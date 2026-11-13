//! #2521: the decided layout installs each build in its own folder,
//! `~/.moai-cli/versions/<version>/bin/moai-cli` (decisions L165), and `install_id` comes
//! from that path. A newer build must keep updating the skill an earlier build in the same
//! layout installed, without weakening the unmanaged-folder protection (#2268/#2483).
//! Every case runs copies of the built binary inside a temporary HOME; the real HOME and
//! the installed binary are never touched.

#![cfg(unix)]

mod common;
use common::CleanEnv;

use serde_json::{json, Value};
use std::{
    fs,
    io::Write as _,
    os::unix::fs::symlink,
    path::{Path, PathBuf},
    process::{Command, Output, Stdio},
    thread::sleep,
    time::Duration,
};
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;
const HOSTS: [&str; 2] = ["claude", "codex"];

struct Fixture {
    _tmp: TempDir,
    root: PathBuf,
    home: PathBuf,
    v1: PathBuf,
    v2: PathBuf,
}

fn skill_md(body: &str) -> String {
    format!("---\nname: moai\ndescription: fixture package\n---\n# MoAI\n{body}\n")
}

fn write_pkg(dir: &Path, files: &[(&str, String)]) {
    for (rel, body) in files {
        let p = dir.join(rel);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        fs::write(p, body).unwrap();
    }
}

fn fixture() -> Fixture {
    let tmp = TempDir::new().unwrap();
    let root = fs::canonicalize(tmp.path()).unwrap();
    let home = root.join("home");
    fs::create_dir_all(&home).unwrap();
    let v1 = root.join("stage-v1");
    let v2 = root.join("stage-v2");
    write_pkg(
        &v1,
        &[
            ("SKILL.md", skill_md("body v1")),
            ("workflows/update.md", "# UPDATE\nv1\n".into()),
            // Unchanged between versions: only the adopted receipt can own it afterwards.
            ("references/routes.json", "{\"routes\":1}\n".into()),
        ],
    );
    write_pkg(
        &v2,
        &[
            ("SKILL.md", skill_md("body v2")),
            ("workflows/update.md", "# UPDATE\nv2\n".into()),
            // Unchanged between versions: only the adopted receipt can own it afterwards.
            ("references/routes.json", "{\"routes\":1}\n".into()),
            ("workflows/new.md", "# NEW\nv2\n".into()),
        ],
    );
    Fixture {
        _tmp: tmp,
        root,
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

fn install(exe: &Path, home: &Path, stage: &Path) {
    let p = plan(exe, home, stage);
    let r = apply(exe, home, stage, &p);
    for h in HOSTS {
        let o = r["hosts"]
            .as_array()
            .unwrap()
            .iter()
            .find(|x| x["host"] == h)
            .unwrap();
        assert_eq!(o["status"], json!("applied"), "{r}");
    }
    // Receipts carry second-resolution timestamps; keep the order of installs observable.
    sleep(Duration::from_millis(1100));
}

fn host<'a>(plan: &'a Value, name: &str) -> &'a Value {
    plan["hosts"]
        .as_array()
        .unwrap()
        .iter()
        .find(|h| h["host"] == name)
        .unwrap()
}

