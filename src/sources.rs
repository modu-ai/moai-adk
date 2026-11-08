//! External skill source candidates, provenance manifest and read receipts (Issue
//! #1996, 03-B).
//!
//! `skill-source-plan` judges typed candidates read-only: license evidence (a
//! `LICENSE*`/`COPYING*` file or a SKILL.md frontmatter `license:` string) decides
//! `reviewed` versus `refused: license_unverified` — a check that evidence exists,
//! never a legal judgment. `skill-source-apply` copies approved fixed-revision trees
//! into the host official skill folders with exclusive creates, records the
//! `moai-package-tree-v1` digest of every installed tree in
//! `.moai-cli/manifests/sources.json` (structured merge on
//! `source_url`+`resolved_revision`, unknown keys preserved, byte-identical rerun)
//! and reports installation separately from host discovery and execution
//! (`host_loaded:false`, `execution_verified:false`). `skill-read-receipt` walks an
//! installed tree and records that every file was actually read.
//!
//! RC1 fetch boundary (HOLD-1 first pass): sources are caller-provided local
//! checkouts; moai-cli computes digests and copies, and never runs git, a network
//! client or the Vercel skills CLI. A remote `source_type` therefore plans as
//! `requires_local_checkout`. `publisher_trust` stays `not_verified` until in-toto,
//! SLSA or Sigstore evidence is actually verified (REQ-03B-06) — unverified evidence
//! is never recorded.

use crate::{digest, package_digest, read as read_bounded, PROJECT_STATE_DIR};
use serde::{Deserialize, Serialize};
use serde_json::{json, Map, Value};
use std::{
    fs,
    io::Write as _,
    path::{Path, PathBuf},
};

type Result<T> = std::result::Result<T, String>;
/// Disambiguates temporary names inside one process: two calls in the same clock tick
/// must not collide on the same staging name.
fn next_temp_seq() -> u64 {
    use std::sync::atomic::{AtomicU64, Ordering};
    static SEQ: AtomicU64 = AtomicU64::new(0);
    SEQ.fetch_add(1, Ordering::Relaxed)
}
/// Same per-file bound as the profile and package readers.
const FILE_LIMIT: u64 = 65536;
/// Installed trees stay within the package digest's own bounds; the walk refuses
/// beyond them instead of silently hashing a huge tree.
const RECEIPT_TOTAL_LIMIT: u64 = 16 * 1024 * 1024;
const MAX_SKILLS_PER_SOURCE: usize = 8;
const MAX_CANDIDATES: usize = 32;

/// Skill names follow the Agent Skills specification (adopted 2026-11-08): 1-64
/// characters, lowercase letters, digits and hyphens, no leading/trailing/consecutive
/// hyphen, matching the directory name.
fn valid_skill_name(s: &str) -> bool {
    !s.is_empty()
        && s.len() <= 64
        && s.bytes()
            .all(|b| b.is_ascii_lowercase() || b.is_ascii_digit() || b == b'-')
        && !s.starts_with('-')
        && !s.ends_with('-')
        && !s.contains("--")
}

fn valid_revision(s: &str) -> bool {
    s.len() == 40
        && s.bytes()
            .all(|b| b.is_ascii_digit() || (b'a'..=b'f').contains(&b))
}

fn valid_profile_digest(s: &str) -> bool {
    let hex = s.strip_prefix("sha256:").unwrap_or(s);
    s.starts_with("sha256:")
        && hex.len() == 64
        && hex
            .bytes()
            .all(|b| b.is_ascii_digit() || (b'a'..=b'f').contains(&b))
}

fn bounded_url(s: &str) -> bool {
    !s.trim().is_empty() && s.len() <= 2048 && !s.chars().any(char::is_control)
}

/// Search terms travel to external tooling one day, so they carry the portable
/// filter: no paths, no credential shapes, no control characters (REQ-03B-08).
fn portable_search_term(s: &str) -> bool {
    let lower = s.to_ascii_lowercase();
    !s.trim().is_empty()
        && s.len() <= 256
        && !s.chars().any(char::is_control)
        && !["/", "\\", "~/", ":\\"]
            .iter()
            .chain(crate::SECRET_SHAPES.iter())
            .any(|x| lower.contains(x))
}

