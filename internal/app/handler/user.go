package handler

import (
	"fmt"
	"lab1/internal/app/ds"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h *Handler) RegisterUserAPI(ctx *gin.Context) {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := ctx.BindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("некорректное тело запроса"))
		return
	}

	user := ds.User{
		Login:    req.Login,
		Password: req.Password,
	}

	if err := h.Repository.CreateUser(&user); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"user": gin.H{
			"id":           user.ID,
			"login":        user.Login,
			"is_moderator": user.IsModerator,
		},
	})
}

func (h *Handler) LoginUserAPI(ctx *gin.Context) {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := ctx.BindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("некорректное тело запроса"))
		return
	}

	user, err := h.Repository.GetUserByLogin(req.Login)
	if err != nil || user.Password != req.Password {
		h.errorHandler(ctx, http.StatusUnauthorized, fmt.Errorf("неверный логин или пароль"))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"user": gin.H{
			"id":           user.ID,
			"login":        user.Login,
			"is_moderator": user.IsModerator,
		},
	})
}

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
		"is_moderator": user.IsModerator,
	})
}

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

func (h *Handler) LogoutUserAPI(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"status": "success"})
}
