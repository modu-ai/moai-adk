//! WT-1 (#2380): `moai-cli worktree-report` on a synthetic repository and a synthetic
//! temporary root. The real `/tmp` and the real project are never scanned.
#![cfg(unix)]
mod common;

use serde_json::Value;
use std::collections::BTreeSet;
use std::fs;
use std::path::{Path, PathBuf};
use std::process::Command;

const BIN: &str = env!("CARGO_BIN_EXE_moai-cli");

fn git(dir: &Path, args: &[&str]) -> String {
    let mut all = vec![
        "-c",
        "user.name=fixture",
        "-c",
        "user.email=fixture@example.invalid",
    ];
    all.extend_from_slice(args);
    common::git::git(dir, &all)
}

/// One 64 KiB file of fixed bytes.
fn blob(path: &Path) {
    fs::write(path, vec![0x5au8; 65_536]).unwrap();
}

struct Fixture {
    _dir: tempfile::TempDir,
    root: PathBuf,
    tmp: PathBuf,
    home: PathBuf,
}

/// A repository with two linked worktrees nested in `.moai-cli/worktrees/`:
/// - `a`: branch `issue/1-a`, locked with a reason, a `target/` holding one 64 KiB file and
///   a hard link to it, plus a 256 KiB ignored file in its checkout;
/// - `b`: detached, a `target/` holding one 64 KiB file, and an untracked file (dirty).
///
/// Both targets carry `CACHEDIR.TAG`. The temporary root holds one folder per class plus a
/// folder that is not `moai-*`.
fn fixture() -> Fixture {
    let dir = tempfile::TempDir::new().unwrap();
    let base = fs::canonicalize(dir.path()).unwrap();
    let root = base.join("repo");
    let tmp = base.join("tmp");
    let home = base.join("home");
    for d in [&root, &tmp, &home] {
        fs::create_dir_all(d).unwrap();
    }
    git(&root, &["init", "-q", "-b", "main"]);
    fs::write(root.join(".gitignore"), ".moai-cli/\ntarget/\n*.bin\n").unwrap();
    fs::write(root.join("a.txt"), "a\n").unwrap();
    git(&root, &["add", ".gitignore", "a.txt"]);
    git(&root, &["commit", "-q", "-m", "init"]);
    let a = root.join(".moai-cli/worktrees/a");
    let b = root.join(".moai-cli/worktrees/b");
    git(
        &root,
        &[
            "worktree",
            "add",
            "-q",
            "-b",
            "issue/1-a",
            a.to_str().unwrap(),
        ],
    );
    git(
        &root,
        &[
            "worktree",
            "lock",
            "--reason",
            "moai claim t e1",
            a.to_str().unwrap(),
        ],
    );
    git(
        &root,
        &[
            "worktree",
            "add",
            "-q",
            "--detach",
            b.to_str().unwrap(),
            "HEAD",
        ],
    );
    for wt in [&a, &b] {
        fs::create_dir(wt.join("target")).unwrap();
        fs::write(
            wt.join("target/CACHEDIR.TAG"),
            "Signature: 8a477f597d28d172789f06886806bc55\n",
        )
        .unwrap();
        blob(&wt.join("target/big.bin"));
    }
    fs::hard_link(a.join("target/big.bin"), a.join("target/big-link.bin")).unwrap();
    fs::write(a.join("ignored.bin"), vec![1u8; 262_144]).unwrap();
    fs::write(b.join("untracked.txt"), "dirty\n").unwrap();

    let rc = tmp.join("moai-rc-target-agent-2-t61-audit-2326");
    fs::create_dir_all(rc.join(".git")).unwrap();
    blob(&rc.join("x.bin"));
    let audit = tmp.join("moai-audit-2369-x");
    fs::create_dir_all(audit.join(".git")).unwrap();
    blob(&audit.join(".git/pack.bin"));
    fs::create_dir(tmp.join("moai-cli-home-abc")).unwrap();
    fs::create_dir(tmp.join("moai-foo")).unwrap();
    fs::write(tmp.join("moai-foo/f"), "f\n").unwrap();
    fs::create_dir(tmp.join("other-tool")).unwrap();
    blob(&tmp.join("other-tool/y.bin"));
    Fixture {
        _dir: dir,
        root,
        tmp,
        home,
    }
}

