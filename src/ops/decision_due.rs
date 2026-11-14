//! `decision-due` (#2619): what each unanswered decision's deadline means at a time.

use moai_cli::decision_file::{self, Due};
use std::path::PathBuf;

/// Sort the unanswered decisions of a bundle
///
/// Sort the unanswered decisions of one bundle at `--now` (default: the clock): still open,
/// default applies (reversible, category `none`, the user's, past its deadline), or on hold
/// (one-way, `guarded` or a HOLD category, past its deadline). Reads only; applies nothing.
#[derive(clap::Args)]
pub struct Args {
    file: PathBuf,
    /// `YYYY-MM-DDTHH:MM:SSZ`; the system clock when absent.
    #[arg(long)]
    now: Option<String>,
}

pub const OPERATION: &str = "decision.due";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let text = std::fs::read_to_string(&args.file)
        .map_err(|e| format!("decision file unavailable: {}: {e}", args.file.display()))?;
    let (rows, problems) = decision_file::parse(&text);
    if !problems.is_empty() {
        return Err(format!(
            "decision file refused: {} problem(s); run decision-file-check",
            problems.len()
        ));
    }
    let now_text = args.now.unwrap_or_else(moai_cli::audit::utc_now);
    let now = decision_file::deadline_secs(&now_text)
        .ok_or_else(|| format!("invalid --now `{now_text}`: expected YYYY-MM-DDTHH:MM:SSZ"))?;
    let (mut open, mut default_applies, mut hold, mut answered) =
        (Vec::new(), Vec::new(), Vec::new(), Vec::new());
    for row in &rows {
        match decision_file::due(row, now) {
            None => answered.push(serde_json::json!({ "id": row.id, "answer": row.answer })),
            Some(Due::Open) => {
                open.push(serde_json::json!({ "id": row.id, "deadline": row.deadline }))
            }
            Some(Due::DefaultApplies) => {
                default_applies.push(serde_json::json!({ "id": row.id, "default": row.default }))
            }
            Some(Due::Hold) => {
                hold.push(serde_json::json!({ "id": row.id, "category": row.category }))
            }
        }
    }
    Ok(serde_json::json!({
        "now": now_text,
        "open": open,
        "default_applies": default_applies,
        "hold": hold,
        "answered": answered,
    }))
}
