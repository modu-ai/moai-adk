//! The host steps of `project-update-plan` (#2936 R9 S5a): the Codex and OMP hook sets and
//! the `.omp/config.yml` line. Each step calls the plan the host's own writer takes
//! (`hooks::plan`) and prints that writer's command; nothing is written here.
//!
//! `project_update` turns a [`HostStep`] into its own step type; this module knows nothing
//! about phases. HOME is read here, so `project_update::plan` keeps its signature.

use crate::hooks;
use serde_json::{json, Value};
use std::path::{Path, PathBuf};

/// The one placeholder a printed command carries (the same word `project_update` prints).
const PLACEHOLDER: &str = crate::project_update::PERMISSION_PLACEHOLDER;

/// The step names this module makes, in plan order.
pub const NAMES: [&str; 3] = ["codex-hooks", "omp-hooks", "omp-config"];

const OMP_CONFIG_REASON: &str = "MoAI writes no .omp/config.yml: no status line key is set \
     (#2936 S2, default A) and an existing file is never read or changed";

/// One host step; `apply` is the printed command as a word list, never with stdin.
pub struct HostStep {
    pub name: &'static str,
    pub state: &'static str,
    pub reason: String,
    pub preview: Value,
    pub plan_digest: Option<String>,
    pub apply: Option<Vec<String>>,
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
    steps_in(root, shown, home.as_deref())
}

/// [`steps`] with the user's HOME given: a project that is the HOME, or lies in a host's own
/// settings folder under it, is refused by `hooks::plan` and shows as `blocked`.
fn steps_in(root: &Path, shown: &str, home: Option<&Path>) -> Vec<HostStep> {
    vec![
        hooks_step("codex-hooks", "codex", root, shown, home),
        hooks_step("omp-hooks", "omp", root, shown, home),
        HostStep::bare("omp-config", "skip", OMP_CONFIG_REASON),
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
        (Some(d), "ready") => Some(
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
        ),
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
        let argv = step.apply.expect("a ready step prints its command");
        assert_eq!(&argv[..2], ["moai", "hooks-apply"]);
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
        for s in steps_in(&root, shown, Some(&root))
            .iter()
            .filter(|s| s.name.ends_with("-hooks"))
        {
            assert_eq!(s.state, "blocked", "{}: {}", s.name, s.reason);
            assert!(s.apply.is_none(), "{}", s.name);
        }
        // The same folder as a project of a different HOME is previewed.
        let other = tempfile::tempdir().unwrap();
        let home = std::fs::canonicalize(other.path()).unwrap();
        for s in steps_in(&root, shown, Some(&home))
            .iter()
            .filter(|s| s.name.ends_with("-hooks"))
        {
            assert_eq!(s.state, "ready", "{}: {}", s.name, s.reason);
        }
    }
}
