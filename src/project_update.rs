//! `project-update-plan` (#2936 S1): one read-only preview of the project update, grouped into
//! the three phases of the user's UPDATE flow (#2936 comment 6876720048): `backup` (the
//! Go-generation state moves aside), `install` (this generation's managed files), `merge`
//! (the text of `AGENTS.local.md` comes back into `AGENTS.md`, #2936 R10; the user's earlier
//! `settings.json` values come back into the new one, #2936 R8; the other backed-up
//! instructions are a later slice).
//!
//! Every step calls an existing plan in this process — `go_state::plan`, `profile::plan`,
//! `generation::generation_plan`, `claude::settings_plan_in`, `github_templates::github_plan`
//! — and prints the apply command that plan's own writer takes, as a word list starting with
//! `moai` whose one placeholder is `<permission-reference>`. Nothing is written here. One
//! step's error makes that step `blocked` and the others still run.

use crate::agents_local_merge;
use crate::claude::{self, SettingsScope};
use crate::claude_settings_merge;
use crate::generation::{self, GenerationRequest};
use crate::github_templates::{self, GithubRequest};
use crate::{
    go_settings_clean, go_state, home_workdir, profile, project_detect, project_hosts,
    reports_move, state_root, status_line, Profile,
};
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
/// Why the hooks step waits while the settings step has a write to make (#3028 audit B1).
pub const AFTER_SETTINGS: &str = "re-run after the claude-settings apply";
/// Why the go-state step waits while the go-settings step has a write to make (#2936 L2a): that
/// write changes what the go-state plan reads from `.claude/settings.json`, so a digest made
/// before it would be stale.
pub const AFTER_GO_SETTINGS: &str = "re-run after the go-settings apply";
/// Why the go-state step waits while the reports step has files to move (C-126): that move
/// changes what the go-state plan lists under the report folders, so a digest made before it
/// would be stale.
pub const AFTER_REPORTS: &str = "re-run after the reports apply";

const BACKUP_REASON: &str = "the Go-generation .moai/ and the earlier .moai-cli/ state, the \
     Go tool files under .claude/ and .agents/, settings and instruction files \
     (go-state-plan); .codex files are classified in the preview and nothing under .codex \
     moves yet (#2936 S1c-2)";
const INSTALL_REASON: &str = "profile, generation (AGENTS.md managed section and generated \
     files), Claude Code settings, Claude Code hooks, GitHub templates, the Codex and OMP hook \
     sets; the Codex config and skill copy and the OMP agents are a later slice (#2936 R9 S5b)";
/// Why the merge step waits while the generation step has files to write: the merge reads and
/// writes `AGENTS.md`, whose bytes that step's digest binds.
pub const AFTER_INSTALL: &str = "re-run after the install phase";
const MERGE_REASON: &str = "the text of AGENTS.local.md merges into AGENTS.md (#2936 R10); the \
     user's earlier settings.json values merge back into the new one (#2936 R8); other \
     backed-up instructions are a later slice (#2936 S1f)";

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

/// #2936 L2a: the `go-settings` step, the backup phase's first. The Go tool's `.moai/` references
/// in `.claude/settings.json` (a status line that runs `.moai/status_line.sh`, `plansDirectory`)
/// hold the Go `.moai/` folder in place; `go-settings-apply` removes exactly the values the Go
/// tool wrote, after backing the file up and showing the Rust launcher is the Rust build.
///
/// - `ready`: Go-made keys to clean; the printed command is `go-settings-apply`.
/// - `unchanged`: no Go-made key (values the user changed are named under `kept` and stay).
/// - `skip`: the Go manifest proves the whole file the earlier generation's, so it moves with
///   the go-state step into the backup and its keys are not cleaned apart.
/// - `needs_decision`: held, the Rust launcher could not be shown; nothing is written and the
///   reason says what to check. `blocked`: the plan cannot be made.
fn go_settings_step(
    root: &Path,
    shown: &str,
    go_state: Option<&go_state::GoStatePlan>,
) -> Result<Step> {
    let home = std::env::var_os("HOME").map(std::path::PathBuf::from);
    let plan = go_settings_clean::plan(root, home.as_deref())?;
    let digest = plan["plan_digest"].as_str().map(String::from);
    let moves_itself = go_state.is_some_and(|p| {
        p.items
            .iter()
            .any(|i| i.path == ".claude/settings.json" && i.action == "move")
    });
    let (state, reason) = match plan["action"].as_str() {
        Some("replace") if moves_itself => (
            "skip",
            "the Go manifest proves .claude/settings.json the earlier generation's: it moves \
             into the backup with the go-state step, so its keys are not cleaned apart"
                .to_string(),
        ),
        Some("replace") => (
            "ready",
            "Go-made .moai/ references in .claude/settings.json to clean (a backup is kept)"
                .to_string(),
        ),
        Some("none") => (
            "unchanged",
            "no value the Go tool wrote names .moai/ in .claude/settings.json".to_string(),
        ),
        Some("held") => (
            "needs_decision",
            format!(
                "{}; {}",
                plan["reason"]
                    .as_str()
                    .unwrap_or("the Rust launcher is not shown"),
                plan["next"].as_str().unwrap_or("nothing was written")
            ),
        ),
        _ => (
            "blocked",
            format!("the clean-up is blocked: {}", plan["blocks"]),
        ),
    };
    let apply = match (&digest, state) {
        (Some(d), "ready") => Some((argv("go-settings-apply", shown, d), Value::Null)),
        _ => None,
    };
    Ok(Step {
        name: "go-settings",
        state,
        reason,
        preview: json!({
            "action": plan["action"],
            "go_keys": plan["go_keys"],
            "kept": plan["kept"],
            "kept_note": plan.get("kept_note"),
            "launcher": plan["launcher"],
            "plans_files": plan["plans_files"],
            "local_settings": plan["local_settings"],
            "blocks": plan["blocks"],
        }),
        plan_digest: digest,
        apply,
    })
}

/// C-126: the `reports` step, the backup phase's last. The report files an earlier generation
/// left under `.moai-cli/reports` and `.moai/reports` that git does not track move to the
/// user-level reports folder; `reports-move-apply` takes the plan back on stdin and makes it
/// again, so a file added, moved or tracked since the preview makes it stale.
///
/// - `ready`: files to move; the printed command is `reports-move-apply`.
/// - `unchanged`: nothing untracked is left in the earlier folders.
/// - `blocked`: the project is not a git repository or git cannot tell what it tracks.
fn reports_step(root: &Path, shown: &str) -> Result<Step> {
    let home = home_workdir::home()?;
    let plan = reports_move::plan(root, &home)?;
    let state = match plan.state.as_str() {
        "ready" => "ready",
        "unchanged" => "unchanged",
        _ => "blocked",
    };
    let reason = match state {
        "ready" => format!(
            "{} untracked report file(s) under the earlier report folders move to {} \
             (tracked files stay; a name already there is kept)",
            plan.moves.len(),
            plan.destination
        ),
        "unchanged" => "no untracked report file is left in the earlier report folders".to_string(),
        _ => plan.reason.clone(),
    };
    let apply = match state {
        "ready" => Some((
            argv("reports-move-apply", shown, &plan.plan_digest),
            encode(&plan)?,
        )),
        _ => None,
    };
    Ok(Step {
        name: "reports",
        state,
        reason,
        preview: plan.preview(),
        plan_digest: Some(plan.plan_digest.clone()),
        apply,
    })
}

/// The key names a `go-settings` step cleans, for another step's reason.
fn go_setting_names(step: &Step) -> String {
    step.preview["go_keys"]
        .as_array()
        .map(|a| {
            a.iter()
                .filter_map(|k| k["key"].as_str())
                .collect::<Vec<_>>()
                .join(", ")
        })
        .unwrap_or_default()
}

