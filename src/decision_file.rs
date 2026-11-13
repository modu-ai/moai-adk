//! Decision bundle files (Issue #2619): one Markdown file under `docs/decisions/` holds a
//! table of open or answered decisions, each with the fields the skill names (SKILL.md: what,
//! options, default, reversible, category, deadline, decider). This module only reads: it
//! checks a file's shape and, given a time, sorts unanswered rows into "the default now
//! applies" and "still on hold". Who decides is never computed here — a row names it.
//!
//! The table is the first Markdown table in the file whose header row is exactly
//! [`COLUMNS`]. A cell holds no `|`. `options` lists `<letter>: <text>` items separated by
//! `; `, and `default` and a non-empty `answer` name one of those letters.

use serde::Serialize;

/// The header of a decision table, in order.
pub const COLUMNS: [&str; 10] = [
    "id",
    "what",
    "options",
    "default",
    "reversible",
    "category",
    "deadline",
    "decider",
    "answer",
    "evidence",
];

/// `none`, `guarded` (permissions, keys, the real HOME and global settings, deletions,
/// repository protection settings — reversible, yet the user's) and the five HOLD categories
/// (SKILL.md, #2619; leader decision 6840688551 on PR #2620).
pub const CATEGORIES: [&str; 7] = [
    "none",
    "guarded",
    "money",
    "legal",
    "external_contract",
    "secret",
    "real_data",
];

#[derive(Debug, Clone, PartialEq, Serialize)]
pub struct Row {
    pub line: usize,
    pub id: String,
    pub what: String,
    pub options: Vec<String>,
    pub default: String,
    pub reversible: bool,
    pub category: String,
    pub deadline: String,
    pub decider: String,
    pub answer: String,
    pub evidence: String,
}

/// One reason a file is refused, with the 1-based line it points at.
#[derive(Debug, Clone, PartialEq, Serialize)]
pub struct Problem {
    pub line: usize,
    pub id: String,
    pub reason: String,
}

/// What an unanswered row's deadline means at a given time.
#[derive(Debug, Clone, PartialEq, Serialize)]
#[serde(rename_all = "snake_case")]
pub enum Due {
    /// Before the deadline: nothing happens yet.
    Open,
    /// Past the deadline, reversible and outside the five categories, handed to the user:
    /// the default applies.
    DefaultApplies,
    /// Past the deadline but one-way or in a HOLD category: it keeps waiting for the user.
    Hold,
}

/// Unix seconds of a `YYYY-MM-DDTHH:MM:SSZ` deadline, only when the text is exactly the
/// canonical form of that instant: a wrong separator or a day the month lacks
/// (`2026-02-31`) does not round-trip and is refused (audit note on PR #2623).
pub fn deadline_secs(text: &str) -> Option<i64> {
    let secs = crate::forge::policy::parse_utc(text)?;
    (crate::baseline::utc_from_epoch(secs) == text).then_some(secs)
}

fn cells(line: &str) -> Option<Vec<String>> {
    let t = line.trim();
    let inner = t.strip_prefix('|')?.strip_suffix('|')?;
    Some(inner.split('|').map(|c| c.trim().to_string()).collect())
}

fn is_rule(cells: &[String]) -> bool {
    cells
        .iter()
        .all(|c| !c.is_empty() && c.chars().all(|ch| matches!(ch, '-' | ':')))
}

/// Option letters of `A: text; B: text`, or `None` when an item has no `<letter>:` prefix.
fn option_letters(options: &str) -> Option<Vec<String>> {
    options
        .split("; ")
        .map(|item| {
            let (key, text) = item.split_once(':')?;
            let key = key.trim();
            (!key.is_empty() && key.chars().all(|c| c.is_alphanumeric()) && !text.trim().is_empty())
                .then(|| key.to_string())
        })
        .collect()
}

