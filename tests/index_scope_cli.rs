//! SPEC-20261119-index-scope (card C-58): `index-build` can be scoped to folders, and a refusal
//! over the file limit names where the files are.
//!
//! Measured first (AC-1): the limit counts the tracked files of the scan only; gitignored,
//! untracked files such as `node_modules` never count, a committed one does. Then the scope
//! itself: a request field `scope`, recorded in the manifest, judged by `index-query`, with its
//! own N and N+1 edge at 16384.
//!
//! Every case runs against a temporary git repository; the real HOME and every other project
//! are untouched. Large counts use index-only entries (`git update-index --index-info`), as
//! `tests/index_cli.rs` does: the scan skips a tracked path that is absent on disk.
mod common;
use serde_json::{json, Value};
use std::{
    fs,
    io::Write,
    path::{Path, PathBuf},
    process::{Command, Stdio},
};

const LIMIT: usize = 16_384;

fn command(dir: &Path, program: &str) -> Command {
    let mut command = Command::new(program);
    command
        .current_dir(dir)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .env("GIT_AUTHOR_NAME", "t")
        .env("GIT_AUTHOR_EMAIL", "t@example.com")
        .env("GIT_COMMITTER_NAME", "t")
        .env("GIT_COMMITTER_EMAIL", "t@example.com");
    command
}

fn git(dir: &Path, args: &[&str]) {
    let status = command(dir, "git")
        .args(["-c", "gc.auto=0", "-c", "maintenance.auto=false"])
        .args(args)
        .status()
        .unwrap();
    assert!(status.success(), "git {args:?} failed");
}

fn git_stdin(dir: &Path, args: &[&str], input: &[u8]) {
    let mut child = command(dir, "git")
        .args(args)
        .stdin(Stdio::piped())
        .stdout(Stdio::null())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    child.stdin.take().unwrap().write_all(input).unwrap();
    let out = child.wait_with_output().unwrap();
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
}

/// A repository with one committed file, `a.txt`.
fn setup() -> (tempfile::TempDir, PathBuf) {
    let tmp = tempfile::tempdir().unwrap();
    let project = tmp.path().join("project");
    fs::create_dir_all(&project).unwrap();
    git(&project, &["init", "-q", "--initial-branch=main"]);
    fs::write(project.join("a.txt"), b"alpha").unwrap();
    git(&project, &["add", "."]);
    git(&project, &["commit", "-q", "-m", "init"]);
    (tmp, project)
}

fn write(project: &Path, relative: &str, text: &str) {
    let path = project.join(relative);
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, text).unwrap();
}

/// Commit every file of the working tree.
fn commit_all(project: &Path) {
    git(project, &["add", "-A"]);
    git(project, &["commit", "-q", "-m", "files"]);
}

/// Make the index hold exactly `paths` (all pointing at the blob of `a.txt`): paths absent on
/// disk are tracked entries the scan skips, so a count of 16 000 costs no files.
fn set_index(project: &Path, paths: &[String]) {
    let out = command(project, "git")
        .args(["rev-parse", "HEAD:a.txt"])
        .output()
        .unwrap();
    let hash = String::from_utf8(out.stdout).unwrap().trim().to_string();
    git(project, &["read-tree", "--empty"]);
    let mut entries = Vec::new();
    for path in paths {
        writeln!(entries, "100644 {hash}\t{path}").unwrap();
    }
    git_stdin(project, &["update-index", "--index-info"], &entries);
}

fn numbered(folder: &str, count: usize) -> Vec<String> {
    (0..count).map(|n| format!("{folder}/f-{n:05}")).collect()
}

fn cli(project: &Path, op: &str, stdin: &str) -> std::process::Output {
    // HOME is a folder of its own, never the real one and never inside the project: the
    // binary must not see a file the build would count.
    let home = tempfile::tempdir().unwrap();
    let mut child = command(project, common::MOAI_BIN)
        .env("HOME", home.path())
        .args([op, "."])
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
    child.wait_with_output().unwrap()
}

