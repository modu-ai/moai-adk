//! SPEC-20261124-ready-receipt-gate: a `LANE-READY-INTEGRATION` or `REPAIRED` note is kept only
//! when a worktree that has the note's head as its HEAD holds the PASS receipt of that head.
//!
//! The receipt is `summary.txt` in `target/check-affected-<head12>/` of that worktree: the one
//! `check-affected` summary line `moai check affected` answers, written by [`write_summary`]. The
//! gate is a check against going on without the run, not a security boundary: a person can write
//! the file by hand.

use super::land::check::git_text;
use std::fs;
use std::path::{Path, PathBuf};

type Result<T> = std::result::Result<T, String>;

/// The receipt file inside `target/check-affected-<head12>/`.
pub const SUMMARY_FILE: &str = "summary.txt";

/// What a note that hands a head back starts with, and the word it is called: `READY` needs its
/// space so that a longer word is no marker.
const MARKERS: [(&str, &str); 3] = [
    ("LANE-READY-INTEGRATION", "LANE-READY-INTEGRATION"),
    ("REPAIRED", "REPAIRED"),
    ("READY ", "READY"),
];

/// A receipt is one short line; a larger file is not one.
const MAX_RECEIPT_BYTES: u64 = 4096;

fn short(head: &str) -> &str {
    head.get(..12).unwrap_or(head)
}

fn is_head(text: &str) -> bool {
    text.len() == 40 && text.bytes().all(|b| matches!(b, b'0'..=b'9' | b'a'..=b'f'))
}

/// The hand-back marker a line starts with (SPEC-20261124-ready-receipt-followup REQ-1, REQ-2,
/// REQ-4): the one judge of the `card note` gate, `card inbox` and the status line. Whatever
/// follows the marker does not matter here; the gate then demands the exact shape. Blanks before
/// the marker are the caller's business: the gate and the inbox skip them, the status line does
/// not (its own rule, pinned by its tests).
pub fn marker_of_line(line: &str) -> Option<&'static str> {
    MARKERS
        .iter()
        .find(|(prefix, _)| line.starts_with(prefix))
        .map(|(_, word)| *word)
}

/// The head a hand-back note names: `Ok(None)` when its first line starts with no marker, the
/// head when the line reads `<marker> <card> head=<40 lowercase hex>` followed by the end of the
/// line or a space, and the refusal when it starts with a marker and reads otherwise.
pub fn marker_head(text: &str, card: &str) -> Result<Option<String>> {
    let first = text.lines().next().unwrap_or("");
    // Blanks before a marker do not hide it (`card inbox` counts such a line); the shape below is
    // read from the line as it is, so such a line is refused.
    let Some(marker) = marker_of_line(first.trim_start()) else {
        return Ok(None);
    };
    let shown: String = first.chars().take(120).collect();
    let wrong = || {
        Err(format!(
            "card note refused: a note that starts with `{marker}` must read \
             `{marker} {card} head=<40 lowercase hex>` for this card on its first line, not `{shown}`"
        ))
    };
    let Some(rest) = first.strip_prefix(marker).and_then(|r| r.strip_prefix(' ')) else {
        return wrong();
    };
    let Some((named, rest)) = rest.split_once(' ') else {
        return wrong();
    };
    let Some(rest) = rest.strip_prefix("head=") else {
        return wrong();
    };
    let Some(head) = rest.get(..40).filter(|h| is_head(h)) else {
        return wrong();
    };
    let after = &rest[40..];
    if named != card || !(after.is_empty() || after.starts_with(' ')) {
        return wrong();
    }
    Ok(Some(head.to_string()))
}

/// What a worktree's receipt folder says about a head.
#[derive(Debug, PartialEq, Eq)]
enum Receipt {
    /// A PASS summary line of the head; the base its range started from, as the line wrote it.
    Pass(String),
    /// Anything else, in words: `missing`, `result=FAIL`, `names another head`, and so on.
    Not(String),
}

fn not(why: impl Into<String>) -> Receipt {
    Receipt::Not(why.into())
}

