//! The model stage of `moai update` (SPEC-20261118-update-model-pick, card C-47 slice 3).
//!
//! On a person's terminal the stage asks the tool pick and the leader and agent model and effort
//! of each picked tool (`model_effort_questions`), shows what would change and writes the choice
//! to the user launcher settings file through the digest-bound apply. In every other run
//! [`report`] lists the model and effort each tool and role would use. Neither function quotes
//! saved text: a message names a key path or a fixed reason.

use crate::model_effort_questions::{self as questions, Change, Outcome, Plan};
use crate::model_effort_settings::{self as settings, Choice, ROLES, TOOLS};
use serde_json::{json, Value};
use std::fs;
use std::io::{BufRead, Write};
use std::path::{Path, PathBuf};

/// The user launcher settings file: `config/launcher.yaml` under the state folder of the home,
/// the file `session.rs` reads.
pub fn settings_path(home: &Path) -> PathBuf {
    crate::home_root::path(home).join("config/launcher.yaml")
}

/// The tools whose folder exists under the home: `claude` for `~/.claude`, `codex` for `~/.codex`.
/// A link or a file of that name is not a folder.
pub fn installed_tools(home: &Path) -> Vec<&'static str> {
    [("claude", ".claude"), ("codex", ".codex")]
        .into_iter()
        .filter(|(_, folder)| {
            fs::symlink_metadata(home.join(folder)).is_ok_and(|m| m.file_type().is_dir())
        })
        .map(|(tool, _)| tool)
        .collect()
}

fn error(message: &str) -> Value {
    json!({"state": "error", "message": message})
}

/// The model and effort each tool and role would use: the saved value when the support table
/// lists it, else the default, with the source, and the saved entries outside the table by key
/// path. An unusable settings file is the state `error`.
pub fn report(home: &Path) -> Value {
    let launcher = match questions::read_launcher(&settings_path(home)) {
        Ok(launcher) => launcher,
        Err(message) => return error(&message),
    };
    let unsupported = match settings::unsupported_entries(&launcher) {
        Ok(entries) => entries,
        Err(message) => return error(&message),
    };
    let mut values = Vec::new();
    for tool in TOOLS {
        for role in ROLES {
            // `saved_choice` answers only a supported entry.
            let saved = settings::saved_choice(&launcher, tool, role);
            let source = if saved.is_some() { "saved" } else { "default" };
            let Some(choice) = saved.or_else(|| settings::default_choice(tool, role)) else {
                continue;
            };
            values.push(json!({
                "tool": tool,
                "role": role,
                "model": choice.model,
                "effort": choice.effort,
                "source": source,
            }));
        }
    }
    json!({"state": "ok", "values": values, "unsupported": unsupported})
}

fn choice_json(choice: &Choice) -> Value {
    json!({"model": choice.model, "effort": choice.effort})
}

fn changes_json(changes: &[Change]) -> Vec<Value> {
    changes
        .iter()
        .map(|c| {
            json!({
                "tool": c.tool,
                "role": c.role,
                "from": c.from.as_ref().map(choice_json),
                "to": choice_json(&c.to),
            })
        })
        .collect()
}

fn unchanged(plan: &Plan) -> bool {
    plan.changes.iter().all(|c| c.from.as_ref() == Some(&c.to))
}

fn preview<W: Write>(out: &mut W, plan: &Plan) -> std::io::Result<()> {
    writeln!(out, "Model settings that would be saved:")?;
    for c in &plan.changes {
        let from = c.from.as_ref().map_or("nothing saved".to_string(), |f| {
            format!("{} {}", f.model, f.effort)
        });
        writeln!(
            out,
            "  {} {}: {from} -> {} {}",
            c.tool, c.role, c.to.model, c.to.effort
        )?;
    }
    for path in &plan.kept_unsupported {
        writeln!(out, "  saved but not supported, kept unchanged: {path}")?;
    }
    write!(out, "Save these model settings? [y/N] ")?;
    out.flush()
}

/// Asks the tool pick and the model questions on `input` and `out`, previews the change and
/// asks once whether to save it. The state is `nothing_picked`, `cancelled`, `unchanged`,
/// `declined`, `applied` (with `changes`) or `error` (with `message`). Nothing is written before
/// a yes, and a refusal of the apply leaves the file as it was.
pub fn stage<R: BufRead, W: Write>(home: &Path, input: &mut R, out: &mut W) -> Value {
    let path = settings_path(home);
    let picks = match questions::run(input, out, &installed_tools(home)) {
        Ok(Outcome::Picks(picks)) if picks.is_empty() => return json!({"state": "nothing_picked"}),
        Ok(Outcome::Picks(picks)) => picks,
        Ok(Outcome::Cancelled) => return json!({"state": "cancelled"}),
        Err(_) => return error("the answers could not be read"),
    };
    let plan = match questions::plan_file(&path, &picks) {
        Ok(plan) => plan,
        Err(message) => return error(&message),
    };
    if unchanged(&plan) {
        return json!({"state": "unchanged", "unsupported": plan.kept_unsupported});
    }
    if preview(out, &plan).is_err() {
        return error("the preview could not be written");
    }
    let answer = match questions::read_line_bounded(input) {
        Ok(Some(line)) => line,
        Ok(None) => String::new(),
        Err(_) => return error("the answer could not be read"),
    };
    let _ = writeln!(out);
    if !matches!(answer.trim().to_ascii_lowercase().as_str(), "y" | "yes") {
        return json!({"state": "declined", "changes": changes_json(&plan.changes), "unsupported": plan.kept_unsupported});
    }
    match questions::apply(&path, &picks, &plan.digest) {
        Ok(done) => json!({
            "state": "applied",
            "changes": changes_json(&done.changes),
            "unsupported": done.kept_unsupported,
        }),
        Err(message) => error(&message),
    }
}
