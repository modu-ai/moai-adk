//! Codex adapter (02-C #1993): projection plan, projection apply (I4 #2241) and read-only
//! doctor. Only `projection_apply` writes, and only inside the caller's explicit target.
//!
//! Sources (read 2026-11-08):
//! - https://learn.chatgpt.com/docs/build-skills — skill locations (`.agents/skills` from the
//!   working directory up to the repository root, `$HOME/.agents/skills`, `/etc/codex/skills`),
//!   required `name`/`description`, same-name skills are not merged, `agents/openai.yaml`
//!   fields, `allow_implicit_invocation` default `true`.
//! - https://learn.chatgpt.com/docs/agent-configuration/agents-md — global `AGENTS.override.md`
//!   else `AGENTS.md` (first non-empty), project root down to the working directory with at most
//!   one file per directory (override, `AGENTS.md`, fallback names), blank-line joins, stop at
//!   `project_doc_max_bytes` (32 KiB by default).
//! - https://learn.chatgpt.com/docs/config-file/config-reference — `projects.<path>.trust_level`;
//!   untrusted projects skip the project `.codex/` layer; `tui.status_line` item identifiers are
//!   not enumerated.
//!
//! RC1 interim decisions (leader, 2026-11-08): this doctor and this projection plan read no
//! TOML, so trust, inline hooks and the status line are `not_evaluated` and never reported as
//! trusted; this projection generates no hooks (hook generation is `hooks-plan` /
//! `hooks-apply --host codex`, rc.2 I1 #2238); the projection target is always explicit.
//! Writing the project `.codex/config.toml` is the separate plan/apply in `codex_config`
//! (#2240); the doctor reading it is a follow-up.
//! The install location (`~/.agents/skills/moai` vs `<repo>/.agents/skills/moai`) stays a
//! HOLD (P04 / contract gap C5): there is no default, the caller names the target.
use crate::{apply, digest, package_digest};
use serde_json::{json, Value};
use std::{
    fs,
    path::{Component, Path, PathBuf},
};

type Result<T> = std::result::Result<T, String>;

/// Default of `project_doc_max_bytes` stated on the AGENTS.md page (32 KiB).
pub const DEFAULT_MAX_BYTES: u64 = 32 * 1024;
const FILE_LIMIT: u64 = 1024 * 1024;
const TOML_HOLD: &str = "not read: this operation reads no TOML (leader decision 2026-11-08, TOML surface 02-B #1992); the project .codex/config.toml is written only by the codex_config plan/apply (#2240), and reading it here is a follow-up";
const STATUS_LINE_HOLD: &str = "tui.status_line item identifiers are not listed in the official config reference (read 2026-11-09); this plan writes nothing, and codex_config (#2240) writes the items unverified";
/// The package's own digest record (02-A); it lists every other package file's sha256.
const SOURCE_MANIFEST: &str = "references/source-manifest.json";
const TARGET_FILES: usize = 4096;
const LOCATION_HOLD: &str = "caller-chosen target; ~/.agents/skills/moai vs <repo>/.agents/skills/moai is a HOLD (P04 / C5), no default";

/// Options for `agents/openai.yaml`; `None` keeps the documented default by omitting the key.
pub struct OpenAiYaml {
    pub display_name: String,
    pub short_description: String,
    pub default_prompt: String,
    pub allow_implicit_invocation: Option<bool>,
}

fn yaml_scalar(field: &str, value: &str) -> Result<String> {
    if value.is_empty() || value.chars().count() > 256 || value.chars().any(char::is_control) {
        return Err(format!(
            "invalid {field}: 1-256 characters without control characters"
        ));
    }
    Ok(format!(
        "\"{}\"",
        value.replace('\\', "\\\\").replace('"', "\\\"")
    ))
}