#[derive(Clone, Debug, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct SourceCandidate {
    /// Local path for `local` sources (absolute, or relative to the project root);
    /// recorded origin for the remote types.
    pub source_url: String,
    /// github | gitlab | git | local | download
    pub source_type: String,
    /// 40-character commit SHA the caller pinned, or null (REQ-03B-01).
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub requested_revision: Option<String>,
    pub skill_names: Vec<String>,
    /// claude | codex
    pub host_targets: Vec<String>,
    /// user | project
    pub scope: String,
}

#[derive(Clone, Debug, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct SourcesRequest {
    pub protocol_version: u32,
    /// Public technology/feature words the caller would search with (REQ-03B-08).
    #[serde(default)]
    pub search_terms: Vec<String>,
    /// `sha256:` digest of the 03-A profile the candidates were selected under; the
    /// provenance link recorded with every entry.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub profile_digest: Option<String>,
    #[serde(default)]
    pub candidates: Vec<SourceCandidate>,
}

/// One candidate's read-only judgment (REQ-03B-01).
#[derive(Clone, Debug, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct CandidatePlan {
    pub source_url: String,
    pub source_type: String,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub requested_revision: Option<String>,
    /// The caller's pin, passed through under HOLD-1 (a): moai-cli never runs git to
    /// resolve it, so the trust boundary stays with the caller.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub resolved_revision: Option<String>,
    pub skill_names: Vec<String>,
    pub host_targets: Vec<String>,
    pub scope: String,
    /// License evidence: `status` reviewed|refused, `spdx` preserves the SKILL.md
    /// frontmatter string verbatim, `evidence` names the license file found.
    pub license: LicenseJudgment,
    /// copy | requires_local_checkout
    pub install_mode: String,
    pub telemetry_policy: String,
    /// install | refused: license_unverified | refused: source_unreadable
    pub decision: String,
    /// Digest of the first resolvable skill tree (representative for a multi-skill
    /// source; every tree is re-digested at apply time).
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub tree_digest: Option<String>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub detail: Option<String>,
}

#[derive(Clone, Debug, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct LicenseJudgment {
    /// reviewed | refused (`unknown` stays in the 00-D vocabulary, but the license
    /// gate maps a missing-evidence candidate straight to refused, REQ-03B-02).
    pub status: String,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub spdx: Option<String>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub evidence: Option<String>,
}

#[derive(Clone, Debug, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct SourcePlan {
    pub protocol_version: u32,
    pub request: SourcesRequest,
    pub target_binding: String,
    pub results: Vec<CandidatePlan>,
    pub plan_digest: String,
}

fn validate(r: &SourcesRequest) -> Result<()> {
    if r.protocol_version != 1 || r.candidates.len() > MAX_CANDIDATES {
        return Err("invalid sources request; no values were logged".into());
    }
    for term in &r.search_terms {
        if !portable_search_term(term) {
            return Err("invalid search terms; no values were logged".into());
        }
    }
    if let Some(d) = &r.profile_digest {
        if !valid_profile_digest(d) {
            return Err("invalid sources request; no values were logged".into());
        }
    }
    for c in &r.candidates {
        if !bounded_url(&c.source_url)
            || !["github", "gitlab", "git", "local", "download"].contains(&c.source_type.as_str())
            || c.skill_names.is_empty()
            || c.skill_names.len() > MAX_SKILLS_PER_SOURCE
            || !c.skill_names.iter().all(|n| valid_skill_name(n))
            || c.host_targets.is_empty()
            || !c.host_targets.iter().all(|h| h == "claude" || h == "codex")
            || !["user", "project"].contains(&c.scope.as_str())
            || c.requested_revision
                .as_ref()
                .is_some_and(|rev| !valid_revision(rev))
        {
            return Err("invalid sources request; no values were logged".into());
        }
    }
    Ok(())
}

/// The telemetry policy of the surrounding environment, recorded verbatim (REQ-03B-07
/// records the state; HOLD-3 owns the default policy decision).
fn telemetry_policy() -> String {
    let disabled = std::env::var_os("DISABLE_TELEMETRY").is_some()
        || std::env::var_os("DO_NOT_TRACK").is_some();
    if disabled {
        "telemetry_off; security-audit requests are disabled with it".into()
    } else {
        "not_configured; Vercel CLI telemetry default is HOLD-3".into()
    }
}

/// The skill directory inside a source checkout: `<dir>/<name>/SKILL.md`, or the
/// directory itself when it holds the SKILL.md and exactly one skill was requested.
fn resolve_skill_dir(source_dir: &Path, name: &str, single: bool) -> Option<PathBuf> {
    let nested = source_dir.join(name).join("SKILL.md");
    if nested.is_file() {
        return Some(source_dir.join(name));
    }
    if single && source_dir.join("SKILL.md").is_file() {
        return Some(source_dir.to_path_buf());
    }
    None
}

