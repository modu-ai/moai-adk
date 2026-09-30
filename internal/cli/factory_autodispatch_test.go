// factory_autodispatch_test.go — SPEC-TODO-CLASSIFY-DISPATCH-001 M4
// acceptance tests (AC-TCD-012): the -f lane's auto-dispatch is DEFAULT-ON,
// the launcher stamps the selection into the lane bootstrap carrier, and
// --no-auto-dispatch selects the manual mode. The default is recorded in
// code — no new config key (REQ-TCD-011).
package cli

import (
	"os"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestParseFactoryEntryAutoDispatchDefaultOn — a plain `-f lane` carries no
// opt-out: the selection the launcher stamps is the code default (auto).
func TestParseFactoryEntryAutoDispatchDefaultOn(t *testing.T) {
	sdScrubLauncherEnv(t)
	entry, err := parseLauncherEntry([]string{"-f", "lane"})
	if err != nil {
		t.Fatalf("parse -f lane: %v", err)
	}
	if !entry.FactoryEnabled || entry.AutoDispatchManual {
		t.Errorf("-f lane: enabled=%v autoDispatchManual=%v, want (true, false) — auto-dispatch is the default", entry.FactoryEnabled, entry.AutoDispatchManual)
	}
}

// TestParseFactoryEntryNoAutoDispatchOptOut — `--no-auto-dispatch` selects
// the manual mode on both lane shapes.
func TestParseFactoryEntryNoAutoDispatchOptOut(t *testing.T) {
	sdScrubLauncherEnv(t)
	for _, args := range [][]string{
		{"-f", "lane", "--no-auto-dispatch"},
		{"-f", "lane-2", "--no-auto-dispatch"},
	} {
		entry, err := parseLauncherEntry(args)
		if err != nil {
			t.Fatalf("parse %v: %v", args, err)
		}
		if !entry.AutoDispatchManual {
			t.Errorf("%v: autoDispatchManual=false, want the opt-out carried", args)
		}
	}
}

// TestParseFactoryEntryNoAutoDispatchLaneOnly — the opt-out is a LANE
// selection (the --clear-policy precedent): on the leader shape or without
// -f it is refused, never silently swallowed.
func TestParseFactoryEntryNoAutoDispatchLaneOnly(t *testing.T) {
	sdScrubLauncherEnv(t)
	for _, args := range [][]string{
		{"-f", "--no-auto-dispatch"},           // leader shape
		{"--no-auto-dispatch"},                 // no factory entry at all
		{"-f", "lane", "--no-auto-dispatch=1"}, // no value form exists
	} {
		if _, err := parseLauncherEntry(args); err == nil {
			t.Errorf("%v: parsed without refusal, want a lane-only usage error", args)
		}
	}
}

// TestEnterFactoryLaneModeStampsAutoDispatch — the launcher stamps the
// selection ALWAYS (the clear-policy overwrite discipline): a manual stamp
// in the child's environment, an auto stamp on the default path, and no
// value inherited from an outer session leaking through.
func TestEnterFactoryLaneModeStampsAutoDispatch(t *testing.T) {
	sdScrubLauncherEnv(t)
	t.Setenv(config.EnvFactoryAutoDispatch, config.FactoryDispatchManual) // poisoned outer value
	restore := enterFactoryLaneMode("lane-1", 1, "", config.FactoryDispatchAuto)
	if got := getenvAutoDispatch(); got != config.FactoryDispatchAuto {
		t.Errorf("auto path stamped %q, want %q (the stamp overwrites, nothing leaks)", got, config.FactoryDispatchAuto)
	}
	restore()
	if got := getenvAutoDispatch(); got != config.FactoryDispatchManual {
		t.Errorf("restore left %q, want the prior outer value %q", got, config.FactoryDispatchManual)
	}
	restore = enterFactoryLaneMode("lane-1", 1, "", config.FactoryDispatchManual)
	if got := getenvAutoDispatch(); got != config.FactoryDispatchManual {
		t.Errorf("opt-out path stamped %q, want %q", got, config.FactoryDispatchManual)
	}
	restore()
}

func getenvAutoDispatch() string { return os.Getenv(config.EnvFactoryAutoDispatch) }
