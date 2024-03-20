package listen

import (
	"fmt"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

type ProcessPacket func(ip string, port uint16)

func Pcap(iface string, processPacket ProcessPacket) error {

	handle, err := pcap.OpenLive(iface, 1600, true, pcap.BlockForever)
	if err != nil {
		return fmt.Errorf("cannot open '%s' : %w", iface, err)
	}
	defer handle.Close()

	// Set filter
	var filter = "udp"
	err = handle.SetBPFFilter(filter)
	if err != nil {
		return fmt.Errorf("cannot filter with 'udp': %w", err)
	}
	fmt.Printf("Capturing UDP packets on %s\n", iface)

	// Use the handle as a packet source to process all packets
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	packetSource.DecodeOptions.Lazy = true
	packetSource.DecodeOptions.NoCopy = true

	for packet := range packetSource.Packets() {
		// Process packet here
		// fmt.Println("Got a UDP packet:")
		ip, port := extractPacketInfo(packet)
		processPacket(ip, port)
	}
	return nil
}

func extractPacketInfo(packet gopacket.Packet) (string, uint16) {
	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	if ipLayer == nil {
		return "", 0
	}
	ip, ok := ipLayer.(*layers.IPv4)
	if !ok {
		return "", 0
	}

	udpLayer := packet.Layer(layers.LayerTypeUDP)
	if udpLayer == nil {
		return ip.SrcIP.String(), 0
	}
	udp, ok := udpLayer.(*layers.UDP)
	if !ok {
		return ip.SrcIP.String(), 0
	}

	return ip.SrcIP.String(), uint16(udp.DstPort)
}
