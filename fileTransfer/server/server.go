package main

import (
	"crypto/sha512"
	"fileTransfer/files"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

func getFilePath(path string, fileName string) string {
	fp := filepath.Join(path, fileName)
	return fp
}

func calcCheckSum(path string) []byte {
	file, err := os.Open(path)
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer file.Close()

	h := sha512.New()

	if _, shaErr := io.Copy(h, file); err != nil {
		log.Fatalln(shaErr)
	}

	return h.Sum(nil)

}

func checkSpace(path string, fileSize int32, fileHandler *files.FileHandler) {
	var stat unix.Statfs_t

	unix.Statfs(path, &stat)

	// fmt.Println("Available space: ", stat.Bavail*uint64(stat.Bsize))
	freeSpace := stat.Bavail * uint64(stat.Bsize)
	if freeSpace < uint64(fileSize) {
		diskSpaceMsg := files.FileDetails{Status: "No Space available"}
		diskSpaceWrapper := &files.Wrapper{
			Operations: &files.Wrapper_FileDetails{
				FileDetails: &diskSpaceMsg,
			},
		}

		fileHandler.Send(diskSpaceWrapper)
	}
}

func checkFileExist(fileName string, fileHandler *files.FileHandler) {
	_, fileErr := os.Stat(fileName)
	if fileErr == nil {
		fileExistMsg := files.FileDetails{Status: "File Already Exist"}
		fileExistWrapper := &files.Wrapper{
			Operations: &files.Wrapper_FileDetails{
				FileDetails: &fileExistMsg,
			},
		}

		fileHandler.Send(fileExistWrapper)
	}
}
func handleClient(fileHandler *files.FileHandler, path string) {
	defer fileHandler.Close()
	// TODO : Create a hashmap to store the details
	var act string
	var fileSize int32
	var fileName string
	var checkSum []byte
	for {
		wrapper, _ := fileHandler.Receive()

		switch opn := wrapper.Operations.(type) {
		case *files.Wrapper_FileDetails:

			act = opn.FileDetails.Action
			fileSize = opn.FileDetails.Size
			fileName = opn.FileDetails.FileName
			checkSum = opn.FileDetails.Checksum
			fmt.Println("File Details")
			fmt.Println("Filename " + fileName + "\n")
			fmt.Printf("Filesize %d\n", fileSize)

			if act == "put" {

				// File exits
				checkFileExist(fileName, fileHandler)

				// Check disk space
				checkSpace(path, fileSize, fileHandler)

				// Send an ok response
				okMsg := files.FileDetails{Status: "OK"}
				okWrapper := &files.Wrapper{
					Operations: &files.Wrapper_FileDetails{FileDetails: &okMsg},
				}
				fileHandler.Send(okWrapper)
			}
		case *files.Wrapper_File:
			if act == "put" {

				// var fileArr []byte

				fileArr := wrapper.GetFile().FileData
				writeErr := os.WriteFile(fileName, fileArr, 0644)
				fmt.Println("File arr" + string(fileArr) + "\n")

				if writeErr != nil {
					log.Fatalln(writeErr.Error())
				} else {
					fmt.Println("Wrote to a file")
					// Calculate checksum of file and check with the original checksum
					fp := getFilePath(path, fileName)
					cs := calcCheckSum(fp)

					// If both the checksums match send Transfer success
					if string(cs) == string(checkSum) {
						transferMsg := files.File{TransferStatus: "success"}
						transferWrapper := &files.Wrapper{
							Operations: &files.Wrapper_File{File: &transferMsg},
						}
						fmt.Println("Send success message")
						fileHandler.Send(transferWrapper)
					}
				}
			}
		case nil:
			log.Printf("Received an empty message. Disconnecting the Client")
			return
		default:
			log.Printf("Unexpected message type %T", opn)
		}

	}

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
