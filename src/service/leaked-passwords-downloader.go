package service

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const apiEndpoint = "https://api.pwnedpasswords.com/range/%s"

type HibpDownloader struct {
	file         string
	nWorkers     uint64
	fp           *os.File
	hex5         <-chan string
	responseData chan []byte
	quit         chan bool
}

func Download(file string, parallelism uint64, overwrite bool) {
	var err error
	hd := &HibpDownloader{
		file:         file,
		nWorkers:     parallelism,
		responseData: make(chan []byte, 100),
		quit:         make(chan bool),
	}

	hd.hex5 = hex5generator()

	if _, err := os.Stat(hd.file); err == nil && !overwrite {
		panic(fmt.Errorf("file `%s` already exists", hd.file))
	}

	hd.fp, err = os.OpenFile(hd.file, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		panic(fmt.Errorf("failed to open file `%s`: %w", hd.file, err))
	}
	defer hd.fp.Close()

	fmt.Printf("Downloading SHA1 hashes with %d workers\n\n", hd.nWorkers)

	var ww sync.WaitGroup
	ww.Add(1)
	go func() {
		hd.writer(hd.responseData, hd.quit)
		ww.Done()
	}()

	var wg sync.WaitGroup
	for i := uint64(0); i < hd.nWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			hd.downloader()
		}()
	}

	wg.Wait()

	time.Sleep(time.Second * 2) // allow in-flight responseData
	hd.quit <- true
	ww.Wait()
}

func (hd *HibpDownloader) downloader() {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	for hex5 := range hd.hex5 {
		const maxRetries = 5
		ok := false
		for attempt := 0; attempt < maxRetries && !ok; attempt++ {
			url := fmt.Sprintf(apiEndpoint, hex5)
			req, _ := http.NewRequest(http.MethodGet, url, nil)
			req.Header.Set("User-Agent", "hibp-downloader/1.0")

			response, err := client.Do(req)
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

func (hd *HibpDownloader) writer(responseData chan []byte, quit chan bool) {
	for {
		select {
		case blob := <-responseData:
			if len(blob) == 0 {
				continue
			}
			if _, err := hd.fp.Write(blob); err != nil {
				panic(err)
			}
		case <-quit:
			return
		}
	}
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
