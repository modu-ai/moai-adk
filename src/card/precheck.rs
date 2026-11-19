//! `moai card precheck` (card C-96, workflow W3): the read-only check the leader runs before
//! readying or dispatching a card. Two dispatch mistakes named the card's birth: cards already
//! done on the integration tip were handed to lanes anyway (C-88 in #3151, C-92 in #3187), and
//! criteria contradicted decisions the ledger already held (C-87 AC 2, C-63 N1).
//!
//! The answer is one JSON report. Its `merged_mentions` lists the commits of the integration
//! ref whose message names the card — `C-<n>`, `card/<n>` or `card <n>` as a word in the
//! subject or body, or a `Moai-Spec:` trailer naming the SPEC the ledger links to the card —
//! with the pull-request number of `Merge pull request #N from …` subjects. Its `spec` names
//! the linked SPEC and every SPEC folder under `.moai/specs/` whose text names the card. Its
//! `decision_notes` lists every `kind = decision` note of the ledger — on any card, this one
//! included — that shares a path with the card's write scope or a distinctive token of its
//! criteria, so a collision is read before a lane starts.
//!
//! Nothing is written and no finding refuses the run: a missing ref, an absent ledger or an
//! unreadable one is an advisory inside an `ok` answer, because the check advises and the
//! leader decides. Only a malformed card name, a card the ledger does not have (and a `#<n>`
//! name, which only the ledger can resolve), or an unreadable environment refuses, the way the
//! other read-only card actions do. The git half needs no network: a merged pull
//! request is its merge commit on the integration ref.

use super::input::{self, CardRef};
use super::spec_link;
use crate::identity::scrub_git_env;
use crate::ledger;
use crate::ledger::command::{project_id_of, resolve_home};
use rusqlite::{params, OptionalExtension};
use serde_json::{json, Value};
use std::path::Path;

/// The integration ref's commits are scanned at most this far by default; the bound is named in
/// the report whenever it cuts the scan short (`truncated`).
pub const DEFAULT_MAX_COMMITS: usize = 500;

/// The largest `--max-commits` honoured: `git log -n` takes a 31-bit count, and a bound that
/// large is already "everything". A larger value is read as this one, never as an overflow.
pub const MAX_COMMITS_CEILING: usize = 1_000_000;

/// A refusal shaped like the other card actions' error frames: `kind` and `message`.
pub struct Refusal {
    pub kind: &'static str,
    pub message: String,
}

type Result<T> = std::result::Result<T, Refusal>;

fn refused<E: std::fmt::Display>(what: &str, error: E) -> Refusal {
    Refusal {
        kind: "invalid_input",
        message: format!("card precheck refused: {what}: {error}"),
    }
}

/// A ledger that cannot be opened. The open error already starts with `ledger blocked`, so the
/// prefix is added only when it is missing.
fn ledger_blocked<E: std::fmt::Display>(error: E) -> Refusal {
    let text = error.to_string();
    let message = if text.starts_with("ledger blocked") {
        text
    } else {
        format!("ledger blocked: {text}")
    };
    Refusal {
        kind: "blocked",
        message,
    }
}

fn not_found(message: String) -> Refusal {
    Refusal {
        kind: "not_found",
        message,
    }
}

/// The cards that came from source number `n` — usually none or one; more than one is the
/// caller's ambiguity to refuse.
fn cards_of_source(conn: &rusqlite::Connection, n: i64) -> Result<Vec<i64>> {
    let mut stmt = conn
        .prepare("SELECT card_id FROM card WHERE source_number = ?1")
        .map_err(|e| refused("ledger read", e))?;
    let rows = stmt
        .query_map(params![n], |r| r.get(0))
        .map_err(|e| refused("ledger read", e))?
        .collect::<std::result::Result<Vec<i64>, _>>()
        .map_err(|e| refused("ledger read", e))?;
    Ok(rows)
}

