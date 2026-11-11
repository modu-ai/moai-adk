//! Claim journal (① single-host) and Issue comment claim (⑤ cross-host, eventual)
//! (Issue #2005, 06-A; P08 decision 2026-11-08: ⑤ comment claim is primary, ① the
//! same-host serializer).
//!
//! ① One clone keeps one journal. All worktrees of a clone resolve the same common dir
//! through [`crate::identity`], so the journal lives at `<common dir>/.moai-cli/claims/`:
//! `journal.log` (append-only, one JSON line per event, fsync — the authority),
//! `<task_key>.json` (the current record, temp + fsync + rename — a human-readable cache),
//! and `.lock` (an advisory `File::try_lock` target). A second process that finds the lock
//! held gets `busy` at once and waits never (REQ-06A-05); the waiting policy is the caller's.
//! The lock serialises same-host claimants only: anything that bypasses `moai-cli` is not
//! serialised, which is why a claim that never went through the journal is invalid by
//! definition (RC1 reading of the advisory-lock boundary).
//!
//! ⑤ A claiming lane posts exactly one Issue comment
//! `moai-claim: session=<id> epoch=<n> lease_until=<utc>` (after the journal check) and
//! records the returned comment id in the record. When live claim comments collide, the
//! smallest comment id wins and the loser stops writing and gets `already_owned`
//! (REQ-06A-13). A winning comment with the caller's own session value is adopted only when
//! this journal recorded posting that comment id; otherwise the outcome is `uncertain` and
//! nothing is written, because a same-named live session may own it (#2165, REQ-AM-04).
//! The token `Moai-Claim: <issue>#<comment-id>` is carried in the branch
//! name (suffix `c<issue>-<id>`), in every commit trailer and in the PR body
//! (REQ-06A-14); [`verify_claim_token`] is the leader's read-side check (REQ-06A-15).
//! An expired lease is never taken over automatically: the outcome names the operator
//! (06-B) instead (REQ-06A-16).
//!
//! Lease (rc.2 #2166, decisions 00:5x 2026-11-09): an absent `lease_until` defaults to
//! now + [`DEFAULT_LEASE`] (4h); an explicit value wins. Expiry is only shown
//! (`claim-show` `expired: true`) and never moves ownership.
//!
//! HOLD (reported, never decided here): the heartbeat numbers (P08/P11); the journal location
//! is an interim derivation under the git common dir until P09/01-B fixes the state-root
//! map; takeover after expiry is 06-B's operator procedure.
use crate::contract::{Kind, OpError};
use crate::digest;
use crate::forge::policy::{format_utc, next_link, parse_utc, Clock};
use crate::forge::transport::{HttpRequest, Transport};
use crate::forge::{Client, ForgeError};
use crate::identity;
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use std::{
    collections::BTreeSet,
    fs::{self, OpenOptions},
    io::Write,
    path::{Path, PathBuf},
    process::{Command, Stdio},
};

type Result<T> = std::result::Result<T, String>;

pub const SCHEMA_VERSION: u32 = 1;
/// Default lease in seconds when a request carries no `lease_until` (#2166, 4h).
pub const DEFAULT_LEASE: i64 = 4 * 60 * 60;
/// Safety bound for the comment-list pagination, aligned with forge's `max_pages`.
const CLAIM_MAX_PAGES: u32 = 50;
/// The states in which a record holds the claim (plan §8). A record in any other state
/// (`unclaimed`, `integrated`, `cancelled`) does not block a new acquire.
pub const ACTIVE_STATES: &[&str] = &[
    "assigned",
    "running",
    "verifying",
    "pr_ready",
    "in_review",
    "blocked",
    "handoff_pending",
    "uncertain",
];

// @MX:NOTE: The journal root is derived from git's common dir (`<clone>/.git/.moai-cli/claims`)
// so every worktree of one clone shares one journal and git status stays clean. HOLD: the final
// state-root map belongs to P09 / 01-B (#1989); 06-B must migrate records if the location moves.
/// One clone keeps one journal; the common dir is the clone boundary (01-A), so all worktrees
/// resolve the same root and git never lists the files.
pub fn journal_root(project: &Path) -> Result<PathBuf> {
    let id = identity::identity(project)?;
    if id["identity_status"].as_str() != Some("ok") {
        return Err(format!(
            "claim journal requires a git repository; identity_status was {}",
            id["identity_status"].as_str().unwrap_or("unknown")
        ));
    }
    // #2175: the P09 opt-in home layout switches the journal store to
    // ~/.moai-cli/projects/<id>/claims once layout.json is in place.
    if let Some(home) = crate::layout::switched_dir(project, "claims") {
        return Ok(home);
    }
    let common = id["clone_anchor"]
        .as_str()
        .ok_or("claim journal requires a git repository")?;
    Ok(Path::new(common).join(".moai-cli").join("claims"))
}

// ---- Request and record ----

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Eq)]
#[serde(deny_unknown_fields)]
pub struct TaskRef {
    pub provider: String,
    pub host: String,
    /// Display data for the API URL; the key is `(provider, host, repository_id, issue_id)`.
    pub owner: String,
    pub repo: String,
    pub repository_id: i64,
    pub issue_id: i64,
    pub issue_number: i64,
}

impl TaskRef {
    /// Every field is required today; an optional field added later gets its default here,
    /// so call sites stay unchanged (#2256).
    pub fn new(
        provider: impl Into<String>,
        host: impl Into<String>,
        owner: impl Into<String>,
        repo: impl Into<String>,
        repository_id: i64,
        issue_id: i64,
        issue_number: i64,
    ) -> Self {
        TaskRef {
            provider: provider.into(),
            host: host.into(),
            owner: owner.into(),
            repo: repo.into(),
            repository_id,
            issue_id,
            issue_number,
        }
    }
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Eq)]
#[serde(deny_unknown_fields)]
pub struct Owner {
    /// `session`, `leader`, `operator`, `local-process`, `remote-task`, `ci` (bounded free
    /// form). `local-process` is the only kind 06-B probes for liveness.
    pub kind: String,
    pub id: String,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub host_id: Option<String>,
    /// 06-B additive: the OS pid of a `local-process` owner, recorded at claim time so a
    /// later handoff can probe liveness (REQ-06B-02). Old records deserialize unchanged.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub pid: Option<u32>,
    /// 06-B additive: the process start time (`ps -o lstart=` under `LC_ALL=C`); comparing
    /// it guards against pid reuse. Recorded with `pid`.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub process_start: Option<String>,
}

impl Owner {
    /// An owner with no host id, pid or process start (#2256).
    pub fn new(kind: impl Into<String>, id: impl Into<String>) -> Self {
        Owner {
            kind: kind.into(),
            id: id.into(),
            host_id: None,
            pid: None,
            process_start: None,
        }
    }

    pub fn host_id(mut self, host_id: impl Into<String>) -> Self {
        self.host_id = Some(host_id.into());
        self
    }

    pub fn pid(mut self, pid: u32) -> Self {
        self.pid = Some(pid);
        self
    }

    pub fn process_start(mut self, process_start: impl Into<String>) -> Self {
        self.process_start = Some(process_start.into());
        self
    }
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Eq)]
#[serde(deny_unknown_fields)]
pub struct SpecRef {
    pub id: String,
    pub revision: String,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub digest: Option<String>,
}

