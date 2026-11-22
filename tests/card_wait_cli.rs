//! `moai card wait` (SPEC-20261122-card-wait): the next note on the named cards, or a note on any
//! card that is not done whose first line says `all lanes`, printed once; or a timeout.
//!
//! Every command runs against a temporary HOME; the real HOME is not read or written. Notes that
//! a case needs with an exact author, origin or card state are inserted with SQL; the note that
//! ends a running wait is written by another `moai card note` process, as a lane would write it.
#![cfg(unix)]

mod common;

use common::home::TempHome;
use moai_cli::ledger;
use serde_json::{json, Value};
use std::io::Write;
use std::process::{Child, Output, Stdio};
use std::time::{Duration, Instant};

/// The longest any one wait of these tests may run before the test fails and kills it.
const HARD_LIMIT: Duration = Duration::from_secs(60);

struct World {
    home: TempHome,
    project: String,
}

/// A HOME with a project and a registered ledger.
fn world() -> World {
    let home = TempHome::new();
    let project = home.project("demo").to_str().unwrap().to_string();
    ledger::register(home.path(), "demo").unwrap();
    World { home, project }
}

impl World {
    /// `moai card wait [args] --project <project>`.
    fn wait_args(&self, extra: &[&str]) -> Vec<String> {
        let mut args: Vec<String> = ["card", "wait"].iter().map(|s| s.to_string()).collect();
        args.extend(extra.iter().map(|s| s.to_string()));
        args.push("--project".to_string());
        args.push(self.project.clone());
        args
    }

    /// Runs `card wait` to its end, failing the test (and killing the child) past [`HARD_LIMIT`].
    fn wait(&self, extra: &[&str]) -> (Output, Duration) {
        let started = Instant::now();
        let child = self.spawn_wait(extra);
        let out = finish(child, started + HARD_LIMIT);
        (out, started.elapsed())
    }

    fn spawn_wait(&self, extra: &[&str]) -> Child {
        let args = self.wait_args(extra);
        let argv: Vec<&str> = args.iter().map(String::as_str).collect();
        self.home
            .command(&argv)
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .spawn()
            .expect("start moai card wait")
    }

    /// `moai card add --title T` for each title; the cards are `C-1`, `C-2`, … in order.
    fn cards(&self, titles: &[&str]) {
        for title in titles {
            let out = self
                .home
                .command(&["card", "add", "--title", title, "--project", &self.project])
                .output()
                .unwrap();
            assert_eq!(out.status.code(), Some(0), "{}", text(&out));
        }
    }

    /// A card that is `done`, inserted whole (an insert has no state check).
    fn done_card(&self, id: i64, title: &str) {
        let w = ledger::open_write(self.home.path(), "demo").unwrap();
        w.conn()
            .execute(
                "INSERT INTO card (card_id, title, state, created_by, created_at, updated_at, closed_at) \
                 VALUES (?1, ?2, 'done', 'tester', 1, 2, 2)",
                rusqlite::params![id, title],
            )
            .unwrap();
    }

    /// One note inserted with SQL; returns its note id.
    fn note(&self, card: i64, author: &str, origin: &str, body: &str) -> i64 {
        let w = ledger::open_write(self.home.path(), "demo").unwrap();
        w.conn()
            .execute(
                "INSERT INTO card_note (card_id, kind, author, origin, body, created_at) \
                 VALUES (?1, 'comment', ?2, ?3, ?4, 100)",
                rusqlite::params![card, author, origin, body],
            )
            .unwrap();
        w.conn().last_insert_rowid()
    }

    /// One note written by another `moai card note` process; returns its note id.
    fn card_note(&self, card: &str, by: &str, body: &str) -> i64 {
        let mut child = self
            .home
            .command(&[
                "card",
                "note",
                card,
                "--kind",
                "comment",
                "--by",
                by,
                "--project",
                &self.project,
            ])
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .spawn()
            .unwrap();
        child
            .stdin
            .take()
            .unwrap()
            .write_all(body.as_bytes())
            .unwrap();
        let out = child.wait_with_output().unwrap();
        result(&out)["note_id"].as_i64().unwrap()
    }

