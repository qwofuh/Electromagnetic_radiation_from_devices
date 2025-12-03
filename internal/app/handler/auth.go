package handler

import (
	"lab1/internal/app/ds"
	"lab1/internal/app/role"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/sirupsen/logrus"
)

// LoginUserAPI godoc
// @Summary      Аутентификация пользователя
// @Description  Авторизует пользователя по логину и паролю, возвращает JWT-токен и время жизни
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      ds.LoginReq  true  "Данные пользователя"
// @Success      200      {object}  ds.LoginResponse
// @Failure      400      {object}  map[string]string  "Неверный формат запроса"
// @Failure      401      {object}  map[string]string  "Пользователь не найден или неверный пароль"
// @Router       /api/users/login [post]
func (h *Handler) LoginUserAPI(ctx *gin.Context) {
	var body struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	// Проверяем тело запроса
	if err := ctx.BindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// ОТЛАДКА: выводим что пришло
	logrus.Printf("Login attempt: login=%s, password=%s", body.Login, body.Password)

	// Получаем пользователя по логину
	user, err := h.Repository.GetUserByLogin(body.Login)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "пользователь не найден"})
		return
	}

	hashedPassword := generateHashString(body.Password)
	logrus.Printf("Input password: %s", body.Password)
	logrus.Printf("Input password hash: %s", hashedPassword)
	logrus.Printf("Stored password hash: %s", user.Password)
	logrus.Printf("Hashes match: %t", user.Password == hashedPassword)

	// Проверяем пароль
	if user.Password != generateHashString(body.Password) {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "неверный пароль"})
		return
	}

	// Генерация JWT токена
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &ds.JWTClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(h.Config.JWT.AccessTokenTTL).Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "bitop-admin",
		},
		UserID: user.ID,
		Role:   user.Role,
	})

	strToken, err := token.SignedString([]byte(h.Config.JWT.AccessSecret))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось сгенерировать токен"})
		return
	}

	// Добавляем токен в заголовок
	ctx.Header("Authorization", "Bearer "+strToken)

	// Возвращаем только данные пользователя без токена
	ctx.JSON(http.StatusOK, gin.H{
		"id":    user.ID,
		"login": user.Login,
		"role":  user.Role,
	})
}

// LogoutUserAPI godoc
// @Summary      Деаутентификация пользователя
// @Description  Добавляет JWT-токен в черный список (блеклист) Redis, чтобы он стал недействительным
// @Tags         auth
// @Security     BearerAuth
// @Produce      json
// @Success      200  "Успешный выход"
// @Failure      400  {object}  ds.ErrorResponse  "Некорректный токен"
// @Failure      500  {object}  ds.ErrorResponse  "Ошибка при записи токена в блеклист"
// @Router       /api/users/logout [post]
func (h *Handler) LogoutUserAPI(ctx *gin.Context) {
	authHeader := ctx.GetHeader("Authorization")
	const prefix = "Bearer "
	if !strings.HasPrefix(authHeader, prefix) {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	jwtStr := authHeader[len(prefix):]
	logrus.Printf("Logout: processing token %s...", jwtStr[:10]) // первые 10 символов

	// Парсинг токена для проверки
	_, err := jwt.ParseWithClaims(jwtStr, &ds.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.Config.JWT.AccessSecret), nil
	})
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	// Записываем в блеклист redis с TTL равным оставшемуся сроку жизни токена
	if h.Redis == nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	logrus.Printf("Redis client: %+v", h.Redis) // выведет структуру

	if err := h.Redis.WriteJWTToBlacklist(ctx.Request.Context(), jwtStr, h.Config.JWT.AccessTokenTTL); err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.Status(http.StatusOK)
}

func (h *Handler) WithAuthCheck(allowedRoles ...role.Role) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		const prefix = "Bearer "

		if !strings.HasPrefix(authHeader, prefix) {
			// Нет токена → 401
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		jwtStr := authHeader[len(prefix):]

		// Проверяем блеклист Redis
		if h.Redis != nil {
			if err := h.Redis.CheckJWTInBlacklist(ctx.Request.Context(), jwtStr); err == nil {
				// токен в блеклисте → 401
				ctx.AbortWithStatus(http.StatusUnauthorized)
				return
			}
		}

		// Разбираем токен
		token, err := jwt.ParseWithClaims(jwtStr, &ds.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(h.Config.JWT.AccessSecret), nil
		})
		if err != nil || !token.Valid {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(*ds.JWTClaims)
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// Проверяем роль, если переданы allowedRoles
		if len(allowedRoles) > 0 {
			allowed := false
			for _, r := range allowedRoles {
				if claims.Role == r {
					allowed = true
					break
				}
			}
			if !allowed {
				// Авторизован, но нет нужной роли → 403
				ctx.AbortWithStatus(http.StatusForbidden)
				return
			}
		}

		// Сохраняем информацию о пользователе в контексте
		ctx.Set("userID", claims.UserID)
		ctx.Set("role", claims.Role)

		ctx.Next()
	}
}
