#!/bin/bash
# ====================================================
#  Kafe Production Server Deploy Script v3.0 (Auto-Update)
#  Loyiha: kafe.ruslandev.uz
#  Ishlatish: sudo bash server-deploy.sh
# ====================================================

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log() { echo -e "${GREEN}[✓]${NC} $1"; }
info() { echo -e "${BLUE}[→]${NC} $1"; }
warn() { echo -e "${YELLOW}[!]${NC} $1"; }
error() { echo -e "${RED}[✗]${NC} $1"; exit 1; }

APP_DIR="/opt/kafe"

# 1. GitHub dan yangi kod olish
info "1. GitHub-dan kod yangilanmoqda..."
if [ -d "$APP_DIR/.git" ]; then
    cd "$APP_DIR"
    git fetch origin main
    git reset --hard origin/main
    log "Kod eng so'nggi versiyaga yangilandi."
else
    git clone https://github.com/Ruslan-Xusenov/kafe "$APP_DIR"
    cd "$APP_DIR"
    log "Loyiha birinchi marta clone qilindi."
fi

# 2. .env faylini tekshirish
info "2. Konfiguratsiya tekshirilmoqda..."
if [ ! -f "$APP_DIR/backend/.env" ]; then
    if [ -f "$APP_DIR/.env.production" ]; then
        cp "$APP_DIR/.env.production" "$APP_DIR/backend/.env"
        log ".env.production dan backend/.env yaratildi."
    else
        warn ".env fayli topilmadi! O'zingiz qo'lda yarating."
    fi
fi

# 3. Docker va Nginx o'rnatilganini tekshirish (faqat kerak bo'lsa)
if ! command -v docker &> /dev/null; then
    info "Docker o'rnatilmoqda..."
    apt-get update -qq && apt-get install -y docker.io docker-compose-v2 -qq
fi

# 4. Nginx sozlash (symlink yo'q bo'lsa)
if [ ! -f "/etc/nginx/sites-enabled/kafe.ruslandev.uz" ]; then
    info "Nginx sozlanyapti..."
    cp "$APP_DIR/nginx.conf" /etc/nginx/sites-available/kafe.ruslandev.uz
    ln -sf /etc/nginx/sites-available/kafe.ruslandev.uz /etc/nginx/sites-enabled/kafe.ruslandev.uz
    rm -f /etc/nginx/sites-enabled/default
    nginx -t && systemctl reload nginx
fi

# 5. Docker Konteynerlarni Build va Start qilish
info "5. Docker konteynerlar build qilinmoqda..."
docker compose up -d --build --remove-orphans
log "Konteynerlar muvaffaqiyatli ishga tushdi."

# 6. Ma'lumotlar bazasini avtomatik migratsiyalash (Yangi ustunlarni qo'shish)
info "6. Ma'lumotlar bazasini tekshirish (Auto Migration)..."
# is_user_controlled ustuni borligini tekshirish va yo'q bo'lsa qo'shish
docker exec kafe-db-prod psql -U postgres -d kafe_db -c "DO \$\$ 
BEGIN 
    BEGIN
        ALTER TABLE categories ADD COLUMN is_user_controlled BOOLEAN DEFAULT FALSE;
    EXCEPTION
        WHEN duplicate_column THEN RAISE NOTICE 'Column is_user_controlled already exists, skipping.';
    END;
END \$\$;" 2>/dev/null || true

log "Baza migratsiyasi tekshirildi."

echo ""
echo "🚀 Tizim muvaffaqiyatli yangilandi!"
docker compose ps