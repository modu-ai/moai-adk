//! The host steps of `project-update-plan`: one information step about the retired Codex and OMP
//! hosts. MoAI no longer plans or writes anything for them; the step only lists the files earlier
//! versions left in the project, so the user can see them and decide. Nothing is written here.
//!
//! `project_update` turns a [`HostStep`] into its own step type; this module knows nothing
//! about phases.

use serde_json::{json, Value};
use std::path::Path;

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

/// The host steps for the project `root`: the one information step, always present.
pub fn steps(root: &Path) -> Vec<HostStep> {
    vec![retired_hosts_step(root)]
}

/// The files and folders earlier MoAI versions wrote for the Codex and OMP hosts.
const RETIRED_HOST_PATHS: [&str; 6] = [
    ".codex/hooks.json",
    ".codex/config.toml",
    ".codex/agents",
    ".agents/skills/moai",
    ".omp/extensions/moai-hooks.ts",
    ".omp/agents",
];

const RETIRED_HOSTS_REASON: &str = "MoAI no longer manages these files and leaves them as they are";

/// The one information step about the retired hosts: it lists those of the paths MoAI used to
/// write that exist under `root`, sorted. A path counts as present by `symlink_metadata`, so a
/// link is listed and never followed and no file content is read. It writes nothing and has no
/// command to apply.
pub fn retired_hosts_step(root: &Path) -> HostStep {
    let mut files: Vec<&str> = RETIRED_HOST_PATHS
        .into_iter()
        .filter(|rel| std::fs::symlink_metadata(root.join(rel)).is_ok())
        .collect();
    files.sort_unstable();
    let mut step = HostStep::bare("retired-hosts", "skip", RETIRED_HOSTS_REASON);
    step.preview = json!({"action": "info", "files": files});
    step
}

#[cfg(test)]
mod tests {
    use super::*;

    /// Every path under `root`, relative and sorted, links not followed.
    fn tree_listing(root: &Path) -> Vec<String> {
        fn walk(dir: &Path, root: &Path, out: &mut Vec<String>) {
            for entry in std::fs::read_dir(dir).unwrap() {
                let path = entry.unwrap().path();
                out.push(path.strip_prefix(root).unwrap().display().to_string());
                if std::fs::symlink_metadata(&path).unwrap().is_dir() {
                    walk(&path, root, out);
                }
            }
        }
        let mut out = Vec::new();
        walk(root, root, &mut out);
        out.sort();
        out
    }

    /// The host steps of a project are the one information step and nothing else.
    #[test]
    fn the_host_steps_are_the_one_information_step() {
        let t = tempfile::tempdir().unwrap();
        let all = steps(t.path());
        assert_eq!(
            all.iter().map(|s| s.name).collect::<Vec<_>>(),
            ["retired-hosts"]
        );
    }

    /// An empty project still gets the step, with an empty list and no command.
    #[test]
    fn the_retired_hosts_step_of_an_empty_project_lists_nothing() {
        let t = tempfile::tempdir().unwrap();
        let step = retired_hosts_step(t.path());
        assert_eq!(step.name, "retired-hosts");
        assert_eq!(step.state, "skip");
        assert!(!step.reason.is_empty());
        assert_eq!(step.preview, json!({"action": "info", "files": []}));
        assert!(step.apply.is_none() && step.plan_digest.is_none());
    }

    /// Files and folders both count; the list is sorted and holds only what exists.
    #[test]
    fn the_retired_hosts_step_lists_what_exists_sorted() {
        let t = tempfile::tempdir().unwrap();
        std::fs::create_dir_all(t.path().join(".codex")).unwrap();
        std::fs::write(t.path().join(".codex/config.toml"), "").unwrap();
        std::fs::create_dir_all(t.path().join(".omp/agents")).unwrap();
        std::fs::create_dir_all(t.path().join(".agents/skills/moai")).unwrap();
        let step = retired_hosts_step(t.path());
        assert_eq!(step.state, "skip");
        assert_eq!(
            step.preview,
            json!({"action": "info",
                   "files": [".agents/skills/moai", ".codex/config.toml", ".omp/agents"]})
        );
    }

    /// A link is listed by its own metadata, even when its target is missing.
    #[cfg(unix)]
    #[test]
    fn the_retired_hosts_step_lists_a_dangling_link() {
        let t = tempfile::tempdir().unwrap();
        std::fs::create_dir_all(t.path().join(".codex")).unwrap();
        std::os::unix::fs::symlink("missing-target", t.path().join(".codex/hooks.json")).unwrap();
        assert!(!t.path().join(".codex/hooks.json").exists());
        let step = retired_hosts_step(t.path());
        assert_eq!(
            step.preview,
            json!({"action": "info", "files": [".codex/hooks.json"]})
        );
    }

    /// The step reads nothing but metadata and writes nothing.
    #[test]
    fn the_retired_hosts_step_writes_nothing() {
        let t = tempfile::tempdir().unwrap();
        std::fs::create_dir_all(t.path().join(".omp/extensions")).unwrap();
        std::fs::write(t.path().join(".omp/extensions/moai-hooks.ts"), "x").unwrap();
        std::fs::create_dir_all(t.path().join(".codex/agents")).unwrap();
        let before = tree_listing(t.path());
        let step = retired_hosts_step(t.path());
        assert_eq!(tree_listing(t.path()), before);
        assert_eq!(
            step.preview,
            json!({"action": "info",
                   "files": [".codex/agents", ".omp/extensions/moai-hooks.ts"]})
        );
        assert_eq!(
            std::fs::read_to_string(t.path().join(".omp/extensions/moai-hooks.ts")).unwrap(),
            "x"
        );
    }
}
