//! The `🌳 <worktree>` item of the first line (SPEC-20261122-statusline-where, D-4). It shows the
//! worktree name the host gave, as it is, before the branch item; the card number is not part of
//! it any more (the label `WT-c<card>-<name>` of the user instruction of 2026-11-19 is gone, and
//! the held card shows as `🎯` on the flow line). The main checkout has no item.
//!
//! Every case runs the real binary through `TempHome` (a temporary HOME, nothing of the real
//! environment but a PATH for git) against a ledger built in that HOME and a real linked
//! worktree made with `git worktree add`. The read is local and read-only: the HOME tree is
//! compared entry by entry, bytes and modification times, before and after each render.

mod common;

use common::home::TempHome;
use moai_cli::ledger;
use std::io::Write;
use std::path::{Path, PathBuf};
use std::process::Stdio;

struct Fixture {
    home: TempHome,
    project: PathBuf,
}

/// A HOME with the project `demo` (profile, a git repository on `main` with one commit) and a
/// registered, empty ledger.
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

/// A linked worktree of the project: its folder (canonical, as a holder records it) and the
/// name the host would give it (the folder name).
struct Worktree {
    dir: PathBuf,
    name: String,
}

impl Fixture {
    /// `git worktree add -b <branch> <home>/wt/<folder>`.
    fn worktree(&self, folder: &str, branch: &str) -> Worktree {
        let dir = self.home.path().join("wt").join(folder);
        std::fs::create_dir_all(dir.parent().unwrap()).unwrap();
        common::git::git_in(
            &self.project,
            &["worktree", "add", "-q", "-b", branch, dir.to_str().unwrap()],
        );
        Worktree {
            dir: dir.canonicalize().unwrap(),
            name: folder.to_string(),
        }
    }

    /// One card in `card_state` and one reservation of it in `status` naming `worktree` and
    /// `branch`. Returns the card number. An `INSERT` has no transition check.
    fn reserve(&self, card_state: &str, status: &str, worktree: &Path, branch: &str) -> i64 {
        let mut l = ledger::open_write(self.home.path(), "demo").unwrap();
        let mut card = 0;
        l.with_write_tx(|tx, now| {
            tx.execute(
                "INSERT INTO card (title, state, created_by, created_at, updated_at) \
                 VALUES ('t', ?1, 'test', ?2, ?2)",
                rusqlite::params![card_state, now],
            )?;
            card = tx.last_insert_rowid();
            tx.execute(
                "INSERT INTO reservation (card_id, generation, holder, status, acquired_at, \
                 lease_until, updated_at, worktree, branch) VALUES (?1, 1, 'test', ?2, ?3, ?4, ?3, ?5, ?6)",
                rusqlite::params![
                    card,
                    status,
                    now,
                    now + 14400,
                    worktree.to_string_lossy(),
                    branch
                ],
            )?;
            Ok(())
        })
        .unwrap();
        card
    }

    /// One render as Claude Code runs it from `wt` (or from the main checkout), NO_COLOR, with
    /// the HOME and the project left exactly as they were. Returns stdout.
    fn plain(&self, wt: Option<&Worktree>) -> String {
        let before = self.home.snapshot();
        let out = self.run(wt);
        assert_eq!(
            self.home.snapshot(),
            before,
            "the render changed the HOME or the project"
        );
        out
    }

    fn run(&self, wt: Option<&Worktree>) -> String {
        let dir = wt.map_or(self.project.clone(), |w| w.dir.clone());
        let mut stdin = serde_json::json!({
            "cwd": dir,
            "model": {"display_name": "Opus"},
            "workspace": {"current_dir": dir, "project_dir": self.project},
        });
        if let Some(w) = wt {
            stdin["worktree"] = serde_json::json!({"name": w.name});
            stdin["workspace"]["git_worktree"] = serde_json::json!(w.name);
        }
        // The render may keep a session cache under `$TMPDIR`: a folder of its own outside the
        // HOME, so the HOME snapshot stays equal and nothing reaches the real temporary folder.
        let tmp = tempfile::tempdir().unwrap();
        let mut cmd = self
            .home
            .command(&["statusline", "--render-budget-ms", "3600000"]);
        cmd.current_dir(&dir)
            .env("PATH", std::env::var_os("PATH").unwrap_or_default())
            .env("TMPDIR", tmp.path())
            .env("NO_COLOR", "1")
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped());
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
        String::from_utf8(out.stdout).unwrap()
    }
}

fn line(out: &str, n: usize) -> String {
    out.lines().nth(n).unwrap_or_default().to_string()
}

/// The usage windows line: it starts with the context window and, in this fixture, ends with
/// the MoAI version.
fn is_windows(line: &str) -> bool {
    line.starts_with("CW n/a")
}