fn go_state_step(
    root: &Path,
    shown: &str,
    plan: &go_state::GoStatePlan,
    settings: &Step,
) -> Result<Step> {
    let pending = plan.items.iter().any(|i| i.action == "move")
        || plan.state_moves.iter().any(|m| m.action == "move")
        || plan.root_files.iter().any(|f| f.action != "leave")
        || plan.restores.iter().any(|r| r.action == "copy");
    // #2936 S1b: how many items fall in each class, so the user sees the shape at a glance.
    let mut classes = std::collections::BTreeMap::<&str, u64>::new();
    for i in &plan.items {
        *classes.entry(i.class.as_str()).or_insert(0) += 1;
    }
    let mut preview = json!({
        "classes": classes,
        "kinds": plan.kinds,
        "counterparts": plan.counterparts,
        "after_move": plan.after_move,
        "secrets": plan.secrets,
        "items": plan.items,
        "layout": plan.layout,
        "go_backup": plan.go_backup,
        "backup_to": plan.backup_to,
        "state_moves": plan.state_moves,
        "restores": plan.restores,
        "root_files": plan.root_files,
        "unrecorded": plan.unrecorded,
        "warnings": plan.warnings,
    });
    // #2936 S-CYCLE: which settings move and what that frees; absent when the rule does not fire.
    if let Some(sm) = &plan.settings_move {
        preview["settings_move"] = encode(sm)?;
    }
    // #2936 S-AGENTS-STAY U1: the backup area's own `.gitignore` leaves the copies to git (or git
    // cannot tell). The apply would refuse before writing, so no command is offered: the user
    // fixes that file (the update never edits it) and previews again. The install steps wait
    // behind this step like behind any backup step that is still to do.
    if let Some(b) = &plan.backup_ignore {
        preview["backup_ignore"] = encode(b)?;
        let named = if b.examples.is_empty() {
            String::new()
        } else {
            format!(" (for example {})", b.examples.join(", "))
        };
        return Ok(Step {
            name: "go-state",
            state: "needs_decision",
            reason: format!(
                "the earlier generation's state would move into a backup area git may offer for \
                 commit{named}: {}",
                b.reason
            ),
            preview,
            apply: None,
            plan_digest: Some(plan.plan_digest.clone()),
        });
    }
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
            plan_digest: Some(plan.plan_digest.clone()),
        }
    } else if let Some((state, reason)) = tied(root, plan, settings) {
        Step {
            name: "go-state",
            state,
            reason,
            preview,
            plan_digest: Some(plan.plan_digest.clone()),
            apply: None,
        }
    } else {
        Step {
            name: "go-state",
            state: "unchanged",
            reason: "nothing to move".into(),
            preview,
            plan_digest: Some(plan.plan_digest.clone()),
            apply: None,
        }
    })
}

/// Earlier state that stays where it is although nothing moves (#2953 F1): `go-state-plan`
/// leaves everything for a `state_root` verdict other than `go`/`none`, and for Go or
/// `.moai-cli/` state it cannot move (settings naming `.moai/`, a worktree or nested checkout
/// inside, no usable Go manifest). The user decides; `interrupted` waits for their revert.
fn tied(
    root: &Path,
    plan: &go_state::GoStatePlan,
    settings: &Step,
) -> Option<(&'static str, String)> {
    let state = state_root::classify(root);
    let held = |path: &str| {
        plan.items
            .iter()
            .find(|i| i.path == path && i.action == "leave")
            .map(|i| i.reason.clone())
    };
    match state.generation {
        "interrupted" => Some((
            "needs_decision",
            format!(
                "state_root `interrupted`: {}; show go-state-revert-plan and move nothing else \
                 until the user decides",
                state.reason
            ),
        )),
        "mixed" | "unknown" | "refuse" => Some((
            "blocked",
            format!(
                "state_root `{}`: {}; nothing moves, the user decides outside this route",
                state.generation, state.reason
            ),
        )),
        "go" => {
            let mut why = held(state_root::MOAI_DIR)
                .unwrap_or_else(|| "no usable Go manifest, so nothing is proven tool-made".into());
            // #2936 L1: name the file and keys that hold the folder, and what to do next.
            // #2936 L2a: the go-settings step above cleans the values the Go tool wrote.
            if why.starts_with("settings reference .moai/") {
                let refs = moai_dir_references(root);
                if !refs.is_empty() {
                    // Values the step leaves: the ones it lists as kept. The local file is never
                    // written (#2936 L2a, D1 condition 3); its keys get a sentence of their own
                    // (#2936 L2b).
                    let yours = settings.preview["kept"]
                        .as_array()
                        .is_some_and(|k| !k.is_empty());
                    let local_keys = refs
                        .iter()
                        .find(|(file, _)| *file == LOCAL_SETTINGS)
                        .map(|(_, keys)| keys.join(", "));
                    let shared_keys = refs.iter().any(|(file, _)| *file != LOCAL_SETTINGS);
                    let mut next = match settings.state {
                        "ready" => format!(
                            "the go-settings step above clears the Go-made keys ({}); apply it \
                             and preview again{}",
                            go_setting_names(settings),
                            if yours {
                                "; the keys it lists as kept are your values: change or remove \
                                 them yourself"
                            } else {
                                ""
                            }
                        ),
                        "needs_decision" => format!(
                            "the go-settings step above could not clear the Go-made keys ({})",
                            settings.reason
                        ),
                        _ if shared_keys => "this update does not change those keys, so change \
                                             or remove them yourself and run the update again"
                            .to_string(),
                        _ => String::new(),
                    };
                    if let Some(keys) = &local_keys {
                        if !next.is_empty() {
                            next.push_str("; ");
                        }
                        next.push_str(&local_settings_next(keys));
                    }
                    why = format!("{why}; held by {}; {next}", moai_dir_reference_note(&refs));
                }
            }
            Some(("blocked", format!("the Go-generation .moai/ stays: {why}")))
        }
        // Without a Rust marker `state_root::dir` names the earlier `.moai-cli` folder.
        "none" => held(state_root::dir(root)).map(|r| {
            (
                "blocked",
                format!("the earlier .moai-cli/ state stays: {r}"),
            )
        }),
        _ => None,
    }
}

/// The local settings file, which this update never writes (#2936 L2a, D1 condition 3).
const LOCAL_SETTINGS: &str = ".claude/settings.local.json";

/// #2936 L2b: what to do about Go references in the local settings file: which file, which keys
/// (names only), that this update does not edit it, and the command that follows the user's edit.
fn local_settings_next(keys: &str) -> String {
    format!(
        "this update does not edit {LOCAL_SETTINGS}; edit or remove these keys yourself ({keys}), \
         then run `moai project-update-plan` again"
    )
}

/// #2936 L1: the top-level keys of the project and local settings files whose values name
/// `.moai/`, per file. Key names only: a value is never printed. A file that cannot be read as
/// a settings object is left out (the settings plan already reports that).
fn moai_dir_references(root: &Path) -> Vec<(&'static str, Vec<String>)> {
    let mut out = Vec::new();
    for (scope, file) in [
        (SettingsScope::Project, ".claude/settings.json"),
        (SettingsScope::Local, LOCAL_SETTINGS),
    ] {
        let Ok((_, current)) = claude::scope_view(root, scope) else {
            continue;
        };
        let mut keys: Vec<String> = current
            .iter()
            .filter(|(_, v)| v.to_string().contains(".moai/"))
            .map(|(k, _)| k.clone())
            .collect();
        keys.sort();
        if !keys.is_empty() {
            out.push((file, keys));
        }
    }
    out
}

