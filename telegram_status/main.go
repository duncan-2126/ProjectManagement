package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "time"
)

func send(message string) error {
    token := os.Getenv("TELEGRAM_BOT_TOKEN")
    chatID := os.Getenv("TELEGRAM_CHAT_ID")
    if token == "" || chatID == "" {
        return fmt.Errorf("missing env vars")
    }
    payload := map[string]string{
        "chat_id": chatID,
        "text":    message,
    }
    body, _ := json.Marshal(payload)
    resp, err := http.Post(fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token),
        "application/json", bytes.NewReader(body))
    if err != nil {
        return err
    }
    resp.Body.Close()
    return nil
}

func buildMessage() string {
    // Append timestamp for debugging
    ts := time.Now().Format("2006-01-02 15:04:05 MST")
    return fmt.Sprintf(`🚀 Sprint status (%s)

✅ Telegram Bot: in_progress
✅ React UI: in_progress
✅ CI Pipeline: in_progress
✅ API Unit Test: in_progress
✅ API design: in_progress
✅ DB migrations: in_progress
❌ GitHub Integration: pending
❌ E2E Test Suite: pending
❌ Deployment Bot: pending
❌ PR review summary: pending`, ts)
}

func main() {
    // Initial send
    msg := buildMessage()
    if err := send(msg); err != nil {
        fmt.Println("send error:", err)
    }
    ticker := time.NewTicker(10 * time.Minute)
    defer ticker.Stop()
    for range ticker.C {
        msg = buildMessage()
        if err := send(msg); err != nil {
            fmt.Println("send error:", err)
        }
    }
}
