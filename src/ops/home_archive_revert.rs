//! `home-archive-revert` (SPEC-20261123-home-state-retire): put the archived earlier HOME
//! state folder back. The logic is `moai_cli::home_state_retire`.

use std::path::PathBuf;

/// Undo the archiving of the earlier HOME state folder
///
/// Replay the archive journal in reverse: the folder goes back to its old place only while it is
/// still the entry that was moved and the old place is free; an entry replaced since stays where
/// it is and the state is `revert_partial`. `--stamp` names the archive; the newest otherwise.
/// Deletes nothing.
#[derive(clap::Args)]
pub struct Args {
    /// The HOME to change (default `$HOME`). Tests pass a temporary folder.
    #[arg(long)]
    home: Option<PathBuf>,
    /// The archive to undo, `YYYYMMDDTHHMMSSZ`; the newest when absent.
    #[arg(long)]
    stamp: Option<String>,
    /// The caller's reference for the user's approval.
    #[arg(long)]
    permission_reference: String,
}

pub const OPERATION: &str = "home_archive.revert";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let home = match args.home {
        Some(h) => h,
        None => std::env::var_os("HOME")
            .filter(|h| !h.is_empty())
            .map(PathBuf::from)
            .ok_or("HOME is not set; pass --home")?,
    };
    moai_cli::home_state_retire::revert(&home, args.stamp.as_deref(), &args.permission_reference)
}
