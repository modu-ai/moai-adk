//! `moai card criterion-add` (SPEC-20261119-card-criterion-add): an acceptance criterion is added
//! to a card that exists, so a card made without one can be closed. The closing rules are
//! `tests/ledger_workflow.rs` and `tests/card_work_cli.rs`; here the action, its answers, its
//! refusal kinds and what it leaves in the ledger.
//!
//! Every command runs against a temporary HOME; the real `~/.moai` is not read or written. The
//! holder is the session id in the environment (`CLAUDE_CODE_SESSION_ID`). `card pr` reads the
//! pull request through a stub `gh` (`MOAI_FORGE_GH_BIN`); no test reaches the network.
#![cfg(unix)]

mod common;

use common::home::TempHome;
use moai_cli::ledger;
use rusqlite::Connection;
use serde_json::{json, Value};
use std::fs;
use std::path::PathBuf;
use std::process::{Output, Stdio};

const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
const S_A: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
const MERGE: &str = "0123456789abcdef0123456789abcdef01234567";
const HEAD: &str = "89abcdef0123456789abcdef0123456789abcdef";

struct World {
    home: TempHome,
    _host: tempfile::TempDir,
    dir: PathBuf,
    project: String,
    gh: PathBuf,
}

fn world() -> World {
    let home = TempHome::new();
    let project = home.project("demo");
    ledger::register(home.path(), "demo").unwrap();
    let host = tempfile::Builder::new()
        .prefix("moai-test-host-")
        .tempdir()
        .unwrap();
    let dir = host.path().to_path_buf();
    let script = format!(
        "#!/bin/sh\nurl=''\nfor a in \"$@\"; do case \"$a\" in https://*) url=\"$a\";; esac; done\n\
         case \"$url\" in\n\
           */pulls/*) n=${{url##*/pulls/}}; if [ -f '{dir}/pull-'$n'.json' ]; then printf 'HTTP/2.0 200 OK\\r\\n\\r\\n'; cat '{dir}/pull-'$n'.json'; exit 0; fi;;\n\
         esac\n\
         printf 'HTTP/2.0 404 Not Found\\r\\n\\r\\n{{}}'; exit 1\n",
        dir = dir.display()
    );
    let gh = dir.join("gh-stub");
    common::write_executable(&gh, &script);
    World {
        home,
        _host: host,
        dir,
        project: project.to_str().unwrap().to_string(),
        gh,
    }
}

impl World {
    /// `moai card <args> --project <project>`, as `session`.
    fn card(&self, session: Option<&str>, args: &[&str]) -> Output {
        let mut all = vec!["card"];
        all.extend_from_slice(args);
        all.extend_from_slice(&["--project", &self.project]);
        let mut cmd = self.home.command(&all);
        cmd.env("MOAI_FORGE_GH_BIN", &self.gh);
        if let Some(id) = session {
            cmd.env(SESSION_VAR, id);
        }
        cmd.stdin(Stdio::null()).output().expect("run moai")
    }

    fn host_pr(&self, number: u64) {
        let pull = json!({
            "number": number, "state": "closed", "merged": true,
            "merged_at": "2026-11-17T00:00:00Z",
            "merge_commit_sha": MERGE,
            "head": {"sha": HEAD},
        });
        fs::write(
            self.dir.join(format!("pull-{number}.json")),
            pull.to_string(),
        )
        .unwrap();
    }

    /// A `ready` card with the given criteria; its name.
    fn ready(&self, criteria: &[&str]) -> String {
        self.ready_in(criteria, "apps/bo/")
    }

    /// The same, with its own write scope (two cards taken at once must not overlap).
    fn ready_in(&self, criteria: &[&str], scope: &str) -> String {
        let mut args = vec!["add", "--title", "work", "--scope", scope];
        for ac in criteria {
            args.extend_from_slice(&["--ac", ac]);
        }
        let name = result(&self.card(None, &args))["name"]
            .as_str()
            .unwrap()
            .to_string();
        result(&self.card(None, &["ready", &name]));
        name
    }

