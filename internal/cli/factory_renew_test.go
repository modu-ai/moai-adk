// factory_renew_test.go — SPEC-FACTORY-LEASE-RENEW-001 M2 acceptance tests (card t1663):
// the lane renew verb (REQ-FLR-001), the factory_renew MCP tool (REQ-FLR-002), the
// non-holder refusal (REQ-FLR-003), and the verb half of the machine-readable refusal
// class (REQ-FLR-011). Each test reaches its surface through the command tree or through
// the MCP registration, so the file names no Go symbol that the milestone adds: a missing
// surface fails on an assertion, never at compile time.
package cli

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// factoryRenewWire is the JSON that the renew verb (--json) and the factory_renew tool
// return: the refusal class in result, the lease expiry in expires_at, and the refusal
// text in detail (absent on a renewal).
type factoryRenewWire struct {
	Result    string `json:"result"`
	ExpiresAt string `json:"expires_at"`
	Detail    string `json:"detail,omitempty"`
}

// factoryRenewPlace places card t1 in the run state, leased to holder until expiry.
func factoryRenewPlace(t *testing.T, root, holder string, expiry time.Time) {
	t.Helper()
	fcPlace(t, root, homestate.Card{
		CardID: "t1", State: homestate.CardRun, Stage: homestate.CardRun,
		OwnerLabel: holder, LeaseHolder: holder,
		LeaseExpiresAt: expiry.Format(time.RFC3339Nano), WorktreePath: root,
	})
}

// factoryRenewRecord reads card t1 and the number of event rows in the factory record.
func factoryRenewRecord(t *testing.T, root string) (homestate.Card, int) {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	defer func() { _ = db.Close() }()
	card, err := db.LoadCard(context.Background(), fcRun, "t1")
	if err != nil {
		t.Fatalf("load t1: %v", err)
	}
	var events int
	if err := db.DB.QueryRow(`SELECT count(*) FROM events`).Scan(&events); err != nil {
		t.Fatal(err)
	}
	return card, events
}

// factoryRenewMCPHandler returns the handler that registerFactoryCardMCPTools binds to
// name. A name the registration does not bind fails the test on this assertion.
func factoryRenewMCPHandler(t *testing.T, name string) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	t.Helper()
	bound := map[string]server.ToolHandlerFunc{}
	registerFactoryCardMCPTools(func(tool string, _ mcp.Tool, handler server.ToolHandlerFunc) {
		bound[tool] = handler
	})
	handler, ok := bound[name]
	if !ok {
		t.Fatalf("MCP tool %s is not bound by registerFactoryCardMCPTools", name)
	}
	return handler
}

// factoryRenewViaMCP calls factory_renew for card t1 through the registered handler and
// decodes the JSON outcome of its text result.
func factoryRenewViaMCP(t *testing.T, root string) factoryRenewWire {
	t.Helper()
	text, err := sdCallTool(t, factoryRenewMCPHandler(t, "factory_renew"),
		map[string]any{projectRootArg: root, "card": "t1", "run": fcRun})
	if err != nil {
		t.Fatalf("factory_renew: %v", err)
	}
	var wire factoryRenewWire
	if jerr := json.Unmarshal([]byte(strings.TrimSpace(text)), &wire); jerr != nil {
		t.Fatalf("factory_renew text %q is not the JSON outcome: %v", text, jerr)
	}
	return wire
}

// factoryRenewViaVerb runs `moai factory renew t1 --json`. It returns the decoded JSON
// outcome and the command's error: a refused renewal exits non-zero and still prints the
// outcome on stdout.
func factoryRenewViaVerb(t *testing.T) (factoryRenewWire, error) {
	t.Helper()
	out, _, err := runFactory(t, "renew", "t1", "--json", "--run", fcRun)
	var wire factoryRenewWire
	if jerr := json.Unmarshal([]byte(strings.TrimSpace(out)), &wire); jerr != nil {
		t.Fatalf("renew --json printed no JSON outcome (stdout %q, error %v): %v", out, err, jerr)
	}
	return wire, err
}

