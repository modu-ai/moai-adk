//! `star-ask` (SPEC-20261122-star-and-reports): the polite star line of
//! install and update. One read-only GET through the user's gh login checks
//! the star state; the line shows once when unknown or not starred, never
//! when starred or dismissed; `--dismiss` records the dismissal in the home
//! moai state. Nothing is written besides the dismissal marker and no write
//! call exists on the check path.

/// The polite star ask
#[derive(clap::Args)]
pub struct Args {
    /// Record the dismissal: the user starred, or asked not to be asked
    #[arg(long)]
    dismiss: bool,
}

pub const OPERATION: &str = "star.ask";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    // The raw HOME path: the star_ask functions resolve the state folder
    // themselves (home_root::path), so passing the resolved folder here
    // would double-resolve.
    let home = std::path::PathBuf::from(
        std::env::var("HOME").map_err(|e| format!("HOME is unreadable: {e}"))?,
    );
    if args.dismiss {
        moai_cli::star_ask::dismiss_star_ask(&home);
        return Ok(serde_json::json!({
            "schema_version": 1,
            "dismissed": true,
        }));
    }
    let transport =
        moai_cli::forge::transport::GhTransport::from_env().map_err(|e| e.to_string())?;
    let state = moai_cli::star_ask::check_star(&transport);
    let line = moai_cli::star_ask::star_line(state, &home);
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
