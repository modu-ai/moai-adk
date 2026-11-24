//! SPEC-20261124-card-flow-icons: the card-flow icon set of the status line and of the shipped
//! style texts is the table `moai_cli::statusline::icons`, written once.
//!
//! The line tests run the real binary through `TempHome` against a ledger built in that HOME
//! (nothing of the real environment is read); the scans walk the checked-in tree like the other
//! repo scans and are listed in `tools/repo_scans.json`.

mod common;

use common::home::TempHome;
use moai_cli::ledger;
use moai_cli::statusline::icons::{
    self, BLOCKED, CHECK, DONE, PAUSED, QUESTION, READY, RETIRED, TODO, WORK,
};
use rusqlite::{params, Connection};
use std::fs;
use std::io::Write;
use std::path::{Path, PathBuf};
use std::process::Stdio;

const HOUR: i64 = 3_600;
const DAY: i64 = 86_400;
/// The render's clock, 2026-11-24T12:00:00Z (`--now-unix`): a fixed clock keeps every count
/// independent of the hour the test runs at.
const T0: i64 = 1_795_521_600;
const HEAD: &str = "0123456789abcdef0123456789abcdef01234567";

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

fn card(tx: &Connection, state: &str, closed_at: Option<i64>) -> i64 {
    tx.execute(
        "INSERT INTO card (title, state, created_by, created_at, updated_at, closed_at) \
         VALUES ('t', ?1, 'test', ?2, ?2, ?3)",
        params![state, T0 - 30 * DAY, closed_at],
    )
    .unwrap();
    tx.last_insert_rowid()
}

fn reservation(tx: &Connection, card: i64, status: &str, taken: i64, lease_until: i64) {
    tx.execute(
        "INSERT INTO reservation (card_id, generation, holder, status, acquired_at, lease_until, \
         updated_at, worktree, branch) VALUES (?1, 1, ?2, ?3, ?4, ?5, ?4, ?6, 'work')",
        params![
            card,
            format!("lane-{card}"),
            status,
            taken,
            lease_until,
            format!("/nonexistent/wt-{card}")
        ],
    )
    .unwrap();
}

fn note(tx: &Connection, card: i64, body: &str, at: i64) {
    tx.execute(
        "INSERT INTO card_note (card_id, kind, author, body, created_at) \
         VALUES (?1, 'report', 'lane', ?2, ?3)",
        params![card, body, at],
    )
    .unwrap();
}

fn event(tx: &Connection, card: i64, kind: &str) {
    tx.execute(
        "INSERT INTO event (event_key, at, actor, card_id, kind, payload) \
         VALUES (?1, ?2, 'test', ?3, ?4, '{}')",
        params![format!("{kind}-{card}"), T0 - HOUR, card, kind],
    )
    .unwrap();
}

/// The flow line (the second line of the output) of one render as Claude Code runs it.
fn flow_line(f: &Fixture, columns: Option<&str>) -> String {
    let before = f.home.snapshot();
    let stdin = serde_json::json!({
        "session_id": "this-session",
        "cwd": f.project,
        "model": {"display_name": "Opus"},
        "workspace": {"current_dir": f.project, "project_dir": f.project},
    });
    let tmp = tempfile::tempdir().unwrap();
    let mut cmd = f.home.command(&[
        "statusline",
        "--render-budget-ms",
        "3600000",
        "--now-unix",
        &T0.to_string(),
    ]);
    cmd.current_dir(&f.project)
        .env("PATH", std::env::var_os("PATH").unwrap_or_default())
        .env("TMPDIR", tmp.path())
        .env("TZ", "UTC")
        .env("NO_COLOR", "1")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped());
    if let Some(columns) = columns {
        cmd.env("COLUMNS", columns);
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
    let line = text.lines().nth(1).unwrap_or_default();
    // The version opens the line: what follows is the flow.
    let version = format!("🗿 {}", moai_cli::BUILD_VERSION);
    line.strip_prefix(&format!("{version} | "))
        .or_else(|| line.strip_prefix(&version))
        .unwrap_or(line)
        .to_string()
}

/// One card in every stage of the road, three cards beside it (one blocked, a withdrawn and a
/// folded hold card, two plain hold cards) and a draft that is on no part of the line.
fn build_every_stage(f: &Fixture) {
    let mut l = ledger::open_write(f.home.path(), "demo").unwrap();
    l.with_write_tx(|tx, _| {
        let live = T0 + HOUR;
        let taken = T0 - 8 * HOUR;
        let later = T0 - 6 * HOUR;
        // to do: three ready cards nobody took.
        for _ in 0..3 {
            card(tx, "ready", None);
        }
        // work: two active cards with a live lease and no marker.
        for _ in 0..2 {
            let c = card(tx, "active", None);
            reservation(tx, c, "held", taken, live);
        }
        // check: one review card.
        let c = card(tx, "review", None);
        reservation(tx, c, "held_for_review", taken, live);
        // ready to land: an audited card a lane handed back and released.
        let c = card(tx, "ready", None);
        reservation(tx, c, "released", taken, later);
        note(
            tx,
            c,
            &format!("AUDIT C-{c} (scope) head {HEAD} base {HEAD} — verdict: PASS"),
            later + 60,
        );
        // done: two cards closed since midnight.
        for _ in 0..2 {
            card(tx, "done", Some(T0 - HOUR));
        }
        // blocked: one active card, the lane asked for a decision.
        let c = card(tx, "active", None);
        reservation(tx, c, "held", taken, live);
        note(tx, c, "LANE-BLOCKED kind=decision C-1 which option", later);
        // paused: a withdrawn and a folded hold card.
        for kind in ["card.withdraw", "card.fold"] {
            let c = card(tx, "hold", None);
            event(tx, c, kind);
        }
        // question: two hold cards nobody withdrew.
        for _ in 0..2 {
            card(tx, "hold", None);
        }
        // a draft: on no part of the line.
        card(tx, "draft", None);
        Ok(())
    })
    .unwrap();
}

