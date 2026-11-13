//! Concise Issue report: one owned comment per run, plus artifact references (07-C #2010).
//!
//! `forge-report` renders the design-22 `issue-report` template from a typed payload and
//! keeps exactly one comment per `stable_run_id` on the Issue, found by its first line
//! `<!-- moai:report run=<id> -->` (marker rules and the outbox belong to 07-B #2009; this
//! module only calls them). Everything is checked before anything is written:
//!
//! 1. permission reference, schema, run id, the `done` gate (every required check `passed`
//!    on the candidate revision), artifact references (sha256 required; `kind: local` files
//!    must match under `<project>/.moai-cli/evidence/<run>/`), the redaction gate (control
//!    characters removed, credential shapes and personal paths refused as
//!    `blocked_sensitive`), and the rendered length (`too_large`).
//! 2. local evidence copies (temporary file + rename).
//! 3. the network: the Issue and repository are re-read against the `ExternalRef`, every
//!    comment page is read without `since`, then one PATCH (marker comment found, body still
//!    what this machine last published) or one POST (no marker comment), then a GET confirms
//!    the body before the outbox records `confirmed`.
//!
//! A comment edited by someone else is never overwritten: the result is `conflict` and the
//! proposed body is saved beside the evidence. A write whose outcome is unknown is left
//! `publish_unknown`; the next run re-reads the comments first and only then PATCHes or POSTs
//! once, so no duplicate comment appears.
//!
//! Numbers that are not decided yet (P11): `max_body_chars` and `tail_lines` are settings
//! with conservative, unmeasured defaults. Uploading artifacts is not implemented (P09):
//! local references are rendered with an export/pin-required marker.
use super::{
    outbox::{self, OutboxEntry, OutboxState, Phase},
    policy::{format_utc, next_link, parse_utc, Clock},
    transport::{HttpRequest, HttpResponse, Transport},
    Client, Config, ForgeError, API,
};
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use std::{
    fs::{self, OpenOptions},
    io::Write,
    path::{Path, PathBuf},
};

const OPERATION: &str = "report";
/// Unmeasured (P11): below the commonly cited, officially undocumented 65,536-character
/// comment limit.
pub const DEFAULT_MAX_BODY_CHARS: usize = 65_535;
/// Unmeasured (P11) default for the per-check output tail.
pub const DEFAULT_TAIL_LINES: usize = 20;
/// Local evidence files are copied whole; larger output belongs to an external store (P09).
const MAX_LOCAL_ARTIFACT: u64 = 64 * 1024 * 1024;

#[derive(Clone, Copy, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum Status {
    Running,
    Blocked,
    Review,
    Done,
}

#[derive(Clone, Copy, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum CheckResult {
    Passed,
    Failed,
    NotRun,
    Blocked,
    Stale,
    Unknown,
}

#[derive(Clone, Debug, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct Check {
    pub name: String,
    pub command: String,
    pub revision: String,
    pub exit_code: Option<i32>,
    pub result: CheckResult,
    /// Counts for `done`. Defaults to true, so a forgotten flag still blocks `done`.
    #[serde(default = "yes")]
    pub required: bool,
    /// Last lines of the output; only `config.tail_lines` of them are rendered.
    #[serde(default)]
    pub tail: Vec<String>,
}
fn yes() -> bool {
    true
}

#[derive(Clone, Copy, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum ArtifactKind {
    Local,
    ActionsArtifact,
    ReleaseAsset,
    IssueArtifact,
    External,
}

#[derive(Clone, Copy, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum Access {
    Public,
    Repo,
    Private,
}

#[derive(Clone, Debug, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct ArtifactRef {
    pub kind: ArtifactKind,
    /// `sha256:<64 hex>`; a reference without it is refused (REQ-REPORT-009).
    #[serde(default)]
    pub digest: Option<String>,
    pub size_bytes: u64,
    /// `kind: local`: a file name inside `<project>/.moai-cli/evidence/<run>/`.
    /// Other kinds: an `https://` URL.
    pub location: String,
    /// `until:<YYYY-MM-DDTHH:MM:SSZ>` or `unknown`.
    pub retention: String,
    pub access: Access,
    pub redacted: bool,
    /// `kind: local` only: a file to copy into the evidence directory (absolute or relative
    /// to the project). Without it the file must already be there.
    #[serde(default)]
    pub source: Option<String>,
}

