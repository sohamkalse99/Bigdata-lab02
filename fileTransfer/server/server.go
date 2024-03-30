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

	"golang.org/x/sys/unix"
)

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
func checkSpace(path string, fileSize int32, fileHandler *files.FileHandler, flag *bool) {
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
		*flag = false
		fileHandler.Send(diskSpaceWrapper)
	}

}

func checkFileExist(fileName string, fileHandler *files.FileHandler, flag *bool) {
	_, fileErr := os.Stat(fileName)
	if fileErr == nil {
		fileExistMsg := files.FileDetails{Status: "File Already Exist"}
		fileExistWrapper := &files.Wrapper{
			Operations: &files.Wrapper_FileDetails{
				FileDetails: &fileExistMsg,
			},
		}
		*flag = false
		fileHandler.Send(fileExistWrapper)
	}
}

func handlePutFileDets(flag *bool, fileName string, fileHandler *files.FileHandler, path string, fileSize int32) {
	*flag = true
	// File exits
	checkFileExist(fileName, fileHandler, flag)

	// Check disk space
	checkSpace(path, fileSize, fileHandler, flag)
	fmt.Println("Flag: ", *flag)
	// Send an ok response
	okMsg := files.FileDetails{Status: "OK"}
	okWrapper := &files.Wrapper{
		Operations: &files.Wrapper_FileDetails{FileDetails: &okMsg},
	}
	fileHandler.Send(okWrapper)
}

func handlePutFiles(wrapper *files.Wrapper, fileName string, path string, checkSum []byte, fileHandler *files.FileHandler) {

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

func handleGetFileDets(fileName string, fileHandler *files.FileHandler, path string) {
	// TODO : Ensure the file requested actually exists
	_, fileErr := os.Stat(fileName)

	// File Exists
	if fileErr == nil {
		// Send a response back to the client with the file’s size and checksum (or indicate failure if it doesn’t exit)
		filePath := getFilePath(path, fileName)
		size := calcFileSize(filePath)
		checkSum := calcCheckSum(filePath)
		status := "details"
		fileDetailsMsg := files.FileDetails{Size: size, Checksum: checkSum, Status: status}
		fileDetailsWrapper := &files.Wrapper{
			Operations: &files.Wrapper_FileDetails{
				FileDetails: &fileDetailsMsg,
			},
		}

		fileHandler.Send(fileDetailsWrapper)

		// TODO : Begin streaming file to the client

		// Convert file into an array of bytes
		fileArr := readFile(filePath)

		fileWrapper := &files.Wrapper{
			Operations: &files.Wrapper_File{
				File: &files.File{
					FileData: fileArr,
				},
			},
		}

		// Send the file
		fileHandler.Send(fileWrapper)
	} else {
		// File does not exist

		fileNotExistMsg := files.FileDetails{Status: "File does not exist"}
		fileExistWrapper := &files.Wrapper{
			Operations: &files.Wrapper_FileDetails{
				FileDetails: &fileNotExistMsg,
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
	var flagVal bool
	flag := &flagVal
	for {
		wrapper, _ := fileHandler.Receive()

		switch opn := wrapper.Operations.(type) {
		case *files.Wrapper_FileDetails:

			act = opn.FileDetails.Action
			fileSize = opn.FileDetails.Size
			fileName = opn.FileDetails.FileName
			checkSum = opn.FileDetails.Checksum
			// fmt.Println("File Details")
			// fmt.Println("Filename " + fileName + "\n")
			// fmt.Printf("Filesize %d\n", fileSize)

			if act == "put" {
				handlePutFileDets(flag, fileName, fileHandler, path, fileSize)
			} else if act == "get" {
				handleGetFileDets(fileName, fileHandler, path)
			}
		case *files.Wrapper_File:
			// If action is put and status is ok (checking this using boolean flag)
			if act == "put" && *flag {
				handlePutFiles(wrapper, fileName, path, checkSum, fileHandler)
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
