package cli

// SPEC-DUAL-HARNESS-HOOK-PARITY-001 M2d, R1: `moai verify codex-review` is the
// out-of-hook codex review runner. It makes the same review call the Claude
// gate makes in-hook (HandleCodexReviewGate) and records the verdict as a
// receipt bound to the current tree, the codex version, and the review
// configuration.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/verify"
)

func runVerifyCodexReview(t *testing.T, root string) (map[string]any, error) {
	t.Helper()
	cmd := newVerifyCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"codex-review", "--project-root", root})
	if err := cmd.Execute(); err != nil {
		return nil, err
	}
	var got map[string]any
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("output is not JSON: %q", out.String())
	}
	return got, nil
}

func TestVerifyCodexReviewRecordsReceipt(t *testing.T) {
	ctx := context.Background()
	fakeCodexVersion(t, "codex-cli 0.0.0-receipt")
	f := newStopFixture(t)
	f.dirty(t, "reviewable")

	current := func(t *testing.T) *verify.Receipt {
		t.Helper()
		key, err := verify.Key(ctx, f.root)
		if err != nil {
			t.Fatal(err)
		}
		state := codexReviewReceiptState(ctx, key, "/fake/codex")
		return verify.LoadReceipt(f.root, state)
	}

	cases := []struct {
		name   string
		review []string
		start  error
		want   string
	}{
		{"review finds issues", codexSessionScript("- [P1] found issues\n- [P2] more"), nil, "fail"},
		{"review passes", codexSessionScript("clean change, approved"), nil, "pass"},
		{"review call errors", nil, errors.New("fake: session start failed"), "inconclusive"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sess := withCodexSession(t, tc.review)
			sess.startErr = tc.start
			got, err := runVerifyCodexReview(t, f.root)
			if err != nil {
				t.Fatalf("moai verify codex-review: %v", err)
			}
			if got["verdict"] != tc.want {
				t.Fatalf("verdict = %v, want %s (%v)", got["verdict"], tc.want, got)
			}
			r := current(t)
			if r == nil || r.Verdict != tc.want {
				t.Fatalf("stored receipt = %+v, want verdict %s", r, tc.want)
			}
			if r.ToolVersion != "codex-cli 0.0.0-receipt" {
				t.Fatalf("tool_version = %q, want the codex --version output", r.ToolVersion)
			}
		})
	}

	t.Run("codex binary missing records nothing", func(t *testing.T) {
		g := newStopFixture(t)
		g.dirty(t, "reviewable")
		withCodexLookPath(t, func(string) (string, error) { return "", errFakeLookPath })
		if _, err := runVerifyCodexReview(t, g.root); err == nil {
			t.Fatal("a missing codex binary must fail the runner, not record a receipt")
		}
		key, err := verify.Key(ctx, g.root)
		if err != nil {
			t.Fatal(err)
		}
		if r := verify.LoadReceipt(g.root, codexReviewReceiptState(ctx, key, "/fake/codex")); r != nil {
			t.Fatalf("a receipt was recorded without a reviewer: %+v", r)
		}
	})
}
