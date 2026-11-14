//! `goal-show` (#2679): the standing goal of this clone, or `state: none`.

use std::path::PathBuf;

/// Show the standing goal of this clone
///
/// Show the standing goal of this clone (#2679): `state` is `active`, `cleared` or `none`,
/// with the stored text, end condition, exclusions, source, start time and revision. The leader
/// and every lane read it at the start of a round. Reads only.
#[derive(clap::Args)]
pub struct Args {
    /// A folder inside the project's git repository (any worktree of the clone).
    project: PathBuf,
}

pub const OPERATION: &str = "goal.show";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::goal::show(&args.project)
}
