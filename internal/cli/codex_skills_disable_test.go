package cli

// codex_skills_disable_test.go — SPEC-CODEX-SKILL-DISABLE-001 (t502).
//
// Every fixture lives under t.TempDir(). No test reads or writes the real
// ~/.codex/config.toml: the config location rides the CODEX_HOME env var and
// the project/home roots ride explicit options, so both are pointed at the
// temp tree.
//
// Tests here reassign package-level seams (codexSkillWriteFileFn,
// skillsProjectRootFn) or set env vars and therefore MUST NOT call
// t.Parallel().

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/printer"
	"github.com/modu-ai/moai-adk/internal/codexwiring"
)

// ─────────────────────────────────────────────────────────────────────────
// helpers
// ─────────────────────────────────────────────────────────────────────────

// countEntriesWithPath counts declared [[skills.config]] entries whose path
// key equals want. It reads through the shipped parser rather than grepping,
// so the count is the one the consuming code will also see.
func countEntriesWithPath(t *testing.T, content []byte, want string) int {
	t.Helper()
	n := 0
	for _, e := range codexwiring.ParseSkillEntries(content) {
		if e.Path == want {
			n++
		}
	}
	return n
}

// entryWithPath returns the single entry declaring want, failing when the
// count is not exactly one.
func entryWithPath(t *testing.T, content []byte, want string) codexwiring.SkillEntry {
	t.Helper()
	var found []codexwiring.SkillEntry
	for _, e := range codexwiring.ParseSkillEntries(content) {
		if e.Path == want {
			found = append(found, e)
		}
	}
	if len(found) != 1 {
		t.Fatalf("entries declaring %q = %d, want 1\n--- content ---\n%s", want, len(found), content)
	}
	return found[0]
}

// entryDeclaresEnabledKey reports whether the entry's own line range carries
// an `enabled` assignment at all. The parser's tri-state cannot answer this on
// its own: Unspecified also covers an unrecognised value. An entry lacking the
// key is a HARD start failure for the user's codex (t504 F1), so the schema
// axis needs its own reading.
func entryDeclaresEnabledKey(t *testing.T, content []byte, e codexwiring.SkillEntry) bool {
	t.Helper()
	lines, _ := codexwiring.SplitConfigLines(content)
	for i := e.StartLine; i < e.EndLine && i < len(lines); i++ {
		if strings.HasPrefix(strings.TrimSpace(lines[i]), "enabled") {
			return true
		}
	}
	return false
}

// mirrorFixture builds a project tree in the shape internal/template/
// skill_mirror.go actually creates, and returns the project root.
//
//	symlink -> MirrorModeSymlink (.agents/skills/<n> -> ../../.claude/skills/<n>)
//	copy    -> MirrorModeCopy, shape-identical to MirrorModeSkipped
func mirrorFixture(t *testing.T, mode, skill string) string {
	t.Helper()
	proj := t.TempDir()
	src := filepath.Join(proj, ".claude", "skills", skill)
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatalf("mkdir src: %v", err)
	}
	if err := os.WriteFile(filepath.Join(src, "SKILL.md"), []byte("---\nname: "+skill+"\n---\n"), 0o644); err != nil {
		t.Fatalf("write src SKILL.md: %v", err)
	}
	mirror := filepath.Join(proj, ".agents", "skills")
	if err := os.MkdirAll(mirror, 0o755); err != nil {
		t.Fatalf("mkdir mirror: %v", err)
	}
	switch mode {
	case "symlink":
		if err := os.Symlink(filepath.Join("..", "..", ".claude", "skills", skill), filepath.Join(mirror, skill)); err != nil {
			t.Skipf("symlinks unavailable here: %v", err)
		}
	case "copy":
		dst := filepath.Join(mirror, skill)
		if err := os.MkdirAll(dst, 0o755); err != nil {
			t.Fatalf("mkdir mirror skill: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dst, "SKILL.md"), []byte("---\nname: "+skill+"\n---\n"), 0o644); err != nil {
			t.Fatalf("write mirror SKILL.md: %v", err)
		}
	default:
		t.Fatalf("unknown mirror mode %q", mode)
	}
	return proj
}

