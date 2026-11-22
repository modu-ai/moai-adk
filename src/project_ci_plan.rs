//! Plan and apply of the CI files (SPEC-20261122-project-ci-design REQ-6 to REQ-8): the
//! generated workflow text of `project_ci_workflows` is planned against what the project
//! already holds and published only on the user's pick.
//!
//! The plan is read-only. It reads the profile's saved answers and declared checks, renders
//! the files, compares each with the file on disk and with its recorded baseline (the same
//! 3-way engine `github-apply` uses: `decide`, `publish_exclusive`/`publish_replace`,
//! `save_before_bytes`, `Receipt`) and prints the estimate. The apply publishes exactly the
//! plan the user picked: its digest and a caller-supplied permission reference. A file the
//! project already has without a baseline, or that the user changed since the last apply, is a
//! conflict and is kept; other files under `.github/` are listed as preserved and never
//! touched. Baselines live in their own manifest, `manifests/ci.json` under the state folder,
//! so this operation adds no writer to the manifests of the other generators.

use crate::apply::{
    decide, new_request_id, publish_exclusive, publish_replace, read_opt, root_of,
    save_before_bytes, target, Decision, Receipt,
};
use crate::digest;
use crate::project_ci_estimate::Estimate;
use crate::spec::{Canonical, SortedKeys};
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use std::path::Path;

type Result<T> = std::result::Result<T, String>;

/// The CI manifest of `root`, under its state folder.
pub fn manifest_rel(root: &Path) -> String {
    crate::state_root::rel(root, "manifests/ci.json")
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Eq)]
#[serde(deny_unknown_fields)]
pub struct CiRequest {
    pub protocol_version: u32,
    /// Commits to the integration branch in the last 30 days, when the caller counted them;
    /// the estimate never uses fewer than the floor of `project_ci_estimate`.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub runs_per_month: Option<u32>,
    /// `ko` writes the comment block of each file in Korean; any other value writes English.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub language: Option<String>,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Eq)]
#[serde(deny_unknown_fields)]
pub struct CiFilePlan {
    pub path: String,
    /// create | merge | unchanged | conflict
    pub action: String,
    pub expected_sha256: Option<String>,
    pub baseline_sha256: Option<String>,
    pub rendered_sha256: String,
    /// True when the file already equals the rendering but no baseline is recorded (an
    /// interrupted apply): the next apply records it instead of reporting a conflict.
    pub adopt_baseline: bool,
    /// The jobs the file holds, as `name (runner)`, for the plain-words preview.
    pub jobs: Vec<String>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub detail: Option<String>,
}

#[derive(Clone, Debug, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct CiPlan {
    pub protocol_version: u32,
    pub request: CiRequest,
    pub files: Vec<CiFilePlan>,
    /// Existing files under `.github/` this plan does not own; they are left as they are.
    pub preserved: Vec<String>,
    /// Plain facts for the preview, in English; the interview relays them in the user's language.
    pub notes: Vec<String>,
    /// The monthly estimate; absent when nothing was planned.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub estimate: Option<Estimate>,
    pub plan_digest: String,
}

/// What the profile and the request render to: the files with their text, the notes and the
/// estimate. Nothing is read from the files the project already holds.
struct Rendered {
    files: Vec<(String, String)>,
    notes: Vec<String>,
    estimate: Option<Estimate>,
}

