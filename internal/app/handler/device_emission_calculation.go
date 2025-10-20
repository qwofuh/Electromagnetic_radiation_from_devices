package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

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
