//! `bin-names-plan` (#2791): preview the `moai` hard link in this build's version folder.
//! Reads only; the logic is `moai_cli::bin_names`.

use std::path::PathBuf;

/// Preview the two command names (#2791 AC-1): in the version folder this binary runs from,
/// `~/.moai-cli/versions/<version>/bin/`, `moai` should be a hard link to `moai-cli`. Says
/// `create`, `unchanged`, `replace_copy` (a separate copy of this same build is moved aside),
/// `blocked` (a link, a folder or another program named `moai`) or `outside_layout`. Writes
/// nothing; PATH and shell files are never touched.
#[derive(clap::Args)]
pub struct Args {
    /// The HOME to read (default `$HOME`). Tests pass a temporary folder.
    #[arg(long)]
    home: Option<PathBuf>,
}

pub const OPERATION: &str = "bin_names.plan";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let home = match args.home {
        Some(h) => h,
        None => std::env::var_os("HOME")
            .filter(|h| !h.is_empty())
            .map(PathBuf::from)
            .ok_or("HOME is not set; pass --home")?,
    };
    let exe = std::env::current_exe().map_err(|_| "current executable unavailable")?;
    Ok(moai_cli::bin_names::plan_for(&home, &exe)?.to_json())
}
