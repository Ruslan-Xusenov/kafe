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

info "2. Konfiguratsiya tekshirilmoqda..."
if [ ! -f "$APP_DIR/backend/.env" ]; then
    if [ -f "$APP_DIR/.env.production" ]; then
        cp "$APP_DIR/.env.production" "$APP_DIR/backend/.env"
        log ".env.production dan backend/.env yaratildi."
    else
        warn ".env fayli topilmadi! O'zingiz qo'lda yarating."
    fi
fi

if ! command -v docker &> /dev/null; then
    info "Docker o'rnatilmoqda..."
    apt-get update -qq && apt-get install -y docker.io docker-compose-v2 -qq
fi

if [ ! -f "/etc/nginx/sites-enabled/kafe.ruslandev.uz" ]; then
    info "Nginx sozlanyapti..."
    cp "$APP_DIR/nginx.conf" /etc/nginx/sites-available/kafe.ruslandev.uz
    ln -sf /etc/nginx/sites-available/kafe.ruslandev.uz /etc/nginx/sites-enabled/kafe.ruslandev.uz
    rm -f /etc/nginx/sites-enabled/default
    nginx -t && systemctl reload nginx
fi

info "5. Docker konteynerlar build qilinmoqda..."
docker compose up -d --build --remove-orphans
log "Konteynerlar muvaffaqiyatli ishga tushdi."

info "6. Ma'lumotlar bazasini tekshirish (Auto Migration)..."
# Categories jadvali uchun
docker exec kafe-db-prod psql -U postgres -d kafe_db -c "
DO \$\$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='categories' AND column_name='is_user_controlled') THEN
        ALTER TABLE categories ADD COLUMN is_user_controlled BOOLEAN DEFAULT FALSE;
    END IF;
END \$\$;"

# Products jadvali uchun etishmayotgan barcha ustunlarni qo'shish
docker exec kafe-db-prod psql -U postgres -d kafe_db -c "
DO \$\$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='products' AND column_name='unit') THEN
        ALTER TABLE products ADD COLUMN unit VARCHAR(20) DEFAULT 'dona';
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='products' AND column_name='min_quantity') THEN
        ALTER TABLE products ADD COLUMN min_quantity DECIMAL(12, 3) DEFAULT 1.0;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='products' AND column_name='quantity_step') THEN
        ALTER TABLE products ADD COLUMN quantity_step DECIMAL(12, 3) DEFAULT 1.0;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='products' AND column_name='has_mandatory_container') THEN
        ALTER TABLE products ADD COLUMN has_mandatory_container BOOLEAN DEFAULT FALSE;
    END IF;
END \$\$;"

# Settings jadvali uchun
docker exec kafe-db-prod psql -U postgres -d kafe_db -c "
CREATE TABLE IF NOT EXISTS settings (
    key VARCHAR(50) PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
INSERT INTO settings (key, value) VALUES ('container_price', '1000') ON CONFLICT (key) DO NOTHING;
INSERT INTO settings (key, value) VALUES ('container_product_id', '7') ON CONFLICT (key) DO NOTHING;"

log "Baza migratsiyasi muvaffaqiyatli yakunlandi."

echo ""
echo "🚀 Tizim muvaffaqiyatli yangilandi!"
docker compose ps