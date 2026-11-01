package template_test

// Card t2012 — the security-null adapter of SPEC-CI-CODEX-REVIEW-001 (Amendment 8, AC-CCR-004a).
//
// F-security-null is the real shape observed on PR #1956 at head a755592815: the identity's state
// comment reports code_review {completed, false, false} and security_review {completed, commit_sha =
// head, has_any_findings null, has_unresolved_findings null}, a security marker completed on the same
// head, and one clean code receipt and one clean security receipt of the identity. The fixture file is
// built from the real objects (testdata/codex-native-security-null.json: the leading marker lines and
// the receipt lines only; ids, timestamps, author and performed_via_github_app are as observed).
//
// The pass row must exit 0 through the adapter. Every refusal row changes exactly one thing and must
// exit with the reason of its row; rows that share an existing reason are told apart by their fixture.
// "null is not false" stays the rule: no row here converts null to false.

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	snHead  = "a7555928155f1df3bcc3a1439670a97ce8dfebf1"
	snOther = "fedcba9876543210fedcba9876543210fedcba98"
	snApp   = "chatgpt-codex-connector"
)

// snFixture returns a deep copy of the stored real-shape fixture: state comment, code receipt,
// security receipt (in that order).
func snFixture(t *testing.T) []cgEntry {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(findProjectRootForMirrorTest(t), "internal/template/testdata/codex-native-security-null.json"))
	if err != nil {
		t.Fatal(err)
	}
	var v []cgEntry
	if err := json.Unmarshal(raw, &v); err != nil {
		t.Fatal(err)
	}
	if len(v) != 3 {
		t.Fatalf("fixture holds %d entries, want 3 (state, code receipt, security receipt)", len(v))
	}
	return v
}

func snClone(e cgEntry) cgEntry {
	b, _ := json.Marshal(e)
	var n cgEntry
	_ = json.Unmarshal(b, &n)
	return n
}

func snBody(e cgEntry) string { return e["body"].(string) }

// snRun writes the issue comments plus optional review and review-comment lists and runs the gate
// for head with --wait-closed closed.
func snRun(t *testing.T, issues []cgEntry, reviews, comments []any, closed string) cgResult {
	t.Helper()
	dir := t.TempDir()
	var items []any
	for _, e := range issues {
		items = append(items, e)
	}
	ip := cgWriteList(t, dir, "issues", items...)
	rp := cgWriteList(t, dir, "reviews", reviews...)
	cp := cgWriteList(t, dir, "comments", comments...)
	args := append(cgReadArgs(rp, cp, snHead, "false", closed), "--issue-comments", ip)
	return cgExec(t, cgScriptPath(t), nil, args...)
}

type snRow struct {
	name     string
	edit     func(v []cgEntry) []cgEntry
	reviews  func() []any
	comments func() []any
	want     int    // exit code with --wait-closed true
	reason   string // substring of the output
	wantOpen *int   // exit code with --wait-closed false, when it differs from want
	text     string // extra substring that must be present (pass row)
}

func snIntp(n int) *int { return &n }

