package cli

// codex_skills_disable_path_test.go — SPEC-CODEX-SKILL-PATH-SLASH-001 (t540) M2.
//
// These tests cover the publisher half of the separator seam: the conversion
// applied before the "cannot carry verbatim" guard, and the comparison that
// decides whether an already-declared entry is UPDATED or a second one is
// appended.
//
// Every test here reassigns the package-level configPathSeparator var and
// therefore MUST NOT call t.Parallel() — the same discipline the osStatFn and
// codexUserHomeDir seams already carry in this package.
//
// They also close t502 F4: before this file, the guard at
// codex_skills_disable.go:245 had ZERO coverage — measured, `/usr/bin/grep -rn
// 'cannot carry verbatim' --include='*.go' internal/` matched the production
// line and nothing else. Both of the guard's branches now have a test: the
// refuse branch (AC-CSPS-003) and the branch a Windows separator no longer
// reaches (AC-CSPS-002).

import (
	"bytes"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/codexwiring"
)

// withConfigPathSeparator points the package seam at sep for the duration of
// one test. Non-parallel by construction: the caller must not call
// t.Parallel(), and t.Cleanup restores the production value even when the
// test fails mid-way.
func withConfigPathSeparator(t *testing.T, sep rune) {
	t.Helper()
	orig := configPathSeparator
	t.Cleanup(func() { configPathSeparator = orig })
	configPathSeparator = sep
}

// countConfigHeaders counts [[skills.config]] headers by reading the raw
// lines rather than the parser, because the property under test in
// AC-CSPS-007 is "no SECOND entry was appended" — a claim about what is
// written, which a parser that groups or skips entries could mask.
func countConfigHeaders(content []byte) int {
	n := 0
	lines, _ := codexwiring.SplitConfigLines(content)
	for _, l := range lines {
		if strings.TrimSpace(strings.TrimSuffix(l, "\r")) == "[[skills.config]]" {
			n++
		}
	}
	return n
}

// ─────────────────────────────────────────────────────────────────────────
// AC-CSPS-002 — a backslash-shaped path is published, not skipped
// ─────────────────────────────────────────────────────────────────────────