#[derive(Clone, Debug, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct ReportConfig {
    #[serde(default = "default_max_body")]
    pub max_body_chars: usize,
    #[serde(default = "default_tail")]
    pub tail_lines: usize,
    #[serde(default)]
    pub forge: Config,
}
fn default_max_body() -> usize {
    DEFAULT_MAX_BODY_CHARS
}
fn default_tail() -> usize {
    DEFAULT_TAIL_LINES
}
impl Default for ReportConfig {
    fn default() -> Self {
        ReportConfig {
            max_body_chars: DEFAULT_MAX_BODY_CHARS,
            tail_lines: DEFAULT_TAIL_LINES,
            forge: Config::default(),
        }
    }
}

#[derive(Clone, Debug, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct ReportPayload {
    pub schema_version: u32,
    pub stable_run_id: String,
    /// The 07-A external reference of the Issue.
    pub issue: Value,
    pub status: Status,
    /// The exact revision under review; `done` requires every required check on it.
    pub candidate_revision: String,
    pub work: String,
    pub checks: Vec<Check>,
    #[serde(default)]
    pub gaps: Vec<String>,
    pub next: String,
    #[serde(default)]
    pub evidence_links: Vec<String>,
    #[serde(default)]
    pub artifacts: Vec<ArtifactRef>,
    #[serde(default)]
    pub config: ReportConfig,
}

/// The parts of the `ExternalRef` a report needs.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct IssueRef {
    pub owner: String,
    pub repo: String,
    pub repository_id: i64,
    pub id: i64,
    pub node_id: String,
    pub number: i64,
}

fn invalid(why: &str) -> ForgeError {
    ForgeError::invalid(format!("invalid report: {why}"))
}

fn issue_ref(v: &Value) -> Result<IssueRef, ForgeError> {
    let s = |p: &str| v.pointer(p).and_then(Value::as_str).map(str::to_string);
    let n = |p: &str| v.pointer(p).and_then(Value::as_i64);
    if s("/provider").as_deref() != Some("github") || s("/host").as_deref() != Some("github.com") {
        return Err(invalid("issue must be a github.com ExternalRef"));
    }
    let r = IssueRef {
        owner: s("/project/owner").unwrap_or_default(),
        repo: s("/project/repo").unwrap_or_default(),
        repository_id: n("/project/repository_id").unwrap_or(0),
        id: n("/issue/id").unwrap_or(0),
        node_id: s("/issue/node_id").unwrap_or_default(),
        number: n("/issue/number").unwrap_or(0),
    };
    if !super::valid_name(&r.owner)
        || !super::valid_name(&r.repo)
        || r.repository_id <= 0
        || r.id <= 0
        || r.node_id.is_empty()
        || r.number <= 0
    {
        return Err(invalid(
            "issue needs project{owner, repo, repository_id} and issue{id, node_id, number}",
        ));
    }
    Ok(r)
}

fn valid_run(s: &str) -> bool {
    !s.is_empty()
        && s.len() <= 64
        && s.bytes()
            .all(|b| b.is_ascii_alphanumeric() || b"-_.".contains(&b))
        && s != "."
        && s != ".."
}

fn valid_digest(s: &str) -> bool {
    s.len() == 71
        && s.starts_with("sha256:")
        && s[7..]
            .bytes()
            .all(|b| b.is_ascii_digit() || (b'a'..=b'f').contains(&b))
}

