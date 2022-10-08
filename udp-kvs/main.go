package main

import (
	"bytes"
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
	data := buf[:]
	log.Printf("received: %q", data)
	defer log.Println()

	if bytes.Contains(data, []byte("=")) {
		tmp := bytes.SplitN(data, []byte("="), 2)

		key := string(tmp[0])
		value := tmp[1]
		if key == "version" {
			log.Println("key is `version`, so ignore")
			return
		}
		log.Printf("key = %s, value = %q, store", key, value)

		kvs[key] = value[:]
	} else {
		key := string(data)
		value, ok := kvs[key]
		if !ok {
			value = []byte("")
		}
		log.Printf("key = %s, value = %q, retrieve", key, value)

		kb := []byte(key)
		kb = append(kb, '=')
		kb = append(kb, value...)
		sendMessage(addr, conn, kb)
	}
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
		buf = make([]byte, 1024)
	}
}