func TestCodexSecurityNullAdapter(t *testing.T) {
	cgTools(t)

	stateBody := func(v []cgEntry) string { return snBody(v[0]) }
	setState := func(v []cgEntry, from, to string) []cgEntry {
		if !strings.Contains(stateBody(v), from) {
			t.Fatalf("fixture lacks %q", from)
		}
		v[0]["body"] = strings.Replace(stateBody(v), from, to, 1)
		return v
	}
	secLeg := `"security_review":{"commit_sha":"` + snHead + `","has_any_findings":null,"has_unresolved_findings":null,"status":"completed"}`
	rows := []snRow{
		{name: "P_real_1956_shape", want: 0, text: "native-review-clean"},
		{name: "N1a_one_null_only_any", edit: func(v []cgEntry) []cgEntry {
			return setState(v, secLeg, strings.Replace(secLeg, `"has_unresolved_findings":null`, `"has_unresolved_findings":false`, 1))
		}, want: 1, reason: "review-not-complete:security_review:has_any_findings"},
		{name: "N1b_one_null_only_unresolved", edit: func(v []cgEntry) []cgEntry {
			return setState(v, secLeg, strings.Replace(secLeg, `"has_any_findings":null`, `"has_any_findings":false`, 1))
		}, want: 1, reason: "review-not-complete:security_review:has_unresolved_findings"},
		{name: "N2_any_true_unresolved_false", edit: func(v []cgEntry) []cgEntry {
			return setState(v, secLeg, strings.Replace(strings.Replace(secLeg, `"has_any_findings":null`, `"has_any_findings":true`, 1), `"has_unresolved_findings":null`, `"has_unresolved_findings":false`, 1))
		}, want: 1, reason: "finding:reported:security_review"},
		{name: "N3_any_true_unresolved_true", edit: func(v []cgEntry) []cgEntry {
			return setState(v, secLeg, strings.Replace(strings.Replace(secLeg, `"has_any_findings":null`, `"has_any_findings":true`, 1), `"has_unresolved_findings":null`, `"has_unresolved_findings":true`, 1))
		}, want: 1, reason: "finding:unresolved:security_review"},
		{name: "N4a_key_any_absent", edit: func(v []cgEntry) []cgEntry {
			return setState(v, secLeg, strings.Replace(secLeg, `"has_any_findings":null,`, ``, 1))
		}, want: 1, reason: "review-not-complete:security_review:has_any_findings"},
		{name: "N4b_key_unresolved_absent", edit: func(v []cgEntry) []cgEntry {
			return setState(v, secLeg, strings.Replace(secLeg, `"has_unresolved_findings":null,`, ``, 1))
		}, want: 1, reason: "review-not-complete:security_review:has_unresolved_findings"},
		{name: "N5_security_commit_sha_null_1828_form", edit: func(v []cgEntry) []cgEntry {
			return setState(v, secLeg, strings.Replace(secLeg, `"commit_sha":"`+snHead+`"`, `"commit_sha":null`, 1))
		}, want: 1, reason: "review-not-complete:security_review:has_any_findings"},
		{name: "N6_security_leg_running", edit: func(v []cgEntry) []cgEntry {
			return setState(v, secLeg, strings.Replace(secLeg, `"status":"completed"`, `"status":"running"`, 1))
		}, want: 1, reason: "review-not-complete:running", wantOpen: snIntp(10)},
		{name: "N7_code_flag_null", edit: func(v []cgEntry) []cgEntry {
			return setState(v, `"code_review":{"commit_sha":"`+snHead+`","has_any_findings":false`, `"code_review":{"commit_sha":"`+snHead+`","has_any_findings":null`)
		}, want: 1, reason: "review-not-complete:code_review:has_any_findings"},
		{name: "N8_marker_running", edit: func(v []cgEntry) []cgEntry {
			return setState(v, `"repository":"modu-ai/moai-adk","status":"completed"`, `"repository":"modu-ai/moai-adk","status":"running"`)
		}, want: 1, reason: "review-not-complete:security_review:has_any_findings"},
		{name: "N9_marker_headSha_other", edit: func(v []cgEntry) []cgEntry {
			return setState(v, `"headSha":"`+snHead+`"`, `"headSha":"`+snOther+`"`)
		}, want: 1, reason: "stale-review:" + snOther[:10]},
		{name: "N10_security_receipt_absent", edit: func(v []cgEntry) []cgEntry { return v[:2] },
			want: 1, reason: "native-receipt-missing:security_review"},
		{name: "N11_security_receipt_not_clean", edit: func(v []cgEntry) []cgEntry {
			v[2]["body"] = strings.Replace(snBody(v[2]), "No security issues were found in this pull request.", "Security issues were found in this pull request.", 1)
			return v
		}, want: 1, reason: "native-receipt-not-clean:security_review"},
		{name: "N12_security_receipt_stale", edit: func(v []cgEntry) []cgEntry {
			v[2]["body"] = strings.Replace(snBody(v[2]), snHead[:10], snOther[:10], 1)
			return v
		}, want: 1, reason: "stale-review:" + snOther[:10]},
		{name: "N13_two_same_head_security_receipts", edit: func(v []cgEntry) []cgEntry {
			n := snClone(v[2])
			n["id"] = 9999999001
			n["created_at"] = "2026-11-01T17:40:00Z"
			n["updated_at"] = "2026-11-01T17:40:00Z"
			return append(v, n)
		}, want: 1, reason: "native-receipt-ambiguous:security_review"},
		{name: "N14a_security_receipt_forged_login", edit: func(v []cgEntry) []cgEntry {
			v[2]["user"].(map[string]any)["login"] = "chatgpt-codex-connector-fake[bot]"
			return v
		}, want: 1, reason: "native-receipt-missing:security_review"},
		{name: "N14b_security_receipt_forged_id", edit: func(v []cgEntry) []cgEntry {
			v[2]["user"].(map[string]any)["id"] = 199175423
			return v
		}, want: 1, reason: "native-receipt-missing:security_review"},
		{name: "N14c_security_receipt_type_user", edit: func(v []cgEntry) []cgEntry {
			v[2]["user"].(map[string]any)["type"] = "User"
			return v
		}, want: 1, reason: "native-receipt-missing:security_review"},
		{name: "N15_state_comment_forged_identity", edit: func(v []cgEntry) []cgEntry {
			v[0]["user"].(map[string]any)["id"] = 199175423
			return v
		}, want: 11, reason: "not-run:review-absent", wantOpen: snIntp(10)},
		{name: "N16_identity_review_comment_on_other_head", comments: func() []any {
			return []any{cgComment(7001, snOther, snOther)}
		}, want: 1, reason: "review-not-complete:security_review:has_any_findings"},
		{name: "N17_identity_commented_review_on_other_head", reviews: func() []any {
			return []any{cgReview(7002, snOther, "COMMENTED")}
		}, want: 1, reason: "review-not-complete:security_review:has_any_findings"},
		{name: "N18_state_comment_wrong_app_slug", edit: func(v []cgEntry) []cgEntry {
			v[0]["performed_via_github_app"].(map[string]any)["slug"] = "other-app"
			return v
		}, want: 1, reason: "review-not-complete:security_review:has_any_findings"},
		{name: "N18b_security_receipt_wrong_app_slug", edit: func(v []cgEntry) []cgEntry {
			v[2]["performed_via_github_app"].(map[string]any)["slug"] = "other-app"
			return v
		}, want: 1, reason: "review-not-complete:security_review:has_any_findings"},
		{name: "N19_second_marker_version_present", edit: func(v []cgEntry) []cgEntry {
			v[0]["body"] = snBody(v[0]) + "\n<!-- codex-security-review:v2 {\"status\":\"completed\"} -->"
			return v
		}, want: 1, reason: "review-not-complete:security_review:has_any_findings"},
		// R76 rows (leader extras). The string forms are regression guards; the duplicate-key and
		// duplicate-artifact rows are RED against the first adapter, which read the LAST duplicate key.
		{name: "R3a_string_null_any", edit: func(v []cgEntry) []cgEntry {
			return setState(v, `"has_any_findings":null`, `"has_any_findings":"null"`)
		}, want: 1, reason: "review-not-complete:security_review:has_any_findings"},
		{name: "R3b_string_null_unresolved", edit: func(v []cgEntry) []cgEntry {
			return setState(v, `"has_unresolved_findings":null`, `"has_unresolved_findings":"null"`)
		}, want: 1, reason: "review-not-complete:security_review:has_unresolved_findings"},
		{name: "R3c_string_false_security", edit: func(v []cgEntry) []cgEntry {
			return setState(v, secLeg, strings.Replace(strings.Replace(secLeg, `"has_any_findings":null`, `"has_any_findings":"false"`, 1), `"has_unresolved_findings":null`, `"has_unresolved_findings":"false"`, 1))
		}, want: 1, reason: "review-not-complete:security_review:has_any_findings"},
		{name: "R3d_string_false_code", edit: func(v []cgEntry) []cgEntry {
			return setState(v, `"code_review":{"commit_sha":"`+snHead+`","has_any_findings":false`, `"code_review":{"commit_sha":"`+snHead+`","has_any_findings":"false"`)
		}, want: 1, reason: "review-not-complete:code_review:has_any_findings"},
		{name: "R4a_duplicate_key_hides_true_any", edit: func(v []cgEntry) []cgEntry {
			return setState(v, `"security_review":{"commit_sha":"`+snHead+`","has_any_findings":null`, `"security_review":{"commit_sha":"`+snHead+`","has_any_findings":true,"has_any_findings":null`)
		}, want: 1, reason: "review-not-complete:security_review:has_any_findings"},
		{name: "R4b_duplicate_key_hides_true_unresolved", edit: func(v []cgEntry) []cgEntry {
			return setState(v, `"has_unresolved_findings":null`, `"has_unresolved_findings":true,"has_unresolved_findings":null`)
		}, want: 1, reason: "review-not-complete:security_review:has_unresolved_findings"},
		{name: "R4c_duplicate_security_leg_object", edit: func(v []cgEntry) []cgEntry {
			return setState(v, `"security_review":{`, `"security_review":{"commit_sha":"`+snHead+`","has_any_findings":true,"has_unresolved_findings":true,"status":"completed"},"security_review":{`)
		}, want: 1, reason: "review-not-complete:security_review:has_any_findings"},
		{name: "R4d_duplicate_commit_sha_hides_stale", edit: func(v []cgEntry) []cgEntry {
			return setState(v, `"security_review":{"commit_sha":"`+snHead+`"`, `"security_review":{"commit_sha":"`+snOther+`","commit_sha":"`+snHead+`"`)
		}, want: 1, reason: "review-not-complete:security_review:has_any_findings"},
		{name: "R4e_second_security_marker_same_version", edit: func(v []cgEntry) []cgEntry {
			v[0]["body"] = snBody(v[0]) + "\n<!-- codex-security-review:v1 {\"headSha\":\"" + snHead + "\",\"status\":\"completed\"} -->"
			return v
		}, want: 1, reason: "security-marker-duplicated"},
		{name: "R4f_second_state_marker_same_version", edit: func(v []cgEntry) []cgEntry {
			v[0]["body"] = snBody(v[0]) + "\n<!-- codex-review-state:v1 {\"code_review\":{\"status\":\"completed\"}} -->"
			return v
		}, want: 1, reason: "state-marker-duplicated"},
		{name: "R4g_second_state_comment", edit: func(v []cgEntry) []cgEntry {
			n := snClone(v[0])
			n["id"] = 9999999002
			n["created_at"] = "2026-11-01T17:41:00Z"
			n["updated_at"] = "2026-11-01T17:41:00Z"
			return append(v, n)
		}, want: 1, reason: "state-comment-duplicated"},
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			v := snFixture(t)
			if row.edit != nil {
				v = row.edit(v)
			}
			var reviews, comments []any
			if row.reviews != nil {
				reviews = row.reviews()
			}
			if row.comments != nil {
				comments = row.comments()
			}
			for _, closed := range []string{"true", "false"} {
				want := row.want
				if closed == "false" && row.wantOpen != nil {
					want = *row.wantOpen
				}
				got := snRun(t, snCopy(v), reviews, comments, closed)
				if got.code != want {
					t.Fatalf("closed=%s: exit=%d want=%d\n%s", closed, got.code, want, got.all())
				}
				if want == row.want && row.reason != "" && !strings.Contains(got.all(), row.reason) {
					t.Fatalf("closed=%s: output lacks %q\n%s", closed, row.reason, got.all())
				}
				if row.text != "" && !strings.Contains(got.all(), row.text) {
					t.Fatalf("closed=%s: output lacks %q\n%s", closed, row.text, got.all())
				}
			}
		})
	}
}

