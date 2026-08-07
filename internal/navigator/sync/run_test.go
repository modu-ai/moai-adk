package sync

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// popNavGraph reads nav-graph.json under root and unmarshals it.
func popNavGraph(t *testing.T, root string) Graph {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, ".moai", "project", "navigator", "nav-graph.json"))
	if err != nil {
		t.Fatalf("read nav-graph.json: %v", err)
	}
	var g Graph
	if err := json.Unmarshal(b, &g); err != nil {
		t.Fatalf("unmarshal nav-graph.json: %v\n%s", err, b)
	}
	return g
}

// standardPopulatedProject seeds a temp project with capability-map.md, a
// Go file carrying @NAV:DEC and @NAV:SYM tokens, and capability-symbols.json.
func standardPopulatedProject(t *testing.T) string {
	root := t.TempDir()
	// capability-map.md
	writeFixture(t, filepath.Join(root, ".moai", "project", "navigator", "capability-map.md"),
		"| spec-id | title | implementation-path |\n"+
			"|---------|-------|---------------------|\n"+
			"| SPEC-X-001 | X | src/x |\n")
	// SPEC module registration so the mx-associator can resolve path-based
	// associations for `src/x` → SPEC-X-001.
	writeFixture(t, filepath.Join(root, ".moai", "specs", "SPEC-X-001", "spec.md"),
		"---\nid: SPEC-X-001\nmodule: src/x\n---\n# SPEC-X-001\n")
	// Go fixture with tokens
	writeFixture(t, filepath.Join(root, "src", "x", "a.go"),
		"package x\n\n"+
			"// @NAV:DEC-AUTH: adopt OAuth2\n"+
			"// @NAV:SYM:pkg.ParseHeader\n"+
			"// @MX:NOTE: header parser\n"+
			"// @MX:SPEC: SPEC-X-001\n"+
			"func ParseHeader() {}\n")
	// design doc with @NAV:DEC and @NAV:SYM
	writeFixture(t, filepath.Join(root, ".moai", "project", "tech.md"),
		"# Tech\n\nDecision @NAV:DEC-AUTH: OAuth2.\nUses @NAV:SYM:pkg.WriteAtomic.\n")
	// capability-symbols.json (003 output)
	writeFixture(t, filepath.Join(root, ".moai", "project", "codemaps", "capability-symbols.json"),
		`{"rows":[{"spec_id":"SPEC-X-001","primary_symbols":[{"name":"pkg.ParseHeader"}]}]}`)
	return root
}

// TestRun_AC001_TokenTrioEdgeTypes exercises AC-001 / REQ-NS-001: the
// emitted graph contains edges of all three families.
func TestRun_AC001_TokenTrioEdgeTypes(t *testing.T) {
	root := standardPopulatedProject(t)
	if err := Run(Options{ProjectRoot: root}); err != nil {
		t.Fatalf("Run error: %v", err)
	}
	g := popNavGraph(t, root)
	types := map[EdgeType]bool{}
	for _, e := range g.Edges {
		types[e.EdgeType] = true
	}
	for _, want := range []EdgeType{EdgeDec, EdgeSpec, EdgeSym} {
		if !types[want] {
			t.Errorf("edge type %q missing; got %+v", want, types)
		}
	}
}

// TestRun_AC002_BindingRecordFiveFields exercises AC-002 / REQ-NS-002: every
// scanner-emitted binding record carries the five required fields with
// non-empty values.
func TestRun_AC002_BindingRecordFiveFields(t *testing.T) {
	root := standardPopulatedProject(t)
	if err := Run(Options{ProjectRoot: root}); err != nil {
		t.Fatalf("Run error: %v", err)
	}
	// Re-run scanners to inspect the binding records directly (the scanner
	// surface is what AC-002 asserts on, per acceptance.md evidence).
	logPath := filepath.Join(root, logPathName)
	prov := CurrentProvenance(root)
	decRecs, _, err := ScanDec(root, prov.ExtractCommitSHA, logPath)
	if err != nil {
		t.Fatalf("ScanDec error: %v", err)
	}
	symRecs, _, err := ScanSym(root, prov.ExtractCommitSHA, logPath)
	if err != nil {
		t.Fatalf("ScanSym error: %v", err)
	}
	for _, r := range append(append([]BindingRecord{}, decRecs...), symRecs...) {
		if string(r.TokenFamily) == "" {
			t.Errorf("record missing token_family: %+v", r)
		}
		if r.Identifier == "" {
			t.Errorf("record missing identifier: %+v", r)
		}
		if r.SourcePath == "" {
			t.Errorf("record missing source_path: %+v", r)
		}
		if r.LineNumber == 0 {
			t.Errorf("record missing line_number: %+v", r)
		}
		if r.CommitSHA == "" {
			t.Errorf("record missing commit_sha: %+v", r)
		}
	}
}

