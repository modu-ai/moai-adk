package cli

// migrate_ownership_helpers_test.go — SPEC-UPDATE-OWNERSHIP-REPAIR-001 M0
// (card t1865): the shared fixtures and helpers every later ownership
// criterion builds on, and the frozen profile call-site baseline.
//
//   - Fixtures use REAL embedded template paths (never invented ones) so a
//     byte-equal-to-the-template mutant can actually fire.
//   - runUpdateFlowForced drives the whole update flow with --force so the run
//     reaches reconciliation; a same-version plain re-run is skipped before
//     reconciliation (update_template_sync.go:1239-1248), so every
//     multi-run criterion goes through it and asserts it reached.
//   - ownTreeSnapshot folds directory mtimes into the tree digest so a
//     "create then delete" temporary write is visible.
//   - The profile call-site scanner and its baseline are the measurement
//     AC-OWR-08 later judges against.
//
// Every helper confines itself to t.TempDir() and a temporary HOME. None
// reads or writes the real home.

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/profile"
	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/internal/userassets"
)

// ---------------------------------------------------------------------------
// Real embedded template paths.
// ---------------------------------------------------------------------------

// The two paths AC-OWR-02 plants as unregistered byte-equal files. Both exist
// in the embedded template tree of this build (acceptance.md § L-1); every use
// asserts the lookup first so a path that stops existing fails loudly instead
// of letting a mutant slip through unfired.
const (
	ownEmbeddedClaudeSkill = ".claude/skills/moai-ref-owasp-checklist/SKILL.md"
	ownEmbeddedAgentsSkill = ".agents/skills/moai-plan/SKILL.md"
)

// ownEmbeddedBytes returns the embedded template bytes at rel, failing the
// test when the path is not in the embedded tree.
func ownEmbeddedBytes(t *testing.T, rel string) []byte {
	t.Helper()
	embedded, err := template.EmbeddedTemplates()
	if err != nil {
		t.Fatalf("load embedded templates: %v", err)
	}
	data, err := fs.ReadFile(embedded, rel)
	if err != nil {
		t.Fatalf("embedded template lookup failed for %s (the fixture would not exercise a byte-equal file): %v", rel, err)
	}
	return data
}

// ---------------------------------------------------------------------------
// File-state builders (registered / modified / user-created / unregistered).
// ---------------------------------------------------------------------------

// ownWriteFile writes data at root/rel, creating parent directories.
func ownWriteFile(t *testing.T, root, rel string, data []byte) {
	t.Helper()
	abs := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatalf("mkdir for %s: %v", rel, err)
	}
	if err := os.WriteFile(abs, data, 0o644); err != nil {
		t.Fatalf("write %s: %v", rel, err)
	}
}

// ownTrack records rel in the project manifest with the given provenance and
// template hash (the file must exist: Track hashes the current bytes).
func ownTrack(t *testing.T, root, rel string, prov manifest.Provenance, templateHash string) {
	t.Helper()
	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	if err := mgr.Track(rel, prov, templateHash); err != nil {
		t.Fatalf("track %s: %v", rel, err)
	}
	if err := mgr.Save(); err != nil {
		t.Fatalf("save manifest: %v", err)
	}
}

// ownAddRegisteredManaged plants rel with data and registers it template_managed
// with a matching template hash — the "healthy managed copy" state.
func ownAddRegisteredManaged(t *testing.T, root, rel string, data []byte) {
	t.Helper()
	ownWriteFile(t, root, rel, data)
	ownTrack(t, root, rel, manifest.TemplateManaged, manifest.HashBytes(data))
}

// ownAddRegisteredEdited reproduces the real "registered, then modified" order:
// the pristine template bytes are written and tracked first (so the recorded
// deployed and current hashes are the pristine hash), then the user's appended
// text lands on disk. prov=TemplateManaged is the stale-record shape,
// prov=UserModified the explicitly recorded one.
func ownAddRegisteredEdited(t *testing.T, root, rel string, template []byte, appended string, prov manifest.Provenance) {
	t.Helper()
	ownWriteFile(t, root, rel, template)
	ownTrack(t, root, rel, prov, manifest.HashBytes(template))
	ownWriteFile(t, root, rel, append(append([]byte{}, template...), []byte(appended)...))
}

// ownAddRegisteredUserCreated plants rel with data and registers it user_created.
func ownAddRegisteredUserCreated(t *testing.T, root, rel string, data []byte) {
	t.Helper()
	ownWriteFile(t, root, rel, data)
	ownTrack(t, root, rel, manifest.UserCreated, manifest.HashBytes(data))
}

// ownAddUnregistered plants rel with data and leaves it out of the manifest.
func ownAddUnregistered(t *testing.T, root, rel string, data []byte) {
	t.Helper()
	ownWriteFile(t, root, rel, data)
	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	if _, found := mgr.GetEntry(rel); found {
		t.Fatalf("fixture error: %s is registered, it must stay unregistered", rel)
	}
}

// ownAddUnregisteredFromTemplate plants an unregistered file whose bytes are
// exactly the embedded template's bytes at the same path, and returns them.
func ownAddUnregisteredFromTemplate(t *testing.T, root, rel string) []byte {
	t.Helper()
	data := ownEmbeddedBytes(t, rel)
	ownAddUnregistered(t, root, rel, data)
	return data
}

