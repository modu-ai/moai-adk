package template_test

// SPEC-CI-CODEX-REVIEW-001 M1 — the native Codex review reader.
//
// These tests drive `scripts/ci/codex-review-gate.sh read-native-review` with
// synthetic fixtures in the documented REST shape of a pull-request review and a
// review comment (research E23). No test claims that the live native review
// produces these objects: the artifact form is unverified (research E29, U2).
//
// The same file owns the helpers the M2 and M3 test files reuse (cg* names).
// The tests run offline: bash and jq are required and their absence fails the
// test on linux and darwin; only a windows run skips, with the stated reason.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"
)

const (
	cgBotLogin  = "chatgpt-codex-connector[bot]"
	cgBotID     = 199175422
	cgBotType   = "Bot"
	cgHead      = "0123456789abcdef0123456789abcdef01234567"
	cgOther     = "fedcba9876543210fedcba9876543210fedcba98"
	cgScriptRel = "scripts/ci/codex-review-gate.sh"
)

// cgEntry is one JSON object of a fixture list.
type cgEntry map[string]any

type cgOpt func(cgEntry)

func cgIdentity() map[string]any {
	return map[string]any{"login": cgBotLogin, "id": cgBotID, "type": cgBotType}
}

func cgBy(login string, id int, typ string) cgOpt {
	return func(e cgEntry) { e["user"] = map[string]any{"login": login, "id": id, "type": typ} }
}

func cgUserNull() cgOpt           { return func(e cgEntry) { e["user"] = nil } }
func cgBody(s string) cgOpt       { return func(e cgEntry) { e["body"] = s } }
func cgSet(k string, v any) cgOpt { return func(e cgEntry) { e[k] = v } }
func cgDrop(keys ...string) cgOpt {
	return func(e cgEntry) {
		for _, k := range keys {
			delete(e, k)
		}
	}
}

// cgReview builds a pull-request review authored by the Codex identity.
func cgReview(id int, commit, state string, opts ...cgOpt) cgEntry {
	e := cgEntry{
		"id": id, "user": cgIdentity(), "commit_id": commit, "state": state,
		"submitted_at": "2026-10-18T10:00:00Z", "body": "",
	}
	for _, o := range opts {
		o(e)
	}
	return e
}

// cgComment builds a review comment authored by the Codex identity.
func cgComment(id int, commit, orig string, opts ...cgOpt) cgEntry {
	e := cgEntry{
		"id": id, "user": cgIdentity(), "commit_id": commit, "original_commit_id": orig,
		"path": "internal/x.go", "body": "a finding",
	}
	for _, o := range opts {
		o(e)
	}
	return e
}

// cgWriteList writes one JSON value per line; a string item is written verbatim.
func cgWriteList(t *testing.T, dir, name string, items ...any) string {
	t.Helper()
	var b bytes.Buffer
	for _, it := range items {
		if s, ok := it.(string); ok {
			b.WriteString(s)
		} else {
			raw, err := json.Marshal(it)
			if err != nil {
				t.Fatalf("marshal fixture: %v", err)
			}
			b.Write(raw)
		}
		b.WriteByte('\n')
	}
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, b.Bytes(), 0o600); err != nil {
		t.Fatalf("write %s: %v", p, err)
	}
	return p
}

type cgResult struct {
	code           int
	stdout, stderr string
}

func (r cgResult) all() string { return r.stdout + r.stderr }

// cgTools fails (never skips) when bash or jq is missing; only windows skips.
func cgTools(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the gate script is bash; the consumer runs on ubuntu and these tests run on linux and darwin")
	}
	for _, tool := range []string{"bash", "jq"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Fatalf("%s is required by the gate tests and is not on PATH: %v", tool, err)
		}
	}
}

// cgScriptPath returns the real gate script; its absence fails the test and names the file.
func cgScriptPath(t *testing.T) string {
	t.Helper()
	p := filepath.Join(findProjectRootForMirrorTest(t), filepath.FromSlash(cgScriptRel))
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("gate script absent: %s: %v", p, err)
	}
	return p
}

