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
//! reads (`<!-- moai-op pr_create:<repository_id>:<task_key>:<epoch> -->`, r33), the lookup
//! key for 06-B's reconcile (REQ-06C-09). [`render_and_record`] records the matching 06-B
//! `effect-intend{kind: pr_create}` (same operation, project, local id and attempt, so
//! `parse_op_markers` reads both markers as one operation) when the record came from the
//! journal. Rendering is not opening: the record moves to `pr_ready` only at the PR's ACK
//! (REQ-06C-11, "PR opening and ACK"), through [`confirm_and_record`] (`pr-confirm`), which confirms
//! the intent with 06-B's `effect_confirm` — or takes an effect a reconcile already
//! confirmed — and never goes further than `pr_ready`. An `uncertain` ACK leaves the state
//! alone; `effects-reconcile` resolves it first.
use crate::claim::{claim_token, ClaimRecord};
use crate::digest;
use crate::forge::policy::Clock;
use crate::run::Receipt;
use crate::verify::CheckResult;
use crate::worktree::{
    commit_update, is_sha, not_written, open_locked_journal, RECORD_UPDATE_OWNER,
};
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
    /// Extra bullets for the remaining-gap section.
    pub gaps: &'a [String],
    /// `(status, path)` from `git diff --name-status <base>...<candidate>`.
    pub changes: &'a [(String, String)],
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
    let change_rows = if input.changes.is_empty() {
        "| (변경 없음) | base와 candidate 사이에 바뀐 경로가 없다 |".to_string()
    } else {
        input
            .changes
            .iter()
            .map(|(status, path)| format!("| `{}` | {} |", cell(path), cell(status)))
            .collect::<Vec<_>>()
            .join("\n")
    };
    let gap_lines: String = input.gaps.iter().map(|g| format!("- {g}\n")).collect();
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
## 변경\n\n\
| 경로 | git 상태 |\n\
| --- | --- |\n\
{change_rows}\n\n\
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
- reconcile: not_run (렌더는 effect intent만 기록한다. 조회는 `effects-reconcile`이 uncertain effect에만 한다)\n\
{gap_lines}\n\
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
        "reconcile": "not_run: effects-reconcile probes only an uncertain effect",
        "warnings": warnings,
        "record_update_request": {
            "task_key": rec.task_key,
            "epoch": rec.epoch,
            "state": "pr_ready",
            "at": "the pr_create ACK (pr-confirm), not the render",
        },
        "record_update_owner": RECORD_UPDATE_OWNER,
    }))
}

/// The intent payload: what is being created, not the body text, so a re-render after a
/// new commit is the same intent (06-B rejects a different payload under one key).
fn intent_payload(rec: &ClaimRecord, base_branch: &str) -> Vec<u8> {
    json!({
        "operation": "pr_create",
        "task_key": rec.task_key,
        "epoch": rec.epoch,
        "issue_number": rec.task_ref.issue_number,
        "base_branch": base_branch,
    })
    .to_string()
    .into_bytes()
}

/// [`render`] plus the record write (REQ-06C-09): the 06-B `pr_create` intent only; the
/// state stays as it is until the ACK ([`confirm_and_record`]). A record from a file is
/// rendered only; nothing is written for it.
pub fn render_and_record<C: Clock>(
    project: &std::path::Path,
    rec: &ClaimRecord,
    from_journal: bool,
    receipts: &[Receipt],
    input: &RenderInput,
    clock: &C,
) -> Result<Value> {
    let journal = if from_journal {
        Some(open_locked_journal(project)?)
    } else {
        None
    };
    let mut out = render(rec, receipts, input)?;
    let Some(journal) = journal else {
        out["record_update"] = not_written();
        out["effect_intent"] = Value::Null;
        return Ok(out);
    };
    let epoch = u32::try_from(rec.epoch).map_err(|_| "epoch out of range for an operation id")?;
    let intent = crate::handoff::effect_intend(
        &journal,
        &rec.task_key,
        rec.epoch,
        "pr_create",
        &rec.task_ref.repository_id.to_string(),
        &rec.task_key,
        epoch,
        "pr_create",
        &intent_payload(rec, input.base_branch),
        clock,
    )?;
    out["effect_intent"] = json!({
        "operation_id": intent["operation_id"],
        "marker": intent["marker"],
        "state": intent["state"],
        "intent_digest": intent["intent_digest"],
    });
    out["record_update"] = json!({
        "applied": intent["notes"].is_null(),
        "event": "effect_intend",
        "state": "unchanged until the pr_create ACK (pr-confirm)",
    });
    Ok(out)
}

/// `pr-confirm` (REQ-06C-11): the PR was opened and its ACK observed. An `intended`
/// `pr_create` effect is confirmed through 06-B's `effect_confirm` (`uncertain` when the
/// answer was lost); a `confirmed` effect — the reconcile case — is taken as is. Only a
/// confirmed effect moves the record to `pr_ready`.
pub fn confirm_and_record<C: Clock>(
    project: &std::path::Path,
    key: &str,
    operation_id: &str,
    response_id: &str,
    observed_state: &str,
    clock: &C,
) -> Result<Value> {
    let journal = open_locked_journal(project)?;
    let current = journal
        .read_events()?
        .into_iter()
        .rev()
        .find(|e| e.record.task_key == key)
        .map(|e| e.record)
        .ok_or("claim record not found for task")?;
    let effect = current
        .effects
        .iter()
        .find(|e| e["operation_id"] == operation_id && e["kind"] == "pr_create")
        .cloned()
        .ok_or_else(|| {
            format!("unknown operation_id {operation_id}; no pr_create was intended under it")
        })?;
    let confirm = if effect["state"] == "intended" {
        crate::handoff::effect_confirm(
            &journal,
            key,
            operation_id,
            response_id,
            observed_state,
            clock,
        )?
    } else {
        Value::Null
    };
    let state = if confirm.is_null() {
        effect["state"].clone()
    } else {
        confirm["state"].clone()
    };
    let record_update = if state == "confirmed" {
        commit_update(&journal, key, current.epoch, "pr_ready", clock, |r| {
            r.state = "pr_ready".into()
        })?
    } else {
        json!({
            "applied": false,
            "reason": format!("the pr_create effect is {state}; pr_ready waits for a confirmed ACK"),
        })
    };
    Ok(json!({
        "operation_id": operation_id,
        "effect_confirm": confirm,
        "effect_state": state,
        "record_update": record_update,
    }))
}
