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
// refreshed bytes' hash — never the stale recorded one. RED-now reason: the
// carry-in at :297 appends the OLD journal entry verbatim, so a run
// interrupted after the refresh mis-classifies its own refresh as a case-3
// mismatch (the entry's hash no longer matches the file it just wrote).
//
// Observation mechanics: the run must be stopped between the stage
// re-persist (:301) and the manifest save (:306) or the journal is cleared
// atomically with a successful save. The watcher polls the journal file for
// the carried entry's appearance (the :301 write), then makes ~/.moai
// read-only so the save fails and the journal survives on disk. The
// pre-seeded manifest carries filler entries purely to widen the save window
// (a big marshal takes milliseconds, giving the watcher a wide target).
func TestJournalRefreshUpdatesCarriedHash(t *testing.T) {
	const recoveredKey = "claude-skills/moai-beta/extra.md"
	const attempts = 40
	var lastObservation string
	caught := false

	var redEvidence string
	for attempt := 0; attempt < attempts && redEvidence == ""; attempt++ {
		observed, ok := attemptRefreshHashRepro(t, recoveredKey)
		lastObservation = observed
		if ok && strings.HasPrefix(observed, "captured") {
			redEvidence = observed
		}
		caught = caught || ok
	}
	if redEvidence != "" {
		t.Fatalf("RED (intended): %s", redEvidence)
	}
	if !caught {
		t.Fatalf("probe inconclusive after %d attempts (tool failure, not RED): the :301→:306 window was never caught; last observation: %s", attempts, lastObservation)
	}
}

// attemptRefreshHashRepro runs one attempt of the AC-003 repro. It returns
// (observation, true) when the carried entry was captured on disk after a
// forced save failure, and (observation, false) when the window was missed
// (the save succeeded and cleared the journal — retry).
func attemptRefreshHashRepro(t *testing.T, recoveredKey string) (string, bool) {
	f := newFixture(t)
	moaiHome := filepath.Join(f.home, ".moai")

	oldBytes := []byte("old recovered bytes the journal still hashes\n")
	newShipped := []byte("refreshed shipped bytes v2\n")
	// The recovered file exists on disk with the OLD bytes; the journal
	// hashes them (a case-2 claim) — the SOURCE carries the NEW shipped
	// bytes, so the re-evaluation arm refreshes the file mid-run.
	abs := filepath.Join(f.home, ".claude", "skills", "moai-beta", "extra.md")
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(abs, oldBytes, 0o644); err != nil {
		t.Fatal(err)
	}
	f.src[".claude/skills/moai-beta/extra.md"] = &fstest.MapFile{Data: newShipped}
	// extra.md needs a catalog entry for fileTarget to enumerate it in
	// reEvaluate? No — reEvaluate re-runs reconcile-claimed CATALOG targets.
	// The refresh arm is driven by applyTarget over reEvaluate, which is
	// built from installTargets (catalog entries). moai-beta is an
	// extras-pack skill; the entry covers the whole tree via dirTargets, so
	// extra.md rides the extras selection. The run therefore passes the
	// extras selection to reach it — and the recorded selection union (RF9)
	// keeps it honest.
	seedRecoveredJournal(t, f.home, recoveredKey, shaHex(oldBytes), true)

	// Widen the save window: filler entries make the manifest marshal slow.
	if err := widenManifestForSaveWindow(f.home, 4000); err != nil {
		t.Fatal(err)
	}

	// Watcher: the moment the journal file carries the recovered path AND is
	// no longer the seeded bytes — i.e. the :301 re-persist landed the
	// carried entry — make ~/.moai read-only so the manifest save at :306
	// fails and the journal survives. The seed marker discriminates the
	// pre-run journal (which contains the path from the start) from the
	// run's own carried write.
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
			if data, err := os.ReadFile(jp); err == nil &&
				strings.Contains(string(data), recoveredKey) &&
				!strings.Contains(string(data), seedJournalStartedAt) {
				_ = os.Chmod(moaiHome, 0o555)
				return
			}
			time.Sleep(100 * time.Microsecond)
		}
	}()

	inst := f.installer(t)
	_, installErr := inst.Install([]string{"extras"})
	close(stop)
	<-watcherDone
	// Restore writability no matter how the attempt ends.
	if err := os.Chmod(moaiHome, 0o755); err != nil {
		t.Fatalf("restore .moai perms: %v", err)
	}

	// The design point: the save must have FAILED (journal survives). A nil
	// error means the watcher lost the race — the save+clear succeeded, so
	// the window was missed and the attempt is retried.
	if installErr == nil {
		return "save succeeded — watcher lost the :301→:306 race", false
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
			if e.ExpectedSHA256 != shaHex(newShipped) {
				return "captured: the carried entry hash " + e.ExpectedSHA256 + " != the refreshed bytes hash " + shaHex(newShipped) + " — the :297 carry-in wrote the stale hash over a file the same run had already refreshed", true
			}
			return "captured: carried hash already correct (unexpected on this tree)", true
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
