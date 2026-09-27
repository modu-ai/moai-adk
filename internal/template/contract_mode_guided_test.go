// contract_mode_guided_test.go — base-ref guards for the contract-mode
// rewiring. Each test compares the working tree against the base ref named by
// MOAI_GR_BASE (the last absorbed integration commit). CI does not know that
// ref, so each test skips when the variable is unset; an acceptance run MUST
// set it and read `--- PASS`, never `--- SKIP`.
//
// The base ref is always read from the environment, never pinned in this
// file, so the guards re-derive cleanly after the base is re-absorbed.
package template_test

import (
	"archive/tar"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/modu-ai/moai-adk/internal/constitution"
)

// grBaseEnv names the base ref for these guards.
const grBaseEnv = "MOAI_GR_BASE"

func grBase(t *testing.T) string {
	t.Helper()
	base := os.Getenv(grBaseEnv)
	if base == "" {
		t.Skipf("%s is not set; this base-ref guard needs the last absorbed integration commit", grBaseEnv)
	}
	return base
}

// grGit runs git in the repository root and returns stdout.
func grGit(t *testing.T, root string, args ...string) []byte {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, stderr.String())
	}
	return out
}

// grShow returns the file at the base ref; ok is false when it is absent there.
func grShow(t *testing.T, root, base, rel string) (string, bool) {
	t.Helper()
	cmd := exec.Command("git", "show", base+":"+rel)
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return "", false
	}
	return string(out), true
}

// grAllCopies lists every block-bearing target in both copies.
func grAllCopies() []string {
	var out []string
	for _, tg := range grTargets {
		out = append(out, tg.path, grTemplatePrefix+tg.path)
	}
	return out
}

// grPreservationFindings compares a stripped document with its base copy.
func grPreservationFindings(rel, stripped, base string) []string {
	if stripped == base {
		return nil
	}
	sl, bl := strings.Split(stripped, "\n"), strings.Split(base, "\n")
	for i := 0; i < len(sl) && i < len(bl); i++ {
		if sl[i] != bl[i] {
			return []string{fmt.Sprintf("%s: text outside the blocks differs from the base at line %d: %q vs %q", rel, i+1, sl[i], bl[i])}
		}
	}
	return []string{fmt.Sprintf("%s: text outside the blocks differs from the base (line count %d vs %d)", rel, len(sl), len(bl))}
}

// TestContractModeGuidedPreservation strips every block and requires the
// result to equal the base copy byte for byte (AC-GR-001).
func TestContractModeGuidedPreservation(t *testing.T) {
	t.Run("falsifier/word-changed-outside-block", func(t *testing.T) {
		base := "alpha\nbeta\n"
		cur := "alpha\n" + grBlockText("a", grCondition+" x") + "\nbetta\n"
		if f := grPreservationFindings("fixture.md", grStrip(cur), base); len(f) == 0 {
			t.Fatal("preservation check accepted a changed word outside the block")
		}
		if f := grPreservationFindings("fixture.md", grStrip("alpha\n"+grBlockText("a", grCondition+" x")+"\nbeta\n"), base); len(f) != 0 {
			t.Fatalf("preservation check rejected a clean additive block: %v", f)
		}
	})
	t.Run("tree", func(t *testing.T) {
		base := grBase(t)
		root := grRoot(t)
		stripped := 0
		for _, rel := range grAllCopies() {
			cur := grRead(t, root, rel)
			blocks, _ := grParse(rel, cur)
			stripped += len(blocks)
			want, ok := grShow(t, root, base, rel)
			if !ok {
				t.Errorf("%s: absent at the base ref", rel)
				continue
			}
			for _, f := range grPreservationFindings(rel, grStrip(cur), want) {
				t.Error(f)
			}
		}
		if stripped == 0 {
			t.Fatal("no block was stripped — empty sweep")
		}
		t.Logf("stripped %d blocks across %d copies", stripped, len(grAllCopies()))
	})
}

