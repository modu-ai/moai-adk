//! First-launch project initialization for `cc|glm|codex` (Issue #2481; user decisions
//! 6810797000 and 6810818827, 2026-11-12).
//!
//! A launch in a folder that is not a git repository runs `git init` there without asking,
//! then creates the MoAI project settings and `AGENTS.md` through the existing code paths —
//! `profile::plan`/`apply` (`project-plan`/`project-apply`) and
//! `generation::generation_plan`/`generation_apply` (`generation-plan`/`generation-apply`,
//! the ordinary-project order of #2274). Nothing new is generated here: this module only
//! decides whether to run them and reports what it did on one line. Files that already exist
//! are never touched, user-level settings and keys are not read or written, and host programs
//! are not installed. HOME itself and the file-system root are never initialized. Without git
//! on PATH the launcher stops with the install guidance for this OS.
use crate::generation::{generation_apply, generation_plan, GenerationRequest};
use crate::{identity::identity, profile, PROJECT_STATE_DIR};
use serde_json::{json, Value};
use std::{
    fs::{self, OpenOptions},
    path::{Path, PathBuf},
    process::{Command, Stdio},
};

/// Prefix of the one stderr line a launch prints after it initialized something.
pub const INIT_LINE: &str = "moai-cli: initialized";
/// Recorded in the apply receipts as the reason for the writes (the user decision).
const PERMISSION: &str = "launcher-auto-init #2481 user decision 6810818827";
/// The generated workflow id; AGENTS.md points at `.moai-cli/workflows/moai/`.
const WORKFLOW_ID: &str = "moai";
/// Test-only: milliseconds a first launch holds `.init-lock` between writing the settings and
/// generating `AGENTS.md` (at most 3000), so a test can show a concurrent launch waits.
pub const INIT_TEST_HOLD_VAR: &str = "MOAI_CLI_TEST_INIT_HOLD_MS";
/// Test-only (Issue #2504, not implemented yet in this RED commit).
pub const INIT_TEST_FAIL_VAR: &str = "MOAI_CLI_TEST_INIT_FAIL_AFTER_PROFILE";

/// What [`prepare`] did or, under `--dry-run`, would do.
#[derive(Debug, Default)]
pub struct Prepared {
    /// One human line for stderr and the session notes; `None` when nothing applies.
    pub note: Option<String>,
    /// `--dry-run` in a folder that is not a repository yet: `git init` is still pending, so
    /// the agent-number store does not exist and the next number is 1.
    pub git_pending: bool,
}