/// Renders `agents/openai.yaml` with only the explicitly chosen fields (REQ-02C-02).
pub fn openai_yaml(o: &OpenAiYaml) -> Result<String> {
    let mut text = format!(
        "interface:\n  display_name: {}\n  short_description: {}\n  default_prompt: {}\n",
        yaml_scalar("display_name", &o.display_name)?,
        yaml_scalar("short_description", &o.short_description)?,
        yaml_scalar("default_prompt", &o.default_prompt)?,
    );
    if let Some(allow) = o.allow_implicit_invocation {
        text.push_str(&format!("policy:\n  allow_implicit_invocation: {allow}\n"));
    }
    Ok(text)
}

fn package_files(root: &Path, dir: &Path, out: &mut Vec<(String, Vec<u8>)>) -> Result<()> {
    let mut entries: Vec<PathBuf> = fs::read_dir(dir)
        .map_err(|_| "package directory unavailable")?
        .map(|e| {
            e.map(|e| e.path())
                .map_err(|_| "package entry unavailable".to_string())
        })
        .collect::<Result<_>>()?;
    entries.sort();
    for path in entries {
        if path.is_dir() {
            package_files(root, &path, out)?;
        } else {
            let rel = path
                .strip_prefix(root)
                .map_err(|_| "package path invalid")?
                .to_str()
                .ok_or("nonportable package path")?
                .replace('\\', "/");
            out.push((rel, fs::read(&path).map_err(|_| "package read failed")?));
        }
    }
    Ok(())
}

/// REQ-02C-01: the projected SKILL.md frontmatter holds only `name: moai` and `description`.
fn check_frontmatter(skill: &[u8]) -> Result<()> {
    let text = std::str::from_utf8(skill).map_err(|_| "invalid package: SKILL.md is not UTF-8")?;
    let body = text
        .strip_prefix("---\n")
        .and_then(|rest| rest.split_once("\n---\n"))
        .map(|(front, _)| front)
        .ok_or("invalid package: SKILL.md frontmatter missing")?;
    let mut keys = Vec::new();
    for line in body.lines().filter(|l| !l.trim().is_empty()) {
        if line.starts_with([' ', '\t']) {
            return Err(
                "invalid package: SKILL.md frontmatter must hold only name and description".into(),
            );
        }
        let (key, value) = line
            .split_once(':')
            .ok_or("invalid package: SKILL.md frontmatter line without a key")?;
        keys.push((key.trim(), value.trim()));
    }
    let names: Vec<&str> = keys.iter().map(|(k, _)| *k).collect();
    if names.len() != 2 || !names.contains(&"name") || !names.contains(&"description") {
        return Err(
            "invalid package: SKILL.md frontmatter must hold only name and description".into(),
        );
    }
    if keys.iter().any(|(k, v)| *k == "name" && *v != "moai") {
        return Err("invalid package: SKILL.md name must be moai".into());
    }
    Ok(())
}

/// REQ-I4-01: compares every package file (except the manifest itself) with its
/// `source-manifest.json` entry, and every entry with a package file. The manifest stores
/// bare hex; `digest` prefixes `sha256:`. Reported, never thrown: the plan shows the state
/// and `projection_apply` refuses anything but `verified`.
fn manifest_check(files: &[(String, Vec<u8>)]) -> Value {
    let Some((_, bytes)) = files.iter().find(|(rel, _)| rel == SOURCE_MANIFEST) else {
        return json!({"path": SOURCE_MANIFEST, "status": "missing", "mismatches": []});
    };
    let Some(entries) = serde_json::from_slice::<Value>(bytes)
        .ok()
        .and_then(|m| m.get("files").and_then(Value::as_object).cloned())
    else {
        return json!({"path": SOURCE_MANIFEST, "status": "unreadable", "mismatches": []});
    };
    let mut mismatches = Vec::new();
    for (rel, bytes) in files.iter().filter(|(rel, _)| rel != SOURCE_MANIFEST) {
        match entries
            .get(rel)
            .and_then(|e| e.get("sha256"))
            .and_then(Value::as_str)
        {
            None => mismatches.push(json!({"path": rel, "reason": "no manifest entry"})),
            Some(hex) if format!("sha256:{hex}") != digest(bytes) => {
                mismatches.push(json!({"path": rel, "reason": "sha256 differs from the manifest"}))
            }
            Some(_) => {}
        }
    }
    for rel in entries.keys() {
        if !files.iter().any(|(r, _)| r == rel) {
            mismatches.push(json!({"path": rel, "reason": "listed in the manifest but absent"}));
        }
    }
    let status = if mismatches.is_empty() {
        "verified"
    } else {
        "mismatch"
    };
    json!({"path": SOURCE_MANIFEST, "status": status, "mismatches": mismatches})
}

