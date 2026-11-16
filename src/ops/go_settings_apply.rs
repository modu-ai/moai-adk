//! `go-settings-apply` (#2936 L2a): clean the Go-made `.moai/` references out of a project's
//! `.claude/settings.json` after the user confirmed the plan. The logic is
//! `moai_cli::go_settings_clean`.

use std::path::PathBuf;

/// Clean Go-era settings keys
///
/// Clean the Go-made `.moai/` references out of `.claude/settings.json` after the user confirmed
/// `go-settings-plan` (#2936 L2a): the old file is kept in a receipt folder first, the Go-made
/// keys go or change, every other key stays, and `.claude/settings.local.json` is never touched.
/// `go-settings-revert` undoes a finished apply.
#[derive(clap::Args)]
pub struct Args {
    /// The `plan_digest` of the confirmed `go-settings-plan`.
    #[arg(long)]
    plan_digest: String,
    /// The caller's reference for the user's approval.
    #[arg(long)]
    permission_reference: String,
    /// The project (default: the current folder).
    #[arg(default_value = ".")]
    project: PathBuf,
}

pub const OPERATION: &str = "go_settings.apply";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let home = std::env::var_os("HOME").map(PathBuf::from);
    moai_cli::go_settings_clean::apply(
        &args.project,
        home.as_deref(),
        &args.plan_digest,
        &args.permission_reference,
    )
}
