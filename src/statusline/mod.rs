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
    let profile = (!ctx.no_color()).then(|| color::Profile::from_colorterm(ctx.var("COLORTERM").unwrap_or("")));
    let status = state::collect(input.as_ref(), ctx);
    render::Renderer::new(&config, profile).render(&status) + "\n"
}

/// The CLI entry: stdin is read only when it is not a terminal (a pipe or a file).
pub fn run() -> String {
    use std::io::{IsTerminal, Read};
    let mut stdin = Vec::new();
    if !std::io::stdin().is_terminal() {
        let _ = std::io::stdin().take(4 * 1024 * 1024).read_to_end(&mut stdin);
    }
    render(&stdin, &state::Context::from_process(RENDER_BUDGET))
}
