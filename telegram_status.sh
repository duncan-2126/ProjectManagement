#!/usr/bin/env bash

# Export env variables for the Telegram bot
export TELEGRAM_BOT_TOKEN="8310508597:AAHLqJC9rTkFI2EpAg6TzHz5u4gRca7RE7o"
export TELEGRAM_CHAT_ID="3664664734"

# Build the status message
status_msg=$(cat <<'EOF'
🚀 Sprint status (UTC) $(date -u '+%Y-%m-%d %H:%M')

✅ Telegram Bot: in_progress
✅ React UI: in_progress
✅ CI Pipeline: in_progress
✅ API Unit Test: in_progress
✅ API design: in_progress
✅ DB migrations: in_progress
❌ GitHub Integration: pending
❌ E2E Test Suite: pending
❌ Deployment Bot: pending
❌ PR review summary: pending
EOF
)

# Send via the helper binary (assumes it is named status_importer)
./status_importer <<< "$status_msg"
