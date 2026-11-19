//! `trailers` (SPEC-20261117-session-trailer, SDD-6; SPEC-20261119-verified-trailer): the
//! identity trailers for a commit, and the verification kinds of a range.

use std::path::PathBuf;

/// Print the identity trailers for a commit, or list a range by verification kind
///
/// Print the trailer lines every agent commit carries (user decision U-B 2026-11-17):
/// `Moai-Spec: <SPEC-id>` and, when the host gives this session's id (`CLAUDE_CODE_SESSION_ID`,
/// then `CODEX_THREAD_ID`, as `session-current` reads them) and the folder is in a git
/// worktree, `Moai-Session: <id>:<worktree key>` (12 hex digits of the sha256 of the
/// worktree's git dir: a subagent inherits its parent's id, the worktree tells them apart). With
/// `--verified` one `Moai-Verified: <kinds>` line is appended, every kind one of `tests`,
/// `mutants`, `e2e`, `smoke`, `none` (comma-separated, order kept); an unknown kind is refused
/// and no text is printed. Without an id the result says so and carries the SPEC line only. Only
/// a new-form SPEC id is accepted; a card number is never a trailer. Reads the environment only;
/// writes nothing.
///
/// `--list` reads `--range` (default `HEAD`) of `--dir` instead and groups the commits by
/// verification kind — a commit with several kinds appears in each, a missing line counts as
/// `none` — and marks the commits a later standard revert message names, so `none` and reverted
/// commits can be counted together. Reads only.
#[derive(clap::Args)]
pub struct Args {
    /// The SPEC the commit belongs to, `SPEC-<yyyymmdd>-<slug>`.
    #[arg(long, required_unless_present = "list")]
    spec: Option<String>,
    /// How the commit was verified: tests, mutants, e2e, smoke or none; repeat or comma-separate.
    #[arg(long, value_delimiter = ',', num_args = 1.., requires = "spec")]
    verified: Vec<String>,
    /// List the commits of a range by verification kind instead of printing one commit's trailers.
    #[arg(long, conflicts_with_all = ["spec", "verified"])]
    list: bool,
    /// The range to list, for example `<base>..HEAD` (with --list).
    #[arg(long, default_value = "HEAD", requires = "list")]
    range: String,
    /// Repository to read (with --list).
    #[arg(long, default_value = ".", requires = "list")]
    dir: PathBuf,
}

pub const OPERATION: &str = "trailers.format";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    if args.list {
        return moai_cli::trailers::list(&args.dir, &args.range);
    }
    let spec = args.spec.as_deref().unwrap_or_default();
    let key = std::env::current_dir()
        .ok()
        .and_then(|dir| moai_cli::trailers::worktree_key(&dir));
    moai_cli::trailers::format_verified(spec, &args.verified, |name| std::env::var(name).ok(), key.as_deref())
}
