//! GitHub templates for a generated project (Issue #2013, 08-B).
//!
//! The product templates under `templates/github/` are derived from this repository's
//! `.github/` forms by structure, without the repository's own text (#2434 AC-6; the byte
//! parity of 08-B is replaced, checked by `tests/github_templates_test.py`); this module
//! embeds those bytes and plans and applies them into a project with the 01-B 3-way engine
//! (`decide`, `publish_exclusive`/`publish_replace`, `save_before_bytes`, `Receipt`).
//!
//! Decisions (leader, under the user's "lift only the HOLD" rule, recorded on #2013):
//! - template bytes are embedded in the binary with `include_str!` (①A, as generation.rs
//!   renders its files from strings in the binary);
//! - Rust handles the installed bytes and the 3-way only; form structure is checked by the Python
//!   static tests (②A: no YAML crate, Cargo.toml keeps its dependencies);
//! - `config.yml` is not a template (③B: the development copy has none).
//!
//! Generation is opted in by the profile key `github_issues` (Issue #2106) or by the
//! request flag of the same name; an explicit request flag wins over the profile, and a
//! folder that merely exists never triggers it (REQ-08B-04).
//! Baselines live in their own manifest `.moai-cli/manifests/github.json` with the
//! `owned_files[].baseline_sha256` shape of `generation.json`: that file's writers are
//! private to agents.rs and generation.rs, so 08-B adds no third writer to it.
//! Files are planned whole (`managed_scope: whole-file`): a user file at a target path
//! without a baseline is a conflict and is never overwritten; other files under
//! `.github/` (user Markdown templates, `config.yml`) are listed as preserved and never
//! touched. GitHub renders forms only from the default branch, so nothing here reports
//! rendering as verified (REQ-08B-10).
use crate::apply::{
    decide, new_request_id, publish_exclusive, publish_replace, read_opt, root_of,
    save_before_bytes, target, Decision, Receipt,
};
use crate::digest;
use crate::spec::{Canonical, SortedKeys};
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use std::{fs, path::Path};

type Result<T> = std::result::Result<T, String>;

pub const MANIFEST: &str = ".moai-cli/manifests/github.json";
/// Marker of an unresolved template variable (`placeholder_pattern` in
/// `templates/github/manifest.json`).
pub const PLACEHOLDER: &str = "{{";

/// `(target path, embedded bytes)` of every template, in manifest order.
pub const TEMPLATES: &[(&str, &str)] = &[
    (
        ".github/ISSUE_TEMPLATE/moai-feature.yml",
        include_str!("../templates/github/ISSUE_TEMPLATE/moai-feature.yml"),
    ),
    (
        ".github/ISSUE_TEMPLATE/moai-bugfix.yml",
        include_str!("../templates/github/ISSUE_TEMPLATE/moai-bugfix.yml"),
    ),
    (
        ".github/ISSUE_TEMPLATE/moai-investigation.yml",
        include_str!("../templates/github/ISSUE_TEMPLATE/moai-investigation.yml"),
    ),
    (
        ".github/ISSUE_TEMPLATE/moai-followup.yml",
        include_str!("../templates/github/ISSUE_TEMPLATE/moai-followup.yml"),
    ),
    (
        ".github/moai/labels.yml",
        include_str!("../templates/github/labels.yml"),
    ),
];

#[derive(Clone, Debug, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct GithubRequest {
    pub protocol_version: u32,
    /// Explicit opt-in for this run. When absent, the profile's `github_issues` decides;
    /// when present, it wins over the profile.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub github_issues: Option<bool>,
}

/// Resolves the opt-in: the request flag first, then the profile key. A missing profile
/// or key means not opted in; an unreadable profile is an error, never a guess.
fn opted_in(root: &Path, request: &GithubRequest) -> Result<(bool, String)> {
    if let Some(flag) = request.github_issues {
        return Ok((flag, format!("opt-in from request.github_issues = {flag}")));
    }
    let profile_path = root.join(crate::PROJECT_STATE_DIR).join("project.yaml");
    if read_opt(&profile_path)?.is_none() {
        return Ok((
            false,
            "opt-in: request.github_issues absent and no project profile".into(),
        ));
    }
    let (profile, _) = crate::profile::read_profile(root)?;
    Ok(match profile.github_issues {
        Some(flag) => (flag, format!("opt-in from profile github_issues = {flag}")),
        None => (
            false,
            "opt-in: request.github_issues and profile github_issues both absent".into(),
        ),
    })
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Eq)]
#[serde(deny_unknown_fields)]
pub struct GithubFilePlan {
    pub path: String,
    /// create | merge | unchanged | conflict
    pub action: String,
    pub expected_sha256: Option<String>,
    pub baseline_sha256: Option<String>,
    pub rendered_sha256: String,
    /// True when the file already equals the rendering but no baseline is recorded
    /// (an interrupted apply): the next apply records it instead of reporting a conflict.
    pub adopt_baseline: bool,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub detail: Option<String>,
}

