//! SPEC-20261124-update-flat-agent-ownership (REQ-3, REQ-7, REQ-9, REQ-10, REQ-11): the wiring of
//! `moai update` around the flat agent file choice, run through the built binary. The library
//! tests pin the choice and the apply; these cases start the real `update` against the real
//! package folders copied beside a copy of the binary in a temporary HOME. A terminal run asks
//! and applies the answer, any other run keeps every conflict and answers `partial`, and a blocked
//! agents folder ends the run partly applied. The real HOME is never touched.

#![cfg(unix)]

mod common;
use common::pty::{output_within, Terminal, END_OF_INPUT};
use common::CleanEnv;

use serde_json::{json, Value};
use std::{
    fs,
    os::unix::fs::symlink,
    path::{Path, PathBuf},
    process::{Command, Output},
    thread::sleep,
    time::Duration,
};
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;
/// The person's flat file: it matches neither the record nor the render nor the package file.
const MINE: &str = "my own old copy\n";

struct Fx {
    _tmp: TempDir,
    home: PathBuf,
    exe: PathBuf,
    version: PathBuf,
}

fn copy_tree(from: &Path, to: &Path) {
    fs::create_dir_all(to).unwrap();
    for e in fs::read_dir(from).unwrap() {
        let e = e.unwrap();
        let target = to.join(e.file_name());
        if e.file_type().unwrap().is_dir() {
            copy_tree(&e.path(), &target);
        } else {
            fs::copy(e.path(), target).unwrap();
        }
    }
}

fn piped(fx: &Fx, args: &[&str]) -> Output {
    let mut cmd = Command::new(&fx.exe);
    cmd.args(args)
        .clean_env()
        .env("PATH", "/usr/bin:/bin")
        .env("HOME", &fx.home)
        .env("LC_ALL", "C")
        .current_dir(&fx.home);
    output_within(&mut cmd, None)
}

