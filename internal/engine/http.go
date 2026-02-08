package engine

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
)

func Download(url, output string, parts int) error {
	resp, err := http.Head(url)
	if err != nil {
		return err
	}

	size := resp.ContentLength
	file, err := os.Create(output)
	if err != nil {
		return err
	}
	defer file.Close()

	file.Truncate(size)

	var wg sync.WaitGroup
	partSize := size / int64(parts)

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

			resp, _ := http.DefaultClient.Do(req)
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
		}(i)
	}
	wg.Wait()
	return nil
}

