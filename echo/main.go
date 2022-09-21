package main

import (
	"errors"
	"io"
	"log"
	"net"
	"strconv"
)

func do(port int) error {
	log.Printf("listen :%d ", port)
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return err
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer func() {
		conn.Close()
		log.Printf("closed: %s", conn.RemoteAddr())
	}()
	log.Printf("accepted: %s", conn.RemoteAddr())

	for {
		buf := make([]byte, 4096)
		n, err := conn.Read(buf)
		if n == 0 || err != nil {
			if !errors.Is(err, io.EOF) {
				log.Println(err)
			}
			return
		}

		_, err = conn.Write(buf[:n])
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func main() {
	port := 1234

	if err := do(port); err != nil {
		log.Fatal(err)
	}
}
