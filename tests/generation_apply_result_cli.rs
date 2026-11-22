//! #2936 G3-2 (F18): the `generation-apply` result says what it did. A file that was already
//! there and got a section appended is `files_merged`, not `files_created`; the manifest is
//! listed; the agents engine's own receipt is named; and what was not checked and what to do
//! next are said.
mod common;
use serde_json::{json, Value};
use std::{
    fs,
    io::Write,
    process::{Command, Stdio},
};
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;
const MANIFEST: &str = ".moai-cli/manifests/generation.json";

fn project() -> TempDir {
    let root = TempDir::new().unwrap();
    fs::create_dir_all(root.path().join("skills/moai/workflows")).unwrap();
    fs::create_dir_all(root.path().join(".moai-cli")).unwrap();
    let profile = json!({"schema_version":2,"project_id":"demo","purpose":"Build a small editor","tools":["Rust"],"development_methods":["TDD"],"hosts":["claude","codex"],"interview_revision":"rc1-fixture-1"});
    let mut bytes = serde_json::to_vec_pretty(&profile).unwrap();
    bytes.push(b'\n');
    fs::write(root.path().join(".moai-cli/project.yaml"), &bytes).unwrap();
    root
}

fn cli(args: &[&str], input: &Value) -> Value {
    let mut child = Command::new(EXE)
        .args(args)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    let mut pipe = child.stdin.take().unwrap();
    let _ = pipe.write_all(input.to_string().as_bytes());
    drop(pipe);
    let out = child.wait_with_output().unwrap();
    serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|_| panic!("no envelope: {}", String::from_utf8_lossy(&out.stderr)))
}

