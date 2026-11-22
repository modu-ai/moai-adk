//! SPEC-20261122-codex-home-update-removed: the home update plans, writes and reverts for the host
//! `claude` only. A home that holds Codex, shared-agents or OMP files keeps them byte for byte and
//! gets one notice per host; an old receipt that names those hosts reverts its Claude entries and
//! skips the rest. Every case runs in a temporary HOME; the real one is never touched.

#![cfg(unix)]

mod common;
use common::CleanEnv;

use moai_cli::{agent_install, home_routing, home_style};
use serde_json::{json, Value};
use std::{
    collections::BTreeSet,
    fs,
    path::{Path, PathBuf},
    process::{Command, Output, Stdio},
};
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;
const PERM: &str = "test-permission";
const NOTICE: &str = "no longer manages";

fn package() -> PathBuf {
    Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai")
}

fn agent_sources() -> PathBuf {
    Path::new(env!("CARGO_MANIFEST_DIR")).join("agents")
}

fn copy_tree(from: &Path, to: &Path) {
    fs::create_dir_all(to).unwrap();
    for entry in fs::read_dir(from).unwrap() {
        let p = entry.unwrap().path();
        let dest = to.join(p.file_name().unwrap());
        if p.is_dir() {
            copy_tree(&p, &dest);
        } else {
            fs::copy(&p, &dest).unwrap();
        }
    }
}

/// A temporary HOME with `~/.moai-cli/versions/a/bin/moai-cli`, the real package beside it and
/// the agent sources next to `skills/`.
fn setup() -> (TempDir, PathBuf, PathBuf) {
    let tmp = TempDir::new().unwrap();
    let root = fs::canonicalize(tmp.path()).unwrap();
    let home = root.join("home");
    let folder = home.join(".moai-cli/versions/a");
    fs::create_dir_all(folder.join("bin")).unwrap();
    let exe = folder.join("bin/moai-cli");
    common::copy_executable(Path::new(EXE), &exe);
    copy_tree(&package(), &folder.join("skills/moai"));
    copy_tree(&agent_sources(), &folder.join("agents"));
    (tmp, home, exe)
}

fn run(exe: &Path, home: &Path, args: &[&str]) -> Output {
    Command::new(exe)
        .args(args)
        .clean_env()
        .env("PATH", "/usr/bin:/bin")
        .env("HOME", home)
        .env("LC_ALL", "C")
        .current_dir(home)
        .stdin(Stdio::null())
        .output()
        .unwrap()
}

fn result(out: &Output) -> Value {
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

/// Every regular file under `root` with its bytes, by path relative to `root`.
fn tree(root: &Path) -> BTreeSet<(String, Vec<u8>)> {
    fn walk(dir: &Path, root: &Path, out: &mut BTreeSet<(String, Vec<u8>)>) {
        let Ok(entries) = fs::read_dir(dir) else {
            return;
        };
        for e in entries {
            let p = e.unwrap().path();
            if p.is_dir() {
                walk(&p, root, out);
            } else {
                let rel = p.strip_prefix(root).unwrap().to_string_lossy().to_string();
                out.insert((rel, fs::read(&p).unwrap()));
            }
        }
    }
    let mut out = BTreeSet::new();
    walk(root, root, &mut out);
    out
}

fn put(home: &Path, rel: &str, text: &str) {
    let path = home.join(rel);
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, text).unwrap();
}

/// What an earlier version left in a home: Codex files, shared-agents skills, OMP files.
fn leave_retired_files(home: &Path) {
    put(home, ".codex/AGENTS.md", "# My Codex notes\n\nkeep this\n");
    put(
        home,
        ".codex/agents/my-own.md",
        "---\nname: my-own\n---\nmine\n",
    );
    put(
        home,
        ".agents/skills/other/SKILL.md",
        "---\nname: other\n---\n",
    );
    put(
        home,
        ".omp/agent/agents/my-own.md",
        "---\nname: my-own\n---\nmine\n",
    );
}

fn retired_trees(home: &Path) -> Vec<BTreeSet<(String, Vec<u8>)>> {
    [".codex", ".agents", ".omp"]
        .iter()
        .map(|d| tree(&home.join(d)))
        .collect()
}

fn notices(layer: &Value) -> Vec<(String, String)> {
    layer["skipped"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|s| s["reason"].as_str().is_some_and(|r| r.contains(NOTICE)))
        .map(|s| {
            (
                s["host"].as_str().unwrap().to_string(),
                s["reason"].as_str().unwrap().to_string(),
            )
        })
        .collect()
}

// ---- the plan names Claude only (REQ-1, REQ-2, REQ-9) --------------------------------------

