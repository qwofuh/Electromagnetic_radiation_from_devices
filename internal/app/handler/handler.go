package handler

import (
	"lab1/internal/app/repository"
	"net/http"
	"strconv"
	"time"

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

func (h *Handler) GetDevices(ctx *gin.Context) {
	var devices []repository.Device
	var err error

	searchQuery := ctx.Query("device_search") // получаем значение из поля поиска
	if searchQuery == "" {                    // если поле поиска пусто, то просто получаем из репозитория все записи
		devices, err = h.Repository.GetDevices()
		if err != nil {
			logrus.Error(err)
		}
	} else {
		devices, err = h.Repository.GetDeviceByTitle(searchQuery) // в ином случае ищем заказ по заголовку
		if err != nil {
			logrus.Error(err)
		}
	}

	ctx.HTML(http.StatusOK, "devices.html", gin.H{
		"time":          time.Now().Format("15:04:05"),
		"devices":       devices,
		"device_search": searchQuery, // передаем введенный запрос обратно на страницу
		// в ином случае оно будет очищаться при нажатии на кнопку
	})
}

func (h *Handler) GetDevice(ctx *gin.Context) {
	idStr := ctx.Param("id") // получаем id заказа из урла (то есть из /order/:id)
	// через двоеточие мы указываем параметры, которые потом сможем считать через функцию выше
	id, err := strconv.Atoi(idStr) // так как функция выше возвращает нам строку, нужно ее преобразовать в int
	if err != nil {
		logrus.Error(err)
	}

	device, err := h.Repository.GetDevice(id)
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "device.html", gin.H{
		"device": device,
	})
}

func (h *Handler) Calculate_emissions(ctx *gin.Context) {

	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id < 0 {
		logrus.Error(err)
	}

	order, _ := h.Repository.GetOrder()

	ctx.HTML(http.StatusOK, "emissions_calculation.html", gin.H{
		"order": order,
	})
}
