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

// Native issue-comment fixture with current code and security receipts; no formal APPROVED review exists.
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
		{"newer_stale_receipt_refuses", func(v []cgEntry) []cgEntry {
			var n cgEntry
			b, _ := json.Marshal(v[1])
			_ = json.Unmarshal(b, &n)
			n["id"] = 9999999999
			n["created_at"] = "2026-10-19T23:00:00Z"
			n["body"] = strings.ReplaceAll(n["body"].(string), head[:10], cgOther[:10])
			return append(v, n)
		}, 1},
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
		{"summary_only", func(v []cgEntry) []cgEntry { return v[:1] }, 1},
		{"receipt_only", func(v []cgEntry) []cgEntry { return v[1:] }, 11},
		{"receipt_wrong_head", func(v []cgEntry) []cgEntry {
			v[1]["body"] = strings.ReplaceAll(v[1]["body"].(string), head[:10], cgOther[:10])
			return v
		}, 1},
		{"receipt_not_clean", func(v []cgEntry) []cgEntry {
			v[1]["body"] = strings.Replace(v[1]["body"].(string), "Didn't find any major issues.", "Found major issues.", 1)
			return v
		}, 1},
		{"receipt_marker_missing", func(v []cgEntry) []cgEntry {
			v[1]["body"] = strings.Replace(v[1]["body"].(string), "<!-- codex-review-receipt:v1 -->", "<!-- other-marker -->", 1)
			return v
		}, 1},
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
		{"stale_receipt_refuses", func(v []cgEntry) []cgEntry {
			v[1]["body"] = strings.ReplaceAll(v[1]["body"].(string), head[:10], cgOther[:10])
			return v
		}, 1},
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

