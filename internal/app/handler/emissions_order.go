package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"lab1/internal/app/ds"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// POST /orders/delete/:id - пометить заказ как удалённый
func (h *Handler) DeleteDeviceOrder(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.SetOrderStatus(id, "удален"); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// После удаления можно редиректить на главную
	ctx.Redirect(http.StatusSeeOther, "/")
}

// GetDraftCartAPI godoc
// @Summary      Получить корзину чернового заказа
// @Description  Возвращает информацию о черновом заказе пользователя: ID заказа и количество устройств в нем
// @Tags         Заявки с устройствами
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "Успешный ответ с данными корзины"
// @Failure      500  {object}  map[string]string  "Ошибка на сервере"
// @Router       /api/emissions_calculation/draft/cart [get]
func (h *Handler) GetDraftCartAPI(ctx *gin.Context) {

	userID, _ := h.getUserFromContext(ctx)

	// Ищем черновик
	order, err := h.Repository.GetDraftOrder(ctx.Request.Context(), userID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	if order == nil {
		// Если черновика нет — возвращаем пустую корзину
		ctx.JSON(http.StatusOK, gin.H{
			"orderID":   0,
			"itemCount": 0,
		})
		return
	}

	// Считаем количество услуг
	count, err := h.Repository.GetOrderDevicesCount(order.ID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"orderID":   order.ID,
		"itemCount": count,
	})
}

// GetOrdersAPI godoc
// @Summary      Получить список заказов
// @Description  Возвращает список заказов с возможностью фильтрации по статусу и диапазону дат. Разрешённые статусы: "сформирован", "завершен", "отклонен".
// @Tags         Заявки с устройствами
// @Accept       json
// @Produce      json
// @Param        status  query     string  false  "Статус заказа (сформирован, завершен, отклонен), можно указать несколько через запятую"
// @Param        start   query     string  false  "Дата начала фильтрации (формат YYYY-MM-DD)"
// @Param        end     query     string  false  "Дата окончания фильтрации (формат YYYY-MM-DD)"
// @Success      200     {object}  ds.EmissionListResponse  "Успешный ответ со списком заказов"
// @Failure      500     {object}  ds.ErrorResponse       "Ошибка на сервере"
// @Security BearerAuth
// @Router       /api/emissions_calculation [get]
func (h *Handler) GetOrdersAPI(ctx *gin.Context) {
	status := ctx.Query("status")
	start := ctx.Query("start")
	end := ctx.Query("end")

	orders, err := h.Repository.GetOrdersFiltered(ctx.Request.Context(), status, start, end)
	if err != nil {
		if err.Error() == "not_found" {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "заказы со статусом 'черновик' или 'удален' не доступны",
			})
			return
		}
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"orders": orders,
	})
}

// GetOrdersAPI godoc
// @Summary      Получить список заказов
// @Description  Возвращает список заказов с возможностью фильтрации по статусу и диапазону дат. Разрешённые статусы: "сформирован", "завершен", "отклонен".
// @Tags         Заявки с устройствами
// @Accept       json
// @Produce      json
// @Param        status  query     string  false  "Статус заказа (сформирован, завершен, отклонен), можно указать несколько через запятую"
// @Param        start   query     string  false  "Дата начала фильтрации (формат YYYY-MM-DD)"
// @Param        end     query     string  false  "Дата окончания фильтрации (формат YYYY-MM-DD)"
// @Success      200     {object}  ds.EmissionListResponse  "Успешный ответ со списком заказов"
// @Failure      500     {object}  ds.ErrorResponse       "Ошибка на сервере"
// @Security BearerAuth
// @Router       /api/emissions_calculation/user [get]
func (h *Handler) GetOrdersUserAPI(ctx *gin.Context) {
	status := ctx.Query("status")
	start := ctx.Query("start")
	end := ctx.Query("end")

	userID, _ := h.getUserFromContext(ctx)

	orders, err := h.Repository.GetOrdersFilteredForUser(ctx.Request.Context(), status, start, end, userID)
	if err != nil {
		if err.Error() == "not_found" {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "заказы со статусом 'черновик' или 'удален' не доступны",
			})
			return
		}
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"orders": orders,
	})
}

