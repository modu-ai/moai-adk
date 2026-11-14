//! `project-update-plan` (#2936 S1): one read-only preview of the project update, grouped into
//! the three phases of the user's UPDATE flow (#2936 comment 6876720048): `backup` (the
//! Go-generation state moves aside), `install` (this generation's managed files), `merge`
//! (asking which backed-up settings and instructions come back; a later slice).
//!
//! Every step calls an existing plan in this process — `go_state::plan`, `profile::plan`,
//! `generation::generation_plan`, `claude::settings_plan_in`, `github_templates::github_plan`
//! — and prints the apply command that plan's own writer takes, as a word list starting with
//! `moai` whose one placeholder is `<permission-reference>`. Nothing is written here. One
//! step's error makes that step `blocked` and the others still run.

use crate::claude::{self, SettingsScope};
use crate::generation::{self, GenerationRequest};
use crate::github_templates::{self, GithubRequest};
use crate::{go_state, profile, project_detect, state_root, Profile};
use serde_json::{json, Map, Value};
use std::{fs, path::Path};

type Result<T> = std::result::Result<T, String>;

/// The one placeholder a printed command carries.
pub const PERMISSION_PLACEHOLDER: &str = "<permission-reference>";
/// Why an install step waits when the project has no profile (Q5: never created here).
pub const NO_PROFILE: &str = "run /moai project";
/// Why every install step waits while the backup phase has a move to make: its preview
/// would be stale after the move.
pub const AFTER_BACKUP: &str = "re-run after the backup phase";
/// Why the generation step waits while the profile step is `ready`: the generation plan
/// binds the profile bytes, so its digest is stale once the profile changes.
pub const AFTER_PROFILE: &str = "re-run after the profile apply";

const BACKUP_REASON: &str = "the Go-generation .moai/ and the earlier .moai-cli/ state \
     (go-state-plan); .agents, .codex, settings and instruction files are not covered yet \
     (a later slice, #2936 S2)";
const INSTALL_REASON: &str = "profile, generation (AGENTS.md managed section and generated \
     files), Claude Code settings, GitHub templates; host bundles for .codex, .agents and .omp \
     are a later slice (#2936 S3)";
const MERGE_REASON: &str = "merge candidates are a later slice (#2936 S1f)";

/// Step states from least to most severe; a phase takes its worst step's state.
/// `skip_in_use` (a file a running session reads, #2936 S2b) is reserved: this slice never
/// emits it.
const SEVERITY: [&str; 6] = [
    "unchanged",
    "skip",
    "ready",
    "needs_decision",
    "conflict",
    "blocked",
];

/// The generation manifest kinds whose request is not recorded, so a plan from the stored
/// state cannot rebuild them (Q1 A, #2936).
const UNREBUILDABLE: [&str; 3] = [
    "generation:native-role:",
    "generation:project-skill:",
    "generation:skill-projection:",
];

struct Step {
    name: &'static str,
    state: &'static str,
    reason: String,
    preview: Value,
    plan_digest: Option<String>,
    /// The printed command and its stdin; only a step that can be applied carries one.
    apply: Option<(Vec<String>, Value)>,
}

impl Step {
    fn waiting(name: &'static str, state: &'static str, reason: &str) -> Step {
        Step {
            name,
            state,
            reason: reason.into(),
            preview: Value::Null,
            plan_digest: None,
            apply: None,
        }
    }

    fn to_json(&self) -> Value {
        json!({
            "step": self.name,
            "state": self.state,
            "reason": self.reason,
            "preview": self.preview,
            "plan_digest": self.plan_digest,
            "apply": self.apply.as_ref().map(|(argv, stdin)| json!({"argv": argv, "stdin": stdin})),
        })
    }
}

/// A step's error becomes that step's `blocked` state; the plan goes on.
fn or_blocked(name: &'static str, r: Result<Step>) -> Step {
    r.unwrap_or_else(|e| Step::waiting(name, "blocked", &e))
}

fn worst(steps: &[Step]) -> &'static str {
    steps
        .iter()
        .map(|s| s.state)
        .max_by_key(|s| SEVERITY.iter().position(|x| x == s))
        .unwrap_or("unchanged")
}

