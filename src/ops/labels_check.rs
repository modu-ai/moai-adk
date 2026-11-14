//! `labels-check` (#2740): read-only report of open Issues that break the label axis rule.

use moai_cli::forge::transport::GhTransport;
use moai_cli::labels_check;
use std::path::PathBuf;

/// Report Issues whose labels break the rule
///
/// Report the open Issues whose labels break the axis rule of the project's
/// `.github/moai/labels.yml` (exactly one `type:*`, exactly one `priority:*`, one or two
/// `area:*`, no `not_used` or unlisted label). Reads GitHub with GET only and never adds or
/// removes a label.
#[derive(clap::Args)]
pub struct Args {
    /// Project root holding `.github/moai/labels.yml`.
    project: PathBuf,
    /// `OWNER/REPO`.
    #[arg(long)]
    repo: String,
}

pub const OPERATION: &str = "labels.check";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    // Static progress lines on a stderr terminal (#2806); silent on a pipe.
    moai_cli::progress::install_stderr();
    let (owner, repo) = args
        .repo
        .split_once('/')
        .ok_or("--repo must be OWNER/REPO")?;
    let file = args.project.join(".github/moai/labels.yml");
    let text = std::fs::read_to_string(&file)
        .map_err(|_| format!("labels file unavailable: {}", file.display()))?;
    let axes = labels_check::parse_axes(&text)?;
    labels_check::check(&GhTransport::from_env()?, owner, repo, &axes)
}
