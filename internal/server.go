package internal

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Message struct {
	Name string `json:"name"`
	Data string `json:"data"`
	Time string `json:"time"`
}

type Hub struct {
	Clients    map[*websocket.Conn]bool
	Register   chan *websocket.Conn
	Unregister chan *websocket.Conn
	Broadcast  chan Message
	Mutex      sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[*websocket.Conn]bool),
		Register:   make(chan *websocket.Conn),
		Unregister: make(chan *websocket.Conn),
		Broadcast:  make(chan Message),
		Mutex:      sync.Mutex{},
	}
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.Register:
			h.Mutex.Lock()
			h.Clients[c] = true
			h.Mutex.Unlock()
		case c := <-h.Unregister:
			h.Mutex.Lock()
			if _, ok := h.Clients[c]; ok {
				delete(h.Clients, c)
			}
			h.Mutex.Unlock()
		case m := <-h.Broadcast:
			h.Mutex.Lock()
			for client := range h.Clients {
				err := client.WriteJSON(m)
				if err != nil {
					_ = client.Close()
					delete(h.Clients, client)
				}
			}
			h.Mutex.Unlock()
		}
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func WebSocketHandler(c *gin.Context, hub *Hub) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		panic(err)
	}
	hub.Register <- conn
	defer func() {
		hub.Unregister <- conn
		_ = conn.Close()
	}()

	for {
		var msg Message
		_, message, err := conn.ReadMessage()
		if err != nil {
			break
		}
		_ = json.Unmarshal(message, &msg)

		hub.Broadcast <- msg

	}
}

type Configuration struct {
	Path string
	Port string
}

func RunServeCommand(cfg Configuration) {
	r := gin.Default()
	hub := NewHub()
	go hub.Run()
	r.GET(cfg.Path, func(c *gin.Context) {
		WebSocketHandler(c, hub)
	})

	err := r.Run(":" + cfg.Port)
	if err != nil {
		panic(err)
	}
}
