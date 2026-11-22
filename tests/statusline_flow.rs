//! SPEC-20261122-statusline-flow, the part of the flow line only the real binary can show: which
//! card is mine (`🎯 c<n> <stage letter>`), what each line of the output is for, and the spacing
//! after every icon.
//!
//! Every case runs the real binary through `TempHome` against a ledger built in that HOME and a
//! real linked worktree made with `git worktree add`. The read is local and read-only: the HOME
//! tree is compared entry by entry, bytes and modification times, before and after each render.

mod common;

use common::home::TempHome;
use moai_cli::ledger;
use moai_cli::statusline::width::ICONS;
use std::io::Write;
use std::path::{Path, PathBuf};
use std::process::Stdio;

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

/// One reservation to write: the card state it belongs to and what it holds.
struct Hold<'a> {
    card_state: &'a str,
    status: &'a str,
    holder: &'a str,
    worktree: &'a Path,
    /// Seconds from the ledger's clock to the end of the lease; negative = already over.
    lease: i64,
}

impl Fixture {
    fn worktree(&self, folder: &str, branch: &str) -> PathBuf {
        let dir = self.home.path().join("wt").join(folder);
        std::fs::create_dir_all(dir.parent().unwrap()).unwrap();
        common::git::git_in(
            &self.project,
            &["worktree", "add", "-q", "-b", branch, dir.to_str().unwrap()],
        );
        dir.canonicalize().unwrap()
    }

    /// One card and one reservation of it. Returns the card number.
    fn hold(&self, h: &Hold) -> i64 {
        let mut l = ledger::open_write(self.home.path(), "demo").unwrap();
        let mut card = 0;
        l.with_write_tx(|tx, now| {
            tx.execute(
                "INSERT INTO card (title, state, created_by, created_at, updated_at) \
                 VALUES ('t', ?1, 'test', ?2, ?2)",
                rusqlite::params![h.card_state, now],
            )?;
            card = tx.last_insert_rowid();
            tx.execute(
                "INSERT INTO reservation (card_id, generation, holder, status, acquired_at, \
                 lease_until, updated_at, worktree, branch) \
                 VALUES (?1, 1, ?2, ?3, ?4, ?5, ?4, ?6, 'work')",
                rusqlite::params![
                    card,
                    h.holder,
                    h.status,
                    now,
                    now + h.lease,
                    h.worktree.to_string_lossy()
                ],
            )?;
            Ok(())
        })
        .unwrap();
        card
    }

    fn note(&self, card: i64, body: &str) {
        let mut l = ledger::open_write(self.home.path(), "demo").unwrap();
        l.with_write_tx(|tx, now| {
            tx.execute(
                "INSERT INTO card_note (card_id, kind, author, body, created_at) \
                 VALUES (?1, 'report', 'test', ?2, ?3)",
                rusqlite::params![card, body, now],
            )?;
            Ok(())
        })
        .unwrap();
    }

    /// One render as Claude Code runs it from `dir` for the session `session_id`.
    fn render(&self, dir: &Path, session_id: &str, worktree_name: Option<&str>) -> String {
        let before = self.home.snapshot();
        let mut stdin = serde_json::json!({
            "session_id": session_id,
            "cwd": dir,
            "model": {"display_name": "Opus"},
            "workspace": {"current_dir": dir, "project_dir": self.project},
        });
        if let Some(name) = worktree_name {
            stdin["worktree"] = serde_json::json!({"name": name});
        }
        let mut cmd = self
            .home
            .command(&["statusline", "--render-budget-ms", "3600000"]);
        cmd.current_dir(dir)
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
        assert_eq!(
            self.home.snapshot(),
            before,
            "the render changed the HOME or the project"
        );
        String::from_utf8(out.stdout).unwrap()
    }
}

fn line(out: &str, n: usize) -> String {
    out.lines().nth(n).unwrap_or_default().to_string()
}

/// A card whose held reservation names the session's worktree is my card, whoever holds it; a
/// live lease means working, and a marker note moves it on.
#[test]
fn my_card_is_the_card_held_in_this_worktree() {
    let f = fixture();
    let wt = f.worktree("demo-1-work", "spec/20260101-demo-work");
    let card = f.hold(&Hold {
        card_state: "active",
        status: "held",
        holder: "another-session",
        worktree: &wt,
        lease: 3600,
    });
    let out = f.render(&wt, "this-session", Some("demo-1-work"));
    assert_eq!(
        line(&out, 2),
        format!("🎯 c{card} W | 🗂️  D0 R0 W1 E0 ✅ 0"),
        "{out}"
    );
    f.note(card, "LANE-READY-INTEGRATION C-1 head=abc");
    let out = f.render(&wt, "this-session", Some("demo-1-work"));
    // The group of the open stages has nothing left to count, so it is left out.
    assert_eq!(
        line(&out, 2),
        format!("🎯 c{card} S | 👀 S1 A0 L0"),
        "{out}"
    );
    f.note(card, "LANE-BLOCKED kind=decision which option");
    let out = f.render(&wt, "this-session", Some("demo-1-work"));
    // The attention group shows both counts when either is not zero.
    assert_eq!(
        line(&out, 2),
        format!("🎯 c{card} B | ⛔ B1 ⏸️  H0"),
        "{out}"
    );
}

