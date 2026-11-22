//! Issue #2455 (user decision 2026-11-11): the statusline reads its HOME state files
//! only from `$HOME/.moai-cli/state/`, never from the Go folder `$HOME/.moai/`, and
//! `MOAI_HOME` (the Go tool's variable) no longer redirects it. `MOAI_LAST_MODEL` is an
//! environment value, not a file read, and keeps its precedence over the model file.
//!
//! Every case runs the real binary with a cleared environment and a temporary HOME; the
//! working directory is a separate temporary folder outside that HOME. Paths are spelled
//! literally on purpose: a test that went through the product constant would follow a
//! regression instead of catching it.

mod common;
use common::CleanEnv;
use std::io::Write;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};

struct Fixture {
    _tmp: tempfile::TempDir,
    home: PathBuf,
    work: PathBuf,
}

fn fixture() -> Fixture {
    let tmp = tempfile::tempdir().unwrap();
    let home = tmp.path().join("home");
    let work = tmp.path().join("work");
    std::fs::create_dir_all(&home).unwrap();
    std::fs::create_dir_all(&work).unwrap();
    Fixture {
        _tmp: tmp,
        home,
        work,
    }
}

fn write(base: &Path, rel: &str, bytes: &[u8]) {
    let path = base.join(rel);
    std::fs::create_dir_all(path.parent().unwrap()).unwrap();
    std::fs::write(path, bytes).unwrap();
}

/// One plain render with stdin `{}`: no model in the input, so the model segment can only
/// come from `MOAI_LAST_MODEL` or the HOME model file.
fn render(f: &Fixture, extra: &[(&str, &str)]) -> String {
    render_in(f, &f.work, extra)
}

fn render_in(f: &Fixture, cwd: &Path, extra: &[(&str, &str)]) -> String {
    // The render may keep a session cache under `$TMPDIR`: a folder of its own outside the
    // HOME, so nothing reaches the real temporary folder.
    let tmp = tempfile::tempdir().unwrap();
    let mut cmd = Command::new(crate::common::MOAI_BIN);
    cmd.args(["statusline", "--render-budget-ms", "3600000"])
        .current_dir(cwd)
        .clean_env()
        .env("PATH", std::env::var_os("PATH").unwrap_or_default())
        .env("HOME", &f.home)
        .env("TMPDIR", tmp.path())
        .env("NO_COLOR", "1")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped());
    for (k, v) in extra {
        cmd.env(k, v);
    }
    let mut child = cmd.spawn().unwrap();
    child.stdin.take().unwrap().write_all(b"{}").unwrap();
    let out = child.wait_with_output().unwrap();
    assert!(
        out.status.success(),
        "statusline exit {:?}: {}",
        out.status.code(),
        String::from_utf8_lossy(&out.stderr)
    );
    String::from_utf8(out.stdout).unwrap()
}

const GO_MODEL: &[u8] = b"go-marker-model\n";
const GO_TASK: &[u8] =
    br#"{"active_task": {"command": "gomarker", "spec_id": "SPEC-GO-1", "stage": "red"}}"#;
const RC_MODEL: &[u8] = b"rc-marker-model\n";
const RC_TASK: &[u8] =
    br#"{"active_task": {"command": "rcmarker", "spec_id": "SPEC-RC-1", "stage": "green"}}"#;

fn write_go_state(f: &Fixture) {
    write(&f.home, ".moai/state/last-model.txt", GO_MODEL);
    write(&f.home, ".moai/state/last-session-state.json", GO_TASK);
}

fn write_rc_state(f: &Fixture) {
    write(&f.home, ".moai-cli/state/last-model.txt", RC_MODEL);
    write(&f.home, ".moai-cli/state/last-session-state.json", RC_TASK);
}

fn assert_no_go_markers(out: &str) {
    assert!(
        !out.contains("go-marker-model"),
        "Go model file was read: {out}"
    );
    assert!(
        !out.contains("gomarker"),
        "Go session state was read: {out}"
    );
}

