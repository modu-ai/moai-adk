//! SPEC-20261123-cleanup-outbox: a closed card queues its worktree and branch cleanup in the
//! ledger in the same transaction that closes it, and an idempotent executor removes them after a
//! fresh re-check.
//!
//! Every test runs on a temporary HOME with a temporary git repository; nothing of the real
//! user state folder, the real checkout or its worktrees and branches is read or written. The closing
//! commands run as the `moai` binary; the executor's own failure paths (a stop in the middle, a
//! second executor, a branch another worktree takes meanwhile) run through the library function
//! with an `Options` value, so no test needs an environment variable that a parallel test could
//! disturb.
mod common;

use common::home::TempHome;
use moai_cli::card::cleanup::{self, Options};
use moai_cli::ledger;
use moai_cli::ledger::cleanup_outbox::{self, Claim, Finish, Outcome, Step};
use serde_json::{json, Value};
use std::cell::RefCell;
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Output, Stdio};
use std::sync::Once;

const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
const S_A: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
const S_EXEC: &str = "eeeeeeee-0000-4000-8000-0000000000ee";
/// The independent reviewer of the PASS a landing relies on.
const S_AUDITOR: &str = "cccccccc-0000-4000-8000-00000000000c";

/// The profile of an m-flow project whose integration branch is `integ`.
const MFLOW_PROFILE: &[u8] = b"{\"schema_version\": 1, \"project_id\": \"demo\", \"purpose\": \"fixture\", \"tools\": [\"rust\"], \"development_methods\": [\"sdd\"], \"hosts\": [\"claude\"], \"work_flow\": {\"flow\": \"m-flow\", \"integration_branch\": \"integ\"}}";
/// The same project with no work_flow: no integration branch to measure "merged" against.
const PLAIN_PROFILE: &[u8] = b"{\"schema_version\": 1, \"project_id\": \"demo\", \"purpose\": \"fixture\", \"tools\": [\"rust\"], \"development_methods\": [\"sdd\"], \"hosts\": [\"claude\"]}";

static CLOCK: Once = Once::new();

/// The cleanup clock three days ahead for this whole test process, so every change in a fixture
/// is old. One value for every test: the variable is set once, before any test reads it.
fn clock_three_days_on() -> String {
    CLOCK.call_once(|| {
        let now = std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap()
            .as_secs();
        std::env::set_var(
            moai_cli::worktree_cleanup::NOW_SEAM,
            (now + 3 * 86_400).to_string(),
        );
    });
    std::env::var(moai_cli::worktree_cleanup::NOW_SEAM).unwrap()
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

/// A commit with the tree of `tree_of`, the given parents in order and one message; no ref moves.
fn commit_with(project: &Path, tree_of: &str, parents: &[&str], message: &str) -> String {
    let tree = git(project, &["rev-parse", &format!("{tree_of}^{{tree}}")]);
    let mut args = vec!["commit-tree", tree.as_str()];
    for parent in parents {
        args.push("-p");
        args.push(parent);
    }
    args.push("-m");
    args.push(message);
    git(project, &args)
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

fn number(name: &str) -> i64 {
    name.trim_start_matches("C-").parse().unwrap()
}

struct World {
    home: TempHome,
    dir: tempfile::TempDir,
    project: PathBuf,
    p: String,
    git_bin: PathBuf,
    extra_env: RefCell<Vec<(String, String)>>,
    profile: RefCell<Vec<u8>>,
    /// The folder the `moai` commands run in; `None` is the test process's own.
    cwd: RefCell<Option<PathBuf>>,
    head: String,
}

impl World {
    fn new() -> Self {
        let now = clock_three_days_on();
        let home = TempHome::new();
        let project = home.project("demo");
        ledger::register(home.path(), "demo").unwrap();
        git(&project, &["init", "-q", "-b", "main"]);
        fs::write(project.join("a.txt"), "a\n").unwrap();
        git(&project, &["add", "a.txt"]);
        git(&project, &["commit", "-q", "-m", "init"]);
        fs::write(project.join("b.txt"), "b\n").unwrap();
        git(&project, &["add", "b.txt"]);
        // Signed by the lane session: a close on a landed head reads the verdict of the head, and
        // that check needs the authors of the head.
        git(
            &project,
            &[
                "commit",
                "-q",
                "-m",
                &format!("the card's change\n\nMoai-Session: {S_A}:0123456789ab"),
            ],
        );
        let head = git(&project, &["rev-parse", "HEAD"]);
        let tip = commit_with(&project, &head, &[&head], "integration moved on");
        git(&project, &["update-ref", "refs/heads/integ", &tip]);
        let world = World {
            p: project.to_str().unwrap().to_string(),
            home,
            dir: tempfile::Builder::new()
                .prefix("moai-test-outbox-")
                .tempdir()
                .unwrap(),
            project,
            git_bin: git_executable(),
            extra_env: RefCell::new(vec![
                (moai_cli::worktree_cleanup::NOW_SEAM.to_string(), now),
                // The cleanup's process check runs `lsof`, which it finds on the caller's PATH.
                (
                    "PATH".to_string(),
                    std::env::var("PATH").unwrap_or_default(),
                ),
            ]),
            profile: RefCell::new(MFLOW_PROFILE.to_vec()),
            cwd: RefCell::new(None),
            head,
        };
        world.write_profile();
        world
    }

    fn write_profile(&self) {
        let cfg = moai_cli::state_root::path(&self.project, "project.yaml");
        fs::create_dir_all(cfg.parent().unwrap()).unwrap();
        fs::write(&cfg, &*self.profile.borrow()).unwrap();
    }

    /// `moai <args>` as `session`, with the git program named.
    fn run(&self, session: Option<&str>, args: &[&str]) -> Output {
        let mut cmd: Command = self.home.command(args);
        cmd.env("MOAI_GIT_BIN", &self.git_bin);
        for (key, value) in self.extra_env.borrow().iter() {
            cmd.env(key, value);
        }
        if let Some(id) = session {
            cmd.env(SESSION_VAR, id);
        }
        if let Some(dir) = self.cwd.borrow().as_ref() {
            cmd.current_dir(dir);
        }
        cmd.stdin(Stdio::null());
        cmd.output().expect("run moai")
    }

    /// `moai card <args> --project <project>` with the profile written last.
    fn card(&self, session: Option<&str>, args: &[&str]) -> Output {
        self.write_profile();
        let mut all = vec!["card"];
        all.extend_from_slice(args);
        all.extend_from_slice(&["--project", &self.p]);
        self.run(session, &all)
    }

    /// A folder of the fixture, outside the project.
    fn outside(&self, name: &str) -> PathBuf {
        fs::canonicalize(self.dir.path()).unwrap().join(name)
    }

    /// A linked worktree of the project, detached at the card's change (which the tip contains),
    /// in a folder of its own outside the project.
    fn linked_worktree(&self, name: &str) -> PathBuf {
        let at = self.outside(name);
        git(
            &self.project,
            &[
                "worktree",
                "add",
                "-q",
                "--detach",
                at.to_str().unwrap(),
                &self.head,
            ],
        );
        at
    }

    /// A branch at the card's change (merged into the tip).
    fn branch(&self, name: &str) {
        git(&self.project, &["branch", name, &self.head]);
    }

    fn has_branch(&self, name: &str) -> bool {
        !git(
            &self.project,
            &["for-each-ref", &format!("refs/heads/{name}")],
        )
        .is_empty()
    }

    fn registered(&self, dir: &Path) -> bool {
        git(&self.project, &["worktree", "list", "--porcelain"])
            .contains(dir.file_name().unwrap().to_str().unwrap())
    }

    /// A task card with `criteria`, its own scope path, taken by `S_A` in `worktree` (on `branch`
    /// when given); returns its name. The generation is 1.
    fn card_in(
        &self,
        scope: &str,
        worktree: &Path,
        branch: Option<&str>,
        criteria: &[&str],
    ) -> String {
        let mut args = vec![
            "add",
            "--title",
            "fixture card",
            "--kind",
            "task",
            "--scope",
            scope,
        ];
        for text in criteria {
            args.push("--ac");
            args.push(text);
        }
        let name = result(&self.card(None, &args))["name"]
            .as_str()
            .unwrap()
            .to_string();
        result(&self.card(None, &["ready", &name]));
        let mut take = vec![
            "take",
            name.as_str(),
            "--worktree",
            worktree.to_str().unwrap(),
        ];
        if let Some(b) = branch {
            take.extend(["--branch", b]);
        }
        result(&self.card(Some(S_A), &take));
        name
    }

    /// A candidate on the current integration tip: a merge commit with the tip as its first
    /// parent and a lane commit as its second.
    fn candidate_now(&self) -> String {
        let base = self.integ();
        let lane = commit_with(
            &self.project,
            &self.head,
            &[&self.head],
            &format!("the lane head\n\nMoai-Session: {S_A}:0123456789ab"),
        );
        commit_with(
            &self.project,
            &base,
            &[&base, &lane],
            "merge the lane head on the tip",
        )
    }

    /// One independent tier 0 PASS of `head` written as `card verdict` writes it: a landing that
    /// closes a card relies on it (SPEC-20261124-land-closes-card).
    fn pass_of(&self, card: &str, head: &str) {
        let base = git(&self.project, &["rev-parse", &format!("{head}^1")]);
        let conn =
            rusqlite::Connection::open(ledger::ledger_path(self.home.path(), "demo").unwrap())
                .unwrap();
        let payload = json!({
            "head": head, "base": base, "verdict": "pass", "tier": 0,
            "evidence": "the fixture's evidence", "receipt": null, "reviewer": S_AUDITOR,
        });
        conn.execute(
            "INSERT INTO event (event_key, at, actor, card_id, kind, payload) \
             VALUES (?1, unixepoch(), ?2, ?3, 'card.verdict', ?4)",
            rusqlite::params![
                format!("test-verdict-{card}-{head}"),
                S_AUDITOR,
                number(card),
                payload.to_string()
            ],
        )
        .unwrap();
    }

    /// [`World::pass_of`] for a head that changes code (`b.txt`): judged at tier 1, so the PASS
    /// cites a green receipt of the head. `card finish` on a landed head reads it
    /// (SPEC-20261125-integration-guard-install).
    fn pass_of_code(&self, card: &str, head: &str) {
        let base = git(&self.project, &["rev-parse", &format!("{head}^1")]);
        let conn =
            rusqlite::Connection::open(ledger::ledger_path(self.home.path(), "demo").unwrap())
                .unwrap();
        let payload = json!({
            "head": head, "base": base, "verdict": "pass", "tier": 1,
            "evidence": "the fixture's evidence",
            "receipt": {"head": head, "steps": [{"name": "test", "exit": 0}]},
            "reviewer": S_AUDITOR,
        });
        conn.execute(
            "INSERT INTO event (event_key, at, actor, card_id, kind, payload) \
             VALUES (?1, unixepoch(), ?2, ?3, 'card.verdict', ?4)",
            rusqlite::params![
                format!("test-verdict-code-{card}-{head}"),
                S_AUDITOR,
                number(card),
                payload.to_string()
            ],
        )
        .unwrap();
    }

    fn integ(&self) -> String {
        git(&self.project, &["rev-parse", "refs/heads/integ"])
    }

    /// Sends the card to review as `S_A` and lands a candidate as the leader; returns the answer.
    fn land(&self, card: &str, met: &[&str]) -> Value {
        result(&self.card(Some(S_A), &["review", card, "--generation", "1"]));
        let candidate = self.candidate_now();
        self.pass_of(card, &candidate);
        let mut args = vec!["land", card, "--head", candidate.as_str()];
        for m in met {
            args.extend(["--met", m]);
        }
        args.extend(["--by", "leader"]);
        result(&self.card(None, &args))
    }

    /// `card finish` as the holder, on the card's change.
    fn finish(&self, card: &str) -> Value {
        self.pass_of_code(card, &self.head);
        result(&self.card(
            Some(S_A),
            &[
                "finish",
                card,
                "--head",
                &self.head,
                "--met",
                "1=the test",
                "--by",
                "leader",
            ],
        ))
    }

    /// `card landed` then `card criterion` then `card done`, as the holder.
    fn done(&self, card: &str) -> Value {
        result(&self.card(Some(S_A), &["review", card, "--generation", "1"]));
        result(&self.card(
            Some(S_A),
            &["landed", card, "--head", &self.head, "--by", "leader"],
        ));
        result(&self.card(
            Some(S_A),
            &[
                "criterion",
                card,
                "1",
                "--generation",
                "1",
                "--met",
                "--evidence",
                "the test",
            ],
        ));
        result(&self.card(Some(S_A), &["done", card, "--generation", "1"]))
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

    fn all_events(&self) -> i64 {
        self.count("SELECT count(*) FROM event")
    }

    /// The payloads of one kind of event of a card, oldest first.
    fn payloads(&self, card: &str, kind: &str) -> Vec<Value> {
        let l = ledger::open_read_only(self.home.path(), "demo").unwrap();
        let mut st = l
            .conn()
            .prepare("SELECT payload FROM event WHERE card_id = ?1 AND kind = ?2 ORDER BY event_id")
            .unwrap();
        st.query_map(rusqlite::params![number(card), kind], |r| {
            r.get::<_, String>(0)
        })
        .unwrap()
        .map(|p| serde_json::from_str(&p.unwrap()).unwrap())
        .collect()
    }

    /// The executor, run as a library call with the default options.
    fn run_pending(&self) -> Value {
        self.run_pending_with(&Options::default())
    }

    fn run_pending_with(&self, opts: &Options<'_>) -> Value {
        self.write_profile();
        cleanup::run_pending(&self.project, self.home.path(), S_EXEC, opts)
    }

    /// The answer's entry of the request of `card`.
    fn entry<'a>(answer: &'a Value, card: &str) -> &'a Value {
        answer["requests"]
            .as_array()
            .unwrap_or_else(|| panic!("no requests list: {answer}"))
            .iter()
            .find(|r| r["card_id"] == json!(number(card)))
            .unwrap_or_else(|| panic!("no request of {card}: {answer}"))
    }

    /// The `cleanup_outbox` entry of a command answer for `card`.
    fn outbox<'a>(got: &'a Value, card: &str) -> &'a Value {
        assert_eq!(got["cleanup_outbox"]["error"], json!(null), "{got}");
        Self::entry(&got["cleanup_outbox"], card)
    }

    /// A card closed while another card still claims its worktree: the request is pending (a live
    /// claim passes by itself). Returns the closed card, the claiming card, the card's worktree
    /// and its branch.
    fn pending(&self, tag: &str) -> (String, String, PathBuf, String) {
        let wt = self.linked_worktree(&format!("wt-{tag}"));
        let branch = format!("lane/{tag}");
        self.branch(&branch);
        let claimer = self.card_in(&format!("apps/claim-{tag}/"), &wt, None, &["held"]);
        let card = self.card_in(&format!("apps/{tag}/"), &wt, Some(&branch), &["one"]);
        let got = self.land(&card, &["1=the test"]);
        assert_eq!(got["closed"], json!(true), "{got}");
        let mine = Self::outbox(&got, &card);
        assert_eq!(mine["status"], json!("pending"), "{got}");
        assert!(wt.is_dir() && self.registered(&wt), "{got}");
        assert!(self.has_branch(&branch), "{got}");
        (card, claimer, wt, branch)
    }

    /// The claiming card gives its reservation back: the claim that passed by itself is over.
    fn release(&self, card: &str) {
        result(&self.card(Some(S_A), &["release", card, "--generation", "1"]));
    }
}

