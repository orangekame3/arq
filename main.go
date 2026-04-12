package main

import (
	"os"

	"github.com/orangekame3/arq/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
