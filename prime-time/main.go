package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net"
	"strconv"
)

const (
	N = 100000000
)

var (
	primeTable = make([]bool, N)
)

func init() {
	for i := 2; i*i <= N; i++ {
		if primeTable[i] {
			continue
		}
		for j := i; j < N; j += i {
			primeTable[j] = true
		}
		primeTable[i] = false
	}
}

type RequestObject struct {
	Method string          `json:"method"`
	Number json.RawMessage `json:"number"`
}

type ResponseObject struct {
	Method string `json:"method"`
	Prime  bool   `json:"prime"`
}

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

func sendResponse(w io.Writer, isPrime bool) {
	response := ResponseObject{
		Method: "isPrime",
		Prime:  isPrime,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println(err)
	}
}

func handleConnection(conn net.Conn) {
	defer func() {
		conn.Close()
		log.Printf("closed: %s", conn.RemoteAddr())
	}()
	log.Printf("accepted: %s", conn.RemoteAddr())

	reader := bufio.NewReader(conn)
	for {
		line, err := reader.ReadBytes('\n')
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Println(err)
			return
		}
		log.Println(string(bytes.TrimRight(line, "\n")))

		var req RequestObject
		if err := json.Unmarshal(line, &req); err != nil {
			log.Println("invalid request json")
			conn.Write([]byte("invalid request"))
			break
		}
		if req.Method != "isPrime" {
			log.Println("invalid method")
			conn.Write([]byte("invalid request"))
			break
		}

		if bytes.ContainsAny(req.Number, "\"'") {
			log.Println("contains quote")
			conn.Write([]byte("invalid request"))
			break
		}

		var numString json.Number
		if err := json.Unmarshal(req.Number, &numString); err != nil {
			log.Println("contains quote")
			conn.Write([]byte("invalid request"))
			break
		}

		isPrime := false
		num, err := numString.Int64()
		if err == nil {
			isPrime = checkPrime(num)
		}
		sendResponse(conn, isPrime)
	}
}

func checkPrime(num int64) bool {
	if num < 2 {
		return false
	}
	return !primeTable[num]
}

func main() {
	port := 1234

	if err := do(port); err != nil {
		log.Fatal(err)
	}
}
