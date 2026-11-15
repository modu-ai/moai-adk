//! `home-install-plan` (#2871 U3a): which top-level entries of `~/.moai-cli` would move to the
//! `~/.moai` top level. Reads only; the logic is `moai_cli::home_install`.

use std::path::PathBuf;

/// Preview moving ~/.moai-cli items to ~/.moai
///
/// Classify each top-level entry of `~/.moai-cli` with the HOME name table (#2871): `move` (a
/// name this generation writes, renamed to `~/.moai/<name>`), `keep` (user data, the old shared
/// worktree root, a name not in the table, a link, the marker, the install journal) or
/// `blocked` (the name is already taken in `~/.moai`, ignoring case). Go items that have not
/// retired yet block the move. Lists the install id of each version folder before and after,
/// how each `path` entry would be re-pointed and the modes of the vault, judged by `lstat`
/// alone: a secret is never opened. Shows how many Rust sessions are registered in the clone at
/// `--project` (the apply refuses while one is) and says that other clones are not visible from
/// the HOME. Gives the `plan_digest` and the confirmation token. Writes nothing.
#[derive(clap::Args)]
pub struct Args {
    /// The HOME to read (default `$HOME`). Tests pass a temporary folder.
    #[arg(long)]
    home: Option<PathBuf>,
    /// `YYYY-MM-DDTHH:MM:SSZ`; the system clock when absent. Names the journal.
    #[arg(long)]
    now: Option<String>,
    /// A folder inside the clone whose Rust sessions are shown (default: the current folder).
    #[arg(long)]
    project: Option<PathBuf>,
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
    let mut plan = moai_cli::home_install::plan(&home, moai_home.as_deref(), &now)?;
    // The session facts are not part of either digest: a session that starts or ends between the
    // preview and the apply is judged again by the apply itself.
    let project = match args.project {
        Some(p) => Ok(p),
        None => std::env::current_dir().map_err(|e| format!("current folder unavailable: {e}")),
    };
    plan["sessions"] = match project {
        Ok(p) => moai_cli::home_install::sessions(&p),
        Err(e) => moai_cli::home_install::session_fact(&Err(e)),
    };
    Ok(plan)
}