/// `claim-acquire` stdin (typed JSON, unknown keys rejected).
#[derive(Clone, Debug, Deserialize, PartialEq, Eq)]
#[serde(deny_unknown_fields)]
pub struct ClaimRequest {
    pub schema_version: u32,
    pub task_ref: TaskRef,
    pub logical_key: String,
    pub owner: Owner,
    /// #2169 additive: GitHub assignee — display projection only (REQ-B2-01), recorded
    /// verbatim; the claim never writes it to GitHub (REQ-06A-08).
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub assignee: Option<String>,
    /// #2169 additive: the host session identifier of the executing session.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub provider_session: Option<String>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub attempt: Option<u64>,
    /// The exact base commit the attempt is bound to (40 lowercase hex).
    pub base_revision: String,
    pub write_scope: Vec<String>,
    pub spec_ref: SpecRef,
    /// RFC3339 `YYYY-MM-DDTHH:MM:SSZ`. Absent or empty: now + [`DEFAULT_LEASE`] at acquire
    /// (#2166); an explicit value wins.
    #[serde(default)]
    pub lease_until: String,
    /// Recorded as intents in `effects[]` only; never executed by this Issue (REQ-06A-08).
    #[serde(default)]
    pub projection_intents: Vec<Value>,
    /// WHEN given and not the epoch the journal would hand out, the request is rejected as
    /// stale (AC-06A-2).
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub expected_epoch: Option<u64>,
}

impl ClaimRequest {
    /// The required fields; every optional field takes the value serde gives an absent key,
    /// so a built request equals the parsed one (#2256).
    pub fn new(
        schema_version: u32,
        task_ref: TaskRef,
        logical_key: impl Into<String>,
        owner: Owner,
        base_revision: impl Into<String>,
        write_scope: Vec<String>,
        spec_ref: SpecRef,
    ) -> Self {
        ClaimRequest {
            schema_version,
            task_ref,
            logical_key: logical_key.into(),
            owner,
            assignee: None,
            provider_session: None,
            attempt: None,
            base_revision: base_revision.into(),
            write_scope,
            spec_ref,
            lease_until: String::new(),
            projection_intents: Vec::new(),
            expected_epoch: None,
        }
    }

    pub fn assignee(mut self, assignee: impl Into<String>) -> Self {
        self.assignee = Some(assignee.into());
        self
    }

    pub fn provider_session(mut self, provider_session: impl Into<String>) -> Self {
        self.provider_session = Some(provider_session.into());
        self
    }

    pub fn attempt(mut self, attempt: u64) -> Self {
        self.attempt = Some(attempt);
        self
    }

    pub fn lease_until(mut self, lease_until: impl Into<String>) -> Self {
        self.lease_until = lease_until.into();
        self
    }

    pub fn projection_intents(mut self, projection_intents: Vec<Value>) -> Self {
        self.projection_intents = projection_intents;
        self
    }

    pub fn expected_epoch(mut self, expected_epoch: u64) -> Self {
        self.expected_epoch = Some(expected_epoch);
        self
    }
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Eq)]
pub struct WorktreeId {
    pub common_dir: String,
    pub git_dir: String,
    pub path: String,
    /// Full 40-hex HEAD of the work tree at claim time.
    pub head: String,
    /// 06-B/2107 additive: the branch name the claim works on (checked against
    /// `pulls?head=<owner>:<branch>` when reconciling a pr_create effect). Old records
    /// deserialize unchanged; `None` falls back to the full pull-list scan.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub branch: Option<String>,
}

impl WorktreeId {
    /// A work tree identity with no branch (#2256).
    pub fn new(
        common_dir: impl Into<String>,
        git_dir: impl Into<String>,
        path: impl Into<String>,
        head: impl Into<String>,
    ) -> Self {
        WorktreeId {
            common_dir: common_dir.into(),
            git_dir: git_dir.into(),
            path: path.into(),
            head: head.into(),
            branch: None,
        }
    }

    pub fn branch(mut self, branch: impl Into<String>) -> Self {
        self.branch = Some(branch.into());
        self
    }
}

/// One claim record (REQ-06A-01). `comment_id`/`claim_token` are the ⑤ fencing token
/// (REQ-06A-12); `claim_token` is the exact `Moai-Claim: <issue>#<comment-id>` string.
#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Eq)]
pub struct ClaimRecord {
    pub schema_version: u32,
    pub task_ref: TaskRef,
    pub task_key: String,
    pub logical_key: String,
    /// The execution owner — the only one of the four ownership fields the acquire path
    /// compares with the winning claim comment's session (REQ-B2-01). Exclusivity itself
    /// comes from that comment and the journal record, never from a field value alone.
    pub owner: Owner,
    /// #2169 additive: GitHub assignee of the issue — display projection only, never
    /// consulted by the journal or any gate (REQ-06A-08: a claim never writes it).
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub assignee: Option<String>,
    /// #2169 additive: the host session identifier of the executing session.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub provider_session: Option<String>,
    /// #2169 additive: the leader-side integration gate identifier (who owns the
    /// integration gate for this claim). Set by the integration flow, not by acquire.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub integration_owner: Option<String>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub attempt: Option<u64>,
    pub epoch: u64,
    pub base_revision: String,
    pub worktree: WorktreeId,
    pub write_scope: Vec<String>,
    pub spec_ref: SpecRef,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub comment_id: Option<i64>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub claim_token: Option<String>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub lease_until: Option<String>,
    pub state: String,
    #[serde(default)]
    pub effects: Vec<Value>,
    pub last_observed_at: String,
}

/// One `journal.log` line. Public so tests and 06-B tooling can read the authority file.
#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct JournalEvent {
    pub event: String,
    pub record: ClaimRecord,
}

/// sha256(provider|host|repository_id|issue_id), first 16 hex chars — path-safe (plan step 4).
pub fn task_key(t: &TaskRef) -> String {
    let s = format!(
        "{}|{}|{}|{}",
        t.provider, t.host, t.repository_id, t.issue_id
    );
    digest(s.as_bytes())[7..23].to_string()
}

pub fn claim_token(issue_number: i64, comment_id: i64) -> String {
    format!("Moai-Claim: {issue_number}#{comment_id}")
}

/// Branch-name encoding of the token (REQ-06A-14); branches cannot carry `#`.
pub fn claim_branch_suffix(issue_number: i64, comment_id: i64) -> String {
    format!("c{issue_number}-{comment_id}")
}

/// Whether `state` holds the claim (blocks a new acquire). `handed_off` (06-B) is not
/// active: the handoff closes the record and the next acquire writes the new epoch.
pub fn is_active(state: &str) -> bool {
    ACTIVE_STATES.contains(&state)
}

fn is_live(lease_until: &str, now: i64) -> bool {
    parse_utc(lease_until).is_some_and(|t| t > now)
}

pub fn is_key(s: &str) -> bool {
    s.len() == 16
        && s.bytes()
            .all(|b| b.is_ascii_digit() || (b'a'..=b'f').contains(&b))
}

fn bounded(s: &str, max: usize) -> bool {
    !s.is_empty() && s.len() <= max && !s.chars().any(char::is_control)
}

fn valid_github_name(s: &str) -> bool {
    !s.is_empty()
        && s.len() <= 100
        && s.bytes()
            .all(|b| b.is_ascii_alphanumeric() || b"-_.".contains(&b))
        && s != "."
        && s != ".."
}