// emptyHome returns a home directory carrying no ~/.agents/skills mirror, so
// the home root contributes no candidate.
func emptyHome(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}

// codexHomeWith writes config.toml into a fresh CODEX_HOME and points the
// env var at it. It returns the home dir and the config path.
func codexHomeWith(t *testing.T, content string, mode fs.FileMode) (string, string) {
	t.Helper()
	home := t.TempDir()
	cfg := filepath.Join(home, "config.toml")
	if err := os.WriteFile(cfg, []byte(content), mode); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := os.Chmod(cfg, mode); err != nil {
		t.Fatalf("chmod config: %v", err)
	}
	t.Setenv("CODEX_HOME", home)
	return home, cfg
}

func mustRead(t *testing.T, p string) []byte {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read %s: %v", p, err)
	}
	return b
}

// ─────────────────────────────────────────────────────────────────────────
// §A publication
// ─────────────────────────────────────────────────────────────────────────

// AC-CSD-001 — a config declaring no entry for P gains exactly ONE, carrying
// both keys, with enabled = false.
//
// Three axes, three mutant cells (plan §F): the append axis (a no-op append
// leaves zero entries), the value axis (enabled = true), and the schema axis
// (no enabled key at all — a HARD start failure for the user's codex).
func TestUpsertCodexSkillDisableAppendsOneEntryWithFalse(t *testing.T) {
	p := "/proj/.agents/skills/probe/SKILL.md"
	in := []byte("[model]\nname = \"gpt\"\n")

	out, v := upsertCodexSkillDisable(in, p)

	// append axis
	if got := countEntriesWithPath(t, out, p); got != 1 {
		t.Fatalf("entries declaring %q = %d, want exactly 1\n--- out ---\n%s", p, got, out)
	}
	if v.Action != codexSkillDisableAppended {
		t.Errorf("verdict action = %v, want appended", v.Action)
	}
	e := entryWithPath(t, out, p)

	// schema axis — the key must be PRESENT, not merely read as false
	if !entryDeclaresEnabledKey(t, out, e) {
		t.Errorf("entry declares no `enabled` key; a codex reading this config fails to start\n--- out ---\n%s", out)
	}
	// value axis
	if e.Enabled != codexwiring.SkillEnabledFalse {
		t.Errorf("enabled = %v, want SkillEnabledFalse\n--- out ---\n%s", e.Enabled, out)
	}
	// the pre-existing table survives
	if !strings.Contains(string(out), "[model]") {
		t.Errorf("pre-existing [model] table lost\n--- out ---\n%s", out)
	}
}

