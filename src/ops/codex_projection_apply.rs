//! `codex-projection-apply` (I4 #2241): installs a planned Codex projection into its explicit
//! target. Moved out of `src/main.rs` unchanged (#2834 R2a).

use moai_cli::hosts::codex;
use std::path::PathBuf;

/// Install a previewed Codex skill copy
///
/// Install a planned Codex projection into its explicit target (I4 #2241). Refuses a
/// stale plan, a source-manifest mismatch or a conflicting target file before writing.
#[derive(clap::Args)]
pub struct Args {
    package: PathBuf,
    /// Absolute skill directory named `moai`; the same target the plan was made for.
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
    #[arg(long)]
    allow_implicit_invocation: Option<bool>,
    #[arg(long)]
    plan_digest: String,
    #[arg(long)]
    permission_reference: String,
}

pub const OPERATION: &str = "codex.projection_apply";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    codex::projection_apply(
        &args.package,
        &args.target,
        &codex::OpenAiYaml {
            display_name: args.display_name,
            short_description: args.short_description,
            default_prompt: args.default_prompt,
            allow_implicit_invocation: args.allow_implicit_invocation,
        },
        &args.plan_digest,
        &args.permission_reference,
    )
}