// ---- AC-1: the request is written with the close, in the same transaction ---------------------

/// AC-1: `card finish` writes one request that names the card's worktree and branch.
#[test]
fn ac1_finish_writes_one_request_naming_the_worktree_and_the_branch() {
    let w = World::new();
    let wt = w.linked_worktree("wt-finish");
    w.branch("lane/finish");
    let card = w.card_in("apps/a/", &wt, Some("lane/finish"), &["one"]);
    let got = w.finish(&card);
    assert_eq!(got["closed"], json!(true), "{got}");
    assert_eq!(w.events(&card, "cleanup.requested"), 1);
    let request = &w.payloads(&card, "cleanup.requested")[0];
    assert_eq!(
        request["worktrees"],
        json!([wt.to_str().unwrap()]),
        "{request}"
    );
    assert_eq!(request["branches"], json!(["lane/finish"]), "{request}");
    assert!(got["cleanup_outbox"].is_object(), "{got}");
}

/// AC-1: `card land` writes one request, and `card done` writes one.
#[test]
fn ac1_land_and_done_each_write_one_request() {
    let w = World::new();
    let wt_a = w.linked_worktree("wt-land");
    w.branch("lane/land");
    let landed = w.card_in("apps/a/", &wt_a, Some("lane/land"), &["one"]);
    let got = w.land(&landed, &["1=the test"]);
    assert_eq!(got["closed"], json!(true), "{got}");
    assert_eq!(w.events(&landed, "cleanup.requested"), 1);
    assert!(got["cleanup_outbox"].is_object(), "{got}");

    let wt_b = w.linked_worktree("wt-done");
    w.branch("lane/done");
    let done = w.card_in("apps/b/", &wt_b, Some("lane/done"), &["one"]);
    let got = w.done(&done);
    assert_eq!(got["state"], json!("done"), "{got}");
    assert_eq!(w.events(&done, "cleanup.requested"), 1);
    assert!(got["cleanup_outbox"].is_object(), "{got}");
    assert_eq!(
        w.payloads(&done, "cleanup.requested")[0]["branches"],
        json!(["lane/done"])
    );
}

