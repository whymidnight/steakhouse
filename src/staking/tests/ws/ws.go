package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/recws-org/recws"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func newWS() {
	http.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		conn, _ := upgrader.Upgrade(w, r, nil) // error ignored for sake of simplicity

		for {
			// Read message from browser
			msgType, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}

			// Print the message to the console
			fmt.Printf("%s sent: %s\n", conn.RemoteAddr(), string(msg))

			// Write message back to browser
			if err = conn.WriteMessage(msgType, msg); err != nil {
				return
			}
		}
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "websockets.html")
	})

	http.ListenAndServe(":8080", nil)
}

func main() {
	// newWS()
	ctx, cancel := context.WithCancel(context.Background())
	ws := recws.RecConn{
		KeepAliveTimeout: 10 * time.Second,
	}
	ws.Dial("wss://api.devnet.solana.com", nil)

	go func() {
		time.Sleep(2 * time.Second)
		cancel()
	}()

	for {
		select {
		case <-ctx.Done():
			go ws.Close()
			log.Printf("Websocket closed %s", ws.GetURL())
			return
		default:
			if !ws.IsConnected() {
				// log.Printf("Websocket disconnected %s", ws.GetURL())
				continue
			}

			if err := ws.WriteMessage(1, []byte("Incoming")); err != nil {
				log.Printf("Error: WriteMessage %s", ws.GetURL())
				return
			}

			_, message, err := ws.ReadMessage()
			if err != nil {
				log.Printf("Error: ReadMessage %s", ws.GetURL())
				return
			}

			log.Printf("Success: %s", message)
		}
	}
}

