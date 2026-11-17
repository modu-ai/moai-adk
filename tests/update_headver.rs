//! Card C-67 slice 1 (SPEC-20261117-headver-version AC-6): `update-plan` records the installed
//! receipt's version as the source and this build's HeadVer version as the target. An install
//! of the earlier `2026.11.0-rc.N[+local.n]` scheme reads as older, so this build updates it
//! cleanly; an install newer than this build blocks the host, so a build never updates back.
//! Every case runs copies of the built binary inside a temporary HOME; the real HOME and the
//! installed binary are never touched.

#![cfg(unix)]

mod common;
use common::CleanEnv;

use serde_json::{json, Value};
use std::{
    fs,
    io::Write as _,
    path::{Path, PathBuf},
    process::{Command, Output, Stdio},
};
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;
const HOSTS: [&str; 2] = ["claude", "codex"];
const EARLIER: &str = "2026.11.0-rc.2+local.4";

struct Fixture {
    _tmp: TempDir,
    home: PathBuf,
    v1: PathBuf,
    v2: PathBuf,
}

fn write_pkg(dir: &Path, body: &str) {
    for (rel, text) in [
        (
            "SKILL.md",
            format!("---\nname: moai\ndescription: fixture package\n---\n# MoAI\n{body}\n"),
        ),
        ("workflows/update.md", format!("# UPDATE\n{body}\n")),
    ] {
        let p = dir.join(rel);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        fs::write(p, text).unwrap();
    }
}

fn fixture() -> Fixture {
    let tmp = TempDir::new().unwrap();
    let root = fs::canonicalize(tmp.path()).unwrap();
    let home = root.join("home");
    fs::create_dir_all(&home).unwrap();
    let (v1, v2) = (root.join("stage-v1"), root.join("stage-v2"));
    write_pkg(&v1, "v1");
    write_pkg(&v2, "v2");
    Fixture {
        _tmp: tmp,
        home,
        v1,
        v2,
    }
}

/// A copy of the built binary at `~/.moai-cli/versions/<version>/bin/moai-cli`.
fn version_exe(home: &Path, version: &str) -> PathBuf {
    let bin = home.join(".moai-cli/versions").join(version).join("bin");
    fs::create_dir_all(&bin).unwrap();
    let exe = bin.join("moai-cli");
    common::copy_executable(Path::new(EXE), &exe);
    exe
}

fn run(exe: &Path, args: &[&str], stdin: Option<&Value>) -> Output {
    let mut child = Command::new(exe)
        .args(args)
        .clean_env()
        .env("PATH", "/usr/bin:/bin")
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

fn plan(exe: &Path, home: &Path, stage: &Path) -> Value {
    ok(&run(
        exe,
        &[
            "update-plan",
            home.to_str().unwrap(),
            stage.to_str().unwrap(),
        ],
        None,
    ))
}

fn apply(exe: &Path, home: &Path, stage: &Path, plan: &Value) -> Value {
    let digest = plan["plan_digest"].as_str().unwrap().to_string();
    ok(&run(
        exe,
        &[
            "update-apply",
            home.to_str().unwrap(),
            stage.to_str().unwrap(),
            "--plan-digest",
            &digest,
            "--permission-reference",
            "fixture-approved",
        ],
        Some(plan),
    ))
}

fn host<'a>(v: &'a Value, name: &str) -> &'a Value {
    v["hosts"]
        .as_array()
        .unwrap()
        .iter()
        .find(|h| h["host"] == name)
        .unwrap_or_else(|| panic!("{name}: {v}"))
}

/// Every host receipt under `~/.moai-cli/installations/`.
fn receipt_files(home: &Path) -> Vec<PathBuf> {
    let mut out = Vec::new();
    let mut stack = vec![home.join(".moai-cli/installations")];
    while let Some(d) = stack.pop() {
        for e in fs::read_dir(&d).unwrap() {
            let p = e.unwrap().path();
            if p.is_dir() {
                stack.push(p);
            } else if HOSTS.iter().any(|h| {
                p.file_name()
                    .is_some_and(|n| n == format!("{h}.json").as_str())
            }) {
                out.push(p);
            }
        }
    }
    out.sort();
    out
}

/// Rewrites every host receipt's `package_version`, as if another build had installed it.
fn set_receipt_versions(home: &Path, version: &str) {
    let files = receipt_files(home);
    assert_eq!(files.len(), HOSTS.len(), "{files:?}");
    for p in files {
        let mut r: Value = serde_json::from_slice(&fs::read(&p).unwrap()).unwrap();
        r["package_version"] = json!(version);
        fs::write(&p, serde_json::to_vec_pretty(&r).unwrap()).unwrap();
    }
}

