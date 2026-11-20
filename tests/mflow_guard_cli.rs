//! SPEC-20261121-mflow-forward-guard: the forward-only guard scenes in temp repos. The
//! generated hook script goes into the repo's own `.git/hooks/reference-transaction` and real
//! git commands run against it; the tool-guard rule is judged on command words.

use moai_cli::mflow_guard::{guard_rule, hook_script, LOG_FILE, OVERRIDE_ENV};
use std::fs;
use std::path::PathBuf;
use std::process::Command;

struct Repo(PathBuf);

impl Repo {
    fn new() -> Repo {
        let stamp = std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap()
            .as_nanos();
        let dir = std::env::temp_dir().join(format!("mflow-guard-{}-{stamp}", std::process::id()));
        fs::create_dir_all(&dir).unwrap();
        let repo = Repo(dir);
        repo.git(&["init", "-b", "v2026.11.0"]);
        repo.git(&["config", "user.email", "t@e.x"]);
        repo.git(&["config", "user.name", "t"]);
        repo
    }

    fn git(&self, args: &[&str]) -> std::process::Output {
        Command::new("git")
            .args(args)
            .current_dir(&self.0)
            .output()
            .unwrap()
    }

    fn git_env(&self, env: &str, value: &str, args: &[&str]) -> std::process::Output {
        Command::new("git")
            .args(args)
            .env(env, value)
            .current_dir(&self.0)
            .output()
            .unwrap()
    }

    fn commit(&self, name: &str) -> String {
        fs::write(self.0.join(name), format!("{name}\n")).unwrap();
        self.git(&["add", name]);
        let out = self.git(&["commit", "-m", name]);
        assert!(
            out.status.success(),
            "commit {name}: {}",
            String::from_utf8_lossy(&out.stderr)
        );
        self.rev("HEAD")
    }

    fn rev(&self, rev: &str) -> String {
        let out = self.git(&["rev-parse", rev]);
        assert!(out.status.success(), "rev-parse {rev}");
        String::from_utf8_lossy(&out.stdout).trim().to_string()
    }

    fn install_hook(&self, branch: &str) {
        let hooks = self.0.join(".git").join("hooks");
        fs::create_dir_all(&hooks).unwrap();
        let path = hooks.join("reference-transaction");
        fs::write(&path, hook_script(branch, OVERRIDE_ENV, LOG_FILE)).unwrap();
        #[cfg(unix)]
        {
            use std::os::unix::fs::PermissionsExt;
            fs::set_permissions(&path, fs::Permissions::from_mode(0o755)).unwrap();
        }
    }
}

#[test]
fn a_backward_update_ref_is_refused_with_the_forward_reason() {
    let repo = Repo::new();
    let old = repo.commit("a");
    repo.commit("b");
    repo.install_hook("v2026.11.0");
    let out = repo.git(&["update-ref", "refs/heads/v2026.11.0", &old]);
    assert!(!out.status.success(), "the backward move ran");
    let err = String::from_utf8_lossy(&out.stderr);
    assert!(
        err.contains("forward"),
        "the reason names the forward way: {err}"
    );
    assert!(
        err.contains("revert commit"),
        "the reason names the revert way: {err}"
    );
}

#[test]
fn deletion_is_refused_and_creation_is_allowed() {
    let repo = Repo::new();
    repo.commit("a");
    let tip = repo.commit("b");
    repo.install_hook("v2026.11.0");
    repo.git(&["checkout", "-b", "side"]);
    let out = repo.git(&["update-ref", "-d", "refs/heads/v2026.11.0"]);
    assert!(!out.status.success(), "the deletion ran: {}", String::from_utf8_lossy(&out.stderr));
    let err = String::from_utf8_lossy(&out.stderr);
    assert!(
        err.contains("may not be deleted"),
        "the hook names the deletion, not just a backward move: {err}"
    );
    let out = repo.git(&["update-ref", "refs/heads/seedling", &tip]);
    assert!(
        out.status.success(),
        "another branch is created: {}",
        String::from_utf8_lossy(&out.stderr)
    );
}

#[test]
fn a_fast_forward_and_a_merge_commit_pass() {
    let repo = Repo::new();
    repo.commit("a");
    repo.commit("b");
    repo.install_hook("v2026.11.0");
    repo.git(&["checkout", "-b", "side"]);
    let forward = repo.commit("c");
    repo.git(&["checkout", "v2026.11.0"]);
    let out = repo.git(&["update-ref", "refs/heads/v2026.11.0", &forward]);
    assert!(
        out.status.success(),
        "the fast-forward ran: {}",
        String::from_utf8_lossy(&out.stderr)
    );
    repo.git(&["checkout", "v2026.11.0"]);
    let out = repo.git(&["merge", "--no-ff", "side", "-m", "merge-side"]);
    assert!(
        out.status.success(),
        "the merge commit ran: {}",
        String::from_utf8_lossy(&out.stderr)
    );
}

