package http

import (
	"net"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestServerAcceptsConnections(t *testing.T) {
	server := NewServer("localhost", "0")

	// Create a waiting group to coordinate shutdowns
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		err := server.ListenAndServe()
		if err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
			// use of closed network connection is expected when the server is shutdown
			t.Errorf("unexpected server error: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	addr := server.listener.Addr().String()

	// Simulate a client connecting to the server
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("could not connect to server: %v", err)
	}

	t.Log("client connected to server")
	conn.Close()

	err = server.Shutdown()
	if err != nil {
		t.Fatalf("could not shutdown server: %v", err)
	}

	// Wait for the server to shutdown
	wg.Wait()
}
