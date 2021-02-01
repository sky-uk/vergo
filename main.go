package main

import (
	"os"
	"sky.uk/vergo/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
