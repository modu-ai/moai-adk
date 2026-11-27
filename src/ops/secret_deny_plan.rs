//! `secret-deny-plan` (SPEC-20261122-secret-file-deny-rules): the deny rules a project could
//! carry for its secret files. Reads only; the logic is `moai_cli::secret_deny`.

use std::path::PathBuf;

/// Preview Read deny rules for the project's key files
///
/// List the `Read` deny rules that could keep Claude Code away from the key files found in a
/// project, and the entries of its `.claudeignore` as rules, each with the files it matches and
/// the settings file it would go into. Files are matched by name and never opened. With `--pick`
/// the answer also holds the `permissions.deny` list the picked rules would make and the digest
/// `secret-deny-apply` needs. Writes nothing.
#[derive(clap::Args)]
pub struct Args {
    /// The project folder (default: the current folder).
    #[arg(default_value = ".")]
    project: PathBuf,
    /// A rule of the preview to add; repeat the flag for each rule.
    #[arg(long = "pick", value_name = "RULE")]
    picks: Vec<String>,
    /// Also move `.claudeignore` into the settings backup folder; needs at least one `--pick`.
    #[arg(long)]
    retire_claudeignore: bool,
}

pub const OPERATION: &str = "secret_deny.plan";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    if args.picks.is_empty() && !args.retire_claudeignore {
        moai_cli::secret_deny::plan(&args.project)
    } else {
        moai_cli::secret_deny::plan_picks(&args.project, &args.picks, args.retire_claudeignore)
    }
}