#[test]
fn another_branch_moves_freely() {
    let repo = Repo::new();
    let old = repo.commit("a");
    repo.commit("b");
    repo.install_hook("v2026.11.0");
    repo.git(&["update-ref", "refs/heads/scratch", "HEAD"]);
    let out = repo.git(&["update-ref", "refs/heads/scratch", &old]);
    assert!(
        out.status.success(),
        "another branch moved backwards: {}",
        String::from_utf8_lossy(&out.stderr)
    );
}

#[test]
fn the_hook_judges_only_the_prepared_phase() {
    let repo = Repo::new();
    repo.commit("a");
    repo.install_hook("v2026.11.0");
    for phase in ["committed", "aborted"] {
        let out = Command::new("sh")
            .arg(".git/hooks/reference-transaction")
            .arg(phase)
            .current_dir(&repo.0)
            .stdin(std::process::Stdio::piped())
            .output();
        let out = out.unwrap();
        assert!(out.status.success(), "the {phase} phase is not judged");
    }
}

#[test]
fn the_override_lets_it_through_and_records_one_line() {
    let repo = Repo::new();
    let old = repo.commit("a");
    repo.commit("b");
    repo.install_hook("v2026.11.0");
    let out = repo.git_env(
        OVERRIDE_ENV,
        "decision-note-1",
        &["update-ref", "refs/heads/v2026.11.0", &old],
    );
    assert!(
        out.status.success(),
        "the override let it through: {}",
        String::from_utf8_lossy(&out.stderr)
    );
    let log = fs::read_to_string(repo.0.join(LOG_FILE)).unwrap();
    let lines: Vec<&str> = log.lines().collect();
    assert_eq!(lines.len(), 1, "one record line: {log:?}");
    assert!(
        log.contains("decision-note-1"),
        "the note reference is recorded: {log:?}"
    );
    assert!(log.contains(&old), "the old value is recorded: {log:?}");
}

#[test]
fn without_the_override_the_refusal_stands() {
    let repo = Repo::new();
    let old = repo.commit("a");
    repo.commit("b");
    repo.install_hook("v2026.11.0");
    let out = repo.git_env(
        OVERRIDE_ENV,
        "",
        &["update-ref", "refs/heads/v2026.11.0", &old],
    );
    assert!(!out.status.success(), "an empty override is no exception");
    assert!(
        !repo.0.join(LOG_FILE).exists(),
        "nothing is logged for a refusal"
    );
}

#[test]
fn the_script_is_pinned_byte_identical() {
    let one = hook_script("v2026.11.0", OVERRIDE_ENV, LOG_FILE);
    let two = hook_script("v2026.11.0", OVERRIDE_ENV, LOG_FILE);
    assert_eq!(one, two, "the same inputs give the same bytes");
    assert!(
        one.contains("refs/heads/v2026.11.0"),
        "the branch is baked in"
    );
    assert!(
        one.contains(OVERRIDE_ENV),
        "the override variable is baked in"
    );
    assert_ne!(
        one,
        hook_script("other", OVERRIDE_ENV, LOG_FILE),
        "another branch differs"
    );
}

#[test]
fn the_guard_refuses_the_listed_commands_on_the_integration_branch() {
    let branch = "v2026.11.0";
    for cmd in [
        "git reset --hard v2026.11.0",
        "git reset --keep v2026.11.0",
        "git push --force origin v2026.11.0",
        "git push --force-with-lease origin v2026.11.0",
        "git branch -f v2026.11.0 main",
        "git branch -D v2026.11.0",
    ] {
        assert!(guard_rule(cmd, branch).is_some(), "{cmd} is refused");
    }
}

#[test]
fn the_guard_leaves_other_branches_and_commands_alone() {
    let branch = "v2026.11.0";
    for cmd in [
        "git reset --hard main",
        "git push origin v2026.11.0",
        "git branch -f other main",
        "git status",
        "git merge --no-ff side",
        "git checkout -B v2026.11.0-origin-base",
    ] {
        assert_eq!(guard_rule(cmd, branch), None, "{cmd} is not refused");
    }
}
