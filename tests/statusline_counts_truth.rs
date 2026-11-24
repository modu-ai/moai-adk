//! SPEC-20261124-statusline-counts-truth: the counts on the flow line against the ledger.
//!
//! The fixture is the shape of the real ledger when the user reported the line
//! `🗂️ 81 → 🟡 3 → 📨 9 → 🛬 6 → ✅ 50 | 🚧 8 🙋 96` (2026-11-24T07:32Z): which cards are in which
//! state, which notes they carry, how their reservations ended. The composition below is a
//! reconstruction made to reproduce those seven numbers; the real ledger is never read. Every
//! case runs the real binary through `TempHome` against a ledger built in that HOME.

mod common;

use common::home::TempHome;
use moai_cli::ledger;
use rusqlite::{params, Connection};
use std::io::Write;
use std::path::PathBuf;
use std::process::Stdio;

const HOUR: i64 = 3_600;
const DAY: i64 = 86_400;
/// The render's clock, 2026-11-24T07:32:00Z (`--now-unix`): the hour of the user's report. A fixed
/// clock keeps every count independent of the hour the test runs at.
const T0: i64 = 1_795_505_520;

struct Fixture {
    home: TempHome,
    project: PathBuf,
}

fn fixture() -> Fixture {
    let home = TempHome::new();
    let project = home.project("demo");
    common::git::git_in(&project, &["init", "-q", "-b", "main"]);
    common::git::git_in(
        &project,
        &["commit", "-q", "--allow-empty", "-m", "fixture"],
    );
    ledger::register(home.path(), "demo").unwrap();
    Fixture { home, project }
}

// ---- ledger rows -------------------------------------------------------------------------

fn card(tx: &Connection, state: &str, now: i64, closed_at: Option<i64>) -> i64 {
    tx.execute(
        "INSERT INTO card (title, state, created_by, created_at, updated_at, closed_at) \
         VALUES ('t', ?1, 'test', ?2, ?2, ?3)",
        params![state, now - 30 * DAY, closed_at],
    )
    .unwrap();
    tx.last_insert_rowid()
}

/// A reservation of `card`: its generation, status, when it was taken and when its lease ends.
fn reservation(
    tx: &Connection,
    card: i64,
    generation: i64,
    status: &str,
    taken: i64,
    lease_until: i64,
) {
    tx.execute(
        "INSERT INTO reservation (card_id, generation, holder, status, acquired_at, lease_until, \
         updated_at, worktree, branch) VALUES (?1, ?2, ?3, ?4, ?5, ?6, ?5, ?7, 'work')",
        params![
            card,
            generation,
            format!("lane-{card}"),
            status,
            taken,
            lease_until,
            format!("/nonexistent/wt-{card}")
        ],
    )
    .unwrap();
}

fn note(tx: &Connection, card: i64, kind: &str, author: &str, body: &str, at: i64) {
    tx.execute(
        "INSERT INTO card_note (card_id, kind, author, body, created_at) \
         VALUES (?1, ?2, ?3, ?4, ?5)",
        params![card, kind, author, body, at],
    )
    .unwrap();
}

fn event(tx: &Connection, card: i64, kind: &str, at: i64) {
    tx.execute(
        "INSERT INTO event (event_key, at, actor, card_id, kind, payload) \
         VALUES (?1, ?2, 'test', ?3, ?4, '{\"reason\":\"fixture\"}')",
        params![format!("{kind}-{card}-{at}"), at, card, kind],
    )
    .unwrap();
}

const HEAD: &str = "0123456789abcdef0123456789abcdef01234567";

fn ready_line(marker: &str, card: i64) -> String {
    format!("{marker} C-{card} head={HEAD}")
}

fn audit_pass(card: i64) -> String {
    format!("AUDIT C-{card} (scope) head {HEAD} base {HEAD} — verdict: PASS")
}

