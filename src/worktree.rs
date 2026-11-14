//! One worktree per claim (Issue #2007, 06-C; REQ-06C-01..04, 10, 12).
//!
//! `worktree-plan` computes the branch and path for a claim and writes nothing.
//! `worktree-apply` re-derives the plan, refuses on a digest mismatch, detects an existing
//! owner itself (branch already used by a worktree, branch ref present, path present, or a
//! worktree locked for the same task) and only then runs
//! `git worktree add --lock --reason "moai claim <task_key> e<epoch>" -b <branch> <path> <base>`.
//! It never forces the add, and the primary worktree's HEAD, branch and index are never
//! written: the only mutating git call in this file is that `worktree add`.
//!
//! Decided inputs: the default path is `<HOME>/.moai/worktrees/<project>-<issue>-<logical_key>`
//! (#2811 W2, `crate::home_workdir`; before it `<primary>/.moai-cli/worktrees/<issue>-<logical_key>`,
//! decisions 16:5x→17:0x, where existing worktrees stay); any other path is accepted as given
//! (REQ-06C-12). `git worktree add` creates the missing parents of the default path (D5). The branch prefix is an open HOLD, so it is a required
//! input with no default.
//!
//! Records: every 06-C command returns its record change as `record_update_request`. When
//! the record came from the local journal (`--task`), the change is also committed through
//! 06-B's public `Journal::commit` under the journal lock ([`commit_update`]): the worktree
//! identity, `blocked(reason=scope_conflict)`, an integration order, and `pr_ready` with the
//! `pr_create` effect intent. 06-A's acquire/release behaviour is unchanged; these are
//! additional events. A record can also come from a file (`--record`) for a claim that never
//! went through the local journal; such a file is the caller's input, not an authority, is
//! accepted only when it agrees with the Issue claim comment it names (`--claim-comment`),
//! and is never written back: the journal holds no record for it.
use crate::claim::{
    self, claim_token, is_key, parse_claim_line, ClaimRecord, Journal, WorktreeId, ACTIVE_STATES,
};
use crate::digest;
use crate::forge::policy::{format_utc, Clock};
use crate::spec::{Canonical, SortedKeys};
use serde_json::{json, Value};
use std::{
    collections::BTreeMap,
    fs,
    path::{Path, PathBuf},
    process::{Command, Stdio},
};

type Result<T> = std::result::Result<T, String>;

pub const RECORD_UPDATE_OWNER: &str =
    "06-C via 06-B Journal::commit when the record came from --task; not written for --record";

/// Free-space floor below which `worktree-apply` refuses a new worktree: 30 GiB (WT-5,
/// #2380 Q5 user decision 2026-11-10, #2389 comment 6774899518 row 3).
pub const DISK_FLOOR_BYTES: u64 = 30 * 1024 * 1024 * 1024;
/// Project total (the worktrees' files plus their `target/`) above which a created worktree
/// carries a warning: 40 GiB (same decision).
pub const PROJECT_WARN_BYTES: u64 = 40 * 1024 * 1024 * 1024;

/// The capacity limits `worktree-apply` checks; the CLI passes the decided values unless a
/// hidden test override is given.
#[derive(Clone, Copy, Debug)]
pub struct Capacity {
    pub floor_bytes: u64,
    pub warn_bytes: u64,
}

impl Default for Capacity {
    fn default() -> Self {
        Capacity {
            floor_bytes: DISK_FLOOR_BYTES,
            warn_bytes: PROJECT_WARN_BYTES,
        }
    }
}

impl Capacity {
    /// True when a limit differs from the decided value (the hidden test flags were used);
    /// the answer shows it as `capacity.override`.
    pub fn is_override(&self) -> bool {
        self.floor_bytes != DISK_FLOOR_BYTES || self.warn_bytes != PROJECT_WARN_BYTES
    }
}

/// True when `free` bytes are at or above the floor. An unmeasured free space (`None`: off
/// Unix, or `df -Pk` failed or was unreadable on Unix) does not block; the answer shows
/// `free_bytes: null`.
pub fn disk_floor_ok(free: Option<u64>, floor: u64) -> bool {
    free.is_none_or(|f| f >= floor)
}

/// True when the project total is above the warning line (exactly the line is no warning).
pub fn over_warning(total: u64, warn: u64) -> bool {
    total > warn
}

