package main

import (
	"flag"
	"fmt"
	"os"

	"supergesit/internal/engine"
)

func main() {
	url := flag.String("url", "", "Download URL")
	out := flag.String("o", "output.bin", "Output file")
	n := flag.Int("n", 4, "Threads")
	insecure := flag.Bool("insecure", false, "Skip TLS verify")

	flag.Parse()

	if *url == "" {
		fmt.Println("URL wajib diisi")
		os.Exit(1)
	}

	fmt.Println("⚡ SuperGesit Downloader")
	fmt.Println("URL     :", *url)
	fmt.Println("Threads :", *n)
	fmt.Println("Output  :", *out)

	err := engine.Download(*url, *out, *n, *insecure)
	if err != nil {
		fmt.Println("❌ Download gagal:", err)
		os.Exit(1)
	}

	fmt.Println("✅ Download selesai")
}