/// License evidence inside a skill directory: a `LICENSE*`/`COPYING*` file, and a
/// SKILL.md frontmatter `license:` string preserved verbatim when present (never a
/// legal judgment — both kinds of evidence are recorded independently).
fn license_judgment(skill_dir: &Path) -> std::result::Result<LicenseJudgment, ()> {
    let mut evidence = None;
    let entries = fs::read_dir(skill_dir).map_err(|_| ())?;
    for entry in entries.flatten() {
        let name = entry.file_name();
        let name = name.to_str().ok_or(())?;
        let is_file = entry.file_type().map_err(|_| ())?.is_file();
        if is_file
            && (name == "LICENSE" || name.starts_with("LICENSE.") || name.starts_with("COPYING"))
        {
            evidence = Some(name.to_string());
            break;
        }
    }
    let mut spdx = None;
    let skill_md = fs::read_to_string(skill_dir.join("SKILL.md")).map_err(|_| ())?;
    let mut lines = skill_md.lines();
    if lines.next().is_some_and(|l| l.trim() == "---") {
        for line in lines {
            let trimmed = line.trim();
            if trimmed == "---" {
                break;
            }
            if let Some(rest) = trimmed.strip_prefix("license:") {
                let value = rest.trim().trim_matches('"').trim_matches('\'');
                if !value.is_empty() && value.len() <= 128 {
                    spdx = Some(value.to_string());
                    break;
                }
            }
        }
    }
    let evidence = match (evidence, spdx.as_ref()) {
        (Some(file), _) => Some(file),
        (None, Some(_)) => Some("SKILL.md frontmatter license".to_string()),
        (None, None) => None,
    };
    Ok(match evidence {
        Some(file) => LicenseJudgment {
            status: "reviewed".into(),
            spdx,
            evidence: Some(file),
        },
        None => LicenseJudgment {
            status: "refused".into(),
            spdx: None,
            evidence: None,
        },
    })
}

fn sources_manifest_path(root: &Path) -> Result<PathBuf> {
    let root = root
        .canonicalize()
        .map_err(|_| "project directory unavailable")?;
    if !root.is_dir() {
        return Err("project must be a directory".into());
    }
    let dir = root.join(PROJECT_STATE_DIR);
    let full = dir.join("manifests/sources.json");
    for path in [&dir, &full] {
        match fs::symlink_metadata(path) {
            Ok(meta) if meta.file_type().is_symlink() => {
                return Err("symlink target refused".into())
            }
            Ok(_) => (),
            Err(e) if e.kind() == std::io::ErrorKind::NotFound => (),
            Err(_) => return Err("target metadata unavailable".into()),
        }
    }
    Ok(full)
}

fn plan_digest(p: &SourcePlan) -> String {
    digest(
        &serde_json::to_vec(&(
            p.protocol_version,
            &p.request,
            &p.target_binding,
            &p.results,
        ))
        .expect("typed plan"),
    )
}

