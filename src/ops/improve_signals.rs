//! `improve-signals` (SPEC-20261121-self-improvement-loop, slice S1): the observed problems of a
//! project, grouped into classes with a count. Reads the card ledger and the project's
//! `improve/notes.jsonl`; writes nothing and proposes nothing. The logic is in this file with
//! its unit tests, as `claim_timeout_check.rs` keeps its tests.

use moai_cli::contract::PROJECT_PATH_UNAVAILABLE;
use moai_cli::forge::policy::parse_utc;
use moai_cli::ledger::{self, LedgerError};
use moai_cli::redact;
use rusqlite::Connection;
use serde_json::{json, Value};
use std::collections::{BTreeMap, BTreeSet};
use std::io::Read;
use std::path::{Path, PathBuf};

/// Read the observed problems of a project as counted classes
///
/// Reads only facts that were observed: the audit-FAIL notes and the `LANE-BLOCKED … kind=<kind>`
/// notes of the project's card ledger, the cards titled `flaky: <test>` with their recurrence
/// notes, and the lines of the project's `improve/notes.jsonl`. Each fact is a signal; signals
/// of one problem form a class with a count, the first and last time and at most five examples.
/// A class seen once is `observed_once`, seen twice or more it is `candidate`. An agent's own
/// claim (a READY, DONE, CANDIDATE or MERGE-READY note) is never a signal. Host memory of any
/// tool is not read. Writes nothing, proposes nothing, creates no file or folder; a project
/// without a ledger or notes file reports each as `absent` and exits 0.
#[derive(clap::Args)]
pub struct Args {
    /// The project folder.
    #[arg(default_value = ".")]
    project: PathBuf,
    /// The HOME that holds the card ledger (default: $HOME).
    #[arg(long)]
    home: Option<PathBuf>,
}

pub const OPERATION: &str = "improve.signals";

/// A class seen this many times is a `candidate` (SPEC D9).
const THRESHOLD: usize = 2;
/// Example signals shown per class.
const EXAMPLES: usize = 5;
/// The characters of a text that leave the command.
const EXCERPT_CHARS: usize = 120;
/// The newest ledger notes read.
const MAX_NOTES: usize = 20_000;
/// The bytes of one note body read.
const BODY_BYTES: usize = 8192;
/// The bytes of the notes file read.
const FILE_BYTES: u64 = 1024 * 1024;
/// File names a cited path may end in.
const EXTENSIONS: [&str; 9] = ["rs", "md", "py", "json", "yml", "yaml", "toml", "sh", "txt"];
/// The kinds a line of the notes file may carry.
const NOTE_KINDS: [&str; 3] = ["correction", "failure", "false_claim"];

/// One observed fact.
#[derive(Debug, Clone, PartialEq)]
struct Signal {
    class: String,
    at: i64,
    reference: String,
    text: String,
}

/// One card note as read: the body is cut to [`BODY_BYTES`] bytes.
#[derive(Debug, Clone, PartialEq)]
struct NoteRow {
    card_id: i64,
    note_id: i64,
    kind: String,
    at: i64,
    body: String,
}

pub fn run(args: Args) -> Result<Value, String> {
    let home = ledger::command::resolve_home(args.home)?;
    let id = ledger::command::project_id_of(&args.project)?;
    let root =
        std::fs::canonicalize(&args.project).map_err(|_| PROJECT_PATH_UNAVAILABLE.to_string())?;

    let notes_path = moai_cli::state_root::path(&root, "improve/notes.jsonl");
    let improve = read_improve_notes(&notes_path)?;

    let mut signals: Vec<Signal> = Vec::new();
    let mut no_rule = 0usize;
    let mut truncated = false;
    let ledger_state = match ledger::open_read_only(&home, &id) {
        Ok(l) => {
            let conn = l.conn();
            let (notes, cut) = read_notes(conn, MAX_NOTES, BODY_BYTES)?;
            truncated = cut;
            let flaky = read_flaky_cards(conn)?;
            let (found, skipped) = signals_of(&notes, &flaky);
            signals.extend(found);
            no_rule += skipped;
            for (card, (test, at)) in &flaky {
                signals.push(Signal {
                    class: flake_class(test),
                    at: *at,
                    reference: format!("C-{card}"),
                    text: excerpt(&format!("flaky: {test}")),
                });
            }
            "ready"
        }
        Err(LedgerError::Absent) => "absent",
        Err(e) => return Err(e.to_string()),
    };
    let mut bad_line = 0usize;
    let notes_file = match improve {
        Some(file) => {
            signals.extend(file.signals);
            bad_line = file.bad_lines;
            truncated |= file.truncated;
            "present"
        }
        None => "absent",
    };

    Ok(json!({
        "action": "signals", "project_id": id, "ledger": ledger_state,
        "notes_file": notes_file, "host_memory": "off", "threshold": THRESHOLD,
        "truncated": truncated, "classes": classes_of(signals),
        "skipped": {"no_rule": no_rule, "bad_line": bad_line}, "writes": false,
    }))
}