/// `file (key, key); file (key)` for [`moai_dir_references`].
fn moai_dir_reference_note(refs: &[(&'static str, Vec<String>)]) -> String {
    refs.iter()
        .map(|(file, keys)| format!("{file} ({})", keys.join(", ")))
        .collect::<Vec<_>>()
        .join("; ")
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

/// The personal-file choice as the manifest records it (#2953 F2): an `AGENTS.local.md` seen
/// once with the `.gitignore` section MoAI owns is `ignored`, without it `tracked`, neither
/// is `none`. `Err` when the record cannot tell: the section without the file, or the file
/// without the section while a user rule in `.gitignore` names it (an `ignored` choice
/// records no section then).
///
/// Since #3009 S3b the section is written for every choice (the state-folder allow-list), so
/// "the section is there" no longer means "the personal file is listed". What the section
/// lists is its recorded `kind`; a section recorded before the kind existed listed the file.
fn local_file(root: &Path, manifest: &Value) -> Result<&'static str> {
    const LOCAL: &str = "AGENTS.local.md";
    let created = manifest["created_once"]
        .as_array()
        .into_iter()
        .flatten()
        .any(|p| p == LOCAL)
        // Manifests written before `created_once` recorded the file in owned_files.
        || manifest["owned_files"]
            .as_array()
            .into_iter()
            .flatten()
            .any(|f| f["path"] == LOCAL && f["section_id"].is_null());
    // `Some(listed)` when the section is recorded: whether it lists the personal file.
    let section: Option<bool> = manifest["owned_files"]
        .as_array()
        .into_iter()
        .flatten()
        .find(|f| f["section_id"] == crate::agents::IGNORE_SECTION)
        .map(|f| crate::agents::ignore_kind_lists_local(f["kind"].as_str()));
    let unknown = "the generation manifest does not tell whether AGENTS.local.md is tracked \
                   or ignored; ask the user and regenerate through /moai project";
    let user_rule_decides = || -> Result<&'static str> {
        let ignore = crate::apply::read_opt(&crate::apply::target(root, ".gitignore")?)?;
        let user_rule = ignore
            .as_deref()
            .and_then(|b| std::str::from_utf8(b).ok())
            .is_some_and(crate::agents::ignore_has_rule);
        if user_rule {
            Err(unknown.into())
        } else {
            Ok("tracked")
        }
    };
    match (created, section) {
        (true, Some(true)) => Ok("ignored"),
        (false, None) | (false, Some(false)) => Ok("none"),
        (false, Some(true)) => Err(unknown.into()),
        (true, None) | (true, Some(false)) => user_rule_decides(),
    }
}

/// #2936 L2b: a workflow reference recorded while the state lived in the earlier `.moai-cli/`
/// follows the state once it has moved to `.moai/` (the go-state step moves that folder's product
/// state, not the path inside the generation manifest). Without this the generation step waits
/// for a folder that no longer exists ("workflow reference directory missing") and no command
/// ever clears it. A reference elsewhere, or a project whose state has not moved, is as recorded.
fn follow_state_dir(root: &Path, recorded: &str) -> String {
    let earlier = format!("{}/", state_root::EARLIER_PROJECT_DIR);
    match recorded.strip_prefix(&earlier) {
        Some(rest) if state_root::dir(root) == state_root::MOAI_DIR => {
            format!("{}/{rest}", state_root::MOAI_DIR)
        }
        _ => recorded.to_string(),
    }
}

/// The generation request rebuilt from the profile and the manifest's record (#2936 S1).
/// `Err` is a question for the user (the step reads `needs_decision`).
fn generation_request(
    root: &Path,
    profile: &Profile,
    manifest: &Value,
) -> Result<GenerationRequest> {
    let language = profile
        .language
        .as_ref()
        .and_then(|l| l.get("instruction_files"))
        .filter(|code| *code == "en" || *code == "ko")
        .cloned()
        .unwrap_or_else(|| "en".into());
    let workflow_reference = manifest["workflow_reference"]
        .as_str()
        .map(|r| follow_state_dir(root, r))
        .unwrap_or_else(|| format!("{}/workflows/project/", state_root::dir(root)));
    let workflow_id = manifest["workflow_id"]
        .as_str()
        .map(String::from)
        .unwrap_or_else(|| "project".into());
    Ok(GenerationRequest {
        protocol_version: 1,
        language,
        workflow_reference,
        local_file: local_file(root, manifest)?.into(),
        workflow_id,
        native_roles: Vec::new(),
        skill_projections: Vec::new(),
        project_skills: Vec::new(),
        picks: Vec::new(),
    })
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
    // Read without following a link on the way, like every other manifest read.
    let rel = format!("{}/manifests/generation.json", state_root::dir(root));
    let manifest: Value = match crate::apply::read_opt(&crate::apply::target(root, &rel)?)? {
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
    // R3 (#2936): a profile that is still the first-launch placeholder cannot generate; the
    // interview comes first, so no apply is printed that `generation-apply` would refuse.
    if profile.is_placeholder() {
        return Ok(Step::waiting(
            "generation",
            "needs_decision",
            "the project profile is still the placeholder a first launch writes; run \
             /moai project to answer the interview before generating",
        ));
    }
    let request = match generation_request(root, profile, &manifest) {
        Ok(r) => r,
        Err(reason) => return Ok(Step::waiting("generation", "needs_decision", &reason)),
    };
    let plan = generation::generation_plan(root, request.clone())?;
    let choices = generation_choices(root, shown, &request, &plan)?;
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
    let mut preview =
        json!({"files": files, "warnings": plan.warnings, "host_notes": plan.host_notes});
    // The keys are absent when there is nothing to choose, so a preview without a conflict or
    // a shadowing CLAUDE.md reads as it did before (SPEC-20261119-agents-choices-step REQ-9).
    if !choices.is_empty() {
        preview["choices"] = Value::Array(choices);
    }
    if let Some(shadow) = &plan.claude_shadow {
        preview["claude_shadow"] = encode(shadow)?;
    }
    Ok(Step {
        name: "generation",
        state,
        reason: reason.into(),
        preview,
        plan_digest: Some(plan.plan_digest),
        apply,
    })
}

/// The choices a session can run after the user's pick (SPEC-20261119-agents-choices-step
/// REQ-6..8, card C-57). Each one is the whole request planned again with that choice in
/// `picks`, so its digest and its `generation-apply` command belong to that choice alone.
///
/// - `AGENTS.md` is a conflict for a missing managed block: `append-managed` (recommended)
///   and `skip-conflicts`, and with a shadowing CLAUDE.md also `append-managed+claude-import`
///   (an import alone would be refused while the conflict stands).
/// - No file conflicts and a CLAUDE.md shadows AGENTS.md: `claude-import`.
///
/// A choice whose own plan still has a conflict row is left out: it could not be applied.
/// Nothing here writes.
fn generation_choices(
    root: &Path,
    shown: &str,
    request: &GenerationRequest,
    plan: &generation::GenerationPlan,
) -> Result<Vec<Value>> {
    let missing_block = plan.files.iter().any(|f| {
        f.path == "AGENTS.md"
            && f.action == "conflict"
            && f.detail
                .as_deref()
                .is_some_and(|d| d.contains(crate::agents::BLOCK_MISSING))
    });
    let any_conflict = plan.files.iter().any(|f| f.action == "conflict");
    let shadow = plan.claude_shadow.is_some();
    let mut sets: Vec<(&str, Vec<&str>, &str)> = Vec::new();
    if missing_block {
        sets.push((
            "append-managed",
            vec!["append-managed"],
            "add the managed section at the end of AGENTS.md; every other byte stays",
        ));
        sets.push((
            "skip-conflicts",
            vec!["skip-conflicts"],
            "keep AGENTS.md as it is; the other files are still written",
        ));
        if shadow {
            sets.push((
                "append-managed+claude-import",
                vec!["append-managed", "claude-import"],
                "append the managed section and add the one import line of AGENTS.md to CLAUDE.md",
            ));
        }
    } else if shadow && !any_conflict {
        sets.push((
            "claude-import",
            vec!["claude-import"],
            "add the one import line of AGENTS.md to CLAUDE.md",
        ));
    }
    let mut out = Vec::new();
    for (id, picks, effect) in sets {
        let mut picked = request.clone();
        picked.picks = picks.iter().map(|p| p.to_string()).collect();
        let Ok(choice_plan) = generation::generation_plan(root, picked) else {
            continue;
        };
        if choice_plan.files.iter().any(|f| f.action == "conflict") {
            continue;
        }
        out.push(json!({
            "id": id,
            "recommended": id == "append-managed",
            "picks": picks,
            "effect": effect,
            "plan_digest": choice_plan.plan_digest,
            "apply": {
                "argv": argv("generation-apply", shown, &choice_plan.plan_digest),
                "stdin": encode(&choice_plan)?,
            },
        }));
    }
    Ok(out)
}

/// Q2 A (#2936): the keys MoAI already owns, without `hooks` (only `hooks-apply` writes
/// that key), are kept as they are; no key is added by default.
fn settings_step(root: &Path, shown: &str) -> Result<Step> {
    let home = std::env::var_os("HOME").map(std::path::PathBuf::from);
    let path_env = std::env::var("PATH").unwrap_or_default();
    settings_step_with(root, shown, home.as_deref(), &path_env)
}

/// [`settings_step`] for a given HOME and PATH.
///
/// #3004 P4-3: the status line is `moai statusline` (`status_line::current`, the same binary
/// name the hooks call). A project with none gets it, and a status line MoAI owns in the form
/// the earlier releases wrote (`moai-cli statusline`) and the file still holds is rewritten to
/// it, both only where [`status_line::gate`] lets it through. When the gate holds, the old
/// status line stays in the printed owned keys (so nothing is rewritten) and the step says why
/// in `status_line_migration`; the hooks step is not touched, its gate is the hook gate's. An
/// edited or foreign status line is never touched and its text is never printed.
fn settings_step_with(
    root: &Path,
    shown: &str,
    home: Option<&Path>,
    path_env: &str,
) -> Result<Step> {
    let mut owned = settings_stdin(claude::owned_settings(root)?);
    let file = claude::current_settings(root)?;
    // SPEC-20261119-hooks-followups (N4): MoAI's record still owns a status line the file no
    // longer has. The plan keeps it removed (the user's edit); the step must not call that
    // "current" without saying so.
    let status_line_removed = owned.contains_key("statusLine") && !file.contains_key("statusLine");
    // #2936 R8 (Q-H1, decision 6880072221): a project that has no status line of its own gets
    // MoAI's. One it has stays as it is, and so does one MoAI wrote before.
    let intent = status_line::intent(owned.get("statusLine"), file.get("statusLine"));
    let verdict = match intent {
        status_line::Intent::Write(kind) => {
            let verdict = status_line::gate(kind, root, home, path_env);
            if verdict.is_ok() {
                owned.insert("statusLine".into(), status_line::current());
            }
            Some(verdict)
        }
        status_line::Intent::Keep(_) => None,
    };
    let desired = claude::user_settings_desired_in(root, &owned, SettingsScope::Project)?;
    let plan = claude::settings_plan_in(root, &desired, SettingsScope::Project)?;
    let mut step = settings_from_plan(shown, &plan, owned);
    step.preview["status_line_migration"] = status_line::migration(intent, verdict.as_ref());
    if let Some(Err(held)) = &verdict {
        // With nothing else to write, the hold is the news, not "current".
        if step.state == "unchanged" {
            step.state = "needs_decision";
            step.reason = held.message.clone();
        }
    }
    if status_line_removed && step.state == "unchanged" {
        step.reason = "the status line MoAI wrote was removed from .claude/settings.json by hand \
                       and is not put back; the other MoAI-owned keys are current"
            .into();
        step.preview["status_line_removed"] = json!(true);
    }
    // #2936 L1: "current" must not hide keys that still name the earlier generation's `.moai/`
    // folder: they keep the go-state step waiting, and nothing here changes them yet.
    if step.state == "unchanged" && state_root::classify(root).generation == "go" {
        let refs = moai_dir_references(root);
        if !refs.is_empty() {
            step.reason = format!(
                "{}; still naming .moai/ and holding the Go .moai/ folder (go-state step): {}",
                step.reason,
                moai_dir_reference_note(&refs)
            );
            step.preview["go_references"] = json!(refs
                .iter()
                .map(|(file, keys)| json!({"file": file, "keys": keys}))
                .collect::<Vec<_>>());
        }
    }
    Ok(step)
}

/// The keys a printed settings apply carries on stdin: the owned keys without `hooks`.
fn settings_stdin(mut owned: Map<String, Value>) -> Map<String, Value> {
    owned.remove("hooks");
    owned
}

/// The settings step from its plan; `owned` (without `hooks`) is the printed stdin.
fn settings_from_plan(shown: &str, plan: &Value, owned: Map<String, Value>) -> Step {
    let digest = plan["plan_digest"].as_str().map(String::from);
    let (state, reason) = match plan["action"].as_str() {
        Some("create" | "update") => ("ready", "MoAI-owned keys to restore"),
        Some("unchanged") if owned.is_empty() => (
            "unchanged",
            "MoAI owns no key in this file; the keys it has are left as they are",
        ),
        Some("unchanged") => ("unchanged", "the MoAI-owned keys are current"),
        _ => (
            "conflict",
            "a key MoAI manages was set by the user; nothing is written",
        ),
    };
    // #2383: a conflict the plan offers to replace prints that replacement as its command; it
    // runs only after the user's explicit confirmation (#2953 F6).
    let token = plan["confirm_replace"]["confirm_token"].as_str();
    let (state, reason) = match (state, token) {
        ("conflict", Some(_)) => (
            "needs_decision",
            "a key MoAI manages was set by the user; replacing it needs the user's explicit \
             confirmation (confirm_replace)",
        ),
        _ => (state, reason),
    };
    let apply = match (&digest, state, token) {
        (Some(d), "ready", _) => Some((
            argv("claude-settings-apply", shown, d),
            Value::Object(owned),
        )),
        (Some(d), "needs_decision", Some(t)) => {
            let mut words = argv("claude-settings-apply", shown, d);
            words.extend(["--confirm-replace".to_string(), t.to_string()]);
            Some((words, Value::Object(owned)))
        }
        _ => None,
    };
    Step {
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
    }
}

/// #2936 R-B (O-1): the Go-generation hook entries in `.claude/settings.json` are replaced by
/// the MoAI hook set with `go-hooks-apply`. `replace` prints that command (`permissions` is
/// not touched), `blocked` (entries of other tools share `hooks`) waits for the user without a
/// command, `none` is `unchanged`. HOME is not read (O-3).
///
/// SPEC-20261119-hooks-followups: where the profile lists `claude`, the hooks plan of the
/// Claude host is consulted even with no Go entry left and no record (leader decision B on N1),
/// so the step's words say what the file holds and name the manual `hooks-plan` command; the
/// step itself stays `unchanged` and prints no command there. Without `claude` in the profile
/// the Go plan alone decides the step.
fn hooks_step(root: &Path, shown: &str, claude: bool) -> Result<Step> {
    let plan = crate::go_hooks::plan(root, None)?;
    let digest = plan["plan_digest"].as_str().map(String::from);
    let names = |key: &str| -> Vec<Value> {
        plan[key]
            .as_array()
            .map(|a| a.iter().filter(|v| !v.is_null()).cloned().collect())
            .unwrap_or_default()
    };
    let preview = json!({
        "action": plan["action"],
        "blocks": plan["blocks"],
        "go_entries": names("go_entries"),
        "kept_entries": plan["kept_entries"],
        "added_events": plan["added"].as_object().map(|m| m.keys().cloned().collect::<Vec<_>>()),
        "scripts": names("scripts"),
        "scripts_unreferenced": names("scripts_unreferenced"),
        "scripts_missing": names("scripts_missing"),
    });
    // #2936 S-CYCLE (OQ2 H1): no Go entry is left because the Go settings moved into a backup;
    // the MoAI hook set is then written by the ordinary `hooks-apply` for Claude. C-55: the
    // same holds where MoAI's record says the hook set is installed — the file may not hold it.
    if plan["action"].as_str() == Some("none") {
        if let Some(step) = hooks_after_settings_move(root, shown, &preview, claude)? {
            return Ok(step);
        }
    }
    let (state, reason) = match plan["action"].as_str() {
        Some("replace") => (
            "ready",
            "Go-generation hook entries to replace with the MoAI hook set".to_string(),
        ),
        Some("none") => ("unchanged", "no Go-generation hooks are left".to_string()),
        _ => (
            "needs_decision",
            format!(
                "the replacement is blocked: {}; the user decides outside this route",
                plan["blocks"]
            ),
        ),
    };
    let apply = (state == "ready")
        .then(|| {
            let d = digest.as_deref().ok_or("hooks plan without a digest")?;
            Ok::<_, String>((argv("go-hooks-apply", shown, d), Value::Null))
        })
        .transpose()?;
    Ok(Step {
        name: "hooks",
        state,
        reason,
        preview,
        plan_digest: digest,
        apply,
    })
}

/// #2936 S-CYCLE (OQ2 H1): the Go settings moved into a backup (a committed `go-state-apply`
/// whose receipt is not reverted holds them), so the Go hook entries are gone with them and
/// `go-hooks-apply` has nothing to replace, yet the new settings carry no hook set. The step
/// then prints `hooks-apply --host claude`. Without such a receipt (a project that never had
/// Go settings, or a reverted move) it offers nothing: `None`.
///
/// C-55: a project whose ownership record names the MoAI hook set is planned the same way, so
/// a settings file that lost its `hooks` key (the record is not tracked by git, the file is)
/// reads `ready` here and never `unchanged`.
///
/// SPEC-20261119-hooks-followups: with `claude` in the profile's hosts the plan is made in every
/// case (no record, no move either), and the words of the step say what the file holds: a
/// `hooks` key that is not the MoAI set is never "current" (N3), and an unreadable record names
/// its file and the next step (N5). Without a move and without a record the step still offers
/// nothing (the rule above, leader decision B on N1): it stays `unchanged` and its reason names
/// the gap and the command to run by hand, `preview.configured` reading as `hooks-status` does.
fn hooks_after_settings_move(
    root: &Path,
    shown: &str,
    base: &Value,
    claude: bool,
) -> Result<Option<Step>> {
    let moved = go_state::backed_up_settings(root)?.is_some();
    let recorded = crate::hooks::installed(root, "claude").map_err(|e| record_error(root, &e))?;
    if !claude && !moved && !recorded {
        return Ok(None);
    }
    let plan = crate::hooks::plan(root, "claude", None)?;
    let digest = plan["plan_digest"].as_str().map(String::from);
    let mut preview = base.clone();
    preview["action"] = plan["action"].clone();
    preview["host"] = json!("claude");
    preview["hooks"] = plan["hooks"].clone();
    let offer = moved || recorded;
    let (state, reason): (&'static str, String) = match plan["action"].as_str() {
        Some("create" | "update") if moved => (
            "ready",
            "the Go settings moved into the backup and took their hook entries with them; the \
             MoAI hook set is not in the new settings yet"
                .into(),
        ),
        Some("create" | "update") if recorded => (
            "ready",
            "MoAI's record names the MoAI hook set, but .claude/settings.json does not hold it"
                .into(),
        ),
        Some("create" | "update") => {
            preview["configured"] = json!(hook_status_configured(root));
            (
                "unchanged",
                format!(
                    "no MoAI hook set is installed in .claude/settings.json (hooks-plan reads \
                     {}); this step prints no command here, to install it run `moai hooks-plan \
                     {shown} --host claude`, then `moai hooks-apply {shown} --host claude` with \
                     its plan_digest",
                    plan["action"].as_str().unwrap_or("update")
                ),
            )
        }
        Some("unchanged") => {
            let (state, reason) = unchanged_hooks(root, &mut preview)?;
            (state, reason.into())
        }
        _ if !offer => (
            "unchanged",
            "the MoAI hook set conflicts with a key the user set in .claude/settings.json (hooks-plan \
             reads conflict); this step prints no command and nothing is written"
                .into(),
        ),
        _ => (
            "conflict",
            "the MoAI hook set conflicts with a key the user set in .claude/settings.json; \
             nothing is written"
                .into(),
        ),
    };
    let apply = match (&digest, state) {
        (Some(d), "ready") => Some((
            [
                "moai",
                "hooks-apply",
                shown,
                "--host",
                "claude",
                "--plan-digest",
                d,
                "--permission-reference",
                PERMISSION_PLACEHOLDER,
            ]
            .map(String::from)
            .to_vec(),
            Value::Null,
        )),
        _ => None,
    };
    Ok(Some(Step {
        name: "hooks",
        state,
        reason,
        preview,
        plan_digest: digest,
        apply,
    }))
}

/// What `hooks-status` reads for the Claude host's `configured` (`yes`, `no`, `drifted`,
/// `legacy_binary`); `drifted` when the status cannot be read.
fn hook_status_configured(root: &Path) -> String {
    crate::hooks::status(root, "claude", None, "")
        .ok()
        .and_then(|s| s["configured"].as_str().map(String::from))
        .unwrap_or_else(|| "drifted".into())
}

/// What a hooks plan that reads `unchanged` means for the `hooks` key the file holds now: the
/// MoAI set is "current" only where the key is that set. An edit of it is kept and said to be
/// one (the status reads `drifted`); an empty key leaves no hook in effect, no command changes
/// a key the user edited, so the step waits for the user and names the way out.
fn unchanged_hooks(root: &Path, preview: &mut Value) -> Result<(&'static str, &'static str)> {
    let held = crate::claude::read_existing(&root.join(".claude/settings.json"))?
        .and_then(|b| serde_json::from_slice::<Value>(&b).ok())
        .and_then(|v| v.get("hooks").cloned())
        .unwrap_or(Value::Null);
    if held == Value::Object(crate::hooks::hooks_object("claude")?) {
        return Ok(("unchanged", "the MoAI hook set is current"));
    }
    let empty = held.is_null() || held.as_object().is_some_and(|m| m.is_empty());
    preview["configured"] = json!("drifted");
    preview["hooks_in_effect"] = json!(!empty);
    Ok(if empty {
        (
            "needs_decision",
            "`hooks` in .claude/settings.json is empty, so no hook is in effect, and no moai \
             command changes a key you edited; to get the MoAI hooks back, remove the `hooks` \
             key from the file by hand and plan again",
        )
    } else {
        (
            "unchanged",
            "`hooks` in .claude/settings.json is not the MoAI hook set (your edit is kept and \
             hooks-status reads drifted); no moai command changes it",
        )
    })
}

/// The message of an unreadable ownership record, with its file and the next step (N5): the
/// record is MoAI's own file under the state folder and no command repairs it. A message that is
/// not about the record (the settings file itself) stays as it is.
fn record_error(root: &Path, e: &str) -> String {
    // The claude record is read through the settings state, whose refusals name the cause
    // ("symlink target refused", "target is not a bounded regular file", "target read failed")
    // but not the file, and none of them carries "receipt" — name the record here, like the
    // receipt errors below do (card C-107 slice 2: a refusal names its file).
    let e = if e.starts_with("symlink target refused")
        || e.starts_with("target is not a bounded regular file")
        || e.starts_with("target read failed")
    {
        format!("{e}: the ownership receipt is unreadable")
    } else {
        e.to_string()
    };
    if !e.contains("receipt") {
        return e;
    }
    format!(
        "the ownership record {} cannot be read ({e}); no moai command repairs it: read it, \
         move it aside by hand if it is damaged, and plan again",
        state_root::rel(root, "manifests/claude-settings.json")
    )
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

/// The install steps in the order they are printed and applied; `hooks` follows
/// `claude-settings` (#2936 R-B), the host steps follow `github` (#2936 R9 S5a).
const INSTALL: [&str; 11] = [
    "profile",
    "generation",
    "claude-settings",
    "hooks",
    "github",
    "codex-hooks",
    "codex-config",
    "codex-skill",
    "omp-hooks",
    "omp-config",
    "omp-agents",
];

/// The host steps (#2936 R9 S5a) as this module's steps; `picked` is the profile's `hosts`, or
/// `None` when no profile could be read (SPEC-20261118-profile-host-omp).
fn host_steps(root: &Path, shown: &str, picked: Option<&[String]>) -> Vec<Step> {
    project_hosts::steps(root, shown, picked)
        .into_iter()
        .map(|h| Step {
            name: h.name,
            state: h.state,
            reason: h.reason,
            preview: h.preview,
            plan_digest: h.plan_digest,
            apply: h.apply,
        })
        .collect()
}

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
            steps.push(or_blocked("hooks", hooks_step(root, shown, false)));
            steps.push(or_blocked("github", github_step(root, shown)));
            steps.extend(host_steps(root, shown, None));
            hooks_after_settings(&mut steps);
            return steps;
        }
    };
    let profile_step = or_blocked("profile", profile_step(root, shown, &profile, detected));
    let generation = if profile_step.state == "ready" {
        Step::waiting("generation", "blocked", AFTER_PROFILE)
    } else {
        or_blocked("generation", generation_step(root, shown, &profile))
    };
    let claude = profile.hosts.iter().any(|h| h == "claude");
    let mut steps = vec![
        profile_step,
        generation,
        or_blocked("claude-settings", settings_step(root, shown)),
        or_blocked("hooks", hooks_step(root, shown, claude)),
        or_blocked("github", github_step(root, shown)),
    ];
    steps.extend(host_steps(root, shown, Some(&profile.hosts)));
    hooks_after_settings(&mut steps);
    steps
}

