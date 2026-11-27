//! SPEC-20261126-shell-cd-guard: the `tool-preflight` hook refuses a `cd`, `pushd` or `popd`
//! that would move the folder of a session working in a linked worktree of a MoAI project, and
//! leaves every command that does not move it alone. Every case runs the real binary
//! (`hook-dispatch --host claude --event tool-preflight`) against a temporary world made with a
//! real `git worktree add`, under a temporary HOME. No judged command is ever executed: the hook
//! only reads the text.
#![cfg(unix)]
mod common;

use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};

const MODES: [&str; 2] = ["default", "bypassPermissions"];
const PREFIX: &str = "moai-boundary/session-cd: ";

/// Commands that move the folder of the session shell.
const MOVES: [&str; 16] = [
    "(ls) && cd src",
    "time cd src",
    "command -- cd src",
    "cd src",
    "cd",
    "ls && cd src",
    "ls; cd src",
    "false || cd src",
    "ls\ncd src",
    "{ cd src; }",
    "if true; then cd src; fi",
    "X=1 cd src",
    "builtin cd src",
    "eval 'cd src'",
    "pushd src",
    "popd",
];

/// Commands that hold the letters or run a `cd` that leaves the session folder where it is.
const STAYS: [&str; 19] = [
    // A `)` that the lexer takes for the end of a substitution or an arithmetic expansion
    // leaves the parentheses of the line unbalanced: such a line is not split.
    "v=$(cat <<'EOF'\n1) a\ncd src\nEOF\n)",
    "v=$(\n# 1) a\ncd src\n)",
    "( echo $((2*(1+1))); cd src )",
    "f () { cd src; }",
    "echo \"cd src\"",
    "grep 'cd ' a.txt",
    "printf '%s' 'a && cd src'",
    "ls # cd src",
    "cat <<'EOF'\ncd src\nEOF",
    "ls cd",
    "(cd src && ls)",
    "( cd src && ls ) && pwd",
    "v=$(cd src && pwd)",
    "echo `cd src; pwd`",
    "bash -c 'cd src && ls'",
    "sh -c \"cd src\"",
    "command -v cd",
    "/usr/bin/cd src",
    "git -C /tmp status",
];

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

struct World {
    _tmp: tempfile::TempDir,
    home: PathBuf,
    primary: PathBuf,
    own: PathBuf,
    /// A linked worktree of the same repository made inside the primary checkout.
    nested: PathBuf,
    /// A linked worktree of another repository.
    other_wt: PathBuf,
    /// A plain folder that is no checkout.
    plain: PathBuf,
}

impl World {
    /// A primary checkout and a linked worktree `lane/own`. With `profile` the primary holds a
    /// MoAI project profile; `tracked` commits it, so the worktree holds it too.
    fn new(profile: bool, tracked: bool) -> World {
        let tmp = tempfile::tempdir().unwrap();
        let root = fs::canonicalize(tmp.path()).unwrap();
        let primary = root.join("primary");
        let home = root.join("home");
        for dir in [&primary, &home] {
            fs::create_dir_all(dir).unwrap();
        }
        git(&primary, &["init", "-q", "-b", "NeXT"]);
        fs::create_dir_all(primary.join("src")).unwrap();
        fs::write(primary.join("src/x"), "x\n").unwrap();
        if profile {
            fs::create_dir_all(moai_cli::state_root::path(&primary, "")).unwrap();
            fs::write(
                moai_cli::state_root::path(&primary, "project.yaml"),
                "{\"schema_version\":2,\"project_id\":\"demo\"}\n",
            )
            .unwrap();
        }
        if tracked {
            git(&primary, &["add", "-A"]);
        } else {
            git(&primary, &["add", "src"]);
        }
        git(&primary, &["commit", "-q", "-m", "fixture"]);
        let own = root.join("wt-own");
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
        let nested = primary.join(".claude/worktrees/x");
        git(
            &primary,
            &[
                "worktree",
                "add",
                "-q",
                "-b",
                "lane/nested",
                nested.to_str().unwrap(),
            ],
        );
        let other = root.join("other");
        fs::create_dir_all(&other).unwrap();
        git(&other, &["init", "-q", "-b", "main"]);
        fs::write(other.join("y"), "y\n").unwrap();
        git(&other, &["add", "y"]);
        git(&other, &["commit", "-q", "-m", "fixture"]);
        let other_wt = root.join("other-wt");
        git(
            &other,
            &[
                "worktree",
                "add",
                "-q",
                "-b",
                "lane/other",
                other_wt.to_str().unwrap(),
            ],
        );
        let plain = root.join("plain");
        fs::create_dir_all(&plain).unwrap();
        std::os::unix::fs::symlink(&own, home.join("link-own")).unwrap();
        World {
            _tmp: tmp,
            home,
            primary,
            own,
            nested,
            other_wt,
            plain,
        }
    }

