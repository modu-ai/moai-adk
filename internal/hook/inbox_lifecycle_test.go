package hook

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// ---------------------------------------------------------------------------
// M1 — schema version field (AC-IBX-007) + cap constants (plan.md §F M1).
// ---------------------------------------------------------------------------

// TestLessonsInboxStub_GoldenJSONCarriesSchemaVersion covers AC-IBX-007 part 1:
// a stub the collector appends parses as JSON carrying an integer version
// field equal to 1 (REQ-IBX-008).
func TestLessonsInboxStub_GoldenJSONCarriesSchemaVersion(t *testing.T) {
	root := t.TempDir()

	appendLessonsInboxStub(root, "tool_failure:Bash:ExitError", "exit status 1", "tool:Bash")

	data, err := os.ReadFile(filepath.Join(root, ".moai", "lessons-inbox.jsonl"))
	if err != nil {
		t.Fatalf("read lessons-inbox.jsonl: %v", err)
	}
	line := strings.TrimSpace(string(data))
	if line == "" {
		t.Fatal("no stub line written")
	}
	var raw map[string]any
	if err := json.Unmarshal([]byte(line), &raw); err != nil {
		t.Fatalf("stub line is not JSON: %v (%s)", err, line)
	}
	got, present := raw["v"]
	if !present {
		t.Fatalf("stub line carries no version field: %s", line)
	}
	num, isNum := got.(float64)
	if !isNum {
		t.Fatalf("version field is not a JSON number: %v (%T)", got, got)
	}
	if num != 1 {
		t.Fatalf("version = %v, want 1", num)
	}
	// Golden form: encoding/json renders the int as the bare literal `"v":1`.
	if !strings.Contains(line, `"v":1`) {
		t.Fatalf("stub line does not carry the integer literal \"v\":1: %s", line)
	}
}

// TestInboxStubVersion_AbsenceReadsAsV1 covers AC-IBX-007 part 2: a pre-upgrade
// line without the version field resolves as version 1 when parsed by the
// SPEC's reader (REQ-IBX-008 absence tolerance).
func TestInboxStubVersion_AbsenceReadsAsV1(t *testing.T) {
	// Pre-upgrade line: the REQ-HRR-006 schema carried no version field.
	var pre map[string]any
	preLine := `{"timestamp":"2026-01-01T00:00:00Z","event_key":"tool_failure:Bash:ExitError","summary":"exit status 1","source":"tool:Bash"}`
	if err := json.Unmarshal([]byte(preLine), &pre); err != nil {
		t.Fatalf("unmarshal pre-upgrade line: %v", err)
	}
	if got := InboxStubVersion(pre); got != 1 {
		t.Errorf("InboxStubVersion(pre-upgrade line) = %d, want 1", got)
	}

	// An explicit version passes through unchanged.
	var explicit map[string]any
	if err := json.Unmarshal([]byte(`{"v":2,"event_key":"k"}`), &explicit); err != nil {
		t.Fatalf("unmarshal explicit line: %v", err)
	}
	if got := InboxStubVersion(explicit); got != 2 {
		t.Errorf("InboxStubVersion(explicit v:2) = %d, want 2", got)
	}

	// An empty object (absence, not just a missing field) also reads as 1.
	if got := InboxStubVersion(map[string]any{}); got != 1 {
		t.Errorf("InboxStubVersion(empty) = %d, want 1", got)
	}
}

// TestInboxCapConstants_PinnedDefaults pins the M1-finalized cap constants
// (plan.md §F M1: "cap default 1 MiB finalized in M1"). The literals here are
// the contract, not duplicates: the test fails if the single-source default is
// retuned without a deliberate decision.
func TestInboxCapConstants_PinnedDefaults(t *testing.T) {
	if config.DefaultInboxMaxBytes != 1<<20 {
		t.Errorf("DefaultInboxMaxBytes = %d, want %d (1 MiB)", config.DefaultInboxMaxBytes, 1<<20)
	}
	if config.DefaultInboxArchiveGenerations != 2 {
		t.Errorf("DefaultInboxArchiveGenerations = %d, want 2", config.DefaultInboxArchiveGenerations)
	}
}

// ---------------------------------------------------------------------------
// M2 — write-time cap + rotation + curator stand-down
// (AC-IBX-001/002/003/008 + lane-adopted NFC-4 state-dir guard).
// ---------------------------------------------------------------------------

// inboxSeedLine is a fixed 1024-byte JSONL line (1023 'x' payload + '\n') used
// to seed the live inbox at measured sizes.
func inboxSeedLine() []byte {
	line := make([]byte, 1024)
	for i := range line {
		line[i] = 'x'
	}
	line[1023] = '\n'
	return line
}

