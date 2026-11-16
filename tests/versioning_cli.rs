//! Card C-69 through the binary: `project-detect` reports the versioning scheme a project
//! already uses from its tag names and root manifests, read only; and the project profile
//! stores the interview's choice under `versioning`, refusing what it does not know.
//!
//! Tag names are read from the files of the project's own `.git` and nothing else: no
//! program is started, no object and no configuration is read. The hostile fixtures below
//! (audit of 5433d7e007) pin that. Every `moai` child runs under a temporary HOME with a
//! clean environment, in temporary project folders only. Some fixtures are made with the
//! shared git runner, to read what a real git writes; the rest are written by hand.

mod common;

use common::cli::tree_state;
use common::home::TempHome;
use serde_json::{json, Value};
use std::{
    fs,
    io::Write,
    path::Path,
    process::{Output, Stdio},
    time::{Duration, Instant},
};
use tempfile::TempDir;

const HEX: &str = "1111111111111111111111111111111111111111";
const FOUR_MIB: usize = 4 * 1024 * 1024;

/// Runs `moai ARGS` and fails when it has not finished after `limit`.
fn run_within(home: &TempHome, args: &[&str], stdin: &[u8], limit: Duration) -> Output {
    let mut child = home
        .command(args)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    let mut pipe = child.stdin.take().unwrap();
    match pipe.write_all(stdin) {
        Ok(()) => {}
        Err(e) if e.kind() == std::io::ErrorKind::BrokenPipe => {}
        Err(e) => panic!("stdin write failed: {e}"),
    }
    drop(pipe);
    let start = Instant::now();
    while child.try_wait().unwrap().is_none() {
        if start.elapsed() > limit {
            child.kill().unwrap();
            panic!("moai {args:?} did not finish within {limit:?}");
        }
        std::thread::sleep(Duration::from_millis(10));
    }
    child.wait_with_output().unwrap()
}

fn run(home: &TempHome, args: &[&str], stdin: &[u8]) -> Output {
    run_within(home, args, stdin, Duration::from_secs(60))
}

