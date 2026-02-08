package main

import (
	"flag"
	"log"
	"turbo-downloader/internal/engine"
)

func main() {
	url := flag.String("url", "", "Download URL")
	out := flag.String("o", "file.bin", "Output file")
	parts := flag.Int("n", 16, "Connections")
	flag.Parse()

	if *url == "" {
		log.Fatal("URL wajib diisi")
	}

	log.Println("Turbo Download started...")
	engine.Download(*url, *out, *parts)
	log.Println("Done!")
}

