//! `moai card land` of a merge commit whose second parent is the audited head
//! (SPEC-20261125-land-closes-audited-merge): when the card has a verdict on the second parent and
//! the merge is the clean merge of its parents, that verdict covers the closing landing, and every
//! check the verdict meets stays. Every test runs on a temporary home and a temporary repository
//! whose integration branch is `integ`; nothing here touches the real ledger.
//!
//! The repository: `main` holds two commits; `head` is the second (the card's change) and `integ`
//! points at `tip`, a child of `head` (it may add a file). A lane commit sits on `head`, and a merge
//! commit is made with `commit-tree`, so the working tree never changes.

use crate::common;

use common::home::TempHome;
use moai_cli::ledger;
use rusqlite::Connection;
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Output, Stdio};

const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
/// The lane's author and card holder, the independent reviewer, the lander (the leader).
const S_A: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
const S_AUD: &str = "cccccccc-0000-4000-8000-00000000000c";
const S_LEAD: &str = "dddddddd-0000-4000-8000-00000000000d";
const WORKTREE_KEY: &str = "0123456789ab";

/// The profile of the fixture project: `mode` is the `land_verdict` word.
fn profile(mode: &str) -> Vec<u8> {
    json!({
        "schema_version": 1, "project_id": "demo", "purpose": "fixture", "tools": ["rust"],
        "development_methods": ["sdd"], "hosts": ["claude"],
        "work_flow": {"flow": "m-flow", "integration_branch": "integ", "land_verdict": mode},
    })
    .to_string()
    .into_bytes()
}

/// A scan table with one scan that every new file under `tests/` needs.
fn table() -> String {
    json!({"scans": [
        {"id": "test-area-layout", "always": true,
         "tests": ["tests/test_area_layout.rs::area_table_maps_every_real_test_file"],
         "checks": ["cargo test --offline --locked --test test_area_layout"],
         "rules": ["tests/test_area_layout.rs", "tests/areas.toml"],
         "scope": [{"prefix": "tests/", "suffixes": [".rs", ".toml"]}],
         "rejects": "a test file whose first word is not in tests/areas.toml"}
    ]})
    .to_string()
}

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

/// A commit message with the trailer block of `session`.
fn signed(subject: &str, session: &str) -> String {
    format!(
        "{subject}\n\nMoai-Spec: SPEC-20261125-fixture\nMoai-Session: {session}:{WORKTREE_KEY}\nMoai-Verified: none"
    )
}

/// A commit on top of `parent` that sets `path` to `content` (adds it or replaces it); no ref moves.
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

fn dump_tables(conn: &Connection, tables: &[&str]) -> Vec<String> {
    let mut out = Vec::new();
    for table in tables {
        let mut st = conn
            .prepare(&format!("SELECT * FROM {table} ORDER BY 1, 2"))
            .unwrap();
        let n = st.column_count();
        let lines: Vec<String> = st
            .query_map([], |r| {
                let mut line = format!("{table}:");
                for i in 0..n {
                    line.push_str(&format!("{:?}|", r.get_ref(i)?));
                }
                Ok(line)
            })
            .unwrap()
            .map(Result::unwrap)
            .collect();
        out.extend(lines);
    }
    out
}

fn number(name: &str) -> i64 {
    name.trim_start_matches("C-").parse().unwrap()
}

struct World {
    home: TempHome,
    project: PathBuf,
    p: String,
    mode: String,
    head: String,
    tip: String,
}

