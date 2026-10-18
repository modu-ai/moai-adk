package cli

// migrate_ownership_noprofileaccess_test.go — SPEC-UPDATE-OWNERSHIP-REPAIR-001
// M2 (card t1865), AC-OWR-08: the update never opens, reads, enumerates, stats,
// walks, supplies or writes a profile root.
//
// A runtime "was it read" observation does not exist on this tree (no access
// recording seam, atime is not evidence), so the proof is two instruments with
// different blind spots, each shown red on a known input:
//
//   - a symbol-level call-site scan of the update path's whole call closure,
//     judged against the frozen baseline (new (file, symbol) rows and grown
//     counts are red; the one named exception must stay at exactly 1), with a
//     pure scan function so mutant sources can be spliced in and observed red;
//   - a differential canary: the same profile-mapped project updated three
//     times with the profile configuration directory's skills/agents in three
//     states (absent, template-equal, differing) must leave an identical
//     normalized observation record, and a classifier that DOES read the canary
//     must be seen to break that identity.
//
// Named limits (not hidden): a path assembled from run-time fragments that names
// none of the listed symbols is invisible to the scan, and a read whose result
// is discarded leaves no observable difference to the canary. The two together
// do not close that hole; it is accepted debt N16 (card t1702 hold scope).

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/config"
)

// ---------------------------------------------------------------------------
// L1 — the scan judged against the baseline (pure).
// ---------------------------------------------------------------------------

var (
	closureOnce    sync.Once
	closureSources map[string]string
)

// cachedClosureSources lists the update path's call-closure sources once per
// test process and hands out a private copy, so a mutant splice never leaks into
// another test.
func cachedClosureSources(t *testing.T) map[string]string {
	t.Helper()
	closureOnce.Do(func() { closureSources = listUpdateClosureSources(t) })
	if len(closureSources) == 0 {
		t.Fatal("the update call closure came back empty; the scan would judge nothing")
	}
	out := make(map[string]string, len(closureSources))
	for k, v := range closureSources {
		out[k] = v
	}
	return out
}

// profileScanProblems scans files and returns one problem per call-site row that
// is new against the baseline, per count that grew, and for a named-exception
// row that is not exactly 1.
func profileScanProblems(files map[string]string, baseline map[string]int) ([]string, error) {
	counts, err := scanProfileCallsites(files)
	if err != nil {
		return nil, err
	}
	var problems []string
	for k, n := range counts {
		switch w, ok := baseline[k]; {
		case !ok:
			problems = append(problems, fmt.Sprintf("NEW call site %q x%d", k, n))
		case n > w:
			problems = append(problems, fmt.Sprintf("GREW %q: %d > baseline %d", k, n, w))
		}
	}
	exception := ownExceptionFile + "\t" + ownExceptionSymbol
	if got := counts[exception]; got != 1 {
		problems = append(problems, fmt.Sprintf("the named exception %q measured %d times, want exactly 1", exception, got))
	}
	sort.Strings(problems)
	return problems, nil
}

func loadProfileBaseline(t *testing.T) map[string]int {
	t.Helper()
	raw, err := os.ReadFile(ownBaselineRel)
	if err != nil {
		t.Fatalf("read the frozen baseline: %v", err)
	}
	base := parseProfileBaseline(t, string(raw))
	if len(base) == 0 {
		t.Fatal("the frozen baseline holds no rows; the comparison would be vacuous")
	}
	return base
}

// closurePackageDirs is the set of module-relative package directories of the
// `go list -deps` closure of internal/cli (union over the three release GOOS).
func closurePackageDirs(t *testing.T) map[string]bool {
	t.Helper()
	dirs := map[string]bool{}
	for _, goos := range []string{"darwin", "linux", "windows"} {
		out := ownRunGoList(t, goos, "-e", "-deps", "-f", "{{.ImportPath}}", ".")
		for _, imp := range strings.Fields(out) {
			if imp == ownModulePath {
				dirs["."] = true
			} else if rest, ok := strings.CutPrefix(imp, ownModulePath+"/"); ok {
				dirs[rest] = true
			}
		}
	}
	return dirs
}