// AC-CSD-001, insert branch — an entry declaring `path` and NO `enabled` is
// precisely the shape codex hard-fails on (`missing field 'enabled' in
// skills.config`, rc=1). The merge must ADD the key, not leave the entry as it
// found it.
//
// Emission happens at more than one site: the append composer builds a whole
// entry, and this branch inserts the single missing key into an entry that
// already exists. Mutant 3 of AC-CSD-001 says an implementation with the
// enabled line removed must go RED, and without this test that is only true of
// the append site — deleting the insert emission passed GREEN, leaving the half
// of the guard that faces a total-outage hazard unexercised.
//
// No pre-implementation RED exists for this test: the branch was already
// correct when it was written. The mutant is therefore the only honest evidence
// that it discriminates — delete the emission at the insert site and this goes
// red. After this test, removing EITHER emission point turns AC-CSD-001 red, so
// the criterion's wording matches what the tests actually cover.
func TestUpsertCodexSkillDisableInsertsMissingEnabledKey(t *testing.T) {
	p := "/proj/.agents/skills/probe/SKILL.md"
	in := []byte("[[skills.config]]\npath = \"" + p + "\"\n\n[tail]\nk = 1\n")

	out, v := upsertCodexSkillDisable(in, p)

	if v.Action != codexSkillDisableUpdated {
		t.Errorf("action = %v, want updated", v.Action)
	}
	if got := countEntriesWithPath(t, out, p); got != 1 {
		t.Fatalf("entries declaring %q = %d, want 1\n--- out ---\n%s", p, got, out)
	}
	e := entryWithPath(t, out, p)
	if !entryDeclaresEnabledKey(t, out, e) {
		t.Errorf("the entry still declares no `enabled` key — codex refuses to start on this config\n--- out ---\n%s", out)
	}
	if e.Enabled != codexwiring.SkillEnabledFalse {
		t.Errorf("enabled = %v, want SkillEnabledFalse\n--- out ---\n%s", e.Enabled, out)
	}
	// The insert shifts every later line by one, so what follows is the part
	// most likely to be damaged by getting that wrong.
	for _, want := range []string{"[tail]", "k = 1"} {
		if !strings.Contains(string(out), want) {
			t.Errorf("lost %q past the insert point\n--- out ---\n%s", want, out)
		}
	}
}

