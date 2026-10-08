// m0_journal_red_test.go — SPEC-USERASSET-DEPLOY-GUARD-001 M0 RED-first
// reproduction battery, journal family: AC-002 (first-stage merge), AC-003
// (carried-hash refresh), AC-004 (per-file WriteCompleted persistence),
// AC-005 (unknown schema refusal).
//
// M0 discipline: observation only. Each test is written to the acceptance.md
// contract and RUN AT HEAD; the verbatim outcome is recorded in progress.md
// §E.2 and no production code is touched. RED must be red for the RIGHT
// reason — the intended assertion is named per test.
package userassets

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"
	"time"
)

// parkingSource wraps a MapFS and parks the Nth ReadFile of one source path
// on a channel — the deterministic mid-loop interruption seam the M0 risk
// note calls for (table tests + injected suspension, no production change).
// Read counting for a skill source path (.claude/skills/<name>/SKILL.md):
// installTargets reads it once per skill-root slug (claude-skills +
// agents-skills = reads #1-#2), then applyTarget reads it once per sorted
// target (reads #3 = the agents-skills target, #4 = the claude-skills
// target). Parking read #4 suspends the write loop AFTER the targets that
// sort before claude-skills — including claude-agents/manager-x.md — were
// fully applied.
type parkingSource struct {
	inner   fs.FS
	parkAt  string
	parkOn  int
	seen    int
	parked  chan struct{} // closed when the park is reached
	release chan struct{} // closed by the test to resume the run
}

func newParkingSource(inner fstest.MapFS, parkAt string, parkOn int) *parkingSource {
	return &parkingSource{
		inner:   inner,
		parkAt:  parkAt,
		parkOn:  parkOn,
		parked:  make(chan struct{}),
		release: make(chan struct{}),
	}
}

func (p *parkingSource) Open(name string) (fs.File, error) { return p.inner.Open(name) }

func (p *parkingSource) ReadFile(name string) ([]byte, error) {
	if name == p.parkAt {
		p.seen++
		if p.seen == p.parkOn {
			close(p.parked)
			<-p.release
		}
	}
	return fs.ReadFile(p.inner, name)
}

func waitChannel(t *testing.T, ch chan struct{}, what string) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(15 * time.Second):
		t.Fatalf("%s did not arrive within 15s — the run never reached the expected point", what)
	}
}

func shaHex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// seedJournalStartedAt marks the pre-seeded journal so a watcher can tell
// the SEEDED bytes from the run's own staging writes (the run stamps
// StartedAt with its now()).
const seedJournalStartedAt = "2000-01-01T00:00:00Z"

// seedRecoveredJournal stages a pending-install journal carrying one entry
// for extraRel whose on-disk bytes hash to entrySHA — a previous run's
// recovery evidence the next run must carry through its own staging.
func seedRecoveredJournal(t *testing.T, home, extraRel, entrySHA string, writeCompleted bool) {
	t.Helper()
	j := &PendingJournal{
		SchemaVersion:    SchemaVersion,
		BundlesSelection: []string{"extras"},
		StartedAt:        seedJournalStartedAt,
		Entries: []JournalEntry{{
			Path:           extraRel,
			ExpectedSHA256: entrySHA,
			Bundle:         "extras",
			MoaiVersion:    "v3.2.0-old",
			InstalledAt:    "2026-10-08T00:00:00Z",
			WriteCompleted: writeCompleted,
		}},
	}
	if err := WriteJournal(JournalPath(home), j); err != nil {
		t.Fatalf("seed journal: %v", err)
	}
}

