package constitution

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// SPEC-CON-AMEND-APPLY-001 — Execute-level acceptance tests for the apply
// step (AC-CAA-001 … 005, 014, 019, 020, 022 … 025). Every fixture lives
// under t.TempDir(); every test that reaches Execute sets
// MOAI_CONSTITUTION_REGISTRY and CLAUDE_PROJECT_DIR itself (REQ-CAA-015); the
// lock path is pinned under a separate t.TempDir(); one Pipeline serves one
// Execute call, because releaseLock clears LockFilePath.
//
// Rule IDs and clause texts avoid the digits 0, 1, and 2, so no digit run of
// an entry id or a clause can supply a count an assertion looks for
// (acceptance.md common rules, numeric-assertion rule).

const (
	fxRuleID   = "CONST-V3R6-789"
	fxBefore   = "Keep every amendment small and reviewable."
	fxAfter    = "Keep every amendment small, reviewable, and reversible."
	fxRuleFile = "rules/target.md"
)

// ruleSpec is one registry entry of a fixture.
type ruleSpec struct {
	id, zone, file, clause string
}

// standardEntries returns the three-entry fixture registry whose target entry
// points at targetFile.
func standardEntries(targetFile string) []ruleSpec {
	return []ruleSpec{
		{"CONST-V3R6-787", "Frozen", "rules/other.md", "Frozen text stays as written."},
		{fxRuleID, "Evolvable", targetFile, fxBefore},
		{"CONST-V3R6-798", "Evolvable", "rules/other.md", "Another evolvable clause stays."},
	}
}

// registryContent renders a registry with a header comment, a comment inside
// the fence, and a blank line between entries.
func registryContent(entries []ruleSpec) string {
	var b strings.Builder
	b.WriteString("# Zone Registry (fixture)\n\nHeader prose before the fence.\n\n```yaml\n")
	b.WriteString("# ============================================================\n")
	b.WriteString("# fixture entries\n")
	b.WriteString("# ============================================================\n")
	for i, e := range entries {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString("- id: " + e.id + "\n")
		b.WriteString("  zone: " + e.zone + "\n")
		b.WriteString("  file: " + e.file + "\n")
		b.WriteString("  anchor: \"#fixture\"\n")
		b.WriteString("  clause: \"" + e.clause + "\"\n")
		b.WriteString("  canary_gate: false\n")
	}
	b.WriteString("```\n\nTrailing prose after the fence.\n")
	return b.String()
}

const (
	fxRuleBody  = "# Target rules\n\nIntro paragraph.\n\n" + fxBefore + "\n\nTrailing paragraph.\n"
	fxOtherBody = "# Other rules\n\nFrozen text stays as written.\n\nAnother evolvable clause stays.\n"
	fxLogBody   = "# Evolution Log\n\n---\n" +
		"id: LEARN-20250505-005\n" +
		"rule_id: CONST-V3R6-798\n" +
		"zone_before: Evolvable\n" +
		"zone_after: Evolvable\n" +
		"clause_before: \"Old text.\"\n" +
		"clause_after: \"Another evolvable clause stays.\"\n" +
		"canary_verdict: skipped\n" +
		"approved_by: human\n" +
		"approved_at: 2025-05-05T05:05:05Z\n" +
		"rolled_back: false\n" +
		"---\n"
)

// project is one fixture project tree.
type project struct {
	dir, registry, rule, log string
}

// writeFixtureFile writes content to path, creating parent directories.
func writeFixtureFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// buildProject writes a fixture project under dir: the registry at the default
// location, rules/other.md, the target rule file (when targetFile is relative
// and ruleBody is non-empty), and the evolution log (when withLog).
func buildProject(t *testing.T, dir string, entries []ruleSpec, targetFile, ruleBody string, withLog bool) project {
	t.Helper()
	p := project{
		dir:      dir,
		registry: filepath.Join(dir, ".claude", "rules", "moai", "core", "zone-registry.md"),
		log:      filepath.Join(dir, ".moai", "research", "evolution-log.md"),
	}
	writeFixtureFile(t, p.registry, registryContent(entries))
	writeFixtureFile(t, filepath.Join(dir, "rules", "other.md"), fxOtherBody)
	if filepath.IsAbs(targetFile) {
		p.rule = targetFile
	} else {
		p.rule = filepath.Join(dir, targetFile)
	}
	if ruleBody != "" && !filepath.IsAbs(targetFile) {
		writeFixtureFile(t, p.rule, ruleBody)
	}
	if withLog {
		writeFixtureFile(t, p.log, fxLogBody)
	}
	return p
}

// standardProject is buildProject with the standard entries, rule body, and log.
func standardProject(t *testing.T, dir string) project {
	t.Helper()
	return buildProject(t, dir, standardEntries(fxRuleFile), fxRuleFile, fxRuleBody, true)
}

// isolateEnv sets both registry-resolution variables to empty (REQ-CAA-015).
func isolateEnv(t *testing.T) {
	t.Helper()
	t.Setenv("MOAI_CONSTITUTION_REGISTRY", "")
	t.Setenv("CLAUDE_PROJECT_DIR", "")
}

// recordingGates implements all five gate interfaces, approves everything, and
// counts calls.
type recordingGates struct{ calls int }

func (g *recordingGates) Check(*AmendmentProposal, Zone) error { g.calls++; return nil }
func (g *recordingGates) Evaluate(*AmendmentProposal, string) (*CanaryResult, error) {
	g.calls++
	return &CanaryResult{Available: true, Passed: true}, nil
}
func (g *recordingGates) Scan(*AmendmentProposal, *Registry) (*ContradictionResult, error) {
	g.calls++
	return &ContradictionResult{}, nil
}
func (g *recordingGates) Admit(*AmendmentProposal, string) error { g.calls++; return nil }
func (g *recordingGates) Approve(*AmendmentProposal, bool) (bool, error) {
	g.calls++
	return true, nil
}

// newGatedPipeline returns a pipeline whose five gates are one recording
// double, with the lock pinned under a fresh t.TempDir(). It returns the lock
// directory as well.
func newGatedPipeline(t *testing.T) (*Pipeline, *recordingGates, string) {
	t.Helper()
	g := &recordingGates{}
	lockDir := t.TempDir()
	p := &Pipeline{
		FrozenGuard:           g,
		Canary:                g,
		ContradictionDetector: g,
		RateLimiter:           g,
		HumanOversight:        g,
		LockFilePath:          filepath.Join(lockDir, "amend.lock"),
	}
	return p, g, lockDir
}

// proposal returns a proposal for the fixture target entry.
func proposal(before, after string) *AmendmentProposal {
	return &AmendmentProposal{RuleID: fxRuleID, Before: before, After: after}
}

