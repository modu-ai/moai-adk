package cli

// update_commands_discovery_test.go — SPEC-UPDATE-OWNERSHIP-REPAIR-001 M6
// (card t1865), REQ-OWR-006: the update gives each command file exactly one
// disposition per run, and after every run — an interrupted and then recovered
// run included — every command and agent file the product delivers is present.
//
// Where the files live. Commands are a PROJECT payload (`.claude/commands/moai`).
// Agents are not: the deployer's project payload carries no common skill or agent
// file in any mode (isCommonAssetRoot, SPEC-USER-ASSET-INSTALL-001 REQ-005), and
// the user-asset phase of `runUpdate` installs the default (L0) agent set into the
// user roots (`~/.claude/agents/<name>.md`, `~/.codex/agents/<name>.toml`). So
// "present" is judged where each kind is delivered, and the oracle is a control
// run of the same entry on a pristine project in a pristine home — the deployer's
// and the installer's own output, which is the only honest oracle for the bytes of
// a `.tmpl` command and for which agents the default selection installs.
//
// AC-OWR-09 is first a MEASUREMENT. The field symptom ("commands left out of the
// migration and then redeployed") is not reproduced by reading the code, so the
// same project shape is driven two ways: V1 (the project's template record equals
// the binary version; only the recorded template tree is stale, so the sync is
// reached without --force) and V2 (the record and the manifest carry an older
// release's values — the field case was a rc.35 record under a rc.36 binary —
// driven through the whole `runUpdate` entry). A variant that is not red is a
// regression guard, and the assertion functions are shown red in front of
// injected wrong outcomes.
//
// AC-OWR-10 recovers a project whose earlier run died with the commands and
// project agent directories emptied. The recovery run is --force (a same-version
// plain re-run is skipped before reconciliation).
//
// Time independence: one flow run per project; no assertion depends on which
// clock second a run lands in.

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/modu-ai/moai-adk/internal/cli/update"
	"github.com/modu-ai/moai-adk/internal/cli/update/report"
	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/pkg/version"
)

// The four command states of AC-OWR-09, one observable outcome each.
const (
	cmdHealthy      = ".claude/commands/moai/gtd.md"      // S1 registered managed, healthy
	cmdUserModified = ".claude/commands/moai/fix.md"      // S2 registered, then the user added a line
	cmdUnregistered = ".claude/commands/moai/todo.md"     // S3 template bytes, no manifest entry
	cmdUserOwn      = ".claude/commands/moai-user-cmd.md" // S4 the user's own command

	cmdDir           = ".claude/commands/moai"
	cmdUserLine      = "USERLINE: keep me"
	cmdBinaryVersion = "v3.2.0-rc.36"
	cmdOldVersion    = "v3.2.0-rc.35"
	cmdStaleTree     = "stale-tree-hash"
)

var cmdUserOwnBytes = []byte("# the user's own command\n")

// cmdControl is what a pristine run of the whole `runUpdate` entry delivers.
type cmdControl struct {
	// commands maps a project-relative command path to the deployer's bytes.
	commands map[string][]byte
	// userFiles maps a home-relative agent path to the installer's bytes.
	userFiles map[string][]byte
}

// cmdPinVersion sets the binary version for one test.
func cmdPinVersion(t *testing.T, v string) {
	t.Helper()
	prev := version.Version
	version.Version = v
	t.Cleanup(func() { version.Version = prev })
}

// cmdWriteProject writes the project's identity: the system config carrying the
// template record version, and the harness choice.
func cmdWriteProject(t *testing.T, root, harness, recordVersion string) {
	t.Helper()
	ownWriteFile(t, root, ".moai/config/sections/system.yaml",
		[]byte(fmt.Sprintf("moai:\n  version: v3.0.0\n  template_version: %q\n", recordVersion)))
	ownWriteFile(t, root, ".moai/config/sections/llm.yaml", []byte("llm:\n  harness: "+harness+"\n"))
}

// cmdStampManifest sets the manifest-level version and recorded template tree.
func cmdStampManifest(t *testing.T, root, ver, treeHash string) {
	t.Helper()
	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	mgr.Manifest().Version = ver
	mgr.Manifest().TemplateTreeHash = treeHash
	if err := mgr.Save(); err != nil {
		t.Fatalf("save manifest: %v", err)
	}
}