// AC-FLR-001 (REQ-FLR-001): the holder of a live lease runs the renew verb. The lease
// extends to the request time plus the configured card lease duration (15 minutes at the
// default), the verb prints the new expiry, and the card's state and version are unchanged.
func TestFactoryRenewExtendsLiveLease(t *testing.T) {
	root, _ := sdMoaiFixture(t)
	sdRegisterLane(t, root, "lane-1")
	factoryRenewPlace(t, root, "lane-1", fcNow.Add(5*time.Minute))
	before, _ := factoryRenewRecord(t, root)
	sdLaneEnv(t, "lane-1", "")
	want := fcNow.Add(15 * time.Minute).Format(time.RFC3339Nano)

	out, _, err := runFactory(t, "renew", "t1", "--run", fcRun)
	if err != nil {
		t.Fatalf("renew: %v", err)
	}
	if !strings.Contains(out, want) {
		t.Errorf("renew output %q does not print the new expiry %s", out, want)
	}
	after, _ := factoryRenewRecord(t, root)
	if after.LeaseExpiresAt != want {
		t.Errorf("lease expiry = %s, want %s (request time plus 15m)", after.LeaseExpiresAt, want)
	}
	if after.Version != before.Version || after.State != before.State {
		t.Errorf("renewal changed the card: v%d %s → v%d %s", before.Version, before.State, after.Version, after.State)
	}
}

// AC-FLR-002 (REQ-FLR-002): the factory_renew tool returns the same expiry fields and the
// same refusal class as the renew verb, for the same card, caller, and lease state.
func TestFactoryRenewMatchesMCP(t *testing.T) {
	t.Run("live lease: the same expiry fields", func(t *testing.T) {
		root, _ := sdMoaiFixture(t)
		sdRegisterLane(t, root, "lane-1")
		factoryRenewPlace(t, root, "lane-1", fcNow.Add(5*time.Minute))
		sdLaneEnv(t, "lane-1", "")
		viaMCP := factoryRenewViaMCP(t, root)
		viaVerb, verbErr := factoryRenewViaVerb(t)
		if verbErr != nil {
			t.Fatalf("renew --json on a live lease: %v", verbErr)
		}
		if viaMCP.Result != "renewed" || viaVerb.Result != "renewed" {
			t.Fatalf("result: mcp = %q, verb = %q, want renewed from both", viaMCP.Result, viaVerb.Result)
		}
		if viaMCP.ExpiresAt != viaVerb.ExpiresAt {
			t.Errorf("expires_at: mcp = %q, verb = %q, want equal", viaMCP.ExpiresAt, viaVerb.ExpiresAt)
		}
	})

	t.Run("expired lease: the same refusal class", func(t *testing.T) {
		root, _ := sdMoaiFixture(t)
		sdRegisterLane(t, root, "lane-1")
		factoryRenewPlace(t, root, "lane-1", fcNow.Add(-2*time.Minute))
		sdLaneEnv(t, "lane-1", "")
		viaMCP := factoryRenewViaMCP(t, root)
		viaVerb, verbErr := factoryRenewViaVerb(t)
		if verbErr == nil {
			t.Errorf("renew --json on an expired lease: err = nil, want a non-zero exit")
		}
		if viaMCP.Result != "refused-expired" || viaVerb.Result != "refused-expired" {
			t.Fatalf("result: mcp = %q, verb = %q, want refused-expired from both", viaMCP.Result, viaVerb.Result)
		}
		if viaMCP.ExpiresAt != viaVerb.ExpiresAt {
			t.Errorf("expires_at: mcp = %q, verb = %q, want equal", viaMCP.ExpiresAt, viaVerb.ExpiresAt)
		}
	})
}

