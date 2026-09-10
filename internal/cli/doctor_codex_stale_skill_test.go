package cli

// SPEC-DOCTOR-STAT-SEAM-001 M1 — unit-level characterization tests for
// codexStaleSkillFinding (the stat seam's site B) and the skill-mirror
// inspection loop's state (site A's inspectSkillMirror).
//
// These tests PIN CURRENT BEHAVIOR on the unmodified tree. They are not
// RED-first: their definition is that they pass BEFORE the M2 seam swap
// exists, so the M2 output-identity proof can show identical findings on the
// same fixtures. They are struct-level — the codexFinding fields and the
// skillMirrorState fields asserted directly — which is finer-grained than the
// rendered-Message/Detail assertions of the indirect checkCodexWiring suite
// in doctor_codex_test.go. That suite is the M2 identity-proof comparison
// set; this file SUPPLEMENTS it and deliberately does not duplicate its
// bucket fixtures.
//
// Seam discipline (REQ-007): osStatFn is NOT overridden here — M1 pins real
// fixture behavior through the real stat. The codexUserHomeDir home seam IS
// overridden, serially (never with t.Parallel), where the unresolvable-home
// arm cannot be driven by real fixtures alone.

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// stubCodexHomeUnresolvable pins the home seam to fail, driving the arms a
// real fixture cannot reach: with CODEX_HOME blank the config itself is then
// unlocatable (a silent skip before any entry is read), and with CODEX_HOME
// set the config is read but a "~" expansion fails mid-loop, which the
// finding must count as INDETERMINATE, never missing (fail-open, not a
// guess). Serial only — the seam is a package-level variable.
func stubCodexHomeUnresolvable(t *testing.T) {
	t.Helper()
	orig := codexUserHomeDir
	codexUserHomeDir = func() (string, error) { return "", errors.New("t563: user home unresolvable") }
	t.Cleanup(func() { codexUserHomeDir = orig })
}

// assertStaleSilent pins the fail-open posture: a config the check cannot
// judge (no entries, no path in any entry, an unresolvable home, an absent
// or unreadable config) yields ok=false and a zero finding.
func assertStaleSilent(t *testing.T, f codexFinding, ok bool) {
	t.Helper()
	if ok {
		t.Fatalf("expected a silent skip (ok=false), got a finding: %+v", f)
	}
	if f.summary != "" || f.detail != "" || f.severity != codexSeverityAdvisory {
		t.Errorf("silent skip carried a non-zero finding: %+v", f)
	}
}

// TestCodexStaleSkillFinding_AbsoluteExistingNotCounted: an absolute
// registration that exists — as a FILE or as a DIRECTORY (directories resolve
// by design; a missing verdict on one would advise deleting a registration
// on a guess) — produces no finding at all.
func TestCodexStaleSkillFinding_AbsoluteExistingNotCounted(t *testing.T) {
	home := writeCodexHomeConfig(t, []codexSkillEntrySpec{
		{Path: liveSkillFile(t), EnabledKey: "true"},
		{Path: t.TempDir(), EnabledKey: "true"}, // a directory resolves too
	})
	stubCodexHome(t, home)

	f, ok := codexStaleSkillFinding()
	assertStaleSilent(t, f, ok)
}

