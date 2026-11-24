//! The `stale` key of `moai card next` (SPEC-20261124-card-fold-stale): one row per card that wants
//! closing, of four kinds (`landed_open`, `lease_ended`, `review_no_holder`, `old_draft`), each with
//! the command that closes it. `next` writes nothing.
//!
//! Every command runs against a temporary HOME and a temporary git repository; the real home
//! state is not read or written. The caller is the session id in the environment
//! (`CLAUDE_CODE_SESSION_ID`).
#![cfg(unix)]

mod common;

use common::home::TempHome;
use moai_cli::ledger;
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Output, Stdio};

const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
const S_A: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
const S_B: &str = "bbbbbbbb-0000-4000-8000-00000000000b";
const S_C: &str = "cccccccc-0000-4000-8000-00000000000c";
const S_D: &str = "dddddddd-0000-4000-8000-00000000000d";
const DAY: i64 = 86_400;
const HOUR: i64 = 3_600;

/// The profile of an m-flow project whose integration branch is `integ`, with an optional
/// `stale_draft_days` value spliced in.
fn profile(extra: &str) -> Vec<u8> {
    format!(
        "{{\"schema_version\": 1, \"project_id\": \"demo\", \"purpose\": \"fixture\", \"tools\": [\"rust\"], \"development_methods\": [\"sdd\"], \"hosts\": [\"claude\"], \"work_flow\": {{\"flow\": \"m-flow\", \"integration_branch\": \"integ\"}}{extra}}}"
    )
    .into_bytes()
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

/// A commit with the tree of `parent` and `parent` as its only parent; `branch` is moved to it
/// without a checkout, so the working tree never changes.
fn commit_on(project: &Path, branch: &str, parent: &str, message: &str) -> String {
    let tree = git(project, &["rev-parse", &format!("{parent}^{{tree}}")]);
    let sha = git(
        project,
        &["commit-tree", tree.as_str(), "-p", parent, "-m", message],
    );
    git(
        project,
        &["update-ref", &format!("refs/heads/{branch}"), &sha],
    );
    sha
}

/// An absolute `git` on this runner's `PATH`, for the child that has none.
fn git_executable() -> Option<PathBuf> {
    let path = std::env::var_os("PATH")?;
    std::env::split_paths(&path)
        .map(|dir| dir.join("git"))
        .find(|p| p.is_file())
}

fn result(out: &Output) -> Value {
    let v: Value = serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
        panic!(
            "stdout is not JSON ({e}): {} {}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    });
    assert_eq!(out.status.code(), Some(0), "{v}");
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

/// The rows of `stale` for one card.
fn rows_of<'a>(listing: &'a Value, name: &str) -> Vec<&'a Value> {
    listing["stale"]
        .as_array()
        .unwrap_or_else(|| panic!("`next` answers no `stale` list: {listing}"))
        .iter()
        .filter(|e| e["name"] == json!(name))
        .collect()
}

fn stale_names(listing: &Value) -> Vec<String> {
    listing["stale"]
        .as_array()
        .unwrap_or_else(|| panic!("`next` answers no `stale` list: {listing}"))
        .iter()
        .map(|e| e["name"].as_str().unwrap().to_string())
        .collect()
}

/// The world: a temporary HOME with a registered ledger and a temporary git repository as the
/// project. `main` is at `head`, `integ` is one commit ahead of it, and `scratch` holds one commit
/// outside `integ` (`orphan`).
struct World {
    home: TempHome,
    project: PathBuf,
    p: String,
    git_bin: Option<PathBuf>,
    head: String,
    orphan: String,
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
        fs::write(project.join("b.txt"), "b\n").unwrap();
        git(&project, &["add", "b.txt"]);
        git(&project, &["commit", "-q", "-m", "the card's change"]);
        let head = git(&project, &["rev-parse", "HEAD"]);
        git(&project, &["branch", "integ"]);
        commit_on(&project, "integ", &head, "integration moved on");
        let orphan = commit_on(&project, "scratch", &head, "not integrated");
        World {
            p: project.to_str().unwrap().to_string(),
            home,
            project,
            git_bin: git_executable(),
            head,
            orphan,
        }
    }

    fn command(&self, session: Option<&str>, args: &[&str]) -> Command {
        let mut cmd = self.home.command(args);
        if let Some(git) = &self.git_bin {
            cmd.env("MOAI_GIT_BIN", git);
        }
        if let Some(id) = session {
            cmd.env(SESSION_VAR, id);
        }
        cmd.stdin(Stdio::null());
        cmd
    }

    /// `moai card <args> --project <project>`, as `session`.
    fn card(&self, session: Option<&str>, args: &[&str]) -> Value {
        let mut all = vec!["card"];
        all.extend_from_slice(args);
        all.extend_from_slice(&["--project", &self.p]);
        result(&self.command(session, &all).output().expect("run moai"))
    }

    fn write_profile(&self, extra: &str) {
        let cfg = moai_cli::state_root::path(&self.project, "project.yaml");
        fs::create_dir_all(cfg.parent().unwrap()).unwrap();
        fs::write(&cfg, profile(extra)).unwrap();
    }

    /// `moai card next <project>` as `session`, with the profile as it was written last.
    fn next(&self, session: &str) -> Value {
        result(
            &self
                .command(Some(session), &["card", "next", &self.p])
                .output()
                .expect("run moai"),
        )
    }

    /// Runs `statement` against the ledger directly (the only way to end a lease or to age a
    /// card).
    fn sql(&self, statement: &str) {
        let mut l = ledger::open_write(self.home.path(), "demo").unwrap();
        l.with_write_tx(|tx, _| {
            tx.execute_batch(statement)?;
            Ok(())
        })
        .unwrap();
    }

    fn end_lease(&self, card: &str) {
        let id = card.trim_start_matches("C-");
        self.sql(&format!(
            "UPDATE reservation SET lease_until = unixepoch() - 5 \
             WHERE card_id = {id} AND status = 'held'"
        ));
    }

    /// The lease of the held reservation of `card` ended `secs_ago` seconds ago.
    fn end_lease_ago(&self, card: &str, secs_ago: i64) {
        let id = card.trim_start_matches("C-");
        self.sql(&format!(
            "UPDATE reservation SET lease_until = unixepoch() - {secs_ago} \
             WHERE card_id = {id} AND status = 'held'"
        ));
    }

    /// The holder of `card` last wrote `secs_ago` seconds ago: the reservation and every event of
    /// the card move back, so nothing newer remains. The event table refuses updates in a real
    /// ledger, so the guard is lifted in this throwaway one (`tests/lane_heartbeat_cli.rs` does
    /// the same).
    fn go_silent(&self, card: &str, secs_ago: i64) {
        let id = card.trim_start_matches("C-");
        let file = ledger::ledger_path(self.home.path(), "demo").unwrap();
        rusqlite::Connection::open(&file)
            .unwrap()
            .execute_batch(&format!(
                "DROP TRIGGER IF EXISTS event_no_update; \
                 UPDATE reservation SET updated_at = unixepoch() - {secs_ago} \
                 WHERE card_id = {id}; \
                 UPDATE event SET at = unixepoch() - {secs_ago} WHERE card_id = {id};"
            ))
            .unwrap();
    }

    /// The events of `card` (the holder's own: nobody else wrote to it) are `secs_ago` seconds
    /// old while the reservation keeps the age `go_silent` gave it.
    fn events_at(&self, card: &str, secs_ago: i64) {
        let id = card.trim_start_matches("C-");
        let file = ledger::ledger_path(self.home.path(), "demo").unwrap();
        rusqlite::Connection::open(&file)
            .unwrap()
            .execute_batch(&format!(
                "DROP TRIGGER IF EXISTS event_no_update; \
                 UPDATE event SET at = unixepoch() - {secs_ago} WHERE card_id = {id};"
            ))
            .unwrap();
    }

    /// Makes the card's last update `days` days old. The `rev` moves with it, so the
    /// `card_touch` trigger does not write the clock over it again.
    fn age(&self, card: &str, days: i64) {
        let id = card.trim_start_matches("C-");
        self.sql(&format!(
            "UPDATE card SET updated_at = unixepoch() - {}, rev = rev + 1 WHERE card_id = {id}",
            days * DAY
        ));
    }

    /// A `draft` card; returns its name.
    fn draft(&self, scope: &str) -> String {
        let added = self.card(
            None,
            &[
                "add", "--title", "fixture", "--kind", "task", "--scope", scope, "--ac", "one",
            ],
        );
        added["name"].as_str().unwrap().to_string()
    }

    /// A `ready` card with one write scope and one open criterion; returns its name.
    fn ready(&self, scope: &str) -> String {
        let name = self.draft(scope);
        self.card(None, &["ready", &name]);
        name
    }

    /// `S_A` takes the card, sends it to review and records `head` as landed, then reopens it. It
    /// still holds the card (generation 1) when this returns.
    fn land_and_reopen(&self, card: &str, head: &str) {
        self.card(Some(S_A), &["take", card]);
        self.card(Some(S_A), &["review", card, "--generation", "1"]);
        self.card(
            Some(S_A),
            &["landed", card, "--head", head, "--by", "leader"],
        );
        self.card(Some(S_A), &["reopen", card, "--generation", "1"]);
    }
}

