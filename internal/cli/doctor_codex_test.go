package cli

// SPEC-CODEX-WIRING-001 M4 — the `moai doctor` "Codex Wiring" diagnostic
// (REQ-CW-010 / AC-CW-012). Advisory and fail-open (checkBinaryFreshness t184
// precedent): the check never blocks the rest of doctor, an inactive project
// is an informational skip, and drift is REPORTED (the wiring writer never
// repairs a user-owned surface).

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/codexwiring"
)

// wireProjectForDoctor wires a fresh temp project doctor should find HEALTHY,
// and returns its root.
//
// It carries a skill mirror because a real `moai init --agent codex` project
// does: wiring and the template deploy that creates `.agents/skills` happen on
// the same run. A wired root WITHOUT a mirror is a state the mirror diagnostic
// reports (SPEC-CODEX-MIRROR-DOCTOR-001 REQ-CMD-004), so leaving it out here
// would make every test in this file that only wants a quiet baseline carry an
// unrelated finding.
//
// The mirror is an EMPTY directory rather than a populated one, and no
// `.claude/skills` is created: a project with no skills has nothing to mirror,
// which observes as neither a finding nor a detail count. It also needs no
// symlink, so this fixture does not become windows-skipped for every caller.
// A test that wants the mirror ABSENT uses wireProjectWithoutMirror.
func wireProjectForDoctor(t *testing.T) string {
	t.Helper()
	root := wireProjectWithoutMirror(t)
	if err := os.MkdirAll(filepath.Join(root, ".agents", "skills"), 0o755); err != nil {
		t.Fatal(err)
	}
	return root
}

// wireProjectWithoutMirror wires a fresh temp project and returns its root,
// leaving `.agents/skills` absent — the state REQ-CMD-004 reports on.
func wireProjectWithoutMirror(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	var out, warn bytes.Buffer
	if _, err := codexwiring.Wire(root, &out, &warn); err != nil {
		t.Fatalf("Wire: %v", err)
	}
	return root
}

// stubMoaiLookup pins the PATH-resolution sub-check deterministically: tests
// install a fake that reports moai found (doctor only needs the verdict).
func stubMoaiLookup(t *testing.T, found bool) {
	t.Helper()
	orig := codexWiringLookPath
	codexWiringLookPath = func(string) (string, error) {
		if found {
			return "/usr/local/bin/moai", nil
		}
		return "", os.ErrNotExist
	}
	t.Cleanup(func() { codexWiringLookPath = orig })
}

// stubCodexLookup pins the PATH-resolution seam per binary name: the "moai"
// and "codex" sub-checks ask different questions, so a test that needs them to
// answer differently cannot use the verdict-only stubMoaiLookup.
func stubCodexLookup(t *testing.T, moaiFound, codexFound bool) {
	t.Helper()
	orig := codexWiringLookPath
	codexWiringLookPath = func(name string) (string, error) {
		found := moaiFound
		if name == "codex" {
			found = codexFound
		}
		if found {
			return "/usr/local/bin/" + name, nil
		}
		return "", os.ErrNotExist
	}
	t.Cleanup(func() { codexWiringLookPath = orig })
}

// stubCodexHome pins the home-directory resolution so the stale-skill
// sub-check never reads the developer's real ~/.codex/config.toml.
//
// BOTH inputs to resolveCodexHomeDir are pinned: the codexUserHomeDir seam
// (so no t.Setenv("HOME", …), which pollutes parallel tests) AND CODEX_HOME,
// which takes precedence over the seam. Pinning only the seam would let a
// developer's exported CODEX_HOME decide the verdict.
func stubCodexHome(t *testing.T, home string) {
	t.Helper()
	t.Setenv(codexHomeEnvVar, "") // blank is treated as unset by the resolver
	orig := codexUserHomeDir
	codexUserHomeDir = func() (string, error) { return home, nil }
	t.Cleanup(func() { codexUserHomeDir = orig })
}

// codexSkillEntrySpec is one [[skills.config]] entry to write out. Both keys
// are optional so a fixture can exercise an entry that declares no path (the
// empty-path guard) and one that declares no enabled key (the unspecified
// tri-state) — neither is expressible with plain string/bool fields.
type codexSkillEntrySpec struct {
	Path       string // written only when non-empty
	EnabledKey string // written verbatim when non-empty ("true", "false", `"true"`, …)
}

// writeCodexHomeConfig writes a config.toml under a fresh home carrying one
// [[skills.config]] entry per spec, and returns the home root.
func writeCodexHomeConfig(t *testing.T, entries []codexSkillEntrySpec) string {
	t.Helper()
	home := t.TempDir()
	dir := filepath.Join(home, ".codex")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	var sb strings.Builder
	sb.WriteString("model = \"gpt-5\"\n\n")
	for _, e := range entries {
		sb.WriteString("[[skills.config]]\n")
		if e.Path != "" {
			sb.WriteString("path = \"" + e.Path + "\"\n")
		}
		if e.EnabledKey != "" {
			sb.WriteString("enabled = " + e.EnabledKey + "\n")
		}
		sb.WriteString("\n")
	}
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(sb.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	return home
}

// codexDetailText collapses Detail's line wrapping back onto one line so an
// ORDERED-PHRASE assertion is unaffected by where a wrap happened to fall.
// Wrap position is a rendering concern; the phrase order is the contract.
func codexDetailText(c DiagnosticCheck) string {
	return strings.Join(strings.Fields(c.Detail), " ")
}

// absentSkillPath derives a definitely-absent ABSOLUTE path under a fresh
// temp root. It must never be a hardcoded leading-slash literal: on a Windows
// host such a literal is not filepath.IsAbs, so once shape classification
// lands (SPEC-CODEX-SKILL-PATH-001) it would reclassify as relative there and
// silently corrupt the regression judgment.
func absentSkillPath(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "definitely-absent-"+name, "SKILL.md")
}

