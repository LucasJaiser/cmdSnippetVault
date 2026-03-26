package main

import (
	"fmt"
	"lucasjaiser/goSnipperVault/internal/cli"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cli.SetVersion(version, commit, date)
	cli.Init()
	err := cli.Execute()
	if err != nil {
		fmt.Printf("Error in execution: %s", err.Error())
	}
}
