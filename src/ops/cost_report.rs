//! `cost-report` (D5 #2191): sums the cost items of run receipts; writes nothing. Moved out of
//! `src/main.rs` unchanged (#2834 R2b).

/// Sum the cost items of run receipts
///
/// Aggregate cost items across run/verify receipts (D5 #2191): typed request
/// (receipt array) on stdin; breakdown envelope on stdout; writes nothing.
/// External backend calls are structurally zero; tokens are absent, never
/// estimated (REQ-D5-01).
#[derive(clap::Args)]
pub struct Args {}

pub const OPERATION: &str = "cost.report";

pub fn run(_args: Args) -> Result<serde_json::Value, String> {
    // cost_report checks protocol_version itself (one check, not two).
    moai_cli::run::cost_report(super::super::input::<moai_cli::run::CostRequest>()?)
}