/// The installed skill files of both hosts.
fn installed(home: &Path) -> Vec<(PathBuf, Vec<u8>)> {
    let mut out = Vec::new();
    for top in [".claude", ".agents"] {
        let mut stack = vec![home.join(top)];
        while let Some(d) = stack.pop() {
            let Ok(rd) = fs::read_dir(&d) else { continue };
            for e in rd {
                let p = e.unwrap().path();
                if p.is_dir() {
                    stack.push(p);
                } else {
                    out.push((p.clone(), fs::read(&p).unwrap()));
                }
            }
        }
    }
    out.sort();
    out
}

#[test]
fn an_earlier_rc_install_reads_older_and_this_build_updates_it() {
    let f = fixture();
    let rc = version_exe(&f.home, EARLIER);
    let p = plan(&rc, &f.home, &f.v1);
    apply(&rc, &f.home, &f.v1, &p);
    set_receipt_versions(&f.home, EARLIER);

    let me = version_exe(&f.home, moai_cli::BUILD_VERSION);
    let p = plan(&me, &f.home, &f.v2);
    assert_eq!(p["source_version"], json!(EARLIER), "{p}");
    assert_eq!(p["target_version"], json!(moai_cli::BUILD_VERSION), "{p}");
    for h in HOSTS {
        assert_eq!(host(&p, h)["preflight"], json!("ok"), "{h}: {p}");
    }
    let r = apply(&me, &f.home, &f.v2, &p);
    for h in HOSTS {
        assert_eq!(host(&r, h)["status"], json!("applied"), "{h}: {r}");
    }
    // The new receipts carry this build's version and name the earlier one it replaced.
    for path in receipt_files(&f.home) {
        let receipt: Value = serde_json::from_slice(&fs::read(&path).unwrap()).unwrap();
        if receipt["package_version"] == json!(moai_cli::BUILD_VERSION) {
            assert_eq!(receipt["version_before"], json!(EARLIER), "{receipt}");
        }
    }
    let again = plan(&me, &f.home, &f.v2);
    assert_eq!(
        again["source_version"],
        json!(moai_cli::BUILD_VERSION),
        "{again}"
    );
}

#[test]
fn an_install_newer_than_this_build_is_never_updated_back() {
    let f = fixture();
    let me = version_exe(&f.home, moai_cli::BUILD_VERSION);
    let p = plan(&me, &f.home, &f.v1);
    apply(&me, &f.home, &f.v1, &p);
    // One head above this build: newer whatever week and build this build has.
    let head: u32 = moai_cli::BUILD_VERSION
        .split('.')
        .next()
        .unwrap()
        .parse()
        .unwrap();
    let newer = format!("{}.0001.1", head + 1);
    set_receipt_versions(&f.home, &newer);
    let before = installed(&f.home);

    let p = plan(&me, &f.home, &f.v2);
    assert_eq!(p["source_version"], json!(newer), "{p}");
    for h in HOSTS {
        let hp = host(&p, h);
        assert_eq!(hp["preflight"], json!("blocked"), "{h}: {p}");
        let reason = hp["blocked_reason"].as_str().unwrap();
        assert!(
            reason.contains(&newer) && reason.contains("never updates back"),
            "{h}: {reason}"
        );
    }
    // The apply either refuses or applies no host; either way nothing changes.
    let digest = p["plan_digest"].as_str().unwrap().to_string();
    let out = run(
        &me,
        &[
            "update-apply",
            f.home.to_str().unwrap(),
            f.v2.to_str().unwrap(),
            "--plan-digest",
            &digest,
            "--permission-reference",
            "fixture-approved",
        ],
        Some(&p),
    );
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    if v["ok"] == json!(true) {
        for h in HOSTS {
            assert_ne!(
                host(&v["result"], h)["status"],
                json!("applied"),
                "{h}: {v}"
            );
        }
    }
    assert_eq!(installed(&f.home), before, "nothing was updated back");
    // A receipt of a shape this build does not read is not a reason to block.
    set_receipt_versions(&f.home, "3.2.1");
    let p = plan(&me, &f.home, &f.v2);
    for h in HOSTS {
        assert_eq!(host(&p, h)["preflight"], json!("ok"), "{h}: {p}");
    }
}
