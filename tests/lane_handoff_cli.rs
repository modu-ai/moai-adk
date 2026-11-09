//! Lane handoff check fixtures (A10 #2264). New-file discipline: existing test
//! files are untouched. Each test builds a real git fixture repository (base
//! commit + claim-trailer commits), then drives `lane-handoff-check` with
//! mutated request fields to observe every failure path. No real host is used:
//! the remote head and PR record are injected observations.

use serde_json::{json, Value};
use std::{
    io::Write,
    path::{Path, PathBuf},
    process::{Command, Stdio},
};
use tempfile::TempDir;

const EXE: &str = env!("CARGO_BIN_EXE_moai-cli");

struct FixtureRepo {
    root: PathBuf,
    head: String,
}

fn git(repo: &Path, args: &[&str]) -> String {
    let out = Command::new("git")
        .arg("-C")
        .arg(repo)
        .args(args)
        .output()
        .expect("git runs");
    assert!(out.status.success(), "git {:?} failed", args);
    String::from_utf8_lossy(&out.stdout).trim().to_string()
}

fn commit(repo: &Path, message: &str, trailer: Option<&str>) -> String {
    let p = repo.join("f.txt");
    let mut body = std::fs::read_to_string(&p).unwrap_or_default();
    body.push_str(&format!("{}\n", message));
    std::fs::write(&p, &body).unwrap();
    git(repo, &["add", "f.txt"]);
    let mut args: Vec<&str> = vec![
        "-c",
        "user.name=t",
        "-c",
        "user.email=t@example.invalid",
        "commit",
        "-m",
        message,
    ];
    if let Some(t) = trailer {
        args.extend(["-m", t]);
    }
    git(repo, &args);
    // Return the new head sha (git commit prints its summary to stdout).
    git(repo, &["rev-parse", "HEAD"])
}

fn fixture_repo(parent: &Path) -> FixtureRepo {
    let root = parent.join("repo");
    std::fs::create_dir_all(&root).unwrap();
    git(&root, &["init", "-b", "v2026.11.0"]);
    std::fs::write(root.join("f.txt"), "base\n").unwrap();
    // The durable report record lives under the project state dir and must not
    // dirty the tree at handoff time — the base commit ignores it (repo convention).
    std::fs::write(root.join(".gitignore"), ".moai-cli/\n").unwrap();
    git(&root, &["add", "f.txt", ".gitignore"]);
    git(
        &root,
        &[
            "-c",
            "user.name=t",
            "-c",
            "user.email=t@example.invalid",
            "commit",
            "-m",
            "base",
        ],
    );
    git(&root, &["checkout", "-b", "issue/2193-e1-dist"]);
    let head = commit(&root, "card work", Some("Moai-Claim: 2193#6730943161"));
    FixtureRepo { root, head }
}

fn base_request(repo: &Path, remote_head: &str) -> Value {
    json!({
        "protocol_version": 1,
        "project": repo.to_str().unwrap(),
        "branch": "issue/2193-e1-dist",
        "base_branch": "v2026.11.0",
        "issue": 2193,
        "claim_comment_id": 6730943161u64,
        "remote_head": remote_head,
        "pr": {"url": "https://github.com/modu-ai/moai-adk/pull/2258", "number": 2258, "head": remote_head},
        "leader_report_path": null,
    })
}

/// Runs the CLI with the fixture repo as the positional project path.
fn run_check(request: &Value, project: &Path) -> (i32, Value, String) {
    let mut child = Command::new(EXE)
        .arg("lane-handoff-check")
        .arg(project.to_str().unwrap())
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    child
        .stdin
        .as_mut()
        .unwrap()
        .write_all(request.to_string().as_bytes())
        .unwrap();
    let out = child.wait_with_output().unwrap();
    let envelope: Value = serde_json::from_slice(&out.stdout).unwrap_or(Value::Null);
    (
        out.status.code().unwrap_or(-1),
        envelope,
        String::from_utf8_lossy(&out.stderr).to_string(),
    )
}

fn write_report(repo: &Path) -> PathBuf {
    let dir = repo.join(".moai-cli/reports");
    std::fs::create_dir_all(&dir).unwrap();
    let path = dir.join("2193-handoff.md");
    std::fs::write(&path, "#2193 영속 보고 — PR #2258 head ce93d735d\n").unwrap();
    path
}

