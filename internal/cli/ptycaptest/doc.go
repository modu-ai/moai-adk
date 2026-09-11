// Package ptycaptest is a test-only pty capture harness. It runs a package's
// own test binary — built with `go test -c` — inside a detached tmux session
// sized 80x30 and reads the painted screen with `tmux capture-pane -p`. That is
// the only place a render defect is observed the way a terminal shows it;
// View() goldens are the regression guard, the pty capture is the repair
// verdict. The package also carries the real-HOME watch list and the render
// helpers (ANSI stripping, display-width columns, golden comparison) the
// capture verdicts use.
//
// Production code MUST NOT import this package. It shells out to tmux and go,
// reads the real HOME, and calls into package testing. The rule is checked,
// not only stated: TestNoProductionImport lists every package in the module
// with `go list` and fails when a non-_test.go file imports this package.
// This package imports neither internal/cli nor any package under it, so a
// test in internal/cli or internal/cli/wizard can import it without a cycle.
//
// Contract for a consuming package:
//   - Every capture test calls Gate first. Without MOAI_PTY_CAPTURE=1 it
//     skips; with the gate on and tmux missing it fails, never skips.
//   - The package defines func TestPtyCaptureChild(t *testing.T) — the name
//     in ChildTestName. It skips when os.Getenv(ChildEnv) is empty, calls
//     RecordEnv before any product code, then runs the case named by
//     os.Getenv(ChildEnv) on its TTY.
//   - BuildChild compiles the package named by its argument. Callers pass ".":
//     `go test` runs a test binary in its own package directory, so "." is
//     the calling package.
//
// Safety contract:
//   - Every session name comes from SessionName (prefix moai-ptycap-), and
//     every session is killed by its exact name from a Cleanup registered
//     right after creation. No kill-server, no prefix sweep.
//   - The child environment is passed variable by variable with tmux -e and
//     then observed: the child records what it actually received, and the
//     parent checks that record (VerifyChildEnv) instead of trusting the -e
//     flags.
//   - Every subprocess is bounded by a context deadline, and the child test
//     binary by -test.timeout.
//   - The harness never writes into the repository on its own: captures go
//     to MOAI_PTY_CAPTURE_OUT when set, otherwise to the test's temp dir.
package ptycaptest
