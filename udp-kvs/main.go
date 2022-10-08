package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
)

var (
	kvs = make(map[string][]byte)
)

func init() {
	kvs["version"] = []byte("Ken's Key-Value Store 1.0")
}

func sendMessage(addr *net.UDPAddr, conn *net.UDPConn, message []byte) {
	log.Printf("send message: len(message)=%d, %q", len(message), message)
	if _, err := conn.WriteTo(message, addr); err != nil {
		log.Println(err)
	}
}

func handleRequest(addr *net.UDPAddr, conn *net.UDPConn, buf []byte) {
	log.Printf("received: %q", buf)
	defer log.Println()

	tmp := bytes.SplitN(buf, []byte("="), 2)
	if len(tmp) >= 2 {
		key := string(tmp[0])
		value := tmp[1]
		if key == "version" {
			log.Println("key is `version`, so ignore")
			return
		}
		log.Printf("key = %s, value = %q, store", key, value)

		kvs[key] = value[:]
		return
	}

	key := string(buf)
	value := kvs[key]
	log.Printf("key = %s, value = %q, retrieve", key, value)

	sendMessage(addr, conn, fmt.Appendf(nil, "%s=%s", key, value))
}

func main() {
	addr := &net.UDPAddr{
		IP:   net.ParseIP("localhost"),
		Port: 1234,
	}
	listener, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Println(err)
	}

	buf := make([]byte, 1024)
	log.Println("starting udp server")
	for {
		n, ad, err := listener.ReadFromUDP(buf)
		if err != nil {
			log.Println(err)
			continue
		}

		handleRequest(ad, listener, buf[:n])
	}
}
