//! SPEC-20261125-agent-team-session-names (user decision 2026-11-25, reversing D-1 of
//! SPEC-20261124-launcher-retire-role-flags): `moai cc|glm -l|--leader` and `-a|--agent` start a
//! named session in the permission bypass and hand the host the first input `/moai leader` or
//! `/moai agent`. Every case runs in a temporary git project with a temporary HOME; no real host
//! is started (a stub `claude` records its arguments). AC-5 (advisor and model defaults of `-a`)
//! waits for the approval of revision 3 of the SPEC and is not in this file yet.
#![cfg(unix)]

mod common;
use common::CleanEnv;
use serde_json::Value;
use std::fs;
use std::path::PathBuf;
use std::process::{Command, Output};

const TOKEN: &str = "dummy-token-for-tests";

/// A temporary git project folder named `name`, and the temporary HOME the launcher sees.
struct Fixture {
    _root: tempfile::TempDir,
    project: PathBuf,
    home: tempfile::TempDir,
    stub: tempfile::TempDir,
    args_file: PathBuf,
}

impl Fixture {
    fn new(name: &str) -> Self {
        let root = tempfile::tempdir().unwrap();
        let project = root.path().join(name);
        fs::create_dir_all(&project).unwrap();
        assert!(common::git::command()
            .args(["init", "-q"])
            .current_dir(&project)
            .status()
            .unwrap()
            .success());
        let stub = tempfile::tempdir().unwrap();
        let args_file = stub.path().join("host-args");
        let script = format!(
            "#!/bin/sh\nfor a in \"$@\"; do printf '%s\\n' \"$a\"; done > '{}'\n",
            args_file.display()
        );
        let path = stub.path().join("claude");
        fs::write(&path, script).unwrap();
        {
            use std::os::unix::fs::PermissionsExt;
            fs::set_permissions(&path, fs::Permissions::from_mode(0o755)).unwrap();
        }
        Fixture {
            _root: root,
            project,
            home: tempfile::tempdir().unwrap(),
            stub,
            args_file,
        }
    }

    fn command(&self, args: &[&str]) -> Command {
        let mut cmd = Command::new(common::MOAI_BIN);
        cmd.args(args)
            .current_dir(&self.project)
            .clean_env()
            .env("HOME", self.home.path())
            .env("ANTHROPIC_AUTH_TOKEN", TOKEN)
            .env(
                "PATH",
                format!("{}:/usr/bin:/bin", self.stub.path().display()),
            );
        cmd
    }

    /// `--dry-run`: the composed command as JSON.
    fn dry(&self, args: &[&str]) -> (i32, Value, String) {
        let mut all = args.to_vec();
        all.push("--dry-run");
        let out = self.command(&all).output().unwrap();
        let v = serde_json::from_slice(&out.stdout).unwrap_or(Value::Null);
        (out.status.code().unwrap_or(-1), v, text(&out))
    }

    /// A real launch against the stub host; the arguments the stub received, if it ran.
    fn run(&self, args: &[&str]) -> (Output, Option<Vec<String>>) {
        let out = self.command(args).output().unwrap();
        let seen = fs::read_to_string(&self.args_file)
            .ok()
            .map(|s| s.lines().map(str::to_string).collect());
        (out, seen)
    }

    fn sessions(&self) -> PathBuf {
        moai_cli::session::agent_seq_dir(&self.project).unwrap()
    }

    /// Holds the live lock of `name` for as long as the returned file lives.
    fn hold(&self, name: &str) -> fs::File {
        moai_cli::session::lock_named_live(&self.sessions(), name).unwrap()
    }
}