    fn take(&self, session: &str, card: &str) {
        result(&self.card(
            Some(session),
            &[
                "take",
                card,
                "--worktree",
                "/work/tree",
                "--branch",
                "card/work",
            ],
        ));
    }

    fn sql(&self, statement: &str) {
        let mut l = ledger::open_write(self.home.path(), "demo").unwrap();
        l.with_write_tx(|tx, _| {
            tx.execute_batch(statement)?;
            Ok(())
        })
        .unwrap();
    }

    fn dump(&self) -> Vec<String> {
        let l = ledger::open_read_only(self.home.path(), "demo").unwrap();
        dump_tables(
            l.conn(),
            &[
                "card",
                "card_ac",
                "card_note",
                "card_link",
                "reservation",
                "reservation_scope",
                "event",
            ],
        )
    }

    fn rows(&self, sql: &str) -> Vec<Value> {
        let l = ledger::open_read_only(self.home.path(), "demo").unwrap();
        let mut st = l.conn().prepare(sql).unwrap();
        let n = st.column_count();
        st.query_map([], |r| {
            Ok(Value::Array(
                (0..n)
                    .map(|i| match r.get_ref(i).unwrap() {
                        rusqlite::types::ValueRef::Integer(v) => json!(v),
                        rusqlite::types::ValueRef::Text(t) => {
                            json!(String::from_utf8_lossy(t).to_string())
                        }
                        _ => Value::Null,
                    })
                    .collect(),
            ))
        })
        .unwrap()
        .map(Result::unwrap)
        .collect()
    }

    fn show(&self, card: &str) -> Value {
        result(&self.card(None, &["show", card]))
    }
}

fn dump_tables(conn: &Connection, tables: &[&str]) -> Vec<String> {
    let mut out = Vec::new();
    for table in tables {
        let mut st = conn
            .prepare(&format!("SELECT * FROM {table} ORDER BY 1, 2"))
            .unwrap();
        let n = st.column_count();
        let lines: Vec<String> = st
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
        out.extend(lines);
    }
    out
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
    assert_eq!(v["operation"], json!("card"), "{v}");
    v["result"].clone()
}

