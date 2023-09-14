package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/novalagung/gubrak/v2"
)

type M map[string]interface{}

const MESSAGE_NEW_USER = "New User"
const MESSAGE_CHAT = "Chat"
const MESSAGE_LEAVE = "Leave"

// var connections = make([]*WebSocketConnection, 0)

var rooms = make(map[string][]*WebSocketConnection)

// var rooms = []string{"1", "2", "3"}

type SocketPayload struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Room    string `json:"room"`
}

type SocketResponse struct {
	From    string
	Type    string
	Message string
	Room    string
}

type WebSocketConnection struct {
	*websocket.Conn
	Username string
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	http.Handle("/", http.FileServer(http.Dir("./frontend")))

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		// socket code here
		currentGorillaConn, err := upgrader.Upgrade(w, r, w.Header())
		if err != nil {
			http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
			return
		}

		username := r.URL.Query().Get("username")
		currentConn := WebSocketConnection{Conn: currentGorillaConn, Username: username}
		// connections = append(connections, &currentConn)

		go handleIO(&currentConn)
	})

	fmt.Println("Server starting at :8080")
	log.Fatalln(http.ListenAndServe(":8080", nil))
}

func handleIO(currentConn *WebSocketConnection) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("ERROR", fmt.Sprintf("%v", r))
		}
	}()

	// broadcastMessage(currentConn, MESSAGE_NEW_USER, "")

	for {
		payload := SocketPayload{}
		err := currentConn.ReadJSON(&payload)
		if err != nil && strings.Contains(err.Error(), "websocket: close") {
			log.Println("ERROR", err.Error())
			continue
		}

		switch payload.Type {
		case "disconnect":
			broadcastMessage(currentConn, payload, MESSAGE_LEAVE)
			ejectConnection(currentConn, payload.Room)
			return
		case "login":
			rooms[payload.Room] = append(rooms[payload.Room], currentConn)
			broadcastMessage(currentConn, payload, MESSAGE_NEW_USER)
		case "chat":
			broadcastMessage(currentConn, payload, MESSAGE_CHAT)
		}

		// broadcastMessage(currentConn, MESSAGE_CHAT, payload.Message)
	}
}

func broadcastMessage(currentConn *WebSocketConnection, payload SocketPayload, kind string) {
	for _, eachConn := range rooms[payload.Room] {
		if eachConn == currentConn {
			continue
		}

		eachConn.WriteJSON(SocketResponse{
			From:    currentConn.Username,
			Type:    kind,
			Message: payload.Message,
			Room:    payload.Room,
		})
	}
}

func ejectConnection(currentConn *WebSocketConnection, room string) {
	filtered := gubrak.From(rooms[room]).Reject(func(each *WebSocketConnection) bool {
		return each == currentConn
	}).Result()
	rooms[room] = filtered.([]*WebSocketConnection)
	if len(rooms[room]) == 0 {
		delete(rooms, room)
	}
}
