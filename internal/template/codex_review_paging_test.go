package template_test

// SPEC-CI-CODEX-REVIEW-001 M2 — complete paging (fetch-pages) and the tamper guard (gate-files-changed).
//
// The stub `gh` written here serves fixture pages by the `page` parameter, can fail a chosen
// page, and also serves the pull request object, `gh pr view`, `gh pr checks` and `gh pr merge`
// by sequence, so the M3 consumer tests reuse it unchanged. Every call is recorded in one
// ordered log (CALL_LOG) that the stub `sleep` and the stub gate script of M3 also write to.

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const cgGhStub = `#!/bin/sh
set -u
if [ -n "${CALL_LOG:-}" ]; then echo "gh $*" >> "$CALL_LOG"; fi
seq_pick() {
  n=0
  if [ -f "$2" ]; then n=$(cat "$2"); fi
  n=$((n + 1))
  echo "$n" > "$2"
  total=$(wc -l < "$1" | tr -d ' ')
  if [ "$total" -eq 0 ]; then return 1; fi
  if [ "$n" -gt "$total" ]; then n=$total; fi
  sed -n "${n}p" "$1"
}
case "${1:-}" in
  api)
    ep="${2:-}"
    path="${ep%%\?*}"
    query=""
    case "$ep" in *\?*) query="${ep#*\?}" ;; esac
    case "$path" in
      repos/*/*/pulls/*/files | repos/*/*/pulls/*/reviews | repos/*/*/pulls/*/comments)
        kind="${path##*/}"
        page="${query##*page=}"
        if [ "${FAIL_PAGE:-}" = "$kind:$page" ] || [ "${FAIL_KIND:-}" = "$kind" ]; then
          echo "stub gh: forced failure for $kind page $page" >&2
          exit 1
        fi
        # A list can change between two collections: page-1 reads are counted per list, and the pages of
        # generation N (if $FIXTURES/$kind/gN/page-P.json exists) win over the base pages.
        if [ "$page" = "1" ]; then
          g=0
          if [ -f "$FIXTURES/$kind.gen" ]; then g=$(cat "$FIXTURES/$kind.gen"); fi
          g=$((g + 1))
          echo "$g" > "$FIXTURES/$kind.gen"
        fi
        g=0
        if [ -f "$FIXTURES/$kind.gen" ]; then g=$(cat "$FIXTURES/$kind.gen"); fi
        if [ -f "$FIXTURES/$kind/g$g/page-$page.json" ]; then cat "$FIXTURES/$kind/g$g/page-$page.json"
        elif [ -f "$FIXTURES/$kind/page-$page.json" ]; then cat "$FIXTURES/$kind/page-$page.json"
        else echo '[]'; fi
        ;;
      repos/*/*/pulls/*)
        f="$(seq_pick "$FIXTURES/pull-seq" "$FIXTURES/pull.count")" || { echo "stub gh: no pull fixture" >&2; exit 1; }
        if [ "$f" = "FAIL" ]; then echo "stub gh: pull object read failed" >&2; exit 1; fi
        cat "$FIXTURES/$f"
        ;;
      *) echo "stub gh: unexpected endpoint $ep" >&2; exit 99 ;;
    esac
    ;;
  pr)
    case "${2:-}" in
      view) seq_pick "$FIXTURES/view-seq" "$FIXTURES/view.count" ;;
      checks) cat "$FIXTURES/checks.json"; exit "${CHECKS_EXIT:-0}" ;;
      merge) exit "${MERGE_EXIT:-0}" ;;
      *) echo "stub gh: unexpected pr subcommand $*" >&2; exit 99 ;;
    esac
    ;;
  *) echo "stub gh: unexpected command $*" >&2; exit 99 ;;
esac
`

const cgSleepStub = `#!/bin/sh
if [ -n "${CALL_LOG:-}" ]; then echo "sleep $*" >> "$CALL_LOG"; fi
exit 0
`

// cgEnvT is one test's offline world: a stub bin directory first on PATH and a fixtures directory.
type cgEnvT struct {
	t        *testing.T
	dir      string
	bin      string
	fixtures string
	callLog  string
}

