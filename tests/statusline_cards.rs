//! Card C-65 (user instructions 2026-11-16 and 2026-11-17): the repository line of the status
//! line shows the cards of the project's local ledger right after the forge counts,
//! `⬇️ 46 🔀 0 | 🗂️ 50 ▶️ 2 🔃 3 ✅ 12 | 📡 owner/repo | …`: `🗂️` cards created and
//! remaining (`draft` + `ready` + `hold`), `▶️` cards being worked on (`active` without a waiting
//! marker note), `🔃` cards pushed and waiting for integration (every `review` card and each
//! `active` card whose newest marker note is a waiting one), `✅` cards `done` (user decision
//! 2026-11-19: this order and these icons).
//!
//! Every case runs the real binary through `TempHome` (a temporary HOME, nothing of the real
//! environment but a PATH for git) against a ledger built in that HOME. The read is local and
//! read-only: the HOME tree is compared entry by entry, bytes and modification times, before and
//! after each render.

mod common;

use common::home::TempHome;
use moai_cli::ledger;
use std::io::Write;
use std::path::{Path, PathBuf};
use std::process::Stdio;

/// The states the fixture ledger holds: 3 + 2 in progress, 20 + 25 + 5 remaining, 12 done.
const MIX: [(&str, usize); 6] = [
    ("active", 3),
    ("review", 2),
    ("draft", 20),
    ("ready", 25),
    ("hold", 5),
    ("done", 12),
];

const FORGE_CACHE: &str = ".moai-cli/state/github/counts.json";

struct Fixture {
    home: TempHome,
    project: PathBuf,
}

/// A HOME with the project `demo` (profile, a git repository on `main` with one commit) and
/// fresh forge counts `46 / 0`. No ledger yet.
fn fixture() -> Fixture {
    let home = TempHome::new();
    let project = home.project("demo");
    common::git::git_in(&project, &["init", "-q", "-b", "main"]);
    common::git::git_in(
        &project,
        &["commit", "-q", "--allow-empty", "-m", "fixture"],
    );
    std::fs::create_dir_all(project.join(FORGE_CACHE).parent().unwrap()).unwrap();
    std::fs::write(
        project.join(FORGE_CACHE),
        "{\"open_issues\":46,\"open_prs\":0,\"fetched_at\":4102444800}\n",
    )
    .unwrap();
    Fixture { home, project }
}

impl Fixture {
    fn register(&self) {
        ledger::register(self.home.path(), "demo").unwrap();
    }

    /// Adds `n` cards in `state`. An `INSERT` has no transition check, so every state can be
    /// written directly.
    fn add(&self, state: &str, n: usize) {
        let mut l = ledger::open_write(self.home.path(), "demo").unwrap();
        l.with_write_tx(|tx, now| {
            for _ in 0..n {
                tx.execute(
                    "INSERT INTO card (title, state, created_by, created_at, updated_at, closed_at) \
                     VALUES ('t', ?1, 'test', ?2, ?2, iif(?1 = 'done', ?2, NULL))",
                    rusqlite::params![state, now],
                )?;
            }
            Ok(())
        })
        .unwrap();
    }

    /// The [`MIX`], and one of the `active` cards carries a note whose first line is the
    /// leader-bound `LANE-READY-INTEGRATION` marker: it waits for integration, the other two
    /// `active` cards (no notes) are being worked on, and the 2 `review` cards wait.
    fn with_mix(&self) {
        self.with_cards();
        let mut l = ledger::open_write(self.home.path(), "demo").unwrap();
        l.with_write_tx(|tx, now| {
            tx.execute(
                "INSERT INTO card_note (card_id, kind, author, body, created_at) \
                 SELECT min(card_id), 'report', 'test', \
                 'LANE-READY-INTEGRATION C-1 head=abc spec=SPEC-20261119-x', ?1 \
                 FROM card WHERE state = 'active'",
                rusqlite::params![now],
            )?;
            Ok(())
        })
        .unwrap();
    }

