//! `home-retire-revert` (#2871 U2): put the retired Go items back. The logic is
//! `moai_cli::home_retire`.

use std::path::PathBuf;

/// Undo a Go-item retire
///
/// Replay the retire journal in `~/.moai-go-retired` in reverse (#2871): each entry moves back
/// only while it is still the one that was moved and its old place is free; an entry replaced
/// since stays where it is and the state is `revert_partial`. Uses the newest retire journal
/// unless `--stamp` names one. Deletes nothing.
#[derive(clap::Args)]
pub struct Args {
    /// The HOME to change (default `$HOME`). Tests pass a temporary folder.
    #[arg(long)]
    home: Option<PathBuf>,
    /// The retire to undo, `YYYYMMDDTHHMMSSZ`; the newest when absent.
    #[arg(long)]
    stamp: Option<String>,
    /// The caller's reference for the user's approval.
    #[arg(long)]
    permission_reference: String,
}

pub const OPERATION: &str = "home_retire.revert";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let home = match args.home {
        Some(h) => h,
        None => std::env::var_os("HOME")
            .filter(|h| !h.is_empty())
            .map(PathBuf::from)
            .ok_or("HOME is not set; pass --home")?,
    };
    moai_cli::home_retire::revert(&home, args.stamp.as_deref(), &args.permission_reference)
}