// cmdTemplateCommandNames returns the command file names the embedded template
// carries, as deployed (a `.tmpl` suffix is the template form of the same file).
func cmdTemplateCommandNames(t *testing.T) []string {
	t.Helper()
	embedded, err := template.EmbeddedTemplates()
	if err != nil {
		t.Fatalf("load embedded templates: %v", err)
	}
	entries, err := fs.ReadDir(embedded, cmdDir)
	if err != nil {
		t.Fatalf("embedded directory %s: %v", cmdDir, err)
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() {
			names = append(names, strings.TrimSuffix(e.Name(), ".tmpl"))
		}
	}
	sort.Strings(names)
	return names
}

// cmdUserAgentDirs are the user-root agent directories, home-relative.
var cmdUserAgentDirs = []string{".claude/agents", ".codex/agents"}

// cmdReadUserAgents returns the agent files under the user roots of home.
func cmdReadUserAgents(t *testing.T, home string) map[string][]byte {
	t.Helper()
	out := map[string][]byte{}
	for _, dir := range cmdUserAgentDirs {
		entries, err := os.ReadDir(filepath.Join(home, filepath.FromSlash(dir)))
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			data, err := os.ReadFile(filepath.Join(home, filepath.FromSlash(dir), e.Name()))
			if err != nil {
				t.Fatalf("read user agent %s/%s: %v", dir, e.Name(), err)
			}
			out[dir+"/"+e.Name()] = data
		}
	}
	return out
}

// cmdRenderControl runs the whole `runUpdate` entry on a pristine project in a
// pristine home and returns what it delivered. The premise is checked: the
// deployer wrote exactly the template-carried command names, and at least one
// agent was installed.
func cmdRenderControl(t *testing.T, harness string) cmdControl {
	t.Helper()
	home := ownIsolateHome(t)
	root := t.TempDir()
	cmdWriteProject(t, root, harness, cmdOldVersion)
	cmdStampManifest(t, root, cmdOldVersion, cmdStaleTree)
	updateLedger.reset()
	out := cmdRunCommand(t, root, false)
	if strings.Contains(out, report.RenderOutcome(report.OutcomeAlreadyUpToDate, 0, "")) {
		t.Fatalf("the control run was ended by the version-match skip; it delivered nothing:\n%s", out)
	}
	c := cmdControl{commands: map[string][]byte{}, userFiles: cmdReadUserAgents(t, home)}
	want := cmdTemplateCommandNames(t)
	entries, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(cmdDir)))
	if err != nil {
		t.Fatalf("control: %s: %v", cmdDir, err)
	}
	got := map[string]bool{}
	for _, e := range entries {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(cmdDir), e.Name()))
		if err != nil {
			t.Fatalf("control: read %s/%s: %v", cmdDir, e.Name(), err)
		}
		c.commands[cmdDir+"/"+e.Name()] = data
		got[e.Name()] = true
	}
	for _, n := range want {
		if !got[n] {
			t.Fatalf("control: the deployer did not write %s/%s although the template carries it", cmdDir, n)
		}
	}
	if len(got) != len(want) {
		t.Fatalf("control: %s holds %d files, the template carries %d", cmdDir, len(got), len(want))
	}
	if len(c.userFiles) == 0 {
		t.Fatalf("control: the user-asset phase installed no agent under %v", cmdUserAgentDirs)
	}
	return c
}

// cmdFrontmatter returns the top-level YAML frontmatter keys of a file.
func cmdFrontmatter(data []byte) (map[string]any, error) {
	rest, ok := bytes.CutPrefix(data, []byte("---\n"))
	if !ok {
		return nil, fmt.Errorf("no frontmatter")
	}
	block, _, ok := bytes.Cut(rest, []byte("\n---"))
	if !ok {
		return nil, fmt.Errorf("unterminated frontmatter")
	}
	var fm map[string]any
	if err := yaml.Unmarshal(block, &fm); err != nil {
		return nil, err
	}
	return fm, nil
}

// cmdNonEmptyField reports whether key is a non-empty string in fm.
func cmdNonEmptyField(fm map[string]any, key string) bool {
	v, ok := fm[key]
	if !ok {
		return false
	}
	s, isString := v.(string)
	return isString && strings.TrimSpace(s) != ""
}

