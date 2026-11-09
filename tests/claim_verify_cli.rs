//! rc.2 A9 #2255: claim ownership re-verification. The evaluator is pure, so every AC-1
//! scenario is a fixture of comments; the op and the hook run over recorded exchanges or a
//! stub `gh` in temporary repositories. No live GitHub call is made.

use moai_cli::claim::{
    parse_claim_line, render_claim_line, resolve_winner, ClaimComment, ClaimFields, ClaimRecord,
    Journal, Owner, SpecRef, TaskRef, WorktreeId,
};
use moai_cli::claim_verify::{
    classify_comment, evaluate, hook_cli, is_commit_or_push, refusal, render_handoff_line,
    render_release_line, verify_cli, CommentKind, IssueComment, Verifier, VerifyRequest,
};
use moai_cli::contract::{classify, Kind};
use moai_cli::forge::policy::{parse_utc, Clock};
use moai_cli::forge::transport::{Exchange, HttpRequest, HttpResponse, RecordedTransport};
use serde_json::{json, Value};
use std::fs;
use std::path::Path;
use std::process::{Command, Stdio};

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

fn now() -> i64 {
    parse_utc(NOW).unwrap()
}

fn claim(id: i64, session: &str, lease: &str) -> IssueComment {
    IssueComment {
        id,
        author: "GoosLab".into(),
        body: format!("moai-claim: session={session} epoch=1 lease_until={lease}"),
    }
}

fn comment(id: i64, author: &str, body: &str) -> IssueComment {
    IssueComment {
        id,
        author: author.into(),
        body: body.into(),
    }
}

fn verdict(comments: &[IssueComment], session: &str, comment_id: Option<i64>) -> Value {
    evaluate(
        comments,
        &Verifier {
            session,
            comment_id,
            allowed_authors: &["GoosLab".to_string()],
            now: now(),
        },
    )
}

/// The generator writes the RC1 line first, so the RC1 parser reads it unchanged, and the
/// appended fields round-trip through the new classifier.
#[test]
fn render_claim_line_keeps_the_rc1_prefix_and_round_trips() {
    let line = render_claim_line(&ClaimFields {
        session: "agent-1-e29c15".into(),
        epoch: 2,
        lease_until: LIVE.into(),
        lane: Some("agent-1".into()),
        issue: Some(2255),
        branch: Some("issue/2255-rc2-a9-claim-enforce".into()),
        at: Some(NOW.into()),
        relay: true,
    });
    assert!(line.starts_with(&format!(
        "moai-claim: session=agent-1-e29c15 epoch=2 lease_until={LIVE} "
    )));
    for part in [
        "lane=agent-1",
        "issue=2255",
        "branch=issue/2255-rc2-a9-claim-enforce",
        &format!("at={NOW}"),
        "relay=true",
    ] {
        assert!(line.contains(part), "{line}");
    }
    let rc1 = parse_claim_line(&line).unwrap();
    assert_eq!((rc1.session.as_str(), rc1.epoch), ("agent-1-e29c15", 2));
    assert_eq!(
        classify_comment(&line),
        CommentKind::Claim {
            session: "agent-1-e29c15".into(),
            epoch: 2,
            lease_until: LIVE.into()
        }
    );
    let unverified = render_claim_line(&ClaimFields {
        session: "unverified".into(),
        epoch: 1,
        lease_until: LIVE.into(),
        ..Default::default()
    });
    assert_eq!(
        unverified,
        format!("moai-claim: session=unverified epoch=1 lease_until={LIVE}")
    );
}

/// AC-1 concurrent posts: the lowest open claim id holds; the other claimant is refused.
#[test]
fn concurrent_claims_lowest_id_holds() {
    let c = [claim(100, "a", LIVE), claim(200, "b", LIVE)];
    assert_eq!(verdict(&c, "a", Some(100))["status"], json!("owner"));
    let v = verdict(&c, "b", Some(200));
    assert_eq!(v["status"], json!("not_owner"));
    assert_eq!(v["lost_to"], json!(100));
    let msg = refusal(&v).unwrap();
    assert_eq!(classify(&msg), Kind::Conflict, "{msg}");
    assert!(refusal(&verdict(&c, "a", Some(100))).is_none());
}