/// AC-3: one card of each kind is listed once, with its kind and a command that names it; a plain
/// `ready` card and a young `draft` are not listed; `next` writes nothing.
#[test]
fn each_kind_of_card_that_wants_closing_is_listed_once_with_its_command() {
    let w = World::new();
    w.write_profile("");
    let landed = w.ready("apps/a/");
    w.land_and_reopen(&landed, &w.head);
    let expired = w.ready("apps/b/");
    w.card(Some(S_A), &["take", &expired]);
    w.end_lease(&expired);
    let review = w.ready("apps/c/");
    w.card(Some(S_A), &["take", &review]);
    w.card(Some(S_A), &["review", &review, "--generation", "1"]);
    w.sql(&format!(
        "UPDATE reservation SET status = 'released' WHERE card_id = {}",
        review.trim_start_matches("C-")
    ));
    let old = w.draft("apps/d/");
    w.age(&old, 15);
    let young = w.draft("apps/e/");
    let plain = w.ready("apps/f/");

    let before = w.home.snapshot();
    let listing = w.next(S_B);
    // `assert!` and not `assert_eq!`: a failure must not print the whole ledger file.
    assert!(w.home.snapshot() == before, "next writes nothing");

    let mut want = vec![landed.clone(), expired.clone(), review.clone(), old.clone()];
    want.sort_by_key(|n| n.trim_start_matches("C-").parse::<i64>().unwrap());
    assert_eq!(stale_names(&listing), want, "{listing}");
    for (card, kind) in [
        (&landed, "landed_open"),
        (&expired, "lease_ended"),
        (&review, "review_no_holder"),
        (&old, "old_draft"),
    ] {
        let rows = rows_of(&listing, card);
        assert_eq!(rows.len(), 1, "{card} is listed once: {listing}");
        assert_eq!(rows[0]["kind"], json!(kind), "{listing}");
        let command = rows[0]["command"].as_str().unwrap_or("");
        assert!(
            command.starts_with("moai card ") && command.contains(card.as_str()),
            "the command names the card: {command}"
        );
    }
    for card in [&young, &plain] {
        assert!(rows_of(&listing, card).is_empty(), "{card}: {listing}");
    }
}

