//! The commands `ledger-recover-plan` and `ledger-recover-apply` (Issue #3068 P6; AC7 and AC6):
//! what they gather outside the ledger (the claim journal, the worktrees, an observation file),
//! what they refuse, and that a plan writes nothing. Every command runs against a temporary
//! HOME; the real `~/.moai` is not read.

mod common;

use common::git::{git, git_ok};
use common::home::TempHome;
use moai_cli::ledger;
use moai_cli::ledger::recover::WorktreeFact;
use moai_cli::ledger::reservation;
use moai_cli::ledger_recover::{self, Inputs, WorktreeProbe};
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::Output;

const PROJECT: &str = "demo";

fn run(home: &TempHome, args: &[&str]) -> Output {
    home.command(args).output().expect("run moai")
}

fn envelope(out: &Output) -> Value {
    serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
        panic!(
            "stdout is not JSON ({e}): {}",
            String::from_utf8_lossy(&out.stdout)
        )
    })
}

fn result(out: &Output) -> Value {
    assert_eq!(
        out.status.code(),
        Some(0),
        "{}\n{}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    let v = envelope(out);
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn error_of(out: &Output) -> Value {
    assert_eq!(
        out.status.code(),
        Some(1),
        "{}",
        String::from_utf8_lossy(&out.stdout)
    );
    let v = envelope(out);
    assert_eq!(v["ok"], json!(false), "{v}");
    v["error"].clone()
}

fn overwrite_head(path: &Path, bytes: &[u8]) {
    use std::io::Write;
    let mut f = fs::OpenOptions::new().write(true).open(path).unwrap();
    f.write_all(bytes).unwrap();
}

/// A project with a ledger holding two held reservations, then the ledger damaged beyond use
/// (its `.bak` copy is the source):
/// - card 1, `alice`, generation 1, worktree `<home>/work/wt-a` (a real git worktree on branch
///   `issue/1`), branch `issue/1`
/// - card 2, `bob`, generation 1, worktree `<home>/work/plain-b` (a folder, not a worktree),
///   branch `issue/2`
struct Fixture {
    home: TempHome,
    project: PathBuf,
    ledger: PathBuf,
    bak: PathBuf,
    journal: PathBuf,
}

impl Fixture {
    fn new() -> Fixture {
        let home = TempHome::new();
        let project = home.project(PROJECT);
        let wt_a = home.path().join("work/wt-a");
        fs::create_dir_all(&wt_a).unwrap();
        git_ok(&wt_a, &["init", "-q", "-b", "issue/1"]);
        git_ok(
            &wt_a,
            &[
                "-c",
                "user.name=t",
                "-c",
                "user.email=t@example.invalid",
                "commit",
                "-q",
                "--allow-empty",
                "-m",
                "x",
            ],
        );
        let plain_b = home.path().join("work/plain-b");
        fs::create_dir_all(&plain_b).unwrap();

        ledger::register(home.path(), PROJECT).unwrap();
        let mut w = ledger::open_write(home.path(), PROJECT).unwrap();
        for _ in 0..2 {
            w.with_write_tx(|tx, now| {
                tx.execute(
                    "INSERT INTO card (title, state, created_by, created_at, updated_at) \
                     VALUES ('c', 'ready', 't', ?1, ?1)",
                    [now],
                )?;
                Ok(())
            })
            .unwrap();
        }
        reservation::reserve(
            &mut w,
            1,
            "alice",
            &["zone1/".to_string()],
            3600,
            Some(wt_a.to_str().unwrap()),
            Some("issue/1"),
        )
        .unwrap();
        reservation::reserve(
            &mut w,
            2,
            "bob",
            &["zone2/".to_string()],
            3600,
            Some(plain_b.to_str().unwrap()),
            Some("issue/2"),
        )
        .unwrap();
        drop(w);
        let ledger = ledger::ledger_path(home.path(), PROJECT).unwrap();
        let bak = home.path().join("sources/ledger-copy.db");
        fs::create_dir_all(bak.parent().unwrap()).unwrap();
        fs::copy(&ledger, &bak).unwrap();
        overwrite_head(&ledger, b"NOT A DATABASE!!");

        // A claim journal folder with both holders' claims, leases ending 10:00Z.
        let journal = home.path().join("journal");
        fs::create_dir_all(&journal).unwrap();
        let line = |holder: &str, key: &str| {
            json!({"event": "assigned", "record": {
                "schema_version": 1,
                "task_ref": {"provider": "github", "host": "github.com", "owner": "o",
                    "repo": "r", "repository_id": 1, "issue_id": 7, "issue_number": 7},
                "task_key": key,
                "logical_key": "l",
                "owner": {"kind": "session", "id": holder},
                "epoch": 1,
                "base_revision": "b",
                "worktree": {"common_dir": "/x/.git", "git_dir": "/x/.git", "path": "/x", "head": "b"},
                "write_scope": [],
                "spec_ref": {"id": "S", "revision": "r"},
                "lease_until": "2026-11-16T10:00:00Z",
                "state": "running",
                "effects": [],
                "last_observed_at": "2026-11-16T08:00:00Z"
            }})
            .to_string()
        };
        fs::write(
            journal.join("journal.log"),
            format!(
                "{}\n{}\n",
                line("alice", "aaaaaaaaaaaaaaaa"),
                line("bob", "bbbbbbbbbbbbbbbb")
            ),
        )
        .unwrap();
        Fixture {
            home,
            project,
            ledger,
            bak,
            journal,
        }
    }

    fn remotes_file(&self, remotes: Value) -> PathBuf {
        let path = self.home.path().join("remote-observations.json");
        fs::write(&path, json!({ "remotes": remotes }).to_string()).unwrap();
        path
    }

    fn args<'a>(&'a self, verb: &'a str, extra: &[&'a str]) -> Vec<&'a str> {
        let mut v = vec![
            verb,
            self.project.to_str().unwrap(),
            "--source",
            self.bak.to_str().unwrap(),
        ];
        v.extend_from_slice(extra);
        v
    }

    fn statuses(&self) -> Vec<(i64, String)> {
        let r = ledger::open_read_only(self.home.path(), PROJECT).unwrap();
        let mut stmt = r
            .conn()
            .prepare("SELECT card_id, status FROM reservation ORDER BY reservation_id")
            .unwrap();
        let rows: Vec<(i64, String)> = stmt
            .query_map([], |row| Ok((row.get(0)?, row.get(1)?)))
            .unwrap()
            .map(Result::unwrap)
            .collect();
        rows
    }
}

