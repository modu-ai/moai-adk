package cli

import (
	"context"
	"database/sql"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/factorymsg"
	mcpcat "github.com/modu-ai/moai-adk/internal/mcp"
)

// storeInventory lists every file under roots that could be a message store,
// a database root, or a private socket: SQLite files and their journals, and
// any socket or named pipe.
func storeInventory(t *testing.T, roots ...string) []string {
	t.Helper()
	var out []string
	for _, root := range roots {
		_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				if d.Name() == ".git" {
					return filepath.SkipDir
				}
				return nil
			}
			info, err := d.Info()
			if err != nil {
				return nil
			}
			name := d.Name()
			isStore := strings.HasSuffix(name, ".db") || strings.HasSuffix(name, ".sqlite") || strings.HasSuffix(name, ".sqlite3") ||
				strings.HasSuffix(name, "-wal") || strings.HasSuffix(name, "-shm") || strings.HasSuffix(name, "-journal")
			if isStore || info.Mode()&(fs.ModeSocket|fs.ModeNamedPipe) != 0 {
				out = append(out, p)
			}
			return nil
		})
	}
	sort.Strings(out)
	return out
}

// storeFiles strips the SQLite journal files, whose presence depends on
// checkpoint timing rather than on which stores exist.
func storeFiles(paths []string) []string {
	var out []string
	for _, p := range paths {
		if strings.HasSuffix(p, "-wal") || strings.HasSuffix(p, "-shm") || strings.HasSuffix(p, "-journal") {
			continue
		}
		out = append(out, p)
	}
	return out
}

// processRecorder instruments the handoff subprocess seam and keeps every
// command it built, so the test can prove each one was a finished git.
type processRecorder struct {
	mu   sync.Mutex
	cmds []*exec.Cmd
}

func (r *processRecorder) install(t *testing.T) {
	t.Helper()
	prev := handoffCommand
	handoffCommand = func(name string, args ...string) *exec.Cmd {
		cmd := prev(name, args...)
		r.mu.Lock()
		r.cmds = append(r.cmds, cmd)
		r.mu.Unlock()
		return cmd
	}
	t.Cleanup(func() { handoffCommand = prev })
}

