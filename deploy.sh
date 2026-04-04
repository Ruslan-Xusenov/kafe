#!/bin/bash

# --- Kafe Production Deployment Script ---
# IP: 46.224.133.140
# --- --------------------------------- ---

echo "🚀 Kafe Production Deployment boshlandi..."

# 1. GitHubdan borini tortish va o'zgarishlarni aniqlash
PREV_COMMIT=$(git rev-parse HEAD)
git pull origin main
NEW_COMMIT=$(git rev-parse HEAD)

CHANGED_FILES=$(git diff --name-only $PREV_COMMIT $NEW_COMMIT)

# --- 🛰 1. BACKEND API TEKSHIRUVI ---
# Eslatma: Backend endi Frontend build (dist) papkasini ham serve qiladi
if [ "$PREV_COMMIT" == "$NEW_COMMIT" ] || echo "$CHANGED_FILES" | grep -q "^backend/" || echo "$CHANGED_FILES" | grep -q "^frontend/"; then
    echo "🔄 Backend yoki Frontend o'zgardi! Yangilash boshlanmoqda..."
    
    # Port 8080 ni tozalash
    pkill -f "go run ./cmd/api/main.go" || true

    # Agar frontend o'zgargan bo'lsa - Build qilamiz
    if echo "$CHANGED_FILES" | grep -q "^frontend/"; then
        echo "🌐 Frontend build qilinmoqda..."
        cd frontend && npm install && npm run build && cd ..
        pkill -f "vite" || true
    fi

    echo "🚀 Backend (Sayt + API) yoqilmoqda..."
    cd backend && nohup go run ./cmd/api/main.go > backend.log 2>&1 &
    cd ..
else
    echo "⏺ Backend va Frontendda o'zgarish yo'q."
fi

# --- 🤖 2. TELEGRAM BOT TEKSHIRUVI ---
if [ "$PREV_COMMIT" == "$NEW_COMMIT" ] || echo "$CHANGED_FILES" | grep -q "^telegram_bot/"; then
    echo "🔄 Telegram Bot yangilanmoqda..."
    # Bot jarayonini to'xtatish
    pkill -f "go run .$" || true
    pkill -f "go run telegram_bot" || true
    
    cd telegram_bot && nohup go run . > bot.log 2>&1 &
    cd ..
    echo "✅ Bot yangilandi."
else
    echo "⏺ Botda o'zgarish yo'q."
fi

echo "✨ PRODUCTION rejimi faol! Barcha tizimlar (Sayt, API, Bot) ishlamoqda."
echo "📜 Loglar: backend/backend.log, telegram_bot/bot.log"