/// The newest `limit` notes in the order they were written, each body cut to `body_bytes`
/// bytes, and whether older notes were left out.
fn read_notes(
    conn: &Connection,
    limit: usize,
    body_bytes: usize,
) -> Result<(Vec<NoteRow>, bool), String> {
    let db = |e: rusqlite::Error| LedgerError::from(e).to_string();
    let mut stmt = conn
        .prepare(
            "SELECT card_id, note_id, kind, created_at, substr(CAST(body AS BLOB), 1, ?2) \
               FROM card_note ORDER BY note_id DESC LIMIT ?1",
        )
        .map_err(db)?;
    let rows = stmt
        .query_map(
            rusqlite::params![limit as i64 + 1, body_bytes as i64],
            |r| {
                let bytes: Vec<u8> = r.get(4)?;
                Ok(NoteRow {
                    card_id: r.get(0)?,
                    note_id: r.get(1)?,
                    kind: r.get(2)?,
                    at: r.get(3)?,
                    body: String::from_utf8_lossy(&bytes).into_owned(),
                })
            },
        )
        .map_err(db)?;
    let mut notes: Vec<NoteRow> = Vec::new();
    for row in rows {
        notes.push(row.map_err(db)?);
    }
    let truncated = notes.len() > limit;
    notes.truncate(limit);
    notes.reverse();
    Ok((notes, truncated))
}

/// The cards titled `flaky: <test>`: card id to (test name, creation time).
fn read_flaky_cards(conn: &Connection) -> Result<BTreeMap<i64, (String, i64)>, String> {
    let db = |e: rusqlite::Error| LedgerError::from(e).to_string();
    let mut stmt = conn
        .prepare(
            "SELECT card_id, substr(title, 1, 300), created_at FROM card \
              WHERE substr(title, 1, 7) = 'flaky: ' ORDER BY card_id",
        )
        .map_err(db)?;
    let rows = stmt
        .query_map([], |r| {
            Ok((
                r.get::<_, i64>(0)?,
                r.get::<_, String>(1)?,
                r.get::<_, i64>(2)?,
            ))
        })
        .map_err(db)?;
    let mut out = BTreeMap::new();
    for row in rows {
        let (card, title, at) = row.map_err(db)?;
        if let Some(test) = title["flaky: ".len()..].split_whitespace().next() {
            out.insert(card, (test.to_string(), at));
        }
    }
    Ok(out)
}

/// The signals of the card notes and how many notes matched no rule. A note on a flaky card of
/// the kind `observation` or `report` is a recurrence of that flake; every other note is read by
/// its first line.
fn signals_of(notes: &[NoteRow], flaky: &BTreeMap<i64, (String, i64)>) -> (Vec<Signal>, usize) {
    let mut out = Vec::new();
    let mut no_rule = 0;
    for n in notes {
        let reference = format!("C-{}#{}", n.card_id, n.note_id);
        if let Some((test, _)) = flaky.get(&n.card_id) {
            if n.kind == "observation" || n.kind == "report" {
                out.push(Signal {
                    class: flake_class(test),
                    at: n.at,
                    reference,
                    text: excerpt(first_line(&n.body)),
                });
                continue;
            }
        }
        let first = first_line(&n.body);
        if is_audit_fail(first) {
            let mut paths = cited_paths(&n.body);
            if paths.is_empty() {
                paths.insert("uncited".to_string());
            }
            for path in paths {
                out.push(Signal {
                    class: class_key("audit_fail", &path),
                    at: n.at,
                    reference: reference.clone(),
                    text: excerpt(first),
                });
            }
        } else if let Some(class) = block_class(first) {
            out.push(Signal {
                class,
                at: n.at,
                reference,
                text: excerpt(first),
            });
        } else {
            no_rule += 1;
        }
    }
    (out, no_rule)
}