// TestFactoryLaneHandoffT1074Compatibility is AC-FLH-015 (REQ-FLH-015): with
// the handoff additions exercised end to end, the t1074 broker, roster, and
// receipt behavior still holds on the same broker, the MCP catalog stays at 39
// tools (15 write-capable, 24 read-only — 36/14/22 when the handoff landed,
// plus the Codex read-only role launcher's start/status/result trio added
// later) and matches the registered server,
// and the handoff adds no broker database root, daemon, private socket, or
// parallel message store.
func TestFactoryLaneHandoffT1074Compatibility(t *testing.T) {
	// MCP catalog: size, write/read split, and registration equality.
	var write, read int
	for _, def := range mcpcat.MoaiMCPTools() {
		if def.WriteCapable {
			write++
		} else {
			read++
		}
	}
	if total := write + read; total != 39 || write != 15 || read != 24 {
		t.Fatalf("MCP catalog = %d total / %d write / %d read, want 39/15/24", total, write, read)
	}
	registered := listToolNames(t)
	if len(registered) != 39 {
		t.Fatalf("registered MCP tools = %d, want 39", len(registered))
	}
	inCatalog := map[string]bool{}
	for _, def := range mcpcat.MoaiMCPTools() {
		inCatalog[def.Name] = true
	}
	for _, n := range registered {
		if !inCatalog[n] {
			t.Fatalf("registered tool %q has no catalog entry", n)
		}
	}

	f := newLaneHandoffFixture(t, "develop", true)
	withFakeHandoffAppServer(t, &fakeHandoffAppServer{newThreadID: "thr-forked"})
	procs := &processRecorder{}
	procs.install(t)
	ctx := context.Background()
	lead := abandonTestLead(t, f)
	home := os.Getenv("MOAI_HOME")
	if home == "" {
		t.Fatal("fixture MOAI_HOME is unset")
	}
	before := storeFiles(storeInventory(t, home, f.primary))
	if len(before) == 0 {
		t.Fatal("store inventory found nothing before the handoff: the probe is blind")
	}

	// The whole handoff surface: reservation, creation, relocation, BOUND,
	// restart recovery, and the operator command on a second lane.
	f.driveHeadless(ctx, t)
	if got := f.reconcile(t); got.Decision != laneRecoveryFinalize {
		t.Fatalf("recovery after BOUND = %s", got.Decision)
	}

	// t1074 broker, roster, and receipt behavior on a lane without a handoff,
	// in the same run.
	pending, err := f.store.RegisterLaunchPending(ctx, factorymsg.Peer{ProjectKey: "project", RunID: f.run, Backend: "codex", Role: "worker", Slot: "lane-2", PID: 999_983, ProcessStart: "lane-2-start"})
	if err != nil {
		t.Fatalf("t1074 launcher registration: %v", err)
	}
	pending.SessionUUID = "lane-2-uuid"
	lane2, ok, err := f.store.BindLaunchPending(ctx, pending)
	if err != nil || !ok || lane2.Generation != pending.Generation+1 {
		t.Fatalf("t1074 launcher bind = %+v ok=%v err=%v", lane2, ok, err)
	}
	env, err := f.store.Send(ctx, factorymsg.SendRequest{From: lead, To: lane2, Kind: factorymsg.KindDispatchNotice, IdempotencyKey: "legacy", TaskRef: "t1082", CorrelationID: "c-legacy", TTL: time.Hour, Payload: []byte("legacy-body")})
	if err != nil {
		t.Fatalf("t1074 send: %v", err)
	}
	claims, err := f.store.Claim(ctx, lane2, factorymsg.MaxBatch, time.Hour)
	if err != nil || len(claims) != 1 || claims[0].ID != env.ID {
		t.Fatalf("t1074 claim = %+v err=%v", claims, err)
	}
	if body, err := f.store.ReadBody(ctx, lane2, env.ID, claims[0].ClaimToken); err != nil || string(body) != "legacy-body" {
		t.Fatalf("t1074 body = %q err=%v", body, err)
	}
	if err := f.store.RecordDisposition(ctx, lane2, env.ID, claims[0].ClaimToken, factorymsg.DispositionAccepted); err != nil {
		t.Fatal(err)
	}
	if err := f.store.Receipt(ctx, lane2, env.ID, claims[0].ClaimToken); err != nil {
		t.Fatalf("t1074 receipt: %v", err)
	}
	status, err := f.store.Status(ctx)
	if err != nil {
		t.Fatal(err)
	}
	roster := map[string]factorymsg.LaneStatus{}
	for _, l := range status.Lanes {
		roster[l.Slot] = l
	}
	if l := roster["lane-2"]; l.SessionUUID != "lane-2-uuid" || l.Generation != lane2.Generation {
		t.Fatalf("t1074 roster lane-2 = %+v", l)
	}
	if l := roster[handoffTestSlot]; l.SessionUUID != "thr-forked" || l.Generation != f.source.Generation+1 {
		t.Fatalf("roster lane-1 after BOUND = %+v", l)
	}
	if _, ok := roster["lead"]; !ok || len(roster) != 3 {
		t.Fatalf("roster = %+v, want lead, lane-1, lane-2", status.Lanes)
	}
	if status.Acknowledged < 1 {
		t.Fatalf("status acknowledged = %d", status.Acknowledged)
	}

	// No new store: the same database files exist, the handoff tables live in
	// the existing broker file, and no socket appeared.
	after := storeInventory(t, home, f.primary)
	if got := storeFiles(after); strings.Join(got, "\n") != strings.Join(before, "\n") {
		t.Fatalf("store inventory changed:\nbefore=%v\nafter=%v", before, got)
	}
	brokerPath, err := factorymsg.BrokerPath(f.primary, f.run)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, p := range before {
		found = found || p == brokerPath
	}
	if !found {
		t.Fatalf("broker %s not in the inventory %v", brokerPath, before)
	}
	db, err := sql.Open("sqlite", "file:"+filepath.ToSlash(brokerPath)+"?mode=ro")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	var tables int
	if err := db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='table' AND name IN ('lane_handoffs','lane_endpoint_tombstones','lane_handoff_receipts','lane_dispatch_releases','peers','messages')`).Scan(&tables); err != nil || tables != 6 {
		t.Fatalf("handoff and t1074 tables in the broker file = %d err=%v, want 6", tables, err)
	}

	// No daemon: every subprocess the handoff started was git and has exited.
	procs.mu.Lock()
	defer procs.mu.Unlock()
	if len(procs.cmds) == 0 {
		t.Fatal("subprocess seam recorded nothing: the probe is blind")
	}
	for _, c := range procs.cmds {
		if filepath.Base(c.Path) != "git" && filepath.Base(c.Path) != "git.exe" {
			t.Fatalf("handoff started %s, not git", c.Path)
		}
		if c.ProcessState == nil {
			t.Fatalf("handoff subprocess %v was never waited: a lingering process", c.Args)
		}
	}
	t.Logf("AC_FLH_015 catalog=39/15/24 stores=%d git_calls=%d", len(before), len(procs.cmds))
}