fn validate_request(req: &ClaimRequest) -> Result<()> {
    if req.schema_version != SCHEMA_VERSION {
        return Err("unsupported request schema_version".into());
    }
    let t = &req.task_ref;
    if t.provider != "github" || t.host != "github.com" {
        return Err("claim task_ref must be provider github host github.com in RC1".into());
    }
    if !valid_github_name(&t.owner) || !valid_github_name(&t.repo) {
        return Err("claim task_ref owner and repo must be GitHub names".into());
    }
    if t.repository_id <= 0 || t.issue_id <= 0 || t.issue_number <= 0 {
        return Err("claim task_ref ids must be positive".into());
    }
    if !bounded(&req.logical_key, 100) {
        return Err("logical_key must be a bounded string without control characters".into());
    }
    if !bounded(&req.owner.kind, 32) || !bounded(&req.owner.id, 128) {
        return Err("owner kind and id must be bounded strings without control characters".into());
    }
    if req.owner.host_id.as_ref().is_some_and(|h| !bounded(h, 128)) {
        return Err("owner host_id must be a bounded string without control characters".into());
    }
    let hex = |s: &str| {
        s.len() == 40
            && s.bytes()
                .all(|b| b.is_ascii_digit() || (b'a'..=b'f').contains(&b))
    };
    if !hex(&req.base_revision) {
        return Err("base_revision must be a 40 lowercase hex commit".into());
    }
    if req.write_scope.len() > 32 || req.write_scope.iter().any(|s| !bounded(s, 256)) {
        return Err("write_scope must be a bounded list of short entries".into());
    }
    if !bounded(&req.spec_ref.id, 128) || !bounded(&req.spec_ref.revision, 128) {
        return Err(
            "spec_ref id and revision must be bounded strings without control characters".into(),
        );
    }
    if let Some(d) = &req.spec_ref.digest {
        let ok = d.len() == 71
            && d.starts_with("sha256:")
            && d[7..]
                .bytes()
                .all(|b| b.is_ascii_digit() || (b'a'..=b'f').contains(&b));
        if !ok {
            return Err("spec_ref digest must be sha256:<64 hex digits>".into());
        }
    }
    if parse_utc(&req.lease_until).is_none() {
        return Err("lease_until must be YYYY-MM-DDTHH:MM:SSZ".into());
    }
    if req.projection_intents.len() > 32 {
        return Err("projection_intents must be a bounded list (32)".into());
    }
    Ok(())
}

// ---- Worktree identity (01-A reuse + a minimal HEAD read) ----

fn git_head(work_tree: &str) -> Result<String> {
    let mut cmd = Command::new("git");
    cmd.arg("--no-optional-locks")
        .arg("-C")
        .arg(work_tree)
        .arg("rev-parse")
        .arg("HEAD");
    // The shared git-environment scrub (Issue #2043) — the same hygiene as `identity`.
    identity::scrub_git_env(&mut cmd);
    let out = cmd
        .env("GIT_OPTIONAL_LOCKS", "0")
        .stdin(Stdio::null())
        .stderr(Stdio::piped())
        .output()
        .map_err(|_| "git unavailable")?;
    if !out.status.success() {
        let stderr = String::from_utf8_lossy(&out.stderr);
        let raw = stderr.lines().next().unwrap_or("").trim();
        let line = raw
            .strip_prefix("fatal:")
            .or_else(|| raw.strip_prefix("error:"))
            .unwrap_or(raw)
            .trim();
        let cut = line.find(['\'', '"', '/', '\\']).unwrap_or(line.len());
        let summary: String = line[..cut].trim_end().chars().take(160).collect();
        let summary = if summary.is_empty() {
            "git rev-parse failed"
        } else {
            &summary
        };
        return Err(format!("git failed: {summary}"));
    }
    let s = String::from_utf8(out.stdout)
        .map_err(|_| "non UTF-8 git output")?
        .trim()
        .to_string();
    let lower_hex = s
        .bytes()
        .all(|b| b.is_ascii_digit() || (b'a'..=b'f').contains(&b));
    if s.len() == 40 && lower_hex {
        Ok(s)
    } else {
        Err("HEAD is not a 40 hex commit".into())
    }
}

/// The checked-out branch of the work tree at claim time (#2126) — the name a later
/// pr_create reconcile can query with `pulls?head=<owner>:<branch>`. `None` on a
/// detached HEAD (git answers `HEAD`) or anything that is not a plain branch name;
/// reconcile then falls back to the full pull-list scan.
fn git_branch(work_tree: &str) -> Result<Option<String>> {
    let mut cmd = Command::new("git");
    cmd.arg("--no-optional-locks")
        .arg("-C")
        .arg(work_tree)
        .arg("rev-parse")
        .arg("--abbrev-ref")
        .arg("HEAD");
    identity::scrub_git_env(&mut cmd);
    let out = cmd
        .env("GIT_OPTIONAL_LOCKS", "0")
        .stdin(Stdio::null())
        .stderr(Stdio::piped())
        .output()
        .map_err(|_| "git unavailable".to_string())?;
    if !out.status.success() {
        // The HEAD read already classified any real git failure; a missing branch name
        // is not one — the claim still records, reconcile just loses the shortcut.
        return Ok(None);
    }
    let s = String::from_utf8(out.stdout)
        .map_err(|_| "non UTF-8 git output")?
        .trim()
        .to_string();
    let plain = !s.is_empty()
        && s != "HEAD"
        && s.len() <= 200
        && !s.chars().any(|c| c.is_control() || c.is_whitespace());
    if plain {
        Ok(Some(s))
    } else {
        Ok(None)
    }
}

pub(crate) fn worktree_id(identity: &Value) -> Result<WorktreeId> {
    let path = identity["work_tree"]
        .as_str()
        .ok_or("claim journal requires a git work tree")?;
    // Same evaluation order as the former literal, so the first failing source wins as before.
    let common_dir = identity["clone_anchor"]
        .as_str()
        .ok_or("claim journal requires a git repository")?;
    let git_dir = identity["workspace_git_dir"]
        .as_str()
        .ok_or("claim journal requires a git repository")?;
    let mut id = WorktreeId::new(common_dir, git_dir, path, git_head(path)?);
    id.branch = git_branch(path)?;
    Ok(id)
}

// ---- Comment claim (⑤) ----

/// One parsed `moai-claim:` line; `comment_id` is attached by the reader.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct ClaimComment {
    pub comment_id: i64,
    pub session: String,
    pub epoch: u64,
    pub lease_until: String,
}

/// Parses the canonical claim line wherever it stands in a comment body.
pub fn parse_claim_line(body: &str) -> Option<ClaimComment> {
    for line in body.lines() {
        let Some(rest) = line.trim().strip_prefix("moai-claim:") else {
            continue;
        };
        let (mut session, mut epoch, mut lease) = (None, None, None);
        for tok in rest.split_whitespace() {
            if let Some(v) = tok.strip_prefix("session=") {
                session = Some(v.to_string());
            } else if let Some(v) = tok.strip_prefix("epoch=") {
                epoch = v.parse().ok();
            } else if let Some(v) = tok.strip_prefix("lease_until=") {
                lease = Some(v.to_string());
            }
        }
        if let (Some(session), Some(epoch), Some(lease_until)) = (session, epoch, lease) {
            return Some(ClaimComment {
                comment_id: 0,
                session,
                epoch,
                lease_until,
            });
        }
    }
    None
}

/// REQ-06A-13: among live claims the smallest comment id wins.
pub fn resolve_winner(live: &[ClaimComment]) -> Option<ClaimComment> {
    live.iter().min_by_key(|c| c.comment_id).cloned()
}