/// `moai <op> <project> --plan-digest <digest> --permission-reference <permission-reference>`.
fn argv(op: &str, project: &str, digest: &str) -> Vec<String> {
    [
        "moai",
        op,
        project,
        "--plan-digest",
        digest,
        "--permission-reference",
        PERMISSION_PLACEHOLDER,
    ]
    .map(String::from)
    .to_vec()
}

fn encode<T: serde::Serialize>(v: &T) -> Result<Value> {
    serde_json::to_value(v).map_err(|_| "encode failed".to_string())
}

fn go_state_step(root: &Path, shown: &str) -> Result<Step> {
    let plan = go_state::plan(root)?;
    let pending = plan.items.iter().any(|i| i.action == "move")
        || plan.state_moves.iter().any(|m| m.action == "move")
        || plan.root_files.iter().any(|f| f.action == "copy")
        || plan.restores.iter().any(|r| r.action == "copy");
    let preview = json!({
        "items": plan.items,
        "go_backup": plan.go_backup,
        "state_moves": plan.state_moves,
        "restores": plan.restores,
        "root_files": plan.root_files,
        "unrecorded": plan.unrecorded,
        "warnings": plan.warnings,
    });
    Ok(if pending {
        Step {
            name: "go-state",
            state: "needs_decision",
            reason: "the earlier generation's state moves aside; ask the user before this move"
                .into(),
            preview,
            apply: Some((
                argv("go-state-apply", shown, &plan.plan_digest),
                Value::Null,
            )),
            plan_digest: Some(plan.plan_digest),
        }
    } else {
        Step {
            name: "go-state",
            state: "unchanged",
            reason: "nothing to move".into(),
            preview,
            plan_digest: Some(plan.plan_digest),
            apply: None,
        }
    })
}

fn profile_step(
    root: &Path,
    shown: &str,
    profile: &Profile,
    detected: &Result<project_detect::Detected>,
) -> Result<Step> {
    let mut desired = profile.clone();
    desired.schema_version = 2;
    desired.detected = Some(detected.clone()?);
    let plan = profile::plan(root, desired)?;
    let (state, reason) = match plan.action.as_str() {
        "create" | "update" => ("ready", "the detected facts changed"),
        "unchanged" => ("unchanged", "the profile already has the detected facts"),
        _ => (
            "conflict",
            "a key changed by hand since the last apply; it is kept",
        ),
    };
    let preview = json!({"action": plan.action, "delta": plan.delta, "drift": plan.drift});
    let apply = (state == "ready")
        .then(|| {
            Ok::<_, String>((
                argv("project-apply", shown, &plan.plan_digest),
                encode(&plan)?,
            ))
        })
        .transpose()?;
    Ok(Step {
        name: "profile",
        state,
        reason: reason.into(),
        preview,
        plan_digest: Some(plan.plan_digest),
        apply,
    })
}

/// The generation request rebuilt from the profile and the manifest's record (#2936 S1).
fn generation_request(root: &Path, profile: &Profile, manifest: &Value) -> GenerationRequest {
    let language = profile
        .language
        .as_ref()
        .and_then(|l| l.get("instruction_files"))
        .filter(|code| *code == "en" || *code == "ko")
        .cloned()
        .unwrap_or_else(|| "en".into());
    let workflow_reference = manifest["workflow_reference"]
        .as_str()
        .map(String::from)
        .unwrap_or_else(|| format!("{}/workflows/project/", state_root::dir(root)));
    let workflow_id = manifest["workflow_id"]
        .as_str()
        .map(String::from)
        .unwrap_or_else(|| "project".into());
    // The manifest records the personal-file choice only through the `.gitignore` section it
    // owns (`ignored`); anything else plans `none`, which never creates or touches the file.
    let ignored =
        owned_entries(manifest).any(|(section, _)| section == crate::agents::IGNORE_SECTION);
    GenerationRequest {
        protocol_version: 1,
        language,
        workflow_reference,
        local_file: if ignored { "ignored" } else { "none" }.into(),
        workflow_id,
        native_roles: Vec::new(),
        skill_projections: Vec::new(),
        project_skills: Vec::new(),
    }
}

/// `(section_id, path)` of every owned manifest entry that has a section.
fn owned_entries(manifest: &Value) -> impl Iterator<Item = (&str, &str)> {
    manifest["owned_files"]
        .as_array()
        .into_iter()
        .flatten()
        .filter_map(|f| Some((f["section_id"].as_str()?, f["path"].as_str()?)))
}