// TestCodexStaleSkillFinding_AbsoluteMissingSplitsEnabledStates pins the
// four-way per-`enabled`-state split at the struct level: the summary counts
// only genuinely-missing paths, the declared denominator is separate, the
// split is quantified in Detail with the fourth member rendered
// unconditionally (REQ-SSF-004), and a path-less entry is counted in the
// denominator but never as missing. Severity is the ADVISORY zero value.
func TestCodexStaleSkillFinding_AbsoluteMissingSplitsEnabledStates(t *testing.T) {
	home := writeCodexHomeConfig(t, []codexSkillEntrySpec{
		{Path: absentSkillPath(t, "t563-enabled"), EnabledKey: "true"},
		{Path: absentSkillPath(t, "t563-disabled"), EnabledKey: "false"},
		{Path: absentSkillPath(t, "t563-unspecified")}, // no `enabled` key
		{Path: absentSkillPath(t, "t563-nonboolean"), EnabledKey: `"quoted"`},
		{EnabledKey: "true"}, // declares no path — in the total, never missing
	})
	stubCodexHome(t, home)

	f, ok := codexStaleSkillFinding()
	if !ok {
		t.Fatalf("four missing paths produced no finding")
	}
	if f.severity != codexSeverityAdvisory {
		t.Errorf("stale finding severity = %v, want advisory (zero value)", f.severity)
	}
	if want := "~/.codex/config.toml: 4 stale skill entries"; f.summary != want {
		t.Errorf("summary = %q, want %q", f.summary, want)
	}
	// Detail's leading token is the RESOLVED config path (a fixture temp path,
	// so only the phrase after it is pinned); the display form is summary-only.
	for _, want := range []string{
		" declares 5 [[skills.config]] entries",
		"; 4 with a path that no longer exists (1 enabled, 1 disabled, 1 unspecified, 1 non-boolean)",
		"remove the stale entries or restore the skill files",
	} {
		if !strings.Contains(f.detail, want) {
			t.Errorf("detail missing %q:\n%s", want, f.detail)
		}
	}
}

// TestCodexStaleSkillFinding_HomeRelativeExpansionBuckets: a "~"-prefixed
// declaration expands against the USER home (the codexUserHomeDir seam). An
// expansion target that exists is NOT counted missing; one that does not is
// counted under its declared enabled state.
func TestCodexStaleSkillFinding_HomeRelativeExpansionBuckets(t *testing.T) {
	home := writeCodexHomeConfig(t, []codexSkillEntrySpec{
		{Path: "~/t563-live/SKILL.md", EnabledKey: "true"},
		{Path: "~/t563-gone/SKILL.md", EnabledKey: "false"},
	})
	t468HomeSkillFile(t, home, "t563-live/SKILL.md")
	stubCodexHome(t, home)

	f, ok := codexStaleSkillFinding()
	if !ok {
		t.Fatalf("missing ~-relative path produced no finding")
	}
	if want := "~/.codex/config.toml: 1 stale skill entry"; f.summary != want {
		t.Errorf("summary = %q, want %q (the live ~ entry must not inflate the count)", f.summary, want)
	}
	if !strings.Contains(f.detail, "; 1 with a path that no longer exists (0 enabled, 1 disabled, 0 unspecified, 0 non-boolean)") {
		t.Errorf("detail does not carry the missing ~-relative entry under its declared state:\n%s", f.detail)
	}
}

// TestCodexStaleSkillFinding_BareTildeResolvesHome: a bare "~" expands to the
// user home itself, which exists — the entry is not stale and stays silent.
func TestCodexStaleSkillFinding_BareTildeResolvesHome(t *testing.T) {
	home := writeCodexHomeConfig(t, []codexSkillEntrySpec{
		{Path: "~", EnabledKey: "true"},
	})
	stubCodexHome(t, home)

	f, ok := codexStaleSkillFinding()
	assertStaleSilent(t, f, ok)
}

// TestCodexStaleSkillFinding_RelativeShapeNeverCountedMissing: a relative
// declaration has no observed resolution base, so it is reported as its own
// classification and NEVER stat'ed — never missing, and the remove directive
// never attaches to it.
func TestCodexStaleSkillFinding_RelativeShapeNeverCountedMissing(t *testing.T) {
	home := writeCodexHomeConfig(t, []codexSkillEntrySpec{
		{Path: "skills/t563-a/SKILL.md", EnabledKey: "true"},
		{Path: "skills/t563-b/SKILL.md", EnabledKey: "true"},
	})
	stubCodexHome(t, home)

	f, ok := codexStaleSkillFinding()
	if !ok {
		t.Fatalf("relative-only config produced no finding")
	}
	if want := "~/.codex/config.toml: 2 skill entries with a path shape this check cannot resolve"; f.summary != want {
		t.Errorf("summary = %q, want the non-destructive shape summary %q", f.summary, want)
	}
	if !strings.Contains(f.detail, "; 2 relative entries (not checked: the resolution base is not observed)") {
		t.Errorf("detail does not carry the relative classification:\n%s", f.detail)
	}
	if strings.Contains(f.detail, "remove the stale entries") {
		t.Errorf("remove directive attached to a relative-only config:\n%s", f.detail)
	}
}

