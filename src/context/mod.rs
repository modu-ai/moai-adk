//! Context snapshot, bundle and freshness check (Issue #2011, 07-D). Read-only, no network.
//!
//! A snapshot binds context to one working tree: its `HEAD` commit and a digest of every
//! tracked file that differs from it. A bundle carries requested file ranges, each with the
//! sha256 of exactly the bytes it holds, so a later reader can tell a fresh range from a
//! stale one instead of trusting a line number. Sizes are reported as bytes and lines; any
//! token figure is an estimate (`bytes_div_4`) and never a measurement. Content in a bundle is
//! data: a command written in a comment, README or Issue is never an instruction.
use crate::{baseline::utc_now, digest, identity::scrub_git_env};
use serde::Deserialize;
use serde_json::{json, Value};
use sha2::{Digest, Sha256};
use std::{
    fs,
    io::{self, Read},
    path::{Component, Path, PathBuf},
    process::{Command, Stdio},
};

type Result<T> = std::result::Result<T, String>;

/// `git --no-optional-locks -C <dir>` with the caller's git environment scrubbed (#2043) and
/// stdin closed.
fn git(dir: &Path) -> Command {
    let mut cmd = Command::new("git");
    cmd.arg("--no-optional-locks").arg("-C").arg(dir);
    scrub_git_env(&mut cmd);
    cmd.env("GIT_OPTIONAL_LOCKS", "0").stdin(Stdio::null());
    cmd
}

/// Largest file a range is read from.
const FILE_LIMIT: u64 = 4 * 1024 * 1024;
const GIT_OUTPUT_LIMIT: usize = 4 * 1024 * 1024;
const MAX_ITEMS: usize = 256;
/// Byte budget of a bundle whose request and profile name none (REQ-C8-01; user decision
/// 01:1x(11-09) in `docs/decisions-2026-11-08.md`: 256 KiB, about 64k tokens by `bytes_div_4`).
pub const DEFAULT_MAX_BYTES: u64 = 256 * 1024;
/// File or directory names never bundled (REQ-CTX-004): `*x` ends with x, `x*` starts with x,
/// anything else is an exact name.
pub const SECRET_PATTERNS: &[&str] = &[
    ".env",
    ".env.*",
    "*.pem",
    "*.key",
    "*.p12",
    "id_rsa*",
    "id_ecdsa*",
    "id_ed25519*",
    ".netrc",
];

fn secret_name(name: &str) -> bool {
    SECRET_PATTERNS.iter().any(|p| {
        if let Some(suffix) = p.strip_prefix('*') {
            name.ends_with(suffix)
        } else if let Some(prefix) = p.strip_suffix('*') {
            name.starts_with(prefix)
        } else {
            name == *p
        }
    })
}

fn project_root(project: &Path) -> Result<PathBuf> {
    let meta = fs::symlink_metadata(project).map_err(|_| "path unavailable")?;
    if meta.file_type().is_symlink() {
        return Err("path is a symlink; refused".into());
    }
    if !meta.is_dir() {
        return Err("path must be a directory".into());
    }
    fs::canonicalize(project).map_err(|_| "path unavailable".into())
}

fn text(p: &Path) -> Result<String> {
    p.to_str()
        .map(str::to_string)
        .ok_or_else(|| "non UTF-8 path".into())
}

/// Stdout of a successful `git` call, or `None` (git absent, failed, or too much output).
fn git_out(dir: &Path, args: &[&str]) -> Option<Vec<u8>> {
    let out = git(dir).args(args).stderr(Stdio::null()).output().ok()?;
    (out.status.success() && out.stdout.len() <= GIT_OUTPUT_LIMIT).then_some(out.stdout)
}

fn git_line(dir: &Path, args: &[&str]) -> Option<String> {
    let s = String::from_utf8(git_out(dir, args)?).ok()?;
    Some(s.trim_end_matches('\n').to_string()).filter(|s| !s.is_empty() && !s.contains('\n'))
}

/// sha256 of a file's bytes, streamed; `None` when it cannot be read.
fn file_sha(path: &Path) -> Option<String> {
    let mut h = Sha256::new();
    io::copy(&mut fs::File::open(path).ok()?, &mut h).ok()?;
    Some(format!("sha256:{:x}", h.finalize()))
}