impl World {
    /// A repository whose first commit holds `table` (none: no table at all); the tip adds
    /// `tip_adds` on top of the card's change when it is given.
    fn new(mode: &str, table: Option<&str>, tip_adds: Option<(&str, &str)>) -> Self {
        let home = TempHome::new();
        let project = home.project("demo");
        ledger::register(home.path(), "demo").unwrap();
        git(&project, &["init", "-q", "-b", "main"]);
        fs::write(project.join("a.txt"), "a\n").unwrap();
        git(&project, &["add", "a.txt"]);
        if let Some(text) = table {
            fs::create_dir_all(project.join("tools")).unwrap();
            fs::write(project.join("tools/repo_scans.json"), text).unwrap();
            git(&project, &["add", "tools/repo_scans.json"]);
        }
        git(&project, &["commit", "-q", "-m", "init"]);
        fs::write(project.join("b.txt"), "b\n").unwrap();
        git(&project, &["add", "b.txt"]);
        git(
            &project,
            &["commit", "-q", "-m", &signed("the card's change", S_A)],
        );
        let head = git(&project, &["rev-parse", "HEAD"]);
        let tip = match tip_adds {
            Some((path, content)) => put_on(&project, &head, path, content, "integration moved on"),
            None => {
                let tree = git(&project, &["rev-parse", &format!("{head}^{{tree}}")]);
                git(
                    &project,
                    &[
                        "commit-tree",
                        &tree,
                        "-p",
                        &head,
                        "-m",
                        "integration moved on",
                    ],
                )
            }
        };
        git(&project, &["update-ref", "refs/heads/integ", &tip]);
        World {
            p: project.to_str().unwrap().to_string(),
            home,
            project,
            mode: mode.to_string(),
            head,
            tip,
        }
    }

    /// The common case: no table, the tip a plain child of the card's change.
    fn plain(mode: &str) -> Self {
        World::new(mode, None, None)
    }

    fn card(&self, session: Option<&str>, args: &[&str]) -> Output {
        let cfg = moai_cli::state_root::path(&self.project, "project.yaml");
        fs::create_dir_all(cfg.parent().unwrap()).unwrap();
        fs::write(&cfg, profile(&self.mode)).unwrap();
        let mut all = vec!["card"];
        all.extend_from_slice(args);
        all.extend_from_slice(&["--project", &self.p]);
        let mut cmd: Command = self.home.command(&all);
        cmd.env("MOAI_GIT_BIN", git_executable());
        if let Some(id) = session {
            cmd.env(SESSION_VAR, id);
        }
        cmd.stdin(Stdio::null());
        cmd.output().expect("run moai")
    }

    /// A task card with one open criterion, taken by `S_A` and sent to review; its name.
    fn review_card(&self) -> String {
        let added = result(&self.card(
            None,
            &[
                "add",
                "--title",
                "fixture card",
                "--kind",
                "task",
                "--scope",
                "apps/a/",
                "--ac",
                "one",
            ],
        ));
        let name = added["name"].as_str().unwrap().to_string();
        result(&self.card(None, &["ready", &name]));
        result(&self.card(Some(S_A), &["take", &name]));
        result(&self.card(Some(S_A), &["review", &name, "--generation", "1"]));
        name
    }

    /// A lane commit on the card's change that sets `path`, written by the card's holder.
    fn lane_setting(&self, path: &str, content: &str) -> String {
        put_on(
            &self.project,
            &self.head,
            path,
            content,
            &signed("the lane change", S_A),
        )
    }

    /// A lane commit on the card's change that adds `path`.
    fn lane_adding(&self, path: &str) -> String {
        self.lane_setting(path, "lane\n")
    }

    /// The tree `git merge-tree --write-tree` makes of two commits (they must merge cleanly).
    fn clean_tree(&self, first: &str, second: &str) -> String {
        git(
            &self.project,
            &["merge-tree", "--write-tree", first, second],
        )
    }

    /// A merge commit with the given tree and parents, signed by the lander.
    fn merge_of(&self, tree: &str, parents: &[&str]) -> String {
        let mut args = vec!["commit-tree", tree];
        for parent in parents {
            args.push("-p");
            args.push(parent);
        }
        let message = signed("merge the lane on the tip", S_LEAD);
        args.push("-m");
        args.push(&message);
        git(&self.project, &args)
    }

    /// The clean join the leader lands: `second` merged on `first` with the tree git makes.
    fn join(&self, first: &str, second: &str) -> String {
        let tree = self.clean_tree(first, second);
        self.merge_of(&tree, &[first, second])
    }