/// Conservative shape filter: the shared redact module's shapes (profile shapes
/// plus the merged gh token prefixes — #2351, single list) plus this report's own
/// extra cloud/chat key shapes and personal absolute paths. The extra shapes below
/// stay here: adding new prefixes elsewhere is out of this card's scope.
fn sensitive(s: &str) -> bool {
    let lower = s.to_ascii_lowercase();
    crate::redact::contains_secret_shape(s)
        || [
            "akia",
            "bearer ",
            "xoxb-",
            "xoxp-",
            "aws_secret",
            "/users/",
            "/home/",
            "~/",
            ":\\",
        ]
        .iter()
        .any(|x| lower.contains(x))
}

/// Removes control characters (newlines included) so every field renders on its line.
fn clean(s: &str) -> String {
    s.chars()
        .map(|c| if c.is_control() { ' ' } else { c })
        .collect::<String>()
        .trim()
        .to_string()
}

/// Every string that can reach the body, already cleaned; any shape hit blocks publication.
fn gate(p: &ReportPayload) -> Result<(), ForgeError> {
    let mut all: Vec<&str> = vec![&p.stable_run_id, &p.candidate_revision, &p.work, &p.next];
    for c in &p.checks {
        all.extend([c.name.as_str(), &c.command, &c.revision]);
        all.extend(c.tail.iter().map(String::as_str));
    }
    all.extend(p.gaps.iter().map(String::as_str));
    all.extend(p.evidence_links.iter().map(String::as_str));
    for a in &p.artifacts {
        all.extend([a.location.as_str(), &a.retention]);
        all.extend(a.digest.as_deref());
    }
    if all.iter().any(|s| sensitive(s)) {
        return Err(ForgeError::blocked(
            "blocked_sensitive: a report field matched a credential shape or a personal path; nothing was published",
            false,
        ));
    }
    Ok(())
}

/// REQ-REPORT-001/002/009 shape checks; nothing is written or sent before they pass.
pub fn validate(p: &ReportPayload) -> Result<IssueRef, ForgeError> {
    if p.schema_version != 1 {
        return Err(ForgeError::of(
            "unsupported_protocol_version",
            "unsupported request schema_version",
        ));
    }
    if !valid_run(&p.stable_run_id) {
        return Err(invalid("stable_run_id must be a portable id"));
    }
    let issue = issue_ref(&p.issue)?;
    let one_line = |s: &str| !clean(s).is_empty() && s.chars().count() <= 500;
    if !one_line(&p.work) || !one_line(&p.next) || clean(&p.candidate_revision).is_empty() {
        return Err(invalid(
            "work, next and candidate_revision must be nonempty and short",
        ));
    }
    if p.checks.iter().any(|c| {
        clean(&c.name).is_empty() || clean(&c.command).is_empty() || clean(&c.revision).is_empty()
    }) {
        return Err(invalid("each check needs a name, command and revision"));
    }
    if p.status == Status::Done {
        let required: Vec<&Check> = p.checks.iter().filter(|c| c.required).collect();
        let ok = !required.is_empty()
            && required
                .iter()
                .all(|c| c.result == CheckResult::Passed && c.revision == p.candidate_revision);
        if !ok {
            return Err(invalid(
                "status done requires every required check passed on candidate_revision",
            ));
        }
    }
    if p.evidence_links.iter().any(|l| !l.starts_with("https://")) {
        return Err(invalid("evidence links must be https URLs"));
    }
    for a in &p.artifacts {
        let bad = |why: &str| ForgeError::invalid(format!("invalid artifact reference: {why}"));
        match a.digest.as_deref() {
            Some(d) if valid_digest(d) => {}
            _ => return Err(bad("a sha256 digest is required")),
        }
        let retention_ok = a.retention == "unknown"
            || a.retention
                .strip_prefix("until:")
                .is_some_and(|t| parse_utc(t).is_some());
        if !retention_ok {
            return Err(bad("retention must be until:<UTC time> or unknown"));
        }
        match a.kind {
            ArtifactKind::Local => {
                if !valid_run(&a.location) {
                    return Err(bad(
                        "a local location is a file name in the evidence directory",
                    ));
                }
            }
            _ => {
                if !a.location.starts_with("https://") || a.source.is_some() {
                    return Err(bad("a remote location is an https URL without source"));
                }
            }
        }
    }
    Ok(issue)
}

