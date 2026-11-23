//! SPEC-20261123-base-agents-discovery-record (REQ-1 .. REQ-6): the installed Claude Code finds
//! the five base agents and a project role in a native config folder and in a folder shaped
//! like an Orca profile (its `agents` entry links to the user's agents folder), and no name is
//! hidden. Every case runs the built binary and the real `claude` on temporary folders with a
//! cleared environment and no credential; the real HOME, `~/.claude` and the Orca profile are
//! never read or written. A case is skipped (and says so) when `claude` is not on the PATH.
#![cfg(unix)]

mod common;
use common::{kill_group, spawn_group, CleanEnv};
use serde_json::Value;
use std::{
    collections::{BTreeMap, BTreeSet},
    ffi::OsString,
    fs,
    os::unix::fs::symlink,
    path::{Path, PathBuf},
    process::{Command, Stdio},
    thread,
    time::{Duration, Instant},
};
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;
const PERMISSION: &str = "fixture-approved";
/// Stops the child's background network traffic (updates, telemetry) that is not needed for a
/// name check: with it unset an unauthenticated run still tried DNS lookups.
const NO_NONESSENTIAL_TRAFFIC: (&str, &str) = ("CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC", "1");
/// What a known agent name must end in here: the temporary config folder holds no credential,
/// so the run stops at the login step. Anything else (a model answer) means a credential reached
/// the child and the run was a real, billed model call.
const LOGIN_STOP: &str = "Not logged in";
/// The longest one child `claude` may run; a name check ends in about a second.
const CHILD_DEADLINE: Duration = Duration::from_secs(60);
/// Variables a POSIX shell (and macOS's launcher) adds to its own environment on start, so the
/// fake program below sees them although the test set none of them.
const SHELL_ADDED: [&str; 5] = ["PWD", "OLDPWD", "SHLVL", "_", "__CF_USER_TEXT_ENCODING"];
const BASE: [&str; 5] = [
    "moai-audit",
    "moai-docs",
    "moai-impl",
    "moai-plan",
    "moai-research",
];
const PROJECT_ROLE: &str = "prj-probe";
const PERSONAL: &str = "my-own";
const UNKNOWN: &str = "no-such-agent-for-the-list";

/// The `claude` program on `path`, if any.
fn find_claude(path: Option<OsString>) -> Option<PathBuf> {
    let paths = path?;
    std::env::split_paths(&paths)
        .map(|d| d.join("claude"))
        .find(|p| p.is_file())
}

fn claude_or_skip() -> Option<PathBuf> {
    let found = find_claude(std::env::var_os("PATH"));
    if found.is_none() {
        eprintln!("skipped: `claude` is not on the PATH");
    }
    found
}

fn copy_tree(from: &Path, to: &Path) {
    fs::create_dir_all(to).unwrap();
    for e in fs::read_dir(from).unwrap() {
        let e = e.unwrap();
        let target = to.join(e.file_name());
        if e.file_type().unwrap().is_dir() {
            copy_tree(&e.path(), &target);
        } else {
            fs::copy(e.path(), target).unwrap();
        }
    }
}

fn agent_file(name: &str) -> String {
    format!("---\nname: {name}\ndescription: probe agent {name}\ntools: Read\n---\n\nProbe.\n")
}

/// `<program> --agent <name> -p hi` in `project`: the environment is cleared, so only the
/// variables set here reach it (no credential among them), and the background traffic is off.
fn claude_command(
    program: &Path,
    project: &Path,
    home: &Path,
    config_dir: &Path,
    name: &str,
) -> Command {
    let mut cmd = Command::new(program);
    cmd.clean_env()
        .args(["--agent", name, "-p", "hi"])
        .current_dir(project)
        .env("PATH", std::env::var_os("PATH").unwrap_or_default())
        .env("HOME", home)
        .env("CLAUDE_CONFIG_DIR", config_dir)
        .envs([NO_NONESSENTIAL_TRAFFIC]);
    cmd
}