fn report(fx: &Fixture, budget_ms: &str) -> Value {
    report_with_tmp(fx, &fx.tmp, budget_ms)
}

fn report_with_tmp(fx: &Fixture, tmp: &Path, budget_ms: &str) -> Value {
    let out = Command::new(BIN)
        .args(["worktree-report"])
        .arg(&fx.root)
        .arg("--tmp-root")
        .arg(tmp)
        .args(["--budget-ms", budget_ms])
        .env("HOME", &fx.home)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .output()
        .unwrap();
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], true, "{v}");
    assert_eq!(v["operation"], "worktree.report");
    v["result"].clone()
}

fn by_path<'a>(r: &'a Value, p: &str) -> &'a Value {
    r["worktrees"]
        .as_array()
        .unwrap()
        .iter()
        .find(|w| w["path"] == p)
        .unwrap_or_else(|| panic!("no worktree {p}: {r}"))
}

#[test]
fn classes_and_worktree_facts_match_the_fixture() {
    let fx = fixture();
    let r = report(&fx, "600000");
    assert_eq!(r["partial"], false, "{r}");
    assert_eq!(r["writes"], false);
    assert_eq!(r["network"], "disabled");
    assert_eq!(r["candidates"]["status"], "not_evaluated");

    let c = &r["classes"];
    assert_eq!(c["worktree_checkout"]["count"], 3);
    assert_eq!(c["worktree_target"]["count"], 2);
    // The build-folder prefix wins over the audit pattern in its name.
    assert_eq!(c["tmp_rc_target"]["count"], 1);
    assert_eq!(c["tmp_audit_clone"]["count"], 1);
    assert_eq!(c["tmp_empty_dir"]["count"], 1);
    assert_eq!(
        c["tmp_other"]["count"], 1,
        "other-tool is not a moai-* folder"
    );
    assert_eq!(c["local_refs"]["branches"], 2);
    assert!(c["tmp_rc_target"]["bytes"].as_u64().unwrap() >= 65_536);

    let primary = by_path(&r, ".");
    let a = by_path(&r, ".moai-cli/worktrees/a");
    let b = by_path(&r, ".moai-cli/worktrees/b");
    assert_eq!(primary["has_target"], false);
    assert_eq!(primary["inside_project"], true);
    assert_eq!(a["branch"], "issue/1-a");
    assert_eq!(a["locked"], true);
    assert_eq!(a["lock_reason"], "moai claim t e1");
    assert_eq!(a["dirty"], false);
    assert_eq!(a["target_cachedir_tag"], true);
    assert_eq!(b["detached"], true);
    assert_eq!(b["branch"], Value::Null);
    assert_eq!(b["dirty"], true);

    // A hard link is counted once: a's target (file + link) equals b's (file only).
    assert_eq!(a["target_bytes"], b["target_bytes"], "{r}");
    assert!(a["target_bytes"].as_u64().unwrap() >= 65_536);
    // The primary checkout does not include the worktrees nested inside it.
    assert!(a["checkout_bytes"].as_u64().unwrap() >= 262_144, "{r}");
    assert!(primary["checkout_bytes"].as_u64().unwrap() < 262_144, "{r}");

    assert!(r["disk"]["free_bytes"].as_u64().unwrap() > 0, "{r}");
    assert!(r["disk"]["total_bytes"].as_u64().unwrap() > 0, "{r}");
}

#[test]
fn an_expired_budget_marks_every_sized_class_partial() {
    let fx = fixture();
    let r = report(&fx, "0");
    assert_eq!(r["partial"], true, "{r}");
    let partial: BTreeSet<&str> = r["partial_classes"]
        .as_array()
        .unwrap()
        .iter()
        .map(|v| v.as_str().unwrap())
        .collect();
    for cls in [
        "worktree_checkout",
        "worktree_target",
        "tmp_rc_target",
        "tmp_audit_clone",
        "tmp_other",
    ] {
        assert!(partial.contains(cls), "{cls}: {r}");
        assert_eq!(r["classes"][cls]["partial"], true, "{cls}");
        assert_eq!(r["classes"][cls]["bytes"], 0, "{cls}");
    }
    // Counts are still known; sizes and status checks are not.
    assert_eq!(r["classes"]["worktree_checkout"]["count"], 3);
    for w in r["worktrees"].as_array().unwrap() {
        assert_eq!(w["dirty"], Value::Null, "{w}");
    }
}