const NOW_EARLY: &str = "2026-11-16T09:00:00Z";
const NOW_LATE: &str = "2026-11-16T11:00:00Z";

const REMOTES: fn() -> Value = || {
    json!([
        {"branch": "issue/1", "exists": true, "holder": "alice", "generation": 1, "card_id": 1},
        {"branch": "issue/2", "exists": true}
    ])
};

// ---- plan and apply end to end ------------------------------------------------------------------

#[test]
fn a_reservation_comes_back_held_only_with_journal_worktree_and_remote_all_confirming() {
    let f = Fixture::new();
    let obs = f.remotes_file(REMOTES());
    let journal = f.journal.to_str().unwrap();
    let obs_arg = obs.to_str().unwrap();
    let common_args = [
        "--journal",
        journal,
        "--observations",
        obs_arg,
        "--now",
        NOW_EARLY,
    ];
    let before = f.home.snapshot();
    let plan = result(&run(&f.home, &f.args("ledger-recover-plan", &common_args)));
    assert_eq!(f.home.snapshot(), before, "the plan wrote nothing");
    assert_eq!(plan["writes"], json!(false));
    assert_eq!(plan["network"], json!("disabled"));
    assert_eq!(plan["restore_held"], json!(1));
    assert_eq!(plan["held_for_review"], json!(1));
    let by_holder = |h: &str| -> Value {
        plan["decisions"]
            .as_array()
            .unwrap()
            .iter()
            .find(|d| d["holder"] == json!(h))
            .unwrap()
            .clone()
    };
    assert_eq!(by_holder("alice")["decision"], json!("restore_held"));
    assert_eq!(by_holder("alice")["reasons"], json!([]));
    assert_eq!(by_holder("bob")["decision"], json!("held_for_review"));
    assert_eq!(
        by_holder("bob")["reasons"],
        json!(["worktree_missing"]),
        "bob's journal and remote agree; his worktree is a plain folder"
    );
    assert_eq!(plan["observed"]["journal"]["records"], json!(2));
    assert_eq!(plan["observed"]["worktrees"][0]["branch"], json!("issue/1"));
    assert_eq!(plan["target_state"], json!("unavailable"));
    let digest = plan["plan_digest"].as_str().unwrap().to_string();

    // The same inputs give the same digest.
    let again = result(&run(&f.home, &f.args("ledger-recover-plan", &common_args)));
    assert_eq!(again["plan_digest"], json!(digest));

    // A later `--now` is another plan: alice's lease has ended.
    let late = result(&run(
        &f.home,
        &f.args(
            "ledger-recover-plan",
            &[
                "--journal",
                journal,
                "--observations",
                obs_arg,
                "--now",
                NOW_LATE,
            ],
        ),
    ));
    assert_eq!(late["restore_held"], json!(0));
    assert_ne!(late["plan_digest"], json!(digest));

    // Apply: a stale digest, a bad `--now`, a missing permission reference are all refused and
    // leave the HOME as it was.
    let stale_args = [
        "--journal",
        journal,
        "--observations",
        obs_arg,
        "--now",
        NOW_LATE,
        "--plan-digest",
        digest.as_str(),
        "--permission-reference",
        "ref-1",
    ];
    let err = error_of(&run(&f.home, &f.args("ledger-recover-apply", &stale_args)));
    assert!(
        err["message"].as_str().unwrap().contains("stale_plan"),
        "{err}"
    );
    assert_eq!(err["kind"], json!("invalid_input"));
    let no_permission = [
        "--journal",
        journal,
        "--observations",
        obs_arg,
        "--now",
        NOW_EARLY,
        "--plan-digest",
        digest.as_str(),
        "--permission-reference",
        "",
    ];
    let err = error_of(&run(
        &f.home,
        &f.args("ledger-recover-apply", &no_permission),
    ));
    assert!(
        err["message"]
            .as_str()
            .unwrap()
            .contains("permission reference"),
        "{err}"
    );
    let bad_now = [
        "--now",
        "yesterday",
        "--plan-digest",
        digest.as_str(),
        "--permission-reference",
        "ref-1",
    ];
    let out = run(&f.home, &f.args("ledger-recover-apply", &bad_now));
    assert_eq!(out.status.code(), Some(1));
    // The refusals changed nothing but the recovery lock file the stale-plan attempt created
    // (kept, empty): the ledger is as it was and no scratch or kept file appeared.
    let project_dir = f.ledger.parent().unwrap();
    let listed = common::home::TempHome::names_in(project_dir);
    assert_eq!(
        listed,
        vec![".recover.lock".to_string(), "ledger.db".to_string()]
    );
    assert_eq!(
        fs::metadata(project_dir.join(".recover.lock"))
            .unwrap()
            .len(),
        0
    );

    // Apply with the preview's inputs.
    let good = [
        "--journal",
        journal,
        "--observations",
        obs_arg,
        "--now",
        NOW_EARLY,
        "--plan-digest",
        digest.as_str(),
        "--permission-reference",
        "ref-1",
    ];
    let applied = result(&run(&f.home, &f.args("ledger-recover-apply", &good)));
    assert_eq!(applied["state"], json!("applied"));
    assert_eq!(applied["built"]["integrity_check"], json!("ok"));
    assert!(applied["set_aside"]["damaged"]
        .as_str()
        .unwrap()
        .contains(".damaged-"));
    assert_eq!(
        f.statuses(),
        vec![(1, "held".to_string()), (2, "held_for_review".to_string())]
    );
    // A second apply answers `unchanged` and writes nothing.
    let after = f.home.snapshot();
    let second = result(&run(&f.home, &f.args("ledger-recover-apply", &good)));
    assert_eq!(second["state"], json!("unchanged"));
    assert_eq!(f.home.snapshot(), after);

    // The status of the recovered ledger is ready and reports no leftovers.
    let status = result(&run(
        &f.home,
        &["ledger-status", f.project.to_str().unwrap()],
    ));
    assert_eq!(status["state"], json!("ready"));
    assert_eq!(status["stale_temp"], json!([]));
    assert_eq!(status["writes"], json!(false));
}

