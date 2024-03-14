package messages

import (
	"encoding/binary"
	"net"

	"google.golang.org/protobuf/proto"
)

type MessageHandler struct {
	conn net.Conn
}

func NewMessageHandler(conn net.Conn) *MessageHandler {
	msgHandler := &MessageHandler{
		conn: conn,
	}

	return msgHandler
}

func (msgHandler *MessageHandler) readN(buff []byte) error {
	byteRead := uint64(0)

	for byteRead < uint64(len(buff)) {
		n, error := msgHandler.conn.Read(buff)

		if error != nil {
			return error
		}

		byteRead += uint64(n)
	}

	return nil
}

func (msgHandler *MessageHandler) writeN(buff []byte) error {
	byteWrite := uint64(0)

	for byteWrite < uint64(len(buff)) {
		n, error := msgHandler.conn.Write(buff)

		if error != nil {
			return error
		}

		byteWrite += uint64(n)
	}

	return nil
}

func (msgHandler *MessageHandler) Send(wrapper *Wrapper) error {
	serialized, err := proto.Marshal(wrapper)

	if err != nil {
		return err
	}

	prefix := make([]byte, 8)
	binary.LittleEndian.PutUint64(prefix, uint64(len(serialized)))
	msgHandler.writeN(prefix)
	msgHandler.writeN(serialized)

	return nil
}

func (msgHandler *MessageHandler) Receive() (*Wrapper, error) {
	// serialized, err := proto.Marshal(wrapper)
	prefix := make([]byte, 8)
	msgHandler.readN(prefix)

	payloadSize := binary.LittleEndian.Uint64(prefix)

	payload := make([]byte, payloadSize)

	msgHandler.readN(payload)

	wrapper := &Wrapper{}
	err := proto.Unmarshal(payload, wrapper)

	return wrapper, err
}

func (msgHandler *MessageHandler) Close() {
	msgHandler.conn.Close()
}
