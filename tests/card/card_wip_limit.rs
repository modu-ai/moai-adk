//! The limit of cards in work and in review (SPEC-20261125-landing-moves-checkout REQ-1 to
//! REQ-4): `card take` and `card add` refuse once the `active` and `review` cards number at least
//! `wip_limit` (default 2) times the live registered lanes (at least 1), `--by user` adds pass,
//! `wip_limit` 0 turns the limit off, and `card next` at the limit names the cards that wait for
//! an audit.
//!
//! Every command runs against a temporary HOME and a temporary git repository; the real home
//! folder and the real session registry are never read or written.
#![cfg(unix)]

use crate::common;

use common::git::git_in;
use common::home::TempHome;
use moai_cli::ledger;
use serde_json::{json, Value};
use std::fs;
use std::path::PathBuf;
use std::process::{Output, Stdio};

const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";

/// A well-formed host session id for number `n`.
fn sid(n: u64) -> String {
    format!("{n:08x}-0000-4000-8000-{n:012x}")
}

fn profile(wip_limit: Option<u32>) -> Vec<u8> {
    let limit = wip_limit.map_or(String::new(), |n| format!(", \"wip_limit\": {n}"));
    format!(
        "{{\"schema_version\": 1, \"project_id\": \"demo\", \"purpose\": \"fixture\", \"tools\": [\"rust\"], \"development_methods\": [\"sdd\"], \"hosts\": [\"claude\"], \"work_flow\": {{\"flow\": \"m-flow\", \"integration_branch\": \"integ\"{limit}}}}}"
    )
    .into_bytes()
}

struct World {
    home: TempHome,
    project: PathBuf,
}

fn world(wip_limit: Option<u32>) -> World {
    let home = TempHome::new();
    let project = home.project("demo");
    ledger::register(home.path(), "demo").unwrap();
    git_in(&project, &["init", "-q", "-b", "main"]);
    fs::write(project.join("a.txt"), "a\n").unwrap();
    git_in(&project, &["add", "-A"]);
    git_in(&project, &["commit", "-q", "-m", "first"]);
    let w = World { home, project };
    w.set_limit(wip_limit);
    w
}

impl World {
    fn set_limit(&self, wip_limit: Option<u32>) {
        let cfg = moai_cli::state_root::path(&self.project, "project.yaml");
        fs::create_dir_all(cfg.parent().unwrap()).unwrap();
        fs::write(&cfg, profile(wip_limit)).unwrap();
    }

    fn run(&self, session: Option<&str>, args: &[&str]) -> Output {
        let mut cmd = self.home.command(args);
        cmd.current_dir(&self.project);
        if let Some(id) = session {
            cmd.env(SESSION_VAR, id);
        }
        cmd.stdin(Stdio::null()).output().expect("run moai")
    }

    fn card(&self, session: Option<&str>, args: &[&str]) -> Output {
        let mut all = vec!["card"];
        all.extend_from_slice(args);
        all.extend_from_slice(&["--project", self.project.to_str().unwrap()]);
        self.run(session, &all)
    }

    /// `moai card next <project>` as `session`.
    fn next(&self, session: &str) -> Output {
        self.run(
            Some(session),
            &["card", "next", self.project.to_str().unwrap()],
        )
    }

    /// A lane entry of the session registry, as `session-register` writes it.
    fn lane(&self, name: &str, session: &str) {
        let out = self.run(
            Some(session),
            &[
                "session-register",
                self.project.to_str().unwrap(),
                "--name",
                name,
            ],
        );
        result(&out);
    }

    /// A leader entry of the session registry (`session-register --mode leader`).
    fn leader(&self, name: &str, session: &str) {
        let out = self.run(
            Some(session),
            &[
                "session-register",
                self.project.to_str().unwrap(),
                "--name",
                name,
                "--mode",
                "leader",
            ],
        );
        result(&out);
    }

    /// Ends the lease of the registry entry `name`: its `lease_until` becomes 1.
    fn expire(&self, name: &str) {
        let dir = moai_cli::lane_registry::registry_dir_unlinked(&self.project).unwrap();
        let file = moai_cli::lane_registry::entry_file(&dir, name)
            .ok()
            .flatten()
            .unwrap_or_else(|| dir.join(format!("{name}.json")));
        let mut entry: Value = serde_json::from_slice(&fs::read(&file).unwrap()).unwrap();
        entry["lease_until"] = json!(1);
        fs::write(&file, entry.to_string()).unwrap();
    }

    /// A ready card with one criterion and its own write scope; returns its name.
    fn ready_card(&self, scope: &str) -> String {
        let added = result(&self.card(
            None,
            &["add", "--title", "t", "--scope", scope, "--ac", "one"],
        ));
        let name = added["name"].as_str().unwrap().to_string();
        result(&self.card(None, &["ready", &name]));
        name
    }

    fn take(&self, session: &str, card: &str) -> Output {
        self.card(Some(session), &["take", card])
    }

