//! `claude-settings-plan` (#2239): previews a three-way merge of the managed Claude Code settings keys. Moved out of `src/main.rs` unchanged (#2834 R4b).

use std::path::PathBuf;

/// Preview the managed Claude Code settings keys
///
/// Preview a three-way merge of the managed keys (`outputStyle`, `statusLine`, `permissions`) on stdin
/// into PROJECT/.claude/settings.json (#2239); other keys are refused. Writes nothing.
#[derive(clap::Args)]
pub struct Args {
    project: PathBuf,
    /// `project` (default): .claude/settings.json. `local`: .claude/settings.local.json, the
    /// file Claude Code also reads for sessions in a linked worktree (#2581).
    #[arg(long, default_value = "project")]
    scope: String,
}

pub const OPERATION: &str = "claude.settings_plan";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let scope = moai_cli::claude::SettingsScope::parse(&args.scope)?;
    let desired =
        moai_cli::claude::user_settings_desired_in(&args.project, &super::super::input()?, scope)?;
    moai_cli::claude::settings_plan_in(&args.project, &desired, scope)
}