/// Time budget for the project-total measurement behind the warning.
const CAPACITY_BUDGET: std::time::Duration = std::time::Duration::from_secs(5);

/// The journal of `project`, opened for writing and locked; a held lock is the one
/// journal lock-busy answer — the explicit `unavailable` kind, retryable (06-A
/// busy-never-waits, REQ-06A-05; #2345 sets the kind at this raise site).
pub fn open_locked_journal(
    project: &Path,
) -> std::result::Result<Journal, crate::contract::OpError> {
    let journal = Journal::open(&claim::journal_root(project)?)?;
    if !journal.try_lock()? {
        return Err(crate::contract::OpError::new(
            crate::contract::Kind::Unavailable,
            crate::claim::JOURNAL_LOCK_BUSY,
        ));
    }
    Ok(journal)
}

/// Applies `change` to the current record of `key` and commits it as one journal event.
/// The record must still be at `epoch` and active; a change that leaves the record as it
/// was writes nothing (`applied: false`), so a repeated command adds no event.
pub fn commit_update<C: Clock>(
    journal: &Journal,
    key: &str,
    epoch: u64,
    event: &str,
    clock: &C,
    change: impl FnOnce(&mut ClaimRecord),
) -> Result<Value> {
    if !is_key(key) {
        return Err("task must be a 16 hex character claim key".into());
    }
    let events = journal.read_events()?;
    let current = events
        .iter()
        .rev()
        .find(|e| e.record.task_key == key)
        .map(|e| e.record.clone())
        .ok_or("claim record not found for task")?;
    if current.epoch != epoch {
        return Err(format!(
            "conflict in claim journal: stale_epoch {epoch}; current record epoch is {}",
            current.epoch
        ));
    }
    if !ACTIVE_STATES.contains(&current.state.as_str()) {
        return Err(format!(
            "conflict in claim journal: the record is not active ({})",
            current.state
        ));
    }
    let mut next = current.clone();
    change(&mut next);
    if next == current {
        return Ok(json!({"applied": false, "event": event, "state": current.state}));
    }
    next.last_observed_at = format_utc(clock.now());
    journal.commit(event, &next)?;
    Ok(json!({"applied": true, "event": event, "state": next.state}))
}

/// `worktree-apply` plus the record write (REQ-06C-02): a created worktree's measured
/// identity becomes `record.worktree` when the record came from the journal.
#[allow(clippy::too_many_arguments)]
pub fn apply_and_record<C: Clock>(
    project: &Path,
    rec: &ClaimRecord,
    from_journal: bool,
    prefix: &str,
    path: Option<&Path>,
    home: Option<&Path>,
    plan_digest: &str,
    capacity: Capacity,
    clock: &C,
) -> std::result::Result<Value, crate::contract::OpError> {
    let journal = if from_journal {
        Some(open_locked_journal(project)?)
    } else {
        None
    };
    let mut out = apply(project, rec, prefix, path, home, plan_digest, capacity)?;
    out["record_update"] = if out["decision"] != "created" {
        json!({"applied": false, "reason": "no worktree was created"})
    } else if let Some(journal) = &journal {
        let id: WorktreeId =
            serde_json::from_value(out["identity"].clone()).map_err(|_| "encode failed")?;
        commit_update(
            journal,
            &rec.task_key,
            rec.epoch,
            "worktree_recorded",
            clock,
            |r| r.worktree = id,
        )?
    } else {
        not_written()
    };
    Ok(out)
}

/// The `record_update` field for a record that came from a file: nothing is written.
pub fn not_written() -> Value {
    json!({"applied": false, "reason": "the record came from --record; the journal holds no record for it"})
}

fn git(dir: &Path, args: &[&str]) -> Result<String> {
    let mut cmd = Command::new("git");
    crate::identity::scrub_git_env(&mut cmd);
    let out = cmd
        .args(args)
        .current_dir(dir)
        .output()
        .map_err(|_| "git unavailable".to_string())?;
    if !out.status.success() {
        let err = String::from_utf8_lossy(&out.stderr);
        return Err(format!(
            "git failed: {}",
            err.lines().next().unwrap_or("no message")
        ));
    }
    String::from_utf8(out.stdout).map_err(|_| "non UTF-8 git output".to_string())
}

