package service

import (
	"github.com/dgraph-io/badger/v4"
	"github.com/robfig/cron/v3"
	"log/slog"
	"os"
)

type ScheduledDownload struct {
	db *badger.DB
}

func NewScheduledDownload(db *badger.DB) *ScheduledDownload {
	return &ScheduledDownload{db: db}
}

func (sd *ScheduledDownload) RunDownload() *cron.Cron {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	c := cron.New()

	// Run once on the first day of every month at midnight
	_, err := c.AddFunc("0 0 1 * *", func() {
		logger.Info("Starting cron job")

		// TODO: Add workers and prefixes as env variables
		downloader := NewHibpDownloader(10, 20, sd.db)
		downloader.DownloadAndSavePwnedPasswords()
	})

	if err != nil {
		logger.Error("An error occurred while running scheduler", err)
	}

	c.Start()

	go func() {
		logger.Info("Running download of service startup")
		downloader := NewHibpDownloader(10, 20, sd.db)
		downloader.DownloadAndSavePwnedPasswords()
	}()

	return c
}
