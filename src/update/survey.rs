//! The update survey (SPEC-20261120-update-guided-migration, REQ-1 to REQ-3, REQ-11): one
//! read-only pass over every layer `moai update` and the project update can change, put into
//! one list of items. Every item has an id, a plain title, a layer, a group (`will_change`,
//! `kept` or `needs_pick`), a state, how many entries it covers, a guarded flag, the ids it
//! depends on, the printed apply command when there is one, and its revert: a command, the
//! statement that the apply prints it, or the statement that none exists.
//!
//! Nothing here reads more than the plans the update already makes (the user layer, the
//! HOME move preview, the project update plan) and nothing writes: the survey is a view over
//! those plans, so what it shows is what the apply code would act on. One digest binds the
//! component plans (REQ-11): a plan that changes between the preview and the answer changes it.
//!
//! The apply order is the order of [`ORDER`] (REQ-6): the HOME retire before the HOME move,
//! the move before the other HOME items, the HOME items before the project's, the project's
//! backup phase before its install phase.
//!
//! Left out on purpose: the model questions. They are asked after an apply and are no migration
//! step, so they never keep a survey from reporting nothing to apply.

use super::{agents_write_anything, plan_writes_anything, routing_writes_anything, UserLayerPlan};
use serde_json::{json, Value};
use std::path::Path;

pub const WILL_CHANGE: &str = "will_change";
pub const KEPT: &str = "kept";
pub const NEEDS_PICK: &str = "needs_pick";

/// The ids in apply order. The HOME items first, then the project steps in the order of the
/// project update plan's phases; `project-plan` is the one item for a project plan that could
/// not be read.
pub const ORDER: [&str; 28] = [
    "home-retire",
    "home-move",
    "skills",
    "agents",
    "routing",
    "style",
    "command-names",
    "hooks-here",
    "keys",
    "cleanup",
    "switch",
    "models",
    "project-go-settings",
    "project-go-state",
    "project-profile",
    "project-generation",
    "project-claude-settings",
    "project-hooks",
    "project-github",
    "project-codex-hooks",
    "project-codex-config",
    "project-codex-skill",
    "project-omp-hooks",
    "project-omp-config",
    "project-omp-agents",
    "project-agents-local-merge",
    "project-claude-settings-merge",
    "project-plan",
];

