package api

import (
	"leaked-passwords-api/src/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	passwordService *service.CheckPassword
	badgerService   *service.BadgerHashReader
}

func RegisterRoutes(router *gin.Engine, checkPassword *service.CheckPassword, badgerDebugService *service.BadgerHashReader) {
	handler := &Handler{passwordService: checkPassword, badgerService: badgerDebugService}

	router.POST("/check", handler.checkPasswordHandler)
	router.GET("/internal/debug/byHash/:hash", handler.getByHash)
	router.GET("/internal/debug/getAll", handler.getAllHashes)
}

func (handler *Handler) getByHash(context *gin.Context) {
	passwordHash := context.Param("hash")

	if passwordHash == "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Missing passwordHash param!"})
		return
	}

	hashEntry, err := handler.badgerService.GetByHash(passwordHash)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if hashEntry == nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "hash not found"})
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"hash":  hashEntry.Key,
		"count": hashEntry.Value,
	})
}

func (handler *Handler) getAllHashes(context *gin.Context) {
	hashes, err := handler.badgerService.GetAll()

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"total_returned": len(hashes),
		"hashes":         hashes,
	})
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
