package cli

// feedback_participation_model.go — the production Summarizer over the
// existing headless claude runner (SPEC-FEEDBACK-PARTICIPATION-001 design
// section 8, DEC-6): the flag set claudeAuditArgs passes today MINUS the
// audit --json-schema, a fixed prompt over the payload's closed fields
// (publish.SummaryPrompt), and a bounded run through runClaudeCommand.
// The consent text names that the summary may spend the user's own
// subscription tokens; the deterministic template text covers every
// unavailable, unauthenticated, or failed path.

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/bugreport"
	"github.com/modu-ai/moai-adk/internal/feedback/publish"
)

// claudeParticipationSummarizer is injected into publish at the flush call
// (runParticipationFlushWork) — publish imports no model helper and no
// internal/cli (REQ-ANON-025, AC-025).
type claudeParticipationSummarizer struct{}

// claudeSummaryArgs is claudeAuditArgs minus the audit --json-schema pair:
// the same headless, tool-free, prompt-only shape.
func claudeSummaryArgs(model, effort string) []string {
	audit := claudeAuditArgs(model, effort)
	args := make([]string, 0, len(audit))
	for i := 0; i < len(audit); i++ {
		if audit[i] == "--json-schema" {
			i++ // skip its value too
			continue
		}
		args = append(args, audit[i])
	}
	return args
}

// claudeParticipationEnv builds the summarizer's subprocess environment:
// the CALLER'S real environment through the audit scrubber — the base env
// (HOME, PATH, the auth store's lookup path) survives, the risky keys
// (ANTHROPIC_*, CLAUDECODE_*, the audit scrub set) are dropped (review-gate
// finding 6: the nil-environ construction built an EMPTY env, leaving the
// claude child without HOME or PATH). The proxy family the audit env
// validator rejects is dropped too (l57-r2 finding 3): the audit path's
// allowlist never lets a proxy through, but this denylist path kept one, and
// Summarize then died at its own validation before summarizing. The
// validator's wording is mirrored exactly (NO_PROXY, any *_PROXY suffix).
// CLAUDE_CONFIG_DIR is then pinned to the subscription directory, as the
// audit env validator requires (SPEC-MCP-NEUTRALITY-001 REQ-MN-012): an
// ambient value is replaced, never inherited.
func claudeParticipationEnv() []string {
	scrubbed := scrubClaudeAuditEnv(os.Environ())
	env := make([]string, 0, len(scrubbed)+1)
	for _, entry := range scrubbed {
		key, _, _ := strings.Cut(entry, "=")
		if key == "NO_PROXY" || strings.HasSuffix(key, "_PROXY") {
			continue
		}
		if key != "CLAUDE_CONFIG_DIR" {
			env = append(env, entry)
		}
	}
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		env = append(env, "CLAUDE_CONFIG_DIR="+filepath.Join(home, ".claude"))
	}
	return env
}

func (claudeParticipationSummarizer) Summarize(ctx context.Context, p bugreport.Payload) (string, error) {
	binary, err := claudeLookPath(claudeBinaryName)
	if err != nil {
		return "", err
	}
	env := claudeParticipationEnv()
	if err := validateClaudeAuditEnv(env); err != nil {
		return "", err
	}
	authBytes, _, authErr := claudeRunner.RunAuthStatus(ctx, binary, env)
	if authErr != nil {
		return "", authErr
	}
	auth, err := parseClaudeAuthStatus(authBytes)
	if err != nil || !auth.subscriptionReady() {
		// An unauthenticated session: the caller falls back to the
		// deterministic template text (design section 8).
		return "", context.Canceled
	}
	prompt := publish.SummaryPrompt(p)
	if len(prompt) > publish.ModelInputMaxBytes() {
		return "", context.Canceled
	}
	stdout, _, runErr := runClaudeCommand(ctx, binary, "", claudeSummaryArgs(claudeAuditDefaultModel, claudeAuditDefaultEffort), env, []byte(prompt))
	if runErr != nil {
		return "", runErr
	}
	return parseClaudeSummaryOutput(stdout)
}

// parseClaudeSummaryOutput extracts the text from the --output-format json
// envelope, falling back to the raw trimmed output.
func parseClaudeSummaryOutput(stdout []byte) (string, error) {
	var envelope struct {
		Result string `json:"result"`
	}
	if err := json.Unmarshal(stdout, &envelope); err == nil && strings.TrimSpace(envelope.Result) != "" {
		return envelope.Result, nil
	}
	if trimmed := strings.TrimSpace(string(stdout)); trimmed != "" {
		return trimmed, nil
	}
	return "", context.Canceled
}
