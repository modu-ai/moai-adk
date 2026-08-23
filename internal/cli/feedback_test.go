package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/feedback"
)

// Credential-shaped fixtures are assembled at run time from a prefix constant
// plus a deterministic dummy tail, mirroring internal/feedback's test
// convention: the repository's PreToolUse guard rejects committed
// credential-looking literals, and such a fixture would itself be a leak
// surface.
const cliGHTokenPrefix = "ghp_"

func cliDummyTail(n int) string {
	const alphabet = "abcdefghij0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteByte(alphabet[i%len(alphabet)])
	}
	return b.String()
}

func cliFakeGitHubToken() string { return cliGHTokenPrefix + cliDummyTail(36) }

// newFeedbackProject creates a temp directory carrying the .moai marker, so
// ResolveProjectRoot recognises it as a project root.
func newFeedbackProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, defs.MoAIDir), 0o755); err != nil {
		t.Fatalf("create project marker: %v", err)
	}
	return root
}

// writeFeedbackSecuritySection writes the project's security section, the file
// the scrubber reads for pattern and environment-name extensions.
func writeFeedbackSecuritySection(t *testing.T, root, content string) {
	t.Helper()
	dir := filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("create sections dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, feedbackSecurityFileName), []byte(content), 0o600); err != nil {
		t.Fatalf("write security section: %v", err)
	}
}

// runFeedbackCLI executes the `feedback` command tree with the given stdin and
// argv, returning stdout, stderr and the RunE error. A non-nil error is what
// the process turns into a non-zero exit code.
func runFeedbackCLI(t *testing.T, stdin string, args ...string) (string, string, error) {
	t.Helper()
	cmd := newFeedbackCmd()
	var out, errOut bytes.Buffer
	cmd.SetIn(strings.NewReader(stdin))
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), errOut.String(), err
}

