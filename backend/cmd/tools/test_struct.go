//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"time"
)

type Table struct {
	ID        int       `json:"id"`
	Number    int       `json:"number"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func main() {
	var payload struct {
		Table
		PaymentMethod string `json:"payment_method"`
	}

	data := []byte(`{"id": 4, "number": 4, "status": "free", "payment_method": "click"}`)
	err := json.Unmarshal(data, &payload)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("Parsed: %+v\n", payload)
}