// cmdPresenceProblems is the pure assertion of AC-OWR-10 (and the presence half
// of AC-OWR-09): every command the template carries is in the project, and every
// agent the control run installed is in the user roots, each non-empty, carrying
// the frontmatter its kind requires, and byte-equal to what the product renders.
// Commands and agents are separate conditions because their formats differ: a
// command has `description` and `allowed-tools` (and no `name`), an agent has
// `name` and `description`. skip names project paths whose bytes are judged
// elsewhere (the user-modified command).
func cmdPresenceProblems(t *testing.T, root, home string, control cmdControl, skip map[string]bool) []string {
	t.Helper()
	var problems []string
	for _, name := range cmdTemplateCommandNames(t) {
		rel := cmdDir + "/" + name
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil {
			problems = append(problems, fmt.Sprintf("%s: missing after the run", rel))
			continue
		}
		if len(bytes.TrimSpace(data)) == 0 {
			problems = append(problems, fmt.Sprintf("%s: empty after the run", rel))
			continue
		}
		if skip[rel] {
			continue
		}
		if exp, ok := control.commands[rel]; ok && !bytes.Equal(data, exp) {
			problems = append(problems, fmt.Sprintf("%s: bytes differ from the deployer's render", rel))
		}
		fm, fmErr := cmdFrontmatter(data)
		if fmErr != nil {
			problems = append(problems, fmt.Sprintf("%s: frontmatter: %v", rel, fmErr))
			continue
		}
		for _, key := range []string{"description", "allowed-tools"} {
			if !cmdNonEmptyField(fm, key) {
				problems = append(problems, fmt.Sprintf("%s: frontmatter %q is missing or empty", rel, key))
			}
		}
	}
	for rel, exp := range control.userFiles {
		data, err := os.ReadFile(filepath.Join(home, filepath.FromSlash(rel)))
		if err != nil {
			problems = append(problems, fmt.Sprintf("~/%s: missing after the run", rel))
			continue
		}
		if len(bytes.TrimSpace(data)) == 0 {
			problems = append(problems, fmt.Sprintf("~/%s: empty after the run", rel))
			continue
		}
		if !bytes.Equal(data, exp) {
			problems = append(problems, fmt.Sprintf("~/%s: bytes differ from the installer's output", rel))
		}
		if !strings.HasSuffix(rel, ".md") {
			continue
		}
		fm, fmErr := cmdFrontmatter(data)
		if fmErr != nil {
			problems = append(problems, fmt.Sprintf("~/%s: frontmatter: %v", rel, fmErr))
			continue
		}
		for _, key := range []string{"name", "description"} {
			if !cmdNonEmptyField(fm, key) {
				problems = append(problems, fmt.Sprintf("~/%s: frontmatter %q is missing or empty", rel, key))
			}
		}
	}
	sort.Strings(problems)
	return problems
}

// cmdPlanProblems classifies the managed roots the way the run's reconciliation
// does (the read-only function the dry-run preview calls) and returns one
// problem per path that appears in more than one disposition list, plus the
// vacuity checks: the plan is not empty and covers the named paths.
func cmdPlanProblems(t *testing.T, root string, mustCover ...string) []string {
	t.Helper()
	embedded, err := template.EmbeddedTemplates()
	if err != nil {
		t.Fatalf("load embedded templates: %v", err)
	}
	mode := resolveUpdateDeployMode(root, false)
	plan, err := update.ClassifyManagedRoots(root,
		reconcileTargets(root, computeRunCleanTargets(root, mode, nil)),
		templateRenderCarriage(embedded, nil, nil), update.LoadManifestReadOnly(root))
	if err != nil {
		t.Fatalf("classify managed roots: %v", err)
	}
	seen := map[string][]string{}
	for class, set := range map[string][]update.ReconcileFile{
		"template-owned": plan.TemplateOwned, "user-modified": plan.UserModified,
		"user-owned": plan.UserOwned, "stale": plan.Stale,
	} {
		for _, f := range set {
			seen[f.RelPath] = append(seen[f.RelPath], class)
		}
	}
	var problems []string
	for rel, classes := range seen {
		if len(classes) > 1 {
			sort.Strings(classes)
			problems = append(problems, fmt.Sprintf("%s: appears in %d dispositions %v", rel, len(classes), classes))
		}
	}
	if plan.Count() == 0 {
		problems = append(problems, "the reconcile plan is empty: the disjointness check would be vacuous")
	}
	for _, rel := range mustCover {
		if plan.ClassOf(rel) == "" {
			problems = append(problems, fmt.Sprintf("%s: not in the reconcile plan at all", rel))
		}
	}
	sort.Strings(problems)
	return problems
}

