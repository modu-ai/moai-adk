//! SPEC-20261123-hook-bash-boundary: `moai hook bash-boundary` (H2). Every case runs the real
//! binary against a temporary world: a primary checkout on the integration branch, linked
//! worktrees (the session's own, a sibling, one on `main`, one nested in the primary), a link that
//! leads out of the session checkout, a scratch folder and a plain folder, all under a temporary
//! HOME. The world lives under this crate's `target/` folder because the system temporary folder
//! is itself a boundary root of the handler. No command of a case is ever executed: the hook only
//! judges the text. The case tables are the only place the bypass shapes are written down.
#![cfg(unix)]
mod common;

use serde_json::{json, Value};
use std::collections::BTreeMap;
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};
use std::time::Instant;

/// A folder that exists on no machine and lies outside every boundary root.
const OUT: &str = "/opt/moai-boundary-outside";
const CLASSES: [&str; 5] = [
    "delete-outside",
    "git-state",
    "protected-ref",
    "hook-skip",
    "unreadable",
];

// ---------------------------------------------------------------- the world

#[derive(Clone, Copy)]
enum S {
    /// The session's own linked worktree, on `lane/own`.
    Own,
    /// A session in a subfolder of its own worktree.
    Sub,
    /// A session in the primary checkout, on the integration branch `NeXT`.
    Pri,
    /// A session in a linked worktree on the protected branch `main`.
    Main,
}

struct World {
    _tmp: tempfile::TempDir,
    root: PathBuf,
    home: PathBuf,
    primary: PathBuf,
    own: PathBuf,
    other: PathBuf,
    inner: PathBuf,
    main_wt: PathBuf,
    elsewhere: PathBuf,
    scratch: PathBuf,
    plain: PathBuf,
}

fn git(dir: &Path, args: &[&str]) {
    let out = Command::new("git")
        .args(["-c", "gc.auto=0", "-c", "maintenance.auto=false"])
        .args([
            "-c",
            "user.name=Fixture",
            "-c",
            "user.email=fixture@example.invalid",
        ])
        .args(args)
        .current_dir(dir)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .env("LC_ALL", "C")
        .stdin(Stdio::null())
        .output()
        .expect("spawn git");
    assert!(
        out.status.success(),
        "git {args:?} failed: {}",
        String::from_utf8_lossy(&out.stderr)
    );
}

/// Whether a path lies under a temporary root of the handler (D-3). The world must lie outside
/// every one of them, or the folders the cases call "outside" would be inside.
fn under_a_temporary_root(path: &Path) -> bool {
    ["/tmp", "/private/tmp", "/var/tmp", "/private/var/tmp"]
        .iter()
        .any(|root| path.starts_with(root))
        || path.starts_with("/var/folders")
        || path.starts_with("/private/var/folders")
}

/// Where the world is built: this crate's `target/` folder, unless the checkout itself lies under
/// a temporary root (an audit copy in `/tmp`); then a folder under the user's cache folder, which
/// is removed with the world. The result never depends on where the repository sits.
fn fixture_base() -> PathBuf {
    let manifest = fs::canonicalize(env!("CARGO_MANIFEST_DIR")).unwrap();
    let mut base = manifest.join("target").join("hook-bash-boundary-tests");
    if under_a_temporary_root(&base) {
        let home = std::env::var_os("HOME")
            .map(PathBuf::from)
            .expect("HOME names the folder for a world outside the temporary roots");
        base = fs::canonicalize(&home)
            .unwrap_or(home)
            .join(".cache")
            .join("moai-hook-bash-boundary-tests");
    }
    assert!(
        !under_a_temporary_root(&base),
        "no folder outside the temporary roots for the fixture world: {}",
        base.display()
    );
    base
}

impl World {
    fn new() -> World {
        let base = fixture_base();
        fs::create_dir_all(&base).unwrap();
        let tmp = tempfile::Builder::new()
            .prefix("hbb-")
            .tempdir_in(&base)
            .unwrap();
        let root = fs::canonicalize(tmp.path()).unwrap();
        let primary = root.join("primary");
        let home = root.join("home");
        let elsewhere = root.join("elsewhere");
        let scratch = root.join("scratch");
        let plain = root.join("plain");
        for dir in [&primary, &home, &elsewhere, &scratch, &plain] {
            fs::create_dir_all(dir).unwrap();
        }
        fs::write(elsewhere.join("x"), "x").unwrap();
        git(&primary, &["init", "-q", "-b", "NeXT"]);
        fs::create_dir_all(moai_cli::state_root::path(&primary, "")).unwrap();
        fs::write(
            moai_cli::state_root::path(&primary, "project.yaml"),
            "{\"schema_version\":2,\"project_id\":\"demo\",\"work_flow\":{\"flow\":\"m-flow\",\"integration_branch\":\"NeXT\"}}\n",
        )
        .unwrap();
        fs::create_dir_all(primary.join("src")).unwrap();
        fs::create_dir_all(primary.join("build")).unwrap();
        fs::write(primary.join("src/a.txt"), "a\n").unwrap();
        fs::write(primary.join("build/x"), "x\n").unwrap();
        git(&primary, &["add", "-A"]);
        git(&primary, &["commit", "-q", "-m", "fixture"]);
        git(&primary, &["branch", "main"]);
        let own = root.join("wt-own");
        let other = root.join("wt-other");
        let main_wt = root.join("wt-main");
        let inner = primary.join(".claude/worktrees/inner");
        git(
            &primary,
            &[
                "worktree",
                "add",
                "-q",
                "-b",
                "lane/own",
                own.to_str().unwrap(),
            ],
        );
        git(
            &primary,
            &[
                "worktree",
                "add",
                "-q",
                "-b",
                "lane/other",
                other.to_str().unwrap(),
            ],
        );
        git(
            &primary,
            &["worktree", "add", "-q", main_wt.to_str().unwrap(), "main"],
        );
        git(
            &primary,
            &[
                "worktree",
                "add",
                "-q",
                "-b",
                "lane/inner",
                inner.to_str().unwrap(),
            ],
        );
        std::os::unix::fs::symlink(&elsewhere, own.join("escape")).unwrap();
        World {
            _tmp: tmp,
            root,
            home,
            primary,
            own,
            other,
            inner,
            main_wt,
            elsewhere,
            scratch,
            plain,
        }
    }

    fn fill(&self, text: &str) -> String {
        let pairs: [(&str, String); 10] = [
            ("{OWN}", self.own.display().to_string()),
            ("{PRI}", self.primary.display().to_string()),
            ("{OTH}", self.other.display().to_string()),
            ("{INNER}", self.inner.display().to_string()),
            ("{MAIN}", self.main_wt.display().to_string()),
            ("{ESC}", self.own.join("escape").display().to_string()),
            ("{ELSE}", self.elsewhere.display().to_string()),
            ("{SCR}", self.scratch.display().to_string()),
            ("{HOME}", self.home.display().to_string()),
            ("{OUT}", OUT.to_owned()),
        ];
        let mut out = text.to_owned();
        for (key, value) in pairs {
            out = out.replace(key, &value);
        }
        out
    }

    /// The folder the hook process starts in and the `cwd` of the hook input for a session.
    fn session(&self, s: S) -> (PathBuf, PathBuf) {
        match s {
            S::Own => (self.own.clone(), self.own.clone()),
            S::Sub => (self.own.clone(), self.own.join("src")),
            S::Pri => (self.primary.clone(), self.primary.clone()),
            S::Main => (self.main_wt.clone(), self.main_wt.clone()),
        }
    }

    fn input(&self, command: &Value, cwd: &Path, mode: &str) -> Value {
        json!({
            "session_id": "fixture",
            "cwd": cwd,
            "scratchpad_dir": self.scratch,
            "permission_mode": mode,
            "hook_event_name": "PreToolUse",
            "tool_name": "Bash",
            "tool_input": { "command": command },
        })
    }

