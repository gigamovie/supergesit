package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	url := flag.String("url", "", "URL file atau video")
	output := flag.String("o", "", "Nama file output")
	threads := flag.Int("n", 16, "Jumlah koneksi (default 16)")
	engine := flag.String("engine", "http", "Engine: http | aria2 | ytdlp")

	flag.Parse()

	if *url == "" {
		fmt.Println("❌ URL wajib diisi")
		flag.Usage()
		os.Exit(1)
	}

	fmt.Println("⚡ SuperGesit Downloader")
	fmt.Println("Engine :", *engine)
	fmt.Println("URL    :", *url)
	fmt.Println("Output :", *output)
	fmt.Println("Threads:", *threads)

	// STEP SELANJUTNYA: panggil engine
	fmt.Println("🚧 Engine belum diaktifkan (STEP 2)")
}
