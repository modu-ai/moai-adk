//! SPEC-20261127-installed-binary-name-moai AC-7 (REQ-7): a new install holds `bin/moai` alone, so
//! its installation id is the hash of that path. The receipt of an earlier version folder, which
//! sits under the id of the `moai-cli` path of a hard-linked folder, must still be found as the
//! update baseline, and the screen must never say `installed none`. Every case runs copies of
//! the built binaries as child processes inside a temporary HOME; the real HOME is never touched.

#![cfg(unix)]

mod common;
use common::CleanEnv;

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

const CLI: &str = common::MOAI_CLI_BIN;
const MOAI: &str = common::MOAI_BIN;
const HOST: &str = "claude";
/// The version the earlier folder's receipt is rewritten to, as if an earlier build wrote it.
const EARLIER: &str = "2026.11.0-rc.2+local.4";

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

/// The state folder this generation uses in `home`, found through the resolver.
fn state(home: &Path) -> PathBuf {
    moai_cli::home_root::path(home)
}

fn bin_dir(home: &Path, version: &str) -> PathBuf {
    let bin = state(home).join("versions").join(version).join("bin");
    fs::create_dir_all(&bin).unwrap();
    bin
}

/// A copy of the built binary `source` at `versions/<version>/bin/<name>`.
fn put(home: &Path, version: &str, name: &str, source: &str) -> PathBuf {
    let exe = bin_dir(home, version).join(name);
    common::copy_executable(Path::new(source), &exe);
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

/// Plans the fixture's package as `exe` and applies it, so a receipt of `exe`'s id exists.
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

/// The earlier folder `a`: `bin/moai-cli` and a hard link `bin/moai`, installed as `moai-cli`, so
/// its receipt sits under the `moai-cli` path's id. Returns that id.
fn earlier_hard_linked_folder(f: &Fixture) -> String {
    let cli = put(&f.home, "a", "moai-cli", CLI);
    fs::hard_link(&cli, cli.with_file_name("moai")).unwrap();
    install(&cli, f);
    let id = id_of(&cli);
    set_version(&f.home, &id, EARLIER);
    id
}

// AC-7 (REQ-7): the new folder `b` holds `bin/moai` alone. A plan run as that `moai` finds the
// receipt of the hard-linked earlier folder and reports its version as the baseline.
#[test]
fn ac7_a_new_folder_with_moai_alone_finds_the_earlier_receipt() {
    let f = fixture();
    let earlier_id = earlier_hard_linked_folder(&f);
    let earlier_bytes = fs::read(receipt_file(&f.home, &earlier_id)).unwrap();

    let moai = put(&f.home, "b", "moai", MOAI);
    assert!(
        !moai.with_file_name("moai-cli").exists(),
        "the new folder must hold bin/moai alone"
    );
    let p = plan(&moai, &f);

    assert_eq!(p["source_version"], json!(EARLIER), "{p}");
    let h = host(&p);
    assert_eq!(h["preflight"], json!("ok"), "{p}");
    assert!(h["receipt_from"].is_string(), "{p}");
    assert!(h.get("adopt").is_none(), "{p}");
    assert_eq!(
        fs::read(receipt_file(&f.home, &earlier_id)).unwrap(),
        earlier_bytes,
        "the plan changed the earlier receipt"
    );
    // The new folder's own id is the hash of `bin/moai`, never of a `moai-cli` it does not hold.
    assert_ne!(id_of(&moai), earlier_id);
}
