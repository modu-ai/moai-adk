//! `home-state-plan` (#2723): whose `~/.moai` is and what moving the Go generation's folder
//! aside would do. Reads only; the logic is `moai_cli::home_state`.

use std::path::PathBuf;

/// Preview moving the Go-era ~/.moai folder
///
/// Preview the HOME state folder (#2723): classify `~/.moai` with the project's verdict table
/// and the HOME's Go markers, name the backup (`~/.moai-backup`, or `~/.moai-backup-<UTC>` when
/// that exists; never overwritten), list what blocks the move (a live Go session, a process
/// inside, a nested checkout, too large to check) and the entries the user may copy back
/// (default none; secrets never offered). Writes nothing.
#[derive(clap::Args)]
pub struct Args {
    /// The HOME to read (default `$HOME`). Tests pass a temporary folder.
    #[arg(long)]
    home: Option<PathBuf>,
    /// `YYYY-MM-DDTHH:MM:SSZ`; the system clock when absent.
    #[arg(long)]
    now: Option<String>,
    /// Classify each top-level entry instead of moving the whole folder (#2815):
    /// keep-in-place, move-to-backup, secret-manual (never opened) or unknown (moves nothing).
    #[arg(long)]
    selective: bool,
}

pub const OPERATION: &str = "home_state.plan";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let home = match args.home {
        Some(h) => h,
        None => std::env::var_os("HOME")
            .filter(|h| !h.is_empty())
            .map(PathBuf::from)
            .ok_or("HOME is not set; pass --home")?,
    };
    let moai_home = std::env::var("MOAI_HOME").ok();
    let now = args.now.unwrap_or_else(moai_cli::audit::utc_now);
    if args.selective {
        return moai_cli::home_state::plan_selective(&home, moai_home.as_deref(), &now);
    }
    moai_cli::home_state::plan(&home, moai_home.as_deref(), &now)
}
