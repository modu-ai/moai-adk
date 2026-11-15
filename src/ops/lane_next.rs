//! `lane-next` (#2831 C1): the one thing a lane should do next, read fresh from GitHub each
//! run. The logic is `moai_cli::lane_next`.

use moai_cli::forge::policy::{Clock as _, SystemClock};
use moai_cli::forge::transport::GhTransport;
use std::path::PathBuf;

/// Find the lane's next task
///
/// Find the one task this lane should do next and why, in the user's P0–P4 order: a red tip,
/// the lane's own pull request (audit FAIL, conflict, failed CI), an audit it reserved (not
/// judged, head moved, re-audit asked — with the head now), a pull request with no auditor, a
/// decision request the leader answered, its claimed card, an unclaimed Issue. Reads GitHub
/// (at most 30 requests a run) and the local clone only. `--wait` reads again every 432
/// seconds for the default ten lanes on one token (within half of GitHub's hourly limit;
/// `--shared-lanes` changes the count) until a task appears. A watching lane stops by itself:
/// the one thing it writes is a count kept per lane and repository in the clone's state
/// folder (`<git common dir>/.moai-cli/lane-idle/`, result `writes: true`; nothing with
/// `--max-empty 0`; a `--snapshot` read counts and writes only when `--max-empty` is given).
/// Each read with no task, nothing to wait for and a full answer from GitHub adds 1; at
/// `--max-empty` (6 on a GitHub read) the result says `status: idle_limit` and carries the
/// `LANE-IDLE` line to print once, and a `--wait` ends at that read. A read with a task,
/// something to wait for or an open pull request of the lane's own (a draft too:
/// `own_open_prs`) sets the count back to 0; a partial read (`partial`, or an unknown
/// `end_condition.met`) is not counted. Two `partial` notes are not a failure and do not
/// block an empty read: no tip base known (nothing is open) and `#<n>: only the newest 100
/// comments read` of a busy Issue.
#[derive(clap::Args)]
pub struct Args {
    /// The lane name as it appears in claims (`session=<lane>`) and audit reservations.
    #[arg(long)]
    lane: String,
    /// `OWNER/REPO`.
    #[arg(long)]
    repo: String,
    /// The integration branch whose CI is P0 (default: the base most open pull requests
    /// target).
    #[arg(long)]
    tip: Option<String>,
    /// Seconds to keep reading until a task appears (0 = read once; at most 3600).
    #[arg(long, default_value_t = 0)]
    wait: u64,
    /// Lanes polling with the same GitHub token: the `--wait` interval is
    /// ⌈3600 × 30 × lanes / 2500⌉ seconds (432 for 10; 1 to 50).
    #[arg(long, default_value_t = moai_cli::lane_next::SHARED_LANES)]
    shared_lanes: u64,
    /// The local clone the tip judgment walks.
    #[arg(long, default_value = ".")]
    dir: PathBuf,
    /// Decide from a saved snapshot (JSON) instead of reading GitHub.
    #[arg(long)]
    snapshot: Option<PathBuf>,
    /// Add `explain`: why each pull request and Issue was passed over.
    #[arg(long)]
    explain: bool,
    /// Empty reads in a row after which the result says `idle_limit` (1 to 1000; 0 = do not
    /// count). Default 6 on a GitHub read; a `--snapshot` read counts only when this is given.
    #[arg(long)]
    max_empty: Option<u64>,
}

pub const OPERATION: &str = "lane.next";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let max_empty = moai_cli::lane_idle::resolve_max(args.max_empty, args.snapshot.is_some())?;
    let anchor = (max_empty > 0)
        .then(|| {
            moai_cli::identity::identity(&args.dir)
                .ok()
                .filter(|id| id["identity_status"].as_str() == Some("ok"))
                .and_then(|id| id["clone_anchor"].as_str().map(PathBuf::from))
        })
        .flatten();
    // The count saved so far, so a `--wait` call stops at the Nth empty read in a row.
    let budget =
        moai_cli::lane_idle::budget(anchor.as_deref(), &args.lane, &args.repo, max_empty);
    let mut v = match &args.snapshot {
        Some(file) => {
            let bytes = std::fs::read(file).map_err(|e| format!("snapshot unreadable: {e}"))?;
            let snap: moai_cli::lane_next::Snapshot =
                serde_json::from_slice(&bytes).map_err(|e| format!("snapshot invalid: {e}"))?;
            moai_cli::lane_next::decide(&args.lane, &snap)
        }
        None => {
            let (owner, repo) = args
                .repo
                .split_once('/')
                .ok_or("--repo must be OWNER/REPO")?;
            let target = moai_cli::lane_next::Target {
                owner,
                repo,
                lane: &args.lane,
                tip_branch: args.tip.as_deref(),
                dir: &args.dir,
            };
            moai_cli::lane_next::run_counting(
                &target,
                &GhTransport::from_env()?,
                &SystemClock,
                args.wait,
                args.shared_lanes,
                budget,
            )?
        }
    };
    if max_empty > 0 {
        let now = moai_cli::forge::policy::format_utc(SystemClock.now());
        moai_cli::lane_idle::apply(
            &mut v,
            anchor.as_deref(),
            &args.lane,
            &args.repo,
            max_empty,
            &now,
        );
    }
    if args.explain {
        v["explain"] = serde_json::json!(moai_cli::lane_next::explain_lines(&v));
    }
    // A person at a terminal gets one line on stderr (#2831 D2); stdout stays the envelope.
    use std::io::IsTerminal as _;
    if std::io::stderr().is_terminal() {
        eprintln!("{}", moai_cli::lane_next::summary_line(&v));
    }
    Ok(v)
}