/// Every file under the fixture (repository with its `.git` and worktree indexes, and the
/// temporary root): path, length and modification time.
fn snapshot(dir: &Path) -> Vec<(String, u64, std::time::SystemTime)> {
    let mut out = Vec::new();
    let mut stack = vec![dir.to_path_buf()];
    while let Some(p) = stack.pop() {
        let m = fs::symlink_metadata(&p).unwrap();
        if m.is_dir() {
            for e in fs::read_dir(&p).unwrap() {
                stack.push(e.unwrap().path());
            }
        } else {
            out.push((
                p.to_string_lossy().into_owned(),
                m.len(),
                m.modified().unwrap(),
            ));
        }
    }
    out.sort();
    out
}

#[test]
fn the_report_writes_nothing() {
    let fx = fixture();
    let before = (snapshot(&fx.root), snapshot(&fx.tmp));
    report(&fx, "600000");
    let after = (snapshot(&fx.root), snapshot(&fx.tmp));
    assert_eq!(before, after);
}

fn partial_set(r: &Value) -> BTreeSet<String> {
    r["partial_classes"]
        .as_array()
        .unwrap()
        .iter()
        .map(|v| v.as_str().unwrap().to_string())
        .collect()
}

/// 4 MiB of bytes that do not compress (a fixed linear congruential sequence), so a git
/// object holding them stays about 4 MiB.
fn noise(path: &Path) {
    let mut x: u32 = 0x2380_2380;
    let bytes: Vec<u8> = (0..4 * 1024 * 1024)
        .map(|_| {
            x = x.wrapping_mul(1_664_525).wrapping_add(1_013_904_223);
            (x >> 24) as u8
        })
        .collect();
    fs::write(path, bytes).unwrap();
}

const FOUR_MIB: u64 = 4 * 1024 * 1024;

/// Pre-audit FAIL 6775068725 B1: symlinks are never followed — not a directory symlink to a
/// large outside folder, not a self loop, not a `moai-*` symlink in the temporary root.
#[test]
fn symlinks_are_never_followed() {
    let fx = fixture();
    let outside = fx.root.parent().unwrap().join("outside");
    fs::create_dir(&outside).unwrap();
    noise(&outside.join("big.dat"));
    let b = fx.root.join(".moai-cli/worktrees/b");
    std::os::unix::fs::symlink(&outside, b.join("outside-link")).unwrap();
    std::os::unix::fs::symlink(".", b.join("loop")).unwrap();
    std::os::unix::fs::symlink(&outside, fx.tmp.join("moai-link")).unwrap();

    let r = report(&fx, "600000");
    assert_eq!(r["partial"], false, "{r}");
    let b = by_path(&r, ".moai-cli/worktrees/b");
    assert!(b["checkout_bytes"].as_u64().unwrap() < FOUR_MIB, "{b}");
    // The symlink is not a temporary folder: the counts of the plain fixture stay.
    let c = &r["classes"];
    assert_eq!(c["tmp_other"]["count"], 1, "{c}");
    let tmp_total: u64 = [
        "tmp_rc_target",
        "tmp_audit_clone",
        "tmp_empty_dir",
        "tmp_other",
    ]
    .iter()
    .map(|k| c[*k]["bytes"].as_u64().unwrap())
    .sum();
    assert!(tmp_total < FOUR_MIB, "{c}");
}