/// Run `cmd` with no input, as a process group of its own, and give back its combined output and
/// whether it succeeded. A child still running at `deadline` is killed with everything it
/// started, and the answer is an error that holds the output so far.
fn run_with_deadline(mut cmd: Command, deadline: Duration) -> Result<(String, bool), String> {
    let dir = TempDir::new().unwrap();
    let log = dir.path().join("out");
    let out = fs::File::create(&log).unwrap();
    let err = out.try_clone().unwrap();
    cmd.stdin(Stdio::null()).stdout(out).stderr(err);
    let mut child = spawn_group(&mut cmd);
    let started = Instant::now();
    let status = loop {
        if let Some(status) = child.try_wait().unwrap() {
            break status;
        }
        if started.elapsed() >= deadline {
            kill_group(&mut child);
            let text = String::from_utf8_lossy(&fs::read(&log).unwrap()).into_owned();
            return Err(format!(
                "no exit within {deadline:?}; output so far: {text}"
            ));
        }
        thread::sleep(Duration::from_millis(50));
    };
    let text = String::from_utf8_lossy(&fs::read(&log).unwrap()).into_owned();
    Ok((text, status.success()))
}

/// Why `text` is not the answer of a known agent name in a folder without a credential, or
/// `None` when it is: the name is accepted (no "not found") and the run stops at the login step.
fn login_stop_error(text: &str) -> Option<String> {
    if text.contains("not found") {
        return Some(format!("the name was not found: {text}"));
    }
    if !text.contains(LOGIN_STOP) {
        return Some(format!(
            "the run did not stop at `{LOGIN_STOP}`, so a credential may have reached the child \
             and a model call was made: {text}"
        ));
    }
    None
}

/// Why a run of `program --agent <name> -p hi` is not the answer of a known agent name in a
/// folder without a credential, or `None` when it is: the run ends in time and stops at the login
/// step (see `login_stop_error`).
fn known_name_error(
    program: &Path,
    project: &Path,
    home: &Path,
    config_dir: &Path,
    name: &str,
) -> Option<String> {
    let cmd = claude_command(program, project, home, config_dir, name);
    match run_with_deadline(cmd, CHILD_DEADLINE) {
        Err(why) => Some(why),
        Ok((text, _)) => login_stop_error(&text),
    }
}

struct Fx {
    _tmp: TempDir,
    home: PathBuf,
    package: PathBuf,
    agents: PathBuf,
    project: PathBuf,
}

/// A temporary account home with the canonical `.claude/agents` holding one personal agent, the
/// package and agent sources copied from the repository, and a project with one nested role.
fn fx() -> Fx {
    let tmp = TempDir::new().unwrap();
    let base = fs::canonicalize(tmp.path()).unwrap();
    let home = base.join("home");
    let repo = Path::new(env!("CARGO_MANIFEST_DIR"));
    let package = base.join("package");
    copy_tree(&repo.join("skills/moai"), &package);
    let agents = base.join("agents");
    copy_tree(&repo.join("agents"), &agents);
    fs::create_dir_all(home.join(".claude/agents")).unwrap();
    fs::write(
        home.join(format!(".claude/agents/{PERSONAL}.md")),
        agent_file(PERSONAL),
    )
    .unwrap();
    let project = base.join("project");
    fs::create_dir_all(project.join(".claude/agents/moai")).unwrap();
    fs::write(
        project.join(format!(".claude/agents/moai/{PROJECT_ROLE}.md")),
        agent_file(PROJECT_ROLE),
    )
    .unwrap();
    Fx {
        _tmp: tmp,
        home,
        package,
        agents,
        project,
    }
}

impl Fx {
    fn canonical_agents(&self) -> PathBuf {
        self.home.join(".claude/agents")
    }

