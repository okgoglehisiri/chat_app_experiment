package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan Message)

type Message struct {
	Type    int
	Message []byte
}

func main() {
	r := gin.Default()

	// WebSocket upgrader configuration
	wsupgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	// Serve the index.html file on the root route
	r.GET("/", func(ctx *gin.Context) {
		http.ServeFile(ctx.Writer, ctx.Request, "client/index.html")
	})

	// Handle WebSocket connections
	r.GET("/ws", func(ctx *gin.Context) {
		conn, err := wsupgrader.Upgrade(ctx.Writer, ctx.Request, nil)
		if err != nil {
			log.Printf("Failed to set websocket upgrade: %+v\n", err)
			return
		}
		defer conn.Close() // Ensure connection is closed when the function exits

		// Add the new connection to the clients map
		clients[conn] = true

		// Continuously read messages from the WebSocket connection
		for {
			t, msg, err := conn.ReadMessage()
			if err != nil {
				log.Printf("ReadMessage Error: %+v\n", err)
				delete(clients, conn)
				break
			}
			// Send the received message to the broadcast channel
			broadcast <- Message{Type: t, Message: msg}
		}
	})

	// Start handling messages in a separate goroutine
	go handleMessages()

	// Run the server on port 4001
	r.Run(":4001")
}

// Handle broadcasting messages to all connected clients
func handleMessages() {
	for {
		message := <-broadcast
		for client := range clients {
			err := client.WriteMessage(message.Type, message.Message)
			if err != nil {
				log.Printf("WriteMessage Error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}
