package hygiene

import (
	"strings"
	"testing"
)

// TestLivenessVerdictMatrix — AC-HYG-006 (L-006). The full signal-
// combination table: any affirmative → LIVE; all-measurable-all-negative
// → DEAD; any unmeasured with none affirmative → INDETERMINATE. The
// registry-absent, entry-missing, and transcript-absent rows can never
// produce DEAD.
func TestLivenessVerdictMatrix(t *testing.T) {
	const key = "01234567-89ab-cdef-0123-456789abcdef"

	yes := func(Signal) {}
	_ = yes

	base := func() *Liveness {
		return &Liveness{
			TranscriptWindow: testActivityWindow,
			HeartbeatWindow:  testStaleHb,
		}
	}

	tests := []struct {
		name string
		seed func(*Liveness)
		want Verdict
		// neverDead asserts the row can never classify as DEAD even when
		// every OTHER signal reads negative: the listed probes flip to
		// negative while the row's structurally unmeasured source stays
		// untouched (registry absence, missing entry, absent transcript,
		// or a missing fingerprint cannot be measured into existence).
		neverDead bool
		flip      []string
	}{
		{
			name: "pid alive with matching fingerprint",
			seed: func(l *Liveness) {
				l.registryEntry = func(string) *RegistryEntry { return &RegistryEntry{PID: 4242} }
				l.pidProbe = func(int) Signal { return SignalAffirmative }
			},
			want: VerdictLive,
		},
		{
			name: "pid alive but fingerprint mismatch (reused pid)",
			seed: func(l *Liveness) {
				l.registryEntry = func(string) *RegistryEntry { return &RegistryEntry{PID: 4242} }
				l.pidProbe = func(int) Signal { return SignalNegative }
				l.transcriptProbe = func(string) Signal { return SignalNegative }
				l.heartbeatProbe = func(string) Signal { return SignalNegative }
			},
			want: VerdictDead,
		},
		{
			name: "pid dead",
			seed: func(l *Liveness) {
				l.registryEntry = func(string) *RegistryEntry { return &RegistryEntry{PID: 4242} }
				l.pidProbe = func(int) Signal { return SignalNegative }
				l.transcriptProbe = func(string) Signal { return SignalNegative }
				l.heartbeatProbe = func(string) Signal { return SignalNegative }
			},
			want: VerdictDead,
		},
		{
			name: "pid alive with no recorded fingerprint",
			seed: func(l *Liveness) {
				l.registryEntry = func(string) *RegistryEntry { return &RegistryEntry{PID: 4242} }
				l.pidProbe = func(int) Signal { return SignalUnmeasured }
				l.transcriptProbe = func(string) Signal { return SignalNegative }
				l.heartbeatProbe = func(string) Signal { return SignalNegative }
			},
			want:      VerdictIndeterminate,
			neverDead: true,
			flip:      []string{"transcript", "heartbeat"},
		},
		{
			name: "registry file absent",
			seed: func(l *Liveness) {
				l.registryEntry = func(string) *RegistryEntry { return nil }
				l.registryPresent = false
				l.transcriptProbe = func(string) Signal { return SignalNegative }
			},
			want:      VerdictIndeterminate,
			neverDead: true,
			flip:      []string{"transcript"},
		},
		{
			name: "registry present but no entry for the key",
			seed: func(l *Liveness) {
				l.registryEntry = func(string) *RegistryEntry { return nil }
				l.registryPresent = true
				l.transcriptProbe = func(string) Signal { return SignalNegative }
			},
			want:      VerdictIndeterminate,
			neverDead: true,
			flip:      []string{"transcript"},
		},
		{
			name: "probe undetermined",
			seed: func(l *Liveness) {
				l.registryEntry = func(string) *RegistryEntry { return &RegistryEntry{PID: 4242} }
				l.pidProbe = func(int) Signal { return SignalUnmeasured }
				l.transcriptProbe = func(string) Signal { return SignalNegative }
				l.heartbeatProbe = func(string) Signal { return SignalNegative }
			},
			want:      VerdictIndeterminate,
			neverDead: true,
			flip:      []string{"transcript", "heartbeat"},
		},
		{
			name: "transcript fresh",
			seed: func(l *Liveness) {
				l.transcriptProbe = func(string) Signal { return SignalAffirmative }
			},
			want: VerdictLive,
		},
		{
			name: "transcript stale",
			seed: func(l *Liveness) {
				l.registryEntry = func(string) *RegistryEntry { return &RegistryEntry{PID: 4242} }
				l.pidProbe = func(int) Signal { return SignalNegative }
				l.transcriptProbe = func(string) Signal { return SignalNegative }
				l.heartbeatProbe = func(string) Signal { return SignalNegative }
			},
			want: VerdictDead,
		},
		{
			name: "transcript absent under a resolvable root",
			seed: func(l *Liveness) {
				l.registryEntry = func(string) *RegistryEntry { return &RegistryEntry{PID: 4242} }
				l.pidProbe = func(int) Signal { return SignalNegative }
				l.transcriptProbe = func(string) Signal { return SignalUnmeasured }
				l.heartbeatProbe = func(string) Signal { return SignalNegative }
			},
			want:      VerdictIndeterminate,
			neverDead: true,
			flip:      []string{"pid", "heartbeat"},
		},
		{
			name: "transcript root unresolvable",
			seed: func(l *Liveness) {
				l.registryEntry = func(string) *RegistryEntry { return &RegistryEntry{PID: 4242} }
				l.pidProbe = func(int) Signal { return SignalNegative }
				l.transcriptProbe = func(string) Signal { return SignalUnmeasured }
				l.heartbeatProbe = func(string) Signal { return SignalNegative }
			},
			want:      VerdictIndeterminate,
			neverDead: true,
			flip:      []string{"pid", "heartbeat"},
		},
		{
			name: "heartbeat younger than the staleness window",
			seed: func(l *Liveness) {
				l.registryEntry = func(string) *RegistryEntry { return &RegistryEntry{PID: 4242} }
				l.pidProbe = func(int) Signal { return SignalNegative }
				l.transcriptProbe = func(string) Signal { return SignalNegative }
				l.heartbeatProbe = func(string) Signal { return SignalAffirmative }
			},
			want: VerdictLive,
		},
		{
			name: "heartbeat older than the staleness window",
			seed: func(l *Liveness) {
				l.registryEntry = func(string) *RegistryEntry { return &RegistryEntry{PID: 4242} }
				l.pidProbe = func(int) Signal { return SignalNegative }
				l.transcriptProbe = func(string) Signal { return SignalNegative }
				l.heartbeatProbe = func(string) Signal { return SignalNegative }
			},
			want: VerdictDead,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			l := base()
			tc.seed(l)
			got, unmeasured, evidence := l.Evaluate(key)
			if got != tc.want {
				t.Fatalf("verdict = %s, want %s (unmeasured=%v evidence=%v)",
					got, tc.want, unmeasured, evidence)
			}
			if tc.neverDead {
				for _, probe := range tc.flip {
					switch probe {
					case "pid":
						l.pidProbe = func(int) Signal { return SignalNegative }
					case "transcript":
						l.transcriptProbe = func(string) Signal { return SignalNegative }
					case "heartbeat":
						l.heartbeatProbe = func(string) Signal { return SignalNegative }
					}
				}
				got2, _, _ := l.Evaluate(key)
				if got2 == VerdictDead {
					t.Fatalf("unmeasured-signal row classified DEAD on an all-negative field")
				}
			}
		})
	}
}

// TestLivenessNeverDeadOnRegistryAbsence — the majority-case rows
// (61 registry entries vs 1,116 files measured) re-asserted standalone.
func TestLivenessNeverDeadOnRegistryAbsence(t *testing.T) {
	l := &Liveness{
		TranscriptWindow: testActivityWindow,
		HeartbeatWindow:  testStaleHb,
		registryPresent:  false,
		transcriptProbe:  func(string) Signal { return SignalNegative },
	}
	got, unmeasured, _ := l.Evaluate("99999999-aaaa-bbbb-cccc-dddddddddddd")
	if got == VerdictDead {
		t.Fatalf("registry-absent candidate classified DEAD")
	}
	if got != VerdictIndeterminate {
		t.Fatalf("verdict = %s, want indeterminate", got)
	}
	joined := strings.Join(unmeasured, ",")
	if !strings.Contains(joined, "pid") || !strings.Contains(joined, "heartbeat") {
		t.Fatalf("unmeasured names missing pid/heartbeat: %v", unmeasured)
	}
}
