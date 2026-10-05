package cli

// codex_skills_prune_test.go — SPEC-CODEX-GHOST-SKILLS-PRUNE-001 (t506).
//
// Every fixture lives under t.TempDir(). No test reads or writes the real
// ~/.codex/config.toml: the config location rides the CODEX_HOME env var and
// "~" expansion rides the codexUserHomeDir seam, so both are pointed at the
// temp tree.
//
// Tests here reassign package-level seams (codexUserHomeDir, osStatFn) and
// therefore MUST NOT call t.Parallel().

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/printer"
	"github.com/modu-ai/moai-adk/internal/codexwiring"
)

// errIndeterminate stands in for a permission denial, a symlink loop, or an
// I/O error: any stat failure that is NOT fs.ErrNotExist. Injected through the
// osStatFn seam rather than through POSIX permission bits, which Windows does
// not model and root bypasses (plan §F.1 — a platform-conditional skip would
// make AC-CGP-004 unachievable, which is a FAIL, not a recorded gap).
var errIndeterminate = errors.New("indeterminate stat failure")

// stubHome points the "~" expansion seam at dir for the duration of the test.
func stubHome(t *testing.T, dir string) {
	t.Helper()
	orig := codexUserHomeDir
	t.Cleanup(func() { codexUserHomeDir = orig })
	codexUserHomeDir = func() (string, error) { return dir, nil }
}

// stubUnresolvableHome makes "~" expansion fail, which must read as
// indeterminate — never as absent (REQ-CGP-008).
func stubUnresolvableHome(t *testing.T) {
	t.Helper()
	orig := codexUserHomeDir
	t.Cleanup(func() { codexUserHomeDir = orig })
	codexUserHomeDir = func() (string, error) { return "", errors.New("no home") }
}

// stubStatIndeterminate makes stat fail with a non-ErrNotExist error for
// exactly one path, leaving every other path on the real filesystem.
func stubStatIndeterminate(t *testing.T, target string) {
	t.Helper()
	orig := osStatFn
	t.Cleanup(func() { osStatFn = orig })
	osStatFn = func(name string) (os.FileInfo, error) {
		if name == target {
			return nil, errIndeterminate
		}
		return orig(name)
	}
}

func TestPruneCodexSkillEntriesRemovesMissingAbsolute(t *testing.T) {
	tmp := t.TempDir()
	gone := filepath.Join(tmp, "gone", "SKILL.md")

	content := []byte("[[skills.config]]\n" +
		"path = \"" + gone + "\"\n" +
		"enabled = false\n" +
		"\n" +
		"[tail]\n" +
		"k = 1\n")

	out, verdicts := pruneCodexSkillEntries(content)

	if len(verdicts) != 1 || !verdicts[0].Eligible {
		t.Fatalf("verdicts = %+v, want one eligible", verdicts)
	}
	got := string(out)
	if strings.Contains(got, gone) || strings.Contains(got, "[[skills.config]]") {
		t.Errorf("entry survived the prune:\n%s", got)
	}
	if !strings.Contains(got, "[tail]\nk = 1\n") {
		t.Errorf("tail table damaged:\n%s", got)
	}
}

// AC-CGP-002 — a home-relative declaration is expanded through
// expandCodexHomeRelativePath (the codexUserHomeDir seed), never through
// CODEX_HOME.
func TestPruneCodexSkillEntriesRemovesMissingHomeRelative(t *testing.T) {
	tmp := t.TempDir()
	stubHome(t, tmp)

	content := []byte("[[skills.config]]\n" +
		"path = \"~/gone/SKILL.md\"\n")

	out, verdicts := pruneCodexSkillEntries(content)
	if len(verdicts) != 1 || !verdicts[0].Eligible {
		t.Fatalf("verdicts = %+v, want one eligible", verdicts)
	}
	if strings.Contains(string(out), "skills.config") {
		t.Errorf("entry survived:\n%s", out)
	}
}

