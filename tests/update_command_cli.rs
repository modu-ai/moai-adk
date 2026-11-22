//! #2522: one-line `moai-cli update` — preview, confirm, apply. The package it applies is
//! the one installed with the binary, `<version folder>/skills/moai` next to
//! `<version folder>/bin/moai-cli`. Every case runs copies of the built binary inside a
//! temporary HOME; the real HOME and the installed binary are never touched.

#![cfg(unix)]

mod common;
use common::pty::{output_within, Terminal, END_OF_INPUT};
use common::CleanEnv;

use moai_cli::model_effort_settings::default_choice;
use serde_json::{json, Value};
use std::{
    fs,
    io::Write as _,
    path::{Path, PathBuf},
    process::{Command, Output},
    thread::sleep,
    time::Duration,
};
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;
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

/// `~/.moai-cli/versions/<version>/bin/moai-cli` with its package beside it, and an empty
/// agent source folder `agents/` (#2721: a release without agent sources installs none).
/// `bin/moai` is already the hard link an installed folder has (#2791), so these cases see
/// only the user layer's own writes.
fn install_version(home: &Path, version: &str, pkg: &[(&str, String)]) -> PathBuf {
    let folder = home.join(".moai-cli/versions").join(version);
    let bin = folder.join("bin");
    fs::create_dir_all(&bin).unwrap();
    fs::create_dir_all(folder.join("agents")).unwrap();
    let exe = bin.join("moai-cli");
    common::copy_executable(Path::new(EXE), &exe);
    fs::hard_link(&exe, bin.join("moai")).unwrap();
    for (rel, body) in pkg {
        let p = folder.join("skills/moai").join(rel);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        fs::write(p, body).unwrap();
    }
    exe
}