/// Plans and applies the generation; returns the apply `result`.
fn generate(root: &TempDir) -> Value {
    let request = json!({
        "protocol_version": 1, "language": "en",
        "workflow_reference": "skills/moai/workflows/", "local_file": "tracked",
        "workflow_id": "review-flow", "native_roles": [], "skill_projections": []
    });
    let p = root.path().to_str().unwrap();
    let plan = cli(&["generation-plan", p], &request);
    assert_eq!(plan["ok"], json!(true), "{plan}");
    let plan = plan["result"].clone();
    let applied = cli(
        &[
            "generation-apply",
            p,
            "--plan-digest",
            plan["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "fixture-approved",
        ],
        &plan,
    );
    assert_eq!(applied["ok"], json!(true), "{applied}");
    applied["result"].clone()
}

fn names(v: &Value) -> Vec<String> {
    v.as_array()
        .unwrap()
        .iter()
        .map(|s| s.as_str().unwrap().to_string())
        .collect()
}

#[test]
fn a_new_project_lists_everything_as_created_with_its_manifest_and_receipt() {
    let root = project();
    let r = generate(&root);
    let created = names(&r["files_created"]);
    for path in ["AGENTS.md", MANIFEST] {
        assert!(created.iter().any(|c| c == path), "{path} in {created:?}");
    }
    assert!(names(&r["files_merged"]).is_empty(), "nothing existed: {r}");

    // The agents engine's receipt is named and is a receipt on disk.
    let id = r["agents_receipt"].as_str().expect("agents_receipt");
    assert_ne!(Some(id), r["receipt"].as_str(), "{r}");
    assert!(
        root.path()
            .join(format!(".moai-cli/receipts/{id}.json"))
            .exists(),
        "{id}"
    );

    // What was not checked and what to do next are said, and the flags stay false.
    assert_eq!(r["host_loaded"], json!(false));
    assert_eq!(r["execution_verified"], json!(false));
    let not_checked = names(&r["not_checked"]).join(" | ");
    assert!(
        not_checked.contains("host_loaded") && not_checked.contains("execution_verified"),
        "{not_checked}"
    );
    let next = names(&r["next_actions"]).join(" | ");
    assert!(
        next.contains("new host session") && next.contains("project-update-plan"),
        "{next}"
    );
}

#[test]
fn a_file_that_was_already_there_is_merged_not_created() {
    let root = project();
    let mine = "# My project notes\n\nKeep this line.\n";
    fs::write(root.path().join("AGENTS.md"), mine).unwrap();
    fs::write(root.path().join(".gitignore"), "target/\n").unwrap();
    let r = generate(&root);

    let created = names(&r["files_created"]);
    let merged = names(&r["files_merged"]);
    assert!(merged.iter().any(|m| m == "AGENTS.md"), "merged {merged:?}");
    // No file that was there before is called created (the request may not touch `.gitignore`).
    for path in ["AGENTS.md", ".gitignore"] {
        assert!(
            !created.iter().any(|c| c == path),
            "{path} in created {created:?}"
        );
    }
    // The manifest did not exist, so it is a creation, and the user's text is still there.
    assert!(created.iter().any(|c| c == MANIFEST), "{created:?}");
    let agents = fs::read_to_string(root.path().join("AGENTS.md")).unwrap();
    assert!(agents.contains("Keep this line."), "{agents}");
}

/// SPEC-20261121-remove-codex-project-layer REQ-8 (AC-8): a project that lists claude, codex and
/// omp, and a request that names a Codex role, a Codex projection target and the default project
/// skill hosts, gets the Claude files only. Nothing exists under `.codex`, `.agents` or `.omp`,
/// and the manifest names no other host's projection or owned entry.
#[test]
fn no_codex_projection_is_written() {
    let root = project();
    let profile = json!({"schema_version":2,"project_id":"demo","purpose":"Build a small editor","tools":["Rust"],"development_methods":["TDD"],"hosts":["claude","codex","omp"],"interview_revision":"rc1-fixture-1"});
    let mut bytes = serde_json::to_vec_pretty(&profile).unwrap();
    bytes.push(b'\n');
    fs::write(root.path().join(".moai-cli/project.yaml"), &bytes).unwrap();
    let sources = json!({"schema_version": 1, "sources": [
        {"source_url": "fixtures/skill-b", "license": {"status": "reviewed", "spdx": "MIT"},
         "resolved_revision": null}]});
    fs::create_dir_all(root.path().join(".moai-cli/manifests")).unwrap();
    fs::write(
        root.path().join(".moai-cli/manifests/sources.json"),
        serde_json::to_vec_pretty(&sources).unwrap(),
    )
    .unwrap();
    let request = json!({
        "protocol_version": 1, "language": "en",
        "workflow_reference": "skills/moai/workflows/", "local_file": "none",
        "workflow_id": "review-flow",
        "native_roles": [
            {"role_id": "prj-fixer", "host": "claude", "purpose": "Fix lint failures"},
            {"role_id": "prj-watcher", "host": "codex", "purpose": "Review diffs only"}
        ],
        "skill_projections": [
            {"id": "ref-b", "source_url": "fixtures/skill-b", "host_targets": ["claude", "codex"]}
        ],
        "project_skills": [{"id": "prj-checks", "purpose": "How this project is checked"}]
    });
    let p = root.path().to_str().unwrap();
    let plan = cli(&["generation-plan", p], &request);
    assert_eq!(plan["ok"], json!(true), "{plan}");
    let plan = plan["result"].clone();
    let applied = cli(
        &[
            "generation-apply",
            p,
            "--plan-digest",
            plan["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "fixture-approved",
        ],
        &plan,
    );
    assert_eq!(applied["ok"], json!(true), "{applied}");

    for folder in [".codex", ".agents", ".omp"] {
        assert!(
            !root.path().join(folder).exists(),
            "{folder} must not exist after a generation"
        );
    }
    for claude in [
        ".claude/agents/prj-fixer.md",
        ".claude/skills/ref-b/SKILL.md",
        ".claude/skills/prj-checks/SKILL.md",
    ] {
        assert!(root.path().join(claude).exists(), "{claude}");
    }
    let manifest: Value =
        serde_json::from_slice(&fs::read(root.path().join(MANIFEST)).unwrap()).unwrap();
    assert_eq!(manifest["projections"], json!(["claude"]), "{manifest}");
    for entry in manifest["owned_files"].as_array().unwrap() {
        let path = entry["path"].as_str().unwrap_or("");
        let tag = entry["section_id"].as_str().unwrap_or("");
        assert!(
            !path.starts_with(".codex/")
                && !path.starts_with(".agents/")
                && !path.starts_with(".omp/")
                && !tag.contains(":codex:")
                && !tag.contains(":omp:"),
            "{entry}"
        );
    }
}