// snapshotTree maps every path under root to its sha256 (files), "dir"
// (directories), or "link:<target>" (symbolic links, not followed).
func snapshotTree(t *testing.T, root string) map[string]string {
	t.Helper()
	m := map[string]string{}
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, p)
		switch {
		case d.Type()&fs.ModeSymlink != 0:
			target, lerr := os.Readlink(p)
			if lerr != nil {
				return lerr
			}
			m[rel] = "link:" + target
		case d.IsDir():
			m[rel] = "dir"
		default:
			data, rerr := os.ReadFile(p)
			if rerr != nil {
				return rerr
			}
			sum := sha256.Sum256(data)
			m[rel] = hex.EncodeToString(sum[:])
		}
		return nil
	})
	if err != nil {
		t.Fatalf("snapshot %s: %v", root, err)
	}
	return m
}

// assertSameTree fails when the snapshot of root differs from before.
func assertSameTree(t *testing.T, label, root string, before map[string]string) {
	t.Helper()
	after := snapshotTree(t, root)
	var diffs []string
	for k, v := range before {
		if av, ok := after[k]; !ok {
			diffs = append(diffs, "removed "+k)
		} else if av != v {
			diffs = append(diffs, "changed "+k)
		}
	}
	for k := range after {
		if _, ok := before[k]; !ok {
			diffs = append(diffs, "added "+k)
		}
	}
	sort.Strings(diffs)
	if len(diffs) > 0 {
		t.Errorf("%s: tree %s changed: %v", label, root, diffs)
	}
}

// fileSHA returns the sha256 of one file.
func fileSHA(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// readString reads one file.
func readString(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

// assertLockReleased fails when a lock file remains under lockDir.
func assertLockReleased(t *testing.T, lockDir string) {
	t.Helper()
	entries, err := os.ReadDir(lockDir)
	if err != nil {
		t.Fatalf("read lock dir: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("lock dir %s not empty after Execute: %d entries", lockDir, len(entries))
	}
}

// resolvedTempDir returns t.TempDir() with its symbolic links resolved, so the
// only links in a fixture are the ones a test creates on purpose.
func resolvedTempDir(t *testing.T) string {
	t.Helper()
	d, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("resolve temp dir: %v", err)
	}
	return d
}

// symlinkOrSkip creates a symbolic link or skips the test when the platform
// refuses (the skip is a Gap, never a PASS — acceptance.md AC-CAA-024).
func symlinkOrSkip(t *testing.T, target, link string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(link), err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("platform refused a symbolic link (%v) — recorded as a Gap", err)
	}
}

// requireNumbers asserts, on the error with every fixture path removed, that
// each of present occurs as a whole number and each of absent does not.
func requireNumbers(t *testing.T, err error, paths []string, present, absent []string) {
	t.Helper()
	stripped := stripFixturePaths(err.Error(), paths...)
	for _, n := range present {
		if !hasWholeNumber(stripped, n) {
			t.Errorf("error %q lacks the number %s (path-stripped: %q)", err, n, stripped)
		}
	}
	for _, n := range absent {
		if hasWholeNumber(stripped, n) {
			t.Errorf("error %q carries the number %s (path-stripped: %q)", err, n, stripped)
		}
	}
}

// AC-CAA-001 — exactly-once replacement lands in all three files.
func TestApply_ExactlyOnce_Success(t *testing.T) {
	isolateEnv(t)
	prj := standardProject(t, t.TempDir())
	ruleBefore := readString(t, prj.rule)
	logsBefore, err := LoadEvolutionLogs(prj.log)
	if err != nil {
		t.Fatal(err)
	}
	p, _, lockDir := newGatedPipeline(t)

	entry, err := p.Execute(proposal(fxBefore, fxAfter), prj.dir, false)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if entry == nil {
		t.Fatal("Execute returned a nil log entry")
	}
	if want := strings.Replace(ruleBefore, fxBefore, fxAfter, 1); readString(t, prj.rule) != want {
		t.Errorf("rule file:\n%s\nwant:\n%s", readString(t, prj.rule), want)
	}
	reg, err := LoadRegistry(prj.registry, prj.dir)
	if err != nil {
		t.Fatalf("LoadRegistry after apply: %v", err)
	}
	for _, e := range standardEntries(fxRuleFile) {
		got, _ := reg.Get(e.id)
		want := e.clause
		if e.id == fxRuleID {
			want = fxAfter
		}
		if got.Clause != want {
			t.Errorf("registry %s clause = %q, want %q", e.id, got.Clause, want)
		}
	}
	logsAfter, err := LoadEvolutionLogs(prj.log)
	if err != nil {
		t.Fatalf("LoadEvolutionLogs after apply: %v", err)
	}
	if len(logsAfter) != len(logsBefore)+1 {
		t.Fatalf("log entries = %d, want %d", len(logsAfter), len(logsBefore)+1)
	}
	last := logsAfter[len(logsAfter)-1]
	if last.RuleID != fxRuleID || last.ClauseBefore != fxBefore || last.ClauseAfter != fxAfter || last.ApprovedAt.IsZero() {
		t.Errorf("new log entry = %+v", last)
	}
	assertLockReleased(t, lockDir)
}

// AC-CAA-002 — zero or multiple occurrences fail before any write.
func TestApply_OccurrenceCount_Rejected(t *testing.T) {
	cases := []struct {
		name, body, count string
	}{
		{"zero", "# Target rules\n\nNo matching clause here.\n", "0"},
		{"two", "# Target rules\n\n" + fxBefore + "\n\nAgain: " + fxBefore + "\n", "2"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			isolateEnv(t)
			dir := t.TempDir()
			prj := buildProject(t, dir, standardEntries(fxRuleFile), fxRuleFile, tc.body, true)
			before := snapshotTree(t, dir)
			p, _, lockDir := newGatedPipeline(t)

			_, err := p.Execute(proposal(fxBefore, fxAfter), dir, false)
			if err == nil {
				t.Fatal("Execute: want an occurrence error, got nil")
			}
			if !containsPathForm(err.Error(), prj.rule) {
				t.Errorf("error %q does not name the rule file %s", err, prj.rule)
			}
			requireNumbers(t, err, []string{prj.rule, dir}, []string{tc.count}, nil)
			assertSameTree(t, tc.name, dir, before)
			assertLockReleased(t, lockDir)
		})
	}
}

// AC-CAA-003 — no whitespace normalization.
func TestApply_NoWhitespaceNormalization(t *testing.T) {
	isolateEnv(t)
	dir := t.TempDir()
	body := "# Target rules\n\n" +
		"Keep every  amendment small and reviewable.\n\n" +
		"Keep every amendment\nsmall and reviewable.\n"
	prj := buildProject(t, dir, standardEntries(fxRuleFile), fxRuleFile, body, true)
	before := snapshotTree(t, dir)
	p, _, _ := newGatedPipeline(t)

	_, err := p.Execute(proposal(fxBefore, fxAfter), dir, false)
	if err == nil {
		t.Fatal("Execute: want an occurrence error, got nil")
	}
	if !containsPathForm(err.Error(), prj.rule) {
		t.Errorf("error %q does not name the rule file %s", err, prj.rule)
	}
	requireNumbers(t, err, []string{prj.rule, dir}, []string{"0"}, []string{"2"})
	assertSameTree(t, "no-normalization", dir, before)
}