fn run(exe: &Path, home: &Path, args: &[&str]) -> Output {
    let mut cmd = Command::new(exe);
    cmd.args(args)
        .clean_env()
        .env("PATH", "/usr/bin:/bin")
        .env("HOME", home)
        .env("LC_ALL", "C")
        .current_dir(home);
    output_within(&mut cmd, None)
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

/// #2638: `--confirm` takes the combined digest of the skills, agents and routing plans; the
/// skill plan's own digest (`plan.plan_digest`) is refused.
fn digest(r: &Value) -> String {
    r["plan_digest"].as_str().unwrap().to_string()
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
        .contains(&format!("moai update --confirm {}", digest(&p))));
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

/// `moai-cli <args>` at a terminal: one pseudo terminal as stdin, stdout and stderr (the shared
/// helper makes it with `openpty`, so Linux and macOS run the same cases), a clean environment, the
/// temporary HOME, English screens and no colour. Add the answer, a hook or an `env` on the result.
fn at_terminal(exe: &Path, home: &Path, args: &[&str]) -> Terminal {
    Terminal::new(exe)
        .args(args)
        .env("PATH", "/usr/bin:/bin")
        .env("HOME", home)
        .env("LC_ALL", "C")
        .env("TERM", "xterm-256color")
        .env("NO_COLOR", "1")
        .env("COLUMNS", "80")
        .env("MOAI_LANG", "en")
        .cwd(home)
}

/// Terminal confirmation: `update` at a terminal. The answer is typed only after the question
/// is on screen, as a person would.
fn run_in_terminal(exe: &Path, home: &Path, answer: &str) -> String {
    let s = at_terminal(exe, home, &["update"]).answer(answer).run();
    // Every terminal path of `update` that is asked or told something ends with exit 0.
    assert_eq!(s.code, Some(0), "{}", s.text);
    s.text
}

/// The saved settings files under the home, by path (the Python case looked at the whole tree).
fn launcher_files(home: &Path) -> Vec<PathBuf> {
    state(home)
        .into_iter()
        .map(|(p, _)| p)
        .filter(|p| p.ends_with("launcher.yaml"))
        .collect()
}

// In a terminal `update` asks once: anything but yes writes nothing, yes applies.
#[test]
fn in_a_terminal_update_asks_and_applies_only_on_yes() {
    let f = fx();
    let a = install_version(&f.home, "a", &v1());
    let before = state(&f.home);
    // #2894 U1: a terminal gets the person's screen, not the envelope.
    let text = run_in_terminal(&a, &f.home, "n\n");
    assert!(text.contains("Apply this update? [y/N]"), "{text}");
    assert!(text.contains("Nothing was written."), "{text}");
    assert!(text.contains("MoAI update"), "{text}");
    assert!(text.contains("Checking"), "{text}");
    assert!(
        !text.contains("sha256:"),
        "a screen shows no digest: {text}"
    );
    assert!(
        !text.contains("\"mode\""),
        "no envelope on a terminal: {text}"
    );
    assert_eq!(state(&f.home), before);
    // Card C-47: an update that wrote files goes on to the model stage; the tool pick, the four
    // model questions of the one installed tool and the save question are answered by Enter and a no.
    let text = run_in_terminal(&a, &f.home, "y\n\n\n\n\n\nn\n");
    assert!(text.contains("Applied") && text.contains("Done"), "{text}");
    assert!(text.contains("Save these model settings? [y/N]"), "{text}");
    assert!(
        !text.contains("\"mode\""),
        "no envelope on a terminal: {text}"
    );
    assert!(f.home.join(".claude/skills/moai/SKILL.md").is_file());
}

// A terminal run that applies the update goes on to the model stage and saves the choice after
// a yes (card C-47, SPEC-20261118-update-model-pick).
#[test]
fn in_a_terminal_an_applied_update_asks_the_model_stage_and_saves_after_a_yes() {
    let f = fx();
    let a = install_version(&f.home, "a", &v1());
    let text = run_in_terminal(&a, &f.home, "y\n\n\n\n\n\ny\n");
    assert!(text.contains("Save these model settings? [y/N]"), "{text}");
    // Exactly one settings file was written, and it is the one the command reads.
    let files = launcher_files(&f.home);
    assert_eq!(files.len(), 1, "{files:?}\n{text}");
    let saved = fs::read_to_string(moai_cli::update_models::settings_path(&f.home)).unwrap();
    let saved: Value = serde_json::from_str(&saved).unwrap();
    // The answers were Enter, so the saved choice is the default of the source, not a copied literal.
    let want = default_choice("claude", "leader").expect("claude has a leader default");
    assert_eq!(
        saved["launcher"]["models"]["claude"]["leader"],
        json!({"model": want.model, "effort": want.effort}),
        "{text}"
    );
    // The model stage opened with the tool question, and the end screen sends each project to
    // its own session's update (the sentence wraps and uses no-break spaces).
    assert!(
        text.contains(&moai_cli::model_effort_questions::tool_question(&[]).prompt),
        "{text}"
    );
    assert!(text.contains("Done"), "{text}");
    let flat = flat_of(&text);
    assert!(
        flat.contains("run /moai update in a session there"),
        "{flat}"
    );
    assert!(!text.contains("\"protocol_version\""), "{text}");
    // A second run has nothing left to apply and points to the same step.
    let again = run_in_terminal_unanswered(&a, &f.home);
    assert!(again.contains("Everything is up to date"), "{again}");
    assert!(!again.contains("[y/N]"), "{again}");
    assert!(flat_of(&again).contains("$moai update"), "{again}");
}

// A scripted run (`--confirm`) asks nothing, also on a terminal: the model stage is for the
// person who answered the update question (card C-47, SPEC-20261118-update-model-pick).
#[test]
fn in_a_terminal_a_confirmed_update_asks_no_model_question() {
    let f = fx();
    let a = install_version(&f.home, "a", &v1());
    let p = preview(&a, &f.home);
    // Typed nothing: a question here would end with the question in the text, not hang.
    let text = at_terminal(&a, &f.home, &["update", "--confirm", &digest(&p)])
        .run()
        .text;
    assert!(
        f.home.join(".claude/skills/moai/SKILL.md").is_file(),
        "{text}"
    );
    assert!(
        !text.contains(&moai_cli::model_effort_questions::tool_question(&[]).prompt),
        "{text}"
    );
    assert!(!text.contains("Save these model settings"), "{text}");
}

/// A screen with its line wraps and no-break spaces read as single spaces.
fn flat_of(text: &str) -> String {
    text.replace("\n  ", " ").replace('\u{a0}', " ")
}

fn skill_file(home: &Path) -> PathBuf {
    home.join(".claude/skills/moai/SKILL.md")
}

// The cases below moved here from the Python terminal test of `update`
// (SPEC-20261120-python-binary-tests-to-rust); the Python module is gone.

// A pipe keeps the envelope, and `--json` on a pipe is the same bytes.
#[test]
fn a_pipe_keeps_the_envelope_and_json_is_the_same_bytes() {
    let f = fx();
    let a = install_version(&f.home, "a", &v1());
    let plain = run(&a, &f.home, &["update"]);
    let json_flag = run(&a, &f.home, &["update", "--json"]);
    let v = envelope(&plain);
    assert_eq!(v["operation"], json!("update.run"), "{v}");
    assert_eq!(v["result"]["mode"], json!("preview"), "{v}");
    assert_eq!(plain.stdout, json_flag.stdout);
}

// `--json` keeps the envelope on a terminal too.
#[test]
fn json_on_a_terminal_prints_the_envelope() {
    let f = fx();
    let a = install_version(&f.home, "a", &v1());
    let before = state(&f.home);
    let s = at_terminal(&a, &f.home, &["update", "--json"]).run();
    assert_eq!(s.code, Some(0), "{}", s.text);
    assert!(s.text.contains("\"protocol_version\""), "{}", s.text);
    assert!(
        s.text.contains("\"operation\":\"update.run\""),
        "{}",
        s.text
    );
    assert_eq!(state(&f.home), before);
}

// End of input (Ctrl-D) at the question is a no.
#[test]
fn end_of_input_at_the_question_writes_nothing() {
    let f = fx();
    let a = install_version(&f.home, "a", &v1());
    let before = state(&f.home);
    let s = at_terminal(&a, &f.home, &["update"])
        .answer([END_OF_INPUT])
        .run();
    assert_eq!(s.code, Some(0), "{}", s.text);
    assert_eq!(state(&f.home), before);
    assert!(!skill_file(&f.home).exists(), "{}", s.text);
}

// End of input at the tool pick ends the model stage as cancelled: the update stays applied
// and no settings file is written.
#[test]
fn end_of_input_at_the_model_stage_keeps_the_applied_update() {
    let f = fx();
    let a = install_version(&f.home, "a", &v1());
    let s = at_terminal(&a, &f.home, &["update"])
        .answer(b"y\n\x04")
        .run();
    assert_eq!(s.code, Some(0), "{}", s.text);
    assert!(skill_file(&f.home).is_file(), "{}", s.text);
    assert!(
        s.text
            .contains(&moai_cli::model_effort_questions::tool_question(&[]).prompt),
        "{}",
        s.text
    );
    assert!(!s.text.contains("Save these model settings?"), "{}", s.text);
    assert!(s.text.contains("Done"), "{}", s.text);
    assert!(
        !moai_cli::update_models::settings_path(&f.home).exists(),
        "{}",
        s.text
    );
    assert_eq!(launcher_files(&f.home), Vec::<PathBuf>::new(), "{}", s.text);
}

// The package changes after the preview, then `y`: nothing is written and the new preview shows.
#[test]
fn a_plan_that_changes_before_the_write_is_refused() {
    let f = fx();
    let a = install_version(&f.home, "a", &v1());
    let changed = a
        .parent()
        .unwrap()
        .parent()
        .unwrap()
        .join("skills/moai/workflows/update.md");
    let s = at_terminal(&a, &f.home, &["update"])
        .answer("y\n")
        .before_answer(move || {
            let mut file = fs::OpenOptions::new().append(true).open(&changed).unwrap();
            file.write_all(b"\nchanged after the preview\n").unwrap();
        })
        .run();
    assert_eq!(s.code, Some(0), "{}", s.text);
    assert!(
        s.text.contains("The plan changed after the preview"),
        "{}",
        s.text
    );
    assert!(!skill_file(&f.home).exists(), "{}", s.text);
}

// stdout on a terminal but stdin not: the screen and the confirming command, no question, no wait.
#[test]
fn stdin_that_is_not_a_terminal_is_never_asked() {
    let f = fx();
    let a = install_version(&f.home, "a", &v1());
    let before = state(&f.home);
    let started = std::time::Instant::now();
    let s = at_terminal(&a, &f.home, &["update"])
        .stdin_not_a_terminal()
        .run();
    assert_eq!(s.code, Some(0), "{}", s.text);
    assert!(!s.text.contains("[y/N]"), "{}", s.text);
    assert!(
        s.text.contains("moai update --confirm sha256:"),
        "{}",
        s.text
    );
    assert!(started.elapsed() < Duration::from_secs(30));
    assert_eq!(state(&f.home), before);
}

// A failed run on a terminal shows the error frame only; a pipe still gets the failure envelope.
#[test]
fn an_error_on_a_terminal_is_the_frame_not_json() {
    let f = fx();
    let a = install_version(&f.home, "a", &v1());
    fs::remove_dir_all(a.parent().unwrap().parent().unwrap().join("skills/moai")).unwrap();
    let s = at_terminal(&a, &f.home, &["update"]).run();
    assert_eq!(s.code, Some(1), "{}", s.text);
    assert!(s.text.contains("no skill package installed"), "{}", s.text);
    assert!(!s.text.contains("\"protocol_version\""), "{}", s.text);
    let piped = run(&a, &f.home, &["update"]);
    assert_eq!(piped.status.code(), Some(1));
    assert_eq!(envelope(&piped)["ok"], json!(false));
}

// The CLI's screens are English whatever the language settings say (#2825 R1).
#[test]
fn a_korean_setting_keeps_the_english_screens() {
    let f = fx();
    let a = install_version(&f.home, "a", &v1());
    let s = at_terminal(&a, &f.home, &["update"])
        .env("MOAI_LANG", "ko")
        .answer("n\n")
        .run();
    assert_eq!(s.code, Some(0), "{}", s.text);
    assert!(s.text.contains("Checking"), "{}", s.text);
    assert!(s.text.contains("Nothing was written."), "{}", s.text);
    assert!(!s.text.contains("검사"), "{}", s.text);
}

// A confirmation piped in without a terminal is not a confirmation: only `--confirm`
// applies outside a terminal.
#[test]
fn a_yes_piped_without_a_terminal_writes_nothing() {
    let f = fx();
    let a = install_version(&f.home, "a", &v1());
    let before = state(&f.home);
    let mut cmd = Command::new(&a);
    cmd.arg("update")
        .clean_env()
        .env("PATH", "/usr/bin:/bin")
        .env("HOME", &f.home)
        .env("LC_ALL", "C")
        .current_dir(&f.home);
    let r = ok(&output_within(&mut cmd, Some(b"y\nyes\n")));
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
/// Makes the config-folder hook layer `unchanged` the only way that is stable across
/// processes: one real update cycle (SPEC-20261122-next-flow-hooks — the exec-form entries
/// name the absolute path of the binary that rendered them, so pre-written bytes from this
/// test process would always read as an update). The skill copies stay blocked, so only the
/// hook layer writes.
fn hooks_current(a: &Path, home: &Path) {
    let p = preview(a, home);
    confirm(a, home, &digest(&p));
}

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
    hooks_current(&a, &f.home);
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
    hooks_current(&a, &f.home);
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

/// Runs `update` at a terminal with nothing typed, and returns everything shown.
fn run_in_terminal_unanswered(exe: &Path, home: &Path) -> String {
    let s = at_terminal(exe, home, &["update"]).run();
    assert_eq!(s.code, Some(0), "{}", s.text);
    s.text
}

/// A model setting saved the way the model step saves one. The model questions open on a terminal
/// when none is saved (SPEC-20261121-model-step-reachable), so a terminal case that is about
/// something else saves one first.
fn with_a_saved_model_setting(home: &Path) {
    let path = moai_cli::update_models::settings_path(home);
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(
        path,
        r#"{"launcher":{"models":{"claude":{"leader":{"model":"claude-opus-5-5","effort":"high"}}}}}"#,
    )
    .unwrap();
}

// #2540 AC-3: in a terminal, a plan with every host blocked is shown without the question.
#[test]
fn in_a_terminal_nothing_to_write_asks_nothing() {
    let f = fx();
    let a = install_version(&f.home, "a", &v1());
    foreign_skill_in_both_hosts(&f.home);
    hooks_current(&a, &f.home);
    with_a_saved_model_setting(&f.home);
    let before = state(&f.home);
    let text = run_in_terminal_unanswered(&a, &f.home);
    assert!(!text.contains("[y/N]"), "{text}");
    // #2894 U1: the blocked screen instead of the preview envelope.
    assert!(
        text.contains("Nothing can be applied — clear the blocked part first"),
        "{text}"
    );
    assert_eq!(state(&f.home), before);
}

// #2540 AC-3 (pre-audit 6824540371 F1): in a terminal, rerunning `update` on an already
// applied package asks nothing.
#[test]
fn in_a_terminal_an_already_applied_package_asks_nothing() {
    let (f, a) = applied_once();
    with_a_saved_model_setting(&f.home);
    let before = state(&f.home);
    let text = run_in_terminal_unanswered(&a, &f.home);
    assert!(!text.contains("[y/N]"), "{text}");
    // #2894 U1: the up-to-date screen instead of `nothing_to_apply` in the envelope.
    assert!(text.contains("Everything is up to date"), "{text}");
    assert_eq!(state(&f.home), before);
}

// #2545 (audit note U5 on #2530): the question goes to stderr, so `update` asks only when both
// stdin and stderr are terminals. With stdin on a terminal but stderr sent to a file, a person
// would wait on a question they cannot see: `update` must not ask and must write nothing.
#[test]
fn with_stderr_away_from_the_terminal_update_does_not_ask() {
    let f = fx();
    let a = install_version(&f.home, "a", &v1());
    let err_file = f.root.join("stderr.txt");
    let before = state(&f.home);
    // stdin and stdout on the terminal, stderr sent to a file by the shell that starts `update`.
    let shown = at_terminal(
        Path::new("/bin/sh"),
        &f.home,
        &[
            "-c",
            "exec \"$0\" update 2>\"$1\"",
            a.to_str().unwrap(),
            err_file.to_str().unwrap(),
        ],
    )
    .run()
    .text;
    let err = fs::read_to_string(&err_file).unwrap_or_default();
    assert!(
        !err.contains("[y/N]"),
        "asked on a stderr nobody sees: {err}"
    );
    // #2894 U1: stdout is a terminal, so the person's screen ends with the way to answer
    // later instead of a question nobody would see.
    assert!(
        shown.contains("next: run it again in a terminal to answer, or"),
        "{shown}"
    );
    assert!(shown.contains("moai update --confirm sha256:"), "{shown}");
    assert_eq!(state(&f.home), before);
}

// #2791 AC-1: a version folder without `bin/moai` gets the hard link from the same confirmed
// update; the preview names it and writes nothing.
#[test]
fn update_links_the_moai_name_in_its_version_folder() {
    use std::os::unix::fs::MetadataExt as _;
    // The user layer is already applied: the missing link alone is something to apply.
    let (f, a) = applied_once();
    let alias = a.with_file_name("moai");
    fs::remove_file(&alias).unwrap();
    let p = preview(&a, &f.home);
    assert_eq!(p["bin_names"]["state"], json!("create"), "{p}");
    assert_eq!(p["nothing_to_apply"], json!(false), "{p}");
    assert!(!alias.exists(), "the preview links nothing");
    let r = confirm(&a, &f.home, &digest(&p));
    assert_eq!(r["bin_names_result"]["state"], json!("applied"), "{r}");
    let (m1, m2) = (fs::metadata(&a).unwrap(), fs::metadata(&alias).unwrap());
    assert_eq!(
        (m1.dev(), m1.ino()),
        (m2.dev(), m2.ino()),
        "one file, two names"
    );
    let again = preview(&a, &f.home);
    assert_eq!(again["bin_names"]["state"], json!("unchanged"), "{again}");
    assert_eq!(again["nothing_to_apply"], json!(true), "{again}");
}
