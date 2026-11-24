//! SPEC-20261124-installed-version-by-name: an installation id is the hash of the binary path,
//! and a build runs under two names, `moai` and `moai-cli` (hard-linked in its version folder).
//! The update baseline of an earlier version folder was read under the running build's name only,
//! so a receipt written under the other name was never found and the screen said
//! `installed none`. Every case runs copies of the built binary as child processes inside a
//! temporary HOME; the real HOME is never touched.

#![cfg(unix)]

mod common;
use common::CleanEnv;

use moai_cli::ui::Painter;
use moai_cli::ui_text::Lang;
use moai_cli::update_tux::{screen, Asked, View};
use serde_json::{json, Value};
use sha2::{Digest, Sha256};
use std::{
    fs,
    io::Write as _,
    path::{Path, PathBuf},
    process::{Command, Output, Stdio},
    thread::sleep,
    time::Duration,
};
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;
const HOST: &str = "claude";
/// The version the earlier folder's receipts are rewritten to, as if an earlier build wrote them.
const EARLIER: &str = "2026.11.0-rc.2+local.4";
const LATER: &str = "2026.11.0-rc.2+local.5";

struct Fixture {
    _tmp: TempDir,
    home: PathBuf,
    stage: PathBuf,
}

fn fixture() -> Fixture {
    let tmp = TempDir::new().unwrap();
    let root = fs::canonicalize(tmp.path()).unwrap();
    let home = root.join("home");
    fs::create_dir_all(&home).unwrap();
    let stage = root.join("stage");
    for (rel, body) in [
        (
            "SKILL.md",
            "---\nname: moai\ndescription: fixture package\n---\n# MoAI\nbody\n",
        ),
        ("workflows/update.md", "# UPDATE\nv1\n"),
    ] {
        let p = stage.join(rel);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        fs::write(p, body).unwrap();
    }
    Fixture {
        _tmp: tmp,
        home,
        stage,
    }
}

fn state(home: &Path) -> PathBuf {
    home.join(".moai-cli")
}

fn bin_dir(home: &Path, version: &str) -> PathBuf {
    let bin = state(home).join("versions").join(version).join("bin");
    fs::create_dir_all(&bin).unwrap();
    bin
}

/// A copy of the built binary at `versions/<version>/bin/<name>`.
fn put(home: &Path, version: &str, name: &str) -> PathBuf {
    let exe = bin_dir(home, version).join(name);
    common::copy_executable(Path::new(EXE), &exe);
    exe
}

fn run(exe: &Path, home: &Path, args: &[&str], stdin: Option<&Value>) -> Output {
    let mut child = Command::new(exe)
        .args(args)
        .clean_env()
        .env("PATH", "/usr/bin:/bin")
        .env("HOME", home)
        .env("LC_ALL", "C")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    let mut pipe = child.stdin.take().unwrap();
    if let Some(v) = stdin {
        let _ = pipe.write_all(v.to_string().as_bytes());
    }
    drop(pipe);
    child.wait_with_output().unwrap()
}