// AC-FLR-003 (REQ-FLR-003): a lane that does not hold a live lease cannot renew it,
// through the verb or through the MCP tool. Both refuse with the holder-refusal class, and
// the card's state, version, holder, expiry, and event count are unchanged.
func TestNonHolderRenewRefusedCLIAndMCP(t *testing.T) {
	root, _ := sdMoaiFixture(t)
	sdRegisterLane(t, root, "lane-1")
	sdRegisterLane(t, root, "lane-2")
	factoryRenewPlace(t, root, "lane-1", fcNow.Add(5*time.Minute))
	before, eventsBefore := factoryRenewRecord(t, root)
	sdLaneEnv(t, "lane-2", "")

	// Each surface is its own subtest, so a red on one does not hide the other.
	t.Run("verb", func(t *testing.T) {
		wire, verbErr := factoryRenewViaVerb(t)
		if verbErr == nil {
			t.Errorf("renew --json by a non-holder: err = nil, want a non-zero exit")
		}
		if wire.Result != "refused-not-holder" {
			t.Errorf("renew --json by a non-holder: result = %q, want refused-not-holder", wire.Result)
		}
	})
	t.Run("MCP tool", func(t *testing.T) {
		if wire := factoryRenewViaMCP(t, root); wire.Result != "refused-not-holder" {
			t.Errorf("factory_renew by a non-holder: result = %q, want refused-not-holder", wire.Result)
		}
	})

	after, eventsAfter := factoryRenewRecord(t, root)
	if after.Version != before.Version || after.State != before.State ||
		after.LeaseHolder != before.LeaseHolder || after.LeaseExpiresAt != before.LeaseExpiresAt {
		t.Errorf("refused renewal changed the card: v%d %s holder=%q expiry=%s → v%d %s holder=%q expiry=%s",
			before.Version, before.State, before.LeaseHolder, before.LeaseExpiresAt,
			after.Version, after.State, after.LeaseHolder, after.LeaseExpiresAt)
	}
	if eventsAfter != eventsBefore {
		t.Errorf("refused renewals appended %d event rows, want none", eventsAfter-eventsBefore)
	}
}

// AC-FLR-011, verb half (REQ-FLR-011): the renew verb's JSON output carries the field
// result with the value refused-expired for a lease that expired two minutes ago, and the
// refusal writes nothing.
func TestFactoryRenewRefusalIsMachineReadable(t *testing.T) {
	root, _ := sdMoaiFixture(t)
	sdRegisterLane(t, root, "lane-1")
	expiry := fcNow.Add(-2 * time.Minute)
	factoryRenewPlace(t, root, "lane-1", expiry)
	before, eventsBefore := factoryRenewRecord(t, root)
	sdLaneEnv(t, "lane-1", "")

	wire, verbErr := factoryRenewViaVerb(t)
	if verbErr == nil {
		t.Errorf("renew --json on an expired lease: err = nil, want a non-zero exit")
	}
	if wire.Result != "refused-expired" {
		t.Fatalf("renew --json result = %q, want refused-expired", wire.Result)
	}
	if want := expiry.Format(time.RFC3339Nano); wire.ExpiresAt != want {
		t.Errorf("renew --json expires_at = %q, want the lapsed expiry %q", wire.ExpiresAt, want)
	}
	after, eventsAfter := factoryRenewRecord(t, root)
	if after.Version != before.Version || after.State != before.State || after.LeaseExpiresAt != before.LeaseExpiresAt {
		t.Errorf("refused renewal changed the card: v%d %s expiry=%s → v%d %s expiry=%s",
			before.Version, before.State, before.LeaseExpiresAt, after.Version, after.State, after.LeaseExpiresAt)
	}
	if eventsAfter != eventsBefore {
		t.Errorf("refused renewal appended %d event rows, want none", eventsAfter-eventsBefore)
	}
}

// factoryRenewAssertUnchanged asserts that a refused renewal left card t1 and the event
// count as they were before the call: no version bump, no holder or expiry change, no event.
func factoryRenewAssertUnchanged(t *testing.T, root string, before homestate.Card, eventsBefore int) {
	t.Helper()
	after, eventsAfter := factoryRenewRecord(t, root)
	if after.Version != before.Version || after.State != before.State ||
		after.LeaseHolder != before.LeaseHolder || after.LeaseExpiresAt != before.LeaseExpiresAt {
		t.Errorf("refused renewal changed the card: v%d %s holder=%q expiry=%s → v%d %s holder=%q expiry=%s",
			before.Version, before.State, before.LeaseHolder, before.LeaseExpiresAt,
			after.Version, after.State, after.LeaseHolder, after.LeaseExpiresAt)
	}
	if eventsAfter != eventsBefore {
		t.Errorf("refused renewal appended %d event rows, want none", eventsAfter-eventsBefore)
	}
}

