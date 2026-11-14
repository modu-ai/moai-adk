//! `language-setup` (#2293): per-item writing-language setup; writes nothing. Moved out of `src/main.rs` unchanged (#2834 R4a).

use std::path::PathBuf;

/// Choose the writing language item by item
///
/// Per-item writing-language setup (#2293): asks one question per item and returns a
/// complete profile for `project-plan`; writes nothing. `--questions` prints the question
/// set for a host question tool, `--answers` reads that tool's `{item: code}` JSON on
/// stdin, and neither flag asks on the terminal (prompts on stderr, answers on stdin).
#[derive(clap::Args)]
pub struct Args {
    project: PathBuf,
    #[arg(long, conflicts_with = "answers")]
    questions: bool,
    #[arg(long)]
    answers: bool,
    /// Conversation language code for the conversation-default items; else the
    /// profile's `language.user_chat`, else `en`.
    #[arg(long)]
    conversation: Option<String>,
}

pub const OPERATION: &str = "language.setup";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    use moai_cli::language::{setup, SetupMode};
    let mode = if args.questions {
        SetupMode::Questions
    } else if args.answers {
        let value = super::super::input::<serde_json::Value>()?;
        SetupMode::Answers(
            value
                .as_object()
                .cloned()
                .ok_or("answers must be a JSON object")?,
        )
    } else {
        SetupMode::Interactive
    };
    setup(
        &args.project,
        mode,
        args.conversation.as_deref(),
        &mut std::io::stdin().lock(),
        &mut std::io::stderr(),
    )
}
