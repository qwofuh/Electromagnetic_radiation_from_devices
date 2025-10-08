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

	router.POST("/orders/draft/add/:id", h.AddDeviceToDraftOrder)
	router.POST("/orders/delete/:id", h.DeleteDeviceOrder)

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