// TestCodexStaleSkillFinding_OddlyFormedShapeNeverCountedMissing: a
// backslash-bearing non-absolute fragment and another user's "~user" form are
// both oddly-formed — reported per classification, never missing.
func TestCodexStaleSkillFinding_OddlyFormedShapeNeverCountedMissing(t *testing.T) {
	home := writeCodexHomeConfig(t, []codexSkillEntrySpec{
		{Path: `skills\t563-foo\SKILL.md`, EnabledKey: "true"},
		{Path: "~otheruser/skills/t563/SKILL.md", EnabledKey: "true"},
	})
	stubCodexHome(t, home)

	f, ok := codexStaleSkillFinding()
	if !ok {
		t.Fatalf("oddly-formed-only config produced no finding")
	}
	if want := "~/.codex/config.toml: 2 skill entries with a path shape this check cannot resolve"; f.summary != want {
		t.Errorf("summary = %q, want the non-destructive shape summary %q", f.summary, want)
	}
	if !strings.Contains(f.detail, "; 2 oddly-formed entries (not checked: backslash or ~other-user shape)") {
		t.Errorf("detail does not carry the oddly-formed classification:\n%s", f.detail)
	}
	if strings.Contains(f.detail, "remove the stale entries") {
		t.Errorf("remove directive attached to an oddly-formed-only config:\n%s", f.detail)
	}
}

// TestCodexStaleSkillFinding_UnresolvableHomeExpansionIndeterminate: with
// CODEX_HOME locating the config, a "~" expansion consults the home seam —
// and a seam that FAILS makes the path's existence INDETERMINATE, disclosed
// in Detail, never folded into missing. The summary keeps counting only the
// genuinely-missing anchor.
func TestCodexStaleSkillFinding_UnresolvableHomeExpansionIndeterminate(t *testing.T) {
	home := writeCodexHomeConfig(t, []codexSkillEntrySpec{
		{Path: absentSkillPath(t, "t563-anchor"), EnabledKey: "true"},
		{Path: "~/t563-unreachable/SKILL.md", EnabledKey: "true"},
	})
	// CODEX_HOME points at the CONFIG DIR (the home's .codex), not the home
	// root: the resolver joins only the file's base name onto it. Located this
	// way, the config is read WITHOUT the home seam.
	t.Setenv(codexHomeEnvVar, filepath.Join(home, ".codex"))
	stubCodexHomeUnresolvable(t)

	f, ok := codexStaleSkillFinding()
	if !ok {
		t.Fatalf("anchor missing + indeterminate expansion produced no finding")
	}
	if want := "$CODEX_HOME/config.toml: 1 stale skill entry"; f.summary != want {
		t.Errorf("summary = %q, want %q (the env-sourced display, counting the anchor only)", f.summary, want)
	}
	for _, want := range []string{
		" declares 2 [[skills.config]] entries",
		"; 1 with a path that no longer exists (1 enabled, 0 disabled, 0 unspecified, 0 non-boolean)",
		"; a further 1 could not be checked and are NOT counted as missing",
	} {
		if !strings.Contains(f.detail, want) {
			t.Errorf("detail missing %q:\n%s", want, f.detail)
		}
	}
}

// TestCodexStaleSkillFinding_IndeterminateOnlyStaysSilent (the t451
// posture): an indeterminate-only config — every path whose expansion cannot
// be resolved — stays silent. An unobserved absence is never reported.
func TestCodexStaleSkillFinding_IndeterminateOnlyStaysSilent(t *testing.T) {
	home := writeCodexHomeConfig(t, []codexSkillEntrySpec{
		{Path: "~/t563-unreachable/SKILL.md", EnabledKey: "true"},
	})
	t.Setenv(codexHomeEnvVar, filepath.Join(home, ".codex")) // config dir, not home root
	stubCodexHomeUnresolvable(t)

	f, ok := codexStaleSkillFinding()
	assertStaleSilent(t, f, ok)
}

// TestCodexStaleSkillFinding_EmptyPathEntrySkipped: an entry declaring no
// path says nothing about a file's existence — it is counted in the declared
// total and never as missing, and a config of only such entries is silent.
func TestCodexStaleSkillFinding_EmptyPathEntrySkipped(t *testing.T) {
	home := writeCodexHomeConfig(t, []codexSkillEntrySpec{
		{EnabledKey: "true"},
	})
	stubCodexHome(t, home)

	f, ok := codexStaleSkillFinding()
	assertStaleSilent(t, f, ok)
}

