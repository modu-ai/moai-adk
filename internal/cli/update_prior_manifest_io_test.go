package cli

// update_prior_manifest_io_test.go — SPEC-UPDATE-OWNERSHIP-REPAIR-001 M1b (card
// t1865), AC-OWR-21: the prior generation is created atomically (a temporary
// file in the same directory, then a rename — the final name never holds a
// partial write), a damaged generation is no evidence, a manifest written before
// the deployed-version field loads unchanged, an older binary that drops the
// field weakens the evidence instead of inventing any, and
// priorManifestGenerations is a lookup bound only — nothing here deletes,
// prunes or shortens the retention of a run backup.

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/cli/update/backup"
	"github.com/modu-ai/moai-adk/internal/manifest"
)

// recordedStep is one observed call of the two atomic-write observation points.
type recordedStep struct {
	op        string // "temp" or "rename"
	dir       string // temp: the directory written to
	temp      string // temp: the temporary path produced
	from, to  string // rename arguments
	finalSeen bool   // whether manifest.prior.json existed when the step began
}

// swapPriorWriteHooks installs the two observation-point doubles for one test
// and returns the recorder. Each double calls the operating function unless an
// override replaces it, so the real file-system work still happens on the
// paths that inject nothing.
func swapPriorWriteHooks(t *testing.T,
	tempOverride func(dir string, data []byte, real func(string, []byte) (string, error)) (string, error),
	renameOverride func(oldPath, newPath string) error,
) *[]recordedStep {
	t.Helper()
	var steps []recordedStep
	prevTemp, prevRename := priorManifestTempWrite, priorManifestRename
	t.Cleanup(func() { priorManifestTempWrite, priorManifestRename = prevTemp, prevRename })

	finalExists := func(dir string) bool {
		_, err := os.Lstat(filepath.Join(dir, priorManifestName))
		return err == nil
	}
	priorManifestTempWrite = func(dir string, data []byte) (string, error) {
		step := recordedStep{op: "temp", dir: dir, finalSeen: finalExists(dir)}
		var path string
		var err error
		if tempOverride != nil {
			path, err = tempOverride(dir, data, prevTemp)
		} else {
			path, err = prevTemp(dir, data)
		}
		step.temp = path
		steps = append(steps, step)
		return path, err
	}
	priorManifestRename = func(oldPath, newPath string) error {
		steps = append(steps, recordedStep{op: "rename", from: oldPath, to: newPath, finalSeen: finalExists(filepath.Dir(newPath))})
		if renameOverride != nil {
			return renameOverride(oldPath, newPath)
		}
		return prevRename(oldPath, newPath)
	}
	return &steps
}