/// A sorted listing of every entry below `dir` with its bytes, to prove nothing changed.
fn snapshot(dir: &Path) -> Vec<(PathBuf, Vec<u8>)> {
    let mut out = Vec::new();
    let mut stack = vec![dir.to_path_buf()];
    while let Some(d) = stack.pop() {
        for e in std::fs::read_dir(&d).unwrap() {
            let p = e.unwrap().path();
            if p.is_dir() {
                stack.push(p.clone());
                out.push((p, Vec::new()));
            } else {
                out.push((p.clone(), std::fs::read(&p).unwrap()));
            }
        }
    }
    out.sort();
    out
}

/// AC-1: only the Go files exist — neither marker appears, with or without `MOAI_HOME`
/// pointing at the Go folder, and the Go folder is left byte for byte as it was.
#[test]
fn go_home_state_files_are_never_read() {
    let f = fixture();
    write_go_state(&f);
    let before = snapshot(&f.home);

    let out = render(&f, &[]);
    assert_no_go_markers(&out);
    assert!(!out.contains("🧠"), "{out}");
    assert!(!out.contains("📋"), "{out}");

    let go_home = f.home.join(".moai");
    let out = render(&f, &[("MOAI_HOME", go_home.to_str().unwrap())]);
    assert_no_go_markers(&out);
    assert!(!out.contains("📋"), "{out}");

    assert_eq!(snapshot(&f.home), before, "HOME changed during render");
}

/// AC-2: the same files under `.moai-cli/state/` are read. With both locations filled the
/// RC files win and the Go markers stay out. The session task (📋) is never shown since the new
/// three-line layout (#2376, user decision 3), even when its state file is present.
#[test]
fn moai_cli_home_state_files_are_read() {
    let f = fixture();
    write_rc_state(&f);
    let out = render(&f, &[]);
    assert!(out.contains("🧠 rc-marker-model"), "{out}");
    assert!(!out.contains("📋") && !out.contains("rcmarker"), "{out}");

    write_go_state(&f);
    let go_home = f.home.join(".moai");
    for extra in [vec![], vec![("MOAI_HOME", go_home.to_str().unwrap())]] {
        let out = render(&f, &extra);
        assert!(out.contains("🧠 rc-marker-model"), "{out}");
        assert!(!out.contains("📋") && !out.contains("rcmarker"), "{out}");
        assert_no_go_markers(&out);
    }
}

/// `MOAI_LAST_MODEL` is kept and still wins over the model file.
#[test]
fn moai_last_model_still_overrides_the_model_file() {
    let f = fixture();
    write_rc_state(&f);
    let out = render(&f, &[("MOAI_LAST_MODEL", "env-marker-model")]);
    assert!(out.contains("🧠 env-marker-model"), "{out}");
    assert!(!out.contains("rc-marker-model"), "{out}");
}

/// AC-2 boundaries for the model file: absent, 0 bytes, whitespace only. Each renders
/// (exit 0) without a model segment; one visible character is enough to show it.
#[test]
fn model_file_boundaries() {
    let f = fixture();
    let out = render(&f, &[]);
    assert!(!out.contains("🧠"), "absent: {out}");

    for (case, bytes) in [("empty", &b""[..]), ("whitespace", &b" \n\t\n"[..])] {
        write(&f.home, ".moai-cli/state/last-model.txt", bytes);
        let out = render(&f, &[]);
        assert!(!out.contains("🧠"), "{case}: {out}");
    }

    write(&f.home, ".moai-cli/state/last-model.txt", b"m");
    let out = render(&f, &[]);
    assert!(out.contains("🧠 m"), "one character: {out}");
}

