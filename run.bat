@echo off
setlocal
echo 🚀 Kafe va Yetkazib berish platformasini ishga tushirish (Windows)...

:: Check for Docker
docker-compose --version >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ Xatolik: docker-compose o'rnatilmagan. Iltimos, Docker Desktop o'rnating.
    pause
    exit /b 1
)

:: Check for .env file
if not exist backend\.env (
    echo ⚠️  backend\.env fayli topilmadi. .env.example dan nusxa olinmoqda...
    copy backend\.env.example backend\.env
)

:: Start services
echo 📦 Konteynerlar qurilmoqda va ishga tushirilmoqda...
docker-compose up -d --build

echo ✅ Tizim muvaffaqiyatli ishga tushdi!
echo 🌐 Frontend: http://localhost
echo ⚙️  Backend API: http://localhost:8080
echo 📊 MB (PostgreSQL): localhost:5432

docker-compose logs -f
pause