    /// Puts a live registry entry of a lane session in the clone: the session id the hook input
    /// carries (`fixture`), mode `agent`, a lease far in the future, registered from `worktree`.
    fn register(&self, worktree: &Path) {
        let dir = moai_cli::clone_root::path(&self.primary.join(".git")).join("sessions");
        fs::create_dir_all(&dir).unwrap();
        let entry = json!({
            "schema_version": 1,
            "name": "agent-1",
            "instance": moai_cli::lane_registry::new_instance(),
            "lease_until": 4_000_000_000_i64,
            "mode": "agent",
            "session": "fixture",
            "worktree": worktree,
        });
        fs::write(dir.join("agent-1.json"), entry.to_string()).unwrap();
    }

    /// One `tool-preflight` call from `cwd` in `mode`; `project_dir` sets `CLAUDE_PROJECT_DIR`.
    /// The deny reason, or `None` when the hook says nothing.
    fn call(
        &self,
        cwd: &Path,
        command: &str,
        mode: &str,
        project_dir: Option<&Path>,
    ) -> Option<String> {
        let input = json!({
            "session_id": "fixture",
            "cwd": cwd,
            "hook_event_name": "PreToolUse",
            "tool_name": "Bash",
            "tool_input": { "command": command },
            "permission_mode": mode,
        });
        let mut cmd = Command::new(common::MOAI_BIN);
        cmd.args([
            "hook-dispatch",
            "--host",
            "claude",
            "--event",
            "tool-preflight",
        ])
        .current_dir(cwd)
        .env("HOME", &self.home)
        .env_remove("CLAUDE_PROJECT_DIR")
        .env_remove("MOAI_FORGE_GH_BIN")
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped());
        if let Some(dir) = project_dir {
            cmd.env("CLAUDE_PROJECT_DIR", dir);
        }
        let mut child = cmd.spawn().unwrap();
        common::feed_stdin(&mut child, input.to_string().as_bytes());
        let out = child.wait_with_output().unwrap();
        assert_eq!(
            out.status.code(),
            Some(0),
            "{command:?}: {}",
            String::from_utf8_lossy(&out.stderr)
        );
        let text = String::from_utf8_lossy(&out.stdout);
        if text.trim().is_empty() {
            return None;
        }
        let value: Value = serde_json::from_str(text.trim())
            .unwrap_or_else(|e| panic!("stdout for {command:?} is not JSON ({e}): {text}"));
        let inner = &value["hookSpecificOutput"];
        assert_eq!(
            inner["permissionDecision"],
            json!("deny"),
            "{command:?}: {value}"
        );
        Some(
            inner["permissionDecisionReason"]
                .as_str()
                .unwrap()
                .to_owned(),
        )
    }
}

/// AC-1 (REQ-1): every way of moving the session folder is denied in a worktree session, with one
/// line that names the other ways; also from a subfolder, from a session the host started in the
/// primary checkout, and where the profile is not tracked into the worktree.
#[test]
fn a_worktree_session_cd_is_denied() {
    let w = World::new(true, true);
    for mode in MODES {
        for command in MOVES {
            let reason = w
                .call(&w.own, command, mode, None)
                .unwrap_or_else(|| panic!("allowed in {mode}: {command:?}"));
            assert!(reason.starts_with(PREFIX), "{command:?}: {reason}");
            assert_eq!(reason.lines().count(), 1, "one line: {reason}");
            for way in [
                "absolute path",
                "--manifest-path",
                "plain single",
                "object id",
            ] {
                assert!(reason.contains(way), "{way} not named: {reason}");
            }
        }
    }
    let from_sub = w.call(&w.own.join("src"), "cd ..", "default", None);
    assert!(from_sub.is_some_and(|r| r.starts_with(PREFIX)), "subfolder");
    let started_in_primary = w.call(&w.own, "cd src", "bypassPermissions", Some(&w.primary));
    assert!(
        started_in_primary.is_some_and(|r| r.starts_with(PREFIX)),
        "a session started in the primary checkout"
    );

    let untracked = World::new(true, false);
    assert!(
        !moai_cli::state_root::path(&untracked.own, "project.yaml").exists(),
        "the profile is not in this worktree"
    );
    let reason = untracked.call(&untracked.own, "cd src", "default", None);
    assert!(
        reason.is_some_and(|r| r.starts_with(PREFIX)),
        "an untracked profile"
    );
}