/// Tracked paths that differ from `HEAD`, with the current content digest of each (deleted
/// files have none, a symlink is digested by its target text, a submodule by its checked-out
/// commit), and one digest over them all. `Err` names what the digest cannot represent; the
/// snapshot is then `unknown` rather than falsely stable.
fn overlay(top: &Path) -> std::result::Result<(String, Vec<Value>), &'static str> {
    let raw = git_out(
        top,
        &[
            "status",
            "--porcelain=v1",
            "-z",
            "--untracked-files=no",
            "--no-renames",
        ],
    )
    .ok_or("git status failed")?;
    let mut entries: Vec<(String, String)> = raw
        .split(|b| *b == 0)
        .filter(|e| e.len() > 3)
        .map(|e| {
            let s = String::from_utf8_lossy(e);
            (s[3..].to_string(), s[..2].trim().to_string())
        })
        .collect();
    entries.sort();
    entries.dedup();
    let mut h = Sha256::new();
    h.update(b"moai-overlay-v1\0");
    let mut files = Vec::new();
    for (path, status) in entries {
        let full = top.join(&path);
        let (tag, sha) = match fs::symlink_metadata(&full) {
            Err(_) => (b'd', None),
            Ok(m) if m.file_type().is_symlink() => (
                b'l',
                fs::read_link(&full)
                    .ok()
                    .and_then(|t| t.to_str().map(|t| digest(t.as_bytes()))),
            ),
            // A changed directory entry is a submodule (gitlink): its checked-out commit is
            // the state. A submodule with uncommitted changes of its own cannot be digested.
            Ok(m) if m.is_dir() => {
                let head = git_line(&full, &["rev-parse", "--verify", "HEAD^{commit}"])
                    .ok_or("a changed submodule has no readable HEAD")?;
                let dirty = git_out(
                    &full,
                    &["status", "--porcelain=v1", "-z", "--untracked-files=no"],
                )
                .ok_or("a changed submodule cannot be inspected")?;
                if !dirty.is_empty() {
                    return Err("a submodule has uncommitted changes");
                }
                (b's', Some(format!("commit:{head}")))
            }
            Ok(_) => (b'f', file_sha(&full)),
        };
        h.update((path.len() as u64).to_be_bytes());
        h.update(path.as_bytes());
        h.update([tag]);
        h.update(sha.as_deref().unwrap_or("").as_bytes());
        files.push(json!({"path": path, "status": status, "sha256": sha}));
    }
    Ok((format!("sha256:{:x}", h.finalize()), files))
}

/// Describes the working tree of `project` (REQ-CTX-001). Writes nothing. Outside git, or
/// without a `HEAD` commit, the base is `unknown` with a reason; nothing is guessed.
pub fn snapshot(project: &Path) -> Result<Value> {
    let root = project_root(project)?;
    let top = git_line(
        &root,
        &["rev-parse", "--path-format=absolute", "--show-toplevel"],
    )
    .map(PathBuf::from);
    let base = git_line(
        &root,
        &["rev-parse", "--verify", "--end-of-options", "HEAD^{commit}"],
    );
    let (status, reason, overlay_digest, files) = match (&top, &base) {
        (Some(top), Some(_)) => match overlay(top) {
            Ok((d, f)) => ("ok", Value::Null, json!(d), json!(f)),
            Err(why) => ("unknown", json!(why), Value::Null, Value::Null),
        },
        (None, _) => (
            "unknown",
            json!("not inside a git work tree"),
            Value::Null,
            Value::Null,
        ),
        (Some(_), None) => ("unknown", json!("no HEAD commit"), Value::Null, Value::Null),
    };
    Ok(json!({
        "schema_version": 1,
        "worktree_path": text(top.as_deref().unwrap_or(&root))?,
        "base_commit": base,
        "base_status": status,
        "reason": reason,
        "dirty_overlay_digest": overlay_digest,
        "dirty_files": files,
        "described_at": utc_now(),
        "network": "disabled",
        "writes": false,
    }))
}

