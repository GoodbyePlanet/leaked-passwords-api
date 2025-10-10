package service

import (
	"crypto/sha1"
	"encoding/hex"
	"leaked-passwords-api/src/models"
	"leaked-passwords-api/src/repository"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

type PasswordService struct {
	passwordRepository *repository.PasswordsRepository
	logger             *slog.Logger
}

func NewPasswordService(passwordRepository *repository.PasswordsRepository) *PasswordService {
	return &PasswordService{
		passwordRepository: passwordRepository,
		logger:             slog.New(slog.NewJSONHandler(os.Stdout, nil)),
	}
}

func (passwordService *PasswordService) CheckPassword(password string) *models.CheckPasswordResponse {
	passwordService.logger.Info("Checking password", "password", password)
	hashBytes := sha1.Sum([]byte(password))
	hashHex := strings.ToUpper(hex.EncodeToString(hashBytes[:]))
	passwordService.logger.Info("Hashed password", "hash", hashHex)

	hash, err := passwordService.passwordRepository.GetByHash(hashHex)
	if err != nil {
		return response(false, hashHex, "0")
	}

	if hash == nil {
		return response(false, hashHex, "0")
	}

	return response(true, hashHex, hash.Value)
}

func response(isLeaked bool, hash string, count string) *models.CheckPasswordResponse {
	c, err := strconv.ParseUint(count, 10, 64)
	if err != nil {
		return nil
	}
	return &models.CheckPasswordResponse{
		IsLeaked:     isLeaked,
		PasswordHash: hash,
		BreachCount:  c,
	}
}
