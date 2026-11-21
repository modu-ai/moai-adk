//! `agents-merge-plan` (#2936 R10): what merging a project's `AGENTS.local.md` into
//! `CLAUDE.local.md` would do. Reads only; the logic is `moai_cli::agents_local_merge`.

use std::path::PathBuf;

/// Preview merging AGENTS.local.md into CLAUDE.local.md
///
/// Preview merging the text of a project's `AGENTS.local.md` into `CLAUDE.local.md`, the personal
/// file Claude Code reads (#2936 R10): the source (the file in place, or the copy a
/// `go-state-apply` moved into its backup), the whole text, where it lands (at the end of
/// `CLAUDE.local.md`, which is created when it is missing), whether git ignores that file or the
/// one line `CLAUDE.local.md` would be added to `.gitignore`, and any warning, such as Claude Code
/// no longer reading `AGENTS.md` once `CLAUDE.local.md` exists (the apply then needs the
/// `confirm_token` printed here). A line shaped like a credential blocks the plan. Writes nothing.
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