    fn spawn(&self, proc_cwd: &Path, stdin: &[u8], project_dir: Option<&Path>) -> Run {
        let mut cmd = Command::new(common::MOAI_BIN);
        cmd.args(["hook", "bash-boundary"])
            .current_dir(proc_cwd)
            .env("HOME", &self.home)
            .env_remove("CLAUDE_PROJECT_DIR")
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped());
        if let Some(dir) = project_dir {
            cmd.env("CLAUDE_PROJECT_DIR", dir);
        }
        let mut child = cmd.spawn().unwrap();
        common::feed_stdin(&mut child, stdin);
        let out = child.wait_with_output().unwrap();
        Run {
            code: out.status.code().unwrap(),
            stdout: String::from_utf8_lossy(&out.stdout).into_owned(),
            stderr: String::from_utf8_lossy(&out.stderr).into_owned(),
        }
    }

    fn run(&self, s: S, command: &str, mode: &str) -> Run {
        let (proc_cwd, cwd) = self.session(s);
        let input = self.input(&Value::String(self.fill(command)), &cwd, mode);
        self.spawn(&proc_cwd, input.to_string().as_bytes(), None)
    }

    fn verdict(&self, s: S, command: &str) -> Option<String> {
        let run = self.run(s, command, "default");
        parse_verdict(&run, command)
    }
}

struct Run {
    code: i32,
    stdout: String,
    stderr: String,
}

/// `None` is no decision; `Some(class)` is a deny with that class word. Any other output fails.
fn parse_verdict(run: &Run, what: &str) -> Option<String> {
    assert_eq!(run.code, 0, "exit code for {what:?}: {}", run.stderr);
    assert!(run.stderr.is_empty(), "stderr for {what:?}: {}", run.stderr);
    if run.stdout.trim().is_empty() {
        return None;
    }
    let value: Value = serde_json::from_str(run.stdout.trim())
        .unwrap_or_else(|e| panic!("stdout for {what:?} is not JSON ({e}): {}", run.stdout));
    let object = value.as_object().expect("an object");
    assert_eq!(object.len(), 1, "one top key for {what:?}");
    let inner = object["hookSpecificOutput"]
        .as_object()
        .expect("hookSpecificOutput");
    let mut keys: Vec<&str> = inner.keys().map(String::as_str).collect();
    keys.sort_unstable();
    assert_eq!(
        keys,
        [
            "hookEventName",
            "permissionDecision",
            "permissionDecisionReason"
        ],
        "keys for {what:?}"
    );
    assert_eq!(inner["hookEventName"], "PreToolUse");
    assert_eq!(
        inner["permissionDecision"], "deny",
        "only deny is ever printed"
    );
    let reason = inner["permissionDecisionReason"]
        .as_str()
        .expect("a reason");
    let rest = reason
        .strip_prefix("moai-boundary/")
        .unwrap_or_else(|| panic!("reason prefix for {what:?}: {reason}"));
    let class = rest.split(':').next().unwrap().to_owned();
    assert!(
        CLASSES.contains(&class.as_str()),
        "class word {class:?} for {what:?}"
    );
    Some(class)
}

// -------------------------------------------------------------- case tables

struct Case {
    s: S,
    cmd: String,
    want: Option<&'static str>,
}

fn a(s: S, cmd: &str) -> Case {
    Case {
        s,
        cmd: cmd.to_owned(),
        want: None,
    }
}

fn d(s: S, cmd: &str, class: &'static str) -> Case {
    Case {
        s,
        cmd: cmd.to_owned(),
        want: Some(class),
    }
}

fn run_table(w: &World, name: &str, cases: &[Case]) {
    let allowed = cases.iter().filter(|c| c.want.is_none()).count();
    let denied = cases.len() - allowed;
    assert!(
        allowed >= 12 && denied >= 12,
        "{name}: {allowed} allowed and {denied} denied cases"
    );
    let mut bad = Vec::new();
    for case in cases {
        let got = w.verdict(case.s, &case.cmd);
        if got.as_deref() != case.want {
            bad.push(format!(
                "  {:?}: want {:?}, got {:?}",
                case.cmd, case.want, got
            ));
        }
    }
    assert!(
        bad.is_empty(),
        "{name}: {} of {} cases differ\n{}",
        bad.len(),
        cases.len(),
        bad.join("\n")
    );
}

fn delete_cases() -> Vec<Case> {
    use S::{Own, Pri, Sub};
    let de = "delete-outside";
    vec![
        a(Own, "rm -rf {OWN}/build"),
        a(Own, "rm -rf build"),
        a(Own, "rm -f src/a.txt"),
        a(Own, "rm -rf /tmp/moai-x/cache"),
        a(Own, "rm -rf /private/tmp/moai-x"),
        a(Own, "rm -rf {SCR}/work"),
        a(Own, "cd {OWN}/src && rm -rf ../build"),
        a(Own, "rm -rf build/{a,b}"),
        a(Own, "rm -rf ./build ./dist"),
        a(Own, "rm -rf /private/var/folders/ab/cdef12/T/moai-xyz"),
        a(Own, "rmdir build"),
        a(Own, "find . -name '*.o' -delete"),
        a(Own, "find build -type f -delete"),
        a(Own, "find /tmp/moai-x -exec rm -f {} +"),
        a(Own, "rm -rf \"{OWN}/build dir\""),
        a(Own, "ls -la && echo done"),
        a(Own, "rm -rf build/*"),
        a(Own, "(cd {OWN}/src; rm -f a.txt)"),
        a(Sub, "rm -rf ../build"),
        a(Own, "rm -rf /tmp/moai-x*"),
        a(Pri, "rm -rf {PRI}/build"),
        a(Own, "rm -f {ESC}"),
        d(Own, "rm -rf {OUT}/x", de),
        d(Own, "rm -rf /", de),
        d(Own, "rm -rf /*", de),
        d(Own, "rm -rf {OWN}", de),
        d(Own, "rm -rf .", de),
        d(Own, "rm -rf {OWN}/.git", de),
        d(Own, "rm -rf {OWN}/..", de),
        d(Own, "rm -rf {OTH}/src", de),
        d(Own, "rm -rf {PRI}/src", de),
        d(Own, "rm -rf {ESC}/x", de),
        d(Own, "rm -rf {ESC}/", de),
        d(Own, "rm -rf ~/x", de),
        d(Own, "rm -rf /tmp", de),
        d(Own, "rm -rf /tmp/../etc", de),
        d(Own, "cd {OUT} && rm -rf x", de),
        d(Own, "cd /tmp && rm -rf ../etc", de),
        d(Pri, "rm -rf {PRI}/.claude/worktrees", de),
        d(Pri, "rm -rf {INNER}", de),
        d(Pri, "rm -rf {PRI}", de),
        d(Own, "find / -delete", de),
        d(Own, "find . -delete", de),
        d(Own, "find {OUT} -name x -exec rm -rf {} +", de),
        d(Own, "rm -rf /tmp/*", de),
        d(Own, "rm -rf {OWN}/src/../../x", de),
        d(Own, "rm -rf .*", de),
        d(Own, "shred {OUT}/y", de),
        d(Own, "unlink {OUT}/y", de),
        d(Own, "rm -rf {OWN}/.git/hooks", de),
        d(Own, "rm -rf {OWN}/build /etc/x", de),
    ]
}

fn git_state_cases() -> Vec<Case> {
    use S::{Own, Pri};
    let gs = "git-state";
    vec![
        a(Own, "git status"),
        a(Own, "git log --oneline -3"),
        a(Own, "git diff"),
        a(Own, "git add src/a.txt"),
        a(Own, "git commit -m x"),
        a(Own, "git checkout -b feature/x"),
        a(Own, "git -C {OWN} status"),
        a(Own, "git -C {OTH} status"),
        a(Own, "git -C {PRI} log -1"),
        a(Own, "git -C {OTH} diff"),
        a(Own, "git stash list"),
        a(Own, "git branch"),
        a(Own, "git branch --show-current"),
        a(Own, "git worktree list"),
        a(Own, "git -C {PRI} rev-parse HEAD"),
        a(Own, "git switch -c feat/y"),
        a(Own, "git reset --hard HEAD"),
        a(Own, "git stash"),
        a(Own, "git rebase main"),
        a(Own, "git clean -fd"),
        a(Own, "git merge main"),
        a(Pri, "git status"),
        a(Pri, "git add src/a.txt"),
        a(Pri, "git fetch origin"),
        a(Pri, "git restore src/a.txt"),
        a(Pri, "git rm --cached src/a.txt"),
        a(Pri, "git diff --stat"),
        a(Pri, "git log -1"),
        d(Own, "git -C {OTH} commit -m x", gs),
        d(Own, "git -C {PRI} checkout x", gs),
        d(Own, "git -C {PRI} reset --hard", gs),
        d(Own, "cd {OTH} && git add src/a.txt", gs),
        d(
            Own,
            "git --git-dir={PRI}/.git --work-tree={PRI} checkout x",
            gs,
        ),
        d(Own, "git --work-tree={OTH} add .", gs),
        d(Own, "GIT_WORK_TREE={OTH} git add .", gs),
        d(Own, "git worktree remove {OTH}", gs),
        d(Own, "git -C {INNER} reset --hard", gs),
        d(Own, "git -C {PRI} stash", gs),
        d(Own, "git -C {OTH} clean -fd", gs),
        d(Own, "git -C {PRI} merge x", gs),
        d(Pri, "git -C {OWN} add .", gs),
        d(Pri, "git checkout -b x", gs),
        d(Pri, "git switch other", gs),
        d(Pri, "git reset --hard HEAD~1", gs),
        d(Pri, "git rebase main", gs),
        d(Pri, "git stash", gs),
        d(Pri, "git stash push", gs),
        d(Pri, "git clean -fd", gs),
        d(Pri, "git merge x", gs),
        d(Pri, "git pull", gs),
        d(Pri, "git cherry-pick abc", gs),
        d(Pri, "git revert abc", gs),
        d(Pri, "git am p.patch", gs),
        d(Pri, "git checkout -- src/a.txt", gs),
    ]
}

