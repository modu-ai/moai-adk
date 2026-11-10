//! `moai-cli statusline` (#2350): the Claude Code `statusLine` command, a port of the Go
//! `moai statusline` (moai-adk-go 4fa2aa9db2) checked byte for byte against the goldens
//! in `tests/fixtures/statusline/` by `tools/statusline_golden.py`.
//!
//! Split so another host adapter can reuse the parts: `input` (stdin contract),
//! `segments` (pure computations), `state` (environment, files and git, read-only),
//! `render` and `color` (layout and escape sequences), `config` (the YAML subset read).
//!
//! The output is raw text, not the JSON envelope every other operation prints: Claude
//! Code shows the command's stdout as is.

pub mod color;
pub mod config;
pub mod godecode;
pub mod input;
pub mod render;
pub mod segments;
pub mod state;

use std::time::Duration;

/// Claude Code's render budget (cli/statusline.go `statuslineRenderBudget`).
pub const RENDER_BUDGET: Duration = Duration::from_millis(800);

/// The complete stdout for one render of `stdin`, final newline included.
pub fn render(stdin: &[u8], ctx: &state::Context) -> String {
    let input = input::parse(stdin);
    // Subagent rows (Claude Code v2.1.293+): a present `tasks` array switches the output
    // to one JSON line per row and skips every collector.
    if let Some(tasks) = input.as_ref().and_then(|i| i.tasks.as_ref()) {
        return segments::subagent_rows(tasks) + "\n";
    }
    let config = state::display_config(input.as_ref(), ctx);
    let profile = (!ctx.no_color())
        .then(|| color::Profile::from_colorterm(ctx.var("COLORTERM").unwrap_or("")));
    let status = state::collect(input.as_ref(), ctx);
    render::Renderer::new(&config, profile).render(&status) + "\n"
}

/// At most this many stdin bytes are read; the rest is ignored (and the JSON then fails
/// to parse, which degrades to empty input).
pub const STDIN_CAP: u64 = 4 * 1024 * 1024;

fn read_capped(reader: impl std::io::Read) -> Vec<u8> {
    use std::io::Read;
    let mut buf = Vec::new();
    let _ = reader.take(STDIN_CAP).read_to_end(&mut buf);
    buf
}

/// The CLI entry: stdin is read only when it is not a terminal (a pipe or a file). `budget`
/// is [`RENDER_BUDGET`] except under the hidden test seam `--render-budget-ms`.
pub fn run(budget: Duration) -> String {
    use std::io::IsTerminal;
    let stdin = if std::io::stdin().is_terminal() {
        Vec::new()
    } else {
        read_capped(std::io::stdin())
    };
    render(&stdin, &state::Context::from_process(budget))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn stdin_is_read_up_to_exactly_four_mebibytes() {
        assert_eq!(read_capped(&vec![b' '; 4_194_304][..]).len(), 4_194_304);
        assert_eq!(read_capped(&vec![b' '; 4_194_305][..]).len(), 4_194_304);
        assert_eq!(read_capped(&vec![b' '; 4_194_303][..]).len(), 4_194_303);
    }

    /// #2350 audit F2: either variable set to a non-empty value turns colour off, as Go's
    /// `NO_COLOR != "" || MOAI_NO_COLOR != ""` (internal/cli/statusline.go).
    #[test]
    fn no_color_and_moai_no_color_turn_escapes_off() {
        let stdin = br#"{"model":{"display_name":"Opus"}}"#;
        let out = |vars: &[(&str, &str)]| render(stdin, &state::Context::for_test(vars, None, 0));
        for colored in [
            out(&[("COLORTERM", "truecolor")]),
            out(&[]),
            out(&[("NO_COLOR", ""), ("MOAI_NO_COLOR", "")]),
        ] {
            assert!(colored.contains('\x1b'), "{colored:?}");
        }
        for plain in [
            out(&[("NO_COLOR", "1")]),
            out(&[("MOAI_NO_COLOR", "1")]),
            out(&[("COLORTERM", "truecolor"), ("MOAI_NO_COLOR", "0")]),
        ] {
            assert!(!plain.contains('\x1b'), "{plain:?}");
            assert!(plain.contains("Opus"), "{plain:?}");
        }
    }

    #[test]
    fn render_budget_is_go_800ms() {
        assert_eq!(RENDER_BUDGET, Duration::from_millis(800));
    }
}
