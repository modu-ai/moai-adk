//! `secret-deny-plan` (SPEC-20261122-secret-file-deny-rules): the deny rules a project could
//! carry for its secret files. Reads only; the logic is `moai_cli::secret_deny`.

use std::path::PathBuf;

/// Preview Read deny rules for the project's key files
///
/// List the `Read` deny rules that could keep Claude Code away from the key files found in a
/// project, and the entries of its `.claudeignore` as rules (SPEC-20261122-secret-file-deny-rules),
/// each with the files it matches and the settings file it would go into. Files are matched by
/// name and never opened. Writes nothing.
#[derive(clap::Args)]
pub struct Args {
    /// The project folder (default: the current folder).
    #[arg(default_value = ".")]
    project: PathBuf,
}

pub const OPERATION: &str = "secret_deny.plan";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::secret_deny::plan(&args.project)
}
