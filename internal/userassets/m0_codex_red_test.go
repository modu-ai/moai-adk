// m0_codex_red_test.go — SPEC-USERASSET-DEPLOY-GUARD-001 M0, Codex
// conversion family: AC-006 (role TOML references converted), AC-007
// (13a user-skill references + 13b role TOML verbatim — the REPRODUCE OR
// REFUTE determination that gates M3 scope), AC-026 (unconvertible
// references reported per file), AC-016 (depends_on closure install).
//
// M0 discipline: observation only — no production change. The M0 finding on
// 13a/13b decides whether M3's 13-axis scope survives or shrinks to 7a.
package userassets

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/modu-ai/moai-adk/internal/template"
)

// referenceLadenTOML is the role-TOML source shape the ledger's 7a/13b axes
// name: Claude-tree references a Codex consumer can never resolve.
const referenceLadenTOML = `name = "manager-x"
instructions = """
Read the shared rules at .claude/rules/moai/core.md and the workflow
guide .claude/skills/moai/workflows/run.md. The product brief is CLAUDE.md.
"""
`

// referenceLadenSkill is the 13a shape: a skill directory source carrying
// CLAUDE.md and .claude/rules/moai/ references, installed to the Codex root.
const referenceLadenSkill = `---
name: moai-alpha
---
Use the rules under .claude/rules/moai/ and the brief CLAUDE.md.
`

// installWithTOML runs a first install over a fixture whose Codex role TOML
// carries the given bytes, returning the installed TOML's bytes.
func installWithTOML(t *testing.T, tomlBytes []byte) (installed []byte, installedPath string) {
	t.Helper()
	f := newFixture(t)
	f.src[".codex/agents/moai/manager-x.toml"] = &fstest.MapFile{Data: tomlBytes}
	if _, err := f.installer(t).Install(nil); err != nil {
		t.Fatalf("Install: %v", err)
	}
	installedPath = filepath.Join(f.home, ".codex", "agents", "manager-x.toml")
	data, err := os.ReadFile(installedPath)
	if err != nil {
		t.Fatalf("read installed TOML: %v", err)
	}
	return data, installedPath
}

// TestCodexAgentTOMLReferencesConverted — AC-006 (ledger 7a, REQ-CNV-001).
// Given a role TOML source carrying .claude/rules, .claude/skills and
// CLAUDE.md references, the installed TOML must carry the target harness's
// deployed faces — never a verbatim copy. RED-now reason: the install path's
// fileTarget (install.go:392) reads the source with a plain fs.ReadFile and
// records those bytes verbatim — the deploy-path converter
// (template.NormalizeCodexRoleForDeploy) is never wired here.
func TestCodexAgentTOMLReferencesConverted(t *testing.T) {
	installed, _ := installWithTOML(t, []byte(referenceLadenTOML))
	got := string(installed)

	converted := 0
	for _, pair := range [][2]string{
		{".claude/rules/moai/", ".moai/policies/"},
		{".claude/skills/moai/workflows/", ".moai/workflows/"},
		{"CLAUDE.md", "AGENTS.md"},
	} {
		if strings.Contains(got, pair[0]) {
			t.Errorf("RED (intended): installed TOML still carries the unconverted reference %q — the install path copied the source verbatim", pair[0])
		}
		if strings.Contains(got, pair[1]) {
			converted++
		}
	}
	if converted == 0 {
		t.Errorf("RED (intended): no converted face found in the installed TOML — NormalizeCodexRoleForDeploy is not wired into the install path (install.go:392)")
	}
}

