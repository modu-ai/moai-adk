//! A landing keeps open the criteria a verdict holds open (SPEC-20261126-land-open-criteria):
//! `card verdict --open <n>` records the hold, and `card land` (also inside `card land-batch`)
//! lands such a head as a slice. Every test runs on a temporary home and a temporary repository
//! whose integration branch is `integ`; nothing here touches the real ledger.
#![cfg(unix)]

use crate::common;
use common::home::TempHome;
use moai_cli::ledger;
use rusqlite::Connection;
use serde_json::{json, Value};
use std::fs;
use std::os::unix::fs::PermissionsExt;
use std::path::{Path, PathBuf};
use std::process::{Command, Output, Stdio};

const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
/// The lane's author and card holder, two independent reviewers, the lander (the leader).
const S_A: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
const S_AUD: &str = "cccccccc-0000-4000-8000-00000000000c";
const S_AUD2: &str = "eeeeeeee-0000-4000-8000-00000000000e";
const S_LEAD: &str = "dddddddd-0000-4000-8000-00000000000d";
const WORKTREE_KEY: &str = "0123456789ab";
/// The SPEC every fixture card names; the batch's join commits are named for it.
const SPEC: &str = "SPEC-20261124-fixture";
const PROFILE: &str = r#"{"schema_version": 1, "project_id": "demo", "purpose": "fixture",
    "tools": ["rust"], "development_methods": ["sdd"], "hosts": ["claude"],
    "work_flow": {"flow": "m-flow", "integration_branch": "integ"}}"#;

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

/// Asserts exit 0 and a success envelope, and unwraps the result.
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

