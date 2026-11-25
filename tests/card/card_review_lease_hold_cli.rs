//! SPEC-20261125-review-lease-hold: a card in `review` keeps its lease until the verdict is
//! recorded, and a reclaimed or reopened reservation brings back the card's own scope. Every
//! command runs the built binary against a temporary HOME and a temporary repository; the real
//! home state folder is not read or written. The clock cannot be set, so a test that needs time
//! to pass moves the lease in the ledger directly (the way `card_reclaim_cli` ends a lease).
#![cfg(unix)]

use crate::common;

use common::home::TempHome;
use moai_cli::ledger;
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Output, Stdio};

const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
const S_A: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
const S_B: &str = "bbbbbbbb-0000-4000-8000-00000000000b";
const S_AUD: &str = "cccccccc-0000-4000-8000-00000000000c";
const DAY: i64 = 86_400;
const FOUR_HOURS: i64 = 14_400;
const REASON: &str = "the lease ended while the card waited for its audit; the holder is gone";
/// Seconds a measured lease may differ from its nominal length: the commands run in between.
const SLACK: i64 = 30;

struct World {
    home: TempHome,
    p: String,
    /// The first commit and the second (the card's change), for the verdict.
    base: String,
    head: String,
    /// The folder the commands run in: not a git worktree of the repository under test.
    host: tempfile::TempDir,
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

/// A refusal: exit 1, `ok` false; the kind and the message.
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
        World {
            p: project.to_str().unwrap().to_string(),
            home,
            base,
            head,
            host: tempfile::Builder::new()
                .prefix("moai-test-host-")
                .tempdir()
                .unwrap(),
        }
    }

    fn project(&self) -> PathBuf {
        PathBuf::from(&self.p)
    }

    /// `moai card <args> --project <project>`, as `session`.
    fn card(&self, session: Option<&str>, args: &[&str]) -> Output {
        let mut all = vec!["card"];
        all.extend_from_slice(args);
        all.extend_from_slice(&["--project", &self.p]);
        let mut cmd: Command = self.home.command(&all);
        cmd.current_dir(self.host.path());
        if let Some(id) = session {
            cmd.env(SESSION_VAR, id);
        }
        cmd.stdin(Stdio::null()).output().expect("run moai")
    }

    /// A ready card with the scope entries given; returns its name.
    fn ready_card(&self, title: &str, scope: &[&str]) -> String {
        let mut args = vec!["add", "--title", title, "--ac", "fixture"];
        for entry in scope {
            args.push("--scope");
            args.push(entry);
        }
        let name = result(&self.card(None, &args))["name"]
            .as_str()
            .unwrap()
            .to_string();
        result(&self.card(None, &["ready", &name]));
        name
    }

    /// A card taken by `S_A` at generation 1 with `lease` seconds (`None`: the default).
    fn taken_card(&self, title: &str, scope: &[&str], lease: Option<&str>) -> String {
        let name = self.ready_card(title, scope);
        let mut args = vec![
            "take",
            &name,
            "--worktree",
            "/work/tree",
            "--branch",
            "card/review-lease-hold",
        ];
        if let Some(lease) = lease {
            args.push("--lease-secs");
            args.push(lease);
        }
        result(&self.card(Some(S_A), &args));
        name
    }

    fn review(&self, card: &str, generation: &str) -> Value {
        result(&self.card(Some(S_A), &["review", card, "--generation", generation]))
    }

    fn reclaim(&self, card: &str, generation: &str) -> Output {
        self.card(
            Some(S_B),
            &[
                "reclaim",
                card,
                "--generation",
                generation,
                "--by",
                "leader",
                "--reason",
                REASON,
            ],
        )
    }

    fn verdict(&self, card: &str, word: &str) -> Output {
        self.card(
            Some(S_AUD),
            &[
                "verdict",
                card,
                "--head",
                &self.head,
                "--base",
                &self.base,
                "--verdict",
                word,
                "--tier",
                "1",
                "--evidence",
                "read the diff and the receipt",
            ],
        )
    }

    /// Runs `statement` against the ledger directly.
    fn sql(&self, statement: &str) {
        let mut l = ledger::open_write(self.home.path(), "demo").unwrap();
        l.with_write_tx(|tx, _| {
            tx.execute_batch(statement)?;
            Ok(())
        })
        .unwrap();
    }

    /// Moves the lease of the card's `held` reservation to `secs` from now (negative: ended).
    fn set_lease(&self, card: i64, secs: i64) {
        self.sql(&format!(
            "UPDATE reservation SET lease_until = unixepoch() + ({secs}) \
             WHERE card_id = {card} AND status = 'held'"
        ));
    }

    /// Seconds left of the lease of the card's `held` reservation.
    fn remaining(&self, card: i64) -> i64 {
        let l = ledger::open_read_only(self.home.path(), "demo").unwrap();
        l.conn()
            .query_row(
                "SELECT lease_until - unixepoch() FROM reservation \
                 WHERE card_id = ?1 AND status = 'held'",
                [card],
                |r| r.get(0),
            )
            .unwrap()
    }

    /// The lease end (seconds since the epoch) of the card's `held` reservation.
    fn lease_until(&self, card: i64) -> i64 {
        self.count(&format!(
            "SELECT lease_until FROM reservation WHERE card_id = {card} AND status = 'held'"
        ))
    }

    fn count(&self, sql: &str) -> i64 {
        let l = ledger::open_read_only(self.home.path(), "demo").unwrap();
        l.conn().query_row(sql, [], |r| r.get(0)).unwrap()
    }

    /// Writes the profile with a `shared_files` list.
    fn shared(&self, list: Value) {
        let profile = json!({"schema_version": 2, "project_id": "demo", "shared_files": list});
        fs::write(
            self.project()
                .join(moai_cli::PROJECT_STATE_DIR)
                .join("project.yaml"),
            format!("{profile}\n"),
        )
        .unwrap();
    }

    /// Three reclaims of card 1, each after its lease ended: the card holds generation 4.
    fn three_reclaims(&self, card: &str) {
        for generation in ["1", "2", "3"] {
            self.set_lease(1, -5);
            let r = result(&self.reclaim(card, generation));
            assert_eq!(r["reservation"]["status"], json!("held"), "{r}");
        }
    }
}