fn git_ok(dir: &Path, args: &[&str]) -> Result<bool> {
    let mut cmd = Command::new("git");
    crate::identity::scrub_git_env(&mut cmd);
    cmd.args(args)
        .current_dir(dir)
        .stdout(Stdio::null())
        .stderr(Stdio::null())
        .status()
        .map(|s| s.success())
        .map_err(|_| "git unavailable".to_string())
}

/// The 40-hex HEAD of `project` (the candidate SHA of a PR body).
pub fn head(project: &Path) -> Result<String> {
    Ok(git(project, &["rev-parse", "HEAD"])?.trim().to_string())
}

/// `(status, path)` of every path changed between the merge base of `base` and HEAD, and HEAD.
pub fn changes(project: &Path, base: &str) -> Result<Vec<(String, String)>> {
    let range = format!("{base}...HEAD");
    let out = git(project, &["diff", "--name-status", "--no-renames", &range])?;
    Ok(out
        .lines()
        .filter_map(|l| l.split_once('\t'))
        .map(|(s, p)| (s.to_string(), p.to_string()))
        .collect())
}

/// The Cargo build folder `path` would sit in, if any (#2934 S3, #2930 R-3: lane worktrees were
/// made inside `target/`, so every cargo clean or trim took them along and the build output
/// nested without end). The path is read with its longest existing ancestor resolved (symbolic
/// links followed); a build folder is an ancestor that holds `CACHEDIR.TAG` (cargo writes one
/// at the top of every target directory) or a folder named `target` beside a `Cargo.toml`,
/// which counts before cargo has made it.
pub fn cargo_build_folder(path: &Path) -> Option<PathBuf> {
    let resolved = resolve_existing(path);
    resolved
        .ancestors()
        .find(|a| {
            a.join("CACHEDIR.TAG").is_file()
                || (a.file_name().is_some_and(|n| n == "target")
                    && a.parent().is_some_and(|p| p.join("Cargo.toml").is_file()))
        })
        .map(Path::to_path_buf)
}

/// Refuses a worktree path inside a Cargo build folder ([`cargo_build_folder`]).
pub fn refuse_in_build_folder(path: &Path, what: &str) -> Result<()> {
    match cargo_build_folder(path) {
        Some(dir) => Err(format!(
            "{what} {} is inside the Cargo build folder {} (#2930 R-3); put worktrees outside \
             build output, under the default <HOME>/.moai/worktrees",
            path.display(),
            dir.display()
        )),
        None => Ok(()),
    }
}

/// `path` with its longest existing ancestor canonicalised (symlinks resolved), the rest kept.
pub(crate) fn resolve_existing(path: &Path) -> PathBuf {
    let mut rest = Vec::new();
    let mut cur = path.to_path_buf();
    loop {
        if let Ok(c) = fs::canonicalize(&cur) {
            let mut out = c;
            for part in rest.iter().rev() {
                out.push(part);
            }
            return out;
        }
        match (cur.file_name().map(|n| n.to_os_string()), cur.parent()) {
            (Some(name), Some(parent)) => {
                rest.push(name);
                cur = parent.to_path_buf();
            }
            _ => return path.to_path_buf(),
        }
    }
}

pub fn is_sha(s: &str) -> bool {
    s.len() == 40
        && s.bytes()
            .all(|b| b.is_ascii_digit() || (b'a'..=b'f').contains(&b))
}

// ---- record input ----

fn read_json(path: &Path, what: &str) -> Result<Value> {
    let bytes = fs::read(path).map_err(|_| format!("{what} file not found"))?;
    serde_json::from_slice(&bytes).map_err(|_| format!("invalid {what} file"))
}

/// The record a command works on: from the local journal (`task`), or from a file checked
/// against the Issue claim comment (`record` + `comment`).
pub fn load_record(
    project: &Path,
    task: Option<&str>,
    record: Option<&Path>,
    comment: Option<&Path>,
) -> Result<ClaimRecord> {
    match (task, record) {
        (Some(key), None) => {
            let shown = claim::show(&claim::journal_root(project)?, key)?;
            serde_json::from_value(shown["record"].clone())
                .map_err(|_| "claim record not found for task".to_string())
        }
        (None, Some(path)) => {
            let rec: ClaimRecord = serde_json::from_value(read_json(path, "record")?)
                .map_err(|_| "invalid record file".to_string())?;
            let comment = comment.ok_or("--record requires --claim-comment")?;
            check_against_comment(&rec, &read_json(comment, "claim comment")?)?;
            Ok(rec)
        }
        _ => Err("exactly one of --task or --record is required".into()),
    }
}

