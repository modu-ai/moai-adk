//! `moai card land`'s remote copy and `moai card remote-copy` (SPEC-20261122-lander-push, parent
//! SPEC-20261122-remote-copy-slim REQ-1): the integration branch is copied to its remote with one
//! fast-forward-only push after every landing; a lost or refused copy leaves the landing intact
//! and is retried by the next landing or on demand. Every test runs on a temporary home and a
//! temporary repository whose remote is one local bare repository — nothing here reaches the real
//! remote and no moai command ever pushes with a force flag.
//!
//! The acceptance lines (a) to (d) are the peer proposal the leader took as this slice's criteria
//! on 2026-11-22; AC-1 also carries the parent SPEC's first acceptance.

mod common;

use common::home::TempHome;
use moai_cli::ledger;
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::Output;

const S_A: &str = "aaaaaaaa-0000-4000-8000-00000000000a";

/// The profile of an m-flow project whose integration branch is `integ` and whose remote copy
/// is on (the leader turned it on after a preview and the user's yes).
const MFLOW_PROFILE: &[u8] = b"{\"schema_version\": 1, \"project_id\": \"demo\", \"purpose\": \"fixture\", \"tools\": [\"rust\"], \"development_methods\": [\"sdd\"], \"hosts\": [\"claude\"], \"work_flow\": {\"flow\": \"m-flow\", \"integration_branch\": \"integ\"}, \"remote_copy\": true}";
/// The same profile with the copy off: the key is absent, and the copy must stay off.
const COPY_OFF_PROFILE: &[u8] = b"{\"schema_version\": 1, \"project_id\": \"demo\", \"purpose\": \"fixture\", \"tools\": [\"rust\"], \"development_methods\": [\"sdd\"], \"hosts\": [\"claude\"], \"work_flow\": {\"flow\": \"m-flow\", \"integration_branch\": \"integ\"}}";
/// The same profile with no integration branch named.
const NO_BRANCH_PROFILE: &[u8] = b"{\"schema_version\": 1, \"project_id\": \"demo\", \"purpose\": \"fixture\", \"tools\": [\"rust\"], \"development_methods\": [\"sdd\"], \"hosts\": [\"claude\"], \"work_flow\": {\"flow\": \"m-flow\"}}";

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

/// The `moai` children get a clean environment, so the git program is named for them the way
/// `card land`'s tests do.
fn git_executable() -> PathBuf {
    let path = std::env::var_os("PATH").expect("PATH");
    std::env::split_paths(&path)
        .map(|dir| dir.join("git"))
        .find(|p| p.is_file())
        .expect("a git on PATH")
}

