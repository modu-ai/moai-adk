//! `claim-link` (Issue #3030): re-point an active claim record at the worktree and branch the
//! lane really works in.
//!
//! The hook guard (`claim-verify --hook`) finds a claim record by the folder a commit runs in.
//! A claim records the folder and branch of the moment it was posted; a worktree made after
//! it, or a branch renamed after it, leaves the record pointing elsewhere and the guard sees
//! "not a lane worktree" (cases: mo.ai.kr #911, #3048). `claim-link` fixes that without
//! touching the claim itself: it changes only `record.worktree` and appends one
//! `worktree_linked` effect and event. The branch and the path are measured from git, never
//! typed.
//!
//! Order, each step before the next (a mutant that reorders them must fail a test):
//! 1. the journal lock, before any forge call (busy is the typed `unavailable` answer);
//! 2. the one active record of the Issue; `--comment` and `--lane` must match it;
//! 3. ownership against the forge: the claim comment is the lowest open claim, its lease is
//!    live, and its `session=` is the lane (`claim-verify` rules, [`verify_owner`]);
//! 4. the local facts of the target folder ([`probe_target`]);
//! 5. no other active record holds the same path or branch ([`held_by_other`]);
//! 6. one journal event through `worktree::commit_update` (nothing is written when the
//!    record already says the same).
//!
//! Every refusal is an `Err` whose message starts `conflict in claim link: <code>: `, so the
//! contract classifies it as `conflict`; a message that has a command to run ends with it in
//! the `run `…`` form `error_next` recognises. A claim-verify refusal (`not owner`, `closed`,
//! `stale`, `uncertain`) keeps its own text and kind.
//!
//! Limits: every lane posts through one shared GitHub account, so `--lane` cannot be
//! authenticated; the check against the claim comment's `session=` only narrows it.

use crate::claim::{ClaimRecord, Journal, JournalEvent, WorktreeId, ACTIVE_STATES};
use crate::claim_verify::{
    classify_comment, evaluate, read_comments, refusal, CommentKind, Verifier,
};
use crate::forge::policy::Clock;
use crate::forge::transport::Transport;
use crate::forge::Client;
use serde_json::{json, Value};
use std::collections::BTreeMap;
use std::path::Path;

type Result<T> = std::result::Result<T, String>;

/// Longest `--lane` value, the same bound `claim-acquire` puts on an owner id.
pub const MAX_LANE_BYTES: usize = 128;

/// A refusal: `conflict in claim link: <code>: <detail>`.
pub fn refuse(code: &str, detail: impl std::fmt::Display) -> String {
    format!("conflict in claim link: {code}: {detail}")
}

/// The latest record of every task in the journal events, active states only.
pub fn latest_active(events: Vec<JournalEvent>) -> Vec<ClaimRecord> {
    let mut latest: BTreeMap<String, ClaimRecord> = BTreeMap::new();
    for e in events {
        latest.insert(e.record.task_key.clone(), e.record);
    }
    latest
        .into_values()
        .filter(|r| ACTIVE_STATES.contains(&r.state.as_str()))
        .collect()
}

/// The one active record of Issue `issue` in `owner/repo`.
pub fn find_record(
    records: &[ClaimRecord],
    issue: i64,
    owner: &str,
    repo: &str,
) -> Result<ClaimRecord> {
    let mut found = records.iter().filter(|r| {
        r.task_ref.issue_number == issue && r.task_ref.owner == owner && r.task_ref.repo == repo
    });
    let first = found.next().ok_or_else(|| {
        refuse(
            "record_not_found",
            format!(
                "no active claim record for Issue #{issue} of {owner}/{repo} in this clone's \
                 journal; claim it first with claim-acquire"
            ),
        )
    })?;
    if found.next().is_some() {
        return Err(refuse(
            "record_ambiguous",
            format!("more than one active claim record for Issue #{issue} of {owner}/{repo}"),
        ));
    }
    Ok(first.clone())
}