#[derive(Deserialize)]
#[serde(deny_unknown_fields)]
pub struct BundleRequest {
    pub protocol_version: u32,
    /// In priority order: under a budget, later items are dropped first (REQ-CTX-005).
    pub items: Vec<ItemRequest>,
    #[serde(default)]
    pub budget: Budget,
}

#[derive(Deserialize, Clone)]
#[serde(deny_unknown_fields)]
pub struct ItemRequest {
    pub path: String,
    #[serde(default)]
    pub start_line: Option<usize>,
    #[serde(default)]
    pub end_line: Option<usize>,
    pub reason: String,
}

/// Optional. A missing `max_bytes` comes from the profile key `context_max_bytes`, else
/// [`DEFAULT_MAX_BYTES`] (REQ-C8-01); `max_items` has no default. A budget is never
/// enlarged on its own.
#[derive(Deserialize, Default)]
#[serde(deny_unknown_fields)]
pub struct Budget {
    #[serde(default)]
    pub max_bytes: Option<u64>,
    #[serde(default)]
    pub max_items: Option<usize>,
}

enum Resolved {
    Item {
        start: usize,
        end: usize,
        content: String,
    },
    /// Never bundled and never read: a secret name, an ignored path, or outside the tree.
    Blocked(&'static str),
    /// Allowed but not readable as a text range.
    Unreadable(&'static str),
}

/// Resolves one requested range under `root`. `ignored` reports `git check-ignore`; `None`
/// when git could not answer.
fn resolve(root: &Path, req: &ItemRequest, ignored: &mut Option<bool>) -> Resolved {
    let rel = Path::new(&req.path);
    if req.path.is_empty()
        || req.path.contains('\0')
        || !rel
            .components()
            .all(|c| matches!(c, Component::Normal(_) | Component::CurDir))
    {
        return Resolved::Blocked("outside_repository");
    }
    let named_secret = |p: &Path| {
        p.components()
            .any(|c| matches!(c, Component::Normal(n) if secret_name(&n.to_string_lossy())))
    };
    if named_secret(rel) {
        return Resolved::Blocked("secret_pattern");
    }
    let Ok(full) = fs::canonicalize(root.join(rel)) else {
        return Resolved::Unreadable("missing");
    };
    // A symlink may lead out of the tree or onto a secret under another name.
    let inside = match full.strip_prefix(root) {
        Err(_) => return Resolved::Blocked("outside_repository"),
        Ok(inside) if named_secret(inside) => return Resolved::Blocked("secret_pattern"),
        Ok(inside) => inside.to_string_lossy().into_owned(),
    };
    // Both the requested name and the file it resolves to: a symlink alias of an ignored
    // file is as ignored as the file (Codex review #2057). A hardlink alias is not caught:
    // canonicalize resolves symlinks only and check-ignore answers on paths, never inodes.
    let check = |path: &str| {
        git(root)
            .args(["check-ignore", "-q", "--", path])
            .stdout(Stdio::null())
            .stderr(Stdio::null())
            .status()
            .ok()
            .and_then(|s| s.code())
    };
    let codes = [check(&req.path), check(&inside)];
    *ignored = if codes.contains(&Some(0)) {
        return Resolved::Blocked("ignored");
    } else if codes.iter().all(|c| *c == Some(1)) {
        Some(false)
    } else {
        None
    };
    match fs::metadata(&full) {
        Ok(m) if !m.is_file() => return Resolved::Unreadable("not_a_file"),
        Ok(m) if m.len() > FILE_LIMIT => return Resolved::Unreadable("too_large"),
        Err(_) => return Resolved::Unreadable("missing"),
        Ok(_) => {}
    }
    let mut bytes = Vec::new();
    if fs::File::open(&full)
        .and_then(|f| f.take(FILE_LIMIT + 1).read_to_end(&mut bytes))
        .is_err()
    {
        return Resolved::Unreadable("missing");
    }
    let Ok(body) = String::from_utf8(bytes) else {
        return Resolved::Unreadable("not_text");
    };
    let lines: Vec<&str> = body.split_inclusive('\n').collect();
    let start = req.start_line.unwrap_or(1);
    let end = req.end_line.unwrap_or(lines.len());
    if lines.is_empty() && req.start_line.is_none() && req.end_line.is_none() {
        return Resolved::Item {
            start: 0,
            end: 0,
            content: String::new(),
        };
    }
    if start == 0 || start > end || end > lines.len() {
        return Resolved::Unreadable("range_out_of_bounds");
    }
    Resolved::Item {
        start,
        end,
        content: lines[start - 1..end].concat(),
    }
}

fn range(start: Option<usize>, end: Option<usize>) -> Value {
    json!({"start_line": start, "end_line": end})
}

/// The byte budget and where it came from: the request, then the profile key
/// `context_max_bytes`, then [`DEFAULT_MAX_BYTES`] (REQ-C8-01). A missing profile means the
/// default; an unreadable profile is an error, never a guess.
fn resolve_max_bytes(root: &Path, budget: &Budget) -> Result<(u64, &'static str)> {
    if let Some(m) = budget.max_bytes {
        return Ok((m, "request"));
    }
    let profile_path = crate::state_root::path(root, "project.yaml");
    if crate::apply::read_opt(&profile_path)?.is_none() {
        return Ok((DEFAULT_MAX_BYTES, "default"));
    }
    let (profile, _) = crate::profile::read_profile(root)?;
    Ok(match profile.context_max_bytes {
        Some(m) => (m, "profile"),
        None => (DEFAULT_MAX_BYTES, "default"),
    })
}

/// Builds a bundle for `req` in `project` (REQ-CTX-002..006). Writes nothing.
pub fn bundle(project: &Path, req: &BundleRequest) -> Result<Value> {
    if req.protocol_version != 1 {
        return Err("unsupported protocol_version; supported: 1".into());
    }
    if req.items.is_empty() || req.items.len() > MAX_ITEMS {
        return Err("bundle request requires 1 to 256 items".into());
    }
    if req.items.iter().any(|i| i.reason.trim().is_empty()) {
        return Err("every bundle item requires a reason".into());
    }
    let snapshot = snapshot(project)?;
    let root = project_root(project)?;
    let (max_bytes, max_bytes_source) = resolve_max_bytes(&root, &req.budget)?;
    let (mut items, mut omitted) = (Vec::new(), Vec::new());
    let (mut used_bytes, mut truncated) = (0u64, false);
    let mut ignore_check = "git";
    for it in &req.items {
        let mut ignored = None;
        let resolved = resolve(&root, it, &mut ignored);
        if ignored.is_none() && !matches!(resolved, Resolved::Blocked(_)) {
            ignore_check = "unavailable";
        }
        let requested = range(it.start_line, it.end_line);
        let omit = |cause: &str, detail: &str| json!({"path": it.path, "range": requested, "reason": it.reason, "cause": cause, "detail": detail});
        match resolved {
            Resolved::Blocked(why) => omitted.push(omit("blocked_sensitive", why)),
            Resolved::Unreadable(why) => omitted.push(omit("unreadable", why)),
            Resolved::Item {
                start,
                end,
                content,
            } => {
                let size = content.len() as u64;
                let over_bytes = used_bytes + size > max_bytes;
                let over_items = req.budget.max_items.is_some_and(|m| items.len() >= m);
                // Once one item does not fit, every later (lower-priority) item is dropped
                // too, so a dropped item never outranks a kept one.
                if truncated || over_bytes || over_items {
                    truncated = true;
                    omitted.push(omit("budget_truncated", "budget"));
                    continue;
                }
                used_bytes += size;
                // The content field is the masked view (rc.2 #2351): it passes the
                // shared redaction so credential shapes (including the merged gh
                // token prefixes) never leave the project verbatim. The digest
                // stays the digest of the raw file bytes — it is the verification
                // key compare against the file, not of this view. Every bundled
                // item is untrusted data (design: source is data, never authority).
                items.push(json!({
                    "path": it.path,
                    "range": {"start_line": start, "end_line": end},
                    "content_sha256": digest(content.as_bytes()),
                    "bytes": size,
                    "lines": if start == 0 { 0 } else { end + 1 - start },
                    "reason": it.reason,
                    "freshness": "fresh",
                    "untrusted": true,
                    "content": crate::redact::redact_secret_shapes(&content),
                }));
            }
        }
    }
    let refs: Vec<Value> = items
        .iter()
        .map(|i| json!([i["path"], i["range"], i["content_sha256"]]))
        .collect();
    Ok(json!({
        "schema_version": 1,
        "snapshot": snapshot,
        "items": items,
        "omitted": omitted,
        "budget": {
            "max_bytes": max_bytes,
            "max_bytes_source": max_bytes_source,
            "max_items": req.budget.max_items,
            "used_bytes": used_bytes,
            "used_items": refs.len(),
        },
        "tokens_estimate": {"value": used_bytes.div_ceil(4), "method": "bytes_div_4", "model_scope": "none"},
        "ignore_check": ignore_check,
        "bundle_digest": digest(&serde_json::to_vec(&refs).expect("refs")),
        "network": "disabled",
        "writes": false,
    }))
}

#[derive(Deserialize)]
#[serde(deny_unknown_fields)]
pub struct VerifyRequest {
    pub protocol_version: u32,
    #[serde(default)]
    pub base_commit: Option<String>,
    #[serde(default)]
    pub dirty_overlay_digest: Option<String>,
    pub items: Vec<VerifyItem>,
}

#[derive(Deserialize)]
pub struct VerifyItem {
    pub path: String,
    pub range: VerifyRange,
    pub content_sha256: String,
}

#[derive(Deserialize)]
pub struct VerifyRange {
    pub start_line: usize,
    pub end_line: usize,
}

/// Re-reads every bundled range and marks it `stale` when its bytes no longer match the
/// recorded digest, or it can no longer be read (REQ-CTX-003). Writes nothing.
pub fn verify(project: &Path, req: &VerifyRequest) -> Result<Value> {
    if req.protocol_version != 1 {
        return Err("unsupported protocol_version; supported: 1".into());
    }
    if req.items.is_empty() || req.items.len() > MAX_ITEMS {
        return Err("verify request requires 1 to 256 items".into());
    }
    let root = project_root(project)?;
    let mut stale = 0;
    let items: Vec<Value> = req
        .items
        .iter()
        .map(|it| {
            let ask = ItemRequest {
                path: it.path.clone(),
                start_line: (it.range.start_line > 0).then_some(it.range.start_line),
                end_line: (it.range.end_line > 0).then_some(it.range.end_line),
                reason: String::new(),
            };
            let (fresh, detail) = match resolve(&root, &ask, &mut None) {
                Resolved::Item { content, .. } => {
                    let same = digest(content.as_bytes()) == it.content_sha256;
                    (
                        same,
                        if same {
                            "digest_matches"
                        } else {
                            "digest_changed"
                        },
                    )
                }
                Resolved::Blocked(why) | Resolved::Unreadable(why) => (false, why),
            };
            stale += usize::from(!fresh);
            json!({
                "path": it.path,
                "range": {"start_line": it.range.start_line, "end_line": it.range.end_line},
                "freshness": if fresh { "fresh" } else { "stale" },
                "reread_required": !fresh,
                "detail": detail,
            })
        })
        .collect();
    let now = snapshot(project)?;
    let snapshot_freshness = match (&req.base_commit, &req.dirty_overlay_digest) {
        (Some(b), Some(d)) if now["base_status"] == "ok" => {
            if now["base_commit"] == json!(b) && now["dirty_overlay_digest"] == json!(d) {
                "fresh"
            } else {
                "stale"
            }
        }
        _ => "unknown",
    };
    Ok(json!({
        "schema_version": 1,
        "items": items,
        "stale_count": stale,
        "snapshot_freshness": snapshot_freshness,
        "network": "disabled",
        "writes": false,
    }))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn secret_patterns_match_names_not_substrings() {
        for name in [
            ".env",
            ".env.local",
            "server.pem",
            "tls.key",
            "id_rsa",
            "id_ed25519.pub",
        ] {
            assert!(secret_name(name), "{name}");
        }
        for name in ["env", "my.env.txt", "pem.md", "keys.rs", "README.md"] {
            assert!(!secret_name(name), "{name}");
        }
    }
}
