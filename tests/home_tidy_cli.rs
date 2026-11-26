//! SPEC-20261124-home-tidy-slice2: `tidy-inventory` and `tidy-plan` read the home moai folder and
//! write nothing. Every test runs in a crafted temporary home; the real HOME is never touched.
//!
//! Stand-ins for "never opens a protected entry": a secret-named file and a secret-named folder
//! the test made unreadable (mode 000), a symlink, a dangling symlink and a named pipe. A tool
//! that opened the pipe would block, and the helper below kills a command that runs too long.
#![cfg(unix)]

mod common;
use common::CleanEnv;

use moai_cli::tidy::apply::{apply_with, Request, Step};
use serde_json::Value;
use std::{
    fs,
    io::Read,
    os::unix::fs::{symlink, PermissionsExt},
    path::{Path, PathBuf},
    process::{Command, Stdio},
    time::{Duration, Instant, SystemTime},
};
use tempfile::TempDir;

const BIN: &str = crate::common::MOAI_BIN;
/// The folder under HOME the operations tidy, named once through the product's constant.
const MOAI: &str = moai_cli::home_root::MOAI_DIR;
const DAY: u64 = 86_400;
const LIMIT: Duration = Duration::from_secs(60);

struct Out {
    ok: bool,
    stdout: String,
    stderr: String,
}

/// Runs `moai ARGS` with a clean environment and an optional PATH; kills it after `LIMIT`.
fn run_with(path: Option<&Path>, args: &[&str]) -> Out {
    // The operations take their home as an argument; HOME is a throwaway folder of its own, so
    // nothing the binary might resolve reaches the real one.
    let own_home = TempDir::new().unwrap();
    let mut cmd = Command::new(BIN);
    cmd.clean_env()
        .env("HOME", own_home.path())
        .args(args)
        .stdin(Stdio::null())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped());
    if let Some(p) = path {
        cmd.env("PATH", p);
    }
    let mut child = cmd.spawn().unwrap();
    let mut so = child.stdout.take().unwrap();
    let mut se = child.stderr.take().unwrap();
    let t_out = std::thread::spawn(move || {
        let mut s = String::new();
        let _ = so.read_to_string(&mut s);
        s
    });
    let t_err = std::thread::spawn(move || {
        let mut s = String::new();
        let _ = se.read_to_string(&mut s);
        s
    });
    let start = Instant::now();
    let status = loop {
        if let Some(st) = child.try_wait().unwrap() {
            break st;
        }
        if start.elapsed() > LIMIT {
            let _ = child.kill();
            let _ = child.wait();
            panic!("moai {args:?} ran longer than {LIMIT:?}: it opened something that blocks");
        }
        std::thread::sleep(Duration::from_millis(20));
    };
    Out {
        ok: status.success(),
        stdout: t_out.join().unwrap(),
        stderr: t_err.join().unwrap(),
    }
}

fn run(args: &[&str]) -> Out {
    run_with(None, args)
}

/// The `result` of a success envelope.
fn result(out: &Out) -> Value {
    assert!(out.ok, "failed: {} {}", out.stdout, out.stderr);
    let v: Value = serde_json::from_str(&out.stdout).unwrap();
    assert_eq!(v["ok"], Value::Bool(true), "{v}");
    v["result"].clone()
}

fn set_age(p: &Path, days: u64) {
    let t = SystemTime::now() - Duration::from_secs(days * DAY);
    fs::File::open(p).unwrap().set_modified(t).unwrap();
}

/// Gives `p` and everything under it the age `days` (children first, so a folder keeps it).
fn age_tree(p: &Path, days: u64) {
    let md = fs::symlink_metadata(p).unwrap();
    if md.is_dir() {
        for e in fs::read_dir(p).unwrap() {
            age_tree(&e.unwrap().path(), days);
        }
    }
    if !md.file_type().is_symlink() {
        set_age(p, days);
    }
}

fn put(p: &Path, bytes: usize) {
    fs::create_dir_all(p.parent().unwrap()).unwrap();
    fs::write(p, vec![b'x'; bytes]).unwrap();
}

struct Fx {
    _tmp: TempDir,
    home: PathBuf,
    moai: PathBuf,
}

impl Drop for Fx {
    fn drop(&mut self) {
        // Let the temporary folder be removed: the test made these unreadable.
        let _ = fs::set_permissions(
            self.moai.join("locked-dir"),
            fs::Permissions::from_mode(0o755),
        );
    }
}

/// A crafted home: one entry of each class and the folder kinds that grow.
fn fixture() -> Fx {
    let tmp = TempDir::new().unwrap();
    let home = fs::canonicalize(tmp.path()).unwrap().join("home");
    let moai = home.join(MOAI);
    fs::create_dir_all(&moai).unwrap();
    // owned, live: a report changed 5 days ago next to two old ones.
    put(&moai.join("reports/projA/new.md"), 1024);
    put(&moai.join("reports/projA/old1.md"), 2048);
    put(&moai.join("reports/projA/old2.md"), 1024);
    age_tree(&moai.join("reports"), 5);
    // The older file sorts last by name, so only an age sort puts it first.
    set_age(&moai.join("reports/projA/old1.md"), 200);
    set_age(&moai.join("reports/projA/old2.md"), 300);
    // owned, stale: nothing changed for a long time.
    put(&moai.join("handoffs/h1.md"), 512);
    age_tree(&moai.join("handoffs"), 100);
    // kinds whose safety needs git proof.
    put(&moai.join("evidence/ev1/e.txt"), 700);
    age_tree(&moai.join("evidence"), 200);
    put(&moai.join("build/b1/x.o"), 900);
    age_tree(&moai.join("build"), 100);
    put(&moai.join("worktrees/projA/w1/f.rs"), 300);
    age_tree(&moai.join("worktrees"), 100);
    put(&moai.join("receipts/r1/r.json"), 300);
    age_tree(&moai.join("receipts"), 100);
    // retired tool: a folder with files, one with files (run), one empty.
    put(&moai.join("releases/v1/bin"), 800);
    age_tree(&moai.join("releases"), 400);
    put(&moai.join("run/pid"), 10);
    age_tree(&moai.join("run"), 400);
    fs::create_dir_all(moai.join("db")).unwrap();
    set_age(&moai.join("db"), 400);
    fs::create_dir_all(moai.join("rescue")).unwrap();
    set_age(&moai.join("rescue"), 400);
    // unknown writer (parent D-10) and a name in no table.
    put(&moai.join("logs/hook.log"), 5 * 1024);
    age_tree(&moai.join("logs"), 10);
    put(&moai.join("mystery/x"), 40);
    age_tree(&moai.join("mystery"), 50);
    // another product: listed, never changed; its file is unreadable.
    put(&moai.join("claude-profiles/p1/auth.json"), 64);
    age_tree(&moai.join("claude-profiles"), 3);
    fs::set_permissions(
        moai.join("claude-profiles/p1/auth.json"),
        fs::Permissions::from_mode(0o000),
    )
    .unwrap();
    // protected: secret-shaped names (unreadable), a link, a dangling link, a named pipe.
    for name in [".env.glm", "credentials.yaml"] {
        put(&moai.join(name), 32);
        fs::set_permissions(moai.join(name), fs::Permissions::from_mode(0o000)).unwrap();
    }
    fs::create_dir_all(moai.join("secrets")).unwrap();
    fs::set_permissions(moai.join("secrets"), fs::Permissions::from_mode(0o000)).unwrap();
    symlink(moai.join("reports"), moai.join("link-to-reports")).unwrap();
    symlink(moai.join("nowhere"), moai.join("dangling-link")).unwrap();
    let mk = Command::new("mkfifo")
        .arg(moai.join("pipe"))
        .status()
        .unwrap();
    assert!(mk.success());
    // unreadable and not secret-shaped: listed as unknown with the reason.
    put(&moai.join("locked-dir/f"), 5);
    fs::set_permissions(moai.join("locked-dir"), fs::Permissions::from_mode(0o000)).unwrap();
    Fx {
        _tmp: tmp,
        home,
        moai,
    }
}

fn inventory(fx: &Fx, extra: &[&str]) -> Value {
    let home = fx.home.to_str().unwrap();
    let mut args = vec!["tidy-inventory", home];
    args.extend_from_slice(extra);
    result(&run(&args))
}

fn plan(fx: &Fx, extra: &[&str]) -> Value {
    let home = fx.home.to_str().unwrap();
    let mut args = vec!["tidy-plan", home];
    args.extend_from_slice(extra);
    result(&run(&args))
}

fn entry<'a>(v: &'a Value, name: &str) -> &'a Value {
    v["entries"]
        .as_array()
        .unwrap()
        .iter()
        .find(|e| e["name"] == name)
        .unwrap_or_else(|| panic!("no entry {name}: {v}"))
}

fn item<'a>(v: &'a Value, path: &str) -> &'a Value {
    v["items"]
        .as_array()
        .unwrap()
        .iter()
        .find(|e| e["path"] == path)
        .unwrap_or_else(|| panic!("no item {path}: {v}"))
}

