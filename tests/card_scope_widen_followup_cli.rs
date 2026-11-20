//! `moai card scope` on a held card, the follow-ups of the C-94 audit
//! (SPEC-20261119-card-scope-widen-followup, card C-129): line separators in a reason or a scope
//! entry, the reason kept without its surrounding blanks, a case variant of a pattern the
//! reservation already has (slice A), and the generation a widening is made at (slice B).
//!
//! Every command runs against a temporary HOME; the real `~/.moai` is not read or written. The
//! holder is the session id in the environment (`CLAUDE_CODE_SESSION_ID`).
#![cfg(unix)]

mod common;

use common::home::TempHome;
use moai_cli::ledger;
use rusqlite::Connection;
use serde_json::{json, Value};
use std::process::{Output, Stdio};

const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
const S_A: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
const WHY: &str = "the shared type lives under apps/web and the task changes it";

struct World {
    home: TempHome,
    project: String,
}

fn world() -> World {
    let home = TempHome::new();
    let project = home.project("demo");
    ledger::register(home.path(), "demo").unwrap();
    World {
        home,
        project: project.to_str().unwrap().to_string(),
    }
}

impl World {
    /// `moai card <args> --project <project>`, as `session`.
    fn card(&self, session: Option<&str>, args: &[&str]) -> Output {
        let mut all = vec!["card"];
        all.extend_from_slice(args);
        all.extend_from_slice(&["--project", &self.project]);
        let mut cmd = self.home.command(&all);
        if let Some(id) = session {
            cmd.env(SESSION_VAR, id);
        }
        cmd.stdin(Stdio::null()).output().expect("run moai")
    }

    /// A ready card with one open criterion and `scope`, taken by `session` (generation 1).
    fn held(&self, session: &str, title: &str, scope: &[&str]) -> String {
        let mut args = vec!["add", "--title", title, "--ac", "one"];
        for entry in scope {
            args.push("--scope");
            args.push(entry);
        }
        let name = result(&self.card(None, &args))["name"]
            .as_str()
            .unwrap()
            .to_string();
        result(&self.card(None, &["ready", &name]));
        result(&self.card(
            Some(session),
            &[
                "take",
                &name,
                "--worktree",
                "/work/tree",
                "--branch",
                "card/x",
            ],
        ));
        name
    }

    /// `card scope <card> --scope… --by leader [--reason …]`.
    fn widen(&self, card: &str, scope: &[&str], reason: Option<&str>) -> Output {
        self.widen_at(card, None, scope, reason)
    }

    /// `card scope <card> [--generation <n>] --scope… --by leader [--reason …]`.
    fn widen_at(
        &self,
        card: &str,
        generation: Option<&str>,
        scope: &[&str],
        reason: Option<&str>,
    ) -> Output {
        let mut args = vec!["scope", card];
        if let Some(g) = generation {
            args.extend_from_slice(&["--generation", g]);
        }
        for entry in scope {
            args.push("--scope");
            args.push(entry);
        }
        args.extend_from_slice(&["--by", "leader"]);
        if let Some(reason) = reason {
            args.extend_from_slice(&["--reason", reason]);
        }
        self.card(None, &args)
    }

    /// Every row of the ledger's tables, as text.
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

    fn rows(&self, sql: &str) -> Vec<Vec<Value>> {
        let l = ledger::open_read_only(self.home.path(), "demo").unwrap();
        let mut st = l.conn().prepare(sql).unwrap();
        let n = st.column_count();
        st.query_map([], |r| {
            Ok((0..n)
                .map(|i| match r.get_ref(i).unwrap() {
                    rusqlite::types::ValueRef::Integer(v) => json!(v),
                    rusqlite::types::ValueRef::Text(t) => {
                        json!(String::from_utf8_lossy(t).to_string())
                    }
                    _ => Value::Null,
                })
                .collect())
        })
        .unwrap()
        .map(Result::unwrap)
        .collect()
    }

    /// The patterns of the reservation of `card_id`, sorted.
    fn patterns(&self, card_id: i64) -> Vec<String> {
        self.rows(&format!(
            "SELECT s.pattern FROM reservation_scope s JOIN reservation r \
             ON r.reservation_id = s.reservation_id WHERE r.card_id = {card_id} \
             ORDER BY s.pattern"
        ))
        .into_iter()
        .map(|row| row[0].as_str().unwrap().to_string())
        .collect()
    }

