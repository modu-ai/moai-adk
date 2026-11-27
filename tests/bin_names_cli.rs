//! #2791 AC-1: `bin-names-plan` / `bin-names-apply` make `bin/moai` a hard link to
//! `bin/moai-cli` in the version folder the binary runs from. Every case runs copies of the
//! built binaries inside a temporary HOME; the real HOME, PATH and shell files are never read.

#![cfg(unix)]

mod common;

use common::pty::output_within;
use common::CleanEnv;
use serde_json::{json, Value};
use std::os::unix::fs::MetadataExt as _;
use std::os::unix::fs::PermissionsExt as _;
use std::{
    fs,
    path::{Path, PathBuf},
    process::{Command, Stdio},
};
use tempfile::TempDir;

const EXE: &str = common::MOAI_CLI_BIN;
const MOAI: &str = common::MOAI_BIN;
const NOW: &str = "2026-11-13T21:00:00Z";

/// A temporary HOME with `~/.moai-cli/versions/a/bin/moai-cli`; returns (guard, home, exe).
fn layout() -> (TempDir, PathBuf, PathBuf) {
    let d = TempDir::new().unwrap();
    let home = fs::canonicalize(d.path()).unwrap();
    let bin = home.join(".moai-cli/versions/a/bin");
    fs::create_dir_all(&bin).unwrap();
    let exe = bin.join("moai-cli");
    common::copy_executable(Path::new(EXE), &exe);
    (d, home, exe)
}

fn call(exe: &Path, home: &Path, args: &[&str]) -> Value {
    let out = Command::new(exe)
        .args(args)
        .env("HOME", home)
        .env("PATH", "/usr/bin:/bin")
        .current_dir(home)
        .stdin(Stdio::null())
        .output()
        .unwrap();
    serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|_| panic!("{}", String::from_utf8_lossy(&out.stderr)))
}

