//! The host steps of `project-update-plan` (#2936 R9 S5): the Codex and OMP hook sets, the Codex
//! config line, the project copy of the `moai` skill for Codex, the base agents in the project's
//! second agent folder and the `.omp/config.yml` line. Each step calls the plan the host's own writer takes and prints that writer's command;
//! nothing is written here.
//!
//! `project_update` turns a [`HostStep`] into its own step type; this module knows nothing
//! about phases. HOME is read here, so `project_update::plan` keeps its signature.

use crate::{hooks, hosts::codex_config, omp_agents, project_skills, update};
use serde_json::{json, Value};
use std::path::{Path, PathBuf};

/// The one placeholder a printed command carries (the same word `project_update` prints).
const PLACEHOLDER: &str = crate::project_update::PERMISSION_PLACEHOLDER;

/// The step names this module makes, in plan order.
pub const NAMES: [&str; 6] = [
    "codex-hooks",
    "codex-config",
    "codex-skill",
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
pub fn steps(root: &Path, shown: &str, picked: Option<&[String]>) -> Vec<HostStep> {
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
        picked,
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
    picked: Option<&[String]>,
) -> Vec<HostStep> {
    // A host the profile does not list gets no plan at all (SPEC-20261118-profile-host-omp);
    // `None` (no readable profile) plans every host as before.
    let listed = |host: &str| picked.is_none_or(|p| p.iter().any(|h| h == host));
    let unlisted = |name: &'static str, host: &str| {
        HostStep::bare(name, "skip", &format!("the profile does not list {host}"))
    };
    let (codex, omp) = (listed("codex"), listed("omp"));
    vec![
        if codex {
            hooks_step("codex-hooks", "codex", root, shown, home)
        } else {
            unlisted("codex-hooks", "codex")
        },
        if codex {
            codex_config_step(root, shown, config_home)
        } else {
            unlisted("codex-config", "codex")
        },
        if codex {
            codex_skill_step(root, shown, home, &installed)
        } else {
            unlisted("codex-skill", "codex")
        },
        if omp {
            hooks_step("omp-hooks", "omp", root, shown, home)
        } else {
            unlisted("omp-hooks", "omp")
        },
        if omp {
            HostStep::bare("omp-config", "skip", OMP_CONFIG_REASON)
        } else {
            unlisted("omp-config", "omp")
        },
        if omp {
            omp_agents_step(root, shown, installed)
        } else {
            unlisted("omp-agents", "omp")
        },
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
            // Which entries of the old `moai-cli` set the write turns into `moai` (#3004 P4-2).
            "migration": plan["binary_migration"],
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
        Ok(plan) => {
            let mut step = config_from_plan(shown, &desired, &plan);
            if step.state == "conflict" {
                // The two values are in the step itself: reading the `diff` needs no command.
                if plan["diff"].is_array() {
                    step.preview["diff"] = plan["diff"].clone();
                }
                // The replacement's own plan (read-only) gives the digest its command needs;
                // without it the user still gets `keep` and `diff`.
                let replace = codex_config::config_plan_with(root, &home, request, true).ok();
                step.preview["choices"] = conflict_choices(shown, &desired, replace.as_ref());
            }
            step
        }
        Err(e) => HostStep::bare(NAME, "blocked", &e),
    }
}