// requireUserCounterpart asserts the user-root counterpart of project path rel
// is confirmed in home (installed, tracked, hash-equal, current content). The
// default user-asset phase already installs the common set, so this is the
// path-specific precondition check rather than a second installer: a fixture
// that thinks a removal is possible fails here when it is not.
func requireUserCounterpart(t *testing.T, home, rel string) {
	t.Helper()
	embedded, err := template.EmbeddedTemplates()
	if err != nil {
		t.Fatalf("load embedded templates: %v", err)
	}
	cat, err := template.LoadEmbeddedCatalog()
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}
	um, err := userassets.Load(userassets.ManifestPath(home))
	if err != nil {
		t.Fatalf("load user manifest: %v", err)
	}
	if !userCounterpartConfirmed(um, cat, embedded, home, rel) {
		t.Fatalf("precondition: the user counterpart of %s is not confirmed in the test home", rel)
	}
}

// ---------------------------------------------------------------------------
// Prior-generation manifest builder (JSON, no dependency on the schema).
// ---------------------------------------------------------------------------

// ownPriorRecord is one path's record in a prior manifest generation. The
// builder writes plain JSON so M0 compiles before the deployed_version field
// exists in the manifest package.
type ownPriorRecord struct {
	Provenance      manifest.Provenance
	TemplateHash    string
	DeployedHash    string
	DeployedVersion string // empty: the key is omitted (an old-schema entry)
}

// ownPriorGenerationJSON renders a prior manifest generation exactly as a
// manifest file would carry it, using the real provenance constants.
func ownPriorGenerationJSON(t *testing.T, version, deployedAt string, files map[string]ownPriorRecord) []byte {
	t.Helper()
	entries := make(map[string]any, len(files))
	for rel, r := range files {
		e := map[string]any{
			"provenance":    string(r.Provenance),
			"template_hash": r.TemplateHash,
			"deployed_hash": r.DeployedHash,
			"current_hash":  r.DeployedHash,
		}
		if r.DeployedVersion != "" {
			e["deployed_version"] = r.DeployedVersion
		}
		entries[rel] = e
	}
	data, err := json.MarshalIndent(map[string]any{
		"version":     version,
		"deployed_at": deployedAt,
		"files":       entries,
	}, "", "  ")
	if err != nil {
		t.Fatalf("marshal prior generation: %v", err)
	}
	return append(data, '\n')
}

// ownWritePriorGeneration writes a prior generation at
// <projectRoot>/.moai-backups/<runName>/manifest.prior.json and returns the
// run directory.
func ownWritePriorGeneration(t *testing.T, projectRoot, runName string, data []byte) string {
	t.Helper()
	runDir := filepath.Join(projectRoot, ".moai-backups", runName)
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		t.Fatalf("mkdir run dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(runDir, "manifest.prior.json"), data, 0o644); err != nil {
		t.Fatalf("write prior generation: %v", err)
	}
	return runDir
}

// ---------------------------------------------------------------------------
// Tree snapshot with directory mtimes.
// ---------------------------------------------------------------------------

// ownTreeOpts selects what a snapshot folds in.
type ownTreeOpts struct {
	// SkipPrefixes are slash-relative path prefixes left out of the snapshot
	// (for example ".moai-backups/" to ignore run-backup directories).
	SkipPrefixes []string
	// DirMtime folds each directory's mtime into its entry, so a temporary file
	// created and deleted inside a directory changes the snapshot.
	DirMtime bool
}

// ownTreeSnapshot returns slash-relative path -> descriptor for every entry
// under root (files: mode and content digest; directories: mode and, with
// DirMtime, mtime in nanoseconds; symlinks: target).
func ownTreeSnapshot(t *testing.T, root string, opts ownTreeOpts) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, relErr := filepath.Rel(root, p)
		if relErr != nil {
			return relErr
		}
		if rel == "." {
			// The root directory itself: a file created and deleted directly in
			// it (a probe file) moves its mtime and nothing else.
			if opts.DirMtime {
				info, infoErr := os.Lstat(p)
				if infoErr != nil {
					return infoErr
				}
				out["./"] = fmt.Sprintf("d %o %d", info.Mode().Perm(), info.ModTime().UnixNano())
			}
			return nil
		}
		key := filepath.ToSlash(rel)
		for _, pre := range opts.SkipPrefixes {
			if key+"/" == pre || strings.HasPrefix(key+"/", pre) {
				if d.IsDir() {
					return fs.SkipDir
				}
				return nil
			}
		}
		info, infoErr := os.Lstat(p)
		if infoErr != nil {
			return infoErr
		}
		switch {
		case info.Mode()&os.ModeSymlink != 0:
			target, _ := os.Readlink(p)
			out[key] = "l " + target
		case d.IsDir():
			desc := fmt.Sprintf("d %o", info.Mode().Perm())
			if opts.DirMtime {
				desc += fmt.Sprintf(" %d", info.ModTime().UnixNano())
			}
			out[key+"/"] = desc
		default:
			data, readErr := os.ReadFile(p)
			if readErr != nil {
				return readErr
			}
			sum := sha256.Sum256(data)
			out[key] = fmt.Sprintf("f %o %s", info.Mode().Perm(), hex.EncodeToString(sum[:]))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("snapshot %s: %v", root, err)
	}
	return out
}