fn result(out: &Output) -> Value {
    let v: Value = serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
        panic!(
            "{}{}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    });
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn versioning(home: &TempHome, root: &Path) -> Value {
    let out = run(home, &["project-detect", root.to_str().unwrap()], b"");
    result(&out)["versioning"].clone()
}

const IDENTITY: [&str; 8] = [
    "-c",
    "user.name=fixture",
    "-c",
    "user.email=fixture@example.invalid",
    "-c",
    "commit.gpgsign=false",
    "-c",
    "tag.gpgsign=false",
];

fn git(root: &Path, args: &[&str]) {
    let mut all = IDENTITY.to_vec();
    all.extend(args);
    common::git::git_pinned(root, &all, Some("1767225600"));
}

/// A real repository in `root` with one commit per tag, in the order given.
fn repo_with_tags(root: &Path, tags: &[&str]) {
    common::git::git_ok(root, &["init", "-q"]);
    for tag in tags {
        git(root, &["commit", "-q", "--allow-empty", "-m", "release"]);
        git(root, &["tag", tag]);
    }
}

fn write(path: &Path, bytes: &[u8]) {
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, bytes).unwrap();
}

/// A `.git` folder written by hand: no object, loose tag files only.
fn handmade_repo(root: &Path, tags: &[&str]) {
    write(&root.join(".git/HEAD"), b"ref: refs/heads/main\n");
    fs::create_dir_all(root.join(".git/objects")).unwrap();
    fs::create_dir_all(root.join(".git/refs/tags")).unwrap();
    for tag in tags {
        write(
            &root.join(".git/refs/tags").join(tag),
            format!("{HEX}\n").as_bytes(),
        );
    }
}

fn tag_values(v: &Value) -> Vec<String> {
    v["evidence"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|e| e["source"] == "tag")
        .map(|e| e["value"].as_str().unwrap().to_string())
        .collect()
}

fn assert_unavailable(v: &Value, why: &str) {
    assert_eq!(v["tags"], "unavailable", "{why}: {v}");
    assert!(
        v["tags_reason"].as_str().is_some_and(|r| !r.is_empty()),
        "{why}: no reason: {v}"
    );
    assert_eq!(v["scheme"], "none", "{why}: {v}");
    assert!(tag_values(v).is_empty(), "{why}: {v}");
}

/// Criterion 1: the tags name the scheme, the five highest versions are evidence, and the
/// read changes nothing in the project or the HOME.
#[test]
fn the_five_highest_tags_name_the_scheme_and_nothing_is_written() {
    let home = TempHome::new();
    let root = TempDir::new().unwrap();
    // v1.10.0 is made before v1.9.0: the order is by version, not by date.
    repo_with_tags(
        root.path(),
        &["v1.0.0", "v1.1.0", "v1.10.0", "v1.3.0", "v1.4.0", "v1.9.0"],
    );
    let before = tree_state(root.path());
    let v = versioning(&home, root.path());
    assert_eq!(v["scheme"], "semver", "{v}");
    assert_eq!(v["latest"], "v1.10.0");
    assert_eq!(v["tags"], "read");
    assert!(v.get("tags_reason").is_none(), "{v}");
    // Six tags exist; the five highest are shown, highest first, and the lowest is left out.
    assert_eq!(
        tag_values(&v),
        ["v1.10.0", "v1.9.0", "v1.4.0", "v1.3.0", "v1.1.0"]
    );
    assert_eq!(
        v["evidence"][0],
        json!({"source": "tag", "value": "v1.10.0", "shape": "semver"})
    );
    assert_eq!(
        tree_state(root.path()),
        before,
        "detection wrote into the project"
    );
    assert!(
        TempHome::names_in(home.path()).is_empty(),
        "detection wrote into HOME: {:?}",
        TempHome::names_in(home.path())
    );

    // Exactly five tags: all five are shown.
    let five = TempDir::new().unwrap();
    handmade_repo(
        five.path(),
        &["v0.1.0", "v0.2.0", "v0.3.0", "v0.4.0", "v0.5.0"],
    );
    assert_eq!(tag_values(&versioning(&home, five.path())).len(), 5);
}

/// What a real git writes is read: packed refs, the peeled line of an annotated tag, and a
/// tag whose name has a folder in it.
#[test]
fn packed_and_annotated_tags_of_a_real_repository_are_read() {
    let home = TempHome::new();
    let root = TempDir::new().unwrap();
    repo_with_tags(root.path(), &["v1.0.0", "release/v1.1.0"]);
    git(root.path(), &["tag", "-a", "-m", "annotated", "v1.2.0"]);
    git(root.path(), &["pack-refs", "--all"]);
    let packed = fs::read_to_string(root.path().join(".git/packed-refs")).unwrap();
    assert!(packed.lines().any(|l| l.starts_with('^')), "{packed}");
    // One more tag after the pack stays a loose file.
    git(root.path(), &["tag", "v1.3.0"]);
    assert!(root.path().join(".git/refs/tags/v1.3.0").is_file());
    let v = versioning(&home, root.path());
    assert_eq!(v["scheme"], "semver", "{v}");
    assert_eq!(
        tag_values(&v),
        ["v1.3.0", "v1.2.0", "release/v1.1.0", "v1.0.0"]
    );
}

#[test]
fn calver_headver_mixed_and_none_are_told_apart() {
    let home = TempHome::new();
    let calver = TempDir::new().unwrap();
    handmade_repo(calver.path(), &["2026.10.0", "2026.11.0-rc.1", "2026.11.0"]);
    let v = versioning(&home, calver.path());
    // A release ranks above its own pre-release.
    assert_eq!(
        (&v["scheme"], &v["latest"]),
        (&json!("calver"), &json!("2026.11.0")),
        "{v}"
    );
    assert_eq!(tag_values(&v), ["2026.11.0", "2026.11.0-rc.1", "2026.10.0"]);

    let headver = TempDir::new().unwrap();
    handmade_repo(headver.path(), &["3.2646.58", "3.2647.59"]);
    let v = versioning(&home, headver.path());
    assert_eq!(
        (&v["scheme"], &v["latest"]),
        (&json!("headver"), &json!("3.2647.59")),
        "{v}"
    );
    // Every HeadVer version is valid SemVer too; the row says so.
    assert!(
        v["evidence"][0]["note"]
            .as_str()
            .is_some_and(|n| n.contains("also valid SemVer")),
        "{v}"
    );

    let mixed = TempDir::new().unwrap();
    handmade_repo(mixed.path(), &["v3.2.1", "2026.11.0"]);
    let v = versioning(&home, mixed.path());
    assert_eq!(
        (&v["scheme"], &v["latest"]),
        (&json!("mixed"), &json!("2026.11.0")),
        "{v}"
    );

    // N2: a minor that walks through four digits is SemVer, not SemVer mixed with HeadVer.
    let walking = TempDir::new().unwrap();
    handmade_repo(walking.path(), &["v1.999.0", "v1.1000.0", "v1.1001.0"]);
    let v = versioning(&home, walking.path());
    assert_eq!(
        (&v["scheme"], &v["latest"]),
        (&json!("semver"), &json!("v1.1001.0")),
        "{v}"
    );

    // A repository without tags, and a folder that is no repository.
    let untagged = TempDir::new().unwrap();
    handmade_repo(untagged.path(), &[]);
    assert_eq!(
        versioning(&home, untagged.path()),
        json!({"scheme": "none", "latest": null, "evidence": [], "tags": "read"})
    );
    let plain = TempDir::new().unwrap();
    assert_eq!(
        versioning(&home, plain.path()),
        json!({"scheme": "none", "latest": null, "evidence": [], "tags": "no_repository"})
    );

    // Tags that are no versions are evidence and decide nothing.
    let named = TempDir::new().unwrap();
    handmade_repo(named.path(), &["nightly"]);
    let v = versioning(&home, named.path());
    assert_eq!(v["scheme"], "none", "{v}");
    assert_eq!(v["evidence"][0]["shape"], "unrecognised");
}

/// Without a recognised tag the root manifests decide, and their versions are evidence.
#[test]
fn manifest_versions_are_evidence_and_decide_without_tags() {
    let home = TempHome::new();
    let root = TempDir::new().unwrap();
    fs::write(
        root.path().join("Cargo.toml"),
        "[package]\nname = \"demo\"\nversion = \"0.3.1\"\n",
    )
    .unwrap();
    fs::write(
        root.path().join("package.json"),
        r#"{"name": "demo", "version": "0.3.1"}"#,
    )
    .unwrap();
    fs::write(
        root.path().join("pyproject.toml"),
        "[project]\nname = \"demo\"\nversion = \"0.3.1\"\n",
    )
    .unwrap();
    let v = versioning(&home, root.path());
    assert_eq!(v["scheme"], "semver", "{v}");
    assert_eq!(v["latest"], "0.3.1");
    assert_eq!(
        v["evidence"],
        json!([
            {"source": "Cargo.toml", "value": "0.3.1", "shape": "semver"},
            {"source": "package.json", "value": "0.3.1", "shape": "semver"},
            {"source": "pyproject.toml", "value": "0.3.1", "shape": "semver"},
        ])
    );

    // The tags decide when one has a shape; the manifest stays evidence.
    handmade_repo(root.path(), &["2026.11.0"]);
    let v = versioning(&home, root.path());
    assert_eq!(v["scheme"], "calver", "{v}");
    assert_eq!(v["latest"], "2026.11.0");
    assert_eq!(v["evidence"].as_array().unwrap().len(), 4);

    // A workspace version, a Poetry version, and values that are no plain version text.
    let other = TempDir::new().unwrap();
    fs::write(
        other.path().join("Cargo.toml"),
        "[workspace.package]\nversion = \"2026.11.0\"\n[package]\nname = \"demo\"\nversion.workspace = true\n",
    )
    .unwrap();
    fs::write(
        other.path().join("pyproject.toml"),
        "[tool.poetry]\nname = \"demo\"\nversion = \"2026.11.1\"\n",
    )
    .unwrap();
    fs::write(
        other.path().join("package.json"),
        r#"{"version": "1.0.0 && curl example.invalid"}"#,
    )
    .unwrap();
    let v = versioning(&home, other.path());
    assert_eq!(
        v["evidence"],
        json!([
            {"source": "Cargo.toml", "value": "2026.11.0", "shape": "calver"},
            {"source": "pyproject.toml", "value": "2026.11.1", "shape": "calver"},
        ]),
        "{v}"
    );
}

/// A folder inside another repository, without a `.git` of its own, reads no tags.
#[test]
fn a_parent_folders_repository_is_not_the_projects() {
    let home = TempHome::new();
    let outer = TempDir::new().unwrap();
    handmade_repo(outer.path(), &["v9.9.9"]);
    let inner = outer.path().join("packages").join("inner");
    fs::create_dir_all(&inner).unwrap();
    let v = versioning(&home, &inner);
    assert_eq!(v["tags"], "no_repository", "{v}");
    assert_eq!(v["scheme"], "none");
}

/// Audit B1: a repository that declares a promisor remote and holds a tag whose object is
/// missing. Reading its tags starts nothing (the remote's `uploadpack` would create FLAG) and
/// writes nothing (a lazy fetch would add a filter line to `.git/config`).
#[test]
fn a_promisor_remote_runs_nothing_and_the_repository_stays_byte_identical() {
    let home = TempHome::new();
    let root = TempDir::new().unwrap();
    let flag = root.path().join("FLAG");
    handmade_repo(root.path(), &["v1.0.0"]);
    let config = format!(
        "[core]\n\trepositoryformatversion = 1\n\tbare = false\n[extensions]\n\tpartialclone = origin\n[remote \"origin\"]\n\turl = ext::sh -c touch% {flag}\n\tpromisor = true\n\tuploadpack = touch {flag}\n\tfetch = +refs/heads/*:refs/remotes/origin/*\n",
        flag = flag.display()
    );
    write(&root.path().join(".git/config"), config.as_bytes());
    let before = tree_state(root.path());
    let v = versioning(&home, root.path());
    assert!(!flag.exists(), "the repository's uploadpack command ran");
    assert_eq!(
        fs::read(root.path().join(".git/config")).unwrap(),
        config.as_bytes(),
        ".git/config changed"
    );
    assert_eq!(tree_state(root.path()), before, "the repository changed");
    assert_eq!(v["tags"], "read", "{v}");
    assert_eq!(
        (&v["scheme"], &v["latest"]),
        (&json!("semver"), &json!("v1.0.0")),
        "{v}"
    );
}

/// Audit B2: `packed-refs` as a named pipe nobody writes to does not hold the command.
#[cfg(unix)]
#[test]
fn a_named_pipe_as_packed_refs_is_unavailable_within_two_seconds() {
    let home = TempHome::new();
    let root = TempDir::new().unwrap();
    handmade_repo(root.path(), &["v1.0.0"]);
    let pipe = root.path().join(".git/packed-refs");
    let made = std::process::Command::new("mkfifo")
        .arg(&pipe)
        .status()
        .unwrap();
    assert!(made.success());
    let out = run_within(
        &home,
        &["project-detect", root.path().to_str().unwrap()],
        b"",
        Duration::from_secs(2),
    );
    assert_unavailable(&result(&out)["versioning"], "named pipe");
}

/// No link is followed: not `.git`, not `packed-refs`, not an entry under `refs/tags`.
#[cfg(unix)]
#[test]
fn links_in_the_repository_are_not_followed() {
    use std::os::unix::fs::symlink;
    let home = TempHome::new();
    let elsewhere = TempDir::new().unwrap();
    handmade_repo(elsewhere.path(), &["v9.9.9"]);
    write(
        &elsewhere.path().join("packed"),
        format!("{HEX} refs/tags/v8.8.8\n").as_bytes(),
    );

    // `.git` itself is a link to another repository.
    let linked = TempDir::new().unwrap();
    symlink(elsewhere.path().join(".git"), linked.path().join(".git")).unwrap();
    assert_unavailable(&versioning(&home, linked.path()), "symlinked .git");

    // `packed-refs` is a link.
    let packed = TempDir::new().unwrap();
    handmade_repo(packed.path(), &["v1.0.0"]);
    symlink(
        elsewhere.path().join("packed"),
        packed.path().join(".git/packed-refs"),
    )
    .unwrap();
    assert_unavailable(&versioning(&home, packed.path()), "symlinked packed-refs");

    // `refs/tags` is a link.
    let tags = TempDir::new().unwrap();
    write(&tags.path().join(".git/HEAD"), b"ref: refs/heads/main\n");
    fs::create_dir_all(tags.path().join(".git/refs")).unwrap();
    symlink(
        elsewhere.path().join(".git/refs/tags"),
        tags.path().join(".git/refs/tags"),
    )
    .unwrap();
    assert_unavailable(&versioning(&home, tags.path()), "symlinked refs/tags");

    // Links inside `refs/tags`, to a file and to a folder, are no tags; the real one is read.
    let inside = TempDir::new().unwrap();
    handmade_repo(inside.path(), &["v1.0.0"]);
    symlink(
        elsewhere.path().join(".git/refs/tags/v9.9.9"),
        inside.path().join(".git/refs/tags/v7.7.7"),
    )
    .unwrap();
    symlink(
        elsewhere.path().join(".git/refs/tags"),
        inside.path().join(".git/refs/tags/release"),
    )
    .unwrap();
    let v = versioning(&home, inside.path());
    assert_eq!(v["tags"], "read", "{v}");
    assert_eq!(tag_values(&v), ["v1.0.0"]);
}

fn packed_refs(tags: usize) -> String {
    let mut text = String::from("# pack-refs with: peeled fully-peeled sorted \n");
    for i in 0..tags {
        text.push_str(&format!("{HEX} refs/tags/v0.0.{i}\n"));
    }
    text
}

/// The limits at their edges: `packed-refs` is read up to 4 MiB and not one byte beyond, and
/// `refs/tags` up to 2,000 entries and not one more. 200,000 packed tags are over the size.
#[test]
fn the_size_and_entry_limits_hold_at_their_edges() {
    let home = TempHome::new();
    let pad_to = |text: &str, len: usize| -> String {
        let mut out = text.to_string();
        let fill = len - out.len();
        out.push('#');
        out.push_str(&"x".repeat(fill - 2));
        out.push('\n');
        assert_eq!(out.len(), len);
        out
    };
    let at = TempDir::new().unwrap();
    handmade_repo(at.path(), &[]);
    write(
        &at.path().join(".git/packed-refs"),
        pad_to(&packed_refs(3), FOUR_MIB).as_bytes(),
    );
    let v = versioning(&home, at.path());
    assert_eq!(v["tags"], "read", "{v}");
    assert_eq!(tag_values(&v), ["v0.0.2", "v0.0.1", "v0.0.0"]);

    let over = TempDir::new().unwrap();
    handmade_repo(over.path(), &[]);
    write(
        &over.path().join(".git/packed-refs"),
        pad_to(&packed_refs(3), FOUR_MIB + 1).as_bytes(),
    );
    assert_unavailable(&versioning(&home, over.path()), "4 MiB + 1");

    let many = TempDir::new().unwrap();
    handmade_repo(many.path(), &[]);
    let text = packed_refs(200_000);
    assert!(text.len() > FOUR_MIB);
    write(&many.path().join(".git/packed-refs"), text.as_bytes());
    let out = run_within(
        &home,
        &["project-detect", many.path().to_str().unwrap()],
        b"",
        Duration::from_secs(5),
    );
    assert_unavailable(&result(&out)["versioning"], "200,000 packed tags");

    let names = |n: usize| -> Vec<String> { (0..n).map(|i| format!("v0.0.{i}")).collect() };
    let full = TempDir::new().unwrap();
    let list = names(2000);
    handmade_repo(
        full.path(),
        &list.iter().map(String::as_str).collect::<Vec<_>>(),
    );
    let v = versioning(&home, full.path());
    assert_eq!(v["tags"], "read", "{v}");
    assert_eq!(v["latest"], "v0.0.1999");
    assert_eq!(tag_values(&v).len(), 5);

    let beyond = TempDir::new().unwrap();
    let list = names(2001);
    handmade_repo(
        beyond.path(),
        &list.iter().map(String::as_str).collect::<Vec<_>>(),
    );
    assert_unavailable(&versioning(&home, beyond.path()), "2,001 entries");

    // Folders are entered four levels deep and no further.
    let deep = TempDir::new().unwrap();
    handmade_repo(deep.path(), &["a/b/c/v1.0.0", "a/b/c/d/v2.0.0"]);
    let v = versioning(&home, deep.path());
    assert_eq!(tag_values(&v), ["a/b/c/v1.0.0"], "{v}");
}

/// A `.git` file is followed only for a linked worktree of a repository that names this
/// folder back; a pointer to anything else is `unavailable`.
#[test]
fn a_git_file_is_followed_only_for_a_linked_worktree() {
    let home = TempHome::new();
    let main = TempDir::new().unwrap();
    repo_with_tags(main.path(), &["2026.10.0", "2026.11.0"]);
    let holder = TempDir::new().unwrap();
    let worktree = holder.path().join("wt");
    git(
        main.path(),
        &[
            "worktree",
            "add",
            "-q",
            "--detach",
            worktree.to_str().unwrap(),
        ],
    );
    assert!(worktree.join(".git").is_file());
    let v = versioning(&home, &worktree);
    assert_eq!(v["tags"], "read", "{v}");
    assert_eq!(
        (&v["scheme"], &v["latest"]),
        (&json!("calver"), &json!("2026.11.0")),
        "{v}"
    );

    // A copy of that pointer in another folder: the repository does not name it back.
    let copy = TempDir::new().unwrap();
    fs::copy(worktree.join(".git"), copy.path().join(".git")).unwrap();
    assert_unavailable(&versioning(&home, copy.path()), "copied worktree pointer");

    // A pointer straight at an unrelated repository's folder.
    let unrelated = TempDir::new().unwrap();
    write(
        &unrelated.path().join(".git"),
        format!("gitdir: {}\n", main.path().join(".git").display()).as_bytes(),
    );
    assert_unavailable(&versioning(&home, unrelated.path()), "unrelated repository");

    // A submodule's pointer: a folder with no `commondir` file.
    let parent = TempDir::new().unwrap();
    handmade_repo(parent.path(), &["v9.9.9"]);
    let module = parent.path().join(".git/modules/sub");
    write(&module.join("HEAD"), b"ref: refs/heads/main\n");
    write(
        &module.join("refs/tags/v5.0.0"),
        format!("{HEX}\n").as_bytes(),
    );
    write(
        &parent.path().join("sub/.git"),
        b"gitdir: ../.git/modules/sub\n",
    );
    assert_unavailable(
        &versioning(&home, &parent.path().join("sub")),
        "submodule pointer",
    );

    // Not a pointer at all, and a pointer to nothing.
    for (text, why) in [
        ("not a pointer\n", "other text"),
        ("gitdir: missing/folder\n", "missing target"),
        ("", "empty file"),
    ] {
        let odd = TempDir::new().unwrap();
        write(&odd.path().join(".git"), text.as_bytes());
        assert_unavailable(&versioning(&home, odd.path()), why);
    }
}

/// Audit N6 and its surviving mutant: a tag name that is not short, plain text never
/// reaches the output, and it does not take the place of a tag that is.
#[test]
fn hostile_tag_names_never_reach_the_output() {
    let home = TempHome::new();
    let root = TempDir::new().unwrap();
    let longest = format!("v1.0.0-{}", "a".repeat(57));
    let too_long = format!("v9.0.0-{}", "a".repeat(58));
    assert_eq!((longest.len(), too_long.len()), (64, 65));
    handmade_repo(
        root.path(),
        &[
            "v1.0.0",
            &longest,
            &too_long,
            "v9.1.0$(touch x)",
            "v9.2.0`id`",
            "v9.3.0<!--x-->",
            "v9.4.0 space",
            "v9.5.0;x",
        ],
    );
    let mut packed = packed_refs(1).into_bytes();
    for hostile in [
        &b"v9.6.0\x1b[31m"[..],
        b"v9.7.0\ttab",
        b"v9.8.0\rcr",
        "v9.9.0-\u{d55c}\u{ae00}".as_bytes(),
        b"v9.10.0-\xff\xfe",
        b"v9.11.0\"quote",
        b"v9.12.0\\slash",
    ] {
        packed.extend(format!("{HEX} refs/tags/").as_bytes());
        packed.extend(hostile);
        packed.push(b'\n');
    }
    // A newline in a name is a line break here: its second half is no ref line.
    packed.extend(format!("{HEX} refs/tags/v9.13.0\n$(second half)\n").as_bytes());
    // Lines that are no tag: a branch, a short id, a peeled line.
    packed.extend(format!("{HEX} refs/heads/v8.0.0\n").as_bytes());
    packed.extend(b"1111 refs/tags/v8.1.0\n");
    packed.extend(format!("^{HEX}\n").as_bytes());
    write(&root.path().join(".git/packed-refs"), &packed);

    let out = run(
        &home,
        &["project-detect", root.path().to_str().unwrap()],
        b"",
    );
    let text = String::from_utf8(out.stdout.clone()).unwrap();
    let v = result(&out)["versioning"].clone();
    assert_eq!(v["tags"], "read", "{v}");
    // The plain names are all there, the highest first; none of the others took a place.
    assert_eq!(
        tag_values(&v),
        ["v9.13.0", "v1.0.0", longest.as_str(), "v0.0.0"],
        "{v}"
    );
    let shown = v.to_string();
    for part in [
        "touch",
        "`",
        "<!--",
        "space",
        ";x",
        "\\u001b",
        "[31m",
        "tab",
        "cr",
        "\u{d55c}",
        "quote",
        "slash",
        "second half",
        "v8.",
        "v9.0.0",
        "v9.1.0",
        "v9.6.0",
        "\u{fffd}",
    ] {
        assert!(
            !shown.contains(part),
            "{part:?} reached the output: {shown}"
        );
    }
    assert!(!text.contains("\\u001b"), "{text}");
}

fn profile(versioning: Option<Value>) -> Value {
    let mut v = json!({"schema_version": 2, "project_id": "demo", "purpose": "Build a small editor",
        "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude"],
        "interview_revision": "c69-fixture-1"});
    if let Some(versioning) = versioning {
        v["versioning"] = versioning;
    }
    v
}