/// What the user can pick when a key MoAI manages was set differently (SPEC-20261119-codex-
/// conflict-choices), recommended first. `keep` has no command; `diff` is the read-only replace
/// plan; `replace` is the apply of that plan, which keeps a copy of the file first, and is left
/// out when there is no replace plan (`None`, or one whose action is not `replace`).
fn conflict_choices(shown: &str, desired: &Value, replace: Option<&Value>) -> Value {
    let words = |list: &[&str]| list.iter().map(|w| w.to_string()).collect::<Vec<_>>();
    let mut choices = vec![
        json!({
            "id": "keep", "recommended": true, "writes": false,
            "label": "keep your Codex config as it is; nothing is written",
            "command": Value::Null,
        }),
        json!({
            "id": "diff", "recommended": false, "writes": false,
            "label": "show your value next to the one MoAI would write; nothing is written",
            "command": {
                "argv": words(&["moai", "codex-config-plan", shown, "--replace-with-backup"]),
                "stdin": desired,
            },
        }),
    ];
    if let Some((replace, digest)) = replace
        .filter(|r| r["action"] == "replace")
        .and_then(|r| Some((r, r["plan_digest"].as_str()?)))
    {
        choices.push(json!({
            "id": "replace", "recommended": false, "writes": true,
            "label": if replace["backup"]["path"].is_string() {
                "keep a copy of your Codex config, then set the MoAI value"
            } else {
                "set the MoAI value (there is no config file to keep a copy of)"
            },
            "backup": replace["backup"]["path"],
            "command": {
                "argv": words(&[
                    "moai", "codex-config-apply", shown, "--replace-with-backup",
                    "--plan-digest", digest, "--permission-reference", PLACEHOLDER,
                ]),
                "stdin": desired,
            },
        }));
    }
    Value::Array(choices)
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

/// The project copy of the `moai` skill for Codex (`PROJECT/.agents/skills/moai`), from the
/// skill package that came with this binary. Codex reads the user's own `~/.agents/skills`
/// too, so a `moai` skill there means the project needs no copy (user decision 6880072221
/// Q-H3). The step never writes HOME: it only looks for that one file. Without an installed
/// package there is nothing to copy (`skip`); without a HOME the user decides (`needs_decision`,
/// no command); a project that is the HOME or lies in a host configuration folder under it is
/// `blocked` (audit #3025 B2).
fn codex_skill_step(
    root: &Path,
    shown: &str,
    home: Option<&Path>,
    installed: &Result<Installed, String>,
) -> HostStep {
    const NAME: &str = "codex-skill";
    let installed = match installed {
        Ok(i) => i,
        Err(e) => return HostStep::bare(NAME, "skip", e),
    };
    let Some(home) = home else {
        return HostStep::bare(
            NAME,
            "needs_decision",
            "HOME is not set, so the user's own Codex skill folder cannot be checked; nothing is copied",
        );
    };
    // The same refusal as the hook steps (audit #3025 B2): a project that is the user's HOME, or
    // lies in a host's configuration folder under it, would make the printed apply write the
    // user's own skill folder, so no command is offered.
    if hooks::project_root(root, Some(home)).is_err() {
        return HostStep::bare(
            NAME,
            "blocked",
            "the project is the user's HOME or lies inside a host configuration folder under it, \
             or cannot be read; no project skill copy is offered",
        );
    }
    let user_copy = home
        .join(".agents")
        .join("skills")
        .join(update::SKILL_DIR)
        .join("SKILL.md");
    // `symlink_metadata`, not `exists`: a broken link there is still something the user put in
    // their own skill folder (audit #3025 N1), so the project offers no copy next to it.
    if user_copy.symlink_metadata().is_ok() {
        return HostStep::bare(
            NAME,
            "skip",
            "the user's Codex skill folder under HOME already has a moai skill; the project needs no copy",
        );
    }
    let request = project_skills::ProjectSkillRequest {
        protocol_version: project_skills::PROJECT_SKILL_SCHEMA_VERSION,
        opt_in: true,
        project_root: root.to_path_buf(),
        home: Some(home.to_path_buf()),
        hosts: vec!["codex".into()],
        source_dir: installed.package.clone(),
    };
    let Ok(stdin) = serde_json::to_value(&request) else {
        return HostStep::bare(NAME, "blocked", "a path is not UTF-8");
    };
    match project_skills::project_skill_plan(&request) {
        Ok(plan) => skill_from_plan(shown, &stdin, &plan),
        Err(e) => HostStep::bare(NAME, "blocked", &e),
    }
}

fn skill_from_plan(
    shown: &str,
    stdin: &Value,
    plan: &project_skills::ProjectSkillPlan,
) -> HostStep {
    const NAME: &str = "codex-skill";
    let Some(entry) = plan.entries.first().filter(|_| plan.entries.len() == 1) else {
        return HostStep::bare(NAME, "blocked", "the skill plan has no single Codex entry");
    };
    let (state, reason) = match entry.action.as_str() {
        "create" | "replace" => ("ready", "MoAI skill copy to write".to_string()),
        "unchanged" => ("unchanged", "the project skill copy is current".to_string()),
        "blocked" => (
            "blocked",
            entry
                .blocked_reason
                .clone()
                .unwrap_or_else(|| "the project skill folder is blocked".into()),
        ),
        _ => ("blocked", "the skill plan has no known action".to_string()),
    };
    let apply = (state == "ready").then(|| {
        (
            [
                "moai",
                "project-skill-apply",
                shown,
                "--plan-digest",
                plan.plan_digest.as_str(),
                "--permission-reference",
                PLACEHOLDER,
            ]
            .map(String::from)
            .to_vec(),
            stdin.clone(),
        )
    });
    HostStep {
        name: NAME,
        state,
        reason,
        preview: json!({
            "host": entry.host,
            "path": entry.path,
            "action": entry.action,
            "package_digest": plan.package_digest,
            "warnings": plan.warnings,
        }),
        plan_digest: Some(plan.plan_digest.clone()),
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
            None,
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
            None,
        )
        .iter()
        .filter(|s| s.name.ends_with("-hooks"))
        {
            assert_eq!(s.state, "ready", "{}: {}", s.name, s.reason);
        }
    }
    fn skill_plan(action: &str, reason: Option<&str>) -> project_skills::ProjectSkillPlan {
        project_skills::ProjectSkillPlan {
            protocol_version: 1,
            opt_in: true,
            package_digest: "sha256:package".into(),
            entries: vec![project_skills::ProjectSkillEntry {
                host: "codex".into(),
                path: ".agents/skills/moai".into(),
                action: action.into(),
                blocked_reason: reason.map(String::from),
                source_sha256: "sha256:package".into(),
                current_sha256: None,
                home_sha256: None,
                duplicated_with_home: false,
                reading_order_note: String::new(),
            }],
            warnings: Vec::new(),
            plan_digest: "sha256:fixture".into(),
        }
    }

    /// Only a copy MoAI will create or replace is offered for apply, and the printed command
    /// always carries both gate flags.
    #[test]
    fn the_skill_actions_set_the_step_state() {
        let stdin = json!({"hosts": ["codex"]});
        for (action, state, has_command) in [
            ("create", "ready", true),
            ("replace", "ready", true),
            ("unchanged", "unchanged", false),
            ("blocked", "blocked", false),
            ("something-new", "blocked", false),
        ] {
            let step = skill_from_plan("/p", &stdin, &skill_plan(action, Some("linked")));
            assert_eq!(step.state, state, "{action}");
            assert_eq!(step.apply.is_some(), has_command, "{action}");
        }
        let step = skill_from_plan("/p", &stdin, &skill_plan("create", None));
        let (argv, sent) = step.apply.unwrap();
        assert_eq!(
            &argv[..4],
            ["moai", "project-skill-apply", "/p", "--plan-digest"]
        );
        assert_eq!(argv[4], "sha256:fixture");
        assert_eq!(&argv[5..], ["--permission-reference", PLACEHOLDER]);
        assert_eq!(sent, stdin);
        let blocked = skill_from_plan("/p", &stdin, &skill_plan("blocked", Some("linked")));
        assert_eq!(blocked.reason, "linked");
    }

    /// Without a HOME the user's own Codex skill folder cannot be checked: the step waits for
    /// the user and prints no command.
    #[test]
    fn an_unknown_home_makes_the_codex_skill_step_wait_for_the_user() {
        let t = tempfile::tempdir().unwrap();
        let root = std::fs::canonicalize(t.path()).unwrap();
        let installed = Ok(Installed {
            package: root.clone(),
            agents: root.clone(),
        });
        let step = codex_skill_step(&root, "/p", None, &installed);
        assert_eq!(step.state, "needs_decision", "{}", step.reason);
        assert!(step.apply.is_none());
    }

    /// Audit #3025 B2: a project that is the HOME, or lies in a host configuration folder or
    /// the HOME state folder under it, is blocked with no command; a plain project under the
    /// same HOME is not refused for that reason.
    #[test]
    fn a_project_that_is_the_home_or_in_its_host_folders_is_blocked_for_the_skill_copy() {
        let t = tempfile::tempdir().unwrap();
        let home = std::fs::canonicalize(t.path()).unwrap();
        let installed = Ok(Installed {
            package: home.clone(),
            agents: home.clone(),
        });
        let mut refused = vec![home.clone()];
        for dir in [".agents", ".codex", ".omp", ".claude"] {
            let inner = home.join(dir).join("work");
            std::fs::create_dir_all(&inner).unwrap();
            refused.push(inner);
        }
        for root in &refused {
            let step = codex_skill_step(root, "/p", Some(&home), &installed);
            assert_eq!(step.state, "blocked", "{}: {}", root.display(), step.reason);
            assert!(step.apply.is_none(), "{}", root.display());
        }
        // A project beside them is not refused for this reason: the step goes on to the plan
        // (the empty package folder is then what blocks it, with another reason).
        let plain = home.join("project");
        std::fs::create_dir_all(&plain).unwrap();
        let step = codex_skill_step(&plain, "/p", Some(&home), &installed);
        assert!(
            !step.reason.contains("no project skill copy is offered"),
            "{}",
            step.reason
        );
    }

    /// Audit #3025 N1: a broken link where the user's own moai skill would be is something
    /// the user put there; the project offers no copy next to it.
    #[cfg(unix)]
    #[test]
    fn a_broken_link_in_the_users_skill_folder_counts_as_their_skill() {
        let t = tempfile::tempdir().unwrap();
        let base = std::fs::canonicalize(t.path()).unwrap();
        let home = base.join("home");
        let project = base.join("project");
        let user_skill = home.join(".agents").join("skills").join(update::SKILL_DIR);
        std::fs::create_dir_all(&user_skill).unwrap();
        std::fs::create_dir_all(&project).unwrap();
        std::os::unix::fs::symlink(base.join("missing"), user_skill.join("SKILL.md")).unwrap();
        let installed = Ok(Installed {
            package: base.clone(),
            agents: base.clone(),
        });
        let step = codex_skill_step(&project, "/p", Some(&home), &installed);
        assert_eq!(step.state, "skip", "{}", step.reason);
        assert!(step.apply.is_none());
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
    /// blocks the config step; it never previews a write. The same project with a home is ready,
    /// so the refusal comes from the home and not from the project.
    #[test]
    fn a_codex_home_that_cannot_be_told_blocks_the_config_step() {
        let project = tempfile::tempdir().unwrap();
        let root = std::fs::canonicalize(project.path()).unwrap();
        let shown = root.to_str().unwrap();
        let s = codex_config_step(
            &root,
            shown,
            Err("home directory missing; config target refused".into()),
        );
        assert_eq!(s.state, "blocked");
        assert!(s.reason.contains("home directory missing"), "{}", s.reason);
        assert!(s.apply.is_none() && s.plan_digest.is_none());
        let home = tempfile::tempdir().unwrap();
        let ok = codex_config_step(
            &root,
            shown,
            Ok(std::fs::canonicalize(home.path()).unwrap()),
        );
        assert_eq!(ok.state, "ready", "{}", ok.reason);
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

    fn ids(choices: &Value) -> Vec<&str> {
        choices
            .as_array()
            .unwrap()
            .iter()
            .map(|c| c["id"].as_str().unwrap())
            .collect()
    }

    /// SPEC-20261119-codex-conflict-choices REQ-10: no replace plan (it could not be made, or it
    /// is not a replacement) leaves `keep` and `diff`; a replace plan adds `replace` last.
    #[test]
    fn conflict_choices_drop_replace_without_a_replace_plan() {
        let want = codex_config_request();
        const BACKUP: &str = crate::project_state!("backups/x.toml");
        let replace = json!({"action": "replace", "plan_digest": "sha256:d",
                             "backup": {"path": BACKUP}});
        let all = conflict_choices("/p", &want, Some(&replace));
        assert_eq!(ids(&all), ["keep", "diff", "replace"]);
        assert_eq!(all[2]["backup"], json!(BACKUP));
        for none_or_other in [
            None,
            Some(json!({"action": "conflict", "plan_digest": "sha256:d"})),
            Some(json!({"action": "replace"})),
        ] {
            let kept = conflict_choices("/p", &want, none_or_other.as_ref());
            assert_eq!(ids(&kept), ["keep", "diff"], "{none_or_other:?}");
        }
    }

    /// REQ-9, REQ-11: `keep` is first, recommended and has no command; only `replace` writes;
    /// the `replace` command carries the digest of the plan it was built from.
    #[test]
    fn conflict_choices_put_keep_first_and_only_replace_writes() {
        let want = codex_config_request();
        let replace = json!({"action": "replace", "plan_digest": "sha256:d", "backup": null});
        let all = conflict_choices("/p", &want, Some(&replace));
        let list = all.as_array().unwrap();
        assert_eq!(list[0]["recommended"], json!(true));
        assert_eq!(list[0]["command"], Value::Null);
        let recommended = list.iter().filter(|c| c["recommended"] == json!(true));
        assert_eq!(recommended.count(), 1);
        let writes: Vec<&str> = list
            .iter()
            .filter(|c| c["writes"] == json!(true))
            .map(|c| c["id"].as_str().unwrap())
            .collect();
        assert_eq!(writes, ["replace"]);
        let argv = list[2]["command"]["argv"].as_array().unwrap();
        let at = argv.iter().position(|w| w == "--plan-digest").unwrap();
        assert_eq!(argv[at + 1], json!("sha256:d"));
    }

    /// SPEC-20261120-hooks-conflict-choices REQ-12: a hook conflict has choices only when there is
    /// a replace plan (an action `replace` with a digest); every other plan gives none, and the
    /// step stays a bare conflict.
    #[test]
    fn hooks_choices_drop_replace_without_a_replace_plan() {
        const BACKUP: &str = crate::project_state!("backups/codex-hooks-x.json");
        let replace = json!({"action": "replace", "plan_digest": "sha256:d",
                             "backup": {"path": BACKUP}});
        let all = hooks_conflict_choices("/p", "codex", Some(&replace)).unwrap();
        assert_eq!(ids(&all), ["keep", "diff", "replace"]);
        assert_eq!(all[2]["backup"], json!(BACKUP));
        for none_or_other in [
            None,
            Some(json!({"action": "conflict", "plan_digest": "sha256:d"})),
            Some(json!({"action": "replace"})),
        ] {
            assert!(
                hooks_conflict_choices("/p", "codex", none_or_other.as_ref()).is_none(),
                "{none_or_other:?}"
            );
        }
    }

    /// REQ-11: `keep` is first, recommended and has no command; only `replace` writes; both
    /// commands name the host and the flag, and `replace` carries the digest of its plan.
    #[test]
    fn hooks_choices_put_keep_first_and_only_replace_writes() {
        let replace = json!({"action": "replace", "plan_digest": "sha256:d", "backup": null});
        for host in ["codex", "omp"] {
            let all = hooks_conflict_choices("/p", host, Some(&replace)).unwrap();
            let list = all.as_array().unwrap();
            assert_eq!(list[0]["recommended"], json!(true));
            assert_eq!(list[0]["command"], Value::Null);
            let recommended = list.iter().filter(|c| c["recommended"] == json!(true));
            assert_eq!(recommended.count(), 1);
            let writes: Vec<&str> = list
                .iter()
                .filter(|c| c["writes"] == json!(true))
                .map(|c| c["id"].as_str().unwrap())
                .collect();
            assert_eq!(writes, ["replace"]);
            for (i, verb) in [(1, "hooks-plan"), (2, "hooks-apply")] {
                let argv = list[i]["command"]["argv"].as_array().unwrap();
                assert_eq!(&argv[..2], [json!("moai"), json!(verb)], "{host}");
                assert!(argv.contains(&json!("--replace-with-backup")), "{argv:?}");
                let at = argv.iter().position(|w| w == "--host").unwrap();
                assert_eq!(argv[at + 1], json!(host));
                assert_eq!(list[i]["command"]["stdin"], Value::Null);
            }
            let argv = list[2]["command"]["argv"].as_array().unwrap();
            let at = argv.iter().position(|w| w == "--plan-digest").unwrap();
            assert_eq!(argv[at + 1], json!("sha256:d"));
            assert!(argv.contains(&json!(PLACEHOLDER)));
        }
    }
}