/// Asserts exit 1 and a refusal; returns its message.
fn refusal(out: &Output) -> String {
    assert_eq!(
        out.status.code(),
        Some(1),
        "{} {}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    let v = envelope(out);
    assert_eq!(v["ok"], json!(false), "{v}");
    v["error"]["message"].as_str().unwrap().to_string()
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

fn number(card: &str) -> i64 {
    card.trim_start_matches("C-").parse().unwrap()
}

struct World {
    home: TempHome,
    project: PathBuf,
    p: String,
    scripts: tempfile::TempDir,
    c0: String,
}

impl World {
    /// A repository whose first commit holds the fixture SPEC, `integ` at that commit and the
    /// leader session registered.
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
        // A root `Cargo.toml` makes the fixture a Rust repository, as the batch fixture does.
        fs::write(project.join("Cargo.toml"), "[workspace]\n").unwrap();
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
        // The command answers where it wrote the SPEC, so the test spells no state folder.
        let spec_rel = result(&made)["path"].as_str().unwrap().to_string();
        let spec_file = project.join(&spec_rel);
        fs::write(&spec_file, filled(&fs::read_to_string(&spec_file).unwrap())).unwrap();
        git(&project, &["add", "a.txt", "Cargo.toml", &spec_rel]);
        git(&project, &["commit", "-q", "-m", "init"]);
        let c0 = git(&project, &["rev-parse", "HEAD"]);
        git(&project, &["update-ref", "refs/heads/integ", &c0]);
        let w = World {
            p: project.to_str().unwrap().to_string(),
            home,
            project,
            scripts: tempfile::tempdir().unwrap(),
            c0,
        };
        result_ok(&w.run(
            Some(S_LEAD),
            &[
                "session-register",
                &w.p,
                "--name",
                "leader",
                "--mode",
                "leader",
            ],
            None,
        ));
        w
    }

    fn run(&self, session: Option<&str>, args: &[&str], check: Option<&Path>) -> Output {
        let mut cmd: Command = self.home.command(args);
        cmd.env("MOAI_GIT_BIN", git_executable());
        if let Some(program) = check {
            cmd.env("MOAI_TEST_LAND_BATCH_CHECK", program);
        }
        if let Some(id) = session {
            cmd.env(SESSION_VAR, id);
        }
        cmd.stdin(Stdio::null());
        cmd.output().expect("run moai")
    }

    fn card_with(&self, session: Option<&str>, args: &[&str], check: Option<&Path>) -> Output {
        let cfg = moai_cli::state_root::path(&self.project, "project.yaml");
        fs::create_dir_all(cfg.parent().unwrap()).unwrap();
        fs::write(&cfg, PROFILE).unwrap();
        let mut all = vec!["card"];
        all.extend_from_slice(args);
        all.extend_from_slice(&["--project", &self.p]);
        self.run(session, &all, check)
    }

    fn card(&self, session: Option<&str>, args: &[&str]) -> Output {
        self.card_with(session, args, None)
    }

    /// A task card with `criteria` open criteria, taken by `S_A`, still `active`; its name.
    fn active_card(&self, scope: &str, criteria: usize) -> String {
        let mut args = vec![
            "add",
            "--title",
            "fixture card",
            "--kind",
            "task",
            "--scope",
            scope,
        ];
        let words = ["one", "two", "three"];
        for word in &words[..criteria] {
            args.extend(["--ac", word]);
        }
        let name = result(&self.card(None, &args))["name"]
            .as_str()
            .unwrap()
            .to_string();
        self.link_spec(&name);
        result(&self.card(None, &["ready", &name]));
        // A temporary path that is no worktree: left out, the ledger would record the folder the
        // test runs in.
        let tree = self.scripts.path().join(&name);
        let tree = tree.to_str().unwrap();
        result(&self.card(Some(S_A), &["take", &name, "--worktree", tree]));
        name
    }

    fn to_review(&self, card: &str) {
        result(&self.card(Some(S_A), &["review", card, "--generation", "1"]));
    }

    /// A card in `review` with `criteria` open criteria.
    fn review_card(&self, scope: &str, criteria: usize) -> String {
        let card = self.active_card(scope, criteria);
        self.to_review(&card);
        card
    }

    fn met(&self, card: &str, seq: i64, evidence: &str) {
        let seq = seq.to_string();
        result(&self.card(
            Some(S_A),
            &[
                "criterion",
                card,
                &seq,
                "--generation",
                "1",
                "--met",
                "--evidence",
                evidence,
            ],
        ));
    }

    /// The SPEC link of `card`, written as `card spec` writes it.
    fn link_spec(&self, card: &str) {
        self.insert_event(
            card,
            S_LEAD,
            "card.spec",
            &json!({"from": null, "to": SPEC}),
        );
    }

    fn insert_event(&self, card: &str, actor: &str, kind: &str, payload: &Value) {
        let conn =
            Connection::open(ledger::ledger_path(self.home.path(), "demo").unwrap()).unwrap();
        let rows: i64 = conn
            .query_row("SELECT count(*) FROM event", [], |r| r.get(0))
            .unwrap();
        conn.execute(
            "INSERT INTO event (event_key, at, actor, card_id, kind, payload) \
             VALUES (?1, unixepoch(), ?2, ?3, ?4, ?5)",
            rusqlite::params![
                format!("test-event-{rows}"),
                actor,
                number(card),
                kind,
                payload.to_string()
            ],
        )
        .unwrap();
    }

    /// A lane head on the base commit that adds `path`; no ref moves.
    fn lane(&self, path: &str) -> String {
        let blob_file = self.project.join("lane-blob.tmp");
        fs::write(&blob_file, "lane\n").unwrap();
        let blob = git(&self.project, &["hash-object", "-w", "lane-blob.tmp"]);
        fs::remove_file(&blob_file).unwrap();
        let index = self.project.join(".git/lane-index");
        let with_index = |args: &[&str]| -> String {
            let out = common::git::command()
                .arg("-C")
                .arg(&self.project)
                .args(args)
                .env("GIT_INDEX_FILE", &index)
                .output()
                .unwrap();
            assert!(out.status.success(), "git {args:?}");
            String::from_utf8_lossy(&out.stdout).trim().to_string()
        };
        with_index(&["read-tree", &self.c0]);
        with_index(&[
            "update-index",
            "--add",
            "--cacheinfo",
            &format!("100644,{blob},{path}"),
        ]);
        let tree = with_index(&["write-tree"]);
        fs::remove_file(&index).unwrap();
        let message = format!(
            "the lane change\n\nMoai-Spec: {SPEC}\nMoai-Session: {S_A}:{WORKTREE_KEY}\nMoai-Verified: none"
        );
        git(
            &self.project,
            &["commit-tree", &tree, "-p", &self.c0, "-m", &message],
        )
    }

    /// A lane head and the clean merge of it onto a moved integration tip, which `integ` points
    /// at: landing the merge is a fast move whose second parent is the audited head.
    fn merge_on_a_moved_tip(&self) -> (String, String) {
        let audited = self.lane("notes/a.md");
        let tip = self.lane("notes/t.md");
        git(&self.project, &["update-ref", "refs/heads/integ", &tip]);
        let tree = git(
            &self.project,
            &["merge-tree", "--write-tree", &tip, &audited],
        );
        let merge = git(
            &self.project,
            &[
                "commit-tree",
                &tree,
                "-p",
                &tip,
                "-p",
                &audited,
                "-m",
                "merge",
            ],
        );
        (audited, merge)
    }

    /// `card verdict` of `head` by `reviewer` with the `--open` numbers.
    fn verdict(&self, card: &str, head: &str, reviewer: &str, word: &str, open: &[i64]) -> Output {
        let numbers: Vec<String> = open.iter().map(i64::to_string).collect();
        let mut args = vec![
            "verdict",
            card,
            "--head",
            head,
            "--base",
            &self.c0,
            "--verdict",
            word,
            "--tier",
            "0",
            "--evidence",
            "the fixture's evidence",
        ];
        for n in &numbers {
            args.extend(["--open", n]);
        }
        self.card(Some(reviewer), &args)
    }

    fn pass(&self, card: &str, head: &str, reviewer: &str, open: &[i64]) -> i64 {
        result(&self.verdict(card, head, reviewer, "pass", open))["event"]
            .as_i64()
            .unwrap()
    }

    fn land(&self, card: &str, head: &str, extra: &[&str]) -> Output {
        let mut args = vec!["land", card, "--head", head, "--by", "leader"];
        args.extend_from_slice(extra);
        self.card(Some(S_LEAD), &args)
    }

    fn integ(&self) -> String {
        git(&self.project, &["rev-parse", "refs/heads/integ"])
    }

    fn query<T: rusqlite::types::FromSql>(&self, sql: &str) -> T {
        let l = ledger::open_read_only(self.home.path(), "demo").unwrap();
        l.conn().query_row(sql, [], |r| r.get(0)).unwrap()
    }

    fn state(&self, card: &str) -> String {
        self.query(&format!(
            "SELECT state FROM card WHERE card_id = {}",
            number(card)
        ))
    }

    /// `(status, evidence_ref)` of every criterion of the card, in order.
    fn criteria(&self, card: &str) -> Vec<(String, Option<String>)> {
        let l = ledger::open_read_only(self.home.path(), "demo").unwrap();
        let mut st = l
            .conn()
            .prepare("SELECT status, evidence_ref FROM card_ac WHERE card_id = ?1 ORDER BY seq")
            .unwrap();
        st.query_map([number(card)], |r| Ok((r.get(0)?, r.get(1)?)))
            .unwrap()
            .map(Result::unwrap)
            .collect()
    }

    fn statuses(&self, card: &str) -> Vec<String> {
        self.criteria(card).into_iter().map(|(s, _)| s).collect()
    }

    fn events(&self, card: &str, kind: &str) -> i64 {
        self.query(&format!(
            "SELECT count(*) FROM event WHERE card_id = {} AND kind = '{kind}'",
            number(card)
        ))
    }

    /// The payload of the newest event of `kind` on the card.
    fn payload(&self, card: &str, kind: &str) -> Value {
        let text: String = self.query(&format!(
            "SELECT payload FROM event WHERE card_id = {} AND kind = '{kind}' \
             ORDER BY event_id DESC LIMIT 1",
            number(card)
        ));
        serde_json::from_str(&text).unwrap()
    }
}

/// Exit 0 of a command whose answer is not a card envelope.
fn result_ok(out: &Output) {
    assert_eq!(
        out.status.code(),
        Some(0),
        "{} {}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
}

/// AC-1 (REQ-3, REQ-4): a PASS that holds criterion 2 lands the head as a slice: the branch
/// moves, the card stays open, no criterion is marked and the answer names the hold.
#[test]
fn a_held_criterion_keeps_the_card_open_and_marks_nothing() {
    let w = World::new();
    let card = w.review_card("apps/a/", 2);
    let head = w.lane("notes/a.md");
    let event = w.pass(&card, &head, S_AUD, &[2]);

    let got = result(&w.land(&card, &head, &[]));
    assert_eq!(got["moved"], json!(true), "{got}");
    assert_eq!(got["closed"], json!(false), "{got}");
    assert_eq!(got["open_criteria"], json!([1, 2]), "{got}");
    assert_eq!(got["met"], json!([]), "{got}");
    assert_eq!(
        got["held"],
        json!([{"seq": 2, "reviewer": S_AUD, "event": event}]),
        "{got}"
    );
    assert_eq!(w.integ(), head);
    assert_ne!(w.state(&card), "done");
    assert_eq!(w.statuses(&card), ["open", "open"]);
    let landed = w.payload(&card, "card.landed");
    assert_eq!(landed["keep_open"], json!(true), "{landed}");
    assert_eq!(w.events(&card, "card.finish"), 0);
}

/// AC-2 (REQ-5): a held criterion the author marked met is open again after the landing, with the
/// replaced evidence on the landing event; the other criterion stays met and the card stays open.
#[test]
fn a_held_criterion_the_author_marked_met_is_opened_again() {
    for extra in [&[][..], &["--keep-open"][..]] {
        let w = World::new();
        let card = w.active_card("apps/a/", 2);
        w.met(&card, 1, "run one");
        w.met(&card, 2, "run two");
        w.to_review(&card);
        let head = w.lane("notes/a.md");
        w.pass(&card, &head, S_AUD, &[2]);

        let got = result(&w.land(&card, &head, extra));
        assert_eq!(got["closed"], json!(false), "{extra:?}: {got}");
        assert_eq!(got["open_criteria"], json!([2]), "{extra:?}: {got}");
        assert_ne!(w.state(&card), "done", "{extra:?}");
        assert_eq!(
            w.criteria(&card),
            [
                ("met".to_string(), Some("run one".to_string())),
                ("open".to_string(), None)
            ],
            "{extra:?}"
        );
        let landed = w.payload(&card, "card.landed");
        assert_eq!(
            landed["reopened"],
            json!([{"seq": 2, "evidence": "run two"}]),
            "{extra:?}: {landed}"
        );

        // A rerun of the same head while the hold stands closes nothing and fills nothing.
        let again = result(&w.land(&card, &head, &[]));
        assert_eq!(again["closed"], json!(false), "{extra:?}: {again}");
        assert_ne!(w.state(&card), "done", "{extra:?}");
        assert_eq!(w.statuses(&card), ["met", "open"], "{extra:?}");
    }
}

/// AC-3 (REQ-1, REQ-2, REQ-8): the verdict stores only a hold it can keep.
#[test]
fn card_verdict_stores_only_a_hold_it_can_keep() {
    let w = World::new();
    let card = w.active_card("apps/a/", 3);
    result(&w.card(
        Some(S_A),
        &[
            "criterion",
            &card,
            "3",
            "--generation",
            "1",
            "--withdraw",
            "--reason",
            "no longer needed",
        ],
    ));
    w.to_review(&card);
    let head = w.lane("notes/a.md");

    for (word, open, needle) in [
        ("fail", &[1][..], "pass"),
        ("pass", &[4][..], "criterion 4"),
        ("pass", &[3][..], "withdrawn"),
    ] {
        let why = refusal(&w.verdict(&card, &head, S_AUD, word, open));
        assert!(why.contains(needle), "{word} {open:?}: {why}");
        assert_eq!(w.events(&card, "card.verdict"), 0, "{word} {open:?}");
    }

    let first = result(&w.verdict(&card, &head, S_AUD, "pass", &[2, 1, 2]));
    assert_eq!(first["open"], json!([1, 2]), "{first}");
    assert_eq!(w.payload(&card, "card.verdict")["open"], json!([1, 2]));

    let plain = result(&w.verdict(&card, &head, S_AUD2, "pass", &[]));
    assert!(plain.get("open").is_none(), "{plain}");
    let payload = w.payload(&card, "card.verdict");
    assert!(payload.get("open").is_none(), "{payload}");
    let repeat = result(&w.verdict(&card, &head, S_AUD2, "pass", &[]));
    assert_eq!(repeat["written"], json!(false), "{repeat}");
    assert_eq!(repeat["event"], plain["event"], "{repeat}");

    // 0 is no criterion number: the command line refuses it before the ledger is read.
    let zero = w.verdict(&card, &head, S_AUD, "pass", &[0]);
    assert_eq!(zero.status.code(), Some(2));
    assert_eq!(w.events(&card, "card.verdict"), 2);

    // A stored `open` that is no list of positive numbers is no verdict: the head has no PASS.
    for stored in [json!("1"), json!(["1"]), json!([1.5]), json!([0])] {
        let w = World::new();
        let card = w.review_card("apps/a/", 1);
        let head = w.lane("notes/a.md");
        w.insert_event(
            &card,
            S_AUD,
            "card.verdict",
            &json!({"head": head, "base": w.c0, "verdict": "pass", "tier": 0,
                    "evidence": "hand written", "receipt": null, "reviewer": S_AUD,
                    "open": stored}),
        );
        let why = refusal(&w.land(&card, &head, &[]));
        assert!(why.contains("no_pass"), "{stored}: {why}");
        assert_eq!(w.integ(), w.c0, "{stored}");
    }
}

/// REQ-3: a verdict recorded after a held landing, on the landed merge or on the audited head,
/// does not drop the hold on a rerun; a newer PASS of the held head without `--open` releases it.
#[test]
fn a_rerun_after_a_later_verdict_keeps_the_hold() {
    for (on_merge, word) in [(true, "fail"), (true, "pass"), (false, "fail")] {
        let case = format!("on_merge={on_merge} {word}");
        let w = World::new();
        let card = w.review_card("apps/a/", 2);
        let (audited, merge) = w.merge_on_a_moved_tip();
        let held_event = w.pass(&card, &audited, S_AUD, &[2]);
        let held = json!([{"seq": 2, "reviewer": S_AUD, "event": held_event}]);

        let first = result(&w.land(&card, &merge, &[]));
        assert_eq!(first["held"], held, "{case}: {first}");
        assert_eq!(first["closed"], json!(false), "{case}: {first}");
        assert_eq!(w.integ(), merge, "{case}");

        let later = if on_merge { &merge } else { &audited };
        result(&w.verdict(&card, later, S_AUD, word, &[]));

        let again = result(&w.land(&card, &merge, &[]));
        assert_eq!(again["closed"], json!(false), "{case}: {again}");
        assert_eq!(again["held"], held, "{case}: {again}");
        assert_eq!(again["met"], json!([]), "{case}: {again}");
        assert_ne!(w.state(&card), "done", "{case}");
        assert_eq!(w.statuses(&card), ["open", "open"], "{case}");
        let why = refusal(&w.land(&card, &merge, &["--met", "2=ref"]));
        assert!(why.contains("criterion 2"), "{case}: {why}");
        assert_eq!(w.events(&card, "card.finish"), 0, "{case}");

        w.pass(&card, &audited, S_AUD, &[]);
        let released = result(&w.land(&card, &merge, &[]));
        assert_eq!(released["closed"], json!(true), "{case}: {released}");
        assert!(released.get("held").is_none(), "{case}: {released}");
        assert_eq!(w.state(&card), "done", "{case}");
    }
}

/// REQ-5: a held criterion marked met after the landing is open again on a rerun, and the
/// replaced evidence is recorded although the landing event of the head is already there.
#[test]
fn a_rerun_opens_a_held_criterion_marked_met_since() {
    let w = World::new();
    let card = w.review_card("apps/a/", 2);
    let head = w.lane("notes/a.md");
    w.pass(&card, &head, S_AUD, &[2]);
    result(&w.land(&card, &head, &[]));
    w.met(&card, 2, "after the landing");

    let again = result(&w.land(&card, &head, &[]));
    assert_eq!(again["closed"], json!(false), "{again}");
    assert_eq!(again["open_criteria"], json!([1, 2]), "{again}");
    assert_eq!(w.statuses(&card), ["open", "open"]);
    let reopened = w.payload(&card, "card.held_reopened");
    assert_eq!(reopened["head"], json!(head), "{reopened}");
    assert_eq!(
        reopened["reopened"],
        json!([{"seq": 2, "evidence": "after the landing"}]),
        "{reopened}"
    );
    assert_eq!(w.events(&card, "card.landed"), 1);

    // With nothing marked met since, another rerun writes no event.
    result(&w.land(&card, &head, &[]));
    assert_eq!(w.events(&card, "card.held_reopened"), 1);
}

/// REQ-3: the same call again changes a reviewer's hold when a newer PASS of that reviewer
/// changed it in between; a FAIL in between is not undone by repeating an earlier PASS.
#[test]
fn a_hold_set_again_after_a_release_applies() {
    let w = World::new();
    let card = w.review_card("apps/a/", 2);
    let head = w.lane("notes/a.md");
    let first = w.pass(&card, &head, S_AUD, &[2]);
    let release = w.pass(&card, &head, S_AUD, &[]);

    let again = result(&w.verdict(&card, &head, S_AUD, "pass", &[2]));
    assert_eq!(again["written"], json!(true), "{again}");
    let again = again["event"].as_i64().unwrap();
    assert!(again > release && release > first);
    let repeat = result(&w.verdict(&card, &head, S_AUD, "pass", &[2]));
    assert_eq!(repeat["written"], json!(false), "{repeat}");
    let got = result(&w.land(&card, &head, &[]));
    assert_eq!(got["closed"], json!(false), "{got}");
    assert_eq!(
        got["held"],
        json!([{"seq": 2, "reviewer": S_AUD, "event": again}]),
        "{got}"
    );

    // The release is the same call as the first release, and it is written again.
    let released = result(&w.verdict(&card, &head, S_AUD, "pass", &[]));
    assert_eq!(released["written"], json!(true), "{released}");
    let got = result(&w.land(&card, &head, &[]));
    assert_eq!(got["closed"], json!(true), "{got}");

    // A FAIL stays the newest verdict when an earlier PASS is repeated.
    let w = World::new();
    let card = w.review_card("apps/a/", 2);
    let head = w.lane("notes/a.md");
    w.pass(&card, &head, S_AUD, &[2]);
    result(&w.verdict(&card, &head, S_AUD, "fail", &[]));
    let repeat = result(&w.verdict(&card, &head, S_AUD, "pass", &[2]));
    assert_eq!(repeat["written"], json!(false), "{repeat}");
    let why = refusal(&w.land(&card, &head, &[]));
    assert!(why.contains("verdict_fail"), "{why}");
    assert_eq!(w.integ(), w.c0);
}

/// AC-4 (REQ-6): `--met` on a held criterion refuses before the first write.
#[test]
fn land_refuses_met_for_a_held_criterion() {
    let w = World::new();
    let card = w.review_card("apps/a/", 2);
    let head = w.lane("notes/a.md");
    let event = w.pass(&card, &head, S_AUD, &[2]);

    let why = refusal(&w.land(&card, &head, &["--met", "2=ref"]));
    assert!(why.contains("criterion 2"), "{why}");
    assert!(why.contains(&format!("event {event}")), "{why}");
    assert_eq!(w.integ(), w.c0);
    assert_eq!(w.events(&card, "card.land_started"), 0);
    assert_eq!(w.events(&card, "card.landed"), 0);
    assert_eq!(w.statuses(&card), ["open", "open"]);
}

/// AC-5 (REQ-3): the hold is the union of each reviewer's newest PASS; a newer PASS without
/// `--open` releases that reviewer's hold only.
#[test]
fn the_hold_is_the_union_of_each_reviewers_newest_pass() {
    let w = World::new();
    let card = w.review_card("apps/a/", 3);
    let head = w.lane("notes/a.md");
    let first = w.pass(&card, &head, S_AUD, &[1]);
    let second = w.pass(&card, &head, S_AUD2, &[3]);

    let got = result(&w.land(&card, &head, &["--keep-open"]));
    assert_eq!(
        got["held"],
        json!([
            {"seq": 1, "reviewer": S_AUD, "event": first},
            {"seq": 3, "reviewer": S_AUD2, "event": second},
        ]),
        "{got}"
    );

    // The first reviewer releases its hold; the other's stays.
    w.pass(&card, &head, S_AUD, &[]);
    let got = result(&w.land(&card, &head, &[]));
    assert_eq!(got["closed"], json!(false), "{got}");
    assert_eq!(
        got["held"],
        json!([{"seq": 3, "reviewer": S_AUD2, "event": second}]),
        "{got}"
    );
    assert_ne!(w.state(&card), "done");

    // With both released the same head closes the card, and the answer names no hold.
    w.pass(&card, &head, S_AUD2, &[]);
    let got = result(&w.land(&card, &head, &[]));
    assert_eq!(got["closed"], json!(true), "{got}");
    assert!(got.get("held").is_none(), "{got}");
    assert_eq!(w.state(&card), "done");
}

/// AC-6 (REQ-3, REQ-4, REQ-7): a batch lands a held head as a slice and closes the clean one.
#[test]
fn land_batch_lands_a_held_head_as_a_slice() {
    let w = World::new();
    let held = w.review_card("apps/a/", 1);
    let clean = w.review_card("apps/b/", 1);
    let held_head = w.lane("notes/a.md");
    let clean_head = w.lane("notes/b.md");
    let event = w.pass(&held, &held_head, S_AUD, &[1]);
    w.pass(&clean, &clean_head, S_AUD, &[]);
    let check = w.scripts.path().join("check.sh");
    fs::write(&check, "#!/bin/sh\nexit 0\n").unwrap();
    fs::set_permissions(&check, fs::Permissions::from_mode(0o755)).unwrap();

    let held_pair = format!("{held}={held_head}");
    let clean_pair = format!("{clean}={clean_head}");
    // The held head goes second, so it lands as a join candidate and the hold is read on the
    // audited head its verdict names.
    let got = result(&w.card_with(
        Some(S_LEAD),
        &["land-batch", &clean_pair, &held_pair, "--by", "leader"],
        Some(&check),
    ));
    assert!(got["stopped"].is_null(), "{got}");
    assert_eq!(got["landed"][0]["card"], json!(clean), "{got}");
    assert_eq!(got["landed"][0]["closed"], json!(true), "{got}");
    assert_eq!(got["landed"][1]["card"], json!(held), "{got}");
    assert_eq!(got["landed"][1]["closed"], json!(false), "{got}");
    assert_eq!(
        got["landed"][1]["held"],
        json!([{"seq": 1, "reviewer": S_AUD, "event": event}]),
        "{got}"
    );
    assert_eq!(got["landed"][1]["open_criteria"], json!([1]), "{got}");
    assert!(got["landed"][0].get("held").is_none(), "{got}");
    assert_ne!(w.state(&held), "done");
    assert_eq!(w.statuses(&held), ["open"]);
    assert_eq!(w.state(&clean), "done");
}

/// A hold opens no way around the PASS a closing landing needs: a head asked to close whose PASS
/// the check does not rely on (tier 1 with no receipt) is refused as before, held or not.
#[test]
fn a_hold_asks_for_the_same_pass_as_a_close() {
    let w = World::new();
    let card = w.review_card("apps/a/", 2);
    let head = w.lane("notes/a.md");
    result(&w.card(
        Some(S_AUD),
        &[
            "verdict",
            &card,
            "--head",
            &head,
            "--base",
            &w.c0,
            "--verdict",
            "pass",
            "--tier",
            "1",
            "--evidence",
            "no receipt",
            "--open",
            "2",
        ],
    ));
    let why = refusal(&w.land(&card, &head, &[]));
    assert!(why.contains("receipt_missing"), "{why}");
    assert_eq!(w.integ(), w.c0);
    assert_eq!(w.events(&card, "card.land_started"), 0);
    assert_eq!(w.events(&card, "card.landed"), 0);
}

/// REQ-3: a criterion withdrawn after the verdict is held no longer.
#[test]
fn a_withdrawn_criterion_is_held_no_longer() {
    let w = World::new();
    let card = w.active_card("apps/a/", 2);
    let head = w.lane("notes/a.md");
    w.pass(&card, &head, S_AUD, &[2]);
    result(&w.card(
        Some(S_A),
        &[
            "criterion",
            &card,
            "2",
            "--generation",
            "1",
            "--withdraw",
            "--reason",
            "no longer needed",
        ],
    ));
    w.to_review(&card);

    let got = result(&w.land(&card, &head, &[]));
    assert_eq!(got["closed"], json!(true), "{got}");
    assert!(got.get("held").is_none(), "{got}");
    assert_eq!(w.statuses(&card), ["met", "withdrawn"]);
}