    fn max_note_id(&self) -> i64 {
        let r = ledger::open_read_only(self.home.path(), "demo").unwrap();
        r.conn()
            .query_row("SELECT COALESCE(MAX(note_id), 0) FROM card_note", [], |r| {
                r.get(0)
            })
            .unwrap()
    }
}

/// Waits for `child` to end by `deadline`; past it the child is killed and the test fails.
fn finish(mut child: Child, deadline: Instant) -> Output {
    loop {
        if child.try_wait().unwrap().is_some() {
            return child.wait_with_output().unwrap();
        }
        if Instant::now() >= deadline {
            let _ = child.kill();
            let out = child.wait_with_output().unwrap();
            panic!("card wait ran past the test's hard limit: {}", text(&out));
        }
        std::thread::sleep(Duration::from_millis(50));
    }
}

fn text(out: &Output) -> String {
    format!(
        "{} {}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    )
}

fn envelope(out: &Output) -> Value {
    serde_json::from_slice(&out.stdout).unwrap_or_else(|e| panic!("{e}: {}", text(out)))
}

fn result(out: &Output) -> Value {
    assert_eq!(out.status.code(), Some(0), "{}", text(out));
    let v = envelope(out);
    assert_eq!(v["ok"], json!(true), "{v}");
    assert_eq!(v["operation"], json!("card"), "{v}");
    v["result"].clone()
}

fn refusal(out: &Output) -> (String, String) {
    assert_eq!(out.status.code(), Some(1), "{}", text(out));
    let v = envelope(out);
    assert_eq!(v["ok"], json!(false), "{v}");
    (
        v["error"]["kind"].as_str().unwrap().to_string(),
        v["error"]["message"].as_str().unwrap().to_string(),
    )
}

/// `(card, note_id)` of every item, in the order printed.
fn items(r: &Value) -> Vec<(String, i64)> {
    r["items"]
        .as_array()
        .unwrap()
        .iter()
        .map(|i| {
            (
                i["card"].as_str().unwrap().to_string(),
                i["note_id"].as_i64().unwrap(),
            )
        })
        .collect()
}

fn item(card: &str, note_id: i64) -> (String, i64) {
    (card.to_string(), note_id)
}

#[test]
fn a_note_written_during_the_wait_ends_it_and_older_notes_are_not_returned() {
    let w = world();
    w.cards(&["watched", "other"]);
    let older = w.card_note("C-1", "leader", "LEADER answer written before the wait");
    let child = w.spawn_wait(&["--card", "C-1", "--timeout", "50s", "--every", "1s"]);
    let started = Instant::now();
    let deadline = started + HARD_LIMIT;
    // The wait fixes its mark when it starts, and that moment cannot be seen from outside. So a
    // note is written, the wait gets three polls to end on it, and only then is the next one
    // written: a note written before the mark is not returned, and the wait ends on the first
    // note after it.
    let mut written = Vec::new();
    let mut child = Some(child);
    let out = loop {
        let id = w.card_note(
            "C-1",
            "leader",
            &format!(
                "LEADER answer {} during the wait\nsecond line",
                written.len()
            ),
        );
        written.push(id);
        let poll_until = Instant::now() + Duration::from_secs(3);
        let running = child.as_mut().unwrap();
        while Instant::now() < poll_until && running.try_wait().unwrap().is_none() {
            std::thread::sleep(Duration::from_millis(50));
        }
        if running.try_wait().unwrap().is_some() {
            break child.take().unwrap().wait_with_output().unwrap();
        }
        if Instant::now() >= deadline {
            break finish(child.take().unwrap(), deadline);
        }
    };
    let r = result(&out);
    assert_eq!(r["action"], json!("wait"), "{r}");
    assert_eq!(r["outcome"], json!("note"), "{r}");
    assert_eq!(r["writes"], json!(false), "{r}");
    let last = *written.last().unwrap();
    assert_eq!(items(&r), vec![item("C-1", last)], "{r}");
    assert_eq!(r["count"], json!(1), "{r}");
    let got = &r["items"][0];
    assert_eq!(got["card_id"], json!(1), "{got}");
    assert_eq!(got["author"], json!("leader"), "{got}");
    assert_eq!(got["kind"], json!("comment"), "{got}");
    assert_eq!(
        got["line"],
        json!(format!(
            "LEADER answer {} during the wait",
            written.len() - 1
        )),
        "only the first line leaves the ledger"
    );
    assert!(got["created_at"].as_i64().unwrap() > 1_700_000_000, "{got}");
    // The mark the wait fixed lies at or above the note written before it started and below
    // the note it returned.
    let mark = r["mark"].as_i64().unwrap();
    assert!(
        mark >= older && mark < last,
        "mark {mark}, older {older}, last {last}"
    );
    assert!(r["waited_secs"].as_u64().unwrap() < 50, "{r}");
}

#[test]
fn without_a_matching_note_the_wait_ends_at_its_timeout() {
    let w = world();
    w.cards(&["watched", "other"]);
    w.card_note("C-1", "leader", "LEADER note before the wait");
    let started = Instant::now();
    let child = w.spawn_wait(&["--card", "C-1", "--timeout", "2s", "--every", "1s"]);
    // A note on a card nobody named, without `all lanes`: it does not end the wait.
    w.card_note("C-2", "leader", "LEADER note on another card");
    let out = finish(child, started + HARD_LIMIT);
    let elapsed = started.elapsed();
    let r = result(&out);
    assert_eq!(r["action"], json!("wait"), "{r}");
    assert_eq!(r["outcome"], json!("timeout"), "{r}");
    assert_eq!(r["items"], json!([]), "{r}");
    assert_eq!(r["count"], json!(0), "{r}");
    assert_eq!(r["writes"], json!(false), "{r}");
    assert!(r["waited_secs"].as_u64().unwrap() >= 2, "{r}");
    assert!(elapsed >= Duration::from_secs(2), "{elapsed:?}");
    // The timeout plus one poll, with room for a slow machine; never the test's hard limit.
    assert!(elapsed < Duration::from_secs(12), "{elapsed:?}");

    // An interval longer than the timeout is cut to the time left: the wait ends at the timeout,
    // not after a whole interval (30 s here; an uncapped sleep would run past the limit above).
    let (out, elapsed) = w.wait(&["--card", "C-1", "--timeout", "2s", "--every", "30"]);
    let r = result(&out);
    assert_eq!(r["outcome"], json!("timeout"), "{r}");
    assert!(r["waited_secs"].as_u64().unwrap() >= 2, "{r}");
    assert!(elapsed >= Duration::from_secs(2), "{elapsed:?}");
    assert!(elapsed < Duration::from_secs(12), "{elapsed:?}");
}

/// A second connection to the ledger file that holds it exclusively until `release`. This process
/// must not open the ledger file by any other path while it is held (closing any descriptor of a
/// file drops the process's POSIX locks on it), so a case lets other processes do the reading and
/// writing.
struct Lock(rusqlite::Connection);

impl Lock {
    fn take(w: &World) -> Lock {
        let path = ledger::ledger_path(w.home.path(), "demo").unwrap();
        let held = rusqlite::Connection::open(path).unwrap();
        held.execute_batch("BEGIN EXCLUSIVE").unwrap();
        Lock(held)
    }

    fn release(self) {
        self.0.execute_batch("ROLLBACK").unwrap();
    }
}

#[test]
fn a_busy_ledger_counts_as_a_poll_that_found_nothing() {
    let w = world();
    w.cards(&["watched"]);
    let started = Instant::now();
    let mut child = w.spawn_wait(&["--card", "C-1", "--timeout", "40s", "--every", "1s"]);
    // The wait reads the ledger once at its start, which cannot be seen from outside: it gets a
    // head start, and the lock is taken only after that. (A start that did meet the lock would
    // wait for it and still pass; only the check of the busy poll below would be skipped.)
    std::thread::sleep(Duration::from_millis(1500));
    let lock = Lock::take(&w);
    // Several polls meet the lock, each giving up after a quarter of a second.
    std::thread::sleep(Duration::from_secs(3));
    if child.try_wait().unwrap().is_some() {
        let out = child.wait_with_output().unwrap();
        panic!(
            "the wait ended while another process held the ledger: {}",
            text(&out)
        );
    }
    lock.release();
    // The wait kept polling: a note written once the ledger is free ends it.
    let id = w.card_note("C-1", "leader", "LEADER answer after the lock");
    let out = finish(child, started + HARD_LIMIT);
    let r = result(&out);
    assert_eq!(r["outcome"], json!("note"), "{r}");
    assert_eq!(items(&r), vec![item("C-1", id)], "{r}");
    assert!(r["waited_secs"].as_u64().unwrap() < 40, "{r}");
}

#[test]
fn a_ledger_held_at_the_start_counts_in_the_wait() {
    let w = world();
    w.cards(&["watched"]);
    let lock = Lock::take(&w);
    let started = Instant::now();
    let child = w.spawn_wait(&["--card", "C-1", "--timeout", "1s", "--every", "1s"]);
    // The start read waits for the lock (the ledger's own limit is 5 s, so the 2.5 s hold leaves
    // it room under load), and the wait ends as a timeout once it is free. The order of the clock
    // and the start read is pinned without a clock by the unit test
    // `a_start_read_that_used_up_the_timeout_leaves_one_poll` in src/card/wait.rs; this test only
    // checks that the wired command reports the time spent at the start. Every bound here is a
    // lower one, which load can only make easier to meet.
    std::thread::sleep(Duration::from_millis(2500));
    lock.release();
    let out = finish(child, started + HARD_LIMIT);
    let r = result(&out);
    assert_eq!(r["outcome"], json!("timeout"), "{r}");
    assert!(r["waited_secs"].as_u64().unwrap() >= 2, "{r}");
}

#[test]
fn a_wait_without_a_valid_timeout_or_card_is_a_usage_error() {
    // A HOME without a ledger: a request that got past the command line would be refused with
    // exit 1 (`not_found`), so exit 2 shows the refusal came before the ledger was read.
    let home = TempHome::new();
    let project = home.project("demo").to_str().unwrap().to_string();
    let before = home.snapshot();
    let run = |extra: &[&str]| {
        let mut args = vec!["card", "wait"];
        args.extend_from_slice(extra);
        args.extend_from_slice(&["--project", project.as_str()]);
        home.command(&args).output().unwrap()
    };
    for bad in [
        vec!["--card", "C-1"],
        vec!["--card", "C-1", "--timeout", "0"],
        vec!["--card", "C-1", "--timeout", "0s"],
        vec!["--card", "C-1", "--timeout", "13h"],
        vec!["--card", "C-1", "--timeout", "43201"],
        vec!["--card", "C-1", "--timeout", "721m"],
        vec!["--card", "C-1", "--timeout", "5x"],
        vec!["--card", "C-1", "--timeout", ""],
        vec!["--card", "C-1", "--timeout", "-1s"],
        vec!["--card", "C-1", "--timeout", "1.5s"],
        vec!["--card", "C-1", "--timeout", "10s", "--every", "0"],
        vec!["--card", "C-1", "--timeout", "10s", "--every", "301"],
        vec!["--card", "C-1", "--timeout", "10s", "--every", "6m"],
        vec!["--card", "C-1", "--timeout", "10s", "--since", "x"],
        vec!["--card", "C-1", "--timeout", "10s", "--since", "-1"],
        vec!["--timeout", "10s"],
    ] {
        let out = run(&bad);
        assert_eq!(out.status.code(), Some(2), "{bad:?}: {}", text(&out));
        assert_eq!(envelope(&out)["error"]["kind"], json!("usage"), "{bad:?}");
    }
    // The last value that passes each limit gets past the command line: the absent ledger
    // refuses it with exit 1.
    for edge in [
        vec!["--card", "C-1", "--timeout", "1"],
        vec!["--card", "C-1", "--timeout", "12h"],
        vec!["--card", "C-1", "--timeout", "43200"],
        vec!["--card", "C-1", "--timeout", "720m"],
        vec!["--card", "C-1", "--timeout", "10s", "--every", "1"],
        vec!["--card", "C-1", "--timeout", "10s", "--every", "300"],
        vec!["--card", "C-1", "--timeout", "10s", "--every", "5m"],
        vec!["--card", "C-1", "--timeout", "10s", "--since", "0"],
    ] {
        let (kind, message) = refusal(&run(&edge));
        assert_eq!(kind, "not_found", "{edge:?}: {message}");
    }
    assert_eq!(home.snapshot(), before, "nothing was created");

    // The first sentence of `--help` says it only reads.
    let help = home.command(&["card", "wait", "--help"]).output().unwrap();
    assert!(help.status.success());
    let head = String::from_utf8_lossy(&help.stdout).into_owned();
    let first = head.lines().find(|l| !l.trim().is_empty()).unwrap_or("");
    assert!(first.starts_with("Reads the ledger only."), "{first}");
}

#[test]
fn own_notes_and_other_cards_are_skipped_and_an_all_lanes_line_is_returned() {
    let w = world();
    w.cards(&["mine", "other", "third"]);
    w.done_card(4, "closed");
    let mark = w.max_note_id();
    let own = w.note(1, "agent-1", "local", "LANE-NOTE agent-1 C-1 progress"); // skipped: own
    let answer = w.note(1, "leader", "local", "LEADER answer to agent-1"); // named card
    let imported = w.note(1, "GoosLab", "import", "an imported comment"); // not local
    let plain = w.note(2, "leader", "local", "LEADER note to agent-2"); // card not named
    let announce = w.note(2, "leader", "local", "ALL LANES: stop pushing\nbody"); // any case
    let later = w.note(3, "leader", "local", "a note\nall lanes on the second line"); // 2nd line
    let closed = w.note(4, "leader", "local", "all lanes on a done card"); // done card
    let own_announce = w.note(3, "agent-1", "local", "All Lanes from agent-1 itself"); // own
    let lower = w.note(3, "agent-3", "local", "notice to all lanes: rebase"); // lower case
    let since = mark.to_string();
    let (out, _) = w.wait(&[
        "--card",
        "C-1",
        "--not-author",
        "agent-1",
        "--not-author",
        "agent-9",
        "--since",
        &since,
        "--timeout",
        "30s",
        "--every",
        "300",
    ]);
    let r = result(&out);
    assert_eq!(r["outcome"], json!("note"), "{r}");
    assert_eq!(
        items(&r),
        vec![
            item("C-1", answer),
            item("C-2", announce),
            item("C-3", lower)
        ],
        "skipped: own {own}, imported {imported}, plain {plain}, second line {later}, \
         done card {closed}, own announcement {own_announce}"
    );
    assert_eq!(r["count"], json!(3));
    assert_eq!(r["items"][1]["line"], json!("ALL LANES: stop pushing"));

    // Without `--not-author` the own notes come back too; with the cards given comma-separated.
    let (out, _) = w.wait(&[
        "--card",
        "C-1,C-3",
        "--since",
        &since,
        "--timeout",
        "30s",
        "--every",
        "300",
    ]);
    let r = result(&out);
    assert_eq!(
        items(&r),
        vec![
            item("C-1", own),
            item("C-1", answer),
            item("C-2", announce),
            item("C-3", later),
            item("C-3", own_announce),
            item("C-3", lower),
        ],
        "{r}"
    );
}

#[test]
fn since_returns_notes_already_above_the_mark_at_once() {
    let w = world();
    w.cards(&["watched"]);
    let first = w.note(1, "leader", "local", "LEADER one");
    let second = w.note(1, "leader", "local", "LEADER two");
    let third = w.note(1, "leader", "local", &format!("{}\nrest", "x".repeat(300)));
    let at_first = first.to_string();
    // A poll every 300 s within a 30 s timeout: only the first poll can return before the
    // timeout, so `note` here means the first poll found them.
    let (out, elapsed) = w.wait(&[
        "--card",
        "C-1",
        "--since",
        &at_first,
        "--timeout",
        "30s",
        "--every",
        "300",
    ]);
    let r = result(&out);
    assert_eq!(r["outcome"], json!("note"), "{r}");
    assert_eq!(r["mark"], json!(first), "{r}");
    assert_eq!(
        items(&r),
        vec![item("C-1", second), item("C-1", third)],
        "{r}"
    );
    assert!(elapsed < Duration::from_secs(20), "{elapsed:?}");
    // A first line longer than the limit is cut to 200 characters, as `card inbox` cuts it.
    assert_eq!(r["items"][1]["line"], json!("x".repeat(200)));

    // The mark is a strict bound: `--since` at the newest note finds nothing and times out.
    let at_third = third.to_string();
    let (out, _) = w.wait(&[
        "--card",
        "C-1",
        "--since",
        &at_third,
        "--timeout",
        "1s",
        "--every",
        "1",
    ]);
    let r = result(&out);
    assert_eq!(r["outcome"], json!("timeout"), "{r}");
    assert_eq!(r["mark"], json!(third), "{r}");
}

#[test]
fn an_unknown_card_or_an_absent_ledger_is_refused_before_the_wait() {
    let w = world();
    w.cards(&["watched"]);
    w.note(1, "leader", "local", "LEADER one");
    let before = w.home.snapshot();
    // An unknown card among known ones, by number or by host number: refused at once, although
    // the timeout is long.
    for card in ["C-9", "C-1,C-9", "#77"] {
        let (out, elapsed) = w.wait(&["--card", card, "--timeout", "50s"]);
        let (kind, message) = refusal(&out);
        assert_eq!(kind, "not_found", "{card}: {message}");
        assert!(elapsed < Duration::from_secs(20), "{card}: {elapsed:?}");
    }
    // A name that is no card name.
    let (out, _) = w.wait(&["--card", "X-1", "--timeout", "50s"]);
    let (kind, message) = refusal(&out);
    assert_eq!(kind, "invalid_input", "{message}");
    assert!(message.contains("card wait refused"), "{message}");
    // A wait that finds its note reads only: no file, folder, size or time changed.
    let (out, _) = w.wait(&["--card", "C-1", "--since", "0", "--timeout", "30s"]);
    assert_eq!(result(&out)["outcome"], json!("note"));
    assert_eq!(w.home.snapshot(), before, "card wait wrote nothing");

    // A HOME without a ledger: refused, and nothing is created.
    let bare = TempHome::new();
    let project = bare.project("demo").to_str().unwrap().to_string();
    let snapshot = bare.snapshot();
    let out = bare
        .command(&[
            "card",
            "wait",
            "--card",
            "C-1",
            "--timeout",
            "50s",
            "--project",
            &project,
        ])
        .output()
        .unwrap();
    let (kind, message) = refusal(&out);
    assert_eq!(kind, "not_found", "{message}");
    assert!(message.contains("ledger not found"), "{message}");
    assert_eq!(bare.snapshot(), snapshot, "nothing was created");

    // A file that is not the ledger: the ledger's own refusal, and the file is left alone.
    let path = ledger::ledger_path(w.home.path(), "demo").unwrap();
    std::fs::write(&path, b"this is not a database").unwrap();
    let before = w.home.snapshot();
    let (out, _) = w.wait(&["--card", "C-1", "--timeout", "50s"]);
    let (kind, message) = refusal(&out);
    assert_eq!(kind, "blocked", "{message}");
    assert!(message.contains("ledger_unavailable"), "{message}");
    assert_eq!(w.home.snapshot(), before);
}