fn within(left: i64, nominal: i64) -> bool {
    left >= nominal - SLACK && left <= nominal + SLACK
}

/// REQ-1 and REQ-2: `card review` leaves a lease that ends a day after the review, and the first
/// `card verdict` on the card leaves one that ends four hours after the verdict.
#[test]
fn ac1_review_holds_the_lease_for_a_day_and_a_verdict_for_four_hours() {
    let w = World::new();
    let card = w.taken_card("short lease", &["apps/bo/"], Some("60"));
    assert!(w.remaining(1) <= 60, "the take did not make a short lease");
    let reviewed = w.review(&card, "1");
    assert_eq!(reviewed["state"], json!("review"), "{reviewed}");
    let left = w.remaining(1);
    assert!(
        left >= DAY - SLACK,
        "the lease after the review ends in {left} seconds, not a day"
    );
    assert_eq!(
        reviewed["reservation"]["lease_until"].as_i64().unwrap(),
        w.lease_until(1),
        "the answer names a different lease end than the ledger holds: {reviewed}"
    );
    // The audit waits: the lease runs low, then the verdict comes.
    w.set_lease(1, 100);
    let verdict = result(&w.verdict(&card, "pass"));
    assert_eq!(verdict["written"], json!(true), "{verdict}");
    let after = w.remaining(1);
    assert!(
        within(after, FOUR_HOURS),
        "the lease after the verdict ends in {after} seconds, not four hours"
    );
}

/// REQ-2: a lease that already ends later than the new end is left as it is, at the review and at
/// the verdict.
#[test]
fn ac1_a_lease_that_already_ends_later_is_left_as_it_is() {
    let w = World::new();
    let card = w.taken_card("long lease", &["apps/bo/"], None);
    w.set_lease(1, 200_000);
    w.review(&card, "1");
    let left = w.remaining(1);
    assert!(
        within(left, 200_000),
        "the review shortened a longer lease to {left} seconds"
    );
    result(&w.verdict(&card, "pass"));
    let after = w.remaining(1);
    assert!(
        within(after, 200_000),
        "the verdict shortened a longer lease to {after} seconds"
    );
}

/// The step after the review can fail: the review stands, the answer says the lease was not held
/// and why, and nothing is swallowed.
#[test]
fn ac1_a_failed_hold_after_review_is_a_warning_and_the_review_stands() {
    use moai_cli::card::work;
    use moai_cli::ledger::reservation::FenceError;
    let w = World::new();
    let card = w.taken_card("hold fails", &["apps/bo/"], Some("60"));
    let failing = |_: &mut ledger::WriteLedger, _: i64, _: &str, _: i64| {
        Err(FenceError::InvalidInput("injected failure".to_string()))
    };
    let answer =
        work::review_keeping_with_hold(&w.project(), w.home.path(), &card, S_A, 1, &[], &failing)
            .expect("the review itself succeeds");
    assert_eq!(answer["state"], json!("review"), "{answer}");
    let warnings = answer["warnings"].as_array().unwrap_or_else(|| {
        panic!("a failed hold is not named in the answer: {answer}");
    });
    let text = warnings
        .iter()
        .filter_map(Value::as_str)
        .collect::<Vec<_>>()
        .join(" | ");
    assert!(text.contains("injected failure"), "{text}");
    assert!(text.contains("lease"), "{text}");
    assert_eq!(
        w.count("SELECT count(*) FROM card WHERE state = 'review'"),
        1
    );
    assert!(
        w.remaining(1) <= 60,
        "the lease was extended although the hold failed"
    );
}

