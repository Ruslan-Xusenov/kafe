#!/bin/bash

# Configuration
PROJECT_DIR="$(pwd)"
LOG_FILE="${PROJECT_DIR}/deploy.log"
TOKEN_FILE="${PROJECT_DIR}/.github_token"

# Function to log messages
log_info() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - [INFO] - $1" >> "$LOG_FILE"
    echo "💡 [INFO]: $1"
}

log_error() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - [ERROR] - $1" >> "$LOG_FILE"
    echo "❌ [ERROR]: $1"
}

# 1. GitHub Tokenni tekshirish
if [ -f "$TOKEN_FILE" ]; then
    GITHUB_TOKEN=$(cat "$TOKEN_FILE" | tr -d '\r\n ')
else
    log_error "'.github_token' fayli topilmadi! Iltimos, serverda tokenni yarating."
    echo "Buyruq: echo 'ghp_sizning_tokeningiz' > .github_token"
    exit 1
fi

REPO_URL="https://Ruslan-Xusenov:${GITHUB_TOKEN}@github.com/Ruslan-Xusenov/kafe.git"

# Start Deployment
echo "------------------------------------------------"
log_info "Deployment jarayoni boshlandi..."

# Navigate to project directory
cd "$PROJECT_DIR" || { log_error "Papka topilmadi: $PROJECT_DIR"; exit 1; }

# Update from GitHub
log_info "Github bilan sinxronizatsiya qilinmoqda..."
git remote set-url origin "$REPO_URL" >> "$LOG_FILE" 2>&1
if git pull origin main >> "$LOG_FILE" 2>&1; then
    log_info "Github'dan yangilanishlar muvaffaqiyatli olindi."
else
    log_error "Github bilan aloqa uzildi. Tokenni tekshiring."
    tail -n 10 "$LOG_FILE"
    exit 1
fi

# Docker Compose buyrug'ini aniqlash (v2 yoki v1)
if docker compose version >/dev/null 2>&1; then
    DOCKER_COMPOSE="docker compose"
elif docker-compose version >/dev/null 2>&1; then
    DOCKER_COMPOSE="docker-compose"
else
    log_error "Docker Compose topilmadi! Iltimos, uni o'rnating."
    exit 1
fi

# Build and Restart Containers
log_info "Docker konteynerlari qayta yig'ilmoqda ($DOCKER_COMPOSE Build & Up)..."
if $DOCKER_COMPOSE up -d --build >> "$LOG_FILE" 2>&1; then
    log_info "Docker muvaffaqiyatli yangilandi."
else
    log_error "Docker bilan bog'liq xatolik yuz berdi!"
    echo "------------------------------------------------"
    echo "↓↓↓ Xatolik tafsilotlari (Logdan) ↓↓↓"
    tail -n 25 "$LOG_FILE"
    echo "------------------------------------------------"
    exit 1
fi

# Cleanup
log_info "Keraksiz Docker tasvirlari tozalanmoqda..."
docker image prune -f >> "$LOG_FILE" 2>&1

log_info "Deployment muvaffaqiyatli yakunlandi! 🚀"
echo "------------------------------------------------"