    /// `n` cards taken by `session`, each in `active`.
    fn working(&self, session: &str, n: usize) -> Vec<String> {
        (0..n)
            .map(|i| {
                let card = self.ready_card(&format!("src/w{i}/"));
                result(&self.take(session, &card));
                card
            })
            .collect()
    }
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

/// REQ-7 (AC-6): with one lane the limit is two: a third take refuses and names the cards that
/// work, a second registered lane doubles the limit, and the ledger is unchanged by the refusal.
#[test]
fn take_stops_at_the_limit_and_a_second_lane_raises_it() {
    let w = world(None);
    w.lane("agent-1", &sid(1));
    let held = w.working(&sid(1), 2);
    let third = w.ready_card("src/third/");

    let (kind, message) = refusal(&w.take(&sid(1), &third));
    assert_eq!(kind, "conflict", "{message}");
    assert!(message.contains("work-in-progress limit"), "{message}");
    for card in &held {
        assert!(
            message.contains(card.as_str()),
            "{card} is named: {message}"
        );
    }
    let states = w.run(None, &["card", "list", w.project.to_str().unwrap()]);
    let listed = result(&states).to_string();
    assert!(
        listed.contains("ready"),
        "the third card is still ready: {listed}"
    );

    w.lane("agent-2", &sid(2));
    result(&w.take(&sid(2), &third));
}

/// REQ-8 (AC-6): at the limit `card add` refuses for a lane or the leader and passes for the
/// user.
#[test]
fn add_stops_at_the_limit_unless_the_user_adds() {
    let w = world(None);
    w.lane("agent-1", &sid(1));
    w.working(&sid(1), 2);

    let out = w.card(
        None,
        &[
            "add",
            "--title",
            "more",
            "--scope",
            "src/more/",
            "--ac",
            "one",
            "--by",
            "leader",
        ],
    );
    let (kind, message) = refusal(&out);
    assert_eq!(kind, "conflict", "{message}");
    assert!(message.contains("work-in-progress limit"), "{message}");

    let out = w.card(
        None,
        &[
            "add",
            "--title",
            "more",
            "--scope",
            "src/more/",
            "--ac",
            "one",
            "--by",
            "user",
        ],
    );
    result(&out);
}

/// REQ-9 (AC-6): `wip_limit` 0 turns the limit off; another number is the multiple per lane.
#[test]
fn a_wip_limit_of_zero_turns_the_limit_off() {
    let w = world(Some(0));
    w.lane("agent-1", &sid(1));
    w.working(&sid(1), 3);

    w.set_limit(Some(1));
    let fourth = w.ready_card("src/fourth/");
    let (kind, message) = refusal(&w.take(&sid(1), &fourth));
    assert_eq!(kind, "conflict", "{message}");
    assert!(message.contains("work-in-progress limit"), "{message}");
}

/// REQ-10 (AC-7): at the limit `card next` names the review cards that no one audited and the
/// asking session did not write, in place of a card to take.
#[test]
fn next_at_the_limit_names_the_cards_that_wait_for_an_audit() {
    let w = world(None);
    w.lane("agent-1", &sid(1));
    let held = w.working(&sid(1), 2);
    result(&w.card(Some(&sid(1)), &["review", &held[1], "--generation", "1"]));
    let waiting = w.ready_card("src/waiting/");

    let got = result(&w.next(&sid(3)));
    assert_eq!(got["wip"]["at_limit"], json!(true), "{got}");
    assert_eq!(got["next_is"], json!("audit"), "{got}");
    assert_eq!(got["next"], json!(held[1]), "{got}");
    let names: Vec<&str> = got["audit_needed"]
        .as_array()
        .unwrap()
        .iter()
        .map(|c| c["name"].as_str().unwrap())
        .collect();
    assert_eq!(names, vec![held[1].as_str()], "{got}");
    assert_ne!(got["next"], json!(waiting), "no card to take at the limit");

    // The session that wrote the card in review is not asked to audit it.
    let mine = result(&w.next(&sid(1)));
    assert_eq!(mine["audit_needed"], json!([]), "{mine}");
}

/// REQ-1 (D-3): with no live registered lane there is no limit: a session that never registered
/// takes three cards, and a lane with an ended lease and a leader entry alone leave it off too.
#[test]
fn with_no_live_lane_registered_there_is_no_limit() {
    let w = world(None);
    w.working(&sid(1), 3);

    let w = world(None);
    w.lane("agent-2", &sid(2));
    w.expire("agent-2");
    w.leader("leader-1", &sid(9));
    w.working(&sid(1), 3);
}

/// REQ-7 (AC-6): a lane whose lease ended and a leader entry are no lanes; only the one live lane
/// counts, so the limit stays two and the third take refuses.
#[test]
fn a_lane_with_an_ended_lease_and_a_leader_entry_do_not_count() {
    let w = world(None);
    w.lane("agent-1", &sid(1));
    w.lane("agent-2", &sid(2));
    w.expire("agent-2");
    w.leader("leader-1", &sid(9));
    w.working(&sid(1), 2);
    let third = w.ready_card("src/third/");
    let (kind, message) = refusal(&w.take(&sid(1), &third));
    assert_eq!(kind, "conflict", "{message}");
    assert!(message.contains("1 lane(s)"), "{message}");
}

/// REQ-10 (AC-7): at the limit with no review card to audit, `card next` names the limit and no
/// card.
#[test]
fn next_at_the_limit_with_nothing_to_audit_names_the_limit() {
    let w = world(None);
    w.lane("agent-1", &sid(1));
    w.working(&sid(1), 2);
    w.ready_card("src/waiting/");

    let got = result(&w.next(&sid(3)));
    assert_eq!(got["wip"]["at_limit"], json!(true), "{got}");
    assert_eq!(got["next_is"], json!("limit"), "{got}");
    assert_eq!(got["next"], Value::Null, "{got}");
    assert_eq!(got["audit_needed"], json!([]), "{got}");
}

// SPEC-20261125-wip-limit-live-leases: the limit counts only the cards whose lease is alive.

/// Runs `statement` against the temporary ledger (the only way to end a lease: its length is at
/// least 60 seconds).
fn sql(w: &World, statement: &str) {
    let mut l = ledger::open_write(w.home.path(), "demo").unwrap();
    l.with_write_tx(|tx, _| {
        tx.execute_batch(statement)?;
        Ok(())
    })
    .unwrap();
}

fn card_id(card: &str) -> i64 {
    card.trim_start_matches("C-").parse().unwrap()
}

/// Ends the lease of the reservation of `card` (held, or held for its review).
fn end_lease(w: &World, card: &str) {
    sql(
        w,
        &format!(
            "UPDATE reservation SET lease_until = unixepoch() - 5 \
             WHERE card_id = {} AND status IN ('held', 'held_for_review')",
            card_id(card)
        ),
    );
}

/// `card add` by `by` with a scope of its own.
fn add_by(w: &World, by: &str, title: &str) -> Output {
    w.card(
        None,
        &[
            "add",
            "--title",
            title,
            "--scope",
            &format!("src/{title}/"),
            "--ac",
            "one",
            "--by",
            by,
        ],
    )
}

/// AC-1 (REQ-1): cards in work whose leases ended do not fill the limit: a new `card add` and a
/// new `card take` pass although two cards are still `active`.
#[test]
fn ac1_cards_with_ended_leases_do_not_fill_the_limit() {
    let w = world(None);
    w.lane("agent-1", &sid(1));
    let held = w.working(&sid(1), 2);
    for card in &held {
        end_lease(&w, card);
    }
    result(&add_by(&w, "leader", "more"));
    let next = w.ready_card("src/next/");
    result(&w.take(&sid(1), &next));
}

/// AC-2 (REQ-1, REQ-2): cards with live leases still fill the limit, and with one card of the two
/// left out the next take passes while the one after it refuses again (the guard against counting
/// too little).
#[test]
fn ac2_live_cards_still_fill_the_limit() {
    let w = world(None);
    w.lane("agent-1", &sid(1));
    let held = w.working(&sid(1), 2);
    let (kind, message) = refusal(&add_by(&w, "leader", "full"));
    assert_eq!(kind, "conflict", "{message}");
    assert!(message.contains("work-in-progress limit"), "{message}");

    end_lease(&w, &held[0]);
    let third = w.ready_card("src/third/");
    let fourth = w.ready_card("src/fourth/");
    result(&w.take(&sid(1), &third));
    let (kind, message) = refusal(&w.take(&sid(1), &fourth));
    assert_eq!(kind, "conflict", "{message}");
    assert!(message.contains("work-in-progress limit"), "{message}");
    assert!(message.contains(&held[1]), "{message}");
    assert!(message.contains(&third), "{message}");
}

/// AC-3 (REQ-3, REQ-4): the refusal names the counted cards first and, apart from them, each
/// card left out with the end of its lease or the words `no reservation`.
#[test]
fn ac3_the_refusal_lists_the_cards_left_out_apart() {
    let w = world(None);
    w.lane("agent-1", &sid(1));
    let old = w.working(&sid(1), 2);
    end_lease(&w, &old[0]);
    sql(
        &w,
        &format!(
            "UPDATE reservation SET status = 'released' WHERE card_id = {}",
            card_id(&old[1])
        ),
    );
    let live = w.working(&sid(1), 2);
    let (kind, message) = refusal(&add_by(&w, "leader", "more"));
    assert_eq!(kind, "conflict", "{message}");
    let (counted, left_out) = message
        .split_once("not counted")
        .unwrap_or_else(|| panic!("the refusal has no `not counted` part: {message}"));
    for card in &live {
        assert!(counted.contains(card.as_str()), "{card} counted: {message}");
        assert!(!left_out.contains(card.as_str()), "{card}: {message}");
    }
    for card in &old {
        assert!(!counted.contains(card.as_str()), "{card}: {message}");
        assert!(
            left_out.contains(card.as_str()),
            "{card} left out: {message}"
        );
    }
    assert!(left_out.contains("lease ended"), "{message}");
    assert!(left_out.contains("no reservation"), "{message}");
}

/// AC-4 (REQ-5): `card next` answers `wip` with the counted number and the cards left out, and a
/// card in review whose lease ended stays in `audit_needed`.
#[test]
fn ac4_next_names_the_cards_left_out_and_keeps_the_audit_list() {
    let w = world(None);
    w.lane("agent-1", &sid(1));
    let held = w.working(&sid(1), 2);
    result(&w.card(Some(&sid(1)), &["review", &held[1], "--generation", "1"]));
    for card in &held {
        end_lease(&w, card);
    }
    let got = result(&w.next(&sid(3)));
    assert_eq!(got["wip"]["count"], json!(0), "{got}");
    assert_eq!(got["wip"]["at_limit"], json!(false), "{got}");
    let left: Vec<&str> = got["wip"]["not_counted"]
        .as_array()
        .unwrap_or_else(|| panic!("no `not_counted` list in wip: {got}"))
        .iter()
        .map(|c| c["name"].as_str().unwrap())
        .collect();
    assert_eq!(left, vec![held[0].as_str(), held[1].as_str()], "{got}");
    let audit: Vec<&str> = got["audit_needed"]
        .as_array()
        .unwrap()
        .iter()
        .map(|c| c["name"].as_str().unwrap())
        .collect();
    assert_eq!(audit, vec![held[1].as_str()], "{got}");
}

/// AC-5 (REQ-6): `wip_limit` 0 and a project with no live lane stay off, whatever the leases.
#[test]
fn ac5_the_off_switches_stay_off() {
    let w = world(Some(0));
    w.lane("agent-1", &sid(1));
    w.working(&sid(1), 3);
    result(&add_by(&w, "leader", "more"));

    let w = world(None);
    w.working(&sid(1), 3);
    result(&add_by(&w, "leader", "more"));
}

/// AC-2 (REQ-1, SPEC D-3): a card in review whose lease is alive still counts, so two live cards,
/// one of them in review, still fill the limit.
#[test]
fn ac2_a_review_card_with_a_live_lease_still_counts() {
    let w = world(None);
    w.lane("agent-1", &sid(1));
    let held = w.working(&sid(1), 2);
    result(&w.card(Some(&sid(1)), &["review", &held[1], "--generation", "1"]));
    let (kind, message) = refusal(&add_by(&w, "leader", "more"));
    assert_eq!(kind, "conflict", "{message}");
    assert!(message.contains(&held[1]), "{message}");
    assert!(!message.contains("not counted"), "{message}");
}

/// AC-2 (REQ-1): a reservation held for its review (what `recover` leaves) with a lease that has
/// not ended still counts, so the two live cards still fill the limit.
#[test]
fn ac2_a_reservation_held_for_review_with_a_live_lease_still_counts() {
    let w = world(None);
    w.lane("agent-1", &sid(1));
    let held = w.working(&sid(1), 2);
    sql(
        &w,
        &format!(
            "UPDATE reservation SET status = 'held_for_review' WHERE card_id = {}",
            card_id(&held[0])
        ),
    );
    let (kind, message) = refusal(&add_by(&w, "leader", "more"));
    assert_eq!(kind, "conflict", "{message}");
    assert!(message.contains(&held[0]), "{message}");
    assert!(!message.contains("not counted"), "{message}");
}

// SPEC-20261126-wip-limit-pass-review: a card in review whose current head has a recorded PASS is
// left out of the count.

/// The commits that hand-back notes and verdicts name: `base` is the first commit of the fixture
/// repository, `a` and `b` are two later ones.
struct Commits {
    base: String,
    a: String,
    b: String,
}

fn commits(w: &World) -> Commits {
    let head = || common::git::git(&w.project, &["rev-parse", "HEAD"]);
    let next = |message: &str| {
        git_in(
            &w.project,
            &["commit", "-q", "--allow-empty", "-m", message],
        );
        head()
    };
    let base = head();
    let a = next("change a");
    let b = next("change b");
    Commits { base, a, b }
}

/// The first line of a note that hands `head` of `card` back.
fn hand_back_line(card: &str, head: &str) -> String {
    format!("LANE-READY-INTEGRATION {card} head={head}")
}

/// Adds the lane's hand-back note of `head`, without the receipt gate.
fn hand_back(w: &World, card: &str, head: &str) {
    crate::ready_receipt::add_note_unchecked(
        w.home.path(),
        "demo",
        card,
        &hand_back_line(card, head),
    );
}

/// Adds one note of `kind` signed `author`.
fn note_by(w: &World, card: &str, kind: &str, author: &str, text: &str) {
    let mut l = ledger::open_write(w.home.path(), "demo").unwrap();
    ledger::workflow::add_note(&mut l, card_id(card), kind, author, None, text, None).unwrap();
}

/// `card verdict` of `head` by the session `reviewer`, which never held the card. Each call has
/// its own evidence, so none is a repeat of an earlier call (a repeat writes nothing).
fn verdict(w: &World, c: &Commits, reviewer: u64, card: &str, head: &str, word: &str) {
    static CALLS: std::sync::atomic::AtomicUsize = std::sync::atomic::AtomicUsize::new(0);
    let call = CALLS.fetch_add(1, std::sync::atomic::Ordering::Relaxed);
    let evidence = format!("read the diff, call {call}");
    let got = result(&w.card(
        Some(&sid(reviewer)),
        &[
            "verdict",
            card,
            "--head",
            head,
            "--base",
            &c.base,
            "--verdict",
            word,
            "--tier",
            "1",
            "--evidence",
            &evidence,
        ],
    ));
    assert_eq!(got["written"], json!(true), "{got}");
}

/// Writes one `card.verdict` event row with `payload` as it is: the command cannot write a row
/// that does not read.
fn verdict_row(w: &World, card: &str, payload: &Value) {
    sql(
        w,
        &format!(
            "INSERT INTO event (event_key, at, actor, card_id, kind, payload) \
             VALUES (lower(hex(randomblob(16))), unixepoch(), 'by-hand', {}, 'card.verdict', \
             '{payload}')",
            card_id(card)
        ),
    );
}

/// `card <verb> <card> --generation <generation>` by the lane session `sid(1)`.
fn lane_step(w: &World, verb: &str, card: &str, generation: &str) {
    result(&w.card(Some(&sid(1)), &[verb, card, "--generation", generation]));
}

/// A card taken by `sid(1)`, handed back with `head` and sent to review (generation 1).
fn handed_back(w: &World, scope: &str, head: &str) -> String {
    let card = w.ready_card(scope);
    result(&w.take(&sid(1), &card));
    hand_back(w, &card, head);
    lane_step(w, "review", &card, "1");
    card
}

/// The `wip` object of `card next`.
fn wip(w: &World) -> Value {
    result(&w.next(&sid(3)))["wip"].clone()
}

/// The `not_counted` entry of `card` in a `wip` object.
fn not_counted_entry<'a>(wip: &'a Value, card: &str) -> Option<&'a Value> {
    wip["not_counted"]
        .as_array()
        .unwrap_or_else(|| panic!("no `not_counted` list in wip: {wip}"))
        .iter()
        .find(|c| c["name"].as_str() == Some(card))
}