/// SPEC-less card C-583: the refusal never recommends a form the lane rules forbid. A subshell
/// `( cd <path> && … )`, `git -C`, a bare `cd <path>`, `pushd` and `popd` must not come back into
/// the text, in either permission mode and for every way of moving the session folder.
#[test]
fn the_session_cd_refusal_recommends_no_cd_form() {
    const FORBIDDEN: [&str; 7] = ["( cd", "(cd", "cd <", "cd ..", "git -c", "pushd", "popd"];
    let w = World::new(true, true);
    for mode in MODES {
        for command in MOVES {
            let reason = w
                .call(&w.own, command, mode, None)
                .unwrap_or_else(|| panic!("allowed in {mode}: {command:?}"));
            let text = reason
                .strip_prefix(PREFIX)
                .unwrap_or_else(|| panic!("prefix: {reason}"))
                .to_ascii_lowercase();
            for form in FORBIDDEN {
                assert!(
                    !text.contains(form),
                    "{command:?}: the refusal recommends {form:?}: {reason}"
                );
            }
        }
    }
}

/// AC-7 (REQ-1): the cargo form the refusal suggests is one cargo accepts. The test takes the
/// backtick span that starts with `cargo` out of the text, requires the subcommand before the
/// flag (cargo refuses a flag in front of it), puts a temporary crate in place of `<path>` and
/// runs that exact form from another folder, so no `cd` is needed for it to work.
#[test]
fn the_session_cd_refusal_cargo_form_is_one_cargo_accepts() {
    let w = World::new(true, true);
    let reason = w
        .call(&w.own, "cd src", "default", None)
        .expect("a worktree session cd is denied");
    assert_cargo_accepts(&reason);
}

/// The backtick span of `reason` that starts with `cargo` is a form cargo accepts (see the test
/// above for how it is run).
fn assert_cargo_accepts(reason: &str) {
    let span = reason
        .split('`')
        .skip(1)
        .step_by(2)
        .find(|s| s.starts_with("cargo "))
        .unwrap_or_else(|| panic!("no cargo form in backticks: {reason}"));
    let words: Vec<&str> = span.split_whitespace().collect();
    assert_eq!(words[0], "cargo", "{span}");
    assert!(
        !words[1].starts_with('-'),
        "the subcommand comes before the flag, cargo refuses it the other way round: {span}"
    );
    assert_eq!(words[1], "test", "{span}");
    let flag_at = words
        .iter()
        .position(|word| *word == "--manifest-path")
        .unwrap_or_else(|| panic!("no --manifest-path in {span}"));
    assert!(flag_at > 1, "{span}");
    let value = words[flag_at + 1];
    assert!(value.starts_with("<path>/"), "{span}");

    let tmp = tempfile::tempdir().unwrap();
    let base = fs::canonicalize(tmp.path()).unwrap();
    let krate = base.join("crate");
    let elsewhere = base.join("elsewhere");
    fs::create_dir_all(krate.join("src")).unwrap();
    fs::create_dir_all(&elsewhere).unwrap();
    fs::write(
        krate.join("Cargo.toml"),
        "[package]\nname = \"suggested\"\nversion = \"0.1.0\"\nedition = \"2021\"\n",
    )
    .unwrap();
    fs::write(krate.join("src/lib.rs"), "pub fn one() -> u32 { 1 }\n").unwrap();
    let args: Vec<String> = words[1..]
        .iter()
        .map(|word| word.replace("<path>", krate.to_str().unwrap()))
        .collect();
    let out = common::cargo::cargo_at(common::cargo::cargo_path())
        .args(&args)
        .args(["--no-run", "--offline"])
        .current_dir(&elsewhere)
        .env_remove("CARGO_TARGET_DIR")
        .env_remove("CARGO_BUILD_TARGET_DIR")
        .env_remove("CARGO_INCREMENTAL")
        .stdin(Stdio::null())
        .output()
        .expect("spawn cargo");
    assert!(
        out.status.success(),
        "cargo refuses the suggested form {span:?}: {}",
        String::from_utf8_lossy(&out.stderr)
    );
}