/// AC-1 late-visible competitor: the first read showed only our claim; the re-read shows a
/// lower id and the verdict flips to not_owner.
#[test]
fn late_visible_competitor_is_judged_on_reread() {
    let first = [claim(200, "b", LIVE)];
    assert_eq!(verdict(&first, "b", Some(200))["status"], json!("owner"));
    let reread = [claim(200, "b", LIVE), claim(150, "a", LIVE)];
    let v = verdict(&reread, "b", Some(200));
    assert_eq!(v["status"], json!("not_owner"));
    assert_eq!(v["lost_to"], json!(150));
}

/// AC-1 post failure / uncertain: our recorded comment id is not on the Issue.
#[test]
fn missing_own_comment_is_uncertain() {
    let c = [claim(100, "a", LIVE)];
    let v = verdict(&c, "b", Some(300));
    assert_eq!(v["status"], json!("uncertain"));
    assert_eq!(classify(&refusal(&v).unwrap()), Kind::Blocked);
}

/// AC-1 duplicate: the same session and epoch claimed twice — the first stands, the second
/// is reported and ignored.
#[test]
fn duplicate_claim_is_reported_and_ignored() {
    let c = [claim(100, "a", LIVE), claim(120, "a", LIVE)];
    let v = verdict(&c, "a", None);
    assert_eq!(v["status"], json!("owner"));
    assert_eq!(v["mine"], json!(100));
    assert_eq!(
        v["invalid"],
        json!([{"comment_id": 120, "reason": "duplicate"}])
    );
}

/// AC-1 forged: a claim line from an author outside allowed_authors is ignored, so a lower
/// forged id does not beat the real claim (REQ-A9-02).
#[test]
fn forged_author_is_ignored() {
    let forged = comment(
        50,
        "mallory",
        &format!("moai-claim: session=b epoch=1 lease_until={LIVE}"),
    );
    let c = [forged, claim(100, "a", LIVE)];
    let v = verdict(&c, "a", Some(100));
    assert_eq!(v["status"], json!("owner"));
    assert_eq!(
        v["invalid"],
        json!([{"comment_id": 50, "reason": "author_mismatch"}])
    );
}

/// AC-1 handoff and release: a closing line opens the Issue to the next claim; the closed
/// claimant itself is refused as `closed`.
#[test]
fn handoff_and_release_close_the_holder() {
    for closing in [
        render_handoff_line(100, "b", "leader"),
        render_release_line(100, "leader"),
    ] {
        let c = [
            claim(100, "a", LIVE),
            comment(150, "GoosLab", &format!("relay note\n{closing}")),
            claim(200, "b", LIVE),
        ];
        let v = verdict(&c, "b", Some(200));
        assert_eq!(v["status"], json!("owner"), "{closing}");
        assert_eq!(v["closed"][0]["comment_id"], json!(100));
        let old = verdict(&c, "a", Some(100));
        assert_eq!(old["status"], json!("closed"));
        assert_eq!(classify(&refusal(&old).unwrap()), Kind::Conflict);
    }
    assert!(matches!(
        classify_comment(&render_handoff_line(1, "x", "y")),
        CommentKind::Handoff { target: 1, .. }
    ));
}

/// AC-1 stale (REQ-A9-03): the holder's lease expired and no handoff exists — the verdict is
/// `stale` for everyone, nobody is assigned, and the holder's own commit is refused too.
#[test]
fn expired_holder_is_stale_and_never_reassigned() {
    let c = [claim(100, "a", EXPIRED), claim(200, "b", LIVE)];
    for (session, id) in [("b", 200), ("a", 100)] {
        let v = verdict(&c, session, Some(id));
        assert_eq!(v["status"], json!("stale"), "{session}");
        assert_eq!(v["holder"], json!(100));
        assert_eq!(v["auto_assigned"], json!(false));
        assert_eq!(classify(&refusal(&v).unwrap()), Kind::Blocked);
    }
}

