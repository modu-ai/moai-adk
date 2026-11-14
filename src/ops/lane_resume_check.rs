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
    /// The comment body, saved to a file (at most 1 MiB is read).
    #[arg(long)]
    file: PathBuf,
}

pub const OPERATION: &str = "lane.resume_check";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    use std::io::Read as _;
    let meta = std::fs::metadata(&args.file).map_err(|e| format!("file unreadable: {e}"))?;
    if !meta.is_file() {
        return Err("file must be a regular file".into());
    }
    let mut body = String::new();
    std::fs::File::open(&args.file)
        .and_then(|f| f.take(1 << 20).read_to_string(&mut body))
        .map_err(|e| format!("file unreadable: {e}"))?;
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
