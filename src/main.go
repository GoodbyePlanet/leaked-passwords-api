package main

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"leaked-passwords-api/src/db"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"leaked-passwords-api/src/api"
	"leaked-passwords-api/src/config"
	"leaked-passwords-api/src/service"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	database := db.Init("./tmp/badger")
	defer database.Close()
	logger.Info("Database initialized!")

	scheduler := service.NewScheduledDownload(database)
	c := scheduler.RunDownload()
	defer c.Stop()

	router := gin.Default()

	checkPasswordService := service.NewCheckPassword()
	badgerDebugService := service.NewBadgerHashReader(database)
	api.RegisterRoutes(router, checkPasswordService, badgerDebugService)

	PORT := config.LoadConfig().PORT
	server := &http.Server{
		Addr:    ":" + PORT,
		Handler: router,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Failed to start server: ", err)
			os.Exit(1)
		}
	}()
	logger.Info("Server started on port " + PORT)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutdown signal received, shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	} else {
		logger.Info("Server shut down gracefully!")
	}

	log.Println("Server and DB shut down finished!")
}