// GetOrderWithDevicesAPI godoc
// @Summary      Получить заказ с устройствами
// @Description  Возвращает полную информацию о заказе с расчетом выбросов, включая список устройств. Доступен только владельцу заказа или администратору
// @Tags         Заявки с устройствами
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      int  true  "ID заказа"
// @Success      200  {object}  map[string]interface{}  "Успешный ответ с данными заказа и устройств"
// @Failure      400  {object}  map[string]string  "Некорректный ID заказа"
// @Failure      403  {object}  map[string]string  "Доступ запрещен (не владелец и не администратор)"
// @Failure      404  {object}  map[string]string  "Заказ не найден"
// @Router       /api/emissions_calculation/:id [get]
func (h *Handler) GetOrderWithDevicesAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	orderID, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	order, devices, err := h.Repository.GetOrderByID(orderID)
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}

	// Проверяем, что запрашиваемый заказ принадлежит текущему пользователю или пользователь — Admin
	userID, roleID := h.getUserFromContext(ctx)
	isAdmin := false
	if roleID == 1 { // role.Admin == 2
		isAdmin = true
	}
	if !isAdmin && order.CreatorID != userID {
		// не владелец и не админ — запрещено
		ctx.AbortWithStatus(http.StatusForbidden)
		return
	}

	// Формируем DTO устройств
	var devicesDTO []ds.DeviceInCalculation
	for _, dec := range devices {
		devicesDTO = append(devicesDTO, ds.DeviceInCalculation{
			ID:          dec.DeviceID,
			Title:       dec.Device.Title,
			Image:       dec.Device.Image,
			AvgMaxPower: dec.Device.AvgMaxPower,
			AvgMinPower: dec.Device.AvgMinPower,
		})
	}

	var FinishAt *time.Time
	if order.FinishAt.Valid {
		FinishAt = &order.FinishAt.Time
	}

	resp := ds.EmissionWithDevices{
		ID:            order.ID,
		CreatorID:     order.CreatorID,
		ModeratorID:   order.ModeratorID,
		Status:        order.Status,
		Distance:      order.Distance,
		TotalEmission: order.TotalEmission,
		CreateAt:      order.CreateAt,
		UpdateAt:      order.UpdateAt,
		FinishAt:      FinishAt,
		Devices:       devicesDTO,
	}

	ctx.JSON(http.StatusOK, gin.H{
		"order": resp,
	})
}

// UpdateDeviceOrderAPI godoc
// @Summary      Обновить заказ
// @Description  Обновляет параметры заказа (например, расстояние). Разрешено обновлять только поле distance
// @Tags         Заявки с устройствами
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id       path      int                     true  "ID заказа"
// @Param        request  body      ds.EmissionUpdateRequest  true  "Данные для обновления заказа"
// @Success      200      {object}  map[string]interface{}  "Заказ успешно обновлен"
// @Failure      400      {object}  map[string]string  "Некорректный ID или недопустимые поля в запросе"
// @Failure      500      {object}  map[string]string  "Ошибка при обновлении заказа"
// @Router       /api/emissions_calculation/:id [put]
func (h *Handler) UpdateDeviceOrderAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	// Читаем тело запроса как map
	var raw map[string]interface{}
	if err := ctx.BindJSON(&raw); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	// Разрешённые поля
	allowed := map[string]bool{
		"distance": true,
	}

	// Проверяем лишние поля
	for k := range raw {
		if !allowed[k] {
			h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("недопустимое поле: %s", k))
			return
		}
	}

	// Преобразуем map в DTO
	var req ds.EmissionUpdateRequest
	if v, ok := raw["distance"]; ok {
		if f, ok := v.(float64); ok {
			req.Distance = &f
		}
	}

	if err := h.Repository.UpdateDeviceOrder(id, req); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	order, _, err := h.Repository.GetOrderByID(id)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"order": order,
	})
}

