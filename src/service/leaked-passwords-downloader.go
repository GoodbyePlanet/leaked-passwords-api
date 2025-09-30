package service

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

const apiEndpoint = "https://api.pwnedpasswords.com/range/%s"

type HibpDownloader struct {
	nWorkers     uint64
	hex5         <-chan string
	responseData chan []byte
	client       *http.Client
}

func Download(parallelism uint64) {
	hd := &HibpDownloader{
		nWorkers:     parallelism,
		responseData: make(chan []byte, 100),
		client:       &http.Client{Timeout: 10 * time.Second},
	}

	hd.hex5 = hex5generator()
	fmt.Printf("Downloading SHA1 hashes with %d workers\n\n", hd.nWorkers)

	var wg sync.WaitGroup
	for i := uint64(0); i < hd.nWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			hd.downloader()
		}()
	}

	go func() {
		wg.Wait()
		close(hd.responseData)
	}()

	for blob := range hd.responseData {
		fmt.Printf("Got %d bytes of data\n", len(blob))
		// TODO: store into Badger
	}
}

func (hd *HibpDownloader) downloader() {
	for hex5 := range hd.hex5 {
		const maxRetries = 5
		ok := false
		for attempt := 0; attempt < maxRetries && !ok; attempt++ {
			url := fmt.Sprintf(apiEndpoint, hex5)
			req, _ := http.NewRequest(http.MethodGet, url, nil)
			req.Header.Set("User-Agent", "hibp-downloader/1.0")

			response, err := hd.client.Do(req)
			if err != nil {
				time.Sleep(time.Second * time.Duration(attempt+1))
				continue
			}
			if response.StatusCode != http.StatusOK {
				response.Body.Close()
				time.Sleep(time.Second * time.Duration(attempt+1))
				continue
			}

			responseBody, err := io.ReadAll(response.Body)
			response.Body.Close()
			if err != nil {
				time.Sleep(time.Second * time.Duration(attempt+1))
				continue
			}

			hd.responseData <- hd.applyHex5Prefix(hex5, responseBody)
			ok = true
		}
		if !ok {
			fmt.Printf("failed after retries: %s\n", hex5)
		}
	}
}

func (hd *HibpDownloader) applyHex5Prefix(hex5 string, responseBody []byte) []byte {
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

func hex5generator() chan string {
	const total = 0x100_000 // 0xFFFFF + 1 = 1,048,576
	ch := make(chan string)

	go func() {
		for i := 0; i < total; i++ {
			ch <- fmt.Sprintf("%05X", i)
		}
		close(ch)
	}()

	return ch
}
