package cmd

import (
    "bytes"
    "encoding/json"
    "fmt"
    "os"
    "net/http"
)

// sendTelegramMessage posts a simple text message to a Telegram bot.
// It expects TELEGRAM_BOT_TOKEN and TELEGRAM_CHAT_ID environment
// variables to be set.
func sendTelegramMessage(text string) error {
    token := os.Getenv("TELEGRAM_BOT_TOKEN")
    chatID := os.Getenv("TELEGRAM_CHAT_ID")
    if token == "" || chatID == "" {
        return fmt.Errorf("missing TELEGRAM_BOT_TOKEN or TELEGRAM_CHAT_ID env var")
    }
    payload := map[string]string{
        "chat_id": chatID,
        "text":    text,
    }
    body, _ := json.Marshal(payload)
    resp, err := http.Post(fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token), "application/json", bytes.NewReader(body))
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("telegram bot returned status %d", resp.StatusCode)
    }
    return nil
}
