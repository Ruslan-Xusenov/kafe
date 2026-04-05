#!/bin/bash
# ====================================================
#  Kafe Production Server Deploy Script
#  Server: 46.224.133.140 | kafe.ruslandev.uz
#  Ishlatish: sudo bash server-deploy.sh
# ====================================================

set -e  # Xato bo'lsa to'xta

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log() { echo -e "${GREEN}[✓]${NC} $1"; }
warn() { echo -e "${YELLOW}[!]${NC} $1"; }
error() { echo -e "${RED}[✗]${NC} $1"; exit 1; }
info() { echo -e "${BLUE}[→]${NC} $1"; }

echo ""
echo "╔══════════════════════════════════════════════╗"
echo "║    🍵 Kafe Deployment Server Script v2.0     ║"
echo "║       kafe.ruslandev.uz | 46.224.133.140     ║"
echo "╚══════════════════════════════════════════════╝"
echo ""

# Root ekanligini tekshir
if [ "$EUID" -ne 0 ]; then
  error "Scriptni sudo bilan ishga tushiring: sudo bash server-deploy.sh"
fi

# ── 1. System paketları ──────────────────────────────
info "1. Tizim paketlarini yangilash..."
apt-get update -qq && apt-get upgrade -y -qq
log "Tizim yangilandi"

# ── 2. Docker o'rnatish ──────────────────────────────
info "2. Docker o'rnatilmoqda..."
if command -v docker &> /dev/null; then
    log "Docker allaqachon o'rnatilgan: $(docker --version)"
else
    apt-get install -y -qq ca-certificates curl gnupg lsb-release
    install -m 0755 -d /etc/apt/keyrings
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
    chmod a+r /etc/apt/keyrings/docker.gpg
    echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] \
        https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" \
        | tee /etc/apt/sources.list.d/docker.list > /dev/null
    apt-get update -qq
    apt-get install -y -qq docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
    systemctl enable docker
    systemctl start docker
    log "Docker o'rnatildi: $(docker --version)"
fi

# ── 3. Nginx o'rnatish ───────────────────────────────
info "3. Nginx o'rnatilmoqda..."
if command -v nginx &> /dev/null; then
    log "Nginx allaqachon mavjud: $(nginx -v 2>&1)"
else
    apt-get install -y -qq nginx
    systemctl enable nginx
    log "Nginx o'rnatildi"
fi

# ── 4. Certbot (SSL) o'rnatish ───────────────────────
info "4. Certbot o'rnatilmoqda..."
if command -v certbot &> /dev/null; then
    log "Certbot allaqachon mavjud"
else
    apt-get install -y -qq certbot python3-certbot-nginx
    log "Certbot o'rnatildi"
fi

# ── 5. Loyihani clone/pull qilish ────────────────────
info "5. GitHub'dan loyiha yuklanmoqda..."
APP_DIR="/opt/kafe"

if [ -d "$APP_DIR/.git" ]; then
    warn "Loyiha allaqachon mavjud. Yangilanmoqda..."
    cd "$APP_DIR"
    git fetch origin main
    git reset --hard origin/main
    log "Kod yangilandi"
else
    git clone https://github.com/Ruslan-Xusenov/kafe "$APP_DIR"
    log "Loyiha clone qilindi"
fi

cd "$APP_DIR"

# ── 6. Production .env sozlash ───────────────────────
info "6. Production muhit konfiguratsiyasi sozlanmoqda..."
if [ -f "$APP_DIR/.env.production" ]; then
    cp "$APP_DIR/.env.production" "$APP_DIR/backend/.env"
    log ".env.production → backend/.env ga ko'chirildi"
else
    warn ".env.production topilmadi! backend/.env ni qo'lda tekshiring."
fi

# ── 7. Nginx sozlash ─────────────────────────────────
info "7. Nginx konfiguratsiyasi sozlanmoqda..."
cp "$APP_DIR/nginx.conf" /etc/nginx/sites-available/kafe.ruslandev.uz
ln -sf /etc/nginx/sites-available/kafe.ruslandev.uz /etc/nginx/sites-enabled/kafe.ruslandev.uz
rm -f /etc/nginx/sites-enabled/default

nginx -t && systemctl reload nginx
log "Nginx konfiguratsiyasi qo'llanildi"

# ── 8. Docker Compose bilan ishga tushirish ──────────
info "8. Docker containerlar ishga tushirilmoqda..."
cd "$APP_DIR"

# Eski containerlarni to'xtatish
docker compose down --remove-orphans 2>/dev/null || true

# Build va ishga tushirish
docker compose up -d --build

log "Docker containerlar ishga tushirildi"

# ── 9. Holat tekshiruvi ──────────────────────────────
info "9. Servislar holati tekshirilmoqda..."
sleep 8

echo ""
echo "📦 Docker Containers:"
docker compose ps

echo ""
echo "🔌 Port holati:"
ss -tlnp | grep -E ':(80|443|3001|8080|5432|5433)\s' || true

# ── 10. SSL sertifikat (Certbot) ─────────────────────
echo ""
warn "10. SSL sertifikat olish uchun quyidagi buyruqni ishga tushiring:"
echo ""
echo "  sudo certbot --nginx -d kafe.ruslandev.uz -d www.kafe.ruslandev.uz"
echo ""
echo "  (Avval DNS A-record 46.224.133.140 ga ko'rsatganligini tekshiring!)"
echo ""

echo "╔══════════════════════════════════════════════╗"
echo "║  ✅ Deploy muvaffaqiyatli bajarildi!         ║"
echo "║                                              ║"
echo "║  🌐 Sayt:   http://kafe.ruslandev.uz         ║"
echo "║  🔧 API:    http://kafe.ruslandev.uz/api     ║"
echo "║  📊 Loglar: docker compose logs -f           ║"
echo "╚══════════════════════════════════════════════╝"