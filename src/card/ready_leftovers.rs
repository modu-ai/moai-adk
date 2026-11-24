//! SPEC-20261124-lane-leftovers: a hand-back note (`LANE-READY-INTEGRATION`, `REPAIRED`, `READY `)
//! carries one `leftovers:` line that names what the lane left behind, or `none`; `card note`
//! refuses the note without it. When it keeps the note it also reads the primary checkout for
//! files that appeared there since the card was reserved and puts them in its answer (observe:
//! the answer lists them, nothing is refused because of them).

use super::land::check::git_text;
use super::ready_receipt;
use std::path::Path;

/// The start of the line.
pub const PREFIX: &str = "leftovers:";

/// What a readable `leftovers:` line says.
#[derive(Debug, PartialEq, Eq)]
pub enum Leftovers {
    /// `leftovers: none`.
    None,
    /// `leftovers: <path> | <path>`.
    Paths(Vec<String>),
}

/// Why the body has no readable `leftovers:` line.
#[derive(Debug, PartialEq, Eq)]
pub enum Why {
    Missing,
    Repeated,
    Empty,
    Placeholder,
    NoneWithPath,
}

/// What the read of the primary checkout found.
#[derive(Debug, PartialEq, Eq)]
pub enum Shared {
    /// The paths (relative to the primary checkout) that are new since the reservation began.
    Checked(Vec<String>),
    /// The check could not be made, and the reason.
    NotChecked(String),
}

/// The `leftovers:` line of a note `body` (its first line is the marker line and is never read).
/// A line inside a code fence, indented four spaces or by a tab is quoted text, not the line.
pub fn read_line(body: &str) -> Result<Leftovers, Why> {
    let mut found: Vec<&str> = Vec::new();
    let mut fenced = false;
    for raw in body.lines().skip(1) {
        let line = raw.trim_start();
        if line.starts_with("```") || line.starts_with("~~~") {
            fenced = !fenced;
            continue;
        }
        let indent = &raw[..raw.len() - line.len()];
        if fenced || indent.starts_with('\t') || indent.len() >= 4 {
            continue;
        }
        let line = line.strip_prefix("- ").unwrap_or(line);
        if let Some(rest) = line.strip_prefix(PREFIX) {
            found.push(rest);
        }
    }
    let rest = match found.as_slice() {
        [] => return Err(Why::Missing),
        [one] => *one,
        _ => return Err(Why::Repeated),
    };
    let fields: Vec<&str> = rest
        .split('|')
        .map(str::trim)
        .filter(|f| !f.is_empty())
        .collect();
    if fields.is_empty() {
        return Err(Why::Empty);
    }
    // A template placeholder left as it is (`<path>`) is not a statement.
    if fields
        .iter()
        .any(|f| f.starts_with('<') && f.ends_with('>'))
    {
        return Err(Why::Placeholder);
    }
    match fields.as_slice() {
        ["none"] => Ok(Leftovers::None),
        _ if fields.contains(&"none") => Err(Why::NoneWithPath),
        _ => Ok(Leftovers::Paths(
            fields.iter().map(|f| (*f).to_string()).collect(),
        )),
    }
}

fn reason(why: &Why) -> &'static str {
    match why {
        Why::Missing => "the body has no `leftovers:` line",
        Why::Repeated => "the body has more than one `leftovers:` line",
        Why::Empty => "the `leftovers:` line is empty",
        Why::Placeholder => "the `leftovers:` line still holds a placeholder such as `<path>`",
        Why::NoneWithPath => "`none` and a path cannot stand on one `leftovers:` line",
    }
}

/// The refusal of a hand-back note whose `leftovers:` line is missing or unreadable; `Ok` for a
/// note that is no hand-back or reads correctly. A first line the receipt gate refuses for its
/// shape is that gate's to refuse, so it passes here.
pub fn check_note(card: &str, text: &str) -> Result<(), String> {
    if !matches!(ready_receipt::marker_head(text, card), Ok(Some(_))) {
        return Ok(());
    }
    match read_line(text) {
        Ok(_) => Ok(()),
        Err(why) => Err(format!(
            "card note refused: a hand-back note needs one `leftovers:` line, and {}.\n\
             Add `leftovers: none` when you left nothing, or `leftovers: <path> | <path>` for each \
             thing you could not remove. Make a throwaway copy, a temporary HOME and a `mktemp` \
             folder only under `<worktree>/target/<name>/` and remove them by a relative path \
             before READY.",
            reason(&why)
        )),
    }
}