/// REQ-3 and REQ-4: a card in review whose lease is live is refused by `card reclaim` with the
/// audit and the end time named; once the day has passed the reclaim succeeds and counts.
#[test]
fn ac2_a_review_card_with_a_live_lease_is_refused_and_reclaimed_after_the_day() {
    let w = World::new();
    let card = w.taken_card("waits for audit", &["apps/bo/"], Some("60"));
    w.review(&card, "1");
    let until = w.lease_until(1);
    let (kind, message) = refusal(&w.reclaim(&card, "1"));
    assert_eq!(kind, "conflict", "{message}");
    assert!(message.contains("audit"), "{message}");
    assert!(message.contains(&until.to_string()), "{message}");
    assert_eq!(
        w.count("SELECT count(*) FROM event WHERE kind = 'card.reclaim'"),
        0
    );
    // The day passes without a verdict: the lease ends as every other lease does.
    w.set_lease(1, -5);
    let taken = result(&w.reclaim(&card, "1"));
    assert_eq!(taken["reservation"]["generation"], json!(2), "{taken}");
    assert_eq!(taken["counted"], json!(true), "{taken}");
    assert_eq!(taken["reclaims_counted"], json!(1), "{taken}");
}

/// REQ-5: a card reviewed without `--keep` has given its paths back; after its lease ended the
/// reclaim brings the card's own scope back.
#[test]
fn ac3_a_reclaim_after_a_review_without_keep_brings_back_the_cards_own_scope() {
    let w = World::new();
    let card = w.taken_card("scope back", &["apps/bo/", "apps/cli/"], None);
    let reviewed = w.review(&card, "1");
    assert_eq!(
        reviewed["reservation"]["write_scope"],
        json!([]),
        "{reviewed}"
    );
    w.set_lease(1, -5);
    let taken = result(&w.reclaim(&card, "1"));
    assert_eq!(
        taken["reservation"]["write_scope"],
        json!(["apps/bo/", "apps/cli/"]),
        "{taken}"
    );
    assert_eq!(
        taken["scope_filled"],
        json!(["apps/bo/", "apps/cli/"]),
        "{taken}"
    );
    assert_eq!(taken["scope_left_out"], json!([]), "{taken}");
}

/// REQ-6: a path another card's live reservation holds is left out and named; the reclaim
/// still succeeds with the rest.
#[test]
fn ac3_a_path_another_card_holds_is_left_out_and_named() {
    let w = World::new();
    let card = w.taken_card("scope back", &["apps/bo/", "apps/cli/"], None);
    w.review(&card, "1");
    let other = w.ready_card("holds a path", &["apps/cli/"]);
    result(&w.card(Some(S_B), &["take", &other, "--worktree", "/work/other"]));
    w.set_lease(1, -5);
    let taken = result(&w.reclaim(&card, "1"));
    assert_eq!(
        taken["reservation"]["write_scope"],
        json!(["apps/bo/"]),
        "{taken}"
    );
    let left_out = taken["scope_left_out"].as_array().unwrap();
    assert_eq!(left_out.len(), 1, "{taken}");
    assert_eq!(left_out[0]["path"], json!("apps/cli/"), "{taken}");
    assert!(
        left_out[0]["reason"].as_str().unwrap().contains("C-2"),
        "{taken}"
    );
}

/// REQ-6: a path that overlaps a shared file of the profile is left out and named, too.
#[test]
fn ac3_a_shared_file_is_left_out_of_the_fill() {
    let w = World::new();
    let card = w.taken_card("scope back", &["apps/bo/", "src/lib.rs"], None);
    w.review(&card, "1");
    w.shared(json!(["src/lib.rs"]));
    w.set_lease(1, -5);
    let taken = result(&w.reclaim(&card, "1"));
    assert_eq!(
        taken["reservation"]["write_scope"],
        json!(["apps/bo/"]),
        "{taken}"
    );
    let left_out = taken["scope_left_out"].as_array().unwrap();
    assert_eq!(left_out.len(), 1, "{taken}");
    assert_eq!(left_out[0]["path"], json!("src/lib.rs"), "{taken}");
    assert!(
        left_out[0]["reason"].as_str().unwrap().contains("shared"),
        "{taken}"
    );
}