// TestFeedbackScrubContract is AC-F-003: stdout carries a single JSON object
// with the five contract fields (title among them — the structural evidence
// that the title no longer bypasses the scrubber), every finding names where it
// matched, and the exit code is zero.
//
// The fixture masks an env value rather than a credential, because a
// credential in the raw text is itself a blocking signal: the contract under
// test is the shape of a passing scrub, so it needs findings AND verdict ok.
// Carrying the value through security.sandbox.env_scrub_extra also observes
// that the CLI reads the section it claims to read.
func TestFeedbackScrubContract(t *testing.T) {
	root := newFeedbackProject(t)
	const envName = "MOAI_FEEDBACK_CONTRACT_VALUE"
	const envValue = "feedback-contract-value-0123456789"
	writeFeedbackSecuritySection(t, root, "security:\n  sandbox:\n    env_scrub_extra:\n      - "+envName+"\n")
	t.Setenv(envName, envValue)

	body := "the failure log contained " + envValue + " right here"
	title := "crash while handling " + envValue

	stdout, _, err := runFeedbackCLI(t, body, "scrub", "--title", title, "--root", root)
	if err != nil {
		t.Fatalf("scrub returned error (want exit 0): %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(stdout), &raw); err != nil {
		t.Fatalf("stdout is not a single JSON object: %v\nstdout=%q", err, stdout)
	}
	for _, field := range []string{"verdict", "title", "body", "findings", "reason"} {
		if _, ok := raw[field]; !ok {
			t.Errorf("stdout JSON missing contract field %q (stdout=%q)", field, stdout)
		}
	}

	var res feedback.Result
	if err := json.Unmarshal([]byte(stdout), &res); err != nil {
		t.Fatalf("decode Result: %v", err)
	}
	if res.Verdict != feedback.VerdictOK {
		t.Errorf("verdict = %q, want %q", res.Verdict, feedback.VerdictOK)
	}
	if strings.Contains(res.Title, envValue) || strings.Contains(res.Body, envValue) {
		t.Error("masked output still carries the raw value")
	}
	if len(res.Findings) == 0 {
		t.Fatal("findings is empty; expected the masked value to be reported")
	}

	sawTitle, sawBody := false, false
	for _, f := range res.Findings {
		switch f.Where {
		case feedback.WhereTitle:
			sawTitle = true
		case feedback.WhereBody:
			sawBody = true
		default:
			t.Errorf("finding.where = %q, want %q or %q", f.Where, feedback.WhereTitle, feedback.WhereBody)
		}
	}
	if !sawTitle {
		t.Error("no finding located in the title; the title bypassed the scrubber")
	}
	if !sawBody {
		t.Error("no finding located in the body")
	}
}

// TestFeedbackScrubToolFailureExitsNonZero is AC-F-004: when the scrubber
// cannot load the policy, the command fails fail-closed — non-zero exit and no
// JSON claiming verdict ok on stdout.
func TestFeedbackScrubToolFailureExitsNonZero(t *testing.T) {
	root := newFeedbackProject(t)
	// Malformed YAML: the user's extra sensitive-content patterns cannot be
	// read, so the masking would be weaker than configured.
	writeFeedbackSecuritySection(t, root, "security:\n  extra_sensitive_content_patterns: [unclosed\n")

	stdout, _, err := runFeedbackCLI(t, "an ordinary report", "scrub", "--title", "t", "--root", root)
	if err == nil {
		t.Fatal("scrub succeeded on an unloadable policy; want a non-zero exit")
	}
	if strings.Contains(stdout, feedback.VerdictOK) {
		t.Errorf("stdout claims verdict ok on a tool failure: %q", stdout)
	}
	if strings.TrimSpace(stdout) != "" {
		t.Errorf("stdout is non-empty on a tool failure: %q", stdout)
	}
}

// TestFeedbackScrubBlockedVerdictExitsZero fixes the P3 axis separation: a
// policy block travels in the JSON, never in the exit code.
func TestFeedbackScrubBlockedVerdictExitsZero(t *testing.T) {
	root := newFeedbackProject(t)
	body := "CVE-2026-12345 lets an attacker read arbitrary files"

	stdout, _, err := runFeedbackCLI(t, body, "scrub", "--title", "report", "--root", root)
	if err != nil {
		t.Fatalf("a blocked verdict must still exit 0: %v", err)
	}
	var res feedback.Result
	if err := json.Unmarshal([]byte(stdout), &res); err != nil {
		t.Fatalf("decode Result: %v", err)
	}
	if res.Verdict != feedback.VerdictBlocked {
		t.Fatalf("verdict = %q, want %q", res.Verdict, feedback.VerdictBlocked)
	}
	if res.Reason == "" {
		t.Error("blocked result carries no reason")
	}
}

// TestFeedbackScrubWritesMaskLogUnderRoot observes the artefact wiring: an
// empty Options.ProjectRoot writes nothing at all, so a scrub that resolved no
// root would silently lose the mask log.
func TestFeedbackScrubWritesMaskLogUnderRoot(t *testing.T) {
	root := newFeedbackProject(t)
	body := "token " + cliFakeGitHubToken() + " leaked"

	if _, _, err := runFeedbackCLI(t, body, "scrub", "--title", "t", "--root", root); err != nil {
		t.Fatalf("scrub: %v", err)
	}

	logPath := feedback.MaskLogPathForRoot(root)
	raw, err := os.ReadFile(logPath) //nolint:gosec // test-owned temp path
	if err != nil {
		t.Fatalf("mask log not written at %s: %v", logPath, err)
	}
	if !strings.Contains(string(raw), "kind=secret") {
		t.Errorf("mask log entry does not report the secret finding: %q", raw)
	}
}

// TestFeedbackRootFallbackWalksUpToMarker covers the --root-unset path: the
// root is resolved by walking up to the .moai marker, not left empty.
func TestFeedbackRootFallbackWalksUpToMarker(t *testing.T) {
	root := newFeedbackProject(t)
	nested := filepath.Join(root, "internal", "deep")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("create nested dir: %v", err)
	}

	got, err := resolveFeedbackRoot("", nested)
	if err != nil {
		t.Fatalf("resolveFeedbackRoot: %v", err)
	}
	if filepath.Clean(got) != filepath.Clean(root) {
		t.Errorf("resolved root = %q, want %q", got, root)
	}
}

