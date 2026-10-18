package cli

// migrate_ownership_test.go — SPEC-UPDATE-OWNERSHIP-REPAIR-001 M1a (card
// t1865): the single ownership verdict function and the preservation fixtures
// that judge it. AC-OWR-01 (the four classes and the disposition gates),
// AC-OWR-02 (an unregistered file byte-equal to a REAL embedded template path
// is never promoted), AC-OWR-03 (a user-modified file is kept), AC-OWR-04 (an
// unregistered file of unknown origin is kept with a reason and its missing
// evidence items).
//
// The preservation criteria carry a reach control (a sibling that must be
// removed in the same run, so the removal arm is shown to be reachable) and a
// mutant injection (a wrong implementation is swapped in through the same
// seams the migration uses, and the pure assertion function must go red).

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/template"
)

// fakeProvenance is a provenanceSource over fixed generations (newest first).
type fakeProvenance struct{ gens []priorGeneration }

func (f fakeProvenance) generations() []priorGeneration { return f.gens }

// swapProvenanceSource installs a provenance source factory for one test.
func swapProvenanceSource(t *testing.T, factory func(projectRoot string) provenanceSource) {
	t.Helper()
	prev := newProvenanceSource
	newProvenanceSource = factory
	t.Cleanup(func() { newProvenanceSource = prev })
}

// swapClassifier installs a classifier for one test (a mutant injection).
func swapClassifier(t *testing.T, fn func(ownershipEnv, string, string) ownershipVerdict) {
	t.Helper()
	prev := classifyOwnershipFn
	classifyOwnershipFn = fn
	t.Cleanup(func() { classifyOwnershipFn = prev })
}

// findRemaining returns the remaining row for rel.
func findRemaining(rc *migrationRunContext, rel string) (remainingRow, bool) {
	for _, r := range rc.remaining {
		if r.Rel == rel {
			return r, true
		}
	}
	return remainingRow{}, false
}

// preserveExpect is what one preserved file must look like after a run.
type preserveExpect struct {
	rel     string
	bytes   []byte
	class   ownershipClass
	reason  keepReason
	missing []evidenceItem // compared only when checkMissing
	// checkMissing makes the missing-evidence list part of the assertion.
	checkMissing bool
}

// ownershipPreserveProblems is the pure preservation assertion: it returns one
// problem per violated expectation, so a meta test can require a non-empty
// list in front of a mutant. reach is a sibling path that must have been
// removed in the same run (the removal arm is reachable); an empty reach skips
// that check.
func ownershipPreserveProblems(root string, removed []string, rc *migrationRunContext, reach string, exps []preserveExpect) []string {
	var problems []string
	if reach != "" && !slices.Contains(removed, reach) {
		problems = append(problems, fmt.Sprintf("reach control %s was not removed (removed=%v): the removal arm was not exercised", reach, removed))
	}
	for _, e := range exps {
		abs := filepath.Join(root, filepath.FromSlash(e.rel))
		got, err := os.ReadFile(abs)
		switch {
		case err != nil:
			problems = append(problems, fmt.Sprintf("%s: file is gone or unreadable: %v", e.rel, err))
		case string(got) != string(e.bytes):
			problems = append(problems, fmt.Sprintf("%s: bytes changed", e.rel))
		}
		if slices.Contains(removed, e.rel) {
			problems = append(problems, fmt.Sprintf("%s: reported as removed", e.rel))
		}
		if matches, _ := filepath.Glob(filepath.Join(root, ".moai-backups", "*", "removed-assets", filepath.FromSlash(e.rel))); len(matches) != 0 {
			problems = append(problems, fmt.Sprintf("%s: a removed-assets copy exists (%v)", e.rel, matches))
		}
		row, ok := findRemaining(rc, e.rel)
		if !ok {
			problems = append(problems, fmt.Sprintf("%s: no remaining row", e.rel))
			continue
		}
		if row.Class != e.class {
			problems = append(problems, fmt.Sprintf("%s: class %q, want %q", e.rel, row.Class, e.class))
		}
		if row.Reason != e.reason {
			problems = append(problems, fmt.Sprintf("%s: reason %q, want %q", e.rel, row.Reason, e.reason))
		}
		if e.checkMissing && !slices.Equal(row.Missing, e.missing) {
			problems = append(problems, fmt.Sprintf("%s: missing evidence %v, want %v", e.rel, row.Missing, e.missing))
		}
	}
	return problems
}