// TestWriteCompletedPersistedPerFile — AC-004 (ledger 10, REQ-JRN-003).
// Given a multi-file install run interrupted after the FIRST file's write,
// the journal ON DISK at interruption time must already persist that file's
// WriteCompleted=true. RED-now reason: install.go applies every flag only
// AFTER the whole loop (:273-278) and persists once at :300-304, so an
// interruption mid-loop leaves every flag unrecorded — the resumed run
// mis-reads its own install as a collision (REQ-010) instead of claiming it.
func TestWriteCompletedPersistedPerFile(t *testing.T) {
	f := newFixture(t)
	// Park read #4 of the alpha SKILL.md source: by then the targets sorting
	// before claude-skills/moai-alpha — agents-skills SKILL.md, its
	// workflows/a.md, and claude-agents/manager-x.md — are fully written.
	src := newParkingSource(f.src, ".claude/skills/moai-alpha/SKILL.md", 4)
	inst := f.installer(t)
	inst.Source = src

	done := make(chan error, 1)
	go func() { _, err := inst.Install(nil); done <- err }()
	waitChannel(t, src.parked, "mid-loop park")

	j, err := LoadJournal(JournalPath(f.home))
	if err != nil {
		t.Fatalf("load journal mid-run: %v", err)
	}
	if j == nil {
		t.Fatal("no journal staged at interruption time — the stage write never happened")
	}
	var first *JournalEntry
	for i := range j.Entries {
		if j.Entries[i].Path == "claude-agents/manager-x.md" {
			first = &j.Entries[i]
			break
		}
	}
	if first == nil {
		t.Fatalf("staged journal lacks the first file's entry; entries: %+v", j.Entries)
	}
	// Resume the suspended run and let it finish before the test returns —
	// a still-running writer races the TempDir cleanup.
	close(src.release)
	select {
	case <-done:
	case <-time.After(30 * time.Second):
		t.Fatal("the resumed install never finished")
	}
	if !first.WriteCompleted {
		t.Fatalf("RED (intended): journal on disk mid-run shows WriteCompleted=false for the already-written first file claude-agents/manager-x.md — flags are batched after the loop instead of persisted per file")
	}
}

// TestJournalStageCarriesRecoveredEntries — AC-002 (ledger 3, REQ-JRN-001).
// Given a journal from an interrupted run and a new install run, the FIRST
// staging (install.go:239) must already contain the recovered entries, so an
// interruption at ANY point of the run loses none. RED-now reason: the
// carry-in happens after the write loop (:287-299) and re-persists at :301,
// so the first staging's file carries only the new delta — an interruption
// between :239 and :301 drops the recovered entry.
func TestJournalStageCarriesRecoveredEntries(t *testing.T) {
	f := newFixture(t)
	const recoveredKey = "claude-skills/moai-beta/extra.md"
	recoveredBytes := []byte("recovered file from the interrupted extras run\n")
	abs := filepath.Join(f.home, ".claude", "skills", "moai-beta", "extra.md")
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(abs, recoveredBytes, 0o644); err != nil {
		t.Fatal(err)
	}
	seedRecoveredJournal(t, f.home, recoveredKey, shaHex(recoveredBytes), true)

	src := newParkingSource(f.src, ".claude/skills/moai-alpha/SKILL.md", 3)
	inst := f.installer(t)
	inst.Source = src

	done := make(chan error, 1)
	go func() { _, err := inst.Install(nil); done <- err }()
	waitChannel(t, src.parked, "mid-loop park")

	j, err := LoadJournal(JournalPath(f.home))
	if err != nil {
		t.Fatalf("load journal mid-run: %v", err)
	}
	if j == nil {
		t.Fatal("no journal staged at interruption time")
	}
	carried := false
	for _, e := range j.Entries {
		if e.Path == recoveredKey {
			carried = true
		}
	}
	// Resume the suspended run and let it finish — a still-running writer
	// races the TempDir cleanup.
	close(src.release)
	select {
	case <-done:
	case <-time.After(30 * time.Second):
		t.Fatal("the resumed install never finished")
	}
	if carried {
		return // carried already at first-staging time — contract held
	}
	t.Fatalf("RED (intended): the journal file staged at :239 lacks the recovered entry %s (present only after the post-loop carry-in) — an interruption before :301 drops it. Entries on disk: %+v", recoveredKey, j.Entries)
}