/// Expected state of a primary checkout: ignored folders and files that a project keeps there on
/// purpose, so they are never a leftover.
fn is_expected(path: &str) -> bool {
    let first = path.split('/').next().unwrap_or("");
    let last = path.trim_end_matches('/').rsplit('/').next().unwrap_or("");
    matches!(
        first,
        "target" | ".claude" | "CLAUDE.local.md" | ".mcp.json"
    ) || last == ".DS_Store"
}

/// The untracked (`??`) and ignored (`!!`) paths of `git status --porcelain=v1 -z` output.
fn untracked_and_ignored(text: &str) -> Vec<String> {
    let mut out = Vec::new();
    let mut fields = text.split('\0');
    while let Some(entry) = fields.next() {
        let Some((xy, path)) = entry.split_at_checked(2) else {
            continue;
        };
        // A rename or copy entry is followed by its source path.
        if xy.starts_with(['R', 'C']) || xy.ends_with(['R', 'C']) {
            fields.next();
        }
        let path = path.trim_start();
        if matches!(xy, "??" | "!!") && !path.is_empty() {
            out.push(path.to_string());
        }
    }
    out
}

/// The most entries below one reported folder that [`newest_secs`] looks at.
const WALK_LIMIT: usize = 2000;

/// The newest creation or modification time of `path` (seconds since the epoch) and, for a
/// folder, of what lies below it, looking at no more than [`WALK_LIMIT`] entries.
fn newest_secs(path: &Path) -> Option<i64> {
    let mut newest: Option<i64> = None;
    let mut visited = 0usize;
    let mut stack = vec![path.to_path_buf()];
    while let Some(p) = stack.pop() {
        visited += 1;
        let Ok(meta) = std::fs::symlink_metadata(&p) else {
            continue;
        };
        for t in [meta.created().ok(), meta.modified().ok()]
            .into_iter()
            .flatten()
        {
            if let Ok(d) = t.duration_since(std::time::UNIX_EPOCH) {
                newest = newest.max(Some(d.as_secs() as i64));
            }
        }
        if meta.is_dir() && visited < WALK_LIMIT {
            if let Ok(entries) = std::fs::read_dir(&p) {
                stack.extend(entries.flatten().map(|e| e.path()));
            }
        }
    }
    newest
}

/// The files of the primary checkout of `project` that are untracked or ignored, are not expected
/// state and were created or changed at or after `since` (seconds since the epoch). Reads only:
/// git runs with `--no-optional-locks` (`git_text` adds it), so the index is not refreshed.
pub fn shared_checkout_new(project: &Path, since: i64) -> Shared {
    let listing = match git_text(project, &["worktree", "list", "--porcelain"]) {
        Ok(t) => t,
        Err(e) => return Shared::NotChecked(e),
    };
    let Some(primary) = listing.lines().find_map(|l| l.strip_prefix("worktree ")) else {
        return Shared::NotChecked("git lists no worktree".to_string());
    };
    let primary = Path::new(primary);
    let status = match git_text(
        primary,
        &[
            "status",
            "--porcelain=v1",
            "-z",
            "--ignored=matching",
            "-uall",
        ],
    ) {
        Ok(t) => t,
        Err(e) => return Shared::NotChecked(e),
    };
    let mut new: Vec<String> = untracked_and_ignored(&status)
        .into_iter()
        .filter(|p| !is_expected(p))
        .filter(|p| newest_secs(&primary.join(p.trim_end_matches('/'))).is_some_and(|t| t >= since))
        .collect();
    new.sort();
    Shared::Checked(new)
}

#[cfg(test)]
mod tests {
    use super::*;

    const FIRST: &str = "LANE-READY-INTEGRATION C-1 head=0000000000000000000000000000000000000000";

