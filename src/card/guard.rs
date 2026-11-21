//! The guard of the integration branch (SPEC-20261121-ref-guard). Red-test skeleton: the shape of
//! the API and no behaviour yet.

use std::path::Path;

/// The record of the protected branch, in the common git folder.
pub const RECORD_FILE: &str = "moai/integration-branch";
/// The intent record of the lander, in the common git folder.
pub const INTENT_FILE: &str = "moai/land-intent";

/// The protected branch as the one reader finds it.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Protected {
    /// The full reference, `refs/heads/<name>`.
    pub full_ref: String,
    /// The profile's branch name when it disagrees with the record.
    pub profile_disagrees: Option<String>,
    /// Whether this call wrote the record.
    pub written: bool,
}

/// The hook text (a POSIX shell script) for the managed-hook installer.
pub fn hook_text() -> &'static str {
    "#!/bin/sh\nexit 0\n"
}

/// The protected branch of the repository of `project`: the record first, the profile's value
/// only for a repository with no record (written as the record when `write_if_absent`).
pub fn protected(project: &Path, write_if_absent: bool) -> Result<Option<Protected>, String> {
    let _ = (project, write_if_absent);
    Ok(None)
}

/// What `card guard <state>` does with the transaction lines git hands the hook.
pub fn run(state: &str, input: &str) -> Result<(), String> {
    let _ = (state, input);
    Ok(())
}