// ownTreeDigest folds a snapshot into one comparable string.
func ownTreeDigest(snap map[string]string) string {
	keys := make([]string, 0, len(snap))
	for k := range snap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	h := sha256.New()
	for _, k := range keys {
		_, _ = fmt.Fprintf(h, "%s\x00%s\n", k, snap[k])
	}
	return hex.EncodeToString(h.Sum(nil))
}

// ownTreeDiff lists the paths whose descriptors differ between two snapshots
// (added, removed, or changed), sorted.
func ownTreeDiff(a, b map[string]string) []string {
	seen := map[string]bool{}
	var diff []string
	for k, v := range a {
		seen[k] = true
		if bv, ok := b[k]; !ok || bv != v {
			diff = append(diff, k)
		}
	}
	for k := range b {
		if !seen[k] {
			diff = append(diff, k)
		}
	}
	sort.Strings(diff)
	return diff
}

// pinTreeDirMtimes sets every directory under root (root included) to a fixed
// past timestamp, so a later create-then-delete inside any of them moves its
// mtime off the pin.
func pinTreeDirMtimes(t *testing.T, root string, when time.Time) {
	t.Helper()
	var dirs []string
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			dirs = append(dirs, p)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
	// Deepest first: touching a child directory never changes its parent's
	// mtime, but ordering keeps the pass independent of that fact.
	sort.Slice(dirs, func(i, j int) bool { return len(dirs[i]) > len(dirs[j]) })
	for _, d := range dirs {
		if err := os.Chtimes(d, when, when); err != nil {
			t.Fatalf("pin mtime of %s: %v", d, err)
		}
	}
}

// ---------------------------------------------------------------------------
// Update-flow drivers.
// ---------------------------------------------------------------------------

// ownIsolateHome points HOME and USERPROFILE at a fresh temporary directory
// and returns it. The migration reads the user manifest from the home the
// flow resolves, so a flow driven without this would read the real one.
func ownIsolateHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	return home
}

// ownAssertHomeIsolated fails the test unless the process HOME is inside the
// system temporary directory — the guard that keeps a flow from touching the
// developer's real home.
func ownAssertHomeIsolated(t *testing.T) {
	t.Helper()
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("resolve HOME: %v", err)
	}
	tmpRoot, err := filepath.EvalSymlinks(os.TempDir())
	if err != nil {
		t.Fatalf("resolve temp dir: %v", err)
	}
	homeResolved, err := filepath.EvalSymlinks(home)
	if err != nil {
		t.Fatalf("resolve HOME %s: %v", home, err)
	}
	if !strings.HasPrefix(homeResolved+string(filepath.Separator), tmpRoot+string(filepath.Separator)) {
		t.Fatalf("HOME %s is outside the temporary directory %s: the flow would touch the real home", homeResolved, tmpRoot)
	}
}

// forcedRun is the observation of one full update flow.
type forcedRun struct {
	// Out is the stdout of the flow.
	Out string
	// Skipped is the flow's skipped result: true when the version-match gate
	// ended the run before reconciliation.
	Skipped bool
	// NewPriorManifest is the path of a manifest.prior.json written during this
	// run (an mtime at or after the run's start), or "" when none was.
	NewPriorManifest string
	// StartManifest is the manifest file's bytes immediately before the run.
	StartManifest []byte
}

// runUpdateFlow drives the real update flow (runTemplateSyncWithProgress, the
// entry `moai update` uses) in root with --yes and the given --force value.
func runUpdateFlow(t *testing.T, root string, force bool) forcedRun {
	t.Helper()
	ownAssertHomeIsolated(t)
	cmd := &cobra.Command{Use: "update"}
	cmd.Flags().Bool("force", force, "")
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
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	start, _ := os.ReadFile(filepath.Join(root, ".moai", "manifest.json"))
	began := time.Now().Add(-time.Second)
	skipped, runErr := runTemplateSyncWithProgress(cmd, true)
	if runErr != nil {
		t.Fatalf("runTemplateSyncWithProgress(force=%v): %v", force, runErr)
	}
	run := forcedRun{Out: out.String(), Skipped: skipped, StartManifest: start}
	matches, _ := filepath.Glob(filepath.Join(root, ".moai-backups", "*", "manifest.prior.json"))
	for _, m := range matches {
		if info, statErr := os.Stat(m); statErr == nil && info.ModTime().After(began) {
			run.NewPriorManifest = m
		}
	}
	return run
}

// runUpdateFlowForced drives the flow with --force and asserts the run was NOT
// skipped by the version-match gate — the reach assertion that keeps a
// multi-run criterion from reading a skipped run as an executed one.
func runUpdateFlowForced(t *testing.T, root string) forcedRun {
	t.Helper()
	run := runUpdateFlow(t, root, true)
	if run.Skipped {
		t.Fatalf("a --force run was skipped before reconciliation; the criterion would judge nothing")
	}
	return run
}

// runUpdateFlowPlain drives the flow without --force and returns the
// observation as it is, skipped or not.
func runUpdateFlowPlain(t *testing.T, root string) forcedRun {
	t.Helper()
	return runUpdateFlow(t, root, false)
}

// ---------------------------------------------------------------------------
// Profile sandbox and canary.
// ---------------------------------------------------------------------------

// ownProfileSandbox confines HOME and the launch ledger to test-owned
// directories (profileReachSandbox), clears CLAUDE_CONFIG_DIR, and returns the
// profile base directory.
func ownProfileSandbox(t *testing.T) string {
	t.Helper()
	profileReachSandbox(t)
	return profile.BaseDirOverride
}

