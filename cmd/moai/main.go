// @MX:ANCHOR: [AUTO] main is the entry point of moai CLI. Returns exit code 1 on error,
// or a custom code when the error implements ExitCoder (used by `moai worktree verify`
// to surface 0/1/2/3 to the orchestrator per SPEC-V3R3-CI-AUTONOMY-001 Wave 5).
// @MX:REASON: Sole entry point of the executable binary; delegates CLI command execution
package main

import (
	"errors"
	"os"

	"github.com/modu-ai/moai-adk/internal/cli"
)

// ExitCoder lets a subcommand surface a non-default exit code through the cobra
// error chain. Used by `moai worktree verify` (0=clean, 1=divergence,
// 2=suspect, 3=both).
type ExitCoder interface {
	ExitCode() int
}

func main() {
	if err := cli.Execute(); err != nil {
		var ec ExitCoder
		if errors.As(err, &ec) {
			os.Exit(ec.ExitCode())
		}
		os.Exit(1)
	}
}