#[test]
fn without_a_journal_or_remote_observation_every_active_reservation_is_held_for_review() {
    let f = Fixture::new();
    // The project is not a git repository, so there is no default journal either.
    let plan = result(&run(&f.home, &f.args("ledger-recover-plan", &[])));
    assert_eq!(plan["restore_held"], json!(0));
    assert_eq!(plan["held_for_review"], json!(2));
    assert!(plan["observed"]["journal"]["error"].is_string());
    let alice = plan["decisions"]
        .as_array()
        .unwrap()
        .iter()
        .find(|d| d["holder"] == json!("alice"))
        .unwrap();
    assert_eq!(
        alice["reasons"],
        json!(["journal_missing", "remote_absent"]),
        "alice's worktree is real; the journal and the remote cannot be confirmed"
    );
}

// ---- refusals -----------------------------------------------------------------------------------

#[test]
fn a_source_that_is_a_link_or_the_wrong_project_is_refused_and_nothing_is_written() {
    let f = Fixture::new();
    // A link needs `std::os::unix`; the wrong-project refusal below is checked everywhere.
    #[cfg(unix)]
    let link = f.home.path().join("sources/link.db");
    #[cfg(unix)]
    std::os::unix::fs::symlink(&f.bak, &link).unwrap();
    let other = f.home.project("other");
    let before = f.home.snapshot();
    #[cfg(unix)]
    {
        let out = run(
            &f.home,
            &[
                "ledger-recover-plan",
                f.project.to_str().unwrap(),
                "--source",
                link.to_str().unwrap(),
            ],
        );
        let err = error_of(&out);
        assert!(
            err["message"]
                .as_str()
                .unwrap()
                .contains("ledger_unavailable"),
            "{err}"
        );
        assert_eq!(err["kind"], json!("blocked"));
    }
    // A project whose id differs from the file's.
    let out = run(
        &f.home,
        &[
            "ledger-recover-plan",
            other.to_str().unwrap(),
            "--source",
            f.bak.to_str().unwrap(),
        ],
    );
    let err = error_of(&out);
    assert!(
        err["message"].as_str().unwrap().contains("id_collision"),
        "{err}"
    );
    assert_eq!(f.home.snapshot(), before, "both refusals wrote nothing");
}

