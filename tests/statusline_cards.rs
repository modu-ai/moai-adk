//! The flow line of the status line (SPEC-20261122-statusline-flow; it grew out of card C-65's
//! card counts, user instructions 2026-11-16 and 2026-11-17): the cards of the project's local
//! ledger by the stage of their progress, `👀 S3 A0 L0 | ⛔ B0 ⏸️  H5 | 🗂️  D20 R25 W0 E2 ✅ 12`,
//! on a line of its own between the first line and the usage windows. The first line says who
//! and where (SPEC-20261122-statusline-where): `🧠 Opus | 📡 owner/repo, ⬇️  46 🔀 0 | 🅱️  …`.
//!
//! Every case runs the real binary through `TempHome` (a temporary HOME, nothing of the real
//! environment but a PATH for git) against a ledger built in that HOME. The read is local and
//! read-only: the HOME tree is compared entry by entry, bytes and modification times, before and
//! after each render.

mod common;

use common::home::TempHome;
use moai_cli::ledger;
use moai_cli::statusline::width::display_width;
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
    /// leader-bound `LANE-READY-INTEGRATION` marker: it is submitted, the other two `active`
    /// cards (no notes, no reservation) are stale, and the 2 `review` cards are submitted.
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

    /// The [`MIX`] cards alone, with no note: the 3 `active` cards have no marker and no
    /// reservation, so all three are stale.
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
        // The render may keep a session cache under `$TMPDIR`: a folder of its own outside the
        // HOME, so the HOME snapshot stays equal and nothing reaches the real temporary folder.
        let tmp = tempfile::tempdir().unwrap();
        let mut cmd = self
            .home
            .command(&["statusline", "--render-budget-ms", "3600000"]);
        cmd.current_dir(cwd)
            .env("PATH", std::env::var_os("PATH").unwrap_or_default())
            .env("TMPDIR", tmp.path())
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

    /// The plain render with `COLUMNS` set to `columns`.
    fn plain_at(&self, columns: &str) -> String {
        self.render_from(&self.project, &[("NO_COLOR", "1"), ("COLUMNS", columns)])
    }

    /// The flow line when the ledger has cards: the second line, under the first line.
    fn flow_line(&self) -> String {
        self.plain().lines().nth(1).unwrap_or_default().to_string()
    }

    /// The first line: who and where.
    fn line1(&self) -> String {
        self.plain().lines().next().unwrap_or_default().to_string()
    }
}

/// The first line of the fixture; `+1`: the project's untracked state folder.
const WITHOUT_CARDS: &str = "🧠 Opus | 📡 modu-ai/moai-adk, ⬇️  46 🔀 0 | 🅱️  main +1";
/// The flow line of the [`MIX`] ledger with the note of [`Fixture::with_mix`]: 3 submitted (2
/// review, 1 by its note), 2 stale, 5 on hold, 20 draft, 25 ready, 12 done today, no card of
/// mine (the session holds none).
const FLOW_MIX: &str = "👀 S3 A0 L0 | ⛔ B0 ⏸️  H5 | 🗂️  D20 R25 W0 E2 ✅ 12";
/// The flow line of the [`MIX`] cards alone: the 3 `active` cards are stale, the 2 `review`
/// cards submitted.
const FLOW_CARDS: &str = "👀 S2 A0 L0 | ⛔ B0 ⏸️  H5 | 🗂️  D20 R25 W0 E3 ✅ 12";

#[test]
fn the_flow_line_stands_between_the_first_line_and_the_usage_windows() {
    let f = fixture();
    let baseline = f.plain();
    assert_eq!(baseline.lines().next(), Some(WITHOUT_CARDS), "no ledger");
    assert_eq!(baseline.lines().count(), 2, "{baseline}");
    assert!(
        !baseline.contains("🗂️"),
        "no ledger: no flow line and no card icon"
    );
    f.with_mix();
    let with = f.plain();
    assert_eq!(
        with.lines().next(),
        Some(WITHOUT_CARDS),
        "line 1 keeps the model, the repository, the forge counts, the branch"
    );
    assert_eq!(with.lines().nth(1), Some(FLOW_MIX), "{with}");
    assert_eq!(
        with.lines().nth(2),
        baseline.lines().nth(1),
        "the usage windows follow, unchanged"
    );
    assert!(!f.line1().contains("🗂️"), "line 1 keeps no card counts");
    // Line 1 does not change from the render without a ledger.
    assert_eq!(with.lines().next(), baseline.lines().next());
    assert_eq!(with.lines().count(), 3);
}

/// Without a marker note an `active` card nobody holds is stale, and a `review` card is
/// submitted whatever its notes say.
#[test]
fn active_cards_nobody_holds_are_stale_and_review_cards_are_submitted() {
    let f = fixture();
    f.with_cards();
    assert_eq!(f.flow_line(), FLOW_CARDS);
}

#[test]
fn the_flow_line_keeps_its_place_without_fresh_forge_counts() {
    let f = fixture();
    f.with_mix();
    std::fs::remove_file(f.project.join(FORGE_CACHE)).unwrap();
    // SPEC-20261121-statusline-detached REQ-3: a failed forge read says `n/a` in place —
    // never a zero and never a hidden segment.
    assert_eq!(
        f.line1(),
        "🧠 Opus | 📡 modu-ai/moai-adk, ⬇️  n/a 🔀 n/a | 🅱️  main +1",
        "without fresh forge counts the slots say unknown, with the comma kept"
    );
    assert_eq!(f.flow_line(), FLOW_MIX);
}

/// SPEC-20261120-statusline-ahead: under M-Flow the local branch runs ahead of its upstream
/// between releases, and the branch says so, before behind and dirty. With ahead 0 nothing
/// appears — every other test here pins that byte-identical shape.
#[test]
fn the_branch_shows_the_ahead_count_against_the_upstream() {
    let f = fixture();
    common::git::git_in(
        &f.project,
        &["commit", "-q", "--allow-empty", "-m", "second"],
    );
    common::git::git_in(&f.project, &["branch", "base", "HEAD~1"]);
    common::git::git_in(&f.project, &["branch", "--set-upstream-to=base"]);
    let line = f.line1();
    assert!(line.contains("🅱️  main ↑1"), "the ahead count: {line}");
    assert!(!line.contains('↓'), "behind 0 stays hidden: {line}");
    assert!(!line.contains("↑0"), "a zero ahead is never shown: {line}");
}