// AC-CGP-005 — `enabled` is not an eligibility gate. An enabled entry whose
// path is absent is a ghost too, and it is removed because it satisfies
// REQ-CGP-004, not because this AC directs the removal.
func TestPruneCodexSkillEntriesEnabledIsNotAGate(t *testing.T) {
	tmp := t.TempDir()
	gone := filepath.Join(tmp, "gone")

	content := []byte("[[skills.config]]\npath = \"" + gone + "\"\nenabled = true\n")
	out, verdicts := pruneCodexSkillEntries(content)

	if len(verdicts) != 1 || !verdicts[0].Eligible {
		t.Fatalf("verdicts = %+v, want one eligible", verdicts)
	}
	if strings.Contains(string(out), "skills.config") {
		t.Errorf("entry survived:\n%s", out)
	}
}

// neverPruneFixture builds the AC-CGP-003 fixture: eight rows covering seven
// never-prune classes. Every declared path that COULD resolve is deliberately
// absent, so only its own class saves it — a row that survives for the wrong
// reason would be indistinguishable from one that survives for the right one.
func neverPruneFixture(t *testing.T, tmp string) (content []byte, wantPresent []string) {
	t.Helper()

	existsFile := filepath.Join(tmp, "present", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(existsFile), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(existsFile, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	existsDir := filepath.Join(tmp, "present-dir")
	if err := os.MkdirAll(existsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	indeterminate := filepath.Join(tmp, "indeterminate", "SKILL.md")
	stubStatIndeterminate(t, indeterminate)

	missingA := filepath.Join(tmp, "gone-a")
	missingB := filepath.Join(tmp, "gone-b")
	missingC := filepath.Join(tmp, "gone-c")

	rows := []string{
		// 1 — relative
		"[[skills.config]]\npath = \"skills/x/SKILL.md\"\n",
		// 2 — oddly formed: another user's home, and a backslash fragment
		"[[skills.config]]\npath = \"~other/x/SKILL.md\"\n",
		"[[skills.config]]\npath = \"sub\\\\dir/SKILL.md\"\n",
		// 3 — home-relative while the home is unresolvable
		"[[skills.config]]\npath = \"~/unresolvable/SKILL.md\"\n",
		// 4 — indeterminate stat error
		"[[skills.config]]\npath = \"" + indeterminate + "\"\n",
		// 5a — no path key at all
		"[[skills.config]]\nenabled = true\n",
		// 5b — an empty path key: skillPathKeyRe matches, Path is ""
		"[[skills.config]]\npath = \"\"\n",
		// 6 — a resolving file, and a resolving DIRECTORY
		"[[skills.config]]\npath = \"" + existsFile + "\"\n",
		"[[skills.config]]\npath = \"" + existsDir + "\"\n",
		// 7a — a multi-line literal inside the entry
		"[[skills.config]]\npath = \"" + missingA + "\"\nnotes = \"\"\"\nprose\n\"\"\"\n",
		// 7b — a key that is neither path nor enabled
		"[[skills.config]]\npath = \"" + missingB + "\"\nnotes = \"hi\"\n",
		// 7c — an array continuation line anyTableRe reads as a table header
		"[[skills.config]]\npath = \"" + missingC + "\"\nmatrix = [\n[\"x\", \"y\"]\n]\n",
	}

	return []byte(strings.Join(rows, "\n") + "\n[tail]\nk = 1\n"), []string{
		"skills/x/SKILL.md",
		"~other/x/SKILL.md",
		// Verbatim as the TOML declares it: the parser does not decode escape
		// sequences, so both backslashes survive into Path.
		`sub\\dir/SKILL.md`,
		"~/unresolvable/SKILL.md",
		indeterminate,
		existsFile,
		existsDir,
		missingA,
		missingB,
		missingC,
	}
}

// AC-CGP-003 — every never-prune row survives a --force run.
func TestPruneCodexSkillEntriesNeverPruneClasses(t *testing.T) {
	tmp := t.TempDir()
	stubUnresolvableHome(t)
	content, wantPresent := neverPruneFixture(t, tmp)

	// Precondition: the fixture must actually declare every row as an entry.
	// A fixture that produced fewer entries would "preserve" the rest by never
	// having seen them — an empty operand asserting nothing.
	entries := codexwiring.ParseSkillEntries(content)
	if len(entries) != 12 {
		t.Fatalf("fixture declared %d entries, want 12 — the fixture did not fire", len(entries))
	}

	out, verdicts := pruneCodexSkillEntries(content)

	for _, v := range verdicts {
		if v.Eligible {
			t.Errorf("entry %q judged eligible; no never-prune row may be removed", v.Entry.Path)
		}
		if v.SkipReason == "" {
			t.Errorf("entry %q skipped with no reason recorded", v.Entry.Path)
		}
	}
	for _, want := range wantPresent {
		if !strings.Contains(string(out), want) {
			t.Errorf("row %q was removed", want)
		}
	}
	if !bytes.Equal(out, content) {
		t.Errorf("output differs from input although nothing was eligible")
	}
}

// AC-CGP-003 row 7-d — the swallow counter-example. A comment carrying an odd
// number of `"""` opens a literal that swallows a whole healthy registration;
// deleting the ghost's extent would destroy it. This is the row that
// separates a parser-state judgment from a text re-scan.
func TestPruneCodexSkillEntriesSwallowedRegistrationSurvives(t *testing.T) {
	tmp := t.TempDir()
	exists := filepath.Join(tmp, "exists", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(exists), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(exists, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	gone := filepath.Join(tmp, "gone")

	content := []byte("[[skills.config]]\n" + // L0
		"path = \"" + gone + "\"\n" + // L1
		"# uses \"\"\" in prose\n" + // L2 — opens the literal
		"[[skills.config]]\n" + // L3 — swallowed
		"path = \"" + exists + "\"\n" + // L4 — swallowed
		"# and \"\"\" again\n" + // L5 — closes
		"[other]\n") // L6

	// Precondition — the PASS state of this row is "nothing happens", so a
	// fixture that never opened a literal would go green identically. Assert
	// the fixture fired BEFORE asserting anything about eligibility.
	entries := codexwiring.ParseSkillEntries(content)
	if len(entries) != 1 || entries[0].Path != gone {
		t.Fatalf("the fixture did not open a literal: entries=%d %+v", len(entries), entries)
	}

	out, verdicts := pruneCodexSkillEntries(content)

	if len(verdicts) != 1 || verdicts[0].Eligible {
		t.Fatalf("the swallowing ghost was judged eligible: %+v", verdicts)
	}
	for _, line := range []string{"[[skills.config]]\npath = \"" + exists + "\"\n# and \"\"\" again\n"} {
		if !strings.Contains(string(out), line) {
			t.Errorf("the swallowed registration was destroyed:\n%s", out)
		}
	}
	if !bytes.Equal(out, content) {
		t.Errorf("bytes changed although the only entry was ineligible")
	}
}

// AC-CGP-003 row 7-d' — the narrow variant: the swallowed span carries no
// header, so a "one header per extent" guard would not fire.
func TestPruneCodexSkillEntriesSwallowedRegistrationNarrowSurvives(t *testing.T) {
	tmp := t.TempDir()
	exists := filepath.Join(tmp, "exists", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(exists), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(exists, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	gone := filepath.Join(tmp, "gone")

	content := []byte("[[skills.config]]\n" +
		"path = \"" + gone + "\"\n" +
		"# uses \"\"\" in prose\n" +
		"path = \"" + exists + "\"\n" +
		"# and \"\"\" again\n" +
		"[other]\n")

	entries := codexwiring.ParseSkillEntries(content)
	if len(entries) != 1 || entries[0].Path != gone {
		t.Fatalf("the variant fixture did not open a literal: entries=%d %+v", len(entries), entries)
	}

	out, verdicts := pruneCodexSkillEntries(content)
	if len(verdicts) != 1 || verdicts[0].Eligible {
		t.Fatalf("the swallowing ghost was judged eligible: %+v", verdicts)
	}
	if !bytes.Equal(out, content) {
		t.Errorf("bytes changed although the only entry was ineligible:\n%s", out)
	}
}

// AC-CGP-007 — a [[skills.config]] header written INSIDE a multi-line literal
// that sits outside any entry is a string, not a registration.
func TestPruneCodexSkillEntriesIgnoresHeaderInsideDocString(t *testing.T) {
	tmp := t.TempDir()
	gone := filepath.Join(tmp, "gone")

	doc := "[docs]\n" +
		"text = \"\"\"\n" +
		"[[skills.config]]\n" +
		"path = \"/not/a/registration\"\n" +
		"\"\"\"\n"
	content := []byte(doc + "\n[[skills.config]]\npath = \"" + gone + "\"\n")

	out, verdicts := pruneCodexSkillEntries(content)
	if len(verdicts) != 1 || !verdicts[0].Eligible {
		t.Fatalf("verdicts = %+v, want exactly one eligible entry", verdicts)
	}
	if !strings.Contains(string(out), doc) {
		t.Errorf("the documentation literal was modified:\n%s", out)
	}
	if strings.Contains(string(out), gone) {
		t.Errorf("the eligible entry survived:\n%s", out)
	}
}

// AC-CGP-014 — the bytes that were not removed are byte-identical, across all
// three line-ending shapes.
func TestPruneCodexSkillEntriesPreservesUntouchedBytes(t *testing.T) {
	tmp := t.TempDir()
	gone := filepath.Join(tmp, "gone")

	base := "# leading comment\n" +
		"[keep]\n" +
		"note = \"\"\"\nmulti\nline\n\"\"\"\n" +
		"\n" +
		"[[skills.config]]\n" +
		"path = \"" + gone + "\"\n" +
		"\n" +
		"[tail]\n" +
		"k = 1\n"
	want := "# leading comment\n" +
		"[keep]\n" +
		"note = \"\"\"\nmulti\nline\n\"\"\"\n" +
		"\n" +
		"\n" +
		"[tail]\n" +
		"k = 1\n"

	cases := []struct {
		name    string
		in, out string
	}{
		{"trailing newline", base, want},
		{"no trailing newline", strings.TrimSuffix(base, "\n"), strings.TrimSuffix(want, "\n")},
		{"CRLF", strings.ReplaceAll(base, "\n", "\r\n"), strings.ReplaceAll(want, "\n", "\r\n")},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, verdicts := pruneCodexSkillEntries([]byte(tc.in))
			if len(verdicts) != 1 || !verdicts[0].Eligible {
				t.Fatalf("verdicts = %+v, want one eligible", verdicts)
			}
			if string(got) != tc.out {
				t.Errorf("output = %q\nwant     %q", got, tc.out)
			}
		})
	}
}

// AC-CGP-016 — an eligible entry that is the file's LAST block, so the parser
// leaves the loop without a closing event.
func TestPruneCodexSkillEntriesEntryAtEOF(t *testing.T) {
	tmp := t.TempDir()
	gone := filepath.Join(tmp, "gone")

	base := "[keep]\nk = 1\n[[skills.config]]\npath = \"" + gone + "\"\nenabled = false\n"
	want := "[keep]\nk = 1\n"

	cases := []struct {
		name    string
		in, out string
	}{
		{"trailing newline", base, want},
		{"no trailing newline", strings.TrimSuffix(base, "\n"), strings.TrimSuffix(want, "\n")},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, verdicts := pruneCodexSkillEntries([]byte(tc.in))
			if len(verdicts) != 1 || !verdicts[0].Eligible {
				t.Fatalf("verdicts = %+v, want one eligible", verdicts)
			}
			if string(got) != tc.out {
				t.Errorf("output = %q\nwant     %q", got, tc.out)
			}
		})
	}
}

// writeCodexConfig seeds a CODEX_HOME with a config.toml and returns its path.
func writeCodexConfig(t *testing.T, body string) (home, cfgPath string) {
	t.Helper()
	home = t.TempDir()
	cfgPath = filepath.Join(home, "config.toml")
	if err := os.WriteFile(cfgPath, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv(codexHomeEnvVar, home)
	return home, cfgPath
}

func runPrune(t *testing.T, force bool) (out, errOut string, err error) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	p := printer.New(printer.WithWriters(&stdout, &stderr))
	err = runCleanCodexSkills(p, force)
	return stdout.String(), stderr.String(), err
}

// AC-CGP-006 — dry-run is the default and writes nothing.
func TestRunCleanCodexSkillsDryRunWritesNothing(t *testing.T) {
	tmp := t.TempDir()
	gone := filepath.Join(tmp, "gone")
	body := "[[skills.config]]\npath = \"" + gone + "\"\n"
	_, cfgPath := writeCodexConfig(t, body)

	before, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}

	stdout, stderr, err := runPrune(t, false)
	if err != nil {
		t.Fatalf("runCleanCodexSkills: %v", err)
	}

	after, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Errorf("dry-run modified the config")
	}
	if !strings.Contains(stdout+stderr, gone) {
		t.Errorf("dry-run did not enumerate the eligible entry:\n%s%s", stdout, stderr)
	}
}

// AC-CGP-008 — the backup lands BEFORE the write, and its path and sha256 are
// reported in a form that copies verbatim into a verdict record.
func TestRunCleanCodexSkillsBacksUpBeforeWriting(t *testing.T) {
	tmp := t.TempDir()
	gone := filepath.Join(tmp, "gone")
	body := "[[skills.config]]\npath = \"" + gone + "\"\n[tail]\nk = 1\n"
	home, cfgPath := writeCodexConfig(t, body)

	stdout, stderr, err := runPrune(t, true)
	if err != nil {
		t.Fatalf("runCleanCodexSkills: %v", err)
	}
	report := stdout + stderr

	entries, err := os.ReadDir(home)
	if err != nil {
		t.Fatal(err)
	}
	var backup string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "config.toml.bak-") {
			backup = filepath.Join(home, e.Name())
		}
	}
	if backup == "" {
		t.Fatalf("no backup was written into %s", home)
	}
	got, err := os.ReadFile(backup)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != body {
		t.Errorf("backup is not the pre-write config:\n%s", got)
	}
	if !strings.Contains(report, backup) {
		t.Errorf("the backup path is absent from the report:\n%s", report)
	}
	if !strings.Contains(report, sha256Hex([]byte(body))) {
		t.Errorf("the backup sha256 is absent from the report:\n%s", report)
	}

	after, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(after), gone) {
		t.Errorf("the eligible entry survived the --force run:\n%s", after)
	}
}