/// `card` is left out of the count for the PASS on `head`.
#[track_caller]
fn assert_passed(w: &World, card: &str, head: &str) {
    let got = wip(w);
    let entry = not_counted_entry(&got, card)
        .unwrap_or_else(|| panic!("{card} is counted, not left out: {got}"));
    assert_eq!(entry["reason"], json!("pass verdict"), "{got}");
    assert_eq!(entry["head"], json!(head), "{got}");
    assert_eq!(entry["lease_until"], Value::Null, "{got}");
}

/// `card` is not left out, and the count is `count`.
#[track_caller]
fn assert_counted(w: &World, card: &str, count: usize) {
    let got = wip(w);
    assert!(
        not_counted_entry(&got, card).is_none(),
        "{card} is left out, not counted: {got}"
    );
    assert_eq!(got["count"], json!(count), "{got}");
}

/// The positive control of the cases below: one lane, one card in review handed back with head
/// `a`, passed by the reviewer `sid(7)` and so left out.
#[track_caller]
fn passed_world() -> (World, Commits, String) {
    let w = world(None);
    let c = commits(&w);
    w.lane("agent-1", &sid(1));
    let card = handed_back(&w, "src/passed/", &c.a);
    verdict(&w, &c, 7, &card, &c.a, "pass");
    assert_passed(&w, &card, &c.a);
    (w, c, card)
}

