package sniff

import (
	"fmt"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

const Action = "sniff"

func Run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("Usage: app sniff <interface>")
	}
	iface := args[0]

	// Open device
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
		fmt.Println("Got a UDP packet:")
		printPacketInfo(packet)
	}
	return nil
}

func printPacketInfo(packet gopacket.Packet) {
	// Print packet timestamp
	fmt.Printf("Timestamp: %s\n", packet.Metadata().Timestamp.Format(time.RFC3339))

	// Print the source and destination IP addresses and ports
	if ipLayer := packet.Layer(layers.LayerTypeIPv4); ipLayer != nil {
		ip, _ := ipLayer.(*layers.IPv4)
		fmt.Printf("From %s to %s\n", ip.SrcIP, ip.DstIP)
	}

	if udpLayer := packet.Layer(layers.LayerTypeUDP); udpLayer != nil {
		udp, _ := udpLayer.(*layers.UDP)
		fmt.Printf("From port %d to port %d\n", udp.SrcPort, udp.DstPort)
	}

	// Print packet payload
	if applicationLayer := packet.ApplicationLayer(); applicationLayer != nil {
		fmt.Println("Payload:")
		fmt.Printf("%s\n", applicationLayer.Payload())
	}

	fmt.Println("---")
}