// factoryRenewCardIn reads card cardID of run runID.
func factoryRenewCardIn(t *testing.T, root, runID, cardID string) homestate.Card {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	defer func() { _ = db.Close() }()
	card, err := db.LoadCard(context.Background(), runID, cardID)
	if err != nil {
		t.Fatalf("load %s in %s: %v", cardID, runID, err)
	}
	return card
}

// factoryRenewWriteLeaseDuration writes the workflow.yaml of root with the card lease
// duration. The renew interval keeps its compiled default.
func factoryRenewWriteLeaseDuration(t *testing.T, root, duration string) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "workflow:\n  factory_lease:\n    duration: " + duration + "\n"
	if err := os.WriteFile(filepath.Join(dir, "workflow.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// AC-FLR-002 (REQ-FLR-002), legacy-label parity — sync-audit F1, card t1663. A lane whose
// MOAI_FACTORY_WORKER carries a legacy spelling is refused by the renew verb before any work
// (factoryLaneLabelFromEnv). The factory_renew tool must refuse it the same way: a tool error
// that names the legacy label, and a lease left exactly as it was.
func TestFactoryRenewRefusesLegacyLaneLabelOnBothSurfaces(t *testing.T) {
	const legacy = "worker-1"
	fresh := func(t *testing.T) string {
		t.Helper()
		root, _ := sdMoaiFixture(t)
		sdRegisterLane(t, root, legacy)
		factoryRenewPlace(t, root, legacy, fcNow.Add(5*time.Minute))
		sdLaneEnv(t, legacy, "")
		return root
	}

	t.Run("MCP tool", func(t *testing.T) {
		root := fresh(t)
		before, eventsBefore := factoryRenewRecord(t, root)
		text, err := sdCallTool(t, factoryRenewMCPHandler(t, "factory_renew"),
			map[string]any{projectRootArg: root, "card": "t1", "run": fcRun})
		switch {
		case err == nil:
			t.Errorf("factory_renew under the legacy label %q was not refused; it returned %q", legacy, text)
		case !strings.Contains(err.Error(), "legacy lane label"):
			t.Errorf("factory_renew refusal does not name the legacy label: %v", err)
		}
		factoryRenewAssertUnchanged(t, root, before, eventsBefore)
	})

	t.Run("verb", func(t *testing.T) {
		root := fresh(t)
		before, eventsBefore := factoryRenewRecord(t, root)
		_, _, err := runFactory(t, "renew", "t1", "--json", "--run", fcRun)
		switch {
		case err == nil:
			t.Errorf("renew --json under the legacy label %q was not refused", legacy)
		case !strings.Contains(err.Error(), "legacy lane label"):
			t.Errorf("renew refusal does not name the legacy label: %v", err)
		}
		factoryRenewAssertUnchanged(t, root, before, eventsBefore)
	})
}

// AC-FLR-002 (REQ-FLR-002), linked-worktree parity — sync-audit F2, card t1663. The factory
// record is kept at the queue root (the primary checkout), but the renewal reads the lease
// keys from the root it is handed. The verb hands it the queue root of its launch tree; the
// tool must hand it the same root for a caller in a linked worktree. Otherwise the two
// surfaces renew one card at two different durations.
func TestFactoryRenewMatchesMCPLinkedWorktree(t *testing.T) {
	for _, tc := range []struct {
		name     string
		duration string
	}{
		{"primary duration 20m", "20m"},
		{"primary duration 0s", "0s"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			primary, _ := sdMoaiFixture(t)
			wt := addGitWorktree(t, primary)
			sdRegisterLane(t, primary, "lane-1")
			factoryRenewPlace(t, primary, "lane-1", fcNow.Add(5*time.Minute))
			// The primary's keys are set after the fixture opened the factory: a 0s duration
			// refuses the open, and the fixture itself must not be the first thing to see it.
			factoryRenewWriteLeaseDuration(t, primary, tc.duration)
			sdLaneEnv(t, "lane-1", "")
			t.Setenv("CLAUDE_PROJECT_DIR", wt)

			viaVerb, verbErr := factoryRenewViaVerb(t)
			text, err := sdCallTool(t, factoryRenewMCPHandler(t, "factory_renew"),
				map[string]any{projectRootArg: wt, "card": "t1", "run": fcRun})
			if err != nil {
				t.Fatalf("factory_renew from the linked worktree: tool error %v", err)
			}
			var viaMCP factoryRenewWire
			if jerr := json.Unmarshal([]byte(strings.TrimSpace(text)), &viaMCP); jerr != nil {
				t.Fatalf("factory_renew text %q is not the JSON outcome: %v", text, jerr)
			}
			if viaMCP != viaVerb {
				t.Errorf("linked worktree: factory_renew = %+v, renew verb = %+v (verb error %v); want equal",
					viaMCP, viaVerb, verbErr)
			}
			// Anchor: the verb reads the primary's keys, so the shared outcome must be the
			// primary's configured one, not the compiled default that a worktree would give.
			switch tc.duration {
			case "20m":
				want := fcNow.Add(20 * time.Minute).Format(time.RFC3339Nano)
				if viaVerb.Result != "renewed" || viaVerb.ExpiresAt != want {
					t.Errorf("verb outcome = %+v, want renewed to %s (the primary's 20m)", viaVerb, want)
				}
			case "0s":
				if viaVerb.Result != "error" {
					t.Errorf("verb outcome = %+v, want result error (the primary's 0s is refused)", viaVerb)
				}
			}
		})
	}
}

// AC-FLR-010 and AC-FLR-011, renewal without --run — sync-audit F3, card t1663. The watchdog's
// renew step runs `moai factory renew <card> --json` with no --run. With two active factory
// runs the single-run resolution refuses, so the renewal must resolve the lane's own run for
// the card from the lane identity: the one run in which this lane holds the card.
func TestFactoryRenewResolvesLaneRunWithoutRunArg(t *testing.T) {
	// fixture seeds two owner-live runs. Card t1 exists in both, held by holderInA in run-a and
	// by holderInB in run-b. The lane under test is lane-1.
	fixture := func(t *testing.T, holderInA, holderInB, bindTo string) string {
		t.Helper()
		root, _ := sdMoaiFixture(t)
		seedLiveRun(t, root, "run-a")
		seedLiveRun(t, root, "run-b")
		sdRegisterLane(t, root, "lane-1")
		sdRegisterLane(t, root, "lane-2")
		if bindTo != "" {
			factoryRenewBindLane(t, root, "lane-1", bindTo)
		}
		live := fcNow.Add(5 * time.Minute).Format(time.RFC3339Nano)
		fcPlace(t, root,
			homestate.Card{RunID: "run-a", CardID: "t1", State: homestate.CardRun, Stage: homestate.CardRun,
				OwnerLabel: holderInA, LeaseHolder: holderInA, LeaseExpiresAt: live, WorktreePath: root},
			homestate.Card{RunID: "run-b", CardID: "t1", State: homestate.CardRun, Stage: homestate.CardRun,
				OwnerLabel: holderInB, LeaseHolder: holderInB, LeaseExpiresAt: live, WorktreePath: root},
		)
		sdLaneEnv(t, "lane-1", "")
		return root
	}
	want := fcNow.Add(15 * time.Minute).Format(time.RFC3339Nano)

	t.Run("verb", func(t *testing.T) {
		root := fixture(t, "lane-2", "lane-1", "run-b")
		before := factoryRenewCardIn(t, root, "run-a", "t1")
		out, _, err := runFactory(t, "renew", "t1", "--json")
		if err != nil {
			t.Fatalf("renew t1 --json with no --run and two active runs: %v (stdout %q)", err, out)
		}
		var wire factoryRenewWire
		if jerr := json.Unmarshal([]byte(strings.TrimSpace(out)), &wire); jerr != nil {
			t.Fatalf("renew --json printed no JSON outcome (stdout %q): %v", out, jerr)
		}
		if wire.Result != "renewed" || wire.ExpiresAt != want {
			t.Errorf("renew outcome = %+v, want renewed to %s", wire, want)
		}
		if got := factoryRenewCardIn(t, root, "run-b", "t1").LeaseExpiresAt; got != want {
			t.Errorf("run-b lease expiry = %s, want %s", got, want)
		}
		if after := factoryRenewCardIn(t, root, "run-a", "t1"); after.LeaseExpiresAt != before.LeaseExpiresAt {
			t.Errorf("run-a lease changed %s → %s; the renewal must touch only the lane's run",
				before.LeaseExpiresAt, after.LeaseExpiresAt)
		}
	})

	t.Run("MCP tool", func(t *testing.T) {
		root := fixture(t, "lane-2", "lane-1", "run-b")
		text, err := sdCallTool(t, factoryRenewMCPHandler(t, "factory_renew"),
			map[string]any{projectRootArg: root, "card": "t1"})
		if err != nil {
			t.Fatalf("factory_renew with no run and two active runs: %v", err)
		}
		var wire factoryRenewWire
		if jerr := json.Unmarshal([]byte(strings.TrimSpace(text)), &wire); jerr != nil {
			t.Fatalf("factory_renew text %q is not the JSON outcome: %v", text, jerr)
		}
		if wire.Result != "renewed" || wire.ExpiresAt != want {
			t.Errorf("factory_renew outcome = %+v, want renewed to %s", wire, want)
		}
	})

	// Guard arms. The renewal stays refused, and the refusal names --run, in two states that the
	// roster cannot resolve to a renewal: the lane holds the card in both active runs and has no
	// roster binding (nothing says which run is the lane's), or the roster binds the lane to run-b
	// and run-b does not hold the card for the lane. The pre-fix code refused both states too, so
	// these pin that the narrowing never turns a refusal into a renewal without a binding.
	for _, tc := range []struct{ name, holderInA, holderInB, bindTo string }{
		{"held in both runs with no roster binding refuses and names --run", "lane-1", "lane-1", ""},
		{"held in neither run, bound to run-b, refuses and names --run", "lane-2", "lane-2", "run-b"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixture(t, tc.holderInA, tc.holderInB, tc.bindTo)
			_, _, err := runFactory(t, "renew", "t1", "--json")
			if err == nil || !strings.Contains(err.Error(), "--run") {
				t.Errorf("renew with no --run (holders %s, %s; bound to %q): err = %v, want a refusal naming --run",
					tc.holderInA, tc.holderInB, tc.bindTo, err)
			}
		})
	}
}