fn kind<'a>(v: &'a Value, name: &str) -> Option<&'a Value> {
    v["kinds"]
        .as_array()
        .unwrap()
        .iter()
        .find(|k| k["kind"] == name)
}

fn text(v: &Value) -> &str {
    v.as_str().unwrap_or("")
}

// ---------------------------------------------------------------- inventory_

/// Runs `program ARGS` in a macOS sandbox that kills the process the moment it opens one of
/// `denied` (a folder with everything below it) to read its data. Listing a name and reading
/// its own directory data stay allowed, so the signal tells an open call from a listing.
#[cfg(target_os = "macos")]
fn killed_on_open(denied: &[PathBuf], program: &str, args: &[&str]) -> std::process::Output {
    let mut profile = String::from("(version 1)(allow default)");
    for p in denied {
        profile.push_str(&format!(
            "(deny file-read-data (subpath \"{}\") (with send-signal SIGKILL))",
            p.display()
        ));
    }
    let own_home = TempDir::new().unwrap();
    Command::new("/usr/bin/sandbox-exec")
        .clean_env()
        .env("HOME", own_home.path())
        .arg("-p")
        .arg(profile)
        .arg(program)
        .args(args)
        .stdin(Stdio::null())
        .output()
        .unwrap()
}

/// The open calls of both operations, observed: no protected entry, no key of another product
/// and nothing a link points at is ever opened. The stand-ins are readable here, so an open
/// would reach the sandbox rule and not stop at the file mode.
#[cfg(target_os = "macos")]
#[test]
fn inventory_and_plan_open_no_protected_entry() {
    use std::os::unix::process::ExitStatusExt;
    let fx = fixture();
    let outside = fx.home.join("outside");
    put(&outside.join("key.txt"), 16);
    put(&outside.join("dir/key.txt"), 16);
    symlink(outside.join("key.txt"), fx.moai.join("link-outside")).unwrap();
    symlink(outside.join("dir"), fx.moai.join("reports/projA/link-dir")).unwrap();
    for (name, mode) in [
        (".env.glm", 0o644),
        ("credentials.yaml", 0o644),
        ("secrets", 0o755),
        ("claude-profiles/p1/auth.json", 0o644),
    ] {
        fs::set_permissions(fx.moai.join(name), fs::Permissions::from_mode(mode)).unwrap();
    }
    put(&fx.moai.join("secrets/token"), 8);
    put(&fx.moai.join("releases/v1/.env"), 8);
    let denied: Vec<PathBuf> = [
        ".env.glm",
        "credentials.yaml",
        "secrets",
        "claude-profiles/p1/auth.json",
        "pipe",
        "releases/v1/.env",
    ]
    .iter()
    .map(|n| fx.moai.join(n))
    .chain([outside.clone()])
    .collect();
    // The probe itself: each denied path kills a program that does open it.
    for p in &denied {
        let program = if p.is_dir() { "/bin/ls" } else { "/bin/cat" };
        let out = killed_on_open(&denied, program, &[p.to_str().unwrap()]);
        assert_eq!(out.status.signal(), Some(9), "{}: {out:?}", p.display());
    }
    for op in ["tidy-inventory", "tidy-plan"] {
        let out = killed_on_open(&denied, BIN, &[op, fx.home.to_str().unwrap()]);
        assert!(
            out.status.success(),
            "{op} was stopped (signal {:?}): {}",
            out.status.signal(),
            String::from_utf8_lossy(&out.stderr)
        );
        let v: Value = serde_json::from_slice(&out.stdout).unwrap();
        assert_eq!(v["ok"], Value::Bool(true), "{op}: {v}");
    }
}

#[test]
fn inventory_lists_every_entry_with_class_size_newest_and_reason() {
    let fx = fixture();
    let inv = inventory(&fx, &[]);
    let mut want: Vec<String> = fs::read_dir(&fx.moai)
        .unwrap()
        .map(|e| e.unwrap().file_name().to_string_lossy().into_owned())
        .collect();
    want.sort();
    let mut got: Vec<String> = inv["entries"]
        .as_array()
        .unwrap()
        .iter()
        .map(|e| text(&e["name"]).to_string())
        .collect();
    got.sort();
    assert_eq!(got, want, "every top-level entry, no more, no fewer");
    for e in inv["entries"].as_array().unwrap() {
        assert!(e["class"].is_string(), "{e}");
        assert!(e["bytes"].is_u64(), "{e}");
        assert!(e["newest"].is_u64(), "{e}");
        assert!(!text(&e["reason"]).is_empty(), "{e}");
    }
    for (name, class) in [
        ("reports", "owned_live"),
        ("handoffs", "owned_stale"),
        ("evidence", "owned_stale"),
        ("releases", "retired_tool"),
        ("run", "retired_tool"),
        ("db", "retired_tool"),
        ("claude-profiles", "other_product"),
        ("logs", "unknown"),
        ("mystery", "unknown"),
    ] {
        assert_eq!(entry(&inv, name)["class"], class, "{name}");
    }
    // The report folder holds 1024 + 2048 + 1024 file bytes.
    assert_eq!(entry(&inv, "reports")["bytes"], 4096);
}

#[test]
fn inventory_protected_entries_are_listed_and_never_opened() {
    let fx = fixture();
    // The pipe blocks a tool that opens it; the unreadable files and the folder fail it.
    let inv = inventory(&fx, &[]);
    for (name, ty) in [
        (".env.glm", "file"),
        ("credentials.yaml", "file"),
        ("secrets", "dir"),
        ("link-to-reports", "symlink"),
        ("dangling-link", "symlink"),
        ("pipe", "fifo"),
    ] {
        let e = entry(&inv, name);
        assert_eq!(e["class"], "protected", "{name}: {e}");
        assert_eq!(e["type"], ty, "{name}: {e}");
        assert!(e["newest"].is_u64(), "{name}: {e}");
        assert!(e.get("sha256").is_none() && e.get("files").is_none(), "{e}");
    }
    // A link is not followed: its size is the link's own, not the report folder's 4096 bytes.
    assert_ne!(entry(&inv, "link-to-reports")["bytes"], 4096);
}

#[test]
fn inventory_unreadable_entry_is_unknown_with_the_reason_and_the_run_goes_on() {
    let fx = fixture();
    let inv = inventory(&fx, &[]);
    let e = entry(&inv, "locked-dir");
    assert_eq!(e["class"], "unknown", "{e}");
    assert!(text(&e["reason"]).contains("unreadable"), "{e}");
    // Entries after it are still listed.
    assert_eq!(entry(&inv, "run")["class"], "retired_tool");
}

#[test]
fn inventory_walk_over_the_entry_limit_keeps_its_group_with_over_limit() {
    let tmp = TempDir::new().unwrap();
    let home = fs::canonicalize(tmp.path()).unwrap().join("home");
    for i in 0..12 {
        put(&home.join(format!("{MOAI}/reports/p/f{i}.md")), 100);
    }
    let fx = Fx {
        _tmp: tmp,
        moai: home.join(MOAI),
        home,
    };
    let inv = inventory(&fx, &["--entry-limit", "5"]);
    let e = entry(&inv, "reports");
    assert_eq!(e["over_limit"], true, "{e}");
    assert!(text(&e["reason"]).contains("over_limit"), "{e}");
    assert!(e["bytes"].as_u64().unwrap() < 1200, "size seen so far: {e}");
    let full = inventory(&fx, &[]);
    assert_eq!(entry(&full, "reports")["bytes"], 1200);
    assert_ne!(entry(&full, "reports")["over_limit"], true);
}

#[test]
fn inventory_a_link_below_a_top_level_folder_is_never_followed() {
    // Links below a folder point outside the home: to a big file and to a folder with a file.
    // A walk that followed them would add their 80000 bytes to the folder it found them in.
    let tmp = TempDir::new().unwrap();
    let base = fs::canonicalize(tmp.path()).unwrap();
    let home = base.join("home");
    let moai = home.join(MOAI);
    let outside = base.join("outside");
    put(&outside.join("big.bin"), 50_000);
    put(&outside.join("dir/inner.bin"), 30_000);
    // `reports` is a folder kind with items; `releases` is a folder of a retired tool.
    put(&moai.join("reports/p/a.md"), 100);
    symlink(outside.join("big.bin"), moai.join("reports/p/link-file")).unwrap();
    symlink(outside.join("dir"), moai.join("reports/p/link-dir")).unwrap();
    put(&moai.join("releases/v1/bin"), 200);
    symlink(outside.join("dir"), moai.join("releases/v1/link-dir")).unwrap();
    symlink(outside.join("big.bin"), moai.join("releases/link-file")).unwrap();
    let fx = Fx {
        _tmp: tmp,
        home,
        moai,
    };
    let inv = inventory(&fx, &[]);
    for (name, bytes) in [("reports", 100), ("releases", 200)] {
        let e = entry(&inv, name);
        assert_eq!(
            e["bytes"], bytes,
            "{name} counted what a link points to: {e}"
        );
        assert_eq!(e["files"], 1, "{name} counted a file behind a link: {e}");
    }
    // The kind row counts the one report that is a file, not the links beside it.
    assert_eq!(kind(&inv, "reports").unwrap()["count"], 1, "{inv}");
    assert_eq!(kind(&inv, "reports").unwrap()["bytes"], 100, "{inv}");
}

