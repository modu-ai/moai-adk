//! SDD-7 (SPEC-20261118-spec-take): `moai spec take <SPEC-ID>` starts work on a SPEC — checks
//! it, creates or switches to its branch `spec/<yyyymmdd>-<slug>`, links a card on request and
//! prints the next commands. Push and draft PR run only when asked. Everything happens in a
//! temporary HOME, a temporary repository, a local bare repository as the remote and a fake
//! `gh`: never the network, the real HOME or the real ledger.

mod common;

use common::home::TempHome;
use moai_cli::ledger;
use serde_json::Value;
use std::fs;
use std::path::{Path, PathBuf};
use std::process::Output;

const BRANCH: &str = "spec/20261117-login-fix";

fn envelope(out: &Output) -> Value {
    serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
        panic!(
            "{} / {}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    })
}

fn ok(out: &Output) -> Value {
    let v = envelope(out);
    assert_eq!(out.status.code(), Some(0), "{v}");
    v["result"].clone()
}

fn refused(out: &Output) -> String {
    let v = envelope(out);
    assert_eq!(out.status.code(), Some(1), "{v}");
    v["error"]["message"]
        .as_str()
        .unwrap_or_default()
        .to_string()
}

/// A HOME, a registered ledger and a git repository `demo` on branch `main` with one commit.
struct World {
    home: TempHome,
    project: PathBuf,
}

impl World {
    fn new() -> Self {
        let home = TempHome::new();
        let project = home.project("demo");
        ledger::register(home.path(), "demo").unwrap();
        let w = World { home, project };
        w.git(&["init", "-q", "-b", "main"]);
        fs::write(w.project.join("a.txt"), "a\n").unwrap();
        w.git(&["add", "-A"]);
        w.commit("init");
        w
    }

    fn p(&self) -> &str {
        self.project.to_str().unwrap()
    }

    fn git(&self, args: &[&str]) -> String {
        common::git::git(&self.project, args)
    }

    fn commit(&self, message: &str) {
        common::git::git_ok(
            &self.project,
            &[
                "-c",
                "user.name=fixture",
                "-c",
                "user.email=fixture@example.invalid",
                "-c",
                "commit.gpgsign=false",
                "commit",
                "-q",
                "-m",
                message,
            ],
        );
    }

    fn run(&self, args: &[&str], env: &[(&str, &str)]) -> Output {
        let mut cmd = self.home.command(args);
        for (k, v) in env {
            cmd.env(k, v);
        }
        cmd.output().expect("run moai")
    }

    /// `moai spec take <id> --project P <extra>`.
    fn take(&self, id: &str, extra: &[&str], env: &[(&str, &str)]) -> Output {
        let mut all = vec!["spec", "take", id, "--project", self.p()];
        all.extend_from_slice(extra);
        self.run(&all, env)
    }

    /// A draft SPEC `SPEC-20261117-<slug>` that passes `spec check` (placeholders filled).
    fn draft_spec(&self, slug: &str) -> String {
        let r = ok(&self.run(
            &[
                "spec",
                "new",
                slug,
                "--title",
                "Login fix",
                "--minimal",
                "--date",
                "2026-11-17",
                "--project",
                self.p(),
            ],
            &[],
        ));
        let id = r["id"].as_str().unwrap().to_string();
        let path = self.project.join(".moai/specs").join(&id).join("spec.md");
        fs::write(&path, filled(&fs::read_to_string(&path).unwrap())).unwrap();
        id
    }

    fn branch(&self) -> String {
        self.git(&["rev-parse", "--abbrev-ref", "HEAD"])
    }

    /// Branches, HEAD and the status: what a refusal must leave as it was.
    fn state(&self) -> (String, String, String) {
        (
            self.git(&[
                "branch",
                "--list",
                "--format=%(refname:short) %(objectname)",
            ]),
            self.git(&["rev-parse", "HEAD"]),
            self.git(&["status", "--porcelain"]),
        )
    }
}

/// A template's text with every `<…>` placeholder filled with the word `filled`.
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

#[test]
fn take_creates_the_branch_and_prints_the_next_commands() {
    let w = World::new();
    let id = w.draft_spec("login-fix");
    let base = w.git(&["rev-parse", "HEAD"]);
    let r = ok(&w.take(&id, &[], &[]));
    assert_eq!(r["action"], "take");
    assert_eq!(r["spec"], id.as_str());
    assert_eq!(r["branch"], BRANCH);
    assert_eq!(r["base"], "main");
    assert_eq!(r["created"], true, "{r}");
    assert_eq!(r["switched"], true, "{r}");
    assert_eq!(r["pushed"], false);
    assert!(r["draft_pr"].is_null(), "{r}");
    assert_eq!(w.branch(), BRANCH);
    assert_eq!(
        w.git(&["rev-parse", BRANCH]),
        base,
        "the branch starts at the base"
    );
    // The new SPEC file travels with the switch, still uncommitted.
    assert!(w
        .project
        .join(".moai/specs")
        .join(&id)
        .join("spec.md")
        .is_file());
    let next: Vec<&str> = r["next"]
        .as_array()
        .unwrap()
        .iter()
        .map(|v| v.as_str().unwrap())
        .collect();
    assert_eq!(
        next,
        [
            format!("moai trailers --spec {id}").as_str(),
            &format!("git push -u origin {BRANCH}"),
            &format!(
                "gh pr create --draft --base main --head {BRANCH} --title 'Login fix' --body 'SPEC: {id}'"
            ),
        ],
        "{r}"
    );
    // Again on the branch: nothing to create or switch; still fine.
    let again = ok(&w.take(&id, &[], &[]));
    assert_eq!(
        (again["created"].clone(), again["switched"].clone()),
        (false.into(), false.into())
    );
    // From another branch it switches back to the existing one.
    w.git(&["switch", "-q", "main"]);
    let back = ok(&w.take(&id, &["--base", "main"], &[]));
    assert_eq!(
        (back["created"].clone(), back["switched"].clone()),
        (false.into(), true.into())
    );
    assert_eq!(w.branch(), BRANCH);
}

#[test]
fn take_refuses_and_changes_nothing() {
    let w = World::new();
    let id = w.draft_spec("login-fix");
    let before = w.state();
    // Not there, a Go-generation id, or an id in no form.
    for missing in ["SPEC-20261117-nothing-here", "SPEC-AUTH-001", "C-87"] {
        let m = refused(&w.take(missing, &[], &[]));
        assert!(m.starts_with("spec take refused: "), "{missing}: {m}");
        assert_eq!(w.state(), before, "{missing}");
    }
    // A SPEC with an error (no frontmatter).
    let broken = w.project.join(".moai/specs/SPEC-20261117-broken");
    fs::create_dir_all(&broken).unwrap();
    fs::write(broken.join("spec.md"), "# no frontmatter\n").unwrap();
    let broken_state = w.state();
    let m = refused(&w.take("SPEC-20261117-broken", &[], &[]));
    assert!(m.contains("error"), "{m}");
    assert_eq!(w.state(), broken_state);
    fs::remove_dir_all(&broken).unwrap();
    // A change outside the SPEC's own folder.
    fs::write(w.project.join("a.txt"), "changed\n").unwrap();
    let dirty = w.state();
    let m = refused(&w.take(&id, &[], &[]));
    assert!(m.contains("a.txt"), "{m}");
    assert_eq!(w.state(), dirty);
    w.git(&["checkout", "-q", "--", "a.txt"]);
    // `--draft-pr` without `--push`: refused before any write.
    let m = refused(&w.take(&id, &["--draft-pr"], &[]));
    assert!(m.contains("--push"), "{m}");
    assert_eq!(w.state(), before);
    // The branch exists and carries another SPEC's commit.
    w.git(&["switch", "-q", "-c", BRANCH]);
    fs::write(w.project.join("b.txt"), "b\n").unwrap();
    w.git(&["add", "b.txt"]);
    w.commit("other work\n\nMoai-Spec: SPEC-20261117-other-fix");
    w.git(&["switch", "-q", "main"]);
    let held = w.state();
    let m = refused(&w.take(&id, &[], &[]));
    assert!(m.contains("SPEC-20261117-other-fix"), "{m}");
    assert_eq!(w.state(), held);
}

#[test]
fn take_links_the_card_in_the_local_ledger() {
    let w = World::new();
    let id = w.draft_spec("login-fix");
    let added = ok(&w.run(
        &[
            "card",
            "add",
            "--title",
            "t",
            "--kind",
            "fix",
            "--scope",
            "src/",
            "--project",
            w.p(),
        ],
        &[],
    ));
    let name = added["name"].as_str().unwrap().to_string();
    let r = ok(&w.take(&id, &["--card", &name], &[]));
    assert_eq!(r["card"]["changed"], true, "{r}");
    assert_eq!(r["card"]["spec"], id.as_str());
    let shown = ok(&w.run(
        &[
            "spec",
            "show",
            &id,
            "--project",
            w.p(),
            "--home",
            w.home.path().to_str().unwrap(),
        ],
        &[],
    ));
    assert_eq!(shown["local_cards"]["count"], 1, "{shown}");
    // No card number reaches a shared place: the printed commands name none.
    assert!(!r["next"].to_string().contains(&name), "{r}");
}

/// A fake `gh` that records its arguments and prints a pull request URL.
fn fake_gh(dir: &Path) -> (PathBuf, PathBuf) {
    let log = dir.join("gh-args.txt");
    let bin = dir.join("fake-gh");
    common::write_executable(
        &bin,
        &format!(
            "#!/bin/sh\nprintf '%s\\n' \"$@\" > '{}'\necho https://github.example/o/r/pull/7\n",
            log.display()
        ),
    );
    (bin, log)
}

#[test]
fn push_and_draft_pr_run_only_when_asked() {
    let w = World::new();
    let id = w.draft_spec("login-fix");
    let remote = w.home.path().join("remote.git");
    common::git::git_ok(
        w.home.path(),
        &["init", "-q", "--bare", remote.to_str().unwrap()],
    );
    w.git(&["remote", "add", "origin", remote.to_str().unwrap()]);
    let (gh, log) = fake_gh(w.home.path());
    let gh_env = [("MOAI_FORGE_GH_BIN", gh.to_str().unwrap())];
    let on_remote = || {
        common::git::git(
            w.home.path(),
            &[
                "--git-dir",
                remote.to_str().unwrap(),
                "branch",
                "--list",
                BRANCH,
            ],
        )
    };
    // Default: nothing leaves the machine.
    let r = ok(&w.take(&id, &[], &gh_env));
    assert_eq!(r["pushed"], false);
    assert_eq!(on_remote(), "");
    assert!(!log.exists(), "gh was not run");
    // --push: the branch reaches the remote; gh still not run.
    let r = ok(&w.take(&id, &["--push"], &gh_env));
    assert_eq!(r["pushed"], true, "{r}");
    assert!(on_remote().contains(BRANCH), "{}", on_remote());
    assert!(!log.exists(), "gh was not run");
    // --push --draft-pr: the fake gh is called with the printed arguments.
    let r = ok(&w.take(&id, &["--push", "--draft-pr"], &gh_env));
    assert_eq!(
        r["draft_pr"]["url"], "https://github.example/o/r/pull/7",
        "{r}"
    );
    let args: Vec<String> = fs::read_to_string(&log)
        .unwrap()
        .lines()
        .map(String::from)
        .collect();
    assert_eq!(
        args,
        [
            "pr",
            "create",
            "--draft",
            "--base",
            "main",
            "--head",
            BRANCH,
            "--title",
            "Login fix",
            "--body",
            &format!("SPEC: {id}"),
        ]
    );
}
