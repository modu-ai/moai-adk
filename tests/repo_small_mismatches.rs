//! SPEC-20261126-small-mismatches: three small mismatches. `moai trailers --no-spec` prints the
//! session line of a card with no SPEC (AC-1 to AC-3); every `moai` command the shipped texts
//! name is a command of this build (AC-4); a slot file that records a dead pid and that nobody
//! holds is taken by the next acquire (AC-5). Every command runs in a clean environment with a
//! temporary HOME; nothing touches the real HOME or the ledger.
#![cfg(unix)]

mod common;

use common::home::TempHome;
use serde_json::Value;
use std::collections::{BTreeMap, BTreeSet};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::Output;

const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
const SESSION: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
const SPEC: &str = "SPEC-20261126-small-mismatches";

/// A temporary HOME with one git repository in it, the folder `moai trailers` reads the worktree
/// key from.
struct World {
    home: TempHome,
    repo: PathBuf,
}

impl World {
    fn new() -> Self {
        let home = TempHome::new();
        let repo = fs::canonicalize(home.path()).unwrap().join("repo");
        fs::create_dir_all(&repo).unwrap();
        let w = World { home, repo };
        let init = w
            .home
            .command_for("git", &["init", "-q", "-b", "main"])
            .current_dir(&w.repo)
            .output()
            .expect("run git");
        assert!(init.status.success(), "git init failed");
        w
    }

    /// `moai <args>` in the repository, with a host session id.
    fn trailers(&self, args: &[&str]) -> Output {
        let mut all = vec!["trailers"];
        all.extend_from_slice(args);
        self.home
            .command(&all)
            .env(SESSION_VAR, SESSION)
            .current_dir(&self.repo)
            .output()
            .expect("run moai")
    }
}

fn envelope(out: &Output) -> Value {
    serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
        panic!(
            "stdout is not JSON ({e}): {} {}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    })
}

fn text_of(out: &Output) -> String {
    let v = envelope(out);
    assert_eq!(out.status.code(), Some(0), "{v}");
    v["result"]["text"].as_str().unwrap_or_default().to_string()
}

/// REQ-1: `--no-spec` prints the session line and no SPEC line; `--verified` adds its line.
#[test]
fn ac1_no_spec_prints_the_session_line_only() {
    let w = World::new();
    let text = text_of(&w.trailers(&["--no-spec"]));
    assert!(
        text.starts_with(&format!("Moai-Session: {SESSION}:")),
        "{text}"
    );
    assert!(!text.contains("Moai-Spec"), "{text}");
    assert_eq!(text.lines().count(), 1, "{text}");
    // The session line is the one `--spec` prints.
    let with_spec = text_of(&w.trailers(&["--spec", SPEC]));
    assert!(
        with_spec.contains(text.trim_end()),
        "--spec: {with_spec} / --no-spec: {text}"
    );
    let verified = text_of(&w.trailers(&["--no-spec", "--verified", "tests"]));
    assert!(verified.contains("Moai-Verified: tests"), "{verified}");
    assert!(!verified.contains("Moai-Spec"), "{verified}");
}

/// REQ-2, REQ-3: no flag is refused and names `--no-spec`; `--no-spec` next to `--spec` or
/// `--list` is refused and prints no trailer line.
#[test]
fn ac2_a_bare_call_names_no_spec_and_the_flag_conflicts_are_refused() {
    let w = World::new();
    let bare = w.trailers(&[]);
    assert!(!bare.status.success(), "a bare call must be refused");
    let said = format!(
        "{} {}",
        String::from_utf8_lossy(&bare.stdout),
        String::from_utf8_lossy(&bare.stderr)
    );
    assert!(said.contains("--no-spec"), "{said}");
    assert!(!said.contains("Moai-Session:"), "{said}");
    for args in [
        vec!["--no-spec", "--spec", SPEC],
        vec!["--no-spec", "--list"],
    ] {
        let out = w.trailers(&args);
        assert!(!out.status.success(), "{args:?} must be refused");
        assert!(
            !String::from_utf8_lossy(&out.stdout).contains("Moai-Session:"),
            "{args:?} printed a trailer line"
        );
    }
}

/// REQ-4: the help names the flag and what it is for.
#[test]
fn ac3_the_help_says_no_spec_is_for_a_card_with_no_spec() {
    let home = TempHome::new();
    let out = home
        .command(&["trailers", "--help"])
        .output()
        .expect("run moai");
    assert!(out.status.success());
    let help = String::from_utf8_lossy(&out.stdout).to_string();
    let flat = help.split_whitespace().collect::<Vec<_>>().join(" ");
    assert!(flat.contains("--no-spec"), "{help}");
    assert!(flat.contains("card with no SPEC"), "{help}");
}

// ---- AC-4: every command a shipped text names is a command of this build ----------------------

fn root() -> &'static Path {
    Path::new(env!("CARGO_MANIFEST_DIR"))
}