struct World {
    home: TempHome,
    project: PathBuf,
    p: String,
    /// The bare repository that is the project's `origin`; a tempdir of its own.
    bare: tempfile::TempDir,
    bare_away: PathBuf,
    profile: Vec<u8>,
    git_bin: PathBuf,
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
        // One local bare repository is the remote; its `integ` starts even with the local tip.
        // The bare's own HEAD names another branch (main), so a test may delete `integ` on it.
        let bare = tempfile::Builder::new()
            .prefix("moai-test-remote-")
            .tempdir()
            .unwrap();
        git(bare.path(), &["init", "-q", "--bare", "-b", "main"]);
        git(
            &project,
            &["remote", "add", "origin", &bare.path().to_string_lossy()],
        );
        git(
            &project,
            &["push", "-q", "origin", "refs/heads/integ:refs/heads/integ"],
        );
        World {
            p: project.to_str().unwrap().to_string(),
            home,
            project,
            bare_away: bare.path().with_extension("away"), // a sibling path, never created
            bare,
            profile: MFLOW_PROFILE.to_vec(),
            git_bin: git_executable(),
            head,
            tip,
        }
    }

    fn write_profile(&self) {
        let cfg = self.project.join(".moai-cli/project.yaml");
        fs::create_dir_all(cfg.parent().unwrap()).unwrap();
        fs::write(&cfg, &self.profile).unwrap();
    }

    fn set_profile(&mut self, bytes: &[u8]) {
        self.profile = bytes.to_vec();
    }

    /// `moai card <args> --project <project>` as `session`, with the git program named.
    fn card(&self, session: Option<&str>, args: &[&str]) -> Output {
        self.write_profile();
        let mut all = vec!["card"];
        all.extend_from_slice(args);
        all.extend(["--project", &self.p]);
        let mut cmd = self.home.command(&all);
        cmd.env("MOAI_GIT_BIN", &self.git_bin);
        if let Some(id) = session {
            cmd.env("CLAUDE_CODE_SESSION_ID", id);
        }
        cmd.stdin(std::process::Stdio::null());
        cmd.output().expect("run moai")
    }

    /// `moai card remote-copy --project <project>`: the on-demand run, a process of its own.
    fn remote_copy(&self) -> Value {
        result(&self.card(None, &["remote-copy"]))
    }

    /// `moai card land <card> --head <candidate> --met 1=… --by leader` as the leader.
    fn land(&self, card: &str, candidate: &str) -> Value {
        result(&self.card(
            None,
            &[
                "land",
                card,
                "--head",
                candidate,
                "--met",
                "1=the test",
                "--by",
                "leader",
            ],
        ))
    }

    /// A review card with one open criterion, held by a session with a live lease.
    fn review_card(&self) -> String {
        let args = vec![
            "add",
            "--title",
            "fixture card",
            "--kind",
            "task",
            "--scope",
            "apps/a/",
            "--ac",
            "one",
        ];
        let added = result(&self.card(None, &args));
        let name = added["name"].as_str().unwrap().to_string();
        result(&self.card(None, &["ready", &name]));
        result(&self.card(Some(S_A), &["take", &name]));
        result(&self.card(Some(S_A), &["review", &name, "--generation", "1"]));
        name
    }

    /// A candidate: a merge of a lane head onto the tip, the shape the leader lands.
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

    /// The commit the bare remote's `integ` points at now; empty when the remote cannot be
    /// reached (the bare is away).
    fn remote_tip(&self) -> String {
        let all = [
            "-c",
            "user.name=fixture",
            "-c",
            "user.email=fixture@example.invalid",
            "ls-remote",
            "origin",
            "refs/heads/integ",
        ];
        let out = common::git::git_raw(&self.project, &all);
        if !out.success {
            return String::new();
        }
        out.stdout
            .split_whitespace()
            .next()
            .unwrap_or_default()
            .to_string()
    }

    /// Moves the bare away, so the remote is unreachable.
    fn bare_away(&self) {
        fs::rename(self.bare.path(), &self.bare_away).unwrap();
    }

    /// Restores the bare after [`bare_away`].
    fn bare_back(&self) {
        fs::rename(&self.bare_away, self.bare.path()).unwrap();
    }

    /// Pushes `commit` to the bare as `integ` with a force flag — plain git in the test setup
    /// only, to place the remote somewhere a moai copy must refuse to reach.
    fn force_remote_to(&self, commit: &str) {
        git(
            &self.project,
            &[
                "push",
                "-q",
                "--force",
                "origin",
                &format!("{commit}:refs/heads/integ"),
            ],
        );
    }

    /// A commit that shares no ancestry with the local history.
    fn unrelated(&self) -> String {
        let tree = git(
            &self.project,
            &["rev-parse", &format!("{}^{{tree}}", self.head)],
        );
        git(
            &self.project,
            &["commit-tree", &tree, "-m", "an unrelated line"],
        )
    }

    /// The copy record: the one file `moai/remote-copy` in the common git folder.
    fn record(&self) -> PathBuf {
        let common = git(&self.project, &["rev-parse", "--git-common-dir"]);
        self.project.join(common.trim()).join("moai/remote-copy")
    }

    fn record_text(&self) -> String {
        fs::read_to_string(self.record()).unwrap()
    }

    /// The copy record as JSON (it is written pretty, so it is parsed, not searched).
    fn record_state(&self) -> String {
        let text = self.record_text();
        let v: Value = serde_json::from_str(&text).unwrap();
        v["state"].as_str().unwrap().to_string()
    }

    fn events(&self, card: &str, kind: &str) -> i64 {
        let l = ledger::open_read_only(self.home.path(), "demo").unwrap();
        l.conn()
            .query_row(
                "SELECT count(*) FROM event WHERE card_id = ?1 AND kind = ?2",
                rusqlite::params![number(card), kind],
                |r| r.get(0),
            )
            .unwrap()
    }
}

// ---- AC-1 (the parent SPEC's first acceptance): fast-forward only, and a refused push reported --

