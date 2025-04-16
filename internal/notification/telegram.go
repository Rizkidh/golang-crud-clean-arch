package notification

import (
	"net/http"
	"net/url"
	"os"
	"strings"
)

// SendTelegramMessage mengirim pesan ke Telegram bot
func SendTelegramMessage(message string) {
	go func() {
		token := os.Getenv("TELEGRAM_BOT_TOKEN")
		chatID := os.Getenv("TELEGRAM_CHAT_ID")

		if token == "" || chatID == "" {
			return // token atau chat ID tidak tersedia, skip
		}

		apiURL := "https://api.telegram.org/bot" + token + "/sendMessage"
		data := url.Values{}
		data.Set("chat_id", chatID)
		data.Set("text", message)

		_, _ = http.Post(apiURL, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	}()
}