/// Relative paths of the regular files already under `target` (empty when it is absent).
/// A symlink anywhere under the target is refused: the apply never follows one.
fn target_files(target: &Path) -> Result<Vec<String>> {
    match fs::symlink_metadata(target) {
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => return Ok(Vec::new()),
        Err(_) => return Err("target metadata unavailable".into()),
        Ok(m) if !m.is_dir() => {
            return Err("invalid target: must be a directory, not a file or symlink".into())
        }
        Ok(_) => {}
    }
    let mut out = Vec::new();
    let mut stack = vec![target.to_path_buf()];
    while let Some(dir) = stack.pop() {
        for entry in fs::read_dir(&dir).map_err(|_| "target directory unavailable")? {
            let path = entry.map_err(|_| "target entry unavailable")?.path();
            let meta = fs::symlink_metadata(&path).map_err(|_| "target metadata unavailable")?;
            if meta.file_type().is_symlink() {
                return Err("invalid target: a symlink under the target is refused".into());
            }
            if meta.is_dir() {
                stack.push(path);
            } else {
                let rel = path
                    .strip_prefix(target)
                    .map_err(|_| "target path invalid")?
                    .to_str()
                    .ok_or("nonportable target path")?
                    .replace('\\', "/");
                out.push(rel);
                if out.len() > TARGET_FILES {
                    return Err("target file limit exceeded".into());
                }
            }
        }
    }
    out.sort();
    Ok(out)
}

/// `create` (absent), `unchanged` (same bytes) or `conflict` (different bytes at the path).
fn target_action(target: &Path, rel: &str, want: &[u8]) -> Result<&'static str> {
    Ok(match apply::read_opt(&target.join(rel))? {
        None => "create",
        Some(have) if have == want => "unchanged",
        Some(_) => "conflict",
    })
}

/// `target` with symlinks resolved, for the outside-the-package check. The target need not
/// exist yet: its nearest existing ancestor is canonicalized and the missing tail appended.
/// `..` is refused because it cannot be resolved through a tail that does not exist, and an
/// entry that exists but cannot be resolved (a dangling symlink) is refused, not skipped.
fn resolved_target(target: &Path) -> Result<PathBuf> {
    if target.components().any(|c| c == Component::ParentDir) {
        return Err("invalid target: '..' components are not allowed".into());
    }
    let mut existing = target;
    let mut tail = Vec::new();
    loop {
        match existing.canonicalize() {
            Ok(mut resolved) => {
                resolved.extend(tail.iter().rev());
                return Ok(resolved);
            }
            Err(_) if fs::symlink_metadata(existing).is_ok() => {
                return Err("invalid target: path cannot be resolved".into());
            }
            Err(_) => {
                tail.push(
                    existing
                        .file_name()
                        .ok_or("invalid target: path cannot be resolved")?,
                );
                existing = existing
                    .parent()
                    .ok_or("invalid target: path cannot be resolved")?;
            }
        }
    }
}

