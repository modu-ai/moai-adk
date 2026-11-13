//! #2268 (rc.2 B-defect) fail-closed fixtures. New-file discipline: existing test
//! files are untouched — the helpers here are a minimal copy of update_cli.rs's.
//!
//! AC-1: the agent-9 repro (a non-RC1 SKILL.md already in the projection root) plans
//! as blocked and update-apply writes nothing (projection tree sha256 identical).
//! REQ-2: the reason distinguishes the unmanaged existing package from a per-file
//! conflict and names the safe alternatives. AC-2: the managed update path is
//! unchanged. REQ-3: recovery-plan lists the files a first apply added into a
//! pre-existing root. Real HOME is never touched — every fixture runs in a
//! temporary HOME.

#![cfg(unix)]

mod common;
use moai_cli::tree_digest;
use serde_json::{json, Value};
use std::{
    fs,
    io::Write as _,
    path::Path,
    process::{Command, Stdio},
};
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;

fn write_pkg(dir: &Path, files: &[(&str, String)]) {
    for (rel, body) in files {
        let p = dir.join(rel);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        fs::write(p, body).unwrap();
    }
}

fn skill_md(body: &str) -> String {
    format!("---\nname: moai\ndescription: fixture package\n---\n# MoAI\n{body}\n")
}

fn stage_v1() -> Vec<(&'static str, String)> {
    vec![
        ("SKILL.md", skill_md("body v1")),
        ("references/routes.json", "{\"routes\":1}\n".into()),
        ("workflows/update.md", "# UPDATE\nv1\n".into()),
    ]
}

fn stage_v2() -> Vec<(&'static str, String)> {
    vec![
        ("SKILL.md", skill_md("body v2")),
        ("references/routes.json", "{\"routes\":1}\n".into()),
        ("workflows/update.md", "# UPDATE\nv2\n".into()),
    ]
}

fn run_cli(args: &[&str], stdin: Option<&Value>, home: &Path) -> std::process::Output {
    let mut child = Command::new(EXE)
        .args(args)
        .env("HOME", home)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    let pipe = child.stdin.take().unwrap();
    match stdin {
        Some(v) => {
            let mut pipe = pipe;
            let data = v.to_string();
            match pipe.write_all(data.as_bytes()) {
                Ok(()) => {}
                Err(e) if e.kind() == std::io::ErrorKind::BrokenPipe => {}
                Err(e) => panic!("stdin write failed: {e}"),
            }
        }
        None => drop(pipe),
    }
    child.wait_with_output().unwrap()
}

fn result_of(out: &std::process::Output) -> Value {
    let v: Value = serde_json::from_slice(&out.stdout).expect("envelope JSON on stdout");
    assert_eq!(
        v["ok"],
        json!(true),
        "{}",
        String::from_utf8_lossy(&out.stdout)
    );
    v["result"].clone()
}

fn plan_via_cli(home: &Path, stage: &Path) -> (Value, String) {
    let out = run_cli(
        &[
            "update-plan",
            home.to_str().unwrap(),
            stage.to_str().unwrap(),
        ],
        None,
        home,
    );
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    let result = result_of(&out);
    let digest = result["plan_digest"].as_str().unwrap().to_string();
    (result, digest)
}

fn apply_via_cli(home: &Path, stage: &Path, plan: &Value, digest: &str) -> std::process::Output {
    run_cli(
        &[
            "update-apply",
            home.to_str().unwrap(),
            stage.to_str().unwrap(),
            "--plan-digest",
            digest,
            "--permission-reference",
            "fixture-approved",
        ],
        Some(plan),
        home,
    )
}

fn host<'a>(plan: &'a Value, name: &str) -> &'a Value {
    plan["hosts"]
        .as_array()
        .unwrap()
        .iter()
        .find(|h| h["host"] == json!(name))
        .expect("host row")
}

fn tree(root: &Path) -> String {
    tree_digest(root, &[]).unwrap()
}

/// AC-1 + REQ-2 (#2268): the agent-9 repro — a non-RC1 SKILL.md already sits in both
/// projection roots. The plan blocks BOTH hosts (fail-closed): the reason names the
/// unmanaged existing package and the safe alternatives, and the per-file actions are
/// informational only. update-apply exits 0 with per-host `blocked` and a
/// byte-identical projection tree (zero writes).
#[test]
fn unmanaged_existing_package_blocks_plan_and_apply() {
    let root = TempDir::new().unwrap();
    let home = root.path().join("home");
    let stage = root.path().join("stage");
    write_pkg(&stage, &stage_v1());
    fs::create_dir_all(home.join(".claude/skills/moai")).unwrap();
    fs::create_dir_all(home.join(".agents/skills/moai")).unwrap();
    // Non-RC1 previous-generation SKILL.md: no RC1 frontmatter, foreign content.
    fs::write(
        home.join(".claude/skills/moai/SKILL.md"),
        "# moai (old generation skill, written by hand)\nlegacy body\n",
    )
    .unwrap();
    fs::write(
        home.join(".agents/skills/moai/SKILL.md"),
        "legacy: also not an RC1 skill\n",
    )
    .unwrap();
    let claude_root = home.join(".claude");
    let agents_root = home.join(".agents");
    let tree_before = (tree(&claude_root), tree(&agents_root));

    let (plan, digest) = plan_via_cli(&home, &stage);
    for name in ["claude", "codex"] {
        let h = host(&plan, name);
        assert_eq!(h["preflight"], json!("blocked"), "{name}: {h}");
        let reason = h["blocked_reason"].as_str().unwrap();
        assert!(reason.contains("unmanaged existing package"), "{reason}");
        assert!(reason.contains("separate location"), "{reason}");
        assert!(reason.contains("backup-then-replace"), "{reason}");
        // REQ-2: the per-file conflict (SKILL.md) and the pending creates are still
        // visible — informational only.
        let skill = h["files"]
            .as_array()
            .unwrap()
            .iter()
            .find(|f| f["path"] == json!("SKILL.md"))
            .unwrap();
        assert_eq!(skill["action"], json!("conflict"));
        assert_eq!(skill["baseline_sha256"], Value::Null, "unowned");
    }

    // update-apply: exit 0 (the envelope reports the block), zero writes.
    let out = apply_via_cli(&home, &stage, &plan, &digest);
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["result"]["status"], json!("blocked"));
    for h in v["result"]["hosts"].as_array().unwrap() {
        assert_eq!(h["status"], json!("blocked"));
        assert_eq!(h["files"], json!([]), "a blocked host writes nothing");
        assert!(h["blocked_reason"]
            .as_str()
            .unwrap()
            .contains("unmanaged existing package"));
    }

    // AC-1: the projection trees are byte-identical before/after the refused apply.
    assert_eq!(tree(&claude_root), tree_before.0, "claude tree unchanged");
    assert_eq!(tree(&agents_root), tree_before.1, "agents tree unchanged");
}

