package cli

// update_prior_manifest_test.go — SPEC-UPDATE-OWNERSHIP-REPAIR-001 M1b (card
// t1865), the evidence-production half (decision-index Q2 (a), conditional):
//
//	AC-OWR-06  deployed version recorded only on entries a run deployed, the
//	           manifest as loaded at the start of each run kept byte for byte in
//	           that run's backup, and a bounded, tolerant reader of the retained
//	           generations;
//	AC-OWR-20  when the run backup cannot hold the prior-manifest copy or a
//	           removal candidate's bytes, the run removes NOTHING — driven through
//	           the production call path (Backup stage -> migration), never
//	           through a stand-in context;
//	AC-OWR-05 (h) the operating evidence source: a prior generation produced by
//	           the real deployer, manifest.Track and the operating writer.
//
// Every flow runs in a temporary project with a temporary HOME; the removal
// switch is turned on inside the fixture only.

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/pkg/version"
)

// setBinaryVersion sets the version the update flow reports as the writing
// binary for one test.
func setBinaryVersion(t *testing.T, v string) {
	t.Helper()
	prev := version.Version
	version.Version = v
	t.Cleanup(func() { version.Version = prev })
}

// loadManifestFile loads the project manifest from disk.
func loadManifestFile(t *testing.T, root string) *manifest.Manifest {
	t.Helper()
	mgr := manifest.NewManager()
	mf, err := mgr.Load(root)
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	return mf
}

// readManifestBytes returns the manifest file's bytes.
func readManifestBytes(t *testing.T, root string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, ".moai", "manifest.json"))
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	return data
}

// rewriteManifestNonCanonical re-serialises the project manifest in a shape
// Save would never produce (four-space indent, trailing blank line), after an
// optional edit of the raw file map, so a copy made by re-serialising cannot
// equal a copy made from the original bytes.
func rewriteManifestNonCanonical(t *testing.T, root string, edit func(files map[string]map[string]any, top map[string]any)) {
	t.Helper()
	raw := readManifestBytes(t, root)
	var top map[string]any
	if err := json.Unmarshal(raw, &top); err != nil {
		t.Fatalf("parse manifest: %v", err)
	}
	files := map[string]map[string]any{}
	if rawFiles, ok := top["files"].(map[string]any); ok {
		for k, v := range rawFiles {
			files[k] = v.(map[string]any)
		}
	}
	if edit != nil {
		edit(files, top)
	}
	asAny := map[string]any{}
	for k, v := range files {
		asAny[k] = v
	}
	top["files"] = asAny
	out, err := json.MarshalIndent(top, "", "    ")
	if err != nil {
		t.Fatal(err)
	}
	out = append(out, '\n', '\n')
	if err := os.WriteFile(filepath.Join(root, ".moai", "manifest.json"), out, 0o644); err != nil {
		t.Fatal(err)
	}
}

// runForcedFlowWithContext drives the production Backup stage -> migration call
// path with --force (the version-match gate would skip a plain same-version
// run) and hands back the migration context the stage filled in.
func runForcedFlowWithContext(t *testing.T, root string) (string, *migrationRunContext) {
	t.Helper()
	ownAssertHomeIsolated(t)
	cmd := &cobra.Command{Use: "update"}
	cmd.Flags().Bool("force", true, "")
	cmd.Flags().Bool("yes", true, "")
	cmd.Flags().Bool("no-hooks", true, "")
	cmd.Flags().Bool("no-plugin", false, "")
	cmd.Flags().Bool("dry-run", false, "")
	cmd.Flags().String("check", "", "")
	var out strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(&strings.Builder{})
	cmd.SetContext(context.Background())

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	rc := &migrationRunContext{userAssetsInstalled: true}
	if err := runTemplateSyncWithReporter(cmd, nil, true, rc); err != nil {
		t.Fatalf("runTemplateSyncWithReporter: %v", err)
	}
	return out.String(), rc
}

// swapPriorManifestWriter installs a prior-manifest writer for one test.
func swapPriorManifestWriter(t *testing.T, w func(runDir string, data []byte) error) {
	t.Helper()
	prev := priorManifestWriter
	priorManifestWriter = w
	t.Cleanup(func() { priorManifestWriter = prev })
}