/// AC-2 boundaries for the session state: absent, 0 bytes, broken JSON, no `active_task`,
/// an empty command, a minimal command. Each renders (exit 0), and none shows a task segment —
/// the new layout never shows 📋 (#2376, user decision 3).
#[test]
fn session_state_boundaries() {
    let f = fixture();
    let out = render(&f, &[]);
    assert!(!out.contains("📋"), "absent: {out}");

    for (case, bytes) in [
        ("empty", &b""[..]),
        ("broken", &br#"{"active_task": {"command": "rcmarker""#[..]),
        ("no active_task", &br#"{"current_branch": "main"}"#[..]),
        ("empty command", &br#"{"active_task": {"command": ""}}"#[..]),
    ] {
        write(&f.home, ".moai-cli/state/last-session-state.json", bytes);
        let out = render(&f, &[]);
        assert!(!out.contains("📋"), "{case}: {out}");
    }

    write(
        &f.home,
        ".moai-cli/state/last-session-state.json",
        br#"{"active_task": {"command": "r"}}"#,
    );
    let out = render(&f, &[]);
    assert!(
        !out.trim().is_empty(),
        "minimal command still renders: {out}"
    );
    assert!(!out.contains("📋"), "minimal command: {out}");
}

/// AC-5 (leader scope 6792437161): with the cwd deep below HOME and only
/// `HOME/.moai/config/sections/system.yaml` present, its template version never reaches
/// the output; at HOME itself neither. A project folder directly below HOME is still read
/// (#2389 project fallback).
#[test]
fn home_dot_moai_is_not_project_configuration() {
    let f = fixture();
    write(
        &f.home,
        ".moai/config/sections/system.yaml",
        b"moai:\n  template_version: v9.9.9\n",
    );
    let deep = f.home.join("a/b/c");
    std::fs::create_dir_all(&deep).unwrap();
    for cwd in [&deep, &f.home] {
        let out = render_in(&f, cwd, &[]);
        assert!(!out.contains("9.9.9"), "{cwd:?}: {out}");
    }

    let proj = f.home.join("proj");
    write(
        &proj,
        ".moai/config/sections/system.yaml",
        b"moai:\n  template_version: v1.2.3\n",
    );
    let sub = proj.join("sub");
    std::fs::create_dir_all(&sub).unwrap();
    for cwd in [&proj, &sub] {
        let out = render_in(&f, cwd, &[]);
        assert!(out.contains("1.2.3"), "{cwd:?}: {out}");
        assert!(!out.contains("9.9.9"), "{cwd:?}: {out}");
    }
}

// ---- SPEC-20261122-statusline-limits REQ-7, AC-5: the usage record ----
//
// The record of the usage windows lives in the status line's cache folder under `TMPDIR`, one
// file per HOME named `usage-<12 hex of the SHA-256 of the canonical HOME path>.record`. The file
// is found here by that name anywhere below `TMPDIR`, not through the product's own function.

fn unix_now() -> i64 {
    std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .unwrap()
        .as_secs() as i64
}

/// A subscriber's stdin with the five-hour window at `used` percent, resetting at `resets_at`.
fn limits_stdin(used: u32, resets_at: i64) -> String {
    format!(
        r#"{{"model":{{"display_name":"Opus"}},"rate_limits":{{"five_hour":{{"used_percentage":{used},"resets_at":{resets_at}}}}}}}"#
    )
}

/// One render of `stdin` with HOME `home`, the given `TMPDIR` and working directory `cwd`.
fn render_with(home: &Path, cwd: &Path, tmpdir: impl AsRef<std::ffi::OsStr>, stdin: &str) {
    let mut child = Command::new(crate::common::MOAI_BIN)
        .args(["statusline", "--render-budget-ms", "3600000"])
        .current_dir(cwd)
        .clean_env()
        .env("PATH", std::env::var_os("PATH").unwrap_or_default())
        .env("HOME", home)
        .env("TMPDIR", tmpdir)
        .env("NO_COLOR", "1")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    child
        .stdin
        .take()
        .unwrap()
        .write_all(stdin.as_bytes())
        .unwrap();
    let out = child.wait_with_output().unwrap();
    assert!(
        out.status.success(),
        "statusline exit {:?}: {}",
        out.status.code(),
        String::from_utf8_lossy(&out.stderr)
    );
}

/// The file name the record of `home` has: `usage-` and the first 12 hex digits of the SHA-256
/// of the canonical path.
fn record_name(home: &Path) -> String {
    use sha2::{Digest, Sha256};
    let canonical = std::fs::canonicalize(home).unwrap();
    let hex = format!(
        "{:x}",
        Sha256::digest(canonical.as_os_str().as_encoded_bytes())
    );
    format!("usage-{}.record", &hex[..12])
}

/// The one record of `home` below `tmpdir`; it must sit in a folder directly under `tmpdir`.
fn record_file(home: &Path, tmpdir: &Path) -> PathBuf {
    let name = record_name(home);
    let found: Vec<PathBuf> = snapshot(tmpdir)
        .into_iter()
        .map(|(p, _)| p)
        .filter(|p| p.file_name().is_some_and(|n| n.to_string_lossy() == name))
        .collect();
    assert_eq!(found.len(), 1, "one {name} below {tmpdir:?}: {found:?}");
    assert_eq!(
        found[0].parent().and_then(Path::parent),
        Some(tmpdir),
        "the record sits in the cache folder directly under TMPDIR"
    );
    found[0].clone()
}

/// `moai ARGS` with the given `TMPDIR`, `input` on stdin; the `result` of its envelope.
fn moai_result(
    args: &[&str],
    tmpdir: impl AsRef<std::ffi::OsStr>,
    input: &str,
) -> serde_json::Value {
    let mut child = Command::new(crate::common::MOAI_BIN)
        .args(args)
        .clean_env()
        .env("PATH", "/usr/bin:/bin")
        .env("TMPDIR", tmpdir)
        .env("LC_ALL", "C")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    child
        .stdin
        .take()
        .unwrap()
        .write_all(input.as_bytes())
        .unwrap();
    let out = child.wait_with_output().unwrap();
    let v: serde_json::Value = serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
        panic!(
            "{}{}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    });
    assert_eq!(v["ok"], true, "{v}");
    v["result"].clone()
}

/// AC-5: a render with a percentage the record does not hold writes the record in the cache
/// folder (0600 in a 0700 folder); a new percentage replaces it.
#[test]
fn usage_record_a_changed_percentage_is_written_in_the_cache_folder() {
    let f = fixture();
    let tmp = tempfile::tempdir().unwrap();
    let reset = unix_now() + 3600;
    render_with(&f.home, &f.work, tmp.path(), &limits_stdin(42, reset));
    let path = record_file(&f.home, tmp.path());
    let first = std::fs::read(&path).unwrap();
    assert!(String::from_utf8_lossy(&first).contains("42"), "{path:?}");
    #[cfg(unix)]
    {
        use std::os::unix::fs::PermissionsExt;
        let mode = |p: &Path| std::fs::symlink_metadata(p).unwrap().permissions().mode() & 0o777;
        assert_eq!(mode(&path), 0o600);
        assert_eq!(mode(path.parent().unwrap()), 0o700);
    }
    render_with(&f.home, &f.work, tmp.path(), &limits_stdin(43, reset));
    let second = std::fs::read(record_file(&f.home, tmp.path())).unwrap();
    assert_ne!(first, second);
    assert!(String::from_utf8_lossy(&second).contains("43"));
}

/// AC-5: the same percentage and reset time leave the record as it is (bytes, modified time and,
/// on Unix, the file itself); a new reset time alone writes it again.
#[test]
fn usage_record_an_unchanged_render_does_not_rewrite_it() {
    let f = fixture();
    let tmp = tempfile::tempdir().unwrap();
    let reset = unix_now() + 3600;
    render_with(&f.home, &f.work, tmp.path(), &limits_stdin(42, reset));
    let path = record_file(&f.home, tmp.path());
    let old = std::time::UNIX_EPOCH + std::time::Duration::from_secs(1_000_000_000);
    std::fs::File::options()
        .write(true)
        .open(&path)
        .unwrap()
        .set_modified(old)
        .unwrap();
    let bytes = std::fs::read(&path).unwrap();
    #[cfg(unix)]
    let inode = std::os::unix::fs::MetadataExt::ino(&std::fs::metadata(&path).unwrap());

    render_with(&f.home, &f.work, tmp.path(), &limits_stdin(42, reset));
    let meta = std::fs::metadata(&path).unwrap();
    assert_eq!(
        meta.modified().unwrap(),
        old,
        "an unchanged record was rewritten"
    );
    assert_eq!(std::fs::read(&path).unwrap(), bytes);
    #[cfg(unix)]
    assert_eq!(std::os::unix::fs::MetadataExt::ino(&meta), inode);

    render_with(&f.home, &f.work, tmp.path(), &limits_stdin(42, reset + 60));
    assert_ne!(
        std::fs::metadata(&path).unwrap().modified().unwrap(),
        old,
        "a new reset time is a change"
    );
}

/// AC-5: renders with and without usage windows leave the HOME tree as it was.
#[test]
fn usage_record_leaves_the_home_tree_as_it_was() {
    let f = fixture();
    let tmp = tempfile::tempdir().unwrap();
    let before = snapshot(&f.home);
    let reset = unix_now() + 3600;
    for stdin in [
        limits_stdin(42, reset),
        limits_stdin(95, reset),
        limits_stdin(95, reset),
        r#"{"model":{"display_name":"Opus"}}"#.to_string(),
    ] {
        render_with(&f.home, &f.work, tmp.path(), &stdin);
    }
    assert_eq!(snapshot(&f.home), before, "HOME changed during render");
    record_file(&f.home, tmp.path());
}

/// AC-5: both lane readers find the new record through the lane's HOME and the same `TMPDIR`:
/// `usage-limit` says `near` at 90 and `lane-screen --home` takes the reset of a window at 100.
/// Another `TMPDIR` holds no record, so the answer is `unknown` there.
#[test]
fn usage_record_is_read_by_the_lane_readers() {
    let f = fixture();
    let tmp = tempfile::tempdir().unwrap();
    let home = f.home.to_str().unwrap();
    let reset = unix_now() + 3600;
    render_with(&f.home, &f.work, tmp.path(), &limits_stdin(90, reset));
    let v = moai_result(&["usage-limit", home], tmp.path(), "");
    assert_eq!(v["state"], "near", "{v}");
    assert_eq!(v["window"], "five_hour", "{v}");
    let other = tempfile::tempdir().unwrap();
    let v = moai_result(&["usage-limit", home], other.path(), "");
    assert_eq!(v["state"], "unknown", "another TMPDIR: {v}");

    render_with(&f.home, &f.work, tmp.path(), &limits_stdin(100, reset));
    let screen = "Your limit will reset at 2026-11-19 17:30:55\n";
    let v = moai_result(&["lane-screen", "--home", home], tmp.path(), screen);
    assert_eq!(v["state"], "usage_limit", "{v}");
    assert_eq!(v["reset"]["source"], "home", "{v}");
    let v = moai_result(&["lane-screen", "--home", home], other.path(), screen);
    assert_ne!(v["reset"]["source"], "home", "another TMPDIR: {v}");
}

/// REQ-7: the session cache's clean-up of old files in the same folder leaves the usage record,
/// however old: a record that has not changed for two days is still there after a render that
/// writes a session cache and cleans, and the lane reader still finds it.
#[test]
fn usage_record_survives_the_session_cache_clean_up() {
    let f = fixture();
    common::git::git_ok(&f.work, &["init", "-q", "-b", "main"]);
    let tmp = tempfile::tempdir().unwrap();
    let reset = unix_now() + 3600;
    render_with(&f.home, &f.work, tmp.path(), &limits_stdin(90, reset));
    let path = record_file(&f.home, tmp.path());
    let two_days_ago = std::time::SystemTime::now() - std::time::Duration::from_secs(2 * 86_400);
    std::fs::File::options()
        .write(true)
        .open(&path)
        .unwrap()
        .set_modified(two_days_ago)
        .unwrap();
    // A render with a session id in a repository writes its session cache and cleans old files.
    render_with(
        &f.home,
        &f.work,
        tmp.path(),
        r#"{"session_id":"s-clean","model":{"display_name":"Opus"}}"#,
    );
    let session = path.parent().unwrap().join("s-clean.json");
    assert!(session.is_file(), "the session cache was written");
    assert!(path.is_file(), "the clean-up removed the usage record");
    let v = moai_result(&["usage-limit", f.home.to_str().unwrap()], tmp.path(), "");
    assert_eq!(v["state"], "near", "{v}");
}

/// REQ-7 / the cache folder rule: a relative `TMPDIR` names no cache folder, so a render writes
/// nothing (neither in the working directory it would resolve against nor under HOME) and the
/// reader with the same `TMPDIR` finds no record.
#[test]
fn usage_record_a_relative_tmpdir_writes_nothing() {
    let f = fixture();
    let home_before = snapshot(&f.home);
    let work_before = snapshot(&f.work);
    render_with(
        &f.home,
        &f.work,
        "rel-tmp",
        &limits_stdin(95, unix_now() + 3600),
    );
    assert_eq!(snapshot(&f.home), home_before, "HOME changed");
    assert_eq!(
        snapshot(&f.work),
        work_before,
        "the working directory changed"
    );
    let v = moai_result(&["usage-limit", f.home.to_str().unwrap()], "rel-tmp", "");
    assert_eq!(v["state"], "unknown", "{v}");
}
