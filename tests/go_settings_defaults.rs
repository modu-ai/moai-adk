//! #2817: `language-setup` offers the Go generation's language and name settings, read from
//! `<project>/.moai-backup/config/sections/`, as defaults — nothing is copied, the backup is
//! only read, and broken or hostile input simply offers nothing. Temporary projects only.

#![cfg(unix)]

mod common;
use serde_json::{json, Value};
use std::{
    fs,
    io::Write as _,
    path::{Path, PathBuf},
    process::{Command, Stdio},
};
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;
/// `go_settings::FILE_LIMIT` as a literal, so a changed bound fails here.
const LIMIT: usize = 64 * 1024;

const GO_LANGUAGE: &str = "# Language Configuration\nlanguage:\n  # Conversation language code (ISO 639-1)\n  conversation_language: ko\n  conversation_language_name: ko\n  agent_prompt_language: en\n  git_commit_messages: \"ja\"\n  code_comments: 'zh'\n  documentation: ko\n  error_messages: en\n";
const GO_PROJECT: &str = "# Project Metadata Configuration\nproject:\n  name: \"go-demo\"\n  description: \"\"\n  initialized: true\n  template_version: \"v3.2.0-rc.23\"\n  mode: personal\n";

/// A migrated project: the Rust `.moai/` with its marker and profile (no `language` yet),
/// and the Go folder kept as `.moai-backup`.
fn migrated(language: Option<&str>, project: Option<&str>) -> (TempDir, PathBuf) {
    let d = TempDir::new().unwrap();
    let root = fs::canonicalize(d.path()).unwrap();
    fs::create_dir_all(root.join(".moai")).unwrap();
    fs::write(
        root.join(".moai/moai-cli.json"),
        r#"{"schema_version":1,"generation":"moai-cli","created_by":"moai-cli test"}"#,
    )
    .unwrap();
    let profile = json!({"schema_version":2,"project_id":"demo","purpose":"Build a small editor","tools":["Rust"],"development_methods":["TDD"],"hosts":["claude","codex"]});
    fs::write(
        root.join(".moai/project.yaml"),
        serde_json::to_vec_pretty(&profile).unwrap(),
    )
    .unwrap();
    let sections = root.join(".moai-backup/config/sections");
    fs::create_dir_all(&sections).unwrap();
    if let Some(text) = language {
        fs::write(sections.join("language.yaml"), text).unwrap();
    }
    if let Some(text) = project {
        fs::write(sections.join("project.yaml"), text).unwrap();
    }
    (d, root)
}

fn setup(root: &Path, args: &[&str], stdin: &str) -> (Vec<u8>, String) {
    let mut child = Command::new(EXE)
        .arg("language-setup")
        .arg(root)
        .args(args)
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
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    (
        out.stdout,
        String::from_utf8_lossy(&out.stderr).into_owned(),
    )
}

