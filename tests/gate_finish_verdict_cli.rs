//! `moai card finish` reads the same verdict of the head that `card land` reads
//! (SPEC-20261125-integration-guard-install REQ-5 and REQ-6): a card that changes the repository
//! closes on a landed head only when that exact head has an independent PASS, in observe mode as
//! well as in enforce mode, and the evidence of the other kinds of card (a report, a note, a
//! merged pull request link on a profile with no integration branch, a research card) is read
//! as before. Every test runs on a temporary home and a temporary repository; nothing here
//! touches a real ledger.
//!
//! The repository: `main` holds `init` and `head` (the card's change, signed by the lane
//! session); `integ` points at `tip`, a child of `head`, so `head` is already landed.

#![cfg(unix)]

mod common;

use common::home::TempHome;
use moai_cli::ledger;
use rusqlite::Connection;
use serde_json::{json, Value};
use std::cell::Cell;
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Output, Stdio};

const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
const S_A: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
const S_AUD: &str = "cccccccc-0000-4000-8000-00000000000c";
const S_LEAD: &str = "dddddddd-0000-4000-8000-00000000000d";
const KEY: &str = "0123456789ab";

fn profile(mode: &str, integration: bool) -> String {
    let flow = if integration {
        format!(
            ", \"work_flow\": {{\"flow\": \"next-flow\", \"integration_branch\": \"integ\", \"land_verdict\": \"{mode}\"}}"
        )
    } else {
        String::new()
    };
    format!(
        "{{\"schema_version\": 1, \"project_id\": \"demo\", \"purpose\": \"fixture\", \"tools\": [\"rust\"], \"development_methods\": [\"sdd\"], \"hosts\": [\"claude\"]{flow}}}"
    )
}

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

struct World {
    home: TempHome,
    project: PathBuf,
    p: String,
    /// The landed head: inside `integ`, a child of `init`.
    head: String,
    /// The head of another change, also inside `integ`.
    other: String,
    events: Cell<u32>,
}

impl World {
    fn new(mode: &str, integration: bool) -> Self {
        let home = TempHome::new();
        let project = home.project("demo");
        ledger::register(home.path(), "demo").unwrap();
        git(&project, &["init", "-q", "-b", "main"]);
        fs::write(project.join("a.txt"), "a\n").unwrap();
        git(&project, &["add", "a.txt"]);
        git(&project, &["commit", "-q", "-m", "init"]);
        let signed = |what: &str| format!("{what}\n\nMoai-Session: {S_A}:{KEY}");
        fs::write(project.join("b.txt"), "b\n").unwrap();
        git(&project, &["add", "b.txt"]);
        git(
            &project,
            &["commit", "-q", "-m", &signed("the card's change")],
        );
        let head = git(&project, &["rev-parse", "HEAD"]);
        fs::write(project.join("c.txt"), "c\n").unwrap();
        git(&project, &["add", "c.txt"]);
        git(&project, &["commit", "-q", "-m", &signed("another change")]);
        let other = git(&project, &["rev-parse", "HEAD"]);
        git(&project, &["branch", "integ", "HEAD"]);
        let world = World {
            p: project.to_str().unwrap().to_string(),
            home,
            project,
            head,
            other,
            events: Cell::new(0),
        };
        world.write_profile(mode, integration);
        world
    }

    fn write_profile(&self, mode: &str, integration: bool) {
        let cfg = moai_cli::state_root::path(&self.project, "project.yaml");
        fs::create_dir_all(cfg.parent().unwrap()).unwrap();
        fs::write(&cfg, profile(mode, integration)).unwrap();
    }

    fn card(&self, session: Option<&str>, args: &[&str]) -> Output {
        let mut all = vec!["card"];
        all.extend_from_slice(args);
        all.extend_from_slice(&["--project", &self.p]);
        let mut cmd = self.home.command(&all);
        cmd.env("LC_ALL", "C").stdin(Stdio::null());
        if let Some(id) = session {
            cmd.env(SESSION_VAR, id);
        }
        cmd.output().expect("run moai")
    }