/// The shipped texts: the skill folder, the agent files and the three instruction documents.
fn shipped_texts() -> Vec<PathBuf> {
    fn walk(dir: &Path, out: &mut Vec<PathBuf>) {
        let mut entries: Vec<PathBuf> = fs::read_dir(dir)
            .unwrap_or_else(|e| panic!("{}: {e}", dir.display()))
            .map(|e| e.unwrap().path())
            .collect();
        entries.sort();
        for p in entries {
            if p.is_dir() {
                walk(&p, out);
            } else if p.extension().is_some_and(|x| x == "md") {
                out.push(p);
            }
        }
    }
    let mut out = Vec::new();
    walk(&root().join("skills/moai"), &mut out);
    walk(&root().join("agents"), &mut out);
    for rel in ["CLAUDE.md", "docs/lane-rules.md", "docs/leader-rules.md"] {
        out.push(root().join(rel));
    }
    out
}

/// A command word: lower-case letters, digits and `-`, starting with a letter.
fn is_word(w: &str) -> bool {
    w.starts_with(|c: char| c.is_ascii_lowercase())
        && w.chars()
            .all(|c| c.is_ascii_lowercase() || c.is_ascii_digit() || c == '-')
}

/// Up to two leading command words of a span that starts with `moai `.
fn words_of(span: &str) -> Vec<String> {
    let Some(rest) = span.trim().strip_prefix("moai ") else {
        return Vec::new();
    };
    rest.split_whitespace()
        .take_while(|w| is_word(w))
        .take(2)
        .map(str::to_string)
        .collect()
}

/// The `moai` commands a text names: every inline code span and every fenced line that starts
/// with `moai `. A span never crosses a line; an unclosed backtick names nothing.
fn named(text: &str) -> Vec<Vec<String>> {
    let mut out = Vec::new();
    let mut fenced = false;
    for line in text.lines() {
        if line.trim_start().starts_with("```") {
            fenced = !fenced;
        } else if fenced {
            out.push(words_of(line.trim().trim_start_matches("$ ")));
        } else {
            let pieces: Vec<&str> = line.split('`').collect();
            for (i, piece) in pieces.iter().enumerate() {
                if i % 2 == 1 && i + 1 < pieces.len() {
                    out.push(words_of(piece));
                }
            }
        }
    }
    out.retain(|w| !w.is_empty());
    out
}

/// What this build answers to `moai <word> --help`: whether it exited 0 and its text.
fn help_of(home: &TempHome, words: &[&str]) -> (bool, String) {
    let mut args: Vec<&str> = words.to_vec();
    args.push("--help");
    let out = home.command(&args).output().expect("run moai");
    (
        out.status.success(),
        String::from_utf8_lossy(&out.stdout).to_string(),
    )
}

/// The subcommand names the `Commands:` list of a help text gives, none for a plain command.
fn subcommands(help: &str) -> BTreeSet<String> {
    let mut out = BTreeSet::new();
    let mut listing = false;
    for line in help.lines() {
        if line.starts_with("Commands:") {
            listing = true;
        } else if listing {
            if !line.starts_with("  ") {
                break;
            }
            // A command line starts two spaces in; a wrapped description starts further in.
            if !line.starts_with("   ") {
                if let Some(name) = line.split_whitespace().next() {
                    out.insert(name.to_string());
                }
            }
        }
    }
    out
}

