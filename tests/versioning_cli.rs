//! Card C-69 through the binary: `project-detect` reports the versioning scheme a project
//! already uses from its newest tags and root manifests, read only; and the project profile
//! stores the interview's choice under `versioning`, refusing what it does not know.
//!
//! Every `moai` child runs under a temporary HOME with a clean environment, in temporary
//! project folders only. The fixture repositories are made with the shared git runner.

mod common;

use common::cli::tree_state;
use common::home::TempHome;
use serde_json::{json, Value};
use std::{
    fs,
    io::Write,
    path::Path,
    process::{Output, Stdio},
};
use tempfile::TempDir;

fn run_env(home: &TempHome, args: &[&str], stdin: &[u8], env: &[(&str, &Path)]) -> Output {
    let mut cmd = home.command(args);
    for (key, value) in env {
        cmd.env(key, value);
    }
    let mut child = cmd
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
    child.wait_with_output().unwrap()
}

fn run(home: &TempHome, args: &[&str], stdin: &[u8]) -> Output {
    run_env(home, args, stdin, &[])
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

/// A repository in `root` with one commit per tag, oldest first, one day apart, so "newest"
/// does not depend on the clock.
fn repo_with_tags(root: &Path, tags: &[&str]) {
    let identity = [
        "-c",
        "user.name=fixture",
        "-c",
        "user.email=fixture@example.invalid",
        "-c",
        "commit.gpgsign=false",
        "-c",
        "tag.gpgsign=false",
    ];
    common::git::git_ok(root, &["init", "-q"]);
    for (i, tag) in tags.iter().enumerate() {
        let date = format!("{}", 1_767_225_600 + 86_400 * i as u64);
        let mut commit = identity.to_vec();
        commit.extend(["commit", "-q", "--allow-empty", "-m", "release"]);
        common::git::git_pinned(root, &commit, Some(&date));
        let mut tagging = identity.to_vec();
        tagging.extend(["tag", tag]);
        common::git::git_pinned(root, &tagging, Some(&date));
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

/// Criterion 1: the newest tags name the scheme, at most five of them are evidence, and the
/// read changes nothing in the project or the HOME.
#[test]
fn the_newest_five_tags_name_the_scheme_and_nothing_is_written() {
    let home = TempHome::new();
    let root = TempDir::new().unwrap();
    repo_with_tags(
        root.path(),
        &["v1.0.0", "v1.1.0", "v1.2.0", "v1.3.0", "v1.4.0", "v1.5.0"],
    );
    let before = tree_state(root.path());
    let v = versioning(&home, root.path());
    assert_eq!(v["scheme"], "semver", "{v}");
    assert_eq!(v["latest"], "v1.5.0");
    assert_eq!(v["tags"], "read");
    // Six tags exist; the five newest are shown, newest first, and the oldest is left out.
    assert_eq!(
        tag_values(&v),
        ["v1.5.0", "v1.4.0", "v1.3.0", "v1.2.0", "v1.1.0"]
    );
    assert_eq!(
        v["evidence"][0],
        json!({"source": "tag", "value": "v1.5.0", "shape": "semver"})
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
    repo_with_tags(
        five.path(),
        &["v0.1.0", "v0.2.0", "v0.3.0", "v0.4.0", "v0.5.0"],
    );
    assert_eq!(tag_values(&versioning(&home, five.path())).len(), 5);
}

#[test]
fn calver_headver_mixed_and_none_are_told_apart() {
    let home = TempHome::new();
    let calver = TempDir::new().unwrap();
    repo_with_tags(calver.path(), &["2026.10.0", "2026.11.0-rc.1", "2026.11.0"]);
    let v = versioning(&home, calver.path());
    assert_eq!(
        (&v["scheme"], &v["latest"]),
        (&json!("calver"), &json!("2026.11.0")),
        "{v}"
    );

    let headver = TempDir::new().unwrap();
    repo_with_tags(headver.path(), &["3.2646.58", "3.2647.59"]);
    let v = versioning(&home, headver.path());
    assert_eq!(
        (&v["scheme"], &v["latest"]),
        (&json!("headver"), &json!("3.2647.59")),
        "{v}"
    );

    let mixed = TempDir::new().unwrap();
    repo_with_tags(mixed.path(), &["v3.2.1", "2026.11.0"]);
    let v = versioning(&home, mixed.path());
    assert_eq!(
        (&v["scheme"], &v["latest"]),
        (&json!("mixed"), &json!("2026.11.0")),
        "{v}"
    );

    // A repository without tags, and a folder that is no repository.
    let untagged = TempDir::new().unwrap();
    repo_with_tags(untagged.path(), &[]);
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
    repo_with_tags(named.path(), &["nightly"]);
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
    repo_with_tags(root.path(), &["2026.11.0"]);
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

/// The tags read are the project's own: a folder inside another repository without its own
/// `.git` reads none, and a `GIT_DIR` in the caller's environment does not redirect the read.
#[test]
fn only_the_projects_own_repository_is_read() {
    let home = TempHome::new();
    let outer = TempDir::new().unwrap();
    repo_with_tags(outer.path(), &["v9.9.9"]);
    let inner = outer.path().join("packages").join("inner");
    fs::create_dir_all(&inner).unwrap();
    let v = versioning(&home, &inner);
    assert_eq!(v["tags"], "no_repository", "{v}");
    assert_eq!(v["scheme"], "none");

    let project = TempDir::new().unwrap();
    repo_with_tags(project.path(), &["2026.11.0"]);
    let foreign = outer.path().join(".git");
    let out = run_env(
        &home,
        &["project-detect", project.path().to_str().unwrap()],
        b"",
        &[("GIT_DIR", &foreign)],
    );
    let v = result(&out)["versioning"].clone();
    assert_eq!(tag_values(&v), ["2026.11.0"], "{v}");
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
        json!({"scheme": "custom", "format": "release train letter plus build number"}),
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
        (json!({"scheme": "headver", "head": -1}), "negative head"),
        (json!({"scheme": "calver", "extra": true}), "unknown field"),
        (json!("calver"), "wrong shape"),
        (json!({}), "no scheme"),
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