fn protected_ref_cases() -> Vec<Case> {
    use S::{Main, Own, Pri};
    let pr = "protected-ref";
    vec![
        a(Own, "git branch feature/new"),
        a(Own, "git branch -D lane/old"),
        a(Own, "git push origin lane/own"),
        a(Own, "git push -u origin HEAD"),
        a(Own, "git push origin lane/own:lane/own"),
        a(Own, "git branch -m old new"),
        a(Own, "git update-ref refs/heads/lane/own HEAD"),
        a(Own, "git fetch origin"),
        a(Own, "git fetch origin NeXT:refs/remotes/origin/NeXT"),
        a(Own, "git checkout -B lane/x"),
        a(Own, "git switch -c feat"),
        a(Own, "git push origin --tags"),
        a(Own, "git commit -m \"mention NeXT and update-ref\""),
        a(Own, "git log NeXT..HEAD"),
        a(Own, "git branch --list NeXT"),
        a(Own, "git diff NeXT"),
        a(Own, "git merge NeXT"),
        a(Own, "git push origin lane/own --force-with-lease"),
        a(Own, "git switch NeXT"),
        a(Main, "git status"),
        a(Main, "git log -1"),
        a(Main, "git add src/a.txt"),
        a(Main, "git reset -- src/a.txt"),
        a(Pri, "git add src/a.txt"),
        d(Own, "git update-ref refs/heads/NeXT HEAD", pr),
        d(Own, "git update-ref -d refs/heads/NeXT", pr),
        d(Own, "git branch -f NeXT HEAD", pr),
        d(Own, "git branch -D NeXT", pr),
        d(Own, "git branch -m NeXT old", pr),
        d(Own, "git branch -M main", pr),
        d(Own, "git push origin NeXT", pr),
        d(Own, "git push origin HEAD:NeXT", pr),
        d(Own, "git push origin :NeXT", pr),
        d(Own, "git push --force origin main", pr),
        d(Own, "git push origin +x:master", pr),
        d(Own, "git push --delete origin NeXT", pr),
        d(Own, "git fetch origin x:NeXT", pr),
        d(Own, "git checkout -B NeXT", pr),
        d(Own, "git switch -C main", pr),
        d(Own, "git symbolic-ref refs/heads/NeXT refs/heads/x", pr),
        d(Own, "git push --mirror", pr),
        d(Own, "git push --all", pr),
        d(Own, "git branch --force master", pr),
        d(Own, "git branch -f nExT HEAD", pr),
        d(Own, "git push origin refs/heads/NeXT", pr),
        d(Own, "git worktree add -B NeXT {OWN}/../x", pr),
        d(Main, "git commit -m x", pr),
        d(Main, "git reset HEAD~1", pr),
        d(Main, "git merge x", pr),
        d(Main, "git pull", pr),
        d(Main, "git rebase x", pr),
        d(Main, "git push", pr),
        d(Main, "git push origin HEAD", pr),
        d(Main, "git cherry-pick abc", pr),
        d(Main, "git revert abc", pr),
        d(Main, "git commit --amend --no-edit", pr),
        d(Pri, "git commit -m x", pr),
        d(Pri, "git commit --amend", pr),
        d(Pri, "git push", pr),
    ]
}

fn hook_skip_cases() -> Vec<Case> {
    use S::Own;
    let hs = "hook-skip";
    vec![
        a(Own, "git commit -m x"),
        a(Own, "git commit -am x"),
        a(Own, "git commit -m n"),
        a(Own, "git commit -mn"),
        a(Own, "git commit -m \"--no-verify\""),
        a(Own, "git commit -m \"fix: never use --no-verify\""),
        a(Own, "git commit --amend --no-edit"),
        a(Own, "git commit --allow-empty -m x"),
        a(Own, "git push -n origin lane/own"),
        a(Own, "git merge -n main"),
        a(Own, "git config user.name x"),
        a(Own, "git config --get core.hooksPath"),
        a(Own, "git config core.hooksPath"),
        a(Own, "git -c user.name=x commit -m y"),
        a(Own, "git commit --verify -m x"),
        a(Own, "git commit -F msg.txt"),
        a(Own, "git commit -S -m x"),
        a(Own, "git commit -m \"$(cat <<'EOF'\nsubject\n\nbody with --no-verify words\nEOF\n)\""),
        d(Own, "git commit --no-verify -m x", hs),
        d(Own, "git commit -n -m x", hs),
        d(Own, "git commit -nm x", hs),
        d(Own, "git commit -an -m x", hs),
        d(Own, "git commit --no-ver -m x", hs),
        d(Own, "git commit --no-v -m x", hs),
        d(Own, "git -c core.hooksPath=/dev/null commit -m x", hs),
        d(Own, "git -c core.hookspath=x commit -m x", hs),
        d(Own, "git config core.hooksPath /x", hs),
        d(Own, "git config --global core.hooksPath /x", hs),
        d(Own, "git config --unset core.hooksPath", hs),
        d(Own, "git push --no-verify origin lane/own", hs),
        d(Own, "git merge --no-verify main", hs),
        d(Own, "git rebase --no-verify main", hs),
        d(Own, "git cherry-pick --no-verify abc", hs),
        d(Own, "GIT_CONFIG_COUNT=1 GIT_CONFIG_KEY_0=core.hooksPath GIT_CONFIG_VALUE_0=/x git commit -m x", hs),
        d(Own, "export GIT_CONFIG_PARAMETERS=\"'core.hookspath'='/x'\"; git commit -m x", hs),
        d(Own, "git --config-env=core.hooksPath=HP commit -m x", hs),
        d(Own, "git commit -a --no-verify", hs),
    ]
}

