package main

import (
	"app/cmd/knocker"
	"app/cmd/knocker_daemon"
	"app/cmd/sniff"
	"fmt"
	"os"
	"strings"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error : ", err)
		os.Exit(1)
	}
}

func run(args []string) error {

	actions := map[string]func([]string) error{
		sniff.Action:          sniff.Run,
		knocker_daemon.Action: knocker_daemon.Run,
		knocker.Action:        knocker.Run,
	}

	listOfAction := make([]string, 0, len(actions))
	for a := range actions {
		listOfAction = append(listOfAction, a)
	}
	invalidErrAction := fmt.Errorf("invalid action :\n\tapp action [args], possible actions are [%s]", strings.Join(listOfAction, " | "))

	if len(args) == 0 {
		return invalidErrAction
	}

	run, ok := actions[args[0]]
	if !ok {
		return invalidErrAction
	}
	return run(args[1:])
}
