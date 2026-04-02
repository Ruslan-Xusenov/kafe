# 🖨️ Printerni Avtomatik Ishlatish Bo‘yicha Qo‘llanma

Ushbu qo‘llanma **Kafe Avtomatizatsiya Platformasida** chek printerini sozlash va undan avtomatik foydalanishni tushuntiradi.

## 1. Printer Sozlamalari (.env)

Loyihaning `backend/.env` faylida printer uchun quyidagi parametrlar mavjud. Printeringiz IP manzili va porti to‘g‘ri ekanligiga ishonch hosil qiling:

```env
# Printer Sozlamalari
PRINTER_IP=192.168.123.10
PRINTER_PORT=9100
PRINTER_ENABLED=true
```

> [!IMPORTANT]
> Printeringiz va backend serveringiz bitta lokal tarmoqda (LAN) bo‘lishi shart. Agar server IP manzili `192.168.123` bilan boshlanmagan bo‘lsa, printerga ulana olmasligi mumkin.

## 2. Printer va Serverni Tekshirish

Printer va server bog‘langanligini bilish uchun terminalda `ping` buyrug‘ini ishlating:

```bash
ping 192.168.123.10
```

Agar javob (reply) kelsa, ulanish tayyor. Agar `Request timeout` bo‘lsa, printer kabelini yoki Wi-Fi tarmog‘ini tekshiring.

## 3. Avtomatik Chop Etish Jarayoni

Tizim quyidagi ketma-ketlikda ishlaydi:
1. **Mijoz buyurtma beradi**: Savatchadagi mahsulotlarni tasdiqlaydi.
2. **Order saqlanadi**: Backend ma'lumotlar bazasida yangi buyurtma yaratadi.
3. **Printer Service ishga tushadi**:
   - Printer bilan TCP ulanish o‘rnatiladi.
   - Printerga **2 marta ovoz chiqarish (Beep)** buyrug‘i yuboriladi (Oshxona xodimlarini ham xabardor qilish uchun).
   - Chek formati (Buyurtma raqami, mahsulotlar, jami narx) yuboriladi.
   - **Avtomatik kesish (Cut)** buyrug‘i ijro etiladi.

## 4. Chop Etilgan Chek Formati

Chek 80mm termal qog‘oz uchun optimallashtirilgan (bir qatorda 48 ta belgi):
- **Header**: Restoran nomi (KAFE) va buyurtma raqami (Katta shriftda).
- **Mijoz malumotlari**: Telefon raqami, Manzil va Mijoz izohi (agar bo‘lsa).
- **Mahsulotlar**: Jadval ko‘rinishida (Nomi, Soni, Narxi, Jami).
- **Footer**: "Xaridingiz uchun rahmat!" yozuvi va 5 qator bo‘sh joy (kesishdan oldin).

## 5. Xatolar va Muammolarni Hal Qilish

| Muammo | Sababi | Yechimi |
| :--- | :--- | :--- |
| **Printer umuman chiqarmayapti** | IP manzil noto‘g‘ri yoki printer o‘chirilgan. | `.env` dagi IP-ni tekshiring va printerni yoqing. |
| **Yozuvlar g‘alati chiqyapti** | Printer kod sahifasi (Code Page) mos kelmayapti. | Tizimda transliteratsiya yoqilgan (`o'`, `g'`), bu barcha printerlarda to‘g‘ri o‘qiladi. |
| **Qog‘oz kesilmayapti** | Printerda avtomatik kesish (Auto-Cutter) yo‘q yoki o‘chirilgan. | Printeringizning DIP switch sozlamalarini selftest chekiga qarab tekshiring. |
| **Oshxonada eshitilmayapti** | Beep buyrug‘i printeringiz tomonidan qo‘llab-quvvatlanmaydi. | Ba'zi printerlarda beep buyrug‘i ESC/POS o‘rniga boshqa kodda bo‘ladi. |

## 6. Layoutni O‘zgartirish (Dasturchilar uchun)

Restoran nomini yoki chek ko‘rinishini o‘zgartirish uchun `backend/internal/service/printer_service.go` faylidagi `PrintOrder` funksiyasini tahrirlang.

---

> [!TIP]
> Agar printeringiz 58mm bo‘lsa, `printer_service.go` dagi `48` belgilik cheklovlarni `32` ga kamaytirib chiqishingiz kerak. Hozirda sizning selftest malumotlaringizga ko‘ra **72mm/48chars** sozlamalari o‘rnatilgan.
