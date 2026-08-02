package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// freshNoticeMarker is the stable substring the notice-count assertions key on.
// Counting occurrences (rather than asserting presence) is what makes a double
// emission detectable: EnsureDir is called from two places after the M3
// reorder, and a notice attached to "wherever EnsureDir runs" would fire twice.
const freshNoticeMarker = "has no Claude Code configuration yet"

func fnSeedClaudeConfig(t *testing.T, profileDir string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(profileDir, ".claude.json"), []byte("{}"), 0o644); err != nil {
		t.Fatalf("seed .claude.json: %v", err)
	}
}

// fnCaptureStderr swaps the launcher stderr seam for a buffer.
func fnCaptureStderr(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	orig := launcherStderr
	t.Cleanup(func() { launcherStderr = orig })
	launcherStderr = &buf
	return &buf
}

// --- AC-PM-011, notice half (REQ-PM-016, 017) ---

// TestFreshProfileNotice_WriterContent judges the notice function in isolation:
// it is self-gating, so calling it on a populated profile writes nothing.
//
// This test CANNOT detect a double emission — it calls the function once, so it
// sees one output by construction. TestUnifiedLaunch_FreshProfileNoticeEmittedExactlyOnce
// owns the firing-count question at the launch level.
func TestFreshProfileNotice_WriterContent(t *testing.T) {
	base, _ := lpSandbox(t)
	lpMkProfile(t, base, "fresh")
	populatedDir := lpMkProfile(t, base, "populated")
	fnSeedClaudeConfig(t, populatedDir)

	var fresh bytes.Buffer
	warnFreshProfile(&fresh, "fresh")
	if !strings.Contains(fresh.String(), freshNoticeMarker) {
		t.Errorf("fresh profile produced no notice; got:\n%s", fresh.String())
	}
	if !strings.Contains(fresh.String(), "fresh") {
		t.Errorf("notice does not name the profile; got:\n%s", fresh.String())
	}

	var populated bytes.Buffer
	warnFreshProfile(&populated, "populated")
	if populated.Len() != 0 {
		t.Errorf("populated profile produced output, want none; got:\n%s", populated.String())
	}
}

// --- AC-PM-018 (REQ-PM-016, launch layer) ---

// TestUnifiedLaunch_FreshProfileNoticeEmittedExactlyOnce judges the firing
// COUNT over one whole launch, across both the explicit -p path and the bare
// path that resolves through the projects map.
//
// Case B is the load-bearing one. Gating step 4.5 on what the user typed rather
// than on the resolved name yields a count of 0 exactly there — the silent
// login screen this SPEC removes — and case A cannot see it, because on the
// explicit path the two variables hold the same value.
func TestUnifiedLaunch_FreshProfileNoticeEmittedExactlyOnce(t *testing.T) {
	t.Run("A_explicit_p_fresh", func(t *testing.T) {
		base, _ := lpSandbox(t)
		lpMkProfile(t, base, "fresh")
		buf := fnCaptureStderr(t)
		lpStubLaunch(t)

		if err := unifiedLaunch("fresh", "claude", nil); err != nil {
			t.Fatalf("unifiedLaunch: %v", err)
		}
		if got := strings.Count(buf.String(), freshNoticeMarker); got != 1 {
			t.Errorf("notice fired %d times, want exactly 1; output:\n%s", got, buf.String())
		}
	})

	t.Run("B_bare_launch_resolves_to_fresh", func(t *testing.T) {
		base, root := lpSandbox(t)
		lpMkProfile(t, base, "fresh")
		lpWriteLedger(t, base, "projects:\n  "+lpProjectKey(t, root)+": fresh\n")
		buf := fnCaptureStderr(t)
		_, gotProfile := lpStubLaunch(t)

		if err := unifiedLaunch("", "claude", nil); err != nil {
			t.Fatalf("unifiedLaunch: %v", err)
		}
		if *gotProfile != "fresh" {
			t.Fatalf("precondition: launch resolved to %q, want fresh", *gotProfile)
		}
		if got := strings.Count(buf.String(), freshNoticeMarker); got != 1 {
			t.Errorf("notice fired %d times on the bare-resolution path, want exactly 1; "+
				"0 means step 4.5 is gated on what the user typed instead of the resolved "+
				"profile. Output:\n%s", got, buf.String())
		}
	})

	t.Run("C1_explicit_p_populated", func(t *testing.T) {
		base, _ := lpSandbox(t)
		dir := lpMkProfile(t, base, "populated")
		fnSeedClaudeConfig(t, dir)
		buf := fnCaptureStderr(t)
		lpStubLaunch(t)

		if err := unifiedLaunch("populated", "claude", nil); err != nil {
			t.Fatalf("unifiedLaunch: %v", err)
		}
		if got := strings.Count(buf.String(), freshNoticeMarker); got != 0 {
			t.Errorf("notice fired %d times for a populated profile, want 0; output:\n%s", got, buf.String())
		}
	})

	t.Run("C2_bare_launch_resolves_to_populated", func(t *testing.T) {
		base, root := lpSandbox(t)
		dir := lpMkProfile(t, base, "populated")
		fnSeedClaudeConfig(t, dir)
		lpWriteLedger(t, base, "projects:\n  "+lpProjectKey(t, root)+": populated\n")
		buf := fnCaptureStderr(t)
		_, gotProfile := lpStubLaunch(t)

		if err := unifiedLaunch("", "claude", nil); err != nil {
			t.Fatalf("unifiedLaunch: %v", err)
		}
		if *gotProfile != "populated" {
			t.Fatalf("precondition: launch resolved to %q, want populated", *gotProfile)
		}
		if got := strings.Count(buf.String(), freshNoticeMarker); got != 0 {
			t.Errorf("notice fired %d times on the bare-resolution path for a populated "+
				"profile, want 0 — the widened gate must not over-fire. Output:\n%s", got, buf.String())
		}
	})
}
