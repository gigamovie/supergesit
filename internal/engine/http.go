package engine

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
)

func DownloadMultiPart(url, output string, threads int, insecure bool) error {
	client := &http.Client{}
	if insecure {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	// HEAD request → ambil ukuran file
	resp, err := client.Head(url)
	if err != nil {
		return err
	}
	resp.Body.Close()

	size := resp.ContentLength
	if size <= 0 {
		return fmt.Errorf("server tidak mendukung Content-Length")
	}

	fmt.Println("📦 Ukuran file:", size, "bytes")

	partSize := size / int64(threads)

	file, err := os.Create(output)
	if err != nil {
		return err
	}
	defer file.Close()

	err = file.Truncate(size)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	errCh := make(chan error, threads)

	for i := 0; i < threads; i++ {
		wg.Add(1)

		start := int64(i) * partSize
		end := start + partSize - 1
		if i == threads-1 {
			end = size - 1
		}

		go func(start, end int64) {
			defer wg.Done()

			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				errCh <- err
				return
			}
			req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))

			resp, err := client.Do(req)
			if err != nil {
				errCh <- err
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusOK {
				errCh <- fmt.Errorf("server tidak mendukung range")
				return
			}

			buf := make([]byte, 32*1024)
			offset := start

			for {
				n, err := resp.Body.Read(buf)
				if n > 0 {
					_, werr := file.WriteAt(buf[:n], offset)
					if werr != nil {
						errCh <- werr
						return
					}
					offset += int64(n)
				}
				if err != nil {
					if err == io.EOF {
						break
					}
					errCh <- err
					return
				}
			}
		}(start, end)
	}

	wg.Wait()
	close(errCh)

	if len(errCh) > 0 {
		return <-errCh
	}

	return nil
}
