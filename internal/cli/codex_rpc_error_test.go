package cli

import (
	"context"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/hook"
)

// stubCodexRunner returns a canned stdout for the codex app-server call.
type stubCodexRunner struct{ stdout string }

func (s stubCodexRunner) run(context.Context, string, []string, string) (string, error) {
	return s.stdout, nil
}

func withStubCodex(t *testing.T, stdout string) {
	t.Helper()
	prevRunner, prevLook := codexRunner, codexLookPath
	codexRunner = stubCodexRunner{stdout: stdout}
	codexLookPath = func(string) (string, error) { return "/usr/bin/true", nil }
	t.Cleanup(func() { codexRunner, codexLookPath = prevRunner, prevLook })
}

// A JSON-RPC rejection observed verbatim from codex-cli 0.146.1 when the client
// sends `target` as a bare string instead of the tagged object the app-server
// protocol requires. Before the error arm existed this decoded into a
// zero-valued Result and read as "the reviewer had no opinion".
const codexProtocolRejection = `{"error":{"code":-32600,"message":"Invalid request: invalid type: string \"uncommittedChanges\", expected internally tagged enum ReviewTarget"},"id":1}`

// TestRunCodexReviewRPC_SurfacesServerError pins that a JSON-RPC error arm is
// reported with the server's own code and message, rather than collapsing into
// the generic no-verdict path.
func TestRunCodexReviewRPC_SurfacesServerError(t *testing.T) {
	withStubCodex(t, codexProtocolRejection)

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

// TestHandleCodexReviewGate_ReturnsRPCErrorWithAllow pins that the gate keeps
// ALLOWing on a reviewer error (fail-open is unchanged) while handing the cause
// back to the caller, which logs it. Swallowing the error made a gate that
// could not reach a verdict look like a gate that had found nothing wrong.
func TestHandleCodexReviewGate_ReturnsRPCErrorWithAllow(t *testing.T) {
	withStubCodex(t, codexProtocolRejection)
	prevDetector := reviewGateChangeDetector
	reviewGateChangeDetector = func(string) bool { return true } // reviewable change present
	t.Cleanup(func() { reviewGateChangeDetector = prevDetector })

	out, err := HandleCodexReviewGate(&hook.HookInput{}, true /* enabled */, t.TempDir())
	if err == nil {
		t.Fatal("the reviewer error must reach the caller so it can be logged")
	}
	if out == nil || out.Decision == hook.DecisionBlock {
		t.Errorf("fail-open must hold: decision = %+v, want ALLOW", out)
	}
}
