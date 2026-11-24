//! `home-archive-plan` (SPEC-20261123-home-state-retire): preview archiving the earlier HOME
//! state folder under `~/.moai/archive/`. The logic is `moai_cli::home_state_retire`.

use std::path::PathBuf;

/// Preview archiving the earlier HOME state folder
///
/// List each top-level entry of the earlier state folder with its kind and mode, read by `lstat`
/// alone (no file is opened, no link followed), and plan one rename of the whole folder to
/// `~/.moai/archive/moai-cli-<UTC>/`. Reports `none` when there is no earlier folder and `blocked`
/// (with the reasons) when it is a link, when `~/.moai` or its `archive` folder is not a folder,
/// when the archive name is taken, when the two are on different file systems, when a program uses
/// a path under the folder or the running binary lies inside it. Shows the `PATH` entries of this
/// process that name the earlier folder and the guidance to run `moai update`; no shell file is
/// read or edited. Gives the `plan_digest` and the `content_digest`. Writes nothing.
#[derive(clap::Args)]
pub struct Args {
    /// The HOME to read (default `$HOME`). Tests pass a temporary folder.
    #[arg(long)]
    home: Option<PathBuf>,
    /// `YYYY-MM-DDTHH:MM:SSZ`; the system clock when absent. Names the archive.
    #[arg(long)]
    now: Option<String>,
}

pub const OPERATION: &str = "home_archive.plan";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let home = match args.home {
        Some(h) => h,
        None => std::env::var_os("HOME")
            .filter(|h| !h.is_empty())
            .map(PathBuf::from)
            .ok_or("HOME is not set; pass --home")?,
    };
    let now = args.now.unwrap_or_else(moai_cli::audit::utc_now);
    moai_cli::home_state_retire::plan(&home, &now, &moai_cli::home_state_retire::Facts::current())
}