func cgExec(t *testing.T, script string, extraEnv []string, args ...string) cgResult {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "bash", append([]string{script}, args...)...)
	cmd.Env = append(os.Environ(), extraEnv...)
	var so, se bytes.Buffer
	cmd.Stdout, cmd.Stderr = &so, &se
	err := cmd.Run()
	res := cgResult{stdout: so.String(), stderr: se.String()}
	var ee *exec.ExitError
	switch {
	case err == nil:
	case errors.As(err, &ee):
		res.code = ee.ExitCode()
	default:
		t.Fatalf("run %s: %v", script, err)
	}
	return res
}

// cgCase is one fixture row of read-native-review.
type cgCase struct {
	name        string
	reviews     []any
	comments    []any
	noReviews   bool // the reviews file is not created
	noComments  bool // the comments file is not created
	noFileFlags bool // neither --reviews nor --comments is passed
	draft       string
	closed      string
	head        string
	both        bool // run with --wait-closed false and true, same expectation
	wantCode    int
	wantText    string
}

func cgReadArgs(reviews, comments, head, draft, closed string) []string {
	args := []string{"read-native-review"}
	if reviews != "" {
		args = append(args, "--reviews", reviews)
	}
	if comments != "" {
		args = append(args, "--comments", comments)
	}
	args = append(args, "--head-sha", head, "--draft", draft)
	if closed != "" {
		args = append(args, "--wait-closed", closed)
	}
	return args
}

func cgRunRow(t *testing.T, script string, tc cgCase, closed string) cgResult {
	t.Helper()
	dir := t.TempDir()
	head, draft := tc.head, tc.draft
	if head == "" {
		head = cgHead
	}
	if draft == "" {
		draft = "false"
	}
	var rp, cp string
	if !tc.noFileFlags {
		rp = filepath.Join(dir, "reviews.ndjson")
		cp = filepath.Join(dir, "comments.ndjson")
		if !tc.noReviews {
			rp = cgWriteList(t, dir, "reviews.ndjson", tc.reviews...)
		}
		if !tc.noComments {
			cp = cgWriteList(t, dir, "comments.ndjson", tc.comments...)
		}
	}
	return cgExec(t, script, nil, cgReadArgs(rp, cp, head, draft, closed)...)
}

func cgRunCases(t *testing.T, cases []cgCase) {
	t.Helper()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cgTools(t)
			script := cgScriptPath(t)
			closeds := []string{tc.closed}
			if tc.both {
				closeds = []string{"false", "true"}
			}
			for _, closed := range closeds {
				res := cgRunRow(t, script, tc, closed)
				if res.code != tc.wantCode {
					t.Fatalf("closed=%q: exit %d, want %d\nstdout: %s\nstderr: %s", closed, res.code, tc.wantCode, res.stdout, res.stderr)
				}
				if !strings.Contains(res.all(), tc.wantText) {
					t.Fatalf("closed=%q: output lacks %q\nstdout: %s\nstderr: %s", closed, tc.wantText, res.stdout, res.stderr)
				}
			}
		})
	}
}

var (
	cgHuman  = cgBy("some-human", 4242, "User")
	cgRabbit = cgBy("coderabbitai[bot]", 136622811, "Bot")
)

