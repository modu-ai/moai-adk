//! SPEC-20261122-home-update-notice-follow-ups: the "no longer managed" notice of the home update
//! tells the person that removing the old Codex and OMP files is their own step, the Korean screens
//! say it in Korean, and the apply screen names the host of every skipped row. Every case that runs
//! the binary uses a temporary HOME; the real one is never touched.

#![cfg(unix)]

mod common;
use common::CleanEnv;

use moai_cli::ui::Painter;
use moai_cli::ui_text::Lang;
use moai_cli::update_tux::{screen, Asked, View};
use serde_json::{json, Value};
use std::{
    fs,
    path::{Path, PathBuf},
    process::{Command, Output, Stdio},
};
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;
const REMOVE_EN: &str = "remove them yourself if you do not use them";
const KEPT_KO: &str = "더는 관리하지 않아요";
const REMOVE_KO: &str = "쓰지 않으면 직접 지우세요";

fn root() -> PathBuf {
    PathBuf::from(env!("CARGO_MANIFEST_DIR"))
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
    let base = fs::canonicalize(tmp.path()).unwrap();
    let home = base.join("home");
    let folder = home.join(".moai-cli/versions/a");
    fs::create_dir_all(folder.join("bin")).unwrap();
    let exe = folder.join("bin/moai-cli");
    common::copy_executable(Path::new(EXE), &exe);
    copy_tree(&root().join("skills/moai"), &folder.join("skills/moai"));
    copy_tree(&root().join("agents"), &folder.join("agents"));
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

fn put(home: &Path, rel: &str, text: &str) {
    let path = home.join(rel);
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, text).unwrap();
}

fn unmanaged(layer: &Value) -> Vec<(String, String)> {
    layer["skipped"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|s| s["part"] == json!("unmanaged"))
        .map(|s| {
            (
                s["host"].as_str().unwrap().to_string(),
                s["reason"].as_str().unwrap().to_string(),
            )
        })
        .collect()
}

/// AC-1: each host's notice ends with the removal sentence, and the plan digest does not move
/// when the folders are there.
#[test]
fn the_notice_says_to_remove_the_old_files_yourself() {
    let (_t, home, exe) = setup();
    let without = result(&run(&exe, &home, &["update"]));
    put(&home, ".codex/AGENTS.md", "# mine\n");
    put(&home, ".omp/agent/agents/mine.md", "mine\n");
    let with = result(&run(&exe, &home, &["update"]));
    assert_eq!(with["plan_digest"], without["plan_digest"], "{with}");
    let notices = unmanaged(&with);
    let hosts: Vec<&str> = notices.iter().map(|(h, _)| h.as_str()).collect();
    assert_eq!(hosts, ["codex", "omp"], "{with}");
    for (host, reason) in &notices {
        assert!(reason.contains("no longer manages"), "{host}: {reason}");
        assert!(reason.ends_with(REMOVE_EN), "{host}: {reason}");
    }
}

fn view(lang: Lang) -> View {
    View {
        painter: Painter {
            color: false,
            unicode: true,
            lang,
        },
        width: 200,
        home: "/home/u".into(),
        project: None,
    }
}

fn fixture(name: &str) -> Value {
    let p = root().join(format!("tests/fixtures/update_tux/{name}.json"));
    serde_json::from_str(&fs::read_to_string(&p).unwrap()).unwrap()
}

fn with_skipped(mut v: Value) -> Value {
    v["skipped"] = json!([
        {"part": "unmanaged", "host": "codex",
         "reason": format!("Codex files under ~/.codex and ~/.agents/skills are left as they are; moai no longer manages them; {REMOVE_EN}")},
        {"part": "unmanaged", "host": "omp",
         "reason": format!("OMP files under ~/.omp are left as they are; moai no longer manages them; {REMOVE_EN}")},
        {"part": "agents", "host": "claude", "reason": "the skill copy is blocked"},
    ]);
    v
}

/// The screen line whose row label starts with `label`.
fn line_with<'a>(text: &'a str, label: &str) -> &'a str {
    text.lines()
        .find(|l| l.trim_start().split_once(' ').is_some() && l.contains(label))
        .unwrap_or_else(|| panic!("no row {label:?} in:\n{text}"))
}

/// AC-2 and AC-3: the preview and the apply screens print the unmanaged rows with their host, in
/// English with the English sentence and in Korean with the Korean words; a Claude skipped part
/// also carries its host on the apply screen.
#[test]
fn the_unmanaged_notice_reads_in_both_languages_and_names_the_host() {
    let preview = with_skipped(fixture("preview-changes"));
    let applied = with_skipped(fixture("applied"));
    for lang in [Lang::En, Lang::Ko] {
        let v = view(lang);
        for (name, text) in [
            ("preview", screen(&preview, Asked::NotAsked, &v)),
            ("applied", screen(&applied, Asked::Yes, &v)),
        ] {
            for host in ["codex", "omp"] {
                let row = line_with(&text, &format!("unmanaged {host}"));
                if lang == Lang::Ko {
                    assert!(row.contains(KEPT_KO), "{name} {host}: {row}");
                    assert!(row.contains(REMOVE_KO), "{name} {host}: {row}");
                    assert!(!row.contains(REMOVE_EN), "{name} {host}: {row}");
                } else {
                    assert!(row.contains(REMOVE_EN), "{name} {host}: {row}");
                }
            }
            line_with(&text, "agents claude");
        }
    }
}
