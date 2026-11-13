//! #2349 (g): a typed `*-plan` result fed back to its `*-apply` on stdin is bounded by the
//! plan limit (16 MiB), not by the 64 KiB limit of hand-written requests. The plan is not
//! trusted for its size: each apply still recomputes or checks it against `--plan-digest`.
mod common;
use serde_json::Value;
use std::{
    fs,
    io::Write,
    path::Path,
    process::{Command, Stdio},
};
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;

fn run(args: &[&str], stdin: &[u8], home: &Path) -> Value {
    let mut child = Command::new(EXE)
        .args(args)
        .env("HOME", home)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    let mut pipe = child.stdin.take().unwrap();
    match pipe.write_all(stdin) {
        Ok(()) => {}
        Err(e) if e.kind() == std::io::ErrorKind::BrokenPipe => {}
        Err(e) => panic!("stdin write failed: {e}"),
    }
    drop(pipe);
    let out = child.wait_with_output().unwrap();
    serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|_| panic!("no envelope: {}", String::from_utf8_lossy(&out.stderr)))
}

fn message(v: &Value) -> String {
    v["error"]["message"].as_str().unwrap_or("").to_string()
}

/// A skill package with `files` files: SKILL.md plus workflow files.
fn stage(dir: &Path, files: usize) {
    fs::create_dir_all(dir.join("workflows")).unwrap();
    fs::write(
        dir.join("SKILL.md"),
        "---\nname: moai\ndescription: fixture package\n---\n# MoAI\n",
    )
    .unwrap();
    for i in 1..files {
        fs::write(
            dir.join(format!("workflows/step-{i:03}.md")),
            format!("# step {i}\n"),
        )
        .unwrap();
    }
}

fn update_plan(home: &Path, stage: &Path) -> Value {
    let v = run(
        &[
            "update-plan",
            home.to_str().unwrap(),
            stage.to_str().unwrap(),
        ],
        b"",
        home,
    );
    assert_eq!(v["ok"], true, "{v}");
    v["result"].clone()
}

fn update_apply(home: &Path, stage: &Path, digest: &str, stdin: &[u8]) -> Value {
    run(
        &[
            "update-apply",
            home.to_str().unwrap(),
            stage.to_str().unwrap(),
            "--plan-digest",
            digest,
            "--permission-reference",
            "fixture-approved",
        ],
        stdin,
        home,
    )
}

/// The measured failure: a 200-file package's plan is far over 64 KiB and must still
/// reach `update-apply`.
#[test]
fn update_apply_accepts_a_200_file_package() {
    let root = TempDir::new().unwrap();
    let home = root.path().join("home");
    let pkg = root.path().join("stage");
    fs::create_dir_all(&home).unwrap();
    stage(&pkg, 200);
    let plan = update_plan(&home, &pkg);
    let bytes = plan.to_string().into_bytes();
    assert!(bytes.len() > 65_536, "plan is only {} bytes", bytes.len());
    let digest = plan["plan_digest"].as_str().unwrap();
    let v = update_apply(&home, &pkg, digest, &bytes);
    assert_eq!(v["ok"], true, "{}", message(&v));
    assert_eq!(v["result"]["status"], "applied");
}

/// The plan limit at its edge, with literal values: 16_777_216 bytes are read, one more
/// byte is refused before anything is applied.
#[test]
fn plan_input_limit_edge_is_16_mib() {
    let root = TempDir::new().unwrap();
    let home = root.path().join("home");
    let pkg = root.path().join("stage");
    fs::create_dir_all(&home).unwrap();
    stage(&pkg, 3);
    let plan = update_plan(&home, &pkg);
    let digest = plan["plan_digest"].as_str().unwrap().to_string();
    let padded = |total: usize| {
        let mut b = plan.to_string().into_bytes();
        b.resize(total, b' ');
        b
    };

    let over = update_apply(&home, &pkg, &digest, &padded(16_777_217));
    assert_eq!(over["ok"], false);
    assert_eq!(message(&over), "plan input exceeds 16 MiB");
    assert!(
        !home.join(".claude/skills/moai/SKILL.md").exists(),
        "a refused plan applied nothing"
    );

    let at = update_apply(&home, &pkg, &digest, &padded(16_777_216));
    assert_eq!(at["ok"], true, "{}", message(&at));
    assert_eq!(at["result"]["status"], "applied");
}

/// Every apply that takes a plan reads it under the plan limit; a hand-written request
/// keeps the 64 KiB limit.
#[test]
fn every_plan_apply_uses_the_plan_limit() {
    let root = TempDir::new().unwrap();
    let home = root.path().join("home");
    let project = root.path().join("project");
    let pkg = root.path().join("stage");
    fs::create_dir_all(&home).unwrap();
    fs::create_dir_all(&project).unwrap();
    stage(&pkg, 1);
    let mut large = b"[]".to_vec();
    large.resize(70_000, b' ');
    let p = project.to_str().unwrap();
    let h = home.to_str().unwrap();
    let s = pkg.to_str().unwrap();
    let flags = ["--plan-digest", "d", "--permission-reference", "r"];
    let cases: [Vec<&str>; 6] = [
        vec!["project-apply", p],
        vec!["agents-apply", p],
        vec!["update-apply", h, s],
        vec!["recovery-apply", h],
        vec!["github-apply", p],
        vec!["generation-apply", p],
    ];
    for case in cases {
        let mut args = case.clone();
        args.extend(flags);
        let v = run(&args, &large, &home);
        assert_eq!(
            message(&v),
            "invalid typed JSON input",
            "{} must read a 70 000-byte plan and then reject its shape",
            case[0]
        );
    }
    let v = run(&["plan-admit", p], &large, &home);
    assert_eq!(message(&v), "input exceeds 64 KiB");
}
