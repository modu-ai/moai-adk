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
pub mod claim;
pub mod claim_cache;
pub mod claim_verify;
pub mod claude;
pub mod code_refs;
pub mod context;
pub mod contract;
pub mod decision;
pub mod decision_file;
pub mod decision_log;
pub mod dispatch;
pub mod flow;
pub mod forge;
pub mod gate;
pub mod generation;
pub mod github_templates;
pub mod go_state;
pub mod handoff;
pub mod home_routing;
pub mod hook_config;
pub mod hook_engine;
pub mod hook_guards;
pub mod hooks;
pub mod hosts;
pub mod identity;
pub mod index;
pub mod index_records;
pub mod install;
pub mod key_vault;
pub mod lane_handoff;
pub mod lane_status;
pub mod language;
pub mod launcher_init;
pub mod layout;
pub mod leader;
pub mod message;
pub mod migrate;
pub mod ontology;
pub mod parity;
pub mod plan;
pub mod pr_link;
pub mod profile;
pub mod project_detect;
pub mod project_skills;
pub mod queue;
pub mod queue_query;
pub mod recovery;
pub mod redact;
pub mod review;
pub mod run;
pub mod scope;
pub mod search;
pub mod self_update;
pub mod session;
pub mod spec;
pub mod spec_links;
pub mod statusline;
pub mod update;
pub mod verify;
pub mod worktree;
pub mod worktree_gc;
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
/// Project state directory of the Rust RC. The Go generation keeps `.moai/`; both coexist
/// untouched until the user performs the final rename in a later step.
pub const PROJECT_STATE_DIR: &str = ".moai-cli";
/// User-home install root of the Rust RC (`~/.moai-cli`). The Go generation keeps `~/.moai`.
/// No command writes there yet; the constant fixes the destination for later install code.
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