/// `skill-source-plan`: judges candidates and writes nothing (REQ-03B-01).
pub fn plan(root: &Path, request: SourcesRequest) -> Result<SourcePlan> {
    validate(&request)?;
    let root = root
        .canonicalize()
        .map_err(|_| "project directory unavailable")?;
    let manifest = sources_manifest_path(&root)?;
    let binding = digest(
        manifest
            .to_str()
            .ok_or("non UTF-8 project path")?
            .as_bytes(),
    );
    let telemetry = telemetry_policy();
    let mut results = Vec::new();
    for c in &request.candidates {
        let single = c.skill_names.len() == 1;
        let mut plan = CandidatePlan {
            source_url: c.source_url.clone(),
            source_type: c.source_type.clone(),
            requested_revision: c.requested_revision.clone(),
            resolved_revision: c.requested_revision.clone(),
            skill_names: c.skill_names.clone(),
            host_targets: c.host_targets.clone(),
            scope: c.scope.clone(),
            license: LicenseJudgment {
                status: "refused".into(),
                spdx: None,
                evidence: None,
            },
            install_mode: "requires_local_checkout".into(),
            telemetry_policy: telemetry.clone(),
            decision: "refused: source_unreadable".into(),
            tree_digest: None,
            detail: None,
        };
        if c.source_type != "local" {
            plan.detail = Some(
                "remote fetch is out of the RC1 boundary; pass a caller-provided local checkout"
                    .into(),
            );
            results.push(plan);
            continue;
        }
        let source_dir = resolve_source_dir(&root, &c.source_url);
        let mut refused = false;
        for name in &c.skill_names {
            let Some(skill_dir) = resolve_skill_dir(&source_dir, name, single) else {
                plan.detail = Some("skill directory with SKILL.md not found".into());
                refused = true;
                break;
            };
            match license_judgment(&skill_dir) {
                Ok(judgment) if judgment.status == "reviewed" => {
                    if plan.license.evidence.is_none() {
                        plan.license = judgment;
                    }
                }
                Ok(judgment) => {
                    plan.license = judgment;
                    // Readable but unlicensed: REQ-03B-02's exact refusal.
                    plan.decision = "refused: license_unverified".into();
                    refused = true;
                    break;
                }
                Err(()) => {
                    plan.detail = Some("skill directory could not be read".into());
                    refused = true;
                    break;
                }
            }
            match package_digest(&skill_dir) {
                Ok(tree) => {
                    if plan.tree_digest.is_none() {
                        plan.tree_digest = Some(tree);
                    }
                }
                Err(e) => {
                    plan.detail = Some(e);
                    refused = true;
                    break;
                }
            }
        }
        if refused {
            results.push(plan);
            continue;
        }
        if plan.license.status != "reviewed" {
            // REQ-03B-02: no license evidence, no clone, install or 03-C reuse.
            plan.decision = "refused: license_unverified".into();
            results.push(plan);
            continue;
        }
        plan.install_mode = "copy".into();
        plan.decision = "install".into();
        results.push(plan);
    }
    let mut p = SourcePlan {
        protocol_version: 1,
        request,
        target_binding: binding,
        results,
        plan_digest: String::new(),
    };
    p.plan_digest = plan_digest(&p);
    Ok(p)
}

fn resolve_source_dir(root: &Path, source_url: &str) -> PathBuf {
    let raw = Path::new(source_url);
    if raw.is_absolute() {
        raw.to_path_buf()
    } else {
        root.join(raw)
    }
}

/// Copies one skill tree with the same refusals the digest applies, into a uniquely
/// named sibling directory, then publishes it with an exclusive rename. The rename
/// fails when the destination appeared meanwhile; an existing destination is reported
/// before anything is copied. Returns the number of files copied.
fn copy_skill_tree(src: &Path, dest: &Path) -> Result<usize> {
    package_digest(src)?;
    let parent = dest.parent().ok_or("target parent unavailable")?;
    fs::create_dir_all(parent).map_err(|_| "target directory unavailable")?;
    let nanos = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map(|d| d.as_nanos())
        .unwrap_or(0);
    let staging = parent.join(format!(
        ".moai-skill-{}-{nanos}-{}.tmp",
        std::process::id(),
        next_temp_seq()
    ));
    fs::create_dir(&staging).map_err(|_| "temporary directory refused")?;
    match copy_recursive(src, &staging, 0) {
        Ok(count) => match fs::rename(&staging, dest) {
            Ok(()) => Ok(count),
            Err(_) => {
                let _ = fs::remove_dir_all(&staging);
                Err("conflict in skill install: existing_skill".into())
            }
        },
        Err(e) => {
            let _ = fs::remove_dir_all(&staging);
            Err(e)
        }
    }
}

fn copy_recursive(src: &Path, dest: &Path, depth: usize) -> Result<usize> {
    if depth > 16 {
        return Err("skill tree nesting limit exceeded".into());
    }
    fs::create_dir_all(dest).map_err(|_| "target directory unavailable")?;
    let mut count = 0usize;
    for entry in fs::read_dir(src).map_err(|_| "source unreadable")? {
        let entry = entry.map_err(|_| "source unreadable")?;
        // symlink_metadata never follows the link: a symlinked skill file is refused,
        // not silently dereferenced into the installed tree.
        let meta = fs::symlink_metadata(entry.path()).map_err(|_| "source unreadable")?;
        let target = dest.join(entry.file_name());
        if meta.file_type().is_symlink() {
            return Err("package symlink refused".into());
        }
        if meta.is_dir() {
            count += copy_recursive(&entry.path(), &target, depth + 1)?;
        } else if meta.is_file() {
            if meta.len() > FILE_LIMIT {
                return Err("target is not a bounded regular file".into());
            }
            fs::copy(entry.path(), &target).map_err(|_| "target write failed")?;
            count += 1;
        } else {
            return Err("package special file refused".into());
        }
    }
    Ok(count)
}