    /// The [`MIX`] cards alone, with no note: the 3 `active` cards are all being worked on.
    fn with_cards(&self) {
        self.register();
        for (state, n) in MIX {
            self.add(state, n);
        }
    }

    fn ledger_file(&self) -> PathBuf {
        ledger::ledger_path(self.home.path(), "demo").unwrap()
    }

    /// [`Fixture::run`], and the HOME (the project is inside it) must be left exactly as it
    /// was: no entry added or removed, no byte and no modification time changed.
    fn render_from(&self, cwd: &Path, env: &[(&str, &str)]) -> String {
        let before = self.home.snapshot();
        let out = self.run(cwd, env);
        assert_eq!(
            self.home.snapshot(),
            before,
            "the render changed the HOME or the project"
        );
        out
    }

    /// One render as Claude Code runs it: the session's JSON on stdin naming the project. The
    /// exit code must be 0. Returns stdout.
    fn run(&self, cwd: &Path, env: &[(&str, &str)]) -> String {
        let stdin = serde_json::json!({
            "cwd": self.project,
            "model": {"display_name": "Opus"},
            "workspace": {
                "current_dir": self.project,
                "project_dir": self.project,
                "repo": {"host": "github.com", "owner": "modu-ai", "name": "moai-adk"},
            },
        })
        .to_string();
        let mut cmd = self
            .home
            .command(&["statusline", "--render-budget-ms", "3600000"]);
        cmd.current_dir(cwd)
            .env("PATH", std::env::var_os("PATH").unwrap_or_default())
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped());
        for (k, v) in env {
            cmd.env(k, v);
        }
        let mut child = cmd.spawn().unwrap();
        child
            .stdin
            .take()
            .unwrap()
            .write_all(stdin.as_bytes())
            .unwrap();
        let out = child.wait_with_output().unwrap();
        assert_eq!(
            out.status.code(),
            Some(0),
            "{}",
            String::from_utf8_lossy(&out.stderr)
        );
        String::from_utf8(out.stdout).unwrap()
    }

    fn plain(&self) -> String {
        self.render_from(&self.project, &[("NO_COLOR", "1")])
    }

    fn line2(&self) -> String {
        self.plain().lines().nth(1).unwrap_or_default().to_string()
    }
}

/// `+1`: the project's untracked state folder.
const WITHOUT_CARDS: &str = "⬇️ 46 🔀 0 | 📡 modu-ai/moai-adk | 🅱️ main +1";

#[test]
fn the_repository_line_shows_the_card_counts_after_the_forge_counts() {
    let f = fixture();
    let baseline = f.plain();
    assert_eq!(baseline.lines().nth(1), Some(WITHOUT_CARDS), "no ledger");
    assert_eq!(baseline.lines().count(), 3, "{baseline}");
    f.with_mix();
    let with = f.plain();
    assert_eq!(
        with.lines().nth(1),
        Some("⬇️ 46 🔀 0 | 🗂️ 50 ▶️ 2 🔃 3 ✅ 12 | 📡 modu-ai/moai-adk | 🅱️ main +1")
    );
    // Only the repository line differs from the render without a ledger.
    assert_eq!(with.lines().next(), baseline.lines().next());
    assert_eq!(with.lines().nth(2), baseline.lines().nth(2));
    assert_eq!(with.lines().count(), 3);
}

#[test]
fn the_card_counts_keep_their_place_without_forge_counts() {
    let f = fixture();
    f.with_mix();
    std::fs::remove_file(f.project.join(FORGE_CACHE)).unwrap();
    assert_eq!(
        f.line2(),
        "🗂️ 50 ▶️ 2 🔃 3 ✅ 12 | 📡 modu-ai/moai-adk | 🅱️ main +1"
    );
}