/// The report of [`precheck`]: every section the inputs could fill, newest commits first. The
/// wiring in `src/ops/card.rs` calls this and prints the envelope.
pub fn precheck(
    card: &str,
    integration_ref: Option<&str>,
    max_commits: Option<usize>,
    project: &Path,
    home: Option<&Path>,
) -> Result<Value> {
    let card_ref = input::parse_ref(card).map_err(|e| refused("bad card name", e))?;
    let max = max_commits
        .unwrap_or(DEFAULT_MAX_COMMITS)
        .clamp(1, MAX_COMMITS_CEILING);
    let home = resolve_home(home.map(Path::to_path_buf)).map_err(|e| refused("no home", e))?;
    let project_id = project_id_of(project).map_err(|e| refused("no project profile", e))?;
    let ledger_file =
        ledger::ledger_path(&home, &project_id).map_err(|e| refused("no ledger path", e))?;
    let ledger_present = ledger_file.is_file();

    // The git half runs with or without a ledger: the card's number comes from the argument.
    // A `#<n>` card needs the ledger to say which card it is, so without one there is no
    // number to scan with and the run refuses as the other card actions do.
    let (card_id, name) = match card_ref {
        CardRef::Id(n) => (n, input::card_name(n)),
        CardRef::SourceNumber(n) if !ledger_present => {
            return Err(not_found(format!(
                "card not found: there is no #{n}; the project has no ledger"
            )))
        }
        CardRef::SourceNumber(n) => {
            let read = ledger::open::open_read_only(&home, &project_id).map_err(ledger_blocked)?;
            let rows = cards_of_source(read.conn(), n)?;
            match rows.as_slice() {
                [one] => (*one, input::card_name(*one)),
                [] => return Err(not_found(format!("card not found: there is no #{n}"))),
                _ => {
                    return Err(refused(
                        "ambiguous card",
                        format!("#{n} names more than one card; write C-<number>"),
                    ))
                }
            }
        }
    };

    // The ledger half first: the linked SPEC feeds the git half's trailer matching.
    let (mut report, linked) = if ledger_present {
        match ledger_report(&home, &project_id, card_id, &name) {
            Ok(done) => done,
            // A card the ledger does not have is the caller's mistake; every other ledger
            // failure (not a database, unreadable rows) is named inside the answer and the
            // sections that need no ledger still fill (REQ-2).
            Err(r) if r.kind != "not_found" => (
                json!({
                    "ledger": "unreadable",
                    "ledger_error": r.message,
                    "spec": {"linked": null, "folders_naming_the_card": []},
                    "decision_notes": [],
                }),
                None,
            ),
            Err(r) => return Err(r),
        }
    } else {
        (
            json!({
                "ledger": "absent",
                "spec": {"linked": null, "folders_naming_the_card": []},
                "decision_notes": [],
            }),
            None,
        )
    };
    // SPEC folders are files of the project, not rows of the ledger: they are scanned in both
    // cases, and the linked SPEC stays what the ledger says.
    report["spec"]["folders_naming_the_card"] = json!(spec_folders_naming(project, card_id));
    let (git, mentions) = git_report(project, integration_ref, max, card_id, linked.as_deref());

    let mut result = json!({
        "action": "precheck",
        "card": name,
    });
    let object = result.as_object_mut().unwrap();
    for (key, value) in report.as_object().unwrap() {
        object.insert(key.clone(), value.clone());
    }
    object.insert("git".into(), git);
    object.insert("merged_mentions".into(), mentions);
    Ok(result)
}

/// One `git` call under the shared environment rules.
fn git(project: &Path, args: &[&str]) -> std::io::Result<std::process::Output> {
    let mut cmd = std::process::Command::new("git");
    scrub_git_env(&mut cmd);
    cmd.arg("--no-optional-locks")
        .arg("-C")
        .arg(project)
        .args(args)
        .output()
}

