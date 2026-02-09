package engine

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Menyamar sebagai Chrome agar tidak kena 403 Forbidden
const browserUA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36"

func Download(url, output string, threads int, insecure bool) error {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure},
		},
	}

	// 1. Dapatkan Ukuran File (Pakai GET + Range 0-0 agar lolos proteksi server)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", browserUA)
	req.Header.Set("Range", "bytes=0-0")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 && resp.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("server menolak akses (HTTP %d)", resp.StatusCode)
	}

	// Ambil ukuran total dari header Content-Range
	var total int64
	cr := resp.Header.Get("Content-Range")
	if cr != "" {
		parts := strings.Split(cr, "/")
		if len(parts) > 1 {
			total, _ = strconv.ParseInt(parts[1], 10, 64)
		}
	} else {
		total, _ = strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
	}

	if total <= 0 {
		fmt.Println("⚠️ Server tidak mendukung Range, mendownload single thread...")
		return singleDownload(client, url, output)
	}

	fmt.Printf("📦 File ditemukan: %.2f MB\n", float64(total)/(1024*1024))

	// 2. Siapkan file kosong (Pre-allocation)
	file, err := os.Create(output)
	if err != nil {
		return err
	}
	defer file.Close()
	file.Truncate(total)

	// 3. Jalankan Multi-thread
	var wg sync.WaitGroup
	chunkSize := total / int64(threads)
	startTime := time.Now()

	

	for i := 0; i < threads; i++ {
		start := int64(i) * chunkSize
		end := start + chunkSize - 1
		if i == threads-1 {
			end = total - 1
		}

		wg.Add(1)
		go func(id int, s, e int64) {
			defer wg.Done()
			// Retry 3x jika koneksi putus
			for retry := 0; retry < 3; retry++ {
				if err := downloadPart(client, url, file, s, e); err == nil {
					return
				}
				time.Sleep(time.Second)
			}
		}(i, start, end)
	}

	wg.Wait()
	fmt.Printf("⚡ Selesai! Kecepatan rata-rata: %.2f MB/s\n", (float64(total)/(1024*1024))/time.Since(startTime).Seconds())
	return nil
}

func downloadPart(client *http.Client, url string, file *os.File, start, end int64) error {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))
	req.Header.Set("User-Agent", browserUA)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Buffer 32KB agar RAM tetap enteng di Termux
	buf := make([]byte, 32*1024)
	curr := start
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			file.WriteAt(buf[:n], curr)
			curr += int64(n)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func singleDownload(client *http.Client, url, output string) error {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", browserUA)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, _ := os.Create(output)
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}