fn evidence_rel(run: &str, name: &str) -> String {
    format!("{}/evidence/{run}/{name}", crate::PROJECT_STATE_DIR)
}

/// The design-22 `issue-report` template; LF only, no trailing newline, so the digest of
/// what GitHub returns can be compared with what was sent.
pub fn render(p: &ReportPayload) -> Result<String, ForgeError> {
    validate(p)?;
    gate(p)?;
    let status = match p.status {
        Status::Running => "running",
        Status::Blocked => "blocked",
        Status::Review => "review",
        Status::Done => "done",
    };
    let label = |r: CheckResult| match r {
        CheckResult::Passed => "passed",
        CheckResult::Failed => "failed",
        CheckResult::NotRun => "not_run",
        CheckResult::Blocked => "blocked",
        CheckResult::Stale => "stale",
        CheckResult::Unknown => "unknown",
    };
    let mut lines = vec![
        outbox::report_marker(&p.stable_run_id),
        format!("MoAI run: {}", p.stable_run_id),
        format!("상태: {status}"),
        format!("수행 내용: {}", clean(&p.work)),
        format!("시험 결과: candidate {}", clean(&p.candidate_revision)),
    ];
    if p.checks.is_empty() {
        lines.push("- 실행한 검사 없음".into());
    }
    for c in &p.checks {
        let exit = c.exit_code.map_or("none".to_string(), |e| e.to_string());
        let required = if c.required { "required" } else { "optional" };
        lines.push(format!(
            "- {}: {} · exit {exit} · revision {} · {required} · {}",
            clean(&c.name),
            label(c.result),
            clean(&c.revision),
            clean(&c.command).replace('`', "'"),
        ));
        let keep = c.tail.len().saturating_sub(p.config.tail_lines);
        let tail: Vec<String> = c.tail[keep..]
            .iter()
            .map(|l| clean(l).replace("```", "'''"))
            .collect();
        if !tail.is_empty() {
            lines.push("  ```text".into());
            lines.extend(tail.into_iter().map(|l| format!("  {l}")));
            lines.push("  ```".into());
        }
    }
    for g in &p.gaps {
        lines.push(format!("- gap: {}", clean(g)));
    }
    lines.push(format!("다음 작업: {}", clean(&p.next)));
    lines.push(if p.evidence_links.is_empty() {
        "근거: 없음".into()
    } else {
        format!(
            "근거: {}",
            p.evidence_links
                .iter()
                .map(|l| clean(l))
                .collect::<Vec<_>>()
                .join(" · ")
        )
    });
    if p.artifacts.is_empty() {
        lines.push("상세 로그: 없음".into());
    } else {
        lines.push("상세 로그:".into());
        for a in &p.artifacts {
            let kind = match a.kind {
                ArtifactKind::Local => "local",
                ArtifactKind::ActionsArtifact => "actions_artifact",
                ArtifactKind::ReleaseAsset => "release_asset",
                ArtifactKind::IssueArtifact => "issue_artifact",
                ArtifactKind::External => "external",
            };
            let access = match a.access {
                Access::Public => "public",
                Access::Repo => "repo",
                Access::Private => "private",
            };
            let location = match a.kind {
                ArtifactKind::Local => evidence_rel(&p.stable_run_id, &a.location),
                _ => clean(&a.location),
            };
            let pin = if a.kind == ArtifactKind::Local {
                " · export/pin 필요"
            } else {
                ""
            };
            lines.push(format!(
                "- {kind} · {location} · {} · {} bytes · retention {} · access {access} · redacted {}{pin}",
                a.digest.as_deref().unwrap_or_default(),
                a.size_bytes,
                clean(&a.retention),
                a.redacted,
            ));
        }
    }
    let body = lines.join("\n");
    let chars = body.chars().count();
    if chars > p.config.max_body_chars {
        return Err(invalid(&format!(
            "too_large: rendered body has {chars} characters, over config.max_body_chars {}",
            p.config.max_body_chars
        )));
    }
    Ok(body)
}

