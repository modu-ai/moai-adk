//! `key-set` (#2393): stores a provider key, encrypted, in the vault; the value comes from
//! stdin only. Moved out of `src/main.rs` unchanged (#2834 R3).

use std::io::IsTerminal;

/// Store a provider key in the vault
///
/// Store a provider's key, encrypted, in the vault (#2393): for the current project, or
/// for every project with `--shared`. The value is read from stdin only (never argv); an
/// interactive terminal is refused. The answer never contains the value.
#[derive(clap::Args)]
pub struct Args {
    provider: String,
    /// The key every project uses (`~/.moai-cli/vault/_shared/`).
    #[arg(long)]
    shared: bool,
}

pub const OPERATION: &str = "key.set";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    if std::io::stdin().is_terminal() {
        return Err(
            "key-set reads the value from stdin and refuses a terminal (the \
                    value would be echoed); pipe it in instead"
                .into(),
        );
    }
    let scope = moai_cli::key_vault::scope_here(args.shared)?;
    let value = moai_cli::key_vault::read_value(std::io::stdin())?;
    moai_cli::key_vault::set(
        &moai_cli::key_vault::home()?,
        &scope,
        &args.provider,
        &value,
    )
}
