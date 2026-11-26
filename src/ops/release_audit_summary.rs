//! `release-audit-summary` (SPEC-20261125-release-audit-summary): gather the verdicts of a release
//! range from the ledger into one pull request comment. The logic is
//! `moai_cli::release_audit_summary`.

use moai_cli::release_audit_summary as summary;
use std::path::PathBuf;

/// Gather a release range's verdicts into one PR comment
///
/// Reads the local git history and the ledger and prints the one comment the `audit-gate` job
/// reads: one line for each first-parent commit of `base..head`, oldest first, with the verdicts
/// of the cards it closed (`PASS`), the commits that a later landing absorbed (`ABS`), and every
/// commit with no passing verdict by name (`NONE`, `FAIL`). The default run only prints the
/// comment, its `plan_digest`, the counts and the missing landings; it starts no `gh` and writes
/// nothing. `--post` sends the comment once, and only with `--pr`, `--repo`, the `plan_digest`
/// of the same run and a `--permission-reference`; a summary with a line that has no passing
/// verdict also needs `--allow-incomplete`. A comment over 60000 bytes is refused, never cut.
#[derive(clap::Args)]
pub struct Args {
    /// The project folder (its profile holds the project id).
    project: PathBuf,
    /// The merge base of the release pull request: 40 lowercase hex, on the first-parent line of
    /// `--head`.
    #[arg(long)]
    base: String,
    /// The head of the release pull request: 40 lowercase hex.
    #[arg(long)]
    head: String,
    /// The HOME holding the ledger (default `$HOME`). Tests pass a temporary folder.
    #[arg(long)]
    home: Option<PathBuf>,
    /// Post the comment instead of only printing it.
    #[arg(long)]
    post: bool,
    /// The pull request number to post to (with `--post`).
    #[arg(long)]
    pr: Option<u64>,
    /// `OWNER/REPO` to post to (with `--post`).
    #[arg(long)]
    repo: Option<String>,
    /// The `plan_digest` the same run printed (with `--post`).
    #[arg(long)]
    plan_digest: Option<String>,
    /// The caller's reference for the user's approval (with `--post`).
    #[arg(long)]
    permission_reference: Option<String>,
    /// Post a summary although some line has no passing verdict (with `--post`).
    #[arg(long)]
    allow_incomplete: bool,
    /// Fail when some line has no passing verdict.
    #[arg(long)]
    require_complete: bool,
}

pub const OPERATION: &str = "release.audit.summary";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let home = moai_cli::ledger::command::resolve_home(args.home)?;
    let request = summary::Request {
        project: &args.project,
        home: &home,
        base: &args.base,
        head: &args.head,
    };
    if !args.post {
        let plan = summary::plan(&request)?;
        if args.require_complete && plan["complete"] != true {
            return Err(summary::incomplete_error(&plan));
        }
        return Ok(plan);
    }
    // Every refusal comes before `gh` is started.
    let approved = summary::approve(
        &request,
        &summary::PostArgs {
            pr: args.pr,
            repo: args.repo.as_deref(),
            plan_digest: args.plan_digest.as_deref(),
            permission_reference: args.permission_reference.as_deref(),
            allow_incomplete: args.allow_incomplete,
            require_complete: args.require_complete,
        },
    )?;
    let transport = moai_cli::forge::transport::GhTransport::from_env()?;
    let clock = moai_cli::forge::policy::SystemClock;
    let connector = moai_cli::forge::connector::GithubConnector::new(
        &transport,
        &clock,
        moai_cli::forge::Config::default(),
    );
    summary::apply(&approved, &connector)
}
