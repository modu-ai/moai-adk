//! `ledger-register-plan` (#3068 P1): what registering the project's local ledger would do.
//! Reads only; the logic is `moai_cli::ledger::command`.

use std::path::PathBuf;

/// Preview registering the project's ledger
///
/// Preview creating the project's local SQLite ledger at `<HOME>/.moai/projects/<project_id>/`
/// (#3068): the project id comes from the project profile, and the preview names the folder and
/// file a registration would make, whether a ledger is already there, and a `plan_digest` for
/// `ledger-register-apply`. A ledger that is already there is checked read-only and never
/// replaced. Writes nothing.
#[derive(clap::Args)]
pub struct Args {
    /// The project folder (its profile holds the project id).
    project: PathBuf,
    /// The HOME to use (default `$HOME`). Tests pass a temporary folder.
    #[arg(long)]
    home: Option<PathBuf>,
}

pub const OPERATION: &str = "ledger.register.plan";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let home = moai_cli::ledger::command::resolve_home(args.home)?;
    let mut plan = moai_cli::ledger::command::register_plan(&args.project, &home)?;
    // One answer with `home-install-plan` for the same HOME (card C-50): an unknown `~/.moai`
    // with no finished Go retire on record is not ready for the ledger either.
    if plan["action"] == "register" {
        if let Some(why) = moai_cli::home_install::registration_blocked(&home) {
            plan["action"] = serde_json::json!("blocked");
            plan["problem"] = serde_json::json!("home_not_ready");
            plan["message"] = serde_json::json!(why);
            plan["would_create"] = serde_json::json!([]);
            plan["would_tighten"] = serde_json::json!([]);
            // The library's digest was made for `register`; a blocked preview gets its own, so
            // it can never authorize the registration once the HOME turns ready.
            plan["plan_digest"] = serde_json::json!(moai_cli::digest(
                format!("blocked:{}:{}", plan["plan_digest"], plan["message"]).as_bytes()
            ));
        }
    }
    Ok(plan)
}