#[test]
fn the_user_layer_plan_names_claude_only() {
    let (_t, home, exe) = setup();
    let p = result(&run(&exe, &home, &["update"]));
    let hosts = |v: &Value| -> Vec<String> {
        v.as_array()
            .unwrap()
            .iter()
            .map(|h| h["host"].as_str().unwrap().to_string())
            .collect()
    };
    assert_eq!(hosts(&p["plan"]["hosts"]), ["claude"], "{p}");
    assert_eq!(p["plan"]["scope"], json!(["claude"]), "{p}");
    assert_eq!(hosts(&p["agents"]["hosts"]), ["claude"], "{p}");
    assert_eq!(hosts(&p["routing"]["hosts"]), ["claude"], "{p}");
    let style: BTreeSet<&str> = p["style"]["targets"]
        .as_array()
        .unwrap()
        .iter()
        .map(|t| t["host"].as_str().unwrap())
        .collect();
    assert_eq!(style, BTreeSet::from(["claude"]), "{p}");
    assert_eq!(p["style"]["targets"].as_array().unwrap().len(), 2, "{p}");
}

#[test]
fn the_three_plan_commands_name_claude_only() {
    let (_t, home, exe) = setup();
    let pkg = package();
    let agents = agent_sources();
    let home_text = home.to_str().unwrap();
    let pkg_text = pkg.to_str().unwrap();
    let agents_text = agents.to_str().unwrap();
    let a = result(&run(
        &exe,
        &home,
        &["agents-home-plan", home_text, pkg_text, agents_text],
    ));
    let hosts: Vec<&str> = a["hosts"]
        .as_array()
        .unwrap()
        .iter()
        .map(|h| h["host"].as_str().unwrap())
        .collect();
    assert_eq!(hosts, ["claude"], "{a}");
    let r = result(&run(
        &exe,
        &home,
        &["home-routing-plan", home_text, pkg_text],
    ));
    let hosts: Vec<&str> = r["hosts"]
        .as_array()
        .unwrap()
        .iter()
        .map(|h| h["host"].as_str().unwrap())
        .collect();
    assert_eq!(hosts, ["claude"], "{r}");
    let s = result(&run(&exe, &home, &["home-style-plan", home_text, pkg_text]));
    assert!(
        s["targets"]
            .as_array()
            .unwrap()
            .iter()
            .all(|t| t["host"] == json!("claude")),
        "{s}"
    );
}

// ---- nothing is written for the retired hosts (REQ-3) --------------------------------------

#[test]
fn apply_leaves_the_codex_and_omp_folders_as_they_are() {
    let (_t, home, exe) = setup();
    leave_retired_files(&home);
    let before = retired_trees(&home);
    let p = result(&run(&exe, &home, &["update"]));
    assert_eq!(retired_trees(&home), before, "a preview writes nothing");
    let top = p["plan_digest"].as_str().unwrap();
    let r = result(&run(&exe, &home, &["update", "--confirm", top]));
    let copied: Vec<&str> = r["result"]["hosts"]
        .as_array()
        .unwrap()
        .iter()
        .map(|h| h["host"].as_str().unwrap())
        .collect();
    assert_eq!(copied, ["claude"], "only the Claude skill copy is written");
    assert_eq!(
        retired_trees(&home),
        before,
        "an apply writes nothing there"
    );
    assert!(!home.join(".agents/skills/moai").exists());
    // The Claude side did go in.
    assert!(home.join(".claude/skills/moai/SKILL.md").is_file());
    assert!(home.join(".claude/CLAUDE.md").is_file());
}

#[test]
fn a_home_without_retired_folders_gets_none_made() {
    let (_t, home, exe) = setup();
    let p = result(&run(&exe, &home, &["update"]));
    let top = p["plan_digest"].as_str().unwrap();
    result(&run(&exe, &home, &["update", "--confirm", top]));
    for d in [".codex", ".agents", ".omp"] {
        assert!(!home.join(d).exists(), "{d} was created");
    }
}

// ---- the notice (REQ-4, REQ-5, REQ-6) -------------------------------------------------------

#[test]
fn a_home_with_codex_files_shows_one_notice_for_codex() {
    let (_t, home, exe) = setup();
    put(&home, ".codex/AGENTS.md", "mine\n");
    put(&home, ".agents/skills/other/SKILL.md", "x\n");
    let p = result(&run(&exe, &home, &["update"]));
    let found = notices(&p);
    assert_eq!(found.len(), 1, "{p}");
    assert_eq!(found[0].0, "codex");
    assert!(found[0].1.contains("Codex"), "{}", found[0].1);
    assert!(found[0].1.contains("~/.codex"), "{}", found[0].1);
}

#[test]
fn a_home_with_omp_files_shows_one_notice_for_omp() {
    let (_t, home, exe) = setup();
    put(&home, ".omp/agent/agents/mine.md", "x\n");
    let p = result(&run(&exe, &home, &["update"]));
    let found = notices(&p);
    assert_eq!(found.len(), 1, "{p}");
    assert_eq!(found[0].0, "omp");
    assert!(found[0].1.contains("OMP"), "{}", found[0].1);
}

