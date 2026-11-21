//! User instruction 2026-11-19: the status line opens line 3 (the bars line) with the worktree,
//! written `🌳 WT-c<card>-<name>`: `<card>` is the number of the card whose held reservation names
//! the session's worktree folder, `<name>` comes from that reservation's branch
//! (`spec/20260101-demo-work` gives `demo-work`). Without such a reservation the segment is the
//! worktree name the host gave, alone, at the same place. The main checkout has no segment.
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
        let mut cmd = self
            .home
            .command(&["statusline", "--render-budget-ms", "3600000"]);
        cmd.current_dir(&dir)
            .env("PATH", std::env::var_os("PATH").unwrap_or_default())
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

/// The whole of line 3 after the label moved to line 2: the bars, unchanged by the label.
const BARS: &str = "CTX: ░░░░░░░░░░ n/a";

#[test]
fn a_held_card_closes_line_2_as_wt_c_card_name() {
    let f = fixture();
    let wt = f.worktree("demo-1-work", "spec/20260101-demo-work");
    let card = f.reserve("active", "held", &wt.dir, "spec/20260101-demo-work");
    assert_eq!(card, 1);
    let out = f.plain(Some(&wt));
    // The picked layout of 2026-11-20T21:5x: the label closes line 2, line 3 keeps the
    // counts and the bars.
    assert!(line(&out, 1).ends_with(" | 🌳 WT-c1-demo-work"), "{out}");
    assert_eq!(
        line(&out, 2),
        format!("🗂️ 0 ▶️ 1 🔃 0 ✅ 0 | {BARS}"),
        "{out}"
    );
    assert_eq!(out.lines().count(), 3, "{out}");
}

#[test]
fn only_the_segment_moves_line_1_and_the_bars_stay() {
    let f = fixture();
    let wt = f.worktree("demo-1-work", "spec/20260101-demo-work");
    let without = f.plain(Some(&wt));
    f.reserve("active", "held", &wt.dir, "spec/20260101-demo-work");
    let with = f.plain(Some(&wt));
    // The card counts open line 3 (a registered ledger shows the group even at zero) and
    // the bars close it; the worktree label closes line 2 in both renders.
    assert!(
        line(&with, 2).starts_with("🗂️ 0 ▶️ 1 🔃 0 ✅ 0 | "),
        "{with}"
    );
    assert!(
        line(&without, 2).starts_with("🗂️ 0 ▶️ 0 🔃 0 ✅ 0 | "),
        "{without}"
    );
    assert!(line(&with, 1).ends_with(" | 🌳 WT-c1-demo-work"), "{with}");
    // Without the reservation the label is the host's name alone; both line 2 shapes keep
    // it last.
    assert!(
        line(&without, 1).ends_with(" | 🌳 demo-1-work"),
        "{without}"
    );
    assert!(line(&with, 2).ends_with(BARS), "{with}");
    // Line 1 holds no worktree and no card counts in either render.
    assert_eq!(line(&with, 0), line(&without, 0));
    assert!(!line(&with, 0).contains('🌳'), "{with}");
    assert!(!line(&without, 0).contains('🌳'), "{without}");
}

#[test]
fn no_reservation_shows_the_host_name_alone_on_line_2() {
    let f = fixture();
    let wt = f.worktree("demo-1-work", "spec/20260101-demo-work");
    let out = f.plain(Some(&wt));
    assert!(line(&out, 1).ends_with(" | 🌳 demo-1-work"), "{out}");
}

#[test]
fn a_released_reservation_is_not_a_holder() {
    let f = fixture();
    let wt = f.worktree("demo-1-work", "spec/20260101-demo-work");
    f.reserve("ready", "released", &wt.dir, "spec/20260101-demo-work");
    let out = f.plain(Some(&wt));
    assert!(line(&out, 1).ends_with(" | 🌳 demo-1-work"), "{out}");
}

#[test]
fn two_cards_holding_one_folder_show_the_lowest_number() {
    let f = fixture();
    let wt = f.worktree("demo-1-work", "spec/20260101-demo-work");
    let first = f.reserve("active", "held", &wt.dir, "spec/20260101-demo-work");
    let second = f.reserve("active", "held", &wt.dir, "spec/20260102-other-work");
    assert!(first < second);
    let out = f.plain(Some(&wt));
    assert!(line(&out, 1).ends_with(" | 🌳 WT-c1-demo-work"), "{out}");
}

#[test]
fn a_branch_of_another_shape_takes_the_rest_of_the_folder_name() {
    let f = fixture();
    let wt = f.worktree("demo-1-hooks", "m-flow/bundle-20260101-6");
    f.reserve("active", "held", &wt.dir, "m-flow/bundle-20260101-6");
    let out = f.plain(Some(&wt));
    assert!(line(&out, 1).ends_with(" | 🌳 WT-c1-hooks"), "{out}");
}

#[test]
fn a_folder_that_keeps_no_rest_gives_the_card_alone() {
    let f = fixture();
    let wt = f.worktree("demo-1", "work-branch");
    f.reserve("active", "held", &wt.dir, "work-branch");
    let out = f.plain(Some(&wt));
    assert!(line(&out, 1).ends_with(" | 🌳 WT-c1"), "{out}");
}

#[test]
fn the_main_checkout_has_no_worktree_segment() {
    let f = fixture();
    let out = f.plain(None);
    assert_eq!(
        line(&out, 2),
        format!("🗂️ 0 ▶️ 0 🔃 0 ✅ 0 | {BARS}"),
        "{out}"
    );
    assert!(!out.contains('🌳'), "{out}");
}

#[test]
fn the_worktree_option_off_hides_the_segment_on_every_line() {
    let f = fixture();
    let wt = f.worktree("demo-1-work", "spec/20260101-demo-work");
    f.reserve("active", "held", &wt.dir, "spec/20260101-demo-work");
    let config = f.project.join(".moai-cli/config/sections/statusline.yaml");
    std::fs::create_dir_all(config.parent().unwrap()).unwrap();
    std::fs::write(&config, "statusline:\n  segments:\n    worktree: false\n").unwrap();
    let out = f.plain(Some(&wt));
    assert_eq!(
        line(&out, 2),
        format!("🗂️ 0 ▶️ 1 🔃 0 ✅ 0 | {BARS}"),
        "{out}"
    );
    assert!(!out.contains('🌳'), "{out}");
}
