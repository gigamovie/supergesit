package engine

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

func Download(url, output string, threads int, insecure bool) error {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure},
		},
	}

	// ==== CHECK RANGE SUPPORT ====
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Range", "bytes=0-0")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusPartialContent {
		fmt.Println("⚠️ Server tidak mendukung Range, fallback single download")
		return singleDownload(client, url, output)
	}

	cr := resp.Header.Get("Content-Range")
	if cr == "" {
		return errors.New("Content-Range tidak ada")
	}

	// bytes 0-0/12345
	parts := strings.Split(cr, "/")
	total, _ := strconv.ParseInt(parts[1], 10, 64)

	fmt.Println("📦 Ukuran file:", total, "bytes")

	file, err := os.Create(output)
	if err != nil {
		return err
	}
	defer file.Close()

	file.Truncate(total)

	chunk := total / int64(threads)
	var wg sync.WaitGroup

	for i := 0; i < threads; i++ {
		start := int64(i) * chunk
		end := start + chunk - 1
		if i == threads-1 {
			end = total - 1
		}

		wg.Add(1)
		go func(id int, s, e int64) {
			defer wg.Done()
			downloadPart(client, url, file, s, e)
			fmt.Println("⚡ Thread", id, "selesai")
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

	buf := make([]byte, 32*1024)
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

	buf := make([]byte, 32*1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			out.Write(buf[:n])
		}
		if err != nil {
			break
		}
	}
	return nil
}