func newCGEnv(t *testing.T) *cgEnvT {
	t.Helper()
	cgTools(t)
	dir := t.TempDir()
	e := &cgEnvT{t: t, dir: dir, bin: filepath.Join(dir, "bin"), fixtures: filepath.Join(dir, "fixtures"), callLog: filepath.Join(dir, "calls.log")}
	for _, d := range []string{e.bin, e.fixtures} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	e.write(filepath.Join(e.bin, "gh"), cgGhStub, 0o755)
	e.write(filepath.Join(e.bin, "sleep"), cgSleepStub, 0o755)
	e.write(e.callLog, "", 0o600)
	return e
}

func (e *cgEnvT) write(path, body string, mode os.FileMode) {
	e.t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		e.t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), mode); err != nil {
		e.t.Fatalf("write %s: %v", path, err)
	}
}

// env is the child environment: the stub bin directory comes first on PATH.
func (e *cgEnvT) env(extra ...string) []string {
	base := []string{
		"PATH=" + e.bin + string(os.PathListSeparator) + os.Getenv("PATH"),
		"FIXTURES=" + e.fixtures, "CALL_LOG=" + e.callLog,
	}
	return append(base, extra...)
}

// setPages writes the entries of one list as pages of 100.
func (e *cgEnvT) setPages(kind string, entries []any) {
	e.t.Helper()
	for p, i := 1, 0; i < len(entries); p, i = p+1, i+100 {
		end := i + 100
		if end > len(entries) {
			end = len(entries)
		}
		raw, err := json.Marshal(entries[i:end])
		if err != nil {
			e.t.Fatal(err)
		}
		e.write(filepath.Join(e.fixtures, kind, fmt.Sprintf("page-%d.json", p)), string(raw), 0o600)
	}
}

func (e *cgEnvT) setRawPage(kind string, page int, raw string) {
	e.t.Helper()
	e.write(filepath.Join(e.fixtures, kind, fmt.Sprintf("page-%d.json", page)), raw, 0o600)
}

func (e *cgEnvT) calls() []string {
	e.t.Helper()
	raw, err := os.ReadFile(e.callLog)
	if err != nil {
		e.t.Fatal(err)
	}
	var out []string
	for _, l := range strings.Split(string(raw), "\n") {
		if strings.TrimSpace(l) != "" {
			out = append(out, l)
		}
	}
	return out
}

// cgFile builds a changed-file entry as the pull-request files listing returns it (research E31).
func cgFile(name, status, ref string, opts ...cgOpt) cgEntry {
	e := cgEntry{
		"filename": name, "status": status,
		"contents_url": "https://api.github.com/repos/modu-ai/moai-adk/contents/" + url.PathEscape(name) + "?ref=" + ref,
	}
	for _, o := range opts {
		o(e)
	}
	return e
}

// cgBenignFiles returns n unrelated changed files bound to ref.
func cgBenignFiles(n int, ref string) []any {
	out := make([]any, 0, n)
	for i := 1; i <= n; i++ {
		out = append(out, cgFile(fmt.Sprintf("docs/f%04d.md", i), "modified", ref))
	}
	return out
}

const cgEndpointPrefix = "repos/modu-ai/moai-adk/pulls/7/"

func (e *cgEnvT) fetch(kind, out string, expect string) cgResult {
	e.t.Helper()
	args := []string{"fetch-pages", "--endpoint", cgEndpointPrefix + kind, "--out", out}
	if expect != "" {
		args = append(args, "--expect-count", expect)
	}
	return cgExec(e.t, cgScriptPath(e.t), e.env(), args...)
}

func (e *cgEnvT) gateFiles(file, head string) cgResult {
	e.t.Helper()
	args := []string{"gate-files-changed", "--file", file}
	if head != "" {
		args = append(args, "--head-sha", head)
	}
	return cgExec(e.t, cgScriptPath(e.t), e.env(), args...)
}

func cgExpect(t *testing.T, res cgResult, code int, text string) {
	t.Helper()
	if res.code != code {
		t.Fatalf("exit %d, want %d\nstdout: %s\nstderr: %s", res.code, code, res.stdout, res.stderr)
	}
	if text != "" && !strings.Contains(res.all(), text) {
		t.Fatalf("output lacks %q\nstdout: %s\nstderr: %s", text, res.stdout, res.stderr)
	}
}

func cgLineCount(t *testing.T, path string) int {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, l := range strings.Split(string(raw), "\n") {
		if strings.TrimSpace(l) != "" {
			n++
		}
	}
	return n
}