/// AC-4: a landed card is closed by `finish` with its head; a head outside the integration branch
/// proves nothing; a card that is both landed and lease-ended is listed once, as `landed_open`.
#[test]
fn a_landed_open_card_names_finish_with_its_head_and_wins_over_an_ended_lease() {
    let w = World::new();
    w.write_profile("");
    let landed = w.ready("apps/a/");
    w.land_and_reopen(&landed, &w.head);
    w.end_lease(&landed);
    let outside = w.ready("apps/b/");
    w.land_and_reopen(&outside, &w.orphan);

    let listing = w.next(S_B);
    let rows = rows_of(&listing, &landed);
    assert_eq!(rows.len(), 1, "{listing}");
    assert_eq!(rows[0]["kind"], json!("landed_open"), "{listing}");
    assert_eq!(rows[0]["head"], json!(w.head), "{listing}");
    assert_eq!(
        rows[0]["command"],
        json!(format!(
            "moai card finish {landed} --head {} --by leader",
            w.head
        )),
        "{listing}"
    );
    assert!(
        rows_of(&listing, &outside).is_empty(),
        "a head outside the integration branch is no landing: {listing}"
    );
}

/// REQ-5: with no profile git is not read; the landed card is listed by its ended lease, the draft
/// by its age, the answer still exits 0 and `landing.note` says why.
#[test]
fn without_a_readable_integration_branch_only_the_proofs_are_missing() {
    let w = World::new();
    let landed = w.ready("apps/a/");
    w.land_and_reopen(&landed, &w.head);
    w.end_lease(&landed);
    let old = w.draft("apps/d/");
    w.age(&old, 15);

    let listing = w.next(S_B);
    assert_eq!(listing["landing"]["checked"], json!(false), "{listing}");
    assert!(
        !listing["landing"]["note"].as_str().unwrap_or("").is_empty(),
        "{listing}"
    );
    let rows = rows_of(&listing, &landed);
    assert_eq!(rows.len(), 1, "{listing}");
    assert_eq!(rows[0]["kind"], json!("lease_ended"), "{listing}");
    assert_eq!(rows_of(&listing, &old)[0]["kind"], json!("old_draft"));
}