/// The decision table of `text`, and every problem that refuses it. An empty problem list
/// means the file is well formed.
pub fn parse(text: &str) -> (Vec<Row>, Vec<Problem>) {
    let mut rows = Vec::new();
    let mut problems = Vec::new();
    // A fenced block (``` or ~~~) is an example, never the table: its lines read as blank,
    // keeping every other line at its own number (Codex review on #2623).
    let mut fence: Option<&str> = None;
    let lines: Vec<&str> = text
        .lines()
        .map(|l| {
            let t = l.trim_start();
            let marker = ["```", "~~~"].into_iter().find(|m| t.starts_with(m));
            match (fence, marker) {
                (None, Some(m)) => {
                    fence = Some(m);
                    ""
                }
                (Some(open), Some(m)) if m == open => {
                    fence = None;
                    ""
                }
                (Some(_), _) => "",
                (None, None) => l,
            }
        })
        .collect();
    let Some(header) = lines.iter().position(|l| {
        cells(l).is_some_and(|c| c.iter().map(String::as_str).eq(COLUMNS.iter().copied()))
    }) else {
        problems.push(Problem {
            line: 0,
            id: String::new(),
            reason: format!("no table with the header | {} |", COLUMNS.join(" | ")),
        });
        return (rows, problems);
    };
    if !lines
        .get(header + 1)
        .and_then(|l| cells(l))
        .is_some_and(|c| c.len() == COLUMNS.len() && is_rule(&c))
    {
        problems.push(Problem {
            line: header + 2,
            id: String::new(),
            reason: "the header is not followed by its separator row".into(),
        });
        return (rows, problems);
    }
    let mut seen = std::collections::BTreeSet::new();
    for (i, raw) in lines.iter().enumerate().skip(header + 2) {
        let line = i + 1;
        if !raw.trim_start().starts_with('|') {
            break;
        }
        // A row that starts as one but lacks its closing pipe is refused, not the table's end.
        let Some(c) = cells(raw) else {
            problems.push(Problem {
                line,
                id: String::new(),
                reason: "a row does not end with `|`".into(),
            });
            continue;
        };
        let id = c.first().cloned().unwrap_or_default();
        let mut bad = |reason: String| {
            problems.push(Problem {
                line,
                id: id.clone(),
                reason,
            })
        };
        if c.len() != COLUMNS.len() {
            bad(format!("{} cells, expected {}", c.len(), COLUMNS.len()));
            continue;
        }
        for (name, value) in COLUMNS.iter().zip(&c) {
            if value.is_empty() && *name != "answer" {
                bad(format!("`{name}` is empty"));
            }
        }
        if !id.is_empty() && !seen.insert(id.clone()) {
            bad(format!("id `{id}` repeats"));
        }
        let letters = option_letters(&c[2]);
        if letters.as_ref().is_none_or(|l| l.len() < 2) {
            bad("`options` needs two or more `<letter>: <text>` items separated by `; `".into());
        }
        let letters = letters.unwrap_or_default();
        if letters.len()
            != letters
                .iter()
                .collect::<std::collections::BTreeSet<_>>()
                .len()
        {
            bad("`options` repeats a letter".into());
        }
        if !c[3].is_empty() && !letters.contains(&c[3]) {
            bad(format!("`default` `{}` is not one of the options", c[3]));
        }
        let reversible = match c[4].as_str() {
            "yes" => true,
            "no" => false,
            other => {
                bad(format!("`reversible` must be yes or no, not `{other}`"));
                true
            }
        };
        if !CATEGORIES.contains(&c[5].as_str()) {
            bad(format!(
                "`category` `{}` is not one of {}",
                c[5],
                CATEGORIES.join(", ")
            ));
        }
        if deadline_secs(&c[6]).is_none() {
            bad(format!("`deadline` `{}` is not YYYY-MM-DDTHH:MM:SSZ", c[6]));
        }
        match c[7].as_str() {
            "leader" | "user" => {}
            other => bad(format!("`decider` must be leader or user, not `{other}`")),
        }
        // A one-way or five-category choice belongs to the user (SKILL.md case 2).
        if (!reversible || c[5] != "none") && c[7] == "leader" {
            bad("a one-way or HOLD-category decision has `decider` user".into());
        }
        // A leader decision is recorded with its answer; there is nothing to wait for.
        if c[7] == "leader" && c[8].is_empty() {
            bad("a leader decision is recorded with its `answer`".into());
        }
        if !c[8].is_empty() && !letters.contains(&c[8]) {
            bad(format!("`answer` `{}` is not one of the options", c[8]));
        }
        rows.push(Row {
            line,
            id: id.clone(),
            what: c[1].clone(),
            options: letters,
            default: c[3].clone(),
            reversible,
            category: c[5].clone(),
            deadline: c[6].clone(),
            decider: c[7].clone(),
            answer: c[8].clone(),
            evidence: c[9].clone(),
        });
    }
    if rows.is_empty() && problems.is_empty() {
        problems.push(Problem {
            line: header + 1,
            id: String::new(),
            reason: "the table has no decision row".into(),
        });
    }
    (rows, problems)
}