// TestPriorManifestRetainedInRunBackup is AC-OWR-06 (b): three update runs, the
// third at the same version as the second.
func TestPriorManifestRetainedInRunBackup(t *testing.T) {
	ownProfileSandbox(t)
	home := installMigrationUserCounterparts(t)
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	root := buildMigrationFixture(t)

	const v1, v2 = "v9.1.0", "v9.2.0"
	const sentinelAt = "2000-01-01T00:00:00Z"
	// Entries no run deploys: project-side common assets are installed in the
	// user home, not by the project deployer. migModifiedSkill carries an OLD
	// stamp that must survive; migIdenticalSkill carries none and must stay so.
	rewriteManifestNonCanonical(t, root, func(files map[string]map[string]any, top map[string]any) {
		files[migModifiedSkill]["deployed_version"] = "v-old"
		top["deployed_at"] = sentinelAt
	})

	// Run 1 (plain, v1): the version differs from the project's, so the run
	// reaches reconciliation.
	setBinaryVersion(t, v1)
	run1 := runUpdateFlowPlain(t, root)
	if run1.Skipped {
		t.Fatalf("run 1 was skipped by the version-match gate; nothing was judged")
	}
	if run1.NewPriorManifest == "" {
		t.Fatalf("run 1 wrote no manifest.prior.json into its run backup (the Backup stage did not preserve the start manifest)")
	}
	prior1, err := os.ReadFile(run1.NewPriorManifest)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(prior1, run1.StartManifest) {
		t.Errorf("run 1: manifest.prior.json is not the manifest as loaded at the start of the run, byte for byte (re-serialised?)")
	}
	assertNoTempResidue(t, filepath.Dir(run1.NewPriorManifest))

	mf1 := loadManifestFile(t, root)
	if mf1.Version != v1 {
		t.Errorf("run 1: manifest-level Version %q, want %q", mf1.Version, v1)
	}
	if mf1.DeployedAt == "" || mf1.DeployedAt == sentinelAt {
		t.Errorf("run 1: manifest-level DeployedAt %q was not refreshed when Version changed", mf1.DeployedAt)
	}
	stamped := 0
	for rel, e := range mf1.Files {
		if e.DeployedVersion == "" {
			continue
		}
		stamped++
		if e.Provenance != manifest.TemplateManaged {
			t.Errorf("run 1: %s carries DeployedVersion %q but provenance %q", rel, e.DeployedVersion, e.Provenance)
		}
		if rel != migModifiedSkill && e.DeployedVersion != v1 {
			t.Errorf("run 1: %s DeployedVersion %q, want %q", rel, e.DeployedVersion, v1)
		}
	}
	if stamped < 2 {
		t.Errorf("run 1: only %d entries carry a deployed version; the deploy path stamped nothing", stamped)
	}
	if mf1.Files[migModifiedSkill].DeployedVersion != "v-old" {
		t.Errorf("run 1: an entry the run did not deploy lost or changed its old stamp: %q", mf1.Files[migModifiedSkill].DeployedVersion)
	}
	if mf1.Files[migIdenticalSkill].DeployedVersion != "" {
		t.Errorf("run 1: an entry the run did not deploy was back-filled: %q", mf1.Files[migIdenticalSkill].DeployedVersion)
	}
	// Pin the manifest-level timestamp so run 2 proves a version change moves it.
	rewriteManifestNonCanonical(t, root, func(_ map[string]map[string]any, top map[string]any) { top["deployed_at"] = sentinelAt })

	// Run 2 (--force, v2).
	setBinaryVersion(t, v2)
	run2 := runUpdateFlowForced(t, root)
	if run2.NewPriorManifest == "" {
		t.Fatalf("run 2 wrote no manifest.prior.json")
	}
	prior2, _ := os.ReadFile(run2.NewPriorManifest)
	if !bytes.Equal(prior2, run2.StartManifest) {
		t.Errorf("run 2: manifest.prior.json differs from the manifest as loaded at the start of run 2")
	}
	if run2.NewPriorManifest == run1.NewPriorManifest {
		t.Errorf("runs 1 and 2 share one run backup directory")
	}
	mf2 := loadManifestFile(t, root)
	if mf2.Version != v2 {
		t.Errorf("run 2: manifest-level Version %q, want %q", mf2.Version, v2)
	}
	if mf2.DeployedAt == sentinelAt {
		t.Errorf("run 2: DeployedAt was not refreshed although Version changed")
	}
	for rel, e := range mf2.Files {
		if e.DeployedVersion == v2 && e.Provenance != manifest.TemplateManaged {
			t.Errorf("run 2: %s was stamped %q with provenance %q", rel, v2, e.Provenance)
		}
	}
	for _, rel := range []string{migModifiedSkill, migIdenticalSkill} {
		if got := mf2.Files[rel].DeployedVersion; got == v2 {
			t.Errorf("run 2: %s (never deployed by a run) was stamped with the new version %q", rel, got)
		}
	}
	if mf2.Files[migModifiedSkill].DeployedVersion != "v-old" {
		t.Errorf("run 2: the old stamp on a non-deployed entry changed: %q", mf2.Files[migModifiedSkill].DeployedVersion)
	}
	assertNoTempResidue(t, filepath.Dir(run2.NewPriorManifest))
	after2 := readManifestBytes(t, root)

	// Run 3 (--force, the SAME version): the manifest bytes do not move. The
	// timestamp has one-second resolution, so wait it out: an implementation
	// that refreshes DeployedAt on every run would otherwise hide inside the
	// same second as run 2.
	time.Sleep(1100 * time.Millisecond)
	run3 := runUpdateFlowForced(t, root)
	if !bytes.Equal(run3.StartManifest, after2) {
		t.Fatalf("test bug: run 3 started from different bytes than run 2 left")
	}
	after3 := readManifestBytes(t, root)
	if !bytes.Equal(after2, after3) {
		t.Errorf("a repeat run at the same version changed the manifest bytes (DeployedAt moved or an entry was rewritten):\n--- after run 2\n%s\n--- after run 3\n%s", after2, after3)
	}
}

