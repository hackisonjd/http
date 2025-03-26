package http

import (
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
}

func NewServer(host, port string) *Server {
	return &Server{
		host: host,
		port: port,
	}
}

func (s *Server) ListenAndServe() error {
	// combine host and port into address
	addr := net.JoinHostPort(s.host, s.port)

	// Create a listener, surrounded by a mutex to prevent problems
	var err error
	s.mu.Lock()
	s.listener, err = net.Listen("tcp", addr)
	s.mu.Unlock()
	if err != nil {
		return err
	}

	// we were able to bind to the address, open a server and wait for connections
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return err
		}

		go s.handleConnection(conn)
	}
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
	defer conn.Close()
	// Read connection data
	buf := make([]byte, 1024)
	_, err := conn.Read(buf)
	if err != nil {
		return
	}

	// Print request data
	response := "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: 13\r\n\r\nHello, World!"

	// Send response back to client
	conn.Write([]byte(response))
}