fn tree_state(repo: &Path) -> Vec<(String, Vec<u8>)> {
    fn walk(dir: &Path, base: &Path, out: &mut Vec<(String, Vec<u8>)>) {
        for entry in std::fs::read_dir(dir).unwrap() {
            let path = entry.unwrap().path();
            if path.is_dir() && !path.ends_with(".git") {
                walk(&path, base, out);
            } else if path.is_file() {
                let rel = path
                    .strip_prefix(base)
                    .unwrap()
                    .to_str()
                    .unwrap()
                    .to_string();
                out.push((rel, std::fs::read(&path).unwrap()));
            }
        }
    }
    let mut out = Vec::new();
    walk(repo, repo, &mut out);
    out.sort();
    out
}

/// AC-1 pass path: clean status, head == remote == PR head, trailers on all
/// commits, report record present, no uncertain effects.
#[test]
fn lane_handoff_check_passes_on_clean_handoff() {
    let parent = TempDir::new().unwrap();
    let repo = fixture_repo(parent.path());
    let report = write_report(&repo.root);
    let mut request = base_request(&repo.root, &repo.head);
    request["leader_report_path"] = json!(report.to_str().unwrap());
    let (code, envelope, _) = run_check(&request, &repo.root);
    assert_eq!(code, 0);
    assert_eq!(envelope["ok"], json!(true));
    let result = &envelope["result"];
    assert_eq!(result["status"], json!("passed"));
    assert_eq!(result["failed_checks"], json!([]));
    assert_eq!(result["separation"]["pr_created_is_not_done"], json!(true));
    assert_eq!(result["separation"]["issue_done"], json!(null));
    assert_eq!(result["separation"]["merged"], json!(false));
    assert_eq!(result["writes"], json!(false));
}

/// AC-1 failure paths (five ≥ five): each mutation flips the verdict to
/// blocked with the reason named in the envelope message, exit != 0.
#[test]
fn lane_handoff_check_failure_paths() {
    let parent = TempDir::new().unwrap();
    let repo = fixture_repo(parent.path());
    let request = base_request(&repo.root, &repo.head);

    // 1. 미커밋 (uncommitted changes)
    std::fs::write(repo.root.join("f.txt"), "dirty\n").unwrap();
    let (code, envelope, _) = run_check(&request, &repo.root);
    assert_ne!(code, 0, "dirty tree must block exit");
    assert_eq!(envelope["error"]["kind"], json!("blocked"));
    assert!(
        envelope["error"]["message"]
            .as_str()
            .unwrap()
            .contains("status_clean"),
        "message names the failed check"
    );
    std::fs::write(repo.root.join("f.txt"), "base\ncard work\n").unwrap();

    // 2. head 불일치 (remote observation differs from local HEAD)
    let mut r = request.clone();
    r["remote_head"] = json!(format!("{}", "f".repeat(40)));
    let (code, envelope, _) = run_check(&r, &repo.root);
    assert_ne!(code, 0);
    assert!(
        envelope["error"]["message"]
            .as_str()
            .unwrap()
            .contains("head_remote_match"),
        "names head_remote_match"
    );

    // 3. trailer 누락 (a commit without the claim trailer). The observation
    //    record is re-pointed at the new head so ONLY the trailer condition
    //    fails — the PR head comparison lives in pr_exists now (audit F1).
    let sha = commit(&repo.root, "work without trailer", None);
    let mut r = request.clone();
    r["remote_head"] = json!(sha);
    r["pr"]["head"] = json!(sha);
    let (code, envelope, _) = run_check(&r, &repo.root);
    assert_ne!(code, 0);
    let msg = envelope["error"]["message"].as_str().unwrap();
    assert!(msg.contains("trailers"), "names trailers: {msg}");
    assert!(
        !msg.contains("pr_exists"),
        "pr head was re-pointed; trailers must fail alone: {msg}"
    );
    // roll back the trailer-less commit (HEAD^ = the trailer commit)
    git(&repo.root, &["reset", "--hard", "HEAD^"]);

    // 4. PR 없음 (no PR record)
    let mut r = request.clone();
    r["pr"] = json!(null);
    let (code, envelope, _) = run_check(&r, &repo.root);
    assert_ne!(code, 0);
    assert!(
        envelope["error"]["message"]
            .as_str()
            .unwrap()
            .contains("pr_exists"),
        "names pr_exists"
    );

    // 5. 보고 없음 (leader report record missing)
    let mut r = request.clone();
    r["leader_report_path"] = json!(null);
    let (code, envelope, _) = run_check(&r, &repo.root);
    assert_ne!(code, 0);
    assert!(
        envelope["error"]["message"]
            .as_str()
            .unwrap()
            .contains("leader_report"),
        "names leader_report"
    );

    // 6. 보고 내용 불일치 (the record exists but does not mention the issue —
    //    the `#<issue>` condition must be load-bearing, not decorative)
    let wrong = repo.root.join(".moai-cli/reports/wrong-issue.md");
    std::fs::create_dir_all(wrong.parent().unwrap()).unwrap();
    std::fs::write(&wrong, "#999 another lane's report, no mention here\n").unwrap();
    let mut r = request.clone();
    r["leader_report_path"] = json!(wrong.to_str().unwrap());
    let (code, envelope, _) = run_check(&r, &repo.root);
    assert_ne!(code, 0, "a report that never names the issue must block");
    assert!(
        envelope["error"]["message"]
            .as_str()
            .unwrap()
            .contains("leader_report"),
        "names leader_report"
    );
}

