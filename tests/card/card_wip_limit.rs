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
        let file = dir.join(format!("{name}.json"));
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