/// AC-1: a landing whose remote is behind pushes fast-forward; a landing whose remote shares no
/// ancestry completes, closes the card, moves the branch, reports the refused push and leaves the
/// remote as it was.
#[test]
fn card_land_pushes_fast_forward_only_and_reports_a_refused_push() {
    let w = World::new();
    // Behind: the landing moves the local branch, the copy fast-forwards the remote to it.
    let card = w.review_card();
    let candidate = w.candidate();
    let landed = w.land(&card, &candidate);
    assert_eq!(landed["closed"], json!(true), "{landed}");
    assert_eq!(landed["remote_copy"]["state"], json!("done"), "{landed}");
    assert_eq!(landed["remote_copy"]["pushed"], json!(true), "{landed}");
    assert_eq!(w.integ(), candidate);
    assert_eq!(w.remote_tip(), candidate, "the remote is fast-forwarded");
    assert_eq!(w.record_state(), "done");

    // No shared ancestry: the landing still completes and reports the refused push as a hold.
    let w2 = World::new();
    let card2 = w2.review_card();
    let candidate2 = w2.candidate();
    let diverged = w2.unrelated();
    w2.force_remote_to(&diverged);
    let refused = w2.land(&card2, &candidate2);
    assert_eq!(refused["closed"], json!(true), "{refused}");
    assert_eq!(refused["moved"], json!(true), "{refused}");
    assert_eq!(refused["remote_copy"]["state"], json!("hold"), "{refused}");
    assert!(w2.integ() == candidate2, "the local branch moved");
    assert_eq!(w2.remote_tip(), diverged, "the remote stays where it was");
    assert_eq!(w2.record_state(), "hold");
}

// ---- AC-2 (peer line a): offline landing, new run pushes ------------------------------------

/// AC-2: the last landing completes with the remote unreachable, the process ends, and a new
/// on-demand run pushes that commit fast-forward without another landing.
#[test]
fn an_offline_landing_completes_and_a_new_run_pushes_the_commit() {
    let w = World::new();
    w.bare_away();
    let card = w.review_card();
    let candidate = w.candidate();
    let landed = w.land(&card, &candidate);
    assert_eq!(landed["closed"], json!(true), "{landed}");
    assert_eq!(landed["remote_copy"]["state"], json!("pending"), "{landed}");
    assert_eq!(w.integ(), candidate, "the landing itself completed");
    assert!(
        w.remote_tip().is_empty(),
        "the unreachable remote answers nothing"
    );
    assert_eq!(w.record_state(), "pending");

    w.bare_back();
    let copied = w.remote_copy();
    assert_eq!(copied["state"], json!("done"), "{copied}");
    assert_eq!(copied["pushed"], json!(true), "{copied}");
    assert_eq!(w.remote_tip(), candidate, "the new run pushed the commit");
}

// ---- AC-3 (peer line b): a lost answer is settled as success --------------------------------

/// AC-3: when the remote already holds the same commit or a descendant and only the answer was
/// lost, the copy is settled as success with no push.
#[test]
fn a_lost_answer_is_settled_as_success() {
    let w = World::new();
    let card = w.review_card();
    let candidate = w.candidate();
    w.land(&card, &candidate);
    assert_eq!(w.remote_tip(), candidate);

    // The same commit: only the answer was lost, nothing is pushed again.
    let same = w.remote_copy();
    assert_eq!(same["state"], json!("done"), "{same}");
    assert_eq!(same["pushed"], json!(false), "{same}");
    assert_eq!(w.remote_tip(), candidate);

    // A descendant: a later copy reached the remote and its answer was lost.
    let ahead = commit_with(
        &w.project,
        &candidate,
        &[&candidate],
        "a later copy reached the remote",
    );
    w.force_remote_to(&ahead);
    let settled = w.remote_copy();
    assert_eq!(settled["state"], json!("done"), "{settled}");
    assert_eq!(settled["pushed"], json!(false), "{settled}");
    assert_eq!(w.remote_tip(), ahead, "the remote keeps its descendant");
}

// ---- AC-4 (peer line c): divergence is a hold ------------------------------------------------

/// AC-4: when the histories diverged, remote and local stay as they are and the reason is
/// recorded as a hold.
#[test]
fn a_diverged_remote_is_recorded_as_a_hold_and_nothing_moves() {
    let w = World::new();
    let before = w.integ();
    let diverged = w.unrelated();
    w.force_remote_to(&diverged);
    let held = w.remote_copy();
    assert_eq!(held["state"], json!("hold"), "{held}");
    assert!(held["why"].is_string(), "{held}");
    assert_eq!(w.integ(), before, "the local branch stays");
    assert_eq!(w.remote_tip(), diverged, "the remote stays");
    assert_eq!(w.record_state(), "hold");
}

// ---- AC-5 (peer line d): the local completion is recorded once -------------------------------

