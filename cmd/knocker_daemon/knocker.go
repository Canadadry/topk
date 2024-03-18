package knocker_daemon

import (
	"app/knocker"
	"app/pcap"
	"fmt"
	"time"
)

const Action = "knockerd"

func Run(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("Usage: app knocker <interface>")
	}
	iface := args[0]

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

	return pcap.Run(iface, func(ip string, port uint16) {
		user, ok := tracker.CheckSequence(ip, port, time.Now())
		if ok {
			fmt.Println(user, "has validated the challenge")
		}
	})

}
