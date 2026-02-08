package main

import (
	"flag"
	"fmt"
	"os"

	"supergesit/internal/engine"
)

func main() {
	url := flag.String("url", "", "URL file yang akan didownload")
	output := flag.String("o", "", "Nama file output")
	threads := flag.Int("n", 16, "Jumlah koneksi (threads)")
	engineType := flag.String("engine", "http", "Engine: http")

	flag.Parse()

	if *url == "" {
		fmt.Println("❌ URL wajib diisi")
		fmt.Println("Contoh:")
		fmt.Println("go run ./cmd/turbo -url https://example.com/file.bin -n 8 -o file.bin")
		os.Exit(1)
	}

	fmt.Println("⚡ SuperGesit Downloader")
	fmt.Println("Engine  :", *engineType)
	fmt.Println("URL     :", *url)
	fmt.Println("Threads :", *threads)

	switch *engineType {
	case "http":
		err := engine.DownloadHTTP(*url, *output, *threads)
		if err != nil {
			fmt.Println("❌ Download gagal:", err)
			os.Exit(1)
		}
	default:
		fmt.Println("❌ Engine tidak dikenal:", *engineType)
		os.Exit(1)
	}

	fmt.Println("✅ Selesai tanpa error")
}
