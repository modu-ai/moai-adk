//! `moai card` secret-screen follow-ups (card C-132, SPEC-20261120-card-secret-followups): the
//! audit N3·N5·N7 shapes the screen of SPEC-20261119-card-secret-screen still lets through —
//! name forms (blanks around the equals sign, quoted names, camel/joined names), the `--reason`
//! of `card reclaim`, values after the separators beyond `= : ,` — each on the four surfaces
//! (note text, title, body, `--ac`), plus the one-long-word time bound with the new separators.
//!
//! Every secret-shaped word is joined from pieces so this file holds none; every command runs
//! against a temporary HOME; the real `~/.moai` is not read or written.

mod common;

use common::home::TempHome;
use moai_cli::card::input;
use moai_cli::ledger;
use serde_json::{json, Value};
use std::path::Path;
use std::process::{Output, Stdio};

const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
const S_A: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
const S_B: &str = "bbbbbbbb-0000-4000-8000-00000000000b";

/// Joins pieces so this file holds no secret-shaped word.
fn words(parts: &[&str]) -> String {
    parts.concat()
}

fn run(home: &TempHome, args: &[&str]) -> Output {
    home.command(args).output().expect("run moai")
}

/// `moai card <args>` as `session`: the holder comes from the environment's session id.
fn card_as(home: &TempHome, session: &str, args: &[&str]) -> Output {
    let mut all = vec!["card"];
    all.extend_from_slice(args);
    let mut cmd = home.command(&all);
    cmd.env(SESSION_VAR, session);
    cmd.output().expect("run moai")
}

fn envelope(out: &Output) -> Value {
    serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
        panic!(
            "stdout is not JSON ({e}): {}",
            String::from_utf8_lossy(&out.stdout)
        )
    })
}