/// SPEC-20261120-statusline-line2-order REQ-8: the renderer shortens nothing and reads no
/// `COLUMNS`, so the whole output, line 2 with all its groups included, has the same bytes under
/// 80 and 40 columns, with and without color.
#[test]
fn line_two_has_the_same_bytes_under_80_and_40_columns() {
    let f = fixture();
    f.with_mix();
    let render = |columns: &str, plain: bool| {
        let env: Vec<(&str, &str)> = if plain {
            vec![("COLUMNS", columns), ("NO_COLOR", "1")]
        } else {
            vec![("COLUMNS", columns), ("COLORTERM", "truecolor")]
        };
        f.render_from(&f.project, &env)
    };
    for plain in [true, false] {
        let wide = render("80", plain);
        let narrow = render("40", plain);
        assert_eq!(wide, narrow, "plain={plain}");
        // The colored line carries the plain text (the colored-line test pins that), so the
        // shape is read from the plain render.
        assert_eq!(
            f.plain().lines().nth(1),
            Some("⬇️ 46 🔀 0 | 🗂️ 50 ▶️ 2 🔃 3 ✅ 12 | 📡 modu-ai/moai-adk | 🅱️ main +1"),
            "plain={plain}"
        );
    }
}

#[test]
fn a_registered_ledger_without_cards_shows_zeros() {
    let f = fixture();
    f.register();
    assert_eq!(
        f.line2(),
        "⬇️ 46 🔀 0 | 🗂️ 0 ▶️ 0 🔃 0 ✅ 0 | 📡 modu-ai/moai-adk | 🅱️ main +1"
    );
}

/// `done` cards are counted in `✅` only; a ledger of nothing but done cards has no card in
/// progress and none remaining, and the other way round.
#[test]
fn done_cards_are_counted_only_in_the_done_segment() {
    let f = fixture();
    f.register();
    f.add("done", 4);
    assert!(
        f.line2().contains("| 🗂️ 0 ▶️ 0 🔃 0 ✅ 4 |"),
        "{}",
        f.line2()
    );
    f.add("active", 1);
    f.add("hold", 2);
    assert!(
        f.line2().contains("| 🗂️ 2 ▶️ 1 🔃 0 ✅ 4 |"),
        "{}",
        f.line2()
    );
}

/// A ledger another process holds exclusively, a file that is not a database, a database cut
/// short and a HOME of the Go generation (where no ledger is registered): exit 0 (checked by
/// `render_from`), the output is byte for byte the one without a ledger, and the file is not
/// touched.
#[test]
fn an_unreadable_ledger_leaves_the_line_as_it_is_without_one() {
    let f = fixture();
    let baseline = f.plain();
    f.with_mix();
    assert_ne!(f.plain(), baseline);

    // While the lock is held this process must not open the ledger file a second time
    // (closing any descriptor of a file drops the process's POSIX locks on it), so the tree is
    // compared around the lock, not inside it.
    let before = f.home.snapshot();
    let holder = rusqlite::Connection::open(f.ledger_file()).unwrap();
    holder.execute_batch("BEGIN EXCLUSIVE").unwrap();
    let started = std::time::Instant::now();
    let locked = f.run(&f.project, &[("NO_COLOR", "1")]);
    let waited = started.elapsed();
    assert_eq!(locked, baseline, "locked");
    assert!(
        waited < ledger::open::BUSY_TIMEOUT / 2,
        "a locked ledger held the render for {waited:?}"
    );
    holder.execute_batch("ROLLBACK").unwrap();
    drop(holder);
    assert_eq!(
        f.home.snapshot(),
        before,
        "the locked render changed nothing"
    );
    assert_ne!(f.plain(), baseline, "readable again once released");

    let good = std::fs::read(f.ledger_file()).unwrap();
    for bytes in [b"not a database".to_vec(), good[..good.len() / 2].to_vec()] {
        std::fs::write(f.ledger_file(), &bytes).unwrap();
        assert_eq!(f.plain(), baseline);
        assert_eq!(std::fs::read(f.ledger_file()).unwrap(), bytes);
    }

    let go = fixture();
    let go_baseline = go.plain();
    go.home.make_go_generation();
    assert!(ledger::register(go.home.path(), "demo").is_err());
    assert_eq!(go.plain(), go_baseline, "a HOME that is not ready");
}