fn unreadable_cases() -> Vec<Case> {
    use S::Own;
    let ur = "unreadable";
    vec![
        a(Own, "echo $HOME"),
        a(Own, "ls $X"),
        a(Own, "cat $(ls)"),
        a(Own, "git commit -m \"$(date)\""),
        a(Own, "git commit -m \"$MSG\""),
        a(Own, "git log --since=$S"),
        a(Own, "git status $X"),
        a(Own, "cargo test --test foo"),
        a(Own, "grep -r \"$PATTERN\" src"),
        a(Own, "cat <<'EOF'\nrm -rf $X\nEOF"),
        a(Own, "echo \"rm -rf $X\""),
        a(Own, "git add \"$FILE\""),
        a(Own, "git commit -F \"$F\""),
        a(Own, "cp $A $B"),
        a(Own, "moai card next"),
        a(Own, "echo 'rm -rf /'"),
        a(Own, "printf '%s\\n' \"rm -rf /\""),
        a(Own, "git commit -m 'rm -rf /'"),
        a(Own, ": ${X:=y}"),
        a(Own, "FOO=$(pwd) cargo build"),
        a(Own, "bash script.sh"),
        a(Own, "source env.sh"),
        a(Own, "./tools/run.sh $ARG"),
        a(Own, "git -C $D status"),
        a(Own, "git log $(git rev-parse HEAD~1)"),
        d(Own, "rm -rf $X", ur),
        d(Own, "rm -rf \"$HOME/x\"", ur),
        d(Own, "rm -rf ${X}/y", ur),
        d(Own, "rm -rf $(echo /x)", ur),
        d(Own, "rm -rf `pwd`/x", ur),
        d(Own, "rm -rf build/$(echo x)", ur),
        d(Own, "cd $X && rm -rf y", ur),
        d(
            Own,
            "cd \"$(git rev-parse --show-toplevel)\" && rm -rf build",
            ur,
        ),
        d(Own, "$X -rf y", ur),
        d(Own, "git $SUB status", ur),
        d(Own, "git push origin $B", ur),
        d(Own, "git -C $D commit -m x", ur),
        d(Own, "git commit $F", ur),
        d(Own, "git branch -D $B", ur),
        d(Own, "git ci -m x", ur),
        d(Own, "eval \"$X\"", ur),
        d(Own, "eval rm -rf $X", ur),
        d(Own, "bash -c \"$X\"", ur),
        d(Own, "sh -c 'rm -rf $X'", ur),
        d(Own, "xargs rm -rf", ur),
        d(Own, "find . -name x | xargs rm", ur),
        d(Own, "echo x | sh", ur),
        d(Own, "echo x | bash -s", ur),
        d(Own, "rm -rf $(pwd)/../x", ur),
        d(Own, "X=/ ; rm -rf $X", ur),
        d(Own, "export D=/; rm -rf $D/etc", ur),
        d(Own, "for d in a b; do rm -rf $d; done", ur),
        d(Own, "rm -rf {OWN}/build $EXTRA", ur),
        d(Own, "GIT_DIR=$D git commit -m x", ur),
        d(Own, "git commit -m x $EXTRA", ur),
        d(Own, "rm -rf \"build", ur),
        d(Own, "f() { rm -rf x; }; f", ur),
        d(Own, "case $x in a) rm -rf y;; esac", ur),
        d(Own, "rm -rf <(echo x)", ur),
        d(Own, "GIT_CONFIG_GLOBAL=/tmp/x git commit -m y", ur),
    ]
}

/// Every denied command of the tables, for the pre-scan superset check.
fn all_cases() -> Vec<Case> {
    let mut all = delete_cases();
    all.extend(git_state_cases());
    all.extend(protected_ref_cases());
    all.extend(hook_skip_cases());
    all.extend(unreadable_cases());
    all.extend(audit_denied_cases());
    all
}

// ------------------------------------------------------------- AC-1 .. AC-3

#[test]
fn ac1_delete_table() {
    run_table(&World::new(), "delete", &delete_cases());
}

#[test]
fn ac1_git_state_table() {
    run_table(&World::new(), "git-state", &git_state_cases());
}

#[test]
fn ac1_protected_ref_table() {
    run_table(&World::new(), "protected-ref", &protected_ref_cases());
}

#[test]
fn ac1_hook_skip_table() {
    run_table(&World::new(), "hook-skip", &hook_skip_cases());
}

#[test]
fn ac1_unreadable_table() {
    run_table(&World::new(), "unreadable", &unreadable_cases());
}

#[test]
fn ac2_outside_a_moai_project_nothing_is_printed_for_any_case() {
    let w = World::new();
    for case in all_cases().iter().filter(|c| c.want.is_some()).step_by(3) {
        let input = w.input(&Value::String(w.fill(&case.cmd)), &w.plain, "default");
        let run = w.spawn(&w.plain, input.to_string().as_bytes(), None);
        assert_eq!(run.code, 0, "{:?}", case.cmd);
        assert!(
            run.stdout.is_empty(),
            "stdout for {:?}: {}",
            case.cmd,
            run.stdout
        );
        assert!(
            run.stderr.is_empty(),
            "stderr for {:?}: {}",
            case.cmd,
            run.stderr
        );
    }
}

#[test]
fn ac2_a_denial_has_exactly_the_three_fields_and_an_allowed_case_prints_nothing() {
    let w = World::new();
    let denied = w.run(S::Own, "rm -rf {OUT}/x", "default");
    assert_eq!(
        parse_verdict(&denied, "deny").as_deref(),
        Some("delete-outside")
    );
    assert_eq!(denied.stdout.trim().lines().count(), 1, "one line of JSON");
    assert!(!denied.stdout.contains("\"ask\"") && !denied.stdout.contains("\"allow\""));
    let allowed = w.run(S::Own, "ls -la", "default");
    assert!(allowed.stdout.is_empty() && allowed.stderr.is_empty() && allowed.code == 0);
}

#[test]
fn ac3_other_tools_and_other_input_shapes_get_no_decision() {
    let w = World::new();
    let (proc_cwd, cwd) = w.session(S::Own);
    let bad = "rm -rf /opt/moai-boundary-outside/x";
    let mut inputs: Vec<Value> = Vec::new();
    let mut other_tool = w.input(&json!(bad), &cwd, "default");
    other_tool["tool_name"] = json!("Edit");
    inputs.push(other_tool);
    for command in [
        json!(null),
        json!(7),
        json!([bad]),
        json!({"c": bad}),
        json!(""),
    ] {
        inputs.push(w.input(&command, &cwd, "default"));
    }
    let mut missing = w.input(&json!(bad), &cwd, "default");
    missing["tool_input"]
        .as_object_mut()
        .unwrap()
        .remove("command");
    inputs.push(missing);
    for input in inputs {
        let run = w.spawn(&proc_cwd, input.to_string().as_bytes(), None);
        assert_eq!(parse_verdict(&run, &input.to_string()), None, "{input}");
    }
}

// ------------------------------------------------------------- AC-4 .. AC-7

#[test]
fn ac4_boundary_roots_and_a_literal_cd() {
    let w = World::new();
    let cases = [
        a(S::Own, "rm -rf {SCR}/deep/er"),
        a(S::Own, "rm -rf /var/tmp/moai-x"),
        a(S::Sub, "cd .. && rm -rf build"),
        a(S::Own, "cd /tmp && rm -rf moai-x"),
        a(S::Own, "pushd {OWN}/src && rm -rf ../build; popd"),
        d(S::Own, "rm -rf {SCR}", "delete-outside"),
        d(S::Own, "rm -rf /var/tmp", "delete-outside"),
        d(S::Own, "cd /tmp; rm -rf ../x", "delete-outside"),
        d(S::Sub, "cd ../.. && rm -rf wt-other", "delete-outside"),
        d(S::Own, "rm -rf {ELSE}/x", "delete-outside"),
        d(S::Own, "rm -rf {ESC}/x", "delete-outside"),
        d(S::Own, "rm -rf {OWN}/.git", "delete-outside"),
        d(S::Own, "cd {OWN} && rm -rf .git", "delete-outside"),
        d(
            S::Own,
            "cd {OWN}/src && rm -rf ../../wt-other",
            "delete-outside",
        ),
    ];
    for case in cases {
        let got = w.verdict(case.s, &case.cmd);
        assert_eq!(got.as_deref(), case.want, "{:?}", case.cmd);
    }
}

#[test]
fn ac4_the_project_folder_names_the_checkout_when_cwd_left_the_repository() {
    let w = World::new();
    // The session started in its own worktree (CLAUDE_PROJECT_DIR) and the shell moved into a
    // sibling worktree: the sibling is a checkout of the same repository, so it counts as the
    // session checkout there (the gap R-3 of the SPEC). A folder of another repository does not.
    let foreign = w.root.join("foreign");
    fs::create_dir_all(&foreign).unwrap();
    git(&foreign, &["init", "-q", "-b", "main"]);
    fs::create_dir_all(foreign.join("data")).unwrap();
    fs::write(foreign.join("data/f"), "f").unwrap();
    git(&foreign, &["add", "-A"]);
    git(&foreign, &["commit", "-q", "-m", "foreign"]);
    let input = w.input(&json!("rm -rf data"), &foreign, "default");
    let run = w.spawn(&w.own, input.to_string().as_bytes(), Some(&w.own));
    assert_eq!(
        parse_verdict(&run, "foreign cwd").as_deref(),
        Some("delete-outside")
    );
    let input = w.input(&json!("rm -rf build"), &w.other, "default");
    let run = w.spawn(&w.own, input.to_string().as_bytes(), Some(&w.own));
    assert_eq!(parse_verdict(&run, "sibling cwd"), None);
}