// ownSetClaudeConfigDir points CLAUDE_CONFIG_DIR at dir for the test.
func ownSetClaudeConfigDir(t *testing.T, dir string) {
	t.Helper()
	t.Setenv(config.EnvClaudeConfigDir, dir)
}

// ownCanaryState is the content state of a canary profile directory.
type ownCanaryState int

const (
	ownCanaryNone         ownCanaryState = iota // skills/ and agents/ do not exist
	ownCanaryTemplateSame                       // counterparts carry the embedded template's bytes
	ownCanaryDiffers                            // counterparts carry different bytes
)

// ownCanaryProfileDir builds <base>/work with skills/x/SKILL.md and agents/y.md
// in the given state and returns the profile directory. The canary is a
// DIFFERENTIAL probe: the update flow's observable record must not depend on
// which state it holds.
func ownCanaryProfileDir(t *testing.T, base string, state ownCanaryState) string {
	t.Helper()
	dir := filepath.Join(base, "work")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir canary profile: %v", err)
	}
	switch state {
	case ownCanaryNone:
	case ownCanaryTemplateSame:
		ownWriteFile(t, dir, "skills/x/SKILL.md", ownEmbeddedBytes(t, ownEmbeddedClaudeSkill))
		ownWriteFile(t, dir, "agents/y.md", ownEmbeddedBytes(t, ownEmbeddedClaudeSkill))
	case ownCanaryDiffers:
		ownWriteFile(t, dir, "skills/x/SKILL.md", []byte("canary: differing skill bytes\n"))
		ownWriteFile(t, dir, "agents/y.md", []byte("canary: differing agent bytes\n"))
	default:
		t.Fatalf("unknown canary state %d", state)
	}
	return dir
}

// ---------------------------------------------------------------------------
// Profile call-site scanner and baseline (AC-OWR-08's measurement).
// ---------------------------------------------------------------------------

const (
	ownModulePath  = "github.com/modu-ai/moai-adk"
	ownProfilePkg  = ownModulePath + "/internal/profile"
	ownBaselineRel = "testdata/profile_callsite_baseline.txt"
	// ownExceptionSymbol is the one named existing contact REQ-OWR-005
	// tolerates (leader ruling OQ-5): launchCandidateIsUsable's os.Stat.
	ownExceptionFile   = "internal/profile/profile.go"
	ownExceptionSymbol = "fs:launchCandidateIsUsable:os.Stat"
)

// ownProfileTokens are the string tokens that name a profile configuration
// directory; they are compared after constant folding.
var ownProfileTokens = []string{"CLAUDE_CONFIG_DIR", "claude-profiles", "claude-skills@", "claude-agents@"}

// ownProfileIdents are identifiers that produce or select a profile path.
var ownProfileIdents = map[string]bool{
	"GetBaseDir": true, "BaseDirOverride": true, "EnvClaudeConfigDir": true,
	"profilesDir": true, "GetCurrentName": true,
}

// ownFSVerbs are the filesystem verbs counted per function inside
// internal/profile. "fs" (io/fs) counts every selector call.
var ownFSVerbs = map[string]map[string]bool{
	"os": {
		"Stat": true, "Lstat": true, "ReadDir": true, "Open": true, "OpenFile": true,
		"ReadFile": true, "WriteFile": true, "MkdirAll": true, "Mkdir": true,
		"Remove": true, "RemoveAll": true, "Rename": true, "Create": true,
		"CreateTemp": true, "MkdirTemp": true, "Symlink": true, "Readlink": true,
		"Chmod": true, "Chtimes": true,
	},
	"filepath": {"Walk": true, "WalkDir": true, "Glob": true, "EvalSymlinks": true},
}

// ownFoldString folds a string-literal expression (literals, parentheses, and
// '+' concatenation) to its value.
func ownFoldString(e ast.Expr) (string, bool) {
	switch n := e.(type) {
	case *ast.BasicLit:
		if n.Kind != token.STRING {
			return "", false
		}
		v, err := strconv.Unquote(n.Value)
		return v, err == nil
	case *ast.ParenExpr:
		return ownFoldString(n.X)
	case *ast.BinaryExpr:
		if n.Op != token.ADD {
			return "", false
		}
		l, lok := ownFoldString(n.X)
		r, rok := ownFoldString(n.Y)
		if lok && rok {
			return l + r, true
		}
	}
	return "", false
}

// ownImportAliases maps each import path of f to the name it is bound to.
func ownImportAliases(f *ast.File) map[string]string {
	out := map[string]string{}
	for _, spec := range f.Imports {
		p, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			continue
		}
		name := path.Base(p)
		if spec.Name != nil {
			if spec.Name.Name == "_" || spec.Name.Name == "." {
				continue
			}
			name = spec.Name.Name
		}
		out[p] = name
	}
	return out
}

