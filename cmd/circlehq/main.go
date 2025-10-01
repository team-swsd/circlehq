package main

import (
	"fmt"
	"os"

	"github.com/team-swsd/circlehq/internal/cmd"
)

var (
	version = "v0.0.1"
)

func main() {
	c := cmd.NewCircleHQCmd(version)
	if err := c.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(-1)
	}
}