/// Plans a Codex projection of the common package into `target` without writing anything.
/// `status_line` and `disable_bundled_skills` are config requests: they are reported as HOLD.
/// Each file carries `target_action` against what is already at the target, so a plan run
/// after an install is its verification (every file `unchanged`).
pub fn projection_plan(
    package: &Path,
    target: &Path,
    yaml: &OpenAiYaml,
    status_line: &[String],
    disable_bundled_skills: bool,
) -> Result<Value> {
    let package_digest = package_digest(package)?;
    if !target.is_absolute() {
        return Err("invalid target: must be an absolute path".into());
    }
    let package_abs = package.canonicalize().map_err(|_| "package unavailable")?;
    // Both sides are compared resolved: a lexical target such as `/tmp/...` against a
    // canonical package `/private/tmp/...` would otherwise pass (#2241 audit FAIL).
    if resolved_target(target)?.starts_with(&package_abs) {
        return Err("invalid target: must be outside the package".into());
    }
    if target.file_name().and_then(|n| n.to_str()) != Some("moai") {
        return Err("invalid target: the skill directory must be named moai".into());
    }
    let mut files = Vec::new();
    package_files(&package_abs, &package_abs, &mut files)?;
    if files.iter().any(|(rel, _)| rel.starts_with("agents/")) {
        return Err(
            "invalid package: agents/ belongs to the Codex projection, not the package".into(),
        );
    }
    let skill = &files
        .iter()
        .find(|(rel, _)| rel == "SKILL.md")
        .ok_or("package SKILL.md missing")?
        .1;
    check_frontmatter(skill)?;
    let existing = target_files(target)?;
    let mut plan: Vec<Value> = Vec::new();
    for (rel, bytes) in &files {
        plan.push(
            json!({"path": rel, "action": "copy", "source": "package", "sha256": digest(bytes),
            "target_action": target_action(target, rel, bytes)?}),
        );
    }
    // REQ-02C-01: the planned SKILL.md is the package's bytes, read again and compared.
    let reread = fs::read(package_abs.join("SKILL.md")).map_err(|_| "package read failed")?;
    let skill_md_identical = plan
        .iter()
        .any(|e| e["path"] == "SKILL.md" && e["sha256"] == digest(&reread));
    let rendered = openai_yaml(yaml)?;
    plan.push(
        json!({"path": "agents/openai.yaml", "action": "create", "source": "adapter",
        "sha256": digest(rendered.as_bytes()), "content": rendered,
        "target_action": target_action(target, "agents/openai.yaml", rendered.as_bytes())?}),
    );
    // Files already at the target that the projection does not own are never touched.
    let preserved: Vec<&String> = existing
        .iter()
        .filter(|rel| !plan.iter().any(|e| e["path"] == rel.as_str()))
        .collect();
    let manifest = manifest_check(&files);
    let bound: Vec<Value> = plan
        .iter()
        .map(|e| json!([e["path"], e["sha256"], e["target_action"]]))
        .collect();
    let plan_digest = digest(
        &serde_json::to_vec(&json!({"target": target, "package_digest": package_digest,
            "files": bound, "preserved": preserved, "manifest": manifest["status"]}))
        .map_err(|_| "plan encode failed")?,
    );
    let unverified: Vec<&String> = status_line.iter().collect();
    let config = if status_line.is_empty() && !disable_bundled_skills {
        json!({"status": "none_requested"})
    } else {
        json!({"status": "hold", "reason": TOML_HOLD,
            "requested": {"tui.status_line": status_line, "skills.bundled.enabled": if disable_bundled_skills { json!(false) } else { Value::Null }},
            "unverified_items": unverified, "unverified_reason": STATUS_LINE_HOLD,
            "writes": []})
    };
    Ok(json!({
        "host": "codex",
        "target": target,
        "package_digest": package_digest,
        "plan_digest": plan_digest,
        "files": plan,
        "preserved": preserved,
        "source_manifest": manifest,
        "location": LOCATION_HOLD,
        "skill_md_identical": skill_md_identical,
        "config_changes": config,
        "hooks": {"generated": [], "reason": "a projection generates no hooks; see hooks-plan / hooks-apply --host codex (#2238)"},
        "invocation": {"desktop": "$moai in the composer", "cli": "/skills or $moai", "ide": "/skills or $moai"},
        "notes": [
            "plan only: nothing is written; codex-projection-apply writes this plan when given its plan_digest",
            "target candidates ~/.agents/skills/moai and <repo>/.agents/skills/moai are a HOLD; the caller chooses",
            "a second skill named moai in another location is not merged; both can appear in selectors"
        ],
        "files_created": [],
        "host_loaded": false,
        "execution_verified": false
    }))
}

