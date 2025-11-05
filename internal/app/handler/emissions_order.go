package handler

import (
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

// GET /api/orders/draft/cart
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
// @Success      200     {object}  ds.OrdersListResponse  "Успешный ответ со списком заказов"
// @Failure      500     {object}  ds.ErrorResponse       "Ошибка на сервере"
// @Security BearerAuth
// @Router       /api/orders [get]

func (h *Handler) GetOrdersAPI(ctx *gin.Context) {
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
	if roleID == 2 { // role.Admin == 2
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

	if err := h.Repository.CompleteOrRejectOrder(orderID, req); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	order, devices, err := h.Repository.GetOrderByID(orderID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"order":   order,
		"devices": devices,
	})
}

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
