//! `moai card verdict` (SPEC-20261123-landing-reads-verdict): a session records the verdict of
//! an audit as one structured event of the card, which `card land` reads. Every test runs on a
//! temporary home and a temporary repository; nothing here touches the real ledger.

mod common;

use common::home::TempHome;
use moai_cli::ledger;
use rusqlite::Connection;
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Output, Stdio};

const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
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

struct World {
    home: TempHome,
    #[allow(dead_code)]
    project: PathBuf,
    p: String,
    /// The first commit, the second (the card's change) and a commit that is no ancestor of it.
    base: String,
    head: String,
    side: String,
}

impl World {
    fn new() -> Self {
        let home = TempHome::new();
        let project = home.project("demo");
        ledger::register(home.path(), "demo").unwrap();
        git(&project, &["init", "-q", "-b", "main"]);
        fs::write(project.join("a.txt"), "a\n").unwrap();
        git(&project, &["add", "a.txt"]);
        git(&project, &["commit", "-q", "-m", "init"]);
        let base = git(&project, &["rev-parse", "HEAD"]);
        fs::write(project.join("b.txt"), "b\n").unwrap();
        git(&project, &["add", "b.txt"]);
        git(&project, &["commit", "-q", "-m", "the card's change"]);
        let head = git(&project, &["rev-parse", "HEAD"]);
        git(&project, &["switch", "-q", "-c", "side", &base]);
        fs::write(project.join("c.txt"), "c\n").unwrap();
        git(&project, &["add", "c.txt"]);
        git(
            &project,
            &["commit", "-q", "-m", "a commit beside the change"],
        );
        let side = git(&project, &["rev-parse", "HEAD"]);
        git(&project, &["switch", "-q", "main"]);
        World {
            p: project.to_str().unwrap().to_string(),
            home,
            project,
            base,
            head,
            side,
        }
    }

    fn run(&self, session: Option<&str>, args: &[&str]) -> Output {
        let mut cmd: Command = self.home.command(args);
        if let Some(id) = session {
            cmd.env(SESSION_VAR, id);
        }
        cmd.stdin(Stdio::null());
        cmd.output().expect("run moai")
    }

    fn card_cmd(&self, session: Option<&str>, args: &[&str]) -> Output {
        let mut all = vec!["card"];
        all.extend_from_slice(args);
        all.extend_from_slice(&["--project", &self.p]);
        self.run(session, &all)
    }

    /// A ready card with one open criterion; returns its name.
    fn card(&self) -> String {
        let added = result(&self.card_cmd(
            None,
            &[
                "add",
                "--title",
                "fixture card",
                "--kind",
                "task",
                "--scope",
                "apps/a/",
                "--ac",
                "one",
            ],
        ));
        added["name"].as_str().unwrap().to_string()
    }

    /// `card verdict <card> --head <head> --base <base> …` as `session`.
    fn verdict(&self, session: Option<&str>, card: &str, extra: &[&str]) -> Output {
        let mut all = vec!["verdict", card];
        all.extend_from_slice(extra);
        self.card_cmd(session, &all)
    }

    fn dump(&self) -> Vec<String> {
        let l = ledger::open_read_only(self.home.path(), "demo").unwrap();
        let mut out = Vec::new();
        for table in ["card", "card_ac", "card_note", "reservation", "event"] {
            let mut st = l
                .conn()
                .prepare(&format!("SELECT * FROM {table} ORDER BY 1, 2"))
                .unwrap();
            let n = st.column_count();
            let rows: Vec<String> = st
                .query_map([], |r| {
                    let mut line = format!("{table}:");
                    for i in 0..n {
                        line.push_str(&format!("{:?}|", r.get_ref(i)?));
                    }
                    Ok(line)
                })
                .unwrap()
                .map(Result::unwrap)
                .collect();
            out.extend(rows);
        }
        out
    }

