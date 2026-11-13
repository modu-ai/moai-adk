//! `decision-file-check` (#2619): the shape check of one decision bundle; reads, writes nothing.

use std::path::PathBuf;

/// Check one decision bundle under `docs/decisions/` (#2619): the table, every field, and
/// that a one-way, `guarded` or HOLD-category decision is the user's. Reads only.
#[derive(clap::Args)]
pub struct Args {
    file: PathBuf,
}

pub const OPERATION: &str = "decision.file_check";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let text = std::fs::read_to_string(&args.file)
        .map_err(|e| format!("decision file unavailable: {}: {e}", args.file.display()))?;
    let (rows, problems) = moai_cli::decision_file::parse(&text);
    if !problems.is_empty() {
        let list: Vec<String> = problems
            .iter()
            .map(|p| format!("line {} {}: {}", p.line, p.id, p.reason))
            .collect();
        return Err(format!(
            "decision file refused: {} problem(s): {}",
            problems.len(),
            list.join("; ")
        ));
    }
    Ok(serde_json::json!({ "rows": rows }))
}
