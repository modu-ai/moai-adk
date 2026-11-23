//! SPEC-20261124-ready-receipt-gate: `moai card note` refuses a `LANE-READY-INTEGRATION` or
//! `REPAIRED` note whose head has no PASS `check affected` receipt, and `moai check affected`
//! leaves the receipt (`summary.txt`) the gate reads. Every case runs against a temporary HOME
//! and a temporary repository that is also the project folder; the real HOME and the real
//! repository are not read or written.
#![cfg(unix)]

mod common;
#[path = "common/ready_receipt.rs"]
mod ready_receipt;

use common::git::{git, git_ok};
use common::home::TempHome;
use moai_cli::card::inbox::{family_of, Family};
use moai_cli::ledger;
use serde_json::Value;
use std::fs;
use std::io::Write;
use std::path::{Path, PathBuf};
use std::process::Stdio;

/// A project folder that is a git repository with two commits: `base` and `head` (the second
/// one carries a trailing blank, so `git diff --check base...head` fails).
struct World {
    home: TempHome,
    project: PathBuf,
    base: String,
    head: String,
}

fn commit(dir: &Path, rel: &str, text: &str, msg: &str) -> String {
    let path = dir.join(rel);
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, text).unwrap();
    git_ok(dir, &["add", "-A"]);
    git_ok(dir, &["commit", "-q", "-m", msg]);
    git(dir, &["rev-parse", "HEAD"])
}

impl World {
    fn new() -> World {
        let home = TempHome::new();
        let project = home.project("demo");
        ledger::register(home.path(), "demo").unwrap();
        git_ok(&project, &["init", "-q", "-b", "main"]);
        git_ok(&project, &["config", "user.email", "t@example.invalid"]);
        git_ok(&project, &["config", "user.name", "t"]);
        fs::write(
            project.join(".gitignore"),
            format!("target/\n{}/\n", moai_cli::PROJECT_STATE_DIR),
        )
        .unwrap();
        fs::write(
            project.join("Cargo.toml"),
            "[package]\nname = \"demo\"\nversion = \"0.1.0\"\nedition = \"2021\"\n",
        )
        .unwrap();
        let base = commit(&project, "src/lib.rs", "pub fn a() {}\n", "base");
        let head = commit(&project, "src/lib.rs", "pub fn a() {}   \n", "head");
        World {
            home,
            project,
            base,
            head,
        }
    }

    fn project_arg(&self) -> String {
        self.project.to_str().unwrap().to_string()
    }

