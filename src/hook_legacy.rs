//! The hook sets of the releases before #3004 P4-2 call `moai-cli`; the set generated now calls
//! `moai`. This module recognises the old entries, previews their rewrite (`binary_migration` in
//! the `hooks-plan` output) and holds the gate that decides whether a plan may be written.
//!
//! Recognition is exact equality with the legacy generator's own output
//! ([`hooks::hooks_object_with`] and friends called with [`hooks::LEGACY_BINARY`]), so the old
//! set cannot drift from the code that wrote it. An entry with an extra argument, a path
//! prefix, another timeout or matcher, or other whitespace is not MoAI's old entry: it is kept
//! as it is and never printed (the preview names its event, matcher and a reason only).
//!
//! Nothing here writes. A rewrite happens only through the writers that already own the files:
//! the Claude `hooks` key (three-way against its ownership receipt, `claude::plan_settings`).
//! A pristine old install is therefore an `update`; an edited one stays a conflict with its
//! bytes untouched. The preview is derived from the same `before` bytes the plan digest binds
//! and sits outside the digest, like `hooks` and `engine_capabilities`.

use crate::hooks::{self, HOOK_BINARY, LEGACY_BINARY};
use serde_json::{json, Map, Value};
use std::path::Path;

/// Longest matcher text a preview repeats; a longer one is cut (it is user-written text).
const MATCHER_SHOWN_MAX: usize = 128;

/// What a plan's `before` bytes were, for [`migration`].
pub struct Subject<'a> {
    pub host: &'a str,
    /// The host file, relative to the project.
    pub file: &'a str,
    /// The host file now (the bytes the plan digest binds).
    pub before: Option<&'a [u8]>,
    /// The plan's `action`: `create`, `update`, `unchanged` or `conflict`.
    pub action: &'a str,
}

/// One hook group of a `hooks` object: where it is and what it holds.
struct Group<'a> {
    event: &'a str,
    index: usize,
    value: &'a Value,
}

fn groups(hooks_value: &Value) -> Vec<Group<'_>> {
    let mut out = Vec::new();
    for (event, list) in hooks_value.as_object().into_iter().flatten() {
        for (index, value) in list.as_array().into_iter().flatten().enumerate() {
            out.push(Group {
                event,
                index,
                value,
            });
        }
    }
    out
}

fn matcher_of(group: &Value) -> Value {
    match group["matcher"].as_str() {
        Some(m) => json!(m.chars().take(MATCHER_SHOWN_MAX).collect::<String>()),
        None => Value::Null,
    }
}

/// The command of a group's first handler; shown only for a group equal to MoAI's own.
fn command_of(group: &Value) -> Value {
    group["hooks"][0]["command"].clone()
}

/// Whether any handler command of `group` names a MoAI binary (`moai-cli` anywhere in it, or
/// `moai` as its first word).
fn mentions_moai(group: &Value) -> bool {
    group["hooks"]
        .as_array()
        .into_iter()
        .flatten()
        .filter_map(|h| h["command"].as_str())
        .any(|c| c.contains(LEGACY_BINARY) || c.split_whitespace().next() == Some(HOOK_BINARY))
}

fn same_position(set: &Map<String, Value>, g: &Group) -> Option<Value> {
    set.get(g.event)?.as_array()?.get(g.index).cloned()
}

fn contains(set: &Map<String, Value>, g: &Group) -> bool {
    set.get(g.event)
        .and_then(Value::as_array)
        .is_some_and(|l| l.contains(g.value))
}

/// The `binary_migration` object of a plan. `none` unless the plan is an `update` (or a
/// `conflict`, which names what is held back) of a file that holds entries of the old set.
pub fn migration(sub: &Subject) -> Value {
    let mut out = json!({
        "from": LEGACY_BINARY,
        "to": HOOK_BINARY,
        "state": "none",
        "entries": [],
        "kept": [],
        "replaced": [],
        "blocked_by": null,
    });
    let rewrites = matches!(sub.action, "update" | "conflict");
    let Some(before) = sub.before.filter(|_| rewrites) else {
        return out;
    };
    hooks_file(sub, before, &mut out);
    out
}

fn hooks_file(sub: &Subject, before: &[u8], out: &mut Value) {
    let Ok(doc) = serde_json::from_slice::<Value>(before) else {
        return;
    };
    let (Ok(old), Ok(new)) = (
        hooks::legacy_hooks(sub.host),
        hooks::current_hooks(sub.host),
    ) else {
        return;
    };
    let (mut entries, mut kept, mut replaced) = (Vec::new(), Vec::new(), Vec::new());
    let updating = sub.action == "update";
    for g in groups(&doc["hooks"]) {
        let matcher = matcher_of(g.value);
        if contains(&old, &g) {
            let after = same_position(&new, &g).map_or(Value::Null, |n| command_of(&n));
            entries.push(json!({
                "file": sub.file, "event": g.event, "matcher": matcher,
                "before": command_of(g.value), "after": after,
                "decision": if updating { "rewrite" } else { "held" },
            }));
            continue;
        }
        let reason = if contains(&new, &g) {
            "current"
        } else if updating {
            // The whole key is MoAI's (it equals its ownership receipt) and is replaced.
            replaced.push(
                json!({"file": sub.file, "event": g.event, "matcher": matcher,
                "reason": "earlier_moai_set"}),
            );
            continue;
        } else if mentions_moai(g.value) {
            "edited"
        } else {
            "not_moai"
        };
        kept.push(
            json!({"file": sub.file, "event": g.event, "matcher": matcher,
            "reason": reason}),
        );
    }
    let foreign = kept.iter().any(|k| k["reason"] != json!("current"));
    out["kept"] = json!(kept);
    out["replaced"] = json!(replaced);
    if updating {
        if !entries.is_empty() || !replaced.is_empty() {
            out["state"] = json!("rewrite");
        }
    } else if !entries.is_empty() {
        // A conflict: nothing is written, not even the entries that are MoAI's own.
        out["state"] = json!("held");
        out["blocked_by"] = json!(if foreign {
            "mixed_entries"
        } else {
            "not_owned"
        });
    }
    out["entries"] = json!(entries);
}

