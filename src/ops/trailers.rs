//! `trailers` (SPEC-20261117-session-trailer, SDD-6): the identity trailers for a commit.

/// Print the identity trailers for a commit
///
/// Print the trailer lines every agent commit carries (user decision U-B 2026-11-17):
/// `Moai-Spec: <SPEC-id>` and, when the host gives this session's id (`CLAUDE_CODE_SESSION_ID`,
/// then `CODEX_THREAD_ID`, as `session-current` reads them) and the folder is in a git
/// worktree, `Moai-Session: <id>:<worktree key>` (12 hex digits of the sha256 of the
/// worktree's git dir: a subagent inherits its parent's id, the worktree tells them apart). Without an
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
    let key = std::env::current_dir()
        .ok()
        .and_then(|dir| moai_cli::trailers::worktree_key(&dir));
    moai_cli::trailers::format(&args.spec, |name| std::env::var(name).ok(), key.as_deref())
}
