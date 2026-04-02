# 🍽️ Yetuk Kafe - Avtomatlashtirilgan Boshqaruv Tizimi

Ushbu loyiha kafe va yetkazib berish xizmatlarini to'liq avtomatlashtirish uchun mo'ljallangan. Tizim onlayn buyurtmalarni qabul qilish, oshxona va kurerlar faoliyatini muvofiqlashtirish hamda lokal printer orqali cheklarni avtomatik chop etish imkonini beradi.

## 🚀 Texnologiyalar

- **Backend**: Go (Golang) + Gin Framework
- **Frontend**: React + Vite + TailwindCSS
- **Database**: PostgreSQL 15
- **Virtualizatsiya**: Docker & Docker Compose
- **Real-time**: WebSockets

---

## 📂 Loyiha Tuzilishi

- `/backend`: Go tilida yozilgan API server.
- `/frontend`: Foydalanuvchilar va ma'murlar uchun veb-interfeys.
- `/bridge`: Windows uchun lokal printer "ko'prik" dasturi.
- `deploy.sh`: Serverda avtomatik yangilanish skripti.
- `docker-compose.yml`: Barcha xizmatlarni birga ishga tushirish qoidasi.

---

## 🛠️ O'rnatish va Ishga tushirish

### 1. Lokal (Docker orqali)
Loyihani o'z kompyuteringizda sinab ko'rish uchun:
```bash
docker-compose up --build
```
Loyiha `http://localhost:3001` (Frontend) va `http://localhost:8080` (Backend) manzillarida ishga tushadi.

### 2. Serverga joylash (Production)
Domen: `kafe.ruslandev.uz`
IP: `46.224.133.140`

Serverdagi `deploy.sh` skriptidan foydalaning:
```bash
chmod +x deploy.sh
./deploy.sh
```
Ushbu skript avtomatik tarzda Github'dan kodni oladi, qayta yig'adi va log yozib boradi.

---

## 🖨️ Avtomatik Printer Tizimi

Loyihaning eng muhim qismi - bu onlayn zakazlarni kafedagi printerdan avtomatik chiqarish.

### Qanday ishlaydi?
1. Saytdan zakaz tushadi -> Server uni qabul qiladi.
2. Server WebSocket orqali kafedagi **Bridge** dasturiga xabar yuboradi.
3. **Bridge** dasturi (Windowsda orqa fonda ishlaydi) xabarni oladi va lokal IP-dagi (`192.168.123.10`) printerga buyruq beradi.

### Bridge sozlamalari (Kafedagi PC uchun):
1. `bridge/main.go` ni Windows uchun yig'ing (`.exe`).
2. `.env` faylida quyidagilarni to'g'rilang:
   - `SERVER_WS_URL=ws://kafe.ruslandev.uz/api/ws`
   - `AUTH_TOKEN=sizning_doimiy_tokeningiz`
   - `PRINTER_IP=192.168.123.10`

---

## 🔒 Xavfsizlik va Tokenlar

Printer serverga ulanishi uchun doimiy (10 yillik) token ishlatiladi. Tokenni yangilash uchun:
```bash
go run backend/gen_token.go
```
Chiqqan tokenni `.env` faylidagi `AUTH_TOKEN` qismiga joylashtiring.

---

## 📝 Nginx Konfiguratsiyasi (Reverse Proxy)

Loyihani boshqa loyihalarga xalaqit bermasligi uchun serverdagi Nginx sozlamalariga `/nginx.conf.example` ni qo'shing. Bu domen orqali so'rovlarni Docker ichidagi `3001` va `8080` portlariga yo'naltiradi.

## 📈 Monitoring

Barcha deploy jarayonlarini `deploy.log` fayli orqali kuzatib borishingiz mumkin:
```bash
tail -f deploy.log
```

---

**Dasturchi**: [Ruslan Xusenov](https://github.com/Ruslan-Xusenov)
**Loyiha versiyasi**: 1.0.0 (Production Ready)