#[test]
fn ac5_git_state_by_session_kind() {
    let w = World::new();
    let cases = [
        a(S::Own, "git -C {OTH} log -1"),
        a(S::Own, "git -C {OTH} status --short"),
        a(S::Own, "git -C {OWN} commit -m x"),
        d(S::Own, "git -C {OTH} add .", "git-state"),
        d(
            S::Own,
            "git --git-dir {PRI}/.git --work-tree {PRI} reset --hard",
            "git-state",
        ),
        d(S::Own, "cd {PRI} && git checkout -b x", "git-state"),
        d(S::Pri, "git reset --hard", "git-state"),
        d(S::Pri, "git stash", "git-state"),
        d(S::Pri, "git switch -c x", "git-state"),
    ];
    for case in cases {
        let want = case.want.map(|_| "git-state");
        let got = w.verdict(case.s, &case.cmd);
        assert_eq!(got.as_deref(), want, "{:?}", case.cmd);
    }
}

#[test]
fn ac6_protected_names_and_the_checked_out_branch() {
    let w = World::new();
    for name in ["NeXT", "main", "master", "next", "MAIN"] {
        let cmd = format!("git branch -f {name} HEAD");
        assert_eq!(
            w.verdict(S::Own, &cmd).as_deref(),
            Some("protected-ref"),
            "{cmd}"
        );
    }
    for name in ["lane/own", "feature/NeXT-notes", "mainline", "release/1"] {
        let cmd = format!("git branch -f {name} HEAD");
        assert_eq!(w.verdict(S::Own, &cmd), None, "{cmd}");
    }
    assert_eq!(
        w.verdict(S::Main, "git commit -m x").as_deref(),
        Some("protected-ref")
    );
    assert_eq!(w.verdict(S::Own, "git commit -m x"), None);
}

#[test]
fn ac7_every_spelling_and_every_hook_running_subcommand() {
    let w = World::new();
    for sub in [
        "commit",
        "merge",
        "push",
        "rebase",
        "cherry-pick",
        "revert",
        "am",
        "pull",
    ] {
        for flag in ["--no-verify", "--no-verif", "--no-ver", "--no-v"] {
            let cmd = format!("git {sub} {flag} lane/own");
            assert_eq!(
                w.verdict(S::Own, &cmd).as_deref(),
                Some("hook-skip"),
                "{cmd}"
            );
        }
    }
    for cmd in [
        "git commit -m x -n",
        "git commit -qn -m x",
        "git commit -a -n",
        "git -c core.hooksPath=/x -c user.name=y commit -m z",
    ] {
        assert_eq!(
            w.verdict(S::Own, cmd).as_deref(),
            Some("hook-skip"),
            "{cmd}"
        );
    }
    for cmd in [
        "git commit -mn",
        "git commit -m \"-n\"",
        "git commit -m x -- src/a.txt",
        "git commit --no-edit --amend",
        "git commit --verify",
        "git push -n origin lane/own",
        "git log --no-verify-signatures",
    ] {
        assert_eq!(w.verdict(S::Own, cmd), None, "{cmd}");
    }
}

// -------------------------------------------------------------------- AC-8

fn nested(inner: &str) -> Vec<String> {
    let q = inner.replace('\'', "'\\''");
    vec![
        format!("sudo {inner}"),
        format!("sudo -u root {inner}"),
        format!("env {inner}"),
        format!("env FOO=1 {inner}"),
        format!("env -i {inner}"),
        format!("command {inner}"),
        format!("nohup {inner}"),
        format!("timeout 5 {inner}"),
        format!("timeout -s KILL 5 {inner}"),
        format!("nice -n 5 {inner}"),
        format!("time {inner}"),
        format!("bash -c '{q}'"),
        format!("sh -c '{q}'"),
        format!("zsh -lc '{q}'"),
        format!("eval '{q}'"),
        format!("eval {inner}"),
        format!("echo $({inner})"),
        format!("echo `{inner}`"),
        format!("({inner})"),
        format!("{{ {inner}; }}"),
        format!("if true; then {inner}; fi"),
        format!("for i in 1 2; do {inner}; done"),
        format!("while false; do {inner}; done"),
        format!("true && {inner}"),
        format!("false || {inner}"),
        format!("true | {inner}"),
        format!("{inner} &"),
        format!("true\n{inner}"),
        format!("bash <<EOF\n{inner}\nEOF"),
        format!("cat <(true) <({inner})"),
        format!("find . -name x -exec sh -c '{q}' \\;"),
        format!("ls $(true) ; {inner}"),
        format!("cat <<EOF\n$({inner})\nEOF"),
    ]
}

#[test]
fn ac8_nesting_is_judged_like_the_bare_command() {
    let w = World::new();
    let mut bad = Vec::new();
    for cmd in nested("rm -rf {OUT}/x") {
        let got = w.verdict(S::Own, &cmd);
        if got.as_deref() != Some("delete-outside") {
            bad.push(format!("  {cmd:?} -> {got:?}"));
        }
    }
    for cmd in nested("ls build") {
        let got = w.verdict(S::Own, &cmd);
        if got.is_some() {
            bad.push(format!("  {cmd:?} (benign) -> {got:?}"));
        }
    }
    for cmd in [
        "cat <<'EOF'\nrm -rf {OUT}/x\nEOF",
        "cat <<EOF\nrm -rf {OUT}/x\nEOF",
        "echo 'rm -rf {OUT}/x'",
        "# rm -rf {OUT}/x",
        "ls # rm -rf {OUT}/x",
    ] {
        let got = w.verdict(S::Own, cmd);
        if got.is_some() {
            bad.push(format!("  {cmd:?} (data) -> {got:?}"));
        }
    }
    assert!(
        bad.is_empty(),
        "{} nesting cases differ\n{}",
        bad.len(),
        bad.join("\n")
    );
}

#[test]
fn ac8_the_depth_the_count_and_the_size_limits_take_their_boundary() {
    let w = World::new();
    let nest = |n: usize| format!("{}true{}", "echo $(".repeat(n), ")".repeat(n));
    assert_eq!(w.verdict(S::Own, &nest(8)), None, "8 levels");
    assert_eq!(
        w.verdict(S::Own, &nest(9)).as_deref(),
        Some("unreadable"),
        "9 levels"
    );
    let commands = |n: usize| format!("git status{}", ";true".repeat(n - 1));
    assert_eq!(w.verdict(S::Own, &commands(2048)), None, "2048 commands");
    assert_eq!(
        w.verdict(S::Own, &commands(2049)).as_deref(),
        Some("unreadable"),
        "2049 commands"
    );
    let sized = |bytes: usize| {
        let head = "git status; echo ";
        format!("{head}{}", "a".repeat(bytes - head.len()))
    };
    assert_eq!(w.verdict(S::Own, &sized(262_144)), None, "256 KiB");
    assert_eq!(
        w.verdict(S::Own, &sized(262_145)).as_deref(),
        Some("unreadable"),
        "256 KiB + 1"
    );
}

// -------------------------------------------------------------------- AC-9

#[test]
fn ac9_reasons_carry_the_class_and_no_command_byte() {
    let w = World::new();
    let marker = "ZZMARKER31337";
    let probes = [
        ("rm -rf {OUT}/ZZMARKER31337", "delete-outside"),
        ("git -C {OTH} commit -m ZZMARKER31337", "git-state"),
        ("git push origin NeXT ZZMARKER31337", "protected-ref"),
        ("git commit --no-verify -m ZZMARKER31337", "hook-skip"),
        ("rm -rf $ZZMARKER31337", "unreadable"),
    ];
    for (cmd, class) in probes {
        let run = w.run(S::Own, cmd, "default");
        assert_eq!(parse_verdict(&run, cmd).as_deref(), Some(class), "{cmd}");
        assert!(
            !run.stdout.contains(marker),
            "the reason echoes the command: {}",
            run.stdout
        );
        let value: Value = serde_json::from_str(run.stdout.trim()).unwrap();
        let reason = value["hookSpecificOutput"]["permissionDecisionReason"]
            .as_str()
            .unwrap();
        assert!(reason.len() < 400, "{} bytes: {reason}", reason.len());
        assert!(
            reason.starts_with(&format!("moai-boundary/{class}:")),
            "{reason}"
        );
        assert!(
            !reason.contains(&w.root.display().to_string()),
            "no fixture path: {reason}"
        );
    }
}

// ------------------------------------------------------------------- AC-10