// liveSkillFile writes a SKILL.md that exists, and returns its path.
func liveSkillFile(t *testing.T) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "SKILL.md")
	if err := os.WriteFile(p, []byte("# skill\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

// TestCheckCodexWiring_UnwiredWithCodexInstalledWarns verifies the branch the
// silent skip used to swallow: codex resolves on PATH but the project carries
// no wiring, so the MCP server is unregistered and the hooks cannot fire here.
// The action directive must ride in Message — Detail renders only under
// --verbose, and a plain `moai doctor` has to show it.
func TestCheckCodexWiring_UnwiredWithCodexInstalledWarns(t *testing.T) {
	stubCodexLookup(t, true, true)
	stubCodexHome(t, t.TempDir())
	check := checkCodexWiring(t.TempDir(), false)
	if check.Status != uikit.CheckWarn {
		t.Errorf("unwired project with codex installed status = %v, want Warn: %+v", check.Status, check)
	}
	if !strings.Contains(check.Message, "moai init --agent codex") {
		t.Errorf("action directive missing from Message (Detail is --verbose-only): %+v", check)
	}
	// The absent paths are evidence, not a directive, so they ride in Detail
	// — naming both in Message is what pushed the row past the panel width.
	for _, want := range []string{codexwiring.HooksRelPath, codexwiring.ConfigRelPath} {
		if !strings.Contains(codexDetailText(check), want) {
			t.Errorf("Detail does not name the absent path %q: %+v", want, check)
		}
	}
	// The claim stays scoped to what was read: only the two PROJECT files were
	// inspected, so a machine-wide "the MoAI MCP server is not registered" is
	// an unobserved premise and must not appear.
	if strings.Contains(check.Message+" "+codexDetailText(check), "the MoAI MCP server is not registered") {
		t.Errorf("finding asserts a machine-wide registration state it never observed: %+v", check)
	}
}

// TestCheckCodexWiring_StaleHomeSkillsReported verifies the second silence:
// ~/.codex/config.toml [[skills.config]] entries whose path no longer exists
// are reported, quantified, and split by enabled state (an enabled missing
// path is live breakage; a disabled one is stale garbage).
// The assertions here are deliberately ORDERED PHRASES on the correct field,
// never bare substrings over Message+Detail. A search for "3" and "4" over the
// concatenation passes just as happily when the numerator and denominator are
// transposed, and a search over the concatenation cannot express "this must be
// in Message" at all — both mutations survived that formulation.
func TestCheckCodexWiring_StaleHomeSkillsReported(t *testing.T) {
	stubCodexLookup(t, true, true)
	live := liveSkillFile(t)
	home := writeCodexHomeConfig(t, []codexSkillEntrySpec{
		{Path: live, EnabledKey: "true"},
		{Path: absentSkillPath(t, "moai-a"), EnabledKey: "true"},
		{Path: absentSkillPath(t, "moai-b"), EnabledKey: "false"},
		{Path: absentSkillPath(t, "moai-c"), EnabledKey: "false"},
		{EnabledKey: "true"}, // declares no path — counted in the total, never as missing
	})
	stubCodexHome(t, home)

	check := checkCodexWiring(wireProjectForDoctor(t), false)
	if check.Status != uikit.CheckWarn {
		t.Errorf("stale home skills status = %v, want Warn: %+v", check.Status, check)
	}
	if !strings.Contains(check.Message, "config.toml") {
		t.Errorf("summary does not name the config file: %q", check.Message)
	}
	// The summary carries the missing COUNT, as an ordered phrase: "3 stale"
	// still reads as a count if the split were transposed, so the count and
	// its noun are asserted together.
	if !strings.Contains(check.Message, "3 stale skill entries") {
		t.Errorf("summary does not carry the missing count as an ordered phrase: %q", check.Message)
	}

	// Detail carries the denominator, the declared split, and the directive.
	// Each is an ORDERED phrase: transposing numerator and denominator, or
	// transposing the enabled and disabled counts, breaks the match.
	for _, want := range []string{
		"declares 5 [[skills.config]] entries",
		"3 with a path that no longer exists",
		"(1 enabled, 2 disabled, 0 unspecified)",
		"remove the stale entries or restore the skill files",
	} {
		if !strings.Contains(codexDetailText(check), want) {
			t.Errorf("Detail does not carry %q: %q", want, check.Detail)
		}
	}
	// The transposed forms must be ABSENT — the positive assertions above
	// would otherwise still pass on a message that merely contains the digits.
	for _, forbidden := range []string{
		"declares 3 [[skills.config]]",
		"5 with a path that no longer exists",
		"(2 enabled, 1 disabled",
	} {
		if strings.Contains(codexDetailText(check), forbidden) {
			t.Errorf("Detail carries the transposed form %q: %q", forbidden, check.Detail)
		}
	}
}

// TestCheckCodexWiring_EmptyPathEntryNotCountedMissing isolates the empty-path
// guard by its EFFECT on the count. os.Stat("") fails with ENOENT, so removing
// the guard turns an entry that declares nothing into a missing path — the
// same fixture with and without the guard differs only in this number.
func TestCheckCodexWiring_EmptyPathEntryNotCountedMissing(t *testing.T) {
	stubCodexLookup(t, true, true)
	home := writeCodexHomeConfig(t, []codexSkillEntrySpec{
		{Path: absentSkillPath(t, "only-one"), EnabledKey: "true"},
		{EnabledKey: "true"},  // no path key
		{EnabledKey: "false"}, // no path key
	})
	stubCodexHome(t, home)

	check := checkCodexWiring(wireProjectForDoctor(t), false)
	if !strings.Contains(check.Message, "1 stale skill entry") {
		t.Errorf("path-less entries inflated the missing count: %q", check.Message)
	}
	if !strings.Contains(codexDetailText(check), "1 with a path that no longer exists") {
		t.Errorf("Detail count wrong — path-less entries must not count as missing: %q", check.Detail)
	}
	if !strings.Contains(codexDetailText(check), "declares 3 [[skills.config]] entries") {
		t.Errorf("path-less entries must still count in the total: %q", check.Detail)
	}
}

// TestCheckCodexWiring_UnspecifiedEnabledReportedSeparately verifies an entry
// declaring no `enabled` key is reported as unspecified rather than folded
// into either side — the repository has not observed Codex's default, so
// claiming one would be an unverified premise in a user-facing message.
func TestCheckCodexWiring_UnspecifiedEnabledReportedSeparately(t *testing.T) {
	stubCodexLookup(t, true, true)
	home := writeCodexHomeConfig(t, []codexSkillEntrySpec{
		{Path: absentSkillPath(t, "a"), EnabledKey: ""},       // no enabled key
		{Path: absentSkillPath(t, "b"), EnabledKey: `"true"`}, // quoted string, still true
	})
	stubCodexHome(t, home)

	check := checkCodexWiring(wireProjectForDoctor(t), false)
	if !strings.Contains(codexDetailText(check), "(1 enabled, 0 disabled, 1 unspecified)") {
		t.Errorf("declared split wrong — quoted true must not demote to disabled, absent must not either: %q", check.Detail)
	}
}

// TestCheckCodexWiring_CodexHomeHonoured verifies CODEX_HOME decides which
// config the sub-check reads. The seam points at a home whose config is
// stale; CODEX_HOME points at a directory with no config at all, and the
// resolver's precedence means the finding must be SILENT. Reading the seam's
// file here would warn about a file Codex never reads.
func TestCheckCodexWiring_CodexHomeHonoured(t *testing.T) {
	stubCodexLookup(t, true, true)
	staleHome := writeCodexHomeConfig(t, []codexSkillEntrySpec{
		{Path: absentSkillPath(t, "moai-a"), EnabledKey: "true"},
	})
	stubCodexHome(t, staleHome)
	t.Setenv(codexHomeEnvVar, t.TempDir()) // empty CODEX_HOME wins

	check := checkCodexWiring(wireProjectForDoctor(t), false)
	if check.Status != uikit.CheckOK {
		t.Errorf("empty CODEX_HOME status = %v, want OK — the seam's stale config must not be read: %+v", check.Status, check)
	}
	if strings.Contains(check.Message+" "+codexDetailText(check), "stale skill") {
		t.Errorf("CODEX_HOME ignored — finding came from the default home: %+v", check)
	}
}

// TestCheckCodexWiring_CodexHomeConfigRead is the other half of the same
// precedence: a stale config INSIDE CODEX_HOME must be found, so the check
// examines the file Codex actually reads.
func TestCheckCodexWiring_CodexHomeConfigRead(t *testing.T) {
	stubCodexLookup(t, true, true)
	stubCodexHome(t, t.TempDir()) // default home carries no config
	envHome := filepath.Join(t.TempDir(), "codex-home")
	if err := os.MkdirAll(envHome, 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := fmt.Sprintf("[[skills.config]]\npath = %q\nenabled = true\n", absentSkillPath(t, "env"))
	if err := os.WriteFile(filepath.Join(envHome, "config.toml"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv(codexHomeEnvVar, envHome)

	check := checkCodexWiring(wireProjectForDoctor(t), false)
	if check.Status != uikit.CheckWarn {
		t.Fatalf("stale config under CODEX_HOME was not read: %+v", check)
	}
	if !strings.Contains(check.Message, codexHomeConfigEnvDisplay) {
		t.Errorf("summary names the default home for an env-sourced config: %q", check.Message)
	}
	if !strings.Contains(codexDetailText(check), envHome) {
		t.Errorf("Detail does not cite the resolved path it actually read: %q", check.Detail)
	}
}

// TestCheckCodexWiring_IndeterminateStatNotMissing verifies a stat that fails
// for a reason OTHER than non-existence is not counted as a missing path. The
// finding advises REMOVING the entry, so a permission error or a symlink loop
// counted as absent would advise deleting a healthy registration.
func TestCheckCodexWiring_IndeterminateStatNotMissing(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation needs privileges on windows")
	}
	stubCodexLookup(t, true, true)
	dir := t.TempDir()
	loopA := filepath.Join(dir, "loopA")
	loopB := filepath.Join(dir, "loopB")
	if err := os.Symlink(loopB, loopA); err != nil {
		t.Skipf("symlink unsupported here: %v", err)
	}
	if err := os.Symlink(loopA, loopB); err != nil {
		t.Skipf("symlink unsupported here: %v", err)
	}
	if _, err := os.Stat(loopA); err == nil || errors.Is(err, fs.ErrNotExist) {
		t.Skipf("symlink loop did not produce a non-ENOENT stat error: %v", err)
	}

	home := writeCodexHomeConfig(t, []codexSkillEntrySpec{
		{Path: absentSkillPath(t, "real-miss"), EnabledKey: "true"},
		{Path: loopA, EnabledKey: "true"},
	})
	stubCodexHome(t, home)

	check := checkCodexWiring(wireProjectForDoctor(t), false)
	if !strings.Contains(check.Message, "1 stale skill entry") {
		t.Errorf("indeterminate stat inflated the missing count: %q", check.Message)
	}
	if !strings.Contains(codexDetailText(check), "1 could not be checked and are NOT counted as missing") {
		t.Errorf("Detail does not disclose the unchecked entry: %q", check.Detail)
	}
}

// TestCheckCodexWiring_DirectoryPathNotMissing verifies a declared path that
// resolves to a DIRECTORY is not reported as missing. It exists; whether
// Codex accepts it is not observed here, and reporting it would advise
// removing a registration on a guess.
func TestCheckCodexWiring_DirectoryPathNotMissing(t *testing.T) {
	stubCodexLookup(t, true, true)
	home := writeCodexHomeConfig(t, []codexSkillEntrySpec{
		{Path: t.TempDir(), EnabledKey: "true"},
	})
	stubCodexHome(t, home)

	check := checkCodexWiring(wireProjectForDoctor(t), false)
	if check.Status != uikit.CheckOK {
		t.Errorf("a directory path was reported as missing: %+v", check)
	}
}

// TestCheckCodexWiring_MessageWidthStaysInBand is the regression guard for the
// panel-width blowout: the doctor box sizes itself to its widest row, so one
// long Message widens every other row past the terminal. The two new branches
// are exercised together — the worst realistic co-occurrence — and Message is
// measured in runes against the band the committed golden rows already occupy.
func TestCheckCodexWiring_MessageWidthStaysInBand(t *testing.T) {
	stubCodexLookup(t, true, true)
	absentRoot := t.TempDir() // one root for all 49 definitely-absent entries
	entries := make([]codexSkillEntrySpec, 0, 49)
	for i := 0; i < 49; i++ { // the real-world census that produced the 1272-column panel
		entries = append(entries, codexSkillEntrySpec{
			Path:       filepath.Join(absentRoot, fmt.Sprintf("moai-skill-%02d", i), "SKILL.md"),
			EnabledKey: "false",
		})
	}
	home := writeCodexHomeConfig(t, entries)
	stubCodexHome(t, home)

	// Unwired project + stale home config: both new findings at once.
	check := checkCodexWiring(t.TempDir(), false)
	if check.Status != uikit.CheckWarn {
		t.Fatalf("premise broken — expected both findings: %+v", check)
	}
	if n := utf8.RuneCountInString(check.Message); n > codexMessageWidthCeiling {
		t.Errorf("Message is %d runes, over the %d-rune band the existing doctor rows occupy: %q",
			n, codexMessageWidthCeiling, check.Message)
	}
	// The bound must not have been bought by dropping the directive.
	if !strings.Contains(check.Message, initCodexAdvice) {
		t.Errorf("width bound dropped the action directive from Message: %q", check.Message)
	}
}

// doctorGoldenPanelWidth is the rune width of the committed doctor golden
// panels (internal/cli/testdata/doctor-{light,dark,nocolor}.golden — every
// box-border row measures 152 runes). It is the width the rendered doctor
// output already occupies, and therefore the ceiling this check must not
// push past.
const doctorGoldenPanelWidth = 152

// TestCheckCodexWiring_RenderedPanelStaysInBand covers the RENDER, which is
// where the width defect actually surfaced and where a Message-length
// assertion alone would not have caught it: the panel sizes itself to its
// widest row, and Detail becomes a row of its own under --verbose.
//
// A golden FIXTURE is the wrong vehicle here. The warn branch's Detail cites
// the resolved config path, which is a t.TempDir() under test and the user's
// real home in production, so a byte-comparison snapshot would either be
// machine-dependent — the exact hermeticity failure the golden pin was added
// to fix — or need the very path normalization that would blank out the
// evidence being asserted. The invariant that broke is the WIDTH, so the
// width is asserted directly, on the real rendered output, in both modes.
func TestCheckCodexWiring_RenderedPanelStaysInBand(t *testing.T) {
	stubCodexLookup(t, true, true)
	absentRoot := t.TempDir() // one root for all 49 definitely-absent entries
	entries := make([]codexSkillEntrySpec, 0, 49)
	for i := 0; i < 49; i++ {
		entries = append(entries, codexSkillEntrySpec{
			Path:       filepath.Join(absentRoot, fmt.Sprintf("moai-skill-%02d", i), "SKILL.md"),
			EnabledKey: "false",
		})
	}
	stubCodexHome(t, writeCodexHomeConfig(t, entries))

	for _, verbose := range []bool{false, true} {
		check := checkCodexWiring(t.TempDir(), verbose)
		if check.Status != uikit.CheckWarn {
			t.Fatalf("verbose=%v premise broken — expected the warn branch: %+v", verbose, check)
		}
		var buf bytes.Buffer
		rendered := renderDoctorGroups(&buf, []checkGroup{{
			title:  "Codex",
			checks: []DiagnosticCheck{check},
		}}, verbose, resolveTheme())

		widest, row := 0, ""
		for _, ln := range strings.Split(rendered, "\n") {
			if n := utf8.RuneCountInString(stripDoctorANSI(ln)); n > widest {
				widest, row = n, ln
			}
		}
		if widest > doctorGoldenPanelWidth {
			t.Errorf("verbose=%v rendered panel is %d runes wide, over the %d-rune band the committed golden panels occupy; widest row: %q",
				verbose, widest, doctorGoldenPanelWidth, row)
		}
	}
}

// stripDoctorANSI removes SGR escape sequences so a rendered row is measured
// in visible runes rather than in styling bytes.
func stripDoctorANSI(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); {
		if s[i] == 0x1b && i+1 < len(s) && s[i+1] == '[' {
			j := i + 2
			for j < len(s) && s[j] != 'm' {
				j++
			}
			if j < len(s) {
				i = j + 1
				continue
			}
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

// TestJoinCodexSummariesDropsTailNotDirective verifies the truncation rule
// directly: when the summaries do not fit, the TAIL is dropped and the count
// disclosed — the leading, directive-bearing summary always survives.
func TestJoinCodexSummariesDropsTailNotDirective(t *testing.T) {
	long := strings.Repeat("x", 90)
	got := joinCodexSummaries([]codexFinding{
		{summary: "lead — " + initCodexAdvice},
		{summary: long},
		{summary: long},
	})
	if n := utf8.RuneCountInString(got); n > codexMessageWidthCeiling {
		t.Errorf("joined summary is %d runes, over the %d ceiling: %q", n, codexMessageWidthCeiling, got)
	}
	if !strings.Contains(got, initCodexAdvice) {
		t.Errorf("leading directive dropped: %q", got)
	}
	if !strings.Contains(got, "(+2 more, see --verbose)") {
		t.Errorf("dropped findings not disclosed: %q", got)
	}
}

// TestCheckCodexWiring_HealthyHomeSkillsNoFinding verifies the no-false-
// positive path: every declared skill path exists, so the sub-check is silent
// and a healthy project stays OK.
func TestCheckCodexWiring_HealthyHomeSkillsNoFinding(t *testing.T) {
	stubCodexLookup(t, true, true)
	home := writeCodexHomeConfig(t, []codexSkillEntrySpec{{Path: liveSkillFile(t), EnabledKey: "true"}})
	stubCodexHome(t, home)

	check := checkCodexWiring(wireProjectForDoctor(t), false)
	if check.Status != uikit.CheckOK {
		t.Errorf("all-present skill paths status = %v, want OK: %+v", check.Status, check)
	}
}

// TestCheckCodexWiring_AbsentHomeConfigSilent verifies an absent
// ~/.codex/config.toml degrades to a silent skip rather than a finding
// (fail-open: an unreadable input is never a failure).
func TestCheckCodexWiring_AbsentHomeConfigSilent(t *testing.T) {
	stubCodexLookup(t, true, true)
	stubCodexHome(t, t.TempDir()) // no .codex/config.toml inside
	check := checkCodexWiring(wireProjectForDoctor(t), false)
	if check.Status != uikit.CheckOK {
		t.Errorf("absent home config status = %v, want OK (silent skip): %+v", check.Status, check)
	}
	if strings.Contains(check.Message+" "+codexDetailText(check), "skills.config") {
		t.Errorf("absent home config produced a skills finding: %+v", check)
	}
}

// TestCheckCodexWiring_ClaudeOnlyMachineStaysSilent verifies the un-nagging
// invariant from the other side: no wiring AND no codex binary means the
// stale-skill sub-check never runs either, even with a stale home config.
func TestCheckCodexWiring_ClaudeOnlyMachineStaysSilent(t *testing.T) {
	stubCodexLookup(t, true, false)
	home := writeCodexHomeConfig(t, []codexSkillEntrySpec{{Path: absentSkillPath(t, "moai-a"), EnabledKey: "true"}})
	stubCodexHome(t, home)

	check := checkCodexWiring(t.TempDir(), false)
	if check.Status != uikit.CheckOK {
		t.Errorf("claude-only machine status = %v, want OK: %+v", check.Status, check)
	}
	if strings.Contains(check.Message+" "+codexDetailText(check), "skills.config") {
		t.Errorf("claude-only machine was nagged about home skills: %+v", check)
	}
}

// TestCheckCodexWiring_InactiveProjectInformationalSkip verifies a project
// without wiring files reports an informational skip (CheckOK), never a
// failure (AC-CW-012 third clause).
func TestCheckCodexWiring_InactiveProjectInformationalSkip(t *testing.T) {
	stubMoaiLookup(t, false) // even a moai-less machine must not turn skip into failure
	stubCodexHome(t, t.TempDir())
	check := checkCodexWiring(t.TempDir(), false)
	if check.Status != uikit.CheckOK {
		t.Errorf("inactive project status = %v, want OK (informational skip): %+v", check.Status, check)
	}
	if !strings.Contains(strings.ToLower(check.Message), "skip") && !strings.Contains(strings.ToLower(check.Message), "not wired") {
		t.Errorf("skip message should say so: %+v", check)
	}
}

// TestCheckCodexWiring_HealthyProjectOK verifies a freshly wired project
// with moai on PATH passes clean (AC-CW-012 first clause).
func TestCheckCodexWiring_HealthyProjectOK(t *testing.T) {
	stubMoaiLookup(t, true)
	stubCodexHome(t, t.TempDir())
	check := checkCodexWiring(wireProjectForDoctor(t), false)
	if check.Status != uikit.CheckOK {
		t.Errorf("healthy project status = %v, want OK: %+v", check.Status, check)
	}
}

// TestCheckCodexWiring_DivergenceAdvisesReTrust verifies the sidecar-hash
// divergence path: an unauthorized hooks.json edit is reported WITH the
// `/hooks to re-trust` action directive (AC-CW-012 second clause; the advice
// is an action instruction, never a claim about Codex's internal state).
func TestCheckCodexWiring_DivergenceAdvisesReTrust(t *testing.T) {
	stubMoaiLookup(t, true)
	stubCodexHome(t, t.TempDir())
	root := wireProjectForDoctor(t)
	hooksPath := filepath.Join(root, codexwiring.HooksRelPath)
	raw, err := os.ReadFile(hooksPath)
	if err != nil {
		t.Fatal(err)
	}
	tampered := string(raw) + "\n" // any byte change diverges the sidecar hash
	if err := os.WriteFile(hooksPath, []byte(tampered), 0o644); err != nil {
		t.Fatal(err)
	}

	check := checkCodexWiring(root, false)
	if check.Status != uikit.CheckWarn {
		t.Errorf("diverged project status = %v, want Warn: %+v", check.Status, check)
	}
	combined := check.Message + " " + check.Detail
	if !strings.Contains(combined, "/hooks to re-trust") {
		t.Errorf("divergence must carry the /hooks to re-trust directive: %+v", check)
	}
}

// TestCheckCodexWiring_ValidationFailureReported verifies a whitelist
// violation in the on-disk hooks.json is surfaced (Codex would silently
// disable the file — doctor is the observability backstop).
func TestCheckCodexWiring_ValidationFailureReported(t *testing.T) {
	stubMoaiLookup(t, true)
	stubCodexHome(t, t.TempDir())
	root := t.TempDir()
	codexDir := filepath.Join(root, ".codex")
	if err := os.MkdirAll(codexDir, 0o755); err != nil {
		t.Fatal(err)
	}
	bad := []byte("{\n  \"version\": 1,\n  \"hooks\": {}\n}\n")
	if err := os.WriteFile(filepath.Join(codexDir, "hooks.json"), bad, 0o644); err != nil {
		t.Fatal(err)
	}

	check := checkCodexWiring(root, false)
	if check.Status != uikit.CheckWarn {
		t.Errorf("violating hooks.json status = %v, want Warn: %+v", check.Status, check)
	}
	if !strings.Contains(check.Message+" "+codexDetailText(check), "version") {
		t.Errorf("diagnostic does not name the violating key: %+v", check)
	}
}

// TestCheckCodexWiring_MoaiNotOnPathReported verifies the PATH-resolution
// sub-check: wiring without a resolvable moai binary means the generated
// hook commands cannot fire.
func TestCheckCodexWiring_MoaiNotOnPathReported(t *testing.T) {
	stubMoaiLookup(t, false)
	stubCodexHome(t, t.TempDir())
	check := checkCodexWiring(wireProjectForDoctor(t), false)
	if check.Status != uikit.CheckWarn {
		t.Errorf("moai-not-on-PATH status = %v, want Warn: %+v", check.Status, check)
	}
	if !strings.Contains(strings.ToLower(check.Message+" "+codexDetailText(check)), "path") {
		t.Errorf("diagnostic does not mention PATH: %+v", check)
	}
}

// TestCheckCodexWiring_ConfigTableDriftReported verifies a user-modified
// [mcp_servers.moai] table is REPORTED (byte-invariant writer, doctor
// reports — REQ-CW-005).
func TestCheckCodexWiring_ConfigTableDriftReported(t *testing.T) {
	stubMoaiLookup(t, true)
	stubCodexHome(t, t.TempDir())
	root := wireProjectForDoctor(t)
	cfgPath := filepath.Join(root, codexwiring.ConfigRelPath)
	raw, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	drifted := strings.Replace(string(raw), `command = "moai"`, `command = "my-custom-moai"`, 1)
	if drifted == string(raw) {
		t.Fatal("drift substitution found nothing — premise broken")
	}
	if err := os.WriteFile(cfgPath, []byte(drifted), 0o644); err != nil {
		t.Fatal(err)
	}

	check := checkCodexWiring(root, false)
	if check.Status != uikit.CheckWarn {
		t.Errorf("drifted config table status = %v, want Warn: %+v", check.Status, check)
	}
	if !strings.Contains(strings.ToLower(check.Message+" "+codexDetailText(check)), "mcp_servers.moai") {
		t.Errorf("diagnostic does not name the drifted table: %+v", check)
	}
}

// TestDoctor_CodexWiringRegistered verifies the check is registered in the
// Workspace group of runGroupedChecksObserved (the --check filter reaches
// it — the registration itself is the AC-CW-012 surface `moai doctor` grep
// relies on).
func TestDoctor_CodexWiringRegistered(t *testing.T) {
	stubMoaiLookup(t, true)
	stubCodexHome(t, t.TempDir())
	groups := runGroupedChecks(false, "Codex Wiring")
	var found bool
	for _, g := range groups {
		for _, c := range g.checks {
			if c.Name == "Codex Wiring" {
				found = true
			}
		}
	}
	if !found {
		t.Error("\"Codex Wiring\" check not reachable via runGroupedChecks — not registered in the Workspace group")
	}
}

// --- SPEC-CODEX-SKILL-PATH-001: declared-path shape resolution ---
//
// The stale-skill sub-check used to stat e.Path RAW, so a ~-relative,
// relative, or backslash-bearing declaration all fell into ENOENT and were
// counted "missing" — advising the user to delete registrations that may be
// perfectly valid. The tests below assert the POST-fix contract: only
// determinately-resolvable shapes (absolute, and ~-relative after expansion
// through the codexUserHomeDir seam) can feed the missing count; relative and
// oddly-formed declarations are reported as their own classifications and are
// never attached to the remove directive.

// t468HomeSkillFile places a SKILL.md at home/rel so a ~-relative
// declaration has an expansion target that EXISTS under the home the seam
// pins.
func t468HomeSkillFile(t *testing.T, home, rel string) {
	t.Helper()
	dir := filepath.Join(home, filepath.Dir(rel))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, rel), []byte("# skill\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestCodexSkillPath_HomeRelativeExistingNotMissing (AC-CSP-001-01): a
// ~/prefixed entry whose expansion target exists under the pinned user home
// is NOT counted missing and emits no finding.
func TestCodexSkillPath_HomeRelativeExistingNotMissing(t *testing.T) {
	stubCodexLookup(t, true, true)
	home := writeCodexHomeConfig(t, []codexSkillEntrySpec{
		{Path: "~/t468-live/SKILL.md", EnabledKey: "true"},
	})
	t468HomeSkillFile(t, home, "t468-live/SKILL.md")
	stubCodexHome(t, home)

	check := checkCodexWiring(wireProjectForDoctor(t), false)
	if check.Status != uikit.CheckOK {
		t.Errorf("an existing ~-relative registration was reported missing: %+v", check)
	}
	if strings.Contains(check.Message+" "+codexDetailText(check), "stale skill") {
		t.Errorf("existing ~-relative registration produced a stale finding: %+v", check)
	}
}

// TestCodexSkillPath_HomeRelativeMissingStillCounted (AC-CSP-001-02): the
// other half of the expansion contract — a ~/prefixed entry whose expansion
// target does NOT exist IS counted missing. Expansion must not create false
// negatives either.
func TestCodexSkillPath_HomeRelativeMissingStillCounted(t *testing.T) {
	stubCodexLookup(t, true, true)
	home := writeCodexHomeConfig(t, []codexSkillEntrySpec{
		{Path: "~/t468-definitely-absent/SKILL.md", EnabledKey: "true"},
	})
	stubCodexHome(t, home)

	check := checkCodexWiring(wireProjectForDoctor(t), false)
	if check.Status != uikit.CheckWarn {
		t.Fatalf("a ~-relative entry with no expansion target was not reported: %+v", check)
	}
	if !strings.Contains(check.Message, "1 stale skill entry") {
		t.Errorf("summary does not count the missing ~-relative entry: %q", check.Message)
	}
	if !strings.Contains(codexDetailText(check), "remove the stale entries") {
		t.Errorf("missing ~-relative entry lost the remove directive: %q", check.Detail)
	}
}

// TestCodexSkillPath_RelativeNotMissingDistinctClassification
// (AC-CSP-001-03): a relative declaration is NOT counted missing, is
// surfaced in Detail as its own relative classification, and the remove
// directive stays bound to the genuinely-missing count only. The resolution
// base for a relative path is unobserved in this repository, so the check
// must not guess one.
func TestCodexSkillPath_RelativeNotMissingDistinctClassification(t *testing.T) {
	stubCodexLookup(t, true, true)
	home := writeCodexHomeConfig(t, []codexSkillEntrySpec{
		// The anchor forces a finding so Detail renders at all; without a
		// missing entry the sub-check is silent and there is no Detail to
		// carry the classification in.
		{Path: absentSkillPath(t, "anchor"), EnabledKey: "true"},
		{Path: "skills/t468-foo/SKILL.md", EnabledKey: "true"},
	})
	stubCodexHome(t, home)

	check := checkCodexWiring(wireProjectForDoctor(t), false)
	if !strings.Contains(check.Message, "1 stale skill entry") {
		t.Errorf("relative declaration inflated the missing count: %q", check.Message)
	}
	detail := codexDetailText(check)
	if !strings.Contains(detail, "1 with a path that no longer exists") {
		t.Errorf("missing count lost: %q", check.Detail)
	}
	if !strings.Contains(detail, "1 relative entry (not checked: the resolution base is not observed)") {
		t.Errorf("Detail does not carry the distinct relative classification: %q", check.Detail)
	}
}

// TestCodexSkillPath_RelativeOnlyNonDestructiveFinding (AC-CSP-001-03): a
// config whose only unusual entry is relative may surface the classification,
// but the finding must be non-destructive — the remove directive never
// attaches to it.
func TestCodexSkillPath_RelativeOnlyNonDestructiveFinding(t *testing.T) {
	stubCodexLookup(t, true, true)
	home := writeCodexHomeConfig(t, []codexSkillEntrySpec{
		{Path: "skills/t468-only/SKILL.md", EnabledKey: "true"},
	})
	stubCodexHome(t, home)

	check := checkCodexWiring(wireProjectForDoctor(t), false)
	combined := check.Message + " " + codexDetailText(check)
	if strings.Contains(combined, "remove the stale entries") {
		t.Errorf("a relative-only config was advised to remove entries: %+v", check)
	}
	if !strings.Contains(codexDetailText(check), "1 relative entry (not checked: the resolution base is not observed)") {
		t.Errorf("relative classification not surfaced for a relative-only config: %q", check.Detail)
	}
}

// TestCodexSkillPath_BackslashAndOtherUserHomeNotMissing (AC-CSP-001-04): a
// backslash-bearing NON-absolute fragment ("skills\foo\SKILL.md" is not
// filepath.IsAbs on darwin OR windows) and another user's ~user home form
// are both oddly-formed — never missing, surfaced as their own
// classification. On a Windows host a native C:\... absolute stays ABSOLUTE
// (IsAbs decides), never oddly-formed.
func TestCodexSkillPath_BackslashAndOtherUserHomeNotMissing(t *testing.T) {
	stubCodexLookup(t, true, true)
	home := writeCodexHomeConfig(t, []codexSkillEntrySpec{
		{Path: absentSkillPath(t, "anchor"), EnabledKey: "true"},
		{Path: `skills\t468-foo\SKILL.md`, EnabledKey: "true"},
		{Path: "~otheruser/skills/t468/SKILL.md", EnabledKey: "true"},
	})
	stubCodexHome(t, home)

	check := checkCodexWiring(wireProjectForDoctor(t), false)
	if !strings.Contains(check.Message, "1 stale skill entry") {
		t.Errorf("oddly-formed declarations inflated the missing count: %q", check.Message)
	}
	detail := codexDetailText(check)
	if !strings.Contains(detail, "1 with a path that no longer exists") {
		t.Errorf("missing count lost: %q", check.Detail)
	}
	if !strings.Contains(detail, "2 oddly-formed entries (not checked: backslash or ~other-user shape)") {
		t.Errorf("Detail does not carry the distinct oddly-formed classification: %q", check.Detail)
	}
}

// TestCodexSkillPath_AbsoluteExistingAndMissing (AC-CSP-001-05): regression
// guard — the pre-existing absolute behavior is untouched: an absolute entry
// that exists yields no finding; an absolute entry that does not is counted
// missing WITH the remove directive. Both fixture paths are temp-root-derived
// absolutes so the same judgment holds on a Windows host.
func TestCodexSkillPath_AbsoluteExistingAndMissing(t *testing.T) {
	stubCodexLookup(t, true, true)
	home := writeCodexHomeConfig(t, []codexSkillEntrySpec{
		{Path: liveSkillFile(t), EnabledKey: "true"},
		{Path: absentSkillPath(t, "gone"), EnabledKey: "false"},
	})
	stubCodexHome(t, home)

	check := checkCodexWiring(wireProjectForDoctor(t), false)
	if !strings.Contains(check.Message, "1 stale skill entry") {
		t.Errorf("absolute missing entry not counted: %q", check.Message)
	}
	detail := codexDetailText(check)
	if !strings.Contains(detail, "(0 enabled, 1 disabled, 0 unspecified)") {
		t.Errorf("declared split wrong for the absolute pair: %q", check.Detail)
	}
	if !strings.Contains(detail, "remove the stale entries") {
		t.Errorf("absolute missing entry lost the remove directive: %q", check.Detail)
	}
}

// TestCodexSkillPath_RealMissingWithSymlinkLoopIndeterminate
// (AC-CSP-001-06): regression guard preserving the t451 stat-error taxonomy
// verbatim — a real missing entry (temp-root absolute that does not exist) is
// counted, a symlink loop (ELOOP, not ErrNotExist) stays indeterminate and
// surfaced in Detail, never folded into missing.
func TestCodexSkillPath_RealMissingWithSymlinkLoopIndeterminate(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation needs privileges on windows")
	}
	stubCodexLookup(t, true, true)
	dir := t.TempDir()
	loopA := filepath.Join(dir, "loopA")
	loopB := filepath.Join(dir, "loopB")
	if err := os.Symlink(loopB, loopA); err != nil {
		t.Skipf("symlink unsupported here: %v", err)
	}
	if err := os.Symlink(loopA, loopB); err != nil {
		t.Skipf("symlink unsupported here: %v", err)
	}
	if _, err := os.Stat(loopA); err == nil || errors.Is(err, fs.ErrNotExist) {
		t.Skipf("symlink loop did not produce a non-ENOENT stat error: %v", err)
	}

	home := writeCodexHomeConfig(t, []codexSkillEntrySpec{
		{Path: absentSkillPath(t, "real-miss"), EnabledKey: "true"},
		{Path: loopA, EnabledKey: "true"},
	})
	stubCodexHome(t, home)

	check := checkCodexWiring(wireProjectForDoctor(t), false)
	if !strings.Contains(check.Message, "1 stale skill entry") {
		t.Errorf("indeterminate stat inflated the missing count: %q", check.Message)
	}
	if !strings.Contains(codexDetailText(check), "1 could not be checked and are NOT counted as missing") {
		t.Errorf("Detail does not disclose the unchecked loop entry: %q", check.Detail)
	}
}

// TestCodexSkillPath_ExpansionUsesUserHomeSeam (AC-CSP-001-08): the ~
// expansion observably routes through the codexUserHomeDir seam (the USER
// home), not through CODEX_HOME or the codex home. The seam home contains
// the expansion target; the codex home (home/.codex) does not — expanding
// against the codex home would report the entry missing. A bare ~ expands to
// the seam-pinned user home itself and is likewise not missing.
func TestCodexSkillPath_ExpansionUsesUserHomeSeam(t *testing.T) {
	stubCodexLookup(t, true, true)
	home := writeCodexHomeConfig(t, []codexSkillEntrySpec{
		{Path: "~/t468-seam-target/SKILL.md", EnabledKey: "true"},
		{Path: "~", EnabledKey: "true"},
	})
	t468HomeSkillFile(t, home, "t468-seam-target/SKILL.md")
	stubCodexHome(t, home)

	check := checkCodexWiring(wireProjectForDoctor(t), false)
	if check.Status != uikit.CheckOK {
		t.Errorf("~ expansion did not use the pinned user home seam: %+v", check)
	}
	if strings.Contains(check.Message+" "+codexDetailText(check), "stale skill") {
		t.Errorf("seam-pinned ~ entries produced a stale finding: %+v", check)
	}
}

// ---------------------------------------------------------------------------
// SPEC-CODEX-MIRROR-DOCTOR-001 — `.agents/skills` mirror-state diagnostic.
//
// Every test below builds its own mirror fixture under t.TempDir(): `.agents/`
// is gitignored in this repository, so there is no in-repo mirror to read
// (plan.md §C.1).
// ---------------------------------------------------------------------------

// The two user-visible strings the ACs pin, written as LITERALS rather than as
// references to the implementation's own constants. Comparing a message
// against the constant that produced it is tautological; comparing it against
// the text the acceptance criteria name is not.
const (
	testMirrorRedeployDirective = "moai update --templates-only --force --yes"
	testMirrorDetailPhrase      = "does not scan .claude/skills"
)

// mirrorEntryKind is how one fixture mirror entry is materialized.
type mirrorEntryKind int

const (
	// mirrorEntryLive is a relative symlink whose target exists.
	mirrorEntryLive mirrorEntryKind = iota
	// mirrorEntryDangling is a relative symlink whose target does not exist.
	mirrorEntryDangling
	// mirrorEntryRealDir is a real directory occupying the mirror path — the
	// copy fallback's materialization (skill_mirror.go MirrorModeCopy).
	mirrorEntryRealDir
)

// mirrorEntrySpec is one entry to build under `.agents/skills`.
type mirrorEntrySpec struct {
	name string
	kind mirrorEntryKind
}

// writeSkillMirror builds `.claude/skills/<canonical...>` and the `.agents/
// skills` entries per spec under root. It mirrors the producer's own layout:
// the link body is the RELATIVE "../../.claude/skills/<name>" that
// skill_mirror.go's mirrorLinkTarget emits, so a fixture link resolves exactly
// as a real one does.
//
// Skips on a host that cannot create symlinks, using this file's existing
// idiom — a fixture the host cannot build asserts nothing.
func writeSkillMirror(t *testing.T, root string, canonical []string, entries []mirrorEntrySpec) {
	t.Helper()
	needsSymlink := false
	for _, e := range entries {
		if e.kind != mirrorEntryRealDir {
			needsSymlink = true
		}
	}
	if needsSymlink && runtime.GOOS == "windows" {
		t.Skip("symlink creation needs privileges on windows")
	}

	claudeSkills := filepath.Join(root, ".claude", "skills")
	if err := os.MkdirAll(claudeSkills, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range canonical {
		if err := os.MkdirAll(filepath.Join(claudeSkills, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	mirrorDir := filepath.Join(root, ".agents", "skills")
	if err := os.MkdirAll(mirrorDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		dst := filepath.Join(mirrorDir, e.name)
		if e.kind == mirrorEntryRealDir {
			if err := os.MkdirAll(dst, 0o755); err != nil {
				t.Fatal(err)
			}
			continue
		}
		if err := os.Symlink("../../.claude/skills/"+e.name, dst); err != nil {
			t.Skipf("symlink unsupported here: %v", err)
		}
	}
}

// mirrorTreeListing records the shape of both skill trees as a lexically
// ordered path + mode + link-target listing. WalkDir never follows a symlink,
// so the listing records the LINK rather than what it points at — which is
// what a read-only assertion has to compare.
func mirrorTreeListing(t *testing.T, root string) string {
	t.Helper()
	var sb strings.Builder
	for _, rel := range []string{".agents", ".claude"} {
		base := filepath.Join(root, rel)
		werr := filepath.WalkDir(base, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				// The error text IS the recorded state: a tree that could not
				// be walked before must be equally unwalkable after.
				fmt.Fprintf(&sb, "%s WALK-ERR %v\n", p, err)
				return nil
			}
			info, ierr := d.Info()
			if ierr != nil {
				fmt.Fprintf(&sb, "%s INFO-ERR %v\n", p, ierr)
				return nil
			}
			target := ""
			if info.Mode()&os.ModeSymlink != 0 {
				target, _ = os.Readlink(p)
			}
			fmt.Fprintf(&sb, "%s %s %s\n", p, info.Mode().String(), target)
			return nil
		})
		if werr != nil {
			fmt.Fprintf(&sb, "%s ROOT-ERR %v\n", base, werr)
		}
	}
	return sb.String()
}

// TestCheckCodexWiring_ClaudeOnlyMachineNoMirrorRow is AC-CMD-001, the
// regression this change most plausibly causes (REQ-CMD-003).
//
// It SUPPLEMENTS — and does not replace — TestCheckCodexWiring_ClaudeOnly
// MachineStaysSilent, which guards the same un-nagging invariant against a
// different intruder (the home-layer skills.config sub-check). This one guards
// it against the mirror sub-check: a claude-only user must gain no new text.
func TestCheckCodexWiring_ClaudeOnlyMachineNoMirrorRow(t *testing.T) {
	stubCodexLookup(t, true, false) // codex NOT on PATH
	stubCodexHome(t, t.TempDir())

	root := t.TempDir() // no .codex/ wiring, no .agents/skills
	check := checkCodexWiring(root, true)

	if check.Status != uikit.CheckOK {
		t.Errorf("claude-only machine status = %v, want OK: %+v", check.Status, check)
	}
	if !strings.Contains(check.Message, "not wired (claude-only project)") {
		t.Errorf("informational-skip message changed: %q", check.Message)
	}
	both := strings.ToLower(check.Message + " " + codexDetailText(check))
	for _, forbidden := range []string{".agents", "mirror"} {
		if strings.Contains(both, forbidden) {
			t.Errorf("claude-only machine was nagged about the mirror (%q present): %+v", forbidden, check)
		}
	}
}

// TestCheckCodexWiring_MirrorAbsentAdvisesRedeploy is AC-CMD-003
// (REQ-CMD-004): on a WIRED project an absent mirror means Codex sees no MoAI
// skills at all, and a routine `moai update` was measured not to restore it
// (root-cause.md Claim 2) — so the summary carries the FORCED re-deploy.
func TestCheckCodexWiring_MirrorAbsentAdvisesRedeploy(t *testing.T) {
	stubCodexLookup(t, true, true)
	stubCodexHome(t, t.TempDir())

	root := wireProjectWithoutMirror(t)
	check := checkCodexWiring(root, false)

	if check.Status != uikit.CheckWarn {
		t.Fatalf("absent mirror on a wired project status = %v, want Warn: %+v", check.Status, check)
	}
	for _, want := range []string{".agents/skills", testMirrorRedeployDirective} {
		if !strings.Contains(check.Message, want) {
			t.Errorf("Message missing %q (Detail is --verbose-only): %q", want, check.Message)
		}
	}
}

// TestCheckCodexWiring_DanglingMirrorEntriesCounted is AC-CMD-004
// (REQ-CMD-005): an entry claiming a skill that is not there. Nothing repairs
// it between deploys, so it is a finding and the COUNT rides in the summary.
func TestCheckCodexWiring_DanglingMirrorEntriesCounted(t *testing.T) {
	stubCodexLookup(t, true, true)
	stubCodexHome(t, t.TempDir())

	root := wireProjectForDoctor(t)
	writeSkillMirror(t, root, []string{"moai-live"}, []mirrorEntrySpec{
		{name: "moai-live", kind: mirrorEntryLive},
		{name: "moai-gone-a", kind: mirrorEntryDangling},
		{name: "moai-gone-b", kind: mirrorEntryDangling},
	})

	check := checkCodexWiring(root, false)
	if check.Status != uikit.CheckWarn {
		t.Fatalf("dangling entries status = %v, want Warn: %+v", check.Status, check)
	}
	if !strings.Contains(check.Message, "2") {
		t.Errorf("Message does not name the dangling count 2: %q", check.Message)
	}
	if !strings.Contains(check.Message, testMirrorRedeployDirective) {
		t.Errorf("Message missing the re-deploy directive: %q", check.Message)
	}
}

// TestCheckCodexWiring_CopyModeDetailOnly is AC-CMD-005 (REQ-CMD-006). A real
// directory in a mirror path is the copy fallback: FUNCTIONAL, and the
// expected materialization wherever symlink creation is unavailable. Warning
// on it would hand every such user a permanent row for a working mirror — the
// un-nagging invariant failing by a different door.
func TestCheckCodexWiring_CopyModeDetailOnly(t *testing.T) {
	stubCodexLookup(t, true, true)
	stubCodexHome(t, t.TempDir())

	root := wireProjectForDoctor(t)
	writeSkillMirror(t, root, []string{"moai-linked", "moai-copied"}, []mirrorEntrySpec{
		{name: "moai-linked", kind: mirrorEntryLive},
		{name: "moai-copied", kind: mirrorEntryRealDir},
	})

	check := checkCodexWiring(root, true)
	if check.Status != uikit.CheckOK {
		t.Fatalf("copy-mode entry was escalated to a finding: %+v", check)
	}
	detail := codexDetailText(check)
	if !strings.Contains(detail, "1") || !strings.Contains(detail, "copy") {
		t.Errorf("Detail does not report the copy-mode count: %q", detail)
	}
	if strings.Contains(strings.ToLower(check.Message), "copy") {
		t.Errorf("copy-mode text leaked into Message: %q", check.Message)
	}
}

// TestCheckCodexWiring_UnmirroredSkillsDetailOnly is AC-CMD-006 (REQ-CMD-007).
// The correct denominator — "skills THIS deploy mirrored" — is not observable
// from doctor, and a project may legitimately carry locally-authored skills no
// deploy ever mirrored. Counted, reported, never warned.
func TestCheckCodexWiring_UnmirroredSkillsDetailOnly(t *testing.T) {
	stubCodexLookup(t, true, true)
	stubCodexHome(t, t.TempDir())

	root := wireProjectForDoctor(t)
	writeSkillMirror(t, root,
		[]string{"moai-mirrored", "local-one", "local-two"},
		[]mirrorEntrySpec{{name: "moai-mirrored", kind: mirrorEntryLive}})

	check := checkCodexWiring(root, true)
	if check.Status != uikit.CheckOK {
		t.Fatalf("unmirrored skills were escalated to a finding: %+v", check)
	}
	if !strings.Contains(codexDetailText(check), "2") {
		t.Errorf("Detail does not report the unmirrored count 2: %q", check.Detail)
	}
	if strings.Contains(check.Message, ".agents/skills") {
		t.Errorf("unmirrored text leaked into Message: %q", check.Message)
	}
}

// TestCheckCodexWiring_UnwiredNoMirrorNag is AC-CMD-007 (REQ-CMD-008, plus the
// unwired clause of REQ-CMD-006/007). Codex is installed but the project never
// opted in: the existing init directive already points at a deploy that
// creates the mirror, so a second directive here would double-nag.
//
// Sub-case (b) pins the unwired clause: a mirror in a REPORTABLE state still
// produces nothing, because the inspector is never called outside the wired
// branch.
func TestCheckCodexWiring_UnwiredNoMirrorNag(t *testing.T) {
	cases := []struct {
		name        string
		buildMirror bool
	}{
		{name: "no_mirror_directory", buildMirror: false},
		{name: "reportable_mirror_still_silent", buildMirror: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stubCodexLookup(t, true, true) // codex FOUND, project unwired
			stubCodexHome(t, t.TempDir())

			root := t.TempDir()
			if tc.buildMirror {
				writeSkillMirror(t, root, []string{"moai-live"}, []mirrorEntrySpec{
					{name: "moai-live", kind: mirrorEntryLive},
					{name: "moai-copied", kind: mirrorEntryRealDir},
					{name: "moai-gone", kind: mirrorEntryDangling},
				})
			}

			check := checkCodexWiring(root, true)
			if !strings.Contains(check.Message, initCodexAdvice) {
				t.Errorf("existing unwired directive missing: %q", check.Message)
			}
			both := check.Message + " " + codexDetailText(check)
			if strings.Contains(both, ".agents/skills") {
				t.Errorf("unwired project was nagged about the mirror: %+v", check)
			}
			if strings.Contains(both, testMirrorRedeployDirective) {
				t.Errorf("unwired project got the re-deploy directive: %+v", check)
			}
		})
	}
}

// TestCheckCodexWiring_MirrorUnreadableIndeterminate is AC-CMD-008
// (REQ-CMD-010): an unobserved absence is never reported as absent. The
// fixture is this file's existing symlink-loop idiom rather than a chmod 0o000
// directory, which was measured to break t.TempDir() cleanup.
func TestCheckCodexWiring_MirrorUnreadableIndeterminate(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation needs privileges on windows")
	}
	stubCodexLookup(t, true, true)
	stubCodexHome(t, t.TempDir())

	// wireProjectWithoutMirror, so `.agents/skills` is free for the loop link
	// to occupy — an existing directory would make os.Symlink fail EEXIST and
	// skip the test, leaving this AC unmet.
	root := wireProjectWithoutMirror(t)
	if err := os.MkdirAll(filepath.Join(root, ".agents"), 0o755); err != nil {
		t.Fatal(err)
	}
	mirrorDir := filepath.Join(root, ".agents", "skills")
	sibling := filepath.Join(root, ".agents", "loop-sibling")
	if err := os.Symlink(sibling, mirrorDir); err != nil {
		t.Skipf("symlink unsupported here: %v", err)
	}
	if err := os.Symlink(mirrorDir, sibling); err != nil {
		t.Skipf("symlink unsupported here: %v", err)
	}
	// Control assertion: without it this AC could pass on a host where the
	// loop resolved, asserting nothing.
	if _, err := os.ReadDir(mirrorDir); err == nil || errors.Is(err, fs.ErrNotExist) {
		t.Skipf("symlink loop did not produce a non-ENOENT read error: %v", err)
	}

	check := checkCodexWiring(root, true)
	if strings.Contains(check.Message, testMirrorRedeployDirective) {
		t.Errorf("an unreadable mirror was reported as absent: %+v", check)
	}
	if !strings.Contains(codexDetailText(check), "not checked") {
		t.Errorf("Detail does not record the indeterminate condition: %q", check.Detail)
	}
}

// TestCheckCodexWiring_MirrorSummaryWidth is AC-CMD-009 clause (a): each
// mirror summary stays inside the width band STANDALONE, so it never widens
// the panel on its own. Each fixture raises exactly one mirror finding and no
// other, which is what makes the measurement standalone by construction.
func TestCheckCodexWiring_MirrorSummaryWidth(t *testing.T) {
	cases := []struct {
		name  string
		build func(t *testing.T, root string)
	}{
		{name: "absent", build: func(*testing.T, string) {}},
		{name: "dangling", build: func(t *testing.T, root string) {
			writeSkillMirror(t, root, []string{"moai-live"}, []mirrorEntrySpec{
				{name: "moai-live", kind: mirrorEntryLive},
				{name: "moai-gone", kind: mirrorEntryDangling},
			})
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stubCodexLookup(t, true, true)
			stubCodexHome(t, t.TempDir())
			root := wireProjectWithoutMirror(t)
			tc.build(t, root)

			check := checkCodexWiring(root, false)
			if check.Status != uikit.CheckWarn {
				t.Fatalf("premise broken — expected exactly one mirror finding: %+v", check)
			}
			if strings.Contains(check.Message, "see --verbose") {
				t.Fatalf("premise broken — the summary was not measured standalone: %q", check.Message)
			}
			if n := utf8.RuneCountInString(check.Message); n > codexMessageWidthCeiling {
				t.Errorf("mirror summary is %d runes, over the %d ceiling: %q", n, codexMessageWidthCeiling, check.Message)
			}
		})
	}
}

// TestCheckCodexWiring_MirrorCheckIsReadOnly is AC-CMD-010 (REQ-CMD-002) and
// the ONLY mechanical guard on the read-only boundary: repair on read would
// make doctor a writer of a surface the deploy path owns.
func TestCheckCodexWiring_MirrorCheckIsReadOnly(t *testing.T) {
	stubCodexLookup(t, true, true)
	stubCodexHome(t, t.TempDir())

	root := wireProjectForDoctor(t)
	writeSkillMirror(t, root,
		[]string{"moai-live", "moai-copied", "local-unmirrored"},
		[]mirrorEntrySpec{
			{name: "moai-live", kind: mirrorEntryLive},
			{name: "moai-gone", kind: mirrorEntryDangling},
			{name: "moai-copied", kind: mirrorEntryRealDir},
		})

	before := mirrorTreeListing(t, root)
	_ = checkCodexWiring(root, true)
	after := mirrorTreeListing(t, root)

	if before != after {
		t.Errorf("the check mutated the skill trees\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

// TestCheckCodexWiring_MirrorFindingParticipatesInTailDrop is AC-CMD-013
// (REQ-CMD-009 clause b): the mirror summary is appended to `problems` and so
// participates in the existing tail-drop UNCHANGED — the bound is a shared
// resource, and a new finding must not be exempted from it.
func TestCheckCodexWiring_MirrorFindingParticipatesInTailDrop(t *testing.T) {
	t.Run("tail_drop_keeps_the_lead_and_detail_keeps_everything", func(t *testing.T) {
		stubCodexLookup(t, true, true)
		stubCodexHome(t, writeCodexHomeConfig(t, []codexSkillEntrySpec{
			{Path: absentSkillPath(t, "moai-a"), EnabledKey: "true"},
		}))

		root := wireProjectWithoutMirror(t) // absent mirror => the mirror finding
		check := checkCodexWiring(root, false)
		if check.Status != uikit.CheckWarn {
			t.Fatalf("premise broken — expected two findings: %+v", check)
		}
		// Premise control: the two summaries must actually overflow, or this
		// test asserts nothing about tail-drop.
		if !strings.Contains(check.Message, "see --verbose") {
			t.Fatalf("premise broken — the joined summaries did not exceed %d runes, so no tail-drop occurred: %q",
				codexMessageWidthCeiling, check.Message)
		}

		if n := utf8.RuneCountInString(check.Message); n > codexMessageWidthCeiling {
			t.Errorf("truncated Message is still %d runes, over the %d ceiling: %q", n, codexMessageWidthCeiling, check.Message)
		}
		if !strings.HasSuffix(check.Message, codexOverflowMarker(1)) {
			t.Errorf("Message does not end with the overflow marker naming 1 dropped summary: %q", check.Message)
		}
		// The dropped finding survives in Detail — including its directive.
		detail := codexDetailText(check)
		for _, want := range []string{testMirrorRedeployDirective, "stale"} {
			if !strings.Contains(detail, want) {
				t.Errorf("Detail dropped %q; it must carry the full text of every finding: %q", want, detail)
			}
		}
	})

	t.Run("lead_summary_exception_stays_unreachable_for_mirror_summaries", func(t *testing.T) {
		// joinCodexSummaries emits a single over-ceiling lead summary whole
		// rather than truncating its own directive. AC-CMD-009 asserts every
		// mirror summary is under the ceiling standalone, so mirror summaries
		// never enter that branch — this sub-case makes the unreachability
		// OBSERVED rather than assumed.
		stubCodexLookup(t, true, true)
		stubCodexHome(t, t.TempDir()) // no home config => no second finding

		check := checkCodexWiring(wireProjectWithoutMirror(t), false)
		if check.Status != uikit.CheckWarn {
			t.Fatalf("premise broken — expected the lone mirror finding: %+v", check)
		}
		if strings.Contains(check.Message, "see --verbose") {
			t.Errorf("a lone mirror finding produced an overflow marker: %q", check.Message)
		}
		if n := utf8.RuneCountInString(check.Message); n > codexMessageWidthCeiling {
			t.Errorf("the lead-summary exception was entered: %d runes > %d: %q", n, codexMessageWidthCeiling, check.Message)
		}
	})
}

// TestCheckCodexWiring_MirrorUsesExistingRowTwoRegisters is AC-CMD-014
// (REQ-CMD-001): the observation rides the EXISTING row in the file's existing
// two-register shape. Clause 3 is asserted on the RENDERED output, never on
// check.Detail: the verbose gate lives at the render layer
// (doctor_render.go:134-137), while the Detail field is assigned regardless.
func TestCheckCodexWiring_MirrorUsesExistingRowTwoRegisters(t *testing.T) {
	stubCodexLookup(t, true, true)
	stubCodexHome(t, t.TempDir())

	check := checkCodexWiring(wireProjectWithoutMirror(t), true)
	if check.Name != "Codex Wiring" {
		t.Errorf("the mirror observation renamed the row: %q", check.Name)
	}
	if !strings.Contains(check.Message, ".agents/skills") {
		t.Errorf("summary register empty of the mirror observation: %q", check.Message)
	}
	if !strings.Contains(codexDetailText(check), testMirrorDetailPhrase) {
		t.Errorf("detail register carries no fuller mirror text: %q", check.Detail)
	}

	var buf bytes.Buffer
	rendered := renderDoctorGroups(&buf, []checkGroup{{
		title:  "Codex",
		checks: []DiagnosticCheck{check},
	}}, false, resolveTheme())
	if strings.Contains(rendered, testMirrorDetailPhrase) {
		t.Errorf("Detail text rendered without --verbose: %q", rendered)
	}
}
