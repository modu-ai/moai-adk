//! `home-install-revert` (#2871 U3c): put the moved items back in `~/.moai-cli`. The logic is
//! `moai_cli::home_install_apply`.

use std::path::PathBuf;

/// Undo a ~/.moai-cli to ~/.moai move
///
/// Take the marker the apply wrote out of `~/.moai` into the install journal folder (an edited
/// marker is refused), point the `path` links back where they still point at the new target, and
/// replay the install journal in reverse (#2871): each entry moves back only while it is still the
/// one that was moved and its old place is free; an entry replaced since stays where it is and the
/// state is `revert_partial`. `~/.moai` goes only when the apply created it and it is empty. Uses
/// the newest install unless `--stamp` names one. Deletes nothing else.
#[derive(clap::Args)]
pub struct Args {
    /// The HOME to change. Required: the real HOME is never a default.
    #[arg(long)]
    home: PathBuf,
    /// The install to undo, `YYYYMMDDTHHMMSSZ`; the newest when absent.
    #[arg(long)]
    stamp: Option<String>,
    /// The caller's reference for the user's approval.
    #[arg(long)]
    permission_reference: String,
}

pub const OPERATION: &str = "home_install.revert";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::home_install_apply::revert(
        &args.home,
        args.stamp.as_deref(),
        &args.permission_reference,
    )
}
