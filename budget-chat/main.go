package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
)

type User struct {
	conn   net.Conn
	reader *bufio.Reader
	Name   string
}

func (u *User) readLine() (string, error) {
	line, err := u.reader.ReadBytes('\n')
	if err != nil {
		return "", err
	}
	reply := strings.TrimRight(string(line), "\r\n")

	return string(reply), nil
}

type Server struct {
	members []*User
}

func (s *Server) addNonameUser(conn net.Conn) (*User, error) {
	_, err := conn.Write([]byte("Welcome to budgetchat! What shall I call you?\n"))
	if err != nil {
		return nil, err
	}

	reader := bufio.NewReader(conn)
	user := &User{
		conn:   conn,
		reader: reader,
	}

	reply, err := user.readLine()
	if err != nil {
		return nil, err
	}
	if len(reply) > 16 {
		return nil, errors.New("name must less than or equal to 16")
	}

	user.Name = reply
	if err := s.addMember(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Server) addMember(user *User) error {
	memberList := make([]string, 0, len(s.members))
	for _, member := range s.members {
		memberList = append(memberList, member.Name)

		if member.Name == user.Name {
			return errors.New("name must be unique")
		}
	}
	s.sendTo(fmt.Sprintf("* The room contains: %s\n", strings.Join(memberList, ", ")), user)

	s.members = append(s.members, user)

	s.broadcast(fmt.Sprintf("* %s has entered the room\n", user.Name), user.Name)
	return nil
}

func (s *Server) sendTo(message string, toUser *User) {
	log.Printf("[[send]]: (%s) %s", toUser.Name, message)
	fmt.Fprint(toUser.conn, message)
}

func (s *Server) broadcast(message string, fromUserName string) {
	for _, user := range s.members {
		if fromUserName != "" && user.Name == fromUserName {
			continue
		}
		log.Printf("[[broadcast]]: (%s -> %s) %s", fromUserName, user.Name, message)

		_, err := user.conn.Write([]byte(message))
		if err != nil {
			log.Println(err)
		}
	}
}

func (s *Server) disconnectUser(user *User) error {
	defer user.conn.Close()

	idx := -1
	for i, member := range s.members {
		if member.Name == user.Name {
			idx = i
		}
	}
	if idx == -1 {
		return nil
	}

	s.members[idx] = s.members[len(s.members)-1]
	s.members = s.members[:len(s.members)-1]

	s.broadcast(fmt.Sprintf("* %s has left the room\n", user.Name), user.Name)

	return nil
}

func (s *Server) handleConnection(conn net.Conn) {
	log.Printf("accepted: %s", conn.RemoteAddr())

	user, err := s.addNonameUser(conn)
	if err != nil {
		log.Println(err)
		conn.Close()
		return
	}

	defer func() {
		s.disconnectUser(user)
		log.Printf("closed: %s, [%s]", conn.RemoteAddr(), user.Name)
	}()

	for {
		line, err := user.readLine()
		if err != nil {
			log.Println(err)
			return
		}

		s.broadcast(fmt.Sprintf("[%s] %s\n", user.Name, line), user.Name)
	}
}

func do(port int) error {
	log.Printf("listen :%d ", port)
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return err
	}

	server := &Server{
		members: make([]*User, 0),
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go server.handleConnection(conn)
	}
}

func main() {
	port := 1234

	if err := do(port); err != nil {
		log.Fatal(err)
	}
}
