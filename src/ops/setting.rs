//! `setting` (#2871 follow-up S1): read the user's CLI defaults. The logic is
//! `moai_cli::setting`.

/// Show the user's CLI defaults
///
/// `setting list` shows every user setting this command knows with its value and where it comes
/// from (`file` or `default`); `setting get KEY` shows one. The settings live in the user-only
/// file `config/launcher.yaml` under the HOME state folder: `launcher.role_access`,
/// `launcher.advisor` and `install.hosts` (which hosts the user-level install covers: claude,
/// codex, omp). A value outside its set is flagged without being quoted. Reads only; the
/// project's own settings are managed by `/moai project`, not here.
#[derive(clap::Args)]
pub struct Args {
    /// `list`, or `get` followed by a KEY.
    action: String,
    /// The setting to show with `get`.
    key: Option<String>,
}

pub const OPERATION: &str = "setting.read";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let home = moai_cli::key_vault::home()?;
    match (args.action.as_str(), args.key.as_deref()) {
        ("list", None) => moai_cli::setting::list(&home),
        ("get", Some(key)) => moai_cli::setting::get(&home, key),
        ("get", None) => Err("setting get needs a KEY".into()),
        ("list", Some(_)) => Err("setting list takes no KEY".into()),
        _ => Err("setting action must be list or get".into()),
    }
}
