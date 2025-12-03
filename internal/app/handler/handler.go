package handler

import (
	"lab1/internal/app/config"
	redisclient "lab1/internal/app/redis"
	"lab1/internal/app/repository"
	"lab1/internal/app/role"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Repository *repository.Repository
	Config     *config.Config
	Redis      *redisclient.Client
}

func NewHandler(r *repository.Repository, cfg *config.Config, redis *redisclient.Client) *Handler {
	return &Handler{
		Repository: r,
		Config:     cfg,
		Redis:      redis,
	}
}

// RegisterHandler регистрируем маршруты
func (h *Handler) RegisterHandler(router *gin.Engine) {
	// ---------------------------
	// Публичные маршруты
	// ---------------------------
	router.GET("/api/devices", h.GetDevicesAPI)
	router.GET("/device/:id", h.GetDevice)
	router.GET("/api/device/:id", h.GetDeviceAPI)
	router.POST("/sign_up", h.Register)
	router.POST("/api/users/login", h.LoginUserAPI)
	router.POST("/api/users/logout", h.LogoutUserAPI)
	router.PUT("/api/material_orders/:id/results", h.ReceiveCalculationResultsAPI)

	// ---------------------------
	// Защищённые маршруты для всех авторизованных (User + Admin)
	// ---------------------------
	auth := router.Group("/api")
	auth.Use(h.WithAuthCheck(role.User, role.Admin))
	{
		// Пользователи
		auth.GET("/users/:id", h.GetUserAPI)
		auth.PUT("/users/:id", h.UpdateUserAPI)

		// Устройства и заказы (все методы кроме админских)
		auth.GET("/devices/:id", h.GetDevicesOrder)
		auth.POST("/emissions_calculation/draft/add/:id", h.AddDeviceToDraftOrderAPI)
		auth.GET("/emissions_calculation/user", h.GetOrdersUserAPI)
		auth.GET("/emissions_calculation/:id", h.GetOrderWithDevicesAPI)
		auth.GET("/emissions_calculation/draft/cart", h.GetDraftCartAPI)
		auth.PUT("/emissions_calculation/:id", h.UpdateDeviceOrderAPI)
		auth.PUT("/emissions_calculation/:id/form", h.FormDeviceOrderAPI)
		auth.PUT("/emissions_calculation/devices/:order_id/:device_id/custom_power", h.UpdateCustomPowerAPI)
		auth.POST("/emissions_calculation/delete/:id", h.DeleteDevicesOrderAPI)
		auth.DELETE("/emissions_calculation/:order_id/device/:device_id", h.DeleteDeviceFromOrderAPI)
	}

	// ---------------------------
	// Только Admin
	// ---------------------------
	admin := router.Group("/api")
	admin.Use(h.WithAuthCheck(role.Admin))
	{
		admin.GET("/emissions_calculation", h.GetOrdersAPI)
		admin.POST("/device", h.CreateDeviceAPI)
		admin.PUT("/device/:id", h.UpdateDeviceAPI)
		admin.POST("/device/:id/image", h.UploadDeviceImage)
		admin.POST("/device/:id/delete", h.DeleteDeviceAPI)
		admin.PUT("/emissions_calculation/:id/complete", h.CompleteOrRejectOrderAPI)
	}
}

func (h *Handler) RegisterStatic(router *gin.Engine) {

	router.LoadHTMLGlob("templates/*")
	router.Static("/static", "./resources")
}

// errorHandler для удобного вывода ошибок
func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{
		"status":      "error",
		"description": err.Error(),
	})
}

// getUserFromContext извлекает ID и роль пользователя из gin.Context
// Возвращает userID и role (int), при отсутствии — нули
func (h *Handler) getUserFromContext(ctx *gin.Context) (int, role.Role) {
	uid, _ := ctx.Get("userID")
	rid, _ := ctx.Get("role")

	userID := 0
	if v, ok := uid.(int); ok {
		userID = v
	}

	var roleID role.Role
	if r, ok := rid.(role.Role); ok {
		roleID = r
	} else if r, ok := rid.(int); ok {
		// На случай, если где-то сохранили как int
		roleID = role.Role(r)
	}

	return userID, roleID
}
