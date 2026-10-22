package cli

// codex_skills_prune_readback_test.go — SPEC-CODEX-SKILL-PATH-READBACK-001
// (t562) M1. The stat target fed to osStatFn must be the fromConfigPath
// conversion of the declared path (the t540 seam), applied in the
// codexPathAbsolute branch only — classification itself stays on the DECLARED
// form (spec.md §B.5, REQ-CSRB-003).
//
// Tests here reassign package-level seams (configPathSeparator, osStatFn,
// codexUserHomeDir) and therefore MUST NOT call t.Parallel(). Every override
// restores via t.Cleanup — the same discipline codex_skills_prune_test.go
// carries.
//
// The separator seam is overridden to '\\' in the conversion tests: on this
// darwin host sep '/' makes fromConfigPath the identity, so only an injected
// separator makes the Windows semantics observable here (the seam file's own
// stated reason for taking sep as a parameter).

import (
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/codexwiring"
)

// overrideSeparator pins the config-path separator seam to '\\' and restores
// it via t.Cleanup.
func overrideSeparator(t *testing.T) {
	t.Helper()
	orig := configPathSeparator
	t.Cleanup(func() { configPathSeparator = orig })
	configPathSeparator = '\\'
}

// pruneReadbackStatRecorder records every path handed to the osStatFn seam.
// Named card-uniquely (not "statRecorder"): the absorbed t563 test file
// doctor_codex_stale_skill_test.go declares its own statRecorder type in this
// package, and a second declaration of that name broke the package test
// binary's compilation when the two files met in the absorb merge.
type pruneReadbackStatRecorder struct {
	calls []string
	err   error // what the stub reports once it has recorded the call
}

// stubRecordingStat replaces osStatFn with a recorder that reports err for
// every call and restores the seam via t.Cleanup.
func stubRecordingStat(t *testing.T, err error) *pruneReadbackStatRecorder {
	t.Helper()
	rec := &pruneReadbackStatRecorder{err: err}
	orig := osStatFn
	t.Cleanup(func() { osStatFn = orig })
	osStatFn = func(name string) (os.FileInfo, error) {
		rec.calls = append(rec.calls, name)
		return nil, rec.err
	}
	return rec
}

// readbackEntry builds a fully-recognised entry whose declared path carries no
// line the parser would disqualify, so the judgment reaches the stat switch.
func readbackEntry(declared string) codexwiring.SkillEntry {
	return codexwiring.SkillEntry{
		Path:                  declared,
		Enabled:               codexwiring.SkillEnabledTrue,
		StartLine:             0,
		EndLine:               3,
		FirstUnrecognizedLine: -1,
	}
}

// TestJudgeCodexSkillEntry_SeparatorConversion (AC-CSRB-002 — the behavioral
// RED) asserts the stat TARGET is the converted form of the declaration, not
// the declaration itself. No deletion is performed: the assertion is on the
// stat argument and the verdict object only (REQ-CSRB-007 boundary).
func TestJudgeCodexSkillEntry_SeparatorConversion(t *testing.T) {
	// A host-absolute slash path classifies codexPathAbsolute on darwin.
	declared := filepath.Join(t.TempDir(), "gone", "SKILL.md")
	overrideSeparator(t)
	rec := stubRecordingStat(t, fs.ErrNotExist)

	v := judgeCodexSkillEntry(readbackEntry(declared))

	want := fromConfigPath(declared, '\\')
	if len(rec.calls) == 0 {
		t.Fatalf("osStatFn was never called; want a stat of %q", want)
	}
	if got := rec.calls[0]; got != want {
		t.Errorf("stat target = %q, want the converted form %q (declared %q)", got, want, declared)
	}
	if !v.Eligible {
		t.Errorf("verdict = %+v, want Eligible — the recorded stat is fs.ErrNotExist", v)
	}
}

// TestJudgeCodexSkillEntry_ClassifiesDeclaredFormBeforeConversion
// (AC-CSRB-003 — ordering guard): classification runs on the DECLARED string,
// so a slash-form relative declaration is skipped without a single stat call
// on every host. The forbidden convert-first order would
// emit the oddly-formed skip reason instead — a distinct string, so the
// ordering is observable here.
func TestJudgeCodexSkillEntry_ClassifiesDeclaredFormBeforeConversion(t *testing.T) {
	overrideSeparator(t)
	rec := stubRecordingStat(t, fs.ErrNotExist)

	v := judgeCodexSkillEntry(readbackEntry("relative/skill/SKILL.md"))

	if len(rec.calls) != 0 {
		t.Errorf("osStatFn called with %v, want zero calls — a relative declaration is never statted", rec.calls)
	}
	want := "relative path — no observed resolution base"
	if v.SkipReason != want {
		t.Errorf("SkipReason = %q, want %q (classification ran on the declared form)", v.SkipReason, want)
	}
	if v.Eligible {
		t.Errorf("verdict = %+v, want not eligible", v)
	}
}