/// SPEC-20261126-wip-limit-pass-review AC-1 (REQ-1, REQ-6, REQ-7): a passed card in review leaves
/// room for `card add` and `card take`, `card next` lists it with its head, and the refusal at the
/// limit names it apart from the counted cards.
#[test]
fn pass_review_ac1_a_passed_review_card_leaves_room_and_is_listed_apart() {
    let w = world(None);
    let c = commits(&w);
    w.lane("agent-1", &sid(1));
    let passed = handed_back(&w, "src/passed/", &c.a);
    let busy = w.ready_card("src/busy/");
    result(&w.take(&sid(1), &busy));
    verdict(&w, &c, 7, &passed, &c.a, "pass");

    let got = wip(&w);
    assert_eq!(got["count"], json!(1), "{got}");
    assert_eq!(got["at_limit"], json!(false), "{got}");
    assert!(not_counted_entry(&got, &busy).is_none(), "{got}");
    assert_passed(&w, &passed, &c.a);
    let got = wip(&w);
    let entry = not_counted_entry(&got, &passed).unwrap();
    assert_eq!(entry["state"], json!("review"), "{got}");

    result(&add_by(&w, "leader", "more"));
    let next = w.ready_card("src/next/");
    result(&w.take(&sid(1), &next));

    // Two cards in work now reach the limit: both commands refuse.
    let got = wip(&w);
    assert_eq!(got["count"], json!(2), "{got}");
    assert_eq!(got["at_limit"], json!(true), "{got}");
    let last = w.ready_card("src/last/");
    for out in [add_by(&w, "leader", "full"), w.take(&sid(1), &last)] {
        let (kind, message) = refusal(&out);
        assert_eq!(kind, "conflict", "{message}");
        assert!(message.contains("work-in-progress limit"), "{message}");
        assert!(!message.contains("(no live lease)"), "{message}");
        let (counted, left_out) = message
            .split_once("; not counted")
            .unwrap_or_else(|| panic!("the refusal has no `not counted` part: {message}"));
        for card in [&busy, &next] {
            assert!(counted.contains(card.as_str()), "{card} counted: {message}");
            assert!(!left_out.contains(card.as_str()), "{card}: {message}");
        }
        assert!(!counted.contains(&passed), "{passed}: {message}");
        assert!(left_out.contains(&passed), "{passed} left out: {message}");
        let why = format!("PASS on head {}", &c.a[..12]);
        assert!(left_out.contains(&why), "{message}");
    }
}

/// SPEC-20261126-wip-limit-pass-review AC-2 (REQ-1, REQ-4): a head is passed only while it has a
/// PASS, every reviewer's newest verdict on it is PASS and the card's newest verdict is PASS.
#[test]
fn pass_review_ac2_only_a_head_that_is_still_passed_leaves_the_count() {
    // No verdict: beside a passed card, a card handed back without a verdict counts.
    let (w, c, passed) = passed_world();
    let unjudged = handed_back(&w, "src/unjudged/", &c.b);
    assert_counted(&w, &unjudged, 1);
    assert_passed(&w, &passed, &c.a);
    verdict(&w, &c, 7, &unjudged, &c.b, "pass");
    assert_passed(&w, &unjudged, &c.b);

    // A FAIL after a PASS on the head.
    let (w, c, card) = passed_world();
    verdict(&w, &c, 7, &card, &c.a, "fail");
    assert_counted(&w, &card, 1);

    // One reviewer's FAIL, then another's PASS: the newest verdict is PASS, one reviewer's is not.
    let (w, c, card) = passed_world();
    verdict(&w, &c, 8, &card, &c.a, "fail");
    assert_counted(&w, &card, 1);
    verdict(&w, &c, 7, &card, &c.a, "pass");
    assert_counted(&w, &card, 1);
    verdict(&w, &c, 8, &card, &c.a, "pass");
    assert_passed(&w, &card, &c.a);

    // A PASS on the head, then a FAIL on another head.
    let (w, c, card) = passed_world();
    verdict(&w, &c, 8, &card, &c.b, "fail");
    assert_counted(&w, &card, 1);

    // A PASS on another head only: the PASS names `b`, then the card is handed back with `a`.
    let w = world(None);
    let c = commits(&w);
    w.lane("agent-1", &sid(1));
    let card = handed_back(&w, "src/other/", &c.b);
    verdict(&w, &c, 7, &card, &c.b, "pass");
    assert_passed(&w, &card, &c.b);
    hand_back(&w, &card, &c.a);
    assert_counted(&w, &card, 1);

    // A PASS on the head, then a PASS on another head: still left out for the current head.
    let (w, c, card) = passed_world();
    verdict(&w, &c, 8, &card, &c.b, "pass");
    assert_passed(&w, &card, &c.a);
}

/// SPEC-20261126-wip-limit-pass-review AC-3 (REQ-2, REQ-3): the current head is the one the newest
/// hand-back note names, and only while that note is newer than every reopen, take and reclaim.
#[test]
fn pass_review_ac3_only_the_head_of_the_current_hand_back_counts_as_passed() {
    // Reopened and handed back with another head.
    let (w, c, card) = passed_world();
    lane_step(&w, "reopen", &card, "1");
    hand_back(&w, &card, &c.b);
    lane_step(&w, "review", &card, "1");
    assert_counted(&w, &card, 1);

    // Reopened and sent to review without a note (`card.reopen` alone follows the note).
    let (w, c, card) = passed_world();
    lane_step(&w, "reopen", &card, "1");
    lane_step(&w, "review", &card, "1");
    assert_counted(&w, &card, 1);
    hand_back(&w, &card, &c.a);
    assert_passed(&w, &card, &c.a);

    // Released, taken again and sent to review without a note (`reserve` alone follows the note).
    // The leader releases no card whose newest verdict is PASS, so a FAIL comes first and the
    // PASS is recorded again after the new review.
    let (w, c, card) = passed_world();
    verdict(&w, &c, 8, &card, &c.a, "fail");
    assert_counted(&w, &card, 1);
    end_lease(&w, &card);
    w.leader("leader", &sid(9));
    result(&w.card(
        Some(&sid(9)),
        &[
            "release",
            &card,
            "--generation",
            "1",
            "--by",
            "leader",
            "--reason",
            "the lane gave the card up",
        ],
    ));
    result(&w.take(&sid(1), &card));
    lane_step(&w, "review", &card, "2");
    verdict(&w, &c, 8, &card, &c.a, "pass");
    assert_counted(&w, &card, 1);
    hand_back(&w, &card, &c.a);
    assert_passed(&w, &card, &c.a);

    // Reclaimed in review (`card.reclaim` alone follows the note).
    let (w, c, card) = passed_world();
    end_lease(&w, &card);
    result(&w.card(
        Some(&sid(1)),
        &[
            "reclaim",
            &card,
            "--generation",
            "1",
            "--by",
            "leader",
            "--reason",
            "the lease ended while the card waited",
        ],
    ));
    assert_counted(&w, &card, 1);
    hand_back(&w, &card, &c.a);
    assert_passed(&w, &card, &c.a);

    // No hand-back note: beside a passed card, a card with a PASS and no note counts; the note
    // is all it lacks.
    let (w, c, passed) = passed_world();
    let bare = w.ready_card("src/bare/");
    result(&w.take(&sid(1), &bare));
    lane_step(&w, "review", &bare, "1");
    verdict(&w, &c, 7, &bare, &c.b, "pass");
    assert_counted(&w, &bare, 1);
    assert_passed(&w, &passed, &c.a);
    hand_back(&w, &bare, &c.b);
    assert_passed(&w, &bare, &c.b);

    // The newest hand-back note does not parse: the older, passed one is not read.
    let (w, _c, card) = passed_world();
    let unparsed = hand_back_line(&card, "nothex");
    note_by(&w, &card, "report", "lane", &unparsed);
    assert_counted(&w, &card, 1);

    // The newest hand-back note has no `card.note` event: it cannot be placed after the take.
    let (w, c, card) = passed_world();
    sql(
        &w,
        &format!(
            "INSERT INTO card_note (card_id, kind, author, body, created_at) \
             VALUES ({}, 'report', 'lane', '{}', unixepoch())",
            card_id(&card),
            hand_back_line(&card, &c.a)
        ),
    );
    assert_counted(&w, &card, 1);

    // A later note by the leader, of kind `decision` or without a marker changes nothing.
    let (w, c, card) = passed_world();
    let quoted = hand_back_line(&card, &c.b);
    note_by(&w, &card, "report", "leader", &quoted);
    assert_passed(&w, &card, &c.a);
    note_by(&w, &card, "decision", "lane", &quoted);
    assert_passed(&w, &card, &c.a);
    note_by(&w, &card, "comment", "lane", "PLAN wait for the landing");
    assert_passed(&w, &card, &c.a);

    // The same head handed back again after a reopen keeps its PASS.
    let (w, c, card) = passed_world();
    lane_step(&w, "reopen", &card, "1");
    hand_back(&w, &card, &c.a);
    lane_step(&w, "review", &card, "1");
    assert_passed(&w, &card, &c.a);
}

