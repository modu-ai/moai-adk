//! A `ready` card with a recorded PASS verdict reaches the landing record
//! (SPEC-20261125-card-ready-pass-landing-path): `card landed` and `card assign` accept it only
//! on a PASS of the exact head that no author of the card wrote, and every other state, verdict
//! and caller is refused with the ledger unchanged. Every test runs on a temporary home and a
//! temporary repository; nothing here touches the real ledger.
//!
//! Sessions: `S_A` took the card and sent it to review, `S_C` audited it (the PASS), `S_D` is a
//! second reviewer that never held the card.

use crate::common;

use common::home::TempHome;
use moai_cli::ledger;
use rusqlite::Connection;
use serde_json::{json, Value};
use std::path::Path;
use std::process::Output;

const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
const S_A: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
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
    /// The first commit, the base every verdict names.
    base: String,
    /// The audited head; the integration branch `integ` points at it.
    head: String,
    /// A later commit that no branch holds, for a PASS of another head.
    other: String,
}

impl World {
    fn new() -> Self {
        let home = TempHome::new();
        let project = home.project("demo");
        ledger::register(home.path(), "demo").unwrap();
        git(&project, &["init", "-q", "-b", "main"]);
        std::fs::create_dir_all(project.join("notes")).unwrap();
        std::fs::write(project.join("notes/x.md"), "a\n").unwrap();
        git(&project, &["add", "notes/x.md"]);
        git(&project, &["commit", "-q", "-m", "init"]);
        let base = git(&project, &["rev-parse", "HEAD"]).trim().to_string();
        // The commits of the card carry the author's `Moai-Session` trailer, as `card finish`
        // reads the authors of a head from it.
        let trailer = format!("\n\nMoai-Session: {S_A}:0123456789ab");
        std::fs::write(project.join("notes/x.md"), "a\nb\n").unwrap();
        git(
            &project,
            &["commit", "-q", "-am", &format!("change{trailer}")],
        );
        let head = git(&project, &["rev-parse", "HEAD"]).trim().to_string();
        git(&project, &["branch", "integ"]);
        std::fs::write(project.join("notes/x.md"), "a\nb\nc\n").unwrap();
        git(
            &project,
            &["commit", "-q", "-am", &format!("later{trailer}")],
        );
        let other = git(&project, &["rev-parse", "HEAD"]).trim().to_string();
        World {
            p: project.to_str().unwrap().to_string(),
            home,
            base,
            head,
            other,
        }
    }

    /// `moai card <args> --project <project>` as `session`, the profile written last.
    fn card(&self, session: Option<&str>, args: &[&str]) -> Output {
        let cfg = moai_cli::state_root::path(Path::new(&self.p), "project.yaml");
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

    /// A card in `ready`, not taken.
    fn ready_card(&self) -> String {
        let added = result(&self.card(
            None,
            &[
                "add", "--title", "fixture", "--kind", "task", "--scope", "apps/a/", "--ac",
                "fixture",
            ],
        ));
        let name = added["name"].as_str().unwrap().to_string();
        result(&self.card(None, &["ready", &name]));
        name
    }

    /// A card in `review`, held by `S_A` at generation 1.
    fn review_card(&self) -> String {
        let name = self.ready_card();
        result(&self.card(Some(S_A), &["take", &name]));
        result(&self.card(Some(S_A), &["review", &name, "--generation", "1"]));
        name
    }

    /// `session` records a verdict of `head` on the card.
    fn verdict(&self, session: &str, card: &str, head: &str, pass: bool, evidence: &str) {
        let word = if pass { "pass" } else { "fail" };
        result(&self.card(
            Some(session),
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
                evidence,
            ],
        ));
    }

    /// The card `S_A` wrote, audited in review with each `(reviewer, head, pass)` of `verdicts`
    /// in order, then sent back to `ready` (`reopen`, `release`).
    fn ready_with(&self, verdicts: &[(&str, &str, bool)]) -> String {
        let name = self.review_card();
        for (n, (reviewer, head, pass)) in verdicts.iter().enumerate() {
            self.verdict(reviewer, &name, head, *pass, &format!("read the diff {n}"));
        }
        result(&self.card(Some(S_A), &["reopen", &name, "--generation", "1"]));
        result(&self.card(Some(S_A), &["release", &name, "--generation", "1"]));
        assert_eq!(self.state(&name), "ready");
        name
    }

    /// [`World::ready_with`] one PASS of `head` by `S_C`.
    fn ready_with_pass(&self) -> String {
        self.ready_with(&[(S_C, &self.head, true)])
    }

    fn landed(&self, session: Option<&str>, card: &str, head: &str, by: &str) -> Output {
        self.card(session, &["landed", card, "--head", head, "--by", by])
    }

    fn assign(&self, card: &str, auditor: &str) -> Output {
        self.card(
            None,
            &["assign", card, "--auditor", auditor, "--by", "leader"],
        )
    }

    /// `moai card next <project>` (it takes the project as an argument, not as an option).
    fn next(&self) -> Value {
        let mut cmd = self.home.command(&["card", "next", &self.p]);
        cmd.stdin(std::process::Stdio::null());
        result(&cmd.output().expect("run moai"))
    }

