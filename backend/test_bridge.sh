#!/bin/bash

echo "🚀 Avtomatlashtirilgan Bridge Testini boshlaymiz..."

# Clean up any previous runs
fuser -k 8080/tcp > /dev/null 2>&1

# 1. Backendni orqa fonda yoqish
echo "1. Serverni yoqyapman..."
go run cmd/api/main.go > backend_test.log 2>&1 &
BACKEND_PID=$!
sleep 5 # Server yonishini kutamiz

# 2. Token yaratish (Mijoz uchun)
echo "2. Test mijozini login qilyapman..."
# Avval mijozni ro'yxatdan o'tkazamiz (agar bo'lmasa)
curl -s -X POST http://localhost:8080/api/auth/register \
-H "Content-Type: application/json" \
-d '{"full_name": "Test User", "phone": "+998901112233", "password": "password123", "role": "customer"}' > /dev/null

# Login qilamiz
LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8080/api/auth/login \
-H "Content-Type: application/json" \
-d '{"phone": "+998901112233", "password": "password123"}')

# Manually extract token since jq might not be available
CUSTOMER_TOKEN=$(echo $LOGIN_RESPONSE | sed -n 's/.*"token":"\([^"]*\)".*/\1/p')

if [ -z "$CUSTOMER_TOKEN" ]; then
    echo "❌ Xato: Token olib bo'lmadi. Backend ishlayotganini tekshiring."
    kill $BACKEND_PID > /dev/null 2>&1
    exit 1
fi

# 3. Printer Ko'prik (Bridge) dasturini yoqish
echo "3. Ko'prik dasturini ulayapman..."
export SERVER_WS_URL=ws://localhost:8080/api/ws
export AUTH_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjIwOTA1MDc5MDEsInJvbGUiOiJwcmludGVyIiwidXNlcl9pZCI6OTk5fQ.dauo2T2LmzMe6Layjz9mpUHdeK9NvfWeplgjBMmiv3c"
export PRINTER_IP=127.0.0.1
go run ./bridge/main.go > bridge_test.log 2>&1 &
BRIDGE_PID=$!
sleep 4

# 4. Yangi Zakaz berish (Simulyatsiya)
echo "4. Saytdan yangi zakaz berayapman..."
# Find a valid product ID
PRODUCT_ID=$(curl -s http://localhost:8080/api/catalog/products | sed -n 's/.*"id":\([0-9]*\).*/\1/p' | head -n 1)
if [ -z "$PRODUCT_ID" ]; then PRODUCT_ID=2; fi

curl -s -X POST http://localhost:8080/api/orders/ \
-H "Content-Type: application/json" \
-H "Authorization: Bearer $CUSTOMER_TOKEN" \
-d "{
  \"address\": \"Test kocha, 12-uy\",
  \"phone\": \"+998901112233\",
  \"items\": [{\"product_id\": $PRODUCT_ID, \"quantity\": 1}]
}" > /dev/null

echo "✅ Zakaz yuborildi! Tekshirilmoqda..."
sleep 3

# 5. Natijani tekshirish
if grep -q "New Order Received" bridge_test.log; then
    echo "✨ MUVAFFAQIYAT! Ko'prik dasturi zakazni qabul qildi."
    grep "New Order Received" bridge_test.log
else
    echo "⚠️ Diqqat: Zakaz qabul qilingani logda ko'rinmadi. 'bridge_test.log' ni tekshiring."
fi

# Tozalash
echo "--- TEST YAKUNLANDI ---"
kill $BACKEND_PID > /dev/null 2>&1
kill $BRIDGE_PID > /dev/null 2>&1
echo "Jarayonlar to'xtatildi."