// AC-CCR-001 (REQ-CCR-001) — the clean reading; noise changes nothing.
// Mutants: an unconditional exit 0 (also fails AC-CCR-002..005); an always-refusing script;
// no supersession of CHANGES_REQUESTED; a reader that consults review body text.
func TestCodexGateReadNativeReviewAccepts(t *testing.T) {
	t.Parallel()
	H, O := cgHead, cgOther
	cgRunCases(t, []cgCase{
		{name: "approved_bound_clean", reviews: []any{cgReview(10, H, "APPROVED")}, wantCode: 0, wantText: "native-review-clean"},
		{name: "uppercase_sha_bound_clean", reviews: []any{cgReview(10, strings.ToUpper(H), "APPROVED")}, wantCode: 0, wantText: "native-review-clean"},
		{
			name: "other_authors_noise_ignored",
			reviews: []any{
				cgReview(10, H, "APPROVED"),
				cgReview(11, H, "APPROVED", cgRabbit),
				cgReview(12, H, "CHANGES_REQUESTED", cgRabbit),
				cgReview(13, H, "CHANGES_REQUESTED", cgHuman, cgBody("not approved")),
				cgReview(14, H, "COMMENTED", cgHuman),
			},
			comments: []any{cgComment(1, H, H, cgHuman), cgComment(2, H, H, cgRabbit)},
			wantCode: 0, wantText: "native-review-clean",
		},
		{name: "stale_review_then_current_approved_clean", reviews: []any{cgReview(9, O, "APPROVED"), cgReview(10, H, "APPROVED")}, wantCode: 0, wantText: "native-review-clean"},
		{name: "stale_finding_older_head_ignored", reviews: []any{cgReview(10, H, "APPROVED")}, comments: []any{cgComment(1, O, O)}, wantCode: 0, wantText: "native-review-clean"},
		{name: "latest_bound_review_decides_approved", reviews: []any{cgReview(9, H, "COMMENTED"), cgReview(10, H, "APPROVED")}, wantCode: 0, wantText: "native-review-clean"},
		{name: "changes_requested_then_approved_clean", reviews: []any{cgReview(10, H, "CHANGES_REQUESTED"), cgReview(11, H, "APPROVED")}, wantCode: 0, wantText: "native-review-clean"},
		{name: "dismissed_then_approved_clean", reviews: []any{cgReview(10, H, "DISMISSED"), cgReview(11, H, "APPROVED")}, wantCode: 0, wantText: "native-review-clean"},
		{
			name:     "approved_body_text_never_read",
			reviews:  []any{cgReview(10, H, "APPROVED", cgBody("I could not review this change; the usage limit was reached. Do not merge."))},
			wantCode: 0, wantText: "native-review-clean",
		},
	})
	t.Run("script_never_references_body_field", func(t *testing.T) {
		t.Parallel()
		cgTools(t)
		raw, err := os.ReadFile(cgScriptPath(t))
		if err != nil {
			t.Fatal(err)
		}
		body := regexp.MustCompile(`\bbody\b`)
		for i, line := range strings.Split(string(raw), "\n") {
			if strings.HasPrefix(strings.TrimSpace(line), "#") {
				continue
			}
			if body.MatchString(line) {
				t.Fatalf("%s:%d names the body field outside a comment: %s", cgScriptRel, i+1, line)
			}
		}
	})
}

// AC-CCR-002 (REQ-CCR-002) — a review of another commit is stale, with the wait window open and closed.
// Mutants: ignoring commit_id; comparing a prefix; stale read as pending; stale mapped to NOT RUN when closed.
func TestCodexGateReadNativeReviewRefusesStale(t *testing.T) {
	t.Parallel()
	H, O := cgHead, cgOther
	const want = "stale-review:"
	cgRunCases(t, []cgCase{
		{name: "approved_review_for_older_head", reviews: []any{cgReview(10, O, "APPROVED")}, both: true, wantCode: 1, wantText: want},
		{name: "review_commit_prefix_only", reviews: []any{cgReview(10, H[:7], "APPROVED")}, both: true, wantCode: 1, wantText: want},
		{name: "review_commit_empty_string", reviews: []any{cgReview(10, "", "APPROVED")}, both: true, wantCode: 1, wantText: want},
		{name: "review_for_other_commit_after_push", reviews: []any{cgReview(10, O, "COMMENTED"), cgReview(11, O, "APPROVED")}, both: true, wantCode: 1, wantText: want},
		{
			name:    "forged_bound_review_beside_real_stale_review_still_stale",
			reviews: []any{cgReview(10, O, "APPROVED"), cgReview(11, H, "APPROVED", cgRabbit)},
			both:    true, wantCode: 1, wantText: want,
		},
		{name: "stale_review_window_closed_still_exit_1", reviews: []any{cgReview(10, O, "APPROVED")}, closed: "true", wantCode: 1, wantText: want},
	})
}

