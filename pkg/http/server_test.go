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

	addr, err := server.GetAddr()
	if err != nil {
		t.Fatalf("could not get server address: %v", err)
	}

	// Simulate a client connecting to the server
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("could not connect to server: %v", err)
	}

	t.Log("client connected to server")

	// Test response
	_, err = conn.Write([]byte("GET / HTTP/1.0\r\n\r\n"))
	if err != nil {
		t.Fatalf("could not write to connection: %v", err)
	}

	if err := conn.Close(); err != nil {
		t.Logf("could not close connection: %v", err)
	}

	err = server.Shutdown()
	if err != nil {
		t.Fatalf("could not shutdown server: %v", err)
	}

	// Wait for the server to shutdown
	wg.Wait()
}

func TestServerGetAddr(t *testing.T) {
	server := NewServer("localhost", "8080")

	// Start the server
	go func() {
		err := server.ListenAndServe()
		if err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
			t.Errorf("unexpected server error: %v", err)
		}
	}()

	addr, err := server.GetAddr()
	if err != nil {
		t.Fatalf("could not get server address: %v", err)
	}

	if !strings.HasPrefix(addr, "127.0.0.1:8080") {
		t.Errorf("unexpected server address: %v", addr)
	}

	err = server.Shutdown()
	if err != nil {
		t.Fatalf("could not shutdown server: %v", err)
	}
}

func TestListenAndServerInitializesListener(t *testing.T) {
	server := NewServer("localhost", "8080")

	go func() {
		err := server.ListenAndServe()
		if err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
			t.Errorf("unexpected server error: %v", err)
		}
	}()

	addr, err := server.GetAddr()
	if err != nil {
		t.Fatalf("could not get server address: %v", err)
	}

	if !strings.HasPrefix(addr, "127.0.0.1:8080") {
		t.Errorf("unexpected server address: %v", addr)
	}

	// Wait for the server to start
	time.Sleep(100 * time.Millisecond)

	if server.listener == nil {
		t.Error("server listener not initialized")
	}

	if err := server.Shutdown(); err != nil {
		t.Fatalf("could not shutdown server: %v", err)
	}
}

func TestListenAndServeHandlesBindError(t *testing.T) {
	// Create a server and bind to a specific port
	server1 := NewServer("localhost", "0")

	// Start first server to occupy the port
	go func() {
		err := server1.ListenAndServe()
		if err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
			t.Errorf("First server error: %v", err)
		}
	}()

	// Wait for first server to be ready
	addr, err := server1.GetAddr()
	if err != nil {
		t.Fatalf("Failed to get first server address: %v", err)
	}

	// Parse the port from the first server's address
	_, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("Failed to split host/port: %v", err)
	}

	// Create a second server trying to bind to the same port
	server2 := NewServer("localhost", portStr)

	// Try to start second server on same port - should fail
	err = server2.ListenAndServe()

	// Verify we got the expected error
	if err == nil {
		t.Fatal("Expected error when binding to in-use port, got nil")
	}

	// Verify error is related to address already in use
	if !strings.Contains(err.Error(), "address already in use") &&
		!strings.Contains(err.Error(), "in use") {
		t.Errorf("Expected 'address already in use' error, got: %v", err)
	}

	// Clean up
	if err := server1.Shutdown(); err != nil {
		t.Fatalf("Failed to shutdown first server: %v", err)
	}
}
