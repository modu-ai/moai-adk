//! `home-retire-plan` (#2871 U2): which top-level entries of the Go generation's `~/.moai`
//! would move to a dated backup. Reads only; the logic is `moai_cli::home_retire`.

use std::path::PathBuf;

/// Preview retiring the Go items in ~/.moai
///
/// Classify each top-level entry of `~/.moai` with the HOME name table (#2871): `retire` (a Go
/// item that moves to `~/.moai-go-retired/<UTC>/`), `keep` (user data, unknown owner, a name
/// not in the table, this generation's own name, a key file still waiting for the merge step,
/// a link) or `skip-in-use` (a running program uses it). A secret is judged by its name only,
/// never opened. Lists what blocks the move and the `plan_digest` the apply needs. Writes
/// nothing.
#[derive(clap::Args)]
pub struct Args {
    /// The HOME to read (default `$HOME`). Tests pass a temporary folder.
    #[arg(long)]
    home: Option<PathBuf>,
    /// `YYYY-MM-DDTHH:MM:SSZ`; the system clock when absent. Names the dated folder.
    #[arg(long)]
    now: Option<String>,
}

pub const OPERATION: &str = "home_retire.plan";

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
    moai_cli::home_retire::plan(&home, moai_home.as_deref(), &now)
}