#[test]
fn a_go_generation_home_blocks_the_recovery() {
    let f = Fixture::new();
    f.home.make_go_generation();
    let out = run(&f.home, &f.args("ledger-recover-plan", &[]));
    let err = error_of(&out);
    assert!(
        err["message"].as_str().unwrap().contains("home_not_ready"),
        "{err}"
    );
}

#[test]
fn the_observation_file_has_a_size_limit_and_a_closed_shape() {
    let f = Fixture::new();
    let limit = ledger_recover::MAX_OBSERVATION_FILE_BYTES as usize;
    let pad = |n: usize| {
        let mut text = String::from("{\"remotes\":[]}");
        text.push_str(&" ".repeat(n - text.len()));
        text
    };
    for (len, ok) in [(limit, true), (limit + 1, false)] {
        let path = f.home.path().join("big.json");
        fs::write(&path, pad(len)).unwrap();
        let out = run(
            &f.home,
            &f.args(
                "ledger-recover-plan",
                &["--observations", path.to_str().unwrap()],
            ),
        );
        if ok {
            result(&out);
        } else {
            let err = error_of(&out);
            assert!(
                err["message"].as_str().unwrap().contains("at most"),
                "{err}"
            );
        }
    }
    for (name, body) in [
        ("unknown.json", "{\"journals\":[]}"),
        ("notjson.json", "nope"),
        ("array.json", "[]"),
        ("badfact.json", "{\"remotes\":[{\"branch\":\"b\"}]}"),
    ] {
        let path = f.home.path().join(name);
        fs::write(&path, body).unwrap();
        let out = run(
            &f.home,
            &f.args(
                "ledger-recover-plan",
                &["--observations", path.to_str().unwrap()],
            ),
        );
        assert_eq!(out.status.code(), Some(1), "{name}");
        let err = error_of(&out);
        assert!(
            err["message"]
                .as_str()
                .unwrap()
                .contains("observation file"),
            "{name}: {err}"
        );
    }
    let err = error_of(&run(
        &f.home,
        &f.args(
            "ledger-recover-plan",
            &[
                "--observations",
                f.home.path().join("missing.json").to_str().unwrap(),
            ],
        ),
    ));
    assert!(
        err["message"]
            .as_str()
            .unwrap()
            .contains("observation file"),
        "{err}"
    );
}

