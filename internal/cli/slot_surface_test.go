package cli

// slot_surface_test.go — the rest of the `moai slot` surface (card t607, M3):
// exit codes that separate held from busy, status in every state, the list
// form, release, and input refusals. Written after the M3 implementation as
// surface coverage; the behavioural decisions they exercise were pinned RED
// in internal/kanban during M1.

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// seedSlotCLILease writes a record straight to the pinned path.
func seedSlotCLILease(t *testing.T, root string, lease kanban.SlotLease) {
	t.Helper()
	path := filepath.Join(root, ".moai", "state", "slot-leases", lease.Resource+".json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	data, err := json.Marshal(&lease)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func slotCLILease(resource, session string, pid int, expires time.Time) kanban.SlotLease {
	return kanban.SlotLease{
		Resource: resource, SessionID: session, SessionName: "lane-" + session, PID: pid,
		PIDSource: kanban.PIDSourceSessionOwner, Command: "heavy-suite",
		AcquiredAt:  time.Now().UTC().Add(-time.Minute).Format(time.RFC3339),
		MaxDuration: time.Hour.String(), ExpiresAt: expires.UTC().Format(time.RFC3339),
	}
}

func TestSlotCLI_HeldAndBusyCarryDistinctExitCodes(t *testing.T) {
	root := t.TempDir()
	seedSlotCLILease(t, root, slotCLILease("demo", "s-1", os.Getpid(), time.Now().Add(time.Hour)))
	_, err := runSlotCLI(t, root, "acquire", "--resource", "demo", "--session", "s-2")
	if code, ok := ResolveExitCode(err); !ok || code != slotExitHeld {
		t.Errorf("acquire over a live holder: exit code (%d, %v), want %d", code, ok, slotExitHeld)
	}
	if !kanban.IsSlotLeaseHeld(err) || !strings.Contains(err.Error(), "lane-s-1") {
		t.Errorf("held error = %v, want the held sentinel naming the holder", err)
	}

	busy := slotResult(fmt.Errorf("%w (waited 1s): x", kanban.ErrSlotLeaseBusy))
	if code, ok := ResolveExitCode(busy); !ok || code != slotExitBusy {
		t.Errorf("busy: exit code (%d, %v), want %d", code, ok, slotExitBusy)
	}
	if !kanban.IsSlotLeaseBusy(busy) || kanban.IsSlotLeaseHeld(busy) || !strings.Contains(busy.Error(), "retry") {
		t.Errorf("busy error = %v, want the busy sentinel (not held) and a retry hint", busy)
	}
	plain := errors.New("other")
	if slotResult(plain) != plain || slotResult(nil) != nil {
		t.Errorf("slotResult must pass other errors and nil through unchanged")
	}
}

func TestSlotCLI_StatusStatesAndList(t *testing.T) {
	root := t.TempDir()
	dead := 0x7FFFFFF0
	if kanban.FactoryProcessAlive(dead) {
		t.Skip("seeded dead pid is live on this machine")
	}
	seedSlotCLILease(t, root, slotCLILease("held-one", "s-1", os.Getpid(), time.Now().Add(time.Hour)))
	seedSlotCLILease(t, root, slotCLILease("gone-one", "s-2", dead, time.Now().Add(time.Hour)))
	seedSlotCLILease(t, root, slotCLILease("late-one", "s-3", os.Getpid(), time.Now().Add(-time.Hour)))

	for resource, want := range map[string]string{
		"held-one": "slot held-one: held\n",
		"gone-one": "held by a session that is gone (reclaimable)",
		"late-one": "past its declared bound (reclaimable)",
		"free-one": "slot free-one: free",
	} {
		out, err := runSlotCLI(t, root, "status", "--resource", resource)
		if err != nil {
			t.Fatalf("status %s: %v", resource, err)
		}
		if !strings.Contains(out, want) {
			t.Errorf("status %s = %q, want it to contain %q", resource, out, want)
		}
	}

	out, err := runSlotCLI(t, root, "status", "--resource", "held-one", "--json")
	if err != nil {
		t.Fatalf("status --json: %v", err)
	}
	var one map[string]any
	if err := json.Unmarshal([]byte(out), &one); err != nil || one["held"] != true || one["lease"] == nil {
		t.Errorf("status --json for a held resource = %s (err %v), want held true with the lease", out, err)
	}

	out, err = runSlotCLI(t, root, "status")
	if err != nil {
		t.Fatalf("status (list): %v", err)
	}
	for _, name := range []string{"held-one", "gone-one", "late-one"} {
		if !strings.Contains(out, "slot "+name+":") {
			t.Errorf("list does not show %s:\n%s", name, out)
		}
	}
	out, err = runSlotCLI(t, root, "status", "--json")
	if err != nil {
		t.Fatalf("status --json (list): %v", err)
	}
	var list struct {
		Leases []map[string]any `json:"leases"`
	}
	if err := json.Unmarshal([]byte(out), &list); err != nil || len(list.Leases) != 3 {
		t.Errorf("list --json = %s (err %v), want 3 leases", out, err)
	}

	empty := t.TempDir()
	if out, err := runSlotCLI(t, empty, "status"); err != nil || !strings.Contains(out, "no slot leases recorded") {
		t.Errorf("empty list = %q (err %v)", out, err)
	}
}

func TestSlotCLI_ReleaseRoundTripAndRefusals(t *testing.T) {
	root := t.TempDir()
	if _, err := runSlotCLI(t, root, "acquire", "--resource", "demo", "--session", "s-1"); err != nil {
		t.Fatalf("acquire: %v", err)
	}
	rec := readSlotCLIRecord(t, root)
	if rec["max_duration"] != (30 * time.Minute).String() {
		t.Errorf("acquire without --max-duration recorded %v, want the configured default 30m0s", rec["max_duration"])
	}
	if _, err := runSlotCLI(t, root, "release", "--resource", "demo", "--session", "s-2"); !kanban.IsSlotLeaseForeign(err) {
		t.Errorf("foreign release: err = %v, want the foreign sentinel", err)
	}
	if _, err := runSlotCLI(t, root, "release", "--resource", "demo"); err == nil || !strings.Contains(err.Error(), "--session") {
		t.Errorf("release without a session id: err = %v, want a --session hint", err)
	}
	out, err := runSlotCLI(t, root, "release", "--resource", "demo", "--session", "s-1", "--json")
	if err != nil {
		t.Fatalf("holder release: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(out), &got); err != nil || got["released"] != true {
		t.Errorf("release --json = %s (err %v)", out, err)
	}
	if _, err := runSlotCLI(t, root, "release", "--resource", "demo", "--session", "s-1"); !kanban.IsSlotLeaseNotHeld(err) {
		t.Errorf("empty release: err = %v, want the not-held sentinel", err)
	}
	if _, err := runSlotCLI(t, root, "acquire", "--resource", "demo", "--session", "s-1"); err != nil {
		t.Fatalf("re-acquire: %v", err)
	}
	if out, err := runSlotCLI(t, root, "release", "--resource", "demo", "--session", "s-1"); err != nil || !strings.Contains(out, "released (was s-1)") {
		t.Errorf("text release = %q (err %v)", out, err)
	}
}

func TestSlotCLI_InputRefusals(t *testing.T) {
	root := t.TempDir()
	for _, args := range [][]string{
		{"acquire", "--resource", "../escape", "--session", "s-1"},
		{"release", "--resource", "Bad Name", "--session", "s-1"},
		{"status", "--resource", "a/b"},
	} {
		if _, err := runSlotCLI(t, root, args...); !kanban.IsSlotResourceNameInvalid(err) {
			t.Errorf("%v: err = %v, want the invalid-name sentinel", args, err)
		}
	}
	for _, bound := range []string{"0", "-1m", "soon"} {
		if _, err := runSlotCLI(t, root, "acquire", "--resource", "demo", "--session", "s-1", "--max-duration", bound); !errors.Is(err, kanban.ErrSlotLeaseBoundInvalid) {
			t.Errorf("--max-duration %s: err = %v, want ErrSlotLeaseBoundInvalid", bound, err)
		}
	}
	if _, err := os.Stat(filepath.Join(root, ".moai", "state", "slot-leases", "demo.json")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("a refused acquire left a record (stat err %v)", err)
	}
}
