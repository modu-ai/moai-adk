//! #2522: one-line `moai-cli update` — preview, confirm, apply. The package it applies is
//! the one installed with the binary, `<version folder>/skills/moai` next to
//! `<version folder>/bin/moai-cli`. Every case runs copies of the built binary inside a
//! temporary HOME; the real HOME and the installed binary are never touched.

#![cfg(unix)]

mod common;

use serde_json::{json, Value};
use std::{
    fs,
    io::Write as _,
    path::{Path, PathBuf},
    process::{Command, Output, Stdio},
    thread::sleep,
    time::Duration,
};
use tempfile::TempDir;

const EXE: &str = env!("CARGO_BIN_EXE_moai-cli");
const HOSTS: [&str; 2] = ["claude", "codex"];

fn skill_md(body: &str) -> String {
    format!("---\nname: moai\ndescription: fixture package\n---\n# MoAI\n{body}\n")
}

fn v1() -> Vec<(&'static str, String)> {
    vec![
        ("SKILL.md", skill_md("body v1")),
        ("workflows/update.md", "# UPDATE\nv1\n".into()),
        ("references/routes.json", "{\"routes\":1}\n".into()),
    ]
}

fn v2() -> Vec<(&'static str, String)> {
    vec![
        ("SKILL.md", skill_md("body v2")),
        ("workflows/update.md", "# UPDATE\nv2\n".into()),
        ("references/routes.json", "{\"routes\":1}\n".into()),
        ("workflows/new.md", "# NEW\nv2\n".into()),
    ]
}

struct Fx {
    _tmp: TempDir,
    root: PathBuf,
    home: PathBuf,
}

fn fx() -> Fx {
    let tmp = TempDir::new().unwrap();
    let root = fs::canonicalize(tmp.path()).unwrap();
    let home = root.join("home");
    fs::create_dir_all(&home).unwrap();
    Fx {
        _tmp: tmp,
        root,
        home,
    }
}

/// `~/.moai-cli/versions/<version>/bin/moai-cli` with its package beside it.
fn install_version(home: &Path, version: &str, pkg: &[(&str, String)]) -> PathBuf {
    let folder = home.join(".moai-cli/versions").join(version);
    let bin = folder.join("bin");
    fs::create_dir_all(&bin).unwrap();
    let exe = bin.join("moai-cli");
    common::copy_executable(Path::new(EXE), &exe);
    for (rel, body) in pkg {
        let p = folder.join("skills/moai").join(rel);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        fs::write(p, body).unwrap();
    }
    exe
}

fn run(exe: &Path, home: &Path, args: &[&str]) -> Output {
    let mut child = Command::new(exe)
        .args(args)
        .env_clear()
        .env("PATH", "/usr/bin:/bin")
        .env("HOME", home)
        .env("LC_ALL", "C")
        .current_dir(home)
        .stdin(Stdio::null())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    let _ = child.stdin.take();
    child.wait_with_output().unwrap()
}

fn envelope(out: &Output) -> Value {
    serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
        panic!(
            "envelope JSON on stdout: {}{}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    })
}

fn ok(out: &Output) -> Value {
    let v = envelope(out);
    assert_eq!(v["ok"], json!(true), "{v}");
    assert!(out.status.success(), "{v}");
    v["result"].clone()
}

/// Every file under the home except the copied binaries and packages, with its bytes.
fn state(home: &Path) -> Vec<(PathBuf, Vec<u8>)> {
    let mut out = Vec::new();
    let mut stack = vec![home.to_path_buf()];
    while let Some(d) = stack.pop() {
        for e in fs::read_dir(&d).unwrap() {
            let p = e.unwrap().path();
            if p.ends_with(".moai-cli/versions") {
                continue;
            }
            if p.is_dir() {
                stack.push(p);
            } else {
                out.push((p.clone(), fs::read(&p).unwrap()));
            }
        }
    }
    out.sort();
    out
}

fn preview(exe: &Path, home: &Path) -> Value {
    let before = state(home);
    let r = ok(&run(exe, home, &["update"]));
    assert_eq!(r["mode"], json!("preview"), "{r}");
    assert_eq!(r["applied"], json!(false), "{r}");
    assert_eq!(state(home), before, "a preview writes nothing");
    r
}

fn confirm(exe: &Path, home: &Path, digest: &str) -> Value {
    let r = ok(&run(exe, home, &["update", "--confirm", digest]));
    assert_eq!(r["mode"], json!("applied"), "{r}");
    sleep(Duration::from_millis(1100)); // receipts carry second-resolution timestamps
    r
}

fn digest(r: &Value) -> String {
    r["plan"]["plan_digest"].as_str().unwrap().to_string()
}

fn host<'a>(r: &'a Value, name: &str) -> &'a Value {
    r.as_array()
        .unwrap()
        .iter()
        .find(|h| h["host"] == name)
        .unwrap()
}