#[test]
fn a_status_reports_the_leftover_of_a_dead_register_and_removes_nothing() {
    let f = Fixture::new();
    // Plain files whose pid cannot belong to a running process on this machine (above any pid
    // limit) are not stale; a stale one needs a dead pid, which the library test makes.
    let dir = f.ledger.parent().unwrap();
    fs::write(dir.join(".creating-4294967295-0"), b"x").unwrap();
    let status = result(&run(
        &f.home,
        &["ledger-status", f.project.to_str().unwrap()],
    ));
    assert_eq!(status["stale_temp"], json!([]));
    assert_eq!(
        status["state"],
        json!("unavailable"),
        "the ledger is damaged"
    );
    assert!(dir.join(".creating-4294967295-0").exists());

    // A process that ran and exited: its pid names a stale leftover, reported by status and by
    // the plan, and still there afterwards.
    let mut child = std::process::Command::new(std::env::current_exe().unwrap())
        .args(["--exact", "no_such_test_exists"])
        .stdout(std::process::Stdio::null())
        .stderr(std::process::Stdio::null())
        .spawn()
        .unwrap();
    let pid = child.id();
    child.wait().unwrap();
    let name = format!(".creating-{pid}-3");
    fs::write(dir.join(&name), b"half a ledger").unwrap();
    let before = f.home.snapshot();
    let status = result(&run(
        &f.home,
        &["ledger-status", f.project.to_str().unwrap()],
    ));
    assert_eq!(status["stale_temp"], json!([{"name": name, "pid": pid}]));
    let plan = result(&run(&f.home, &f.args("ledger-recover-plan", &[])));
    assert_eq!(plan["stale_temp"], json!([{"name": name, "pid": pid}]));
    assert_eq!(
        f.home.snapshot(),
        before,
        "reporting removed and changed nothing"
    );
    assert!(dir.join(&name).exists());
}

// ---- the layer between: injected facts and the real git probe ------------------------------------

struct Fake(Vec<WorktreeFact>);

impl WorktreeProbe for Fake {
    fn probe(&self, path: &str) -> WorktreeFact {
        self.0
            .iter()
            .find(|w| w.path == path)
            .cloned()
            .unwrap_or(WorktreeFact {
                path: path.to_string(),
                exists: false,
                branch: None,
                clean: false,
            })
    }
}