#[test]
fn inventory_a_moai_folder_that_is_a_link_is_refused_by_both_operations() {
    // The state folder under the home is itself a link to a real folder: neither operation
    // follows it.
    let tmp = TempDir::new().unwrap();
    let base = fs::canonicalize(tmp.path()).unwrap();
    let home = base.join("home");
    let real = base.join("elsewhere");
    put(&real.join("reports/p/a.md"), 100);
    fs::create_dir_all(&home).unwrap();
    symlink(&real, home.join(MOAI)).unwrap();
    let home_arg = home.to_str().unwrap();
    for op in ["tidy-inventory", "tidy-plan"] {
        let out = run(&[op, home_arg]);
        assert!(
            !out.ok,
            "{op} followed the link: {} {}",
            out.stdout, out.stderr
        );
        let all = format!("{}{}", out.stdout, out.stderr);
        assert!(all.contains("link"), "{op} did not say it is a link: {all}");
        assert!(
            !all.contains("a.md") && !all.contains("\"reports\""),
            "{op} listed what is behind the link: {all}"
        );
    }
}

// ------------------------------------------------------------------- kinds_

#[test]
fn kinds_rows_carry_count_size_oldest_newest_and_the_passed_thresholds() {
    let fx = fixture();
    let inv = inventory(&fx, &["--threshold", "reports.max_files=2"]);
    let r = kind(&inv, "reports").expect("reports row");
    assert_eq!(r["count"], 3);
    assert_eq!(r["bytes"], 4096);
    assert!(
        r["oldest"].as_u64().unwrap() < r["newest"].as_u64().unwrap(),
        "{r}"
    );
    assert_eq!(r["passed"], serde_json::json!(["max_files"]), "{r}");
    assert_eq!(r["thresholds"]["max_files"], 2);
    // The other defaults of the parent D-5 table are shown unchanged.
    assert_eq!(r["thresholds"]["max_mib"], 500);
    assert_eq!(r["thresholds"]["archive_after_days"], 90);
    // Defaults pass nothing on this small home.
    let plain = inventory(&fx, &[]);
    for k in [
        "reports",
        "logs",
        "evidence",
        "worktrees",
        "build",
        "receipts",
    ] {
        let row = kind(&plain, k).unwrap_or_else(|| panic!("no row {k}"));
        assert_eq!(row["passed"], serde_json::json!([]), "{k}: {row}");
    }
}

#[test]
fn kinds_a_kind_whose_folder_is_absent_has_no_row() {
    let tmp = TempDir::new().unwrap();
    let home = fs::canonicalize(tmp.path()).unwrap().join("home");
    put(&home.join(MOAI).join("reports/p/a.md"), 10);
    let fx = Fx {
        _tmp: tmp,
        moai: home.join(MOAI),
        home,
    };
    let inv = inventory(&fx, &[]);
    assert!(kind(&inv, "reports").is_some());
    for k in ["logs", "evidence", "worktrees", "build", "receipts"] {
        assert!(kind(&inv, k).is_none(), "{k}");
    }
}

// --------------------------------------------------------------- threshold_

#[test]
fn threshold_values_override_the_defaults_for_each_kind_and_may_be_fractions() {
    let fx = fixture();
    let inv = inventory(
        &fx,
        &[
            "--threshold",
            "reports.max_mib=0.001",
            "--threshold",
            "worktrees.max_count=0",
            "--threshold",
            "build.max_gib=0",
            "--threshold",
            "evidence.max_mib=0",
            "--threshold",
            "receipts.max_mib=0",
            "--threshold",
            "logs.file_mib=0.001",
        ],
    );
    // 4096 bytes is over 0.001 MiB (1048.576 bytes).
    assert_eq!(
        kind(&inv, "reports").unwrap()["passed"],
        serde_json::json!(["max_mib"])
    );
    for (k, f) in [
        ("worktrees", "max_count"),
        ("build", "max_gib"),
        ("evidence", "max_mib"),
        ("receipts", "max_mib"),
        ("logs", "file_mib"),
    ] {
        let passed = kind(&inv, k).unwrap()["passed"].clone();
        assert!(
            passed.as_array().unwrap().iter().any(|p| p == f),
            "{k}: {passed}"
        );
    }
}

#[test]
fn threshold_a_bad_argument_refuses_before_any_read_and_names_the_argument() {
    // The home does not exist: an error about the argument shows the arguments are read first.
    for bad in [
        "bogus.max_files=1",
        "reports.bogus=1",
        "reports.max_files=abc",
        "reports.max_files",
        "reports.max_files=-1",
    ] {
        for op in ["tidy-inventory", "tidy-plan"] {
            let out = run(&[op, "/nonexistent-home-for-tidy-test", "--threshold", bad]);
            assert!(!out.ok, "{op} {bad}: {} {}", out.stdout, out.stderr);
            let all = format!("{}{}", out.stdout, out.stderr);
            assert!(all.contains(bad), "{op} {bad} not named: {all}");
            assert!(
                !all.contains("nonexistent-home-for-tidy-test"),
                "{op} {bad} read the home first: {all}"
            );
        }
    }
}

// -------------------------------------------------------------------- plan_

#[test]
fn plan_names_action_rule_and_reason_for_each_candidate() {
    let fx = fixture();
    let p = plan(&fx, &[]);
    let items = p["items"].as_array().unwrap();
    assert!(!items.is_empty());
    for i in items {
        let action = text(&i["action"]);
        assert!(["archive", "delete", "keep"].contains(&action), "{i}");
        assert!(!text(&i["rule"]).is_empty(), "{i}");
        assert!(!text(&i["reason"]).is_empty(), "{i}");
        assert!(i["kind"].is_string(), "{i}");
    }
    let kept = item(&p, "releases");
    assert_eq!(kept["action"], "keep", "{kept}");
    assert!(text(&kept["reason"]).contains("go_run_data_kept"), "{kept}");
    assert_eq!(item(&p, "run")["action"], "archive");
    // Never offered: another product, an unknown entry, an unknown writer, a protected entry.
    for (path, reason) in [
        ("claude-profiles", "other_product"),
        ("mystery", "unknown"),
        ("logs", "unknown"),
    ] {
        let i = item(&p, path);
        assert_eq!(i["action"], "keep", "{path}: {i}");
        assert!(text(&i["reason"]).contains(reason), "{path}: {i}");
    }
    for path in [".env.glm", "secrets", "pipe", "link-to-reports"] {
        assert_eq!(item(&p, path)["action"], "keep", "{path}");
    }
}

#[test]
fn plan_the_only_delete_is_an_empty_folder_of_the_retired_tool_class() {
    let fx = fixture();
    for extra in [
        vec![],
        vec![
            "--threshold",
            "reports.max_files=0",
            "--threshold",
            "build.max_gib=0",
            "--threshold",
            "worktrees.max_count=0",
        ],
    ] {
        let p = plan(&fx, &extra);
        let deletes: Vec<&str> = p["items"]
            .as_array()
            .unwrap()
            .iter()
            .filter(|i| i["action"] == "delete")
            .map(|i| text(&i["path"]))
            .collect();
        assert_eq!(deletes, vec!["rescue"], "{extra:?}");
    }
    // The same folder with a file in it is archived, not deleted.
    put(&fx.moai.join("rescue/leftover"), 3);
    let p = plan(&fx, &[]);
    assert_eq!(item(&p, "rescue")["action"], "archive");
}

#[test]
fn plan_over_threshold_offers_the_old_reports_oldest_first() {
    let fx = fixture();
    let p = plan(&fx, &["--threshold", "reports.max_files=2"]);
    let archives: Vec<&str> = p["items"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|i| i["kind"] == "reports" && i["action"] == "archive")
        .map(|i| text(&i["path"]))
        .collect();
    assert_eq!(
        archives,
        vec!["reports/projA/old2.md", "reports/projA/old1.md"],
        "300 days before 200 days"
    );
    let fresh = item(&p, "reports/projA/new.md");
    assert_eq!(fresh["action"], "keep");
    assert!(text(&fresh["reason"]).contains("not_old_enough"), "{fresh}");
    assert_eq!(
        kind(&p, "reports").unwrap()["passed"],
        serde_json::json!(["max_files"])
    );
}

#[test]
fn plan_under_every_threshold_offers_nothing_from_that_kind() {
    let fx = fixture();
    let p = plan(&fx, &[]);
    let from_reports: Vec<&Value> = p["items"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|i| i["kind"] == "reports")
        .collect();
    assert_eq!(from_reports.len(), 3, "every report is listed");
    for i in from_reports {
        assert_eq!(i["action"], "keep", "{i}");
        assert!(text(&i["reason"]).contains("under_threshold"), "{i}");
    }
}