/// `statusline.yaml` switches the segment off by its id `cards`, like every other segment; the
/// other segments of the line stay.
#[test]
fn the_cards_option_hides_the_segment() {
    let f = fixture();
    let baseline = f.plain();
    f.with_mix();
    let config = f.project.join(".moai-cli/config/sections/statusline.yaml");
    std::fs::create_dir_all(config.parent().unwrap()).unwrap();
    std::fs::write(&config, "statusline:\n  segments:\n    cards: false\n").unwrap();
    assert_eq!(f.plain(), baseline);
    std::fs::write(&config, "statusline:\n  segments:\n    cards: true\n").unwrap();
    assert!(f.line2().contains("🗂️ 50 ▶️ 2 🔃 3 ✅ 12"), "{}", f.line2());
    std::fs::write(
        &config,
        "statusline:\n  segments:\n    github: false\n    git_branch: false\n",
    )
    .unwrap();
    assert_eq!(f.line2(), "🗂️ 50 ▶️ 2 🔃 3 ✅ 12");
}

/// The project is the one the input names, not the working directory: from a folder outside
/// the project (no repository there, so no branch) the counts are still this project's.
#[test]
fn the_counts_follow_the_project_of_the_input_not_the_working_directory() {
    let f = fixture();
    f.with_mix();
    let elsewhere = tempfile::tempdir().unwrap();
    let out = f.render_from(elsewhere.path(), &[("NO_COLOR", "1")]);
    assert_eq!(out.lines().nth(1), Some("🗂️ 50 ▶️ 2 🔃 3 ✅ 12"), "{out}");
}

/// With color the text is the same once the escapes are removed: each number keeps its icon.
#[test]
fn the_colored_line_carries_the_same_text() {
    let f = fixture();
    f.with_mix();
    let colored = f.render_from(&f.project, &[("COLORTERM", "truecolor")]);
    assert!(colored.contains("\x1b[38;2;"), "{colored:?}");
    let mut stripped = String::new();
    let mut rest = colored.as_str();
    while let Some(at) = rest.find("\x1b[") {
        stripped.push_str(&rest[..at]);
        let tail = &rest[at..];
        rest = &tail[tail.find('m').unwrap() + 1..];
    }
    stripped.push_str(rest);
    assert_eq!(stripped, f.plain());
}

/// Writes the two ledgers the golden cases copy into their HOME, project `demo`, both the
/// [`MIX`]: `cards-5-50-12.sqlite` (case `b13-card-counts`) with no note, and
/// `cards-5-50-12-waiting.sqlite` (case `b14-card-counts-waiting`) with the
/// `LANE-READY-INTEGRATION` note of [`Fixture::with_mix`].
/// Not a check: run it by name with `--ignored` after a ledger schema change, then record the
/// goldens again.
#[test]
#[ignore = "regenerates the checked-in fixtures"]
fn write_the_golden_ledger_fixture() {
    let dir = Path::new(env!("CARGO_MANIFEST_DIR")).join("tests/fixtures/statusline-claude/ledger");
    std::fs::create_dir_all(&dir).unwrap();
    let plain = fixture();
    plain.with_cards();
    std::fs::copy(plain.ledger_file(), dir.join("cards-5-50-12.sqlite")).unwrap();
    let waiting = fixture();
    waiting.with_mix();
    std::fs::copy(
        waiting.ledger_file(),
        dir.join("cards-5-50-12-waiting.sqlite"),
    )
    .unwrap();
}
