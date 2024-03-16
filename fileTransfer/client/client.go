package main

import (
	"fileTransfer/files"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

func connectToServer(host string, action string, fileName string, path string) {
	conn, err := net.Dial("tcp", host)

	if err != nil {
		log.Fatalln(err.Error())
	}

	defer conn.Close()

	if strings.ToLower(action) == "put" {
		fileHandler := files.NewFileHandler(conn)

		// for {
		if len(fileName) != 0 {
			fmt.Println("Filename " + fileName)
			fileMsg := files.FileDetails{FileName: fileName, Action: action}

			wrapper := &files.Wrapper{
				Operations: &files.Wrapper_FileDetails{FileDetails: &fileMsg},
			}

			fileHandler.Send(wrapper)

			// Wait for servers response to the file details
			wrapper, _ = fileHandler.Receive()

			fmt.Printf("Message from server: %s", wrapper.GetFileDetails().GetStatus())
		}

		// }
	} /*else if strings.ToLower(action) == "get" {

	}*/

}
func main() {

	if len(os.Args) < 3 {
		fmt.Println("Need to enter atleast action and filename")
		os.Exit(0)
	}
	host := os.Args[1]
	action := os.Args[2]
	fileName := os.Args[3]

	path, dirErr := os.Getwd()

	if dirErr != nil {
		log.Fatalln(dirErr.Error())
	}
	if len(os.Args) == 4 {
		path = os.Args[3]
	}

	// Connect to the server
	connectToServer(host, action, fileName, path)

}
