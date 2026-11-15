//! #3009 S3a: `check-tree`, the read-only warning about what the tracked MoAI folders hold.
//! Temporary git repositories only: the real HOME and real projects are never read. Findings
//! name a path and a count or a size, never the matched text, so each case also checks that the
//! planted user name, UUID and id do not appear in the answer.
mod common;
use common::cli::{tree_state, write_file};
use common::git::git_ok;
use serde_json::{json, Value};
use std::path::Path;
use std::process::{Command, Stdio};

const BIN: &str = crate::common::MOAI_BIN;
const LIMIT: usize = 1024 * 1024;
const UUID: &str = "123e4567-e89b-12d3-a456-426614174000";

fn repo() -> tempfile::TempDir {
    let d = tempfile::tempdir().unwrap();
    git_ok(d.path(), &["init", "-q"]);
    d
}

fn track(dir: &Path) {
    git_ok(dir, &["add", "-A"]);
}

/// The whole answer: `(ok, result or error, raw stdout)`.
fn run(project: &Path) -> (bool, Value, String) {
    let out = Command::new(BIN)
        .args(["check-tree"])
        .arg(project)
        .env_remove("MOAI_HOME")
        .stdin(Stdio::null())
        .output()
        .unwrap();
    let text = String::from_utf8_lossy(&out.stdout).to_string();
    let v: Value = serde_json::from_str(&text).unwrap_or_else(|_| {
        panic!("{text} / {}", String::from_utf8_lossy(&out.stderr));
    });
    let ok = v["ok"] == json!(true);
    let body = if ok {
        v["result"].clone()
    } else {
        v["error"].clone()
    };
    (ok, body, text)
}

fn kinds(result: &Value) -> Vec<(String, String)> {
    result["findings"]
        .as_array()
        .unwrap()
        .iter()
        .map(|f| {
            (
                f["path"].as_str().unwrap().to_string(),
                f["kind"].as_str().unwrap().to_string(),
            )
        })
        .collect()
}

fn pair(path: &str, kind: &str) -> (String, String) {
    (path.to_string(), kind.to_string())
}

#[test]
fn a_clean_tree_answers_clean_with_the_counts_and_writes_nothing() {
    let d = repo();
    write_file(d.path(), ".moai/specs/a.md", b"# a spec\n");
    write_file(d.path(), "src/x.rs", b"fn main() {}\n");
    track(d.path());
    let before = tree_state(d.path());
    let (ok, r, _) = run(d.path());
    assert!(ok, "{r}");
    assert_eq!(r["status"], json!("clean"), "{r}");
    assert_eq!(r["stage"], json!("warning"));
    assert_eq!(r["writes"], json!(false));
    assert_eq!(
        (r["tracked"].clone(), r["in_scope"].clone()),
        (json!(2), json!(1))
    );
    assert_eq!(r["findings"], json!([]));
    assert_eq!(r["omitted"], json!(0));
    assert_eq!(r["size_limit_bytes"], json!(LIMIT));
    assert_eq!(tree_state(d.path()), before, "the check changed the tree");
}

#[test]
fn a_file_of_exactly_1_mib_passes_and_one_byte_more_is_flagged() {
    let d = repo();
    write_file(d.path(), ".moai/at.txt", &vec![b'a'; LIMIT]);
    write_file(d.path(), ".moai/over.txt", &vec![b'a'; LIMIT + 1]);
    write_file(d.path(), ".moai/under.txt", &vec![b'a'; LIMIT - 1]);
    track(d.path());
    let (ok, r, _) = run(d.path());
    assert!(ok, "{r}");
    assert_eq!(kinds(&r), vec![pair(".moai/over.txt", "large_file")], "{r}");
    assert_eq!(r["findings"][0]["bytes"], json!(LIMIT + 1));
    assert_eq!(r["counts"], json!({"large_file": 1}));
}