fn text(out: &Output) -> String {
    format!(
        "{}{}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    )
}

fn argv(v: &Value) -> Vec<String> {
    serde_json::from_value(v["argv"].clone()).unwrap()
}

fn value_after(argv: &[String], flag: &str) -> Option<String> {
    argv.iter()
        .position(|a| a == flag)
        .map(|i| argv[i + 1].clone())
}

/// AC-1 (SPEC-20261125-session-names-project-prefix REQ-1, REQ-5): `-l` and `--leader` name the
/// session `<project name>-leader`, start in the bypass and hand `/moai leader` as the last
/// argument; `--name` replaces the name; `glm` takes the same flags with its own effort.
#[test]
fn ac1_leader_names_the_session_bypasses_and_hands_the_first_input() {
    let f = Fixture::new("myproj");
    for flag in ["-l", "--leader"] {
        let (code, v, said) = f.dry(&["cc", flag]);
        assert_eq!(code, 0, "{flag}: {said}");
        assert_eq!(
            argv(&v),
            [
                "claude",
                "--name",
                "myproj-leader",
                "--model",
                "opus",
                "--effort",
                "medium",
                "--dangerously-skip-permissions",
                "/moai leader"
            ],
            "{flag}"
        );
        assert_eq!(v["name"], "myproj-leader", "{flag}");
    }
    let (code, v, said) = f.dry(&["glm", "-l"]);
    assert_eq!(code, 0, "{said}");
    let a = argv(&v);
    assert_eq!(value_after(&a, "--name").as_deref(), Some("myproj-leader"));
    assert_eq!(value_after(&a, "--effort").as_deref(), Some("high"));
    assert!(
        a.contains(&"--dangerously-skip-permissions".to_string()),
        "{a:?}"
    );
    assert_eq!(a.last().map(String::as_str), Some("/moai leader"));
    // `--name` is used as given, after the registry's name check.
    let (code, v, said) = f.dry(&["cc", "-l", "--name", "boss-1"]);
    assert_eq!(code, 0, "{said}");
    assert_eq!(value_after(&argv(&v), "--name").as_deref(), Some("boss-1"));
    let (code, _, said) = f.dry(&["cc", "-l", "--name", "bad name"]);
    assert_ne!(code, 0, "a name the check refuses stays refused: {said}");
}

/// AC-1 (REQ-4, decisions D-2 and D-3): the project name is the `project_id` of the profile when
/// it is a name a person chose, the folder name of the primary checkout when the id is of the
/// random 26-character form (never the folder of a worktree), reduced to the registry's
/// characters, and short enough that the longest session name fits the registry's 64 characters.
#[test]
fn ac1_the_project_name_comes_from_the_profile_else_the_folder() {
    let f = Fixture::new("some folder.name");
    let (code, v, said) = f.dry(&["cc", "-l"]);
    assert_eq!(code, 0, "{said}");
    assert_eq!(v["name"], "some-folder.name-leader", "{said}");
    let state = f.project.join(moai_cli::PROJECT_STATE_DIR);
    fs::create_dir_all(&state).unwrap();
    fs::write(
        state.join("project.yaml"),
        "{\"project_id\": \"demo-id\"}\n",
    )
    .unwrap();
    let (code, v, said) = f.dry(&["cc", "-l"]);
    assert_eq!(code, 0, "{said}");
    assert_eq!(v["name"], "demo-id-leader", "{said}");
    let (code, v, said) = f.dry(&["cc", "-a"]);
    assert_eq!(code, 0, "{said}");
    assert_eq!(
        v["name"],
        format!("demo-id-agent-{}", moai_cli::session::AGENT_NAMES[0]),
        "{said}"
    );
    // A random-form id names no project to a person: the folder name is used.
    fs::write(
        state.join("project.yaml"),
        "{\"project_id\": \"abcdefghijklmnopqrstuvwxyz\"}\n",
    )
    .unwrap();
    assert!(moai_cli::project_id::is_random_form(
        "abcdefghijklmnopqrstuvwxyz"
    ));
    let (code, v, said) = f.dry(&["cc", "-l"]);
    assert_eq!(code, 0, "{said}");
    assert_eq!(v["name"], "some-folder.name-leader", "{said}");
    // The longest session name stays within the registry's 64 characters.
    let long = Fixture::new(&"p".repeat(80));
    for (flag, tail) in [("-l", "-leader"), ("-a", "-agent-")] {
        let (code, v, said) = long.dry(&["cc", flag]);
        assert_eq!(code, 0, "{said}");
        let name = v["name"].as_str().unwrap().to_string();
        assert!(name.starts_with('p') && name.contains(tail), "{name}");
        assert!(moai_cli::lane_registry::valid_name(&name), "{name}");
    }
    let longest = moai_cli::session::AGENT_NAMES
        .iter()
        .map(|n| n.len())
        .max()
        .unwrap();
    let project = moai_cli::session::project_name(&long.project);
    assert!(
        project.len() + "-agent-".len() + longest <= moai_cli::lane_registry::MAX_NAME_LEN,
        "{project}"
    );
}

/// AC-1 (REQ-4, decision D-2): a worktree of the project gets the project's name, not the
/// folder name of the worktree.
#[test]
fn ac1_a_worktree_uses_the_primary_checkouts_name() {
    let f = Fixture::new("wtproj");
    let git = |args: &[&str]| {
        assert!(common::git::command()
            .args(args)
            .current_dir(&f.project)
            .status()
            .unwrap()
            .success());
    };
    git(&["config", "user.email", "t@example.invalid"]);
    git(&["config", "user.name", "t"]);
    git(&["commit", "-q", "--allow-empty", "-m", "init"]);
    let wt = f._root.path().join("card-worktree-name");
    git(&["worktree", "add", "-q", wt.to_str().unwrap()]);
    let out = f
        .command(&["cc", "-l", "--dry-run"])
        .current_dir(&wt)
        .output()
        .unwrap();
    let v: Value = serde_json::from_slice(&out.stdout).unwrap_or(Value::Null);
    assert_eq!(v["name"], "wtproj-leader", "{}", text(&out));
}

/// AC-2 (REQ-15): a held live lock of `<project name>-leader` refuses the launch and starts no
/// host; once it is free the host starts and receives the first input.
#[test]
fn ac2_a_held_leader_lock_refuses_the_launch() {
    let f = Fixture::new("lockproj");
    let held = f.hold("lockproj-leader");
    let (out, seen) = f.run(&["cc", "-l"]);
    assert!(!out.status.success(), "{}", text(&out));
    assert!(text(&out).contains("lockproj-leader"), "{}", text(&out));
    assert!(seen.is_none(), "the host started: {seen:?}");
    drop(held);
    let (out, seen) = f.run(&["cc", "-l"]);
    assert!(out.status.success(), "{}", text(&out));
    let seen = seen.expect("the host ran");
    assert_eq!(
        seen.last().map(String::as_str),
        Some("/moai leader"),
        "{seen:?}"
    );
    assert!(seen.contains(&"lockproj-leader".to_string()), "{seen:?}");
}

/// AC-2 (REQ-2, REQ-3, REQ-5): `-a` without `--name` takes the first `<project name>-agent-<name>`
/// of the list that has no held live lock and no live registry entry, and `--name` replaces it.
#[test]
fn ac2_agent_takes_the_first_free_english_name() {
    let names = moai_cli::session::AGENT_NAMES;
    assert!(names.len() >= 12, "the list holds at least 12 names");
    assert!(names
        .iter()
        .all(|n| moai_cli::lane_registry::valid_name(&format!("agentproj-agent-{n}"))));
    let unique: std::collections::BTreeSet<_> = names.iter().collect();
    assert_eq!(unique.len(), names.len(), "no name twice");
    let f = Fixture::new("agentproj");
    let (code, v, said) = f.dry(&["cc", "-a"]);
    assert_eq!(code, 0, "{said}");
    assert_eq!(v["name"], format!("agentproj-agent-{}", names[0]), "{said}");
    // A held live lock skips a name.
    let _first = f.hold(&format!("agentproj-agent-{}", names[0]));
    let (_, v, said) = f.dry(&["cc", "--agent"]);
    assert_eq!(v["name"], format!("agentproj-agent-{}", names[1]), "{said}");
    // So does a live registry entry.
    let registered = Command::new(common::MOAI_BIN)
        .args([
            "session-register",
            "--name",
            &format!("agentproj-agent-{}", names[1]),
        ])
        .args(["--mode", "agent"])
        .arg(&f.project)
        .clean_env()
        .env("HOME", f.home.path())
        .env("PATH", "/usr/bin:/bin")
        .output()
        .unwrap();
    assert!(registered.status.success(), "{}", text(&registered));
    let (_, v, said) = f.dry(&["cc", "-a"]);
    assert_eq!(v["name"], format!("agentproj-agent-{}", names[2]), "{said}");
    // `--name` is used as given.
    let (_, v, said) = f.dry(&["cc", "-a", "--name", "scout"]);
    assert_eq!(v["name"], "scout", "{said}");
}

/// AC-2 (REQ-4): a list that is used up refuses the launch with one message that names
/// `--name` and starts no host; `--name` still works then.
#[test]
fn ac2_a_used_up_list_refuses_and_names_the_name_option() {
    let f = Fixture::new("fullproj");
    let held: Vec<fs::File> = moai_cli::session::AGENT_NAMES
        .iter()
        .map(|n| f.hold(&format!("fullproj-agent-{n}")))
        .collect();
    let (out, seen) = f.run(&["cc", "-a"]);
    assert!(!out.status.success(), "{}", text(&out));
    assert!(text(&out).contains("--name"), "{}", text(&out));
    assert!(seen.is_none(), "the host started: {seen:?}");
    let (code, v, said) = f.dry(&["cc", "-a", "--name", "extra"]);
    assert_eq!(code, 0, "{said}");
    assert_eq!(v["name"], "extra");
    drop(held);
}

/// AC-3 (REQ-6, REQ-11): `-l` and `-a` start in the bypass on both hosts and `--permissions`
/// keeps the host's own prompts; a launch without a role flag is unchanged.
#[test]
fn ac3_role_flags_bypass_and_permissions_keeps_the_prompts() {
    let f = Fixture::new("accessproj");
    for host in ["cc", "glm"] {
        for flag in ["-l", "-a"] {
            let (code, v, said) = f.dry(&[host, flag]);
            assert_eq!(code, 0, "{host} {flag}: {said}");
            let a = argv(&v);
            assert!(
                a.contains(&"--dangerously-skip-permissions".to_string()),
                "{a:?}"
            );
            assert!(!a.contains(&"--permission-mode".to_string()), "{a:?}");
            let (code, v, said) = f.dry(&[host, flag, "--permissions"]);
            assert_eq!(code, 0, "{host} {flag} --permissions: {said}");
            let a = argv(&v);
            assert!(
                !a.contains(&"--dangerously-skip-permissions".to_string()),
                "{a:?}"
            );
            assert!(!a.contains(&"--permission-mode".to_string()), "{a:?}");
        }
    }
    // Without a role flag: no name, no first input; `cc` in auto mode, `glm` in the bypass.
    let (_, v, _) = f.dry(&["cc"]);
    assert_eq!(
        argv(&v),
        [
            "claude",
            "--model",
            "opus",
            "--effort",
            "medium",
            "--permission-mode",
            "auto"
        ]
    );
    let (_, v, _) = f.dry(&["glm"]);
    assert_eq!(
        argv(&v),
        [
            "claude",
            "--model",
            "opus",
            "--effort",
            "high",
            "--dangerously-skip-permissions"
        ]
    );
}

/// AC-3 (REQ-12): `-l` with `-a`, or a PROMPT with a role flag, is refused with one message and
/// starts no host.
#[test]
fn ac3_a_role_flag_pair_or_a_prompt_with_one_is_refused() {
    let f = Fixture::new("refuseproj");
    for args in [
        &["cc", "-l", "-a"][..],
        &["cc", "--leader", "--agent"][..],
        &["glm", "-a", "-l"][..],
        &["cc", "-l", "do the thing"][..],
        &["cc", "-a", "do the thing"][..],
    ] {
        let (out, seen) = f.run(args);
        assert_eq!(out.status.code(), Some(1), "{args:?}: {}", text(&out));
        assert!(seen.is_none(), "{args:?} started the host: {seen:?}");
        let said = text(&out);
        assert!(
            said.contains("-l") && said.contains("-a"),
            "{args:?}: {said}"
        );
        assert!(
            !said.contains("retired"),
            "{args:?} still says retired: {said}"
        );
    }
}

/// AC-4 (REQ-7, REQ-8): `-l` hands `/moai leader` and `-a` hands `/moai agent` as the host's
/// last argument, one argument each, in a real launch too.
#[test]
fn ac4_the_first_input_is_the_last_host_argument() {
    let f = Fixture::new("inputproj");
    for (flag, input) in [("-l", "/moai leader"), ("-a", "/moai agent")] {
        let (_, v, said) = f.dry(&["cc", flag]);
        let a = argv(&v);
        assert_eq!(a.last().map(String::as_str), Some(input), "{said}");
        assert_eq!(a.iter().filter(|x| x.as_str() == input).count(), 1, "{a:?}");
        let (out, seen) = f.run(&["cc", flag]);
        assert!(out.status.success(), "{flag}: {}", text(&out));
        let seen = seen.expect("the host ran");
        assert_eq!(seen.last().map(String::as_str), Some(input), "{seen:?}");
        let _ = fs::remove_file(&f.args_file);
    }
    // Without a role flag the host gets no first input.
    let (_, v, _) = f.dry(&["cc"]);
    assert!(!argv(&v).iter().any(|a| a.starts_with("/moai")));
}

/// AC-6 (REQ-13): no live text calls the role flags retired and the launcher no longer refuses
/// them. The help text names them with what each does.
#[test]
fn ac6_the_launcher_help_names_the_four_flags() {
    let home = tempfile::tempdir().unwrap();
    for args in [&["cc", "--help"][..], &["glm", "--help"][..]] {
        let out = Command::new(common::MOAI_BIN)
            .args(args)
            .clean_env()
            .env("HOME", home.path())
            .output()
            .unwrap();
        let said = text(&out);
        for want in ["-l, --leader", "-a, --agent"] {
            assert!(said.contains(want), "{args:?} lacks {want}: {said}");
        }
        assert!(!said.contains("retired"), "{args:?}: {said}");
    }
}

impl Fixture {
    /// Writes the user launcher settings file the launcher reads from the temporary HOME.
    fn user_launcher(&self, text: &str) {
        let dir = moai_cli::home_root::path(self.home.path()).join("config");
        fs::create_dir_all(&dir).unwrap();
        fs::write(dir.join("launcher.yaml"), text).unwrap();
    }
}

fn pair(argv: &[String], flag: &str, value: &str) -> bool {
    argv.windows(2).any(|w| w[0] == flag && w[1] == value)
}

/// AC-5 (REQ-9, REQ-10, D-6, D-13): `-a` on `cc` passes `--advisor opus` unless the option, the
/// setting or `--no-advisor` says otherwise; `-l`, a launch without a role flag and `glm` pass none.
#[test]
fn ac5_agent_advisor_is_opus_unless_told_otherwise() {
    let f = Fixture::new("advproj");
    let (_, v, said) = f.dry(&["cc", "-a"]);
    assert!(pair(&argv(&v), "--advisor", "opus"), "{said}");
    let (_, v, said) = f.dry(&["cc", "-a", "--no-advisor"]);
    assert!(!argv(&v).contains(&"--advisor".to_string()), "{said}");
    let (_, v, said) = f.dry(&["cc", "-a", "--advisor", "fable"]);
    assert!(pair(&argv(&v), "--advisor", "fable"), "{said}");
    for args in [
        &["cc", "-l"][..],
        &["cc"][..],
        &["glm", "-a"][..],
        &["glm", "-l"][..],
    ] {
        let (code, v, said) = f.dry(args);
        assert_eq!(code, 0, "{args:?}: {said}");
        assert!(
            !argv(&v).contains(&"--advisor".to_string()),
            "{args:?} passed an advisor: {said}"
        );
    }
    // D-13: the user setting beats the `-a` default, and `none` in it keeps no advisor.
    f.user_launcher(r#"{"launcher":{"advisor":"fable"}}"#);
    let (_, v, said) = f.dry(&["cc", "-a"]);
    assert!(pair(&argv(&v), "--advisor", "fable"), "{said}");
    f.user_launcher(r#"{"launcher":{"advisor":"none"}}"#);
    let (_, v, said) = f.dry(&["cc", "-a"]);
    assert!(pair(&argv(&v), "--advisor", "opus"), "{said}");
}

/// AC-5 (D-14): a main model that ranks above the default advisor leaves the default out with
/// one line of reason; an `--advisor` the person gave that ranks lower is refused as before.
#[test]
fn ac5_a_higher_main_model_drops_the_default_advisor_and_refuses_a_lower_given_one() {
    let f = Fixture::new("rankproj");
    let (code, v, said) = f.dry(&["cc", "-a", "--model", "fable"]);
    assert_eq!(code, 0, "{said}");
    assert!(!argv(&v).contains(&"--advisor".to_string()), "{said}");
    assert!(
        v["notes"]
            .to_string()
            .contains("ranks below the main model"),
        "{said}"
    );
    let (code, _, said) = f.dry(&["cc", "-a", "--model", "fable", "--advisor", "opus"]);
    assert_ne!(code, 0, "{said}");
    assert!(said.contains("advisor must rank at or above"), "{said}");
}

/// AC-5 (REQ-28, D-12): `-a` starts the host with the saved agent choice, else the agent row of
/// the default table of its host; `--model` and `--effort` win; `-l` and a launch without a role
/// flag keep their model and effort.
#[test]
fn ac5_agent_model_and_effort_come_from_the_agent_row() {
    let f = Fixture::new("modelproj");
    let (_, v, said) = f.dry(&["cc", "-a"]);
    let a = argv(&v);
    assert!(
        pair(&a, "--model", "claude-sonnet-5-5") && pair(&a, "--effort", "high"),
        "{said}"
    );
    let (_, v, said) = f.dry(&["glm", "-a"]);
    let a = argv(&v);
    assert!(
        pair(&a, "--model", "glm-5.3-flash") && pair(&a, "--effort", "max"),
        "{said}"
    );
    // The options win, each over its own half.
    let (_, v, said) = f.dry(&["cc", "-a", "--model", "opus"]);
    let a = argv(&v);
    assert!(
        pair(&a, "--model", "opus") && pair(&a, "--effort", "high"),
        "{said}"
    );
    let (_, v, said) = f.dry(&["cc", "-a", "--effort", "low"]);
    let a = argv(&v);
    assert!(
        pair(&a, "--model", "claude-sonnet-5-5") && pair(&a, "--effort", "low"),
        "{said}"
    );
    // A choice saved for the agent role is the one `-a` takes.
    f.user_launcher(
        r#"{"launcher":{"models":{"claude":{"agent":{"model":"claude-opus-5-5","effort":"medium"}}}}}"#,
    );
    let (_, v, said) = f.dry(&["cc", "-a"]);
    let a = argv(&v);
    assert!(
        pair(&a, "--model", "claude-opus-5-5") && pair(&a, "--effort", "medium"),
        "{said}"
    );
    // `-l` and a launch without a role flag keep today's model and effort.
    for args in [&["cc", "-l"][..], &["cc"][..]] {
        let (_, v, said) = f.dry(args);
        let a = argv(&v);
        assert!(
            pair(&a, "--model", "opus") && pair(&a, "--effort", "medium"),
            "{args:?}: {said}"
        );
    }
}

/// The list of names lives in code and is checked by the registry's rule (REQ-3).
#[test]
fn the_name_list_is_lowercase_single_words() {
    for n in moai_cli::session::AGENT_NAMES {
        assert!(
            !n.is_empty() && n.bytes().all(|b| b.is_ascii_lowercase()),
            "{n}"
        );
    }
}
