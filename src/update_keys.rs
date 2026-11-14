//! The keys layer of `moai update` (#2871 I2 S3): a legacy `.env.<provider>` key file under HOME
//! — the Go generation's `~/.moai/` or the earlier `~/.moai-cli/` — is previewed beside the rest
//! of the user layer and, once confirmed, copied into the vault's `_shared` folder.
//!
//! Leader decisions (PLAN 6876533944 Q5–Q7, survey 6876590509 R8): the key is copied and the
//! original file is kept exactly as it was; only `glm` is imported (what the launcher reads,
//! `session.rs`), every other `.env.<provider>` is reported as `keep`; the preview reads a source
//! only as `key-import-plan` does, after the safety rules below, and never shows the value.
//!
//! A source is kept unread when it is a symlink, a FIFO or any other non-regular file, a folder,
//! larger than 64 KiB, has a mode looser than `0600` (or another owner than HOME's), or names a
//! provider other than `glm`. The layer's digest is part of the one `plan_digest` `--confirm`
//! takes; an entry already in the vault is never replaced. A confirmed apply records a receipt
//! (paths, provider, the sealed entry's digest — never the value), and `keys-revert` removes an
//! entry only while it still holds what the apply wrote; the original is never touched.

use crate::key_vault::{self, Scope};
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};

type Result<T> = std::result::Result<T, String>;

/// The only provider imported (PLAN Q6).
const PROVIDER: &str = "glm";
/// The receipt name prefix under the HOME receipts folder.
const RECEIPT_PREFIX: &str = "keys-";

/// One source file of the keys layer.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Row {
    pub provider: String,
    pub source: PathBuf,
    /// `import`, `already_in_vault` or `keep`.
    pub action: &'static str,
    /// Why a `keep` row is kept.
    pub reason: Option<String>,
    /// The metadata of an `import` source (bound into the digest), else null.
    identity: Value,
}

/// A key read for an import. Never printed: its `Debug` shows nothing.
struct Secret(String);

impl std::fmt::Debug for Secret {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.write_str("Secret(..)")
    }
}

/// The keys layer of one `moai update` preview.
#[derive(Debug)]
pub struct Layer {
    pub rows: Vec<Row>,
    /// The vault entry an import writes.
    pub destination: PathBuf,
    /// The destination's state: `absent`, `present`, `empty` or `unsafe`.
    pub vault_entry: &'static str,
    /// Binds every row, the import source's metadata and the destination's state.
    pub plan_digest: String,
    /// The key of the `import` row, when there is one.
    token: Option<Secret>,
}

impl Layer {
    /// True when confirming would import a key.
    pub fn writes(&self) -> bool {
        self.token.is_some() && self.rows.iter().any(|r| r.action == "import")
    }

    pub fn to_json(&self) -> Value {
        json!({
            "rows": self.rows.iter().map(|r| json!({
                "provider": r.provider,
                "source": r.source.display().to_string(),
                "action": r.action,
                "reason": r.reason,
            })).collect::<Vec<_>>(),
            "destination": self.destination.display().to_string(),
            "vault_entry": self.vault_entry,
            "original_kept": true,
            "plan_digest": self.plan_digest,
            "value_shown": false,
        })
    }

    /// The terminal question's lines for this layer; empty when it has no rows.
    pub fn summary(&self) -> String {
        if self.rows.is_empty() {
            return String::new();
        }
        let mut out = String::from("keys\n");
        for r in &self.rows {
            let action = match &r.reason {
                Some(reason) => format!("{}: {reason}", r.action),
                None => r.action.to_string(),
            };
            out.push_str(&format!(
                "  {} {}: {action}\n",
                r.provider,
                r.source.display()
            ));
        }
        out
    }
}

