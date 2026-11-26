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
             thing you could not remove. Make a throwaway copy, a scratch HOME and a `mktemp` \
             folder only under `<worktree>/target/<name>/` and remove them by a relative path \
             before READY.",
            reason(&why)
        )),
    }
}

/// The most paths one answer lists; the rest are counted in a last entry.
pub const LIST_LIMIT: usize = 200;

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
/// folder, of what lies below it, looking at no more than `limit` entries (the folder itself
/// counts as one). The second value is how many entries it looked at.
fn newest_secs_within(path: &Path, limit: usize) -> (Option<i64>, usize) {
    let mut newest: Option<i64> = None;
    let mut visited = 0usize;
    let mut stack = vec![path.to_path_buf()];
    while visited < limit {
        let Some(p) = stack.pop() else {
            break;
        };
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
        if meta.is_dir() {
            if let Ok(entries) = std::fs::read_dir(&p) {
                stack.extend(entries.flatten().map(|e| e.path()));
            }
        }
    }
    (newest, visited)
}

fn newest_secs(path: &Path) -> Option<i64> {
    newest_secs_within(path, WALK_LIMIT).0
}

/// Whether something last touched at `newest` is new since the reservation began at `since`
/// (a file made in the same second counts).
fn is_new_since(newest: Option<i64>, since: i64) -> bool {
    newest.is_some_and(|t| t >= since)
}

/// `git status` of the primary checkout. Success with anything on stderr (an unreadable folder
/// is a warning, not a failure) is an incomplete read, so it is an `Err`.
fn status_text(primary: &Path) -> Result<String, String> {
    let program = std::env::var("MOAI_GIT_BIN").unwrap_or_else(|_| "git".to_string());
    let mut cmd = std::process::Command::new(program);
    crate::identity::scrub_git_env(&mut cmd);
    let out = cmd
        .args(["--no-optional-locks", "-C"])
        .arg(primary)
        .args([
            "status",
            "--porcelain=v1",
            "-z",
            "--ignored=matching",
            "-uall",
        ])
        .stdin(std::process::Stdio::null())
        .output()
        .map_err(|e| format!("git could not run: {e}"))?;
    status_outcome(out.status.success(), &out.stdout, &out.stderr)
}

/// The text of a finished `git status`, or why it cannot be trusted.
fn status_outcome(success: bool, stdout: &[u8], stderr: &[u8]) -> Result<String, String> {
    let err = String::from_utf8_lossy(stderr);
    let err = err.trim();
    if !success {
        return Err(format!("git status failed: {err}"));
    }
    if !err.is_empty() {
        let first = err.lines().next().unwrap_or("");
        return Err(format!("git status could not read everything: {first}"));
    }
    Ok(String::from_utf8_lossy(stdout).into_owned())
}

/// At most [`LIST_LIMIT`] of the sorted `paths`; a longer list ends in one entry that counts the
/// rest, so the answer stays small whatever a lane left behind.
fn limited(mut paths: Vec<String>) -> Vec<String> {
    if paths.len() > LIST_LIMIT {
        let more = paths.len() - LIST_LIMIT;
        paths.truncate(LIST_LIMIT);
        paths.push(format!("... and {more} more (the list shows {LIST_LIMIT})"));
    }
    paths
}

