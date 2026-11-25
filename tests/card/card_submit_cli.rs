//! `moai card submit` (card C-503, NeXT-Flow 2 step 1): one command that checks the worktree and
//! the cheap gates, takes or reuses the PASS receipt, writes the READY note from it, sends the
//! card to review and gives its paths back. Every test runs on a temporary home and a temporary
//! repository whose integration branch is `integ`; the lane works in a linked worktree on `lane`.
//! The long check is replaced by a debug-only test program (`MOAI_TEST_SUBMIT_CHECK`).

use crate::common;

use common::home::TempHome;
use moai_cli::ledger;
use rusqlite::Connection;
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Output, Stdio};

const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
const S_A: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
const WORKTREE_KEY: &str = "0123456789ab";

fn profile() -> Vec<u8> {
    json!({
        "schema_version": 1, "project_id": "demo", "purpose": "fixture", "tools": ["rust"],
        "development_methods": ["sdd"], "hosts": ["claude"],
        "work_flow": {"flow": "m-flow", "integration_branch": "integ"},
    })
    .to_string()
    .into_bytes()
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

/// A commit message with the trailer block of the lane session.
fn signed(subject: &str) -> String {
    format!(
        "{subject}\n\nMoai-Spec: SPEC-20261125-fixture\nMoai-Session: {S_A}:{WORKTREE_KEY}\nMoai-Verified: none"
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

struct World {
    home: TempHome,
    project: PathBuf,
    p: String,
    wt: PathBuf,
}

impl World {
    /// A repository on `integ` (target/ ignored) and the lane's linked worktree on `lane`.
    fn new() -> Self {
        let home = TempHome::new();
        let project = home.project("demo");
        ledger::register(home.path(), "demo").unwrap();
        git(&project, &["init", "-q", "-b", "integ"]);
        fs::write(project.join("a.txt"), "a\n").unwrap();
        fs::write(project.join(".gitignore"), "target/\n").unwrap();
        git(&project, &["add", "a.txt", ".gitignore"]);
        git(&project, &["commit", "-q", "-m", "init"]);
        let wt = home.path().join("lane-wt");
        git(
            &project,
            &[
                "worktree",
                "add",
                "-q",
                "-b",
                "lane",
                wt.to_str().unwrap(),
                "integ",
            ],
        );
        World {
            p: project.to_str().unwrap().to_string(),
            home,
            project,
            wt,
        }
    }

    fn card_env(&self, session: Option<&str>, args: &[&str], env: &[(&str, &Path)]) -> Output {
        let cfg = moai_cli::state_root::path(&self.project, "project.yaml");
        fs::create_dir_all(cfg.parent().unwrap()).unwrap();
        fs::write(&cfg, profile()).unwrap();
        let mut all = vec!["card"];
        all.extend_from_slice(args);
        all.extend_from_slice(&["--project", &self.p]);
        let mut cmd: Command = self.home.command(&all);
        cmd.env("MOAI_GIT_BIN", git_executable());
        for (k, v) in env {
            cmd.env(k, v);
        }
        if let Some(id) = session {
            cmd.env(SESSION_VAR, id);
        }
        cmd.stdin(Stdio::null());
        cmd.output().expect("run moai")
    }

    fn card(&self, session: Option<&str>, args: &[&str]) -> Output {
        self.card_env(session, args, &[])
    }

    /// A task card with the given criteria and scope, taken by the lane in its worktree; its
    /// name and the generation of its reservation.
    fn taken(&self, scope: &str, criteria: &[&str]) -> (String, i64) {
        let mut args = vec![
            "add", "--title", "fixture", "--kind", "task", "--scope", scope,
        ];
        for c in criteria {
            args.push("--ac");
            args.push(c);
        }
        let added = result(&self.card(None, &args));
        let name = added["name"].as_str().unwrap().to_string();
        result(&self.card(None, &["ready", &name]));
        let took = result(&self.card(
            Some(S_A),
            &[
                "take",
                &name,
                "--worktree",
                self.wt.to_str().unwrap(),
                "--branch",
                "lane",
            ],
        ));
        (name, took["reservation"]["generation"].as_i64().unwrap())
    }

    /// Marks criterion `seq` met as the lane, with `evidence`.
    fn met(&self, card: &str, generation: i64, seq: i64, evidence: &str) {
        let (generation, seq) = (generation.to_string(), seq.to_string());
        result(&self.card(
            Some(S_A),
            &[
                "criterion",
                card,
                &seq,
                "--generation",
                &generation,
                "--met",
                "--evidence",
                evidence,
            ],
        ));
    }

    /// Like [`World::taken`], with every criterion already met.
    fn taken_met(&self, scope: &str, criteria: &[&str]) -> (String, i64) {
        let (name, g) = self.taken(scope, criteria);
        for seq in 1..=criteria.len() as i64 {
            self.met(&name, g, seq, "the fixture test");
        }
        (name, g)
    }

    /// A commit on `lane` that sets `path`.
    fn commit(&self, path: &str, content: &str, message: &str) {
        let full = self.wt.join(path);
        fs::create_dir_all(full.parent().unwrap()).unwrap();
        fs::write(&full, content).unwrap();
        git(&self.wt, &["add", path]);
        git(&self.wt, &["commit", "-q", "-m", message]);
    }

    fn head(&self) -> String {
        git(&self.wt, &["rev-parse", "HEAD"])
    }

    fn base(&self) -> String {
        git(&self.project, &["rev-parse", "integ"])
    }

    fn summary_line(&self, result: &str) -> String {
        format!(
            "check-affected base={} head={} targets=3 scans=listed:0 selection=card test-threads=5 diff_check=0 fmt=0 clippy=0 test=0 result={result}",
            &self.base()[..12],
            &self.head()[..12]
        )
    }

    /// The receipt of the head as a worktree that has run `check affected` leaves it.
    fn plant(&self, result: &str) {
        let dir = self
            .wt
            .join(format!("target/check-affected-{}", &self.head()[..12]));
        fs::create_dir_all(&dir).unwrap();
        fs::write(dir.join("summary.txt"), self.summary_line(result) + "\n").unwrap();
    }

    /// A program that stands in for the long check: it records that it ran and writes a receipt
    /// with `result`. Arguments: worktree, base, head, out folder.
    fn seam(&self, result: &str) -> PathBuf {
        let script = self.home.path().join(format!("seam-{result}.sh"));
        let marker = self.marker();
        fs::write(
            &script,
            format!(
                "#!/bin/sh\necho ran >> '{}'\nmkdir -p \"$4\"\nprintf 'check-affected base=%s head=%s targets=1 scans=listed:0 selection=card test-threads=5 diff_check=0 fmt=0 clippy=0 test=0 result={result}\\n' \"$(printf %s \"$2\" | cut -c1-12)\" \"$(printf %s \"$3\" | cut -c1-12)\" > \"$4/summary.txt\"\n",
                marker.display()
            ),
        )
        .unwrap();
        #[cfg(unix)]
        {
            use std::os::unix::fs::PermissionsExt;
            fs::set_permissions(&script, fs::Permissions::from_mode(0o755)).unwrap();
        }
        script
    }

    fn marker(&self) -> PathBuf {
        self.home.path().join("seam-ran")
    }

    fn submit(&self, card: &str, generation: i64, extra: &[&str], seam: Option<&Path>) -> Output {
        let generation = generation.to_string();
        let mut args = vec!["submit", card, "--generation", &generation];
        args.extend_from_slice(extra);
        match seam {
            Some(s) => self.card_env(Some(S_A), &args, &[("MOAI_TEST_SUBMIT_CHECK", s)]),
            None => self.card(Some(S_A), &args),
        }
    }

    fn dump(&self) -> Vec<String> {
        let l = ledger::open_read_only(self.home.path(), "demo").unwrap();
        dump_tables(
            l.conn(),
            &[
                "card",
                "card_ac",
                "card_note",
                "reservation",
                "reservation_scope",
                "event",
            ],
        )
    }

    fn shown(&self, card: &str) -> Value {
        result(&self.card(None, &["show", card]))
    }

    /// Asserts a refusal that wrote nothing: exit 1, `ok` false, a message that holds every
    /// needle and the ledger as it was. Returns the message.
    fn refuse(&self, out: &Output, needles: &[&str], before: &[String]) -> String {
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
        message
    }
}

/// AC-1: with a PASS receipt of the head the command writes the READY note (first line
/// `LANE-READY-INTEGRATION <card> head=<40hex>`), sends the card to review and gives the paths
/// back, and its answer names the card, the head, the tier and the next card.
#[test]
fn a_passing_receipt_writes_the_ready_note_and_sends_the_card_to_review() {
    let w = World::new();
    let (card, g) = w.taken_met("docs/a/", &["one"]);
    w.commit("docs/a/x.txt", "x\n", &signed("the lane change"));
    w.plant("PASS");
    let head = w.head();

    let got = result(&w.submit(&card, g, &[], None));
    assert_eq!(got["head"], json!(head), "{got}");
    assert_eq!(got["tier"], json!(0), "{got}");
    assert_eq!(got["state"], json!("review"), "{got}");
    assert_eq!(got["receipt"]["reused"], json!(true), "{got}");
    assert_eq!(got["given_back"], json!(["docs/a/"]), "{got}");
    assert!(
        got.get("next").is_some(),
        "the answer names the next card: {got}"
    );

    let shown = w.shown(&card);
    assert_eq!(shown["state"], json!("review"), "{shown}");
    let notes = shown["notes"].to_string();
    assert!(
        notes.contains(&format!("LANE-READY-INTEGRATION {card} head={head}")),
        "{notes}"
    );
    for line in [
        "criteria:",
        "1 met: the fixture test",
        "outside-scope: none",
        "tier: 0",
        "leftovers: none",
        "result=PASS",
    ] {
        assert!(notes.contains(line), "the note lacks {line}: {notes}");
    }
}

/// AC-2: a failing cheap gate (here the whitespace check and the trailer check) is reported with
/// every other failing gate, the long check is not started and the ledger is not written.
#[test]
fn a_failing_fast_gate_lists_every_failure_and_starts_nothing() {
    let w = World::new();
    let (card, g) = w.taken_met("docs/a/", &["one"]);
    w.commit(
        "docs/a/ws.txt",
        "trailing \n",
        "a commit with a bad SPEC trailer\n\nMoai-Spec: not-a-spec-id",
    );
    let seam = w.seam("PASS");
    let before = w.dump();

    let out = w.submit(&card, g, &[], Some(&seam));
    w.refuse(&out, &["diff_check", "trailers"], &before);
    assert!(!w.marker().exists(), "the long check was started");
}

/// AC-3 (first half): a tree that is not clean and a criterion that is still open are refused by
/// name, and nothing is written.
#[test]
fn a_dirty_tree_and_an_open_criterion_are_refused_by_name() {
    let w = World::new();
    let (card, g) = w.taken("docs/a/", &["one", "two"]);
    w.met(&card, g, 1, "the fixture test");
    w.commit("docs/a/x.txt", "x\n", &signed("the lane change"));
    w.plant("PASS");
    fs::write(w.wt.join("scratch.txt"), "left over\n").unwrap();
    let before = w.dump();

    let out = w.submit(&card, g, &[], None);
    w.refuse(&out, &["scratch.txt"], &before);
    fs::remove_file(w.wt.join("scratch.txt")).unwrap();

    let out = w.submit(&card, g, &[], None);
    let message = w.refuse(&out, &["criterion 2"], &before);
    assert!(
        message.contains("two"),
        "names the criterion text: {message}"
    );
}

/// AC-3 (second half): a path outside the reservation's scope is not refused; the note names it
/// on its `outside-scope:` line.
#[test]
fn a_path_outside_the_scope_is_written_to_the_note_and_not_refused() {
    let w = World::new();
    let (card, g) = w.taken_met("docs/a/", &["one"]);
    w.commit("docs/a/x.txt", "x\n", &signed("the lane change"));
    w.commit("other/y.txt", "y\n", &signed("a line in another place"));
    w.plant("PASS");

    let got = result(&w.submit(&card, g, &[], None));
    assert_eq!(got["outside_scope"], json!(["other/y.txt"]), "{got}");
    let notes = w.shown(&card)["notes"].to_string();
    assert!(notes.contains("outside-scope: other/y.txt"), "{notes}");
}

/// AC-4: a PASS receipt of the same head and base is quoted and the check is not started again;
/// without one the check runs once and its summary is quoted.
#[test]
fn a_receipt_of_the_same_head_is_quoted_and_without_one_the_check_runs_once() {
    let w = World::new();
    let (card, g) = w.taken_met("docs/a/", &["one"]);
    w.commit("docs/a/x.txt", "x\n", &signed("the lane change"));
    let seam = w.seam("PASS");

    // With a receipt: not run again.
    w.plant("PASS");
    let got = result(&w.submit(&card, g, &[], Some(&seam)));
    assert_eq!(got["receipt"]["reused"], json!(true), "{got}");
    assert!(
        !w.marker().exists(),
        "the check ran although a receipt existed"
    );

    // Without a receipt on a second card: run once, summary quoted.
    let w = World::new();
    let (card2, g2) = w.taken_met("docs/a/", &["one"]);
    w.commit("docs/a/x.txt", "x\n", &signed("the lane change"));
    let seam = w.seam("PASS");
    let got = result(&w.submit(&card2, g2, &[], Some(&seam)));
    assert_eq!(got["receipt"]["reused"], json!(false), "{got}");
    assert_eq!(fs::read_to_string(w.marker()).unwrap().lines().count(), 1);
    let notes = w.shown(&card2)["notes"].to_string();
    assert!(notes.contains("result=PASS"), "{notes}");
    assert_ne!(card, "", "the first card was used");
}

/// A check that does not pass refuses the submission: the summary is named and nothing is
/// written.
#[test]
fn a_failed_check_refuses_and_writes_nothing() {
    let w = World::new();
    let (card, g) = w.taken_met("docs/a/", &["one"]);
    w.commit("docs/a/x.txt", "x\n", &signed("the lane change"));
    let seam = w.seam("FAIL");
    let before = w.dump();

    let out = w.submit(&card, g, &[], Some(&seam));
    w.refuse(&out, &["result=FAIL"], &before);
}

/// The tier is the floor of the changed paths; a tier 2 head needs the lane's guard-mutant lines
/// (`--mutants FILE`) and the note carries them.
#[test]
fn a_tier_two_head_needs_the_guard_mutant_lines_and_the_note_carries_them() {
    let w = World::new();
    let (card, g) = w.taken_met(".github/", &["one"]);
    w.commit(".github/ci.yml", "name: x\n", &signed("a workflow"));
    w.plant("PASS");
    let before = w.dump();

    let out = w.submit(&card, g, &[], None);
    w.refuse(&out, &["--mutants"], &before);

    let file = w.home.path().join("mutants.txt");
    fs::write(
        &file,
        "guard-mutant: test_a | removed the check | exit 100\nguard-mutant: test_b | removed the scan | exit 100\n",
    )
    .unwrap();
    let got = result(&w.submit(&card, g, &["--mutants", file.to_str().unwrap()], None));
    assert_eq!(got["tier"], json!(2), "{got}");
    let notes = w.shown(&card)["notes"].to_string();
    for line in [
        "tier: 2",
        "guard-mutant: test_a | removed the check | exit 100",
        "guard-mutant: test_b | removed the scan | exit 100",
    ] {
        assert!(notes.contains(line), "the note lacks {line}: {notes}");
    }
}