// TestPriorManifestAtomicWriteFailures is AC-OWR-21 (a): the operating write
// function reaches the final name only through a rename of a same-directory
// temporary file, and a failure at any step leaves no valid generation.
func TestPriorManifestAtomicWriteFailures(t *testing.T) {
	data := ownPriorGenerationJSON(t, "v1", "2026-01-01T00:00:00Z", map[string]ownPriorRecord{
		"a.md": {Provenance: manifest.TemplateManaged, TemplateHash: "sha256:aa", DeployedHash: "sha256:aa", DeployedVersion: "v1"},
	})
	injected := errors.New("injected failure")
	finalIn := func(dir string) string { return filepath.Join(dir, priorManifestName) }
	assertNotAGeneration := func(t *testing.T, dir string) {
		t.Helper()
		info, err := os.Lstat(finalIn(dir))
		if err == nil && info.Mode().IsRegular() {
			t.Errorf("a regular file stands at the final name after a failed write (a partial generation would read as valid)")
		}
		src := newPriorGenerationSource(filepath.Dir(filepath.Dir(dir)), "")
		if gens := src.generations(); len(gens) != 0 {
			t.Errorf("the reader counted %d generations after a failed write", len(gens))
		}
	}
	newRun := func(t *testing.T) (project, runDir string) {
		t.Helper()
		project = t.TempDir()
		runDir = filepath.Join(project, ".moai-backups", "20260101_000001")
		if err := os.MkdirAll(runDir, 0o755); err != nil {
			t.Fatal(err)
		}
		return project, runDir
	}

	t.Run("o-normal-run-goes-through-temp-then-rename", func(t *testing.T) {
		_, runDir := newRun(t)
		steps := swapPriorWriteHooks(t, nil, nil)
		if err := writePriorManifest(runDir, data); err != nil {
			t.Fatalf("write: %v", err)
		}
		if len(*steps) != 2 || (*steps)[0].op != "temp" || (*steps)[1].op != "rename" {
			t.Fatalf("observed steps %+v, want exactly one temporary write followed by one rename", *steps)
		}
		tmp, ren := (*steps)[0], (*steps)[1]
		if tmp.dir != runDir {
			t.Errorf("temporary write targeted %q, want the run directory %q", tmp.dir, runDir)
		}
		if filepath.Base(tmp.temp) == priorManifestName || filepath.Dir(tmp.temp) != runDir {
			t.Errorf("temporary path %q must sit in the run directory under a name other than %s", tmp.temp, priorManifestName)
		}
		if ren.from != tmp.temp || ren.to != finalIn(runDir) {
			t.Errorf("rename %q -> %q, want %q -> %q", ren.from, ren.to, tmp.temp, finalIn(runDir))
		}
		if tmp.finalSeen || ren.finalSeen {
			t.Errorf("the final name already existed when a step began (temp=%v rename=%v): it must come into being by the rename alone", tmp.finalSeen, ren.finalSeen)
		}
		got, err := os.ReadFile(finalIn(runDir))
		if err != nil || !bytes.Equal(got, data) {
			t.Errorf("final bytes differ from the input: %v", err)
		}
		if _, err := os.Lstat(tmp.temp); !os.IsNotExist(err) {
			t.Errorf("the temporary file %s was not consumed (%v)", tmp.temp, err)
		}
	})

	t.Run("p-failure-during-the-temporary-write", func(t *testing.T) {
		project, runDir := newRun(t)
		swapPriorWriteHooks(t, func(dir string, data []byte, real func(string, []byte) (string, error)) (string, error) {
			path, err := real(dir, data[:len(data)/2]) // only the first half reaches the temporary file
			if err != nil {
				return path, err
			}
			return path, injected
		}, nil)
		err := writePriorManifest(runDir, data)
		if err == nil {
			t.Fatalf("a failed temporary write returned nil: the hand-off would read it as success")
		}
		if _, statErr := os.Lstat(finalIn(runDir)); !os.IsNotExist(statErr) {
			t.Errorf("the final name exists after a failed temporary write (%v)", statErr)
		}
		_ = project
		assertNotAGeneration(t, runDir)
	})

	t.Run("q-failure-between-the-temporary-write-and-the-rename", func(t *testing.T) {
		_, runDir := newRun(t)
		var atRename []byte
		var atRenameErr error
		steps := swapPriorWriteHooks(t, nil, func(oldPath, _ string) error {
			// The premise is read where it holds: at the moment of the rename the
			// temporary file carries the complete bytes. (The writer removes the
			// temporary file once the rename has failed, so it is not read after.)
			atRename, atRenameErr = os.ReadFile(oldPath)
			return injected // the rename double reports failure without renaming
		})
		err := writePriorManifest(runDir, data)
		if err == nil {
			t.Fatalf("a failed rename returned nil")
		}
		if len(*steps) != 2 || (*steps)[1].op != "rename" {
			t.Fatalf("steps %+v: the rename observation point was not reached", *steps)
		}
		if atRenameErr != nil || !bytes.Equal(atRename, data) {
			t.Fatalf("test premise: the temporary file should hold the complete bytes at the rename: %v", atRenameErr)
		}
		if _, statErr := os.Lstat(finalIn(runDir)); !os.IsNotExist(statErr) {
			t.Errorf("the final name exists after the rename failed (%v): a complete-looking temporary file must NOT count as a generation", statErr)
		}
		assertNotAGeneration(t, runDir)
	})

	t.Run("r-real-rename-failure-onto-an-occupied-name", func(t *testing.T) {
		_, runDir := newRun(t)
		occupant := filepath.Join(finalIn(runDir), "keep")
		if err := os.MkdirAll(filepath.Dir(occupant), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(occupant, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := writePriorManifest(runDir, data); err == nil {
			t.Fatalf("renaming onto a non-empty directory returned nil")
		}
		got, err := os.ReadFile(occupant)
		if err != nil || string(got) != "x" {
			t.Errorf("the occupant directory's content was disturbed: %v", err)
		}
		assertNotAGeneration(t, runDir)
	})
}

// TestPriorGenerationCorruptIsNoEvidence is AC-OWR-21 (b): every shape of
// damaged generation is skipped, and a file whose only evidence is damaged is
// kept with an evidence reason — the run does not die and nothing is promoted.
func TestPriorGenerationCorruptIsNoEvidence(t *testing.T) {
	const p = ".agents/skills/moai-plan/SKILL.md"
	pBytes := ownEmbeddedBytes(t, p)
	good := ownPriorGenerationJSON(t, "v-gen", "2026-01-01T00:00:00Z", map[string]ownPriorRecord{
		p: {Provenance: manifest.TemplateManaged, TemplateHash: manifest.HashBytes(pBytes), DeployedHash: manifest.HashBytes(pBytes), DeployedVersion: "v-gen"},
	})

	shapes := []struct {
		name  string
		setup func(t *testing.T, project string)
		valid bool
	}{
		{"intact", func(t *testing.T, project string) { ownWritePriorGeneration(t, project, "20260101_000001", good) }, true},
		{"truncated-json", func(t *testing.T, project string) {
			ownWritePriorGeneration(t, project, "20260101_000001", good[:len(good)/2])
		}, false},
		{"empty-file", func(t *testing.T, project string) { ownWritePriorGeneration(t, project, "20260101_000001", nil) }, false},
		{"json-but-not-a-manifest", func(t *testing.T, project string) {
			ownWritePriorGeneration(t, project, "20260101_000001", []byte(`["not","a","manifest"]`))
		}, false},
		{"json-object-without-files", func(t *testing.T, project string) {
			ownWritePriorGeneration(t, project, "20260101_000001", []byte(`{"version":"v"}`))
		}, false},
		{"directory-in-place-of-the-file", func(t *testing.T, project string) {
			if err := os.MkdirAll(filepath.Join(project, ".moai-backups", "20260101_000001", priorManifestName, "x"), 0o755); err != nil {
				t.Fatal(err)
			}
		}, false},
		{"only-a-temporary-name-exists", func(t *testing.T, project string) {
			// The complete-looking temporary file of an interrupted write: the
			// reader reads the exact name only, so it is not a generation.
			dir := filepath.Join(project, ".moai-backups", "20260101_000001")
			if err := os.MkdirAll(dir, 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, priorManifestName+".tmp-123456"), good, 0o644); err != nil {
				t.Fatal(err)
			}
		}, false},
	}
	for _, sh := range shapes {
		t.Run(sh.name, func(t *testing.T) {
			project := t.TempDir()
			sh.setup(t, project)
			src := newPriorGenerationSource(project, "")
			gens := src.generations()
			if sh.valid != (len(gens) == 1) {
				t.Fatalf("valid=%v but the reader returned %d generations", sh.valid, len(gens))
			}
			rootResolved, _ := filepath.EvalSymlinks(project)
			ownWriteFile(t, project, p, pBytes)
			v := classifyOwnership(ownershipEnv{
				entry:                func(string) (*manifest.FileEntry, bool) { return nil, false },
				evidence:             src,
				counterpartConfirmed: func(string) bool { return true },
				notInCatalog:         func(string) bool { return false },
				removalEnabled:       true,
				rootResolved:         rootResolved,
			}, p, filepath.Join(project, filepath.FromSlash(p)))
			if sh.valid {
				if v.class != ownershipManaged {
					t.Errorf("an intact generation should attest the file, got %+v", v)
				}
				return
			}
			if v.class != ownershipUnregistered || v.removalEligible || v.attestation != nil {
				t.Errorf("a damaged generation produced %+v: it must be no evidence", v)
			}
			if v.keep != keepNoProvenanceRecord {
				t.Errorf("keep reason %q, want %q (no readable generation at all)", v.keep, keepNoProvenanceRecord)
			}
		})
	}
}

// TestLegacyManifestRoundTrip is AC-OWR-21 (c): a manifest written before the
// field existed loads, saves byte for byte, and its entries read as lacking the
// deployed-version evidence.
func TestLegacyManifestRoundTrip(t *testing.T) {
	project := t.TempDir()
	const p = ".agents/skills/moai-plan/SKILL.md"
	pBytes := ownEmbeddedBytes(t, p)
	h := manifest.HashBytes(pBytes)
	legacy := &manifest.Manifest{
		Version:    "v3.0.0",
		DeployedAt: "2026-01-02T03:04:05Z",
		Files: map[string]manifest.FileEntry{
			p:          {Provenance: manifest.TemplateManaged, TemplateHash: h, DeployedHash: h, CurrentHash: h},
			"other.md": {Provenance: manifest.TemplateManaged, TemplateHash: "sha256:x", DeployedHash: "sha256:x", CurrentHash: "sha256:x"},
			"third.md": {Provenance: manifest.UserModified, TemplateHash: "sha256:y", DeployedHash: "sha256:y", CurrentHash: "sha256:z"},
		},
	}
	want, err := json.MarshalIndent(legacy, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	want = append(want, '\n')
	if bytes.Contains(want, []byte("deployed_version")) {
		t.Fatalf("the legacy fixture must not carry the new key")
	}
	if err := os.MkdirAll(filepath.Join(project, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, ".moai", "manifest.json"), want, 0o644); err != nil {
		t.Fatal(err)
	}
	mgr := manifest.NewManager()
	if _, err := mgr.Load(project); err != nil {
		t.Fatal(err)
	}
	if err := mgr.Save(); err != nil {
		t.Fatal(err)
	}
	if got := readManifestBytes(t, project); !bytes.Equal(got, want) {
		t.Errorf("legacy manifest did not round-trip byte for byte (the load/save invented a field)")
	}
	for rel, e := range mgr.Manifest().Files {
		if e.DeployedVersion != "" {
			t.Errorf("%s: loading invented DeployedVersion %q", rel, e.DeployedVersion)
		}
	}

	// The same legacy bytes as a prior generation: the entry exists, so the
	// path and the provenance are fine, but the deployed version is missing.
	ownWritePriorGeneration(t, project, "20260101_000001", want)
	ownWriteFile(t, project, p, pBytes)
	rootResolved, _ := filepath.EvalSymlinks(project)
	v := classifyOwnership(ownershipEnv{
		entry:                func(string) (*manifest.FileEntry, bool) { return nil, false },
		evidence:             newPriorGenerationSource(project, ""),
		counterpartConfirmed: func(string) bool { return true },
		notInCatalog:         func(string) bool { return false },
		removalEnabled:       true,
		rootResolved:         rootResolved,
	}, p, filepath.Join(project, filepath.FromSlash(p)))
	if v.class != ownershipUnregistered || v.keep != evidenceIncompleteReason(evidenceDeployedVersion) || !reflect.DeepEqual(v.missing, []evidenceItem{evidenceDeployedVersion}) {
		t.Errorf("a legacy generation should read as lacking only the deployed version, got %+v", v)
	}
}

// TestDowngradeDropsEvidenceNotSafety is AC-OWR-21 (d): an older binary that
// does not know deployed_version drops it on its next save; the next newer
// binary then finds less evidence and keeps the file. Evidence weakens; no
// removal-direction error appears.
func TestDowngradeDropsEvidenceNotSafety(t *testing.T) {
	// The older binary's view of the manifest: the same JSON shape without the field.
	type oldFileEntry struct {
		Provenance   manifest.Provenance `json:"provenance"`
		TemplateHash string              `json:"template_hash"`
		DeployedHash string              `json:"deployed_hash"`
		CurrentHash  string              `json:"current_hash"`
	}
	type oldManifest struct {
		Version          string                  `json:"version"`
		DeployedAt       string                  `json:"deployed_at"`
		Files            map[string]oldFileEntry `json:"files"`
		TemplateTreeHash string                  `json:"template_tree_hash,omitempty"`
	}

	const p = ".agents/skills/moai-plan/SKILL.md"
	pBytes := ownEmbeddedBytes(t, p)
	h := manifest.HashBytes(pBytes)
	withField := ownPriorGenerationJSON(t, "v-new", "2026-01-01T00:00:00Z", map[string]ownPriorRecord{
		p: {Provenance: manifest.TemplateManaged, TemplateHash: h, DeployedHash: h, DeployedVersion: "v-new"},
	})
	if !bytes.Contains(withField, []byte(`"deployed_version"`)) {
		t.Fatal("test premise: the manifest written by the newer binary carries the field")
	}
	var seen oldManifest
	if err := json.Unmarshal(withField, &seen); err != nil {
		t.Fatalf("the older binary must still load the newer manifest: %v", err)
	}
	downgraded, err := json.MarshalIndent(seen, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(downgraded, []byte("deployed_version")) {
		t.Fatal("test premise: the older binary's save drops the field")
	}

	for _, tc := range []struct {
		name      string
		gen       []byte
		attested  bool
		wantKeep  keepReason
		wantClass ownershipClass
	}{
		{"newer binary's generation attests", withField, true, "", ownershipManaged},
		{"after the older binary saved, the file is kept", downgraded, false, evidenceIncompleteReason(evidenceDeployedVersion), ownershipUnregistered},
	} {
		t.Run(tc.name, func(t *testing.T) {
			project := t.TempDir()
			ownWritePriorGeneration(t, project, "20260101_000001", tc.gen)
			ownWriteFile(t, project, p, pBytes)
			rootResolved, _ := filepath.EvalSymlinks(project)
			v := classifyOwnership(ownershipEnv{
				entry:                func(string) (*manifest.FileEntry, bool) { return nil, false },
				evidence:             newPriorGenerationSource(project, ""),
				counterpartConfirmed: func(string) bool { return true },
				notInCatalog:         func(string) bool { return false },
				removalEnabled:       true,
				rootResolved:         rootResolved,
			}, p, filepath.Join(project, filepath.FromSlash(p)))
			if v.class != tc.wantClass || v.keep != tc.wantKeep || (v.attestation != nil) != tc.attested {
				t.Errorf("got class=%q keep=%q attested=%v, want class=%q keep=%q attested=%v", v.class, v.keep, v.attestation != nil, tc.wantClass, tc.wantKeep, tc.attested)
			}
			if !tc.attested && v.removalEligible {
				t.Errorf("a downgraded generation led to a removal-eligible verdict")
			}
		})
	}
}

// priorManifestGenerationsUses returns the names of the non-test functions in
// this package that mention the priorManifestGenerations constant, excluding
// its own declaration.
func priorManifestGenerationsUses(t *testing.T) []string {
	t.Helper()
	fset := token.NewFileSet()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	var uses []string
	for _, e := range entries {
		n := e.Name()
		if e.IsDir() || !strings.HasSuffix(n, ".go") || strings.HasSuffix(n, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, n, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, decl := range f.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			ast.Inspect(fd, func(node ast.Node) bool {
				if id, ok := node.(*ast.Ident); ok && id.Name == "priorManifestGenerations" {
					uses = append(uses, fd.Name.Name)
				}
				return true
			})
		}
	}
	sort.Strings(uses)
	return uses
}

// TestPriorGenerationsAreLookupBoundOnly is AC-OWR-21 (e): priorManifestGenerations
// bounds how many generations are READ and is consulted by no deletion; the
// existing retention policy selects run backups with and without a prior
// manifest by one rule.
func TestPriorGenerationsAreLookupBoundOnly(t *testing.T) {
	t.Run("OnlyTheReaderUsesTheBound", func(t *testing.T) {
		uses := priorManifestGenerationsUses(t)
		if len(uses) != 1 || uses[0] != "readPriorGenerations" {
			t.Errorf("priorManifestGenerations is used by %v, want exactly the reading function readPriorGenerations", uses)
		}
	})

	// A project with 12 run backups. The retention policy keeps the newest 5
	// whole; the rest survive only if the selection rule requires them.
	build := func(t *testing.T, withPrior bool, distinctPrior bool) string {
		t.Helper()
		project := t.TempDir()
		base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		for i := 1; i <= 12; i++ {
			name := fmt.Sprintf("20260101_0000%02d", i)
			dir := filepath.Join(project, ".moai-backups", name)
			if err := os.MkdirAll(filepath.Join(dir, "sections"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "sections", "user.yaml"), []byte("a: 1\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			if withPrior {
				body := []byte(`{"version":"v","deployed_at":"x","files":{}}` + "\n")
				if distinctPrior {
					body = []byte(fmt.Sprintf(`{"version":"v%d","deployed_at":"x","files":{}}`+"\n", i))
				}
				if err := os.WriteFile(filepath.Join(dir, priorManifestName), body, 0o644); err != nil {
					t.Fatal(err)
				}
			}
			when := base.Add(time.Duration(i) * time.Hour)
			if err := os.Chtimes(dir, when, when); err != nil {
				t.Fatal(err)
			}
		}
		return project
	}
	survivors := func(t *testing.T, project string) []string {
		t.Helper()
		entries, err := os.ReadDir(filepath.Join(project, ".moai-backups"))
		if err != nil {
			t.Fatal(err)
		}
		var names []string
		for _, e := range entries {
			names = append(names, e.Name())
		}
		sort.Strings(names)
		return names
	}

	t.Run("SameRuleWithAndWithoutPriorManifest", func(t *testing.T) {
		without := build(t, false, false)
		with := build(t, true, false)
		rw := backup.PruneBackups(without, 5, 0)
		rp := backup.PruneBackups(with, 5, 0)
		if rw.Runs != 12 || rp.Runs != 12 {
			t.Fatalf("premise: both projects count 12 runs, got %d and %d", rw.Runs, rp.Runs)
		}
		if a, b := survivors(t, without), survivors(t, with); !slices.Equal(a, b) {
			t.Errorf("run backups carrying a prior manifest are selected by a different rule:\nwithout: %v\nwith:    %v", a, b)
		}
		if got := len(survivors(t, with)); got < 5 {
			t.Errorf("the retention window (newest 5) was not kept: %d survive", got)
		}
	})

	t.Run("MeasuredWhenEveryRunHoldsADifferentPriorManifest", func(t *testing.T) {
		// Gap G-K, measured: the census hashes every regular file of a run, so
		// a prior manifest that differs per run could make the policy keep
		// older runs. The policy is unchanged either way; the invariant this
		// asserts is that the window survives whole and nothing but the policy
		// decided what else stayed.
		without := build(t, false, false)
		with := build(t, true, true)
		backup.PruneBackups(without, 5, 0)
		backup.PruneBackups(with, 5, 0)
		kept := survivors(t, with)
		t.Logf("G-K measurement: without prior manifests %d run backups survive %v; with distinct prior manifests %d survive %v",
			len(survivors(t, without)), survivors(t, without), len(kept), kept)
		for i := 8; i <= 12; i++ {
			name := fmt.Sprintf("20260101_0000%02d", i)
			if !slices.Contains(kept, name) {
				t.Errorf("window run %s did not survive", name)
			}
		}
	})
}