fn plan_actions(r: &Value, name: &str) -> Vec<(String, String)> {
    host(&r["plan"]["hosts"], name)["files"]
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

// AC-1: one `update` previews, `--confirm <digest>` applies (receipts committed), and the
// next preview is unchanged everywhere.
#[test]
fn update_previews_then_applies_on_confirmation_then_is_unchanged() {
    let f = fx();
    let a = install_version(&f.home, "a", &v1());
    let p = preview(&a, &f.home);
    assert_eq!(
        p["package"],
        json!(f.home.join(".moai-cli/versions/a/skills/moai"))
    );
    assert!(p["confirm_with"]
        .as_str()
        .unwrap()
        .contains(&format!("moai-cli update --confirm {}", digest(&p))));
    for h in HOSTS {
        assert_eq!(
            host(&p["plan"]["hosts"], h)["preflight"],
            json!("ok"),
            "{p}"
        );
        assert!(
            plan_actions(&p, h).iter().all(|(_, a)| a == "create"),
            "{p}"
        );
    }
    let r = confirm(&a, &f.home, &digest(&p));
    assert_eq!(r["result"]["status"], json!("applied"), "{r}");
    for h in HOSTS {
        let o = host(&r["result"]["hosts"], h);
        assert_eq!(o["status"], json!("applied"), "{r}");
        assert_eq!(o["receipt_status"], json!("committed"), "{r}");
        let skill = f.home.join(match h {
            "claude" => ".claude/skills/moai/SKILL.md",
            _ => ".agents/skills/moai/SKILL.md",
        });
        assert_eq!(fs::read_to_string(skill).unwrap(), skill_md("body v1"));
    }
    let again = preview(&a, &f.home);
    for h in HOSTS {
        assert!(
            plan_actions(&again, h)
                .iter()
                .all(|(_, a)| a == "unchanged"),
            "{again}"
        );
    }
}

// AC-2: without a confirmation nothing is written; a stale or foreign digest writes nothing.
#[test]
fn without_a_matching_confirmation_nothing_is_written() {
    let f = fx();
    let a = install_version(&f.home, "a", &v1());
    let p = preview(&a, &f.home);
    let before = state(&f.home);
    let out = run(&a, &f.home, &["update", "--confirm", "sha256:"]);
    assert!(!out.status.success());
    assert_eq!(envelope(&out)["ok"], json!(false));
    let wrong = format!("sha256:{}", "0".repeat(64));
    let out = run(&a, &f.home, &["update", "--confirm", &wrong]);
    assert!(!out.status.success());
    assert_eq!(state(&f.home), before);
    // The preview's own digest still applies.
    confirm(&a, &f.home, &digest(&p));
}

// AC-2: a file the user edited is kept (conflict), the others update.
#[test]
fn a_user_edit_is_preserved() {
    let f = fx();
    let a = install_version(&f.home, "a", &v1());
    let p = preview(&a, &f.home);
    confirm(&a, &f.home, &digest(&p));
    let edited = f.home.join(".claude/skills/moai/workflows/update.md");
    fs::write(&edited, "# UPDATE\nmy own notes\n").unwrap();
    fs::write(
        f.home
            .join(".moai-cli/versions/a/skills/moai/workflows/update.md"),
        "# UPDATE\nv2\n",
    )
    .unwrap();
    fs::write(
        f.home.join(".moai-cli/versions/a/skills/moai/SKILL.md"),
        skill_md("body v2"),
    )
    .unwrap();
    let p = preview(&a, &f.home);
    assert!(plan_actions(&p, "claude").contains(&("workflows/update.md".into(), "conflict".into())));
    confirm(&a, &f.home, &digest(&p));
    assert_eq!(
        fs::read_to_string(&edited).unwrap(),
        "# UPDATE\nmy own notes\n"
    );
    assert_eq!(
        fs::read_to_string(f.home.join(".claude/skills/moai/SKILL.md")).unwrap(),
        skill_md("body v2")
    );
}

// AC-3: version A installs, then version B's `update` previews ok and applies (#2521).
#[test]
fn a_newer_version_folder_updates_what_the_older_installed() {
    let f = fx();
    let a = install_version(&f.home, "a", &v1());
    let p = preview(&a, &f.home);
    confirm(&a, &f.home, &digest(&p));
    let b = install_version(&f.home, "b", &v2());
    let p = preview(&b, &f.home);
    for h in HOSTS {
        let hp = host(&p["plan"]["hosts"], h);
        assert_eq!(hp["preflight"], json!("ok"), "{p}");
        assert!(hp["receipt_from"].is_string(), "{p}");
    }
    let r = confirm(&b, &f.home, &digest(&p));
    assert_eq!(r["result"]["status"], json!("applied"), "{r}");
    let again = preview(&b, &f.home);
    for h in HOSTS {
        assert!(
            plan_actions(&again, h)
                .iter()
                .all(|(_, a)| a == "unchanged"),
            "{again}"
        );
    }
}

// A binary with no package beside it says where the package is expected and writes nothing.
#[test]
fn no_package_beside_the_binary_is_an_error() {
    let f = fx();
    let bin = f.root.join("elsewhere/bin");
    fs::create_dir_all(&bin).unwrap();
    let exe = bin.join("moai-cli");
    common::copy_executable(Path::new(EXE), &exe);
    let before = state(&f.home);
    let out = run(&exe, &f.home, &["update"]);
    assert!(!out.status.success());
    let v = envelope(&out);
    assert_eq!(v["ok"], json!(false));
    let text = v.to_string();
    assert!(
        text.contains("skills/moai") && text.contains("update-plan"),
        "{text}"
    );
    assert_eq!(state(&f.home), before);
}

/// Terminal confirmation through a pseudo terminal (BSD `script`, macOS only here). The
/// answer is typed only after the question is on screen, as a person would.
#[cfg(target_os = "macos")]
fn run_in_terminal(exe: &Path, home: &Path, answer: &str) -> String {
    use std::io::Read as _;
    let mut child = Command::new("/usr/bin/script")
        .args(["-q", "/dev/null"])
        .arg(exe)
        .arg("update")
        .env_clear()
        .env("PATH", "/usr/bin:/bin")
        .env("HOME", home)
        .env("LC_ALL", "C")
        .current_dir(home)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::null())
        .spawn()
        .unwrap();
    let mut stdout = child.stdout.take().unwrap();
    let mut seen = Vec::new();
    let mut buf = [0u8; 4096];
    while !String::from_utf8_lossy(&seen).contains("[y/N]") {
        let n = stdout.read(&mut buf).unwrap();
        assert!(n > 0, "no question: {}", String::from_utf8_lossy(&seen));
        seen.extend_from_slice(&buf[..n]);
    }
    let mut stdin = child.stdin.take().unwrap();
    stdin.write_all(answer.as_bytes()).unwrap();
    stdout.read_to_end(&mut seen).unwrap();
    drop(stdin);
    child.wait().unwrap();
    String::from_utf8_lossy(&seen).to_string()
}