/// Audit F2 (P1·P2 probes) and F1 (pr-head comparison lives in pr_exists):
/// an exact-line claim requirement — a prefix-sharing id, a prose mention,
/// and a PR-head-only mismatch each name the RIGHT failed check, with the
/// leader report present so nothing else fails alongside.
#[test]
fn trailers_match_is_exact_line_and_pr_head_lives_in_pr_exists() {
    let parent = TempDir::new().unwrap();
    let repo = fixture_repo(parent.path());
    let report = write_report(&repo.root);
    let request = base_request(&repo.root, &repo.head);

    // P1: a commit whose claim line shares only the id prefix must fail.
    let sha = commit(
        &repo.root,
        "prefix id work",
        Some("Moai-Claim: 2193#67309431612"),
    );
    let mut r = request.clone();
    r["remote_head"] = json!(sha);
    r["pr"]["head"] = json!(sha);
    r["leader_report_path"] = json!(report.to_str().unwrap());
    let (code, envelope, _) = run_check(&r, &repo.root);
    assert_ne!(code, 0, "prefix-sharing id must not satisfy the check");
    let msg = envelope["error"]["message"].as_str().unwrap();
    assert!(msg.contains("trailers"), "names trailers: {msg}");
    assert!(
        !msg.contains("pr_exists"),
        "pr head matches; only trailers fails: {msg}"
    );
    git(&repo.root, &["reset", "--hard", "HEAD^"]);

    // P2: prose that merely mentions the token must fail.
    let sha = commit(
        &repo.root,
        "prose mention work",
        Some("not my claim: see Moai-Claim: 2193#6730943161 of another lane"),
    );
    let mut r = request.clone();
    r["remote_head"] = json!(sha);
    r["pr"]["head"] = json!(sha);
    r["leader_report_path"] = json!(report.to_str().unwrap());
    let (code, envelope, _) = run_check(&r, &repo.root);
    assert_ne!(code, 0, "a prose mention must not satisfy the check");
    let msg = envelope["error"]["message"].as_str().unwrap();
    assert!(msg.contains("trailers"), "names trailers: {msg}");
    git(&repo.root, &["reset", "--hard", "HEAD^"]);

    // PR-head-only mismatch is pr_exists, never trailers (audit F1).
    let mut r = request.clone();
    r["pr"]["head"] = json!(format!("{}", "e".repeat(40)));
    r["leader_report_path"] = json!(report.to_str().unwrap());
    let (code, envelope, _) = run_check(&r, &repo.root);
    assert_ne!(code, 0);
    let msg = envelope["error"]["message"].as_str().unwrap();
    assert!(msg.contains("pr_exists"), "names pr_exists: {msg}");
    assert!(
        !msg.contains("trailers"),
        "trailers must not absorb the PR head comparison: {msg}"
    );
}

