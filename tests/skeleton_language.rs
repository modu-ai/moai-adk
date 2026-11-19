//! SPEC-20261119-skeleton-language (card C-60): the text `agents-apply` writes outside its
//! managed markers when it creates a file follows the instruction language, and a file that
//! exists is never rewritten outside its block. Temporary projects only.

mod common;
use serde_json::{json, Value};
use std::fs;
use std::io::Write;
use std::path::Path;
use std::process::{Command, Stdio};
use tempfile::TempDir;

fn cli(args: &[&str], input: &Value) -> std::process::Output {
    let mut child = Command::new(crate::common::MOAI_BIN)
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
        .write_all(input.to_string().as_bytes())
        .unwrap();
    child.wait_with_output().unwrap()
}

fn result(out: &std::process::Output) -> Value {
    assert!(
        out.status.success(),
        "{}{}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    v["result"].clone()
}

/// A project whose profile has review answers and, when given, one `instruction_files` code.
fn project(instruction_files: Option<&str>) -> TempDir {
    let root = TempDir::new().unwrap();
    fs::create_dir_all(root.path().join("skills/moai/workflows")).unwrap();
    fs::create_dir_all(root.path().join(".moai-cli")).unwrap();
    let decision = |id: &str, value: &str| json!({"id": id, "topic": id, "value": value, "source": "user", "scope": "project"});
    let mut profile = json!({"schema_version": 2, "project_id": "demo", "purpose": "Build a small editor",
        "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude"],
        "interview_revision": "r1",
        "decisions": [decision("review_blocks", "a change without a test"),
                      decision("protected", "the billing module")]});
    if let Some(code) = instruction_files {
        profile["language"] = json!({ "instruction_files": code });
    }
    let mut bytes = serde_json::to_vec_pretty(&profile).unwrap();
    bytes.push(b'\n');
    fs::write(root.path().join(".moai-cli/project.yaml"), bytes).unwrap();
    root
}

fn request(language: &str, local_file: &str) -> Value {
    json!({"protocol_version": 1, "language": language,
        "workflow_reference": "skills/moai/workflows/", "local_file": local_file})
}

fn plan(root: &Path, language: &str, local_file: &str) -> Value {
    result(&cli(
        &["agents-plan", root.to_str().unwrap()],
        &request(language, local_file),
    ))
}

fn apply(root: &Path, plan: &Value) {
    result(&cli(
        &[
            "agents-apply",
            root.to_str().unwrap(),
            "--plan-digest",
            plan["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "fixture-approved",
        ],
        plan,
    ));
}

fn create(root: &Path, language: &str, local_file: &str) {
    let planned = plan(root, language, local_file);
    apply(root, &planned);
}

fn read(root: &Path, rel: &str) -> String {
    fs::read_to_string(root.join(rel)).unwrap()
}

/// The text of `file` outside its managed block: from the begin marker line to the end marker
/// line is cut out (html comment markers and `#` markers).
fn outside_markers(text: &str) -> String {
    let mut out = String::new();
    let mut inside = false;
    for line in text.split_inclusive('\n') {
        if line.contains("moai:managed:begin") {
            inside = true;
            continue;
        }
        if line.contains("moai:managed:end") {
            inside = false;
            continue;
        }
        if !inside {
            out.push_str(line);
        }
    }
    out
}

const KO_HEADING: &str = "# 검토 지침\n\n<!-- moai:managed:begin id=moai-review -->\n";
const EN_HEADING: &str = "# Review instructions\n\n<!-- moai:managed:begin id=moai-review -->\n";

#[test]
fn a_new_review_file_has_a_heading_in_the_instruction_language() {
    // (instruction_files, request language, expected start)
    let cases: [(Option<&str>, &str, &str); 7] = [
        (Some("ko"), "ko", KO_HEADING),
        (Some("en"), "en", EN_HEADING),
        (Some("ko"), "en", KO_HEADING),
        (Some("en"), "ko", EN_HEADING),
        (None, "ko", KO_HEADING),
        (None, "en", EN_HEADING),
        (Some("ja"), "ko", KO_HEADING),
    ];
    for (setting, language, expected) in cases {
        let root = project(setting);
        create(root.path(), language, "none");
        let text = read(root.path(), "REVIEW.md");
        assert!(
            text.starts_with(expected),
            "instruction_files {setting:?}, request {language}: {text}"
        );
    }
}

#[test]
fn an_existing_review_file_keeps_its_own_heading() {
    for (setting, language, own) in [
        (
            Some("ko"),
            "ko",
            "# Review instructions\n\nHand-written intro.\n\n",
        ),
        (Some("en"), "en", "# 검토 지침\n\n손으로 쓴 소개.\n\n"),
    ] {
        let root = project(setting);
        fs::write(root.path().join("REVIEW.md"), own).unwrap();
        create(root.path(), language, "none");
        let text = read(root.path(), "REVIEW.md");
        assert!(text.starts_with(own), "{text}");
        assert_eq!(
            outside_markers(&text),
            own,
            "nothing outside the block changed"
        );
        assert_eq!(text.matches("moai:managed:begin id=moai-review").count(), 1);
    }
}

#[test]
fn every_first_creation_text_follows_its_language() {
    let english_skeleton = [
        "Review instructions",
        "Personal supplement",
        "Response language",
    ];
    for (language, korean) in [("ko", true), ("en", false)] {
        let root = project(Some(language));
        create(root.path(), language, "ignored");
        for file in ["AGENTS.md", "REVIEW.md", "AGENTS.local.md", ".gitignore"] {
            let text = read(root.path(), file);
            let outside = outside_markers(&text);
            let has_hangul = outside
                .chars()
                .any(|c| ('\u{ac00}'..='\u{d7a3}').contains(&c));
            if korean {
                for word in english_skeleton {
                    assert!(
                        !outside.contains(word),
                        "{file} in ko holds `{word}` outside its markers: {outside}"
                    );
                }
            } else {
                assert!(
                    !has_hangul,
                    "{file} in en holds Korean outside its markers: {outside}"
                );
            }
        }
        // The two files that carry words: one first heading each, in the project language.
        assert!(read(root.path(), "AGENTS.local.md").starts_with(if korean {
            "# 개인 작업 보충"
        } else {
            "# Personal supplement"
        }));
        assert!(read(root.path(), "AGENTS.md").starts_with("# demo\n\n"));
        assert_eq!(
            outside_markers(&read(root.path(), ".gitignore")),
            "",
            "the ignore file has no skeleton outside its block"
        );
    }
}

#[test]
fn the_plan_reads_unchanged_right_after_the_apply() {
    for language in ["ko", "en"] {
        let root = project(Some(language));
        create(root.path(), language, "ignored");
        let again = plan(root.path(), language, "ignored");
        let rows = again["files"].as_array().unwrap();
        assert!(!rows.is_empty());
        for row in rows {
            assert_eq!(
                row["action"],
                json!("unchanged"),
                "{language}: {} is {}",
                row["path"],
                row["action"]
            );
        }
    }
}