    /// One verdict event of `card` on `head`, written as the `card verdict` command writes it;
    /// returns the event id.
    fn verdict(
        &self,
        card: &str,
        head: &str,
        pass: bool,
        tier: u8,
        reviewer: &str,
        receipt: Option<Value>,
    ) -> i64 {
        let conn =
            Connection::open(ledger::ledger_path(self.home.path(), "demo").unwrap()).unwrap();
        let rows = self.count("SELECT count(*) FROM event");
        let payload = json!({
            "head": head, "base": self.tip, "verdict": if pass { "pass" } else { "fail" },
            "tier": tier, "evidence": "the fixture's evidence", "receipt": receipt,
            "reviewer": reviewer,
        });
        conn.execute(
            "INSERT INTO event (event_key, at, actor, card_id, kind, payload) \
             VALUES (?1, unixepoch(), ?2, ?3, 'card.verdict', ?4)",
            rusqlite::params![
                format!("test-verdict-{rows}"),
                reviewer,
                number(card),
                payload.to_string()
            ],
        )
        .unwrap();
        conn.last_insert_rowid()
    }

    /// A receipt of `head` with two steps at exit 0.
    fn good_receipt(&self, head: &str) -> Value {
        json!({"head": head, "steps": [{"name": "fmt", "exit": 0}, {"name": "test", "exit": 0}]})
    }

    /// `card land` of `head` as the leader session, naming no `--met` option so the verdict fills
    /// every open criterion.
    fn land(&self, card: &str, head: &str) -> Output {
        self.card(
            Some(S_LEAD),
            &["land", card, "--head", head, "--by", "leader"],
        )
    }

    fn integ(&self) -> String {
        git(&self.project, &["rev-parse", "refs/heads/integ"])
    }

    fn dump(&self) -> Vec<String> {
        let l = ledger::open_read_only(self.home.path(), "demo").unwrap();
        dump_tables(
            l.conn(),
            &[
                "card",
                "card_ac",
                "card_note",
                "card_link",
                "reservation",
                "reservation_scope",
                "event",
            ],
        )
    }

    fn count(&self, sql: &str) -> i64 {
        let l = ledger::open_read_only(self.home.path(), "demo").unwrap();
        l.conn().query_row(sql, [], |r| r.get(0)).unwrap()
    }

    fn events(&self, card: &str, kind: &str) -> i64 {
        self.count(&format!(
            "SELECT count(*) FROM event WHERE card_id = {} AND kind = '{kind}'",
            number(card)
        ))
    }

