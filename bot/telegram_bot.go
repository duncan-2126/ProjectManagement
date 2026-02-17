package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
)

// telegramRequest defines the payload for the Telegram sendMessage API.
type telegramRequest struct {
    ChatID string `json:"chat_id"`
    Text   string `json:"text"`
    ParseMode string `json:"parse_mode,omitempty"`
}

func main() {
    token := os.Getenv("TELEGRAM_TOKEN")
    if token == "" {
        log.Fatalf("TELEGRAM_TOKEN not set")
    }
    chatID := os.Getenv("TELEGRAM_CHAT_ID")
    if chatID == "" {
        log.Fatalf("TELEGRAM_CHAT_ID not set")
    }

    message := `Hello @duncan2126_bot!  ✅ *Feature‑Request README* processed.  ✓ Web‑GUI tests now passing (go test ./...).  ✏️ Planned next steps: 1️⃣ Unit‑testing all remaining API endpoints. 2️⃣ Adding a minimal React bundle. 3️⃣ Improving CI workflow & linting.`

    payload := telegramRequest{ChatID: chatID, Text: message, ParseMode: "Markdown"}
    body, err := json.Marshal(payload)
    if err != nil {
        log.Fatalf("failed to marshal json: %v", err)
    }

    url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
    resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
    if err != nil {
        log.Fatalf("request failed: %v", err)
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        log.Fatalf("telegram returned non-200: %s", resp.Status)
    }
    fmt.Println("Message sent!")
}
