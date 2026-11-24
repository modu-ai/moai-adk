//! SPEC-20261124-card-select-batch-land: `card land-batch` joins audited heads onto the tip,
//! runs one wide check on the last join and lands each head in order; a failing check is split in
//! halves until the failing head is found. A throwaway repository, a throwaway HOME and a shell
//! script as the check program: the real checks, HOME and slots are never used.
#![cfg(unix)]
mod common;

use common::home::TempHome;
use moai_cli::ledger;
use rusqlite::Connection;
use serde_json::{json, Value};
use std::fs;
use std::os::unix::fs::PermissionsExt;
use std::path::{Path, PathBuf};
use std::process::{Command, Output, Stdio};

const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
const S_A: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
const S_AUD: &str = "cccccccc-0000-4000-8000-00000000000c";
const S_LEAD: &str = "dddddddd-0000-4000-8000-00000000000d";
const WORKTREE_KEY: &str = "0123456789ab";
/// The SPEC every fixture card names; the join commits are named for it.
const SPEC: &str = "SPEC-20261124-fixture";
const PROFILE: &[u8] = b"{\"schema_version\": 1, \"project_id\": \"demo\", \"purpose\": \"fixture\", \"tools\": [\"rust\"], \"development_methods\": [\"sdd\"], \"hosts\": [\"claude\"], \"work_flow\": {\"flow\": \"m-flow\", \"integration_branch\": \"integ\"}}";

/// A script that logs the range it was given and fails (exit 3) where the checkout holds
/// `notes/poison.md`.
const CHECK: &str = "#!/bin/sh\necho \"$1\" >> \"$BATCH_LOG\"\nif [ -e notes/poison.md ]; then exit 3; fi\nexit 0\n";

fn git(dir: &Path, args: &[&str]) -> String {
    let mut all = vec![
        "-c",
        "user.name=fixture",
        "-c",
        "user.email=fixture@example.invalid",
    ];
    all.extend_from_slice(args);
    common::git::git(dir, &all)
}

fn signed(subject: &str, session: &str) -> String {
    format!(
        "{subject}\n\nMoai-Spec: SPEC-20261124-fixture\nMoai-Session: {session}:{WORKTREE_KEY}\nMoai-Verified: none"
    )
}

/// A commit on `parent` that adds or changes `path`; no ref moves.
fn put_on(project: &Path, parent: &str, path: &str, content: &str, message: &str) -> String {
    let blob_file = project.join("put-on-blob.tmp");
    fs::write(&blob_file, content).unwrap();
    let blob = git(project, &["hash-object", "-w", "put-on-blob.tmp"]);
    fs::remove_file(&blob_file).unwrap();
    let index = project.join(".git/put-on-index");
    let with_index = |args: &[&str]| -> String {
        let out = common::git::command()
            .arg("-C")
            .arg(project)
            .args(args)
            .env("GIT_INDEX_FILE", &index)
            .output()
            .unwrap();
        assert!(out.status.success(), "git {args:?}");
        String::from_utf8_lossy(&out.stdout).trim().to_string()
    };
    with_index(&["read-tree", parent]);
    with_index(&[
        "update-index",
        "--add",
        "--cacheinfo",
        &format!("100644,{blob},{path}"),
    ]);
    let tree = with_index(&["write-tree"]);
    fs::remove_file(&index).unwrap();
    git(
        project,
        &["commit-tree", &tree, "-p", parent, "-m", message],
    )
}

fn git_executable() -> PathBuf {
    let path = std::env::var_os("PATH").expect("PATH");
    std::env::split_paths(&path)
        .map(|dir| dir.join("git"))
        .find(|p| p.is_file())
        .expect("a git on PATH")
}

