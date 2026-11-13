//! PLAN and task DAG admission (Issue #2002, 05-B).
mod common;
use moai_cli::plan::{admit, build_plan, current_revision, AdmissionRequest, Approval, Task};
use moai_cli::spec::{amend, approve, SortedKeys, SpecDocument};
use serde_json::{json, Value};
use std::{
    fs,
    io::Write,
    path::Path,
    process::{Command, Stdio},
};

const REV: &str = "1111111111111111111111111111111111111111";

fn root(rel: &str) -> std::path::PathBuf {
    Path::new(env!("CARGO_MANIFEST_DIR")).join(rel)
}
fn tasks(name: &str) -> Vec<Task> {
    serde_json::from_slice(&fs::read(root(&format!("tests/fixtures/plan/{name}.json"))).unwrap())
        .unwrap()
}
/// The 05-A fixture, approved through the 05-A API so its snapshot digest is real.
fn approved() -> SpecDocument {
    let d: SpecDocument = serde_json::from_slice(
        &fs::read(root("tests/fixtures/spec/snapshot_canonical_a.json")).unwrap(),
    )
    .unwrap();
    approve(&d, &SortedKeys)
}
fn approval_for(spec: &SpecDocument) -> Approval {
    let snap = spec.snapshot.clone().expect("approved");
    Approval {
        id: "APR-1".into(),
        approver: "operator".into(),
        basis: "plan review 2026-11-08".into(),
        snapshot_reference: snap.reference,
        snapshot_digest: snap.digest,
        allowed_actions: vec!["run".into()],
        revoked: false,
    }
}
fn request(name: &str) -> AdmissionRequest {
    let spec = approved();
    AdmissionRequest {
        protocol_version: 1,
        approval: approval_for(&spec),
        spec,
        tasks: tasks(name),
        expected_revision: REV.into(),
        capabilities: vec!["claude".into()],
    }
}
fn decide(req: &AdmissionRequest, current: &str) -> Value {
    serde_json::to_value(admit(req, current)).unwrap()
}
fn reasons(v: &Value) -> Vec<String> {
    v["reasons"]
        .as_array()
        .unwrap()
        .iter()
        .map(|r| r.as_str().unwrap().to_string())
        .collect()
}

#[test]
fn dag_cycle_is_refused_with_only_its_members() {
    let req = request("dag_cycle");
    let plan = serde_json::to_value(build_plan(&req.spec, &req.tasks, REV)).unwrap();
    assert_eq!(plan["order"], json!([]));
    // E depends on the cycle but is not part of it.
    assert_eq!(plan["cycles"], json!([["A", "B", "C"], ["D"]]));
    let v = decide(&req, REV);
    assert_eq!(v["decision"], "blocked");
    assert!(reasons(&v).contains(&"cycle".to_string()));
    assert!(v["pinned"].as_array().unwrap().is_empty());
}

#[test]
fn dag_linear_order_matches_golden_and_repeats() {
    let req = request("dag_linear");
    let golden: Value = serde_json::from_slice(
        &fs::read(root("tests/fixtures/plan/golden/dag_linear_order.json")).unwrap(),
    )
    .unwrap();
    let a = build_plan(&req.spec, &req.tasks, REV);
    let b = build_plan(&req.spec, &req.tasks, REV);
    assert_eq!(serde_json::to_value(&a.order).unwrap(), golden);
    assert_eq!(a.order, b.order);
    assert_eq!(a.plan_digest, b.plan_digest);
    assert!(a.plan_digest.starts_with("sha256:") && a.plan_digest.len() == 71);
    let v = decide(&req, REV);
    assert_eq!(v["decision"], "ready", "{v}");
    assert_eq!(v["pinned"][0]["task_id"], "T-types");
    assert_eq!(v["pinned"][0]["expected_revision"], REV);
    assert_eq!(v["pinned"][0]["approval_ref"], "APR-1");
    assert_eq!(v["pinned"][0]["spec_ref"], "SPEC-T-SNAP@1");
}

#[test]
fn dag_tie_breaks_by_id_and_task_input_order_does_not_matter() {
    let req = request("dag_tie");
    let plan = build_plan(&req.spec, &req.tasks, REV);
    assert_eq!(plan.order, ["root", "alpha", "mid", "zeta"]);
    let mut reversed = req.tasks.clone();
    reversed.reverse();
    let again = build_plan(&req.spec, &reversed, REV);
    assert_eq!(again.order, plan.order);
    assert_eq!(again.plan_digest, plan.plan_digest);
}

