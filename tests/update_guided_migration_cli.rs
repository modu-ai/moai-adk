//! SPEC-20261120-update-guided-migration, slice 2: the update survey. `moai update` (a preview, no
//! confirmation) reads every layer the update can change in one pass and carries the result as
//! `survey` in its JSON: items with a group, a guard, an order and a revert, and one digest.
//! Temporary HOME folders and a temporary project only; the real HOME is never read or written.

#![cfg(unix)]

mod common;
use common::CleanEnv;

use serde_json::{json, Value};
use sha2::{Digest, Sha256};
use std::{
    fs,
    path::{Path, PathBuf},
    process::{Command, Output, Stdio},
};
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;

/// The ids the survey names, in the order the update applies them (REQ-6).
const ORDER: [&str; 28] = [
    "home-retire",
    "home-move",
    "skills",
    "agents",
    "routing",
    "style",
    "command-names",
    "hooks-here",
    "keys",
    "cleanup",
    "switch",
    "models",
    "project-go-settings",
    "project-go-state",
    "project-profile",
    "project-generation",
    "project-claude-settings",
    "project-hooks",
    "project-github",
    "project-codex-hooks",
    "project-codex-config",
    "project-codex-skill",
    "project-omp-hooks",
    "project-omp-config",
    "project-omp-agents",
    "project-agents-local-merge",
    "project-claude-settings-merge",
    "project-plan",
];

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

struct Fixture {
    _tmp: TempDir,
    home: PathBuf,
    exe: PathBuf,
    project: PathBuf,
    path: String,
}

fn write(root: &Path, rel: &str, body: &str) {
    let p = root.join(rel);
    fs::create_dir_all(p.parent().unwrap()).unwrap();
    fs::write(p, body).unwrap();
}

/// A temporary HOME holding the earlier generation (`~/.moai-cli` with this build's version
/// folder, and a Go `~/.moai`) and a git project that holds the earlier generation too (a Go
/// `.moai/` folder, an earlier `.moai-cli/project.yaml`, Go settings and a Go hook).
fn setup() -> Fixture {
    let tmp = TempDir::new().unwrap();
    let root = fs::canonicalize(tmp.path()).unwrap();
    let home = root.join("home");
    let folder = home.join(".moai-cli/versions/a");
    fs::create_dir_all(folder.join("bin")).unwrap();
    let exe = folder.join("bin/moai-cli");
    common::copy_executable(Path::new(EXE), &exe);
    let repo = Path::new(env!("CARGO_MANIFEST_DIR"));
    copy_tree(&repo.join("skills/moai"), &folder.join("skills/moai"));
    copy_tree(&repo.join("agents"), &folder.join("agents"));
    // The Go generation's `~/.moai`.
    write(
        &home,
        ".moai/user-assets.json",
        r#"{"schema_version": 1, "files": {}}"#,
    );
    write(&home, ".moai/state/active-sessions.json", "[]");
    write(&home, ".moai/cache/a.bin", "x");
    write(&home, ".moai/specs/SPEC-1/spec.md", "# user spec\n");
    // The project.
    let project = root.join("project");
    fs::create_dir_all(&project).unwrap();
    common::git::git_ok(&project, &["init", "-q"]);
    let body = "a: 1\n";
    let h = format!("sha256:{:x}", Sha256::digest(body.as_bytes()));
    write(&project, ".moai/config/sections/a.yaml", body);
    write(&project, ".moai-cli/project.yaml", "project_id: demo\n");
    write(
        &project,
        ".moai/manifest.json",
        &json!({"version": "v3.2.1-rc.6", "files": {".moai/config/sections/a.yaml": {
            "provenance": "template_managed", "template_hash": h,
            "deployed_hash": h, "current_hash": h}}})
        .to_string(),
    );
    write(
        &project,
        ".claude/settings.json",
        r#"{"statusLine": {"type": "command", "command": ".moai/status_line.sh"}, "hooks": {"SessionStart": [{"hooks": [{"type": "command", "command": "${CLAUDE_PROJECT_DIR}/.claude/hooks/moai/handle-session-start.sh"}]}]}}"#,
    );
    write(
        &project,
        ".claude/hooks/moai/handle-session-start.sh",
        "#!/bin/sh\nexit 0\n",
    );
    let bin = root.join("bin");
    fs::create_dir_all(&bin).unwrap();
    std::os::unix::fs::symlink(&exe, bin.join("moai")).unwrap();
    Fixture {
        _tmp: tmp,
        home,
        exe,
        project,
        path: format!("{}:/usr/sbin:/usr/bin:/bin:/sbin", bin.display()),
    }
}

