package handlers

import (
	"context"
	"deepseek-trader/api/middleware"
	"deepseek-trader/services"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// @Summary      Get the trades history
// @Description  Get the trades history
// @Tags         Trades
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /trades/history [get]
func (h *Handler) TradesHistory(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	limit := 100
	if qs := c.Query("limit"); qs != "" {
		if n, err := strconv.Atoi(qs); err == nil {
			limit = n
		}
	}
	ctx, cancel := context.WithTimeout(c, 15*time.Second)
	defer cancel()

	w, err := h.wallet.FindLatestByUser(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.hl.SetWalletAddress(w.Address)

	raw, err := h.hl.HistoricalOrders(ctx, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Map to a simple list compatible with frontend
	res := make([]map[string]any, 0, len(raw))
	for _, item := range raw {
		ord, _ := item["order"].(map[string]any)
		if ord == nil {
			continue
		}
		coin, _ := ord["coin"].(string)
		side, _ := ord["side"].(string)
		pxStr, _ := ord["limitPx"].(string)
		szStr, _ := ord["origSz"].(string)
		ts, _ := ord["timestamp"].(float64)
		price, _ := strconv.ParseFloat(pxStr, 64)
		qty, _ := strconv.ParseFloat(szStr, 64)
		sideText := "BUY"
		if strings.EqualFold(side, "A") {
			sideText = "SELL"
		}
		res = append(res, map[string]any{
			"id":        ord["oid"],
			"symbol":    coin + "USDT",
			"side":      sideText,
			"qty":       qty,
			"price":     price,
			"pnl":       0,
			"createdAt": time.UnixMilli(int64(ts)).UTC().Format(time.RFC3339),
		})
	}
	c.JSON(http.StatusOK, res)
}

// @Summary      Get the trades summary
// @Description  Get the trades summary
// @Tags         Trades
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /trades/summary [get]
func (h *Handler) TradesSummary(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	limit := 200
	if qs := c.Query("limit"); qs != "" {
		if n, err := strconv.Atoi(qs); err == nil {
			limit = n
		}
	}

	ctx, cancel := context.WithTimeout(c, 20*time.Second)
	defer cancel()

	w, err := h.wallet.FindLatestByUser(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.hl.SetWalletAddress(w.Address)

	raw, err := h.hl.HistoricalOrders(ctx, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	fees, _ := h.hl.UserFees(ctx)
	feeParams := services.FeeParams{Maker: 0.00015, Taker: 0.00045, Discount: 0}
	if fees != nil {
		feeParams.Maker = fees.UserAddRate
		feeParams.Taker = fees.UserCrossRate
		feeParams.Discount = fees.ReferralDiscount + fees.StakingActiveDiscount
		if feeParams.Discount < 0 {
			feeParams.Discount = 0
		}
		if feeParams.Discount > 0.9 {
			feeParams.Discount = 0.9
		}
	}
	summary := services.BuildTradeSummary(raw, feeParams)
	c.JSON(http.StatusOK, summary)

	c.JSON(http.StatusOK, summary)
}
