package handlers

import (
	"deepseek-trader/api/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

// @Summary      Register a new user
// @Description  Register a new user with email and password
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request   body      models.User  true  "Register request"
// @Success      200  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var b struct{ Email, Password string }
	if err := c.ShouldBindJSON(&b); err != nil || b.Email == "" || b.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email and password required"})
		return
	}
	if _, err := h.authSvc.Register(c, b.Email, b.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "registered"})
}

// @Summary      Login a user
// @Description  Login a user with email and password
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request   body      models.User  true  "Login request"
// @Success      200  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var b struct{ Email, Password string }
	if err := c.ShouldBindJSON(&b); err != nil || b.Email == "" || b.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email and password required"})
		return
	}
	tok, err := h.authSvc.Login(c, b.Email, b.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": tok})
}

// @Summary      Get the current user
// @Description  Get the current user
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /auth/me [get]
func (h *Handler) Me(c *gin.Context) {
	uid, _ := middleware.GetUserID(c)
	email, _ := middleware.GetEmail(c)
	c.JSON(http.StatusOK, gin.H{"id": uid, "email": email})
}