// AC-CAA-004 — the registry rewrite touches one line and round-trips.
func TestUpdateRegistryClause_SingleLineRoundTrip(t *testing.T) {
	isolateEnv(t)
	prj := standardProject(t, t.TempDir())
	newClause := `Quote "this", keep C:\temp: safe # not a comment`
	regBefore := readString(t, prj.registry)
	p, _, _ := newGatedPipeline(t)

	if _, err := p.Execute(proposal(fxBefore, newClause), prj.dir, false); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	beforeLines := strings.Split(regBefore, "\n")
	afterLines := strings.Split(readString(t, prj.registry), "\n")
	if len(afterLines) != len(beforeLines) {
		t.Fatalf("registry line count %d -> %d", len(beforeLines), len(afterLines))
	}
	var changed []int
	for i := range beforeLines {
		if beforeLines[i] != afterLines[i] {
			changed = append(changed, i)
		}
	}
	if len(changed) != 1 {
		t.Fatalf("registry lines changed = %d, want exactly one", len(changed))
	}
	if !strings.HasPrefix(strings.TrimSpace(beforeLines[changed[0]]), "clause: \""+fxBefore) {
		t.Errorf("changed line %q is not the target entry's clause line", beforeLines[changed[0]])
	}
	reg, err := LoadRegistry(prj.registry, prj.dir)
	if err != nil {
		t.Fatalf("LoadRegistry after apply: %v", err)
	}
	if got, _ := reg.Get(fxRuleID); got.Clause != newClause {
		t.Errorf("decoded clause = %q, want %q", got.Clause, newClause)
	}
}

// AC-CAA-005 — the re-parse verification blocks a corrupt registry.
func TestApply_RegistryReparse_Rejects(t *testing.T) {
	continuation := strings.Replace(registryContent(standardEntries(fxRuleFile)),
		"  clause: \""+fxBefore+"\"\n",
		"  clause: \"Keep every amendment small\n    and reviewable.\"\n", 1)
	flow := strings.Replace(registryContent(standardEntries(fxRuleFile)),
		"- id: "+fxRuleID+"\n  zone: Evolvable\n  file: "+fxRuleFile+"\n  anchor: \"#fixture\"\n  clause: \""+fxBefore+"\"\n  canary_gate: false\n",
		"- {id: "+fxRuleID+", zone: Evolvable, file: "+fxRuleFile+", anchor: \"#fixture\", clause: \""+fxBefore+"\", canary_gate: false}\n", 1)
	cases := []struct {
		name, registry, after string
	}{
		{"continuation", continuation, fxAfter},
		{"no_clause_line", flow, fxAfter},
		{"newline_clause", "", "Keep every amendment small,\nreviewable, and reversible."},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			isolateEnv(t)
			dir := t.TempDir()
			prj := standardProject(t, dir)
			if tc.registry != "" {
				writeFixtureFile(t, prj.registry, tc.registry)
			}
			if reg, err := LoadRegistry(prj.registry, dir); err != nil {
				t.Fatalf("fixture registry does not load: %v", err)
			} else if got, _ := reg.Get(fxRuleID); got.Clause != fxBefore {
				t.Fatalf("fixture target clause = %q, want %q", got.Clause, fxBefore)
			}
			before := snapshotTree(t, dir)
			p, _, _ := newGatedPipeline(t)

			_, err := p.Execute(proposal(fxBefore, tc.after), dir, false)
			if err == nil {
				t.Fatal("Execute: want a registry error, got nil")
			}
			if !containsPathForm(err.Error(), prj.registry) {
				t.Errorf("error %q does not name the registry %s", err, prj.registry)
			}
			assertSameTree(t, tc.name, dir, before)
		})
	}
}

// AC-CAA-019 — a new clause already present in the source is rejected.
func TestApply_NewClausePresent_Rejected(t *testing.T) {
	cases := []struct {
		name, body, after string
	}{
		{"elsewhere", fxRuleBody + "\nAlready here: " + fxAfter + "\n", fxAfter},
		{"inside_current_clause", fxRuleBody, "Keep every amendment small"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			isolateEnv(t)
			dir := t.TempDir()
			prj := buildProject(t, dir, standardEntries(fxRuleFile), fxRuleFile, tc.body, true)
			before := snapshotTree(t, dir)
			p, _, lockDir := newGatedPipeline(t)

			_, err := p.Execute(proposal(fxBefore, tc.after), dir, false)
			if err == nil {
				t.Fatal("Execute: want a new-clause error, got nil")
			}
			if !containsPathForm(err.Error(), prj.rule) {
				t.Errorf("error %q does not name the rule file %s", err, prj.rule)
			}
			requireNumbers(t, err, []string{prj.rule, dir}, []string{"1"}, nil)
			assertSameTree(t, tc.name, dir, before)
			assertLockReleased(t, lockDir)
		})
	}
}

// AC-CAA-014 — dry-run runs the validation and writes nothing.
func TestPipeline_Execute_DryRun_Validates(t *testing.T) {
	type fixture struct {
		name  string
		build func(t *testing.T, dir string) project
		prop  *AmendmentProposal
		want  func(prj project) (paths []string, numbers []string, texts []string)
	}
	fixtures := []fixture{
		{
			name:  "valid",
			build: standardProject,
			prop:  proposal(fxBefore, fxAfter),
		},
		{
			name: "two_occurrences",
			build: func(t *testing.T, dir string) project {
				return buildProject(t, dir, standardEntries(fxRuleFile), fxRuleFile, fxRuleBody+"\n"+fxBefore+"\n", true)
			},
			prop: proposal(fxBefore, fxAfter),
			want: func(prj project) ([]string, []string, []string) { return []string{prj.rule}, []string{"2"}, nil },
		},
		{
			name: "registry_continuation",
			build: func(t *testing.T, dir string) project {
				prj := standardProject(t, dir)
				writeFixtureFile(t, prj.registry, strings.Replace(registryContent(standardEntries(fxRuleFile)),
					"  clause: \""+fxBefore+"\"\n",
					"  clause: \"Keep every amendment small\n    and reviewable.\"\n", 1))
				return prj
			},
			prop: proposal(fxBefore, fxAfter),
			want: func(prj project) ([]string, []string, []string) { return []string{prj.registry}, nil, nil },
		},
		{
			name: "missing_rule_file",
			build: func(t *testing.T, dir string) project {
				return buildProject(t, dir, standardEntries(fxRuleFile), fxRuleFile, "", true)
			},
			prop: proposal(fxBefore, fxAfter),
			want: func(prj project) ([]string, []string, []string) { return []string{prj.rule}, nil, nil },
		},
		{
			name: "new_clause_present",
			build: func(t *testing.T, dir string) project {
				return buildProject(t, dir, standardEntries(fxRuleFile), fxRuleFile, fxRuleBody+"\n"+fxAfter+"\n", true)
			},
			prop: proposal(fxBefore, fxAfter),
			want: func(prj project) ([]string, []string, []string) { return []string{prj.rule}, []string{"1"}, nil },
		},
		{
			name:  "stale_before",
			build: standardProject,
			prop:  proposal(fxBefore+"!", fxAfter),
			want:  func(prj project) ([]string, []string, []string) { return nil, nil, []string{fxRuleID} },
		},
	}
	for _, fx := range fixtures {
		t.Run(fx.name, func(t *testing.T) {
			for _, dryRun := range []bool{true, false} {
				isolateEnv(t)
				dir := t.TempDir()
				prj := fx.build(t, dir)
				before := snapshotTree(t, dir)
				p, _, lockDir := newGatedPipeline(t)
				prop := *fx.prop

				entry, err := p.Execute(&prop, dir, dryRun)
				if fx.want == nil {
					if !dryRun {
						continue // the valid fixture's real run is AC-CAA-001's
					}
					if err != nil || entry == nil {
						t.Fatalf("dry-run on the valid fixture = (%v, %v), want a log entry and no error", entry, err)
					}
				} else {
					if err == nil {
						t.Fatalf("dryRun=%v: want an error, got nil", dryRun)
					}
					paths, numbers, texts := fx.want(prj)
					for _, path := range paths {
						if !containsPathForm(err.Error(), path) {
							t.Errorf("dryRun=%v: error %q does not name %s", dryRun, err, path)
						}
					}
					for _, text := range texts {
						if !strings.Contains(err.Error(), text) {
							t.Errorf("dryRun=%v: error %q lacks %q", dryRun, err, text)
						}
					}
					requireNumbers(t, err, append(paths, dir), numbers, nil)
				}
				if dryRun {
					assertSameTree(t, fx.name+" dry-run", dir, before)
					assertLockReleased(t, lockDir)
				}
			}
		})
	}
}