fn err_transport(message: impl Into<String>) -> ForgeError {
    ForgeError {
        kind: "transport",
        message: format!("gh transport: {}", message.into()),
        retryable: true,
    }
}

/// Fields of a claim comment line (#2255, REQ-A9-02). The first three are the RC1 line and
/// stay first, so every reader of the RC1 form keeps working; the rest are appended tokens.
#[derive(Clone, Debug, Default, PartialEq, Eq)]
pub struct ClaimFields {
    /// Session id of the claimant; `unverified` when the poster could not confirm it.
    pub session: String,
    pub epoch: u64,
    pub lease_until: String,
    /// Lane label (`agent-1`, ...), when known.
    pub lane: Option<String>,
    pub issue: Option<i64>,
    pub branch: Option<String>,
    /// UTC time the line was written (`YYYY-MM-DDTHH:MM:SSZ`).
    pub at: Option<String>,
    /// True when another session (the leader) posted the line on the claimant's behalf.
    pub relay: bool,
}

/// The one claim-line generator: acquire, queue-next and a relaying leader all render
/// through it, and `claim_verify::classify_comment` parses what it writes.
pub fn render_claim_line(f: &ClaimFields) -> String {
    let mut line = format!(
        "moai-claim: session={} epoch={} lease_until={}",
        f.session, f.epoch, f.lease_until
    );
    if let Some(lane) = &f.lane {
        line.push_str(&format!(" lane={lane}"));
    }
    if let Some(issue) = f.issue {
        line.push_str(&format!(" issue={issue}"));
    }
    if let Some(branch) = &f.branch {
        line.push_str(&format!(" branch={branch}"));
    }
    if let Some(at) = &f.at {
        line.push_str(&format!(" at={at}"));
    }
    // Absent means not relayed, so a line with no extra fields is the RC1 line byte for byte.
    if f.relay {
        line.push_str(" relay=true");
    }
    line
}

/// Reads every comment page and returns claims not explicitly closed.
/// Release/handoff markers can close only an earlier referenced claim comment.
fn read_claims<T: Transport, C: Clock>(
    client: &Client<'_, T, C>,
    t: &TaskRef,
) -> std::result::Result<Vec<ClaimComment>, ForgeError> {
    let mut url = Some(format!(
        "https://api.github.com/repos/{}/{}/issues/{}/comments?per_page=100",
        t.owner, t.repo, t.issue_number
    ));
    let mut claims = Vec::new();
    let mut closed = BTreeSet::new();
    let mut pages = 0;
    while let Some(u) = url {
        pages += 1;
        if pages > CLAIM_MAX_PAGES {
            return Err(ForgeError {
                kind: "invalid_input",
                message: "claim comment page limit exceeded (50)".into(),
                retryable: false,
            });
        }
        let resp = client.get(&u, None)?;
        let arr = resp
            .body
            .as_ref()
            .and_then(Value::as_array)
            .cloned()
            .unwrap_or_default();
        for c in &arr {
            let (Some(id), Some(body)) = (c["id"].as_i64(), c["body"].as_str()) else {
                continue;
            };
            match crate::claim_verify::classify_comment(body) {
                crate::claim_verify::CommentKind::Claim {
                    session,
                    epoch,
                    lease_until,
                } => claims.push(ClaimComment {
                    comment_id: id,
                    session,
                    epoch,
                    lease_until,
                }),
                crate::claim_verify::CommentKind::Release { target }
                | crate::claim_verify::CommentKind::Handoff { target, .. }
                    if target < id =>
                {
                    closed.insert(target);
                }
                _ => {}
            }
        }
        url = resp.header("link").and_then(|l| next_link(Some(l)));
    }
    claims.retain(|c| !closed.contains(&c.comment_id));
    Ok(claims)
}

/// The one allowed GitHub write of a claim (REQ-06A-12): exactly one comment POST.
fn post_claim<T: Transport, C: Clock>(
    client: &Client<'_, T, C>,
    t: &TaskRef,
    line: &str,
) -> std::result::Result<i64, ForgeError> {
    let url = format!(
        "https://api.github.com/repos/{}/{}/issues/{}/comments",
        t.owner, t.repo, t.issue_number
    );
    let req = HttpRequest::new("POST", url).with_body(json!({"body": line}));
    let resp = client.send_with_backoff(&req)?;
    if !(200..300).contains(&resp.status) {
        return Err(if resp.status == 404 {
            ForgeError {
                kind: "not_found",
                message: "forge resource not found or not visible to this account (404)".into(),
                retryable: false,
            }
        } else {
            err_transport(format!("claim comment POST returned HTTP {}", resp.status))
        });
    }
    resp.body
        .as_ref()
        .and_then(|b| b["id"].as_i64())
        .ok_or_else(|| err_transport("claim comment response carried no id"))
}

// ---- Journal I/O ----

fn refuse_symlink(path: &Path) -> Result<()> {
    if fs::symlink_metadata(path)
        .map(|m| m.file_type().is_symlink())
        .unwrap_or(false)
    {
        return Err("journal path is a symlink; refused".into());
    }
    Ok(())
}

/// The journal handle. `lock: None` is the read-only view (`show`): it creates nothing.
#[derive(Debug)]
pub struct Journal {
    dir: PathBuf,
    lock: Option<fs::File>,
}

impl Journal {
    /// Read-only view; never creates a file or directory.
    pub fn open_read(root: &Path) -> Result<Journal> {
        refuse_symlink(root)?;
        Ok(Journal {
            dir: root.to_path_buf(),
            lock: None,
        })
    }

    /// Opens (creating directories and `.lock`) for a writing command; [`Journal::try_lock`]
    /// follows. Symlinked paths are refused (REQ-06A-07); a symlink swap between the check and
    /// the open is a documented residual risk of the advisory local model.
    pub fn open(root: &Path) -> Result<Journal> {
        refuse_symlink(root)?;
        if root.exists() && !root.is_dir() {
            return Err("journal path is not a directory".into());
        }
        fs::create_dir_all(root).map_err(|_| "journal directory unavailable".to_string())?;
        let lock_path = root.join(".lock");
        refuse_symlink(&lock_path)?;
        let lock = OpenOptions::new()
            .read(true)
            .write(true)
            .create(true)
            .truncate(false)
            .open(&lock_path)
            .map_err(|_| "journal lock unavailable".to_string())?;
        Ok(Journal {
            dir: root.to_path_buf(),
            lock: Some(lock),
        })
    }

    /// REQ-06A-05: never waits. `Ok(true)` holds the lock until the handle drops;
    /// `Ok(false)` is `busy` — the waiting policy is the caller's.
    pub fn try_lock(&self) -> Result<bool> {
        match self.lock.as_ref().map(|f| f.try_lock()) {
            Some(Ok(())) => Ok(true),
            Some(Err(std::fs::TryLockError::WouldBlock)) => Ok(false),
            Some(Err(std::fs::TryLockError::Error(_))) => Err("journal lock unavailable".into()),
            None => Err("journal was opened read-only; lock needs Journal::open".into()),
        }
    }

