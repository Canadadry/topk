package listen

import (
	"context"
	"net"
	"testing"
	"time"
)

// MockProcessPacket keeps track of received packets
type MockProcessPacket struct {
	Calls []PacketInfo // Records calls to ProcessPacket
}

func (m *MockProcessPacket) ProcessPacket(srcIP string, port int) {
	m.Calls = append(m.Calls, PacketInfo{Port: port, SrcIP: srcIP})
}

func TestServer(t *testing.T) {
	// Setup
	startPort, endPort := uint16(49152), uint16(49153) // Small range for testing
	mockProcessor := &MockProcessPacket{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensure cancellation at the end of the test

	go func() {
		err := Server(ctx, startPort, endPort, mockProcessor.ProcessPacket)
		if err != nil {
			t.Errorf("Server exited with error: %v", err)
		}
	}()

	// Give the server a moment to start
	time.Sleep(time.Second)

	// Simulate a client sending a UDP packet
	conn, err := net.Dial("udp", net.JoinHostPort("localhost", "49152"))
	if err != nil {
		t.Fatalf("Failed to dial UDP server: %v", err)
	}
	defer conn.Close()

	_, err = conn.Write([]byte("test"))
	if err != nil {
		t.Fatalf("Failed to write to UDP server: %v", err)
	}

	// Give the server time to process the packet
	time.Sleep(time.Second)

	// Test that the packet was processed
	if len(mockProcessor.Calls) == 0 {
		t.Errorf("Expected ProcessPacket to be called at least once, but it wasn't")
	}

	// Test server shutdown
	cancel() // Trigger cancellation

	// Allow some time for shutdown
	time.Sleep(time.Second)

	// Additional assertions can be made here if there are observable effects of shutdown.
	// For example, if the server releases resources or logs a message, verify that behavior.
}
