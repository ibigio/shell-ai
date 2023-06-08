package main

import (
	"q/cli"
)

func main() {
	if err := cli.RootCmd.Execute(); err != nil {
		panic(err)
	}
}