#[test]
fn an_absolute_path_is_counted_and_the_user_name_is_never_printed() {
    let d = repo();
    write_file(
        d.path(),
        ".moai/notes.md",
        b"see /Users/zebulon/work and \"/home/quincy/x\"\n/Users/\n",
    );
    write_file(d.path(), ".moai/clean.md", b"lib/Users/x is relative\n");
    track(d.path());
    let (ok, r, raw) = run(d.path());
    assert!(ok, "{r}");
    assert_eq!(
        kinds(&r),
        vec![pair(".moai/notes.md", "absolute_path")],
        "{r}"
    );
    assert_eq!(r["findings"][0]["hits"], json!(2));
    for secret in ["zebulon", "quincy"] {
        assert!(
            !raw.contains(secret),
            "the answer repeated `{secret}`: {raw}"
        );
    }
}

#[test]
fn a_session_uuid_in_a_path_or_in_a_session_line_is_flagged_without_the_uuid() {
    let d = repo();
    write_file(d.path(), &format!(".moai/sessions/{UUID}.md"), b"summary\n");
    write_file(
        d.path(),
        ".moai/log.md",
        format!("Session {UUID} ended\nother {UUID} no word\n").as_bytes(),
    );
    write_file(
        d.path(),
        ".moai/other.md",
        format!("a commit {UUID} with no such word\n").as_bytes(),
    );
    track(d.path());
    let (ok, r, raw) = run(d.path());
    assert!(ok, "{r}");
    assert_eq!(
        kinds(&r),
        vec![
            pair(".moai/log.md", "session_id"),
            pair(&format!(".moai/sessions/{UUID}.md"), "session_id"),
        ],
        "{r}"
    );
    assert_eq!(r["findings"][0]["hits"], json!(1));
    assert!(
        !raw.replace(&format!(".moai/sessions/{UUID}.md"), "")
            .contains(UUID),
        "the answer repeated the UUID outside the path: {raw}"
    );
}

#[test]
fn credential_names_are_flagged_and_example_files_are_not() {
    let d = repo();
    for name in [".env", "prod.env", "server.pem", "api.key", ".env.local"] {
        write_file(d.path(), &format!(".moai/{name}"), b"PLACEHOLDER\n");
    }
    for name in [".env.example", "readme.md"] {
        write_file(d.path(), &format!(".moai/{name}"), b"PLACEHOLDER\n");
    }
    track(d.path());
    let (ok, r, _) = run(d.path());
    assert!(ok, "{r}");
    let flagged: Vec<String> = kinds(&r).into_iter().map(|(p, _)| p).collect();
    assert_eq!(
        flagged,
        [
            ".moai/.env",
            ".moai/.env.local",
            ".moai/api.key",
            ".moai/prod.env",
            ".moai/server.pem"
        ],
        "{r}"
    );
    assert_eq!(r["counts"], json!({"secret_name": 5}));
}

#[test]
fn only_tracked_files_under_the_two_moai_folders_are_looked_at() {
    let d = repo();
    write_file(d.path(), "src/a.rs", b"/Users/outside/x\n");
    write_file(d.path(), "app/.moai/a.md", b"/Users/nested/x\n");
    write_file(d.path(), ".moai/tracked.md", b"/Users/tracked/x\n");
    write_file(d.path(), ".moai-cli/state/x.txt", b"/Users/state/x\n");
    track(d.path());
    // Written after `git add`: untracked, so not looked at.
    write_file(d.path(), ".moai/untracked.md", b"/Users/untracked/x\n");
    write_file(d.path(), ".moai-cli/state/y.txt", b"/Users/ignored/x\n");
    let (ok, r, _) = run(d.path());
    assert!(ok, "{r}");
    assert_eq!(
        kinds(&r),
        vec![
            pair(".moai-cli/state/x.txt", "absolute_path"),
            pair(".moai/tracked.md", "absolute_path"),
        ],
        "{r}"
    );
    assert_eq!(r["in_scope"], json!(2));
}

