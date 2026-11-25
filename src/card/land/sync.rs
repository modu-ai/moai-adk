//! `moai card land-sync` (SPEC-20261125-land-sync-merge): the leader merges another branch into
//! the integration branch as one merge commit and moves the branch by one compare-and-swap.
//!
//! This file is the scaffold the command line calls; the body follows in the next commit.

use serde_json::Value;
use std::path::Path;

/// What `card land-sync` was given.
pub struct SyncRequest<'a> {
    /// This session's id, when it has one.
    pub session: Option<&'a str>,
    /// `--by`: who syncs; only the leader word is accepted.
    pub by: &'a str,
    /// The local branch to merge in.
    pub from: &'a str,
    /// `--spec`: the approved SPEC that decided the sync.
    pub spec: &'a str,
    /// The tests' seam for the check program (a debug build reads it from the environment).
    pub check_cmd: Option<&'a Path>,
}

/// `card land-sync`: see the module comment.
pub fn land_sync(_project: &Path, _home: &Path, _req: &SyncRequest<'_>) -> Result<Value, String> {
    Err(
        "card land-sync refused: the command is not implemented yet; nothing was changed"
            .to_string(),
    )
}
