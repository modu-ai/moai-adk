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
/// `COLUMNS` − 2 − padding. The windows are the same bytes at every width; the flow line changes
/// form exactly at each boundary, and the first line has its own forms (AC-6 of
/// SPEC-20261122-statusline-where).
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
        // The windows line is the same bytes at every width.
        for out in [&kept, &at(w + 1)] {
            assert_eq!(out.lines().last(), wide.lines().last(), "the windows line");
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
        ("113", "🧠 Opus 5.5 high·t | 📡 modu-ai/moai-adk | 🌳 statusline-flow ↑3 +2 | 🧭 next ↑3 ↓0 | 🔗 #12 pending"),
        ("102", "🧠 Opus 5.5 high·t | 📡 modu-ai/moai-adk | 🌳 statusline-flow ↑3 +2 | 🧭 next ↑3 ↓0 | 🔗 #12 pending"),
        ("101", "🧠 Opus high | 📡 moai-adk | 🌳 statusline-flow ↑3 +2 | 🧭 next ↑3 ↓0 | 🔗 #12"),
        ("80", "🧠 Opus high | 📡 moai-adk | 🌳 statusline-flow ↑3 +2 | 🧭 next ↑3 ↓0 | 🔗 #12"),
        ("79", "🧠 Opus | 🌳 statusline-flow ↑3 +2 | 🧭 next ↓0 | 🔗 #12"),
        ("58", "🧠 Opus | 🌳 statusline-flow ↑3 +2 | 🧭 next ↓0 | 🔗 #12"),
        ("57", "🧠 Opus | 🌳 statusline-flow ↑3 +2"),
        ("10", "🧠 Opus | 🌳 statusline-flow ↑3 +2"),
        ("abc", LINE_WORKTREE),
    ] {
        assert_eq!(s.first_line(&[("COLUMNS", columns)]), want, "COLUMNS {columns}");
    }
}
