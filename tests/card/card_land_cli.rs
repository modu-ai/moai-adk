//! `moai card land` (SPEC-20261121-card-land): the leader moves the integration branch to a
//! candidate commit by one compare-and-swap and closes the card in the one transaction
//! `card finish` uses. Every test runs on a temporary home and a temporary repository whose
//! integration branch is `integ`; nothing here touches the real ledger.
//!
//! The repository: `main` holds two commits; `head` is the second (the card's change), `integ`
//! points at `tip`, a child of `head`, and a candidate is a commit made with `commit-tree` so the
//! working tree never changes.

use crate::common;

use common::home::TempHome;
use moai_cli::ledger;
use rusqlite::Connection;
use serde_json::{json, Value};
use std::cell::RefCell;
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Output, Stdio};

const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
const S_A: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
const UNKNOWN_HEAD: &str = "fedcba9876543210fedcba9876543210fedcba98";

/// The profile of an m-flow project whose integration branch is `integ`.
const MFLOW_PROFILE: &[u8] = b"{\"schema_version\": 1, \"project_id\": \"demo\", \"purpose\": \"fixture\", \"tools\": [\"rust\"], \"development_methods\": [\"sdd\"], \"hosts\": [\"claude\"], \"work_flow\": {\"flow\": \"m-flow\", \"integration_branch\": \"integ\"}}";
/// The same profile with an integration branch that names a tag path.
const TAG_PROFILE: &[u8] = b"{\"schema_version\": 1, \"project_id\": \"demo\", \"purpose\": \"fixture\", \"tools\": [\"rust\"], \"development_methods\": [\"sdd\"], \"hosts\": [\"claude\"], \"work_flow\": {\"flow\": \"m-flow\", \"integration_branch\": \"refs/tags/v1\"}}";

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
    // A message with no session trailer is signed by the lane session, so the authors of a head
    // built from these commits are known (SPEC-20261124-land-closes-card: a closing landing
    // relies on an independent PASS, and that check needs the authors).
    let signed_message;
    let message = if message.contains("Moai-Session:") {
        message
    } else {
        signed_message = format!("{message}\n\nMoai-Session: {S_A}:{WORKTREE_KEY}");
        signed_message.as_str()
    };
    args.push("-m");
    args.push(message);
    git(project, &args)
}

/// A commit on top of `parent` that adds `path` with `content`; no ref moves.
fn put_on(project: &Path, parent: &str, path: &str, content: &str) -> String {
    put_on_msg(
        project,
        parent,
        path,
        content,
        &format!("docs: add a file\n\nMoai-Session: {S_A}:{WORKTREE_KEY}"),
    )
}

