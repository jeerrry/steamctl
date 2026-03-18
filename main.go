package main

import (
	"os"

	"github.com/jeerrry/steamctl/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