/// The summary line `text` holds, as a receipt of `head`.
fn judge(text: &str, head: &str) -> Receipt {
    let line = text.strip_suffix('\n').unwrap_or(text);
    if line.contains('\n') {
        return not("not one line");
    }
    let words: Vec<&str> = line.split(' ').collect();
    if words.first() != Some(&"check-affected") {
        return not("not a check-affected summary line");
    }
    match words.iter().find_map(|w| w.strip_prefix("head=")) {
        Some(named) if named == short(head) => {}
        Some(named) => return not(format!("names another head ({named})")),
        None => return not("names no head"),
    }
    // A `--lane` run checked a part of what the head changed (SPEC-20261124-ready-receipt-followup).
    if words.contains(&"selection=lane") {
        return not(
            "a --lane run (selection=lane) checks only a part; run the bare `moai check affected`",
        );
    }
    let base = words.iter().find_map(|w| w.strip_prefix("base="));
    match words.last().and_then(|w| w.strip_prefix("result=")) {
        Some("PASS") => match base {
            Some(b) if (12..=40).contains(&b.len()) && b.bytes().all(|c| c.is_ascii_hexdigit()) => {
                Receipt::Pass(b.to_string())
            }
            Some(b) => not(format!("names a base that is no commit id ({b})")),
            None => not("names no base"),
        },
        Some(other) => not(format!("result={other}")),
        None => not("has no result word"),
    }
}

fn read_receipt(path: &Path, head: &str) -> Receipt {
    let meta = match fs::symlink_metadata(path) {
        Ok(m) => m,
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => return not("missing"),
        Err(e) => return not(format!("unreadable ({e})")),
    };
    if !meta.is_file() {
        return not("not a regular file");
    }
    if meta.len() > MAX_RECEIPT_BYTES {
        return not("larger than a receipt line");
    }
    match fs::read(path) {
        Ok(bytes) => match String::from_utf8(bytes) {
            Ok(text) => judge(&text, head),
            Err(_) => not("not UTF-8 text"),
        },
        Err(e) => not(format!("unreadable ({e})")),
    }
}

/// The worktrees of `project`'s repository that have `head` as their HEAD, by `git worktree list
/// --porcelain`.
fn worktrees_at(project: &Path, head: &str) -> Result<Vec<PathBuf>> {
    let text = git_text(project, &["worktree", "list", "--porcelain"]).map_err(|e| {
        format!(
            "card note refused: cannot list the worktrees of {}: {e}",
            project.display()
        )
    })?;
    let mut found = Vec::new();
    for block in text.split("\n\n") {
        let path = block.lines().find_map(|l| l.strip_prefix("worktree "));
        let at = block.lines().find_map(|l| l.strip_prefix("HEAD "));
        if let (Some(path), Some(at)) = (path, at) {
            if at == head {
                found.push(PathBuf::from(path));
            }
        }
    }
    Ok(found)
}

/// The refusal of a hand-back note whose head has no PASS receipt, or `Ok` when one holds it.
/// `text` is the note; a note that starts with no marker is not judged.
pub fn check_note(project: &Path, card: &str, text: &str) -> Result<()> {
    let Some(head) = marker_head(text, card)? else {
        return Ok(());
    };
    let mut seen = Vec::new();
    let worktrees = worktrees_at(project, &head)?;
    for wt in &worktrees {
        let path = wt
            .join("target")
            .join(format!("check-affected-{}", short(&head)))
            .join(SUMMARY_FILE);
        match read_receipt(&path, &head) {
            // The range of the run began at a commit the head grew from.
            Receipt::Pass(base) => {
                if git_text(project, &["merge-base", "--is-ancestor", &base, &head]).is_ok() {
                    return Ok(());
                }
                seen.push(format!(
                    "  - {}: base {base} is not an ancestor of the head (or not a commit of this repository)",
                    path.display()
                ));
            }
            Receipt::Not(why) => seen.push(format!("  - {}: {why}", path.display())),
        }
    }
    if worktrees.is_empty() {
        seen.push(format!(
            "  - no worktree of {} has {head} as its HEAD (`git worktree list`)",
            project.display()
        ));
    }
    Err(format!(
        "card note refused: head {head} has no PASS check-affected receipt:\n{}\n\
         Run `moai check affected <base>..{head}` (bare, with every tracked file committed) in a \
         worktree that is at that head until it ends in `result=PASS`, then write the note again.",
        seen.join("\n")
    ))
}

/// Writes the summary line of a finished `moai check affected` run to `<out>/summary.txt`.
pub fn write_summary(out: &Path, summary: &str) -> Result<()> {
    let path = out.join(SUMMARY_FILE);
    let refused = |e: std::io::Error| {
        format!(
            "check affected refused: cannot write {}: {e}",
            path.display()
        )
    };
    // A link at the receipt's place would send the write somewhere else.
    if fs::symlink_metadata(&path).is_ok_and(|m| m.file_type().is_symlink()) {
        fs::remove_file(&path).map_err(refused)?;
    }
    fs::write(&path, format!("{summary}\n")).map_err(refused)
}

