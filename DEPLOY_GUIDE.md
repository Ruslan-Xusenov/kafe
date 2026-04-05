# 🍵 Kafe — Server Deploy Qo'llanmasi

> **Server:** `**.***.**.****` | **Domen:** `kafe.ruslandev.uz`  
> **Stack:** Go + React (Vite) + PostgreSQL + Telegram Bot | Docker Compose + Nginx

---

## Arxitektura

```
Internet
   │
   ▼
[Nginx :80/:443]  ← kafe.ruslandev.uz
   ├── /          → kafe-web-prod   (React, :3001)
   ├── /api       → kafe-api-prod   (Go API, :8080)
   ├── /api/ws    → kafe-api-prod   (WebSocket)
   └── /uploads   → kafe-api-prod   (Fayllar)

Docker Network: kafe-network
   ├── kafe-db-prod     (PostgreSQL)
   ├── kafe-api-prod    (Go Backend :8080)
   ├── kafe-web-prod    (Nginx+React :3001)
   └── kafe-bot-prod    (Telegram Bot)
```

---

## BOSQICH 1 — DNS Sozlash

Domen provayderingiz paneliga kiring va A-recordlar qo'shing:

| Nom | Tur | Qiymat | TTL |
|-----|-----|--------|-----|
| `kafe.ruslandev.uz` | A | `46.224.133.140` | 300 |
| `www.kafe.ruslandev.uz` | A | `46.224.133.140` | 300 |

> DNS tarqalishi 5-30 daqiqa vaqt oladi. Tekshirish: `ping kafe.ruslandev.uz`

---

## BOSQICH 2 — Fayllarni GitHub'ga Push Qilish

Lokal kompyuterda bajaring:

```bash
cd /home/kali/Desktop/projects/Django/Kafe

# .gitignore'ni tekshir — .env chiqqan bo'lmasin!
cat .gitignore | grep .env

git add docker-compose.yml nginx.conf server-deploy.sh
git add telegram_bot/Dockerfile telegram_bot/main.go
git add backend/Dockerfile .env.production
git commit -m "chore: production deployment config"
git push origin main
```

> DIQQAT: `backend/.env` va `.env.production` GitHub'ga chiqmasin!

---

## BOSQICH 3 — Serverga Ulanish

```bash
ssh root@46.224.133.140
```

---

## BOSQICH 4 — Loyihani Clone va Deploy

Serverda:

```bash
# Loyihani /opt/kafe ga clone qilish
git clone https://github.com/Ruslan-Xusenov/kafe /opt/kafe
cd /opt/kafe

# .env.production faylni qo'lda yaratish (GitHub'ga push qilinmagan)
cat > .env.production << 'EOF'
PORT=8080
GIN_MODE=release
DB_HOST=db
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=kafe2026
DB_NAME=kafe_db
JWT_SECRET=Yetuk_Kafe_2026
TELEGRAM_BOT_TOKEN=8263024052:AAES9kj_XBqqRlLbTFBLuY1z0qG8-FluRWc
SUPER_ADMIN_ID=8462446411
TELEGRAM_CHAT_ID=8462446411
PRINTER_ENABLED=false
API_URL=http://kafe.ruslandev.uz
EOF

# Deploy scriptni ishga tushirish
sudo bash server-deploy.sh
```

Script avtomatik bajaradi:
- Docker & Docker Compose o'rnatish
- Nginx o'rnatish va `kafe.ruslandev.uz` sozlash
- `.env.production` → `backend/.env` ga ko'chirish
- 4 ta containerni build qilib ishga tushirish

---

## BOSQICH 5 — SSL Sertifikat (HTTPS)

DNS tarqalgandan keyin serverda:

```bash
sudo certbot --nginx -d kafe.ruslandev.uz -d www.kafe.ruslandev.uz
```

So'ralganda:
- Email kiriting
- Terms of Service → `A`
- HTTP → HTTPS redirect → `2`

Yangilashni avtomatik qilish:
```bash
sudo systemctl enable certbot.timer
sudo certbot renew --dry-run
```

---

## Foydali Buyruqlar

```bash
# Containerlar holati
cd /opt/kafe && docker compose ps

# Loglar (real-time)
docker compose logs -f
docker compose logs -f backend
docker compose logs -f telegram-bot

# Yangi kod deploy
cd /opt/kafe
git pull origin main
docker compose down && docker compose up -d --build

# Database backup
docker exec kafe-db-prod pg_dump -U postgres kafe_db > backup_$(date +%Y%m%d).sql

# PostgreSQL ichiga kirish
docker exec -it kafe-db-prod psql -U postgres -d kafe_db
```

---

## Muammolar va Yechimlar

| Muammo | Sabab | Yechim |
|--------|-------|--------|
| Container ishlamayapti | Kod xatosi | `docker compose logs backend` |
| Nginx 502 Bad Gateway | Container to'xtagan | `docker compose restart backend` |
| DB ulanmaydi | `DB_HOST` noto'g'ri | `.env` da `DB_HOST=db` bo'lishi shart |
| Bot ishlamaydi | Token xato | `docker compose logs telegram-bot` |
| Port band | Boshqa jarayon | `ss -tlnp \| grep 8080` |

---

## Xavfsizlik Sozlamalari

```bash
ufw allow ssh
ufw allow 80/tcp
ufw allow 443/tcp
ufw deny 8080/tcp   # API tashqaridan yopiq
ufw deny 3001/tcp   # Frontend tashqaridan yopiq
ufw deny 5432/tcp   # DB faqat Docker ichida
ufw enable
```

---

## Deploy Tekshiruvi

| Test | URL | Natija |
|------|-----|--------|
| Frontend | `http://kafe.ruslandev.uz` | React sahifa |
| API | `http://kafe.ruslandev.uz/api/health` | `{"status":"ok"}` |
| HTTPS | `https://kafe.ruslandev.uz` | 🔒 SSL lock |
| WebSocket | Browser DevTools → Network → WS | Connected |