/// `<project>/.moai-cli/evidence/<run>/`, created on demand; symlinks are refused.
fn evidence_dir(project: &Path, run: &str) -> Result<PathBuf, ForgeError> {
    let mut dir = project.to_path_buf();
    for part in [crate::PROJECT_STATE_DIR, "evidence", run] {
        dir.push(part);
        match fs::symlink_metadata(&dir) {
            Ok(m) if m.file_type().is_symlink() || !m.is_dir() => {
                return Err(ForgeError::invalid(
                    "invalid evidence directory: a symlink or not a directory; refused",
                ))
            }
            Ok(_) => {}
            Err(_) => fs::create_dir(&dir)
                .map_err(|_| ForgeError::io("evidence directory unavailable"))?,
        }
    }
    Ok(dir)
}

fn regular_file(path: &Path) -> Option<u64> {
    match fs::symlink_metadata(path) {
        Ok(m) if m.is_file() && !m.file_type().is_symlink() => Some(m.len()),
        _ => None,
    }
}

fn write_atomic(dir: &Path, name: &str, bytes: &[u8]) -> Result<(), ForgeError> {
    let tmp = dir.join(format!(".{name}.tmp-{}", std::process::id()));
    let mut f = OpenOptions::new()
        .write(true)
        .create_new(true)
        .open(&tmp)
        .map_err(|_| ForgeError::io("evidence write failed"))?;
    f.write_all(bytes)
        .and_then(|_| f.sync_all())
        .map_err(|_| ForgeError::io("evidence write failed"))?;
    fs::rename(&tmp, dir.join(name)).map_err(|_| {
        let _ = fs::remove_file(&tmp);
        ForgeError::io("evidence replace failed")
    })
}

/// Checks every `kind: local` reference, then copies the sources (REQ-REPORT-010). A file
/// that does not hash to its declared digest or size is refused before anything is copied.
pub fn preserve_local(project: &Path, p: &ReportPayload) -> Result<Vec<Value>, ForgeError> {
    let bad = |why: &str| ForgeError::invalid(format!("invalid artifact reference: {why}"));
    let local: Vec<&ArtifactRef> = p
        .artifacts
        .iter()
        .filter(|a| a.kind == ArtifactKind::Local)
        .collect();
    if local.is_empty() {
        return Ok(Vec::new());
    }
    let target_dir = project
        .join(crate::PROJECT_STATE_DIR)
        .join("evidence")
        .join(&p.stable_run_id);
    let mut copies = Vec::new();
    for a in &local {
        let from = match &a.source {
            Some(s) => {
                let sp = Path::new(s);
                if sp.is_absolute() {
                    sp.to_path_buf()
                } else {
                    project.join(sp)
                }
            }
            None => target_dir.join(&a.location),
        };
        match regular_file(&from) {
            Some(len) if len <= MAX_LOCAL_ARTIFACT => {}
            Some(_) => return Err(bad("a local file is over the 64 MiB local limit")),
            None => return Err(bad("the local file is not a regular file in place")),
        }
        let bytes = fs::read(&from).map_err(|_| ForgeError::io("evidence file unreadable"))?;
        if Some(crate::digest(&bytes).as_str()) != a.digest.as_deref()
            || bytes.len() as u64 != a.size_bytes
        {
            return Err(bad(
                "the local file does not hash to its declared sha256 and size",
            ));
        }
        copies.push((a, bytes, a.source.is_some()));
    }
    let dir = evidence_dir(project, &p.stable_run_id)?;
    let mut out = Vec::new();
    for (a, bytes, copy) in copies {
        if copy {
            write_atomic(&dir, &a.location, &bytes)?;
        }
        out.push(json!({"location": evidence_rel(&p.stable_run_id, &a.location), "digest": a.digest, "size_bytes": a.size_bytes}));
    }
    Ok(out)
}