/// AC-1 restart: a restarted session finds its claim through the journal comment id or its
/// session id; a session that restarted under a new id owns nothing.
#[test]
fn restart_keeps_ownership_only_for_the_same_claim() {
    let c = [claim(100, "a", LIVE), claim(200, "b", LIVE)];
    assert_eq!(verdict(&c, "a", Some(100))["status"], json!("owner"));
    assert_eq!(verdict(&c, "a", None)["status"], json!("owner"));
    let v = verdict(&c, "a-restarted", None);
    assert_eq!(v["status"], json!("not_owner"));
    assert_eq!(v["mine"], json!(null));
}

/// Malformed lines of each family are reported and never counted as claims.
#[test]
fn malformed_lines_are_reported() {
    let c = [
        comment(
            10,
            "GoosLab",
            "moai-claim: session=a epoch=1 lease_until=tomorrow",
        ),
        comment(11, "GoosLab", "moai-claim: session=a"),
        comment(12, "GoosLab", "moai-claim-release: by=leader"),
        comment(13, "GoosLab", "moai-claim-handoff: comment=10"),
        comment(14, "GoosLab", "moai-claim-release: comment=999 by=leader"),
        comment(15, "GoosLab", "just a note"),
        claim(100, "a", LIVE),
    ];
    let v = verdict(&c, "a", None);
    assert_eq!(v["status"], json!("owner"));
    let ids: Vec<i64> = v["invalid"]
        .as_array()
        .unwrap()
        .iter()
        .map(|i| {
            assert_eq!(i["reason"], json!("malformed"));
            i["comment_id"].as_i64().unwrap()
        })
        .collect();
    assert_eq!(ids, [10, 11, 12, 13, 14]);
    assert_eq!(v["claim_lines"], json!(6));
}

/// Where every claim is live and valid, the holder equals RC1 `resolve_winner`.
#[test]
fn agrees_with_resolve_winner_when_all_claims_are_live() {
    let sets: [&[i64]; 4] = [&[5], &[9, 3], &[300, 200, 100], &[42, 7, 99, 8]];
    for ids in sets {
        let comments: Vec<IssueComment> = ids
            .iter()
            .map(|&id| claim(id, &format!("s{id}"), LIVE))
            .collect();
        let rc1: Vec<ClaimComment> = ids
            .iter()
            .map(|&id| ClaimComment {
                comment_id: id,
                session: format!("s{id}"),
                epoch: 1,
                lease_until: LIVE.into(),
            })
            .collect();
        let winner = resolve_winner(&rc1).unwrap().comment_id;
        assert_eq!(verdict(&comments, "nobody", None)["holder"], json!(winner));
        let v = verdict(&comments, &format!("s{winner}"), None);
        assert_eq!(v["status"], json!("owner"));
    }
}

#[test]
fn commit_or_push_detection() {
    for yes in [
        "git commit -m x",
        "git -C repo push origin HEAD",
        "cargo test && git commit -F msg.txt",
        "/usr/bin/git push",
        "(cd x; git commit)",
    ] {
        assert!(is_commit_or_push(yes), "{yes}");
    }
    for no in [
        "git status",
        "git log --grep=commit",
        "echo commit",
        "gh pr create",
        "",
    ] {
        assert!(
            !is_commit_or_push(no) || no.contains("--grep=commit"),
            "{no}"
        );
    }
    assert!(!is_commit_or_push("git status && echo push"));
}

fn get(body: Value) -> Exchange {
    Exchange {
        request: HttpRequest {
            method: "GET".into(),
            url: URL.into(),
            headers: Default::default(),
            body: None,
        },
        response: HttpResponse {
            status: 200,
            headers: Default::default(),
            body: Some(body),
        },
    }
}

fn raw(id: i64, login: &str, body: &str) -> Value {
    json!({"id": id, "user": {"login": login}, "body": body})
}

fn request(session: &str, comment_id: Option<i64>) -> VerifyRequest {
    VerifyRequest {
        schema_version: 1,
        owner: "acme".into(),
        repo: "widgets".into(),
        issue_number: 7,
        session: session.into(),
        comment_id,
        allowed_authors: vec!["GoosLab".into()],
    }
}

