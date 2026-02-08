package engine

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
)

func DownloadHTTP(url, output string, parts int) error {
	// HEAD request untuk ambil ukuran file
	resp, err := http.Head(url)
	if err != nil {
		return err
	}

	size := resp.ContentLength
	if size <= 0 {
		return fmt.Errorf("tidak bisa mendapatkan ukuran file")
	}

	if output == "" {
		output = "output.bin"
	}

	file, err := os.Create(output)
	if err != nil {
		return err
	}
	defer file.Close()

	// Set ukuran file
	err = file.Truncate(size)
	if err != nil {
		return err
	}

	fmt.Println("📦 File size:", size, "bytes")
	fmt.Println("⚡ Mulai download dengan", parts, "koneksi")

	partSize := size / int64(parts)
	var wg sync.WaitGroup

	for i := 0; i < parts; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			start := int64(i) * partSize
			end := start + partSize - 1
			if i == parts-1 {
				end = size - 1
			}

			req, _ := http.NewRequest("GET", url, nil)
			req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				fmt.Println("❌ Error part", i, err)
				return
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
					if err != io.EOF {
						fmt.Println("❌ Read error:", err)
					}
					break
				}
			}

			fmt.Println("✅ Part", i, "selesai")
		}(i)
	}

	wg.Wait()
	fmt.Println("🎉 Download selesai")
	return nil
}