/// The `hooks` step waits while the `claude-settings` step has a command to run (#3028 audit
/// B1): both rewrite `.claude/settings.json` and the `go-hooks-apply` digest binds that file's
/// bytes, so the second command of one preview would always be refused as a stale plan. The
/// next preview, made after the settings apply, plans the hooks step against the new file.
fn hooks_after_settings(steps: &mut [Step]) {
    let pending = steps
        .iter()
        .any(|s| s.name == "claude-settings" && s.apply.is_some());
    if !pending {
        return;
    }
    if let Some(hooks) = steps
        .iter_mut()
        .find(|s| s.name == "hooks" && s.apply.is_some())
    {
        *hooks = Step::waiting("hooks", "blocked", AFTER_SETTINGS);
    }
}

/// Whether the backup phase still has something to do that makes the later previews stale: the
/// go-state step wants a decision (a move, or a user fix it cannot print), or the go-settings
/// step has a command to run (#2936 L2a). A held go-settings step (`needs_decision` without a
/// command) changes nothing, so it does not count.
fn backup_pending(backup: &[Step]) -> bool {
    backup.iter().any(|s| match s.name {
        "go-state" => s.state == "needs_decision",
        "go-settings" => s.apply.is_some(),
        _ => false,
    })
}

/// The merge phase (#2936 R10): `AGENTS.local.md` into `AGENTS.md`. It waits while the backup
/// phase has a move to make, and while the generation step has files to write.
fn merge_steps(root: &Path, shown: &str, backup: &[Step], install: &[Step]) -> Vec<Step> {
    const NAME: &str = "agents-local-merge";
    const SETTINGS: &str = "claude-settings-merge";
    if backup_pending(backup) {
        return vec![
            Step::waiting(NAME, "blocked", AFTER_BACKUP),
            Step::waiting(SETTINGS, "blocked", AFTER_BACKUP),
        ];
    }
    let pending = |step: &str| {
        install
            .iter()
            .any(|s| s.name == step && matches!(s.state, "ready" | "needs_decision"))
    };
    let agents = if install
        .iter()
        .any(|s| s.name == "generation" && s.state == "ready")
    {
        Step::waiting(NAME, "blocked", AFTER_INSTALL)
    } else {
        or_blocked(NAME, agents_local_merge_step(root, shown))
    };
    // The settings merge reads the file the install step writes: with something to merge, its
    // digest would be stale before that step ran.
    let settings = or_blocked(
        SETTINGS,
        claude_settings_merge_step(root, shown, pending("claude-settings")),
    );
    vec![agents, settings]
}