fn host_skill_dir(home: &Path, root: &Path, host: &str, scope: &str, name: &str) -> PathBuf {
    let folder = match host {
        "codex" => ".agents/skills",
        _ => ".claude/skills",
    };
    match scope {
        "project" => root.join(folder).join(name),
        _ => home.join(folder).join(name),
    }
}

/// Merges one candidate's record into the manifest Value under its
/// `source_url`+`resolved_revision` key, preserving unknown user keys (REQ-03B-09).
fn merge_entry(
    manifest: &mut Value,
    profile_sha256: Option<&str>,
    result: &CandidatePlan,
    installed: &Value,
) {
    let Some(map) = manifest.as_object_mut() else {
        return;
    };
    let sources = map
        .entry("sources")
        .or_insert_with(|| json!([]))
        .as_array_mut()
        .expect("sources array");
    let record = json!({
        "source_url": result.source_url,
        "source_type": result.source_type,
        "requested_revision": result.requested_revision,
        "resolved_revision": result.resolved_revision,
        "skill_names": result.skill_names,
        "host_targets": result.host_targets,
        "scope": result.scope,
        "license": result.license,
        "install_mode": result.install_mode,
        "telemetry_policy": result.telemetry_policy,
        "decision": result.decision,
        "tree_digest": result.tree_digest,
        "profile_sha256": profile_sha256,
        "publisher_trust": "not_verified",
        "host_loaded": false,
        "execution_verified": false,
        "installed": installed,
    });
    let position = sources.iter().position(|s| {
        s.get("source_url").and_then(Value::as_str) == Some(result.source_url.as_str())
            && s.get("resolved_revision").and_then(Value::as_str)
                == result.resolved_revision.as_deref()
    });
    match position {
        Some(i) => {
            let mut merged = match sources[i].as_object() {
                Some(m) => m.clone(),
                None => Map::new(),
            };
            if let Some(new) = record.as_object() {
                for (k, v) in new {
                    merged.insert(k.clone(), v.clone());
                }
            }
            sources[i] = Value::Object(merged);
        }
        None => sources.push(record),
    }
}

fn write_replace_file(path: &Path, before: Option<&[u8]>, after: &[u8]) -> Result<()> {
    if read_bounded(path)?.as_deref() != before {
        return Err("target changed since plan; re-preview required".into());
    }
    let parent = path.parent().ok_or("target parent unavailable")?;
    fs::create_dir_all(parent).map_err(|_| "target directory unavailable")?;
    let nanos = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map(|d| d.as_nanos())
        .unwrap_or(0);
    let temp = parent.join(format!(
        ".moai-sources-{}-{nanos}-{}.tmp",
        std::process::id(),
        next_temp_seq()
    ));
    let mut file = fs::OpenOptions::new()
        .write(true)
        .create_new(true)
        .open(&temp)
        .map_err(|_| "temporary file refused")?;
    let written = file
        .write_all(after)
        .and_then(|_| file.sync_all())
        .map_err(|_| "temporary write failed".to_string());
    if written.is_err() {
        let _ = fs::remove_file(&temp);
        return Err("temporary write failed".into());
    }
    if before.is_some() {
        fs::rename(&temp, path).map_err(|_| "atomic replace failed")?;
    } else {
        match fs::hard_link(&temp, path) {
            Ok(()) => {
                let _ = fs::remove_file(&temp);
            }
            Err(_) => {
                let _ = fs::remove_file(&temp);
                return Err("exclusive create refused".into());
            }
        }
    }
    Ok(())
}

