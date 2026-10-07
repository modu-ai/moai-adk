// @MX:ANCHOR: [AUTO] main is the entry point of moai CLI. Returns exit code 1 on error,
// or a custom code when the error carries an intentional ExitCoder (used by `moai worktree verify`
// to surface 0/1/2/3 to the orchestrator per SPEC-V3R3-CI-AUTONOMY-001 Wave 5).
// @MX:REASON: Sole entry point of the executable binary; delegates CLI command execution
package main

import (
	"os"

	"github.com/modu-ai/moai-adk/internal/bugreport"
	"github.com/modu-ai/moai-adk/internal/cli"
)

func main() {
	// SPEC-FEEDBACK-PARTICIPATION-001 (REQ-ANON-006): the main goroutine's
	// deferred recover is the pipeline's panic emit site. Capture records the
	// signal (fail-open, consent-gated, user-scoped), then re-panics so the
	// runtime's crash output and exit code survive untouched.
	defer func() {
		if r := recover(); r != nil {
			bugreport.Capture(bugreport.KindPanic, nil, "", nil)
			panic(r)
		}
	}()
	if err := cli.Execute(); err != nil {
		// ResolveExitCode rejects raw *exec.ExitError: a wrapped subprocess
		// failure must fall through to the default exit 1, never adopt the
		// subprocess's own code (card t130 — the rc=128 silent failures).
		if code, ok := cli.ResolveExitCode(err); ok {
			os.Exit(code)
		}
		os.Exit(1)
	}
}
