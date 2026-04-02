package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

// Simplified Order struct for printing
type OrderPrint struct {
	ID         int      `json:"id"`
	Phone      string   `json:"phone"`
	Address    string   `json:"address"`
	Items      []Item   `json:"items"`
	TotalPrice float64  `json:"total_price"`
}

type Item struct {
	ProductName string `json:"product_name"`
	Quantity    int    `json:"quantity"`
}

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// In a real scenario, this would be the server address
	serverAddr := os.Getenv("API_URL")
	if serverAddr == "" {
		serverAddr = "localhost:8080"
	}

	u := url.URL{Scheme: "ws", Host: serverAddr, Path: "/api/ws"}
	log.Printf("Connecting to %s", u.String())

	// Simulated Auth Token for Printer (Admin or specific Printer role)
	token := os.Getenv("PRINTER_TOKEN")
	header := http.Header{}
	header.Add("Authorization", "Bearer "+token)

	c, _, err := websocket.DefaultDialer.Dial(u.String(), header)
	if err != nil {
		log.Fatal("Dial Error:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("Read Error:", err)
				return
			}

			var event struct {
				Type  string      `json:"type"`
				Order OrderPrint  `json:"order"`
			}
			
			if err := json.Unmarshal(message, &event); err != nil {
				continue
			}

			if event.Type == "new_order" {
				printOrder(event.Order)
			}
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-interrupt:
			log.Println("Interrupt received, closing...")
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("Write Close Error:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}

func printOrder(o OrderPrint) {
	fmt.Println("\n==============================")
	fmt.Println("       YANGI BUYURTMA         ")
	fmt.Println("==============================")
	fmt.Printf("ID:      #%d\n", o.ID)
	fmt.Printf("Mijoz:   %s\n", o.Phone)
	fmt.Printf("Vaqt:    %s\n", time.Now().Format("15:04:05"))
	fmt.Println("------------------------------")
	for _, item := range o.Items {
		fmt.Printf("%-20s x%-3d\n", item.ProductName, item.Quantity)
	}
	fmt.Println("------------------------------")
	fmt.Printf("JAMI:    %.0f so'm\n", o.TotalPrice)
	fmt.Println("==============================")
	
	// In real setup, you would send ESC/POS commands to /dev/usb/lp0 or similar
}