/// AC-2 (REQ-2, REQ-3): letters in a string, a comment or a here-document, a subshell, a
/// substitution, another shell's string and a program called `cd` are not denied.
#[test]
fn a_cd_that_does_not_move_the_session_is_allowed() {
    let w = World::new(true, true);
    for mode in MODES {
        for command in STAYS {
            let got = w.call(&w.own, command, mode, None);
            assert!(
                !got.as_deref().is_some_and(|r| r.starts_with(PREFIX)),
                "denied in {mode}: {command:?}: {got:?}"
            );
        }
    }
}

/// SPEC-20261126-shell-cd-guard revision 2, REQ-2 and REQ-3, AC-6: these inputs leave the
/// session folder where it is, yet the rule denies them because the shell reader does not follow
/// these forms; the test pins the behaviour until the reader does.
#[test]
fn a_known_false_denial_stays_denied() {
    let w = World::new(true, true);
    for mode in MODES {
        for command in [
            "v=$(\n# 1) cd src (\n)",
            "v=$(\n# 1) a; cd src; echo (\n)",
            "v=$(cat <<'EOF'\n1) a\ncd src\n(\nEOF\n)",
            "cat <<EOF$X\nEOF\ncd src\nEOF$X",
        ] {
            let reason = w
                .call(&w.own, command, mode, None)
                .unwrap_or_else(|| panic!("allowed in {mode}: {command:?}"));
            assert!(reason.starts_with(PREFIX), "{command:?}: {reason}");
        }
    }
}

/// The dangerous path of the rule (REQ-1, REQ-4): a session folder reached through a link is
/// judged by the checkout it leads to, and judging a compound command runs nothing of it.
#[test]
fn a_linked_session_folder_is_judged_by_where_it_leads() {
    let w = World::new(true, true);
    let into_own = w.primary.parent().unwrap().join("link-into-own");
    std::os::unix::fs::symlink(w.own.join("src"), &into_own).unwrap();
    let into_primary = w.own.join("link-into-primary");
    std::os::unix::fs::symlink(w.primary.join("src"), &into_primary).unwrap();
    for mode in MODES {
        let got = w.call(&into_own, "ls && cd .. && rm -f src/x", mode, None);
        assert!(
            got.as_deref().is_some_and(|r| r.starts_with(PREFIX)),
            "a link into the worktree, {mode}: {got:?}"
        );
        let got = w.call(&into_primary, "cd .. && ls", mode, None);
        assert!(
            !got.as_deref().is_some_and(|r| r.starts_with(PREFIX)),
            "a link into the primary checkout, {mode}: {got:?}"
        );
    }
    for checkout in [&w.own, &w.primary] {
        assert!(checkout.join("src/x").is_file(), "a judged command ran");
    }
}

/// AC-3 (REQ-4): the primary checkout session, a linked worktree of a folder that is no MoAI
/// project and a command the lexer cannot split keep their `cd`.
#[test]
fn a_session_outside_a_worktree_keeps_cd() {
    let w = World::new(true, true);
    for mode in MODES {
        for command in ["cd src", "ls && cd src", "pushd src"] {
            let got = w.call(&w.primary, command, mode, None);
            assert_eq!(got, None, "primary session, {mode}: {command:?}");
        }
        let got = w.call(&w.own, "case a in a) cd src;; esac", mode, None);
        assert!(
            !got.as_deref().is_some_and(|r| r.starts_with(PREFIX)),
            "unsplittable, {mode}: {got:?}"
        );
    }
    let plain = World::new(false, false);
    for mode in MODES {
        let got = plain.call(&plain.own, "cd src", mode, None);
        assert_eq!(got, None, "no MoAI project, {mode}");
    }
}

/// Whether the deny reason is one of the session-cd rule.
fn is_session_cd(got: &Option<String>) -> bool {
    got.as_deref().is_some_and(|r| r.starts_with(PREFIX))
}