/// The ledger of 2026-11-24T07:32Z. Old numbers in the comments are what the status line drew.
fn build_ledger(f: &Fixture) {
    let mut l = ledger::open_write(f.home.path(), "demo").unwrap();
    l.with_write_tx(|tx, _| {
        let now = T0;
        let live = now + HOUR;
        let ended = now - 3 * DAY;
        let taken = now - 8 * HOUR;
        let later = now - 6 * HOUR;

        // 🗂️ 81 = every `ready` card. 77 of them are what the user means (waiting for a lane).
        for _ in 0..76 {
            card(tx, "ready", now, None);
        }
        // A card taken again after its READY: the READY is older than the last take, the card
        // waits for a lane (hand-back is read after the last reserve only).
        let again = card(tx, "ready", now, None);
        reservation(tx, again, 1, "released", now - 20 * HOUR, now - 19 * HOUR);
        note(
            tx,
            again,
            "report",
            "lane",
            &ready_line("LANE-READY-INTEGRATION", again),
            now - 18 * HOUR,
        );
        reservation(tx, again, 2, "released", now - 10 * HOUR, now - 9 * HOUR);
        // Four cards a lane handed back and released: the release moved them to `ready`.
        // 2 handed over (📨) …
        for marker in ["LANE-READY-INTEGRATION", "REPAIRED"] {
            let c = card(tx, "ready", now, None);
            reservation(tx, c, 1, "released", taken, later);
            note(tx, c, "report", "lane", &ready_line(marker, c), later + 60);
        }
        // … and 2 audited (🛬).
        for _ in 0..2 {
            let c = card(tx, "ready", now, None);
            reservation(tx, c, 1, "released", taken, later);
            note(tx, c, "report", "auditor", &audit_pass(c), later + 60);
        }
        // Drafts are on no part of the line.
        for _ in 0..3 {
            card(tx, "draft", now, None);
        }

        // 🟡 3 = working with a live lease and no newer marker. One of the three is a card whose
        // lane handed it back with REPAIRED (the line did not read it).
        for _ in 0..2 {
            let c = card(tx, "active", now, None);
            reservation(tx, c, 1, "held", taken, live);
        }
        let repaired = card(tx, "active", now, None);
        reservation(tx, repaired, 1, "held", taken, live);
        note(
            tx,
            repaired,
            "report",
            "lane",
            &ready_line("REPAIRED", repaired),
            later,
        );

        // Not drawn: 5 working cards whose lease ended without a hand-back (stale).
        for _ in 0..5 {
            let c = card(tx, "active", now, None);
            reservation(tx, c, 1, "held", taken, ended);
        }

        // 📨 9 = 5 `review` cards (3 with an ended lease) and 4 `active` cards with a READY
        // (3 ended).
        for i in 0..5 {
            let c = card(tx, "review", now, None);
            let lease = if i < 3 { ended } else { live };
            reservation(tx, c, 1, "held_for_review", taken, lease);
        }
        for i in 0..4 {
            let c = card(tx, "active", now, None);
            let lease = if i < 3 { ended } else { live };
            reservation(tx, c, 1, "held", taken, lease);
            note(
                tx,
                c,
                "report",
                "lane",
                &ready_line("LANE-READY-INTEGRATION", c),
                later,
            );
        }

        // 🛬 6 = 3 audited, 2 landed, and one card that landed once and was taken again.
        for _ in 0..3 {
            let c = card(tx, "active", now, None);
            reservation(tx, c, 1, "held", taken, ended);
            note(tx, c, "report", "auditor", &audit_pass(c), later);
        }
        for _ in 0..2 {
            let c = card(tx, "review", now, None);
            reservation(tx, c, 1, "held_for_review", taken, ended);
            event(tx, c, "card.landed", later);
        }
        let reopened = card(tx, "active", now, None);
        reservation(tx, reopened, 1, "released", now - 5 * DAY, now - 5 * DAY + HOUR);
        event(tx, reopened, "card.landed", now - 4 * DAY);
        reservation(tx, reopened, 2, "held", now - HOUR, live);
        note(tx, reopened, "report", "lane", "ACK C-1 taking it again", now - HOUR + 60);

        // 🚧 8 = LANE-BLOCKED as the newest marker; 4 of them were answered by the leader's
        // decision note and the lane works on (live lease).
        for _ in 0..4 {
            let c = card(tx, "active", now, None);
            reservation(tx, c, 1, "held", taken, ended);
            note(tx, c, "report", "lane", "LANE-BLOCKED kind=decision C-1 which option", later);
        }
        for _ in 0..4 {
            let c = card(tx, "active", now, None);
            reservation(tx, c, 1, "held", taken, live);
            note(tx, c, "report", "lane", "LANE-BLOCKED kind=decision C-1 which option", later);
            note(tx, c, "decision", "leader", "LEADER 결정: 둘째 안으로 가요", later + 60);
        }

        // 🙋 96 = hold cards, every one of them withdrawn.
        for _ in 0..96 {
            let c = card(tx, "hold", now, None);
            event(tx, c, "card.withdraw", later);
        }

        // ✅ 50 = closed since 00:00 UTC; 53 more closed between 00:00 Seoul time (15:00Z the day
        // before) and 00:00 UTC: 103 for a user in Seoul, where it is 16:32.
        let day_start = now.div_euclid(DAY) * DAY;
        for _ in 0..50 {
            card(tx, "done", now, Some(day_start + HOUR));
        }
        for _ in 0..53 {
            card(tx, "done", now, Some(day_start - HOUR));
        }
        Ok(())
    })
    .unwrap();
}