// TestCodexStaleSkillFinding_UnreadableConfigFailsOpen: a CODEX_HOME whose
// config.toml does not exist yields a silent skip — a missing input never
// becomes a finding.
func TestCodexStaleSkillFinding_UnreadableConfigFailsOpen(t *testing.T) {
	t.Setenv(codexHomeEnvVar, t.TempDir()) // no .codex/config.toml inside

	f, ok := codexStaleSkillFinding()
	assertStaleSilent(t, f, ok)
}

// TestCodexStaleSkillFinding_UnresolvableCodexHomeSkipsLoop: with the home
// seam failing AND CODEX_HOME unset, the config cannot even be located — the
// entry loop is never entered and the whole sub-check fails open.
func TestCodexStaleSkillFinding_UnresolvableCodexHomeSkipsLoop(t *testing.T) {
	t.Setenv(codexHomeEnvVar, "")
	stubCodexHomeUnresolvable(t)

	f, ok := codexStaleSkillFinding()
	assertStaleSilent(t, f, ok)
}

// TestCodexStaleSkillFinding_SymlinkLoopIndeterminateNotMissing preserves the
// t451 stat-error taxonomy at the struct level: a real missing entry is
// counted, a symlink loop (ELOOP, not ErrNotExist) stays indeterminate and is
// disclosed in Detail, never folded into missing.
func TestCodexStaleSkillFinding_SymlinkLoopIndeterminateNotMissing(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation needs privileges on windows")
	}
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
		{Path: absentSkillPath(t, "t563-real-miss"), EnabledKey: "true"},
		{Path: loopA, EnabledKey: "true"},
	})
	stubCodexHome(t, home)

	f, ok := codexStaleSkillFinding()
	if !ok {
		t.Fatalf("missing + looped entries produced no finding")
	}
	if want := "~/.codex/config.toml: 1 stale skill entry"; f.summary != want {
		t.Errorf("summary = %q, want %q (the loop must not inflate missing)", f.summary, want)
	}
	if !strings.Contains(f.detail, "; a further 1 could not be checked and are NOT counted as missing") {
		t.Errorf("detail does not disclose the unchecked loop entry:\n%s", f.detail)
	}
}

// ---------------------------------------------------------------------------
// M2 — stat-seam observation tests (REQ-006). These override osStatFn with a
// RECORDING or INJECTING function (save → replace → t.Cleanup restore, never
// t.Parallel — the seam is a package-level variable) to pin WHICH path each
// stat site probes and WHICH bucket each stat result drives. The M1 tests
// above never touch the seam; these are the only tests in this file that do.
// ---------------------------------------------------------------------------

// statRecorder collects every path argument passed through the overridden
// osStatFn while forwarding to the real os.Stat, so the observation changes
// nothing about the check's behavior.
type statRecorder struct {
	paths []string
}

// stubStatRecording installs a forwarding recording osStatFn and returns the
// recorder. Serial only.
func stubStatRecording(t *testing.T) *statRecorder {
	t.Helper()
	rec := &statRecorder{}
	orig := osStatFn
	osStatFn = func(p string) (fs.FileInfo, error) {
		rec.paths = append(rec.paths, p)
		return os.Stat(p)
	}
	t.Cleanup(func() { osStatFn = orig })
	return rec
}

// stubStatInject overrides osStatFn so mapped paths return their mapped error
// VERBATIM (a nil mapped error resolves, backed by a real FileInfo from a
// live directory) while everything else forwards to the real os.Stat. This is
// what proves the bucket follows the INJECTED result rather than the disk
// state — the portable stat-failure injection the seam exists to enable.
func stubStatInject(t *testing.T, results map[string]error) {
	t.Helper()
	info, err := os.Stat(t.TempDir()) // a real resolving FileInfo to hand back
	if err != nil {
		t.Fatal(err)
	}
	orig := osStatFn
	osStatFn = func(p string) (fs.FileInfo, error) {
		if e, mapped := results[p]; mapped {
			if e != nil {
				return nil, e
			}
			return info, nil
		}
		return os.Stat(p)
	}
	t.Cleanup(func() { osStatFn = orig })
}