/// How an item is taken back.
enum Revert {
    /// The command, with its placeholders.
    Command(&'static str),
    /// The apply result prints the revert.
    AfterApply(&'static str),
    /// No revert command exists; the text says what to know instead.
    None(&'static str),
}

struct Item {
    id: String,
    title: &'static str,
    layer: &'static str,
    group: &'static str,
    state: String,
    entries: u64,
    guarded: bool,
    depends_on: Vec<String>,
    reason: Option<String>,
    apply: Option<String>,
    revert: Revert,
}

impl Item {
    fn to_json(&self, order: usize) -> Value {
        let revert = match &self.revert {
            Revert::Command(c) => json!({"command": c}),
            Revert::AfterApply(t) => json!({"after_apply": t}),
            Revert::None(t) => json!({"none": t}),
        };
        json!({
            "id": self.id,
            "title": self.title,
            "layer": self.layer,
            "group": self.group,
            "state": self.state,
            "count": {"entries": self.entries},
            "guarded": self.guarded,
            "depends_on": self.depends_on,
            "reason": self.reason,
            "apply": self.apply,
            "revert": revert,
            "order": order,
        })
    }
}

/// What the survey reads besides the user layer. Reads only.
pub struct Gathered {
    /// The `~/.moai-cli` to `~/.moai` move preview (`home_install::plan`).
    pub home_move: Result<Value, String>,
    /// The project update plan of the work tree the command runs in; `None` outside one.
    pub project: Option<Result<Value, String>>,
}

/// Reads the plans the user layer does not carry. Writes nothing.
pub fn gather(home: &Path, layer: &UserLayerPlan) -> Gathered {
    let moai_home = std::env::var("MOAI_HOME").ok();
    // The move plan reads and digests the whole earlier state folder, so it is made only when
    // there is one: after the move (the usual state) the survey costs nothing here.
    let home_move = if old_folder_absent(home) {
        Ok(json!({
            "state_dir": "absent", "action": "none", "blocks": [], "counts": {"move": 0},
            "content_digest": null,
        }))
    } else {
        crate::home_install::plan(home, moai_home.as_deref(), &crate::audit::utc_now())
    };
    Gathered {
        home_move,
        project: layer
            .hooks
            .project
            .as_deref()
            .map(crate::project_update::plan),
    }
}

/// True when `~/.moai-cli` is plainly not there (not found); a folder that cannot be read is
/// not "absent" and goes to the move plan, which names it.
fn old_folder_absent(home: &Path) -> bool {
    matches!(
        std::fs::symlink_metadata(crate::home_install::old_root(home)),
        Err(e) if e.kind() == std::io::ErrorKind::NotFound
    )
}

/// Changes win over picks: a step that has something to apply and something held back is
/// `will_change` (the held part is named in its reason); only a step with nothing to apply and
/// something held back needs a pick.
fn decide(changes: usize, picks: usize, held: &str) -> (&'static str, &'static str) {
    if changes > 0 {
        (WILL_CHANGE, "ready")
    } else if picks > 0 {
        (NEEDS_PICK, held_state(held))
    } else {
        (KEPT, "unchanged")
    }
}

fn held_state(held: &str) -> &'static str {
    if held == "conflict" {
        "conflict"
    } else {
        "blocked"
    }
}

fn picks_note(changes: usize, picks: usize) -> Option<String> {
    (changes > 0 && picks > 0).then(|| {
        format!("{picks} host(s) or file(s) are held back and need a pick; they are not part of this step")
    })
}

fn confirm_flag(layer: &UserLayerPlan) -> String {
    format!("moai update --confirm {}", layer.plan_digest)
}

fn skills_item(layer: &UserLayerPlan) -> Item {
    let hosts = &layer.skills.hosts;
    let changes = hosts
        .iter()
        .filter(|h| h.preflight != "blocked")
        .flat_map(|h| &h.files)
        .filter(|f| matches!(f.action.as_str(), "create" | "update"))
        .count();
    let blocked = hosts.iter().filter(|h| h.preflight == "blocked").count();
    let conflicts = hosts
        .iter()
        .flat_map(|h| &h.files)
        .filter(|f| f.action == "conflict")
        .count();
    let picks = blocked + conflicts;
    let (group, state) = decide(
        changes,
        picks,
        if conflicts > 0 { "conflict" } else { "blocked" },
    );
    let reason = picks_note(changes, picks).or_else(|| {
        hosts
            .iter()
            .find_map(|h| h.blocked_reason.clone())
            .filter(|_| group == NEEDS_PICK)
    });
    Item {
        id: "skills".into(),
        title: "Skill copies in the Claude and Codex folders",
        layer: "home",
        group,
        state: state.into(),
        entries: changes as u64,
        guarded: false,
        depends_on: vec![],
        reason,
        apply: (changes > 0).then(|| confirm_flag(layer)),
        revert: Revert::None(
            "skill copies have no revert command; a host left pending is repaired with recovery-plan and recovery-apply",
        ),
    }
}

fn agents_item(layer: &UserLayerPlan) -> Item {
    let hosts = &layer.agents.hosts;
    let changes = hosts
        .iter()
        .filter(|h| h.preflight == "ok")
        .flat_map(|h| &h.files)
        .filter(|f| matches!(f.action.as_str(), "create" | "update"))
        .count();
    let blocked = hosts.iter().filter(|h| h.preflight != "ok").count();
    let conflicts = hosts
        .iter()
        .flat_map(|h| &h.files)
        .filter(|f| matches!(f.action.as_str(), "conflict" | "remove_blocked"))
        .count();
    let picks = blocked + conflicts;
    let (group, state) = decide(
        changes,
        picks,
        if conflicts > 0 { "conflict" } else { "blocked" },
    );
    debug_assert_eq!(agents_write_anything(&layer.agents), changes > 0);
    Item {
        id: "agents".into(),
        title: "Base agents in the Claude and Codex folders",
        layer: "home",
        group,
        state: state.into(),
        entries: changes as u64,
        guarded: false,
        depends_on: vec!["skills".into()],
        reason: picks_note(changes, picks),
        apply: (changes > 0).then(|| confirm_flag(layer)),
        revert: Revert::Command(
            "moai agents-home-revert <HOME> --receipt <id> --permission-reference <ref>",
        ),
    }
}

fn routing_item(layer: &UserLayerPlan) -> Item {
    let hosts = &layer.routing.hosts;
    let action = |h: &crate::home_routing::HostPlan| {
        h.section
            .as_ref()
            .map(|s| s.action.clone())
            .unwrap_or_default()
    };
    let changes = hosts
        .iter()
        .filter(|h| h.preflight == "ok")
        .filter(|h| matches!(action(h).as_str(), "create" | "append" | "merge"))
        .count();
    let blocked = hosts.iter().filter(|h| h.preflight != "ok").count();
    let conflicts = hosts.iter().filter(|h| action(h) == "conflict").count();
    let picks = blocked + conflicts;
    let (group, state) = decide(
        changes,
        picks,
        if conflicts > 0 { "conflict" } else { "blocked" },
    );
    debug_assert_eq!(routing_writes_anything(&layer.routing), changes > 0);
    Item {
        id: "routing".into(),
        title: "The MoAI routing block in the global instruction files",
        layer: "home",
        group,
        state: state.into(),
        entries: changes as u64,
        guarded: false,
        depends_on: vec!["skills".into()],
        reason: picks_note(changes, picks),
        apply: (changes > 0).then(|| confirm_flag(layer)),
        revert: Revert::Command(
            "moai home-routing-revert <HOME> --receipt <id> --permission-reference <ref>",
        ),
    }
}

fn style_item(layer: &UserLayerPlan) -> Item {
    let targets = &layer.style.targets;
    let changes = targets
        .iter()
        .filter(|t| matches!(t.action.as_str(), "create" | "append" | "merge" | "set"))
        .count();
    let picks = targets
        .iter()
        .filter(|t| matches!(t.action.as_str(), "conflict" | "blocked"))
        .count();
    let (group, state) = decide(changes, picks, "conflict");
    Item {
        id: "style".into(),
        title: "The MoAI-Easy output style",
        layer: "home",
        group,
        state: state.into(),
        entries: changes as u64,
        guarded: false,
        depends_on: vec!["skills".into()],
        reason: picks_note(changes, picks),
        apply: (changes > 0).then(|| confirm_flag(layer)),
        revert: Revert::Command(
            "moai home-style-revert <HOME> --receipt <id> --permission-reference <ref>",
        ),
    }
}

fn command_names_item(layer: &UserLayerPlan) -> Item {
    let p = &layer.bin_names;
    let (group, state) = match p.state {
        "create" | "replace_copy" => (WILL_CHANGE, "ready"),
        "blocked" => (NEEDS_PICK, "blocked"),
        _ => (KEPT, "unchanged"),
    };
    Item {
        id: "command-names".into(),
        title: "The moai command name (a link to moai-cli)",
        layer: "home",
        group,
        state: if group == KEPT && p.state != "unchanged" {
            p.state.to_string()
        } else {
            state.to_string()
        },
        entries: u64::from(p.writes()),
        guarded: false,
        depends_on: vec![],
        reason: p.reason.clone(),
        apply: p.writes().then(|| confirm_flag(layer)),
        revert: Revert::None("the command name has no revert command"),
    }
}

fn hooks_here_item(layer: &UserLayerPlan) -> Option<Item> {
    let project = layer.hooks.project.as_ref()?;
    if layer.hooks.hosts.is_empty() {
        return None;
    }
    let changes = layer
        .hooks
        .hosts
        .iter()
        .filter(|h| h.plan_digest.is_some() && matches!(h.action.as_str(), "create" | "update"))
        .count();
    let picks = layer
        .hooks
        .hosts
        .iter()
        .filter(|h| h.action == "conflict")
        .count();
    let (group, state) = decide(changes, picks, "conflict");
    Some(Item {
        id: "hooks-here".into(),
        title: "The MoAI hooks already installed in this project",
        layer: "home",
        group,
        state: state.into(),
        entries: changes as u64,
        guarded: false,
        depends_on: vec![],
        reason: picks_note(changes, picks)
            .or_else(|| Some(format!("project {}", project.display()))),
        apply: (changes > 0).then(|| confirm_flag(layer)),
        revert: Revert::Command(
            "moai hooks-revert <PROJECT> --receipt <id> --permission-reference <ref>",
        ),
    })
}

fn keys_item(layer: &UserLayerPlan) -> Option<Item> {
    if layer.keys.rows.is_empty() {
        return None;
    }
    let imports = layer
        .keys
        .rows
        .iter()
        .filter(|r| r.action == "import")
        .count();
    let (group, state) = if layer.keys.writes() {
        (WILL_CHANGE, "ready")
    } else {
        (KEPT, "unchanged")
    };
    Some(Item {
        id: "keys".into(),
        title: "The glm key file moves into the key vault",
        layer: "home",
        group,
        state: state.into(),
        entries: imports as u64,
        guarded: true,
        depends_on: vec![],
        reason: layer
            .keys
            .rows
            .iter()
            .find_map(|r| r.reason.clone())
            .filter(|_| group == KEPT),
        apply: layer.keys.writes().then(|| confirm_flag(layer)),
        revert: Revert::Command(
            "moai keys-revert <HOME> --receipt <keys-id> --permission-reference <ref>",
        ),
    })
}

fn cleanup_item(layer: &UserLayerPlan) -> Option<Item> {
    let c = layer.cleanup.as_ref()?;
    let token = c["confirm_token"].as_str();
    let (group, state) = if token.is_some() {
        (WILL_CHANGE, "ready")
    } else {
        (KEPT, "unchanged")
    };
    Some(Item {
        id: "cleanup".into(),
        title: "Go leftovers move into the cleanup archive",
        layer: "home",
        group,
        state: state.into(),
        entries: c["counts"]["archive"].as_u64().unwrap_or(0),
        guarded: true,
        depends_on: vec![],
        reason: None,
        apply: token.map(|t| format!("moai update --confirm-cleanup {t}")),
        revert: Revert::Command(
            "moai cleanup-revert-plan <HOME> --stamp <stamp>, then moai cleanup-revert <HOME> --stamp <stamp> --permission-reference <ref>",
        ),
    })
}

fn switch_item(layer: &UserLayerPlan) -> Option<Item> {
    let p = layer.switch.to_json()?;
    let token = layer.switch.token();
    let (group, state) = if token.is_some() {
        (WILL_CHANGE, "ready")
    } else {
        (KEPT, "unchanged")
    };
    Some(Item {
        id: "switch".into(),
        title: "The name-switch script of the moai command",
        layer: "home",
        group,
        state: state.into(),
        entries: u64::from(token.is_some()),
        guarded: false,
        depends_on: vec![],
        reason: p["summary"].as_str().map(str::to_string),
        apply: token.map(|t| format!("moai update --confirm-switch {t}")),
        revert: Revert::Command("moai-cli update --revert-switch <stamp>"),
    })
}

fn retire_item(layer: &UserLayerPlan) -> Option<Item> {
    let p = layer.home.to_json()?;
    let action = p["action"].as_str().unwrap_or("none");
    let (group, state) = match action {
        "retire" => (WILL_CHANGE, "ready"),
        "blocked" => (NEEDS_PICK, "blocked"),
        _ => (KEPT, "unchanged"),
    };
    let blocks: Vec<&str> = p["blocks"]
        .as_array()
        .into_iter()
        .flatten()
        .filter_map(Value::as_str)
        .collect();
    Some(Item {
        id: "home-retire".into(),
        title: "The earlier generation's items in the HOME state folder move aside",
        layer: "home",
        group,
        state: state.into(),
        entries: p["counts"]["retire"].as_u64().unwrap_or(0),
        guarded: true,
        depends_on: vec![],
        reason: (!blocks.is_empty()).then(|| blocks.join(", ")),
        apply: p["apply_flag"].as_str().map(|f| format!("moai update {f}")),
        revert: Revert::Command(
            "moai home-retire-revert --home <HOME> --stamp <stamp> --permission-reference <ref>",
        ),
    })
}

fn home_move_item(plan: &Result<Value, String>, retire: bool) -> Option<Item> {
    const TITLE: &str = "The HOME state folder moves from ~/.moai-cli to ~/.moai";
    const REVERT: Revert =
        Revert::Command("moai home-install-revert --home <HOME> --permission-reference <ref>");
    let p = match plan {
        Err(e) => {
            return Some(Item {
                id: "home-move".into(),
                title: TITLE,
                layer: "home",
                group: NEEDS_PICK,
                state: "unread".into(),
                entries: 0,
                guarded: true,
                depends_on: vec![],
                reason: Some(format!("the plan could not be read: {e}")),
                apply: None,
                revert: REVERT,
            })
        }
        Ok(p) => p,
    };
    if p["state_dir"] == json!("absent") {
        return None;
    }
    let blocks: Vec<&str> = p["blocks"]
        .as_array()
        .into_iter()
        .flatten()
        .filter_map(Value::as_str)
        .collect();
    let waits = blocks.iter().any(|b| b.starts_with("go_items_not_retired"));
    let other: Vec<&str> = blocks
        .iter()
        .copied()
        .filter(|b| !b.starts_with("go_items_not_retired"))
        .collect();
    let (group, state, reason) = match p["action"].as_str().unwrap_or("none") {
        "install" => (WILL_CHANGE, "ready", None),
        "blocked" if waits && other.is_empty() => (
            WILL_CHANGE,
            "waiting",
            Some("waits for the retire of the earlier generation's items".to_string()),
        ),
        "blocked" => (NEEDS_PICK, "blocked", Some(other.join(", "))),
        _ => (KEPT, "unchanged", None),
    };
    Some(Item {
        id: "home-move".into(),
        title: TITLE,
        layer: "home",
        group,
        state: state.into(),
        entries: p["counts"]["move"].as_u64().unwrap_or(0),
        guarded: true,
        depends_on: if retire { vec!["home-retire".into()] } else { vec![] },
        reason,
        apply: (group == WILL_CHANGE && state == "ready").then(|| {
            "moai home-install-apply --home <HOME> --now <now> --plan-digest <digest> --permission-reference <ref>".to_string()
        }),
        revert: REVERT,
    })
}

/// The title, guard and revert of a project step, from the detection table of the SPEC.
fn project_facts(step: &str) -> (&'static str, bool, Revert) {
    match step {
        "go-settings" => (
            "Go settings in .claude/settings.json are cleaned",
            false,
            Revert::Command(
                "moai go-settings-revert --receipt <id> --permission-reference <ref> <PROJECT>",
            ),
        ),
        "go-state" => (
            "The project's earlier state folder moves aside and the new .moai folder starts",
            true,
            Revert::Command(
                "moai go-state-revert-plan <PROJECT> --backup <UTC>, then moai go-state-revert <PROJECT> --backup <UTC> --plan-digest <digest> --permission-reference <ref>",
            ),
        ),
        "profile" => (
            "The project profile and managed files",
            false,
            Revert::AfterApply("the apply result of project-apply prints its receipt"),
        ),
        "generation" => (
            "The project agents, skills and the AGENTS.md managed block",
            false,
            Revert::Command(
                "moai apply-recover --receipt <agents_receipt> --permission-reference <ref> <PROJECT>, then moai apply-recover --receipt <receipt> --permission-reference <ref> <PROJECT>",
            ),
        ),
        "claude-settings" => (
            "The keys MoAI owns in the project's Claude settings",
            false,
            Revert::AfterApply("the old file is kept in settings-backups; the apply result prints the copy-back command"),
        ),
        "hooks" => (
            "The project's Go hooks become the Rust hooks",
            false,
            Revert::Command(
                "moai go-hooks-revert --receipt <receipt> --permission-reference <ref> <PROJECT>, or moai hooks-revert <PROJECT> --receipt <id> --permission-reference <ref> after hooks-apply",
            ),
        ),
        "github" => (
            "The GitHub templates",
            false,
            Revert::AfterApply("the apply result of github-apply prints its receipt"),
        ),
        "codex-hooks" => (
            "The project's Codex hooks",
            false,
            Revert::Command(
                "moai hooks-revert <PROJECT> --receipt <id> --permission-reference <ref>",
            ),
        ),
        "codex-config" => (
            "The Codex status line in the project's config",
            false,
            Revert::None("no revert command exists for this step"),
        ),
        "codex-skill" => (
            "The Codex skill copy in the project",
            false,
            Revert::None("no revert command exists for this step"),
        ),
        "omp-hooks" => (
            "The project's OMP hooks",
            false,
            Revert::Command(
                "moai hooks-revert <PROJECT> --receipt <id> --permission-reference <ref>",
            ),
        ),
        "omp-config" => (
            "The OMP config",
            false,
            Revert::None("this step never writes"),
        ),
        "omp-agents" => (
            "The project's OMP agents",
            false,
            Revert::None("no revert command exists for this step; nothing is ever deleted"),
        ),
        "agents-local-merge" => (
            "AGENTS.local.md joins the AGENTS.md managed block",
            false,
            Revert::Command(
                "moai agents-merge-revert <PROJECT> --receipt <id> --permission-reference <ref>",
            ),
        ),
        "claude-settings-merge" => (
            "Your own Claude settings keys come back one by one",
            false,
            Revert::AfterApply("the apply result prints the copy-back command"),
        ),
        _ => (
            "A project update step",
            false,
            Revert::None("this step is not in the detection table; it has no known revert"),
        ),
    }
}

fn project_entries(preview: &Value) -> u64 {
    for key in [
        "files",
        "entries",
        "hosts",
        "moves",
        "state_moves",
        "keys",
        "candidates",
        "targets",
    ] {
        if let Some(a) = preview[key].as_array() {
            return a.len() as u64;
        }
    }
    0
}

fn project_items(plan: &Result<Value, String>) -> Vec<Item> {
    let p = match plan {
        Err(e) => {
            return vec![Item {
                id: "project-plan".into(),
                title: "The project update plan",
                layer: "project",
                group: NEEDS_PICK,
                state: "unread".into(),
                entries: 0,
                guarded: false,
                depends_on: vec![],
                reason: Some(format!("the plan could not be read: {e}")),
                apply: None,
                revert: Revert::None("nothing was applied"),
            }]
        }
        Ok(p) => p,
    };
    let mut out = Vec::new();
    for phase in p["phases"].as_array().into_iter().flatten() {
        let phase_name = phase["phase"].as_str().unwrap_or("");
        for s in phase["steps"].as_array().into_iter().flatten() {
            let name = s["step"].as_str().unwrap_or("");
            let reason = s["reason"].as_str().unwrap_or("");
            let state = s["state"].as_str().unwrap_or("unchanged");
            let (group, state_out) = match state {
                "unchanged" | "skip" => (KEPT, state),
                "ready" => (WILL_CHANGE, "ready"),
                _ if reason.starts_with("re-run after") => (WILL_CHANGE, "waiting"),
                other => (NEEDS_PICK, other),
            };
            let (title, guarded, revert) = project_facts(name);
            let depends_on = match (phase_name, name) {
                (_, "go-settings") => vec![],
                (_, "go-state") => vec!["project-go-settings".to_string()],
                ("install", _) => vec!["project-go-state".to_string()],
                ("merge", _) => vec!["project-generation".to_string()],
                _ => vec![],
            };
            let mut entries = project_entries(&s["preview"]);
            if group == WILL_CHANGE && state_out == "ready" && entries == 0 {
                entries = 1;
            }
            out.push(Item {
                id: format!("project-{name}"),
                title,
                layer: "project",
                group,
                state: state_out.to_string(),
                entries,
                guarded,
                depends_on,
                reason: (!reason.is_empty()).then(|| reason.to_string()),
                apply: s["apply"]["argv"].as_array().map(|a| {
                    a.iter()
                        .filter_map(Value::as_str)
                        .collect::<Vec<_>>()
                        .join(" ")
                }),
                revert,
            });
        }
    }
    out
}

/// A survey has nothing to apply when no item will change; items that only wait for a pick
/// leave a second run with nothing to apply (REQ-9).
fn nothing_to_apply(items: &[Item]) -> bool {
    items.iter().all(|i| i.group != WILL_CHANGE)
}

fn position(id: &str) -> usize {
    ORDER.iter().position(|x| *x == id).unwrap_or(ORDER.len())
}

/// The survey of one update preview: items, groups, components and the digest. Pure over its
/// inputs.
pub fn build(layer: &UserLayerPlan, g: &Gathered) -> Value {
    let mut items: Vec<Item> = Vec::new();
    let retire = retire_item(layer);
    let has_retire = retire.is_some();
    items.extend(retire);
    let mv = home_move_item(&g.home_move, has_retire);
    let moving = mv
        .as_ref()
        .is_some_and(|m| m.group == WILL_CHANGE || m.state == "unread");
    items.extend(mv);
    items.push(skills_item(layer));
    items.push(agents_item(layer));
    items.push(routing_item(layer));
    items.push(style_item(layer));
    items.push(command_names_item(layer));
    items.extend(hooks_here_item(layer));
    items.extend(keys_item(layer));
    items.extend(cleanup_item(layer));
    items.extend(switch_item(layer));
    if let Some(project) = &g.project {
        items.extend(project_items(project));
    }
    // The items that live in the HOME state folder are surveyed again once it has moved.
    if moving {
        for i in &mut items {
            if matches!(i.id.as_str(), "keys" | "cleanup" | "switch") {
                i.depends_on.push("home-move".into());
            }
        }
    }
    items.sort_by_key(|i| position(&i.id));
    let items_json: Vec<Value> = items
        .iter()
        .enumerate()
        .map(|(n, i)| i.to_json(n + 1))
        .collect();
    let ids_of = |group: &str| -> Vec<&str> {
        items
            .iter()
            .filter(|i| i.group == group)
            .map(|i| i.id.as_str())
            .collect()
    };
    let components = components(layer, g);
    let survey_digest = crate::digest(
        serde_json::to_string(&components)
            .unwrap_or_default()
            .as_bytes(),
    );
    json!({
        "survey_digest": survey_digest,
        "components": components,
        "items": items_json,
        "groups": {
            WILL_CHANGE: ids_of(WILL_CHANGE),
            KEPT: ids_of(KEPT),
            NEEDS_PICK: ids_of(NEEDS_PICK),
        },
        "nothing_to_apply": nothing_to_apply(&items),
        "writes": false,
    })
}

/// One digest per component plan; the survey digest binds all of them (REQ-11).
fn components(layer: &UserLayerPlan, g: &Gathered) -> Value {
    let text = |v: &Value| crate::digest(serde_json::to_string(v).unwrap_or_default().as_bytes());
    let home_move = match &g.home_move {
        Ok(p) => p["content_digest"].as_str().unwrap_or("none").to_string(),
        Err(e) => format!("unread:{e}"),
    };
    let project = match &g.project {
        None => "none".to_string(),
        Some(Err(e)) => format!("unread:{e}"),
        Some(Ok(p)) => {
            let steps: Vec<Value> = p["phases"]
                .as_array()
                .into_iter()
                .flatten()
                .flat_map(|ph| ph["steps"].as_array().into_iter().flatten())
                .map(|s| json!([s["step"], s["state"], s["plan_digest"], s["reason"]]))
                .collect();
            text(&json!(steps))
        }
    };
    json!([
        {"id": "user-layer", "digest": layer.plan_digest},
        {"id": "home-retire", "digest": layer.home.to_json().map_or_else(|| "none".to_string(), text)},
        {"id": "cleanup", "digest": layer.cleanup.as_ref().and_then(|c| c["confirm_token"].as_str()).unwrap_or("none")},
        {"id": "switch", "digest": layer.switch.token().unwrap_or("none")},
        {"id": "home-move", "digest": home_move},
        {"id": "project", "digest": project},
    ])
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn the_order_has_no_duplicate_and_the_home_move_follows_the_retire() {
        let mut seen = std::collections::BTreeSet::new();
        assert!(
            ORDER.iter().all(|i| seen.insert(*i)),
            "an id is listed twice"
        );
        assert!(position("home-retire") < position("home-move"));
        assert!(position("home-move") < position("skills"));
        assert!(position("project-go-settings") < position("project-go-state"));
        assert!(position("project-go-state") < position("project-profile"));
        assert!(position("switch") < position("project-hooks"));
        assert!(position("project-generation") < position("project-agents-local-merge"));
    }

    /// AC-2 (REQ-2): a plan that could not be read is named unread with its reason, needs a
    /// pick and applies nothing; the layers beside it are still listed.
    fn item_in(group: &'static str) -> Item {
        Item {
            id: "skills".into(),
            title: "x",
            layer: "home",
            group,
            state: "x".into(),
            entries: 0,
            guarded: false,
            depends_on: vec![],
            reason: None,
            apply: None,
            revert: Revert::None("x"),
        }
    }

    #[test]
    fn changes_win_over_picks_and_picks_win_over_nothing() {
        assert_eq!(decide(2, 3, "blocked"), (WILL_CHANGE, "ready"));
        assert_eq!(decide(0, 3, "blocked"), (NEEDS_PICK, "blocked"));
        assert_eq!(decide(0, 1, "conflict"), (NEEDS_PICK, "conflict"));
        assert_eq!(decide(0, 0, "blocked"), (KEPT, "unchanged"));
        assert!(picks_note(1, 2).is_some());
        assert!(picks_note(0, 2).is_none());
        assert!(picks_note(2, 0).is_none());
    }

    #[test]
    fn nothing_to_apply_only_when_no_item_will_change() {
        assert!(nothing_to_apply(&[]));
        assert!(nothing_to_apply(&[item_in(KEPT), item_in(NEEDS_PICK)]));
        assert!(!nothing_to_apply(&[item_in(KEPT), item_in(WILL_CHANGE)]));
    }

    #[cfg(unix)]
    #[test]
    fn a_home_that_cannot_be_read_is_not_an_absent_old_folder() {
        use std::os::unix::fs::PermissionsExt;
        let tmp = tempfile::tempdir().unwrap();
        let home = tmp.path().join("h");
        std::fs::create_dir(&home).unwrap();
        std::fs::create_dir(home.join(crate::HOME_STATE_DIR)).unwrap();
        std::fs::set_permissions(&home, std::fs::Permissions::from_mode(0o000)).unwrap();
        let unreadable = !old_folder_absent(&home);
        std::fs::set_permissions(&home, std::fs::Permissions::from_mode(0o755)).unwrap();
        // Run as root the folder stays readable and the check proves nothing: skip then.
        let readable = std::fs::symlink_metadata(home.join(crate::HOME_STATE_DIR)).is_ok();
        assert!(readable, "the folder is back");
        let as_root = std::fs::read_dir(&home).is_ok() && {
            std::fs::set_permissions(&home, std::fs::Permissions::from_mode(0o000)).unwrap();
            let r = std::fs::read_dir(&home).is_ok();
            std::fs::set_permissions(&home, std::fs::Permissions::from_mode(0o755)).unwrap();
            r
        };
        assert!(
            unreadable || as_root,
            "an unreadable HOME was taken for an absent old folder"
        );
    }

    #[test]
    fn the_move_plan_is_skipped_only_when_the_old_folder_is_not_found() {
        let tmp = tempfile::tempdir().unwrap();
        assert!(old_folder_absent(tmp.path()));
        std::fs::create_dir(tmp.path().join(crate::HOME_STATE_DIR)).unwrap();
        assert!(!old_folder_absent(tmp.path()));
    }

    #[test]
    fn unread_layer_is_named() {
        let home = home_move_item(&Err("the folder cannot be listed".into()), false).unwrap();
        assert_eq!(home.state, "unread");
        assert_eq!(home.group, NEEDS_PICK);
        assert!(home.apply.is_none());
        assert!(home.reason.as_deref().unwrap().contains("cannot be listed"));
        let project = project_items(&Err("not a project".into()));
        assert_eq!(project.len(), 1);
        assert_eq!(project[0].id, "project-plan");
        assert_eq!(project[0].state, "unread");
        assert_eq!(project[0].group, NEEDS_PICK);
        assert!(project[0].apply.is_none());
    }

    #[test]
    fn a_step_that_waits_is_will_change_waiting_and_a_blocked_one_needs_a_pick() {
        let plan = Ok(json!({"phases": [
            {"phase": "backup", "steps": [
                {"step": "go-settings", "state": "ready", "reason": "", "preview": {"files": [1, 2]},
                 "plan_digest": "d1", "apply": {"argv": ["moai", "go-settings-apply"]}},
                {"step": "go-state", "state": "blocked", "reason": "re-run after the go-settings apply",
                 "preview": null, "plan_digest": null, "apply": null}]},
            {"phase": "install", "steps": [
                {"step": "profile", "state": "blocked", "reason": "re-run after the backup phase",
                 "preview": null, "plan_digest": null, "apply": null},
                {"step": "github", "state": "skip", "reason": "no opt-in", "preview": null,
                 "plan_digest": null, "apply": null},
                {"step": "hooks", "state": "conflict", "reason": "edited", "preview": null,
                 "plan_digest": null, "apply": null}]},
        ]}));
        let items = project_items(&plan);
        let by = |id: &str| items.iter().find(|i| i.id == id).unwrap();
        assert_eq!(by("project-go-settings").group, WILL_CHANGE);
        assert_eq!(by("project-go-settings").entries, 2);
        assert_eq!(
            by("project-go-settings").apply.as_deref(),
            Some("moai go-settings-apply")
        );
        assert_eq!(by("project-go-state").state, "waiting");
        assert_eq!(by("project-go-state").group, WILL_CHANGE);
        assert!(by("project-go-state").guarded);
        assert_eq!(by("project-profile").state, "waiting");
        assert_eq!(by("project-github").group, KEPT);
        assert_eq!(by("project-hooks").group, NEEDS_PICK);
        assert_eq!(by("project-hooks").state, "conflict");
        assert_eq!(by("project-profile").depends_on, vec!["project-go-state"]);
    }

    #[test]
    fn the_home_move_waits_for_the_retire_and_is_guarded() {
        let waiting = Ok(json!({
            "state_dir": "folder", "action": "blocked",
            "blocks": ["go_items_not_retired:3"], "counts": {"move": 4},
            "content_digest": "sha256:aa",
        }));
        let m = home_move_item(&waiting, true).unwrap();
        assert_eq!((m.group, m.state.as_str()), (WILL_CHANGE, "waiting"));
        assert!(m.guarded);
        assert_eq!(m.depends_on, vec!["home-retire"]);
        assert!(m.apply.is_none());
        let stuck = Ok(json!({
            "state_dir": "folder", "action": "blocked",
            "blocks": ["go_items_not_retired:3", "different_file_system"], "counts": {"move": 0},
        }));
        let m = home_move_item(&stuck, true).unwrap();
        assert_eq!(m.group, NEEDS_PICK);
        assert!(m.reason.unwrap().contains("different_file_system"));
        let ready = Ok(json!({
            "state_dir": "folder", "action": "install", "blocks": [], "counts": {"move": 4},
        }));
        let m = home_move_item(&ready, false).unwrap();
        assert_eq!(
            (m.group, m.state.as_str(), m.entries),
            (WILL_CHANGE, "ready", 4)
        );
        assert!(m.apply.is_some());
        let none = Ok(json!({"state_dir": "absent", "action": "none", "blocks": []}));
        assert!(home_move_item(&none, false).is_none());
    }
}
