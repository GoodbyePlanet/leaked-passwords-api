package service

import (
	"crypto/sha1"
	"encoding/hex"
	"leaked-passwords-api/src/repository"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

type CheckPassword struct {
	badgerRepo *repository.BadgerRepository
	logger     *slog.Logger
}

type CheckPasswordResponse struct {
	IsLeaked     bool
	PasswordHash string
	BreachCount  uint64
}

func NewCheckPassword(badgerRepo *repository.BadgerRepository) *CheckPassword {
	return &CheckPassword{
		badgerRepo: badgerRepo,
		logger:     slog.New(slog.NewJSONHandler(os.Stdout, nil)),
	}
}

func (checkPassword *CheckPassword) CheckPassword(password string) *CheckPasswordResponse {
	checkPassword.logger.Info("Checking password", "password", password)
	hashBytes := sha1.Sum([]byte(password))
	hashHex := strings.ToUpper(hex.EncodeToString(hashBytes[:]))
	checkPassword.logger.Info("Hashed password", "hash", hashHex)

	hash, err := checkPassword.badgerRepo.GetByHash(hashHex)
	if err != nil {
		return response(false, hashHex, "0")
	}

	if hash == nil {
		return response(false, hashHex, "0")
	}

	return response(true, hashHex, hash.Value)
}

func response(isLeaked bool, hash string, count string) *CheckPasswordResponse {
	c, err := strconv.ParseUint(count, 10, 64)
	if err != nil {
		return nil
	}
	return &CheckPasswordResponse{IsLeaked: isLeaked, PasswordHash: hash, BreachCount: c}
}
