package engine

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Gunakan User Agent Chrome agar tidak dianggap bot oleh server
const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36"

func Download(url, output string, threads int, insecure bool) error {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure},
		// Optimasi koneksi untuk banyak thread
		MaxIdleConns:        threads,
		IdleConnTimeout:     90 * time.Second,
		DisableKeepAlives:   false,
	}
	client := &http.Client{Transport: transport}

	// 1. Dapatkan Ukuran File (Gunakan GET Range 0-0 sebagai ganti HEAD)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Range", "bytes=0-0")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 && resp.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("server menolak akses: %s", resp.Status)
	}

	// Ambil ukuran total dari header Content-Range atau Content-Length
	var total int64
	contentRange := resp.Header.Get("Content-Range")
	if contentRange != "" {
		parts := strings.Split(contentRange, "/")
		if len(parts) > 1 {
			total, _ = strconv.ParseInt(parts[1], 10, 64)
		}
	} else {
		sizeStr := resp.Header.Get("Content-Length")
		total, _ = strconv.ParseInt(sizeStr, 10, 64)
	}

	if total <= 0 {
		fmt.Println("⚠️ Server tidak memberikan ukuran file, download mode single...")
		return singleDownload(client, url, output)
	}

	fmt.Printf("📦 Ukuran: %.2f MB | Threads: %d\n", float64(total)/(1024*1024), threads)

	// 2. Siapkan File Kosong (Pre-allocation)
	file, err := os.Create(output)
	if err != nil {
		return err
	}
	defer file.Close()
	file.Truncate(total)

	// 3. Eksekusi Download Paralel
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
			// Sistem Retry: Coba 3 kali jika gagal di tengah jalan
			for retry := 0; retry < 3; retry++ {
				if err := downloadPart(client, url, file, s, e); err == nil {
					return
				}
				time.Sleep(2 * time.Second)
			}
			fmt.Printf("❌ Thread %d gagal setelah 3 percobaan\n", id)
		}(i, start, end)
	}

	wg.Wait()
	fmt.Printf("⚡ Berhasil! Waktu: %v\n", time.Since(startTime))
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

	// Gunakan buffer agar hemat RAM (Penting untuk Termux)
	buf := make([]byte, 64*1024)
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
	req.Header.Set("User-Agent", userAgent)
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