func TestNoProfileRootAccessInUpdateSources(t *testing.T) {
	files := cachedClosureSources(t)

	t.Run("ScannedSetIsTheGoListClosure", func(t *testing.T) {
		scanned := map[string]bool{}
		for name := range files {
			scanned[path.Dir(name)] = true
		}
		want := closurePackageDirs(t)
		if len(want) == 0 {
			t.Fatal("go list returned no module-internal package; the equality below would be vacuous")
		}
		for d := range want {
			if !scanned[d] {
				t.Errorf("package %s is in the go list closure but was not scanned", d)
			}
		}
		for d := range scanned {
			if !want[d] {
				t.Errorf("package %s was scanned but is not in the go list closure", d)
			}
		}
		for _, must := range []string{"internal/profile", "internal/config", "internal/manifest", "internal/template", "internal/core/project", "internal/userassets"} {
			if !scanned[must] {
				t.Errorf("the scanned closure lacks %s; the scan would be blind to it", must)
			}
		}
	})

	t.Run("NoNewCallSiteAgainstTheBaseline", func(t *testing.T) {
		problems, err := profileScanProblems(files, loadProfileBaseline(t))
		if err != nil {
			t.Fatal(err)
		}
		for _, p := range problems {
			t.Error(p)
		}
	})

	t.Run("MigrationPreviewAndVerdictFilesAddNothing", func(t *testing.T) {
		counts, err := scanProfileCallsites(files)
		if err != nil {
			t.Fatal(err)
		}
		// The code this SPEC adds or edits holds no profile call site at all, and
		// the migration keeps exactly the one baseline row it had before.
		zero := func(file string) bool {
			base := path.Base(file)
			return path.Dir(file) == "internal/cli" &&
				(strings.HasPrefix(base, "migrate_ownership") || base == "update_prior_manifest.go" ||
					base == "update_template_sync.go" || base == "update_dryrun_preview.go")
		}
		got := map[string]int{}
		for k, n := range counts {
			file, _, _ := strings.Cut(k, "\t")
			switch {
			case zero(file):
				t.Errorf("new judgment, sync or preview code carries a profile call site: %q x%d", k, n)
			case file == "internal/cli/migrate_project_assets.go":
				got[k] = n
			}
		}
		want := map[string]int{"internal/cli/migrate_project_assets.go\tprofile.ResolveLaunchProfileForProject": 1}
		if fmt.Sprint(got) != fmt.Sprint(want) {
			t.Errorf("call sites in migrate_project_assets.go: %v, want exactly %v", got, want)
		}
		for _, name := range []string{"internal/cli/migrate_ownership.go", "internal/cli/migrate_ownership_evidence.go", "internal/cli/update_prior_manifest.go", "internal/cli/update_template_sync.go", "internal/cli/update_dryrun_preview.go"} {
			if _, ok := files[name]; !ok {
				t.Errorf("%s is not in the scanned closure; its zero would be vacuous", name)
			}
		}
	})
}

// scanMutant is one spliced-in source change and the row it must make visible.
type scanMutant struct {
	name   string
	apply  func(files map[string]string)
	expect string
}

func appendSource(files map[string]string, name, src string) { files[name] += src }

