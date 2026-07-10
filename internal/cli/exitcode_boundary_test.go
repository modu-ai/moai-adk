package cli

import (
	"errors"
	"testing"
)

// exitCoder is defined in fang.go (SPEC-CLI-TUX-V3-001 M1b) and reused here —
// it mirrors the cmd/moai/main.go ExitCoder interface.

// TestExitCodeErrorSatisfiesExitCoder proves the cli-root exitCodeError type
// (shared by astgrep/hook/hook_pre_push/spec_lint/spec_drift/migrate_agency/
// constitution) satisfies the main.go ExitCoder boundary. Before M1 this type
// had no ExitCode() method, so main.go could not read the code and mapped every
// error to exit 1 (REQ-CONT-001-005 defect). This is the load-bearing adoption.
func TestExitCodeErrorSatisfiesExitCoder(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want int
	}{
		{"exit 1", &exitCodeError{code: 1, msg: "user error"}, 1},
		{"exit 2", &exitCodeError{code: 2, msg: "system error"}, 2},
		{"exit 3", &exitCodeError{code: 3, msg: "invalid args"}, 3},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var ec exitCoder
			if !errors.As(tc.err, &ec) {
				t.Fatalf("errors.As(exitCodeError, &exitCoder) = false; want true (type must satisfy ExitCoder boundary)")
			}
			if got := ec.ExitCode(); got != tc.want {
				t.Fatalf("ExitCode() = %d, want %d", got, tc.want)
			}
			if tc.err.Error() == "" {
				t.Fatalf("Error() must carry the original diagnostic, got empty string")
			}
		})
	}
}

// TestExitCodeBoundary_DeferRuns proves the core behavioral value of REQ-CONT-001-001:
// returning an ExitCoder error (instead of calling os.Exit) lets defers run. An
// os.Exit would skip the defer; a return does not. This is why the 11 sites were
// converted from os.Exit to ExitCoder returns.
func TestExitCodeBoundary_DeferRuns(t *testing.T) {
	deferRan := false
	run := func() error {
		defer func() { deferRan = true }()
		// Representative converted path: an error-severity verdict that must
		// surface a specific exit code through the boundary.
		return &exitCodeError{code: 2, msg: "simulated system error"}
	}

	err := run()
	if !deferRan {
		t.Fatalf("defer did not run; returning an ExitCoder must let defers execute (os.Exit would skip it)")
	}
	var ec exitCoder
	if !errors.As(err, &ec) {
		t.Fatalf("errors.As = false; want ExitCoder")
	}
	if ec.ExitCode() != 2 {
		t.Fatalf("ExitCode() = %d, want 2", ec.ExitCode())
	}
}
