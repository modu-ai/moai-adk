package cli

import (
	"context"
	"strings"
	"testing"
)

// withStubCodex installs BOTH the single-shot runner seam (for codex_setup /
// auth probes) and the session seam (for the review-gate JSON-RPC session), so a
// test drives whichever path the code under exercise takes. The session seam
// replays `lines` as the codex stdout NDJSON transcript and records every sent
// request line into `sent`.
func withStubCodex(t *testing.T, lines []string) *fakeCodexSession {
	t.Helper()
	prevRunner, prevLook, prevSess := codexRunner, codexLookPath, codexSession
	codexRunner = stubCodexRunner{}
	codexLookPath = func(string) (string, error) { return "/usr/bin/true", nil }
	sess := &fakeCodexSession{lines: lines}
	codexSession = sess
	t.Cleanup(func() { codexRunner, codexLookPath, codexSession = prevRunner, prevLook, prevSess })
	return sess
}

// stubCodexRunner is the no-op single-shot runner for tests that exercise the
// session path (it must not be called).
type stubCodexRunner struct{}

func (stubCodexRunner) run(context.Context, string, []string, string) (string, error) {
	return "", nil
}

// A JSON-RPC rejection observed verbatim from codex-cli 0.146.1 when the client
// sends `target` as a bare string instead of the tagged object the app-server
// protocol requires. Before the error arm existed this decoded into a
// zero-valued Result and read as "the reviewer had no opinion".
//
// Under the corrected session protocol the rejection can no longer come from
// review/start's target (the client now sends an object), but the SAME server
// rejection shape can still fire from any handshake step — so the canned line is
// replayed as the initialize (id=1) response to pin that ANY JSON-RPC error arm
// is surfaced with the server's own code and message.
const codexProtocolRejection = `{"error":{"code":-32600,"message":"Invalid request: invalid type: string \"uncommittedChanges\", expected internally tagged enum ReviewTarget"},"id":1}`

// TestRunCodexReviewRPC_SurfacesServerError pins that a JSON-RPC error arm is
// reported with the server's own code and message, rather than collapsing into
// the generic no-verdict path. The rejection is replayed as the initialize
// handshake response, so it is the FIRST thing the session client sees.
func TestRunCodexReviewRPC_SurfacesServerError(t *testing.T) {
	withStubCodex(t, []string{codexProtocolRejection})

	out, err := runCodexReviewRPC(context.Background(), "/usr/bin/true", codexMethodReviewStart, map[string]any{})
	if err == nil {
		t.Fatal("a JSON-RPC error arm must return a non-nil error")
	}
	for _, want := range []string{"-32600", "expected internally tagged enum ReviewTarget"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error must carry %q verbatim; got %q", want, err.Error())
		}
	}
	if out.Verdict != VerdictInconclusive {
		t.Errorf("verdict = %q, want %q (fail-open shape preserved)", out.Verdict, VerdictInconclusive)
	}
	if strings.Contains(out.Summary, "carried no verdict") {
		t.Error("a rejected request must not be described as a verdict-less review")
	}
}

// TestProduceCodexReviewReceipt_RPCErrorRecordsInconclusive pins the reviewer
// error contract at its M2 home, the receipt producer (the gate no longer
// calls the RPC in-hook — REQ-GBN-002): a protocol-rejected review records
// inconclusive with the non-zero exit — fail-open holds, and the record never
// reads as a local pass. Swallowing the error made a pipeline that could not
// reach a verdict look like one that had found nothing wrong.
func TestProduceCodexReviewReceipt_RPCErrorRecordsInconclusive(t *testing.T) {
	withStubCodex(t, []string{codexProtocolRejection})
	root := cacheTestRoot(t)

	r, err := produceCodexReviewReceipt(context.Background(), root)
	if err != nil {
		t.Fatalf("a reviewer error must not fail the producer (fail-open): %v", err)
	}
	if r.Verdict != codexReviewVerdictInconclusive || r.ExitCode != 2 {
		t.Fatalf("a protocol-rejected review must record inconclusive/2, got %q/%d", r.Verdict, r.ExitCode)
	}
}