fn first_line(body: &str) -> &str {
    body.lines().next().unwrap_or("").trim()
}

/// `AUDIT … verdict: FAIL`, the line a leader writes on a card (a pull-request comment has the
/// same two markers).
fn is_audit_fail(first: &str) -> bool {
    first.starts_with("AUDIT ") && first.contains("verdict: FAIL")
}

/// `block:<kind>` of a `LANE-BLOCKED … kind=<kind>` line; `block:unknown` without a kind.
fn block_class(first: &str) -> Option<String> {
    if !first.starts_with("LANE-BLOCKED") {
        return None;
    }
    let kind = first
        .split_whitespace()
        .find_map(|w| w.strip_prefix("kind="))
        .map(|k| {
            k.chars()
                .take_while(|c| c.is_ascii_lowercase() || c.is_ascii_digit() || *c == '-')
                .collect::<String>()
        })
        .filter(|k| !k.is_empty())
        .unwrap_or_else(|| "unknown".to_string());
    Some(class_key("block", &kind))
}

fn flake_class(test: &str) -> String {
    class_key("flake", test)
}

/// `<kind>:<name>`; a name that has a secret shape is masked, so a key never carries one.
fn class_key(kind: &str, name: &str) -> String {
    if redact::contains_secret_shape(name) {
        format!("{kind}:[redacted]")
    } else {
        format!("{kind}:{name}")
    }
}

/// The relative file paths a text names, each once: a word of letters, digits and `_ . / -`
/// ending in a known extension (`src/a.rs:12` is `src/a.rs`).
fn cited_paths(text: &str) -> BTreeSet<String> {
    let mut out = BTreeSet::new();
    for word in
        text.split(|c: char| !(c.is_ascii_alphanumeric() || matches!(c, '_' | '.' | '/' | '-')))
    {
        let word = word
            .strip_prefix("./")
            .unwrap_or(word)
            .trim_end_matches('.');
        if word.is_empty()
            || word.len() > 200
            || word.starts_with('/')
            || word.starts_with('-')
            || word.contains("..")
        {
            continue;
        }
        let Some((stem, ext)) = word.rsplit_once('.') else {
            continue;
        };
        if !stem.is_empty() && !stem.ends_with('/') && EXTENSIONS.contains(&ext) {
            out.insert(word.to_string());
        }
    }
    out
}

/// A text as it may leave the command: masked first, then cut, so a cut never leaves half of a
/// secret.
fn excerpt(text: &str) -> String {
    let (masked, _) = redact::redact_outgoing(&redact::redact_secret_shapes(text));
    masked.chars().take(EXCERPT_CHARS).collect()
}

/// What the notes file held.
struct NotesFile {
    signals: Vec<Signal>,
    bad_lines: usize,
    truncated: bool,
}

/// The lines of the notes file, or `None` when there is no file. A symbolic link or anything
/// but a regular file is refused before it is opened (a named pipe would block).
fn read_improve_notes(path: &Path) -> Result<Option<NotesFile>, String> {
    let meta = match std::fs::symlink_metadata(path) {
        Ok(m) => m,
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => return Ok(None),
        Err(_) => return Err("improve-signals refused: the notes file cannot be inspected".into()),
    };
    if meta.file_type().is_symlink() {
        return Err(
            "improve-signals refused: the notes file is a symbolic link; it is not read".into(),
        );
    }
    if !meta.is_file() {
        return Err("improve-signals refused: the notes file is not a regular file".into());
    }
    let mut bytes = Vec::new();
    std::fs::File::open(path)
        .and_then(|f| f.take(FILE_BYTES + 1).read_to_end(&mut bytes))
        .map_err(|_| "improve-signals refused: the notes file cannot be read".to_string())?;
    let truncated = bytes.len() as u64 > FILE_BYTES;
    if truncated {
        let end = bytes[..FILE_BYTES as usize]
            .iter()
            .rposition(|b| *b == b'\n')
            .map_or(0, |i| i + 1);
        bytes.truncate(end);
    }
    let text = String::from_utf8_lossy(&bytes);
    let mut file = NotesFile {
        signals: Vec::new(),
        bad_lines: 0,
        truncated,
    };
    for (i, line) in text.lines().enumerate() {
        if line.trim().is_empty() {
            continue;
        }
        match parse_improve_line(line, i + 1) {
            Some(signal) => file.signals.push(signal),
            None => file.bad_lines += 1,
        }
    }
    Ok(Some(file))
}

