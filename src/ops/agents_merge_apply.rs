//! `agents-merge-apply` (#2936 R10): merge a project's `AGENTS.local.md` into
//! `AGENTS.md` after the user confirmed the plan. The logic is `moai_cli::agents_local_merge`.

use std::path::PathBuf;

/// Merge AGENTS.local.md into AGENTS.md
///
/// Merge the text of `AGENTS.local.md` into `AGENTS.md` (#2936 R10) after the user confirmed
/// `agents-merge-plan`: the text is added once at the end of `AGENTS.md` between two
/// marker lines, and a receipt keeps the earlier `AGENTS.md`. The source file is not moved or
/// deleted; `agents-merge-revert` undoes a finished apply.
#[derive(clap::Args)]
pub struct Args {
    /// The `plan_digest` of the confirmed `agents-merge-plan`.
    #[arg(long)]
    plan_digest: String,
    /// The caller's reference for the user's approval.
    #[arg(long)]
    permission_reference: String,
    /// The project (default: the current folder).
    #[arg(default_value = ".")]
    project: PathBuf,
}

pub const OPERATION: &str = "agents_merge.apply";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::agents_local_merge::apply(
        &args.project,
        &args.plan_digest,
        &args.permission_reference,
    )
}