// FormDeviceOrderAPI godoc
// @Summary      Сформировать заказ
// @Description  Переводит черновой заказ в статус "сформирован". Требуется, чтобы все устройства имели заполненное поле custom_power
// @Tags         Заявки с устройствами
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      int  true  "ID заказа"
// @Success      200  {object}  map[string]interface{}  "Заказ успешно сформирован"
// @Failure      400  {object}  map[string]string  "Некорректный ID заказа или не все custom_power заполнены"
// @Failure      500  {object}  map[string]string  "Ошибка при формировании заказа"
// @Router       /api/emissions_calculation/:id/form [put]
func (h *Handler) FormDeviceOrderAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	orderID, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.FormDeviceOrder(orderID); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	// Возвращаем обновлённый заказ
	order, _, err := h.Repository.GetOrderByID(orderID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"order": order,
	})
}

// CompleteOrRejectOrderAPI godoc
// @Summary      Завершить или отклонить заказ
// @Description  Переводит заказ в статус "завершен" или "отклонен". Доступно только администраторам
// @Tags         Заявки с устройствами
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id       path      int                      true  "ID заказа"
// @Param        request  body      ds.EmissionCompleteRequest  true  "Данные для завершения/отклонения заказа"
// @Success      200      {object}  map[string]interface{}  "Заказ успешно обновлен"
// @Failure      400      {object}  map[string]string  "Некорректный ID или данные запроса"
// @Failure      403      {object}  map[string]string  "Доступ запрещен (требуется роль Admin)"
// @Failure      500      {object}  map[string]string  "Ошибка при обновлении заказа"
// @Router       /api/emissions_calculation/:id/complete [put]
func (h *Handler) CompleteOrRejectOrderAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	orderID, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	var req ds.EmissionCompleteRequest
	if err := ctx.BindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	// Обновляем статус заказа
	if err := h.Repository.CompleteOrRejectOrder(orderID, req); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// Получаем обновленный заказ и устройства
	order, devices, err := h.Repository.GetOrderByID(orderID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// Если заказ завершён успешно, запускаем асинхронный расчёт выбросов для устройств
	if req.Status == "завершен" {
		go func() {
			// Формируем payload для асинхронного сервиса (Django)
			type devicePayload struct {
				DeviceID int     `json:"device_id"`
				Power    float64 `json:"power"`
			}

			var deviceList []devicePayload
			for _, device := range devices {
				// Используем кастомную мощность устройства, если она задана
				power := device.Device.AvgMaxPower
				if device.CustomPower.Valid {
					power = device.CustomPower.Float64
				}

				deviceList = append(deviceList, devicePayload{
					DeviceID: device.DeviceID,
					Power:    power,
				})
			}

			payload := map[string]interface{}{
				"order_id":       orderID, // ← должно быть order_id, не emission_id
				"distance":       order.Distance,
				"status":         order.Status,
				"moderator_id":   order.ModeratorID,
				"devices":        deviceList,
				"callback_url":   fmt.Sprintf("http://localhost:8080/api/material_orders/%d/results", orderID),
				"callback_token": CalcCallbackToken,
			}

			// Отправляем запрос в асинхронный сервис
			b, _ := json.Marshal(payload)
			_, err := http.Post(AsyncServiceURL, "application/json", bytes.NewReader(b))
			if err != nil {
				// Логируем ошибку, но не прерываем выполнение
				fmt.Printf("Ошибка отправки в асинхронный сервис: %v\n", err)
			}
		}()
	}

	// Формируем DTO устройств для ответа
	var devicesDTO []ds.DeviceInCalculation
	for _, dec := range devices {
		devicesDTO = append(devicesDTO, ds.DeviceInCalculation{
			ID:          dec.DeviceID,
			Title:       dec.Device.Title,
			Image:       dec.Device.Image,
			AvgMaxPower: dec.Device.AvgMaxPower,
			AvgMinPower: dec.Device.AvgMinPower,
		})
	}

	var FinishAt *time.Time
	if order.FinishAt.Valid {
		FinishAt = &order.FinishAt.Time
	}

	orderWithDevices := ds.EmissionWithDevices{
		ID:            order.ID,
		CreatorID:     order.CreatorID,
		ModeratorID:   order.ModeratorID,
		Status:        order.Status,
		Distance:      order.Distance,
		TotalEmission: order.TotalEmission,
		CreateAt:      order.CreateAt,
		UpdateAt:      order.UpdateAt,
		FinishAt:      FinishAt,
		Devices:       devicesDTO,
	}

	ctx.JSON(http.StatusOK, gin.H{
		"order":   orderWithDevices,
		"devices": devicesDTO,
	})
}

