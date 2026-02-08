package main

import (
	"flag"
	"fmt"
	"os"

	"supergesit/internal/engine"
)

func main() {
	url := flag.String("url", "", "URL file")
	out := flag.String("o", "output.bin", "Output file")
	flag.Parse()

	if *url == "" {
		fmt.Println("❌ URL wajib diisi")
		os.Exit(1)
	}

	fmt.Println("⬇️ Downloading:", *url)

	err := engine.Download(*url, *out)
	if err != nil {
		fmt.Println("❌ Error:", err)
		os.Exit(1)
	}

	fmt.Println("✅ Selesai:", *out)
}