/// The `.env.<provider>` files of one HOME state folder, `.env.glm` left out (it is planned
/// first, in the launcher's order). A folder that is a symlink or not a folder is not listed.
fn other_sources(home: &Path, dir: &str) -> Vec<(String, PathBuf)> {
    let root = home.join(dir);
    if !fs::symlink_metadata(&root).is_ok_and(|m| m.is_dir()) {
        return Vec::new();
    }
    let mut out: Vec<(String, PathBuf)> = fs::read_dir(&root)
        .into_iter()
        .flatten()
        .flatten()
        .filter_map(|e| {
            let name = e.file_name().to_str()?.to_string();
            let provider = name.strip_prefix(".env.")?.to_string();
            (provider != PROVIDER).then(|| (provider, e.path()))
        })
        .collect();
    out.sort();
    out
}

/// Plans the keys layer. Reads a glm source only after the safety rules; never writes.
pub fn plan(home: &Path) -> Result<Layer> {
    let scope = Scope::Shared;
    let destination = key_vault::entry_path(home, &scope, PROVIDER);
    let vault_entry = match key_vault::read_entry(home, &scope, PROVIDER) {
        Ok(Some(_)) => "present",
        Ok(None) => "absent",
        Err(b) if b.code == "key_absent" => "empty",
        Err(_) => "unsafe",
    };
    let mut rows = Vec::new();
    let mut token = None;
    for rel in key_vault::LEGACY_GLM_FILES {
        let source = home.join(rel);
        // The folder holding it must be a real folder: a linked `~/.moai` is not followed.
        let parent_ok = source
            .parent()
            .is_some_and(|p| fs::symlink_metadata(p).is_ok_and(|m| m.is_dir()));
        if !parent_ok || fs::symlink_metadata(&source).is_err() {
            continue;
        }
        let mut row = Row {
            provider: PROVIDER.into(),
            source,
            action: "keep",
            reason: None,
            identity: Value::Null,
        };
        match vault_entry {
            "present" => row.action = "already_in_vault",
            "absent" => match key_vault::layer_source(home, &row.source) {
                Err(why) => row.reason = Some(why.to_string()),
                Ok((None, _)) => row.reason = Some("no usable key in the file".into()),
                Ok((Some(_), _)) if token.is_some() => {
                    row.reason = Some("another glm file is imported".into())
                }
                Ok((Some(t), identity)) => {
                    row.action = "import";
                    row.identity = identity;
                    token = Some(Secret(t));
                }
            },
            _ => {
                row.reason = Some(
                    "the vault's glm entry is unusable; check it with key-status glm --shared"
                        .into(),
                )
            }
        }
        rows.push(row);
    }
    for dir in [crate::state_root::MOAI_DIR, crate::HOME_STATE_DIR] {
        for (provider, source) in other_sources(home, dir) {
            rows.push(Row {
                provider,
                source,
                action: "keep",
                reason: Some("unknown provider: only glm is imported".into()),
                identity: Value::Null,
            });
        }
    }
    let digest_input = json!({
        "schema_version": 1,
        "rows": rows.iter().map(|r| json!([r.provider, r.source.display().to_string(), r.action, r.reason, r.identity])).collect::<Vec<_>>(),
        "destination": destination.display().to_string(),
        "vault_entry": vault_entry,
    });
    let plan_digest = crate::digest(digest_input.to_string().as_bytes());
    Ok(Layer {
        rows,
        destination,
        vault_entry,
        plan_digest,
        token,
    })
}