// cmdOutcomeProblems is the pure assertion on the four command states after a
// run and on the report: each state has its one outcome, and no fixture path is
// described by more than one report line (one disposition per path per run).
func cmdOutcomeProblems(root, out string, control cmdControl) []string {
	var problems []string
	read := func(rel string) ([]byte, bool) {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		return data, err == nil
	}
	if got, ok := read(cmdHealthy); !ok || !bytes.Equal(got, control.commands[cmdHealthy]) {
		problems = append(problems, fmt.Sprintf("S1 %s: bytes are not the current render (present=%v)", cmdHealthy, ok))
	}
	got, ok := read(cmdUserModified)
	switch {
	case !ok:
		problems = append(problems, fmt.Sprintf("S2 %s: the user-modified command is gone", cmdUserModified))
	case !strings.Contains(string(got), cmdUserLine):
		problems = append(problems, fmt.Sprintf("S2 %s: the user's added line is not in the file", cmdUserModified))
	}
	if got, ok := read(cmdUnregistered); !ok || !bytes.Equal(got, control.commands[cmdUnregistered]) {
		problems = append(problems, fmt.Sprintf("S3 %s: missing or not the template render after the run (present=%v)", cmdUnregistered, ok))
	}
	if got, ok := read(cmdUserOwn); !ok || !bytes.Equal(got, cmdUserOwnBytes) {
		problems = append(problems, fmt.Sprintf("S4 %s: the user's own command changed (present=%v)", cmdUserOwn, ok))
	}
	for _, rel := range []string{cmdHealthy, cmdUserModified, cmdUnregistered, cmdUserOwn} {
		n := 0
		for _, line := range strings.Split(out, "\n") {
			if strings.Contains(line, rel) {
				n++
			}
		}
		if n > 1 {
			problems = append(problems, fmt.Sprintf("%s: %d report lines describe it, want at most one disposition line", rel, n))
		}
	}
	return problems
}

// cmdBuildProject writes the AC-OWR-09 fixture: the four command states over a
// project whose template record is recordVersion, with a stale recorded tree so
// the version-match skip cannot end the run.
func cmdBuildProject(t *testing.T, root, harness, recordVersion string, control cmdControl) {
	t.Helper()
	cmdWriteProject(t, root, harness, recordVersion)
	ownAddRegisteredManaged(t, root, cmdHealthy, control.commands[cmdHealthy])
	ownAddRegisteredEdited(t, root, cmdUserModified, control.commands[cmdUserModified], "\n"+cmdUserLine+"\n", manifest.TemplateManaged)
	ownAddUnregistered(t, root, cmdUnregistered, control.commands[cmdUnregistered])
	ownAddUnregistered(t, root, cmdUserOwn, cmdUserOwnBytes)
	cmdStampManifest(t, root, recordVersion, cmdStaleTree)
}

// cmdRunCommand drives the whole `runUpdate` entry (the one `moai update`
// dispatches) as a non-dry-run update with --yes and --templates-only, which
// skips only the binary self-update. The binary version is pinned by the caller
// and HOME is the caller's isolated directory.
func cmdRunCommand(t *testing.T, root string, force bool) string {
	t.Helper()
	ownAssertHomeIsolated(t)
	origDeps := deps
	t.Cleanup(func() { deps = origDeps })
	deps = &Dependencies{}
	snapChdir(t, root)
	var out, errOut bytes.Buffer
	cmd := &cobra.Command{Use: "update-commands-test"}
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SetContext(context.Background())
	cmd.Flags().Bool("check", false, "")
	cmd.Flags().Bool("shell-env", false, "")
	cmd.Flags().Bool("config", false, "")
	cmd.Flags().Bool("binary", false, "")
	cmd.Flags().Bool("templates-only", true, "")
	cmd.Flags().Bool("yes", true, "")
	cmd.Flags().Bool("force", force, "")
	cmd.Flags().Bool("dry-run", false, "")
	cmd.Flags().Bool("no-hooks", true, "")
	cmd.Flags().Bool("verbose", false, "")
	cmd.Flags().String("profile", "", "")
	if err := runUpdate(cmd, nil); err != nil {
		t.Fatalf("runUpdate: %v\nout:\n%s\nerr:\n%s", err, out.String(), errOut.String())
	}
	return out.String()
}