    fn state(&self, card: &str) -> String {
        let shown = result(&self.card(None, &["show", card]));
        shown["state"].as_str().unwrap().to_string()
    }

    /// Runs `f` on a read-only handle to the ledger.
    fn read<T>(&self, f: impl FnOnce(&Connection) -> T) -> T {
        let l = ledger::open_read_only(self.home.path(), "demo").unwrap();
        f(l.conn())
    }

    fn events(&self) -> i64 {
        self.read(|c| {
            c.query_row("SELECT count(*) FROM event", [], |r| r.get(0))
                .unwrap()
        })
    }

    fn count(&self, kind: &str) -> i64 {
        self.read(|c| {
            c.query_row("SELECT count(*) FROM event WHERE kind = ?1", [kind], |r| {
                r.get(0)
            })
            .unwrap()
        })
    }

    fn payload_of_last(&self, kind: &str) -> Value {
        let payload: String = self.read(|c| {
            c.query_row(
                "SELECT payload FROM event WHERE kind = ?1 ORDER BY event_id DESC LIMIT 1",
                [kind],
                |r| r.get(0),
            )
            .unwrap()
        });
        serde_json::from_str(&payload).unwrap()
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

/// Asserts exit 1 and a `conflict` failure envelope; returns the message.
fn conflict(out: &Output) -> String {
    assert_eq!(
        out.status.code(),
        Some(1),
        "{} {}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], json!(false), "{v}");
    assert_eq!(v["error"]["kind"], json!("conflict"), "{v}");
    v["error"]["message"]
        .as_str()
        .unwrap_or_default()
        .to_string()
}

/// AC-1 (REQ-1): a `ready` card with a PASS of the head takes the landing record from the
/// leader: one `card.landed` event names the head, the by and `from_state` `ready`.
#[test]
fn ac1_a_ready_card_with_a_pass_takes_the_landing_record() {
    let w = World::new();
    let name = w.ready_with_pass();
    let out = w.landed(Some(S_D), &name, &w.head, "leader");
    let v = result(&out);
    assert_eq!(v["action"], json!("landed"), "{v}");
    assert_eq!(w.count("card.landed"), 1);
    let payload = w.payload_of_last("card.landed");
    assert_eq!(payload["head"], json!(w.head), "{payload}");
    assert_eq!(payload["by"], json!("leader"), "{payload}");
    assert_eq!(payload["from_state"], json!("ready"), "{payload}");
    assert_eq!(w.state(&name), "ready");
}

/// AC-2 (REQ-2, REQ-3): no verdict, a PASS of another head only, and a newest FAIL (of the same
/// reviewer or of a second one) each refuse and leave the ledger unchanged.
#[test]
fn ac2_a_ready_card_without_a_pass_of_the_head_is_refused() {
    let w = World::new();

    let bare = w.ready_card();
    let before = w.events();
    let message = conflict(&w.landed(None, &bare, &w.head, "leader"));
    assert!(message.contains("PASS"), "{message}");
    assert_eq!(w.events(), before, "a refusal wrote an event");

    // A PASS that names another head is no PASS of this head.
    let other_head = w.ready_with(&[(S_D, &w.other, true)]);
    let before = w.events();
    let message = conflict(&w.landed(None, &other_head, &w.head, "leader"));
    assert!(message.contains("PASS"), "{message}");
    assert_eq!(w.events(), before);

    // The newest verdict of a reviewer on the head is a FAIL, after an earlier PASS.
    let same_reviewer = w.ready_with(&[(S_C, &w.head, true), (S_C, &w.head, false)]);
    let before = w.events();
    let message = conflict(&w.landed(None, &same_reviewer, &w.head, "leader"));
    assert!(message.contains("FAIL"), "{message}");
    assert_eq!(w.events(), before);

    // A second reviewer's FAIL stands next to the first one's PASS.
    let second_reviewer = w.ready_with(&[(S_C, &w.head, true), (S_D, &w.head, false)]);
    let before = w.events();
    let message = conflict(&w.landed(None, &second_reviewer, &w.head, "leader"));
    assert!(message.contains("FAIL"), "{message}");
    assert_eq!(w.events(), before);
    assert_eq!(w.count("card.landed"), 0, "no refusal wrote a landing");
}

/// AC-3 (REQ-4, REQ-5): a `--by` other than `leader`, and a PASS whose reviewer held a
/// reservation of the card, each refuse and leave the ledger unchanged.
#[test]
fn ac3_the_ready_path_needs_the_leader_and_an_independent_reviewer() {
    let w = World::new();
    let name = w.ready_with_pass();
    let before = w.events();
    let message = conflict(&w.landed(Some(S_D), &name, &w.head, "agent-1"));
    assert!(message.contains("leader"), "{message}");
    assert_eq!(w.events(), before, "a refusal wrote an event");

    // `S_C` audited the card, then takes it and gives it back: the reviewer is now an author.
    result(&w.card(Some(S_C), &["take", &name]));
    result(&w.card(Some(S_C), &["release", &name, "--generation", "2"]));
    assert_eq!(w.state(&name), "ready");
    let before = w.events();
    let message = conflict(&w.landed(Some(S_D), &name, &w.head, "leader"));
    assert!(message.contains("reservation"), "{message}");
    assert_eq!(w.events(), before);
    assert_eq!(w.count("card.landed"), 0);
}

/// AC-4 (REQ-6, REQ-7): the PASS reviewer is assigned to the `ready` card and shows as its
/// auditor; no PASS, or a session that is not a PASS reviewer, is refused and writes nothing.
#[test]
fn ac4_assign_on_a_ready_card_names_the_pass_reviewer_only() {
    let w = World::new();
    let bare = w.ready_card();
    let before = w.events();
    let message = conflict(&w.assign(&bare, S_C));
    assert!(message.contains("PASS"), "{message}");
    assert_eq!(w.events(), before, "a refusal wrote an event");

    let name = w.ready_with_pass();
    let before = w.events();
    let message = conflict(&w.assign(&name, S_D));
    assert!(message.contains(S_D), "{message}");
    assert_eq!(w.events(), before);
    // A reviewer whose only verdict is a FAIL is no PASS reviewer.
    let failed = w.ready_with(&[(S_D, &w.head, false)]);
    let before = w.events();
    let message = conflict(&w.assign(&failed, S_D));
    assert!(message.contains("PASS"), "{message}");
    assert_eq!(w.events(), before);
    let before = w.events();
    // The author of the card stays refused, as in review.
    let message = conflict(&w.assign(&name, S_A));
    assert!(message.contains(S_A), "{message}");
    assert_eq!(w.events(), before);
    assert_eq!(w.count("card.audit_assigned"), 0);

    let v = result(&w.assign(&name, S_C));
    assert_eq!(v["action"], json!("assign"), "{v}");
    assert_eq!(w.count("card.audit_assigned"), 1);
    let shown = result(&w.card(None, &["show", &name]));
    assert_eq!(shown["roles"]["auditor"], json!(S_C), "{shown}");
}

/// AC-5 (REQ-8, REQ-9): a card in `review` keeps both commands as they are (a free `--by`, no
/// verdict needed), and a `draft`, an `active` and a `done` card are refused naming `ready` and
/// `review`.
#[test]
fn ac5_review_is_unchanged_and_other_states_name_the_accepted_ones() {
    let w = World::new();
    let review = w.review_card();
    result(&w.landed(Some(S_D), &review, &w.head, "agent-1"));
    result(&w.assign(&review, S_C));
    w.verdict(S_C, &review, &w.head, true, "read the diff");
    assert_eq!(w.count("card.landed"), 1);
    assert_eq!(w.count("card.audit_assigned"), 1);
    let payload = w.payload_of_last("card.landed");
    assert_eq!(payload["by"], json!("agent-1"), "{payload}");
    assert_eq!(
        payload.get("from_state"),
        None,
        "the review record is as it was: {payload}"
    );

    let draft = result(&w.card(
        None,
        &[
            "add", "--title", "draft", "--kind", "task", "--scope", "apps/b/", "--ac", "fixture",
        ],
    ))["name"]
        .as_str()
        .unwrap()
        .to_string();
    let active = w.ready_card();
    result(&w.card(Some(S_A), &["take", &active]));
    // `done`: the review card above closes on its landed head, which `integ` holds.
    result(&w.card(
        Some(S_D),
        &["finish", &review, "--by", "leader", "--met", "1=fixture"],
    ));
    assert_eq!(w.state(&review), "done");

    for card in [&draft, &active, &review] {
        let before = w.events();
        for message in [
            conflict(&w.landed(None, card, &w.head, "leader")),
            conflict(&w.assign(card, S_C)),
        ] {
            assert!(message.contains("`ready`"), "{card}: {message}");
            assert!(message.contains("`review`"), "{card}: {message}");
        }
        assert_eq!(w.events(), before, "{card}: a refusal wrote an event");
    }
}

/// AC-6 (REQ-10, REQ-11): after the record of a `ready` card whose head `integ` holds,
/// `card next` skips the card as already landed and `card finish` closes it on the head.
#[test]
fn ac6_the_ready_record_lets_next_skip_and_finish_close() {
    let w = World::new();
    let name = w.ready_with_pass();
    result(&w.landed(Some(S_D), &name, &w.head, "leader"));
    let listing = w.next();
    let skipped = listing["skipped"].as_array().cloned().unwrap_or_default();
    let entry = skipped
        .iter()
        .find(|e| e["name"] == json!(name))
        .unwrap_or_else(|| panic!("{name} is not skipped: {listing}"));
    assert_eq!(entry["reason"], json!("already_landed"), "{entry}");

    let done = result(&w.card(
        Some(S_D),
        &["finish", &name, "--by", "leader", "--met", "1=fixture"],
    ));
    assert_eq!(done["state"], json!("done"), "{done}");
    assert_eq!(w.state(&name), "done");
}