fn plan(home: &TempHome, root: &TempDir, input: &Value) -> Output {
    run(
        home,
        &["project-plan", root.path().to_str().unwrap()],
        input.to_string().as_bytes(),
    )
}

fn apply(home: &TempHome, root: &TempDir, plan: &Value) {
    let out = run(
        home,
        &[
            "project-apply",
            root.path().to_str().unwrap(),
            "--plan-digest",
            plan["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "fixture-approved",
        ],
        plan.to_string().as_bytes(),
    );
    result(&out);
}

/// Criterion 3: the choice, with the CalVer format or the HeadVer head, is planned, written
/// and read back unchanged; a profile without it does not gain the key.
#[test]
fn the_choice_is_stored_in_the_profile_and_read_back() {
    let home = TempHome::new();
    for choice in [
        json!({"scheme": "semver"}),
        json!({"scheme": "calver", "format": "YYYY.0M.MICRO"}),
        json!({"scheme": "headver", "head": 3}),
        json!({"scheme": "headver", "head": 9999}),
        json!({"scheme": "custom", "format": "release train letter plus build number"}),
        json!({"scheme": "custom", "format": "\u{d55c}".repeat(128)}),
    ] {
        let root = TempDir::new().unwrap();
        let first = result(&plan(&home, &root, &profile(Some(choice.clone()))));
        assert_eq!(first["action"], "create", "{choice}");
        assert_eq!(first["profile"]["versioning"], choice);
        apply(&home, &root, &first);
        let written: Value =
            serde_json::from_slice(&fs::read(root.path().join(".moai/project.yaml")).unwrap())
                .unwrap();
        assert_eq!(written["versioning"], choice);
        let again = result(&plan(&home, &root, &profile(Some(choice.clone()))));
        assert_eq!(again["action"], "unchanged", "{choice}");
        // `project-detect` hands the stored choice back for the next plan.
        let detect = result(&run(
            &home,
            &["project-detect", root.path().to_str().unwrap()],
            b"",
        ));
        assert_eq!(detect["profile"]["versioning"], choice);
    }

    let root = TempDir::new().unwrap();
    let p = result(&plan(&home, &root, &profile(None)));
    assert!(p["profile"].get("versioning").is_none(), "{p}");
    apply(&home, &root, &p);
    let written: Value =
        serde_json::from_slice(&fs::read(root.path().join(".moai/project.yaml")).unwrap()).unwrap();
    assert!(written.get("versioning").is_none(), "{written}");
    // An interview answer added later is an update of that one key.
    let later = result(&plan(
        &home,
        &root,
        &profile(Some(json!({"scheme": "semver"}))),
    ));
    assert_eq!(later["action"], "update", "{later}");
    assert_eq!(later["delta"]["added"], json!(["versioning"]));
}

#[test]
fn an_unknown_choice_is_refused_and_nothing_is_written() {
    let home = TempHome::new();
    for (bad, why) in [
        (json!({"scheme": "romantic"}), "unknown scheme"),
        (json!({"scheme": "SemVer"}), "upper-case name"),
        (
            json!({"scheme": "semver", "head": 2}),
            "head without HeadVer",
        ),
        (
            json!({"scheme": "semver", "format": "MAJOR.MINOR.PATCH"}),
            "format without CalVer or custom",
        ),
        (
            json!({"scheme": "calver", "format": "YYYY-ish"}),
            "unknown CalVer token",
        ),
        (
            json!({"scheme": "calver", "format": "YYYY.0M.0M"}),
            "a token twice",
        ),
        (
            json!({"scheme": "calver", "format": "MAJOR.YYYY.MICRO"}),
            "the year is not first",
        ),
        (json!({"scheme": "headver", "head": -1}), "negative head"),
        (
            json!({"scheme": "headver", "head": 10000}),
            "head over 9999",
        ),
        (json!({"scheme": "calver", "extra": true}), "unknown field"),
        (json!("calver"), "a string"),
        (json!(["semver"]), "an array"),
        (json!(["calver", "YYYY.0M.MICRO"]), "an array of two"),
        (json!({}), "no scheme"),
        (
            json!({"scheme": "custom", "format": "\u{d55c}".repeat(129)}),
            "129 characters",
        ),
        (
            json!({"scheme": "custom", "format": "x <!-- y"}),
            "an angle bracket",
        ),
        (
            json!({"scheme": "custom", "format": "run `id` first"}),
            "a backtick",
        ),
        (
            json!({"scheme": "custom", "format": "# Heading"}),
            "a heading mark at the line start",
        ),
        (
            json!({"scheme": "custom", "format": "a\u{202e}b"}),
            "a direction override",
        ),
    ] {
        let root = TempDir::new().unwrap();
        let out = plan(&home, &root, &profile(Some(bad)));
        assert!(!out.status.success(), "{why}");
        assert!(
            !root.path().join(".moai/project.yaml").exists(),
            "{why}: a refused plan wrote a profile"
        );
    }
}