#[cfg(test)]
mod tests {
    use super::*;

    const HEAD: &str = "0123456789abcdef0123456789abcdef01234567";

    #[test]
    fn the_marker_line_is_read_or_refused() {
        let ok = |t: &str| marker_head(t, "C-1");
        assert_eq!(
            ok(&format!("LANE-READY-INTEGRATION C-1 head={HEAD}\nbase=x")),
            Ok(Some(HEAD.to_string()))
        );
        assert_eq!(
            ok(&format!("REPAIRED C-1 head={HEAD} more")),
            Ok(Some(HEAD.to_string()))
        );
        assert_eq!(ok("LANE-BLOCKED kind=decision C-1"), Ok(None));
        assert_eq!(ok(&format!("x\nREPAIRED C-1 head={HEAD}")), Ok(None));
        assert_eq!(
            ok(&format!("READY C-1 head={HEAD}")),
            Ok(Some(HEAD.to_string()))
        );
        assert_eq!(ok(&format!("READYING C-1 head={HEAD}")), Ok(None));
        assert_eq!(ok(&format!("MERGE-READY C-1 head={HEAD}")), Ok(None));
        for bad in [
            format!("REPAIRED-X C-1 head={HEAD}"),
            format!("LANE-READY-INTEGRATION: C-1 head={HEAD}"),
            format!(" LANE-READY-INTEGRATION C-1 head={HEAD}"),
            format!("READY (head {})", &HEAD[..7]),
            format!("READY C-1 head={}", &HEAD[..12]),
            "REPAIRED".to_string(),
            "REPAIRED C-1".to_string(),
            format!("REPAIRED C-2 head={HEAD}"),
            format!("REPAIRED C-1 head={}", &HEAD[..39]),
            format!("REPAIRED C-1 head={HEAD}0"),
            format!("REPAIRED C-1 head={}", HEAD.to_uppercase()),
            format!("REPAIRED  C-1 head={HEAD}"),
        ] {
            assert!(ok(&bad).is_err(), "{bad}");
        }
    }

    /// The one judge the gate, `card inbox` and the status line share (SPEC-20261124-ready-
    /// receipt-followup REQ-1, REQ-2, REQ-4): a first line that starts with a marker is that
    /// marker, whatever follows; `READY` needs its space; a leading blank is no marker here.
    #[test]
    fn marker_of_line_table() {
        for (line, want) in [
            (
                "LANE-READY-INTEGRATION C-1 head=abc",
                Some("LANE-READY-INTEGRATION"),
            ),
            (
                "LANE-READY-INTEGRATION: C-1",
                Some("LANE-READY-INTEGRATION"),
            ),
            ("LANE-READY-INTEGRATIONC-1", Some("LANE-READY-INTEGRATION")),
            // Blanks before the marker are the caller's business (see the doc of the function).
            ("  LANE-READY-INTEGRATION C-1", None),
            ("LANE-READY-INTEGRATION", Some("LANE-READY-INTEGRATION")),
            ("REPAIRED C-1 head=abc", Some("REPAIRED")),
            ("REPAIRED-X C-1", Some("REPAIRED")),
            ("READY C-1 head=abc", Some("READY")),
            ("READY (head abc)", Some("READY")),
            ("READY", None),
            ("READYING C-1", None),
            ("MERGE-READY C-1", None),
            ("XREADY C-1", None),
            ("see REPAIRED in the middle", None),
            ("repaired in lower case", None),
            ("LANE-BLOCKED kind=decision", None),
            ("", None),
        ] {
            assert_eq!(marker_of_line(line), want, "{line:?}");
        }
    }

    #[test]
    fn a_lane_summary_line_is_no_receipt() {
        let own = short(HEAD);
        let line = |result: &str, sel: &str| {
            format!("check-affected base=aaaaaaaaaaaa head={own} targets=1 scans=none{sel} test-threads=4 test=0 {result}\n")
        };
        assert_eq!(
            judge(&line("result=PASS", ""), HEAD),
            Receipt::Pass("aaaaaaaaaaaa".to_string())
        );
        let lane = judge(&line("result=PASS", " selection=lane"), HEAD);
        match lane {
            Receipt::Not(why) => {
                assert!(why.contains("selection=lane"), "{why}");
                assert!(why.contains("--lane"), "{why}");
            }
            other => panic!("a lane run was taken as a receipt: {other:?}"),
        }
    }

