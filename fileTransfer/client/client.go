package main

import (
	"bufio"
	"crypto/sha512"
	"fileTransfer/files"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
)

func extractFileName(path string) string {
	fileName := filepath.Base(path)
	return fileName
}
func calcFileSize(fileName string) int32 {
	file, err := os.Open(fileName)

	if err != nil {
		log.Fatalln(err.Error())
	}

	info, _ := file.Stat()
	size := info.Size()
	size32 := int32(size)
	defer file.Close()
	return size32
}

func readFile(fileName string) []byte {
	file, err := os.Open(fileName)

	if err != nil {
		log.Fatalln(err.Error())
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var fileBytes []byte
	for scanner.Scan() {
		fileBytes = append(fileBytes, scanner.Bytes()...)
		fileBytes = append(fileBytes, '\n')
	}

	if scanner.Err(); err != nil {
		log.Fatalln(err.Error())
	}

	return fileBytes
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
func connectToServer(host string, action string, fileName string, size int32, checkSum []byte, fileArr []byte) {
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
			fileMsg := files.FileDetails{FileName: fileName, Size: size, Checksum: checkSum, Action: action}

			wrapper := &files.Wrapper{
				Operations: &files.Wrapper_FileDetails{FileDetails: &fileMsg},
			}
			fileHandler.Send(wrapper)

			// Wait for servers response to the file details
			wrapper, _ = fileHandler.Receive()

			fmt.Printf("Message from server: %s", wrapper.GetFileDetails().GetStatus())
			status := wrapper.GetFileDetails().GetStatus()

			if status == "OK" {

				fileWrapper := &files.Wrapper{
					Operations: &files.Wrapper_File{File: &files.File{
						FileData: fileArr,
					}},
				}

				fileHandler.Send(fileWrapper)
				fmt.Println("Sent files to the server")

				wrapper, _ = fileHandler.Receive()
				transferStatus := wrapper.GetFile().TransferStatus

				if transferStatus == "success" {
					fmt.Println("Received Success message from server")
				}
			}
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
	// path := os.Args[3]

	path, dirErr := os.Getwd()

	if dirErr != nil {
		log.Fatalln(dirErr.Error())
	}
	if len(os.Args) == 4 {
		path = os.Args[3]
	}

	fileName := extractFileName(path)
	size := calcFileSize(path)
	fmt.Printf("Filesize %d\n", size)
	fileArr := readFile(path)
	checkSum := calcCheckSum(path)
	// Connect to the server
	connectToServer(host, action, fileName, size, checkSum, fileArr)

}
