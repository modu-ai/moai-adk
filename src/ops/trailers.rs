//! `trailers` (SPEC-20261117-session-trailer, SDD-6): the identity trailers for a commit.

/// Print the identity trailers for a commit
///
/// Print the trailer lines every agent commit carries (user decision U-B 2026-11-17):
/// `Moai-Spec: <SPEC-id>` and, when the host gives this session's id (`CLAUDE_CODE_SESSION_ID`,
/// then `CODEX_THREAD_ID`, as `session-current` reads them), `Moai-Session: <id>`. Without an
/// id the result says so and carries the SPEC line only. Only a new-form SPEC id is accepted;
/// a card number is never a trailer. Reads the environment only; writes nothing.
#[derive(clap::Args)]
pub struct Args {
    /// The SPEC the commit belongs to, `SPEC-<yyyymmdd>-<slug>`.
    #[arg(long)]
    spec: String,
}

pub const OPERATION: &str = "trailers.format";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::trailers::format(&args.spec, |name| std::env::var(name).ok())
}
