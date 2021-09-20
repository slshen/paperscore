package main

import (
	"log"

	"github.com/slshen/sb/cmd"
)

func main() {
	err := cmd.Root().Execute()
	if err != nil {
		log.Fatal(err)
	}
}
