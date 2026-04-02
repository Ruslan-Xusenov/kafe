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

# 🛠️ Docker va Compose'ni avtomatik o'rnatish funksiyasi
install_docker_if_missing() {
    if ! command -v docker &> /dev/null; then
        log_info "Docker topilmadi. O'rnatish boshlanmoqda..."
        sudo apt-get update -y >> "$LOG_FILE" 2>&1
        sudo apt-get install -y ca-certificates curl gnupg >> "$LOG_FILE" 2>&1
        sudo install -m 0755 -d /etc/apt/keyrings >> "$LOG_FILE" 2>&1
        curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg >> "$LOG_FILE" 2>&1
        sudo chmod a+r /etc/apt/keyrings/docker.gpg >> "$LOG_FILE" 2>&1
        echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
        sudo apt-get update -y >> "$LOG_FILE" 2>&1
        sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin >> "$LOG_FILE" 2>&1
        log_info "Docker muvaffaqiyatli o'rnatildi."
    else
        log_info "Docker allaqachon o'rnatilgan."
    fi
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

# Docker o'rnatishni tekshirish
install_docker_if_missing

# Docker Compose buyrug'ini aniqlash (v2 yoki v1)
if docker compose version >/dev/null 2>&1; then
    DOCKER_COMPOSE="docker compose"
elif docker-compose version >/dev/null 2>&1; then
    DOCKER_COMPOSE="docker-compose"
else
    log_info "Docker Compose topilmadi. O'rnatishga harakat qilinmoqda..."
    sudo apt-get update -y >> "$LOG_FILE" 2>&1
    sudo apt-get install -y docker-compose-plugin >> "$LOG_FILE" 2>&1
    DOCKER_COMPOSE="docker compose"
fi

# Navigate to project directory
cd "$PROJECT_DIR" || { log_error "Papka topilmadi: $PROJECT_DIR"; exit 1; }

# 🔄 Smart Update: Mahalliy o'zgarishlarni vaqtinchalik saqlab turish
log_info "Mahalliy o'zgarishlarni tekshirish va stash qilish..."
HAS_CHANGES=$(git status --porcelain)
if [ -n "$HAS_CHANGES" ]; then
    log_info "Mahalliy o'zgarishlar topildi. Stash qilinmoqda..."
    git stash push -m "Auto-stash before deploy $(date)" >> "$LOG_FILE" 2>&1
    STASHED=true
else
    log_info "Mahalliy o'zgarishlar yo'q."
    STASHED=false
fi

# Update from GitHub
log_info "Github bilan sinxronizatsiya qilinmoqda..."
git remote set-url origin "$REPO_URL" >> "$LOG_FILE" 2>&1
if git pull origin main >> "$LOG_FILE" 2>&1; then
    log_info "Github'dan yangilanishlar muvaffaqiyatli olindi."
else
    log_error "Github bilan bog'liq xatolik yuz berdi! Stash tiklanishi mumkin."
    [ "$STASHED" = true ] && git stash pop >> "$LOG_FILE" 2>&1
    tail -n 10 "$LOG_FILE"
    exit 1
fi

# 🔄 Smart Update: O'zgarishlarni qayta tiklash
if [ "$STASHED" = true ]; then
    log_info "Mahalliy o'zgarishlar qayta tiklanmoqda (git stash pop)..."
    if git stash pop >> "$LOG_FILE" 2>&1; then
        log_info "O'zgarishlar muvaffaqiyatli saqlandi."
    else
        log_error "Stash pop paytida to'qnashuv (conflict) yuz berdi! Iltimos, qo'lda tekshiring."
    fi
fi

# Environment faylini tekshirish (Docker uchun .env shart)
if [ ! -f "backend/.env" ]; then
    log_info "backend/.env topilmadi. .env.example'dan nusxa olinmoqda..."
    cp backend/.env.example backend/.env
    log_info "backend/.env fayli yaratildi."
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