/// A record file must agree with the Issue claim comment: comment id, session, epoch, lease
/// and (when present) the claim token. Any difference refuses the record.
pub fn check_against_comment(rec: &ClaimRecord, comment: &Value) -> Result<()> {
    let id = comment["id"]
        .as_i64()
        .ok_or("invalid claim comment file: id must be an integer")?;
    let line = comment["body"]
        .as_str()
        .and_then(parse_claim_line)
        .ok_or("invalid claim comment file: no moai-claim line in body")?;
    let mut diffs = Vec::new();
    if rec.comment_id != Some(id) {
        diffs.push("comment_id");
    }
    if rec.owner.id != line.session {
        diffs.push("session");
    }
    if rec.epoch != line.epoch {
        diffs.push("epoch");
    }
    if rec.lease_until.as_deref() != Some(line.lease_until.as_str()) {
        diffs.push("lease_until");
    }
    if let Some(token) = &rec.claim_token {
        if *token != claim_token(rec.task_ref.issue_number, id) {
            diffs.push("claim_token");
        }
    }
    if diffs.is_empty() {
        Ok(())
    } else {
        Err(format!(
            "conflict in record: {} differs from the Issue claim comment; record refused",
            diffs.join(", ")
        ))
    }
}

/// The latest record of every task in the local journal, active ones only.
pub fn active_records(project: &Path) -> Result<Vec<ClaimRecord>> {
    let events = Journal::open_read(&claim::journal_root(project)?)?.read_events()?;
    let mut latest: BTreeMap<String, ClaimRecord> = BTreeMap::new();
    for e in events {
        latest.insert(e.record.task_key.clone(), e.record);
    }
    Ok(latest
        .into_values()
        .filter(|r| ACTIVE_STATES.contains(&r.state.as_str()))
        .collect())
}

// ---- plan ----

pub(crate) struct Entry {
    pub(crate) path: String,
    pub(crate) head: String,
    pub(crate) branch: Option<String>,
    pub(crate) locked: Option<String>,
}

pub(crate) fn worktrees(project: &Path) -> Result<Vec<Entry>> {
    let text = git(project, &["worktree", "list", "--porcelain"])?;
    let mut out = Vec::new();
    for block in text.split("\n\n").filter(|b| !b.trim().is_empty()) {
        let mut e = Entry {
            path: String::new(),
            head: String::new(),
            branch: None,
            locked: None,
        };
        for line in block.lines() {
            if let Some(v) = line.strip_prefix("worktree ") {
                e.path = v.into();
            } else if let Some(v) = line.strip_prefix("HEAD ") {
                e.head = v.into();
            } else if let Some(v) = line.strip_prefix("branch ") {
                e.branch = Some(v.into());
            } else if line == "locked" {
                e.locked = Some(String::new());
            } else if let Some(v) = line.strip_prefix("locked ") {
                e.locked = Some(v.into());
            }
        }
        out.push(e);
    }
    Ok(out)
}

fn entry_json(e: &Entry) -> Value {
    json!({"path": e.path, "head": e.head, "branch": e.branch, "locked": e.locked})
}

fn valid_prefix(p: &str) -> bool {
    !p.is_empty()
        && !p.starts_with('/')
        && !p.ends_with('/')
        && !p.split('/').any(|s| s.is_empty() || s == "." || s == "..")
        && p.bytes()
            .all(|b| b.is_ascii_alphanumeric() || b"-_./".contains(&b))
}

