package knocker

import (
	"app/knocker"
	"app/pkg/publicip"
	"flag"
	"fmt"
	"net"
	"time"
)

const Action = "knock"

func Run(args []string) error {
	var server string
	var sleep time.Duration
	var sequenceIntervalInSecond int
	var secret string
	var sequenceLen int
	var minPort int
	var maxPort int

	fs := flag.NewFlagSet("app", flag.ContinueOnError)
	fs.StringVar(&server, "to", server, "server to knock on")
	fs.DurationVar(&sleep, "sleep", sleep, "duration between knock")
	fs.IntVar(&sequenceIntervalInSecond, "rotation", sequenceIntervalInSecond, "duration between sequence rotation in second")
	fs.StringVar(&secret, "secret", secret, "secret to generate sequence")
	fs.IntVar(&sequenceLen, "len", sequenceLen, "knocking sequence len")
	fs.IntVar(&minPort, "min", minPort, "knocking sequence minimum port")
	fs.IntVar(&maxPort, "max", maxPort, "knocking sequence maximum port")
	err := fs.Parse(args)
	if err != nil {
		return err
	}

	srcIP, err := publicip.Get()
	if err != nil {
		return fmt.Errorf("cannot get public ip : %w", err)
	}

	provider := knocker.TimeBasedSequenceProvider{
		MinPort:          uint16(minPort),
		MaxPort:          uint16(maxPort),
		Salt:             secret,
		IntervalInSecond: sequenceIntervalInSecond,
		SequenceLen:      sequenceLen,
	}

	// ensure DNS lookup cached or first ports may not be knocked
	_, err = net.LookupHost(server)
	if err != nil {
		return fmt.Errorf("lookup host failed '%s' : %w", server, err)
	}

	sequence := provider.GetSequence(srcIP, time.Now())

	for _, port := range sequence {
		delay := time.NewTicker(sleep)

		addr := fmt.Sprintf("%s:%d", server, port)
		timeout := time.Second
		con, _ := net.DialTimeout("udp", addr, timeout)
		if con != nil {
			con.Close()
		}

		<-delay.C
	}
	return nil
}