#[test]
fn injected_worktree_facts_decide_like_real_ones() {
    let f = Fixture::new();
    let obs = f.remotes_file(REMOTES());
    let now = moai_cli::forge::policy::parse_utc(NOW_EARLY).unwrap();
    let facts = vec![
        WorktreeFact {
            path: f.home.path().join("work/wt-a").display().to_string(),
            exists: true,
            branch: Some("issue/1".into()),
            clean: true,
        },
        WorktreeFact {
            path: f.home.path().join("work/plain-b").display().to_string(),
            exists: true,
            branch: Some("issue/2".into()),
            clean: false,
        },
    ];
    let inputs = Inputs {
        project: &f.project,
        home: f.home.path(),
        source: &f.bak,
        journal: Some(&f.journal),
        observations: Some(&obs),
        now,
    };
    let plan = ledger_recover::plan(&inputs, &Fake(facts)).unwrap();
    assert_eq!(plan["restore_held"], json!(2), "both confirmed three ways");
    // `clean` is informational: bob's dirty worktree does not change the decision.
    assert_eq!(plan["observed"]["worktrees"][1]["clean"], json!(false));
    // A worktree the machine lacks can be supplied by the observation file; one it has cannot
    // be overridden.
    let file = f.home.path().join("wt.json");
    let plain_b = f.home.path().join("work/plain-b").display().to_string();
    fs::write(
        &file,
        json!({
            "remotes": REMOTES(),
            "worktrees": [{"path": plain_b, "exists": true, "branch": "issue/2"}],
        })
        .to_string(),
    )
    .unwrap();
    let inputs = Inputs {
        observations: Some(&file),
        ..inputs
    };
    let none = Fake(vec![]);
    let plan = ledger_recover::plan(&inputs, &none).unwrap();
    let reasons = |h: &str| -> Value {
        plan["decisions"]
            .as_array()
            .unwrap()
            .iter()
            .find(|d| d["holder"] == json!(h))
            .unwrap()["reasons"]
            .clone()
    };
    assert_eq!(reasons("alice"), json!(["worktree_missing"]));
    assert_eq!(
        reasons("bob"),
        json!([]),
        "the file supplied the missing one"
    );
    let local = Fake(vec![WorktreeFact {
        path: plain_b,
        exists: true,
        branch: Some("elsewhere".into()),
        clean: true,
    }]);
    let plan = ledger_recover::plan(&inputs, &local).unwrap();
    assert_eq!(
        plan["decisions"]
            .as_array()
            .unwrap()
            .iter()
            .find(|d| d["holder"] == json!("bob"))
            .unwrap()["reasons"],
        json!(["branch_mismatch"]),
        "a worktree the machine has is judged by the machine, not the file"
    );
}

#[test]
fn the_git_probe_reports_a_worktree_only_at_its_own_top() {
    use ledger_recover::GitProbe;
    let home = TempHome::new();
    let top = home.path().join("repo");
    fs::create_dir_all(top.join("sub")).unwrap();
    git_ok(&top, &["init", "-q", "-b", "feature/x"]);
    let commit = [
        "-c",
        "user.name=t",
        "-c",
        "user.email=t@example.invalid",
        "commit",
        "-q",
        "--allow-empty",
        "-m",
        "x",
    ];
    git_ok(&top, &commit);
    let probe = |p: &Path| GitProbe.probe(p.to_str().unwrap());

    let fact = probe(&top);
    assert!(fact.exists && fact.clean, "{fact:?}");
    assert_eq!(fact.branch.as_deref(), Some("feature/x"));
    // A folder inside the worktree is not the worktree.
    assert!(!probe(&top.join("sub")).exists);
    // A missing path, a relative path, a plain file.
    assert!(!probe(&home.path().join("nowhere")).exists);
    assert!(!GitProbe.probe("relative/path").exists);
    fs::write(top.join("file"), "x").unwrap();
    assert!(!probe(&top.join("file")).exists);
    // An untracked file makes it unclean but not absent.
    let dirty = probe(&top);
    assert!(dirty.exists && !dirty.clean);
    // A detached HEAD has no branch.
    let head = git(&top, &["rev-parse", "HEAD"]);
    git_ok(&top, &["checkout", "-q", "--detach", &head]);
    let detached = probe(&top);
    assert!(detached.exists);
    assert_eq!(detached.branch, None);
    // A real linked worktree is found at its own path.
    git_ok(
        &top,
        &[
            "worktree",
            "add",
            "-q",
            "-b",
            "other",
            home.path().join("linked").to_str().unwrap(),
        ],
    );
    let linked = probe(&home.path().join("linked"));
    assert!(linked.exists);
    assert_eq!(linked.branch.as_deref(), Some("other"));
}

