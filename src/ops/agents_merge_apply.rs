//! `agents-merge-apply` (#2936 R10): merge a project's `AGENTS.local.md` into `CLAUDE.local.md`
//! after the user confirmed the plan. The logic is `moai_cli::agents_local_merge`.

use std::path::PathBuf;

/// Merge AGENTS.local.md into CLAUDE.local.md
///
/// Merge the text of `AGENTS.local.md` into `CLAUDE.local.md`, the personal file Claude Code
/// reads (#2936 R10), after the user confirmed `agents-merge-plan`: the text is added once at the
/// end of `CLAUDE.local.md` between two marker lines (the file is created when it is missing), the
/// one line `CLAUDE.local.md` goes into `.gitignore` when git does not ignore it yet, and a
/// receipt keeps the earlier files. The source file is not moved or deleted; `agents-merge-revert`
/// undoes a finished apply. A plan that warns Claude Code would stop reading `AGENTS.md` is
/// refused unless `--confirm-warnings` carries the token printed with that plan.
#[derive(clap::Args)]
pub struct Args {
    /// The `plan_digest` of the confirmed `agents-merge-plan`.
    #[arg(long)]
    plan_digest: String,
    /// The caller's reference for the user's approval.
    #[arg(long)]
    permission_reference: String,
    /// The `confirm_token` of a plan that carries a warning, passed only after the user read the
    /// warning and said yes to it.
    #[arg(long)]
    confirm_warnings: Option<String>,
    /// The project (default: the current folder).
    #[arg(default_value = ".")]
    project: PathBuf,
}

pub const OPERATION: &str = "agents_merge.apply";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::agents_local_merge::apply_confirmed(
        moai_cli::agents_local_merge::Kind::Local,
        &args.project,
        &args.plan_digest,
        &args.permission_reference,
        args.confirm_warnings.as_deref(),
    )
}
