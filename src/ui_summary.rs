//! The human summary of a plan/apply result on stderr (Issue #2805, design #2799 D2 §6, leader
//! decision D3 6865936070; extraction rules from agent-28's preparation 6866050307).
//!
//! The summary reads the same `result` the stdout envelope carries — nothing is computed twice
//! and the envelope is never touched. It prints only when stderr is a terminal and
//! `MOAI_OUTPUT` is not `quiet`; `MOAI_OUTPUT=plain` keeps the table without color or UTF-8
//! symbols. Hook mode never reaches it (the caller skips it).
//!
//! One extraction serves every plan/apply operation: the first array of objects in `result`
//! whose rows carry an action word. Row name: `path` → `host` → `target` → `selection` →
//! `name`. Action: `action` → `decision` → `status` → `kind`. Note: `reason` → `detail` →
//! `reasons[0]` (+N more) → `status` when it is not the action. Every action word is printed
//! next to its symbol, so the meaning never rides on color alone.

use crate::ui::{Painter, Stream, Token};
use serde_json::Value;
use unicode_width::{UnicodeWidthChar, UnicodeWidthStr};

/// Rows shown before the rest is counted.
pub const MAX_ROWS: usize = 50;
/// The width used when neither `COLUMNS` nor the terminal tells one.
pub const DEFAULT_WIDTH: usize = 80;
/// The narrowest the name column is squeezed to before the note is cut instead.
const MIN_NAME: usize = 12;

const NAME_KEYS: [&str; 5] = ["path", "host", "target", "selection", "name"];
const ACTION_KEYS: [&str; 4] = ["action", "decision", "status", "kind"];

/// Whether `op` is one the summary is for: previews, applies and reverts.
pub fn is_summary_op(op: &str) -> bool {
    let last = op.rsplit('.').next().unwrap_or(op);
    last.contains("plan") || last.contains("apply") || last.contains("revert") || op == "update.run"
}

/// The command a user types for `op` (`go_state.revert_plan` → `go-state-revert-plan`).
pub fn command_of(op: &str) -> String {
    op.replace(['.', '_'], "-")
}

/// The command that applies a preview of `op`, by naming rule: `…-plan` → `…-apply`, except
/// `…-revert-plan` → `…-revert`. The caller checks that the command exists.
pub fn apply_command_of(op: &str) -> Option<String> {
    let cmd = command_of(op);
    if let Some(stem) = cmd.strip_suffix("-revert-plan") {
        return Some(format!("{stem}-revert"));
    }
    cmd.strip_suffix("-plan")
        .map(|stem| format!("{stem}-apply"))
}

#[derive(Debug, PartialEq, Eq)]
struct Row {
    action: String,
    name: String,
    note: String,
}

/// A result value as display text. Strings pass through [`crate::ui::visible`] here, where the
/// data enters the table, so every cell — painted or not — and every width is computed on text
/// without control characters (#2818 audit B1).
fn text(v: &Value) -> Option<String> {
    match v {
        Value::String(s) if !s.is_empty() => Some(crate::ui::visible(s).into_owned()),
        Value::Number(n) => Some(n.to_string()),
        Value::Bool(b) => Some(b.to_string()),
        _ => None,
    }
}

