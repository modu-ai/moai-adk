//! Card roles (SPEC-20261122-card-roles): a card shows its authors, its auditor and its lander
//! from ledger rows, an audit assignment never names an author, and a typed verdict row from a
//! session that held the card is refused. Every test runs on a temporary home and a temporary
//! repository; nothing here touches the real ledger.
//!
//! Sessions are made with the host's session variable, one per role: `S_A` took the card first,
//! `S_B` took it after `S_A` gave it back and sent it to review, `S_C` and `S_D` never held it.

mod common;

use common::home::TempHome;
use moai_cli::ledger;
use rusqlite::Connection;
use serde_json::{json, Value};
use std::path::Path;
use std::process::Output;

const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
const S_A: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
const S_B: &str = "bbbbbbbb-0000-4000-8000-00000000000b";
const S_C: &str = "cccccccc-0000-4000-8000-00000000000c";
const S_D: &str = "dddddddd-0000-4000-8000-00000000000d";
const PROFILE: &[u8] = b"{\"schema_version\": 1, \"project_id\": \"demo\", \"purpose\": \"fixture\", \"tools\": [\"rust\"], \"development_methods\": [\"sdd\"], \"hosts\": [\"claude\"], \"work_flow\": {\"flow\": \"m-flow\", \"integration_branch\": \"integ\"}}";

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

struct World {
    home: TempHome,
    p: String,
    head: String,
}

impl World {
    fn new() -> Self {
        let home = TempHome::new();
        let project = home.project("demo");
        ledger::register(home.path(), "demo").unwrap();
        git(&project, &["init", "-q", "-b", "main"]);
        std::fs::write(project.join("a.txt"), "a\n").unwrap();
        git(&project, &["add", "a.txt"]);
        git(&project, &["commit", "-q", "-m", "init"]);
        let head = git(&project, &["rev-parse", "HEAD"]).trim().to_string();
        git(&project, &["branch", "integ"]);
        World {
            p: project.to_str().unwrap().to_string(),
            home,
            head,
        }
    }

    /// `moai card <args> --project <project>` as `session`, the profile written last.
    fn card(&self, session: Option<&str>, args: &[&str]) -> Output {
        let cfg = Path::new(&self.p).join(".moai-cli/project.yaml");
        std::fs::create_dir_all(cfg.parent().unwrap()).unwrap();
        std::fs::write(&cfg, PROFILE).unwrap();
        let mut all = vec!["card"];
        all.extend_from_slice(args);
        all.extend_from_slice(&["--project", &self.p]);
        let mut cmd = self.home.command(&all);
        if let Some(id) = session {
            cmd.env(SESSION_VAR, id);
        }
        cmd.stdin(std::process::Stdio::null());
        cmd.output().expect("run moai")
    }

    /// A card taken by `S_A` (generation 1), given back, taken by `S_B` (generation 2) and sent
    /// to review by `S_B`; returns its name.
    fn review_card(&self) -> String {
        let name = self.ready_card();
        result(&self.card(Some(S_A), &["take", &name]));
        result(&self.card(Some(S_A), &["release", &name, "--generation", "1"]));
        result(&self.card(Some(S_B), &["take", &name]));
        result(&self.card(Some(S_B), &["review", &name, "--generation", "2"]));
        name
    }

    /// A card in `ready`, not taken.
    fn ready_card(&self) -> String {
        let added = result(&self.card(
            None,
            &[
                "add", "--title", "fixture", "--kind", "task", "--scope", "apps/a/",
            ],
        ));
        let name = added["name"].as_str().unwrap().to_string();
        result(&self.card(None, &["ready", &name]));
        name
    }

    /// An active card: taken by `S_A` and not sent to review.
    fn active_card(&self) -> String {
        let name = self.ready_card();
        result(&self.card(Some(S_A), &["take", &name]));
        name
    }

    fn verdict(&self, session: &str, card: &str, head: &str, result_word: &str) -> Output {
        self.card(
            Some(session),
            &[
                "verdict",
                card,
                "--head",
                head,
                "--result",
                result_word,
                "--evidence",
                "note-1",
            ],
        )
    }