// TestCodexStaleSkillFinding_StatSeamRecordsClassifiedPaths (site B): the
// stat path argument is the CLASSIFIED path — the declared absolute, and the
// seam-expanded home-relative — and ONLY for those two classifications:
// relative, oddly-formed, and path-less entries produce ZERO recorded stat
// calls (REQ-003's skip-without-stat contract, now observable).
func TestCodexStaleSkillFinding_StatSeamRecordsClassifiedPaths(t *testing.T) {
	absent := absentSkillPath(t, "t563-seam-abs")
	home := writeCodexHomeConfig(t, []codexSkillEntrySpec{
		{Path: absent, EnabledKey: "true"},
		{Path: "~/t563-seam-gone/SKILL.md", EnabledKey: "true"},
		{Path: "skills/t563-rel/SKILL.md", EnabledKey: "true"},
		{Path: `skills\t563-odd\SKILL.md`, EnabledKey: "true"},
		{Path: "~otheruser/x", EnabledKey: "true"},
		{EnabledKey: "true"},
	})
	stubCodexHome(t, home)
	rec := stubStatRecording(t)

	f, ok := codexStaleSkillFinding()
	if !ok {
		t.Fatalf("missing classified paths produced no finding")
	}
	if want := "~/.codex/config.toml: 2 stale skill entries"; f.summary != want {
		t.Errorf("summary = %q, want %q", f.summary, want)
	}
	wantPaths := []string{
		absent, // codexPathAbsolute: declared verbatim
		filepath.Join(home, "t563-seam-gone", "SKILL.md"), // codexPathHomeRelative: expanded
	}
	if len(rec.paths) != len(wantPaths) {
		t.Fatalf("recorded stat calls = %v, want exactly %v", rec.paths, wantPaths)
	}
	for i, want := range wantPaths {
		if rec.paths[i] != want {
			t.Errorf("recorded stat arg[%d] = %q, want %q", i, rec.paths[i], want)
		}
	}
}

// TestCodexStaleSkillFinding_InjectedErrNotExistDrivesMissing: a path that
// EXISTS on disk is counted missing when the injected stat result is
// fs.ErrNotExist — the bucket follows the injected result, not the disk.
func TestCodexStaleSkillFinding_InjectedErrNotExistDrivesMissing(t *testing.T) {
	live := liveSkillFile(t)
	home := writeCodexHomeConfig(t, []codexSkillEntrySpec{{Path: live, EnabledKey: "true"}})
	stubCodexHome(t, home)
	stubStatInject(t, map[string]error{live: fs.ErrNotExist})

	f, ok := codexStaleSkillFinding()
	if !ok {
		t.Fatalf("injected ErrNotExist produced no finding")
	}
	if want := "~/.codex/config.toml: 1 stale skill entry"; f.summary != want {
		t.Errorf("summary = %q, want %q (injected ErrNotExist must drive missing)", f.summary, want)
	}
	if !strings.Contains(f.detail, "; 1 with a path that no longer exists (1 enabled, 0 disabled, 0 unspecified, 0 non-boolean)") {
		t.Errorf("detail does not carry the injected-missing entry:\n%s", f.detail)
	}
}

// TestCodexStaleSkillFinding_InjectedNilDrivesResolves: a path that is ABSENT
// on disk resolves when the injected stat result is nil — the resolves arm is
// the injected verdict, so the entry stays silent.
func TestCodexStaleSkillFinding_InjectedNilDrivesResolves(t *testing.T) {
	absent := absentSkillPath(t, "t563-seam-nil")
	home := writeCodexHomeConfig(t, []codexSkillEntrySpec{{Path: absent, EnabledKey: "true"}})
	stubCodexHome(t, home)
	stubStatInject(t, map[string]error{absent: nil})

	f, ok := codexStaleSkillFinding()
	assertStaleSilent(t, f, ok)
}

