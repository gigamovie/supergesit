package main

import (
	"flag"
	"fmt"
	"os"

	"supergesit/internal/engine"
)

func main() {
	url := flag.String("url", "", "URL file")
	output := flag.String("o", "output.bin", "Nama file output")
	threads := flag.Int("n", 8, "Jumlah koneksi")

	flag.Parse()

	if *url == "" {
		fmt.Println("❌ URL wajib diisi")
		fmt.Println("Contoh:")
		fmt.Println("go run ./cmd/turbo -url https://example.com/file.bin -o file.bin")
		os.Exit(1)
	}

	fmt.Println("⚡ SuperGesit Downloader")
	fmt.Println("URL     :", *url)
	fmt.Println("Threads :", *threads)
	fmt.Println("Output  :", *output)

	err := engine.DownloadHTTP(*url, *output, *threads)
	if err != nil {
		fmt.Println("❌ Download gagal:", err)
		os.Exit(1)
	}

	fmt.Println("✅ Download selesai")
}
