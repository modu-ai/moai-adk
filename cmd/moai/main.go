// @MX:ANCHOR: [AUTO] main is the entry point of moai CLI. Returns exit code 1 on error.
// @MX:REASON: Sole entry point of the executable binary; delegates CLI command execution
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