// scanProfileCallsites counts, per (file, symbol), the profile-touching call
// sites in the given non-test sources (slash-relative path -> source). It is a
// pure function over its input so a mutant source can be spliced into the map.
//
// Counted symbol classes:
//
//	profile.<X>            selector uses of every exported identifier of internal/profile
//	token:<T>              a string expression (after '+' folding) containing a profile-directory token
//	ident:<I>              an identifier that produces or selects a profile path
//	env:nonconst:os.<F>    os.Getenv / os.LookupEnv with a non-constant name, and os.Environ
//	fs:<Func>:<alias>.<V>  filesystem verbs, per function, inside internal/profile only
//
// Limits (stated, not hidden): a path built from run-time fragments that names
// none of the symbols above is not seen; an argument that is an imported
// identifier is assumed to be a constant; build-tagged files outside the
// scanned set are not seen.
func scanProfileCallsites(files map[string]string) (map[string]int, error) {
	fset := token.NewFileSet()
	parsed := make(map[string]*ast.File, len(files))
	for name, src := range files {
		f, err := parser.ParseFile(fset, name, src, 0)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", name, err)
		}
		parsed[name] = f
	}

	// Exported top-level identifiers of internal/profile.
	exported := map[string]bool{}
	for name, f := range parsed {
		rest, ok := strings.CutPrefix(name, "internal/profile/")
		if !ok || strings.Contains(rest, "/") {
			continue
		}
		for _, d := range f.Decls {
			switch decl := d.(type) {
			case *ast.FuncDecl:
				if decl.Recv == nil && decl.Name.IsExported() {
					exported[decl.Name.Name] = true
				}
			case *ast.GenDecl:
				for _, spec := range decl.Specs {
					switch s := spec.(type) {
					case *ast.TypeSpec:
						if s.Name.IsExported() {
							exported[s.Name.Name] = true
						}
					case *ast.ValueSpec:
						for _, n := range s.Names {
							if n.IsExported() {
								exported[n.Name] = true
							}
						}
					}
				}
			}
		}
	}

	// Constant names declared anywhere in a directory's package.
	consts := map[string]map[string]bool{}
	for name, f := range parsed {
		dir := path.Dir(name)
		if consts[dir] == nil {
			consts[dir] = map[string]bool{}
		}
		ast.Inspect(f, func(n ast.Node) bool {
			if gd, ok := n.(*ast.GenDecl); ok && gd.Tok == token.CONST {
				for _, spec := range gd.Specs {
					if vs, ok := spec.(*ast.ValueSpec); ok {
						for _, id := range vs.Names {
							consts[dir][id.Name] = true
						}
					}
				}
			}
			return true
		})
	}

	counts := map[string]int{}
	add := func(file, sym string) { counts[file+"\t"+sym]++ }

	for name, f := range parsed {
		aliases := ownImportAliases(f)
		profileAlias := aliases[ownProfilePkg]
		osAlias := aliases["os"]
		importNames := map[string]bool{}
		for _, a := range aliases {
			importNames[a] = true
		}
		dir := path.Dir(name)
		inProfile := strings.HasPrefix(name, "internal/profile/")

		isConstArg := func(e ast.Expr) bool {
			for {
				p, ok := e.(*ast.ParenExpr)
				if !ok {
					break
				}
				e = p.X
			}
			if _, ok := ownFoldString(e); ok {
				return true
			}
			switch n := e.(type) {
			case *ast.Ident:
				return consts[dir][n.Name]
			case *ast.SelectorExpr:
				if x, ok := n.X.(*ast.Ident); ok {
					return importNames[x.Name]
				}
			}
			return false
		}

		for _, decl := range f.Decls {
			funcName := "<package>"
			if fd, ok := decl.(*ast.FuncDecl); ok {
				funcName = fd.Name.Name
				if fd.Recv != nil && len(fd.Recv.List) == 1 {
					switch rt := fd.Recv.List[0].Type.(type) {
					case *ast.StarExpr:
						if id, ok := rt.X.(*ast.Ident); ok {
							funcName = id.Name + "." + funcName
						}
					case *ast.Ident:
						funcName = rt.Name + "." + funcName
					}
				}
			}
			ast.Inspect(decl, func(n ast.Node) bool {
				if expr, ok := n.(ast.Expr); ok {
					if v, folded := ownFoldString(expr); folded {
						for _, tok := range ownProfileTokens {
							if strings.Contains(v, tok) {
								add(name, "token:"+tok)
							}
						}
						return false
					}
				}
				switch node := n.(type) {
				case *ast.Ident:
					if ownProfileIdents[node.Name] {
						add(name, "ident:"+node.Name)
					}
				case *ast.SelectorExpr:
					if x, ok := node.X.(*ast.Ident); ok && profileAlias != "" && x.Name == profileAlias && exported[node.Sel.Name] {
						add(name, "profile."+node.Sel.Name)
					}
				case *ast.CallExpr:
					sel, ok := node.Fun.(*ast.SelectorExpr)
					if !ok {
						break
					}
					x, ok := sel.X.(*ast.Ident)
					if !ok {
						break
					}
					if osAlias != "" && x.Name == osAlias {
						switch sel.Sel.Name {
						case "Getenv", "LookupEnv":
							if len(node.Args) == 1 && !isConstArg(node.Args[0]) {
								add(name, "env:nonconst:os."+sel.Sel.Name)
							}
						case "Environ":
							add(name, "env:nonconst:os.Environ")
						}
					}
					if inProfile {
						fsAlias := aliases["io/fs"]
						if verbs, known := ownFSVerbs[x.Name]; known && verbs[sel.Sel.Name] {
							add(name, "fs:"+funcName+":"+x.Name+"."+sel.Sel.Name)
						} else if fsAlias != "" && x.Name == fsAlias {
							add(name, "fs:"+funcName+":"+x.Name+"."+sel.Sel.Name)
						}
					}
				}
				return true
			})
		}
	}
	return counts, nil
}