// seedInbox writes n 1024-byte lines to the live inbox and returns the bytes.
func seedInbox(t *testing.T, root string, nLines int) []byte {
	t.Helper()
	data := bytes.Repeat(inboxSeedLine(), nLines)
	path := filepath.Join(root, ".moai", "lessons-inbox.jsonl")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir .moai: %v", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("seed inbox: %v", err)
	}
	return data
}

// archiveGens counts archive generation files (lessons-inbox.jsonl.<n>) under
// the inbox directory.
func archiveGens(t *testing.T, root string) int {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(root, ".moai"))
	if err != nil {
		t.Fatalf("read .moai: %v", err)
	}
	count := 0
	for _, e := range entries {
		if !e.IsDir() && strings.HasPrefix(e.Name(), "lessons-inbox.jsonl.") {
			count++
		}
	}
	return count
}

// dirSnapshot hashes every entry under dir (path, size, mode, content) into a
// comparable map. Used by the lane-adopted NFC-4 guard: the LSEL state dir
// must be byte-unchanged by any cap-path activity.
func dirSnapshot(t *testing.T, dir string) map[string]string {
	t.Helper()
	snap := map[string]string{}
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, relErr := filepath.Rel(dir, path)
		if relErr != nil {
			return relErr
		}
		info, infoErr := d.Info()
		if infoErr != nil {
			return infoErr
		}
		entry := fmt.Sprintf("mode=%s size=%d", info.Mode(), info.Size())
		if !d.IsDir() {
			data, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			sum := sha256.Sum256(data)
			entry += " content=" + hex.EncodeToString(sum[:])
		}
		snap[rel] = entry
		return nil
	})
	if err != nil {
		t.Fatalf("snapshot %s: %v", dir, err)
	}
	return snap
}

// TestInboxCap_RotatesOverCapInbox covers AC-IBX-001 (REQ-IBX-001, REQ-IBX-003):
// an append observing the live inbox over the cap rotates the pre-append
// content into generation .1 and lands the new stub on a fresh live file whose
// size stays under cap + one-stub margin. Also guards the NFC-4 flip side: the
// cap path must not CREATE the marker tree (.moai/state) on a marker-absent
// install — creating it would stand the cap down forever after.
func TestInboxCap_RotatesOverCapInbox(t *testing.T) {
	root := t.TempDir()
	seed := seedInbox(t, root, int(config.DefaultInboxMaxBytes/1024)+2) // cap + 2KiB

	appendLessonsInboxStub(root, "tool_failure:Bash:ExitError", "exit status 1", "tool:Bash")

	gen1, err := os.ReadFile(filepath.Join(root, ".moai", "lessons-inbox.jsonl.1"))
	if err != nil {
		t.Fatalf("archive generation .1 not created: %v", err)
	}
	if !bytes.Equal(gen1, seed) {
		t.Errorf("archive .1 does not carry the pre-append content (got %d bytes, seeded %d)", len(gen1), len(seed))
	}
	live, err := os.ReadFile(filepath.Join(root, ".moai", "lessons-inbox.jsonl"))
	if err != nil {
		t.Fatalf("live inbox missing after rotation: %v", err)
	}
	liveLines := strings.Count(strings.TrimRight(string(live), "\n"), "\n") + 1
	if liveLines != 1 {
		t.Errorf("live file is not fresh after rotation: %d lines", liveLines)
	}
	if int64(len(live)) >= config.DefaultInboxMaxBytes+1024 {
		t.Errorf("live size %d exceeds cap + one-stub margin", len(live))
	}
	if _, err := os.Stat(filepath.Join(root, ".moai", "state")); !os.IsNotExist(err) {
		t.Errorf("cap path created .moai/state (NFC-4: the marker tree must never be created): %v", err)
	}
}

// TestInboxCap_BoundaryAtCapRotates covers the acceptance §B boundary edge:
// a live inbox EXACTLY at the cap rotates ("at or over", REQ-IBX-003).
func TestInboxCap_BoundaryAtCapRotates(t *testing.T) {
	root := t.TempDir()
	seedInbox(t, root, int(config.DefaultInboxMaxBytes/1024)) // exactly 1 MiB

	appendLessonsInboxStub(root, "k", "boundary", "test")

	if _, err := os.Stat(filepath.Join(root, ".moai", "lessons-inbox.jsonl.1")); err != nil {
		t.Fatalf("rotation did not fire at the exact-cap boundary: %v", err)
	}
}

