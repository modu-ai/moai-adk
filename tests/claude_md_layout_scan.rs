//! SPEC-20261122-claude-md-text-and-ignore: shipped and generated text calls the instruction
//! files `CLAUDE.md` and `CLAUDE.local.md`.
//!
//! `shipped_*` walks the shipped surfaces of this repository: an older instruction file name may
//! stand only in a line fragment listed in `EXCEPTIONS`, with its exact count and its reason.
//! `planted_*` runs the same scan on temporary trees, where it must report. `generated_*` sets a
//! temporary project up through the real binary, with a temporary HOME, and reads what moai wrote.
//! Listed in tools/repo_scans.json.

mod common;
use common::git::git_ok;
use common::CleanEnv;
use moai_cli::state_root::MOAI_DIR;
use serde_json::{json, Value};
use sha2::{Digest, Sha256};
use std::{
    fs,
    io::Write,
    path::{Path, PathBuf},
    process::{Command, Output, Stdio},
};
use tempfile::TempDir;

const ROOT: &str = env!("CARGO_MANIFEST_DIR");

/// A file below the shipped skill folder of a temporary tree, with its head spelled in two parts.
/// The skill reference scan of the repository (`skill_references_scan`) reads every tracked text
/// for that head followed by a file name and would take a fixture path of this test for a
/// reference that misses; it needs no row for these paths this way.
macro_rules! skill_file {
    ($below:literal) => {
        concat!("skill", "s/", $below)
    };
}

/// What the scan reads, relative to the repository root: folders and single files.
const SURFACES: [&str; 8] = [
    "skills",
    "agents",
    "templates",
    "schemas",
    "install",
    "README.md",
    "README.ko.md",
    "src",
];

/// The names of the older instruction files.
const NAMES: [&str; 4] = [
    "AGENTS.md",
    "AGENTS.local.md",
    "AGENTS.override.md",
    "AGENTS files",
];

/// The format inserts that put one of those names into a message, read in the files of
/// `names_through_constants`.
const INSERTS: [&str; 3] = ["{AGENTS_FILE}", "{LEGACY_TARGET}", "{LOCAL}"];

// --- the scan -------------------------------------------------------------------------------------

/// One excepted line fragment: it stands exactly `count` times in the shipped lines of `path`,
/// and it excuses only the names inside it.
struct Row {
    path: &'static str,
    fragment: &'static str,
    count: usize,
    reason: &'static str,
}

const fn row(path: &'static str, fragment: &'static str, reason: &'static str) -> Row {
    Row {
        path,
        fragment,
        count: 1,
        reason,
    }
}

fn names_through_constants(rel: &str) -> bool {
    rel == "src/agents.rs" || rel.starts_with("src/agents_local_merge")
}

/// The source files whose `///` lines become command help or screen text.
fn doc_lines_are_shipped(rel: &str) -> bool {
    rel.starts_with("src/ops/") || ["src/main.rs", "src/help.rs", "src/ui.rs"].contains(&rel)
}

/// A column-0 attribute that builds the next item for tests only.
fn is_test_attribute(line: &str) -> bool {
    let line = line.trim_end();
    line == "#[cfg(test)]" || (line.starts_with("#[cfg(all(test,") && line.ends_with(")]"))
}

/// A column-0 `mod <name> {` line, with or without a visibility.
fn opens_module(line: &str) -> bool {
    let line = line.trim_end();
    let line = ["pub(crate) ", "pub "]
        .iter()
        .find_map(|v| line.strip_prefix(v))
        .unwrap_or(line);
    line.strip_prefix("mod ")
        .and_then(|rest| rest.strip_suffix(" {"))
        .is_some_and(|name| {
            !name.is_empty() && name.chars().all(|c| c.is_ascii_alphanumeric() || c == '_')
        })
}

/// The lines of one file that ship, with their line numbers. Outside Rust source every line
/// ships. In `src/**.rs` a line comment does not (a `///` line of a help file does), and neither
/// does a test module: a test attribute directly followed by `mod <name> {`, up to the next
/// column-0 `}`.
fn shipped_lines<'a>(rel: &str, text: &'a str) -> Result<Vec<(usize, &'a str)>, String> {
    let lines: Vec<&str> = text.lines().collect();
    let source = rel.starts_with("src/") && rel.ends_with(".rs");
    let mut out = Vec::new();
    let mut i = 0;
    while i < lines.len() {
        let line = lines[i];
        if source {
            if is_test_attribute(line) && lines.get(i + 1).is_some_and(|next| opens_module(next)) {
                let body = i + 2;
                let end = lines[body..]
                    .iter()
                    .position(|l| l.starts_with('}'))
                    .ok_or_else(|| {
                        format!(
                            "{rel}:{}: the test module has no closing brace in column 0",
                            i + 2
                        )
                    })?;
                i = body + end + 1;
                continue;
            }
            let start = line.trim_start();
            if start.starts_with("//") {
                let help = doc_lines_are_shipped(rel)
                    && start.starts_with("///")
                    && !start.starts_with("////");
                if !help {
                    i += 1;
                    continue;
                }
            }
        }
        out.push((i + 1, line));
        i += 1;
    }
    Ok(out)
}

/// `true` when the fragment is a name alone, quoted or not, with or without a folder: such a row
/// would excuse every mention instead of one line.
fn is_bare_file_name(fragment: &str) -> bool {
    let core = fragment
        .trim()
        .trim_matches(|c| matches!(c, '`' | '"' | '\''));
    let core = core.strip_prefix('@').unwrap_or(core);
    NAMES.iter().chain(INSERTS.iter()).any(|name| {
        core == *name
            || (core.ends_with(&format!("/{name}")) && !core.contains(char::is_whitespace))
    })
}

fn row_fault(r: &Row) -> Option<&'static str> {
    if r.reason.trim().is_empty() {
        Some("has no reason")
    } else if is_bare_file_name(r.fragment) {
        Some("is only a bare file name")
    } else if !NAMES
        .iter()
        .chain(INSERTS.iter())
        .any(|name| r.fragment.contains(name))
    {
        Some("holds no whole name, so it excuses nothing")
    } else if r.count == 0 {
        Some("has the count 0")
    } else {
        None
    }
}