fn envelope(out: &Output) -> Value {
    serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
        panic!(
            "stdout is not JSON ({e}): {} {}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    })
}

fn result(out: &Output) -> Value {
    assert_eq!(
        out.status.code(),
        Some(0),
        "{} {}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    let v = envelope(out);
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

/// A template's text with every `<...>` placeholder filled with the word `filled`.
fn filled(text: &str) -> String {
    let mut out = String::new();
    let mut rest = text;
    while let Some(open) = rest.find('<') {
        let Some(close) = rest[open..].find('>') else {
            break;
        };
        out.push_str(&rest[..open]);
        out.push_str("filled");
        rest = &rest[open + close + 1..];
    }
    out.push_str(rest);
    out
}

struct World {
    home: TempHome,
    project: PathBuf,
    p: String,
    git_bin: PathBuf,
    scripts: tempfile::TempDir,
    c0: String,
}

impl World {
    fn new() -> Self {
        let home = TempHome::new();
        let project = home.project("demo");
        ledger::register(home.path(), "demo").unwrap();
        git(&project, &["init", "-q", "-b", "main"]);
        git(&project, &["config", "user.name", "fixture"]);
        git(
            &project,
            &["config", "user.email", "fixture@example.invalid"],
        );
        fs::write(project.join("a.txt"), "a\n").unwrap();
        // The SPEC every fixture card names: a card is landed only when its SPEC is in the tip.
        let made = home
            .command(&[
                "spec",
                "new",
                "fixture",
                "--title",
                "t",
                "--minimal",
                "--date",
                "2026-11-24",
                "--project",
                project.to_str().unwrap(),
            ])
            .stdin(Stdio::null())
            .output()
            .expect("run moai spec new");
        let spec_file = project.join(".moai/specs").join(SPEC).join("spec.md");
        assert!(
            spec_file.is_file(),
            "{}",
            String::from_utf8_lossy(&made.stdout)
        );
        fs::write(&spec_file, filled(&fs::read_to_string(&spec_file).unwrap())).unwrap();
        git(&project, &["add", "a.txt", ".moai/specs"]);
        git(&project, &["commit", "-q", "-m", "init"]);
        let c0 = git(&project, &["rev-parse", "HEAD"]);
        git(&project, &["update-ref", "refs/heads/integ", &c0]);
        let scripts = tempfile::tempdir().unwrap();
        World {
            p: project.to_str().unwrap().to_string(),
            home,
            project,
            git_bin: git_executable(),
            scripts,
            c0,
        }
    }

    fn script(&self, name: &str, text: &str) -> PathBuf {
        let path = self.scripts.path().join(name);
        fs::write(&path, text).unwrap();
        fs::set_permissions(&path, fs::Permissions::from_mode(0o755)).unwrap();
        path
    }

    fn log(&self) -> PathBuf {
        self.scripts.path().join("batch.log")
    }

    /// The ranges the check program was given, one per run, in order.
    fn ranges(&self) -> Vec<String> {
        fs::read_to_string(self.log())
            .unwrap_or_default()
            .lines()
            .map(String::from)
            .collect()
    }

    fn run(&self, session: Option<&str>, args: &[&str]) -> Output {
        let mut cmd: Command = self.home.command(args);
        cmd.env("MOAI_GIT_BIN", &self.git_bin);
        cmd.env("BATCH_LOG", self.log());
        if let Some(id) = session {
            cmd.env(SESSION_VAR, id);
        }
        cmd.stdin(Stdio::null());
        cmd.output().expect("run moai")
    }

    fn card(&self, session: Option<&str>, args: &[&str]) -> Output {
        let cfg = self.project.join(".moai-cli/project.yaml");
        fs::create_dir_all(cfg.parent().unwrap()).unwrap();
        fs::write(&cfg, PROFILE).unwrap();
        let mut all = vec!["card"];
        all.extend_from_slice(args);
        all.extend_from_slice(&["--project", &self.p]);
        self.run(session, &all)
    }

    /// A card with its own scope, taken by the lane and sent to review; returns its name.
    fn review_card(&self, scope: &str) -> String {
        let added = result(&self.card(
            None,
            &[
                "add",
                "--title",
                "fixture card",
                "--kind",
                "task",
                "--scope",
                scope,
                "--ac",
                "one",
            ],
        ));
        let name = added["name"].as_str().unwrap().to_string();
        self.link_spec(&name, SPEC);
        result(&self.card(None, &["ready", &name]));
        result(&self.card(Some(S_A), &["take", &name]));
        result(&self.card(Some(S_A), &["review", &name, "--generation", "1"]));
        name
    }

    /// The SPEC link of `card`, written as `card spec` writes it (the event, without the SPEC
    /// file the command wants on disk).
    fn link_spec(&self, card: &str, spec: &str) {
        let conn =
            Connection::open(ledger::ledger_path(self.home.path(), "demo").unwrap()).unwrap();
        let rows: i64 = conn
            .query_row("SELECT count(*) FROM event", [], |r| r.get(0))
            .unwrap();
        conn.execute(
            "INSERT INTO event (event_key, at, actor, card_id, kind, payload) \
             VALUES (?1, unixepoch(), ?2, ?3, 'card.spec', ?4)",
            rusqlite::params![
                format!("test-spec-{rows}"),
                S_LEAD,
                card.trim_start_matches("C-").parse::<i64>().unwrap(),
                json!({"from": null, "to": spec}).to_string()
            ],
        )
        .unwrap();
    }

    /// A lane head on the base commit that adds `path`.
    fn lane(&self, path: &str, content: &str) -> String {
        put_on(
            &self.project,
            &self.c0,
            path,
            content,
            &signed("the lane change", S_A),
        )
    }

    /// A PASS of `head` by the independent reviewer, written as `card verdict` writes it.
    fn pass(&self, card: &str, head: &str) {
        let conn =
            Connection::open(ledger::ledger_path(self.home.path(), "demo").unwrap()).unwrap();
        let rows: i64 = conn
            .query_row("SELECT count(*) FROM event", [], |r| r.get(0))
            .unwrap();
        let payload = json!({
            "head": head, "base": self.c0, "verdict": "pass", "tier": 0,
            "evidence": "the fixture's evidence", "receipt": null, "reviewer": S_AUD,
        });
        conn.execute(
            "INSERT INTO event (event_key, at, actor, card_id, kind, payload) \
             VALUES (?1, unixepoch(), ?2, ?3, 'card.verdict', ?4)",
            rusqlite::params![
                format!("test-verdict-{rows}"),
                S_AUD,
                card.trim_start_matches("C-").parse::<i64>().unwrap(),
                payload.to_string()
            ],
        )
        .unwrap();
    }

    /// `card land-batch` as the leader with the check script at `check`.
    fn batch(&self, pairs: &[(&str, &str)], check: &Path) -> Output {
        let joined: Vec<String> = pairs.iter().map(|(c, h)| format!("{c}={h}")).collect();
        let mut args = vec!["land-batch"];
        args.extend(joined.iter().map(String::as_str));
        args.extend(["--by", "leader", "--check-cmd", check.to_str().unwrap()]);
        self.card(Some(S_LEAD), &args)
    }

    fn integ(&self) -> String {
        git(&self.project, &["rev-parse", "refs/heads/integ"])
    }

    fn state(&self, card: &str) -> String {
        let l = ledger::open_read_only(self.home.path(), "demo").unwrap();
        l.conn()
            .query_row(
                "SELECT state FROM card WHERE card_id = ?1",
                [card.trim_start_matches("C-").parse::<i64>().unwrap()],
                |r| r.get(0),
            )
            .unwrap()
    }

    fn is_ancestor(&self, head: &str, of: &str) -> bool {
        common::git::git_raw(&self.project, &["merge-base", "--is-ancestor", head, of]).success
    }

    /// Three cards with their lane heads, each added and passed; `poison` marks the head (by
    /// index) that carries the file the check script fails on.
    fn three(&self, poison: Option<usize>) -> Vec<(String, String)> {
        (0..3)
            .map(|n| {
                let card = self.review_card(&format!("apps/{n}/"));
                let (path, content) = if poison == Some(n) {
                    ("notes/poison.md".to_string(), "bad\n")
                } else {
                    (format!("notes/{n}.md"), "ok\n")
                };
                let head = self.lane(&path, content);
                self.pass(&card, &head);
                (card, head)
            })
            .collect()
    }
}

fn pairs(v: &[(String, String)]) -> Vec<(&str, &str)> {
    v.iter().map(|(c, h)| (c.as_str(), h.as_str())).collect()
}

fn names(v: &Value) -> Vec<String> {
    v.as_array()
        .unwrap()
        .iter()
        .map(|e| e["card"].as_str().unwrap().to_string())
        .collect()
}

#[test]
fn batch_three_passing_heads_run_one_check_and_land_in_order() {
    let w = World::new();
    let cards = w.three(None);
    let check = w.script("check.sh", CHECK);
    let r = result(&w.batch(&pairs(&cards), &check));
    assert_eq!(
        names(&r["landed"]),
        cards.iter().map(|(c, _)| c.clone()).collect::<Vec<_>>(),
        "{r}"
    );
    assert!(r["dropped"].as_array().unwrap().is_empty(), "{r}");
    assert!(r["not_run"].as_array().unwrap().is_empty(), "{r}");
    assert!(r["stopped"].is_null(), "{r}");
    // One run for the three heads, on the range from the tip to the last join.
    assert_eq!(r["checks_run"], 1, "{r}");
    let ranges = w.ranges();
    assert_eq!(ranges.len(), 1, "{ranges:?}");
    assert_eq!(ranges[0], format!("{}..{}", w.c0, w.integ()));
    // Every head and card landed; the first head stands on the tip as it is, the others joined.
    for (card, head) in &cards {
        assert!(w.is_ancestor(head, &w.integ()), "{head}");
        assert_eq!(w.state(card), "done", "{card}");
    }
    assert_eq!(r["landed"][0]["candidate"], json!(cards[0].1), "{r}");
    let join = git(&w.project, &["cat-file", "-p", &w.integ()]);
    assert!(join.contains("Moai-Verified: none"), "{join}");
    assert!(join.contains(&format!("parent {}", cards[2].1)), "{join}");
}

#[test]
fn batch_a_tip_that_moved_during_the_check_is_not_reused() {
    let w = World::new();
    let cards = w.three(None);
    // Someone else lands `x` on the integration branch while the first check runs.
    let x = w.lane("notes/x.md", "x\n");
    let script = format!(
        "#!/bin/sh\necho \"$1\" >> \"$BATCH_LOG\"\nif [ ! -e \"$BATCH_LOG.moved\" ]; then\n  : > \"$BATCH_LOG.moved\"\n  \"$MOAI_GIT_BIN\" -C '{}' update-ref refs/heads/integ {x}\nfi\nexit 0\n",
        w.p
    );
    let check = w.script("moving.sh", &script);
    let r = result(&w.batch(&pairs(&cards), &check));
    assert_eq!(names(&r["landed"]).len(), 3, "{r}");
    let ranges = w.ranges();
    // The first result belongs to the old tip and lands nothing; the second starts from `x`.
    assert_eq!(ranges.len(), 2, "{ranges:?}");
    assert!(ranges[0].starts_with(&w.c0), "{ranges:?}");
    assert!(ranges[1].starts_with(&x), "{ranges:?}");
    assert!(w.is_ancestor(&x, &w.integ()));
    for (_, head) in &cards {
        assert!(w.is_ancestor(head, &w.integ()), "{head}");
    }
}

#[test]
fn split_a_failing_head_is_found_by_halving_and_the_others_land() {
    let w = World::new();
    let cards = w.three(Some(1));
    let check = w.script("check.sh", CHECK);
    let r = result(&w.batch(&pairs(&cards), &check));
    assert_eq!(
        names(&r["landed"]),
        vec![cards[0].0.clone(), cards[2].0.clone()],
        "{r}"
    );
    assert_eq!(names(&r["dropped"]), vec![cards[1].0.clone()], "{r}");
    assert!(
        r["dropped"][0]["reason"]
            .as_str()
            .unwrap()
            .starts_with("check_failed"),
        "{r}"
    );
    assert_eq!(w.state(&cards[0].0), "done");
    assert_eq!(w.state(&cards[2].0), "done");
    assert_ne!(w.state(&cards[1].0), "done");
    assert!(!w.is_ancestor(&cards[1].1, &w.integ()));
    // The whole set, its first half, the first head alone, then the rest on the new tip: the
    // first head landed, so the later runs start from it and never from the old tip (REQ-9).
    let ranges = w.ranges();
    assert_eq!(ranges.len(), 6, "{ranges:?}");
    assert!(
        ranges[..3].iter().all(|r| r.starts_with(&w.c0)),
        "{ranges:?}"
    );
    assert!(
        ranges[3..].iter().all(|r| r.starts_with(&cards[0].1)),
        "{ranges:?}"
    );
}

#[test]
fn split_a_passing_first_half_lands_before_the_rest_is_checked() {
    let w = World::new();
    let cards = w.three(Some(2));
    let check = w.script("check.sh", CHECK);
    let r = result(&w.batch(&pairs(&cards), &check));
    assert_eq!(
        names(&r["landed"]),
        vec![cards[0].0.clone(), cards[1].0.clone()],
        "{r}"
    );
    assert_eq!(names(&r["dropped"]), vec![cards[2].0.clone()], "{r}");
    let ranges = w.ranges();
    // {1,2,3} fails, {1,2} passes and lands, {3} runs on the new tip and fails.
    assert_eq!(ranges.len(), 3, "{ranges:?}");
    assert!(
        ranges[0].starts_with(&w.c0) && ranges[1].starts_with(&w.c0),
        "{ranges:?}"
    );
    assert!(!ranges[2].starts_with(&w.c0), "{ranges:?}");
    assert!(ranges[2].starts_with(&w.integ()), "{ranges:?}");
}

#[test]
fn refuse_a_head_that_conflicts_is_dropped_and_the_others_land() {
    let w = World::new();
    let c1 = w.review_card("apps/1/");
    let c2 = w.review_card("apps/2/");
    let c3 = w.review_card("apps/3/");
    // Both heads add the same new file with different text: an add/add conflict at tier 0.
    let h1 = w.lane("notes/shared.md", "one\n");
    let h2 = w.lane("notes/shared.md", "two\n");
    let h3 = w.lane("notes/3.md", "ok\n");
    for (c, h) in [(&c1, &h1), (&c2, &h2), (&c3, &h3)] {
        w.pass(c, h);
    }
    let check = w.script("check.sh", CHECK);
    let r = result(&w.batch(&[(&c1, &h1), (&c2, &h2), (&c3, &h3)], &check));
    assert_eq!(names(&r["landed"]), vec![c1.clone(), c3.clone()], "{r}");
    assert_eq!(names(&r["dropped"]), vec![c2.clone()], "{r}");
    assert!(
        r["dropped"][0]["reason"]
            .as_str()
            .unwrap()
            .starts_with("join_failed"),
        "{r}"
    );
    assert_ne!(w.state(&c2), "done");
}

#[test]
fn refuse_a_check_that_cannot_start_lands_nothing() {
    let w = World::new();
    let cards = w.three(None);
    let before = w.integ();
    let r = result(&w.batch(&pairs(&cards), &w.scripts.path().join("missing.sh")));
    assert!(r["landed"].as_array().unwrap().is_empty(), "{r}");
    assert_eq!(r["stopped"]["code"], "check_not_started", "{r}");
    assert_eq!(r["not_run"].as_array().unwrap().len(), 3, "{r}");
    assert_eq!(w.integ(), before);
    for (card, _) in &cards {
        assert_ne!(w.state(card), "done");
    }
    // The private worktree of the run is gone again.
    let listed = git(&w.project, &["worktree", "list", "--porcelain"]);
    assert!(!listed.contains("land-batch"), "{listed}");
}

#[test]
fn refuse_a_head_with_no_verdict_is_dropped_before_any_head_lands_and_the_rest_is_rechecked() {
    let w = World::new();
    let c1 = w.review_card("apps/1/");
    let c2 = w.review_card("apps/2/");
    let c3 = w.review_card("apps/3/");
    let h1 = w.lane("notes/1.md", "ok\n");
    let h2 = w.lane("notes/2.md", "ok\n");
    let h3 = w.lane("notes/3.md", "ok\n");
    // The second head has no verdict: the landing closes no card without a PASS.
    w.pass(&c1, &h1);
    w.pass(&c3, &h3);
    let check = w.script("check.sh", CHECK);
    let r = result(&w.batch(&[(&c1, &h1), (&c2, &h2), (&c3, &h3)], &check));
    // The refusal is found before the first head lands (N2): the branch never stops at a join of
    // the first heads that no check ran on, and the heads that can land are joined without it and
    // checked again.
    assert_eq!(names(&r["dropped"]), vec![c2.clone()], "{r}");
    assert!(
        r["dropped"][0]["reason"]
            .as_str()
            .unwrap()
            .starts_with("land_refused: card land refused: no_pass"),
        "{r}"
    );
    assert_eq!(names(&r["landed"]), vec![c1.clone(), c3.clone()], "{r}");
    assert!(r["stopped"].is_null(), "{r}");
    assert_eq!(w.state(&c1), "done");
    assert_eq!(w.state(&c3), "done");
    assert_ne!(w.state(&c2), "done");
    // Every check ran on a join that holds exactly the heads that then landed: the second run is
    // the one of the two that remained, and the branch moved only to joins of that run.
    let ranges = w.ranges();
    assert_eq!(ranges.len(), 2, "{ranges:?}");
    let last = ranges[1].split("..").nth(1).unwrap().to_string();
    assert!(
        w.is_ancestor(&h1, &last) && w.is_ancestor(&h3, &last),
        "{ranges:?}"
    );
    assert!(!w.is_ancestor(&h2, &w.integ()));
    assert_eq!(w.integ(), last);
}

impl World {
    /// A `git` that is the real one except that `commit-tree` of a join commit (`-p <cursor> -p
    /// <head>`) first runs `rewrite`, a shell fragment that may set `FIRST` and `SECOND` (the two
    /// parents) and `TREE`. It makes the join candidates the batch itself never makes.
    fn join_rewriter(&mut self, rewrite: &str) {
        let real = self.git_bin.to_str().unwrap().to_string();
        let text = format!(
            "#!/bin/sh\nREAL='{real}'\nif [ \"$4\" = commit-tree ] && [ \"$6\" = -p ] && [ \"$8\" = -p ]; then\n  TREE=\"$5\"; FIRST=\"$7\"; SECOND=\"$9\"; DIR=\"$3\"\n  {rewrite}\n  exec \"$REAL\" \"$1\" \"$2\" \"$3\" commit-tree \"$TREE\" -p \"$FIRST\" -p \"$SECOND\" \"${{10}}\" \"${{11}}\"\nfi\nexec \"$REAL\" \"$@\"\n"
        );
        self.git_bin = self.script("git-rewrite.sh", &text);
    }

    /// Two cards whose heads sit on the base, each with its PASS; the first head stands on the tip
    /// and lands as it is, the second is joined.
    fn two(&self) -> Vec<(String, String)> {
        (0..2)
            .map(|n| {
                let card = self.review_card(&format!("apps/{n}/"));
                let head = self.lane(&format!("notes/{n}.md"), "ok\n");
                self.pass(&card, &head);
                (card, head)
            })
            .collect()
    }
}

/// The reason a dropped head carries, by card name.
fn reason_of<'a>(r: &'a Value, card: &str) -> &'a str {
    r["dropped"]
        .as_array()
        .unwrap()
        .iter()
        .find(|e| e["card"] == card)
        .unwrap_or_else(|| panic!("{card} is not dropped: {r}"))["reason"]
        .as_str()
        .unwrap()
}

/// D-5 pinned through the command (audit N9): a join whose second parent is not the audited head
/// is refused by the landing, so the verdict of that head covers nothing else.
#[test]
fn refuse_a_join_whose_second_parent_is_not_the_audited_head() {
    let mut w = World::new();
    let cards = w.two();
    // The parents swapped: the audited head becomes the first parent of its own join.
    w.join_rewriter("T=\"$FIRST\"; FIRST=\"$SECOND\"; SECOND=\"$T\"");
    let check = w.script("check.sh", CHECK);
    let r = result(&w.batch(&pairs(&cards), &check));
    let (c1, h1) = (&cards[0].0, &cards[0].1);
    let c2 = &cards[1].0;
    assert_eq!(names(&r["landed"]), vec![c1.clone()], "{r}");
    assert!(
        reason_of(&r, c2).contains("is not the second parent of a two-parent candidate"),
        "{r}"
    );
    assert_ne!(w.state(c2), "done");
    assert_eq!(&w.integ(), h1);
}

/// Audit N1: a join candidate lands only from the tip it was made on. A first parent that is
/// beyond the tip, with the same tree, carries a commit nobody landed.
#[test]
fn refuse_a_join_whose_first_parent_is_not_the_tip() {
    let mut w = World::new();
    let cards = w.two();
    w.join_rewriter(
        "FIRST=$(\"$REAL\" -C \"$DIR\" commit-tree \"$(\"$REAL\" -C \"$DIR\" rev-parse \"$FIRST^{tree}\")\" -p \"$FIRST\" -m beyond)",
    );
    let check = w.script("check.sh", CHECK);
    let r = result(&w.batch(&pairs(&cards), &check));
    let (c1, h1) = (&cards[0].0, &cards[0].1);
    let c2 = &cards[1].0;
    assert_eq!(names(&r["landed"]), vec![c1.clone()], "{r}");
    assert!(reason_of(&r, c2).contains("not the integration tip"), "{r}");
    assert_ne!(w.state(c2), "done");
    assert_eq!(&w.integ(), h1);
}

/// Audit N5: a join commit is named for the SPEC, carries `Moai-Spec`, and names no card.
#[test]
fn a_join_commit_names_the_spec_and_no_card_number() {
    let w = World::new();
    let cards = w.three(None);
    let check = w.script("check.sh", CHECK);
    let r = result(&w.batch(&pairs(&cards), &check));
    assert_eq!(names(&r["landed"]).len(), 3, "{r}");
    let log = git(
        &w.project,
        &["log", "--merges", "--format=%B%x1e", "refs/heads/integ"],
    );
    let messages: Vec<&str> = log
        .split('\x1e')
        .map(str::trim)
        .filter(|m| !m.is_empty())
        .collect();
    assert_eq!(messages.len(), 2, "{log}");
    for m in messages {
        assert!(
            m.starts_with(&format!("merge(next-flow): {SPEC} into integ, head ")),
            "{m}"
        );
        assert!(m.contains(&format!("\nMoai-Spec: {SPEC}\n")), "{m}");
        // The SPEC id holds "C-2" inside "SPEC-2026…", so look for a card number outside it.
        let without_spec = m.replace(SPEC, "");
        assert!(
            !without_spec.contains("C-1")
                && !without_spec.contains("C-2")
                && !without_spec.contains("C-3"),
            "{m}"
        );
    }
}

#[test]
fn refuse_a_call_that_is_not_two_to_four_pairs_by_the_leader() {
    let w = World::new();
    let cards = w.three(None);
    let check = w.script("check.sh", CHECK);
    let before = w.integ();
    let one = w.batch(&pairs(&cards)[..1], &check);
    assert_eq!(
        one.status.code(),
        Some(1),
        "{}",
        String::from_utf8_lossy(&one.stdout)
    );
    let repeated = w.batch(&[pairs(&cards)[0], pairs(&cards)[0]], &check);
    assert_eq!(repeated.status.code(), Some(1));
    let args = [
        "land-batch",
        &format!("{}={}", cards[0].0, cards[0].1),
        &format!("{}={}", cards[1].0, cards[1].1),
        "--by",
        "agent-1",
        "--check-cmd",
        check.to_str().unwrap(),
    ];
    let not_leader = w.card(Some(S_LEAD), &args);
    assert_eq!(not_leader.status.code(), Some(1));
    assert_eq!(w.integ(), before);
    assert!(w.ranges().is_empty());
}