// TestCodexStaleSkillFinding_InjectedOtherErrorDrivesIndeterminate: a stat
// error that is neither nil nor ErrNotExist folds the entry into the
// indeterminate counter — disclosed in Detail, never counted missing.
func TestCodexStaleSkillFinding_InjectedOtherErrorDrivesIndeterminate(t *testing.T) {
	pMissing := liveSkillFile(t)
	pOther := liveSkillFile(t)
	home := writeCodexHomeConfig(t, []codexSkillEntrySpec{
		{Path: pMissing, EnabledKey: "true"},
		{Path: pOther, EnabledKey: "true"},
	})
	stubCodexHome(t, home)
	stubStatInject(t, map[string]error{
		pMissing: fs.ErrNotExist,
		pOther:   errors.New("t563: injected stat failure (not ENOENT)"),
	})

	f, ok := codexStaleSkillFinding()
	if !ok {
		t.Fatalf("missing + injected-error entries produced no finding")
	}
	if want := "~/.codex/config.toml: 1 stale skill entry"; f.summary != want {
		t.Errorf("summary = %q, want %q (the other-error entry must not inflate missing)", f.summary, want)
	}
	if !strings.Contains(f.detail, "; a further 1 could not be checked and are NOT counted as missing") {
		t.Errorf("detail does not disclose the injected-error entry:\n%s", f.detail)
	}
}

// ---------------------------------------------------------------------------
// Site A — the skill-mirror inspection. inspectSkillMirror is READ-only, so
// the skillMirrorState it returns is asserted directly per fixture shape.
// ---------------------------------------------------------------------------

// TestInspectSkillMirror_AbsentMirrorIsNotIndeterminate: a root with no
// `.agents/skills` reports genuinely-absent — the one state that may raise
// the absent finding — and nothing else.
func TestInspectSkillMirror_AbsentMirrorIsNotIndeterminate(t *testing.T) {
	st := inspectSkillMirror(t.TempDir())

	if st.dirPresent {
		t.Errorf("absent mirror reported dirPresent: %+v", st)
	}
	if st.dirIndeterminate {
		t.Errorf("absent mirror reported dirIndeterminate: %+v", st)
	}
	if st.copyMode != 0 || st.unmirrored != 0 || st.indeterminate != 0 || len(st.dangling) != 0 {
		t.Errorf("absent mirror carried non-zero counters: %+v", st)
	}
}

// TestInspectSkillMirror_ResolvingSymlinkUncounted: a relative symlink whose
// target exists (the producer's own link-body shape) hits no bucket at all.
func TestInspectSkillMirror_ResolvingSymlinkUncounted(t *testing.T) {
	root := t.TempDir()
	writeSkillMirror(t, root, []string{"live"}, []mirrorEntrySpec{
		{name: "live", kind: mirrorEntryLive},
	})

	st := inspectSkillMirror(root)
	if !st.dirPresent {
		t.Fatalf("present mirror reported dirPresent=false: %+v", st)
	}
	if st.copyMode != 0 || st.unmirrored != 0 || st.indeterminate != 0 || len(st.dangling) != 0 {
		t.Errorf("resolving symlink counted in some bucket: %+v", st)
	}
}

// TestInspectSkillMirror_DanglingSymlinkNamedInDangling: an entry whose link
// body no longer resolves is recorded BY NAME in dangling — exactly the
// enumeration the dangling finding renders.
func TestInspectSkillMirror_DanglingSymlinkNamedInDangling(t *testing.T) {
	root := t.TempDir()
	writeSkillMirror(t, root, []string{"live"}, []mirrorEntrySpec{
		{name: "live", kind: mirrorEntryLive},
		{name: "ghost", kind: mirrorEntryDangling},
	})

	st := inspectSkillMirror(root)
	if len(st.dangling) != 1 || st.dangling[0] != "ghost" {
		t.Errorf("dangling = %v, want [ghost]: %+v", st.dangling, st)
	}
	if st.indeterminate != 0 || st.copyMode != 0 || st.unmirrored != 0 {
		t.Errorf("dangling entry leaked into another counter: %+v", st)
	}
}

