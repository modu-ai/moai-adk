//! `moai card next` and a `ready` card with no counted criterion whose landing is proven
//! (SPEC-20261124-card-next-skip-reason): the landing proof wins, so the card is skipped as
//! `already_landed` with the proof and not as `no_criteria`. A card with no proof keeps
//! `no_criteria`; a held card keeps `held`.
//!
//! Every command runs against a temporary HOME and a temporary git repository; the real `~/.moai`
//! is not read or written. The caller is the session id in the environment
//! (`CLAUDE_CODE_SESSION_ID`).
#![cfg(unix)]

mod common;

use common::home::TempHome;
use moai_cli::ledger;
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Output, Stdio};

const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
const S_A: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
const S_B: &str = "bbbbbbbb-0000-4000-8000-00000000000b";

/// The profile of an m-flow project whose integration branch is `integ`.
const MFLOW_PROFILE: &[u8] = b"{\"schema_version\": 1, \"project_id\": \"demo\", \"purpose\": \"fixture\", \"tools\": [\"rust\"], \"development_methods\": [\"sdd\"], \"hosts\": [\"claude\"], \"work_flow\": {\"flow\": \"m-flow\", \"integration_branch\": \"integ\"}}";

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

/// A commit with the tree of `parent` and `parent` as its only parent; `branch` is moved to it
/// without a checkout, so the working tree never changes.
fn commit_on(project: &Path, branch: &str, parent: &str, message: &str) -> String {
    let tree = git(project, &["rev-parse", &format!("{parent}^{{tree}}")]);
    let sha = git(
        project,
        &["commit-tree", tree.as_str(), "-p", parent, "-m", message],
    );
    git(
        project,
        &["update-ref", &format!("refs/heads/{branch}"), &sha],
    );
    sha
}

/// An absolute `git` on this runner's `PATH`, for the child that has none.
fn git_executable() -> Option<PathBuf> {
    let path = std::env::var_os("PATH")?;
    std::env::split_paths(&path)
        .map(|dir| dir.join("git"))
        .find(|p| p.is_file())
}

fn result(out: &Output) -> Value {
    let v: Value = serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
        panic!(
            "stdout is not JSON ({e}): {} {}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    });
    assert_eq!(out.status.code(), Some(0), "{v}");
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn names(list: &Value) -> Vec<String> {
    list.as_array()
        .unwrap()
        .iter()
        .map(|e| e["name"].as_str().unwrap().to_string())
        .collect()
}

/// The entries of `skipped` for one card.
fn skipped_entries<'a>(listing: &'a Value, name: &str) -> Vec<&'a Value> {
    listing["skipped"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|e| e["name"] == json!(name))
        .collect()
}

/// The world: a temporary HOME with a registered ledger and a temporary git repository as the
/// project. `main` is at `head` (the card's commit), `integ` is one commit ahead of `head`, and
/// `scratch` has one commit outside `integ` (`orphan`).
struct World {
    home: TempHome,
    project: PathBuf,
    p: String,
    git_bin: Option<PathBuf>,
    head: String,
    orphan: String,
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
        git(&project, &["branch", "integ"]);
        commit_on(&project, "integ", &head, "integration moved on");
        let orphan = commit_on(&project, "scratch", &head, "not integrated");
        World {
            p: project.to_str().unwrap().to_string(),
            home,
            project,
            git_bin: git_executable(),
            head,
            orphan,
        }
    }

    fn command(&self, session: Option<&str>, args: &[&str]) -> Command {
        let mut cmd = self.home.command(args);
        if let Some(git) = &self.git_bin {
            cmd.env("MOAI_GIT_BIN", git);
        }
        if let Some(id) = session {
            cmd.env(SESSION_VAR, id);
        }
        cmd.stdin(Stdio::null());
        cmd
    }

    /// `moai card <args> --project <project>`, as `session`.
    fn card(&self, session: Option<&str>, args: &[&str]) -> Value {
        let mut all = vec!["card"];
        all.extend_from_slice(args);
        all.extend_from_slice(&["--project", &self.p]);
        result(&self.command(session, &all).output().expect("run moai"))
    }

    fn write_profile(&self) {
        let cfg = self.project.join(".moai-cli/project.yaml");
        fs::create_dir_all(cfg.parent().unwrap()).unwrap();
        fs::write(&cfg, MFLOW_PROFILE).unwrap();
    }

    /// `moai card next <project>` as `session`, with the profile as it was written last.
    fn next_unwritten(&self, session: &str) -> Value {
        result(
            &self
                .command(Some(session), &["card", "next", &self.p])
                .output()
                .expect("run moai"),
        )
    }

    /// `moai card next <project>` as `session`, with the profile written last.
    fn next(&self, session: &str) -> Value {
        self.write_profile();
        self.next_unwritten(session)
    }

    /// A `ready` card with one write scope and one open criterion; returns its name.
    fn card_with_one_criterion(&self, scope: &str) -> String {
        let added = self.card(
            None,
            &[
                "add", "--title", "fixture", "--kind", "task", "--scope", scope, "--ac", "one",
            ],
        );
        let name = added["name"].as_str().unwrap().to_string();
        self.card(None, &["ready", &name]);
        name
    }

    /// `S_A` takes the card, sends it to review and records `head` as landed, then reopens it. It
    /// still holds the card (generation 1) when this returns.
    fn land_and_reopen(&self, card: &str, head: &str) {
        self.card(Some(S_A), &["take", card]);
        self.card(Some(S_A), &["review", card, "--generation", "1"]);
        self.card(
            Some(S_A),
            &["landed", card, "--head", head, "--by", "leader"],
        );
        self.card(Some(S_A), &["reopen", card, "--generation", "1"]);
    }

    fn withdraw_criterion_one(&self, card: &str) {
        self.card(
            Some(S_A),
            &[
                "criterion",
                card,
                "1",
                "--generation",
                "1",
                "--withdraw",
                "--reason",
                "fixture: no criterion counts",
            ],
        );
    }

    fn release(&self, card: &str) {
        self.card(Some(S_A), &["release", card, "--generation", "1"]);
    }

    /// A `ready` card with a landing event for `head` and no counted criterion (its only
    /// criterion is withdrawn); nobody holds it.
    fn landed_card_without_criterion(&self, scope: &str, head: &str) -> String {
        let card = self.card_with_one_criterion(scope);
        self.land_and_reopen(&card, head);
        self.withdraw_criterion_one(&card);
        self.release(&card);
        card
    }
}

