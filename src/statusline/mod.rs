//! `moai-cli statusline`: the fixed three-line Claude Code `statusLine` command
//! and the separate `subagentStatusLine` JSONL protocol.
//!
//! Split so another host adapter can reuse the parts: `input` (stdin contract),
//! `segments` (pure computations), `state` (environment, files and git, read-only),
//! `render` and `color` (layout and escape sequences), `config` (the YAML subset read),
//! `counts` (the forge counts cache and its detached refresh, #2588).
//!
//! The output is raw text, not the JSON envelope every other operation prints: Claude
//! Code shows the command's stdout as is.

pub mod color;
pub mod config;
pub mod counts;
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
    let input = (stdin.len() <= STDIN_CAP as usize)
        .then(|| input::parse(stdin))
        .flatten();
    // Subagent rows (Claude Code v2.1.293+): a present `tasks` array switches the output
    // to one JSON line per row and skips every collector.
    if let Some(tasks) = input.as_ref().and_then(|i| i.tasks.as_ref()) {
        return segments::subagent_rows(tasks) + "\n";
    }
    let config = state::display_config(input.as_ref(), ctx);
    // #2559: Claude Code reads stdout through a pipe, so color never depends on a TTY
    // (the regression PR #2552 introduced); only NO_COLOR / MOAI_NO_COLOR turn it off.
    let profile = (!ctx.no_color())
        .then(|| color::Profile::from_colorterm(ctx.var("COLORTERM").unwrap_or("")));
    let status = state::collect(input.as_ref(), ctx);
    render::Renderer::new(&config, profile).render(&status) + "\n"
}

/// Largest accepted stdin payload. One extra byte is read so an oversized stream
/// degrades to no input even when its prefix is valid JSON.
pub const STDIN_CAP: u64 = 4 * 1024 * 1024;

fn read_capped(reader: impl std::io::Read) -> Vec<u8> {
    use std::io::Read;
    let mut buf = Vec::new();
    let _ = reader.take(STDIN_CAP + 1).read_to_end(&mut buf);
    buf
}

/// The CLI entry: stdin is read only when it is not a terminal (a pipe or a file). `budget`
/// is [`RENDER_BUDGET`] except under the hidden test seam `--render-budget-ms`; `now`
/// replaces the wall clock only under the hidden test seam `--now-unix` (#2756).
pub fn run(budget: Duration, now: Option<i64>) -> String {
    use std::io::IsTerminal;
    let stdin = if std::io::stdin().is_terminal() {
        Vec::new()
    } else {
        read_capped(std::io::stdin())
    };
    let mut ctx = state::Context::from_process(budget);
    if let Some(now) = now {
        ctx.now = now;
    }
    let out = render(&stdin, &ctx);
    refresh_counts_after_render(&stdin);
    if let Some(home) = std::env::var_os("HOME").filter(|h| !h.is_empty()) {
        record_usage_after_render(&stdin, std::path::Path::new(&home), unix_now());
    }
    out
}

fn unix_now() -> i64 {
    std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map(|d| d.as_secs() as i64)
        .unwrap_or(0)
}

/// #2668 S1: keep the usage windows Claude Code gave this render in
/// `$HOME/.moai-cli/state/usage-limits.json`, so a lane can renew its leases before a limit
/// stops it. Subagent rows and inputs without `rate_limits` write nothing; a failed write
/// never changes the rendered line.
fn record_usage_after_render(stdin: &[u8], home: &std::path::Path, now: i64) -> bool {
    let Some(input) = (stdin.len() <= STDIN_CAP as usize)
        .then(|| input::parse(stdin))
        .flatten()
        .filter(|i| i.tasks.is_none())
    else {
        return false;
    };
    let Some(limits) = input.rate_limits.as_ref() else {
        return false;
    };
    let pair = |name, w: Option<input::RateWindow>| {
        (
            name,
            w.and_then(|w| w.used_percentage),
            w.and_then(|w| w.resets_at),
        )
    };
    let windows = crate::usage_limit::windows_from(
        &[
            pair("five_hour", limits.five_hour),
            pair("seven_day", limits.seven_day),
        ],
        now,
    );
    crate::usage_limit::record(home, windows, now).unwrap_or(false)
}

