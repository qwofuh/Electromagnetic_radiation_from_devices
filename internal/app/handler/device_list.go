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
	userID, _ := h.getUserFromContext(ctx)

	// Ищем черновой заказ пользователя
	order, err := h.Repository.GetDraftOrder(ctx.Request.Context(), userID)
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

	userID, _ := h.getUserFromContext(ctx)

	// Получаем черновой заказ
	order, err := h.Repository.GetDraftOrder(ctx.Request.Context(), userID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	if order == nil {
		order, err = h.Repository.CreateDraftOrder(ctx.Request.Context(), userID)
		if err != nil {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
			return
		}
	}

	if err := h.Repository.AddDeviceToOrder(ctx.Request.Context(), order.ID, deviceID); err != nil {
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
			"description": "заказ не найден или удален",
		})
		return
	}

	order, device, err := h.Repository.GetOrderByID(id)
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}

	// Если статус заказа не черновик, считаем, что заказа нет/он удалён
	if order.Status != "черновик" {
		ctx.JSON(http.StatusNotFound, gin.H{
			"description": "заказ не найден или удален",
		})
		return
	}

	ctx.HTML(http.StatusOK, "emissions_calculation.html", gin.H{
		"order":  order,
		"device": device,
	})
}

// GET /api/devices/:id
func (h *Handler) GetDeviceAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	device, err := h.Repository.GetDeviceByID(id)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	if device == nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"status":      "error",
			"description": "устройство не найдено",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"device": device,
	})
}

// GET /api/devices?title=<название>
func (h *Handler) GetDevicesAPI(ctx *gin.Context) {
	title := ctx.Query("title")

	devices, err := h.Repository.GetDevicesFiltered(title)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"devices": devices,
	})
}

// POST /api/device
func (h *Handler) CreateDeviceAPI(ctx *gin.Context) {
	var input ds.Device

	// Привязываем JSON из запроса
	if err := ctx.ShouldBindJSON(&input); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	// Создаём устройство через репозиторий
	if err := h.Repository.CreateDevice(&input); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"device": input,
	})
}

// PUT /api/device/:id
func (h *Handler) UpdateDeviceAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	var input ds.Device
	if err := ctx.ShouldBindJSON(&input); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	err = h.Repository.UpdateDevice(id, &input)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"device": input,
	})
}

// POST /api/device/:id/delete
func (h *Handler) DeleteDeviceAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	err = h.Repository.DeleteDevice(id)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "устройство успешно скрыто",
	})
}

// POST /api/orders/draft/add/:id
func (h *Handler) AddDeviceToDraftOrderAPI(ctx *gin.Context) {

	deviceIDStr := ctx.Param("id")
	deviceID, err := strconv.Atoi(deviceIDStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	userID, _ := h.getUserFromContext(ctx)

	// Получаем черновой заказ пользователя
	order, err := h.Repository.GetDraftOrder(ctx.Request.Context(), userID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// Если чернового заказа нет — создаём новый
	if order == nil {
		order, err = h.Repository.CreateDraftOrder(ctx.Request.Context(), userID)
		if err != nil {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
			return
		}
	}

	// Добавляем материал в заказ
	if err := h.Repository.AddDeviceToOrder(ctx.Request.Context(), order.ID, deviceID); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// Получаем новое количество материалов в заказе
	count, _ := h.Repository.GetOrderDevicesCount(order.ID)

	ctx.JSON(http.StatusOK, gin.H{
		"message":   "устройство добавлено в черновой заказ",
		"orderID":   order.ID,
		"itemCount": count,
	})
}

// Загрузить/заменить изображение материала
func (h *Handler) UploadDeviceImage(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid material id"})
		return
	}

	fileHeader, err := ctx.FormFile("image")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "no image file"})
		return
	}

	if err := h.Repository.UploadDeviceImage(id, fileHeader); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "image uploaded"})
}