    /// A card of `kind` `task` with one open criterion, taken by the lane and sent to review.
    fn review_card(&self, kind: &str) -> String {
        let added = result(&self.card(
            None,
            &[
                "add",
                "--title",
                "fixture card",
                "--kind",
                kind,
                "--scope",
                "apps/a/",
                "--ac",
                "one",
            ],
        ));
        let name = added["name"].as_str().unwrap().to_string();
        result(&self.card(None, &["ready", &name]));
        result(&self.card(Some(S_A), &["take", &name]));
        result(&self.card(Some(S_A), &["review", &name, "--generation", "1"]));
        name
    }

    /// One verdict event written as `card verdict` writes it: tier 1 with a green receipt of the
    /// head, the tier every head that changes code is judged at.
    fn verdict(&self, card: &str, head: &str, pass: bool, reviewer: &str) {
        let conn =
            Connection::open(ledger::ledger_path(self.home.path(), "demo").unwrap()).unwrap();
        let base = git(&self.project, &["rev-parse", &format!("{head}^")]);
        let payload = json!({
            "head": head, "base": base, "verdict": if pass { "pass" } else { "fail" },
            "tier": 1, "evidence": "the fixture's evidence",
            "receipt": {"head": head, "steps": [{"name": "test", "exit": 0}]},
            "reviewer": reviewer,
        });
        self.events.set(self.events.get() + 1);
        conn.execute(
            "INSERT INTO event (event_key, at, actor, card_id, kind, payload) \
             VALUES (?1, unixepoch(), ?2, ?3, 'card.verdict', ?4)",
            rusqlite::params![
                format!("test-verdict-{}", self.events.get()),
                reviewer,
                card.trim_start_matches("C-").parse::<i64>().unwrap(),
                payload.to_string()
            ],
        )
        .unwrap();
    }

    /// `card finish` as the leader's session, a session that is neither a reviewer nor an author.
    fn finish(&self, card: &str, extra: &[&str]) -> Output {
        self.finish_as(S_LEAD, card, extra)
    }

