//! `flake-classify` (#2666 F1): which tests of one CI log failed and later passed.

use std::path::PathBuf;

/// Classify flaky tests from two CI logs
///
/// Classify flaky tests (#2666): given the log of a failed CI run and the log of a later run of
/// the same head (`gh run view <id> --log` saved to a file), list the tests that failed first
/// and passed later (flaky), failed in both (still failing), or are absent from the later log.
/// Reads the two files only; writes nothing.
#[derive(clap::Args)]
pub struct Args {
    /// Log of the run where tests failed.
    #[arg(long)]
    failed: PathBuf,
    /// Log of a later run of the same head.
    #[arg(long)]
    later: PathBuf,
    /// Only this job (exact name, for example `ci / macos-arm64`).
    #[arg(long)]
    job: Option<String>,
}

pub const OPERATION: &str = "flake.classify";

const LIMIT: u64 = 64 * 1024 * 1024;

fn read(path: &PathBuf) -> Result<String, String> {
    let meta = std::fs::metadata(path).map_err(|_| format!("log unavailable: {}", path.display()))?;
    if !meta.is_file() || meta.len() > LIMIT {
        return Err(format!("log must be a regular file up to 64 MiB: {}", path.display()));
    }
    let bytes = std::fs::read(path).map_err(|_| format!("log unreadable: {}", path.display()))?;
    Ok(String::from_utf8_lossy(&bytes).into_owned())
}

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let c = moai_cli::flake::classify(&read(&args.failed)?, &read(&args.later)?, args.job.as_deref());
    serde_json::to_value(c).map_err(|_| "encode failed".into())
}
