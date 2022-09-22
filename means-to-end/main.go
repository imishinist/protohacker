package main

import (
	"bytes"
	"encoding/binary"
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

	prices := make([]*Request, 0)

	rest := make([]byte, 0)
	for {
		buf := make([]byte, 4096)
		n, err := conn.Read(buf)
		if n == 0 || err != nil {
			if !errors.Is(err, io.EOF) {
				log.Println(err)
			}
			return
		}
		rest = append(rest, buf[:n]...)

		for len(rest) >= 9 {
			req, err := parseRequest(rest[:9])
			if err != nil {
				log.Println(err)
				return
			}
			if req.Type == "I" {
				prices = append(prices, req)
			} else {
				mean := meanPrice(prices, req.MinTime, req.MaxTime)

				if err := binary.Write(conn, binary.BigEndian, mean); err != nil {
					log.Println(err)
				}
			}
			rest = rest[9:]
		}
	}
}

func meanPrice(prices []*Request, minTime, maxTime int32) int32 {
	total := int64(0)
	count := 0
	for _, price := range prices {
		if price.Type != "I" {
			continue
		}
		if minTime <= price.Timestamp && price.Timestamp <= maxTime {
			total += int64(price.Price)
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return int32(float64(total) / float64(count))
}

type Request struct {
	Type string

	// when Type == "I"
	Timestamp int32
	Price     int32

	// when Type == "Q"
	MinTime int32
	MaxTime int32
}

func parseRequest(data []byte) (*Request, error) {
	if len(data) < 9 {
		panic("len(data) < 9")
	}

	buf := bytes.NewReader(data)

	fb, err := buf.ReadByte()
	if err != nil {
		return nil, err
	}

	request := Request{
		Type: string(fb),
	}
	var a, b int32
	if err := binary.Read(buf, binary.BigEndian, &a); err != nil {
		return nil, err
	}
	if err := binary.Read(buf, binary.BigEndian, &b); err != nil {
		return nil, err
	}

	if request.Type == "I" {
		request.Timestamp = a
		request.Price = b
	} else {
		request.MinTime = a
		request.MaxTime = b
	}
	return &request, nil
}

func main() {
	port := 1234

	if err := do(port); err != nil {
		log.Fatal(err)
	}
}
