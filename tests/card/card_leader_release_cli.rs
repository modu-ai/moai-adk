//! `moai card release --by leader` (SPEC-20261125-card-take-records-worktree, revision 2): the
//! confirmed leader session ends the reservation of another session, so the card goes back to
//! `ready` and any lane can take it again. It is no reclaim: no `card.reclaim` event is written
//! and the lifetime limit of three reclaims is not used.
//!
//! Every command runs the built binary against a temporary HOME and a temporary git repository
//! that holds the session registry; the real home state folder is not read or written. The session
//! registry is built only through `moai session-register`, an empty `<name>.live` file and the
//! registry-lease rewrite of `card_reclaim_early_cli`. The clock cannot be set, so a test that
//! needs a lease to end moves it in the ledger directly.
#![cfg(unix)]

use crate::common;

use common::git::{git, git_in};
use common::home::TempHome;
use moai_cli::ledger;
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Child, Command, Output, Stdio};

const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
const REASON: &str = "the holder's terminal is gone and the card waits for a new lane";
const NAME: &str = "agent-1";
/// The session of the registered leader entry and the session that audits.
const LEADER: u64 = 100;
const AUDITOR: u64 = 7;

/// A well-formed session id for number `n`.
fn sid(n: u64) -> String {
    format!("{n:08x}-0000-4000-8000-{n:012x}")
}

/// `C-12` as `12`.
fn id_of(card: &str) -> i64 {
    card.trim_start_matches("C-").parse().unwrap()
}

struct World {
    home: TempHome,
    /// The folder the commands run in: no git worktree.
    host: tempfile::TempDir,
    project: PathBuf,
    /// A worktree with its work pushed: clean.
    wt: PathBuf,
    path_env: std::ffi::OsString,
    /// The two commits of the project, for a verdict.
    base: String,
    head: String,
}

fn world() -> World {
    let home = TempHome::new();
    let project = home.project("demo");
    ledger::register(home.path(), "demo").unwrap();
    // The project is a git repository: it holds the session registry and the commits a verdict
    // names.
    git_in(&project, &["init", "-q", "-b", "main"]);
    fs::write(project.join("a.txt"), "a\n").unwrap();
    git_in(&project, &["add", "a.txt"]);
    git_in(&project, &["commit", "-q", "-m", "init"]);
    let base = git(&project, &["rev-parse", "HEAD"]);
    fs::write(project.join("b.txt"), "b\n").unwrap();
    git_in(&project, &["add", "b.txt"]);
    git_in(&project, &["commit", "-q", "-m", "the card's change"]);
    let head = git(&project, &["rev-parse", "HEAD"]);
    let host = tempfile::Builder::new()
        .prefix("moai-test-host-")
        .tempdir()
        .unwrap();
    let dir = host.path();
    let bare = dir.join("remote.git");
    fs::create_dir_all(&bare).unwrap();
    git_in(&bare, &["init", "-q", "--bare", "-b", "main"]);
    let wt = dir.join("wt");
    fs::create_dir_all(&wt).unwrap();
    git_in(&wt, &["init", "-q", "-b", "main"]);
    fs::write(wt.join("a.txt"), "a\n").unwrap();
    git_in(&wt, &["add", "a.txt"]);
    git_in(&wt, &["commit", "-q", "-m", "first"]);
    git_in(&wt, &["remote", "add", "origin", bare.to_str().unwrap()]);
    git_in(&wt, &["push", "-q", "origin", "main"]);
    // The command runs with a clean environment; the registry lookup spawns `git`, so the
    // directory of this runner's `git` goes back into a minimal PATH.
    let git_dir = std::env::var_os("PATH")
        .and_then(|p| {
            std::env::split_paths(&p)
                .find(|d| d.join("git").is_file())
                .map(|d| d.into_os_string())
        })
        .unwrap_or_default();
    let mut path_env = git_dir;
    path_env.push(":/usr/bin:/bin");
    World {
        home,
        host,
        project,
        wt,
        path_env,
        base,
        head,
    }
}