/// AC-1: a refused done, a refused finish and a landing that leaves the card open write no
/// request, and the ledger keeps every other row as it was.
#[test]
fn ac1_a_refusal_and_a_landing_that_leaves_the_card_open_write_no_request() {
    let w = World::new();
    let wt = w.linked_worktree("wt-open");
    w.branch("lane/open");
    let card = w.card_in("apps/a/", &wt, Some("lane/open"), &["one", "two"]);

    // A done in the wrong state (the card is active, not in review).
    let before = w.all_events();
    let out = w.card(Some(S_A), &["done", &card, "--generation", "1"]);
    assert_eq!(
        out.status.code(),
        Some(1),
        "{}",
        String::from_utf8_lossy(&out.stdout)
    );
    assert_eq!(w.all_events(), before, "a refused done wrote an event");
    assert_eq!(w.events(&card, "cleanup.requested"), 0);

    // A finish that names no met option for either open criterion records the landing only.
    let out = w.card(
        Some(S_A),
        &[
            "finish",
            &card,
            "--head",
            &w.head,
            "--met",
            "1=the test",
            "--by",
            "leader",
        ],
    );
    let got = result(&out);
    assert_eq!(got["closed"], json!(false), "{got}");
    assert_eq!(w.events(&card, "cleanup.requested"), 0, "{got}");
    assert!(wt.is_dir() && w.registered(&wt) && w.has_branch("lane/open"));
}

// ---- AC-2: the whole path ------------------------------------------------------------------

/// AC-2: a card whose worktree is clean and merged and whose branch is merged leaves neither
/// behind; the answer and the ledger say what was done.
#[test]
fn ac2_a_clean_merged_card_loses_its_worktree_and_its_branch() {
    let w = World::new();
    let wt = w.linked_worktree("wt-mine");
    let other = w.linked_worktree("wt-other");
    w.branch("lane/mine");
    w.branch("lane/other");
    let mine = w.card_in("apps/a/", &wt, Some("lane/mine"), &["one"]);
    let _held = w.card_in("apps/b/", &other, Some("lane/other"), &["two"]);
    let got = w.land(&mine, &["1=the test"]);
    assert_eq!(got["closed"], json!(true), "{got}");
    let entry = World::outbox(&got, &mine);
    assert_eq!(entry["status"], json!("settled"), "{got}");
    assert_eq!(entry["worktrees"][0]["status"], json!("removed"), "{got}");
    assert_eq!(entry["branches"][0]["branch"], json!("lane/mine"), "{got}");
    assert_eq!(entry["branches"][0]["status"], json!("deleted"), "{got}");

    assert!(
        !wt.exists() && !w.registered(&wt),
        "the worktree is still there"
    );
    assert!(!w.has_branch("lane/mine"), "the branch is still there");
    // Another card's worktree and branch stay.
    assert!(other.is_dir() && w.registered(&other));
    assert!(w.has_branch("lane/other"));
    // The integration branch moved with the landing and nothing else of it.
    assert_eq!(w.integ(), got["to_tip"].as_str().unwrap());

    assert_eq!(w.events(&mine, "cleanup.requested"), 1);
    assert_eq!(w.events(&mine, "cleanup.started"), 1);
    assert!(w.events(&mine, "cleanup.step") >= 2);
    let results = w.payloads(&mine, "cleanup.result");
    assert_eq!(results.len(), 1, "{results:?}");
    assert_eq!(results[0]["status"], json!("settled"), "{results:?}");
}

/// AC-2 (N7): `card finish` closes a card whose worktree is clean and merged and whose branch is
/// merged, and removes both; another card's worktree and branch stay.
#[test]
fn ac2_card_finish_removes_the_worktree_and_the_branch() {
    let w = World::new();
    let wt = w.linked_worktree("wt-fin");
    let other = w.linked_worktree("keepme-fin");
    w.branch("lane/fin");
    w.branch("lane/fin-other");
    let card = w.card_in("apps/a/", &wt, Some("lane/fin"), &["one"]);
    let _held = w.card_in("apps/b/", &other, Some("lane/fin-other"), &["two"]);
    let got = w.finish(&card);
    assert_eq!(got["closed"], json!(true), "{got}");
    let entry = World::outbox(&got, &card);
    assert_eq!(entry["status"], json!("settled"), "{got}");
    assert_eq!(entry["worktrees"][0]["status"], json!("removed"), "{got}");
    assert_eq!(entry["branches"][0]["status"], json!("deleted"), "{got}");
    assert!(
        !wt.exists() && !w.registered(&wt),
        "the worktree is still there"
    );
    assert!(!w.has_branch("lane/fin"), "the branch is still there");
    assert!(other.is_dir() && w.registered(&other));
    assert!(w.has_branch("lane/fin-other"));
    let results = w.payloads(&card, "cleanup.result");
    assert_eq!(results.len(), 1, "{results:?}");
    assert_eq!(results[0]["status"], json!("settled"), "{results:?}");
}

/// AC-2 (N7): `card done` closes the card and removes the same two things.
#[test]
fn ac2_card_done_removes_the_worktree_and_the_branch() {
    let w = World::new();
    let wt = w.linked_worktree("wt-done");
    let other = w.linked_worktree("keepme-done");
    w.branch("lane/done");
    w.branch("lane/done-other");
    let card = w.card_in("apps/a/", &wt, Some("lane/done"), &["one"]);
    let _held = w.card_in("apps/b/", &other, Some("lane/done-other"), &["two"]);
    let got = w.done(&card);
    let entry = World::outbox(&got, &card);
    assert_eq!(entry["status"], json!("settled"), "{got}");
    assert_eq!(entry["worktrees"][0]["status"], json!("removed"), "{got}");
    assert_eq!(entry["branches"][0]["status"], json!("deleted"), "{got}");
    assert!(
        !wt.exists() && !w.registered(&wt),
        "the worktree is still there"
    );
    assert!(!w.has_branch("lane/done"), "the branch is still there");
    assert!(other.is_dir() && w.registered(&other));
    assert!(w.has_branch("lane/done-other"));
}

/// AC-2 (B2): a worktree the reservation names by a relative path is read against the project, not
/// against the folder the command runs in: a run in the project and a run in another folder give
/// the same answer, and both remove the worktree and delete the branch.
#[test]
fn ac2_a_relative_worktree_path_gives_the_same_answer_in_any_folder() {
    let mut answers = Vec::new();
    for from_project in [true, false] {
        let w = World::new();
        let wt = w.project.join("wt-rel");
        git(
            &w.project,
            &[
                "worktree",
                "add",
                "-q",
                "--detach",
                wt.to_str().unwrap(),
                &w.head,
            ],
        );
        w.branch("lane/rel");
        let elsewhere = w.outside("elsewhere");
        fs::create_dir_all(&elsewhere).unwrap();
        *w.cwd.borrow_mut() = Some(if from_project {
            w.project.clone()
        } else {
            elsewhere
        });
        let card = w.card_in("apps/a/", Path::new("wt-rel"), Some("lane/rel"), &["one"]);
        let got = w.finish(&card);
        let entry = World::outbox(&got, &card);
        assert_eq!(entry["worktrees"][0]["status"], json!("removed"), "{got}");
        assert_eq!(entry["branches"][0]["status"], json!("deleted"), "{got}");
        assert!(
            !wt.exists() && !w.registered(&wt),
            "the worktree is still there: {got}"
        );
        assert!(
            !w.has_branch("lane/rel"),
            "the branch is still there: {got}"
        );
        answers.push(json!({
            "status": entry["status"], "worktrees": entry["worktrees"], "branches": entry["branches"],
        }));
    }
    assert_eq!(
        answers[0], answers[1],
        "the answer depends on the folder of the run"
    );
}

/// AC-3 (B2): a relative path that no registered worktree of the project matches is `absent`
/// whatever the folder of the run, and a folder of that name in the run's own folder is not read.
#[test]
fn ac3_a_relative_path_is_not_matched_against_the_folder_of_the_run() {
    let w = World::new();
    // A real worktree sits in the run's folder under the relative name; the card's relative name
    // is the same, but the project has no such folder, so it is the project's answer that counts.
    let wt = w.linked_worktree("wt-cwd");
    w.branch("lane/cwd");
    *w.cwd.borrow_mut() = Some(w.outside(""));
    let card = w.card_in("apps/a/", Path::new("wt-cwd"), Some("lane/cwd"), &["one"]);
    let got = w.finish(&card);
    let entry = World::outbox(&got, &card);
    assert!(
        wt.is_dir() && w.registered(&wt),
        "a worktree found through the folder of the run was removed: {got}"
    );
    assert_eq!(entry["worktrees"][0]["status"], json!("absent"), "{got}");
}

// ---- AC-3: what stays ----------------------------------------------------------------------

/// AC-3: a dirty worktree stays with its branch, and the request ends as settled with kept.
#[test]
fn ac3_a_dirty_worktree_and_its_branch_stay() {
    let w = World::new();
    let wt = w.linked_worktree("wt-dirty");
    w.branch("lane/dirty");
    fs::write(wt.join("untracked.txt"), "work in progress\n").unwrap();
    let card = w.card_in("apps/a/", &wt, Some("lane/dirty"), &["one"]);
    let got = w.land(&card, &["1=the test"]);
    let entry = World::outbox(&got, &card);
    assert_eq!(entry["status"], json!("settled_with_kept"), "{got}");
    assert_eq!(entry["worktrees"][0]["status"], json!("kept"), "{got}");
    assert_eq!(entry["branches"][0]["status"], json!("kept"), "{got}");
    assert!(
        entry["branches"][0]["reasons"]
            .to_string()
            .contains("worktree_kept"),
        "{got}"
    );
    assert!(wt.join("untracked.txt").is_file(), "the dirty file is gone");
    assert!(w.registered(&wt) && w.has_branch("lane/dirty"));
}

