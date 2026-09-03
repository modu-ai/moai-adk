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

// wireProjectForDoctor wires a fresh temp project and returns its root.
func wireProjectForDoctor(t *testing.T) string {
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
		{Path: "/nonexistent/moai-a/SKILL.md", EnabledKey: "true"},
		{Path: "/nonexistent/moai-b/SKILL.md", EnabledKey: "false"},
		{Path: "/nonexistent/moai-c/SKILL.md", EnabledKey: "false"},
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
		{Path: "/nonexistent/only-one/SKILL.md", EnabledKey: "true"},
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
		{Path: "/nonexistent/a/SKILL.md"},                       // no enabled key
		{Path: "/nonexistent/b/SKILL.md", EnabledKey: `"true"`}, // quoted string, still true
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
		{Path: "/nonexistent/moai-a/SKILL.md", EnabledKey: "true"},
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
	cfg := "[[skills.config]]\npath = \"/nonexistent/env/SKILL.md\"\nenabled = true\n"
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
		{Path: "/nonexistent/real-miss/SKILL.md", EnabledKey: "true"},
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
	entries := make([]codexSkillEntrySpec, 0, 49)
	for i := 0; i < 49; i++ { // the real-world census that produced the 1272-column panel
		entries = append(entries, codexSkillEntrySpec{
			Path:       fmt.Sprintf("/nonexistent/moai-skill-%02d/SKILL.md", i),
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
	entries := make([]codexSkillEntrySpec, 0, 49)
	for i := 0; i < 49; i++ {
		entries = append(entries, codexSkillEntrySpec{
			Path:       fmt.Sprintf("/nonexistent/moai-skill-%02d/SKILL.md", i),
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
	home := writeCodexHomeConfig(t, []codexSkillEntrySpec{{Path: "/nonexistent/moai-a/SKILL.md", EnabledKey: "true"}})
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
