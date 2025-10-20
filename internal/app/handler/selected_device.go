package handler

import (
	"net/http"
	"strconv"

	"lab1/internal/app/ds"

	"github.com/gin-gonic/gin"
)

// Получение конкретного устройства по ID
func (h *Handler) GetDevice(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	var device ds.Device
	device, err = h.Repository.GetDevice(id)
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}

	ctx.HTML(http.StatusOK, "device.html", gin.H{
		"device": device,
	})
}
