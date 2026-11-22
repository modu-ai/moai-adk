//! `agents-merge-revert` (#2936 R10): undo one `agents-merge-apply`. The logic is
//! `moai_cli::agents_local_merge`.

use std::path::PathBuf;

/// Undo an AGENTS.local.md merge
///
/// Undo one `agents-merge-apply` (#2936 R10) while `CLAUDE.local.md` still holds the bytes the
/// apply wrote: it gets its earlier bytes back, or is removed when the apply created it. The
/// `.gitignore` line the apply added is taken back too, unless that file changed since (it is then
/// kept as it is). The source file and its backup are not touched. A receipt of an earlier
/// generation, which wrote into `AGENTS.md`, still restores `AGENTS.md`.
///
/// A receipt `agents-pick-…` of `agents-merge-apply --pick` is undone part by part in the reverse
/// order of the apply: the personal text and its `.gitignore` line, the text of `AGENTS.md` or the
/// import line, then the managed block. A file that changed after the apply is named under `kept`
/// and left as it is; the other files still go back.
#[derive(clap::Args)]
pub struct Args {
    /// The receipt id the apply printed (`agents-local-merge-…` or `agents-pick-…`).
    #[arg(long)]
    receipt: String,
    /// The caller's reference for the user's approval.
    #[arg(long)]
    permission_reference: String,
    /// The project (default: the current folder).
    #[arg(default_value = ".")]
    project: PathBuf,
}

pub const OPERATION: &str = "agents_merge.revert";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    if args
        .receipt
        .starts_with(moai_cli::agents_local_merge::picks::RECEIPT_PREFIX)
    {
        return moai_cli::agents_local_merge::picks::revert(
            &args.project,
            &args.receipt,
            &args.permission_reference,
        );
    }
    moai_cli::agents_local_merge::revert(&args.project, &args.receipt, &args.permission_reference)
}
