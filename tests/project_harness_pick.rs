//! SPEC-20261120-project-harness-research revision 2: a request that carries the pick
//! `harness-only` plans and applies the requested project agents and skills and the manifest
//! alone, so a hand-edited `CLAUDE.md`, `REVIEW.md` or `.gitignore` never blocks the change of
//! one agent. Every case runs the built binary in a temporary project with a temporary HOME.

#![cfg(unix)]

mod common;

use common::CleanEnv;
use serde_json::{json, Value};
use std::fs;
use std::io::Write;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;
const PICK: &str = "harness-only";
const ROLE: &str = ".claude/agents/moai/prj-reviewer.md";
const CHECKS: &str = ".claude/skills/moai-prj-checks/SKILL.md";
const OTHER: &str = ".claude/skills/moai-prj-other/SKILL.md";

struct Fx {
    _tmp: TempDir,
    home: PathBuf,
    root: PathBuf,
}

/// The project's state folder, named by the resolver the product uses.
fn state(root: &Path) -> PathBuf {
    root.join(moai_cli::state_root::dir(root))
}

fn decision(id: &str, value: &str) -> Value {
    json!({"id": id, "topic": id, "value": value, "source": "user", "scope": "project"})
}

fn write_profile(root: &Path, revision: &str, purpose: &str, tools: &[&str], review_blocks: &str) {
    let profile = json!({"schema_version": 2, "project_id": "demo", "purpose": purpose,
    "tools": tools, "development_methods": ["TDD"], "hosts": ["claude"],
    "interview_revision": revision, "decisions": [
        decision("test_command", "the project test command"),
        decision("review_blocks", review_blocks),
        decision("protected", "the billing module"),
    ]});
    let mut bytes = serde_json::to_vec_pretty(&profile).unwrap();
    bytes.push(b'\n');
    fs::write(state(root).join("project.yaml"), bytes).unwrap();
}

fn fixture() -> Fx {
    let tmp = TempDir::new().unwrap();
    let base = fs::canonicalize(tmp.path()).unwrap();
    let (home, root) = (base.join("home"), base.join("project"));
    fs::create_dir_all(&home).unwrap();
    fs::create_dir_all(root.join("skills/moai/workflows")).unwrap();
    fs::create_dir_all(state(&root)).unwrap();
    write_profile(
        &root,
        "r1",
        "Build a small editor",
        &["Rust"],
        "a change without a test",
    );
    Fx {
        _tmp: tmp,
        home,
        root,
    }
}