/// `skill-source-apply`: installs approved candidates and records provenance
/// (REQ-03B-03/04/09/10). All-or-nothing on destination conflicts: every destination
/// is checked before the first byte is copied. A candidate whose manifest record is
/// already byte-identical is skipped, so re-running the same input changes no bytes
/// (REQ-03B-09).
pub fn sources_apply(
    root: &Path,
    home: &Path,
    p: SourcePlan,
    confirmed: &str,
    permission: &str,
) -> Result<Value> {
    // Caller authorization remains host-owned; this reference is not an authorization verifier.
    if ["", "null", "placeholder"].contains(&permission.trim())
        || permission.len() > 128
        || permission.chars().any(char::is_control)
    {
        return Err("permission reference required".into());
    }
    let current = plan(root, p.request.clone())?;
    if p.protocol_version != 1
        || confirmed != p.plan_digest
        || plan_digest(&p) != p.plan_digest
        || current.plan_digest != p.plan_digest
    {
        return Err("stale or altered plan; re-preview required".into());
    }
    let root = root
        .canonicalize()
        .map_err(|_| "project directory unavailable")?;
    let manifest_path = sources_manifest_path(&root)?;
    let mut manifest: Value = match read_bounded(&manifest_path)? {
        Some(bytes) => serde_json::from_slice(&bytes)
            .map_err(|_| "sources manifest unreadable; inspect before applying".to_string())?,
        None => json!({"schema_version": 1, "sources": []}),
    };
    let profile_sha256 = p.request.profile_digest.clone();
    // Pre-render every record: an already-identical record means this apply was done
    // before, and nothing may change (REQ-03B-09).
    let mut rendered: Vec<(usize, &CandidatePlan, Value)> = Vec::new();
    let sources_list = manifest
        .get("sources")
        .and_then(Value::as_array)
        .cloned()
        .unwrap_or_default();
    for (idx, r) in p.results.iter().enumerate() {
        let mut record = json!({});
        merge_entry(&mut record, profile_sha256.as_deref(), r, &json!(null));
        let entry = record
            .get("sources")
            .and_then(Value::as_array)
            .and_then(|a| a.first())
            .cloned()
            .unwrap_or(Value::Null);
        // The skip compares what THIS apply would install — including the absolute
        // destination paths — so a different HOME (or scope) is never silently
        // absorbed by a previous run's record (REQ-03B-09).
        let expected_installed = if r.decision == "install" {
            let single = r.skill_names.len() == 1;
            let source_dir = resolve_source_dir(&root, &r.source_url);
            let mut arr = Vec::new();
            let mut resolvable = true;
            for name in &r.skill_names {
                let Some(skill_dir) = resolve_skill_dir(&source_dir, name, single) else {
                    resolvable = false;
                    break;
                };
                let Ok(tree) = package_digest(&skill_dir) else {
                    resolvable = false;
                    break;
                };
                for host in &r.host_targets {
                    let dest = host_skill_dir(home, &root, host, &r.scope, name);
                    arr.push(json!({
                        "skill_name": name,
                        "host": host,
                        "scope": r.scope,
                        "path": dest.to_str().unwrap_or(""),
                        "tree_digest": tree,
                        "read_complete": false,
                    }));
                }
            }
            if resolvable {
                Value::Array(arr)
            } else {
                Value::Null
            }
        } else {
            Value::Array(vec![])
        };
        let existing = sources_list.iter().find(|s| {
            s.get("source_url").and_then(Value::as_str) == Some(r.source_url.as_str())
                && s.get("resolved_revision").and_then(Value::as_str)
                    == r.resolved_revision.as_deref()
        });
        let done = existing.is_some_and(|s| {
            s.get("installed") == Some(&expected_installed)
                && s.get("decision") == entry.get("decision")
                && s.get("tree_digest") == entry.get("tree_digest")
                && s.get("license") == entry.get("license")
        });
        if !done {
            rendered.push((idx, r, record));
        }
    }
    // Destination pre-flight: every install target must be free before the first copy.
    for (_, r, _) in &rendered {
        if r.decision != "install" {
            continue;
        }
        let single = r.skill_names.len() == 1;
        let source_dir = resolve_source_dir(&root, &r.source_url);
        for name in &r.skill_names {
            if resolve_skill_dir(&source_dir, name, single).is_none() {
                return Err("stale or altered plan; re-preview required".into());
            }
            for host in &r.host_targets {
                let dest = host_skill_dir(home, &root, host, &r.scope, name);
                if fs::symlink_metadata(&dest).is_ok() {
                    return Err(format!(
                        "conflict in skill install: existing_skill at {name} ({host}/{}); nothing applied",
                        r.scope
                    ));
                }
            }
        }
    }
    let mut installed = Vec::new();
    let mut files_installed = 0usize;
    let mut applied_records = 0usize;
    for (idx, r, record) in &rendered {
        let mut per_source = Vec::new();
        if r.decision == "install" {
            let single = r.skill_names.len() == 1;
            let source_dir = resolve_source_dir(&root, &r.source_url);
            for name in &r.skill_names {
                let Some(skill_dir) = resolve_skill_dir(&source_dir, name, single) else {
                    return Err("stale or altered plan; re-preview required".into());
                };
                let tree = package_digest(&skill_dir)?;
                for host in &r.host_targets {
                    let dest = host_skill_dir(home, &root, host, &r.scope, name);
                    let count = copy_skill_tree(&skill_dir, &dest).map_err(|e| {
                        format!(
                            "partial apply: applied [{}]; failed at {name} ({host}): {e}",
                            installed.join(", ")
                        )
                    })?;
                    let installed_digest = package_digest(&dest)?;
                    if installed_digest != tree {
                        return Err(format!(
                            "partial apply: applied [{}]; installed tree digest mismatch at {name} ({host})",
                            installed.join(", ")
                        ));
                    }
                    files_installed += count;
                    per_source.push(json!({
                        "skill_name": name,
                        "host": host,
                        "scope": r.scope,
                        "path": dest.to_str().ok_or("non UTF-8 install path")?,
                        "tree_digest": installed_digest,
                        "read_complete": false,
                    }));
                    installed.push(format!("{name} ({host})"));
                }
            }
        }
        applied_records += 1;
        let _ = idx;
        merge_entry(
            &mut manifest,
            profile_sha256.as_deref(),
            r,
            &json!(per_source),
        );
        let _ = record;
    }
    if applied_records > 0 {
        if let Some(obj) = manifest.as_object_mut() {
            obj.insert(
                "profile_sha256".into(),
                profile_sha256
                    .clone()
                    .map(Value::String)
                    .unwrap_or(Value::Null),
            );
        }
        let mut out = serde_json::to_vec_pretty(&manifest).map_err(|_| "encode failed")?;
        out.push(b'\n');
        let before = read_bounded(&manifest_path)?;
        if before.as_deref() != Some(out.as_slice()) {
            write_replace_file(&manifest_path, before.as_deref(), &out)?;
        }
    }
    Ok(json!({
        "state": if installed.is_empty() { "unchanged" } else { "applied" },
        "files_installed": files_installed,
        "installed": installed,
        "read_complete": false,
        "host_loaded": false,
        "execution_verified": false,
    }))
}