/// From the main checkout, with no worktree of mine, the newest live card the session holds is
/// mine; a card held by someone else is not.
#[test]
fn my_card_falls_back_to_the_newest_live_card_of_the_session() {
    let f = fixture();
    let elsewhere = f.worktree("demo-2-else", "spec/20260102-else");
    let mine = f.hold(&Hold {
        card_state: "active",
        status: "held",
        holder: "this-session",
        worktree: &elsewhere,
        lease: 3600,
    });
    let theirs = f.worktree("demo-3-theirs", "spec/20260103-theirs");
    f.hold(&Hold {
        card_state: "active",
        status: "held",
        holder: "someone-else",
        worktree: &theirs,
        lease: 3600,
    });
    let out = f.render(&f.project, "this-session", None);
    assert!(
        line(&out, 2).starts_with(&format!("🎯 c{mine} W | ")),
        "{out}"
    );
    let out = f.render(&f.project, "a-stranger", None);
    assert!(!line(&out, 2).contains('🎯'), "no card of mine: {out}");
}

/// A lease that is over makes the card stale (E), and it stays my card while its held
/// reservation names this worktree (D-3): the line says my card needs attention. For a card of
/// the session's holder in another folder, an ended lease means it is no longer mine.
#[test]
fn an_expired_lease_makes_my_card_stale_and_a_session_card_not_mine() {
    let f = fixture();
    let wt = f.worktree("demo-1-work", "spec/20260101-demo-work");
    f.hold(&Hold {
        card_state: "active",
        status: "held",
        holder: "this-session",
        worktree: &wt,
        lease: -60,
    });
    let out = f.render(&wt, "this-session", Some("demo-1-work"));
    assert_eq!(line(&out, 2), "🎯 c1 E | 🗂️  D0 R0 W0 E1 ✅ 0", "{out}");
    // From the main checkout the worktree does not match, and the session's own card has no
    // live lease: nothing is mine.
    let out = f.render(&f.project, "this-session", None);
    assert_eq!(line(&out, 2), "🗂️  D0 R0 W0 E1 ✅ 0", "{out}");
}

/// Each line has one role: line 1 the session, line 2 the repository and the worktree, line 3
/// the flow, the last line the usage windows. No icon of one role shows on another's line.
#[test]
fn three_lines_have_their_roles() {
    let f = fixture();
    let wt = f.worktree("demo-1-work", "spec/20260101-demo-work");
    f.hold(&Hold {
        card_state: "active",
        status: "held",
        holder: "this-session",
        worktree: &wt,
        lease: 3600,
    });
    let out = f.render(&wt, "this-session", Some("demo-1-work"));
    let lines: Vec<&str> = out.lines().collect();
    assert_eq!(lines.len(), 4, "{out}");
    assert!(lines[0].starts_with("🏷️"), "{out}");
    assert!(lines[1].contains("🅱️") && lines[1].contains("🌳"), "{out}");
    assert!(
        lines[2].starts_with("🎯") && lines[2].contains("🗂️"),
        "{out}"
    );
    assert!(lines[3].contains("CW"), "{out}");
    for (n, role_icons) in [
        (0, ["🎯", "👀", "⛔", "⏸️", "🗂️", "✅"].as_slice()),
        (1, ["🎯", "👀", "⛔", "⏸️", "🗂️", "✅"].as_slice()),
        (2, ["🌳", "🅱️", "📡", "🧠"].as_slice()),
        (3, ["🎯", "🗂️", "🌳", "🅱️"].as_slice()),
    ] {
        for icon in role_icons {
            assert!(!lines[n].contains(icon), "line {n} carries {icon}: {out}");
        }
    }
}

/// Every icon the flow line writes is followed by the spacing its width class asks for, in the
/// real output: two spaces after an icon that is narrow by default, one after the others.
#[test]
fn icon_spacing_in_the_real_output() {
    let f = fixture();
    let wt = f.worktree("demo-1-work", "spec/20260101-demo-work");
    let card = f.hold(&Hold {
        card_state: "active",
        status: "held",
        holder: "this-session",
        worktree: &wt,
        lease: 3600,
    });
    f.note(card, "LANE-BLOCKED kind=decision x");
    {
        let mut l = ledger::open_write(f.home.path(), "demo").unwrap();
        l.with_write_tx(|tx, now| {
            for state in ["hold", "draft", "ready"] {
                tx.execute(
                    "INSERT INTO card (title, state, created_by, created_at, updated_at) \
                     VALUES ('t', ?1, 'test', ?2, ?2)",
                    rusqlite::params![state, now],
                )?;
            }
            tx.execute(
                "INSERT INTO card (title, state, created_by, created_at, updated_at, closed_at) \
                 VALUES ('t', 'done', 'test', ?1, ?1, ?1)",
                rusqlite::params![now],
            )?;
            Ok(())
        })
        .unwrap();
    }
    let out = f.render(&wt, "this-session", Some("demo-1-work"));
    let flow = line(&out, 2);
    for icon in ["🎯", "⛔", "⏸️", "🗂️", "✅"] {
        assert!(flow.contains(icon), "{icon} missing in {flow}");
    }
    for (icon, gap) in ICONS {
        for (at, _) in flow.match_indices(icon) {
            let after = &flow[at + icon.len()..];
            assert!(after.starts_with(gap), "{icon}: `{after}` in {flow}");
            assert!(
                !after[gap.len()..].starts_with(' '),
                "{icon}: too many spaces in {flow}"
            );
        }
    }
}