/// Installs the planned projection into `target` (I4 #2241). Gates, in order: a permission
/// reference; the caller's `plan_digest` equals a fresh plan's; the source manifest is
/// `verified` (REQ-I4-01: a copy whose digest differs from the manifest is never written);
/// no planned file conflicts with different bytes at the target. Every check runs before
/// the first byte. Files are created exclusively (never overwritten), then every planned
/// file is re-read and compared with the plan. Re-running on an installed target writes
/// nothing. Config requests (status line, bundled skills) stay HOLD and are not taken.
pub fn projection_apply(
    package: &Path,
    target: &Path,
    yaml: &OpenAiYaml,
    confirmed: &str,
    permission: &str,
) -> Result<Value> {
    // Caller authorization remains host-owned; this reference is not an authorization verifier.
    if !crate::contract::permission_reference_ok(permission) {
        return Err("permission reference required".into());
    }
    let plan = projection_plan(package, target, yaml, &[], false)?;
    if plan["plan_digest"] != confirmed {
        return Err("stale or altered plan; re-preview required".into());
    }
    let manifest = &plan["source_manifest"];
    if manifest["status"] != "verified" {
        let paths: Vec<&str> = manifest["mismatches"]
            .as_array()
            .map(|a| a.iter().filter_map(|m| m["path"].as_str()).collect())
            .unwrap_or_default();
        return Err(format!(
            "source manifest digest mismatch ({}): {}; nothing applied",
            manifest["status"].as_str().unwrap_or("unknown"),
            paths.join(", ")
        ));
    }
    let files = plan["files"].as_array().ok_or("plan encode failed")?;
    let conflicts: Vec<&str> = files
        .iter()
        .filter(|e| e["target_action"] == "conflict")
        .filter_map(|e| e["path"].as_str())
        .collect();
    if !conflicts.is_empty() {
        return Err(format!(
            "conflict in codex projection: {} differ from the planned bytes; nothing applied",
            conflicts.join(", ")
        ));
    }
    let package_abs = package.canonicalize().map_err(|_| "package unavailable")?;
    let (mut created, mut unchanged) = (Vec::new(), Vec::new());
    for entry in files {
        let rel = entry["path"].as_str().ok_or("plan encode failed")?;
        if entry["target_action"] != "create" {
            unchanged.push(rel);
            continue;
        }
        let bytes = match entry["content"].as_str() {
            Some(content) => content.as_bytes().to_vec(),
            None => fs::read(package_abs.join(rel)).map_err(|_| "package read failed")?,
        };
        if digest(&bytes) != entry["sha256"] {
            return Err(partial(&created, rel, "package changed since plan"));
        }
        let path = target.join(rel);
        if let Some(parent) = path.parent() {
            fs::create_dir_all(parent)
                .map_err(|_| partial(&created, rel, "target directory write failed"))?;
        }
        apply::publish_exclusive(&path, &bytes).map_err(|e| partial(&created, rel, &e))?;
        created.push(rel);
    }
    // Verification: every planned file now holds the planned bytes.
    for entry in files {
        let rel = entry["path"].as_str().ok_or("plan encode failed")?;
        let have = apply::read_opt(&target.join(rel))?.map(|b| digest(&b));
        if have.as_deref() != entry["sha256"].as_str() {
            return Err(format!(
                "codex projection digest mismatch after write: {rel}"
            ));
        }
    }
    Ok(json!({
        "host": "codex",
        "target": target,
        "plan_digest": plan["plan_digest"],
        "package_digest": plan["package_digest"],
        "source_manifest": manifest["status"],
        "files_created": created,
        "files_unchanged": unchanged,
        "preserved": plan["preserved"],
        "verified": true,
        "location": LOCATION_HOLD,
        "config_changes": {"status": "hold", "reason": TOML_HOLD, "writes": []},
        "hooks": plan["hooks"],
        "host_loaded": false,
        "execution_verified": false
    }))
}