    fn assign(&self, by: &str, card: &str, auditor: &str) -> Output {
        self.card(None, &["assign", card, "--auditor", auditor, "--by", by])
    }

    /// Runs `f` on a read-only handle to the ledger.
    fn read<T>(&self, f: impl FnOnce(&Connection) -> T) -> T {
        let l = ledger::open_read_only(self.home.path(), "demo").unwrap();
        f(l.conn())
    }

    fn count(&self, kind: &str) -> i64 {
        self.read(|c| {
            c.query_row("SELECT count(*) FROM event WHERE kind = ?1", [kind], |r| {
                r.get(0)
            })
            .unwrap()
        })
    }

    fn events(&self) -> i64 {
        self.read(|c| {
            c.query_row("SELECT count(*) FROM event", [], |r| r.get(0))
                .unwrap()
        })
    }

    fn payload_of_last(&self, kind: &str) -> (Value, String) {
        let (payload, actor): (String, String) = self.read(|c| {
            c.query_row(
                "SELECT payload, actor FROM event WHERE kind = ?1 ORDER BY event_id DESC LIMIT 1",
                [kind],
                |r| Ok((r.get(0)?, r.get(1)?)),
            )
            .unwrap()
        });
        (serde_json::from_str(&payload).unwrap(), actor)
    }
}

