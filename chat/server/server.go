package main

import (
	"chat/messages"
	"fmt"
	"log"
	"net"
	"os"
)

func handleClient(msgHandler *messages.MessageHandler) {
	defer msgHandler.Close()

	// To handle multiple messages from the same client
	for {
		wrapper, _ := msgHandler.Receive()

		switch msg := wrapper.Msg.(type) {
		case *messages.Wrapper_RegMsg:
			fmt.Println("Got a registration message. Not implemented yet")

		case *messages.Wrapper_ChatMsg:
			fmt.Println("<" + msg.ChatMsg.GetUsername() + "> " + msg.ChatMsg.GetMessageBody())
		case nil:
			log.Println("Received an empty message. Terminating Client")
			return
		default:
			log.Printf("Unexpected message type: %d", msg)
		}

	}

}

func startServer() {

	listner, ListenErr := net.Listen("tcp", ":"+os.Args[1])

	if ListenErr != nil {
		log.Fatalln(ListenErr.Error())
	}

	// Infinite loop to continuously handle requests
	for {
		fmt.Println("Started an infinte loop")
		conn, err := listner.Accept()
		fmt.Println("Accepted a client")
		if err == nil {
			msgHandler := messages.NewMessageHandler(conn)

			go handleClient(msgHandler)
		}
	}
}
func main() {
	path, dirErr := os.Getwd()

	if dirErr != nil {
		log.Fatalln(dirErr.Error())
	}

	if len(os.Args) == 3 {
		path = os.Args[2]
	}

	_, err := os.Stat(path)

	// Path does not exist
	if os.IsNotExist(err) {
		log.Fatalln(err.Error())
	}

	// Path exists or path is not given
	if err == nil || path == "" {
		startServer()
	}

}
