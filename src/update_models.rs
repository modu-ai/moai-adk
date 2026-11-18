//! The model stage of `moai update` (SPEC-20261118-update-model-pick, card C-47 slice 3).
//!
//! On a person's terminal the stage asks the tool pick and the leader and agent model and effort
//! of each picked tool (`model_effort_questions`), shows what would change and writes the choice
//! to the user launcher settings file through the digest-bound apply. In every other run
//! [`report`] lists the model and effort each tool and role would use. Neither function quotes
//! saved text: a message names a key path or a fixed reason.

use crate::key_vault;
use crate::model_effort_questions::{self as questions, Change, Outcome, Plan};
use crate::model_effort_settings::{self as settings, Choice, ROLES, TOOLS};
use serde_json::{json, Value};
use sha2::{Digest, Sha256};
use std::fs;
use std::io::{BufRead, Write};
use std::path::{Path, PathBuf};

/// The permission reference of the receipt: the person answered yes to the save question.
const PERMISSION: &str = "user-pick:update-models";
/// The prefix of the receipt ids of this write.
const RECEIPT_PREFIX: &str = "models-";
/// A receipt larger than this is not written (the reader of receipts stops at 65536 bytes).
const RECEIPT_MAX_BYTES: usize = 60_000;

fn sha(text: &str) -> String {
    format!("{:x}", Sha256::digest(text.as_bytes()))
}

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
    // The saved choices are the defaults of the questions; an unusable file ends the stage here.
    let launcher = match questions::read_launcher(&path) {
        Ok(launcher) => launcher,
        Err(message) => return error(&message),
    };
    let picks = match questions::run_with(input, out, &installed_tools(home), &launcher) {
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
    save(home, &path, &picks, &plan)
}

/// The save after a yes: the receipt first (the text before the write, the hashes before and
/// after), then the digest-bound apply, then a read back against the promised hash. No receipt
/// means no write.
fn save(home: &Path, path: &Path, picks: &[questions::Pick], plan: &Plan) -> Value {
    let before = match questions::read_text(path) {
        Ok(text) => text,
        Err(message) => return error(&message),
    };
    let after = match questions::after_text(path, plan) {
        Ok(text) => text,
        Err(message) => return error(&message),
    };
    // A plan that leaves the launcher object as it is writes nothing (the apply keeps the file).
    if questions::read_launcher(path).is_ok_and(|launcher| launcher == plan.after) {
        return json!({"state": "unchanged", "unsupported": plan.kept_unsupported});
    }
    let unix_ms = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map(|d| d.as_millis() as u64)
        .unwrap_or(0);
    let id = format!("{RECEIPT_PREFIX}{unix_ms}-{}", std::process::id());
    let receipt = json!({
        "schema_version": 1,
        "operation": "update.models",
        "path": path.display().to_string(),
        "before_text": before,
        "before_sha256": before.as_deref().map(sha),
        "after_sha256": sha(&after),
        "permission_reference": PERMISSION,
        "unix_ms": unix_ms,
    });
    if receipt.to_string().len() > RECEIPT_MAX_BYTES {
        return error("the launcher settings are too large to keep a receipt (60000 bytes); nothing was changed");
    }
    if key_vault::write_home_receipt(home, &id, &receipt).is_err() {
        return error("the receipt could not be written; nothing was changed");
    }
    let done = match questions::apply(path, picks, &plan.digest) {
        Ok(done) => done,
        Err(message) => return error(&message),
    };
    match questions::read_text(path) {
        Ok(Some(text)) if sha(&text) == sha(&after) => json!({
            "state": "applied",
            "receipt": id,
            "changes": changes_json(&done.changes),
            "unsupported": done.kept_unsupported,
        }),
        _ => json!({
            "state": "error",
            "receipt": id,
            "message": "the settings file differs from what the receipt promised; check the file",
        }),
    }
}

fn receipt_id_ok(id: &str) -> bool {
    id.strip_prefix(RECEIPT_PREFIX).is_some_and(|rest| {
        !rest.is_empty()
            && rest.len() <= 64
            && rest.bytes().all(|b| b.is_ascii_digit() || b == b'-')
    })
}

/// Puts the settings file back as the receipt kept it, while the file is still the file the save
/// wrote (its hash equals the receipt's): the text before the write byte for byte, or no file when
/// there was none. A file changed since is left alone (`skipped`, no marker). Nothing is changed
/// when the reference, the id, the receipt or the settings path is not the one that wrote.
pub fn revert(home: &Path, receipt_id: &str, permission: &str) -> Result<Value, String> {
    if !crate::contract::permission_reference_ok(permission) {
        return Err("permission reference required".into());
    }
    if !receipt_id_ok(receipt_id) {
        return Err("unknown models receipt id".into());
    }
    let marker = format!("{receipt_id}.reverted");
    if key_vault::read_home_receipt(home, &marker)?.is_some() {
        return Err("models receipt already reverted".into());
    }
    let receipt =
        key_vault::read_home_receipt(home, receipt_id)?.ok_or("models receipt not found")?;
    if receipt["operation"] != json!("update.models") {
        return Err("not a models receipt".into());
    }
    let path = settings_path(home);
    if receipt["path"] != json!(path.display().to_string()) {
        return Err("models receipt names another settings file".into());
    }
    let after_sha = receipt["after_sha256"]
        .as_str()
        .ok_or("models receipt is damaged")?;
    let before = match &receipt["before_text"] {
        Value::String(text) => Some(text.as_str()),
        Value::Null => None,
        _ => return Err("models receipt is damaged".into()),
    };
    if receipt["before_sha256"].as_str() != before.map(sha).as_deref() {
        return Err("models receipt is damaged".into());
    }
    let current = questions::read_text(&path)?;
    if current.as_deref().map(sha).as_deref() != Some(after_sha) {
        return Ok(json!({
            "state": "skipped",
            "reason": "changed_since_apply",
            "receipt": receipt_id,
            "path": path.display().to_string(),
        }));
    }
    match before {
        Some(text) => questions::write_text(&path, text)?,
        None => questions::remove_settings(&path)?,
    }
    if questions::read_text(&path)?.as_deref() != before {
        return Err(
            "the settings file read back differently after the revert; check the file".into(),
        );
    }
    key_vault::write_home_receipt(
        home,
        &marker,
        &json!({
            "schema_version": 1,
            "operation": "update.models_revert",
            "receipt": receipt_id,
            "permission_reference": permission,
        }),
    )?;
    Ok(json!({
        "state": "reverted",
        "receipt": receipt_id,
        "path": path.display().to_string(),
        "file_removed": before.is_none(),
    }))
}
