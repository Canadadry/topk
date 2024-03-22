package listen

import (
	"context"
	"net"
	"testing"
	"time"
)

type MockProcessPacket struct {
	Calls []PacketInfo
}

func (m *MockProcessPacket) ProcessPacket(srcIP string, port uint16) {
	m.Calls = append(m.Calls, PacketInfo{Port: port, SrcIP: srcIP})
}

func TestServer(t *testing.T) {
	startPort, endPort := uint16(49152), uint16(49153)
	mockProcessor := &MockProcessPacket{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		err := Server(ctx, startPort, endPort, mockProcessor.ProcessPacket)
		if err != nil {
			t.Errorf("Server exited with error: %v", err)
		}
	}()

	time.Sleep(time.Second)

	conn, err := net.Dial("udp", net.JoinHostPort("localhost", "49152"))
	if err != nil {
		t.Fatalf("Failed to dial UDP server: %v", err)
	}
	defer conn.Close()

	_, err = conn.Write([]byte("test"))
	if err != nil {
		t.Fatalf("Failed to write to UDP server: %v", err)
	}

	time.Sleep(time.Second)

	if len(mockProcessor.Calls) == 0 {
		t.Errorf("Expected ProcessPacket to be called at least once, but it wasn't")
	}

	cancel()
	time.Sleep(time.Second)
}