#[test]
fn plan_worktrees_build_evidence_and_receipts_wait_for_proof() {
    let fx = fixture();
    let p = plan(
        &fx,
        &[
            "--threshold",
            "worktrees.max_count=0",
            "--threshold",
            "build.max_gib=0",
            "--threshold",
            "evidence.max_mib=0",
            "--threshold",
            "receipts.max_mib=0",
        ],
    );
    for path in [
        "worktrees/projA/w1",
        "build/b1",
        "evidence/ev1",
        "receipts/r1",
    ] {
        let i = item(&p, path);
        assert_eq!(i["action"], "keep", "{path}: {i}");
        assert!(text(&i["reason"]).contains("proof_pending"), "{path}: {i}");
    }
    for k in ["worktrees", "build", "evidence", "receipts"] {
        assert!(
            !kind(&p, k).unwrap()["passed"]
                .as_array()
                .unwrap()
                .is_empty(),
            "{k}"
        );
    }
}

#[test]
fn plan_digest_is_stable_for_one_folder_and_changes_with_the_candidates() {
    let fx = fixture();
    let a = plan(&fx, &[]);
    let b = plan(&fx, &[]);
    assert!(text(&a["digest"]).starts_with("sha256:"), "{}", a["digest"]);
    assert_eq!(a["digest"], b["digest"]);
    put(&fx.moai.join("run/another"), 77);
    age_tree(&fx.moai.join("run"), 400);
    let c = plan(&fx, &[]);
    assert_ne!(
        a["digest"], c["digest"],
        "a changed candidate changes the digest"
    );
    let d = plan(&fx, &["--threshold", "reports.max_files=2"]);
    assert_ne!(
        c["digest"], d["digest"],
        "a different offer changes the digest"
    );
}

/// Makes a folder readable again when dropped, so the temporary home can be removed.
struct Unlock(PathBuf);

impl Drop for Unlock {
    fn drop(&mut self) {
        let _ = fs::set_permissions(&self.0, fs::Permissions::from_mode(0o755));
    }
}

/// A home whose only entry is the retired-tool folder `cache` with a readable file.
fn retired_tool_home() -> Fx {
    let tmp = TempDir::new().unwrap();
    let home = fs::canonicalize(tmp.path()).unwrap().join("home");
    let moai = home.join(MOAI);
    put(&moai.join("cache/v1/bin"), 800);
    age_tree(&moai.join("cache"), 400);
    Fx {
        _tmp: tmp,
        home,
        moai,
    }
}

#[test]
fn plan_a_retired_tool_folder_with_an_unreadable_part_is_kept_not_archived() {
    let fx = retired_tool_home();
    // The whole folder looks old and readable until the walk meets the one part it cannot list.
    let whole = plan(&fx, &[]);
    assert_eq!(
        item(&whole, "cache")["action"],
        "archive",
        "control: {whole}"
    );
    put(&fx.moai.join("cache/v1/locked/f"), 5);
    let locked = fx.moai.join("cache/v1/locked");
    fs::set_permissions(&locked, fs::Permissions::from_mode(0o000)).unwrap();
    let _unlock = Unlock(locked);
    let p = plan(&fx, &[]);
    let i = item(&p, "cache");
    assert_eq!(
        i["action"], "keep",
        "archived what it could not read in full: {i}"
    );
    assert_eq!(i["rule"], "keep_unknown", "{i}");
    assert!(text(&i["reason"]).contains("unknown"), "{i}");
    assert!(
        p["items"]
            .as_array()
            .unwrap()
            .iter()
            .all(|i| i["action"] == "keep"),
        "nothing is offered from a folder read in part: {p}"
    );
}

#[test]
fn plan_a_retired_tool_folder_over_the_entry_limit_is_kept_not_archived() {
    let fx = retired_tool_home();
    for n in 0..12 {
        put(&fx.moai.join(format!("cache/v1/f{n}")), 10);
    }
    age_tree(&fx.moai.join("cache"), 400);
    let whole = plan(&fx, &[]);
    assert_eq!(
        item(&whole, "cache")["action"],
        "archive",
        "control: {whole}"
    );
    // A lowered limit stops the walk partway: the size is only what was seen.
    let p = plan(&fx, &["--entry-limit", "5"]);
    let i = item(&p, "cache");
    assert_eq!(
        i["action"], "keep",
        "archived a folder it only partly counted: {i}"
    );
    assert_eq!(i["rule"], "keep_unknown", "{i}");
    assert!(text(&i["reason"]).contains("unknown"), "{i}");
}

// one memory index of `lines` lines of `width` bytes each (the newline included).
fn memory_file(dir: &Path, name: &str, lines: usize, width: usize) -> PathBuf {
    let p = dir.join(name);
    let line = format!("{}\n", "m".repeat(width - 1));
    fs::write(&p, line.repeat(lines)).unwrap();
    p
}

// ------------------------------------------------------------------ memory_

#[test]
fn memory_index_size_line_reports_lines_bytes_percent_and_the_over_80_flag() {
    let fx = fixture();
    let dir = fx.home.join("mem");
    fs::create_dir_all(&dir).unwrap();
    let m150 = memory_file(&dir, "m150.md", 150, 10);
    let m170 = memory_file(&dir, "m170.md", 170, 10);
    let m30k = memory_file(&dir, "m30k.md", 30, 1024);
    let inv = inventory(
        &fx,
        &[
            "--memory-index",
            m150.to_str().unwrap(),
            "--memory-index",
            m170.to_str().unwrap(),
            "--memory-index",
            m30k.to_str().unwrap(),
        ],
    );
    let rows = inv["memory_index"].as_array().unwrap();
    assert_eq!(rows.len(), 3);
    let by = |n: &str| rows.iter().find(|r| text(&r["path"]).ends_with(n)).unwrap();
    let r = by("m150.md");
    assert_eq!(
        (r["lines"].as_u64(), r["bytes"].as_u64()),
        (Some(150), Some(1500))
    );
    assert_eq!(r["percent"].as_f64(), Some(75.0));
    assert_eq!(r["over_80"], false);
    let r = by("m170.md");
    assert_eq!(r["percent"].as_f64(), Some(85.0));
    assert_eq!(r["over_80"], true);
    let r = by("m30k.md");
    assert_eq!(r["bytes"], 30 * 1024);
    assert_eq!(
        r["percent"].as_f64(),
        Some(120.0),
        "30 KiB of the 25 KiB cap"
    );
    assert_eq!(r["over_80"], true);
    // The targets of the leader's note: under 100 lines, 12 KB, 120 bytes a line.
    assert_eq!(r["targets"]["lines"], 100);
    assert_eq!(r["targets"]["bytes"], 12 * 1024);
    assert_eq!(r["targets"]["line_bytes"], 120);
    // The plan carries the same rows.
    let p = plan(&fx, &["--memory-index", m170.to_str().unwrap()]);
    assert_eq!(p["memory_index"][0]["percent"].as_f64(), Some(85.0));
}

#[test]
fn memory_plan_text_for_an_index_over_80_percent_and_none_under_it() {
    let fx = fixture();
    let dir = fx.home.join("mem");
    fs::create_dir_all(&dir).unwrap();
    let under = memory_file(&dir, "under.md", 150, 10);
    let over = memory_file(&dir, "over.md", 170, 10);
    let p = plan(&fx, &["--memory-index", under.to_str().unwrap()]);
    assert!(
        p.get("memory_tidy").is_none() || p["memory_tidy"].is_null(),
        "{p}"
    );
    let p = plan(&fx, &["--memory-index", over.to_str().unwrap()]);
    let t = &p["memory_tidy"];
    assert_eq!(
        t["groups"],
        serde_json::json!(["rules_to_instructions", "retired_items", "merges"])
    );
    assert!(!text(&t["run_in_claude_session"]).is_empty(), "{t}");
    assert_eq!(t["purge_reference"]["command"], "claude purge --all");
    assert_eq!(t["purge_reference"]["never_run"], true);
}

#[test]
fn memory_a_missing_file_is_reported_with_a_reason_and_the_run_goes_on() {
    let fx = fixture();
    let missing = fx.home.join("mem/none.md");
    let inv = inventory(&fx, &["--memory-index", missing.to_str().unwrap()]);
    let r = &inv["memory_index"][0];
    assert_eq!(r["found"], false, "{r}");
    assert!(!text(&r["reason"]).is_empty(), "{r}");
    assert!(inv["entries"].as_array().unwrap().len() > 5);
}

#[test]
fn memory_an_index_over_the_read_limit_is_reported_with_a_reason() {
    let fx = fixture();
    let dir = fx.home.join("mem");
    fs::create_dir_all(&dir).unwrap();
    let big = dir.join("big.md");
    fs::write(&big, vec![b'a'; 3 * 1024 * 1024]).unwrap();
    let inv = inventory(&fx, &["--memory-index", big.to_str().unwrap()]);
    let r = &inv["memory_index"][0];
    assert_eq!(r["found"], true, "{r}");
    assert!(text(&r["reason"]).contains("over_limit"), "{r}");
}

// ----------------------------------------------------------------- readonly_

