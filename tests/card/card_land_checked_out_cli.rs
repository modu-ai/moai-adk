//! `moai card land` with the integration branch checked out in a worktree
//! (SPEC-20261125-land-checked-out-branch). The branch moves by the one compare-and-swap, and the
//! worktree that has it checked out is brought to the new commit by git's own two-tree merge; a
//! worktree that cannot be brought along refuses the landing before the first write.
//!
//! Every test runs on a temporary home and a temporary repository whose integration branch is
//! `integ` and is checked out in the project folder itself, the shape of the real repository
//! (the primary checkout always stands on the integration branch). Nothing here touches the real
//! ledger or the real repository.

use crate::common;

use common::home::TempHome;
use moai_cli::ledger;
use rusqlite::Connection;
use serde_json::{json, Value};
use std::cell::RefCell;
use std::collections::BTreeMap;
use std::fs;
use std::io::Write;
use std::path::{Path, PathBuf};
use std::process::{Command, Output, Stdio};

const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
const S_A: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
const S_AUD: &str = "cccccccc-0000-4000-8000-00000000000c";
const S_LEAD: &str = "dddddddd-0000-4000-8000-00000000000d";
const WORKTREE_KEY: &str = "0123456789ab";

/// The profile of an m-flow project whose integration branch is `integ`.
const PROFILE: &[u8] = b"{\"schema_version\": 1, \"project_id\": \"demo\", \"purpose\": \"fixture\", \"tools\": [\"rust\"], \"development_methods\": [\"sdd\"], \"hosts\": [\"claude\"], \"work_flow\": {\"flow\": \"m-flow\", \"integration_branch\": \"integ\"}}";

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

