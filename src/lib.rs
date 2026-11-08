pub mod agents;
pub mod apply;
pub mod baseline;
pub mod claude;
pub mod context;
pub mod contract;
pub mod decision;
pub mod forge;
pub mod hosts;
pub mod identity;
pub mod install;
pub mod plan;
pub mod run;
pub mod session;
pub mod spec;
pub mod verify;
pub use agents::{agents_apply, agents_plan, AgentsPlan, AgentsRequest};
pub use baseline::tree_digest;
pub use session::{launch as session_launch, SessionRequest};

use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use sha2::{Digest, Sha256};
use std::{
    fs::{self},
    path::{Path, PathBuf},
};

type Result<T> = std::result::Result<T, String>;
const LIMIT: u64 = 65536;
/// Project state directory of the Rust RC. The Go generation keeps `.moai/`; both coexist
/// untouched until the user performs the final rename in a later step.
pub const PROJECT_STATE_DIR: &str = ".moai-cli";
/// User-home install root of the Rust RC (`~/.moai-cli`). The Go generation keeps `~/.moai`.
/// No command writes there yet; the constant fixes the destination for later install code.
pub const HOME_STATE_DIR: &str = ".moai-cli";
#[derive(Clone, Debug, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct Profile {
    pub schema_version: u32,
    pub project_id: String,
    pub purpose: String,
    pub tools: Vec<String>,
    pub development_methods: Vec<String>,
    pub hosts: Vec<String>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub package_pin: Option<String>,
}
#[derive(Clone, Debug, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct Plan {
    pub protocol_version: u32,
    pub profile: Profile,
    pub expected_sha256: Option<String>,
    pub target_binding: String,
    pub action: String,
    pub plan_digest: String,
}
pub fn digest(bytes: &[u8]) -> String {
    format!("sha256:{:x}", Sha256::digest(bytes))
}
fn valid_digest(s: &str) -> bool {
    s.len() == 71
        && s.starts_with("sha256:")
        && s[7..]
            .bytes()
            .all(|b| b.is_ascii_digit() || (b'a'..=b'f').contains(&b))
}
/// Credential-shaped substrings of the conservative profile filter, lower-cased. Shared with
/// the 07-B promotion gate (`forge::promote`) so both refuse the same shapes; this is not a
/// general secret detector (the common redaction module belongs to 03-B #1996).
pub(crate) const SECRET_SHAPES: [&str; 7] = [
    "sk-",
    "ghp_",
    "github_pat_",
    "-----begin",
    "password=",
    "token=",
    "api_key=",
];
fn portable(s: &str) -> bool {
    let lower = s.to_ascii_lowercase();
    !s.trim().is_empty()
        && s.len() <= 2048
        && !s.chars().any(char::is_control)
        && !["/", "\\", "~/", ":\\"]
            .iter()
            .chain(SECRET_SHAPES.iter())
            .any(|x| lower.contains(x))
}
fn validate(p: &Profile) -> Result<()> {
    let id_ok = !p.project_id.is_empty()
        && p.project_id.len() <= 80
        && p.project_id
            .bytes()
            .all(|b| b.is_ascii_alphanumeric() || b == b'-' || b == b'_');
    if p.schema_version != 1
        || !id_ok
        || !portable(&p.purpose)
        || p.tools.is_empty()
        || p.development_methods.is_empty()
        || p.hosts.is_empty()
        || p.tools.len() > 32
        || p.development_methods.len() > 32
        || p.hosts.len() > 2
        || !p
            .tools
            .iter()
            .chain(&p.development_methods)
            .all(|s| portable(s))
        || !p.hosts.iter().all(|s| s == "claude" || s == "codex")
        || p.package_pin.as_ref().is_some_and(|s| !valid_digest(s))
    {
        return Err("invalid or nonportable profile; no values were logged".into());
    }
    Ok(())
}
fn target(root: &Path) -> Result<PathBuf> {
    let root = root
        .canonicalize()
        .map_err(|_| "project directory unavailable")?;
    if !root.is_dir() {
        return Err("project must be a directory".into());
    }
    let dir = root.join(PROJECT_STATE_DIR);
    for path in [&dir, &dir.join("project.yaml")] {
        match fs::symlink_metadata(path) {
            Ok(meta) if meta.file_type().is_symlink() => {
                return Err("symlink target refused".into())
            }
            Ok(_) => (),
            Err(e) if e.kind() == std::io::ErrorKind::NotFound => (),
            Err(_) => return Err("target metadata unavailable".into()),
        }
    }
    Ok(dir.join("project.yaml"))
}
fn read(path: &Path) -> Result<Option<Vec<u8>>> {
    match fs::metadata(path) {
        Ok(m) if !m.is_file() || m.len() > LIMIT => {
            Err("target is not a bounded regular file".into())
        }
        Ok(_) => fs::read(path)
            .map(Some)
            .map_err(|_| "target read failed".into()),
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => Ok(None),
        Err(_) => Err("target metadata unavailable".into()),
    }
}
fn plan_digest(p: &Plan) -> String {
    digest(
        &serde_json::to_vec(&(
            p.protocol_version,
            &p.profile,
            &p.expected_sha256,
            &p.target_binding,
            &p.action,
        ))
        .expect("typed plan"),
    )
}
pub fn plan(root: &Path, profile: Profile) -> Result<Plan> {
    validate(&profile)?;
    let old = read(&target(root)?)?;
    let action = match &old {
        None => "create",
        Some(bytes) => {
            // JSON is a YAML 1.2 subset. General YAML and comments are preserved by refusing edits.
            let current: Value = serde_json::from_slice(bytes)
                .map_err(|_| "existing YAML requires preservation review")?;
            let desired = serde_json::to_value(&profile).expect("typed profile");
            if !desired
                .as_object()
                .expect("profile object")
                .iter()
                .all(|(k, v)| current.get(k) == Some(v))
                || (profile.package_pin.is_none() && current.get("package_pin").is_some())
            {
                return Err(
                    "existing profile differs; preserve user edits and review merge".into(),
                );
            }
            "unchanged"
        }
    };
    let binding = digest(
        target(root)?
            .to_str()
            .ok_or("non UTF-8 project path")?
            .as_bytes(),
    );
    let mut p = Plan {
        protocol_version: 1,
        profile,
        target_binding: binding,
        expected_sha256: old.as_deref().map(digest),
        action: action.into(),
        plan_digest: String::new(),
    };
    p.plan_digest = plan_digest(&p);
    Ok(p)
}
pub fn apply(root: &Path, p: Plan, confirmed: &str, permission: &str) -> Result<Value> {
    // Caller authorization remains host-owned; this reference is not an authorization verifier.
    if ["", "null", "placeholder"].contains(&permission.trim())
        || permission.len() > 128
        || permission.chars().any(char::is_control)
    {
        return Err("permission reference required".into());
    }
    let current_plan = plan(root, p.profile.clone())?;
    if p.protocol_version != 1
        || confirmed != p.plan_digest
        || plan_digest(&p) != p.plan_digest
        || current_plan.plan_digest != p.plan_digest
    {
        return Err("stale or altered plan; re-preview required".into());
    }
    let path = target(root)?;
    let request_id = apply::new_request_id("project.apply", &p.plan_digest);
    let mut receipt = apply::Receipt::new(&request_id, "project.apply", root);
    if p.action == "unchanged" {
        // REQ-01B-10: a re-run writes nothing; the receipt records zero publishes.
        receipt.status = "complete".into();
        receipt
            .save(root)
            .map_err(|e| format!("receipt not recorded: {e}"))?;
        return Ok(json!({
            "state":"unchanged","host_verified":false,"receipt":request_id
        }));
    }
    fs::create_dir_all(path.parent().expect("target parent"))
        .map_err(|_| "project state directory unavailable")?;
    let path = target(root)?;
    // Exclusive create through the apply engine (HOLD-B2): stage in a unique temp
    // sibling and publish via hard link, so a crash mid-write never leaves a partial
    // profile file behind.
    let mut bytes = serde_json::to_vec_pretty(&p.profile).expect("typed profile");
    bytes.push(b'\n');
    apply::publish_exclusive(&path, &bytes)?;
    receipt.entry(
        &format!("{PROJECT_STATE_DIR}/project.yaml"),
        "create",
        None,
        Some(&digest(&bytes)),
        true,
        None,
    );
    receipt.status = "complete".into();
    receipt.save(root).map_err(|e| {
        format!("published {PROJECT_STATE_DIR}/project.yaml but its receipt was not recorded: {e}")
    })?;
    Ok(json!({
        "state":"created","profile_digest":digest(&bytes),"host_verified":false,"receipt":request_id
    }))
}
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
/// Reads the existing profile, keeping only known keys. Unknown user keys are left in the
/// file untouched; the raw bytes are returned so callers can bind to the exact revision.
pub fn read_profile(root: &Path) -> Result<(Profile, Vec<u8>)> {
    let bytes = read(&target(root)?)?.ok_or("project profile missing")?;
    let value: Value =
        serde_json::from_slice(&bytes).map_err(|_| "existing YAML requires preservation review")?;
    let map = value.as_object().ok_or("invalid profile object")?;
    if map.get("package_pin").is_some_and(Value::is_null) {
        return Err("explicit null pin is invalid".into());
    }
    let known = [
        "schema_version",
        "project_id",
        "purpose",
        "tools",
        "development_methods",
        "hosts",
        "package_pin",
    ];
    let filtered: serde_json::Map<String, Value> = map
        .iter()
        .filter(|(k, _)| known.contains(&k.as_str()))
        .map(|(k, v)| (k.clone(), v.clone()))
        .collect();
    let profile: Profile =
        serde_json::from_value(Value::Object(filtered)).map_err(|_| "invalid profile")?;
    validate(&profile)?;
    Ok((profile, bytes))
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