// ownRunGoList runs `go list` in the test's package directory for one GOOS.
func ownRunGoList(t *testing.T, goos string, args ...string) string {
	t.Helper()
	cmd := exec.Command("go", append([]string{"list"}, args...)...)
	cmd.Env = append(os.Environ(), "GOOS="+goos, "CGO_ENABLED=0")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("go list (GOOS=%s) %v: %v\n%s", goos, args, err, stderr.String())
	}
	return string(out)
}

// listUpdateClosureSources derives the update path's call closure from
// `go list -deps` (the union over darwin, linux and windows so the baseline
// does not depend on the host OS) and returns every non-test .go file of every
// module-internal package in it, keyed by module-relative slash path.
func listUpdateClosureSources(t *testing.T) map[string]string {
	t.Helper()
	moduleDir := strings.TrimSpace(ownRunGoList(t, "linux", "-m", "-f", "{{.Dir}}"))
	if moduleDir == "" {
		t.Fatal("go list -m returned no module directory")
	}
	pkgDirs := map[string]string{}
	for _, goos := range []string{"darwin", "linux", "windows"} {
		out := ownRunGoList(t, goos, "-e", "-deps", "-f", "{{.ImportPath}}\t{{.Dir}}", ".")
		for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
			imp, dir, ok := strings.Cut(line, "\t")
			if !ok || dir == "" {
				continue
			}
			if imp == ownModulePath || strings.HasPrefix(imp, ownModulePath+"/") {
				pkgDirs[imp] = dir
			}
		}
	}
	if len(pkgDirs) == 0 {
		t.Fatal("the update call closure came back empty; the scan would judge nothing")
	}
	files := map[string]string{}
	for imp, dir := range pkgDirs {
		rel := strings.TrimPrefix(strings.TrimPrefix(imp, ownModulePath), "/")
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatalf("read %s: %v", dir, err)
		}
		for _, e := range entries {
			n := e.Name()
			if e.IsDir() || !strings.HasSuffix(n, ".go") || strings.HasSuffix(n, "_test.go") {
				continue
			}
			data, err := os.ReadFile(filepath.Join(dir, n))
			if err != nil {
				t.Fatalf("read %s: %v", n, err)
			}
			key := n
			if rel != "" {
				key = rel + "/" + n
			}
			files[key] = string(data)
		}
	}
	_ = moduleDir
	return files
}

// formatProfileBaseline renders counts as sorted "file<TAB>symbol<TAB>count"
// lines; the one named exception carries a trailing "exception" column.
func formatProfileBaseline(counts map[string]int) string {
	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for _, k := range keys {
		line := fmt.Sprintf("%s\t%d", k, counts[k])
		if k == ownExceptionFile+"\t"+ownExceptionSymbol {
			line += "\texception"
		}
		b.WriteString(line + "\n")
	}
	return b.String()
}

// parseProfileBaseline reads a baseline file body into counts (comment lines
// start with '#').
func parseProfileBaseline(t *testing.T, body string) map[string]int {
	t.Helper()
	out := map[string]int{}
	for _, line := range strings.Split(body, "\n") {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		cols := strings.Split(line, "\t")
		if len(cols) < 3 {
			t.Fatalf("malformed baseline line %q", line)
		}
		n, err := strconv.Atoi(cols[2])
		if err != nil {
			t.Fatalf("malformed baseline count in %q: %v", line, err)
		}
		out[cols[0]+"\t"+cols[1]] = n
	}
	return out
}

var updateProfileBaseline = flag.Bool("update-profile-baseline", false,
	"rewrite internal/cli/testdata/profile_callsite_baseline.txt from the current tree (M0 freeze only)")

// ---------------------------------------------------------------------------
// Tests of the helpers themselves (characterization: GREEN on arrival).
// ---------------------------------------------------------------------------