// AC-CGP-003 (reporting half) — a kept entry is enumerated with its reason.
// Preserving a ghost silently is not enough: the user asked for the ghosts to
// go and would otherwise believe every one of them did.
func TestRunCleanCodexSkillsReportsSkippedEntries(t *testing.T) {
	tmp := t.TempDir()
	missing := filepath.Join(tmp, "gone")
	body := "[[skills.config]]\n" +
		"path = \"" + missing + "\"\n" +
		"notes = \"hi\"\n" + // unrecognised: disqualifies the entry
		"\n" +
		"[[skills.config]]\n" +
		"enabled = true\n" // declares no path
	writeCodexConfig(t, body)

	stdout, stderr, err := runPrune(t, true)
	if err != nil {
		t.Fatalf("runCleanCodexSkills: %v", err)
	}
	report := stdout + stderr

	if !strings.Contains(report, missing) || !strings.Contains(report, "not recognised") {
		t.Errorf("the unrecognised-line skip was not enumerated:\n%s", report)
	}
	// An entry with no path has no path to name, so it is named by its line.
	if !strings.Contains(report, "entry at line 5") || !strings.Contains(report, "declares no path") {
		t.Errorf("the no-path skip was not enumerated:\n%s", report)
	}
}

// AC-CGP-009 — fail-open on every missing input.
func TestRunCleanCodexSkillsFailsOpen(t *testing.T) {
	t.Run("unresolvable home", func(t *testing.T) {
		t.Setenv(codexHomeEnvVar, "")
		stubUnresolvableHome(t)
		if _, _, err := runPrune(t, true); err != nil {
			t.Errorf("err = %v, want nil", err)
		}
	})

	t.Run("config absent", func(t *testing.T) {
		t.Setenv(codexHomeEnvVar, t.TempDir())
		if _, _, err := runPrune(t, true); err != nil {
			t.Errorf("err = %v, want nil", err)
		}
	})

	t.Run("config unreadable", func(t *testing.T) {
		home := t.TempDir()
		// A directory in the config's place: os.ReadFile fails with a
		// non-ENOENT error on every platform.
		if err := os.MkdirAll(filepath.Join(home, "config.toml"), 0o755); err != nil {
			t.Fatal(err)
		}
		t.Setenv(codexHomeEnvVar, home)
		if _, _, err := runPrune(t, true); err != nil {
			t.Errorf("err = %v, want nil", err)
		}
	})

	t.Run("zero entries", func(t *testing.T) {
		body := "[tail]\nk = 1\n"
		_, cfgPath := writeCodexConfig(t, body)
		if _, _, err := runPrune(t, true); err != nil {
			t.Errorf("err = %v, want nil", err)
		}
		after, err := os.ReadFile(cfgPath)
		if err != nil {
			t.Fatal(err)
		}
		if string(after) != body {
			t.Errorf("config changed on a zero-entry config:\n%s", after)
		}
	})
}