// grLineDelta is the multiset difference a−b of two documents' lines, sorted.
func grLineDelta(a, b string) []string {
	count := map[string]int{}
	for _, l := range strings.Split(a, "\n") {
		count[l]++
	}
	for _, l := range strings.Split(b, "\n") {
		count[l]--
	}
	var out []string
	for l, n := range count {
		for ; n > 0; n-- {
			out = append(out, l)
		}
	}
	sort.Strings(out)
	return out
}

// TestContractModeInheritedDivergence requires the local↔template divergence
// outside the blocks to equal the base's divergence (AC-GR-002 half).
func TestContractModeInheritedDivergence(t *testing.T) {
	base := grBase(t)
	root := grRoot(t)
	for _, tg := range grTargets {
		bl, ok1 := grShow(t, root, base, tg.path)
		bt, ok2 := grShow(t, root, base, grTemplatePrefix+tg.path)
		if !ok1 || !ok2 {
			t.Errorf("%s: a copy is absent at the base ref", tg.path)
			continue
		}
		cl := grStrip(grRead(t, root, tg.path))
		ct := grStrip(grRead(t, root, grTemplatePrefix+tg.path))
		if !slices.Equal(grLineDelta(cl, ct), grLineDelta(bl, bt)) || !slices.Equal(grLineDelta(ct, cl), grLineDelta(bt, bl)) {
			t.Errorf("%s: local↔template divergence outside the blocks changed from the base", tg.path)
		}
	}
}

// grAllowed reports whether a changed path is on the design.md §2 edit
// allowlist (template mirrors and the SPEC directory included).
func grAllowed(p string) bool {
	exact := []string{
		grSSOTPath,
		".claude/rules/moai/core/moai-mcp-tools-catalogue.md",
		".moai/config/sections/workflow.yaml",
		".moai/specs/SPEC-JEV-CORE-001/spec.md",
		"CLAUDE.local.md",
		"internal/contract/receipt.go",
		"internal/contract/receipt_test.go",
		"internal/contract/doc.go",
		"internal/cli/contract.go",
		"internal/template/contract_mode_blocks_test.go",
		"internal/template/contract_mode_guided_test.go",
		// Mechanical cascade of the skill edits: the embedded catalog carries
		// a whole-tree hash of the moai skill, regenerated by `make build`.
		"internal/template/catalog.yaml",
	}
	for _, tg := range grTargets {
		exact = append(exact, tg.path)
	}
	for _, e := range exact {
		if p == e || p == grTemplatePrefix+e {
			return true
		}
	}
	prefixes := []string{
		"internal/contract/receipt/",
		"internal/contract/kickoff/",
		"internal/contract/revoke/",
		"internal/contract/sign/",
		".moai/specs/SPEC-AUTONOMY-GATE-REWIRE-001/",
		// The card's evidence path: plan-audit reports and the verdict the
		// plan phase committed after the base. Evidence, not implementation.
		".moai/reports/t1236/",
	}
	for _, pre := range prefixes {
		if strings.HasPrefix(p, pre) {
			return true
		}
	}
	for _, stem := range []string{"contract_decide", "contract_kickoff_check", "contract_revoke"} {
		if strings.HasPrefix(p, "internal/cli/"+stem) && strings.HasSuffix(p, ".go") {
			return true
		}
	}
	return false
}

// TestContractModeChangeSetAllowlist requires every path changed since the
// base (tracked diff plus untracked files) to be on the allowlist (AC-GR-003).
func TestContractModeChangeSetAllowlist(t *testing.T) {
	t.Run("falsifier/forbidden-paths", func(t *testing.T) {
		for _, p := range []string{
			".claude/rules/moai/core/moai-constitution.md",
			".claude/rules/moai/core/zone-registry.md",
			".claude/agents/moai/manager-develop.md",
			".claude/output-styles/moai/moai.md",
			".claude/rules/moai/core/agent-common-protocol.md",
			".claude/rules/moai/workflow/spec-workflow.md",
			"internal/kanban/backlog.go",
			"internal/escalation/record.go",
		} {
			if grAllowed(p) {
				t.Errorf("allowlist admits forbidden path %s", p)
			}
		}
	})
	t.Run("tree", func(t *testing.T) {
		base := grBase(t)
		root := grRoot(t)
		changed := strings.Fields(string(grGit(t, root, "diff", "--name-only", base)))
		changed = append(changed, strings.Fields(string(grGit(t, root, "ls-files", "--others", "--exclude-standard")))...)
		if len(changed) == 0 {
			t.Fatal("no changed path since the base — the sweep measured nothing")
		}
		for _, p := range changed {
			if !grAllowed(p) {
				t.Errorf("changed path outside the allowlist: %s", p)
			}
		}
		t.Logf("%d changed paths checked", len(changed))
	})
}