fn actions(plan: &Value, name: &str) -> Vec<(String, String)> {
    host(plan, name)["files"]
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

fn assert_ok_for_both(plan: &Value) {
    for h in HOSTS {
        assert_eq!(host(plan, h)["preflight"], json!("ok"), "{h}: {plan}");
    }
}

fn assert_blocked_unmanaged(plan: &Value) {
    for h in HOSTS {
        let hp = host(plan, h);
        assert_eq!(hp["preflight"], json!("blocked"), "{h}: {plan}");
        let reason = hp["blocked_reason"].as_str().unwrap();
        assert!(
            reason.contains("unmanaged existing package"),
            "{h}: {reason}"
        );
    }
}

fn installations(home: &Path) -> PathBuf {
    home.join(".moai-cli/installations")
}

fn receipt_dirs(home: &Path) -> Vec<String> {
    let mut v: Vec<String> = fs::read_dir(installations(home))
        .unwrap()
        .map(|e| e.unwrap().file_name().into_string().unwrap())
        .collect();
    v.sort();
    v
}

fn tree_bytes(dir: &Path) -> Vec<(PathBuf, Vec<u8>)> {
    let mut out = Vec::new();
    let mut stack = vec![dir.to_path_buf()];
    while let Some(d) = stack.pop() {
        for e in fs::read_dir(&d).unwrap() {
            let p = e.unwrap().path();
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

// AC-1 / AC-3: a skill the build in versions/A installed (receipts under A's per-version
// id, the state an existing HOME already has) is planned and applied `ok` by the build in
// versions/B; the next plan is a no-op, and A's receipts are left byte-for-byte alone.
#[test]
fn a_newer_version_folder_keeps_updating_the_skill() {
    let f = fixture();
    let a = version_exe(&f.home, "2026.11.0-rc.1+a");
    let b = version_exe(&f.home, "2026.11.0-rc.1+b");
    install(&a, &f.home, &f.v1);
    let [a_id]: [String; 1] = receipt_dirs(&f.home).try_into().unwrap();
    let a_before = tree_bytes(&installations(&f.home).join(&a_id));

    let p = plan(&b, &f.home, &f.v2);
    assert_ok_for_both(&p);
    for h in HOSTS {
        assert_eq!(host(&p, h)["receipt_from"], json!(a_id), "{h}: {p}");
        assert!(
            actions(&p, h).contains(&("SKILL.md".into(), "update".into())),
            "{p}"
        );
    }
    let r = apply(&b, &f.home, &f.v2, &p);
    for h in HOSTS {
        let o = r["hosts"]
            .as_array()
            .unwrap()
            .iter()
            .find(|x| x["host"] == h)
            .unwrap();
        assert_eq!(o["status"], json!("applied"), "{r}");
    }
    let again = plan(&b, &f.home, &f.v2);
    assert_ok_for_both(&again);
    for h in HOSTS {
        assert!(
            host(&again, h).get("receipt_from").is_none(),
            "own receipt now: {again}"
        );
        assert!(
            actions(&again, h).iter().all(|(_, a)| a == "unchanged"),
            "{h}: {again}"
        );
    }
    assert_eq!(receipt_dirs(&f.home).len(), 2);
    assert_eq!(tree_bytes(&installations(&f.home).join(&a_id)), a_before);
}

// A → B → A: the earlier build sees the newer state, not its own stale receipt.
#[test]
fn going_back_to_an_older_folder_reads_the_newest_receipt() {
    let f = fixture();
    let a = version_exe(&f.home, "a");
    let b = version_exe(&f.home, "b");
    install(&a, &f.home, &f.v1);
    install(&b, &f.home, &f.v2);
    let p = plan(&a, &f.home, &f.v2);
    assert_ok_for_both(&p);
    for h in HOSTS {
        assert!(
            actions(&p, h).iter().all(|(_, a)| a == "unchanged"),
            "{h}: {p}"
        );
    }
}

// AC-2: a build outside the layout neither adopts a layout receipt nor lends its own.
#[test]
fn a_build_outside_the_layout_stays_blocked_both_ways() {
    let f = fixture();
    let a = version_exe(&f.home, "a");
    install(&a, &f.home, &f.v1);
    let outside = f.root.join("elsewhere/bin/moai-cli");
    fs::create_dir_all(outside.parent().unwrap()).unwrap();
    common::copy_executable(Path::new(EXE), &outside);
    assert_blocked_unmanaged(&plan(&outside, &f.home, &f.v2));

    let g = fixture();
    let outside = g.root.join("elsewhere/bin/moai-cli");
    fs::create_dir_all(outside.parent().unwrap()).unwrap();
    common::copy_executable(Path::new(EXE), &outside);
    install(&outside, &g.home, &g.v1);
    let b = version_exe(&g.home, "b");
    assert_blocked_unmanaged(&plan(&b, &g.home, &g.v2));
}

// AC-2: a linked version folder or a linked receipt directory lends nothing.
#[test]
fn links_in_the_layout_or_the_receipts_lend_nothing() {
    let f = fixture();
    let a = version_exe(&f.home, "a");
    install(&a, &f.home, &f.v1);
    let [a_id]: [String; 1] = receipt_dirs(&f.home).try_into().unwrap();
    let b = version_exe(&f.home, "b");

    // versions/a replaced by a link to the same folder moved elsewhere.
    let versions = f.home.join(".moai-cli/versions");
    let moved = f.root.join("moved-a");
    fs::rename(versions.join("a"), &moved).unwrap();
    symlink(&moved, versions.join("a")).unwrap();
    assert_blocked_unmanaged(&plan(&b, &f.home, &f.v2));
    fs::remove_file(versions.join("a")).unwrap();
    fs::rename(&moved, versions.join("a")).unwrap();
    assert_ok_for_both(&plan(&b, &f.home, &f.v2));

    // installations/<a_id> replaced by a link to the same directory moved elsewhere.
    let inst = installations(&f.home);
    let moved = f.root.join("moved-receipts");
    fs::rename(inst.join(&a_id), &moved).unwrap();
    symlink(&moved, inst.join(&a_id)).unwrap();
    assert_blocked_unmanaged(&plan(&b, &f.home, &f.v2));
}

// AC-2: an earlier build's pending receipt is never adopted; the plan says how to recover.
#[test]
fn a_pending_receipt_of_an_earlier_folder_blocks_with_guidance() {
    let f = fixture();
    let a = version_exe(&f.home, "a");
    install(&a, &f.home, &f.v1);
    let [a_id]: [String; 1] = receipt_dirs(&f.home).try_into().unwrap();
    for h in HOSTS {
        let path = installations(&f.home)
            .join(&a_id)
            .join(format!("hosts/moai/{h}.json"));
        let mut v: Value = serde_json::from_slice(&fs::read(&path).unwrap()).unwrap();
        v["status"] = json!("pending");
        fs::write(&path, serde_json::to_vec_pretty(&v).unwrap()).unwrap();
    }
    let b = version_exe(&f.home, "b");
    let p = plan(&b, &f.home, &f.v2);
    for h in HOSTS {
        let hp = host(&p, h);
        assert_eq!(hp["preflight"], json!("blocked"), "{p}");
        let reason = hp["blocked_reason"].as_str().unwrap();
        assert!(
            reason.contains("pending") && reason.contains(&a_id),
            "{reason}"
        );
    }
}

// Two earlier folders: the receipt that finished last is the baseline.
#[test]
fn of_two_earlier_folders_the_latest_receipt_wins() {
    let f = fixture();
    let a = version_exe(&f.home, "a");
    let c = version_exe(&f.home, "c");
    install(&a, &f.home, &f.v1);
    let [a_id]: [String; 1] = receipt_dirs(&f.home).try_into().unwrap();
    install(&c, &f.home, &f.v2);
    let c_id = receipt_dirs(&f.home)
        .into_iter()
        .find(|id| *id != a_id)
        .unwrap();
    let b = version_exe(&f.home, "b");
    let p = plan(&b, &f.home, &f.v2);
    assert_ok_for_both(&p);
    for h in HOSTS {
        assert_eq!(host(&p, h)["receipt_from"], json!(c_id), "{p}");
        assert!(actions(&p, h).iter().all(|(_, a)| a == "unchanged"), "{p}");
    }
}

fn host_receipt(home: &Path, id: &str, h: &str) -> PathBuf {
    installations(home)
        .join(id)
        .join(format!("hosts/moai/{h}.json"))
}

fn edit_receipt(home: &Path, id: &str, h: &str, edit: impl Fn(&mut Value)) {
    let path = host_receipt(home, id, h);
    let mut v: Value = serde_json::from_slice(&fs::read(&path).unwrap()).unwrap();
    edit(&mut v);
    fs::write(&path, serde_json::to_vec_pretty(&v).unwrap()).unwrap();
}

fn assert_blocked_with(plan: &Value, h: &str, needle: &str) {
    let hp = host(plan, h);
    assert_eq!(hp["preflight"], json!("blocked"), "{h}: {plan}");
    let reason = hp["blocked_reason"].as_str().unwrap();
    assert!(reason.contains(needle), "{h}: {reason}");
    assert!(hp.get("receipt_from").is_none(), "{h}: {plan}");
}

// AC-2 (audit 6820396218 V7): a receipt that names another installation or another host
// (a copied file) lends nothing.
#[test]
fn a_receipt_naming_another_installation_or_host_lends_nothing() {
    let f = fixture();
    let a = version_exe(&f.home, "a");
    install(&a, &f.home, &f.v1);
    let [a_id]: [String; 1] = receipt_dirs(&f.home).try_into().unwrap();
    edit_receipt(&f.home, &a_id, "claude", |v| {
        v["install_id"] = json!("0".repeat(64))
    });
    edit_receipt(&f.home, &a_id, "codex", |v| v["host"] = json!("claude"));
    let b = version_exe(&f.home, "b");
    let p = plan(&b, &f.home, &f.v2);
    for h in HOSTS {
        assert_blocked_with(&p, h, "names another installation or host");
    }
}

// AC-2 (V10): an unreadable earlier receipt blocks instead of being skipped.
#[test]
fn an_unreadable_earlier_receipt_blocks() {
    let f = fixture();
    let a = version_exe(&f.home, "a");
    install(&a, &f.home, &f.v1);
    let [a_id]: [String; 1] = receipt_dirs(&f.home).try_into().unwrap();
    for h in HOSTS {
        fs::write(host_receipt(&f.home, &a_id, h), b"{ not json").unwrap();
    }
    let b = version_exe(&f.home, "b");
    let p = plan(&b, &f.home, &f.v2);
    for h in HOSTS {
        assert_blocked_with(&p, h, "is unreadable");
    }
}

// AC-2 (V4): two different earlier folders finishing in the same second are ambiguous.
#[test]
fn two_earlier_folders_finishing_in_the_same_second_block() {
    let f = fixture();
    let a = version_exe(&f.home, "a");
    let c = version_exe(&f.home, "c");
    install(&a, &f.home, &f.v1);
    let [a_id]: [String; 1] = receipt_dirs(&f.home).try_into().unwrap();
    install(&c, &f.home, &f.v2);
    let c_id = receipt_dirs(&f.home)
        .into_iter()
        .find(|id| *id != a_id)
        .unwrap();
    for h in HOSTS {
        let when: Value =
            serde_json::from_slice(&fs::read(host_receipt(&f.home, &c_id, h)).unwrap()).unwrap();
        edit_receipt(&f.home, &a_id, h, |v| {
            v["finished_at"] = when["finished_at"].clone()
        });
    }
    let b = version_exe(&f.home, "b");
    let p = plan(&b, &f.home, &f.v2);
    for h in HOSTS {
        assert_blocked_with(&p, h, "ambiguous");
    }
}

// AC-2 (V9): an earlier pending receipt that started in the same second the newest
// finished one ended is not ordered after it, so it still blocks.
#[test]
fn a_pending_receipt_in_the_same_second_as_the_newest_finished_one_blocks() {
    let f = fixture();
    let a = version_exe(&f.home, "a");
    let c = version_exe(&f.home, "c");
    install(&a, &f.home, &f.v1);
    let [a_id]: [String; 1] = receipt_dirs(&f.home).try_into().unwrap();
    install(&c, &f.home, &f.v2);
    let c_id = receipt_dirs(&f.home)
        .into_iter()
        .find(|id| *id != a_id)
        .unwrap();
    for h in HOSTS {
        let a_receipt: Value =
            serde_json::from_slice(&fs::read(host_receipt(&f.home, &a_id, h)).unwrap()).unwrap();
        edit_receipt(&f.home, &c_id, h, |v| {
            v["status"] = json!("pending");
            v["started_at"] = a_receipt["finished_at"].clone();
            v["finished_at"] = Value::Null;
        });
    }
    let b = version_exe(&f.home, "b");
    let p = plan(&b, &f.home, &f.v2);
    for h in HOSTS {
        assert_blocked_with(&p, h, "is pending");
    }
}

// AC-2 (V1): a binary inside a version folder but not in its `bin/` is outside the layout,
// so it adopts nothing.
#[test]
fn a_binary_outside_the_bin_folder_of_a_version_adopts_nothing() {
    let f = fixture();
    let a = version_exe(&f.home, "a");
    install(&a, &f.home, &f.v1);
    let libexec = f.home.join(".moai-cli/versions/b/libexec");
    fs::create_dir_all(&libexec).unwrap();
    let b = libexec.join("moai-cli");
    common::copy_executable(Path::new(EXE), &b);
    assert_blocked_unmanaged(&plan(&b, &f.home, &f.v2));
}

// #2543: a version folder whose name is not UTF-8 lends nothing and never fails the plan; the
// earlier folder's receipt is still adopted. Linux only: macOS APFS refuses such a name
// (EILSEQ), so there the unit test `update::version_folder_tests` covers the rule.
#[cfg(target_os = "linux")]
#[test]
fn a_version_folder_named_in_bytes_that_are_not_utf8_lends_nothing() {
    use std::os::unix::ffi::OsStrExt;
    let f = fixture();
    let a = version_exe(&f.home, "2026.11.0-rc.1+a");
    install(&a, &f.home, &f.v1);
    let [a_id]: [String; 1] = receipt_dirs(&f.home).try_into().unwrap();
    let bad = f
        .home
        .join(".moai-cli/versions")
        .join(std::ffi::OsStr::from_bytes(b"\xff\xfe-not-utf8"))
        .join("bin");
    fs::create_dir_all(&bad).unwrap();
    common::copy_executable(Path::new(EXE), &bad.join("moai-cli"));
    let b = version_exe(&f.home, "2026.11.0-rc.1+b");
    let p = plan(&b, &f.home, &f.v2);
    assert_ok_for_both(&p);
    for h in HOSTS {
        assert_eq!(host(&p, h)["receipt_from"], json!(a_id), "{h}: {p}");
    }
}