// TestInboxStandDown_MarkerPresentNoRotation covers AC-IBX-002 (REQ-IBX-002,
// the NFC-4 parity proof): with the LSEL drain-ownership marker present, an
// over-cap inbox is never rotated — the local drain owns the inbox lifecycle.
// Lane-adopted strengthening (plan-audit N-D2): the assertion ALSO pins the
// state dir byte-unchanged across the capped appends — the read-only stat
// probe is the only permitted touch of .moai/state/lsel/.
func TestInboxStandDown_MarkerPresentNoRotation(t *testing.T) {
	root := t.TempDir()
	stateDir := filepath.Join(root, ".moai", "state", "lsel")
	if err := os.MkdirAll(filepath.Join(stateDir, "clusters-history"), 0o755); err != nil {
		t.Fatalf("mkdir state dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, "clusters.json"), []byte(`{"candidates":[]}`), 0o600); err != nil {
		t.Fatalf("write clusters.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, "clusters-history", "snap.json"), []byte(`{"candidates":[]}`), 0o600); err != nil {
		t.Fatalf("write clusters-history snapshot: %v", err)
	}
	before := dirSnapshot(t, stateDir)
	if len(before) < 3 {
		t.Fatalf("state-dir snapshot is near-empty (%d entries) — the guard would be vacuous", len(before))
	}

	seedInbox(t, root, int(config.DefaultInboxMaxBytes/1024)+2)
	for i := 0; i < 5; i++ {
		appendLessonsInboxStub(root, fmt.Sprintf("standdown:%d", i), "stub", "test")
	}

	if got := archiveGens(t, root); got != 0 {
		t.Errorf("stand-down violated: %d archive generations created", got)
	}
	live, err := os.ReadFile(filepath.Join(root, ".moai", "lessons-inbox.jsonl"))
	if err != nil {
		t.Fatalf("live inbox missing: %v", err)
	}
	if int64(len(live)) <= config.DefaultInboxMaxBytes {
		t.Errorf("live inbox did not grow past the cap under stand-down: %d bytes", len(live))
	}
	after := dirSnapshot(t, stateDir)
	if !reflect.DeepEqual(before, after) {
		t.Errorf("LSEL state dir changed under the cap path (NFC-4 violation): before=%v after=%v", before, after)
	}
}

// TestInboxRetention_MaxTwoGenerations covers AC-IBX-003 (REQ-IBX-004): after
// four forced rotations at most 2 archive generations exist. The bound is the
// LITERAL 2 (acceptance §D PASS term) — asserting against the config constant
// would let a retuned constant validate itself.
func TestInboxRetention_MaxTwoGenerations(t *testing.T) {
	root := t.TempDir()
	rotations := 0
	for round := 0; round < 4; round++ {
		seedInbox(t, root, int(config.DefaultInboxMaxBytes/1024)+1)
		appendLessonsInboxStub(root, fmt.Sprintf("retention:%d", round), "stub", "test")
		rotations++
		if got := archiveGens(t, root); got > 2 {
			t.Fatalf("after rotation %d: %d archive generations (bound 2)", rotations, got)
		}
	}
	if archiveGens(t, root) < 1 {
		t.Fatal("no archive generation after 4 over-cap rounds — rotations never fired (vacuous pass)")
	}
}

// TestInboxRotation_PreEraLeftoverArchivesIdempotent covers the acceptance §B
// pre-era edge: archive paths left by a prior SPEC-less era are absorbed by
// the delete-then-rename chain without error, keeping the generation count
// bounded.
func TestInboxRotation_PreEraLeftoverArchivesIdempotent(t *testing.T) {
	root := t.TempDir()
	moai := filepath.Join(root, ".moai")
	if err := os.MkdirAll(moai, 0o755); err != nil {
		t.Fatalf("mkdir .moai: %v", err)
	}
	if err := os.WriteFile(filepath.Join(moai, "lessons-inbox.jsonl.1"), []byte("pre-era-gen1"), 0o600); err != nil {
		t.Fatalf("seed .1: %v", err)
	}
	if err := os.WriteFile(filepath.Join(moai, "lessons-inbox.jsonl.2"), []byte("pre-era-gen2"), 0o600); err != nil {
		t.Fatalf("seed .2: %v", err)
	}
	seed := seedInbox(t, root, int(config.DefaultInboxMaxBytes/1024)+1)

	appendLessonsInboxStub(root, "k", "post-era", "test") // must not panic

	if got := archiveGens(t, root); got > 2 {
		t.Fatalf("%d archive generations after absorbing pre-era leftovers (bound 2)", got)
	}
	gen2, err := os.ReadFile(filepath.Join(moai, "lessons-inbox.jsonl.2"))
	if err != nil {
		t.Fatalf("shift .1 -> .2 lost the pre-era generation .1: %v", err)
	}
	if string(gen2) != "pre-era-gen1" {
		t.Errorf("generation shift wrong: .2 = %q, want the shifted pre-era .1", gen2)
	}
	gen1, err := os.ReadFile(filepath.Join(moai, "lessons-inbox.jsonl.1"))
	if err != nil {
		t.Fatalf("live -> .1 rotation missing: %v", err)
	}
	if !bytes.Equal(gen1, seed) {
		t.Errorf("archive .1 does not carry the rotated live content")
	}
}

// TestInboxCap_FailOpenOnRotationFailure covers AC-IBX-008 (REQ-IBX-009):
// rotation sabotaged by an undeletable/under-shiftable archive destination →
// no panic, the failure is logged as a warning, and the append still lands
// best-effort on the existing live file. The session must never block.
func TestInboxCap_FailOpenOnRotationFailure(t *testing.T) {
	root := t.TempDir()
	seedInbox(t, root, int(config.DefaultInboxMaxBytes/1024)+1)
	// Sabotage: generation .2 is a non-empty DIRECTORY and generation .1 is a
	// real file, so the .1 -> .2 shift (file renamed over a non-empty dir)
	// fails and the chain errors at that link.
	moaiDir := filepath.Join(root, ".moai")
	if err := os.WriteFile(filepath.Join(moaiDir, "lessons-inbox.jsonl.1"), []byte("prior-gen1"), 0o600); err != nil {
		t.Fatalf("seed .1: %v", err)
	}
	sabotage := filepath.Join(moaiDir, "lessons-inbox.jsonl.2")
	if err := os.MkdirAll(sabotage, 0o755); err != nil {
		t.Fatalf("sabotage dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sabotage, "occupied"), []byte("x"), 0o600); err != nil {
		t.Fatalf("sabotage payload: %v", err)
	}

	var buf bytes.Buffer
	orig := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})))
	t.Cleanup(func() { slog.SetDefault(orig) })

	appendLessonsInboxStub(root, "tool_failure:Bash:ExitError", "after sabotage", "tool:Bash")

	if !strings.Contains(buf.String(), "rotation failed") {
		t.Errorf("rotation failure was not logged as a warning; log = %q", buf.String())
	}
	live, err := os.ReadFile(filepath.Join(root, ".moai", "lessons-inbox.jsonl"))
	if err != nil {
		t.Fatalf("live inbox missing: %v", err)
	}
	if !strings.Contains(string(live), `"event_key":"tool_failure:Bash:ExitError"`) {
		t.Errorf("appended stub did not land on the existing live file after failed rotation")
	}
	if strings.Count(strings.TrimRight(string(live), "\n"), "\n")+1 < int(config.DefaultInboxMaxBytes/1024)+1 {
		t.Errorf("live file was rotated despite sabotaged chain (append must land on the EXISTING file)")
	}
}

