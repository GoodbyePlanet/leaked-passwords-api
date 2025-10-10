package api

import (
	"leaked-passwords-api/src/models"
	"leaked-passwords-api/src/repository"
	"leaked-passwords-api/src/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	passwordService *service.PasswordService
	passwordsRepo   *repository.PasswordsRepository
}

func NewHandler(passwordService *service.PasswordService, passwordsRepo *repository.PasswordsRepository) *Handler {
	return &Handler{passwordService: passwordService, passwordsRepo: passwordsRepo}
}

func (handler *Handler) GetByHash(c *gin.Context) {
	passwordHash := c.Param("hash")

	if passwordHash == "" {
		respondWithError(c, http.StatusBadRequest, "Missing passwordHash param!")
		return
	}

	hashEntry, err := handler.passwordsRepo.GetByHash(passwordHash)

	if err != nil {
		respondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	if hashEntry == nil {
		respondWithError(c, http.StatusNotFound, "Hash not found")
		return
	}

	c.JSON(http.StatusOK, models.HashResponse{
		Hash:  hashEntry.Key,
		Count: hashEntry.Value,
	})
}

func (handler *Handler) GetAllHashes(c *gin.Context) {
	hashes, err := handler.passwordsRepo.GetAll()

	if err != nil {
		respondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, models.HashesListResponse{
		TotalReturned: len(hashes),
		Hashes:        hashes,
	})
}

func (handler *Handler) CheckPasswordHandler(c *gin.Context) {
	var request models.CheckPasswordRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondWithError(c, http.StatusBadRequest, "Invalid request")
		return
	}

	r := handler.passwordService.CheckPassword(request.Password)
	c.JSON(http.StatusOK, models.CheckPasswordResponse{
		PasswordHash: r.PasswordHash,
		BreachCount:  r.BreachCount,
		IsLeaked:     r.IsLeaked,
	})
}

func respondWithError(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{"error": message})
}