// grExtract writes the tree at ref into dir with `git archive`.
func grExtract(t *testing.T, root, ref, dir string) {
	t.Helper()
	data := grGit(t, root, "archive", "--format=tar", ref)
	tr := tar.NewReader(bytes.NewReader(data))
	for {
		h, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return
		}
		if err != nil {
			t.Fatalf("read archive of %s: %v", ref, err)
		}
		target := filepath.Join(dir, filepath.FromSlash(h.Name))
		switch h.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				t.Fatal(err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				t.Fatal(err)
			}
			body, err := io.ReadAll(tr)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(target, body, 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func grValidate(t *testing.T, root string) constitution.ValidationResult {
	t.Helper()
	res, err := constitution.Validate(constitution.ValidateOptions{
		RegistryPath: filepath.Join(root, filepath.FromSlash(constitution.RegistryRelPath)),
		ProjectDir:   root,
	})
	if err != nil {
		var ve *constitution.ValidationError
		if !constitution.AsValidationError(err, &ve) {
			t.Fatalf("constitution validate %s: %v", root, err)
		}
	}
	if res.Skipped || res.Status == "" {
		t.Fatalf("constitution validate %s was skipped or empty (status %q)", root, res.Status)
	}
	return res
}

// grNonOK returns the sorted "SENTINEL id" pairs of a result.
func grNonOK(r constitution.ValidationResult) []string {
	var out []string
	for _, e := range r.Entries {
		if e.SentinelKey == "" || e.SentinelKey == "OK" {
			continue
		}
		out = append(out, e.SentinelKey+" "+e.ID)
	}
	sort.Strings(out)
	return slices.Compact(out)
}

// grConstitutionFindings compares the current result with the base: every
// non-OK pair must already exist at the base and no count may grow.
func grConstitutionFindings(base, cur constitution.ValidationResult) []string {
	var findings []string
	b := grNonOK(base)
	for _, p := range grNonOK(cur) {
		if !slices.Contains(b, p) {
			findings = append(findings, "new constitution finding not present at the base: "+p)
		}
	}
	if cur.DriftCount > base.DriftCount {
		findings = append(findings, fmt.Sprintf("drift_count %d > base %d", cur.DriftCount, base.DriftCount))
	}
	if cur.MissingCount > base.MissingCount {
		findings = append(findings, fmt.Sprintf("missing_count %d > base %d", cur.MissingCount, base.MissingCount))
	}
	if cur.UnregisteredCount > base.UnregisteredCount {
		findings = append(findings, fmt.Sprintf("unregistered_count %d > base %d", cur.UnregisteredCount, base.UnregisteredCount))
	}
	return findings
}

// grHardTargets are the always-loaded target copies whose [HARD] line sets
// must not grow.
func grHardTargets() []string {
	var out []string
	for _, tg := range grTargets {
		if tg.alwaysLoaded {
			out = append(out, tg.path, grTemplatePrefix+tg.path)
		}
	}
	return out
}

// grHardSet is the de-duplicated set of lines carrying the literal [HARD].
func grHardSet(text string) []string {
	var out []string
	for _, l := range strings.Split(text, "\n") {
		if strings.Contains(l, "[HARD]") {
			out = append(out, l)
		}
	}
	sort.Strings(out)
	return slices.Compact(out)
}

// grHardFindings reports [HARD] lines present now but absent at the base.
func grHardFindings(rel, base, cur string) []string {
	b := grHardSet(base)
	var findings []string
	for _, l := range grHardSet(cur) {
		if !slices.Contains(b, l) {
			findings = append(findings, fmt.Sprintf("%s: unregistered [HARD] line not present at the base: %q", rel, l))
		}
	}
	return findings
}

// grBaseDriftIDs are the DRIFT ids the base carries (the ledger's EV-6
// baseline). They describe the base tree, not a constant of the code: when the
// base is re-absorbed this set is re-measured and updated with it.
var grBaseDriftIDs = []string{
	"CONST-V3R2-013", "CONST-V3R2-014", "CONST-V3R2-015", "CONST-V3R2-016", "CONST-V3R2-017",
	"CONST-V3R2-033", "CONST-V3R2-049", "CONST-V3R2-152", "CONST-V3R2-153",
}

// TestContractModeConstitutionDriftNotIncreased requires the constitution
// validation not to get worse than the base in any category, and the
// always-loaded targets' [HARD] line sets not to grow (AC-GR-003).
func TestContractModeConstitutionDriftNotIncreased(t *testing.T) {
	base := grBase(t)
	root := grRoot(t)
	t.Setenv("MOAI_CONSTITUTION_SKIP_VALIDATE", "")

	baseDir := t.TempDir()
	grExtract(t, root, base, baseDir)
	baseRes := grValidate(t, baseDir)
	curRes := grValidate(t, root)

	var baseDrift []string
	for _, e := range baseRes.Entries {
		if e.SentinelKey == constitution.SentinelDrift {
			baseDrift = append(baseDrift, e.ID)
		}
	}
	sort.Strings(baseDrift)
	if !slices.Equal(baseDrift, grBaseDriftIDs) || baseRes.MissingCount != 0 || baseRes.UnregisteredCount != 0 {
		t.Fatalf("base premise broken: drift ids %v (want %v), missing %d, unregistered %d",
			baseDrift, grBaseDriftIDs, baseRes.MissingCount, baseRes.UnregisteredCount)
	}
	t.Logf("base non-OK pairs: %v (drift %d missing %d unregistered %d)", grNonOK(baseRes), baseRes.DriftCount, baseRes.MissingCount, baseRes.UnregisteredCount)
	t.Logf("current non-OK pairs: %v (drift %d missing %d unregistered %d)", grNonOK(curRes), curRes.DriftCount, curRes.MissingCount, curRes.UnregisteredCount)
	for _, f := range grConstitutionFindings(baseRes, curRes) {
		t.Error(f)
	}

	for _, rel := range grHardTargets() {
		bt, ok := grShow(t, root, base, rel)
		if !ok {
			t.Errorf("%s: absent at the base ref", rel)
			continue
		}
		ct := grRead(t, root, rel)
		t.Logf("%s: [HARD] lines base=%d current=%d", rel, len(grHardSet(bt)), len(grHardSet(ct)))
		for _, f := range grHardFindings(rel, bt, ct) {
			t.Error(f)
		}
	}

	t.Run("falsifier/registered-clause-removed", func(t *testing.T) {
		mut := t.TempDir()
		grExtract(t, root, "HEAD", mut)
		reg, err := constitution.LoadRegistry(filepath.Join(mut, filepath.FromSlash(constitution.RegistryRelPath)), mut)
		if err != nil {
			t.Fatal(err)
		}
		before := grValidate(t, mut)
		drifted := map[string]bool{}
		for _, e := range before.Entries {
			drifted[e.ID] = true
		}
		for _, e := range reg.Entries {
			if drifted[e.ID] || e.Clause == "" || strings.HasPrefix(e.Clause, "[SUPERSEDED") {
				continue
			}
			path := filepath.Join(mut, filepath.FromSlash(e.File))
			data, err := os.ReadFile(path)
			if err != nil || !strings.Contains(string(data), e.Clause) {
				continue
			}
			if err := os.WriteFile(path, []byte(strings.ReplaceAll(string(data), e.Clause, "")), 0o644); err != nil {
				t.Fatal(err)
			}
			after := grValidate(t, mut)
			f := grConstitutionFindings(baseRes, after)
			if len(f) == 0 {
				t.Fatalf("comparison accepted a removed registered clause (%s)", e.ID)
			}
			t.Logf("removed %s: drift %d → findings %v", e.ID, after.DriftCount, f)
			return
		}
		t.Fatal("no registered clause found to remove — the falsifier measured nothing")
	})
	t.Run("falsifier/unregistered-hard-line", func(t *testing.T) {
		rel := ".claude/rules/moai/core/askuser-protocol.md"
		bt, _ := grShow(t, root, base, rel)
		mut := grRead(t, root, rel) + "\n[HARD] Probe-only unregistered rule inserted by the falsifier.\n"
		f := grHardFindings(rel, bt, mut)
		if len(f) == 0 {
			t.Fatal("[HARD] comparison accepted an unregistered [HARD] line")
		}
		t.Logf("observed: %v", f)
	})
}

// TestContractModeAlwaysLoadedBudget bounds the context growth guided
// sessions pay (AC-GR-010, design.md §4).
func TestContractModeAlwaysLoadedBudget(t *testing.T) {
	base := grBase(t)
	root := grRoot(t)
	runes := func(s string) int { return utf8.RuneCountInString(s) }
	for _, prefix := range []string{"", grTemplatePrefix} {
		total := 0
		for _, tg := range grTargets {
			if !tg.alwaysLoaded {
				continue
			}
			rel := prefix + tg.path
			bt, ok := grShow(t, root, base, rel)
			if !ok {
				t.Errorf("%s: absent at the base ref", rel)
				continue
			}
			cur := grRead(t, root, rel)
			grow := runes(cur) - runes(bt)
			total += grow
			t.Logf("%s: +%d characters (now %d)", rel, grow, runes(cur))
			if runes(cur) >= 40000 {
				t.Errorf("%s: %d characters (limit < 40000)", rel, runes(cur))
			}
			blocks, _ := grParse(rel, cur)
			for _, b := range blocks {
				if n := runes(b.body); n > grAlwaysLoadedBlockCap {
					t.Errorf("%s: block %q is %d characters (cap %d)", rel, b.id, n, grAlwaysLoadedBlockCap)
				}
			}
		}
		t.Logf("always-loaded growth (%q copies): +%d characters", prefix, total)
		if total > 1500 {
			t.Errorf("always-loaded growth %d characters exceeds 1500 (%q copies)", total, prefix)
		}

		mcp := prefix + ".claude/rules/moai/core/moai-mcp-tools.md"
		if bt, ok := grShow(t, root, base, mcp); ok && bt != grRead(t, root, mcp) {
			t.Errorf("%s changed; it is not an edit target", mcp)
		}

		cat := prefix + ".claude/rules/moai/core/moai-mcp-tools-catalogue.md"
		bt, ok := grShow(t, root, base, cat)
		if !ok {
			t.Errorf("%s: absent at the base ref", cat)
			continue
		}
		cur := grRead(t, root, cat)
		if grow := runes(cur) - runes(bt); grow > 600 {
			t.Errorf("%s: +%d characters (cap 600)", cat, grow)
		}
		for _, key := range []string{"`mcp__moai__jev_ask`", "| Judgment (gated) |"} {
			lineOf := func(text string) string {
				for _, l := range strings.Split(text, "\n") {
					if strings.Contains(l, key) {
						return l
					}
				}
				return ""
			}
			if grow := runes(lineOf(cur)) - runes(lineOf(bt)); grow > 300 {
				t.Errorf("%s: row %s grew %d characters (cap 300)", cat, key, grow)
			}
		}
	}
}

// grKickoffClasses is the research §1.2 classification of every document
// carrying "Kickoff" at the base: E (emitter), R (reference), H (homonym),
// L (local-only harness). It is a measurement of the base tree and is
// re-measured whenever the base is re-absorbed.
var grKickoffClasses = map[string]string{
	"CLAUDE.md": "E",
	".claude/rules/moai/workflow/orchestration-mode-selection.md": "E",
	".claude/rules/moai/core/askuser-protocol.md":                 "E",
	".claude/rules/moai/workflow/goal-directive.md":               "E",
	".claude/skills/moai/SKILL.md":                                "E",
	".claude/skills/moai/workflows/moai.md":                       "E",
	".claude/skills/moai/workflows/plan.md":                       "E",
	".claude/skills/moai/workflows/plan/spec-assembly.md":         "E",
	".claude/skills/moai/workflows/run.md":                        "E",
	".claude/skills/moai/workflows/goal.md":                       "E",
	".claude/skills/moai/workflows/project/doc-generation.md":     "R",
	".claude/skills/moai/workflows/design.md":                     "R",
	".claude/skills/moai/workflows/factory.md":                    "R",
	".claude/skills/moai/workflows/run/mode-orchestration.md":     "R",
	".claude/skills/moai/workflows/run/phase-execution.md":        "R",
	".claude/skills/moai/workflows/run/task-decomposition.md":     "R",
	".claude/rules/moai/workflow/session-handoff.md":              "R",
	".claude/rules/moai/workflow/session-handoff-examples.md":     "R",
	".claude/rules/moai/workflow/spec-workflow.md":                "R",
	".claude/rules/moai/workflow/goal-directive-detail.md":        "R",
	".claude/rules/moai/workflow/dynamic-workflows.md":            "R",
	".claude/rules/moai/workflow/cadence-bridge.md":               "R",
	".claude/rules/moai/workflow/kanban-dispatch.md":              "R",
	".claude/rules/moai/workflow/cache-aware-execution.md":        "R",
	".claude/rules/moai/workflow/archived-agent-rejection.md":     "R",
	".claude/rules/moai/development/coding-standards.md":          "R",
	".claude/output-styles/moai/moai.md":                          "R",
	".claude/agents/moai/manager-develop.md":                      "R",
	".claude/agents/moai/manager-design.md":                       "R",
	".claude/agents/moai/plan-auditor.md":                         "R",
	".claude/skills/moai/workflows/e2e.md":                        "H",
	".claude/skills/moai/workflows/harness-build-entry.md":        "H",
	".claude/agents/harness/workflow-specialist.md":               "L",
}

// TestContractModeEmitterSites requires every base document carrying
// "Kickoff" to be classified and every emitter copy to carry a block
// (AC-GR-004).
func TestContractModeEmitterSites(t *testing.T) {
	base := grBase(t)
	root := grRoot(t)
	scopes := []string{"CLAUDE.md", ".claude/rules", ".claude/skills/moai", ".claude/output-styles", ".claude/agents"}
	var args []string
	args = append(args, "grep", "-l", "Kickoff", base, "--")
	for _, s := range scopes {
		args = append(args, s, grTemplatePrefix+s)
	}
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git grep at the base: %v", err)
	}
	var found []string
	for _, l := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		found = append(found, strings.TrimPrefix(l, base+":"))
	}
	if len(found) == 0 {
		t.Fatal("no Kickoff document found at the base — empty sweep")
	}
	for _, p := range found {
		if _, ok := grKickoffClasses[strings.TrimPrefix(p, grTemplatePrefix)]; !ok {
			t.Errorf("unclassified Kickoff document at the base: %s", p)
		}
	}
	emitters := 0
	for p, class := range grKickoffClasses {
		if class != "E" {
			continue
		}
		for _, rel := range []string{p, grTemplatePrefix + p} {
			emitters++
			blocks, _ := grParse(rel, grRead(t, root, rel))
			if len(blocks) == 0 {
				t.Errorf("emitter %s carries no contract-mode block", rel)
			}
		}
	}
	t.Logf("%d Kickoff documents at the base, %d emitter copies checked", len(found), emitters)
	if emitters != 20 {
		t.Errorf("emitter copies = %d, want 20", emitters)
	}
}