fn refusal(out: &Output) -> (String, String) {
    assert_eq!(
        out.status.code(),
        Some(1),
        "{} {}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    let v = envelope(out);
    assert_eq!(v["ok"], json!(false), "{v}");
    (
        v["error"]["kind"].as_str().unwrap().to_string(),
        v["error"]["message"].as_str().unwrap().to_string(),
    )
}

fn criteria_of(show: &Value) -> Vec<(i64, String, String)> {
    show["criteria"]
        .as_array()
        .unwrap()
        .iter()
        .map(|c| {
            (
                c["seq"].as_i64().unwrap(),
                c["text"].as_str().unwrap().to_string(),
                c["status"].as_str().unwrap().to_string(),
            )
        })
        .collect()
}

#[test]
fn criteria_are_appended_after_the_highest_number_with_one_event() {
    let w = world();
    let card = w.ready(&["first", "second"]);
    // Criterion 2 is withdrawn: its number is not given again.
    w.take(S_A, &card);
    result(&w.card(
        Some(S_A),
        &[
            "criterion",
            &card,
            "2",
            "--generation",
            "1",
            "--withdraw",
            "--reason",
            "no longer applies",
        ],
    ));
    let rev_before = w.show(&card)["rev"].as_i64().unwrap();
    let events_before = w.rows("SELECT count(*) FROM event")[0][0].as_i64().unwrap();

    let answer = result(&w.card(
        None,
        &[
            "criterion-add",
            &card,
            "--ac",
            "  a  ",
            "--ac",
            "b",
            "--by",
            "leader",
        ],
    ));
    assert_eq!(answer["action"], json!("criterion-add"), "{answer}");
    assert_eq!(answer["name"], json!(card), "{answer}");
    assert_eq!(answer["by"], json!("leader"), "{answer}");
    assert_eq!(answer["added"], json!([3, 4]), "{answer}");
    assert_eq!(answer["criteria"], json!(4), "{answer}");
    assert_eq!(answer["events"], json!(1), "{answer}");
    assert_eq!(answer["writes"], json!(true), "{answer}");

    let shown = w.show(&card);
    assert_eq!(
        criteria_of(&shown),
        vec![
            (1, "first".to_string(), "open".to_string()),
            (2, "second".to_string(), "withdrawn".to_string()),
            (3, "a".to_string(), "open".to_string()),
            (4, "b".to_string(), "open".to_string()),
        ]
    );
    assert!(shown["rev"].as_i64().unwrap() > rev_before, "{shown}");
    let events = w.rows(
        "SELECT actor, payload FROM event WHERE kind = 'card.criterion_add' ORDER BY event_id",
    );
    assert_eq!(events.len(), 1, "{events:?}");
    assert_eq!(events[0][0], json!("leader"));
    let payload: Value = serde_json::from_str(events[0][1].as_str().unwrap()).unwrap();
    assert_eq!(payload["first"], json!(3), "{payload}");
    assert_eq!(payload["count"], json!(2), "{payload}");
    assert_eq!(
        w.rows("SELECT count(*) FROM event")[0][0].as_i64().unwrap(),
        events_before + 1
    );

    // A card with no criterion gets number 1, and the actor defaults to `user`.
    let bare = w.ready(&[]);
    let answer = result(&w.card(None, &["criterion-add", &bare, "--ac", "only"]));
    assert_eq!(answer["added"], json!([1]), "{answer}");
    assert_eq!(answer["by"], json!("user"), "{answer}");
    assert_eq!(
        criteria_of(&w.show(&bare)),
        vec![(1, "only".to_string(), "open".to_string())]
    );
}

#[test]
fn a_card_made_without_a_criterion_can_be_closed_once_one_is_added_and_met() {
    let w = world();
    w.host_pr(7);
    let card = w.ready(&[]);
    w.take(S_A, &card);
    result(&w.card(
        Some(S_A),
        &["pr", &card, "7", "--generation", "1", "--repo", "octo/demo"],
    ));
    result(&w.card(Some(S_A), &["review", &card, "--generation", "1"]));

    // Zero criteria: closing is refused, as before.
    let (kind, message) = refusal(&w.card(Some(S_A), &["done", &card, "--generation", "1"]));
    assert_eq!(kind, "invalid_input", "{message}");
    assert!(
        message.ends_with("the card has no acceptance criterion that is met"),
        "{message}"
    );

    // A criterion is added; it is open, so nothing is met yet and closing is still refused.
    result(&w.card(
        None,
        &["criterion-add", &card, "--ac", "the fix is on the tip"],
    ));
    let (_, message) = refusal(&w.card(Some(S_A), &["done", &card, "--generation", "1"]));
    assert!(
        message.ends_with("the card has no acceptance criterion that is met"),
        "{message}"
    );

    // Met, it would close the card; a second criterion added now is open, and that refuses it.
    let met = |n: &str| {
        result(&w.card(
            Some(S_A),
            &[
                "criterion",
                &card,
                n,
                "--generation",
                "1",
                "--met",
                "--evidence",
                "cargo test x: ok",
            ],
        ))
    };
    met("1");
    result(&w.card(None, &["criterion-add", &card, "--ac", "the doc says it"]));
    let (_, message) = refusal(&w.card(Some(S_A), &["done", &card, "--generation", "1"]));
    assert!(
        message.ends_with("an acceptance criterion is still open"),
        "{message}"
    );

    // Both met, the card closes.
    met("2");
    let done = result(&w.card(Some(S_A), &["done", &card, "--generation", "1"]));
    assert_eq!(done["state"], json!("done"), "{done}");
}

#[test]
fn a_done_card_and_a_missing_card_are_refused_and_nothing_is_written() {
    let w = world();
    let card = w.ready(&["one"]);
    w.sql(
        "INSERT INTO card (title, state, created_by, created_at, updated_at, closed_at) \
         VALUES ('closed', 'done', 'user', 1, 1, 1)",
    );
    let done = "C-2";
    let before = w.dump();

    let (kind, message) = refusal(&w.card(None, &["criterion-add", done, "--ac", "late"]));
    assert_eq!(kind, "conflict", "{message}");
    assert!(message.contains("C-2 is `done`"), "{message}");

    let (kind, message) = refusal(&w.card(None, &["criterion-add", "C-99", "--ac", "x"]));
    assert_eq!(kind, "not_found", "{message}");
    assert!(message.contains("C-99"), "{message}");

    // A name that is not a card reference, and a host number (`#12`, which needs a lookup),
    // are refused before the ledger is opened.
    let (kind, _) = refusal(&w.card(None, &["criterion-add", "card-1", "--ac", "x"]));
    assert_eq!(kind, "invalid_input");
    let (kind, message) = refusal(&w.card(None, &["criterion-add", "#12", "--ac", "x"]));
    assert_eq!(kind, "invalid_input", "{message}");
    assert!(message.contains("C-number"), "{message}");

    assert_eq!(w.dump(), before);
    // The card that is not closed still takes one.
    result(&w.card(None, &["criterion-add", &card, "--ac", "two"]));
}

#[test]
fn a_text_that_cannot_be_stored_refuses_the_whole_request() {
    let w = world();
    let card = w.ready(&["one"]);
    let before = w.dump();

    for (args, says) in [
        (vec!["--ac", ""], "criterion 1 is empty"),
        (vec!["--ac", "  "], "criterion 1 is empty"),
        (vec!["--ac", "two\nlines"], "criterion 1 must be one line"),
        (
            vec!["--ac", "good", "--ac", "bad\nline", "--ac", "good too"],
            "criterion 2 must be one line",
        ),
        (vec!["--ac", "good", "--by", " "], "--by must be one line"),
    ] {
        let mut all = vec!["criterion-add", card.as_str()];
        all.extend(args.iter().copied());
        let (kind, message) = refusal(&w.card(None, &all));
        assert_eq!(kind, "invalid_input", "{args:?}: {message}");
        assert!(message.contains(says), "{args:?}: {message}");
    }
    let long_by = "b".repeat(81);
    let (kind, message) = refusal(&w.card(
        None,
        &["criterion-add", &card, "--ac", "good", "--by", &long_by],
    ));
    assert_eq!(kind, "invalid_input", "{message}");
    assert_eq!(w.dump(), before, "nothing was written, the good text too");

    // 80 characters are accepted.
    let ok_by = "b".repeat(80);
    result(&w.card(
        None,
        &["criterion-add", &card, "--ac", "good", "--by", &ok_by],
    ));

    // No text at all is a missing option: the parser refuses it (usage, exit 2).
    let out = w.card(None, &["criterion-add", &card]);
    assert_eq!(out.status.code(), Some(2));
    assert_eq!(envelope(&out)["error"]["kind"], json!("usage"));
}

#[test]
fn adding_to_a_held_card_leaves_its_reservation_alone() {
    let w = world();
    let active = w.ready(&["one"]);
    w.take(S_A, &active);
    let review = w.ready_in(&["one"], "apps/web/");
    w.take(S_A, &review);
    result(&w.card(
        Some(S_A),
        &[
            "criterion",
            &review,
            "1",
            "--generation",
            "1",
            "--met",
            "--evidence",
            "x",
        ],
    ));
    w.host_pr(7);
    result(&w.card(
        Some(S_A),
        &[
            "pr",
            &review,
            "7",
            "--generation",
            "1",
            "--repo",
            "octo/demo",
        ],
    ));
    result(&w.card(Some(S_A), &["review", &review, "--generation", "1"]));

    let reservation = || {
        w.rows(
            "SELECT card_id, generation, holder, status, acquired_at, lease_until, worktree, \
             branch FROM reservation ORDER BY reservation_id",
        )
    };
    let scope = || {
        w.rows("SELECT reservation_id, pattern FROM reservation_scope ORDER BY reservation_id, pattern")
    };
    let (reservations, scopes) = (reservation(), scope());
    let check = |card: &str| {
        envelope(&w.card(
            Some(S_A),
            &["check", card, "--generation", "1", "--path", "apps/bo/x.rs"],
        ))
    };
    let (check_active, check_review) = (check(&active), check(&review));

    for card in [&active, &review] {
        result(&w.card(None, &["criterion-add", card, "--ac", "added later"]));
    }
    assert_eq!(reservation(), reservations);
    assert_eq!(scope(), scopes);
    assert_eq!(
        (check(&active), check(&review)),
        (check_active, check_review)
    );
    assert_eq!(w.show(&active)["state"], json!("active"));
    assert_eq!(w.show(&review)["state"], json!("review"));
    // The review card now has an open criterion: it cannot close until that one is met.
    let (_, message) = refusal(&w.card(Some(S_A), &["done", &review, "--generation", "1"]));
    assert!(
        message.ends_with("an acceptance criterion is still open"),
        "{message}"
    );
}

#[test]
fn two_adds_at_once_number_every_criterion_once() {
    let w = world();
    let card = w.ready(&["one"]);
    let spawn = |text: &str| {
        w.home
            .command(&[
                "card",
                "criterion-add",
                &card,
                "--ac",
                text,
                "--by",
                "leader",
                "--project",
                &w.project,
            ])
            .stdin(Stdio::null())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .spawn()
            .expect("run moai")
    };
    let children: Vec<_> = (0..4).map(|i| spawn(&format!("added {i}"))).collect();
    for child in children {
        let out = child.wait_with_output().unwrap();
        result(&out);
    }
    let numbers: Vec<i64> = criteria_of(&w.show(&card)).iter().map(|c| c.0).collect();
    assert_eq!(numbers, vec![1, 2, 3, 4, 5]);
    let mut texts: Vec<String> = criteria_of(&w.show(&card))
        .into_iter()
        .skip(1)
        .map(|c| c.1)
        .collect();
    texts.sort();
    assert_eq!(texts, ["added 0", "added 1", "added 2", "added 3"]);
}

#[test]
fn the_help_and_the_ops_doc_name_the_rule() {
    let w = world();
    let out = w
        .home
        .command(&["card", "criterion-add", "--help"])
        .output()
        .expect("run moai");
    let help = String::from_utf8_lossy(&out.stdout);
    // The help is wrapped at the terminal width: compare words, not line breaks.
    let help = help.split_whitespace().collect::<Vec<_>>().join(" ");
    assert!(
        help.starts_with("Writes the ledger."),
        "first line says it writes: {help}"
    );
    for phrase in [
        "criteria are only added",
        "needs no generation",
        "a `done` card is refused",
    ] {
        assert!(help.contains(phrase), "{phrase} missing: {help}");
    }
    let doc = std::fs::read_to_string(
        std::path::Path::new(env!("CARGO_MANIFEST_DIR"))
            .join("docs/design-review/appendix/operations/card.md"),
    )
    .unwrap();
    let row = doc
        .lines()
        .find(|line| line.starts_with("| `card criterion-add"))
        .expect("the card criterion-add row");
    for phrase in [
        "`workflow::add_criteria`",
        "더하기만",
        "세대가 필요 없",
        "`done` 카드는 거절",
    ] {
        assert!(row.contains(phrase), "{phrase} missing in the row: {row}");
    }
}
