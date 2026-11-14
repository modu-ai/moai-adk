pub mod agent_defs;
pub mod agent_install;
pub mod agents;
pub mod apply;
pub mod archive;
pub mod atlas_query;
pub mod atlas_tag;
pub mod audit;
pub mod audit_claim;
pub mod audit_clone;
pub mod audit_delta;
pub mod banner;
pub mod baseline;
pub mod bin_names;
pub mod claim;
pub mod claim_cache;
pub mod claim_timeout;
pub mod claim_verify;
pub mod claude;
pub mod clone_root;
pub mod close_check;
pub mod code_refs;
pub mod context;
pub mod contract;
pub mod decision;
pub mod decision_file;
pub mod decision_log;
pub mod dispatch;
pub mod error_next;
pub mod flake;
pub mod flake_record;
pub mod flow;
pub mod forge;
pub mod gate;
pub mod gate_collect;
pub mod gate_loop;
pub mod gate_merge;
pub mod gate_schedule;
pub mod generation;
pub mod github_templates;
pub mod go_hooks;
pub mod go_settings;
pub mod go_state;
pub mod goal;
pub mod goal_block;
pub mod handoff;
pub mod heavy_slot;
pub mod help;
pub mod home_root;
pub mod home_routing;
pub mod home_state;
pub mod home_style;
pub mod home_workdir;
pub mod hook_config;
pub mod hook_engine;
pub mod hook_guards;
pub mod hooks;
pub mod hosts;
pub mod hygiene;
pub mod identity;
pub mod index;
pub mod index_records;
pub mod install;
pub mod key_vault;
pub mod labels_check;
pub mod lane_duty;
pub mod lane_env;
pub mod lane_handoff;
pub mod lane_loop;
pub mod lane_next;
pub mod lane_registry;
pub mod lane_reset;
pub mod lane_status;
pub mod language;
pub mod launcher_init;
pub mod layout;
pub mod leader;
pub mod local_build;
pub mod message;
pub mod migrate;
pub mod ontology;
pub mod optimize;
pub mod parity;
pub mod plan;
pub mod pr_link;
pub mod profile;
pub mod progress;
pub mod project_detect;
pub mod project_skills;
pub mod queue;
pub mod queue_query;
pub mod recovery;
pub mod redact;
pub mod repo_scans;
pub mod review;
pub mod run;
pub mod scope;
pub mod search;
pub mod self_audit;
pub mod self_update;
pub mod session;
pub mod skills_legacy;
pub mod spec;
pub mod spec_links;
pub mod startable;
pub mod state_root;
pub mod statusline;
pub mod ui;
pub mod ui_summary;
pub mod ui_text;
pub mod update;
pub mod update_cleanup;
pub mod update_hooks;
pub mod update_tux;
pub mod usage_limit;
pub mod verify;
pub mod worktree;
pub mod worktree_gc;
pub mod worktree_include;
pub mod worktree_move;
pub mod worktree_report;
pub use agents::{agents_apply, agents_plan, AgentsPlan, AgentsRequest};
pub use baseline::tree_digest;
pub use profile::{apply, plan, read_profile, DecisionRecord, Delta, Plan, Profile};
pub use session::{launch as session_launch, SessionRequest};

use profile::read;

use serde_json::{json, Value};
use sha2::{Digest, Sha256};
use std::{
    fs::{self},
    path::{Path, PathBuf},
};

type Result<T> = std::result::Result<T, String>;
/// The version this binary reports: the package version, plus `+local.<n>` when the build set
/// `MOAI_LOCAL_BUILD=<n>` (#2722 AC-5; see [`local_build`]). Set by `build.rs`.
pub const BUILD_VERSION: &str = env!("MOAI_BUILD_VERSION");
/// What `--version` prints after the command name (#2790 R1): the build version and the
/// generation mark `(rust)`, so `moai --version` tells this build from the Go `moai`
/// (`moai version v3…`). Both binary names print the same line, `moai-cli <ver> (rust)`.
pub const VERSION_SIGNATURE: &str = concat!(env!("MOAI_BUILD_VERSION"), " (rust)");
/// The one spelling of the project state directory (Issue #2725 S1): `project_state!()` is the
/// directory, `project_state!("sub/path")` a path under it, both `&'static str` so constants
/// and tables can use them. Every project-tree path in `src` goes through this macro or
/// [`PROJECT_STATE_DIR`]; the move to `.moai/` (#2725 S2) then changes one place.
#[macro_export]
macro_rules! project_state {
    () => {
        ".moai-cli"
    };
    ($sub:literal) => {
        concat!(".moai-cli/", $sub)
    };
}
/// Paths that keep the `.moai-cli` name when the project state moves (#2725 decisions
/// 6854854683): the clone-shared state under the git common dir (D3), files under HOME (D4),
/// the lane worktrees until #2723 moves them (D6) and the earlier project root read as a
/// legacy location (D7).
#[macro_export]
macro_rules! fixed_state {
    () => {
        ".moai-cli"
    };
    ($sub:literal) => {
        concat!(".moai-cli/", $sub)
    };
}
/// Project state directory of the Rust RC. The Go generation keeps `.moai/`; both coexist
/// untouched until the user performs the final rename in a later step.
pub const PROJECT_STATE_DIR: &str = project_state!();
/// Clone-shared state under the git common dir: claims, goal, sessions, run baselines (#2725 D3).
pub const CLONE_STATE_DIR: &str = fixed_state!();
/// Lane worktrees of the primary checkout, relative to it; they stay here until #2723 (#2725 D6).
pub const WORKTREES_DIR: &str = fixed_state!("worktrees");
/// The earlier user-home state root (`~/.moai-cli`). Production code asks
/// [`home_root::dir`]/[`home_root::path`], which give `~/.moai` once its marker is there (#2696 C4b).
pub const HOME_STATE_DIR: &str = ".moai-cli";