/// AC-5: after the offline landing and its retry the ledger holds one landing event and one close
/// for the card, and the copy state is one record.
#[test]
fn the_local_completion_is_recorded_once() {
    let w = World::new();
    w.bare_away();
    let card = w.review_card();
    let candidate = w.candidate();
    w.land(&card, &candidate);
    w.bare_back();
    w.remote_copy();
    w.remote_copy(); // a second on-demand run changes nothing

    assert_eq!(w.events(&card, "card.landed"), 1);
    assert_eq!(w.events(&card, "card.finish"), 1);
    let record = w.record();
    assert!(record.is_file(), "exactly one record file exists");
    let dir = record.parent().unwrap();
    let moai_entries: Vec<_> = fs::read_dir(dir).unwrap().collect();
    assert_eq!(moai_entries.len(), 1, "no second record was written");
    assert_eq!(w.record_state(), "done");
}

// ---- AC-6: an absent remote branch is created; a profile without a branch is reported --------

/// AC-6: a remote without the branch takes the copy as a creation that destroys nothing, and a
/// profile with no integration branch answers pending with the reason and no push.
#[test]
fn an_absent_remote_branch_is_created_and_a_profile_without_a_branch_is_reported() {
    let mut w = World::new();
    git(
        &w.project,
        &["push", "-q", "origin", "--delete", "refs/heads/integ"],
    );
    let card = w.review_card();
    let candidate = w.candidate();
    let landed = w.land(&card, &candidate);
    assert_eq!(landed["remote_copy"]["state"], json!("done"), "{landed}");
    assert_eq!(landed["remote_copy"]["pushed"], json!(true), "{landed}");
    assert_eq!(w.remote_tip(), candidate, "the branch was created");

    w.set_profile(NO_BRANCH_PROFILE);
    let reported = w.remote_copy();
    assert_eq!(reported["state"], json!("pending"), "{reported}");
    assert!(reported["why"].is_string(), "{reported}");
    assert_eq!(w.integ(), candidate, "nothing moved");
}

// ---- B1: the copy is off by default and answers off with a preview ---------------------------

/// B1: with the key absent a landing pushes nothing, answers `off`, and the remote stays where
/// it was.
#[test]
fn a_landing_with_the_copy_off_pushes_nothing_and_answers_off() {
    let mut w = World::new();
    w.set_profile(COPY_OFF_PROFILE);
    let seeded = w.remote_tip();
    let card = w.review_card();
    let candidate = w.candidate();
    let landed = w.land(&card, &candidate);
    assert_eq!(landed["closed"], json!(true), "{landed}");
    assert_eq!(landed["remote_copy"]["state"], json!("off"), "{landed}");
    assert_eq!(landed["remote_copy"]["pushed"], json!(false), "{landed}");
    assert_eq!(w.integ(), candidate, "the landing itself completed");
    assert_eq!(w.remote_tip(), seeded, "nothing reached the remote");
    assert_eq!(w.record_state(), "off");
}

/// B1: with the copy off the on-demand command answers a preview and writes nothing — no push,
/// no record.
#[test]
fn the_off_command_answers_a_preview_and_writes_nothing() {
    let mut w = World::new();
    w.set_profile(COPY_OFF_PROFILE);
    let seeded = w.remote_tip();
    let answered = w.remote_copy();
    assert_eq!(answered["state"], json!("off"), "{answered}");
    let preview = &answered["preview"];
    assert_eq!(preview["remote"], json!("origin"), "{answered}");
    assert_eq!(preview["branch"], json!("integ"), "{answered}");
    assert!(preview["behind_by"].is_number(), "{answered}");
    assert!(preview["ci"].is_string(), "{answered}");
    assert!(!w.record().exists(), "the off answer writes no record");
    assert_eq!(w.remote_tip(), seeded, "nothing reached the remote");
}

// ---- B2: a token inside a push-remote URL never reaches an answer or the record --------------

/// B2: a pushRemote URL holding a token is redacted wherever the copy answers or records it.
#[test]
fn a_token_in_the_push_remote_url_is_redacted() {
    let w = World::new();
    git(
        &w.project,
        &[
            "config",
            "branch.integ.pushRemote",
            "https://user:secret-tok@example.invalid/x.git",
        ],
    );
    let answered = w.remote_copy();
    let text = serde_json::to_string(&answered).unwrap();
    assert!(!text.contains("secret-tok"), "{answered}");
    assert!(!text.contains("user:"), "{answered}");
    if w.record().exists() {
        assert!(!w.record_text().contains("secret-tok"));
    }
}

// ---- B3: no tag follows the push, and the push arguments carry the safety flags --------------