/// The forms of the flow line for the [`FLOW_MIX`] ledger, from the widest to the narrowest.
/// The narrowest form that has anything to show is the waits group: with no blocked card the
/// attention group is left out of the form without the hold count.
const MIX_FORMS: [&str; 3] = [FLOW_MIX, "👀 S3 A0 L0 | ⛔ B0 ⏸️  H5", "👀 S3 A0 L0"];

/// SPEC-20261122-statusline-flow REQ-9 (it replaces the same-bytes-at-80-and-40 rule of
/// SPEC-20261120-statusline-line2-order): each line takes the widest of its forms that fits
/// `COLUMNS` − 2 − padding. The flow line changes form exactly at each boundary; the first line
/// (AC-6 of SPEC-20261122-statusline-where) and line 3 (SPEC-20261122-statusline-limits AC-9,
/// the `line_three_ladder_` tests) have their own forms, so line 3 here is its widest form
/// wherever that fits.
#[test]
fn width_ladder_changes_the_flow_line_exactly_at_each_boundary() {
    let f = fixture();
    f.with_mix();
    let wide = f.plain_at("200");
    for (i, form) in MIX_FORMS.iter().enumerate() {
        let w = display_width(form);
        let at = |columns: usize| f.plain_at(&columns.to_string());
        let kept = at(w + 2);
        assert_eq!(
            kept.lines().nth(1),
            Some(*form),
            "COLUMNS {} keeps form {i}",
            w + 2
        );
        if let Some(next) = MIX_FORMS.get(i + 1) {
            let dropped = at(w + 1);
            assert_eq!(
                dropped.lines().nth(1),
                Some(*next),
                "COLUMNS {} drops form {i}",
                w + 1
            );
        }
        // Line 3 is its widest form wherever that fits the budget.
        let widest = wide.lines().last().unwrap_or_default();
        for (columns, out) in [(w + 2, &kept), (w + 1, &at(w + 1))] {
            if display_width(widest) <= columns - 2 {
                assert_eq!(out.lines().last(), Some(widest), "line 3 at {columns}");
            }
            assert_eq!(out.lines().count(), 3);
        }
    }
    assert_eq!(wide.lines().nth(1), Some(FLOW_MIX));
    assert_eq!(f.plain_at("120").lines().nth(1), Some(FLOW_MIX));
    assert_eq!(f.plain_at("80").lines().nth(1), Some(FLOW_MIX));
    // Below the narrowest form the narrowest is still written.
    for columns in ["20", "5", "3", "2", "1"] {
        assert_eq!(
            f.plain_at(columns).lines().nth(1),
            Some(MIX_FORMS[2]),
            "COLUMNS {columns}"
        );
    }
}

/// A missing, empty, zero or non-numeric `COLUMNS` writes the widest form: there is no width to
/// fit.
#[test]
fn width_ladder_writes_the_widest_form_without_a_usable_columns() {
    let f = fixture();
    f.with_mix();
    assert_eq!(f.flow_line(), FLOW_MIX, "COLUMNS unset");
    for columns in ["", "0", "abc", "-5", "8 0", "12x"] {
        assert_eq!(
            f.plain_at(columns).lines().nth(1),
            Some(FLOW_MIX),
            "COLUMNS {columns:?}"
        );
    }
}

/// The padding of `statusline.yaml` (the same number the generated host settings carry) is
/// subtracted from the budget on top of the margin of two.
#[test]
fn width_ladder_subtracts_the_configured_padding() {
    let f = fixture();
    f.with_mix();
    let config = f.project.join(".moai-cli/config/sections/statusline.yaml");
    std::fs::create_dir_all(config.parent().unwrap()).unwrap();
    let w = display_width(FLOW_MIX);
    for padding in [1usize, 3, 10] {
        std::fs::write(&config, format!("statusline:\n  padding: {padding}\n")).unwrap();
        assert_eq!(
            f.plain_at(&(w + 2 + padding).to_string()).lines().nth(1),
            Some(FLOW_MIX),
            "padding {padding} keeps the widest form at its edge"
        );
        assert_eq!(
            f.plain_at(&(w + 1 + padding).to_string()).lines().nth(1),
            Some(MIX_FORMS[1]),
            "padding {padding}: one column less drops a form"
        );
    }
    // A padding that is not a number reads as 0.
    for text in ["abc", "-2", "", "1.5"] {
        std::fs::write(&config, format!("statusline:\n  padding: {text}\n")).unwrap();
        assert_eq!(
            f.plain_at(&(w + 2).to_string()).lines().nth(1),
            Some(FLOW_MIX),
            "padding {text:?}"
        );
    }
}

/// A ledger with cards but no card in any open stage and nothing closed today has nothing to
/// show: the flow line is left out (zeros are not shown) and the output is the one without a
/// ledger.
#[test]
fn a_registered_ledger_without_cards_shows_no_flow_line() {
    let f = fixture();
    let baseline = f.plain();
    f.register();
    assert_eq!(f.line1(), WITHOUT_CARDS);
    assert_eq!(
        f.plain(),
        baseline,
        "nothing to show: the output is unchanged"
    );
}