pub fn digest(bytes: &[u8]) -> String {
    format!("sha256:{:x}", Sha256::digest(bytes))
}
/// Credential-shaped substrings of the conservative profile filter, lower-cased. Shared with
/// the common redaction module (`crate::redact`, 03-B #1996) so both refuse the same shapes;
/// this is not a general secret detector.
pub(crate) const SECRET_SHAPES: [&str; 7] = [
    "sk-",
    "ghp_",
    "github_pat_",
    "-----begin",
    "password=",
    "token=",
    "api_key=",
];
// Digest v1: sorted UTF-8 relative file paths and bytes, each framed by a u64 big-endian length.
// Symlinks and special files are rejected. This is content integrity, not publisher trust.
pub fn package_digest(root: &Path) -> Result<String> {
    if fs::symlink_metadata(root)
        .map_err(|_| "package unavailable")?
        .file_type()
        .is_symlink()
    {
        return Err("package symlink refused".into());
    }
    fn walk(
        root: &Path,
        dir: &Path,
        files: &mut Vec<PathBuf>,
        depth: usize,
        entries: &mut usize,
    ) -> Result<()> {
        if depth > 32 {
            return Err("package nesting limit exceeded".into());
        }
        for entry in fs::read_dir(dir).map_err(|_| "package directory unavailable")? {
            *entries += 1;
            if *entries > 8192 {
                return Err("package entry limit exceeded".into());
            }
            let path = entry.map_err(|_| "package entry unavailable")?.path();
            let m = fs::symlink_metadata(&path).map_err(|_| "package metadata unavailable")?;
            if m.file_type().is_symlink() {
                return Err("package symlink refused".into());
            }
            if m.is_dir() {
                walk(root, &path, files, depth + 1, entries)?;
            } else if m.is_file() {
                files.push(
                    path.strip_prefix(root)
                        .map_err(|_| "package path invalid")?
                        .into(),
                );
            } else {
                return Err("package special file refused".into());
            }
            if files.len() > 4096 {
                return Err("package too large".into());
            }
        }
        Ok(())
    }
    let mut files = Vec::new();
    walk(root, root, &mut files, 0, &mut 0)?;
    files.sort();
    if !files.contains(&PathBuf::from("SKILL.md")) {
        return Err("package SKILL.md missing".into());
    }
    let mut h = Sha256::new();
    h.update(b"moai-package-tree-v1\0");
    let mut total = 0;
    for rel in files {
        let raw = rel.to_str().ok_or("non UTF-8 package path")?;
        if raw.contains('\\') {
            return Err("nonportable package path".into());
        }
        let name = raw.replace('\\', "/");
        let bytes = read(&root.join(&rel))?.ok_or("package file disappeared")?;
        total += bytes.len();
        if total > 16 * 1024 * 1024 {
            return Err("package too large".into());
        }
        h.update((name.len() as u64).to_be_bytes());
        h.update(name.as_bytes());
        h.update((bytes.len() as u64).to_be_bytes());
        h.update(bytes);
    }
    Ok(format!("sha256:{:x}", h.finalize()))
}

pub fn resolve(root: &Path, repo: Option<&Path>, user: Option<&Path>) -> Result<Value> {
    let (profile, _) = read_profile(root)?;
    let (source, path) = if profile.package_pin.is_some() {
        (
            "repo_pin",
            repo.ok_or("explicit pin requires repository package; fallback forbidden")?,
        )
    } else {
        ("user", user.ok_or("user package required without pin")?)
    };
    let actual = package_digest(path)?;
    if profile
        .package_pin
        .as_ref()
        .is_some_and(|pin| pin != &actual)
    {
        return Err("pinned package digest mismatch; fallback forbidden".into());
    }
    Ok(
        json!({"source":source,"digest":actual,"digest_algorithm":"moai-package-tree-v1","host_loaded":false,"publisher_trust":"not_verified"}),
    )
}

#[cfg(test)]
mod state_dir_tests {
    /// Code lines outside this file that spell a `.moai-cli` path as a string literal.
    /// Comments are skipped; a literal counts when it is `".moai-cli"` or starts with
    /// `".moai-cli/` (`.moai-cli-synthetic-home` is another name).
    fn spelled_outside_lib(text: &str) -> Vec<String> {
        text.lines()
            .filter(|l| {
                let t = l.trim_start();
                !(t.starts_with("//") || t.starts_with('*') || t.starts_with("/*"))
            })
            .filter(|l| l.contains("\".moai-cli\"") || l.contains("\".moai-cli/"))
            .map(str::to_string)
            .collect()
    }

