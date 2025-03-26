package http

import (
	"fmt"
	"net"
	"sync"
)

// Client sends a request, including
// - the request method (GET, POST etc etc)
// - the target URI
// - headers
// - optional body

// Server code
type Server struct {
	host     string
	port     string
	listener net.Listener // listener for the server
	mu       sync.Mutex   // mutex to protect the listener
	ready    chan struct{}
}

func NewServer(host, port string) *Server {
	return &Server{
		host:  host,
		port:  port,
		ready: make(chan struct{}),
	}
}

func (s *Server) ListenAndServe() error {
	// combine host and port into address
	addr := net.JoinHostPort(s.host, s.port)

	// Create a listener, surrounded by a mutex to prevent problems
	var err error

	s.mu.Lock()
	s.listener, err = net.Listen("tcp", addr)
	if err != nil {
		s.mu.Unlock()
		return err
	}
	s.mu.Unlock()

	// Signal that the server is ready
	close(s.ready)

	// we were able to bind to the address, open a server and wait for connections
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return err
		}

		go s.handleConnection(conn)
	}
}

func (s *Server) GetAddr() (string, error) {
	<-s.ready

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.listener == nil {
		return "", fmt.Errorf("server not started")
	}

	return s.listener.Addr().String(), nil
}

func (s *Server) Shutdown() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

func (s *Server) handleConnection(conn net.Conn) {
	defer func() {
		if err := conn.Close(); err != nil {
			fmt.Printf("could not close connection: %v\n", err)
		}
	}()
	// Read connection data
	buf := make([]byte, 1024)
	_, err := conn.Read(buf)
	if err != nil {
		return
	}

	// Error check for conn.Close()

	// Print request data
	response := "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: 13\r\n\r\nHello, World!"

	// Send response back to client
	if _, err := conn.Write([]byte(response)); err != nil {
		fmt.Printf("could not write response: %v\n", err)
		return
	}
}