func cgMustNotExist(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Fatalf("%s exists; a refusal must remove the output list", path)
	}
}

type cgSub struct {
	name string
	run  func(t *testing.T, e *cgEnvT)
}

func cgRunSubs(t *testing.T, subs []cgSub) {
	t.Helper()
	for _, s := range subs {
		t.Run(s.name, func(t *testing.T) {
			t.Parallel()
			s.run(t, newCGEnv(t))
		})
	}
}

func cgReviewList(n int, author cgOpt) []any {
	out := make([]any, 0, n)
	for i := 1; i <= n; i++ {
		out = append(out, cgReview(i, cgHead, "COMMENTED", author))
	}
	return out
}

// AC-CCR-010 (REQ-CCR-010) — every page is collected, count-checked where a count exists, and any doubt refuses.
// Mutants: page 1 only; the exit status of gh ignored; no count comparison; no 3000 ceiling; no empty-page
// rule at an exact multiple of 100; a repeated entry accepted.
func TestCodexGateFetchPagesCompleteness(t *testing.T) {
	t.Parallel()
	gate := func(t *testing.T, e *cgEnvT, n int, workflowPage int) {
		// n files, the (n)th being the workflow when workflowPage > 0.
		files := cgBenignFiles(n, cgHead)
		if workflowPage > 0 {
			files[n-1] = cgFile(".github/workflows/evil.yml", "added", cgHead)
		}
		e.setPages("files", files)
		out := filepath.Join(e.dir, "files.ndjson")
		cgExpect(t, e.fetch("files", out, fmt.Sprint(n)), 0, fmt.Sprintf("fetched:%d count-proof:yes", n))
		res := e.gateFiles(out, cgHead)
		if workflowPage > 0 {
			cgExpect(t, res, 1, "gate-path-changed:.github/workflows/evil.yml")
		} else {
			cgExpect(t, res, 0, "")
		}
	}
	cgRunSubs(t, []cgSub{
		{"single_page_clear", func(t *testing.T, e *cgEnvT) {
			e.setPages("files", cgBenignFiles(5, cgHead))
			out := filepath.Join(e.dir, "files.ndjson")
			cgExpect(t, e.fetch("files", out, "5"), 0, "fetched:5 count-proof:yes")
			if n := cgLineCount(t, out); n != 5 {
				t.Fatalf("out holds %d lines, want 5", n)
			}
		}},
		{"exact_200_ends_on_empty_page", func(t *testing.T, e *cgEnvT) {
			e.setPages("files", cgBenignFiles(200, cgHead))
			out := filepath.Join(e.dir, "files.ndjson")
			cgExpect(t, e.fetch("files", out, "200"), 0, "fetched:200 count-proof:yes")
			if n := cgLineCount(t, out); n != 200 {
				t.Fatalf("out holds %d lines, want 200", n)
			}
		}},
		{"entries_101_gate_file_on_page_2_refused", func(t *testing.T, e *cgEnvT) { gate(t, e, 101, 2) }},
		{"entries_301_gate_file_on_page_4_refused", func(t *testing.T, e *cgEnvT) { gate(t, e, 301, 4) }},
		{"entries_350_unrelated_clear", func(t *testing.T, e *cgEnvT) { gate(t, e, 350, 0) }},
		{"page_2_error_refuses_and_removes_out", func(t *testing.T, e *cgEnvT) {
			e.setPages("files", cgBenignFiles(150, cgHead))
			out := filepath.Join(e.dir, "files.ndjson")
			res := cgExec(t, cgScriptPath(t), e.env("FAIL_PAGE=files:2"), "fetch-pages", "--endpoint", cgEndpointPrefix+"files", "--out", out, "--expect-count", "150")
			cgExpect(t, res, 1, "page-error:2")
			cgMustNotExist(t, out)
		}},
		{"page_invalid_json_refuses", func(t *testing.T, e *cgEnvT) {
			e.setRawPage("files", 1, "this is not json")
			out := filepath.Join(e.dir, "files.ndjson")
			cgExpect(t, e.fetch("files", out, ""), 1, "page-invalid:1")
			cgMustNotExist(t, out)
		}},
		{"fewer_entries_than_expected_refused", func(t *testing.T, e *cgEnvT) {
			e.setPages("files", cgBenignFiles(9, cgHead))
			out := filepath.Join(e.dir, "files.ndjson")
			cgExpect(t, e.fetch("files", out, "10"), 1, "count-mismatch:9:10")
			cgMustNotExist(t, out)
		}},
		{"more_entries_than_expected_refused", func(t *testing.T, e *cgEnvT) {
			e.setPages("files", cgBenignFiles(9, cgHead))
			out := filepath.Join(e.dir, "files.ndjson")
			cgExpect(t, e.fetch("files", out, "8"), 1, "count-mismatch:9:8")
			cgMustNotExist(t, out)
		}},
		{"expect_count_not_integer_refused", func(t *testing.T, e *cgEnvT) {
			e.setPages("files", cgBenignFiles(3, cgHead))
			out := filepath.Join(e.dir, "files.ndjson")
			cgExpect(t, e.fetch("files", out, "abc"), 1, "expect-count-unusable")
			cgMustNotExist(t, out)
		}},
		{"ceiling_3000_refused", func(t *testing.T, e *cgEnvT) {
			e.setPages("files", cgBenignFiles(3000, cgHead))
			out := filepath.Join(e.dir, "files.ndjson")
			cgExpect(t, e.fetch("files", out, "3000"), 1, "page-ceiling")
			cgMustNotExist(t, out)
		}},
		{"reviews_without_count_proof_three_pages", func(t *testing.T, e *cgEnvT) {
			e.setPages("reviews", cgReviewList(250, cgRabbit))
			out := filepath.Join(e.dir, "reviews.ndjson")
			cgExpect(t, e.fetch("reviews", out, ""), 0, "fetched:250 count-proof:none")
			if n := cgLineCount(t, out); n != 250 {
				t.Fatalf("out holds %d lines, want 250", n)
			}
		}},
		{"review_comments_count_mismatch_refused", func(t *testing.T, e *cgEnvT) {
			var cs []any
			for i := 1; i <= 6; i++ {
				cs = append(cs, cgComment(i, cgHead, cgHead, cgHuman))
			}
			e.setPages("comments", cs)
			out := filepath.Join(e.dir, "comments.ndjson")
			cgExpect(t, e.fetch("comments", out, "5"), 1, "count-mismatch:6:5")
			cgMustNotExist(t, out)
		}},
		{"codex_review_on_page_2_is_read", func(t *testing.T, e *cgEnvT) {
			reviews := cgReviewList(100, cgRabbit)
			reviews = append(reviews, cgReview(101, cgHead, "APPROVED"))
			e.setPages("reviews", reviews)
			rp := filepath.Join(e.dir, "reviews.ndjson")
			cp := filepath.Join(e.dir, "comments.ndjson")
			cgExpect(t, e.fetch("reviews", rp, ""), 0, "fetched:101")
			cgExpect(t, e.fetch("comments", cp, "0"), 0, "fetched:0")
			res := cgExec(t, cgScriptPath(t), e.env(), cgReadArgs(rp, cp, cgHead, "false", "false")...)
			cgExpect(t, res, 0, "native-review-clean")
		}},
		{"finding_comment_on_page_3_refused", func(t *testing.T, e *cgEnvT) {
			e.setPages("reviews", []any{cgReview(1, cgHead, "APPROVED")})
			var cs []any
			for i := 1; i <= 249; i++ {
				cs = append(cs, cgComment(i, cgHead, cgHead, cgHuman))
			}
			cs = append(cs, cgComment(250, cgHead, cgHead))
			e.setPages("comments", cs)
			rp := filepath.Join(e.dir, "reviews.ndjson")
			cp := filepath.Join(e.dir, "comments.ndjson")
			cgExpect(t, e.fetch("reviews", rp, ""), 0, "fetched:1")
			cgExpect(t, e.fetch("comments", cp, "250"), 0, "fetched:250")
			res := cgExec(t, cgScriptPath(t), e.env(), cgReadArgs(rp, cp, cgHead, "false", "false")...)
			cgExpect(t, res, 1, "finding:250")
		}},
		{"duplicate_review_id_across_pages_refused", func(t *testing.T, e *cgEnvT) {
			// A list that shifted by one between two page reads: id 100 appears on both pages.
			var all []any
			for i := 1; i <= 100; i++ {
				all = append(all, cgReview(i, cgHead, "COMMENTED", cgRabbit))
			}
			e.setPages("reviews", all)
			var second []any
			for i := 100; i <= 149; i++ {
				second = append(second, cgReview(i, cgHead, "COMMENTED", cgRabbit))
			}
			raw, _ := json.Marshal(second)
			e.setRawPage("reviews", 2, string(raw))
			out := filepath.Join(e.dir, "reviews.ndjson")
			cgExpect(t, e.fetch("reviews", out, ""), 1, "page-duplicate:100")
			cgMustNotExist(t, out)
		}},
		{"duplicate_filename_across_pages_refused", func(t *testing.T, e *cgEnvT) {
			// The count still equals the expected count; only the repeated entry reveals the shift.
			first := cgBenignFiles(100, cgHead)
			e.setPages("files", first)
			var second []any
			for i := 100; i <= 149; i++ {
				second = append(second, cgFile(fmt.Sprintf("docs/f%04d.md", i), "modified", cgHead))
			}
			raw, _ := json.Marshal(second)
			e.setRawPage("files", 2, string(raw))
			out := filepath.Join(e.dir, "files.ndjson")
			cgExpect(t, e.fetch("files", out, "150"), 1, "page-duplicate:docs/f0100.md")
			cgMustNotExist(t, out)
		}},
		{"entry_without_id_or_filename_refused", func(t *testing.T, e *cgEnvT) {
			e.setRawPage("reviews", 1, `[{"state":"APPROVED"}]`)
			out := filepath.Join(e.dir, "reviews.ndjson")
			cgExpect(t, e.fetch("reviews", out, ""), 1, "page-invalid:1")
			cgMustNotExist(t, out)
		}},
	})
}

