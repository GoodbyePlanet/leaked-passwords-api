package api

import (
	"leaked-passwords-api/src/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	passwordService *service.CheckPassword
}

func RegisterRoutes(router *gin.Engine, checkPassword *service.CheckPassword) {
	handler := &Handler{passwordService: checkPassword}

	router.POST("/check", handler.checkPasswordHandler)
}

func (handler *Handler) checkPasswordHandler(context *gin.Context) {
	var request struct {
		Password string `json:"password"`
	}
	if err := context.ShouldBindJSON(&request); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	isPwned := handler.passwordService.CheckPassword(request.Password)

	context.JSON(http.StatusOK, gin.H{
		"password": request.Password,
		"pwned":    isPwned,
	})
}
