//! `project-detect` (#2496): detects a project's facts for `/moai project`. Moved out of `src/main.rs` unchanged (#2834 R4a).

use std::path::PathBuf;

/// Detect the project's facts
///
/// Detect the project's facts for the `/moai project` route (#2496): new or existing,
/// manifests, languages, language servers, configured MCP server names, databases. Reads
/// only and writes nothing; returns the profile with `detected` set for `project-plan`.
#[derive(clap::Args)]
pub struct Args {
    project: PathBuf,
}

pub const OPERATION: &str = "project.detect";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::project_detect::detect_cli(&args.project)
}