func base64Decode(s string) (string, error) {
	b, err := base64.StdEncoding.DecodeString(s)
	return string(b), err
}

func snCopy(v []cgEntry) []cgEntry {
	out := make([]cgEntry, 0, len(v))
	for _, e := range v {
		out = append(out, snClone(e))
	}
	return out
}

// TestCodexSecurityNullAdapterNamesItsPath: the pass reason distinguishes the adapter path from a
// boolean pass, and the printed summary keeps the null values as read (null is not converted).
func TestCodexSecurityNullAdapterNamesItsPath(t *testing.T) {
	cgTools(t)
	got := snRun(t, snFixture(t), nil, nil, "true")
	if got.code != 0 {
		t.Fatalf("exit=%d want 0\n%s", got.code, got.all())
	}
	if !strings.Contains(got.stdout, "security-null-adapter") {
		t.Fatalf("pass reason does not name the adapter path:\n%s", got.stdout)
	}
	// The summary markers are carried base64; decode and require the nulls survive.
	idx := strings.LastIndex(got.stdout, "markers ")
	if idx < 0 {
		t.Fatalf("no markers field in the pass line:\n%s", got.stdout)
	}
	b64 := strings.TrimSpace(got.stdout[idx+len("markers "):])
	raw, err := base64Decode(b64)
	if err != nil {
		t.Fatalf("markers field is not base64: %v\n%s", err, got.stdout)
	}
	if !strings.Contains(raw, `"security_review":{"status":"completed","commit_sha":"`+snHead+`","has_unresolved_findings":null}`) {
		t.Fatalf("the printed summary does not keep the security null:\n%s", raw)
	}
}