// TestOwnershipHelperTreeSnapshotSeesDirMtime: the snapshot must distinguish a
// tree where a temporary file was created and deleted from the untouched tree
// only when directory mtimes are folded in, and must always see byte changes.
func TestOwnershipHelperTreeSnapshotSeesDirMtime(t *testing.T) {
	root := t.TempDir()
	ownWriteFile(t, root, "a/b/file.txt", []byte("one\n"))
	pin := time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)
	pinTreeDirMtimes(t, root, pin)

	withMtime := ownTreeOpts{DirMtime: true}
	noMtime := ownTreeOpts{}
	beforeM := ownTreeDigest(ownTreeSnapshot(t, root, withMtime))
	beforeN := ownTreeDigest(ownTreeSnapshot(t, root, noMtime))
	if again := ownTreeDigest(ownTreeSnapshot(t, root, withMtime)); again != beforeM {
		t.Fatalf("snapshot is not deterministic on an untouched tree")
	}

	// Create-then-delete inside a/b: content-only digest cannot see it.
	tmp := filepath.Join(root, "a", "b", "probe.tmp")
	if err := os.WriteFile(tmp, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(tmp); err != nil {
		t.Fatal(err)
	}
	if got := ownTreeDigest(ownTreeSnapshot(t, root, noMtime)); got != beforeN {
		t.Errorf("content-only digest changed after a create-then-delete; it must not")
	}
	if got := ownTreeDigest(ownTreeSnapshot(t, root, withMtime)); got == beforeM {
		t.Errorf("mtime-folding digest did not change after a create-then-delete; a temporary write would go unseen")
	}

	// Create-then-delete directly in the root directory: only the root's own
	// mtime moves, so the snapshot must carry it.
	pinTreeDirMtimes(t, root, pin)
	rootBeforeM := ownTreeDigest(ownTreeSnapshot(t, root, withMtime))
	rootBeforeN := ownTreeDigest(ownTreeSnapshot(t, root, noMtime))
	rootTmp := filepath.Join(root, "probe.tmp")
	if err := os.WriteFile(rootTmp, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(rootTmp); err != nil {
		t.Fatal(err)
	}
	if got := ownTreeDigest(ownTreeSnapshot(t, root, noMtime)); got != rootBeforeN {
		t.Errorf("content-only digest changed after a create-then-delete in the root; it must not")
	}
	if got := ownTreeSnapshot(t, root, withMtime); ownTreeDigest(got) == rootBeforeM {
		t.Errorf("mtime-folding digest did not change after a create-then-delete in the root directory; a probe in the project root would go unseen")
	}

	// A byte change is always visible, and SkipPrefixes hides a subtree.
	ownWriteFile(t, root, "a/b/file.txt", []byte("two\n"))
	if got := ownTreeDigest(ownTreeSnapshot(t, root, noMtime)); got == beforeN {
		t.Errorf("content-only digest did not see a byte change")
	}
	ownWriteFile(t, root, ".moai-backups/run/x", []byte("y"))
	skipped := ownTreeSnapshot(t, root, ownTreeOpts{SkipPrefixes: []string{".moai-backups/"}})
	for k := range skipped {
		if strings.HasPrefix(k, ".moai-backups") {
			t.Errorf("SkipPrefixes left %s in the snapshot", k)
		}
	}
}

// TestOwnershipHelperForcedRunReachesReconciliation measures the version-match
// gate on this tree through the helpers: a plain first run reaches
// reconciliation, a plain same-version re-run is skipped, and a --force run is
// not. It is the reach assertion every multi-run criterion relies on.
func TestOwnershipHelperForcedRunReachesReconciliation(t *testing.T) {
	ownProfileSandbox(t)
	home := installMigrationUserCounterparts(t)
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	root := buildMigrationFixture(t)

	first := runUpdateFlowPlain(t, root)
	t.Logf("force=false run=1 skipped=%v", first.Skipped)
	if first.Skipped {
		t.Fatalf("plain run 1 was skipped; the fixture is already up to date")
	}
	second := runUpdateFlowPlain(t, root)
	t.Logf("force=false run=2 skipped=%v", second.Skipped)
	if !second.Skipped {
		t.Errorf("plain run 2 (same version) was not skipped; the gate this SPEC leaves unchanged moved")
	}
	forced := runUpdateFlowForced(t, root)
	t.Logf("force=true  run=3 skipped=%v", forced.Skipped)
	if forced.Skipped {
		t.Errorf("forced run reported skipped")
	}
	third := runUpdateFlowPlain(t, root)
	t.Logf("force=false run=4 skipped=%v", third.Skipped)
	if !third.Skipped {
		t.Errorf("plain run 4 (same version after a forced run) was not skipped")
	}
}

// TestOwnershipHelperUsesRealEmbeddedPaths pins the fixtures' premise: the two
// paths AC-OWR-02 plants exist in the embedded tree, and a user counterpart for
// each is confirmable in a freshly installed test home.
func TestOwnershipHelperUsesRealEmbeddedPaths(t *testing.T) {
	for _, rel := range []string{ownEmbeddedClaudeSkill, ownEmbeddedAgentsSkill} {
		if len(ownEmbeddedBytes(t, rel)) == 0 {
			t.Errorf("embedded %s is empty", rel)
		}
	}
	home := installMigrationUserCounterparts(t)
	requireUserCounterpart(t, home, ownEmbeddedClaudeSkill)
	requireUserCounterpart(t, home, ownEmbeddedAgentsSkill)

	root := t.TempDir()
	got := ownAddUnregisteredFromTemplate(t, root, ownEmbeddedAgentsSkill)
	onDisk, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(ownEmbeddedAgentsSkill)))
	if err != nil || !bytes.Equal(onDisk, got) {
		t.Errorf("unregistered-from-template fixture does not carry the template bytes: %v", err)
	}
}