fn ok(out: &Output) -> Value {
    let v: Value = serde_json::from_slice(&out.stdout).expect("envelope JSON on stdout");
    assert_eq!(
        v["ok"],
        json!(true),
        "{}{}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    v["result"].clone()
}

fn plan(exe: &Path, f: &Fixture) -> Value {
    ok(&run(
        exe,
        &f.home,
        &[
            "update-plan",
            f.home.to_str().unwrap(),
            f.stage.to_str().unwrap(),
        ],
        None,
    ))
}

fn install(exe: &Path, f: &Fixture) {
    let p = plan(exe, f);
    let digest = p["plan_digest"].as_str().unwrap().to_string();
    let r = ok(&run(
        exe,
        &f.home,
        &[
            "update-apply",
            f.home.to_str().unwrap(),
            f.stage.to_str().unwrap(),
            "--plan-digest",
            &digest,
            "--permission-reference",
            "fixture-approved",
        ],
        Some(&p),
    ));
    let h = r["hosts"]
        .as_array()
        .unwrap()
        .iter()
        .find(|h| h["host"] == HOST)
        .unwrap();
    assert_eq!(h["status"], json!("applied"), "{r}");
    // Receipts carry second-resolution timestamps; keep the order of installs observable.
    sleep(Duration::from_millis(1100));
}

fn host(v: &Value) -> &Value {
    v["hosts"]
        .as_array()
        .unwrap()
        .iter()
        .find(|h| h["host"] == HOST)
        .unwrap_or_else(|| panic!("{v}"))
}

/// The install id of a binary: the hash of its canonical path.
fn id_of(exe: &Path) -> String {
    format!(
        "{:x}",
        Sha256::digest(fs::canonicalize(exe).unwrap().to_str().unwrap().as_bytes())
    )
}

fn receipt_file(home: &Path, id: &str) -> PathBuf {
    state(home)
        .join("installations")
        .join(id)
        .join(format!("hosts/moai/{HOST}.json"))
}

fn set_version(home: &Path, id: &str, version: &str) {
    let path = receipt_file(home, id);
    let mut r: Value = serde_json::from_slice(&fs::read(&path).unwrap()).unwrap();
    r["package_version"] = json!(version);
    fs::write(&path, serde_json::to_vec_pretty(&r).unwrap()).unwrap();
}

/// The earlier folder `a`: a build run as `bin/moai` alone (so its receipts sit under the `moai`
/// name's id), then a `bin/moai-cli` of other bytes beside it. Nothing in `a` runs again.
fn earlier_folder(f: &Fixture) -> PathBuf {
    let moai = put(&f.home, "a", "moai");
    install(&moai, f);
    set_version(&f.home, &id_of(&moai), EARLIER);
    let cli = bin_dir(&f.home, "a").join("moai-cli");
    let mut bytes = fs::read(&moai).unwrap();
    bytes.push(0);
    fs::write(&cli, bytes).unwrap();
    moai
}

/// The running build `b`: `bin/moai-cli` and a hard link `bin/moai` to it; the plan runs as `moai`.
fn running_build(f: &Fixture) -> PathBuf {
    let cli = put(&f.home, "b", "moai-cli");
    let moai = cli.with_file_name("moai");
    fs::hard_link(&cli, &moai).unwrap();
    moai
}

// AC-1 (REQ-1, REQ-6): the receipt of the earlier folder sits under the `moai` name's id while
// the running build is `moai-cli`'s identity; the baseline is still found and nothing is adopted.
#[test]
fn a_receipt_under_the_moai_name_of_an_earlier_folder_is_the_baseline() {
    let f = fixture();
    earlier_folder(&f);
    let b = running_build(&f);
    let p = plan(&b, &f);
    assert_eq!(p["source_version"], json!(EARLIER), "{p}");
    let h = host(&p);
    assert_eq!(h["preflight"], json!("ok"), "{p}");
    assert!(h["receipt_from"].is_string(), "{p}");
    assert!(h.get("adopt").is_none(), "{p}");
}

// AC-2 (REQ-4): the header of the screen drawn from that plan names the earlier version.
#[test]
fn the_screen_header_names_the_earlier_version_and_never_none() {
    let f = fixture();
    earlier_folder(&f);
    let b = running_build(&f);
    let p = plan(&b, &f);
    let path =
        Path::new(env!("CARGO_MANIFEST_DIR")).join("tests/fixtures/update_tux/preview-latest.json");
    let mut result: Value = serde_json::from_str(&fs::read_to_string(path).unwrap()).unwrap();
    result["plan"] = p;
    let view = View {
        painter: Painter {
            color: false,
            unicode: true,
            lang: Lang::En,
        },
        width: 100,
        home: "/home/u".into(),
        project: None,
    };
    let text = screen(&result, Asked::NotAsked, &view);
    let line = text
        .lines()
        .find(|l| l.contains("installed"))
        .unwrap_or_else(|| panic!("{text}"));
    assert!(
        line.contains(&format!("installed {EARLIER} → package")),
        "{line}"
    );
    assert!(!line.contains("none"), "{line}");
}

// AC-3 (REQ-3): both names of one folder hold a finished receipt; the newer one is the baseline
// and the folder never ties with itself.
#[test]
fn of_both_names_of_one_folder_the_newer_receipt_is_the_baseline() {
    let f = fixture();
    // `bin/moai-cli` installs first; then its bytes change so that `bin/moai`, a copy of the
    // build, is its own identity and installs second: the newer receipt sits under the `moai`
    // name and carries `LATER`, the older one under `moai-cli` carries `EARLIER`.
    let cli = put(&f.home, "a", "moai-cli");
    install(&cli, &f);
    set_version(&f.home, &id_of(&cli), EARLIER);
    let mut bytes = fs::read(&cli).unwrap();
    bytes.push(0);
    fs::write(&cli, bytes).unwrap();
    let moai = put(&f.home, "a", "moai");
    install(&moai, &f);
    set_version(&f.home, &id_of(&moai), LATER);
    assert_ne!(id_of(&cli), id_of(&moai));
    let b = running_build(&f);
    let p = plan(&b, &f);
    assert_eq!(p["source_version"], json!(LATER), "{p}");
    assert_eq!(host(&p)["preflight"], json!("ok"), "{p}");
}

// AC-4 (REQ-2): an unreadable receipt under the `moai` name blocks the host and adopts nothing.
#[test]
fn an_unreadable_receipt_under_the_moai_name_blocks_the_host() {
    let f = fixture();
    let moai = earlier_folder(&f);
    fs::write(receipt_file(&f.home, &id_of(&moai)), b"{ not json").unwrap();
    let b = running_build(&f);
    let p = plan(&b, &f);
    let h = host(&p);
    assert_eq!(h["preflight"], json!("blocked"), "{p}");
    assert!(
        h["blocked_reason"]
            .as_str()
            .is_some_and(|r| r.contains("unreadable")),
        "{p}"
    );
    assert!(h.get("adopt").is_none(), "{p}");
}