// TestJournalRefreshUpdatesCarriedHash — AC-003 (ledger 6a, REQ-JRN-002).
// Given a recovered entry whose target is refreshed to shipped bytes during
// the run, the CARRIED journal entry's ExpectedSHA256 must equal the
// refreshed bytes' hash — never the stale recorded one. The M0 RED (stale
// hash carried verbatim) is recorded in progress.md §E.2; the observation
// mechanism was re-aimed at M1's persistence timing (the M0 trigger watched
// the post-loop re-persist, which the per-file persistence replaced) while
// the assertion stands on the same contract: the ON-DISK journal, after an
// interrupted run, carries the refreshed hash.
//
// Observation mechanics: the watcher polls for the REFRESHED hash landing
// in the journal file — only the REQ-JRN-002 hash sync can write it — then
// makes ~/.moai read-only so the manifest save fails and the journal
// survives on disk with its synced state. The filler-padded manifest keeps
// the save window wide.
func TestJournalRefreshUpdatesCarriedHash(t *testing.T) {
	const recoveredKey = "codex-agents/manager-x.toml"
	const attempts = 40
	var lastObservation string
	captured := false

	for attempt := 0; attempt < attempts && !captured; attempt++ {
		observed, ok := attemptRefreshHashRepro(t, recoveredKey)
		lastObservation = observed
		captured = captured || ok
	}
	if !captured {
		t.Fatalf("probe inconclusive after %d attempts (tool failure, not GREEN): the hash-sync persist never landed; last observation: %s", attempts, lastObservation)
	}
}

// attemptRefreshHashRepro runs one attempt of the AC-003 repro. It returns
// (observation, true) when the carried entry was captured on disk after a
// forced save failure, and (observation, false) when the window was missed
// (the save succeeded and cleared the journal — retry).
func attemptRefreshHashRepro(t *testing.T, recoveredKey string) (string, bool) {
	f := newFixture(t)
	moaiHome := filepath.Join(f.home, ".moai")

	oldBytes := []byte("name = \"manager-x\"\nversion = \"old\"\n")
	newShipped := []byte("name = \"manager-x\"\nversion = \"new\"\n")
	newSHA := shaHex(newShipped)
	// The carrier is the codex-agents TOML — a FLAT, single-root target. A
	// skill-tree carrier enumerates the refreshed file under BOTH skill
	// roots, and the unclaimed twin root's stage entry carries the new hash
	// from the FIRST staging, firing the watcher before the hash sync even
	// ran (observed). A flat root has no twin. The recovered file exists on
	// disk with the OLD bytes; the journal hashes them (a case-2 claim) —
	// the SOURCE carries the NEW shipped bytes, so the re-evaluation arm
	// refreshes the file mid-run. The codex TOML is an L0 agent target, so
	// Install(nil) reaches it.
	abs := filepath.Join(f.home, ".codex", "agents", "manager-x.toml")
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(abs, oldBytes, 0o644); err != nil {
		t.Fatal(err)
	}
	f.src[".codex/agents/moai/manager-x.toml"] = &fstest.MapFile{Data: newShipped}
	seedRecoveredJournal(t, f.home, recoveredKey, shaHex(oldBytes), true)

	// Widen the save window: filler entries make the manifest marshal slow.
	if err := widenManifestForSaveWindow(f.home, 4000); err != nil {
		t.Fatal(err)
	}

	// Watcher: the moment the journal carries the REFRESHED hash ON THE
	// RECOVERED ENTRY, make ~/.moai read-only so the manifest save fails
	// and the journal survives with its synced state. The condition parses
	// the journal and inspects the recovered entry's hash specifically
	// (gate round 14: a raw substring match over the whole file also hits
	// staged NEW entries carrying the same bytes under other roots — the
	// watcher then fires before the hash sync even ran).
	stop := make(chan struct{})
	watcherDone := make(chan struct{})
	go func() {
		defer close(watcherDone)
		jp := JournalPath(f.home)
		for {
			select {
			case <-stop:
				return
			default:
			}
			if data, err := os.ReadFile(jp); err == nil {
				var probe PendingJournal
				if json.Unmarshal(data, &probe) == nil {
					for _, e := range probe.Entries {
						if e.Path == recoveredKey && e.ExpectedSHA256 == newSHA {
							_ = os.Chmod(moaiHome, 0o555)
							return
						}
					}
				}
			}
			time.Sleep(100 * time.Microsecond)
		}
	}()

	inst := f.installer(t)
	_, installErr := inst.Install(nil)
	close(stop)
	<-watcherDone
	// Restore writability no matter how the attempt ends.
	if err := os.Chmod(moaiHome, 0o755); err != nil {
		t.Fatalf("restore .moai perms: %v", err)
	}

	// The save must have FAILED (journal survives). A nil error means the
	// watcher lost the race — the save+clear succeeded, so retry.
	if installErr == nil {
		return "save succeeded — watcher lost the persist→save race", false
	}
	if !strings.Contains(installErr.Error(), "save user manifest") {
		return "run failed differently: " + installErr.Error(), false
	}

	data, err := os.ReadFile(JournalPath(f.home))
	if err != nil {
		return "journal unreadable after failed save: " + err.Error(), false
	}
	var j PendingJournal
	if err := json.Unmarshal(data, &j); err != nil {
		return "journal unparseable: " + err.Error(), false
	}
	diskBytes, err := os.ReadFile(abs)
	if err != nil {
		return "refreshed file unreadable: " + err.Error(), false
	}
	diskSHA := shaHex(diskBytes)
	if diskSHA != shaHex(newShipped) {
		return "file was not refreshed to shipped bytes in-run (sha " + diskSHA + ")", false
	}
	for _, e := range j.Entries {
		if e.Path == recoveredKey {
			if e.ExpectedSHA256 != newSHA {
				t.Errorf("stale carried hash survived on disk after the interrupted run: %q, want the refreshed %q — REQ-JRN-002 regressed", e.ExpectedSHA256, newSHA)
				return "captured", true
			}
			// Gate round 14 bucket B4: the hash sync carries the PROVENANCE
			// with it — new bytes recorded under the old run's origin would
			// violate REQ-006 (the per-file version names the build that
			// produced the bytes on disk).
			if e.MoaiVersion != "v3.2.0-test" {
				t.Errorf("the refreshed carried entry kept the stale provenance: MoaiVersion=%q, want the current run's v3.2.0-test", e.MoaiVersion)
			}
			return "captured — the carried entry carries the refreshed hash " + newSHA, true
		}
	}
	return "journal survived but lost the carried entry (entries: " + strconv.Itoa(len(j.Entries)) + ")", false
}

