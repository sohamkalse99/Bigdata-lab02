package files

import (
	"encoding/binary"
	"net"

	"google.golang.org/protobuf/proto"
)

type FileHandler struct {
	conn net.Conn
}

func NewFileHandler(conn net.Conn) *FileHandler {
	fileHandler := &FileHandler{
		conn: conn,
	}

	return fileHandler
}

func (fileHandler *FileHandler) readN(buff []byte) error {
	byteRead := uint64(0)

	for byteRead < uint64(len(buff)) {
		n, error := fileHandler.conn.Read(buff)

		if error != nil {
			return error
		}

		byteRead += uint64(n)
	}

	return nil
}

func (fileHandler *FileHandler) writeN(buff []byte) error {
	byteWrite := uint64(0)

	for byteWrite < uint64(len(buff)) {
		n, error := fileHandler.conn.Write(buff)

		if error != nil {
			return error
		}

		byteWrite += uint64(n)
	}

	return nil
}

func (fileHandler *FileHandler) Send(wrapper *Wrapper) error {
	serialized, err := proto.Marshal(wrapper)

	if err != nil {
		return err
	}

	prefix := make([]byte, 8)
	binary.LittleEndian.PutUint64(prefix, uint64(len(serialized)))
	fileHandler.writeN(prefix)
	fileHandler.writeN(serialized)

	return nil
}

func (fileHandler *FileHandler) Receive() (*Wrapper, error) {
	// serialized, err := proto.Marshal(wrapper)
	prefix := make([]byte, 8)
	fileHandler.readN(prefix)

	payloadSize := binary.LittleEndian.Uint64(prefix)

	payload := make([]byte, payloadSize)

	fileHandler.readN(payload)

	wrapper := &Wrapper{}
	err := proto.Unmarshal(payload, wrapper)

	return wrapper, err
}

func (fileHandler *FileHandler) Close() {
	fileHandler.conn.Close()
}
