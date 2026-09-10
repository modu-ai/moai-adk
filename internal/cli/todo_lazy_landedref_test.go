package cli

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"charm.land/fang/v2"
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

// manCmdRegistered runs a throwaway root command through the real fang.Execute
// with the given options and reports whether fang added its hidden `man`
// subcommand. It probes the BUILT command tree, which is the surface that
// decides whether .Long gets a second consumer.
func manCmdRegistered(t *testing.T, opts ...fang.Option) bool {
	t.Helper()
	root := &cobra.Command{Use: "probe", Run: func(*cobra.Command, []string) {}}
	root.SetArgs([]string{})
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	if err := fang.Execute(context.Background(), root, opts...); err != nil {
		t.Fatalf("fang.Execute on the probe root: %v", err)
	}
	for _, c := range root.Commands() {
		if c.Name() == "man" {
			return true
		}
	}
	return false
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
// the protection: the failure lands on whoever admits the second consumer, and
// it names what breaks rather than merely what changed.
//
// The probe is behavioural: it runs the real fangOptions() through the real
// fang.Execute on a throwaway root and asks whether fang registered its `man`
// command — the command whose RunE hands the tree to mango, which reads .Long.
// It binds the PROPERTY (the fang surface registers no second .Long consumer)
// rather than the spelling of an option literal, so any equivalent way of
// disabling man pages satisfies it and any way of enabling them trips it.
//
// The control assertion exists because the guard is a negative: if an upstream
// fang change stopped registering `man` under ANY configuration, the guard
// would pass forever while asserting nothing. The control fails loudly instead.
//
// The limit that REMAINS: this binds fang's man surface only. A second reader
// of .Long arriving by another route — a `cobra/doc` generator wired up
// elsewhere, or a hand-rolled man/markdown emitter — is not bound here.
func TestLandedRefDeferralHasExactlyOneConsumerSurface(t *testing.T) {
	// Control: with no options, fang MUST register `man`. Without this the
	// assertion below could pass while asserting nothing.
	if !manCmdRegistered(t) {
		t.Fatal("control failed: fang.Execute with no options did NOT register a `man` command.\n" +
			"This probe no longer detects fang's man surface, so the guard below has gone blind —\n" +
			"it would pass whether or not moai enables man pages. Re-derive the probe against the\n" +
			"current fang version before trusting this test again.")
	}

	// The guard: the real configuration must admit no man surface.
	if manCmdRegistered(t, fangOptions()...) {
		t.Error("fangOptions() now yields a fang `man` command, which admits a SECOND reader of .Long.\n" +
			"`todo pr` and `todo done` defer their .Long until the help function runs (todo.go, withResolvedLandedRef),\n" +
			"so cobra's man generator — which reads .Long directly and runs no help function — would emit them EMPTY.\n" +
			"Enabling man pages therefore means populating .Long on that path too, not just flipping this option.")
	}
}