/// SPEC-20261126-wip-limit-pass-review AC-4 (REQ-5): a verdict row that cannot be read beside a
/// PASS makes the card count.
#[test]
fn pass_review_ac4_a_verdict_row_that_cannot_be_read_makes_the_card_count() {
    let by_hand = |c: &Commits| {
        json!({
            "head": c.a, "base": c.base, "verdict": "pass", "tier": 1,
            "evidence": "by hand", "receipt": null, "reviewer": sid(8),
        })
    };

    // A row written by hand that reads as a PASS changes nothing.
    let (w, c, card) = passed_world();
    verdict_row(&w, &card, &by_hand(&c));
    assert_passed(&w, &card, &c.a);

    // No verdict word, a word that is no verdict, no head.
    for (field, value) in [
        ("verdict", None),
        ("verdict", Some(json!("maybe"))),
        ("head", None),
    ] {
        let (w, c, card) = passed_world();
        let mut payload = by_hand(&c);
        match value {
            Some(v) => payload[field] = v,
            None => {
                payload.as_object_mut().unwrap().remove(field);
            }
        }
        verdict_row(&w, &card, &payload);
        assert_counted(&w, &card, 1);
        // A later PASS does not hide the older row that cannot be read.
        verdict(&w, &c, 7, &card, &c.a, "pass");
        assert_counted(&w, &card, 1);
    }
}

/// SPEC-20261126-wip-limit-pass-review AC-5 (REQ-8): only a card in review is left out for a PASS;
/// the limit's numbers, the off switch and the two lease reasons stay.
#[test]
fn pass_review_ac5_only_a_review_card_changes_and_the_rest_stays() {
    let w = world(None);
    let c = commits(&w);
    w.lane("agent-1", &sid(1));
    let reviewed = handed_back(&w, "src/reviewed/", &c.a);
    assert_counted(&w, &reviewed, 1);
    let before = wip(&w);
    verdict(&w, &c, 7, &reviewed, &c.a, "pass");
    assert_passed(&w, &reviewed, &c.a);
    let after = wip(&w);
    for (key, number) in [("per_lane", 2), ("lanes", 1), ("limit", 2)] {
        assert_eq!(before[key], json!(number), "{before}");
        assert_eq!(after[key], json!(number), "{after}");
    }

    // A card in work that never went to review counts although its head has a PASS.
    let active = w.ready_card("src/active/");
    result(&w.take(&sid(1), &active));
    hand_back(&w, &active, &c.b);
    verdict(&w, &c, 7, &active, &c.b, "pass");
    assert_counted(&w, &active, 1);
    assert_passed(&w, &reviewed, &c.a);
    lane_step(&w, "review", &active, "1");
    assert_passed(&w, &active, &c.b);

    // A reopened card counts, also with a new hand-back note of the passed head.
    lane_step(&w, "reopen", &reviewed, "1");
    assert_counted(&w, &reviewed, 1);
    hand_back(&w, &reviewed, &c.a);
    assert_counted(&w, &reviewed, 1);
    lane_step(&w, "review", &reviewed, "1");
    assert_passed(&w, &reviewed, &c.a);

    // An ended lease and a released reservation keep their reasons, a PASS or not.
    end_lease(&w, &reviewed);
    sql(
        &w,
        &format!(
            "UPDATE reservation SET status = 'released' WHERE card_id = {}",
            card_id(&active)
        ),
    );
    let got = wip(&w);
    assert_eq!(got["count"], json!(0), "{got}");
    let ended = not_counted_entry(&got, &reviewed).unwrap_or_else(|| panic!("{got}"));
    assert_eq!(ended["reason"], json!("lease ended"), "{got}");
    assert!(ended["lease_until"].is_i64(), "{got}");
    let released = not_counted_entry(&got, &active).unwrap_or_else(|| panic!("{got}"));
    assert_eq!(released["reason"], json!("no reservation"), "{got}");
    assert_eq!(released["lease_until"], Value::Null, "{got}");

    // `wip_limit` 0 stays off.
    let w = world(Some(0));
    let c = commits(&w);
    w.lane("agent-1", &sid(1));
    let card = handed_back(&w, "src/off/", &c.a);
    verdict(&w, &c, 7, &card, &c.a, "pass");
    w.working(&sid(1), 3);
    result(&add_by(&w, "leader", "more"));
    let got = wip(&w);
    assert_eq!(got["limit"], json!(0), "{got}");
    assert_eq!(got["at_limit"], json!(false), "{got}");
}

// SPEC-20261126-audit-needed-current-head: `card next` leaves a card in review out of the audit
// list only for a verdict on its current head.

/// The names `card next` lists in `audit_needed` when `session` asks.
fn audit_names(w: &World, session: &str) -> Vec<String> {
    let got = result(&w.next(session));
    got["audit_needed"]
        .as_array()
        .unwrap_or_else(|| panic!("no `audit_needed` list: {got}"))
        .iter()
        .map(|c| c["name"].as_str().unwrap().to_string())
        .collect()
}

/// SPEC-20261126-audit-needed-current-head AC-1 (REQ-1, REQ-2): a card judged on head `a` is
/// listed again once head `b` is handed back, whether it was reopened or got a `REPAIRED` note.
#[test]
fn audit_head_ac1_a_new_head_after_a_verdict_is_listed() {
    let mut not_listed = Vec::new();
    for word in ["pass", "fail"] {
        for reopened in [true, false] {
            let w = world(Some(0));
            let c = commits(&w);
            let card = handed_back(&w, "src/resubmitted/", &c.a);
            assert_eq!(audit_names(&w, &sid(3)), vec![card.clone()], "unjudged");
            verdict(&w, &c, 7, &card, &c.a, word);
            assert_eq!(
                audit_names(&w, &sid(3)),
                Vec::<String>::new(),
                "{word} on a"
            );

            if reopened {
                lane_step(&w, "reopen", &card, "1");
                hand_back(&w, &card, &c.b);
                lane_step(&w, "review", &card, "1");
            } else {
                let repaired = format!("REPAIRED {card} head={}", c.b);
                note_by(&w, &card, "report", "lane", &repaired);
            }
            let got = audit_names(&w, &sid(3));
            if got != vec![card] {
                not_listed.push(format!("{word} on a, reopened {reopened}: {got:?}"));
            }
        }
    }
    assert!(
        not_listed.is_empty(),
        "head b has no verdict, yet the card is not listed: {not_listed:#?}"
    );
}

/// SPEC-20261126-audit-needed-current-head AC-1 (REQ-1, REQ-2): at the limit `card next` offers
/// the resubmitted card for an audit.
#[test]
fn audit_head_ac1_at_the_limit_next_offers_the_resubmitted_card() {
    let w = world(None);
    let c = commits(&w);
    w.lane("agent-1", &sid(1));
    let resubmitted = handed_back(&w, "src/resubmitted/", &c.a);
    let busy = w.ready_card("src/busy/");
    result(&w.take(&sid(1), &busy));
    verdict(&w, &c, 7, &resubmitted, &c.a, "fail");
    lane_step(&w, "reopen", &resubmitted, "1");
    hand_back(&w, &resubmitted, &c.b);
    lane_step(&w, "review", &resubmitted, "1");

    let got = result(&w.next(&sid(3)));
    assert_eq!(got["wip"]["count"], json!(2), "{got}");
    assert_eq!(got["wip"]["at_limit"], json!(true), "{got}");
    assert_eq!(got["next_is"], json!("audit"), "{got}");
    assert_eq!(got["next"], json!(resubmitted), "{got}");
    assert_eq!(audit_names(&w, &sid(3)), vec![resubmitted], "{got}");
}

