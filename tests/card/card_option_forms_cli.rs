//! The usage mistakes a lane repeats on `moai card note`, `moai card spec` and `moai card
//! verdict` (card C-608): the other card commands take `--generation`, so a lane gives it to
//! `note` and `spec` by habit and got a bare usage error; `--verdict PASS` in capitals was refused
//! though the word is the right one. `note` and `spec` now take `--generation` as `card scope`
//! does (given: the card's live reservation must be at it; left out: nothing is checked) and
//! `--verdict` is read without regard to case. The note-shape gates are not touched.
//!
//! Every command runs against a temporary HOME and a temporary git project; nothing of this
//! machine is read or written.
#![cfg(unix)]

use crate::common;

use common::home::TempHome;
use moai_cli::ledger;
use serde_json::{json, Value};
use std::fs;
use std::io::Write;
use std::path::Path;
use std::process::{Output, Stdio};

const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
const S_LANE: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
const S_AUD: &str = "cccccccc-0000-4000-8000-00000000000c";

fn git(dir: &Path, args: &[&str]) -> String {
    let mut all = vec![
        "-c",
        "user.name=fixture",
        "-c",
        "user.email=fixture@example.invalid",
    ];
    all.extend_from_slice(args);
    common::git::git(dir, &all)
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
    v["result"].clone()
}