    /// Every complete line; a trailing partial line (no newline) is dropped — the
    /// "last complete line" rule of the design. A corrupt complete line is an error.
    pub fn read_events(&self) -> Result<Vec<JournalEvent>> {
        let path = self.dir.join("journal.log");
        refuse_symlink(&path)?;
        let bytes = match fs::read(&path) {
            Ok(b) => b,
            Err(e) if e.kind() == std::io::ErrorKind::NotFound => return Ok(Vec::new()),
            Err(_) => return Err("claim journal read failed".into()),
        };
        let text = String::from_utf8(bytes).map_err(|_| "claim journal read failed".to_string())?;
        let segs: Vec<&str> = text.split('\n').collect();
        // The last element is "" after a final newline, or a partial line without one;
        // it is never a complete record, so it is dropped either way.
        let complete = &segs[..segs.len().saturating_sub(1)];
        let mut out = Vec::new();
        for (i, seg) in complete.iter().enumerate() {
            if seg.trim().is_empty() {
                continue;
            }
            match serde_json::from_str::<JournalEvent>(seg) {
                Ok(ev) => out.push(ev),
                Err(_) => return Err(format!("claim journal read failed at line {}", i + 1)),
            }
        }
        Ok(out)
    }

    /// Commits one journal event: the append (authority, fsync) lands first, then the
    /// record file (cache) is replaced atomically. The single write surface for every
    /// consumer that records effects, handoffs or transitions on a record (06-B and later;
    /// tests seed journals through it).
    pub fn commit(&self, event: &str, record: &ClaimRecord) -> Result<()> {
        append_event(&self.dir, event, record)?;
        write_record_file(&self.dir, record)
    }
}

fn append_event(dir: &Path, event: &str, record: &ClaimRecord) -> Result<()> {
    let path = dir.join("journal.log");
    refuse_symlink(&path)?;
    let mut f = OpenOptions::new()
        .create(true)
        .append(true)
        .open(&path)
        .map_err(|_| "claim journal write failed".to_string())?;
    let ev = JournalEvent {
        event: event.to_string(),
        record: record.clone(),
    };
    let mut line = serde_json::to_string(&ev).map_err(|_| "encode failed".to_string())?;
    line.push('\n');
    f.write_all(line.as_bytes())
        .and_then(|_| f.sync_all())
        .map_err(|_| "claim journal write failed".to_string())
}

/// Temp file + fsync + rename (REQ-06A-04). Crash semantics: `journal.log` is appended and
/// fsynced FIRST, so the authority always carries the newer of the two writes.
fn write_record_file(dir: &Path, record: &ClaimRecord) -> Result<()> {
    let path = dir.join(format!("{}.json", record.task_key));
    refuse_symlink(&path)?;
    let tmp = dir.join(format!(
        "{}.json.tmp-{}",
        record.task_key,
        std::process::id()
    ));
    let bytes = serde_json::to_vec_pretty(record).map_err(|_| "encode failed".to_string())?;
    let mut f = OpenOptions::new()
        .write(true)
        .create_new(true)
        .open(&tmp)
        .map_err(|_| "claim journal write failed".to_string())?;
    let write = f.write_all(&bytes).and_then(|_| f.sync_all());
    if write.is_err() {
        let _ = fs::remove_file(&tmp);
        return Err("claim journal write failed".into());
    }
    match fs::rename(&tmp, &path) {
        Ok(()) => Ok(()),
        Err(_) => {
            let _ = fs::remove_file(&tmp);
            Err("claim record replace failed".into())
        }
    }
}

// ---- Filesystem honesty (REQ-06A-07) ----

/// Best-effort filesystem type of the journal directory. macOS `stat -f %T` and coreutils
/// `stat -f -c %T` are both tried; when neither answers, the honest fallback says so.
fn fs_kind_note(dir: &Path) -> String {
    let arg_sets: [&[&str]; 2] = [&["-f", "%T"], &["-f", "-c", "%T"]];
    for args in arg_sets {
        let out = Command::new("stat")
            .args(args)
            .arg(dir)
            .env("LC_ALL", "C")
            .stdin(Stdio::null())
            .stderr(Stdio::piped())
            .output();
        if let Ok(out) = out {
            if out.status.success() {
                if let Ok(s) = String::from_utf8(out.stdout) {
                    return fs_note(&s.trim().to_ascii_lowercase());
                }
            }
        }
    }
    "journal filesystem type unknown; atomicity and locking are guaranteed only on local filesystems".into()
}

fn fs_note(kind: &str) -> String {
    let network = ["nfs", "cifs", "smb", "afp", "webdav", "9p", "fuse"]
        .iter()
        .any(|k| kind.contains(k));
    if network {
        format!("journal path is on network filesystem '{kind}'; atomicity and advisory locking are not guaranteed")
    } else {
        format!("journal filesystem '{kind}'; locking is advisory and local")
    }
}

// ---- Operations ----

#[derive(Debug)]
pub struct AcquireOutcome {
    pub outcome: &'static str,
    pub record: Option<ClaimRecord>,
    pub foreign_owner: Option<Value>,
    pub notes: Vec<String>,
}

fn scope_of(outcome: &str, record: Option<&ClaimRecord>) -> &'static str {
    // REQ-06A-11: a claim decided by a comment is cross-host (eventually); the journal-only
    // paths are single-host by construction.
    match outcome {
        "claimed" => "cross-host-eventual",
        "already_owned" => {
            if record.is_none() {
                // lost to another host's live comment
                "cross-host-eventual"
            } else {
                "single-host"
            }
        }
        // The winner is unresolved on the comment ledger, which is the cross-host layer.
        "uncertain" => "cross-host-eventual",
        _ => "single-host",
    }
}

impl AcquireOutcome {
    pub(crate) fn to_json(&self) -> Value {
        json!({
            "outcome": self.outcome,
            "record": self.record.as_ref().map(|r| json!(r)).unwrap_or(Value::Null),
            "foreign_owner": self.foreign_owner.clone().unwrap_or(Value::Null),
            "coordination_scope": scope_of(self.outcome, self.record.as_ref()),
            "cross_host_atomic": false,
            "notes": self.notes,
        })
    }
}

