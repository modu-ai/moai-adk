//! SPEC-20261123-base-agents-discovery-record (REQ-1 .. REQ-6): the installed Claude Code finds
//! the five base agents and a project role in a native config folder and in a folder shaped
//! like an Orca profile (its `agents` entry links to the user's agents folder), and no name is
//! hidden. Every case runs the built binary and the real `claude` on temporary folders with a
//! cleared environment and no credential; the real HOME, `~/.claude` and the Orca profile are
//! never read or written. A case is skipped (and says so) when `claude` is not on the PATH.
#![cfg(unix)]

mod common;
use common::CleanEnv;
use serde_json::Value;
use std::{
    ffi::OsString,
    fs,
    os::unix::fs::symlink,
    path::{Path, PathBuf},
    process::Command,
};
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;
const PERMISSION: &str = "fixture-approved";
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
        let out = Command::new(claude)
            .clean_env()
            .args(["--agent", name, "-p", "hi"])
            .current_dir(&self.project)
            .env("PATH", std::env::var_os("PATH").unwrap())
            .env("HOME", &self.home)
            .env("CLAUDE_CONFIG_DIR", config_dir)
            .output()
            .unwrap();
        let text = format!(
            "{}{}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        );
        (text, out.status.success())
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
    /// check: it never answers that the agent is not found (it stops at the login step, because
    /// the temporary config folder holds no credential and none is ever copied).
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
            let (text, _) = self.claude(claude, config_dir, name);
            assert!(
                !text.contains("not found"),
                "{label}: `--agent {name}` was not found: {text}"
            );
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
}

/// REQ-6: no `claude` on the path is a skip, not a failure.
#[test]
fn no_claude_on_the_path_is_a_skip() {
    assert!(find_claude(None).is_none());
    assert!(find_claude(Some(OsString::new())).is_none());
    let empty = TempDir::new().unwrap();
    assert!(find_claude(Some(empty.path().into())).is_none());
}