/// How many entries a plan's `binary_migration` rewrites.
pub fn rewrite_count(plan: &Value) -> usize {
    let m = &plan["binary_migration"];
    if m["state"] != json!("rewrite") {
        return 0;
    }
    ["entries", "replaced"]
        .iter()
        .map(|k| m[*k].as_array().map_or(0, Vec::len))
        .sum()
}

/// Puts the gate's verdict into a plan's `binary_migration.blocked_by` and a top-level `gate`
/// object, from the `hook_binary` state the caller probed. Reports only.
pub fn annotate(plan: &mut Value, binary_state: &str) {
    let verdict = check(plan, binary_state);
    plan["gate"] = match &verdict {
        Ok(()) => json!({"allowed": true, "reason": null}),
        Err(reason) => json!({"allowed": false, "reason": reason}),
    };
    if verdict.is_err() && plan["binary_migration"]["state"] == json!("rewrite") {
        plan["binary_migration"]["blocked_by"] = json!(format!("moai_{binary_state}"));
    }
}

/// The gate decision for a plan, given the state of the `moai` first on PATH.
///
/// - A plan that rewrites old `moai-cli` entries needs `resolved`: the hooks would call `moai`,
///   and a `moai` that is not the Rust build (or is not there) leaves the claim guard off while
///   everything reports success.
/// - A fresh install (a `create`, or an `update` that rewrites nothing) is refused on
///   `wrong_generation` only. With no `moai` on PATH (`binary_unresolved`, or a `not_probed`
///   hit inside the project) it is written and reported, as `hooks-apply` always did.
/// - A plan that writes nothing (`unchanged`, `conflict`) is never gated.
pub fn check(plan: &Value, binary_state: &str) -> Result<(), String> {
    if !matches!(
        plan["action"].as_str(),
        Some("create" | "update" | "replace")
    ) {
        return Ok(());
    }
    if plan["binary_migration"]["state"] == json!("rewrite") {
        if binary_state == "resolved" {
            return Ok(());
        }
        return Err(format!(
            "hooks not rewritten: the hooks would call moai, and the moai first on PATH is not \
             the Rust build ({binary_state}); the claim guard would stay off. Put the Rust moai \
             first on PATH and plan again"
        ));
    }
    if binary_state == "wrong_generation" {
        return Err(
            "hooks not written: the moai first on PATH is not the Rust build \
             (wrong_generation), so the generated hooks would reach it and the claim guard would \
             stay off. Put the Rust moai first on PATH and plan again"
                .to_string(),
        );
    }
    Ok(())
}

/// The gate of an install of the MoAI set into a project that has none of it: a fresh create,
/// as [`check`] sees it. The Go hooks replacement (`go-hooks-apply`) installs the set after it
/// took the Go entries out, so its own `hooks` plan is not a create yet when it is gated; this
/// answers for the install it is about to make. The `moai` is probed once.
pub fn fresh_install(
    project: &Path,
    home: Option<&Path>,
    path_env: &str,
) -> std::result::Result<(), String> {
    let state = hooks::moai_state(project, path_env, home);
    check(
        &json!({"action": "create"}),
        state["state"].as_str().unwrap_or("wrong_generation"),
    )
}

/// The single gate: plans `host` in `project` again, now, and refuses what [`check`] refuses
/// for the `moai` on `path_env`. `hooks-apply` and the `moai update` hook layer both call it, at
/// plan time and again at apply time, because the plan digest binds the files and not the PATH.
/// The `moai` is probed only when the plan writes something.
pub fn gate(project: &Path, host: &str, home: Option<&Path>, path_env: &str) -> crate::Result<()> {
    let plan = hooks::plan(project, host, home)?;
    gate_plan(&plan, project, home, path_env)
}

/// [`gate`] for a plan already made.
pub fn gate_plan(
    plan: &Value,
    project: &Path,
    home: Option<&Path>,
    path_env: &str,
) -> crate::Result<()> {
    if !matches!(
        plan["action"].as_str(),
        Some("create" | "update" | "replace")
    ) {
        return Ok(());
    }
    let state = hooks::moai_state(project, path_env, home);
    check(plan, state["state"].as_str().unwrap_or("wrong_generation"))
}
