//! `home-install-plan` (#2871 U3a): which top-level entries of `~/.moai-cli` would move to the
//! `~/.moai` top level. Reads only; the logic is `moai_cli::home_install`.

use std::path::PathBuf;

/// Preview moving this generation's items from ~/.moai-cli to ~/.moai
///
/// Classify each top-level entry of `~/.moai-cli` with the HOME name table (#2871): `move` (a
/// name this generation writes, renamed to `~/.moai/<name>`), `keep` (user data, the old shared
/// worktree root, a name not in the table, a link, the marker, the install journal) or
/// `blocked` (the name is already taken in `~/.moai`, ignoring case). Go items that have not
/// retired yet block the move. Lists the install id of each version folder before and after,
/// how each `path` entry would be re-pointed and the modes of the vault, judged by `lstat`
/// alone: a secret is never opened. Gives the `plan_digest` and the confirmation token. Writes
/// nothing.
#[derive(clap::Args)]
pub struct Args {
    /// The HOME to read (default `$HOME`). Tests pass a temporary folder.
    #[arg(long)]
    home: Option<PathBuf>,
    /// `YYYY-MM-DDTHH:MM:SSZ`; the system clock when absent. Names the journal.
    #[arg(long)]
    now: Option<String>,
}

pub const OPERATION: &str = "home_install.plan";

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
    moai_cli::home_install::plan(&home, moai_home.as_deref(), &now)
}