/// What names card `n` in a message: the three word forms, tried in this order. The answer is
/// the canonical form that matched, so `c-1` and `C-1` both answer `C-1`.
fn names_card(text: &str, n: i64) -> Option<String> {
    let forms = [format!("C-{n}"), format!("card/{n}"), format!("card {n}")];
    forms
        .into_iter()
        .find(|form| word_match(text, &form.to_lowercase()))
}

/// Case-insensitive `needle` with a word edge on both sides: the neighbouring characters are
/// neither alphanumeric nor `_`, or the edge is the text's end. `C-1` does not match `C-10`,
/// `C-19` or `AC-100`; `card/1` does not match `card/10`.
fn word_match(haystack: &str, needle: &str) -> bool {
    let Some(first) = needle.chars().next() else {
        return false;
    };
    let hay = haystack.to_lowercase();
    let mut at = 0;
    while let Some(found) = hay[at..].find(needle) {
        let start = at + found;
        let end = start + needle.len();
        let before_ok = start == 0
            || !hay[..start]
                .chars()
                .next_back()
                .is_some_and(|c| c.is_alphanumeric() || c == '_');
        let after_ok = end == hay.len()
            || !hay[end..]
                .chars()
                .next()
                .is_some_and(|c| c.is_alphanumeric() || c == '_');
        if before_ok && after_ok {
            return true;
        }
        // Step over the whole first character of the match: a byte step lands inside a
        // multi-byte character (Hangul is three bytes) and slicing there panics.
        at = start + first.len_utf8();
    }
    false
}

/// The distinctive tokens of one criteria word: an identifier-ish token (a digit, `-`, `_`,
/// `/`, `#` or `::` inside, three characters or more) or a plain alphanumeric word of five
/// characters or more that is not a function word. They are matched case-insensitively as
/// whole words against a decision note's body.
fn phrase_token(word: &str) -> Option<String> {
    const STOPWORDS: [&str; 16] = [
        "every", "other", "which", "where", "their", "there", "these", "those", "about", "after",
        "before", "between", "during", "while", "would", "should",
    ];
    const EDGES: [char; 18] = [
        '(', ')', '[', ']', '{', '}', '<', '>', '"', '\'', '.', ',', ';', ':', '!', '?', '`', '…',
    ];
    let trimmed = word.trim_matches(|c: char| EDGES.contains(&c) || c == '“' || c == '”');
    if trimmed.chars().count() < 3 {
        return None;
    }
    let lower = trimmed.to_lowercase();
    let identifier = lower
        .chars()
        .any(|c| c.is_ascii_digit() || "-_/#".contains(c))
        || lower.contains("::");
    if identifier {
        return Some(lower);
    }
    let plain = lower.chars().all(|c| c.is_alphanumeric());
    if plain && lower.chars().count() >= 5 && !STOPWORDS.contains(&lower.as_str()) {
        return Some(lower);
    }
    None
}

/// Every phrase key of the card's criteria, in text order, without duplicates.
fn phrase_tokens(criteria: &[String]) -> Vec<String> {
    let mut out: Vec<String> = Vec::new();
    for text in criteria {
        for word in text.split_whitespace() {
            if let Some(token) = phrase_token(word) {
                if !out.contains(&token) {
                    out.push(token);
                }
            }
        }
    }
    out
}

/// Does a write-scope pattern cover `path`? A pattern is a path, or a directory with `/**`
/// covering everything below it (the card scope's glob, C-53 `card check`).
fn scope_covers(pattern: &str, path: &str) -> bool {
    if let Some(dir) = pattern.strip_suffix("/**") {
        path == dir || path.starts_with(&format!("{dir}/"))
    } else {
        pattern == path
    }
}

/// Path-looking tokens of a note body: whitespace-split tokens with a `/` inside, their edges
/// stripped of punctuation. `src/alpha.rs` and `tests/beta` qualify; `deploy/` qualifies too
/// and matches nothing until a scope covers it.
fn path_candidates(body: &str) -> Vec<String> {
    const EDGES: [char; 13] = [
        '(', ')', '[', ']', '{', '}', '<', '>', '"', '\'', '`', ',', '…',
    ];
    body.split_whitespace()
        .map(|w| w.trim_matches(|c: char| EDGES.contains(&c) || ".,;:!?".contains(c)))
        .filter(|w| w.len() >= 3 && w.contains('/') && !w.contains("://"))
        .map(String::from)
        .collect()
}