// assertNoTempResidue fails when a temporary prior-manifest file is left in dir.
func assertNoTempResidue(t *testing.T, dir string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.Contains(e.Name(), ".tmp") {
			t.Errorf("temporary file %s left behind in %s after a successful run", e.Name(), dir)
		}
	}
}

// priorGenJSON renders a minimal valid prior generation naming one path.
func priorGenJSON(t *testing.T, rel, ver string) []byte {
	t.Helper()
	h := manifest.HashBytes([]byte("bytes of " + rel))
	return ownPriorGenerationJSON(t, "v-gen", "2026-01-01T00:00:00Z", map[string]ownPriorRecord{
		rel: {Provenance: manifest.TemplateManaged, TemplateHash: h, DeployedHash: h, DeployedVersion: ver},
	})
}

// TestPriorGenerationSourceBoundedAndTolerant is AC-OWR-06 (c): the operating
// evidence source reads at most priorManifestGenerations of the newest run
// backups and treats every damaged generation as no evidence.
func TestPriorGenerationSourceBoundedAndTolerant(t *testing.T) {
	t.Run("BoundedToTheNewestEight", func(t *testing.T) {
		root := t.TempDir()
		for i := 1; i <= 10; i++ {
			name := fmt.Sprintf("20260101_0000%02d", i)
			ownWritePriorGeneration(t, root, name, priorGenJSON(t, fmt.Sprintf("gen-%02d-only.md", i), fmt.Sprintf("v%d", i)))
		}
		src := newPriorGenerationSource(root, "")
		gens := src.generations()
		if len(gens) != priorManifestGenerations {
			t.Fatalf("read %d generations, want exactly the bound %d", len(gens), priorManifestGenerations)
		}
		if src.reads != priorManifestGenerations {
			t.Errorf("read %d generation files, want %d (the 9th and later must not be opened)", src.reads, priorManifestGenerations)
		}
		if _, ok := gens[0].files["gen-10-only.md"]; !ok {
			t.Errorf("generations are not newest first: first holds %v", priorKeysOf(gens[0].files))
		}
		for _, g := range gens {
			if _, ok := g.files["gen-01-only.md"]; ok {
				t.Errorf("the oldest generation was read although it is beyond the bound")
			}
			if _, ok := g.files["gen-02-only.md"]; ok {
				t.Errorf("the second-oldest generation was read although it is beyond the bound")
			}
		}
	})

	t.Run("DamagedGenerationsAreNoEvidenceAndDoNotKillTheRun", func(t *testing.T) {
		root := t.TempDir()
		good := priorGenJSON(t, "good-newest.md", "v-new")
		older := priorGenJSON(t, "good-older.md", "v-old")
		full := string(priorGenJSON(t, "truncated.md", "v-x"))
		ownWritePriorGeneration(t, root, "20260101_000006", good)
		ownWritePriorGeneration(t, root, "20260101_000005", []byte(full[:len(full)/2])) // partial / truncated
		// A directory occupies the file's name: unreadable as a file.
		if err := os.MkdirAll(filepath.Join(root, ".moai-backups", "20260101_000004", priorManifestName, "inner"), 0o755); err != nil {
			t.Fatal(err)
		}
		ownWritePriorGeneration(t, root, "20260101_000003", nil)                                   // empty file
		ownWritePriorGeneration(t, root, "20260101_000002", []byte(`{"hello": "not a manifest"}`)) // JSON, not a manifest
		ownWritePriorGeneration(t, root, "20260101_000001", older)

		src := newPriorGenerationSource(root, "")
		var gens []priorGeneration
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("a damaged generation panicked the reader: %v", r)
				}
			}()
			gens = src.generations()
		}()
		var names []string
		for _, g := range gens {
			names = append(names, g.name)
		}
		if len(gens) != 2 {
			t.Fatalf("valid generations %v, want exactly the 2 intact ones", names)
		}
		if _, ok := gens[0].files["good-newest.md"]; !ok {
			t.Errorf("newest valid generation missing: %v", priorKeysOf(gens[0].files))
		}
		if _, ok := gens[1].files["good-older.md"]; !ok {
			t.Errorf("older valid generation missing: %v", priorKeysOf(gens[1].files))
		}
		if src.reads != 6 {
			t.Errorf("attempted %d generation reads, want 6 (every retained run backup is tried once)", src.reads)
		}
	})

	t.Run("CurrentRunDirectoryIsExcluded", func(t *testing.T) {
		root := t.TempDir()
		ownWritePriorGeneration(t, root, "20260101_000002", priorGenJSON(t, "cur.md", "v"))
		ownWritePriorGeneration(t, root, "20260101_000001", priorGenJSON(t, "old.md", "v"))
		src := newPriorGenerationSource(root, filepath.Join(root, ".moai-backups", "20260101_000002"))
		gens := src.generations()
		if len(gens) != 1 || gens[0].name != "20260101_000001" {
			t.Errorf("the current run's own copy must not be read as a prior generation: %+v", gens)
		}
	})
}

