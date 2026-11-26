//! `moai card guard-repoint` (SPEC-20261126-guard-record-repoint): re-point a stale guard record.
//!
//! The entry point is wired first so the command exists; the body is filled by the next commit.

use serde_json::Value;
use std::path::Path;

type Result<T> = std::result::Result<T, String>;

/// The options of `card guard-repoint`.
pub struct RepointRequest<'a> {
    /// `--apply`: rewrite the record.
    pub apply: bool,
    /// `--revert <stamp>`: restore the bytes a receipt kept.
    pub revert: Option<&'a str>,
    /// `--plan-digest`: the digest of the preview.
    pub plan_digest: Option<&'a str>,
    /// `--permission-reference`: the reference of the user's yes.
    pub permission_reference: Option<&'a str>,
}

/// `card guard-repoint`: see the module comment.
pub fn repoint(_project: &Path, _req: &RepointRequest<'_>) -> Result<Value> {
    Err("card guard-repoint refused: not implemented yet".to_string())
}