#[test]
fn ac10_the_module_makes_no_process_network_environment_or_write_call() {
    let source =
        fs::read_to_string(Path::new(env!("CARGO_MANIFEST_DIR")).join("src/hook_bash_boundary.rs"))
            .unwrap();
    let code: String = source
        .lines()
        .filter(|l| !l.trim_start().starts_with("//"))
        .collect::<Vec<_>>()
        .join("\n");
    for word in [
        "std::process",
        "Command::new",
        "std::net",
        "TcpStream",
        "std::env",
        "env::var",
        "fs::write",
        "File::create",
        "OpenOptions",
        "remove_file",
        "remove_dir",
        "create_dir",
        "fs::rename",
        "fs::copy",
        "set_permissions",
    ] {
        assert!(
            !code.contains(word),
            "src/hook_bash_boundary.rs names {word}"
        );
    }
}

/// Every entry of the tree with its kind, size and a hash of its bytes (a link: its target), so a
/// rewrite that keeps the size, an emptied folder and a changed link all show.
fn tree_state(root: &Path) -> BTreeMap<String, String> {
    fn fnv(bytes: &[u8]) -> u64 {
        bytes.iter().fold(0xcbf2_9ce4_8422_2325, |h, b| {
            (h ^ u64::from(*b)).wrapping_mul(0x0100_0000_01b3)
        })
    }
    fn walk(base: &Path, dir: &Path, out: &mut BTreeMap<String, String>) {
        let Ok(entries) = fs::read_dir(dir) else {
            return;
        };
        for entry in entries.flatten() {
            let path = entry.path();
            let meta = fs::symlink_metadata(&path).unwrap();
            let rel = path
                .strip_prefix(base)
                .unwrap()
                .to_string_lossy()
                .into_owned();
            if meta.file_type().is_symlink() {
                let target = fs::read_link(&path).unwrap();
                out.insert(rel, format!("link {}", target.display()));
            } else if meta.is_dir() {
                out.insert(rel, "dir".to_owned());
                walk(base, &path, out);
            } else {
                let bytes = fs::read(&path).unwrap_or_default();
                out.insert(
                    rel,
                    format!("file {} bytes, hash {:016x}", meta.len(), fnv(&bytes)),
                );
            }
        }
    }
    let mut out = BTreeMap::new();
    walk(root, root, &mut out);
    out
}

#[test]
fn ac10_no_case_writes_a_file_and_the_mode_never_changes_the_answer() {
    let w = World::new();
    let before = tree_state(&w.root);
    let probes = [
        (S::Own, "rm -rf {OUT}/x"),
        (S::Own, "git push origin NeXT"),
        (S::Own, "git commit --no-verify -m x"),
        (S::Own, "rm -rf $X"),
        (S::Own, "rm -rf build"),
        (S::Pri, "git stash"),
    ];
    for (s, cmd) in probes {
        let mut seen = Vec::new();
        for mode in [
            "default",
            "plan",
            "acceptEdits",
            "auto",
            "dontAsk",
            "bypassPermissions",
        ] {
            let run = w.run(s, cmd, mode);
            seen.push(parse_verdict(&run, cmd));
        }
        assert!(seen.windows(2).all(|p| p[0] == p[1]), "{cmd}: {seen:?}");
    }
    assert_eq!(tree_state(&w.root), before, "the fixture changed");
    for dir in [&w.home, &w.primary, &w.own] {
        assert!(!dir.join(".claude/settings.json").exists());
        assert!(!dir.join(".claude/settings.local.json").exists());
    }
}

// ------------------------------------------------------------------- AC-11

#[test]
fn ac11_the_prescan_never_hides_a_trigger_word() {
    let w = World::new();
    for case in all_cases().iter().filter(|c| c.want.is_some()) {
        let filled = w.fill(&case.cmd);
        assert!(
            moai_cli::hook_bash_boundary::needs_walk(&filled),
            "the pre-scan skips a denied command: {filled:?}"
        );
    }
    for harmless in [
        "ls -la",
        "cargo test --test foo",
        "echo hi",
        "moai card next",
        "pwd",
    ] {
        assert!(
            !moai_cli::hook_bash_boundary::needs_walk(harmless),
            "{harmless}"
        );
    }
    for disguised in [
        "r\"\"m -rf x",
        "r\\m -rf x",
        "'g'it status",
        "ls $X",
        "/bin/r? -rf x",
        "/???/?? -rf x",
        "/bin/r[m] -rf x",
        "{rm,-rf,x}",
        "=rm -rf x",
        "ls; =rm -rf x",
        "$\"rm\" -rf x",
    ] {
        assert!(
            moai_cli::hook_bash_boundary::needs_walk(disguised),
            "{disguised}"
        );
    }
}

fn percentile(sorted: &[f64], share: f64) -> f64 {
    let rank = (share * sorted.len() as f64).ceil() as usize;
    sorted[rank.clamp(1, sorted.len()) - 1]
}

#[test]
fn ac11_measured_latency_is_printed_and_only_a_tenfold_miss_fails() {
    let w = World::new();
    let load = Command::new("uptime")
        .output()
        .map(|o| String::from_utf8_lossy(&o.stdout).trim().to_owned())
        .unwrap_or_default();
    let profile = if cfg!(debug_assertions) {
        "debug"
    } else {
        "release"
    };
    for (label, s, cmd) in [
        ("allowed", S::Own, "git status"),
        ("denied", S::Own, "rm -rf {OUT}/x"),
    ] {
        let (proc_cwd, cwd) = w.session(s);
        let input = w
            .input(&Value::String(w.fill(cmd)), &cwd, "default")
            .to_string();
        let mut ms = Vec::new();
        for _ in 0..200 {
            let start = Instant::now();
            let run = w.spawn(&proc_cwd, input.as_bytes(), None);
            ms.push(start.elapsed().as_secs_f64() * 1000.0);
            assert_eq!(run.code, 0);
        }
        ms.sort_by(|x, y| x.partial_cmp(y).unwrap());
        let (p50, p90) = (percentile(&ms, 0.5), percentile(&ms, 0.9));
        println!(
            "AC-11 {label}: p50 {p50:.1} ms, p90 {p90:.1} ms, max {:.1} ms, 200 samples, {profile} build, load: {load}",
            ms[ms.len() - 1]
        );
        assert!(
            p50 <= 200.0 && p90 <= 400.0,
            "{label}: p50 {p50:.1} ms, p90 {p90:.1} ms"
        );
    }
}

// ------------------------------------------------------------------- AC-14
//
// The repair table of the independent audit (AUDIT #403, note 7629). Every shape the audit ran
// against the real binary is here with the answer this SPEC wants: the rule is that a command
// word or a folder that the text does not fix statically is `unreadable`, and a command that
// runs another command is judged as that command. The audit's own probe lists (the wrongly
// allowed ones and the list that must stay allowed) are copied whole.