fn generation_step(root: &Path, shown: &str, profile: &Profile) -> Result<Step> {
    let path = state_root::path(root, "manifests/generation.json");
    let manifest: Value = match crate::apply::read_opt(&path)? {
        None => Value::Null,
        Some(bytes) => serde_json::from_slice(&bytes)
            .map_err(|_| "generation manifest unreadable; inspect before regenerating")?,
    };
    // Q1 A (#2936): until the manifest records the request fields of project agents and
    // skills, a plan from the stored state would drop them (a project skill plans `remove`).
    // The user decides instead; nothing is planned.
    let unrebuildable: Vec<&str> = owned_entries(&manifest)
        .filter(|(section, _)| UNREBUILDABLE.iter().any(|k| section.starts_with(k)))
        .map(|(_, path)| path)
        .collect();
    if !unrebuildable.is_empty() {
        return Ok(Step {
            name: "generation",
            state: "needs_decision",
            reason: "the generation manifest owns project agent or skill files whose request is \
                     not recorded; regenerate them through /moai project"
                .into(),
            preview: json!({"not_rebuilt": unrebuildable}),
            plan_digest: None,
            apply: None,
        });
    }
    let request = generation_request(root, profile, &manifest);
    let plan = generation::generation_plan(root, request)?;
    let files: Vec<Value> = plan
        .files
        .iter()
        .map(|f| json!({"path": f.path, "action": f.action, "detail": f.detail}))
        .collect();
    let (state, reason) = if plan.files.iter().any(|f| f.action == "conflict") {
        ("conflict", "a managed part was edited by hand; it is kept")
    } else if plan.files.iter().any(|f| f.action != "unchanged") {
        ("ready", "managed files to write")
    } else {
        ("unchanged", "every managed file is current")
    };
    let apply = (state == "ready")
        .then(|| {
            Ok::<_, String>((
                argv("generation-apply", shown, &plan.plan_digest),
                encode(&plan)?,
            ))
        })
        .transpose()?;
    Ok(Step {
        name: "generation",
        state,
        reason: reason.into(),
        preview: json!({"files": files, "warnings": plan.warnings, "host_notes": plan.host_notes}),
        plan_digest: Some(plan.plan_digest),
        apply,
    })
}

/// Q2 A (#2936): the keys MoAI already owns, without `hooks` (only `hooks-apply` writes
/// that key), are kept as they are; no key is added by default.
fn settings_step(root: &Path, shown: &str) -> Result<Step> {
    let mut owned: Map<String, Value> = claude::owned_settings(root)?;
    owned.remove("hooks");
    let desired = claude::user_settings_desired_in(root, &owned, SettingsScope::Project)?;
    let plan = claude::settings_plan_in(root, &desired, SettingsScope::Project)?;
    let digest = plan["plan_digest"].as_str().map(String::from);
    let (state, reason) = match plan["action"].as_str() {
        Some("create" | "update") => ("ready", "MoAI-owned keys to restore"),
        Some("unchanged") => ("unchanged", "the MoAI-owned keys are current"),
        _ => (
            "conflict",
            "a key MoAI manages was set by the user; nothing is written",
        ),
    };
    let apply = match (&digest, state) {
        (Some(d), "ready") => Some((
            argv("claude-settings-apply", shown, d),
            Value::Object(owned),
        )),
        _ => None,
    };
    Ok(Step {
        name: "claude-settings",
        state,
        reason: reason.into(),
        preview: json!({
            "action": plan["action"],
            "settings_path": plan["settings_path"],
            "conflicts": plan["conflicts"],
            "confirm_replace": plan.get("confirm_replace"),
        }),
        plan_digest: digest,
        apply,
    })
}

fn github_step(root: &Path, shown: &str) -> Result<Step> {
    let plan = github_templates::github_plan(
        root,
        GithubRequest {
            protocol_version: 1,
            github_issues: None,
            language: None,
        },
    )?;
    let files: Vec<Value> = plan
        .files
        .iter()
        .map(|f| json!({"path": f.path, "action": f.action, "detail": f.detail}))
        .collect();
    let (state, reason) = if plan.files.is_empty() {
        ("skip", "the project did not opt in to GitHub templates")
    } else if plan.files.iter().any(|f| f.action == "conflict") {
        ("conflict", "a template was edited by hand; it is kept")
    } else if plan
        .files
        .iter()
        .any(|f| f.action != "unchanged" || f.adopt_baseline)
    {
        ("ready", "templates to write or record")
    } else {
        ("unchanged", "every template is current")
    };
    let apply = (state == "ready")
        .then(|| {
            Ok::<_, String>((
                argv("github-apply", shown, &plan.plan_digest),
                encode(&plan)?,
            ))
        })
        .transpose()?;
    Ok(Step {
        name: "github",
        state,
        reason: reason.into(),
        preview: json!({"files": files, "notes": plan.notes}),
        plan_digest: Some(plan.plan_digest),
        apply,
    })
}