func priorKeysOf(m map[string]priorRecord) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

// threeManagedCopies registers two more removable managed copies beside the
// fixture's migIdenticalSkill and confirms all three counterparts.
func threeManagedCopies(t *testing.T, root, home string) []string {
	t.Helper()
	rels := []string{migIdenticalSkill, ".claude/skills/moai-workflow-testing/SKILL.md", ownEmbeddedClaudeSkill}
	for _, rel := range rels[1:] {
		ownAddRegisteredManaged(t, root, rel, ownEmbeddedBytes(t, rel))
	}
	for _, rel := range rels {
		requireUserCounterpart(t, home, rel)
	}
	return rels
}

// TestBackupFailureRemovesNothingThroughProductionPath is AC-OWR-20: the
// production Backup stage hands its result to the migration, and when the run
// backup cannot hold the prior-manifest copy (a) or a candidate's bytes (b) the
// run removes no file at all; (c) is the twin that removes all three.
func TestBackupFailureRemovesNothingThroughProductionPath(t *testing.T) {
	setup := func(t *testing.T) (root string, rels []string, before map[string]string) {
		t.Helper()
		enableRemovalForTest(t)
		ownProfileSandbox(t)
		home := installMigrationUserCounterparts(t)
		t.Setenv("HOME", home)
		t.Setenv("USERPROFILE", home)
		root = buildMigrationFixture(t)
		rels = threeManagedCopies(t, root, home)
		before = map[string]string{}
		for _, rel := range rels {
			data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
			if err != nil {
				t.Fatal(err)
			}
			before[rel] = string(data)
		}
		return root, rels, before
	}
	assertAllKept := func(t *testing.T, root string, rels []string, before map[string]string, rc *migrationRunContext, out string) {
		t.Helper()
		if len(rc.removed) != 0 {
			t.Errorf("the run removed %v although its backup failed", rc.removed)
		}
		for _, rel := range rels {
			got, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
			if err != nil || string(got) != before[rel] {
				t.Errorf("%s is gone or changed after a failed-backup run: %v", rel, err)
			}
			row, ok := findRemaining(rc, rel)
			if !ok || row.Reason != keepBackupFailed {
				t.Errorf("%s: remaining row %+v (found=%v), want reason %q", rel, row, ok, keepBackupFailed)
			}
		}
		if !strings.Contains(out, string(keepBackupFailed)) {
			t.Errorf("no output line names the backup failure (%q):\n%s", keepBackupFailed, out)
		}
		mf := loadManifestFile(t, root)
		for _, rel := range rels {
			if _, ok := mf.Files[rel]; !ok {
				t.Errorf("manifest entry for %s was dropped although the file stayed", rel)
			}
		}
		// The rest of the update continues: a template-carried command exists.
		if _, err := os.Stat(filepath.Join(root, ".claude", "commands", "moai", "gtd.md")); err != nil {
			t.Errorf("the update did not continue past the failed backup: %v", err)
		}
	}

	t.Run("a-prior-manifest-copy-fails", func(t *testing.T) {
		root, rels, before := setup(t)
		calls := 0
		orig := priorManifestWriter
		swapPriorManifestWriter(t, func(runDir string, data []byte) error {
			calls++
			blocker := filepath.Join(runDir, priorManifestName)
			if err := os.MkdirAll(blocker, 0o755); err != nil {
				return err
			}
			if err := os.WriteFile(filepath.Join(blocker, "occupant"), []byte("x"), 0o644); err != nil {
				return err
			}
			return orig(runDir, data) // the real writer, now unable to rename onto a non-empty directory
		})
		out, rc := runForcedFlowWithContext(t, root)
		if calls != 1 {
			t.Fatalf("the Backup stage called the prior-manifest writer %d times, want 1 (the flow did not reach the stage)", calls)
		}
		assertAllKept(t, root, rels, before, rc, out)
	})

	t.Run("b-candidate-bytes-copy-fails", func(t *testing.T) {
		root, rels, before := setup(t)
		orig := priorManifestWriter
		swapPriorManifestWriter(t, func(runDir string, data []byte) error {
			// Make ONE candidate's removed-assets slot unwritable: a directory
			// with content stands where the byte copy must be written.
			blocker := filepath.Join(runDir, migrationBackupSubdir, filepath.FromSlash(rels[1]))
			if err := os.MkdirAll(filepath.Join(blocker, "occupant"), 0o755); err != nil {
				return err
			}
			return orig(runDir, data)
		})
		out, rc := runForcedFlowWithContext(t, root)
		if rc.priorManifest != priorManifestWritten {
			t.Fatalf("the prior-manifest copy should have succeeded here (state %v)", rc.priorManifest)
		}
		assertAllKept(t, root, rels, before, rc, out)
	})

	t.Run("c-twin-without-injection-removes-all-three", func(t *testing.T) {
		root, rels, _ := setup(t)
		calls := 0
		orig := priorManifestWriter
		swapPriorManifestWriter(t, func(runDir string, data []byte) error {
			calls++
			return orig(runDir, data)
		})
		_, rc := runForcedFlowWithContext(t, root)
		if calls != 1 {
			t.Fatalf("the Backup stage called the prior-manifest writer %d times, want 1", calls)
		}
		for _, rel := range rels {
			if !slices.Contains(rc.removed, rel) {
				t.Errorf("twin: %s was not removed (removed=%v): the failure arms above do not contrast with a working run", rel, rc.removed)
			}
			if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel))); !os.IsNotExist(err) {
				t.Errorf("twin: %s still on disk (%v)", rel, err)
			}
		}
	})
}