#[derive(Clone, Debug, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct GithubPlan {
    pub protocol_version: u32,
    pub request: GithubRequest,
    pub files: Vec<GithubFilePlan>,
    /// Existing `.github/` files this plan does not own; they are left as they are.
    pub preserved: Vec<String>,
    pub notes: Vec<String>,
    pub plan_digest: String,
}

/// Paths of an unresolved placeholder among `templates` (REQ-08B-08).
pub fn unresolved(templates: &[(&str, &str)]) -> Vec<String> {
    templates
        .iter()
        .filter(|(_, text)| text.contains(PLACEHOLDER))
        .map(|(path, _)| path.to_string())
        .collect()
}

fn manifest_baselines(root: &Path) -> Result<Vec<(String, String)>> {
    let Some(bytes) = read_opt(&root.join(MANIFEST))? else {
        return Ok(Vec::new());
    };
    let v: Value = serde_json::from_slice(&bytes)
        .map_err(|_| "github manifest unreadable; inspect before generating".to_string())?;
    Ok(v["owned_files"]
        .as_array()
        .map(|a| {
            a.iter()
                .filter_map(|f| {
                    Some((
                        f["path"].as_str()?.to_string(),
                        f["baseline_sha256"].as_str()?.to_string(),
                    ))
                })
                .collect()
        })
        .unwrap_or_default())
}

fn preserved(root: &Path) -> Vec<String> {
    let mut out = Vec::new();
    let mut stack = vec![root.join(".github")];
    while let Some(dir) = stack.pop() {
        let Ok(entries) = fs::read_dir(&dir) else {
            continue;
        };
        for e in entries.flatten() {
            let path = e.path();
            if path.is_dir() {
                stack.push(path);
            } else if let Ok(rel) = path.strip_prefix(root) {
                let rel = rel.to_string_lossy().replace('\\', "/");
                if !TEMPLATES.iter().any(|(t, _)| *t == rel) {
                    out.push(rel);
                }
            }
        }
    }
    out.sort();
    out
}

fn plan_digest(p: &GithubPlan) -> String {
    let mut v = serde_json::to_value(p).expect("plan serializes");
    v["plan_digest"] = Value::Null;
    digest(&SortedKeys.bytes(&v))
}

/// `github-plan`: read-only.
pub fn github_plan(root: &Path, request: GithubRequest) -> Result<GithubPlan> {
    if request.protocol_version != 1 {
        return Err("unsupported protocol_version; supported: 1".into());
    }
    let root = root_of(root)?;
    let mut files = Vec::new();
    let mut notes = Vec::new();
    let (enabled, source) = opted_in(&root, &request)?;
    if !enabled {
        notes.push("not generated: not opted in; an existing .github/ never opts in".into());
    } else {
        let baselines = manifest_baselines(&root)?;
        for (rel, text) in TEMPLATES {
            let current = read_opt(&target(&root, rel)?)?;
            let current_digest = current.as_deref().map(digest);
            let baseline = baselines
                .iter()
                .find(|(p, _)| p == rel)
                .map(|(_, d)| d.clone());
            let rendered = digest(text.as_bytes());
            let (action, detail) = match decide(
                baseline.as_deref(),
                current_digest.as_deref(),
                Some(&rendered),
            ) {
                Decision::KeepCurrent => ("unchanged", None),
                Decision::Apply if current.is_none() => ("create", None),
                Decision::Apply => ("merge", None),
                Decision::Conflict if baseline.is_none() => (
                    "conflict",
                    Some("an existing file has no recorded baseline; it is kept".to_string()),
                ),
                Decision::Conflict => (
                    "conflict",
                    Some("the file changed since the last apply; it is kept".to_string()),
                ),
            };
            let adopt = action == "unchanged" && baseline.is_none();
            files.push(GithubFilePlan {
                path: rel.to_string(),
                action: action.into(),
                expected_sha256: current_digest,
                baseline_sha256: baseline,
                rendered_sha256: rendered,
                adopt_baseline: adopt,
                detail,
            });
        }
        notes.push(
            "GitHub reads Issue forms only from the default branch; rendering is not_run".into(),
        );
    }
    notes.push(source);
    let mut plan = GithubPlan {
        protocol_version: 1,
        request,
        files,
        preserved: preserved(&root),
        notes,
        plan_digest: String::new(),
    };
    plan.plan_digest = plan_digest(&plan);
    Ok(plan)
}

