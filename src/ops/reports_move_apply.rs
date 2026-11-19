//! `reports-move-apply` (C-126): move the previewed untracked report files to the user-level
//! reports folder after the user confirmed the plan. The logic is `moai_cli::reports_move`.

use std::path::PathBuf;

/// Move the earlier report files out of the repository
///
/// Move the files `reports-move-plan` listed to `<HOME>/.moai/reports/<project>`, keeping the
/// relative paths, after the user confirmed that plan. The plan comes on stdin and is made again:
/// a file added, moved or tracked since the preview makes it stale and nothing moves. A name that
/// already exists at the destination is skipped, never overwritten; folders made are owner-only
/// (0700) and moved files lose their group and other permission bits. Running it again moves
/// nothing.
#[derive(clap::Args)]
pub struct Args {
    /// The `plan_digest` of the confirmed `reports-move-plan`.
    #[arg(long)]
    plan_digest: String,
    /// The caller's reference for the user's approval.
    #[arg(long)]
    permission_reference: String,
    /// The project (default: the current folder).
    #[arg(default_value = ".")]
    project: PathBuf,
}

pub const OPERATION: &str = "reports_move.apply";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let home = moai_cli::home_workdir::home()?;
    moai_cli::reports_move::apply(
        &args.project,
        &home,
        super::super::plan_input::<moai_cli::reports_move::MovePlan>()?,
        &args.plan_digest,
        &args.permission_reference,
    )
}