fn envelope(out: &std::process::Output) -> Value {
    serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
        panic!(
            "stdout is not JSON: {e}\n{}\n{}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    })
}

fn build(project: &Path, request: Value) -> std::process::Output {
    cli(project, "index-build", &request.to_string())
}

fn built(project: &Path, request: Value) -> Value {
    let out = build(project, request);
    assert!(
        out.status.success(),
        "{}\n{}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    envelope(&out)["result"].clone()
}

/// The error of a refused request: `(kind, retryable, message)`.
fn refused(out: &std::process::Output) -> (String, bool, String) {
    assert_eq!(
        out.status.code(),
        Some(1),
        "{}",
        String::from_utf8_lossy(&out.stdout)
    );
    let v = envelope(out);
    assert_eq!(v["ok"], json!(false), "{v}");
    (
        v["error"]["kind"].as_str().unwrap().to_string(),
        v["error"]["retryable"].as_bool().unwrap(),
        v["error"]["message"].as_str().unwrap().to_string(),
    )
}

fn query(project: &Path) -> Value {
    let out = cli(project, "index-query", r#"{"protocol_version":1}"#);
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    envelope(&out)["result"].clone()
}

fn index_dir(project: &Path) -> PathBuf {
    project.join(".moai-cli/index")
}

fn pointer_gen(project: &Path) -> Option<String> {
    let bytes = fs::read(index_dir(project).join("current.json")).ok()?;
    let v: Value = serde_json::from_slice(&bytes).ok()?;
    v["generation_id"].as_str().map(str::to_string)
}

fn manifest(project: &Path) -> (String, Value) {
    let gen = pointer_gen(project).expect("a published generation");
    let text = fs::read_to_string(
        index_dir(project)
            .join("generations")
            .join(&gen)
            .join("manifest.json"),
    )
    .unwrap();
    let v = serde_json::from_str(&text).unwrap();
    (text, v)
}

fn manifest_paths(project: &Path) -> Vec<String> {
    manifest(project).1["files"]
        .as_array()
        .unwrap()
        .iter()
        .map(|f| f["path"].as_str().unwrap().to_string())
        .collect()
}

/// AC-1 (REQ-4): measured in a fixture. The limit counts tracked files; ignored untracked
/// files do not count, and the same folder committed does.
#[test]
fn the_limit_counts_tracked_files_only() {
    let (_tmp, project) = setup();
    write(&project, ".gitignore", "node_modules/\n");
    git(&project, &["add", ".gitignore"]);
    git(&project, &["commit", "-q", "-m", "ignore"]);
    for n in 0..20 {
        write(&project, &format!("node_modules/pkg/x{n}.js"), "x");
    }
    let ignored = built(&project, json!({"protocol_version": 1}));
    assert_eq!(
        ignored["files"],
        json!(2),
        "a.txt and .gitignore: {ignored}"
    );

    git(&project, &["add", "-f", "node_modules"]);
    git(&project, &["commit", "-q", "-m", "vendor"]);
    let tracked = built(&project, json!({"protocol_version": 1}));
    assert_eq!(
        tracked["files"],
        json!(22),
        "the committed folder counts: {tracked}"
    );
}

/// AC-2 (REQ-3): 16384 files build, 16385 are refused, and the refusal says where the files are.
#[test]
fn the_limit_is_n_built_and_n_plus_one_refused_naming_the_folders() {
    let (_tmp, project) = setup();
    let mut paths = vec!["a.txt".to_string()];
    paths.extend(numbered("apps/bo", 9000));
    paths.extend(numbered("apps/web", 5000));
    paths.extend(numbered("packages/ui", LIMIT - 1 - 9000 - 5000));
    assert_eq!(paths.len(), LIMIT);
    set_index(&project, &paths);
    let accepted = built(&project, json!({"protocol_version": 1}));
    assert_eq!(
        accepted["files"],
        json!(1),
        "only a.txt is on disk: {accepted}"
    );
    let previous = pointer_gen(&project).unwrap();

    paths.push("apps/bo/one-more".into());
    set_index(&project, &paths);
    let (kind, retryable, message) = refused(&build(&project, json!({"protocol_version": 1})));
    assert_eq!(kind, "invalid_input", "{message}");
    assert!(!retryable);
    assert!(
        message.contains("index file limit exceeded (16384)"),
        "{message}"
    );
    assert!(message.contains("16385"), "the count is named: {message}");
    let bo = message
        .find("apps/bo (9001)")
        .unwrap_or_else(|| panic!("{message}"));
    let web = message
        .find("apps/web (5000)")
        .unwrap_or_else(|| panic!("{message}"));
    let ui = message
        .find("packages/ui (2383)")
        .unwrap_or_else(|| panic!("{message}"));
    assert!(bo < web && web < ui, "largest first: {message}");
    assert!(
        message.contains(r#"{"protocol_version":1,"scope":["apps/bo"]}"#),
        "the accepted request is shown: {message}"
    );
    assert_eq!(
        pointer_gen(&project).as_deref(),
        Some(previous.as_str()),
        "the previous generation stays"
    );
}

/// AC-3 (REQ-1): only the files under the scope are indexed, whole path components are
/// matched, a trailing slash is the same folder, and the manifest records the scope.
#[test]
fn a_scoped_build_indexes_only_the_scope_and_records_it() {
    let (_tmp, project) = setup();
    write(&project, "apps/bo/a.ts", "bo");
    write(&project, "apps/bobby/x.ts", "bobby");
    write(&project, "apps/web/b.ts", "web");
    write(&project, "packages/ui/c.ts", "ui");
    commit_all(&project);

    let one = built(
        &project,
        json!({"protocol_version": 1, "scope": ["apps/bo"]}),
    );
    assert_eq!(one["files"], json!(1), "{one}");
    assert_eq!(manifest_paths(&project), vec!["apps/bo/a.ts"]);
    assert_eq!(manifest(&project).1["scope"], json!(["apps/bo"]));

    let two = built(
        &project,
        json!({"protocol_version": 1, "scope": ["packages/ui/", "apps/bo", "apps/bo/"]}),
    );
    assert_eq!(two["files"], json!(2), "{two}");
    assert_eq!(
        manifest_paths(&project),
        vec!["apps/bo/a.ts", "packages/ui/c.ts"]
    );
    assert_eq!(
        manifest(&project).1["scope"],
        json!(["apps/bo", "packages/ui"])
    );
}

/// AC-3 (REQ-6, REQ-7): an unscoped manifest has no `scope` key, an empty list is unscoped,
/// and a scope changes the generation id.
#[test]
fn an_unscoped_manifest_has_no_scope_key() {
    let (_tmp, project) = setup();
    write(&project, "apps/bo/a.ts", "bo");
    commit_all(&project);
    let plain = built(&project, json!({"protocol_version": 1}));
    assert!(
        !manifest(&project).0.contains("\"scope\""),
        "{}",
        manifest(&project).0
    );
    let empty = built(&project, json!({"protocol_version": 1, "scope": []}));
    assert_eq!(
        empty["generation_id"], plain["generation_id"],
        "an empty scope is no scope"
    );
    assert_eq!(empty["action"], json!("unchanged"));

    let everything = built(
        &project,
        json!({"protocol_version": 1, "scope": ["apps", "a.txt"]}),
    );
    assert_eq!(everything["files"], plain["files"], "the same files");
    assert_ne!(
        everything["generation_id"], plain["generation_id"],
        "the scope is part of the id even when it covers the same files"
    );
}

/// AC-4 (REQ-1, REQ-3): a repository over the limit builds with a scope under it, and the
/// scoped count has its own N and N+1 edge.
#[test]
fn a_scope_builds_a_repository_over_the_limit() {
    let (_tmp, project) = setup();
    let mut paths = vec!["a.txt".to_string()];
    paths.extend(numbered("apps/bo", LIMIT - 1));
    paths.extend(numbered("apps/web", 100));
    set_index(&project, &paths);
    let (kind, _, message) = refused(&build(&project, json!({"protocol_version": 1})));
    assert_eq!(kind, "invalid_input", "{message}");
    assert!(message.contains("apps/bo (16383)"), "{message}");

    // 16383 under apps/bo (a.txt is outside the scope): below the edge.
    let below = built(
        &project,
        json!({"protocol_version": 1, "scope": ["apps/bo"]}),
    );
    assert_eq!(
        below["files"],
        json!(0),
        "nothing of the scope is on disk: {below}"
    );
    // Exactly 16384 under the scope: accepted.
    let mut at = paths.clone();
    at.push("apps/bo/edge-0".into());
    set_index(&project, &at);
    built(
        &project,
        json!({"protocol_version": 1, "scope": ["apps/bo"]}),
    );
    // 16385 under the scope: refused, naming the folder.
    at.push("apps/bo/edge-1".into());
    set_index(&project, &at);
    let (kind, retryable, message) = refused(&build(
        &project,
        json!({"protocol_version": 1, "scope": ["apps/bo"]}),
    ));
    assert_eq!(
        (kind.as_str(), retryable),
        ("invalid_input", false),
        "{message}"
    );
    assert!(
        message.contains("index file limit exceeded (16384)"),
        "{message}"
    );
    assert!(message.contains("apps/bo (16385)"), "{message}");
    assert!(
        message.contains("16385 tracked files in the scope"),
        "a refusal inside a scope says so: {message}"
    );
}

/// The request a refusal shows, read back out of its text: the JSON after `for example`.
fn suggested_request(message: &str) -> Value {
    let at = message
        .find("for example ")
        .unwrap_or_else(|| panic!("no example request in: {message}"))
        + "for example ".len();
    serde_json::from_str(message[at..].trim_end_matches('.'))
        .unwrap_or_else(|e| panic!("the example is not JSON ({e}): {message}"))
}

/// AC-2 (REQ-3), audit B1: the request a refusal shows is one that is accepted. A folder that is
/// over the limit by itself is never the suggestion; with no folder that fits there is no
/// example; the root bucket (`.`) is never one; and a folder grouped by its first component is
/// judged by everything under it, not by its direct files.
#[test]
fn the_request_a_refusal_shows_is_one_that_is_accepted() {
    // One folder holds everything: nothing fits, so there is nothing to suggest.
    let (_tmp, project) = setup();
    let mut paths = vec!["a.txt".to_string()];
    paths.extend(numbered("apps/bo", LIMIT + 1));
    set_index(&project, &paths);
    let (kind, _, message) = refused(&build(&project, json!({"protocol_version": 1})));
    assert_eq!(kind, "invalid_input", "{message}");
    assert!(message.contains("apps/bo (16385)"), "{message}");
    assert!(
        !message.contains(r#""scope":["#),
        "the folder is over the limit by itself, no example: {message}"
    );
    assert!(
        message.contains("No listed folder fits under the limit by itself"),
        "{message}"
    );
    // Following the old advice inside the scope gives the same refusal, still without an example.
    let (_, _, again) = refused(&build(
        &project,
        json!({"protocol_version": 1, "scope": ["apps/bo"]}),
    ));
    assert!(!again.contains(r#""scope":["#), "{again}");

    // Root files are the largest group: the example skips `.` and names the folder that fits,
    // and sending it builds.
    let (_tmp, project) = setup();
    let mut paths: Vec<String> = (0..LIMIT).map(|n| format!("r-{n:05}")).collect();
    paths.extend(numbered("apps/bo", 3));
    set_index(&project, &paths);
    let (_, _, message) = refused(&build(&project, json!({"protocol_version": 1})));
    assert!(message.contains(". (16384)"), "the r-* files: {message}");
    let suggestion = suggested_request(&message);
    assert_eq!(
        suggestion,
        json!({"protocol_version": 1, "scope": ["apps/bo"]}),
        "{message}"
    );
    built(&project, suggestion);

    // `tools` has 6000 direct files and five subfolders of 5000: it is the largest bucket but
    // 31000 files in all, so the example is the next folder, and sending it builds.
    let (_tmp, project) = setup();
    let mut paths = vec!["a.txt".to_string()];
    paths.extend(numbered("tools", 6000));
    for sub in ["a", "b", "c", "d", "e"] {
        paths.extend(numbered(&format!("tools/{sub}"), 5000));
    }
    set_index(&project, &paths);
    let (_, _, message) = refused(&build(&project, json!({"protocol_version": 1})));
    assert!(message.contains("tools (6000)"), "{message}");
    let suggestion = suggested_request(&message);
    assert_eq!(
        suggestion,
        json!({"protocol_version": 1, "scope": ["tools/a"]}),
        "{message}"
    );
    built(&project, suggestion);
}

/// Audit N3: after a refused build leaves a failure marker, a scoped generation is still judged
/// inside its scope (the marker branch of `index-query` reads the same scope).
#[test]
fn a_failure_marker_keeps_a_scoped_generation_scoped() {
    let (_tmp, project) = setup();
    write(&project, "apps/bo/a.ts", "bo");
    commit_all(&project);
    built(
        &project,
        json!({"protocol_version": 1, "scope": ["apps/bo"]}),
    );
    // An unscoped build now fails (16385 tracked files outside the scope) and leaves the marker.
    // `set_index` replaces the whole list, so the tracked files of the scope are named again.
    let mut tracked = vec!["a.txt".to_string(), "apps/bo/a.ts".to_string()];
    tracked.extend(numbered("apps/web", LIMIT + 1));
    set_index(&project, &tracked);
    let (kind, _, _) = refused(&build(&project, json!({"protocol_version": 1})));
    assert_eq!(kind, "invalid_input");
    write(&project, "apps/bo/a.ts", "bo changed");
    let partial = query(&project);
    assert_eq!(partial["state"], json!("partial"), "{partial}");
    assert_eq!(
        partial["changes"]["modified"],
        json!(["apps/bo/a.ts"]),
        "the changes are read inside the scope, not over 16385 files outside it: {partial}"
    );
}

/// Audit N4: a repository with no tracked file still builds without a scope; only a scope that
/// matches nothing is refused.
#[test]
fn an_unscoped_build_of_a_repository_with_no_file_still_succeeds() {
    let tmp = tempfile::tempdir().unwrap();
    let project = tmp.path().join("project");
    fs::create_dir_all(&project).unwrap();
    git(&project, &["init", "-q", "--initial-branch=main"]);
    git(&project, &["commit", "-q", "--allow-empty", "-m", "init"]);
    let empty = built(&project, json!({"protocol_version": 1}));
    assert_eq!(empty["files"], json!(0), "{empty}");
    let (kind, _, message) = refused(&build(
        &project,
        json!({"protocol_version": 1, "scope": ["apps"]}),
    ));
    assert_eq!(kind, "invalid_input", "{message}");
}

/// AC-5 (REQ-2): a request that is not a plain relative folder list, or that matches no
/// tracked file, is refused as `invalid_input` and writes no generation and no failure marker.
#[test]
fn scope_refusals_write_nothing() {
    let (_tmp, project) = setup();
    write(&project, "d0/x.txt", "x");
    commit_all(&project);
    let too_many: Vec<String> = (0..65).map(|n| format!("d{n}")).collect();
    let mut bad: Vec<Value> = vec![
        json!(["/abs"]),
        json!(["../up"]),
        json!(["a/../b"]),
        json!(["./a"]),
        json!([""]),
        json!(["a//b"]),
        json!(["a\u{1}b"]),
        json!(["back\\slash"]),
        json!(["nowhere"]),
        json!(["nowhere/deeper"]),
        json!("not-a-list"),
        json!(too_many),
    ];
    bad.push(json!(["x".repeat(513)]));
    for scope in bad {
        let out = build(&project, json!({"protocol_version": 1, "scope": scope}));
        assert_eq!(
            out.status.code(),
            Some(1),
            "accepted a bad scope {scope}: {}",
            String::from_utf8_lossy(&out.stdout)
        );
        let (kind, retryable, message) = refused(&out);
        assert_eq!(
            (kind.as_str(), retryable),
            ("invalid_input", false),
            "{scope}: {message}"
        );
        assert!(
            !index_dir(&project).join("current.json").exists()
                && !index_dir(&project).join("last-build.json").exists(),
            "{scope}: a refused request wrote a pointer or a failure marker"
        );
    }
    // The edges that are accepted: 64 entries (one matches), 512 bytes.
    let at_limit: Vec<String> = (0..64).map(|n| format!("d{n}")).collect();
    built(&project, json!({"protocol_version": 1, "scope": at_limit}));
    // 512 bytes in three components (a file name over 255 bytes is not allowed on disk).
    let long = format!(
        "{}/{}/{}",
        "a".repeat(170),
        "b".repeat(170),
        "c".repeat(170)
    );
    assert_eq!(long.len(), 512);
    write(&project, &format!("{long}/x"), "y");
    let long = long.to_string();
    commit_all(&project);
    built(&project, json!({"protocol_version": 1, "scope": [long]}));
}

/// AC-6 (REQ-5): freshness is judged inside the scope only.
#[test]
fn a_scoped_generation_ignores_changes_outside_the_scope() {
    let (_tmp, project) = setup();
    write(&project, "apps/bo/a.ts", "bo");
    write(&project, "apps/web/b.ts", "web");
    commit_all(&project);
    built(
        &project,
        json!({"protocol_version": 1, "scope": ["apps/bo"]}),
    );
    assert_eq!(query(&project)["state"], json!("fresh"));

    // Outside the scope: a modified tracked file, an untracked file, a staged new file.
    write(&project, "apps/web/b.ts", "web changed");
    write(&project, "apps/web/untracked.ts", "u");
    write(&project, "root-new.txt", "r");
    git(&project, &["add", "root-new.txt"]);
    let outside = query(&project);
    assert_eq!(outside["state"], json!("fresh"), "{outside}");

    // Inside the scope the same kinds of change are seen.
    write(&project, "apps/bo/a.ts", "bo changed");
    let modified = query(&project);
    assert_eq!(modified["state"], json!("stale"), "{modified}");
    assert_eq!(modified["changes"]["modified"], json!(["apps/bo/a.ts"]));
    write(&project, "apps/bo/a.ts", "bo");
    write(&project, "apps/bo/new.ts", "n");
    let added = query(&project);
    assert_eq!(added["state"], json!("stale"), "{added}");
    assert_eq!(added["changes"]["new_files"], json!(["apps/bo/new.ts"]));
}
