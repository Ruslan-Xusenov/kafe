#!/bin/bash

# Configuration
PROJECT_DIR="$(pwd)"
GITHUB_TOKEN="YOUR_GITHUB_TOKEN_HERE"
REPO_URL="https://Ruslan-Xusenov:${GITHUB_TOKEN}@github.com/Ruslan-Xusenov/kafe.git"
LOG_FILE="${PROJECT_DIR}/deploy.log"

# Function to log messages
log_info() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - [INFO] - $1" >> "$LOG_FILE"
    echo "💡 [INFO]: $1"
}

log_error() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - [ERROR] - $1" >> "$LOG_FILE"
    echo "❌ [ERROR]: $1"
}

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
    exit 1
fi

# Build and Restart Containers
log_info "Docker konteynerlari qayta yig'ilmoqda (Build & Up)..."
if docker-compose up -d --build >> "$LOG_FILE" 2>&1; then
    log_info "Docker muvaffaqiyatli yangilandi."
else
    log_error "Docker bilan bog'liq xatolik yuz berdi."
    exit 1
fi

# Cleanup
log_info "Keraksiz Docker tasvirlari tozalanmoqda..."
docker image prune -f >> "$LOG_FILE" 2>&1

log_info "Deployment muvaffaqiyatli yakunlandi! 🚀"
echo "------------------------------------------------"