#[test]
fn dangling_dependency_fails_structure_and_blocks() {
    let req = request("dangling_dependency");
    let v = decide(&req, REV);
    assert_eq!(v["plan"]["structure"], "failed");
    assert_eq!(
        v["plan"]["dangling"],
        json!([{"from": "T1", "to": "T-missing"}])
    );
    assert_eq!(v["decision"], "blocked");
    assert!(reasons(&v).contains(&"dangling-dependency".to_string()));
}

#[test]
fn admission_rejects_mismatch_of_snapshot_revision_or_reused_approval() {
    // One byte of the recorded snapshot digest changed.
    let mut req = request("dag_linear");
    let snap = req.spec.snapshot.as_mut().unwrap();
    let last = snap.digest.pop().unwrap();
    snap.digest.push(if last == '0' { '1' } else { '0' });
    let v = decide(&req, REV);
    assert_eq!(v["decision"], "stale");
    assert!(reasons(&v).contains(&"snapshot-digest-mismatch".to_string()));

    // The repository moved after the plan pinned its revision.
    let other = "2222222222222222222222222222222222222222";
    let v = decide(&request("dag_linear"), other);
    assert_eq!(v["decision"], "stale");
    assert_eq!(v["current_revision"], other);
    assert!(reasons(&v).contains(&"expected-revision-mismatch".to_string()));

    // A body change opens revision 2; the approval of revision 1 is not reused for it.
    let req = request("dag_linear");
    let mut changed = req.spec.requirements.clone();
    changed[0].text = changed[0].text.replace("receipt", "Receipt");
    let next = approve(
        &amend(&req.spec, changed, req.spec.acceptance.clone()),
        &SortedKeys,
    );
    let reused = AdmissionRequest {
        spec: next,
        ..request("dag_linear")
    };
    let v = decide(&reused, REV);
    assert_eq!(v["decision"], "stale");
    assert!(reasons(&v).contains(&"approval-reference-mismatch".to_string()));
    assert!(reasons(&v).contains(&"approval-snapshot-mismatch".to_string()));

    // An amended revision that is not approved yet cannot be admitted at all.
    let req = request("dag_linear");
    let proposed = amend(
        &req.spec,
        req.spec.requirements.clone(),
        req.spec.acceptance.clone(),
    );
    let v = decide(
        &AdmissionRequest {
            spec: proposed,
            ..request("dag_linear")
        },
        REV,
    );
    assert_ne!(v["decision"], "ready");
    assert!(reasons(&v).contains(&"not-approved".to_string()));
    assert!(v["pinned"].as_array().unwrap().is_empty());
}

#[test]
fn blocked_reasons_name_the_cause_and_never_promote() {
    // A task with no verifier is blocked and a verifier task is suggested before it.
    let v = decide(&request("verifier_missing"), REV);
    assert_eq!(v["decision"], "blocked");
    assert!(reasons(&v).contains(&"verifier-missing".to_string()));
    let blocked = v["plan"]["blocked"].as_array().unwrap();
    assert_eq!(blocked.len(), 1, "{v}");
    assert_eq!(blocked[0]["task"], "T-none");
    assert_eq!(blocked[0]["reason"], "verifier-missing");
    assert!(blocked[0]["suggestion"]
        .as_str()
        .unwrap()
        .contains("verifier"));

    // A criterion whose verification has no selector is no verifier either.
    let mut req = request("dag_linear");
    let mut spec = req.spec.clone();
    spec.acceptance[0].verification.selector = String::new();
    req.spec = approve(&spec, &SortedKeys);
    req.approval = approval_for(&req.spec);
    let v = decide(&req, REV);
    assert_eq!(v["decision"], "blocked");
    assert_eq!(v["plan"]["blocked"].as_array().unwrap().len(), 4);

    // A revoked approval blocks admission.
    let mut req = request("dag_linear");
    req.approval.revoked = true;
    let v = decide(&req, REV);
    assert_eq!(v["decision"], "blocked");
    assert!(reasons(&v).contains(&"approval-revoked".to_string()));

    // An approval that does not allow `run` blocks admission.
    let mut req = request("dag_linear");
    req.approval.allowed_actions = vec!["plan".into()];
    assert!(reasons(&decide(&req, REV)).contains(&"approval-scope".to_string()));

    // A task bound to a host the session does not have is blocked.
    let mut req = request("dag_linear");
    req.tasks[0].host = Some("codex".into());
    let v = decide(&req, REV);
    assert_eq!(v["decision"], "blocked");
    assert!(reasons(&v).contains(&"capability-missing".to_string()));

    // Execution stays not_run on every path, ready included.
    for name in ["verifier_missing", "dag_cycle", "dag_linear"] {
        let v = decide(&request(name), REV);
        assert_eq!(v["plan"]["execution"], "not_run", "{name}");
    }
}