// TestInboxCap_ConcurrentAppendsCrossCapBoundary covers AC-IBX-010 (NFC-1):
// N concurrent appenders crossing the cap boundary race the rotation. Run
// under -race (plan.md §E E5). PASS terms: no panic (completion), the
// retention bound holds under race, and the final live size is bounded — the
// cap machinery stayed active despite the racing (the no-op-cap mutant leaves
// the seeded ~cap bytes in live and fails this bound).
func TestInboxCap_ConcurrentAppendsCrossCapBoundary(t *testing.T) {
	root := t.TempDir()
	// Seed just under the cap so the first racing appends cross the boundary
	// inside one rotation window.
	seedInbox(t, root, int(config.DefaultInboxMaxBytes/1024)-1)

	const workers = 8
	const perWorker = 25
	const maxStubBytes = 1024
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for i := 0; i < perWorker; i++ {
				appendLessonsInboxStub(root, fmt.Sprintf("race:%d:%d", w, i), "race stub", "test")
			}
		}(w)
	}
	wg.Wait()

	if got := archiveGens(t, root); got > 2 {
		t.Errorf("retention bound broken under race: %d generations", got)
	}
	if archiveGens(t, root) < 1 {
		t.Fatal("no rotation observed across the cap boundary — boundary never crossed (vacuous)")
	}
	info, err := os.Stat(filepath.Join(root, ".moai", "lessons-inbox.jsonl"))
	if err != nil {
		t.Fatalf("live inbox missing after race: %v", err)
	}
	if bound := int64(config.DefaultInboxMaxBytes) + int64(workers*perWorker*maxStubBytes); info.Size() > bound {
		t.Errorf("final live size %d exceeds the bounded expectation %d (cap machinery inert?)", info.Size(), bound)
	}
}
