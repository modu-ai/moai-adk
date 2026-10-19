package template_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func cgNativeIssueComments(t *testing.T, head string) []any {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(findProjectRootForMirrorTest(t), "internal/template/testdata/codex-native-clean.json"))
	if err != nil {
		t.Fatal(err)
	}
	const fixtureHead = "e7cceb2cb5d6e1cd9dbd2f465c4b9d5a1d631f21"
	text := strings.ReplaceAll(string(raw), fixtureHead, head)
	text = strings.ReplaceAll(text, fixtureHead[:10], head[:10])
	var comments []any
	if err := json.Unmarshal([]byte(text), &comments); err != nil {
		t.Fatal(err)
	}
	return comments
}

// Captured native issue-comment artifacts; no formal APPROVED review exists.
func TestCodexNativeReceipt(t *testing.T) {
	cgTools(t)
	raw, err := os.ReadFile(filepath.Join(findProjectRootForMirrorTest(t), "internal/template/testdata/codex-native-clean.json"))
	if err != nil {
		t.Fatal(err)
	}
	var real []cgEntry
	if err := json.Unmarshal(raw, &real); err != nil {
		t.Fatal(err)
	}
	const head = "e7cceb2cb5d6e1cd9dbd2f465c4b9d5a1d631f21"
	for _, tc := range []struct {
		name string
		edit func([]cgEntry) []cgEntry
		want int
	}{
		{"real_clean", nil, 0},
		{"large_human_comment", func(v []cgEntry) []cgEntry {
			for i := 0; i < 50; i++ {
				v = append(v, cgEntry{"id": i + 7, "user": map[string]any{"id": 1, "login": "human", "type": "User"}, "body": strings.Repeat("x", 60000)})
			}
			return v
		}, 0},
		{"invalid_id", func(v []cgEntry) []cgEntry { v[0]["id"] = "6260255148"; return v }, 1},
		{"missing_timestamp", func(v []cgEntry) []cgEntry { delete(v[0], "updated_at"); return v }, 1},
		{"newer_error_summary", func(v []cgEntry) []cgEntry {
			var n cgEntry
			b, _ := json.Marshal(v[0])
			_ = json.Unmarshal(b, &n)
			n["id"] = 9999999999
			n["updated_at"] = "2026-10-19T23:00:00Z"
			n["body"] = strings.ReplaceAll(n["body"].(string), `"status":"completed"`, `"status":"error"`)
			return append(v, n)
		}, 1},
		{"newer_stale_receipt_is_not_verdict", func(v []cgEntry) []cgEntry {
			var n cgEntry
			b, _ := json.Marshal(v[1])
			_ = json.Unmarshal(b, &n)
			n["id"] = 9999999999
			n["created_at"] = "2026-10-19T23:00:00Z"
			n["body"] = strings.ReplaceAll(n["body"].(string), head[:10], cgOther[:10])
			return append(v, n)
		}, 0},
		{"duplicate_summary", func(v []cgEntry) []cgEntry {
			var n cgEntry
			b, _ := json.Marshal(v[0])
			_ = json.Unmarshal(b, &n)
			n["id"] = 9999999999
			n["created_at"] = "2026-10-19T23:00:00Z"
			n["updated_at"] = "2026-10-19T23:00:00Z"
			return append(v, n)
		}, 1},
		{"missing_status", func(v []cgEntry) []cgEntry {
			v[0]["body"] = strings.ReplaceAll(v[0]["body"].(string), `"status":"completed"`, `"other":"completed"`)
			return v
		}, 1},
		{"unknown_findings", func(v []cgEntry) []cgEntry {
			v[0]["body"] = strings.ReplaceAll(v[0]["body"].(string), `"has_unresolved_findings":false`, `"has_unresolved_findings":null`)
			return v
		}, 1},
		{"absent", func(v []cgEntry) []cgEntry { return nil }, 11},
		{"summary_only", func(v []cgEntry) []cgEntry { return v[:1] }, 0},
		{"receipt_only", func(v []cgEntry) []cgEntry { return v[1:] }, 11},
		{"missing_security_marker", func(v []cgEntry) []cgEntry {
			v[0]["body"] = strings.Replace(v[0]["body"].(string), "<!-- codex-security-review:v1 ", "<!-- missing-security-marker ", 1)
			return v
		}, 1},
		{"stale_security_leg", func(v []cgEntry) []cgEntry {
			v[0]["body"] = strings.Replace(v[0]["body"].(string), `"security_review":{"commit_sha":"e7cceb2cb5d6e1cd9dbd2f465c4b9d5a1d631f21"`, `"security_review":{"commit_sha":"`+cgOther+`"`, 1)
			return v
		}, 1},
		{"unresolved_security_finding", func(v []cgEntry) []cgEntry {
			v[0]["body"] = strings.Replace(v[0]["body"].(string), `"security_review":{"commit_sha":"e7cceb2cb5d6e1cd9dbd2f465c4b9d5a1d631f21","has_any_findings":false,"has_unresolved_findings":false`, `"security_review":{"commit_sha":"e7cceb2cb5d6e1cd9dbd2f465c4b9d5a1d631f21","has_any_findings":false,"has_unresolved_findings":true`, 1)
			return v
		}, 1},
		{"security_leg_running_waits", func(v []cgEntry) []cgEntry {
			v[0]["body"] = strings.Replace(v[0]["body"].(string), `"security_review":{"commit_sha":"e7cceb2cb5d6e1cd9dbd2f465c4b9d5a1d631f21","has_any_findings":false,"has_unresolved_findings":false,"status":"completed"}`, `"security_review":{"commit_sha":"e7cceb2cb5d6e1cd9dbd2f465c4b9d5a1d631f21","has_any_findings":false,"has_unresolved_findings":false,"status":"running"}`, 1)
			return v
		}, 1},
		{"stale", func(v []cgEntry) []cgEntry {
			v[0]["body"] = strings.ReplaceAll(v[0]["body"].(string), head, cgOther)
			return v
		}, 1},
		{"error", func(v []cgEntry) []cgEntry {
			v[0]["body"] = strings.ReplaceAll(v[0]["body"].(string), `"status":"completed"`, `"status":"error"`)
			return v
		}, 1},
		{"running", func(v []cgEntry) []cgEntry {
			v[0]["body"] = strings.ReplaceAll(v[0]["body"].(string), `"status":"completed"`, `"status":"running"`)
			return v
		}, 1},
		{"unresolved", func(v []cgEntry) []cgEntry {
			v[0]["body"] = strings.ReplaceAll(v[0]["body"].(string), `"has_unresolved_findings":false`, `"has_unresolved_findings":true`)
			return v
		}, 1},
		{"malformed", func(v []cgEntry) []cgEntry {
			v[0]["body"] = strings.ReplaceAll(v[0]["body"].(string), `{"code_review":`, `{broken:`)
			return v
		}, 1},
		{"spoof_id", func(v []cgEntry) []cgEntry {
			for _, x := range v {
				x["user"] = map[string]any{"id": 1, "login": cgBotLogin, "type": "Bot"}
			}
			return v
		}, 11},
		{"stale_receipt_is_not_verdict", func(v []cgEntry) []cgEntry {
			v[1]["body"] = strings.ReplaceAll(v[1]["body"].(string), head[:10], cgOther[:10])
			return v
		}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var v []cgEntry
			if err := json.Unmarshal(raw, &v); err != nil {
				t.Fatal(err)
			}
			if tc.edit != nil {
				v = tc.edit(v)
			}
			dir := t.TempDir()
			var entries []any
			for _, x := range v {
				entries = append(entries, x)
			}
			issues := cgWriteList(t, dir, "issues", entries...)
			reviews := cgWriteList(t, dir, "reviews")
			comments := cgWriteList(t, dir, "comments")
			args := append(cgReadArgs(reviews, comments, head, "false", "true"), "--issue-comments", issues)
			got := cgExec(t, cgScriptPath(t), nil, args...)
			if got.code != tc.want {
				t.Fatalf("exit=%d want=%d: %s", got.code, tc.want, got.all())
			}
		})
	}
}

