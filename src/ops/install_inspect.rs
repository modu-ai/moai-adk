//! `install-inspect` (04-A #1998): read-only installation inspection; on a terminal the
//! PATH-first note once on stderr (#2867). Moved out of `src/main.rs` unchanged (#2834 R3).

use std::path::PathBuf;

/// Show the installed MoAI binaries
///
/// Read-only installation inspection (04-A #1998): PATH generations, receipts under
/// `~/.moai-cli/installations/`, Go-generation path existence. Writes nothing.
#[derive(clap::Args)]
pub struct Args {
    /// Accepted for the REQ-04A-02 invocation form; output is always the JSON envelope.
    #[arg(long)]
    json: bool,
}

pub const OPERATION: &str = "install.inspect";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let Args { json: _ } = args;
    let home = std::env::var_os("HOME")
        .filter(|h| !h.is_empty())
        .map(PathBuf::from)
        .ok_or("home directory missing; install-inspect requires HOME")?;
    let path_env = std::env::var("PATH").unwrap_or_default();
    let env_dir = |k: &str| {
        std::env::var_os(k)
            .filter(|v| !v.is_empty())
            .map(PathBuf::from)
    };
    let (project_dir, gobin) = (env_dir("CLAUDE_PROJECT_DIR"), env_dir("GOBIN"));
    let places = moai_cli::install::CandidatePlaces {
        project_dir: project_dir.as_deref(),
        gobin: gobin.as_deref(),
    };
    let result = moai_cli::install::inspect(&home, &path_env, &places)?;
    // #2867: a person at a terminal sees the note once, in stderr's language (#2825);
    // the envelope carries the English note in every language.
    if result["path_first_moai"]["note"].is_string() && moai_cli::ui::stderr_is_terminal() {
        let lang = moai_cli::ui::ui_lang(moai_cli::ui::Stream::Stderr);
        let note = moai_cli::ui_text::text(moai_cli::ui_text::Msg::PathFirstNote, lang);
        eprintln!("moai-cli: {note}");
    }
    Ok(result)
}
