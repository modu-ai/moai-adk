//! Read-only report on a `moai` candidate that is a shell routing shim (#3021 S0).
//!
//! `install-inspect` lists every `moai` it can find. When one of them is a short `sh` script that
//! hands `mcp-server` to another build without `MOAI_MCP_ONESHOT=1`, a Go audit job started
//! through it can vanish (`unknown job`). This module only reads the script text, with a size
//! bound, and says what the text shows. It never runs the file and never writes anything. A text
//! match is a hypothesis, not a finding: the report says so in `basis` and claims no defect.

use serde_json::{json, Value};
use std::{fs, io::Read, path::Path};

/// Largest script read: 64 KiB is read, one byte more is `too_large`.
pub const MAX_SCRIPT_BYTES: usize = 65536;

/// What the report rests on.
pub const BASIS: &str = "script text, not run";

/// The note a person sees for `mcp_route_without_oneshot`.
pub const NOTE: &str = "this shim hands mcp-server to the Go build without MOAI_MCP_ONESHOT=1; \
     a Go audit job started through it can vanish (unknown job). If a job answers unknown job, \
     reconnect the MCP server. Nothing was changed.";

const SHELLS: [&str; 6] = ["sh", "bash", "zsh", "dash", "ash", "ksh"];
const ROUTE_WORD: &str = "mcp-server";
const ONESHOT_WORD: &str = "MOAI_MCP_ONESHOT";

fn report(status: &str, routes: bool, sets: bool, note: Option<&str>) -> Value {
    json!({
        "status": status,
        "routes_mcp_server": routes,
        "sets_mcp_oneshot": sets,
        "basis": BASIS,
        "note": note,
    })
}

/// The `shim_mcp` field of one `moai_candidates` item. Repository content (`probed == false`)
/// is not read; anything else is read as text, at most `MAX_SCRIPT_BYTES + 1` bytes, never run.
pub fn candidate_report(path: &Path, probed: bool) -> Value {
    if !probed {
        return json!({ "status": "not_read" });
    }
    match read_bounded(path) {
        Some(bytes) => report_for_bytes(&bytes),
        None => report("unreadable", false, false, None),
    }
}

/// The file's first `MAX_SCRIPT_BYTES + 1` bytes; `None` unless it is a regular file that opens.
fn read_bounded(path: &Path) -> Option<Vec<u8>> {
    if !fs::metadata(path).ok()?.is_file() {
        return None;
    }
    let file = fs::File::open(path).ok()?;
    let mut bytes = Vec::new();
    file.take(MAX_SCRIPT_BYTES as u64 + 1)
        .read_to_end(&mut bytes)
        .ok()?;
    Some(bytes)
}

/// Classifies script bytes. Pure: no file is touched.
pub fn report_for_bytes(bytes: &[u8]) -> Value {
    if bytes.len() > MAX_SCRIPT_BYTES {
        return report("too_large", false, false, None);
    }
    let text = String::from_utf8_lossy(bytes);
    if !is_shell_script(&text) {
        return report("not_a_script", false, false, None);
    }
    let routes = text.lines().any(line_routes_mcp_server);
    let sets = text.lines().any(line_sets_oneshot);
    match (routes, sets) {
        (true, true) => report("mcp_route_with_oneshot", true, true, None),
        (true, false) => report("mcp_route_without_oneshot", true, false, Some(NOTE)),
        (false, _) => report("no_mcp_route", false, sets, None),
    }
}

/// `#!` first line naming a shell, directly or through `env`.
fn is_shell_script(text: &str) -> bool {
    let first = text.lines().next().unwrap_or("");
    let Some(rest) = first.strip_prefix("#!") else {
        return false;
    };
    let mut tokens = rest.split_whitespace();
    let base = |t: &str| t.rsplit('/').next().unwrap_or("").to_string();
    let Some(program) = tokens.next().map(base) else {
        return false;
    };
    if program == "env" {
        let named = tokens.find(|t| !t.starts_with('-') && !t.contains('='));
        return named.is_some_and(|t| SHELLS.contains(&base(t).as_str()));
    }
    SHELLS.contains(&program.as_str())
}

/// A shell `case` pattern line with `mcp-server` as a whole alternative. Comment lines never
/// match. The alternatives are the text before the first `)`; a line that ends in `|` or `|\`
/// continues the pattern list on the next line and is read whole.
fn line_routes_mcp_server(line: &str) -> bool {
    let line = line.trim();
    if line.starts_with('#') {
        return false;
    }
    let head = match line.find(')') {
        Some(i) => &line[..i],
        None => {
            let open = line.strip_suffix('\\').unwrap_or(line).trim_end();
            if !open.ends_with('|') {
                return false;
            }
            open
        }
    };
    head.trim_start_matches('(')
        .split('|')
        .any(|alternative| alternative.trim() == ROUTE_WORD)
}