// AC-CAA-020 — Execute rejects a stale Before before Layer 1.
func TestPipeline_Execute_StaleBefore_Rejected(t *testing.T) {
	for _, mode := range []struct {
		name   string
		dryRun bool
	}{{"dry_run", true}, {"real", false}} {
		t.Run(mode.name, func(t *testing.T) {
			isolateEnv(t)
			dir := t.TempDir()
			standardProject(t, dir)
			before := snapshotTree(t, dir)
			p, gates, lockDir := newGatedPipeline(t)

			_, err := p.Execute(proposal(fxBefore+"!", fxAfter), dir, mode.dryRun)
			if err == nil {
				t.Fatal("Execute: want a stale-Before error, got nil")
			}
			if !strings.Contains(err.Error(), fxRuleID) {
				t.Errorf("error %q does not name the rule %s", err, fxRuleID)
			}
			if gates.calls != 0 {
				t.Errorf("gate doubles called %d times, want 0 (check runs before Layer 1)", gates.calls)
			}
			assertSameTree(t, mode.name, dir, before)
			assertLockReleased(t, lockDir)
		})
	}
}

// AC-CAA-022 — the CLI and Execute resolve the same registry.
func TestExecute_UsesSharedRegistryResolver(t *testing.T) {
	const clauseB = "Alternate registry clause text."
	const afterB = "Alternate registry clause text, amended."
	P := t.TempDir()
	Q := t.TempDir()
	prjP := standardProject(t, P)
	altRegistry := filepath.Join(P, "alt", "zone-registry.md")
	altEntries := []ruleSpec{
		{"CONST-V3R6-787", "Frozen", "rules/other.md", "Frozen text stays as written."},
		{fxRuleID, "Evolvable", "alt/rule.md", clauseB},
	}
	writeFixtureFile(t, altRegistry, registryContent(altEntries))
	writeFixtureFile(t, filepath.Join(P, "alt", "rule.md"), "# Alt rules\n\n"+clauseB+"\n")
	standardProject(t, Q)
	t.Setenv("MOAI_CONSTITUTION_REGISTRY", altRegistry)
	t.Setenv("CLAUDE_PROJECT_DIR", Q)
	if got := ResolveRegistryPath(P); got != altRegistry {
		t.Errorf("ResolveRegistryPath(P) = %q, want %q (the override outranks CLAUDE_PROJECT_DIR)", got, altRegistry)
	}
	defaultSHA := fileSHA(t, prjP.registry)
	beforeQ := snapshotTree(t, Q)
	p, _, _ := newGatedPipeline(t)

	if _, err := p.Execute(proposal(clauseB, afterB), P, false); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	reg, err := LoadRegistry(altRegistry, P)
	if err != nil {
		t.Fatalf("LoadRegistry(alt): %v", err)
	}
	if got, _ := reg.Get(fxRuleID); got.Clause != afterB {
		t.Errorf("alt registry clause = %q, want %q", got.Clause, afterB)
	}
	if fileSHA(t, prjP.registry) != defaultSHA {
		t.Error("the default-location registry in P changed")
	}
	assertSameTree(t, "Q", Q, beforeQ)
}

// AC-CAA-023 — a registry in another tree stops Execute before any write.
func TestExecute_RegistryOutsideProjectDir_Refused(t *testing.T) {
	build := func(t *testing.T) (P, Q string) {
		P, Q = t.TempDir(), t.TempDir()
		prj := standardProject(t, P)
		writeFixtureFile(t, filepath.Join(Q, ".claude", "rules", "moai", "core", "zone-registry.md"), readString(t, prj.registry))
		return P, Q
	}
	for _, mode := range []struct {
		name   string
		dryRun bool
	}{{"divergent_root_real", false}, {"divergent_root_dry_run", true}} {
		t.Run(mode.name, func(t *testing.T) {
			P, Q := build(t)
			t.Setenv("MOAI_CONSTITUTION_REGISTRY", "")
			t.Setenv("CLAUDE_PROJECT_DIR", Q)
			beforeP, beforeQ := snapshotTree(t, P), snapshotTree(t, Q)
			p, gates, lockDir := newGatedPipeline(t)

			_, err := p.Execute(proposal(fxBefore, fxAfter), P, mode.dryRun)
			if err == nil {
				t.Fatal("Execute: want a registry load error, got nil")
			}
			offending := filepath.Join(Q, ".claude", "rules", "moai", "core", "zone-registry.md")
			if !containsPathForm(err.Error(), offending) {
				t.Errorf("error %q does not name the offending registry %s", err, offending)
			}
			if gates.calls != 0 {
				t.Errorf("gate doubles called %d times, want 0", gates.calls)
			}
			assertSameTree(t, "P", P, beforeP)
			assertSameTree(t, "Q", Q, beforeQ)
			assertLockReleased(t, lockDir)
		})
	}
	t.Run("same_root_control", func(t *testing.T) {
		P, Q := build(t)
		t.Setenv("MOAI_CONSTITUTION_REGISTRY", "")
		t.Setenv("CLAUDE_PROJECT_DIR", P)
		beforeQ := snapshotTree(t, Q)
		p, _, _ := newGatedPipeline(t)

		if _, err := p.Execute(proposal(fxBefore, fxAfter), P, false); err != nil {
			t.Fatalf("Execute: %v", err)
		}
		if !strings.Contains(readString(t, filepath.Join(P, fxRuleFile)), fxAfter) {
			t.Error("P's rule file does not carry the amendment")
		}
		reg, err := LoadRegistry(filepath.Join(P, ".claude", "rules", "moai", "core", "zone-registry.md"), P)
		if err != nil {
			t.Fatal(err)
		}
		if got, _ := reg.Get(fxRuleID); got.Clause != fxAfter {
			t.Errorf("P's registry clause = %q, want %q", got.Clause, fxAfter)
		}
		if logs, err := LoadEvolutionLogs(filepath.Join(P, ".moai", "research", "evolution-log.md")); err != nil || len(logs) != 2 {
			t.Errorf("P's log = %d entries (%v), want 2", len(logs), err)
		}
		assertSameTree(t, "Q", Q, beforeQ)
	})
}

