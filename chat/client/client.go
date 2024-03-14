package main

import (
	"bufio"
	"chat/messages"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	usrnme := os.Args[1]
	fmt.Println("Hello", usrnme)

	host := os.Args[2]
	conn, err := net.Dial("tcp", host)

	if err != nil {
		log.Fatalln(err.Error())
	}

	defer conn.Close()

	msgHandler := messages.NewMessageHandler(conn)

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println("message>")
		// Take message from user
		result := scanner.Scan() // Reads single line

		if result == false {
			break
		}

		message := scanner.Text()

		if len(message) != 0 {
			msg := messages.Chat{Username: usrnme, MessageBody: message}

			wrapper := &messages.Wrapper{
				Msg: &messages.Wrapper_ChatMsg{ChatMsg: &msg},
			}
			msgHandler.Send(wrapper)
		}
	}
}