/// The keys by which one decision note collides with the card: `path:` for a scope pattern the
/// note names or a note path a scope covers, `phrase:` for a criteria token the note contains.
/// The keys keep first-seen order.
fn note_collisions(body: &str, scope: &[String], criteria: &[String]) -> Vec<String> {
    let mut keys: Vec<String> = Vec::new();
    let push = |key: String, keys: &mut Vec<String>| {
        if !keys.contains(&key) {
            keys.push(key);
        }
    };
    for candidate in path_candidates(body) {
        for pattern in scope {
            if scope_covers(pattern, &candidate) {
                push(format!("path:{candidate}"), &mut keys);
            }
        }
    }
    for pattern in scope {
        let literal = pattern.strip_suffix("/**").unwrap_or(pattern);
        if body.contains(literal) {
            push(format!("path:{pattern}"), &mut keys);
        }
    }
    for token in phrase_tokens(criteria) {
        if word_match(body, &token) {
            push(format!("phrase:{token}"), &mut keys);
        }
    }
    keys
}

/// The pull-request number of a `Merge pull request #N from …` subject, the shape GitHub gives
/// a merge commit.
fn pr_of(subject: &str) -> Option<i64> {
    let rest = subject.strip_prefix("Merge pull request #")?;
    let digits: String = rest.chars().take_while(|c| c.is_ascii_digit()).collect();
    digits.parse().ok()
}

/// One scanned commit: where and how it names the card. A trailer counts as its own place: a
/// `Moai-Spec:` trailer naming the card's linked SPEC says "this commit belongs to the card's
/// SPEC" even when no `C-<n>` token is there.
fn mention_of(sha: &str, subject: &str, body: &str, n: i64, linked: Option<&str>) -> Option<Value> {
    let place = if let Some(form) = names_card(subject, n) {
        Some(("subject", form))
    } else if let Some(id) = linked {
        let trailer = body.lines().any(|l| {
            l.trim_start()
                .strip_prefix("Moai-Spec:")
                .is_some_and(|v| v.trim() == id)
        });
        if trailer {
            Some(("trailer", format!("spec:{id}")))
        } else {
            names_card(body, n).map(|form| ("body", form))
        }
    } else {
        names_card(body, n).map(|form| ("body", form))
    };
    let (place, matched) = place?;
    let pr = pr_of(subject);
    Some(json!({
        "sha": sha,
        "kind": if pr.is_some() { "pr_merge" } else { "commit" },
        "pr": pr,
        "subject": subject,
        "where": place,
        "matched": matched,
    }))
}