#[test]
fn a_home_without_them_shows_no_notice() {
    let (_t, home, exe) = setup();
    let p = result(&run(&exe, &home, &["update"]));
    assert!(notices(&p).is_empty(), "{p}");
}

#[test]
fn the_apply_result_carries_the_notice() {
    let (_t, home, exe) = setup();
    leave_retired_files(&home);
    let p = result(&run(&exe, &home, &["update"]));
    let top = p["plan_digest"].as_str().unwrap();
    let r = result(&run(&exe, &home, &["update", "--confirm", top]));
    let hosts: Vec<String> = notices(&r).into_iter().map(|(h, _)| h).collect();
    assert_eq!(hosts, ["codex", "omp"], "{r}");
}

#[test]
fn the_notice_does_not_change_the_plan_digest() {
    let (_t, home, exe) = setup();
    let without = result(&run(&exe, &home, &["update"]));
    leave_retired_files(&home);
    let with = result(&run(&exe, &home, &["update"]));
    assert_eq!(with["plan_digest"], without["plan_digest"]);
    assert!(notices(&without).is_empty() && !notices(&with).is_empty());
}

// ---- an old receipt that names the retired hosts (REQ-7) -----------------------------------

/// Adds one entry for `host` to the receipt `id` of the part whose state folder is `part`: a
/// copy of the first entry with the host changed, the way an earlier version wrote it.
fn add_entry(home: &Path, part: &str, id: &str, host: &str, kind: Option<&str>) {
    let path = moai_cli::home_root::path(home)
        .join(part)
        .join("receipts")
        .join(id)
        .join("receipt.json");
    let mut receipt: Value = serde_json::from_slice(&fs::read(&path).unwrap()).unwrap();
    let mut entry = receipt["entries"][0].clone();
    entry["host"] = json!(host);
    if let Some(kind) = kind {
        entry["kind"] = json!(kind);
    }
    receipt["entries"].as_array_mut().unwrap().push(entry);
    fs::write(&path, receipt.to_string()).unwrap();
}

#[test]
fn an_old_receipt_with_codex_entries_reverts_the_claude_ones() {
    let tmp = TempDir::new().unwrap();
    let home = fs::canonicalize(tmp.path()).unwrap();
    put(&home, ".codex/AGENTS.md", "# kept\n");
    put(&home, ".codex/agents/mine.md", "kept\n");
    put(&home, ".omp/agent/agents/mine.md", "kept\n");
    let codex_before = retired_trees(&home);

    // Routing.
    let p = home_routing::plan(&home, &package()).unwrap();
    let id = home_routing::apply(&home, &package(), &p.plan_digest, PERM).unwrap()["receipt"]
        .as_str()
        .unwrap()
        .to_string();
    add_entry(&home, "home-routing", &id, "codex", None);
    let v = home_routing::revert(&home, &id, PERM).unwrap();
    assert_eq!(v["reverted"], json!(["claude"]), "{v}");
    assert_eq!(
        v["skipped"],
        json!([{"host": "codex", "reason": "host_retired"}]),
        "{v}"
    );

    // Style.
    let p = home_style::plan(&home, &package()).unwrap();
    let id = home_style::apply(&home, &package(), &p.plan_digest, PERM).unwrap()["receipt"]
        .as_str()
        .unwrap()
        .to_string();
    add_entry(&home, "home-style", &id, "codex", Some("block"));
    let v = home_style::revert(&home, &id, PERM).unwrap();
    assert_eq!(
        v["skipped"],
        json!([{"target": "codex:block", "reason": "host_retired"}]),
        "{v}"
    );
    let reverted: Vec<String> = serde_json::from_value(v["reverted"].clone()).unwrap();
    assert!(
        !reverted.is_empty() && reverted.iter().all(|t| t.starts_with("claude:")),
        "{v}"
    );

    // Agents.
    let p = agent_install::plan(&home, &package(), &agent_sources()).unwrap();
    let id = agent_install::apply(&home, &package(), &agent_sources(), &p.plan_digest, PERM)
        .unwrap()["receipt"]
        .as_str()
        .unwrap()
        .to_string();
    add_entry(&home, "agents", &id, "codex", None);
    add_entry(&home, "agents", &id, "omp", None);
    let v = agent_install::revert(&home, &id, PERM).unwrap();
    let hosts: BTreeSet<&str> = v["skipped"]
        .as_array()
        .unwrap()
        .iter()
        .map(|s| s["host"].as_str().unwrap())
        .collect();
    assert_eq!(hosts, BTreeSet::from(["codex", "omp"]), "{v}");
    assert!(v["skipped"]
        .as_array()
        .unwrap()
        .iter()
        .all(|s| s["reason"] == "host_retired"));
    let reverted: Vec<String> = serde_json::from_value(v["reverted"].clone()).unwrap();
    assert!(
        !reverted.is_empty() && reverted.iter().all(|t| t.starts_with("claude:")),
        "{v}"
    );

    assert_eq!(
        retired_trees(&home),
        codex_before,
        "their files are left alone"
    );
}
