package main

import (
	"fmt"
	"log"
	"path"
	"strings"
	"time"

	"supergesit/internal/engine" // Pastikan import sesuai nama module di go.mod
	tele "gopkg.in/telebot.v3"
)

func main() {
	// 1. Inisialisasi Bot
	pref := tele.Settings{
		Token:  "8503364188:AAFCbVtSulyr2ifm5uc4BdsQ9qmQJetGCeI", 
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Println("⚡ Bot SuperGesit Berhasil Berjalan di Termux!")

	// 2. Handler untuk pesan teks (termasuk link terusan)
	b.Handle(tele.OnText, func(c tele.Context) error {
		txt := c.Text()
		
		// Deteksi apakah ada link (sederhana)
		if !strings.HasPrefix(txt, "http") {
			return c.Send("❌ Kirimkan link video/file yang valid (dimulai dengan http/https).")
		}

		// Tentukan nama file dari ujung URL
		fileName := path.Base(txt)
		if strings.Contains(fileName, "?") {
			fileName = strings.Split(fileName, "?")[0]
		}
		
		c.Send(fmt.Sprintf("🚀 Link diterima!\n📦 Nama File: %s\n⚡ Sedang mengunduh dengan 16 thread...", fileName))

		// 3. Panggil Mesin SuperGesit
		start := time.Now()
		err := engine.Download(txt, fileName, 16, true)
		
		if err != nil {
			return c.Send("❌ Gagal: " + err.Error())
		}

		durasi := time.Since(start).Round(time.Second)
		return c.Send(fmt.Sprintf("✅ DOWNLOAD SELESAI!\n⏱️ Waktu: %v\n📍 Lokasi: %s", durasi, fileName))
	})

	b.Start()
}