fn run(f: &Fixture, cwd: &Path, args: &[&str]) -> Output {
    Command::new(&f.exe)
        .args(args)
        .clean_env()
        .env("PATH", &f.path)
        .env("HOME", &f.home)
        .env("LC_ALL", "C")
        .current_dir(cwd)
        .stdin(Stdio::null())
        .output()
        .unwrap()
}

fn result(out: &Output) -> Value {
    let v: Value = serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
        panic!(
            "not JSON ({e}): {} {}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    });
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn preview(f: &Fixture) -> Value {
    result(&run(f, &f.project, &["update"]))
}

/// Every entry under `dir`: relative path, kind and bytes, except this build's version folder.
fn tree(dir: &Path) -> Vec<(String, String, Vec<u8>)> {
    fn walk(at: &Path, base: &Path, skip: &Path, out: &mut Vec<(String, String, Vec<u8>)>) {
        let mut names: Vec<_> = fs::read_dir(at)
            .unwrap()
            .map(|e| e.unwrap().path())
            .collect();
        names.sort();
        for p in names {
            if p.starts_with(skip) {
                continue;
            }
            let rel = p.strip_prefix(base).unwrap().display().to_string();
            let meta = fs::symlink_metadata(&p).unwrap();
            if meta.file_type().is_symlink() {
                out.push((rel, "link".into(), Vec::new()));
            } else if meta.is_dir() {
                out.push((rel, "dir".into(), Vec::new()));
                walk(&p, base, skip, out);
            } else {
                out.push((rel, "file".into(), fs::read(&p).unwrap()));
            }
        }
    }
    let mut out = Vec::new();
    walk(dir, dir, &dir.join(".moai-cli/versions"), &mut out);
    out
}

fn items(p: &Value) -> Vec<Value> {
    p["survey"]["items"].as_array().cloned().unwrap_or_default()
}

fn item<'a>(all: &'a [Value], id: &str) -> &'a Value {
    all.iter()
        .find(|i| i["id"] == json!(id))
        .unwrap_or_else(|| panic!("the survey has no item {id}: {all:?}"))
}