// cmdNewPriorManifest returns the path of a manifest.prior.json written at or
// after began, or "".
func cmdNewPriorManifest(root string, began int64) string {
	matches, _ := filepath.Glob(filepath.Join(root, ".moai-backups", "*", priorManifestName))
	sort.Strings(matches)
	found := ""
	for _, m := range matches {
		if info, err := os.Stat(m); err == nil && info.ModTime().UnixNano() >= began {
			found = m
		}
	}
	return found
}

// cmdSkipped reports whether the run output is the version-match skip's.
func cmdSkipped(out string) bool {
	return strings.Contains(out, report.RenderOutcome(report.OutcomeAlreadyUpToDate, 0, ""))
}

// cmdLogObservation writes what a variant observed into the test log: the report
// lines that mention the command directory, the reconcile plan's class for the
// fixture paths, and the Reconciliation summary line. It is evidence for the
// reader of a -v run; nothing asserts on it.
func cmdLogObservation(t *testing.T, label, root, out string) {
	t.Helper()
	embedded, err := template.EmbeddedTemplates()
	if err != nil {
		return
	}
	plan, err := update.ClassifyManagedRoots(root,
		reconcileTargets(root, computeRunCleanTargets(root, resolveUpdateDeployMode(root, false), nil)),
		templateRenderCarriage(embedded, nil, nil), update.LoadManifestReadOnly(root))
	if err == nil {
		t.Logf("%s: plan after the run (read-only classification): %d files, %d template-owned, %d user-modified, %d user-owned, %d stale",
			label, plan.Count(), len(plan.TemplateOwned), len(plan.UserModified), len(plan.UserOwned), len(plan.Stale))
	}
	for _, line := range strings.Split(ansiStrip.ReplaceAllString(out, ""), "\n") {
		if strings.Contains(line, "Reconciliation") || strings.Contains(line, "commands") || strings.Contains(line, "conflict") {
			t.Logf("%s: report line: %s", label, strings.TrimSpace(line))
		}
	}
}