// ---- the render --------------------------------------------------------------------------

/// The flow line (the second line of the output) of one render as Claude Code runs it.
/// `tz` goes to the child only (SPEC REQ: the local day of the machine's zone).
fn flow_line(f: &Fixture, tz: Option<&str>) -> String {
    let before = f.home.snapshot();
    let stdin = serde_json::json!({
        "session_id": "this-session",
        "cwd": f.project,
        "model": {"display_name": "Opus"},
        "workspace": {"current_dir": f.project, "project_dir": f.project},
    });
    let tmp = tempfile::tempdir().unwrap();
    let mut cmd = f
        .home
        .command(&[
            "statusline",
            "--render-budget-ms",
            "3600000",
            "--now-unix",
            &T0.to_string(),
        ]);
    cmd.current_dir(&f.project)
        .env("PATH", std::env::var_os("PATH").unwrap_or_default())
        .env("TMPDIR", tmp.path())
        .env("NO_COLOR", "1")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped());
    if let Some(tz) = tz {
        cmd.env("TZ", tz);
    }
    let mut child = cmd.spawn().unwrap();
    child
        .stdin
        .take()
        .unwrap()
        .write_all(stdin.to_string().as_bytes())
        .unwrap();
    let out = child.wait_with_output().unwrap();
    assert_eq!(
        out.status.code(),
        Some(0),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    assert_eq!(f.home.snapshot(), before, "the render changed the HOME");
    let text = String::from_utf8(out.stdout).unwrap();
    text.lines().nth(1).unwrap_or_default().to_string()
}

fn moai() -> String {
    format!("🗿 {}", moai_cli::BUILD_VERSION)
}

/// Characterization of the defect the user reported (2026-11-24T07:32Z): the fixture above draws
/// exactly the seven numbers of the report. The truth of the same ledger is in the tests that
/// replace this one with the revision (SPEC-20261124-statusline-counts-truth).
#[test]
fn the_0732z_ledger_shape_draws_the_seven_numbers_of_the_report() {
    let f = fixture();
    build_ledger(&f);
    assert_eq!(
        flow_line(&f, Some("UTC")),
        format!("{} | 🗂️ 81 → 🟡 3 → 📨 9 → 🛬 6 → ✅ 50 | 🚧 8 🙋 96", moai())
    );
}

/// The same ledger drawn for a user in Seoul (`TZ` for the child only): their day began at 15:00Z
/// the day before, so 103 cards were closed today, and the line still says 50 (the UTC day).
#[test]
fn the_done_count_ignores_the_time_zone_of_the_machine() {
    let f = fixture();
    build_ledger(&f);
    let seoul = flow_line(&f, Some("Asia/Seoul"));
    assert!(seoul.contains("→ ✅ 50 |"), "{seoul}");
}