/// The ledger half of the report: the card's row, its linked SPEC and the colliding decision
/// notes. The card must exist; `linked` also feeds the git half's trailer matching.
fn ledger_report(
    home: &Path,
    project_id: &str,
    card_id: i64,
    name: &str,
) -> Result<(Value, Option<String>)> {
    let read = ledger::open::open_read_only(home, project_id).map_err(ledger_blocked)?;
    let conn = read.conn();

    let row: Option<(String, String)> = conn
        .query_row(
            "SELECT kind, write_scope FROM card WHERE card_id = ?1",
            params![card_id],
            |r| Ok((r.get(0)?, r.get(1)?)),
        )
        .optional()
        .map_err(|e| refused("ledger read", e))?;
    let Some((_kind, write_scope)) = row else {
        return Err(not_found(format!("card not found: there is no {name}")));
    };
    let scope: Vec<String> =
        serde_json::from_str(&write_scope).map_err(|e| refused("scope of the card", e))?;
    let criteria: Vec<String> = {
        let mut stmt = conn
            .prepare("SELECT text FROM card_ac WHERE card_id = ?1 ORDER BY seq")
            .map_err(|e| refused("ledger read", e))?;
        let out = stmt
            .query_map(params![card_id], |r| r.get(0))
            .map_err(|e| refused("ledger read", e))?
            .collect::<std::result::Result<Vec<_>, _>>()
            .map_err(|e| refused("ledger read", e))?;
        out
    };

    let linked = spec_link::kind_and_spec(conn, card_id)
        .map_err(|e| refused("ledger read", e))?
        .and_then(|(_kind, spec)| spec);

    let mut notes: Vec<Value> = Vec::new();
    {
        let mut stmt = conn
            .prepare(
                "SELECT n.note_id, n.card_id, n.author, n.ref, n.body FROM card_note n \
                 WHERE n.kind = 'decision' ORDER BY n.card_id, n.note_id",
            )
            .map_err(|e| refused("ledger read", e))?;
        let rows = stmt
            .query_map([], |r| {
                Ok((
                    r.get::<_, i64>(0)?,
                    r.get::<_, i64>(1)?,
                    r.get::<_, String>(2)?,
                    r.get::<_, Option<String>>(3)?,
                    r.get::<_, String>(4)?,
                ))
            })
            .map_err(|e| refused("ledger read", e))?
            .collect::<std::result::Result<Vec<_>, _>>()
            .map_err(|e| refused("ledger read", e))?;
        for (note_id, card_of, author, ref_, body) in rows {
            let matched_by = note_collisions(&body, &scope, &criteria);
            if matched_by.is_empty() {
                continue;
            }
            notes.push(json!({
                "card": input::card_name(card_of),
                "note_id": note_id,
                "ref": ref_,
                "author": author,
                "matched_by": matched_by,
            }));
        }
    }

    Ok((
        json!({
            "ledger": "ready",
            "spec": {"linked": linked, "folders_naming_the_card": []},
            "decision_notes": notes,
        }),
        linked,
    ))
}

/// Every SPEC folder of the project whose `spec.md` text names card `n`, sorted by id. The scan
/// reads the file's text only; a folder that does not parse is still scanned for the card's
/// name, because a half-written SPEC naming the card is exactly what the leader needs to see.
fn spec_folders_naming(project: &Path, n: i64) -> Vec<String> {
    let mut out = Vec::new();
    let Ok(entries) = std::fs::read_dir(project.join(".moai/specs")) else {
        return out;
    };
    for entry in entries.flatten() {
        let file_name = entry.file_name();
        let Some(id) = file_name.to_str() else {
            continue;
        };
        if !id.starts_with("SPEC-") {
            continue;
        }
        if let Ok(text) = std::fs::read_to_string(entry.path().join("spec.md")) {
            if names_card(&text, n).is_some() {
                out.push(id.to_string());
            }
        }
    }
    out.sort();
    out
}

