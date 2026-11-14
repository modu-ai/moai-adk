//! The host steps of `project-update-plan` (#2936 R9 S5): the Codex and OMP hook sets, the Codex
//! config line, the base agents in the project's second agent folder and the `.omp/config.yml`
//! line. Each step calls the plan the host's own writer takes and prints that writer's command;
//! nothing is written here.
//!
//! `project_update` turns a [`HostStep`] into its own step type; this module knows nothing
//! about phases. HOME is read here, so `project_update::plan` keeps its signature.

use crate::{hooks, hosts::codex_config, omp_agents, update};
use serde_json::{json, Value};
use std::path::{Path, PathBuf};

/// The one placeholder a printed command carries (the same word `project_update` prints).
const PLACEHOLDER: &str = crate::project_update::PERMISSION_PLACEHOLDER;

/// The step names this module makes, in plan order.
pub const NAMES: [&str; 5] = [
    "codex-hooks",
    "codex-config",
    "omp-hooks",
    "omp-config",
    "omp-agents",
];

/// What `codex-config-apply` takes on stdin: the one key MoAI manages (the `moai` status line
/// preset, #2936 S2).
fn codex_config_request() -> Value {
    json!({"tui.status_line": "moai"})
}

const OMP_CONFIG_REASON: &str = "MoAI writes no .omp/config.yml: no status line key is set \
     (#2936 S2, default A) and an existing file is never read or changed";

/// One host step; `apply` is the printed command as a word list and the JSON it takes on stdin
/// (`Null` when it takes none).
pub struct HostStep {
    pub name: &'static str,
    pub state: &'static str,
    pub reason: String,
    pub preview: Value,
    pub plan_digest: Option<String>,
    pub apply: Option<(Vec<String>, Value)>,
}

impl HostStep {
    fn bare(name: &'static str, state: &'static str, reason: &str) -> HostStep {
        HostStep {
            name,
            state,
            reason: reason.into(),
            preview: Value::Null,
            plan_digest: None,
            apply: None,
        }
    }
}

/// The host steps for the project `root`; `shown` is the project word the printed commands
/// carry. One step's error makes that step `blocked` (or `conflict` when the host file was
/// edited by hand) and the others still run.
pub fn steps(root: &Path, shown: &str) -> Vec<HostStep> {
    let home = std::env::var_os("HOME").map(PathBuf::from);
    let installed = update::installed_package_dir().and_then(|package| {
        update::installed_agents_dir().map(|agents| Installed { package, agents })
    });
    steps_in(
        root,
        shown,
        home.as_deref(),
        codex_config::config_home(root),
        installed,
    )
}

/// The skill package and the agent sources that came with this binary.
struct Installed {
    package: PathBuf,
    agents: PathBuf,
}

/// [`steps`] with the user's HOME, the Codex config home and the installed folders given: a
/// project that is the HOME, or lies in a host's own settings folder under it, is refused by
/// `hooks::plan` and shows as `blocked`.
fn steps_in(
    root: &Path,
    shown: &str,
    home: Option<&Path>,
    config_home: Result<PathBuf, String>,
    installed: Result<Installed, String>,
) -> Vec<HostStep> {
    vec![
        hooks_step("codex-hooks", "codex", root, shown, home),
        codex_config_step(root, shown, config_home),
        hooks_step("omp-hooks", "omp", root, shown, home),
        HostStep::bare("omp-config", "skip", OMP_CONFIG_REASON),
        omp_agents_step(root, shown, installed),
    ]
}

fn hooks_step(
    name: &'static str,
    host: &str,
    root: &Path,
    shown: &str,
    home: Option<&Path>,
) -> HostStep {
    match hooks::plan(root, host, home) {
        Ok(plan) => from_plan(name, host, shown, &plan),
        // A host file the user edited is kept as it is (the writer never replaces it).
        Err(e) if e.contains("preservation review") => HostStep::bare(name, "conflict", &e),
        Err(e) => HostStep::bare(name, "blocked", &e),
    }
}