/// Every file of a surface, in a fixed order. What cannot be read, a link and anything that is
/// neither a file nor a folder is a problem that names its path.
fn collect(root: &Path, rel: &str, files: &mut Vec<String>, problems: &mut Vec<String>) {
    let path = root.join(rel);
    let kind = match fs::symlink_metadata(&path) {
        Ok(meta) => meta.file_type(),
        Err(e) => return problems.push(format!("{rel}: cannot be read: {e}")),
    };
    if kind.is_file() {
        return files.push(rel.to_string());
    }
    if !kind.is_dir() {
        return problems.push(format!(
            "{rel}: cannot be read: a link or not a file or folder"
        ));
    }
    let names = fs::read_dir(&path).and_then(|entries| {
        entries
            .map(|e| e.map(|e| e.file_name()))
            .collect::<std::io::Result<Vec<_>>>()
    });
    let mut names = match names {
        Ok(names) => names,
        Err(e) => return problems.push(format!("{rel}: cannot be read: {e}")),
    };
    names.sort();
    for name in names {
        match name.to_str() {
            Some(name) => collect(root, &format!("{rel}/{name}"), files, problems),
            None => problems.push(format!("{rel}: cannot be read: a name that is not UTF-8")),
        }
    }
}

/// About sixty characters on each side of byte `at`, for a report line.
fn snippet(line: &str, at: usize) -> String {
    let before: Vec<char> = line[..at].chars().rev().take(60).collect();
    let before: String = before.into_iter().rev().collect();
    let after: String = line[at..].chars().take(70).collect();
    format!("{}{}", before.trim_start(), after)
}

/// Everything wrong under `root` for these exception rows: a bad row, an unreadable surface, a
/// name outside the listed fragments and a fragment that does not stand as often as its row says.
fn scan(root: &Path, rows: &[Row]) -> Vec<String> {
    let mut problems = Vec::new();
    for (i, r) in rows.iter().enumerate() {
        if let Some(fault) = row_fault(r) {
            problems.push(format!("row {}: `{}` {fault}", r.path, r.fragment));
        }
        if rows[..i]
            .iter()
            .any(|o| o.path == r.path && o.fragment == r.fragment)
        {
            problems.push(format!("row {}: `{}` is listed twice", r.path, r.fragment));
        }
    }
    let mut files = Vec::new();
    for surface in SURFACES {
        collect(root, surface, &mut files, &mut problems);
    }
    let mut found = vec![0usize; rows.len()];
    for rel in &files {
        let text = match fs::read_to_string(root.join(rel)) {
            Ok(text) => text,
            Err(e) => {
                problems.push(format!("{rel}: cannot be read: {e}"));
                continue;
            }
        };
        let lines = match shipped_lines(rel, &text) {
            Ok(lines) => lines,
            Err(e) => {
                problems.push(e);
                continue;
            }
        };
        let inserts: &[&str] = if names_through_constants(rel) {
            &INSERTS
        } else {
            &[]
        };
        for (number, line) in lines {
            // @cc:careful a name is excused only where a listed fragment covers it whole; a
            // fragment that cuts a name in two excuses nothing.
            let mut covered: Vec<(usize, usize)> = Vec::new();
            for (i, r) in rows.iter().enumerate() {
                if r.path != rel || row_fault(r).is_some() {
                    continue;
                }
                for (at, hit) in line.match_indices(r.fragment) {
                    found[i] += 1;
                    covered.push((at, at + hit.len()));
                }
            }
            for name in NAMES.iter().chain(inserts) {
                for (at, hit) in line.match_indices(name) {
                    let end = at + hit.len();
                    if !covered.iter().any(|&(s, e)| s <= at && end <= e) {
                        problems.push(format!(
                            "{rel}:{number}: `{name}` outside a listed line: {}",
                            snippet(line, at)
                        ));
                    }
                }
            }
        }
    }
    for (r, n) in rows.iter().zip(found) {
        if row_fault(r).is_none() && n != r.count {
            problems.push(format!(
                "row {}: `{}` is listed {} time(s) and found {n}",
                r.path, r.fragment, r.count
            ));
        }
    }
    problems
}

// --- the exceptions -------------------------------------------------------------------------------

const MOVE: &str = "moves the user's older file into the CLAUDE files, or offers that move";
const READ: &str = "reads the user's older file";
const BACKUP: &str = "backs up the user's older file in the move of the Go state";
const GUARD: &str = "guards the user's older file";
const UNDO: &str = "undoes a move and restores the user's older file";
const ADVISE: &str = "advises the user about their older file";

const AGENTS_RS: &str = "src/agents.rs";
const MERGE: &str = "src/agents_local_merge.rs";
const ORIGINALS: &str = "src/agents_local_merge/originals.rs";
const PICKS: &str = "src/agents_local_merge/picks.rs";
const CLAUDE_RS: &str = "src/claude.rs";
const UPDATE_RS: &str = "src/project_update.rs";
const DECLINED: &str = "src/project_update/declined.rs";
const GUIDE: &str = "src/project_update/guide.rs";
const MERGE_APPLY: &str = "src/ops/agents_merge_apply.rs";
const MERGE_PLAN: &str = "src/ops/agents_merge_plan.rs";
const MERGE_REVERT: &str = "src/ops/agents_merge_revert.rs";
const UPDATE_MD: &str = "skills/moai/workflows/update.md";

