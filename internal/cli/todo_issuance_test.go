package cli

import (
	"bytes"
	"context"
	"crypto/sha256"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/factory"
)

// M1 CLI acceptance tests (SPEC-TODO-CARD-ISSUANCE-001, AC-TCI-002/004/005/006).

// withIssuanceProbe swaps the production lane-branch probe for the test's.
func withIssuanceProbe(t *testing.T, p factory.LaneFilesProbe) {
	t.Helper()
	old := todoLaneFilesProbe
	todoLaneFilesProbe = p
	t.Cleanup(func() { todoLaneFilesProbe = old })
}

// withIssuanceSpecs swaps the completed-SPEC directory reader for fixed
// entries.
func withIssuanceSpecs(t *testing.T, specs []factory.IssuanceCompletedSpec) {
	t.Helper()
	old := todoCompletedSpecsReader
	todoCompletedSpecsReader = func(string, time.Time) []factory.IssuanceCompletedSpec {
		return specs
	}
	t.Cleanup(func() { todoCompletedSpecsReader = old })
}

// runTodoRaw is runTodo without any t.* call inside — safe to drive from a
// goroutine (the stall tests run add concurrently with lock probes).
func runTodoRaw(args ...string) (string, string, error) {
	cmd := newTodoCmd()
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), errBuf.String(), err
}

// queueBytes hashes every file under root/.moai — the byte-identity input
// for the dry-run contract.
func queueBytes(t *testing.T, root string) []byte {
	t.Helper()
	h := sha256.New()
	base := filepath.Join(root, ".moai")
	err := filepath.WalkDir(base, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		h.Write([]byte(path))
		h.Write(data)
		return nil
	})
	if err != nil {
		t.Fatalf("walk queue: %v", err)
	}
	return h.Sum(nil)
}

func lastSeqOf(t *testing.T, store *factory.BacklogStore) int {
	t.Helper()
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	return rec.LastSeq
}