#[test]
fn scope_overlap_without_dependency_is_not_parallel() {
    let req = request("scope_overlap");
    let v = decide(&req, REV);
    assert_eq!(
        v["plan"]["parallel_conflicts"],
        json!([
            {"tasks": ["T1", "T2"], "paths": ["src/receipt", "src/receipt/apply.rs"]},
            {"tasks": ["T2", "T3"], "paths": ["src/receipt", "src/receipt/apply.rs"]}
        ])
    );
    // A conflict is a warning about parallelism, not a refusal.
    assert_eq!(v["decision"], "ready");
}

// ---- real git repository (AC-05B-5, AC-05B-6) --------------------------------------------

fn git_available() -> bool {
    Command::new("git")
        .arg("--version")
        .output()
        .is_ok_and(|o| o.status.success())
}
fn git(dir: &Path, args: &[&str]) {
    let out = Command::new("git")
        .arg("-C")
        .arg(dir)
        .args([
            "-c",
            "user.name=fixture",
            "-c",
            "user.email=fixture@example.invalid",
        ])
        .args(["-c", "commit.gpgsign=false"])
        .args(["-c", "gc.auto=0", "-c", "maintenance.auto=false"])
        .args(args)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_NOSYSTEM", "1")
        .output()
        .unwrap();
    assert!(
        out.status.success(),
        "git {args:?}: {}",
        String::from_utf8_lossy(&out.stderr)
    );
}
fn repo() -> tempfile::TempDir {
    let dir = tempfile::TempDir::new().unwrap();
    git(dir.path(), &["init", "-q"]);
    fs::write(dir.path().join("a.txt"), "one\n").unwrap();
    git(dir.path(), &["add", "a.txt"]);
    git(dir.path(), &["commit", "-q", "-m", "one"]);
    dir
}
fn plan_admit(project: &Path, input: &[u8]) -> Value {
    let (ok, v) = plan_admit_env(project, input, &[]);
    assert!(ok, "{v}");
    v
}
/// `plan-admit` with extra caller environment; returns (exit 0, envelope).
fn plan_admit_env(project: &Path, input: &[u8], envs: &[(&str, &str)]) -> (bool, Value) {
    let mut child = Command::new(crate::common::MOAI_BIN)
        .arg("plan-admit")
        .arg(project)
        .envs(envs.iter().copied())
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    match child.stdin.take().unwrap().write_all(input) {
        Ok(()) => {}
        Err(e) if e.kind() == std::io::ErrorKind::BrokenPipe => {}
        Err(e) => panic!("stdin write failed: {e}"),
    }
    let out = child.wait_with_output().unwrap();
    let v = serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|_| panic!("no envelope: {}", String::from_utf8_lossy(&out.stderr)));
    (out.status.success(), v)
}

