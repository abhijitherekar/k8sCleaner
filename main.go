package main

import (
	"fmt"
	"github.com/k8sCleanner/cmd"
	"os"
)

func main() {

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
