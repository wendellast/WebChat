package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

type Message struct {
	Username string `json:"username"`
	Message  string `json:"message"`
}

var broadcast = make(chan Message)
var clients = make(map[*websocket.Conn]bool)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	fs := http.FileServer(http.Dir("."))
	http.Handle("/", fs)
	http.HandleFunc("/ws", handlerConection)

	go handlerMessages()

	fmt.Println("Runing...")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

func handlerConection(w http.ResponseWriter, r *http.Request) {
	WS, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		panic(err)
	}

	defer WS.Close()
	clients[WS] = true

	for {
		var msg Message
		if err := WS.ReadJSON(&msg); err != nil {
			delete(clients, WS)
			break
		}

		broadcast <- msg
	}
}

func handlerMessages() {
	for {
		msg := <-broadcast
		for client := range clients {
			err := client.WriteJSON(msg)

			if err != nil {
				delete(clients, client)
				client.Close()
			}
		}
	}
}