/// What `row` means at `now` (Unix seconds). An answered row is never due.
pub fn due(row: &Row, now: i64) -> Option<Due> {
    if !row.answer.is_empty() {
        return None;
    }
    let deadline = deadline_secs(&row.deadline)?;
    Some(if now < deadline {
        Due::Open
    } else if row.reversible && row.category == "none" && row.decider == "user" {
        Due::DefaultApplies
    } else {
        Due::Hold
    })
}

#[cfg(test)]
mod tests {
    use super::*;

    const HEAD: &str = "| id | what | options | default | reversible | category | deadline | decider | answer | evidence |\n| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |\n";

    fn file(rows: &[&str]) -> String {
        format!("# Decisions\n\n{HEAD}{}\n", rows.join("\n"))
    }

    const LEADER: &str = "| D1 | PR limit | A: 8; B: 4 | A | yes | none | 2026-11-14T00:00:00Z | leader | A | #1975 c1 |";
    const TO_USER: &str = "| D2 | wake ceiling | A: 1800 s; B: 300 s | B | yes | none | 2026-11-14T00:00:00Z | user |  | #1975 c2 |";
    const HOLD: &str = "| D3 | deploy docs | A: now; B: later | B | no | external_contract | 2026-11-14T00:00:00Z | user |  | #2298 |";

    #[test]
    fn a_well_formed_bundle_parses_without_problems() {
        let (rows, problems) = parse(&file(&[LEADER, TO_USER, HOLD]));
        assert!(problems.is_empty(), "{problems:?}");
        assert_eq!(rows.len(), 3);
        assert_eq!(rows[0].options, ["A", "B"]);
        assert_eq!(rows[2].category, "external_contract");
        assert!(!rows[2].reversible);
    }

    fn refused(row: &str) -> Vec<String> {
        parse(&file(&[row]))
            .1
            .into_iter()
            .map(|p| p.reason)
            .collect()
    }

    #[test]
    fn every_required_field_is_refused_when_empty() {
        for (i, name) in COLUMNS.iter().enumerate() {
            if *name == "answer" {
                continue;
            }
            let mut c: Vec<&str> = TO_USER
                .trim_matches('|')
                .split('|')
                .map(str::trim)
                .collect();
            c[i] = "";
            let row = format!("| {} |", c.join(" | "));
            let reasons = refused(&row);
            assert!(
                reasons.iter().any(|r| r == &format!("`{name}` is empty")),
                "{name}: {reasons:?}"
            );
        }
    }

    #[test]
    fn one_way_or_category_rows_must_be_the_users() {
        let one_way = LEADER.replace("| yes | none |", "| no | none |");
        assert!(refused(&one_way)
            .contains(&"a one-way or HOLD-category decision has `decider` user".to_string()));
        for category in ["money", "guarded"] {
            let row = LEADER.replace("| yes | none |", &format!("| yes | {category} |"));
            assert!(
                refused(&row).contains(
                    &"a one-way or HOLD-category decision has `decider` user".to_string()
                ),
                "{category}"
            );
        }
    }