    /// The patterns of the reservation of `card_id` at `generation`, sorted.
    fn patterns_at(&self, card_id: i64, generation: i64) -> Vec<String> {
        self.rows(&format!(
            "SELECT s.pattern FROM reservation_scope s JOIN reservation r \
             ON r.reservation_id = s.reservation_id WHERE r.card_id = {card_id} \
             AND r.generation = {generation} ORDER BY s.pattern"
        ))
        .into_iter()
        .map(|row| row[0].as_str().unwrap().to_string())
        .collect()
    }

    /// The `card.scope` events as payloads.
    fn scope_events(&self) -> Vec<Value> {
        self.rows("SELECT payload FROM event WHERE kind = 'card.scope' ORDER BY event_id")
            .into_iter()
            .map(|row| serde_json::from_str(row[0].as_str().unwrap()).unwrap())
            .collect()
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

/// AC-1 (REQ-1): U+2028 and U+2029 are not control characters (category Zl and Zp), `str::trim`
/// strips them at the ends: the check reads the raw text, so an inner one, a leading one and a
/// trailing one are all refused, for a reason and for a scope entry, on every command that takes
/// them, and nothing is written.
#[test]
fn line_separators_are_refused_in_a_reason_and_a_scope_entry() {
    let w = world();
    let held = w.held(S_A, "held", &["apps/bo/"]);
    let before = w.dump();
    for sep in ['\u{2028}', '\u{2029}'] {
        for reason in [
            format!("first{sep}second"),
            format!("{sep}leading"),
            format!("trailing{sep}"),
        ] {
            let (kind, message) =
                refusal(&w.widen(&held, &["apps/bo/", "apps/web/"], Some(&reason)));
            assert_eq!(kind, "invalid_input", "{reason:?}: {message}");
            assert!(message.contains("--reason"), "{message}");
        }
        for entry in [
            format!("apps/{sep}web/"),
            format!("{sep}apps/web/"),
            format!("apps/web/{sep}"),
        ] {
            let (kind, message) = refusal(&w.widen(&held, &["apps/bo/", &entry], Some(WHY)));
            assert_eq!(kind, "invalid_input", "{entry:?}: {message}");
            assert!(message.contains("scope"), "{message}");
        }
    }
    // `add` and `take --scope` use the same entry check.
    let (kind, _) = refusal(&w.card(
        None,
        &[
            "add",
            "--title",
            "t",
            "--ac",
            "one",
            "--scope",
            "src/\u{2028}x/",
        ],
    ));
    assert_eq!(kind, "invalid_input");
    let ready = result(&w.card(
        None,
        &["add", "--title", "ready", "--ac", "one", "--scope", "docs/"],
    ))["name"]
        .as_str()
        .unwrap()
        .to_string();
    result(&w.card(None, &["ready", &ready]));
    let (kind, _) = refusal(&w.card(
        Some(S_A),
        &[
            "take",
            &ready,
            "--worktree",
            "/work/t2",
            "--branch",
            "card/y",
            "--scope",
            "docs/\u{2029}",
        ],
    ));
    assert_eq!(kind, "invalid_input");
    let after = w.dump();
    // Only the `ready` card and its events were added by the setup above; the held card is as it was.
    let held_rows = |rows: &[String]| -> Vec<String> {
        rows.iter()
            .filter(|l| l.starts_with("reservation") || l.starts_with("reservation_scope"))
            .cloned()
            .collect()
    };
    assert_eq!(held_rows(&before), held_rows(&after));
    assert!(w.scope_events().is_empty(), "no widening was recorded");
}

/// AC-1, the reclaim half: `card reclaim --reason` shares the reason check. The lease of a fresh
/// reservation has not ended, so the command refuses for that reason too; the test pins that the
/// separator is named first, before any ledger read.
#[test]
fn a_reclaim_reason_with_a_line_separator_is_invalid_input() {
    let w = world();
    let held = w.held(S_A, "held", &["apps/bo/"]);
    let (kind, message) = refusal(&w.card(
        Some(S_A),
        &[
            "reclaim",
            &held,
            "--generation",
            "1",
            "--by",
            "leader",
            "--reason",
            "the holder is gone\u{2028}really",
        ],
    ));
    assert_eq!(kind, "invalid_input", "{message}");
    assert!(message.contains("--reason"), "{message}");
}

/// AC-2 (REQ-2): the reason is stored without its surrounding blanks, in the answer and the event.
#[test]
fn a_reason_is_kept_without_its_surrounding_blanks() {
    let w = world();
    let held = w.held(S_A, "held", &["apps/bo/"]);
    // Spaces only: a tab is a control character and is refused on its own.
    let padded = format!("  \u{3000} {WHY} \u{a0}  ");
    let answer = result(&w.widen(&held, &["apps/bo/", "apps/web/"], Some(&padded)));
    assert_eq!(answer["reason"], json!(WHY), "{answer}");
    let events = w.scope_events();
    assert_eq!(events.len(), 1, "{events:?}");
    assert_eq!(events[0]["reason"], json!(WHY), "{:?}", events[0]);
}

/// AC-3 (REQ-3): a pattern that differs from one the reservation has, or from another added
/// pattern, only by letter case is not a new pattern. The refusal names both spellings.
#[test]
fn a_case_variant_is_not_a_new_pattern() {
    let w = world();
    let held = w.held(S_A, "held", &["apps/bo/", "apps/web/"]);
    let before = w.dump();

    // A variant of a pattern the reservation has.
    let (kind, message) =
        refusal(&w.widen(&held, &["apps/bo/", "apps/web/", "Apps/Web/"], Some(WHY)));
    assert_eq!(kind, "invalid_input", "{message}");
    assert!(
        message.contains("Apps/Web/") && message.contains("apps/web/"),
        "{message}"
    );

    // Two variants of each other, both new.
    let (kind, message) = refusal(&w.widen(
        &held,
        &["apps/bo/", "apps/web/", "libs/Util/", "libs/util/"],
        Some(WHY),
    ));
    assert_eq!(kind, "invalid_input", "{message}");
    assert!(
        message.contains("libs/Util/") && message.contains("libs/util/"),
        "{message}"
    );
    assert_eq!(w.dump(), before, "a refusal writes nothing");
    assert_eq!(w.patterns(1), vec!["apps/bo/", "apps/web/"]);

    // The same letters in another folder are a different path: still added.
    let answer = result(&w.widen(
        &held,
        &["apps/bo/", "apps/web/", "other/apps/web/"],
        Some(WHY),
    ));
    assert_eq!(answer["added"], json!(["other/apps/web/"]), "{answer}");

    // A variant of a pattern added just now is refused as well.
    let (kind, message) = refusal(&w.widen(
        &held,
        &[
            "apps/bo/",
            "apps/web/",
            "other/apps/web/",
            "Other/Apps/Web/",
        ],
        Some(WHY),
    ));
    assert_eq!(kind, "invalid_input", "{message}");
    assert_eq!(
        w.patterns(1),
        vec!["apps/bo/", "apps/web/", "other/apps/web/"]
    );
}

/// AC-5 (REQ-4, REQ-5): `--generation <n>` widens only a reservation that is at generation `n`.
/// A lane that read generation 1 and a leader that acts after the card was released and taken
/// again (generation 2) do not act on different screens: the stale number is a `conflict` that
/// names the generation the card has, and nothing is written. A card nobody holds has no
/// generation to match. Without `--generation` the command is as it was.
#[test]
fn a_stale_generation_is_refused() {
    let w = world();
    let held = w.held(S_A, "held", &["apps/bo/"]);
    // Released and taken again: the reservation is now at generation 2.
    result(&w.card(Some(S_A), &["release", &held, "--generation", "1"]));
    result(&w.card(
        Some(S_A),
        &[
            "take",
            &held,
            "--worktree",
            "/work/tree",
            "--branch",
            "card/x",
        ],
    ));
    let before = w.dump();

    for stale in ["1", "3"] {
        let (kind, message) =
            refusal(&w.widen_at(&held, Some(stale), &["apps/bo/", "apps/web/"], Some(WHY)));
        assert_eq!(kind, "conflict", "{message}");
        assert!(
            message.contains("generation 2") && message.contains(stale),
            "names the generation the card has and the one given: {message}"
        );
    }
    assert_eq!(w.dump(), before, "a refusal writes nothing");
    assert!(w.scope_events().is_empty());

    // The right generation widens, and the answer and the event keep it.
    let answer = result(&w.widen_at(&held, Some("2"), &["apps/bo/", "apps/web/"], Some(WHY)));
    assert_eq!(answer["generation"], json!(2), "{answer}");
    assert_eq!(answer["added"], json!(["apps/web/"]), "{answer}");
    assert_eq!(w.patterns_at(1, 2), vec!["apps/bo/", "apps/web/"]);
    assert_eq!(
        w.patterns_at(1, 1),
        vec!["apps/bo/"],
        "the released one is as it was"
    );

    // Without `--generation` nothing is asked: the command is as before.
    let answer = result(&w.widen(&held, &["apps/bo/", "apps/web/", "docs/"], Some(WHY)));
    assert_eq!(answer["added"], json!(["docs/"]), "{answer}");

    // A card nobody holds has no generation to match: `conflict`, nothing written.
    let ready = result(&w.card(
        None,
        &["add", "--title", "ready", "--ac", "one", "--scope", "libs/"],
    ))["name"]
        .as_str()
        .unwrap()
        .to_string();
    result(&w.card(None, &["ready", &ready]));
    let before = w.dump();
    let (kind, message) = refusal(&w.widen_at(&ready, Some("1"), &["libs/", "libs/x/"], None));
    assert_eq!(kind, "conflict", "{message}");
    assert!(message.contains("not held"), "{message}");
    assert_eq!(w.dump(), before);
}

/// A non-numeric or non-positive `--generation` is not a generation.
#[test]
fn a_generation_that_is_not_a_positive_number_is_a_usage_error() {
    let w = world();
    let held = w.held(S_A, "held", &["apps/bo/"]);
    let before = w.dump();
    for bad in ["0", "-1", "x", ""] {
        let out = w.widen_at(&held, Some(bad), &["apps/bo/", "apps/web/"], Some(WHY));
        // clap refuses the value before the command runs: exit 2, a `usage` envelope.
        assert_eq!(out.status.code(), Some(2), "`{bad}`");
        let v = envelope(&out);
        assert_eq!(v["ok"], json!(false), "`{bad}`: {v}");
        assert_eq!(v["error"]["kind"], json!("usage"), "`{bad}`: {v}");
    }
    assert_eq!(w.dump(), before);
}

/// Audit note N-1 of slice A: the fold is `to_lowercase`, so a non-ASCII variant is a variant too.
#[test]
fn a_non_ascii_case_variant_is_not_a_new_pattern() {
    let w = world();
    let held = w.held(S_A, "held", &["caf\u{e9}/"]);
    let before = w.dump();
    let (kind, message) = refusal(&w.widen(&held, &["caf\u{e9}/", "CAF\u{c9}/"], Some(WHY)));
    assert_eq!(kind, "invalid_input", "{message}");
    assert!(
        message.contains("CAF\u{c9}/") && message.contains("caf\u{e9}/"),
        "{message}"
    );
    assert_eq!(w.dump(), before);
}

/// Audit note N-2 of slice A: the reason is trimmed before it is counted, so 500 characters with
/// blanks around are kept as 500, and 501 characters with blanks around are refused.
#[test]
fn a_reason_of_500_characters_with_blanks_is_kept_and_501_is_refused() {
    let w = world();
    let held = w.held(S_A, "held", &["apps/bo/"]);
    let five_hundred = "x".repeat(500);
    let answer = result(&w.widen(
        &held,
        &["apps/bo/", "apps/web/"],
        Some(&format!("  {five_hundred}  ")),
    ));
    assert_eq!(answer["reason"], json!(five_hundred), "{answer}");
    let events = w.scope_events();
    assert_eq!(events[0]["reason"], json!(five_hundred));

    let before = w.dump();
    let (kind, message) = refusal(&w.widen(
        &held,
        &["apps/bo/", "apps/web/", "docs/"],
        Some(&format!("  {}  ", "x".repeat(501))),
    ));
    assert_eq!(kind, "invalid_input", "{message}");
    assert!(message.contains("--reason"), "{message}");
    assert_eq!(w.dump(), before, "a refusal writes nothing");
}