/// `github-apply`: publish the approved plan, then record baselines. A failure after a
/// file and before the manifest leaves files equal to the rendering, which the next plan
/// adopts as baselines (REQ-08B-07).
pub fn github_apply(
    root: &Path,
    p: GithubPlan,
    confirmed: &str,
    permission: &str,
) -> Result<Value> {
    if !crate::contract::permission_reference_ok(permission) {
        return Err("permission reference required".into());
    }
    let current = github_plan(root, p.request.clone())?;
    if p.protocol_version != 1
        || confirmed != p.plan_digest
        || plan_digest(&p) != p.plan_digest
        || current.plan_digest != p.plan_digest
    {
        return Err("stale or altered plan; re-preview required".into());
    }
    if let Some(f) = p.files.iter().find(|f| f.action == "conflict") {
        return Err(format!(
            "conflict in {}: {}; nothing applied",
            f.path,
            f.detail.as_deref().unwrap_or("review required")
        ));
    }
    if let Some(path) = unresolved(TEMPLATES).first() {
        return Err(format!(
            "conflict in {path}: unresolved template placeholder; nothing applied"
        ));
    }
    let root = root_of(root)?;
    let request_id = new_request_id("github.apply", &p.plan_digest);
    let mut receipt = Receipt::new(&request_id, "github.apply", &root);
    let mut written: Vec<String> = Vec::new();
    for f in p
        .files
        .iter()
        .filter(|f| f.action == "create" || f.action == "merge")
    {
        let text = TEMPLATES
            .iter()
            .find(|(t, _)| *t == f.path)
            .map(|(_, text)| *text)
            .ok_or("rendered output missing")?;
        let path = target(&root, &f.path)?;
        let before = read_opt(&path)?;
        let before_digest = before.as_deref().map(digest);
        let failed = |receipt: &mut Receipt, e: String| -> String {
            receipt.entry(
                &f.path,
                &f.action,
                before_digest.as_deref(),
                None,
                false,
                Some(e.clone()),
            );
            receipt.status = "partial".into();
            let saved = receipt.save(&root);
            let message = format!(
                "partial apply: applied [{}]; failed at {}: {e}",
                written.join(", "),
                f.path
            );
            match saved {
                Ok(()) => message,
                Err(se) => format!("{message}; receipt not recorded: {se}"),
            }
        };
        if before_digest != f.expected_sha256 {
            return Err(failed(&mut receipt, "target changed since plan".into()));
        }
        if let Some(b) = before.as_deref() {
            save_before_bytes(&root, &request_id, &f.path, b)
                .map_err(|e| failed(&mut receipt, e))?;
        }
        let publish = match before.as_deref() {
            None => publish_exclusive(&path, text.as_bytes()),
            Some(b) => publish_replace(&path, Some(b), text.as_bytes()),
        };
        publish.map_err(|e| failed(&mut receipt, e))?;
        receipt.entry(
            &f.path,
            &f.action,
            before_digest.as_deref(),
            Some(&f.rendered_sha256),
            true,
            None,
        );
        receipt.save(&root)?;
        written.push(f.path.clone());
    }
    // Baselines last: every planned file that is now equal to its rendering.
    let manifest_path = target(&root, MANIFEST)?;
    let before = read_opt(&manifest_path)?;
    let owned: Vec<Value> = p
        .files
        .iter()
        .map(|f| json!({"path": f.path, "baseline_sha256": f.rendered_sha256}))
        .collect();
    let mut manifest = json!({"schema_version": 1, "template_set": "github", "owned_files": owned});
    if p.files.is_empty() {
        manifest = match &before {
            Some(b) => serde_json::from_slice(b).map_err(|_| "github manifest unreadable")?,
            None => Value::Null,
        };
    }
    let mut adopted = Vec::new();
    if !manifest.is_null() {
        let mut out = serde_json::to_vec_pretty(&manifest).map_err(|_| "encode failed")?;
        out.push(b'\n');
        if before.as_deref() != Some(out.as_slice()) {
            if let Some(b) = before.as_deref() {
                save_before_bytes(&root, &request_id, MANIFEST, b)?;
            }
            match before.as_deref() {
                None => publish_exclusive(&manifest_path, &out),
                Some(b) => publish_replace(&manifest_path, Some(b), &out),
            }
            .map_err(|e| {
                format!(
                    "partial apply: applied [{}]; failed at {MANIFEST}: {e}",
                    written.join(", ")
                )
            })?;
            receipt.entry(
                MANIFEST,
                if before.is_none() { "create" } else { "merge" },
                before.as_deref().map(digest).as_deref(),
                Some(&digest(&out)),
                true,
                None,
            );
        }
        adopted = p
            .files
            .iter()
            .filter(|f| f.adopt_baseline)
            .map(|f| f.path.clone())
            .collect();
    }
    // A run that published nothing leaves no receipt behind (second run: diff 0).
    let request_id = if receipt.entries.is_empty() {
        None
    } else {
        receipt.status = "complete".into();
        receipt.save(&root)?;
        Some(request_id)
    };
    Ok(json!({
        "request_id": request_id,
        "written": written,
        "adopted_baselines": adopted,
        "unchanged": p.files.iter().filter(|f| f.action == "unchanged").map(|f| f.path.clone()).collect::<Vec<_>>(),
        "preserved": p.preserved,
        "manifest": MANIFEST,
        "rendering": "not_run: GitHub reads Issue forms only from the default branch",
    }))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn placeholder_scan_finds_unresolved_markers() {
        assert!(unresolved(TEMPLATES).is_empty());
        assert_eq!(
            unresolved(&[("a.yml", "name: {{project}}"), ("b.yml", "ok")]),
            vec!["a.yml".to_string()]
        );
    }
}