/// The acquire body; the caller holds the journal lock. Journal owner check first
/// (REQ-06A-03), then the comment claim (REQ-06A-12), then the journal append.
pub fn acquire<T: Transport, C: Clock>(
    journal: &Journal,
    req: &ClaimRequest,
    worktree: WorktreeId,
    client: &Client<'_, T, C>,
    clock: &C,
) -> Result<AcquireOutcome> {
    let now = clock.now();
    // #2166: an absent lease becomes now + DEFAULT_LEASE; an explicit one is kept as is.
    let filled;
    let req = if req.lease_until.is_empty() {
        filled = ClaimRequest {
            lease_until: format_utc(now + DEFAULT_LEASE),
            ..req.clone()
        };
        &filled
    } else {
        req
    };
    validate_request(req)?;
    let key = task_key(&req.task_ref);
    let events = journal.read_events()?;
    let mine: Vec<&ClaimRecord> = events
        .iter()
        .filter(|e| e.record.task_key == key)
        .map(|e| &e.record)
        .collect();
    if let Some(&existing) = mine.last() {
        if is_active(&existing.state) {
            // Never changed, never taken over (REQ-06A-03, REQ-06A-16).
            let expired = existing
                .lease_until
                .as_deref()
                .is_some_and(|l| !is_live(l, now));
            return Ok(if expired {
                AcquireOutcome {
                    outcome: "expired",
                    record: Some(existing.clone()),
                    foreign_owner: None,
                    notes: vec![
                        "the active record's lease has expired; takeover needs operator approval (06-B)".into(),
                    ],
                }
            } else {
                AcquireOutcome {
                    outcome: "already_owned",
                    record: Some(existing.clone()),
                    foreign_owner: None,
                    notes: vec![],
                }
            });
        }
    }
    let epoch = mine.iter().map(|r| r.epoch).max().unwrap_or(0) + 1;
    if let Some(want) = req.expected_epoch {
        if want != epoch {
            return Err(format!(
                "conflict in claim journal: stale epoch {want}; journal epoch would be {epoch}"
            ));
        }
    }
    // ⑤: read before posting; a live foreign claim wins by the smaller comment id.
    let claims = read_claims(client, &req.task_ref).map_err(String::from)?;
    let live: Vec<ClaimComment> = claims
        .into_iter()
        .filter(|c| is_live(&c.lease_until, now))
        .collect();
    let adopted;
    let comment_id = match resolve_winner(&live) {
        Some(winner) if winner.session != req.owner.id => {
            return Ok(AcquireOutcome {
                outcome: "already_owned",
                record: None,
                foreign_owner: Some(json!({
                    "session": winner.session, "epoch": winner.epoch,
                    "comment_id": winner.comment_id, "lease_until": winner.lease_until,
                })),
                notes: vec![
                    "a live claim comment with a smaller comment id is the winner; nothing was written".into(),
                ],
            });
        }
        // The winner carries our session value. A comment body proves nothing about who
        // posted it (every lane shares one forge account), so it is adopted only when this
        // journal recorded posting that very comment; otherwise a same-named live session
        // may own it (#2165 REQ-A6-01, REQ-AM-04).
        Some(winner) if !mine.iter().any(|r| r.comment_id == Some(winner.comment_id)) => {
            return Ok(AcquireOutcome {
                outcome: "uncertain",
                record: None,
                foreign_owner: Some(json!({
                    "session": winner.session, "epoch": winner.epoch,
                    "comment_id": winner.comment_id, "lease_until": winner.lease_until,
                })),
                notes: vec![format!(
                    "the winning live claim comment (id {}) carries this session value but this \
                     clone's journal has no record of posting it; it was not adopted — another \
                     live session may share the name (REQ-AM-04); nothing was written",
                    winner.comment_id
                )],
            });
        }
        Some(winner) => {
            adopted = true;
            winner.comment_id
        }
        None => {
            adopted = false;
            post_claim(
                client,
                &req.task_ref,
                // The acquire line stays the RC1 line byte for byte (#2255 AC-3): the issue is
                // the URL's, and the comment's own created_at is the server's time.
                &render_claim_line(&ClaimFields {
                    session: req.owner.id.clone(),
                    epoch,
                    lease_until: req.lease_until.clone(),
                    ..Default::default()
                }),
            )
            .map_err(String::from)?
        }
    };
    // The POST does not settle a cross-host race: another claimant can pass the same
    // pre-check and post between our GET and our POST (REQ-06A-13). The re-read decides —
    // the smallest live comment id wins. A loser keeps its posted comment (it is inert
    // under the rule, and deleting it is no authorized claim write) but records nothing.
    let comment_id = if adopted {
        comment_id
    } else {
        match read_claims(client, &req.task_ref) {
            Err(e) => {
                return Ok(AcquireOutcome {
                    outcome: "uncertain",
                    record: None,
                    foreign_owner: None,
                    notes: vec![format!(
                        "the comment was posted (id {comment_id}) but the re-read failed ({}); \
                             no journal record was written and the winner is unresolved — \
                             re-check with claim-show and the issue comments",
                        e.message
                    )],
                });
            }
            Ok(claims2) => {
                let live2: Vec<ClaimComment> = claims2
                    .into_iter()
                    .filter(|c| is_live(&c.lease_until, now))
                    .collect();
                // Our own POST returned an id, so the comment exists; if the re-read does
                // not show it, the list cannot be trusted to judge the winner.
                let Some(ours) = live2.iter().find(|c| c.comment_id == comment_id) else {
                    return Ok(AcquireOutcome {
                        outcome: "uncertain",
                        record: None,
                        foreign_owner: None,
                        notes: vec![format!(
                            "the comment was posted (id {comment_id}) but the re-read does not \
                             show it; without read-your-writes visibility the winner cannot be \
                             judged — no journal record was written, re-check with claim-show"
                        )],
                    });
                };
                match resolve_winner(&live2) {
                    Some(winner) if winner.comment_id != ours.comment_id => {
                        return Ok(AcquireOutcome {
                            outcome: "already_owned",
                            record: None,
                            foreign_owner: Some(json!({
                                "session": winner.session, "epoch": winner.epoch,
                                "comment_id": winner.comment_id, "lease_until": winner.lease_until,
                            })),
                            notes: vec![format!(
                                "a concurrent claimant posted the smaller live comment id {winner_id}; \
                                 our posted comment id {mine} remains on the issue and is inert under \
                                 the smallest-live-id rule; no journal record was written",
                                winner_id = winner.comment_id,
                                mine = comment_id,
                            )],
                        });
                    }
                    _ => comment_id,
                }
            }
        }
    };
    let token = claim_token(req.task_ref.issue_number, comment_id);
    let mut effects = req.projection_intents.clone();
    effects.push(json!({
        "kind": if adopted { "comment_adopted" } else { "comment_posted" },
        "comment_id": comment_id,
    }));
    let record = ClaimRecord {
        schema_version: SCHEMA_VERSION,
        task_ref: req.task_ref.clone(),
        task_key: key.clone(),
        logical_key: req.logical_key.clone(),
        owner: req.owner.clone(),
        assignee: req.assignee.clone(),
        provider_session: req.provider_session.clone(),
        // REQ-B2-01: the integration owner is the leader gate's identity — the acquire
        // path never sets it; the integration flow (leader) records it later.
        integration_owner: None,
        attempt: req.attempt,
        epoch,
        base_revision: req.base_revision.clone(),
        worktree,
        write_scope: req.write_scope.clone(),
        spec_ref: req.spec_ref.clone(),
        comment_id: Some(comment_id),
        claim_token: Some(token),
        lease_until: Some(req.lease_until.clone()),
        state: "assigned".into(),
        effects,
        last_observed_at: format_utc(now),
    };
    // The append (authority) lands before the record file (cache).
    append_event(&journal.dir, "acquire", &record)?;
    write_record_file(&journal.dir, &record)?;
    let mut notes = vec![
        "the journal lock is advisory: processes that bypass MoAI are not serialized".into(),
        fs_kind_note(&journal.dir),
    ];
    if adopted {
        notes.push(
            "a live claim comment of this session already existed and was adopted; no comment was posted".into(),
        );
    }
    Ok(AcquireOutcome {
        outcome: "claimed",
        record: Some(record),
        foreign_owner: None,
        notes,
    })
}

