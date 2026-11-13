//! `home-state-revert` (#2723): put the Go `~/.moai` back after an apply. The logic is
//! `moai_cli::home_state`.

use std::path::PathBuf;

/// Put the Go `~/.moai` back (#2723): the Rust `~/.moai` an apply made (valid marker and
/// receipt) moves aside to `~/.moai.rust-<UTC>`, and the backup its receipt names moves back
/// to `~/.moai`. Deletes nothing.
#[derive(clap::Args)]
pub struct Args {
    /// The HOME to change (default `$HOME`). Tests pass a temporary folder.
    #[arg(long)]
    home: Option<PathBuf>,
    /// `YYYY-MM-DDTHH:MM:SSZ`; the system clock when absent (names the aside folder).
    #[arg(long)]
    now: Option<String>,
    /// The caller's reference for the user's approval.
    #[arg(long)]
    permission_reference: String,
}

pub const OPERATION: &str = "home_state.revert";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let home = match args.home {
        Some(h) => h,
        None => std::env::var_os("HOME")
            .filter(|h| !h.is_empty())
            .map(PathBuf::from)
            .ok_or("HOME is not set; pass --home")?,
    };
    let now = args.now.unwrap_or_else(moai_cli::audit::utc_now);
    moai_cli::home_state::revert(&home, &now, &args.permission_reference)
}