// TestJudgeCodexSkillEntry_HomeRelativeStatTargetStaysNative
// (AC-CSRB-004 — no-double-conversion guard): the home-relative branch hands
// the native filepath.Join product to stat EXACTLY. The blanket-wrap mutant
// (fromConfigPath applied at the stat site rather than in the absolute branch)
// rewrites the Join product under the injected '\\' and fails this test.
func TestJudgeCodexSkillEntry_HomeRelativeStatTargetStaysNative(t *testing.T) {
	tmp := t.TempDir()
	stubHome(t, tmp)
	overrideSeparator(t)
	rec := stubRecordingStat(t, fs.ErrNotExist)

	v := judgeCodexSkillEntry(readbackEntry("~/x/SKILL.md"))

	want := filepath.Join(tmp, "x/SKILL.md")
	if len(rec.calls) == 0 {
		t.Fatalf("osStatFn was never called; want a stat of %q", want)
	}
	if got := rec.calls[0]; got != want {
		t.Errorf("stat target = %q, want the native Join product %q unrewritten", got, want)
	}
	if !v.Eligible {
		t.Errorf("verdict = %+v, want Eligible — the recorded stat is fs.ErrNotExist", v)
	}
}

// TestJudgeCodexSkillEntry_EligibilityGatingPins (AC-CSRB-005): the exact
// two-condition eligibility gate — classification ∈ {absolute, home-relative}
// AND stat error fs.ErrNotExist — unchanged by this card. Verdicts only; no
// arm performs a deletion.
func TestJudgeCodexSkillEntry_EligibilityGatingPins(t *testing.T) {
	t.Run("missing means eligible", func(t *testing.T) {
		declared := filepath.Join(t.TempDir(), "gone", "SKILL.md")
		rec := stubRecordingStat(t, fs.ErrNotExist)

		v := judgeCodexSkillEntry(readbackEntry(declared))

		if !v.Eligible || v.SkipReason != "" {
			t.Errorf("verdict = %+v, want Eligible with empty SkipReason", v)
		}
		if len(rec.calls) != 1 {
			t.Errorf("osStatFn calls = %v, want exactly one", rec.calls)
		}
	})

	t.Run("resolving means kept", func(t *testing.T) {
		declared := filepath.Join(t.TempDir(), "present", "SKILL.md")
		rec := stubRecordingStat(t, nil)

		v := judgeCodexSkillEntry(readbackEntry(declared))

		want := "the path resolves"
		if v.Eligible {
			t.Errorf("verdict = %+v, want not eligible", v)
		}
		if v.SkipReason != want {
			t.Errorf("SkipReason = %q, want %q", v.SkipReason, want)
		}
		if len(rec.calls) != 1 {
			t.Errorf("osStatFn calls = %v, want exactly one", rec.calls)
		}
	})

	t.Run("backslash declaration is oddly formed and never statted", func(t *testing.T) {
		// Keep this declaration non-absolute on every supported host.
		// A native Windows drive path is legitimately absolute and belongs
		// to the positive ordering control below, not the no-stat arm.
		rec := stubRecordingStat(t, fs.ErrNotExist)

		v := judgeCodexSkillEntry(readbackEntry(`relative\skill\SKILL.md`))

		if len(rec.calls) != 0 {
			t.Errorf("osStatFn calls = %v, want zero — an oddly-formed declaration is never statted", rec.calls)
		}
		want := "oddly-formed path — not resolvable here"
		if v.SkipReason != want {
			t.Errorf("SkipReason = %q, want %q", v.SkipReason, want)
		}
		if v.Eligible {
			t.Errorf("verdict = %+v, want not eligible", v)
		}
	})
}

// Keep the original drive-shaped declarations as explicit host-truth controls.
// All rows execute: native absolute paths are statted exactly once and become
// eligible only for the observed ErrNotExist; foreign shapes remain non-destructive.
func TestJudgeCodexSkillEntry_DriveDeclarationsRespectHost(t *testing.T) {
	for _, declared := range []string{"C:/Users/u/SKILL.md", `C:\Users\u\SKILL.md`} {
		t.Run(declared, func(t *testing.T) {
			rec := stubRecordingStat(t, fs.ErrNotExist)
			v := judgeCodexSkillEntry(readbackEntry(declared))
			if runtime.GOOS == "windows" {
				if !v.Eligible || v.SkipReason != "" || len(rec.calls) != 1 || rec.calls[0] != fromConfigPath(declared, configPathSeparator) {
					t.Errorf("native absolute declaration lost its exact stat/eligibility contract: verdict=%+v calls=%q", v, rec.calls)
				}
				return
			}
			wantReason := "relative path — no observed resolution base"
			if strings.ContainsRune(declared, '\\') {
				wantReason = "oddly-formed path — not resolvable here"
			}
			if v.Eligible || v.SkipReason != wantReason || len(rec.calls) != 0 {
				t.Errorf("foreign declaration must remain non-destructive: verdict=%+v calls=%q, want reason %q", v, rec.calls, wantReason)
			}
		})
	}
}
