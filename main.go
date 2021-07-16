package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type msg struct {
	Error error       `json:"error"`
	Data  interface{} `json:"data"`
}

var clients = make(map[string]*websocket.Conn)
var upgrader = websocket.Upgrader{} // use default options

func main() {
	http.HandleFunc("/chat", chat)
	log.Println("listening on localhost:8080/chat")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func chat(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("problem upgrading the connection: %s", err)
	}
	defer c.Close()

	clients[c.RemoteAddr().String()] = c
	c.SetCloseHandler(func(code int, text string) error {
		delete(clients, c.RemoteAddr().String())
		broadcastWithout(c, fmt.Sprintf("%s has disconnected", c.RemoteAddr()))
		return nil
	})

	broadcastWithout(c, fmt.Sprintf("%s connected", c.RemoteAddr()))

	var incoming msg
	for {
		c.ReadJSON(&incoming)
		broadcastWithout(c, incoming.Data.(string))
	}
}

func sendTo(conn *websocket.Conn, data string) {
	log.Printf("sending %s to %s", data, conn.RemoteAddr())
	m := &msg{Data: data}
	conn.WriteJSON(m)
}

func broadcast(msg string) {
	for _, c := range clients {
		sendTo(c, msg)
	}
}

func broadcastWithout(conn *websocket.Conn, msg string) {
	for n, c := range clients {
		if n != conn.RemoteAddr().String() {
			sendTo(c, msg)
		}
	}
}
