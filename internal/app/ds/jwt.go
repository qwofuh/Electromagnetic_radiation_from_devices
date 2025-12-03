package ds

import (
	"lab1/internal/app/role"
	"time"

	"github.com/golang-jwt/jwt"
)

type JWTClaims struct {
	jwt.StandardClaims
	UserID int       `json:"user_id"`
	Role   role.Role `json:"role"`
}

type LoginReq struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginResp struct {
	ExpiresIn   time.Duration `json:"expires_in"`
	AccessToken string        `json:"access_token"`
	TokenType   string        `json:"token_type"`
}

type LoginResponse struct {
	ExpiresIn   int64  `json:"expires_in" example:"3600"` // время жизни токена (в секундах)
	AccessToken string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"`
	TokenType   string `json:"token_type" example:"Bearer"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"Некорректный токен"`
}
