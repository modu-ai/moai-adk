package main

import (
	"os"

	"github.com/anthropics/moai-adk-go/internal/cli"
)

func main() {
	cmd := cli.NewRootCommand()
	if err := cmd.Execute(); err != nil {
		// Cobra already prints the error
		// Exit with non-zero code
		os.Exit(1)
	}
}
