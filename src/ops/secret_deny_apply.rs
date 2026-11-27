//! `secret-deny-apply` (SPEC-20261122-secret-file-deny-rules): adds the picked `Read` deny rules
//! to the project's root settings file. The logic is `moai_cli::secret_deny::apply`.

use std::path::PathBuf;

/// Add the picked Read deny rules to the project settings
///
/// Add only the picked rules to `permissions.deny` of the project's `.claude/settings.json` after
/// the user confirmed `secret-deny-plan`: every other key and permission entry stays, the
/// original file is kept in the settings backup folder, and every key-file rule ends behind the
/// last `!` entry. Nothing is written while git ignores the settings file or when the digest or
/// the picks differ from the preview. With `--retire-claudeignore`, `.claudeignore` is moved into
/// the backup folder as the last step; it is never deleted.
#[derive(clap::Args)]
pub struct Args {
    /// The project folder (default: the current folder).
    #[arg(default_value = ".")]
    project: PathBuf,
    /// The `apply.plan_digest` of the confirmed `secret-deny-plan` run with the same flags.
    #[arg(long)]
    plan_digest: String,
    /// The caller's reference for the user's approval.
    #[arg(long)]
    permission_reference: String,
    /// A rule of the preview to add; repeat the flag for each rule, in the preview's order.
    #[arg(long = "pick", value_name = "RULE")]
    picks: Vec<String>,
    /// Also move `.claudeignore` into the settings backup folder; needs at least one `--pick`.
    #[arg(long)]
    retire_claudeignore: bool,
}

pub const OPERATION: &str = "secret_deny.apply";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    moai_cli::secret_deny::apply(
        &args.project,
        &args.plan_digest,
        &args.permission_reference,
        &args.picks,
        args.retire_claudeignore,
    )
}
