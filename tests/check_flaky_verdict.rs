//! SPEC-20261126-check-affected-trailers-flaky (AC-6, AC-7): the flaky ledger lives in the
//! `local` folder of the project's state folder, and that folder alone must not change what the
//! project is. The state verdict and the update preview read the same before and after a record,
//! and a state move beside such a folder, and its revert, leave the folder in place. Every case
//! builds a temporary project; the real checkout is never read.

#![cfg(unix)]

mod common;

use common::git::git_ok;
use common::CleanEnv;
use moai_cli::check_flaky::{record, TestName};
use moai_cli::state_root::{
    adopt_new, classify, EARLIER_PROJECT_DIR, LOCAL_DIR, MOAI_DIR, RUST_MARKER,
};
use serde_json::{json, Value};
use std::fs;
use std::os::unix::fs::symlink;
use std::path::{Path, PathBuf};
use std::process::Command;
use tempfile::TempDir;

const BIN: &str = crate::common::MOAI_BIN;
const PERMISSION: &str = "test-permission-ref";
const LEDGER: &str = "check-flaky.jsonl";

/// A temporary project folder and a temporary HOME beside it.
struct Project {
    tmp: TempDir,
    root: PathBuf,
}

impl Project {
    fn new() -> Project {
        let tmp = TempDir::new().unwrap();
        let base = fs::canonicalize(tmp.path()).unwrap();
        let root = base.join("project");
        fs::create_dir_all(&root).unwrap();
        fs::create_dir_all(base.join("home")).unwrap();
        Project { tmp, root }
    }

    /// `<root>/<folder>/<rest>` holding `body`.
    fn write(&self, folder: &str, rest: &str, body: &str) {
        let path = self.root.join(folder).join(rest);
        fs::create_dir_all(path.parent().unwrap()).unwrap();
        fs::write(path, body).unwrap();
    }

    fn read(&self, folder: &str, rest: &str) -> String {
        let path = self.root.join(folder).join(rest);
        fs::read_to_string(&path).unwrap_or_else(|e| panic!("{}: {e}", path.display()))
    }

    fn call(&self, args: &[&str]) -> Value {
        let home = self.tmp.path().join("home");
        let out = Command::new(BIN)
            .args(args)
            .clean_env()
            .env("HOME", home)
            .env("PATH", "/usr/bin:/bin")
            .env("LC_ALL", "C")
            .output()
            .unwrap();
        serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
            panic!(
                "{args:?}: {}{}",
                String::from_utf8_lossy(&out.stdout),
                String::from_utf8_lossy(&out.stderr)
            )
        })
    }

    fn arg(&self) -> &str {
        self.root.to_str().unwrap()
    }

    /// The state of the update preview's `go-state` step.
    fn state_step(&self) -> String {
        let v = self.call(&["project-update-plan", self.arg()]);
        v["result"]["phases"]
            .as_array()
            .unwrap_or_else(|| panic!("{v}"))
            .iter()
            .flat_map(|p| p["steps"].as_array().into_iter().flatten())
            .find(|s| s["step"] == "go-state")
            .and_then(|s| s["state"].as_str())
            .unwrap_or_else(|| panic!("no go-state step: {v}"))
            .to_string()
    }

    /// The layout and the planned rows of the state move preview.
    fn planned_move(&self) -> Value {
        let v = self.call(&["go-state-plan", self.arg()]);
        json!([v["result"]["layout"], v["result"]["state_moves"]])
    }

    fn record_one_flaky_name(&self) {
        let out = self.tmp.path().join("check");
        fs::create_dir_all(&out).unwrap();
        let name = TestName {
            binary: "moai-cli::demo_cli".to_string(),
            name: "t_flaky".to_string(),
        };
        record(&self.root, &out, "0123456789ab", &[name]).unwrap();
    }

    fn generation(&self) -> &'static str {
        classify(&self.root).generation
    }
}

