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
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// factoryRenewWire is the JSON that the renew verb (--json) and the factory_renew tool
// return: the refusal class in result, and the lease expiry in expires_at.
type factoryRenewWire struct {
	Result    string `json:"result"`
	ExpiresAt string `json:"expires_at"`
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
