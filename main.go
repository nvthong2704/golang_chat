package chat_socket

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var clients = make(map[*websocket.Conn]bool) // connected clients
var broadcast = make(chan Message)           // broadcast channel

// configure the upgrader
var upgrader = websocket.Upgrader{}

// define out message object
type Message struct {
	Email   string `json:"email"`
	User    string `json:"user"`
	Message string `json:"message"`
}

func main() {
	// create a simple file saver
	fs := http.FileServer(http.Dir("../public"))
	http.Handle("/", fs)

	// config websocket route
	http.HandleFunc("/ws", handleConnections)

	// start listening for incoming chat messages
	go handleMessages()

	// start the server on localhost port 8000 and log any errors
	log.Println("http server started on : 8000")
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatalf("listenAndServer: ", err)
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// update initial GET request to a websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	// make sure we close the connection when the function returns
	defer ws.Close()

	// register our new client
	clients[ws] = true

	for {
		var msg Message
		// read in a new message as json and map it to a message object
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("err: %v", err)
			delete(clients, ws)
			break
		}
		// send the newly received message to the broadcast channel
		broadcast <- msg
	}
}

func handleMessages() {
	for {
		// grab the next message from the broadcast channel
		msg := <-broadcast
		// send it out to every client that is currently connected
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("err : %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}