/// Path, kind, length, mode, mtime and link target of every entry under `p`, never opened.
fn snapshot(p: &Path, out: &mut Vec<String>) {
    let md = fs::symlink_metadata(p).unwrap();
    let mtime = md
        .modified()
        .unwrap()
        .duration_since(SystemTime::UNIX_EPOCH)
        .unwrap();
    let link = if md.file_type().is_symlink() {
        fs::read_link(p).unwrap().display().to_string()
    } else {
        String::new()
    };
    out.push(format!(
        "{} {:?} {} {:o} {} {}",
        p.display(),
        md.file_type(),
        md.len(),
        md.permissions().mode(),
        mtime.as_nanos(),
        link
    ));
    if md.is_dir() && md.permissions().mode() & 0o500 == 0o500 {
        let mut kids: Vec<_> = fs::read_dir(p)
            .unwrap()
            .map(|e| e.unwrap().path())
            .collect();
        kids.sort();
        for k in kids {
            snapshot(&k, out);
        }
    }
}

fn snap(dir: &Path) -> Vec<String> {
    let mut v = Vec::new();
    snapshot(dir, &mut v);
    v
}

#[test]
fn readonly_both_operations_leave_the_home_the_memory_index_and_the_path_commands_alone() {
    let fx = fixture();
    let mem = fx.home.join("mem");
    fs::create_dir_all(&mem).unwrap();
    let idx = memory_file(&mem, "MEMORY.md", 170, 10);
    // `claude` and `git` stand-ins that record every call.
    let tools = TempDir::new().unwrap();
    let bin = fs::canonicalize(tools.path()).unwrap().join("bin");
    fs::create_dir_all(&bin).unwrap();
    let log = bin.join("calls.log");
    for name in ["claude", "git"] {
        let script = bin.join(name);
        fs::write(
            &script,
            format!(
                "#!/bin/sh\necho \"{name} $*\" >> \"{}\"\nexit 0\n",
                log.display()
            ),
        )
        .unwrap();
        fs::set_permissions(&script, fs::Permissions::from_mode(0o755)).unwrap();
    }
    let path = format!("{}:/usr/bin:/bin", bin.display());
    let before = snap(&fx.home);
    for op in ["tidy-inventory", "tidy-plan"] {
        let out = run_with(
            Some(Path::new(&path)),
            &[
                op,
                fx.home.to_str().unwrap(),
                "--memory-index",
                idx.to_str().unwrap(),
            ],
        );
        assert!(out.ok, "{op}: {} {}", out.stdout, out.stderr);
    }
    assert_eq!(snap(&fx.home), before, "nothing in the home changed");
    let calls = fs::read_to_string(&log).unwrap_or_default();
    assert!(!calls.contains("claude"), "claude must not run: {calls}");
    for l in calls.lines() {
        let sub = l.split_whitespace().nth(1).unwrap_or("");
        assert!(
            ["status", "rev-parse", "log", "diff", "show", "branch"].contains(&sub)
                || l.starts_with("git worktree list"),
            "a git call that may write: {l}"
        );
    }
}

// -------------------------------------------------------------------- help_

#[test]
fn help_the_operations_say_they_write_nothing() {
    // The operations document fragments are checked by `operations_doc` in the binary's tests.
    for op in ["tidy-inventory", "tidy-plan"] {
        let out = run(&[op, "--help"]);
        assert!(out.ok, "{op}: {}", out.stderr);
        assert!(
            out.stdout.contains("Writes nothing"),
            "{op}: {}",
            out.stdout
        );
    }
}

#[test]
fn help_apply_and_restore_say_what_they_change() {
    for (op, words) in [
        (
            "tidy-apply",
            ["--confirm", "--permission-reference", "--pick", "Removes"],
        ),
        (
            "tidy-restore",
            [
                "--archive",
                "--permission-reference",
                "--pick",
                "overwrites",
            ],
        ),
    ] {
        let out = run(&[op, "--help"]);
        assert!(out.ok, "{op}: {}", out.stderr);
        for w in words {
            assert!(out.stdout.contains(w), "{op} lacks {w}: {}", out.stdout);
        }
    }
}

// ---------------------------------------------- SPEC-20261126-home-tidy-apply

/// The permission reference the tests hand to the operations that write.
const PERMIT: &str = "test:picked";
/// Lowers the report file count so the old reports of the fixture are offered.
const OVER: [&str; 2] = ["--threshold", "reports.max_files=0"];

fn all(out: &Out) -> String {
    format!("{}{}", out.stdout, out.stderr)
}

fn token(fx: &Fx) -> String {
    text(&plan(fx, &OVER)["digest"]).to_string()
}

/// `tidy-apply` over the fixture with the given picks, confirm token and permission reference.
fn apply_args(fx: &Fx, picks: &[&str], confirm: &str, permit: &str) -> Vec<String> {
    let mut args = vec![
        "tidy-apply".to_string(),
        fx.home.display().to_string(),
        "--confirm".to_string(),
        confirm.to_string(),
        "--permission-reference".to_string(),
        permit.to_string(),
    ];
    for p in picks {
        args.push("--pick".to_string());
        args.push((*p).to_string());
    }
    args.extend(OVER.iter().map(|s| s.to_string()));
    args
}

fn restore_args(fx: &Fx, stamp: &str, picks: &[&str], permit: &str) -> Vec<String> {
    let mut args = vec![
        "tidy-restore".to_string(),
        fx.home.display().to_string(),
        "--archive".to_string(),
        stamp.to_string(),
        "--permission-reference".to_string(),
        permit.to_string(),
    ];
    for p in picks {
        args.push("--pick".to_string());
        args.push((*p).to_string());
    }
    args
}

fn call(args: &[String]) -> Out {
    run(&args.iter().map(String::as_str).collect::<Vec<_>>())
}

/// Applies the picks with the token of a fresh plan and returns the result.
fn applied(fx: &Fx, picks: &[&str]) -> Value {
    result(&call(&apply_args(fx, picks, &token(fx), PERMIT)))
}

fn read_json(p: &Path) -> Value {
    serde_json::from_slice(&fs::read(p).unwrap_or_else(|e| panic!("{}: {e}", p.display()))).unwrap()
}

fn mode_of(p: &Path) -> u32 {
    fs::symlink_metadata(p).unwrap().permissions().mode() & 0o7777
}

/// The archive folder of the item `path` in an apply answer.
fn item_dir(answer: &Value, path: &str) -> PathBuf {
    PathBuf::from(text(&answer["archive_root"])).join(text(&item(answer, path)["archive_item"]))
}

/// The fixture with a nested folder, an empty folder and files of distinct modes in `run`.
fn apply_fixture() -> Fx {
    let fx = fixture();
    put(&fx.moai.join("run/deep/tool.sh"), 300);
    fs::create_dir_all(fx.moai.join("run/deep/empty")).unwrap();
    for (p, mode) in [
        ("run/deep/tool.sh", 0o750),
        ("run/pid", 0o600),
        ("run/deep/empty", 0o750),
    ] {
        fs::set_permissions(fx.moai.join(p), fs::Permissions::from_mode(mode)).unwrap();
    }
    age_tree(&fx.moai.join("run"), 400);
    fx
}

/// Every file below `p` as `(path below p, bytes)`, in path order.
fn contents(p: &Path) -> Vec<(PathBuf, Vec<u8>)> {
    fn walk(root: &Path, p: &Path, out: &mut Vec<(PathBuf, Vec<u8>)>) {
        if fs::symlink_metadata(p).unwrap().is_dir() {
            let mut kids: Vec<_> = fs::read_dir(p)
                .unwrap()
                .map(|e| e.unwrap().path())
                .collect();
            kids.sort();
            for k in kids {
                walk(root, &k, out);
            }
        } else {
            out.push((
                p.strip_prefix(root).unwrap().to_path_buf(),
                fs::read(p).unwrap(),
            ));
        }
    }
    let mut out = Vec::new();
    walk(p, p, &mut out);
    out
}

// ------------------------------------------------------------------- apply_

#[test]
fn apply_runs_only_the_picked_items_and_leaves_the_rest() {
    let fx = apply_fixture();
    let old1 = snap(&fx.moai.join("reports/projA/old1.md"));
    let rescue = snap(&fx.moai.join("rescue"));
    let r = applied(&fx, &["run", "reports/projA/old2.md"]);
    assert_eq!(r["state"], "committed", "{r}");
    assert_eq!(item(&r, "run")["state"], "archived", "{r}");
    assert_eq!(item(&r, "reports/projA/old2.md")["state"], "archived");
    assert_eq!(r["items"].as_array().unwrap().len(), 2, "{r}");
    assert!(!fx.moai.join("run").exists());
    assert!(!fx.moai.join("reports/projA/old2.md").exists());
    assert_eq!(snap(&fx.moai.join("reports/projA/old1.md")), old1);
    assert_eq!(snap(&fx.moai.join("rescue")), rescue);
    assert!(fx.moai.join("reports/projA/new.md").exists());
}

