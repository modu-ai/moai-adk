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
const HOSTS: [&str; 1] = ["claude"];
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

/// Runs a copy at `<home>/.moai-cli/versions/<v>/bin/moai-cli` with that temporary HOME.
fn run(exe: &Path, args: &[&str], stdin: Option<&Value>) -> Output {
    let home = exe
        .ancestors()
        .nth(5)
        .expect("the copy sits under a temporary HOME");
    let mut child = Command::new(exe)
        .args(args)
        .clean_env()
        .env("HOME", home)
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

/// Every skill host receipt under `~/.moai-cli/installations/<id>/hosts/moai/` (the other
/// layers' receipts are not the version the update compares).
fn receipt_files(home: &Path) -> Vec<PathBuf> {
    let mut out = Vec::new();
    let mut stack = vec![home.join(".moai-cli/installations")];
    while let Some(d) = stack.pop() {
        for e in fs::read_dir(&d).unwrap() {
            let p = e.unwrap().path();
            if p.is_dir() {
                stack.push(p);
            } else if p.parent().is_some_and(|d| d.ends_with("hosts/moai"))
                && HOSTS.iter().any(|h| {
                    p.file_name()
                        .is_some_and(|n| n == format!("{h}.json").as_str())
                })
            {
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
    // Audit N5: the same version installed again is not "newer", so the re-plan is not blocked.
    for h in HOSTS {
        assert_eq!(host(&again, h)["preflight"], json!("ok"), "{h}: {again}");
        assert!(host(&again, h)["blocked_reason"].is_null(), "{h}: {again}");
    }
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
            reason.contains(&newer)
                && reason.contains("never updates back")
                && reason.contains("moai update --allow-older"),
            "{h}: {reason}"
        );
        assert!(hp["going_back"].is_null(), "{h}: {p}");
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

/// `moai-cli update` from the copy's own folder (its package and agents beside it), with the
/// temporary HOME as the current folder too, so no project outside it is read.
fn update(exe: &Path, home: &Path, args: &[&str]) -> Value {
    let out = Command::new(exe)
        .arg("update")
        .args(args)
        .clean_env()
        .env("HOME", home)
        .env("PATH", "/usr/bin:/bin")
        .env("LC_ALL", "C")
        .current_dir(home)
        .stdin(Stdio::null())
        .output()
        .unwrap();
    serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
        panic!(
            "{}{}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    })
}

/// Audit N1 (leader decision, #2521 A → B → A): going back to an older build is blocked by
/// default; `moai-cli update --allow-older` is the explicit way back. Its preview says so in a
/// warning per host, its digest binds the choice (a confirm without the flag is refused as
/// stale) and only the confirmed run with the flag writes.
#[test]
fn going_back_to_an_older_build_needs_allow_older() {
    let f = fixture();
    let me = version_exe(&f.home, moai_cli::BUILD_VERSION);
    let folder = me.parent().unwrap().parent().unwrap().to_path_buf();
    let repo = Path::new(env!("CARGO_MANIFEST_DIR"));
    copy_tree(&repo.join("skills/moai"), &folder.join("skills/moai"));
    copy_tree(&repo.join("agents"), &folder.join("agents"));
    let first = update(&me, &f.home, &[]);
    let digest = first["result"]["plan_digest"].as_str().unwrap().to_string();
    let applied = update(&me, &f.home, &["--confirm", &digest]);
    assert_eq!(applied["ok"], json!(true), "{applied}");

    // A newer build installed since: one head above this build.
    let head: u32 = moai_cli::BUILD_VERSION
        .split('.')
        .next()
        .unwrap()
        .parse()
        .unwrap();
    let newer = format!("{}.0001.1", head + 1);
    set_receipt_versions(&f.home, &newer);
    // A changed skill file, so going back has something to write.
    let skill = folder.join("skills/moai/SKILL.md");
    let mut text = fs::read_to_string(&skill).unwrap();
    text.push_str("\nolder build\n");
    fs::write(&skill, text).unwrap();
    let before = installed(&f.home);

    // Default: blocked, and the reason names the way back.
    let p = update(&me, &f.home, &[]);
    for h in HOSTS {
        let hp = host(&p["result"]["plan"], h);
        assert_eq!(hp["preflight"], json!("blocked"), "{h}: {p}");
        assert!(
            hp["blocked_reason"]
                .as_str()
                .unwrap()
                .contains("moai update --allow-older"),
            "{h}: {p}"
        );
    }

    // With the flag: planned, with a warning line per host and a confirm line that keeps it.
    let p = update(&me, &f.home, &["--allow-older"]);
    let r = &p["result"];
    for h in HOSTS {
        let hp = host(&r["plan"], h);
        assert_eq!(hp["preflight"], json!("ok"), "{h}: {p}");
        let warning = hp["going_back"].as_str().unwrap_or_default();
        assert!(
            warning.contains(&newer) && warning.contains(moai_cli::BUILD_VERSION),
            "{h}: {p}"
        );
    }
    let digest = r["plan_digest"].as_str().unwrap().to_string();
    assert_eq!(
        r["confirm_with"],
        json!(format!("moai update --allow-older --confirm {digest}")),
        "{p}"
    );
    assert_eq!(installed(&f.home), before, "a preview writes nothing");

    // The digest binds the choice: confirmed without the flag, it is stale.
    let refused = update(&me, &f.home, &["--confirm", &digest]);
    assert_eq!(refused["ok"], json!(false), "{refused}");
    assert_eq!(installed(&f.home), before, "nothing was updated back");

    let back = update(&me, &f.home, &["--allow-older", "--confirm", &digest]);
    assert_eq!(back["ok"], json!(true), "{back}");
    assert_ne!(installed(&f.home), before, "the older build was installed");
    for path in receipt_files(&f.home) {
        let receipt: Value = serde_json::from_slice(&fs::read(&path).unwrap()).unwrap();
        if receipt.get("package_version").is_some() {
            assert_eq!(
                receipt["package_version"],
                json!(moai_cli::BUILD_VERSION),
                "{receipt}"
            );
        }
    }
}

/// This build's `(head, yearweek, build)` from `BUILD_VERSION`.
fn build_parts() -> (u32, u32, u32) {
    let mut parts = moai_cli::BUILD_VERSION
        .split('.')
        .map(|p| p.parse::<u32>().unwrap());
    (
        parts.next().unwrap(),
        parts.next().unwrap(),
        parts.next().unwrap(),
    )
}

/// SPEC-20261119-headver-release-workflow D-6 (leader decision, card C-93 note 2260): only the
/// numbers `(head, yearweek, build)` order a local build and a release. A release counter
/// starts at 1 and a local one is already higher, so a release of this build's own week with a
/// lower number reads as older: this build updates over it without `--allow-older`. The
/// install record holds the version text only, so `update` cannot tell the two kinds apart.
#[test]
fn a_build_of_the_same_week_with_a_higher_number_reads_newer_than_the_installed_release() {
    let (head, week, build) = build_parts();
    let f = fixture();
    let me = version_exe(&f.home, moai_cli::BUILD_VERSION);
    let p = plan(&me, &f.home, &f.v1);
    apply(&me, &f.home, &f.v1, &p);
    if build < 2 {
        // Card C-93 follow-up N3: with no build number (a local run or a CI re-run) or the
        // first-ever run's 1, no release of this week sits below this build — the release
        // {head}.{week}.{build + 1} reads newer, and the plan says so with `going_back`. This
        // used to be a silent early return.
        let release = format!("{head}.{week:04}.{}", build + 1);
        set_receipt_versions(&f.home, &release);
        let p = plan(&me, &f.home, &f.v2);
        assert_eq!(p["source_version"], json!(release), "{p}");
        for h in HOSTS {
            let hp = host(&p, h);
            assert_eq!(hp["preflight"], json!("blocked"), "{h}: {p}");
            let reason = hp["blocked_reason"].as_str().unwrap();
            assert!(
                reason.contains(&release) && reason.contains("moai update --allow-older"),
                "{h}: {reason}"
            );
        }
        return;
    }
    let release = format!("{head}.{week:04}.{}", build - 1);
    set_receipt_versions(&f.home, &release);

    let p = plan(&me, &f.home, &f.v2);
    assert_eq!(p["source_version"], json!(release), "{p}");
    assert_eq!(p["target_version"], json!(moai_cli::BUILD_VERSION), "{p}");
    for h in HOSTS {
        let hp = host(&p, h);
        assert_eq!(hp["preflight"], json!("ok"), "{h}: {p}");
        assert!(hp["blocked_reason"].is_null(), "{h}: {p}");
        assert!(hp["going_back"].is_null(), "{h}: {p}");
    }
}

/// The other side of the same order: an installed release of this build's own week with a
/// higher number is newer, so this build updates back over it only with `--allow-older`.
#[test]
fn a_build_below_the_installed_release_of_the_same_week_needs_allow_older() {
    let (head, week, build) = build_parts();
    let f = fixture();
    let me = version_exe(&f.home, moai_cli::BUILD_VERSION);
    let p = plan(&me, &f.home, &f.v1);
    apply(&me, &f.home, &f.v1, &p);
    let release = format!("{head}.{week:04}.{}", build + 1);
    set_receipt_versions(&f.home, &release);

    let p = plan(&me, &f.home, &f.v2);
    assert_eq!(p["source_version"], json!(release), "{p}");
    for h in HOSTS {
        let hp = host(&p, h);
        assert_eq!(hp["preflight"], json!("blocked"), "{h}: {p}");
        let reason = hp["blocked_reason"].as_str().unwrap();
        assert!(
            reason.contains(&release) && reason.contains("moai update --allow-older"),
            "{h}: {reason}"
        );
    }
}