/// Initializes `cwd` for a launch on `host` (`claude`, `glm` or `codex`). Refuses — before any
/// write — when git is missing from PATH, or when `cwd` is not a repository and is HOME itself
/// or the file-system root. Under `dry_run` it writes nothing and reports the plan.
pub fn prepare(cwd: &Path, host: &str, dry_run: bool) -> Result<Prepared, String> {
    // An identity error (a git failure, a refused path) initializes nothing and leaves the
    // launch as it was before #2481: numbering reports its own error, other launches go on.
    let Ok(id) = identity(cwd) else {
        return Ok(Prepared::default());
    };
    let home = crate::key_vault::home()
        .ok()
        .and_then(|h| h.canonicalize().ok());
    let guard = |dir: &Path| home_or_root(dir, home.as_deref());
    match id["identity_status"].as_str() {
        Some("git_unavailable") => Err(install_guidance()),
        Some("no_git") => {
            let dir = cwd
                .canonicalize()
                .map_err(|_| "project check: the current directory is unreadable".to_string())?;
            if let Some(what) = guard(&dir) {
                return Err(format!(
                    "moai-cli: {} is {what}; it is not turned into a git repository. Run the launcher in a project folder, or run `git init` here yourself first.",
                    dir.display()
                ));
            }
            if dry_run {
                return Ok(Prepared {
                    note: Some(format!(
                        "would initialize {}: git init, {PROJECT_STATE_DIR}/project.yaml, AGENTS.md (--dry-run writes nothing)",
                        dir.display()
                    )),
                    git_pending: true,
                });
            }
            git_init(&dir)?;
            let mut done = vec!["git init".to_string()];
            done.extend(settings(&dir, &dir.join(".git"), host)?);
            Ok(Prepared {
                note: Some(format!(
                    "{INIT_LINE} {}: {}",
                    dir.display(),
                    done.join(", ")
                )),
                git_pending: false,
            })
        }
        Some("ok") => {
            let Some(top) = id["work_tree"].as_str().map(PathBuf::from) else {
                return Ok(Prepared::default()); // bare repository or inside `.git`
            };
            if exists(&top.join(PROJECT_STATE_DIR).join("project.yaml")) {
                // A first launch that is still generating AGENTS.md holds `.init-lock` (created
                // before the settings file): wait for it, so this host does not start before the
                // initialization it found has finished. No lock file, no wait and no write.
                if let Some(common) = id["clone_anchor"].as_str() {
                    wait_for_initializer(&init_lock_dir(Path::new(common)).join(".init-lock"));
                }
                return Ok(Prepared::default());
            }
            if let Some(what) = guard(&top) {
                return Ok(Prepared {
                    note: Some(format!(
                        "moai-cli: {} is {what}; MoAI project settings are not created there",
                        top.display()
                    )),
                    git_pending: false,
                });
            }
            if dry_run {
                let agents = if exists(&top.join("AGENTS.md")) {
                    ""
                } else {
                    ", AGENTS.md"
                };
                return Ok(Prepared {
                    note: Some(format!(
                        "would initialize {}: {PROJECT_STATE_DIR}/project.yaml{agents} (--dry-run writes nothing)",
                        top.display()
                    )),
                    git_pending: false,
                });
            }
            let anchor = id["clone_anchor"]
                .as_str()
                .map_or_else(|| top.join(".git"), PathBuf::from);
            let done = settings(&top, &anchor, host)?;
            Ok(Prepared {
                note: (!done.is_empty())
                    .then(|| format!("{INIT_LINE} {}: {}", top.display(), done.join(", "))),
                git_pending: false,
            })
        }
        _ => Ok(Prepared::default()),
    }
}

/// `Some(description)` when `dir` (canonical) is a file-system root or `home` (HOME, already
/// resolved, so a HOME given through a symlink is recognized).
pub fn home_or_root(dir: &Path, home: Option<&Path>) -> Option<&'static str> {
    if dir.parent().is_none() {
        return Some("the file-system root");
    }
    (home == Some(dir)).then_some("your HOME folder")
}

fn exists(path: &Path) -> bool {
    fs::symlink_metadata(path).is_ok()
}

/// `<clone common dir>/.moai-cli/sessions`, the folder of the agent-number store
/// (`session::agent_seq_dir`), read from an identity already resolved.
fn init_lock_dir(common: &Path) -> PathBuf {
    common.join(".moai-cli").join("sessions")
}

/// Waits until no first launch holds `lock` (a regular file only; never created here).
fn wait_for_initializer(lock: &Path) {
    if !fs::symlink_metadata(lock).is_ok_and(|m| m.file_type().is_file()) {
        return;
    }
    if let Ok(f) = OpenOptions::new().read(true).open(lock) {
        let _ = f.lock_shared();
    }
}

/// `git init -q` in `dir` with the same variable scrub as every other git subprocess. A
/// concurrent launch may have initialized the folder meanwhile: a failed `init` that leaves a
/// repository behind is accepted.
fn git_init(dir: &Path) -> Result<(), String> {
    let mut cmd = Command::new("git");
    cmd.arg("init").arg("-q").arg(dir);
    crate::identity::scrub_git_env(&mut cmd);
    let ok = cmd
        .stdin(Stdio::null())
        .stdout(Stdio::null())
        .stderr(Stdio::null())
        .status()
        .is_ok_and(|s| s.success());
    if ok || identity(dir).is_ok_and(|id| id["identity_status"] == "ok") {
        Ok(())
    } else {
        Err(format!(
            "moai-cli: git init failed in {}; run `git init` there yourself or pass --name",
            dir.display()
        ))
    }
}

