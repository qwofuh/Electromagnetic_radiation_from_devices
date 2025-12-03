package handler

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"lab1/internal/app/ds"
	"lab1/internal/app/role"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Register godoc
// @Summary      Регистрация нового пользователя
// @Description  Регистрирует нового пользователя с логином, паролем и (опционально) ролью.
// @Tags         Управление пользователями
// @Accept       json
// @Produce      json
// @Param        request  body      ds.RegisterReq  true  "Данные нового пользователя"
// @Success      200      {object}  ds.RegisterResp  "Пользователь успешно зарегистрирован"
// @Failure      400      {object}  map[string]string  "Некорректные данные или пустые поля"
// @Failure      500      {object}  map[string]string  "Ошибка при сохранении пользователя"
// @Router       /sign_up [post]

func (h *Handler) Register(ctx *gin.Context) {
	req := &ds.RegisterReq{}

	if err := json.NewDecoder(ctx.Request.Body).Decode(req); err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	if req.Login == "" || req.Password == "" {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("login or password is empty"))
		return
	}

	// Всегда ставим роль User (1)
	userRole := role.User

	user := &ds.User{
		Login:    req.Login,
		Password: generateHashString(req.Password),
		Role:     userRole,
	}

	if err := h.Repository.Register(user); err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, &ds.RegisterResp{
		Ok:   true,
		Role: user.Role,
	})
}

// GetUserAPI godoc
// @Summary      Получить информацию о пользователе
// @Description  Возвращает информацию о пользователе по его ID (логин и роль)
// @Tags         Управление пользователями
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      int  true  "ID пользователя"
// @Success      200  {object}  map[string]interface{}  "Успешный ответ с данными пользователя"
// @Failure      400  {object}  map[string]string  "Некорректный ID пользователя"
// @Failure      404  {object}  map[string]string  "Пользователь не найден"
// @Router       /api/users/:id [get]
func (h *Handler) GetUserAPI(ctx *gin.Context) {
	userIDStr := ctx.Param("id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("некорректный ID пользователя"))
		return
	}

	user, err := h.Repository.GetUserByID(userID)
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"id":           user.ID,
		"login":        user.Login,
		"is_moderator": user.Role,
	})
}

// UpdateUserAPI godoc
// @Summary      Обновить данные пользователя
// @Description  Обновляет логин и/или пароль пользователя по его ID
// @Tags         Управление пользователями
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id       path      int     true  "ID пользователя"
// @Param        request  body      object  true  "Данные для обновления"  example:{"login":"newlogin","password":"newpassword"}
// @Success      200      {object}  map[string]string  "Пользователь успешно обновлен"
// @Failure      400      {object}  map[string]string  "Некорректный ID или тело запроса"
// @Failure      500      {object}  map[string]string  "Ошибка при обновлении пользователя"
// @Router       /api/users/:id [put]
func (h *Handler) UpdateUserAPI(ctx *gin.Context) {
	userIDStr := ctx.Param("id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("некорректный ID пользователя"))
		return
	}

	var req struct {
		Login    string `json:"login,omitempty"`
		Password string `json:"password,omitempty"`
	}

	if err := ctx.BindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("некорректное тело запроса"))
		return
	}

	updates := make(map[string]interface{})
	if req.Login != "" {
		updates["login"] = req.Login
	}
	if req.Password != "" {
		updates["password"] = req.Password
	}

	if err := h.Repository.UpdateUser(userID, updates); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success"})
}

// Функция для хеширования пароля
func generateHashString(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}
