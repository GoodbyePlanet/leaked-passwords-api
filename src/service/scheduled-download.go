package service

import (
	"github.com/robfig/cron/v3"
	"log/slog"
	"os"

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
	c := cron.New()

	// Run once on the first day of every month at midnight
	_, err := c.AddFunc("0 0 1 * *", func() {
		logger.Info("Starting cron job")

		// TODO: Add workers and prefixes as env variables
		downloader := NewHibpDownloader(10, 20, sd.passwordsRepo)
		downloader.DownloadAndSavePwnedPasswords()
	})

	if err != nil {
		logger.Error("An error occurred while running scheduler", err)
	}

	c.Start()

	go func() {
		logger.Info("Running download of service startup")
		downloader := NewHibpDownloader(10, 20, sd.passwordsRepo)
		downloader.DownloadAndSavePwnedPasswords()
	}()

	return c
}