/// Creates the project settings and, when `AGENTS.md` is absent, the generated workflow and
/// `AGENTS.md` in work tree `top`, under a lock in the clone's common dir so concurrent first
/// launches initialize once. Returns what it created.
fn settings(top: &Path, anchor: &Path, host: &str) -> Result<Vec<String>, String> {
    let lock_dir = crate::session::agent_seq_dir(top)?;
    for p in [&lock_dir, &lock_dir.join(".init-lock")] {
        if fs::symlink_metadata(p).is_ok_and(|m| m.file_type().is_symlink()) {
            return Err(format!("{} is a symlink; refused", p.display()));
        }
    }
    fs::create_dir_all(&lock_dir)
        .map_err(|_| format!("{} cannot be created", lock_dir.display()))?;
    let lock = OpenOptions::new()
        .read(true)
        .write(true)
        .create(true)
        .truncate(false)
        .open(lock_dir.join(".init-lock"))
        .and_then(|f| f.lock().map(|()| f))
        .map_err(|_| "the project initialization lock is unavailable".to_string())?;
    let mut done = Vec::new();
    if !exists(&top.join(PROJECT_STATE_DIR).join("project.yaml")) {
        let plan = profile::plan(top, default_profile(&primary_name(anchor), host)?)?;
        if plan.action != "create" {
            return Err(format!(
                "project settings plan is {}, not create; nothing written",
                plan.action
            ));
        }
        let digest = plan.plan_digest.clone();
        profile::apply(top, plan, &digest, PERMISSION)?;
        done.push(format!("{PROJECT_STATE_DIR}/project.yaml"));
        if let Some(ms) = std::env::var(INIT_TEST_HOLD_VAR)
            .ok()
            .and_then(|v| v.parse::<u64>().ok())
        {
            std::thread::sleep(std::time::Duration::from_millis(ms.min(3000)));
        }
        if exists(&top.join("AGENTS.md")) {
            done.push("AGENTS.md present, unchanged".into());
        } else {
            done.extend(agents(top)?);
        }
    }
    drop(lock);
    Ok(done)
}

/// `generation-plan`/`generation-apply` with the workflow under `.moai-cli/workflows/moai/`.
/// A plan with any conflict (a user file already on a generated path) is not applied.
fn agents(top: &Path) -> Result<Vec<String>, String> {
    let request = GenerationRequest {
        protocol_version: 1,
        language: "en".into(),
        workflow_reference: format!("{PROJECT_STATE_DIR}/workflows/{WORKFLOW_ID}/"),
        local_file: "none".into(),
        workflow_id: WORKFLOW_ID.into(),
        native_roles: Vec::new(),
        skill_projections: Vec::new(),
    };
    let plan = generation_plan(top, request)?;
    if plan.files.iter().any(|f| f.action == "conflict") {
        return Ok(vec![
            "AGENTS.md not generated (a generated path is already in use)".into(),
        ]);
    }
    let digest = plan.plan_digest.clone();
    generation_apply(top, plan, &digest, PERMISSION)?;
    Ok(vec!["AGENTS.md".into()])
}

/// The primary folder name exactly as `key_vault::project_scope` derives it from the clone
/// common dir `anchor`: the folder holding `.git` (else `anchor` itself), with trailing `.git`
/// removed. Keeping both the same means writing the profile does not move the vault folder.
fn primary_name(anchor: &Path) -> String {
    let primary = if anchor.file_name().is_some_and(|n| n == ".git") {
        anchor.parent().unwrap_or(anchor)
    } else {
        anchor
    };
    primary
        .file_name()
        .map(|n| n.to_string_lossy().trim_end_matches(".git").to_string())
        .unwrap_or_default()
}

