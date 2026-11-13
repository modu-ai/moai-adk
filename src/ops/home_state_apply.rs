//! `home-state-apply` (#2723): move the Go generation's `~/.moai` aside and create the Rust
//! one, only on an identical fresh preview. The logic is `moai_cli::home_state`.

use std::path::PathBuf;

/// Move a previewed Go-era ~/.moai folder
///
/// Move the Go `~/.moai` aside to the backup the preview named (never overwriting), create
/// the Rust `~/.moai`, copy only the entries named with `--copy` (never secrets or Go
/// markers), write the receipt and then the Rust marker (#2723). Runs only when the preview
/// recomputed from the same inputs has the same `plan_digest` and the action `migrate`. On a
/// real HOME the leader runs it only after the user confirms.
#[derive(clap::Args)]
pub struct Args {
    /// The HOME to change (default `$HOME`). Tests pass a temporary folder.
    #[arg(long)]
    home: Option<PathBuf>,
    /// The same `--now` the preview read.
    #[arg(long)]
    now: String,
    /// The preview's `plan_digest`.
    #[arg(long)]
    plan_digest: String,
    /// The caller's reference for the user's approval.
    #[arg(long)]
    permission_reference: String,
    /// A top-level entry of `~/.moai` to copy into the Rust folder; repeat for more.
    #[arg(long)]
    copy: Vec<String>,
}

pub const OPERATION: &str = "home_state.apply";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let home = match args.home {
        Some(h) => h,
        None => std::env::var_os("HOME")
            .filter(|h| !h.is_empty())
            .map(PathBuf::from)
            .ok_or("HOME is not set; pass --home")?,
    };
    let moai_home = std::env::var("MOAI_HOME").ok();
    moai_cli::home_state::apply(
        &home,
        moai_home.as_deref(),
        &args.now,
        &args.plan_digest,
        &args.permission_reference,
        &args.copy,
    )
}