// AC-CCR-006 (script rows) — fetch-pages refuses every endpoint outside the three listings, before any gh call.
// Mutant: a fetch-pages without the endpoint check.
func TestCodexGateFetchPagesRefusesDisallowedEndpoint(t *testing.T) {
	t.Parallel()
	refused := func(endpoint string) func(t *testing.T, e *cgEnvT) {
		return func(t *testing.T, e *cgEnvT) {
			out := filepath.Join(e.dir, "x.ndjson")
			res := cgExec(t, cgScriptPath(t), e.env(), "fetch-pages", "--endpoint", endpoint, "--out", out)
			cgExpect(t, res, 1, "endpoint-not-allowed:"+endpoint)
			if calls := e.calls(); len(calls) != 0 {
				t.Fatalf("a refused endpoint reached gh: %v", calls)
			}
			cgMustNotExist(t, out)
		}
	}
	cgRunSubs(t, []cgSub{
		{"issue_comments_endpoint_refused", refused("repos/modu-ai/moai-adk/issues/7/comments")},
		{"check_runs_endpoint_refused", refused("repos/modu-ai/moai-adk/commits/" + cgHead + "/check-runs")},
		{"commit_status_endpoint_refused", refused("repos/modu-ai/moai-adk/commits/" + cgHead + "/status")},
		{"path_with_extra_segment_refused", func(t *testing.T, e *cgEnvT) {
			refused("repos/modu-ai/moai-adk/pulls/7/files/extra")(t, e)
			refused("repos/modu-ai/moai-adk/pulls/7")(t, e)
		}},
		{"files_reviews_comments_endpoints_accepted", func(t *testing.T, e *cgEnvT) {
			e.setPages("files", cgBenignFiles(2, cgHead))
			e.setPages("reviews", []any{cgReview(1, cgHead, "APPROVED")})
			e.setPages("comments", []any{cgComment(1, cgHead, cgHead, cgHuman)})
			for _, kind := range []string{"files", "reviews", "comments"} {
				out := filepath.Join(e.dir, kind+".ndjson")
				cgExpect(t, e.fetch(kind, out, ""), 0, "fetched:")
			}
		}},
	})
}