/// The profile a first launch writes: `project_id` from the primary folder name with the key
/// vault's folder-name mapping (so the vault folder name does not change), `hosts` from the
/// launching host, and placeholder values marked `source: default` that the project interview
/// replaces.
fn default_profile(folder: &str, host: &str) -> Result<profile::Profile, String> {
    let named = crate::key_vault::folder_name(None, folder, "");
    let project_id = named.strip_suffix('-').unwrap_or(&named).to_string();
    let host = if host == "codex" { "codex" } else { "claude" };
    let value: Value = json!({
        "schema_version": 2,
        "project_id": project_id,
        "purpose": "Created by moai-cli on first launch; refine it in the project interview",
        "tools": ["unspecified"],
        "development_methods": ["unspecified"],
        "hosts": [host],
        "decisions": [
            {"id": "auto-init-tools", "topic": "tools", "value": "unspecified", "source": "default", "scope": "project"},
            {"id": "auto-init-methods", "topic": "development_methods", "value": "unspecified", "source": "default", "scope": "project"},
            {"id": "auto-init-hosts", "topic": "hosts", "value": host, "source": "default", "scope": "project"}
        ]
    });
    serde_json::from_value(value).map_err(|_| "default project profile is invalid".into())
}

/// The git install guidance for this OS, from git-scm.com (read 2026-11-12, Issue #2481).
pub fn install_guidance() -> String {
    let body = if cfg!(target_os = "macos") {
        "  xcode-select --install     (Apple Command Line Tools)\n  brew install git           (Homebrew)\n  sudo port install git      (MacPorts)\nSource: https://git-scm.com/install/mac"
    } else if cfg!(windows) {
        "  winget install --id Git.Git -e --source winget\n  or the Git for Windows installer\nSource: https://git-scm.com/install/windows"
    } else if cfg!(target_os = "linux") {
        "  Debian/Ubuntu: sudo apt-get install git\n  Fedora:        sudo dnf install git\n  Arch Linux:    sudo pacman -S git\n  openSUSE:      sudo zypper install git\n  Alpine:        sudo apk add git\n  other distributions: see the source page\nSource: https://git-scm.com/install/linux"
    } else {
        "Source: https://git-scm.com/install"
    };
    format!("moai-cli: git is not installed (not found on PATH). Install it, then run the launcher again:\n{body}")
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn the_file_system_root_is_never_a_project_folder() {
        let root = if cfg!(windows) {
            PathBuf::from(r"C:\")
        } else {
            PathBuf::from("/")
        };
        let home = root.join("home-x");
        assert_eq!(home_or_root(&root, None), Some("the file-system root"));
        assert_eq!(
            home_or_root(&root, Some(&home)),
            Some("the file-system root")
        );
        assert_eq!(home_or_root(&home, Some(&home)), Some("your HOME folder"));
        assert!(home_or_root(&home.join("project"), Some(&home)).is_none());
        assert!(home_or_root(&home, None).is_none());
    }

    #[test]
    fn the_first_launch_profile_passes_the_profile_rules() {
        let p = default_profile("My Project", "glm").unwrap();
        assert_eq!(p.project_id, "my-project");
        assert_eq!(p.hosts, ["claude"]);
        let p = default_profile("___", "codex").unwrap();
        assert_eq!(p.project_id, "project");
        assert_eq!(p.hosts, ["codex"]);
    }

    #[test]
    fn writing_the_first_launch_profile_keeps_the_key_vault_folder_name() {
        // #2504 AC-2: the key vault names a project folder from the profile's project_id when
        // there is one, else from the primary folder name. The first-launch project_id must make
        // both give the same folder, so writing the profile never moves the vault folder.
        let hex = "0123456789abcdef";
        for folder in [
            "My Project",
            "example.git",
            "a__b",
            "---",
            "Ünïcode Dir",
            "x".repeat(60).as_str(),
            "Mixed.Case-Name_1",
        ] {
            let anchor = PathBuf::from("/x").join(folder).join(".git");
            let name = primary_name(&anchor);
            let p = default_profile(&name, "claude").unwrap();
            assert_eq!(
                crate::key_vault::folder_name(Some(&p.project_id), &name, hex),
                crate::key_vault::folder_name(None, &name, hex),
                "{folder}"
            );
        }
    }

    #[test]
    fn the_primary_name_matches_the_key_vault_derivation() {
        assert_eq!(primary_name(Path::new("/x/example.git/.git")), "example");
        assert_eq!(primary_name(Path::new("/x/plain/.git")), "plain");
        assert_eq!(primary_name(Path::new("/x/bare.git")), "bare");
    }
}