// ownershipRun is one direct migration call over a fixture.
type ownershipRun struct {
	root    string
	home    string
	removed []string
	rc      *migrationRunContext
	report  []string
}

// runOwnershipMigration runs the migration directly (removal switch ON for the
// fixture only, profile sandbox, no profile mapping) and collects its outcome.
func runOwnershipMigration(t *testing.T, root, home string) ownershipRun {
	t.Helper()
	rc := &migrationRunContext{userAssetsInstalled: true}
	var report []string
	removed, err := migrateProjectCommonAssets(root, home, true, nil, func(format string, args ...interface{}) {
		report = append(report, fmt.Sprintf(format, args...))
	}, rc)
	if err != nil {
		t.Fatalf("migrateProjectCommonAssets: %v", err)
	}
	return ownershipRun{root: root, home: home, removed: removed, rc: rc, report: report}
}

// ownershipFixtureBase prepares the removal switch, the profile sandbox, the
// user counterparts and the record-bearing fixture project.
func ownershipFixtureBase(t *testing.T) (root, home string) {
	t.Helper()
	enableRemovalForTest(t)
	ownProfileSandbox(t)
	root = buildMigrationFixture(t)
	home = installMigrationUserCounterparts(t)
	requireUserCounterpart(t, home, migIdenticalSkill)
	return root, home
}

// bytesEqualPromotingSource is the wrong implementation AC-OWR-02 injects: it
// answers all three evidence items "present" for every project common-asset
// file whose bytes equal the embedded template's bytes at the same path.
type bytesEqualPromotingSource struct {
	root     string
	embedded fs.FS
}

func (s bytesEqualPromotingSource) generations() []priorGeneration {
	files := map[string]priorRecord{}
	for _, rootRel := range projectCommonAssetRels {
		_ = filepath.WalkDir(filepath.Join(s.root, filepath.FromSlash(rootRel)), func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			rel, relErr := filepath.Rel(s.root, p)
			if relErr != nil {
				return nil
			}
			rel = filepath.ToSlash(rel)
			disk, rErr := os.ReadFile(p)
			tpl, tErr := fs.ReadFile(s.embedded, rel)
			if rErr == nil && tErr == nil && string(disk) == string(tpl) {
				h := manifest.HashBytes(disk)
				files[rel] = priorRecord{provenance: manifest.TemplateManaged, templateHash: h, deployedHash: h, deployedVersion: "v-mutant"}
			}
			return nil
		})
	}
	return []priorGeneration{{name: "mutant-bytes-equal", files: files}}
}

// dirNamePromotingSource is the wrong implementation AC-OWR-04 injects: it
// claims a template-managed record for every project common-asset file whose
// unit directory (a skill directory, or the agents directory) exists in the
// embedded template tree.
type dirNamePromotingSource struct {
	root     string
	embedded fs.FS
}

func (s dirNamePromotingSource) generations() []priorGeneration {
	files := map[string]priorRecord{}
	for _, rootRel := range projectCommonAssetRels {
		_ = filepath.WalkDir(filepath.Join(s.root, filepath.FromSlash(rootRel)), func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			rel, relErr := filepath.Rel(s.root, p)
			if relErr != nil {
				return nil
			}
			rel = filepath.ToSlash(rel)
			unit := strings.TrimSuffix(rootRel, "/")
			if strings.HasSuffix(rootRel, "skills/") {
				parts := strings.Split(strings.TrimPrefix(rel, rootRel), "/")
				if len(parts) < 2 {
					return nil
				}
				unit = rootRel + parts[0]
			}
			if _, statErr := fs.Stat(s.embedded, unit); statErr != nil {
				return nil
			}
			disk, rErr := os.ReadFile(p)
			if rErr != nil {
				return nil
			}
			h := manifest.HashBytes(disk)
			files[rel] = priorRecord{provenance: manifest.TemplateManaged, templateHash: h, deployedHash: h, deployedVersion: "v-mutant"}
			return nil
		})
	}
	return []priorGeneration{{name: "mutant-dir-name", files: files}}
}