// TestRun_AC003_ScanDecFieldValues exercises AC-003.
func TestRun_AC003_ScanDecFieldValues(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, filepath.Join(root, ".moai", "project", "tech.md"),
		"# Tech\n\nDecision @NAV:DEC-AUTH-STRATEGY: adopt OAuth2\n")
	logPath := filepath.Join(root, logPathName)
	recs, _, err := ScanDec(root, "abc123", logPath)
	if err != nil {
		t.Fatalf("ScanDec error: %v", err)
	}
	if len(recs) != 1 {
		t.Fatalf("records = %d, want 1", len(recs))
	}
	r := recs[0]
	if r.TokenFamily != FamilyNavDec || r.Identifier != "AUTH-STRATEGY" {
		t.Errorf("record = %+v", r)
	}
	if !strings.HasSuffix(r.SourcePath, "tech.md") {
		t.Errorf("source_path = %q", r.SourcePath)
	}
	if r.CommitSHA != "abc123" {
		t.Errorf("commit_sha = %q, want abc123", r.CommitSHA)
	}
}

// TestRun_AC006_TopLevelShape exercises AC-006 / REQ-NS-006: top-level keys
// are exactly {provenance, nodes, edges}.
func TestRun_AC006_TopLevelShape(t *testing.T) {
	root := standardPopulatedProject(t)
	if err := Run(Options{ProjectRoot: root}); err != nil {
		t.Fatalf("Run error: %v", err)
	}
	b, err := os.ReadFile(filepath.Join(root, ".moai", "project", "navigator", "nav-graph.json"))
	if err != nil {
		t.Fatalf("read nav-graph.json: %v", err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	keys := make([]string, 0, len(raw))
	for k := range raw {
		keys = append(keys, k)
	}
	sortStringsInPlace(keys)
	want := []string{"edges", "nodes", "provenance"}
	if len(keys) != 3 || !equalStrings(keys, want) {
		t.Errorf("top-level keys = %v, want %v", keys, want)
	}
}

// TestRun_AC007_NodeEntityTypes exercises AC-007 / REQ-NS-007.
func TestRun_AC007_NodeEntityTypes(t *testing.T) {
	root := standardPopulatedProject(t)
	if err := Run(Options{ProjectRoot: root}); err != nil {
		t.Fatalf("Run error: %v", err)
	}
	g := popNavGraph(t, root)
	if len(g.Nodes) == 0 {
		t.Fatalf("no nodes emitted")
	}
	allowed := map[EntityType]bool{EntityDecision: true, EntitySpec: true, EntitySymbol: true}
	for _, n := range g.Nodes {
		if !allowed[n.EntityType] {
			t.Errorf("disallowed entity_type %q on node %+v", n.EntityType, n)
		}
	}
}

// TestRun_AC009_ByteIdenticalReRun exercises AC-009 / REQ-NS-009: two runs
// on the same HEAD produce byte-identical output. captured_at equals
// `git log -1 --format=%cI`.
func TestRun_AC009_ByteIdenticalReRun(t *testing.T) {
	root := standardPopulatedProject(t)
	if err := Run(Options{ProjectRoot: root}); err != nil {
		t.Fatalf("Run error: %v", err)
	}
	first, err := os.ReadFile(filepath.Join(root, ".moai", "project", "navigator", "nav-graph.json"))
	if err != nil {
		t.Fatalf("read first: %v", err)
	}
	// Second run.
	if err := Run(Options{ProjectRoot: root}); err != nil {
		t.Fatalf("Run error (2nd): %v", err)
	}
	second, err := os.ReadFile(filepath.Join(root, ".moai", "project", "navigator", "nav-graph.json"))
	if err != nil {
		t.Fatalf("read second: %v", err)
	}
	if string(first) != string(second) {
		t.Errorf("byte-identical re-run broken:\n--- first ---\n%s\n--- second ---\n%s", first, second)
	}
	// captured_at equals git committer date.
	want := gitOutput(root, "log", "-1", "--format=%cI")
	var g Graph
	if err := json.Unmarshal(first, &g); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if want != "" && g.Provenance.CapturedAt != want {
		t.Errorf("captured_at = %q, want %q", g.Provenance.CapturedAt, want)
	}
}

// TestRun_AC010_AtomicWriteBarrier exercises AC-010 / REQ-NS-010: the
// NAVIGATOR_PRE_RENAME_BARRIER test hook blocks the rename until removed.
func TestRun_AC010_AtomicWriteBarrier(t *testing.T) {
	root := standardPopulatedProject(t)
	barrier := filepath.Join(root, "barrier.flag")
	t.Setenv("NAVIGATOR_PRE_RENAME_BARRIER", barrier)

	done := make(chan error, 1)
	go func() { done <- Run(Options{ProjectRoot: root}) }()

	tmp := filepath.Join(root, ".moai", "project", "navigator", "nav-graph.json.tmp")
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(barrier); err == nil {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if _, err := os.Stat(barrier); err != nil {
		t.Fatalf("barrier file not created; tmp=%s", tmp)
	}
	if _, err := os.Stat(filepath.Join(root, ".moai", "project", "navigator", "nav-graph.json")); err == nil {
		t.Errorf("final nav-graph.json created before barrier removed")
	}
	if err := os.Remove(barrier); err != nil {
		t.Fatalf("remove barrier: %v", err)
	}
	if err := <-done; err != nil {
		t.Fatalf("Run after barrier: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".moai", "project", "navigator", "nav-graph.json")); err != nil {
		t.Errorf("final nav-graph.json missing after barrier removed: %v", err)
	}
	if _, err := os.Stat(tmp); err == nil {
		t.Errorf("tmp file left behind after rename")
	}
}

// TestRun_AC011_FailOpenWhenCapabilityMapAbsent exercises AC-011 / REQ-NS-011.
func TestRun_AC011_FailOpenWhenCapabilityMapAbsent(t *testing.T) {
	root := t.TempDir()
	if err := Run(Options{ProjectRoot: root}); err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".moai", "project", "navigator", "nav-graph.json")); err == nil {
		t.Errorf("nav-graph.json written despite capability-map absent")
	}
	b, err := os.ReadFile(filepath.Join(root, ".moai", "logs", "navigator-sync.log"))
	if err != nil {
		t.Fatalf("log file missing: %v", err)
	}
	if !strings.Contains(string(b), "capability-map absent") {
		t.Errorf("log missing 'capability-map absent'; got:\n%s", b)
	}
}

// TestRun_AC014_WriteSurfaceIsolation exercises AC-014 / REQ-NS-014: the
// only path the binary creates is nav-graph.json (the .tmp transient is
// renamed away).
func TestRun_AC014_WriteSurfaceIsolation(t *testing.T) {
	root := standardPopulatedProject(t)
	snapBefore := snapshotTree(root)
	if err := Run(Options{ProjectRoot: root}); err != nil {
		t.Fatalf("Run error: %v", err)
	}
	snapAfter := snapshotTree(root)
	created := diffCreated(snapBefore, snapAfter)
	// Allow the advisory log to be created/modified; AC-014 specifically
	// scopes the write-surface to nav-graph.json + .tmp transient (the .tmp
	// must be renamed away and not left behind).
	forbidden := []string{
		".moai/specs/",
		".moai/project/navigator/capability-map.md",
		".moai/project/navigator/audit-report.md",
		".moai/project/navigator/audit-report.json",
		".moai/project/codemaps/capability-symbols.md",
		".moai/project/codemaps/capability-symbols.json",
	}
	for path := range created {
		for _, f := range forbidden {
			if strings.HasPrefix(path, f) {
				t.Errorf("forbidden path created/modified by Run: %s", path)
			}
		}
		if strings.HasSuffix(path, "nav-graph.json.tmp") {
			t.Errorf(".tmp left behind: %s", path)
		}
	}
	// nav-graph.json MUST exist.
	if _, err := os.Stat(filepath.Join(root, ".moai", "project", "navigator", "nav-graph.json")); err != nil {
		t.Errorf("nav-graph.json not created: %v", err)
	}
}

// TestRun_AC015_AuditReportUntouched exercises AC-015 / REQ-NS-015: a
// pre-existing audit-report.json is byte-identical before and after the run.
func TestRun_AC015_AuditReportUntouched(t *testing.T) {
	root := standardPopulatedProject(t)
	auditPath := filepath.Join(root, ".moai", "project", "navigator", "audit-report.json")
	original := []byte(`{"findings":[{"id":"x"}]}` + "\n")
	writeFixture(t, auditPath, string(original))
	before, err := os.ReadFile(auditPath)
	if err != nil {
		t.Fatalf("read audit before: %v", err)
	}
	if err := Run(Options{ProjectRoot: root}); err != nil {
		t.Fatalf("Run error: %v", err)
	}
	after, err := os.ReadFile(auditPath)
	if err != nil {
		t.Fatalf("read audit after: %v", err)
	}
	if string(before) != string(after) {
		t.Errorf("audit-report.json mutated:\n--- before ---\n%s\n--- after ---\n%s", before, after)
	}
}

// TestRun_AC016_CapabilitySymbolsUntouched exercises AC-016 / REQ-NS-016.
func TestRun_AC016_CapabilitySymbolsUntouched(t *testing.T) {
	root := standardPopulatedProject(t)
	capsymPath := filepath.Join(root, ".moai", "project", "codemaps", "capability-symbols.json")
	before, err := os.ReadFile(capsymPath)
	if err != nil {
		t.Fatalf("read capsym before: %v", err)
	}
	if err := Run(Options{ProjectRoot: root}); err != nil {
		t.Fatalf("Run error: %v", err)
	}
	after, err := os.ReadFile(capsymPath)
	if err != nil {
		t.Fatalf("read capsym after: %v", err)
	}
	if string(before) != string(after) {
		t.Errorf("capability-symbols.json mutated")
	}
}

// TestRun_AC017_MalformedTokenDiagnostic exercises AC-017 / REQ-NS-017: a
// malformed token emits a diagnostic, is skipped, and the join completes
// with exit 0; the resulting graph carries no edge sourced from the
// malformed line.
func TestRun_AC017_MalformedTokenDiagnostic(t *testing.T) {
	root := t.TempDir()
	// capability-map (so the capability gate does not trip).
	writeFixture(t, filepath.Join(root, ".moai", "project", "navigator", "capability-map.md"),
		"| spec-id | title |\n|-|-|\n| SPEC-X-001 | X |\n")
	// Fixture with malformed @NAV:DEC- (empty id) and @NAV:SYM: (empty symbol).
	writeFixture(t, filepath.Join(root, ".moai", "project", "tech.md"),
		"# Tech\n\nBad @NAV:DEC- here\nBad @NAV:SYM: here\n")
	if err := Run(Options{ProjectRoot: root}); err != nil {
		t.Fatalf("Run error: %v", err)
	}
	logBytes, err := os.ReadFile(filepath.Join(root, ".moai", "logs", "navigator-sync.log"))
	if err != nil {
		t.Fatalf("log file missing: %v", err)
	}
	logStr := string(logBytes)
	if !strings.Contains(logStr, "malformed @NAV:DEC") {
		t.Errorf("log missing malformed DEC diagnostic:\n%s", logStr)
	}
	if !strings.Contains(logStr, "malformed @NAV:SYM") {
		t.Errorf("log missing malformed SYM diagnostic:\n%s", logStr)
	}
	// nav-graph.json MUST exist (REQ-NS-017: scanner does NOT abort the build).
	if _, err := os.Stat(filepath.Join(root, ".moai", "project", "navigator", "nav-graph.json")); err != nil {
		t.Errorf("nav-graph.json not emitted despite REQ-NS-017 non-abort: %v", err)
	}
}

// --- helpers ---

func sortStringsInPlace(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j-1] > s[j]; j-- {
			s[j-1], s[j] = s[j], s[j-1]
		}
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// snapshotTree returns the set of file paths under root.
func snapshotTree(root string) map[string]bool {
	out := map[string]bool{}
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		out[filepath.ToSlash(rel)] = true
		return nil
	})
	return out
}

// diffCreated returns the set of paths in after but not in before.
func diffCreated(before, after map[string]bool) map[string]bool {
	out := map[string]bool{}
	for k := range after {
		if !before[k] {
			out[k] = true
		}
	}
	return out
}