/// The refusal of a call: exit 1 and the message, never a usage error (exit 2).
fn refusal(out: &Output) -> String {
    assert_eq!(
        out.status.code(),
        Some(1),
        "{} {}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    let v = envelope(out);
    assert_eq!(v["ok"], json!(false), "{v}");
    v["error"]["message"].as_str().unwrap().to_string()
}

/// A text with every `<…>` placeholder of a SPEC template filled, so `spec check` passes it.
fn filled(text: &str) -> String {
    let mut out = String::new();
    let mut rest = text;
    while let Some(open) = rest.find('<') {
        let Some(close) = rest[open..].find('>') else {
            break;
        };
        out.push_str(&rest[..open]);
        out.push_str("filled");
        rest = &rest[open + close + 1..];
    }
    out.push_str(rest);
    out
}

struct World {
    home: TempHome,
    project: std::path::PathBuf,
    p: String,
    /// Three commits in a line: the base, the head a PASS is recorded on and the head of a FAIL.
    base: String,
    head: String,
    head2: String,
}

impl World {
    fn new() -> Self {
        let home = TempHome::new();
        let project = home.project("demo");
        ledger::register(home.path(), "demo").unwrap();
        git(&project, &["init", "-q", "-b", "main"]);
        let mut shas = Vec::new();
        for name in ["a.txt", "b.txt", "c.txt"] {
            fs::write(project.join(name), format!("{name}\n")).unwrap();
            git(&project, &["add", name]);
            git(&project, &["commit", "-q", "-m", name]);
            shas.push(git(&project, &["rev-parse", "HEAD"]));
        }
        World {
            p: project.to_str().unwrap().to_string(),
            home,
            project,
            base: shas[0].clone(),
            head: shas[1].clone(),
            head2: shas[2].clone(),
        }
    }

    /// `moai card <args> --project <p>` as `session`, with `stdin` on standard input.
    fn card(&self, session: Option<&str>, args: &[&str], stdin: &str) -> Output {
        let mut all = vec!["card"];
        all.extend_from_slice(args);
        all.extend_from_slice(&["--project", &self.p]);
        let mut cmd = self.home.command(&all);
        if let Some(id) = session {
            cmd.env(SESSION_VAR, id);
        }
        let mut child = cmd
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .spawn()
            .expect("run moai");
        child
            .stdin
            .take()
            .unwrap()
            .write_all(stdin.as_bytes())
            .unwrap();
        child.wait_with_output().unwrap()
    }

    /// A ready card with one open criterion; returns its name.
    fn ready(&self) -> String {
        let name = result(&self.card(
            None,
            &[
                "add", "--title", "fixture", "--kind", "task", "--scope", "apps/a/", "--ac", "one",
            ],
            "",
        ))["name"]
            .as_str()
            .unwrap()
            .to_string();
        result(&self.card(None, &["ready", &name], ""));
        name
    }

    /// A ready card taken by the lane session at generation 1.
    fn held(&self) -> String {
        let name = self.ready();
        let taken = result(&self.card(
            Some(S_LANE),
            &[
                "take",
                &name,
                "--worktree",
                "/work/tree",
                "--branch",
                "card/x",
            ],
            "",
        ));
        assert_eq!(taken["reservation"]["generation"], json!(1), "{taken}");
        name
    }

    /// `card note <card> --kind report --by lane [--generation g]` with a plain body.
    fn note(&self, card: &str, generation: Option<&str>) -> Output {
        let mut args = vec!["note", card, "--kind", "report", "--by", "lane"];
        if let Some(g) = generation {
            args.extend_from_slice(&["--generation", g]);
        }
        self.card(Some(S_LANE), &args, "a plain progress note\n")
    }

    fn count(&self, sql: &str) -> i64 {
        let l = ledger::open_read_only(self.home.path(), "demo").unwrap();
        l.conn().query_row(sql, [], |r| r.get(0)).unwrap()
    }

    fn notes(&self) -> i64 {
        self.count("SELECT COUNT(*) FROM card_note")
    }

    fn spec_links(&self) -> i64 {
        self.count("SELECT COUNT(*) FROM event WHERE kind = 'card.spec'")
    }

    /// A SPEC every `spec check` passes; returns its id.
    fn spec(&self, slug: &str) -> String {
        let made = result(
            &self
                .home
                .command(&[
                    "spec",
                    "new",
                    slug,
                    "--title",
                    "t",
                    "--minimal",
                    "--date",
                    "2026-11-17",
                    "--project",
                    &self.p,
                ])
                .output()
                .expect("run moai"),
        );
        let id = made["id"].as_str().unwrap().to_string();
        // The SPEC folder is `MOAI_DIR/specs` whatever the project's own state folder resolves
        // to (`spec_doc::specs_dir` joins the same constant).
        let path = self
            .project
            .join(moai_cli::state_root::MOAI_DIR)
            .join("specs")
            .join(&id)
            .join("spec.md");
        fs::write(&path, filled(&fs::read_to_string(&path).unwrap())).unwrap();
        id
    }

    /// `card verdict <card> --head <head> --base <base> --verdict <word> --tier 0 --evidence …`
    /// as the auditor session.
    fn verdict(&self, card: &str, head: &str, word: &str) -> Output {
        self.card(
            Some(S_AUD),
            &[
                "verdict",
                card,
                "--head",
                head,
                "--base",
                &self.base,
                "--verdict",
                word,
                "--tier",
                "0",
                "--evidence",
                "read the diff, nothing to run",
            ],
            "",
        )
    }
}

#[test]
fn a_note_with_the_generation_the_card_is_held_at_is_written() {
    let w = World::new();
    let card = w.held();
    let before = w.notes();
    let r = result(&w.note(&card, Some("1")));
    assert_eq!(r["action"], json!("note"), "{r}");
    assert_eq!(w.notes(), before + 1, "the note is kept");
}

#[test]
fn a_note_with_another_generation_is_refused_naming_the_live_one_and_writes_nothing() {
    let w = World::new();
    let card = w.held();
    let before = w.notes();
    let message = refusal(&w.note(&card, Some("2")));
    assert!(
        message.contains("generation 1") && message.contains("not 2"),
        "the refusal names the live generation and the given one: {message}"
    );
    assert_eq!(w.notes(), before, "nothing is written");
}

#[test]
fn a_note_with_a_generation_on_a_card_nobody_holds_says_it_is_not_held() {
    let w = World::new();
    let card = w.ready();
    let before = w.notes();
    let message = refusal(&w.note(&card, Some("1")));
    assert!(message.contains("not held"), "{message}");
    assert!(
        message.contains("--generation"),
        "the refusal names the option to leave out: {message}"
    );
    assert_eq!(w.notes(), before, "nothing is written");
}

#[test]
fn a_note_without_a_generation_is_checked_against_nothing() {
    let w = World::new();
    let held = w.held();
    let free = w.ready();
    let before = w.notes();
    result(&w.note(&held, None));
    result(&w.note(&free, None));
    assert_eq!(w.notes(), before + 2, "held or not, the note is kept");
}

#[test]
fn the_note_gates_still_refuse_when_the_generation_is_right() {
    let w = World::new();
    let card = w.held();
    let before = w.notes();
    // A decision note without `--by`, and a note with an unknown kind: the same refusals as ever.
    let decision = w.card(
        Some(S_LANE),
        &["note", &card, "--kind", "decision", "--generation", "1"],
        "a decision\n",
    );
    assert!(refusal(&decision).contains("--by"));
    let kind = w.card(
        Some(S_LANE),
        &[
            "note",
            &card,
            "--kind",
            "plan",
            "--by",
            "lane",
            "--generation",
            "1",
        ],
        "a plan\n",
    );
    assert_eq!(kind.status.code(), Some(1), "unknown kind");
    assert_eq!(w.notes(), before, "neither is written");
}

#[test]
fn a_spec_link_with_the_generation_the_card_is_held_at_is_written() {
    let w = World::new();
    let card = w.held();
    let spec = w.spec("option-forms");
    let r = result(&w.card(
        None,
        &["spec", &card, &spec, "--generation", "1", "--by", "leader"],
        "",
    ));
    assert_eq!(r["spec"], json!(spec), "{r}");
    assert_eq!(w.spec_links(), 1, "one link event");
}

#[test]
fn a_spec_link_with_another_generation_is_refused_and_links_nothing() {
    let w = World::new();
    let card = w.held();
    let spec = w.spec("option-forms");
    let message = refusal(&w.card(
        None,
        &["spec", &card, &spec, "--generation", "3", "--by", "leader"],
        "",
    ));
    assert!(
        message.contains("generation 1") && message.contains("not 3"),
        "{message}"
    );
    assert_eq!(w.spec_links(), 0, "nothing is linked");
}

#[test]
fn a_spec_link_without_a_generation_is_checked_against_nothing() {
    let w = World::new();
    let card = w.held();
    let spec = w.spec("option-forms");
    result(&w.card(None, &["spec", &card, &spec, "--by", "leader"], ""));
    assert_eq!(w.spec_links(), 1);
}

#[test]
fn a_verdict_is_read_without_regard_to_case() {
    let w = World::new();
    let card = w.held();
    let pass = result(&w.verdict(&card, &w.head, "PASS"));
    assert_eq!(pass["verdict"], json!("pass"), "{pass}");
    let fail = result(&w.verdict(&card, &w.head2, " Fail "));
    assert_eq!(fail["verdict"], json!("fail"), "{fail}");
}

#[test]
fn a_word_that_is_not_pass_or_fail_is_still_refused_naming_both_words() {
    let w = World::new();
    let card = w.held();
    for word in ["maybe", "passed", "ok"] {
        let message = refusal(&w.verdict(&card, &w.head, word));
        assert!(
            message.contains("`pass`") && message.contains("`fail`") && message.contains(word),
            "{word}: {message}"
        );
    }
    assert_eq!(
        w.count("SELECT COUNT(*) FROM event WHERE kind = 'card.verdict'"),
        0,
        "nothing is recorded"
    );
}