fn result(bytes: &[u8]) -> Value {
    let v: Value = serde_json::from_slice(bytes).unwrap();
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn suggested(questions: &Value) -> Vec<(String, String)> {
    questions["questions"]["items"]
        .as_array()
        .unwrap()
        .iter()
        .map(|i| {
            (
                i["id"].as_str().unwrap().to_string(),
                i["suggested"].as_str().unwrap().to_string(),
            )
        })
        .collect()
}

fn tree(root: &Path) -> Vec<(String, Vec<u8>)> {
    let mut out = Vec::new();
    let mut stack = vec![root.to_path_buf()];
    while let Some(d) = stack.pop() {
        for e in fs::read_dir(&d).unwrap().flatten() {
            let p = e.path();
            let rel = p.strip_prefix(root).unwrap().display().to_string();
            let m = fs::symlink_metadata(&p).unwrap();
            if m.is_dir() {
                out.push((format!("{rel}/"), Vec::new()));
                stack.push(p);
            } else if m.file_type().is_symlink() {
                out.push((
                    format!("{rel}@"),
                    fs::read_link(&p)
                        .unwrap()
                        .to_string_lossy()
                        .into_owned()
                        .into_bytes(),
                ));
            } else {
                out.push((rel, fs::read(&p).unwrap()));
            }
        }
    }
    out.sort();
    out
}

/// AC-1 and AC-3: the Go values become the defaults (`suggested`), named in `go_backup`, the
/// terminal step says where they came from, the answers flow keeps them, and the backup is
/// byte-identical afterwards.
#[test]
fn the_go_backup_values_are_offered_as_defaults_and_never_written() {
    let (_d, root) = migrated(Some(GO_LANGUAGE), Some(GO_PROJECT));
    let backup = root.join(".moai-backup");
    let before = tree(&backup);

    let (stdout, _) = setup(&root, &["--questions"], "");
    let q = result(&stdout);
    assert_eq!(
        suggested(&q),
        vec![
            ("user_chat".into(), "ko".into()),
            ("reports".into(), "ko".into()),
            ("issue_pr_records".into(), "ko".into()),
            ("commit_messages".into(), "ja".into()),
            ("agent_messages".into(), "en".into()),
            ("instruction_files".into(), "en".into()),
            ("skill_instructions".into(), "en".into()),
            ("code_comments".into(), "zh".into()),
        ]
    );
    assert_eq!(q["questions"]["conversation_language"], json!("ko"));
    let go = &q["go_backup"];
    assert_eq!(go["source"], json!(".moai-backup/config/sections"), "{q}");
    assert_eq!(go["project_id"], json!("go-demo"));
    assert_eq!(go["copied"], json!(false));
    assert!(go["not_moved"]
        .as_array()
        .unwrap()
        .contains(&json!("language.documentation")));

    // No answers: the profile the step returns carries the Go values.
    let (stdout, _) = setup(&root, &["--answers"], "{}");
    let r = result(&stdout);
    assert_eq!(r["language"]["commit_messages"], json!("ja"), "{r}");
    assert_eq!(r["language"]["user_chat"], json!("ko"));
    assert_eq!(r["profile_written"], json!(false));

    // An answer still wins over the Go value; an explicit --conversation over the Go one.
    let (stdout, _) = setup(
        &root,
        &["--answers", "--conversation", "en"],
        r#"{"code_comments":"en"}"#,
    );
    let r = result(&stdout);
    assert_eq!(r["language"]["code_comments"], json!("en"), "{r}");
    assert_eq!(r["language"]["user_chat"], json!("en"));
    assert_eq!(r["language"]["commit_messages"], json!("ja"));

    // The terminal step says where its defaults came from (stderr) and keeps them on Enter.
    let (stdout, stderr) = setup(&root, &[], &"\n".repeat(8));
    assert!(stderr.contains(".moai-backup/config/sections"), "{stderr}");
    assert_eq!(result(&stdout)["language"]["code_comments"], json!("zh"));

    assert_eq!(tree(&backup), before, "the backup is only read");
}

/// D4: without a backup the output is byte-for-byte what it was — no `go_backup` key.
#[test]
fn without_a_backup_the_output_bytes_do_not_change() {
    let (_d, root) = migrated(Some(GO_LANGUAGE), Some(GO_PROJECT));
    let (with, _) = setup(&root, &["--answers"], "{}");
    assert!(result(&with).get("go_backup").is_some());
    fs::rename(root.join(".moai-backup"), root.join("moved-away")).unwrap();
    let (q, _) = setup(&root, &["--questions"], "");
    let (a, _) = setup(&root, &["--answers"], "{}");
    for bytes in [&q, &a] {
        let text = String::from_utf8_lossy(bytes);
        assert!(!text.contains("go_backup"), "{text}");
    }
    assert_eq!(
        result(&q)["questions"]["conversation_language"],
        json!("en")
    );
}

/// AC-2: a broken file, a link, an oversized file (64 KiB kept, one byte more refused), an
/// unknown language code or a name that is not a portable id offers no default and never
/// fails the step.
#[test]
fn bad_or_hostile_input_offers_nothing_and_never_fails() {
    // Unknown language code and an unportable name: those values only are skipped.
    let (_d, root) = migrated(
        Some("language:\n  conversation_language: es\n  code_comments: ko\n"),
        Some("project:\n  name: \"my project\"\n"),
    );
    let (stdout, _) = setup(&root, &["--questions"], "");
    let q = result(&stdout);
    assert_eq!(q["go_backup"]["conversation_language"], json!(null), "{q}");
    assert_eq!(q["go_backup"]["items"], json!({"code_comments": "ko"}));
    assert_eq!(q["go_backup"]["project_id"], json!(null));

    // Broken YAML: the whole file gives nothing.
    let (_d, root) = migrated(
        Some("language:\n  code_comments:\n    - ko\n"),
        Some("project: [\n"),
    );
    let q = result(&setup(&root, &["--questions"], "").0);
    assert_eq!(q["go_backup"]["items"], json!({}), "{q}");
    assert_eq!(q["go_backup"]["project_id"], json!(null));

    // The size bound: exactly LIMIT bytes is read, LIMIT + 1 is not.
    for (extra, read) in [(0usize, true), (1, false)] {
        let body = "language:\n  code_comments: ko\n";
        let pad = LIMIT - body.len() - 1 + extra;
        let text = format!("{body}#{}", "x".repeat(pad));
        assert_eq!(text.len(), LIMIT + extra);
        let (_d, root) = migrated(Some(&text), None);
        let q = result(&setup(&root, &["--questions"], "").0);
        let expect = if read {
            json!({"code_comments": "ko"})
        } else {
            json!({})
        };
        assert_eq!(q["go_backup"]["items"], expect, "{} bytes", text.len());
    }

    // A linked file, and a linked sections folder: nothing is followed.
    let (_d, root) = migrated(None, None);
    let elsewhere = root.join("elsewhere.yaml");
    fs::write(&elsewhere, "language:\n  code_comments: ko\n").unwrap();
    let sections = root.join(".moai-backup/config/sections");
    std::os::unix::fs::symlink(&elsewhere, sections.join("language.yaml")).unwrap();
    let q = result(&setup(&root, &["--questions"], "").0);
    assert_eq!(q["go_backup"]["items"], json!({}), "{q}");

    let (_d, root) = migrated(None, None);
    let real = root.join("real-sections");
    fs::create_dir(&real).unwrap();
    fs::write(
        real.join("language.yaml"),
        "language:\n  code_comments: ko\n",
    )
    .unwrap();
    fs::remove_dir(root.join(".moai-backup/config/sections")).unwrap();
    std::os::unix::fs::symlink(&real, root.join(".moai-backup/config/sections")).unwrap();
    let (stdout, _) = setup(&root, &["--questions"], "");
    assert!(
        result(&stdout).get("go_backup").is_none(),
        "a linked folder is no backup"
    );
}

/// #2820 audit B1: a value the profile already stores wins over the Go backup's — in the
/// suggested answers, the conversation language and the profile `--answers '{}'` returns;
/// only the items the profile leaves empty take the Go value.
#[test]
fn the_profile_value_wins_over_the_go_backup() {
    let (_d, root) = migrated(
        Some("language:\n  conversation_language: ko\n  code_comments: zh\n  git_commit_messages: ja\n"),
        None,
    );
    let profile = json!({"schema_version":2,"project_id":"demo","purpose":"Build a small editor","tools":["Rust"],"development_methods":["TDD"],"hosts":["claude","codex"],"language":{"user_chat":"en","code_comments":"en"}});
    fs::write(
        root.join(".moai/project.yaml"),
        serde_json::to_vec_pretty(&profile).unwrap(),
    )
    .unwrap();

    let q = result(&setup(&root, &["--questions"], "").0);
    assert_eq!(
        q["questions"]["conversation_language"],
        json!("en"),
        "the profile's user_chat, not the Go conversation language: {q}"
    );
    let s: std::collections::BTreeMap<_, _> = suggested(&q).into_iter().collect();
    assert_eq!(s["user_chat"], "en");
    assert_eq!(s["code_comments"], "en", "profile over Go");
    assert_eq!(s["commit_messages"], "ja", "empty in the profile: Go value");
    assert_eq!(
        s["reports"], "en",
        "conversation default follows the profile"
    );

    let r = result(&setup(&root, &["--answers"], "{}").0);
    assert_eq!(r["language"]["user_chat"], json!("en"), "{r}");
    assert_eq!(r["language"]["code_comments"], json!("en"));
    assert_eq!(r["language"]["commit_messages"], json!("ja"));
}
