//! SPEC-20261124-home-tidy-slice2: `tidy-inventory` and `tidy-plan` read the home moai folder and
//! write nothing. Every test runs in a crafted temporary home; the real HOME is never touched.
//!
//! Stand-ins for "never opens a protected entry": a secret-named file and a secret-named folder
//! the test made unreadable (mode 000), a symlink, a dangling symlink and a named pipe. A tool
//! that opened the pipe would block, and the helper below kills a command that runs too long.
#![cfg(unix)]

mod common;
use common::CleanEnv;

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
    assert_eq!(item(&p, "releases")["action"], "archive");
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
        assert_eq!(deletes, vec!["db"], "{extra:?}");
    }
    // The same folder with a file in it is archived, not deleted.
    put(&fx.moai.join("db/leftover"), 3);
    let p = plan(&fx, &[]);
    assert_eq!(item(&p, "db")["action"], "archive");
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