/// SPEC-20261126-shell-cd-guard revision 4, AC-8 (REQ-6): a session in the primary checkout, or in
/// a folder of it, that would move its shell into a linked worktree of the same repository is
/// denied with one line that names `EnterWorktree`, in both permission modes.
#[test]
fn a_main_checkout_cd_into_a_worktree_is_denied() {
    let w = World::new(true, true);
    let own = w.own.display();
    let nested = w.nested.display();
    let from_primary = [
        format!("cd {own}"),
        format!("cd {own}/src"),
        "cd .claude/worktrees/x".to_owned(),
        format!("pushd {own}"),
        format!("ls && cd {own}"),
        format!("cd src && cd {own}"),
        format!("X=1 cd {own}"),
        format!("builtin cd {own}"),
        format!("eval 'cd {own}'"),
        format!("cd {nested}"),
        "cd ~/link-own".to_owned(),
    ];
    let from_sub = [format!("cd {own}"), "cd ../.claude/worktrees/x".to_owned()];
    for mode in MODES {
        for (cwd, commands) in [
            (w.primary.clone(), &from_primary[..]),
            (w.primary.join("src"), &from_sub[..]),
        ] {
            for command in commands {
                let reason = w
                    .call(&cwd, command, mode, None)
                    .unwrap_or_else(|| panic!("allowed in {mode}: {command:?} from {cwd:?}"));
                assert!(reason.starts_with(PREFIX), "{command:?}: {reason}");
                assert_eq!(reason.lines().count(), 1, "one line: {reason}");
                assert!(reason.contains("EnterWorktree"), "{command:?}: {reason}");
            }
        }
    }
}

/// AC-9 (REQ-7, REQ-8): a `cd` that does not enter a linked worktree of the same repository, a
/// target the text does not hold, a subshell and a string given to another shell are not denied
/// by the entering rule.
#[test]
fn a_main_checkout_cd_that_does_not_enter_a_worktree_is_allowed() {
    let w = World::new(true, true);
    let own = w.own.display();
    let primary = w.primary.display();
    let commands = [
        "cd src".to_owned(),
        "cd .".to_owned(),
        format!("cd {primary}/src"),
        format!("cd {}", w.plain.display()),
        format!("cd {primary}/does-not-exist"),
        "cd \"$W\"".to_owned(),
        "cd $W".to_owned(),
        "cd -".to_owned(),
        "cd".to_owned(),
        "cd ~".to_owned(),
        "cd ../*".to_owned(),
        format!("cd {{{own},src}}"),
        format!("(cd {own} && ls)"),
        format!("v=$(cd {own} && pwd)"),
        format!("bash -c 'cd {own}'"),
        format!("echo \"cd {own}\""),
        format!("git -C {own} status"),
        format!("cd {}", w.other_wt.display()),
        format!("pushd {primary}/src"),
        "popd".to_owned(),
    ];
    for mode in MODES {
        for command in &commands {
            let got = w.call(&w.primary, command, mode, None);
            assert!(
                !is_session_cd(&got),
                "denied in {mode}: {command:?}: {got:?}"
            );
        }
    }
}

/// AC-10 (REQ-9): the refusal of a worktree session tells a session that a shell `cd` moved there
/// how to get out, and the entering refusal names the other way in; both are one line, recommend
/// no `cd` form and hold one cargo form that cargo accepts.
#[test]
fn the_session_cd_refusal_tells_a_moved_session_how_to_get_out() {
    const FORBIDDEN: [&str; 7] = ["( cd", "(cd", "cd <", "cd ..", "git -c", "pushd", "popd"];
    let w = World::new(true, true);
    let own = w.own.display();
    let in_worktree = w
        .call(&w.own, "cd src", "default", None)
        .expect("a worktree session cd is denied");
    let entering = w
        .call(&w.primary, &format!("cd {own}"), "default", None)
        .expect("an entering cd is denied");
    for reason in [&in_worktree, &entering] {
        assert!(reason.starts_with(PREFIX), "{reason}");
        assert_eq!(reason.lines().count(), 1, "one line: {reason}");
        assert!(reason.contains("EnterWorktree"), "{reason}");
        assert_eq!(
            reason.matches('`').count(),
            2,
            "one backtick span: {reason}"
        );
        let text = reason.to_ascii_lowercase();
        for form in FORBIDDEN {
            assert!(!text.contains(form), "recommends {form:?}: {reason}");
        }
        assert_cargo_accepts(reason);
    }
    for words in ["end the session", "start it again from the main checkout"] {
        assert!(in_worktree.contains(words), "{words}: {in_worktree}");
    }
    assert!(
        entering.contains("main checkout") && entering.contains("linked worktree"),
        "{entering}"
    );
}