// @MX:DEBT: [AUTO] the branch prefix is a required input with no default (HOLD, leader 2026-11-08)
// @MX:CEILING: REQ-06C-01 layout `<prefix>/<issue>-<logical_key>-e<epoch>` (Issue first since
//   #2336, like the lanes' `issue/<n>-<key>`, so `claim_verify::lane_branch` reads the same
//   Issue); it still differs from 06-A's `claim_branch_suffix` (`c<issue>-<comment_id>`)
// @MX:UPGRADE: when the user decides the prefix and its relation to the 06-A suffix
/// The plan for a claim's worktree (REQ-06C-01): no git change, byte-identical for equal input.
/// `home` is the user's home (`crate::home_workdir::home`); it is needed only when `path` is
/// `None` and names the default `<HOME>/.moai/worktrees/<project>-<issue>-<logical_key>`.
pub fn plan(
    project: &Path,
    rec: &ClaimRecord,
    prefix: &str,
    path: Option<&Path>,
    home: Option<&Path>,
) -> Result<Value> {
    if !valid_prefix(prefix) {
        return Err("invalid --branch-prefix: segments of [A-Za-z0-9._-] separated by /".into());
    }
    if !is_sha(&rec.base_revision) {
        return Err("invalid record: base_revision must be 40 lowercase hex".into());
    }
    let issue = rec.task_ref.issue_number;
    let branch = format!("{prefix}/{issue}-{}-e{}", rec.logical_key, rec.epoch);
    if !git_ok(project, &["check-ref-format", "--branch", &branch])? {
        return Err(format!("invalid branch name {branch}"));
    }
    let primary = worktrees(project)?
        .into_iter()
        .next()
        .map(|e| PathBuf::from(e.path))
        .ok_or("project root unavailable")?;
    let path = match path {
        Some(p) if p.is_absolute() => p.to_path_buf(),
        Some(p) => primary.join(p),
        None => crate::home_workdir::worktree_dir(
            home.ok_or("home directory missing; pass --path for the worktree")?,
            &primary,
            &format!("{issue}-{}", rec.logical_key),
        )?,
    };
    refuse_in_build_folder(&path, "worktree path")?;
    let content = json!({
        "task_key": rec.task_key,
        "logical_key": rec.logical_key,
        "issue_number": issue,
        "epoch": rec.epoch,
        "branch": branch,
        "path": path.to_string_lossy(),
        "base_sha": rec.base_revision,
        "primary": primary.to_string_lossy(),
        "lock_reason": format!("moai claim {} e{}", rec.task_key, rec.epoch),
    });
    let digest = digest(&SortedKeys.bytes(&content));
    let mut out = content;
    out["plan_digest"] = json!(digest);
    out["writes"] = json!(false);
    Ok(out)
}

// ---- apply ----

/// Measured identity of a worktree; never inferred from the `.git` shape or a basename (D42).
pub fn identity(path: &Path) -> Result<WorktreeId> {
    let out = git(
        path,
        &[
            "rev-parse",
            "--path-format=absolute",
            "--show-toplevel",
            "--git-common-dir",
            "--git-dir",
            "HEAD",
            "--abbrev-ref",
            "HEAD",
        ],
    )?;
    let lines: Vec<&str> = out.lines().collect();
    let [top, common, gitdir, head, abbrev] = lines[..] else {
        return Err("unexpected git rev-parse output".into());
    };
    let mut id = WorktreeId::new(common, gitdir, top, head);
    // #2131: the branch checked out at claim time, from the same rev-parse call.
    // A detached HEAD prints `HEAD` (git forbids a branch of that name), so it is None.
    id.branch = (abbrev != "HEAD").then(|| abbrev.to_string());
    Ok(id)
}

fn conflict(reason: &str, plan: &Value, existing: Vec<Value>) -> Value {
    json!({
        "decision": "conflict",
        "reason": reason,
        "branch": plan["branch"],
        "path": plan["path"],
        "existing": existing,
        "record_update_request": Value::Null,
    })
}

