package main

import (
	"log"
	"time"

	"github.com/joho/godotenv"
	"github.com/username/kafe-backend/internal/models"
	"github.com/username/kafe-backend/internal/service"
)

func main() {
	// 1. .env faylni yuklash
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env fayl topilmadi, standart sozlamalar ishlatiladi")
	}

	// 2. PrinterService-ni yaratish
	printer := service.NewPrinterService()
	if !printer.Enabled {
		log.Fatal("Xato: PRINTER_ENABLED=.env faylda 'true' bo'lishi kerak!")
	}

	log.Printf("Printerga ulanishga harakat qilinmoqda: %s:%s...", printer.IP, printer.Port)

	// 3. Test Buyurtma yaratish
	comment := "Iltimos, tezroq bo'lsin!"
	testOrder := &models.Order{
		ID:         999,
		Phone:      "+998 90 123 45 67",
		Address:    "Toshkent shahri, Chilonzor 1-kvartal",
		TotalPrice: 125000,
		Comment:    &comment,
		CreatedAt:  time.Now(),
		Items: []models.OrderItem{
			{
				ProductName: "Osh (Maxsus)",
				Quantity:    2,
				Price:       35000,
			},
			{
				ProductName: "Shashlik (Mol go'shti)",
				Quantity:    3,
				Price:       15000,
			},
			{
				ProductName: "Coca-Cola 1.5L",
				Quantity:    1,
				Price:       10000,
			},
		},
	}

	// 4. Chop etish buyrug'ini yuborish
	printer.PrintOrder(testOrder)

	log.Println("Test chek yuborildi! Iltimos, printerni tekshiring.")
}