// TestInspectSkillMirror_RealDirCountsCopyModeAndUnmirrored: a real directory
// on a mirror path is the copy fallback's materialization (copyMode), and a
// canonical skill directory absent from the mirror is unmirrored — both
// reported-only counts, neither a finding.
func TestInspectSkillMirror_RealDirCountsCopyModeAndUnmirrored(t *testing.T) {
	root := t.TempDir()
	writeSkillMirror(t, root, []string{"copied", "unlinked"}, []mirrorEntrySpec{
		{name: "copied", kind: mirrorEntryRealDir},
	})

	st := inspectSkillMirror(root)
	if !st.dirPresent {
		t.Fatalf("present mirror reported dirPresent=false: %+v", st)
	}
	if st.copyMode != 1 {
		t.Errorf("copyMode = %d, want 1: %+v", st.copyMode, st)
	}
	if st.unmirrored != 1 {
		t.Errorf("unmirrored = %d, want 1 (the canonical dir with no mirror entry): %+v", st.unmirrored, st)
	}
	if st.indeterminate != 0 || len(st.dangling) != 0 {
		t.Errorf("copy-mode fixture leaked into a finding counter: %+v", st)
	}
}

// TestInspectSkillMirror_RegularFileOnMirrorPathUncounted pins the source
// comment's contract: a regular file occupying a mirror path is a shape no
// requirement classifies, so it is counted as NOTHING — no copyMode, no
// dangling, no indeterminate.
func TestInspectSkillMirror_RegularFileOnMirrorPathUncounted(t *testing.T) {
	root := t.TempDir()
	mirrorDir := filepath.Join(root, ".agents", "skills")
	if err := os.MkdirAll(mirrorDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, ".claude", "skills"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mirrorDir, "stray.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	st := inspectSkillMirror(root)
	if !st.dirPresent {
		t.Fatalf("present mirror reported dirPresent=false: %+v", st)
	}
	if st.copyMode != 0 || st.indeterminate != 0 || len(st.dangling) != 0 || st.unmirrored != 0 {
		t.Errorf("regular file on a mirror path was counted somewhere: %+v", st)
	}
}

// TestInspectSkillMirror_SymlinkLoopEntryIsIndeterminate: a stat error that
// is NOT ErrNotExist (a symlink loop's ELOOP) folds into indeterminate —
// never into dangling, which asserts an unobserved absence.
func TestInspectSkillMirror_SymlinkLoopEntryIsIndeterminate(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation needs privileges on windows")
	}
	root := t.TempDir()
	mirrorDir := filepath.Join(root, ".agents", "skills")
	if err := os.MkdirAll(mirrorDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, ".claude", "skills"), 0o755); err != nil {
		t.Fatal(err)
	}
	loopA := filepath.Join(mirrorDir, "loopA")
	loopB := filepath.Join(mirrorDir, "loopB")
	if err := os.Symlink(loopB, loopA); err != nil {
		t.Skipf("symlink unsupported here: %v", err)
	}
	if err := os.Symlink(loopA, loopB); err != nil {
		t.Skipf("symlink unsupported here: %v", err)
	}
	if _, err := os.Stat(loopA); err == nil || errors.Is(err, fs.ErrNotExist) {
		t.Skipf("symlink loop did not produce a non-ENOENT stat error: %v", err)
	}

	// The loop is TWO mirror entries (loopA and loopB are both symlinks whose
	// stat fails with ELOOP), so each folds into indeterminate separately.
	st := inspectSkillMirror(root)
	if st.indeterminate != 2 {
		t.Errorf("indeterminate = %d, want 2 (one per loop entry): %+v", st.indeterminate, st)
	}
	if len(st.dangling) != 0 {
		t.Errorf("loop entry leaked into dangling %v: %+v", st.dangling, st)
	}
}