// AC-CGP-012 — the two home-scope flags are mutually exclusive.
func TestCleanCmdRejectsBothScopeFlags(t *testing.T) {
	cmd := newCleanCmd()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--home", "--codex-skills"})

	if err := cmd.Execute(); err == nil {
		t.Fatalf("both scope flags were accepted; want a usage error")
	}
}

// AC-CGP-015 — the help text matches the command's real blast radius.
func TestCleanCmdHelpNamesCodexScope(t *testing.T) {
	long := newCleanCmd().Long
	if strings.Contains(long, "Only ~/.moai is touched") {
		t.Errorf("the help still claims only ~/.moai is touched:\n%s", long)
	}
	if !strings.Contains(long, "~/.codex/config.toml") {
		t.Errorf("the help does not disclose the ~/.codex/config.toml scope:\n%s", long)
	}
}

// AC-CGP-010 — one path-shape classifier, not two.
func TestSinglePathShapeClassifier(t *testing.T) {
	for _, dir := range []string{"internal/cli", "internal/codexwiring"} {
		root := filepath.Join("..", "..", dir)
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return err
			}
			src, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			body := string(src)
			// The one sanctioned combination lives in classifyCodexSkillPath.
			if strings.Contains(body, "filepath.IsAbs") && strings.Contains(body, `HasPrefix(p, "~`) &&
				!strings.Contains(body, "func classifyCodexSkillPath") {
				return fmt.Errorf("%s recombines IsAbs with a ~ prefix check: a second shape classifier", path)
			}
			return nil
		})
		if err != nil {
			t.Error(err)
		}
	}
}

// AC-CGP-011 — the parser stays read-only.
func TestSkillsParserStaysReadOnly(t *testing.T) {
	src, err := os.ReadFile(filepath.Join("..", "..", "internal", "codexwiring", "skills.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"os.WriteFile", "os.Create", "os.OpenFile", "io.Writer"} {
		if strings.Contains(string(src), forbidden) {
			t.Errorf("skills.go references %s; the parser must not write", forbidden)
		}
	}
}
