#!/bin/bash

# --- Kafe Production Deployment Script (Final) ---
# IP: 46.224.133.140
# --- --------------------------------------- ---

echo "🚀 Kafe Production Deployment boshlandi..."

# ROOT papkani aniqlash
ROOT_DIR=$(pwd)

# 1. GitHubdan borini tortish va o'zgarishlarni aniqlash
git fetch origin
PREV_COMMIT=$(git rev-parse HEAD)
git reset --hard origin/main
NEW_COMMIT=$(git rev-parse HEAD)

CHANGED_FILES=$(git diff --name-only $PREV_COMMIT $NEW_COMMIT)

# --- 🛰 1. BACKEND API TEKSHIRUVI ---
if [ "$PREV_COMMIT" == "$NEW_COMMIT" ] || echo "$CHANGED_FILES" | grep -q "^backend/" || echo "$CHANGED_FILES" | grep -q "^frontend/"; then
    echo "🔄 Backend yoki Frontend o'zgardi! Yangilash boshlanmoqda..."
    
    # Port 8080 ni tozalash
    pkill -f "go run ./cmd/api/main.go" || true

    # Agar frontend o'zgargan bo'lsa - Build qilamiz
    if echo "$CHANGED_FILES" | grep -q "^frontend/"; then
        echo "🌐 Frontend build qilinmoqda..."
        cd "$ROOT_DIR/frontend" && npm install && npm run build
    fi

    echo "🚀 Backend (Sayt + API) yoqilmoqda..."
    cd "$ROOT_DIR/backend" && nohup go run ./cmd/api/main.go > backend.log 2>&1 &
    echo "✅ Backend yangilandi."
else
    echo "⏺ Backend va Frontendda o'zgarish yo'q."
fi

# --- 🤖 2. TELEGRAM BOT TEKSHIRUVI ---
if [ "$PREV_COMMIT" == "$NEW_COMMIT" ] || echo "$CHANGED_FILES" | grep -q "^telegram_bot/"; then
    echo "🔄 Telegram Bot yangilanmoqda..."
    # Bot jarayonini to'xtatish
    pkill -f "go run .$" || true
    pkill -f "go run telegram_bot" || true
    
    cd "$ROOT_DIR/telegram_bot" && nohup go run . > bot.log 2>&1 &
    echo "✅ Bot yangilandi."
else
    echo "⏺ Botda o'zgarish yo'q."
fi

echo "✨ PRODUCTION rejimi faol! Barcha tizimlar (Sayt, API, Bot) ishlamoqda."
echo "📜 Loglar: $ROOT_DIR/backend/backend.log, $ROOT_DIR/telegram_bot/bot.log"