// DeleteDevicesOrderAPI godoc
// @Summary      Удалить заказ
// @Description  Помечает заказ как удаленный (soft delete), устанавливая статус "удален"
// @Tags         Заявки с устройствами
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      int  true  "ID заказа"
// @Success      200  {object}  map[string]interface{}  "Заказ успешно удален"
// @Failure      400  {object}  map[string]string  "Некорректный ID заказа"
// @Failure      500  {object}  map[string]string  "Ошибка при удалении заказа"
// @Router       /api/emissions_calculation/delete/:id [post]
func (h *Handler) DeleteDevicesOrderAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	orderID, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("некорректный ID заказа"))
		return
	}

	// Проставляем статус "удален" и дату завершения
	if err := h.Repository.SoftDeleteOrder(orderID); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"orderID": orderID,
	})
}

const (

	// URL асинхронного сервиса (Django). Измените по необходимости.
	AsyncServiceURL = "http://localhost:8000/api/compute/"
	// Токен псевдо-авторизации (8 байт), который будет передавать асинхронный сервис при POST результатов
	CalcCallbackToken = "ABCDEFGH"
)

func (h *Handler) ReceiveCalculationResultsAPI(ctx *gin.Context) {

	// Псевдо-авторизация по заголовку
	token := ctx.GetHeader("X-Calc-Token")

	if token != CalcCallbackToken {
		fmt.Printf("!!! Неверный токен!\n")
		ctx.AbortWithStatus(http.StatusForbidden)
		return
	}

	idStr := ctx.Param("id")
	orderID, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Printf("!!! Ошибка парсинга ID: %v\n", err)
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	// Читаем сырые данные для отладки
	rawData, _ := ctx.GetRawData()

	// Парсим в map чтобы увидеть реальную структуру
	var rawMap map[string]interface{}
	json.Unmarshal(rawData, &rawMap)

	// Правильная структура для данных от Django
	var body struct {
		Results []struct {
			DeviceID        int     `json:"device_id"`
			CurrentEmission float64 `json:"current_emission"` // ИЗМЕНИЛОСЬ!
			Power           float64 `json:"power"`            // НОВОЕ ПОЛЕ
		} `json:"results"`
		TotalEmission float64 `json:"total_emission"` // ОБЩИЙ выброс
		OrderID       int     `json:"order_id"`
		Status        string  `json:"status"`
		ModeratorID   int     `json:"moderator_id"`
	}

	// Парсим JSON
	decoder := json.NewDecoder(bytes.NewReader(rawData))
	decoder.UseNumber() // Для точного парсинга чисел
	if err := decoder.Decode(&body); err != nil {
		fmt.Printf("!!! Ошибка парсинга JSON: %v\n", err)
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	// Преобразуем в структуры репозитория
	var repoResults []ds.MMResult
	for _, r := range body.Results {
		repoResults = append(repoResults, ds.MMResult{
			DeviceID:      r.DeviceID,
			TotalEmission: r.CurrentEmission, // Берем current_emission!
		})
	}

	// Вариант 1: Используйте общий total_emission из JSON
	// (это сумма всех current_emission)
	if err := h.Repository.UpdateDeviceEmissionCalculationResults(ctx.Request.Context(), orderID, body.TotalEmission); err != nil {
		fmt.Printf("!!! Ошибка репозитория: %v\n", err)
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// Вариант 2: Или обновите напрямую
	/*
		result := h.Repository.DB.Exec(`
			UPDATE emissions
			SET total_emission = ?,
				status = 'завершен',
				updated_at = NOW()
			WHERE id = ?
		`, body.TotalEmission, orderID)

		fmt.Printf("SQL результат: rows=%d, error=%v\n",
			result.RowsAffected, result.Error)
	*/

	ctx.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"message": fmt.Sprintf("Заказ %d обновлен, total_emission=%f",
			orderID, body.TotalEmission),
	})
}
