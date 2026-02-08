package main

import (
	"flag"
	"fmt"
	"os"

	"supergesit/internal/engine"
)

func main() {
	url := flag.String("url", "", "URL file yang akan diunduh")
	output := flag.String("o", "output.bin", "Nama file output")
	threads := flag.Int("n", 8, "Jumlah thread (belum dipakai)")
	insecure := flag.Bool("insecure", false, "Lewati verifikasi TLS (HANYA UNTUK PRIBADI)")

	flag.Parse()

	if *url == "" {
		fmt.Println("❌ URL wajib diisi")
		os.Exit(1)
	}

	fmt.Println("⚡ SuperGesit Downloader")
	fmt.Println("URL     :", *url)
	fmt.Println("Threads :", *threads)
	fmt.Println("Output  :", *output)
	if *insecure {
		fmt.Println("⚠️  TLS  : INSECURE MODE")
	}

	err := engine.DownloadHTTP(*url, *output, *insecure)
	if err != nil {
		fmt.Println("❌ Download gagal:", err)
		os.Exit(1)
	}

	fmt.Println("✅ Download selesai")
}