fn envelope(out: &Output) -> Value {
    let v: Value = serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
        panic!(
            "envelope JSON on stdout: {}{}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    });
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

/// A temporary HOME holding a copy of the binary in a version folder with the real skill package
/// and agent sources beside it, after one scripted `update --confirm` has applied everything: the
/// base agents are installed nested and the next `update` has nothing to apply.
fn installed() -> Fx {
    let tmp = TempDir::new().unwrap();
    let home = fs::canonicalize(tmp.path()).unwrap().join("home");
    let version = moai_cli::home_root::path(&home).join("versions/a");
    fs::create_dir_all(version.join("bin")).unwrap();
    let exe = version.join("bin/moai-cli");
    common::copy_executable(Path::new(EXE), &exe);
    fs::hard_link(&exe, version.join("bin/moai")).unwrap();
    let repo = Path::new(env!("CARGO_MANIFEST_DIR"));
    copy_tree(&repo.join("skills/moai"), &version.join("skills/moai"));
    copy_tree(&repo.join("agents"), &version.join("agents"));
    let fx = Fx {
        _tmp: tmp,
        home,
        exe,
        version,
    };
    let preview = envelope(&piped(&fx, &["update"]));
    let digest = preview["plan_digest"].as_str().unwrap().to_string();
    let applied = envelope(&piped(&fx, &["update", "--confirm", &digest]));
    assert_eq!(applied["mode"], json!("applied"), "{applied}");
    assert_eq!(applied["agents_result"]["state"], json!("applied"));
    sleep(Duration::from_millis(1100)); // receipts carry second-resolution timestamps
    let again = envelope(&piped(&fx, &["update"]));
    assert_eq!(again["nothing_to_apply"], json!(true), "{again}");
    fx
}

impl Fx {
    fn agents(&self) -> PathBuf {
        self.home.join(".claude/agents")
    }
    fn flat_impl(&self) -> PathBuf {
        self.agents().join("moai-impl.md")
    }
    fn nested_impl(&self) -> PathBuf {
        self.agents().join("moai/moai-impl.md")
    }
    /// A flat copy of the impl agent beside the nested one: a conflict and nothing else to write.
    fn put_the_persons_flat_file(&self) {
        fs::write(self.flat_impl(), MINE).unwrap();
    }
    /// The receipt folders that kept the person's flat file, as the bytes they kept.
    fn kept_in_receipts(&self) -> Vec<String> {
        let dir = moai_cli::home_root::path(&self.home).join("agents/receipts");
        fs::read_dir(dir)
            .unwrap()
            .filter_map(|e| {
                let p = e.unwrap().path().join("before/claude/legacy/moai-impl.md");
                fs::read_to_string(p).ok()
            })
            .collect()
    }
}

/// `moai-cli update` at a terminal: one pseudo terminal as stdin, stdout and stderr, a clean
/// environment, the temporary HOME, English screens and no colour. The answers are typed once
/// the first question is on screen; an end of input closes any later question the run opens (the
/// model stage), so a run never waits for a person.
fn at_terminal(fx: &Fx, answers: &str) -> String {
    let typed = format!("{answers}{}", END_OF_INPUT as char);
    let s = Terminal::new(&fx.exe)
        .args(["update"])
        .env("PATH", "/usr/bin:/bin")
        .env("HOME", &fx.home)
        .env("LC_ALL", "C")
        .env("TERM", "xterm-256color")
        .env("NO_COLOR", "1")
        .env("COLUMNS", "80")
        .env("MOAI_LANG", "en")
        .cwd(&fx.home)
        .answer(typed)
        .run();
    s.text
}

/// REQ-11, REQ-3, REQ-5: a conflict alone is no up-to-date run. At a terminal the update asks the
/// confirming question, then asks per file, and the answer use MoAI's moves the file with the
/// person's bytes kept in the receipt.
#[test]
fn a_terminal_run_asks_about_a_conflict_alone_and_use_moais_moves_the_file() {
    let fx = installed();
    let nested_before = fs::read(fx.nested_impl()).unwrap();
    fx.put_the_persons_flat_file();
    let text = at_terminal(&fx, "y\nm\n");
    assert!(text.contains("[y/N]"), "no confirming question: {text}");
    assert!(text.contains("keep mine (k)"), "no choice: {text}");
    assert!(text.contains("moai-impl.md"), "{text}");
    assert!(!fx.flat_impl().exists(), "the flat file stayed: {text}");
    assert_eq!(fs::read(fx.nested_impl()).unwrap(), nested_before);
    assert_eq!(fx.kept_in_receipts(), [MINE], "{text}");
    assert!(!text.contains("Partly applied"), "{text}");
}

/// REQ-4, REQ-7, REQ-9, REQ-10: show the difference, then keep mine: both paths stay as they
/// were, and the run ends partly applied with the kept file counted on the agents row.
#[test]
fn a_terminal_run_shows_the_difference_and_keep_mine_ends_partly_applied() {
    let fx = installed();
    let nested_before = fs::read(fx.nested_impl()).unwrap();
    fx.put_the_persons_flat_file();
    let text = at_terminal(&fx, "y\nd\nk\n");
    assert!(
        text.lines().any(|l| l.starts_with("- my own old copy")),
        "no difference: {text}"
    );
    assert_eq!(fs::read_to_string(fx.flat_impl()).unwrap(), MINE, "{text}");
    assert_eq!(fs::read(fx.nested_impl()).unwrap(), nested_before);
    assert!(fx.kept_in_receipts().is_empty(), "{text}");
    assert!(text.contains("Partly applied"), "{text}");
    assert!(
        text.lines()
            .any(|l| l.contains("agents") && l.contains("1 kept")),
        "{text}"
    );
}

/// REQ-7, REQ-9, REQ-11: a run that is no terminal never asks: the preview is not empty, the
/// confirmed apply keeps the file and answers `partial` with the kept name.
#[test]
fn a_run_without_a_terminal_keeps_every_conflict_and_answers_partial() {
    let fx = installed();
    fx.put_the_persons_flat_file();
    let preview = envelope(&piped(&fx, &["update"]));
    assert_eq!(preview["nothing_to_apply"], json!(false), "{preview}");
    let digest = preview["plan_digest"].as_str().unwrap().to_string();
    let applied = envelope(&piped(&fx, &["update", "--confirm", &digest]));
    assert_eq!(applied["agents_result"]["state"], json!("partial"));
    assert_eq!(
        applied["agents_result"]["kept"],
        json!(["claude:moai-impl.md"]),
        "{applied}"
    );
    assert_eq!(fs::read_to_string(fx.flat_impl()).unwrap(), MINE);
}

/// A link at the agents folder blocks the host: nothing is written through it, and the run that
/// applied the rest ends partly applied with the reason on the agents row, not on Done.
#[test]
fn a_blocked_agents_folder_ends_the_update_partly_applied() {
    let fx = installed();
    // Something else to apply: the package skill text changed since the install.
    let skill = fx.version.join("skills/moai/SKILL.md");
    let text = fs::read_to_string(&skill).unwrap();
    fs::write(&skill, format!("{text}\nA line added after the install.\n")).unwrap();
    let outside = fx.home.join("outside-agents");
    fs::rename(fx.agents(), &outside).unwrap();
    symlink(&outside, fx.agents()).unwrap();
    let before = fs::read_dir(&outside).unwrap().count();
    let text = at_terminal(&fx, "y\n");
    assert!(text.contains("Partly applied"), "{text}");
    assert!(
        text.lines()
            .any(|l| l.contains("agents") && l.contains("claude:")),
        "the agents row names the host and the reason: {text}"
    );
    assert_eq!(
        fs::read_dir(&outside).unwrap().count(),
        before,
        "wrote through the link"
    );
}