impl Fx {
    fn run(&self, op: &str, extra: &[&str], input: &Value) -> Value {
        let mut child = Command::new(EXE)
            .arg(op)
            .arg(&self.root)
            .args(extra)
            .clean_env()
            .env("HOME", &self.home)
            .env("PATH", "/usr/bin:/bin")
            .env("LC_ALL", "C")
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

    fn plan(&self, request: &Value) -> Value {
        let v = self.run("generation-plan", &[], request);
        assert_eq!(v["ok"], json!(true), "{v}");
        v["result"].clone()
    }

    /// The envelope of `generation-apply` for a plan.
    fn apply(&self, plan: &Value) -> Value {
        let digest = plan["plan_digest"].as_str().unwrap();
        self.run(
            "generation-apply",
            &[
                "--plan-digest",
                digest,
                "--permission-reference",
                "fixture-approved",
            ],
            plan,
        )
    }

    fn read(&self, rel: &str) -> String {
        fs::read_to_string(self.root.join(rel)).unwrap()
    }

    fn manifest(&self) -> Value {
        let bytes = fs::read(state(&self.root).join("manifests/generation.json")).unwrap();
        serde_json::from_slice(&bytes).unwrap()
    }
}

fn request(reviewer_purpose: &str, skills: &[&str], picks: &[&str]) -> Value {
    let skills: Vec<Value> = skills
        .iter()
        .map(|id| json!({"id": id, "purpose": format!("What {id} is for")}))
        .collect();
    let mut r = json!({"protocol_version": 1, "language": "en",
        "workflow_reference": "skills/moai/workflows", "local_file": "none",
        "workflow_id": "review-flow", "skill_projections": [],
        "native_roles": [{"role_id": "prj-reviewer", "host": "claude", "purpose": reviewer_purpose}],
        "project_skills": skills});
    if !picks.is_empty() {
        r["picks"] = json!(picks);
    }
    r
}

fn actions(plan: &Value) -> Vec<(String, String)> {
    plan["files"]
        .as_array()
        .unwrap()
        .iter()
        .map(|f| {
            (
                f["path"].as_str().unwrap().to_string(),
                f["action"].as_str().unwrap().to_string(),
            )
        })
        .collect()
}

/// Puts one line of the user's own right after the begin marker of a managed block.
fn hand_edit(fx: &Fx, rel: &str, marker: &str) {
    let text = fx.read(rel);
    let at = text
        .find(marker)
        .unwrap_or_else(|| panic!("{rel} has no marker {marker}"));
    let line_end = at + text[at..].find('\n').unwrap() + 1;
    let edited = format!("{}HAND EDIT\n{}", &text[..line_end], &text[line_end..]);
    fs::write(fx.root.join(rel), edited).unwrap();
}

const TRACKED: [&str; 3] = ["CLAUDE.md", "REVIEW.md", ".gitignore"];

/// A project whose whole harness was generated once, whose three tracked files the user then
/// edited inside their managed blocks, and whose profile has since moved on. A block conflicts
/// only when both sides moved after the baseline, so the profile moves what two of the blocks
/// render: CLAUDE.md renders the tools, and REVIEW.md the review rules. The `.gitignore` block
/// renders from no profile field (each one was moved on its own and left it `unchanged`), so it
/// stands for a block an earlier build of the tool generated, whose recorded baseline is not
/// what this build renders — the state of a project generated before a tool update.
fn generated_then_edited() -> Fx {
    let fx = fixture();
    let first = fx.plan(&request(
        "Reviews a change",
        &["prj-checks", "prj-other"],
        &[],
    ));
    let applied = fx.apply(&first);
    assert_eq!(applied["ok"], json!(true), "{applied}");
    for rel in TRACKED {
        assert!(fx.root.join(rel).exists(), "{rel} was not generated");
    }
    earlier_gitignore_baseline(&fx);
    hand_edit(
        &fx,
        "CLAUDE.md",
        "<!-- moai:managed:begin id=moai-workflow -->",
    );
    hand_edit(
        &fx,
        "REVIEW.md",
        "<!-- moai:managed:begin id=moai-review -->",
    );
    hand_edit(&fx, ".gitignore", "# moai:managed:begin id=moai-local");
    // Both sides move: the files were edited by hand and what they would render changed.
    write_profile(
        &fx.root,
        "r2",
        "Build a large editor",
        &["Rust", "Python"],
        "a change without a test or a review",
    );
    fx
}

/// Makes the recorded baseline of the `.gitignore` block one this build no longer renders, the
/// way a manifest written by an earlier build of the tool holds it.
fn earlier_gitignore_baseline(fx: &Fx) {
    let path = state(&fx.root).join("manifests/generation.json");
    let mut manifest: Value = serde_json::from_slice(&fs::read(&path).unwrap()).unwrap();
    let entry = manifest["owned_files"]
        .as_array_mut()
        .expect("owned_files")
        .iter_mut()
        .find(|f| f["path"] == ".gitignore")
        .expect("the manifest records .gitignore");
    entry["baseline_sha256"] = json!(moai_cli::digest(
        b"# moai:managed:begin id=moai-local\nwhat an earlier build rendered here\n# moai:managed:end id=moai-local\n"
    ));
    fs::write(&path, serde_json::to_vec_pretty(&manifest).unwrap()).unwrap();
}

#[test]
fn without_the_pick_a_hand_edited_tracked_file_blocks_the_change_of_one_agent() {
    let fx = generated_then_edited();
    let plan = fx.plan(&request("Reviews a change twice", &["prj-checks"], &[]));
    // Every tracked file the user edited by hand while what it would render moved is a
    // conflict, and nothing else is: the paths are named, not only the fact of one conflict.
    let mut conflicts: Vec<String> = actions(&plan)
        .into_iter()
        .filter(|(_, action)| action == "conflict")
        .map(|(path, _)| path)
        .collect();
    conflicts.sort();
    let mut expected: Vec<String> = TRACKED.iter().map(|rel| rel.to_string()).collect();
    expected.sort();
    assert_eq!(conflicts, expected, "the conflicts without the pick");
    assert_ne!(
        fx.apply(&plan)["ok"],
        json!(true),
        "a conflict stops the apply"
    );
}

#[test]
fn the_pick_plans_only_the_requested_agents_and_skills_and_the_manifest() {
    let fx = generated_then_edited();
    let plan = fx.plan(&request("Reviews a change twice", &["prj-checks"], &[PICK]));
    let listed = actions(&plan);
    for (path, action) in &listed {
        let allowed = path == ROLE || path == CHECKS || path.ends_with("manifests/generation.json");
        assert!(
            allowed,
            "{path} ({action}) is not part of the harness: {listed:?}"
        );
        assert_ne!(action, "conflict", "{path}: {listed:?}");
    }
    assert!(
        listed.contains(&(ROLE.to_string(), "merge".to_string())),
        "the renamed agent is planned: {listed:?}"
    );
    assert!(
        !listed.iter().any(|(p, _)| p == OTHER),
        "a skill the request does not name is not planned for removal: {listed:?}"
    );
}

#[test]
fn the_pick_applies_the_agent_and_leaves_every_other_file_and_the_linkage_alone() {
    let fx = generated_then_edited();
    let kept: Vec<(String, String)> = TRACKED
        .iter()
        .map(|rel| (rel.to_string(), fx.read(rel)))
        .chain([(
            format!(
                "{}/workflows/review-flow/SKILL.md",
                moai_cli::state_root::dir(&fx.root)
            ),
            fx.read(&format!(
                "{}/workflows/review-flow/SKILL.md",
                moai_cli::state_root::dir(&fx.root)
            )),
        )])
        .collect();
    let before = fx.manifest();
    let other_before = fx.read(OTHER);

    let plan = fx.plan(&request("Reviews a change twice", &["prj-checks"], &[PICK]));
    let applied = fx.apply(&plan);
    assert_eq!(applied["ok"], json!(true), "{applied}");

    assert!(fx.read(ROLE).contains("Reviews a change twice"));
    for (rel, text) in &kept {
        assert_eq!(&fx.read(rel), text, "{rel} must stay as the user left it");
    }
    assert_eq!(fx.read(OTHER), other_before, "a skill nobody named stays");
    let after = fx.manifest();
    for key in ["workflow_id", "interview_revision", "profile_sha256"] {
        assert_eq!(
            after[key], before[key],
            "the manifest key {key} must not move"
        );
    }
    assert!(
        after["owned_files"]
            .as_array()
            .unwrap()
            .iter()
            .any(|f| f["path"] == json!(OTHER)),
        "the skill nobody named is still recorded: {after}"
    );
}

#[test]
fn without_the_pick_a_skill_the_request_leaves_out_is_still_planned_for_removal() {
    let fx = fixture();
    let first = fx.plan(&request(
        "Reviews a change",
        &["prj-checks", "prj-other"],
        &[],
    ));
    assert_eq!(fx.apply(&first)["ok"], json!(true));
    let plan = fx.plan(&request("Reviews a change", &["prj-checks"], &[]));
    assert!(
        actions(&plan).contains(&(OTHER.to_string(), "remove".to_string())),
        "{:?}",
        actions(&plan)
    );
}

/// A project that was never generated has no earlier record to go back to: under the pick its
/// manifest then holds the digest of the profile the apply ran with, and no workflow id and no
/// interview revision (the comment of `record_generation` in `src/generation.rs` says the same).
#[test]
fn a_never_generated_project_keeps_the_current_profile_digest_under_the_pick() {
    let fx = fixture();
    assert!(
        !state(&fx.root).join("manifests/generation.json").exists(),
        "the fixture must start without a manifest"
    );
    let plan = fx.plan(&request("Reviews a change", &["prj-checks"], &[PICK]));
    assert_eq!(fx.apply(&plan)["ok"], json!(true));
    let after = fx.manifest();
    let profile = fs::read(state(&fx.root).join("project.yaml")).unwrap();
    assert_eq!(
        after["profile_sha256"],
        json!(moai_cli::digest(&profile)),
        "{after}"
    );
    assert!(after["interview_revision"].is_null(), "{after}");
    assert!(after["workflow_id"].is_null(), "{after}");
}
