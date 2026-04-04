#!/bin/bash

# 🚀 Kafe & Bot Universal Production Deployer (Master v4.0)
# Bu skript Sayt, API, Bot va Printerni yagona logikaga birlashtiradi.

ROOT_DIR=$(pwd)
BACKEND_DIR="$ROOT_DIR/backend"
BOT_DIR="$ROOT_DIR/telegram_bot"
FRONTEND_DIR="$ROOT_DIR/frontend"

echo "🧹 1. Serverni 'Ghost' (zombi) jarayonlardan tozalash..."
killall -9 main 2>/dev/null || true
pkill -9 -f "go run" 2>/dev/null || true
fuser -k 8080/tcp 2>/dev/null || true
sleep 2

echo "📥 2. Eng so'nggi kodlarni yuklash (Hard Force)..."
git fetch origin main
git reset --hard origin/main
git pull origin main

echo "📦 3. Frontendni build qilish (Production Mode)..."
if [ -d "$FRONTEND_DIR" ]; then
    cd "$FRONTEND_DIR"
    npm install --silent
    npm run build
fi

echo "🚀 4. Backend (Site + API) ishga tushirilmoqda..."
cd "$BACKEND_DIR"
# .env dagi DB_HOST ni localhost ekanligini ta'minlash
sed -i 's/DB_HOST=db/DB_HOST=localhost/g' .env 2>/dev/null || true

# Papkalarni tekshirish
mkdir -p uploads

# Backendni qat'iy (nohup) yoqish
nohup go run ./cmd/api/main.go > backend.log 2>&1 &
echo "✅ Backend (Port 8080) yoqildi."

echo "🤖 5. Telegram Bot ishga tushirilmoqda..."
cd "$BOT_DIR"
nohup go run . > bot.log 2>&1 &
echo "✅ Bot yoqildi."

echo "✨ FINAL: Tizim Production rejimida!"
echo "📍 Sayt: kafe.ruslandev.uz"
echo "📍 Loglar: backend/backend.log, telegram_bot/bot.log"

# Diagnostic Check
sleep 3
netstat -tpln | grep 8080