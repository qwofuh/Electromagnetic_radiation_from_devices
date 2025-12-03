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

// GetDeviceAPI godoc
// @Summary      Получить устройство по ID
// @Description  Возвращает информацию об устройстве по его идентификатору
// @Tags         Устройства
// @Produce      json
// @Param        id   path      int  true  "ID устройства"
// @Success      200  {object}  map[string]interface{}  "Успешный ответ с данными устройства"
// @Failure      400  {object}  map[string]string  "Некорректный ID устройства"
// @Failure      404  {object}  map[string]string  "Устройство не найдено"
// @Failure      500  {object}  map[string]string  "Ошибка на сервере"
// @Router       /api/device/:id [get]
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

// GetDevicesAPI godoc
// @Summary      Получить список устройств
// @Description  Возвращает список всех устройств с возможностью фильтрации по названию
// @Tags         Устройства
// @Produce      json
// @Param        title  query     string  false  "Фильтр по названию устройства"
// @Success      200    {object}  map[string]interface{}  "Успешный ответ со списком устройств"
// @Failure      500    {object}  map[string]string  "Ошибка на сервере"
// @Router       /api/devices [get]
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

// CreateDeviceAPI godoc
// @Summary      Создать новое устройство
// @Description  Создает новое устройство с указанными параметрами. Доступно только администраторам
// @Tags         Устройства
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request  body      ds.Device  true  "Данные нового устройства"
// @Success      201      {object}  map[string]interface{}  "Устройство успешно создано"
// @Failure      400      {object}  map[string]string  "Некорректные данные устройства"
// @Failure      403      {object}  map[string]string  "Доступ запрещен (требуется роль Admin)"
// @Failure      500      {object}  map[string]string  "Ошибка при создании устройства"
// @Router       /api/device [post]
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

// UpdateDeviceAPI godoc
// @Summary      Обновить устройство
// @Description  Обновляет данные существующего устройства по его ID. Доступно только администраторам
// @Tags         Устройства
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id       path      int         true  "ID устройства"
// @Param        request  body      ds.Device   true  "Обновленные данные устройства"
// @Success      200      {object}  map[string]interface{}  "Устройство успешно обновлено"
// @Failure      400      {object}  map[string]string  "Некорректный ID или данные устройства"
// @Failure      403      {object}  map[string]string  "Доступ запрещен (требуется роль Admin)"
// @Failure      500      {object}  map[string]string  "Ошибка при обновлении устройства"
// @Router       /api/device/:id [put]
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

// DeleteDeviceAPI godoc
// @Summary      Удалить устройство
// @Description  Помечает устройство как скрытое (soft delete). Доступно только администраторам
// @Tags         Устройства
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      int  true  "ID устройства"
// @Success      200  {object}  map[string]string  "Устройство успешно скрыто"
// @Failure      400  {object}  map[string]string  "Некорректный ID устройства"
// @Failure      403  {object}  map[string]string  "Доступ запрещен (требуется роль Admin)"
// @Failure      500  {object}  map[string]string  "Ошибка при удалении устройства"
// @Router       /api/device/:id/delete [post]
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

// AddDeviceToDraftOrderAPI godoc
// @Summary      Добавить устройство в черновой заказ
// @Description  Добавляет устройство в черновой заказ пользователя. Если чернового заказа нет, создает новый
// @Tags         Заявки с устройствами
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      int  true  "ID устройства"
// @Success      200  {object}  map[string]interface{}  "Устройство успешно добавлено в черновой заказ"
// @Failure      400  {object}  map[string]string  "Некорректный ID устройства"
// @Failure      500  {object}  map[string]string  "Ошибка при добавлении устройства"
// @Router       /api/emissions_calculation/draft/add/:id [post]
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

// UploadDeviceImage godoc
// @Summary      Загрузить изображение устройства
// @Description  Загружает или заменяет изображение для устройства. Доступно только администраторам
// @Tags         Устройства
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        id     path      int   true  "ID устройства"
// @Param        image  formData  file  true  "Изображение устройства"
// @Success      200    {object}  map[string]string  "Изображение успешно загружено"
// @Failure      400    {object}  map[string]string  "Некорректный ID или отсутствует файл изображения"
// @Failure      403    {object}  map[string]string  "Доступ запрещен (требуется роль Admin)"
// @Failure      500    {object}  map[string]string  "Ошибка при загрузке изображения"
// @Router       /api/device/:id/image [post]
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
