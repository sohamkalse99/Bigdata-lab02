package main

import (
	"fileTransfer/files"
	"fmt"
	"log"
	"net"
	"os"
)

func handleClient(fileHandler *files.FileHandler, path string) {
	defer fileHandler.Close()

	// for {
	wrapper, _ := fileHandler.Receive()

	switch opn := wrapper.Operations.(type) {
	case *files.Wrapper_FileDetails:
		fmt.Println("File Details")
		fmt.Println("Filename " + opn.FileDetails.FileName + "\n")

		if opn.FileDetails.Action == "put" {
			// Send an ok response
			okMsg := files.FileDetails{Status: "OK"}
			okWrapper := &files.Wrapper{
				Operations: &files.Wrapper_FileDetails{FileDetails: &okMsg},
			}
			fileHandler.Send(okWrapper)
		}
	case *files.Wrapper_File:

	case nil:
		log.Printf("Received an empty message. Terminating Client")
		return
	default:
		log.Printf("Unexpected message type %T", opn)
	}

	// }

}

func startServer(path string) {

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
			// msgHandler := messages.NewMessageHandler(conn)
			fileHandler := files.NewFileHandler(conn)
			go handleClient(fileHandler, path)
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
		startServer(path)
	}

}