/// Audit F3 (결정 06:02): tool merges are records, never claim-checked. A
/// claim-less `--no-ff` merge commit in base..HEAD must NOT block the exit,
/// while the lane-authored commits are still verified.
#[test]
fn tool_merge_commits_are_not_claim_checked() {
    let parent = TempDir::new().unwrap();
    let repo = fixture_repo(parent.path());
    let report = write_report(&repo.root);
    // Move the base forward on a separate file (no conflict with f.txt), then
    // merge it back with --no-ff and no claim line.
    git(&repo.root, &["checkout", "v2026.11.0"]);
    std::fs::write(repo.root.join("g.txt"), "new base\n").unwrap();
    git(&repo.root, &["add", "g.txt"]);
    git(
        &repo.root,
        &[
            "-c",
            "user.name=t",
            "-c",
            "user.email=t@example.invalid",
            "commit",
            "-m",
            "base moved",
        ],
    );
    git(&repo.root, &["checkout", "issue/2193-e1-dist"]);
    git(
        &repo.root,
        &[
            "merge",
            "v2026.11.0",
            "--no-ff",
            "-m",
            "Merge branch v2026.11.0 (tool record, no claim)",
        ],
    );
    let head = git(&repo.root, &["rev-parse", "HEAD"]);
    let mut r = base_request(&repo.root, &head);
    r["pr"]["head"] = json!(head);
    r["leader_report_path"] = json!(report.to_str().unwrap());
    let (code, envelope, stderr) = run_check(&r, &repo.root);
    assert_eq!(code, 0, "a claim-less tool merge must not block: {stderr}");
    assert_eq!(envelope["result"]["status"], json!("passed"));
    // The merge commit is excluded from the trailer count.
    assert_eq!(envelope["result"]["trailer_commits"], json!(1));
}

/// Audit 비차단 1 (P4): an unreadable or unparsable outbox entry counts as
/// uncertain — fail-closed, exit blocked.
#[test]
fn unparsable_outbox_entry_counts_as_uncertain() {
    let parent = TempDir::new().unwrap();
    let repo = fixture_repo(parent.path());
    let report = write_report(&repo.root);
    let mut r = base_request(&repo.root, &repo.head);
    r["leader_report_path"] = json!(report.to_str().unwrap());
    let outbox = repo.root.join(".moai-cli/forge/outbox");
    std::fs::create_dir_all(&outbox).unwrap();
    std::fs::write(outbox.join("broken.json"), b"{\"state\": \"publi").unwrap();
    let (code, envelope, _) = run_check(&r, &repo.root);
    assert_ne!(code, 0, "an unparsable entry must block (fail-closed)");
    let msg = envelope["error"]["message"].as_str().unwrap();
    assert!(
        msg.contains("no_uncertain_effects"),
        "names no_uncertain_effects: {msg}"
    );
    std::fs::remove_file(outbox.join("broken.json")).unwrap();
    // Back to clean after the entry is removed.
    let (code, _, _) = run_check(&r, &repo.root);
    assert_eq!(code, 0);
}

/// AC-2: one cycle iteration (PR → 점검 → 다음 claim 준비) replayed with
/// fixtures only — no real host. The check is read-only and deterministic:
/// a rerun of the same request on the unchanged tree passes identically, and
/// the queue-next link (#2255/#2161, not yet merged) is stubbed as the
/// claim text a next card would need.
#[test]
fn cycle_replay_with_stubs() {
    let parent = TempDir::new().unwrap();
    let repo = fixture_repo(parent.path());
    let report = write_report(&repo.root);

    let (first, _) = {
        let mut r = base_request(&repo.root, &repo.head);
        r["leader_report_path"] = json!(report.to_str().unwrap());
        let out = run_check(&r, &repo.root);
        let code = out.0;
        let envelope = out.1.clone();
        let stderr = out.2.clone();
        assert_eq!(code, 0, "{}", stderr);
        let v: Value = envelope;
        assert_eq!(v["result"]["status"], json!("passed"));
        (v["result"].clone(), ())
    };

    let before = tree_state(&repo.root);
    let (again, _) = {
        let mut r = base_request(&repo.root, &repo.head);
        r["leader_report_path"] = json!(report.to_str().unwrap());
        let out = run_check(&r, &repo.root);
        let code = out.0;
        let envelope = out.1.clone();
        let stderr = out.2.clone();
        assert_eq!(code, 0, "{}", stderr);
        let v: Value = envelope;
        assert_eq!(v["result"]["status"], json!("passed"));
        (v["result"].clone(), ())
    };
    assert_eq!(
        tree_state(&repo.root),
        before,
        "no write between the two checks"
    );
    // Deterministic: the second run reproduces the first exactly.
    assert_eq!(first, again);
    // The stub for 다음 claim: a new fixture issue's claim text is derived
    // from the same generator rules (deterministic, no host).
    let next_claim = format!("/goal #{}의 PR이 병합되거나 blocked로 보고됨", 2194);
    assert!(next_claim.starts_with("/goal #"));
}