// escapeFixture is one AC-CAA-024 fixture: base B (resolved), root P = B/root.
type escapeFixture struct {
	B, P      string
	offending string
}

// newEscapeBase creates B and the trees AC-CAA-024 names.
func newEscapeBase(t *testing.T) (B, P string) {
	t.Helper()
	B = resolvedTempDir(t)
	P = filepath.Join(B, "root")
	for _, d := range []string{P, filepath.Join(B, "root-evil"), filepath.Join(B, "other"), filepath.Join(B, "outside"), filepath.Join(B, "outside-research")} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return B, P
}

// buildEscapeFixture builds row/caseName of AC-CAA-024 and sets the
// environment and working directory the row names.
func buildEscapeFixture(t *testing.T, row, caseName string) escapeFixture {
	t.Helper()
	B, P := newEscapeBase(t)
	fx := escapeFixture{B: B, P: P}
	t.Setenv("MOAI_CONSTITUTION_REGISTRY", "")
	t.Setenv("CLAUDE_PROJECT_DIR", "")
	outsideRule := filepath.Join(B, "outside", "rule.md")
	defaultRel := filepath.Join(".claude", "rules", "moai", "core", "zone-registry.md")
	switch row {
	case "relative_env_escape":
		prj := standardProject(t, P)
		writeFixtureFile(t, filepath.Join(B, "other", defaultRel), readString(t, prj.registry))
		t.Chdir(P)
		if caseName == "registry_var" {
			t.Setenv("MOAI_CONSTITUTION_REGISTRY", filepath.Join("..", "other", defaultRel))
		} else {
			t.Setenv("CLAUDE_PROJECT_DIR", filepath.Join("..", "other"))
		}
		fx.offending = filepath.Join(B, "other", defaultRel)
	case "absolute_file":
		if caseName == "non_target" {
			entries := standardEntries(fxRuleFile)
			entries[2].file = outsideRule
			buildProject(t, P, entries, fxRuleFile, fxRuleBody, true)
		} else {
			buildProject(t, P, standardEntries(outsideRule), outsideRule, "", true)
		}
		writeFixtureFile(t, outsideRule, fxRuleBody)
		fx.offending = outsideRule
	case "dotdot_file":
		rel := filepath.Join("..", "outside", "rule.md")
		buildProject(t, P, standardEntries(rel), fxRuleFile, "", true)
		writeFixtureFile(t, outsideRule, fxRuleBody)
		fx.offending = outsideRule
	case "sibling_prefix_file":
		evil := filepath.Join(B, "root-evil", "rule.md")
		buildProject(t, P, standardEntries(evil), evil, "", true)
		writeFixtureFile(t, evil, fxRuleBody)
		fx.offending = evil
	case "symlinked_registry":
		prj := standardProject(t, P)
		writeFixtureFile(t, filepath.Join(B, "other", "zone-registry.md"), readString(t, prj.registry))
		symlinkOrSkip(t, filepath.Join(B, "other"), filepath.Join(P, "linkdir"))
		t.Setenv("MOAI_CONSTITUTION_REGISTRY", filepath.Join(P, "linkdir", "zone-registry.md"))
		fx.offending = filepath.Join(P, "linkdir", "zone-registry.md")
	case "symlinked_file":
		rel := filepath.Join("linked", "rule.md")
		buildProject(t, P, standardEntries(rel), fxRuleFile, "", true)
		writeFixtureFile(t, outsideRule, fxRuleBody)
		symlinkOrSkip(t, filepath.Join(B, "outside"), filepath.Join(P, "linked"))
		fx.offending = filepath.Join(P, "linked", "rule.md")
	case "symlinked_log":
		buildProject(t, P, standardEntries(fxRuleFile), fxRuleFile, fxRuleBody, false)
		writeFixtureFile(t, filepath.Join(B, "outside-research", "evolution-log.md"), fxLogBody)
		symlinkOrSkip(t, filepath.Join(B, "outside-research"), filepath.Join(P, ".moai", "research"))
		fx.offending = filepath.Join(P, ".moai", "research", "evolution-log.md")
	default:
		t.Fatalf("unknown row %s", row)
	}
	return fx
}