/// A failure after the first created file is a partial apply; before it, the inner error.
fn partial(created: &[&str], at: &str, inner: &str) -> String {
    if created.is_empty() {
        return format!("{inner} ({at}); nothing applied");
    }
    format!(
        "partial apply: created {} file(s) before {at}; failed at {at}: {inner}",
        created.len()
    )
}

/// Reads a regular file of bounded size; `Ok(None)` when it is absent.
fn read_text(path: &Path) -> Result<Option<String>> {
    match fs::symlink_metadata(path) {
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => Ok(None),
        Err(_) => Err("target metadata unavailable".into()),
        Ok(m) if !m.is_file() || m.len() > FILE_LIMIT => {
            Err("target is not a bounded regular file".into())
        }
        Ok(_) => fs::read_to_string(path)
            .map(Some)
            .map_err(|_| "target read failed".into()),
    }
}

/// The first of `names` in `dir` that exists and is not empty ("Codex skips empty files").
fn first_doc(dir: &Path, names: &[String]) -> Result<Option<(PathBuf, String)>> {
    for name in names {
        let path = dir.join(name);
        if let Some(text) = read_text(&path)? {
            if text.trim().is_empty() {
                continue;
            }
            return Ok(Some((path, text)));
        }
    }
    Ok(None)
}

/// `roots` are skill roots (directories that hold `<name>/SKILL.md`). The package
/// name is `moai` (#2286); an earlier-generation (Go-era) MoAI `moai` skill sits at
/// the same path, so this read-only doctor reports it too and cannot tell the two
/// apart — ownership is judged by the update and project-copy plans, which block it.
fn moai_skills(roots: &[(PathBuf, &str)]) -> Vec<Value> {
    roots
        .iter()
        .map(|(root, scope)| (root.join("moai/SKILL.md"), scope))
        .filter(|(path, _)| path.is_file())
        .map(|(path, scope)| json!({"path": path, "scope": scope}))
        .collect()
}

