//! The repository text test of the flow rename (SPEC-20261122-next-flow REQ-3, REQ-4): no
//! shipped text names M-Flow except the one line that says it is the old name, and a text
//! that plants the old display name is refused. The scanner is a function over a root, so
//! the planted case runs it on a temporary copy instead of touching the repository.

use std::{
    fs,
    path::{Path, PathBuf},
};

/// Every line of every shipped text under `root` that carries the old display name,
/// as `relative/path:line-number: the line`, sorted for a stable message.
fn m_flow_lines(root: &Path) -> Vec<String> {
    let skills = root.join("skills");
    let mut hits = Vec::new();
    let mut stack = vec![skills.clone()];
    let mut files = Vec::new();
    while let Some(dir) = stack.pop() {
        let Ok(entries) = fs::read_dir(&dir) else {
            continue;
        };
        for entry in entries.flatten() {
            let path = entry.path();
            if path.is_dir() {
                stack.push(path);
            } else if path.extension().is_some_and(|e| e == "md") {
                files.push(path);
            }
        }
    }
    files.sort();
    for file in files {
        let Ok(text) = fs::read_to_string(&file) else {
            continue;
        };
        for (number, line) in text.lines().enumerate() {
            if line.contains("M-Flow") {
                let rel = file.strip_prefix(root).unwrap_or(&file);
                hits.push(format!("{}:{}: {}", rel.display(), number + 1, line.trim()));
            }
        }
    }
    hits
}

fn repo_root() -> PathBuf {
    PathBuf::from(env!("CARGO_MANIFEST_DIR"))
}

/// REQ-4: exactly one line of the shipped texts carries M-Flow, and it is the
/// old-name line of the one text that explains the flow.
#[test]
fn flow_name_one_old_name_line_only() {
    let hits = m_flow_lines(&repo_root());
    assert_eq!(
        hits.len(),
        1,
        "shipped texts must say NeXT-Flow; the old display name may appear once, on the old-name line. hits: {hits:#?}"
    );
    assert!(
        hits[0].contains("old name"),
        "the one allowed line must say it is the old name: {}",
        hits[0]
    );
}

/// The `/moai project` interview offers the flow by its name: the `work_flow` row of the
/// question bank says NeXT-Flow with the value next-flow (never the old value), and the
/// PROJECT workflow offers the same name when the user changes the flow.
#[test]
fn flow_name_the_interview_offers_next_flow_by_name() {
    let root = repo_root();
    let bank = fs::read_to_string(root.join("skills/moai/references/internal/interview.md"))
        .expect("the question bank is shipped");
    let row = bank
        .lines()
        .find(|l| l.starts_with("| `work_flow` |"))
        .expect("the question bank has a work_flow row");
    assert!(row.contains("next-flow (NeXT-Flow)"), "{row}");
    assert!(!row.contains("m-flow"), "{row}");
    assert!(
        bank.contains(r#"{"flow": "next-flow" | "github-flow" | "trunk-based" | "git-flow""#),
        "the pick is recorded as next-flow"
    );
    let project = fs::read_to_string(root.join("skills/moai/workflows/project.md"))
        .expect("the PROJECT workflow is shipped");
    let line = project
        .lines()
        .find(|l| l.contains("Git work flow (SPEC-20261117-mflow-project)"))
        .expect("the PROJECT workflow has a git work flow line");
    assert!(line.contains("next-flow (NeXT-Flow"), "{line}");
    assert!(!line.contains("then m-flow"), "{line}");
}

/// The name a person reads: both stored values of the local flow show as NeXT-Flow and the
/// other flows show as they are; the one list of accepted values holds both local values.
#[test]
fn flow_display_name_is_next_flow_for_both_local_values() {
    use moai_cli::profile::{flow_display_name, is_local_flow, WORK_FLOWS};
    for value in ["next-flow", "m-flow"] {
        assert!(is_local_flow(value), "{value}");
        assert_eq!(flow_display_name(value), "NeXT-Flow", "{value}");
        assert!(WORK_FLOWS.contains(&value), "{value} is accepted");
    }
    for value in ["github-flow", "trunk-based", "git-flow"] {
        assert!(!is_local_flow(value), "{value}");
        assert_eq!(flow_display_name(value), value);
    }
}

/// REQ-4's refusal bites: a planted shipped text is detected, so the scan above
/// cannot pass by scanning nothing.
#[test]
fn flow_name_a_planted_text_is_refused() {
    let tmp = tempfile::TempDir::new().unwrap();
    let dir = tmp.path().join("skills/moai/workflows");
    fs::create_dir_all(&dir).unwrap();
    fs::write(
        dir.join("planted.md"),
        "The M-Flow text that must not ship.\n",
    )
    .unwrap();
    let hits = m_flow_lines(tmp.path());
    assert_eq!(hits.len(), 1, "the scanner must find the planted text");
    assert!(hits[0].contains("planted.md"));
}
