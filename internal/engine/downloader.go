package engine

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

const userAgent = "SuperGesit/1.0 (+https://github.com/gigamovie/supergesit)"

func Download(url, output string, threads int, insecure bool) error {
	client := &http.Client{
		Timeout: 0,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			req.Header.Set("User-Agent", userAgent)
			return nil
		},
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure},
		},
	}

	// ===== HEAD =====
	headReq, _ := http.NewRequest("HEAD", url, nil)
	headReq.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(headReq)
	if err != nil {
		return err
	}
	resp.Body.Close()

	if resp.StatusCode >= 400 {
		return errors.New("HTTP error: " + resp.Status)
	}

	sizeStr := resp.Header.Get("Content-Length")
	if sizeStr == "" {
		fmt.Println("⚠️ Tidak ada Content-Length, fallback single")
		return singleDownload(client, url, output)
	}

	total, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil || total <= 0 {
		fmt.Println("⚠️ Content-Length tidak valid")
		return singleDownload(client, url, output)
	}

	fmt.Println("📦 Ukuran file:", total, "bytes")

	if resp.Header.Get("Accept-Ranges") != "bytes" || threads < 2 {
		fmt.Println("⚠️ Server tidak mendukung Range, fallback single")
		return singleDownload(client, url, output)
	}

	// ===== FILE PREP =====
	file, err := os.Create(output)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := file.Truncate(total); err != nil {
		return err
	}

	chunk := total / int64(threads)
	var wg sync.WaitGroup

	startTime := time.Now()

	for i := 0; i < threads; i++ {
		start := int64(i) * chunk
		end := start + chunk - 1
		if i == threads-1 {
			end = total - 1
		}

		wg.Add(1)
		go func(id int, s, e int64) {
			defer wg.Done()
			if err := downloadPart(client, url, file, s, e); err == nil {
				fmt.Println("⚡ Thread", id, "selesai")
			}
		}(i, start, end)
	}

	wg.Wait()
	fmt.Println("⏱️ Waktu:", time.Since(startTime))
	return nil
}

func downloadPart(client *http.Client, url string, file *os.File, start, end int64) error {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusPartialContent {
		return errors.New("server tidak balas 206")
	}

	buf := make([]byte, 128*1024)
	offset := start

	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			file.WriteAt(buf[:n], offset)
			offset += int64(n)
		}
		if err != nil {
			break
		}
	}
	return nil
}

func singleDownload(client *http.Client, url, output string) error {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.New("HTTP error: " + resp.Status)
	}

	out, err := os.Create(output)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
