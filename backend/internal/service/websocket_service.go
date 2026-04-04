package service

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // In production, restrict this
	},
}

type Client struct {
	ID   int
	Role string
	Conn *websocket.Conn
}

type WebsocketService struct {
	clients map[*Client]bool
	mu      sync.Mutex
}

func NewWebsocketService() *WebsocketService {
	return &WebsocketService{
		clients: make(map[*Client]bool),
	}
}

func (s *WebsocketService) HandleConnection(w http.ResponseWriter, r *http.Request, userID int, role string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WS Upgrade Error: %v", err)
		return
	}

	client := &Client{
		ID:   userID,
		Role: role,
		Conn: conn,
	}

	s.mu.Lock()
	s.clients[client] = true
	s.mu.Unlock()

	log.Printf("📡 [WS] New client: ID=%d, Role=%s, RemoteAddr=%s", userID, role, conn.RemoteAddr())

	// Keep connection alive/read messages if needed
	defer func() {
		s.mu.Lock()
		delete(s.clients, client)
		s.mu.Unlock()
		log.Printf("📉 [WS] Client disconnected: ID=%d, Role=%s", userID, role)
		conn.Close()
	}()

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}

func (s *WebsocketService) BroadcastToRole(role string, message interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	count := 0
	for client := range s.clients {
		if client.Role == role || role == "all" {
			err := client.Conn.WriteJSON(message)
			if err != nil {
				log.Printf("❌ [WS] Write Error to %d (%s): %v", client.ID, client.Role, err)
				client.Conn.Close()
				delete(s.clients, client)
			} else {
				count++
			}
		}
	}
	log.Printf("📢 [WS] Broadcast to Role='%s': Sent to %d clients", role, count)
}

func (s *WebsocketService) BroadcastToUser(userID int, message interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for client := range s.clients {
		if client.ID == userID {
			client.Conn.WriteJSON(message)
		}
	}
}
