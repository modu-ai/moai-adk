//! PR body for a claim (Issue #2007, 06-C; REQ-06C-07..11, 13).
//!
//! `pr-body-render` fixes the evidence a PR is judged by: `Refs #<n>`, the Issue URL,
//! logical key, SPEC reference, exact base and candidate SHAs, epoch, the claim token, the
//! checks from VERIFY receipts (05-C #2003), the independent review slot and the not_run list.
//! It never writes a closing keyword: GitHub reads them only on PRs into the default branch,
//! and the leader judges completion (REQ-06C-07/08). The layout follows
//! `.github/pull_request_template.md`; that template is applied by GitHub only on the default
//! branch, so `v2026.11.0` PRs pass this output to `gh pr create --body-file`.
//!
//! The body ends with an operation marker in the form `forge::outbox::parse_op_markers`
//! reads (`<!-- moai-op pr_create:<repository_id>:<task_key>:<epoch> -->`), the lookup key for
//! 06-B's reconcile (REQ-06C-09). 06-B (#2006) has not landed, so no effect intent is recorded
//! and reconcile is reported `not_run`. The `pr_ready` state is returned as a
//! `record_update_request` for 06-A's owner; nothing here promotes a record (REQ-06C-11).
use crate::claim::{claim_token, ClaimRecord};
use crate::digest;
use crate::run::Receipt;
use crate::verify::CheckResult;
use crate::worktree::{is_sha, RECORD_UPDATE_OWNER};
use serde_json::{json, Value};

type Result<T> = std::result::Result<T, String>;

/// Every GitHub closing keyword; none may appear in a rendered body.
const CLOSING_KEYWORDS: &[&str] = &[
    "close", "closes", "closed", "fix", "fixes", "fixed", "resolve", "resolves", "resolved",
];

pub struct RenderInput<'a> {
    pub base_branch: &'a str,
    pub default_branch: &'a str,
    pub base_sha: &'a str,
    pub candidate_sha: &'a str,
    pub allow_close: bool,
    pub not_run: &'a [String],
    pub review_link: Option<&'a str>,
}

/// Closing keywords found in `text`, as ASCII words (any case).
pub fn closing_keywords(text: &str) -> Vec<String> {
    text.split(|c: char| !c.is_ascii_alphabetic())
        .map(|w| w.to_ascii_lowercase())
        .filter(|w| CLOSING_KEYWORDS.contains(&w.as_str()))
        .collect()
}

fn cell(s: &str) -> String {
    s.replace('|', "\\|").replace('\n', " ")
}

pub fn operation_id(rec: &ClaimRecord) -> String {
    format!(
        "pr_create:{}:{}:{}",
        rec.task_ref.repository_id, rec.task_key, rec.epoch
    )
}

pub fn render(rec: &ClaimRecord, receipts: &[Receipt], input: &RenderInput) -> Result<Value> {
    for (name, sha) in [
        ("base_sha", input.base_sha),
        ("candidate_sha", input.candidate_sha),
    ] {
        if !is_sha(sha) {
            return Err(format!("invalid {name}: must be 40 lowercase hex"));
        }
    }
    let t = &rec.task_ref;
    let n = t.issue_number;
    let url = format!("https://github.com/{}/{}/issues/{n}", t.owner, t.repo);
    let token = rec
        .comment_id
        .map(|id| format!("`{}`", claim_token(n, id)))
        .unwrap_or_else(|| "없음 (claim 댓글 id가 record에 없다)".into());
    let spec_digest = rec.spec_ref.digest.as_deref().unwrap_or("null");

    let mut rows = Vec::new();
    let mut not_run: Vec<String> = Vec::new();
    for r in receipts {
        for c in &r.checks {
            let exit = match (c.result, c.exit_code) {
                (CheckResult::NotRun, _) => "not_run".to_string(),
                (_, Some(code)) => code.to_string(),
                (_, None) => "none".to_string(),
            };
            let result = serde_json::to_value(c.result)
                .ok()
                .and_then(|v| v.as_str().map(String::from))
                .unwrap_or_default();
            let tests = match (c.tests_passed, c.tests_failed) {
                (Some(p), Some(f)) => format!(" · {p} passed / {f} failed"),
                _ => String::new(),
            };
            rows.push(format!(
                "| `{}` | {exit} | {result}{tests} (receipt {}) |",
                cell(&c.command),
                cell(&r.attempt_id)
            ));
            if c.result == CheckResult::NotRun {
                not_run.push(format!("`{}` (receipt {})", c.command, r.attempt_id));
            }
        }
    }
    if rows.is_empty() {
        rows.push("| (receipt 없음) | not_run | 검사 receipt가 입력되지 않았다 |".into());
    }
    not_run.extend(input.not_run.iter().cloned());
    let not_run_lines = if not_run.is_empty() {
        "- not_run: 없음".to_string()
    } else {
        not_run
            .iter()
            .map(|s| format!("- not_run: {s}"))
            .collect::<Vec<_>>()
            .join("\n")
    };
    let review = input
        .review_link
        .unwrap_or("<대기: 독립 review 판정 뒤 채운다>");
    let op = operation_id(rec);
    let marker = format!("<!-- moai-op {op} -->");

    let body = format!(
        "Refs #{n}\n\n\
## 대상과 claim\n\n\
- 대상 Issue: #{n} ({lk}) · {url}\n\
- logical_key: `{lk}`\n\
- claim 토큰: {token}\n\
- epoch: {epoch}\n\
- base: `{base_branch}` @ `{base}`\n\
- candidate head: `{head}` (exact SHA, 독립 review가 본 SHA와 같아야 한다)\n\
- SPEC: `{sid}` · revision `{srev}` · digest `{spec_digest}`\n\n\
## 실행한 검사와 exit\n\n\
| 명령 | exit | 요약 |\n\
| --- | --- | --- |\n\
{rows}\n\n\
## 독립 review\n\n\
- 판정: <대기> · 본 SHA: `{head}`\n\
- 판정 댓글 링크: {review}\n\n\
## not_run / blocked\n\n\
{not_run_lines}\n\n\
## 남은 gap과 위험\n\n\
- PR 개설은 완료가 아니다. 통합과 Issue 종료는 리더가 판정한다.\n\
- reconcile: not_run (06-B #2006 미완, op marker만 기록)\n\n\
{marker}\n",
        lk = rec.logical_key,
        epoch = rec.epoch,
        base_branch = input.base_branch,
        base = input.base_sha,
        head = input.candidate_sha,
        sid = rec.spec_ref.id,
        srev = rec.spec_ref.revision,
        rows = rows.join("\n"),
    );
    let found = closing_keywords(&body);
    if !found.is_empty() {
        return Err(format!(
            "conflict in PR body: closing keyword {} would be rendered; nothing written",
            found.join(", ")
        ));
    }
    let mut warnings = Vec::new();
    if input.allow_close {
        warnings.push(if input.base_branch != input.default_branch {
            format!(
                "closing keywords are read only on PRs into the default branch ({}); none were rendered",
                input.default_branch
            )
        } else {
            "closing keywords are never rendered; the leader closes the Issue".to_string()
        });
    }
    Ok(json!({
        "body": body,
        "body_digest": digest(body.as_bytes()),
        "closing_keywords": 0,
        "operation_id": op,
        "marker": marker,
        "reconcile": "not_run: 06-B #2006 not landed; marker recorded only",
        "warnings": warnings,
        "record_update_request": {
            "task_key": rec.task_key,
            "epoch": rec.epoch,
            "state": "pr_ready",
        },
        "record_update_owner": RECORD_UPDATE_OWNER,
    }))
}