/// SPEC-20261126-audit-needed-current-head AC-2 (REQ-2): any verdict on the current head, of any
/// reviewer, hides the card, and nothing that leaves the head as it is brings it back. The card
/// without a verdict beside it stays listed.
#[test]
fn audit_head_ac2_a_verdict_on_the_current_head_hides_the_card() {
    // A card handed back with head `a` and judged on it by each of `verdicts`, beside a card
    // without a verdict, which alone is listed.
    let judged_world = |verdicts: &[(u64, &str)]| {
        let w = world(Some(0));
        let c = commits(&w);
        let judged = handed_back(&w, "src/judged/", &c.a);
        let unjudged = handed_back(&w, "src/unjudged/", &c.b);
        assert_eq!(
            audit_names(&w, &sid(3)),
            vec![judged.clone(), unjudged.clone()],
            "before any verdict"
        );
        for (reviewer, word) in verdicts {
            verdict(&w, &c, *reviewer, &judged, &c.a, word);
        }
        assert_eq!(
            audit_names(&w, &sid(3)),
            vec![unjudged.clone()],
            "{verdicts:?}"
        );
        (w, c, judged, unjudged)
    };

    // PASS, FAIL, and two reviewers who disagree, in both orders.
    judged_world(&[(7, "pass")]);
    judged_world(&[(7, "fail")]);
    judged_world(&[(7, "pass"), (8, "fail")]);
    judged_world(&[(7, "fail"), (8, "pass")]);

    for (word, later) in [("pass", "fail"), ("fail", "pass")] {
        // A later verdict on another head.
        let (w, c, judged, unjudged) = judged_world(&[(7, word)]);
        verdict(&w, &c, 8, &judged, &c.b, later);
        assert_eq!(
            audit_names(&w, &sid(3)),
            vec![unjudged],
            "{word} on a, then {later} on b"
        );

        // A later note by the leader or of kind `decision` that names another head.
        let (w, c, judged, unjudged) = judged_world(&[(7, word)]);
        let quoted = hand_back_line(&judged, &c.b);
        note_by(&w, &judged, "report", "leader", &quoted);
        assert_eq!(
            audit_names(&w, &sid(3)),
            vec![unjudged.clone()],
            "{word}, then a leader note"
        );
        note_by(&w, &judged, "decision", "lane", &quoted);
        assert_eq!(
            audit_names(&w, &sid(3)),
            vec![unjudged],
            "{word}, then a decision note"
        );

        // The same head handed back again, after a reopen and by a `REPAIRED` note in review.
        let (w, c, judged, unjudged) = judged_world(&[(7, word)]);
        lane_step(&w, "reopen", &judged, "1");
        hand_back(&w, &judged, &c.a);
        lane_step(&w, "review", &judged, "1");
        assert_eq!(
            audit_names(&w, &sid(3)),
            vec![unjudged.clone()],
            "{word}, then head a again after a reopen"
        );
        let repaired = format!("REPAIRED {judged} head={}", c.a);
        note_by(&w, &judged, "report", "lane", &repaired);
        assert_eq!(
            audit_names(&w, &sid(3)),
            vec![unjudged],
            "{word}, then a REPAIRED note of head a"
        );
    }
}

/// SPEC-20261126-audit-needed-current-head AC-3 (REQ-3): a card whose current head cannot be
/// confirmed is listed although a verdict names the head it was handed back with.
#[test]
fn audit_head_ac3_a_card_whose_current_head_cannot_be_confirmed_is_listed() {
    // One card handed back with head `a` and passed on it, so hidden.
    let passed = || {
        let w = world(Some(0));
        let c = commits(&w);
        let card = handed_back(&w, "src/passed/", &c.a);
        verdict(&w, &c, 7, &card, &c.a, "pass");
        assert_eq!(audit_names(&w, &sid(3)), Vec::<String>::new(), "pass on a");
        (w, c, card)
    };
    let mut not_listed = Vec::new();
    let mut expect_listed = |case: &str, w: &World, card: &str| {
        let got = audit_names(w, &sid(3));
        if got != vec![card.to_string()] {
            not_listed.push(format!("{case}: {got:?}"));
        }
    };

    let (w, _c, card) = passed();
    lane_step(&w, "reopen", &card, "1");
    lane_step(&w, "review", &card, "1");
    expect_listed("review without a new note after a reopen", &w, &card);

    let (w, _c, card) = passed();
    let unparsed = hand_back_line(&card, "nothex");
    note_by(&w, &card, "report", "lane", &unparsed);
    expect_listed("the newest hand-back note does not parse", &w, &card);

    let (w, c, card) = passed();
    sql(
        &w,
        &format!(
            "INSERT INTO card_note (card_id, kind, author, body, created_at) \
             VALUES ({}, 'report', 'lane', '{}', unixepoch())",
            card_id(&card),
            hand_back_line(&card, &c.a)
        ),
    );
    expect_listed("a note without its card.note event", &w, &card);

    let w = world(Some(0));
    let c = commits(&w);
    let bare = w.ready_card("src/bare/");
    result(&w.take(&sid(1), &bare));
    lane_step(&w, "review", &bare, "1");
    verdict(&w, &c, 7, &bare, &c.a, "pass");
    expect_listed("no hand-back note", &w, &bare);

    let (w, _c, card) = passed();
    end_lease(&w, &card);
    result(&w.card(
        Some(&sid(1)),
        &[
            "reclaim",
            &card,
            "--generation",
            "1",
            "--by",
            "leader",
            "--reason",
            "the lease ended while the card waited",
        ],
    ));
    expect_listed("after a reclaim", &w, &card);

    assert!(
        not_listed.is_empty(),
        "the current head cannot be confirmed, yet the card is not listed: {not_listed:#?}"
    );
}

/// SPEC-20261126-audit-needed-current-head AC-4 (REQ-4): a verdict row that cannot be read keeps
/// the card listed, alone, beside a PASS on the head and before a later PASS.
#[test]
fn audit_head_ac4_a_card_with_an_unreadable_verdict_row_is_listed() {
    let by_hand = |c: &Commits| {
        json!({
            "head": c.a, "base": c.base, "verdict": "pass", "tier": 1,
            "evidence": "by hand", "receipt": null, "reviewer": sid(8),
        })
    };
    // One card handed back with head `a`, without a verdict and so listed.
    let unjudged = || {
        let w = world(Some(0));
        let c = commits(&w);
        let card = handed_back(&w, "src/judged/", &c.a);
        assert_eq!(audit_names(&w, &sid(3)), vec![card.clone()], "unjudged");
        (w, c, card)
    };

    // A row written by hand that reads as a PASS on the head hides the card.
    let (w, c, card) = unjudged();
    verdict_row(&w, &card, &by_hand(&c));
    assert_eq!(
        audit_names(&w, &sid(3)),
        Vec::<String>::new(),
        "a readable row written by hand"
    );

    let mut not_listed = Vec::new();
    let mut expect_listed = |case: String, w: &World, card: &str| {
        let got = audit_names(w, &sid(3));
        if got != vec![card.to_string()] {
            not_listed.push(format!("{case}: {got:?}"));
        }
    };
    // No verdict word, a word that is no verdict, no head.
    for (defect, field, value) in [
        ("no verdict word", "verdict", None),
        ("the word maybe", "verdict", Some(json!("maybe"))),
        ("no head", "head", None),
    ] {
        let unreadable = |c: &Commits| {
            let mut payload = by_hand(c);
            match &value {
                Some(v) => payload[field] = v.clone(),
                None => {
                    payload.as_object_mut().unwrap().remove(field);
                }
            }
            payload
        };

        let (w, c, card) = unjudged();
        verdict_row(&w, &card, &unreadable(&c));
        expect_listed(format!("{defect}, alone"), &w, &card);

        let (w, c, card) = unjudged();
        verdict(&w, &c, 7, &card, &c.a, "pass");
        assert_eq!(audit_names(&w, &sid(3)), Vec::<String>::new(), "pass on a");
        verdict_row(&w, &card, &unreadable(&c));
        expect_listed(format!("{defect}, beside a PASS on the head"), &w, &card);

        let (w, c, card) = unjudged();
        verdict_row(&w, &card, &unreadable(&c));
        verdict(&w, &c, 7, &card, &c.a, "pass");
        expect_listed(format!("{defect}, before a later PASS"), &w, &card);
    }
    assert!(
        not_listed.is_empty(),
        "a verdict row cannot be read, yet the card is not listed: {not_listed:#?}"
    );
}