/// AC-1: a `ready` card with no counted criterion and a landed head inside the integration branch
/// is listed once under skipped as `already_landed` with its proof, is not in `ready` or `next`,
/// and `next` writes nothing.
#[test]
fn a_landed_card_with_no_counted_criterion_is_already_landed_and_not_no_criteria() {
    let w = World::new();
    let landed = w.landed_card_without_criterion("apps/a/", &w.head);

    w.write_profile();
    let before = w.home.snapshot();
    let listing = w.next_unwritten(S_B);
    // `assert!` and not `assert_eq!`: a failure must not print the whole ledger file.
    assert!(w.home.snapshot() == before, "next writes nothing");

    let entries = skipped_entries(&listing, &landed);
    assert_eq!(
        entries.len(),
        1,
        "the card is skipped exactly once: {listing}"
    );
    let entry = entries[0];
    assert_eq!(entry["reason"], json!("already_landed"), "{entry}");
    assert_eq!(entry["head"], json!(w.head), "{entry}");
    assert!(
        !entry["via"].as_str().unwrap_or("").is_empty(),
        "the skip entry names its proof: {entry}"
    );
    assert!(
        !names(&listing["ready"]).contains(&landed),
        "{landed} is offered under ready: {listing}"
    );
    assert_eq!(listing["next"], Value::Null, "{listing}");
    assert_eq!(listing["landing"]["checked"], json!(true), "{listing}");
}

/// AC-2: with no landing a card that has no counted criterion keeps `no_criteria`, and so does one
/// whose recorded head is outside the integration branch.
#[test]
fn a_card_with_no_counted_criterion_and_no_landing_keeps_no_criteria() {
    let w = World::new();
    let plain = w.card_with_one_criterion("apps/a/");
    w.card(Some(S_A), &["take", &plain]);
    w.withdraw_criterion_one(&plain);
    w.release(&plain);
    let outside = w.landed_card_without_criterion("apps/b/", &w.orphan);

    let listing = w.next(S_B);
    for card in [&plain, &outside] {
        let entries = skipped_entries(&listing, card);
        assert_eq!(entries.len(), 1, "{card} is skipped once: {listing}");
        assert_eq!(entries[0]["reason"], json!("no_criteria"), "{listing}");
        assert!(!names(&listing["ready"]).contains(card), "{listing}");
    }
}

/// AC-2: a card with one open criterion and a landed head is `already_landed` as before, and a
/// card held by another session keeps `held` whether or not it has a criterion.
#[test]
fn the_other_skip_reasons_stay_as_they_were() {
    let w = World::new();
    let with_criterion = w.card_with_one_criterion("apps/a/");
    w.land_and_reopen(&with_criterion, &w.head);
    w.release(&with_criterion);
    let held = w.card_with_one_criterion("apps/b/");
    w.card(Some(S_A), &["take", &held]);
    w.withdraw_criterion_one(&held);

    let listing = w.next(S_B);
    let landed = skipped_entries(&listing, &with_criterion);
    assert_eq!(landed.len(), 1, "{listing}");
    assert_eq!(landed[0]["reason"], json!("already_landed"), "{listing}");
    let held_entries = skipped_entries(&listing, &held);
    assert_eq!(held_entries.len(), 1, "{listing}");
    assert_eq!(held_entries[0]["reason"], json!("held"), "{listing}");
}