/// The record must be the one `--comment` and `--lane` name, and must be linkable. A claim
/// waiting for a handoff (`handoff_pending`) is not changed by its old owner.
pub fn check_record(record: &ClaimRecord, comment: i64, lane: &str) -> Result<()> {
    if record.comment_id != Some(comment) {
        return Err(refuse(
            "comment_differs",
            format!(
                "--comment {comment} is not the claim comment of the record ({})",
                record
                    .comment_id
                    .map_or("none".to_string(), |c| c.to_string())
            ),
        ));
    }
    if record.owner.id != lane {
        return Err(refuse(
            "lane_differs",
            format!(
                "--lane `{lane}` is not the owner of the record (`{}`)",
                record.owner.id
            ),
        ));
    }
    if record.state == "handoff_pending" {
        return Err(refuse(
            "state_not_linkable",
            "the record waits for a handoff (handoff_pending); the leader decides it",
        ));
    }
    Ok(())
}

/// Ownership of `record`'s claim comment, read from the forge once (GET only): the
/// `claim-verify` verdict must be `owner`, and the comment's `session=` must be `lane` — the
/// evaluator ignores the session when a comment id is given. Returns the verdict.
pub fn verify_owner<T: Transport, C: Clock>(
    record: &ClaimRecord,
    comment: i64,
    lane: &str,
    transport: &T,
    clock: &C,
) -> Result<Value> {
    let client = Client::new(transport, clock, Default::default(), Default::default());
    let comments = read_comments(
        &client,
        &record.task_ref.owner,
        &record.task_ref.repo,
        record.task_ref.issue_number,
    )?;
    let verdict = evaluate(
        &comments,
        &Verifier {
            session: lane,
            comment_id: Some(comment),
            allowed_authors: &[],
            now: clock.now(),
        },
    );
    if let Some(message) = refusal(&verdict) {
        return Err(message);
    }
    let session = comments
        .iter()
        .find(|c| c.id == comment)
        .map(|c| classify_comment(&c.body));
    match session {
        Some(CommentKind::Claim { session, .. }) if session == lane => Ok(verdict),
        Some(CommentKind::Claim { session, .. }) => Err(refuse(
            "session_differs",
            format!("the claim comment {comment} was posted by session `{session}`, not `{lane}`"),
        )),
        _ => Err(refuse(
            "session_differs",
            format!("comment {comment} carries no readable claim line"),
        )),
    }
}

/// `OWNER/REPO` of `--repo`, or of the project's `origin` remote on GitHub (as
/// `claim-verify --issue` reads it).
pub fn resolve_repo(project: &Path, repo: Option<&str>) -> Result<(String, String)> {
    let req =
        crate::claim_verify::request_from_flags(project, 1, 1, "lane", repo, &[]).map_err(|m| {
            match m.strip_prefix("claim-verify:") {
                Some(rest) => format!("claim-link:{rest}"),
                None => m,
            }
        })?;
    Ok((req.owner, req.repo))
}

/// Opens nothing: the journal events of an already locked journal.
pub fn active_records(journal: &Journal) -> Result<Vec<ClaimRecord>> {
    Ok(latest_active(journal.read_events()?))
}

/// What the target folder measures as, and how its HEAD relates to the claim base.
#[derive(Clone, Debug)]
pub struct Probe {
    /// Path, git dir, common dir, HEAD and branch, measured by git (canonical paths, as
    /// `claim-acquire` records them). `branch` is `None` on a detached HEAD.
    pub id: WorktreeId,
    /// `equal` (HEAD is the base) or `descends` (HEAD is a later commit of the base).
    pub relation: &'static str,
}

/// Exit code and stdout of `git -C dir args`; `Err` only when git cannot be started.
fn git_code(dir: &Path, args: &[&str]) -> Result<(i32, String)> {
    let mut cmd = std::process::Command::new("git");
    crate::identity::scrub_git_env(&mut cmd);
    let out = cmd
        .arg("-C")
        .arg(dir)
        .args(args)
        .stdin(std::process::Stdio::null())
        .stderr(std::process::Stdio::null())
        .output()
        .map_err(|_| "git unavailable".to_string())?;
    Ok((
        out.status.code().unwrap_or(-1),
        String::from_utf8_lossy(&out.stdout).trim().to_string(),
    ))
}