    #[test]
    fn project_state_dir_is_spelled_once() {
        // #2725 S1: every `.moai-cli` path in src goes through `project_state!`, `fixed_state!`
        // or the constants built from them; only src/lib.rs spells the name.
        let root = std::path::Path::new(env!("CARGO_MANIFEST_DIR")).join("src");
        let lib = root.join("lib.rs");
        let mut stack = vec![root];
        let mut found = Vec::new();
        let mut files = 0;
        while let Some(dir) = stack.pop() {
            for entry in std::fs::read_dir(&dir).unwrap() {
                let path = entry.unwrap().path();
                if path.is_dir() {
                    stack.push(path);
                } else if path.extension().is_some_and(|e| e == "rs") && path != lib {
                    files += 1;
                    let text = std::fs::read_to_string(&path).unwrap();
                    for line in spelled_outside_lib(&text) {
                        found.push(format!("{}: {}", path.display(), line.trim()));
                    }
                }
            }
        }
        assert!(files > 100, "the walk saw {files} files; src moved?");
        assert!(
            found.is_empty(),
            "spell these through project_state!/fixed_state!:\n{}",
            found.join("\n")
        );
    }

    #[test]
    fn the_spelling_check_sees_literals_and_skips_comments() {
        let cases = [
            ("let p = root.join(\".moai-cli\");", 1),
            ("const M: &str = \".moai-cli/manifests/x.json\";", 1),
            ("    // the old \".moai-cli/receipts\" folder", 0),
            ("/// see `.moai-cli/receipts`", 0),
            (" * \".moai-cli/x\"", 0),
            ("const S: &str = \".moai-cli-synthetic-home\";", 0),
            ("format!(\"{}/x\", crate::PROJECT_STATE_DIR)", 0),
        ];
        for (line, want) in cases {
            assert_eq!(spelled_outside_lib(line).len(), want, "{line}");
        }
        assert_eq!(
            crate::project_state!("manifests/a.json"),
            ".moai-cli/manifests/a.json"
        );
        assert_eq!(crate::PROJECT_STATE_DIR, ".moai-cli");
        assert_eq!(crate::CLONE_STATE_DIR, ".moai-cli");
        assert_eq!(crate::WORKTREES_DIR, ".moai-cli/worktrees");
    }

    /// Production lines (before the file's first test module) that name the earlier project
    /// root directly instead of asking `state_root`.
    fn earlier_root_uses(text: &str) -> usize {
        text.lines()
            .take_while(|l| !l.starts_with("#[cfg(test)]") && !l.starts_with("#[cfg(all(test"))
            .filter(|l| {
                let t = l.trim_start();
                !(t.starts_with("//") || t.starts_with('*'))
            })
            .filter(|l| l.contains("PROJECT_STATE_DIR") || l.contains("project_state!"))
            .count()
    }

    #[test]
    fn project_state_paths_are_resolved_per_project() {
        // #2725 S2b: production code asks `state_root::{dir, path, rel}` for the project's
        // state folder. These files keep the earlier root on purpose, each for its reason.
        let allowed: &[(&str, usize, &str)] = &[
            (
                "apply.rs",
                1,
                "the earlier receipt store stays out of recovery too",
            ),
            (
                "forge/report.rs",
                1,
                "`render` without a project keeps the earlier spelling",
            ),
            (
                "go_state.rs",
                1,
                "the Go move carries the earlier root's product state into .moai/ (#2696)",
            ),
            (
                "migrate.rs",
                2,
                "HOME migration rehearsal targets (#2459 HOLD)",
            ),
            (
                "project_detect.rs",
                2,
                "the code walk skips the earlier root",
            ),
            (
                "scope.rs",
                1,
                "the shared-interface glob (REQ-06C-06, pinned by worktree_cli)",
            ),
        ];
        let root = std::path::Path::new(env!("CARGO_MANIFEST_DIR")).join("src");
        let mut stack = vec![root.clone()];
        let mut found = std::collections::BTreeMap::new();
        while let Some(dir) = stack.pop() {
            for entry in std::fs::read_dir(&dir).unwrap() {
                let path = entry.unwrap().path();
                if path.is_dir() {
                    stack.push(path);
                } else if path.extension().is_some_and(|e| e == "rs") {
                    let rel = path
                        .strip_prefix(&root)
                        .unwrap()
                        .to_string_lossy()
                        .replace('\\', "/");
                    if rel == "lib.rs" || rel == "state_root.rs" {
                        continue;
                    }
                    let n = earlier_root_uses(&std::fs::read_to_string(&path).unwrap());
                    if n > 0 {
                        found.insert(rel, n);
                    }
                }
            }
        }
        let want: std::collections::BTreeMap<String, usize> = allowed
            .iter()
            .map(|(f, n, _)| (f.to_string(), *n))
            .collect();
        assert_eq!(
            found, want,
            "ask crate::state_root::{{dir, path, rel}} instead"
        );
    }
}
