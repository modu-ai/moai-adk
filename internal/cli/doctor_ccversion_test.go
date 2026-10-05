package cli

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/session"
)

// saveDoctorCCVersionSeams captures the doctor check's two seams and restores
// them at cleanup.
func saveDoctorCCVersionSeams(t *testing.T) {
	t.Helper()
	prevEntries := doctorCCVersionEntries
	prevResolve := session.ResolveCCVersions
	t.Cleanup(func() {
		doctorCCVersionEntries = prevEntries
		session.ResolveCCVersions = prevResolve
	})
}

// doctorCCFixtureEntry is one registry row the staleness check reads.
func doctorCCFixtureEntry(sessionID string, pid int) session.Entry {
	return session.Entry{SessionID: sessionID, SpecID: "SPEC-CC-VERSION-FIXTURE", PID: pid}
}

// fixtureResolve returns a session.ResolveCCVersions substitution keyed on
// the fixture pid.
func fixtureResolve(running, installed string) func(int) session.CCVersions {
	return func(pid int) session.CCVersions {
		if pid != 4242 {
			return session.CCVersions{Running: session.UnknownCCVersion, Installed: session.UnknownCCVersion}
		}
		return session.CCVersions{Running: running, Installed: installed}
	}
}

// TestDoctorCCVersionStaleness (AC-SCV-007, REQ-SCV-007) — the staleness
// check warns naming both versions when a live session runs older, reads
// CheckOK when equal, reads CheckOK with the unknown noted when either side
// is unknown (never a warn on unknown), and never returns CheckFail in any
// fixture — the advisory property pinned at the unit layer.
func TestDoctorCCVersionStaleness(t *testing.T) {
	subtests := []struct {
		name       string
		entries    []session.Entry
		resolve    func(int) session.CCVersions
		wantStatus uikit.CheckStatus
		contains   []string
	}{
		{
			name:       "behind warns naming both versions",
			entries:    []session.Entry{doctorCCFixtureEntry("uuid-behind-1", 4242)},
			resolve:    fixtureResolve("2.1.281", "2.1.288"),
			wantStatus: uikit.CheckWarn,
			contains:   []string{"2.1.281", "2.1.288"},
		},
		{
			name:       "equal versions reads OK",
			entries:    []session.Entry{doctorCCFixtureEntry("uuid-equal-1", 4242)},
			resolve:    fixtureResolve("2.1.288", "2.1.288"),
			wantStatus: uikit.CheckOK,
		},
		{
			name:       "unknown running reads OK with the unknown noted",
			entries:    []session.Entry{doctorCCFixtureEntry("uuid-unkrun-1", 4242)},
			resolve:    fixtureResolve(session.UnknownCCVersion, "2.1.288"),
			wantStatus: uikit.CheckOK,
			contains:   []string{session.UnknownCCVersion},
		},
		{
			name:       "unknown installed reads OK with the unknown noted",
			entries:    []session.Entry{doctorCCFixtureEntry("uuid-unkinst-1", 4242)},
			resolve:    fixtureResolve("2.1.281", session.UnknownCCVersion),
			wantStatus: uikit.CheckOK,
			contains:   []string{session.UnknownCCVersion},
		},
		{
			name:       "empty registry reads OK",
			entries:    []session.Entry{},
			resolve:    fixtureResolve("2.1.281", "2.1.288"),
			wantStatus: uikit.CheckOK,
		},
		{
			name:       "numeric compare not lexicographic",
			entries:    []session.Entry{doctorCCFixtureEntry("uuid-numeric-1", 4242)},
			resolve:    fixtureResolve("2.1.9", "2.1.288"),
			wantStatus: uikit.CheckWarn,
			contains:   []string{"2.1.9", "2.1.288"},
		},
	}
	for _, tc := range subtests {
		t.Run(tc.name, func(t *testing.T) {
			saveDoctorCCVersionSeams(t)
			doctorCCVersionEntries = func() ([]session.Entry, error) { return tc.entries, nil }
			session.ResolveCCVersions = tc.resolve

			check := checkSessionCCVersionStaleness("", false)
			if check.Status != tc.wantStatus {
				t.Fatalf("status = %v, want %v (message: %s)", check.Status, tc.wantStatus, check.Message)
			}
			if check.Status == uikit.CheckFail {
				t.Fatalf("the advisory check returned CheckFail — it must never gate doctor")
			}
			for _, want := range tc.contains {
				if !strings.Contains(check.Message, want) && !strings.Contains(check.Detail, want) {
					t.Errorf("output missing %q (message: %s; detail: %s)", want, check.Message, check.Detail)
				}
			}
		})
	}

	// The warn message itself (not only the detail) names both versions —
	// REQ-SCV-007's "warn naming both versions" is about the warning the
	// operator reads, so the single-behind case asserts the message field.
	t.Run("warn message names both versions", func(t *testing.T) {
		saveDoctorCCVersionSeams(t)
		doctorCCVersionEntries = func() ([]session.Entry, error) {
			return []session.Entry{doctorCCFixtureEntry("uuid-msg-1", 4242)}, nil
		}
		session.ResolveCCVersions = fixtureResolve("2.1.281", "2.1.288")

		check := checkSessionCCVersionStaleness("", false)
		if check.Status != uikit.CheckWarn {
			t.Fatalf("status = %v, want CheckWarn", check.Status)
		}
		for _, want := range []string{"2.1.281", "2.1.288"} {
			if !strings.Contains(check.Message, want) {
				t.Errorf("warn message missing %q: %s", want, check.Message)
			}
		}
	})
}
