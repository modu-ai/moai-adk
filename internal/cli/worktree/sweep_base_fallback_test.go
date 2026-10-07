package worktree

// sweep_base_fallback_test.go — SPEC-CUTOVER-RESIDUE-001 M1.
//
// The sweep derives its default base from the configured integration target;
// after the GitHub Flow cutover that target named a ref the remote no longer
// carries, and every tree's landing predicate failed with cause=fetch-failed
// (185 of 187 evaluated trees, measured on the pre-fix tree cb2a011d0). The
// behavior under test is the derived-base resolution: when the derived ref is
// absent on its remote, the sweep falls back to the remote's own default
// branch — resolved from the remote, never hardcoded — and keeps the derived
// base in every case it cannot answer affirmatively.

import (
	"errors"
	"testing"
)

// sweepFallbackMock installs canned answers for the two remote-probe seams
// and restores both when the test ends.
func sweepFallbackMock(t *testing.T, refExists bool, refExistsErr error, head string, headErr error) {
	t.Helper()
	origExists, origHead := sweepRemoteRefExists, sweepRemoteHead
	t.Cleanup(func() {
		sweepRemoteRefExists, sweepRemoteHead = origExists, origHead
	})
	sweepRemoteRefExists = func(_, _, _ string) (bool, error) {
		return refExists, refExistsErr
	}
	sweepRemoteHead = func(_, _ string) (string, error) {
		return head, headErr
	}
}

func TestSweepEffectiveBase(t *testing.T) {
	cases := []struct {
		name         string
		base         string
		refExists    bool
		refExistsErr error
		head         string
		headErr      error
		wantBase     string
		wantFallback bool
	}{
		{
			name:         "absent derived ref falls back to the remote default branch",
			base:         "origin/develop",
			refExists:    false,
			head:         "main",
			wantBase:     "origin/main",
			wantFallback: true,
		},
		{
			name:         "present derived ref is used unchanged and HEAD is not consulted",
			base:         "origin/develop",
			refExists:    true,
			head:         "main",
			wantBase:     "origin/develop",
			wantFallback: false,
		},
		{
			name:         "undeterminable probe keeps the derived base",
			base:         "origin/develop",
			refExistsErr: errors.New("ls-remote failed: network down"),
			head:         "main",
			wantBase:     "origin/develop",
			wantFallback: false,
		},
		{
			name:         "unresolvable remote HEAD keeps the derived base",
			base:         "origin/develop",
			refExists:    false,
			headErr:      errors.New("ls-remote --symref returned no symbolic ref line"),
			wantBase:     "origin/develop",
			wantFallback: false,
		},
		{
			name:         "resolved head equal to the absent ref keeps the base",
			base:         "origin/develop",
			refExists:    false,
			head:         "develop",
			wantBase:     "origin/develop",
			wantFallback: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sweepFallbackMock(t, tc.refExists, tc.refExistsErr, tc.head, tc.headErr)
			gotBase, gotFallback := sweepEffectiveBase("/repo", tc.base)
			if gotBase != tc.wantBase {
				t.Fatalf("effective base = %q, want %q", gotBase, tc.wantBase)
			}
			if gotFallback != tc.wantFallback {
				t.Fatalf("fallback = %v, want %v", gotFallback, tc.wantFallback)
			}
		})
	}
}