fn plan(exe: &Path, home: &Path) -> Value {
    let v = call(
        exe,
        home,
        &["bin-names-plan", "--home", home.to_str().unwrap()],
    );
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn apply(exe: &Path, home: &Path, digest: &str, perm: &str) -> Value {
    call(
        exe,
        home,
        &[
            "bin-names-apply",
            "--home",
            home.to_str().unwrap(),
            "--plan-digest",
            digest,
            "--permission-reference",
            perm,
            "--now",
            NOW,
        ],
    )
}

fn names(dir: &Path) -> Vec<(String, Vec<u8>)> {
    let mut v: Vec<_> = fs::read_dir(dir)
        .unwrap()
        .flatten()
        .map(|e| {
            let p = e.path();
            let bytes = fs::read(&p).unwrap_or_default();
            (e.file_name().to_string_lossy().into_owned(), bytes)
        })
        .collect();
    v.sort();
    v
}

fn same_file(a: &Path, b: &Path) -> bool {
    let (m1, m2) = (fs::metadata(a).unwrap(), fs::metadata(b).unwrap());
    (m1.dev(), m1.ino()) == (m2.dev(), m2.ino())
}

fn version_line(exe: &Path) -> String {
    let out = Command::new(exe).arg("--version").output().unwrap();
    String::from_utf8_lossy(&out.stdout)
        .lines()
        .next()
        .unwrap_or("")
        .to_string()
}

/// The two reasons the plan gives today for a name it will not touch (REQ-3 and REQ-4 keep them).
const CLI_REASON: &str = "bin/moai-cli is missing or not a regular file";
const MOAI_REASON: &str = "bin/moai is a link or a folder";

/// Every entry of `dir` sorted by name, judged by its own metadata and never through a link:
/// the kind, the raw link text of a symlink, and `(dev, ino)` with the link count and length
/// of a file. `names` reads through links, so it cannot tell these cases apart.
fn snapshot(dir: &Path) -> Vec<(String, String)> {
    let mut v: Vec<_> = fs::read_dir(dir)
        .unwrap()
        .flatten()
        .map(|e| {
            let p = e.path();
            let m = fs::symlink_metadata(&p).unwrap();
            let kind = m.file_type();
            let what = if kind.is_symlink() {
                format!("symlink -> {}", fs::read_link(&p).unwrap().display())
            } else if kind.is_dir() {
                "dir".to_string()
            } else {
                format!(
                    "file dev={} ino={} nlink={} len={}",
                    m.dev(),
                    m.ino(),
                    m.nlink(),
                    m.len()
                )
            };
            (e.file_name().to_string_lossy().into_owned(), what)
        })
        .collect();
    v.sort();
    v
}

/// How the second name of a pair is made from the first.
#[derive(Debug, Clone, Copy)]
enum Link {
    Hard,
    Relative,
    Absolute,
}

fn link_to(target: &Path, link: &Path, how: Link) {
    match how {
        Link::Hard => fs::hard_link(target, link),
        Link::Relative => std::os::unix::fs::symlink(target.file_name().unwrap(), link),
        Link::Absolute => std::os::unix::fs::symlink(target, link),
    }
    .unwrap();
}

/// In `bin`, makes `regular` a regular copy of `binary` and `linked` a link to it, replacing
/// both names. Returns `(regular, linked)`.
fn pair(bin: &Path, binary: &str, regular: &str, linked: &str, how: Link) -> (PathBuf, PathBuf) {
    let (r, l) = (bin.join(regular), bin.join(linked));
    for p in [&r, &l] {
        let _ = fs::remove_file(p);
    }
    common::copy_executable(Path::new(binary), &r);
    link_to(&r, &l, how);
    (r, l)
}

fn bin_dir(exe: &Path) -> PathBuf {
    exe.parent().unwrap().to_path_buf()
}

/// A regular copy of the binary that is no part of the pair under test: the plan and the apply
/// of a negative case run from it, so they never run through the link being judged.
fn probe_beside(exe: &Path) -> PathBuf {
    let probe = exe.with_file_name("probe");
    common::copy_executable(Path::new(EXE), &probe);
    probe
}

/// The `bin/` of another version folder of the same temporary HOME (`versions/b/bin`).
fn other_version_bin(exe: &Path) -> PathBuf {
    let versions = exe.parent().unwrap().parent().unwrap().parent().unwrap();
    let dir = versions.join("b/bin");
    fs::create_dir_all(&dir).unwrap();
    dir
}

/// The plan says `unchanged` and the apply writes nothing: the names, their link texts and
/// their inodes are as they were.
fn assert_unchanged_untouched(runner: &Path, home: &Path, bin: &Path, case: &str) {
    let before = snapshot(bin);
    let p = plan(runner, home);
    assert_eq!(p["state"], json!("unchanged"), "{case}: {p}");
    let a = apply(runner, home, p["plan_digest"].as_str().unwrap(), "approved");
    assert_eq!(a["ok"], json!(true), "{case}: {a}");
    assert_eq!(a["result"]["written"], json!([]), "{case}: {a}");
    assert_eq!(snapshot(bin), before, "{case}");
}

/// The plan says `blocked` with `reason` and the apply writes nothing.
fn assert_blocked_untouched(runner: &Path, home: &Path, bin: &Path, reason: &str, case: &str) {
    let before = snapshot(bin);
    let p = plan(runner, home);
    assert_eq!(p["state"], json!("blocked"), "{case}: {p}");
    assert!(
        p["reason"].as_str().unwrap().contains(reason),
        "{case}: {p}"
    );
    let a = apply(runner, home, p["plan_digest"].as_str().unwrap(), "approved");
    assert_eq!(a["ok"], json!(true), "{case}: {a}");
    assert_eq!(a["result"]["written"], json!([]), "{case}: {a}");
    assert_eq!(snapshot(bin), before, "{case}");
}

/// The package a version folder carries beside `bin/` for a full `update` run.
const PACKAGE: [(&str, &str); 3] = [
    (
        "SKILL.md",
        "---\nname: moai\ndescription: fixture package\n---\n# MoAI\nbody\n",
    ),
    ("workflows/update.md", "# UPDATE\nv1\n"),
    ("references/routes.json", "{\"routes\":1}\n"),
];

/// An agent source folder and the package under `folder`, as `install_version` of
/// `tests/update_command_cli.rs` lays them out.
fn put_package(folder: &Path) {
    fs::create_dir_all(folder.join("agents")).unwrap();
    for (rel, body) in PACKAGE {
        let p = folder.join("skills/moai").join(rel);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        fs::write(p, body).unwrap();
    }
}

/// One `moai update` run with a cleared environment: an inherited Claude folder variable could
/// write into the real Claude folder. Returns the whole envelope.
fn update(runner: &Path, home: &Path, args: &[&str]) -> Value {
    let mut cmd = Command::new(runner);
    cmd.arg("update")
        .args(args)
        .clean_env()
        .env("PATH", "/usr/bin:/bin")
        .env("HOME", home)
        .env("LC_ALL", "C")
        .current_dir(home);
    let out = output_within(&mut cmd, None);
    serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
        panic!(
            "envelope JSON on stdout: {}{}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    })
}

#[test]
fn create_links_the_alias_and_a_rerun_is_unchanged() {
    let (_d, home, exe) = layout();
    let bin = exe.parent().unwrap().to_path_buf();
    let alias = bin.join("moai");
    let p = plan(&exe, &home);
    assert_eq!(p["state"], json!("create"), "{p}");
    let d = p["plan_digest"].as_str().unwrap().to_string();

    // Refusals write nothing.
    let before = names(&bin);
    for (digest, perm, why) in [
        (d.as_str(), "", "permission reference"),
        ("sha256:00", "approved", "stale or altered plan"),
    ] {
        let a = apply(&exe, &home, digest, perm);
        assert_eq!(a["ok"], json!(false), "{why}: {a}");
        assert!(a.to_string().contains(why), "{why}: {a}");
        assert_eq!(names(&bin), before, "{why}");
    }

    let a = apply(&exe, &home, &d, "approved");
    assert_eq!(a["ok"], json!(true), "{a}");
    assert_eq!(a["result"]["state"], json!("applied"), "{a}");
    assert!(same_file(&exe, &alias), "one file, two names");
    assert_eq!(version_line(&alias), version_line(&exe));
    assert!(version_line(&alias).ends_with(" (rust)"));

    // Planned from the `moai` name the same folder is found, and nothing is left to do.
    let again = plan(&alias, &home);
    assert_eq!(again["state"], json!("unchanged"), "{again}");
    let a = apply(
        &alias,
        &home,
        again["plan_digest"].as_str().unwrap(),
        "approved",
    );
    assert_eq!(a["result"]["written"], json!([]), "{a}");
}

#[test]
fn a_separate_copy_of_this_build_is_moved_aside_never_deleted() {
    let (_d, home, exe) = layout();
    let bin = exe.parent().unwrap().to_path_buf();
    let alias = bin.join("moai");
    common::copy_executable(Path::new(MOAI), &alias);
    let copy_bytes = fs::read(&alias).unwrap();
    assert_ne!(
        copy_bytes,
        fs::read(&exe).unwrap(),
        "two binaries of one build"
    );
    let p = plan(&exe, &home);
    assert_eq!(p["state"], json!("replace_copy"), "{p}");

    // The aside name is taken: refused before anything moves.
    let aside = bin.join("moai.replaced-20261113T210000Z");
    fs::write(&aside, "taken").unwrap();
    let before = names(&bin);
    let a = apply(&exe, &home, p["plan_digest"].as_str().unwrap(), "approved");
    assert_eq!(a["ok"], json!(false), "{a}");
    assert!(a.to_string().contains("already exists"), "{a}");
    assert_eq!(names(&bin), before);
    fs::remove_file(&aside).unwrap();

    let a = apply(&exe, &home, p["plan_digest"].as_str().unwrap(), "approved");
    assert_eq!(a["ok"], json!(true), "{a}");
    assert_eq!(fs::read(&aside).unwrap(), copy_bytes, "the copy is kept");
    assert!(same_file(&exe, &alias));
}

#[test]
fn another_program_a_link_or_a_folder_is_left_alone() {
    let (_d, home, exe) = layout();
    let bin = exe.parent().unwrap().to_path_buf();
    let alias = bin.join("moai");
    // The Go `moai` under this name: reported, never replaced.
    common::write_executable(&alias, "#!/bin/sh\necho 'moai-adk 3.2.0'\n");
    let p = plan(&exe, &home);
    assert_eq!(p["state"], json!("blocked"), "{p}");
    assert!(
        p["reason"].as_str().unwrap().contains("moai-adk 3.2.0"),
        "{p}"
    );
    let before = names(&bin);
    let a = apply(&exe, &home, p["plan_digest"].as_str().unwrap(), "approved");
    assert_eq!(a["result"]["written"], json!([]), "{a}");
    assert_eq!(names(&bin), before);

    // A link to a regular file of another version folder is not this folder's own `moai`
    // (a link to the sibling `moai-cli` is, SPEC-20261127-launcher-bin-layout D-2).
    fs::remove_file(&alias).unwrap();
    let elsewhere = other_version_bin(&exe).join("moai");
    common::copy_executable(Path::new(MOAI), &elsewhere);
    std::os::unix::fs::symlink(&elsewhere, &alias).unwrap();
    let p = plan(&exe, &home);
    assert_eq!(p["state"], json!("blocked"), "{p}");
    assert!(p["reason"].as_str().unwrap().contains(MOAI_REASON), "{p}");
    fs::remove_file(&alias).unwrap();
    fs::create_dir(&alias).unwrap();
    assert_eq!(plan(&exe, &home)["state"], json!("blocked"));
}

#[test]
fn outside_the_version_folder_nothing_is_linked() {
    let d = TempDir::new().unwrap();
    let root = fs::canonicalize(d.path()).unwrap();
    let home = root.join("home");
    fs::create_dir_all(&home).unwrap();
    // `bin/` but not under `versions/<v>/`, and a `versions/<v>/` folder not named `bin`.
    for rel in ["elsewhere/bin", "home/.moai-cli/versions/a/other"] {
        let dir = root.join(rel);
        fs::create_dir_all(&dir).unwrap();
        let exe = dir.join("moai-cli");
        common::copy_executable(Path::new(EXE), &exe);
        let p = plan(&exe, &home);
        assert_eq!(p["state"], json!("outside_layout"), "{rel}: {p}");
        let a = apply(&exe, &home, p["plan_digest"].as_str().unwrap(), "approved");
        assert_eq!(a["result"]["written"], json!([]), "{rel}: {a}");
        assert!(!dir.join("moai").exists(), "{rel}");
    }
}

// SPEC-20261127-launcher-bin-layout AC-1 (REQ-1, REQ-5): `moai` is the regular file and
// `moai-cli` a link to it, relative and absolute, run from both names.
#[test]
fn ac1_a_cli_link_to_the_sibling_moai_is_unchanged() {
    for how in [Link::Relative, Link::Absolute] {
        let (_d, home, exe) = layout();
        let bin = bin_dir(&exe);
        let (moai, cli) = pair(&bin, MOAI, "moai", "moai-cli", how);
        for runner in [&moai, &cli] {
            let case = format!(
                "{how:?} via {}",
                runner.file_name().unwrap().to_string_lossy()
            );
            assert_unchanged_untouched(runner, &home, &bin, &case);
        }
    }
}

// SPEC-20261127-launcher-bin-layout AC-2 (REQ-2, REQ-5): `moai-cli` stays the regular file and
// `moai` is a link to it, relative and absolute, run from both names.
#[test]
fn ac2_a_moai_link_to_the_sibling_moai_cli_is_unchanged() {
    for how in [Link::Relative, Link::Absolute] {
        let (_d, home, exe) = layout();
        let bin = bin_dir(&exe);
        let moai = bin.join("moai");
        link_to(&exe, &moai, how);
        for runner in [&exe, &moai] {
            let case = format!(
                "{how:?} via {}",
                runner.file_name().unwrap().to_string_lossy()
            );
            assert_unchanged_untouched(runner, &home, &bin, &case);
        }
    }
}

// SPEC-20261127-launcher-bin-layout AC-3 (REQ-3): a `moai-cli` link that does not lead to the
// sibling regular `moai` stays blocked. Run from a regular copy, never through the link.
#[test]
fn ac3_a_cli_link_that_leaves_the_pair_stays_blocked() {
    let cases = [
        "dangling",
        "folder",
        "sibling moai is a folder",
        "file of another version folder",
        "file outside the state folder",
    ];
    for case in cases {
        let (_d, home, exe) = layout();
        let bin = bin_dir(&exe);
        let probe = probe_beside(&exe);
        let outside = TempDir::new().unwrap();
        let target = match case {
            "dangling" => PathBuf::from("nowhere"),
            "folder" => {
                fs::create_dir(bin.join("adir")).unwrap();
                PathBuf::from("adir")
            }
            "sibling moai is a folder" => {
                fs::create_dir(bin.join("moai")).unwrap();
                PathBuf::from("moai")
            }
            "file of another version folder" => {
                let f = other_version_bin(&exe).join("moai");
                common::copy_executable(Path::new(MOAI), &f);
                f
            }
            _ => {
                let f = outside.path().join("moai");
                common::copy_executable(Path::new(MOAI), &f);
                f
            }
        };
        fs::remove_file(&exe).unwrap();
        std::os::unix::fs::symlink(&target, &exe).unwrap();
        assert_blocked_untouched(&probe, &home, &bin, CLI_REASON, case);
    }
}

// SPEC-20261127-launcher-bin-layout AC-4 (REQ-3, REQ-6): a link to another file with the very
// same bytes as the sibling `moai` is no link to the pair.
#[test]
fn ac4_a_cli_link_to_the_same_bytes_elsewhere_stays_blocked() {
    for how in [Link::Relative, Link::Absolute] {
        let (_d, home, exe) = layout();
        let bin = bin_dir(&exe);
        let probe = probe_beside(&exe);
        let moai = bin.join("moai");
        let other = bin.join("other");
        common::copy_executable(Path::new(MOAI), &moai);
        common::copy_executable(Path::new(MOAI), &other);
        assert_eq!(fs::read(&moai).unwrap(), fs::read(&other).unwrap(), "setup");
        fs::remove_file(&exe).unwrap();
        link_to(&other, &exe, how);
        assert_blocked_untouched(&probe, &home, &bin, CLI_REASON, &format!("{how:?}"));
    }
}

// SPEC-20261127-launcher-bin-layout AC-5 (REQ-4, REQ-6): a `moai` link that does not lead to
// the sibling regular `moai-cli` stays blocked. Run from a regular copy.
#[test]
fn ac5_a_moai_link_that_leaves_the_pair_stays_blocked() {
    let cases = [
        "dangling",
        "folder",
        "same bytes elsewhere",
        "file of another version folder",
    ];
    for case in cases {
        let (_d, home, exe) = layout();
        let bin = bin_dir(&exe);
        let probe = probe_beside(&exe);
        let target = match case {
            "dangling" => PathBuf::from("nowhere"),
            "folder" => {
                fs::create_dir(bin.join("adir")).unwrap();
                PathBuf::from("adir")
            }
            "same bytes elsewhere" => {
                common::copy_executable(Path::new(EXE), &bin.join("other"));
                assert_eq!(
                    fs::read(&exe).unwrap(),
                    fs::read(bin.join("other")).unwrap(),
                    "setup"
                );
                PathBuf::from("other")
            }
            _ => {
                let f = other_version_bin(&exe).join("moai-cli");
                common::copy_executable(Path::new(EXE), &f);
                f
            }
        };
        std::os::unix::fs::symlink(&target, bin.join("moai")).unwrap();
        assert_blocked_untouched(&probe, &home, &bin, MOAI_REASON, case);
    }
}

// SPEC-20261127-launcher-bin-layout AC-6 (REQ-5, REQ-7): a full `update` run from a version
// folder in each layout reports `unchanged` and leaves both names as they were. The hard link
// layout passes today; the other four are the red part.
#[test]
fn ac6_update_does_not_block_itself_on_a_linked_layout() {
    let shapes = [
        ("hard link", "moai-cli", "moai", Link::Hard),
        ("reversed relative", "moai", "moai-cli", Link::Relative),
        ("reversed absolute", "moai", "moai-cli", Link::Absolute),
        ("mirror relative", "moai-cli", "moai", Link::Relative),
        ("mirror absolute", "moai-cli", "moai", Link::Absolute),
    ];
    let mut problems: Vec<String> = Vec::new();
    for (label, regular, linked, how) in shapes {
        let (_d, home, first) = layout();
        let bin = bin_dir(&first);
        put_package(bin.parent().unwrap());
        let (runner, _) = pair(&bin, MOAI, regular, linked, how);
        let before = snapshot(&bin);

        let v = update(&runner, &home, &["--json"]);
        let r = &v["result"];
        if v["ok"] != json!(true) || r["mode"] != json!("preview") {
            problems.push(format!("{label}: the preview answered {v}"));
            continue;
        }
        if r["bin_names"]["state"] != json!("unchanged") {
            problems.push(format!(
                "{label}: the preview says bin_names {}",
                r["bin_names"]
            ));
            continue;
        }
        if snapshot(&bin) != before {
            problems.push(format!("{label}: the preview changed the names"));
        }

        let digest = r["plan_digest"].as_str().unwrap().to_string();
        let v = update(&runner, &home, &["--confirm", &digest]);
        let r = &v["result"];
        if v["ok"] != json!(true) || r["mode"] != json!("applied") {
            problems.push(format!("{label}: the confirm answered {v}"));
            continue;
        }
        if r["bin_names_result"]["state"] != json!("unchanged") {
            problems.push(format!(
                "{label}: the confirm says bin_names_result {}",
                r["bin_names_result"]
            ));
        }
        if snapshot(&bin) != before {
            problems.push(format!("{label}: the confirm changed the names"));
        }
    }
    assert!(problems.is_empty(), "{problems:#?}");
}

/// A version folder that holds `bin/moai` alone: a regular copy of the `moai` binary and no
/// `bin/moai-cli`, the shape of a new install (SPEC-20261127-installed-binary-name-moai).
/// Returns `(guard, home, moai)`.
fn only_moai() -> (TempDir, PathBuf, PathBuf) {
    let (d, home, cli) = layout();
    let bin = bin_dir(&cli);
    fs::remove_file(&cli).unwrap();
    let moai = bin.join("moai");
    common::copy_executable(Path::new(MOAI), &moai);
    (d, home, moai)
}

/// Whether `path` names anything at all, judged by its own metadata and never through a link.
fn exists_by_lstat(path: &Path) -> bool {
    fs::symlink_metadata(path).is_ok()
}

// SPEC-20261127-installed-binary-name-moai AC-5 (REQ-5): a version folder with `bin/moai`
// alone is the installed shape. The plan says `unchanged`, the apply writes nothing, and a full
// `update` run reports `unchanged` before and after the confirm and never makes a `moai-cli`.
#[test]
fn ac5_only_moai_is_unchanged_by_the_plan_and_by_update() {
    let (_d, home, moai) = only_moai();
    let bin = bin_dir(&moai);
    put_package(bin.parent().unwrap());
    assert_unchanged_untouched(&moai, &home, &bin, "only moai");
    assert!(
        !exists_by_lstat(&bin.join("moai-cli")),
        "the plan or the apply made a moai-cli"
    );

    let before = snapshot(&bin);
    let v = update(&moai, &home, &["--json"]);
    let r = &v["result"];
    assert_eq!(v["ok"], json!(true), "{v}");
    assert_eq!(r["mode"], json!("preview"), "{v}");
    assert_eq!(
        r["bin_names"]["state"],
        json!("unchanged"),
        "{}",
        r["bin_names"]
    );
    assert_eq!(snapshot(&bin), before, "the preview changed the names");

    let digest = r["plan_digest"].as_str().unwrap().to_string();
    let v = update(&moai, &home, &["--confirm", &digest]);
    let r = &v["result"];
    assert_eq!(v["ok"], json!(true), "{v}");
    assert_eq!(r["mode"], json!("applied"), "{v}");
    assert_eq!(
        r["bin_names_result"]["state"],
        json!("unchanged"),
        "{}",
        r["bin_names_result"]
    );
    assert_eq!(snapshot(&bin), before, "the confirm changed the names");
    assert!(!exists_by_lstat(&bin.join("moai-cli")));
}

// SPEC-20261127-installed-binary-name-moai AC-6 (REQ-6): the earlier shapes keep their states
// and no state makes or refreshes a `moai-cli`. A `moai-cli` that was there keeps its inode, one
// that was not there stays absent. A `moai` that is another program beside `moai-cli` stays
// blocked and names the program. This one passes before the change: it pins the earlier shapes.
#[test]
fn ac6_no_name_state_makes_or_refreshes_a_moai_cli() {
    // The folder holds `moai-cli` alone: `create` links `moai` and leaves `moai-cli` as it is.
    {
        let (_d, home, cli) = layout();
        let bin = bin_dir(&cli);
        let before = snapshot(&bin);
        let p = plan(&cli, &home);
        assert_eq!(p["state"], json!("create"), "{p}");
        let a = apply(&cli, &home, p["plan_digest"].as_str().unwrap(), "approved");
        assert_eq!(a["ok"], json!(true), "{a}");
        let after = snapshot(&bin);
        let cli_before = before.iter().find(|(n, _)| n == "moai-cli").unwrap();
        let cli_after = after.iter().find(|(n, _)| n == "moai-cli").unwrap();
        assert!(cli_after.1.starts_with("file dev="), "{cli_after:?}");
        assert_eq!(
            cli_before.1.split(" nlink=").next(),
            cli_after.1.split(" nlink=").next(),
            "moai-cli was rewritten"
        );
    }
    // The hard link layout stays `unchanged` and gains nothing.
    {
        let (_d, home, cli) = layout();
        let bin = bin_dir(&cli);
        fs::hard_link(&cli, bin.join("moai")).unwrap();
        assert_unchanged_untouched(&cli, &home, &bin, "hard link");
    }
    // A `moai` that is another program beside `moai-cli`: blocked, naming the program, and both
    // names keep their inodes.
    {
        let (_d, home, cli) = layout();
        let bin = bin_dir(&cli);
        let moai = bin.join("moai");
        fs::write(&moai, "#!/bin/sh\necho 'other 9.9.9 (go)'\n").unwrap();
        fs::set_permissions(&moai, fs::Permissions::from_mode(0o755)).unwrap();
        let probe = probe_beside(&cli);
        assert_blocked_untouched(&probe, &home, &bin, "another program", "other program");
    }
}

// SPEC-20261127-installed-binary-name-moai AC-8 (REQ-8): a `bin/moai-cli` and a link
// `moai-cli` in the link folder of the state folder are listed as `left_in_place` by the plan
// and by the update preview, and neither command changes them.
#[test]
fn ac8_a_name_left_in_place_is_listed_and_never_touched() {
    // The hard link layout lists its `bin/moai-cli`.
    {
        let (_d, home, cli) = layout();
        let bin = bin_dir(&cli);
        fs::hard_link(&cli, bin.join("moai")).unwrap();
        let p = plan(&cli, &home);
        assert_eq!(
            p["left_in_place"],
            json!([bin.join("moai-cli").to_string_lossy()]),
            "{p}"
        );
        assert_unchanged_untouched(&cli, &home, &bin, "hard link");
    }
    // `bin/moai` alone with a link `moai-cli` in the link folder lists that link.
    {
        let (_d, home, moai) = only_moai();
        let bin = bin_dir(&moai);
        let state = bin.parent().unwrap().parent().unwrap().parent().unwrap();
        let links = state.join("path");
        fs::create_dir_all(&links).unwrap();
        let link = links.join("moai-cli");
        std::os::unix::fs::symlink("../versions/a/bin/moai-cli", &link).unwrap();
        let before = snapshot(&links);
        let p = plan(&moai, &home);
        assert_eq!(p["state"], json!("unchanged"), "{p}");
        assert_eq!(p["left_in_place"], json!([link.to_string_lossy()]), "{p}");
        assert_unchanged_untouched(&moai, &home, &bin, "only moai with a link");
        assert_eq!(snapshot(&links), before, "the link folder changed");

        put_package(bin.parent().unwrap());
        let v = update(&moai, &home, &["--json"]);
        assert_eq!(
            v["result"]["bin_names"]["left_in_place"],
            json!([link.to_string_lossy()]),
            "{v}"
        );
        assert_eq!(
            snapshot(&links),
            before,
            "the preview changed the link folder"
        );
    }
    // `bin/moai` alone and nothing else lists nothing.
    {
        let (_d, home, moai) = only_moai();
        let p = plan(&moai, &home);
        assert_eq!(p["left_in_place"], json!([]), "{p}");
    }
}
