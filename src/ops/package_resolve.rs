//! `package-resolve`: which skill package a project uses. Moved out of `src/main.rs`
//! unchanged (#2834 R2c).

use std::path::PathBuf;

/// Resolve which skill package a project uses
#[derive(clap::Args)]
pub struct Args {
    project: PathBuf,
    #[arg(long)]
    repo_package: Option<PathBuf>,
    #[arg(long)]
    user_package: Option<PathBuf>,
}

pub const OPERATION: &str = "package.resolve";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::resolve(
        &args.project,
        args.repo_package.as_deref(),
        args.user_package.as_deref(),
    )
}