    /// The tool's own command on this home; `config_dir` is what an Orca session would set.
    fn moai(&self, args: &[&str], config_dir: &Path) -> Value {
        let out = Command::new(EXE)
            .args(args)
            .clean_env()
            .env("PATH", "/usr/bin:/bin")
            .env("LC_ALL", "C")
            .env("HOME", &self.home)
            .env("MOAI_ACCOUNT_HOME", &self.home)
            .env("CLAUDE_CONFIG_DIR", config_dir)
            .output()
            .unwrap();
        serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
            panic!(
                "{args:?}: {}{}",
                String::from_utf8_lossy(&out.stdout),
                String::from_utf8_lossy(&out.stderr)
            )
        })
    }

    fn plan(&self, config_dir: &Path) -> Value {
        let v = self.moai(
            &[
                "agents-home-plan",
                self.home.to_str().unwrap(),
                self.package.to_str().unwrap(),
                self.agents.to_str().unwrap(),
            ],
            config_dir,
        );
        assert_eq!(v["ok"], Value::Bool(true), "{v}");
        v["result"].clone()
    }

    fn install(&self, config_dir: &Path) {
        let plan = self.plan(config_dir);
        let v = self.moai(
            &[
                "agents-home-apply",
                self.home.to_str().unwrap(),
                self.package.to_str().unwrap(),
                self.agents.to_str().unwrap(),
                "--plan-digest",
                plan["plan_digest"].as_str().unwrap(),
                "--permission-reference",
                PERMISSION,
            ],
            config_dir,
        );
        assert_eq!(v["result"]["state"], "applied", "{v}");
    }

    /// `claude --agent <name> -p hi` in the project: its output and exit status, with the
    /// environment cleared (no credential reaches it) and `config_dir` as the config folder.
    fn claude(&self, claude: &Path, config_dir: &Path, name: &str) -> (String, bool) {
        let cmd = claude_command(claude, &self.project, &self.home, config_dir, name);
        run_with_deadline(cmd, CHILD_DEADLINE)
            .unwrap_or_else(|why| panic!("`claude --agent {name}`: {why}"))
    }

    /// The names of the available-agents list `claude` prints for an unknown `--agent` name.
    fn available(&self, claude: &Path, config_dir: &Path) -> Vec<String> {
        let (text, _) = self.claude(claude, config_dir, UNKNOWN);
        let list = text
            .split("Available agents:")
            .nth(1)
            .unwrap_or_else(|| panic!("no `Available agents:` line: {text}"));
        list.lines()
            .next()
            .unwrap()
            .split(',')
            .map(|n| n.trim().to_string())
            .filter(|n| !n.is_empty())
            .collect()
    }

    /// Each name of both kinds is in the list once, and `claude --agent <name>` passes the name
    /// check and stops at the login step (`Not logged in`, asserted): the temporary config folder
    /// holds no credential and none is ever copied, so any other answer is a leaked credential.
    fn assert_found_and_selectable(&self, claude: &Path, config_dir: &Path, label: &str) {
        let names = self.available(claude, config_dir);
        let wanted: Vec<&str> = BASE
            .iter()
            .copied()
            .chain([PROJECT_ROLE, PERSONAL])
            .collect();
        for name in &wanted {
            let count = names.iter().filter(|n| n == name).count();
            assert_eq!(
                count, 1,
                "{label}: `{name}` listed {count} times in {names:?}"
            );
        }
        // The detector below can fire: an unknown name is answered with "not found".
        let (unknown, ok) = self.claude(claude, config_dir, UNKNOWN);
        assert!(!ok && unknown.contains("not found"), "{label}: {unknown}");
        for name in &wanted {
            if let Some(why) = known_name_error(claude, &self.project, &self.home, config_dir, name)
            {
                panic!("{label}: `--agent {name}`: {why}");
            }
        }
    }
}

