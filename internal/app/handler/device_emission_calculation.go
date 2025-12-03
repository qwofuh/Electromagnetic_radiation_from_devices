package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// DeleteDeviceFromOrderAPI godoc
// @Summary      Удалить устройство из заказа
// @Description  Удаляет устройство из расчета выбросов по ID заказа и ID устройства
// @Tags         Заявки с устройствами
// @Security     BearerAuth
// @Produce      json
// @Param        order_id   path      int  true  "ID заказа"
// @Param        material_id  path      int  true  "ID устройства (material_id)"
// @Success      200        {object}  map[string]interface{}  "Устройство успешно удалено из расчета"
// @Failure      400        {object}  map[string]string  "Некорректный ID заказа или устройства"
// @Failure      500        {object}  map[string]string  "Ошибка при удалении устройства"
// @Router       /api/emissions_calculation/:order_id/device/:material_id [delete]
func (h *Handler) DeleteDeviceFromOrderAPI(ctx *gin.Context) {
	orderIDStr := ctx.Param("order_id")
	deviceIDStr := ctx.Param("device_id")

	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("некорректный ID заказа"))
		return
	}

	deviceID, err := strconv.Atoi(deviceIDStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("некорректный ID устройства"))
		return
	}

	if err := h.Repository.DeleteDeviceFromOrder(deviceID, orderID); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"device_id": deviceID,
		"order_id":  orderID,
		"message":   "устройство удалено из расчета",
	})
}

// UpdateCustomPowerAPI godoc
// @Summary      Обновить пользовательскую мощность устройства
// @Description  Обновляет значение пользовательской мощности (custom_power) для устройства в заказе
// @Tags         Заявки с устройствами
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        order_id     path      int     true  "ID заказа"
// @Param        material_id  path      int     true  "ID устройства (material_id)"
// @Param        request      body      object  true  "Данные для обновления"  example:{"custom_power":150.5}
// @Success      200          {object}  map[string]interface{}  "Мощность успешно обновлена"
// @Failure      400          {object}  map[string]string  "Некорректный ID или тело запроса"
// @Failure      500          {object}  map[string]string  "Ошибка при обновлении мощности"
// @Router       /api/emissions_calculation/devices/:order_id/:material_id/custom_power [put]
func (h *Handler) UpdateCustomPowerAPI(ctx *gin.Context) {
	orderIDStr := ctx.Param("order_id")
	deviceIDStr := ctx.Param("device_id")

	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("некорректный ID заказа"))
		return
	}

	deviceID, err := strconv.Atoi(deviceIDStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("некорректный ID устройства"))
		return
	}

	var req struct {
		CustomPower float64 `json:"custom_power"`
	}

	if err := ctx.BindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("некорректное тело запроса"))
		return
	}

	if err := h.Repository.UpdateCustomPower(deviceID, orderID, req.CustomPower); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":       "success",
		"device_id":    deviceID,
		"order_id":     orderID,
		"custom_power": req.CustomPower,
	})
}
