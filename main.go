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
			fmt.Println("[error] acceptLoop", err.Error())
			continue
		}

		fmt.Printf("[client] %s connected", conn.RemoteAddr().String())

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
				fmt.Printf("[client] %s close the connection\n", conn.RemoteAddr().String())
				break
			}
			fmt.Println("[error] readLoop", err.Error())
			continue
		}

		s.msgch <- Message{
			from:    conn.RemoteAddr().String(),
			payload: buf[:n],
		}

		conn.Write([]byte("[response]: OK!"))
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
	close(s.msgch)

	return nil
}

func main() {
	s := NewServer("localhost:3000")
	fmt.Printf("[listen] %s\n", s.listenAddr)
	go func() {
		for msg := range s.msgch {
			fmt.Printf("[message] from %s : %s\n", msg.from, string(msg.payload))
		}
	}()
	log.Fatal(s.Start())
}
