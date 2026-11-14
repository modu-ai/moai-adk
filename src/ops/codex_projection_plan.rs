//! `codex-projection-plan`: plans a Codex projection of the common package into an explicit
//! target; writes nothing. Moved out of `src/main.rs` unchanged (#2834 R2a).

use moai_cli::hosts::codex;
use std::path::PathBuf;

/// Preview a Codex copy of the skill package
///
/// Plan a Codex projection of the common package into an explicit target; writes nothing.
#[derive(clap::Args)]
pub struct Args {
    package: PathBuf,
    /// Absolute skill directory named `moai`, e.g. ~/.agents/skills/moai (no default).
    #[arg(long)]
    target: PathBuf,
    #[arg(long, default_value = "MoAI")]
    display_name: String,
    #[arg(
        long,
        default_value = "Set up, plan, implement, verify, audit and sync through one workflow"
    )]
    short_description: String,
    #[arg(long, default_value = "Use $moai to read this project's status")]
    default_prompt: String,
    /// Write `policy.allow_implicit_invocation`; omitted keeps Codex's default.
    #[arg(long)]
    allow_implicit_invocation: Option<bool>,
    /// Requested `tui.status_line` item (repeatable); reported as HOLD, never written.
    #[arg(long = "status-line")]
    status_line: Vec<String>,
    /// Requested `[skills.bundled] enabled = false`; reported as HOLD, never written.
    #[arg(long)]
    disable_bundled_skills: bool,
}

pub const OPERATION: &str = "codex.projection_plan";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    codex::projection_plan(
        &args.package,
        &args.target,
        &codex::OpenAiYaml {
            display_name: args.display_name,
            short_description: args.short_description,
            default_prompt: args.default_prompt,
            allow_implicit_invocation: args.allow_implicit_invocation,
        },
        &args.status_line,
        args.disable_bundled_skills,
    )
}
