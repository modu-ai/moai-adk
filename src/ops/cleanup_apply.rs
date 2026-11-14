//! `cleanup-apply` (#2872 stage 1): move the archive items of a fresh cleanup plan into one
//! archive folder. The logic is `moai_cli::update_cleanup`.

use std::path::PathBuf;

/// Archive the previewed leftover MoAI files
///
/// Recompute the cleanup plan and, only when its `confirm_token` equals the one given, move
/// every archive item into `<state>/cleanup/<UTC>` (a new folder; `.2`, `.3` … in the same
/// second) beside `manifest.tsv` and `receipt.json` (#2872). Refuses before moving anything
/// when an item is on another file system. Deletes nothing.
#[derive(clap::Args)]
pub struct Args {
    /// User home root.
    home: PathBuf,
    /// Package directory carrying SKILL.md (the same stage the plan read).
    stage: PathBuf,
    /// The plan's `confirm_token`.
    #[arg(long, value_name = "TOKEN")]
    confirm_cleanup: String,
    /// The caller's reference for the user's approval.
    #[arg(long)]
    permission_reference: String,
    /// `YYYY-MM-DDTHH:MM:SSZ`; the system clock when absent (names the archive folder).
    #[arg(long)]
    now: Option<String>,
}

pub const OPERATION: &str = "cleanup.apply";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let staged = moai_cli::update::stage_package(&args.stage)?;
    let now = args.now.unwrap_or_else(moai_cli::audit::utc_now);
    moai_cli::update_cleanup::apply(
        &args.home,
        &staged,
        &args.confirm_cleanup,
        &args.permission_reference,
        &now,
    )
}
