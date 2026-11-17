//! `trailer-check` (SPEC-20261117-session-trailer, SDD-6): the identity trailers of a range.

use std::path::PathBuf;

/// Report commits with missing or malformed trailers
///
/// Read the commits of RANGE (merges skipped, oldest first) and report each missing or
/// malformed `Moai-Spec:` / `Moai-Session:` trailer (user decision U-B 2026-11-17). A missing
/// `Moai-Spec` fails only on a `spec/` branch; a malformed value always fails; a missing
/// `Moai-Session` is reported and never fails (the audit gate then reads the authors as
/// unknown). `ok` is true when nothing fails; otherwise the result is printed and the command
/// exits 1. A value holding the log's field separator is `malformed_trailers`. Reads only.
#[derive(clap::Args)]
pub struct Args {
    /// Revision range, for example `<base>..HEAD`.
    range: String,
    /// Repository to read.
    #[arg(long, default_value = ".")]
    dir: PathBuf,
}

pub const OPERATION: &str = "trailers.check";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::trailers::check(&args.dir, &args.range)
}