    /// Asserts a refusal that wrote nothing and moved nothing: exit 1, `ok` false, a message that
    /// holds every one of `needles`, every table as it was, `integ` at the tip and no start row.
    /// Returns the message.
    fn refuse(&self, out: &Output, needles: &[&str], before: &[String], card: &str) -> String {
        assert_eq!(
            out.status.code(),
            Some(1),
            "{} {}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        );
        let v = envelope(out);
        assert_eq!(v["ok"], json!(false), "{v}");
        let message = v["error"]["message"].as_str().unwrap().to_string();
        for needle in needles {
            assert!(
                message.contains(needle),
                "the reason lacks {needle}: {message}"
            );
        }
        assert_eq!(&self.dump(), before, "the refused call wrote to the ledger");
        assert_eq!(self.integ(), self.tip, "the refused call moved integ");
        assert_eq!(self.events(card, "card.land_started"), 0);
        message
    }
}

/// AC-1 (REQ-1, REQ-7): the clean merge of the audited head on the tip lands and closes the card
/// in one call, in both modes of the verdict check; the answer names the audited head and the
/// criterion's evidence names the verdict event and that head.
#[test]
fn a_clean_merge_of_the_audited_head_closes_the_card_in_one_call() {
    for mode in ["enforce", "observe"] {
        let w = World::plain(mode);
        let card = w.review_card();
        let lane = w.lane_adding("notes/x.md");
        let merge = w.join(&w.tip, &lane);
        let event = w.verdict(&card, &lane, true, 0, S_AUD, None);

        let got = result(&w.land(&card, &merge));
        assert_eq!(got["moved"], json!(true), "{mode}: {got}");
        assert_eq!(got["closed"], json!(true), "{mode}: {got}");
        assert_eq!(got["state"], json!("done"), "{mode}: {got}");
        assert_eq!(got["head"], json!(merge), "{mode}: {got}");
        assert_eq!(got["audited_head"], json!(lane), "{mode}: {got}");
        assert_eq!(got["open_criteria"], json!([]), "{mode}: {got}");
        assert_eq!(
            got["met"][0]["evidence"],
            json!(format!("verdict event {event} PASS head {}", &lane[..12])),
            "{mode}: {got}"
        );
        assert_eq!(w.integ(), merge, "{mode}: the branch is at the merge");
        assert_eq!(w.events(&card, "card.landed"), 1, "{mode}");
    }
}

/// AC-2 (REQ-2): every check the verdict of the second parent meets still refuses the merge, with
/// the code of the check, and nothing is written or moved.
#[test]
fn every_check_of_the_audited_verdict_still_refuses_the_merge() {
    // Each case: what the case is, the table of the repository, the path the lane adds and the code
    // the refusal names.
    let cases: [(&str, Option<String>, &str, &str); 5] = [
        (
            "a reviewer who wrote the lane",
            None,
            "notes/x.md",
            "self_verdict",
        ),
        (
            "a reviewer who is the lander",
            None,
            "notes/x.md",
            "lander_review",
        ),
        (
            "a tier 1 verdict without a receipt",
            None,
            "notes/x.md",
            "receipt_missing",
        ),
        (
            "a newer FAIL of the same reviewer",
            None,
            "notes/x.md",
            "verdict_fail",
        ),
        (
            "a scan the receipt does not cover",
            Some(table()),
            "tests/branch_gc_cli.rs",
            "scans_missing",
        ),
    ];
    for (name, table, path, code) in cases {
        let w = World::new("enforce", table.as_deref(), None);
        let card = w.review_card();
        let lane = w.lane_adding(path);
        let merge = w.join(&w.tip, &lane);
        match code {
            "self_verdict" => {
                w.verdict(&card, &lane, true, 0, S_A, None);
            }
            "lander_review" => {
                w.verdict(&card, &lane, true, 1, S_LEAD, Some(w.good_receipt(&lane)));
            }
            "receipt_missing" => {
                w.verdict(&card, &lane, true, 1, S_AUD, None);
            }
            "verdict_fail" => {
                w.verdict(&card, &lane, true, 0, S_AUD, None);
                w.verdict(&card, &lane, false, 0, S_AUD, None);
            }
            _ => {
                w.verdict(&card, &lane, true, 1, S_AUD, Some(w.good_receipt(&lane)));
            }
        }
        let before = w.dump();
        let out = w.land(&card, &merge);
        w.refuse(&out, &[code], &before, &card);
        assert_eq!(w.events(&card, "card.landed"), 0, "{name}");
    }
}

/// AC-3 (REQ-3): a merge whose first parent is not the tip, with the tip on its first-parent line,
/// is refused and nothing moves.
#[test]
fn a_merge_whose_first_parent_is_not_the_tip_is_refused() {
    let w = World::plain("enforce");
    let card = w.review_card();
    let lane = w.lane_adding("notes/x.md");
    // A commit nobody landed, on top of the tip: the merge is made on it, not on the tip.
    let beyond = put_on(
        &w.project,
        &w.tip,
        "notes/y.md",
        "beyond\n",
        &signed("a commit nobody landed", S_LEAD),
    );
    let merge = w.join(&beyond, &lane);
    w.verdict(&card, &lane, true, 0, S_AUD, None);
    let before = w.dump();
    let out = w.land(&card, &merge);
    let message = w.refuse(&out, &["first parent", "integration tip"], &before, &card);
    assert!(
        message.contains(&beyond),
        "names the first parent: {message}"
    );
}

/// AC-4 (REQ-4): a merge with conflict resolution and a merge edited after it are refused with a
/// message that starts `no_pass:`, names what differs from the clean merge and tells the auditor
/// to record a verdict on the merge itself.
#[test]
fn a_resolved_or_edited_merge_is_refused_with_the_paths_the_auditor_must_read() {
    // The merge conflicts: the tip and the lane both change a.txt. The leader resolves it by hand.
    let w = World::new("enforce", None, Some(("a.txt", "tip\n")));
    let card = w.review_card();
    let lane = w.lane_setting("a.txt", "lane\n");
    let resolved_on = put_on(
        &w.project,
        &w.tip,
        "a.txt",
        "both\n",
        &signed("the resolution", S_LEAD),
    );
    let tree = git(
        &w.project,
        &["rev-parse", &format!("{resolved_on}^{{tree}}")],
    );
    let resolved = w.merge_of(&tree, &[&w.tip, &lane]);
    w.verdict(&card, &lane, true, 0, S_AUD, None);
    let before = w.dump();
    let out = w.land(&card, &resolved);
    let message = w.refuse(
        &out,
        &["no_pass:", "conflict", "card verdict", &resolved],
        &before,
        &card,
    );
    assert!(message.contains("--head"), "{message}");

    // The merge is clean, and a file is added to it afterwards.
    let w = World::plain("enforce");
    let card = w.review_card();
    let lane = w.lane_adding("notes/x.md");
    let clean = w.join(&w.tip, &lane);
    let edited_on = put_on(
        &w.project,
        &clean,
        "extra.txt",
        "extra\n",
        &signed("an edit after the merge", S_LEAD),
    );
    let tree = git(&w.project, &["rev-parse", &format!("{edited_on}^{{tree}}")]);
    let edited = w.merge_of(&tree, &[&w.tip, &lane]);
    w.verdict(&card, &lane, true, 0, S_AUD, None);
    let before = w.dump();
    let out = w.land(&card, &edited);
    w.refuse(
        &out,
        &["no_pass:", "extra.txt", "card verdict", &edited],
        &before,
        &card,
    );
}

/// AC-5 (REQ-5): a merge of three parents and a merge whose second parent has no verdict are
/// refused as `no_pass`, as before, and nothing is written.
#[test]
fn a_merge_without_two_parents_or_without_a_verdict_on_the_second_is_refused_as_before() {
    let w = World::plain("observe");
    let card = w.review_card();
    let lane = w.lane_adding("notes/x.md");
    let other = w.lane_adding("notes/z.md");

    // Three parents; the second one has the PASS.
    w.verdict(&card, &lane, true, 0, S_AUD, None);
    let tree = w.clean_tree(&w.tip, &lane);
    let octopus = w.merge_of(&tree, &[&w.tip, &lane, &other]);
    let before = w.dump();
    let out = w.land(&card, &octopus);
    w.refuse(&out, &["no_pass"], &before, &card);

    // Two parents; the PASS names another head than the second parent.
    let merge = w.join(&w.tip, &other);
    let before = w.dump();
    let out = w.land(&card, &merge);
    w.refuse(&out, &["no_pass"], &before, &card);
}

/// AC-6 (REQ-6): a verdict on the merge itself decides. A FAIL on the merge refuses although the
/// second parent has a PASS; a PASS on the merge lands it on that verdict and the answer names no
/// audited head.
#[test]
fn a_verdict_on_the_merge_itself_decides_and_the_second_parent_is_not_read() {
    let w = World::plain("enforce");
    let card = w.review_card();
    let lane = w.lane_adding("notes/x.md");
    let merge = w.join(&w.tip, &lane);
    w.verdict(&card, &lane, true, 0, S_AUD, None);
    w.verdict(&card, &merge, false, 0, S_AUD, None);
    let before = w.dump();
    let out = w.land(&card, &merge);
    w.refuse(&out, &["verdict_fail"], &before, &card);

    let w = World::plain("enforce");
    let card = w.review_card();
    let lane = w.lane_adding("notes/x.md");
    let merge = w.join(&w.tip, &lane);
    let event = w.verdict(&card, &merge, true, 0, S_AUD, None);
    let got = result(&w.land(&card, &merge));
    assert_eq!(got["closed"], json!(true), "{got}");
    assert_eq!(got["audited_head"], Value::Null, "{got}");
    assert_eq!(got["verdict"]["event"], json!(event), "{got}");
    assert_eq!(
        got["met"][0]["evidence"],
        json!(format!("verdict event {event} PASS head {}", &merge[..12])),
        "{got}"
    );
}
