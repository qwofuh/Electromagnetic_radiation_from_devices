package handler

import (
	"net/http"
	"strconv"

	"lab1/internal/app/ds"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *Handler) GetDevices(ctx *gin.Context) {
	searchQuery := ctx.Query("device_search")
	var devices []ds.Device
	var err error

	if searchQuery == "" {
		devices, err = h.Repository.GetDevices()
	} else {
		devices, err = h.Repository.GetDeviceByTitle(searchQuery)
	}

	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// Для примера используем userID = 1
	userID := 1

	// Ищем черновой заказ пользователя
	order, err := h.Repository.GetDraftOrder(userID)
	var orderCount int64 = 0
	if err != nil {
		logrus.Warn("Не удалось получить черновой заказ: ", err)
	} else if order != nil {
		// Получаем количество материалов в черновике
		orderCount, err = h.Repository.GetOrderDevicesCount(order.ID)
		if err != nil {
			logrus.Warn("Не удалось получить количество материалов в черновике: ", err)
			orderCount = 0
		}
	}

	logrus.Infof("Found %d devices", len(devices))
	for i, device := range devices {
		logrus.Infof("Device %d: %s, Image: %s", i, device.Title, device.Image)
	}

	ctx.HTML(http.StatusOK, "devices.html", gin.H{
		"devices":    devices,
		"query":      searchQuery,
		"orderCount": orderCount,
		// передаём ID чернового заказа в шаблон (0 если заказа нет)
		"orderID": func() int {
			if order != nil {
				return order.ID
			}
			return 0
		}(),
	})
}

// POST /orders/draft/add/:id
func (h *Handler) AddDeviceToDraftOrder(ctx *gin.Context) {
	deviceIDStr := ctx.Param("id")
	deviceID, err := strconv.Atoi(deviceIDStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	userID := 1

	// Получаем черновой заказ
	order, err := h.Repository.GetDraftOrder(userID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	if order == nil {
		order, err = h.Repository.CreateDraftOrder(userID)
		if err != nil {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
			return
		}
	}

	if err := h.Repository.AddDeviceToOrder(order.ID, deviceID); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// Сохраняем count в сессии или просто редиректим на текущую страницу
	ctx.Redirect(http.StatusSeeOther, ctx.Request.Referer())
}

// Получение заказа по ID
func (h *Handler) GetDevicesOrder(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	// Если id == 0 — отвечаем унифицированным сообщением, не перенаправляя/не показывая внутреннюю ошибку
	if id == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{
			"status":      "error",
			"description": "заказ не найден или удален",
		})
		return
	}

	order, device, err := h.Repository.GetOrderByID(id)
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}

	if order.Status != "черновик" {
		ctx.JSON(http.StatusNotFound, gin.H{
			"status":      "error",
			"description": "заказ не найден или удален",
		})
		return
	}

	ctx.HTML(http.StatusOK, "emissions_calculation.html", gin.H{
		"order":  order,
		"device": device,
	})
}
