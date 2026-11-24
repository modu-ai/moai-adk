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

/// The tools whose folder exists under the home: `claude` for `~/.claude`. A link or a file of
/// that name is not a folder.
pub fn installed_tools(home: &Path) -> Vec<&'static str> {
    [("claude", ".claude")]
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
        // The receipt is already kept: name it, because the file may or may not have changed.
        Err(message) => return json!({"state": "error", "receipt": id, "message": message}),
    };
    if file_has_hash(path, &sha(&after)) {
        json!({
            "state": "applied",
            "receipt": id,
            "changes": changes_json(&done.changes),
            "unsupported": done.kept_unsupported,
        })
    } else {
        json!({
            "state": "error",
            "receipt": id,
            "message": "the settings file differs from what the receipt promised; check the file",
        })
    }
}

fn receipt_id_ok(id: &str) -> bool {
    id.strip_prefix(RECEIPT_PREFIX).is_some_and(|rest| {
        !rest.is_empty()
            && rest.len() <= 64
            && rest.bytes().all(|b| b.is_ascii_digit() || b == b'-')
    })
}

/// The text before the save that a receipt kept: `Some(Some(text))` for a file that existed,
/// `Some(None)` for a file that was absent, `None` when the receipt is damaged. Both keys must be
/// present: a text with its hash, or two nulls — a key that is missing never reads as "absent".
fn before_of(receipt: &Value) -> Option<Option<&str>> {
    let object = receipt.as_object()?;
    let (text, hash) = (object.get("before_text")?, object.get("before_sha256")?);
    match (text, hash) {
        (Value::Null, Value::Null) => Some(None),
        (Value::String(text), Value::String(hash)) if sha(text) == *hash => Some(Some(text)),
        _ => None,
    }
}

/// Whether the file at `path` is exactly `expected` (`None` = no file).
fn file_is(path: &Path, expected: Option<&str>) -> bool {
    questions::read_text(path).is_ok_and(|now| now.as_deref() == expected)
}

/// Whether the file at `path` hashes to `expected_sha`.
fn file_has_hash(path: &Path, expected_sha: &str) -> bool {
    questions::read_text(path)
        .ok()
        .flatten()
        .is_some_and(|text| sha(&text) == expected_sha)
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
        return Err("invalid models receipt id".into());
    }
    let marker = format!("{receipt_id}.reverted");
    // An unreadable marker reads as "not written": the honest failure then comes from the
    // marker write itself, which names the path and says nothing was changed.
    let marker_written = matches!(key_vault::read_home_receipt(home, &marker), Ok(Some(_)));
    let receipt =
        key_vault::read_home_receipt(home, receipt_id)?.ok_or("models receipt not found")?;
    if receipt["operation"] != json!("update.models") {
        return Err("conflict in models revert: the receipt is not a models receipt".into());
    }
    let path = settings_path(home);
    if receipt["path"] != json!(path.display().to_string()) {
        return Err("conflict in models revert: the receipt names another settings file".into());
    }
    let after_sha = receipt["after_sha256"]
        .as_str()
        .ok_or("invalid models receipt: the fields the revert needs are missing or do not hash")?;
    let before = before_of(&receipt)
        .ok_or("invalid models receipt: the fields the revert needs are missing or do not hash")?;
    let current = questions::read_text(&path)?;
    if current.as_deref().map(sha).as_deref() != Some(after_sha) {
        // A marker without the reverted file is a dead run's window: resume below. A file
        // that no longer matches is either already reverted or changed since — refused.
        if marker_written {
            return Err("models receipt already reverted".into());
        }
        return Ok(json!({
            "state": "skipped",
            "reason": "changed_since_apply",
            "receipt": receipt_id,
            "path": path.display().to_string(),
        }));
    }
    // The marker goes first (SPEC-20261119-models-revert-reporting REQ-1): a failed marker
    // write leaves the settings untouched and retryable, and a run that dies after the
    // marker resumes here through the window check above.
    if !marker_written {
        key_vault::write_home_receipt(
            home,
            &marker,
            &json!({
                "schema_version": 1,
                "operation": "update.models_revert",
                "receipt": receipt_id,
                "permission_reference": permission,
            }),
        )
        .map_err(|e| format!("models revert marker write failed: {e}; nothing was changed"))?;
    }
    let restored = (|| -> std::result::Result<(), String> {
        match before {
            Some(text) => questions::write_text(&path, text)?,
            None => questions::remove_settings(&path)?,
        }
        if !file_is(&path, before) {
            return Err(
                "conflict in models revert: the settings file read back differently after the \
                 revert; check the file"
                    .into(),
            );
        }
        Ok(())
    })();
    restored.inspect_err(|_| {
        // The settings are not back, so the marker must not say they were: best effort
        // removal. A marker that survives this cleanup is a dead run's window, and the
        // marker+hash check above resumes it on the next attempt.
        let _ = std::fs::remove_file(
            crate::home_root::path(home)
                .join("receipts")
                .join(format!("{marker}.json")),
        );
    })?;
    Ok(json!({
        "state": "reverted",
        "receipt": receipt_id,
        "path": path.display().to_string(),
        "file_removed": before.is_none(),
    }))
}

#[cfg(test)]
mod tests {
    use super::*;

    fn temp_file(text: Option<&str>) -> (tempfile::TempDir, std::path::PathBuf) {
        let dir = tempfile::TempDir::new().unwrap();
        let path = dir.path().join("launcher.yaml");
        if let Some(text) = text {
            std::fs::write(&path, text).unwrap();
        }
        (dir, path)
    }

    #[test]
    fn the_read_back_after_a_revert_compares_the_whole_file() {
        let (_dir, path) = temp_file(Some("one"));
        assert!(file_is(&path, Some("one")));
        assert!(!file_is(&path, Some("two")), "a different text");
        assert!(!file_is(&path, None), "a file where none was expected");
        let (_dir, absent) = temp_file(None);
        assert!(file_is(&absent, None));
        assert!(
            !file_is(&absent, Some("one")),
            "no file where one was expected"
        );
    }

    #[test]
    fn the_promise_check_after_a_save_compares_the_hash_of_the_file() {
        let (_dir, path) = temp_file(Some("promised"));
        assert!(file_has_hash(&path, &sha("promised")));
        assert!(!file_has_hash(&path, &sha("something else")));
        let (_dir, absent) = temp_file(None);
        assert!(
            !file_has_hash(&absent, &sha("promised")),
            "no file keeps no promise"
        );
    }

    #[test]
    fn a_receipt_reads_back_only_with_both_before_fields() {
        let text = "earlier";
        let good = json!({"before_text": text, "before_sha256": sha(text)});
        assert_eq!(before_of(&good), Some(Some(text)));
        let absent = json!({"before_text": null, "before_sha256": null});
        assert_eq!(before_of(&absent), Some(None));
        for bad in [
            json!({}),
            json!({"before_text": text}),
            json!({"before_sha256": sha(text)}),
            json!({"before_text": null}),
            json!({"before_text": text, "before_sha256": null}),
            json!({"before_text": null, "before_sha256": sha(text)}),
            json!({"before_sha256": null}),
            json!({"before_text": text, "before_sha256": sha("other")}),
            json!({"before_text": 1, "before_sha256": 1}),
            json!("not an object"),
        ] {
            assert_eq!(before_of(&bad), None, "{bad}");
        }
    }
}