// In a terminal `update` asks once: anything but yes writes nothing, yes applies.
#[cfg(target_os = "macos")]
#[test]
fn in_a_terminal_update_asks_and_applies_only_on_yes() {
    let f = fx();
    let a = install_version(&f.home, "a", &v1());
    let before = state(&f.home);
    let text = run_in_terminal(&a, &f.home, "n\n");
    assert!(text.contains("Apply this update? [y/N]"), "{text}");
    assert!(text.contains("\"mode\":\"preview\""), "{text}");
    assert_eq!(state(&f.home), before);
    let text = run_in_terminal(&a, &f.home, "y\n");
    assert!(text.contains("\"mode\":\"applied\""), "{text}");
    assert!(f.home.join(".claude/skills/moai/SKILL.md").is_file());
}

// A confirmation piped in without a terminal is not a confirmation: only `--confirm`
// applies outside a terminal.
#[test]
fn a_yes_piped_without_a_terminal_writes_nothing() {
    let f = fx();
    let a = install_version(&f.home, "a", &v1());
    let before = state(&f.home);
    let mut child = Command::new(&a)
        .arg("update")
        .env_clear()
        .env("PATH", "/usr/bin:/bin")
        .env("HOME", &f.home)
        .env("LC_ALL", "C")
        .current_dir(&f.home)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    let mut stdin = child.stdin.take().unwrap();
    let _ = stdin.write_all(b"y\nyes\n");
    drop(stdin);
    let r = ok(&child.wait_with_output().unwrap());
    assert_eq!(r["mode"], json!("preview"), "{r}");
    assert_eq!(state(&f.home), before);
}

