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
    fixture_in("")
}

/// The same fixture with every folder under `parent` (for example a name with a space).
fn fixture_in(parent: &str) -> Fixture {
    let dir = tempfile::TempDir::new().unwrap();
    let b = fs::canonicalize(dir.path()).unwrap().join(parent);
    fs::create_dir_all(&b).unwrap();
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

/// Runs the planned commands from `cwd`, with `pre` as git options before each subcommand.
fn run_commands(r: &Value, cwd: &Path, pre: &[&str]) -> Vec<String> {
    r["commands"]
        .as_array()
        .unwrap()
        .iter()
        .map(|argv| {
            let argv: Vec<&str> = argv
                .as_array()
                .unwrap()
                .iter()
                .map(|a| a.as_str().unwrap())
                .collect();
            assert_eq!(argv[0], "git");
            let mut all = pre.to_vec();
            all.extend_from_slice(&argv[1..]);
            git(cwd, &all)
        })
        .collect()
}

#[test]
fn a_relative_local_origin_is_resolved_against_the_project() {
    let fx = fixture();
    // `project` and `forge.git` are siblings, so `../forge.git` works only from the project.
    git(
        &fx.project,
        &["remote", "set-url", "origin", "../forge.git"],
    );
    let v = plan(&fx, &[]);
    assert_eq!(v["ok"], true, "{v}");
    let r = &v["result"];
    assert_eq!(r["repo_url"], fx.bare.to_str().unwrap());
    // Run from a folder two levels down, where `../forge.git` would name nothing.
    let deep = fx.tmp.join("a").join("b");
    fs::create_dir_all(&deep).unwrap();
    let out = run_commands(r, &deep, &[]);
    assert_eq!(out[3], fx.head);
    assert_eq!(out[4], fx.base);
}

#[test]
fn a_relative_origin_under_a_project_path_with_a_space_is_planned_and_runs() {
    // The written origin has no space; the resolved path inherits one from the parent.
    let fx = fixture_in("audit space");
    git(
        &fx.project,
        &["remote", "set-url", "origin", "../forge.git"],
    );
    let v = plan(&fx, &[]);
    assert_eq!(v["ok"], true, "{v}");
    let r = &v["result"];
    let url = r["repo_url"].as_str().unwrap();
    assert_eq!(url, fx.bare.to_str().unwrap());
    assert!(url.contains("audit space"), "{url}");
    let deep = fx.tmp.join("a").join("b");
    fs::create_dir_all(&deep).unwrap();
    let out = run_commands(r, &deep, &[]);
    assert_eq!(out[3], fx.head);
    assert_eq!(out[4], fx.base);
}

#[test]
fn a_relative_origin_that_itself_has_a_space_is_still_refused() {
    let fx = fixture();
    git(
        &fx.project,
        &["remote", "set-url", "origin", "../my forge.git"],
    );
    let v = plan(&fx, &[]);
    assert_eq!(v["ok"], false, "{v}");
    assert!(v.to_string().contains("not usable"), "{v}");
}

/// Raw origin boundary through the real CLI path (`git remote get-url` → plan). Every value
/// is stored with `git config` exactly as written; the plan must refuse it rather than plan a
/// trimmed or otherwise different URL. Literal values, one per boundary cell.
#[test]
fn raw_origin_values_with_surrounding_or_control_characters_are_refused() {
    for raw in [
        " ../forge.git",     // leading space
        "../forge.git ",     // trailing space (re-audit 6792085218 reproduction)
        "\t../forge.git",    // leading tab
        "../forge.git\t",    // trailing tab
        "../forge.git\r",    // trailing CR
        "../a\nb",           // newline inside the value
        "../forge\u{1}.git", // other control character
        "-x",                // leading dash
    ] {
        let fx = fixture();
        git(&fx.project, &["config", "remote.origin.url", raw]);
        let v = plan(&fx, &[]);
        assert_eq!(v["ok"], false, "{raw:?}: {v}");
        assert!(v.to_string().contains("not usable"), "{raw:?}: {v}");
    }
}

#[test]
fn a_raw_origin_that_is_not_utf8_is_refused() {
    let fx = fixture();
    git(&fx.project, &["config", "--unset-all", "remote.origin.url"]);
    let config = fx.project.join(".git").join("config");
    let mut bytes = fs::read(&config).unwrap();
    bytes.extend_from_slice(b"[remote \"origin\"]\n\turl = ../forge\xff.git\n");
    fs::write(&config, bytes).unwrap();
    let v = plan(&fx, &[]);
    assert_eq!(v["ok"], false, "{v}");
    assert!(v.to_string().contains("not UTF-8"), "{v}");
}

#[test]
fn an_empty_origin_is_planned_as_git_reads_it_the_remote_name() {
    // Git answers an empty remote URL with the remote's name, so the plan uses what git uses.
    let fx = fixture();
    git(&fx.project, &["config", "remote.origin.url", ""]);
    assert_eq!(git(&fx.project, &["remote", "get-url", "origin"]), "origin");
    let v = plan(&fx, &[]);
    assert_eq!(v["ok"], true, "{v}");
    let url = v["result"]["repo_url"].as_str().unwrap();
    assert_eq!(
        url,
        fx.project.join("origin").to_str().unwrap(),
        "resolved against the project top level"
    );
}

#[test]
fn the_clone_names_its_remote_origin_whatever_the_default_name() {
    let fx = fixture();
    let v = plan(&fx, &[]);
    let r = &v["result"];
    let out = run_commands(r, &fx.tmp, &["-c", "clone.defaultRemoteName=upstream"]);
    assert_eq!(out[3], fx.head);
    assert_eq!(
        out[4], fx.base,
        "the base fetch into refs/remotes/origin worked"
    );
    let dir = PathBuf::from(r["clone_dir"].as_str().unwrap());
    assert_eq!(git(&dir, &["remote"]), "origin");
}

#[test]
fn a_local_source_with_alternates_gives_an_independent_clone() {
    let fx = fixture();
    // `alt.git` borrows every object from `forge.git` through objects/info/alternates.
    let alt = fx.bare.with_file_name("alt.git");
    git(
        &fx.tmp,
        &[
            "clone",
            "-q",
            "--bare",
            "--shared",
            fx.bare.to_str().unwrap(),
            alt.to_str().unwrap(),
        ],
    );
    assert!(alt.join("objects/info/alternates").exists());
    let url_flag = format!("--repo-url={}", alt.to_str().unwrap());
    let v = plan(&fx, &[url_flag.as_str()]);
    assert_eq!(v["ok"], true, "{v}");
    let r = &v["result"];
    let out = run_commands(r, &fx.tmp, &[]);
    assert_eq!(out[3], fx.head);
    let dir = PathBuf::from(r["clone_dir"].as_str().unwrap());
    assert!(
        !dir.join(".git/objects/info/alternates").exists(),
        "the clone must not borrow the source's objects"
    );
    // With the borrowed-from store gone, the clone still reads its own objects.
    fs::rename(&fx.bare, fx.bare.with_file_name("forge.moved")).unwrap();
    assert_eq!(git(&dir, &["cat-file", "-t", &fx.head]), "commit");
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
