package handler

import (
	"net/http"
	"strconv"

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

	if err := h.Repository.SetOrderStatus(id, "удалён"); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// После удаления можно редиректить на главную
	ctx.Redirect(http.StatusSeeOther, "/")
}