// TestUpsertCodexSkillDisablePublishesWindowsPathInSlashForm is the measured
// half of inference I1. Before this change, every path on a Windows host
// carried the host separator, so the guard refused every invocation and the
// verb reported rc=0 while doing nothing — inoperative on the one platform
// this card exists to serve.
//
// The separator seam is what makes that measurable here rather than only on a
// Windows host: filepath.ToSlash is the identity on this host (spec.md M6), so
// without the injected separator the backslash literal below would survive to
// the guard and this test could not pass on darwin at all.
//
// RED before the change: Action == codexSkillDisableSkipped.
func TestUpsertCodexSkillDisablePublishesWindowsPathInSlashForm(t *testing.T) {
	withConfigPathSeparator(t, '\\')

	native := `C:\Users\u\.codex\skills\probe\SKILL.md`
	slash := "C:/Users/u/.codex/skills/probe/SKILL.md"
	in := []byte("[model]\nname = \"gpt\"\n")

	out, v := upsertCodexSkillDisable(in, native)

	if v.Action != codexSkillDisableAppended {
		t.Fatalf("action = %v (reason %q), want appended", v.Action, v.Reason)
	}
	if want := `path = "` + slash + `"`; !strings.Contains(string(out), want) {
		t.Errorf("emitted content does not carry %s\n--- out ---\n%s", want, out)
	}
	// No escape sequence, and no raw backslash either: the parser reads the
	// value verbatim, so either would name a file nothing on the moai side
	// can resolve.
	if strings.Contains(string(out), `\`) {
		t.Errorf("emitted content carries a backslash; the config must never hold one\n--- out ---\n%s", out)
	}
	// The parser — the consumer that actually decides what this entry names —
	// must read back exactly the slash form.
	if got := countEntriesWithPath(t, out, slash); got != 1 {
		t.Errorf("entries declaring %q = %d, want exactly 1\n--- out ---\n%s", slash, got, out)
	}
	if e := entryWithPath(t, out, slash); e.Enabled != codexwiring.SkillEnabledFalse {
		t.Errorf("enabled = %v, want SkillEnabledFalse\n--- out ---\n%s", e.Enabled, out)
	}
	if !strings.Contains(string(out), "[model]") {
		t.Errorf("pre-existing [model] table lost\n--- out ---\n%s", out)
	}
}

// ─────────────────────────────────────────────────────────────────────────
// AC-CSPS-003 — genuinely unrepresentable characters are still refused
// ─────────────────────────────────────────────────────────────────────────

// TestUpsertCodexSkillDisableStillRefusesUnrepresentableChars keeps the guard
// honest in the direction the card does NOT change. A quote, an LF, or a CR
// remains unrepresentable after conversion on every separator, so each must
// still be refused and the content handed back byte-unchanged.
//
// Running it under both separators is deliberate: REQ-CSPS-003 says the
// refusal applies to what remains AFTER conversion, and only the '\\' arm
// exercises that ordering.
func TestUpsertCodexSkillDisableStillRefusesUnrepresentableChars(t *testing.T) {
	in := []byte("[model]\nname = \"gpt\"\n")

	for _, sep := range []rune{'/', '\\'} {
		for _, tc := range []struct {
			name string
			path string
		}{
			{"quote", `/proj/.agents/skills/pr"obe/SKILL.md`},
			{"lf", "/proj/.agents/skills/pr\nobe/SKILL.md"},
			{"cr", "/proj/.agents/skills/pr\robe/SKILL.md"},
		} {
			t.Run(string(sep)+"/"+tc.name, func(t *testing.T) {
				withConfigPathSeparator(t, sep)

				out, v := upsertCodexSkillDisable(in, tc.path)

				if v.Action != codexSkillDisableSkipped {
					t.Fatalf("action = %v, want skipped", v.Action)
				}
				if !strings.Contains(v.Reason, "cannot carry verbatim") {
					t.Errorf("reason = %q, want the verbatim-refusal reason", v.Reason)
				}
				if !bytes.Equal(out, in) {
					t.Errorf("content changed on a refusal\n--- out ---\n%s", out)
				}
			})
		}
	}
}

// TestUpsertCodexSkillDisableRefusesBackslashOnSlashHost is REQ-CSPS-010, and
// it is the assertion that forbids an unconditional
// strings.ReplaceAll(p, "\\", "/") inside toConfigPath.
//
// A unix filename may legally contain a backslash. Today such a path is
// REFUSED, which is the correct outcome — the format cannot carry it
// verbatim. An unconditional replacement would publish a DIFFERENT path
// instead, one naming a file that does not exist, which the prune verb then
// classifies as absent and deletes. That is strictly worse than the defect
// this card repairs, so the refusal must survive.
//
// The separator is pinned to '/' rather than left at its production value so
// the criterion's state-driven premise ("while the host separator is '/'") is
// the tested state on every host, instead of being true on darwin by accident
// and silently skipped on Windows.
func TestUpsertCodexSkillDisableRefusesBackslashOnSlashHost(t *testing.T) {
	withConfigPathSeparator(t, '/')

	in := []byte("[model]\nname = \"gpt\"\n")
	// A legal unix filename that happens to carry a backslash.
	p := `/proj/.agents/skills/we\ird/SKILL.md`

	out, v := upsertCodexSkillDisable(in, p)

	if v.Action != codexSkillDisableSkipped {
		t.Fatalf("action = %v, want skipped — a backslash-bearing unix path must still be refused, never rewritten", v.Action)
	}
	if !strings.Contains(v.Reason, "cannot carry verbatim") {
		t.Errorf("reason = %q, want the verbatim-refusal reason", v.Reason)
	}
	if !bytes.Equal(out, in) {
		t.Errorf("content changed on a refusal\n--- out ---\n%s", out)
	}
}