// AC-CCR-003 (REQ-CCR-003) — only the identity counts; exclusion is by author, never by body text.
// Mutants: login only; id only; any bound review of any author; body text read for "approved";
// exclusion of any review whose body says approved; another login's CHANGES_REQUESTED blocking.
func TestCodexGateReadNativeReviewIgnoresForged(t *testing.T) {
	t.Parallel()
	H, O := cgHead, cgOther
	pending := "pending:no-review"
	cgRunCases(t, []cgCase{
		{name: "other_bot_approved_bound", reviews: []any{cgReview(10, H, "APPROVED", cgRabbit)}, wantCode: 10, wantText: pending},
		{name: "human_review_says_approved", reviews: []any{cgReview(10, H, "APPROVED", cgHuman, cgBody("approved"))}, wantCode: 10, wantText: pending},
		{name: "lookalike_login_without_bot_suffix_same_id", reviews: []any{cgReview(10, H, "APPROVED", cgBy("chatgpt-codex-connector", cgBotID, "Bot"))}, wantCode: 10, wantText: pending},
		{name: "same_login_other_id", reviews: []any{cgReview(10, H, "APPROVED", cgBy(cgBotLogin, 1, "Bot"))}, wantCode: 10, wantText: pending},
		{name: "same_login_type_user", reviews: []any{cgReview(10, H, "APPROVED", cgBy(cgBotLogin, cgBotID, "User"))}, wantCode: 10, wantText: pending},
		{name: "review_user_null", reviews: []any{cgReview(10, H, "APPROVED", cgUserNull())}, wantCode: 10, wantText: pending},
		{
			name:     "forged_bound_review_beside_real_stale_review_not_pass",
			reviews:  []any{cgReview(10, O, "APPROVED"), cgReview(11, H, "APPROVED", cgHuman, cgBody("approved"))},
			wantCode: 1, wantText: "stale-review:",
		},
		{
			name:     "identity_approved_beside_other_login_changes_requested_clean",
			reviews:  []any{cgReview(10, H, "APPROVED"), cgReview(11, H, "CHANGES_REQUESTED", cgHuman)},
			wantCode: 0, wantText: "native-review-clean",
		},
		{
			name:     "other_login_commented_never_completes_identity_review",
			reviews:  []any{cgReview(10, H, "COMMENTED"), cgReview(11, H, "APPROVED", cgHuman, cgBody("approved"))},
			wantCode: 1, wantText: "review-not-conclusive:COMMENTED",
		},
		{
			name:     "identity_approved_body_says_approved_not_excluded",
			reviews:  []any{cgReview(10, H, "APPROVED", cgBody("approved")), cgReview(11, H, "APPROVED", cgHuman, cgBody("approved"))},
			wantCode: 0, wantText: "native-review-clean",
		},
	})
}