/// AC-6 (REQ-9): in the three layouts a record leaves the verdict `none` and the update preview's
/// state step as they were; a project without a state folder can still be adopted. The same holds
/// for a tracked earlier state whose `.gitignore` hides `.moai/`.
#[test]
fn a_flaky_record_leaves_the_state_verdict_and_the_update_preview_as_they_were() {
    let mut problems = Vec::new();
    for layout in ["earlier and kept names", "earlier only", "new"] {
        let p = Project::new();
        git_ok(&p.root, &["init", "-q"]);
        if layout != "new" {
            p.write(EARLIER_PROJECT_DIR, "project.yaml", "schema_version: 2\n");
        }
        if layout == "earlier and kept names" {
            p.write(MOAI_DIR, "project.yaml", "schema_version: 2\n");
        }
        let before = (p.generation(), p.state_step());
        assert_eq!(before.0, "none", "{layout}");
        p.record_one_flaky_name();
        assert!(
            p.root.join(MOAI_DIR).join(LOCAL_DIR).join(LEDGER).is_file(),
            "{layout}: the premise, the ledger is in the local folder"
        );
        let after = (p.generation(), p.state_step());
        if after != before {
            problems.push(format!("{layout}: before {before:?}, after {after:?}"));
        }
        if layout == "new" {
            match adopt_new(&p.root, "test") {
                Ok(true) if p.root.join(MOAI_DIR).join(RUST_MARKER).is_file() => {
                    assert!(p.root.join(MOAI_DIR).join(LOCAL_DIR).join(LEDGER).is_file());
                }
                other => problems.push(format!("{layout}: adopt_new answered {other:?}")),
            }
        }
    }
    // The same three kinds of project with tracked files and a rule that hides `.moai/`. Beside a
    // name the team keeps the move stays blocked, before the record and after it.
    for (layout, allow, beside, step) in [
        ("hidden, earlier only", "", None, "needs_decision"),
        (
            "hidden, earlier and SPECs",
            "!/.moai/specs/\n",
            Some("specs/SPEC-20260101-a/spec.md"),
            "needs_decision",
        ),
        (
            "hidden, earlier and kept names",
            "!/.moai/project.yaml\n",
            Some("project.yaml"),
            "blocked",
        ),
    ] {
        let p = Project::new();
        git_ok(&p.root, &["init", "-q"]);
        fs::write(p.root.join(".gitignore"), format!("/.moai/*\n{allow}")).unwrap();
        p.write(EARLIER_PROJECT_DIR, "project.yaml", "schema_version: 2\n");
        p.write(EARLIER_PROJECT_DIR, "index/a.json", "{}\n");
        if let Some(rest) = beside {
            p.write(MOAI_DIR, rest, "schema_version: 2\n");
        }
        git_ok(&p.root, &["add", "-A"]);
        let who = ["-c", "user.name=t", "-c", "user.email=t@example.invalid"];
        git_ok(
            &p.root,
            &[&who[..], &["commit", "-q", "-m", "one"]].concat(),
        );
        let before = (p.generation(), p.state_step(), p.planned_move());
        assert_eq!((before.0, before.1.as_str()), ("none", step), "{layout}");
        p.record_one_flaky_name();
        let after = (p.generation(), p.state_step(), p.planned_move());
        if after != before {
            problems.push(format!("{layout}: before {before:?}, after {after:?}"));
        }
    }
    assert!(problems.is_empty(), "{}", problems.join("\n"));
}

/// AC-6 (REQ-10): a `local` that is a link or a file, an unlisted name beside it, a Go marker or
/// an open go-state receipt keep the verdict the folder had before this change.
#[test]
fn a_linked_or_plain_file_local_entry_keeps_the_earlier_verdict() {
    // A link to a real folder is not a local folder, beside a kept name and alone.
    for kept in [true, false] {
        let p = Project::new();
        if kept {
            p.write(MOAI_DIR, "project.yaml", "name: p\n");
        }
        fs::create_dir_all(p.root.join(MOAI_DIR)).unwrap();
        let elsewhere = p.root.join("elsewhere");
        fs::create_dir_all(&elsewhere).unwrap();
        symlink(&elsewhere, p.root.join(MOAI_DIR).join(LOCAL_DIR)).unwrap();
        assert_eq!(p.generation(), "unknown", "a linked local, kept {kept}");
    }

    // A file of that name.
    let p = Project::new();
    p.write(MOAI_DIR, LOCAL_DIR, "mine\n");
    assert_eq!(p.generation(), "unknown", "a local file");

    // An unlisted name beside the local folder.
    let p = Project::new();
    p.write(MOAI_DIR, &format!("{LOCAL_DIR}/{LEDGER}"), "{}\n");
    p.write(MOAI_DIR, "notes.txt", "mine\n");
    assert_eq!(p.generation(), "unknown", "an unlisted name");

    // A Go marker: the same verdict with and without the local folder.
    let verdict_with = |local: bool| {
        let p = Project::new();
        p.write(MOAI_DIR, "manifest.json", "{}\n");
        if local {
            p.write(MOAI_DIR, &format!("{LOCAL_DIR}/{LEDGER}"), "{}\n");
        }
        p.generation()
    };
    assert_ne!(verdict_with(true), "none", "a Go marker");
    assert_eq!(verdict_with(true), verdict_with(false), "a Go marker");

    // A go-state apply that did not finish.
    let p = Project::new();
    p.write(MOAI_DIR, &format!("{LOCAL_DIR}/{LEDGER}"), "{}\n");
    p.write(
        MOAI_DIR,
        "backups/20261126T000000Z/go-state-receipt.json",
        "{\"state\":\"pending\"}\n",
    );
    assert_eq!(p.generation(), "interrupted", "an open receipt");

    // An empty folder stays empty.
    let p = Project::new();
    fs::create_dir_all(p.root.join(MOAI_DIR)).unwrap();
    assert_eq!(p.generation(), "empty");
}