fn audit_denied_cases() -> Vec<Case> {
    use S::{Main, Own, Pri};
    let (de, gs, pr, hs, ur) = (
        "delete-outside",
        "git-state",
        "protected-ref",
        "hook-skip",
        "unreadable",
    );
    vec![
        // The probes the first head already answered (kept as regression).
        d(Own, "rm -rf {OUT}/x; ls", de),
        d(Own, "ls && rm -rf {OUT}/x", de),
        d(Own, "ls | rm -rf {OUT}/x", de),
        d(Own, "ls || rm -rf {OUT}/x", de),
        d(Own, "ls & rm -rf {OUT}/x", de),
        d(Own, "ls |& rm -rf {OUT}/x", de),
        d(Own, "echo $(rm -rf {OUT}/x)", de),
        d(Own, "echo \"$(rm -rf {OUT}/x)\"", de),
        d(Own, "echo `rm -rf {OUT}/x`", de),
        d(Own, "(rm -rf {OUT}/x)", de),
        d(Own, "{ rm -rf {OUT}/x; }", de),
        d(Own, "env rm -rf {OUT}/x", de),
        d(Own, "env -i PATH=/bin rm -rf {OUT}/x", de),
        d(Own, "command rm -rf {OUT}/x", de),
        d(Own, "builtin command rm -rf {OUT}/x", de),
        d(Own, "exec rm -rf {OUT}/x", de),
        d(Own, "sudo -n rm -rf {OUT}/x", de),
        d(Own, "/bin/rm -rf {OUT}/x", de),
        d(Own, "\\rm -rf {OUT}/x", de),
        d(Own, "'rm' -rf {OUT}/x", de),
        d(Own, "r''m -rf {OUT}/x", de),
        d(Own, "   \n\n  rm -rf {OUT}/x", de),
        d(Own, "\trm -rf {OUT}/x", de),
        d(Own, "bash -c 'rm -rf {OUT}/x'", de),
        d(Own, "/bin/bash -c \"ls; rm -rf {OUT}/x\"", de),
        d(Own, "zsh -c 'cd {OUT} && rm -rf y'", de),
        d(Own, "bash <<'EOF'\nrm -rf {OUT}/x\nEOF", de),
        d(Own, "bash <<< 'rm -rf {OUT}/x'", de),
        d(Own, "echo 'rm -rf {OUT}/x' | bash", ur),
        d(Own, "cd {OUT}; rm -rf ./*", de),
        d(Own, "cd -P {ESC} && rm -rf ./*", de),
        d(Own, "rm -rf {ESC}/", de),
        d(Own, "rm -rf escape/*", de),
        d(Own, "rm -rf ../wt-other", de),
        d(Own, "find {OUT} -delete", de),
        d(Own, "find {OUT} -exec rm {} \\;", de),
        d(Own, "find . -exec sh -c 'rm -rf {OUT}/x' \\;", de),
        d(Own, "xargs -0 rm < list", ur),
        d(Own, "nice -n 5 timeout 5 nohup rm -rf {OUT}/x", de),
        d(Own, "eval 'rm -rf {OUT}/x'", de),
        d(Own, "f() { rm -rf {OUT}/x; }; f", ur),
        d(Own, "if true; then rm -rf {OUT}/x; fi", de),
        d(Own, "while true; do rm -rf {OUT}/x; break; done", de),
        d(Own, "for i in 1; do rm -rf {OUT}/x; done", de),
        d(Own, "! rm -rf {OUT}/x", de),
        d(Own, "time rm -rf {OUT}/x", de),
        d(Own, "$'rm' -rf {OUT}/x", de),
        // B1: a command word that an expansion decides is unreadable.
        d(Own, "/bin/r? -rf {OUT}/x", ur),
        d(Own, "/bin/r[m] -rf {OUT}/x", ur),
        d(Own, "/???/?? -rf {OUT}/x", ur),
        d(Own, "/bin/r* -rf {OUT}/x", ur),
        d(Own, "~/bin/r? -rf {OUT}/x", ur),
        d(Own, "bash -c '/bin/r? -rf {OUT}/x'", ur),
        d(Own, "=rm -rf {OUT}/x", ur),
        d(Own, "$\"rm\" -rf {OUT}/x", ur),
        d(Own, "bash -c '{rm,-rf,{OUT}/x}'", ur),
        d(Own, "{rm,-rf,{OUT}/x}", ur),
        d(Own, "r{m,} -rf {OUT}/x", ur),
        d(Own, "sudo /bin/r? -rf {OUT}/x", ur),
        d(Own, "env /bin/r? -rf {OUT}/x", ur),
        d(Own, "nohup /bin/r? -rf {OUT}/x", ur),
        d(Own, "time {rm,-rf,{OUT}/x}", ur),
        d(Own, "find {OWN} -exec /bin/r? {} +", ur),
        d(Own, "echo {OUT}/x | xargs /bin/r? -rf", ur),
        d(Own, "CDPATH=/opt cd moai-boundary-outside && rm -rf ./*", ur),
        d(
            Own,
            "export CDPATH=/opt; cd moai-boundary-outside; rm -rf ./*",
            ur,
        ),
        // B2: the directory stack and CDPATH are not followed; the folder becomes unreadable.
        d(Own, "pushd {OUT} && popd +1 && rm -rf ./*", ur),
        d(Own, "pushd {OUT} && popd -n && rm -rf ./*", ur),
        d(Own, "pushd {OUT} && pushd {OWN} && pushd +1 && rm -rf ./*", ur),
        d(Own, "pushd -n {OUT}; rm -rf ./*", ur),
        d(Own, "popd -0; rm -rf build", ur),
        d(Own, "pushd; rm -rf build", ur),
        // B3: a protected ref changed by a spelling the first table did not hold.
        d(Main, "git push origin @", pr),
        d(Main, "git update-ref HEAD HEAD~1", pr),
        d(Pri, "git update-ref HEAD HEAD~1", pr),
        d(Main, "git update-ref -d HEAD", pr),
        d(Main, "git branch -m renamed", pr),
        d(Main, "git branch -M renamed", pr),
        d(Own, "git rebase lane/own NeXT", pr),
        d(Own, "git rebase --onto HEAD~1 HEAD main", pr),
        d(Own, "git rebase --onto NeXT lane/own main", pr),
        d(Own, "git filter-branch -f --msg-filter cat -- main", pr),
        d(Own, "git filter-branch -f -- --all", pr),
        d(Main, "git filter-branch -f --msg-filter cat", pr),
        d(Own, "git checkout --track origin/main", pr),
        d(Own, "git checkout -t origin/NeXT", pr),
        d(Own, "git switch --track origin/master", pr),
        d(Own, "git switch -t origin/main", pr),
        d(Own, "git push origin HEAD:refs/heads/NeXT", pr),
        d(Own, "git push . HEAD:NeXT", pr),
        d(Own, "git fetch . HEAD:main", pr),
        d(Own, "git push origin x:Main", pr),
        d(Own, "git push -f origin lane/own:master", pr),
        d(Own, "git branch --copy lane/own main", pr),
        d(Own, "git checkout -bmain", pr),
        d(Own, "git -C {PRI} update-ref refs/heads/NeXT HEAD", gs),
        d(
            Own,
            "GIT_DIR={PRI}/.git git update-ref refs/heads/lane/x HEAD",
            gs,
        ),
        d(Own, "cd {PRI} && git stash", gs),
        d(Own, "cd {OTH}; git reset --hard", gs),
        d(Own, "git -C ../wt-other commit -m x", gs),
        d(Own, "(cd ../wt-other && git commit -m x)", gs),
        d(Own, "bash -c 'cd ../wt-other && git commit -m x'", gs),
        d(Own, "env GIT_WORK_TREE={OTH} git add .", gs),
        d(Own, "export GIT_DIR={PRI}/.git; git checkout x", gs),
        d(Own, "git worktree remove --force ../wt-other", gs),
        // B4: a hooks path set through every spelling of `git config`, and the include route.
        d(Own, "git config --type path core.hooksPath /dev/null", hs),
        d(Own, "git config --type=path core.hooksPath /dev/null", hs),
        d(Own, "git config set core.hooksPath /dev/null", hs),
        d(Own, "git config --local core.hooksPath /dev/null", hs),
        d(Own, "git config --global core.hooksPath /x", hs),
        d(Own, "git config --type path --add core.hooksPath /x", hs),
        d(Own, "git config include.path /opt/x.cfg", hs),
        d(Own, "git config --add include.path /opt/x.cfg", hs),
        d(Own, "git config includeIf.gitdir:/x/.path /opt/x.cfg", hs),
        d(Own, "git -c Core.HooksPath=/x commit -m x", hs),
        d(Own, "git commit -anm x", hs),
        d(Own, "git commit -vn -m x", hs),
        d(Own, "git commit --no-verif -m x", hs),
        d(
            Own,
            "export GIT_CONFIG_COUNT=1 GIT_CONFIG_KEY_0=core.hooksPath GIT_CONFIG_VALUE_0=/x; git commit -m x",
            hs,
        ),
        d(
            Own,
            "env GIT_CONFIG_PARAMETERS=\"'core.hooksPath'='/x'\" git commit -m x",
            hs,
        ),
        d(Own, "GIT_CONFIG_GLOBAL=/opt/x git commit -m x", ur),
        d(Own, "git push --no-verify", hs),
        d(Own, "git am --no-verify p.patch", hs),
        d(Own, "git commit -m \"$(date)\" --no-verify", hs),
        d(Own, "git ci -m x", ur),
        d(Own, "git $(echo commit) --no-verify", ur),
        // B5: a command that runs another command is judged as that command.
        d(Own, "echo {OUT}/x | xargs -J % rm -rf %", ur),
        d(Own, "echo {OUT}/x | xargs -R 1 -I % rm -rf %", ur),
        d(Own, "echo {OUT}/x | xargs -S 255 rm -rf", ur),
        d(Own, "arch -arm64 rm -rf {OUT}/x", de),
        d(Own, "arch -arch arm64 rm -rf {OUT}/x", de),
        d(Own, "arch -x86_64 -e A=b rm -rf {OUT}/x", de),
        d(Own, "script -q /dev/null rm -rf {OUT}/x", de),
        d(Own, "script -q -c 'rm -rf {OUT}/x' /dev/null", de),
        d(Own, "script --command='rm -rf {OUT}/x' /dev/null", de),
        d(Own, "git rebase -x 'rm -rf {OUT}/x' HEAD~1", de),
        d(Own, "git rebase --exec 'rm -rf {OUT}/x' main", de),
        d(Own, "git rebase --exec='rm -rf {OUT}/x' main", de),
        d(Own, "git rebase -x'rm -rf {OUT}/x' main", de),
        d(Own, "git rebase -i -x 'ls; rm -rf {OUT}/x' main", de),
        d(Own, "git rebase -x \"$CMD\" main", ur),
        d(Own, "git difftool -y -x 'rm -rf {OUT}/x' HEAD~1", de),
        d(Own, "git difftool --extcmd='rm -rf {OUT}/x' HEAD~1", de),
        d(Own, "git submodule foreach 'rm -rf {OUT}/x'", de),
        d(Own, "git submodule foreach --recursive 'rm -rf {OUT}/x'", de),
        d(Own, "git submodule foreach \"$CMD\"", ur),
        d(Own, "git bisect run rm -rf {OUT}/x", de),
        d(Own, "git bisect run sh -c 'rm -rf {OUT}/x'", de),
        d(Own, "git bisect run $CMD", ur),
        d(
            Own,
            "git filter-branch --tree-filter 'rm -rf {OUT}/x' -- lane/own",
            de,
        ),
        d(Own, "git filter-branch --msg-filter \"$CMD\" -- lane/own", ur),
        // N3: the scratch folder itself is a root and never a target.
        d(Own, "rm -rf {SCR}", de),
        d(Own, "rm -rf {SCR}/", de),
        d(Own, "rm -rf {SCR}/.", de),
        // N2: other checkouts stay denied for the commands that are now judged by subcommand.
        d(Own, "git -C {OTH} lfs pull", gs),
        d(Own, "git lfs frobnicate", ur),
    ]
}