#[test]
fn apply_refuses_before_any_write_and_names_the_cause() {
    let fx = apply_fixture();
    let good = token(&fx);
    let cases: [(&[&str], &str, &str, &str); 6] = [
        (&[], &good, PERMIT, "no item is picked"),
        (&["run", "nope"], &good, PERMIT, "`nope`"),
        (&["run", "mystery"], &good, PERMIT, "`mystery`"),
        (&["run"], "sha256:00", PERMIT, "digest"),
        (&["run"], &good, "", "permission reference"),
        (&["run"], &good, "placeholder", "permission reference"),
    ];
    let before = snap(&fx.home);
    for (picks, confirm, permit, cause) in cases {
        let out = call(&apply_args(&fx, picks, confirm, permit));
        assert!(!out.ok, "{cause}: {}", all(&out));
        assert!(all(&out).contains(cause), "{cause}: {}", all(&out));
        assert_eq!(snap(&fx.home), before, "{cause}: the home changed");
    }
    // The token of an earlier plan does not run a changed folder.
    put(&fx.moai.join("run/later"), 5);
    let before = snap(&fx.home);
    let out = call(&apply_args(&fx, &["run"], &good, PERMIT));
    assert!(!out.ok && all(&out).contains("digest"), "{}", all(&out));
    assert_eq!(snap(&fx.home), before);
}

#[test]
fn apply_refuses_an_archive_folder_that_is_a_link_or_a_file() {
    for as_link in [true, false] {
        let fx = apply_fixture();
        let elsewhere = fx.home.join("elsewhere");
        fs::create_dir_all(&elsewhere).unwrap();
        if as_link {
            symlink(&elsewhere, fx.moai.join("archive")).unwrap();
        } else {
            put(&fx.moai.join("archive"), 3);
        }
        let good = token(&fx);
        let before = snap(&fx.home);
        let out = call(&apply_args(&fx, &["run"], &good, PERMIT));
        assert!(!out.ok, "{as_link}: {}", all(&out));
        assert!(all(&out).contains("archive"), "{}", all(&out));
        assert_eq!(snap(&fx.home), before, "{as_link}");
    }
}

#[test]
fn apply_refuses_a_moai_folder_that_is_a_link() {
    let fx = apply_fixture();
    let good = token(&fx);
    let real = fx.home.join("real");
    fs::rename(&fx.moai, &real).unwrap();
    symlink(&real, &fx.moai).unwrap();
    let before = snap(&fx.home);
    let out = call(&apply_args(&fx, &["run"], &good, PERMIT));
    assert!(!out.ok && all(&out).contains("link"), "{}", all(&out));
    assert_eq!(snap(&fx.home), before);
    // The fixture restores modes below `moai`; give it the real folder back.
    fs::remove_file(&fx.moai).unwrap();
    fs::rename(&real, &fx.moai).unwrap();
}

// ----------------------------------------------------------------- archive_

#[test]
fn archive_writes_a_private_list_of_every_file_before_the_original_goes() {
    let fx = apply_fixture();
    let tool_mtime = fs::metadata(fx.moai.join("run/deep/tool.sh"))
        .unwrap()
        .modified()
        .unwrap()
        .duration_since(SystemTime::UNIX_EPOCH)
        .unwrap();
    let r = applied(&fx, &["run"]);
    let root = PathBuf::from(text(&r["archive_root"]));
    assert_eq!(root.parent().unwrap(), fx.moai.join("archive"));
    assert_eq!(root.file_name().unwrap().to_str(), r["archive"].as_str());
    let dir = item_dir(&r, "run");
    assert_eq!(mode_of(&fx.moai.join("archive")), 0o700);
    assert_eq!(mode_of(&root), 0o700);
    assert_eq!(mode_of(&dir), 0o700);
    for f in ["list.json", "data.tar.gz"] {
        assert_eq!(mode_of(&dir.join(f)), 0o600, "{f}");
    }
    assert_eq!(mode_of(&root.join("receipt.json")), 0o600);
    let list = read_json(&dir.join("list.json"));
    assert_eq!(list["item"], "run", "{list}");
    assert_eq!(list["compression"], "gzip", "{list}");
    let rows = list["entries"].as_array().unwrap();
    let mut paths: Vec<&str> = rows.iter().map(|e| text(&e["path"])).collect();
    paths.sort();
    assert_eq!(
        paths,
        [
            "run",
            "run/deep",
            "run/deep/empty",
            "run/deep/tool.sh",
            "run/pid"
        ]
    );
    let tool = rows
        .iter()
        .find(|e| e["path"] == "run/deep/tool.sh")
        .unwrap();
    assert_eq!(tool["type"], "file", "{tool}");
    assert_eq!(tool["bytes"], 300, "{tool}");
    assert_eq!(tool["mode"], 0o750, "{tool}");
    assert_eq!(tool["sha256"], moai_cli::digest(&[b'x'; 300]), "{tool}");
    assert_eq!(tool["mtime_secs"], tool_mtime.as_secs(), "{tool}");
    assert!(!fx.moai.join("run").exists());
    // A gzip stream.
    let data = fs::read(dir.join("data.tar.gz")).unwrap();
    assert_eq!(&data[..2], &[0x1f, 0x8b]);
}

// --------------------------------------------------------------- protected_

#[test]
fn protected_items_with_a_secret_name_a_link_or_a_pipe_are_skipped_whole() {
    let fx = fixture();
    let outside = fx.home.join("outside");
    put(&outside.join("key.txt"), 16);
    put(&fx.moai.join("cache/a"), 5);
    put(&fx.moai.join("cache/deep/.env"), 5);
    fs::set_permissions(
        fx.moai.join("cache/deep/.env"),
        fs::Permissions::from_mode(0o000),
    )
    .unwrap();
    put(&fx.moai.join("run/b"), 5);
    symlink(outside.join("key.txt"), fx.moai.join("run/link")).unwrap();
    put(&fx.moai.join("scratch/c"), 5);
    assert!(Command::new("mkfifo")
        .arg(fx.moai.join("scratch/pipe"))
        .status()
        .unwrap()
        .success());
    put(&fx.moai.join("reports/projA/.env"), 5);
    set_age(&fx.moai.join("reports/projA/.env"), 300);
    // A report below a secret-named folder.
    put(&fx.moai.join("reports/secrets/old.md"), 5);
    set_age(&fx.moai.join("reports/secrets/old.md"), 300);
    // A file that cannot be read.
    put(&fx.moai.join("tmp-x/locked"), 5);
    fs::set_permissions(
        fx.moai.join("tmp-x/locked"),
        fs::Permissions::from_mode(0o000),
    )
    .unwrap();
    let picks = [
        "cache",
        "run",
        "scratch",
        "reports/projA/.env",
        "reports/secrets/old.md",
        "tmp-x",
    ];
    let kept =
        |fx: &Fx| -> Vec<Vec<String>> { picks.iter().map(|p| snap(&fx.moai.join(p))).collect() };
    let before = kept(&fx);
    let away = snap(&outside);
    let r = applied(&fx, &picks);
    assert_eq!(r["state"], "partial", "{r}");
    for (path, reason) in [
        ("cache", "secret_name"),
        ("run", "link"),
        ("scratch", "special_file"),
        ("reports/projA/.env", "secret_name"),
        ("reports/secrets/old.md", "secret_name"),
        ("tmp-x", "unreadable"),
    ] {
        let i = item(&r, path);
        assert_eq!(i["state"], "skipped", "{path}: {i}");
        assert!(text(&i["reason"]).starts_with(reason), "{path}: {i}");
    }
    assert_eq!(kept(&fx), before);
    assert_eq!(snap(&outside), away);
}

#[test]
fn protected_an_item_the_walk_could_not_finish_cannot_be_picked() {
    let fx = apply_fixture();
    // `run` holds more than one entry: the plan keeps it, so the apply refuses the pick.
    let p = plan(&fx, &[OVER[0], OVER[1], "--entry-limit", "1"]);
    assert_eq!(item(&p, "run")["action"], "keep", "{p}");
    let mut args = apply_args(&fx, &["run"], text(&p["digest"]), PERMIT);
    args.extend(["--entry-limit".to_string(), "1".to_string()]);
    let before = snap(&fx.home);
    let out = call(&args);
    assert!(!out.ok && all(&out).contains("`run`"), "{}", all(&out));
    assert_eq!(snap(&fx.home), before);
}

#[cfg(target_os = "macos")]
#[test]
fn protected_files_are_never_opened_by_the_apply() {
    let fx = fixture();
    let outside = fx.home.join("outside");
    put(&outside.join("key.txt"), 16);
    put(&fx.moai.join("cache/deep/.env"), 5);
    put(&fx.moai.join("run/b"), 5);
    symlink(outside.join("key.txt"), fx.moai.join("run/link")).unwrap();
    let denied = [fx.moai.join("cache/deep/.env"), outside.clone()];
    let args = apply_args(&fx, &["cache", "run"], &token(&fx), PERMIT);
    let args: Vec<&str> = args.iter().map(String::as_str).collect();
    let out = killed_on_open(&denied, BIN, &args);
    assert!(
        out.status.success(),
        "signal {:?}: {}",
        std::os::unix::process::ExitStatusExt::signal(&out.status),
        String::from_utf8_lossy(&out.stderr)
    );
}

// ------------------------------------------------------------------- inuse_

