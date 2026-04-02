#!/bin/bash

# Cafe Automation Platform Startup Script (Linux/macOS)

echo "🚀 Kafe va Yetkazib berish platformasini ishga tushirish..."

# Check dependencies
if ! [ -x "$(command -v docker-compose)" ]; then
  echo "❌ Xatolik: docker-compose o'rnatilmagan."
  exit 1
fi

# Check for .env file
if [ ! -f backend/.env ]; then
    echo "⚠️  backend/.env fayli topilmadi. .env.example dan nusxa olinmoqda..."
    cp backend/.env.example backend/.env
fi

# Start services
echo "📦 Konteynerlar qurilmoqda va ishga tushirilmoqda..."
docker-compose up -d --build

echo "✅ Tizim muvaffaqiyatli ishga tushdi!"
echo "🌐 Frontend: http://localhost"
echo "⚙️  Backend API: http://localhost:8080"
echo "📊 MB (PostgreSQL): localhost:5432"

docker-compose logs -f