// TestCodexSecurityNullAdapterBooleanPathUnchanged: a boolean clean state still exits 0 WITHOUT the
// adapter path, so a verdict basis is distinguishable.
func TestCodexSecurityNullAdapterBooleanPathUnchanged(t *testing.T) {
	cgTools(t)
	v := snFixture(t)
	v[0]["body"] = strings.Replace(snBody(v[0]),
		`"security_review":{"commit_sha":"`+snHead+`","has_any_findings":null,"has_unresolved_findings":null`,
		`"security_review":{"commit_sha":"`+snHead+`","has_any_findings":false,"has_unresolved_findings":false`, 1)
	got := snRun(t, v, nil, nil, "true")
	if got.code != 0 || !strings.Contains(got.stdout, "native-review-clean") {
		t.Fatalf("boolean clean state: exit=%d\n%s", got.code, got.all())
	}
	if strings.Contains(got.stdout, "security-null-adapter") {
		t.Fatalf("a boolean pass must not name the adapter path:\n%s", got.stdout)
	}
}

// TestCodexSecurityNullAdapterCollectionKeepsAppField: the consumer's collection path (fetch-pages,
// used by auto-merge) hands the gate whole objects, so performed_via_github_app survives and
// condition 1's slug test is not vacuous there. Proven on the code path, not by inspection: a gh
// double answers one page and the output list keeps the field.
func TestCodexSecurityNullAdapterCollectionKeepsAppField(t *testing.T) {
	cgTools(t)
	dir := t.TempDir()
	bin := filepath.Join(dir, "bin")
	if err := os.MkdirAll(bin, 0o755); err != nil {
		t.Fatal(err)
	}
	page, err := os.ReadFile(filepath.Join(findProjectRootForMirrorTest(t), "internal/template/testdata/codex-native-security-null.json"))
	if err != nil {
		t.Fatal(err)
	}
	pagePath := filepath.Join(dir, "page.json")
	if err := os.WriteFile(pagePath, page, 0o600); err != nil {
		t.Fatal(err)
	}
	// fetch-pages treats a short page (< 100 entries) as the last one, so a single answer ends the loop.
	stub := "#!/bin/sh\ncat \"" + pagePath + "\"\n"
	if err := os.WriteFile(filepath.Join(bin, "gh"), []byte(stub), 0o755); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "issues.ndjson")
	env := []string{"PATH=" + bin + string(os.PathListSeparator) + os.Getenv("PATH")}
	got := cgExec(t, cgScriptPath(t), env, "fetch-pages", "--endpoint", "repos/modu-ai/moai-adk/issues/1956/comments", "--out", out, "--expect-count", "3")
	if got.code != 0 {
		t.Fatalf("fetch-pages exit=%d\n%s", got.code, got.all())
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if n := strings.Count(string(data), `"slug":"`+snApp+`"`); n != 3 {
		t.Fatalf("collected list keeps performed_via_github_app on %d of 3 comments:\n%s", n, data)
	}
}

