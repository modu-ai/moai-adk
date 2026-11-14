//! `flake-record` (#2666 F2): one `[flake] <test>` Issue per flaky test, a comment per run.

use moai_cli::forge::transport::GhTransport;
use std::path::PathBuf;

/// Record classified flaky tests on their Issues
///
/// Record flaky tests (#2666): for every flaky row of a `flake-classify` result, comment on the
/// open Issue that tracks the test (hidden marker on its first line), or create `[flake] <test>`
/// when none is open. A run and job already recorded writes nothing. Needs a permission
/// reference.
#[derive(clap::Args)]
pub struct Args {
    /// The `flake-classify` JSON output.
    #[arg(long)]
    classified: PathBuf,
    /// Id of the failed CI run the rows came from.
    #[arg(long)]
    run: u64,
    /// The operator's permission reference for these writes.
    #[arg(long)]
    permission_reference: String,
    /// `OWNER/REPO`.
    #[arg(long, default_value = "modu-ai/moai-adk")]
    repo: String,
}

pub const OPERATION: &str = "flake.record";

const LIMIT: u64 = 16 * 1024 * 1024;

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let (owner, repo) = args
        .repo
        .split_once('/')
        .ok_or("--repo must be OWNER/REPO")?;
    let path = &args.classified;
    let meta = std::fs::metadata(path)
        .map_err(|_| format!("classified file unavailable: {}", path.display()))?;
    if !meta.is_file() || meta.len() > LIMIT {
        return Err(format!(
            "classified file must be a regular file up to 16 MiB: {}",
            path.display()
        ));
    }
    let text = std::fs::read_to_string(path)
        .map_err(|_| format!("classified file unreadable: {}", path.display()))?;
    let classified: serde_json::Value =
        serde_json::from_str(&text).map_err(|e| format!("classified file is not JSON: {e}"))?;
    let transport = GhTransport::from_env()?;
    moai_cli::flake_record::record(
        owner,
        repo,
        args.run,
        &classified,
        &args.permission_reference,
        &transport,
    )
}
