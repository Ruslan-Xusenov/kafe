#!/bin/bash

# --- Kafe Smart Deployment Script ---
# IP: 46.224.133.140
# --- ---------------------------- ---

echo "🚀 Tizim tekshirilmoqda..."

# 1. GitHubdan borini tortish va o'zgarishlarni aniqlash
PREV_COMMIT=$(git rev-parse HEAD)
git pull origin main
NEW_COMMIT=$(git rev-parse HEAD)

if [ "$PREV_COMMIT" == "$NEW_COMMIT" ]; then
    echo "✅ Hammasi yangi holatda, o'zgarish yo'q."
    # Agar birinchi marta ishga tushirayotgan bo'lsangiz ham, davom etaveradi
fi

# O'zgargan fayllar ro'yxati
CHANGED_FILES=$(git diff --name-only $PREV_COMMIT $NEW_COMMIT)

# --- 🛰 1. BACKEND API TEKSHIRUVI ---
if echo "$CHANGED_FILES" | grep -q "^backend/"; then
    echo "🔄 Backend o'zgardi! API qayta ishga tushirilmoqda..."
    # Backend jarayonini to'xtatish
    pkill -f "go run ./cmd/api/main.go" || true
    # Orqa fonda boshlash
    cd backend && nohup go run ./cmd/api/main.go > backend.log 2>&1 &
    cd ..
    echo "✅ Backend yangilandi."
else
    echo "⏺ Backendda o'zgarish yo'q."
fi

# --- 🤖 2. TELEGRAM BOT TEKSHIRUVI ---
if echo "$CHANGED_FILES" | grep -q "^telegram_bot/"; then
    echo "🔄 Telegram Bot o'zgardi! Bot qayta ishga tushirilmoqda..."
    # Bot jarayonini to'xtatish
    pkill -f "go run telegram_bot" || true # Agar papka nomi bilan run qilsak
    # Yoki aniq 'telegram_bot/main.go' bo'lsa:
    pkill -f "go run ./main.go" || true
    
    cd telegram_bot && nohup go run . > bot.log 2>&1 &
    cd ..
    echo "✅ Bot yangilandi."
else
    echo "⏺ Botda o'zgarish yo'q."
fi

# --- 🌐 3. FRONTEND TEKSHIRUVI ---
if echo "$CHANGED_FILES" | grep -q "^frontend/"; then
    echo "🔄 Frontend o'zgardi! Sayt qayta ishga tushirilmoqda..."
    # Vite/Frontend jarayonini to'xtatish
    pkill -f "vite" || true
    
    cd frontend && nohup npm run dev -- --host > frontend.log 2>&1 &
    cd ..
    echo "✅ Frontend yangilandi."
else
    echo "⏺ Frontendda o'zgarish yo'q."
fi

echo "✨ Jarayon yakunlandi! Tizim sog'lom ishlamoqda."
echo "📜 Loglar: backend.log, bot.log, frontend.log"