/// Pre-audit FAIL 6775068725 B2: a folder that cannot be read marks its class partial, and
/// an unreadable temporary root marks every temporary class partial.
#[test]
fn unreadable_folders_mark_their_class_partial() {
    use std::os::unix::fs::PermissionsExt;
    let fx = fixture();
    let locked = fx.tmp.join("moai-locked");
    fs::create_dir(&locked).unwrap();
    fs::write(locked.join("f"), "f\n").unwrap();
    fs::set_permissions(&locked, fs::Permissions::from_mode(0o000)).unwrap();
    let readable = fs::read_dir(&locked).is_ok();
    let r = report(&fx, "600000");
    // Restore before asserting so the temporary directory can always be removed.
    fs::set_permissions(&locked, fs::Permissions::from_mode(0o755)).unwrap();
    if readable {
        eprintln!("skipped: this user can read a 000 folder (root)");
        return;
    }
    assert_eq!(r["partial"], true, "{r}");
    assert_eq!(partial_set(&r), BTreeSet::from(["tmp_other".to_string()]));
    assert_eq!(r["classes"]["tmp_other"]["partial"], true);

    let missing = fx.root.parent().unwrap().join("no-such-tmp");
    let r = report_with_tmp(&fx, &missing, "600000");
    assert_eq!(r["partial"], true, "{r}");
    let want: BTreeSet<String> = [
        "tmp_rc_target",
        "tmp_audit_clone",
        "tmp_empty_dir",
        "tmp_other",
    ]
    .iter()
    .map(|s| s.to_string())
    .collect();
    assert_eq!(partial_set(&r), want);
}

/// Pre-audit FAIL 6775068725 B3: a checkout's bytes leave out its git directory, measured
/// with a `.git` far larger than the checkout. Non-blocking N3: `refs/remotes/pr/*` counted.
#[test]
fn the_checkout_excludes_its_git_directory_and_pr_refs_are_counted() {
    let fx = fixture();
    noise(&fx.root.join("big.dat"));
    git(&fx.root, &["add", "-f", "big.dat"]);
    git(&fx.root, &["commit", "-q", "-m", "big object"]);
    git(&fx.root, &["rm", "-q", "big.dat"]);
    git(
        &fx.root,
        &["commit", "-q", "-m", "drop it from the checkout"],
    );
    git(&fx.root, &["update-ref", "refs/remotes/pr/1", "HEAD"]);
    let git_bytes: u64 = snapshot(&fx.root.join(".git")).iter().map(|e| e.1).sum();
    assert!(git_bytes >= FOUR_MIB, "fixture .git is {git_bytes} bytes");

    let r = report(&fx, "600000");
    let primary = by_path(&r, ".");
    assert!(
        primary["checkout_bytes"].as_u64().unwrap() < FOUR_MIB,
        "{primary}"
    );
    assert_eq!(r["classes"]["local_refs"]["remote_pr_refs"], 1);
}

fn keys(v: &Value) -> BTreeSet<String> {
    v.as_object().unwrap().keys().cloned().collect()
}

fn schema_keys(s: &Value) -> BTreeSet<String> {
    s["properties"]
        .as_object()
        .unwrap()
        .keys()
        .cloned()
        .collect()
}

fn required(s: &Value) -> BTreeSet<String> {
    s["required"]
        .as_array()
        .unwrap()
        .iter()
        .map(|v| v.as_str().unwrap().to_string())
        .collect()
}

#[test]
fn output_keys_equal_the_schema_both_ways() {
    let schema: Value = serde_json::from_str(
        &fs::read_to_string(
            Path::new(env!("CARGO_MANIFEST_DIR")).join("schemas/worktree-report.schema.json"),
        )
        .unwrap(),
    )
    .unwrap();
    let fx = fixture();
    let r = report(&fx, "600000");
    for (out, sch) in [
        (&r, &schema),
        (&r["disk"], &schema["properties"]["disk"]),
        (&r["classes"], &schema["properties"]["classes"]),
        (&r["classes"]["worktree_target"], &schema["$defs"]["class"]),
        (
            &r["classes"]["local_refs"],
            &schema["properties"]["classes"]["properties"]["local_refs"],
        ),
        (
            &r["worktrees"][0],
            &schema["properties"]["worktrees"]["items"],
        ),
        (&r["candidates"], &schema["properties"]["candidates"]),
    ] {
        assert_eq!(keys(out), schema_keys(sch), "{out}");
        assert_eq!(keys(out), required(sch), "{out}");
        assert_eq!(sch["additionalProperties"], false);
    }
}
