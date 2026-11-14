//! `lane-resume-check` (#2904 O5): read a card's resume note back from a comment body. The
//! logic is `moai_cli::lane_loop`.

use std::path::PathBuf;

/// Read a card's resume note from a comment
///
/// Read the resume note in a saved comment body (`--file`): its one `moai-resume:` line at a
/// line start and the `next:` and `wait:` lines after it. Answers `ok: true` with the fields,
/// or the problems found (no note, two notes, a missing or unknown field, uncommitted work, a
/// short head). The first thing a lane reads after a compact or a clear. Writes nothing.
#[derive(clap::Args)]
pub struct Args {
    /// The comment body, saved to a file (over 1 MiB is refused, never read in part).
    #[arg(long)]
    file: PathBuf,
}

pub const OPERATION: &str = "lane.resume_check";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    // A body over 1 MiB is refused whole, never read in part (AUDIT #2917 I1).
    let bytes = moai_cli::lane_loop::read_bounded(&args.file).map_err(|e| format!("file {e}"))?;
    let body = String::from_utf8(bytes).map_err(|_| "file is not UTF-8".to_string())?;
    Ok(match moai_cli::lane_loop::parse(&body) {
        Ok(n) => serde_json::json!({
            "ok": true,
            "note": {
                "lane": n.lane, "issue": n.issue, "claim": n.claim,
                "lease_until": n.lease_until, "branch": n.branch, "head": n.head,
                "uncommitted": n.uncommitted, "next": n.next,
                "wait": n.wait.map(|(c, w)| serde_json::json!({"condition": c, "clears": w})),
            },
            "writes": false,
        }),
        Err(problems) => serde_json::json!({"ok": false, "problems": problems, "writes": false}),
    })
}