// TestCommandsSingleDispositionAndPresent is AC-OWR-09.
func TestCommandsSingleDispositionAndPresent(t *testing.T) {
	cmdPinVersion(t, cmdBinaryVersion)
	var control cmdControl
	t.Run("Control", func(t *testing.T) { control = cmdRenderControl(t, "claude") })
	if control.commands == nil {
		t.Fatalf("the control run did not complete; nothing below can be judged")
	}
	var v1Tree, v2Tree map[string]string

	t.Run("V1_SameVersion", func(t *testing.T) {
		home := ownIsolateHome(t)
		root := t.TempDir()
		// The record equals the binary; only the recorded tree is stale, so the
		// sync is reached without --force.
		cmdBuildProject(t, root, "claude", cmdBinaryVersion, control)
		for _, p := range cmdPlanProblems(t, root, cmdHealthy, cmdUnregistered) {
			t.Errorf("plan: %s", p)
		}
		updateLedger.reset()
		run := runUpdateFlowPlain(t, root)
		if run.Skipped {
			t.Fatalf("V1 was skipped before reconciliation; it judged nothing")
		}
		if run.NewPriorManifest == "" {
			t.Fatalf("V1 left no new manifest.prior.json; the Backup stage was not reached")
		}
		cmdLogObservation(t, "V1", root, run.Out)
		for _, p := range cmdOutcomeProblems(root, run.Out, control) {
			t.Errorf("outcome: %s", p)
		}
		for _, p := range cmdPresenceProblems(t, root, home, cmdControl{commands: control.commands}, map[string]bool{cmdUserModified: true}) {
			t.Errorf("presence: %s", p)
		}
		v1Tree = ownTreeSnapshot(t, filepath.Join(root, ".claude", "commands"), ownTreeOpts{})
	})

	t.Run("V2_StaleRecord", func(t *testing.T) {
		home := ownIsolateHome(t)
		root := t.TempDir()
		// The record and the manifest are one release behind the binary; the
		// whole runUpdate entry (the version-match gate included) is driven.
		cmdBuildProject(t, root, "claude", cmdOldVersion, control)
		for _, p := range cmdPlanProblems(t, root, cmdHealthy, cmdUnregistered) {
			t.Errorf("plan: %s", p)
		}
		updateLedger.reset()
		began := time.Now().Add(-time.Second).UnixNano()
		out := cmdRunCommand(t, root, false)
		if cmdSkipped(out) {
			t.Fatalf("V2 was ended by the version-match skip; it judged nothing:\n%s", out)
		}
		if cmdNewPriorManifest(root, began) == "" {
			t.Fatalf("V2 left no new manifest.prior.json; the Backup stage was not reached")
		}
		cmdLogObservation(t, "V2", root, out)
		for _, p := range cmdOutcomeProblems(root, out, control) {
			t.Errorf("outcome: %s", p)
		}
		for _, p := range cmdPresenceProblems(t, root, home, control, map[string]bool{cmdUserModified: true}) {
			t.Errorf("presence: %s", p)
		}
		v2Tree = ownTreeSnapshot(t, filepath.Join(root, ".claude", "commands"), ownTreeOpts{})
	})

	t.Run("V1AndV2ReachTheSameCommandsTree", func(t *testing.T) {
		if v1Tree == nil || v2Tree == nil {
			t.Fatalf("a variant did not complete (V1 %d entries, V2 %d entries); there is nothing to compare", len(v1Tree), len(v2Tree))
		}
		if diff := ownTreeDiff(v1Tree, v2Tree); len(diff) != 0 {
			t.Errorf("V1 and V2 left different commands trees; differing paths: %v", diff)
		}
	})

	// The pure assertions are shown red in front of wrong outcomes.
	t.Run("MutantTwoDispositionsForOnePathIsCaught", func(t *testing.T) {
		root := t.TempDir()
		cmdBuildProject(t, root, "claude", cmdOldVersion, control)
		out := strings.Repeat("  reconcile: "+cmdUserModified+"\n", 2)
		if got := cmdOutcomeProblems(root, out, control); !hasProblemFor(got, cmdUserModified, "report lines") {
			t.Errorf("two report lines for one path were not reported; problems: %v", got)
		}
	})
	t.Run("MutantRedeployOverTheUserEditIsCaught", func(t *testing.T) {
		root := t.TempDir()
		cmdBuildProject(t, root, "claude", cmdOldVersion, control)
		ownWriteFile(t, root, cmdUserModified, control.commands[cmdUserModified])
		if got := cmdOutcomeProblems(root, "", control); !hasProblemFor(got, cmdUserModified, "user's added line") {
			t.Errorf("a redeploy over the user's edit was not reported; problems: %v", got)
		}
	})
	t.Run("MutantRemovedCommandIsCaught", func(t *testing.T) {
		root := t.TempDir()
		cmdBuildProject(t, root, "claude", cmdOldVersion, control)
		if err := os.Remove(filepath.Join(root, filepath.FromSlash(cmdUnregistered))); err != nil {
			t.Fatal(err)
		}
		if got := cmdOutcomeProblems(root, "", control); !hasProblemFor(got, cmdUnregistered, "missing") {
			t.Errorf("a removed command was not reported; problems: %v", got)
		}
	})
}

// hasProblemFor reports whether a problem names rel and contains frag.
func hasProblemFor(problems []string, rel, frag string) bool {
	for _, p := range problems {
		if strings.Contains(p, rel) && strings.Contains(p, frag) {
			return true
		}
	}
	return false
}

var keptInPlaceRe = regexp.MustCompile(`kept in place: (\d+)`)
var ansiStrip = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// keptRowCount reads the number of remaining rows from a run's output: the sum
// of every "kept in place: N" line of its terminal block.
func keptRowCount(out string) int {
	total := 0
	for _, line := range strings.Split(ansiStrip.ReplaceAllString(out, ""), "\n") {
		if m := keptInPlaceRe.FindStringSubmatch(line); m != nil {
			n, _ := strconv.Atoi(m[1])
			total += n
		}
	}
	return total
}

// cmdEmptyProjectCommands deletes every file under the project command directory
// and leaves the directory in place: the state a run that died after the clean
// and before the deploy leaves behind.
func cmdEmptyProjectCommands(t *testing.T, root string) {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(cmdDir)))
	if err != nil {
		t.Fatalf("read %s: %v", cmdDir, err)
	}
	for _, e := range entries {
		if err := os.Remove(filepath.Join(root, filepath.FromSlash(cmdDir), e.Name())); err != nil {
			t.Fatalf("remove %s/%s: %v", cmdDir, e.Name(), err)
		}
	}
}