/// `skill-read-receipt`: walks an installed tree and records that every file was
/// actually read (REQ-03B-05). A symlink, special file or unreadable file is a gap:
/// `read_complete:false` with its path. The receipt persists under the project's
/// `.moai-cli/manifests/read-receipts/<name>.json`.
pub fn read_receipt(root: &Path, dir: &Path, name: &str) -> Result<Value> {
    if !valid_skill_name(name) {
        return Err("invalid skill name; no values were logged".into());
    }
    let dir = dir
        .canonicalize()
        .map_err(|_| "install directory unavailable".to_string())?;
    if !dir.is_dir() {
        return Err("install directory unavailable".into());
    }
    let mut files = Vec::new();
    let mut unread = Vec::new();
    let mut total_bytes = 0u64;
    fn walk(
        dir: &Path,
        rel: &str,
        files: &mut Vec<Value>,
        unread: &mut Vec<String>,
        total: &mut u64,
        depth: usize,
    ) -> Result<()> {
        if depth > 16 {
            return Err("skill tree nesting limit exceeded".into());
        }
        for entry in fs::read_dir(dir).map_err(|_| "install directory unavailable".to_string())? {
            let entry = entry.map_err(|_| "install directory unavailable".to_string())?;
            let child_rel = if rel.is_empty() {
                entry.file_name().to_string_lossy().into_owned()
            } else {
                format!("{rel}/{}", entry.file_name().to_string_lossy())
            };
            let meta = match fs::symlink_metadata(entry.path()) {
                Ok(m) => m,
                Err(_) => {
                    unread.push(child_rel);
                    continue;
                }
            };
            if meta.file_type().is_symlink() {
                unread.push(format!("{child_rel} (symlink)"));
            } else if meta.is_dir() {
                walk(&entry.path(), &child_rel, files, unread, total, depth + 1)?;
            } else if meta.is_file() {
                if *total + meta.len() > RECEIPT_TOTAL_LIMIT {
                    return Err("installed tree exceeds the read-receipt bound".into());
                }
                match fs::read(entry.path()) {
                    Ok(bytes) => {
                        *total += bytes.len() as u64;
                        files.push(json!({"path": child_rel, "sha256": digest(&bytes)}));
                    }
                    Err(_) => unread.push(child_rel),
                }
            } else {
                unread.push(format!("{child_rel} (special file)"));
            }
        }
        Ok(())
    }
    walk(&dir, "", &mut files, &mut unread, &mut total_bytes, 0)?;
    let read_complete = unread.is_empty();
    let receipt = json!({
        "schema_version": 1,
        "skill_name": name,
        "read_complete": read_complete,
        "files": files,
        "total_bytes": total_bytes,
        "unread": unread,
        "host_loaded": false,
        "execution_verified": false,
    });
    let mut out = serde_json::to_vec_pretty(&receipt).map_err(|_| "encode failed")?;
    out.push(b'\n');
    let root = root
        .canonicalize()
        .map_err(|_| "project directory unavailable")?;
    let receipts_dir = root.join(PROJECT_STATE_DIR).join("manifests/read-receipts");
    fs::create_dir_all(&receipts_dir).map_err(|_| "project state directory unavailable")?;
    let receipt_path = receipts_dir.join(format!("{name}.json"));
    let before = read_bounded(&receipt_path)?;
    write_replace_file(&receipt_path, before.as_deref(), &out)?;
    Ok(json!({
        "skill_name": name,
        "read_complete": read_complete,
        "files_read": files.len(),
        "total_bytes": total_bytes,
        "unread": unread,
        "receipt_path": receipt_path.to_str().ok_or("non UTF-8 receipt path")?,
        "host_loaded": false,
        "execution_verified": false,
    }))
}