/// SPEC-20261126-audit-needed-current-head AC-5 (REQ-5): the session that holds a resubmitted
/// card is not asked to audit it; another session is.
#[test]
fn audit_head_ac5_the_holder_is_not_asked_to_audit_its_own_card() {
    let w = world(Some(0));
    let c = commits(&w);
    let card = handed_back(&w, "src/mine/", &c.a);
    verdict(&w, &c, 7, &card, &c.a, "fail");
    lane_step(&w, "reopen", &card, "1");
    hand_back(&w, &card, &c.b);
    lane_step(&w, "review", &card, "1");

    assert_eq!(audit_names(&w, &sid(1)), Vec::<String>::new(), "the holder");
    assert_eq!(audit_names(&w, &sid(3)), vec![card], "another session");
}

// SPEC-20261125-wip-limit-live-lock-lanes: a lane is live while its launcher holds the
// `<name>.live` lock or its registration lease runs, and a registry that cannot be read does not
// switch the limit off.

/// The sessions folder of the fixture project (the registry and the `.live` locks).
fn sessions_dir(w: &World) -> PathBuf {
    moai_cli::lane_registry::registry_dir_unlinked(&w.project).unwrap()
}

/// Holds the `<name>.live` lock the way a launcher does; the lock ends when the file is dropped.
fn hold_lock(w: &World, name: &str) -> fs::File {
    moai_cli::session::lock_named_live(&sessions_dir(w), name).unwrap()
}

/// REQ-1 (AC-1): three lanes whose registration leases ended but whose launchers hold their locks
/// are three live lanes, so the limit is six: six cards are taken and the seventh refuses.
#[test]
fn lock_ac1_lanes_with_a_held_lock_count_after_their_lease_ended() {
    let w = world(None);
    let mut locks = Vec::new();
    for n in 1..=3u64 {
        let name = format!("agent-{n}");
        w.lane(&name, &sid(n));
        w.expire(&name);
        locks.push(hold_lock(&w, &name));
    }
    w.working(&sid(1), 6);
    let seventh = w.ready_card("src/seventh/");
    let (kind, message) = refusal(&w.take(&sid(1), &seventh));
    assert_eq!(kind, "conflict", "{message}");
    assert!(message.contains("3 lane(s)"), "{message}");
    assert!(message.contains("limit is 6"), "{message}");
}

/// REQ-1 (AC-2): a lane with no lock and an ended lease stays out, its lock makes it count, and
/// letting the lock go takes it out again (the guard against counting too much).
#[test]
fn lock_ac2_a_lane_counts_only_while_its_lock_is_held() {
    let w = world(None);
    w.lane("agent-1", &sid(1));
    w.lane("agent-2", &sid(2));
    w.expire("agent-2");
    w.working(&sid(1), 2);
    let third = w.ready_card("src/third/");
    let (_, message) = refusal(&w.take(&sid(1), &third));
    assert!(message.contains("1 lane(s)"), "{message}");

    let lock = hold_lock(&w, "agent-2");
    result(&w.take(&sid(1), &third));
    let fourth = w.ready_card("src/fourth/");
    result(&w.take(&sid(1), &fourth));
    let fifth = w.ready_card("src/fifth/");
    let (_, message) = refusal(&w.take(&sid(1), &fifth));
    assert!(message.contains("2 lane(s)"), "{message}");

    drop(lock);
    let (_, message) = refusal(&w.take(&sid(1), &fifth));
    assert!(message.contains("1 lane(s)"), "{message}");
}

/// REQ-2, REQ-3 (AC-3): a leader entry with a held lock and an ended lease, the project's leader
/// name and a name no lane has all stay out; a held lock named like a lane of this project counts
/// although it has no entry.
#[test]
fn lock_ac3_leaders_and_unknown_names_do_not_count_but_a_lane_name_does() {
    let w = world(None);
    w.lane("agent-1", &sid(1));
    w.leader("leader-1", &sid(9));
    w.expire("leader-1");
    let _leader_lock = hold_lock(&w, "leader-1");
    let _named_leader = hold_lock(&w, &moai_cli::session::leader_session_name(&w.project));
    let _unknown = hold_lock(&w, "misc-session");
    w.working(&sid(1), 2);
    let third = w.ready_card("src/third/");
    let (_, message) = refusal(&w.take(&sid(1), &third));
    assert!(message.contains("1 lane(s)"), "{message}");

    let lane_name = format!("{}zed", moai_cli::session::agent_prefix(&w.project));
    let _lane_lock = hold_lock(&w, &lane_name);
    result(&w.take(&sid(1), &third));
    let fourth = w.ready_card("src/fourth/");
    result(&w.take(&sid(1), &fourth));
    let fifth = w.ready_card("src/fifth/");
    let (_, message) = refusal(&w.take(&sid(1), &fifth));
    assert!(message.contains("2 lane(s)"), "{message}");
}

/// SPEC-20261127-launcher-session-names REQ-11 (AC-7): the leader is read as `leader-<project
/// name>` or `<project name>-leader`, so an entry of either name is no lane even without the
/// leader mode; a lane is read as `agent-<name>` or `<project name>-agent-<name>`, and a lane that
/// holds both forms counts once.
#[test]
fn lock_ac7_both_leader_names_are_no_lane_and_both_lane_forms_are_one_lane() {
    let w = world(None);
    w.lane("agent-1", &sid(1));
    w.lane("leader-demo", &sid(9));
    w.lane("demo-leader", &sid(8));
    w.working(&sid(1), 2);
    let third = w.ready_card("src/third/");
    let (_, message) = refusal(&w.take(&sid(1), &third));
    assert!(message.contains("1 lane(s)"), "{message}");

    // One lane that holds both forms of its name counts once: two lanes allow four cards.
    let _short = hold_lock(&w, "agent-ben");
    let _long = hold_lock(&w, "demo-agent-ben");
    result(&w.take(&sid(1), &third));
    let fourth = w.ready_card("src/fourth/");
    result(&w.take(&sid(1), &fourth));
    let fifth = w.ready_card("src/fifth/");
    let (_, message) = refusal(&w.take(&sid(1), &fifth));
    assert!(message.contains("2 lane(s)"), "{message}");
}

/// REQ-4, REQ-6, REQ-7 (AC-4): with the sessions folder replaced by a link the limit is not off:
/// `card take` refuses and names the registry, `card next` answers `registry_error` at the limit
/// and names no card, and `wip_limit` 0 still takes.
#[test]
fn lock_ac4_a_registry_folder_that_cannot_be_read_stops_the_take() {
    let w = world(None);
    w.lane("agent-1", &sid(1));
    let card = w.ready_card("src/waiting/");
    let dir = sessions_dir(&w);
    let elsewhere = w.home.path().join("elsewhere");
    fs::create_dir_all(&elsewhere).unwrap();
    fs::remove_dir_all(&dir).unwrap();
    std::os::unix::fs::symlink(&elsewhere, &dir).unwrap();

    let (kind, message) = refusal(&w.take(&sid(1), &card));
    assert_eq!(kind, "conflict", "{message}");
    assert!(message.contains("session registry"), "{message}");
    assert!(message.contains("symlink"), "{message}");
    let (kind, message) = refusal(&add_by(&w, "leader", "more"));
    assert_eq!(kind, "conflict", "{message}");

    let got = result(&w.next(&sid(1)));
    assert!(got["wip"]["registry_error"].is_string(), "{got}");
    assert_eq!(got["wip"]["at_limit"], json!(true), "{got}");
    assert_eq!(got["next_is"], json!("limit"), "{got}");
    assert_eq!(got["next"], Value::Null, "{got}");

    w.set_limit(Some(0));
    result(&w.take(&sid(1), &card));
}

