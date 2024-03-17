package main

import (
	"app/knocker"
	"fmt"
	"strconv"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

const Action = "knocker"

func Run(args []string) error {
	if len(args) > 5 {
		fmt.Errorf("Usage: app knocker <interface> <port1> <port2> <port3> <port4>")
	}
	iface := args[1]
	ports := make([]uint16, 4)
	for i := 0; i < 4; i++ {
		port, err := strconv.Atoi(args[i+1])
		if err != nil || port < 1 || port > 65535 {
			return fmt.Errorf("Invalid port number: %s", args[i+1])
		}
		ports[i] = uint16(port)
	}

	// Open device
	handle, err := pcap.OpenLive(iface, 1600, true, pcap.BlockForever)
	if err != nil {
		return fmt.Errorf("cannot open '%s' : %w", iface, err)
	}
	defer handle.Close()

	var filter = "udp"
	err = handle.SetBPFFilter(filter)
	if err != nil {
		return fmt.Errorf("cannot filter with 'udp': %w", err)
	}
	fmt.Printf("Monitoring for port knocking sequence on %s: %v\n", iface, ports)
	// Initialize the sequence tracker with the provided ports
	tracker := knocker.NewUserKnocker(knocker.UserKnockerConfig{
		Users: map[string]string{
			"admin": "password",
		},
		SequenceLen:              10,
		MinPort:                  16000,
		MaxPort:                  65535,
		SequenceIntervalInSecond: 15,
		Timeout:                  10 * time.Second,
		MinInterval:              time.Second,
	})

	// Use the handle as a packet source to process all packets
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	packetSource.DecodeOptions.Lazy = true
	packetSource.DecodeOptions.NoCopy = true

	for packet := range packetSource.Packets() {
		// Process packet here
		if udpLayer := packet.Layer(layers.LayerTypeUDP); udpLayer != nil {
			udp, _ := udpLayer.(*layers.UDP)
			srcIP := packet.NetworkLayer().NetworkFlow().Src().String()
			dstPort := udp.DstPort

			// Check if the current packet is part of the sequence
			if user, ok := tracker.CheckSequence(srcIP, uint16(dstPort), time.Now()); ok {
				fmt.Printf("%s has completed the sequence from %s \n", user, srcIP)
			}
		}
	}
	return nil
}