/// AC-5: the profile key moves the threshold, and with no key it is 14 days.
#[test]
fn the_stale_draft_days_come_from_the_profile_and_default_to_fourteen() {
    let w = World::new();
    let ten = w.draft("apps/a/");
    w.age(&ten, 10);
    let fifteen = w.draft("apps/b/");
    w.age(&fifteen, 15);

    w.write_profile("");
    let listing = w.next(S_B);
    assert!(rows_of(&listing, &ten).is_empty(), "{listing}");
    assert_eq!(rows_of(&listing, &fifteen).len(), 1, "{listing}");

    w.write_profile(", \"stale_draft_days\": 3");
    let listing = w.next(S_B);
    assert_eq!(rows_of(&listing, &ten).len(), 1, "{listing}");
    assert_eq!(rows_of(&listing, &fifteen).len(), 1, "{listing}");

    w.write_profile(", \"stale_draft_days\": 12");
    let listing = w.next(S_B);
    assert!(rows_of(&listing, &ten).is_empty(), "{listing}");
    assert_eq!(rows_of(&listing, &fifteen).len(), 1, "{listing}");
}

/// AC-5: a profile that sets the key to 0 is refused when it is read.
#[test]
fn a_profile_with_zero_stale_draft_days_is_refused() {
    let w = World::new();
    w.write_profile(", \"stale_draft_days\": 0");
    assert!(
        moai_cli::profile::read_profile(&w.project).is_err(),
        "stale_draft_days: 0 must not read"
    );
    w.write_profile(", \"stale_draft_days\": 1");
    assert!(
        moai_cli::profile::read_profile(&w.project).is_ok(),
        "stale_draft_days: 1 reads"
    );
}

/// A profile that cannot be read gives the 14-day default and adds no warning: `card_spec_cli`
/// and other fixtures count `warnings` exactly, and their profiles are often not valid.
#[test]
fn an_unreadable_profile_uses_the_default_days_and_adds_no_warning() {
    let w = World::new();
    // The profile `TempHome::project` leaves names the project and is not a complete profile.
    assert!(moai_cli::profile::read_profile(&w.project).is_err());
    let old = w.draft("apps/a/");
    w.age(&old, 15);

    let listing = w.next(S_B);
    let rows = rows_of(&listing, &old);
    assert_eq!(rows.len(), 1, "{listing}");
    assert_eq!(rows[0]["threshold_days"], json!(14), "{listing}");
    // SPEC-20261124-card-help-gates AC-8 (REQ-7): no warning of any wording is added by the stale
    // draft: the whole `warnings` array equals that of the same fixture without a stale draft.
    let fresh = World::new();
    assert!(moai_cli::profile::read_profile(&fresh.project).is_err());
    let young = fresh.draft("apps/a/");
    fresh.age(&young, 1);
    let fresh_listing = fresh.next(S_B);
    assert!(
        rows_of(&fresh_listing, &young).is_empty(),
        "{fresh_listing}"
    );
    assert_eq!(
        listing["warnings"], fresh_listing["warnings"],
        "the stale draft changed the warnings: {listing} against {fresh_listing}"
    );
    // The fixture holds no card that wants a SPEC, so the one array both listings carry is empty:
    // a warning of the unreadable profile, in any wording, is not in it.
    assert_eq!(listing["warnings"], json!([]), "{listing}");
}

fn unix_now() -> i64 {
    std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .unwrap()
        .as_secs() as i64
}