/// AC-3 (dangerous path): a worktree path that is a link to a folder holding someone's files is
/// never followed: the folder behind the link and its files stay, the worktree is kept with the
/// reason `link_path`, and so is the card's branch.
#[test]
#[cfg(unix)]
fn ac3_a_worktree_path_that_is_a_link_is_not_followed() {
    let w = World::new();
    let precious = w.outside("precious");
    fs::create_dir_all(&precious).unwrap();
    fs::write(precious.join("keep.txt"), "not a worktree\n").unwrap();
    let link = w.outside("wt-link");
    std::os::unix::fs::symlink(&precious, &link).unwrap();
    w.branch("lane/link");
    let card = w.card_in("apps/a/", &link, Some("lane/link"), &["one"]);
    let got = w.land(&card, &["1=the test"]);
    assert_eq!(got["closed"], json!(true), "{got}");
    assert!(
        precious.join("keep.txt").is_file(),
        "the folder behind the link lost a file: {got}"
    );
    assert!(link.exists(), "the link or its target was removed: {got}");
    let entry = World::outbox(&got, &card);
    assert_eq!(entry["worktrees"][0]["status"], json!("kept"), "{got}");
    assert!(
        entry["worktrees"][0]["reasons"]
            .to_string()
            .contains("link_path"),
        "{got}"
    );
    assert_eq!(entry["branches"][0]["status"], json!("kept"), "{got}");
    assert!(w.has_branch("lane/link"), "{got}");
}

/// AC-3 (B3, dangerous path): the reservation names a link whose target is a clean, merged, registered
/// worktree of the project. Nothing is touched: the link and the worktree behind it stay registered
/// with their files, the answer says `link_path`, and the branch stays.
#[test]
#[cfg(unix)]
fn ac3_a_link_to_a_registered_worktree_leaves_that_worktree_standing() {
    let w = World::new();
    let real = w.linked_worktree("wt-real");
    let link = w.outside("wt-alias");
    std::os::unix::fs::symlink(&real, &link).unwrap();
    w.branch("lane/alias");
    let card = w.card_in("apps/a/", &link, Some("lane/alias"), &["one"]);
    let got = w.finish(&card);
    assert_eq!(got["closed"], json!(true), "{got}");
    let entry = World::outbox(&got, &card);
    assert_eq!(entry["worktrees"][0]["status"], json!("kept"), "{got}");
    assert!(
        entry["worktrees"][0]["reasons"]
            .to_string()
            .contains("link_path"),
        "{got}"
    );
    assert!(
        real.is_dir() && w.registered(&real),
        "the worktree behind the link was removed: {got}"
    );
    assert!(link.exists(), "{got}");
    assert!(w.has_branch("lane/alias"), "{got}");
    assert_eq!(entry["branches"][0]["status"], json!("kept"), "{got}");
    // The step row says kept too, never removed.
    let steps = w.payloads(&card, "cleanup.step");
    assert!(
        steps.iter().all(|s| s["status"] != json!("removed")),
        "{steps:?}"
    );
}

/// AC-2 (B3): a link in a parent folder is allowed (macOS keeps `/tmp` and `/var` behind links):
/// only the last element of the path is refused. The worktree is removed and the branch deleted.
#[test]
#[cfg(unix)]
fn ac2_a_link_in_a_parent_folder_does_not_stop_the_cleanup() {
    let w = World::new();
    let real_parent = w.outside("real-parent");
    fs::create_dir_all(&real_parent).unwrap();
    let wt = real_parent.join("wt-inner");
    git(
        &w.project,
        &[
            "worktree",
            "add",
            "-q",
            "--detach",
            wt.to_str().unwrap(),
            &w.head,
        ],
    );
    let parent_link = w.outside("parent-link");
    std::os::unix::fs::symlink(&real_parent, &parent_link).unwrap();
    w.branch("lane/parent");
    let card = w.card_in(
        "apps/a/",
        &parent_link.join("wt-inner"),
        Some("lane/parent"),
        &["one"],
    );
    let got = w.finish(&card);
    let entry = World::outbox(&got, &card);
    assert_eq!(entry["worktrees"][0]["status"], json!("removed"), "{got}");
    assert_eq!(entry["branches"][0]["status"], json!("deleted"), "{got}");
    assert!(!wt.exists() && !w.registered(&wt), "{got}");
    assert!(!w.has_branch("lane/parent"), "{got}");
}

/// AC-3 (B1): the plan trims a worktree that holds a file of its own next to a build folder: only
/// the build output goes, the worktree stays registered, so the answer and the step row say `kept`
/// with `trimmed_only` (never `removed`), and the card's branch is not deleted.
#[test]
fn ac3_a_trimmed_worktree_is_kept_and_so_is_its_branch() {
    let w = World::new();
    let wt = w.linked_worktree("wt-trim");
    w.branch("lane/trim");
    fs::write(wt.join("notes.txt"), "work in progress\n").unwrap();
    let build = wt.join("target/debug");
    fs::create_dir_all(&build).unwrap();
    fs::write(
        build.join("CACHEDIR.TAG"),
        "Signature: 8a477f597d28d172789f06886806bc55\n",
    )
    .unwrap();
    fs::write(build.join("artifact.o"), "x").unwrap();
    let card = w.card_in("apps/a/", &wt, Some("lane/trim"), &["one"]);
    let got = w.finish(&card);
    assert_eq!(got["closed"], json!(true), "{got}");
    let entry = World::outbox(&got, &card);
    assert_eq!(entry["status"], json!("settled_with_kept"), "{got}");
    assert_eq!(entry["worktrees"][0]["status"], json!("kept"), "{got}");
    assert!(
        entry["worktrees"][0]["reasons"]
            .to_string()
            .contains("trimmed_only"),
        "{got}"
    );
    assert_eq!(entry["branches"][0]["status"], json!("kept"), "{got}");
    assert!(
        entry["branches"][0]["reasons"]
            .to_string()
            .contains("worktree_kept"),
        "{got}"
    );
    assert!(wt.is_dir() && w.registered(&wt), "the worktree went: {got}");
    assert!(wt.join("notes.txt").is_file(), "{got}");
    assert!(w.has_branch("lane/trim"), "the branch was deleted: {got}");
    let steps = w.payloads(&card, "cleanup.step");
    assert!(
        steps.iter().all(|s| s["status"] != json!("removed")),
        "a step row says removed for a worktree that stands: {steps:?}"
    );
}

/// AC-3: a branch the integration tip does not contain (a squash leaves no ancestry) stays, with
/// the reason, while the clean merged worktree goes.
#[test]
fn ac3_a_branch_the_tip_does_not_contain_stays() {
    let w = World::new();
    let wt = w.linked_worktree("wt-squash");
    let extra = commit_with(&w.project, &w.head, &[&w.head], "a commit the tip lacks");
    git(&w.project, &["branch", "lane/squash", &extra]);
    let card = w.card_in("apps/a/", &wt, Some("lane/squash"), &["one"]);
    let got = w.land(&card, &["1=the test"]);
    let entry = World::outbox(&got, &card);
    assert_eq!(entry["status"], json!("settled_with_kept"), "{got}");
    assert_eq!(entry["worktrees"][0]["status"], json!("removed"), "{got}");
    assert_eq!(entry["branches"][0]["status"], json!("kept"), "{got}");
    assert!(
        entry["branches"][0]["reasons"]
            .to_string()
            .contains("not_merged"),
        "{got}"
    );
    assert!(!wt.exists());
    assert_eq!(git(&w.project, &["rev-parse", "lane/squash"]), extra);
}

/// AC-3: a preserved branch name stays, and so does a branch another worktree has checked out.
#[test]
fn ac3_a_preserved_branch_and_a_branch_checked_out_elsewhere_stay() {
    let w = World::new();
    // Preserved: the `archive/` prefix.
    let wt_p = w.linked_worktree("wt-pre");
    w.branch("archive/old");
    let pre = w.card_in("apps/a/", &wt_p, Some("archive/old"), &["one"]);
    let got = w.land(&pre, &["1=the test"]);
    let entry = World::outbox(&got, &pre);
    assert_eq!(entry["worktrees"][0]["status"], json!("removed"), "{got}");
    assert_eq!(entry["branches"][0]["status"], json!("kept"), "{got}");
    assert!(
        entry["branches"][0]["reasons"]
            .to_string()
            .contains("preserved"),
        "{got}"
    );
    assert!(w.has_branch("archive/old"));

    // Checked out in a worktree the card never named.
    let wt_c = w.linked_worktree("wt-card");
    w.branch("lane/shared");
    let elsewhere = w.outside("wt-elsewhere");
    git(
        &w.project,
        &[
            "worktree",
            "add",
            "-q",
            elsewhere.to_str().unwrap(),
            "lane/shared",
        ],
    );
    let shared = w.card_in("apps/b/", &wt_c, Some("lane/shared"), &["two"]);
    let got = w.land(&shared, &["1=the test"]);
    let entry = World::outbox(&got, &shared);
    assert_eq!(entry["branches"][0]["status"], json!("kept"), "{got}");
    assert!(
        entry["branches"][0]["reasons"]
            .to_string()
            .contains("checked_out"),
        "{got}"
    );
    assert!(w.has_branch("lane/shared"));
    assert!(elsewhere.is_dir() && w.registered(&elsewhere));
}

