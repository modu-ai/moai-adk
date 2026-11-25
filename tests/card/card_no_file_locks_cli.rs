//! SPEC-20261125-card-no-file-locks: the file-level write lock between cards is gone. A take, a
//! scope widening and `card next` no longer refuse or skip a card because its write scope can be
//! written under another card's reservation; each answers the overlap instead. What stays: one
//! holder per card, the generation fence, HOLD, empty scopes and cards that are not ready.
//!
//! Every command runs against a temporary HOME; the real home state is not read or written. The
//! holder is the session id in the environment (`CLAUDE_CODE_SESSION_ID`).
#![cfg(unix)]

use crate::common;

use common::git::{git, git_ok};
use common::home::TempHome;
use moai_cli::ledger;
use serde_json::{json, Value};
use std::io::Write;
use std::path::PathBuf;
use std::process::{Output, Stdio};

const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
const S_A: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
const S_B: &str = "bbbbbbbb-0000-4000-8000-00000000000b";

struct World {
    home: TempHome,
    project: String,
}

fn world() -> World {
    let home = TempHome::new();
    let root: PathBuf = home.project("demo");
    ledger::register(home.path(), "demo").unwrap();
    World {
        project: root.to_str().unwrap().to_string(),
        home,
    }
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

/// The message of a refused command (exit 1, `ok: false`).
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

impl World {
    /// `moai card <args> --project <project>`, as `session`.
    fn card(&self, session: Option<&str>, args: &[&str]) -> Output {
        let mut all = vec!["card"];
        all.extend_from_slice(args);
        all.extend_from_slice(&["--project", &self.project]);
        self.run(session, &all)
    }

    fn run(&self, session: Option<&str>, args: &[&str]) -> Output {
        let mut cmd = self.home.command(args);
        if let Some(id) = session {
            cmd.env(SESSION_VAR, id);
        }
        cmd.stdin(Stdio::null()).output().expect("run moai")
    }

    fn next(&self, session: Option<&str>) -> Value {
        result(&self.run(session, &["card", "next", &self.project]))
    }

    /// A `ready` card with one criterion and the write scope given (none when empty).
    fn ready_card(&self, title: &str, scope: &[&str]) -> String {
        let mut args = vec!["add", "--title", title, "--ac", "fixture"];
        for s in scope {
            args.extend_from_slice(&["--scope", s]);
        }
        let added = result(&self.card(None, &args));
        let name = added["name"].as_str().unwrap().to_string();
        result(&self.card(None, &["ready", &name]));
        name
    }

    /// Takes a card for `session` with `scope`; returns the whole answer.
    fn take(&self, session: &str, card: &str, scope: &[&str]) -> Value {
        let mut args = vec!["take", card];
        for s in scope {
            args.extend_from_slice(&["--scope", s]);
        }
        result(&self.card(Some(session), &args))
    }

    fn count(&self, sql: &str) -> i64 {
        let l = ledger::open_read_only(self.home.path(), "demo").unwrap();
        l.conn().query_row(sql, [], |r| r.get(0)).unwrap()
    }

    /// The scope of the active reservation of `card` as the ledger holds it.
    fn reserved(&self, card: &str) -> Vec<String> {
        let id: i64 = card.trim_start_matches("C-").parse().unwrap();
        let l = ledger::open_read_only(self.home.path(), "demo").unwrap();
        let mut stmt = l
            .conn()
            .prepare(
                "SELECT s.pattern FROM reservation r JOIN reservation_scope s \
                 ON s.reservation_id = r.reservation_id \
                 WHERE r.card_id = ?1 AND r.status = 'held' ORDER BY s.pattern",
            )
            .unwrap();
        stmt.query_map([id], |r| r.get::<_, String>(0))
            .unwrap()
            .map(|r| r.unwrap())
            .collect()
    }
}

/// AC-1 (REQ-1, REQ-8): a take whose scope overlaps another card's reservation is made, answers
/// the overlap, and leaves the other card's reservation and the event log alone.
#[test]
fn ac1_a_take_over_another_cards_paths_is_made_and_reports_the_overlap() {
    let w = world();
    let first = w.ready_card("first", &["src/a.rs"]);
    let second = w.ready_card("second", &["src/**"]);
    w.take(S_A, &first, &["src/a.rs"]);
    let yields = w.count("SELECT count(*) FROM event WHERE kind = 'card.yield'");
    let took = w.take(S_B, &second, &["src/**"]);
    assert_eq!(
        took["reservation"]["write_scope"],
        json!(["src/**"]),
        "{took}"
    );
    assert_eq!(took["state"], json!("active"), "{took}");
    let overlaps = took["overlaps"]
        .as_array()
        .unwrap_or_else(|| panic!("{took}"));
    assert_eq!(overlaps.len(), 1, "{took}");
    assert_eq!(overlaps[0]["card"], json!(first), "{took}");
    assert_eq!(overlaps[0]["holder"], json!(S_A), "{took}");
    assert_eq!(overlaps[0]["theirs"], json!("src/a.rs"), "{took}");
    assert_eq!(overlaps[0]["mine"], json!("src/**"), "{took}");
    assert_eq!(
        w.reserved(&first),
        ["src/a.rs"],
        "the first card is left alone"
    );
    assert_eq!(
        w.count("SELECT count(*) FROM event WHERE kind = 'card.yield'"),
        yields,
        "no card.yield event"
    );
}

/// AC-1 (REQ-1, REQ-8): a reservation whose lease ended is not given up to the new take.
#[test]
fn ac1_an_ended_lease_keeps_its_paths_when_another_card_takes_over_them() {
    let w = world();
    let first = w.ready_card("first", &["src/a.rs"]);
    let second = w.ready_card("second", &["src/**"]);
    w.take(S_A, &first, &["src/a.rs"]);
    {
        let mut l = ledger::open_write(w.home.path(), "demo").unwrap();
        l.with_write_tx(|tx, _| {
            tx.execute_batch("UPDATE reservation SET lease_until = 1 WHERE status = 'held'")?;
            Ok(())
        })
        .unwrap();
    }
    let took = w.take(S_B, &second, &["src/**"]);
    assert_eq!(took["state"], json!("active"), "{took}");
    assert_eq!(
        w.count("SELECT count(*) FROM event WHERE kind = 'card.yield'"),
        0,
        "no card.yield event"
    );
    let kept: i64 = w.count(
        "SELECT count(*) FROM reservation_scope s JOIN reservation r \
         ON r.reservation_id = s.reservation_id WHERE s.pattern = 'src/a.rs'",
    );
    assert_eq!(kept, 1, "the ended reservation keeps its scope");
}

/// AC-2 (REQ-2, REQ-8): widening a held card by a path another card holds is made and reports it.
#[test]
fn ac2_a_widening_over_another_cards_paths_is_made_and_reports_the_overlap() {
    let w = world();
    let first = w.ready_card("first", &["tests/x.rs"]);
    let second = w.ready_card("second", &["src/b.rs"]);
    w.take(S_A, &first, &["tests/x.rs"]);
    w.take(S_B, &second, &["src/b.rs"]);
    let widened = result(&w.card(
        Some(S_B),
        &[
            "scope",
            &second,
            "--generation",
            "1",
            "--by",
            "agent-b",
            "--reason",
            "needs the test",
            "--scope",
            "src/b.rs",
            "--scope",
            "tests/x.rs",
        ],
    ));
    assert_eq!(widened["added"], json!(["tests/x.rs"]), "{widened}");
    let overlaps = widened["overlaps"]
        .as_array()
        .unwrap_or_else(|| panic!("{widened}"));
    assert_eq!(overlaps.len(), 1, "{widened}");
    assert_eq!(overlaps[0]["card"], json!(first), "{widened}");
    assert_eq!(overlaps[0]["theirs"], json!("tests/x.rs"), "{widened}");
    assert_eq!(
        w.reserved(&first),
        ["tests/x.rs"],
        "the first card is left alone"
    );
    assert_eq!(w.reserved(&second), ["src/b.rs", "tests/x.rs"]);
}

/// AC-3 (REQ-3): `card next` lists a card whose scope overlaps another reservation as ready, and
/// no skipped card carries the reason `scope_overlap`; a HOLD card is still skipped.
#[test]
fn ac3_card_next_no_longer_skips_a_card_for_an_overlap() {
    let w = world();
    let first = w.ready_card("first", &["src/a.rs"]);
    let second = w.ready_card("second", &["src/**"]);
    let held = w.ready_card("held", &["docs/**"]);
    w.take(S_A, &first, &["src/a.rs"]);
    result(&w.card(
        None,
        &[
            "hold",
            &held,
            "--category",
            "legal",
            "--reason",
            "why",
            "--by",
            "user",
        ],
    ));
    let next = w.next(Some(S_B));
    let ready: Vec<&str> = next["ready"]
        .as_array()
        .unwrap()
        .iter()
        .map(|c| c["name"].as_str().unwrap())
        .collect();
    assert_eq!(ready, [second.as_str()], "{next}");
    let skipped = next["skipped"].as_array().unwrap();
    assert!(
        skipped
            .iter()
            .all(|s| s["reason"] != json!("scope_overlap")),
        "{next}"
    );
    let hold = skipped
        .iter()
        .find(|s| s["name"] == json!(held))
        .unwrap_or_else(|| panic!("{next}"));
    assert_eq!(hold["reason"], json!("hold"), "{next}");
}

/// AC-4 (REQ-4): `card check` allows a path another card also holds and names that card; a path
/// no other card holds has no overlap; a path outside the reservation is still refused.
#[test]
fn ac4_card_check_names_the_other_card_that_holds_the_path() {
    let w = world();
    let first = w.ready_card("first", &["src/a.rs"]);
    let second = w.ready_card("second", &["src/**"]);
    w.take(S_A, &first, &["src/a.rs"]);
    w.take(S_B, &second, &["src/**"]);
    let both = result(&w.card(
        Some(S_B),
        &["check", &second, "--generation", "1", "--path", "src/a.rs"],
    ));
    assert_eq!(both["allowed"], json!(true), "{both}");
    let overlaps = both["overlaps"]
        .as_array()
        .unwrap_or_else(|| panic!("{both}"));
    assert_eq!(overlaps.len(), 1, "{both}");
    assert_eq!(overlaps[0]["path"], json!("src/a.rs"), "{both}");
    assert_eq!(overlaps[0]["card"], json!(first), "{both}");
    assert_eq!(overlaps[0]["pattern"], json!("src/a.rs"), "{both}");
    let alone = result(&w.card(
        Some(S_B),
        &["check", &second, "--generation", "1", "--path", "src/z.rs"],
    ));
    assert_eq!(alone["allowed"], json!(true), "{alone}");
    assert_eq!(alone["overlaps"], json!([]), "{alone}");
    let outside = refusal(&w.card(
        Some(S_B),
        &["check", &second, "--generation", "1", "--path", "docs/a.md"],
    ));
    assert!(outside.contains("is not under what"), "{outside}");
}

/// A project folder that is a git repository whose integration branch is `main`, with the base
/// commit on it and the receipt helpers of the READY gate.
struct Repo {
    home: TempHome,
    root: PathBuf,
    base: String,
}

fn commit(dir: &std::path::Path, files: &[&str], msg: &str) -> String {
    for f in files {
        let path = dir.join(f);
        std::fs::create_dir_all(path.parent().unwrap()).unwrap();
        std::fs::write(&path, format!("{msg}\n")).unwrap();
    }
    git_ok(dir, &["add", "-A"]);
    git_ok(dir, &["commit", "-q", "-m", msg]);
    git(dir, &["rev-parse", "HEAD"])
}

impl Repo {
    /// `with_profile` false leaves the project without a work flow, so no integration branch.
    fn new(with_profile: bool) -> Repo {
        let home = TempHome::new();
        let root = home.project("demo");
        ledger::register(home.path(), "demo").unwrap();
        git_ok(&root, &["init", "-q", "-b", "main"]);
        git_ok(&root, &["config", "user.email", "t@example.invalid"]);
        git_ok(&root, &["config", "user.name", "t"]);
        std::fs::write(
            root.join(".gitignore"),
            format!("target/\n{}/\n", moai_cli::PROJECT_STATE_DIR),
        )
        .unwrap();
        let base = commit(&root, &["src/a.rs", "src/b.rs"], "base");
        if with_profile {
            let cfg = moai_cli::state_root::path(&root, "project.yaml");
            std::fs::create_dir_all(cfg.parent().unwrap()).unwrap();
            std::fs::write(
                &cfg,
                "{\"schema_version\": 1, \"project_id\": \"demo\", \"purpose\": \"fixture\", \
                 \"tools\": [\"rust\"], \"development_methods\": [\"sdd\"], \"hosts\": [\"claude\"], \
                 \"work_flow\": {\"flow\": \"next-flow\", \"integration_branch\": \"main\"}}",
            )
            .unwrap();
        }
        Repo { home, root, base }
    }

    /// A lane commit on branch `lane` changing `lane_files`, then the integration branch moves
    /// by one commit changing `tip_files`; the project ends at the lane head, which is returned.
    fn diverge(&self, lane_files: &[&str], tip_files: &[&str]) -> String {
        git_ok(&self.root, &["switch", "-q", "-c", "lane"]);
        let head = commit(&self.root, lane_files, "lane");
        git_ok(&self.root, &["switch", "-q", "main"]);
        commit(&self.root, tip_files, "tip");
        git_ok(&self.root, &["switch", "-q", "lane"]);
        head
    }

    /// Writes the card, the PASS receipt of `head` and a READY note; returns the note answer.
    fn ready_note(&self, head: &str) -> Value {
        let project = self.root.to_str().unwrap();
        let added = result(
            &self
                .home
                .command(&[
                    "card",
                    "add",
                    "--title",
                    "lane work",
                    "--scope",
                    "area/",
                    "--ac",
                    "done",
                    "--project",
                    project,
                ])
                .stdin(Stdio::null())
                .output()
                .unwrap(),
        );
        let card = added["name"].as_str().unwrap().to_string();
        crate::ready_receipt::write_receipt_based(&self.root, head, &self.base);
        let text = format!(
            "LANE-READY-INTEGRATION {card} head={head}\nbase={}\nleftovers: none\n",
            self.base
        );
        let mut child = self
            .home
            .command(&[
                "card",
                "note",
                &card,
                "--kind",
                "report",
                "--by",
                "lane",
                "--project",
                project,
            ])
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .spawn()
            .unwrap();
        child
            .stdin
            .take()
            .unwrap()
            .write_all(text.as_bytes())
            .unwrap();
        result(&child.wait_with_output().unwrap())
    }
}

/// AC-5 (REQ-5): a READY note's answer lists the files the lane and the integration branch both
/// changed since the base.
#[test]
fn ac5_a_ready_note_lists_the_files_the_tip_changed_under_the_lane() {
    let r = Repo::new(true);
    let head = r.diverge(&["src/a.rs", "src/new.rs"], &["src/a.rs", "src/c.rs"]);
    let answer = r.ready_note(&head);
    assert_eq!(answer["tip_overlap"], json!(["src/a.rs"]), "{answer}");
}

#[test]
fn ac5_a_ready_note_with_no_shared_file_answers_an_empty_list() {
    let r = Repo::new(true);
    let head = r.diverge(&["src/b.rs"], &["src/c.rs"]);
    let answer = r.ready_note(&head);
    assert_eq!(answer["tip_overlap"], json!([]), "{answer}");
}

#[test]
fn ac5_a_ready_note_keeps_the_note_when_the_integration_branch_is_unknown() {
    let r = Repo::new(false);
    let head = r.diverge(&["src/a.rs"], &["src/a.rs"]);
    let answer = r.ready_note(&head);
    assert_eq!(answer["tip_overlap"], json!("not_checked"), "{answer}");
    assert!(
        answer["tip_overlap_reason"]
            .as_str()
            .is_some_and(|s| !s.is_empty()),
        "{answer}"
    );
    assert!(
        answer["note_id"].as_i64().is_some(),
        "the note is kept: {answer}"
    );
}

/// AC-6 (REQ-6, REQ-9): what stays refused. A card another session holds, a card on HOLD, a
/// card with no write scope and a stale generation are refused as before.
#[test]
fn ac6_the_card_owner_hold_scope_and_generation_checks_stay() {
    let w = world();
    let first = w.ready_card("first", &["src/a.rs"]);
    let held = w.ready_card("held", &["docs/**"]);
    let empty = w.ready_card("empty", &[]);
    w.take(S_A, &first, &["src/a.rs"]);
    let again = refusal(&w.card(Some(S_B), &["take", &first, "--scope", "src/a.rs"]));
    assert!(again.contains("already held by"), "{again}");
    result(&w.card(
        None,
        &[
            "hold",
            &held,
            "--category",
            "legal",
            "--reason",
            "why",
            "--by",
            "user",
        ],
    ));
    let on_hold = refusal(&w.card(Some(S_B), &["take", &held, "--scope", "docs/**"]));
    assert!(on_hold.contains("hold"), "{on_hold}");
    let none = refusal(&w.card(Some(S_B), &["take", &empty]));
    assert!(none.contains("scope"), "{none}");
    let stale = refusal(&w.card(
        Some(S_A),
        &["check", &first, "--generation", "7", "--path", "src/a.rs"],
    ));
    assert!(stale.contains("generation"), "{stale}");
    let widen_stale = refusal(&w.card(
        Some(S_A),
        &[
            "scope",
            &first,
            "--generation",
            "7",
            "--by",
            "agent-a",
            "--reason",
            "r",
            "--scope",
            "src/a.rs",
            "--scope",
            "src/b.rs",
        ],
    ));
    assert!(widen_stale.contains("generation"), "{widen_stale}");
}