fn audit_allowed_cases() -> Vec<Case> {
    use S::{Own, Pri};
    vec![
        // The audit's list of commands that must stay allowed (no false positive).
        a(Own, "rm -rf {OWN}/build"),
        a(Own, "echo \"rm -rf {OUT}\""),
        a(Own, "grep -r 'rm -rf' src"),
        a(Own, "git commit -m 'drop rm -rf {OUT} and --no-verify'"),
        a(Own, "cat {OUT}/rm/git/notes.txt"),
        a(Own, "ls /opt/firmware/sh"),
        a(Own, "cargo build --release"),
        a(
            Own,
            "cargo test -p moai-cli --test hook_bash_boundary_cli",
        ),
        a(Own, "rg 'git push' docs"),
        a(Own, "echo $HOME"),
        a(Own, "git log --format='%H %s' -3"),
        a(Own, "git status && git diff --stat"),
        a(Own, "rm -f /tmp/foo.log"),
        a(
            Own,
            "moai card note C-1 \"$(cat <<'EOF'\nrm -rf / and git push origin NeXT\nEOF\n)\"",
        ),
        a(
            Own,
            "cat > notes.md <<EOF\nuse git push --no-verify never\nEOF",
        ),
        a(Own, "date -u +%Y-%m-%dT%H:%M:%SZ"),
        a(Own, "ps aux | grep cargo"),
        a(Own, "git push origin @"),
        a(Own, "git --git-dir={OTH}/../primary/.git status"),
        a(Own, "git worktree prune"),
        a(Pri, "git merge-tree --write-tree NeXT lane/own"),
        a(Own, "git merge-tree --write-tree NeXT HEAD"),
        a(Own, "git lfs pull"),
        a(Own, "git lfs ls-files"),
        a(Own, "rm -rf target/debug"),
        a(Pri, "git worktree add -b lane/new .claude/worktrees/new NeXT"),
        a(
            Own,
            "env -u CLAUDE_CODE_SESSION_ID RUSTC_WRAPPER=sccache moai heavy-slot-acquire --max 4 --name x -- cargo nextest run --test a",
        ),
        a(
            Own,
            "git -C {OWN} log --oneline -3 && git -C {PRI} rev-parse NeXT",
        ),
        a(Own, "sqlite3 -readonly ~/.moai/x.db 'select 1'"),
        a(Own, "git show NeXT:REVIEW.md"),
        a(Own, "git stash push -u -m tag"),
        a(
            Own,
            "git commit -m x --trailer 'Co-Authored-By: a <b>'",
        ),
        // Guards against the new rules over-blocking an ordinary command.
        a(Own, "[ -d build ] && ls"),
        a(Own, "[[ -f src/a.txt ]] && echo ok"),
        a(Own, "test -f src/a.txt"),
        a(Own, "ls src/*.txt"),
        a(Own, "ls src/a.t?t"),
        a(Own, "echo ?"),
        a(Own, "cd {OWN}/src && ls ../build"),
        a(Own, "pushd {OWN}/src && popd && rm -rf build"),
        a(Own, "git rebase NeXT"),
        a(Own, "git rebase -x 'cargo test' NeXT"),
        a(Own, "git rebase -i NeXT"),
        a(Own, "git bisect start"),
        a(Own, "git bisect run cargo test"),
        a(Own, "git submodule update --init"),
        a(Own, "git submodule foreach 'git status'"),
        a(Own, "git difftool HEAD~1"),
        a(Own, "git difftool -x 'diff -u' HEAD~1"),
        a(Own, "git filter-branch --msg-filter cat -- lane/own"),
        a(Own, "arch -x86_64 cargo build"),
        a(Own, "script -q /dev/null cargo test"),
        a(Own, "echo x | xargs -J % echo %"),
        a(Own, "git branch -m renamed"),
        a(Own, "git branch -m lane/own lane/renamed"),
        a(Own, "git update-ref HEAD HEAD~1"),
        a(Own, "git checkout --track origin/feature/x"),
        a(Own, "git config --get include.path"),
        a(Own, "git config user.name x"),
        a(Own, "git config --type bool core.filemode true"),
        a(Own, "git config --type path --get core.excludesfile"),
        a(Own, "rm -rf {SCR}/work"),
    ]
}

#[test]
fn ac14_a_scratchpad_root_under_a_temporary_root_is_still_never_a_target() {
    let w = World::new();
    let scratch = "/tmp/moai-bb-scratch-fixture";
    let ask = |cmd: &str| {
        let (proc_cwd, cwd) = w.session(S::Own);
        let mut input = w.input(&Value::String(w.fill(cmd)), &cwd, "default");
        input["scratchpad_dir"] = Value::String(scratch.to_owned());
        let run = w.spawn(&proc_cwd, input.to_string().as_bytes(), None);
        parse_verdict(&run, cmd)
    };
    for cmd in [
        "rm -rf /tmp/moai-bb-scratch-fixture",
        "rm -rf /tmp/moai-bb-scratch-fixture/",
        "rm -rf /tmp/moai-bb-scratch-fixture/.",
        "cd /tmp/moai-bb-scratch-fixture && rm -rf .",
    ] {
        assert_eq!(ask(cmd).as_deref(), Some("delete-outside"), "{cmd}");
    }
    for cmd in [
        "rm -rf /tmp/moai-bb-scratch-fixture/work",
        "rm -rf /tmp/moai-bb-scratch-fixture/work/x",
        "rm -rf /tmp/moai-bb-other",
    ] {
        assert_eq!(ask(cmd), None, "{cmd}");
    }
}

#[test]
fn ac14_the_audit_repair_table_denied_and_allowed_shapes() {
    let mut cases = audit_denied_cases();
    cases.extend(audit_allowed_cases());
    run_table(&World::new(), "audit-repair", &cases);
}

#[test]
fn ac14_deep_nesting_is_unreadable_and_never_overflows_the_stack() {
    let w = World::new();
    let shapes = [
        (
            "quoted substitutions in substitutions",
            format!("{}x{}", "echo \"$(".repeat(15_000), ")\"".repeat(15_000)),
        ),
        (
            "double quote then substitution",
            format!(
                "git status {}x{}",
                "\"$(".repeat(30_000),
                ")\"".repeat(30_000)
            ),
        ),
        (
            "open parentheses",
            format!("git status {}", "(".repeat(50_000)),
        ),
    ];
    for (label, cmd) in &shapes {
        assert!(cmd.len() < 262_144, "{label}: {} bytes", cmd.len());
        let run = w.run(S::Own, cmd, "default");
        assert_eq!(
            parse_verdict(&run, label).as_deref(),
            Some("unreadable"),
            "{label}"
        );
    }
}
