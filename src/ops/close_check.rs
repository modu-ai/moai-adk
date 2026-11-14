//! `close-check` (#2903, #2899 O4): may this Issue be closed now? The logic is
//! `moai_cli::close_check`.

use moai_cli::forge::transport::GhTransport;
use std::path::PathBuf;

/// Check whether an Issue may close
///
/// Answer whether a lane may close its own card now: the newest comment whose first line is
/// `CLOSE-TABLE #<n>` holds criteria tables (a column headed 판정, 상태, 결과, verdict or
/// status) whose every row is met (충족, 해당 없음, 끝, 병합, 지킴, 완료, closed, met, pass,
/// done, n/a), and no open pull request claims the Issue (`Moai-Claim: <n>#…` at a body line
/// start). Verdicts: close, keep_open, no_table, already_closed, uncertain. One GraphQL
/// request; writes nothing.
#[derive(clap::Args)]
pub struct Args {
    /// The Issue number.
    issue: u64,
    /// `OWNER/REPO`.
    #[arg(long)]
    repo: String,
    /// Decide from a saved snapshot (JSON) instead of reading GitHub.
    #[arg(long)]
    snapshot: Option<PathBuf>,
}

pub const OPERATION: &str = "close.check";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let snap = match &args.snapshot {
        Some(file) => {
            let bytes = std::fs::read(file).map_err(|e| format!("snapshot unreadable: {e}"))?;
            serde_json::from_slice::<moai_cli::close_check::Snapshot>(&bytes)
                .map_err(|e| format!("snapshot invalid: {e}"))?
        }
        None => {
            let (owner, repo) = args
                .repo
                .split_once('/')
                .ok_or("--repo must be OWNER/REPO")?;
            moai_cli::close_check::collect(owner, repo, args.issue, &GhTransport::from_env()?)?
        }
    };
    if snap.number != args.issue {
        return Err(format!(
            "snapshot is Issue #{}, not #{}",
            snap.number, args.issue
        ));
    }
    Ok(moai_cli::close_check::decide(&snap))
}
