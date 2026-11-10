//! WT-3 (#2380): `moai-cli audit-clone-plan`. The plan is read-only; this test then runs the
//! planned commands against a local bare repository (file path URL, no network) and checks
//! the clone holds only the head and base branches, sits detached at the exact head, and
//! reports the merge base.
#![cfg(unix)]
mod common;

use serde_json::Value;
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

struct Fixture {
    _dir: tempfile::TempDir,
    project: PathBuf,
    bare: PathBuf,
    tmp: PathBuf,
    home: PathBuf,
    head: String,
    base: String,
}

/// A bare "forge" with `main` (base), `issue/1-x` (head, one commit on main) and an unrelated
/// `other` branch; the project is a clone of it.
fn fixture() -> Fixture {
    let dir = tempfile::TempDir::new().unwrap();
    let b = fs::canonicalize(dir.path()).unwrap();
    let (work, bare, project, tmp, home) = (
        b.join("work"),
        b.join("forge.git"),
        b.join("project"),
        b.join("tmp"),
        b.join("home"),
    );
    for d in [&work, &tmp, &home] {
        fs::create_dir_all(d).unwrap();
    }
    git(&work, &["init", "-q", "-b", "main"]);
    fs::write(work.join("a.txt"), "a\n").unwrap();
    git(&work, &["add", "a.txt"]);
    git(&work, &["commit", "-q", "-m", "base"]);
    let base = git(&work, &["rev-parse", "HEAD"]);
    git(&work, &["switch", "-q", "-c", "other"]);
    fs::write(work.join("o.txt"), "o\n").unwrap();
    git(&work, &["add", "o.txt"]);
    git(&work, &["commit", "-q", "-m", "other"]);
    git(&work, &["switch", "-q", "-c", "issue/1-x", "main"]);
    fs::write(work.join("x.txt"), "x\n").unwrap();
    git(&work, &["add", "x.txt"]);
    git(&work, &["commit", "-q", "-m", "head"]);
    let head = git(&work, &["rev-parse", "HEAD"]);
    git(
        &b,
        &[
            "clone",
            "-q",
            "--bare",
            work.to_str().unwrap(),
            bare.to_str().unwrap(),
        ],
    );
    git(
        &b,
        &[
            "clone",
            "-q",
            bare.to_str().unwrap(),
            project.to_str().unwrap(),
        ],
    );
    Fixture {
        _dir: dir,
        project,
        bare,
        tmp,
        home,
        head,
        base,
    }
}

fn plan(fx: &Fixture, extra: &[&str]) -> Value {
    let out = Command::new(BIN)
        .arg("audit-clone-plan")
        .arg(&fx.project)
        .args([
            "--pr",
            "2380",
            "--head",
            &fx.head,
            "--head-branch",
            "issue/1-x",
            "--base-branch",
            "main",
        ])
        .arg("--tmp-root")
        .arg(&fx.tmp)
        .args(extra)
        .env("HOME", &fx.home)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .output()
        .unwrap();
    serde_json::from_slice(&out.stdout).unwrap()
}

#[test]
fn the_planned_commands_make_a_minimal_detached_clone() {
    let fx = fixture();
    let v = plan(&fx, &[]);
    assert_eq!(v["ok"], true, "{v}");
    assert_eq!(v["operation"], "audit.clone.plan");
    let r = &v["result"];
    assert_eq!(r["writes"], false);
    // Planning wrote nothing: the clone folder does not exist yet.
    let dir = PathBuf::from(r["clone_dir"].as_str().unwrap());
    assert_eq!(
        dir,
        fx.tmp.join(format!("moai-audit-2380-{}", &fx.head[..8]))
    );
    assert!(!dir.exists());
    // The default URL is the project's origin.
    assert_eq!(r["repo_url"], fx.bare.to_str().unwrap());
    // Without a project profile there is no project id, so no evidence folder.
    assert_eq!(r["evidence_dir"], Value::Null);
    assert!(r["evidence_reason"]
        .as_str()
        .unwrap()
        .contains("project_id"));

    let mut outputs = Vec::new();
    for argv in r["commands"].as_array().unwrap() {
        let argv: Vec<&str> = argv
            .as_array()
            .unwrap()
            .iter()
            .map(|a| a.as_str().unwrap())
            .collect();
        assert_eq!(argv[0], "git");
        outputs.push(git(&fx.tmp, &argv[1..]));
    }
    assert_eq!(outputs[3], fx.head, "command 4 is rev-parse HEAD");
    assert_eq!(outputs[4], fx.base, "command 5 is the merge base");
    assert_eq!(
        git(&dir, &["rev-parse", "--abbrev-ref", "HEAD"]),
        "HEAD",
        "detached"
    );
    // Only the head branch and the base branch were fetched; `other` was not.
    let refs = git(&dir, &["for-each-ref", "--format=%(refname)"]);
    assert!(refs.contains("refs/remotes/origin/main"), "{refs}");
    assert!(refs.contains("refs/remotes/origin/issue/1-x"), "{refs}");
    assert!(!refs.contains("other"), "{refs}");
}

#[test]
fn bad_inputs_are_refused_before_anything_is_planned() {
    let fx = fixture();
    for (args, needle) in [
        (vec!["--pr", "0"], "pull request number"),
        (vec!["--head", "abc"], "40-hex"),
        (vec!["--head-branch", "bad..name"], "--head-branch"),
        (vec!["--base-branch", "-x"], "--base-branch"),
        (vec!["--repo-url", "has space"], "URL"),
    ] {
        // Good values for every flag, with the one under test replaced (`--flag=value` so a
        // value starting with `-` stays a value).
        let mut flags = vec![
            ("--pr", "2380".to_string()),
            ("--head", fx.head.clone()),
            ("--head-branch", "issue/1-x".to_string()),
            ("--base-branch", "main".to_string()),
        ];
        match flags.iter_mut().find(|(f, _)| *f == args[0]) {
            Some(slot) => slot.1 = args[1].to_string(),
            None => flags.push((args[0], args[1].to_string())),
        }
        let out = Command::new(BIN)
            .arg("audit-clone-plan")
            .arg(&fx.project)
            .args(flags.iter().map(|(f, v)| format!("{f}={v}")))
            .env("HOME", &fx.home)
            .output()
            .unwrap();
        let v: Value = serde_json::from_slice(&out.stdout).unwrap_or(Value::Null);
        let text = format!("{v}{}", String::from_utf8_lossy(&out.stderr));
        assert!(!out.status.success(), "{args:?}: {text}");
        assert!(text.contains(needle), "{args:?}: {text}");
    }
}
