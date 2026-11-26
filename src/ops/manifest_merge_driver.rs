//! `manifest-merge-driver` (SPEC-20261126-manifest-merge-driver-rust): the git merge driver of
//! the skill source manifest; the logic is `moai_cli::manifest_merge`. It reads no HOME, no
//! identity and no ledger.

use std::{ffi::OsString, path::PathBuf};

/// Git merge driver of the skill source manifest
///
/// Run by git, not by hand: it merges `skills/moai/references/source-manifest.json` three-way,
/// takes the sha256 of a skill file both sides changed from git's histogram merge of that file,
/// writes the result into the `%A` file and exits 0. When it cannot decide it leaves git's text
/// merge in the `%A` file, names the reason on standard error and exits 1. Register it once per
/// clone with the line in `docs/agents-md-maintenance.md`.
#[derive(clap::Args)]
pub struct Args {
    /// `%O`: the file holding the common version
    base: PathBuf,
    /// `%A`: the file holding our version; the result is written into it
    ours: PathBuf,
    /// `%B`: the file holding their version
    theirs: PathBuf,
    /// `%P`: the path of the manifest in the repository
    path: OsString,
}

pub const OPERATION: &str = "manifest.merge_driver";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::manifest_merge::drive(&args.base, &args.ours, &args.theirs, &args.path)
        // The shared command entry exits 1 for every operation error, which is `REFUSED_STATUS`;
        // `refused.status` is not read here.
        .map_err(|refused| refused.message)?;
    Ok(serde_json::json!({
        "path": args.path.to_string_lossy(),
        "merged": true,
    }))
}