// TestCodexUserSkillInstallNoClaudeOnlyRefs — AC-007 arm 13a (ledger 13a,
// REQ-CNV-001, reproduce-or-refute). Given a skill directory source carrying
// CLAUDE-only references, the installed COPY under the Codex skill root
// (.agents/skills) must not carry references the Codex side cannot resolve.
func TestCodexUserSkillInstallNoClaudeOnlyRefs(t *testing.T) {
	f := newFixture(t)
	f.src[".claude/skills/moai-alpha/SKILL.md"] = &fstest.MapFile{Data: []byte(referenceLadenSkill)}
	f.src[".agents/skills/moai-alpha/SKILL.md"] = &fstest.MapFile{Data: []byte(referenceLadenSkill)}
	if _, err := f.installer(t).Install(nil); err != nil {
		t.Fatalf("Install: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(f.home, ".agents", "skills", "moai-alpha", "SKILL.md"))
	if err != nil {
		t.Fatalf("read installed skill: %v", err)
	}
	got := string(data)
	for _, ref := range []string{".claude/rules/moai/", "CLAUDE.md"} {
		if strings.Contains(got, ref) {
			t.Errorf("REPRODUCED 13a: the installed Codex-root skill still carries the unconverted reference %q", ref)
		}
	}
}

// TestCodexRoleTOMLNotVerbatim — AC-007 arm 13b (ledger 13b, REQ-CNV-001,
// reproduce-or-refute; the live copy path is install.go:392 — the ledger's
// :584 coordinate drifted). The TOML must not be a byte-identical copy of a
// reference-laden source.
func TestCodexRoleTOMLNotVerbatim(t *testing.T) {
	installed, _ := installWithTOML(t, []byte(referenceLadenTOML))
	if string(installed) == referenceLadenTOML {
		t.Fatalf("REPRODUCED 13b: the installed role TOML is byte-identical to the reference-laden source — same verbatim-copy defect as AC-006 (install.go:392); 13b absorbs into AC-006")
	}
}

// TestCodexUnconvertibleReferenceReported — AC-026 (ledger 7a/13,
// REQ-CNV-002). Given an installed asset carrying a harness-specific
// reference with NO Codex-side counterpart (.claude/agents/moai/ — the Codex
// layout has no such face to map it to), the install RESULT REPORT must name
// that file per file. RED-now reason: the user-asset install path has no
// conversion wiring and no per-file report at all (REQ-CNV-002 lives only in
// acceptance at M3 today).
//
// Faithfulness note (recorded as a divergence in the M0 report): the report
// surface at M0 is the Result category set the cli summary renders
// (collisions/divergences/failures/shared/deferred). This test asserts the
// file name surfaces through that rendered report; M3 must land the
// REQ-CNV-002 report so this observable flips.
func TestCodexUnconvertibleReferenceReported(t *testing.T) {
	f := newFixture(t)
	unconvertible := `name = "manager-x"
instructions = "Coordinate with .claude/agents/moai/plan-auditor.md."
`
	f.src[".codex/agents/moai/manager-x.toml"] = &fstest.MapFile{Data: []byte(unconvertible)}
	res, err := f.installer(t).Install(nil)
	if err != nil {
		t.Fatalf("Install: %v", err)
	}

	var report strings.Builder
	writeM0ResultReport(&report, res)
	if !strings.Contains(report.String(), "codex-agents/manager-x.toml") {
		t.Fatalf("RED (intended): the install result report names no file for the unconvertible reference — report was:\n%s", report.String())
	}
}

// writeM0ResultReport renders the Result's report categories the way the
// cli layer's writeInstallSummary does — the M0-observable install report.
// M3 landing note: the unconverted-reference category (REQ-CNV-002) was
// added to the Result and to this render together — the M0 test comment
// sanctioned exactly this flip.
func writeM0ResultReport(b *strings.Builder, res *Result) {
	for _, c := range res.Collisions {
		b.WriteString("collision: " + c + "\n")
	}
	for _, d := range res.Divergences {
		b.WriteString("divergence: " + d + "\n")
	}
	for _, f := range res.Failures {
		b.WriteString("failure: " + f.Path + ": " + f.Reason + "\n")
	}
	for _, s := range res.SharedSurvivors {
		b.WriteString("kept: " + s + "\n")
	}
	for _, d := range res.DeferredDeps {
		b.WriteString("deferred: " + d + "\n")
	}
	for _, u := range res.Unconverted {
		b.WriteString("unconverted: " + u + "\n")
	}
}

// TestBundleDependsOnClosureInstall — AC-016 (ledger 8b, REQ-SRF-005).
// Given a selected bundle declaring depends_on, the dependency bundle's
// assets must install with it. RED-now reason: collectEntries
// (install.go:455-468) gathers core + the named packs only — no closure
// expansion over Pack.DependsOn. (The prune half of the ledger item follows
// this M0 discriminating result per acceptance.md; it is not asserted here.)
func TestBundleDependsOnClosureInstall(t *testing.T) {
	f := newFixture(t)
	gammaV1 := []byte("gamma skill body v1\n")
	f.cat.Catalog.OptionalPacks["extras"].DependsOn = []string{"extras2"}
	f.cat.Catalog.OptionalPacks["extras2"] = &template.Pack{
		Description: "dependency of extras",
		Skills: []template.Entry{
			{Name: "moai-gamma", Tier: "optional-pack:extras2", Path: "templates/.claude/skills/moai-gamma/", Version: "1.0.0"},
		},
	}
	f.src[".claude/skills/moai-gamma/SKILL.md"] = &fstest.MapFile{Data: gammaV1}
	f.src[".agents/skills/moai-gamma/SKILL.md"] = &fstest.MapFile{Data: gammaV1}

	if _, err := f.installer(t).Install([]string{"extras"}); err != nil {
		t.Fatalf("Install: %v", err)
	}
	if _, err := os.Stat(filepath.Join(f.home, ".claude", "skills", "moai-gamma")); err != nil {
		t.Fatalf("RED (intended): the depends_on dependency bundle (extras2) was not installed with the selected bundle (extras) — collectEntries has no closure expansion: %v", err)
	}
}