#[test]
fn inuse_a_go_session_that_may_be_live_keeps_the_retired_items() {
    let live = format!("[{{\"pid\": {}}}]", std::process::id());
    for (registry, reason) in [(live.as_str(), "in_use"), ("not json", "in_use_unknown")] {
        let fx = apply_fixture();
        fs::create_dir_all(fx.moai.join("state")).unwrap();
        fs::write(fx.moai.join("state/active-sessions.json"), registry).unwrap();
        let run_before = snap(&fx.moai.join("run"));
        let r = applied(&fx, &["run", "rescue", "reports/projA/old2.md"]);
        for path in ["run", "rescue"] {
            let i = item(&r, path);
            assert_eq!(i["state"], "skipped", "{reason}: {i}");
            assert_eq!(i["reason"], reason, "{i}");
        }
        assert_eq!(
            item(&r, "reports/projA/old2.md")["state"],
            "archived",
            "{r}"
        );
        assert_eq!(snap(&fx.moai.join("run")), run_before);
        assert!(fx.moai.join("rescue").is_dir());
    }
}

// -------------------------------------------------------------------- kept_

#[test]
fn kept_go_run_data_is_listed_as_kept_and_cannot_be_picked() {
    let fx = fixture();
    put(&fx.moai.join("backups/b.tar"), 9);
    put(&fx.moai.join("user-assets.json"), 9);
    let p = plan(&fx, &[]);
    for name in ["releases", "db", "backups", "user-assets.json"] {
        let i = item(&p, name);
        assert_eq!(i["action"], "keep", "{name}: {i}");
        assert!(text(&i["reason"]).contains("go_run_data_kept"), "{i}");
        let before = snap(&fx.home);
        let out = call(&apply_args(&fx, &[name], &token(&fx), PERMIT));
        assert!(!out.ok, "{name}: {}", all(&out));
        assert!(all(&out).contains(name), "{}", all(&out));
        assert_eq!(snap(&fx.home), before, "{name}");
    }
}

// ------------------------------------------------------------------ delete_

#[test]
fn delete_removes_the_empty_retired_folder_and_archives_nothing_for_it() {
    let fx = fixture();
    let r = applied(&fx, &["rescue"]);
    assert_eq!(r["state"], "committed", "{r}");
    let i = item(&r, "rescue");
    assert_eq!(i["action"], "delete", "{i}");
    assert_eq!(i["state"], "deleted", "{i}");
    assert!(!fx.moai.join("rescue").exists());
    let names: Vec<_> = fs::read_dir(text(&r["archive_root"]))
        .unwrap()
        .map(|e| e.unwrap().file_name().to_string_lossy().into_owned())
        .collect();
    assert_eq!(names, ["receipt.json"], "{names:?}");
}

// ----------------------------------------------------------------- restore_

#[test]
fn restore_puts_back_the_same_bytes_modes_and_times() {
    let fx = apply_fixture();
    let run = fx.moai.join("run");
    let report = fx.moai.join("reports/projA/old2.md");
    let before = (snap(&run), contents(&run));
    let report_before = (snap(&report), fs::read(&report).unwrap());
    let stamp = text(&applied(&fx, &["run", "reports/projA/old2.md"])["archive"]).to_string();
    let r = result(&call(&restore_args(&fx, &stamp, &[], PERMIT)));
    assert_eq!(r["state"], "restored", "{r}");
    assert_eq!(item(&r, "run")["state"], "restored", "{r}");
    assert_eq!((snap(&run), contents(&run)), before);
    assert_eq!((snap(&report), fs::read(&report).unwrap()), report_before);
    // The archive is kept.
    assert!(fx
        .moai
        .join("archive")
        .join(&stamp)
        .join("receipt.json")
        .exists());
}

#[test]
fn restore_runs_only_the_picked_item() {
    let fx = apply_fixture();
    let stamp = text(&applied(&fx, &["run", "reports/projA/old2.md"])["archive"]).to_string();
    let r = result(&call(&restore_args(
        &fx,
        &stamp,
        &["reports/projA/old2.md"],
        PERMIT,
    )));
    assert_eq!(r["items"].as_array().unwrap().len(), 1, "{r}");
    assert!(fx.moai.join("reports/projA/old2.md").exists());
    assert!(!fx.moai.join("run").exists());
}

#[test]
fn restore_names_what_exists_and_fills_only_what_is_missing() {
    let fx = apply_fixture();
    let stamp = text(&applied(&fx, &["run"])["archive"]).to_string();
    // As after a removal that stopped half way: one file is still there, with newer bytes.
    put(&fx.moai.join("run/pid"), 3);
    let r = result(&call(&restore_args(&fx, &stamp, &[], PERMIT)));
    assert_eq!(r["state"], "partial", "{r}");
    let i = item(&r, "run");
    assert_eq!(i["existing"], serde_json::json!(["run", "run/pid"]), "{i}");
    assert_eq!(fs::read(fx.moai.join("run/pid")).unwrap(), b"xxx");
    assert_eq!(
        fs::read(fx.moai.join("run/deep/tool.sh")).unwrap().len(),
        300
    );
}

#[test]
fn restore_writes_nothing_of_an_item_whose_list_or_archive_does_not_check() {
    for damage in [
        "escape",
        "secret",
        "missing_row",
        "flipped_byte",
        "missing_list",
        "linked_data",
    ] {
        let fx = apply_fixture();
        let r = applied(&fx, &["run"]);
        let stamp = text(&r["archive"]).to_string();
        let dir = item_dir(&r, "run");
        let mut list = read_json(&dir.join("list.json"));
        let rows = list["entries"].as_array_mut().unwrap();
        let file = rows.iter().position(|e| e["path"] == "run/pid").unwrap();
        match damage {
            "escape" => rows[file]["path"] = "run/../../escaped".into(),
            "secret" => rows[file]["path"] = "run/.env".into(),
            "missing_row" => {
                rows.remove(file);
            }
            "flipped_byte" => flip(&dir.join("data.tar.gz")),
            _ => {}
        }
        fs::write(dir.join("list.json"), serde_json::to_vec(&list).unwrap()).unwrap();
        match damage {
            "missing_list" => fs::remove_file(dir.join("list.json")).unwrap(),
            "linked_data" => {
                let copy = fx.home.join("data.copy");
                fs::rename(dir.join("data.tar.gz"), &copy).unwrap();
                symlink(&copy, dir.join("data.tar.gz")).unwrap();
            }
            _ => {}
        }
        let before = snap(&fx.home);
        let r = result(&call(&restore_args(&fx, &stamp, &[], PERMIT)));
        assert_eq!(r["state"], "partial", "{damage}: {r}");
        assert_eq!(item(&r, "run")["state"], "failed", "{damage}: {r}");
        assert_eq!(snap(&fx.home), before, "{damage}: something was written");
    }
}

#[test]
fn restore_follows_no_link_that_took_the_place_of_a_parent() {
    let fx = apply_fixture();
    let stamp = text(&applied(&fx, &["reports/projA/old2.md"])["archive"]).to_string();
    let outside = fx.home.join("outside");
    fs::create_dir_all(&outside).unwrap();
    fs::rename(fx.moai.join("reports/projA"), fx.home.join("moved")).unwrap();
    symlink(&outside, fx.moai.join("reports/projA")).unwrap();
    let before = snap(&fx.home);
    let r = result(&call(&restore_args(&fx, &stamp, &[], PERMIT)));
    assert_eq!(r["state"], "partial", "{r}");
    assert_eq!(item(&r, "reports/projA/old2.md")["state"], "failed", "{r}");
    assert_eq!(snap(&fx.home), before, "something was written");
}

#[test]
fn restore_refuses_a_bad_pick_stamp_or_permission_before_any_write() {
    let fx = apply_fixture();
    let stamp = text(&applied(&fx, &["run"])["archive"]).to_string();
    let before = snap(&fx.home);
    for (stamp, picks, permit, cause) in [
        (stamp.as_str(), vec!["nope"], PERMIT, "`nope`"),
        ("../reports", vec![], PERMIT, "archive"),
        ("20000101T000000Z", vec![], PERMIT, "archive"),
        (stamp.as_str(), vec![], "null", "permission reference"),
    ] {
        let out = call(&restore_args(&fx, stamp, &picks, permit));
        assert!(!out.ok, "{cause}: {}", all(&out));
        assert!(all(&out).contains(cause), "{cause}: {}", all(&out));
        assert_eq!(snap(&fx.home), before, "{cause}");
    }
    // An archive folder that is a link.
    let archive = fx.moai.join("archive");
    symlink(archive.join(&stamp), archive.join("20000101T000000Z")).unwrap();
    let before = snap(&fx.home);
    let out = call(&restore_args(&fx, "20000101T000000Z", &[], PERMIT));
    assert!(!out.ok && all(&out).contains("archive"), "{}", all(&out));
    assert_eq!(snap(&fx.home), before);
}

// ----------------------------------------------------------------- receipt_