// cmdProjectAgentDirs are the project agent directories an older deploy filled
// and the interrupted run emptied (the deployer no longer writes them).
var cmdProjectAgentDirs = []string{".claude/agents/moai", ".codex/agents/moai"}

// cmdInterruptedProject builds a project whose earlier run died with the commands
// and the project agent directories emptied: every shipped file was registered
// (so the manifest still remembers it) and then deleted, and the user roots hold
// no agent yet. An unregistered skill copy supplies a positive remaining-row
// count, which shows the migration was reached, and a user command shows it
// survives.
func cmdInterruptedProject(t *testing.T, harness string, control cmdControl) string {
	t.Helper()
	root := t.TempDir()
	cmdWriteProject(t, root, harness, cmdOldVersion)
	for rel, data := range control.commands {
		ownAddRegisteredManaged(t, root, rel, data)
	}
	embedded, err := template.EmbeddedTemplates()
	if err != nil {
		t.Fatalf("load embedded templates: %v", err)
	}
	for _, dir := range cmdProjectAgentDirs {
		entries, err := fs.ReadDir(embedded, dir)
		if err != nil {
			t.Fatalf("embedded directory %s: %v", dir, err)
		}
		for _, e := range entries {
			data, err := fs.ReadFile(embedded, dir+"/"+e.Name())
			if err != nil {
				t.Fatalf("embedded read %s/%s: %v", dir, e.Name(), err)
			}
			ownAddRegisteredManaged(t, root, dir+"/"+e.Name(), data)
		}
	}
	// The earlier run emptied these directories.
	cmdEmptyProjectCommands(t, root)
	for _, dir := range cmdProjectAgentDirs {
		entries, _ := os.ReadDir(filepath.Join(root, filepath.FromSlash(dir)))
		for _, e := range entries {
			if err := os.Remove(filepath.Join(root, filepath.FromSlash(dir), e.Name())); err != nil {
				t.Fatalf("remove %s/%s: %v", dir, e.Name(), err)
			}
		}
	}
	ownAddUnregisteredFromTemplate(t, root, ownEmbeddedAgentsSkill)
	ownAddUnregistered(t, root, cmdUserOwn, cmdUserOwnBytes)
	cmdStampManifest(t, root, cmdOldVersion, cmdStaleTree)
	return root
}