/// The files of the primary checkout of `project` that are untracked or ignored, are not expected
/// state and were created or changed at or after `since` (seconds since the epoch). Reads only:
/// git runs with `--no-optional-locks`, so the index is not refreshed.
pub fn shared_checkout_new(project: &Path, since: i64) -> Shared {
    let listing = match git_text(project, &["worktree", "list", "--porcelain"]) {
        Ok(t) => t,
        Err(e) => return Shared::NotChecked(e),
    };
    let Some(primary) = listing.lines().find_map(|l| l.strip_prefix("worktree ")) else {
        return Shared::NotChecked("git lists no worktree".to_string());
    };
    let primary = Path::new(primary);
    let status = match status_text(primary) {
        Ok(t) => t,
        Err(e) => return Shared::NotChecked(e),
    };
    let mut new: Vec<String> = untracked_and_ignored(&status)
        .into_iter()
        .filter(|p| !is_expected(p))
        .filter(|p| is_new_since(newest_secs(&primary.join(p.trim_end_matches('/'))), since))
        .collect();
    new.sort();
    Shared::Checked(limited(new))
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
    fn every_refusal_is_an_input_refusal_that_a_resend_does_not_change() {
        use crate::contract::{classify, Kind};
        let bodies = [
            ("", Why::Missing),
            ("leftovers:\n", Why::Empty),
            ("leftovers: <path>\n", Why::Placeholder),
            ("leftovers: none | x\n", Why::NoneWithPath),
            ("leftovers: a\nleftovers: b\n", Why::Repeated),
        ];
        for (rest, why) in bodies {
            let refused = check_note("C-1", &body(rest)).unwrap_err();
            assert!(refused.contains(reason(&why)), "{refused}");
            let kind = classify(&refused);
            assert_eq!(kind, Kind::InvalidInput, "{rest:?}: {refused}");
            assert!(!kind.retryable(), "{rest:?}");
        }
    }

    #[test]
    fn expected_state_is_never_a_leftover() {
        for path in [
            "target/",
            "target/debug/x.o",
            ".claude/worktrees/other/file.txt",
            ".claude/settings.local.json",
            "CLAUDE.local.md",
            ".mcp.json",
            ".DS_Store",
            "sub/.DS_Store",
            "sub/.DS_Store/",
        ] {
            assert!(is_expected(path), "{path}");
        }
        for path in [
            "new.txt",
            "dir/new.txt",
            "targets/x",
            "claude/x",
            "sub/CLAUDE.local.md",
            "sub/.mcp.json",
            ".DS_Store.bak",
        ] {
            assert!(!is_expected(path), "{path}");
        }
    }

    #[test]
    fn a_touch_in_the_second_the_card_was_reserved_is_new() {
        assert!(!is_new_since(None, 100));
        assert!(!is_new_since(Some(99), 100));
        assert!(is_new_since(Some(100), 100));
        assert!(is_new_since(Some(101), 100));
    }

    #[test]
    fn the_walk_looks_at_no_more_than_its_limit() {
        let dir = tempfile::tempdir().unwrap();
        let root = dir.path().join("keep");
        std::fs::create_dir_all(root.join("sub")).unwrap();
        for f in ["a", "b", "sub/c"] {
            std::fs::write(root.join(f), "x").unwrap();
        }
        // The folder, `sub`, and three files: five entries.
        let total = 5;
        assert_eq!(newest_secs_within(&root, total).1, total, "N");
        assert_eq!(newest_secs_within(&root, total + 1).1, total, "N+1 limit");
        assert_eq!(
            newest_secs_within(&root, total - 1).1,
            total - 1,
            "N-1 limit"
        );
        assert_eq!(newest_secs_within(&root, 1).1, 1);
        assert_eq!(newest_secs_within(&root, 0), (None, 0), "empty limit");
        // Looking at the whole folder finds a time; a file alone is one entry.
        assert!(newest_secs_within(&root, total).0.is_some());
        assert_eq!(newest_secs_within(&root.join("a"), total).1, 1);
        // A path that is not there has an entry that cannot be read and no time.
        assert_eq!(newest_secs_within(&root.join("gone"), total), (None, 1));
    }

    #[test]
    fn a_walk_goes_below_the_folder_it_is_given() {
        let dir = tempfile::tempdir().unwrap();
        let root = dir.path().join("keep");
        std::fs::create_dir_all(root.join("sub")).unwrap();
        let file = std::fs::File::create(root.join("sub/new.bin")).unwrap();
        // A file stamped in the future is the newest thing under the folder.
        let future = std::time::SystemTime::now() + std::time::Duration::from_secs(10_000);
        file.set_modified(future).unwrap();
        let want = future
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap()
            .as_secs() as i64;
        assert_eq!(newest_secs(&root), Some(want));
    }

    #[test]
    fn a_long_list_is_cut_and_the_rest_counted() {
        let paths = |n: usize| (0..n).map(|i| format!("p{i:05}")).collect::<Vec<_>>();
        assert_eq!(limited(paths(0)), Vec::<String>::new());
        assert_eq!(limited(paths(LIST_LIMIT)), paths(LIST_LIMIT), "N");
        let cut = limited(paths(LIST_LIMIT + 1));
        assert_eq!(cut.len(), LIST_LIMIT + 1, "N+1");
        assert_eq!(cut[..LIST_LIMIT], paths(LIST_LIMIT)[..]);
        assert!(cut[LIST_LIMIT].contains("1 more"), "{}", cut[LIST_LIMIT]);
        let big = limited(paths(60_000));
        assert_eq!(big.len(), LIST_LIMIT + 1);
        assert!(
            big[LIST_LIMIT].contains("59800 more"),
            "{}",
            big[LIST_LIMIT]
        );
    }

    #[test]
    fn a_status_that_wrote_a_warning_is_not_trusted() {
        assert_eq!(
            status_outcome(true, b"?? a\0", b""),
            Ok("?? a\0".to_string())
        );
        assert_eq!(
            status_outcome(true, b"?? a\0", b" \n"),
            Ok("?? a\0".to_string())
        );
        let warned = status_outcome(
            true,
            b"?? a\0",
            b"warning: could not open directory 'x/': Permission denied\n",
        )
        .unwrap_err();
        assert!(warned.contains("Permission denied"), "{warned}");
        let failed = status_outcome(false, b"", b"fatal: bad\n").unwrap_err();
        assert!(failed.contains("fatal: bad"), "{failed}");
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
