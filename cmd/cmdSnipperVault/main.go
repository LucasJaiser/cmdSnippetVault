package main

import (
	"fmt"
	"lucasjaiser/goSnipperVault/internal/cli"
)

func main() {
	cli.Init()
	err := cli.Execute()
	if err != nil {
		fmt.Printf("Error in execution: %s", err.Error())
	}
}