    #[test]
    fn categories_options_answers_and_deadlines_are_checked() {
        let eighth = TO_USER.replace("| none |", "| reputation |");
        assert!(refused(&eighth)[0].starts_with("`category` `reputation`"));
        let one_option = TO_USER.replace("A: 1800 s; B: 300 s", "A: 1800 s");
        assert!(refused(&one_option)[0].starts_with("`options` needs two or more"));
        let wrong_default = TO_USER.replace("| B | yes |", "| C | yes |");
        assert!(refused(&wrong_default)[0].starts_with("`default` `C`"));
        let wrong_answer = LEADER.replace("| leader | A |", "| leader | Z |");
        assert!(refused(&wrong_answer)[0].starts_with("`answer` `Z`"));
        let unanswered_leader = LEADER.replace("| leader | A |", "| leader |  |");
        assert!(refused(&unanswered_leader)
            .contains(&"a leader decision is recorded with its `answer`".to_string()));
        for bad_deadline in ["2026-11-14", "2026-11-14T00-00-00Z", "2026-02-31T00:00:00Z"] {
            let row = TO_USER.replace("2026-11-14T00:00:00Z", bad_deadline);
            assert!(
                refused(&row)[0].starts_with(&format!("`deadline` `{bad_deadline}`")),
                "{bad_deadline}"
            );
        }
        let repeated = TO_USER.replace("A: 1800 s; B: 300 s", "A: 1800 s; A: 300 s");
        assert!(refused(&repeated).contains(&"`options` repeats a letter".to_string()));
        let date_only = TO_USER.replace("2026-11-14T00:00:00Z", "2026-11-14");
        assert!(refused(&date_only)[0].starts_with("`deadline` `2026-11-14`"));
        let maybe = TO_USER.replace("| yes | none |", "| maybe | none |");
        assert!(refused(&maybe)[0].starts_with("`reversible` must be yes or no"));
    }

    #[test]
    fn a_broken_row_is_refused_and_a_fenced_example_is_never_the_table() {
        let broken = LEADER.trim_end_matches('|').to_string();
        let (rows, problems) = parse(&file(&[TO_USER, &broken, HOLD]));
        assert_eq!(rows.len(), 2, "the row after the broken one is still read");
        assert_eq!(problems.len(), 1, "{problems:?}");
        assert_eq!(problems[0].reason, "a row does not end with `|`");
        assert_eq!(problems[0].line, 6);
        for fence in ["```", "~~~"] {
            let example = format!(
                "# Decisions\n\n{fence}markdown\n{HEAD}| X | x | A: a; B: b | Z | maybe | none | soon | nobody |  | - |\n{fence}\n\n{HEAD}{TO_USER}\n"
            );
            let (rows, problems) = parse(&example);
            assert!(problems.is_empty(), "{fence}: {problems:?}");
            assert_eq!(rows.len(), 1);
            assert_eq!(rows[0].id, "D2");
        }
    }

    #[test]
    fn a_file_without_the_table_or_a_row_is_refused() {
        assert!(parse("# nothing here\n").1[0]
            .reason
            .starts_with("no table"));
        assert_eq!(
            parse(&file(&[])).1[0].reason,
            "the table has no decision row"
        );
        let twice = file(&[LEADER, LEADER]);
        assert!(parse(&twice)
            .1
            .iter()
            .any(|p| p.reason == "id `D1` repeats"));
    }

    #[test]
    fn past_its_deadline_only_a_reversible_none_user_row_takes_its_default() {
        let (rows, _) = parse(&file(&[LEADER, TO_USER, HOLD]));
        let deadline = deadline_secs("2026-11-14T00:00:00Z").unwrap();
        // Boundary: one second before the deadline nothing is due; at the deadline it is.
        assert_eq!(due(&rows[1], deadline - 1), Some(Due::Open));
        assert_eq!(due(&rows[1], deadline), Some(Due::DefaultApplies));
        assert_eq!(due(&rows[2], deadline), Some(Due::Hold));
        assert_eq!(due(&rows[2], deadline + 86_400 * 365), Some(Due::Hold));
        // Audit B1 on #2623: a reversible row in any of the five categories keeps waiting.
        for category in ["money", "legal", "external_contract", "secret", "real_data"] {
            let row = TO_USER.replace("| none |", &format!("| {category} |"));
            let (r, problems) = parse(&file(&[&row]));
            assert!(problems.is_empty(), "{category}: {problems:?}");
            assert!(r[0].reversible);
            assert_eq!(due(&r[0], deadline), Some(Due::Hold), "{category}");
        }
        // A guarded setting is reversible, yet it never takes its default either.
        let guarded = TO_USER.replace("| none |", "| guarded |");
        let (g, problems) = parse(&file(&[&guarded]));
        assert!(problems.is_empty(), "{problems:?}");
        assert_eq!(due(&g[0], deadline), Some(Due::Hold));
        // An answered row is never due.
        assert_eq!(due(&rows[0], deadline), None);
    }
}