// TestFeedbackQueueVerbsRoundTrip covers the minimal verb set the skill body
// calls: enqueue a scrubber Result, list it, resolve it.
func TestFeedbackQueueVerbsRoundTrip(t *testing.T) {
	root := newFeedbackProject(t)

	scrubbed, _, err := runFeedbackCLI(t, "a plain report body", "scrub", "--title", "a plain title", "--root", root)
	if err != nil {
		t.Fatalf("scrub: %v", err)
	}

	enqueued, _, err := runFeedbackCLI(t, scrubbed, "queue", "enqueue", "--root", root)
	if err != nil {
		t.Fatalf("queue enqueue: %v", err)
	}
	var item feedback.QueueItem
	if err := json.Unmarshal([]byte(enqueued), &item); err != nil {
		t.Fatalf("decode enqueued item: %v (stdout=%q)", err, enqueued)
	}
	if item.ID == "" {
		t.Fatal("enqueued item has no id")
	}
	if item.Title != "a plain title" {
		t.Errorf("queued title = %q, want the masked title", item.Title)
	}

	listed, _, err := runFeedbackCLI(t, "", "queue", "list", "--root", root)
	if err != nil {
		t.Fatalf("queue list: %v", err)
	}
	var rec feedback.QueueRecord
	if err := json.Unmarshal([]byte(listed), &rec); err != nil {
		t.Fatalf("decode queue record: %v (stdout=%q)", err, listed)
	}
	if len(rec.Items) != 1 || rec.Items[0].ID != item.ID {
		t.Fatalf("queue list = %+v, want exactly the enqueued item", rec.Items)
	}

	if _, _, err := runFeedbackCLI(t, "", "queue", "resolve", item.ID, "--root", root); err != nil {
		t.Fatalf("queue resolve: %v", err)
	}
	listed, _, err = runFeedbackCLI(t, "", "queue", "list", "--root", root)
	if err != nil {
		t.Fatalf("queue list after resolve: %v", err)
	}
	if err := json.Unmarshal([]byte(listed), &rec); err != nil {
		t.Fatalf("decode queue record: %v", err)
	}
	if len(rec.Items) != 0 {
		t.Errorf("queue still holds %d item(s) after resolve", len(rec.Items))
	}
}

// TestFeedbackQueueRefusesBlockedResult keeps the queue from becoming the
// retry path for a report the classifier declined.
func TestFeedbackQueueRefusesBlockedResult(t *testing.T) {
	root := newFeedbackProject(t)

	scrubbed, _, err := runFeedbackCLI(t, "CVE-2026-12345 remote code execution", "scrub", "--title", "t", "--root", root)
	if err != nil {
		t.Fatalf("scrub: %v", err)
	}
	if _, _, err := runFeedbackCLI(t, scrubbed, "queue", "enqueue", "--root", root); err == nil {
		t.Fatal("queue enqueue accepted a blocked result")
	}
}

// TestFeedbackQueueNeverReadsPreScrubDraft is the CLI-level D4 guard. The
// pre-submit draft holds PRE-SCRUB RAW text for a different failure; a queue
// verb that globbed it would put unmasked text into a public issue.
func TestFeedbackQueueNeverReadsPreScrubDraft(t *testing.T) {
	root := newFeedbackProject(t)
	stateDir := filepath.Join(root, defs.MoAIDir, defs.StateSubdir)
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatalf("create state dir: %v", err)
	}
	rawSentinel := "RAW-DRAFT-" + cliFakeGitHubToken()
	draft := filepath.Join(stateDir, "feedback-draft-20260823.md")
	if err := os.WriteFile(draft, []byte(rawSentinel), 0o600); err != nil {
		t.Fatalf("write draft: %v", err)
	}

	listed, errOut, err := runFeedbackCLI(t, "", "queue", "list", "--root", root)
	if err != nil {
		t.Fatalf("queue list: %v", err)
	}
	if strings.Contains(listed, rawSentinel) || strings.Contains(errOut, rawSentinel) {
		t.Fatal("queue list surfaced pre-scrub draft text")
	}
	var rec feedback.QueueRecord
	if err := json.Unmarshal([]byte(listed), &rec); err != nil {
		t.Fatalf("decode queue record: %v", err)
	}
	if len(rec.Items) != 0 {
		t.Errorf("queue list adopted %d draft(s) as queue items", len(rec.Items))
	}
}

// TestFeedbackQueueRequiresRoot: the queue verbs write, so an unresolvable root
// is an error rather than a silent no-op.
func TestFeedbackQueueRequiresRoot(t *testing.T) {
	notAProject := t.TempDir()
	if _, _, err := runFeedbackCLI(t, "", "queue", "list", "--root", notAProject); err == nil {
		t.Fatal("queue list accepted a directory that is not a project root")
	}
}