// TestCommandsAndAgentsDiscoverableAfterInterruptedRun is AC-OWR-10.
func TestCommandsAndAgentsDiscoverableAfterInterruptedRun(t *testing.T) {
	cmdPinVersion(t, cmdBinaryVersion)
	for _, harness := range []string{"claude", "both"} {
		t.Run(harness, func(t *testing.T) {
			var control cmdControl
			t.Run("Control", func(t *testing.T) { control = cmdRenderControl(t, harness) })
			if control.commands == nil {
				t.Fatalf("the control run did not complete; nothing below can be judged")
			}
			home := ownIsolateHome(t)
			root := cmdInterruptedProject(t, harness, control)

			// Premise: the interrupted state really is emptied, and the presence
			// assertion is red on it.
			for _, dir := range append([]string{cmdDir}, cmdProjectAgentDirs...) {
				entries, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(dir)))
				if err != nil || len(entries) != 0 {
					t.Fatalf("fixture error: %s is not an empty directory (entries=%d, err=%v)", dir, len(entries), err)
				}
			}
			if before := cmdPresenceProblems(t, root, home, control, nil); len(before) == 0 {
				t.Fatalf("the presence assertion stayed green on the emptied state; it asserts nothing")
			}

			t.Run("RecoveryRun", func(t *testing.T) {
				updateLedger.reset()
				began := time.Now().Add(-time.Second).UnixNano()
				out := cmdRunCommand(t, root, true)
				if cmdSkipped(out) {
					t.Fatalf("the recovery run was ended by the version-match skip; it judged nothing:\n%s", out)
				}
				if cmdNewPriorManifest(root, began) == "" {
					t.Fatalf("the recovery run left no new manifest.prior.json; the Backup stage was not reached")
				}
				if n := keptRowCount(out); n < 1 {
					t.Fatalf("the recovery run listed %d remaining rows; the migration was not reached:\n%s", n, out)
				}
				for _, p := range cmdPresenceProblems(t, root, home, control, nil) {
					t.Errorf("%s", p)
				}
				if got, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(cmdUserOwn))); err != nil || !bytes.Equal(got, cmdUserOwnBytes) {
					t.Errorf("the user's own command did not survive the recovery run (%v)", err)
				}
				for _, dir := range cmdProjectAgentDirs {
					if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(dir))); err != nil {
						t.Errorf("the project agent directory %s is gone after the run: %v", dir, err)
					}
				}
			})

			t.Run("PlainSameVersionRerunIsSkipped", func(t *testing.T) {
				// The operational fact the criterion records: after the recovery the
				// record equals the binary and the recorded tree is current, so a
				// plain re-run is skipped before reconciliation and emptied project
				// commands would stay empty. The recovery guarantee belongs to a run
				// that reaches reconciliation (plan.md OQ-6).
				cmdEmptyProjectCommands(t, root)
				updateLedger.reset()
				out := cmdRunCommand(t, root, false)
				if !cmdSkipped(out) {
					t.Fatalf("a plain same-version re-run was not skipped:\n%s", out)
				}
				if entries, _ := os.ReadDir(filepath.Join(root, filepath.FromSlash(cmdDir))); len(entries) != 0 {
					t.Errorf("the skipped run changed %s (%d entries)", cmdDir, len(entries))
				}
			})

			// The pure assertion is shown red in front of wrong trees.
			t.Run("MutantsAreCaught", func(t *testing.T) {
				home := ownIsolateHome(t)
				recovered := cmdInterruptedProject(t, harness, control)
				updateLedger.reset()
				if out := cmdRunCommand(t, recovered, true); cmdSkipped(out) {
					t.Fatalf("the mutant base run was skipped:\n%s", out)
				}
				if got := cmdPresenceProblems(t, recovered, home, control, nil); len(got) != 0 {
					t.Fatalf("the recovered tree is not clean, so mutants prove nothing: %v", got)
				}
				cmdRel := cmdDir + "/" + cmdTemplateCommandNames(t)[0]
				var agentRel string
				for rel := range control.userFiles {
					if strings.HasSuffix(rel, ".md") && (agentRel == "" || rel < agentRel) {
						agentRel = rel
					}
				}
				if agentRel == "" {
					t.Fatalf("no markdown agent in the control run")
				}

				// A recovery that does not redeploy a command: the name set is short.
				orig, err := os.ReadFile(filepath.Join(recovered, filepath.FromSlash(cmdRel)))
				if err != nil {
					t.Fatal(err)
				}
				if err := os.Remove(filepath.Join(recovered, filepath.FromSlash(cmdRel))); err != nil {
					t.Fatal(err)
				}
				if got := cmdPresenceProblems(t, recovered, home, control, nil); !hasProblemFor(got, cmdRel, "missing") {
					t.Errorf("a missing command was not reported; problems: %v", got)
				}
				// One file written empty.
				ownWriteFile(t, recovered, cmdRel, nil)
				if got := cmdPresenceProblems(t, recovered, home, control, nil); !hasProblemFor(got, cmdRel, "empty") {
					t.Errorf("an empty command was not reported; problems: %v", got)
				}
				ownWriteFile(t, recovered, cmdRel, orig)
				// An agent without its name.
				agent, err := os.ReadFile(filepath.Join(home, filepath.FromSlash(agentRel)))
				if err != nil {
					t.Fatal(err)
				}
				ownWriteFile(t, home, agentRel, bytes.Replace(agent, []byte("\nname:"), []byte("\nnom:"), 1))
				if got := cmdPresenceProblems(t, recovered, home, control, nil); !hasProblemFor(got, agentRel, "name") {
					t.Errorf("an agent without a name was not reported; problems: %v", got)
				}
				// An agent that was not installed.
				if err := os.Remove(filepath.Join(home, filepath.FromSlash(agentRel))); err != nil {
					t.Fatal(err)
				}
				if got := cmdPresenceProblems(t, recovered, home, control, nil); !hasProblemFor(got, agentRel, "missing") {
					t.Errorf("a missing agent was not reported; problems: %v", got)
				}
			})
		})
	}
}