/// The settings merge (#2936 R8): what the user had set in the Go generation's
/// `.claude/settings.json`, read from the copy `go-state-apply` moved into its backup, offered
/// back key by key. The printed command is the ordinary `claude-settings-apply` with the
/// owned keys plus every carriable candidate on stdin, so ownership, the three-way plan and the
/// digest are the ones that file already has; a merge of only some keys is a new
/// `claude-settings-plan` with fewer keys.
fn claude_settings_merge_step(root: &Path, shown: &str, install_pending: bool) -> Result<Step> {
    const NAME: &str = "claude-settings-merge";
    let Some(found) = claude_settings_merge::candidates(root)? else {
        return Ok(Step::waiting(
            NAME,
            "skip",
            "no earlier settings.json was moved into a backup, so there is nothing to merge back",
        ));
    };
    let preview = |extra: Value| {
        json!({
            "backup": found.rel,
            "backup_sha256": found.sha256,
            "candidates": found.rows,
            "commit_notice": "a merged value is written to .claude/settings.json, a file that \
                              is usually committed",
            "plan": extra,
        })
    };
    if found.carry.is_empty() {
        return Ok(Step {
            name: NAME,
            state: "skip",
            reason: "nothing in the earlier settings can be merged back; the file stays in the \
                     backup"
                .into(),
            preview: preview(Value::Null),
            plan_digest: None,
            apply: None,
        });
    }
    if install_pending {
        return Ok(Step {
            name: NAME,
            state: "blocked",
            reason: AFTER_INSTALL.into(),
            preview: preview(Value::Null),
            plan_digest: None,
            apply: None,
        });
    }
    let mut stdin = settings_stdin(claude::owned_settings(root)?);
    for (key, value) in &found.carry {
        stdin.insert(key.clone(), value.clone());
    }
    let desired = claude::user_settings_desired_in(root, &stdin, SettingsScope::Project)?;
    let plan = claude::settings_plan_in(root, &desired, SettingsScope::Project)?;
    let digest = plan["plan_digest"].as_str().map(String::from);
    let (state, reason) = match plan["action"].as_str() {
        Some("create" | "update") => (
            "needs_decision",
            "earlier settings the user had can be merged back; which ones is the user's choice",
        ),
        Some("unchanged") => ("unchanged", "the earlier settings are already in the file"),
        _ => (
            "conflict",
            "a key was changed in the file since MoAI wrote it; nothing is merged",
        ),
    };
    let apply = match (&digest, state) {
        (Some(d), "needs_decision") => Some((
            argv("claude-settings-apply", shown, d),
            Value::Object(stdin),
        )),
        _ => None,
    };
    Ok(Step {
        name: NAME,
        state,
        reason: reason.into(),
        preview: preview(json!({
            "action": plan["action"],
            "settings_path": plan["settings_path"],
            "conflicts": plan["conflicts"],
        })),
        plan_digest: digest,
        apply,
    })
}

