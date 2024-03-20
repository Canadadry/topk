package sniff

import (
	"app/listen"
	"fmt"
	"time"
)

const Action = "sniff"

func Run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("Usage: app sniff <interface>")
	}
	iface := args[0]

	return listen.Pcap(iface, func(ip string, port uint16) {
		fmt.Println("rvc packet at", port, "from", ip, "at", time.Now())
	})
}