// AC-CAA-024 — the containment check refuses every escaping path shape before
// any write; in-root controls pass; the loader keeps its present behaviour.
func TestExecute_ContainmentCheck_RefusesEscapes(t *testing.T) {
	rows := []struct {
		row   string
		cases []string
	}{
		{"relative_env_escape", []string{"registry_var", "project_dir_var"}},
		{"absolute_file", []string{"target", "non_target"}},
		{"dotdot_file", []string{""}},
		{"sibling_prefix_file", []string{""}},
		{"symlinked_registry", []string{""}},
		{"symlinked_file", []string{""}},
		{"symlinked_log", []string{""}},
	}
	for _, r := range rows {
		t.Run(r.row, func(t *testing.T) {
			for _, c := range r.cases {
				for _, mode := range []struct {
					name   string
					dryRun bool
				}{{"real", false}, {"dry_run", true}} {
					name := mode.name
					if c != "" {
						name = c + "/" + mode.name
					}
					t.Run(name, func(t *testing.T) {
						fx := buildEscapeFixture(t, r.row, c)
						before := snapshotTree(t, fx.B)
						p, gates, lockDir := newGatedPipeline(t)

						_, err := p.Execute(proposal(fxBefore, fxAfter), fx.P, mode.dryRun)
						if err == nil {
							t.Fatal("Execute: want a containment refusal, got nil")
						}
						if !containsPathForm(err.Error(), fx.offending) {
							t.Errorf("error %q does not name the offending path %s", err, fx.offending)
						}
						if gates.calls != 0 {
							t.Errorf("gate doubles called %d times, want 0", gates.calls)
						}
						assertSameTree(t, "B", fx.B, before)
						assertLockReleased(t, lockDir)
					})
				}
			}
		})
	}

	t.Run("in_root_control", func(t *testing.T) {
		for _, c := range []string{"plain", "symlinked_root"} {
			for _, mode := range []struct {
				name   string
				dryRun bool
			}{{"real", false}, {"dry_run", true}} {
				t.Run(c+"/"+mode.name, func(t *testing.T) {
					B, P := newEscapeBase(t)
					prj := buildProject(t, P, standardEntries(fxRuleFile), fxRuleFile, fxRuleBody, false)
					if err := os.MkdirAll(filepath.Join(P, ".moai", "research"), 0o755); err != nil {
						t.Fatal(err)
					}
					projectDir := P
					if c == "symlinked_root" {
						symlinkOrSkip(t, P, filepath.Join(B, "link"))
						projectDir = filepath.Join(B, "link")
					}
					t.Chdir(projectDir)
					t.Setenv("MOAI_CONSTITUTION_REGISTRY", filepath.Join(".claude", "rules", "moai", "core", "zone-registry.md"))
					t.Setenv("CLAUDE_PROJECT_DIR", "")
					before := snapshotTree(t, B)
					p, _, _ := newGatedPipeline(t)

					entry, err := p.Execute(proposal(fxBefore, fxAfter), projectDir, mode.dryRun)
					if err != nil || entry == nil {
						t.Fatalf("Execute = (%v, %v), want a log entry and no error", entry, err)
					}
					if mode.dryRun {
						assertSameTree(t, "B", B, before)
						return
					}
					if !strings.Contains(readString(t, prj.rule), fxAfter) {
						t.Error("rule file does not carry the amendment")
					}
					reg, err := LoadRegistry(prj.registry, P)
					if err != nil {
						t.Fatal(err)
					}
					if got, _ := reg.Get(fxRuleID); got.Clause != fxAfter {
						t.Errorf("registry clause = %q, want %q", got.Clause, fxAfter)
					}
					logs, err := LoadEvolutionLogs(prj.log)
					if err != nil || len(logs) != 1 {
						t.Errorf("new log = %d entries (%v), want exactly one", len(logs), err)
					}
					after := snapshotTree(t, B)
					for k, v := range before {
						if strings.HasPrefix(k, "root") && (k == "root" || strings.HasPrefix(k, "root"+string(filepath.Separator))) {
							continue
						}
						if after[k] != v {
							t.Errorf("path %s outside P changed", k)
						}
					}
				})
			}
		}
	})

	t.Run("loader_unchanged", func(t *testing.T) {
		countIDs := func(t *testing.T, registry string) int {
			fence, err := extractYAMLFence(readString(t, registry))
			if err != nil {
				t.Fatal(err)
			}
			return len(regexp.MustCompile(`(?m)^- id:`).FindAllString(fence, -1))
		}
		t.Run("relative_registry", func(t *testing.T) {
			fx := buildEscapeFixture(t, "relative_env_escape", "registry_var")
			before := snapshotTree(t, fx.B)
			reg, err := LoadRegistry(filepath.Join("..", "other", ".claude", "rules", "moai", "core", "zone-registry.md"), fx.P)
			if err != nil {
				t.Fatalf("LoadRegistry: %v", err)
			}
			if want := countIDs(t, fx.offending); len(reg.Entries) != want {
				t.Errorf("entries = %d, want %d", len(reg.Entries), want)
			}
			assertSameTree(t, "B", fx.B, before)
		})
		t.Run("absolute_file", func(t *testing.T) {
			fx := buildEscapeFixture(t, "absolute_file", "target")
			registry := filepath.Join(fx.P, ".claude", "rules", "moai", "core", "zone-registry.md")
			before := snapshotTree(t, fx.B)
			reg, err := LoadRegistry(registry, fx.P)
			if err != nil {
				t.Fatalf("LoadRegistry: %v", err)
			}
			if want := countIDs(t, registry); len(reg.Entries) != want {
				t.Errorf("entries = %d, want %d", len(reg.Entries), want)
			}
			if got, _ := reg.Get(fxRuleID); got.File != filepath.Join(fx.B, "outside", "rule.md") {
				t.Errorf("target File = %q, want %q", got.File, filepath.Join(fx.B, "outside", "rule.md"))
			}
			assertSameTree(t, "B", fx.B, before)
		})
		t.Run("symlinked_registry", func(t *testing.T) {
			fx := buildEscapeFixture(t, "symlinked_registry", "")
			before := snapshotTree(t, fx.B)
			reg, err := LoadRegistry(fx.offending, fx.P)
			if err != nil {
				t.Fatalf("LoadRegistry: %v", err)
			}
			if want := countIDs(t, fx.offending); len(reg.Entries) != want {
				t.Errorf("entries = %d, want %d", len(reg.Entries), want)
			}
			assertSameTree(t, "B", fx.B, before)
		})
		t.Run("absolute_escape", func(t *testing.T) {
			fx := buildEscapeFixture(t, "relative_env_escape", "registry_var")
			before := snapshotTree(t, fx.B)
			_, err := LoadRegistry(fx.offending, fx.P)
			if err == nil || !strings.Contains(err.Error(), "escapes project dir") {
				t.Errorf("LoadRegistry(absolute escape) = %v, want the loader's 'escapes project dir' error", err)
			}
			assertSameTree(t, "B", fx.B, before)
		})
	})
}

// realRegistryFixture is one AC-CAA-025 fixture built on a byte copy of the
// repository's real registry.
type realRegistryFixture struct {
	B, P, link, registry string
	target               rawEntry
	targetFile           string
}

// buildRealRegistryFixture copies the real registry (read-only) into P, creates
// every distinct file: path under P, and picks the target entry by rule.
func buildRealRegistryFixture(t *testing.T) realRegistryFixture {
	t.Helper()
	realRegistry := filepath.Join("..", "..", ".claude", "rules", "moai", "core", "zone-registry.md")
	data, err := os.ReadFile(realRegistry)
	if err != nil {
		t.Fatalf("read real registry: %v", err)
	}
	B := resolvedTempDir(t)
	fx := realRegistryFixture{B: B, P: filepath.Join(B, "root"), link: filepath.Join(B, "link")}
	fx.registry = filepath.Join(fx.P, ".claude", "rules", "moai", "core", "zone-registry.md")
	writeFixtureFile(t, fx.registry, string(data))
	if err := os.MkdirAll(filepath.Join(fx.P, ".moai", "research"), 0o755); err != nil {
		t.Fatal(err)
	}
	fence, err := extractYAMLFence(string(data))
	if err != nil {
		t.Fatal(err)
	}
	var raw []rawEntry
	if err := yaml.Unmarshal([]byte(fence), &raw); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, e := range raw {
		if strings.EqualFold(e.Zone, "Evolvable") && !strings.HasPrefix(e.Clause, "[SUPERSEDED") {
			fx.target, found = e, true
			break
		}
	}
	if !found {
		t.Fatal("real registry has no live Evolvable entry")
	}
	fx.targetFile = filepath.Join(fx.P, fx.target.File)
	placeholder := "# placeholder\n\nPlaceholder text for the fixture.\n"
	seen := map[string]bool{}
	for _, e := range raw {
		if seen[e.File] {
			continue
		}
		seen[e.File] = true
		body := placeholder
		if e.File == fx.target.File {
			body += "\n" + fx.target.Clause + "\n"
		}
		writeFixtureFile(t, filepath.Join(fx.P, e.File), body)
	}
	symlinkOrSkip(t, fx.P, fx.link)
	return fx
}

