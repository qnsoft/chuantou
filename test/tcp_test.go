package test

import (
	"bytes"
	"encoding/binary"
	"github.com/denisbrodbeck/machineid"
	"log"
	"net"
	"testing"
)

// 测试连接中断

func TestTCPServer(t *testing.T) {
	listener, _ := net.Listen("tcp", ":6666")
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		var length uint32

		if err = binary.Read(conn, binary.BigEndian, &length); err != nil {
			log.Println(err)
			_ = conn.Close()
			return
		}

		// 读取协议内容
		body := make([]byte, length)
		if err = binary.Read(conn, binary.BigEndian, &body); err != nil {
			log.Println(err)
			_ = conn.Close()
			return
		}
		log.Println(string(body))
		_ = conn.Close()
	}
}

func TestTCPClient(t *testing.T) {
	conn, err := net.Dial("tcp", ":6666")
	if err != nil {
		log.Println("dial failed")
		return
	}
	body, _ := machineid.ID()

	var buf = make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(len(body)))

	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(buf)
	buffer.WriteString(body)

	if _, err := conn.Write(buffer.Bytes()); err != nil {
		log.Println(err)

		_ = conn.Close()
	}
}