// snStubGH writes a gh double that answers `gh api <endpoint>?per_page=N&page=P` from the given page
// bodies (an absent page answers an empty list) and returns the environment that puts it on PATH.
func snStubGH(t *testing.T, dir string, pages map[int]string) []string {
	t.Helper()
	bin := filepath.Join(dir, "bin")
	if err := os.MkdirAll(bin, 0o755); err != nil {
		t.Fatal(err)
	}
	var sh strings.Builder
	sh.WriteString("#!/bin/sh\ncase \"$2\" in\n")
	for n, body := range pages {
		p := filepath.Join(dir, "page"+string(rune('0'+n))+".json")
		if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
		sh.WriteString("*page=" + string(rune('0'+n)) + ") cat \"" + p + "\" ;;\n")
	}
	sh.WriteString("*) echo '[]' ;;\nesac\n")
	if err := os.WriteFile(filepath.Join(bin, "gh"), []byte(sh.String()), 0o755); err != nil {
		t.Fatal(err)
	}
	return []string{"PATH=" + bin + string(os.PathListSeparator) + os.Getenv("PATH")}
}

// TestCodexSecurityNullAdapterPaginationSeesPageTwo (R76-1): an identity finding that sits only on
// page 2 of the review-comment list must reach the gate, and the adapter must then refuse. Page 1 is a
// full page of other users' comments, so fetch-pages has to take the second page.
func TestCodexSecurityNullAdapterPaginationSeesPageTwo(t *testing.T) {
	cgTools(t)
	dir := t.TempDir()
	var p1 []map[string]any
	for i := 0; i < 100; i++ {
		p1 = append(p1, map[string]any{"id": 8000 + i, "user": map[string]any{"login": "someone", "id": 1, "type": "User"}, "commit_id": snHead, "original_commit_id": snHead, "path": "x.go", "body": "note"})
	}
	b1, _ := json.Marshal(p1)
	b2, _ := json.Marshal([]cgEntry{cgComment(7003, snOther, snOther)})
	env := snStubGH(t, dir, map[int]string{1: string(b1), 2: string(b2)})
	out := filepath.Join(dir, "comments.ndjson")
	got := cgExec(t, cgScriptPath(t), env, "fetch-pages", "--endpoint", "repos/modu-ai/moai-adk/pulls/1956/comments", "--out", out, "--expect-count", "101")
	if got.code != 0 {
		t.Fatalf("fetch-pages exit=%d\n%s", got.code, got.all())
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var comments []any
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		var e map[string]any
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			t.Fatalf("collected line is not JSON: %v", err)
		}
		comments = append(comments, e)
	}
	if len(comments) != 101 {
		t.Fatalf("collected %d comments, want 101 (both pages)", len(comments))
	}
	res := snRun(t, snFixture(t), nil, comments, "true")
	if res.code != 1 || strings.Contains(res.all(), "security-null-adapter") {
		t.Fatalf("a finding only on page 2 must refuse the adapter: exit=%d\n%s", res.code, res.all())
	}
}

