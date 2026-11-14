//! `agents-merge-plan` (#2936 R10): what merging a project's `AGENTS.local.md` into
//! `AGENTS.md` would do. Reads only; the logic is `moai_cli::agents_local_merge`.

use std::path::PathBuf;

/// Preview merging AGENTS.local.md into AGENTS.md
///
/// Preview merging the text of a project's `AGENTS.local.md` into `AGENTS.md` (#2936 R10): the
/// source (the file in place, or the copy a `go-state-apply` moved into its backup), the whole
/// text, where it lands (at the end of `AGENTS.md`, outside the managed section) and that it
/// goes into a committed file. A line shaped like a credential blocks the plan. Writes nothing.
#[derive(clap::Args)]
pub struct Args {
    /// The project (default: the current folder).
    #[arg(default_value = ".")]
    project: PathBuf,
}

pub const OPERATION: &str = "agents_merge.plan";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::agents_local_merge::plan(&args.project)
}