/// AC-3: a worktree under `.claude/worktrees/` is preserved by the cleanup plan, so the worktree
/// and its branch stay and the answer names the reason; no remote ref exists or changes.
#[test]
fn ac3_a_worktree_under_the_preserved_place_stays_with_its_branch() {
    let w = World::new();
    let under = w.project.join(".claude/worktrees/lane-x");
    fs::create_dir_all(under.parent().unwrap()).unwrap();
    git(
        &w.project,
        &[
            "worktree",
            "add",
            "-q",
            "--detach",
            under.to_str().unwrap(),
            &w.head,
        ],
    );
    w.branch("lane/x");
    let card = w.card_in("apps/a/", &under, Some("lane/x"), &["one"]);
    let got = w.land(&card, &["1=the test"]);
    let entry = World::outbox(&got, &card);
    assert_eq!(entry["status"], json!("settled_with_kept"), "{got}");
    assert_eq!(entry["worktrees"][0]["status"], json!("kept"), "{got}");
    assert!(
        entry["worktrees"][0]["reasons"]
            .to_string()
            .contains("preserved_location"),
        "{got}"
    );
    assert!(under.is_dir() && w.registered(&under) && w.has_branch("lane/x"));
    assert!(
        git(&w.project, &["for-each-ref", "refs/remotes"]).is_empty(),
        "a remote ref appeared"
    );
}

// ---- AC-4: a second run does nothing --------------------------------------------------------

/// AC-4: after a settled request another executor run writes no event and moves no ref.
#[test]
fn ac4_a_second_run_after_a_settled_request_writes_nothing() {
    let w = World::new();
    let wt = w.linked_worktree("wt-once");
    w.branch("lane/once");
    let card = w.card_in("apps/a/", &wt, Some("lane/once"), &["one"]);
    let got = w.land(&card, &["1=the test"]);
    assert_eq!(World::outbox(&got, &card)["status"], json!("settled"));

    let events = w.all_events();
    let refs = git(&w.project, &["for-each-ref"]);
    let listed = git(&w.project, &["worktree", "list", "--porcelain"]);
    let again = w.run_pending();
    assert_eq!(again["error"], json!(null), "{again}");
    assert_eq!(again["requests"], json!([]), "{again}");
    assert_eq!(
        w.all_events(),
        events,
        "a settled request was acted on again"
    );
    assert_eq!(git(&w.project, &["for-each-ref"]), refs);
    assert_eq!(
        git(&w.project, &["worktree", "list", "--porcelain"]),
        listed
    );
}

// ---- AC-5: a duplicate, a restart ----------------------------------------------------------

/// AC-5: two executors that start together on one request leave one claim and one result.
#[test]
fn ac5_two_executors_on_one_request_leave_one_claim_and_one_result() {
    let w = World::new();
    let (card, claimer, wt, branch) = w.pending("dup");
    // The close's own run took attempt 1 and gave the claim back.
    let before = w.events(&card, "cleanup.started");
    assert_eq!(before, 1);
    w.release(&claimer);
    w.write_profile();
    let (project, home) = (&w.project, w.home.path());
    let answers: Vec<Value> = std::thread::scope(|s| {
        let handles: Vec<_> = (0..2)
            .map(|_| s.spawn(|| cleanup::run_pending(project, home, S_EXEC, &Options::default())))
            .collect();
        handles.into_iter().map(|h| h.join().unwrap()).collect()
    });
    for a in &answers {
        assert_eq!(a["error"], json!(null), "{a}");
    }
    assert_eq!(
        w.events(&card, "cleanup.started"),
        before + 1,
        "{answers:?}"
    );
    assert_eq!(w.events(&card, "cleanup.result"), 1, "{answers:?}");
    assert!(!wt.exists() && !w.registered(&wt));
    assert!(!w.has_branch(&branch));
}

/// AC-5: an executor that stopped after the worktree step leaves the branch and the request; the
/// next run, once the lease ended, deletes the branch and writes one result.
#[test]
fn ac5_a_run_that_stopped_after_the_worktree_step_is_finished_by_the_next() {
    let w = World::new();
    let (card, claimer, wt, branch) = w.pending("stop");
    w.release(&claimer);

    let stopped = w.run_pending_with(&Options {
        // Long enough that a loaded machine does not end it before the busy check below.
        lease_secs: 12,
        stop_after_worktrees: true,
        ..Options::default()
    });
    assert_eq!(
        World::entry(&stopped, &card)["status"],
        json!("interrupted"),
        "{stopped}"
    );
    assert!(!wt.exists(), "the worktree step did not run");
    assert!(w.has_branch(&branch), "the branch went before the stop");
    assert_eq!(w.events(&card, "cleanup.result"), 0);

    // The lease is still live: another run leaves the request to its claim.
    let busy = w.run_pending();
    assert_eq!(
        World::entry(&busy, &card)["status"],
        json!("busy"),
        "{busy}"
    );
    assert!(w.has_branch(&branch));

    // Wait for the lease to end (the database clock counts whole seconds), then run again.
    let mut done = w.run_pending();
    for _ in 0..60 {
        if World::entry(&done, &card)["status"] != json!("busy") {
            break;
        }
        std::thread::sleep(std::time::Duration::from_millis(500));
        done = w.run_pending();
    }
    assert_eq!(
        World::entry(&done, &card)["status"],
        json!("settled"),
        "{done}"
    );
    assert!(!w.has_branch(&branch));
    // Attempt 1 (the close), attempt 2 (stopped, lease ended), attempt 3 (this run).
    assert_eq!(w.events(&card, "cleanup.started"), 3);
    assert_eq!(w.events(&card, "cleanup.result"), 1);
}

// ---- AC-6: pending requests -----------------------------------------------------------------

/// AC-6: a worktree another card still claims keeps the request pending; the next close of any
/// card runs it once the claim is over.
#[test]
fn ac6_a_claimed_worktree_keeps_the_request_pending_until_a_later_close() {
    let w = World::new();
    let (card, claimer, wt, branch) = w.pending("claim");
    assert_eq!(w.events(&card, "cleanup.result"), 0);

    // Still claimed: a run changes nothing and the request stays pending.
    let again = w.run_pending();
    assert_eq!(
        World::entry(&again, &card)["status"],
        json!("pending"),
        "{again}"
    );
    assert!(wt.is_dir() && w.has_branch(&branch));

    w.release(&claimer);
    // Another card closes: its command runs every pending request of the project.
    let wt2 = w.linked_worktree("wt-next");
    w.branch("lane/next");
    let next = w.card_in("apps/next/", &wt2, Some("lane/next"), &["one"]);
    let got = w.land(&next, &["1=the test"]);
    assert_eq!(
        World::outbox(&got, &next)["status"],
        json!("settled"),
        "{got}"
    );
    let older = World::entry(&got["cleanup_outbox"], &card);
    assert_eq!(older["status"], json!("settled"), "{got}");
    assert!(!wt.exists() && !w.has_branch(&branch));
    assert_eq!(w.events(&card, "cleanup.result"), 1);
}

/// AC-6: a git failure on the branch delete (a lock file on the ref) leaves the request pending
/// until the cause is gone.
#[test]
fn ac6_a_git_failure_on_the_branch_delete_leaves_the_request_pending() {
    let w = World::new();
    let wt = w.linked_worktree("wt-lock");
    w.branch("lane/lock");
    let lock = w.project.join(".git/refs/heads/lane/lock.lock");
    fs::write(&lock, "").unwrap();
    let card = w.card_in("apps/a/", &wt, Some("lane/lock"), &["one"]);
    let got = w.land(&card, &["1=the test"]);
    assert_eq!(got["closed"], json!(true), "{got}");
    let entry = World::outbox(&got, &card);
    assert_eq!(entry["status"], json!("pending"), "{got}");
    assert_eq!(entry["worktrees"][0]["status"], json!("removed"), "{got}");
    assert!(
        matches!(
            entry["branches"][0]["status"].as_str(),
            Some("failed" | "refused")
        ),
        "{got}"
    );
    assert!(w.has_branch("lane/lock"));
    assert_eq!(w.events(&card, "cleanup.result"), 0);

    // The failed attempt gave its claim back, so the next run needs no wait.
    fs::remove_file(&lock).unwrap();
    let done = w.run_pending();
    assert_eq!(
        World::entry(&done, &card)["status"],
        json!("settled"),
        "{done}"
    );
    assert!(!w.has_branch("lane/lock"));
    assert_eq!(w.events(&card, "cleanup.result"), 1);
}

// ---- AC-7: a race, a reopened card ----------------------------------------------------------

