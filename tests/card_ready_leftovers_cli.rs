//! SPEC-20261124-lane-leftovers: a hand-back note needs one `leftovers:` line, `moai card note`
//! lists the files that appeared in the primary checkout since the card was reserved, and the lane
//! guidance says where a lane makes its throwaway copies. Every case runs against a temporary HOME
//! and a temporary repository that is also the project folder; the real HOME and the real
//! repository are not read or written.
#![cfg(unix)]

mod common;
#[path = "common/ready_receipt.rs"]
mod ready_receipt;

use common::git::{git, git_ok};
use common::home::TempHome;
use moai_cli::ledger;
use serde_json::Value;
use std::fs;
use std::io::Write;
use std::path::{Path, PathBuf};
use std::process::Stdio;
use std::time::Duration;

const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
const SESSION: &str = "aaaaaaaa-0000-4000-8000-0000000000a1";

struct World {
    home: TempHome,
    project: PathBuf,
    head: String,
}

impl World {
    fn new() -> World {
        let home = TempHome::new();
        let project = home.project("demo");
        ledger::register(home.path(), "demo").unwrap();
        git_ok(&project, &["init", "-q", "-b", "main"]);
        git_ok(&project, &["config", "user.email", "t@example.invalid"]);
        git_ok(&project, &["config", "user.name", "t"]);
        // The ignore rules live in the repository's own exclude file, so the fixture names no
        // checked-in file of this repository.
        fs::create_dir_all(project.join(".git/info")).unwrap();
        fs::write(
            project.join(".git/info/exclude"),
            format!(
                "target/\n{}/\n*.log\nscratch/\n",
                moai_cli::PROJECT_STATE_DIR
            ),
        )
        .unwrap();
        fs::create_dir_all(project.join("app")).unwrap();
        fs::write(project.join("app/demo.rs"), "pub fn a() {}\n").unwrap();
        git_ok(&project, &["add", "-A"]);
        git_ok(&project, &["commit", "-q", "-m", "head"]);
        let head = git(&project, &["rev-parse", "HEAD"]);
        ready_receipt::write_receipt(&project, &head, "PASS");
        World {
            home,
            project,
            head,
        }
    }

    fn project_arg(&self) -> String {
        self.project.to_str().unwrap().to_string()
    }

