package cli

import (
	"strings"
	"testing"
)

// t2089: `moai factory wait` had no --run flag, so with two active factory runs the
// run it needs (to read the card's --after hint when --on is omitted) was resolved
// from an empty id and refused as ambiguous. Every other factory verb takes --run.
func TestFactoryWaitTakesRunWhenSeveralRunsAreActive(t *testing.T) {
	root := t.TempDir()
	seedLiveRun(t, factoryRootOf(root), "runa01")
	seedLiveRun(t, factoryRootOf(root), "runb02")

	// control: without --run the two live runs are ambiguous (the incident shape)
	out, err := runFactoryWaitCmd(t, root, "t2089", "--waker", "lane-3", "--reason", "r")
	if err == nil || !strings.Contains(err.Error()+out, "run is unresolved") {
		t.Fatalf("control: want the unresolved-run refusal, got err=%v out=%q", err, out)
	}

	// subject: --run names the run, so the run resolves; the card has no record in
	// it, so the verb now fails on the card, not on the run
	out, err = runFactoryWaitCmd(t, root, "t2089", "--waker", "lane-3", "--reason", "r", "--run", "runa01")
	if err == nil {
		t.Fatalf("a card with no record and no --on should still be refused: %q", out)
	}
	msg := err.Error() + out
	if strings.Contains(msg, "unknown flag") {
		t.Fatalf("wait does not take --run: %s", msg)
	}
	if strings.Contains(msg, "run is unresolved") {
		t.Errorf("--run runa01 was not used to resolve the run: %s", msg)
	}
	if !strings.Contains(msg, "card record is unreadable") {
		t.Errorf("want the card-record refusal once the run resolves, got: %s", msg)
	}
}

// --run with --on changes nothing (the predecessor is given, no run is read), so a
// lane can pass --run on every call.
func TestFactoryWaitAcceptsRunAlongsideOn(t *testing.T) {
	root := t.TempDir()
	seedLiveRun(t, factoryRootOf(root), "runa01")
	seedLiveRun(t, factoryRootOf(root), "runb02")
	out, err := runFactoryWaitCmd(t, root, "t2089", "--on", "t2010", "--waker", "lane-3", "--reason", "r", "--run", "runa01")
	if err != nil {
		t.Fatalf("wait with --on and --run: %v\n%s", err, out)
	}
	if !strings.Contains(out, "wait recorded: card=t2089 condition=predecessor-landed t2010") {
		t.Errorf("output = %q", out)
	}
}
