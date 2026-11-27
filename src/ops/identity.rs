//! `identity`: read-only clone/workspace identity and state roots of a path. Moved out of
//! `src/main.rs` unchanged (#2834 R3).

use std::path::PathBuf;

/// Show where a path sits in its clone
///
/// Read-only clone/workspace identity and state roots of a path (typed JSON on stdout).
#[derive(clap::Args)]
pub struct Args {
    path: PathBuf,
}

pub const OPERATION: &str = "identity";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    // The one read of the environment HOME for this command; an unset or empty HOME is no HOME.
    let home = std::env::var_os("HOME")
        .filter(|h| !h.is_empty())
        .map(PathBuf::from);
    moai_cli::identity::identity_in(&args.path, home.as_deref())
}
