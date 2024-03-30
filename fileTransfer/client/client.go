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

func getFilePath(path string, fileName string) string {
	fp := filepath.Join(path, fileName)
	return fp
}

func handlePutRequest(fileHandler *files.FileHandler, host string, action string, fileName string, size int32, checkSum []byte, fileArr []byte) {

	if len(fileName) != 0 {
		fmt.Println("Filename " + fileName)
		fileMsg := files.FileDetails{FileName: fileName, Size: size, Checksum: checkSum, Action: action}

		wrapper := &files.Wrapper{
			Operations: &files.Wrapper_FileDetails{FileDetails: &fileMsg},
		}
		fileHandler.Send(wrapper)

		// Wait for servers response to the file details
		wrapper, _ = fileHandler.Receive()

		fmt.Printf("Message from server: %s\n", wrapper.GetFileDetails().GetStatus())
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
}

func handleGetRequest(fileHandler *files.FileHandler, action string, fileName string, path string) {

	if len(fileName) != 0 {
		fileDetails := files.FileDetails{FileName: fileName, Action: action}

		// Create a wrapper message with filename and action to the server
		fileDetailsWrapper := &files.Wrapper{
			Operations: &files.Wrapper_FileDetails{
				FileDetails: &fileDetails,
			},
		}

		fileHandler.Send(fileDetailsWrapper)

		// Either size, checksum or filedoes not exit
		wrapper, _ := fileHandler.Receive()
		serverChecksum := wrapper.GetFileDetails().Checksum
		status := wrapper.GetFileDetails().Status
		// size := wrapper.GetFileDetails().Size
		fmt.Println("Status: ", status)
		if status != "details" {
			fmt.Println("Server does not contain the file")
		} else {
			// TODO : Create the file and begin storing its data

			// Get the file array
			wrapper, _ := fileHandler.Receive()
			fileArr := wrapper.GetFile().FileData
			writeErr := os.WriteFile(fileName, fileArr, 0644)
			fmt.Println("File arr" + string(fileArr) + "\n")

			if writeErr != nil {
				log.Fatalln(writeErr.Error())
			} else {
				fmt.Println("Wrote to a file")
				fp := getFilePath(path, fileName)
				cs := calcCheckSum(fp)

				if string(cs) == string(serverChecksum) {
					fmt.Println("Checksum matched")
					fmt.Println("success")
				}
			}

		}
	}

}
func connectToServer(host string, action string, filePath string, path string) {
	conn, err := net.Dial("tcp", host)

	if err != nil {
		log.Fatalln(err.Error())
	}

	defer conn.Close()

	fileHandler := files.NewFileHandler(conn)

	if strings.ToLower(action) == "put" {

		fileName := extractFileName(filePath)
		size := calcFileSize(filePath)
		fmt.Printf("Filesize %d\n", size)
		fileArr := readFile(filePath)
		checkSum := calcCheckSum(filePath)

		handlePutRequest(fileHandler, host, action, fileName, size, checkSum, fileArr)

	} else if strings.ToLower(action) == "get" {

		handleGetRequest(fileHandler, action, filePath, path)

	}

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

	// If get request it is file name else it is file path
	filePath := os.Args[3]

	if action == "get" {

		if len(os.Args) == 5 {
			path = os.Args[4]
		}
	}

	// Connect to the server
	connectToServer(host, action, filePath, path)

}
