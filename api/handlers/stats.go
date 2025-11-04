package handlers

import (
	"context"
	"deepseek-trader/api/middleware"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// @Summary      Get the latest stats
// @Description  Get the latest stats
// @Tags         Stats
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /stats/latest [get]
func (h *Handler) Stats(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	ctx, cancel := context.WithTimeout(c, 10*time.Second)
	defer cancel()
	st, err := h.stats.Latest(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	w, err := h.wallet.FindLatestByUser(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.hl.SetWalletAddress(w.Address)

	if live, err := h.hl.GetLiveStats(ctx); err == nil {
		st.Balance = live.Balance
		st.PnL = live.PnL
		st.ROE = live.ROE
	}
	c.JSON(http.StatusOK, st)
}