/// Applies a confirmed keys layer (the caller compared the combined digest of a fresh plan):
/// writes the `import` row's key into the vault without replacing an entry, reads it back, and
/// records a receipt. The source file is left as it was.
pub fn apply(home: &Path, layer: &Layer, permission: &str) -> Result<Value> {
    let (Some(row), Some(token)) = (
        layer.rows.iter().find(|r| r.action == "import"),
        layer.token.as_ref(),
    ) else {
        return Ok(json!({"state": "unchanged", "written": [], "original_kept": true}));
    };
    let scope = Scope::Shared;
    key_vault::set_new(home, &scope, &row.provider, &token.0)?;
    let matched =
        matches!(key_vault::read_entry(home, &scope, &row.provider), Ok(Some(v)) if v == token.0);
    let entry_sha256 = key_vault::entry_digest(home, &scope, &row.provider)?;
    let unix_ms = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map(|d| d.as_millis() as u64)
        .unwrap_or(0);
    let id = format!("{RECEIPT_PREFIX}{unix_ms}-{}", std::process::id());
    let destination = layer.destination.display().to_string();
    let receipt = json!({
        "schema_version": 1,
        "operation": "update.keys",
        "status": "applied",
        "imported": [{
            "provider": row.provider,
            "folder": scope.folder(),
            "source": row.source.display().to_string(),
            "destination": destination,
            "entry_sha256": entry_sha256,
            "matched": matched,
        }],
        "original_kept": true,
        "permission_reference": permission,
        "unix_ms": unix_ms,
    });
    key_vault::write_home_receipt(home, &id, &receipt)?;
    if !matched {
        return Err(format!(
            "the vault read-back did not match; receipt {id} (revert with keys-revert)"
        ));
    }
    Ok(json!({
        "state": "applied",
        "written": [destination],
        "imported": [{"provider": row.provider, "source": row.source.display().to_string()}],
        "receipt": id,
        "original_kept": true,
        "value_shown": false,
    }))
}

fn receipt_id_ok(id: &str) -> bool {
    id.strip_prefix(RECEIPT_PREFIX).is_some_and(|rest| {
        !rest.is_empty()
            && rest.len() <= 64
            && rest.bytes().all(|b| b.is_ascii_digit() || b == b'-')
    })
}

/// `keys-revert`: removes each entry a confirmed keys layer imported, only while it still holds
/// what the apply wrote (an entry replaced since — a later `key-set` — is skipped and named).
/// The legacy source file is never touched. A receipt reverts once.
pub fn revert(home: &Path, receipt_id: &str, permission: &str) -> Result<Value> {
    if !crate::contract::permission_reference_ok(permission) {
        return Err("permission reference required".into());
    }
    if !receipt_id_ok(receipt_id) {
        return Err("unknown keys receipt id".into());
    }
    let done = format!("{receipt_id}.reverted");
    if key_vault::read_home_receipt(home, &done)?.is_some() {
        return Err("keys receipt already reverted".into());
    }
    let receipt =
        key_vault::read_home_receipt(home, receipt_id)?.ok_or("keys receipt not found")?;
    if receipt["operation"] != json!("update.keys") {
        return Err("not a keys receipt".into());
    }
    let mut reverted = Vec::new();
    let mut skipped = Vec::new();
    for e in receipt["imported"].as_array().into_iter().flatten() {
        let provider = e["provider"].as_str().unwrap_or("");
        key_vault::check_provider(provider)?;
        if e["folder"] != json!(Scope::Shared.folder()) {
            return Err("keys receipt names a folder other than the shared one".into());
        }
        let scope = Scope::Shared;
        let path = key_vault::entry_path(home, &scope, provider)
            .display()
            .to_string();
        let sha = e["entry_sha256"].as_str().unwrap_or("");
        if key_vault::remove_if_digest(home, &scope, provider, sha)? {
            reverted.push(path);
        } else {
            skipped.push(json!({"path": path, "reason": "changed_since_apply"}));
        }
    }
    key_vault::write_home_receipt(
        home,
        &done,
        &json!({
            "schema_version": 1,
            "operation": "update.keys_revert",
            "receipt": receipt_id,
            "reverted": reverted,
            "skipped": skipped,
            "permission_reference": permission,
        }),
    )?;
    Ok(json!({
        "state": "reverted",
        "receipt": receipt_id,
        "reverted": reverted,
        "skipped": skipped,
        "original_kept": true,
    }))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn receipt_ids_are_the_layers_own() {
        assert!(receipt_id_ok("keys-1731580000000-42"));
        for bad in [
            "keys-",
            "keys-1/2",
            "hooks-1",
            "keys-1.reverted",
            "../keys-1",
        ] {
            assert!(!receipt_id_ok(bad), "{bad}");
        }
    }

    #[test]
    fn a_secret_never_debug_prints() {
        let s = Secret("DUMMY-GLM-KEY-123".into());
        assert!(!format!("{s:?}").contains("DUMMY"));
    }
}