fn result(out: &Output) -> Value {
    assert_eq!(
        out.status.code(),
        Some(0),
        "{} {}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    let v = envelope(out);
    assert_eq!(v["ok"], json!(true), "{v}");
    assert_eq!(v["operation"], json!("card"), "{v}");
    v["result"].clone()
}

/// The refusal: exit 1, its `kind` and its message.
fn refusal(out: &Output) -> (String, String) {
    assert_eq!(
        out.status.code(),
        Some(1),
        "{}",
        String::from_utf8_lossy(&out.stdout)
    );
    let v = envelope(out);
    assert_eq!(v["ok"], json!(false), "{v}");
    (
        v["error"]["kind"].as_str().unwrap().to_string(),
        v["error"]["message"].as_str().unwrap().to_string(),
    )
}

fn p(project: &Path) -> String {
    project.to_str().unwrap().to_string()
}

/// A HOME with a project `demo` and its registered ledger.
fn registered() -> (TempHome, String) {
    let home = TempHome::new();
    let project = home.project("demo");
    ledger::register(home.path(), "demo").unwrap();
    (home, p(&project))
}

fn count(home: &TempHome, table: &str) -> i64 {
    let l = ledger::open_read_only(home.path(), "demo").unwrap();
    l.conn()
        .query_row(&format!("SELECT count(*) FROM {table}"), [], |r| r.get(0))
        .unwrap()
}

/// The card `C-1` of these tests, so a note has a card to go on.
fn with_seed_card() -> (TempHome, String) {
    let (home, project) = registered();
    result(&run(
        &home,
        &["card", "add", "--title", "seed", "--project", &project],
    ));
    (home, project)
}

/// `card note C-1 --kind report --by agent-7` with `text` on stdin.
fn note_with(home: &TempHome, project: &str, text: &str) -> Output {
    let args = [
        "card",
        "note",
        "C-1",
        "--kind",
        "report",
        "--by",
        "agent-7",
        "--project",
        project,
    ];
    let mut child = home
        .command(&args)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .expect("run moai");
    common::feed_stdin(&mut child, text.as_bytes());
    child.wait_with_output().unwrap()
}

/// The four text surfaces the screen covers: a note, a title, a body and an `--ac` entry.
fn on_every_surface(home: &TempHome, project: &str, text: &str) -> Vec<(&'static str, Output)> {
    vec![
        ("note", note_with(home, project, text)),
        (
            "title",
            run(
                home,
                &["card", "add", "--title", text, "--project", project],
            ),
        ),
        (
            "body",
            run(
                home,
                &[
                    "card",
                    "add",
                    "--title",
                    "t",
                    "--body",
                    text,
                    "--project",
                    project,
                ],
            ),
        ),
        (
            "criterion",
            run(
                home,
                &[
                    "card",
                    "add",
                    "--title",
                    "t",
                    "--ac",
                    text,
                    "--project",
                    project,
                ],
            ),
        ),
    ]
}

/// A held card `C-2` (added with a scope, made ready, taken by `S_A`), so `card reclaim` has a
/// card to name.
fn held_card(home: &TempHome, project: &str) {
    result(&run(
        home,
        &[
            "card",
            "add",
            "--title",
            "held",
            "--ac",
            "one",
            "--scope",
            "src/a/",
            "--project",
            project,
        ],
    ));
    result(&run(home, &["card", "ready", "C-2", "--project", project]));
    result(&card_as(home, S_A, &["take", "C-2", "--project", project]));
}

/// AC-1 (REQ-1..4): the three N3 name forms are refused on every surface, and plain prose that
/// holds the same shapes without an assignment — or a spaced equals sign behind a plain name —
/// still passes.
#[test]
fn name_forms_blanks_quotes_and_camel_are_refused_on_every_surface() {
    let (home, project) = with_seed_card();
    let (tk, ak) = (words(&["to", "ken"]), words(&["api", "-", "key"]));
    let refused = [
        words(&["run --my-", &tk, " = 7"]),
        words(&["run ", &ak, " = 7"]),
        words(&["run \"my-", &ak, "\"=x"]),
        words(&["run myApi", "Key=1"]),
        words(&["run my", "Token", "Count=1"]),
    ];
    for text in &refused {
        for (surface, out) in on_every_surface(&home, &project, text) {
            let (kind, message) = refusal(&out);
            assert_eq!(kind, "invalid_input", "{surface}: {text}: {message}");
            assert!(
                message.contains("shape of a secret"),
                "{surface}: {text}: {message}"
            );
        }
    }
    assert_eq!(count(&home, "card"), 1, "only the seed card exists");
    assert_eq!(count(&home, "card_note"), 0);
    assert_eq!(count(&home, "card_ac"), 0);

    let plain = [
        words(&["the ", &tk, " count is seven"]),
        words(&["a quoted \"my-", &ak, "\" mention"]),
        "count = total".to_string(),
        words(&["see the ", &tk, "s list"]),
        words(&["run ", &tk, "s=3 wide=1"]),
        "run --config=prod.yaml --api-version=2".to_string(),
    ];
    for text in &plain {
        for (surface, out) in on_every_surface(&home, &project, text) {
            let r = result(&out);
            assert!(
                r["action"] == json!("note") || r["action"] == json!("add"),
                "{surface}: {text}: {r}"
            );
        }
    }
}

/// AC-2 (REQ-5): the reclaim reason meets the same one screen — a secret-named assignment in it
/// is refused before any ledger read, a clean reason still passes that check.
#[test]
fn reclaim_reason_shares_the_one_screen() {
    let (home, project) = with_seed_card();
    held_card(&home, &project);
    let bad = words(&["holder gone; --max-", "to", "ken-age=5 stayed plain"]);
    let (kind, message) = refusal(&card_as(
        &home,
        S_B,
        &[
            "reclaim",
            "C-2",
            "--generation",
            "1",
            "--by",
            "leader",
            "--reason",
            &bad,
            "--project",
            &project,
        ],
    ));
    assert_eq!(kind, "invalid_input", "{message}");
    assert!(message.contains("shape of a secret"), "{message}");

    // A clean reason passes the screen; the command refuses later, for the live lease, without
    // naming a secret shape.
    let (kind, message) = refusal(&card_as(
        &home,
        S_B,
        &[
            "reclaim",
            "C-2",
            "--generation",
            "1",
            "--by",
            "leader",
            "--reason",
            "holder gone; lease ended cleanly",
            "--project",
            &project,
        ],
    ));
    assert_ne!(kind, "invalid_input", "{message}");
    assert!(!message.contains("shape of a secret"), "{message}");
}

/// AC-3 (REQ-6): a credential-shaped value after each separator beyond `= : ,` is refused on
/// every surface, and ordinary separator prose passes.
#[test]
fn separators_values_are_refused_on_every_surface() {
    let (home, project) = with_seed_card();
    let key = words(&["AK", "IA", "ABCDEFGHIJKLMNOP"]);
    for sep in [";", "/", "|", "@", "?", "&", ".", "+", "#", "\\", "("] {
        let text = format!("see{sep}{key}");
        for (surface, out) in on_every_surface(&home, &project, &text) {
            let (kind, message) = refusal(&out);
            assert_eq!(kind, "invalid_input", "{surface} `{sep}`: {message}");
        }
    }
    assert_eq!(count(&home, "card_note"), 0);
    assert_eq!(count(&home, "card_ac"), 0);

    let plain = [
        "see; the docs/guide.md, page 3.5".to_string(),
        "a|b c&d e?f g+h".to_string(),
        "user@site.com x#y (see)".to_string(),
    ];
    for text in &plain {
        for (surface, out) in on_every_surface(&home, &project, text) {
            let r = result(&out);
            assert!(
                r["action"] == json!("note") || r["action"] == json!("add"),
                "{surface}: {text}: {r}"
            );
        }
    }
}

/// AC-3 (REQ-7): with the new separators in the word, one long word without a space is still
/// answered in time that grows with its length (the bound is far above linear, far below
/// square), and the value after the last separator is seen.
#[test]
fn separators_long_one_word_stays_within_the_time_bound() {
    fn timed(text: String) -> Result<usize, String> {
        let (tx, rx) = std::sync::mpsc::channel();
        // The screen runs on its own thread, so a square-time regression fails after the bound
        // instead of holding the test for minutes (the thread dies with the test process).
        std::thread::spawn(move || {
            let _ = tx.send(input::secret_words(&text));
        });
        rx.recv_timeout(std::time::Duration::from_secs(15))
            .map_err(|_| "no answer within 15 s".to_string())
    }
    let key = words(&["AK", "IA", "ABCDEFGHIJKLMNOP"]);
    for size in [200_000usize, 1_048_575] {
        for unit in ["a;", "a.", "a/"] {
            let mut text = unit.repeat(size / unit.len());
            text.push_str(&key);
            let refused = timed(text).unwrap_or_else(|e| panic!("size {size} unit {unit}: {e}"));
            assert_eq!(
                refused, 1,
                "size {size} unit {unit}: the value after the last separator is refused"
            );
        }
    }
}
