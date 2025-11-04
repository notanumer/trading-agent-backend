package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// @Summary      Start the bot
// @Description  Start the bot
// @Tags         Bot
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /bot/start [post]
func (h *Handler) Start(c *gin.Context) {
	h.botSvc.Start()
	c.JSON(http.StatusOK, gin.H{"status": "started", "on": true})
}

// @Summary      Stop the bot
// @Description  Stop the bot
// @Tags         Bot
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /bot/stop [post]
func (h *Handler) Stop(c *gin.Context) {
	h.botSvc.Stop()
	c.JSON(http.StatusOK, gin.H{"status": "stopped", "on": false})
}

// @Summary      Get the bot status
// @Description  Get the bot status
// @Tags         Bot
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /bot/status [get]
func (h *Handler) Status(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"on": h.botSvc.IsOn()})
}
