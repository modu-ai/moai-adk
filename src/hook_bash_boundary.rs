//! SPEC-20261123-hook-bash-boundary: the judgment of `moai hook bash-boundary` (H2).
//! RED stub: it decides nothing yet.

use serde_json::{Map, Value};
use std::path::Path;

/// What the handler needs besides the hook input.
pub struct Env<'a> {
    /// The folder the entry found `.moai/project.yaml` in.
    pub project_root: &'a Path,
    /// `CLAUDE_PROJECT_DIR`, when set.
    pub project_dir: Option<&'a Path>,
    pub home: Option<&'a Path>,
    /// The working folder of the hook process, used when the input has no `cwd`.
    pub process_cwd: &'a Path,
}

pub fn decide(_input: &Map<String, Value>, _env: &Env) -> Option<Value> {
    None
}

/// Whether the command holds a word the checks react to (D-11).
pub fn needs_walk(_command: &str) -> bool {
    false
}