/// AC-7: a branch another worktree checks out after the worktree step is kept and reported, and
/// that worktree is not touched.
#[test]
fn ac7_a_branch_checked_out_after_the_worktree_step_is_kept() {
    let w = World::new();
    let (card, claimer, wt, branch) = w.pending("race");
    w.release(&claimer);
    let taker = w.outside("wt-taker");
    let take = || {
        git(
            &w.project,
            &["worktree", "add", "-q", taker.to_str().unwrap(), &branch],
        );
    };
    let got = w.run_pending_with(&Options {
        before_branch: Some(&take),
        ..Options::default()
    });
    let entry = World::entry(&got, &card);
    assert_eq!(entry["status"], json!("settled_with_kept"), "{got}");
    assert_eq!(entry["branches"][0]["status"], json!("kept"), "{got}");
    assert!(
        entry["branches"][0]["reasons"]
            .to_string()
            .contains("checked_out"),
        "{got}"
    );
    assert!(!wt.exists(), "the card's own worktree stays removed");
    assert!(w.has_branch(&branch));
    assert!(taker.is_dir() && w.registered(&taker));
}

/// AC-7: a card that is no longer done in the ledger is skipped: nothing is removed.
#[test]
fn ac7_a_card_reopened_after_its_request_is_skipped() {
    let w = World::new();
    let (card, _claimer, wt, branch) = w.pending("reopen");
    let file = ledger::ledger_path(w.home.path(), "demo").unwrap();
    let conn = rusqlite::Connection::open(&file).unwrap();
    conn.execute(
        "UPDATE card SET state = 'ready', closed_at = NULL WHERE card_id = ?1",
        [number(&card)],
    )
    .unwrap();
    drop(conn);
    let got = w.run_pending();
    let entry = World::entry(&got, &card);
    assert_eq!(entry["status"], json!("skipped"), "{got}");
    assert!(wt.is_dir() && w.registered(&wt) && w.has_branch(&branch));
    assert_eq!(w.events(&card, "cleanup.result"), 1);
}

// ---- AC-8: an executor problem never undoes the close ---------------------------------------

/// AC-8: with no integration branch to measure against, the executor reports an error, removes
/// nothing, leaves the request pending, and the card stays closed.
#[test]
fn ac8_a_problem_of_the_executor_is_reported_and_the_card_stays_closed() {
    let w = World::new();
    let (card, claimer, wt, branch) = w.pending("noref");
    w.release(&claimer);
    // No work_flow in the profile and no recorded integration branch.
    *w.profile.borrow_mut() = PLAIN_PROFILE.to_vec();
    let _ = fs::remove_file(w.project.join(".git/moai/integration-branch"));
    let got = w.run_pending();
    assert!(got["error"].is_string(), "{got}");
    assert!(wt.is_dir() && w.registered(&wt) && w.has_branch(&branch));
    assert_eq!(w.events(&card, "cleanup.result"), 0);
    let state: String = {
        let l = ledger::open_read_only(w.home.path(), "demo").unwrap();
        l.conn()
            .query_row(
                "SELECT state FROM card WHERE card_id = ?1",
                [number(&card)],
                |r| r.get(0),
            )
            .unwrap()
    };
    assert_eq!(state, "done");
}

// ---- SPEC-20261123-cleanup-outbox-followups ---------------------------------------------------

fn a_step(target: &str) -> Step {
    Step {
        kind: "worktree",
        target: target.to_string(),
        status: "removed".to_string(),
        reasons: Vec::new(),
        detail: None,
    }
}

impl World {
    /// The name of the one pending request of the project.
    fn the_request(&self) -> i64 {
        let l = ledger::open_read_only(self.home.path(), "demo").unwrap();
        let pending = cleanup_outbox::pending(l.conn()).unwrap();
        assert_eq!(pending.len(), 1, "{pending:?}");
        pending[0].request
    }

    /// The `cleanup.step` payloads of a card whose `kind` field is `kind`.
    fn steps_of_kind(&self, card: &str, kind: &str) -> Vec<Value> {
        self.payloads(card, "cleanup.step")
            .into_iter()
            .filter(|p| p["kind"] == json!(kind))
            .collect()
    }

    /// What a card's close wrote under the cleanup reference-transaction queue.
    fn queue_file(&self) -> PathBuf {
        self.project
            .join(".git")
            .join(moai_cli::worktree_cleanup::QUEUE_FILE)
    }
}

/// FU-1 (REQ-1, REQ-2): an attempt whose lease ended, or that a newer attempt replaced, writes no
/// step and no result; the newest live attempt finishes as before.
#[test]
fn fu1_an_attempt_that_is_not_live_writes_nothing_and_the_newest_live_one_finishes() {
    let w = World::new();
    let (card, claimer, _wt, _branch) = w.pending("fu1-ledger");
    w.release(&claimer);
    let (request, id) = (w.the_request(), number(&card));
    let mut write = ledger::open_write(w.home.path(), "demo").unwrap();
    let end = Outcome::Result("settled".to_string());

    // A lease of zero seconds is over the moment it starts.
    let Claim::Taken { attempt: dead, .. } =
        cleanup_outbox::claim(&mut write, request, id, S_EXEC, 0).unwrap()
    else {
        panic!("the request was not claimable")
    };
    let refused = cleanup_outbox::finish_attempt(
        &mut write,
        request,
        id,
        dead,
        S_EXEC,
        &[a_step("dead-target")],
        &end,
    )
    .unwrap();
    assert_eq!(refused, Finish::LeaseEnded);
    assert_eq!(w.events(&card, "cleanup.result"), 0);
    assert!(w
        .steps_of_kind(&card, "worktree")
        .iter()
        .all(|s| s["target"] != json!("dead-target")));

    // A newer attempt takes over; the older one stays refused, the newer one finishes.
    let Claim::Taken { attempt: live, .. } =
        cleanup_outbox::claim(&mut write, request, id, S_EXEC, 300).unwrap()
    else {
        panic!("the dead attempt's request was not claimable")
    };
    assert_eq!(live, dead + 1);
    let late = cleanup_outbox::finish_attempt(
        &mut write,
        request,
        id,
        dead,
        S_EXEC,
        &[a_step("late-target")],
        &end,
    )
    .unwrap();
    assert_eq!(late, Finish::LeaseEnded);
    assert_eq!(w.events(&card, "cleanup.result"), 0);

    // The live attempt's own number and holder are both needed: an attempt nobody started and
    // another holder's call are refused.
    let unknown = cleanup_outbox::finish_attempt(
        &mut write,
        request,
        id,
        live + 5,
        S_EXEC,
        &[a_step("unknown-target")],
        &end,
    )
    .unwrap();
    assert_eq!(unknown, Finish::LeaseEnded);
    let stranger = cleanup_outbox::finish_attempt(
        &mut write,
        request,
        id,
        live,
        S_A,
        &[a_step("stranger-target")],
        &end,
    )
    .unwrap();
    assert_eq!(stranger, Finish::LeaseEnded);

    // An attempt that gave its claim back cannot finish afterwards, even inside its lease.
    let gave_back = cleanup_outbox::finish_attempt(
        &mut write,
        request,
        id,
        live,
        S_EXEC,
        &[],
        &Outcome::Release,
    )
    .unwrap();
    assert_eq!(gave_back, Finish::Written);
    let after = cleanup_outbox::finish_attempt(
        &mut write,
        request,
        id,
        live,
        S_EXEC,
        &[a_step("after-release-target")],
        &end,
    )
    .unwrap();
    assert_eq!(after, Finish::LeaseEnded);
    assert_eq!(w.events(&card, "cleanup.result"), 0);

    let Claim::Taken { attempt: live, .. } =
        cleanup_outbox::claim(&mut write, request, id, S_EXEC, 300).unwrap()
    else {
        panic!("the released attempt's request was not claimable")
    };
    let written = cleanup_outbox::finish_attempt(
        &mut write,
        request,
        id,
        live,
        S_EXEC,
        &[a_step("live-target")],
        &end,
    )
    .unwrap();
    assert_eq!(written, Finish::Written);
    assert_eq!(w.events(&card, "cleanup.result"), 1);
    let targets: Vec<Value> = w
        .steps_of_kind(&card, "worktree")
        .into_iter()
        .map(|s| s["target"].clone())
        .collect();
    assert!(targets.contains(&json!("live-target")), "{targets:?}");
    assert!(!targets.contains(&json!("late-target")), "{targets:?}");
    assert!(!targets.contains(&json!("dead-target")), "{targets:?}");
    for refused in ["unknown-target", "stranger-target", "after-release-target"] {
        assert!(!targets.contains(&json!(refused)), "{refused}: {targets:?}");
    }

    // A finished request is over for everyone, a dead attempt included.
    let over =
        cleanup_outbox::finish_attempt(&mut write, request, id, dead, S_EXEC, &[], &end).unwrap();
    assert_eq!(over, Finish::Over);
}

