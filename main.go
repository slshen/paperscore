package main

import (
	"os"

	"github.com/slshen/sb/cmd"
)

func main() {
	err := cmd.Root().Execute()
	if err != nil {
		os.Exit(2)
	}
}