// AC-CCR-004 (REQ-CCR-004) — APPROVED is the only clean state; every other form is refused with a named reason.
// Mutants: the revision-3 reader (the newest bound review decides, COMMENTED or APPROVED counts as complete);
// the oldest bound review deciding; body text read for a clean result; DISMISSED or PENDING complete;
// submitted_at ignored; an entry with a missing field passed.
func TestCodexGateReadNativeReviewRefusesIncomplete(t *testing.T) {
	t.Parallel()
	H, O := cgHead, cgOther
	const (
		commented = "review-not-conclusive:COMMENTED"
		changes   = "changes-requested"
	)
	cgRunCases(t, []cgCase{
		{name: "commented_empty_body_refused", reviews: []any{cgReview(10, H, "COMMENTED")}, wantCode: 1, wantText: commented},
		{name: "commented_body_lgtm_refused", reviews: []any{cgReview(10, H, "COMMENTED", cgBody("LGTM"))}, wantCode: 1, wantText: commented},
		{name: "commented_findings_only_in_body_refused", reviews: []any{cgReview(10, H, "COMMENTED", cgBody("1. nil dereference in x.go\n2. missing error check"))}, wantCode: 1, wantText: commented},
		{name: "commented_could_not_review_body_refused", reviews: []any{cgReview(10, H, "COMMENTED", cgBody("I could not review this change"))}, wantCode: 1, wantText: commented},
		{name: "commented_usage_limit_notice_refused", reviews: []any{cgReview(10, H, "COMMENTED", cgBody("You have reached your Codex usage limits for code reviews."))}, wantCode: 1, wantText: commented},
		{name: "changes_requested_alone_refused", reviews: []any{cgReview(10, H, "CHANGES_REQUESTED")}, wantCode: 1, wantText: changes},
		{name: "approved_then_changes_requested_refused", reviews: []any{cgReview(10, H, "APPROVED"), cgReview(11, H, "CHANGES_REQUESTED")}, wantCode: 1, wantText: changes},
		{name: "changes_requested_then_commented_refused", reviews: []any{cgReview(10, H, "CHANGES_REQUESTED"), cgReview(11, H, "COMMENTED")}, wantCode: 1, wantText: changes},
		{name: "changes_requested_then_dismissed_refused", reviews: []any{cgReview(10, H, "CHANGES_REQUESTED"), cgReview(11, H, "DISMISSED")}, wantCode: 1, wantText: changes},
		{name: "approved_then_commented_refused", reviews: []any{cgReview(10, H, "APPROVED"), cgReview(11, H, "COMMENTED")}, wantCode: 1, wantText: commented},
		{name: "approved_then_dismissed_refused", reviews: []any{cgReview(10, H, "APPROVED"), cgReview(11, H, "DISMISSED")}, wantCode: 1, wantText: "review-not-complete:DISMISSED"},
		{name: "approved_other_sha_and_commented_head_refused", reviews: []any{cgReview(10, O, "APPROVED"), cgReview(11, H, "COMMENTED")}, wantCode: 1, wantText: commented},
		{name: "state_dismissed", reviews: []any{cgReview(10, H, "DISMISSED")}, wantCode: 1, wantText: "review-not-complete:DISMISSED"},
		{name: "state_pending", reviews: []any{cgReview(10, H, "PENDING")}, wantCode: 1, wantText: "review-not-complete:PENDING"},
		{name: "state_unknown_value", reviews: []any{cgReview(10, H, "WEIRD_STATE")}, wantCode: 1, wantText: "review-not-complete:WEIRD_STATE"},
		{name: "approved_submitted_at_null", reviews: []any{cgReview(10, H, "APPROVED", cgSet("submitted_at", nil))}, wantCode: 1, wantText: "review-not-complete:no-submitted-at"},
		{name: "approved_submitted_at_empty", reviews: []any{cgReview(10, H, "APPROVED", cgSet("submitted_at", ""))}, wantCode: 1, wantText: "review-not-complete:no-submitted-at"},
		{name: "commit_id_missing", reviews: []any{cgReview(10, H, "APPROVED", cgDrop("commit_id"))}, wantCode: 1, wantText: "review-entry-invalid:10"},
		{name: "state_missing", reviews: []any{cgReview(10, H, "APPROVED", cgDrop("state"))}, wantCode: 1, wantText: "review-entry-invalid:10"},
		{name: "id_missing", reviews: []any{cgReview(10, H, "APPROVED", cgDrop("id"))}, wantCode: 1, wantText: "review-entry-invalid:"},
		{name: "duplicate_review_id", reviews: []any{cgReview(10, H, "APPROVED"), cgReview(10, H, "COMMENTED")}, wantCode: 1, wantText: "review-entry-invalid:duplicate-id:10"},
		{name: "review_entry_not_object", reviews: []any{"42"}, wantCode: 1, wantText: "input-invalid"},
		{name: "reviews_file_missing", noReviews: true, wantCode: 1, wantText: "input-invalid"},
		{name: "comments_file_missing", reviews: []any{cgReview(10, H, "APPROVED")}, noComments: true, wantCode: 1, wantText: "input-invalid"},
	})
}

// AC-CCR-005 (REQ-CCR-005) — a finding comment of the identity on the head refuses the review.
// Mutants: comments never read; only commit_id counted; comments of any author counted; an unbound comment ignored.
func TestCodexGateReadNativeReviewRefusesFinding(t *testing.T) {
	t.Parallel()
	H, O := cgHead, cgOther
	cgRunCases(t, []cgCase{
		{name: "finding_commit_id_equals_head", reviews: []any{cgReview(10, H, "APPROVED")}, comments: []any{cgComment(77, H, O)}, wantCode: 1, wantText: "finding:77"},
		{name: "finding_original_commit_id_equals_head", reviews: []any{cgReview(10, H, "APPROVED")}, comments: []any{cgComment(78, O, H)}, wantCode: 1, wantText: "finding:78"},
		{name: "finding_without_any_commit_field", reviews: []any{cgReview(10, H, "APPROVED")}, comments: []any{cgComment(79, "", "", cgDrop("commit_id", "original_commit_id"))}, wantCode: 1, wantText: "finding-unbound:79"},
		{name: "finding_by_other_author_ignored", reviews: []any{cgReview(10, H, "APPROVED")}, comments: []any{cgComment(80, H, H, cgRabbit), cgComment(81, H, H, cgHuman)}, wantCode: 0, wantText: "native-review-clean"},
		{name: "finding_bound_to_older_head_ignored", reviews: []any{cgReview(10, H, "APPROVED")}, comments: []any{cgComment(82, O, O)}, wantCode: 0, wantText: "native-review-clean"},
		{name: "comments_file_empty_clean", reviews: []any{cgReview(10, H, "APPROVED")}, wantCode: 0, wantText: "native-review-clean"},
		{name: "approved_review_with_finding_refused", reviews: []any{cgReview(10, H, "APPROVED")}, comments: []any{cgComment(83, H, H)}, wantCode: 1, wantText: "finding:83"},
		{name: "commented_review_with_inline_finding_refused", reviews: []any{cgReview(10, H, "COMMENTED")}, comments: []any{cgComment(84, H, H)}, wantCode: 1, wantText: "finding:84"},
	})
}