#[test]
fn receipt_names_each_item_and_the_answer_has_the_free_bytes_without_any_command() {
    let fx = apply_fixture();
    let tools = TempDir::new().unwrap();
    let bin = fs::canonicalize(tools.path()).unwrap();
    let log = bin.join("calls.log");
    for name in ["claude", "git", "tar", "gzip", "df", "rm"] {
        let script = bin.join(name);
        fs::write(
            &script,
            format!(
                "#!/bin/sh\necho \"{name} $*\" >> \"{}\"\nexit 1\n",
                log.display()
            ),
        )
        .unwrap();
        fs::set_permissions(&script, fs::Permissions::from_mode(0o755)).unwrap();
    }
    let args = apply_args(&fx, &["run", "rescue"], &token(&fx), PERMIT);
    let args: Vec<&str> = args.iter().map(String::as_str).collect();
    let r = result(&run_with(Some(&bin), &args));
    assert!(r["free_bytes"]["before"].is_u64(), "{r}");
    assert!(r["free_bytes"]["after"].is_u64(), "{r}");
    let receipt = read_json(&PathBuf::from(text(&r["archive_root"])).join("receipt.json"));
    assert_eq!(receipt["status"], "committed", "{receipt}");
    assert_eq!(receipt["permission_reference"], PERMIT, "{receipt}");
    assert_eq!(item(&receipt, "run")["state"], "archived", "{receipt}");
    assert_eq!(item(&receipt, "rescue")["state"], "deleted", "{receipt}");
    let stamp = text(&r["archive"]).to_string();
    let args = restore_args(&fx, &stamp, &[], PERMIT);
    let args: Vec<&str> = args.iter().map(String::as_str).collect();
    result(&run_with(Some(&bin), &args));
    assert!(!log.exists(), "{}", fs::read_to_string(&log).unwrap());
}

// ------------------------------------------- a run changed half way (SPEC D-4)

/// Flips one byte in the middle of a file.
fn flip(p: &Path) {
    let mut data = fs::read(p).unwrap();
    let at = data.len() / 2;
    data[at] ^= 0xff;
    fs::write(p, data).unwrap();
}

/// `tidy-apply` through the library over the temporary home, with `at` called at each step.
fn stepped(fx: &Fx, picks: &[&str], at: &mut dyn FnMut(Step<'_>)) -> Value {
    let opts = moai_cli::tidy::Options::new(
        &[OVER[1].to_string()],
        Vec::new(),
        moai_cli::tidy::DEFAULT_ENTRY_LIMIT,
        moai_cli::tidy::DEFAULT_TIME_LIMIT_SECS,
    )
    .unwrap();
    let picks: Vec<String> = picks.iter().map(|p| (*p).to_string()).collect();
    let confirm = token(fx);
    let request = Request {
        picks: &picks,
        confirm: &confirm,
        permission: PERMIT,
    };
    apply_with(&fx.home, &opts, &request, at).unwrap()
}

#[test]
fn archive_a_changed_archive_or_source_removes_nothing_of_the_item() {
    for damage in ["flipped_byte", "changed_source", "removed_source"] {
        let fx = apply_fixture();
        let run = fx.moai.join("run");
        let mut then = None;
        let mut folder = PathBuf::new();
        let r = stepped(&fx, &["run", "reports/projA/old2.md"], &mut |step| {
            if let Step::Written { item: "run", dir } = step {
                folder = dir.to_path_buf();
                match damage {
                    "flipped_byte" => flip(&dir.join("data.tar.gz")),
                    "changed_source" => put(&run.join("pid"), 11),
                    _ => fs::remove_file(run.join("pid")).unwrap(),
                }
                then = Some((snap(&run), contents(&run)));
            }
        });
        assert_eq!(item(&r, "run")["state"], "failed", "{damage}: {r}");
        assert_eq!(r["state"], "partial", "{damage}: {r}");
        assert_eq!(
            Some((snap(&run), contents(&run))),
            then,
            "{damage}: the original changed"
        );
        // Only the unfinished files of this item are gone; the other item is archived.
        assert!(!folder.exists(), "{damage}: {}", folder.display());
        assert_eq!(item(&r, "reports/projA/old2.md")["state"], "archived");
        assert!(item_dir(&r, "reports/projA/old2.md")
            .join("data.tar.gz")
            .exists());
        let receipt = read_json(&folder.parent().unwrap().join("receipt.json"));
        assert_eq!(item(&receipt, "run")["state"], "failed", "{receipt}");
        assert_eq!(receipt["status"], "partial", "{receipt}");
    }
}

#[test]
fn archive_a_file_that_appears_after_the_listing_stays() {
    let fx = apply_fixture();
    let late = fx.moai.join("run/deep/late.txt");
    let r = stepped(&fx, &["run"], &mut |step| {
        if let Step::Written { .. } = step {
            put(&late, 4);
        }
    });
    assert_eq!(r["state"], "partial", "{r}");
    let i = item(&r, "run");
    assert_eq!(i["state"], "archived", "{i}");
    assert_eq!(
        i["left"],
        serde_json::json!(["run", "run/deep", "run/deep/late.txt"]),
        "{i}"
    );
    assert_eq!(fs::read(&late).unwrap(), b"xxxx");
    for gone in ["run/pid", "run/deep/tool.sh", "run/deep/empty"] {
        assert!(!fx.moai.join(gone).exists(), "{gone}");
    }
}

#[test]
fn archive_a_file_changed_just_before_its_removal_stays() {
    let fx = apply_fixture();
    let pid = fx.moai.join("run/pid");
    let r = stepped(&fx, &["run"], &mut |step| {
        if let Step::Removing { .. } = step {
            put(&pid, 11);
        }
    });
    assert_eq!(r["state"], "partial", "{r}");
    let i = item(&r, "run");
    assert_eq!(i["state"], "partly_removed", "{i}");
    assert_eq!(i["left"], serde_json::json!(["run", "run/pid"]), "{i}");
    assert_eq!(fs::read(&pid).unwrap().len(), 11);
    assert!(!fx.moai.join("run/deep").exists());
}

#[test]
fn protected_a_folder_swapped_for_a_link_is_not_followed_at_removal() {
    let fx = apply_fixture();
    // Same names, bytes and mode as the listed ones.
    let outside = fx.home.join("outside");
    put(&outside.join("tool.sh"), 300);
    fs::set_permissions(outside.join("tool.sh"), fs::Permissions::from_mode(0o750)).unwrap();
    fs::create_dir_all(outside.join("empty")).unwrap();
    let deep = fx.moai.join("run/deep");
    let moved = fx.home.join("moved");
    let mut away = Vec::new();
    let r = stepped(&fx, &["run"], &mut |step| {
        if let Step::Removing { .. } = step {
            fs::rename(&deep, &moved).unwrap();
            symlink(&outside, &deep).unwrap();
            away = snap(&outside);
        }
    });
    assert_eq!(item(&r, "run")["state"], "partly_removed", "{r}");
    assert_eq!(snap(&outside), away);
    assert!(moved.join("tool.sh").exists() && moved.join("empty").exists());
    assert!(fs::symlink_metadata(&deep)
        .unwrap()
        .file_type()
        .is_symlink());
}

#[test]
fn delete_keeps_a_folder_that_got_a_file_just_before_the_removal() {
    let fx = fixture();
    let new = fx.moai.join("rescue/new");
    let r = stepped(&fx, &["rescue"], &mut |step| {
        if let Step::Removing { .. } = step {
            put(&new, 2);
        }
    });
    assert_eq!(r["state"], "partial", "{r}");
    let i = item(&r, "rescue");
    assert_eq!(i["state"], "kept", "{i}");
    assert!(text(&i["reason"]).starts_with("not_empty"), "{i}");
    assert_eq!(fs::read(&new).unwrap(), b"xx");
}

#[test]
fn receipt_is_pending_before_the_first_removal() {
    let fx = apply_fixture();
    let run = fx.moai.join("run");
    let before = snap(&run);
    let mut seen = 0;
    stepped(&fx, &["run", "rescue"], &mut |step| {
        if let Step::Removing { run: folder, .. } = step {
            if seen == 0 {
                let receipt = read_json(&folder.join("receipt.json"));
                assert_eq!(receipt["status"], "pending", "{receipt}");
                assert!(fx.moai.join("rescue").is_dir());
            }
            seen += 1;
        }
    });
    assert_eq!(seen, 2);
    assert_ne!(snap_if_present(&run), before);
}

fn snap_if_present(p: &Path) -> Vec<String> {
    if p.exists() {
        snap(p)
    } else {
        Vec::new()
    }
}

#[test]
fn inuse_a_held_session_lock_or_a_linked_state_folder_keeps_the_retired_items() {
    for (case, reason) in [("held", "in_use"), ("linked", "in_use_unknown")] {
        let fx = apply_fixture();
        let mut lock = None;
        if case == "held" {
            put(&fx.moai.join("state/active-sessions.json.lock"), 0);
            let f = fs::File::open(fx.moai.join("state/active-sessions.json.lock")).unwrap();
            f.lock().unwrap();
            lock = Some(f);
        } else {
            let elsewhere = fx.home.join("state");
            fs::create_dir_all(&elsewhere).unwrap();
            symlink(&elsewhere, fx.moai.join("state")).unwrap();
        }
        let r = applied(&fx, &["run"]);
        assert_eq!(item(&r, "run")["reason"], reason, "{case}: {r}");
        assert!(fx.moai.join("run/pid").exists(), "{case}");
        drop(lock);
    }
}