/// `claim-release`: closes the active record as `cancelled` (REQ-06A-16 takeover stays with
/// the operator). A stale epoch is a conflict (AC-06A-2).
pub fn release<C: Clock>(
    journal: &Journal,
    key: &str,
    epoch: u64,
    reason: &str,
    clock: &C,
) -> Result<AcquireOutcome> {
    if !is_key(key) {
        return Err("task must be a 16 hex character claim key".into());
    }
    let events = journal.read_events()?;
    let existing = events
        .iter()
        .rfind(|e| e.record.task_key == key)
        .map(|e| e.record.clone());
    let Some(rec) = existing else {
        return Ok(AcquireOutcome {
            outcome: "unclaimed",
            record: None,
            foreign_owner: None,
            notes: vec!["no record exists for this task key".into()],
        });
    };
    if !is_active(&rec.state) {
        return Ok(AcquireOutcome {
            outcome: "unclaimed",
            record: Some(rec),
            foreign_owner: None,
            notes: vec!["the record is not in an active state".into()],
        });
    }
    if rec.epoch != epoch {
        return Err(format!(
            "conflict in claim journal: stale epoch {epoch}; current record epoch is {}",
            rec.epoch
        ));
    }
    let now = clock.now();
    let mut new = rec.clone();
    new.state = "cancelled".into();
    new.last_observed_at = format_utc(now);
    new.effects
        .push(json!({"kind": "release", "reason": reason}));
    append_event(&journal.dir, "release", &new)?;
    write_record_file(&journal.dir, &new)?;
    Ok(AcquireOutcome {
        outcome: "released",
        record: Some(new),
        foreign_owner: None,
        notes: vec![],
    })
}

/// `show` plus `expired` against `now` (#2166): true only for an active record whose lease
/// has passed. Display only — ownership never moves on expiry (REQ-A7-02). Byte-identical
/// when repeated at the same `now`; no lock, no write.
pub fn show_at(root: &Path, key: &str, now: i64) -> Result<Value> {
    let mut v = show(root, key)?;
    let expired = v["record"]["state"].as_str().is_some_and(is_active)
        && v["record"]["lease_until"]
            .as_str()
            .is_some_and(|l| !is_live(l, now));
    v["expired"] = json!(expired);
    Ok(v)
}

/// `claim-show`: byte-identical when repeated (REQ-06A-10) — no lock, no write, no clock.
pub fn show(root: &Path, key: &str) -> Result<Value> {
    if !is_key(key) {
        return Err("task must be a 16 hex character claim key".into());
    }
    let journal = Journal::open_read(root)?;
    let events = journal.read_events()?;
    let record = events
        .iter()
        .rfind(|e| e.record.task_key == key)
        .map(|e| e.record.clone());
    let outcome = if record.is_some() {
        "held"
    } else {
        "unclaimed"
    };
    let scope = match &record {
        Some(r) if r.comment_id.is_some() => "cross-host-eventual",
        _ => "single-host",
    };
    Ok(json!({
        "outcome": outcome,
        "record": record.map(|r| json!(r)).unwrap_or(Value::Null),
        "coordination_scope": scope,
        "cross_host_atomic": false,
    }))
}

// ---- Leader-side token verification (REQ-06A-15) ----

#[derive(Clone, Debug, PartialEq, Eq)]
pub struct CommitCarrier {
    pub sha: String,
    /// The `Moai-Claim:` trailer text exactly as the commit carries it.
    pub moai_claim_trailer: Option<String>,
}

#[derive(Clone, Debug, Default, PartialEq, Eq)]
pub struct TokenCarriers {
    /// The PR head branch name; `None` fails the verdict (it is always knowable to the leader).
    pub branch: Option<String>,
    pub commits: Vec<CommitCarrier>,
    pub pr_body: String,
}

#[derive(Clone, Debug, PartialEq, Eq)]
pub struct TokenVerdict {
    pub pass: bool,
    pub reasons: Vec<String>,
}

/// Checks that branch name, every commit trailer and the PR body carry exactly the winner's
/// token (REQ-06A-14/15, AC-06A-9/12). Any absent or different token is a reason; a different
/// token is a fencing violation, not a near-miss.
pub fn verify_claim_token(
    issue_number: i64,
    comment_id: i64,
    carriers: &TokenCarriers,
) -> TokenVerdict {
    let token = claim_token(issue_number, comment_id);
    let suffix = claim_branch_suffix(issue_number, comment_id);
    let mut reasons = Vec::new();
    match &carriers.branch {
        Some(b) if b.contains(&suffix) => {}
        Some(_) => reasons.push(format!(
            "branch name does not carry the claim token suffix {suffix}"
        )),
        None => reasons.push("branch name missing from carriers".into()),
    }
    if carriers.commits.is_empty() {
        reasons.push("no commits supplied".into());
    }
    for c in &carriers.commits {
        match &c.moai_claim_trailer {
            Some(t) if t == &token => {}
            Some(t) => reasons.push(format!(
                "commit {} carries a different claim token ({t})",
                c.sha
            )),
            None => reasons.push(format!("commit {} carries no Moai-Claim trailer", c.sha)),
        }
    }
    if !carriers.pr_body.contains(&token) {
        reasons.push("PR body does not carry the claim token".into());
    }
    TokenVerdict {
        pass: reasons.is_empty(),
        reasons,
    }
}

// ---- CLI entries (`main` wraps the results in the 01-C envelope) ----

/// The one journal lock-busy answer (Issue #2345): every journal-writing operation returns
/// this message with the explicit `unavailable` kind, retryable, exit 1, when another
/// process holds the lock. The tool never waits (REQ-06A-05); the waiting policy stays the
/// caller's, so the message names the retry rule (a bounded wait within the caller's cap).
pub const JOURNAL_LOCK_BUSY: &str = "journal busy: the claim journal lock is unavailable (another process holds it); the caller may retry the same command after a bounded wait, within its own cap (REQ-06A-05)";

/// CLI acquire: identity + worktree, journal open + try_lock (busy never waits), then
/// [`acquire`]. The transport and clock come from the caller like `forge_read_cli`.
pub fn acquire_cli<T: Transport, C: Clock>(
    project: &Path,
    req: ClaimRequest,
    transport: &T,
    clock: &C,
) -> std::result::Result<Value, OpError> {
    let root = journal_root(project)?;
    let id = identity::identity(project)?;
    let worktree = worktree_id(&id)?;
    let journal = Journal::open(&root)?;
    if !journal.try_lock()? {
        // #2345: the kind is set here, at the raise site — classify() never guesses it.
        return Err(OpError::new(Kind::Unavailable, JOURNAL_LOCK_BUSY));
    }
    let client = Client::new(transport, clock, Default::default(), Default::default());
    let out = acquire(&journal, &req, worktree, &client, clock)?;
    Ok(out.to_json())
}

pub fn show_cli(project: &Path, key: &str) -> Result<Value> {
    let root = journal_root(project)?;
    show_at(&root, key, crate::forge::policy::SystemClock.now())
}