func TestProfileScanCatchesMutants(t *testing.T) {
	base := loadProfileBaseline(t)
	real := cachedClosureSources(t)
	const migration = "internal/cli/migrate_project_assets.go"
	const profileGo = "internal/profile/profile.go"
	for _, name := range []string{migration, profileGo} {
		if _, ok := real[name]; !ok {
			t.Fatalf("the scanned closure lacks %s; a mutant could not be spliced into it", name)
		}
	}
	const stat = "info, err := os.Stat(filepath.Join(baseDir, name))"
	if !strings.Contains(real[profileGo], stat) {
		t.Fatalf("launchCandidateIsUsable no longer holds %q; the M-c mutant would not fire", stat)
	}

	mutants := []scanMutant{
		{"M-a read-and-discard through an exported helper", func(f map[string]string) {
			appendSource(f, migration, "\nfunc zzMutantA() {\n\t_, _ = os.ReadDir(filepath.Join(profile.GetProfileDir(\"work\"), \"skills\"))\n}\n")
		}, "profile.GetProfileDir"},
		{"M-b spliced token read of the config-dir variable", func(f map[string]string) {
			appendSource(f, migration, "\nfunc zzMutantB() string {\n\treturn os.Getenv(\"CLAUDE_\" + \"CONFIG_DIR\")\n}\n")
		}, "token:CLAUDE_CONFIG_DIR"},
		{"M-c a second os.Stat in launchCandidateIsUsable", func(f map[string]string) {
			f[profileGo] = strings.Replace(f[profileGo], stat, "_, _ = os.Stat(filepath.Join(baseDir, name))\n\t"+stat, 1)
		}, "launchCandidateIsUsable:os.Stat"},
		{"M-d a new function in internal/profile that lists a directory", func(f map[string]string) {
			appendSource(f, profileGo, "\nfunc zzMutantD() {\n\t_, _ = os.ReadDir(\".\")\n}\n")
		}, "fs:zzMutantD:os.ReadDir"},
		{"M-e a new file calling an exported profile helper", func(f map[string]string) {
			f["internal/cli/zz_mutant_e.go"] = "package cli\n\nimport \"github.com/modu-ai/moai-adk/internal/profile\"\n\nfunc zzMutantE() bool { return profile.HasClaudeConfig(\"work\") }\n"
		}, "profile.HasClaudeConfig"},
		{"M-f an environment read with a non-constant name", func(f map[string]string) {
			appendSource(f, migration, "\nfunc zzMutantF(name string) string {\n\treturn os.Getenv(name)\n}\n")
		}, "env:nonconst:os.Getenv"},
	}

	t.Run("UnmutatedTreeIsClean", func(t *testing.T) {
		problems, err := profileScanProblems(real, base)
		if err != nil {
			t.Fatal(err)
		}
		if len(problems) != 0 {
			t.Errorf("the unmutated closure is not clean: %v", problems)
		}
	})

	for _, m := range mutants {
		t.Run(m.name, func(t *testing.T) {
			files := cachedClosureSources(t)
			m.apply(files)
			problems, err := profileScanProblems(files, base)
			if err != nil {
				t.Fatalf("the mutated source did not parse (the mutant is malformed, not caught): %v", err)
			}
			if len(problems) == 0 {
				t.Fatalf("the scan stayed green in front of mutant %q", m.name)
			}
			if !slices.ContainsFunc(problems, func(p string) bool { return strings.Contains(p, m.expect) }) {
				t.Errorf("mutant %q turned the scan red but not on %q: %v", m.name, m.expect, problems)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// L3 — the differential canary with a normalized observation record.
// ---------------------------------------------------------------------------

var (
	canaryRunDirRe = regexp.MustCompile(`\b\d{8}_\d{6}\b`)
	// canaryTimestampRe matches RFC3339 and its space-separated and
	// filesystem-safe (dashes for the time colons) spellings.
	canaryTimestampRe = regexp.MustCompile(`\d{4}-\d{2}-\d{2}(?:[T ]\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:?\d{2})?|T\d{2}-\d{2}-\d{2}Z)`)
)

// canaryNormalizer erases exactly three kinds of run-specific text: the
// temporary project root (<ROOT>), run backup directory names (<RUN>), and
// timestamps including the manifest's deployed_at value (<TS>). Everything else
// is compared byte for byte.
type canaryNormalizer struct{ roots []string }

func newCanaryNormalizer(root string) canaryNormalizer {
	roots := []string{root}
	if resolved, err := filepath.EvalSymlinks(root); err == nil && resolved != root {
		roots = append(roots, resolved)
	}
	sort.Slice(roots, func(i, j int) bool { return len(roots[i]) > len(roots[j]) })
	return canaryNormalizer{roots: roots}
}

func (n canaryNormalizer) apply(s string) string {
	for _, r := range n.roots {
		s = strings.ReplaceAll(s, r, "<ROOT>")
	}
	s = canaryTimestampRe.ReplaceAllString(s, "<TS>")
	return canaryRunDirRe.ReplaceAllString(s, "<RUN>")
}

// canaryObservation is everything one update run leaves that a profile-side read
// could influence: the removal list, the remaining rows, the stored manifest,
// both output streams, the run backup layout and the project tree.
type canaryObservation struct {
	Removed   string
	Remaining string
	Manifest  string
	Out       string
	Err       string
	Backups   string
	Tree      string
}

// canaryDifferences lists, per field, how the observations differ from the first.
func canaryDifferences(obs []canaryObservation) []string {
	var diffs []string
	for i := 1; i < len(obs); i++ {
		for _, f := range []struct {
			name string
			a, b string
		}{
			{"removal list", obs[0].Removed, obs[i].Removed},
			{"remaining rows", obs[0].Remaining, obs[i].Remaining},
			{"stored manifest", obs[0].Manifest, obs[i].Manifest},
			{"stdout", obs[0].Out, obs[i].Out},
			{"stderr", obs[0].Err, obs[i].Err},
			{"backup layout", obs[0].Backups, obs[i].Backups},
			{"project tree", obs[0].Tree, obs[i].Tree},
		} {
			if f.a != f.b {
				diffs = append(diffs, fmt.Sprintf("run %d differs from run 0 in the %s", i, f.name))
			}
		}
	}
	return diffs
}

// canaryTreeDigest hashes the project tree (the run backups are listed
// separately) with each file's text normalized first.
func canaryTreeDigest(t *testing.T, root string, n canaryNormalizer) string {
	t.Helper()
	var lines []string
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, relErr := filepath.Rel(root, p)
		if relErr != nil || rel == "." {
			return relErr
		}
		key := filepath.ToSlash(rel)
		if key == ".moai-backups" || strings.HasPrefix(key, ".moai-backups/") || key == ".git" || strings.HasPrefix(key, ".git/") {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		info, infoErr := os.Lstat(p)
		if infoErr != nil {
			return infoErr
		}
		switch {
		case info.Mode()&os.ModeSymlink != 0:
			target, _ := os.Readlink(p)
			lines = append(lines, "l "+n.apply(key)+" "+n.apply(target))
		case d.IsDir():
			lines = append(lines, "d "+n.apply(key))
		default:
			data, readErr := os.ReadFile(p)
			if readErr != nil {
				return readErr
			}
			sum := sha256.Sum256([]byte(n.apply(string(data))))
			lines = append(lines, fmt.Sprintf("f %s %o %s", n.apply(key), info.Mode().Perm(), hex.EncodeToString(sum[:])))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
	sort.Strings(lines)
	return strings.Join(lines, "\n")
}

// canaryBackupLayout lists the run backup tree relative to its run directory.
func canaryBackupLayout(root string, n canaryNormalizer) string {
	var rels []string
	base := filepath.Join(root, ".moai-backups")
	_ = filepath.WalkDir(base, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(base, p)
		rels = append(rels, n.apply(filepath.ToSlash(rel)))
		return nil
	})
	sort.Strings(rels)
	return strings.Join(rels, "\n")
}

// canaryFixture builds the profile-mapped project of the canary runs: the
// record-bearing migration fixture plus the AC-OWR-07 agents, mapped to the
// profile "work".
func canaryFixture(t *testing.T) string {
	t.Helper()
	root := buildMigrationFixture(t)
	for _, rel := range []string{profAgentA, profAgentB} {
		ownAddRegisteredManaged(t, root, rel, ownEmbeddedBytes(t, rel))
	}
	mapProjectToWork(t, root)
	return root
}

// observedFlow is the update flow driven with --force through the reconciliation
// entry, keeping the migration run context so the removal list and the remaining
// rows are observable. The version-match gate sits in the outer entry only, so
// this entry always reaches reconciliation; the reach assertions below still
// check the Backup stage and the migration were reached.
func observedFlow(t *testing.T, root string) (out, errOut string, rc *migrationRunContext, startManifest []byte) {
	t.Helper()
	ownAssertHomeIsolated(t)
	cmd := &cobra.Command{Use: "update"}
	cmd.Flags().Bool("force", true, "")
	cmd.Flags().Bool("yes", true, "")
	cmd.Flags().Bool("no-hooks", true, "")
	cmd.Flags().Bool("no-plugin", false, "")
	cmd.Flags().Bool("dry-run", false, "")
	cmd.Flags().String("check", "", "")
	var stdout, stderr strings.Builder
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetContext(t.Context())

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	startManifest, _ = os.ReadFile(filepath.Join(root, ".moai", "manifest.json"))
	rc = &migrationRunContext{userAssetsInstalled: true}
	if err := runTemplateSyncWithReporter(cmd, nil, true, rc); err != nil {
		t.Fatalf("runTemplateSyncWithReporter: %v (stderr: %s)", err, stderr.String())
	}
	verifyMigrationFinalState(".", rc, &stderr)
	return stdout.String(), stderr.String(), rc, startManifest
}

// observeCanaryRun runs the flow in a fresh canary project and returns the
// normalized observation plus the reach problems (the run must have reached the
// Backup stage and the migration, and every row must record the profile side as
// not-checked).
func observeCanaryRun(t *testing.T) (canaryObservation, []string) {
	t.Helper()
	root := canaryFixture(t)
	out, errOut, rc, start := observedFlow(t, root)
	n := newCanaryNormalizer(root)

	var reach []string
	priors, _ := filepath.Glob(filepath.Join(root, ".moai-backups", "*", "manifest.prior.json"))
	switch {
	case len(priors) != 1:
		reach = append(reach, fmt.Sprintf("the run left %d manifest.prior.json copies, want 1 (the Backup stage was not reached)", len(priors)))
	default:
		if got, err := os.ReadFile(priors[0]); err != nil || !bytes.Equal(got, start) {
			reach = append(reach, "manifest.prior.json is not the manifest as the run loaded it")
		}
	}
	if len(rc.remaining) == 0 {
		reach = append(reach, "the migration produced no remaining row (the migration was not reached)")
	}
	var rows []string
	for _, r := range rc.remaining {
		if r.ProfileSide != profileSideNotChecked {
			reach = append(reach, fmt.Sprintf("%s: profile side %q, want %q", r.Rel, r.ProfileSide, profileSideNotChecked))
		}
		rows = append(rows, fmt.Sprintf("%s|%s|%s|%v|%s|%s", r.Rel, r.Class, r.Reason, r.Missing, r.Counterpart, r.ProfileSide))
	}
	for _, rel := range profileMappedRels {
		if _, ok := findRemaining(rc, rel); !ok {
			reach = append(reach, fmt.Sprintf("%s has no remaining row", rel))
		}
	}
	sort.Strings(rows)
	removed := append([]string(nil), rc.removed...)
	sort.Strings(removed)

	manifestBytes, _ := os.ReadFile(filepath.Join(root, ".moai", "manifest.json"))
	return canaryObservation{
		Removed:   strings.Join(removed, "\n"),
		Remaining: strings.Join(rows, "\n"),
		Manifest:  n.apply(string(manifestBytes)),
		Out:       n.apply(out),
		Err:       n.apply(errOut),
		Backups:   canaryBackupLayout(root, n),
		Tree:      canaryTreeDigest(t, root, n),
	}, reach
}

// canaryStates are the three states of the profile configuration directory's
// skills/ and agents/ the canary cycles through.
var canaryStates = []struct {
	name  string
	state ownCanaryState
}{
	{"absent", ownCanaryNone},
	{"template-equal", ownCanaryTemplateSame},
	{"differing", ownCanaryDiffers},
}

// observeAcrossCanaryStates runs one canary project per profile-directory state
// (each with CLAUDE_CONFIG_DIR pointed at the canary directory) and returns the
// observations in the order of canaryStates.
func observeAcrossCanaryStates(t *testing.T, base string, states []ownCanaryState) ([]canaryObservation, []string) {
	t.Helper()
	var obs []canaryObservation
	var reach []string
	for _, st := range states {
		dir := ownCanaryProfileDir(t, base, st)
		ownSetClaudeConfigDir(t, dir)
		o, r := observeCanaryRun(t)
		obs = append(obs, o)
		reach = append(reach, r...)
		for _, sub := range []string{"skills", "agents"} {
			if err := os.RemoveAll(filepath.Join(dir, sub)); err != nil {
				t.Fatal(err)
			}
		}
	}
	return obs, reach
}

func TestProfileEnvDoesNotReachProfileRoot(t *testing.T) {
	base := ownProfileSandbox(t)
	home := t.TempDir() // an empty user root: the same for every run
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	t.Run("NormalizationRules", func(t *testing.T) {
		root := t.TempDir()
		n := newCanaryNormalizer(root)
		resolved, _ := filepath.EvalSymlinks(root)
		raw := strings.Join([]string{
			"project " + root + "/.claude/settings.json",
			"resolved " + resolved + "/.moai/manifest.json",
			"run .moai-backups/20261018_153045/removed-assets/x",
			`"deployed_at": "2026-10-18T15:30:45.123456+09:00"`,
			"stamp 2026-10-18 15:30:45",
			// Text that must survive untouched.
			"version v3.2.0-rc.36 digits 20261018 count 313",
			"elsewhere /etc/hosts and sha256:" + strings.Repeat("ab", 32),
			"canary " + filepath.Join(base, "work", "skills", "x", "SKILL.md"),
		}, "\n")
		got := n.apply(raw)
		for _, gone := range []string{root, "20261018_153045", "15:30:45"} {
			if strings.Contains(got, gone) {
				t.Errorf("the normalizer left %q in place", gone)
			}
		}
		for _, kept := range []string{"v3.2.0-rc.36 digits 20261018 count 313", "/etc/hosts", "sha256:" + strings.Repeat("ab", 32), filepath.Join(base, "work", "skills", "x", "SKILL.md")} {
			if !strings.Contains(got, kept) {
				t.Errorf("the normalizer erased %q, which is not one of the three kinds", kept)
			}
		}
		for _, marker := range []string{"<ROOT>", "<RUN>", "<TS>"} {
			if !strings.Contains(got, marker) {
				t.Errorf("the normalizer produced no %s marker", marker)
			}
		}
		if again := n.apply(got); again != got {
			t.Errorf("normalizing twice changed the result:\n%s\nvs\n%s", got, again)
		}
	})

	t.Run("ControlTwoRunsOfTheSameStateAreIdentical", func(t *testing.T) {
		obs, reach := observeAcrossCanaryStates(t, base, []ownCanaryState{ownCanaryNone, ownCanaryNone})
		for _, r := range reach {
			t.Error(r)
		}
		for _, d := range canaryDifferences(obs) {
			t.Errorf("the normalization does not make two identical-state runs equal (the three-state comparison below would judge noise): %s", d)
		}
	})

	t.Run("ThreeProfileStatesLeaveTheSameRecord", func(t *testing.T) {
		states := make([]ownCanaryState, len(canaryStates))
		for i, s := range canaryStates {
			states[i] = s.state
		}
		obs, reach := observeAcrossCanaryStates(t, base, states)
		for _, r := range reach {
			t.Error(r)
		}
		if len(obs) != len(canaryStates) {
			t.Fatalf("observed %d runs, want %d", len(obs), len(canaryStates))
		}
		for _, d := range canaryDifferences(obs) {
			t.Error(d)
		}
	})

	t.Run("MutantReadingTheCanaryIsSeen", func(t *testing.T) {
		swapClassifier(t, func(env ownershipEnv, rel, abs string) ownershipVerdict {
			v := classifyOwnership(env, rel, abs)
			// The wrong implementation: look into the profile configuration
			// directory and let what it finds decide the recorded profile side.
			data, err := os.ReadFile(filepath.Join(os.Getenv(config.EnvClaudeConfigDir), "skills", "x", "SKILL.md"))
			switch {
			case err != nil:
				v.profileSide = profileSideStatus("absent")
			case bytes.Equal(data, ownEmbeddedBytes(t, ownEmbeddedClaudeSkill)):
				v.profileSide = profileSideStatus("confirmed")
			default:
				v.profileSide = profileSideStatus("unconfirmed")
			}
			return v
		})
		states := make([]ownCanaryState, len(canaryStates))
		for i, s := range canaryStates {
			states[i] = s.state
		}
		obs, _ := observeAcrossCanaryStates(t, base, states)
		if len(canaryDifferences(obs)) == 0 {
			t.Errorf("the canary stayed identical in front of a classifier that reads the profile directory")
		}
	})
}