/// Read-only Codex doctor for `project` (the project root) and `cwd` inside it.
pub fn doctor(
    project: &Path,
    cwd: &Path,
    codex_home: &Path,
    home: &Path,
    fallback: &[String],
    max_bytes: Option<u64>,
) -> Result<Value> {
    let root = project
        .canonicalize()
        .map_err(|_| "project directory unavailable")?;
    let cwd = cwd.canonicalize().map_err(|_| "path unavailable")?;
    if !root.is_dir() {
        return Err("project must be a directory".into());
    }
    if !cwd.starts_with(&root) {
        return Err("invalid cwd: must be inside the project".into());
    }
    for name in fallback {
        if name.is_empty() || name.contains('/') || name.contains('\\') || name == ".." {
            return Err("invalid fallback name: a plain file name is required".into());
        }
    }
    let (budget, budget_source) = match max_bytes {
        Some(0) => return Err("invalid max-bytes: must be positive".into()),
        Some(n) => (n, "argument"),
        None => (
            DEFAULT_MAX_BYTES,
            "default 32 KiB stated on the AGENTS.md page",
        ),
    };
    // Global level: AGENTS.override.md if present, else AGENTS.md; the first non-empty file.
    let global_names = vec!["AGENTS.override.md".to_string(), "AGENTS.md".to_string()];
    let g_hit = first_doc(codex_home, &global_names)?;
    let global = json!({"dir": codex_home, "considered": global_names,
        "included": g_hit.as_ref().map(|(p, _)| p), "bytes": g_hit.as_ref().map(|(_, t)| t.len())});
    // Project level: root down to cwd, at most one file per directory.
    let mut dirs = vec![root.clone()];
    for part in cwd
        .strip_prefix(&root)
        .map_err(|_| "invalid cwd: must be inside the project")?
        .components()
    {
        let next = dirs.last().unwrap().join(part);
        dirs.push(next);
    }
    let mut names = vec!["AGENTS.override.md".to_string(), "AGENTS.md".to_string()];
    names.extend(fallback.iter().cloned());
    let mut chain = Vec::new();
    let mut included = Vec::new();
    let mut total: u64 = 0;
    let mut over_at: Option<PathBuf> = None;
    for dir in &dirs {
        let considered = &names;
        // "Codex skips empty files" (AGENTS.md page). Moving on to the next name in the same
        // directory after an empty one is inferred; the page does not say so for projects.
        let hit = first_doc(dir, &names)?;
        let entry = match &hit {
            None => json!({"dir": dir, "considered": considered, "included": null}),
            Some((path, text)) => {
                let bytes = text.len() as u64;
                let joined = if included.is_empty() {
                    bytes
                } else {
                    bytes + 2
                };
                if over_at.is_none() && total + joined > budget {
                    over_at = Some(path.clone());
                }
                total += joined;
                included.push(json!({"path": path, "bytes": bytes}));
                json!({"dir": dir, "considered": considered, "included": path})
            }
        };
        chain.push(entry);
    }
    // Skill locations from the build-skills page, for the duplicate-name warning (REQ-02C-04).
    let mut skill_dirs: Vec<(PathBuf, &str)> = dirs
        .iter()
        .rev()
        .map(|d| (d.join(".agents/skills"), "repository"))
        .collect();
    skill_dirs.push((home.join(".agents/skills"), "user"));
    skill_dirs.push((PathBuf::from("/etc/codex/skills"), "admin"));
    let found = moai_skills(&skill_dirs);
    let duplicate = found.len() > 1;
    let layer = root.join(".codex");
    Ok(json!({
        "host": "codex",
        "project_root": root,
        "cwd": cwd,
        "agents": {
            "global": global,
            "project_chain": chain,
            "included": included,
            "project_bytes": total,
            "max_bytes": budget,
            "max_bytes_source": budget_source,
            "fallback_names": fallback,
            "fallback_source": if fallback.is_empty() { "none given; project_doc_fallback_filenames is not read (this doctor reads no TOML)" } else { "argument" },
            "over_budget": over_at.is_some(),
            "over_budget_at": over_at,
            "budget_note": "the project chain is counted with blank-line joins; whether Codex counts the global file in the same budget and how it cuts the crossing file are not stated on the page",
            "agents_override_created": false
        },
        "trust": {"status": "not_evaluated", "reason": TOML_HOLD,
            "consequence": "if the project is untrusted, Codex skips .codex/config.toml, hooks and rules (blocked_by_trust); this doctor cannot tell and never reports trusted"},
        "codex_layer": {
            "config_toml": layer.join("config.toml").is_file(),
            "hooks_json": layer.join("hooks.json").is_file(),
            "inline_hooks": "not_evaluated",
            "status_line": "not_evaluated",
            "reason": TOML_HOLD
        },
        "skills": {"moai": found, "duplicate_warning": if duplicate {
            json!("more than one skill named moai was found; Codex does not merge them and both can appear in selectors")
        } else { Value::Null }},
        "writes": [],
        "host_loaded": false,
        "execution_verified": false
    }))
}