/// AC-1 (REQ-1, REQ-3): the survey names every layer of the temporary HOME and project in the
/// three groups, each with a revert command or the no-revert statement, in the apply order, and
/// writes nothing.
#[test]
fn survey_reads_every_layer() {
    let f = setup();
    let home_before = tree(&f.home);
    let project_before = tree(&f.project);
    let p = preview(&f);
    assert_eq!(tree(&f.home), home_before, "the survey wrote into the HOME");
    assert_eq!(
        tree(&f.project),
        project_before,
        "the survey wrote into the project"
    );
    let all = items(&p);
    assert!(!all.is_empty(), "no survey in the preview: {p}");
    // Every item has the fields REQ-2 names.
    for i in &all {
        let id = i["id"].as_str().unwrap_or("?");
        assert!(ORDER.contains(&id), "item {id} is not a known id");
        assert!(i["title"].as_str().is_some_and(|t| !t.is_empty()), "{i}");
        assert!(
            matches!(i["layer"].as_str(), Some("home" | "project")),
            "{i}"
        );
        assert!(
            matches!(
                i["group"].as_str(),
                Some("will_change" | "kept" | "needs_pick")
            ),
            "{i}"
        );
        assert!(i["count"]["entries"].is_u64(), "{i}");
        assert!(i["guarded"].is_boolean(), "{i}");
        assert!(i["depends_on"].is_array(), "{i}");
        let kinds = ["command", "after_apply", "none"]
            .iter()
            .filter(|k| i["revert"][**k].as_str().is_some())
            .count();
        assert_eq!(
            kinds, 1,
            "an item carries a revert command, the statement that the apply prints it, or the no-revert statement, one of them: {i}"
        );
    }
    // The apply order (REQ-6): the items come in the order of ORDER and `order` counts up.
    let ids: Vec<&str> = all.iter().map(|i| i["id"].as_str().unwrap()).collect();
    let mut expected: Vec<&str> = ORDER.iter().copied().filter(|x| ids.contains(x)).collect();
    expected.dedup();
    assert_eq!(ids, expected, "the items are not in the apply order");
    let orders: Vec<u64> = all.iter().map(|i| i["order"].as_u64().unwrap()).collect();
    assert!(
        orders.windows(2).all(|w| w[0] < w[1]),
        "order does not count up: {orders:?}"
    );
    // The layers of this fixture.
    for id in [
        "home-retire",
        "home-move",
        "skills",
        "agents",
        "routing",
        "style",
        "command-names",
        "project-go-settings",
        "project-go-state",
    ] {
        item(&all, id);
    }
    // The retire of the Go HOME items is guarded and will change; the HOME move waits for it.
    let retire = item(&all, "home-retire");
    assert_eq!(retire["group"], json!("will_change"), "{retire}");
    assert_eq!(retire["guarded"], json!(true), "{retire}");
    let mv = item(&all, "home-move");
    assert_eq!(mv["guarded"], json!(true), "{mv}");
    assert!(
        mv["depends_on"]
            .as_array()
            .unwrap()
            .contains(&json!("home-retire")),
        "the HOME move depends on the retire: {mv}"
    );
    // The project's state-folder move is guarded; the skills are not.
    assert_eq!(item(&all, "project-go-state")["guarded"], json!(true));
    assert_eq!(item(&all, "skills")["guarded"], json!(false));
    assert_eq!(item(&all, "skills")["group"], json!("will_change"));
    // The groups name the same ids as the items.
    for group in ["will_change", "kept", "needs_pick"] {
        let listed: Vec<&str> = p["survey"]["groups"][group]
            .as_array()
            .unwrap_or_else(|| panic!("no group {group}: {}", p["survey"]))
            .iter()
            .map(|v| v.as_str().unwrap())
            .collect();
        let by_items: Vec<&str> = all
            .iter()
            .filter(|i| i["group"] == json!(group))
            .map(|i| i["id"].as_str().unwrap())
            .collect();
        assert_eq!(listed, by_items, "group {group}");
    }
    assert_eq!(p["survey"]["nothing_to_apply"], json!(false));
}

/// REQ-11: one digest binds the components, it is the same on a second preview of the same
/// state and it changes when one layer changes.
#[test]
fn survey_digest_is_stable_and_binds_the_layers() {
    let f = setup();
    let a = preview(&f);
    let b = preview(&f);
    let digest = |p: &Value| p["survey"]["survey_digest"].as_str().map(str::to_string);
    assert!(digest(&a).is_some(), "no survey digest: {}", a["survey"]);
    assert_eq!(
        digest(&a),
        digest(&b),
        "a second preview gives another digest"
    );
    let parts = a["survey"]["components"].as_array().unwrap();
    assert!(
        parts
            .iter()
            .all(|c| c["id"].is_string() && c["digest"].is_string()),
        "{parts:?}"
    );
    // A Go item of the HOME changes: the digest changes.
    write(&f.home, ".moai/cache/b.bin", "y");
    let c = preview(&f);
    assert_ne!(
        digest(&a),
        digest(&c),
        "a changed layer left the digest as it was"
    );
    // A Go setting of the project changes (the go-settings step has a plan of its own; the
    // go-state step waits for it and has none yet): the digest changes again.
    write(
        &f.project,
        ".claude/settings.json",
        r#"{"statusLine": {"type": "command", "command": ".moai/status_line.sh"}, "plansDirectory": ".moai/plans", "hooks": {"SessionStart": [{"hooks": [{"type": "command", "command": "${CLAUDE_PROJECT_DIR}/.claude/hooks/moai/handle-session-start.sh"}]}]}}"#,
    );
    let d = preview(&f);
    assert_ne!(
        digest(&c),
        digest(&d),
        "a changed project left the digest as it was"
    );
}
