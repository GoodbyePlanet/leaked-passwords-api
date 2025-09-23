package api

import (
	"leaked-passwords-api/src/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	passwordService *service.CheckPassword
}

func RegisterRoutes(r *gin.Engine, checkPassword *service.CheckPassword) {
	h := &Handler{passwordService: checkPassword}

	// Routes
	r.POST("/check", h.checkPasswordHandler)
}

func (h *Handler) checkPasswordHandler(c *gin.Context) {
	var req struct {
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	isPwned := h.passwordService.CheckPassword(req.Password)

	c.JSON(http.StatusOK, gin.H{
		"password": req.Password,
		"pwned":    isPwned,
	})
}