// AC-TCI-002: the presentation goes to stderr only; stdout stays the bare
// machine line; live, dropped (stripped, reason) and archived neighbors all
// appear within the limit with score and measure.
func TestTodoAddPresentationStderrOnly(t *testing.T) {
	todoFixture(t)
	body := "align the CLI flag parser with the wizard defaults"
	if _, _, err := runTodo(t, "add", body+" quickly"); err != nil { // t1 live
		t.Fatal(err)
	}
	if _, _, err := runTodo(t, "add", body+" now"); err != nil { // t2 → dropped
		t.Fatal(err)
	}
	if _, _, err := runTodo(t, "drop", "t2", "operator decision"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := runTodo(t, "add", body+" for launch"); err != nil { // t3 → archived
		t.Fatal(err)
	}
	if _, _, err := runTodo(t, "done", "t3"); err != nil {
		t.Fatal(err)
	}
	out, errOut, err := runTodo(t, "add", body) // t4, similar to all three
	if err != nil {
		t.Fatal(err)
	}
	if out != "t4 2\n" {
		t.Errorf("stdout = %q, want exactly %q", out, "t4 2\n")
	}
	if strings.Count(errOut, "similar") > 3 {
		t.Errorf("neighbor lines exceed the limit:\n%s", errOut)
	}
	for _, id := range []string{"t1", "t2", "t3"} {
		if !strings.Contains(errOut, id) {
			t.Errorf("neighbor %s missing from stderr:\n%s", id, errOut)
		}
	}
	for _, want := range []string{"measure=token-set-jaccard", "live", "dropped", "archived"} {
		if !strings.Contains(errOut, want) {
			t.Errorf("stderr missing %q:\n%s", want, errOut)
		}
	}
	if strings.Contains(errOut, "[DROPPED") {
		t.Errorf("dropped prefix leaked:\n%s", errOut)
	}
	if !strings.Contains(errOut, "(reason: operator decision)") {
		t.Errorf("drop reason missing:\n%s", errOut)
	}
}

// AC-TCI-002 (e) + AC-TCI-003 (d)(f)(g): the in-flight overlap reports the
// shared path with lane and measure, says nothing for a measured none, and
// says unmeasured only when the candidate carries no input.
func TestTodoAddPresentationShowsInFlightOverlap(t *testing.T) {
	root, store := todoFixture(t)
	withIssuanceProbe(t, func(string, string) ([]string, bool) { return nil, false })
	if _, _, err := runTodo(t, "add", "refactor internal/cli/todo.go for issuance"); err != nil {
		t.Fatal(err)
	}
	if err := store.Mutate(func(rec *factory.BacklogRecord) error {
		rec.Items[0].State = factory.BacklogStatePicked
		stamp := "2026-10-05T12:00:00+09:00"
		rec.Items[0].PickedAt = &stamp
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := factory.RecordFactoryCardAssignment(root, "run-1", "t1", "lane-1", ""); err != nil {
		t.Fatal(err)
	}
	// Positive: the candidate shares internal/cli/todo.go with t1.
	_, errOut, err := runTodo(t, "add", "make internal/cli/todo.go issuance aware")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"t1", "lane-1", "internal/cli/todo.go", "measure=file-overlap"} {
		if !strings.Contains(errOut, want) {
			t.Errorf("overlap stderr missing %q:\n%s", want, errOut)
		}
	}
	// Measured none: the candidate carries input, nothing shared — no line.
	_, errOut, err = runTodo(t, "add", "touch internal/graph/graph.go only")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(errOut, "measure=file-overlap") || strings.Contains(errOut, "unmeasured") {
		t.Errorf("measured none must stay silent:\n%s", errOut)
	}
	// Unmeasured: the candidate carries no path at all.
	_, errOut, err = runTodo(t, "add", "no paths named in this body")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(errOut, "unmeasured (new card carries no expected files)") {
		t.Errorf("unmeasured line missing:\n%s", errOut)
	}
}

// AC-TCI-004 (c): the probe waits WITHOUT holding the queue lock — another
// process's Mutate completes while the presentation's probe is pending.
func TestTodoAddPresentationProbeOutsideLock(t *testing.T) {
	root, store := todoFixture(t)
	// The probe fires once per in-flight card: seed one picked card with a
	// lane assignment, or the presentation never probes and `started` never
	// closes.
	if _, _, err := runTodo(t, "add", "in-flight card naming internal/cli/todo.go"); err != nil {
		t.Fatal(err)
	}
	if err := store.Mutate(func(rec *factory.BacklogRecord) error {
		rec.Items[0].State = factory.BacklogStatePicked
		stamp := "2026-10-05T12:00:00+09:00"
		rec.Items[0].PickedAt = &stamp
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := factory.RecordFactoryCardAssignment(root, "run-1", "t1", "lane-1", ""); err != nil {
		t.Fatal(err)
	}
	started := make(chan struct{})
	release := make(chan struct{})
	var once int
	withIssuanceProbe(t, func(string, string) ([]string, bool) {
		if once == 0 {
			once++
			close(started)
		}
		<-release
		return nil, false
	})
	type result struct {
		out, errOut string
		err         error
	}
	done := make(chan result, 1)
	go func() {
		out, errOut, err := runTodoRaw("add", "card whose probe waits")
		done <- result{out, errOut, err}
	}()
	<-started
	mutated := make(chan error, 1)
	go func() {
		mutated <- store.Mutate(func(rec *factory.BacklogRecord) error { return nil })
	}()
	select {
	case err := <-mutated:
		if err != nil {
			t.Fatalf("concurrent mutate failed while the probe waited: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("concurrent Mutate blocked while the probe waited — the probe holds the lock")
	}
	close(release)
	r := <-done
	if r.err != nil {
		t.Fatalf("add: %v", r.err)
	}
	if !strings.Contains(r.out, "t2") {
		t.Errorf("stdout = %q", r.out)
	}
}

// AC-TCI-004 (d): a probe past the time bound degrades to
// `unmeasured (time bound)` and the admission still succeeds.
func TestTodoAddPresentationTimeBound(t *testing.T) {
	root, store := todoFixture(t)
	// Seed one in-flight card whose only input WOULD be the probe (its body
	// names no path), so the timed-out probe is the comparison's lost input.
	if _, _, err := runTodo(t, "add", "in-flight card without any path"); err != nil {
		t.Fatal(err)
	}
	if err := store.Mutate(func(rec *factory.BacklogRecord) error {
		rec.Items[0].State = factory.BacklogStatePicked
		stamp := "2026-10-05T12:00:00+09:00"
		rec.Items[0].PickedAt = &stamp
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := factory.RecordFactoryCardAssignment(root, "run-1", "t1", "lane-1", ""); err != nil {
		t.Fatal(err)
	}
	withIssuanceProbe(t, func(string, string) ([]string, bool) {
		time.Sleep(3 * time.Second) // > IssuanceProbeTimeBound (2s)
		return nil, false
	})
	start := time.Now()
	out, errOut, err := runTodoRaw("add", "touches internal/cli/todo.go while the probe stalls")
	if err != nil {
		t.Fatalf("add must succeed past the bound: %v", err)
	}
	if elapsed := time.Since(start); elapsed > 6*time.Second {
		t.Fatalf("add took %s — the bound did not fire", elapsed)
	}
	if !strings.Contains(errOut, "unmeasured (time bound)") {
		t.Errorf("time-bound line missing:\n%s", errOut)
	}
	if !strings.Contains(out, "t2") {
		t.Errorf("stdout = %q", out)
	}
}

// AC-TCI-004 (a): a failing presentation source leaves the admission
// identical to a presentation-free run.
func TestTodoAddPresentationNeverBlocks(t *testing.T) {
	root, _ := todoFixture(t)
	withIssuanceSpecs(t, nil)
	withIssuanceProbe(t, func(string, string) ([]string, bool) {
		time.Sleep(3 * time.Second)
		return nil, false
	})
	out, _, err := runTodoRaw("add", "card that survives a failing source")
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	if !strings.Contains(out, "t1 1") {
		t.Errorf("stdout = %q", out)
	}
	_ = root
}

// AC-TCI-005: --dry-run writes nothing, exits 0, and reports a would-be
// refusal on an exact duplicate instead of refusing.
func TestTodoAddPresentationDryRunWritesNothing(t *testing.T) {
	root, store := todoFixture(t)
	if _, _, err := runTodo(t, "add", "existing card about the exporter queue"); err != nil {
		t.Fatal(err)
	}
	before := queueBytes(t, root)
	seqBefore := lastSeqOf(t, store)

	out, errOut, err := runTodo(t, "add", "--dry-run", "another card about the exporter queue with retry")
	if err != nil {
		t.Fatalf("dry-run must exit 0: %v", err)
	}
	if out != "" {
		t.Errorf("dry-run stdout = %q, want empty", out)
	}
	if !strings.Contains(errOut, "dry-run: nothing was written") {
		t.Errorf("dry-run marker missing:\n%s", errOut)
	}
	if !strings.Contains(errOut, "measure=token-set-jaccard") {
		t.Errorf("dry-run should show below-floor neighbors too:\n%s", errOut)
	}
	if after := queueBytes(t, root); !bytes.Equal(before, after) {
		t.Errorf("queue bytes changed under --dry-run")
	}
	if seqAfter := lastSeqOf(t, store); seqAfter != seqBefore {
		t.Errorf("last_seq = %d, want %d (no id consumed)", seqAfter, seqBefore)
	}
	// Exact duplicate: reported, still exit 0.
	_, errOut, err = runTodo(t, "add", "--dry-run", "existing card about the exporter queue")
	if err != nil {
		t.Fatalf("dry-run over an exact duplicate must exit 0: %v", err)
	}
	if !strings.Contains(errOut, "a real add would refuse: t1 already holds this card") {
		t.Errorf("would-refuse line missing:\n%s", errOut)
	}
	// Regression: an unknown flag still refuses the way it did before.
	if _, _, err = runTodo(t, "add", "--bogus", "x"); err == nil || !strings.Contains(err.Error(), "unknown flag") {
		t.Errorf("unknown flag err = %v, want unknown flag refusal", err)
	}
}

// AC-TCI-005: --dry-run parses on the add surface; the natural-language
// fallthrough takes no flags — a leading --dry-run reaches the text verbatim.
func TestTodoAddDryRunFlagParsed(t *testing.T) {
	scan, err := scanTodoAddArgs([]string{"--dry-run", "some text"})
	if err != nil {
		t.Fatal(err)
	}
	if !scan.dryRun || scan.text != "some text" {
		t.Errorf("scan = %+v", scan)
	}
	_, store := todoFixture(t)
	// The fallthrough takes NO flags: a leading --dry-run refuses as an
	// unknown flag instead of being consumed as text (REQ-TCI-004's flag
	// belongs to the add surface alone).
	if _, _, err = runTodo(t, "--dry-run", "text"); err == nil || !strings.Contains(err.Error(), "unknown flag") {
		t.Fatalf("fallthrough err = %v, want unknown flag refusal", err)
	}
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	if len(rec.Items) != 0 {
		t.Errorf("fallthrough wrote a card: %+v", rec.Items)
	}
	_ = store
}

// AC-TCI-006 (a)(b): the MCP result keeps its first line, carries the
// presentation after one blank line, and stays byte-equal to the CLI stdout
// when the presentation is empty.
func TestTodoAddMCPCarriesPresentation(t *testing.T) {
	// Empty queue: no presentation, result == CLI stdout.
	rootA, _ := todoFixture(t)
	cliOut, _, err := runTodo(t, "add", "mcp parity card")
	if err != nil {
		t.Fatal(err)
	}
	rootB, _ := sdMoaiFixture(t)
	mcpOut, err := sdCallTool(t, handleTodoAdd, map[string]any{"text": "mcp parity card", "project_root": rootB})
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(mcpOut); got != strings.TrimSpace(cliOut) {
		t.Errorf("empty-queue mcp = %q, cli = %q", got, strings.TrimSpace(cliOut))
	}
	// Neighbours present: first line, blank line, presentation.
	mcpOut2, err := sdCallTool(t, handleTodoAdd, map[string]any{"text": "mcp parity card with the wizard defaults", "project_root": rootB})
	if err != nil {
		t.Fatal(err)
	}
	first := strings.SplitN(mcpOut2, "\n", 2)[0]
	if !strings.Contains(first, "t2") {
		t.Errorf("first line = %q, want the issued id line", first)
	}
	if !strings.Contains(mcpOut2, "\n\nissuance:") {
		t.Errorf("presentation block missing after a blank line:\n%s", mcpOut2)
	}
	_ = rootA
}

// AC-TCI-010: the drop stores the reason in the drop-reason attribute while
// the text keeps its prefix.
func TestTodoDropStoresReason(t *testing.T) {
	_, store := todoFixture(t)
	if _, _, err := runTodo(t, "add", "card that will be dropped"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := runTodo(t, "drop", "t1", "premise dead"); err != nil {
		t.Fatal(err)
	}
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	it := rec.Items[0]
	if it.State != factory.BacklogStateDropped {
		t.Fatalf("state = %s", it.State)
	}
	if it.Issuance == nil || it.Issuance.DropReason != "premise dead" {
		t.Errorf("drop reason attribute = %+v, want premise dead", it.Issuance)
	}
	if !strings.HasPrefix(it.Text, "[DROPPED — premise dead] ") {
		t.Errorf("text prefix missing: %q", it.Text)
	}
}

// AC-TCI-010: the prefix convention is unchanged — the stored text carries
// the marker and the reason verbatim.
func TestTodoDropKeepsTextPrefix(t *testing.T) {
	_, store := todoFixture(t)
	if _, _, err := runTodo(t, "add", "prefix keeper"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := runTodo(t, "drop", "t1", "superseded by t2"); err != nil {
		t.Fatal(err)
	}
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	want := "[DROPPED — superseded by t2] prefix keeper"
	if rec.Items[0].Text != want {
		t.Errorf("text = %q, want %q", rec.Items[0].Text, want)
	}
}

// AC-TCI-011: the relate verb's --disposition form records the disposition
// on the matching finding and records no new relation; an out-of-set value
// is refused.
func TestTodoRelateDispositionVerb(t *testing.T) {
	_, store := todoFixture(t)
	if _, _, err := runTodo(t, "add", "first card for the disposition pair"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := runTodo(t, "add", "second card for the disposition pair"); err != nil {
		t.Fatal(err)
	}
	if err := store.Mutate(func(rec *factory.BacklogRecord) error {
		rec.Findings = append(rec.Findings, factory.BacklogFinding{
			SubjectID: "t2", RelatedID: "t1", Relation: "near-duplicate",
			Source: "jev", Score: 0.85, Note: "pair", At: "2026-10-05T12:00:00+09:00",
		})
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	findingsBefore := len(rec.Findings)
	out, _, err := runTodo(t, "relate", "t2", "t1", "--disposition", "merge")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "disposition merge recorded") {
		t.Errorf("confirmation = %q", out)
	}
	rec, err = store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	if len(rec.Findings) != findingsBefore {
		t.Errorf("findings = %d, want %d (no new relation)", len(rec.Findings), findingsBefore)
	}
	matched := 0
	for i := range rec.Findings {
		f := &rec.Findings[i]
		if f.Names("t1") && f.Names("t2") {
			matched++
			if f.Disposition == nil || *f.Disposition != "merge" {
				t.Errorf("disposition = %+v, want merge", f.Disposition)
			}
		}
	}
	if matched != 1 {
		t.Errorf("matched = %d, want 1", matched)
	}
	// Out-of-set: refused, nothing written.
	_, _, err = runTodo(t, "relate", "t2", "t1", "--disposition", "maybe")
	if err == nil || !strings.Contains(err.Error(), "disposition must be one of") {
		t.Errorf("out-of-set err = %v, want the closed-set refusal", err)
	}
	rec, _ = store.LoadPure()
	for i := range rec.Findings {
		if rec.Findings[i].Names("t1") && rec.Findings[i].Names("t2") &&
			rec.Findings[i].Disposition != nil && *rec.Findings[i].Disposition == "maybe" {
			t.Errorf("out-of-set value was written")
		}
	}
}

// AC-TCI-013 (add clause): the issuance flags validate before the write —
// a repeated flag, an out-of-set origin, a nonexistent parent and a
// non-integer size all refuse with nothing written and no id consumed;
// the happy path attaches the attributes.
func TestTodoAddIssuanceFlagRefusals(t *testing.T) {
	root, store := todoFixture(t)
	if _, _, err := runTodo(t, "add", "parent card"); err != nil {
		t.Fatal(err)
	}
	seqBefore := lastSeqOf(t, store)

	cases := []struct {
		name, wantErr string
		args          []string
	}{
		{"repeated origin", "flag repeated: --origin", []string{"add", "--origin", "operator", "--origin", "leader", "x"}},
		{"repeated parent", "flag repeated: --parent", []string{"add", "--parent", "t1", "--parent", "t1", "x"}},
		{"out-of-set origin", "--origin must be one of", []string{"add", "--origin", "nobody-knows", "x"}},
		{"nonexistent parent", "--parent names no card", []string{"add", "--parent", "t999", "x"}},
		{"non-integer size", "--size-lines must be an integer", []string{"add", "--size-lines", "big", "x"}},
	}
	for _, tc := range cases {
		_, _, err := runTodo(t, tc.args...)
		if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
			t.Errorf("%s: err = %v, want %q", tc.name, err, tc.wantErr)
		}
	}
	if seqAfter := lastSeqOf(t, store); seqAfter != seqBefore {
		t.Errorf("last_seq moved %d → %d on refusals (ids consumed)", seqBefore, seqAfter)
	}
	// Happy path: archived parent counts as existing; the attributes attach.
	if _, _, err := runTodo(t, "done", "t1"); err != nil {
		t.Fatal(err)
	}
	out, _, err := runTodo(t, "add", "--parent", "t1", "--origin", "follow-up", "--size-lines", "120", "--files", "internal/cli/todo.go, internal/cli/todo_issuance.go", "follow-up card")
	if err != nil {
		t.Fatalf("flagged add: %v", err)
	}
	if !strings.HasPrefix(out, "t2") {
		t.Errorf("stdout = %q", out)
	}
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	iss := rec.Items[0].Issuance
	if iss == nil {
		t.Fatal("issuance not attached")
	}
	if iss.SpawnedBy != "t1" || iss.Origin != "follow-up" || iss.SizeLines == nil || *iss.SizeLines != 120 {
		t.Errorf("issuance = %+v", iss)
	}
	if len(iss.Files) != 2 || iss.Files[0] != "internal/cli/todo.go" {
		t.Errorf("files = %v", iss.Files)
	}
	_ = root
}

// AC-TCI-006 (c): engage prints the presentation on stderr, records no
// finding, and refuses nothing.
func TestGTDEngagePresentationRecordsNothing(t *testing.T) {
	root, store := todoFixture(t)
	ctx := context.Background()
	item, err := factory.CaptureGTDItem(ctx, store, factory.CaptureInput{Content: "align the CLI flag parser with the wizard defaults", Source: "user", SourceAllowed: true, Sensitivity: factory.SensitivityPrivate, EventID: "issuance-engage"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := factory.ClarifyGTDItem(ctx, store, factory.ClarifyInput{ItemID: item.ItemID, Disposition: factory.DispositionAction, DesiredOutcome: "done", CompletionEvidence: "assignment", Authority: "dispatch", SourceTrusted: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := factory.OrganizeGTDItem(ctx, store, factory.OrganizeInput{ItemID: item.ItemID, Class: factory.ClassAction}); err != nil {
		t.Fatal(err)
	}
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	findingsBefore := len(rec.Findings)

	cmd := newGTDEngageCmd()
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs([]string{item.ItemID, "--approve", "--fresh", "--dependencies-ready", "--resources", "--lane", "test-lane"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("engage: %v", err)
	}
	if !strings.Contains(errBuf.String(), "issuance:") {
		t.Errorf("engage stderr missing the presentation:\n%s", errBuf.String())
	}
	rec, err = store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	if len(rec.Findings) != findingsBefore {
		t.Errorf("findings = %d, want %d (engage records none)", len(rec.Findings), findingsBefore)
	}
	_ = root
}