    fn body(rest: &str) -> String {
        format!("{FIRST}\n{rest}")
    }

    #[test]
    fn the_line_is_read_in_its_forms() {
        assert_eq!(read_line(&body("leftovers: none\n")), Ok(Leftovers::None));
        assert_eq!(
            read_line(&body("- leftovers: target/x/copy | target/y\n")),
            Ok(Leftovers::Paths(vec![
                "target/x/copy".to_string(),
                "target/y".to_string()
            ]))
        );
        // A line among other lines, a path with a space, and blanks around the fields.
        assert_eq!(
            read_line(&body("base=x\nleftovers:  a b  |  c \nmore\n")),
            Ok(Leftovers::Paths(vec!["a b".to_string(), "c".to_string()]))
        );
    }

    #[test]
    fn a_line_that_is_missing_or_unreadable_gives_its_reason() {
        let why = |rest: &str| read_line(&body(rest)).unwrap_err();
        assert_eq!(why("base=x\n"), Why::Missing);
        assert_eq!(why("leftovers:\n"), Why::Empty);
        assert_eq!(why("leftovers:  | \n"), Why::Empty);
        assert_eq!(why("leftovers: <path>\n"), Why::Placeholder);
        assert_eq!(why("leftovers: a | <path>\n"), Why::Placeholder);
        assert_eq!(why("leftovers: none | build/x\n"), Why::NoneWithPath);
        assert_eq!(why("leftovers: a\nleftovers: none\n"), Why::Repeated);
        assert_eq!(read_line(FIRST).unwrap_err(), Why::Missing);
    }

    #[test]
    fn a_quoted_line_is_not_a_leftovers_line() {
        let why = |rest: &str| read_line(&body(rest)).unwrap_err();
        assert_eq!(why("```\nleftovers: none\n```\n"), Why::Missing);
        assert_eq!(why("~~~\nleftovers: none\n~~~\n"), Why::Missing);
        assert_eq!(why("    leftovers: none\n"), Why::Missing);
        assert_eq!(why("\tleftovers: none\n"), Why::Missing);
        // The marker line itself is never read, even when it holds the word.
        let first_only = format!("{FIRST} leftovers: none");
        assert_eq!(read_line(&first_only).unwrap_err(), Why::Missing);
        // The first line is never read, whatever it says: a body that is only the line has none.
        assert_eq!(read_line("leftovers: none").unwrap_err(), Why::Missing);
        // A fence that closes lets the next line count again.
        assert_eq!(
            read_line(&body("```\nx\n```\nleftovers: none\n")),
            Ok(Leftovers::None)
        );
    }

    #[test]
    fn only_a_hand_back_note_is_judged() {
        assert_eq!(
            check_note("C-1", "LANE-BLOCKED kind=decision C-1\nbody"),
            Ok(())
        );
        assert_eq!(check_note("C-1", "plain note"), Ok(()));
        for marker in ["LANE-READY-INTEGRATION", "REPAIRED", "READY"] {
            let text =
                format!("{marker} C-1 head=0000000000000000000000000000000000000000\nbase=x");
            let refused = check_note("C-1", &text).unwrap_err();
            assert!(refused.contains("card note refused"), "{refused}");
            assert!(refused.contains("leftovers:"), "{refused}");
            let kept = format!("{text}\nleftovers: none");
            assert_eq!(check_note("C-1", &kept), Ok(()), "{marker}");
        }
    }

    #[test]
    fn the_refusal_names_the_reason_and_a_line_to_paste() {
        let text = format!("{FIRST}\nleftovers: <path>");
        let refused = check_note("C-1", &text).unwrap_err();
        assert!(refused.contains("placeholder"), "{refused}");
        assert!(refused.contains("leftovers: none"), "{refused}");
        assert!(refused.contains("target/"), "{refused}");
    }

    #[test]
    fn a_checkout_that_cannot_be_read_is_not_checked() {
        let gone = Path::new("/nonexistent/lane-leftovers-test-checkout");
        match shared_checkout_new(gone, 0) {
            Shared::NotChecked(why) => assert!(!why.is_empty()),
            other => panic!("expected NotChecked, got {other:?}"),
        }
    }
}
