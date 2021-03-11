package main

import (
	"fmt"
	"log"
	"net"
	"time"
)

func main() {
	service := ":3333"
	tcpAddr, err := net.ResolveTCPAddr("tcp", service)
	if err != nil {
		log.Printf("err, %s",err)
	}
	listner, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Printf("err, %s",err)
	}
	for {
		conn, err := listner.Accept()
		if err != nil {
			continue
		}

		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	fmt.Println("client accept")
	messageBuf := make([]byte, 1024)
	messageLen, err := conn.Read(messageBuf)
	if err != nil {
		log.Printf("err, %s",err)
	}

	message := string(messageBuf[:messageLen])
	message = message + ""

	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	conn.Write([]byte(message))
}