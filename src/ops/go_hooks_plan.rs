//! `go-hooks-plan` (#2877): what replacing a project's Go-generation hooks would do. Reads
//! only; the logic is `moai_cli::go_hooks`.

use std::path::PathBuf;

/// Preview replacing Go-era project hooks
///
/// Preview replacing the Go-generation hooks of a project (#2877): the `.claude/settings.json`
/// entries that run `.claude/hooks/moai/*` go, every other entry stays, the MoAI hook set is
/// added, and the scripts move to an archive folder (never deleted). Entries of other tools in
/// `hooks` block the plan for now. Writes nothing.
#[derive(clap::Args)]
pub struct Args {
    /// The project (default: the current folder).
    #[arg(default_value = ".")]
    project: PathBuf,
}

pub const OPERATION: &str = "go_hooks.plan";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let home = std::env::var_os("HOME").map(PathBuf::from);
    moai_cli::go_hooks::plan(&args.project, home.as_deref())
}