// AC-CCR-007 (script rows) — no review of the identity yet is pending, never a pass.
// Mutant: an absent review read as success.
func TestCodexGateReadNativeReviewPending(t *testing.T) {
	t.Parallel()
	cgRunCases(t, []cgCase{
		{name: "empty_reviews_pending", comments: []any{cgComment(1, cgHead, cgHead, cgHuman)}, closed: "false", wantCode: 10, wantText: "pending:no-review"},
		{name: "empty_reviews_and_comments_pending", closed: "false", wantCode: 10, wantText: "pending:no-review"},
	})
}

// AC-CCR-008 (script rows) — NOT RUN is one named result for a draft and for a review absent at window close.
// Mutants: a draft merged (exit 0); NOT RUN indistinguishable from failure (exit 1); absence never final
// (10 at window close); a stale review mapped to 11.
func TestCodexGateReadNativeReviewDraftNotRun(t *testing.T) {
	t.Parallel()
	H, O := cgHead, cgOther
	cgRunCases(t, []cgCase{
		{name: "draft_true_exit_11", draft: "true", closed: "false", wantCode: 11, wantText: "not-run:draft"},
		{name: "draft_true_with_clean_bound_review_still_11", draft: "true", reviews: []any{cgReview(10, H, "APPROVED")}, closed: "false", wantCode: 11, wantText: "not-run:draft"},
		{name: "draft_true_without_input_files_still_11", draft: "true", noFileFlags: true, closed: "false", wantCode: 11, wantText: "not-run:draft"},
		{name: "draft_false_approved_exit_0", draft: "false", reviews: []any{cgReview(10, H, "APPROVED")}, closed: "true", wantCode: 0, wantText: "native-review-clean"},
		{name: "draft_flag_invalid_exit_1", draft: "maybe", closed: "false", wantCode: 1, wantText: "draft-flag-unusable"},
		{name: "wait_closed_flag_invalid_exit_1", draft: "false", closed: "perhaps", wantCode: 1, wantText: "wait-closed-flag-unusable"},
		{name: "review_absent_window_open_exit_10", draft: "false", closed: "false", wantCode: 10, wantText: "pending:no-review"},
		{name: "review_absent_window_closed_exit_11", draft: "false", closed: "true", wantCode: 11, wantText: "not-run:review-absent"},
		{name: "changes_requested_window_closed_still_exit_1", draft: "false", reviews: []any{cgReview(10, H, "CHANGES_REQUESTED")}, closed: "true", wantCode: 1, wantText: "changes-requested"},
		{name: "stale_window_closed_is_not_exit_11", draft: "false", reviews: []any{cgReview(10, O, "APPROVED")}, closed: "true", wantCode: 1, wantText: "stale-review:"},
	})
	t.Run("four_classes_pairwise_distinct", func(t *testing.T) {
		t.Parallel()
		cgTools(t)
		script := cgScriptPath(t)
		rows := []cgCase{
			{reviews: []any{cgReview(10, H, "APPROVED")}, closed: "true"},          // pass
			{reviews: []any{cgReview(10, H, "CHANGES_REQUESTED")}, closed: "true"}, // fail
			{closed: "false"}, // pending
			{draft: "true", closed: "false", noFileFlags: true}, // NOT RUN
		}
		seen := map[int]bool{}
		var codes []int
		for _, row := range rows {
			res := cgRunRow(t, script, row, row.closed)
			codes = append(codes, res.code)
			seen[res.code] = true
		}
		want := []int{0, 1, 10, 11}
		for i, c := range codes {
			if c != want[i] {
				t.Fatalf("exit classes %v, want %v", codes, want)
			}
		}
		if len(seen) != 4 {
			t.Fatalf("exit classes are not pairwise distinct: %v", codes)
		}
	})
}
