package api

import (
	"assignment1/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

type RateHandler struct {
	Service *service.RateService
}

func NewRateHandler(rs *service.RateService) *RateHandler {
	return &RateHandler{Service: rs}
}

func (h *RateHandler) GetRate(c *gin.Context) {
	base := c.Query("base")
	target := c.Query("target")
	if base == "" || target == "" {
		RespondError(c, http.StatusBadRequest, "Missing base or target parameter")
		return
	}

	data, err := h.Service.GetRate(base, target)
	if err != nil {
		RespondError(c, http.StatusNotFound, "Rate not found or fetch failed")
		return
	}
	RespondSuccess(c, data, "Rate fetched successfully")
}