/// #2588: after a render, start the detached counts refresh when the cache needs one.
/// The render itself stays read-only and never waits for it.
fn refresh_counts_after_render(stdin: &[u8]) {
    let Some(anchor) = (stdin.len() <= STDIN_CAP as usize)
        .then(|| input::parse(stdin))
        .flatten()
        .filter(|i| i.tasks.is_none())
        .as_ref()
        .and_then(state::resolve_anchor)
    else {
        return;
    };
    if let Ok(exe) = std::env::current_exe() {
        counts::maybe_refresh(&anchor, &exe, state::forge_suppressed(&anchor));
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    /// #2668 S1: a render with `rate_limits` leaves the record a lane reads; subagent rows,
    /// inputs without the field and garbage leave none.
    #[test]
    fn render_input_with_rate_limits_records_the_usage_windows() {
        let now = 1_800_000_000;
        let home = tempfile::TempDir::new().unwrap();
        let none: [&[u8]; 3] = [
            br#"{"model":{"display_name":"Opus"}}"#,
            br#"{"tasks":[],"rate_limits":{"five_hour":{"used_percentage":95,"resets_at":1800003600}}}"#,
            b"not json",
        ];
        for stdin in none {
            assert!(!record_usage_after_render(stdin, home.path(), now));
        }
        assert!(!crate::usage_limit::record_path(home.path()).exists());
        let stdin = br#"{"rate_limits":{"five_hour":{"used_percentage":95,"resets_at":1800003600},"seven_day":{"used_percentage":40,"resets_at":1800600000}}}"#;
        assert!(record_usage_after_render(stdin, home.path(), now));
        let v = crate::usage_limit::evaluate(home.path(), now);
        assert_eq!(v["state"], "near");
        assert_eq!(v["window"], "five_hour");
        assert_eq!(
            v["windows"].as_object().unwrap().keys().collect::<Vec<_>>(),
            vec!["five_hour", "seven_day"]
        );
    }

    #[test]
    fn stdin_boundary_accepts_the_cap_and_rejects_valid_prefixes_above_it() {
        let ctx = state::Context::for_test(&[], None);
        let prefix = br#"{"model":{"display_name":"CAP_SENTINEL"}}"#;
        for size in [4_194_303, 4_194_304, 4_194_305, 4_195_328] {
            let mut stdin = prefix.to_vec();
            stdin.resize(size as usize, b' ');
            let output = render(&read_capped(stdin.as_slice()), &ctx);
            if size <= 4_194_304 {
                assert!(output.contains("CAP_SENTINEL"), "{size}: {output:?}");
            } else {
                assert_eq!(output, render(&[], &ctx), "{size}");
                assert!(!output.contains("CAP_SENTINEL"), "{size}: {output:?}");
            }
        }
    }

    /// #2350 audit F2: either variable set to a non-empty value turns colour off, as Go's
    /// `NO_COLOR != "" || MOAI_NO_COLOR != ""` (internal/cli/statusline.go).
    #[test]
    fn no_color_and_moai_no_color_turn_escapes_off() {
        let stdin = br#"{"model":{"display_name":"Opus"}}"#;
        let out = |vars: &[(&str, &str)]| render(stdin, &state::Context::for_test(vars, None));
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

    /// #2559 regression (PR #2552): Claude Code reads the statusline through a pipe, so
    /// color must not depend on stdout being a terminal. Only NO_COLOR / MOAI_NO_COLOR
    /// turn it off; without them the default context (as Claude Code runs it) is colored,
    /// and the colored text equals the NO_COLOR text once the escapes are removed.
    #[test]
    fn piped_stdout_keeps_color_and_only_no_color_removes_it() {
        let stdin = br#"{"model":{"display_name":"Opus"},"rate_limits":{"five_hour":{"used_percentage":0}}}"#;
        for (colorterm, sgr) in [("truecolor", "38;2;"), ("", "38;5;")] {
            let colored = render(
                stdin,
                &state::Context::for_test(&[("COLORTERM", colorterm)], None),
            );
            assert!(colored.contains(sgr), "{colorterm:?}: {colored:?}");
            let plain = render(
                stdin,
                &state::Context::for_test(&[("COLORTERM", colorterm), ("NO_COLOR", "1")], None),
            );
            assert!(!plain.contains('\x1b'), "{plain:?}");
            let stripped: String = {
                let mut out = String::new();
                let mut rest = colored.as_str();
                while let Some(i) = rest.find('\x1b') {
                    out.push_str(&rest[..i]);
                    rest = &rest[i..];
                    let end = rest.find('m').map(|e| e + 1).unwrap_or(rest.len());
                    rest = &rest[end..];
                }
                out.push_str(rest);
                out
            };
            assert_eq!(stripped, plain, "{colorterm:?}");
        }
    }

    #[test]
    fn subagent_setting_keeps_its_separate_jsonl_output() {
        let stdin = br#"{
            "model":{"display_name":"MAIN_SENTINEL"},
            "tasks":[{"id":"t1","agentType":"Explore","name":"explorer",
                      "status":"running","tokenCount":12000}]
        }"#;
        assert_eq!(
            render(stdin, &state::Context::for_test(&[], None)),
            "{\"id\":\"t1\",\"content\":\"⚙ [Explore] explorer (running) · 12.0k tokens\"}\n\n"
        );
    }

    #[test]
    fn hidden_main_fields_still_receive_typed_input_validation() {
        let ctx = state::Context::for_test(&[], None);
        for stdin in [
            br#"{"model":"INVALID_SENTINEL","agent":{"name":123}}"#.as_slice(),
            br#"{"model":"INVALID_SENTINEL","pr":{"number":"wrong"}}"#.as_slice(),
        ] {
            assert_eq!(render(stdin, &ctx), render(&[], &ctx));
        }
    }
}