/// The word `MOAI_MCP_ONESHOT` on a line that is not a comment.
fn line_sets_oneshot(line: &str) -> bool {
    if line.trim_start().starts_with('#') {
        return false;
    }
    let is_name = |c: char| c.is_ascii_alphanumeric() || c == '_';
    line.match_indices(ONESHOT_WORD).any(|(at, word)| {
        let before = line[..at].chars().next_back();
        let after = line[at + word.len()..].chars().next();
        !before.is_some_and(is_name) && !after.is_some_and(is_name)
    })
}

#[cfg(test)]
mod tests {
    use super::*;

    /// `~/go/bin/moai` on the author's machine (the name-switch shim), verbatim.
    const REAL_SHIM: &str = r####"#!/bin/sh
# MoAI name-switch shim, installed 2026-11-10 by the ADK leader on the user's instruction
# ("use the name moai for the Rust build, keep Go as moai.old").
#
# `moai` now means the Rust build. The Go build is kept as moai.old and still serves the
# subcommands only it has: hooks, the status line, the MCP server and the harness commands that
# running Claude Code sessions call. Without this routing the Rust binary would answer
# `moai hook ...` with exit 2, which Claude Code treats as "block this tool call".
#
# Undo: mv "$HOME/go/bin/moai.old" "$HOME/go/bin/moai"  (and remove ~/.moai-cli/path/moai)
RUST="${MOAI_RUST_BIN:-$HOME/.moai-cli/path/moai-cli}"
GO="${MOAI_GO_BIN:-$HOME/go/bin/moai.old}"

case "${1:-}" in
  # Launcher words always go to Rust.
  cc|glm|codex|omp|-l|-a|--leader|--agent)
    exec "$RUST" "$@" ;;
  # Subcommands only the Go build has.
  af|agent|ast-edit|ast-grep|bundle|chain|ci-verdict|clean|completion|config|constitution|contract|decision|doctor|epic|factory|feedback|fetch|gate|github|goal|graph|gtd|handoff|harness|hook|inbox|init|integration|inventory|jev|jevmeasure|loop|lsp|mcp|mcp-server|memory|migrate|migration|moai|mx|paths|plan|pr|preference|profile|research|session|slot|spec|state|status|statusline|telemetry|that|todo|tokens|tool-policy|verify|version|web|workflow|worktree)
    if [ -x "$GO" ]; then
      exec "$GO" "$@"
    fi
    echo "moai: '$1' belongs to the Go build, which was not found at $GO" >&2
    exit 1 ;;