#[test]
fn ac4_every_command_the_shipped_texts_name_is_a_command_of_this_build() {
    let home = TempHome::new();
    // The first word of each named command, then its second word when the first is a group.
    let mut first: BTreeMap<String, Option<BTreeSet<String>>> = BTreeMap::new();
    let mut missing: Vec<String> = Vec::new();
    for path in shipped_texts() {
        let text = fs::read_to_string(&path).unwrap_or_else(|e| panic!("{}: {e}", path.display()));
        let rel = path.strip_prefix(root()).unwrap().display().to_string();
        for words in named(&text) {
            let head = words[0].clone();
            let entry = first.entry(head.clone()).or_insert_with(|| {
                let (ok, help) = help_of(&home, &[&head]);
                ok.then(|| subcommands(&help))
            });
            match entry {
                None => missing.push(format!(
                    "{rel}: `moai {head}` is not a command of this build"
                )),
                Some(subs) if !subs.is_empty() && words.len() > 1 && !subs.contains(&words[1]) => {
                    missing.push(format!(
                        "{rel}: `moai {head} {}` is not a subcommand of `moai {head}`",
                        words[1]
                    ));
                }
                Some(_) => {}
            }
        }
    }
    missing.sort();
    missing.dedup();
    assert!(
        missing.is_empty(),
        "shipped texts name commands this build does not have:\n{}",
        missing.join("\n")
    );
}

#[test]
fn ac4_the_scan_finds_a_command_the_build_does_not_have() {
    // The scan's own reading: a span and a fenced line name commands; prose and placeholders name none.
    let text = "Run `moai card show` and `moai nothing-here --flag`.\n\
                Not a command: `moai-adk`, moai worktree, `moai <name>`.\n\
                ```\n$ moai spec check\nmoai version\n```\n";
    let got = named(text);
    assert_eq!(
        got,
        vec![
            vec!["card".to_string(), "show".to_string()],
            vec!["nothing-here".to_string()],
            vec!["spec".to_string(), "check".to_string()],
            vec!["version".to_string()],
        ]
    );
    let home = TempHome::new();
    assert!(!help_of(&home, &["nothing-here"]).0);
    assert!(
        !help_of(&home, &["version"]).0,
        "`moai version` is not a command"
    );
    let (ok, card) = help_of(&home, &["card"]);
    assert!(ok);
    assert!(subcommands(&card).contains("show"));
    assert!(!subcommands(&card).contains("nothing-here"));
}

// ---- AC-5: a slot of a dead pid that nobody holds is taken --------------------------------------

#[test]
fn ac5_a_slot_of_a_dead_pid_that_nobody_holds_is_taken_by_the_next_acquire() {
    let home = TempHome::new();
    let slots = home.path().join("slots");
    fs::create_dir_all(&slots).unwrap();
    // A pid that was alive a moment ago and has been reaped.
    let mut gone = std::process::Command::new("true").spawn().unwrap();
    let pid = gone.id();
    gone.wait().unwrap();
    fs::write(
        slots.join("slot-0"),
        format!(
            "{{\"name\":\"gone\",\"command\":[\"sleep\",\"600\"],\"pid\":{pid},\"started_unix\":1,\"slot\":0}}\n"
        ),
    )
    .unwrap();
    let dir = slots.to_str().unwrap();
    let status = |home: &TempHome| -> Value {
        let out = home
            .command(&["heavy-slot-status", "--dir", dir])
            .output()
            .expect("run moai");
        envelope(&out)["result"].clone()
    };
    assert_eq!(status(&home)["held"], serde_json::json!([]), "free at rest");
    let out = home
        .command(&[
            "heavy-slot-acquire",
            "--dir",
            dir,
            "--name",
            "next",
            "--max",
            "1",
            "--min-free-gb",
            "0",
            "--",
            "true",
        ])
        .output()
        .expect("run moai");
    assert_eq!(
        out.status.code(),
        Some(0),
        "the next acquire waited or failed: {}",
        String::from_utf8_lossy(&out.stderr)
    );
    assert_eq!(status(&home)["held"], serde_json::json!([]), "free after");
}
