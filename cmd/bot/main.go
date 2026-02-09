package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"supergesit/internal/engine" // Sesuaikan dengan module path kamu
	tele "gopkg.in/telebot.v3"
)

func main() {
	pref := tele.Settings{
		Token:  "TOKEN_BOT_KAMU_DISINI",
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("🤖 Bot SuperGesit sedang standby...")

	b.Handle(tele.OnText, func(c tele.Context) error {
		msg := c.Message()
		var url string

		// Cek apakah ada URL di dalam teks (baik pesan biasa atau terusan)
		for _, entity := range msg.Entities {
			if entity.Type == tele.EntityURL {
				url = msg.Text[entity.Offset : entity.Offset+entity.Length]
			}
		}

		if url == "" {
			return c.Send("❌ Kirimkan pesan yang berisi link video/file.")
		}

		c.Send(fmt.Sprintf("🔍 Link Terdeteksi: %s\n⚡ Memulai download super cepat...", url))

		// Nama file output (ambil dari ujung URL atau random)
		output := "downloaded_file_" + time.Now().Format("150405") + ".bin"
		
		// Panggil engine SuperGesit (n=16 threads)
		err := engine.Download(url, output, 16, true)
		if err != nil {
			return c.Send("❌ Gagal download: " + err.Error())
		}

		// Kirim info file berhasil diunduh ke server bot
		return c.Send(fmt.Sprintf("✅ Berhasil diunduh ke server!\nOutput: %s\nUkuran file tersimpan di storage server.", output))
	})

	b.Start()
}