// factoryRenewBindLane binds a registered lane label to a factory run in the roster: it sets the
// run_id of the label's workers row, the column the claim path writes at join
// (internal/factory/factory_slots.go:355; the column is internal/homestate/factory.go:29).
func factoryRenewBindLane(t *testing.T, root, label, runID string) {
	t.Helper()
	db := fcOpen(t, root)
	if _, err := db.DB.Exec(`UPDATE workers SET run_id=? WHERE label=?`, runID, label); err != nil {
		t.Fatalf("bind %s to %s in the roster: %v", label, runID, err)
	}
}

// factoryRenewNoRun renews card t1 with no --run through the surface named by surface ("verb" or
// "MCP tool") and decodes the JSON outcome. The error is the verb's exit error or the tool error.
func factoryRenewNoRun(t *testing.T, root, surface string) (factoryRenewWire, error) {
	t.Helper()
	var wire factoryRenewWire
	if surface == "verb" {
		out, _, err := runFactory(t, "renew", "t1", "--json")
		if jerr := json.Unmarshal([]byte(strings.TrimSpace(out)), &wire); jerr != nil {
			t.Fatalf("renew --json printed no JSON outcome (stdout %q, error %v): %v", out, err, jerr)
		}
		return wire, err
	}
	text, err := sdCallTool(t, factoryRenewMCPHandler(t, "factory_renew"), map[string]any{projectRootArg: root, "card": "t1"})
	if err != nil {
		return wire, err
	}
	if jerr := json.Unmarshal([]byte(strings.TrimSpace(text)), &wire); jerr != nil {
		t.Fatalf("factory_renew text %q is not the JSON outcome: %v", text, jerr)
	}
	return wire, nil
}