/// [`put_on`] with the commit message given.
fn put_on_msg(project: &Path, parent: &str, path: &str, content: &str, message: &str) -> String {
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

/// The card number of a name like `C-12`.
fn number(name: &str) -> i64 {
    name.trim_start_matches("C-").parse().unwrap()
}

struct World {
    home: TempHome,
    dir: tempfile::TempDir,
    project: PathBuf,
    p: String,
    git_bin: RefCell<PathBuf>,
    /// Variables the next `moai` children get on top of the clean environment.
    extra_env: RefCell<Vec<(String, String)>>,
    profile: RefCell<Vec<u8>>,
    head: String,
    tip: String,
}

impl World {
    fn new() -> Self {
        let home = TempHome::new();
        let project = home.project("demo");
        ledger::register(home.path(), "demo").unwrap();
        git(&project, &["init", "-q", "-b", "main"]);
        fs::write(project.join("a.txt"), "a\n").unwrap();
        git(&project, &["add", "a.txt"]);
        git(&project, &["commit", "-q", "-m", "init"]);
        fs::write(project.join("b.txt"), "b\n").unwrap();
        git(&project, &["add", "b.txt"]);
        // Signed by the lane session, so a PASS of an independent reviewer can be relied on for
        // a head that is already inside the tip (SPEC-20261124-land-closes-card).
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
        let tip = commit_with(&project, &head, &[&head], "integration moved on");
        git(&project, &["update-ref", "refs/heads/integ", &tip]);
        World {
            p: project.to_str().unwrap().to_string(),
            home,
            dir: tempfile::Builder::new()
                .prefix("moai-test-land-")
                .tempdir()
                .unwrap(),
            project,
            git_bin: RefCell::new(git_executable()),
            extra_env: RefCell::new(Vec::new()),
            profile: RefCell::new(MFLOW_PROFILE.to_vec()),
            head,
            tip,
        }
    }

    fn set_profile(&self, bytes: &[u8]) {
        *self.profile.borrow_mut() = bytes.to_vec();
    }

    fn write_profile(&self) {
        let cfg = self.project.join(".moai-cli/project.yaml");
        fs::create_dir_all(cfg.parent().unwrap()).unwrap();
        fs::write(&cfg, &*self.profile.borrow()).unwrap();
    }

    /// `moai <args>` as `session`, with the git program named.
    fn run(&self, session: Option<&str>, args: &[&str]) -> Output {
        let mut cmd: Command = self.home.command(args);
        cmd.env("MOAI_GIT_BIN", &*self.git_bin.borrow());
        for (key, value) in self.extra_env.borrow().iter() {
            cmd.env(key, value);
        }
        if let Some(id) = session {
            cmd.env(SESSION_VAR, id);
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

    /// `moai card land <card> <extra>` as the leader (no session of its own).
    fn land(&self, card: &str, extra: &[&str]) -> Output {
        let mut all = vec!["land", card];
        all.extend_from_slice(extra);
        self.card(None, &all)
    }

    /// A card of `kind` `task` with one write scope and one open criterion per text, taken by
    /// `S_A` and sent to review (the reservation stays held with a live lease); returns its name.
    fn review_card(&self, criteria: &[&str]) -> String {
        let mut args = vec![
            "add",
            "--title",
            "fixture card",
            "--kind",
            "task",
            "--scope",
            "apps/a/",
        ];
        for text in criteria {
            args.push("--ac");
            args.push(text);
        }
        let added = result(&self.card(None, &args));
        let name = added["name"].as_str().unwrap().to_string();
        result(&self.card(None, &["ready", &name]));
        result(&self.card(Some(S_A), &["take", &name]));
        result(&self.card(Some(S_A), &["review", &name, "--generation", "1"]));
        name
    }

    /// A candidate: a merge commit with `tip` as its first parent and a lane commit as its second
    /// (no change to the tree), the shape the leader lands.
    fn candidate(&self) -> String {
        let lane = commit_with(
            &self.project,
            &self.head,
            &[&self.head],
            &format!("the lane head\n\nMoai-Session: {S_A}:{WORKTREE_KEY}"),
        );
        commit_with(
            &self.project,
            &self.tip,
            &[&self.tip, &lane],
            &format!("merge the lane head on the tip\n\nMoai-Session: {S_LEAD}:{WORKTREE_KEY}"),
        )
    }

    /// The commit `integ` points at now.
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

    fn text(&self, sql: &str) -> String {
        let l = ledger::open_read_only(self.home.path(), "demo").unwrap();
        l.conn().query_row(sql, [], |r| r.get(0)).unwrap()
    }

    fn events(&self, card: &str, kind: &str) -> i64 {
        self.count(&format!(
            "SELECT count(*) FROM event WHERE card_id = {} AND kind = '{kind}'",
            number(card)
        ))
    }

    /// Asserts a refusal that wrote nothing: exit 1, `ok` false, the error kind, a reason that
    /// holds one of `any_of`, every table equal to `before`, `integ` still at `integ_before` and
    /// no `fence_reject` event. Returns the message.
    fn refuse(
        &self,
        out: &Output,
        kind: &str,
        any_of: &[&str],
        before: &[String],
        integ_before: &str,
    ) -> String {
        assert_eq!(
            out.status.code(),
            Some(1),
            "{} {}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        );
        let v = envelope(out);
        assert_eq!(v["ok"], json!(false), "{v}");
        assert_eq!(v["error"]["kind"], json!(kind), "{v}");
        let message = v["error"]["message"].as_str().unwrap().to_string();
        assert!(
            any_of.iter().any(|needle| message.contains(needle)),
            "the reason names none of {any_of:?}: {message}"
        );
        assert_eq!(&self.dump(), before, "the refused call wrote to the ledger");
        assert_eq!(self.integ(), integ_before, "the refused call moved integ");
        assert_eq!(
            self.count("SELECT count(*) FROM event WHERE kind = 'fence_reject'"),
            0,
            "the call left a fence_reject event"
        );
        message
    }

    /// The common git folder of the fixture repository.
    fn common(&self) -> PathBuf {
        self.project.join(".git")
    }

    /// The guard's hook, its record folder and the moai the hook finds in `$HOME/.moai/bin`,
    /// and a path for the git calls of the command that holds the tools the hook needs.
    fn install_guard(&self) {
        let hooks = self.common().join("hooks");
        fs::create_dir_all(&hooks).unwrap();
        common::write_executable(
            &hooks.join("reference-transaction"),
            moai_cli::card::guard::hook_text(),
        );
        let bin = self.home.path().join(".moai/bin");
        fs::create_dir_all(&bin).unwrap();
        link_moai(&bin);
        let git_dir = git_executable().parent().unwrap().to_path_buf();
        let path =
            std::env::join_paths([git_dir, PathBuf::from("/usr/bin"), PathBuf::from("/bin")])
                .unwrap();
        self.extra_env
            .borrow_mut()
            .push(("PATH".to_string(), path.to_string_lossy().into_owned()));
    }

    /// Moves the integration branch to `commit` the way `git branch -C` does: no reference
    /// transaction, so no hook sees it.
    fn copy_over(&self, commit: &str) {
        git(&self.project, &["branch", "-f", "copy-source", commit]);
        git(&self.project, &["branch", "-C", "copy-source", "integ"]);
        git(&self.project, &["branch", "-D", "copy-source"]);
        assert_eq!(self.integ(), commit);
    }

    /// A git wrapper that runs `before` (shell text) just before the real `update-ref` and
    /// `after` just after it, then returns the real exit code.
    fn wrap_update_ref(&self, name: &str, before: &str, after: &str) {
        let wrapper = self.dir.path().join(name);
        common::write_executable(
            &wrapper,
            &format!(
                "#!/bin/sh\nfor a in \"$@\"; do\n  if [ \"$a\" = update-ref ]; then\n    {before}\n    '{git}' \"$@\"\n    rc=$?\n    {after}\n    exit $rc\n  fi\ndone\nexec '{git}' \"$@\"\n",
                git = git_executable().display(),
            ),
        );
        *self.git_bin.borrow_mut() = wrapper;
    }

    /// A card of kind `task` that is ready and held by nobody; returns its name.
    fn ready_card(&self, criteria: &[&str]) -> String {
        let mut args = vec![
            "add",
            "--title",
            "ready card",
            "--kind",
            "task",
            "--scope",
            "apps/a/",
        ];
        for text in criteria {
            args.push("--ac");
            args.push(text);
        }
        let added = result(&self.card(None, &args));
        let name = added["name"].as_str().unwrap().to_string();
        result(&self.card(None, &["ready", &name]));
        name
    }

    /// Asserts a refusal of the move itself: exit 1, `ok` false, the error kind, a reason that
    /// holds one of `any_of`, `integ` still at `integ_before`, every table but `event` equal to
    /// `before` and exactly one more event than before, the start row of `card` (SPEC-20261121-
    /// ref-guard: the row is written before the move, so a refused update leaves it). Returns the
    /// message.
    fn refuse_the_move(
        &self,
        out: &Output,
        kind: &str,
        any_of: &[&str],
        before: &[String],
        integ_before: &str,
        card: &str,
    ) -> String {
        assert_eq!(
            out.status.code(),
            Some(1),
            "{}",
            String::from_utf8_lossy(&out.stdout)
        );
        let v = envelope(out);
        assert_eq!(v["ok"], json!(false), "{v}");
        assert_eq!(v["error"]["kind"], json!(kind), "{v}");
        let message = v["error"]["message"].as_str().unwrap().to_string();
        assert!(
            any_of.iter().any(|needle| message.contains(needle)),
            "the reason names none of {any_of:?}: {message}"
        );
        assert!(
            !message.contains("nothing was changed") && message.contains("start row"),
            "a refusal after the start row says what was written: {message}"
        );
        assert_eq!(self.integ(), integ_before, "the refused call moved integ");
        let split = |d: &[String]| {
            let rest: Vec<String> = d
                .iter()
                .filter(|l| !l.starts_with("event:"))
                .cloned()
                .collect();
            (rest, d.len() - rest_len(d))
        };
        fn rest_len(d: &[String]) -> usize {
            d.iter().filter(|l| !l.starts_with("event:")).count()
        }
        let (after_rest, after_events) = split(&self.dump());
        let (before_rest, before_events) = split(before);
        assert_eq!(
            after_rest, before_rest,
            "the refused move wrote beyond the start row"
        );
        assert_eq!(after_events, before_events + 1, "expected one new event");
        assert_eq!(
            self.events(card, "card.land_started"),
            1,
            "the start row is missing"
        );
        assert_eq!(self.events(card, "card.landed"), 0);
        message
    }

    /// A draft SPEC with every placeholder filled, so `spec check` passes it; returns its id.
    fn draft_spec(&self, slug: &str) -> String {
        let out = self.run(
            None,
            &[
                "spec",
                "new",
                slug,
                "--title",
                "t",
                "--minimal",
                "--date",
                "2026-11-21",
                "--project",
                &self.p,
            ],
        );
        let id = result(&out)["id"].as_str().unwrap().to_string();
        let path = self.project.join(".moai/specs").join(&id).join("spec.md");
        fs::write(&path, filled(&fs::read_to_string(&path).unwrap())).unwrap();
        id
    }
}

// ---- AC-1: the move and the close in one call ----------------------------------------------

/// AC-1: a card in review whose candidate is a commit on the tip lands in one call.
#[test]
fn ac1_a_candidate_on_the_tip_moves_the_branch_and_closes_the_card() {
    let w = World::new();
    let card = w.review_card(&["one"]);
    let candidate = w.candidate();
    let out = w.land_pass(
        &card,
        &[
            "--head",
            &candidate,
            "--met",
            "1=the test",
            "--by",
            "leader",
        ],
    );
    let got = result(&out);
    assert_eq!(got["action"], json!("land"), "{got}");
    assert_eq!(got["moved"], json!(true), "{got}");
    assert_eq!(got["branch"], json!("integ"), "{got}");
    assert_eq!(got["from_tip"], json!(w.tip), "{got}");
    assert_eq!(got["to_tip"], json!(candidate), "{got}");
    assert_eq!(got["closed"], json!(true), "{got}");
    assert_eq!(got["state"], json!("done"), "{got}");
    assert_eq!(w.integ(), candidate);
    let id = number(&card);
    assert_eq!(
        w.text(&format!("SELECT state FROM card WHERE card_id = {id}")),
        "done"
    );
    assert_eq!(
        w.text(&format!(
            "SELECT status FROM reservation WHERE card_id = {id} ORDER BY generation DESC LIMIT 1"
        )),
        "released"
    );
    assert_eq!(w.events(&card, "card.landed"), 1);
    assert_eq!(w.events(&card, "card.finish"), 1);
    // The reference log names the candidate and no card number.
    let log = git(
        &w.project,
        &[
            "reflog",
            "show",
            "--format=%gs",
            "-n",
            "1",
            "refs/heads/integ",
        ],
    );
    assert!(log.contains(&candidate[..12]), "{log}");
    assert!(!log.contains("C-"), "{log}");
}

// ---- AC-2: a branch that moved under the call ------------------------------------------------

/// AC-2: another process moves the branch between the read of the tip and the update, so the
/// compare-and-swap refuses and everything stays as it was.
#[test]
fn ac2_a_branch_that_moved_under_the_call_refuses_with_tip_moved() {
    let w = World::new();
    let card = w.review_card(&["one"]);
    let candidate = w.candidate();
    let other = commit_with(&w.project, &w.tip, &[&w.tip], "somebody else landed first");
    let real = git_executable();
    let wrapper = w.dir.path().join("git-moves-integ");
    common::write_executable(
        &wrapper,
        &format!(
            "#!/bin/sh\nfor a in \"$@\"; do\n  if [ \"$a\" = update-ref ]; then\n    '{git}' -C '{project}' update-ref refs/heads/integ {other}\n  fi\ndone\nexec '{git}' \"$@\"\n",
            git = real.display(),
            project = w.project.display(),
            other = other,
        ),
    );
    *w.git_bin.borrow_mut() = wrapper;
    w.pass_of(&card, &candidate);
    let before = w.dump();
    let out = w.land_lead(
        &card,
        &[
            "--head",
            &candidate,
            "--met",
            "1=the test",
            "--by",
            "leader",
        ],
    );
    w.refuse_the_move(&out, "conflict", &["tip_moved"], &before, &other, &card);
}

// ---- AC-3: a candidate already inside the tip ------------------------------------------------

/// AC-3: a candidate that is already inside the tip closes the card with no reference update; a
/// second call answers done and writes nothing, and another head refuses.
#[test]
fn ac3_a_candidate_inside_the_tip_closes_the_card_without_a_move() {
    let w = World::new();
    let card = w.review_card(&["one"]);
    let out = w.land_pass(
        &card,
        &["--head", &w.head, "--met", "1=the test", "--by", "leader"],
    );
    let got = result(&out);
    assert_eq!(got["moved"], json!(false), "{got}");
    assert_eq!(got["closed"], json!(true), "{got}");
    assert_eq!(got["to_tip"], json!(w.tip), "{got}");
    assert_eq!(w.integ(), w.tip);
    // The same call again: done, nothing written.
    let before = w.dump();
    let again = result(&w.land_lead(
        &card,
        &["--head", &w.head, "--met", "1=the test", "--by", "leader"],
    ));
    assert_eq!(again["state"], json!("done"), "{again}");
    assert_eq!(again["moved"], json!(false), "{again}");
    assert_eq!(w.dump(), before, "a call on a done card wrote");
    // Another head on the done card refuses.
    let out = w.land_lead(
        &card,
        &["--head", &w.tip, "--met", "1=the test", "--by", "leader"],
    );
    w.refuse(&out, "invalid_input", &["is done"], &before, &w.tip);
}

// ---- AC-4: stale base --------------------------------------------------------------------------

/// AC-4: a candidate that does not have the tip on its first-parent line refuses with stale_base.
#[test]
fn ac4_a_candidate_without_the_tip_on_its_first_parent_line_is_stale() {
    let w = World::new();
    let card = w.review_card(&["one"]);
    let before = w.dump();
    // Branched from an older commit: the tip is not even an ancestor.
    let older = commit_with(&w.project, &w.head, &[&w.head], "from an older tip");
    let out = w.land(
        &card,
        &["--head", &older, "--met", "1=the test", "--by", "leader"],
    );
    let message = w.refuse(&out, "conflict", &["stale_base"], &before, &w.tip);
    assert!(
        message.contains(&w.tip) && message.contains(&older),
        "{message}"
    );
    // The tip only on a second parent: its first-parent line is another history.
    let second = commit_with(&w.project, &w.head, &[&w.head, &w.tip], "the tip merged in");
    let out = w.land(
        &card,
        &["--head", &second, "--met", "1=the test", "--by", "leader"],
    );
    w.refuse(&out, "conflict", &["stale_base"], &before, &w.tip);
}

// ---- AC-5: every check of finish comes before the move --------------------------------------

/// AC-5: what `card finish` would refuse is refused before the branch moves.
#[test]
fn ac5_every_refusal_of_finish_comes_before_the_move() {
    let w = World::new();
    let candidate = w.candidate();
    let before_ref = w.tip.clone();

    // A wrong generation.
    let card = w.review_card(&["one"]);
    let before = w.dump();
    let out = w.land(
        &card,
        &[
            "--head",
            &candidate,
            "--met",
            "1=x",
            "--generation",
            "7",
            "--by",
            "leader",
        ],
    );
    let message = w.refuse(&out, "invalid_input", &["generation"], &before, &before_ref);
    assert_eq!(
        message.matches("nothing was changed").count(),
        1,
        "the refusal says it twice: {message}"
    );

    // A met option for a criterion that is not open, and one named twice.
    let out = w.land(
        &card,
        &["--head", &candidate, "--met", "9=x", "--by", "leader"],
    );
    w.refuse(
        &out,
        "invalid_input",
        &["not an open criterion"],
        &before,
        &before_ref,
    );
    let out = w.land(
        &card,
        &[
            "--head", &candidate, "--met", "1=a", "--met", "1=b", "--by", "leader",
        ],
    );
    w.refuse(
        &out,
        "invalid_input",
        &["named twice"],
        &before,
        &before_ref,
    );

    // A card in a state that cannot be finished: a draft.
    let added = result(&w.card(
        None,
        &[
            "add", "--title", "a draft", "--kind", "task", "--scope", "apps/b/", "--ac", "one",
        ],
    ));
    let draft = added["name"].as_str().unwrap().to_string();
    let before = w.dump();
    let out = w.land(
        &draft,
        &["--head", &candidate, "--met", "1=x", "--by", "leader"],
    );
    w.refuse(&out, "conflict", &["draft"], &before, &before_ref);

    // A SPEC with errors at the candidate: the checkout copy is good, the candidate's is not.
    let spec = w.draft_spec("land-fixture");
    let spec_path = format!(".moai/specs/{spec}/spec.md");
    let good = fs::read_to_string(w.project.join(&spec_path)).unwrap();
    let card = w.review_card(&["one"]);
    result(&w.card(None, &["spec", &card, &spec]));
    let broken_tip = put_on(&w.project, &w.tip, &spec_path, "not a spec\n");
    git(&w.project, &["update-ref", "refs/heads/integ", &broken_tip]);
    let candidate = commit_with(&w.project, &broken_tip, &[&broken_tip], "on a broken SPEC");
    assert_ne!(good, "not a spec\n");
    let before = w.dump();
    let out = w.land(
        &card,
        &["--head", &candidate, "--met", "1=x", "--by", "leader"],
    );
    w.refuse(
        &out,
        "invalid_input",
        &["error(s)", "SPEC"],
        &before,
        &broken_tip,
    );
}

// ---- AC-6: a checked-out branch and a head that is no commit ----------------------------------

/// AC-6: the integration branch checked out in another worktree refuses and names it, and a
/// candidate that is no commit refuses; both before any write.
#[test]
fn ac6_a_checked_out_branch_and_a_head_that_is_no_commit_refuse_before_the_move() {
    let w = World::new();
    let card = w.review_card(&["one"]);
    let candidate = w.candidate();
    let before = w.dump();
    let out = w.land(
        &card,
        &["--head", UNKNOWN_HEAD, "--met", "1=x", "--by", "leader"],
    );
    w.refuse(&out, "invalid_input", &["not a commit"], &before, &w.tip);

    let tree = w.dir.path().join("integ-checkout");
    git(
        &w.project,
        &["worktree", "add", "-q", tree.to_str().unwrap(), "integ"],
    );
    // A clean checkout is moved with the branch (SPEC-20261125-landing-moves-checkout); one with a
    // change to a tracked file still refuses and is named.
    fs::write(tree.join("a.txt"), "changed\n").unwrap();
    let out = w.land(
        &card,
        &["--head", &candidate, "--met", "1=x", "--by", "leader"],
    );
    let message = w.refuse(&out, "conflict", &["checked out"], &before, &w.tip);
    assert!(
        message.contains("integ-checkout"),
        "the refusal names the worktree: {message}"
    );
}

// ---- AC-7: the caller ---------------------------------------------------------------------------

/// AC-7: a caller word other than the leader word refuses.
#[test]
fn ac7_a_caller_other_than_the_leader_refuses() {
    let w = World::new();
    let card = w.review_card(&["one"]);
    let candidate = w.candidate();
    let before = w.dump();
    let out = w.land(
        &card,
        &["--head", &candidate, "--met", "1=x", "--by", "agent-16"],
    );
    let message = w.refuse(&out, "invalid_input", &["leader"], &before, &w.tip);
    assert!(message.contains("typed verdict"), "{message}");
    // The holder's own session is no leader either.
    w.write_profile();
    let out = w.card(
        Some(S_A),
        &[
            "land", &card, "--head", &candidate, "--met", "1=x", "--by", "agent-1",
        ],
    );
    w.refuse(&out, "invalid_input", &["leader"], &before, &w.tip);
}

// ---- AC-8: a landing that closes nothing ---------------------------------------------------------

/// AC-8: an open criterion with no met option moves the branch, records one landing event and
/// leaves the card open; the same call with the met options closes it with no second move.
#[test]
fn ac8_an_open_criterion_with_no_met_option_lands_and_leaves_the_card_open() {
    let w = World::new();
    let card = w.review_card(&["one", "two"]);
    let candidate = w.candidate();
    // A slice is a `--keep-open` landing now: a landing that names no `--met` for a criterion
    // fills it from the PASS it relies on (SPEC-20261124-land-closes-card).
    let got = result(&w.land_lead(
        &card,
        &["--head", &candidate, "--keep-open", "--by", "leader"],
    ));
    assert_eq!(got["moved"], json!(true), "{got}");
    assert_eq!(got["closed"], json!(false), "{got}");
    assert_eq!(got["open_criteria"], json!([1, 2]), "{got}");
    assert_eq!(w.integ(), candidate);
    assert_eq!(w.events(&card, "card.landed"), 1);
    assert_eq!(w.events(&card, "card.finish"), 0);
    assert_eq!(
        w.text(&format!(
            "SELECT state FROM card WHERE card_id = {}",
            number(&card)
        )),
        "review"
    );
    let done = result(&w.land_pass(
        &card,
        &[
            "--head", &candidate, "--met", "1=x", "--met", "2=y", "--by", "leader",
        ],
    ));
    assert_eq!(done["moved"], json!(false), "{done}");
    assert_eq!(done["closed"], json!(true), "{done}");
    assert_eq!(w.events(&card, "card.landed"), 1);
    assert_eq!(w.events(&card, "card.finish"), 1);
}

// ---- AC-9: a rerun after a stop -------------------------------------------------------------------

/// AC-9: the branch was moved by hand to the candidate and the card is still open; one call
/// closes it with one landing event and one finish event.
#[test]
fn ac9_a_rerun_after_a_stop_between_the_move_and_the_close_closes_the_card() {
    let w = World::new();
    let card = w.review_card(&["one"]);
    let candidate = w.candidate();
    git(&w.project, &["update-ref", "refs/heads/integ", &candidate]);
    let got = result(&w.land_pass(
        &card,
        &["--head", &candidate, "--met", "1=x", "--by", "leader"],
    ));
    assert_eq!(got["moved"], json!(false), "{got}");
    assert_eq!(got["closed"], json!(true), "{got}");
    assert_eq!(w.integ(), candidate);
    assert_eq!(w.events(&card, "card.landed"), 1);
    assert_eq!(w.events(&card, "card.finish"), 1);
}

// ---- AC-10: the profile branch and git ---------------------------------------------------------------

/// AC-10: a profile whose integration branch names a tag path refuses with the words "is not a
/// branch name", and a git that cannot run refuses before any write.
#[test]
fn ac10_a_tag_path_and_a_git_that_cannot_run_refuse_before_any_write() {
    let w = World::new();
    let card = w.review_card(&["one"]);
    let candidate = w.candidate();
    let before = w.dump();
    w.set_profile(TAG_PROFILE);
    let out = w.land(
        &card,
        &["--head", &candidate, "--met", "1=x", "--by", "leader"],
    );
    w.refuse(
        &out,
        "invalid_input",
        &["is not a branch name"],
        &before,
        &w.tip,
    );

    w.set_profile(MFLOW_PROFILE);
    *w.git_bin.borrow_mut() = w.dir.path().join("no-such-git");
    let out = w.land(
        &card,
        &["--head", &candidate, "--met", "1=x", "--by", "leader"],
    );
    w.refuse(
        &out,
        "invalid_input",
        &["git could not run"],
        &before,
        &w.tip,
    );
}

// ---- the repairs of the audit of the first slice ---------------------------------------------

/// B-1, G: a close refused after the branch moved says the branch moved and never that nothing was
/// changed; the same call run again closes the card. A call that gets in between is made by the
/// git wrapper: after the real `update-ref` it marks criterion 1 met as the holder, so the close
/// names a criterion that is no longer open.
#[test]
fn ac13_a_close_refused_after_the_move_says_the_branch_moved() {
    let w = World::new();
    let card = w.review_card(&["one"]);
    let candidate = w.candidate();
    let real = git_executable();
    let wrapper = w.dir.path().join("git-races-the-ledger");
    common::write_executable(
        &wrapper,
        &format!(
            "#!/bin/sh\nfor a in \"$@\"; do\n  if [ \"$a\" = update-ref ]; then\n    '{git}' \"$@\"\n    rc=$?\n    CLAUDE_CODE_SESSION_ID={holder} HOME='{home}' '{moai}' card criterion {card} 1 --generation 1 --met --evidence raced --project '{project}' --home '{home}' >/dev/null 2>&1\n    exit $rc\n  fi\ndone\nexec '{git}' \"$@\"\n",
            git = real.display(),
            holder = S_A,
            home = w.home.path().display(),
            moai = common::MOAI_BIN,
            card = card,
            project = w.project.display(),
        ),
    );
    *w.git_bin.borrow_mut() = wrapper;
    let out = w.land_pass(
        &card,
        &[
            "--head",
            &candidate,
            "--met",
            "1=the test",
            "--by",
            "leader",
        ],
    );
    assert_eq!(
        out.status.code(),
        Some(1),
        "{}",
        String::from_utf8_lossy(&out.stdout)
    );
    let v = envelope(&out);
    assert_eq!(v["ok"], json!(false), "{v}");
    let message = v["error"]["message"].as_str().unwrap().to_string();
    assert!(message.contains("not an open criterion"), "{message}");
    assert!(
        message.contains(&format!("already moved to {candidate}")),
        "the refusal does not say the branch moved: {message}"
    );
    assert!(
        !message.contains("nothing was changed"),
        "the refusal says nothing changed after the branch moved: {message}"
    );
    assert_eq!(w.integ(), candidate, "the branch moved");
    assert_eq!(w.events(&card, "card.landed"), 0);
    // The same call again, with git running as it should: no move, the card closes.
    *w.git_bin.borrow_mut() = git_executable();
    let again = result(&w.land_pass(&card, &["--head", &candidate, "--by", "leader"]));
    assert_eq!(again["moved"], json!(false), "{again}");
    assert_eq!(again["closed"], json!(true), "{again}");
}

/// D: the SPEC is read at the candidate and not at the old tip: a good SPEC on the tip and a
/// broken one at the candidate refuses, and a broken one on the tip with a good one at the
/// candidate lands.
#[test]
fn ac14_the_spec_is_read_at_the_candidate_and_not_at_the_old_tip() {
    let w = World::new();
    let spec = w.draft_spec("land-candidate");
    let spec_path = format!(".moai/specs/{spec}/spec.md");
    let good = fs::read_to_string(w.project.join(&spec_path)).unwrap();
    let card = w.review_card(&["one"]);
    result(&w.card(None, &["spec", &card, &spec]));
    let good_tip = put_on(&w.project, &w.tip, &spec_path, &good);
    git(&w.project, &["update-ref", "refs/heads/integ", &good_tip]);
    let broken_candidate = put_on(&w.project, &good_tip, &spec_path, "not a spec\n");
    w.pass_of(&card, &broken_candidate);
    let before = w.dump();
    let out = w.land_lead(
        &card,
        &[
            "--head",
            &broken_candidate,
            "--met",
            "1=x",
            "--by",
            "leader",
        ],
    );
    w.refuse(&out, "invalid_input", &["error(s)"], &before, &good_tip);
    // The other way round: the old tip is broken, the candidate is good.
    let broken_tip = put_on(&w.project, &w.tip, &spec_path, "not a spec\n");
    git(&w.project, &["update-ref", "refs/heads/integ", &broken_tip]);
    let good_candidate = put_on(&w.project, &broken_tip, &spec_path, &good);
    let got = result(&w.land_pass(
        &card,
        &["--head", &good_candidate, "--met", "1=x", "--by", "leader"],
    ));
    assert_eq!(got["closed"], json!(true), "{got}");
    assert_eq!(w.integ(), good_candidate);
}

/// F: the landing event records the head and the new tip (the candidate after a move, the tip
/// that was read when the head was already inside).
#[test]
fn ac15_the_landing_event_records_the_head_and_the_new_tip() {
    let w = World::new();
    let moved_card = w.review_card(&["one"]);
    let candidate = w.candidate();
    result(&w.land_pass(
        &moved_card,
        &["--head", &candidate, "--met", "1=x", "--by", "leader"],
    ));
    let payload = |card: &str| -> Value {
        let text = w.text(&format!(
            "SELECT payload FROM event WHERE card_id = {} AND kind = 'card.landed' \
             ORDER BY event_id DESC LIMIT 1",
            number(card)
        ));
        serde_json::from_str(&text).unwrap()
    };
    let landed = payload(&moved_card);
    assert_eq!(landed["head"], json!(candidate), "{landed}");
    assert_eq!(landed["tip"], json!(candidate), "{landed}");
    assert_eq!(landed["by"], json!("leader"), "{landed}");
    // A head already inside the tip: the tip is the one that was read, not the head.
    let inside_card = w.review_card(&["one"]);
    result(&w.land_pass(
        &inside_card,
        &["--head", &w.head, "--met", "1=x", "--by", "leader"],
    ));
    let landed = payload(&inside_card);
    assert_eq!(landed["head"], json!(w.head), "{landed}");
    assert_eq!(landed["tip"], json!(candidate), "{landed}");
}

/// N: a card that closes on a report is `card finish`'s, and nothing moves.
#[test]
fn ac16_a_research_card_is_refused_and_nothing_moves() {
    let w = World::new();
    let added = result(&w.card(
        None,
        &[
            "add",
            "--title",
            "a question",
            "--kind",
            "research",
            "--scope",
            "apps/r/",
            "--ac",
            "one",
        ],
    ));
    let card = added["name"].as_str().unwrap().to_string();
    result(&w.card(None, &["ready", &card]));
    let candidate = w.candidate();
    let before = w.dump();
    let out = w.land(
        &card,
        &["--head", &candidate, "--met", "1=x", "--by", "leader"],
    );
    let message = w.refuse(
        &out,
        "invalid_input",
        &["closes on a report"],
        &before,
        &w.tip,
    );
    assert!(message.contains("card finish"), "{message}");
}

/// H: a hostile git environment of the caller changes nothing: the command runs git through the
/// scrub, so a repository named by `GIT_DIR` and the like is never the one it works in.
#[test]
fn ac17_a_hostile_git_environment_is_ignored() {
    let w = World::new();
    let card = w.review_card(&["one"]);
    let candidate = w.candidate();
    let nowhere = w.dir.path().join("nowhere");
    for (key, value) in [
        ("GIT_DIR", nowhere.join("git-dir")),
        ("GIT_WORK_TREE", nowhere.join("work-tree")),
        ("GIT_INDEX_FILE", nowhere.join("index")),
    ] {
        w.extra_env
            .borrow_mut()
            .push((key.to_string(), value.to_str().unwrap().to_string()));
    }
    w.extra_env
        .borrow_mut()
        .push(("GIT_NAMESPACE".to_string(), "hostile".to_string()));
    let got = result(&w.land_pass(
        &card,
        &["--head", &candidate, "--met", "1=x", "--by", "leader"],
    ));
    assert_eq!(got["moved"], json!(true), "{got}");
    assert_eq!(w.integ(), candidate);
}

/// N-1: a reference update that git refuses for another reason than a moved branch (a stale lock
/// file here) is reported with git's reason, not as `tip_moved`.
#[test]
fn ac18_an_update_ref_failure_that_is_no_move_is_not_called_tip_moved() {
    let w = World::new();
    let card = w.review_card(&["one"]);
    let candidate = w.candidate();
    fs::write(w.project.join(".git/refs/heads/integ.lock"), "").unwrap();
    w.pass_of(&card, &candidate);
    let before = w.dump();
    let out = w.land_lead(
        &card,
        &["--head", &candidate, "--met", "1=x", "--by", "leader"],
    );
    let message = w.refuse_the_move(
        &out,
        "invalid_input",
        &["git refused"],
        &before,
        &w.tip,
        &card,
    );
    assert!(!message.contains("tip_moved"), "{message}");
    assert!(message.contains("lock"), "git's reason is named: {message}");
}

// ---- slice 1b: the guard and the ledger row ------------------------------------------------

/// AC-7 (SPEC-20261121-ref-guard): `card land` under the guard moves the branch, the intent record
/// exists before the update and is gone after, a record left by a stopped call is replaced, and a
/// move the guard refuses answers with the guard's message and moves nothing.
#[test]
fn ac19_the_landing_command_moves_the_branch_under_the_guard() {
    let w = World::new();
    w.install_guard();
    let side = w.dir.path().join("intent-seen");
    let intent = w.common().join("moai/land-intent");
    w.wrap_update_ref(
        "git-copies-intent",
        &format!(
            "[ -f '{i}' ] && cp '{i}' '{s}'",
            i = intent.display(),
            s = side.display()
        ),
        "",
    );
    let card = w.review_card(&["one"]);
    let candidate = w.candidate();
    // A record left by a stopped call that names another move is replaced by this call.
    fs::create_dir_all(w.common().join("moai")).unwrap();
    fs::write(
        &intent,
        "{\"ref\":\"refs/heads/integ\",\"old\":\"1\",\"new\":\"2\"}",
    )
    .unwrap();
    let got = result(&w.land_pass(
        &card,
        &["--head", &candidate, "--met", "1=x", "--by", "leader"],
    ));
    assert_eq!(got["moved"], json!(true), "{got}");
    assert_eq!(w.integ(), candidate);
    let seen: Value = serde_json::from_str(&fs::read_to_string(&side).unwrap()).unwrap();
    assert_eq!(seen["ref"], json!("refs/heads/integ"), "{seen}");
    assert_eq!(seen["old"], json!(w.tip), "{seen}");
    assert_eq!(seen["new"], json!(candidate), "{seen}");
    assert!(!intent.exists(), "the record stays after the call");
    assert_eq!(
        fs::read_to_string(w.common().join("moai/integration-branch")).unwrap(),
        "refs/heads/integ\n",
        "the lander seeds the record from the profile"
    );

    // A move the guard refuses: the record is removed just before the update.
    let second = w.review_card(&["one"]);
    let next = commit_with(&w.project, &candidate, &[&candidate], "the next lane head");
    w.wrap_update_ref(
        "git-loses-intent",
        &format!("rm -f '{}'", intent.display()),
        "",
    );
    let out = w.land_pass(
        &second,
        &["--head", &next, "--met", "1=x", "--by", "leader"],
    );
    assert_eq!(
        out.status.code(),
        Some(1),
        "{}",
        String::from_utf8_lossy(&out.stdout)
    );
    let v = envelope(&out);
    let message = v["error"]["message"].as_str().unwrap().to_string();
    assert!(
        message.contains("moai ref guard"),
        "the guard's message is lost: {message}"
    );
    assert!(!message.contains("tip_moved"), "{message}");
    assert_eq!(w.integ(), candidate, "the guard let the branch move");
    assert_eq!(w.events(&second, "card.landed"), 0);
}

/// AC-8: the start row is written before the landing event and names branch, tip, head and
/// caller; a call that refuses writes no row; a rerun for the same card, head and tip writes no
/// second row; a close refused after the move leaves the row.
#[test]
fn ac20_the_start_row_comes_before_the_move_and_once() {
    let w = World::new();
    // A landing: the row first, the landing event after.
    let card = w.review_card(&["one"]);
    let candidate = w.candidate();
    result(&w.land_pass(
        &card,
        &["--head", &candidate, "--met", "1=x", "--by", "leader"],
    ));
    let row = |card: &str| -> Vec<(i64, Value)> {
        let l = ledger::open_read_only(w.home.path(), "demo").unwrap();
        let mut st = l
            .conn()
            .prepare(
                "SELECT event_id, payload FROM event WHERE card_id = ?1 \
                 AND kind = 'card.land_started' ORDER BY event_id",
            )
            .unwrap();
        st.query_map([number(card)], |r| {
            Ok((
                r.get::<_, i64>(0)?,
                serde_json::from_str(&r.get::<_, String>(1)?).unwrap(),
            ))
        })
        .unwrap()
        .map(Result::unwrap)
        .collect()
    };
    let rows = row(&card);
    assert_eq!(rows.len(), 1, "{rows:?}");
    let landed_id = w.count(&format!(
        "SELECT event_id FROM event WHERE card_id = {} AND kind = 'card.landed'",
        number(&card)
    ));
    assert!(rows[0].0 < landed_id, "the row does not come first");
    assert_eq!(rows[0].1["branch"], json!("integ"), "{:?}", rows[0].1);
    assert_eq!(rows[0].1["tip"], json!(w.tip), "{:?}", rows[0].1);
    assert_eq!(rows[0].1["head"], json!(candidate), "{:?}", rows[0].1);
    assert_eq!(rows[0].1["by"], json!("leader"), "{:?}", rows[0].1);

    // A card that cannot finish refuses before any row.
    let waiting = w.ready_card(&["one"]);
    let next = commit_with(&w.project, &candidate, &[&candidate], "another lane head");
    let out = w.land_pass(
        &waiting,
        &["--head", &next, "--generation", "7", "--by", "leader"],
    );
    assert_eq!(out.status.code(), Some(1));
    assert_eq!(row(&waiting).len(), 0);
    assert_eq!(w.integ(), candidate);

    // A stop between the row and the move (the guard refuses), then the same call: one row.
    let held = w.review_card(&["one"]);
    w.install_guard();
    let intent = w.common().join("moai/land-intent");
    w.wrap_update_ref(
        "git-loses-intent",
        &format!("rm -f '{}'", intent.display()),
        "",
    );
    let out = w.land_pass(&held, &["--head", &next, "--met", "1=x", "--by", "leader"]);
    assert_eq!(out.status.code(), Some(1));
    assert_eq!(row(&held).len(), 1, "the refused move left no row");
    assert_eq!(w.integ(), candidate);
    *w.git_bin.borrow_mut() = git_executable();
    result(&w.land_pass(&held, &["--head", &next, "--met", "1=x", "--by", "leader"]));
    assert_eq!(row(&held).len(), 1, "the rerun wrote a second row");
    assert_eq!(w.integ(), next);

    // A close refused after the move leaves the row: another session marks the criterion met
    // right after the update.
    let raced = w.review_card(&["one"]);
    let after = commit_with(&w.project, &next, &[&next], "the last lane head");
    w.wrap_update_ref(
        "git-races-the-ledger",
        "",
        &format!(
            "CLAUDE_CODE_SESSION_ID={holder} HOME='{home}' '{moai}' card criterion {raced} 1 --generation 1 --met --evidence raced --project '{project}' --home '{home}' >/dev/null 2>&1",
            holder = S_A,
            home = w.home.path().display(),
            moai = common::MOAI_BIN,
            project = w.project.display(),
        ),
    );
    let out = w.land_pass(
        &raced,
        &["--head", &after, "--met", "1=x", "--by", "leader"],
    );
    assert_eq!(
        out.status.code(),
        Some(1),
        "{}",
        String::from_utf8_lossy(&out.stdout)
    );
    assert_eq!(w.integ(), after);
    assert_eq!(row(&raced).len(), 1, "the refused close lost the row");

    // The same call again after the refused close: the branch is already at the head, so there
    // is no move and no second row, and the refusal is the close's again.
    let again = w.land_pass(
        &raced,
        &["--head", &after, "--met", "1=x", "--by", "leader"],
    );
    assert_eq!(again.status.code(), Some(1));
    assert_eq!(w.integ(), after);
    assert_eq!(
        row(&raced).len(),
        1,
        "the rerun after the move wrote a second row"
    );
}

/// AC-11: a lane takes the ready card between the checks and the close while the call carries a
/// generation: the refusal after the move names the cause and the moved branch and does not say
/// that nothing was changed.
#[test]
fn ac21_a_take_between_the_checks_and_the_close_is_named_after_the_move() {
    let w = World::new();
    let card = w.ready_card(&["one"]);
    let candidate = w.candidate();
    w.wrap_update_ref(
        "git-takes-the-card",
        "",
        &format!(
            "CLAUDE_CODE_SESSION_ID={holder} HOME='{home}' '{moai}' card take {card} --project '{project}' --home '{home}' >/dev/null 2>&1",
            holder = S_A,
            home = w.home.path().display(),
            moai = common::MOAI_BIN,
            project = w.project.display(),
        ),
    );
    let out = w.land_pass(
        &card,
        &[
            "--head",
            &candidate,
            "--met",
            "1=x",
            "--generation",
            "0",
            "--by",
            "leader",
        ],
    );
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
        message.contains("does not match"),
        "the cause is not named: {message}"
    );
    assert!(
        message.contains(&format!("already moved to {candidate}")),
        "the moved branch is not named: {message}"
    );
    assert!(
        !message.contains("nothing was changed"),
        "the refusal says nothing changed after the move: {message}"
    );
    assert_eq!(w.integ(), candidate);
}

/// AC-14: the answer names a case variant that already exists, and a profile that names another
/// integration branch than the record is reported while the recorded branch is the one moved.
#[test]
fn ac22_the_answer_reports_a_variant_and_a_disagreeing_profile() {
    let w = World::new();
    // A packed variant of the branch's name, made before the call: the branch is packed first,
    // because on a case-insensitive file system a loose variant is the branch's own file.
    git(&w.project, &["pack-refs", "--all"]);
    git(&w.project, &["update-ref", "refs/heads/INTEG", &w.tip]);
    git(&w.project, &["pack-refs", "--all"]);
    // The record exists and the profile names another branch.
    fs::create_dir_all(w.common().join("moai")).unwrap();
    fs::write(
        w.common().join("moai/integration-branch"),
        "refs/heads/integ\n",
    )
    .unwrap();
    w.set_profile(
        &String::from_utf8_lossy(MFLOW_PROFILE)
            .replace(
                "\"integration_branch\": \"integ\"",
                "\"integration_branch\": \"other\"",
            )
            .into_bytes(),
    );
    let card = w.review_card(&["one"]);
    let candidate = w.candidate();
    let got = result(&w.land_pass(
        &card,
        &["--head", &candidate, "--met", "1=x", "--by", "leader"],
    ));
    assert_eq!(got["moved"], json!(true), "{got}");
    assert_eq!(
        got["branch"],
        json!("integ"),
        "the recorded branch is the one moved: {got}"
    );
    assert_eq!(got["profile_disagrees"], json!("other"), "{got}");
    assert_eq!(got["variants"], json!(["refs/heads/INTEG"]), "{got}");
    assert_eq!(w.integ(), candidate);
}

/// Links the moai under test as `bin/moai`, where the hook looks for it. The hook is a shell
/// script, so the guard tests run on unix only.
fn link_moai(bin: &Path) {
    #[cfg(unix)]
    std::os::unix::fs::symlink(common::MOAI_BIN, bin.join("moai")).unwrap();
    #[cfg(not(unix))]
    {
        let _ = bin;
        panic!("the guard hook is a shell script: unix only");
    }
}

/// AC-23 (SPEC-20261121-ref-guard): a branch that something moved outside the landing command is a
/// drift. `git branch -C` copies a branch over the integration branch with no reference
/// transaction, so no hook sees it; the next `card land` compares the tip with the tip the last
/// landing left (or the head of a move that is under way) and refuses, naming the drift and the
/// command that puts the branch back. Nothing is written, and after the recovery the call lands.
#[test]
fn ac23_a_branch_moved_outside_the_landing_command_is_a_drift() {
    let w = World::new();
    w.install_guard();
    // No landing recorded yet: there is nothing to compare, so a first landing goes through.
    let first = w.review_card(&["one"]);
    let landed = w.candidate();
    result(&w.land_pass(
        &first,
        &["--head", &landed, "--met", "1=x", "--by", "leader"],
    ));
    assert_eq!(w.integ(), landed);

    // A copy over the branch: an unrelated commit that is not on the first-parent line.
    let stray = commit_with(&w.project, &landed, &[&w.head], "a commit nobody landed");
    git(&w.project, &["branch", "stray", &stray]);
    git(&w.project, &["branch", "-C", "stray", "integ"]);
    assert_eq!(w.integ(), stray);

    let second = w.review_card(&["one"]);
    let next = commit_with(&w.project, &stray, &[&stray], "a head on the drifted tip");
    w.pass_of(&second, &next);
    let before = w.dump();
    let out = w.land_lead(
        &second,
        &["--head", &next, "--met", "1=x", "--by", "leader"],
    );
    let message = w.refuse(&out, "conflict", &["drift"], &before, &stray);
    for needle in [stray.as_str(), landed.as_str(), "moai card drift --restore"] {
        assert!(
            message.contains(needle),
            "the drift answer lacks {needle}: {message}"
        );
    }
    assert!(
        !message.contains("git update-ref"),
        "the recovery is a moai command, never a bare git call: {message}"
    );
    assert_eq!(w.events(&second, "card.land_started"), 0);

    // After the recovery (the command's preview, then its apply with the user's yes) the same
    // call lands.
    let plan = result(&w.card(None, &["drift", "--restore"]));
    let digest = plan["plan"]["plan_digest"].as_str().unwrap().to_string();
    result(&w.card(
        None,
        &[
            "drift",
            "--restore",
            "--apply",
            "--plan-digest",
            &digest,
            "--permission-reference",
            "user-pick-ac23",
            "--by",
            "leader",
        ],
    ));
    assert_eq!(w.integ(), landed);
    let after = commit_with(
        &w.project,
        &landed,
        &[&landed],
        "a head on the recorded tip",
    );
    result(&w.land_pass(
        &second,
        &["--head", &after, "--met", "1=x", "--by", "leader"],
    ));
    assert_eq!(w.integ(), after);
}

/// AC-24 (SPEC-20261121-ref-guard): the lander is killed right after the branch moved and before
/// the close; the same call run again moves nothing, writes no second start row and closes. Only
/// a call that dies at that point shows a rerun that writes the row for a head already inside
/// the tip.
#[test]
fn ac24_a_kill_after_the_move_then_a_rerun_writes_no_second_row() {
    let w = World::new();
    let card = w.review_card(&["one"]);
    let candidate = w.candidate();
    w.wrap_update_ref("git-kills-the-lander", "", "kill -9 $PPID");
    let out = w.land_pass(
        &card,
        &["--head", &candidate, "--met", "1=x", "--by", "leader"],
    );
    assert_ne!(out.status.code(), Some(0), "the lander was not killed");
    assert_eq!(w.integ(), candidate, "the branch did not move");
    assert_eq!(w.events(&card, "card.land_started"), 1);
    assert_eq!(w.events(&card, "card.landed"), 0);

    *w.git_bin.borrow_mut() = git_executable();
    let rerun = result(&w.land_pass(
        &card,
        &["--head", &candidate, "--met", "1=x", "--by", "leader"],
    ));
    assert_eq!(rerun["moved"], json!(false), "{rerun}");
    assert_eq!(
        w.events(&card, "card.land_started"),
        1,
        "a second start row"
    );
    assert_eq!(w.events(&card, "card.landed"), 1);
    assert_eq!(w.integ(), candidate);
}

/// AC-25 (SPEC-20261121-ref-guard): an unlanded descendant is drift. A fast-forward made by hand
/// keeps the last landed head inside the branch, but the commits it added were never landed by
/// this command, so the next landing refuses and names the recorded commit.
#[test]
fn ac25_an_unlanded_descendant_of_the_last_landing_is_a_drift() {
    let w = World::new();
    let first = w.review_card(&["one"]);
    let landed = w.candidate();
    result(&w.land_pass(
        &first,
        &["--head", &landed, "--met", "1=x", "--by", "leader"],
    ));

    let ahead = commit_with(&w.project, &landed, &[&landed], "a fast-forward by hand");
    git(
        &w.project,
        &["update-ref", "refs/heads/integ", &ahead, &landed],
    );
    let second = w.review_card(&["one"]);
    let next = commit_with(&w.project, &ahead, &[&ahead], "a head on the moved tip");
    w.pass_of(&second, &next);
    let before = w.dump();
    let out = w.land_lead(
        &second,
        &["--head", &next, "--met", "1=x", "--by", "leader"],
    );
    let message = w.refuse(&out, "conflict", &["drift"], &before, &ahead);
    assert!(message.contains(&landed), "{message}");
    assert!(message.contains(&ahead), "{message}");
}

/// A fixture with two landings: card one lands `first`, card two lands `second` on top of it.
/// Returns the world and the two result commits.
fn two_landings() -> (World, String, String) {
    let w = World::new();
    let one = w.review_card(&["one"]);
    let first = w.candidate();
    result(&w.land_pass(&one, &["--head", &first, "--met", "1=x", "--by", "leader"]));
    let two = w.review_card(&["one"]);
    let lane = commit_with(&w.project, &w.head, &[&w.head], "the second lane head");
    let second = commit_with(
        &w.project,
        &first,
        &[&first, &lane],
        "merge the second lane",
    );
    result(&w.land_pass(&two, &["--head", &second, "--met", "1=x", "--by", "leader"]));
    assert_eq!(w.integ(), second);
    (w, first, second)
}

/// AC-26 (SPEC-20261121-ref-guard, the second re-check's B-3): a copy that rewinds the branch to
/// where the previous landing started is drift. The tip must be the result commit of the newest
/// landing the ledger recorded; the tip the landing started from is not accepted, or the next
/// landing would drop a landed change.
#[test]
fn ac26_a_rewind_to_the_previous_landings_start_is_a_drift() {
    let (w, first, second) = two_landings();
    git(&w.project, &["branch", "old", &first]);
    git(&w.project, &["branch", "-C", "old", "integ"]);
    assert_eq!(w.integ(), first);

    let three = w.review_card(&["one"]);
    let next = commit_with(&w.project, &first, &[&first], "a head on the rewound tip");
    w.pass_of(&three, &next);
    let before = w.dump();
    let out = w.land_lead(&three, &["--head", &next, "--met", "1=x", "--by", "leader"]);
    let message = w.refuse(&out, "conflict", &["drift"], &before, &first);
    assert!(
        message.contains(&second),
        "the result commit is not named: {message}"
    );
    assert_eq!(w.events(&three, "card.land_started"), 0);
}

/// AC-27: the two interrupted landings keep working. A move that git refused leaves a start row
/// with no landed row and the branch at its from-tip; a move that was done and not closed leaves
/// the branch at the row's head. A landing of another card is accepted on either tip.
#[test]
fn ac27_an_interrupted_landing_is_accepted_on_its_from_tip_and_its_target() {
    // (a) The move was refused: git's update-ref fails and nothing moved.
    let w = World::new();
    let stuck = w.review_card(&["one"]);
    let candidate = w.candidate();
    w.wrap_update_ref("git-refuses", "exit 1", "");
    let out = w.land_pass(
        &stuck,
        &["--head", &candidate, "--met", "1=x", "--by", "leader"],
    );
    assert_ne!(out.status.code(), Some(0));
    assert_eq!(w.integ(), w.tip, "the refused move changed the branch");
    assert_eq!(w.events(&stuck, "card.land_started"), 1);
    assert_eq!(w.events(&stuck, "card.landed"), 0);
    *w.git_bin.borrow_mut() = git_executable();
    let other = w.review_card(&["one"]);
    result(&w.land_pass(
        &other,
        &["--head", &candidate, "--met", "1=x", "--by", "leader"],
    ));
    assert_eq!(w.integ(), candidate);

    // (b) The move was done and the close is missing: the lander was killed after the move.
    let v = World::new();
    let killed = v.review_card(&["one"]);
    let moved = v.candidate();
    v.wrap_update_ref("git-kills-the-lander", "", "kill -9 $PPID");
    let out = v.land_pass(
        &killed,
        &["--head", &moved, "--met", "1=x", "--by", "leader"],
    );
    assert_ne!(out.status.code(), Some(0));
    assert_eq!(v.integ(), moved);
    *v.git_bin.borrow_mut() = git_executable();
    let next_card = v.review_card(&["one"]);
    let lane = commit_with(&v.project, &v.head, &[&v.head], "another lane head");
    let on_top = commit_with(
        &v.project,
        &moved,
        &[&moved, &lane],
        "merge on the moved tip",
    );
    result(&v.land_pass(
        &next_card,
        &["--head", &on_top, "--met", "1=x", "--by", "leader"],
    ));
    assert_eq!(v.integ(), on_top);
}

/// AC-28: `card drift` reads and writes nothing. It says `ok` on the recorded commit, `drift`
/// with the restore command off it, and `no_landing` for a branch the ledger never moved.
#[test]
fn ac28_the_drift_view_compares_the_tip_with_the_ledger_and_writes_nothing() {
    let w = World::new();
    let none = result(&w.card(None, &["drift"]));
    assert_eq!(none["state"], json!("no_landing"), "{none}");
    assert_eq!(none["writes"], json!(false), "{none}");

    let (w, first, second) = two_landings();
    let before = w.dump();
    let ok = result(&w.card(None, &["drift"]));
    assert_eq!(ok["state"], json!("ok"), "{ok}");
    assert_eq!(ok["branch"], json!("integ"), "{ok}");
    assert_eq!(ok["tip"], json!(second), "{ok}");
    assert_eq!(ok["expected"]["commit"], json!(second), "{ok}");

    git(&w.project, &["branch", "old", &first]);
    git(&w.project, &["branch", "-C", "old", "integ"]);
    let drift = result(&w.card(None, &["drift"]));
    assert_eq!(drift["state"], json!("drift"), "{drift}");
    assert_eq!(drift["tip"], json!(first), "{drift}");
    assert_eq!(drift["expected"]["commit"], json!(second), "{drift}");
    assert!(
        drift["restore_command"]
            .as_str()
            .unwrap()
            .contains("moai card drift --restore"),
        "{drift}"
    );
    assert_eq!(w.dump(), before, "the view wrote to the ledger");
    assert_eq!(w.integ(), first, "the view moved the branch");
}

/// AC-31: an interrupted landing after a successful one does not become the recorded commit. The
/// third landing starts and git refuses the move; the branch stays on the second result, which is
/// no drift. A copy that then rewinds the branch is drift, and the commit a restore names is the
/// second landing's result, not the head the interrupted landing was heading for.
#[test]
fn ac31_an_interrupted_landing_after_a_landing_is_not_the_recorded_commit() {
    let (w, first, second) = two_landings();
    let three = w.review_card(&["one"]);
    let wanted = commit_with(
        &w.project,
        &second,
        &[&second],
        "the head of the stuck landing",
    );
    w.wrap_update_ref("git-refuses", "exit 1", "");
    let out = w.land_pass(
        &three,
        &["--head", &wanted, "--met", "1=x", "--by", "leader"],
    );
    assert_ne!(out.status.code(), Some(0));
    assert_eq!(w.events(&three, "card.land_started"), 1);
    assert_eq!(w.events(&three, "card.landed"), 0);
    *w.git_bin.borrow_mut() = git_executable();
    let ok = result(&w.card(None, &["drift"]));
    assert_eq!(ok["state"], json!("ok"), "{ok}");
    assert_eq!(ok["expected"]["commit"], json!(second), "{ok}");

    git(&w.project, &["branch", "old", &first]);
    git(&w.project, &["branch", "-C", "old", "integ"]);
    let drift = result(&w.card(None, &["drift", "--restore"]));
    assert_eq!(drift["state"], json!("drift"), "{drift}");
    assert_eq!(drift["expected"]["commit"], json!(second), "{drift}");
    assert_eq!(drift["plan"]["to"], json!(second), "{drift}");
    assert_ne!(drift["expected"]["card"], json!(three), "{drift}");
}

/// A name-to-bytes snapshot of every file under `dir`, so two moments can be compared.
fn tree_of(dir: &Path) -> std::collections::BTreeMap<String, Vec<u8>> {
    fn walk(base: &Path, dir: &Path, out: &mut std::collections::BTreeMap<String, Vec<u8>>) {
        for entry in fs::read_dir(dir).unwrap().flatten() {
            let path = entry.path();
            if path.is_dir() {
                walk(base, &path, out);
            } else {
                out.insert(
                    path.strip_prefix(base).unwrap().display().to_string(),
                    fs::read(&path).unwrap_or_default(),
                );
            }
        }
    }
    let mut out = std::collections::BTreeMap::new();
    walk(dir, dir, &mut out);
    out
}

/// AC-32 (B-4 of the third re-check): with the hook installed, the restore passes the guard for a
/// descendant nobody landed and for an unrelated copy, not only for a rewind. Every restore test
/// runs with the guard; a bare `git update-ref` over the branch is still refused.
#[test]
fn ac32_the_restore_passes_the_installed_guard_for_a_descendant_and_an_unrelated_copy() {
    let (w, first, second) = two_landings();
    w.install_guard();
    let apply = |digest: &str, reference: &str| {
        w.card(
            None,
            &[
                "drift",
                "--restore",
                "--apply",
                "--plan-digest",
                digest,
                "--permission-reference",
                reference,
                "--by",
                "leader",
            ],
        )
    };
    let preview = || {
        let plan = result(&w.card(None, &["drift", "--restore"]));
        plan["plan"]["plan_digest"].as_str().unwrap().to_string()
    };

    // (a) A descendant of the last landing that nobody landed.
    let ahead = commit_with(&w.project, &second, &[&second], "an unlanded change");
    w.copy_over(&ahead);
    let digest = preview();
    result(&apply(&digest, "user-pick-descendant"));
    assert_eq!(w.integ(), second, "the descendant was not restored");

    // (b) An unrelated copy: neither an ancestor nor a descendant of the last landing.
    let stray = commit_with(&w.project, &first, &[&w.head], "a commit nobody landed");
    w.copy_over(&stray);
    let digest = preview();
    result(&apply(&digest, "user-pick-unrelated"));
    assert_eq!(w.integ(), second, "the unrelated copy was not restored");
    assert_eq!(
        w.count("SELECT count(*) FROM event WHERE kind = 'card.tip_restored'"),
        2
    );
    assert!(
        !w.common().join("moai/land-intent").exists(),
        "the intent record stays after the restore"
    );

    // A bare update-ref over the branch is refused: the restore intent belongs to the command.
    let bare = Command::new(git_executable())
        .args(["update-ref", "refs/heads/integ", &stray, &second])
        .current_dir(&w.project)
        .env(
            "PATH",
            w.extra_env
                .borrow()
                .iter()
                .find(|(k, _)| k == "PATH")
                .unwrap()
                .1
                .clone(),
        )
        .env("HOME", w.home.path())
        .output()
        .unwrap();
    assert!(!bare.status.success(), "a bare update-ref passed the guard");
    assert_eq!(w.integ(), second);
}

/// AC-33 (N-20): the digest is bound to a nonce the preview writes into the git folder. A digest
/// made without the preview (the plain hash of the move), one from an older preview and one that
/// was already used are all refused; the apply consumes the nonce.
#[test]
fn ac33_the_digest_needs_the_nonce_of_the_latest_preview_and_is_used_once() {
    use sha2::{Digest, Sha256};
    let (w, first, second) = two_landings();
    w.install_guard();
    w.copy_over(&first);
    let apply = |digest: &str| {
        w.card(
            None,
            &[
                "drift",
                "--restore",
                "--apply",
                "--plan-digest",
                digest,
                "--permission-reference",
                "user-pick",
                "--by",
                "leader",
            ],
        )
    };
    let before = w.dump();
    let one = result(&w.card(None, &["drift", "--restore"]));
    let first_digest = one["plan"]["plan_digest"].as_str().unwrap().to_string();
    assert!(
        w.common().join("moai/drift-nonce").is_file(),
        "the preview wrote no nonce"
    );
    assert_eq!(w.dump(), before, "the preview wrote to the ledger");
    let two = result(&w.card(None, &["drift", "--restore"]));
    let second_digest = two["plan"]["plan_digest"].as_str().unwrap().to_string();
    assert_ne!(first_digest, second_digest, "two previews share a digest");

    // The plain hash of the move, computable without the preview, is refused.
    let card_number = two["plan"]["card"]
        .as_str()
        .unwrap()
        .trim_start_matches("C-")
        .to_string();
    let plain = format!(
        "{:x}",
        Sha256::digest(format!(
            "moai card drift --restore\ninteg\n{first}\n{second}\n{card_number}\n"
        ))
    );
    for (what, digest) in [
        ("the plain hash", plain.as_str()),
        ("an older preview's digest", first_digest.as_str()),
    ] {
        let out = apply(digest);
        assert_eq!(
            out.status.code(),
            Some(1),
            "{what}: {}",
            String::from_utf8_lossy(&out.stdout)
        );
        assert_eq!(w.integ(), first, "{what} moved the branch");
        assert_eq!(w.dump(), before, "{what} wrote to the ledger");
    }

    result(&apply(&second_digest));
    assert_eq!(w.integ(), second);
    assert!(
        !w.common().join("moai/drift-nonce").exists(),
        "the apply left the nonce"
    );

    // The same digest after the state drifts again: the nonce is gone, so it is refused.
    w.copy_over(&first);
    let replay = apply(&second_digest);
    assert_eq!(
        replay.status.code(),
        Some(1),
        "{}",
        String::from_utf8_lossy(&replay.stdout)
    );
    assert_eq!(w.integ(), first, "a used digest moved the branch");
}

/// AC-34 (N-21): the plain view writes no file, neither in the ledger nor in the git folder, in
/// the states ok, drift and no_landing; the preview writes only the nonce file.
#[test]
fn ac34_the_view_writes_no_file_and_the_preview_writes_only_the_nonce() {
    let w = World::new();
    w.install_guard();
    let git_before = tree_of(&w.common());
    let ledger_before = w.dump();
    let none = result(&w.card(None, &["drift"]));
    assert_eq!(none["state"], json!("no_landing"), "{none}");
    assert_eq!(tree_of(&w.common()), git_before, "the view wrote a file");
    assert_eq!(w.dump(), ledger_before);

    let (w, first, _second) = two_landings();
    w.install_guard();
    let git_before = tree_of(&w.common());
    let ledger_before = w.dump();
    let ok = result(&w.card(None, &["drift"]));
    assert_eq!(ok["state"], json!("ok"), "{ok}");
    assert_eq!(tree_of(&w.common()), git_before, "the view wrote a file");
    assert_eq!(w.dump(), ledger_before);

    w.copy_over(&first);
    let git_before = tree_of(&w.common());
    let drift = result(&w.card(None, &["drift"]));
    assert_eq!(drift["state"], json!("drift"), "{drift}");
    assert_eq!(tree_of(&w.common()), git_before, "the view wrote a file");
    assert_eq!(w.dump(), ledger_before);

    result(&w.card(None, &["drift", "--restore"]));
    let after = tree_of(&w.common());
    let added: Vec<&String> = after
        .keys()
        .filter(|k| !git_before.contains_key(*k))
        .collect();
    assert_eq!(
        added,
        vec!["moai/drift-nonce"],
        "the preview wrote more than the nonce"
    );
    for (name, bytes) in &git_before {
        assert_eq!(after.get(name), Some(bytes), "the preview changed {name}");
    }
    assert_eq!(w.dump(), ledger_before);
}

/// AC-29: the restore previews first and moves only on the user's yes. The preview names the
/// move and carries a digest and moves nothing; the apply needs the leader word, the digest of a
/// preview of the same state and a permission reference; it moves the branch to the recorded
/// commit under the guard and records one event with the reference.
#[test]
fn ac29_the_restore_needs_the_preview_digest_and_a_permission_reference() {
    let (w, first, second) = two_landings();
    w.install_guard();
    git(&w.project, &["branch", "old", &first]);
    git(&w.project, &["branch", "-C", "old", "integ"]);
    let before = w.dump();

    let plan = result(&w.card(None, &["drift", "--restore"]));
    assert_eq!(plan["plan"]["from"], json!(first), "{plan}");
    assert_eq!(plan["plan"]["to"], json!(second), "{plan}");
    assert_eq!(plan["plan"]["moved"], json!(false), "{plan}");
    let digest = plan["plan"]["plan_digest"].as_str().unwrap().to_string();
    assert_eq!(w.dump(), before, "the preview wrote to the ledger");
    assert_eq!(w.integ(), first, "the preview moved the branch");

    // Every apply that lacks one of its three inputs, or has a wrong one, refuses and moves nothing.
    let apply = |extra: &[&str]| {
        let mut args = vec!["drift", "--restore", "--apply"];
        args.extend_from_slice(extra);
        w.card(None, &args)
    };
    for (what, extra) in [
        (
            "no digest",
            vec!["--permission-reference", "user-pick", "--by", "leader"],
        ),
        (
            "no reference",
            vec!["--plan-digest", digest.as_str(), "--by", "leader"],
        ),
        (
            "a wrong digest",
            vec![
                "--plan-digest",
                "0000000000000000000000000000000000000000000000000000000000000000",
                "--permission-reference",
                "user-pick",
                "--by",
                "leader",
            ],
        ),
        (
            "another caller",
            vec![
                "--plan-digest",
                digest.as_str(),
                "--permission-reference",
                "user-pick",
                "--by",
                "agent-1",
            ],
        ),
    ] {
        let out = apply(&extra);
        assert_eq!(
            out.status.code(),
            Some(1),
            "{what}: {}",
            String::from_utf8_lossy(&out.stdout)
        );
        assert_eq!(w.integ(), first, "{what} moved the branch");
        assert_eq!(w.dump(), before, "{what} wrote to the ledger");
    }
    // A bare git call is still refused by the guard: only the command's intent record passes.
    let bare = Command::new(git_executable())
        .args(["update-ref", "refs/heads/integ", &second, &first])
        .current_dir(&w.project)
        .env(
            "PATH",
            w.extra_env
                .borrow()
                .iter()
                .find(|(k, _)| k == "PATH")
                .unwrap()
                .1
                .clone(),
        )
        .env("HOME", w.home.path())
        .output()
        .unwrap();
    assert!(!bare.status.success(), "a bare update-ref passed the guard");
    assert_eq!(w.integ(), first);

    let done = result(&apply(&[
        "--plan-digest",
        &digest,
        "--permission-reference",
        "user-pick",
        "--by",
        "leader",
    ]));
    assert_eq!(done["plan"]["moved"], json!(true), "{done}");
    assert_eq!(w.integ(), second);
    assert_eq!(
        w.count("SELECT count(*) FROM event WHERE kind = 'card.tip_restored'"),
        1
    );
    assert!(
        w.text("SELECT payload FROM event WHERE kind = 'card.tip_restored'")
            .contains("user-pick"),
        "the event lacks the permission reference"
    );
    assert!(
        !w.common().join("moai/land-intent").exists(),
        "the intent record stays after the restore"
    );
    // The branch is where the ledger says: the view is ok and the next landing goes through.
    assert_eq!(result(&w.card(None, &["drift"]))["state"], json!("ok"));
}

/// AC-30: a restore of an unlanded descendant lists what it drops; a restore with no drift, and a
/// digest of a state that has since changed, refuse and move nothing.
#[test]
fn ac30_the_restore_names_what_it_drops_and_refuses_a_stale_digest() {
    let w = World::new();
    w.install_guard();
    let first = w.review_card(&["one"]);
    let landed = w.candidate();
    result(&w.land_pass(
        &first,
        &["--head", &landed, "--met", "1=x", "--by", "leader"],
    ));
    let none = w.card(None, &["drift", "--restore"]);
    assert_eq!(
        none.status.code(),
        Some(1),
        "{}",
        String::from_utf8_lossy(&none.stdout)
    );
    assert!(
        envelope(&none)["error"]["message"]
            .as_str()
            .unwrap()
            .contains("no drift"),
        "{}",
        String::from_utf8_lossy(&none.stdout)
    );

    let ahead = commit_with(&w.project, &landed, &[&landed], "an unlanded change");
    w.copy_over(&ahead);
    let plan = result(&w.card(None, &["drift", "--restore"]));
    let drops = plan["plan"]["drops"].as_array().unwrap();
    assert_eq!(drops.len(), 1, "{plan}");
    assert!(
        drops[0].as_str().unwrap().starts_with(&ahead[..12]),
        "{plan}"
    );
    let digest = plan["plan"]["plan_digest"].as_str().unwrap().to_string();

    // The state changes after the preview: the digest no longer matches.
    let later = commit_with(&w.project, &ahead, &[&ahead], "another unlanded change");
    w.copy_over(&later);
    let out = w.card(
        None,
        &[
            "drift",
            "--restore",
            "--apply",
            "--plan-digest",
            &digest,
            "--permission-reference",
            "user-pick",
            "--by",
            "leader",
        ],
    );
    assert_eq!(
        out.status.code(),
        Some(1),
        "{}",
        String::from_utf8_lossy(&out.stdout)
    );
    assert_eq!(w.integ(), later, "a stale digest moved the branch");
}

// ---- SPEC-20261123-landing-reads-verdict --------------------------------------------------------
//
// The landing reads the verdict events of the card. The events are written here as ledger rows
// with the payload the SPEC names (D-1), so the record's shape is pinned by these tests and the
// landing is tested apart from the `card verdict` command (`tests/card_verdict_cli.rs`).

/// The profile of the fixture with the verdict check in enforce mode.
const ENFORCE_PROFILE: &[u8] = b"{\"schema_version\": 1, \"project_id\": \"demo\", \"purpose\": \"fixture\", \"tools\": [\"rust\"], \"development_methods\": [\"sdd\"], \"hosts\": [\"claude\"], \"work_flow\": {\"flow\": \"m-flow\", \"integration_branch\": \"integ\", \"land_verdict\": \"enforce\"}}";
/// The same profile in observe mode.
const OBSERVE_PROFILE: &[u8] = b"{\"schema_version\": 1, \"project_id\": \"demo\", \"purpose\": \"fixture\", \"tools\": [\"rust\"], \"development_methods\": [\"sdd\"], \"hosts\": [\"claude\"], \"work_flow\": {\"flow\": \"m-flow\", \"integration_branch\": \"integ\", \"land_verdict\": \"observe\"}}";
/// The same profile with a word the key does not take.
const BAD_MODE_PROFILE: &[u8] = b"{\"schema_version\": 1, \"project_id\": \"demo\", \"purpose\": \"fixture\", \"tools\": [\"rust\"], \"development_methods\": [\"sdd\"], \"hosts\": [\"claude\"], \"work_flow\": {\"flow\": \"m-flow\", \"integration_branch\": \"integ\", \"land_verdict\": \"block\"}}";
/// The independent reviewer, another reviewer and the lander (the leader's session).
const S_AUD: &str = "cccccccc-0000-4000-8000-00000000000c";
const S_OTHER: &str = "eeeeeeee-0000-4000-8000-00000000000e";
const S_LEAD: &str = "dddddddd-0000-4000-8000-00000000000d";
const WORKTREE_KEY: &str = "0123456789ab";

/// A commit message with the trailer block of `session`.
fn signed(subject: &str, session: &str) -> String {
    format!(
        "{subject}\n\nMoai-Spec: SPEC-20261123-fixture\nMoai-Session: {session}:{WORKTREE_KEY}\nMoai-Verified: none"
    )
}

impl World {
    /// A fixture whose profile asks for the verdict check.
    fn enforced() -> Self {
        let w = World::new();
        w.set_profile(ENFORCE_PROFILE);
        w
    }

    /// A lane commit on the card's head that adds `path`, written by `session`.
    fn lane_adding(&self, path: &str, session: &str) -> String {
        put_on_msg(
            &self.project,
            &self.head,
            path,
            "lane\n",
            &signed("the lane change", session),
        )
    }

    /// The commit the leader lands: the lane on the tip, the tree git makes of the two, signed
    /// by the lander.
    fn join(&self, lane: &str, lander: &str) -> String {
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
                &signed("merge the lane on the tip", lander),
            ],
        )
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
            "head": head, "base": self.pass_base(head), "verdict": if pass { "pass" } else { "fail" },
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

    /// `card land` as `session` (the lander) for `head`.
    fn land_as(&self, session: Option<&str>, card: &str, head: &str) -> Output {
        self.card(
            session,
            &[
                "land",
                card,
                "--head",
                head,
                "--met",
                "1=the test",
                "--by",
                "leader",
            ],
        )
    }

    /// The payload of the newest event of `kind` of `card`.
    fn payload(&self, card: &str, kind: &str) -> Value {
        let text = self.text(&format!(
            "SELECT payload FROM event WHERE card_id = {} AND kind = '{kind}' \
             ORDER BY event_id DESC LIMIT 1",
            number(card)
        ));
        serde_json::from_str(&text).unwrap()
    }
}

/// AC-4 (REQ-7, REQ-8, REQ-9, REQ-10): the four refusals of the card, each with its own code,
/// the branch and every table unchanged and not even a start row.
#[test]
fn ac40_the_four_refusals_of_the_card_each_name_their_code_and_write_nothing() {
    let w = World::enforced();
    let card = w.review_card(&["one"]);
    let lane = w.lane_adding("notes/x.md", S_A);
    let cand = w.join(&lane, S_LEAD);

    // No verdict at all.
    let before = w.dump();
    let out = w.land_as(Some(S_LEAD), &card, &cand);
    w.refuse(&out, "invalid_input", &["no_verdict"], &before, &w.tip);

    // A PASS recorded for another head only.
    w.verdict(&card, &w.head, true, 0, S_AUD, None);
    let before = w.dump();
    let out = w.land_as(Some(S_LEAD), &card, &cand);
    let message = w.refuse(
        &out,
        "invalid_input",
        &["verdict_other_head"],
        &before,
        &w.tip,
    );
    assert!(
        message.contains(&w.head),
        "names the newest verdict head: {message}"
    );

    // A FAIL on the head.
    w.verdict(&card, &cand, false, 0, S_AUD, None);
    let before = w.dump();
    let out = w.land_as(Some(S_LEAD), &card, &cand);
    w.refuse(&out, "invalid_input", &["verdict_fail"], &before, &w.tip);
    assert_eq!(w.events(&card, "card.land_started"), 0);

    // A PASS by the session that wrote the lane commit.
    let w = World::enforced();
    let card = w.review_card(&["one"]);
    let lane = w.lane_adding("notes/x.md", S_A);
    let cand = w.join(&lane, S_LEAD);
    w.verdict(&card, &cand, true, 0, S_A, None);
    let before = w.dump();
    let out = w.land_as(Some(S_LEAD), &card, &cand);
    w.refuse(&out, "invalid_input", &["self_verdict"], &before, &w.tip);
    assert_eq!(w.events(&card, "card.land_started"), 0);
}

/// AC-5 (REQ-8): one reviewer's newest FAIL refuses whoever else passed the head; the same
/// reviewer's later PASS lands it.
#[test]
fn ac41_one_reviewers_newest_fail_refuses_until_the_same_reviewer_passes() {
    let w = World::enforced();
    let card = w.review_card(&["one"]);
    let lane = w.lane_adding("notes/x.md", S_A);
    let cand = w.join(&lane, S_LEAD);
    w.verdict(&card, &cand, false, 0, S_AUD, None);
    w.verdict(&card, &cand, true, 0, S_OTHER, None);
    let before = w.dump();
    let out = w.land_as(Some(S_LEAD), &card, &cand);
    w.refuse(&out, "invalid_input", &["verdict_fail"], &before, &w.tip);

    w.verdict(&card, &cand, true, 0, S_AUD, None);
    let got = result(&w.land_as(Some(S_LEAD), &card, &cand));
    assert_eq!(got["moved"], json!(true), "{got}");
    assert_eq!(got["closed"], json!(true), "{got}");
    assert_eq!(w.integ(), cand);
}

/// AC-6 (REQ-10): the holder of a reservation of the card is an author even when no commit
/// carries the holder's trailer, and a bare session id in a trailer is the same session as the
/// full value.
#[test]
fn ac42_a_reservation_holder_and_a_bare_trailer_id_are_authors() {
    let w = World::enforced();
    let card = w.review_card(&["one"]);
    // The lane commit is signed by another session; the holder of the card is `S_A`.
    let lane = w.lane_adding("notes/x.md", S_OTHER);
    let cand = w.join(&lane, S_LEAD);
    w.verdict(&card, &cand, true, 0, S_A, None);
    let before = w.dump();
    let out = w.land_as(Some(S_LEAD), &card, &cand);
    w.refuse(&out, "invalid_input", &["self_verdict"], &before, &w.tip);

    // A trailer with the bare session id (no worktree key) names the same session.
    let bare = put_on_msg(
        &w.project,
        &w.head,
        "notes/y.md",
        "y\n",
        &format!("the lane change\n\nMoai-Spec: SPEC-20261123-fixture\nMoai-Session: {S_AUD}\nMoai-Verified: none"),
    );
    let cand = w.join(&bare, S_LEAD);
    w.verdict(&card, &cand, true, 0, S_AUD, None);
    let before = w.dump();
    let out = w.land_as(Some(S_LEAD), &card, &cand);
    w.refuse(&out, "invalid_input", &["self_verdict"], &before, &w.tip);
}

/// AC-7 (REQ-11): commits that carry no `Moai-Session` trailer leave the authors unknown.
#[test]
fn ac43_a_range_with_no_session_trailer_refuses_as_authors_unknown() {
    let w = World::enforced();
    let card = w.review_card(&["one"]);
    let lane = put_on_msg(&w.project, &w.head, "notes/x.md", "x\n", "the lane change");
    let cand = w.join(&lane, S_LEAD);
    w.verdict(&card, &cand, true, 0, S_AUD, None);
    let before = w.dump();
    let out = w.land_as(Some(S_LEAD), &card, &cand);
    w.refuse(&out, "invalid_input", &["authors_unknown"], &before, &w.tip);
}

/// AC-8 (REQ-12): at tier 1 and above a PASS needs a receipt of the head with a step and every
/// step at exit 0; the tier is the higher of the stated one and the floor of the paths.
#[test]
fn ac44_a_tier_one_pass_needs_a_receipt_of_the_head_with_every_step_green() {
    let w = World::enforced();
    let card = w.review_card(&["one"]);
    let lane = w.lane_adding("src/a.rs", S_A);
    let cand = w.join(&lane, S_LEAD);
    let bad = [
        None,
        Some(w.good_receipt(&w.tip)),
        Some(json!({"head": cand, "steps": []})),
        Some(json!({"head": cand, "steps": [{"name": "test", "exit": 1}]})),
    ];
    for receipt in bad {
        // A stated tier 0 on a path that is not tier 0 safe is judged at tier 1.
        w.verdict(&card, &cand, true, 0, S_AUD, receipt.clone());
        let before = w.dump();
        let out = w.land_as(Some(S_LEAD), &card, &cand);
        w.refuse(&out, "invalid_input", &["receipt_missing"], &before, &w.tip);
    }
    w.verdict(&card, &cand, true, 1, S_AUD, Some(w.good_receipt(&cand)));
    let got = result(&w.land_as(Some(S_LEAD), &card, &cand));
    assert_eq!(got["closed"], json!(true), "{got}");

    // The floor of a tier 2 path beats a stated tier 0.
    let w = World::enforced();
    let card = w.review_card(&["one"]);
    let lane = w.lane_adding("src/gate_x.rs", S_A);
    let cand = w.join(&lane, S_LEAD);
    w.verdict(&card, &cand, true, 0, S_AUD, None);
    let before = w.dump();
    let out = w.land_as(Some(S_LEAD), &card, &cand);
    w.refuse(&out, "invalid_input", &["receipt_missing"], &before, &w.tip);
}

/// AC-9 and AC-17 (REQ-6, REQ-13): an independent PASS lands the head in one call, and the start
/// row, the landing event and the answer name the verdict event, the reviewer, the tier and the
/// lander as separate values.
#[test]
fn ac45_an_independent_pass_lands_and_the_record_names_the_verdict() {
    let w = World::enforced();
    let card = w.review_card(&["one"]);
    let lane = w.lane_adding("src/a.rs", S_A);
    let cand = w.join(&lane, S_LEAD);
    let event = w.verdict(&card, &cand, true, 1, S_AUD, Some(w.good_receipt(&cand)));
    let got = result(&w.land_as(Some(S_LEAD), &card, &cand));
    assert_eq!(got["moved"], json!(true), "{got}");
    assert_eq!(got["closed"], json!(true), "{got}");
    assert_eq!(w.integ(), cand);
    let named = json!({"event": event, "reviewer": S_AUD, "tier": 1, "lander": S_LEAD});
    assert_eq!(got["verdict"], named, "{got}");
    assert_eq!(w.payload(&card, "card.land_started")["verdict"], named);
    assert_eq!(w.payload(&card, "card.landed")["verdict"], named);
    let id = number(&card);
    assert_eq!(
        w.text(&format!(
            "SELECT status FROM reservation WHERE card_id = {id} ORDER BY generation DESC LIMIT 1"
        )),
        "released"
    );
}

/// AC-10 (REQ-16): a second call on the landed card answers done and writes nothing, also after a
/// FAIL was recorded for the head.
#[test]
fn ac46_a_done_card_answers_done_whatever_the_verdicts_say_now() {
    let w = World::enforced();
    let card = w.review_card(&["one"]);
    let lane = w.lane_adding("notes/x.md", S_A);
    let cand = w.join(&lane, S_LEAD);
    w.verdict(&card, &cand, true, 0, S_AUD, None);
    result(&w.land_as(Some(S_LEAD), &card, &cand));
    w.verdict(&card, &cand, false, 0, S_OTHER, None);
    let before = w.dump();
    let again = result(&w.land_as(Some(S_LEAD), &card, &cand));
    assert_eq!(again["already_done"], json!(true), "{again}");
    assert_eq!(w.dump(), before, "a call on a done card wrote");
}

/// AC-11 (REQ-15, REQ-17): a landing killed after the move is closed by the same call run again
/// from the verdict the start row names, and a FAIL recorded in between is named, not obeyed.
#[test]
fn ac47_a_rerun_after_the_move_uses_the_start_rows_verdict_and_names_a_late_fail() {
    let w = World::enforced();
    let card = w.review_card(&["one"]);
    let lane = w.lane_adding("notes/x.md", S_A);
    let cand = w.join(&lane, S_LEAD);
    let event = w.verdict(&card, &cand, true, 0, S_AUD, None);
    w.wrap_update_ref("git-kills-the-lander", "", "kill -9 $PPID");
    let out = w.land_as(Some(S_LEAD), &card, &cand);
    assert_ne!(out.status.code(), Some(0), "the lander was not killed");
    assert_eq!(w.integ(), cand);
    assert_eq!(w.events(&card, "card.land_started"), 1);
    assert_eq!(w.events(&card, "card.landed"), 0);

    let late = w.verdict(&card, &cand, false, 0, S_OTHER, None);
    *w.git_bin.borrow_mut() = git_executable();
    let rerun = result(&w.land_as(Some(S_LEAD), &card, &cand));
    assert_eq!(rerun["moved"], json!(false), "{rerun}");
    assert_eq!(rerun["closed"], json!(true), "{rerun}");
    assert_eq!(rerun["verdict"]["event"], json!(event), "{rerun}");
    assert_eq!(
        rerun["late_fail"],
        json!({"event": late, "reviewer": S_OTHER}),
        "{rerun}"
    );
    assert_eq!(w.events(&card, "card.land_started"), 1);
    assert_eq!(w.events(&card, "card.landed"), 1);
}

/// AC-12 (REQ-18): a landing killed after the start row and before the move, and one killed
/// after the move and before the close, each end with one start row, one landing event and one
/// finish event when a fresh process runs the same call again. (The close is one write
/// transaction, so it has no stop point of its own.)
#[test]
fn ac48_a_kill_before_the_move_and_a_kill_after_it_each_end_in_one_landing() {
    // The script keeps running after it kills its parent, so the first kill also stops the
    // script and the branch is not moved.
    for (before, after, moved_again) in [
        ("kill -9 $PPID; exit 9", "", true),
        ("", "kill -9 $PPID", false),
    ] {
        let w = World::enforced();
        let card = w.review_card(&["one"]);
        let lane = w.lane_adding("notes/x.md", S_A);
        let cand = w.join(&lane, S_LEAD);
        w.verdict(&card, &cand, true, 0, S_AUD, None);
        w.wrap_update_ref("git-kills-the-lander", before, after);
        let out = w.land_as(Some(S_LEAD), &card, &cand);
        assert_ne!(out.status.code(), Some(0), "the lander was not killed");
        assert_eq!(w.events(&card, "card.land_started"), 1);
        assert_eq!(w.events(&card, "card.landed"), 0);
        *w.git_bin.borrow_mut() = git_executable();
        let rerun = result(&w.land_as(Some(S_LEAD), &card, &cand));
        assert_eq!(rerun["moved"], json!(moved_again), "{rerun}");
        assert_eq!(rerun["closed"], json!(true), "{rerun}");
        assert_eq!(w.integ(), cand);
        assert_eq!(
            w.events(&card, "card.land_started"),
            1,
            "a second start row"
        );
        assert_eq!(w.events(&card, "card.landed"), 1);
        assert_eq!(w.events(&card, "card.finish"), 1);
    }
}

/// AC-13 (REQ-18): a lock file on the branch reference refuses with git's reason, and the same
/// call run again after the cause is gone lands once.
#[test]
fn ac49_a_failure_between_the_steps_is_repaired_by_running_the_call_again() {
    let w = World::enforced();
    let card = w.review_card(&["one"]);
    let lane = w.lane_adding("notes/x.md", S_A);
    let cand = w.join(&lane, S_LEAD);
    w.verdict(&card, &cand, true, 0, S_AUD, None);
    let lock = w.project.join(".git/refs/heads/integ.lock");
    fs::write(&lock, "").unwrap();
    let out = w.land_as(Some(S_LEAD), &card, &cand);
    assert_eq!(out.status.code(), Some(1));
    assert_eq!(w.integ(), w.tip);
    assert_eq!(w.events(&card, "card.land_started"), 1);
    fs::remove_file(&lock).unwrap();
    let got = result(&w.land_as(Some(S_LEAD), &card, &cand));
    assert_eq!(got["moved"], json!(true), "{got}");
    assert_eq!(
        w.events(&card, "card.land_started"),
        1,
        "a second start row"
    );
    assert_eq!(w.events(&card, "card.landed"), 1);
}

/// AC-14 (REQ-6, REQ-18): a moved integration branch gives its three outcomes: `tip_moved` under
/// the call, which keeps the verdict valid for a later call; `stale_base` when the tip moved
/// before the call, which does not carry the PASS to the head that merges the new tip; and a
/// branch that already holds the head, which closes with no second move.
#[test]
fn ac50_a_moved_integration_branch_gives_tip_moved_stale_base_or_a_close() {
    let w = World::enforced();
    let card = w.review_card(&["one"]);
    let lane = w.lane_adding("notes/x.md", S_A);
    let cand = w.join(&lane, S_LEAD);
    w.verdict(&card, &cand, true, 0, S_AUD, None);
    let other = commit_with(&w.project, &w.tip, &[&w.tip], "somebody else landed first");

    // The tip read is still current, then another process moves the branch under the update.
    let wrapper = w.dir.path().join("git-moves-integ");
    common::write_executable(
        &wrapper,
        &format!(
            "#!/bin/sh\nfor a in \"$@\"; do\n  if [ \"$a\" = update-ref ]; then\n    '{git}' -C '{project}' update-ref refs/heads/integ {other}\n  fi\ndone\nexec '{git}' \"$@\"\n",
            git = git_executable().display(),
            project = w.project.display(),
        ),
    );
    *w.git_bin.borrow_mut() = wrapper;
    let before = w.dump();
    let out = w.land_as(Some(S_LEAD), &card, &cand);
    w.refuse_the_move(&out, "conflict", &["tip_moved"], &before, &other, &card);

    // The tip moved before the call: the head no longer has it on its first-parent line.
    *w.git_bin.borrow_mut() = git_executable();
    let before = w.dump();
    let out = w.land_as(Some(S_LEAD), &card, &cand);
    w.refuse(&out, "conflict", &["stale_base"], &before, &other);

    // The branch is put back; the PASS of the first head is still valid and lands it.
    git(&w.project, &["update-ref", "refs/heads/integ", &w.tip]);
    let got = result(&w.land_as(Some(S_LEAD), &card, &cand));
    assert_eq!(got["moved"], json!(true), "{got}");
    assert_eq!(got["closed"], json!(true), "{got}");

    // A tip that moved before any landing of the card: the head that merges the new tip is
    // another head, so the PASS of the first head does not carry over to it.
    let w = World::enforced();
    let card = w.review_card(&["one"]);
    let lane = w.lane_adding("notes/x.md", S_A);
    let cand = w.join(&lane, S_LEAD);
    w.verdict(&card, &cand, true, 0, S_AUD, None);
    let other = commit_with(&w.project, &w.tip, &[&w.tip], "somebody else landed first");
    git(&w.project, &["update-ref", "refs/heads/integ", &other]);
    let tree = git(&w.project, &["merge-tree", "--write-tree", &other, &lane]);
    let merged = git(
        &w.project,
        &[
            "commit-tree",
            &tree,
            "-p",
            &other,
            "-p",
            &lane,
            "-m",
            &signed("merge the lane on the new tip", S_LEAD),
        ],
    );
    let before = w.dump();
    let out = w.land_as(Some(S_LEAD), &card, &merged);
    w.refuse(
        &out,
        "invalid_input",
        &["verdict_other_head"],
        &before,
        &other,
    );

    // The branch moved outside the call to contain the head: it closes the card with no second
    // move and no start row.
    let w = World::enforced();
    let card = w.review_card(&["one"]);
    let lane = w.lane_adding("notes/x.md", S_A);
    let cand = w.join(&lane, S_LEAD);
    w.verdict(&card, &cand, true, 0, S_AUD, None);
    git(&w.project, &["update-ref", "refs/heads/integ", &cand]);
    let got = result(&w.land_as(Some(S_LEAD), &card, &cand));
    assert_eq!(got["moved"], json!(false), "{got}");
    assert_eq!(got["closed"], json!(true), "{got}");
    assert_eq!(w.events(&card, "card.land_started"), 0);
    assert_eq!(w.integ(), cand);
}

/// AC-15 (REQ-19): a note whose first line is a well-formed verdict line is no verdict.
#[test]
fn ac51_a_note_that_reads_like_a_verdict_is_no_verdict() {
    let w = World::enforced();
    let card = w.review_card(&["one"]);
    let lane = w.lane_adding("notes/x.md", S_A);
    let cand = w.join(&lane, S_LEAD);
    let note = format!(
        "AUDIT {card} (the fixture) head {cand} base {} — verdict: PASS",
        w.tip
    );
    let note_file = w.dir.path().join("verdict-note.md");
    fs::write(&note_file, &note).unwrap();
    let mut cmd = w
        .home
        .command(&["card", "note", &card, "--kind", "report", "--project", &w.p]);
    cmd.env(SESSION_VAR, S_AUD)
        .stdin(fs::File::open(&note_file).unwrap());
    result(&cmd.output().unwrap());
    let before = w.dump();
    let out = w.land_as(Some(S_LEAD), &card, &cand);
    w.refuse(&out, "invalid_input", &["no_verdict"], &before, &w.tip);
}

/// AC-16 (REQ-14): in observe mode a head with no verdict lands as it did before, with the
/// refusal it would have made in the answer; the same call in enforce mode refuses.
#[test]
fn ac52_observe_mode_reports_would_refuse_and_lands() {
    let w = World::new();
    w.set_profile(OBSERVE_PROFILE);
    let card = w.review_card(&["one"]);
    let lane = w.lane_adding("notes/x.md", S_A);
    let cand = w.join(&lane, S_LEAD);
    // A landing that would close the card needs the PASS in every mode (SPEC-20261124-land-
    // closes-card); a slice lands in observe mode and carries the refusal it would have had.
    let before = w.dump();
    let out = w.land_as(Some(S_LEAD), &card, &cand);
    w.refuse(&out, "invalid_input", &["no_pass"], &before, &w.tip);
    let got = result(&w.land_plain(Some(S_LEAD), &card, &cand, &["--keep-open"]));
    assert_eq!(got["moved"], json!(true), "{got}");
    assert_eq!(got["closed"], json!(false), "{got}");
    assert_eq!(got["would_refuse"]["code"], json!("no_verdict"), "{got}");
    assert!(got.get("verdict").is_none(), "{got}");
    assert!(w
        .payload(&card, "card.land_started")
        .get("verdict")
        .is_none());
}

/// AC-17 (REQ-13, REQ-20): at tier 0 the reviewer and the lander may be one session and both are
/// recorded, as two values.
#[test]
fn ac53_the_lander_may_review_tier_zero_and_both_identities_are_recorded() {
    let w = World::enforced();
    let card = w.review_card(&["one"]);
    let lane = w.lane_adding("notes/x.md", S_A);
    let cand = w.join(&lane, S_LEAD);
    let event = w.verdict(&card, &cand, true, 0, S_LEAD, None);
    let got = result(&w.land_as(Some(S_LEAD), &card, &cand));
    assert_eq!(
        got["verdict"],
        json!({"event": event, "reviewer": S_LEAD, "tier": 0, "lander": S_LEAD}),
        "{got}"
    );
}

/// AC-19 (REQ-20): the lander's own PASS lands tier 0 and refuses with `lander_review` at tier 1
/// and 2 (also when the floor of the paths raises a stated tier 0), and a landing with no
/// session cannot be told from the reviewer at tier 1.
#[test]
fn ac54_the_lander_may_not_review_above_tier_zero() {
    let w = World::enforced();
    let card = w.review_card(&["one"]);
    let lane = w.lane_adding("src/a.rs", S_A);
    let cand = w.join(&lane, S_LEAD);
    w.verdict(&card, &cand, true, 1, S_LEAD, Some(w.good_receipt(&cand)));
    let before = w.dump();
    let out = w.land_as(Some(S_LEAD), &card, &cand);
    w.refuse(&out, "invalid_input", &["lander_review"], &before, &w.tip);

    // Stated tier 0 on a tier 2 path: judged at tier 2, so the lander's PASS does not count.
    let w = World::enforced();
    let card = w.review_card(&["one"]);
    let lane = w.lane_adding("src/gate_x.rs", S_A);
    let cand = w.join(&lane, S_LEAD);
    w.verdict(&card, &cand, true, 0, S_LEAD, None);
    let before = w.dump();
    let out = w.land_as(Some(S_LEAD), &card, &cand);
    w.refuse(&out, "invalid_input", &["lander_review"], &before, &w.tip);

    // An independent PASS at tier 1 with a landing that has no session of its own.
    let w = World::enforced();
    let card = w.review_card(&["one"]);
    let lane = w.lane_adding("src/a.rs", S_A);
    let cand = w.join(&lane, S_LEAD);
    w.verdict(&card, &cand, true, 1, S_AUD, Some(w.good_receipt(&cand)));
    let before = w.dump();
    let out = w.land_as(None, &card, &cand);
    w.refuse(&out, "invalid_input", &["lander_review"], &before, &w.tip);
}

/// AC-20 (REQ-10, REQ-21): the lander's clean merge commit is no authorship; a merge commit with
/// another tree is, and so is a merge that git cannot compute.
#[test]
fn ac55_a_clean_join_is_no_authorship_and_an_edited_merge_is() {
    // A clean join signed by the lander: the lander may review tier 0 (done by ac53_); an edited
    // merge commit of the lander counts as authored and refuses the lander's PASS.
    let w = World::enforced();
    let card = w.review_card(&["one"]);
    let lane = w.lane_adding("notes/x.md", S_A);
    let clean = w.join(&lane, S_LEAD);
    let edited_tree = {
        let with_extra = put_on_msg(&w.project, &clean, "notes/evil.md", "e\n", "tmp");
        git(
            &w.project,
            &["rev-parse", &format!("{with_extra}^{{tree}}")],
        )
    };
    let edited = git(
        &w.project,
        &[
            "commit-tree",
            &edited_tree,
            "-p",
            &w.tip,
            "-p",
            &lane,
            "-m",
            &signed("merge the lane and add a file", S_LEAD),
        ],
    );
    w.verdict(&card, &edited, true, 0, S_LEAD, None);
    let before = w.dump();
    let out = w.land_as(Some(S_LEAD), &card, &edited);
    w.refuse(&out, "invalid_input", &["self_verdict"], &before, &w.tip);

    // A merge that git cannot compute (the two sides change one file) counts as authored.
    let w = World::enforced();
    let card = w.review_card(&["one"]);
    let tip_b = put_on_msg(
        &w.project,
        &w.head,
        "notes/z.md",
        "B\n",
        "the tip changes z",
    );
    git(&w.project, &["update-ref", "refs/heads/integ", &tip_b]);
    let lane = put_on_msg(
        &w.project,
        &w.head,
        "notes/z.md",
        "A\n",
        &signed("the lane changes z", S_A),
    );
    let tree = git(&w.project, &["rev-parse", &format!("{lane}^{{tree}}")]);
    let cand = git(
        &w.project,
        &[
            "commit-tree",
            &tree,
            "-p",
            &tip_b,
            "-p",
            &lane,
            "-m",
            &signed("merge by hand", S_LEAD),
        ],
    );
    w.verdict(&card, &cand, true, 0, S_LEAD, None);
    let before = w.dump();
    let out = w.land_as(Some(S_LEAD), &card, &cand);
    w.refuse(&out, "invalid_input", &["self_verdict"], &before, &tip_b);
}

/// AC-21 (REQ-22, REQ-23): the profile's `land_verdict` picks the mode, none means observe, and
/// another word makes the profile unreadable for the landing.
#[test]
fn ac56_the_profile_names_the_mode_and_another_word_is_refused() {
    let w = World::new();
    let card = w.review_card(&["one"]);
    let lane = w.lane_adding("notes/x.md", S_A);
    let cand = w.join(&lane, S_LEAD);

    w.set_profile(BAD_MODE_PROFILE);
    let before = w.dump();
    let out = w.land_as(Some(S_LEAD), &card, &cand);
    assert_eq!(out.status.code(), Some(1));
    assert_eq!(w.dump(), before);
    assert_eq!(w.integ(), w.tip);

    w.set_profile(ENFORCE_PROFILE);
    let before = w.dump();
    let out = w.land_as(Some(S_LEAD), &card, &cand);
    w.refuse(&out, "invalid_input", &["no_verdict"], &before, &w.tip);

    // No `land_verdict` key at all: observe. A slice lands and reports the refusal; a landing
    // that would close the card is refused as `no_pass` in every mode.
    w.set_profile(MFLOW_PROFILE);
    let before = w.dump();
    let out = w.land_as(Some(S_LEAD), &card, &cand);
    w.refuse(&out, "invalid_input", &["no_pass"], &before, &w.tip);
    let got = result(&w.land_plain(Some(S_LEAD), &card, &cand, &["--keep-open"]));
    assert_eq!(got["would_refuse"]["code"], json!("no_verdict"), "{got}");
    assert_eq!(got["closed"], json!(false), "{got}");
}

// ---- AC-35 to AC-38: a landing cleans the card's worktrees (SPEC-20261123-worktree-slots AC-6) ---

impl World {
    /// A linked worktree of the project, detached at the card's change (which the tip contains),
    /// in a folder of its own outside the project.
    fn linked_worktree(&self, name: &str) -> PathBuf {
        let at = fs::canonicalize(self.dir.path()).unwrap().join(name);
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

    /// A task card with `criteria`, its own scope path, taken by `S_A` in `worktree` (on `branch`
    /// when given); returns its name.
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

    /// The cleanup clock three days ahead, so every change in the fixture is old.
    fn clock_three_days_on(&self) {
        let now = std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap()
            .as_secs();
        let mut env = self.extra_env.borrow_mut();
        env.push((
            moai_cli::worktree_cleanup::NOW_SEAM.to_string(),
            (now + 3 * 86_400).to_string(),
        ));
        // The cleanup's process check runs `lsof`, which it finds on the caller's PATH.
        env.push((
            "PATH".to_string(),
            std::env::var("PATH").unwrap_or_default(),
        ));
    }

    /// Sends the card to review and lands a candidate with the given `--met` options.
    fn lands(&self, card: &str, met: &[&str]) -> Value {
        result(&self.card(Some(S_A), &["review", card, "--generation", "1"]));
        let candidate = self.candidate();
        let mut args = vec!["--head", candidate.as_str()];
        for m in met {
            args.extend(["--met", m]);
        }
        args.extend(["--by", "leader"]);
        result(&self.land_pass(card, &args))
    }

    fn registered(&self, dir: &Path) -> bool {
        git(&self.project, &["worktree", "list", "--porcelain"])
            .contains(dir.file_name().unwrap().to_str().unwrap())
    }
}

/// AC-35: after a landing closes the card, the cleanup plan lists the card's worktree and the
/// apply removes it on the plan's digest; another card's worktree stays.
#[test]
fn ac35_a_landing_removes_the_closed_cards_merged_worktree_on_the_plans_digest() {
    let w = World::new();
    let mine = w.linked_worktree("wt-mine");
    let other = w.linked_worktree("wt-other");
    w.clock_three_days_on();
    let card = w.card_in("apps/a/", &mine, None, &["one"]);
    let _held = w.card_in("apps/b/", &other, None, &["two"]);
    let got = w.lands(&card, &["1=the test"]);
    assert_eq!(got["closed"], json!(true), "{got}");
    let cleanup = &got["cleanup"];
    assert_eq!(cleanup["trigger"], json!("landing"), "{got}");
    assert_eq!(cleanup["error"], json!(null), "{got}");
    assert_eq!(cleanup["applied"], json!(true), "{got}");
    let digest = cleanup["plan_digest"].as_str().unwrap();
    assert!(
        digest.starts_with("sha256:") && digest.len() == "sha256:".len() + 64,
        "{got}"
    );
    let listed = cleanup["worktrees"].as_array().unwrap();
    assert_eq!(listed.len(), 1, "only the closed card's worktree: {got}");
    assert_eq!(listed[0]["action"], json!("remove"), "{got}");
    assert_eq!(listed[0]["status"], json!("removed"), "{got}");
    assert!(!mine.exists(), "the merged worktree is still there");
    assert!(!w.registered(&mine));
    assert!(
        other.is_dir() && w.registered(&other),
        "another card's worktree was touched"
    );
    assert_eq!(
        w.integ(),
        got["to_tip"].as_str().unwrap(),
        "the branch moved with the landing"
    );
}

/// AC-36: a worktree under `.claude/worktrees` stays preserved: the plan names the reason, the
/// apply acts on nothing and the folder is still a registered worktree.
#[test]
fn ac36_a_worktree_under_the_preserved_place_is_kept() {
    let w = World::new();
    w.clock_three_days_on();
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
    let card = w.card_in("apps/a/", &under, None, &["one"]);
    let got = w.lands(&card, &["1=the test"]);
    assert_eq!(got["closed"], json!(true), "{got}");
    let listed = got["cleanup"]["worktrees"].as_array().unwrap();
    assert_eq!(listed[0]["action"], json!("skip"), "{got}");
    assert!(
        listed[0]["reasons"]
            .to_string()
            .contains("preserved_location"),
        "{got}"
    );
    assert_eq!(got["cleanup"]["applied"], json!(false), "{got}");
    assert!(
        under.is_dir() && w.registered(&under),
        "a preserved worktree was removed"
    );
}

/// AC-36 (second half): a card that stays open because it was landed with `--keep-open` is not
/// cleaned.
#[test]
fn ac36_a_card_that_stays_open_is_not_cleaned() {
    let w = World::new();
    w.clock_three_days_on();
    let kept = w.linked_worktree("wt-open");
    let open = w.card_in("apps/c/", &kept, None, &["a", "b"]);
    result(&w.card(Some(S_A), &["review", &open, "--generation", "1"]));
    let candidate = w.candidate();
    let got = result(&w.land_lead(
        &open,
        &["--head", &candidate, "--keep-open", "--by", "leader"],
    ));
    assert_eq!(got["closed"], json!(false), "{got}");
    assert_eq!(
        got["cleanup"],
        json!(null),
        "an open card was cleaned: {got}"
    );
    assert!(kept.is_dir() && w.registered(&kept));
}

/// AC-37: the queue the hook fills is read at a landing. A queued branch that is gone cleans the
/// worktrees of the cards that held it; a queued branch that still exists only loses its line;
/// the queue file is gone afterwards.
#[test]
fn ac37_a_landing_reads_the_queue_and_cleans_for_a_branch_that_is_gone() {
    let w = World::new();
    w.clock_three_days_on();
    let mine = w.linked_worktree("wt-mine");
    let old = w.linked_worktree("wt-old");
    let keeps = w.linked_worktree("wt-keeps");
    let card = w.card_in("apps/a/", &mine, None, &["one"]);
    // Two earlier cards on branches; both gave their reservation back.
    git(&w.project, &["branch", "card/live"]);
    for (scope, wt, branch) in [
        ("apps/d/", &old, "card/gone"),
        ("apps/e/", &keeps, "card/live"),
    ] {
        let name = w.card_in(scope, wt, Some(branch), &["x"]);
        result(&w.card(Some(S_A), &["release", &name, "--generation", "1"]));
    }
    let queue = w.common().join("moai/worktree-cleanup-queue");
    fs::create_dir_all(queue.parent().unwrap()).unwrap();
    fs::write(
        &queue,
        "refs/heads/card/gone\nrefs/heads/card/live\nrefs/heads/card/gone\n",
    )
    .unwrap();

    let got = w.lands(&card, &["1=the test"]);
    let cleanup = &got["cleanup"];
    assert_eq!(cleanup["error"], json!(null), "{got}");
    assert_eq!(
        cleanup["queue"]["gone"],
        json!(["refs/heads/card/gone"]),
        "{got}"
    );
    assert_eq!(
        cleanup["queue"]["still_there"],
        json!(["refs/heads/card/live"]),
        "{got}"
    );
    assert!(!old.exists(), "the gone branch's worktree was not cleaned");
    assert!(
        keeps.is_dir(),
        "a branch that still exists had its worktree removed"
    );
    assert!(!mine.exists());
    assert!(!queue.exists(), "the queue was not used up");
}

/// AC-38: a cleanup that cannot run never undoes the landing: the card is closed, the branch
/// moved, the answer carries the cleanup's `error`, and the queue keeps its lines.
#[cfg(unix)]
#[test]
fn ac38_a_cleanup_that_cannot_run_is_reported_and_the_landing_stands() {
    use std::os::unix::fs::PermissionsExt;
    let w = World::new();
    let mine = w.linked_worktree("wt-mine");
    let card = w.card_in("apps/a/", &mine, None, &["one"]);
    // A queue nobody can read.
    let queue = w.common().join("moai/worktree-cleanup-queue");
    fs::create_dir_all(queue.parent().unwrap()).unwrap();
    fs::write(&queue, "refs/heads/card/gone\n").unwrap();
    fs::set_permissions(&queue, fs::Permissions::from_mode(0o000)).unwrap();
    let got = w.lands(&card, &["1=the test"]);
    fs::set_permissions(&queue, fs::Permissions::from_mode(0o600)).unwrap();
    assert_eq!(got["closed"], json!(true), "{got}");
    assert_eq!(got["moved"], json!(true), "{got}");
    assert_eq!(got["cleanup"]["applied"], json!(false), "{got}");
    assert!(got["cleanup"]["error"].is_string(), "{got}");
    assert!(mine.is_dir(), "a cleanup that failed removed a worktree");
    assert_eq!(
        fs::read_to_string(&queue).unwrap(),
        "refs/heads/card/gone\n",
        "the queue lost its line"
    );
}

/// AC-39 (SPEC-20261123-landing-cleanup-real-clock AC-1, AC-4): on the real clock, with no clock
/// seam, a landing removes the closed card's clean merged worktree and counts the removal in
/// `writes`. Every worktree in this test is seconds old.
#[test]
fn ac39_a_landing_on_the_real_clock_removes_the_closed_cards_clean_worktree() {
    let w = World::new();
    let mine = w.linked_worktree("wt-mine");
    let card = w.card_in("apps/a/", &mine, None, &["one"]);
    let got = w.lands(&card, &["1=the test"]);
    assert_eq!(got["closed"], json!(true), "{got}");
    assert_eq!(got["cleanup"]["error"], json!(null), "{got}");
    assert_eq!(got["cleanup"]["applied"], json!(true), "{got}");
    assert_eq!(
        got["cleanup"]["worktrees"][0]["status"],
        json!("removed"),
        "{got}"
    );
    assert!(!mine.exists() && !w.registered(&mine), "{got}");
    assert_eq!(got["writes"], json!(true), "{got}");
}

/// AC-39 (third part; SPEC-20261123-landing-cleanup-real-clock AC-1, D-5): `card land --home DIR` reads the
/// claims of the ledger it landed in, not of the process's own `HOME`. A second live card holds the
/// landed card's worktree; the process `HOME` is an empty folder, `--home` is the ledger's. The
/// worktree stays, whatever `HOME` the process has; a worktree nobody else holds is removed.
#[test]
fn ac39_a_landing_reads_the_other_cards_claim_from_the_given_home() {
    let w = World::new();
    let shared = w.linked_worktree("wt-shared");
    let card = w.card_in("apps/a/", &shared, None, &["one"]);
    let _second = w.card_in("apps/b/", &shared, None, &["two"]);
    result(&w.card(Some(S_A), &["review", &card, "--generation", "1"]));
    let candidate = w.candidate();
    w.pass_of(&card, &candidate);
    let process_home = TempHome::new();
    let ledger_home = w.home.path().to_str().unwrap().to_string();
    let mut cmd = w.home.command(&[
        "card",
        "land",
        &card,
        "--head",
        &candidate,
        "--met",
        "1=the test",
        "--by",
        "leader",
        "--project",
        &w.p,
        "--home",
        &ledger_home,
    ]);
    cmd.env("HOME", process_home.path())
        .env("USERPROFILE", process_home.path())
        .env(SESSION_VAR, S_LEAD)
        .env("MOAI_GIT_BIN", &*w.git_bin.borrow())
        // The cleanup's process check runs `lsof`, which it finds on the caller's PATH.
        .env("PATH", std::env::var("PATH").unwrap_or_default())
        .stdin(Stdio::null());
    let got = result(&cmd.output().expect("run moai"));
    assert_eq!(got["closed"], json!(true), "{got}");
    assert_eq!(got["cleanup"]["error"], json!(null), "{got}");
    assert_eq!(got["cleanup"]["applied"], json!(false), "{got}");
    assert_eq!(
        got["cleanup"]["worktrees"][0]["action"],
        json!("skip"),
        "{got}"
    );
    assert!(
        got["cleanup"]["worktrees"][0]["reasons"]
            .to_string()
            .contains("live_claim"),
        "{got}"
    );
    assert!(
        shared.is_dir() && w.registered(&shared),
        "a worktree another live card holds was removed: {got}"
    );
}

/// AC-39 (second half): on the real clock a landing keeps a worktree with a tracked change and one
/// whose `target` links outside; what lies outside, and the changed file, are as they were.
#[cfg(unix)]
#[test]
fn ac39_a_landing_on_the_real_clock_keeps_a_dirty_and_a_linked_target_worktree() {
    // One landing per fixture: the candidate is made on the fixture's own tip.
    for case in ["dirty", "linked"] {
        let w = World::new();
        let outside = fs::canonicalize(w.dir.path()).unwrap().join("outside");
        fs::create_dir_all(&outside).unwrap();
        fs::write(outside.join("keep.txt"), "do not touch\n").unwrap();
        let wt = w.linked_worktree("wt-case");
        let mut tracked = None;
        if case == "dirty" {
            let listed = git(&wt, &["ls-files"]);
            let file = wt.join(listed.lines().next().expect("a tracked file"));
            fs::write(&file, "changed after the card\n").unwrap();
            tracked = Some(file);
        } else {
            std::os::unix::fs::symlink(&outside, wt.join("target")).unwrap();
        }
        let card = w.card_in("apps/a/", &wt, None, &["one"]);
        let got = w.lands(&card, &["1=the test"]);
        assert_eq!(got["closed"], json!(true), "{case}: {got}");
        assert_eq!(got["cleanup"]["applied"], json!(false), "{case}: {got}");
        assert_eq!(
            got["cleanup"]["worktrees"][0]["action"],
            json!("skip"),
            "{case}: {got}"
        );
        assert!(wt.is_dir() && w.registered(&wt), "{case}: it was removed");
        assert_eq!(
            fs::read_to_string(outside.join("keep.txt")).unwrap(),
            "do not touch\n",
            "{case}"
        );
        if let Some(file) = tracked {
            assert_eq!(
                fs::read_to_string(&file).unwrap(),
                "changed after the card\n"
            );
        } else {
            assert!(fs::symlink_metadata(wt.join("target"))
                .unwrap()
                .file_type()
                .is_symlink());
        }
    }
}

/// AC-22 (REQ-8, REQ-24): a call killed after the start row and before the move, then a FAIL by
/// another reviewer: the call run again refuses with `verdict_fail`, the branch stays at the tip
/// and no landing is written; the same reviewer's later PASS lands the head once, and the landing
/// event and the answer name that newest PASS and its reviewer, not the PASS the start row of the
/// killed call named (REQ-13, AC-17).
#[test]
fn ac57_a_fail_recorded_after_a_kill_before_the_move_refuses_the_rerun() {
    let w = World::enforced();
    let card = w.review_card(&["one"]);
    let lane = w.lane_adding("notes/x.md", S_A);
    let cand = w.join(&lane, S_LEAD);
    let first = w.verdict(&card, &cand, true, 0, S_AUD, None);
    // The script keeps running after it kills its parent, so it exits too and the branch stays.
    w.wrap_update_ref("git-kills-the-lander", "kill -9 $PPID; exit 9", "");
    let out = w.land_as(Some(S_LEAD), &card, &cand);
    assert_ne!(out.status.code(), Some(0), "the lander was not killed");
    assert_eq!(w.integ(), w.tip, "the kill came after the move");
    assert_eq!(w.events(&card, "card.land_started"), 1);
    assert_eq!(w.events(&card, "card.landed"), 0);

    // Another reviewer fails the head while the branch has not moved.
    w.verdict(&card, &cand, false, 0, S_OTHER, None);
    *w.git_bin.borrow_mut() = git_executable();
    let before = w.dump();
    let out = w.land_as(Some(S_LEAD), &card, &cand);
    w.refuse(&out, "invalid_input", &["verdict_fail"], &before, &w.tip);
    assert_eq!(
        w.events(&card, "card.land_started"),
        1,
        "a second start row"
    );
    assert_eq!(w.events(&card, "card.landed"), 0);

    // The reviewer that failed it passes it later: the landing goes through once.
    let latest = w.verdict(&card, &cand, true, 0, S_OTHER, None);
    assert_ne!(latest, first);
    let got = result(&w.land_as(Some(S_LEAD), &card, &cand));
    assert_eq!(got["moved"], json!(true), "{got}");
    assert_eq!(got["closed"], json!(true), "{got}");
    assert_eq!(w.integ(), cand);
    assert_eq!(
        w.events(&card, "card.land_started"),
        1,
        "a second start row"
    );
    assert_eq!(w.events(&card, "card.landed"), 1);
    // The call relied on the newest PASS (the reviewer that failed the head and then passed it),
    // and the record of the landing says so: the answer and the landing event name the same
    // verdict, and it is not the PASS the killed call's start row names.
    let newest = json!({"event": latest, "reviewer": S_OTHER, "tier": 0, "lander": S_LEAD});
    assert_eq!(got["verdict"], newest, "{got}");
    assert_eq!(w.payload(&card, "card.landed")["verdict"], newest);
    assert_eq!(
        w.payload(&card, "card.land_started")["verdict"]["event"],
        json!(first),
        "the start row of the killed call is not rewritten"
    );
}

/// AC-23 (REQ-9): a verdict event whose head differs from the candidate in the last hex digit only
/// is a verdict for another head; the verdict of the full head lands it.
#[test]
fn ac58_a_verdict_names_the_whole_head_and_not_a_prefix_of_it() {
    let w = World::enforced();
    let card = w.review_card(&["one"]);
    let lane = w.lane_adding("notes/x.md", S_A);
    let cand = w.join(&lane, S_LEAD);
    let last = cand.chars().last().unwrap();
    let other_last = if last == '0' { '1' } else { '0' };
    let near = format!("{}{other_last}", &cand[..cand.len() - 1]);
    assert_ne!(near, cand);
    w.verdict(&card, &near, true, 0, S_AUD, None);
    let before = w.dump();
    let out = w.land_as(Some(S_LEAD), &card, &cand);
    let message = w.refuse(
        &out,
        "invalid_input",
        &["verdict_other_head"],
        &before,
        &w.tip,
    );
    assert!(
        message.contains(&near),
        "names the verdict's own head: {message}"
    );
    w.verdict(&card, &cand, true, 0, S_AUD, None);
    let got = result(&w.land_as(Some(S_LEAD), &card, &cand));
    assert_eq!(got["moved"], json!(true), "{got}");
}

/// AC-16 (REQ-14): in observe mode a git that fails inside the verdict check only is reported as
/// `would_refuse` with the code `check_unreadable` and the head lands; in enforce mode the same
/// call refuses.
#[test]
fn ac59_a_check_that_cannot_read_git_is_reported_in_observe_and_refuses_in_enforce() {
    for observe in [true, false] {
        let w = World::new();
        w.set_profile(if observe {
            OBSERVE_PROFILE
        } else {
            ENFORCE_PROFILE
        });
        let card = w.review_card(&["one"]);
        let lane = w.lane_adding("notes/x.md", S_A);
        let cand = w.join(&lane, S_LEAD);
        w.verdict(&card, &cand, true, 0, S_AUD, None);
        // Only the `git log` of the verdict check (its format argument) fails.
        let wrapper = w.dir.path().join("git-fails-in-the-check");
        common::write_executable(
            &wrapper,
            &format!(
                "#!/bin/sh\nfor a in \"$@\"; do\n  case \"$a\" in --format=%H%x1f%P%x1f%B%x1e) echo 'fatal: the check cannot read' >&2; exit 128;; esac\ndone\nexec '{git}' \"$@\"\n",
                git = git_executable().display(),
            ),
        );
        *w.git_bin.borrow_mut() = wrapper;
        let before = w.dump();
        let out = w.land_as(Some(S_LEAD), &card, &cand);
        if observe {
            // A landing that would close the card cannot rely on a PASS it could not read; a
            // slice lands and reports the finding.
            w.refuse(&out, "io", &["check_unreadable"], &before, &w.tip);
            let got = result(&w.land_plain(Some(S_LEAD), &card, &cand, &["--keep-open"]));
            assert_eq!(got["moved"], json!(true), "{got}");
            assert_eq!(got["closed"], json!(false), "{got}");
            assert_eq!(
                got["would_refuse"]["code"],
                json!("check_unreadable"),
                "{got}"
            );
            assert!(got.get("verdict").is_none(), "{got}");
        } else {
            w.refuse(&out, "invalid_input", &["cannot read"], &before, &w.tip);
        }
    }
}

// ---- SPEC-20261124-land-closes-card: a landing on a recorded PASS closes the card ---------------

impl World {
    /// The base a reviewer names for `head`: the integration tip while the head stands on it, the
    /// head's first parent once the branch holds it (the range the head added).
    fn pass_base(&self, head: &str) -> String {
        let inside = common::git::command()
            .arg("-C")
            .arg(&self.project)
            .args(["merge-base", "--is-ancestor", head, &self.tip])
            .status()
            .map(|s| s.success())
            .unwrap_or(false);
        let parent = common::git::command()
            .arg("-C")
            .arg(&self.project)
            .args(["rev-parse", "--verify", "-q", &format!("{head}^")])
            .output()
            .ok()
            .filter(|o| o.status.success())
            .map(|o| String::from_utf8_lossy(&o.stdout).trim().to_string());
        match (inside, parent) {
            (true, Some(parent)) => parent,
            _ => self.tip.clone(),
        }
    }

    /// `card land` for the head named in `extra` after an independent tier 0 PASS of that head is
    /// recorded: the fixture of every test that lands a card and is not about the verdict.
    fn land_pass(&self, card: &str, extra: &[&str]) -> Output {
        if let Some(i) = extra.iter().position(|a| *a == "--head") {
            self.pass_of(card, extra[i + 1]);
        }
        self.land_lead(card, extra)
    }

    /// `card land` as the lander session `S_LEAD`, with no verdict recorded by the call: the
    /// test of a refusal records the PASS first and takes its ledger snapshot after.
    fn land_lead(&self, card: &str, extra: &[&str]) -> Output {
        let mut all = vec!["land", card];
        all.extend_from_slice(extra);
        self.card(Some(S_LEAD), &all)
    }

    /// An independent PASS of `head` with a green receipt: judged at tier 1 whatever the head
    /// changes, so it holds for a head that changes code as well as for one that changes notes.
    fn pass_of(&self, card: &str, head: &str) -> i64 {
        self.verdict(card, head, true, 1, S_AUD, Some(self.good_receipt(head)))
    }

    /// A review card with one open criterion per text and the tier 0 candidate of a lane that
    /// `S_A` wrote and `S_LEAD` joined; returns the card and the candidate.
    fn signed_card(&self, criteria: &[&str]) -> (String, String) {
        let card = self.review_card(criteria);
        let lane = self.lane_adding("notes/x.md", S_A);
        let cand = self.join(&lane, S_LEAD);
        (card, cand)
    }

    /// `card land` as `session` with no `--met` of its own unless `extra` names one.
    fn land_plain(&self, session: Option<&str>, card: &str, head: &str, extra: &[&str]) -> Output {
        let mut all = vec!["land", card, "--head", head];
        all.extend_from_slice(extra);
        all.extend_from_slice(&["--by", "leader"]);
        self.card(session, &all)
    }

    /// The status and the evidence of one criterion, as `status|evidence`.
    fn criterion_row(&self, card: &str, seq: i64) -> String {
        self.text(&format!(
            "SELECT status || '|' || COALESCE(evidence_ref, '') FROM card_ac \
             WHERE card_id = {} AND seq = {seq}",
            number(card)
        ))
    }

    fn state_of(&self, card: &str) -> String {
        self.text(&format!(
            "SELECT state FROM card WHERE card_id = {}",
            number(card)
        ))
    }
}

/// The evidence line the landing writes for a criterion it marks met.
fn pass_line(event: i64, head: &str) -> String {
    format!("verdict event {event} PASS head {}", &head[..12])
}

/// AC-1 (REQ-1, REQ-3, REQ-11): a head with an independent PASS lands with no `--met` of its own,
/// every open criterion becomes met with the verdict line and the card is done.
#[test]
fn ac_closes_a_head_with_a_pass_marks_every_open_criterion_met() {
    let w = World::new();
    let (card, cand) = w.signed_card(&["one", "two", "three"]);
    let event = w.verdict(&card, &cand, true, 0, S_AUD, None);
    let got = result(&w.land_plain(Some(S_LEAD), &card, &cand, &[]));
    assert_eq!(got["moved"], json!(true), "{got}");
    assert_eq!(got["closed"], json!(true), "{got}");
    assert_eq!(got["state"], json!("done"), "{got}");
    assert_eq!(got["open_criteria"], json!([]), "{got}");
    let line = pass_line(event, &cand);
    for seq in 1..=3 {
        assert_eq!(w.criterion_row(&card, seq), format!("met|{line}"));
    }
    let met = got["met"].as_array().unwrap();
    assert_eq!(met.len(), 3, "{got}");
    assert!(met.iter().all(|m| m["evidence"] == json!(line)), "{got}");
    assert_eq!(w.state_of(&card), "done");
    assert_eq!(w.events(&card, "card.finish"), 1);
}

/// AC-2 (REQ-2): an explicit `--met` keeps its own reference and the verdict line fills the rest.
#[test]
fn ac_keeps_met_a_given_reference_and_fills_the_others_from_the_verdict() {
    let w = World::new();
    let (card, cand) = w.signed_card(&["one", "two", "three"]);
    let event = w.verdict(&card, &cand, true, 0, S_AUD, None);
    let got = result(&w.land_plain(Some(S_LEAD), &card, &cand, &["--met", "2=my own reference"]));
    assert_eq!(got["closed"], json!(true), "{got}");
    let line = pass_line(event, &cand);
    assert_eq!(w.criterion_row(&card, 1), format!("met|{line}"));
    assert_eq!(w.criterion_row(&card, 2), "met|my own reference");
    assert_eq!(w.criterion_row(&card, 3), format!("met|{line}"));
}

/// AC-3 (REQ-4): without a relied PASS and without `--keep-open` the landing refuses in observe
/// mode too: `no_pass` for no verdict and for a PASS of another head, the check's own code for a
/// FAIL; an explicit `--met` for every criterion does not bypass it.
#[test]
fn ac_no_pass_refuses_a_closing_landing_in_observe_mode() {
    let w = World::new();
    let (card, cand) = w.signed_card(&["one"]);

    let before = w.dump();
    let out = w.land_plain(Some(S_LEAD), &card, &cand, &[]);
    w.refuse(&out, "invalid_input", &["no_pass"], &before, &w.tip);
    let out = w.land_plain(Some(S_LEAD), &card, &cand, &["--met", "1=my reference"]);
    w.refuse(&out, "invalid_input", &["no_pass"], &before, &w.tip);

    // A PASS recorded for another head only.
    w.verdict(&card, &w.head, true, 0, S_AUD, None);
    let before = w.dump();
    let out = w.land_plain(Some(S_LEAD), &card, &cand, &[]);
    w.refuse(&out, "invalid_input", &["no_pass"], &before, &w.tip);

    // A FAIL on the head.
    w.verdict(&card, &cand, false, 0, S_AUD, None);
    let before = w.dump();
    let out = w.land_plain(Some(S_LEAD), &card, &cand, &[]);
    w.refuse(&out, "invalid_input", &["verdict_fail"], &before, &w.tip);
    assert_eq!(w.state_of(&card), "review");
}

/// AC-12 (REQ-4): the refusal comes before the start row and before the intent record.
#[test]
fn ac_no_pass_writes_nothing_before_the_start_row_or_the_intent_record() {
    let w = World::new();
    let (card, cand) = w.signed_card(&["one"]);
    let before = w.dump();
    let out = w.land_plain(Some(S_LEAD), &card, &cand, &[]);
    w.refuse(&out, "invalid_input", &["no_pass"], &before, &w.tip);
    assert_eq!(w.events(&card, "card.land_started"), 0);
    assert_eq!(w.events(&card, "card.landed"), 0);
    assert!(
        !w.common().join("moai/land-intent").exists(),
        "the refused call wrote the intent record"
    );
}

/// AC-4 (REQ-5, REQ-8): `--keep-open` lands a head with no verdict in observe mode, closes
/// nothing, names every open criterion and marks the landing event.
#[test]
fn ac_keep_open_lands_a_slice_and_leaves_every_criterion_open() {
    let w = World::new();
    let (card, cand) = w.signed_card(&["one", "two"]);
    let got = result(&w.land_plain(Some(S_LEAD), &card, &cand, &["--keep-open"]));
    assert_eq!(got["moved"], json!(true), "{got}");
    assert_eq!(got["closed"], json!(false), "{got}");
    assert_eq!(got["open_criteria"], json!([1, 2]), "{got}");
    assert_eq!(got["would_refuse"]["code"], json!("no_verdict"), "{got}");
    assert_eq!(w.integ(), cand);
    assert_eq!(w.events(&card, "card.landed"), 1);
    assert_eq!(w.events(&card, "card.finish"), 0);
    assert_eq!(w.payload(&card, "card.landed")["keep_open"], json!(true));
    assert_eq!(w.state_of(&card), "review");
    assert_eq!(w.criterion_row(&card, 1), "open|");
    assert_eq!(w.criterion_row(&card, 2), "open|");
}

/// AC-5 (REQ-6): in enforce mode `--keep-open` is no way round the verdict check; with an
/// independent PASS the same call lands.
#[test]
fn ac_keep_enforce_applies_the_verdict_check_to_a_slice() {
    let w = World::enforced();
    let (card, cand) = w.signed_card(&["one", "two"]);
    let before = w.dump();
    let out = w.land_plain(Some(S_LEAD), &card, &cand, &["--keep-open"]);
    w.refuse(&out, "invalid_input", &["no_verdict"], &before, &w.tip);

    w.verdict(&card, &cand, true, 0, S_AUD, None);
    let got = result(&w.land_plain(Some(S_LEAD), &card, &cand, &["--keep-open"]));
    assert_eq!(got["moved"], json!(true), "{got}");
    assert_eq!(got["closed"], json!(false), "{got}");
    assert_eq!(got["open_criteria"], json!([1, 2]), "{got}");
    assert_eq!(w.payload(&card, "card.landed")["keep_open"], json!(true));
}

/// AC-6 (REQ-7): `--keep-open` with a `--met` option and `--keep-open` on a card with no open
/// criterion each refuse and write nothing.
#[test]
fn ac_keep_invalid_refuses_a_met_option_and_a_card_with_nothing_open() {
    let w = World::new();
    let (card, cand) = w.signed_card(&["one", "two"]);
    let before = w.dump();
    let out = w.land_plain(Some(S_LEAD), &card, &cand, &["--keep-open", "--met", "1=x"]);
    w.refuse(&out, "invalid_input", &["--keep-open"], &before, &w.tip);

    let w = World::new();
    let (card, cand) = w.signed_card(&["one"]);
    result(&w.card(
        Some(S_A),
        &[
            "criterion",
            &card,
            "1",
            "--generation",
            "1",
            "--met",
            "--evidence",
            "the test",
        ],
    ));
    let before = w.dump();
    let out = w.land_plain(Some(S_LEAD), &card, &cand, &["--keep-open"]);
    w.refuse(&out, "invalid_input", &["--keep-open"], &before, &w.tip);
}

/// AC-8 (REQ-8): a slice landed with `--keep-open` closes when the same head is landed again
/// with an independent PASS; the landing is recorded once.
#[test]
fn ac_slice_then_close_lands_once_and_closes_on_the_later_pass() {
    let w = World::new();
    let (card, cand) = w.signed_card(&["one", "two"]);
    result(&w.land_plain(Some(S_LEAD), &card, &cand, &["--keep-open"]));
    assert_eq!(w.state_of(&card), "review");

    let event = w.verdict(&card, &cand, true, 0, S_AUD, None);
    let got = result(&w.land_plain(Some(S_LEAD), &card, &cand, &[]));
    assert_eq!(got["moved"], json!(false), "{got}");
    assert_eq!(got["closed"], json!(true), "{got}");
    assert_eq!(w.events(&card, "card.landed"), 1);
    assert_eq!(w.events(&card, "card.finish"), 1);
    let line = pass_line(event, &cand);
    assert_eq!(w.criterion_row(&card, 1), format!("met|{line}"));
    assert_eq!(w.criterion_row(&card, 2), format!("met|{line}"));
    assert_eq!(w.state_of(&card), "done");
}

/// AC-9 (REQ-9): a landing killed after the move is closed by the same call run again with no
/// `--keep-open`, on the verdict the start row names, also after a later FAIL.
#[test]
fn ac_rerun_closes_on_the_verdict_the_start_row_names() {
    let w = World::enforced();
    let (card, cand) = w.signed_card(&["one", "two"]);
    let event = w.verdict(&card, &cand, true, 0, S_AUD, None);
    w.wrap_update_ref("git-kills-the-lander", "", "kill -9 $PPID");
    let out = w.land_plain(Some(S_LEAD), &card, &cand, &[]);
    assert_ne!(out.status.code(), Some(0), "the lander was not killed");
    assert_eq!(w.integ(), cand);
    assert_eq!(w.events(&card, "card.landed"), 0);

    w.verdict(&card, &cand, false, 0, S_OTHER, None);
    *w.git_bin.borrow_mut() = git_executable();
    let rerun = result(&w.land_plain(Some(S_LEAD), &card, &cand, &[]));
    assert_eq!(rerun["moved"], json!(false), "{rerun}");
    assert_eq!(rerun["closed"], json!(true), "{rerun}");
    assert_eq!(rerun["verdict"]["event"], json!(event), "{rerun}");
    let line = pass_line(event, &cand);
    assert_eq!(w.criterion_row(&card, 1), format!("met|{line}"));
    assert_eq!(w.criterion_row(&card, 2), format!("met|{line}"));
}

/// AC-10 (REQ-1): a card that another session holds in `active` with a live lease keeps today's
/// refusal of the finish check, with a PASS recorded.
#[test]
fn ac_held_keeps_the_refusal_for_a_live_lease_of_another_session() {
    let w = World::new();
    let card = w.ready_card(&["one"]);
    result(&w.card(Some(S_A), &["take", &card]));
    let lane = w.lane_adding("notes/x.md", S_A);
    let cand = w.join(&lane, S_LEAD);
    w.verdict(&card, &cand, true, 0, S_AUD, None);
    let before = w.dump();
    let out = w.land_plain(Some(S_LEAD), &card, &cand, &[]);
    w.refuse(&out, "conflict", &["held by"], &before, &w.tip);
}

// ---- a checked-out integration branch (SPEC-20261125-landing-moves-checkout) -----------------------

/// A review card, its lane and the join the leader lands, with an independent PASS recorded, and
/// the integration branch checked out in the project folder: the state the checkout tests start in.
fn checkout_fixture() -> (World, String, String) {
    let w = World::enforced();
    let card = w.review_card(&["one"]);
    let lane = w.lane_adding("src/a.rs", S_A);
    let cand = w.join(&lane, S_LEAD);
    w.verdict(&card, &cand, true, 1, S_AUD, Some(w.good_receipt(&cand)));
    git(&w.project, &["checkout", "-q", "integ"]);
    (w, card, cand)
}

/// AC-1 (REQ-1): a clean tree that has the integration branch checked out moves with the branch.
#[test]
fn checkout_a_clean_checked_out_branch_moves_with_its_tree() {
    let (w, card, cand) = checkout_fixture();
    let got = result(&w.land_as(Some(S_LEAD), &card, &cand));
    assert_eq!(got["moved"], json!(true), "{got}");
    assert_eq!(w.integ(), cand);
    assert_eq!(
        git(&w.project, &["rev-parse", "HEAD"]),
        cand,
        "the tree's HEAD"
    );
    assert!(
        w.project.join("src/a.rs").is_file(),
        "the file of the head is in the tree"
    );
    assert_eq!(
        git(
            &w.project,
            &["status", "--porcelain", "--untracked-files=no"]
        ),
        "",
        "the tree is clean at the head"
    );
}

/// AC-5 (REQ-6): a landing without `--keep-open` leaves the card done and no criterion open.
#[test]
fn checkout_a_landing_closes_the_card() {
    let (w, card, cand) = checkout_fixture();
    result(&w.land_as(Some(S_LEAD), &card, &cand));
    assert_eq!(w.state_of(&card), "done");
    assert_eq!(
        w.text(&format!(
            "SELECT count(*) FROM card_ac WHERE card_id = {} AND status = 'open'",
            number(&card)
        )),
        "0"
    );
}

/// AC-2 (REQ-2, REQ-3): a tracked change in the tree, and a second worktree that has the same
/// branch checked out, each refuse with nothing written: the ledger, the branch and the tree stay.
#[test]
fn checkout_a_dirty_or_moved_or_shared_checkout_refuses_and_writes_nothing() {
    let (w, card, cand) = checkout_fixture();
    fs::write(w.project.join("a.txt"), "changed\n").unwrap();
    let before = w.dump();
    let out = w.land_as(Some(S_LEAD), &card, &cand);
    let message = w.refuse(&out, "conflict", &["changed"], &before, &w.tip);
    assert!(
        message.contains(w.project.to_str().unwrap()),
        "the refusal names the tree: {message}"
    );
    assert_eq!(
        fs::read_to_string(w.project.join("a.txt")).unwrap(),
        "changed\n"
    );
    assert_eq!(git(&w.project, &["rev-parse", "HEAD"]), w.tip);

    // Clean again, but a second worktree has the branch checked out as well.
    git(&w.project, &["checkout", "--", "a.txt"]);
    let second = w.dir.path().join("second-checkout");
    git(
        &w.project,
        &[
            "worktree",
            "add",
            "-q",
            "-f",
            second.to_str().unwrap(),
            "integ",
        ],
    );
    let before = w.dump();
    let out = w.land_as(Some(S_LEAD), &card, &cand);
    w.refuse(&out, "conflict", &["checked out"], &before, &w.tip);
    assert_eq!(git(&w.project, &["rev-parse", "HEAD"]), w.tip);
    assert_eq!(git(&second, &["rev-parse", "HEAD"]), w.tip);
}

/// AC-3 (REQ-4): a file git ignores does not stop the landing and stays where it is.
#[test]
fn checkout_an_ignored_file_does_not_stop_the_landing() {
    let (w, card, cand) = checkout_fixture();
    let exclude = w.project.join(".git/info/exclude");
    fs::write(&exclude, "scratch.tmp\n").unwrap();
    fs::write(w.project.join("scratch.tmp"), "mine\n").unwrap();
    let got = result(&w.land_as(Some(S_LEAD), &card, &cand));
    assert_eq!(got["moved"], json!(true), "{got}");
    assert_eq!(w.integ(), cand);
    assert_eq!(
        fs::read_to_string(w.project.join("scratch.tmp")).unwrap(),
        "mine\n"
    );
}