    /// SPEC-20261124-card-select-batch-land REQ-4, REQ-5: a `--card` run is a receipt, a `--lane`
    /// run still is not, and a `--card` run that fell back to the full set writes the bare line.
    #[test]
    fn a_card_summary_line_is_a_receipt_and_a_lane_one_is_not() {
        let own = short(HEAD);
        let line = |sel: &str, result: &str| {
            format!("check-affected base=aaaaaaaaaaaa head={own} targets=7 scans=listed:1{sel} test-threads=4 test=0 {result}\n")
        };
        assert_eq!(
            judge(&line(" selection=card", "result=PASS"), HEAD),
            Receipt::Pass("aaaaaaaaaaaa".to_string())
        );
        assert_eq!(
            judge(&line("", "result=PASS"), HEAD),
            Receipt::Pass("aaaaaaaaaaaa".to_string())
        );
        assert_eq!(
            judge(&line(" selection=card", "result=FAIL"), HEAD),
            not("result=FAIL")
        );
        assert!(matches!(
            judge(&line(" selection=lane", "result=PASS"), HEAD),
            Receipt::Not(_)
        ));
    }

    /// SPEC-20261125-check-affected-card-areas REQ-12: the `card_by=` word that a card run writes
    /// after `selection=card` changes nothing for the gate (it reads `result=` and `selection=`).
    #[test]
    fn a_card_summary_line_with_its_reasons_is_a_receipt_and_a_lane_one_is_not() {
        let own = short(HEAD);
        let line = |sel: &str, result: &str| {
            format!("check-affected base=aaaaaaaaaaaa head={own} targets=7 scans=listed:1{sel} test-threads=4 test=0 {result}\n")
        };
        let by = "card_by=area:card,module:card,units:card::work";
        assert_eq!(
            judge(&line(&format!(" selection=card {by}"), "result=PASS"), HEAD),
            Receipt::Pass("aaaaaaaaaaaa".to_string())
        );
        assert_eq!(
            judge(&line(&format!(" selection=card {by}"), "result=FAIL"), HEAD),
            not("result=FAIL")
        );
        assert!(matches!(
            judge(&line(&format!(" selection=lane {by}"), "result=PASS"), HEAD),
            Receipt::Not(_)
        ));
    }

    #[test]
    fn a_summary_line_is_a_receipt_only_for_its_own_head_and_a_pass() {
        let line = |head: &str, result: &str| {
            format!("check-affected base=aaaaaaaaaaaa head={head} targets=1 test=0 {result}\n")
        };
        let own = short(HEAD);
        assert_eq!(
            judge(&line(own, "result=PASS"), HEAD),
            Receipt::Pass("aaaaaaaaaaaa".to_string())
        );
        assert_eq!(
            judge(&format!("check-affected head={own} result=PASS"), HEAD),
            not("names no base")
        );
        assert_eq!(
            judge(
                &format!("check-affected base=--all head={own} result=PASS"),
                HEAD
            ),
            not("names a base that is no commit id (--all)")
        );
        assert_eq!(judge(&line(own, "result=FAIL"), HEAD), not("result=FAIL"));
        assert_eq!(
            judge(&line("bbbbbbbbbbbb", "result=PASS"), HEAD),
            not("names another head (bbbbbbbbbbbb)")
        );
        assert_eq!(
            judge(&format!("{}x\n", line(own, "result=PASS")), HEAD),
            not("not one line")
        );
        assert_eq!(
            judge("result=PASS\n", HEAD),
            not("not a check-affected summary line")
        );
        assert_eq!(
            judge(&format!("check-affected head={own} done"), HEAD),
            not("has no result word")
        );
    }

    #[test]
    fn write_summary_names_the_path_and_the_reason() {
        let dir = tempfile::tempdir().unwrap();
        // The receipt's place is a folder: the write cannot happen.
        fs::create_dir(dir.path().join(SUMMARY_FILE)).unwrap();
        let err = write_summary(dir.path(), "check-affected x").unwrap_err();
        assert!(
            err.starts_with("check affected refused: cannot write "),
            "{err}"
        );
        assert!(
            err.contains(dir.path().join(SUMMARY_FILE).to_str().unwrap()),
            "{err}"
        );
        // A free place gets the line and one line break.
        let free = tempfile::tempdir().unwrap();
        write_summary(free.path(), "check-affected y").unwrap();
        assert_eq!(
            fs::read_to_string(free.path().join(SUMMARY_FILE)).unwrap(),
            "check-affected y\n"
        );
    }
}