func TestCodexNativeRunningSecurityLegWaits(t *testing.T) {
	cgTools(t)
	comments := cgNativeIssueComments(t, cgHead)
	state := comments[0].(map[string]any)
	body := state["body"].(string)
	body = strings.Replace(body, `"security_review":{"commit_sha":"`+cgHead+`","has_any_findings":false,"has_unresolved_findings":false,"status":"completed"}`, `"security_review":{"commit_sha":"`+cgHead+`","has_any_findings":false,"has_unresolved_findings":false,"status":"running"}`, 1)
	state["body"] = strings.ReplaceAll(body, "<!-- codex-security-review:v1 "+`{"blockingSeverityThreshold":"P0","headSha":"`+cgHead+`","mergeGateEnabled":false,"pullRequestNumber":1842,"repository":"modu-ai/moai-adk","status":"completed"} -->`+"\n", "")
	dir := t.TempDir()
	issues := cgWriteList(t, dir, "issues", comments...)
	reviews := cgWriteList(t, dir, "reviews")
	reviewComments := cgWriteList(t, dir, "review-comments")
	args := append(cgReadArgs(reviews, reviewComments, cgHead, "false", "false"), "--issue-comments", issues)
	got := cgExec(t, cgScriptPath(t), nil, args...)
	if got.code != 10 || !strings.Contains(got.all(), "pending:running") {
		t.Fatalf("running security leg must wait for the bounded review window: exit=%d output=%s", got.code, got.all())
	}
}