// TestCodexSecurityNullAdapterApiFailureIsNeverAPass (R76-2): a failed or truncated read is a refusal
// with its own reason, never a pass, on both the collection path and the gate's own input check.
func TestCodexSecurityNullAdapterApiFailureIsNeverAPass(t *testing.T) {
	cgTools(t)
	t.Run("page_one_truncated_json", func(t *testing.T) {
		dir := t.TempDir()
		env := snStubGH(t, dir, map[int]string{1: `[{"id":1,"user":`})
		got := cgExec(t, cgScriptPath(t), env, "fetch-pages", "--endpoint", "repos/modu-ai/moai-adk/issues/1956/comments", "--out", filepath.Join(dir, "o"))
		if got.code == 0 || !strings.Contains(got.all(), "page-invalid:1") {
			t.Fatalf("truncated page: exit=%d\n%s", got.code, got.all())
		}
	})
	t.Run("gh_exits_nonzero", func(t *testing.T) {
		dir := t.TempDir()
		bin := filepath.Join(dir, "bin")
		if err := os.MkdirAll(bin, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(bin, "gh"), []byte("#!/bin/sh\nexit 1\n"), 0o755); err != nil {
			t.Fatal(err)
		}
		env := []string{"PATH=" + bin + string(os.PathListSeparator) + os.Getenv("PATH")}
		got := cgExec(t, cgScriptPath(t), env, "fetch-pages", "--endpoint", "repos/modu-ai/moai-adk/issues/1956/comments", "--out", filepath.Join(dir, "o"))
		if got.code == 0 || !strings.Contains(got.all(), "page-error:1") {
			t.Fatalf("failed gh: exit=%d\n%s", got.code, got.all())
		}
	})
	t.Run("gate_input_truncated_issue_comments", func(t *testing.T) {
		dir := t.TempDir()
		rp := cgWriteList(t, dir, "reviews")
		cp := cgWriteList(t, dir, "comments")
		ip := cgWriteList(t, dir, "issues", `{"id":1,"user":`)
		args := append(cgReadArgs(rp, cp, snHead, "false", "true"), "--issue-comments", ip)
		got := cgExec(t, cgScriptPath(t), nil, args...)
		if got.code != 1 || !strings.Contains(got.all(), "input-invalid") {
			t.Fatalf("truncated gate input: exit=%d\n%s", got.code, got.all())
		}
	})
}
