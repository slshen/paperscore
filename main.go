package main

import (
	"os"

	"github.com/slshen/paperscore/cmd"
)

func main() {
	err := cmd.Root().Execute()
	if err != nil {
		os.Exit(2)
	}
}
