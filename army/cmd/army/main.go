package main

import (
	"fmt"
	"os"

	"github.com/smahovkic/agent-army/army/internal/cli"
)

func main() {
	if err := cli.NewRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