// AC-CSD-002 — the published notation is the absolute LITERAL mirror path, in
// BOTH mirror shapes. The resolved `.claude/…` twin is silently inert under a
// copy-fallback mirror (gate-path-shape.md, cell Cres), so publishing it works
// on a symlinking machine and dies quietly on everyone else's.
func TestResolveCodexSkillMirrorPathPublishesLiteralMirrorPath(t *testing.T) {
	const skill = "t502probe"
	for _, mode := range []string{"symlink", "copy"} {
		t.Run(mode, func(t *testing.T) {
			proj := mirrorFixture(t, mode, skill)
			home := emptyHome(t)

			got := resolveCodexSkillMirrorPath(proj, home, skill)

			if got.Outcome != codexSkillResolved {
				t.Fatalf("outcome = %v (%s), want resolved", got.Outcome, got.Reason)
			}
			want := filepath.Join(proj, ".agents", "skills", skill, "SKILL.md")
			if got.Path != want {
				t.Errorf("path = %q, want %q", got.Path, want)
			}
			// The forbidden notation, named explicitly: it must never be what
			// we publish, in either shape.
			forbidden := filepath.Join(proj, ".claude", "skills", skill, "SKILL.md")
			if got.Path == forbidden {
				t.Errorf("published the resolved `.claude/…` twin %q — inert under a copy-fallback mirror", got.Path)
			}
			if !filepath.IsAbs(got.Path) {
				t.Errorf("path %q is not absolute — a relative notation is inert (cell E6)", got.Path)
			}
			if strings.HasSuffix(got.Path, string(filepath.Separator)+skill) {
				t.Errorf("path %q is directory-shaped — inert in both shapes (cells E4/E5)", got.Path)
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────
// §B name resolution — three failures, distinguishable exits
// ─────────────────────────────────────────────────────────────────────────

// AC-CSD-010 — the dry-run report names the absolute mirror path and exits 0.
func TestRunCodexSkillDisableDryRunNamesResolvedPath(t *testing.T) {
	const skill = "t502probe"
	proj := mirrorFixture(t, "copy", skill)
	_, cfg := codexHomeWith(t, "# empty\n", 0o600)
	before := mustRead(t, cfg)

	var out, errBuf bytes.Buffer
	p := printer.New(printer.WithWriters(&out, &errBuf))
	err := runCodexSkillDisable(p, codexSkillDisableOptions{
		Skill: skill, ProjectRoot: proj, HomeDir: emptyHome(t),
	})
	if err != nil {
		t.Fatalf("dry run returned error: %v", err)
	}
	want := filepath.Join(proj, ".agents", "skills", skill, "SKILL.md")
	if body := out.String() + errBuf.String(); !strings.Contains(body, want) {
		t.Errorf("report does not name %q\n--- report ---\n%s", want, body)
	}
	if !bytes.Equal(before, mustRead(t, cfg)) {
		t.Errorf("dry run modified the config")
	}
}

// AC-CSD-011 — an unresolvable name: NON-ZERO exit, config byte-invariant. A
// typo must be detectable from a script.
func TestRunCodexSkillDisableUnresolvedNameFailsWithoutWriting(t *testing.T) {
	proj := mirrorFixture(t, "copy", "t502probe")
	_, cfg := codexHomeWith(t, "# empty\n", 0o600)
	before := mustRead(t, cfg)

	var out, errBuf bytes.Buffer
	p := printer.New(printer.WithWriters(&out, &errBuf))
	err := runCodexSkillDisable(p, codexSkillDisableOptions{
		Skill: "no-such-skill", ProjectRoot: proj, HomeDir: emptyHome(t), Force: true,
	})
	if err == nil {
		t.Fatalf("unresolvable name returned nil error — a typo would exit 0")
	}
	if !bytes.Equal(before, mustRead(t, cfg)) {
		t.Errorf("config changed on the reject path")
	}
}

// AC-CSD-012 — the same name in two candidate roots: NON-ZERO exit, config
// byte-invariant. The home mirror really does reach codex's root list even
// under CODEX_HOME isolation (gate-path-shape.md, root r0), so this branch is
// a reachable state rather than a hypothetical.
func TestRunCodexSkillDisableAmbiguousNameFailsWithoutWriting(t *testing.T) {
	const skill = "t502probe"
	proj := mirrorFixture(t, "copy", skill)
	home := t.TempDir()
	homeSkill := filepath.Join(home, ".agents", "skills", skill)
	if err := os.MkdirAll(homeSkill, 0o755); err != nil {
		t.Fatalf("mkdir home mirror: %v", err)
	}
	if err := os.WriteFile(filepath.Join(homeSkill, "SKILL.md"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write home SKILL.md: %v", err)
	}
	_, cfg := codexHomeWith(t, "# empty\n", 0o600)
	before := mustRead(t, cfg)

	var out, errBuf bytes.Buffer
	p := printer.New(printer.WithWriters(&out, &errBuf))
	err := runCodexSkillDisable(p, codexSkillDisableOptions{
		Skill: skill, ProjectRoot: proj, HomeDir: home, Force: true,
	})
	if err == nil {
		t.Fatalf("ambiguous name returned nil error")
	}
	if !bytes.Equal(before, mustRead(t, cfg)) {
		t.Errorf("config changed on the ambiguity path")
	}
}

// AC-CSD-013 — no mirror at all: exit 0, config byte-invariant, and the
// reason NAMED. The mirror is produced by a deployment run, not by a
// checkout, so its absence is an ordinary state (this repository is in it).
func TestRunCodexSkillDisableMirrorAbsentExitsZero(t *testing.T) {
	proj := t.TempDir() // no .agents/skills at all
	_, cfg := codexHomeWith(t, "# empty\n", 0o600)
	before := mustRead(t, cfg)

	var out, errBuf bytes.Buffer
	p := printer.New(printer.WithWriters(&out, &errBuf))
	err := runCodexSkillDisable(p, codexSkillDisableOptions{
		Skill: "anything", ProjectRoot: proj, HomeDir: emptyHome(t), Force: true,
	})
	if err != nil {
		t.Fatalf("mirror absence returned error %v — absence is an ordinary state", err)
	}
	body := out.String() + errBuf.String()
	if !strings.Contains(body, "mirror") {
		t.Errorf("reason does not name the mirror\n--- report ---\n%s", body)
	}
	if !bytes.Equal(before, mustRead(t, cfg)) {
		t.Errorf("config changed")
	}
}

// ─────────────────────────────────────────────────────────────────────────
// §C merge — idempotent, non-destructive
// ─────────────────────────────────────────────────────────────────────────

// AC-CSD-020 — an existing entry is UPDATED in place; no duplicate appears.
func TestUpsertCodexSkillDisableUpdatesExistingEntry(t *testing.T) {
	p := "/proj/.agents/skills/probe/SKILL.md"
	in := []byte("[[skills.config]]\npath = \"" + p + "\"\nenabled = true\n")

	out, v := upsertCodexSkillDisable(in, p)

	if got := countEntriesWithPath(t, out, p); got != 1 {
		t.Fatalf("entries declaring %q = %d, want 1\n--- out ---\n%s", p, got, out)
	}
	if v.Action != codexSkillDisableUpdated {
		t.Errorf("action = %v, want updated", v.Action)
	}
	if e := entryWithPath(t, out, p); e.Enabled != codexwiring.SkillEnabledFalse {
		t.Errorf("enabled = %v, want false\n--- out ---\n%s", e.Enabled, out)
	}
}

// AC-CSD-021 — re-running over an already-false entry returns the INPUT
// bytes. Returning a re-joined copy would let a round-trip defect masquerade
// as a no-op run (the discipline prune already keeps).
func TestUpsertCodexSkillDisableIdempotentBytes(t *testing.T) {
	p := "/proj/.agents/skills/probe/SKILL.md"
	in := []byte("# lead comment\n[[skills.config]]\npath = \"" + p + "\"\nenabled = false\n")

	out, v := upsertCodexSkillDisable(in, p)

	if !bytes.Equal(in, out) {
		t.Errorf("bytes changed on a no-op run\n--- in ---\n%s\n--- out ---\n%s", in, out)
	}
	if v.Action != codexSkillDisableUnchanged {
		t.Errorf("action = %v, want unchanged", v.Action)
	}
	if v.Reason == "" {
		t.Errorf("no-op run reports no reason")
	}
}

// AC-CSD-022 — unrelated entries, comments, blank lines, CRLF endings and a
// missing final terminator all survive the merge.
func TestUpsertCodexSkillDisablePreservesSurroundings(t *testing.T) {
	target := "/proj/.agents/skills/probe/SKILL.md"
	other := "/proj/.agents/skills/other/SKILL.md"

	t.Run("unrelated entries and comments", func(t *testing.T) {
		in := []byte("# keep me\n\n[[skills.config]]\npath = \"" + other + "\"\nenabled = true\n\n[tail]\nk = 1\n")
		out, _ := upsertCodexSkillDisable(in, target)
		for _, want := range []string{"# keep me", other, "enabled = true", "[tail]", "k = 1"} {
			if !strings.Contains(string(out), want) {
				t.Errorf("lost %q\n--- out ---\n%s", want, out)
			}
		}
		if got := countEntriesWithPath(t, out, other); got != 1 {
			t.Errorf("unrelated entry count = %d, want 1", got)
		}
		// Without this the subtest would pass under a no-op merge: everything
		// is preserved when nothing happens.
		if got := countEntriesWithPath(t, out, target); got != 1 {
			t.Errorf("target entry count = %d, want 1 — preservation is only meaningful alongside the append", got)
		}
		if e := entryWithPath(t, out, other); e.Enabled != codexwiring.SkillEnabledTrue {
			t.Errorf("unrelated entry's enabled changed to %v", e.Enabled)
		}
	})

	t.Run("CRLF line endings", func(t *testing.T) {
		in := []byte("# keep me\r\n[[skills.config]]\r\npath = \"" + other + "\"\r\nenabled = true\r\n")
		out, _ := upsertCodexSkillDisable(in, target)
		if bytes.Contains(bytes.ReplaceAll(out, []byte("\r\n"), nil), []byte("\n")) {
			t.Errorf("a bare LF appeared in a CRLF file\n--- out ---\n%q", out)
		}
		if !bytes.HasPrefix(out, in) {
			t.Errorf("the original CRLF prefix was rewritten\n--- out ---\n%q", out)
		}
		// A no-op merge preserves CRLF perfectly, so the ending check only
		// discriminates once the appended entry is actually there.
		if got := countEntriesWithPath(t, out, target); got != 1 {
			t.Errorf("target entry count = %d, want 1", got)
		}
		if !bytes.Contains(out, []byte("enabled = false\r\n")) {
			t.Errorf("the appended entry does not use the file's CRLF ending\n--- out ---\n%q", out)
		}
	})

	t.Run("no trailing newline", func(t *testing.T) {
		in := []byte("[model]\nname = \"gpt\"")
		out, _ := upsertCodexSkillDisable(in, target)
		if bytes.HasSuffix(out, []byte("\n")) {
			t.Errorf("a trailing terminator was introduced\n--- out ---\n%q", out)
		}
		if got := countEntriesWithPath(t, out, target); got != 1 {
			t.Errorf("entry count = %d, want 1\n--- out ---\n%s", got, out)
		}
	})
}

// AC-CSD-023 — an entry whose range holds a line the parser did not recognise
// is left alone, and the skip is reported. Every line a multi-line literal
// swallowed still LOOKS like a header or a key; only the parser can tell.
func TestUpsertCodexSkillDisableSkipsUnrecognisedEntry(t *testing.T) {
	p := "/proj/.agents/skills/probe/SKILL.md"
	in := []byte("[[skills.config]]\npath = \"" + p + "\"\nenabled = true\nstray = \"\"\"\n")

	out, v := upsertCodexSkillDisable(in, p)

	if !bytes.Equal(in, out) {
		t.Errorf("an entry carrying an unrecognised line was modified\n--- out ---\n%q", out)
	}
	if v.Action != codexSkillDisableSkipped {
		t.Errorf("action = %v, want skipped", v.Action)
	}
	if v.Reason == "" {
		t.Errorf("skip reports no reason")
	}
}

// ─────────────────────────────────────────────────────────────────────────
// §D runner — dry-run, backup, mode, fail-open
// ─────────────────────────────────────────────────────────────────────────

// AC-CSD-030(b) — the layer must be named in the invocation. Writing into the
// user's HOME happens only when the user asked for that layer by name.
func TestSkillsDisableRequiresCodexFlag(t *testing.T) {
	cmd := newSkillsCmd()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"disable", "anything"})
	if err := cmd.Execute(); err == nil {
		t.Fatalf("disable ran without --codex — the write layer was never named")
	}
}

// AC-CSD-031 — the backup lands first, at 0600, and the TARGET keeps its own
// permission mode. A verb whose headline property is "non-destructive" must
// not silently tighten the user's 0644 config.
//
// Mutant: writing the target unconditionally at 0600 must turn this RED.
func TestRunCodexSkillDisableBacksUpAndPreservesMode(t *testing.T) {
	const skill = "t502probe"
	proj := mirrorFixture(t, "copy", skill)
	home, cfg := codexHomeWith(t, "# start\n", 0o644)
	before := mustRead(t, cfg)

	var out, errBuf bytes.Buffer
	p := printer.New(printer.WithWriters(&out, &errBuf))
	if err := runCodexSkillDisable(p, codexSkillDisableOptions{
		Skill: skill, ProjectRoot: proj, HomeDir: emptyHome(t), Force: true,
	}); err != nil {
		t.Fatalf("forced run: %v", err)
	}

	st, err := os.Stat(cfg)
	if err != nil {
		t.Fatalf("stat config: %v", err)
	}
	if got := st.Mode().Perm(); got != 0o644 {
		t.Errorf("config mode = %04o, want 0644 preserved", got)
	}

	entries, err := os.ReadDir(home)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	var backup string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "config.toml.bak-") {
			backup = filepath.Join(home, e.Name())
		}
	}
	if backup == "" {
		t.Fatalf("no backup written next to the config")
	}
	if !bytes.Equal(before, mustRead(t, backup)) {
		t.Errorf("backup content differs from the pre-write original")
	}
	bst, err := os.Stat(backup)
	if err != nil {
		t.Fatalf("stat backup: %v", err)
	}
	if got := bst.Mode().Perm(); got != 0o600 {
		t.Errorf("backup mode = %04o, want 0600", got)
	}
	if body := out.String() + errBuf.String(); !strings.Contains(body, sha256Hex(before)) {
		t.Errorf("the pre-write sha256 was not reported\n--- report ---\n%s", body)
	}
	if got := countEntriesWithPath(t, mustRead(t, cfg), filepath.Join(proj, ".agents", "skills", skill, "SKILL.md")); got != 1 {
		t.Errorf("post-write entry count = %d, want 1", got)
	}
}

// AC-CSD-031 (second half) — when the backup cannot be written, the target is
// byte-invariant. A write that outruns its backup is unrecoverable.
func TestRunCodexSkillDisableBackupFailureLeavesTargetIntact(t *testing.T) {
	const skill = "t502probe"
	proj := mirrorFixture(t, "copy", skill)
	_, cfg := codexHomeWith(t, "# start\n", 0o644)
	before := mustRead(t, cfg)

	orig := codexSkillWriteFileFn
	t.Cleanup(func() { codexSkillWriteFileFn = orig })
	codexSkillWriteFileFn = func(name string, data []byte, perm fs.FileMode) error {
		if strings.Contains(filepath.Base(name), ".bak-") {
			return errors.New("injected backup failure")
		}
		return orig(name, data, perm)
	}

	var out, errBuf bytes.Buffer
	p := printer.New(printer.WithWriters(&out, &errBuf))
	if err := runCodexSkillDisable(p, codexSkillDisableOptions{
		Skill: skill, ProjectRoot: proj, HomeDir: emptyHome(t), Force: true,
	}); err == nil {
		t.Fatalf("backup failure returned nil error")
	}
	if !bytes.Equal(before, mustRead(t, cfg)) {
		t.Errorf("the target was written despite the backup failing")
	}
}

// AC-CSD-030(a) — the default run reports what it WOULD write and writes
// nothing.
func TestRunCodexSkillDisableDryRunIsDefault(t *testing.T) {
	const skill = "t502probe"
	proj := mirrorFixture(t, "copy", skill)
	home, cfg := codexHomeWith(t, "# start\n", 0o600)
	before := mustRead(t, cfg)

	var out, errBuf bytes.Buffer
	p := printer.New(printer.WithWriters(&out, &errBuf))
	if err := runCodexSkillDisable(p, codexSkillDisableOptions{
		Skill: skill, ProjectRoot: proj, HomeDir: emptyHome(t),
	}); err != nil {
		t.Fatalf("dry run: %v", err)
	}
	if !bytes.Equal(before, mustRead(t, cfg)) {
		t.Errorf("the dry run wrote to the config")
	}
	entries, _ := os.ReadDir(home)
	for _, e := range entries {
		if strings.Contains(e.Name(), ".bak-") {
			t.Errorf("the dry run left a backup %s", e.Name())
		}
	}
	if body := out.String() + errBuf.String(); !strings.Contains(strings.ToLower(body), "dry-run") {
		t.Errorf("the dry run does not announce itself\n--- report ---\n%s", body)
	}
}

// AC-CSD-032 — an absent input is not an error. There is simply nothing to
// disable.
func TestRunCodexSkillDisableFailsOpenOnAbsentConfig(t *testing.T) {
	const skill = "t502probe"
	proj := mirrorFixture(t, "copy", skill)
	t.Setenv("CODEX_HOME", filepath.Join(t.TempDir(), "no-such-home"))

	var out, errBuf bytes.Buffer
	p := printer.New(printer.WithWriters(&out, &errBuf))
	if err := runCodexSkillDisable(p, codexSkillDisableOptions{
		Skill: skill, ProjectRoot: proj, HomeDir: emptyHome(t), Force: true,
	}); err != nil {
		t.Fatalf("absent config returned error %v — a missing input is not an error", err)
	}
}

// AC-CSD-033 — five states, five distinct reasons. One phrase covering all of
// them teaches the user nothing about what to fix.
func TestRunCodexSkillDisableReasonsAreDistinct(t *testing.T) {
	const skill = "t502probe"
	target := func(proj string) string {
		return filepath.Join(proj, ".agents", "skills", skill, "SKILL.md")
	}

	reasons := map[string]string{}
	record := func(name string, opts func(t *testing.T) (codexSkillDisableOptions, string)) {
		t.Run(name, func(t *testing.T) {
			o, cfg := opts(t)
			var out, errBuf bytes.Buffer
			p := printer.New(printer.WithWriters(&out, &errBuf))
			_ = runCodexSkillDisable(p, o)
			_ = cfg
			reasons[name] = out.String() + errBuf.String()
		})
	}

	record("unresolved", func(t *testing.T) (codexSkillDisableOptions, string) {
		proj := mirrorFixture(t, "copy", skill)
		_, cfg := codexHomeWith(t, "# x\n", 0o600)
		return codexSkillDisableOptions{Skill: "no-such", ProjectRoot: proj, HomeDir: emptyHome(t), Force: true}, cfg
	})
	record("ambiguous", func(t *testing.T) (codexSkillDisableOptions, string) {
		proj := mirrorFixture(t, "copy", skill)
		home := t.TempDir()
		d := filepath.Join(home, ".agents", "skills", skill)
		_ = os.MkdirAll(d, 0o755)
		_ = os.WriteFile(filepath.Join(d, "SKILL.md"), []byte("x"), 0o644)
		_, cfg := codexHomeWith(t, "# x\n", 0o600)
		return codexSkillDisableOptions{Skill: skill, ProjectRoot: proj, HomeDir: home, Force: true}, cfg
	})
	record("mirror-absent", func(t *testing.T) (codexSkillDisableOptions, string) {
		_, cfg := codexHomeWith(t, "# x\n", 0o600)
		return codexSkillDisableOptions{Skill: skill, ProjectRoot: t.TempDir(), HomeDir: emptyHome(t), Force: true}, cfg
	})
	record("unrecognised-line", func(t *testing.T) (codexSkillDisableOptions, string) {
		proj := mirrorFixture(t, "copy", skill)
		_, cfg := codexHomeWith(t, "[[skills.config]]\npath = \""+target(proj)+"\"\nenabled = true\nstray = \"\"\"\n", 0o600)
		return codexSkillDisableOptions{Skill: skill, ProjectRoot: proj, HomeDir: emptyHome(t), Force: true}, cfg
	})
	record("already-false", func(t *testing.T) (codexSkillDisableOptions, string) {
		proj := mirrorFixture(t, "copy", skill)
		_, cfg := codexHomeWith(t, "[[skills.config]]\npath = \""+target(proj)+"\"\nenabled = false\n", 0o600)
		return codexSkillDisableOptions{Skill: skill, ProjectRoot: proj, HomeDir: emptyHome(t), Force: true}, cfg
	})

	if len(reasons) != 5 {
		t.Fatalf("collected %d reports, want 5", len(reasons))
	}
	seen := map[string]string{}
	for name, body := range reasons {
		if body == "" {
			t.Errorf("%s: reported nothing", name)
			continue
		}
		if prev, dup := seen[body]; dup {
			t.Errorf("%s and %s report the same text:\n%s", prev, name, body)
			continue
		}
		seen[body] = name
	}
}
