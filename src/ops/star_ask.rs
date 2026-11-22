//! `star-ask` (SPEC-20261122-star-and-reports): the polite star line.
//! B1 of the audit: a dismissed home makes no network call — the dismissal
//! is checked before the transport is built. The line comes in the stream's
//! language (B4: English and Korean). `--dismiss` records the marker in the
//! current-generation home folder and reports the truth when nothing was
//! written (B2, B3). The one read-only GET (the audit's clean finding) goes
//! to `api.github.com/user/starred/modu-ai/moai-adk` through the user's gh
//! login, only when this command runs explicitly.

/// The polite star ask
#[derive(clap::Args)]
pub struct Args {
    /// Record the dismissal: the user starred, or asked not to be asked
    #[arg(long)]
    dismiss: bool,
}

pub const OPERATION: &str = "star.ask";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    // N2: a HOME that is not an absolute path is refused.
    let raw_home = std::env::var("HOME").map_err(|e| format!("HOME is unreadable: {e}"))?;
    let home = std::path::PathBuf::from(&raw_home);
    if !home.is_absolute() {
        return Err(format!(
            "star-ask refused: HOME is not an absolute path: {raw_home}"
        ));
    }
    if args.dismiss {
        // B2, B3: the write outcome is the truth — written, the retired
        // folder, or a refused write with its reason.
        let answer = match moai_cli::star_ask::dismiss_star_ask(&home) {
            moai_cli::star_ask::WriteOutcome::Written => serde_json::json!(true),
            moai_cli::star_ask::WriteOutcome::RetiredFolder => serde_json::json!(false),
            moai_cli::star_ask::WriteOutcome::WriteRefused(reason) => {
                serde_json::json!({ "written": false, "reason": reason })
            }
        };
        return Ok(serde_json::json!({
            "schema_version": 1,
            "dismissed": answer,
        }));
    }
    // B1: a dismissed home makes no network call at all.
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
    let lang = moai_cli::ui_text::Lang::En; // the JSON envelope is English (D3)
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
