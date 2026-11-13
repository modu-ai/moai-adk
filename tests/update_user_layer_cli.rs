//! #2638: `moai-cli update` installs the user layer — the skill copies, the base agents and the
//! routing section of the global instruction files — from the real `skills/moai` package placed
//! beside a copy of the binary, inside a temporary HOME. The real HOME is never touched.

#![cfg(unix)]

mod common;

use serde_json::{json, Value};
use std::{
    fs,
    path::{Path, PathBuf},
    process::{Command, Output, Stdio},
};
use tempfile::TempDir;

const EXE: &str = env!("CARGO_BIN_EXE_moai-cli");

fn package() -> PathBuf {
    Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai")
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

/// A temporary HOME with `~/.moai-cli/versions/a/bin/moai-cli` and the real package beside it.
fn setup() -> (TempDir, PathBuf, PathBuf, PathBuf) {
    let tmp = TempDir::new().unwrap();
    let root = fs::canonicalize(tmp.path()).unwrap();
    let home = root.join("home");
    let folder = home.join(".moai-cli/versions/a");
    fs::create_dir_all(folder.join("bin")).unwrap();
    let exe = folder.join("bin/moai-cli");
    common::copy_executable(Path::new(EXE), &exe);
    let pkg = folder.join("skills/moai");
    copy_tree(&package(), &pkg);
    (tmp, home, exe, pkg)
}

fn run(exe: &Path, home: &Path, args: &[&str]) -> Output {
    Command::new(exe)
        .args(args)
        .env_clear()
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

/// Every file under HOME except the version folder, with its bytes.
fn state(home: &Path) -> Vec<(PathBuf, Vec<u8>)> {
    fn walk(dir: &Path, skip: &Path, out: &mut Vec<(PathBuf, Vec<u8>)>) {
        let Ok(entries) = fs::read_dir(dir) else {
            return;
        };
        for e in entries {
            let p = e.unwrap().path();
            if p.starts_with(skip) {
                continue;
            }
            if p.is_dir() {
                walk(&p, skip, out);
            } else {
                out.push((p.clone(), fs::read(&p).unwrap()));
            }
        }
    }
    let mut out = Vec::new();
    walk(home, &home.join(".moai-cli/versions"), &mut out);
    out.sort();
    out
}

fn agents_dir(home: &Path, host: &str) -> PathBuf {
    match host {
        "claude" => home.join(".claude/agents"),
        "codex" => home.join(".codex/agents"),
        _ => home.join(".omp/agent/agents"),
    }
}

fn skill_dir(home: &Path, agent_host: &str) -> PathBuf {
    match agent_host {
        "claude" => home.join(".claude/skills/moai"),
        _ => home.join(".agents/skills/moai"),
    }
}

#[test]
fn update_previews_and_installs_the_whole_user_layer_then_has_nothing_to_apply() {
    let (_t, home, exe, _) = setup();
    let before = state(&home);
    let p = result(&run(&exe, &home, &["update"]));
    assert_eq!(state(&home), before, "a preview writes nothing");
    assert_eq!(p["nothing_to_apply"], json!(false), "{p}");
    let top = p["plan_digest"].as_str().unwrap();
    assert_eq!(
        p["confirm_with"],
        json!(format!("moai-cli update --confirm {top}"))
    );
    assert_eq!(p["agents"]["hosts"].as_array().unwrap().len(), 3, "{p}");
    assert_eq!(p["routing"]["hosts"].as_array().unwrap().len(), 2, "{p}");
    // A part's own digest is not a confirmation of the whole layer.
    for part in ["plan", "agents", "routing"] {
        let d = p[part]["plan_digest"].as_str().unwrap();
        assert_ne!(d, top);
        let out = run(&exe, &home, &["update", "--confirm", d]);
        assert!(!out.status.success(), "{part}");
        assert_eq!(state(&home), before, "a refused confirm writes nothing");
    }
    let r = result(&run(&exe, &home, &["update", "--confirm", top]));
    assert_eq!(r["applied"], json!(true), "{r}");
    assert_eq!(r["agents_result"]["state"], json!("applied"), "{r}");
    assert_eq!(r["routing_result"]["state"], json!("applied"), "{r}");
    assert!(r["revert"]["agents"]
        .as_str()
        .unwrap()
        .starts_with("moai-cli agents-home-revert"));
    assert!(r["revert"]["routing"]
        .as_str()
        .unwrap()
        .starts_with("moai-cli home-routing-revert"));
    // The global instruction files carry the section; `~/.agents/AGENTS.md` is never made.
    assert!(fs::read_to_string(home.join(".claude/CLAUDE.md"))
        .unwrap()
        .contains("(/moai)"));
    assert!(fs::read_to_string(home.join(".codex/AGENTS.md"))
        .unwrap()
        .contains("($moai)"));
    assert!(!home.join(".agents/AGENTS.md").exists());
    // Every `references/…` path an installed agent names exists in the skill copy it reads.
    let mut checked = 0;
    for host in ["claude", "codex", "omp"] {
        let dir = agents_dir(&home, host);
        let files: Vec<PathBuf> = fs::read_dir(&dir)
            .unwrap()
            .map(|e| e.unwrap().path())
            .collect();
        assert_eq!(files.len(), 5, "{host}: {files:?}");
        for f in files {
            let text = fs::read_to_string(&f).unwrap();
            for part in text.split('`').filter(|s| s.starts_with("references/")) {
                assert!(
                    skill_dir(&home, host).join(part).is_file(),
                    "{host} {}: {part}",
                    f.display()
                );
                checked += 1;
            }
        }
    }
    assert!(checked > 0, "no references checked");
    // A second run has nothing left to do.
    let again = result(&run(&exe, &home, &["update"]));
    assert_eq!(again["nothing_to_apply"], json!(true), "{again}");
}

/// Skills already current (installed through `update-plan`/`update-apply`), agents and section
/// missing: `update` still has something to apply and installs only those two parts.
#[test]
fn current_skills_with_missing_agents_still_have_something_to_apply() {
    let (_t, home, exe, pkg) = setup();
    let (h, p) = (home.to_str().unwrap(), pkg.to_str().unwrap());
    let plan = result(&run(&exe, &home, &["update-plan", h, p]));
    let digest = plan["plan_digest"].as_str().unwrap().to_string();
    let mut child = Command::new(&exe)
        .args([
            "update-apply",
            h,
            p,
            "--plan-digest",
            &digest,
            "--permission-reference",
            "fixture-approved",
        ])
        .env_clear()
        .env("PATH", "/usr/bin:/bin")
        .env("HOME", &home)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    use std::io::Write as _;
    child
        .stdin
        .take()
        .unwrap()
        .write_all(plan.to_string().as_bytes())
        .unwrap();
    let out = child.wait_with_output().unwrap();
    assert_eq!(result(&out)["status"], json!("applied"));
    let p = result(&run(&exe, &home, &["update"]));
    for h in p["plan"]["hosts"].as_array().unwrap() {
        assert!(
            h["files"]
                .as_array()
                .unwrap()
                .iter()
                .all(|f| f["action"] == json!("unchanged")),
            "{h}"
        );
    }
    assert_eq!(p["nothing_to_apply"], json!(false), "{p}");
    let top = p["plan_digest"].as_str().unwrap().to_string();
    let r = result(&run(&exe, &home, &["update", "--confirm", &top]));
    assert_eq!(r["applied"], json!(true), "{r}");
    assert_eq!(r["agents_result"]["state"], json!("applied"), "{r}");
    assert!(home.join(".claude/CLAUDE.md").is_file());
}

#[test]
fn a_blocked_skill_copy_gets_no_agents_and_no_section() {
    let (_t, home, exe, _) = setup();
    // A skill this tool did not install blocks the Codex copy (`~/.agents/skills/moai`).
    let foreign = home.join(".agents/skills/moai");
    fs::create_dir_all(&foreign).unwrap();
    fs::write(
        foreign.join("SKILL.md"),
        "---\nname: moai\ndescription: someone else's\n---\n",
    )
    .unwrap();
    let p = result(&run(&exe, &home, &["update"]));
    let hosts = |part: &str| -> Vec<String> {
        p[part]["hosts"]
            .as_array()
            .unwrap()
            .iter()
            .map(|h| h["host"].as_str().unwrap().to_string())
            .collect()
    };
    assert_eq!(hosts("agents"), ["claude"], "{p}");
    assert_eq!(hosts("routing"), ["claude"], "{p}");
    let skipped: Vec<String> = p["skipped"]
        .as_array()
        .unwrap()
        .iter()
        .map(|s| {
            format!(
                "{}:{}",
                s["part"].as_str().unwrap(),
                s["host"].as_str().unwrap()
            )
        })
        .collect();
    assert_eq!(
        skipped,
        ["agents:codex", "agents:omp", "routing:codex"],
        "{p}"
    );
    let top = p["plan_digest"].as_str().unwrap().to_string();
    let r = result(&run(&exe, &home, &["update", "--confirm", &top]));
    assert_eq!(r["applied"], json!(true), "{r}");
    assert!(agents_dir(&home, "claude").join("moai-impl.md").is_file());
    assert!(!agents_dir(&home, "codex").exists());
    assert!(!agents_dir(&home, "omp").exists());
    assert!(home.join(".claude/CLAUDE.md").is_file());
    assert!(!home.join(".codex/AGENTS.md").exists());
}

#[test]
fn a_package_changed_after_the_preview_is_refused_with_nothing_written() {
    let (_t, home, exe, pkg) = setup();
    let p = result(&run(&exe, &home, &["update"]));
    let top = p["plan_digest"].as_str().unwrap().to_string();
    let src = pkg.join("references/home/routing.md");
    let text = fs::read_to_string(&src).unwrap();
    fs::write(&src, text.replace("before asking", "before you ask")).unwrap();
    let before = state(&home);
    let out = run(&exe, &home, &["update", "--confirm", &top]);
    assert!(!out.status.success());
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert!(
        v["error"]["message"].as_str().unwrap().contains("stale"),
        "{v}"
    );
    assert_eq!(state(&home), before, "nothing was written");
}
