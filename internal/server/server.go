package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/bmamha/httpfromtcp/internal/request"
	"github.com/bmamha/httpfromtcp/internal/response"
)

type Server struct {
	listener net.Listener
	closed   atomic.Bool
	handler  Handler
}

type Handler func(w *response.Writer, r *request.Request)

func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	server := &Server{
		listener: listener,
		handler:  handler,
	}
	go server.listen()

	return server, nil
}

func (s *Server) Close() error {
	s.closed.Store(true)
	return s.listener.Close()
}

func (s *Server) listen() {
	for {
		if s.closed.Load() {
			return
		}
		conn, err := s.listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	w := response.NewWriter(conn)
	r, err := request.RequestFromReader(conn)

	log.Printf("request parsed for %s]n", r.RequestLine.RequestTarget)
	if err != nil {
		fmt.Println("handling bad request")
		w.WriteStatusLine(400)
		w.Headers.Set("Content-Type", "text/html")
		w.Write([]byte("error: request not parsed"))
		return
	}

	s.handler(w, r)
}