// Execute the actual workflow with native issue comments and no formal reviews.
func TestAutoMergeNativeReceiptRecheck(t *testing.T) {
	cgTools(t)
	raw, err := os.ReadFile(filepath.Join(findProjectRootForMirrorTest(t), "internal/template/testdata/codex-native-clean.json"))
	if err != nil {
		t.Fatal(err)
	}
	raw = []byte(strings.ReplaceAll(string(raw), "e7cceb2cb5d6e1cd9dbd2f465c4b9d5a1d631f21", cgHead))
	raw = []byte(strings.ReplaceAll(string(raw), "e7cceb2cb5", cgHead[:10]))
	for _, tc := range []struct {
		name   string
		change bool
	}{{"clean", false}, {"summary_changes_same_count", true}} {
		t.Run(tc.name, func(t *testing.T) {
			var entries []any
			if err := json.Unmarshal(raw, &entries); err != nil {
				t.Fatal(err)
			}
			r := newAMRig(t, "real")
			r.e.setPages("files", cgBenignFiles(2, cgHead))
			r.e.setPages("issues", entries)
			pull := strings.ReplaceAll(amPull(cgHead, false, 2, 0), `"comments":0`, `"comments":2`)
			r.setPulls(pull)
			if tc.change {
				var v []any
				changed := strings.ReplaceAll(string(raw), `\"status\":\"completed\"`, `\"status\":\"error\"`)
				if changed == string(raw) {
					t.Fatal("mutation did not change summary")
				}
				if err := json.Unmarshal([]byte(changed), &v); err != nil {
					t.Fatal(err)
				}
				amSetGen(t, r, "issues", 2, v...)
			}
			r.runChain()
			got := len(r.mergeCalls())
			want := 1
			if tc.change {
				want = 0
			}
			if got != want {
				t.Fatalf("merges=%d want=%d review=%s", got, want, r.results["review"].all())
			}
		})
	}
}

func TestCodexNativeIssuePaging(t *testing.T) {
	for _, tc := range []struct {
		name  string
		fail  bool
		count string
		want  int
	}{{"two_pages", false, "150", 0}, {"count_mismatch", false, "151", 1}, {"page_failure", true, "150", 1}} {
		t.Run(tc.name, func(t *testing.T) {
			e := newCGEnv(t)
			e.setPages("issues", cgReviewList(150, cgHuman))
			out := filepath.Join(e.dir, "issues.ndjson")
			env := e.env()
			if tc.fail {
				env = e.env("FAIL_PAGE=issues:2")
			}
			got := cgExec(t, cgScriptPath(t), env, "fetch-pages", "--endpoint", "repos/modu-ai/moai-adk/issues/7/comments", "--out", out, "--expect-count", tc.count)
			if got.code != tc.want {
				t.Fatalf("exit=%d want=%d: %s", got.code, tc.want, got.all())
			}
			if tc.want != 0 {
				cgMustNotExist(t, out)
			}
		})
	}
}