/// FU-1 (REQ-3): an executor whose lease is over when it ends answers `lease_lost` with an error
/// that names the attempt, writes no step and no result, and the next run settles the request.
#[test]
fn fu1_an_executor_whose_lease_ended_reports_lease_lost_and_the_next_run_settles() {
    let w = World::new();
    let (card, claimer, wt, branch) = w.pending("fu1-lost");
    w.release(&claimer);
    let got = w.run_pending_with(&Options {
        lease_secs: 0,
        ..Options::default()
    });
    let entry = World::entry(&got, &card);
    assert_eq!(entry["status"], json!("lease_lost"), "{got}");
    let why = entry["error"].as_str().unwrap_or_default();
    assert!(why.contains("attempt 2"), "{got}");
    assert_eq!(got["error"], entry["error"], "{got}");
    assert_eq!(w.events(&card, "cleanup.result"), 0);
    // Attempt 1 is the close's own run, which gave its claim back and left its steps.
    let by_refused: Vec<Value> = ["worktree", "branch"]
        .iter()
        .flat_map(|kind| w.steps_of_kind(&card, kind))
        .filter(|s| s["attempt"] == json!(2))
        .collect();
    assert!(
        by_refused.is_empty(),
        "a refused attempt wrote steps: {by_refused:?}"
    );
    // The run did remove what it removed; the ledger only lacks its rows.
    assert!(!wt.exists() && !w.has_branch(&branch));

    let next = w.run_pending();
    assert_eq!(
        World::entry(&next, &card)["status"],
        json!("settled"),
        "{next}"
    );
    assert_eq!(w.events(&card, "cleanup.result"), 1);
}

/// FU-1 (REQ-3, audit N-1): the `lease_lost` answer of a run that did remove its worktree and
/// delete its branch says so, so `cleanup::wrote` is true for it, while the ledger still gets no
/// line from that run.
#[test]
fn fu1_a_lease_lost_answer_names_what_the_run_removed_and_the_ledger_gets_no_line() {
    let w = World::new();
    let (card, claimer, wt, branch) = w.pending("fu1-carry");
    w.release(&claimer);
    let steps_before = w.events(&card, "cleanup.step");
    let events_before = w.all_events();
    let got = w.run_pending_with(&Options {
        lease_secs: 0,
        ..Options::default()
    });
    let entry = World::entry(&got, &card);
    assert_eq!(entry["status"], json!("lease_lost"), "{got}");
    assert!(!wt.exists() && !w.has_branch(&branch), "{got}");

    // The answer names the worktree and the branch this run really removed.
    let worktrees = entry["worktrees"].as_array().unwrap();
    assert_eq!(worktrees.len(), 1, "{got}");
    assert_eq!(worktrees[0]["status"], json!("removed"), "{got}");
    let branches = entry["branches"].as_array().unwrap();
    assert_eq!(branches.len(), 1, "{got}");
    assert_eq!(branches[0]["branch"], json!(branch), "{got}");
    assert_eq!(branches[0]["status"], json!("deleted"), "{got}");
    assert!(cleanup::wrote(&got), "{got}");
    assert_eq!(entry["attempt"], json!(2), "{got}");

    // The ledger gets nothing from this run: no step, no result, no new event but the claim.
    assert_eq!(w.events(&card, "cleanup.step"), steps_before);
    assert_eq!(w.events(&card, "cleanup.result"), 0);
    assert_eq!(w.all_events(), events_before + 1, "only the claim is new");

    // The next run finds both gone and settles the request.
    let next = w.run_pending();
    let again = World::entry(&next, &card);
    assert_eq!(again["status"], json!("settled"), "{next}");
    assert_eq!(again["worktrees"][0]["status"], json!("absent"), "{next}");
    assert_eq!(again["branches"][0]["status"], json!("absent"), "{next}");
    assert!(!cleanup::wrote(&next), "{next}");
}

/// FU-1 (REQ-3, audit N-1): a `lease_lost` answer of a run that removed nothing carries no
/// removal, so `wrote` stays false for it.
#[test]
fn fu1_a_lease_lost_answer_of_a_run_that_removed_nothing_says_so() {
    let w = World::new();
    let (card, _claimer, wt, branch) = w.pending("fu1-kept");
    // The claimer keeps its reservation: the worktree is held, nothing can be removed, and a lease
    // of zero seconds ends before the run does.
    let got = w.run_pending_with(&Options {
        lease_secs: 0,
        ..Options::default()
    });
    let entry = World::entry(&got, &card);
    assert_eq!(entry["status"], json!("lease_lost"), "{got}");
    assert!(wt.is_dir() && w.has_branch(&branch), "{got}");
    assert!(!cleanup::wrote(&got), "{got}");
    let worktrees = entry["worktrees"].as_array().unwrap();
    assert_eq!(worktrees.len(), 1, "{got}");
    assert_ne!(worktrees[0]["status"], json!("removed"), "{got}");
}

/// FU-2 (REQ-4): after a run that died without a trace, the next attempt leaves one `abandoned`
/// line for it, and only for it.
#[test]
fn fu2_the_next_attempt_leaves_one_abandoned_line_for_a_run_that_died() {
    let w = World::new();
    let (card, claimer, _wt, _branch) = w.pending("fu2");
    w.release(&claimer);
    // Attempt 1 (the close's own run) gave its claim back. Attempt 2 stops as a killed run would.
    let died = w.run_pending_with(&Options {
        lease_secs: 0,
        stop_after_worktrees: true,
        ..Options::default()
    });
    assert_eq!(
        World::entry(&died, &card)["status"],
        json!("interrupted"),
        "{died}"
    );
    assert!(w.steps_of_kind(&card, "attempt").is_empty());

    let done = w.run_pending();
    let entry = World::entry(&done, &card);
    assert_eq!(entry["status"], json!("settled"), "{done}");
    assert_eq!(entry["abandoned_attempt"], json!(2), "{done}");
    let lines = w.steps_of_kind(&card, "attempt");
    assert_eq!(lines.len(), 1, "{lines:?}");
    assert_eq!(lines[0]["attempt"], json!(2));
    assert_eq!(lines[0]["status"], json!("abandoned"));
    assert_eq!(lines[0]["reasons"], json!(["lease_ended"]));
    assert_eq!(w.events(&card, "cleanup.result"), 1);

    // A settled request is not read again: no second line.
    let again = w.run_pending();
    assert_eq!(again["requests"], json!([]), "{again}");
    assert_eq!(w.steps_of_kind(&card, "attempt").len(), 1);
}

/// FU-2 (REQ-4, audit N-3): the ledger cannot tell a run that died from a run that finished after
/// its lease ended (it wrote nothing either way), so the `abandoned` line says both.
#[test]
fn fu2_the_abandoned_line_of_a_run_whose_lease_ended_does_not_claim_it_died() {
    let w = World::new();
    let (card, claimer, _wt, _branch) = w.pending("fu2-lost");
    w.release(&claimer);
    let lost = w.run_pending_with(&Options {
        lease_secs: 0,
        ..Options::default()
    });
    assert_eq!(
        World::entry(&lost, &card)["status"],
        json!("lease_lost"),
        "{lost}"
    );
    let done = w.run_pending();
    assert_eq!(World::entry(&done, &card)["abandoned_attempt"], json!(2));
    let lines = w.steps_of_kind(&card, "attempt");
    assert_eq!(lines.len(), 1, "{lines:?}");
    assert_eq!(lines[0]["status"], json!("abandoned"));
    let detail = lines[0]["detail"].as_str().unwrap_or_default();
    assert!(
        detail.contains("no end of it was recorded")
            && detail.contains("died, or its lease ended before it finished"),
        "{detail}"
    );
}

/// FU-3 (REQ-5): the ref queue a run drains leaves one `queue` step per ref; no queue, no step.
#[test]
fn fu3_a_drained_ref_queue_leaves_one_step_per_ref() {
    let w = World::new();
    w.branch("lane/fu3-keep");
    let (card, claimer, _wt, _branch) = w.pending("fu3");
    w.release(&claimer);
    // What the reference-transaction hook writes: one full ref name per deleted branch (a hint
    // only; the cleanup reads each ref again).
    let queue = w.queue_file();
    fs::create_dir_all(queue.parent().unwrap()).unwrap();
    fs::write(
        &queue,
        "refs/heads/lane/fu3-was\nrefs/heads/lane/fu3-keep\n",
    )
    .unwrap();

    let got = w.run_pending();
    assert_eq!(
        World::entry(&got, &card)["status"],
        json!("settled"),
        "{got}"
    );
    let lines = w.steps_of_kind(&card, "queue");
    let seen: Vec<(String, String)> = lines
        .iter()
        .map(|s| {
            (
                s["target"].as_str().unwrap().to_string(),
                s["status"].as_str().unwrap().to_string(),
            )
        })
        .collect();
    assert_eq!(
        seen,
        vec![
            (
                "refs/heads/lane/fu3-keep".to_string(),
                "still_there".to_string()
            ),
            ("refs/heads/lane/fu3-was".to_string(), "gone".to_string()),
        ],
        "{lines:?}"
    );

    // The next card's run finds no queue and writes no queue step.
    let wt2 = w.linked_worktree("wt-fu3-next");
    w.branch("lane/fu3-next");
    let next = w.card_in("apps/fu3-next/", &wt2, Some("lane/fu3-next"), &["one"]);
    let landed = w.land(&next, &["1=the test"]);
    assert_eq!(
        World::outbox(&landed, &next)["status"],
        json!("settled"),
        "{landed}"
    );
    assert!(w.steps_of_kind(&next, "queue").is_empty());
}

