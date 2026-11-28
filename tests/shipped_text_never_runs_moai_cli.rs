//! SPEC-20261127-installed-binary-name-moai AC-10 (REQ-10, REQ-11, D-6): the shipped texts name
//! `moai` as the command to run and never `moai-cli`. The scan reads the skill and agent texts,
//! the templates, `install/README.md`, `README*.md`, `src/help.rs` and the help goldens. A line
//! may still spell `moai-cli` where it names the earlier file as history, as an asset name or as
//! a path to report, and each such place is a row of the allow list with its reason. A row names
//! the phrase it allows: the phrase is taken out of the line and any `moai-cli` left over is a
//! finding, so a command added to an allowed line is not hidden by the row. The state folder name
//! is taken out first (it holds the same letters). The checks sit in one test so the verify
//! command of the SPEC counts one passing test.

use std::fs;
use std::path::{Path, PathBuf};

const ALLOW: &str = "tests/fixtures/shipped_text_moai_cli/allow.tsv";
const WORD: &str = "moai-cli";

/// One allowed phrase of one file, with the reason it is allowed.
struct Row {
    path: String,
    phrase: String,
    reason: String,
}

fn parse(text: &str) -> Result<Vec<Row>, String> {
    let mut rows = Vec::new();
    for (n, line) in text.lines().enumerate() {
        if line.trim().is_empty() || line.starts_with('#') {
            continue;
        }
        let cols: Vec<&str> = line.split('\t').collect();
        if cols.len() != 3 {
            return Err(format!(
                "allow list line {}: want path, phrase, reason",
                n + 1
            ));
        }
        let (path, phrase, reason) = (cols[0].trim(), cols[1], cols[2].trim());
        if !phrase.contains(WORD) {
            return Err(format!(
                "allow list line {}: the phrase must contain the word it allows",
                n + 1
            ));
        }
        if reason.chars().count() < 20 {
            return Err(format!(
                "allow list line {}: the reason is missing or too short to say why",
                n + 1
            ));
        }
        rows.push(Row {
            path: path.to_string(),
            phrase: phrase.to_string(),
            reason: reason.to_string(),
        });
    }
    Ok(rows)
}

/// Whether `line` of `path` still spells the word after each allowed phrase of that file and the
/// state folder name are taken out. `used` marks the rows that matched.
fn leftover(path: &str, line: &str, rows: &[Row], used: &mut [bool]) -> bool {
    let mut rest = line.to_string();
    for (i, row) in rows.iter().enumerate() {
        if row.path == path && rest.contains(&row.phrase) {
            rest = rest.replace(&row.phrase, "");
            used[i] = true;
        }
    }
    rest.replace(moai_cli::HOME_STATE_DIR, "").contains(WORD)
}

fn root() -> PathBuf {
    PathBuf::from(env!("CARGO_MANIFEST_DIR"))
}

/// The files the scan reads, as repository-relative paths with `/`.
fn scanned() -> Vec<String> {
    fn walk(dir: &Path, base: &Path, out: &mut Vec<String>) {
        let Ok(entries) = fs::read_dir(dir) else {
            return;
        };
        for entry in entries.flatten() {
            let path = entry.path();
            if path.is_dir() {
                walk(&path, base, out);
            } else if let Ok(rel) = path.strip_prefix(base) {
                out.push(rel.to_string_lossy().replace('\\', "/"));
            }
        }
    }
    let base = root();
    let mut out = Vec::new();
    for dir in ["skills", "agents", "templates"] {
        walk(&base.join(dir), &base, &mut out);
    }
    for file in ["install/README.md", "src/help.rs"] {
        out.push(file.to_string());
    }
    for dir in [".", "tests/fixtures/help"] {
        if let Ok(entries) = fs::read_dir(base.join(dir)) {
            for entry in entries.flatten() {
                let name = entry.file_name().to_string_lossy().into_owned();
                let wanted = if dir == "." {
                    name.starts_with("README") && name.ends_with(".md")
                } else {
                    name.ends_with(".txt")
                };
                if wanted && entry.path().is_file() {
                    out.push(if dir == "." {
                        name
                    } else {
                        format!("{dir}/{name}")
                    });
                }
            }
        }
    }
    out.sort();
    out
}

#[test]
fn shipped_text_never_runs_moai_cli() {
    // The matcher on made-up lines: a command is a finding, an allowed phrase is not, a command
    // beside an allowed phrase still is, and the state folder name alone is not.
    let fake = parse("a.md\tthe earlier `moai-cli` link\tnames the earlier file as history, never as a command\n")
        .unwrap();
    let mut used = vec![false];
    assert!(leftover(
        "a.md",
        "run `moai-cli update` now",
        &fake,
        &mut used
    ));
    assert!(!used[0]);
    assert!(!leftover(
        "a.md",
        "the earlier `moai-cli` link stays",
        &fake,
        &mut used
    ));
    assert!(used[0]);
    assert!(leftover(
        "a.md",
        "run `moai-cli` and the earlier `moai-cli` link stays",
        &fake,
        &mut used
    ));
    assert!(leftover(
        "b.md",
        "the earlier `moai-cli` link stays",
        &fake,
        &mut used
    ));
    let state_line = format!("keeps {}/path", moai_cli::HOME_STATE_DIR);
    assert!(!leftover("a.md", &state_line, &fake, &mut used));
    assert!(parse("a.md\tthe `moai-cli` link\tshort\n").is_err());
    assert!(
        parse("a.md\tno word here\tnames the earlier file as history, never as a command\n")
            .is_err()
    );
    assert!(parse("a.md\tonly two columns\n").is_err());

    // The repository.
    let rows = parse(&fs::read_to_string(root().join(ALLOW)).expect("the allow list")).unwrap();
    let mut used = vec![false; rows.len()];
    let mut findings = Vec::new();
    for rel in scanned() {
        let Ok(text) = fs::read_to_string(root().join(&rel)) else {
            continue;
        };
        for (n, line) in text.lines().enumerate() {
            if leftover(&rel, line, &rows, &mut used) {
                findings.push(format!("{rel}:{}: {}", n + 1, line.trim()));
            }
        }
    }
    assert!(
        findings.is_empty(),
        "a shipped text names `{WORD}` outside the allow list ({ALLOW}):\n{}",
        findings.join("\n")
    );
    let stale: Vec<String> = rows
        .iter()
        .zip(&used)
        .filter(|(_, hit)| !**hit)
        .map(|(row, _)| format!("{}: {}", row.path, row.phrase))
        .collect();
    assert!(
        stale.is_empty(),
        "allow list rows that match no line (remove them):\n{}",
        stale.join("\n")
    );
    assert!(rows.iter().all(|r| !r.reason.is_empty()));

    // D-6: when `moai --version` does not answer with the Rust signature, the skill reports the
    // block and stops; it gives no other command to run.
    let skill = fs::read_to_string(root().join("skills/moai/SKILL.md")).unwrap();
    let line = skill
        .lines()
        .find(|l| l.contains("`moai --version`"))
        .expect("the skill's version check line");
    assert!(
        line.contains("blocked") && line.contains("stop"),
        "the skill reports the block and stops: {line}"
    );
    assert!(
        !line.contains("instead") && !line.contains("use `moai-cli`"),
        "the skill names no other command to run: {line}"
    );
}
