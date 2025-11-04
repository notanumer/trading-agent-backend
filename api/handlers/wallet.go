package handlers

import (
	"context"
	"deepseek-trader/api/middleware"
	"deepseek-trader/services"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// @Summary      Connect a wallet
// @Description  Connect a wallet
// @Tags         Wallet
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /wallet/connect [post]
func (h *Handler) Connect(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req services.ConnectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c, 15*time.Second)
	defer cancel()
	res, err := h.wallet.Connect(ctx, req, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

// @Summary      Disconnect a wallet
// @Description  Disconnect a wallet
// @Tags         Wallet
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /wallet/disconnect [delete]
func (h *Handler) Disconnect(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	ctx, cancel := context.WithTimeout(c, 10*time.Second)
	defer cancel()
	if err := h.wallet.Disconnect(ctx, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary      Get the latest wallet
// @Description  Get the latest wallet
// @Tags         Wallet
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /wallet/latest [get]
func (h *Handler) Latest(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	ctx, cancel := context.WithTimeout(c, 10*time.Second)
	defer cancel()
	w, err := h.wallet.FindLatestByUser(ctx, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "wallet not found"})
		return
	}
	w.APIKey = ""
	c.JSON(http.StatusOK, w)
}