// A linked package folder is refused; nothing is written.
#[test]
fn a_linked_package_folder_is_refused() {
    let f = fx();
    let a = install_version(&f.home, "a", &v1());
    let skills = f.home.join(".moai-cli/versions/a/skills");
    let moved = f.root.join("pkg-elsewhere");
    fs::rename(skills.join("moai"), &moved).unwrap();
    std::os::unix::fs::symlink(&moved, skills.join("moai")).unwrap();
    let before = state(&f.home);
    let out = run(&a, &f.home, &["update"]);
    assert!(!out.status.success());
    let v = envelope(&out);
    assert!(v.to_string().contains("symlink; refused"), "{v}");
    assert_eq!(state(&f.home), before);
}

// ---- #2540: report what was actually applied; ask only when something would be written ----

/// A `skills/moai` folder this tool did not install, in both hosts' projection roots, so
/// every host of the plan is `blocked`.
fn foreign_skill_in_both_hosts(home: &Path) {
    for root in [".claude/skills/moai", ".agents/skills/moai"] {
        let dir = home.join(root);
        fs::create_dir_all(&dir).unwrap();
        fs::write(dir.join("SKILL.md"), skill_md("someone else's skill")).unwrap();
    }
}

// #2540 AC-1: a confirmed plan whose hosts are all blocked writes nothing and is not
// reported as applied.
#[test]
fn a_confirmed_plan_with_every_host_blocked_is_not_reported_as_applied() {
    let f = fx();
    let a = install_version(&f.home, "a", &v1());
    foreign_skill_in_both_hosts(&f.home);
    let p = preview(&a, &f.home);
    for h in HOSTS {
        assert_eq!(
            host(&p["plan"]["hosts"], h)["preflight"],
            json!("blocked"),
            "{p}"
        );
    }
    let before = state(&f.home);
    let r = ok(&run(&a, &f.home, &["update", "--confirm", &digest(&p)]));
    assert_eq!(r["result"]["status"], json!("blocked"), "{r}");
    assert_eq!(r["applied"], json!(false), "{r}");
    assert_eq!(state(&f.home), before, "nothing was written");
}

// #2540 AC-1: confirming a plan that is unchanged everywhere is not reported as applied.
#[test]
fn a_confirmed_plan_with_nothing_to_write_is_not_reported_as_applied() {
    let f = fx();
    let a = install_version(&f.home, "a", &v1());
    let p = preview(&a, &f.home);
    let r = confirm(&a, &f.home, &digest(&p));
    assert_eq!(r["applied"], json!(true), "{r}");
    let again = preview(&a, &f.home);
    let r = confirm(&a, &f.home, &digest(&again));
    assert_eq!(r["result"]["status"], json!("unchanged"), "{r}");
    assert_eq!(r["applied"], json!(false), "{r}");
}

// #2540 AC-3: a preview with nothing to write says so and offers no confirmation.
#[test]
fn a_preview_with_nothing_to_write_says_so() {
    let f = fx();
    let a = install_version(&f.home, "a", &v1());
    foreign_skill_in_both_hosts(&f.home);
    let p = preview(&a, &f.home);
    assert_eq!(p["nothing_to_apply"], json!(true), "{p}");
    let g = fx();
    let b = install_version(&g.home, "a", &v1());
    let p = preview(&b, &g.home);
    assert_eq!(p["nothing_to_apply"], json!(false), "{p}");
}

/// The everyday rerun: the package was applied once, so every host is `unchanged`.
fn applied_once() -> (Fx, PathBuf) {
    let f = fx();
    let a = install_version(&f.home, "a", &v1());
    let p = preview(&a, &f.home);
    confirm(&a, &f.home, &digest(&p));
    (f, a)
}

// #2540 AC-3 (pre-audit 6824540371 F1): a preview where every host is unchanged has nothing
// to apply.
#[test]
fn a_preview_of_an_already_applied_package_has_nothing_to_apply() {
    let (f, a) = applied_once();
    let p = preview(&a, &f.home);
    for h in HOSTS {
        assert_eq!(
            host(&p["plan"]["hosts"], h)["preflight"],
            json!("ok"),
            "{p}"
        );
        assert!(
            plan_actions(&p, h).iter().all(|(_, a)| a == "unchanged"),
            "{p}"
        );
    }
    assert_eq!(p["nothing_to_apply"], json!(true), "{p}");
}

