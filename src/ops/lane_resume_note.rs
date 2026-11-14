//! `lane-resume-note` (#2904 O5): the resume note a card keeps on its Issue. The logic is
//! `moai_cli::lane_loop`.

/// Write a card's resume note
///
/// Write the resume note a card keeps on its Issue, so a lane continues after a compact or a
/// clear: the first line `moai-resume: lane= issue= claim= lease_until= branch= head=
/// uncommitted=0`, then one `next:` line per remaining step and an optional `wait: <condition>
/// clears=<leader|user|lane|author>`. Refuses a note with uncommitted work, a head that is not
/// the full 40-hex id, a claim id under seven digits or no next step. Prints the text in the
/// result (`text`); posting it is the lane's own comment. Writes nothing.
#[derive(clap::Args)]
pub struct Args {
    /// The lane name as it appears in claims.
    #[arg(long)]
    lane: String,
    /// The card's Issue number.
    #[arg(long)]
    issue: u64,
    /// The claim comment id.
    #[arg(long)]
    claim: u64,
    /// The claim's lease end, `YYYY-MM-DDTHH:MM:SSZ`.
    #[arg(long)]
    lease_until: String,
    /// The card's branch.
    #[arg(long)]
    branch: String,
    /// The branch head, the full 40-hex id (read it from `git rev-parse HEAD`).
    #[arg(long)]
    head: String,
    /// The uncommitted count read from `git status --porcelain` (must be 0).
    #[arg(long)]
    uncommitted: u64,
    /// A remaining step; give it once per step, in order.
    #[arg(long, required = true)]
    next: Vec<String>,
    /// What the card waits on.
    #[arg(long, requires = "clears")]
    wait: Option<String>,
    /// Who clears the wait: leader, user, lane or author.
    #[arg(long, requires = "wait")]
    clears: Option<String>,
}

pub const OPERATION: &str = "lane.resume_note";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let note = moai_cli::lane_loop::ResumeNote {
        lane: args.lane,
        issue: args.issue,
        claim: args.claim,
        lease_until: args.lease_until,
        branch: args.branch,
        head: args.head,
        uncommitted: args.uncommitted,
        next: args.next,
        wait: args.wait.zip(args.clears),
    };
    let text = moai_cli::lane_loop::render(&note).map_err(|p| p.join("; "))?;
    Ok(serde_json::json!({"text": text, "writes": false}))
}