/// AC-2 (REQ-2, REQ-3, REQ-4): a card in every stage draws the road in the order of the table,
/// the blocked, paused and question counts stand beside it, and the withdrawn and the folded hold
/// cards are paused while the plain ones are questions.
#[test]
fn every_stage_draws_its_item_of_the_table() {
    let f = fixture();
    build_every_stage(&f);
    assert_eq!(
        flow_line(&f, None),
        format!(
            "{TODO} 3 → {WORK} 3 → {CHECK} 1 → {READY} 1 → {DONE} 2 | {BLOCKED} 1 {PAUSED} 2 {QUESTION} 2"
        )
    );
}

/// AC-2 (REQ-3): `🙋` shows only when a user question exists, `⏸️` shows what was withdrawn.
#[test]
fn the_question_item_shows_only_when_a_user_pick_is_owed() {
    let f = fixture();
    {
        let mut l = ledger::open_write(f.home.path(), "demo").unwrap();
        l.with_write_tx(|tx, _| {
            card(tx, "ready", None);
            for kind in ["card.withdraw", "card.fold"] {
                let c = card(tx, "hold", None);
                event(tx, c, kind);
            }
            Ok(())
        })
        .unwrap();
    }
    assert_eq!(flow_line(&f, None), format!("{TODO} 1 | {PAUSED} 2"));
    {
        let mut l = ledger::open_write(f.home.path(), "demo").unwrap();
        l.with_write_tx(|tx, _| {
            card(tx, "hold", None);
            Ok(())
        })
        .unwrap();
    }
    assert_eq!(
        flow_line(&f, None),
        format!("{TODO} 1 | {PAUSED} 2 {QUESTION} 1")
    );
}

/// AC-2 (REQ-6): at a narrow width the leftmost flow item goes first and what stands beside the
/// road stays to the end.
#[test]
fn a_narrow_width_drops_the_road_and_keeps_what_stands_beside_it() {
    let f = fixture();
    build_every_stage(&f);
    let beside = format!("{BLOCKED} 1 {PAUSED} 2 {QUESTION} 2");
    let narrow = flow_line(&f, Some("24"));
    assert!(narrow.ends_with(&beside), "{narrow}");
    assert!(!narrow.contains(TODO), "{narrow}");
    let narrowest = flow_line(&f, Some("5"));
    assert_eq!(narrowest, beside);
}

// ---- the scans ---------------------------------------------------------------------------

const ROOT: &str = env!("CARGO_MANIFEST_DIR");

/// AC-3 (REQ-10): none of the retired icons is in the plain goldens the status line draws the
/// flow in. The shipped texts are scanned by `style_banner_scan`, which walks the skill and the
/// templates; here the goldens are read by name, and the rest of them are held by the golden
/// test itself.
#[test]
fn no_retired_icon_is_in_the_plain_goldens_of_the_flow_line() {
    let mut seen = 0;
    for name in ["b13-card-counts", "b14-card-counts-waiting"] {
        let path = Path::new(ROOT)
            .join("tests/fixtures/statusline-claude/golden")
            .join(format!("{name}.plain.txt"));
        let text = fs::read_to_string(&path).unwrap();
        for icon in RETIRED {
            assert!(!text.contains(icon), "{icon} in {}", path.display());
        }
        // The flow line is there and draws the table's icons.
        assert!(text.contains(&format!("{TODO} ")), "{}", path.display());
        seen += 1;
    }
    assert_eq!(seen, 2);
}

/// AC-6 (REQ-12): the tests of the status line and of the style texts build the icons of the road
/// and of what stands beside it from the table, so a changed icon is a change of the table. A
/// literal of one of these icons in those test files is the miss; `✅` and `🙋` are also banner
/// and board icons elsewhere and are not scanned.
#[test]
fn the_tests_read_the_icons_from_the_table() {
    let scanned = [
        "tests/statusline_flow.rs",
        "tests/statusline_cards.rs",
        "tests/statusline_counts_truth.rs",
        "tests/statusline_card_flow_icons.rs",
        "tests/style_banner_scan.rs",
        "tests/home_style_ownership_cli.rs",
    ];
    let literals = [TODO, WORK, CHECK, READY, BLOCKED, PAUSED];
    let mut offenders = Vec::new();
    for rel in scanned {
        let text = fs::read_to_string(Path::new(ROOT).join(rel)).unwrap();
        for (n, line) in text.lines().enumerate() {
            // Comments say what a line drew in the report that started the work; the check is on
            // code and strings.
            if line.trim_start().starts_with("//") {
                continue;
            }
            for icon in literals {
                if line.contains(icon) {
                    offenders.push(format!("{rel}:{}: {icon}", n + 1));
                }
            }
        }
    }
    assert!(
        offenders.is_empty(),
        "an icon of the set is spelled by hand:\n{}",
        offenders.join("\n")
    );
    // The table the scan reads is the one the line draws.
    assert_eq!(icons::status_line_icons().count(), 8);
}
