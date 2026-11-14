//! `parity-check` (D6 #2192): semantic parity of a template and its generated files; read-only.
//! Moved out of `src/main.rs` unchanged (#2834 R1).

use std::path::PathBuf;

/// Compare a template and its generated files
///
/// Semantic parity of TEMPLATE vs GENERATED (optional third DOC set, D6 #2192):
/// per-key classification with the 00-A vocabulary (missing / extra /
/// value_mismatch / identical / doc_only). Read-only.
#[derive(clap::Args)]
pub struct Args {
    template: PathBuf,
    generated: PathBuf,
    #[arg(long)]
    doc: Option<PathBuf>,
}

pub const OPERATION: &str = "parity.check";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::parity::parity_check(&args.template, &args.generated, args.doc.as_deref())
}