    /// `moai <args>`: the success flag and what it printed on both streams.
    fn moai(&self, args: &[&str], stdin: Option<&str>) -> (bool, String) {
        let mut cmd = self.home.command(args);
        cmd.stdin(if stdin.is_some() {
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

    fn new_card(&self) -> String {
        self.card_json(&["add", "--title", "work", "--scope", "area/", "--ac", "done"])["name"]
            .as_str()
            .unwrap()
            .to_string()
    }

    /// `card note` with `text`; the success flag and the printed text.
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

    /// `moai check affected <base>..<head>` in the project; the answer's `result`. The run's
    /// first step (`git diff --check`) fails, so no cargo step starts.
    fn check_affected(&self) -> Value {
        let slots = self.home.path().join("slots");
        let range = format!("{}..{}", self.base, self.head);
        let project = self.project_arg();
        let slots = slots.to_str().unwrap();
        let (_, text) = self.moai(
            &[
                "check",
                "affected",
                &range,
                "--dir",
                &project,
                "--slot-dir",
                slots,
                "--max",
                "1",
                "--min-free-gb",
                "0",
            ],
            None,
        );
        let line = text.lines().next().unwrap_or("");
        let v: Value =
            serde_json::from_str(line).unwrap_or_else(|e| panic!("{e}: check affected: {text}"));
        assert_eq!(v["ok"], true, "{text}");
        v["result"].clone()
    }

    fn ready(&self, card: &str, head: &str) -> String {
        format!(
            "LANE-READY-INTEGRATION {card} head={head}\nbase={}\n",
            self.base
        )
    }

    fn repaired(&self, card: &str, head: &str) -> String {
        format!("REPAIRED {card} head={head}\nbase={}\n", self.base)
    }
}

#[test]
fn check_affected_leaves_a_summary_file() {
    let w = World::new();
    let r = w.check_affected();
    let summary = r["summary"].as_str().unwrap();
    let out = Path::new(r["out"].as_str().unwrap());
    let real = fs::canonicalize(&w.project).unwrap();
    assert_eq!(out, ready_receipt::receipt_dir(&real, &w.head));
    let file = fs::read_to_string(out.join("summary.txt")).expect("summary.txt exists");
    assert_eq!(file, format!("{summary}\n"));
}

#[test]
fn a_ready_and_a_repaired_note_with_a_pass_receipt_are_kept() {
    let w = World::new();
    let card = w.new_card();
    ready_receipt::write_receipt(&w.project, &w.head, "PASS");
    let (ok, text) = w.note(&card, &w.ready(&card, &w.head));
    assert!(ok, "{text}");
    let (ok, text) = w.note(&card, &w.repaired(&card, &w.head));
    assert!(ok, "{text}");
    assert_eq!(w.note_count(&card), 2);
}

#[test]
fn a_ready_note_without_a_receipt_is_refused_with_the_way_out() {
    let w = World::new();
    let card = w.new_card();
    let (ok, text) = w.note(&card, &w.ready(&card, &w.head));
    assert!(!ok, "{text}");
    assert!(text.contains("card note refused"), "{text}");
    assert!(text.contains(&w.head), "names the head: {text}");
    // git names a worktree by its real path (macOS temporary folders sit behind a link).
    let real = fs::canonicalize(&w.project).unwrap();
    let dir = ready_receipt::receipt_dir(&real, &w.head);
    assert!(
        text.contains(dir.join("summary.txt").to_str().unwrap()),
        "names the receipt path: {text}"
    );
    assert!(text.contains("missing"), "{text}");
    assert!(text.contains("moai check affected"), "{text}");
    assert_eq!(w.note_count(&card), 0, "nothing was written");
}

#[test]
fn a_bad_receipt_counts_as_no_receipt() {
    let w = World::new();
    let card = w.new_card();
    let path = ready_receipt::write_receipt(&w.project, &w.head, "FAIL");
    let good = fs::read_to_string(&path).unwrap();
    let other = ready_receipt::summary_line(&w.base, "PASS");
    let bad: [(&str, Vec<u8>); 4] = [
        ("a FAIL result", good.clone().into_bytes()),
        ("another head", format!("{other}\n").into_bytes()),
        (
            "two lines",
            format!("{}\n{}\n", good.trim_end().replace("FAIL", "PASS"), "more").into_bytes(),
        ),
        ("not UTF-8", vec![0xff, 0xfe, 0xfd]),
    ];
    for (what, bytes) in bad {
        fs::write(&path, bytes).unwrap();
        let (ok, text) = w.note(&card, &w.ready(&card, &w.head));
        assert!(!ok, "{what} was kept: {text}");
        assert!(text.contains("card note refused"), "{what}: {text}");
    }
    assert_eq!(w.note_count(&card), 0);
}

#[test]
fn the_receipt_is_looked_for_in_the_worktree_that_has_the_head() {
    let w = World::new();
    let card = w.new_card();
    // The project has `head` checked out; the receipt of `base` sits in its target folder, but
    // no worktree has `base` as its HEAD.
    ready_receipt::write_receipt(&w.project, &w.base, "PASS");
    let (ok, text) = w.note(&card, &w.ready(&card, &w.base));
    assert!(!ok, "{text}");
    // A linked worktree at `base` with its own receipt makes `base` acceptable.
    let linked = w.home.path().join("linked");
    git_ok(
        &w.project,
        &[
            "worktree",
            "add",
            "-q",
            "--detach",
            linked.to_str().unwrap(),
            &w.base,
        ],
    );
    let (ok, text) = w.note(&card, &w.ready(&card, &w.base));
    assert!(!ok, "no receipt in the linked worktree yet: {text}");
    ready_receipt::write_receipt(&linked, &w.base, "PASS");
    let (ok, text) = w.note(&card, &w.ready(&card, &w.base));
    assert!(ok, "{text}");
}

#[test]
fn a_receipt_whose_base_is_not_an_ancestor_of_the_head_is_refused() {
    let w = World::new();
    let card = w.new_card();
    // A commit with the head's tree and no parent: it is in no history of `head`.
    let tree = git(&w.project, &["rev-parse", &format!("{}^{{tree}}", w.head)]);
    let stranger = git(&w.project, &["commit-tree", &tree, "-m", "stranger"]);
    ready_receipt::write_receipt_based(&w.project, &w.head, &stranger);
    let (ok, text) = w.note(&card, &w.ready(&card, &w.head));
    assert!(!ok, "{text}");
    assert!(text.contains("card note refused"), "{text}");
    assert!(text.contains(&stranger[..12]), "names the base: {text}");
    assert!(text.contains("not an ancestor"), "{text}");
    assert_eq!(w.note_count(&card), 0);
    // A base that is not a commit of this repository counts the same.
    ready_receipt::write_receipt_based(
        &w.project,
        &w.head,
        "deadbeefdead00000000000000000000deadbeef",
    );
    let (ok, text) = w.note(&card, &w.ready(&card, &w.head));
    assert!(!ok, "{text}");
    // The real base is an ancestor.
    ready_receipt::write_receipt_based(&w.project, &w.head, &w.base);
    let (ok, text) = w.note(&card, &w.repaired(&card, &w.head));
    assert!(ok, "{text}");
}

#[test]
fn a_marker_line_of_the_wrong_shape_is_refused() {
    let w = World::new();
    let card = w.new_card();
    ready_receipt::write_receipt(&w.project, &w.head, "PASS");
    let short = &w.head[..12];
    let lines = [
        format!("LANE-READY-INTEGRATION {card} head={short}"),
        format!("LANE-READY-INTEGRATION C-9999 head={}", w.head),
        format!("LANE-READY-INTEGRATION {card} {}", w.head),
        format!("REPAIRED {card} head={}x", w.head),
        "REPAIRED".to_string(),
        // A marker with a character glued to it is still a marker (SPEC-20261124-ready-receipt-followup).
        format!("LANE-READY-INTEGRATION: {card} head={}", w.head),
        format!("REPAIRED-X {card} head={}", w.head),
        format!("READY {card} head={short}"),
        format!("READY {card}"),
        format!("READY (head {short})"),
        format!(
            "LANE-READY-INTEGRATION {card} head={}",
            w.head.to_uppercase()
        ),
    ];
    for line in lines {
        let (ok, text) = w.note(&card, &line);
        assert!(!ok, "`{line}` was kept: {text}");
        assert!(text.contains("card note refused"), "{line}: {text}");
    }
    assert_eq!(w.note_count(&card), 0);
}

#[test]
fn other_notes_are_kept_and_there_is_no_skip_flag() {
    let w = World::new();
    let card = w.new_card();
    let texts = [
        format!("LANE-BLOCKED kind=decision {card} x"),
        format!("HEAD-MOVED {card} old={} new={}", w.base, w.head),
        format!(
            "AUDIT {card} head {} base {} — verdict: PASS",
            w.head, w.base
        ),
        format!(
            "LANE-NOTE only a word\nLANE-READY-INTEGRATION {card} head={}",
            w.head
        ),
        format!("READYING {card} head={}", w.head),
        format!("MERGE-READY {card} head={}", w.head),
    ];
    for t in &texts {
        let (ok, text) = w.note(&card, t);
        assert!(ok, "`{t}` was refused: {text}");
    }
    assert_eq!(w.note_count(&card), texts.len());
    let (ok, help) = w.moai(&["card", "note", "--help"], None);
    assert!(ok, "{help}");
    for word in [
        "--force",
        "--skip",
        "--no-receipt",
        "--no-check",
        "--unchecked",
    ] {
        assert!(!help.contains(word), "{word} in the help: {help}");
    }
}

#[test]
fn a_failed_run_leaves_a_fail_receipt_and_the_refusal_says_so() {
    let w = World::new();
    let card = w.new_card();
    let r = w.check_affected();
    assert_eq!(r["result"], "FAIL", "{r}");
    let file =
        fs::read_to_string(ready_receipt::receipt_dir(&w.project, &w.head).join("summary.txt"))
            .unwrap();
    assert!(file.trim_end().ends_with("result=FAIL"), "{file}");
    let (ok, text) = w.note(&card, &w.ready(&card, &w.head));
    assert!(!ok, "{text}");
    assert!(text.contains("result=FAIL"), "{text}");
    assert_eq!(w.note_count(&card), 0);
}

#[test]
fn any_worktree_with_the_head_and_a_pass_receipt_is_enough() {
    let w = World::new();
    let card = w.new_card();
    let linked = w.home.path().join("slot");
    git_ok(
        &w.project,
        &[
            "worktree",
            "add",
            "-q",
            "--detach",
            linked.to_str().unwrap(),
            &w.head,
        ],
    );
    ready_receipt::write_receipt(&w.project, &w.head, "FAIL");
    let (ok, text) = w.note(&card, &w.ready(&card, &w.head));
    assert!(!ok, "{text}");
    let project = fs::canonicalize(&w.project).unwrap();
    let slot = fs::canonicalize(&linked).unwrap();
    assert!(
        text.contains(project.to_str().unwrap()),
        "names the project worktree: {text}"
    );
    assert!(
        text.contains(slot.to_str().unwrap()),
        "names the slot worktree: {text}"
    );
    assert!(text.contains("missing"), "{text}");
    assert!(text.contains("result=FAIL"), "{text}");
    ready_receipt::write_receipt(&linked, &w.head, "PASS");
    let (ok, text) = w.note(&card, &w.ready(&card, &w.head));
    assert!(ok, "{text}");
    assert_eq!(w.note_count(&card), 1);
}

// SPEC-20261124-ready-receipt-followup AC-1: a marker with a character glued to it is a marker for
// the gate as it is for `card inbox` (the status line is pinned by its own unit test).
#[test]
fn a_glued_marker_is_one_rule_for_gate_inbox_and_status_line() {
    let w = World::new();
    let card = w.new_card();
    // A PASS receipt exists, so a refusal below is the shape of the line and nothing else.
    ready_receipt::write_receipt(&w.project, &w.head, "PASS");
    for line in [
        format!("LANE-READY-INTEGRATION: {card} head={}", w.head),
        format!("LANE-READY-INTEGRATION{card} head={}", w.head),
        format!("  LANE-READY-INTEGRATION {card} head={}", w.head),
    ] {
        let (ok, text) = w.note(&card, &line);
        assert!(!ok, "`{line}` was kept: {text}");
        assert!(text.contains("card note refused"), "{line}: {text}");
        assert_eq!(
            family_of(&line),
            Some(Family::ReadyForIntegration),
            "{line}"
        );
    }
    let line = format!("REPAIRED-X {card} head={}", w.head);
    let (ok, text) = w.note(&card, &line);
    assert!(!ok, "`{line}` was kept: {text}");
    assert_eq!(w.note_count(&card), 0);
}

// AC-3: a plain `READY` hand-back needs the receipt, `MERGE-READY` is the leader's and does not.
#[test]
fn a_plain_ready_note_needs_the_receipt_and_merge_ready_does_not() {
    let w = World::new();
    let card = w.new_card();
    let ready = format!("READY {card} head={}\nbase={}\n", w.head, w.base);
    let (ok, text) = w.note(&card, &ready);
    assert!(!ok, "{text}");
    assert!(text.contains("card note refused"), "{text}");
    assert!(text.contains("moai check affected"), "{text}");
    assert_eq!(w.note_count(&card), 0);
    let merge = format!("MERGE-READY {card} head={}\n", w.base);
    let (ok, text) = w.note(&card, &merge);
    assert!(ok, "a MERGE-READY note needs no receipt: {text}");
    ready_receipt::write_receipt(&w.project, &w.head, "PASS");
    let (ok, text) = w.note(&card, &ready);
    assert!(ok, "{text}");
    assert_eq!(w.note_count(&card), 2);
}

// AC-5: the receipt of a `--lane` run is no receipt; the sentence names the bare command.
#[test]
fn a_lane_receipt_is_refused_with_the_bare_command() {
    let w = World::new();
    let card = w.new_card();
    let path = ready_receipt::write_receipt_lane(&w.project, &w.head);
    let (ok, text) = w.note(&card, &w.ready(&card, &w.head));
    assert!(!ok, "{text}");
    assert!(text.contains("selection=lane"), "{text}");
    assert!(text.contains("--lane"), "{text}");
    assert!(text.contains("moai check affected"), "{text}");
    assert!(
        text.contains(path.to_str().unwrap_or("")) || text.contains("summary.txt"),
        "{text}"
    );
    assert_eq!(w.note_count(&card), 0);
    ready_receipt::write_receipt(&w.project, &w.head, "PASS");
    let (ok, text) = w.note(&card, &w.ready(&card, &w.head));
    assert!(ok, "the bare receipt is kept: {text}");
}
