package main

import (
	"fmt"
	"io"
	"log"
	"net"
)

type (
	Message struct {
		from    string
		payload []byte
	}

	Server struct {
		listenAddr string
		ln         net.Listener
		quitch     chan struct{}
		msgch      chan Message
	}
)

func NewServer(listenAddr string) *Server {
	return &Server{
		listenAddr: listenAddr,
		quitch:     make(chan struct{}),
		msgch:      make(chan Message, 10), // make 10 -> non blocant sinon figé en attente de message
	}
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			fmt.Println("accept error:", err.Error())
			continue
		}

		fmt.Println("connection accepted for ", conn.RemoteAddr().String())
		go s.readLoop(conn)
	}
}

func (s *Server) readLoop(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 2048)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			// Used to check if the client close the connection
			if err == io.EOF {
				fmt.Println("client closed", conn.RemoteAddr().String())
				break
			}
			fmt.Println("read error:", err.Error())
			continue
		}

		msg := Message{
			from:    conn.RemoteAddr().String(),
			payload: buf[:n],
		}
		fmt.Println("[message]", msg.from, string(msg.payload))
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return err
	}
	defer ln.Close()

	s.ln = ln

	go s.acceptLoop()

	<-s.quitch

	return nil
}

func main() {
	s := NewServer("localhost:3000")
	fmt.Println("[listen]", s.listenAddr)
	log.Fatal(s.Start())
}