/// B3: with push.followTags on, a tag on the tip does not reach the remote.
#[test]
fn a_tag_on_the_tip_is_not_pushed() {
    let w = World::new();
    git(&w.project, &["config", "push.followTags", "true"]);
    let card = w.review_card();
    let candidate = w.candidate();
    git(&w.project, &["tag", "probe-tag", &candidate]);
    let landed = w.land(&card, &candidate);
    assert_eq!(landed["remote_copy"]["state"], json!("done"), "{landed}");
    assert!(w.remote_tip() == candidate, "the branch was copied");
    let tags = git(&w.project, &["ls-remote", "--tags", "origin"]);
    assert!(tags.trim().is_empty(), "no tag reached the remote: {tags}");
}

/// B3 and the auditor's probe: the push argv carries --no-follow-tags and `--` before the
/// remote name, and never --force, --mirror or a +refspec. A wrapper around git records it.
#[test]
fn the_push_arguments_carry_the_safety_flags() {
    let w = World::new();
    let log = w.bare.path().with_extension("argv");
    let wrapper = w.bare.path().with_extension("gitwrap");
    let real = git_executable();
    fs::write(
        &wrapper,
        format!(
            "#!/bin/sh\nfor a in \"$@\"; do [ \"$a\" = push ] && printf '%s\\n' \"$*\" >> {}; done\nexec {} \"$@\"\n",
            log.to_string_lossy(),
            real.to_string_lossy(),
        ),
    )
    .unwrap();
    {
        use std::os::unix::fs::PermissionsExt;
        fs::set_permissions(&wrapper, fs::Permissions::from_mode(0o755)).unwrap();
    }
    w.write_profile();
    let mut cmd = w.home.command(&["card", "remote-copy", "--project", &w.p]);
    cmd.env("MOAI_GIT_BIN", &wrapper);
    let out = cmd.output().unwrap();
    let answered = result(&out);
    assert_eq!(answered["state"], json!("done"), "{answered}");
    let argv = fs::read_to_string(&log).unwrap();
    let push_line = argv
        .lines()
        .find(|l| l.contains(" push "))
        .or_else(|| argv.lines().next())
        .expect("a recorded push call");
    assert!(push_line.contains("--no-follow-tags"), "{push_line}");
    assert!(push_line.contains("-- origin refs/heads/"), "{push_line}");
    assert!(!push_line.contains("--force"), "{push_line}");
    assert!(!push_line.contains("--mirror"), "{push_line}");
    assert!(!push_line.contains("+refs/"), "{push_line}");
}

// ---- N2 and N3: the unreachable push and the unknown descendant -------------------------------

/// N3: a remote holding a descendant the local repository does not know settles as success,
/// never as a hold (the parent SPEC's AC-3).
#[test]
fn an_unknown_descendant_settles_as_done() {
    let w = World::new();
    // A third clone makes a commit on the remote that the project never fetched.
    let third = tempfile::Builder::new()
        .prefix("moai-test-third-")
        .tempdir()
        .unwrap();
    git(
        third.path(),
        &["clone", "-q", &w.bare.path().to_string_lossy(), "c"],
    );
    let clone = third.path().join("c");
    git(&clone, &["checkout", "-q", "-b", "integ", "origin/integ"]);
    git(
        &clone,
        &["commit", "-q", "--allow-empty", "-m", "a remote child"],
    );
    let pushed = git(&clone, &["rev-parse", "HEAD"]);
    git(&clone, &["push", "-q", "origin", "integ:refs/heads/integ"]);
    let settled = w.remote_copy();
    assert_eq!(settled["state"], json!("done"), "{settled}");
    assert_eq!(settled["pushed"], json!(false), "{settled}");
    assert_eq!(w.remote_tip(), pushed);
    assert!(!w.record_text().contains("\"hold\""));
}

/// N2: a push that cannot reach the remote reads pending, never a hold.
#[test]
fn an_unreachable_push_reads_pending() {
    let w = World::new();
    let wrapper = w.bare.path().with_extension("pushfail");
    let real = git_executable();
    fs::write(
        &wrapper,
        format!(
            "#!/bin/sh\nfor a in \"$@\"; do [ \"$a\" = push ] && echo 'fatal: test unreachable' >&2 && exit 128; done\nexec {} \"$@\"\n",
            real.to_string_lossy(),
        ),
    )
    .unwrap();
    {
        use std::os::unix::fs::PermissionsExt;
        fs::set_permissions(&wrapper, fs::Permissions::from_mode(0o755)).unwrap();
    }
    w.write_profile();
    let card = w.review_card();
    let candidate = w.candidate();
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
    ]);
    cmd.env("MOAI_GIT_BIN", &wrapper);
    let out = cmd.output().unwrap();
    let landed = result(&out);
    assert_eq!(landed["closed"], json!(true), "{landed}");
    assert_eq!(landed["remote_copy"]["state"], json!("pending"), "{landed}");
    assert_eq!(w.record_state(), "pending");
}
