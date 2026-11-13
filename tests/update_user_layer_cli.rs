//! #2638: `moai-cli update` installs the user layer — the skill copies, the base agents and the
//! routing section of the global instruction files — from the real `skills/moai` package placed
//! beside a copy of the binary, inside a temporary HOME. The real HOME is never touched.

#![cfg(unix)]

mod common;
use common::CleanEnv;

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

/// A temporary HOME with `~/.moai-cli/versions/a/bin/moai-cli`, the real package beside it and
/// the agent sources in `agents/` next to `skills/` (#2721).
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
    copy_tree(
        &Path::new(env!("CARGO_MANIFEST_DIR")).join("agents"),
        &folder.join("agents"),
    );
    (tmp, home, exe, pkg)
}

/// #2721 AC-3: an install from a package that still carried the agent sources inside the skill
/// (`references/agents/`) moves to the new layout. The base agents are found by name, so they
/// stay unchanged; the five old skill-tree files are reported `remove_blocked` and kept — no
/// host loads an agent from a skill folder, and the updater never deletes.
#[test]
fn an_install_with_agents_inside_the_skill_moves_to_the_new_layout() {
    let (_t, home, exe, pkg) = setup();
    let old = pkg.join("references/agents");
    copy_tree(&Path::new(env!("CARGO_MANIFEST_DIR")).join("agents"), &old);
    let p = result(&run(&exe, &home, &["update"]));
    let top = p["plan_digest"].as_str().unwrap().to_string();
    let r = result(&run(&exe, &home, &["update", "--confirm", &top]));
    assert_eq!(r["agents_result"]["state"], json!("applied"), "{r}");
    let copied = skill_dir(&home, "claude").join("references/agents/impl.md");
    assert!(copied.is_file());
    // The next release ships the sources only beside the skill.
    fs::remove_dir_all(&old).unwrap();
    let p = result(&run(&exe, &home, &["update"]));
    for host in p["agents"]["hosts"].as_array().unwrap() {
        for f in host["files"].as_array().unwrap() {
            assert_eq!(f["action"], json!("unchanged"), "{f}");
        }
    }
    let claude = p["plan"]["hosts"]
        .as_array()
        .unwrap()
        .iter()
        .find(|h| h["host"] == json!("claude"))
        .unwrap();
    let stale: Vec<&str> = claude["files"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|f| {
            f["path"]
                .as_str()
                .unwrap()
                .starts_with("references/agents/")
        })
        .map(|f| f["action"].as_str().unwrap())
        .collect();
    assert_eq!(stale, ["remove_blocked"; 5], "{p}");
    assert!(copied.is_file(), "a preview deletes nothing");
}

/// #2721: a release folder without `agents/` (or with a symlink there) is refused with the
/// expected path; nothing is written.
#[test]
fn a_release_without_a_real_agent_folder_is_refused() {
    let (_t, home, exe, pkg) = setup();
    let agents = pkg.parent().unwrap().parent().unwrap().join("agents");
    fs::remove_dir_all(&agents).unwrap();
    let before = state(&home);
    let out = run(&exe, &home, &["update"]);
    assert!(!out.status.success());
    let text = String::from_utf8_lossy(&out.stdout);
    assert!(text.contains("no agent sources installed"), "{text}");
    std::os::unix::fs::symlink(
        Path::new(env!("CARGO_MANIFEST_DIR")).join("agents"),
        &agents,
    )
    .unwrap();
    let out = run(&exe, &home, &["update"]);
    assert!(!out.status.success());
    // Refused where the release folder is resolved, before any source is read.
    let text = String::from_utf8_lossy(&out.stdout);
    assert!(
        text.contains("installed agent source folder is a symlink"),
        "{text}"
    );
    // A file in its place is refused the same way.
    fs::remove_file(&agents).unwrap();
    fs::write(&agents, "x").unwrap();
    let out = run(&exe, &home, &["update"]);
    let text = String::from_utf8_lossy(&out.stdout);
    assert!(
        text.contains("the installed agent source path is not a folder"),
        "{text}"
    );
    assert_eq!(state(&home), before);
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
    // Each part counts on its own: with only the section (then only the agents) reverted, the
    // preview still has something to apply.
    let h = home.to_str().unwrap();
    for (op, part) in [
        ("home-routing-revert", "routing_result"),
        ("agents-home-revert", "agents_result"),
    ] {
        let receipt = r[part]["receipt"].as_str().unwrap();
        result(&run(
            &exe,
            &home,
            &[
                op,
                h,
                "--receipt",
                receipt,
                "--permission-reference",
                "fixture",
            ],
        ));
        let p = result(&run(&exe, &home, &["update"]));
        assert_eq!(p["nothing_to_apply"], json!(false), "{op}: {p}");
        // Put the part back, so the next case misses only its own part.
        let top = p["plan_digest"].as_str().unwrap();
        result(&run(&exe, &home, &["update", "--confirm", top]));
        let full = result(&run(&exe, &home, &["update"]));
        assert_eq!(full["nothing_to_apply"], json!(true), "{op}: {full}");
    }
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
        .clean_env()
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