// AC-CAA-025 — the real registry's file: shape is still admitted.
func TestExecute_RealRegistryShape_Admitted(t *testing.T) {
	const after = "Amended fixture clause for the real registry shape."
	realRegistry := filepath.Join("..", "..", ".claude", "rules", "moai", "core", "zone-registry.md")
	realLog := filepath.Join("..", "..", ".moai", "research", "evolution-log.md")
	realRegistrySHA := fileSHA(t, realRegistry)
	realLogSHA := fileSHA(t, realLog)
	t.Cleanup(func() {
		if fileSHA(t, realRegistry) != realRegistrySHA || fileSHA(t, realLog) != realLogSHA {
			t.Error("the real registry or the real log changed (REQ-CAA-015)")
		}
	})

	// Drift witness on the copy's file: shape.
	fence, err := extractYAMLFence(readString(t, realRegistry))
	if err != nil {
		t.Fatal(err)
	}
	fileLine := regexp.MustCompile(`(?m)^\s+file:\s*(.*)$`)
	var absolute, dotdot int
	for _, m := range fileLine.FindAllStringSubmatch(fence, -1) {
		v := strings.Trim(strings.TrimSpace(m[1]), `"'`)
		if filepath.IsAbs(v) || strings.HasPrefix(v, "/") {
			absolute++
		}
		if strings.Contains(v, "..") {
			dotdot++
		}
	}
	if absolute != 0 || dotdot != 0 {
		t.Fatalf("real registry file: shape drifted: %d absolute, %d containing '..' (review the registry, do not relax the witness)", absolute, dotdot)
	}

	t.Run("load", func(t *testing.T) {
		isolateEnv(t)
		fx := buildRealRegistryFixture(t)
		reg, err := LoadRegistry(filepath.Join(fx.link, ".claude", "rules", "moai", "core", "zone-registry.md"), fx.link)
		if err != nil {
			t.Fatalf("LoadRegistry: %v", err)
		}
		f, _ := extractYAMLFence(readString(t, fx.registry))
		if want := len(regexp.MustCompile(`(?m)^- id:`).FindAllString(f, -1)); len(reg.Entries) != want {
			t.Errorf("entries = %d, want %d", len(reg.Entries), want)
		}
	})

	t.Run("dry_run", func(t *testing.T) {
		isolateEnv(t)
		fx := buildRealRegistryFixture(t)
		before := snapshotTree(t, fx.B)
		p, _, _ := newGatedPipeline(t)
		entry, err := p.Execute(&AmendmentProposal{RuleID: fx.target.ID, Before: fx.target.Clause, After: after}, fx.link, true)
		if err != nil || entry == nil {
			t.Fatalf("Execute dry-run = (%v, %v), want a log entry and no error", entry, err)
		}
		assertSameTree(t, "B", fx.B, before)
	})

	t.Run("real", func(t *testing.T) {
		isolateEnv(t)
		fx := buildRealRegistryFixture(t)
		before := snapshotTree(t, fx.B)
		p, _, _ := newGatedPipeline(t)
		entry, err := p.Execute(&AmendmentProposal{RuleID: fx.target.ID, Before: fx.target.Clause, After: after}, fx.link, false)
		if err != nil || entry == nil {
			t.Fatalf("Execute = (%v, %v), want a log entry and no error", entry, err)
		}
		if !strings.Contains(readString(t, fx.targetFile), after) {
			t.Error("the target rule file does not carry the amendment")
		}
		reg, err := LoadRegistry(fx.registry, fx.P)
		if err != nil {
			t.Fatal(err)
		}
		if got, _ := reg.Get(fx.target.ID); got.Clause != after {
			t.Errorf("registry clause = %q, want %q", got.Clause, after)
		}
		logs, err := LoadEvolutionLogs(filepath.Join(fx.P, ".moai", "research", "evolution-log.md"))
		if err != nil || len(logs) != 1 || logs[0].RuleID != fx.target.ID {
			t.Errorf("new log = %+v (%v), want one entry for %s", logs, err, fx.target.ID)
		}
		afterSnap := snapshotTree(t, fx.B)
		changed := map[string]bool{}
		for _, p := range []string{fx.targetFile, fx.registry, filepath.Join(fx.P, ".moai", "research", "evolution-log.md")} {
			rel, _ := filepath.Rel(fx.B, p)
			changed[rel] = true
		}
		for k, v := range before {
			if !changed[k] && afterSnap[k] != v {
				t.Errorf("unrelated path %s changed", k)
			}
		}
		for k := range afterSnap {
			if _, ok := before[k]; !ok && !changed[k] {
				t.Errorf("unexpected new path %s", k)
			}
		}
	})
}

// Out-of-SPEC guards (card t659, accepted by the lead as trust-boundary input
// validation): each is pinned by one test and one mutant. They carry no AC id.

// Execute rejects a proposal whose After is empty before any gate runs, in
// both modes.
func TestExecute_EmptyAfter_Rejected(t *testing.T) {
	const wantMsg = "After is empty"
	for _, mode := range []struct {
		name   string
		dryRun bool
	}{{"dry_run", true}, {"real", false}} {
		t.Run(mode.name, func(t *testing.T) {
			isolateEnv(t)
			dir := t.TempDir()
			standardProject(t, dir)
			before := snapshotTree(t, dir)
			p, gates, lockDir := newGatedPipeline(t)

			_, err := p.Execute(proposal(fxBefore, ""), dir, mode.dryRun)
			if err == nil {
				t.Fatal("Execute: want an empty-After error, got nil")
			}
			if !strings.Contains(err.Error(), wantMsg) {
				t.Errorf("error %q lacks %q", err, wantMsg)
			}
			if !strings.Contains(err.Error(), fxRuleID) {
				t.Errorf("error %q does not name the rule %s", err, fxRuleID)
			}
			if gates.calls != 0 {
				t.Errorf("gate doubles called %d times, want 0 (the check runs before Layer 1)", gates.calls)
			}
			assertSameTree(t, mode.name, dir, before)
			assertLockReleased(t, lockDir)
		})
	}
}