/// The flow line when the one card of the ledger is held by a live lease in the session's
/// worktree: it is my card, working, and the only card of the flow.
const FLOW_HELD: &str = "🎯 c1 W | 🗂️  D0 R0 W1 E0 ✅ 0";

/// The first line of the worktree session: the host's name, then the branch of the worktree
/// (the names differ here, so both items stand).
const LINE_WORKTREE: &str = "🧠 Opus | 🌳 demo-1-work | 🅱️  spec/20260101-demo-work";

#[test]
fn a_held_card_does_not_name_the_worktree_item() {
    let f = fixture();
    let wt = f.worktree("demo-1-work", "spec/20260101-demo-work");
    let card = f.reserve("active", "held", &wt.dir, "spec/20260101-demo-work");
    assert_eq!(card, 1);
    let out = f.plain(Some(&wt));
    // The worktree item is the host's name; the card shows on the flow line, which stands under
    // line 1, and the usage windows close the output.
    assert_eq!(line(&out, 0), LINE_WORKTREE, "{out}");
    assert!(!out.contains("WT-c"), "{out}");
    assert_eq!(line(&out, 1), FLOW_HELD, "{out}");
    assert!(is_windows(&line(&out, 2)), "{out}");
    assert_eq!(out.lines().count(), 3, "{out}");
}

#[test]
fn only_the_flow_line_moves_when_a_card_is_held() {
    let f = fixture();
    let wt = f.worktree("demo-1-work", "spec/20260101-demo-work");
    let without = f.plain(Some(&wt));
    f.reserve("active", "held", &wt.dir, "spec/20260101-demo-work");
    let with = f.plain(Some(&wt));
    // The flow line stands under line 1 once a card exists (a ledger with no card has nothing to
    // show) and the usage windows close the output; line 1 is the same bytes in both renders.
    assert_eq!(line(&without, 0), LINE_WORKTREE, "{without}");
    assert_eq!(line(&with, 0), line(&without, 0));
    assert_eq!(line(&with, 1), FLOW_HELD, "{with}");
    assert!(
        is_windows(&line(&without, 1)),
        "no card: no flow line: {without}"
    );
    assert_eq!(line(&with, 2), line(&without, 1), "{with}");
    assert_eq!(without.lines().count(), 2, "{without}");
    assert_eq!(with.lines().count(), 3, "{with}");
}

#[test]
fn no_reservation_shows_the_host_name() {
    let f = fixture();
    let wt = f.worktree("demo-1-work", "spec/20260101-demo-work");
    let out = f.plain(Some(&wt));
    assert!(line(&out, 0).contains(" | 🌳 demo-1-work | "), "{out}");
}

#[test]
fn a_released_reservation_is_not_a_holder() {
    let f = fixture();
    let wt = f.worktree("demo-1-work", "spec/20260101-demo-work");
    f.reserve("ready", "released", &wt.dir, "spec/20260101-demo-work");
    let out = f.plain(Some(&wt));
    assert_eq!(line(&out, 0), LINE_WORKTREE, "{out}");
    assert!(!out.contains('🎯'), "{out}");
}

/// The branch and the worktree have one name: the worktree item stands alone.
#[test]
fn a_branch_named_like_the_worktree_shows_the_worktree_item_alone() {
    let f = fixture();
    let wt = f.worktree("demo-work", "demo-work");
    let out = f.plain(Some(&wt));
    assert_eq!(line(&out, 0), "🧠 Opus | 🌳 demo-work", "{out}");
}

#[test]
fn the_main_checkout_has_no_worktree_item() {
    let f = fixture();
    let out = f.plain(None);
    // `+1`: the project's untracked state folder.
    assert_eq!(line(&out, 0), "🧠 Opus | 🅱️  main +1", "{out}");
    assert!(
        is_windows(&line(&out, 1)),
        "an empty ledger shows no flow line: {out}"
    );
    assert!(!out.contains('🌳'), "{out}");
}

#[test]
fn the_worktree_option_off_hides_the_item_on_every_line() {
    let f = fixture();
    let wt = f.worktree("demo-1-work", "spec/20260101-demo-work");
    f.reserve("active", "held", &wt.dir, "spec/20260101-demo-work");
    let config = f.project.join(".moai-cli/config/sections/statusline.yaml");
    std::fs::create_dir_all(config.parent().unwrap()).unwrap();
    std::fs::write(&config, "statusline:\n  segments:\n    worktree: false\n").unwrap();
    let out = f.plain(Some(&wt));
    assert_eq!(
        line(&out, 0),
        "🧠 Opus | 🅱️  spec/20260101-demo-work",
        "{out}"
    );
    assert_eq!(line(&out, 1), FLOW_HELD, "{out}");
    assert!(is_windows(&line(&out, 2)), "{out}");
    assert!(!out.contains('🌳'), "{out}");
}
