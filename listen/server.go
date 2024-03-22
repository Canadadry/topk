package listen

import (
	"context"
	"fmt"
	"net"
)

type PacketInfo struct {
	SrcIP string
	Port  uint16
}

func Server(ctx context.Context, startPort, endPort uint16, processPacket ProcessPacket) error {
	packetChan := make(chan PacketInfo)
	errChan := make(chan error, 1)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		for {
			select {
			case packetInfo := <-packetChan:
				processPacket(packetInfo.SrcIP, packetInfo.Port)
			case <-ctx.Done():
				fmt.Println("Packet processing goroutine shutting down...")
				return
			}
		}
	}()

	listenOnPort := func(ctx context.Context, port uint16, packetChan chan<- PacketInfo, errChan chan<- error) {
		addr := net.UDPAddr{
			Port: int(port),
			IP:   net.ParseIP("0.0.0.0"),
		}
		conn, err := net.ListenUDP("udp", &addr)
		if err != nil {
			errChan <- fmt.Errorf("error setting up listener on port %d: %v", port, err)
			return
		}
		defer conn.Close()

		buffer := make([]byte, 1024)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				_, remoteAddr, err := conn.ReadFromUDP(buffer)
				if err != nil {
					fmt.Printf("Error reading from UDP port %d: %v\n", port, err)
					continue
				}
				srcIP := remoteAddr.IP.String()
				packetChan <- PacketInfo{Port: port, SrcIP: srcIP}
			}
		}
	}

	for port := startPort; port <= endPort; port++ {
		go listenOnPort(ctx, port, packetChan, errChan)
	}

	select {
	case <-ctx.Done():
		if ctx.Err() != context.Canceled {
			return ctx.Err()
		}
		return nil
	case err := <-errChan:
		cancel()
		return err
	}
}