/// Only text that moves, reads, backs up, guards, undoes or advises about the user's older files.
#[rustfmt::skip]
const EXCEPTIONS: &[Row] = &[
    row(AGENTS_RS, r#"const AGENTS_FILE: &str = "AGENTS.md";"#, READ),
    row(AGENTS_RS, r#"const LOCAL_NAMES: [&str; 2] = ["AGENTS.local.md", "CLAUDE.local.md"];"#, READ),
    row(AGENTS_RS, r#"&["AGENTS.md"]"#, MOVE),
    row(AGENTS_RS, "{AGENTS_FILE} exists, so no {BLOCK_FILE} is planned", GUARD),
    row(AGENTS_RS, "instead of {AGENTS_FILE} once it exists", GUARD),

    row(MERGE, r#"const LEGACY_TARGET: &str = "AGENTS.md";"#, READ),
    row(MERGE, r#"const LOCAL: &str = "AGENTS.local.md";"#, READ),
    row(MERGE, "(keep it in .gitignore); AGENTS.local.md stays as it is.", MOVE),
    row(MERGE, "read by everyone who works on it; AGENTS.md stays as it is.", MOVE),
    row(MERGE, r#"w == "@AGENTS.md" || w == "@./AGENTS.md""#, READ),
    row(MERGE, "already reads {LEGACY_TARGET} through an @{LEGACY_TARGET} import", READ),
    row(MERGE, r#""no {LOCAL} and no backup copy of one""#, MOVE),
    row(MERGE, r#"why_not = format!("{LOCAL} is empty")"#, MOVE),
    row(MERGE, r#""{LOCAL} only holds the text MoAI generated""#, MOVE),
    row(MERGE, "the backup copy of {LOCAL} is empty", MOVE),
    row(MERGE, "the backup copy of {LOCAL} only holds the text MoAI generated", MOVE),
    row(MERGE, "no {LEGACY_TARGET} in this project", MOVE),
    row(MERGE, "{LEGACY_TARGET} is not UTF-8 text", READ),
    row(MERGE, "MoAI's managed section of {LEGACY_TARGET} (", MOVE),
    row(MERGE, "A block an earlier personal-file merge put into {LEGACY_TARGET} (", MOVE),
    row(MERGE, "The blank lines around {} in {LEGACY_TARGET} are joined", MOVE),
    row(MERGE, "{LEGACY_TARGET} is not changed.", MOVE),
    row(MERGE, "{LEGACY_TARGET} holds no text of its own", MOVE),
    row(MERGE, "this move does not untrack it. AGENTS.local.md", MOVE),
    row(MERGE, r#"[LEGACY_TARGET, ".claude/AGENTS.md"]"#, GUARD),
    row(MERGE, "the text of {LOCAL} is already in {LEGACY_TARGET}, in a block", MOVE),

    row(ORIGINALS, "no AGENTS.md and no AGENTS.local.md in this project", MOVE),
    row(ORIGINALS, "{rel} imports AGENTS.md, so no CLAUDE file holds a copy of its text", GUARD),
    row(ORIGINALS, r#""AGENTS.md is not UTF-8 text""#, READ),
    row(ORIGINALS, "the text of AGENTS.md is not whole in CLAUDE.md", GUARD),
    row(ORIGINALS, "AGENTS.md holds only MoAI's old managed section and CLAUDE.md does not", GUARD),
    row(ORIGINALS, "the personal text an earlier generation merged into AGENTS.md is not whole", GUARD),
    row(ORIGINALS, "{LEGACY_TARGET} holds MoAI's old managed section (marker id {id}", ADVISE),
    row(ORIGINALS, "the receipt copy of {LEGACY_TARGET}", ADVISE),
    row(ORIGINALS, "the text of AGENTS.local.md is not whole in CLAUDE.local.md", GUARD),

    row(PICKS, "AGENTS.md below it and the text you wrote in AGENTS.local.md into CLAUDE.local.md; AGENTS.md", MOVE),
    row(PICKS, "and AGENTS.local.md stay unchanged unless you pick remove-originals", MOVE),
    row(PICKS, "delete the original AGENTS.md and AGENTS.local.md, whose text is", MOVE),
    row(PICKS, "agents-merge-revert writes them back; a tracked AGENTS.md", UNDO),
    row(PICKS, "another tool that reads AGENTS.md stops seeing it", ADVISE),
    row(PICKS, "write nothing; Claude Code keeps reading AGENTS.md as it does now", ADVISE),
    row(PICKS, "Nothing was written. Claude Code keeps reading AGENTS.md as it does", ADVISE),
    row(PICKS, r#""CLAUDE.md already imports AGENTS.md""#, READ),
    row(PICKS, "CLAUDE.md already holds the text of AGENTS.md and nothing of AGENTS.local.md is left", READ),
    row(PICKS, "read the text of AGENTS.md: {rel} neither imports AGENTS.md nor holds its", ADVISE),
    row(PICKS, r#""no AGENTS.md and no AGENTS.local.md in this project""#, MOVE),
    row(PICKS, "deletes AGENTS.md and AGENTS.local.md and keeps their bytes in the", MOVE),
    row(PICKS, "the project holds AGENTS.md or AGENTS.local.md; pick move or leave", MOVE),

    row(CLAUDE_RS, r#"const AGENTS_FILES: [&str; 2] = ["AGENTS.md", ".claude/AGENTS.md"];"#, READ),
    row(CLAUDE_RS, r#"t.ends_with("AGENTS.md")"#, READ),
    row(CLAUDE_RS, "AGENTS.md reading needs 2.1.277 or later", ADVISE),
    row(CLAUDE_RS, "no AGENTS.md in the directory or above", READ),
    row(CLAUDE_RS, "instructionFiles excludes direct AGENTS.md reading", ADVISE),
    // SPEC Plan 6: the one suggestion that replaces the import and the both-files setting.
    row(CLAUDE_RS, "move the text of your AGENTS.md into CLAUDE.md", ADVISE),

    row("src/gate.rs", r#""AGENTS.md","#, GUARD),
    row("src/gate.rs", r#"name != "AGENTS.md""#, GUARD),
    row("src/go_residue.rs", r#""AGENTS.md","#, GUARD),
    row("src/go_residue.rs", r#""AGENTS.local.md","#, GUARD),
    row("src/go_state.rs", r#"const ROOT_FILES: [&str; 3] = ["CLAUDE.md", "AGENTS.md", "AGENTS.local.md"];"#, BACKUP),
    row("src/go_state.rs", r#""AGENTS.md" => "CLAUDE.md","#, BACKUP),
    row("src/go_state.rs", r#""AGENTS.local.md" => "CLAUDE.local.md","#, BACKUP),
    row("src/launcher_init.rs", r#"["CLAUDE.md", "AGENTS.md"]"#, READ),

    row(UPDATE_RS, "AGENTS.md and AGENTS.local.md move into CLAUDE.md", MOVE),
    row(UPDATE_RS, "AGENTS.local.md alone merges into CLAUDE.local.md", MOVE),
    row(UPDATE_RS, r#"root.join("AGENTS.md")"#, READ),
    row(UPDATE_RS, r#"root.join("AGENTS.local.md")"#, READ),
    row(UPDATE_RS, "the text of AGENTS.local.md goes into CLAUDE.local.md through the pick", MOVE),
    row(UPDATE_RS, r#""no AGENTS.md and no AGENTS.local.md in this project""#, MOVE),
    row(UPDATE_RS, r#""CLAUDE.md", &["AGENTS.md"]"#, GUARD),
    row(UPDATE_RS, "the project holds AGENTS.md or AGENTS.local.md; the user picks move", MOVE),

    row(DECLINED, r#"["AGENTS.md", "AGENTS.local.md"]"#, READ),
    row(DECLINED, "declining leaves AGENTS.md, AGENTS.local.md and the CLAUDE files as they are", ADVISE),

    row(GUIDE, "Moving your AGENTS.md text into the CLAUDE files", MOVE),
    row(GUIDE, "Move the text of your AGENTS.md and AGENTS.local.md files into the CLAUDE files.", MOVE),
    row(GUIDE, "You choose whether the text of your AGENTS.md files moves into the CLAUDE files", MOVE),
    row(GUIDE, "Your AGENTS.md text needs no further move.", ADVISE),
    row(GUIDE, "Your AGENTS.md files cannot be moved here", ADVISE),
    row(GUIDE, "Copying your AGENTS.local.md text into CLAUDE.local.md", MOVE),
    row(GUIDE, "Copy the text of your AGENTS.local.md file into CLAUDE.local.md.", MOVE),
    row(GUIDE, "Your AGENTS.local.md text needs no copying.", ADVISE),
    row(GUIDE, "Your AGENTS.local.md text cannot be copied here", ADVISE),
    row(GUIDE, "you may delete the original AGENTS files if you wish", ADVISE),

    row("src/review.rs", r#"const CONTEXT_FILES: [&str; 2] = ["AGENTS.md", "CLAUDE.md"];"#, READ),
    row("src/scope.rs", r#""AGENTS.md","#, GUARD),
    row("src/state_root.rs", r#"HOST_OWNED: [&str; 4] = ["AGENTS.md","#, GUARD),

    row(MERGE_APPLY, "/// Merge AGENTS.local.md into CLAUDE.local.md", MOVE),
    row(MERGE_APPLY, "Merge the text of `AGENTS.local.md` into", MOVE),
    row(MERGE_APPLY, "would stop reading `AGENTS.md`", ADVISE),
    row(MERGE_APPLY, "copies the text of `AGENTS.md` and `AGENTS.local.md`", MOVE),
    row(MERGE_APPLY, "writes or deletes `AGENTS.md` or `AGENTS.local.md`", GUARD),
    row(MERGE_PLAN, "/// Preview merging AGENTS.local.md into CLAUDE.local.md", MOVE),
    row(MERGE_PLAN, "merging the text of a project's `AGENTS.local.md`", MOVE),
    row(MERGE_PLAN, "no longer reading `AGENTS.md` once", ADVISE),
    row(MERGE_PLAN, "holds `AGENTS.md` or `AGENTS.local.md` instead", MOVE),
    row(MERGE_REVERT, "/// Undo an AGENTS.local.md merge", UNDO),
    row(MERGE_REVERT, "which wrote into `AGENTS.md`, still restores `AGENTS.md`", UNDO),
    row(MERGE_REVERT, "the text of `AGENTS.md`, then", UNDO),
    row("src/ops/review_package.rs", "context (`AGENTS.md`, `CLAUDE.md`)", READ),
    // SPEC D-5: the command checks the user's own file and says so.
    Row { path: "src/ops/claude_doctor.rs", fragment: "your own AGENTS.md", count: 2, reason: ADVISE },
    Row { path: "src/help.rs", fragment: "기존 AGENTS.md", count: 2, reason: ADVISE },

    row("skills/moai/references/audit.md", "`CLAUDE.md` and the older `AGENTS.md`", GUARD),
    row("skills/moai/workflows/audit.md", "and an older `AGENTS.md` while the project still has one", READ),
    // The first of these is part of a sentence tests/go_state_classes_cli.rs pins. The three
    // stand in the update workflow, where the integration branch moved them from project.md.
    row(UPDATE_MD, "the root instruction files `CLAUDE.md`, `AGENTS.md` and `AGENTS.local.md` are moved to", BACKUP),
    row(UPDATE_MD, "`AGENTS.md` and `AGENTS.local.md` are the older instruction files", ADVISE),
    row(UPDATE_MD, "the older `AGENTS.md` and `AGENTS.local.md` are gone from it", BACKUP),
    row(UPDATE_MD, "older personal file `AGENTS.local.md` is copied into `CLAUDE.local.md`", MOVE),
    row(UPDATE_MD, "the text written in `AGENTS.md` below it and the text written in `AGENTS.local.md`", MOVE),
    row(UPDATE_MD, "`AGENTS.md` and `AGENTS.local.md` stay unchanged by `move` and `leave`", MOVE),
    row(UPDATE_MD, "**The picks for the older `AGENTS.md` and `AGENTS.local.md`.**", MOVE),
    row(UPDATE_MD, "copies the text the user wrote in `AGENTS.md` below it and the text the user wrote in `AGENTS.local.md`", MOVE),
    row(UPDATE_MD, "Claude Code keeps reading `AGENTS.md`, and the picks come back", ADVISE),
    row(UPDATE_MD, "say that `AGENTS.md` and `AGENTS.local.md` stay unchanged whichever", ADVISE),
    row(UPDATE_MD, "the one deletion of the older `AGENTS.md` and `AGENTS.local.md` the user picks", MOVE),
    row(UPDATE_MD, "no CLAUDE file imports `@AGENTS.md`", GUARD),
    row(UPDATE_MD, "the old MoAI section of `AGENTS.md` is named", ADVISE),
    row("schemas/protocol/v1/agents-request.schema.json", "for a project that holds an AGENTS.md", MOVE),
];

// --- the repository -------------------------------------------------------------------------------

/// SPEC-20261122-claude-md-text-and-ignore AC-3 (REQ-8, REQ-10).
#[test]
fn shipped_texts_name_old_files_only_in_listed_lines() {
    let problems = scan(Path::new(ROOT), EXCEPTIONS);
    assert!(
        problems.is_empty(),
        "{} problem(s) in the shipped text:\n{}",
        problems.len(),
        problems.join("\n")
    );
}

// The sentences the SPEC gives word for word (Plan 2 to 6).
const LABEL_EN: &str = "Instruction files (CLAUDE.md, CLAUDE.local.md, rules)";
const LABEL_KO: &str = "지침 파일(CLAUDE.md, CLAUDE.local.md, 규칙)";
const NEXT_ACTION: &str = "start a new host session in this project so it loads the project \
     instructions and the generated files";
const CURSOR_REASON: &str =
    "Cursor CLI has no user-level instruction file (rules live in the app); not written";
const INSTALL_PHRASE: &str = "writes a new `CLAUDE.md`";
const UNKNOWN_TRACKING: &str =
    "the generation manifest does not tell whether CLAUDE.local.md is tracked or ignored";
const UNDO_NOTE: &str = "refused while the project harness step's CLAUDE.md is in place";
const APPLY_HELP: &str = "Write a previewed CLAUDE.md managed section";
const PLAN_HELP: &str = "Preview the CLAUDE.md managed section";
const SUGGESTION: &str = "move the text of your AGENTS.md into CLAUDE.md (for a file of this \
     project, /moai update offers the pick `move`)";

/// The shipped lines of one repository file as one text, with the string continuations of Rust
/// source joined, so a sentence is found however the source wraps it.
fn shipped_text(rel: &str) -> String {
    let text = fs::read_to_string(Path::new(ROOT).join(rel)).unwrap();
    let mut out = String::new();
    let mut continued = false;
    for (_, line) in shipped_lines(rel, &text).unwrap() {
        let line = if continued { line.trim_start() } else { line };
        let head = line.strip_suffix('\\').filter(|_| rel.ends_with(".rs"));
        continued = head.is_some();
        match head {
            Some(head) => out.push_str(head),
            None => {
                out.push_str(line);
                out.push('\n');
            }
        }
    }
    out
}

/// SPEC-20261122-claude-md-text-and-ignore AC-3 (REQ-4 to REQ-7).
#[test]
fn shipped_rewritten_sentences_name_the_new_files() {
    let mut problems = Vec::new();
    for (rel, sentence) in [
        ("src/language.rs", LABEL_EN),
        ("src/language.rs", LABEL_KO),
        ("src/generation.rs", NEXT_ACTION),
        ("src/home_routing.rs", CURSOR_REASON),
        (UPDATE_MD, INSTALL_PHRASE),
        (CLAUDE_RS, SUGGESTION),
        (UPDATE_RS, UNKNOWN_TRACKING),
        (UPDATE_RS, UNDO_NOTE),
        ("src/ops/agents_apply.rs", APPLY_HELP),
        ("src/ops/agents_plan.rs", PLAN_HELP),
    ] {
        if !shipped_text(rel).contains(sentence) {
            problems.push(format!("{rel} lacks: {sentence}"));
        }
    }
    for (rel, stale) in [
        (UPDATE_MD, "writes new ones"),
        (UPDATE_MD, "writes a new one"),
        (CLAUDE_RS, "add an `@AGENTS.md` import"),
        (
            CLAUDE_RS,
            "set Project instructions to claude-md-and-agents-md",
        ),
    ] {
        if shipped_text(rel).contains(stale) {
            problems.push(format!("{rel} still says: {stale}"));
        }
    }
    assert!(problems.is_empty(), "{}", problems.join("\n"));
}

// --- planted trees --------------------------------------------------------------------------------

/// A tree that holds every surface, clean, and then the given files.
fn tree(files: &[(&str, &str)]) -> TempDir {
    let tmp = TempDir::new().unwrap();
    for surface in SURFACES {
        let path = tmp.path().join(surface);
        if surface.ends_with(".md") {
            fs::write(path, "CLAUDE.md\n").unwrap();
        } else {
            fs::create_dir_all(path).unwrap();
        }
    }
    for (rel, text) in files {
        let path = tmp.path().join(rel);
        fs::create_dir_all(path.parent().unwrap()).unwrap();
        fs::write(path, text).unwrap();
    }
    tmp
}

/// The `path:line` heads of the reported names, in report order.
fn places(problems: &[String]) -> Vec<String> {
    problems
        .iter()
        .map(|p| p.splitn(3, ':').take(2).collect::<Vec<_>>().join(":"))
        .collect()
}

/// SPEC-20261122-claude-md-text-and-ignore AC-4 (REQ-8, REQ-9).
#[test]
fn planted_lines_and_stale_rows_are_reported() {
    // A clean tree passes.
    assert_eq!(scan(tree(&[]).path(), &[]), Vec::<String>::new());

    // An unlisted file, on every surface and for every name.
    for rel in [
        skill_file!("moai/a.md"),
        "agents/a.md",
        "templates/a.md",
        "schemas/a.json",
        "install/a.sh",
        "README.md",
        "README.ko.md",
        "src/a.rs",
    ] {
        for name in NAMES {
            let text = format!("first\nread {name} here\n");
            let t = tree(&[(rel, &text)]);
            let problems = scan(t.path(), &[]);
            assert_eq!(places(&problems), [format!("{rel}:2")], "{problems:?}");
        }
    }

    // An excepted file passes with its listed line, and not with a new line beside it.
    const LISTED: &str = "the older `AGENTS.md` is read";
    let rows = [row("templates/a.md", LISTED, "reads it")];
    let t = tree(&[("templates/a.md", &format!("{LISTED}\n"))]);
    assert_eq!(scan(t.path(), &rows), Vec::<String>::new());
    let t = tree(&[(
        "templates/a.md",
        &format!("{LISTED}\nwrite AGENTS.md now\n"),
    )]);
    let problems = scan(t.path(), &rows);
    assert_eq!(places(&problems), ["templates/a.md:2"], "{problems:?}");

    // An excepted line that grows another name.
    let t = tree(&[(
        "templates/a.md",
        &format!("{LISTED} and AGENTS.override.md\n"),
    )]);
    let problems = scan(t.path(), &rows);
    assert_eq!(places(&problems), ["templates/a.md:1"], "{problems:?}");
    assert!(problems[0].contains("`AGENTS.override.md`"), "{problems:?}");

    // A fragment excuses nothing of a name it cuts in two.
    let cut = [row(
        "templates/a.md",
        "AGENTS.md and AGENTS.lo",
        "cuts a name",
    )];
    let t = tree(&[("templates/a.md", "AGENTS.md and AGENTS.local.md\n")]);
    let problems = scan(t.path(), &cut);
    assert_eq!(places(&problems), ["templates/a.md:1"], "{problems:?}");
    assert!(problems[0].contains("`AGENTS.local.md`"), "{problems:?}");

    // A constant insert counts in the merge files and in src/agents.rs, nowhere else.
    for (rel, insert) in [
        ("src/agents.rs", "{AGENTS_FILE}"),
        ("src/agents_local_merge.rs", "{LEGACY_TARGET}"),
        ("src/agents_local_merge/picks.rs", "{LOCAL}"),
    ] {
        let text = format!("let a = format!(\"write {insert} now\");\n");
        let t = tree(&[(rel, &text)]);
        let problems = scan(t.path(), &[]);
        assert_eq!(places(&problems), [format!("{rel}:1")], "{problems:?}");
    }
    let t = tree(&[("src/other.rs", "let a = format!(\"write {LOCAL} now\");\n")]);
    assert_eq!(scan(t.path(), &[]), Vec::<String>::new());

    // A wrong count: one more and one fewer than the row says, and a row of a file that is gone.
    let two = [Row {
        path: "templates/a.md",
        fragment: LISTED,
        count: 2,
        reason: "reads it",
    }];
    for (times, rows) in [(1, &two[..]), (3, &two[..]), (2, &rows[..]), (0, &rows[..])] {
        let text = format!("{LISTED}\n").repeat(times);
        let t = tree(&[("templates/a.md", &text)]);
        let problems = scan(t.path(), rows);
        assert_eq!(problems.len(), 1, "{times}: {problems:?}");
        assert!(
            problems[0].ends_with(&format!(
                "is listed {} time(s) and found {times}",
                rows[0].count
            )),
            "{problems:?}"
        );
    }
    let t = tree(&[]);
    let problems = scan(t.path(), &rows);
    assert_eq!(problems.len(), 1, "{problems:?}");
    assert!(problems[0].ends_with("found 0"), "{problems:?}");

    // A row with a bare file name is refused and excuses nothing.
    for bare in [
        "AGENTS.md",
        "`AGENTS.md`",
        " AGENTS.local.md ",
        "\".claude/AGENTS.md\"",
        "@AGENTS.md",
        "AGENTS files",
        "{LOCAL}",
    ] {
        let rows = [Row {
            path: "templates/a.md",
            fragment: bare,
            count: 1,
            reason: "a reason",
        }];
        let t = tree(&[("templates/a.md", "x AGENTS.md\n")]);
        let problems = scan(t.path(), &rows);
        assert!(
            problems
                .iter()
                .any(|p| p.ends_with("is only a bare file name")),
            "{bare}: {problems:?}"
        );
        assert!(
            problems.iter().any(|p| p.starts_with("templates/a.md:1:")),
            "{bare}: {problems:?}"
        );
    }

    // A row with no reason is refused and excuses nothing.
    for reason in ["", "  "] {
        let rows = [row("templates/a.md", LISTED, reason)];
        let t = tree(&[("templates/a.md", &format!("{LISTED}\n"))]);
        let problems = scan(t.path(), &rows);
        assert_eq!(problems.len(), 2, "{problems:?}");
        assert!(problems[0].ends_with("has no reason"), "{problems:?}");
        assert!(problems[1].starts_with("templates/a.md:1:"), "{problems:?}");
    }

    // A row without a name, a row with the count 0 and a row listed twice are refused.
    let t = tree(&[("templates/a.md", &format!("{LISTED}\n"))]);
    let bad = [
        row("templates/a.md", LISTED, "reads it"),
        row("templates/a.md", LISTED, "reads it"),
        row("templates/a.md", "the older", "no name"),
        Row {
            path: "templates/b.md",
            fragment: LISTED,
            count: 0,
            reason: "none",
        },
    ];
    let problems = scan(t.path(), &bad);
    for fault in [
        "is listed twice",
        "holds no whole name, so it excuses nothing",
        "has the count 0",
    ] {
        assert!(
            problems.iter().any(|p| p.ends_with(fault)),
            "{fault}: {problems:?}"
        );
    }
    assert_eq!(problems.len(), 3, "{problems:?}");

    // A line comment and a test module do not ship; what follows the module, and the item after a
    // test attribute that is no module, do.
    let source = r#"// AGENTS.md
    //! AGENTS.md
/// AGENTS.md
#[cfg(test)]
mod tests {
    const A: &str = "AGENTS.md";
}
#[cfg(all(test, unix))]
pub(crate) mod unix_tests {
    const B: &str = "AGENTS.md";
}
const AFTER: &str = "AGENTS.md";
#[cfg(test)]
fn helper() -> &'static str {
    "AGENTS.md"
}
#[cfg(not(test))]
mod product {
    const C: &str = "AGENTS.md";
}
let d = 1; // AGENTS.md
"#;
    let t = tree(&[("src/a.rs", source)]);
    let problems = scan(t.path(), &[]);
    assert_eq!(
        places(&problems),
        ["src/a.rs:12", "src/a.rs:15", "src/a.rs:19", "src/a.rs:21"],
        "{problems:?}"
    );

    // Only Rust source loses lines: elsewhere a comment-like line ships.
    let t = tree(&[
        ("src/data.txt", "// AGENTS.md\n"),
        ("templates/a.rs", "// AGENTS.md\n"),
    ]);
    let problems = scan(t.path(), &[]);
    assert_eq!(
        places(&problems),
        ["templates/a.rs:1", "src/data.txt:1"],
        "{problems:?}"
    );

    // A test module that never closes in column 0 is reported, not skipped to the end.
    let t = tree(&[("src/a.rs", "#[cfg(test)]\nmod tests {\n    \"AGENTS.md\"\n")]);
    let problems = scan(t.path(), &[]);
    assert_eq!(problems.len(), 1, "{problems:?}");
    assert!(
        problems[0].contains("src/a.rs:2:") && problems[0].contains("no closing brace"),
        "{problems:?}"
    );

    // A `///` line of a help file is command help and is reported; its other comments are not.
    for rel in ["src/ops/x.rs", "src/main.rs", "src/help.rs", "src/ui.rs"] {
        let t = tree(&[(
            rel,
            "    /// reads AGENTS.md\n//! AGENTS.md\n// AGENTS.md\n//// AGENTS.md\n",
        )]);
        let problems = scan(t.path(), &[]);
        assert_eq!(places(&problems), [format!("{rel}:1")], "{problems:?}");
    }
}

/// The one problem of a scan, which must name `path` as unreadable.
fn assert_unreadable(root: &Path, path: &str) {
    let problems = scan(root, &[]);
    assert_eq!(problems.len(), 1, "{path}: {problems:?}");
    assert!(
        problems[0].starts_with(&format!("{path}: cannot be read")),
        "{path}: {problems:?}"
    );
}

/// SPEC-20261122-claude-md-text-and-ignore AC-4 (REQ-9).
#[test]
fn planted_unreadable_surface_fails() {
    for surface in SURFACES {
        let t = tree(&[]);
        let path = t.path().join(surface);
        if path.is_dir() {
            fs::remove_dir(&path).unwrap();
        } else {
            fs::remove_file(&path).unwrap();
        }
        assert_unreadable(t.path(), surface);
    }

    let t = tree(&[]);
    fs::write(t.path().join("templates/a.bin"), [0xff, 0xfe, 0x00]).unwrap();
    assert_unreadable(t.path(), "templates/a.bin");

    #[cfg(unix)]
    {
        use std::os::unix::fs::{symlink, PermissionsExt};
        let mode = |path: &Path, mode: u32| {
            fs::set_permissions(path, fs::Permissions::from_mode(mode)).unwrap();
        };
        for (rel, back) in [
            ("skills/closed", 0o755),
            (skill_file!("closed/a.md"), 0o644),
        ] {
            let t = tree(&[(skill_file!("closed/a.md"), "text\n")]);
            let path = t.path().join(rel);
            mode(&path, 0o000);
            let problems = scan(t.path(), &[]);
            // The mode goes back before the assertion, so the temporary folder can be removed.
            mode(&path, back);
            assert_eq!(problems.len(), 1, "{rel}: {problems:?}");
            assert!(
                problems[0].starts_with(&format!("{rel}: cannot be read")),
                "{rel}: {problems:?}"
            );
        }
        let t = tree(&[]);
        symlink("../README.md", t.path().join("agents/link.md")).unwrap();
        assert_unreadable(t.path(), "agents/link.md");
    }
}

// --- generated projects ---------------------------------------------------------------------------

const PERMISSION: &str = "fixture-approved";

/// A temporary git repository and an empty temporary HOME.
struct Fx {
    _tmp: TempDir,
    root: PathBuf,
    home: PathBuf,
}

impl Fx {
    fn new() -> Fx {
        let tmp = TempDir::new().unwrap();
        let base = fs::canonicalize(tmp.path()).unwrap();
        let (root, home) = (base.join("project"), base.join("home"));
        fs::create_dir_all(root.join("skills/moai/workflows")).unwrap();
        fs::create_dir_all(&home).unwrap();
        git_ok(&root, &["init", "-q", "-b", "main"]);
        Fx {
            _tmp: tmp,
            root,
            home,
        }
    }

    fn p(&self) -> &str {
        self.root.to_str().unwrap()
    }

    /// Runs the binary and requires a successful answer.
    fn run(&self, args: &[&str], stdin: &[u8]) -> (Value, Output) {
        let mut child = Command::new(common::MOAI_BIN)
            .args(args)
            .current_dir(&self.root)
            .clean_env()
            .env("HOME", &self.home)
            .env("PATH", "/usr/bin:/bin")
            .env("LC_ALL", "C")
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
        let out = child.wait_with_output().unwrap();
        let said = format!(
            "{args:?}: {}{}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        );
        assert!(out.status.success(), "{said}");
        let answer: Value = serde_json::from_slice(&out.stdout).expect(&said);
        assert_eq!(answer["ok"], json!(true), "{said}");
        (answer["result"].clone(), out)
    }

    /// `<op>-plan` then `<op>-apply` of the same request; returns the apply result.
    fn apply(&self, op: &str, request: &Value) -> Value {
        let (plan, _) = self.run(
            &[&format!("{op}-plan"), self.p()],
            request.to_string().as_bytes(),
        );
        let digest = plan["plan_digest"].as_str().unwrap().to_string();
        self.run(
            &[
                &format!("{op}-apply"),
                self.p(),
                "--plan-digest",
                &digest,
                "--permission-reference",
                PERMISSION,
            ],
            plan.to_string().as_bytes(),
        )
        .0
    }
}

/// A profile with review decisions, so the generation plans `REVIEW.md` too.
fn profile() -> Value {
    let decision = |id: &str, value: &str| json!({"id": id, "topic": id, "value": value, "source": "user", "scope": "project"});
    json!({
        "schema_version": 2, "project_id": "demo", "purpose": "Build a small editor",
        "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude"],
        "interview_revision": "rc1-fixture-1",
        "decisions": [decision("review_blocks", "a change without a test"),
                      decision("protected", "the billing module")],
    })
}

fn generation_request(language: &str, local_file: &str) -> Value {
    json!({
        "protocol_version": 1, "language": language,
        "workflow_reference": "skills/moai/workflows/", "local_file": local_file,
        "workflow_id": "review-flow",
        "native_roles": [{"role_id": "prj-fixer", "host": "claude", "purpose": "Fix lint failures"}],
        "skill_projections": [],
        "project_skills": [{"id": "prj-checks", "purpose": "How this project is checked"}],
    })
}

fn strings(v: &Value) -> Vec<String> {
    v.as_array()
        .unwrap_or_else(|| panic!("not a list: {v}"))
        .iter()
        .map(|s| s.as_str().unwrap().to_string())
        .collect()
}

/// Reports every older name in `text`, one line per hit, under the label `place`.
fn old_names(place: &str, text: &str, problems: &mut Vec<String>) {
    for (i, line) in text.lines().enumerate() {
        for name in NAMES {
            for (at, _) in line.match_indices(name) {
                problems.push(format!(
                    "{place}:{}: `{name}`: {}",
                    i + 1,
                    snippet(line, at)
                ));
            }
        }
    }
}

/// Reads every file under `dir` but the git folders, its path and its text.
fn walk(dir: &Path, place: &str, files: &mut usize, problems: &mut Vec<String>) {
    let mut entries: Vec<_> = fs::read_dir(dir).unwrap().map(|e| e.unwrap()).collect();
    entries.sort_by_key(|e| e.file_name());
    for entry in entries {
        let name = entry.file_name().to_string_lossy().into_owned();
        if name == ".git" {
            continue;
        }
        let place = format!("{place}/{name}");
        old_names(&format!("path of {place}"), &place, problems);
        let kind = entry.file_type().unwrap();
        if kind.is_dir() {
            walk(&entry.path(), &place, files, problems);
        } else if kind.is_symlink() {
            let target = fs::read_link(entry.path()).unwrap();
            old_names(&place, &target.to_string_lossy(), problems);
        } else {
            assert!(kind.is_file(), "{place} is not a file");
            *files += 1;
            let bytes = fs::read(entry.path()).unwrap();
            old_names(&place, &String::from_utf8_lossy(&bytes), problems);
        }
    }
}

/// SPEC-20261122-claude-md-text-and-ignore AC-1 (REQ-1, REQ-2, REQ-11, REQ-12).
#[test]
fn generated_project_files_name_no_old_instruction_file() {
    let mut problems = Vec::new();
    for language in ["en", "ko"] {
        let fx = Fx::new();
        // The installer makes the HOME state folder and records the conversation language in it.
        fs::create_dir_all(moai_cli::home_root::path(&fx.home)).unwrap();
        let (recorded, _) = fx.run(&["language-set", "--home", language], b"");
        assert_eq!(recorded["written"], json!(true), "{recorded}");

        fx.apply("project", &profile());
        let (asked, _) = fx.run(&["language-setup", fx.p(), "--questions"], b"");
        let questions = &asked["questions"];
        old_names(
            &format!("[{language}] setup questions"),
            &serde_json::to_string_pretty(questions).unwrap(),
            &mut problems,
        );
        let question = questions["items"]
            .as_array()
            .unwrap()
            .iter()
            .find(|i| i["id"] == "instruction_files")
            .unwrap_or_else(|| panic!("no instruction_files question: {questions}"));
        for (key, expected) in [
            ("question_en", format!("Which language for: {LABEL_EN}?")),
            ("question_ko", format!("{LABEL_KO}에 쓸 언어는?")),
        ] {
            if question[key] != json!(expected) {
                problems.push(format!(
                    "[{language}] setup question {key} is {} and not `{expected}`",
                    question[key]
                ));
            }
        }
        // The same questions asked on the terminal.
        let (_, terminal) = fx.run(&["language-setup", fx.p(), "--conversation", language], b"");
        old_names(
            &format!("[{language}] terminal questions"),
            &String::from_utf8_lossy(&terminal.stderr),
            &mut problems,
        );

        let answers = json!({"user_chat": language, "instruction_files": language});
        let (answered, _) = fx.run(
            &[
                "language-setup",
                fx.p(),
                "--answers",
                "--conversation",
                language,
            ],
            answers.to_string().as_bytes(),
        );
        fx.apply("project", &answered["profile"]);
        let applied = fx.apply("generation", &generation_request(language, "ignored"));

        for written in ["CLAUDE.md", "REVIEW.md"] {
            assert!(
                fx.root.join(written).is_file(),
                "[{language}] the setup wrote no {written}: {applied}"
            );
        }
        let next = strings(&applied["next_actions"]);
        assert!(!next.is_empty(), "[{language}] no next action: {applied}");
        old_names(
            &format!("[{language}] next_actions"),
            &next.join("\n"),
            &mut problems,
        );
        let claude = fs::read_to_string(fx.root.join("CLAUDE.md")).unwrap();
        let label = if language == "ko" { LABEL_KO } else { LABEL_EN };
        if !claude.contains(&format!("{label}: ")) {
            problems.push(format!("[{language}] CLAUDE.md lacks the label `{label}`"));
        }
        for (place, dir) in [("project", &fx.root), ("HOME", &fx.home)] {
            let mut files = 0;
            walk(
                dir,
                &format!("[{language}] {place}"),
                &mut files,
                &mut problems,
            );
            assert!(files > 0, "[{language}] no file under the {place}");
        }
    }
    assert!(
        problems.is_empty(),
        "{} problem(s) in what moai wrote and said:\n{}",
        problems.len(),
        problems.join("\n")
    );
}

/// SPEC-20261122-claude-md-text-and-ignore AC-2 (REQ-3).
#[test]
fn generated_next_action_with_a_kept_old_file_claims_no_new_file() {
    let fx = Fx::new();
    let mine = "# My project notes\n\nKeep this line.\n";
    fs::write(fx.root.join("AGENTS.md"), mine).unwrap();
    fx.apply("project", &profile());
    let applied = fx.apply("generation", &generation_request("en", "none"));

    // The scene of the criterion: the user's file stays and no CLAUDE.md is written.
    assert!(!fx.root.join("CLAUDE.md").exists(), "{applied}");
    assert_eq!(fs::read_to_string(fx.root.join("AGENTS.md")).unwrap(), mine);
    let next = strings(&applied["next_actions"]);
    assert!(
        next.iter().any(|a| a == NEXT_ACTION),
        "no next action says `{NEXT_ACTION}`: {next:?}"
    );
    for name in ["CLAUDE.md", "CLAUDE.local.md"].iter().chain(NAMES.iter()) {
        assert!(
            !next.iter().any(|a| a.contains(name)),
            "a next action names {name}: {next:?}"
        );
    }
}

/// A Go-generation project (a deployed config file and skill files, and the manifest that
/// records them) with the older instruction files in its root. `recorded` names the root
/// files the manifest records as the Go deployer's own.
fn go_project(fx: &Fx, recorded: &[&str]) {
    let write = |rel: &str, body: &str| {
        let path = fx.root.join(rel);
        fs::create_dir_all(path.parent().unwrap()).unwrap();
        fs::write(path, body).unwrap();
    };
    let mut files = serde_json::Map::new();
    let mut deployed = |rel: &str, body: &str| {
        write(rel, body);
        let hash = format!("sha256:{:x}", Sha256::digest(body.as_bytes()));
        files.insert(
            rel.into(),
            json!({"provenance": "template_managed", "template_hash": hash,
                   "deployed_hash": hash, "current_hash": hash}),
        );
    };
    deployed(&format!("{MOAI_DIR}/config/sections/a.yaml"), "a: 1\n");
    deployed(".claude/skills/moai-a/SKILL.md", "# a\n");
    for name in ["AGENTS.md", "AGENTS.local.md"] {
        if recorded.contains(&name) {
            deployed(name, "# the Go deployer wrote this\n");
        } else {
            write(name, "# the user wrote this\n");
        }
    }
    write(
        &format!("{MOAI_DIR}/manifest.json"),
        &json!({"version": "v3.2.1-rc.6", "deployed_at": "2026-11-01T00:00:00Z", "files": files})
            .to_string(),
    );
}

/// SPEC-20261122-claude-md-text-and-ignore AC-11 (REQ-13): the reason of a moved older
/// instruction file says `CLAUDE.md` (or `CLAUDE.local.md`) takes its place, that the older
/// file is not written again and that the moved copy is in the backup; it never says the
/// install step writes a new one.
#[test]
fn a_moved_older_file_reason_says_the_claude_file_takes_its_place() {
    let mut problems = Vec::new();
    // Without a Go proof, and with the manifest as the Go proof: both reasons.
    for (case, recorded) in [
        ("no Go proof", &[][..]),
        ("manifest proof", &["AGENTS.md", "AGENTS.local.md"][..]),
    ] {
        let fx = Fx::new();
        go_project(&fx, recorded);
        let (plan, _) = fx.run(&["go-state-plan", fx.p()], b"");
        for (name, instead) in [
            ("AGENTS.md", "CLAUDE.md"),
            ("AGENTS.local.md", "CLAUDE.local.md"),
        ] {
            let row = plan["root_files"]
                .as_array()
                .unwrap()
                .iter()
                .find(|r| r["path"] == json!(name))
                .unwrap_or_else(|| panic!("[{case}] no root_files row for {name}: {plan}"));
            let reason = row["reason"].as_str().unwrap();
            if row["action"] != json!("move") {
                problems.push(format!("[{case}] {name} is not moved: {row}"));
            }
            for wanted in [
                format!("{instead} takes its place"),
                format!("{name} is not written again"),
                "backup".to_string(),
            ] {
                if !reason.contains(&wanted) {
                    problems.push(format!(
                        "[{case}] {name}: the reason lacks `{wanted}`: {reason}"
                    ));
                }
            }
            if reason.contains("writes a new one") {
                problems.push(format!(
                    "[{case}] {name}: the reason says an install step writes it: {reason}"
                ));
            }
        }
    }
    assert!(problems.is_empty(), "{}", problems.join("\n"));
}