// TestInspectSkillMirror_UnreadableCanonicalIsIndeterminateNotUnmirrored: a
// canonical skills path that exists but cannot be read as a directory (here:
// a regular file, ENOTDIR) folds into indeterminate — the mirror's state is
// unknown, never reported as unmirrored or absent.
func TestInspectSkillMirror_UnreadableCanonicalIsIndeterminateNotUnmirrored(t *testing.T) {
	root := t.TempDir()
	mirrorDir := filepath.Join(root, ".agents", "skills")
	if err := os.MkdirAll(mirrorDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// .claude/skills as a REGULAR FILE: present, but not a readable directory.
	claudeSkills := filepath.Join(root, ".claude")
	if err := os.MkdirAll(claudeSkills, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(claudeSkills, "skills"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mirrorDir, "entry.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	st := inspectSkillMirror(root)
	if !st.dirPresent {
		t.Fatalf("mirror dirPresent=false: %+v", st)
	}
	if st.indeterminate != 1 {
		t.Errorf("indeterminate = %d, want 1 (canonical read error): %+v", st.indeterminate, st)
	}
	if st.unmirrored != 0 {
		t.Errorf("canonical read error leaked into unmirrored: %+v", st)
	}
}

// TestInspectSkillMirror_AbsentCanonicalNotIndeterminate: an absent canonical
// skills directory is the expected "nothing canonical to compare against"
// state — it is NOT an indeterminate count, and the mirror entries observed
// so far stand.
func TestInspectSkillMirror_AbsentCanonicalNotIndeterminate(t *testing.T) {
	root := t.TempDir()
	mirrorDir := filepath.Join(root, ".agents", "skills")
	if err := os.MkdirAll(mirrorDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mirrorDir, "entry.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	// No .claude/ at all.

	st := inspectSkillMirror(root)
	if !st.dirPresent {
		t.Fatalf("mirror dirPresent=false: %+v", st)
	}
	if st.dirIndeterminate {
		t.Errorf("absent canonical set dirIndeterminate: %+v", st)
	}
	if st.indeterminate != 0 {
		t.Errorf("absent canonical counted as indeterminate: %+v", st)
	}
}

// TestInspectSkillMirror_StatSeamRecordsMirrorJoinedPath (site A): the stat
// path argument is the MIRROR-RELATIVE joined path — filepath.Join(mirrorDir,
// e.Name()) — for every symlink entry, resolving the producer's relative link
// body against the mirror directory exactly as the OS does for Codex (REQ-002).
func TestInspectSkillMirror_StatSeamRecordsMirrorJoinedPath(t *testing.T) {
	root := t.TempDir()
	writeSkillMirror(t, root, []string{"live"}, []mirrorEntrySpec{
		{name: "live", kind: mirrorEntryLive},
		{name: "ghost", kind: mirrorEntryDangling},
	})
	rec := stubStatRecording(t)

	st := inspectSkillMirror(root)
	if len(st.dangling) != 1 || st.dangling[0] != "ghost" {
		t.Fatalf("dangling = %v, want [ghost]: %+v", st.dangling, st)
	}
	mirrorDir := filepath.Join(root, ".agents", "skills")
	// ReadDir yields entries lexically: ghost, then live. Both are symlinks,
	// so both are stat'ed (the stat follows the link).
	wantPaths := []string{
		filepath.Join(mirrorDir, "ghost"),
		filepath.Join(mirrorDir, "live"),
	}
	if len(rec.paths) != len(wantPaths) {
		t.Fatalf("recorded stat calls = %v, want exactly %v", rec.paths, wantPaths)
	}
	for i, want := range wantPaths {
		if rec.paths[i] != want {
			t.Errorf("recorded stat arg[%d] = %q, want %q", i, rec.paths[i], want)
		}
	}
}

// TestInspectSkillMirror_InjectedResultsDriveBuckets: two mirror entries
// whose links BOTH resolve on disk are split across the dangling and
// indeterminate buckets purely by the injected stat results — ErrNotExist →
// dangling (named), any other error → indeterminate, and neither ever lands
// in the silent resolve arm.
func TestInspectSkillMirror_InjectedResultsDriveBuckets(t *testing.T) {
	root := t.TempDir()
	writeSkillMirror(t, root, []string{"t-a", "t-b"}, []mirrorEntrySpec{
		{name: "t-a", kind: mirrorEntryLive},
		{name: "t-b", kind: mirrorEntryLive},
	})
	mirrorDir := filepath.Join(root, ".agents", "skills")
	stubStatInject(t, map[string]error{
		filepath.Join(mirrorDir, "t-a"): fs.ErrNotExist,
		filepath.Join(mirrorDir, "t-b"): errors.New("t563: injected stat failure (not ENOENT)"),
	})

	st := inspectSkillMirror(root)
	if len(st.dangling) != 1 || st.dangling[0] != "t-a" {
		t.Errorf("dangling = %v, want [t-a] (injected ErrNotExist must drive dangling): %+v", st.dangling, st)
	}
	if st.indeterminate != 1 {
		t.Errorf("indeterminate = %d, want 1 (injected other error): %+v", st.indeterminate, st)
	}
}