// widenManifestForSaveWindow pads the user manifest with filler file records
// so the post-loop save marshal takes long enough for the watcher to win the
// :301→:306 race reliably.
func widenManifestForSaveWindow(home string, n int) error {
	path := ManifestPath(home)
	m, err := Load(path)
	if err != nil {
		return err
	}
	for i := 0; i < n; i++ {
		m.Files[string(RootClaudeSkills)+"/filler-"+strconv.Itoa(i)+".md"] = FileEntry{
			SHA256: "0000000000000000000000000000000000000000000000000000000000000000",
			Bundle: "core",
		}
	}
	return m.Save(path)
}

// TestJournalUnknownSchemaRefused — AC-005 (ledger 10a-재정식, REQ-JRN-004).
// Given a journal whose schema_version is not the current value, LoadJournal
// must refuse it with a diagnostic (no silent decode). RED-now reason: the
// current LoadJournal (journal.go:86-99) gates on nothing — any
// schema_version decodes silently.
func TestJournalUnknownSchemaRefused(t *testing.T) {
	f := newFixture(t)
	raw := `{
  "schema_version": 99,
  "bundles_selection": ["extras"],
  "entries": [{"path": "claude-skills/moai-alpha/SKILL.md", "expected_sha256": "ab"}],
  "started_at": "2026-10-08T00:00:00Z"
}
`
	if err := os.MkdirAll(filepath.Join(f.home, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(JournalPath(f.home), []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}
	j, err := LoadJournal(JournalPath(f.home))
	if err == nil {
		t.Fatalf("RED (intended): LoadJournal silently decoded an unknown schema_version=99 journal (returned %+v) — no version gate exists", j)
	}
	if j != nil {
		t.Fatalf("LoadJournal returned a journal alongside its error: %+v", j)
	}
	if !strings.Contains(err.Error(), "schema") {
		t.Logf("note: refusal error does not name the schema field: %v", err)
	}
}