fn first_key(obj: &serde_json::Map<String, Value>, keys: &[&'static str]) -> Option<&'static str> {
    keys.iter()
        .copied()
        .find(|k| obj.get(*k).and_then(text).is_some())
}

type Obj = serde_json::Map<String, Value>;

/// Where a row's action word is: on the row itself, or one level down in a nested object
/// (`home-routing-plan` rows are `{host, …, section: {path, action, …}}`).
fn action_source(o: &Obj) -> Option<(&Obj, &'static str)> {
    if let Some(k) = first_key(o, &ACTION_KEYS) {
        return Some((o, k));
    }
    o.values()
        .filter_map(Value::as_object)
        .find_map(|inner| first_key(inner, &ACTION_KEYS).map(|k| (inner, k)))
}

/// The first array of objects that carries action words, with its name key.
fn extract(result: &Value) -> Option<(&'static str, Vec<Row>)> {
    let fields = result.as_object()?;
    for value in fields.values() {
        let Some(items) = value.as_array() else {
            continue;
        };
        let objects: Vec<&Obj> = items.iter().filter_map(Value::as_object).collect();
        if objects.is_empty() || objects.len() != items.len() {
            continue;
        }
        if !objects.iter().any(|o| action_source(o).is_some()) {
            continue;
        }
        let name_key = objects
            .iter()
            .find_map(|o| first_key(o, &NAME_KEYS))
            .unwrap_or("name");
        let header = match name_key {
            "path" => "PATH",
            "host" => "HOST",
            "target" => "TARGET",
            "selection" => "SELECTION",
            _ => "NAME",
        };
        let rows = objects
            .iter()
            .map(|o| {
                let name = o.get(name_key).and_then(text).unwrap_or_else(|| "-".into());
                let Some((src, key)) = action_source(o) else {
                    return Row {
                        action: "-".into(),
                        name,
                        note: note_of(o, ""),
                    };
                };
                let action = src.get(key).and_then(text).unwrap_or_else(|| "-".into());
                let mut note = note_of(o, key);
                if note.is_empty() && !std::ptr::eq(src, *o) {
                    note = note_of(src, key);
                }
                Row { action, name, note }
            })
            .collect();
        return Some((header, rows));
    }
    None
}

fn note_of(o: &Obj, action_key: &str) -> String {
    for key in ["reason", "blocked_reason", "detail"] {
        if let Some(t) = o.get(key).and_then(text) {
            return t;
        }
    }
    if let Some(list) = o.get("reasons").and_then(Value::as_array) {
        let texts: Vec<String> = list.iter().filter_map(text).collect();
        if let Some(first) = texts.first() {
            return match texts.len() {
                1 => first.clone(),
                n => format!("{first} (+{} more)", n - 1),
            };
        }
    }
    if action_key != "status" {
        if let Some(t) = o.get("status").and_then(text) {
            return t;
        }
    }
    String::new()
}

/// Which token an action word gets; unknown words are guidance (`info`), printed as they are.
pub fn token_of(action: &str) -> Token {
    match action {
        "unchanged" | "keep" | "kept" | "skip" | "skipped" | "same" | "none" | "ok" | "applied"
        | "created" | "updated" | "removed" | "moved" | "reverted" | "committed" => Token::Ok,
        "leave" | "blocked" | "preserved" | "user_choice" | "remove_blocked" | "partial"
        | "keep_changed" => Token::Warn,
        "conflict" | "refused" | "error" | "invalid" | "failed" | "stale_target" => Token::Error,
        _ => Token::Info,
    }
}

fn width(s: &str) -> usize {
    UnicodeWidthStr::width(s)
}

/// `s` cut to `max` cells at the end, with the ellipsis inside the limit.
fn cut_end(s: &str, max: usize, ellipsis: &str) -> String {
    if width(s) <= max {
        return s.to_string();
    }
    let room = max.saturating_sub(width(ellipsis));
    let mut out = String::new();
    let mut used = 0;
    for c in s.chars() {
        let w = c.width().unwrap_or(0);
        if used + w > room {
            break;
        }
        out.push(c);
        used += w;
    }
    out + ellipsis
}

/// `s` cut to `max` cells in the middle (a third of the room in front, the rest at the end, so a
/// file name survives).
fn cut_middle(s: &str, max: usize, ellipsis: &str) -> String {
    if width(s) <= max {
        return s.to_string();
    }
    let room = max.saturating_sub(width(ellipsis));
    let front_room = room / 3;
    let back_room = room - front_room;
    let mut front = String::new();
    let mut used = 0;
    for c in s.chars() {
        let w = c.width().unwrap_or(0);
        if used + w > front_room {
            break;
        }
        front.push(c);
        used += w;
    }
    let mut back: Vec<char> = Vec::new();
    let mut used = 0;
    for c in s.chars().rev() {
        let w = c.width().unwrap_or(0);
        if used + w > back_room {
            break;
        }
        back.push(c);
        used += w;
    }
    back.reverse();
    format!("{front}{ellipsis}{}", back.into_iter().collect::<String>())
}

fn pad(painted: String, plain_width: usize, to: usize) -> String {
    painted + &" ".repeat(to.saturating_sub(plain_width))
}

/// The summary text for `op`'s `result`, laid out in `columns` cells. `apply_command` is the
/// verified command a preview's `plan_digest` goes to, if any.
pub fn render(
    op: &str,
    result: &Value,
    painter: &Painter,
    columns: usize,
    apply_command: Option<&str>,
) -> String {
    let ellipsis = if painter.unicode { "…" } else { "..." };
    let mut out = String::new();
    let target = ["project", "home", "root"]
        .iter()
        .find_map(|k| result.get(*k).and_then(text));
    let mark = Token::Accent.symbol(painter.unicode);
    let command = command_of(op);
    let mut title = format!(
        "{} {}",
        painter.mark(Token::Accent),
        painter.paint(Token::Strong, &command)
    );
    if let Some(t) = target {
        let room = columns.saturating_sub(width(mark) + 1 + width(&command) + 2);
        if room >= MIN_NAME {
            title.push_str("  ");
            title.push_str(&painter.paint(Token::Muted, &cut_middle(&t, room, ellipsis)));
        }
    }
    out.push_str(&title);
    out.push('\n');

    let writes_nothing = op.rsplit('.').next().is_some_and(|l| l.contains("plan"))
        || result.get("writes") == Some(&Value::Bool(false));
    let mut totals: Vec<(String, usize)> = Vec::new();
    if let Some((header, rows)) = extract(result) {
        for r in &rows {
            match totals.iter_mut().find(|(a, _)| *a == r.action) {
                Some((_, n)) => *n += 1,
                None => totals.push((r.action.clone(), 1)),
            }
        }
        let shown = &rows[..rows.len().min(MAX_ROWS)];
        let status_text = |r: &Row| {
            let symbol = token_of(&r.action).symbol(painter.unicode);
            if symbol.is_empty() {
                r.action.clone()
            } else {
                format!("{symbol} {}", r.action)
            }
        };
        let act_w = shown
            .iter()
            .map(|r| width(&status_text(r)))
            .chain([width("ACTION")])
            .max()
            .unwrap_or(6);
        let natural_name = shown
            .iter()
            .map(|r| width(&r.name))
            .chain([width(header)])
            .max()
            .unwrap_or(4);
        let natural_note = shown.iter().map(|r| width(&r.note)).max().unwrap_or(0);
        let has_note = natural_note > 0;
        let gaps = if has_note { 4 } else { 2 };
        let budget = columns.saturating_sub(act_w + gaps);
        let (name_w, note_w) = if !has_note {
            (natural_name.min(budget.max(MIN_NAME)), 0)
        } else if natural_name + natural_note <= budget {
            (natural_name, natural_note)
        } else {
            let name_w = natural_name
                .min(budget.saturating_sub(natural_note.min(budget / 2)))
                .max(MIN_NAME.min(natural_name));
            (name_w, budget.saturating_sub(name_w))
        };
        out.push('\n');
        let mut head = pad(
            painter.paint(Token::Strong, "ACTION"),
            width("ACTION"),
            act_w + 2,
        );
        if has_note {
            head.push_str(&pad(
                painter.paint(Token::Strong, header),
                width(header),
                name_w + 2,
            ));
            head.push_str(&painter.paint(Token::Strong, "NOTE"));
        } else {
            head.push_str(&painter.paint(Token::Strong, header));
        }
        out.push_str(head.trim_end());
        out.push('\n');
        for r in shown {
            let status = status_text(r);
            let token = token_of(&r.action);
            let mut line = pad(painter.paint(token, &status), width(&status), act_w + 2);
            let name = cut_middle(&r.name, name_w, ellipsis);
            if has_note {
                line.push_str(&pad(name.clone(), width(&name), name_w + 2));
                let note = cut_end(&r.note, note_w, ellipsis);
                line.push_str(&painter.paint(Token::Muted, &note));
            } else {
                line.push_str(&name);
            }
            out.push_str(line.trim_end());
            out.push('\n');
        }
        if rows.len() > MAX_ROWS {
            out.push_str(&painter.paint(
                Token::Muted,
                &format!("{ellipsis} {} more", rows.len() - MAX_ROWS),
            ));
            out.push('\n');
        }
        out.push('\n');
    }

    let mut parts: Vec<String> = totals
        .iter()
        .map(|(a, n)| painter.paint(token_of(a), &format!("{n} {a}")))
        .collect();
    if writes_nothing {
        parts.push("writes nothing".to_string());
    }
    if parts.is_empty() {
        parts.push("done".to_string());
    }
    out.push_str(&parts.join(if painter.unicode { " · " } else { " | " }));
    out.push('\n');

    if let (Some(cmd), Some(digest)) = (apply_command, result.get("plan_digest").and_then(text)) {
        // The digest is copied as it is, so it is never cut: it gets a line of its own, the
        // only line that may run past a narrow terminal.
        out.push_str(&format!(
            "{} {}\n  {}\n",
            painter.paint(
                Token::Info,
                &format!("{} next:", Token::Info.symbol(painter.unicode))
            ),
            painter.paint(Token::Strong, &format!("moai {cmd} <same arguments>")),
            painter.paint(Token::Strong, &format!("--plan-digest {digest}"))
        ));
    }
    out
}

/// What `MOAI_OUTPUT` asks for.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum OutputMode {
    Normal,
    Plain,
    Quiet,
}

pub fn output_mode(value: Option<&std::ffi::OsStr>) -> OutputMode {
    match value.and_then(|v| v.to_str()).map(str::trim) {
        Some("quiet") => OutputMode::Quiet,
        Some("plain") => OutputMode::Plain,
        _ => OutputMode::Normal,
    }
}

/// Print the summary of a successful `op` to stderr when stderr is a terminal (D2 §1, §6).
/// `command_exists` tells whether a command name is a real subcommand, so the `next:` line
/// never names a command that does not exist.
pub fn maybe_print(op: &str, result: &Value, command_exists: impl Fn(&str) -> bool) {
    use std::io::{IsTerminal, Write};
    if !is_summary_op(op) || !std::io::stderr().is_terminal() {
        return;
    }
    let mode = output_mode(std::env::var_os("MOAI_OUTPUT").as_deref());
    let painter = match mode {
        OutputMode::Quiet => return,
        OutputMode::Plain => Painter::plain(),
        OutputMode::Normal => Painter::for_stream(Stream::Stderr, crate::ui::ColorFlag::Auto),
    };
    let env = crate::ui::Env::from_process();
    let columns = crate::ui::width(Stream::Stderr, &env).unwrap_or(DEFAULT_WIDTH);
    let apply = apply_command_of(op).filter(|c| command_exists(c));
    let text = render(op, result, &painter, columns, apply.as_deref());
    let _ = std::io::stderr().write_all(text.as_bytes());
}

#[cfg(test)]
mod tests {
    use super::*;
    use serde_json::json;

    fn plain() -> Painter {
        Painter::plain()
    }

    fn utf8() -> Painter {
        Painter {
            color: false,
            unicode: true,
        }
    }

    #[test]
    fn summary_ops_are_previews_applies_and_reverts() {
        for op in [
            "go_state.plan",
            "go_state.apply",
            "go_state.revert_plan",
            "go_state.revert",
            "agents.home_plan",
            "worktree.gc.apply",
            "update.run",
        ] {
            assert!(is_summary_op(op), "{op}");
        }
        for op in [
            "claim.verify",
            "identity",
            "project.detect",
            "repo_scans.judge",
        ] {
            assert!(!is_summary_op(op), "{op}");
        }
    }

    #[test]
    fn apply_commands_follow_the_naming_rule() {
        assert_eq!(
            apply_command_of("go_state.plan").as_deref(),
            Some("go-state-apply")
        );
        assert_eq!(
            apply_command_of("go_state.revert_plan").as_deref(),
            Some("go-state-revert")
        );
        assert_eq!(
            apply_command_of("agents.home_plan").as_deref(),
            Some("agents-home-apply")
        );
        assert_eq!(
            apply_command_of("worktree.gc.plan").as_deref(),
            Some("worktree-gc-apply")
        );
        assert_eq!(apply_command_of("go_state.apply"), None);
    }

    #[test]
    fn each_known_row_shape_is_extracted() {
        let go = json!({"project": "/w", "items": [
            {"path": ".moai/config/", "action": "move", "reason": "to backup"},
            {"path": ".moai/specs/", "action": "leave", "reason": "kept"}]});
        let (h, rows) = extract(&go).unwrap();
        assert_eq!(h, "PATH");
        assert_eq!(
            rows[0],
            Row {
                action: "move".into(),
                name: ".moai/config/".into(),
                note: "to backup".into()
            }
        );
        let gc = json!({"worktrees": [{"path": "a", "action": "skip", "reasons": ["claim", "hold", "dirty"]}]});
        assert_eq!(extract(&gc).unwrap().1[0].note, "claim (+2 more)");
        let routing = json!({"hosts": [{"host": "claude", "action": "create"}]});
        assert_eq!(extract(&routing).unwrap().0, "HOST");
        // `home-routing-plan`: the action sits one level down, the name on the row.
        let nested = json!({"home": "/h", "hosts": [
            {"host": "claude", "file": "/h/.claude/CLAUDE.md", "preflight": "ok",
             "section": {"path": "/h/.claude/CLAUDE.md", "action": "create", "detail": null}},
            {"host": "codex", "preflight": "blocked", "blocked_reason": "not a regular file",
             "section": {"path": "/h/.codex/AGENTS.md", "action": "conflict"}}],
            "not_supported": [{"host": "omp", "reason": "shadows other files"}]});
        let (h, rows) = extract(&nested).unwrap();
        assert_eq!(h, "HOST");
        assert_eq!(
            rows,
            vec![
                Row {
                    action: "create".into(),
                    name: "claude".into(),
                    note: String::new()
                },
                Row {
                    action: "conflict".into(),
                    name: "codex".into(),
                    note: "not a regular file".into()
                },
            ]
        );
        let receipt = json!({"entries": [{"path": "x", "decision": "applied", "status": "ok"}]});
        let r = &extract(&receipt).unwrap().1[0];
        assert_eq!((r.action.as_str(), r.note.as_str()), ("applied", "ok"));
        let migrate =
            json!({"rows": [{"selection": "skills", "ownership": "go", "action": "copy"}]});
        assert_eq!(extract(&migrate).unwrap().0, "SELECTION");
    }

    #[test]
    fn arrays_without_action_words_are_not_tables() {
        assert!(extract(&json!({"warnings": ["a", "b"]})).is_none());
        assert!(extract(&json!({"files": [{"path": "a"}]})).is_none());
        assert!(extract(&json!({"items": []})).is_none());
        let text = render(
            "layout.plan",
            &json!({"warnings": ["a"]}),
            &plain(),
            80,
            None,
        );
        assert_eq!(text, "* layout-plan\nwrites nothing\n");
    }

    #[test]
    fn unknown_action_words_print_as_they_are() {
        assert_eq!(token_of("frobnicate"), Token::Info);
        let text = render(
            "x.apply",
            &json!({"items": [{"path": "a", "action": "frobnicate"}]}),
            &plain(),
            80,
            None,
        );
        assert!(text.contains("-> frobnicate  a"), "{text}");
        assert!(text.contains("1 frobnicate"), "{text}");
    }

    #[test]
    fn the_table_has_a_title_a_head_rows_totals_and_next() {
        let result = json!({"project": "/work/demo", "plan_digest": "sha256:abc", "items": [
            {"path": ".moai/config/", "action": "move", "reason": "to .moai-backup/"},
            {"path": ".moai/specs/", "action": "keep", "reason": "copied back"},
            {"path": ".moai/worktrees/", "action": "leave", "reason": "1 git worktree registered"}]});
        let text = render(
            "go_state.plan",
            &result,
            &utf8(),
            80,
            Some("go-state-apply"),
        );
        let want = "\
◆ go-state-plan  /work/demo

ACTION   PATH              NOTE
→ move   .moai/config/     to .moai-backup/
✓ keep   .moai/specs/      copied back
! leave  .moai/worktrees/  1 git worktree registered

1 move · 1 keep · 1 leave · writes nothing
→ next: moai go-state-apply <same arguments>
  --plan-digest sha256:abc
";
        assert_eq!(text, want);
        // Plain (pipe-like) symbols are ASCII and there is no escape anywhere.
        let ascii = render(
            "go_state.plan",
            &result,
            &plain(),
            80,
            Some("go-state-apply"),
        );
        assert!(ascii.is_ascii() && !ascii.contains('\x1b'), "{ascii}");
        // A missing apply command means no next line.
        assert!(!render("go_state.plan", &result, &utf8(), 80, None).contains("next:"));
    }

    #[test]
    fn color_is_four_bit_and_every_status_keeps_its_word() {
        let p = Painter {
            color: true,
            unicode: true,
        };
        let text = render(
            "x.plan",
            &json!({"items": [{"path": "a", "action": "conflict"}, {"path": "b", "action": "create"}]}),
            &p,
            80,
            None,
        );
        assert!(text.contains("\x1b[31m✗ conflict\x1b[0m"), "{text:?}");
        assert!(text.contains("\x1b[36m→ create\x1b[0m"), "{text:?}");
        assert!(!text.contains("38;5;") && !text.contains("38;2;"));
    }

    #[test]
    fn hangul_counts_two_cells_and_long_paths_lose_their_middle() {
        assert_eq!(width("한글"), 4);
        // Exact fit: nothing cut. One over: cut, still within the limit.
        assert_eq!(cut_middle("가나다라", 8, "…"), "가나다라");
        let cut = cut_middle("가나다라마", 9, "…");
        assert!(width(&cut) <= 9, "{cut}");
        assert!(cut.contains('…') && cut.ends_with("마"), "{cut}");
        assert_eq!(cut_end("abcdef", 6, "..."), "abcdef");
        assert_eq!(cut_end("abcdefg", 6, "..."), "abc...");
        let long = format!("src/{}/파일이름.rs", "a".repeat(80));
        let result = json!({"items": [{"path": long, "action": "create", "reason": "새 파일"}]});
        for columns in [60usize, 80] {
            let text = render("x.plan", &result, &utf8(), columns, None);
            for line in text.lines() {
                assert!(
                    width(line) <= columns,
                    "{columns}: {line:?} is {} cells",
                    width(line)
                );
            }
            assert!(text.contains("파일이름.rs"), "{text}");
            assert!(text.contains("새 파일"), "{text}");
        }
    }

    #[test]
    fn more_than_fifty_rows_are_counted_not_listed() {
        let rows = |n: usize| json!({"items": (0..n).map(|i| json!({"path": format!("f{i}"), "action": "create"})).collect::<Vec<_>>()});
        let fifty = render("x.plan", &rows(MAX_ROWS), &plain(), 80, None);
        assert!(!fifty.contains("more"), "{fifty}");
        assert_eq!(
            fifty.lines().filter(|l| l.starts_with("-> create")).count(),
            MAX_ROWS
        );
        let more = render("x.plan", &rows(MAX_ROWS + 1), &plain(), 80, None);
        assert_eq!(
            more.lines().filter(|l| l.starts_with("-> create")).count(),
            MAX_ROWS
        );
        assert!(more.contains("... 1 more"), "{more}");
        assert!(more.contains("51 create"), "{more}");
    }

    #[test]
    fn moai_output_selects_quiet_or_plain() {
        use std::ffi::OsStr;
        assert_eq!(output_mode(None), OutputMode::Normal);
        assert_eq!(output_mode(Some(OsStr::new("quiet"))), OutputMode::Quiet);
        assert_eq!(output_mode(Some(OsStr::new(" plain "))), OutputMode::Plain);
        assert_eq!(output_mode(Some(OsStr::new("json"))), OutputMode::Normal);
    }

    /// #2818 audit B1 (CWE-150): control characters from the result never reach the terminal,
    /// in any cell — title target, row name, action word or note — painted or not.
    #[test]
    fn control_characters_in_the_result_are_shown_not_sent() {
        let evil = "h\u{1b}]0;PWNED\u{7}\u{1b}[2Jz\u{9b}31m\r\n\t\u{7f}";
        let result = json!({"home": format!("/tmp/{evil}"), "plan_digest": "sha256:ab", "items": [
            {"path": format!("dir/{evil}/f.rs"), "action": format!("create{evil}"),
             "reason": format!("note {evil}")}]});
        for p in [
            utf8(),
            plain(),
            Painter {
                color: true,
                unicode: true,
            },
        ] {
            let text = render("x.plan", &result, &p, 200, Some("x-apply"));
            // Remove the painter's own SGR codes; nothing else may be a control character.
            let mut rest = String::new();
            let mut chars = text.chars().peekable();
            while let Some(c) = chars.next() {
                if c == '\u{1b}' && chars.peek() == Some(&'[') {
                    for d in chars.by_ref() {
                        if d == 'm' {
                            break;
                        }
                    }
                    continue;
                }
                rest.push(c);
            }
            let bad: Vec<char> = rest
                .chars()
                .filter(|c| c.is_control() && *c != '\n')
                .collect();
            assert!(bad.is_empty(), "{bad:?} in {text:?}");
            assert!(
                text.contains("\\x1b]0;PWNED\\x07\\x1b[2Jz\\x9b31m"),
                "{text}"
            );
            // The table stays one line per row: the CR/LF in the data were made visible.
            assert_eq!(
                text.lines().filter(|l| l.contains("dir/")).count(),
                1,
                "{text}"
            );
        }
    }
}