/// AC-11 (REQ-10, default of D-12): a live registry entry of a lane session does not exempt the
/// session from the entering rule, whichever worktree the entry was registered from.
#[test]
fn a_registered_session_is_not_exempt_from_the_entering_cd_rule() {
    let w = World::new(true, true);
    let command = format!("cd {}", w.own.display());
    for registered_from in [&w.primary, &w.own] {
        w.register(registered_from);
        for mode in MODES {
            let got = w.call(&w.primary, &command, mode, None);
            assert!(
                is_session_cd(&got),
                "registered from {registered_from:?}, {mode}: {got:?}"
            );
        }
    }
}

/// The id of the empty tree: a plain object name, so the command text holds no brace pattern.
const EMPTY_TREE: &str = "4b825dc642cb6eb9a060e54bf8d69288fbee4904";

/// AC-12 (REQ-11, default of D-14): the subcommands that write objects only are still refused when
/// they aim at a checkout that is not the session's own, and left alone when they do not.
#[test]
fn an_objects_only_git_command_toward_another_checkout_stays_denied() {
    let w = World::new(true, true);
    let primary = w.primary.display();
    let toward = [
        format!("git -C {primary} checkout x"),
        format!("git -C {primary} commit-tree -m x {EMPTY_TREE}"),
        format!("git -C {primary} hash-object -w src/x"),
        format!("git -C {primary} mktree"),
        format!("git -C {primary} mktag"),
        format!("git -C {primary} write-tree"),
        format!("git --git-dir={primary}/.git commit-tree -m x {EMPTY_TREE}"),
    ];
    let own_checkout = [
        "git commit-tree -m x 4b825dc642cb6eb9a060e54bf8d69288fbee4904",
        "git hash-object -w src/x",
        "git mktree",
        "git mktag",
        "git write-tree",
    ];
    for mode in MODES {
        // The bash boundary is a deny only in a session that skips permission prompts
        // (`bypass_boundary_deny`); the first entry is the control that shows the fixture world
        // refuses a plain git write toward the primary checkout there.
        for command in toward.iter().filter(|_| mode == "bypassPermissions") {
            let reason = w
                .call(&w.own, command, mode, None)
                .unwrap_or_else(|| panic!("allowed in {mode}: {command:?}"));
            assert!(
                reason.starts_with("moai-boundary/git-state: "),
                "{command:?}: {reason}"
            );
        }
        for command in own_checkout {
            let got = w.call(&w.own, command, mode, None);
            assert_eq!(got, None, "{mode}: {command:?}");
        }
    }
}

/// AC-13 (REQ-1, REQ-2, REQ-3): the forms the SPEC lists as already allowed stay allowed for the
/// session that may use them.
#[test]
fn the_read_forms_stay_allowed() {
    let w = World::new(true, true);
    let own = w.own.display();
    let primary = w.primary.display();
    let main_checkout = [
        "cd src".to_owned(),
        format!("cd {primary}/src"),
        "cd /tmp/x".to_owned(),
        "pushd src".to_owned(),
        "popd".to_owned(),
        "cd".to_owned(),
        "cd -".to_owned(),
        "cd \"$X\"".to_owned(),
        format!("(cd {own} && ls)"),
        format!("bash -c 'cd {own}'"),
        format!("git -C {own} status"),
        format!("git -C {own} log -1"),
        format!("git -C {own} diff"),
        format!("git -C {own} rev-parse HEAD"),
        format!("git -C {own} cat-file -p HEAD"),
        format!("git -C {own} ls-tree HEAD"),
        format!("git -C {own} show HEAD"),
        format!("git -C {own} branch --show-current"),
        format!("git -C {own} worktree list"),
        "git add src/x".to_owned(),
        "git fetch origin".to_owned(),
        "git restore src/x".to_owned(),
        "git rm --cached src/x".to_owned(),
    ];
    let worktree = [
        format!("git -C {primary} status"),
        format!("git -C {primary} log -1"),
        format!("git -C {primary} diff"),
        format!("git -C {primary} rev-parse HEAD"),
        format!("git -C {} status", w.nested.display()),
        "git -C /tmp status".to_owned(),
    ];
    for mode in MODES {
        for command in &main_checkout {
            let got = w.call(&w.primary, command, mode, None);
            assert_eq!(got, None, "main checkout, {mode}: {command:?}");
        }
        for command in &worktree {
            let got = w.call(&w.own, command, mode, None);
            assert_eq!(got, None, "worktree, {mode}: {command:?}");
        }
        for command in STAYS {
            let got = w.call(&w.own, command, mode, None);
            assert!(
                !is_session_cd(&got),
                "worktree, {mode}: {command:?}: {got:?}"
            );
        }
    }
}