fn render(root: &Path, request: &CiRequest) -> Result<Rendered> {
    if request.protocol_version != 1 {
        return Err("unsupported protocol_version; supported: 1".into());
    }
    let (profile, _) = crate::profile::read_profile(root)?;
    let empty = |note: &str| Rendered {
        files: Vec::new(),
        notes: vec![note.to_string()],
        estimate: None,
    };
    let Some(answers) = &profile.ci else {
        return Ok(empty(
            "No CI answers are saved in the profile, so nothing was planned; ask the plain questions first.",
        ));
    };
    let Some(checks) = profile.checks.as_ref().filter(|c| !c.is_empty()) else {
        return Ok(empty(
            "No declared checks are saved in the profile, so nothing was planned; confirm the detected check commands first.",
        ));
    };
    let languages = profile
        .detected
        .as_ref()
        .map(|d| d.languages.clone())
        .unwrap_or_default();
    let generated = crate::project_ci_workflows::generate(&crate::project_ci_workflows::Input {
        flow: profile.work_flow.as_ref().map(|w| w.flow.as_str()),
        integration_branch: profile
            .work_flow
            .as_ref()
            .and_then(|w| w.integration_branch.as_deref()),
        answers,
        checks,
        languages: &languages,
        comments_in_korean: request.language.as_deref() == Some("ko"),
    })?;
    let mut notes = generated.notes.clone();
    let nightly = generated
        .files
        .iter()
        .any(|f| f.path.ends_with("moai-nightly.yml"));
    let has_build = checks.iter().any(|c| c.id == "build");
    let jobs = crate::project_ci_estimate::plan_jobs(
        answers,
        nightly,
        has_build,
        request.runs_per_month.unwrap_or(0),
    );
    let prices = crate::project_ci_estimate::prices();
    let estimate = crate::project_ci_estimate::estimate(&jobs, answers, &prices);
    let dollars = crate::project_ci_estimate::dollars;
    if estimate.free {
        notes.push(format!(
            "A public repository's standard runners are free: about {} billed minutes a month cost nothing; larger runners are billed. Prices read on {}.",
            estimate.minutes, estimate.prices_read_on
        ));
    } else {
        notes.push(format!(
            "Estimated cost: about {} dollars a month (range {} to {}), {} billed minutes, at prices read on {}. This is an estimate.",
            dollars(estimate.mills),
            dollars(estimate.low_mills),
            dollars(estimate.high_mills),
            estimate.minutes,
            estimate.prices_read_on
        ));
    }
    if estimate.over_budget {
        let mut cheaper: Vec<String> = Vec::new();
        for system in ["macos", "windows"] {
            if jobs.iter().any(|j| j.system == system) {
                cheaper.push(format!("drop the {system} build job"));
            }
        }
        if nightly {
            cheaper.push("drop the nightly run".into());
        }
        cheaper.push("run fewer checks".into());
        notes.push(format!(
            "The estimate is above your budget of {} dollars. Any of these makes a cheaper plan: {}.",
            answers.monthly_budget_usd.unwrap_or(0),
            cheaper.join(", ")
        ));
    }
    if !estimate.free && answers.monthly_budget_usd.is_some() {
        notes.push(
            "This tool cannot set a budget. To create one open Billing, Budgets and alerts, New budget, and turn on Stop usage when budget limit is reached.".into(),
        );
    }
    Ok(Rendered {
        files: generated
            .files
            .into_iter()
            .map(|f| (f.path, f.text))
            .collect(),
        notes,
        estimate: Some(estimate),
    })
}

/// The jobs a workflow file holds, as `name (runner)`.
fn jobs_of(text: &str) -> Vec<String> {
    let mut jobs = Vec::new();
    let mut inside = false;
    let mut name: Option<String> = None;
    for line in text.lines() {
        if line == "jobs:" {
            inside = true;
        } else if inside
            && line.starts_with("  ")
            && !line.starts_with("   ")
            && line.ends_with(':')
        {
            name = Some(line.trim().trim_end_matches(':').to_string());
        } else if let (true, Some(n), Some(runner)) =
            (inside, &name, line.strip_prefix("    runs-on: "))
        {
            jobs.push(format!("{n} ({runner})"));
            name = None;
        }
    }
    jobs
}