#[cfg(test)]
mod tests {
    use super::*;

    fn candidate() -> SourceCandidate {
        SourceCandidate {
            source_url: "/tmp/demo-skill".into(),
            source_type: "local".into(),
            requested_revision: None,
            skill_names: vec!["demo-skill".into()],
            host_targets: vec!["claude".into()],
            scope: "user".into(),
        }
    }

    #[test]
    fn skill_names_follow_the_agents_skills_spec() {
        assert!(valid_skill_name("demo-skill"));
        assert!(valid_skill_name("rust-review-2"));
        assert!(!valid_skill_name("Demo"));
        assert!(!valid_skill_name("-lead"));
        assert!(!valid_skill_name("a--b"));
        assert!(!valid_skill_name(""));
    }

    #[test]
    fn search_terms_reject_paths_and_secret_shapes() {
        assert!(portable_search_term("rust code review"));
        assert!(!portable_search_term("/Users/example/private"));
        assert!(!portable_search_term("token=abc"));
        assert!(!portable_search_term("C:\\tmp"));
    }

    #[test]
    fn request_validation_rejects_unknown_shapes() {
        let mut r = SourcesRequest {
            protocol_version: 1,
            search_terms: vec![],
            profile_digest: None,
            candidates: vec![candidate()],
        };
        assert!(validate(&r).is_ok());
        r.candidates[0].source_type = "ftp".into();
        assert!(validate(&r).is_err());
        r.candidates[0].source_type = "local".into();
        r.candidates[0].skill_names = vec!["Bad_Name".into()];
        assert!(validate(&r).is_err());
        r.candidates[0].skill_names = vec!["demo-skill".into()];
        r.profile_digest = Some("sha256:zz".into());
        assert!(validate(&r).is_err());
        r.profile_digest =
            Some("sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa".into());
        assert!(validate(&r).is_ok());
    }

    #[test]
    fn merge_preserves_unknown_user_keys() {
        let mut manifest = json!({
            "schema_version": 1,
            "sources": [{
                "source_url": "/tmp/demo-skill", "resolved_revision": null,
                "user_note": {"kept": true}
            }]
        });
        let plan = CandidatePlan {
            source_url: "/tmp/demo-skill".into(),
            source_type: "local".into(),
            requested_revision: None,
            resolved_revision: None,
            skill_names: vec!["demo-skill".into()],
            host_targets: vec!["claude".into()],
            scope: "user".into(),
            license: LicenseJudgment {
                status: "reviewed".into(),
                spdx: Some("MIT".into()),
                evidence: Some("LICENSE".into()),
            },
            install_mode: "copy".into(),
            telemetry_policy: "not_configured".into(),
            decision: "install".into(),
            tree_digest: Some("sha256:aa".into()),
            detail: None,
        };
        merge_entry(&mut manifest, None, &plan, &json!([]));
        let entry = &manifest["sources"][0];
        assert_eq!(entry["user_note"]["kept"], json!(true));
        assert_eq!(entry["publisher_trust"], json!("not_verified"));
        assert_eq!(entry["license"]["spdx"], json!("MIT"));
    }
}