    /// `moai <args>` as the lane session: the success flag and what it printed on both streams.
    fn moai(&self, args: &[&str], stdin: Option<&str>) -> (bool, String) {
        let mut cmd = self.home.command(args);
        cmd.env(SESSION_VAR, SESSION)
            .stdin(if stdin.is_some() {
                Stdio::piped()
            } else {
                Stdio::null()
            })
            .stdout(Stdio::piped())
            .stderr(Stdio::piped());
        let mut child = cmd.spawn().expect("run moai");
        if let Some(t) = stdin {
            child.stdin.take().unwrap().write_all(t.as_bytes()).unwrap();
        }
        let out = child.wait_with_output().unwrap();
        let text = format!(
            "{}{}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        );
        (out.status.success(), text)
    }

    fn card_json(&self, args: &[&str]) -> Value {
        let project = self.project_arg();
        let mut all = vec!["card"];
        all.extend_from_slice(args);
        all.extend_from_slice(&["--project", &project]);
        let (ok, text) = self.moai(&all, None);
        assert!(ok, "card {args:?}: {text}");
        let line = text.lines().next().unwrap_or("");
        let v: Value = serde_json::from_str(line).unwrap_or_else(|e| panic!("{e}: {text}"));
        v["result"].clone()
    }

    /// A ready card, not taken.
    fn new_card(&self) -> String {
        let name = self.card_json(&["add", "--title", "work", "--scope", "area/", "--ac", "done"])
            ["name"]
            .as_str()
            .unwrap()
            .to_string();
        self.card_json(&["ready", &name]);
        name
    }

    /// Takes `card` for the lane session in a linked worktree at the head.
    fn take(&self, card: &str) -> PathBuf {
        let linked = self.home.path().join("lane-worktree");
        git_ok(
            &self.project,
            &[
                "worktree",
                "add",
                "-q",
                "--detach",
                linked.to_str().unwrap(),
                &self.head,
            ],
        );
        self.card_json(&[
            "take",
            card,
            "--scope",
            "area/",
            "--worktree",
            linked.to_str().unwrap(),
            "--branch",
            "lane/x",
        ]);
        linked
    }

    /// `card note` with `text`: the success flag and the printed text.
    fn note(&self, card: &str, text: &str) -> (bool, String) {
        let project = self.project_arg();
        self.moai(
            &[
                "card",
                "note",
                card,
                "--kind",
                "report",
                "--by",
                "lane",
                "--project",
                &project,
            ],
            Some(text),
        )
    }

    fn note_count(&self, card: &str) -> usize {
        self.card_json(&["show", card])["notes"]
            .as_array()
            .unwrap()
            .len()
    }

    fn marker_line(&self, marker: &str, card: &str) -> String {
        format!("{marker} {card} head={}\nbase={}\n", self.head, self.head)
    }
}

fn answer(text: &str) -> Value {
    let line = text.lines().next().unwrap_or("");
    let v: Value = serde_json::from_str(line).unwrap_or_else(|e| panic!("{e}: {text}"));
    v["result"].clone()
}

fn repo_file(rel: &str) -> String {
    let path = Path::new(env!("CARGO_MANIFEST_DIR")).join(rel);
    fs::read_to_string(&path).unwrap_or_else(|e| panic!("{}: {e}", path.display()))
}

#[test]
fn the_lane_guidance_names_the_throwaway_place() {
    for rel in [
        "docs/lane-rules.md",
        "skills/moai/references/flows/next-flow.md",
    ] {
        let text = repo_file(rel);
        for needle in ["target/<name>/", "TMPDIR", "relative path", "leftovers:"] {
            assert!(text.contains(needle), "{rel} lacks `{needle}`");
        }
    }
    let rules = repo_file("docs/lane-rules.md");
    assert!(
        rules.contains("<worktree>/target/<name>/"),
        "docs/lane-rules.md names the place as <worktree>/target/<name>/"
    );
}

#[test]
fn a_ready_note_without_a_readable_leftovers_line_is_refused() {
    let w = World::new();
    let card = w.new_card();
    for marker in ["LANE-READY-INTEGRATION", "REPAIRED", "READY"] {
        let base = w.marker_line(marker, &card);
        let (ok, text) = w.note(&card, &base);
        assert!(!ok, "{marker} without the line was kept: {text}");
        assert!(text.contains("card note refused"), "{marker}: {text}");
        assert!(text.contains("leftovers: none"), "names the line: {text}");
        assert_eq!(w.note_count(&card), 0, "{marker}: nothing was written");
    }
    let base = w.marker_line("LANE-READY-INTEGRATION", &card);
    let bad = [
        "leftovers:\n",
        "leftovers: <path>\n",
        "leftovers: none | build/x\n",
        "leftovers: none\nleftovers: none\n",
        "```\nleftovers: none\n```\n",
        "    leftovers: none\n",
    ];
    for tail in bad {
        let (ok, text) = w.note(&card, &format!("{base}{tail}"));
        assert!(!ok, "{tail:?} was kept: {text}");
        assert!(text.contains("card note refused"), "{tail:?}: {text}");
    }
    assert_eq!(w.note_count(&card), 0);
    // The readable forms are kept, and a note that is no hand-back needs no line.
    for tail in [
        "leftovers: none\n",
        "- leftovers: target/a/copy | target/b\n",
    ] {
        let (ok, text) = w.note(&card, &format!("{base}{tail}"));
        assert!(ok, "{tail:?} was refused: {text}");
    }
    let (ok, text) = w.note(&card, "LANE-NOTE progress, no hand-back");
    assert!(ok, "{text}");
    assert_eq!(w.note_count(&card), 3);
}

#[test]
fn the_receipt_is_still_judged_when_the_line_is_there() {
    let w = World::new();
    let card = w.new_card();
    // A hand-back of another head has no receipt: the line does not replace it.
    let other = "0123456789abcdef0123456789abcdef01234567";
    let text = format!("LANE-READY-INTEGRATION {card} head={other}\nleftovers: none\n");
    let (ok, out) = w.note(&card, &text);
    assert!(!ok, "{out}");
    assert!(out.contains("receipt"), "{out}");
    assert_eq!(w.note_count(&card), 0);
}

#[test]
fn the_shared_checkout_is_read_without_changing_it() {
    let w = World::new();
    let card = w.new_card();
    // Older than the reservation: not listed.
    fs::write(w.project.join("old_untracked.txt"), "x").unwrap();
    fs::write(w.project.join("old.log"), "x").unwrap();
    std::thread::sleep(Duration::from_millis(2100));
    let lane = w.take(&card);
    std::thread::sleep(Duration::from_millis(2100));
    // New in the primary checkout after the reservation: listed.
    fs::write(w.project.join("new_untracked.txt"), "x").unwrap();
    fs::write(w.project.join("cache.log"), "x").unwrap();
    fs::create_dir_all(w.project.join("newdir")).unwrap();
    fs::write(w.project.join("newdir/inner.txt"), "x").unwrap();
    fs::create_dir_all(w.project.join("scratch/deep")).unwrap();
    fs::write(w.project.join("scratch/deep/copy.bin"), "x").unwrap();
    // Expected state: not listed.
    for name in [".DS_Store", "CLAUDE.local.md", ".mcp.json"] {
        fs::write(w.project.join(name), "x").unwrap();
    }
    fs::create_dir_all(w.project.join(".claude/worktrees/other")).unwrap();
    fs::write(w.project.join("target/built.o"), "x").unwrap();
    // A file in the lane's own worktree is no business of the primary checkout.
    fs::write(lane.join("lane_only.txt"), "x").unwrap();
    // A tracked file touched without a change in content: a plain `git status` would refresh the
    // index's stat cache and write the index file.
    fs::write(w.project.join("app/demo.rs"), "pub fn a() {}\n").unwrap();

    let index = w.project.join(".git/index");
    let bytes_before = fs::read(&index).unwrap();
    let mtime_before = fs::metadata(&index).unwrap().modified().unwrap();
    let status = |w: &World| {
        let mut cmd = common::git::command();
        cmd.current_dir(&w.project)
            .env("GIT_OPTIONAL_LOCKS", "0")
            .args(["status", "--porcelain=v1", "--ignored=matching", "-uall"]);
        String::from_utf8(cmd.output().unwrap().stdout).unwrap()
    };
    let status_before = status(&w);

    let text = format!(
        "{}leftovers: none\n",
        w.marker_line("LANE-READY-INTEGRATION", &card)
    );
    let (ok, out) = w.note(&card, &text);
    assert!(ok, "the note is kept while the list is observed: {out}");
    let a = answer(&out);
    assert_eq!(a["shared_checkout"], "checked", "{out}");
    let mut listed: Vec<String> = a["shared_checkout_new"]
        .as_array()
        .unwrap()
        .iter()
        .map(|v| v.as_str().unwrap().to_string())
        .collect();
    listed.sort();
    assert_eq!(
        listed,
        [
            "cache.log",
            "new_untracked.txt",
            "newdir/inner.txt",
            "scratch/"
        ],
        "{out}"
    );

    assert_eq!(fs::read(&index).unwrap(), bytes_before, "index bytes");
    assert_eq!(
        fs::metadata(&index).unwrap().modified().unwrap(),
        mtime_before,
        "index mtime"
    );
    assert_eq!(status(&w), status_before, "git status output");
}

#[test]
fn a_card_without_a_reservation_is_not_checked_and_the_note_is_kept() {
    let w = World::new();
    let card = w.new_card();
    fs::write(w.project.join("new_untracked.txt"), "x").unwrap();
    let text = format!(
        "{}leftovers: none\n",
        w.marker_line("LANE-READY-INTEGRATION", &card)
    );
    let (ok, out) = w.note(&card, &text);
    assert!(ok, "{out}");
    let a = answer(&out);
    assert_eq!(a["shared_checkout"], "not_checked", "{out}");
    assert!(
        a["shared_checkout_reason"]
            .as_str()
            .is_some_and(|r| !r.is_empty()),
        "{out}"
    );
    assert_eq!(
        a["shared_checkout_new"].as_array().unwrap().len(),
        0,
        "{out}"
    );
}

#[test]
fn the_leftovers_line_is_shown_where_a_lane_looks() {
    for rel in [
        "docs/lane-rules.md",
        "docs/design-review/appendix/operations/card.md",
    ] {
        assert!(
            repo_file(rel).contains("leftovers:"),
            "{rel} shows the `leftovers:` line"
        );
    }
    let rules = repo_file("docs/lane-rules.md");
    let hand_off = rules
        .split("## Hand-off")
        .nth(1)
        .and_then(|s| s.split("\n## ").next())
        .expect("the Hand-off section");
    assert!(hand_off.contains("leftovers:"), "the Hand-off section");
    let w = TempHome::new();
    let out = w
        .command(&["card", "note", "--help"])
        .output()
        .expect("run moai");
    let help = String::from_utf8_lossy(&out.stdout);
    assert!(help.contains("leftovers:"), "card note --help: {help}");
}
