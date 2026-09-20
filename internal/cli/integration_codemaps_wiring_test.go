package cli

// integration_codemaps_wiring_test.go — the standing source's attachment to
// `moai integration release` (card t1018).
//
// Two things are guarded here and nothing else: that the release verb DOES
// run the standing source, and that the standing source can NEVER fail the
// release. Everything about which card text is issued lives in
// integration_codemaps_card_test.go — this file is about the seam.

import (
	"strings"
	"testing"
)

// TestReleaseRunsTheStandingSourceAndCannotFail drives the lane's round trip
// in a throwaway root with no codemaps artifact, so the layer comes back
// unjudgeable. That case is the interesting one twice over:
//
//   - the release must still succeed, because the window was already given
//     back before the standing source ran — a failure reported afterwards
//     would name a release that did happen as a failure; and
//   - the unjudgeability must be VISIBLE on the error stream, because a
//     permanently unmeasurable tree otherwise looks exactly like a batch that
//     keeps coming in clean.
//
// The visible line is also the positive control for the first assertion: an
// unwired standing source would let the release pass too, and only the line
// tells the two apart.
func TestReleaseRunsTheStandingSourceAndCannotFail(t *testing.T) {
	root := t.TempDir()

	if _, err := runIntegration(t, root, "acquire",
		"--session", "sess-lane8", "--name", "lane-8", "--branch", "release/v9.9.9"); err != nil {
		t.Fatalf("acquire: %v", err)
	}

	stdout, stderr, err := runIntegrationStreams(t, root, "release", "--session", "sess-lane8")
	if err != nil {
		t.Fatalf("the standing source must not be able to fail the release: %v\nstderr: %s", err, stderr)
	}
	if !strings.Contains(stdout, "window released") {
		t.Errorf("release confirmation missing from stdout: %q", stdout)
	}
	if !strings.Contains(stderr, "codemaps debt check") {
		t.Errorf("the standing source left no trace — either it is not wired into release, "+
			"or it swallowed its own failure silently; stderr: %q", stderr)
	}
	if strings.Contains(stdout, "codemaps debt card issued") {
		t.Errorf("an unmeasurable tree must not issue a card: %q", stdout)
	}
}

// TestReleaseStandingSourceKeepsJSONStdoutClean guards the `--json` path: a
// consumer parsing stdout must receive the release document and nothing else,
// so the standing source's own line goes to the error stream there.
func TestReleaseStandingSourceKeepsJSONStdoutClean(t *testing.T) {
	root := t.TempDir()

	if _, err := runIntegration(t, root, "acquire",
		"--session", "sess-lane9", "--name", "lane-9", "--branch", "release/v9.9.9"); err != nil {
		t.Fatalf("acquire: %v", err)
	}

	stdout, stderr, err := runIntegrationStreams(t, root, "release", "--session", "sess-lane9", "--json")
	if err != nil {
		t.Fatalf("release --json: %v\nstderr: %s", err, stderr)
	}
	if strings.Contains(stdout, "codemaps") {
		t.Errorf("standing-source output leaked into the JSON document: %q", stdout)
	}
	if !strings.Contains(stderr, "codemaps debt check") {
		t.Errorf("the standing source must still run on the --json path; stderr: %q", stderr)
	}
}