func cgLoadNative1828Fixture(t *testing.T, name string) []cgEntry {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(findProjectRootForMirrorTest(t), "internal/template/testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	var entries []cgEntry
	if err := json.Unmarshal(raw, &entries); err != nil {
		t.Fatal(err)
	}
	return entries
}

func TestCodexNative1828ReceiptTypesAndHeadBinding(t *testing.T) {
	cgTools(t)
	unresolved := cgLoadNative1828Fixture(t, "codex-native-1828-unresolved.json")
	clean := cgLoadNative1828Fixture(t, "codex-native-1828-clean.json")

	run := func(t *testing.T, entries []cgEntry, wantCode int, wantText string) {
		t.Helper()
		dir := t.TempDir()
		var issueEntries []any
		for _, entry := range entries {
			issueEntries = append(issueEntries, entry)
		}
		issues := cgWriteList(t, dir, "issues", issueEntries...)
		reviews := cgWriteList(t, dir, "reviews")
		comments := cgWriteList(t, dir, "comments")
		args := append(cgReadArgs(reviews, comments, "ab609cafbf8f421164100c626db93048f888655f", "false", "true"), "--issue-comments", issues)
		got := cgExec(t, cgScriptPath(t), nil, args...)
		if got.code != wantCode || !strings.Contains(got.all(), wantText) {
			t.Fatalf("exit=%d want=%d output=%s; want text %q", got.code, wantCode, got.all(), wantText)
		}
	}

	t.Run("actual_unresolved_review_remains_blocked", func(t *testing.T) {
		run(t, unresolved, 1, "finding:unresolved:code_review")
	})
	t.Run("actual_unresolved_review_stays_blocked_with_keep_it_up_receipt", func(t *testing.T) {
		entries := make([]cgEntry, 0, len(unresolved))
		for _, entry := range unresolved {
			clone := make(cgEntry, len(entry))
			for key, value := range entry {
				clone[key] = value
			}
			if body, ok := clone["body"].(string); ok {
				clone["body"] = strings.Replace(body, "Codex Review: Didn't find any major issues. Chef's kiss.", "Codex Review: Didn't find any major issues. Keep it up!", 1)
			}
			entries = append(entries, clone)
		}
		run(t, entries, 1, "finding:unresolved:code_review")
	})
	t.Run("untrusted_marker_only_does_not_block_valid_native_receipts", func(t *testing.T) {
		entries := make([]cgEntry, 0, len(clean)+1)
		for _, entry := range clean {
			clone := make(cgEntry, len(entry))
			for key, value := range entry {
				clone[key] = value
			}
			entries = append(entries, clone)
		}
		entries = append(entries, cgEntry{
			"id":         9999999999,
			"user":       map[string]any{"id": 90000001, "login": "untrusted-test-user", "type": "User"},
			"body":       "<!-- codex-review-receipt:v1 -->",
			"created_at": "2026-10-19T14:40:00Z",
			"updated_at": "2026-10-19T14:40:00Z",
		})
		run(t, entries, 0, "native-review-clean")
	})

	for _, tc := range []struct {
		name   string
		edit   func([]cgEntry) []cgEntry
		want   int
		reason string
	}{
		{name: "security_receipt_newest_actual_clean_text", want: 0, reason: "native-review-clean"},
		{name: "code_receipt_newest_actual_clean_text", edit: func(v []cgEntry) []cgEntry {
			for _, entry := range v {
				if entry["id"] == float64(6267776147) {
					entry["id"] = float64(6267807386)
				} else if entry["id"] == float64(6267807385) {
					entry["id"] = float64(6267807385)
				}
			}
			return v
		}, want: 0, reason: "native-review-clean"},
		{name: "legacy_code_clean_text_security_newest", edit: func(v []cgEntry) []cgEntry {
			for _, entry := range v {
				if entry["id"] == float64(6267776147) {
					entry["body"] = strings.Replace(entry["body"].(string), "Codex Review: Didn't find any major issues. Chef's kiss.", "Codex Review: Didn't find any major issues. :+1:", 1)
				}
			}
			return v
		}, want: 0, reason: "native-review-clean"},
		{name: "code_receipt_actual_bravo_clean_text", edit: func(v []cgEntry) []cgEntry {
			for _, entry := range v {
				if entry["id"] == float64(6267776147) {
					entry["body"] = strings.Replace(entry["body"].(string), "Codex Review: Didn't find any major issues. Chef's kiss.", "Codex Review: Didn't find any major issues. Bravo.", 1)
				}
			}
			return v
		}, want: 0, reason: "native-review-clean"},
		{name: "code_receipt_actual_keep_it_up_clean_text_security_newest", edit: func(v []cgEntry) []cgEntry {
			for _, entry := range v {
				if entry["id"] == float64(6267776147) {
					entry["body"] = strings.Replace(entry["body"].(string), "Codex Review: Didn't find any major issues. Chef's kiss.", "Codex Review: Didn't find any major issues. Keep it up!", 1)
				}
			}
			return v
		}, want: 0, reason: "native-review-clean"},
		{name: "code_receipt_actual_keep_it_up_clean_text_code_newest", edit: func(v []cgEntry) []cgEntry {
			for _, entry := range v {
				if entry["id"] == float64(6267776147) {
					entry["id"] = float64(6267807386)
					entry["body"] = strings.Replace(entry["body"].(string), "Codex Review: Didn't find any major issues. Chef's kiss.", "Codex Review: Didn't find any major issues. Keep it up!", 1)
				}
			}
			return v
		}, want: 0, reason: "native-review-clean"},
		{name: "code_receipt_requires_exact_clean_declaration_at_line_start", edit: func(v []cgEntry) []cgEntry {
			for _, entry := range v {
				if entry["id"] == float64(6267776147) {
					entry["body"] = strings.Replace(entry["body"].(string), "Codex Review: Didn't find any major issues. Chef's kiss.", "Context: Codex Review: Didn't find any major issues. Chef's kiss.", 1)
				}
			}
			return v
		}, want: 1, reason: "native-receipt-type-unknown"},
		{name: "code_receipt_duplicate_declaration_lines_are_ambiguous", edit: func(v []cgEntry) []cgEntry {
			for _, entry := range v {
				if entry["id"] == float64(6267776147) {
					entry["body"] = strings.Replace(entry["body"].(string), "**Reviewed commit:**", "Codex Review: Found major issues.\n**Reviewed commit:**", 1)
				}
			}
			return v
		}, want: 1, reason: "native-receipt-ambiguous:code_review"},
		{name: "code_receipt_actual_bravo_clean_text_code_newest", edit: func(v []cgEntry) []cgEntry {
			for _, entry := range v {
				if entry["id"] == float64(6267776147) {
					entry["id"] = float64(6267807386)
					entry["body"] = strings.Replace(entry["body"].(string), "Codex Review: Didn't find any major issues. Chef's kiss.", "Codex Review: Didn't find any major issues. Bravo.", 1)
				}
			}
			return v
		}, want: 0, reason: "native-review-clean"},
		{name: "legacy_code_clean_text_code_newest", edit: func(v []cgEntry) []cgEntry {
			for _, entry := range v {
				if entry["id"] == float64(6267776147) {
					entry["id"] = float64(6267807386)
					entry["body"] = strings.Replace(entry["body"].(string), "Codex Review: Didn't find any major issues. Chef's kiss.", "Codex Review: Didn't find any major issues. :+1:", 1)
				}
			}
			return v
		}, want: 0, reason: "native-review-clean"},
		{name: "duplicate_current_code_receipt_is_ambiguous", edit: func(v []cgEntry) []cgEntry {
			for _, entry := range v {
				if entry["id"] == float64(6267776147) {
					duplicate := make(cgEntry, len(entry))
					for key, value := range entry {
						duplicate[key] = value
					}
					duplicate["id"] = float64(6267807386)
					return append(v, duplicate)
				}
			}
			return v
		}, want: 1, reason: "native-receipt-ambiguous:code_review"},
		{name: "code_receipt_with_security_type_cue_is_rejected", edit: func(v []cgEntry) []cgEntry {
			for _, entry := range v {
				if entry["id"] == float64(6267776147) {
					entry["body"] = strings.Replace(entry["body"].(string), "Codex Review:", "### Codex Security Review\nCodex Review:", 1)
				}
			}
			return v
		}, want: 1, reason: "native-receipt-type-ambiguous"},
		{name: "unknown_receipt_type_is_rejected", edit: func(v []cgEntry) []cgEntry {
			for _, entry := range v {
				if entry["id"] == float64(6267776147) {
					entry["body"] = strings.Replace(entry["body"].(string), "Codex Review:", "Code review completed:", 1)
				}
			}
			return v
		}, want: 1, reason: "native-receipt-type-unknown"},
		{name: "missing_security_receipt_is_rejected", edit: func(v []cgEntry) []cgEntry {
			filtered := v[:0]
			for _, entry := range v {
				if entry["id"] != float64(6266869543) && entry["id"] != float64(6267807385) {
					filtered = append(filtered, entry)
				}
			}
			return filtered
		}, want: 1, reason: "native-receipt-missing:security_review"},
		{name: "forged_code_receipt_identity_is_rejected", edit: func(v []cgEntry) []cgEntry {
			for _, entry := range v {
				if entry["id"] == float64(6267776147) {
					entry["user"] = map[string]any{"id": 42, "login": "chatgpt-codex-connector[bot]", "type": "Bot"}
				}
			}
			return v
		}, want: 1, reason: "native-receipt-missing:code_review"},
		{name: "forged_security_receipt_identity_is_rejected", edit: func(v []cgEntry) []cgEntry {
			for _, entry := range v {
				if entry["id"] == float64(6267807385) {
					entry["user"] = map[string]any{"id": 42, "login": "chatgpt-codex-connector[bot]", "type": "Bot"}
				}
			}
			return v
		}, want: 1, reason: "stale-review:"},
		{name: "forged_duplicate_receipt_does_not_override_valid_receipt", edit: func(v []cgEntry) []cgEntry {
			for _, entry := range v {
				if entry["id"] == float64(6267776147) {
					forged := make(cgEntry, len(entry))
					for key, value := range entry {
						forged[key] = value
					}
					forged["id"] = float64(6267807391)
					forged["user"] = map[string]any{"id": 42, "login": "chatgpt-codex-connector[bot]", "type": "Bot"}
					return append(v, forged)
				}
			}
			return v
		}, want: 0, reason: "native-review-clean"},
		{name: "stale_code_receipt_is_rejected", edit: func(v []cgEntry) []cgEntry {
			for _, entry := range v {
				if entry["id"] == float64(6267776147) {
					entry["body"] = strings.Replace(entry["body"].(string), "ab609cafbf", cgOther[:10], 1)
				}
			}
			return v
		}, want: 1, reason: "stale-review:"},
		{name: "security_receipt_must_have_its_own_clean_text", edit: func(v []cgEntry) []cgEntry {
			for _, entry := range v {
				if entry["id"] == float64(6267807385) {
					entry["body"] = strings.Replace(entry["body"].(string), "No security issues were found", "Security issues were found", 1)
				}
			}
			return v
		}, want: 1, reason: "native-receipt-not-clean:security_review"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var entries []cgEntry
			for _, entry := range clean {
				clone := make(cgEntry, len(entry))
				for key, value := range entry {
					clone[key] = value
				}
				entries = append(entries, clone)
			}
			if tc.edit != nil {
				entries = tc.edit(entries)
			}
			run(t, entries, tc.want, tc.reason)
		})
	}
}

