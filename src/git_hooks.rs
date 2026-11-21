//! The project's git commit-msg hook (SPEC-20261120-commit-message-card-numbers, repair
//! slices): where git looks for hooks, whether the shipped hook is already there, and
//! installing it. `project-update`'s git-hook step reads [`plan`]; the operation
//! `git-hook-apply` (`src/ops/git_hook_apply.rs`) runs [`install`]. The shipped hook text is
//! compiled in from `install/commit-msg`, so every installed moai carries it without this
//! repository; a user's own hook is never replaced — a different `commit-msg` is
//! `needs_decision`.

use serde_json::{json, Value};
use std::fs;
use std::io::Write;
use std::path::{Path, PathBuf};
use std::process::Command;

/// The hook file the product ships, compiled in (the manual-install copy lives beside the
/// installer scripts; this is the same bytes).
pub const HOOK: &str = include_str!("../install/commit-msg");

/// The file name git runs as the commit-message hook.
pub const HOOK_NAME: &str = "commit-msg";

/// `git rev-parse --git-path hooks` in `root` — git's own hook folder, honoring a
/// `core.hooksPath` setting — made absolute against `root` when git names a relative path.
/// The spawn goes through `scrub_git_env` and reads only; a non-repository is the refusal.
pub fn hooks_dir(root: &Path) -> Result<PathBuf, String> {
    let mut cmd = Command::new("git");
    cmd.arg("-C")
        .arg(root)
        .args(["rev-parse", "--git-path", "hooks"]);
    crate::identity::scrub_git_env(&mut cmd);
    let out = cmd
        .output()
        .map_err(|e| format!("git-hook: git could not be run: {e}"))?;
    if !out.status.success() {
        return Err(format!(
            "git-hook: {} is not a git repository (git rev-parse failed)",
            root.display()
        ));
    }
    let named = String::from_utf8_lossy(&out.stdout).trim().to_string();
    if named.is_empty() {
        return Err("git-hook: git named no hooks folder".to_string());
    }
    let p = PathBuf::from(named);
    Ok(if p.is_absolute() { p } else { root.join(p) })
}

/// What the hook folder holds: `install` when no `commit-msg` is there, `unchanged` when the
/// shipped bytes already are, `needs_decision` when a different hook file is present, and
/// `none` when the folder cannot be used (a link or a file on the folder's path is refused,
/// never followed). `needs_decision` names nothing but the decision.
pub fn plan(root: &Path) -> Result<Value, String> {
    let dir = hooks_dir(root)?;
    match fs::symlink_metadata(&dir) {
        Ok(m) if m.is_dir() => {}
        Ok(_) | Err(_) => {
            return Ok(json!({
                "action": "none",
                "reason": "the hooks folder is missing or not a folder; nothing is created here",
                "hooks_dir": dir.to_string_lossy(),
            }))
        }
    }
    let hook = dir.join(HOOK_NAME);
    let state = match fs::symlink_metadata(&hook) {
        Ok(m) if m.is_file() => {
            let bytes = fs::read(&hook)
                .map_err(|e| format!("git-hook: {} could not be read: {e}", hook.display()))?;
            if bytes == HOOK.as_bytes() {
                "unchanged"
            } else {
                "needs_decision"
            }
        }
        // A link or a folder named commit-msg is a user's own hook: never replaced.
        Ok(_) => "needs_decision",
        Err(_) => "install",
    };
    Ok(json!({
        "action": state,
        "hooks_dir": dir.to_string_lossy(),
        "hook": hook.to_string_lossy(),
    }))
}

/// Installs the shipped hook: writes the compiled-in bytes, mode 0700 bits off — 0o755 — and
/// nothing else. The caller has read [`plan`]; a hook that appeared meanwhile (or a different
/// one) is refused, so a plan and an apply racing stay safe.
pub fn install(root: &Path) -> Result<Value, String> {
    let p = plan(root)?;
    let action = p["action"].as_str().unwrap_or_default();
    if action != "install" {
        return Err(format!(
            "git-hook apply refused: the plan answers `{action}`"
        ));
    }
    let hook = PathBuf::from(p["hook"].as_str().unwrap_or_default());
    let mut options = fs::OpenOptions::new();
    options.write(true).create_new(true);
    #[cfg(unix)]
    {
        use std::os::unix::fs::OpenOptionsExt;
        options.mode(0o755);
    }
    let mut file = options.open(&hook).map_err(|e| {
        format!(
            "git-hook apply failed: {} is in the way: {e}",
            hook.display()
        )
    })?;
    file.write_all(HOOK.as_bytes())
        .map_err(|e| format!("git-hook apply failed: {e}"))?;
    Ok(json!({
        "action": "install",
        "hook": hook.to_string_lossy(),
        "bytes": HOOK.len(),
        "writes": true,
    }))
}

#[cfg(test)]
mod tests {
    use super::*;
    #[cfg(unix)]
    use std::os::unix::fs::PermissionsExt;

    /// A fixture git repository, like `tests/common/git.rs` makes one: `git init` with the
    /// maintenance guards, so a fixture keeps stable under parallel tests.
    fn repo() -> (tempfile::TempDir, PathBuf) {
        let dir = tempfile::TempDir::new().unwrap();
        let root = dir.path().join("repo");
        fs::create_dir_all(&root).unwrap();
        let mut cmd = Command::new("git");
        cmd.args([
            "-c",
            "gc.auto=0",
            "-c",
            "maintenance.auto=false",
            "init",
            "-q",
        ]);
        cmd.arg(&root);
        crate::identity::scrub_git_env(&mut cmd);
        let out = cmd.output().unwrap();
        assert!(
            out.status.success(),
            "git init failed: {}",
            String::from_utf8_lossy(&out.stderr)
        );
        (dir, root)
    }

    #[test]
    fn plan_answers_install_then_unchanged_and_never_replaces() {
        let (_d, root) = repo();
        let first = plan(&root).unwrap();
        assert_eq!(first["action"], "install", "{first}");
        let installed = install(&root).unwrap();
        assert_eq!(installed["action"], "install", "{installed}");
        let hook = PathBuf::from(first["hook"].as_str().unwrap());
        assert_eq!(fs::read(&hook).unwrap(), HOOK.as_bytes());
        let mode = fs::metadata(&hook).unwrap().permissions().mode();
        assert_eq!(mode & 0o111, 0o111, "the hook is executable");
        // A second plan is `unchanged`; install refuses it.
        assert_eq!(plan(&root).unwrap()["action"], "unchanged");
        assert!(install(&root).is_err(), "a second install is refused");
    }

    #[test]
    fn a_user_hook_is_needs_decision_and_stays() {
        let (_d, root) = repo();
        let first = plan(&root).unwrap();
        let hook = PathBuf::from(first["hook"].as_str().unwrap());
        fs::write(&hook, "#!/bin/sh\n# the user's own hook\n").unwrap();
        assert_eq!(plan(&root).unwrap()["action"], "needs_decision");
        assert!(install(&root).is_err());
        assert_eq!(
            fs::read(&hook).unwrap(),
            b"#!/bin/sh\n# the user's own hook\n".to_vec(),
            "the user's hook stays"
        );
    }

    #[test]
    fn a_non_repository_refuses() {
        let dir = tempfile::TempDir::new().unwrap();
        let err = plan(dir.path()).unwrap_err();
        assert!(err.contains("not a git repository"), "{err}");
    }
}
