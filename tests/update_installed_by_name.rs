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
    os::unix::fs::PermissionsExt as _,
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

/// Plans the fixture's package as `exe` and applies it; the raw output, whatever it says.
fn apply_raw(exe: &Path, f: &Fixture) -> Output {
    let p = plan(exe, f);
    let digest = p["plan_digest"].as_str().unwrap().to_string();
    run(
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
    )
}

/// Applies a package that writes no file for the host: the run succeeds and the host's status is
/// `status` (`unchanged`, or `conflict` for a host whose files are in conflict).
fn apply_unchanged(exe: &Path, f: &Fixture, status: &str) -> Value {
    let r = ok(&apply_raw(exe, f));
    assert_eq!(host(&r)["status"], json!(status), "{r}");
    r
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

fn set_status(home: &Path, id: &str, status: &str) {
    let path = receipt_file(home, id);
    let mut r: Value = serde_json::from_slice(&fs::read(&path).unwrap()).unwrap();
    r["status"] = json!(status);
    fs::write(&path, serde_json::to_vec_pretty(&r).unwrap()).unwrap();
}

fn receipt(home: &Path, id: &str) -> Value {
    serde_json::from_slice(&fs::read(receipt_file(home, id)).unwrap()).unwrap()
}

/// The host's projected `SKILL.md` under the temporary HOME.
fn skill_file(home: &Path) -> PathBuf {
    home.join(".claude/skills/moai/SKILL.md")
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

/// The header line (the one that names the installed version) of the screen drawn from the
/// `preview-latest.json` fixture with `edit` applied to it.
fn header_line(lang: Lang, edit: impl FnOnce(&mut Value)) -> String {
    let path =
        Path::new(env!("CARGO_MANIFEST_DIR")).join("tests/fixtures/update_tux/preview-latest.json");
    let mut result: Value = serde_json::from_str(&fs::read_to_string(path).unwrap()).unwrap();
    edit(&mut result);
    let view = View {
        painter: Painter {
            color: false,
            unicode: true,
            lang,
        },
        width: 100,
        home: "/home/u".into(),
        project: None,
    };
    let text = screen(&result, Asked::NotAsked, &view);
    let marker = match lang {
        Lang::En => "installed",
        Lang::Ko => "설치본",
    };
    text.lines()
        .find(|l| l.contains(marker))
        .unwrap_or_else(|| panic!("{text}"))
        .to_string()
}

// SPEC-20261125-update-installed-from-link AC-3 (REQ-3, REQ-9): the preview's `installed_version`
// is the installed side of the header in English and in Korean; without it the header keeps the
// receipt's source version.
#[test]
fn the_header_prefers_the_link_version_in_english_and_korean() {
    let link = "4.2648.26";
    let en = header_line(Lang::En, |r| r["installed_version"] = json!(link));
    assert!(en.contains(&format!("installed {link} → package")), "{en}");
    let ko = header_line(Lang::Ko, |r| r["installed_version"] = json!(link));
    assert!(ko.contains(&format!("설치본 {link} → 패키지")), "{ko}");

    // Without the key the receipt's source version stays, as the neighbouring test pins.
    let source = "2026.11.0-rc.1";
    let en = header_line(Lang::En, |_| {});
    assert!(
        en.contains(&format!("installed {source} = package")),
        "{en}"
    );
    assert!(!en.contains(link), "{en}");
    let ko = header_line(Lang::Ko, |_| {});
    assert!(ko.contains(&format!("설치본 {source} = 패키지")), "{ko}");

    // A receipt-less plan (no source version) with the key still shows the link's version.
    let none = header_line(Lang::En, |r| {
        r["plan"] = json!({"target_version": source});
        r["installed_version"] = json!(link);
    });
    assert!(
        none.contains(&format!("installed {link} → package")),
        "{none}"
    );
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

// SPEC-20261125-update-installed-from-link AC-4 (REQ-4, REQ-6, REQ-7): an apply that writes no file
// for a host leaves this build's version in the host's receipt, with the owned list, the entries
// and the skill bytes as they were. A `recovered` receipt counts like a `committed` one.
#[test]
fn a_no_write_apply_leaves_this_builds_version_in_the_receipt() {
    let f = fixture();
    let exe = put(&f.home, "a", "moai");
    install(&exe, &f);
    let id = id_of(&exe);
    let first = receipt(&f.home, &id);
    let build = first["package_version"].as_str().unwrap().to_string();
    assert_ne!(build, EARLIER);
    let skill = fs::read(skill_file(&f.home)).unwrap();

    for status in ["committed", "recovered"] {
        set_version(&f.home, &id, EARLIER);
        set_status(&f.home, &id, status);
        let r = apply_unchanged(&exe, &f, "unchanged");
        assert_eq!(r["status"], json!("unchanged"), "{status}: {r}");
        assert_eq!(host(&r)["receipt_status"], json!("committed"), "{r}");
        let after = receipt(&f.home, &id);
        assert_eq!(after["package_version"], json!(build), "{status}: {after}");
        assert_eq!(after["version_before"], json!(EARLIER), "{status}: {after}");
        assert_eq!(after["status"], json!("committed"), "{status}: {after}");
        assert_eq!(after["operation"], json!("update.apply"), "{after}");
        assert!(after["finished_at"].is_string(), "{after}");
        assert_eq!(after["owned"], first["owned"], "{status}: {after}");
        assert_eq!(after["entries"], json!([]), "{status}: {after}");
        assert!(after.get("backup").is_none_or(Value::is_null), "{after}");
        assert_eq!(fs::read(skill_file(&f.home)).unwrap(), skill, "{status}");
    }
}

// AC-4 (REQ-4, D4): a host whose only difference is a user-edited owned file writes no file either;
// its receipt takes this build's version and the user's bytes stay.
#[test]
fn a_conflict_only_no_write_apply_also_refreshes_the_receipt_version() {
    let f = fixture();
    let exe = put(&f.home, "a", "moai");
    install(&exe, &f);
    let id = id_of(&exe);
    let first = receipt(&f.home, &id);
    let build = first["package_version"].as_str().unwrap().to_string();
    // The user edits an owned file and the package moves that same file on: a conflict.
    let edited = b"---\nname: moai\ndescription: edited by the user\n---\n";
    fs::write(skill_file(&f.home), edited).unwrap();
    fs::write(
        f.stage.join("SKILL.md"),
        "---\nname: moai\ndescription: fixture package\n---\n# MoAI\nbody v2\n",
    )
    .unwrap();
    set_version(&f.home, &id, EARLIER);

    let r = apply_unchanged(&exe, &f, "conflict");
    assert_eq!(r["status"], json!("conflict"), "{r}");
    assert_eq!(fs::read(skill_file(&f.home)).unwrap(), edited);
    let after = receipt(&f.home, &id);
    assert_eq!(after["package_version"], json!(build), "{after}");
    assert_eq!(after["version_before"], json!(EARLIER), "{after}");
    assert_eq!(after["status"], json!("committed"), "{after}");
    assert_eq!(after["owned"], first["owned"], "{after}");
}

// AC-4 (REQ-5): once the receipt names this build, the next no-write apply writes no receipt.
#[test]
fn a_second_no_write_apply_writes_no_receipt() {
    let f = fixture();
    let exe = put(&f.home, "a", "moai");
    install(&exe, &f);
    let id = id_of(&exe);
    set_version(&f.home, &id, EARLIER);
    apply_unchanged(&exe, &f, "unchanged");
    let written = fs::read(receipt_file(&f.home, &id)).unwrap();
    let finished = receipt(&f.home, &id)["finished_at"].clone();
    assert!(finished.is_string(), "{finished}");

    // Receipts carry second-resolution timestamps: a rewrite would change `finished_at`.
    sleep(Duration::from_millis(1100));
    apply_unchanged(&exe, &f, "unchanged");
    assert_eq!(fs::read(receipt_file(&f.home, &id)).unwrap(), written);
    assert_eq!(receipt(&f.home, &id)["finished_at"], finished);
}

// AC-4 (REQ-4, REQ-6): the reported case. A new version folder has no receipt of its own and its
// apply writes no file: the earlier folder's receipt is the baseline, this build's version goes
// into a receipt under the running build's installation id, and the earlier receipt stays.
#[test]
fn a_no_write_apply_over_an_earlier_folders_receipt_writes_this_installations_receipt() {
    let f = fixture();
    let earlier_id = id_of(&earlier_folder(&f));
    let b = running_build(&f);
    // `moai` beside the same build's `moai-cli` writes under the canonical name's id.
    let own_id = id_of(&b.with_file_name("moai-cli"));
    assert!(!receipt_file(&f.home, &own_id).exists());
    let earlier = receipt(&f.home, &earlier_id);
    let earlier_bytes = fs::read(receipt_file(&f.home, &earlier_id)).unwrap();
    let skill = fs::read(skill_file(&f.home)).unwrap();

    let r = apply_unchanged(&b, &f, "unchanged");

    assert_eq!(host(&r)["receipt_status"], json!("committed"), "{r}");
    let after = receipt(&f.home, &own_id);
    assert_eq!(after["install_id"], json!(own_id), "{after}");
    assert_eq!(
        after["package_version"],
        json!(moai_cli::BUILD_VERSION),
        "{after}"
    );
    assert_eq!(after["version_before"], json!(EARLIER), "{after}");
    assert_eq!(after["status"], json!("committed"), "{after}");
    assert_eq!(after["owned"], earlier["owned"], "{after}");
    assert_eq!(
        fs::read(receipt_file(&f.home, &earlier_id)).unwrap(),
        earlier_bytes
    );
    assert_eq!(fs::read(skill_file(&f.home)).unwrap(), skill);
}

/// Gives a folder its mode back when the test ends, even when an assertion failed first.
struct ModeBack(PathBuf, u32);

impl Drop for ModeBack {
    fn drop(&mut self) {
        let _ = fs::set_permissions(&self.0, fs::Permissions::from_mode(self.1));
    }
}

// AC-5 (REQ-5, REQ-8): a pending host is blocked and writes no receipt; a receipt that is neither
// committed nor recovered is left as it is; a failing save is an error, never `unchanged`.
#[test]
fn a_blocked_or_pending_host_writes_no_version_only_receipt() {
    let f = fixture();
    let exe = put(&f.home, "a", "moai");
    install(&exe, &f);
    let id = id_of(&exe);
    let file = receipt_file(&f.home, &id);

    // A pending receipt blocks the host in the plan, so the apply never reaches a receipt write.
    set_version(&f.home, &id, EARLIER);
    set_status(&f.home, &id, "pending");
    let before = fs::read(&file).unwrap();
    assert_eq!(host(&plan(&exe, &f))["preflight"], json!("blocked"));
    let out = apply_raw(&exe, &f);
    assert_eq!(fs::read(&file).unwrap(), before, "{out:?}");

    // A receipt whose status is neither `committed` nor `recovered` plans without a block; the
    // apply writes no file and leaves its receipt alone.
    set_status(&f.home, &id, "failed");
    let before = fs::read(&file).unwrap();
    assert_eq!(host(&plan(&exe, &f))["preflight"], json!("ok"));
    apply_unchanged(&exe, &f, "unchanged");
    assert_eq!(fs::read(&file).unwrap(), before);

    // The version-only save is an error when the receipt folder cannot take a file.
    set_status(&f.home, &id, "committed");
    let before = fs::read(&file).unwrap();
    let dir = file.parent().unwrap().to_path_buf();
    let mode = fs::metadata(&dir).unwrap().permissions().mode();
    let _back = ModeBack(dir.clone(), mode);
    fs::set_permissions(&dir, fs::Permissions::from_mode(0o555)).unwrap();
    let probe = dir.join("probe");
    if fs::File::create(&probe).is_ok() {
        // A process that ignores folder modes (root) cannot make the save fail: not run.
        fs::remove_file(&probe).unwrap();
        eprintln!("not_run: the failing-save case, this process can write a read-only folder");
        return;
    }
    let out = apply_raw(&exe, &f);
    fs::set_permissions(&dir, fs::Permissions::from_mode(mode)).unwrap();
    assert!(!out.status.success(), "{out:?}");
    let v: Value = serde_json::from_slice(&out.stdout).expect("envelope JSON on stdout");
    assert_eq!(v["ok"], json!(false), "{v}");
    assert_eq!(fs::read(&file).unwrap(), before);
}
