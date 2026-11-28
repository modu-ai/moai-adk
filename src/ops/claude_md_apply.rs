//! `claude-md-apply`: replace a Go-generation project's instruction files after the user
//! confirmed `claude-md-plan`. The logic is `moai_cli::project_update::claude_md`.

use std::path::PathBuf;

/// Replace the old instruction files with new ones
///
/// Apply the `claude-md-plan` preview the user confirmed, with the same `--keep` picks: every
/// instruction file the preview lists moves into an owner-only folder under the state folder's
/// `backups/`, at its own path, and then the new `CLAUDE.md` is written (the managed block of the
/// profile, then the hunks kept as `shared`). A hunk kept as `personal` also writes
/// `CLAUDE.local.md`, readable by its owner alone, and adds the one line `CLAUDE.local.md` to
/// `.gitignore` when no rule ignores that file. No file of yours is deleted or overwritten
/// (a failed apply removes only what it wrote itself); emptied folders stay.
///
/// It refuses before the first write, naming the reason, without a permission reference, while
/// the preview lists a reason under `apply_blocked` (the root `CLAUDE.md` is no Go-generation
/// file, a source is refused, the profile is unusable, a shared hunk holds a credential-shaped
/// line, git would not ignore the backup folder or the personal file, a link stands in the way)
/// and when the sources, the profile or the picks differ from the previewed digest.
///
/// A receipt in the backup folder is saved before the first move and after every step. A step
/// that fails is undone from that receipt and the error names the step: the receipt then reads
/// `restored`, or `partial` with every path that could not be put back, and `claude-md-revert`
/// undoes the rest. An apply that was stopped leaves the receipt `pending`; `claude-md-revert`
/// undoes it too. Settings apply at once; running sessions and lanes read changed instructions
/// only after a restart.
#[derive(clap::Args)]
pub struct Args {
    /// The project (default: the current folder).
    #[arg(default_value = ".")]
    project: PathBuf,
    /// Keep one hunk of the preview: `<id>=shared` (the new `CLAUDE.md`) or `<id>=personal`
    /// (`CLAUDE.local.md`). Repeat once per hunk, as the printed command does.
    #[arg(long = "keep", value_name = "ID=DESTINATION")]
    keep: Vec<String>,
    /// The `plan_digest` of the confirmed `claude-md-plan`.
    #[arg(long)]
    plan_digest: String,
    /// The caller's reference for the user's approval.
    #[arg(long)]
    permission_reference: String,
}

pub const OPERATION: &str = "claude_md.apply";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::project_update::claude_md::apply(
        &args.project,
        &args.keep,
        &args.plan_digest,
        &args.permission_reference,
    )
}