/// The state move of `from` in a plan answer.
fn state_move<'a>(plan: &'a Value, from: &str) -> &'a Value {
    plan["result"]["state_moves"]
        .as_array()
        .unwrap_or_else(|| panic!("{plan}"))
        .iter()
        .find(|m| m["from"] == from)
        .unwrap_or_else(|| panic!("{from} not in state_moves: {plan}"))
}

/// The set-aside folders a revert would leave beside the project.
fn set_aside_folders(root: &Path) -> Vec<String> {
    fs::read_dir(root)
        .unwrap()
        .map(|e| e.unwrap().file_name().to_string_lossy().into_owned())
        .filter(|n| n.starts_with(&format!("{MOAI_DIR}.")))
        .collect()
}

/// AC-7 (REQ-11, REQ-12): with a local folder that was there first, the apply moves the earlier
/// state beside it and records it as held before; the revert leaves the state folder and the
/// ledger where they are and brings the earlier state back.
#[test]
fn a_state_move_and_its_revert_leave_a_local_folder_in_place() {
    let ledger_rest = format!("{LOCAL_DIR}/{LEDGER}");
    let ledger_body = "{\"name\":\"moai-cli::demo_cli t_flaky\"}\n";
    for kept in [false, true] {
        let p = Project::new();
        p.write(EARLIER_PROJECT_DIR, "index/a.json", "{\"a\":1}");
        p.write(MOAI_DIR, &ledger_rest, ledger_body);
        if kept {
            p.write(MOAI_DIR, "evals/e.md", "# e\n");
        }

        let plan = p.call(&["go-state-plan", p.arg()]);
        assert_eq!(plan["ok"], json!(true), "kept {kept}: {plan}");
        let from = format!("{EARLIER_PROJECT_DIR}/index");
        assert_eq!(
            state_move(&plan, &from)["action"],
            json!("move"),
            "kept {kept}: {plan}"
        );
        let digest = plan["result"]["plan_digest"].as_str().unwrap().to_string();
        let applied = p.call(&[
            "go-state-apply",
            p.arg(),
            "--plan-digest",
            &digest,
            "--permission-reference",
            PERMISSION,
        ]);
        assert_eq!(applied["ok"], json!(true), "kept {kept}: {applied}");
        assert_eq!(applied["result"]["state"], json!("committed"), "{applied}");
        let held = applied["result"]["kept_before"]
            .as_array()
            .unwrap_or_else(|| panic!("kept {kept}: no kept_before: {applied}"));
        assert!(held.contains(&json!(LOCAL_DIR)), "kept {kept}: {applied}");
        assert_eq!(p.read(MOAI_DIR, "index/a.json"), "{\"a\":1}");
        assert_eq!(p.read(MOAI_DIR, &ledger_rest), ledger_body);

        let backup = applied["result"]["backup"].as_str().unwrap().to_string();
        let back = p.call(&["go-state-revert-plan", p.arg(), "--backup", &backup]);
        assert_eq!(back["ok"], json!(true), "kept {kept}: {back}");
        assert!(
            !back["result"]["items"]
                .as_array()
                .unwrap()
                .iter()
                .any(|i| i["action"] == "set_aside" && i["path"] == MOAI_DIR),
            "kept {kept}: the revert would set the state folder aside: {back}"
        );
        let digest = back["result"]["plan_digest"].as_str().unwrap().to_string();
        let reverted = p.call(&[
            "go-state-revert",
            p.arg(),
            "--backup",
            &backup,
            "--plan-digest",
            &digest,
            "--permission-reference",
            PERMISSION,
        ]);
        assert_eq!(reverted["ok"], json!(true), "kept {kept}: {reverted}");
        assert_eq!(p.read(MOAI_DIR, &ledger_rest), ledger_body, "kept {kept}");
        assert_eq!(p.read(EARLIER_PROJECT_DIR, "index/a.json"), "{\"a\":1}");
        assert!(set_aside_folders(&p.root).is_empty(), "kept {kept}");
        assert_eq!(p.generation(), "none", "kept {kept}");
    }
}