#[cfg(unix)]
impl World {
    /// `moai card <args>` started in a process group of its own and not waited for, so a test can
    /// kill the whole group the way `kill -9` on a stuck run would.
    fn spawn_card(&self, session: &str, args: &[&str]) -> std::process::Child {
        self.spawn_card_with(session, args, Stdio::null())
    }

    /// The same, with the answer read from a pipe (`wait_with_output`).
    fn spawn_card_piped(&self, session: &str, args: &[&str]) -> std::process::Child {
        self.spawn_card_with(session, args, Stdio::piped())
    }

    fn spawn_card_with(&self, session: &str, args: &[&str], stdout: Stdio) -> std::process::Child {
        use std::os::unix::process::CommandExt;
        self.write_profile();
        let mut all = vec!["card"];
        all.extend_from_slice(args);
        all.extend_from_slice(&["--project", &self.p]);
        let mut cmd: Command = self.home.command(&all);
        cmd.env("MOAI_GIT_BIN", &self.git_bin);
        for (key, value) in self.extra_env.borrow().iter() {
            cmd.env(key, value);
        }
        cmd.env(SESSION_VAR, session)
            .stdin(Stdio::null())
            .stdout(stdout)
            .stderr(Stdio::null())
            .process_group(0);
        cmd.spawn().expect("start moai")
    }
}

/// FU-2 (REQ-4), the real thing: `kill -9` on the process group of a closing command, in the middle
/// of the branch delete. The worktree is gone, the branch and the request stay, the live lease
/// keeps other runs away, and once it ended the next run leaves the `abandoned` line for the dead
/// attempt and settles the request.
#[cfg(unix)]
#[test]
fn fu2_a_killed_executor_leaves_its_lease_and_the_next_attempt_names_it() {
    let w = World::new();
    let wt = w.linked_worktree("wt-killed");
    w.branch("lane/kill");
    let marker = w.outside("kill-marker");
    let hook = w.project.join(".git/hooks/reference-transaction");
    // Stops a delete of the card's branch in its `prepared` phase, as a hung disk would.
    common::write_executable(
        &hook,
        &format!(
            "#!/bin/sh\n[ \"$1\" = prepared ] || exit 0\nwhile read old new ref; do\n  case \"$ref\" in\n    refs/heads/lane/kill) : > '{}'; sleep 30; exit 1 ;;\n  esac\ndone\nexit 0\n",
            marker.display()
        ),
    );
    let card = w.card_in("apps/a/", &wt, Some("lane/kill"), &["one"]);
    w.pass_of_code(&card, &w.head);
    let mut child = w.spawn_card(
        S_A,
        &[
            "finish",
            &card,
            "--head",
            &w.head,
            "--met",
            "1=the test",
            "--by",
            "leader",
        ],
    );
    for _ in 0..200 {
        if marker.exists() {
            break;
        }
        std::thread::sleep(std::time::Duration::from_millis(100));
    }
    assert!(marker.exists(), "the branch delete never started");
    let group = format!("-{}", child.id());
    let killed = Command::new("kill")
        .args(["-KILL", "--", &group])
        .status()
        .expect("run kill");
    assert!(killed.success());
    child.wait().expect("reap the killed command");

    // What the dead run left: the worktree is gone, the branch, the request and its lease stay.
    assert!(!wt.exists() && !w.registered(&wt));
    assert!(w.has_branch("lane/kill"));
    assert_eq!(w.events(&card, "cleanup.started"), 1);
    assert_eq!(w.events(&card, "cleanup.result"), 0);
    assert!(w.steps_of_kind(&card, "worktree").is_empty());
    assert!(w.steps_of_kind(&card, "attempt").is_empty());

    // The hook is gone, the ref lock the killed git left is cleared (what an operator would do),
    // and the lease of the dead attempt is still live: another run leaves the request alone.
    fs::remove_file(&hook).unwrap();
    for lock in ["refs/heads/lane/kill.lock", "packed-refs.lock"] {
        let _ = fs::remove_file(w.project.join(".git").join(lock));
    }
    let busy = w.run_pending();
    assert_eq!(
        World::entry(&busy, &card)["status"],
        json!("busy"),
        "{busy}"
    );
    assert!(w.has_branch("lane/kill"));

    // Time passes: the lease of attempt 1 is over, and the next run names it.
    let file = ledger::ledger_path(w.home.path(), "demo").unwrap();
    let conn = rusqlite::Connection::open(&file).unwrap();
    // Events are append-only in a real ledger; in this throwaway one the guard is lifted so the
    // clock can be moved past the lease without waiting for it.
    conn.execute_batch(
        "DROP TRIGGER event_no_update; \
         UPDATE event SET payload = json_set(payload, '$.lease_until', 1) \
         WHERE kind = 'cleanup.started';",
    )
    .unwrap();
    drop(conn);
    let done = w.run_pending();
    let entry = World::entry(&done, &card);
    assert_eq!(entry["status"], json!("settled"), "{done}");
    assert_eq!(entry["abandoned_attempt"], json!(1), "{done}");
    let lines = w.steps_of_kind(&card, "attempt");
    assert_eq!(lines.len(), 1, "{lines:?}");
    assert_eq!(lines[0]["attempt"], json!(1));
    assert_eq!(lines[0]["status"], json!("abandoned"));
    assert!(!w.has_branch("lane/kill"));
    assert_eq!(w.events(&card, "cleanup.result"), 1);
}

/// FU-1 (REQ-3, audit N-1), the real thing: the closing command is held in the middle of the
/// branch delete while its lease is handed over (the lease of its attempt ends in the ledger), then
/// let go. It deletes the branch, finds its lease ended, and answers `lease_lost` with the
/// worktree and the branch it did remove, and the ledger gets no line from it.
#[cfg(unix)]
#[test]
fn fu1_a_real_executor_whose_lease_ends_under_it_answers_what_it_deleted() {
    let w = World::new();
    let wt = w.linked_worktree("wt-handover");
    w.branch("lane/handover");
    let marker = w.outside("handover-marker");
    let go = w.outside("handover-go");
    let hook = w.project.join(".git/hooks/reference-transaction");
    // Holds the delete of the card's branch in its `prepared` phase until the test lets it go.
    common::write_executable(
        &hook,
        &format!(
            "#!/bin/sh\n[ \"$1\" = prepared ] || exit 0\nwhile read old new ref; do\n  case \"$ref\" in\n    refs/heads/lane/handover) : > '{}'; n=0; while [ ! -e '{}' ] && [ $n -lt 300 ]; do sleep 0.1; n=$((n+1)); done; exit 0 ;;\n  esac\ndone\nexit 0\n",
            marker.display(),
            go.display()
        ),
    );
    let card = w.card_in("apps/h/", &wt, Some("lane/handover"), &["one"]);
    w.pass_of_code(&card, &w.head);
    let child = w.spawn_card_piped(
        S_A,
        &[
            "finish",
            &card,
            "--head",
            &w.head,
            "--met",
            "1=the test",
            "--by",
            "leader",
        ],
    );
    for _ in 0..200 {
        if marker.exists() {
            break;
        }
        std::thread::sleep(std::time::Duration::from_millis(100));
    }
    assert!(marker.exists(), "the branch delete never started");
    assert!(!wt.exists(), "the worktree step ran before the branch step");

    // The lease is handed over: in this throwaway ledger the append-only guard is lifted so the
    // lease of the running attempt can end now.
    let file = ledger::ledger_path(w.home.path(), "demo").unwrap();
    let conn = rusqlite::Connection::open(&file).unwrap();
    conn.execute_batch(
        "DROP TRIGGER event_no_update; \
         UPDATE event SET payload = json_set(payload, '$.lease_until', 1) \
         WHERE kind = 'cleanup.started';",
    )
    .unwrap();
    drop(conn);
    fs::write(&go, b"").unwrap();
    let out = child
        .wait_with_output()
        .expect("wait for the closing command");
    let got = result(&out);
    // `World::outbox` wants an answer without an `error`; a lost lease is one by design.
    let entry = World::entry(&got["cleanup_outbox"], &card);
    eprintln!("LEASE-LOST-ANSWER {entry}");
    assert_eq!(entry["status"], json!("lease_lost"), "{got}");
    assert_eq!(entry["worktrees"][0]["status"], json!("removed"), "{got}");
    assert_eq!(entry["branches"][0]["status"], json!("deleted"), "{got}");
    assert!(cleanup::wrote(&got["cleanup_outbox"]), "{got}");
    assert!(!w.has_branch("lane/handover"), "{got}");

    // The ledger holds the request and the claim of the lost attempt, and nothing more.
    assert_eq!(w.events(&card, "cleanup.requested"), 1);
    assert_eq!(w.events(&card, "cleanup.started"), 1);
    assert_eq!(w.events(&card, "cleanup.result"), 0);
    assert_eq!(w.events(&card, "cleanup.step"), 0);
}
