//! `moai card take` records the worktree as an absolute path, and `moai card repair` lets the
//! holder of a live reservation correct the branch or the worktree path it recorded
//! (SPEC-20261127-reservation-repair). Every test runs the binary on a temporary HOME and a
//! temporary repository whose integration branch is `integ`; the lane works in a linked worktree
//! on `lane`.

mod common;
use common::home::TempHome;
use moai_cli::ledger;
use rusqlite::Connection;
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Output, Stdio};

const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
const S_A: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
const S_B: &str = "bbbbbbbb-0000-4000-8000-00000000000b";
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

fn signed(subject: &str) -> String {
    format!(
        "{subject}\n\nMoai-Spec: SPEC-20261127-fixture\nMoai-Session: {S_A}:{WORKTREE_KEY}\nMoai-Verified: none"
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
    /// The lane's linked worktree on `lane`, a sibling folder of the home's other folders.
    wt: PathBuf,
}

impl World {
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

    /// The top of the lane worktree as git spells it (what a recorded absolute path is).
    fn top(&self) -> String {
        git(&self.wt, &["rev-parse", "--show-toplevel"])
    }

    /// `moai card <args>` as `session` from the folder `cwd` (the test's own folder when `None`).
    fn card_in(&self, session: Option<&str>, cwd: Option<&Path>, args: &[&str]) -> Output {
        let cfg = moai_cli::state_root::path(&self.project, "project.yaml");
        fs::create_dir_all(cfg.parent().unwrap()).unwrap();
        fs::write(&cfg, profile()).unwrap();
        let mut all = vec!["card"];
        all.extend_from_slice(args);
        all.extend_from_slice(&["--project", &self.p]);
        let mut cmd: Command = self.home.command(&all);
        cmd.env("MOAI_GIT_BIN", git_executable());
        if let Some(id) = session {
            cmd.env(SESSION_VAR, id);
        }
        if let Some(dir) = cwd {
            cmd.current_dir(dir);
        }
        cmd.stdin(Stdio::null());
        cmd.output().expect("run moai")
    }

    fn card(&self, session: Option<&str>, args: &[&str]) -> Output {
        self.card_in(session, None, args)
    }

    /// A `ready` task card with one criterion.
    fn ready_card(&self) -> String {
        let added = result(&self.card(
            None,
            &[
                "add", "--title", "fixture", "--kind", "task", "--scope", "docs/a/", "--ac", "one",
            ],
        ));
        let name = added["name"].as_str().unwrap().to_string();
        result(&self.card(None, &["ready", &name]));
        name
    }

    /// A card taken by `S_A` in the lane worktree (an absolute path) on `branch`.
    fn taken(&self, branch: &str) -> (String, i64) {
        let name = self.ready_card();
        let took = result(&self.card(
            Some(S_A),
            &[
                "take",
                &name,
                "--worktree",
                self.wt.to_str().unwrap(),
                "--branch",
                branch,
            ],
        ));
        (name, took["reservation"]["generation"].as_i64().unwrap())
    }

    fn shown(&self, card: &str) -> Value {
        result(&self.card(None, &["show", card]))
    }

    /// The tables a call may write, as text, for a before and after comparison.
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

    /// The payloads of the `card.repair` events, oldest first, with their actor.
    fn repair_events(&self) -> Vec<(String, Value)> {
        let l = ledger::open_read_only(self.home.path(), "demo").unwrap();
        let mut st = l
            .conn()
            .prepare("SELECT actor, payload FROM event WHERE kind = 'card.repair' ORDER BY rowid")
            .unwrap();
        st.query_map([], |r| {
            let payload: String = r.get(1)?;
            Ok((
                r.get::<_, String>(0)?,
                serde_json::from_str(&payload).unwrap(),
            ))
        })
        .unwrap()
        .map(Result::unwrap)
        .collect()
    }

    /// Makes the reservation of `card` record `worktree` as it is, the way a take before
    /// SPEC-20261127-reservation-repair left a relative path.
    fn force_worktree(&self, card: &str, worktree: &str) {
        let id: i64 = card.trim_start_matches("C-").parse().unwrap();
        let mut w = ledger::open_write(self.home.path(), "demo").unwrap();
        w.with_write_tx(|tx, _| {
            tx.execute(
                "UPDATE reservation SET worktree = ?2 WHERE card_id = ?1",
                rusqlite::params![id, worktree],
            )?;
            Ok(())
        })
        .unwrap();
    }

    /// Ends the lease of the reservation of `card` one second ago.
    fn end_lease(&self, card: &str) {
        let id: i64 = card.trim_start_matches("C-").parse().unwrap();
        let mut w = ledger::open_write(self.home.path(), "demo").unwrap();
        w.with_write_tx(|tx, now| {
            tx.execute(
                "UPDATE reservation SET lease_until = ?2 WHERE card_id = ?1",
                rusqlite::params![id, now - 1],
            )?;
            Ok(())
        })
        .unwrap();
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

/// AC-1: a relative `--worktree` is recorded as the absolute path of the checkout's top, resolved
/// against the folder the command runs in; a folder that is missing, not a checkout, or not the
/// top of one is refused with the resolved path named and nothing is written. An absolute path
/// is recorded as it was given.
#[test]
fn take_records_a_relative_worktree_as_an_absolute_path() {
    let w = World::new();
    let card = w.ready_card();
    let took = result(&w.card_in(
        Some(S_A),
        Some(w.home.path()),
        &["take", &card, "--worktree", "lane-wt", "--branch", "lane"],
    ));
    let recorded = took["reservation"]["worktree"].as_str().unwrap();
    assert!(Path::new(recorded).is_absolute(), "{took}");
    assert_eq!(recorded, w.top(), "{took}");
    assert_eq!(w.shown(&card)["reservation"]["worktree"], json!(w.top()));

    // An absolute path is recorded exactly as it was given.
    let other = w.ready_card();
    let given = w.wt.to_str().unwrap();
    let took = result(&w.card_in(
        Some(S_A),
        Some(w.home.path()),
        &["take", &other, "--worktree", given],
    ));
    assert_eq!(took["reservation"]["worktree"], json!(given), "{took}");

    // Refusals: each leaves the ledger as it was and names the folder it resolved to.
    fs::create_dir_all(w.home.path().join("plain")).unwrap();
    fs::create_dir_all(w.wt.join("docs")).unwrap();
    for (rel, why) in [
        ("missing", "is not an existing folder"),
        ("plain", "is not the top of a git checkout"),
        ("lane-wt/docs", "is not the top of a git checkout"),
    ] {
        let card = w.ready_card();
        let before = w.dump();
        let out = w.card_in(
            Some(S_A),
            Some(w.home.path()),
            &["take", &card, "--worktree", rel],
        );
        let message = w.refuse(&out, &[why, rel], &before);
        assert!(
            message.contains(w.home.path().file_name().unwrap().to_str().unwrap()),
            "the resolved absolute path is named: {message}"
        );
        assert_eq!(w.shown(&card)["state"], json!("ready"));
    }
}

/// AC-2: the holder corrects the recorded branch; one `card.repair` event names the old value,
/// the new value, `--by` and `--reason`, and nothing else of the reservation changes.
#[test]
fn repair_changes_only_the_recorded_branch() {
    let w = World::new();
    let (card, g) = w.taken("lane-typo");
    let before = w.shown(&card)["reservation"].clone();
    assert_eq!(before["branch"], json!("lane-typo"));
    let g = g.to_string();

    let done = result(&w.card(
        Some(S_A),
        &[
            "repair",
            &card,
            "--generation",
            &g,
            "--branch",
            "lane",
            "--by",
            "eli",
            "--reason",
            "the branch was cut again under its right name",
        ],
    ));
    assert_eq!(done["action"], json!("repair"), "{done}");
    assert_eq!(done["events"], json!(1), "{done}");

    let after = w.shown(&card)["reservation"].clone();
    assert_eq!(after["branch"], json!("lane"), "{after}");
    let mut same_before = before.clone();
    let mut same_after = after.clone();
    for v in [&mut same_before, &mut same_after] {
        let o = v.as_object_mut().unwrap();
        o.remove("branch");
        o.remove("updated_at");
    }
    assert_eq!(
        same_before, same_after,
        "scope, lease, generation, holder and worktree stay"
    );

    let events = w.repair_events();
    assert_eq!(events.len(), 1, "{events:?}");
    let (actor, payload) = &events[0];
    assert_eq!(actor, "eli");
    assert_eq!(
        payload["branch"],
        json!({"from": "lane-typo", "to": "lane"})
    );
    assert_eq!(payload["by"], json!("eli"), "{payload}");
    assert_eq!(
        payload["reason"],
        json!("the branch was cut again under its right name")
    );
    assert!(payload.get("worktree").is_none(), "{payload}");
    assert_eq!(payload["generation"], json!(g.parse::<i64>().unwrap()));
}

/// AC-3: a reservation that recorded a relative worktree stops `card submit`; `card repair
/// --worktree <absolute path>` makes the same submit run at that path and send the card to review.
#[test]
fn repair_makes_a_relative_reservation_submittable() {
    let w = World::new();
    let (card, g) = w.taken("lane");
    let gen = g.to_string();
    result(&w.card(
        Some(S_A),
        &[
            "criterion",
            &card,
            "1",
            "--generation",
            &gen,
            "--met",
            "--evidence",
            "the fixture test",
        ],
    ));
    fs::create_dir_all(w.wt.join("docs/a")).unwrap();
    fs::write(w.wt.join("docs/a/x.txt"), "x\n").unwrap();
    git(&w.wt, &["add", "docs/a/x.txt"]);
    git(&w.wt, &["commit", "-q", "-m", &signed("the lane change")]);
    let head = git(&w.wt, &["rev-parse", "HEAD"]);
    let base = git(&w.project, &["rev-parse", "integ"]);
    let receipt = w.wt.join(format!("target/check-affected-{}", &head[..12]));
    fs::create_dir_all(&receipt).unwrap();
    fs::write(
        receipt.join("summary.txt"),
        format!(
            "check-affected base={} head={} targets=3 scans=listed:0 selection=card test-threads=5 diff_check=0 fmt=0 clippy=0 test=0 result=PASS\n",
            &base[..12],
            &head[..12]
        ),
    )
    .unwrap();

    // The reservation as an earlier take left it: the worktree is a relative path.
    w.force_worktree(&card, "lane-wt");
    let before = w.dump();
    let out = w.card(Some(S_A), &["submit", &card, "--generation", &gen]);
    let message = w.refuse(&out, &["lane-wt"], &before);
    assert!(
        message.contains("is not a folder") || message.contains("cannot change to"),
        "{message}"
    );

    result(&w.card(
        Some(S_A),
        &[
            "repair",
            &card,
            "--generation",
            &gen,
            "--worktree",
            &w.top(),
            "--by",
            "eli",
            "--reason",
            "the take recorded a relative path",
        ],
    ));
    assert_eq!(w.shown(&card)["reservation"]["worktree"], json!(w.top()));
    let events = w.repair_events();
    assert_eq!(events[0].1["worktree"]["from"], json!("lane-wt"));
    assert_eq!(events[0].1["worktree"]["to"], json!(w.top()));

    let submitted = result(&w.card(Some(S_A), &["submit", &card, "--generation", &gen]));
    assert_eq!(submitted["head"], json!(head), "{submitted}");
    assert_eq!(submitted["state"], json!("review"), "{submitted}");
}

/// AC-4: a session that does not hold the reservation, a stale generation, an ended lease, a
/// card nobody holds, a relative or missing path, a folder that is not the top of a checkout, a
/// branch name git refuses, a call with no value and a bad reason are each refused and the ledger
/// is as it was.
#[test]
fn repair_refusals_write_nothing() {
    let w = World::new();
    let (card, g) = w.taken("lane");
    let gen = g.to_string();
    fs::create_dir_all(w.home.path().join("plain")).unwrap();
    fs::create_dir_all(w.wt.join("docs")).unwrap();
    let top = w.top();
    let docs = w.wt.join("docs");
    let plain = w.home.path().join("plain");
    let long = "x".repeat(501);
    let repair = |session: &str, generation: &str, extra: &[&str]| -> Output {
        let mut args = vec![
            "repair",
            card.as_str(),
            "--generation",
            generation,
            "--by",
            "eli",
        ];
        args.extend_from_slice(extra);
        w.card(Some(session), &args)
    };

    let before = w.dump();
    // Who and which reservation.
    let out = repair(S_B, &gen, &["--branch", "x", "--reason", "r"]);
    w.refuse(&out, &["held by"], &before);
    let out = repair(S_A, "7", &["--branch", "x", "--reason", "r"]);
    w.refuse(&out, &["generation"], &before);
    let nobody = w.ready_card();
    let before_nobody = w.dump();
    let out = w.card(
        Some(S_A),
        &[
            "repair",
            &nobody,
            "--generation",
            "1",
            "--branch",
            "x",
            "--by",
            "eli",
            "--reason",
            "r",
        ],
    );
    w.refuse(&out, &[], &before_nobody);

    // The new values.
    let before = w.dump();
    let out = repair(S_A, &gen, &["--worktree", "lane-wt", "--reason", "r"]);
    w.refuse(&out, &["absolute"], &before);
    let missing = w.home.path().join("missing");
    let out = repair(
        S_A,
        &gen,
        &["--worktree", missing.to_str().unwrap(), "--reason", "r"],
    );
    w.refuse(&out, &["is not an existing folder"], &before);
    let out = repair(
        S_A,
        &gen,
        &["--worktree", plain.to_str().unwrap(), "--reason", "r"],
    );
    w.refuse(&out, &["is not the top of a git checkout"], &before);
    let out = repair(
        S_A,
        &gen,
        &["--worktree", docs.to_str().unwrap(), "--reason", "r"],
    );
    w.refuse(&out, &["is not the top of a git checkout", &top], &before);
    for bad in [
        "has space",
        "a..b",
        "-x",
        "refs/heads/x",
        "x.lock",
        "",
        "a~b",
    ] {
        let out = repair(S_A, &gen, &["--branch", bad, "--reason", "r"]);
        w.refuse(&out, &["--branch"], &before);
    }
    let out = repair(S_A, &gen, &["--reason", "r"]);
    w.refuse(&out, &["--branch", "--worktree"], &before);

    // The reason.
    for bad in ["", "   ", long.as_str(), "password=hunter2hunter2"] {
        let out = repair(S_A, &gen, &["--branch", "x", "--reason", bad]);
        w.refuse(&out, &["--reason"], &before);
    }

    // A lease that has ended: the holder renews first.
    w.end_lease(&card);
    let before = w.dump();
    let out = repair(S_A, &gen, &["--branch", "x", "--reason", "r"]);
    w.refuse(&out, &["lease"], &before);

    assert!(w.repair_events().is_empty(), "no refusal left an event");
}