fn from_plan(name: &'static str, host: &str, shown: &str, plan: &Value) -> HostStep {
    let digest = plan["plan_digest"].as_str().map(String::from);
    let (state, reason) = match plan["action"].as_str() {
        Some("create" | "update") => ("ready", "MoAI hook set to write"),
        Some("unchanged") => ("unchanged", "the MoAI hook set is current"),
        _ => ("blocked", "the hook plan has no known action"),
    };
    let apply = match (&digest, state) {
        (Some(d), "ready") => Some((
            [
                "moai",
                "hooks-apply",
                shown,
                "--host",
                host,
                "--plan-digest",
                d,
                "--permission-reference",
                PLACEHOLDER,
            ]
            .map(String::from)
            .to_vec(),
            Value::Null,
        )),
        _ => None,
    };
    HostStep {
        name,
        state,
        reason: reason.into(),
        preview: json!({
            "host": host,
            "action": plan["action"],
            "path": plan["path"],
            "hooks": plan["hooks"],
            "trust_required": hooks::trust_notice(host),
        }),
        plan_digest: digest,
        apply,
    }
}

/// The Codex config line: previews `tui.status_line = "moai"` in `PROJECT/.codex/config.toml`.
/// The preview names the file and the status line only; Codex trust is the user's and is
/// neither previewed nor changed here.
fn codex_config_step(root: &Path, shown: &str, config_home: Result<PathBuf, String>) -> HostStep {
    const NAME: &str = "codex-config";
    let desired = codex_config_request();
    let home = match config_home {
        Ok(h) => h,
        Err(e) => return HostStep::bare(NAME, "blocked", &e),
    };
    let Some(request) = desired.as_object() else {
        return HostStep::bare(NAME, "blocked", "the config request is not an object");
    };
    match codex_config::config_plan(root, &home, request) {
        Ok(plan) => config_from_plan(shown, &desired, &plan),
        Err(e) => HostStep::bare(NAME, "blocked", &e),
    }
}

fn config_from_plan(shown: &str, desired: &Value, plan: &Value) -> HostStep {
    let digest = plan["plan_digest"].as_str().map(String::from);
    let (state, reason) = match plan["action"].as_str() {
        Some("create" | "update") => ("ready", "MoAI status line to write"),
        Some("unchanged") => ("unchanged", "the MoAI status line is current"),
        Some("conflict") => (
            "conflict",
            "a key MoAI manages was changed outside MoAI; nothing is written",
        ),
        _ => ("blocked", "the config plan has no known action"),
    };
    let apply = match (&digest, state) {
        (Some(d), "ready") => Some((
            [
                "moai",
                "codex-config-apply",
                shown,
                "--plan-digest",
                d,
                "--permission-reference",
                PLACEHOLDER,
            ]
            .map(String::from)
            .to_vec(),
            desired.clone(),
        )),
        _ => None,
    };
    HostStep {
        name: "codex-config",
        state,
        reason: reason.into(),
        preview: json!({
            "host": "codex",
            "action": plan["action"],
            "config_path": plan["config_path"],
            "conflicts": plan["conflicts"],
            "status_line": plan["status_line"],
            "status_line_unverified": plan["status_line_unverified"],
        }),
        plan_digest: digest,
        apply,
    }
}

/// The base MoAI agents in `PROJECT/.omp/agents`, from the skill package and agent sources that
/// came with this binary. Without them there is nothing to copy, so the step is skipped.
fn omp_agents_step(root: &Path, shown: &str, installed: Result<Installed, String>) -> HostStep {
    const NAME: &str = "omp-agents";
    let installed = match installed {
        Ok(i) => i,
        Err(e) => return HostStep::bare(NAME, "skip", &e),
    };
    let (Some(package), Some(agents)) = (installed.package.to_str(), installed.agents.to_str())
    else {
        return HostStep::bare(NAME, "blocked", "the install folder path is not UTF-8");
    };
    match omp_agents::plan(root, &installed.package, &installed.agents) {
        Ok(plan) => agents_from_plan(shown, package, agents, &plan),
        Err(e) => HostStep::bare(NAME, "blocked", &e),
    }
}