/// SPEC-20261119-live-lease-takeover REQ-10: a `lease_ended` row carries the holder's heartbeat
/// and the seconds since, the value the sweep reads.
#[test]
fn stale_heartbeat_row_shows_last_seen_at_and_silent_secs() {
    let w = World::new();
    w.write_profile("");
    let silent = w.ready("apps/a/");
    w.card(Some(S_A), &["take", &silent]);
    w.end_lease(&silent);
    w.go_silent(&silent, 3 * HOUR);
    let recent = w.ready("apps/b/");
    w.card(Some(S_B), &["take", &recent]);
    w.end_lease(&recent);
    w.go_silent(&recent, 120);

    let before = w.home.snapshot();
    let listing = w.next(S_D);
    assert!(w.home.snapshot() == before, "next writes nothing");
    let now = unix_now();
    for (card, ago) in [(&silent, 3 * HOUR), (&recent, 120)] {
        let rows = rows_of(&listing, card);
        assert_eq!(rows.len(), 1, "{card}: {listing}");
        assert_eq!(rows[0]["kind"], json!("lease_ended"), "{listing}");
        let seen = rows[0]["last_seen_at"]
            .as_i64()
            .unwrap_or_else(|| panic!("{card} has no last_seen_at: {listing}"));
        let secs = rows[0]["silent_secs"]
            .as_i64()
            .unwrap_or_else(|| panic!("{card} has no silent_secs: {listing}"));
        assert!(
            (ago..ago + 60).contains(&secs),
            "{card}: silent_secs {secs} for a holder silent {ago}s: {listing}"
        );
        assert!(
            (now - seen - secs).abs() <= 30,
            "{card}: last_seen_at {seen} and silent_secs {secs} name one moment: {listing}"
        );
    }
}

/// REQ-10, REQ-1: the heartbeat of a row is the later of the reservation's update and the holder's
/// latest event, as the sweep reads it: a holder that wrote a minute ago keeps its card whatever
/// the age of the reservation row.
#[test]
fn stale_heartbeat_counts_the_holders_latest_event_and_not_only_the_reservation() {
    let w = World::new();
    w.write_profile("");
    let card = w.ready("apps/a/");
    w.card(Some(S_A), &["take", &card]);
    w.end_lease_ago(&card, 60);
    w.go_silent(&card, 3 * HOUR);
    w.events_at(&card, 60);

    let listing = w.next(S_D);
    let row = rows_of(&listing, &card)[0];
    let secs = row["silent_secs"].as_i64().unwrap();
    assert!(
        (60..120).contains(&secs),
        "the holder's event a minute ago is the heartbeat, not the reservation's three hours: {listing}"
    );
    assert!(
        row["command"]
            .as_str()
            .unwrap()
            .starts_with("moai card withdraw "),
        "a holder that wrote a minute ago is not silent: {listing}"
    );
    assert!(row.get("via").is_none(), "{listing}");
}

/// REQ-11: the row names the sweep for a card the sweep would take, and keeps the withdraw for a
/// card it would not; the command it names does sweep the card.
#[test]
fn stale_heartbeat_names_the_sweep_for_a_silent_card_and_keeps_withdraw_for_a_recent_one() {
    let w = World::new();
    w.write_profile("");
    let silent = w.ready("apps/a/");
    w.card(Some(S_A), &["take", &silent]);
    w.end_lease_ago(&silent, 60);
    w.go_silent(&silent, 3 * HOUR);
    let old = w.ready("apps/b/");
    w.card(Some(S_B), &["take", &old]);
    w.end_lease_ago(&old, 25 * HOUR);
    w.go_silent(&old, 25 * HOUR);
    let recent = w.ready("apps/c/");
    w.card(Some(S_C), &["take", &recent]);
    w.end_lease_ago(&recent, 60);
    w.go_silent(&recent, 120);

    let listing = w.next(S_D);
    let row = |card: &String| rows_of(&listing, card)[0].clone();
    let sweep = |card: &String| format!("moai card sweep {card} --by leader");
    assert_eq!(row(&silent)["command"], json!(sweep(&silent)), "{listing}");
    assert_eq!(row(&silent)["via"], json!("heartbeat"), "{listing}");
    assert_eq!(row(&old)["command"], json!(sweep(&old)), "{listing}");
    assert_eq!(row(&old)["via"], json!("grace"), "{listing}");
    let kept = row(&recent);
    assert!(
        kept["command"]
            .as_str()
            .unwrap()
            .starts_with("moai card withdraw "),
        "a card with a recent heartbeat keeps the withdraw: {listing}"
    );
    assert!(kept.get("via").is_none(), "{listing}");

    // The command the row names takes the card: the row and the sweep agree.
    let swept = w.card(None, &["sweep", &silent, "--by", "leader"]);
    assert_eq!(swept["swept"][0]["name"], json!(silent), "{swept}");
    assert_eq!(swept["swept"][0]["via"], json!("heartbeat"), "{swept}");
}