impl World {
    /// `moai card <args> --project <project>`, as `session` (none: no session id).
    fn card(&self, session: Option<&str>, args: &[&str]) -> Output {
        let mut all = vec!["card"];
        all.extend_from_slice(args);
        let project = self.project.to_str().unwrap();
        all.extend_from_slice(&["--project", project]);
        let mut cmd = self.home.command(&all);
        // A folder that is no git worktree: nothing here records a worktree by itself.
        cmd.current_dir(self.host.path());
        cmd.env("PATH", &self.path_env);
        if let Some(id) = session {
            cmd.env(SESSION_VAR, id);
        }
        cmd.stdin(Stdio::null()).output().expect("run moai")
    }

    fn sessions_dir(&self) -> PathBuf {
        moai_cli::clone_root::path(&self.project.join(".git")).join("sessions")
    }

    /// `moai session-register` for `name` with `extra` options, as the host session `session`.
    fn register_with(&self, name: &str, session: &str, extra: &[&str]) {
        let mut cmd = Command::new(common::MOAI_BIN);
        cmd.arg("session-register")
            .arg(&self.project)
            .args(["--name", name])
            .args(extra)
            .env("HOME", self.home.path())
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .env("GIT_CONFIG_SYSTEM", "/dev/null")
            .env_remove(SESSION_VAR)
            .env_remove("CODEX_THREAD_ID")
            .env(SESSION_VAR, session);
        let out = cmd.output().unwrap();
        assert!(
            out.status.success(),
            "{}",
            String::from_utf8_lossy(&out.stdout)
        );
    }

    fn register(&self, name: &str, session: &str) {
        self.register_with(name, session, &[]);
    }

    /// The registry entry of the leader window: name `leader`, mode `leader`.
    fn register_leader(&self, session: &str) {
        self.register_with("leader", session, &["--mode", "leader"]);
    }

    /// The state of a session whose launcher really ran and has exited: registered under
    /// `name`, a `.live` file nobody locks, and a registry lease that has ended.
    fn register_gone(&self, name: &str, session: &str) {
        self.register(name, session);
        fs::write(self.sessions_dir().join(format!("{name}.live")), b"").unwrap();
        self.set_registry_lease(name, -60);
    }

    /// Moves the registry lease of `name` to `offset` seconds from now (negative: ended).
    fn set_registry_lease(&self, name: &str, offset: i64) {
        // A session-id registration is filed under its session id: find the entry by its name.
        let file = moai_cli::lane_registry::entry_file(&self.sessions_dir(), name)
            .ok()
            .flatten()
            .unwrap_or_else(|| self.sessions_dir().join(format!("{name}.json")));
        let mut entry: Value = serde_json::from_str(&fs::read_to_string(&file).unwrap()).unwrap();
        let now = self.count("SELECT unixepoch()");
        entry["lease_until"] = json!(now + offset);
        fs::write(&file, serde_json::to_vec(&entry).unwrap()).unwrap();
    }

    /// A ready card taken by `holder` in the worktree `worktree` on the branch `main`.
    fn taken(&self, holder: &str, worktree: &Path, scope: &str) -> String {
        let added = result(&self.card(
            None,
            &["add", "--title", "t", "--scope", scope, "--ac", "fixture"],
        ));
        let name = added["name"].as_str().unwrap().to_string();
        result(&self.card(None, &["ready", &name]));
        result(&self.take(holder, &name, worktree));
        name
    }

    fn take(&self, session: &str, card: &str, worktree: &Path) -> Output {
        self.card(
            Some(session),
            &[
                "take",
                card,
                "--worktree",
                worktree.to_str().unwrap(),
                "--branch",
                "main",
            ],
        )
    }

    /// `card review` by the holder at `generation`.
    fn review(&self, holder: &str, card: &str, generation: &str) {
        result(&self.card(Some(holder), &["review", card, "--generation", generation]));
    }