    fn finish_as(&self, session: &str, card: &str, extra: &[&str]) -> Output {
        let mut all = vec!["finish", card];
        all.extend_from_slice(extra);
        all.extend_from_slice(&["--by", "leader"]);
        self.card(Some(session), &all)
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

    /// A refused finish: exit 1, the code in the message, and no table changed.
    fn refused(&self, card: &str, extra: &[&str], code: &str) {
        let before = self.dump();
        let out = self.finish(card, extra);
        assert_eq!(
            out.status.code(),
            Some(1),
            "{}",
            String::from_utf8_lossy(&out.stdout)
        );
        let v = envelope(&out);
        assert_eq!(v["ok"], json!(false), "{v}");
        let message = v["error"]["message"].as_str().unwrap_or_default();
        assert!(message.contains(code), "{code} not in: {message}");
        assert_eq!(before, self.dump(), "the refusal wrote to the ledger");
    }
}

fn each_mode(f: impl Fn(&World, &str)) {
    for mode in ["observe", "enforce"] {
        f(&World::new(mode, true), mode);
    }
}

// ---- AC-5 -----------------------------------------------------------------------------------

/// AC-5: a card with no verdict record is refused with `no_verdict` in both modes.
#[test]
fn ac5_a_landed_head_with_no_verdict_record_is_refused() {
    each_mode(|w, mode| {
        let card = w.review_card("task");
        w.refused(&card, &["--head", &w.head, "--met", "1=x"], "no_verdict");
        let _ = mode;
    });
}

/// AC-5: a verdict for another head is refused with `verdict_other_head`.
#[test]
fn ac5_a_verdict_for_another_head_is_refused() {
    each_mode(|w, _| {
        let card = w.review_card("task");
        w.verdict(&card, &w.other, true, S_AUD);
        w.refused(
            &card,
            &["--head", &w.head, "--met", "1=x"],
            "verdict_other_head",
        );
    });
}

/// AC-5: a FAIL on the head is refused with `verdict_fail`.
#[test]
fn ac5_a_failed_verdict_is_refused() {
    each_mode(|w, _| {
        let card = w.review_card("task");
        w.verdict(&card, &w.head, false, S_AUD);
        w.refused(&card, &["--head", &w.head, "--met", "1=x"], "verdict_fail");
    });
}

/// AC-5: a PASS of a session other than the lander and the authors closes the card, in both
/// modes, and the answer says it closed on the landing.
#[test]
fn ac5_an_independent_pass_closes_the_card() {
    each_mode(|w, mode| {
        let card = w.review_card("task");
        w.verdict(&card, &w.head, true, S_AUD);
        let done = result(&w.finish(&card, &["--head", &w.head, "--met", "1=x"]));
        assert_eq!(done["closed"], json!(true), "{mode}: {done}");
        assert_eq!(done["via"], "landing", "{done}");
        assert_eq!(done["state"], "done", "{done}");
    });
}

/// AC-5: a PASS of the session that wrote the head is no independent verdict.
#[test]
fn ac5_a_pass_of_the_author_is_refused() {
    each_mode(|w, _| {
        let card = w.review_card("task");
        w.verdict(&card, &w.head, true, S_A);
        w.refused(&card, &["--head", &w.head, "--met", "1=x"], "self_verdict");
    });
}

/// AC-5: a call that records the landing and closes nothing (a criterion stays open) reads no
/// verdict: the landing is recorded as it was.
#[test]
fn ac5_a_landing_only_call_reads_no_verdict() {
    each_mode(|w, mode| {
        let card = w.review_card("task");
        let done = result(&w.finish(&card, &["--head", &w.head]));
        assert_eq!(done["closed"], json!(false), "{mode}: {done}");
        assert_eq!(done["events"], json!(1), "{done}");
    });
}

// ---- AC-6 -----------------------------------------------------------------------------------

/// AC-6: a research card closes on a note or a report with no verdict read, and a merged pull
/// request link closes a card on a profile with no integration branch.
#[test]
fn ac6_the_other_evidence_reads_no_verdict() {
    each_mode(|w, mode| {
        let card = w.review_card("research");
        let reports = w.project.join("reports");
        fs::create_dir_all(&reports).unwrap();
        fs::write(reports.join("r.md"), "the report\n").unwrap();
        // The holder of the card closes it: a research card has no landing to pass the gate of
        // another session's live lease.
        let done = result(&w.finish_as(
            S_A,
            &card,
            &["--evidence-report", "reports/r.md", "--met", "1=x"],
        ));
        assert_eq!(done["closed"], json!(true), "{mode}: {done}");
        assert_eq!(done["via"], "report", "{done}");
    });
    // A profile with no integration branch: a card closes on its merged pull request link, no
    // verdict is read.
    let w = World::new("observe", false);
    let card = w.review_card("task");
    let conn = Connection::open(ledger::ledger_path(w.home.path(), "demo").unwrap()).unwrap();
    let id: i64 = card.trim_start_matches("C-").parse().unwrap();
    conn.execute(
        "INSERT INTO card_link (from_card, kind, target_host, target_project, target_number, \
         pr_state, merge_ref, state, created_at, updated_at) \
         VALUES (?1, 'pr', 'github', 'o/r', 1, 'merged', 'abc', 'active', unixepoch(), unixepoch())",
        [id],
    )
    .unwrap();
    drop(conn);
    let done = result(&w.finish_as(S_A, &card, &["--met", "1=x"]));
    assert_eq!(done["closed"], json!(true), "{done}");
    assert_eq!(done["via"], "merged_pr", "{done}");
}
