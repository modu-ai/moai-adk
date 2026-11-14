//! `gate-schedule-apply` (#2906 S2): write a previewed scheduler entry. The logic is
//! `moai_cli::gate_schedule`.

use super::gate_schedule_plan::Entry;

/// Write a previewed gate-loop scheduler entry
///
/// Write a previewed scheduler entry for `gate-loop --once` (#2906 S2): the one plist or
/// crontab file, only when a fresh preview of the same inputs has the same digest, with a
/// receipt holding the previous bytes. Runs no `launchctl` or `crontab`; it prints the enable
/// and disable commands for the operator.
#[derive(clap::Args)]
pub struct Args {
    #[command(flatten)]
    entry: Entry,
    /// The preview's `plan_digest`.
    #[arg(long)]
    plan_digest: String,
    /// Caller-supplied permission reference for this write (recorded; not an authorization
    /// verifier).
    #[arg(long)]
    permission_reference: String,
}

pub const OPERATION: &str = "gate.schedule_apply";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::gate_schedule::apply(
        &args.entry.request()?,
        &args.plan_digest,
        &args.permission_reference,
    )
}