    /// A verdict of the auditor session on the card's two commits.
    fn verdict(&self, card: &str, word: &str) -> Output {
        self.card(
            Some(&sid(AUDITOR)),
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

    fn reclaim(&self, session: &str, card: &str, generation: &str) -> Output {
        self.card(
            Some(session),
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

    /// `card release --by leader` as `session` (none: no session id).
    fn leader_release(&self, session: Option<&str>, card: &str, generation: &str) -> Output {
        self.card(
            session,
            &[
                "release",
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

    /// The leader release by the session of the registered leader entry.
    fn release_as_leader(&self, card: &str, generation: &str) -> Output {
        self.leader_release(Some(&sid(LEADER)), card, generation)
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

    fn count(&self, sql: &str) -> i64 {
        self.rows(sql)[0][0].as_i64().unwrap()
    }

    /// Every row of the tables a release could touch, as text.
    fn dump(&self) -> Vec<String> {
        let mut out = Vec::new();
        for table in [
            "card",
            "card_ac",
            "card_note",
            "reservation",
            "reservation_scope",
            "event",
        ] {
            for r in self.rows(&format!("SELECT * FROM {table} ORDER BY 1, 2")) {
                out.push(format!("{table}:{r}"));
            }
        }
        out
    }

    /// Runs `statement` against the ledger directly: the only way to end a lease (its length is
    /// at least 60 seconds) or to add a note without a pipe.
    fn sql(&self, statement: &str) {
        let mut l = ledger::open_write(self.home.path(), "demo").unwrap();
        l.with_write_tx(|tx, _| {
            tx.execute_batch(statement)?;
            Ok(())
        })
        .unwrap();
    }

    /// Ends the lease of the card's reservation.
    fn end_lease(&self, card: &str) {
        self.sql(&format!(
            "UPDATE reservation SET lease_until = unixepoch() - 5 \
             WHERE card_id = {} AND status IN ('held', 'held_for_review')",
            id_of(card)
        ));
    }

    /// The payloads of the card's `card.release` events, oldest first.
    fn release_events(&self, card: &str) -> Vec<Value> {
        self.rows(&format!(
            "SELECT payload FROM event WHERE card_id = {} AND kind = 'card.release' \
             ORDER BY event_id",
            id_of(card)
        ))
        .iter()
        .map(|r| serde_json::from_str(r[0].as_str().unwrap()).unwrap())
        .collect()
    }

    /// The refusal `out` is a `conflict` that names `code` and says `conflict in`, and it wrote
    /// nothing. Returns the message.
    fn refused_and_unchanged(&self, out: Output, code: &str, before: &[String]) -> String {
        let (kind, message) = refusal(&out);
        assert_eq!(kind, "conflict", "{message}");
        assert!(message.contains(code), "{code}: {message}");
        assert!(message.contains("conflict in"), "conflict in: {message}");
        assert_eq!(&self.dump(), before, "a refusal wrote something");
        message
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

fn sleeper_in(dir: &Path) -> Child {
    Command::new("sleep")
        .arg("30")
        .current_dir(dir)
        .stdin(Stdio::null())
        .spawn()
        .expect("start sleep")
}

/// AC-7 (REQ-8, REQ-12): the confirmed leader releases an `active` card and a `review` card of
/// other sessions whose leases ended. The reservation is `released`, the card `ready`, one
/// `card.release` event names the old holder, generation, lease end, `--by` and `--reason`, no
/// `card.reclaim` event is written, and notes and criteria stay. A verdict after the release of
/// the review card is still recorded and extends no lease.
#[test]
fn ac7_the_leader_releases_an_active_and_a_review_card_whose_lease_ended() {
    let w = world();
    w.register_leader(&sid(LEADER));
    let active = w.taken(&sid(1), &w.wt, "src/active/");
    let review = w.taken(&sid(2), &w.wt, "src/review/");
    w.review(&sid(2), &review, "1");
    for card in [&active, &review] {
        w.sql(&format!(
            "INSERT INTO card_note (card_id, kind, author, body, created_at) \
             VALUES ({}, 'observation', 'agent-9', 'a note that stays', unixepoch())",
            id_of(card)
        ));
        w.end_lease(card);
    }
    let notes_before = w.rows("SELECT * FROM card_note ORDER BY 1");
    let criteria_before = w.rows("SELECT * FROM card_ac ORDER BY 1");
    assert_eq!(notes_before.len(), 2);
    assert_eq!(criteria_before.len(), 2);
    let reclaims = "SELECT count(*) FROM event WHERE kind = 'card.reclaim'";
    let reclaims_before = w.count(reclaims);
    let reservation_sql = |card: &str| {
        format!(
            "SELECT holder, generation, lease_until, worktree, branch FROM reservation \
             WHERE card_id = {}",
            id_of(card)
        )
    };

    for (card, holder, state_before) in [(&active, sid(1), "active"), (&review, sid(2), "review")] {
        let id = id_of(card);
        let state_sql = format!("SELECT state FROM card WHERE card_id = {id}");
        assert_eq!(w.rows(&state_sql), vec![json!([state_before])]);
        let reservation_before = w.rows(&reservation_sql(card));
        let events_before = w.count("SELECT count(*) FROM event");

        let got = result(&w.release_as_leader(card, "1"));
        assert_eq!(got["action"], json!("release"), "{got}");
        assert_eq!(got["reservation"]["status"], json!("released"), "{got}");
        assert_eq!(
            w.rows(&format!(
                "SELECT status FROM reservation WHERE card_id = {id}"
            )),
            vec![json!(["released"])],
            "{card}"
        );
        assert_eq!(w.rows(&state_sql), vec![json!(["ready"])], "{card}");
        // The holder, generation, lease end, worktree and branch stay as the record.
        assert_eq!(w.rows(&reservation_sql(card)), reservation_before, "{card}");
        assert_eq!(
            w.count("SELECT count(*) FROM event"),
            events_before + 1,
            "exactly one new event for {card}"
        );

        let events = w.release_events(card);
        assert_eq!(events.len(), 1, "{card}: {events:?}");
        let payload = &events[0];
        for key in [
            "from_holder",
            "from_generation",
            "from_lease_until",
            "by",
            "reason",
            "card_state",
        ] {
            assert!(payload.get(key).is_some(), "{card}: no `{key}`: {payload}");
        }
        assert_eq!(payload["from_holder"], json!(holder), "{payload}");
        assert_eq!(payload["from_generation"], json!(1), "{payload}");
        assert_eq!(
            payload["from_lease_until"], reservation_before[0][2],
            "{payload}"
        );
        assert_eq!(payload["by"], json!("leader"), "{payload}");
        assert_eq!(payload["reason"], json!(REASON), "{payload}");
        assert!(
            payload.get("early_proof").is_none(),
            "an ended lease needs no proof: {payload}"
        );
    }

    assert_eq!(
        w.count(reclaims),
        reclaims_before,
        "a release is no reclaim: the counted total moved"
    );
    assert_eq!(w.count("SELECT count(*) FROM reservation"), 2);
    assert_eq!(w.rows("SELECT * FROM card_note ORDER BY 1"), notes_before);
    assert_eq!(w.rows("SELECT * FROM card_ac ORDER BY 1"), criteria_before);

    // The verdict still comes for the card that was in review, and it holds no lease.
    let review_reservation = w.rows(&reservation_sql(&review));
    let verdict = result(&w.verdict(&review, "pass"));
    assert_eq!(verdict["written"], json!(true), "{verdict}");
    assert_eq!(
        w.count(&format!(
            "SELECT count(*) FROM event WHERE card_id = {} AND kind = 'card.verdict'",
            id_of(&review)
        )),
        1
    );
    assert_eq!(
        w.count(&format!(
            "SELECT count(*) FROM reservation WHERE card_id = {} \
             AND status IN ('held', 'held_for_review', 'expired_candidate')",
            id_of(&review)
        )),
        0,
        "the verdict made a lease"
    );
    assert_eq!(w.rows(&reservation_sql(&review)), review_reservation);
}

/// A card taken by session 1 in the clean pushed worktree, with the leader registered.
fn live_card(w: &World) -> String {
    w.register_leader(&sid(LEADER));
    w.taken(&sid(1), &w.wt, "src/live/")
}

/// AC-8 (REQ-9, REQ-11): a reservation whose lease is live is released only when every proof of
/// the early reclaim holds, and the event carries `early_proof`; a missing proof refuses with
/// `lease_live`, names the first missing proof and writes nothing.
#[test]
fn ac8_a_live_lease_is_released_only_with_every_early_proof() {
    // Every proof holds: the launcher exited, the worktree is clean and pushed, nobody is in it.
    let w = world();
    let card = live_card(&w);
    w.register_gone(NAME, &sid(1));
    let lease = w.count("SELECT lease_until - unixepoch() FROM reservation WHERE card_id = 1");
    assert!(lease > 3000, "the lease is live: {lease} s left");
    let got = result(&w.release_as_leader(&card, "1"));
    assert_eq!(got["action"], json!("release"), "{got}");
    assert_eq!(got["reservation"]["status"], json!("released"), "{got}");
    assert_eq!(
        w.rows("SELECT state FROM card WHERE card_id = 1"),
        vec![json!(["ready"])]
    );
    let events = w.release_events(&card);
    assert_eq!(events.len(), 1, "{events:?}");
    let proof = &events[0]["early_proof"];
    assert!(!proof.is_null(), "no early_proof in {}", events[0]);
    assert_eq!(events[0]["from_holder"], json!(sid(1)));
    assert_eq!(
        w.count("SELECT count(*) FROM event WHERE kind = 'card.reclaim'"),
        0
    );

    // The liveness of the holder is unknown: no registry entry names its session.
    let w = world();
    let card = live_card(&w);
    let before = w.dump();
    let m = w.refused_and_unchanged(w.release_as_leader(&card, "1"), "lease_live", &before);
    assert!(m.contains("liveness is unknown"), "{m}");

    // The launcher is still running: the entry names the holder and its lock is held.
    let w = world();
    let card = live_card(&w);
    w.register("agent-3", &sid(1));
    let held = moai_cli::session::lock_named_live(&w.sessions_dir(), "agent-3").unwrap();
    let before = w.dump();
    let m = w.refused_and_unchanged(w.release_as_leader(&card, "1"), "lease_live", &before);
    assert!(m.contains("launcher is still running"), "{m}");
    drop(held);

    // No `.live` file: a session registered with `session-register` alone never had a launcher.
    let w = world();
    let card = live_card(&w);
    w.register(NAME, &sid(1));
    w.set_registry_lease(NAME, -60);
    let before = w.dump();
    let m = w.refused_and_unchanged(w.release_as_leader(&card, "1"), "lease_live", &before);
    assert!(m.contains("no launcher lock file"), "{m}");

    // The launcher left a `.live` file but the registry lease of the session has not ended.
    let w = world();
    let card = live_card(&w);
    w.register(NAME, &sid(1));
    fs::write(w.sessions_dir().join(format!("{NAME}.live")), b"").unwrap();
    let before = w.dump();
    let m = w.refused_and_unchanged(w.release_as_leader(&card, "1"), "lease_live", &before);
    assert!(m.contains("has not ended"), "{m}");

    // The worktree holds an uncommitted change.
    let w = world();
    let card = live_card(&w);
    w.register_gone(NAME, &sid(1));
    fs::write(w.wt.join("new.txt"), "x\n").unwrap();
    let before = w.dump();
    let m = w.refused_and_unchanged(w.release_as_leader(&card, "1"), "lease_live", &before);
    assert!(m.contains("worktree dirty"), "{m}");

    // A process has the worktree as its working folder; once it is gone the release goes through.
    let w = world();
    let card = live_card(&w);
    w.register_gone(NAME, &sid(1));
    let before = w.dump();
    let mut inside = sleeper_in(&w.wt);
    let refused = w.release_as_leader(&card, "1");
    let _ = inside.kill();
    let _ = inside.wait();
    let m = w.refused_and_unchanged(refused, "lease_live", &before);
    assert!(m.contains("running process"), "{m}");
    let got = result(&w.release_as_leader(&card, "1"));
    assert_eq!(got["reservation"]["status"], json!("released"), "{got}");
}

/// AC-9 (REQ-10): only the session of the live registered leader entry releases for the leader.
/// No entry, an entry of another session, a call with no session id, an entry whose lease ended
/// and an entry that is not in leader mode each refuse with `not_leader` and write nothing, even
/// though the lease of the card has ended.
#[test]
fn ac9_a_release_by_anyone_but_the_registered_leader_is_refused() {
    let ended_card = |w: &World| {
        let card = w.taken(&sid(1), &w.wt, "src/ended/");
        w.end_lease(&card);
        card
    };

    // No leader entry at all.
    let w = world();
    let card = ended_card(&w);
    let before = w.dump();
    w.refused_and_unchanged(w.release_as_leader(&card, "1"), "not_leader", &before);

    // A leader entry of another session.
    let w = world();
    let card = ended_card(&w);
    w.register_leader(&sid(LEADER));
    let before = w.dump();
    w.refused_and_unchanged(
        w.leader_release(Some(&sid(9)), &card, "1"),
        "not_leader",
        &before,
    );

    // A valid entry, and a call with no session id.
    let w = world();
    let card = ended_card(&w);
    w.register_leader(&sid(LEADER));
    let before = w.dump();
    w.refused_and_unchanged(w.leader_release(None, &card, "1"), "not_leader", &before);

    // An entry of the right session whose lease has ended.
    let w = world();
    let card = ended_card(&w);
    w.register_leader(&sid(LEADER));
    w.set_registry_lease("leader", -60);
    let before = w.dump();
    w.refused_and_unchanged(w.release_as_leader(&card, "1"), "not_leader", &before);

    // An entry named `leader` that is not in leader mode.
    let w = world();
    let card = ended_card(&w);
    w.register("leader", &sid(LEADER));
    let before = w.dump();
    w.refused_and_unchanged(w.release_as_leader(&card, "1"), "not_leader", &before);
}

/// AC-10 (REQ-13): a card in `review` whose newest verdict is `pass` is not released, whether its
/// lease is live (every early proof holds, so only the verdict stops it) or ended. A `fail`
/// verdict does not stop the release.
#[test]
fn ac10_a_review_card_with_a_pass_verdict_is_not_released() {
    let w = world();
    w.register_leader(&sid(LEADER));

    // A live lease whose holder is proven gone.
    let live = w.taken(&sid(1), &w.wt, "src/live/");
    w.review(&sid(1), &live, "1");
    w.register_gone(NAME, &sid(1));
    result(&w.verdict(&live, "pass"));
    let lease = w.count(&format!(
        "SELECT lease_until - unixepoch() FROM reservation WHERE card_id = {}",
        id_of(&live)
    ));
    assert!(lease > 3000, "the lease is live: {lease} s left");
    let before = w.dump();
    w.refused_and_unchanged(w.release_as_leader(&live, "1"), "verdict_pass", &before);

    // An ended lease.
    let ended = w.taken(&sid(2), &w.wt, "src/ended/");
    w.review(&sid(2), &ended, "1");
    result(&w.verdict(&ended, "pass"));
    w.end_lease(&ended);
    let before = w.dump();
    w.refused_and_unchanged(w.release_as_leader(&ended, "1"), "verdict_pass", &before);

    // The other direction: the newest verdict is `fail`, so the ended lease is released.
    let failed = w.taken(&sid(3), &w.wt, "src/failed/");
    w.review(&sid(3), &failed, "1");
    result(&w.verdict(&failed, "fail"));
    w.end_lease(&failed);
    let got = result(&w.release_as_leader(&failed, "1"));
    assert_eq!(got["reservation"]["status"], json!("released"), "{got}");
    assert_eq!(
        w.rows(&format!(
            "SELECT state FROM card WHERE card_id = {}",
            id_of(&failed)
        )),
        vec![json!(["ready"])]
    );
}

/// AC-11 (REQ-14, REQ-16): a card with three counted reclaims refuses a fourth `card reclaim`,
/// and after the leader's release `card take` takes it as it takes any ready card. A holder's own
/// `card release` without `--by` is unchanged: it writes no `card.release` event.
#[test]
fn ac11_a_card_at_the_reclaim_limit_is_released_and_taken() {
    let w = world();
    w.register_leader(&sid(LEADER));
    let card = w.taken(&sid(1), &w.wt, "src/limit/");
    let reclaims = "SELECT count(*) FROM event WHERE kind = 'card.reclaim'";
    for (generation, session) in [("1", 2), ("2", 3), ("3", 4)] {
        w.end_lease(&card);
        let got = result(&w.reclaim(&sid(session), &card, generation));
        assert_eq!(got["reservation"]["status"], json!("held"), "{got}");
    }
    assert_eq!(w.count(reclaims), 3);

    // The fourth reclaim is refused as before.
    w.end_lease(&card);
    let before = w.dump();
    w.refused_and_unchanged(w.reclaim(&sid(5), &card, "4"), "reclaim limit", &before);

    // A holder's own plain release is what it was: no `--by`, no `card.release` event.
    let own = w.taken(&sid(8), &w.wt, "src/own/");
    let got = result(&w.card(Some(&sid(8)), &["release", &own, "--generation", "1"]));
    assert_eq!(got["action"], json!("release"), "{got}");
    assert_eq!(got["state"], json!("ready"), "{got}");
    assert!(w.release_events(&own).is_empty());

    // The leader releases the card at the limit: no reclaim is counted, and a lane takes it.
    let got = result(&w.release_as_leader(&card, "4"));
    assert_eq!(got["reservation"]["status"], json!("released"), "{got}");
    assert_eq!(w.count(reclaims), 3, "the release used a reclaim");
    let taken = result(&w.take(&sid(6), &card, &w.wt));
    assert_eq!(taken["reservation"]["status"], json!("held"), "{taken}");
    assert_eq!(taken["reservation"]["holder"], json!(sid(6)), "{taken}");
    assert_eq!(w.count(reclaims), 3);
}

/// The text of a skill file of this repository.
fn skill_text(name: &str) -> String {
    let path = Path::new(env!("CARGO_MANIFEST_DIR"))
        .join("skills/moai/workflows")
        .join(name);
    fs::read_to_string(&path).unwrap_or_else(|e| panic!("{}: {e}", path.display()))
}

/// AC-12 (REQ-15): the leader workflow names the leader release (a line with `card release` and
/// `--by leader`, because `--by leader` alone is in the `card land` step), its three refusal
/// codes, the reclaim limit and the `card take` that follows; the agent workflow names
/// `--by leader` and the limit exit.
#[test]
fn ac12_the_workflow_texts_name_the_leader_release_and_its_refusals() {
    let leader = skill_text("leader.md");
    let work = skill_text("work.md");
    let mut missing = Vec::new();
    if !leader
        .lines()
        .any(|l| l.contains("card release") && l.contains("--by leader"))
    {
        missing.push("leader.md: a line with `card release` and `--by leader`".to_string());
    }
    for needle in ["not_leader", "lease_live", "verdict_pass", "card take"] {
        if !leader.contains(needle) {
            missing.push(format!("leader.md: `{needle}`"));
        }
    }
    if !(leader.contains("reclaim_limit") || leader.contains("reclaim limit")) {
        missing.push("leader.md: `reclaim_limit` (or `reclaim limit`)".to_string());
    }
    for needle in ["--by leader", "reclaim_limit"] {
        if !work.contains(needle) {
            missing.push(format!("work.md: `{needle}`"));
        }
    }
    assert!(
        missing.is_empty(),
        "the skill text lacks: {}",
        missing.join("; ")
    );
}
