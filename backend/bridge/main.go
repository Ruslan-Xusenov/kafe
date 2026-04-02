package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

// Simplified models for the bridge
type Order struct {
	ID         int         `json:"id"`
	Phone      string      `json:"phone"`
	Address    string      `json:"address"`
	Comment    *string     `json:"comment"`
	TotalPrice float64     `json:"total_price"`
	CreatedAt  time.Time   `json:"created_at"`
	Items      []OrderItem `json:"items"`
}

type OrderItem struct {
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
	Comment     *string `json:"comment"`
}

var (
	ESC_INIT      = []byte{0x1B, 0x40}           // ESC @
	ALIGN_LEFT    = []byte{0x1B, 0x61, 0x00}      // ESC a 0
	ALIGN_CENTER  = []byte{0x1B, 0x61, 0x01}      // ESC a 1
	ALIGN_RIGHT   = []byte{0x1B, 0x61, 0x02}      // ESC a 2
	FONT_NORMAL   = []byte{0x1D, 0x21, 0x00}      // GS ! 0
	FONT_DOUBLE_H = []byte{0x1D, 0x21, 0x01}      // GS ! 1
	FONT_DOUBLE_W = []byte{0x1D, 0x21, 0x10}      // GS ! 16
	FONT_BIG      = []byte{0x1D, 0x21, 0x11}      // GS ! 17 (Double width + height)
	PAPER_CUT     = []byte{0x1D, 0x56, 0x42, 0x00} // GS V B 0
	BEEP          = []byte{0x1B, 0x42, 0x02, 0x02} // ESC B n t (Beep 2 times)
)

func main() {
	log.Println("--- Printer Bridge Started ---")
	godotenv.Load(".env") // Optional local .env for the bridge

	serverURL := os.Getenv("SERVER_WS_URL")
	token := os.Getenv("AUTH_TOKEN")
	printerIP := os.Getenv("PRINTER_IP")
	if printerIP == "" {
		printerIP = "192.168.123.10"
	}
	printerPort := os.Getenv("PRINTER_PORT")
	if printerPort == "" {
		printerPort = "9100"
	}

	if serverURL == "" || token == "" {
		log.Fatal("ERROR: SERVER_WS_URL and AUTH_TOKEN are required in environment variables")
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	for {
		log.Printf("Connecting to %s...", serverURL)
		header := http.Header{}
		header.Add("Authorization", "Bearer "+token)

		c, _, err := websocket.DefaultDialer.Dial(serverURL, header)
		if err != nil {
			log.Printf("Dial Error: %v. Retrying in 5s...", err)
			time.Sleep(5 * time.Second)
			continue
		}
		log.Println("Connected to server successfully!")

		done := make(chan struct{})
		go func() {
			defer close(done)
			for {
				_, message, err := c.ReadMessage()
				if err != nil {
					log.Println("Read Error:", err)
					return
				}
				handleMessage(message, printerIP, printerPort)
			}
		}()

		select {
		case <-done:
			log.Println("Connection lost. Reconnecting in 5s...")
			time.Sleep(5 * time.Second)
		case <-interrupt:
			log.Println("Interrupted. Closing connection...")
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("Close Error:", err)
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}

func handleMessage(message []byte, ip, port string) {
	var payload struct {
		Type  string `json:"type"`
		Order *Order `json:"order"`
	}
	if err := json.Unmarshal(message, &payload); err != nil {
		log.Println("Unmarshal Error:", err)
		return
	}

	if payload.Type == "new_order" && payload.Order != nil {
		log.Printf("New Order Received: #%d. Printing...", payload.Order.ID)
		printOrder(payload.Order, ip, port)
	}
}

func printOrder(order *Order, ip, port string) {
	address := net.JoinHostPort(ip, port)
	
	// Conflict handling: Retry if busy (another software printing)
	var conn net.Conn
	var err error
	for i := 0; i < 5; i++ {
		conn, err = net.DialTimeout("tcp", address, 3*time.Second)
		if err == nil {
			break
		}
		log.Printf("Printer busy or unreachable, retrying in 3s... (attempt %d/5)", i+1)
		time.Sleep(3 * time.Second)
	}

	if err != nil {
		log.Printf("Printer Error after 5 attempts: %v", err)
		return
	}
	defer conn.Close()

	// ESC/POS Formatting
	conn.Write(ESC_INIT)
	conn.Write(BEEP)

	conn.Write(ALIGN_CENTER)
	conn.Write(FONT_BIG)
	conn.Write([]byte("ONLAYN BUYURTMA\n"))
	conn.Write(FONT_NORMAL)
	conn.Write([]byte("------------------------------------------------\n"))
	
	conn.Write(FONT_DOUBLE_H)
	conn.Write([]byte(fmt.Sprintf("BUYURTMA #%d\n", order.ID)))
	conn.Write(FONT_NORMAL)
	
	createdAt := order.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	conn.Write([]byte(fmt.Sprintf("Sana: %s\n", createdAt.Format("02.01.2006 15:04"))))
	conn.Write([]byte("------------------------------------------------\n"))

	conn.Write(ALIGN_LEFT)
	conn.Write([]byte(fmt.Sprintf("Mijoz: %s\n", order.Phone)))
	conn.Write([]byte(fmt.Sprintf("Manzil: %s\n", transliterate(order.Address))))
	if order.Comment != nil && *order.Comment != "" {
		conn.Write([]byte(fmt.Sprintf("Izoh: %s\n", transliterate(*order.Comment))))
	}
	conn.Write([]byte("------------------------------------------------\n"))

	conn.Write([]byte("Mahsulot               Soni   Narxi      Jami\n"))
	for _, item := range order.Items {
		name := transliterate(item.ProductName)
		if len(name) > 22 {
			name = name[:19] + "..."
		}
		
		line := fmt.Sprintf("%-22s %-6d %-10.0f %-10.0f\n", 
			name, item.Quantity, item.Price, item.Price*float64(item.Quantity))
		conn.Write([]byte(line))
		
		if item.Comment != nil && *item.Comment != "" {
			conn.Write([]byte(fmt.Sprintf("  * %s\n", transliterate(*item.Comment))))
		}
	}
	conn.Write([]byte("------------------------------------------------\n"))

	conn.Write(ALIGN_RIGHT)
	conn.Write(FONT_DOUBLE_W)
	conn.Write([]byte(fmt.Sprintf("JAMI: %.0f so'm\n", order.TotalPrice)))
	conn.Write(FONT_NORMAL)
	conn.Write([]byte("\n\n"))

	conn.Write(ALIGN_CENTER)
	conn.Write([]byte("Tayyor bo'lgach xabar beriladi.\n"))
	conn.Write([]byte("\n\n\n\n\n"))

	conn.Write(PAPER_CUT)
}

func transliterate(text string) string {
	r := strings.NewReplacer(
		"ў", "o'", "Ў", "O'",
		"қ", "q", "Қ", "Q",
		"ғ", "g'", "Ғ", "G'",
		"ҳ", "h", "Ҳ", "H",
		"‘", "'", "’", "'",
	)
	return r.Replace(text)
}