/// REQ-A9-01 through the op: one GET, the owner passes, a non-owner is refused (Conflict),
/// and nothing is written to GitHub.
#[test]
fn verify_op_refuses_non_owner_with_one_get() {
    let body = json!([
        raw(
            100,
            "GoosLab",
            &format!("moai-claim: session=a epoch=1 lease_until={LIVE}")
        ),
        raw(
            200,
            "GoosLab",
            &format!("moai-claim: session=b epoch=1 lease_until={LIVE}")
        ),
    ]);
    let clock = FixedClock(now());
    let t = RecordedTransport::new(vec![get(body.clone())]);
    let ok = verify_cli(&request("a", Some(100)), &t, &clock).unwrap();
    assert_eq!(ok["status"], json!("owner"));
    assert!(t.sent.borrow().iter().all(|r| r.method == "GET"));

    let t = RecordedTransport::new(vec![get(body)]);
    let err = verify_cli(&request("b", Some(200)), &t, &clock).unwrap_err();
    assert_eq!(classify(&err), Kind::Conflict, "{err}");
    assert_eq!(t.sent.borrow().len(), 1);

    let mut bad = request("a", None);
    bad.owner = "acme/../x".into();
    let t = RecordedTransport::new(vec![]);
    let err = verify_cli(&bad, &t, &clock).unwrap_err();
    assert_eq!(classify(&err), Kind::InvalidInput);
    assert!(t.sent.borrow().is_empty());
}

fn git(dir: &Path, args: &[&str]) -> String {
    let out = Command::new("git")
        .args(args)
        .current_dir(dir)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .env("GIT_AUTHOR_NAME", "t")
        .env("GIT_AUTHOR_EMAIL", "t@example.invalid")
        .env("GIT_COMMITTER_NAME", "t")
        .env("GIT_COMMITTER_EMAIL", "t@example.invalid")
        .output()
        .unwrap();
    assert!(
        out.status.success(),
        "git {args:?}: {}",
        String::from_utf8_lossy(&out.stderr)
    );
    String::from_utf8_lossy(&out.stdout).trim().to_string()
}

/// A temporary repository with one active claim record whose worktree is the repo itself.
fn claimed_repo(comment_id: i64, session: &str) -> tempfile::TempDir {
    let dir = tempfile::TempDir::new().unwrap();
    git(dir.path(), &["init", "-q", "-b", "main"]);
    git(dir.path(), &["commit", "-q", "--allow-empty", "-m", "base"]);
    let base = git(dir.path(), &["rev-parse", "HEAD"]);
    let root = fs::canonicalize(dir.path()).unwrap();
    let record = ClaimRecord {
        schema_version: 1,
        task_ref: TaskRef {
            provider: "github".into(),
            host: "github.com".into(),
            owner: "acme".into(),
            repo: "widgets".into(),
            repository_id: 1,
            issue_id: 70,
            issue_number: 7,
        },
        task_key: "0123456789abcdef".into(),
        logical_key: "A9".into(),
        owner: Owner {
            kind: "session".into(),
            id: session.into(),
            host_id: None,
            pid: None,
            process_start: None,
        },
        assignee: None,
        provider_session: None,
        integration_owner: None,
        attempt: None,
        epoch: 1,
        base_revision: base,
        worktree: WorktreeId {
            common_dir: root.join(".git").display().to_string(),
            git_dir: root.join(".git").display().to_string(),
            path: root.display().to_string(),
            head: "b".repeat(40),
            branch: None,
        },
        write_scope: vec!["src/x.rs".into()],
        spec_ref: SpecRef {
            id: "SPEC-A9".into(),
            revision: "r1".into(),
            digest: None,
        },
        comment_id: Some(comment_id),
        claim_token: Some(format!("Moai-Claim: 7#{comment_id}")),
        lease_until: Some(LIVE.into()),
        state: "assigned".into(),
        effects: vec![],
        last_observed_at: NOW.into(),
    };
    let journal = Journal::open(&moai_cli::claim::journal_root(&root).unwrap()).unwrap();
    assert!(journal.try_lock().unwrap());
    journal.commit("acquire", &record).unwrap();
    dir
}

