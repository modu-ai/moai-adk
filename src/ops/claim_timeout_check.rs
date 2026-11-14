//! `claim-timeout-check` (#2906 S1): whether a lane may close another lane's claim with a
//! `by=timeout` handoff line. Reads GitHub with GET only; the logic is
//! `moai_cli::claim_timeout`.

use moai_cli::forge::policy::{parse_utc, SystemClock};
use moai_cli::forge::transport::GhTransport;

/// Check the three timeout takeover conditions of a claim
///
/// Decision D2 (#1975): a lane may post `moai-claim-handoff: comment=<holder> to=<lane>
/// by=timeout` only when the holder's lease has expired (`now >= lease_until`), the lane's
/// request naming the holder's claim id has gone 1800 s without an answer, and the holder has
/// shown no activity for 7200 s (lease renewal, a comment of its lane anywhere in the
/// repository in the last two hours, a commit with its claim token on the claim branch).
/// Prints `met` with each condition's evidence, and the line and a comment body when all
/// three hold. A read that fails is `uncertain`. Reads only; posts nothing.
#[derive(clap::Args)]
pub struct Args {
    /// The Issue number the claim is on.
    #[arg(long)]
    issue: i64,
    /// The holder's claim comment id.
    #[arg(long)]
    claim: i64,
    /// The comment id of this lane's request naming that claim.
    #[arg(long)]
    request: i64,
    /// This lane's name, the `to=` of the handoff line.
    #[arg(long)]
    lane: String,
    /// `OWNER/REPO` (default: the project's `origin` remote).
    #[arg(long)]
    repo: Option<String>,
    /// `YYYY-MM-DDTHH:MM:SSZ` to judge against; the system clock when absent.
    #[arg(long)]
    now: Option<String>,
}

pub const OPERATION: &str = "claim.timeout_check";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let now_text = args.now.unwrap_or_else(moai_cli::audit::utc_now);
    let now = parse_utc(&now_text)
        .ok_or_else(|| format!("invalid --now `{now_text}`: expected YYYY-MM-DDTHH:MM:SSZ"))?;
    if args.issue <= 0 || args.claim <= 0 || args.request <= 0 || args.lane.is_empty() {
        return Err("invalid claim-timeout-check: --issue, --claim, --request and --lane are required".into());
    }
    // The same OWNER/REPO reading as `claim-verify` (flag first, then `origin`).
    let request = moai_cli::claim_verify::request_from_flags(
        std::path::Path::new("."),
        args.issue,
        args.claim,
        &args.lane,
        args.repo.as_deref(),
        &[],
    )?;
    let target = moai_cli::claim_timeout::Target {
        owner: &request.owner,
        repo: &request.repo,
        issue: args.issue,
        claim: args.claim,
        request: args.request,
        me: &args.lane,
    };
    Ok(moai_cli::claim_timeout::check(
        &target,
        &GhTransport::from_env()?,
        &SystemClock,
        now,
    ))
}