// provenanceOnlyClassifier is the wrong implementation AC-OWR-03 injects: it
// ignores the file's current bytes and trusts the recorded provenance alone,
// by rewriting the registered template hash to the current bytes' hash.
func provenanceOnlyClassifier(env ownershipEnv, rel, abs string) ownershipVerdict {
	inner := env.entry
	env.entry = func(r string) (*manifest.FileEntry, bool) {
		e, ok := inner(r)
		if !ok || e == nil {
			return e, ok
		}
		cp := *e
		if r == rel {
			if data, err := os.ReadFile(abs); err == nil {
				cp.TemplateHash = manifest.HashBytes(data)
			}
		}
		return &cp, true
	}
	return classifyOwnership(env, rel, abs)
}

// TestOwnershipVerdictClasses is AC-OWR-01: the verdict function maps every
// combination of (registered provenance) x (bytes equal to the recorded
// template hash or not) and the unregistered shapes into the closed class set,
// and a managed file becomes removal-eligible only when every gate holds.
func TestOwnershipVerdictClasses(t *testing.T) {
	root := t.TempDir()
	rootResolved, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	tpl := []byte("# template body\n")
	tplHash := manifest.HashBytes(tpl)
	entries := map[string]*manifest.FileEntry{}

	type want struct {
		class    ownershipClass
		eligible bool
		keep     keepReason
	}
	cases := map[string]want{}
	for _, prov := range []manifest.Provenance{manifest.TemplateManaged, manifest.UserModified, manifest.UserCreated, manifest.Deprecated} {
		for _, match := range []bool{true, false} {
			rel := fmt.Sprintf(".claude/skills/matrix-%s-%t/SKILL.md", prov, match)
			data := append([]byte{}, tpl...)
			if !match {
				data = append(data, []byte("an edit\n")...)
			}
			ownWriteFile(t, root, rel, data)
			entries[rel] = &manifest.FileEntry{Provenance: prov, TemplateHash: tplHash, DeployedHash: tplHash, CurrentHash: manifest.HashBytes(data)}
			switch {
			case prov == manifest.UserCreated:
				cases[rel] = want{class: ownershipUserCreated, keep: keepUserCreated}
			case prov == manifest.UserModified:
				cases[rel] = want{class: ownershipUserModified, keep: keepUserModified}
			case match: // template_managed or deprecated, bytes equal
				cases[rel] = want{class: ownershipManaged, eligible: true}
			default:
				cases[rel] = want{class: ownershipUserModified, keep: keepUserModified}
			}
		}
	}
	// Two unregistered shapes: x1 matches no template, x2 is byte-equal to a
	// REAL embedded template path (a hash match is not ownership).
	const x1 = ".agents/skills/x1-custom/SKILL.md"
	ownWriteFile(t, root, x1, []byte("# nothing like any template\n"))
	cases[x1] = want{class: ownershipUnregistered, keep: keepNoProvenanceRecord}
	x2 := ownEmbeddedClaudeSkill
	ownWriteFile(t, root, x2, ownEmbeddedBytes(t, x2))
	cases[x2] = want{class: ownershipUnregistered, keep: keepNoProvenanceRecord}

	baseEnv := func() ownershipEnv {
		return ownershipEnv{
			entry: func(rel string) (*manifest.FileEntry, bool) {
				e, ok := entries[rel]
				return e, ok
			},
			evidence:             fakeProvenance{},
			counterpartConfirmed: func(string) bool { return true },
			notInCatalog:         func(string) bool { return false },
			removalEnabled:       true,
			rootResolved:         rootResolved,
		}
	}

	seen := map[ownershipClass]bool{}
	for rel, w := range cases {
		v := classifyOwnership(baseEnv(), rel, filepath.Join(root, filepath.FromSlash(rel)))
		seen[v.class] = true
		if !v.class.valid() {
			t.Errorf("%s: class %q is outside the closed set", rel, v.class)
		}
		if v.class != w.class {
			t.Errorf("%s: class %q, want %q", rel, v.class, w.class)
		}
		if v.removalEligible != w.eligible {
			t.Errorf("%s: removal-eligible=%v, want %v", rel, v.removalEligible, w.eligible)
		}
		if w.eligible {
			if v.keep != "" {
				t.Errorf("%s: eligible but keep reason %q set", rel, v.keep)
			}
		} else if v.keep != w.keep {
			t.Errorf("%s: keep %q, want %q", rel, v.keep, w.keep)
		}
		if v.keep != "" && !isKnownKeepReason(v.keep) {
			t.Errorf("%s: keep reason %q is outside the closed reason set", rel, v.keep)
		}
	}
	for _, c := range []ownershipClass{ownershipManaged, ownershipUserModified, ownershipUserCreated, ownershipUnregistered} {
		if !seen[c] {
			t.Errorf("the matrix never produced class %q", c)
		}
	}

	// Disposition gates: from the all-true managed baseline, removing exactly
	// one condition yields keep(<its reason>).
	m := fmt.Sprintf(".claude/skills/matrix-%s-%t/SKILL.md", manifest.TemplateManaged, true)
	mAbs := filepath.Join(root, filepath.FromSlash(m))
	if v := classifyOwnership(baseEnv(), m, mAbs); !v.removalEligible || v.class != ownershipManaged {
		t.Fatalf("baseline managed copy must be removal-eligible, got %+v", v)
	}
	gates := []struct {
		name   string
		mutate func(*ownershipEnv)
		want   keepReason
	}{
		{"profile reach", func(e *ownershipEnv) { e.profileMapped = true }, keepProfileReach},
		{"counterpart unconfirmed", func(e *ownershipEnv) { e.counterpartConfirmed = func(string) bool { return false } }, keepCounterpartUnconfirmed},
		{"not in catalog", func(e *ownershipEnv) {
			e.counterpartConfirmed = func(string) bool { return false }
			e.notInCatalog = func(string) bool { return true }
		}, keepNonCatalog},
		{"switch off", func(e *ownershipEnv) { e.removalEnabled = false }, keepRemovalGateOff},
	}
	for _, g := range gates {
		env := baseEnv()
		g.mutate(&env)
		v := classifyOwnership(env, m, mAbs)
		if v.class != ownershipManaged || v.removalEligible || v.keep != g.want {
			t.Errorf("gate %q: got class=%q eligible=%v keep=%q, want managed/false/%q", g.name, v.class, v.removalEligible, v.keep, g.want)
		}
	}

	// Root boundary: a registered copy that is a symlink out of the root.
	outside := t.TempDir()
	outBytes := []byte("# outside the project\n")
	if err := os.WriteFile(filepath.Join(outside, "SKILL.md"), outBytes, 0o644); err != nil {
		t.Fatal(err)
	}
	const esc = ".claude/skills/escape/SKILL.md"
	if err := os.MkdirAll(filepath.Join(root, ".claude", "skills", "escape"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(outside, "SKILL.md"), filepath.Join(root, filepath.FromSlash(esc))); err != nil {
		t.Skipf("symlinks unavailable here: %v", err)
	}
	entries[esc] = &manifest.FileEntry{Provenance: manifest.TemplateManaged, TemplateHash: manifest.HashBytes(outBytes)}
	if v := classifyOwnership(baseEnv(), esc, filepath.Join(root, filepath.FromSlash(esc))); v.class != ownershipManaged || v.removalEligible || v.keep != keepSymlinkEscape {
		t.Errorf("symlink escape: got class=%q eligible=%v keep=%q, want managed/false/%q", v.class, v.removalEligible, v.keep, keepSymlinkEscape)
	}

	// A retired (deprecated) copy is not in the profile-reach kept set.
	dep := fmt.Sprintf(".claude/skills/matrix-%s-%t/SKILL.md", manifest.Deprecated, true)
	env := baseEnv()
	env.profileMapped = true
	if v := classifyOwnership(env, dep, filepath.Join(root, filepath.FromSlash(dep))); !v.removalEligible {
		t.Errorf("deprecated copy under profile mapping: got %+v, want removal-eligible (retired copies are outside the kept set)", v)
	}
}

// TestOwnershipPreserveByteEqualUserFile is AC-OWR-02: unregistered files whose
// bytes equal the embedded template at a REAL template path stay put, with the
// reason and all three missing evidence items on their remaining rows.
func TestOwnershipPreserveByteEqualUserFile(t *testing.T) {
	paths := []string{ownEmbeddedClaudeSkill, ownEmbeddedAgentsSkill}

	setup := func(t *testing.T) (root, home string, exps []preserveExpect) {
		t.Helper()
		root, home = ownershipFixtureBase(t)
		for _, rel := range paths {
			requireUserCounterpart(t, home, rel)
			data := ownAddUnregisteredFromTemplate(t, root, rel)
			exps = append(exps, preserveExpect{
				rel: rel, bytes: data, class: ownershipUnregistered, reason: keepNoProvenanceRecord,
				missing: []evidenceItem{evidenceDeployedVersion, evidenceOriginalPath, evidencePriorRecord}, checkMissing: true,
			})
		}
		return root, home, exps
	}

	t.Run("Premise", func(t *testing.T) {
		root, home, _ := setup(t)
		for _, rel := range paths {
			embedded := ownEmbeddedBytes(t, rel)
			onDisk, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
			if err != nil || string(onDisk) != string(embedded) {
				t.Fatalf("%s: the planted bytes are not the embedded template's bytes: %v", rel, err)
			}
			requireUserCounterpart(t, home, rel)
		}
		mgr := manifest.NewManager()
		if _, err := mgr.Load(root); err != nil {
			t.Fatal(err)
		}
		for _, rel := range paths {
			if _, found := mgr.GetEntry(rel); found {
				t.Errorf("%s must be unregistered in the fixture", rel)
			}
		}
		if _, found := mgr.GetEntry(migIdenticalSkill); !found {
			t.Errorf("the reach-control sibling %s must be registered", migIdenticalSkill)
		}
	})

	t.Run("Preserve", func(t *testing.T) {
		root, home, exps := setup(t)
		run := runOwnershipMigration(t, root, home)
		if problems := ownershipPreserveProblems(root, run.removed, run.rc, migIdenticalSkill, exps); len(problems) != 0 {
			t.Errorf("preservation violated:\n  %s", strings.Join(problems, "\n  "))
		}
		mgr := manifest.NewManager()
		if _, err := mgr.Load(root); err != nil {
			t.Fatal(err)
		}
		for _, rel := range paths {
			if _, found := mgr.GetEntry(rel); found {
				t.Errorf("%s was adopted into the manifest; an unregistered file is never auto-adopted", rel)
			}
		}
	})

	t.Run("MutantIsCaught", func(t *testing.T) {
		root, home, exps := setup(t)
		embedded, err := template.EmbeddedTemplates()
		if err != nil {
			t.Fatal(err)
		}
		swapProvenanceSource(t, func(projectRoot string) provenanceSource {
			return bytesEqualPromotingSource{root: projectRoot, embedded: embedded}
		})
		run := runOwnershipMigration(t, root, home)
		problems := ownershipPreserveProblems(root, run.removed, run.rc, migIdenticalSkill, exps)
		if len(problems) == 0 {
			t.Fatalf("the byte-equal promoting source passed the preservation assertion; the criterion is too shallow (removed=%v)", run.removed)
		}
		joined := strings.Join(problems, "\n")
		for _, rel := range paths {
			if !strings.Contains(joined, rel) {
				t.Errorf("the mutant's damage to %s went unnoticed:\n%s", rel, joined)
			}
		}
	})
}

// TestOwnershipPreserveUserModifiedFile is AC-OWR-03: a registered file whose
// bytes differ from its template hash, and one recorded user_modified, are
// kept and classified user-modified.
func TestOwnershipPreserveUserModifiedFile(t *testing.T) {
	const agentRel = ".claude/agents/moai/manager-develop.md"

	setup := func(t *testing.T) (root, home string, exps []preserveExpect) {
		t.Helper()
		root, home = ownershipFixtureBase(t)
		requireUserCounterpart(t, home, migModifiedSkill)
		ownAddRegisteredEdited(t, root, agentRel, ownEmbeddedBytes(t, agentRel), "\n<!-- user edit -->\n", manifest.UserModified)
		for _, rel := range []string{migModifiedSkill, agentRel} {
			data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
			if err != nil {
				t.Fatal(err)
			}
			exps = append(exps, preserveExpect{rel: rel, bytes: data, class: ownershipUserModified, reason: keepUserModified})
		}
		return root, home, exps
	}

	t.Run("Reach", func(t *testing.T) {
		root, home, _ := setup(t)
		run := runOwnershipMigration(t, root, home)
		if !slices.Contains(run.removed, migIdenticalSkill) {
			t.Fatalf("the reach-control sibling %s was not removed (removed=%v)", migIdenticalSkill, run.removed)
		}
	})

	t.Run("Preserve", func(t *testing.T) {
		root, home, exps := setup(t)
		run := runOwnershipMigration(t, root, home)
		if problems := ownershipPreserveProblems(root, run.removed, run.rc, migIdenticalSkill, exps); len(problems) != 0 {
			t.Errorf("preservation violated:\n  %s", strings.Join(problems, "\n  "))
		}
	})

	t.Run("MutantIsCaught", func(t *testing.T) {
		root, home, exps := setup(t)
		swapClassifier(t, provenanceOnlyClassifier)
		run := runOwnershipMigration(t, root, home)
		problems := ownershipPreserveProblems(root, run.removed, run.rc, migIdenticalSkill, exps)
		if len(problems) == 0 {
			t.Fatalf("a classifier that ignores the current bytes passed the preservation assertion (removed=%v)", run.removed)
		}
		if !strings.Contains(strings.Join(problems, "\n"), migModifiedSkill) {
			t.Errorf("the mutant removing %s went unnoticed: %v", migModifiedSkill, problems)
		}
	})
}

// TestOwnershipUnregisteredUnknownOriginKeptWithReason is AC-OWR-04:
// unregistered files that match no template stay put; each remaining row and
// each report line carries the reason, and the rows name the missing evidence.
func TestOwnershipUnregisteredUnknownOriginKeptWithReason(t *testing.T) {
	files := map[string]string{
		".agents/skills/moai-custom-x/SKILL.md": "# a custom skill nobody shipped\n",
		".claude/agents/moai/custom-agent.md":   "# a custom agent nobody shipped\n",
		".agents/skills/moai-plan/extra.md":     "# an extra file beside a real template skill\n",
	}
	setup := func(t *testing.T) (root, home string, exps []preserveExpect) {
		t.Helper()
		root, home = ownershipFixtureBase(t)
		rels := make([]string, 0, len(files))
		for rel := range files {
			rels = append(rels, rel)
		}
		slices.Sort(rels)
		for _, rel := range rels {
			ownAddUnregistered(t, root, rel, []byte(files[rel]))
			exps = append(exps, preserveExpect{
				rel: rel, bytes: []byte(files[rel]), class: ownershipUnregistered, reason: keepNoProvenanceRecord,
				missing: []evidenceItem{evidenceDeployedVersion, evidenceOriginalPath, evidencePriorRecord}, checkMissing: true,
			})
		}
		return root, home, exps
	}

	t.Run("Reach", func(t *testing.T) {
		root, home, _ := setup(t)
		run := runOwnershipMigration(t, root, home)
		if !slices.Contains(run.removed, migIdenticalSkill) {
			t.Fatalf("the reach-control sibling %s was not removed (removed=%v)", migIdenticalSkill, run.removed)
		}
	})

	t.Run("Reasons", func(t *testing.T) {
		root, home, exps := setup(t)
		run := runOwnershipMigration(t, root, home)
		if problems := ownershipPreserveProblems(root, run.removed, run.rc, migIdenticalSkill, exps); len(problems) != 0 {
			t.Errorf("preservation violated:\n  %s", strings.Join(problems, "\n  "))
		}
		out := strings.Join(run.report, "\n")
		for _, e := range exps {
			line := ""
			for _, l := range run.report {
				if strings.Contains(l, e.rel) {
					line = l
					break
				}
			}
			if line == "" {
				t.Errorf("the migration output never names %s", e.rel)
				continue
			}
			if !strings.Contains(line, string(keepNoProvenanceRecord)) {
				t.Errorf("the output line for %s does not name its reason %q: %q", e.rel, keepNoProvenanceRecord, line)
			}
		}
		_ = out
	})

	t.Run("MutantIsCaught", func(t *testing.T) {
		root, home, exps := setup(t)
		embedded, err := template.EmbeddedTemplates()
		if err != nil {
			t.Fatal(err)
		}
		swapProvenanceSource(t, func(projectRoot string) provenanceSource {
			return dirNamePromotingSource{root: projectRoot, embedded: embedded}
		})
		run := runOwnershipMigration(t, root, home)
		problems := ownershipPreserveProblems(root, run.removed, run.rc, migIdenticalSkill, exps)
		if len(problems) == 0 {
			t.Fatalf("a source promoting by template directory name passed the reason/class assertions (rows=%+v)", run.rc.remaining)
		}
	})
}