/// One line `{"v":1,"at":"…Z","kind":"correction|failure|false_claim","key":"…","text":"…"}`.
fn parse_improve_line(line: &str, number: usize) -> Option<Signal> {
    let v: Value = serde_json::from_str(line).ok()?;
    if v.get("v") != Some(&json!(1)) {
        return None;
    }
    let kind = v.get("kind")?.as_str().filter(|k| NOTE_KINDS.contains(k))?;
    let key = v.get("key")?.as_str().filter(|k| {
        let n = k.chars().count();
        (1..=80).contains(&n) && !k.chars().any(char::is_control)
    })?;
    let at = parse_utc(v.get("at")?.as_str()?)?;
    let text = v.get("text").and_then(Value::as_str).unwrap_or("");
    Some(Signal {
        class: class_key(kind, key),
        at,
        reference: format!("notes.jsonl:{number}"),
        text: excerpt(first_line(text)),
    })
}

/// The classes, most seen first.
fn classes_of(signals: Vec<Signal>) -> Vec<Value> {
    let mut by_class: BTreeMap<String, Vec<Signal>> = BTreeMap::new();
    for s in signals {
        by_class.entry(s.class.clone()).or_default().push(s);
    }
    let mut out: Vec<(usize, String, Value)> = Vec::new();
    for (key, mut list) in by_class {
        list.sort_by(|a, b| (a.at, &a.reference).cmp(&(b.at, &b.reference)));
        let count = list.len();
        let kind = key.split(':').next().unwrap_or("").to_string();
        let examples: Vec<Value> = list[count.saturating_sub(EXAMPLES)..]
            .iter()
            .map(|s| json!({"ref": s.reference, "at": s.at, "text": s.text}))
            .collect();
        let value = json!({
            "key": key, "kind": kind, "count": count, "status": status(count),
            "first_seen": list[0].at, "last_seen": list[count - 1].at, "signals": examples,
        });
        out.push((count, key, value));
    }
    out.sort_by(|a, b| b.0.cmp(&a.0).then_with(|| a.1.cmp(&b.1)));
    out.into_iter().map(|(_, _, v)| v).collect()
}

