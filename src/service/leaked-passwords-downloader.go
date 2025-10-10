package service

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"leaked-passwords-api/src/repository"
)

const pwnedRangeApiUrl = "https://api.pwnedpasswords.com/range/%s"

type HibpDownloader struct {
	nWorkers     uint64
	nPrefixes    int
	hex5         <-chan string
	responseData chan []byte
	client       *http.Client
	repo         *repository.PasswordsRepository
	logger       *slog.Logger
}

func NewHibpDownloader(workers uint64, prefixes int, repo *repository.PasswordsRepository) *HibpDownloader {
	return &HibpDownloader{
		nWorkers:     workers,
		nPrefixes:    prefixes,
		responseData: make(chan []byte, 100),
		client:       &http.Client{Timeout: 10 * time.Second},
		repo:         repo,
		logger:       slog.New(slog.NewJSONHandler(os.Stdout, nil)),
	}
}

func (hd *HibpDownloader) DownloadAndSavePwnedPasswords() {
	if hd.nPrefixes > 0 {
		hd.hex5 = generateHex5Prefixes(hd.nPrefixes)
	} else {
		hd.hex5 = generateHex5Prefixes()
	}

	hd.logger.Info("Starting download", "workers", hd.nWorkers, "prefixes", hd.nPrefixes)
	var wg sync.WaitGroup
	for i := uint64(0); i < hd.nWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			hd.downloader()
		}()
	}

	go func() {
		defer hd.logger.Info("DB writer finished")

		for blob := range hd.responseData {
			hd.repo.Save(blob)
		}
	}()

	wg.Wait()
	close(hd.responseData) // ensure a DB writer finishes
	hd.logger.Info("Downloading and storing finished")
}

func (hd *HibpDownloader) downloader() {
	for hex5 := range hd.hex5 {
		const maxRetries = 5
		ok := false
		for attempt := 0; attempt < maxRetries && !ok; attempt++ {
			url := fmt.Sprintf(pwnedRangeApiUrl, hex5)
			req, _ := http.NewRequest(http.MethodGet, url, nil)
			req.Header.Set("User-Agent", "hibp-downloader-go")

			response, err := hd.client.Do(req)
			if err != nil {
				hd.logger.Warn("HTTP request failed",
					slog.String("prefix", hex5),
					slog.Int("attempt", attempt+1),
					slog.Any("error", err),
				)
				time.Sleep(time.Second * time.Duration(attempt+1))
				continue
			}

			if response.StatusCode != http.StatusOK {
				hd.logger.Warn("Unexpected status code",
					slog.String("prefix", hex5),
					slog.Int("status", response.StatusCode),
				)
				response.Body.Close()
				time.Sleep(time.Second * time.Duration(attempt+1))
				continue
			}

			responseBody, err := io.ReadAll(response.Body)
			response.Body.Close()

			if err != nil {
				hd.logger.Error("Failed to read response body",
					slog.String("prefix", hex5),
					slog.Any("error", err),
				)
				time.Sleep(time.Second * time.Duration(attempt+1))
				continue
			}

			hd.responseData <- hd.appendPrefix(hex5, responseBody)
			ok = true
		}
		if !ok {
			hd.logger.Error("Failed after retries", slog.String("prefix", hex5))
		}
	}
}

func (hd *HibpDownloader) appendPrefix(hex5 string, responseBody []byte) []byte {
	var b strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(string(responseBody)))
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			fmt.Fprintf(&b, "%s%s\r\n", hex5, strings.ToUpper(line))
		}
	}
	return []byte(b.String())
}

func generateHex5Prefixes(limit ...int) chan string {
	total := 0x100_000 // 0xFFFFF + 1 = 1,048,576
	if len(limit) > 0 {
		total = limit[0]
	}

	ch := make(chan string)

	go func() {
		for i := 0; i < total; i++ {
			ch <- fmt.Sprintf("%05X", i)
		}
		close(ch)
	}()

	return ch
}