esac
exec "$RUST" "$@"
"####;

    fn status(bytes: &[u8]) -> String {
        report_for_bytes(bytes)["status"]
            .as_str()
            .unwrap()
            .to_string()
    }

    fn flags(bytes: &[u8]) -> (bool, bool) {
        let r = report_for_bytes(bytes);
        (
            r["routes_mcp_server"].as_bool().unwrap(),
            r["sets_mcp_oneshot"].as_bool().unwrap(),
        )
    }

    /// A shell script whose only case pattern line is `pattern`.
    fn with_pattern(pattern: &str) -> String {
        format!("#!/bin/sh\ncase \"$1\" in\n{pattern}\n    exec go \"$@\" ;;\nesac\n")
    }

    /// `core` padded with a final comment line to exactly `len` bytes.
    fn padded(core: &str, len: usize) -> Vec<u8> {
        let pad = len - core.len() - 1;
        let mut s = String::from(core);
        s.push('#');
        s.push_str(&"x".repeat(pad));
        assert_eq!(s.len(), len);
        s.into_bytes()
    }

    #[test]
    fn the_real_shim_routes_mcp_server_without_oneshot() {
        let r = report_for_bytes(REAL_SHIM.as_bytes());
        assert_eq!(r["status"], json!("mcp_route_without_oneshot"), "{r}");
        assert_eq!(r["routes_mcp_server"], json!(true));
        assert_eq!(r["sets_mcp_oneshot"], json!(false));
        assert_eq!(r["basis"], json!("script text, not run"));
        assert_eq!(
            r["note"],
            json!(
                "this shim hands mcp-server to the Go build without MOAI_MCP_ONESHOT=1; a Go \
                 audit job started through it can vanish (unknown job). If a job answers \
                 unknown job, reconnect the MCP server. Nothing was changed (#3021)"
            )
        );
    }

    #[test]
    fn the_real_shim_with_the_env_line_sets_oneshot() {
        let fixed = REAL_SHIM.replace(
            "      exec \"$GO\" \"$@\"\n",
            "      export MOAI_MCP_ONESHOT=1\n      exec \"$GO\" \"$@\"\n",
        );
        assert_ne!(fixed, REAL_SHIM, "the line was inserted");
        let r = report_for_bytes(fixed.as_bytes());
        assert_eq!(r["status"], json!("mcp_route_with_oneshot"), "{r}");
        assert_eq!(r["note"], json!(null));
        assert_eq!(flags(fixed.as_bytes()), (true, true));
    }

    #[test]
    fn a_commented_out_route_is_no_route() {
        let s = with_pattern("  # mcp-server)");
        assert_eq!(status(s.as_bytes()), "no_mcp_route");
        let s = with_pattern("# cc|mcp-server|hook)");
        assert_eq!(flags(s.as_bytes()), (false, false));
        // The same text uncommented is a route: the comment check is what decides.
        let s = with_pattern("  mcp-server)");
        assert_eq!(status(s.as_bytes()), "mcp_route_without_oneshot");
    }

    #[test]
    fn only_a_whole_alternative_is_a_route() {
        for pattern in [
            "  mcp-server-x)",
            "  xmcp-server)",
            "  mcp)",
            "  hook|mcp-server-x|mcp)",
            "  mcp-server2|x-mcp-server)",
        ] {
            let s = with_pattern(pattern);
            assert_eq!(status(s.as_bytes()), "no_mcp_route", "{pattern}");
        }
        for pattern in [
            "  mcp-server)",
            "  mcp-server|hook)",
            "  hook | mcp-server )",
            "  (mcp-server)",
            "  mcp-server) exec go \"$@\" ;;",
        ] {
            let s = with_pattern(pattern);
            assert_eq!(
                status(s.as_bytes()),
                "mcp_route_without_oneshot",
                "{pattern}"
            );
        }
    }

    #[test]
    fn the_last_alternative_and_a_split_list_are_routes() {
        let last = with_pattern("  hook|status|mcp-server)");
        assert_eq!(flags(last.as_bytes()), (true, false));
        let split = "#!/bin/sh\ncase \"$1\" in\n  hook|\\\n  status|\\\n  mcp-server)\n    exec go ;;\nesac\n";
        assert_eq!(flags(split.as_bytes()), (true, false));
        let split_first =
            "#!/bin/sh\ncase \"$1\" in\n  mcp-server|\\\n  hook)\n    exec go ;;\nesac\n";
        assert_eq!(flags(split_first.as_bytes()), (true, false));
        // A command that merely mentions the word is not a pattern line.
        let prose = "#!/bin/sh\necho mcp-server\nexec go mcp-server\n";
        assert_eq!(flags(prose.as_bytes()), (false, false));
        // Nor is a bare word that does not end a pattern list.
        let bare = "#!/bin/sh\nmcp-server\n";
        assert_eq!(flags(bare.as_bytes()), (false, false));
    }

    #[test]
    fn a_commented_oneshot_word_does_not_count() {
        let only_comment = REAL_SHIM.replace(
            "      exec \"$GO\" \"$@\"\n",
            "      # export MOAI_MCP_ONESHOT=1\n      exec \"$GO\" \"$@\"\n",
        );
        assert_ne!(only_comment, REAL_SHIM);
        assert_eq!(flags(only_comment.as_bytes()), (true, false));
        // The word only, not a longer name.
        let longer = REAL_SHIM.replace(
            "      exec \"$GO\" \"$@\"\n",
            "      export MOAI_MCP_ONESHOT_X=1\n      exec \"$GO\" \"$@\"\n",
        );
        assert_eq!(flags(longer.as_bytes()), (true, false));
        // Used inline in a command prefix.
        let inline = REAL_SHIM.replace(
            "      exec \"$GO\" \"$@\"\n",
            "      MOAI_MCP_ONESHOT=1 exec \"$GO\" \"$@\"\n",
        );
        assert_eq!(flags(inline.as_bytes()), (true, true));
    }

    #[test]
    fn a_oneshot_word_without_a_route_is_no_route() {
        let s = "#!/bin/sh\nexport MOAI_MCP_ONESHOT=1\nexec go \"$@\"\n";
        let r = report_for_bytes(s.as_bytes());
        assert_eq!(r["status"], json!("no_mcp_route"), "{r}");
        assert_eq!(r["note"], json!(null));
    }

    #[test]
    fn only_a_shell_shebang_is_a_script() {
        let body = "\ncase \"$1\" in\n  mcp-server) exec go ;;\nesac\n";
        for first in [
            "#!/bin/sh",
            "#!/bin/bash",
            "#!/usr/bin/env sh",
            "#!/usr/bin/env -S bash -e",
            "#!/bin/zsh -f",
            "#! /bin/sh",
            "#!/bin/sh\r",
        ] {
            let s = format!("{first}{body}");
            assert_eq!(status(s.as_bytes()), "mcp_route_without_oneshot", "{first}");
        }
        for first in [
            "#!/usr/bin/env python3",
            "#!/usr/bin/perl",
            "#!",
            "#!/usr/bin/env",
            "# /bin/sh",
            "/bin/sh",
        ] {
            let s = format!("{first}{body}");
            assert_eq!(status(s.as_bytes()), "not_a_script", "{first}");
        }
        // The line must be the first line.
        let late = format!("\n#!/bin/sh{body}");
        assert_eq!(status(late.as_bytes()), "not_a_script");
    }

    #[test]
    fn a_binary_is_not_a_script() {
        let mut macho = vec![0xCF, 0xFA, 0xED, 0xFE, 0x0C, 0x00, 0x00, 0x01];
        macho.extend_from_slice(b"mcp-server)\nMOAI_MCP_ONESHOT\n");
        let r = report_for_bytes(&macho);
        assert_eq!(r["status"], json!("not_a_script"), "{r}");
        assert_eq!(flags(&macho), (false, false));
        assert_eq!(status(b"\x7fELF\x02\x01\x01"), "not_a_script");
    }

    #[test]
    fn an_empty_file_is_not_a_script() {
        let r = report_for_bytes(b"");
        assert_eq!(r["status"], json!("not_a_script"), "{r}");
        assert_eq!(flags(b""), (false, false));
    }

    #[test]
    fn the_size_limit_holds_at_its_edge() {
        let core = "#!/bin/sh\ncase \"$1\" in\n  mcp-server)\n    exec go ;;\nesac\n";
        let at_limit = padded(core, MAX_SCRIPT_BYTES);
        assert_eq!(at_limit.len(), 65536);
        assert_eq!(status(&at_limit), "mcp_route_without_oneshot", "N is read");
        let over = padded(core, MAX_SCRIPT_BYTES + 1);
        assert_eq!(over.len(), 65537);
        let r = report_for_bytes(&over);
        assert_eq!(r["status"], json!("too_large"), "N+1 is not");
        assert_eq!(r["routes_mcp_server"], json!(false));
        assert_eq!(r["note"], json!(null));
        assert_eq!(
            status(&padded(core, MAX_SCRIPT_BYTES - 1)),
            "mcp_route_without_oneshot"
        );
    }

    #[test]
    fn crlf_line_endings_are_read() {
        let crlf = REAL_SHIM.replace('\n', "\r\n");
        let r = report_for_bytes(crlf.as_bytes());
        assert_eq!(r["status"], json!("mcp_route_without_oneshot"), "{r}");
        let fixed = crlf.replace(
            "      exec \"$GO\" \"$@\"\r\n",
            "      export MOAI_MCP_ONESHOT=1\r\n      exec \"$GO\" \"$@\"\r\n",
        );
        assert_eq!(status(fixed.as_bytes()), "mcp_route_with_oneshot");
        // `mcp-server)` as the whole of a CRLF line.
        let s = "#!/bin/sh\r\ncase \"$1\" in\r\n  mcp-server)\r\n    exec go ;;\r\nesac\r\n";
        assert_eq!(flags(s.as_bytes()), (true, false));
    }

    #[test]
    fn a_repository_candidate_is_not_read() {
        let dir = tempfile::TempDir::new().unwrap();
        let p = dir.path().join("moai");
        fs::write(&p, REAL_SHIM).unwrap();
        assert_eq!(candidate_report(&p, false), json!({ "status": "not_read" }));
        let read = candidate_report(&p, true);
        assert_eq!(read["status"], json!("mcp_route_without_oneshot"), "{read}");
    }

    #[test]
    fn a_missing_file_or_a_folder_is_unreadable() {
        let dir = tempfile::TempDir::new().unwrap();
        for p in [dir.path().join("absent"), dir.path().to_path_buf()] {
            let r = candidate_report(&p, true);
            assert_eq!(r["status"], json!("unreadable"), "{}: {r}", p.display());
            assert_eq!(r["routes_mcp_server"], json!(false));
            assert_eq!(r["basis"], json!("script text, not run"));
            assert_eq!(r["note"], json!(null));
        }
    }

    #[test]
    fn a_large_file_is_reported_without_reading_it_whole() {
        let dir = tempfile::TempDir::new().unwrap();
        let p = dir.path().join("moai");
        fs::write(&p, vec![b'#'; MAX_SCRIPT_BYTES * 4]).unwrap();
        assert_eq!(candidate_report(&p, true)["status"], json!("too_large"));
    }
}
