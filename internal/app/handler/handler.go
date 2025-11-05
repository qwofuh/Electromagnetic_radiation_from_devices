package handler

import (
	"lab1/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

// RegisterHandler регистрируем маршруты
func (h *Handler) RegisterHandler(router *gin.Engine) {

	router.GET("/", h.GetDevices)
	router.GET("/device/:id", h.GetDevice)
	router.GET("/emissions_calculation/:id", h.GetDevicesOrder)

	router.GET("/api/device/:id", h.GetDeviceAPI)
	router.GET("/api/devices", h.GetDevicesAPI)
	router.GET("/api/emissions_calculation/draft/cart", h.GetDraftCartAPI)
	router.GET("/api/emissions_calculation", h.GetOrdersAPI)
	router.GET("/api/emissions_calculation/:id", h.GetOrderWithDevicesAPI)
	router.GET("/api/users/:id", h.GetUserAPI)

	router.POST("/emissions_calculation/draft/add/:id", h.AddDeviceToDraftOrder)
	router.POST("/emissions_calculation/delete/:id", h.DeleteDeviceOrder)

	router.POST("/api/device", h.CreateDeviceAPI)
	router.POST("/api/emissions_calculation/draft/add/:id", h.AddDeviceToDraftOrderAPI)
	router.POST("/api/device/:id/image", h.UploadDeviceImage)
	router.POST("/api/device/:id/delete", h.DeleteDeviceAPI)
	router.POST("/api/emissions_calculation/delete/:id", h.DeleteDevicesOrderAPI)
	router.POST("/api/users/register", h.RegisterUserAPI)
	router.POST("/api/users/login", h.LoginUserAPI)
	router.POST("/api/users/logout", h.LogoutUserAPI)

	router.PUT("/api/device/:id", h.UpdateDeviceAPI)
	router.PUT("/api/emissions_calculation/:id", h.UpdateDeviceOrderAPI)
	router.PUT("/api/emissions_calculation/:id/form", h.FormDeviceOrderAPI)
	router.PUT("/api/emissions_calculation/:id/complete", h.CompleteOrRejectOrderAPI)
	router.PUT("/api/emissions_calculation/devices/:order_id/:device_id/custom_power", h.UpdateCustomPowerAPI)
	router.PUT("/api/users/:id", h.UpdateUserAPI)

	router.DELETE("/api/emissions_calculation/:order_id/device/:device_id", h.DeleteDeviceFromOrderAPI)

}

func (h *Handler) RegisterStatic(router *gin.Engine) {

	router.LoadHTMLGlob("templates/*")
	router.Static("/static", "./resources")
}

// errorHandler для удобного вывода ошибок
func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{
		"description": err.Error(),
	})
}
