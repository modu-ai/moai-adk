//! `git-hooks-revert` (SPEC-20261121-go-path-removal): put the retired hooks of an applied
//! git-hook change back; the logic is `moai_cli::git_hooks`.

use std::path::PathBuf;

/// Revert a git-hook apply
///
/// Put the retired hooks back from `moai-retired-<stamp>/` (SPEC-20261121-go-path-removal):
/// only over hooks this feature installed (byte-equal to its own texts) or that were absent,
/// and only while the retire folder holds nothing else.
#[derive(clap::Args)]
pub struct Args {
    /// The stamp the apply reported.
    stamp: String,
    /// The repository (default: the current folder).
    #[arg(default_value = ".")]
    project: PathBuf,
}

pub const OPERATION: &str = "git_hooks.revert";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::git_hooks::revert(&args.project, &args.stamp)
}