/// Asserts exit 0 and a success envelope, and unwraps the result.
fn result(out: &Output) -> Value {
    assert_eq!(
        out.status.code(),
        Some(0),
        "{} {}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

/// Asserts exit 1 and a failure envelope; returns the message.
fn refusal(out: &Output) -> String {
    assert_eq!(
        out.status.code(),
        Some(1),
        "{} {}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], json!(false), "{v}");
    v["error"]["message"]
        .as_str()
        .unwrap_or_default()
        .to_string()
}

// ---- roles_ -----------------------------------------------------------------------------------

/// AC-1: the roles of a card come from ledger rows: every session that held a generation is an
/// author, the assignment names the auditor and the landing row names the lander; a card with none
/// of these shows each role as absent.
#[test]
fn roles_are_read_from_reservations_the_assignment_and_the_landing() {
    let w = World::new();
    let bare = w.ready_card();
    let shown = result(&w.card(None, &["show", &bare]));
    assert_eq!(shown["roles"]["authors"], json!([]), "{shown}");
    assert_eq!(shown["roles"]["auditor"], Value::Null, "{shown}");
    assert_eq!(shown["roles"]["lander"], Value::Null, "{shown}");

    let name = w.review_card();
    let shown = result(&w.card(None, &["show", &name]));
    assert_eq!(shown["roles"]["authors"], json!([S_A, S_B]), "{shown}");
    assert_eq!(shown["roles"]["auditor"], Value::Null, "{shown}");

    result(&w.assign("leader", &name, S_C));
    result(&w.card(
        Some(S_B),
        &["landed", &name, "--head", &w.head, "--by", "leader"],
    ));
    let shown = result(&w.card(None, &["show", &name]));
    assert_eq!(shown["roles"]["authors"], json!([S_A, S_B]), "{shown}");
    assert_eq!(shown["roles"]["auditor"], json!(S_C), "{shown}");
    assert_eq!(shown["roles"]["lander"]["by"], json!("leader"), "{shown}");
    assert_eq!(shown["roles"]["lander"]["head"], json!(w.head), "{shown}");
}

// ---- verdict_ ---------------------------------------------------------------------------------

/// AC-2: the session that holds the card and a session that held an earlier generation are both
/// refused a verdict, with the card and the reason named, and nothing is written.
#[test]
fn verdict_author_sessions_are_refused_with_the_reason() {
    let w = World::new();
    let name = w.review_card();
    let before = w.events();
    for (who, label) in [(S_B, "the current holder"), (S_A, "an earlier holder")] {
        let message = refusal(&w.verdict(who, &name, &w.head, "pass"));
        assert!(message.contains(&name), "{label}: {message}");
        assert!(message.contains("held"), "{label}: {message}");
        assert_eq!(w.events(), before, "{label}: an event was written");
    }
    assert_eq!(w.count("card.verdict"), 0);
}

/// AC-3: a session that never held a card in review writes one row with the head, the result, the
/// evidence reference and itself as the actor; a short head, an unknown result and an active card
/// each refuse before any write.
#[test]
fn verdict_rows_hold_the_head_the_result_the_evidence_and_the_writer() {
    let w = World::new();
    let name = w.review_card();
    let before = w.events();
    for (label, head, word) in [
        ("a short head", &w.head[..39], "pass"),
        ("a result of maybe", w.head.as_str(), "maybe"),
    ] {
        refusal(&w.verdict(S_C, &name, head, word));
        assert_eq!(w.events(), before, "{label} wrote an event");
    }
    let active = w.active_card();
    let before = w.events();
    refusal(&w.verdict(S_C, &active, &w.head, "pass"));
    assert_eq!(w.events(), before, "an active card took a verdict");

    result(&w.verdict(S_C, &name, &w.head, "pass"));
    assert_eq!(w.count("card.verdict"), 1);
    let (payload, actor) = w.payload_of_last("card.verdict");
    assert_eq!(payload["head"], json!(w.head), "{payload}");
    assert_eq!(payload["result"], json!("pass"), "{payload}");
    assert_eq!(payload["evidence"], json!("note-1"), "{payload}");
    assert_eq!(actor, S_C);
}

/// AC-6: after an assignment, a verdict from another session that never held the card is refused
/// and names the assigned auditor; the assigned auditor's verdict is accepted.
#[test]
fn verdict_assigned_auditor_is_the_only_one_accepted() {
    let w = World::new();
    let name = w.review_card();
    result(&w.assign("leader", &name, S_C));
    let before = w.events();
    let message = refusal(&w.verdict(S_D, &name, &w.head, "pass"));
    assert!(message.contains(S_C), "the auditor is not named: {message}");
    assert_eq!(w.events(), before);
    result(&w.verdict(S_C, &name, &w.head, "fail"));
    assert_eq!(w.count("card.verdict"), 1);
}

/// AC-7: the same session repeating the same head and result writes no second row and says so; a
/// different result for the same head is a new row.
#[test]
fn verdict_repeat_writes_no_second_row() {
    let w = World::new();
    let name = w.review_card();
    result(&w.verdict(S_C, &name, &w.head, "pass"));
    let again = result(&w.verdict(S_C, &name, &w.head, "pass"));
    assert_eq!(again["written"], json!(false), "{again}");
    assert_eq!(w.count("card.verdict"), 1);
    let other = result(&w.verdict(S_C, &name, &w.head, "fail"));
    assert_eq!(other["written"], json!(true), "{other}");
    assert_eq!(w.count("card.verdict"), 2);
}

// ---- assign_ ----------------------------------------------------------------------------------

/// AC-4: an assignment that names the holder of any generation refuses and writes nothing.
#[test]
fn assign_author_identities_are_refused() {
    let w = World::new();
    let name = w.review_card();
    let before = w.events();
    for who in [S_A, S_B] {
        let message = refusal(&w.assign("leader", &name, who));
        assert!(message.contains(who), "{message}");
        assert!(message.contains(&name), "{message}");
        assert_eq!(w.events(), before, "{who}: an event was written");
    }
    assert_eq!(w.count("card.audit_assigned"), 0);
}

/// AC-5: the leader's assignment of a session that never held the card writes one row with the
/// auditor and the caller; another caller word and a card that is not in review each refuse
/// before any write.
#[test]
fn assign_rows_hold_the_auditor_and_the_caller_and_only_the_leader_writes_them() {
    let w = World::new();
    let name = w.review_card();
    let before = w.events();
    refusal(&w.assign("agent-1", &name, S_C));
    let active = w.active_card();
    let after_active = w.events();
    refusal(&w.assign("leader", &active, S_C));
    assert_eq!(w.events(), after_active);
    assert!(after_active >= before);
    assert_eq!(w.count("card.audit_assigned"), 0);

    result(&w.assign("leader", &name, S_C));
    assert_eq!(w.count("card.audit_assigned"), 1);
    let (payload, _actor) = w.payload_of_last("card.audit_assigned");
    assert_eq!(payload["auditor"], json!(S_C), "{payload}");
    assert_eq!(payload["by"], json!("leader"), "{payload}");
}