fn status(count: usize) -> &'static str {
    if count >= THRESHOLD {
        "candidate"
    } else {
        "observed_once"
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn db(notes: &[(&str, &str)]) -> Connection {
        let c = Connection::open_in_memory().unwrap();
        c.execute_batch(
            "CREATE TABLE card_note (card_id INTEGER, note_id INTEGER PRIMARY KEY AUTOINCREMENT, \
             kind TEXT, created_at INTEGER, body TEXT);",
        )
        .unwrap();
        for (kind, body) in notes {
            c.execute(
                "INSERT INTO card_note (card_id, kind, created_at, body) VALUES (1, ?1, 5, ?2)",
                rusqlite::params![kind, body],
            )
            .unwrap();
        }
        c
    }

    #[test]
    fn improve_signals_status_flips_at_two() {
        assert_eq!(status(1), "observed_once");
        assert_eq!(status(2), "candidate");
        assert_eq!(status(3), "candidate");
    }

    #[test]
    fn improve_signals_reads_the_newest_n_notes_and_says_when_it_cut() {
        let notes: Vec<(&str, &str)> = (0..6).map(|_| ("comment", "x")).collect();
        let conn = db(&notes);
        let (rows, cut) = read_notes(&conn, 6, BODY_BYTES).unwrap();
        assert_eq!((rows.len(), cut), (6, false), "N notes are read whole");
        let (rows, cut) = read_notes(&conn, 5, BODY_BYTES).unwrap();
        assert_eq!((rows.len(), cut), (5, true), "N+1 notes are cut to N");
        assert_eq!(rows[0].note_id, 2, "the oldest one is the one left out");
        assert_eq!(rows[4].note_id, 6);
        let (rows, cut) = read_notes(&conn, 7, BODY_BYTES).unwrap();
        assert_eq!((rows.len(), cut), (6, false));
    }

    #[test]
    fn improve_signals_cuts_a_body_at_the_byte_limit() {
        let long = "a".repeat(BODY_BYTES + 1);
        let exact = "b".repeat(BODY_BYTES);
        let conn = db(&[("comment", long.as_str()), ("comment", exact.as_str())]);
        let (rows, _) = read_notes(&conn, 10, BODY_BYTES).unwrap();
        assert_eq!(rows[0].body.len(), BODY_BYTES, "one byte over is cut");
        assert_eq!(
            rows[1].body.len(),
            BODY_BYTES,
            "the limit itself is kept whole"
        );
        // A cut inside a multi-byte letter still reads as text.
        let hangul = "한".repeat(BODY_BYTES);
        let conn = db(&[("comment", hangul.as_str())]);
        let (rows, _) = read_notes(&conn, 10, BODY_BYTES).unwrap();
        assert!(rows[0].body.len() <= BODY_BYTES + 3);
    }

    #[test]
    fn improve_signals_cited_paths_are_relative_and_once() {
        let p = cited_paths("src/a.rs:12 and src/a.rs:99, docs/x.md. also ./tests/t.py (x.toml)");
        let got: Vec<&str> = p.iter().map(String::as_str).collect();
        assert_eq!(got, vec!["docs/x.md", "src/a.rs", "tests/t.py", "x.toml"]);
        assert!(cited_paths("/etc/passwd.txt ../up.rs -rf.rs a..b.rs e.g. v1.2 a.rsx").is_empty());
        assert!(
            cited_paths(".rs a/.rs").is_empty(),
            "an extension alone is not a path"
        );
    }

    #[test]
    fn improve_signals_reads_the_audit_and_blocked_lines_only() {
        assert!(is_audit_fail("AUDIT C-1 head x — verdict: FAIL"));
        assert!(is_audit_fail(
            "AUDIT #12 (key) head x base y — verdict: FAIL"
        ));
        assert!(!is_audit_fail("AUDIT C-1 head x — verdict: PASS"));
        assert!(!is_audit_fail("verdict: FAIL"));
        assert_eq!(
            block_class("LANE-BLOCKED a C-1 r · kind=tip-red owner=l").as_deref(),
            Some("block:tip-red")
        );
        assert_eq!(
            block_class("LANE-BLOCKED a C-1 r").as_deref(),
            Some("block:unknown")
        );
        assert_eq!(
            block_class("LANE-BLOCKED a C-1 kind=").as_deref(),
            Some("block:unknown")
        );
        assert_eq!(block_class("LANE-READY a C-1"), None);
    }

    #[test]
    fn improve_signals_masks_then_cuts_and_never_keys_a_secret() {
        let text = format!("{} password=hunter2 tail", "z".repeat(EXCERPT_CHARS - 5));
        let shown = excerpt(&text);
        assert!(!shown.contains("hunter2"), "{shown}");
        assert!(shown.chars().count() <= EXCERPT_CHARS);
        assert_eq!(
            excerpt(&"q".repeat(EXCERPT_CHARS)).chars().count(),
            EXCERPT_CHARS
        );
        assert_eq!(
            excerpt(&"q".repeat(EXCERPT_CHARS + 1)).chars().count(),
            EXCERPT_CHARS
        );
        assert_eq!(
            class_key("correction", "token=abc"),
            "correction:[redacted]"
        );
        assert_eq!(class_key("correction", "naming"), "correction:naming");
    }

    #[test]
    fn improve_signals_a_line_needs_version_kind_key_and_time() {
        let ok =
            r#"{"v":1,"at":"2026-11-01T00:00:00Z","kind":"failure","key":"cargo","text":"red"}"#;
        assert_eq!(parse_improve_line(ok, 3).unwrap().class, "failure:cargo");
        assert_eq!(
            parse_improve_line(ok, 3).unwrap().reference,
            "notes.jsonl:3"
        );
        for bad in [
            r#"{"v":2,"at":"2026-11-01T00:00:00Z","kind":"failure","key":"k"}"#,
            r#"{"v":1,"at":"2026-11-01T00:00:00Z","kind":"praise","key":"k"}"#,
            r#"{"v":1,"at":"2026-11-01","kind":"failure","key":"k"}"#,
            r#"{"v":1,"at":"2026-11-01T00:00:00Z","kind":"failure","key":""}"#,
            "[]",
            "nope",
        ] {
            assert!(parse_improve_line(bad, 1).is_none(), "{bad}");
        }
        let key80 = "k".repeat(80);
        let line = |k: &str| {
            format!(r#"{{"v":1,"at":"2026-11-01T00:00:00Z","kind":"failure","key":"{k}"}}"#)
        };
        assert!(
            parse_improve_line(&line(&key80), 1).is_some(),
            "80 characters pass"
        );
        assert!(
            parse_improve_line(&line(&format!("{key80}k")), 1).is_none(),
            "81 are refused"
        );
    }
}