/// `git <args>` in `dir` with `input` on its stdin; the call must succeed.
fn git_with_input(dir: &Path, args: &[&str], input: &str) {
    let mut child = common::git::command()
        .arg("-C")
        .arg(dir)
        .args(args)
        .stdin(Stdio::piped())
        .stdout(Stdio::null())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    child
        .stdin
        .take()
        .unwrap()
        .write_all(input.as_bytes())
        .unwrap();
    let out = child.wait_with_output().unwrap();
    assert!(
        out.status.success(),
        "git {args:?}: {}",
        String::from_utf8_lossy(&out.stderr)
    );
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

/// Prints what a run answered when `C483_SHOW` is set (run with `--no-capture`), so the observed
/// answers of the dangerous path can be copied into the SPEC Result; it checks nothing.
fn show(what: &str, text: &str) {
    if std::env::var_os("C483_SHOW").is_some() {
        eprintln!("SHOW {what}: {text}");
    }
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
    show("answer", &v["result"]["worktree_follow"].to_string());
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

/// A commit message with the trailer block of `session`.
fn signed(subject: &str, session: &str) -> String {
    format!(
        "{subject}\n\nMoai-Spec: SPEC-20261123-fixture\nMoai-Session: {session}:{WORKTREE_KEY}\nMoai-Verified: none"
    )
}

/// Every file of `dir` outside `.git`, path to bytes: what a refused landing must leave as it was.
fn files_of(dir: &Path) -> BTreeMap<String, Vec<u8>> {
    fn walk(root: &Path, at: &Path, out: &mut BTreeMap<String, Vec<u8>>) {
        for entry in fs::read_dir(at).unwrap() {
            let entry = entry.unwrap();
            let path = entry.path();
            let rel = path
                .strip_prefix(root)
                .unwrap()
                .to_string_lossy()
                .into_owned();
            if rel == ".git" {
                continue;
            }
            let kind = entry.file_type().unwrap();
            if kind.is_dir() {
                walk(root, &path, out);
            } else if kind.is_symlink() {
                out.insert(rel, b"<symlink>".to_vec());
            } else {
                out.insert(rel, fs::read(&path).unwrap());
            }
        }
    }
    let mut out = BTreeMap::new();
    walk(dir, dir, &mut out);
    out
}

/// What a landing that refused must leave of the checkout: the files, the index and the status.
#[derive(Debug, PartialEq, Eq)]
struct Checkout {
    files: BTreeMap<String, Vec<u8>>,
    index: String,
    status: String,
    head: String,
}

struct World {
    home: TempHome,
    dir: tempfile::TempDir,
    project: PathBuf,
    p: String,
    git_bin: RefCell<PathBuf>,
    head: String,
    tip: String,
}

impl World {
    /// The fixture of `card_land_cli`, with `integ` checked out in the project folder.
    fn new() -> Self {
        let home = TempHome::new();
        let project = home.project("demo");
        ledger::register(home.path(), "demo").unwrap();
        git(&project, &["init", "-q", "-b", "main"]);
        fs::write(project.join("a.txt"), "a\n").unwrap();
        fs::create_dir(project.join("d")).unwrap();
        fs::write(project.join("d/x"), "x\n").unwrap();
        git(&project, &["add", "a.txt", "d/x"]);
        git(&project, &["commit", "-q", "-m", "init"]);
        fs::write(project.join("b.txt"), "b\n").unwrap();
        git(&project, &["add", "b.txt"]);
        git(
            &project,
            &[
                "commit",
                "-q",
                "-m",
                &format!("the card's change\n\nMoai-Session: {S_A}:{WORKTREE_KEY}"),
            ],
        );
        let head = git(&project, &["rev-parse", "HEAD"]);
        let tree = git(&project, &["rev-parse", &format!("{head}^{{tree}}")]);
        let tip = git(
            &project,
            &[
                "commit-tree",
                &tree,
                "-p",
                &head,
                "-m",
                &format!("integration moved on\n\nMoai-Session: {S_A}:{WORKTREE_KEY}"),
            ],
        );
        git(&project, &["update-ref", "refs/heads/integ", &tip]);
        git(&project, &["checkout", "-q", "integ"]);
        World {
            p: project.to_str().unwrap().to_string(),
            home,
            dir: tempfile::Builder::new()
                .prefix("moai-test-follow-")
                .tempdir()
                .unwrap(),
            project,
            git_bin: RefCell::new(git_executable()),
            head,
            tip,
        }
    }

    /// `moai <args>` as `session`, with the git program named.
    fn run(&self, session: Option<&str>, args: &[&str]) -> Output {
        let mut cmd: Command = self.home.command(args);
        cmd.env("MOAI_GIT_BIN", &*self.git_bin.borrow());
        if let Some(id) = session {
            cmd.env(SESSION_VAR, id);
        }
        cmd.stdin(Stdio::null());
        cmd.output().expect("run moai")
    }

    /// `moai card <args> --project <project>` with the profile written last.
    fn card(&self, session: Option<&str>, args: &[&str]) -> Output {
        let cfg = self
            .project
            .join(moai_cli::PROJECT_STATE_DIR)
            .join("project.yaml");
        fs::create_dir_all(cfg.parent().unwrap()).unwrap();
        fs::write(&cfg, PROFILE).unwrap();
        let mut all = vec!["card"];
        all.extend_from_slice(args);
        all.extend_from_slice(&["--project", &self.p]);
        self.run(session, &all)
    }

    /// A card of kind `task` with one open criterion, taken by `S_A` and sent to review.
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

    /// A lane commit on the card's head that adds `path` with `content` in `mode`, no ref moves.
    fn lane_adding_as(&self, path: &str, mode: &str, content: &str, session: &str) -> String {
        let blob_file = self.project.join("put-on-blob.tmp");
        fs::write(&blob_file, content).unwrap();
        let blob = git(&self.project, &["hash-object", "-w", "put-on-blob.tmp"]);
        fs::remove_file(&blob_file).unwrap();
        let object = if mode == "160000" { &self.head } else { &blob };
        let index = self.project.join(".git/put-on-index");
        let with_index = |args: &[&str]| -> String {
            let out = common::git::command()
                .arg("-C")
                .arg(&self.project)
                .args(args)
                .env("GIT_INDEX_FILE", &index)
                .output()
                .unwrap();
            assert!(
                out.status.success(),
                "git {args:?}: {}",
                String::from_utf8_lossy(&out.stderr)
            );
            String::from_utf8_lossy(&out.stdout).trim().to_string()
        };
        with_index(&["read-tree", &self.head]);
        with_index(&[
            "update-index",
            "--add",
            "--cacheinfo",
            &format!("{mode},{object},{path}"),
        ]);
        let tree = with_index(&["write-tree"]);
        fs::remove_file(&index).unwrap();
        git(
            &self.project,
            &[
                "commit-tree",
                &tree,
                "-p",
                &self.head,
                "-m",
                &signed("the lane change", session),
            ],
        )
    }

    /// A lane commit on the card's head that removes each of `removes` and adds each of `adds`
    /// (path, content) as a regular file, no ref moves.
    fn lane_editing(&self, removes: &[&str], adds: &[(&str, &str)]) -> String {
        let index = self.project.join(".git/put-on-index");
        let with_index = |args: &[&str]| -> String {
            let out = common::git::command()
                .arg("-C")
                .arg(&self.project)
                .args(args)
                .env("GIT_INDEX_FILE", &index)
                .output()
                .unwrap();
            assert!(
                out.status.success(),
                "git {args:?}: {}",
                String::from_utf8_lossy(&out.stderr)
            );
            String::from_utf8_lossy(&out.stdout).trim().to_string()
        };
        with_index(&["read-tree", &self.head]);
        for path in removes {
            with_index(&["update-index", "--force-remove", path]);
        }
        for (path, content) in adds {
            let blob_file = self.project.join("put-on-blob.tmp");
            fs::write(&blob_file, content).unwrap();
            let blob = git(&self.project, &["hash-object", "-w", "put-on-blob.tmp"]);
            fs::remove_file(&blob_file).unwrap();
            with_index(&[
                "update-index",
                "--add",
                "--cacheinfo",
                &format!("100644,{blob},{path}"),
            ]);
        }
        let tree = with_index(&["write-tree"]);
        fs::remove_file(&index).unwrap();
        git(
            &self.project,
            &[
                "commit-tree",
                &tree,
                "-p",
                &self.head,
                "-m",
                &signed("the lane change", S_A),
            ],
        )
    }

    /// A card, the lane commit of `lane_editing`, its join on the tip and the PASS of it.
    fn ready_editing(&self, removes: &[&str], adds: &[(&str, &str)]) -> (String, String) {
        let card = self.review_card();
        let lane = self.lane_editing(removes, adds);
        let cand = self.join(&lane);
        self.pass_of(&card, &cand);
        (card, cand)
    }

    fn lane_adding(&self, path: &str) -> String {
        self.lane_adding_as(path, "100644", "lane\n", S_A)
    }

    /// The commit the leader lands: the lane on the tip, the tree git makes of the two, signed by
    /// the lander.
    fn join(&self, lane: &str) -> String {
        let tree = git(
            &self.project,
            &["merge-tree", "--write-tree", &self.tip, lane],
        );
        git(
            &self.project,
            &[
                "commit-tree",
                &tree,
                "-p",
                &self.tip,
                "-p",
                lane,
                "-m",
                &signed("merge the lane on the tip", S_LEAD),
            ],
        )
    }

    /// An independent PASS of `head` with a green receipt, written as `card verdict` writes it.
    fn pass_of(&self, card: &str, head: &str) {
        let conn =
            Connection::open(ledger::ledger_path(self.home.path(), "demo").unwrap()).unwrap();
        let rows = self.count("SELECT count(*) FROM event");
        let payload = json!({
            "head": head, "base": self.tip, "verdict": "pass", "tier": 1,
            "evidence": "the fixture's evidence",
            "receipt": {"head": head, "steps": [{"name": "fmt", "exit": 0}, {"name": "test", "exit": 0}]},
            "reviewer": S_AUD,
        });
        conn.execute(
            "INSERT INTO event (event_key, at, actor, card_id, kind, payload) \
             VALUES (?1, unixepoch(), ?2, ?3, 'card.verdict', ?4)",
            rusqlite::params![
                format!("test-verdict-{rows}"),
                S_AUD,
                number(card),
                payload.to_string()
            ],
        )
        .unwrap();
    }

    /// `card land` of `cand` as the lander session.
    fn land(&self, card: &str, cand: &str) -> Output {
        self.card(
            Some(S_LEAD),
            &[
                "land",
                card,
                "--head",
                cand,
                "--met",
                "1=the test",
                "--by",
                "leader",
            ],
        )
    }

    /// A review card, a lane commit adding `notes/x.md`, its join on the tip and the PASS of it.
    fn ready_to_land(&self) -> (String, String) {
        let card = self.review_card();
        let lane = self.lane_adding("notes/x.md");
        let cand = self.join(&lane);
        self.pass_of(&card, &cand);
        (card, cand)
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

    fn state(&self, card: &str) -> String {
        let l = ledger::open_read_only(self.home.path(), "demo").unwrap();
        l.conn()
            .query_row(
                "SELECT state FROM card WHERE card_id = ?1",
                [number(card)],
                |r| r.get(0),
            )
            .unwrap()
    }

    fn events(&self, card: &str, kind: &str) -> i64 {
        self.count(&format!(
            "SELECT count(*) FROM event WHERE card_id = {} AND kind = '{kind}'",
            number(card)
        ))
    }

    /// What the checkout holds now, for a before and after of a refusal.
    fn checkout(&self) -> Checkout {
        Checkout {
            files: files_of(&self.project),
            index: git(&self.project, &["ls-files", "-s"]),
            status: git(&self.project, &["status", "--porcelain=v1", "--ignored"]),
            head: git(&self.project, &["rev-parse", "HEAD"]),
        }
    }

    /// The tracked changes of the checkout (`git status` without untracked files).
    fn tracked_changes(&self) -> String {
        git(
            &self.project,
            &["status", "--porcelain=v1", "--untracked-files=no"],
        )
    }

    /// Asserts a refusal of the worktree check: exit 1, `conflict`, a reason that holds each of
    /// `reasons`, a next step and `nothing was changed`; the branch, the ledger and the checkout
    /// are as they were.
    fn refuses(
        &self,
        out: &Output,
        reasons: &[&str],
        before: &(Vec<String>, Checkout),
        what: &str,
    ) -> String {
        assert_eq!(
            out.status.code(),
            Some(1),
            "{what}: {} {}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        );
        let v = envelope(out);
        assert_eq!(v["ok"], json!(false), "{what}: {v}");
        assert_eq!(v["error"]["kind"], json!("conflict"), "{what}: {v}");
        let message = v["error"]["message"].as_str().unwrap().to_string();
        show(what, &message);
        for reason in reasons {
            assert!(
                message.contains(reason),
                "{what}: the reason lacks {reason:?}: {message}"
            );
        }
        assert!(
            message.contains("nothing was changed"),
            "{what}: no `nothing was changed`: {message}"
        );
        assert!(
            ["run the call again", "land again", "land without it"]
                .iter()
                .any(|step| message.contains(step)),
            "{what}: no next step: {message}"
        );
        assert_eq!(
            self.integ(),
            self.tip,
            "{what}: the refused call moved integ"
        );
        assert_eq!(&self.dump(), &before.0, "{what}: the ledger changed");
        assert_eq!(&self.checkout(), &before.1, "{what}: the checkout changed");
        message
    }

    /// The state to compare a refusal with, taken after the PASS was recorded.
    fn before(&self) -> (Vec<String>, Checkout) {
        (self.dump(), self.checkout())
    }

    /// A git wrapper that runs `before` (shell text) just before the real `update-ref` and
    /// `after` just after it, then returns the real exit code.
    fn wrap_update_ref(&self, name: &str, after: &str) {
        let wrapper = self.dir.path().join(name);
        common::write_executable(
            &wrapper,
            &format!(
                "#!/bin/sh\nfor a in \"$@\"; do\n  if [ \"$a\" = update-ref ]; then\n    '{git}' \"$@\"\n    rc=$?\n    {after}\n    exit $rc\n  fi\ndone\nexec '{git}' \"$@\"\n",
                git = git_executable().display(),
            ),
        );
        *self.git_bin.borrow_mut() = wrapper;
    }

    /// A git wrapper around `read-tree`: `dry` (shell text) runs for the dry run `-n` and may
    /// exit; for the real run `before` runs first, then git, then `after`, and the real exit code
    /// is returned. Every other git command runs as it is.
    fn wrap_read_tree(&self, name: &str, dry: &str, before: &str, after: &str) {
        let wrapper = self.dir.path().join(name);
        common::write_executable(
            &wrapper,
            &format!(
                "#!/bin/sh\nis_rt=0\ndry=0\nfor a in \"$@\"; do\n  [ \"$a\" = read-tree ] && is_rt=1\n  [ \"$a\" = -n ] && dry=1\ndone\nif [ $is_rt = 1 ]; then\n  if [ $dry = 1 ]; then\n    {dry}\n  else\n    {before}\n    '{git}' \"$@\"\n    rc=$?\n    {after}\n    exit $rc\n  fi\nfi\nexec '{git}' \"$@\"\n",
                git = git_executable().display(),
            ),
        );
        *self.git_bin.borrow_mut() = wrapper;
    }

    /// A git wrapper that fails the real `read-tree` (the follow step; the dry run passes).
    fn wrap_read_tree_failing(&self, name: &str) {
        self.wrap_read_tree(
            name,
            ":",
            "echo 'fatal: the fixture made read-tree fail' >&2; exit 128",
            ":",
        );
    }

    /// A git wrapper that answers `rev-parse HEAD` with another commit.
    fn wrap_head_elsewhere(&self, name: &str) {
        let wrapper = self.dir.path().join(name);
        common::write_executable(
            &wrapper,
            &format!(
                "#!/bin/sh\nis_rp=0\nis_head=0\nfor a in \"$@\"; do\n  [ \"$a\" = rev-parse ] && is_rp=1\n  [ \"$a\" = HEAD ] && is_head=1\ndone\nif [ $is_rp = 1 ] && [ $is_head = 1 ]; then\n  echo fedcba9876543210fedcba9876543210fedcba98\n  exit 0\nfi\nexec '{git}' \"$@\"\n",
                git = git_executable().display(),
            ),
        );
        *self.git_bin.borrow_mut() = wrapper;
    }

    fn plain_git(&self) {
        *self.git_bin.borrow_mut() = git_executable();
    }
}

fn same_path(a: &str, b: &Path) -> bool {
    fs::canonicalize(a).unwrap() == fs::canonicalize(b).unwrap()
}

// ---- AC-1: a clean checkout follows ------------------------------------------------------------

/// AC-1 (REQ-1, REQ-2, REQ-9): the branch, the index and the working tree equal the candidate, an
/// untracked file that does not overlap stays, the card closes and the answer names the worktree.
#[test]
fn a_clean_checkout_follows_the_branch_and_the_card_closes() {
    let w = World::new();
    fs::write(w.project.join("scratch.txt"), "mine\n").unwrap();
    let (card, cand) = w.ready_to_land();
    let got = result(&w.land(&card, &cand));
    assert_eq!(got["moved"], json!(true), "{got}");
    assert_eq!(got["closed"], json!(true), "{got}");
    assert_eq!(w.integ(), cand);
    assert_eq!(w.state(&card), "done");
    let follow = &got["worktree_follow"];
    assert_eq!(follow["status"], json!("followed"), "{got}");
    assert!(
        same_path(follow["worktree"].as_str().unwrap(), &w.project),
        "{follow}"
    );
    assert_eq!(follow["from"], json!(w.tip), "{follow}");
    assert_eq!(follow["to"], json!(cand), "{follow}");
    // The index and the working tree are the candidate's.
    assert_eq!(git(&w.project, &["rev-parse", "HEAD"]), cand);
    assert_eq!(w.tracked_changes(), "", "tracked changes after the landing");
    assert_eq!(
        fs::read_to_string(w.project.join("notes/x.md")).unwrap(),
        "lane\n"
    );
    assert_eq!(
        git(
            &w.project,
            &["diff-index", "--cached", "--name-only", &cand]
        ),
        ""
    );
    assert_eq!(
        fs::read_to_string(w.project.join("scratch.txt")).unwrap(),
        "mine\n",
        "an untracked file elsewhere stays"
    );
}

// ---- AC-2: a tracked change refuses -------------------------------------------------------------

/// AC-2 (REQ-3, REQ-7): a staged change, an unstaged change and a deleted tracked file each
/// refuse before the first write and name the path and the way out.
#[test]
fn a_tracked_change_refuses_and_changes_nothing() {
    type Mutate = fn(&World);
    let cases: [(&str, Mutate, &str); 3] = [
        (
            "a staged change",
            |w| {
                fs::write(w.project.join("b.txt"), "staged\n").unwrap();
                git(&w.project, &["add", "b.txt"]);
            },
            "b.txt",
        ),
        (
            "an unstaged change",
            |w| fs::write(w.project.join("a.txt"), "edited\n").unwrap(),
            "a.txt",
        ),
        (
            "a deleted tracked file",
            |w| fs::remove_file(w.project.join("b.txt")).unwrap(),
            "b.txt",
        ),
    ];
    for (what, mutate, path) in cases {
        let w = World::new();
        let (card, cand) = w.ready_to_land();
        mutate(&w);
        let before = w.before();
        let out = w.land(&card, &cand);
        w.refuses(&out, &["tracked change", path], &before, what);
    }
}

// ---- AC-3: a moved HEAD or an operation in progress ---------------------------------------------

/// AC-3 (REQ-4, REQ-7): each mark of a merge, a rebase, a cherry-pick, a revert, a bisect or an
/// index lock refuses, and so does a worktree whose HEAD is not the tip the landing read.
#[test]
fn a_moved_head_or_a_pending_operation_refuses_and_changes_nothing() {
    let w = World::new();
    let (card, cand) = w.ready_to_land();
    let git_dir = w.project.join(".git");
    let before = w.before();
    for name in [
        "MERGE_HEAD",
        "CHERRY_PICK_HEAD",
        "REVERT_HEAD",
        "REBASE_HEAD",
        "BISECT_LOG",
        "index.lock",
    ] {
        fs::write(git_dir.join(name), "").unwrap();
        let out = w.land(&card, &cand);
        w.refuses(&out, &["operation in progress", name], &before, name);
        fs::remove_file(git_dir.join(name)).unwrap();
    }
    for name in ["rebase-merge", "rebase-apply", "sequencer"] {
        fs::create_dir(git_dir.join(name)).unwrap();
        let out = w.land(&card, &cand);
        w.refuses(&out, &["operation in progress", name], &before, name);
        fs::remove_dir(git_dir.join(name)).unwrap();
    }
    // A HEAD that is another commit than the tip the landing read.
    w.wrap_head_elsewhere("git-head-elsewhere");
    let out = w.land(&card, &cand);
    w.refuses(
        &out,
        &["fedcba9876543210fedcba9876543210fedcba98", "not the tip"],
        &before,
        "a HEAD that is not the tip",
    );
}

// ---- AC-4: a file at a path the candidate adds --------------------------------------------------

/// AC-4 (REQ-5, REQ-7): an untracked file, an ignored file and a file where the candidate needs a
/// folder each refuse; an untracked file elsewhere does not stop the landing.
#[test]
fn an_untracked_file_at_an_added_path_refuses_and_other_untracked_files_stay() {
    let w = World::new();
    let (card, cand) = w.ready_to_land();
    fs::write(w.project.join("elsewhere.txt"), "elsewhere\n").unwrap();

    fs::create_dir(w.project.join("notes")).unwrap();
    let blocker = w.project.join("notes/x.md");
    fs::write(&blocker, "mine\n").unwrap();
    let before = w.before();
    let out = w.land(&card, &cand);
    w.refuses(
        &out,
        &["no commit tracks", "notes/x.md"],
        &before,
        "an untracked file at an added path",
    );

    // The same file, ignored by the repository's own exclude list.
    let exclude = w.project.join(".git/info/exclude");
    fs::write(&exclude, "notes/x.md\n").unwrap();
    let before = w.before();
    let out = w.land(&card, &cand);
    w.refuses(
        &out,
        &["no commit tracks", "notes/x.md"],
        &before,
        "an ignored file at an added path",
    );
    fs::write(&exclude, "").unwrap();
    fs::remove_file(&blocker).unwrap();
    fs::remove_dir(w.project.join("notes")).unwrap();

    // A file where the candidate needs the folder `notes`.
    fs::write(w.project.join("notes"), "a file\n").unwrap();
    let before = w.before();
    let out = w.land(&card, &cand);
    w.refuses(
        &out,
        &["needs a folder", "notes"],
        &before,
        "a file where a folder is needed",
    );
    fs::remove_file(w.project.join("notes")).unwrap();

    // Nothing in the way: the untracked file elsewhere does not stop it.
    let got = result(&w.land(&card, &cand));
    assert_eq!(got["worktree_follow"]["status"], json!("followed"), "{got}");
    assert_eq!(
        fs::read_to_string(w.project.join("elsewhere.txt")).unwrap(),
        "elsewhere\n"
    );
}

// ---- AC-5: an unsupported checkout --------------------------------------------------------------

/// AC-5 (REQ-6, REQ-7): two worktrees on the branch, a sparse checkout, a skip-worktree entry, an
/// assume-unchanged entry and a submodule change each refuse before the first write.
#[test]
fn an_unsupported_checkout_refuses_and_changes_nothing() {
    type Prepare = fn(&World) -> (String, String);
    let cases: [(&str, Prepare, &[&str]); 6] = [
        (
            "an unmerged entry",
            |w| {
                let (card, cand) = w.ready_to_land();
                let blob = git(&w.project, &["rev-parse", "HEAD:a.txt"]);
                let zero = "0".repeat(40);
                git_with_input(
                    &w.project,
                    &["update-index", "--index-info"],
                    &format!(
                        "0 {zero}\ta.txt\n100644 {blob} 1\ta.txt\n100644 {blob} 2\ta.txt\n100644 {blob} 3\ta.txt\n"
                    ),
                );
                (card, cand)
            },
            &["unmerged paths: a.txt;"],
        ),
        (
            "two worktrees on the branch",
            |w| {
                let other = w.dir.path().join("second-checkout");
                git(
                    &w.project,
                    &[
                        "worktree",
                        "add",
                        "-q",
                        "--force",
                        other.to_str().unwrap(),
                        "integ",
                    ],
                );
                w.ready_to_land()
            },
            &["more than one worktree"],
        ),
        (
            "a sparse checkout",
            |w| {
                git(&w.project, &["config", "core.sparseCheckout", "true"]);
                w.ready_to_land()
            },
            &["sparse checkout"],
        ),
        (
            "a skip-worktree entry",
            |w| {
                git(&w.project, &["update-index", "--skip-worktree", "a.txt"]);
                w.ready_to_land()
            },
            &["skip-worktree", "a.txt"],
        ),
        (
            "an assume-unchanged entry",
            |w| {
                git(&w.project, &["update-index", "--assume-unchanged", "a.txt"]);
                w.ready_to_land()
            },
            &["assume-unchanged", "a.txt"],
        ),
        (
            "a submodule entry",
            |w| {
                let card = w.review_card();
                let lane = w.lane_adding_as("vendor/sub", "160000", "", S_A);
                let cand = w.join(&lane);
                w.pass_of(&card, &cand);
                (card, cand)
            },
            &["submodule entry", "vendor/sub"],
        ),
    ];
    for (what, prepare, reasons) in cases {
        let w = World::new();
        let (card, cand) = prepare(&w);
        let before = w.before();
        let out = w.land(&card, &cand);
        w.refuses(&out, reasons, &before, what);
    }
}

// ---- AC-6, AC-7: a stopped run ------------------------------------------------------------------

/// The commits git's reference log of `integ` records for `card land`.
fn landings_in_reflog(w: &World) -> usize {
    git(
        &w.project,
        &["reflog", "show", "--format=%gs", "refs/heads/integ"],
    )
    .lines()
    .filter(|line| line.contains("moai card land"))
    .count()
}

/// AC-6 (REQ-8): a run whose follow step fails after the branch moved leaves the card open; the
/// same call run again finishes the worktree, moves the branch no second time and closes.
#[test]
fn a_stopped_run_is_finished_by_the_same_call_and_moves_once() {
    let w = World::new();
    let (card, cand) = w.ready_to_land();
    w.wrap_read_tree_failing("git-read-tree-fails");
    let out = w.land(&card, &cand);
    assert_eq!(
        out.status.code(),
        Some(1),
        "{}",
        String::from_utf8_lossy(&out.stdout)
    );
    let message = envelope(&out)["error"]["message"]
        .as_str()
        .unwrap()
        .to_string();
    assert!(
        message.contains("could not follow") && message.contains("run the same call again"),
        "{message}"
    );
    assert_eq!(w.integ(), cand, "the branch moved");
    assert_eq!(w.state(&card), "review", "the card is still open");
    assert_eq!(w.events(&card, "card.landed"), 0);
    assert!(
        !w.tracked_changes().is_empty(),
        "the worktree still lags behind the branch"
    );

    w.plain_git();
    let got = result(&w.land(&card, &cand));
    assert_eq!(got["moved"], json!(false), "{got}");
    assert_eq!(got["worktree_follow"]["status"], json!("finished"), "{got}");
    assert_eq!(got["worktree_follow"]["from"], json!(w.tip), "{got}");
    assert_eq!(got["worktree_follow"]["to"], json!(cand), "{got}");
    assert_eq!(w.integ(), cand);
    assert_eq!(landings_in_reflog(&w), 1, "the branch moved more than once");
    assert_eq!(w.events(&card, "card.land_started"), 1);
    assert_eq!(w.state(&card), "done");
    assert_eq!(w.tracked_changes(), "");
    assert_eq!(
        fs::read_to_string(w.project.join("notes/x.md")).unwrap(),
        "lane\n"
    );
}

/// AC-7 (REQ-8): a run killed with SIGKILL between the branch move and the follow step is finished
/// the same way by the same call.
#[test]
fn a_killed_run_is_finished_by_the_same_call() {
    let w = World::new();
    let (card, cand) = w.ready_to_land();
    w.wrap_update_ref("git-kills-the-lander", "kill -9 $PPID");
    let out = w.land(&card, &cand);
    assert_ne!(out.status.code(), Some(0), "the lander was not killed");
    assert_eq!(w.integ(), cand, "the branch moved");
    assert_eq!(w.events(&card, "card.landed"), 0);
    assert!(
        !w.tracked_changes().is_empty(),
        "the worktree lags behind the branch"
    );

    w.plain_git();
    let got = result(&w.land(&card, &cand));
    assert_eq!(got["moved"], json!(false), "{got}");
    assert_eq!(got["worktree_follow"]["status"], json!("finished"), "{got}");
    assert_eq!(w.integ(), cand);
    assert_eq!(landings_in_reflog(&w), 1);
    assert_eq!(w.events(&card, "card.land_started"), 1);
    assert_eq!(w.state(&card), "done");
    assert_eq!(w.tracked_changes(), "");
    assert_eq!(
        fs::read_to_string(w.project.join("notes/x.md")).unwrap(),
        "lane\n"
    );
}

/// REQ-8, D-5: a worktree whose index is neither the old tip's tree nor the new one's was changed
/// by someone; the rerun leaves it alone, says so and closes the card.
#[test]
fn a_rerun_leaves_a_worktree_alone_whose_index_matches_neither_commit() {
    let w = World::new();
    let (card, cand) = w.ready_to_land();
    w.wrap_update_ref("git-kills-the-lander", "kill -9 $PPID");
    let out = w.land(&card, &cand);
    assert_ne!(out.status.code(), Some(0), "the lander was not killed");
    // Someone stages a file of their own while the index still lags, and edits a tracked file
    // without staging it.
    fs::write(w.project.join("mine.txt"), "mine\n").unwrap();
    git(&w.project, &["add", "mine.txt"]);
    fs::write(w.project.join("d/x"), "edited, never staged\n").unwrap();
    let staged = w.checkout();

    w.plain_git();
    let got = result(&w.land(&card, &cand));
    assert_eq!(got["worktree_follow"]["status"], json!("none"), "{got}");
    let note = got["worktree_follow"]["note"].as_str().unwrap_or_default();
    let old = got["worktree_follow"]["from"].as_str().unwrap();
    assert!(
        note.contains("neither")
            && note.contains(&format!("git diff --binary {old} > keep.patch"))
            && !note.contains("--cached")
            && note.contains("staged or not")
            && note.contains("git reset --hard")
            && note.contains("git apply --3way keep.patch"),
        "the note names a way out that saves the unstaged edits too: {got}"
    );
    assert_eq!(w.checkout(), staged, "the rerun touched the worktree");
    assert_eq!(w.state(&card), "done");

    // The advice, followed to the letter, loses neither the staged file nor the unstaged edit.
    let patch = git(&w.project, &["diff", "--binary", old]);
    assert!(
        patch.contains("mine.txt") && patch.contains("edited, never staged"),
        "the saved patch lacks a change: {patch}"
    );
    fs::write(w.project.join("keep.patch"), format!("{patch}\n")).unwrap();
    git(&w.project, &["reset", "--hard"]);
    git(&w.project, &["apply", "--3way", "keep.patch"]);
    assert_eq!(
        fs::read_to_string(w.project.join("mine.txt")).unwrap(),
        "mine\n"
    );
    assert_eq!(
        fs::read_to_string(w.project.join("d/x")).unwrap(),
        "edited, never staged\n"
    );
}

/// The advice of a lagging foreign index ends in `git reset --hard`, which also overwrites a file
/// no commit tracks (ignored or not) where the new commit adds one, and which no patch holds: the
/// note names such files, and the advice followed to the letter (move them, save, reset, apply)
/// loses neither them nor a staged binary file.
#[test]
fn a_rerun_note_names_the_files_the_reset_would_overwrite() {
    let w = World::new();
    let (card, cand) = w.ready_editing(&[], &[("notes/x.md", "lane\n"), ("keep.ign", "lane\n")]);
    w.wrap_update_ref("git-kills-the-lander", "kill -9 $PPID");
    let out = w.land(&card, &cand);
    assert_ne!(out.status.code(), Some(0), "the lander was not killed");
    // The owner keeps a file of their own where the new commit adds a file, an ignored one at
    // another added path, and stages a binary file: the index matches neither commit.
    fs::write(w.project.join(".git/info/exclude"), "keep.ign\n").unwrap();
    fs::create_dir_all(w.project.join("notes")).unwrap();
    fs::write(w.project.join("notes/x.md"), "mine, untracked\n").unwrap();
    fs::write(w.project.join("keep.ign"), "mine, ignored\n").unwrap();
    let binary: &[u8] = &[0, 1, 2, 254, 255, 0];
    fs::write(w.project.join("bin.dat"), binary).unwrap();
    git(&w.project, &["add", "bin.dat"]);
    let before = w.checkout();

    w.plain_git();
    let got = result(&w.land(&card, &cand));
    assert_eq!(got["worktree_follow"]["status"], json!("none"), "{got}");
    let note = got["worktree_follow"]["note"].as_str().unwrap_or_default();
    let old = got["worktree_follow"]["from"].as_str().unwrap();
    assert!(
        note.contains("Move these away before the reset")
            && note.contains("notes/x.md")
            && note.contains("keep.ign")
            && note.contains(&format!("git diff --binary {old} > keep.patch")),
        "the note must name the files the reset overwrites: {got}"
    );
    assert_eq!(w.checkout(), before, "the rerun touched the worktree");
    assert_eq!(w.state(&card), "done");

    // The advice, to the letter: move the named files away, save, reset, apply.
    let saved = w.dir.path().join("saved");
    fs::create_dir_all(&saved).unwrap();
    fs::rename(w.project.join("notes/x.md"), saved.join("x.md")).unwrap();
    fs::rename(w.project.join("keep.ign"), saved.join("keep.ign")).unwrap();
    // `> keep.patch` keeps the bytes as they are; the `git` helper trims the blank line that ends
    // a binary hunk.
    let diff = common::git::command()
        .arg("-C")
        .arg(&w.project)
        .args(["diff", "--binary", old])
        .output()
        .unwrap();
    assert!(diff.status.success());
    fs::write(w.project.join("keep.patch"), &diff.stdout).unwrap();
    git(&w.project, &["reset", "--hard"]);
    git(&w.project, &["apply", "--3way", "keep.patch"]);
    assert_eq!(fs::read(w.project.join("bin.dat")).unwrap(), binary);
    assert_eq!(
        fs::read_to_string(saved.join("x.md")).unwrap(),
        "mine, untracked\n"
    );
    assert_eq!(
        fs::read_to_string(saved.join("keep.ign")).unwrap(),
        "mine, ignored\n"
    );
    assert_eq!(
        fs::read_to_string(w.project.join("notes/x.md")).unwrap(),
        "lane\n",
        "the reset brings the new commit's file"
    );
}

// ---- the rest of the guards: each rule has a test that only it fails -----------------------------

impl World {
    /// A lane commit on the card's head that replaces the tracked file `a.txt` by a folder that
    /// holds `a.txt/x`.
    fn lane_replacing_a_file_by_a_folder(&self) -> String {
        let blob_file = self.project.join("put-on-blob.tmp");
        fs::write(&blob_file, "inside\n").unwrap();
        let blob = git(&self.project, &["hash-object", "-w", "put-on-blob.tmp"]);
        fs::remove_file(&blob_file).unwrap();
        let index = self.project.join(".git/put-on-index");
        let with_index = |args: &[&str]| -> String {
            let out = common::git::command()
                .arg("-C")
                .arg(&self.project)
                .args(args)
                .env("GIT_INDEX_FILE", &index)
                .output()
                .unwrap();
            assert!(
                out.status.success(),
                "git {args:?}: {}",
                String::from_utf8_lossy(&out.stderr)
            );
            String::from_utf8_lossy(&out.stdout).trim().to_string()
        };
        with_index(&["read-tree", &self.head]);
        with_index(&["update-index", "--force-remove", "a.txt"]);
        with_index(&[
            "update-index",
            "--add",
            "--cacheinfo",
            &format!("100644,{blob},a.txt/x"),
        ]);
        let tree = with_index(&["write-tree"]);
        fs::remove_file(&index).unwrap();
        git(
            &self.project,
            &[
                "commit-tree",
                &tree,
                "-p",
                &self.head,
                "-m",
                &signed("the lane change", S_A),
            ],
        )
    }
}

/// A tracked file the candidate deletes and replaces by a folder is no obstacle: git removes the
/// file and the worktree follows.
#[test]
fn a_tracked_file_the_candidate_replaces_by_a_folder_follows() {
    let w = World::new();
    let card = w.review_card();
    let lane = w.lane_replacing_a_file_by_a_folder();
    let cand = w.join(&lane);
    w.pass_of(&card, &cand);
    let got = result(&w.land(&card, &cand));
    assert_eq!(got["worktree_follow"]["status"], json!("followed"), "{got}");
    assert_eq!(
        fs::read_to_string(w.project.join("a.txt/x")).unwrap(),
        "inside\n"
    );
    assert_eq!(w.tracked_changes(), "");
}

/// A refusal of git's own dry run (a case the command's checks did not name) comes before the
/// move and changes nothing.
#[test]
fn a_dry_run_that_git_refuses_stops_the_landing_before_the_move() {
    let w = World::new();
    let (card, cand) = w.ready_to_land();
    w.wrap_read_tree(
        "git-dry-run-fails",
        "echo 'error: Entry not uptodate. Cannot merge.' >&2; exit 128",
        ":",
        ":",
    );
    let before = w.before();
    let out = w.land(&card, &cand);
    w.refuses(
        &out,
        &["git would refuse to bring", "Cannot merge"],
        &before,
        "a dry run that git refuses",
    );
}

/// A `read-tree` that exits 0 and leaves the index where it was is not a follow: the step checks
/// the index against the candidate and says the branch moved and the call can be run again.
#[test]
fn a_follow_that_leaves_the_index_behind_is_refused_after_the_move() {
    let w = World::new();
    let (card, cand) = w.ready_to_land();
    w.wrap_read_tree("git-read-tree-does-nothing", ":", "exit 0", ":");
    let out = w.land(&card, &cand);
    assert_eq!(
        out.status.code(),
        Some(1),
        "{}",
        String::from_utf8_lossy(&out.stdout)
    );
    let message = envelope(&out)["error"]["message"]
        .as_str()
        .unwrap()
        .to_string();
    assert!(
        message.contains("does not equal") && message.contains("run the same call again"),
        "{message}"
    );
    assert_eq!(w.integ(), cand, "the branch moved");
    assert_eq!(w.state(&card), "review");
}

/// A run killed after the follow, before the close: the rerun finds the index already at the new
/// commit, follows nothing a second time and closes the card.
#[test]
fn a_run_killed_after_the_follow_closes_on_the_rerun_without_following_again() {
    let w = World::new();
    let (card, cand) = w.ready_to_land();
    w.wrap_read_tree("git-kills-after-the-follow", ":", ":", "kill -9 $PPID");
    let out = w.land(&card, &cand);
    assert_ne!(out.status.code(), Some(0), "the lander was not killed");
    assert_eq!(w.integ(), cand);
    assert_eq!(w.tracked_changes(), "", "the follow had finished");
    assert_eq!(w.events(&card, "card.landed"), 0);

    w.plain_git();
    let got = result(&w.land(&card, &cand));
    assert_eq!(got["worktree_follow"]["status"], json!("none"), "{got}");
    assert_eq!(got["worktree_follow"]["note"], Value::Null, "{got}");
    assert_eq!(w.state(&card), "done");
    assert_eq!(landings_in_reflog(&w), 1);
}

/// A rerun refuses what a first run would refuse: a file that no commit tracks now stands at a
/// path the candidate adds, or a tracked file was edited while the index lags.
#[test]
fn a_rerun_refuses_what_a_first_run_would_refuse() {
    let w = World::new();
    let (card, cand) = w.ready_to_land();
    w.wrap_update_ref("git-kills-the-lander", "kill -9 $PPID");
    let out = w.land(&card, &cand);
    assert_ne!(out.status.code(), Some(0), "the lander was not killed");
    w.plain_git();

    // An untracked file at an added path.
    fs::create_dir(w.project.join("notes")).unwrap();
    fs::write(w.project.join("notes/x.md"), "mine\n").unwrap();
    let before = w.checkout();
    let out = w.land(&card, &cand);
    assert_eq!(
        out.status.code(),
        Some(1),
        "{}",
        String::from_utf8_lossy(&out.stdout)
    );
    let message = envelope(&out)["error"]["message"]
        .as_str()
        .unwrap()
        .to_string();
    assert!(
        message.contains("notes/x.md") && message.contains("run the same call again"),
        "{message}"
    );
    assert_eq!(
        w.checkout(),
        before,
        "the refused rerun touched the worktree"
    );
    assert_eq!(w.state(&card), "review");

    // The same file ignored by the exclude list: git's dry run passes it, so only the command's
    // own check of the added paths refuses it.
    let exclude = w.project.join(".git/info/exclude");
    fs::write(&exclude, "notes/x.md\n").unwrap();
    let before = w.checkout();
    let out = w.land(&card, &cand);
    assert_eq!(
        out.status.code(),
        Some(1),
        "{}",
        String::from_utf8_lossy(&out.stdout)
    );
    let message = envelope(&out)["error"]["message"]
        .as_str()
        .unwrap()
        .to_string();
    assert!(
        message.contains("no commit tracks") && message.contains("notes/x.md"),
        "{message}"
    );
    assert_eq!(
        w.checkout(),
        before,
        "the refused rerun touched the worktree"
    );
    fs::write(&exclude, "").unwrap();
    fs::remove_file(w.project.join("notes/x.md")).unwrap();
    fs::remove_dir(w.project.join("notes")).unwrap();

    // A tracked file edited while the index lags.
    fs::write(w.project.join("a.txt"), "edited\n").unwrap();
    let before = w.checkout();
    let out = w.land(&card, &cand);
    assert_eq!(
        out.status.code(),
        Some(1),
        "{}",
        String::from_utf8_lossy(&out.stdout)
    );
    let message = envelope(&out)["error"]["message"]
        .as_str()
        .unwrap()
        .to_string();
    assert!(
        message.contains("edited tracked files") && message.contains("git restore ."),
        "{message}"
    );
    assert_eq!(w.checkout(), before);
    fs::write(w.project.join("a.txt"), "a\n").unwrap();

    // Nothing in the way any more: the same call finishes it.
    let got = result(&w.land(&card, &cand));
    assert_eq!(got["worktree_follow"]["status"], json!("finished"), "{got}");
    assert_eq!(w.state(&card), "done");
}

// ---- the repair of the first audit (B1, B2) --------------------------------------------------

/// B1: an ignored file that appears after the plan (the ledger rows, the scans and the branch move
/// come in between) at a path the head adds is never overwritten. The branch already moved, so the
/// refusal says so and keeps the card open; the same call run again after the cause is gone
/// finishes it.
#[test]
fn an_ignored_file_that_appears_after_the_plan_is_never_overwritten() {
    let w = World::new();
    let (card, cand) = w.ready_to_land();
    fs::write(w.project.join(".git/info/exclude"), "notes/x.md\n").unwrap();
    let p = w.p.clone();
    w.wrap_update_ref(
        "git-an-ignored-file-appears",
        &format!("mkdir -p '{p}/notes'; echo precious > '{p}/notes/x.md'"),
    );
    let out = w.land(&card, &cand);
    assert_eq!(
        out.status.code(),
        Some(1),
        "{}",
        String::from_utf8_lossy(&out.stdout)
    );
    let message = envelope(&out)["error"]["message"]
        .as_str()
        .unwrap()
        .to_string();
    show("an ignored file after the plan", &message);
    for part in [
        "no commit tracks",
        "notes/x.md",
        "already moved",
        "run the same call again",
    ] {
        assert!(message.contains(part), "lacks {part:?}: {message}");
    }
    assert_eq!(w.integ(), cand, "the branch moved");
    assert_eq!(w.state(&card), "review", "the card stays open");
    assert_eq!(
        fs::read_to_string(w.project.join("notes/x.md")).unwrap(),
        "precious\n",
        "the ignored file was overwritten"
    );
    assert!(
        !w.tracked_changes().is_empty(),
        "the worktree still lags behind the branch"
    );

    // The same call again with the file still there refuses the same way and touches nothing.
    w.plain_git();
    let before = w.checkout();
    let out = w.land(&card, &cand);
    assert_eq!(out.status.code(), Some(1));
    assert_eq!(
        w.checkout(),
        before,
        "the refused rerun touched the worktree"
    );

    // Once the file is moved away the same call finishes the worktree and closes the card.
    fs::remove_file(w.project.join("notes/x.md")).unwrap();
    let got = result(&w.land(&card, &cand));
    assert_eq!(got["worktree_follow"]["status"], json!("finished"), "{got}");
    assert_eq!(
        fs::read_to_string(w.project.join("notes/x.md")).unwrap(),
        "lane\n"
    );
    assert_eq!(w.state(&card), "done");
    assert_eq!(landings_in_reflog(&w), 1);
}

/// B2: a tracked folder the head turns into a file is no obstacle: the same move deletes what
/// stands there.
#[test]
fn a_tracked_folder_the_candidate_turns_into_a_file_follows() {
    let w = World::new();
    let (card, cand) = w.ready_editing(&["d/x"], &[("d", "now a file\n")]);
    let got = result(&w.land(&card, &cand));
    assert_eq!(got["worktree_follow"]["status"], json!("followed"), "{got}");
    assert_eq!(
        fs::read_to_string(w.project.join("d")).unwrap(),
        "now a file\n"
    );
    assert_eq!(w.tracked_changes(), "");
    assert_eq!(w.state(&card), "done");
}

/// A folder the head turns into a file refuses when it also holds a file no commit tracks.
#[test]
fn a_folder_with_a_stranger_in_it_is_not_turned_into_a_file() {
    let w = World::new();
    let (card, cand) = w.ready_editing(&["d/x"], &[("d", "now a file\n")]);
    fs::write(w.project.join("d/mine.txt"), "mine\n").unwrap();
    let before = w.before();
    let out = w.land(&card, &cand);
    w.refuses(
        &out,
        &["no commit tracks", "d/mine.txt"],
        &before,
        "a folder with an untracked file in it",
    );
}

/// B2: a case-only rename (`b.txt` to `B.txt`) is the same file under another spelling on a file
/// system that ignores case (this Mac); on one that does not, it is a plain delete and add.
#[test]
fn a_case_only_rename_follows() {
    let w = World::new();
    let (card, cand) = w.ready_editing(&["b.txt"], &[("B.txt", "b\n")]);
    let got = result(&w.land(&card, &cand));
    assert_eq!(got["worktree_follow"]["status"], json!("followed"), "{got}");
    assert_eq!(w.tracked_changes(), "");
    let names = git(&w.project, &["ls-files"]);
    assert!(
        names.lines().any(|n| n == "B.txt") && !names.lines().any(|n| n == "b.txt"),
        "{names}"
    );
    assert_eq!(w.state(&card), "done");
}