/// The publication history of one run on one Issue, oldest first.
fn history(project: &Path, scope: &str, number: &str) -> Result<Vec<OutboxEntry>, ForgeError> {
    let mut out = Vec::new();
    for attempt in 1.. {
        match outbox::load(project, &outbox::op_id(OPERATION, scope, number, attempt)?)? {
            Some(e) => out.push(e),
            None => break,
        }
    }
    Ok(out)
}

/// What this machine last put on the remote, or tried to: (state, comment id, body digest).
fn digest_of(detail: &Value) -> String {
    detail["body_digest"].as_str().unwrap_or("").to_string()
}

/// The body this machine last saw confirmed on the remote: (comment id, body digest).
fn last_confirmed(h: &[OutboxEntry]) -> Option<(Option<i64>, String)> {
    h.iter().rev().find_map(|e| match (e.state, &e.observed) {
        (OutboxState::Confirmed, Some(p)) => {
            Some((p.detail["comment_id"].as_i64(), digest_of(&p.detail)))
        }
        _ => None,
    })
}

/// The newest write whose outcome is unknown, if it is the newest entry: its intent digest.
fn pending(h: &[OutboxEntry]) -> Option<String> {
    match h.last() {
        Some(e)
            if matches!(
                e.state,
                OutboxState::PublishUnknown | OutboxState::Publishing
            ) =>
        {
            e.intent.as_ref().map(|p| digest_of(&p.detail))
        }
        _ => None,
    }
}

/// The P11 limits in force; their defaults are not measured (leader interim decision).
fn limits(p: &ReportPayload) -> Value {
    json!({"max_body_chars": p.config.max_body_chars, "tail_lines": p.config.tail_lines,
           "defaults": "unmeasured"})
}

fn save_proposal(project: &Path, run: &str, body: &str) -> Result<String, ForgeError> {
    let dir = evidence_dir(project, run)?;
    write_atomic(&dir, "proposed-comment.md", body.as_bytes())?;
    Ok(evidence_rel(run, "proposed-comment.md"))
}

