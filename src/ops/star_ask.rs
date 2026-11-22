//! `star-ask` (SPEC-20261122-star-and-reports): the polite star line.
//! A dismissed home makes no network call — the dismissal is checked before
//! the transport is built. The line comes from the shared UI text table in
//! the stream's language, like every other screen. `--dismiss` records the
//! marker in the current-generation home folder and answers one shape on
//! every outcome — an object with `written` and `reason` (null when the
//! marker landed), exit 0 always, so a state write never blocks a run; the
//! retired folder and a refused write are named in `reason`. The one
//! read-only GET goes to `api.github.com/user/starred/modu-ai/moai-adk`
//! through the user's gh login, only when this command runs explicitly.

/// The polite star ask
#[derive(clap::Args)]
pub struct Args {
    /// Record the dismissal: the user starred, or asked not to be asked
    #[arg(long)]
    dismiss: bool,
}

pub const OPERATION: &str = "star.ask";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    // A HOME that is not an absolute path is refused.
    let raw_home = std::env::var("HOME").map_err(|e| format!("HOME is unreadable: {e}"))?;
    let home = std::path::PathBuf::from(&raw_home);
    if !home.is_absolute() {
        return Err(format!(
            "star-ask refused: HOME is not an absolute path: {raw_home}"
        ));
    }
    if args.dismiss {
        // One shape on every outcome (SPEC D-6): an object with `written`
        // and `reason` — null when the marker landed, the retired folder or
        // the refusal named otherwise. Exit 0 always: a state write never
        // blocks a run.
        let (written, reason) = match moai_cli::star_ask::dismiss_star_ask(&home) {
            moai_cli::star_ask::WriteOutcome::Written => (true, None),
            moai_cli::star_ask::WriteOutcome::RetiredFolder => (
                false,
                Some(
                    "the home answers the retired folder (.moai-cli): nothing new is ever created there"
                        .to_string(),
                ),
            ),
            moai_cli::star_ask::WriteOutcome::WriteRefused(reason) => (false, Some(reason)),
        };
        return Ok(serde_json::json!({
            "schema_version": 1,
            "dismissed": { "written": written, "reason": reason },
        }));
    }
    // A dismissed home makes no network call at all.
    if moai_cli::star_ask::star_ask_dismissed(&home) {
        return Ok(serde_json::json!({
            "schema_version": 1,
            "state": "dismissed",
            "line": serde_json::Value::Null,
            "writes": false,
        }));
    }
    let transport =
        moai_cli::forge::transport::GhTransport::from_env().map_err(|e| e.to_string())?;
    let state = moai_cli::star_ask::check_star(&transport);
    // The line follows the stream's language, like every other screen.
    let lang = moai_cli::ui::ui_lang(moai_cli::ui::Stream::Stdout);
    let line = moai_cli::star_ask::star_line(state, &home, lang);
    Ok(serde_json::json!({
        "schema_version": 1,
        "state": match state {
            moai_cli::star_ask::StarState::Starred => "starred",
            moai_cli::star_ask::StarState::NotStarred => "not_starred",
            moai_cli::star_ask::StarState::Unknown => "unknown",
        },
        "line": line,
        "writes": false,
    }))
}