// AC-CCR-009 (REQ-CCR-009) — the tamper guard: any gate path, an unusable list, or an entry not bound to the head refuses.
// Mutants: a guard that matches only auto-merge.yml and the script; previous_filename ignored; a case-sensitive
// match; an empty list read as clear; no 3000 cap; a list read by pull request number with no entry binding.
func TestCodexGateFilesChangedRefusals(t *testing.T) {
	t.Parallel()
	H, O := cgHead, cgOther
	type row struct {
		name     string
		entries  []any
		noFile   bool
		noHead   bool
		wantCode int
		wantText string
	}
	benign := cgBenignFiles(2, H)
	with := func(extra ...any) []any { return append(append([]any{}, benign...), extra...) }
	rows := []row{
		{name: "workflow_auto_merge_edited", entries: with(cgFile(".github/workflows/auto-merge.yml", "modified", H)), wantCode: 1, wantText: "gate-path-changed:.github/workflows/auto-merge.yml"},
		{name: "workflow_added_any_name", entries: with(cgFile(".github/workflows/zzz-new.yml", "added", H)), wantCode: 1, wantText: "gate-path-changed:.github/workflows/zzz-new.yml"},
		{name: "workflow_edited_other", entries: with(cgFile(".github/workflows/ci.yml", "modified", H)), wantCode: 1, wantText: "gate-path-changed:.github/workflows/ci.yml"},
		{name: "workflow_deleted", entries: with(cgFile(".github/workflows/old.yml", "removed", O)), wantCode: 1, wantText: "gate-path-changed:.github/workflows/old.yml"},
		{name: "workflow_case_variant_prefix", entries: with(cgFile(".GitHub/Workflows/ci.yml", "modified", H)), wantCode: 1, wantText: "gate-path-changed:"},
		{name: "rename_into_workflows", entries: with(cgFile(".github/workflows/new-name.yml", "renamed", H, cgSet("previous_filename", "docs/was.md"))), wantCode: 1, wantText: "gate-path-changed:"},
		{name: "rename_from_workflows", entries: with(cgFile("docs/now.md", "renamed", H, cgSet("previous_filename", ".github/workflows/was.yml"))), wantCode: 1, wantText: "gate-path-changed:.github/workflows/was.yml"},
		{name: "gate_script_edited", entries: with(cgFile("scripts/ci/codex-review-gate.sh", "modified", H)), wantCode: 1, wantText: "gate-path-changed:scripts/ci/codex-review-gate.sh"},
		{name: "unrelated_files_clear", entries: benign, wantCode: 0},
		{name: "docs_and_go_files_clear", entries: []any{cgFile("README.md", "modified", H), cgFile("internal/x/y.go", "added", H), cgFile(".github/CODEOWNERS", "modified", H), cgFile("scripts/ci/retry-flaky.sh", "modified", H)}, wantCode: 0},
		{name: "empty_list_refused", entries: nil, wantCode: 1, wantText: "file-list-unusable"},
		{name: "missing_file_refused", noFile: true, wantCode: 1, wantText: "file-list-unusable"},
		{name: "entry_without_filename_refused", entries: with(cgEntry{"status": "modified"}), wantCode: 1, wantText: "file-entry-invalid"},
		{name: "list_of_3000_refused", entries: cgBenignFiles(3000, H), wantCode: 1, wantText: "file-list-at-api-ceiling"},
		{name: "list_of_2999_unrelated_clear", entries: cgBenignFiles(2999, H), wantCode: 0},
		{name: "non_removed_entry_ref_other_commit_refused", entries: with(cgFile("docs/other.md", "modified", O)), wantCode: 1, wantText: "file-entry-unbound:docs/other.md"},
		{name: "non_removed_entry_without_contents_url_refused", entries: with(cgFile("docs/nourl.md", "modified", H, cgDrop("contents_url"))), wantCode: 1, wantText: "file-entry-unbound:docs/nourl.md"},
		{name: "entries_bound_to_head_clear", entries: with(cgFile("docs/a.md", "added", H), cgFile("docs/b.md", "renamed", H, cgSet("previous_filename", "docs/c.md"))), wantCode: 0},
		{name: "removed_entry_ref_not_checked_clear", entries: with(cgFile("docs/gone.md", "removed", O)), wantCode: 0},
		{name: "ref_case_insensitive_equal_clear", entries: with(cgFile("docs/up.md", "modified", strings.ToUpper(H))), wantCode: 0},
		{name: "head_sha_missing_refused", entries: benign, noHead: true, wantCode: 1, wantText: "head-sha-unusable"},
	}
	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			t.Parallel()
			e := newCGEnv(t)
			file := filepath.Join(e.dir, "files.ndjson")
			if !r.noFile {
				cgWriteList(t, e.dir, "files.ndjson", r.entries...)
			}
			head := H
			if r.noHead {
				head = ""
			}
			cgExpect(t, e.gateFiles(file, head), r.wantCode, r.wantText)
		})
	}
}
