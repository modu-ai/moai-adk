//! `moai leader brief` (SPEC-20261124-leader-brief): the short state a leader reads when it
//! wakes, in place of `card inbox`, `card orchestrate plan` and the bodies of the notes they
//! list. It adds no judgement and no query of its own: it shrinks the answer of
//! `card orchestrate plan` to the landing queue, the blocks and the decisions, each with the
//! number of the note (or verdict event) it comes from, and keeps the whole printed answer within
//! [`LIMIT_BYTES`].
//!
//! Nothing here reads the clock, so an unchanged ledger prints the same bytes twice. The command
//! writes nothing, starts no host program and reads no terminal.

use crate::card::orchestrate;
use serde_json::{json, Value};
use std::path::Path;

/// The cap of the whole printed envelope, in bytes.
pub const LIMIT_BYTES: usize = 4096;
/// What the envelope around `result` takes: `{"protocol_version":1,"operation":"leader.brief",
/// "ok":true,"result":` and `}` and the line break, rounded up.
const ENVELOPE_BYTES: usize = 96;
/// The longest first line printed, in bytes.
pub const TEXT_BYTES: usize = 96;

/// The three lists in the order a cut looks at them when two are equally long.
const LISTS: [&str; 3] = ["blocked", "queue", "decisions"];

/// `leader brief`. Reads only; a project without a ledger has empty lists (`ledger: "absent"`,
/// exit 0) and nothing is created.
pub fn brief(project: &Path, home: &Path) -> Result<Value, String> {
    let plan = orchestrate::plan(project, home, None)?;
    Ok(shrink(&plan))
}

/// `text` cut to at most `max` bytes, on a character boundary.
pub fn cut(text: &str, max: usize) -> &str {
    if text.len() <= max {
        return text;
    }
    let mut end = max;
    while !text.is_char_boundary(end) {
        end -= 1;
    }
    &text[..end]
}

/// Whether a line says `kind=decision` as a word of its own: `kind=decisions` does not.
fn says_decision(line: &str) -> bool {
    const KEY: &str = "kind=decision";
    let mut from = 0;
    while let Some(at) = line[from..].find(KEY) {
        let start = from + at;
        let end = start + KEY.len();
        let before_ok = line[..start]
            .chars()
            .next_back()
            .is_none_or(|c| !(c.is_alphanumeric() || c == '_' || c == '-'));
        let after_ok = line[end..]
            .chars()
            .next()
            .is_none_or(|c| !(c.is_alphanumeric() || c == '_' || c == '-'));
        if before_ok && after_ok {
            return true;
        }
        from = end;
    }
    false
}

/// A note's first line without its marker word (`LANE-BLOCKED`, `DECISION-QUESTION`), cut.
fn text_of(line: &str) -> String {
    let rest = line
        .trim_start()
        .split_once(char::is_whitespace)
        .map_or("", |(_, rest)| rest.trim_start());
    cut(rest.trim_end(), TEXT_BYTES).to_string()
}

fn list_len(list: &Value) -> usize {
    list.as_array().map_or(0, Vec::len)
}

/// The brief of one plan: the plan is `card orchestrate plan`'s answer (or the empty plan of a
/// project without a ledger).
pub fn shrink(plan: &Value) -> Value {
    let mut queue: Vec<Value> = Vec::new();
    for entry in plan["land_queue"].as_array().into_iter().flatten() {
        let reference = match (entry["note"].as_i64(), entry["event"].as_i64()) {
            (Some(note), _) => format!("note:{note}"),
            (None, Some(event)) => format!("event:{event}"),
            (None, None) => continue,
        };
        let mut item = json!({
            "card": entry["card"], "head": entry["head"], "ref": reference,
        });
        if entry["current"] == json!(false) {
            item["old"] = json!(true);
        }
        queue.push(item);
    }
    let audit = plan["actions"]
        .as_array()
        .into_iter()
        .flatten()
        .filter(|a| a["kind"] == json!("audit"))
        .count();

    // (note number, item) of each unanswered block or decision.
    let mut blocked: Vec<(i64, Value)> = Vec::new();
    let mut decisions: Vec<(i64, Value)> = Vec::new();
    let mut unread: Vec<&str> = Vec::new();
    for item in plan["attention"].as_array().into_iter().flatten() {
        match item["kind"].as_str() {
            Some("note") => {
                let (Some(note), Some(family)) = (item["note_id"].as_i64(), item["family"].as_str())
                else {
                    continue;
                };
                let line = item["line"].as_str().unwrap_or_default();
                let shown = json!({
                    "card": item["card"], "note": note, "text": text_of(line),
                });
                if family == "decision_question" || (family == "blocked" && says_decision(line)) {
                    decisions.push((note, shown));
                } else if family == "blocked" {
                    blocked.push((note, shown));
                }
            }
            Some("integration") => unread.push("integration"),
            Some("registry") => unread.push("registry"),
            _ => {}
        }
    }
    unread.dedup();
    // The newest note first: a cut drops the oldest, and a leader acts on the newest.
    blocked.sort_by_key(|(note, _)| std::cmp::Reverse(*note));
    decisions.sort_by_key(|(note, _)| std::cmp::Reverse(*note));

    let counts = json!({
        "queue": queue.len(), "audit": audit,
        "blocked": blocked.len(), "decisions": decisions.len(),
    });
    let mut out = json!({
        "action": "brief",
        "project_id": plan["project_id"],
        "ledger": plan["ledger"],
        "writes": false,
        "counts": counts,
        "queue": queue,
        "blocked": blocked.into_iter().map(|(_, v)| v).collect::<Vec<_>>(),
        "decisions": decisions.into_iter().map(|(_, v)| v).collect::<Vec<_>>(),
        "omitted": {"queue": 0, "blocked": 0, "decisions": 0},
        "unread": unread,
    });
    // Drop the last item of the longest list until the answer fits.
    while out.to_string().len() + ENVELOPE_BYTES > LIMIT_BYTES {
        let mut longest: Option<&str> = None;
        for name in LISTS {
            if list_len(&out[name]) > longest.map_or(0, |l| list_len(&out[l])) {
                longest = Some(name);
            }
        }
        let Some(name) = longest else { break };
        if let Some(list) = out[name].as_array_mut() {
            list.pop();
        }
        let dropped = out["omitted"][name].as_i64().unwrap_or(0) + 1;
        out["omitted"][name] = json!(dropped);
    }
    out
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn a_cut_never_splits_a_character() {
        assert_eq!(cut("abc", 3), "abc");
        assert_eq!(cut("abcd", 3), "abc");
        // Each Korean syllable takes three bytes.
        assert_eq!(cut("가나다", 7), "가나");
        assert_eq!(cut("가나다", 2), "");
    }

    #[test]
    fn kind_decision_is_a_word_of_its_own() {
        for (line, want) in [
            ("x · kind=decision owner=leader", true),
            ("kind=decision", true),
            ("x kind=decisions owner=leader", false),
            ("x xkind=decision", false),
            ("x kind=decision-x", false),
            ("x kind=claim then kind=decision", true),
            ("x kind=decisionkind=decision", false),
            ("", false),
        ] {
            assert_eq!(says_decision(line), want, "{line:?}");
        }
    }

    #[test]
    fn the_marker_word_is_left_out_of_the_text() {
        assert_eq!(text_of("LANE-BLOCKED agent-1 C-1 wait"), "agent-1 C-1 wait");
        assert_eq!(text_of("LANE-BLOCKED"), "");
    }
}
