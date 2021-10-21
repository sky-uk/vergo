package main

import (
	"github.com/sky-uk/umc-shared/vergo/cmd"
	"os"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
