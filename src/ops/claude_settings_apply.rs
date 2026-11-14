//! `claude-settings-apply`: applies a previewed Claude Code settings merge. Moved out of `src/main.rs` unchanged (#2834 R4b).

use std::path::PathBuf;

/// Write previewed Claude Code settings keys
///
/// Apply a previewed settings merge: same stdin, the preview's digest and a permission
/// reference. Writes the project settings file and its receipt only; home is refused.
#[derive(clap::Args)]
pub struct Args {
    project: PathBuf,
    #[arg(long)]
    plan_digest: String,
    #[arg(long)]
    permission_reference: String,
    /// The scope the plan was made for (#2581); the digest binds it.
    #[arg(long, default_value = "project")]
    scope: String,
    /// The `confirm_token` a conflict plan offered (#2383): after the user confirms, keep
    /// the existing file in `.moai-cli/settings-backups/` and replace the conflicting keys.
    #[arg(long)]
    confirm_replace: Option<String>,
}

pub const OPERATION: &str = "claude.settings_apply";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let scope = moai_cli::claude::SettingsScope::parse(&args.scope)?;
    let desired =
        moai_cli::claude::user_settings_desired_in(&args.project, &super::super::input()?, scope)?;
    moai_cli::claude::settings_apply_in(
        &args.project,
        &desired,
        &args.plan_digest,
        &args.permission_reference,
        scope,
        args.confirm_replace.as_deref(),
    )
}