fn agents_from_plan(shown: &str, package: &str, agents: &str, plan: &Value) -> HostStep {
    let files = plan["files"].as_array().cloned().unwrap_or_default();
    let count = |action: &str| files.iter().filter(|f| f["action"] == action).count();
    let digest = plan["plan_digest"].as_str().map(String::from);
    let (state, reason) = if plan["preflight"] != "ok" {
        (
            "blocked",
            plan["blocked_reason"]
                .as_str()
                .unwrap_or("the agent folder is blocked")
                .to_string(),
        )
    } else if count("create") + count("update") > 0 {
        ("ready", "MoAI base agents to write".to_string())
    } else if count("conflict") > 0 {
        (
            "conflict",
            "an agent file was changed outside MoAI; it is kept as it is".to_string(),
        )
    } else {
        ("unchanged", "the MoAI base agents are current".to_string())
    };
    let apply = match (&digest, state) {
        (Some(d), "ready") => Some((
            [
                "moai",
                "agents-project-apply",
                shown,
                package,
                agents,
                "--plan-digest",
                d,
                "--permission-reference",
                PLACEHOLDER,
            ]
            .map(String::from)
            .to_vec(),
            Value::Null,
        )),
        _ => None,
    };
    HostStep {
        name: "omp-agents",
        state,
        reason,
        preview: json!({
            "host": plan["host"],
            "dir": plan["dir"],
            "files": files
                .iter()
                .map(|f| json!({"file": f["file"], "action": f["action"]}))
                .collect::<Vec<_>>(),
            "shadows_user_agents": plan["shadows_user_agents"],
        }),
        plan_digest: digest,
        apply,
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn plan(action: &str) -> Value {
        json!({"action": action, "plan_digest": "sha256:fixture", "path": ".codex/hooks.json",
               "hooks": []})
    }

    /// agent-25's audit of #2993 P19: a hook file MoAI wrote earlier is brought up to date.
    #[test]
    fn an_update_is_ready_and_prints_its_command() {
        let step = from_plan("codex-hooks", "codex", "/p", &plan("update"));
        assert_eq!(step.state, "ready");
        let (argv, stdin) = step.apply.expect("a ready step prints its command");
        assert_eq!(&argv[..2], ["moai", "hooks-apply"]);
        assert_eq!(stdin, Value::Null);
        assert_eq!(argv.last().map(String::as_str), Some(PLACEHOLDER));
    }

    /// P18: an action this module does not know is never offered for apply.
    #[test]
    fn an_unknown_action_is_blocked_without_a_command() {
        let step = from_plan("codex-hooks", "codex", "/p", &plan("something-new"));
        assert_eq!(step.state, "blocked");
        assert!(step.apply.is_none());
    }

    /// P12: a project that is the user's HOME is refused, not previewed.
    #[test]
    fn a_project_that_is_the_home_is_blocked() {
        let t = tempfile::tempdir().unwrap();
        let root = std::fs::canonicalize(t.path()).unwrap();
        let shown = root.to_str().unwrap();
        for s in steps_in(
            &root,
            shown,
            Some(&root),
            Err("none".into()),
            Err("none".into()),
        )
        .iter()
        .filter(|s| s.name.ends_with("-hooks"))
        {
            assert_eq!(s.state, "blocked", "{}: {}", s.name, s.reason);
            assert!(s.apply.is_none(), "{}", s.name);
        }
        // The same folder as a project of a different HOME is previewed.
        let other = tempfile::tempdir().unwrap();
        let home = std::fs::canonicalize(other.path()).unwrap();
        for s in steps_in(
            &root,
            shown,
            Some(&home),
            Err("none".into()),
            Err("none".into()),
        )
        .iter()
        .filter(|s| s.name.ends_with("-hooks"))
        {
            assert_eq!(s.state, "ready", "{}: {}", s.name, s.reason);
        }
    }
    fn agents_plan(files: &[(&str, &str)], preflight: &str) -> Value {
        json!({"host": "omp", "dir": ".omp/agents", "preflight": preflight,
               "blocked_reason": if preflight == "ok" { Value::Null } else { json!("linked") },
               "plan_digest": "sha256:fixture", "shadows_user_agents": "shadows",
               "files": files.iter().map(|(f, a)| json!({"file": f, "action": a})).collect::<Vec<_>>()})
    }

    fn agents_state(files: &[(&str, &str)], preflight: &str) -> HostStep {
        agents_from_plan(
            "/p",
            "/i/skills/moai",
            "/i/agents",
            &agents_plan(files, preflight),
        )
    }

    /// What each file action means for the step: a create or update is ready (with the other
    /// files reported), an edit that is kept leaves nothing to do, a conflict alone is a conflict.
    #[test]
    fn the_agent_file_actions_set_the_step_state() {
        for (files, state) in [
            (vec![("a.md", "create")], "ready"),
            (vec![("a.md", "update"), ("b.md", "conflict")], "ready"),
            (
                vec![("a.md", "unchanged"), ("b.md", "preserved")],
                "unchanged",
            ),
            (
                vec![("a.md", "unchanged"), ("b.md", "remove_blocked")],
                "unchanged",
            ),
            (vec![("a.md", "conflict")], "conflict"),
            (vec![], "unchanged"),
        ] {
            let s = agents_state(&files, "ok");
            assert_eq!(s.state, state, "{files:?}");
            assert_eq!(s.apply.is_some(), state == "ready", "{files:?}");
        }
        let blocked = agents_state(&[("a.md", "create")], "blocked");
        assert_eq!(
            (blocked.state, blocked.reason.as_str()),
            ("blocked", "linked")
        );
        assert!(blocked.apply.is_none());
    }

    #[test]
    fn the_agents_command_names_the_installed_folders_and_takes_no_stdin() {
        let (argv, stdin) = agents_state(&[("a.md", "create")], "ok").apply.unwrap();
        assert_eq!(
            argv[..5],
            [
                "moai",
                "agents-project-apply",
                "/p",
                "/i/skills/moai",
                "/i/agents"
            ]
        );
        assert_eq!(argv.last().map(String::as_str), Some(PLACEHOLDER));
        assert_eq!(stdin, Value::Null);
    }

    /// No install folders: the agents step is skipped, never blocked and never offered.
    #[test]
    fn without_an_install_the_agents_step_is_skipped() {
        let s = omp_agents_step(
            Path::new("/p"),
            "/p",
            Err("no skill package installed".into()),
        );
        assert_eq!((s.state, s.apply.is_none()), ("skip", true));
        assert!(s.reason.contains("installed"));
    }

    /// A Codex home that cannot be told (or a project `.codex` that is the user's config folder)
    /// blocks the config step; it never previews a write.
    #[test]
    fn a_codex_home_that_cannot_be_told_blocks_the_config_step() {
        let s = codex_config_step(
            Path::new("/p"),
            "/p",
            Err("home directory missing; config target refused".into()),
        );
        assert_eq!(s.state, "blocked");
        assert!(s.apply.is_none() && s.plan_digest.is_none());
    }
    fn config_plan(action: &str) -> Value {
        json!({"action": action, "plan_digest": "sha256:fixture", "config_path": ".codex/config.toml",
               "conflicts": [], "status_line": {}, "status_line_unverified": {}})
    }

    /// The config actions: write only for create and update, the stdin carries the one key.
    #[test]
    fn the_config_actions_set_the_step_state() {
        for (action, state) in [
            ("create", "ready"),
            ("update", "ready"),
            ("unchanged", "unchanged"),
            ("conflict", "conflict"),
            ("something-new", "blocked"),
        ] {
            let s = config_from_plan("/p", &codex_config_request(), &config_plan(action));
            assert_eq!(s.state, state, "{action}");
            match s.apply {
                Some((argv, stdin)) => {
                    assert_eq!(state, "ready", "{action}");
                    assert_eq!(&argv[..2], ["moai", "codex-config-apply"]);
                    assert_eq!(stdin, json!({"tui.status_line": "moai"}));
                    assert_eq!(argv.last().map(String::as_str), Some(PLACEHOLDER));
                }
                None => assert_ne!(state, "ready", "{action}"),
            }
        }
    }
}
