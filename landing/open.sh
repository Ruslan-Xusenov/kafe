#!/bin/bash
# KafePlat Landing Page Launcher
# Bu faylni ikki marta bosing — brauzer avtomatik ochiladi

LANDING_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PORT=8765

# Eski serverni o'chirish
pkill -f "python3 -m http.server $PORT" 2>/dev/null
sleep 0.5

# Yangi server ishga tushirish
cd "$LANDING_DIR"
python3 -m http.server $PORT > /dev/null 2>&1 &
SERVER_PID=$!

echo "✅ Server ishga tushdi (PID: $SERVER_PID)"
echo "🌐 Sayt manzili: http://localhost:$PORT"

sleep 1

# Brauzerda ochish
if command -v xdg-open &>/dev/null; then
    xdg-open "http://localhost:$PORT"
elif command -v google-chrome &>/dev/null; then
    google-chrome "http://localhost:$PORT"
elif command -v firefox &>/dev/null; then
    firefox "http://localhost:$PORT"
fi

echo "🚀 KafePlat Landing sahifasi ochildi!"
echo "   Yopish uchun: kill $SERVER_PID"