/// Measures `target` and decides whether it may be linked to a record whose claim base is
/// `base`. Reads only (git rev-parse, worktree list, cat-file, merge-base). Refuses, in this
/// order: a path that is not UTF-8; a symlink, a missing path or a file (the identity
/// reader's own refusals); a folder that is not inside a git work tree, or is a bare
/// repository or a bare git directory (`not_a_worktree`); a work tree of another clone
/// (`common_dir_differs`); a work tree `git worktree list` of this clone does not register
/// (`not_a_worktree`, which also catches a forged `.git` file and a pruned worktree); a base
/// commit this clone does not have (`base_missing`); and a HEAD that does not contain the
/// base (`base_not_in_head`, naming the one git command that fixes it). A detached HEAD is
/// linkable, with no branch.
pub fn probe_target(project: &Path, target: &Path, base: &str) -> Result<Probe> {
    let Some(shown) = target.to_str() else {
        return Err("invalid claim link request: the worktree path is not UTF-8".into());
    };
    if !crate::worktree::is_sha(base) {
        return Err("invalid record: base_revision must be 40 lowercase hex".into());
    }
    let ident = crate::identity::identity(target)?;
    match ident["identity_status"].as_str() {
        Some("ok") => {}
        Some("git_unavailable") => return Err("git unavailable".into()),
        _ => {
            return Err(refuse(
                "not_a_worktree",
                format!("{shown} is not inside a git work tree"),
            ))
        }
    }
    if !matches!(ident["workspace_kind"].as_str(), Some("primary" | "linked")) {
        return Err(refuse(
            "not_a_worktree",
            format!("{shown} is a bare repository or a git directory, not a work tree"),
        ));
    }
    let mine = crate::identity::identity(project)?;
    if mine["clone_anchor"].is_null() || mine["clone_anchor"] != ident["clone_anchor"] {
        return Err(refuse(
            "common_dir_differs",
            format!("{shown} belongs to another clone than the journal's"),
        ));
    }
    let top = ident["work_tree"].as_str().unwrap_or_default().to_string();
    let registered = crate::worktree::worktrees(project)?
        .iter()
        .any(|e| std::fs::canonicalize(&e.path).is_ok_and(|p| p == Path::new(&top)));
    if !registered {
        return Err(refuse(
            "not_a_worktree",
            format!("{top} is not a worktree git lists for this clone (git worktree list)"),
        ));
    }
    let id = crate::claim::worktree_id(&ident)?;
    let commit = format!("{base}^{{commit}}");
    if git_code(Path::new(&top), &["cat-file", "-e", &commit])?.0 != 0 {
        return Err(refuse(
            "base_missing",
            format!("the claim base {base} is not a commit of this clone"),
        ));
    }
    let quoted = crate::lane_env::quote(Path::new(&id.path));
    match git_code(
        Path::new(&top),
        &["merge-base", "--is-ancestor", base, "HEAD"],
    )?
    .0
    {
        0 => Ok(Probe {
            relation: if id.head == base { "equal" } else { "descends" },
            id,
        }),
        1 => {
            // HEAD an ancestor of the base: it only lags and fast-forwards; otherwise it has
            // commits of its own and the base has to be merged in.
            let behind = git_code(
                Path::new(&top),
                &["merge-base", "--is-ancestor", "HEAD", base],
            )?
            .0 == 0;
            let fix = if behind {
                format!("git -C {quoted} merge --ff-only {base}")
            } else {
                format!("git -C {quoted} merge {base}")
            };
            Err(refuse(
                "base_not_in_head",
                format!(
                    "HEAD {} of this worktree does not contain the claim base {base}; run `{fix}`",
                    id.head
                ),
            ))
        }
        _ => Err("git failed: merge-base could not compare HEAD with the claim base".into()),
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::claim::{Owner, SpecRef, TaskRef, WorktreeId};
    use crate::forge::policy::{parse_utc, Clock};
    use crate::forge::transport::{Exchange, HttpRequest, HttpResponse, RecordedTransport};

    const NOW: &str = "2026-11-09T05:00:00Z";
    const LIVE: &str = "2026-11-09T09:00:00Z";
    const EXPIRED: &str = "2026-11-09T04:00:00Z";
    const URL: &str = "https://api.github.com/repos/acme/widgets/issues/7/comments?per_page=100";

    struct FixedClock(i64);
    impl Clock for FixedClock {
        fn now(&self) -> i64 {
            self.0
        }
        fn sleep(&self, _: u64) {}
    }

    fn clock() -> FixedClock {
        FixedClock(parse_utc(NOW).unwrap())
    }

    fn record(key: &str, issue: i64, comment: i64, lane: &str, state: &str) -> ClaimRecord {
        ClaimRecord {
            schema_version: 1,
            task_ref: TaskRef::new(
                "github",
                "github.com",
                "acme",
                "widgets",
                1,
                70 + issue,
                issue,
            ),
            task_key: key.into(),
            logical_key: format!("L{issue}"),
            owner: Owner::new("session", lane),
            assignee: None,
            provider_session: None,
            integration_owner: None,
            attempt: None,
            epoch: 1,
            base_revision: "a".repeat(40),
            worktree: WorktreeId::new("/c/.git", "/c/.git", "/c", "b".repeat(40)),
            write_scope: vec!["src/x.rs".into()],
            spec_ref: SpecRef {
                id: "SPEC".into(),
                revision: "r1".into(),
                digest: None,
            },
            comment_id: Some(comment),
            claim_token: Some(format!("Moai-Claim: {issue}#{comment}")),
            lease_until: Some(LIVE.into()),
            state: state.into(),
            effects: vec![],
            last_observed_at: NOW.into(),
        }
    }

    fn event(r: &ClaimRecord) -> JournalEvent {
        JournalEvent {
            event: "acquire".into(),
            record: r.clone(),
        }
    }

    fn comment(id: i64, body: &str) -> Value {
        json!({"id": id, "user": {"login": "GoosLab"}, "body": body})
    }

    fn claim_body(session: &str, lease: &str) -> String {
        format!("moai-claim: session={session} epoch=1 lease_until={lease}")
    }

    fn get(body: Value) -> RecordedTransport {
        RecordedTransport::new(vec![Exchange {
            request: HttpRequest::new("GET", URL),
            response: HttpResponse::new(200).with_body(body),
        }])
    }

    #[test]
    fn find_record_takes_the_latest_active_record_of_the_issue() {
        let mut old = record("0000000000000001", 7, 100, "a", "assigned");
        let mut newer = old.clone();
        newer.epoch = 2;
        old.state = "assigned".into();
        let other = record("0000000000000002", 8, 200, "b", "running");
        let released = record("0000000000000003", 9, 300, "c", "cancelled");
        let records = latest_active(vec![
            event(&old),
            event(&other),
            event(&newer),
            event(&released),
        ]);
        assert_eq!(records.len(), 2, "the cancelled record is not active");
        let found = find_record(&records, 7, "acme", "widgets").unwrap();
        assert_eq!(found.epoch, 2, "the later event of the task wins");
        let err = find_record(&records, 9, "acme", "widgets").unwrap_err();
        assert!(
            err.starts_with("conflict in claim link: record_not_found: "),
            "{err}"
        );
        let err = find_record(&records, 7, "acme", "other").unwrap_err();
        assert!(err.contains("record_not_found"), "{err}");
        // Two active records for one Issue number of one repository cannot be told apart.
        let twin = record("0000000000000004", 7, 150, "d", "assigned");
        let records = latest_active(vec![event(&old), event(&twin)]);
        let err = find_record(&records, 7, "acme", "widgets").unwrap_err();
        assert!(err.contains("record_ambiguous"), "{err}");
    }

    #[test]
    fn check_record_names_the_differing_input() {
        let r = record("0000000000000001", 7, 100, "lane-a", "assigned");
        assert!(check_record(&r, 100, "lane-a").is_ok());
        let err = check_record(&r, 101, "lane-a").unwrap_err();
        assert!(
            err.starts_with("conflict in claim link: comment_differs: "),
            "{err}"
        );
        let err = check_record(&r, 100, "lane-b").unwrap_err();
        assert!(
            err.starts_with("conflict in claim link: lane_differs: "),
            "{err}"
        );
        for state in [
            "assigned",
            "running",
            "verifying",
            "pr_ready",
            "in_review",
            "blocked",
            "uncertain",
        ] {
            let mut s = r.clone();
            s.state = state.into();
            assert!(check_record(&s, 100, "lane-a").is_ok(), "{state}");
        }
        let mut pending = r.clone();
        pending.state = "handoff_pending".into();
        let err = check_record(&pending, 100, "lane-a").unwrap_err();
        assert!(err.contains("state_not_linkable"), "{err}");
        // A record without a comment id can never match --comment.
        let mut none = r;
        none.comment_id = None;
        assert!(check_record(&none, 100, "lane-a")
            .unwrap_err()
            .contains("comment_differs"));
    }

    #[test]
    fn only_the_live_lowest_claim_of_the_lane_passes() {
        let r = record("0000000000000001", 7, 100, "lane-a", "assigned");
        let owner = json!([comment(100, &claim_body("lane-a", LIVE))]);
        let t = get(owner);
        let v = verify_owner(&r, 100, "lane-a", &t, &clock()).unwrap();
        assert_eq!(v["status"], json!("owner"));
        assert_eq!(t.sent.borrow().len(), 1);
        assert!(t.sent.borrow().iter().all(|q| q.method == "GET"));

        // A lower open claim of another session holds the Issue.
        let lower = json!([
            comment(50, &claim_body("other", LIVE)),
            comment(100, &claim_body("lane-a", LIVE)),
        ]);
        let err = verify_owner(&r, 100, "lane-a", &get(lower), &clock()).unwrap_err();
        assert!(err.starts_with("claim verify: not owner"), "{err}");

        // A claim closed by a release line.
        let closed = json!([
            comment(100, &claim_body("lane-a", LIVE)),
            comment(110, "moai-claim-release: comment=100 by=leader"),
        ]);
        let err = verify_owner(&r, 100, "lane-a", &get(closed), &clock()).unwrap_err();
        assert!(err.starts_with("claim verify: closed"), "{err}");

        // An expired lease stays the holder: stale, not owner.
        let stale = json!([comment(100, &claim_body("lane-a", EXPIRED))]);
        let err = verify_owner(&r, 100, "lane-a", &get(stale), &clock()).unwrap_err();
        assert!(err.starts_with("claim verify: stale"), "{err}");

        // A comment that is not on the Issue is uncertain.
        let missing = json!([comment(999, &claim_body("lane-a", LIVE))]);
        let err = verify_owner(&r, 100, "lane-a", &get(missing), &clock()).unwrap_err();
        assert!(err.starts_with("claim verify: uncertain"), "{err}");
    }

    #[test]
    fn the_claim_comment_session_must_be_the_lane() {
        let r = record("0000000000000001", 7, 100, "lane-a", "assigned");
        // The comment id is right and the claim is the owner's, but another session wrote it.
        let forged = json!([comment(100, &claim_body("lane-b", LIVE))]);
        let err = verify_owner(&r, 100, "lane-a", &get(forged), &clock()).unwrap_err();
        assert!(
            err.starts_with("conflict in claim link: session_differs: "),
            "{err}"
        );
        assert_eq!(
            crate::contract::classify(&err),
            crate::contract::Kind::Conflict
        );
    }

    #[test]
    fn a_forge_failure_is_a_transport_error() {
        let r = record("0000000000000001", 7, 100, "lane-a", "assigned");
        let t = RecordedTransport::new(vec![]);
        let err = verify_owner(&r, 100, "lane-a", &t, &clock()).unwrap_err();
        assert_eq!(
            crate::contract::classify(&err),
            crate::contract::Kind::Transport,
            "{err}"
        );
    }
}

#[cfg(test)]
mod probe_tests {
    use super::*;
    use std::fs;
    use std::path::PathBuf;

    /// `git` with the fixture guards; returns trimmed stdout, panics on failure.
    fn git(dir: &Path, args: &[&str]) -> String {
        let out = std::process::Command::new("git")
            .args(["-c", "gc.auto=0", "-c", "maintenance.auto=false"])
            .args(["-c", "user.name=t", "-c", "user.email=t@example.invalid"])
            .args([
                "-c",
                "commit.gpgsign=false",
                "-c",
                "init.defaultBranch=main",
            ])
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .env("GIT_CONFIG_SYSTEM", "/dev/null")
            .env("LC_ALL", "C")
            .arg("-C")
            .arg(dir)
            .args(args)
            .output()
            .unwrap();
        assert!(
            out.status.success(),
            "git {args:?}: {}",
            String::from_utf8_lossy(&out.stderr)
        );
        String::from_utf8_lossy(&out.stdout).trim().to_string()
    }

    /// A clone with one commit on `main`; the returned folder is canonical.
    fn repo() -> (tempfile::TempDir, PathBuf) {
        let dir = tempfile::TempDir::new().unwrap();
        let root = fs::canonicalize(dir.path()).unwrap();
        git(&root, &["init", "-q", "-b", "main"]);
        git(&root, &["commit", "-q", "--allow-empty", "-m", "base"]);
        (dir, root)
    }

    fn add_worktree(root: &Path, name: &str, branch: &str, from: &str) -> PathBuf {
        let path = root.parent().unwrap().join(name);
        git(
            root,
            &[
                "worktree",
                "add",
                "-q",
                "-b",
                branch,
                path.to_str().unwrap(),
                from,
            ],
        );
        fs::canonicalize(path).unwrap()
    }

    fn code(err: &str) -> &str {
        err.strip_prefix("conflict in claim link: ")
            .and_then(|r| r.split(':').next())
            .unwrap_or("")
    }

    #[test]
    fn head_equal_to_the_base_and_a_later_commit_both_link() {
        let (_d, root) = repo();
        let base = git(&root, &["rev-parse", "HEAD"]);
        let w = add_worktree(&root, "w-equal", "feat/x", "HEAD");
        let p = probe_target(&root, &w, &base).unwrap();
        assert_eq!(p.relation, "equal");
        assert_eq!(p.id.path, w.to_str().unwrap());
        assert_eq!(p.id.branch.as_deref(), Some("feat/x"));
        assert_eq!(p.id.head, base);
        assert_eq!(p.id.common_dir, root.join(".git").to_str().unwrap());
        assert!(p.id.git_dir.contains("worktrees"), "{}", p.id.git_dir);
        // base + 1 commit: HEAD descends from the base.
        git(&w, &["commit", "-q", "--allow-empty", "-m", "one more"]);
        let p = probe_target(&root, &w, &base).unwrap();
        assert_eq!(p.relation, "descends");
        assert_ne!(p.id.head, base);
        // The primary checkout is a work tree too, and links with its own branch.
        let p = probe_target(&root, &root, &base).unwrap();
        assert_eq!(p.id.branch.as_deref(), Some("main"));
    }

    #[test]
    fn a_detached_head_links_with_no_branch() {
        let (_d, root) = repo();
        let base = git(&root, &["rev-parse", "HEAD"]);
        let w = add_worktree(&root, "w-detached", "feat/d", "HEAD");
        git(&w, &["switch", "-q", "--detach"]);
        let p = probe_target(&root, &w, &base).unwrap();
        assert_eq!(p.id.branch, None);
    }

    #[test]
    fn a_path_with_spaces_and_a_subfolder_resolve_to_the_work_tree_top() {
        let (_d, root) = repo();
        let base = git(&root, &["rev-parse", "HEAD"]);
        let w = add_worktree(&root, "my work tree", "feat/s", "HEAD");
        let sub = w.join("sub dir");
        fs::create_dir(&sub).unwrap();
        let p = probe_target(&root, &sub, &base).unwrap();
        assert_eq!(p.id.path, w.to_str().unwrap(), "the record holds the top");
    }

    #[test]
    fn a_head_that_lacks_the_base_is_refused_with_the_git_command_that_fixes_it() {
        let (_d, root) = repo();
        let old = git(&root, &["rev-parse", "HEAD"]);
        git(&root, &["commit", "-q", "--allow-empty", "-m", "newer"]);
        let base = git(&root, &["rev-parse", "HEAD"]);
        // Behind: a worktree made from the older commit fast-forwards.
        let w = add_worktree(&root, "w-behind", "feat/b", &old);
        let err = probe_target(&root, &w, &base).unwrap_err();
        assert_eq!(code(&err), "base_not_in_head", "{err}");
        let fix = format!("run `git -C '{}' merge --ff-only {base}`", w.display());
        assert!(err.ends_with(&fix), "{err}");
        // Diverged: the worktree has a commit the base does not, so a merge is needed.
        git(&w, &["commit", "-q", "--allow-empty", "-m", "own"]);
        let err = probe_target(&root, &w, &base).unwrap_err();
        let fix = format!("run `git -C '{}' merge {base}`", w.display());
        assert!(err.ends_with(&fix), "{err}");
        // After the merge the link is possible.
        git(&w, &["merge", "-q", "--no-edit", &base]);
        assert_eq!(probe_target(&root, &w, &base).unwrap().relation, "descends");
        // A base this clone does not have.
        let err = probe_target(&root, &w, &"c".repeat(40)).unwrap_err();
        assert_eq!(code(&err), "base_missing", "{err}");
    }

    #[test]
    fn the_target_must_be_a_registered_worktree_of_this_clone() {
        let (_d, root) = repo();
        let base = git(&root, &["rev-parse", "HEAD"]);
        // A plain folder.
        let plain = tempfile::TempDir::new().unwrap();
        let err = probe_target(&root, &fs::canonicalize(plain.path()).unwrap(), &base).unwrap_err();
        assert_eq!(code(&err), "not_a_worktree", "{err}");
        // A work tree of another clone.
        let (_d2, other) = repo();
        let err = probe_target(&root, &other, &base).unwrap_err();
        assert_eq!(code(&err), "common_dir_differs", "{err}");
        // A forged `.git` file in a plain folder that points at this clone's git dir: the
        // folder is not in `git worktree list`.
        let forged = root.parent().unwrap().join("forged");
        fs::create_dir(&forged).unwrap();
        fs::write(
            forged.join(".git"),
            format!("gitdir: {}\n", root.join(".git").display()),
        )
        .unwrap();
        let err = probe_target(&root, &forged, &base).unwrap_err();
        assert!(
            matches!(code(&err), "not_a_worktree" | "common_dir_differs"),
            "{err}"
        );
        // A pruned worktree: the folder stays but its admin directory is gone.
        let w = add_worktree(&root, "w-pruned", "feat/p", "HEAD");
        let admin = fs::read_dir(root.join(".git").join("worktrees"))
            .unwrap()
            .map(|e| e.unwrap().path())
            .find(|p| {
                fs::read_to_string(p.join("gitdir"))
                    .is_ok_and(|g| g.trim().starts_with(w.to_str().unwrap()))
            })
            .unwrap();
        fs::remove_dir_all(admin).unwrap();
        let err = probe_target(&root, &w, &base).unwrap_err();
        assert_eq!(code(&err), "not_a_worktree", "{err}");
        // A deleted folder.
        let gone = root.parent().unwrap().join("never-made");
        let err = probe_target(&root, &gone, &base).unwrap_err();
        assert_eq!(
            crate::contract::classify(&err),
            crate::contract::Kind::NotFound,
            "{err}"
        );
        // A bare repository.
        let bare = tempfile::TempDir::new().unwrap();
        let bare_root = fs::canonicalize(bare.path()).unwrap();
        git(&bare_root, &["init", "-q", "--bare"]);
        let err = probe_target(&root, &bare_root, &base).unwrap_err();
        assert_eq!(code(&err), "not_a_worktree", "{err}");
    }

    #[cfg(unix)]
    #[test]
    fn a_symlinked_target_is_refused() {
        let (_d, root) = repo();
        let base = git(&root, &["rev-parse", "HEAD"]);
        let w = add_worktree(&root, "w-real", "feat/l", "HEAD");
        let link = root.parent().unwrap().join("w-link");
        std::os::unix::fs::symlink(&w, &link).unwrap();
        let err = probe_target(&root, &link, &base).unwrap_err();
        assert!(err.contains("path is a symlink; refused"), "{err}");
        assert_eq!(
            crate::contract::classify(&err),
            crate::contract::Kind::InvalidInput
        );
    }

    #[cfg(unix)]
    #[test]
    fn a_path_that_is_not_utf8_is_refused_before_git_runs() {
        use std::os::unix::ffi::OsStrExt;
        let (_d, root) = repo();
        let base = git(&root, &["rev-parse", "HEAD"]);
        let odd = Path::new(std::ffi::OsStr::from_bytes(b"/tmp/claim-link-\xff"));
        let err = probe_target(&root, odd, &base).unwrap_err();
        assert!(err.contains("not UTF-8"), "{err}");
        assert_eq!(
            crate::contract::classify(&err),
            crate::contract::Kind::InvalidInput
        );
    }

    #[test]
    fn a_base_that_is_not_a_full_sha_is_refused() {
        let (_d, root) = repo();
        let err = probe_target(&root, &root, "HEAD").unwrap_err();
        assert!(err.contains("base_revision"), "{err}");
    }
}