    fn raw(&self) -> Connection {
        Connection::open(ledger::ledger_path(self.home.path(), "demo").unwrap()).unwrap()
    }

    /// The `card.verdict` events of the card: id, actor and payload.
    fn events(&self, card: &str) -> Vec<(i64, String, Value)> {
        let n: i64 = card.trim_start_matches("C-").parse().unwrap();
        let conn = self.raw();
        let mut st = conn
            .prepare(
                "SELECT event_id, actor, payload FROM event WHERE card_id = ?1 \
                 AND kind = 'card.verdict' ORDER BY event_id",
            )
            .unwrap();
        st.query_map([n], |r| {
            let payload: String = r.get(2)?;
            Ok((
                r.get(0)?,
                r.get(1)?,
                serde_json::from_str(&payload).unwrap(),
            ))
        })
        .unwrap()
        .map(Result::unwrap)
        .collect()
    }

    /// Asserts a refusal that wrote nothing: exit 1, `ok` false, a reason with `needle`.
    fn refused(&self, out: &Output, needle: &str, before: &[String]) {
        assert_eq!(
            out.status.code(),
            Some(1),
            "{} {}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        );
        let v = envelope(out);
        assert_eq!(v["ok"], json!(false), "{v}");
        let message = v["error"]["message"].as_str().unwrap();
        assert!(
            message.contains(needle),
            "the reason lacks {needle:?}: {message}"
        );
        assert_eq!(&self.dump(), before, "the refused call wrote to the ledger");
    }
}

/// AC-1 (REQ-1, REQ-4, REQ-5): the record holds the values and the reviewer read from the host
/// session id; the same call again writes nothing and answers the first event; an option that
/// names a reviewer does not exist.
#[test]
fn ac1_a_recorded_verdict_names_its_values_and_a_repeat_writes_nothing() {
    let w = World::new();
    let card = w.card();
    let args = [
        "--head",
        &w.head,
        "--base",
        &w.base,
        "--verdict",
        "pass",
        "--tier",
        "1",
        "--evidence",
        "read the diff and ran the changed targets",
    ];
    let first = result(&w.verdict(Some(S_AUD), &card, &args));
    assert_eq!(first["action"], json!("verdict"), "{first}");
    assert_eq!(first["written"], json!(true), "{first}");
    let events = w.events(&card);
    assert_eq!(events.len(), 1);
    let (id, actor, payload) = &events[0];
    assert_eq!(first["event"], json!(id), "{first}");
    assert_eq!(actor, S_AUD);
    assert_eq!(payload["head"], json!(w.head));
    assert_eq!(payload["base"], json!(w.base));
    assert_eq!(payload["verdict"], json!("pass"));
    assert_eq!(payload["tier"], json!(1));
    assert_eq!(
        payload["evidence"],
        json!("read the diff and ran the changed targets")
    );
    assert_eq!(payload["reviewer"], json!(S_AUD));
    assert_eq!(payload["receipt"], Value::Null);

    // The same call again: nothing written, the first event answered.
    let before = w.dump();
    let again = result(&w.verdict(Some(S_AUD), &card, &args));
    assert_eq!(again["written"], json!(false), "{again}");
    assert_eq!(again["event"], json!(id), "{again}");
    assert_eq!(w.dump(), before, "a repeat wrote");

    // Another reviewer is another event; a typed reviewer is no option at all.
    let mut typed = args.to_vec();
    typed.extend_from_slice(&["--reviewer", "someone-else"]);
    let out = w.verdict(Some(S_AUD), &card, &typed);
    assert_ne!(out.status.code(), Some(0));
    assert_eq!(w.dump(), before, "a refused option wrote");
}

/// AC-2 (REQ-2): each malformed input refuses and the ledger is as it was.
#[test]
fn ac2_a_malformed_verdict_is_refused_and_writes_nothing() {
    let w = World::new();
    let card = w.card();
    let before = w.dump();
    let good = |head: &str, base: &str, evidence: &str| -> Vec<String> {
        [
            "--head",
            head,
            "--base",
            base,
            "--verdict",
            "pass",
            "--tier",
            "0",
            "--evidence",
            evidence,
        ]
        .iter()
        .map(|s| s.to_string())
        .collect()
    };
    let run = |session: Option<&str>, args: Vec<String>| -> Output {
        let refs: Vec<&str> = args.iter().map(String::as_str).collect();
        w.verdict(session, &card, &refs)
    };
    let unknown = "fedcba9876543210fedcba9876543210fedcba98";
    for (args, needle) in [
        (good(&w.head[..12], &w.base, "read it"), "40 lowercase hex"),
        (
            good(&w.head.to_uppercase(), &w.base, "read it"),
            "40 lowercase hex",
        ),
        (good(unknown, &w.base, "read it"), "not a commit"),
        (good(&w.head, unknown, "read it"), "not a commit"),
        (good(&w.head, &w.side, "read it"), "not an ancestor"),
        (good(&w.head, &w.base, "   "), "evidence"),
        (
            good(
                &w.head,
                &w.base,
                "the key is sk-ant-api03-AbCdEf0123456789xyz",
            ),
            "secret",
        ),
    ] {
        let out = run(Some(S_AUD), args);
        w.refused(&out, needle, &before);
    }
    // No host session id: the reviewer cannot be read.
    let out = run(None, good(&w.head, &w.base, "read it"));
    w.refused(&out, "session", &before);
    // A tier above 2 and a verdict word that is neither pass nor fail are command line errors.
    for bad in [["--tier", "3"], ["--verdict", "maybe"]] {
        let mut args = good(&w.head, &w.base, "read it");
        let key = bad[0].to_string();
        let at = args.iter().position(|a| *a == key).unwrap();
        args[at + 1] = bad[1].to_string();
        let out = run(Some(S_AUD), args);
        assert_ne!(out.status.code(), Some(0));
        assert_eq!(w.dump(), before);
    }
}

/// AC-3 (REQ-3): a receipt is stored with the head it names and each step with its exit code; a
/// step with no exit code is refused.
#[test]
fn ac3_a_receipt_is_stored_with_its_head_and_each_step_exit() {
    let w = World::new();
    let card = w.card();
    let out = w.verdict(
        Some(S_AUD),
        &card,
        &[
            "--head",
            &w.head,
            "--base",
            &w.base,
            "--verdict",
            "pass",
            "--tier",
            "1",
            "--evidence",
            "the lane's receipt of the same head",
            "--receipt-head",
            &w.head,
            "--receipt-step",
            "fmt=0",
            "--receipt-step",
            "test=0",
        ],
    );
    result(&out);
    let events = w.events(&card);
    assert_eq!(
        events[0].2["receipt"],
        json!({"head": w.head, "steps": [{"name": "fmt", "exit": 0}, {"name": "test", "exit": 0}]})
    );

    let before = w.dump();
    let out = w.verdict(
        Some(S_AUD),
        &card,
        &[
            "--head",
            &w.head,
            "--base",
            &w.base,
            "--verdict",
            "pass",
            "--tier",
            "1",
            "--evidence",
            "a step with no exit code",
            "--receipt-head",
            &w.head,
            "--receipt-step",
            "test",
        ],
    );
    w.refused(&out, "NAME=EXIT", &before);
    // A step without a receipt head is no receipt either.
    let out = w.verdict(
        Some(S_AUD),
        &card,
        &[
            "--head",
            &w.head,
            "--base",
            &w.base,
            "--verdict",
            "pass",
            "--tier",
            "1",
            "--evidence",
            "a step with no receipt head",
            "--receipt-step",
            "test=0",
        ],
    );
    assert_ne!(out.status.code(), Some(0));
    assert_eq!(w.dump(), before);
}