/// REQ-1, REQ-3, REQ-5: a native config folder.
#[test]
fn native_layout() {
    let Some(claude) = claude_or_skip() else {
        return;
    };
    let fx = fx();
    let config = fx.home.join(".claude");
    fx.install(&config);
    for name in BASE {
        assert!(
            fx.canonical_agents()
                .join(format!("moai/{name}.md"))
                .is_file(),
            "{name} is not under agents/moai"
        );
        assert!(
            !fx.canonical_agents().join(format!("{name}.md")).exists(),
            "{name} is also flat"
        );
    }
    fx.assert_found_and_selectable(&claude, &config, "native");
}

/// REQ-2, REQ-3, REQ-5: a profile folder whose `agents` entry links to the canonical agents
/// folder, the install run with the profile as `CLAUDE_CONFIG_DIR` (as an Orca session does).
#[test]
fn orca_profile_layout() {
    let Some(claude) = claude_or_skip() else {
        return;
    };
    let fx = fx();
    let profile = fx.home.parent().unwrap().join("profile");
    fs::create_dir_all(&profile).unwrap();
    symlink(fx.canonical_agents(), profile.join("agents")).unwrap();
    fx.install(&profile);
    // The install wrote through the link into the canonical folder and nothing else into the
    // profile: its `agents` entry is still a link and it holds no other entry.
    assert!(
        fs::symlink_metadata(profile.join("agents"))
            .unwrap()
            .file_type()
            .is_symlink(),
        "the profile's agents link was replaced"
    );
    let entries: Vec<String> = fs::read_dir(&profile)
        .unwrap()
        .map(|e| e.unwrap().file_name().to_string_lossy().into_owned())
        .collect();
    assert_eq!(entries, vec!["agents"], "the profile gained entries");
    for name in BASE {
        assert!(
            fx.canonical_agents()
                .join(format!("moai/{name}.md"))
                .is_file(),
            "{name} is not under the canonical agents/moai"
        );
        assert!(
            !fx.canonical_agents().join(format!("{name}.md")).exists(),
            "{name} is also flat in the canonical agents folder"
        );
    }
    // A reader through the profile finds both kinds and the personal agent.
    fx.assert_found_and_selectable(&claude, &profile, "orca profile");
    // The canonical folder read natively agrees with the profile read.
    let native = fx.available(&claude, &fx.home.join(".claude"));
    let through = fx.available(&claude, &profile);
    assert_eq!(native, through, "the two readers see different agents");
}

