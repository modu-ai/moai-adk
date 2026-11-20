//! SPEC-20261120-index-state-folders: the code map leaves MoAI's own state folders (`.moai/` and
//! `.moai-cli/` at the project root) out of the file count, the index and the freshness
//! judgement by default, and a scope that names a place inside one indexes it.
//!
//! The case behind it: a project whose tracked `.moai/reports` held 16,760 of 25,437 files was
//! refused at 16,384. Every case runs against a temporary git repository; the real HOME and
//! every other project are untouched. Large counts use index-only entries
//! (`git update-index --index-info`), as `tests/index_scope_cli.rs` does: the scan skips a
//! tracked path that is absent on disk, but the limit counts it.
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

/// Make the index hold exactly `paths` (all pointing at the blob of `a.txt`): a path absent on
/// disk is a tracked entry the scan skips, so a count of 40 000 costs no files.
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
    // HOME is a folder of its own, never the real one and never inside the project.
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

/// The message of a refused request, `invalid_input` and not retryable.
fn refused(out: &std::process::Output) -> String {
    assert_eq!(
        out.status.code(),
        Some(1),
        "{}",
        String::from_utf8_lossy(&out.stdout)
    );
    let v = envelope(out);
    assert_eq!(v["ok"], json!(false), "{v}");
    assert_eq!(v["error"]["kind"], json!("invalid_input"), "{v}");
    assert_eq!(v["error"]["retryable"], json!(false), "{v}");
    v["error"]["message"].as_str().unwrap().to_string()
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

fn manifest_paths(project: &Path) -> Vec<String> {
    let pointer: Value =
        serde_json::from_slice(&fs::read(project.join(".moai-cli/index/current.json")).unwrap())
            .unwrap();
    let gen = pointer["generation_id"].as_str().unwrap();
    let text = fs::read_to_string(
        project
            .join(".moai-cli/index/generations")
            .join(gen)
            .join("manifest.json"),
    )
    .unwrap();
    let v: Value = serde_json::from_str(&text).unwrap();
    v["files"]
        .as_array()
        .unwrap()
        .iter()
        .map(|f| f["path"].as_str().unwrap().to_string())
        .collect()
}

/// A project with files in both state folders and in look-alike folders that are not one.
fn mixed_project() -> (tempfile::TempDir, PathBuf) {
    let (tmp, project) = setup();
    write(&project, ".moai/specs/SPEC-x/spec.md", "spec");
    write(&project, ".moai/config/a.yaml", "a: 1\n");
    write(&project, ".moai/reports/r-1.md", "report");
    write(&project, ".moai-cli/project.yaml", "p: 1\n");
    write(&project, ".moai-notes/a.txt", "notes");
    write(&project, "docs/.moai/x", "nested");
    write(&project, "src/main.rs", "fn main() {}\n");
    commit_all(&project);
    (tmp, project)
}

/// AC-1 (REQ-1): 20 000 tracked files in each state folder's `reports` neither count nor get
/// indexed. Before the change the 40 003 tracked paths were refused at 16 384.
#[test]
fn tracked_reports_in_a_state_folder_do_not_count() {
    let (_tmp, project) = setup();
    write(&project, "src/main.rs", "fn main() {}\n");
    write(&project, "src/lib.rs", "pub fn f() {}\n");
    commit_all(&project);
    let mut paths = vec![
        "a.txt".to_string(),
        "src/main.rs".to_string(),
        "src/lib.rs".to_string(),
    ];
    paths.extend(numbered(".moai/reports", 20_000));
    paths.extend(numbered(".moai-cli/reports", 20_000));
    set_index(&project, &paths);

    let result = built(&project, json!({"protocol_version": 1}));
    assert_eq!(result["files"], json!(3), "{result}");
    assert_eq!(
        manifest_paths(&project),
        vec!["a.txt", "src/lib.rs", "src/main.rs"]
    );
}

/// AC-1 (REQ-1), the edge: 16 384 source files beside 20 000 tracked reports build, one more
/// is refused, and the refusal counts the source files only.
#[test]
fn the_limit_counts_source_files_next_to_a_state_folder_full_of_reports() {
    let (_tmp, project) = setup();
    let mut paths = vec!["a.txt".to_string()];
    paths.extend(numbered("src", LIMIT - 1));
    paths.extend(numbered(".moai/reports", 20_000));
    set_index(&project, &paths);
    let accepted = built(&project, json!({"protocol_version": 1}));
    assert_eq!(accepted["files"], json!(1), "only a.txt is on disk");

    paths.push("src/one-more".into());
    set_index(&project, &paths);
    let message = refused(&build(&project, json!({"protocol_version": 1})));
    assert!(
        message.contains("index file limit exceeded (16384)"),
        "{message}"
    );
    assert!(
        message.contains("16385 tracked files"),
        "the reports are not in the count: {message}"
    );
    assert!(
        !message.contains(".moai"),
        "no state folder is named as a place to scope: {message}"
    );
}

/// AC-2 (REQ-1, REQ-4): both state folders are left out whole, and only the first path
/// component of the project is matched.
#[test]
fn the_state_folder_is_matched_on_the_first_component_only() {
    let (_tmp, project) = mixed_project();
    let result = built(&project, json!({"protocol_version": 1}));
    assert_eq!(result["files"], json!(4), "{result}");
    assert_eq!(
        manifest_paths(&project),
        vec![".moai-notes/a.txt", "a.txt", "docs/.moai/x", "src/main.rs"]
    );
}

/// AC-3 (REQ-2): a scope that names a place inside a state folder indexes it; the rest of the
/// folder and an unscoped build still leave it out.
#[test]
fn a_scope_inside_a_state_folder_indexes_it() {
    let (_tmp, project) = mixed_project();
    built(
        &project,
        json!({"protocol_version": 1, "scope": [".moai/specs"]}),
    );
    assert_eq!(manifest_paths(&project), vec![".moai/specs/SPEC-x/spec.md"]);

    built(
        &project,
        json!({"protocol_version": 1, "scope": ["src", ".moai/specs"]}),
    );
    assert_eq!(
        manifest_paths(&project),
        vec![".moai/specs/SPEC-x/spec.md", "src/main.rs"]
    );

    built(
        &project,
        json!({"protocol_version": 1, "scope": [".moai-cli"]}),
    );
    assert_eq!(manifest_paths(&project), vec![".moai-cli/project.yaml"]);

    built(&project, json!({"protocol_version": 1}));
    assert!(
        manifest_paths(&project)
            .iter()
            .all(|p| !p.starts_with(".moai/") && !p.starts_with(".moai-cli/")),
        "{:?}",
        manifest_paths(&project)
    );
}

/// AC-4 (REQ-3): untracked files under a state folder, and a change to a tracked file there,
/// do not change what `index-query` says. 17 000 untracked files are above the 16 384 cap that
/// makes a query `unavailable` anywhere else.
#[test]
fn untracked_reports_in_a_state_folder_leave_a_query_fresh() {
    let (_tmp, project) = setup();
    write(&project, ".moai/reports/tracked.md", "one");
    commit_all(&project);
    built(&project, json!({"protocol_version": 1}));
    assert_eq!(query(&project)["state"], json!("fresh"));

    for n in 0..17_000 {
        write(&project, &format!(".moai/reports/new-{n:05}.md"), "");
    }
    write(&project, ".moai/reports/tracked.md", "two");
    let after = query(&project);
    assert_eq!(after["state"], json!("fresh"), "{after}");
}

/// A report path of about 810 bytes: 5 500 of them are over git's 4 MiB output cap while far
/// under the 16 384 file cap, so only git being told to leave the folder out can avoid it.
fn long_report_path(n: usize) -> String {
    let part = "d".repeat(200);
    format!(
        ".moai/reports/{part}1/{part}2/{part}3/{}-{n:05}",
        "f".repeat(190)
    )
}

/// AC-1 (REQ-1): the tracked list of a folder of long report paths never reaches git's output
/// cap, because git is told to leave the state folders out.
#[test]
fn a_tracked_folder_of_long_report_paths_never_reaches_gits_output_cap() {
    let (_tmp, project) = setup();
    let mut paths = vec!["a.txt".to_string()];
    paths.extend((0..5_500).map(long_report_path));
    set_index(&project, &paths);
    let result = built(&project, json!({"protocol_version": 1}));
    assert_eq!(result["files"], json!(1), "{result}");
}

/// AC-4 (REQ-3): the untracked list of the same folder never reaches git's output cap either.
#[test]
fn an_untracked_folder_of_long_report_paths_never_reaches_gits_output_cap() {
    let (_tmp, project) = setup();
    built(&project, json!({"protocol_version": 1}));
    for n in 0..5_500 {
        write(&project, &long_report_path(n), "");
    }
    let after = query(&project);
    assert_eq!(after["state"], json!("fresh"), "{after}");
}

/// AC-4 (REQ-3, REQ-4): the same count outside a state folder still makes a query unavailable,
/// also inside a folder that only looks like one (`docs/.moai`).
#[test]
fn untracked_files_outside_a_state_folder_still_count() {
    let (_tmp, project) = setup();
    built(&project, json!({"protocol_version": 1}));
    for n in 0..=LIMIT {
        write(&project, &format!("docs/.moai/new-{n:05}"), "");
    }
    let overflow = query(&project);
    assert_eq!(overflow["state"], json!("unavailable"), "{overflow}");
    assert!(overflow["reason"]
        .as_str()
        .unwrap()
        .contains("untracked file limit exceeded"));
}