// TestProfileCallsiteBaselineMatchesTree measures the profile call sites of the
// update path's whole call closure and compares them to the frozen baseline.
// With -update-profile-baseline it rewrites the baseline instead (M0 freeze).
func TestProfileCallsiteBaselineMatchesTree(t *testing.T) {
	files := listUpdateClosureSources(t)
	if _, ok := files["internal/profile/profile.go"]; !ok {
		t.Fatalf("the scanned closure does not contain internal/profile/profile.go; the scan would be vacuous")
	}
	counts, err := scanProfileCallsites(files)
	if err != nil {
		t.Fatal(err)
	}
	if got := counts[ownExceptionFile+"\t"+ownExceptionSymbol]; got != 1 {
		t.Fatalf("the named exception %s %s measured %d times, want exactly 1", ownExceptionFile, ownExceptionSymbol, got)
	}
	if *updateProfileBaseline {
		header := "# Profile call-site baseline (SPEC-UPDATE-OWNERSHIP-REPAIR-001 M0, card t1865).\n" +
			"# Frozen from an untouched tree by TestProfileCallsiteBaselineMatchesTree -update-profile-baseline.\n" +
			"# Measured over the call closure of `go list -deps` (union of GOOS=darwin,linux,windows) of internal/cli,\n" +
			"# non-test .go files of every module-internal package; scanner: scanProfileCallsites.\n" +
			"# Columns: file<TAB>symbol<TAB>count[<TAB>exception]. The one exception row is the existing\n" +
			"# launchCandidateIsUsable os.Stat REQ-OWR-005 tolerates (leader ruling OQ-5).\n"
		if sha := strings.TrimSpace(os.Getenv("OWR_BASELINE_TREE")); sha != "" {
			header += "# tree: " + sha + "\n"
		}
		body := header + formatProfileBaseline(counts)
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(ownBaselineRel, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("rewrote %s with %d rows over %d scanned files", ownBaselineRel, len(counts), len(files))
		return
	}
	raw, err := os.ReadFile(ownBaselineRel)
	if err != nil {
		t.Fatalf("read the frozen baseline: %v", err)
	}
	want := parseProfileBaseline(t, string(raw))
	if len(want) == 0 {
		t.Fatalf("the frozen baseline holds no rows; the comparison would be vacuous")
	}
	var problems []string
	for k, n := range counts {
		if w, ok := want[k]; !ok {
			problems = append(problems, fmt.Sprintf("NEW call site %q x%d", k, n))
		} else if n > w {
			problems = append(problems, fmt.Sprintf("GREW %q: %d > baseline %d", k, n, w))
		} else if n < w {
			problems = append(problems, fmt.Sprintf("SHRANK %q: %d < baseline %d (re-freeze deliberately)", k, n, w))
		}
	}
	for k, w := range want {
		if _, ok := counts[k]; !ok {
			problems = append(problems, fmt.Sprintf("GONE %q (baseline %d)", k, w))
		}
	}
	sort.Strings(problems)
	for _, p := range problems {
		t.Errorf("profile call-site baseline drift: %s", p)
	}
}

// TestOwnershipHelperFixtureBuilders characterizes the builders the later
// milestones share: the user-created state, the tree diff, the temporary-home
// guard, and the three canary profile states. A helper nothing exercises is a
// helper nothing can trust.
func TestOwnershipHelperFixtureBuilders(t *testing.T) {
	t.Run("RegisteredUserCreatedAndTreeDiff", func(t *testing.T) {
		root := t.TempDir()
		const rel = ".claude/skills/mine/SKILL.md"
		ownAddRegisteredUserCreated(t, root, rel, []byte("# mine\n"))
		mgr := manifest.NewManager()
		if _, err := mgr.Load(root); err != nil {
			t.Fatal(err)
		}
		if e, ok := mgr.GetEntry(rel); !ok || e.Provenance != manifest.UserCreated {
			t.Fatalf("user-created builder recorded %+v (found=%v)", e, ok)
		}
		before := ownTreeSnapshot(t, root, ownTreeOpts{})
		ownWriteFile(t, root, "added/file.txt", []byte("x"))
		ownWriteFile(t, root, rel, []byte("# mine, edited\n"))
		diff := ownTreeDiff(before, ownTreeSnapshot(t, root, ownTreeOpts{}))
		want := []string{".claude/skills/mine/SKILL.md", "added/", "added/file.txt"}
		if !reflect.DeepEqual(diff, want) {
			t.Errorf("tree diff %v, want %v", diff, want)
		}
	})

	t.Run("TemporaryHomeGuard", func(t *testing.T) {
		home := ownIsolateHome(t)
		ownAssertHomeIsolated(t)
		got, err := os.UserHomeDir()
		if err != nil || got != home {
			t.Errorf("isolated HOME resolves to %q (%v), want %q", got, err, home)
		}
	})

	t.Run("CanaryProfileStates", func(t *testing.T) {
		base := ownProfileSandbox(t)
		for _, tc := range []struct {
			state     ownCanaryState
			wantFiles bool
			sameAsTpl bool
		}{{ownCanaryNone, false, false}, {ownCanaryTemplateSame, true, true}, {ownCanaryDiffers, true, false}} {
			dir := ownCanaryProfileDir(t, base, tc.state)
			skill := filepath.Join(dir, "skills", "x", "SKILL.md")
			data, err := os.ReadFile(skill)
			if tc.wantFiles != (err == nil) {
				t.Fatalf("state %d: skill present=%v, want %v", tc.state, err == nil, tc.wantFiles)
			}
			if tc.wantFiles && tc.sameAsTpl != bytes.Equal(data, ownEmbeddedBytes(t, ownEmbeddedClaudeSkill)) {
				t.Errorf("state %d: template-equality of the canary skill is %v, want %v", tc.state, !tc.sameAsTpl, tc.sameAsTpl)
			}
			// A fresh state replaces the previous one.
			if err := os.RemoveAll(filepath.Join(dir, "skills")); err != nil {
				t.Fatal(err)
			}
			if err := os.RemoveAll(filepath.Join(dir, "agents")); err != nil {
				t.Fatal(err)
			}
			ownSetClaudeConfigDir(t, dir)
			if os.Getenv(config.EnvClaudeConfigDir) != dir {
				t.Errorf("CLAUDE_CONFIG_DIR was not pointed at the canary directory")
			}
		}
	})
}