fn agents_local_merge_step(root: &Path, shown: &str) -> Result<Step> {
    let plan = agents_local_merge::plan(root)?;
    let digest = plan["plan_digest"].as_str().map(String::from);
    let reason = plan["reason"].as_str().unwrap_or_default().to_string();
    let state = match plan["action"].as_str() {
        Some("merge") => "needs_decision",
        Some("unchanged") => "unchanged",
        Some("none") => "skip",
        Some("conflict") => "conflict",
        _ => "blocked",
    };
    let apply = match (&digest, state) {
        (Some(d), "needs_decision") => Some((argv("agents-merge-apply", shown, d), Value::Null)),
        _ => None,
    };
    Ok(Step {
        name: "agents-local-merge",
        state,
        reason,
        preview: json!({
            "action": plan["action"],
            "source": plan["source"],
            "target": plan["target"],
            "body": plan["body"],
            "commit_notice": plan["commit_notice"],
            "secret_suspects": plan["secret_suspects"],
        }),
        plan_digest: digest,
        apply,
    })
}

/// What every leftover says it is not (#2936 L2b).
const LEFTOVER_NOTE: &str = "not removed by update";
/// The Go tool's update lock, relative to the project.
const GO_UPDATE_LOCK: &str = ".moai/.update.lock";
/// A lock file is a few lines of JSON; this only refuses absurd inputs.
const LOCK_LIMIT: u64 = 4096;
/// How many entry names the `.moai-cli/` fact lists.
const LEFTOVER_NAMES: usize = 8;