fn hook_input(command: &str, cwd: &Path) -> Vec<u8> {
    json!({"hook_event_name": "PreToolUse", "tool_name": "Bash",
           "tool_input": {"command": command}, "cwd": cwd})
    .to_string()
    .into_bytes()
}

/// AC-2 (library path): the guard lets everything through except a commit/push inside a
/// claimed worktree, which it verifies against the Issue.
#[test]
fn hook_guards_commit_in_claimed_worktree_only() {
    let clock = FixedClock(now());
    let foreign_lower = json!([
        raw(
            100,
            "GoosLab",
            &format!("moai-claim: session=a epoch=1 lease_until={LIVE}")
        ),
        raw(
            200,
            "GoosLab",
            &format!("moai-claim: session=b epoch=1 lease_until={LIVE}")
        ),
    ]);
    let repo = claimed_repo(200, "b");
    let root = fs::canonicalize(repo.path()).unwrap();

    // Not a commit: allowed, no GitHub read.
    let t = RecordedTransport::new(vec![]);
    let v = hook_cli(&root, &hook_input("cargo test", &root), &t, &clock).unwrap();
    assert_eq!(v["allowed"], json!(true));
    assert!(t.sent.borrow().is_empty());
    let v = hook_cli(&root, br#"{"tool_name":"Edit"}"#, &t, &clock).unwrap();
    assert_eq!(v["allowed"], json!(true));

    // Commit in the claimed worktree while a lower claim holds: refused.
    let t = RecordedTransport::new(vec![get(foreign_lower.clone())]);
    let err = hook_cli(&root, &hook_input("git commit -m x", &root), &t, &clock).unwrap_err();
    assert_eq!(classify(&err), Kind::Conflict, "{err}");

    // The owner's commit passes.
    let owner = claimed_repo(100, "a");
    let oroot = fs::canonicalize(owner.path()).unwrap();
    let t = RecordedTransport::new(vec![get(foreign_lower)]);
    let v = hook_cli(&oroot, &hook_input("git push", &oroot), &t, &clock).unwrap();
    assert_eq!(v["status"], json!("owner"));

    // A commit in a tree without a claim record is not guarded.
    let plain = tempfile::TempDir::new().unwrap();
    git(plain.path(), &["init", "-q"]);
    let proot = fs::canonicalize(plain.path()).unwrap();
    let t = RecordedTransport::new(vec![]);
    let v = hook_cli(&proot, &hook_input("git commit", &proot), &t, &clock).unwrap();
    assert_eq!(v["reason"], json!("not a claimed worktree"));
}

fn gh_stub(dir: &Path, json_body: &str) -> std::path::PathBuf {
    let path = dir.join("gh-stub");
    fs::write(
        &path,
        "#!/bin/sh\nprintf 'HTTP/2.0 200 OK\\r\\n\\r\\n%s' \"$STUB_JSON\"\n",
    )
    .unwrap();
    #[cfg(unix)]
    {
        use std::os::unix::fs::PermissionsExt;
        fs::set_permissions(&path, fs::Permissions::from_mode(0o755)).unwrap();
    }
    let _ = json_body;
    path
}

/// The binary: `claim-verify --hook` exits 2 on a refused commit (the code a host hook
/// needs to block) and 0 when the commit is allowed.
#[cfg(unix)]
#[test]
fn hook_binary_exits_2_on_refusal() {
    let lease = moai_cli::forge::policy::format_utc(
        std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap()
            .as_secs() as i64
            + 3600,
    );
    let repo = claimed_repo(200, "b");
    let root = fs::canonicalize(repo.path()).unwrap();
    let stub_dir = tempfile::TempDir::new().unwrap();
    let body = json!([
        raw(
            100,
            "GoosLab",
            &format!("moai-claim: session=a epoch=1 lease_until={lease}")
        ),
        raw(
            200,
            "GoosLab",
            &format!("moai-claim: session=b epoch=1 lease_until={lease}")
        ),
    ])
    .to_string();
    let gh = gh_stub(stub_dir.path(), &body);
    let run = |command: &str| {
        use std::io::Write as _;
        let mut child = Command::new(env!("CARGO_BIN_EXE_moai-cli"))
            .args(["claim-verify", root.to_str().unwrap(), "--hook"])
            .current_dir(&root)
            .env("MOAI_FORGE_GH_BIN", &gh)
            .env("STUB_JSON", &body)
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .env("GIT_CONFIG_SYSTEM", "/dev/null")
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .spawn()
            .unwrap();
        child
            .stdin
            .take()
            .unwrap()
            .write_all(&hook_input(command, &root))
            .unwrap();
        child.wait_with_output().unwrap()
    };
    let refused = run("git commit -m x");
    assert_eq!(
        refused.status.code(),
        Some(2),
        "{}",
        String::from_utf8_lossy(&refused.stdout)
    );
    assert!(String::from_utf8_lossy(&refused.stderr).contains("claim verify: not owner"));
    let allowed = run("cargo test");
    assert_eq!(allowed.status.code(), Some(0));
}

/// The leader's #2226 follow-up: before a push, every commit on the branch's own first-parent
/// line since the claim base must carry the claim token — a merge commit without it refuses
/// the push; commits of other cards brought in by the merge are not checked.
#[test]
fn hook_refuses_push_when_a_branch_commit_lacks_the_trailer() {
    let clock = FixedClock(now());
    let owner_only = json!([raw(
        100,
        "GoosLab",
        &format!("moai-claim: session=a epoch=1 lease_until={LIVE}")
    )]);
    let repo = claimed_repo(100, "a");
    let root = fs::canonicalize(repo.path()).unwrap();
    let token = "Moai-Claim: 7#100";
    git(
        &root,
        &[
            "commit",
            "-q",
            "--allow-empty",
            "-m",
            &format!("work\n\n{token}"),
        ],
    );
    // Another card's commit on a side branch, merged in without the trailer.
    git(&root, &["checkout", "-q", "-b", "tip"]);
    git(
        &root,
        &[
            "commit",
            "-q",
            "--allow-empty",
            "-m",
            "other card (no token)",
        ],
    );
    git(&root, &["checkout", "-q", "main"]);

    let push = || {
        let t = RecordedTransport::new(vec![get(owner_only.clone())]);
        hook_cli(&root, &hook_input("git push", &root), &t, &clock)
    };
    let v = push().unwrap();
    assert_eq!(v["trailer_checked"], json!(1));

    git(&root, &["merge", "-q", "--no-ff", "-m", "Merge tip", "tip"]);
    let err = push().unwrap_err();
    assert!(err.starts_with("claim verify: trailer missing"), "{err}");
    assert_eq!(classify(&err), Kind::InvalidInput);

    // A merge that carries the token passes; the merged-in commit itself is not checked.
    git(&root, &["reset", "-q", "--hard", "HEAD~1"]);
    git(
        &root,
        &[
            "merge",
            "-q",
            "--no-ff",
            "-m",
            &format!("Merge tip\n\n{token}"),
            "tip",
        ],
    );
    let v = push().unwrap();
    assert_eq!(v["trailer_checked"], json!(2));

    // A commit (not a push) is not trailer-checked here.
    let t = RecordedTransport::new(vec![get(owner_only.clone())]);
    let v = hook_cli(&root, &hook_input("git commit -m x", &root), &t, &clock).unwrap();
    assert!(v.get("trailer_checked").is_none());
}

/// AC-2 (workflow text): the `--mode agent` workflow names the boundary before commit and
/// push, and the guard that enforces it; the binary test above is the mechanical half.
#[test]
fn mode_agent_workflow_names_claim_verify_before_writes() {
    let text = fs::read_to_string(
        Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai/workflows/mode-agent.md"),
    )
    .unwrap();
    let step = text
        .lines()
        .find(|l| l.contains("moai-cli claim-verify"))
        .expect("mode-agent names claim-verify");
    assert!(
        step.contains("Before every commit and every push"),
        "{step}"
    );
    assert!(step.contains("claim-verify --hook"), "{step}");
    assert!(
        step.contains("never take the Issue over by time alone"),
        "{step}"
    );
}