// TestOwnershipAttestedManagedThroughOperatingSource is AC-OWR-05 (h): the
// prior generation is produced by the REAL deployer and manifest.Track, kept by
// the operating writer, read by the operating source, and judged by the verdict
// function. A source that only works when its generations are injected fails
// here.
func TestOwnershipAttestedManagedThroughOperatingSource(t *testing.T) {
	const p = ".claude/commands/moai/operating-source-case.md"
	content := []byte("# a template-carried file\n")

	judge := func(t *testing.T, root string, gateOn bool) ownershipVerdict {
		t.Helper()
		rootResolved, err := filepath.EvalSymlinks(root)
		if err != nil {
			t.Fatal(err)
		}
		mgr := manifest.NewManager()
		if _, err := mgr.Load(root); err != nil {
			t.Fatal(err)
		}
		return classifyOwnership(ownershipEnv{
			entry:                mgr.GetEntry,
			evidence:             newPriorGenerationSource(root, ""),
			counterpartConfirmed: func(string) bool { return true },
			notInCatalog:         func(string) bool { return false },
			removalEnabled:       gateOn,
			rootResolved:         rootResolved,
		}, p, filepath.Join(root, filepath.FromSlash(p)))
	}
	// keepAsPriorAndDropEntry is what a later run sees: the manifest of the run
	// that deployed (or skipped) the file survives only as the previous run's
	// manifest.prior.json, and the current manifest has lost the entry.
	keepAsPriorAndDropEntry := func(t *testing.T, root string) {
		t.Helper()
		raw := readManifestBytes(t, root)
		runDir := filepath.Join(root, ".moai-backups", "20260101_000001")
		if err := os.MkdirAll(runDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := priorManifestWriter(runDir, raw); err != nil {
			t.Fatalf("operating prior-manifest writer: %v", err)
		}
		mgr := manifest.NewManager()
		if _, err := mgr.Load(root); err != nil {
			t.Fatal(err)
		}
		if err := mgr.Remove(p); err != nil {
			t.Fatal(err)
		}
		if err := mgr.Save(); err != nil {
			t.Fatal(err)
		}
	}
	deploy := func(t *testing.T, root, deployingVersion string) manifest.Manager {
		t.Helper()
		mgr := manifest.NewManager()
		if _, err := mgr.Load(root); err != nil {
			t.Fatal(err)
		}
		mgr.SetDeployingVersion(deployingVersion)
		d := template.NewDeployer(fstest.MapFS{p: &fstest.MapFile{Data: content}})
		if err := d.Deploy(context.Background(), root, mgr, nil); err != nil {
			t.Fatalf("real deployer: %v", err)
		}
		mgr.SetDeployingVersion("")
		if err := mgr.Save(); err != nil {
			t.Fatal(err)
		}
		return mgr
	}

	t.Run("h1-existing-untracked-file-is-recorded-user-created-and-stays-unregistered", func(t *testing.T) {
		root := t.TempDir()
		ownWriteFile(t, root, p, content) // the user's byte-identical file exists first
		mgr := deploy(t, root, "v-h1")
		e, ok := mgr.GetEntry(p)
		if !ok || e.Provenance != manifest.UserCreated {
			t.Fatalf("precondition: the real deployer must record the pre-existing file user_created, got %+v (found=%v)", e, ok)
		}
		if e.DeployedVersion != "" {
			t.Errorf("a file the deployer skipped carries DeployedVersion %q (N20: the binary did not write it)", e.DeployedVersion)
		}
		keepAsPriorAndDropEntry(t, root)
		v := judge(t, root, true)
		if v.class != ownershipUnregistered || v.removalEligible || v.attestation != nil {
			t.Fatalf("a byte-identical user file promoted to managed through the operating source: %+v", v)
		}
		if !slices.Contains(v.missing, evidencePriorRecord) {
			t.Errorf("missing evidence %v should name %s", v.missing, evidencePriorRecord)
		}
	})

	t.Run("h2-file-the-run-deployed-is-attested-managed", func(t *testing.T) {
		root := t.TempDir()
		mgr := deploy(t, root, "v-h2")
		e, ok := mgr.GetEntry(p)
		if !ok || e.Provenance != manifest.TemplateManaged || e.DeployedVersion != "v-h2" {
			t.Fatalf("precondition: the real deployer must record template_managed with this run's version, got %+v (found=%v)", e, ok)
		}
		keepAsPriorAndDropEntry(t, root)
		before := readManifestBytes(t, root)
		v := judge(t, root, false)
		if v.class != ownershipManaged || v.attestation == nil {
			t.Fatalf("a file the previous run deployed was not attested managed: %+v", v)
		}
		if v.attestation.DeployedVersion != "v-h2" || v.attestation.OriginalPath != p || v.attestation.Generation != "20260101_000001" {
			t.Errorf("attestation %+v does not name version/path/run backup", *v.attestation)
		}
		if v.removalEligible || v.keep != keepRemovalGateOff {
			t.Errorf("switch off: eligible=%v keep=%q, want keep(%s)", v.removalEligible, v.keep, keepRemovalGateOff)
		}
		if !bytes.Equal(readManifestBytes(t, root), before) {
			t.Errorf("judging an attested file changed the manifest (no auto-adoption is allowed)")
		}
		if on := judge(t, root, true); !on.removalEligible {
			t.Errorf("switch on + confirmed counterpart: %+v, want removal-eligible", on)
		}
	})
}