/// AC-2: a fresh HOME (no pre-existing projection root) keeps the normal managed
/// update path — v1 applies, v2 plans `ok` and applies. No regression from the
/// fail-closed check.
#[test]
fn managed_update_path_is_unchanged() {
    let root = TempDir::new().unwrap();
    let home = root.path().join("home");
    fs::create_dir_all(&home).unwrap();
    let stage1 = root.path().join("stage-v1");
    let stage2 = root.path().join("stage-v2");
    write_pkg(&stage1, &stage_v1());
    write_pkg(&stage2, &stage_v2());

    let (v1_plan, v1_digest) = plan_via_cli(&home, &stage1);
    assert_eq!(
        host(&v1_plan, "claude")["preflight"],
        json!("ok"),
        "a fresh install has no unmanaged content"
    );
    let out = apply_via_cli(&home, &stage1, &v1_plan, &v1_digest);
    let v1 = result_of(&out);
    assert_eq!(v1["status"], json!("applied"));

    // Managed v1 → v2: preflight ok (nothing unmanaged), applied.
    let (v2_plan, v2_digest) = plan_via_cli(&home, &stage2);
    assert_eq!(host(&v2_plan, "claude")["preflight"], json!("ok"));
    let out = apply_via_cli(&home, &stage2, &v2_plan, &v2_digest);
    let v2 = result_of(&out);
    assert_eq!(v2["status"], json!("applied"));
}

/// REQ-3 (#2268): recovery-plan lists the files a FIRST apply added into a
/// pre-existing root (the folder existed — e.g. an empty leftover directory). A
/// fresh install into a non-existent folder lists nothing (control).
#[test]
fn recovery_lists_files_added_into_pre_existing_root() {
    // Control: fresh install into a non-existent folder — no listing.
    let tmp = TempDir::new().unwrap();
    let home = tmp.path().join("fresh-home");
    let stage = tmp.path().join("stage");
    write_pkg(&stage, &stage_v1());
    fs::create_dir_all(&home).unwrap();
    let out = cli_apply(&home, &stage);
    assert!(out.status.success());
    let rec = run_cli(&["recovery-plan", home.to_str().unwrap()], None, &home);
    let plan = result_of(&rec);
    let claude = plan["hosts"]
        .as_array()
        .unwrap()
        .iter()
        .find(|h| h["host"] == json!("claude"))
        .unwrap();
    assert_eq!(
        claude["files"],
        json!([]),
        "a fresh empty-folder install adds nothing listed"
    );

    // The repro-shaped case: the projection folder already existed (empty leftover).
    let tmp2 = TempDir::new().unwrap();
    let home2 = tmp2.path().join("home2");
    let stage2 = tmp2.path().join("stage2");
    write_pkg(&stage2, &stage_v1());
    fs::create_dir_all(&home2).unwrap();
    fs::create_dir_all(home2.join(".claude/skills/moai")).unwrap();
    let out2 = cli_apply(&home2, &stage2);
    assert!(
        out2.status.success(),
        "{}",
        String::from_utf8_lossy(&out2.stderr)
    );
    let rec2 = run_cli(&["recovery-plan", home2.to_str().unwrap()], None, &home2);
    let plan2 = result_of(&rec2);
    let claude2 = plan2["hosts"]
        .as_array()
        .unwrap()
        .iter()
        .find(|h| h["host"] == json!("claude"))
        .unwrap();
    let added: Vec<&str> = claude2["files"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|f| f["class"] == json!("added_into_pre_existing_root"))
        .map(|f| f["path"].as_str().unwrap())
        .collect();
    assert!(
        added.contains(&"SKILL.md"),
        "created files in a pre-existing root are listed: {claude2}"
    );
}

fn cli_apply(home: &Path, stage: &Path) -> std::process::Output {
    let out = run_cli(
        &[
            "update-plan",
            home.to_str().unwrap(),
            stage.to_str().unwrap(),
        ],
        None,
        home,
    );
    let plan = result_of(&out);
    let digest = plan["plan_digest"].as_str().unwrap().to_string();
    run_cli(
        &[
            "update-apply",
            home.to_str().unwrap(),
            stage.to_str().unwrap(),
            "--plan-digest",
            &digest,
            "--permission-reference",
            "fixture-approved",
        ],
        Some(&plan),
        home,
    )
}