/// REQ-5: `card reopen` on a review card whose reservation is empty fills it the same way, and a
/// reservation that has paths is not changed.
#[test]
fn ac3_reopen_fills_an_empty_reservation_and_leaves_a_full_one() {
    let w = World::new();
    let card = w.taken_card("reopened", &["apps/bo/"], None);
    w.review(&card, "1");
    let reopened = result(&w.card(Some(S_A), &["reopen", &card, "--generation", "1"]));
    assert_eq!(reopened["state"], json!("active"), "{reopened}");
    assert_eq!(
        reopened["reservation"]["write_scope"],
        json!(["apps/bo/"]),
        "{reopened}"
    );
    assert_eq!(reopened["scope_filled"], json!(["apps/bo/"]), "{reopened}");
    // A review that kept its paths: the reopen changes nothing about the scope.
    let kept = w.taken_card("kept", &["apps/kept/"], None);
    result(&w.card(
        Some(S_A),
        &["review", &kept, "--generation", "1", "--keep", "apps/kept/"],
    ));
    let again = result(&w.card(Some(S_A), &["reopen", &kept, "--generation", "1"]));
    assert_eq!(
        again["reservation"]["write_scope"],
        json!(["apps/kept/"]),
        "{again}"
    );
    assert!(again.get("scope_filled").is_none(), "{again}");
}

/// REQ-7 and REQ-8: a card with three counted reclaims and a FAIL verdict is reclaimed once
/// more without counting; a second reclaim after the same verdict is refused.
#[test]
fn ac4_a_reclaim_after_a_fail_verdict_is_free_once() {
    let w = World::new();
    let card = w.taken_card("fails", &["apps/bo/"], None);
    w.review(&card, "1");
    w.three_reclaims(&card);
    assert_eq!(
        w.count("SELECT count(*) FROM event WHERE kind = 'card.reclaim'"),
        3
    );
    // Without a verdict the fourth reclaim is refused, as before.
    w.set_lease(1, -5);
    let (kind, message) = refusal(&w.reclaim(&card, "4"));
    assert_eq!(kind, "conflict", "{message}");
    assert!(message.contains("reclaim limit"), "{message}");
    // The audit fails the card: the next reclaim is outside the allowance.
    w.set_lease(1, 100);
    result(&w.verdict(&card, "fail"));
    w.set_lease(1, -5);
    let free = result(&w.reclaim(&card, "4"));
    assert_eq!(free["counted"], json!(false), "{free}");
    assert_eq!(free["reclaims_counted"], json!(3), "{free}");
    assert_eq!(free["reclaim_allowance"], json!(3), "{free}");
    assert_eq!(free["reservation"]["generation"], json!(5), "{free}");
    assert_eq!(
        w.count(
            "SELECT count(*) FROM event WHERE kind = 'card.reclaim' \
             AND json_extract(payload, '$.counted') = 0"
        ),
        1
    );
    // The same verdict frees one reclaim only.
    w.set_lease(1, -5);
    let (kind, message) = refusal(&w.reclaim(&card, "5"));
    assert_eq!(kind, "conflict", "{message}");
    assert!(message.contains("reclaim limit"), "{message}");
}

/// REQ-7: a PASS verdict frees nothing, and an ordinary reclaim answers that it counted.
#[test]
fn ac4_a_pass_verdict_frees_nothing_and_an_ordinary_reclaim_counts() {
    let w = World::new();
    let card = w.taken_card("passes", &["apps/bo/"], None);
    w.review(&card, "1");
    w.three_reclaims(&card);
    w.set_lease(1, 100);
    result(&w.verdict(&card, "pass"));
    w.set_lease(1, -5);
    let (kind, message) = refusal(&w.reclaim(&card, "4"));
    assert_eq!(kind, "conflict", "{message}");
    assert!(message.contains("reclaim limit"), "{message}");

    let first = World::new();
    let one = first.taken_card("counts", &["apps/bo/"], None);
    first.set_lease(1, -5);
    let taken = result(&first.reclaim(&one, "1"));
    assert_eq!(taken["counted"], json!(true), "{taken}");
    assert_eq!(taken["reclaims_counted"], json!(1), "{taken}");
    assert_eq!(taken["reclaim_allowance"], json!(3), "{taken}");
}
