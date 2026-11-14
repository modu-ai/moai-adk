//! `key-remove` (#2393): removes a provider's vault entry; without `--yes` a preview, and on a
//! terminal one question (#2806). Moved out of `src/main.rs` unchanged (#2834 R3).

use std::io::IsTerminal as _;

/// Remove a provider key from the vault
///
/// Remove a provider's key from the vault (#2393): the current project's entry, or the
/// shared one with `--shared`. Without `--yes` it only reports what would be removed. The
/// machine key and any legacy key file are never touched.
#[derive(clap::Args)]
pub struct Args {
    provider: String,
    #[arg(long)]
    yes: bool,
    /// The key every project uses (`~/.moai-cli/vault/_shared/`).
    #[arg(long)]
    shared: bool,
}

pub const OPERATION: &str = "key.remove";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let provider = args.provider;
    let scope = moai_cli::key_vault::scope_here(args.shared)?;
    let home = moai_cli::key_vault::home()?;
    if args.yes {
        return moai_cli::key_vault::remove(&home, &scope, &provider, true);
    }
    // Without `--yes`: the preview removes nothing. On a terminal (stdin and stderr)
    // with an entry to delete, ask once (#2806, design #2799 D2 §7); elsewhere the
    // preview with its `next:` hint is the answer, byte for byte as before.
    let preview = moai_cli::key_vault::remove(&home, &scope, &provider, false)?;
    let terminal = std::io::stdin().is_terminal() && std::io::stderr().is_terminal();
    if !(terminal && preview["exists"] == serde_json::json!(true)) {
        return Ok(preview);
    }
    use moai_cli::ui_text::Msg;
    let what = |p: &moai_cli::ui::Painter| {
        format!(
            "{}\n  {}\n",
            p.text(Msg::KeyRemoveDeletes),
            p.fill(
                Msg::KeyRemoveWhich,
                &[
                    ("provider", provider.as_str()),
                    ("scope", preview["scope"].as_str().unwrap_or("?")),
                    ("path", preview["path"].as_str().unwrap_or("?")),
                ],
            )
        )
    };
    if moai_cli::progress::ask_terminal(what, Msg::RemoveIt)? {
        moai_cli::key_vault::remove(&home, &scope, &provider, true)
    } else {
        let lang = moai_cli::ui::ui_lang(moai_cli::ui::Stream::Stderr);
        eprintln!(
            "{}: {}",
            moai_cli::ui::Token::Info.word_in(lang),
            moai_cli::ui_text::text(Msg::RemoveWithYes, lang)
        );
        Ok(preview)
    }
}
