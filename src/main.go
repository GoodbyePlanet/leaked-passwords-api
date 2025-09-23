package main

import (
	"github.com/gin-gonic/gin"
	"log/slog"
	"os"

	"leaked-passwords-api/src/api"
	"leaked-passwords-api/src/config"
	"leaked-passwords-api/src/service"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	router := gin.Default()

	checkPasswordService := service.NewCheckPassword()

	api.RegisterRoutes(router, checkPasswordService)

	err := router.Run(":" + config.LoadConfig().PORT)
	if err != nil {
		logger.Error("Failed to start server: ", err, "")
		return
	}
}