pub fn release_cli<C: Clock>(
    project: &Path,
    key: &str,
    epoch: u64,
    reason: &str,
    clock: &C,
) -> std::result::Result<Value, OpError> {
    let root = journal_root(project)?;
    let journal = Journal::open(&root)?;
    if !journal.try_lock()? {
        return Err(OpError::new(Kind::Unavailable, JOURNAL_LOCK_BUSY));
    }
    let out = release(&journal, key, epoch, reason, clock)?;
    Ok(out.to_json())
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::contract;

    fn task_ref() -> TaskRef {
        TaskRef::new("github", "github.com", "acme", "widgets", 101, 7001, 7)
    }

    #[test]
    fn task_key_is_stable_and_id_distinct() {
        let a = task_ref();
        // Display names are not key material: a renamed owner keeps the key.
        let mut renamed = task_ref();
        renamed.owner = "other".to_string();
        assert_eq!(task_key(&a), task_key(&renamed));
        let mut other_repo = task_ref();
        other_repo.repository_id = 202;
        assert_ne!(task_key(&a), task_key(&other_repo));
        let key = task_key(&a);
        assert_eq!(key.len(), 16);
        assert!(key
            .bytes()
            .all(|c| c.is_ascii_hexdigit() && !c.is_ascii_uppercase()));
    }

    #[test]
    fn token_shapes_are_canonical() {
        assert_eq!(claim_token(2005, 6720575543), "Moai-Claim: 2005#6720575543");
        assert_eq!(claim_branch_suffix(2005, 6720575543), "c2005-6720575543");
    }

    #[test]
    fn claim_line_parses_from_prose() {
        let body = "leader relay post\n\nmoai-claim: session=agent-6-742cb6 epoch=1 lease_until=2026-11-08T13:47:52Z";
        let c = parse_claim_line(body).unwrap();
        assert_eq!(c.session, "agent-6-742cb6");
        assert_eq!(c.epoch, 1);
        assert_eq!(c.lease_until, "2026-11-08T13:47:52Z");
        assert!(parse_claim_line("no claim here").is_none());
        assert!(parse_claim_line("moai-claim: session=x epoch=oops lease_until=y").is_none());
    }

    #[test]
    fn winner_is_the_smallest_live_comment_id() {
        let c = |id: i64, session: &str, lease: &str| ClaimComment {
            comment_id: id,
            session: session.into(),
            epoch: 1,
            lease_until: lease.into(),
        };
        let now = 1000;
        let live: Vec<ClaimComment> = [
            c(200, "b", "1970-01-01T00:16:40Z"),
            c(100, "a", "1970-01-01T00:30:00Z"),
        ]
        .into_iter()
        .filter(|x| is_live(&x.lease_until, now))
        .collect();
        assert_eq!(resolve_winner(&live).unwrap().comment_id, 100);
        // An expired comment (lease 1970-01-01T00:16:39Z) is not live.
        assert!(!is_live("1970-01-01T00:16:39Z", now));
        assert_eq!(resolve_winner(&[]), None);
    }

    fn carriers(branch: Option<&str>, trailers: &[Option<&str>], pr_body: &str) -> TokenCarriers {
        TokenCarriers {
            branch: branch.map(str::to_string),
            commits: trailers
                .iter()
                .enumerate()
                .map(|(i, t)| CommitCarrier {
                    sha: format!("{i:040}"),
                    moai_claim_trailer: t.map(str::to_string),
                })
                .collect(),
            pr_body: pr_body.into(),
        }
    }

    #[test]
    fn token_verification_passes_only_when_all_three_carry_it() {
        let token = claim_token(2005, 42);
        let good = carriers(
            Some("issue/2005-06-A-c2005-42"),
            &[Some(&token), Some(&token)],
            &format!("Refs #2005\n{token}"),
        );
        assert!(verify_claim_token(2005, 42, &good).pass);

        let v = verify_claim_token(2005, 42, &carriers(None, &[Some(&token)], &token));
        assert!(!v.pass);
        assert!(v.reasons[0].contains("branch name missing"));

        let v = verify_claim_token(
            2005,
            42,
            &carriers(Some("plain-branch"), &[Some(&token)], &token),
        );
        assert!(!v.pass && v.reasons[0].contains("suffix"));

        let v = verify_claim_token(
            2005,
            42,
            &carriers(Some("issue/2005-06-A-c2005-42"), &[None], &token),
        );
        assert!(!v.pass && v.reasons[0].contains("no Moai-Claim trailer"));

        let v = verify_claim_token(
            2005,
            42,
            &carriers(
                Some("issue/2005-06-A-c2005-42"),
                &[Some("Moai-Claim: 2005#7")],
                &token,
            ),
        );
        assert!(!v.pass && v.reasons[0].contains("different claim token"));

        let v = verify_claim_token(
            2005,
            42,
            &carriers(
                Some("issue/2005-06-A-c2005-42"),
                &[Some(&token)],
                "no token here",
            ),
        );
        assert!(!v.pass && v.reasons[0].contains("PR body"));
    }

    #[test]
    fn fs_note_warns_only_on_network_filesystems() {
        let local = fs_note("apfs");
        assert!(local.contains("advisory") && !local.contains("not guaranteed"));
        let nfs = fs_note("nfs");
        assert!(nfs.contains("not guaranteed") && nfs.contains("nfs"));
        assert!(fs_note("fuseblk").contains("not guaranteed"));
    }

    #[test]
    fn every_error_message_has_a_deliberate_kind() {
        use contract::Kind;
        let cases = [
            (
                "unsupported request schema_version",
                Kind::UnsupportedProtocolVersion,
            ),
            (
                "claim task_ref must be provider github host github.com in RC1",
                Kind::InvalidInput,
            ),
            (
                "claim task_ref owner and repo must be GitHub names",
                Kind::InvalidInput,
            ),
            ("claim task_ref ids must be positive", Kind::InvalidInput),
            (
                "logical_key must be a bounded string without control characters",
                Kind::InvalidInput,
            ),
            (
                "owner kind and id must be bounded strings without control characters",
                Kind::InvalidInput,
            ),
            (
                "owner host_id must be a bounded string without control characters",
                Kind::InvalidInput,
            ),
            (
                "base_revision must be a 40 lowercase hex commit",
                Kind::InvalidInput,
            ),
            (
                "write_scope must be a bounded list of short entries",
                Kind::InvalidInput,
            ),
            (
                "spec_ref id and revision must be bounded strings without control characters",
                Kind::InvalidInput,
            ),
            (
                "spec_ref digest must be sha256:<64 hex digits>",
                Kind::InvalidInput,
            ),
            (
                "lease_until must be YYYY-MM-DDTHH:MM:SSZ",
                Kind::InvalidInput,
            ),
            (
                "projection_intents must be a bounded list (32)",
                Kind::InvalidInput,
            ),
            (
                "claim journal requires a git repository; identity_status was no_git",
                Kind::InvalidInput,
            ),
            ("claim journal requires a git work tree", Kind::InvalidInput),
            ("journal path is a symlink; refused", Kind::InvalidInput),
            ("journal path is not a directory", Kind::InvalidInput),
            ("HEAD is not a 40 hex commit", Kind::InvalidInput),
            (
                "task must be a 16 hex character claim key",
                Kind::InvalidInput,
            ),
            ("claim comment page limit exceeded (50)", Kind::InvalidInput),
            ("git failed: dubious ownership", Kind::GitFailed),
            ("git unavailable", Kind::Io),
            ("git output was not UTF-8", Kind::Io),
            ("journal directory unavailable", Kind::Io),
            ("journal lock unavailable", Kind::Io),
            ("claim journal read failed", Kind::Io),
            ("claim journal read failed at line 3", Kind::Io),
            ("claim journal write failed", Kind::Io),
            ("claim record replace failed", Kind::Io),
            ("non UTF-8 git output", Kind::Io),
            ("encode failed", Kind::Internal),
            (
                "conflict in claim journal: stale epoch 2; journal epoch would be 3",
                Kind::Conflict,
            ),
            (
                "conflict in claim journal: stale epoch 2; current record epoch is 3",
                Kind::Conflict,
            ),
            (
                "gh transport: claim comment POST returned HTTP 422",
                Kind::Transport,
            ),
            (
                "gh transport: claim comment response carried no id",
                Kind::Transport,
            ),
            (
                "forge resource not found or not visible to this account (404)",
                Kind::NotFound,
            ),
            ("gh transport: cannot start gh: NotFound", Kind::Transport),
            ("forge blocked: authentication failed (401)", Kind::Blocked),
        ];
        for (message, kind) in cases {
            assert_eq!(contract::classify(message), kind, "{message}");
        }
    }
}
