package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// The landed ref used to be resolved while cobra built the command tree, which
// made EVERY `moai` invocation — `moai statusline`, once per render — pay a
// `git rev-parse` for help text it was never going to print. The resolution is
// now deferred to the help and usage paths. These tests pin both halves: that
// nothing ref-dependent is materialized at construction, and that help still
// prints the resolved ref, which is the regression deferral could cause.
//
// `--help` returns before RunE, so a PreRun hook would have printed help with
// the ref silently missing; that is why the hook is the help function.

// TestLandedRefNotMaterializedAtConstruction asserts the ref-dependent parts of
// the help surface are absent until something renders help. It is the
// unit-level shadow of the end-to-end property (a non-todo command spawns no
// git on this account), which is measured with a PATH shim rather than here.
func TestLandedRefNotMaterializedAtConstruction(t *testing.T) {
	pr := newTodoPRCmd()
	if pr.Long != "" {
		t.Errorf("todo pr Long is populated at construction:\n%s", pr.Long)
	}

	done := newTodoDoneCmd()
	if done.Long != "" {
		t.Errorf("todo done Long is populated at construction:\n%s", done.Long)
	}
	if usage := done.Flags().Lookup("require-landed").Usage; usage != "" {
		t.Errorf("--require-landed usage is populated at construction: %q", usage)
	}
}

// helpOutput renders `moai todo <sub> --help` through the real command tree and
// returns what a user would see.
func helpOutput(t *testing.T, sub string) string {
	t.Helper()
	cmd := newTodoCmd()
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs([]string{sub, "--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("todo %s --help: %v", sub, err)
	}
	return out.String() + errBuf.String()
}

// TestHelpNamesTheResolvedLandedRef is the guard on the deferral. It compares
// the help text against the ref the code itself resolves rather than against a
// hardcoded name, because that equality — help, flag description, refusal and
// query can never name different refs — is the property the resolution exists
// to hold.
func TestHelpNamesTheResolvedLandedRef(t *testing.T) {
	ref := todoLandedRefOnce()
	if ref == "" {
		t.Fatal("todoLandedRefOnce resolved to the empty string")
	}

	prHelp := helpOutput(t, "pr")
	if !strings.Contains(prHelp, ref) {
		t.Errorf("todo pr --help does not name the resolved ref %q:\n%s", ref, prHelp)
	}
	if !strings.Contains(prHelp, "whether its work has already landed on") {
		t.Errorf("todo pr --help lost its Long body:\n%s", prHelp)
	}

	doneHelp := helpOutput(t, "done")
	if !strings.Contains(doneHelp, ref) {
		t.Errorf("todo done --help does not name the resolved ref %q:\n%s", ref, doneHelp)
	}
	if !strings.Contains(doneHelp, "Move the addressed card out of the live queue") {
		t.Errorf("todo done --help lost its Long body:\n%s", doneHelp)
	}
	if !strings.Contains(doneHelp, "names the card (opt-in; see --help for its limit)") {
		t.Errorf("todo done --help lost the --require-landed flag description:\n%s", doneHelp)
	}
}

// TestUsageNamesTheResolvedLandedRef covers the second render path: an argument
// error prints flag usage WITHOUT printing help, so hooking only the help
// function would have left that surface naming no ref at all.
func TestUsageNamesTheResolvedLandedRef(t *testing.T) {
	ref := todoLandedRefOnce()
	// strings.Contains(anything, "") is true, so without this the assertion
	// below passes for every possible usage string. Measured, not supposed:
	// forcing todoLandedRefOnce to return "" left this test green while its
	// sibling above caught the same mutant.
	if ref == "" {
		t.Fatal("todoLandedRefOnce resolved to the empty string")
	}
	done := newTodoDoneCmd()
	// Attach a parent so the usage function has one to delegate to, exactly as
	// the real tree does.
	parent := &cobra.Command{Use: "todo"}
	parent.AddCommand(done)

	usage := done.UsageString()
	if !strings.Contains(usage, ref) {
		t.Errorf("todo done usage does not name the resolved ref %q:\n%s", ref, usage)
	}
}

// TestLandedRefDeferralHasExactlyOneConsumerSurface guards the coupling this
// deferral CREATED.
//
// Deferring `.Long` until help renders is safe only while the help path is the
// only thing that reads it. Cobra's man-page generator reads `.Long` directly,
// with no hook to run first, so if fang's man surface were enabled the man
// pages for `todo pr` and `todo done` would ship with empty descriptions —
// silently, because nothing else would change.
//
// TestLandedRefNotMaterializedAtConstruction documents that emptiness; it
// cannot protect against this, because it asserts the very state that would be
// wrong. It would stay green while the man pages shipped blank. This test is
// the protection: the failure lands on whoever removes the option, and it names
// what breaks rather than merely what changed.
//
// It reads the source rather than the built command because fang.Option values
// are functions and cannot be compared. That is a real limit — a rename that
// preserved the behaviour would trip it, and an equivalent option spelled
// differently would not. It sits here, beside the deferral it protects, so the
// reader who changes the deferral meets the constraint.
func TestLandedRefDeferralHasExactlyOneConsumerSurface(t *testing.T) {
	src, err := os.ReadFile("fang.go")
	if err != nil {
		t.Fatalf("read fang.go: %v", err)
	}
	if !strings.Contains(string(src), "fang.WithoutManpage()") {
		t.Error("fang.WithoutManpage() is gone from fang.go, which admits a SECOND reader of .Long.\n" +
			"`todo pr` and `todo done` defer their .Long until the help function runs (todo.go, withResolvedLandedRef),\n" +
			"so cobra's man generator — which reads .Long directly and runs no help function — would emit them EMPTY.\n" +
			"Enabling man pages therefore means populating .Long on that path too, not just deleting this option.")
	}
}
