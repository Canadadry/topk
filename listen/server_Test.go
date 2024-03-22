package listen

import (
	"context"
	"fmt"
	"net"
	"reflect"
	"testing"
	"time"
)

type MockProcessPacket struct {
	PortCalled map[int]int
	T          *testing.T
}

func (m *MockProcessPacket) ProcessPacket(srcIP string, port uint16) {
	m.T.Logf("ProcessPacket called with %v", port)
	m.PortCalled[int(port)] = m.PortCalled[int(port)] + 1
}

func (m *MockProcessPacket) Reset() {
	m.PortCalled = map[int]int{}
}

func TestServer(t *testing.T) {
	startPort := 12000
	nbPortOpen := 10
	tests := map[string]struct {
		in        map[int]int
		expect    map[int]int
		goroutine int
		duration  time.Duration
	}{
		"one packet one goroutine on open port": {
			in: map[int]int{
				startPort: 1,
			},
			expect: map[int]int{
				startPort: 1,
			},
			goroutine: 1,
			duration:  100 * time.Millisecond,
		},
		"one packet one goroutine on closed port": {
			in: map[int]int{
				startPort - 1: 1,
			},
			expect:    map[int]int{},
			goroutine: 1,
			duration:  100 * time.Millisecond,
		},
		"multiple packet one goroutine on open and closed port": {
			in: map[int]int{
				startPort:     50,
				startPort - 1: 50,
			},
			expect: map[int]int{
				startPort: 50,
			},
			goroutine: 1,
			duration:  100 * time.Millisecond,
		},
		"one packet per goroutine on open port": {
			in: map[int]int{
				startPort:     50,
				startPort + 1: 50,
				startPort + 2: 50,
			},
			expect: map[int]int{
				startPort:     50,
				startPort + 1: 50,
				startPort + 2: 50,
			},
			goroutine: 150,
			duration:  100 * time.Millisecond,
		},
	}

	mpp := &MockProcessPacket{T: t}
	close := RunServer(t, uint16(startPort), uint16(startPort+nbPortOpen), mpp.ProcessPacket)
	defer close()
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mpp.Reset()
			testPacketStream(t, tt.in, tt.goroutine, tt.duration)
			if !reflect.DeepEqual(tt.expect, mpp.PortCalled) {
				t.Fatalf("failed want %#v got %#v", tt.expect, mpp.PortCalled)
			}
		})
	}
}

func testPacketStream(t *testing.T, ports map[int]int, goroutine int, duration time.Duration) {
	t.Helper()
	call := func(t *testing.T, port uint16) {
		t.Logf("sending packet to %v", port)
		conn, err := net.Dial("udp", net.JoinHostPort("localhost", fmt.Sprintf("%d", port)))
		if err != nil {
			t.Fatalf("Failed to dial UDP server: %v", err)
		}
		defer conn.Close()
		_, err = conn.Write([]byte("test"))
		if err != nil {
			t.Fatalf("Failed to write to UDP server: %v", err)
		}
	}

	c := make(chan int)
	for i := range goroutine {
		go func() {
			t.Logf("spawning goroutine %v", i)
			for port := range c {
				call(t, uint16(port))
			}
		}()
	}

	for port, count := range ports {
		for range count {
			c <- port
		}
	}

	close(c)
	time.Sleep(duration)
}

func RunServer(t *testing.T, startPort, endPort uint16, processPacket ProcessPacket) func() {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		err := Server(ctx, startPort, endPort, processPacket)
		if err != nil {
			t.Errorf("Server exited with error: %v", err)
		}
	}()

	return cancel
}
