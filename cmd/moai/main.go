package main

import (
	"os"

	"github.com/modu-ai/moai-adk/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