/// REQ-REPORT-003..008: render, preserve, then create or update the single marker comment.
pub fn publish<T: Transport, C: Clock>(
    project: &Path,
    payload: &ReportPayload,
    permission: &str,
    client: &Client<'_, T, C>,
) -> Result<Value, ForgeError> {
    // Caller authorization remains host-owned; the reference is not an authorization verifier.
    if !crate::contract::permission_reference_ok(permission) {
        return Err(ForgeError::of(
            "permission_reference_missing",
            "permission reference required",
        ));
    }
    let issue = validate(payload)?;
    let body = render(payload)?;
    let body_digest = crate::digest(body.as_bytes());
    let run = payload.stable_run_id.as_str();
    let marker = outbox::report_marker(run);
    let preserved = preserve_local(project, payload)?;

    // Re-read the Issue and repository: the number alone is never a key (07-A).
    let repo_url = format!("{API}/repos/{}/{}", issue.owner, issue.repo);
    let repo = client.get(&repo_url, None)?.body.unwrap_or(Value::Null);
    let got = client
        .get(&format!("{repo_url}/issues/{}", issue.number), None)?
        .body
        .unwrap_or(Value::Null);
    if repo["id"].as_i64() != Some(issue.repository_id)
        || got["id"].as_i64() != Some(issue.id)
        || got["node_id"].as_str() != Some(issue.node_id.as_str())
    {
        return Err(ForgeError::of(
            "conflict",
            "conflict in issue reference: the repository or Issue no longer matches the ExternalRef; nothing published",
        ));
    }

    // Every comment page, without `since` (REQ-REPORT-004); bodies reduced to digests at once.
    let project_page = |r: &HttpResponse| {
        let items: Vec<Value> = r
            .body
            .as_ref()
            .and_then(Value::as_array)
            .map(|a| {
                a.iter()
                    .map(|c| {
                        let b = c["body"].as_str().unwrap_or("");
                        json!({"id": c["id"], "node_id": c["node_id"], "html_url": c["html_url"],
                               "updated_at": c["updated_at"],
                               "marker": b.lines().next() == Some(marker.as_str()),
                               "body_digest": crate::digest(b.as_bytes())})
                    })
                    .collect()
            })
            .unwrap_or_default();
        json!({"items": items, "next": next_link(r.header("link"))})
    };
    let (comments, _) = client.pages(
        format!("{repo_url}/issues/{}/comments?per_page=100", issue.number),
        &project_page,
    )?;
    let marked: Vec<&Value> = comments
        .iter()
        .filter(|c| c["marker"] == json!(true))
        .collect();
    if marked.len() > 1 {
        return Err(ForgeError::of(
            "conflict",
            "conflict in report comment: more than one comment carries this run marker; nothing changed",
        ));
    }

    let scope = format!("{}.{}", issue.owner, issue.repo);
    let number = issue.number.to_string();
    let hist = history(project, &scope, &number)?;
    let confirmed = last_confirmed(&hist);
    let unknown_digest = pending(&hist);
    let now = || format_utc(client.clock.now());

    // Decide PATCH, POST or nothing. With a write of unknown outcome pending, the remote body
    // tells which happened: the intended body (it landed: confirm it), the last confirmed body
    // (it did not land: write again), or anything else (someone edited it: conflict).
    let mut deleted: Option<i64> = None;
    let mut superseded: Option<String> = None;
    let remote_digest = marked
        .first()
        .and_then(|c| c["body_digest"].as_str())
        .map(str::to_string);
    let current: Option<String> = match (&remote_digest, &unknown_digest, &confirmed) {
        (Some(r), Some(p), _) if r == p => {
            if let Some(mut e) = hist.last().cloned() {
                e.advance(OutboxState::Confirmed, Phase {
                    at: now(),
                    detail: json!({"comment_id": marked[0]["id"], "body_digest": p, "via": "marker"}),
                })?;
                outbox::save(project, &e)?;
            }
            Some(p.clone())
        }
        (Some(r), Some(_), Some((_, d))) if r == d => {
            superseded = hist.last().map(|e| e.op_id.clone());
            Some(d.clone())
        }
        (Some(r), None, Some((_, d))) if r == d => Some(d.clone()),
        _ => None,
    };
    let target: Option<&Value> = match (marked.first(), &current, &confirmed) {
        (Some(c), Some(d), _) => {
            if d == &body_digest {
                return Ok(
                    json!({"schema_version": 1, "stable_run_id": run, "action": "unchanged",
                    "comment": {"id": c["id"], "node_id": c["node_id"], "html_url": c["html_url"]},
                    "body_digest": body_digest, "body_chars": body.chars().count(), "limits": limits(payload), "artifacts": preserved}),
                );
            }
            Some(c)
        }
        (Some(_), None, _) => {
            let proposal = save_proposal(project, run, &body)?;
            return Err(ForgeError::of(
                "conflict",
                format!("conflict in report comment: it was edited since this machine last published it (or was never published from here); nothing overwritten, proposal saved at {proposal}"),
            ));
        }
        (None, _, Some((id, _))) if unknown_digest.is_none() => {
            // Someone deleted the comment: record that, then create it again.
            deleted = *id;
            None
        }
        (None, _, _) => None,
    };

    let attempt = hist.len() as u32 + 1;
    let op_id = outbox::op_id(OPERATION, &scope, &number, attempt)?;
    let mut entry = OutboxEntry::new(&op_id, &marker);
    entry.advance(OutboxState::Publishing, Phase {
        at: now(),
        detail: json!({"body_digest": body_digest, "comment_id": target.map(|c| c["id"].clone()),
                       "observed_deleted": deleted, "supersedes": superseded,
                       "method": if target.is_some() { "PATCH" } else { "POST" }}),
    })?;
    outbox::save(project, &entry)?;
    let request = match target {
        Some(c) => HttpRequest::new("PATCH", format!("{repo_url}/issues/comments/{}", c["id"]))
            .with_body(json!({"body": body})),
        None => HttpRequest::new(
            "POST",
            format!("{repo_url}/issues/{}/comments", issue.number),
        )
        .with_body(json!({"body": body})),
    };
    let unknown = |entry: &mut OutboxEntry, why: String| -> Result<ForgeError, ForgeError> {
        entry.advance(
            OutboxState::PublishUnknown,
            Phase {
                at: now(),
                detail: json!({"reason": why}),
            },
        )?;
        outbox::save(project, entry)?;
        Ok(ForgeError::transport(format!(
            "report publish outcome unknown ({why}); the next run re-reads the comments before writing"
        )))
    };
    let resp = match client.send_with_backoff(&request) {
        Ok(r) => r,
        Err(e) if e.kind == "blocked" => {
            entry.advance(
                OutboxState::Blocked,
                Phase {
                    at: now(),
                    detail: json!({"reason": "blocked"}),
                },
            )?;
            outbox::save(project, &entry)?;
            return Err(e);
        }
        Err(e) => return Err(unknown(&mut entry, e.kind.to_string())?),
    };
    if resp.status >= 500 || resp.status == 408 {
        return Err(unknown(&mut entry, format!("HTTP {}", resp.status))?);
    }
    if !(200..300).contains(&resp.status) {
        entry.advance(
            OutboxState::Blocked,
            Phase {
                at: now(),
                detail: json!({"status": resp.status}),
            },
        )?;
        outbox::save(project, &entry)?;
        return Err(ForgeError::blocked(
            format!("report comment refused (HTTP {})", resp.status),
            false,
        ));
    }
    let sent = resp.body.clone().unwrap_or(Value::Null);
    entry.record_response(Phase {
        at: now(),
        detail: json!({"status": resp.status, "comment_id": sent["id"]}),
    })?;
    outbox::save(project, &entry)?;

    // Confirm by a fresh GET before recording `confirmed`.
    let confirm = match client.get(&format!("{repo_url}/issues/comments/{}", sent["id"]), None) {
        Ok(r) => r.body.unwrap_or(Value::Null),
        Err(e) => {
            return Err(unknown(
                &mut entry,
                format!("confirmation read failed: {}", e.kind),
            )?)
        }
    };
    let text = confirm["body"].as_str().unwrap_or("");
    if confirm["id"] != sent["id"]
        || sent["id"].is_null()
        || crate::digest(text.as_bytes()) != body_digest
        || text.lines().next() != Some(marker.as_str())
    {
        entry.advance(
            OutboxState::Conflict,
            Phase {
                at: now(),
                detail: json!({"reason": "re-read did not match what was sent"}),
            },
        )?;
        outbox::save(project, &entry)?;
        return Err(ForgeError::of(
            "conflict",
            "conflict in report comment: the re-read comment does not match what was sent",
        ));
    }
    entry.advance(OutboxState::Confirmed, Phase {
        at: now(),
        detail: json!({"comment_id": confirm["id"], "node_id": confirm["node_id"], "body_digest": body_digest,
                       "updated_at": confirm["updated_at"]}),
    })?;
    outbox::save(project, &entry)?;
    Ok(json!({"schema_version": 1, "stable_run_id": run,
        "action": if target.is_some() { "updated" } else { "created" },
        "comment": {"id": confirm["id"], "node_id": confirm["node_id"], "html_url": confirm["html_url"]},
        "body_digest": body_digest, "body_chars": body.chars().count(), "limits": limits(payload), "op_id": op_id,
        "observed_deleted": deleted, "artifacts": preserved}))
}

/// CLI entry (`forge.report`): errors are their messages.
pub fn report_cli<T: Transport, C: Clock>(
    project: &Path,
    payload: ReportPayload,
    permission: &str,
    transport: &T,
    clock: &C,
) -> Result<Value, String> {
    let client = Client::new(
        transport,
        clock,
        payload.config.forge.clone(),
        Default::default(),
    );
    Ok(publish(project, &payload, permission, &client)?)
}
