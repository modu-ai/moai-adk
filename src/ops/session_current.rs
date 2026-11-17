//! `session-current` (#3015 C4): this session's own id from the host's environment; the logic
//! is `moai_cli::session_current`.

/// Print this session's own id
///
/// Print the session id the host put in this process's environment (`CLAUDE_CODE_SESSION_ID`,
/// then `CODEX_THREAD_ID`), or `null` with the Go build's fallback line when neither is set to
/// an id-shaped value (#3015 C4). It never reads the per-project session record, which names
/// whichever session started last, not this one. Reads and writes no file.
#[derive(clap::Args)]
pub struct Args {}

pub const OPERATION: &str = "session.current";

pub fn run(_args: Args) -> Result<serde_json::Value, String> {
    Ok(moai_cli::session_current::resolve(|name| {
        std::env::var(name).ok()
    }))
}
