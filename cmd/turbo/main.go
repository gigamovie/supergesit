package main

import (
	"flag"
	"fmt"
	"os"
	"supergesit/internal/engine"
)

func main() {
	url := flag.String("url", "", "URL Video/File")
	out := flag.String("o", "hasil_gesit.mp4", "Nama File Output")
	n := flag.Int("n", 16, "Jumlah Thread (Saran: 16)")
	insecure := flag.Bool("k", true, "Abaikan error SSL (Insecure)")

	flag.Parse()

	if *url == "" {
		fmt.Println("❌ Mana link-nya woy? Gunakan: turbo -url LINK")
		os.Exit(1)
	}

	fmt.Println("🚀 Menyalakan Mesin SuperGesit...")
	err := engine.Download(*url, *out, *n, *insecure)
	if err != nil {
		fmt.Printf("💥 Mesin Meledak: %v\n", err)
		os.Exit(1)
	}
}
