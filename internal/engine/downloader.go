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
)

func Download(url, output string, threads int, insecure bool) error {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure},
		},
	}

	// ==== STEP 1: HEAD REQUEST ====
	resp, err := client.Head(url)
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

	total, _ := strconv.ParseInt(sizeStr, 10, 64)
	fmt.Println("📦 Ukuran file:", total, "bytes")

	// ==== STEP 2: RANGE SUPPORT ====
	if resp.Header.Get("Accept-Ranges") != "bytes" || threads <= 1 {
		fmt.Println("⚠️ Server tidak mendukung Range, fallback single")
		return singleDownload(client, url, output)
	}

	// ==== STEP 3: PREPARE FILE ====
	file, err := os.Create(output)
	if err != nil {
		return err
	}
	defer file.Close()

	file.Truncate(total)

	chunk := total / int64(threads)
	var wg sync.WaitGroup

	// ==== STEP 4: MULTIPART ====
	for i := 0; i < threads; i++ {
		start := int64(i) * chunk
		end := start + chunk - 1
		if i == threads-1 {
			end = total - 1
		}

		wg.Add(1)
		go func(id int, s, e int64) {
			defer wg.Done()
			err := downloadPart(client, url, file, s, e)
			if err == nil {
				fmt.Println("⚡ Thread", id, "selesai")
			}
		}(i, start, end)
	}

	wg.Wait()
	return nil
}

func downloadPart(client *http.Client, url string, file *os.File, start, end int64) error {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusPartialContent {
		return errors.New("server tidak balas 206")
	}

	buf := make([]byte, 64*1024)
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
	resp, err := client.Get(url)
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