// Execute rejects a target entry whose file: names the registry or the
// evolution log, in both modes, before any gate runs: a rejection that came
// after Layer 5 would waste the user's approval, so the gate doubles must not
// be called at all. Each half of the condition has its own case; the
// control case, a distinct rule file, is not rejected. The registry carries
// the current clause once by construction; the evolution-log case puts it
// into the log's prose once, so without the check the later steps would
// accept that file as a rule file too.
func TestExecute_RuleFileIsRegistryOrLog_Rejected(t *testing.T) {
	const wantMsg = "is also the registry or the evolution log"
	logWithClause := strings.Replace(fxLogBody, "# Evolution Log\n\n", "# Evolution Log\n\n"+fxBefore+"\n\n", 1)
	cases := []struct {
		name, targetFile, ruleBody, logBody string
		wantReject                          bool
	}{
		{"registry", RegistryRelPath, "", "", true},
		{"evolution_log", ".moai/research/evolution-log.md", "", logWithClause, true},
		{"distinct_control", fxRuleFile, fxRuleBody, "", false},
	}
	for _, tc := range cases {
		for _, mode := range []struct {
			name   string
			dryRun bool
		}{{"dry_run", true}, {"real", false}} {
			t.Run(tc.name+"/"+mode.name, func(t *testing.T) {
				isolateEnv(t)
				dir := t.TempDir()
				prj := buildProject(t, dir, standardEntries(tc.targetFile), tc.targetFile, tc.ruleBody, true)
				if tc.logBody != "" {
					writeFixtureFile(t, prj.log, tc.logBody)
				}
				before := snapshotTree(t, dir)
				p, gates, lockDir := newGatedPipeline(t)

				_, err := p.Execute(proposal(fxBefore, fxAfter), dir, mode.dryRun)
				if !tc.wantReject {
					if err != nil {
						t.Fatalf("Execute on a distinct rule file: %v", err)
					}
					if prj.rule == prj.registry || prj.rule == prj.log {
						t.Fatalf("control fixture is not distinct: rule %s", prj.rule)
					}
					assertLockReleased(t, lockDir)
					return
				}
				if err == nil {
					t.Fatal("Execute: want a rule-file-is-registry-or-log error, got nil")
				}
				if !strings.Contains(err.Error(), wantMsg) {
					t.Errorf("error %q lacks %q", err, wantMsg)
				}
				if !containsPathForm(err.Error(), prj.rule) {
					t.Errorf("error %q does not name the rule file %s", err, prj.rule)
				}
				if gates.calls != 0 {
					t.Errorf("gate doubles called %d times, want 0 (the check runs before Layer 1)", gates.calls)
				}
				assertSameTree(t, tc.name+" "+mode.name, dir, before)
				assertLockReleased(t, lockDir)
			})
		}
	}
}

// Sync-audit F1 — G-B decides by file identity when both paths exist, so an
// alias of the registry is rejected before Layer 1 exactly like the registry
// path itself: a hard link, a case-variant name on a case-insensitive
// filesystem, and a symbolic link. hardlink_log (delta-audit D2) pins the log
// half of the condition the same way: a hard link to an existing evolution log.
// absent_log_fallback pins the other branch: a rule file naming an evolution
// log that does not exist yet is still rejected by the cleaned-absolute-path
// comparison.
func TestExecute_RuleFileAliasOfRegistryOrLog_Rejected(t *testing.T) {
	const wantMsg = "is also the registry or the evolution log"
	const aliasFile = "rules/alias.md"
	cases := []struct {
		name, entryFile string
		withLog         bool
		aliasOfLog      bool // the alias names the evolution log, not the registry
		makeAlias       func(t *testing.T, prj project)
	}{
		{"hardlink_registry", aliasFile, true, false, func(t *testing.T, prj project) {
			if err := os.Link(prj.registry, prj.rule); err != nil {
				t.Skipf("platform refused a hard link (%v) — recorded as a Gap", err)
			}
		}},
		{"case_variant_registry", ".claude/rules/moai/core/ZONE-REGISTRY.md", true, false, func(t *testing.T, _ project) {
			skipUnlessCaseInsensitive(t)
		}},
		{"symlink_registry", aliasFile, true, false, func(t *testing.T, prj project) {
			symlinkOrSkip(t, prj.registry, prj.rule)
		}},
		{"hardlink_log", aliasFile, true, true, func(t *testing.T, prj project) {
			if err := os.Link(prj.log, prj.rule); err != nil {
				t.Skipf("platform refused a hard link (%v) — recorded as a Gap", err)
			}
		}},
		{"absent_log_fallback", ".moai/research/evolution-log.md", false, false, nil},
	}
	for _, tc := range cases {
		for _, mode := range []struct {
			name   string
			dryRun bool
		}{{"dry_run", true}, {"real", false}} {
			t.Run(tc.name+"/"+mode.name, func(t *testing.T) {
				isolateEnv(t)
				dir := t.TempDir()
				prj := buildProject(t, dir, standardEntries(tc.entryFile), tc.entryFile, "", tc.withLog)
				if tc.makeAlias != nil {
					tc.makeAlias(t, prj)
					// Setup sanity, independent of sameFile: the alias must
					// read back as its target's bytes.
					target, targetName := prj.registry, "registry"
					if tc.aliasOfLog {
						target, targetName = prj.log, "evolution log"
					}
					if readString(t, prj.rule) != readString(t, target) {
						t.Fatalf("alias %s does not read as the %s", prj.rule, targetName)
					}
				} else if _, err := os.Stat(prj.rule); err == nil {
					t.Fatalf("fallback fixture: %s exists, want it absent", prj.rule)
				}
				before := snapshotTree(t, dir)
				p, gates, lockDir := newGatedPipeline(t)

				_, err := p.Execute(proposal(fxBefore, fxAfter), dir, mode.dryRun)
				if err == nil {
					t.Fatal("Execute: want a rule-file-is-registry-or-log error, got nil")
				}
				if !strings.Contains(err.Error(), wantMsg) {
					t.Errorf("error %q lacks %q", err, wantMsg)
				}
				if !containsPathForm(err.Error(), prj.rule) {
					t.Errorf("error %q does not name the rule file %s", err, prj.rule)
				}
				if gates.calls != 0 {
					t.Errorf("gate doubles called %d times, want 0 (the check runs before Layer 1)", gates.calls)
				}
				assertSameTree(t, tc.name+" "+mode.name, dir, before)
				assertLockReleased(t, lockDir)
			})
		}
	}
}

// skipUnlessCaseInsensitive skips the test when the temporary filesystem tells
// a file name from its case-flipped form, since a case-variant alias cannot
// exist there.
func skipUnlessCaseInsensitive(t *testing.T) {
	t.Helper()
	probe := filepath.Join(t.TempDir(), "casefold-probe.txt")
	if err := os.WriteFile(probe, []byte("probe"), 0o644); err != nil {
		t.Fatalf("write probe: %v", err)
	}
	flipped := filepath.Join(filepath.Dir(probe), "CASEFOLD-PROBE.TXT")
	if _, err := os.Stat(flipped); err != nil {
		t.Skipf("temporary filesystem is case-sensitive (%v): a case-variant alias cannot exist here", err)
	}
}