/// SPEC-20261125-wip-limit-next-registry-error REQ-1: a registry folder that cannot be read
/// leaves `card next` with no card at all, a card that waits for an audit included, because the
/// answer says the limit is reached and `card take` refuses in the same case.
#[test]
fn registry_error_next_names_no_card_not_even_one_that_waits_for_an_audit() {
    let w = world(None);
    w.lane("agent-1", &sid(1));
    let held = w.working(&sid(1), 1);
    result(&w.card(Some(&sid(1)), &["review", &held[0], "--generation", "1"]));
    let dir = sessions_dir(&w);
    let elsewhere = w.home.path().join("elsewhere");
    fs::create_dir_all(&elsewhere).unwrap();
    fs::remove_dir_all(&dir).unwrap();
    std::os::unix::fs::symlink(&elsewhere, &dir).unwrap();

    let got = result(&w.next(&sid(3)));
    assert!(got["wip"]["registry_error"].is_string(), "{got}");
    assert_eq!(got["next_is"], json!("limit"), "{got}");
    assert_eq!(got["next"], Value::Null, "{got}");
    assert_eq!(got["audit_needed"], json!([]), "{got}");
}

/// REQ-5, REQ-6 (AC-5): one malformed entry beside two live lanes leaves the lane count at two,
/// does not switch the limit off, and is listed as unreadable in the refusal and in `card next`.
#[test]
fn lock_ac5_one_broken_entry_is_left_out_and_listed() {
    let w = world(None);
    w.lane("agent-1", &sid(1));
    w.lane("agent-2", &sid(2));
    fs::write(sessions_dir(&w).join("agent-9.json"), "{not json").unwrap();
    w.working(&sid(1), 4);
    let fifth = w.ready_card("src/fifth/");
    let (kind, message) = refusal(&w.take(&sid(1), &fifth));
    assert_eq!(kind, "conflict", "{message}");
    assert!(message.contains("2 lane(s)"), "{message}");
    assert!(message.contains("unreadable"), "{message}");
    assert!(message.contains("agent-9"), "{message}");

    let got = result(&w.next(&sid(3)));
    assert_eq!(got["wip"]["registry_error"], Value::Null, "{got}");
    assert_eq!(got["wip"]["unreadable"], json!(["agent-9"]), "{got}");
    assert_eq!(got["wip"]["lanes"], json!(2), "{got}");
}

/// REQ-8 (AC-6): `wip_limit` 0 stays off with locks held, and a registry whose only lane has an
/// ended lease and no lock stays off too.
#[test]
fn lock_ac6_the_off_switches_stay_off() {
    let w = world(Some(0));
    w.lane("agent-1", &sid(1));
    let _lock = hold_lock(&w, "agent-1");
    w.working(&sid(1), 3);

    let w = world(None);
    w.lane("agent-1", &sid(1));
    w.expire("agent-1");
    w.working(&sid(1), 3);
    let got = result(&w.next(&sid(3)));
    assert_eq!(got["wip"]["registry_error"], Value::Null, "{got}");
    assert_eq!(got["wip"]["unreadable"], json!([]), "{got}");
}

// SPEC-20261126-wip-limit-all-unreadable: a registry that holds items but none that can be read
// does not switch the limit off.

/// REQ-1 (AC-1): a registry whose only items are malformed entries does not switch the limit off:
/// `card take` and `card add` refuse with the registry text that names each entry, and `card next`
/// answers `registry_error` as a string, `at_limit` true, `unreadable` with the names and no card.
#[test]
fn all_unreadable_ac1_a_registry_with_only_unreadable_items_stops_the_take() {
    let w = world(None);
    fs::create_dir_all(sessions_dir(&w)).unwrap();
    fs::write(sessions_dir(&w).join("agent-9.json"), "{not json").unwrap();
    let card = w.ready_card("src/waiting/");

    let (kind, message) = refusal(&w.take(&sid(1), &card));
    assert_eq!(kind, "conflict", "{message}");
    assert!(message.contains("session registry"), "{message}");
    assert!(message.contains("agent-9"), "{message}");
    let (kind, message) = refusal(&add_by(&w, "leader", "more"));
    assert_eq!(kind, "conflict", "{message}");
    assert!(message.contains("session registry"), "{message}");
    assert!(message.contains("agent-9"), "{message}");

    let got = result(&w.next(&sid(1)));
    assert!(got["wip"]["registry_error"].is_string(), "{got}");
    assert_eq!(got["wip"]["at_limit"], json!(true), "{got}");
    assert_eq!(got["wip"]["unreadable"], json!(["agent-9"]), "{got}");
    assert_eq!(got["next_is"], json!("limit"), "{got}");
    assert_eq!(got["next"], Value::Null, "{got}");

    // Two unreadable items: the refusal names both.
    let w = world(None);
    fs::create_dir_all(sessions_dir(&w)).unwrap();
    fs::write(sessions_dir(&w).join("agent-8.json"), "{not json").unwrap();
    fs::write(sessions_dir(&w).join("agent-9.json"), "{not json").unwrap();
    let card = w.ready_card("src/waiting/");
    let (kind, message) = refusal(&w.take(&sid(1), &card));
    assert_eq!(kind, "conflict", "{message}");
    assert!(message.contains("session registry"), "{message}");
    assert!(message.contains("agent-8"), "{message}");
    assert!(message.contains("agent-9"), "{message}");
}

/// REQ-2 (AC-2): an empty registry folder, a readable entry whose lease ended beside a malformed
/// entry, and a live lane beside a malformed entry each keep the answer they had before.
#[test]
fn all_unreadable_ac2_a_readable_item_or_no_item_keeps_the_earlier_answer() {
    // (a) The sessions folder exists and holds nothing: the limit is off, no registry error.
    let w = world(None);
    fs::create_dir_all(sessions_dir(&w)).unwrap();
    w.working(&sid(1), 3);
    let got = result(&w.next(&sid(3)));
    assert_eq!(got["wip"]["registry_error"], Value::Null, "{got}");
    assert_eq!(got["wip"]["lanes"], json!(0), "{got}");

    // (b) One readable entry whose lease ended (no lock) beside a malformed entry: no live lane,
    // so the limit stays off; the malformed entry is only named.
    let w = world(None);
    w.lane("agent-1", &sid(1));
    w.expire("agent-1");
    fs::write(sessions_dir(&w).join("agent-9.json"), "{not json").unwrap();
    w.working(&sid(1), 3);
    let got = result(&w.next(&sid(3)));
    assert_eq!(got["wip"]["registry_error"], Value::Null, "{got}");
    assert_eq!(got["wip"]["unreadable"], json!(["agent-9"]), "{got}");
    assert_eq!(got["wip"]["lanes"], json!(0), "{got}");

    // (c) One live lane beside a malformed entry: the limit is two and the third take refuses
    // with the ordinary limit text that also lists the unreadable entry.
    let w = world(None);
    w.lane("agent-1", &sid(1));
    fs::write(sessions_dir(&w).join("agent-9.json"), "{not json").unwrap();
    w.working(&sid(1), 2);
    let third = w.ready_card("src/third/");
    let (kind, message) = refusal(&w.take(&sid(1), &third));
    assert_eq!(kind, "conflict", "{message}");
    assert!(message.contains("limit is 2"), "{message}");
    assert!(message.contains("agent-9"), "{message}");
    let got = result(&w.next(&sid(3)));
    assert_eq!(got["wip"]["registry_error"], Value::Null, "{got}");
}

/// REQ-3 (AC-3): with `wip_limit` 0 a registry whose only item is malformed takes and adds without
/// a refusal, and `card next` does not report the limit as reached.
#[test]
fn all_unreadable_ac3_wip_limit_zero_stays_off() {
    let w = world(Some(0));
    fs::create_dir_all(sessions_dir(&w)).unwrap();
    fs::write(sessions_dir(&w).join("agent-9.json"), "{not json").unwrap();
    let card = w.ready_card("src/waiting/");
    result(&w.take(&sid(1), &card));
    result(&add_by(&w, "leader", "more"));
    let got = result(&w.next(&sid(1)));
    assert_eq!(got["wip"]["at_limit"], json!(false), "{got}");
}
