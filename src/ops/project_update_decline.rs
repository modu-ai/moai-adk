//! `project-update-decline`: record that the user declined a choice step of the project update, or
//! ask again. The logic is `moai_cli::project_update::declined`.

use std::path::PathBuf;

/// Decline a choice step of the project update, or ask again
///
/// Record that the user declined one choice step of `project-update-plan` (`agents-picks`,
/// `agents-local-merge` or `claude-settings-merge`) for the source it offered: the step name and
/// the `source_sha256` the preview printed go into `.moai/manifests/update-declined.json`. While
/// the step's source keeps that sha256, the preview reads the step as `skip`; a source with other
/// bytes asks again by itself. Nothing but that one file is written: no project file is changed or
/// deleted. A sha256 that is not the step's source now, a step that cannot be declined, and a
/// link for the state folder or the record are refused with nothing written. With `--ask-again`
/// the step's record is removed and the other records stay.
#[derive(clap::Args)]
pub struct Args {
    /// The step: `agents-picks`, `agents-local-merge` or `claude-settings-merge`.
    #[arg(long)]
    step: String,
    /// The `source_sha256` the update preview printed for the step (not with `--ask-again`).
    #[arg(long)]
    source_sha256: Option<String>,
    /// Remove the step's record, so the next preview asks again.
    #[arg(long)]
    ask_again: bool,
    /// The caller's reference for the user's approval.
    #[arg(long)]
    permission_reference: String,
    /// The project (default: the current folder).
    #[arg(default_value = ".")]
    project: PathBuf,
}

pub const OPERATION: &str = "project_update.decline";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    use moai_cli::project_update::declined;
    match (args.ask_again, &args.source_sha256) {
        (true, None) => declined::ask_again(&args.project, &args.step, &args.permission_reference),
        (false, Some(sha)) => {
            declined::decline(&args.project, &args.step, sha, &args.permission_reference)
        }
        (true, Some(_)) => Err(
            "invalid arguments: --ask-again does not go with --source-sha256; nothing was written"
                .into(),
        ),
        (false, None) => {
            Err("invalid arguments: --source-sha256 is required; nothing was written".into())
        }
    }
}
