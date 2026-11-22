//! `moai card land` (SPEC-20261121-card-land): the leader moves the integration branch to a
//! candidate commit by one compare-and-swap and closes the card in the one transaction
//! `card finish` uses. Every test runs on a temporary home and a temporary repository whose
//! integration branch is `integ`; nothing here touches the real ledger.
//!
//! The repository: `main` holds two commits; `head` is the second (the card's change), `integ`
//! points at `tip`, a child of `head`, and a candidate is a commit made with `commit-tree` so the
//! working tree never changes.

mod common;

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
    args.push("-m");
    args.push(message);
    git(project, &args)
}

/// A commit on top of `parent` that adds `path` with `content`; no ref moves.
fn put_on(project: &Path, parent: &str, path: &str, content: &str) -> String {
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
        &["commit-tree", &tree, "-p", parent, "-m", "docs: add a file"],
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
        git(&project, &["commit", "-q", "-m", "the card's change"]);
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
        let lane = commit_with(&self.project, &self.head, &[&self.head], "the lane head");
        commit_with(
            &self.project,
            &self.tip,
            &[&self.tip, &lane],
            "merge the lane head on the tip",
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
    let out = w.land(
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
    let before = w.dump();
    let out = w.land(
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
    let out = w.land(
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
    let again = result(&w.land(
        &card,
        &["--head", &w.head, "--met", "1=the test", "--by", "leader"],
    ));
    assert_eq!(again["state"], json!("done"), "{again}");
    assert_eq!(again["moved"], json!(false), "{again}");
    assert_eq!(w.dump(), before, "a call on a done card wrote");
    // Another head on the done card refuses.
    let out = w.land(
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
    let got = result(&w.land(
        &card,
        &["--head", &candidate, "--met", "1=x", "--by", "leader"],
    ));
    assert_eq!(got["moved"], json!(true), "{got}");
    assert_eq!(got["closed"], json!(false), "{got}");
    assert_eq!(got["open_criteria"], json!([2]), "{got}");
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
    let done = result(&w.land(
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
    let got = result(&w.land(
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
    let out = w.land(
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
    let again = result(&w.land(&card, &["--head", &candidate, "--by", "leader"]));
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
    let before = w.dump();
    let out = w.land(
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
    let got = result(&w.land(
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
    result(&w.land(
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
    result(&w.land(
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
    let got = result(&w.land(
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
    let before = w.dump();
    let out = w.land(
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
    let got = result(&w.land(
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
    let out = w.land(
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
    result(&w.land(
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
    let out = w.land(
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
    let out = w.land(&held, &["--head", &next, "--met", "1=x", "--by", "leader"]);
    assert_eq!(out.status.code(), Some(1));
    assert_eq!(row(&held).len(), 1, "the refused move left no row");
    assert_eq!(w.integ(), candidate);
    *w.git_bin.borrow_mut() = git_executable();
    result(&w.land(&held, &["--head", &next, "--met", "1=x", "--by", "leader"]));
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
    let out = w.land(
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
    let again = w.land(
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
    let out = w.land(
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
    let got = result(&w.land(
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
    // No landing recorded yet: there is nothing to compare, so a first landing goes through.
    let first = w.review_card(&["one"]);
    let landed = w.candidate();
    result(&w.land(
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
    let before = w.dump();
    let out = w.land(
        &second,
        &["--head", &next, "--met", "1=x", "--by", "leader"],
    );
    let message = w.refuse(&out, "conflict", &["drift"], &before, &stray);
    for needle in [
        stray.as_str(),
        landed.as_str(),
        "git update-ref refs/heads/integ",
    ] {
        assert!(
            message.contains(needle),
            "the drift answer lacks {needle}: {message}"
        );
    }
    assert_eq!(w.events(&second, "card.land_started"), 0);

    // After the recovery the same call lands.
    git(
        &w.project,
        &["update-ref", "refs/heads/integ", &landed, &stray],
    );
    let after = commit_with(
        &w.project,
        &landed,
        &[&landed],
        "a head on the recorded tip",
    );
    result(&w.land(
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
    let out = w.land(
        &card,
        &["--head", &candidate, "--met", "1=x", "--by", "leader"],
    );
    assert_ne!(out.status.code(), Some(0), "the lander was not killed");
    assert_eq!(w.integ(), candidate, "the branch did not move");
    assert_eq!(w.events(&card, "card.land_started"), 1);
    assert_eq!(w.events(&card, "card.landed"), 0);

    *w.git_bin.borrow_mut() = git_executable();
    let rerun = result(&w.land(
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

/// AC-25 (SPEC-20261121-ref-guard): a tip that moved forward from the last landing is no drift: a
/// fast-forward made by hand keeps the last landed head inside the branch, so the next landing
/// goes through; only a tip that does not contain it is refused.
#[test]
fn ac25_a_tip_that_contains_the_last_landed_head_is_no_drift() {
    let w = World::new();
    let first = w.review_card(&["one"]);
    let landed = w.candidate();
    result(&w.land(
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
    result(&w.land(
        &second,
        &["--head", &next, "--met", "1=x", "--by", "leader"],
    ));
    assert_eq!(w.integ(), next);
}