func TestCodexNativeOldReviewObjectsDoNotBlockCurrentCleanReceipt(t *testing.T) {
	cgTools(t)
	for _, tc := range []struct {
		name  string
		head  string
		state string
		want  int
	}{
		{name: "old_comment_does_not_block", head: cgOther, state: "COMMENTED", want: 0},
		{name: "current_comment_still_blocks", head: cgHead, state: "COMMENTED", want: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			issues := cgWriteList(t, dir, "issues", cgNativeIssueComments(t, cgHead)...)
			reviews := cgWriteList(t, dir, "reviews", cgReview(7, tc.head, tc.state))
			comments := cgWriteList(t, dir, "comments")
			got := cgExec(t, cgScriptPath(t), nil, append(cgReadArgs(reviews, comments, cgHead, "false", "true"), "--issue-comments", issues)...)
			if got.code != tc.want {
				t.Fatalf("exit=%d want=%d: %s", got.code, tc.want, got.all())
			}
		})
	}
}

func TestCodexNativeRunningSecurityLegWaits(t *testing.T) {
	cgTools(t)
	const fixtureHead = "ab609cafbf8f421164100c626db93048f888655f"
	for _, tc := range []struct {
		name           string
		waitClosed     string
		unresolved     bool
		currentFinding bool
		malformed      bool
		stale          bool
		failed         bool
		wantCode       int
		wantText       string
	}{
		{name: "running_without_receipt_waits", waitClosed: "false", wantCode: 10, wantText: "pending:running"},
		{name: "running_without_receipt_at_window_end_fails", waitClosed: "true", wantCode: 1, wantText: "review-not-complete:running"},
		{name: "running_unresolved_still_fails", waitClosed: "false", unresolved: true, wantCode: 1, wantText: "finding:unresolved:security_review"},
		{name: "running_current_finding_still_fails", waitClosed: "false", currentFinding: true, wantCode: 1, wantText: "finding:"},
		{name: "running_malformed_summary_still_fails", waitClosed: "false", malformed: true, wantCode: 1, wantText: "review-not-complete:security_review:has_unresolved_findings"},
		{name: "running_stale_summary_still_fails", waitClosed: "false", stale: true, wantCode: 1, wantText: "stale-review:"},
		{name: "running_peer_failure_still_fails", waitClosed: "false", failed: true, wantCode: 1, wantText: "review-failed:code_review"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			entries := cgLoadNative1828Fixture(t, "codex-native-1828-clean.json")
			for _, entry := range entries {
				body, _ := entry["body"].(string)
				if strings.Contains(body, "<!-- codex-review-state:v1 ") {
					body = strings.Replace(body,
						`"security_review":{"commit_sha":"`+fixtureHead+`","has_any_findings":false,"has_unresolved_findings":false,"status":"completed"}`,
						`"security_review":{"commit_sha":"`+fixtureHead+`","has_any_findings":null,"has_unresolved_findings":null,"status":"running"}`, 1)
					body = strings.Replace(body, `"status":"completed"} -->`, `"status":"running"} -->`, 1)
					if tc.unresolved {
						body = strings.Replace(body, `"has_any_findings":null,"has_unresolved_findings":null,"status":"running"`, `"has_any_findings":true,"has_unresolved_findings":true,"status":"running"`, 1)
					}
					if tc.malformed {
						body = strings.Replace(body, `"has_any_findings":null,"has_unresolved_findings":null,"status":"running"`, `"has_any_findings":null,"has_unresolved_findings":"unknown","status":"running"`, 1)
					}
					if tc.stale {
						body = strings.Replace(body, `"security_review":{"commit_sha":"`+fixtureHead+`"`, `"security_review":{"commit_sha":"`+cgOther+`"`, 1)
					}
					if tc.failed {
						body = strings.Replace(body,
							`"code_review":{"commit_sha":"`+fixtureHead+`","has_any_findings":false,"has_unresolved_findings":false,"status":"completed"}`,
							`"code_review":{"commit_sha":"`+fixtureHead+`","has_any_findings":false,"has_unresolved_findings":false,"status":"failed"}`, 1)
					}
					entry["body"] = body
				}
			}
			kept := entries[:0]
			for _, entry := range entries {
				body, _ := entry["body"].(string)
				if strings.Contains(body, "Codex Security Review") || strings.Contains(body, "Security review completed.") {
					continue
				}
				kept = append(kept, entry)
			}
			dir := t.TempDir()
			var issueEntries []any
			for _, entry := range kept {
				issueEntries = append(issueEntries, entry)
			}
			issues := cgWriteList(t, dir, "issues", issueEntries...)
			reviews := cgWriteList(t, dir, "reviews")
			reviewComments := cgWriteList(t, dir, "review-comments")
			if tc.currentFinding {
				reviewComments = cgWriteList(t, dir, "review-comments", cgComment(77, fixtureHead, fixtureHead))
			}
			args := append(cgReadArgs(reviews, reviewComments, fixtureHead, "false", tc.waitClosed), "--issue-comments", issues)
			got := cgExec(t, cgScriptPath(t), nil, args...)
			if got.code != tc.wantCode || !strings.Contains(got.all(), tc.wantText) {
				t.Fatalf("exit=%d want=%d output=%s; want text %q", got.code, tc.wantCode, got.all(), tc.wantText)
			}
		})
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
			pull := strings.ReplaceAll(amPull(cgHead, false, 2, 0), `"comments":0`, `"comments":3`)
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