/// REQ-06C-02/03: create the worktree, or report the existing owner; the add is never forced.
pub fn apply(
    project: &Path,
    rec: &ClaimRecord,
    prefix: &str,
    path: Option<&Path>,
    home: Option<&Path>,
    plan_digest: &str,
    capacity: Capacity,
) -> Result<Value> {
    let p = plan(project, rec, prefix, path, home)?;
    if p["plan_digest"] != plan_digest {
        return Err("stale or altered plan; re-preview required".into());
    }
    let branch = p["branch"].as_str().unwrap_or_default().to_string();
    let target = PathBuf::from(p["path"].as_str().unwrap_or_default());
    let reason = p["lock_reason"].as_str().unwrap_or_default().to_string();
    let task_lock = format!("moai claim {} ", rec.task_key);
    let entries = worktrees(project)?;
    let full_ref = format!("refs/heads/{branch}");
    let in_use: Vec<Value> = entries
        .iter()
        .filter(|e| e.branch.as_deref() == Some(full_ref.as_str()))
        .map(entry_json)
        .collect();
    if !in_use.is_empty() {
        return Ok(conflict("branch_in_use", &p, in_use));
    }
    let same_task: Vec<Value> = entries
        .iter()
        .filter(|e| {
            e.locked
                .as_deref()
                .is_some_and(|l| l.starts_with(&task_lock))
        })
        .map(entry_json)
        .collect();
    if !same_task.is_empty() {
        return Ok(conflict("task_has_worktree", &p, same_task));
    }
    if git_ok(project, &["show-ref", "--verify", "--quiet", &full_ref])? {
        let head = git(project, &["rev-parse", &full_ref])?;
        return Ok(conflict(
            "branch_exists",
            &p,
            vec![json!({"branch": full_ref, "head": head.trim()})],
        ));
    }
    if target.exists() {
        let at: Vec<Value> = entries
            .iter()
            .filter(|e| Path::new(&e.path) == target)
            .map(entry_json)
            .collect();
        return Ok(conflict("path_exists", &p, at));
    }
    // WT-5: refuse below the free-space floor, measured right before the add.
    let free = crate::worktree_report::disk_bytes(project).map(|(free, _)| free);
    if !disk_floor_ok(free, capacity.floor_bytes) {
        return Ok(conflict(
            "disk_free_below_floor",
            &p,
            vec![json!({
                "free_bytes": free,
                "floor_bytes": capacity.floor_bytes,
                "warn_bytes": capacity.warn_bytes,
                "override": capacity.is_override(),
            })],
        ));
    }
    let target_s = target.to_string_lossy().to_string();
    let base = p["base_sha"].as_str().unwrap_or_default().to_string();
    git(
        project,
        &[
            "worktree", "add", "--lock", "--reason", &reason, "-b", &branch, &target_s, &base,
        ],
    )?;
    let id = identity(&target)?;
    // WT-5: warn when the project's worktrees hold more than the warning line. The size
    // walk is bounded; a partial total is a lower bound and says so.
    let measured = crate::worktree_report::project_bytes(project, CAPACITY_BUDGET);
    let capacity_json = json!({
        "free_bytes": free,
        "floor_bytes": capacity.floor_bytes,
        "project_bytes": measured.map(|m| m.0),
        "project_bytes_partial": measured.map(|m| m.1),
        "warn_bytes": capacity.warn_bytes,
        "warning": measured.is_some_and(|m| over_warning(m.0, capacity.warn_bytes)),
        "override": capacity.is_override(),
    });
    Ok(json!({
        "decision": "created",
        "capacity": capacity_json,
        "plan_digest": plan_digest,
        "branch": branch,
        "path": target_s,
        "locked": reason,
        "identity": id,
        "record_update_request": {
            "task_key": rec.task_key,
            "epoch": rec.epoch,
            "worktree": id,
        },
        "record_update_owner": RECORD_UPDATE_OWNER,
    }))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn the_floor_is_30_gib_and_exactly_the_floor_is_enough() {
        assert_eq!(DISK_FLOOR_BYTES, 32_212_254_720);
        assert_eq!(PROJECT_WARN_BYTES, 42_949_672_960);
        assert!(disk_floor_ok(Some(32_212_254_720), DISK_FLOOR_BYTES));
        assert!(!disk_floor_ok(Some(32_212_254_719), DISK_FLOOR_BYTES));
        // Unmeasured free space does not block.
        assert!(disk_floor_ok(None, DISK_FLOOR_BYTES));
    }

    #[test]
    fn the_warning_starts_one_byte_above_40_gib() {
        assert!(!over_warning(42_949_672_960, PROJECT_WARN_BYTES));
        assert!(over_warning(42_949_672_961, PROJECT_WARN_BYTES));
    }

    #[test]
    fn only_the_decided_values_are_not_an_override() {
        assert!(!Capacity::default().is_override());
        let floor = Capacity {
            floor_bytes: 0,
            ..Capacity::default()
        };
        assert!(floor.is_override());
        let warn = Capacity {
            warn_bytes: 0,
            ..Capacity::default()
        };
        assert!(warn.is_override());
    }
}