/// REQ-4: a profile whose `agents` link points at another folder is blocked, and a reader
/// through it finds no base agent — the control that shows the check can fail.
#[test]
fn a_profile_link_to_another_folder_is_blocked_and_finds_nothing() {
    let Some(claude) = claude_or_skip() else {
        return;
    };
    let fx = fx();
    let base = fx.home.parent().unwrap();
    let profile = base.join("profile");
    let elsewhere = base.join("elsewhere/agents");
    fs::create_dir_all(&profile).unwrap();
    fs::create_dir_all(&elsewhere).unwrap();
    symlink(&elsewhere, profile.join("agents")).unwrap();
    let plan = fx.plan(&profile);
    let claude_host = plan["hosts"]
        .as_array()
        .unwrap()
        .iter()
        .find(|h| h["host"] == "claude")
        .unwrap();
    assert_eq!(claude_host["preflight"], "blocked", "{claude_host}");
    assert!(
        claude_host["blocked_reason"]
            .as_str()
            .unwrap()
            .contains("is a link to a folder other than"),
        "{claude_host}"
    );
    let names = fx.available(&claude, &profile);
    for name in BASE {
        assert!(
            !names.iter().any(|n| n == name),
            "{name} is listed through a blocked profile: {names:?}"
        );
    }
    assert!(names.iter().any(|n| n == PROJECT_ROLE), "{names:?}");
    // The install itself refuses too: an apply with the blocked plan's digest writes no agent
    // file, neither through the link nor into the canonical folder.
    let applied = fx.moai(
        &[
            "agents-home-apply",
            fx.home.to_str().unwrap(),
            fx.package.to_str().unwrap(),
            fx.agents.to_str().unwrap(),
            "--plan-digest",
            plan["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            PERMISSION,
        ],
        &profile,
    );
    assert_eq!(applied["ok"], Value::Bool(true), "{applied}");
    assert_eq!(applied["result"]["state"], "unchanged", "{applied}");
    assert_eq!(
        applied["result"]["written"],
        serde_json::json!([]),
        "{applied}"
    );
    for dir in [elsewhere.join("moai"), fx.canonical_agents().join("moai")] {
        assert!(!dir.exists(), "a blocked apply created {}", dir.display());
    }
}

/// REQ-7: the child `claude` gets the traffic switch and nothing that could carry a credential.
/// The command built by `claude_command` is really started, with a fake program that prints the
/// environment it received, so a dropped `.clean_env()` shows (the test process's own variables
/// reach the child) where a look at the command's `get_envs()` would not. The keys the child
/// sees are exactly the allowed ones. Needs no `claude`, so it runs everywhere.
#[test]
fn the_child_claude_gets_the_traffic_switch_and_only_allowed_variables() {
    // The test process carries variables a credential could ride on; without this the check
    // below proves nothing.
    assert!(
        std::env::var_os("CARGO_MANIFEST_DIR").is_some() && std::env::var_os("PATH").is_some(),
        "this test must run under cargo or nextest"
    );
    let dir = TempDir::new().unwrap();
    let (home, config_dir, project) = (
        dir.path().join("home"),
        dir.path().join("config"),
        dir.path().join("project"),
    );
    fs::create_dir_all(&project).unwrap();
    let printer = dir.path().join("claude");
    common::write_executable(&printer, "#!/bin/sh\nexec /usr/bin/env\n");
    let cmd = claude_command(&printer, &project, &home, &config_dir, "moai-plan");
    let (text, ok) = run_with_deadline(cmd, Duration::from_secs(30)).unwrap();
    assert!(ok, "the fake program failed: {text}");
    let seen: BTreeMap<String, String> = text
        .lines()
        .filter_map(|l| l.split_once('='))
        .map(|(k, v)| (k.to_string(), v.to_string()))
        .filter(|(k, _)| !SHELL_ADDED.contains(&k.as_str()))
        .collect();
    let (key, value) = NO_NONESSENTIAL_TRAFFIC;
    let mut expected: BTreeSet<&str> = BTreeSet::from([
        "PATH",
        "HOME",
        "CLAUDE_CONFIG_DIR",
        key,
        common::PROCESS_TABLE_SWITCH,
    ]);
    // `clean_env` passes DEVELOPER_DIR on only when it names a folder that exists.
    if std::env::var_os("DEVELOPER_DIR").is_some_and(|d| Path::new(&d).is_dir()) {
        expected.insert("DEVELOPER_DIR");
    }
    let keys: BTreeSet<&str> = seen.keys().map(String::as_str).collect();
    assert_eq!(
        keys, expected,
        "the child `claude` saw other variables than the allowed ones (names only, a value \
         could be a secret)"
    );
    assert_eq!(seen[key], value, "the traffic switch has the wrong value");
    assert_eq!(seen["HOME"], home.to_string_lossy());
    assert_eq!(seen["CLAUDE_CONFIG_DIR"], config_dir.to_string_lossy());
}

/// REQ-7: a known name is refused when the answer is not the login stop, end to end through the
/// command a real run uses. A fake `claude` placed first on the PATH answers like a model (a
/// credential reached the child), like a missing name, or like the login stop; needs no
/// credential and no real `claude`, so it checks the guard in CI too.
#[test]
fn a_fake_claude_first_on_the_path_is_judged_by_the_known_name_check() {
    let dir = TempDir::new().unwrap();
    let (home, config_dir, project) = (
        dir.path().join("home"),
        dir.path().join("config"),
        dir.path().join("project"),
    );
    fs::create_dir_all(&project).unwrap();
    let cases = [
        // (what the fake prints, exit status, the error text expected or None for accepted)
        (
            "Hello! How can I help you today?",
            0,
            Some("model call was made"),
        ),
        (
            "--agent 'moai-plan' not found. Available agents: claude",
            1,
            Some("the name was not found"),
        ),
        ("Not logged in · Please run /login", 1, None),
    ];
    for (n, (answer, status, expected)) in cases.into_iter().enumerate() {
        let bin = dir.path().join(format!("bin{n}"));
        fs::create_dir_all(&bin).unwrap();
        common::write_executable(
            &bin.join("claude"),
            &format!("#!/bin/sh\necho '{answer}'\nexit {status}\n"),
        );
        let mut path = OsString::from(&bin);
        path.push(":");
        path.push(std::env::var_os("PATH").unwrap_or_default());
        let found = find_claude(Some(path)).expect("a claude on the PATH");
        assert_eq!(
            found,
            bin.join("claude"),
            "the fake is not first on the PATH"
        );
        let why = known_name_error(&found, &project, &home, &config_dir, "moai-plan");
        match (expected, why) {
            (None, None) => {}
            (Some(part), Some(why)) => {
                assert!(why.contains(part), "case {n}: wrong reason: {why}");
            }
            (expected, why) => panic!("case {n} ({answer}): expected {expected:?}, got {why:?}"),
        }
    }
}

/// REQ-7: only a run that stops at the login step is accepted for a known name; a model answer
/// (a credential reached the child) and a "not found" answer are both refused.
#[test]
fn a_known_name_is_accepted_only_at_the_login_stop() {
    let login = "Not logged in · Please run /login";
    assert_eq!(login_stop_error(login), None);
    let answered = login_stop_error("Hello! How can I help you today?");
    assert!(
        answered.is_some_and(|m| m.contains("model call")),
        "a model answer passed"
    );
    let missing = login_stop_error("--agent 'x' not found. Available agents: claude");
    assert!(
        missing.is_some_and(|m| m.contains("not found")),
        "not found passed"
    );
    assert!(login_stop_error("").is_some(), "an empty answer passed");
}

/// REQ-7: a child that never exits is killed at the deadline with the output it printed, and a
/// child that exits is read back whole with its status.
#[test]
fn a_child_is_killed_at_the_deadline_and_a_finished_one_is_read_whole() {
    let dir = TempDir::new().unwrap();
    let hang = dir.path().join("hang");
    common::write_executable(&hang, "#!/bin/sh\necho started\nsleep 30\n");
    let started = Instant::now();
    let err = run_with_deadline(Command::new(&hang), Duration::from_secs(1)).unwrap_err();
    assert!(
        err.contains("no exit within") && err.contains("started"),
        "{err}"
    );
    assert!(
        started.elapsed() < Duration::from_secs(15),
        "the hung child was not killed at the deadline: {:?}",
        started.elapsed()
    );
    let done = dir.path().join("done");
    common::write_executable(
        &done,
        "#!/bin/sh\necho Not logged in\necho oops >&2\nexit 1\n",
    );
    let (text, ok) = run_with_deadline(Command::new(&done), Duration::from_secs(30)).unwrap();
    assert!(
        !ok && text.contains("Not logged in") && text.contains("oops"),
        "{text}"
    );
}

/// REQ-6: no `claude` on the path is a skip, not a failure.
#[test]
fn no_claude_on_the_path_is_a_skip() {
    assert!(find_claude(None).is_none());
    assert!(find_claude(Some(OsString::new())).is_none());
    let empty = TempDir::new().unwrap();
    assert!(find_claude(Some(empty.path().into())).is_none());
}
