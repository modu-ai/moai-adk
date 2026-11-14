//! `goal-plan` (#2679): preview a change of the standing goal; writes nothing.

use std::path::PathBuf;

/// Preview a change of the standing goal
///
/// Preview a change of the standing goal (#2679): `change` is `create`, `update`, `clear`,
/// `unchanged` or `none`, with the changed fields and the `plan_digest` that `goal-apply`
/// needs. `asks_person` is true only when something would change. Writes nothing.
#[derive(clap::Args)]
pub struct Args {
    /// A folder inside the project's git repository (any worktree of the clone).
    project: PathBuf,
    /// JSON request: `schema_version`, `action` (`set`|`clear`), `text`, `end_condition`,
    /// `exclusions` (`number`, `reason`), `source`.
    #[arg(long)]
    request: PathBuf,
}

pub const OPERATION: &str = "goal.plan";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let request = moai_cli::goal::read_request(&args.request)?;
    moai_cli::goal::plan(&args.project, &request)
}