/// Issue #2043 AC-BUG-1: caller variables that point git at another repository or inject
/// configuration must not change which HEAD admission reads; other git failures are
/// `git_failed` as in #2033.
#[test]
fn plan_revision_ignores_injected_git_environment() {
    if !git_available() {
        eprintln!("not_run: git is not available on this machine");
        return;
    }
    let a = repo();
    let b = repo();
    fs::write(b.path().join("a.txt"), "other\n").unwrap();
    git(b.path(), &["commit", "-q", "-am", "other"]);
    let head_a = current_revision(a.path()).unwrap();
    let head_b = current_revision(b.path()).unwrap();
    assert_ne!(head_a, head_b);
    let mut req = request("dag_linear");
    req.expected_revision = head_a.clone();
    let input = serde_json::to_vec(&req).unwrap();
    let b_git = b.path().join(".git");
    let b_git = b_git.to_str().unwrap();
    let b_dir = b.path().to_str().unwrap();
    let params = format!("'core.worktree'='{b_dir}'");
    let cases: [&[(&str, &str)]; 6] = [
        &[("GIT_DIR", b_git)],
        &[("GIT_DIR", b_git), ("GIT_WORK_TREE", b_dir)],
        &[("GIT_COMMON_DIR", b_git)],
        &[("GIT_CONFIG_PARAMETERS", &params)],
        &[
            ("GIT_CONFIG_COUNT", "1"),
            ("GIT_CONFIG_KEY_0", "core.bare"),
            ("GIT_CONFIG_VALUE_0", "true"),
        ],
        &[("GIT_CEILING_DIRECTORIES", a.path().to_str().unwrap())],
    ];
    for envs in cases {
        let (ok, v) = plan_admit_env(a.path(), &input, envs);
        assert!(ok, "{envs:?}: {v}");
        assert_eq!(
            v["result"]["current_revision"],
            json!(head_a),
            "{envs:?} moved HEAD"
        );
        assert_eq!(v["result"]["decision"], "ready", "{envs:?}");
    }
    // Pinned to B's HEAD under a GIT_DIR pointing at B: still A's HEAD, so stale, not ready.
    let mut wrong = request("dag_linear");
    wrong.expected_revision = head_b.clone();
    let (ok, v) = plan_admit_env(
        a.path(),
        &serde_json::to_vec(&wrong).unwrap(),
        &[("GIT_DIR", b_git)],
    );
    assert!(ok);
    assert_eq!(v["result"]["decision"], "stale");

    // git present but refusing (safe.directory ownership): git_failed, no path in the message.
    let (ok, v) = plan_admit_env(
        a.path(),
        &input,
        &[
            ("GIT_TEST_ASSUME_DIFFERENT_OWNER", "1"),
            ("GIT_CONFIG_GLOBAL", "/dev/null"),
            ("GIT_CONFIG_NOSYSTEM", "1"),
        ],
    );
    assert!(!ok);
    assert_eq!(v["error"]["kind"], "git_failed", "{v}");
    let msg = v["error"]["message"].as_str().unwrap();
    assert!(
        msg.starts_with("git failed: ") && !msg.contains('/'),
        "{msg}"
    );
    // Not a repository keeps its earlier, non-git_failed classification.
    let plain = tempfile::TempDir::new().unwrap();
    let (ok, v) = plan_admit_env(plain.path(), &input, &[("GIT_DIR", b_git)]);
    assert!(!ok);
    assert_eq!(v["error"]["kind"], "not_found", "{v}");
}

#[test]
fn expected_revision_from_git_moves_admission_to_stale_after_a_commit() {
    if !git_available() {
        eprintln!("not_run: git is not available on this machine");
        return;
    }
    let dir = repo();
    let head = current_revision(dir.path()).unwrap();
    assert_eq!(head.len(), 40);
    let mut req = request("dag_linear");
    req.expected_revision = head.clone();
    assert_eq!(decide(&req, &head)["decision"], "ready");
    fs::write(dir.path().join("a.txt"), "two\n").unwrap();
    git(dir.path(), &["commit", "-q", "-am", "two"]);
    let moved = current_revision(dir.path()).unwrap();
    assert_ne!(moved, head);
    let v = decide(&req, &moved);
    assert_eq!(v["decision"], "stale");
    assert!(reasons(&v).contains(&"expected-revision-mismatch".to_string()));
    // Not a repository: an error, never a guessed revision.
    let plain = tempfile::TempDir::new().unwrap();
    assert!(current_revision(plain.path()).is_err());
}

#[test]
fn plan_is_idempotent_across_processes() {
    if !git_available() {
        eprintln!("not_run: git is not available on this machine");
        return;
    }
    let dir = repo();
    let mut req = request("dag_linear");
    req.expected_revision = current_revision(dir.path()).unwrap();
    // Saved by one process, re-read by two fresh ones.
    let saved = tempfile::TempDir::new().unwrap();
    let file = saved.path().join("request.json");
    fs::write(&file, serde_json::to_vec(&req).unwrap()).unwrap();
    let first = plan_admit(dir.path(), &fs::read(&file).unwrap());
    let second = plan_admit(dir.path(), &fs::read(&file).unwrap());
    assert_eq!(first["ok"], json!(true));
    assert_eq!(first["operation"], "plan.admit");
    assert_eq!(first["result"]["decision"], "ready", "{first}");
    assert_eq!(first, second);
    let in_process = build_plan(&req.spec, &req.tasks, &req.expected_revision);
    assert_eq!(
        first["result"]["plan"]["plan_digest"],
        json!(in_process.plan_digest)
    );
}