/// The git half: the `git` section and the `merged_mentions` of the report. `max` bounds the
/// scan (one commit beyond it is read to learn whether the scan was cut short); a failure is
/// an advisory `error` inside the section, never a refusal. With no `--integration-ref` the
/// current branch's upstream is the ref, the branch a leader checks out before dispatching.
fn git_report(
    project: &Path,
    integration_ref: Option<&str>,
    max: usize,
    n: i64,
    linked: Option<&str>,
) -> (Value, Value) {
    let failure = |ref_name: &str, error: String| {
        (
            json!({"ref": ref_name, "head": null, "scanned": 0, "truncated": false,
                   "error": error}),
            json!([]),
        )
    };
    let (ref_arg, ref_name) = match integration_ref {
        Some(r) => (r.to_string(), r.to_string()),
        None => {
            let upstream = "the current branch's upstream";
            match git(
                project,
                &[
                    "rev-parse",
                    "--abbrev-ref",
                    "--symbolic-full-name",
                    "@{upstream}",
                ],
            ) {
                Ok(out) if out.status.success() => {
                    let name = String::from_utf8_lossy(&out.stdout).trim().to_string();
                    if name.is_empty() {
                        return failure(
                            upstream,
                            "no upstream for the current branch; pass --integration-ref".into(),
                        );
                    }
                    (name.clone(), name)
                }
                Ok(out) => {
                    let stderr = String::from_utf8_lossy(&out.stderr).trim().to_string();
                    return failure(upstream, format!("{stderr}; pass --integration-ref"));
                }
                Err(e) => return failure(upstream, format!("git did not run: {e}")),
            }
        }
    };
    // `--end-of-options` makes a ref that starts with `-` a revision, never an option
    // (`--output=<file>` once overwrote that file). `-z` with NUL between the fields: a commit
    // message cannot hold a NUL, so no message can forge a record.
    let log = git(
        project,
        &[
            "log",
            "--format=%H%x00%s%x00%b",
            "-z",
            "-n",
            &max.saturating_add(1).to_string(),
            "--end-of-options",
            &ref_arg,
        ],
    );
    let out = match log {
        Ok(out) if out.status.success() => out,
        Ok(out) => {
            let stderr = String::from_utf8_lossy(&out.stderr);
            let first = stderr.lines().next().unwrap_or("git log failed");
            return failure(&ref_name, first.to_string());
        }
        Err(e) => return failure(&ref_name, format!("git did not run: {e}")),
    };
    let text = String::from_utf8_lossy(&out.stdout).into_owned();
    let mut records: Vec<(String, String, String)> = Vec::new();
    let mut fields: Vec<&str> = text.split('\0').collect();
    if fields.last() == Some(&"") {
        fields.pop(); // the terminator after the last commit
    }
    for record in fields.chunks_exact(3) {
        records.push((
            record[0].trim().to_string(),
            record[1].to_string(),
            record[2].to_string(),
        ));
    }
    let truncated = records.len() > max;
    records.truncate(max);
    let head = records.first().map(|(sha, _, _)| sha.clone());
    let mentions: Vec<Value> = records
        .iter()
        .filter_map(|(sha, subject, body)| mention_of(sha, subject, body, n, linked))
        .collect();
    (
        json!({"ref": ref_name, "head": head, "scanned": records.len(), "truncated": truncated}),
        json!(mentions),
    )
}

#[cfg(test)]
mod tests {
    use super::*;

    /// AC-4: the canonical form answers whatever case the message uses.
    #[test]
    fn names_card_is_case_insensitive_and_canonical() {
        assert_eq!(
            names_card("fix: lands (c-1) too", 1).as_deref(),
            Some("C-1")
        );
        assert_eq!(names_card("fix: (C-1)", 1).as_deref(), Some("C-1"));
        assert_eq!(
            names_card("Merge pull request #3 from modu-ai/card/1-demo", 1).as_deref(),
            Some("card/1")
        );
        assert_eq!(
            names_card("the PR title (card 1)", 1).as_deref(),
            Some("card 1")
        );
    }

    /// Audit B2: the scan steps by whole characters, so a needle that starts with a multi-byte
    /// character does not slice inside it after a failed word edge.
    #[test]
    fn word_match_steps_by_whole_characters() {
        assert!(!word_match("그원장에서는 다르게", "원장에서는"));
        assert!(word_match("그원장에서는 원장에서는", "원장에서는"));
        assert!(word_match("가 C-1 나", "c-1"));
        assert!(!word_match("anything", ""));
    }

    /// AC-4: a longer number is a different card — `C-1` does not match `C-10`, `C-19`,
    /// `AC-100`, and `card/1` does not match `card/10`.
    #[test]
    fn names_card_keeps_word_edges() {
        assert_eq!(names_card("fix: (C-10)", 1), None);
        assert_eq!(names_card("see C-19 and card/10", 1), None);
        assert_eq!(names_card("chore: AC-100 cleanup", 1), None);
        assert_eq!(names_card("XC-96 hidden", 96), None);
        assert!(names_card("(C-1)", 1).is_some());
    }