// #2540 (pre-audit 6824540371 N1): a plan whose only difference is a file the user changed
// (a conflict, never written) has nothing to apply either.
#[test]
fn a_preview_with_only_a_conflict_has_nothing_to_apply() {
    let (f, a) = applied_once();
    fs::write(
        f.home.join(".claude/skills/moai/workflows/update.md"),
        "# UPDATE\nmy own notes\n",
    )
    .unwrap();
    fs::write(
        f.home
            .join(".moai-cli/versions/a/skills/moai/workflows/update.md"),
        "# UPDATE\nv2\n",
    )
    .unwrap();
    let p = preview(&a, &f.home);
    assert!(
        plan_actions(&p, "claude").contains(&("workflows/update.md".into(), "conflict".into())),
        "{p}"
    );
    assert!(
        !plan_actions(&p, "claude")
            .iter()
            .any(|(_, a)| a == "create" || a == "update"),
        "{p}"
    );
    // codex has no user edit, so it would update; only claude is conflict-only.
    assert_eq!(p["nothing_to_apply"], json!(false), "{p}");
    fs::write(
        f.home.join(".agents/skills/moai/workflows/update.md"),
        "# UPDATE\nmy own notes too\n",
    )
    .unwrap();
    let p = preview(&a, &f.home);
    assert_eq!(p["nothing_to_apply"], json!(true), "{p}");
}

/// Runs `update` in a pseudo terminal with nothing typed, and returns everything shown.
#[cfg(target_os = "macos")]
fn run_in_terminal_unanswered(exe: &Path, home: &Path) -> String {
    let out = Command::new("/usr/bin/script")
        .args(["-q", "/dev/null"])
        .arg(exe)
        .arg("update")
        .env_clear()
        .env("PATH", "/usr/bin:/bin")
        .env("HOME", home)
        .env("LC_ALL", "C")
        .current_dir(home)
        .stdin(Stdio::null())
        .stdout(Stdio::piped())
        .stderr(Stdio::null())
        .output()
        .unwrap();
    String::from_utf8_lossy(&out.stdout).to_string()
}

// #2540 AC-3: in a terminal, a plan with every host blocked is shown without the question.
#[cfg(target_os = "macos")]
#[test]
fn in_a_terminal_nothing_to_write_asks_nothing() {
    let f = fx();
    let a = install_version(&f.home, "a", &v1());
    foreign_skill_in_both_hosts(&f.home);
    let before = state(&f.home);
    let text = run_in_terminal_unanswered(&a, &f.home);
    assert!(!text.contains("[y/N]"), "{text}");
    assert!(text.contains("\"mode\":\"preview\""), "{text}");
    assert_eq!(state(&f.home), before);
}

// #2540 AC-3 (pre-audit 6824540371 F1): in a terminal, rerunning `update` on an already
// applied package asks nothing.
#[cfg(target_os = "macos")]
#[test]
fn in_a_terminal_an_already_applied_package_asks_nothing() {
    let (f, a) = applied_once();
    let before = state(&f.home);
    let text = run_in_terminal_unanswered(&a, &f.home);
    assert!(!text.contains("[y/N]"), "{text}");
    assert!(text.contains("\"nothing_to_apply\":true"), "{text}");
    assert_eq!(state(&f.home), before);
}

// #2545 (audit note U5 on #2530): the question goes to stderr, so `update` asks only when both
// stdin and stderr are terminals. With stdin on a terminal but stderr sent to a file, a person
// would wait on a question they cannot see: `update` must not ask and must write nothing.
#[cfg(target_os = "macos")]
#[test]
fn with_stderr_away_from_the_terminal_update_does_not_ask() {
    let f = fx();
    let a = install_version(&f.home, "a", &v1());
    let err_file = f.root.join("stderr.txt");
    let before = state(&f.home);
    let out = Command::new("/usr/bin/script")
        .args([
            "-q",
            "/dev/null",
            "/bin/sh",
            "-c",
            "exec \"$0\" update 2>\"$1\"",
        ])
        .arg(&a)
        .arg(&err_file)
        .env_clear()
        .env("PATH", "/usr/bin:/bin")
        .env("HOME", &f.home)
        .env("LC_ALL", "C")
        .current_dir(&f.home)
        .stdin(Stdio::null())
        .stdout(Stdio::piped())
        .stderr(Stdio::null())
        .output()
        .unwrap();
    let shown = String::from_utf8_lossy(&out.stdout);
    let err = fs::read_to_string(&err_file).unwrap_or_default();
    assert!(
        !err.contains("[y/N]"),
        "asked on a stderr nobody sees: {err}"
    );
    assert!(shown.contains("\"mode\":\"preview\""), "{shown}");
    assert_eq!(state(&f.home), before);
}