fn leftover(path: &str, fact: String) -> Value {
    json!({"path": path, "fact": fact, "note": LEFTOVER_NOTE})
}

/// #2936 L2b (completion condition 3 of the Issue): what earlier generations leave behind that no
/// step of this update removes, as facts (path, one line, and that the update does not remove
/// it). Read only, no deletion offered: the user decides. A fact needs its proof:
///
/// - `.moai-cli/`: a real folder while the state already lives in `.moai/` (state_root `rust`).
///   Before that it is the live state folder (or moves with the go-state step) and is no leftover.
/// - `CLAUDE.md`: a regular file with Go proof (the Go manifest records it, or it holds a Go
///   curator block) and no managed section of this generation, while the Go `.moai/` folder is
///   not there for the go-state step to move it with.
/// - `.moai/.update.lock`: the Go tool's lock file whose recorded process is not running
///   (`kill -0`; a process that cannot be judged, or a lock that cannot be read, is not stated).
fn leftovers(root: &Path, generation: &str) -> Vec<Value> {
    let mut out = Vec::new();
    if generation == "rust" {
        if let Ok(m) = fs::symlink_metadata(root.join(state_root::EARLIER_PROJECT_DIR)) {
            if m.is_dir() {
                let mut names: Vec<String> =
                    fs::read_dir(root.join(state_root::EARLIER_PROJECT_DIR))
                        .map(|rd| {
                            rd.flatten()
                                .map(|e| e.file_name().to_string_lossy().into_owned())
                                .collect()
                        })
                        .unwrap_or_default();
                names.sort();
                let held = match names.len() {
                    0 => "is an empty folder".to_string(),
                    n => {
                        // File names are the project's, not ours (AUDIT #3117 E4): a long one is
                        // cut and a control or direction character shows as `?`.
                        let shown = names
                            .iter()
                            .take(LEFTOVER_NAMES)
                            .map(|n| crate::card::input::echo(n))
                            .collect::<Vec<_>>()
                            .join(", ");
                        let more = n.saturating_sub(LEFTOVER_NAMES);
                        if more > 0 {
                            format!("still holds {n} entries ({shown}, and {more} more)")
                        } else {
                            format!("still holds {n} entries ({shown})")
                        }
                    }
                };
                out.push(leftover(
                    &format!("{}/", state_root::EARLIER_PROJECT_DIR),
                    format!(
                        "the earlier generation's state folder {held}; this generation's state \
                         is in .moai/"
                    ),
                ));
            }
        }
    }
    if generation != "go" {
        if let Some(proof) = go_state::go_claude_md_proof(root) {
            let how = if proof == "manifest" {
                "the Go manifest records it"
            } else {
                "it holds a Go curator block"
            };
            out.push(leftover(
                "CLAUDE.md",
                format!("a Go-generation instruction file ({how}) is still in the project root"),
            ));
        }
    }
    if let Some(fact) = dead_update_lock(root) {
        out.push(leftover(GO_UPDATE_LOCK, fact));
    }
    out
}

/// The fact for a Go update lock whose recorded process is not running; `None` for no lock, a
/// link, a lock that cannot be read as the Go payload `{pid, started_at, hostname}`, or a process
/// that is alive or cannot be judged.
fn dead_update_lock(root: &Path) -> Option<String> {
    let path = root.join(GO_UPDATE_LOCK);
    let m = fs::symlink_metadata(&path).ok()?;
    if !m.is_file() || m.len() > LOCK_LIMIT || crate::apply::target(root, GO_UPDATE_LOCK).is_err() {
        return None;
    }
    let payload: Value = serde_json::from_slice(&fs::read(&path).ok()?).ok()?;
    let pid = payload["pid"]
        .as_i64()
        .filter(|p| *p > 0 && *p <= i64::from(i32::MAX))?;
    if crate::home_state::pid_alive(pid)? {
        return None;
    }
    // The time is shown only when it looks like one: the file is not trusted text.
    let since = payload["started_at"]
        .as_str()
        .filter(|s| {
            s.len() <= 32
                && s.chars()
                    .all(|c| c.is_ascii_digit() || matches!(c, '-' | ':' | 'T' | 'Z' | '.' | '+'))
        })
        .map(|s| format!(" (started {s})"))
        .unwrap_or_default();
    Some(format!(
        "the Go update lock names process {pid}{since}, which is not running"
    ))
}

/// #2936 E1 (AUDIT #3117): how to take the update chain back, last applied first. Printed with
/// every preview, because the order is not a matter of taste: `go-state-revert` refuses while the
/// generation step's `AGENTS.md` is still in place ("revert refused: AGENTS.md is conflict"), and
/// the generation step's `AGENTS.md` is a second receipt (`agents_receipt`) that its own undo does
/// not cover. Run in this order the project is back to its start bytes (what the steps added —
/// receipts, `.moai.rust-<UTC>/`, the Codex and OMP files — stays). Each id is printed by the
/// apply result of its step.
fn undo_order() -> Value {
    let none = "no revert command exists for this step; its files stay and do not block the commands below";
    json!({
        "rule": "undo the step applied last first; every command needs its own permission reference",
        "steps": [
            {"step": "omp-hooks", "commands": [], "note": none},
            {"step": "codex-config", "commands": [], "note": none},
            {"step": "codex-hooks", "commands": [], "note": none},
            {"step": "hooks", "commands": [
                "moai go-hooks-revert --receipt <receipt of the hooks step> --permission-reference <reference> <PROJECT>",
            ], "note": "for a hooks step that ran go-hooks-apply (the Go hook set); a hooks step that ran hooks-apply --host claude over an existing settings file printed its own `revert` (moai hooks-revert <PROJECT> --receipt <id> --permission-reference <reference>), and a first install has nothing to revert"},
            {"step": "generation", "commands": [
                "moai apply-recover --receipt <agents_receipt of the generation step> --permission-reference <reference> <PROJECT>",
                "moai apply-recover --receipt <receipt of the generation step> --permission-reference <reference> <PROJECT>",
            ], "note": "the instruction file has a receipt of its own; undo it before the workflow files"},
            {"step": "go-state", "commands": [
                "moai go-state-revert-plan <PROJECT> --backup <backup of the go-state step>",
                "moai go-state-revert --backup <backup of the go-state step> --plan-digest <digest of that plan> --permission-reference <reference> <PROJECT>",
            ], "note": "refused while the generation step's AGENTS.md is in place"},
            {"step": "go-settings", "commands": [
                "moai go-settings-revert --receipt <receipt of the go-settings step> --permission-reference <reference> <PROJECT>",
            ]},
        ],
        "stays": "what the commands above do not take away: the receipts under .moai-cli/receipts/, the Rust state's set-aside folder .moai.rust-<UTC>/, and what the three steps without a revert command wrote (.codex/, .omp/, .moai-cli/manifests/codex-config.json, codex-hooks.json, omp-hooks.json); remove those by hand only after reading them, nothing here deletes them",
        "other_steps": "a step not listed (profile, claude-settings, github, the merge steps) prints its own revert or receipt in its apply result",
    })
}