// TestFactoryRenewRefusesLabelReuseAcrossRuns is sync-audit F7 (card t1663; codex P1). A lane label can
// be reused across factory runs. The roster (workers.run_id, one row per label) binds the label to the run
// the lane joined, while a card lease in another active run still names the label as holder. A renewal
// with no --run must act only in the run the roster binds the lane to.
//
// The first arm is the defect shape: the roster binds lane-1 to run-b, and the card is held by lane-1 in
// run-a. The renewal must refuse, name --run, and leave run-a's lease unchanged. The second arm is the
// positive control: the roster binds lane-1 to run-a, where the card is held, and the renewal renews run-a.
// The third arm holds the card by lane-1 in both runs with the roster binding run-b: the bound run is
// renewed, and the stale lease in run-a is left alone.
func TestFactoryRenewRefusesLabelReuseAcrossRuns(t *testing.T) {
	// fixture: run-a and run-b are both live, and lane-1 and lane-2 are registered. The roster binds
	// lane-1 to bound. Card t1 is held by lane-1 in run-a and by holderInB in run-b.
	fixture := func(t *testing.T, bound, holderInB string) string {
		t.Helper()
		root, _ := sdMoaiFixture(t)
		seedLiveRun(t, root, "run-a")
		seedLiveRun(t, root, "run-b")
		sdRegisterLane(t, root, "lane-1")
		sdRegisterLane(t, root, "lane-2")
		factoryRenewBindLane(t, root, "lane-1", bound)
		live := fcNow.Add(5 * time.Minute).Format(time.RFC3339Nano)
		fcPlace(t, root,
			homestate.Card{RunID: "run-a", CardID: "t1", State: homestate.CardRun, Stage: homestate.CardRun,
				OwnerLabel: "lane-1", LeaseHolder: "lane-1", LeaseExpiresAt: live, WorktreePath: root},
			homestate.Card{RunID: "run-b", CardID: "t1", State: homestate.CardRun, Stage: homestate.CardRun,
				OwnerLabel: holderInB, LeaseHolder: holderInB, LeaseExpiresAt: live, WorktreePath: root},
		)
		sdLaneEnv(t, "lane-1", "")
		return root
	}
	want := fcNow.Add(15 * time.Minute).Format(time.RFC3339Nano)
	arms := []struct{ name, bound, holderInB, renews string }{
		{"reused label refuses and names --run", "run-b", "lane-2", ""},
		{"positive control renews the bound run", "run-a", "lane-2", "run-a"},
		{"held in both runs renews the bound run", "run-b", "lane-1", "run-b"},
	}
	for _, arm := range arms {
		for _, surface := range []string{"verb", "MCP tool"} {
			t.Run(arm.name+"/"+surface, func(t *testing.T) {
				root := fixture(t, arm.bound, arm.holderInB)
				beforeA := factoryRenewCardIn(t, root, "run-a", "t1")
				beforeB := factoryRenewCardIn(t, root, "run-b", "t1")
				wire, callErr := factoryRenewNoRun(t, root, surface)
				afterA := factoryRenewCardIn(t, root, "run-a", "t1")
				afterB := factoryRenewCardIn(t, root, "run-b", "t1")
				if arm.renews == "" {
					if wire.Result == "renewed" {
						t.Errorf("renew with no --run renewed (expires %s) although the roster binds lane-1 to %s",
							wire.ExpiresAt, arm.bound)
					}
					if wire.Result != "error" || !strings.Contains(wire.Detail, "--run") {
						t.Errorf("refusal does not name --run: result=%s detail=%q err=%v", wire.Result, wire.Detail, callErr)
					}
				} else if callErr != nil || wire.Result != "renewed" || wire.ExpiresAt != want {
					t.Errorf("renew with no --run: err=%v result=%s expires=%s, want renewed to %s in %s",
						callErr, wire.Result, wire.ExpiresAt, want, arm.renews)
				}
				for _, run := range []string{"run-a", "run-b"} {
					before, after := beforeA, afterA
					if run == "run-b" {
						before, after = beforeB, afterB
					}
					wantExpiry := before.LeaseExpiresAt
					if run == arm.renews {
						wantExpiry = want
					}
					if after.LeaseExpiresAt != wantExpiry {
						t.Errorf("%s lease expiry = %s, want %s", run, after.LeaseExpiresAt, wantExpiry)
					}
				}
			})
		}
	}
}