    /// AC-4: phrase tokens keep identifier-ish words and long plain words, drop short words
    /// and function words.
    #[test]
    fn phrase_tokens_keep_identifiers_and_drop_function_words() {
        let tokens = phrase_tokens(&[
            "N1: the hooks offer reads ready when #2936 names it; user_chat follows Lang::En"
                .to_string(),
            "every branch between builds".to_string(),
        ]);
        let keys = tokens.iter().map(String::as_str).collect::<Vec<_>>();
        for expected in [
            "hooks",
            "offer",
            "reads",
            "ready",
            "#2936",
            "names",
            "user_chat",
            "lang::en",
        ] {
            assert!(keys.contains(&expected), "{expected} missing: {keys:?}");
        }
        for excluded in ["n1", "the", "when", "it", "every", "between"] {
            assert!(!keys.contains(&excluded), "{excluded} kept: {keys:?}");
        }
        // A plain word of five characters that is not a function word stays ('branch').
        assert!(keys.contains(&"branch"), "{keys:?}");
    }

    /// AC-4: `dir/**` covers the directory and everything below it, and only that.
    #[test]
    fn scope_covers_glob_and_exact() {
        assert!(scope_covers("tests/beta/**", "tests/beta"));
        assert!(scope_covers("tests/beta/**", "tests/beta/x.rs"));
        assert!(!scope_covers("tests/beta/**", "tests/betax"));
        assert!(scope_covers("src/alpha.rs", "src/alpha.rs"));
        assert!(!scope_covers("src/alpha.rs", "src/alpha.rs2"));
    }

    /// AC-4: a note that names a scope's directory matches by the pattern — the other
    /// direction of the path rule — and a note path under the directory matches by itself and
    /// by the pattern.
    #[test]
    fn note_collisions_match_both_path_directions_and_phrases() {
        let scope = vec!["tests/beta/**".to_string()];
        let criteria = vec!["N1: the hooks offer".to_string()];
        assert_eq!(
            note_collisions("note names tests/beta for another lane", &scope, &criteria),
            vec![
                "path:tests/beta".to_string(),
                "path:tests/beta/**".to_string(),
            ]
        );
        assert_eq!(
            note_collisions("note holds tests/beta/deep/file.rs", &scope, &criteria),
            vec![
                "path:tests/beta/deep/file.rs".to_string(),
                "path:tests/beta/**".to_string(),
            ]
        );
        assert_eq!(
            note_collisions("the hooks decision stands", &[], &criteria),
            vec!["phrase:hooks".to_string()]
        );
        assert!(note_collisions("unrelated deploy/ text", &scope, &criteria).is_empty());
    }

    /// AC-4: the pull-request number of a merge subject, and nothing else's.
    #[test]
    fn pr_of_reads_merge_subjects_only() {
        assert_eq!(
            pr_of("Merge pull request #3151 from modu-ai/card/88-slot"),
            Some(3151)
        );
        assert_eq!(pr_of("Merge branch 'x'"), None);
        assert_eq!(pr_of("fix: (C-1)"), None);
    }

    /// AC-4: a `Moai-Spec:` trailer names the card only when it names the card's linked SPEC;
    /// empty criteria and an empty scope produce no keys at all.
    #[test]
    fn trailer_and_empty_inputs() {
        assert_eq!(
            mention_of(
                "a1b2",
                "chore: one",
                "Moai-Spec: SPEC-20261119-demo-linked\n",
                1,
                Some("SPEC-20261119-demo-linked")
            )
            .unwrap()["matched"],
            json!("spec:SPEC-20261119-demo-linked")
        );
        assert_eq!(
            mention_of(
                "a1b2",
                "chore: one",
                "Moai-Spec: SPEC-other\n",
                1,
                Some("SPEC-x")
            ),
            None
        );
        assert!(phrase_tokens(&[]).is_empty());
        assert!(note_collisions("any text at all src/x.rs", &[], &[]).is_empty());
    }
}