/// The whole preview. Writes nothing.
pub fn plan(project: &Path) -> Result<Value> {
    let root = fs::canonicalize(project).map_err(|_| crate::contract::PROJECT_PATH_UNAVAILABLE)?;
    let shown = root.to_str().ok_or("non UTF-8 project path")?.to_string();
    let detected = project_detect::detect(&root);
    // #2936 L2a: `go-settings` first. Its apply rewrites `.claude/settings.json`, which the
    // go-state plan reads; the go-state step with a command of its own waits until a new
    // preview is made after it.
    let moves = go_state::plan(&root);
    let settings = or_blocked(
        "go-settings",
        go_settings_step(&root, &shown, moves.as_ref().ok()),
    );
    let mut state = match &moves {
        Ok(p) => or_blocked("go-state", go_state_step(&root, &shown, p, &settings)),
        Err(e) => Step::waiting("go-state", "blocked", e),
    };
    if settings.apply.is_some() && state.apply.is_some() {
        state = Step::waiting("go-state", "blocked", AFTER_GO_SETTINGS);
    }
    // C-126: `reports` last. Moving the untracked report files out changes the go-state plan's
    // report listing, so a go-state command waits until a new preview is made after it.
    let reports = or_blocked("reports", reports_step(&root, &shown));
    if reports.apply.is_some() && state.apply.is_some() {
        state = Step::waiting("go-state", "blocked", AFTER_REPORTS);
    }
    let backup = vec![settings, state, reports];
    // A move still to make: every install preview would be stale after it.
    let install = if backup_pending(&backup) {
        INSTALL
            .map(|n| Step::waiting(n, "blocked", AFTER_BACKUP))
            .into_iter()
            .collect()
    } else {
        install_steps(&root, &shown, &detected)
    };
    let merge = merge_steps(&root, &shown, &backup, &install);
    let all_unchanged = backup
        .iter()
        .chain(&install)
        .chain(&merge)
        .all(|s| s.state == "unchanged" || s.state == "skip");
    let phase = |name: &str, reason: &str, steps: &[Step]| {
        json!({
            "phase": name,
            "state": worst(steps),
            "reason": reason,
            "steps": steps.iter().map(Step::to_json).collect::<Vec<_>>(),
        })
    };
    let classified = state_root::classify(&root);
    Ok(json!({
        "project": shown,
        "detected": detected.ok(),
        "state_root": classified.to_json(),
        "leftovers": leftovers(&root, classified.generation),
        "phases": [
            phase("backup", BACKUP_REASON, &backup),
            phase("install", INSTALL_REASON, &install),
            phase("merge", MERGE_REASON, &merge),
        ],
        "all_unchanged": all_unchanged,
        "undo_order": undo_order(),
    }))
}

#[cfg(test)]
mod tests {
    use super::*;

    /// AUDIT #3117 E2: a workflow reference recorded in the earlier state folder follows the state
    /// to `.moai/` once it has moved, and only then; any other reference is as recorded.
    #[test]
    fn a_workflow_reference_in_the_earlier_state_folder_follows_the_state() {
        let tmp = tempfile::tempdir().unwrap();
        let root = tmp.path();
        let earlier = format!("{}/workflows/moai/", state_root::EARLIER_PROJECT_DIR);
        let moved = format!("{}/workflows/moai/", state_root::MOAI_DIR);
        // The state has not moved: as recorded.
        assert_eq!(follow_state_dir(root, &earlier), earlier);
        fs::create_dir_all(root.join(state_root::MOAI_DIR)).unwrap();
        fs::write(
            root.join(state_root::MOAI_DIR)
                .join(state_root::RUST_MARKER),
            state_root::marker_bytes("test", &[]),
        )
        .unwrap();
        assert_eq!(follow_state_dir(root, &earlier), moved);
        // Anything that is not that prefix is left alone, look-alikes included.
        for other in [
            format!("{}/workflows/x/", state_root::MOAI_DIR),
            format!("docs/{earlier}"),
            format!("{}x/workflows/", state_root::EARLIER_PROJECT_DIR),
            "workflows/moai/".to_string(),
            state_root::EARLIER_PROJECT_DIR.to_string(),
        ] {
            assert_eq!(follow_state_dir(root, &other), other);
        }
    }

    #[test]
    fn a_phase_takes_its_worst_step() {
        let s = |state| Step::waiting("x", state, "");
        assert_eq!(worst(&[s("unchanged"), s("skip")]), "skip");
        assert_eq!(worst(&[s("ready"), s("blocked"), s("conflict")]), "blocked");
        assert_eq!(worst(&[s("needs_decision"), s("ready")]), "needs_decision");
        assert_eq!(worst(&[]), "unchanged");
    }

    /// #2953 F6: a conflict whose plan offers `confirm_replace` prints that replacement — the
    /// argv with `--confirm-replace <token>` and the owned keys on stdin — as a decision; a
    /// conflict without it prints nothing.
    #[test]
    fn a_replaceable_settings_conflict_prints_the_confirmed_apply() {
        let owned: Map<String, Value> =
            serde_json::from_value(json!({"outputStyle": "Explanatory"})).unwrap();
        let mut plan = json!({
            "action": "conflict", "plan_digest": "sha256:d",
            "settings_path": ".claude/settings.json", "conflicts": ["outputStyle"],
        });
        let kept = settings_from_plan("/p", &plan, owned.clone());
        assert_eq!((kept.state, kept.apply.is_none()), ("conflict", true));

        plan["confirm_replace"] =
            json!({"keys": ["outputStyle"], "confirm_token": "settings-replace:project:t"});
        let step = settings_from_plan("/p", &plan, owned.clone());
        assert_eq!(step.state, "needs_decision");
        let (argv, stdin) = step.apply.expect("a printed command");
        assert_eq!(
            argv,
            [
                "moai",
                "claude-settings-apply",
                "/p",
                "--plan-digest",
                "sha256:d",
                "--permission-reference",
                PERMISSION_PLACEHOLDER,
                "--confirm-replace",
                "settings-replace:project:t",
            ]
        );
        assert_eq!(stdin, Value::Object(owned));
    }

    /// #2953 F5: the printed settings stdin carries the owned keys without `hooks`.
    #[test]
    fn the_settings_stdin_has_no_hooks() {
        let owned: Map<String, Value> =
            serde_json::from_value(json!({"outputStyle": "Explanatory", "hooks": {"Stop": []}}))
                .unwrap();
        assert_eq!(
            Value::Object(settings_stdin(owned)),
            json!({"outputStyle": "Explanatory"})
        );
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