const INSTALL: [&str; 4] = ["profile", "generation", "claude-settings", "github"];

fn install_steps(
    root: &Path,
    shown: &str,
    detected: &Result<project_detect::Detected>,
) -> Vec<Step> {
    let profile = match crate::read_profile(root) {
        Ok((p, _)) => p,
        Err(e) if e == "project profile missing" => {
            return INSTALL
                .map(|n| Step::waiting(n, "needs_decision", NO_PROFILE))
                .into_iter()
                .collect();
        }
        Err(e) => {
            let mut steps = vec![Step::waiting("profile", "blocked", &e)];
            steps.push(Step::waiting("generation", "blocked", &e));
            steps.push(or_blocked("claude-settings", settings_step(root, shown)));
            steps.push(or_blocked("github", github_step(root, shown)));
            return steps;
        }
    };
    let profile_step = or_blocked("profile", profile_step(root, shown, &profile, detected));
    let generation = if profile_step.state == "ready" {
        Step::waiting("generation", "blocked", AFTER_PROFILE)
    } else {
        or_blocked("generation", generation_step(root, shown, &profile))
    };
    vec![
        profile_step,
        generation,
        or_blocked("claude-settings", settings_step(root, shown)),
        or_blocked("github", github_step(root, shown)),
    ]
}

/// The whole preview. Writes nothing.
pub fn plan(project: &Path) -> Result<Value> {
    let root = fs::canonicalize(project).map_err(|_| "project folder unavailable")?;
    let shown = root.to_str().ok_or("non UTF-8 project path")?.to_string();
    let detected = project_detect::detect(&root);
    let backup = vec![or_blocked("go-state", go_state_step(&root, &shown))];
    // A move still to make: every install preview would be stale after it.
    let install = if backup[0].state == "needs_decision" {
        INSTALL
            .map(|n| Step::waiting(n, "blocked", AFTER_BACKUP))
            .into_iter()
            .collect()
    } else {
        install_steps(&root, &shown, &detected)
    };
    let all_unchanged = backup
        .iter()
        .chain(&install)
        .all(|s| s.state == "unchanged" || s.state == "skip");
    let phase = |name: &str, reason: &str, steps: &[Step]| {
        json!({
            "phase": name,
            "state": worst(steps),
            "reason": reason,
            "steps": steps.iter().map(Step::to_json).collect::<Vec<_>>(),
        })
    };
    Ok(json!({
        "project": shown,
        "detected": detected.ok(),
        "state_root": state_root::classify(&root).to_json(),
        "phases": [
            phase("backup", BACKUP_REASON, &backup),
            phase("install", INSTALL_REASON, &install),
            // The merge phase (asking which backed-up settings and instructions come back)
            // is a later slice; it is never counted in `all_unchanged`.
            {"phase": "merge", "state": "not_run", "reason": MERGE_REASON, "candidates": []},
        ],
        "all_unchanged": all_unchanged,
    }))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn a_phase_takes_its_worst_step() {
        let s = |state| Step::waiting("x", state, "");
        assert_eq!(worst(&[s("unchanged"), s("skip")]), "skip");
        assert_eq!(worst(&[s("ready"), s("blocked"), s("conflict")]), "blocked");
        assert_eq!(worst(&[s("needs_decision"), s("ready")]), "needs_decision");
        assert_eq!(worst(&[]), "unchanged");
    }

    #[test]
    fn printed_commands_carry_one_placeholder() {
        let a = argv("go-state-apply", "/p", "sha256:x");
        assert_eq!(a[0], "moai");
        assert_eq!(
            a.iter().filter(|w| w.starts_with('<')).collect::<Vec<_>>(),
            [PERMISSION_PLACEHOLDER]
        );
    }
}