// ─────────────────────────────────────────────────────────────────────────
// AC-CSPS-007 — an existing backslash entry is updated, not duplicated
// ─────────────────────────────────────────────────────────────────────────

// TestUpsertCodexSkillDisableUpdatesExistingBackslashEntry pins the second
// half of M2. Normalizing only the publication point would leave the single
// comparison in this verb (codex_skills_disable.go:252) failing against an
// entry already stored in backslash form — one written by hand or by another
// tool — and the fall-through would append a SECOND entry for the same skill.
// The next invocation would then trip the duplicate skip, handing cleanup to
// `moai clean --codex-skills` (t506). The card would be manufacturing exactly
// the debt that verb exists to clear.
//
// The stored path line stays in its original backslash form: this verb
// rewrites only `enabled`, and rewriting a declaration it did not author is
// out of scope. Asserting that non-change is part of the criterion.
//
// RED before the change: Action == codexSkillDisableAppended.
func TestUpsertCodexSkillDisableUpdatesExistingBackslashEntry(t *testing.T) {
	withConfigPathSeparator(t, '\\')

	native := `C:\Users\u\.codex\skills\probe\SKILL.md`
	pathLine := `path = "` + native + `"`
	in := []byte("[[skills.config]]\n" + pathLine + "\nenabled = true\n")

	out, v := upsertCodexSkillDisable(in, native)

	if v.Action != codexSkillDisableUpdated {
		t.Fatalf("action = %v (reason %q), want updated\n--- out ---\n%s", v.Action, v.Reason, out)
	}
	if got := countConfigHeaders(out); got != 1 {
		t.Errorf("[[skills.config]] headers = %d, want exactly 1 — a second entry was appended\n--- out ---\n%s", got, out)
	}
	if !strings.Contains(string(out), pathLine) {
		t.Errorf("the stored path line was rewritten; this verb rewrites only `enabled`\n--- out ---\n%s", out)
	}
	if strings.Contains(string(out), "enabled = true") {
		t.Errorf("enabled is still true\n--- out ---\n%s", out)
	}
	e := entryWithPath(t, out, native)
	if e.Enabled != codexwiring.SkillEnabledFalse {
		t.Errorf("enabled = %v, want SkillEnabledFalse\n--- out ---\n%s", e.Enabled, out)
	}
}

// TestUpsertCodexSkillDisableSkipsMixedShapeDuplicates is AC-CSPS-007's
// second arm.
//
// It is NOT guarding against two comparison sites disagreeing — there is
// exactly one comparison, at :252, and the duplicate branch at :257 performs
// none of its own, it counts the slice that comparison filled. The arm exists
// because normalization changes WHICH entries land in that slice: a config
// holding one backslash-shaped and one slash-shaped entry for the same skill
// now matches BOTH where it previously matched one. The duplicate branch must
// still be reached, rather than one entry being silently updated or a third
// appended.
func TestUpsertCodexSkillDisableSkipsMixedShapeDuplicates(t *testing.T) {
	withConfigPathSeparator(t, '\\')

	native := `C:\Users\u\.codex\skills\probe\SKILL.md`
	slash := "C:/Users/u/.codex/skills/probe/SKILL.md"
	in := []byte("" +
		"[[skills.config]]\npath = \"" + native + "\"\nenabled = true\n\n" +
		"[[skills.config]]\npath = \"" + slash + "\"\nenabled = true\n")

	out, v := upsertCodexSkillDisable(in, native)

	if v.Action != codexSkillDisableSkipped {
		t.Fatalf("action = %v, want skipped\n--- out ---\n%s", v.Action, out)
	}
	if !strings.Contains(v.Reason, "2 entries") {
		t.Errorf("reason = %q, want the duplicate-count skip naming 2 entries", v.Reason)
	}
	if !bytes.Equal(out, in) {
		t.Errorf("content changed on a duplicate skip\n--- out ---\n%s", out)
	}
}