/// Closed cards are counted in `✅` only (closed today, UTC); open ones in their own stage.
#[test]
fn done_cards_are_counted_only_in_the_done_segment() {
    let f = fixture();
    f.register();
    f.add("done", 4);
    assert_eq!(f.flow_line(), "🗂️  D0 R0 W0 E0 ✅ 4");
    f.add("active", 1);
    f.add("hold", 2);
    assert_eq!(
        f.flow_line(),
        "⛔ B0 ⏸️  H2 | 🗂️  D0 R0 W0 E1 ✅ 4",
        "the active card has no reservation, so it is stale"
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
    assert_eq!(f.flow_line(), FLOW_MIX);
    std::fs::write(
        &config,
        "statusline:\n  segments:\n    github: false\n    git_branch: false\n",
    )
    .unwrap();
    // Without forge counts and a branch line 1 keeps the model: the first line, the flow line,
    // the windows.
    let out = f.plain();
    assert_eq!(out.lines().count(), 3, "{out}");
    assert_eq!(out.lines().nth(1), Some(FLOW_MIX), "{out}");
}

/// The project is the one the input names, not the working directory: from a folder outside
/// the project (no repository there, so no branch) the flow is still this project's.
#[test]
fn the_flow_follows_the_project_of_the_input_not_the_working_directory() {
    let f = fixture();
    f.with_mix();
    let elsewhere = tempfile::tempdir().unwrap();
    let out = f.render_from(elsewhere.path(), &[("NO_COLOR", "1")]);
    assert!(out.lines().any(|l| l == FLOW_MIX), "{out}");
}

/// With color the text is the same once the escapes are removed: each number keeps its letter.
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

// ---- SPEC-20261122-statusline-where AC-1: the items of the first line ----
//
// Each case is a real repository in a temporary HOME and the real binary; `COLUMNS` is unset, so
// the widest form is written. The strings are the ones of the Layouts section of the SPEC.

/// The first line of the worktree session (form W0, 112 columns).
const LINE_WORKTREE: &str = "🧠 Opus 5.5 high·t | 📡 modu-ai/moai-adk, ⬇️  7 🔀 3 | 🌳 statusline-flow ↑3 +2 | 🧭 next ↑3 ↓0 | 🔗 #12 pending";
/// The first line of the main checkout on the integration branch (86 columns).
const LINE_MAIN: &str =
    "🧠 Opus 5.5 high·t | 📡 modu-ai/moai-adk, ⬇️  7 🔀 3 | 🅱️  next ↑2 +1 | 🔗 #12 pending";
/// The first line of a project without an integration branch (77 columns).
const LINE_GENERIC: &str =
    "🧠 Sonnet 5.5 | 📡 acme/webapp | 🅱️  feature/login ↑2 ↓1 +4 | 🔗 #88 approved";

/// One session: a HOME, the folder it runs in and the JSON the host would write on stdin.
struct Session {
    home: TempHome,
    cwd: PathBuf,
    stdin: serde_json::Value,
}

fn write_profile(project: &Path, integration: Option<&str>) {
    let work_flow = integration
        .map(|b| format!(",\"work_flow\":{{\"flow\":\"m-flow\",\"integration_branch\":\"{b}\"}}"))
        .unwrap_or_default();
    std::fs::write(
        moai_cli::state_root::path(project, "project.yaml"),
        format!("{{\"schema_version\":2,\"project_id\":\"demo\"{work_flow}}}\n"),
    )
    .unwrap();
}

fn write_forge_counts(project: &Path, issues: i64, prs: i64) {
    let file = project.join(FORGE_CACHE);
    std::fs::create_dir_all(file.parent().unwrap()).unwrap();
    std::fs::write(
        file,
        format!("{{\"open_issues\":{issues},\"open_prs\":{prs},\"fetched_at\":4102444800}}\n"),
    )
    .unwrap();
}

fn commit(dir: &Path, message: &str) {
    common::git::git_in(dir, &["commit", "-q", "--allow-empty", "-m", message]);
}

fn host_stdin(project: &Path, cwd: &Path, repo: (&str, &str)) -> serde_json::Value {
    serde_json::json!({
        "session_id": "s-where",
        "cwd": cwd,
        "model": {"display_name": "Opus 5.5"},
        "effort": {"level": "high"},
        "thinking": {"enabled": true},
        "workspace": {
            "current_dir": cwd,
            "project_dir": project,
            "repo": {"host": "github.com", "owner": repo.0, "name": repo.1},
        },
        "pr": {"number": 12, "review_state": "pending"},
    })
}

/// A linked worktree `statusline-flow` on its own branch of the same name, three commits ahead
/// of the integration branch `next`, one file changed and one untracked, a pull request, and
/// forge counts of 7 issues and 3 pull requests.
fn worktree_session() -> Session {
    let home = TempHome::new();
    let main = home.project("demo");
    common::git::git_in(&main, &["init", "-q", "-b", "next"]);
    std::fs::write(main.join("a.txt"), "a\n").unwrap();
    common::git::git_in(&main, &["add", "a.txt"]);
    commit(&main, "base");
    write_profile(&main, Some("next"));
    write_forge_counts(&main, 7, 3);
    let wt = home.path().join("work/statusline-flow");
    common::git::git_in(
        &main,
        &[
            "worktree",
            "add",
            "-q",
            "-b",
            "statusline-flow",
            wt.to_str().unwrap(),
        ],
    );
    for i in 1..=3 {
        commit(&wt, &format!("lane {i}"));
    }
    std::fs::write(wt.join("a.txt"), "changed\n").unwrap();
    std::fs::write(wt.join("new.txt"), "new\n").unwrap();
    let mut stdin = host_stdin(&main, &wt, ("modu-ai", "moai-adk"));
    stdin["worktree"] = serde_json::json!({
        "name": "statusline-flow", "path": wt, "branch": "statusline-flow", "original_cwd": main
    });
    Session {
        home,
        cwd: wt,
        stdin,
    }
}

/// The main checkout on `next`, two commits ahead of its upstream, with the project's untracked
/// state folder as its one change.
fn main_session() -> Session {
    let home = TempHome::new();
    let main = home.project("demo");
    common::git::git_in(&main, &["init", "-q", "-b", "next"]);
    std::fs::write(main.join("a.txt"), "a\n").unwrap();
    common::git::git_in(&main, &["add", "a.txt"]);
    commit(&main, "base");
    commit(&main, "landing 1");
    commit(&main, "landing 2");
    common::git::git_in(&main, &["branch", "pushed", "HEAD~2"]);
    common::git::git_in(&main, &["branch", "--set-upstream-to=pushed"]);
    write_profile(&main, Some("next"));
    write_forge_counts(&main, 7, 3);
    let stdin = host_stdin(&main, &main, ("modu-ai", "moai-adk"));
    Session {
        home,
        cwd: main,
        stdin,
    }
}

/// A project with no integration branch: `feature/login` is two commits ahead of and one behind
/// its upstream, with the state folder and three more files untracked; the forge counts are off.
fn generic_session() -> Session {
    let home = TempHome::new();
    let main = home.project("demo");
    common::git::git_in(&main, &["init", "-q", "-b", "main"]);
    std::fs::write(main.join("a.txt"), "a\n").unwrap();
    common::git::git_in(&main, &["add", "a.txt"]);
    commit(&main, "base");
    common::git::git_in(&main, &["branch", "up"]);
    common::git::git_in(&main, &["checkout", "-q", "-b", "feature/login"]);
    commit(&main, "login 1");
    commit(&main, "login 2");
    common::git::git_in(&main, &["checkout", "-q", "up"]);
    commit(&main, "upstream 1");
    common::git::git_in(&main, &["checkout", "-q", "feature/login"]);
    common::git::git_in(&main, &["branch", "--set-upstream-to=up"]);
    for name in ["u1.txt", "u2.txt", "u3.txt"] {
        std::fs::write(main.join(name), "u\n").unwrap();
    }
    write_profile(&main, None);
    let config = moai_cli::state_root::path(&main, "config/sections/statusline.yaml");
    std::fs::create_dir_all(config.parent().unwrap()).unwrap();
    std::fs::write(&config, "statusline:\n  segments:\n    github: false\n").unwrap();
    let mut stdin = host_stdin(&main, &main, ("acme", "webapp"));
    stdin["model"] = serde_json::json!({"display_name": "Sonnet 5.5"});
    stdin["effort"] = serde_json::Value::Null;
    stdin["thinking"] = serde_json::Value::Null;
    stdin["pr"] = serde_json::json!({"number": 88, "review_state": "approved"});
    Session {
        home,
        cwd: main,
        stdin,
    }
}

/// `text` without its color escapes.
fn without_escapes(text: &str) -> String {
    let mut out = String::new();
    let mut rest = text;
    while let Some(at) = rest.find("\x1b[") {
        out.push_str(&rest[..at]);
        let tail = &rest[at..];
        rest = &tail[tail.find('m').unwrap() + 1..];
    }
    out.push_str(rest);
    out
}

impl Session {
    /// One render; the exit code must be 0. Returns stdout.
    fn output(&self, env: &[(&str, &str)]) -> String {
        let mut cmd = self
            .home
            .command(&["statusline", "--render-budget-ms", "3600000"]);
        cmd.current_dir(&self.cwd)
            .env("PATH", std::env::var_os("PATH").unwrap_or_default())
            .env("TMPDIR", self.home.path().join("tmp"))
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped());
        std::fs::create_dir_all(self.home.path().join("tmp")).unwrap();
        for (k, v) in env {
            cmd.env(k, v);
        }
        let mut child = cmd.spawn().unwrap();
        child
            .stdin
            .take()
            .unwrap()
            .write_all(self.stdin.to_string().as_bytes())
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

    fn first_line(&self, env: &[(&str, &str)]) -> String {
        let mut env = env.to_vec();
        env.push(("NO_COLOR", "1"));
        self.output(&env).lines().next().unwrap_or("").to_string()
    }

    /// The first line with color, and its text once the escapes are removed.
    fn first_line_colored(&self) -> (String, String) {
        let out = self.output(&[("COLORTERM", "truecolor")]);
        let line = out.lines().next().unwrap_or("").to_string();
        let text = without_escapes(&line);
        (line, text)
    }
}

/// REQ-1, REQ-3, REQ-4, REQ-9: a worktree session whose branch has the worktree's name shows
/// the worktree item alone, with the distance to the integration branch and the pull request.
#[test]
fn line_one_items_worktree_session_with_a_pull_request() {
    let s = worktree_session();
    assert_eq!(s.first_line(&[]), LINE_WORKTREE);
    let (colored, text) = s.first_line_colored();
    assert!(colored.contains('\x1b'), "{colored:?}");
    assert_eq!(text, LINE_WORKTREE, "color changes no text");
}

/// REQ-3, REQ-4: the main checkout shows the branch item alone; on the integration branch `↑`
/// counts the landings not yet pushed and there is no distance item.
#[test]
fn line_one_items_main_checkout() {
    let s = main_session();
    assert_eq!(s.first_line(&[]), LINE_MAIN);
    let (_, text) = s.first_line_colored();
    assert_eq!(text, LINE_MAIN);
}

/// REQ-4: a project with no integration branch has no distance item, and the branch item keeps
/// the numbers against its own upstream.
#[test]
fn line_one_items_generic_project() {
    let s = generic_session();
    assert_eq!(s.first_line(&[]), LINE_GENERIC);
    let (_, text) = s.first_line_colored();
    assert_eq!(text, LINE_GENERIC);
}

/// REQ-9: the host says the request is a merge request (`pr.kind` is `mr`, as the official
/// statusline page defines it), so the item is `!<n>`; any other kind, or a kind of another
/// type, is a pull request and does not fail the input.
#[test]
fn line_one_items_merge_request() {
    let mut s = generic_session();
    s.stdin["pr"] = serde_json::json!({"number": 31, "review_state": "pending", "kind": "mr"});
    let want = "🧠 Sonnet 5.5 | 📡 acme/webapp | 🅱️  feature/login ↑2 ↓1 +4 | 🔗 !31 pending";
    assert_eq!(s.first_line(&[]), want);
    assert_eq!(s.first_line_colored().1, want);
    for kind in [
        serde_json::json!("pull_request"),
        serde_json::json!("merge_request"),
        serde_json::json!("MR"),
        serde_json::json!(7),
        serde_json::json!(null),
    ] {
        s.stdin["pr"] = serde_json::json!({"number": 31, "review_state": "pending", "kind": kind});
        assert_eq!(s.first_line(&[]), want.replace("!31", "#31"), "kind {kind}");
    }
    s.stdin["pr"] = serde_json::Value::Null;
    assert_eq!(
        s.first_line(&[]),
        "🧠 Sonnet 5.5 | 📡 acme/webapp | 🅱️  feature/login ↑2 ↓1 +4",
        "no request: no item and no separator"
    );
}

/// REQ-10: a missing model leaves the item out, a model id alone shows its shortened name, and
/// the effort and thinking fields give `high`, `high·t` or nothing.
#[test]
fn line_one_items_missing_model() {
    let mut s = generic_session();
    let rest = "📡 acme/webapp | 🅱️  feature/login ↑2 ↓1 +4 | 🔗 #88 approved";
    s.stdin["model"] = serde_json::Value::Null;
    assert_eq!(s.first_line(&[]), rest);
    s.stdin["model"] = serde_json::json!({});
    assert_eq!(s.first_line(&[]), rest);
    s.stdin["model"] = serde_json::json!({"id": "claude-opus-4-1-20250805"});
    assert_eq!(s.first_line(&[]), format!("🧠 Opus 4.1 | {rest}"));
    assert_eq!(s.first_line_colored().1, format!("🧠 Opus 4.1 | {rest}"));
    s.stdin["effort"] = serde_json::json!({"level": "high"});
    assert_eq!(s.first_line(&[]), format!("🧠 Opus 4.1 high | {rest}"));
    s.stdin["thinking"] = serde_json::json!({"enabled": true});
    assert_eq!(s.first_line(&[]), format!("🧠 Opus 4.1 high·t | {rest}"));
    s.stdin["thinking"] = serde_json::json!({"enabled": false});
    assert_eq!(s.first_line(&[]), format!("🧠 Opus 4.1 high | {rest}"));
}

/// REQ-1: with nothing to show the first line is left out and the usage line moves up, and the
/// old tag never appears.
#[test]
fn line_one_items_empty_input_has_no_first_line() {
    let home = TempHome::new();
    let cwd = home.path().join("work/plain");
    std::fs::create_dir_all(&cwd).unwrap();
    let s = Session {
        home,
        cwd,
        stdin: serde_json::json!({}),
    };
    let out = s.output(&[("NO_COLOR", "1")]);
    assert!(out.starts_with("CW n/a"), "{out:?}");
    assert!(!out.contains("🏷️"), "{out:?}");
    assert!(!out.contains("🧠"), "{out:?}");
}

/// REQ-11: the widest form that fits `COLUMNS` − 2, in the binary.
#[test]
fn line_one_items_follow_columns() {
    let s = worktree_session();
    for (columns, want) in [
        ("114", LINE_WORKTREE),
        (
            "113",
            "🧠 Opus 5.5 high·t | 📡 modu-ai/moai-adk | 🌳 statusline-flow ↑3 +2 | 🧭 next ↑3 ↓0 | 🔗 #12 pending",
        ),
        (
            "102",
            "🧠 Opus 5.5 high·t | 📡 modu-ai/moai-adk | 🌳 statusline-flow ↑3 +2 | 🧭 next ↑3 ↓0 | 🔗 #12 pending",
        ),
        (
            "101",
            "🧠 Opus high | 📡 moai-adk | 🌳 statusline-flow ↑3 +2 | 🧭 next ↑3 ↓0 | 🔗 #12",
        ),
        (
            "80",
            "🧠 Opus high | 📡 moai-adk | 🌳 statusline-flow ↑3 +2 | 🧭 next ↑3 ↓0 | 🔗 #12",
        ),
        (
            "79",
            "🧠 Opus | 🌳 statusline-flow ↑3 +2 | 🧭 next ↓0 | 🔗 #12",
        ),
        (
            "58",
            "🧠 Opus | 🌳 statusline-flow ↑3 +2 | 🧭 next ↓0 | 🔗 #12",
        ),
        ("57", "🧠 Opus | 🌳 statusline-flow ↑3 +2"),
        ("10", "🧠 Opus | 🌳 statusline-flow ↑3 +2"),
        ("abc", LINE_WORKTREE),
    ] {
        assert_eq!(
            s.first_line(&[("COLUMNS", columns)]),
            want,
            "COLUMNS {columns}"
        );
    }
}

// ---- SPEC-20261122-statusline-limits AC-1 and AC-9: the items and forms of line 3 ----
//
// Each render is the real binary in a temporary HOME and TMPDIR, at the fixed clock of the
// `--now-unix` seam, from a folder that is no repository. The strings are the ones of the Layouts
// section of the SPEC. Two tokens of a line depend on the build: the MoAI version of this build
// stands where the section writes `4.2647.16`, and the update case uses a template of an earlier
// week (`4.2600.15`, older than any build of this head) where the section writes `4.2647.15`;
// [`as_in_the_layouts`] maps them back, and the widths are measured on the line the binary wrote.
//
// The process-count half of AC-9 (a fake `git` and a fake Orca script, at most one process each,
// only the cache and the record written) belongs to the Orca pass-through step and has no test
// here yet.

/// The render clock of line 3 (`--now-unix` seam).
const LIMITS_NOW: i64 = 1_738_429_000;
/// The template version the update case writes into the project's `system.yaml`.
const OLD_TEMPLATE: &str = "4.2600.15";

/// The JSON a subscriber's host would write: the context window at 74 percent, the five-hour
/// window at `five_hour` percent resetting in 1h 10m, the week at 21 percent resetting in 6d 21h,
/// a cost and the Claude Code version.
fn subscriber(five_hour: i64) -> serde_json::Value {
    serde_json::json!({
        "model": {"display_name": "Opus"},
        "version": "2.1.330",
        "context_window": {"used_percentage": 74, "context_window_size": 200000},
        "cost": {"total_cost_usd": 1.23},
        "rate_limits": {
            "five_hour": {"used_percentage": five_hour, "resets_at": LIMITS_NOW + 3_600 + 600},
            "seven_day": {"used_percentage": 21, "resets_at": LIMITS_NOW + 6 * 86_400 + 21 * 3_600},
        },
    })
}

/// The JSON of an API key user: no `rate_limits`, the context window at 41 percent and a cost.
fn api_key_user() -> serde_json::Value {
    serde_json::json!({
        "model": {"display_name": "Opus"},
        "version": "2.1.330",
        "context_window": {"used_percentage": 41, "context_window_size": 200000},
        "cost": {"total_cost_usd": 3.87},
    })
}

/// Line 3 of one render of `stdin` (the last line), with `env` added; with `template`, the
/// project's `system.yaml` names that template version. The exit code must be 0.
fn line_three(stdin: &serde_json::Value, env: &[(&str, &str)], template: Option<&str>) -> String {
    let home = TempHome::new();
    let cwd = home.path().join("work/plain");
    std::fs::create_dir_all(&cwd).unwrap();
    if let Some(template) = template {
        let sections = cwd
            .join(moai_cli::state_root::MOAI_DIR)
            .join("config/sections");
        std::fs::create_dir_all(&sections).unwrap();
        std::fs::write(
            sections.join("system.yaml"),
            format!("moai:\n  template_version: {template}\n"),
        )
        .unwrap();
    }
    let tmp = home.path().join("tmp");
    std::fs::create_dir_all(&tmp).unwrap();
    let now = LIMITS_NOW.to_string();
    let mut cmd = home.command(&[
        "statusline",
        "--render-budget-ms",
        "3600000",
        "--now-unix",
        &now,
    ]);
    cmd.current_dir(&cwd)
        .env("PATH", std::env::var_os("PATH").unwrap_or_default())
        .env("TMPDIR", &tmp)
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
        .write_all(stdin.to_string().as_bytes())
        .unwrap();
    let out = child.wait_with_output().unwrap();
    assert_eq!(
        out.status.code(),
        Some(0),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    String::from_utf8(out.stdout)
        .unwrap()
        .lines()
        .last()
        .unwrap_or_default()
        .to_string()
}

/// The plain line 3 of `stdin`.
fn plain_three(stdin: &serde_json::Value, template: Option<&str>) -> String {
    line_three(stdin, &[("NO_COLOR", "1")], template)
}

/// The plain line 3 of `stdin` at `columns`.
fn plain_three_at(stdin: &serde_json::Value, template: Option<&str>, columns: &str) -> String {
    line_three(stdin, &[("NO_COLOR", "1"), ("COLUMNS", columns)], template)
}

/// The build's tokens of a line written as the Layouts section writes them.
fn as_in_the_layouts(line: &str) -> String {
    let build = moai_cli::BUILD_VERSION;
    line.replace(
        &format!("🗿 {OLD_TEMPLATE} -> {build}"),
        "🗿 4.2647.15 -> 4.2647.16",
    )
    .replace(&format!("🗿 {build}"), "🗿 4.2647.16")
}

/// A line of the Layouts section with the build's tokens: what the binary writes.
fn as_the_build_writes(line: &str) -> String {
    let build = moai_cli::BUILD_VERSION;
    line.replace(
        "🗿 4.2647.15 -> 4.2647.16",
        &format!("🗿 {OLD_TEMPLATE} -> {build}"),
    )
    .replace("🗿 4.2647.16", &format!("🗿 {build}"))
}

/// The case `stdin` gives `want` (a Layouts string) in plain output, and the same text with color
/// once the escapes are removed.
fn assert_line_three(stdin: &serde_json::Value, template: Option<&str>, want: &str) {
    let plain = plain_three(stdin, template);
    assert_eq!(as_in_the_layouts(&plain), want, "plain");
    let colored = line_three(stdin, &[("COLORTERM", "truecolor")], template);
    assert!(colored.contains('\x1b'), "{colored:?}");
    assert_eq!(without_escapes(&colored), plain, "color changes no text");
}

/// Z0 of the subscriber whose five-hour window is over 70 percent.
const Z0_FIVE_HOUR: &str =
    "🪫 CW 74% | 🪫 5H 85% (1h 10m) | 🔋 7D 21% (6d 21h) | 🤖 2.1.330 | 🗿 4.2647.16";
/// Z0 of the subscriber whose five-hour window is hidden.
const Z0_NO_FIVE_HOUR: &str = "🪫 CW 74% | 🔋 7D 21% (6d 21h) | 🤖 2.1.330 | 🗿 4.2647.16";
/// Z0 of the API key user. The Layouts section writes `🪫 CW 41%`; the battery rule of the same
/// SPEC (AC-2: `🔋` up to 70 percent, `🪫` from 71) gives `🔋` at 41, so the icon here follows the
/// rule and the widths are the section's.
const Z0_API_KEY: &str = "🔋 CW 41% | 💰 3.87 | 🤖 2.1.330 | 🗿 4.2647.16";

/// REQ-1, REQ-4: a subscriber whose five-hour window is over 70 percent sees every window with
/// its countdown, then the two versions, and no cost.
#[test]
fn line_three_items_subscriber_with_the_five_hour_window() {
    assert_line_three(&subscriber(85), None, Z0_FIVE_HOUR);
}

/// REQ-4, REQ-5: at 70 percent or below the five-hour window is left out with its separator, as
/// when the host sends none; the week alone still hides the cost.
#[test]
fn line_three_items_subscriber_without_the_five_hour_window() {
    assert_line_three(&subscriber(70), None, Z0_NO_FIVE_HOUR);
    assert_line_three(&subscriber(23), None, Z0_NO_FIVE_HOUR);
    let mut no_five = subscriber(85);
    no_five["rate_limits"]
        .as_object_mut()
        .unwrap()
        .remove("five_hour");
    assert_line_three(&no_five, None, Z0_NO_FIVE_HOUR);
}

/// REQ-5: with no usage window the cost stands after the context window, to two decimals.
#[test]
fn line_three_items_api_key_user() {
    assert_line_three(&api_key_user(), None, Z0_API_KEY);
}

/// REQ-2: an empty context window says `CW n/a` with no battery icon, never `0%`; the other items
/// keep their places.
#[test]
fn line_three_items_empty_context_window() {
    let mut stdin = api_key_user();
    stdin["context_window"] = serde_json::json!({});
    assert_line_three(&stdin, None, "CW n/a | 💰 3.87 | 🤖 2.1.330 | 🗿 4.2647.16");
    stdin.as_object_mut().unwrap().remove("context_window");
    assert_line_three(&stdin, None, "CW n/a | 💰 3.87 | 🤖 2.1.330 | 🗿 4.2647.16");
}

/// REQ-2 and D-1: a null percentage falls back to the input tokens of `current_usage` over the
/// window size, rescaled by nothing (60,000 of 200,000 is 30 percent, not the 35 the 85 percent
/// budget gave); without `current_usage` it says `CW n/a`.
#[test]
fn line_three_items_null_percentage() {
    let mut stdin = api_key_user();
    stdin["context_window"] = serde_json::json!({
        "used_percentage": null,
        "context_window_size": 200000,
        "current_usage": {
            "input_tokens": 30000,
            "output_tokens": 9999,
            "cache_creation_input_tokens": 10000,
            "cache_read_input_tokens": 20000,
        },
    });
    assert_line_three(
        &stdin,
        None,
        "🔋 CW 30% | 💰 3.87 | 🤖 2.1.330 | 🗿 4.2647.16",
    );
    stdin["context_window"] =
        serde_json::json!({"used_percentage": null, "context_window_size": 200000});
    assert_line_three(&stdin, None, "CW n/a | 💰 3.87 | 🤖 2.1.330 | 🗿 4.2647.16");
}

/// REQ-2: `CW` is the host's percentage unchanged — 25, 85 and 100 stay 25, 85 and 100 (the 85
/// percent rescale wrote 29, 100 and 100).
#[test]
fn line_three_items_context_window_as_the_host_gives_it() {
    for (pct, want) in [(25, "🔋 CW 25%"), (85, "🪫 CW 85%"), (100, "🪫 CW 100%")] {
        let mut stdin = api_key_user();
        stdin["context_window"]["used_percentage"] = serde_json::json!(pct);
        let line = plain_three(&stdin, None);
        assert_eq!(line.split(" | ").next(), Some(want), "{pct}: {line}");
    }
}

/// Every icon of line 3 is followed by one space and then a letter or a number.
#[test]
fn line_three_items_every_icon_is_followed_by_a_space() {
    let line = plain_three(&subscriber(85), None) + " | " + &plain_three(&api_key_user(), None);
    let mut seen = 0;
    for icon in ["🪫", "🔋", "💰", "🤖", "🗿"] {
        for (at, _) in line.match_indices(icon) {
            seen += 1;
            let after = &line[at + icon.len()..];
            assert!(after.starts_with(' '), "{icon}: `{after}`");
            assert!(
                after[1..].chars().next().is_some_and(char::is_alphanumeric),
                "{icon}: `{after}`"
            );
        }
    }
    assert_eq!(seen, 9, "{line}");
}

/// The forms of each Layouts case from the widest to the narrowest, with the width the section
/// gives. The section lists the forms by their number; a form it does not list for a case (Z1 of
/// the subscriber and of the API key user) follows the drop order under the table: the Claude Code
/// version goes first.
const LADDERS: [(&str, &[(&str, usize)]); 4] = [
    (
        "subscriber, 5H over 70",
        &[
            (Z0_FIVE_HOUR, 79),
            (
                "🪫 CW 74% | 🪫 5H 85% (1h 10m) | 🔋 7D 21% (6d 21h) | 🗿 4.2647.16",
                66,
            ),
            ("🪫 CW 74% | 🪫 5H 85% (1h 10m) | 🔋 7D 21% (6d 21h)", 51),
            ("🪫 CW 74% | 🪫 5H 85% | 🔋 7D 21%", 33),
            ("🪫 CW 74% | 🔋 7D 21%", 21),
        ],
    ),
    (
        "subscriber, 5H hidden",
        &[
            (Z0_NO_FIVE_HOUR, 58),
            ("🪫 CW 74% | 🔋 7D 21% (6d 21h) | 🗿 4.2647.16", 45),
            ("🪫 CW 74% | 🔋 7D 21% (6d 21h)", 30),
            ("🪫 CW 74% | 🔋 7D 21%", 21),
        ],
    ),
    (
        "update pending",
        &[
            (
                "🪫 CW 74% | 🪫 5H 85% (1h 10m) | 🔋 7D 21% (6d 21h) | 🤖 2.1.330 | 🗿 4.2647.15 -> 4.2647.16",
                92,
            ),
            (
                "🪫 CW 74% | 🪫 5H 85% (1h 10m) | 🔋 7D 21% (6d 21h) | 🗿 4.2647.15 -> 4.2647.16",
                79,
            ),
            ("🪫 CW 74% | 🪫 5H 85% (1h 10m) | 🔋 7D 21% (6d 21h)", 51),
            ("🪫 CW 74% | 🪫 5H 85% | 🔋 7D 21%", 33),
            ("🪫 CW 74% | 🔋 7D 21%", 21),
        ],
    ),
    (
        "API key user",
        &[
            (Z0_API_KEY, 47),
            ("🔋 CW 41% | 💰 3.87 | 🗿 4.2647.16", 34),
            ("🔋 CW 41% | 💰 3.87", 19),
            ("🔋 CW 41%", 9),
        ],
    ),
];

/// The stdin and the template of a Layouts case.
fn ladder_case(name: &str) -> (serde_json::Value, Option<&'static str>) {
    match name {
        "subscriber, 5H over 70" => (subscriber(85), None),
        "subscriber, 5H hidden" => (subscriber(23), None),
        "update pending" => (subscriber(85), Some(OLD_TEMPLATE)),
        _ => (api_key_user(), None),
    }
}

/// AC-9 (REQ-1): every form is exactly the string of the Layouts section, with the section's
/// width, and changes at its boundary: `COLUMNS` of the form's width plus the margin of two keeps
/// it, one less drops to the next form.
#[test]
fn line_three_ladder_every_form_is_the_string_of_the_layouts_section() {
    for (name, forms) in LADDERS {
        let (stdin, template) = ladder_case(name);
        for (i, (text, cols)) in forms.iter().enumerate() {
            assert_eq!(display_width(text), *cols, "{name}: the width of `{text}`");
            let written = display_width(&as_the_build_writes(text));
            let kept = plain_three_at(&stdin, template, &(written + 2).to_string());
            assert_eq!(
                &as_in_the_layouts(&kept),
                text,
                "{name}: COLUMNS {}",
                written + 2
            );
            if let Some((next, _)) = forms.get(i + 1) {
                let dropped = plain_three_at(&stdin, template, &(written + 1).to_string());
                assert_eq!(
                    &as_in_the_layouts(&dropped),
                    next,
                    "{name}: COLUMNS {}",
                    written + 1
                );
            }
        }
    }
}

/// REQ-1 and the width rule of the parent SPEC: no usable `COLUMNS` writes the widest form, and a
/// `COLUMNS` below the narrowest form still writes the narrowest.
#[test]
fn line_three_ladder_without_a_usable_columns_and_below_the_narrowest() {
    let (stdin, _) = ladder_case("subscriber, 5H over 70");
    for columns in ["", "0", "abc", "8 0"] {
        let line = plain_three_at(&stdin, None, columns);
        assert_eq!(
            as_in_the_layouts(&line),
            Z0_FIVE_HOUR,
            "COLUMNS {columns:?}"
        );
    }
    for columns in ["22", "10", "2", "1"] {
        let line = plain_three_at(&stdin, None, columns);
        assert_eq!(line, "🪫 CW 74% | 🔋 7D 21%", "COLUMNS {columns}");
    }
    let line = plain_three_at(&api_key_user(), None, "3");
    assert_eq!(line, "🔋 CW 41%");
}

/// AC-9 (REQ-12): one render with `rate_limits`, the Orca hook variables and the Orca wrapper
/// in place starts at most one `git` and exactly one wrapper, and writes nothing but the session
/// cache and the usage record (both in the temporary folder); the HOME is left as it was.
#[test]
fn line_three_ladder_one_render_starts_one_git_and_one_orca_wrapper_at_most() {
    let f = fixture();
    let outside = tempfile::tempdir().unwrap();
    let log = outside.path().join("spawns.log");
    let bin = outside.path().join("bin");
    std::fs::create_dir_all(&bin).unwrap();
    let real_git = String::from_utf8(
        std::process::Command::new("sh")
            .args(["-c", "command -v git"])
            .output()
            .unwrap()
            .stdout,
    )
    .unwrap();
    common::write_executable(
        &bin.join("git"),
        &format!(
            "#!/bin/sh\necho git >> '{}'\nexec '{}' \"$@\"\n",
            log.display(),
            real_git.trim()
        ),
    );
    let wrapper = f.home.path().join(".orca/agent-hooks/claude-statusline.sh");
    std::fs::create_dir_all(wrapper.parent().unwrap()).unwrap();
    common::write_executable(
        &wrapper,
        &format!(
            "#!/bin/sh\n/bin/cat > /dev/null\necho orca >> '{}'\n",
            log.display()
        ),
    );
    let mut stdin = subscriber(85);
    stdin["session_id"] = serde_json::json!("s-ac9");
    // The usage record keeps a window by the wall clock (not the `--now-unix` seam), so its
    // resets lie an hour after the real time.
    let soon = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .unwrap()
        .as_secs() as i64
        + 3_600;
    for w in ["five_hour", "seven_day"] {
        stdin["rate_limits"][w]["resets_at"] = serde_json::json!(soon);
    }
    stdin["cwd"] = serde_json::json!(f.project);
    stdin["workspace"] = serde_json::json!({
        "current_dir": f.project,
        "project_dir": f.project,
        "repo": {"host": "github.com", "owner": "modu-ai", "name": "moai-adk"},
    });
    let tmp = tempfile::tempdir().unwrap();
    let before = f.home.snapshot();
    let mut cmd = f
        .home
        .command(&["statusline", "--render-budget-ms", "3600000"]);
    cmd.current_dir(&f.project)
        .env("PATH", &bin)
        .env("TMPDIR", tmp.path())
        .env("NO_COLOR", "1")
        .env("ORCA_AGENT_HOOK_PORT", "1")
        .env("ORCA_AGENT_HOOK_TOKEN", "fake")
        .env("ORCA_PANE_KEY", "tab:1")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::null());
    let mut child = cmd.spawn().unwrap();
    child
        .stdin
        .take()
        .unwrap()
        .write_all(stdin.to_string().as_bytes())
        .unwrap();
    let out = child.wait_with_output().unwrap();
    assert_eq!(out.status.code(), Some(0));
    let count = |what: &str| {
        std::fs::read_to_string(&log)
            .unwrap_or_default()
            .lines()
            .filter(|l| *l == what)
            .count()
    };
    let deadline = std::time::Instant::now() + std::time::Duration::from_secs(30);
    while count("orca") == 0 && std::time::Instant::now() < deadline {
        std::thread::sleep(std::time::Duration::from_millis(50));
    }
    std::thread::sleep(std::time::Duration::from_millis(300));
    assert_eq!(count("orca"), 1, "the wrapper starts once");
    assert!(count("git") <= 1, "git started {} times", count("git"));
    assert_eq!(f.home.snapshot(), before, "the render changed the HOME");
    let written: Vec<String> = common::cli::tree_state(tmp.path())
        .into_iter()
        .map(|(rel, _)| rel)
        .filter(|rel| !tmp.path().join(rel).is_dir())
        .collect();
    assert_eq!(written.len(), 2, "{written:?}");
    assert_eq!(written[0], "moai-statusline/s-ac9.json", "{written:?}");
    assert!(
        written[1].starts_with("moai-statusline/usage-") && written[1].ends_with(".record"),
        "{written:?}"
    );
}