fn manifest_baselines(root: &Path) -> Result<Vec<(String, String)>> {
    let Some(bytes) = read_opt(&root.join(manifest_rel(root)))? else {
        return Ok(Vec::new());
    };
    let v: Value = serde_json::from_slice(&bytes)
        .map_err(|_| "ci manifest unreadable; inspect before generating".to_string())?;
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

/// Files under `.github/workflows` and the Dependabot file that the plan does not own.
fn preserved(root: &Path, owned: &[String]) -> Vec<String> {
    let real_dir = |p: &Path| std::fs::symlink_metadata(p).is_ok_and(|m| m.is_dir());
    let mut out = Vec::new();
    let dir = root.join(".github").join("workflows");
    if real_dir(&root.join(".github")) && real_dir(&dir) {
        if let Ok(entries) = std::fs::read_dir(&dir) {
            for e in entries.flatten() {
                if e.file_type().is_ok_and(|t| t.is_file()) {
                    if let Some(name) = e.file_name().to_str() {
                        out.push(format!(".github/workflows/{name}"));
                    }
                }
            }
        }
    }
    let dependabot = root.join(".github").join("dependabot.yml");
    if std::fs::symlink_metadata(&dependabot).is_ok_and(|m| m.is_file()) {
        out.push(".github/dependabot.yml".into());
    }
    out.retain(|p| !owned.contains(p));
    out.sort();
    out
}

fn plan_digest(p: &CiPlan) -> String {
    let mut v = serde_json::to_value(p).expect("plan serializes");
    v["plan_digest"] = Value::Null;
    digest(&SortedKeys.bytes(&v))
}

/// `project-ci-plan`: read-only.
pub fn ci_plan(root: &Path, request: CiRequest) -> Result<CiPlan> {
    let root = root_of(root)?;
    let rendered = render(&root, &request)?;
    let baselines = manifest_baselines(&root)?;
    let mut files = Vec::new();
    for (rel, text) in &rendered.files {
        let current = read_opt(&target(&root, rel)?)?;
        let current_digest = current.as_deref().map(digest);
        let baseline = baselines
            .iter()
            .find(|(p, _)| p == rel)
            .map(|(_, d)| d.clone());
        let rendered_digest = digest(text.as_bytes());
        let (action, detail) = match decide(
            baseline.as_deref(),
            current_digest.as_deref(),
            Some(&rendered_digest),
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
        files.push(CiFilePlan {
            path: rel.clone(),
            action: action.into(),
            expected_sha256: current_digest,
            baseline_sha256: baseline,
            rendered_sha256: rendered_digest,
            adopt_baseline: adopt,
            jobs: jobs_of(text),
            detail,
        });
    }
    let owned: Vec<String> = files.iter().map(|f| f.path.clone()).collect();
    let mut plan = CiPlan {
        protocol_version: 1,
        request,
        files,
        preserved: preserved(&root, &owned),
        notes: rendered.notes,
        estimate: rendered.estimate,
        plan_digest: String::new(),
    };
    plan.plan_digest = plan_digest(&plan);
    Ok(plan)
}

/// `project-ci-apply`: publish the approved plan, then record baselines. A failure after a file
/// and before the manifest leaves files equal to the rendering, which the next plan adopts as
/// baselines.
pub fn ci_apply(root: &Path, p: CiPlan, confirmed: &str, permission: &str) -> Result<Value> {
    if !crate::contract::permission_reference_ok(permission) {
        return Err("permission reference required".into());
    }
    let current = ci_plan(root, p.request.clone())?;
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
    let root = root_of(root)?;
    if p.files.is_empty() {
        return Ok(json!({
            "request_id": null, "written": [], "adopted_baselines": [], "unchanged": [],
            "preserved": p.preserved, "manifest": manifest_rel(&root),
        }));
    }
    let rendered = render(&root, &p.request)?;
    let request_id = new_request_id("ci.apply", &p.plan_digest);
    let mut receipt = Receipt::new(&request_id, "ci.apply", &root);
    let mut written: Vec<String> = Vec::new();
    for f in p
        .files
        .iter()
        .filter(|f| f.action == "create" || f.action == "merge")
    {
        let text = rendered
            .files
            .iter()
            .find(|(path, _)| *path == f.path)
            .map(|(_, text)| text.as_str())
            .filter(|text| digest(text.as_bytes()) == f.rendered_sha256)
            .ok_or("rendered output missing or changed")?;
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
    // Baselines last: every planned file is now equal to its rendering.
    let manifest_rel = manifest_rel(&root);
    let manifest_path = target(&root, &manifest_rel)?;
    let before = read_opt(&manifest_path)?;
    let owned: Vec<Value> = p
        .files
        .iter()
        .map(|f| json!({"path": f.path, "baseline_sha256": f.rendered_sha256}))
        .collect();
    let manifest = json!({"schema_version": 1, "template_set": "ci", "owned_files": owned});
    let mut out = serde_json::to_vec_pretty(&manifest).map_err(|_| "encode failed")?;
    out.push(b'\n');
    if before.as_deref() != Some(out.as_slice()) {
        if let Some(b) = before.as_deref() {
            save_before_bytes(&root, &request_id, &manifest_rel, b)?;
        }
        match before.as_deref() {
            None => publish_exclusive(&manifest_path, &out),
            Some(b) => publish_replace(&manifest_path, Some(b), &out),
        }
        .map_err(|e| {
            format!(
                "partial apply: applied [{}]; failed at {manifest_rel}: {e}",
                written.join(", ")
            )
        })?;
        receipt.entry(
            &manifest_rel,
            if before.is_none() { "create" } else { "merge" },
            before.as_deref().map(digest).as_deref(),
            Some(&digest(&out)),
            true,
            None,
        );
    }
    let adopted: Vec<String> = p
        .files
        .iter()
        .filter(|f| f.adopt_baseline)
        .map(|f| f.path.clone())
        .collect();
    // A run that published nothing leaves no receipt behind (the second run changes nothing).
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
        "manifest": manifest_rel,
        "rendering": "not_run: GitHub runs a workflow only from the default branch and the first run is the real test",
    }))
}
