package service

import (
	"leaked-passwords-api/src/config"
	"log/slog"
	"os"
	"strconv"

	"github.com/robfig/cron/v3"

	"leaked-passwords-api/src/repository"
)

type ScheduledDownload struct {
	passwordsRepo *repository.PasswordsRepository
}

func NewScheduledDownload(passwordsRepo *repository.PasswordsRepository) *ScheduledDownload {
	return &ScheduledDownload{passwordsRepo: passwordsRepo}
}

func (sd *ScheduledDownload) RunDownload() *cron.Cron {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	env := config.LoadConfig()
	workers, _ := strconv.ParseUint(env.WORKERS, 10, 64)
	prefixes, _ := strconv.Atoi(env.PREFIXES)
	c := cron.New()

	// Run once on the first day of every month at midnight
	_, err := c.AddFunc("0 0 1 * *", func() {
		logger.Info("Starting cron job")

		downloader := NewHibpDownloader(workers, prefixes, sd.passwordsRepo)
		downloader.DownloadAndSavePwnedPasswords()
	})

	if err != nil {
		logger.Error("An error occurred while running scheduler", err)
	}

	c.Start()

	go func() {
		logger.Info("Running download on service startup")
		downloader := NewHibpDownloader(workers, prefixes, sd.passwordsRepo)
		downloader.DownloadAndSavePwnedPasswords()
	}()

	return c
}