#[test]
fn the_journal_facts_follow_the_last_record_of_each_claim() {
    let dir = TempHome::new();
    let root = dir.path().join("journal");
    fs::create_dir_all(&root).unwrap();
    let record = |holder: &str, key: &str, epoch: u64, state: &str, lease: &str| {
        json!({"event": "x", "record": {
            "schema_version": 1,
            "task_ref": {"provider": "github", "host": "github.com", "owner": "o",
                "repo": "r", "repository_id": 1, "issue_id": 7, "issue_number": 7},
            "task_key": key, "logical_key": "l",
            "owner": {"kind": "session", "id": holder}, "epoch": epoch,
            "base_revision": "b",
            "worktree": {"common_dir": "/x", "git_dir": "/x", "path": "/x", "head": "b"},
            "write_scope": [], "spec_ref": {"id": "S", "revision": "r"},
            "lease_until": lease, "state": state, "effects": [],
            "last_observed_at": "2026-11-16T08:00:00Z"
        }})
        .to_string()
    };
    let lines = [
        // claim A: first held by x (epoch 1), later released: not live
        record(
            "x",
            "aaaaaaaaaaaaaaaa",
            1,
            "running",
            "2026-11-16T10:00:00Z",
        ),
        record(
            "x",
            "aaaaaaaaaaaaaaaa",
            1,
            "released",
            "2026-11-16T10:00:00Z",
        ),
        // claim B: live
        record(
            "y",
            "bbbbbbbbbbbbbbbb",
            4,
            "running",
            "2026-11-16T10:00:00Z",
        ),
        // claim C: lease ends exactly at `now` (the lease-end second itself is expired)
        record(
            "z",
            "cccccccccccccccc",
            2,
            "running",
            "2026-11-16T09:00:00Z",
        ),
    ];
    fs::write(root.join("journal.log"), lines.join("\n") + "\n").unwrap();
    let now = moai_cli::forge::policy::parse_utc("2026-11-16T09:00:00Z").unwrap();
    let mut facts = ledger_recover::journal_facts(&root, now).unwrap();
    facts.sort_by(|a, b| a.holder.cmp(&b.holder));
    let shown: Vec<(String, i64, bool)> = facts
        .into_iter()
        .map(|f| (f.holder, f.generation, f.live))
        .collect();
    assert_eq!(
        shown,
        vec![
            ("x".to_string(), 1, false),
            ("y".to_string(), 4, true),
            ("z".to_string(), 2, false),
        ]
    );
    // One second earlier the lease of claim C still holds.
    let facts = ledger_recover::journal_facts(&root, now - 1).unwrap();
    assert!(facts.iter().any(|f| f.holder == "z" && f.live));
    // No journal log at all is no facts, not an error.
    let empty = dir.path().join("empty");
    fs::create_dir_all(&empty).unwrap();
    assert_eq!(ledger_recover::journal_facts(&empty, now).unwrap(), vec![]);
}