#[test]
fn a_tracked_file_that_was_deleted_is_counted_as_missing_not_an_error() {
    let d = repo();
    write_file(d.path(), ".moai/gone.md", b"/Users/a/x\n");
    write_file(d.path(), ".moai/here.md", b"fine\n");
    track(d.path());
    std::fs::remove_file(d.path().join(".moai/gone.md")).unwrap();
    let (ok, r, _) = run(d.path());
    assert!(ok, "{r}");
    assert_eq!(
        (r["missing"].clone(), r["in_scope"].clone()),
        (json!(1), json!(2))
    );
    assert_eq!(r["findings"], json!([]));
}

#[cfg(unix)]
#[test]
fn a_link_and_a_binary_file_are_not_read_as_text() {
    let d = repo();
    write_file(d.path(), "src/target.txt", b"/Users/linked/x\n");
    std::fs::create_dir_all(d.path().join(".moai")).unwrap();
    std::os::unix::fs::symlink("../src/target.txt", d.path().join(".moai/link.txt")).unwrap();
    let mut bin = vec![0u8, 1, 2];
    bin.extend_from_slice(b"/Users/binary/x");
    write_file(d.path(), ".moai/blob.bin", &bin);
    track(d.path());
    let (ok, r, _) = run(d.path());
    assert!(ok, "{r}");
    assert_eq!(r["findings"], json!([]), "{r}");
    assert_eq!(r["not_read"], json!(1), "the link: {r}");
}

#[test]
fn a_project_id_a_person_chose_is_a_warning_and_a_random_one_is_not() {
    let cases: [(&str, &str, Option<&str>); 4] = [
        (
            "{\"project_id\":\"demo\"}",
            "a name",
            Some("project_id_not_random"),
        ),
        (
            "{\"project_id\":\"abcdefghijklmnopqrstuvwxyz\"}",
            "random",
            None,
        ),
        (
            "{\"project_id\":\"abcdefghijklmnopqrstuvwxy\"}",
            "25 chars",
            Some("project_id_not_random"),
        ),
        (
            "project_id: demo\n",
            "yaml text",
            Some("project_id_unusable"),
        ),
    ];
    for (text, label, want) in cases {
        let d = repo();
        write_file(d.path(), ".moai-cli/project.yaml", text.as_bytes());
        let (ok, r, raw) = run(d.path());
        assert!(ok, "{label}: {r}");
        let got: Vec<String> = kinds(&r).into_iter().map(|(_, k)| k).collect();
        assert_eq!(
            got,
            want.map(String::from).into_iter().collect::<Vec<_>>(),
            "{label}: {r}"
        );
        assert!(!raw.contains("demo"), "{label}: the id was printed: {raw}");
    }
}

#[test]
fn at_most_200_findings_are_listed_and_the_counts_cover_all() {
    let d = repo();
    for i in 0..201 {
        write_file(d.path(), &format!(".moai/k/{i:03}.pem"), b"PLACEHOLDER\n");
    }
    track(d.path());
    let (ok, r, _) = run(d.path());
    assert!(ok, "{r}");
    assert_eq!(r["findings"].as_array().unwrap().len(), 200);
    assert_eq!(r["omitted"], json!(1));
    assert_eq!(r["counts"], json!({"secret_name": 201}));
    assert_eq!(r["status"], json!("warnings"));
}

#[test]
fn a_subfolder_is_checked_on_its_own() {
    let d = repo();
    write_file(d.path(), ".moai/top.md", b"/Users/top/x\n");
    write_file(d.path(), "sub/.moai/in.md", b"fine\n");
    track(d.path());
    let (ok, r, _) = run(&d.path().join("sub"));
    assert!(ok, "{r}");
    assert_eq!(
        (r["status"].clone(), r["in_scope"].clone()),
        (json!("clean"), json!(1))
    );
}

#[test]
fn a_folder_that_is_not_a_git_work_tree_and_a_missing_folder_are_refused() {
    let plain = tempfile::tempdir().unwrap();
    let (ok, e, _) = run(plain.path());
    assert!(!ok);
    assert_eq!(e["kind"], json!("invalid_input"), "{e}");
    let (ok, e, _) = run(&plain.path().join("nowhere"));
    assert!(!ok);
    assert_eq!(e["kind"], json!("not_found"), "{e}");
}
