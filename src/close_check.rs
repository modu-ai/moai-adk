//! `close-check` (#2903, #2899 O4): may this Issue be closed now? A lane closes its own card
//! after the last pull request merged and the close table is posted; this reads the Issue and
//! answers, writing nothing.
//!
//! The close table is the newest comment whose first line starts `CLOSE-TABLE #<n>`. A
//! Markdown table in it is a criteria table when a header cell after the first starts with a
//! verdict heading ([`VERDICT_HEADINGS`]); a row whose verdict cell does not start with a word
//! that means "met" ([`MET`]) — and is not undone by a later word ([`UNDOING`]) or a `?` —
//! keeps the Issue open, and so does an open pull request whose body names this Issue in
//! `Moai-Claim: <n>#<comment>`. Fenced code blocks are not read as tables. Both word lists come from the close tables
//! of the 25 newest closed Issues on 2026-11-14 (상태 / 판정 / 결과 columns; 충족 76, 끝 20,
//! 병합 3, 지킴 2, `CLOSED …` 6 rows); tables without such a column (남은 것 | 근거 | 몫) list
//! leftovers, not criteria.

use crate::forge::transport::{HttpRequest, Transport};
use crate::lane_next::Comment;
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};

/// Comments read (GraphQL's largest page); a longer thread is read newest first.
pub const COMMENTS_READ: u64 = 100;

/// Verdict words that mean a row is met. A cell is met when it starts with one of them and the
/// next character is not a letter or digit (so `metadata` is not `met`, `미충족` is not `충족`,
/// `끝남` is not `끝`), case ignored.
pub const MET: [&str; 11] = [
    "충족",
    "해당 없음",
    "끝",
    "병합",
    "지킴",
    "완료",
    "closed",
    "met",
    "pass",
    "done",
    "n/a",
];

/// Words that undo a met word when they follow it anywhere in the cell, as a whole word (case
/// ignored): negation and not-yet (`충족 아님`, `병합 대기`, `완료 예정`, `충족(일부)`,
/// `충족 — B2 남음`, `pass 아님`, `met? no` — audit 6874717869 B1). A `?` anywhere after the
/// met word does the same. Wrong in this direction leaves an Issue open for a person to close;
/// the other way an unmet criterion would vanish.
pub const UNDOING: [&str; 33] = [
    "아님",
    "아니",
    "아니요",
    "아니오",
    "않음",
    "안",
    "전",
    "전에",
    "예정",
    "대기",
    "남음",
    "남은",
    "남아",
    "일부",
    "부분",
    "보류",
    "미정",
    "미완",
    "후속",
    "진행",
    "진행중",
    "실패",
    "no",
    "not",
    "pending",
    "partial",
    "partially",
    "todo",
    "tbd",
    "wip",
    "later",
    "fail",
    "failed",
];

/// Korean attaches particles and endings to a word (`대기중`, `안됨`, `일부만`, `미확인`), so a
/// word after the met word undoes it when it *starts* with one of these (re-audit
/// 6875342005 B1). Wrong in this direction only leaves an Issue open.
pub const UNDOING_KO_PREFIXES: [&str; 16] = [
    "대기", "예정", "일부", "부분", "남", "보류", "진행", "실패", "안", "않", "아니", "아님",
    "아직", "미", "후속", "전에",
];

/// Hangul words that start like an undoing prefix but say nothing is left: `안내`, `안전`,
/// `전역`, `미리`(보기). A prefix match, checked before [`UNDOING_KO_PREFIXES`].
pub const KO_NOT_UNDOING: [&str; 4] = ["안내", "안전", "전역", "미리"];

/// Header cells (after the first column) that name the verdict column; a prefix match, case
/// ignored, the first such cell wins.
pub const VERDICT_HEADINGS: [&str; 5] = ["판정", "상태", "결과", "verdict", "status"];

/// A pull request that refers to the Issue.
#[derive(Clone, Debug, Default, Serialize, Deserialize, PartialEq)]
pub struct RefPr {
    pub number: u64,
    /// `OPEN`, `MERGED` or `CLOSED`.
    pub state: String,
    /// The Issue its body's `Moai-Claim:` line names.
    #[serde(default)]
    pub claim_issue: Option<u64>,
}

/// Everything one answer reads. Tests and `--snapshot` give it as JSON; `collect` reads it.
#[derive(Clone, Debug, Default, Serialize, Deserialize, PartialEq)]
pub struct Snapshot {
    pub number: u64,
    /// `OPEN` or `CLOSED`.
    pub state: String,
    /// Oldest first.
    #[serde(default)]
    pub comments: Vec<Comment>,
    #[serde(default)]
    pub prs: Vec<RefPr>,
    /// What was left unread (older comments, a failed read).
    #[serde(default)]
    pub partial: Vec<String>,
}

/// The search for open pull requests whose text names `Moai-Claim: <n>#`; the hits are then
/// read by [`claim_issue`], so a loose search match is never taken as a claim.
pub fn claim_search(owner: &str, repo: &str, n: u64) -> String {
    format!("repo:{owner}/{repo} is:pr is:open \"Moai-Claim: {n}#\"")
}

/// `Moai-Claim: <issue>#<comment>` at the start of a body line: the Issue.
pub fn claim_issue(body: &str) -> Option<u64> {
    body.lines().find_map(|l| {
        let rest = l.trim().strip_prefix("Moai-Claim:")?.trim();
        let (issue, comment) = rest.split_once('#')?;
        comment.trim().parse::<u64>().ok()?;
        issue.trim().parse().ok()
    })
}

/// The first line is `CLOSE-TABLE #<n>`, with nothing or a non-digit after the number.
fn is_table_of(body: &str, n: u64) -> bool {
    let first = body.lines().next().unwrap_or("").trim();
    first
        .strip_prefix(&format!("CLOSE-TABLE #{n}"))
        .is_some_and(|rest| !rest.starts_with(|c: char| c.is_ascii_digit()))
}

/// The cells of a table row; `\|` is a pipe inside a cell (GitHub Flavored Markdown, "Tables").
fn cells(line: &str) -> Vec<String> {
    let row = line.trim();
    let row = row.strip_prefix('|').unwrap_or(row);
    let row = match row.strip_suffix('|') {
        Some(r) if !r.ends_with('\\') => r,
        _ => row,
    };
    let mut out = vec![String::new()];
    let mut chars = row.chars().peekable();
    while let Some(ch) = chars.next() {
        match ch {
            '\\' if chars.peek() == Some(&'|') => {
                out.last_mut().unwrap().push('|');
                chars.next();
            }
            '|' => out.push(String::new()),
            _ => out.last_mut().unwrap().push(ch),
        }
    }
    out.iter()
        .map(|c| c.trim().replace(['*', '`'], "").trim().to_string())
        .collect()
}

fn is_separator(line: &str) -> bool {
    cells(line)
        .iter()
        .all(|c| !c.is_empty() && c.chars().all(|ch| matches!(ch, '-' | ':' | ' ')))
}

/// A verdict cell that means the row is met.
pub fn is_met(cell: &str) -> bool {
    let lower = cell.trim().to_lowercase();
    MET.iter().any(|w| {
        lower.strip_prefix(w).is_some_and(|rest| {
            !rest.starts_with(char::is_alphanumeric) && !rest.contains('?') && !undone(rest)
        })
    })
}

/// The rest of a cell after its met word holds an [`UNDOING`] word.
fn undone(rest: &str) -> bool {
    rest.split(|c: char| !c.is_alphanumeric()).any(|word| {
        UNDOING.contains(&word)
            // Every prefix is Hangul, so only a Hangul word can match one.
            || (!KO_NOT_UNDOING.iter().any(|p| word.starts_with(p))
                && UNDOING_KO_PREFIXES.iter().any(|p| word.starts_with(p)))
    })
}

/// The verdict column of a header row: the first cell after the first that starts with a
/// verdict heading.
fn verdict_column(header: &str) -> Option<usize> {
    cells(header).iter().enumerate().skip(1).find_map(|(i, h)| {
        let h = h.to_lowercase();
        VERDICT_HEADINGS
            .iter()
            .any(|v| h.starts_with(v))
            .then_some(i)
    })
}

/// Criterion rows of the criteria tables in `body`, and those whose verdict cell is not met.
/// The first row of each table is its header and the row under it the separator; neither is a
/// criterion. A row too short to have the verdict cell is not met. Lines inside a fenced code
/// block (```` ``` ```` or `~~~`) are text, not a table (audit 6874717869 N2).
pub fn unmet_rows(body: &str) -> (usize, Vec<String>) {
    let mut rows = 0;
    let mut unmet = Vec::new();
    // `None` outside a table; `Some(col)` inside one, `col` its verdict column if any.
    let mut table: Option<Option<usize>> = None;
    let mut fence: Option<&str> = None;
    for line in body.lines() {
        let t = line.trim_start();
        let marker = ["```", "~~~"].into_iter().find(|m| t.starts_with(m));
        match (fence, marker) {
            (None, Some(m)) => {
                fence = Some(m);
                table = None;
                continue;
            }
            (Some(f), Some(m)) if f == m => {
                fence = None;
                continue;
            }
            (Some(_), _) => continue,
            (None, None) => {}
        }
        if !t.starts_with('|') {
            table = None;
            continue;
        }
        let Some(col) = table else {
            table = Some(verdict_column(line));
            continue;
        };
        let Some(col) = col else {
            continue;
        };
        if is_separator(line) {
            continue;
        }
        rows += 1;
        let met = cells(line).get(col).is_some_and(|c| is_met(c));
        if !met {
            unmet.push(line.trim().to_string());
        }
    }
    (rows, unmet)
}

/// Decides from `snap`. Pure: the same snapshot gives the same bytes.
pub fn decide(snap: &Snapshot) -> Value {
    let n = snap.number;
    let table = snap.comments.iter().rev().find(|c| is_table_of(&c.body, n));
    let open_prs: Vec<u64> = snap
        .prs
        .iter()
        .filter(|p| p.state == "OPEN" && p.claim_issue == Some(n))
        .map(|p| p.number)
        .collect();
    let merged_prs: Vec<u64> = snap
        .prs
        .iter()
        .filter(|p| p.state == "MERGED" && p.claim_issue == Some(n))
        .map(|p| p.number)
        .collect();
    let (rows, unmet) = table.map(|t| unmet_rows(&t.body)).unwrap_or_default();
    let mut reasons = Vec::new();
    let verdict = if snap.state == "CLOSED" {
        "already_closed"
    } else if table.is_none() {
        if snap.partial.is_empty() {
            reasons.push("no close table (first line CLOSE-TABLE #<n>)".to_string());
            "no_table"
        } else {
            reasons.push("no close table among the comments read; older ones are unread".into());
            "uncertain"
        }
    } else {
        if rows == 0 {
            reasons.push("the close table has no criterion rows".to_string());
        }
        if !unmet.is_empty() {
            reasons.push(format!(
                "{} row(s) not met: name the follow-up card in the table and keep the Issue open",
                unmet.len()
            ));
        }
        if !open_prs.is_empty() {
            reasons.push("an open pull request still claims this Issue".to_string());
        }
        if reasons.is_empty() {
            "close"
        } else {
            "keep_open"
        }
    };
    let next = match verdict {
        "close" => format!(
            "close the Issue, naming the table: gh issue close {n} --comment \"CLOSE-TABLE {} 확인 — 닫아요\"",
            table.map_or(0, |t| t.id)
        ),
        "no_table" => format!("post the close table (first line CLOSE-TABLE #{n}) after the last pull request merges"),
        "keep_open" => "keep the Issue open; the reasons say what is left".to_string(),
        "uncertain" => "read the older comments before deciding".to_string(),
        _ => "nothing to do".to_string(),
    };
    json!({
        "issue": n,
        "state": snap.state,
        "verdict": verdict,
        "close_table_comment": table.map(|t| t.id),
        "rows": rows,
        "unmet": unmet,
        "open_prs": open_prs,
        "merged_prs": merged_prs,
        "reasons": reasons,
        "next": next,
        "partial": snap.partial,
        "writes": false,
    })
}

const GRAPHQL_URL: &str = "https://api.github.com/graphql";

const QUERY: &str =
    "query($owner:String!,$repo:String!,$n:Int!,$claims:String!){repository(owner:$owner,name:$repo){\
issue(number:$n){number state comments(last:100){totalCount nodes{databaseId body createdAt}} \
timelineItems(itemTypes:[CROSS_REFERENCED_EVENT],last:100){nodes{...on CrossReferencedEvent{\
source{...on PullRequest{number state body}}}}}}} \
claimers:search(query:$claims,type:ISSUE,first:30){nodes{...on PullRequest{number state body}}}}";

/// Reads the Issue, its newest comments and the pull requests that refer to it in one GraphQL
/// request: those that cross-reference it (`#<n>` in their text) and the open ones a search for
/// `Moai-Claim: <n>#` finds, since a body with only the claim line may make no cross-reference
/// (audit 6874717869 N1).
pub fn collect<T: Transport>(owner: &str, repo: &str, n: u64, t: &T) -> Result<Snapshot, String> {
    let req = HttpRequest::new("POST", GRAPHQL_URL).with_body(
        json!({"query": QUERY, "variables": {"owner": owner, "repo": repo, "n": n,
            "claims": claim_search(owner, repo, n)}}),
    );
    let resp = t.send(&req)?;
    if resp.status != 200 {
        return Err(format!("GraphQL returned HTTP {}", resp.status));
    }
    let body = resp.body.unwrap_or(Value::Null);
    let issue = &body["data"]["repository"]["issue"];
    if !issue.is_object() {
        let msg = body["errors"][0]["message"]
            .as_str()
            .unwrap_or("issue not found");
        return Err(format!("Issue #{n} unreadable: {msg}"));
    }
    let comments: Vec<Comment> = issue["comments"]["nodes"]
        .as_array()
        .map(|a| {
            a.iter()
                .map(|c| Comment {
                    id: c["databaseId"].as_u64().unwrap_or(0),
                    body: c["body"].as_str().unwrap_or("").to_string(),
                    at: c["createdAt"].as_str().unwrap_or("").to_string(),
                })
                .collect()
        })
        .unwrap_or_default();
    let mut partial = Vec::new();
    let total = issue["comments"]["totalCount"].as_u64().unwrap_or(0);
    if total > comments.len() as u64 {
        partial.push(format!(
            "comments: read the newest {} of {total}",
            comments.len()
        ));
    }
    let mut prs: Vec<RefPr> = Vec::new();
    let referring = issue["timelineItems"]["nodes"]
        .as_array()
        .into_iter()
        .flatten()
        .map(|e| &e["source"]);
    let claiming = body["data"]["claimers"]["nodes"]
        .as_array()
        .into_iter()
        .flatten();
    for pr in referring.chain(claiming) {
        let Some(number) = pr["number"].as_u64() else {
            continue;
        };
        if prs.iter().any(|p| p.number == number) {
            continue;
        }
        prs.push(RefPr {
            number,
            state: pr["state"].as_str().unwrap_or("").to_string(),
            claim_issue: claim_issue(pr["body"].as_str().unwrap_or("")),
        });
    }
    Ok(Snapshot {
        number: issue["number"].as_u64().unwrap_or(n),
        state: issue["state"].as_str().unwrap_or("").to_string(),
        comments,
        prs,
        partial,
    })
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::forge::transport::HttpResponse;
    use std::cell::RefCell;

    const TABLE: &str = "CLOSE-TABLE #7 — agent-7, PR #9 병합\n\n| 기준 | 증거 | 판정 |\n| --- | --- | --- |\n| AC-1 | test | 충족 |\n| AC-2 | none | **해당 없음** |\n\n닫기는 레인이 해요.";

    #[test]
    fn met_words_need_a_boundary() {
        for yes in [
            "충족",
            "**충족**",
            "met",
            "PASS (agent-9)",
            "n/a",
            "해당 없음",
            "충족 — 감사 확인",
            "충족(범위 밖 종료)",
            "끝",
            "병합 `cf05c3ea6f`",
            "지킴",
            "CLOSED completed 2026-11-09",
            "완료",
            "done",
        ] {
            assert!(is_met(&yes.replace(['*', '`'], "")), "{yes}");
        }
        for no in [
            "미충족",
            "부분 충족",
            "metadata",
            "passed?",
            "",
            "후속 #12",
            "pending",
            "끝남",
            "일부",
            "리더 판단",
            "열림(감사 대기)",
            "metric",
            "closedown",
            // Audit 6874717869 B1: a met word undone by what follows it.
            "충족 아님",
            "충족(일부)",
            "충족 — B2 남음",
            "pass 아님",
            "met? no",
            "끝 아님",
            "done? 아니요",
            "병합 전",
            "완료 예정",
            "n/a?",
            "병합 대기",
            "충족 (pending)",
            "done, not yet merged",
            "PASS — 재감사 후속",
            // Re-audit 6875342005 B1: Korean particles and endings attached to the word.
            "병합 대기중",
            "충족 안됨",
            "충족(일부만)",
            "충족 아직",
            "충족 — 미확인",
            "충족, 단 B2 미해결",
            "끝 (남은 일 있음)",
            "완료 예정임",
        ] {
            assert!(!is_met(no), "{no}");
        }
        // Words that merely contain an undoing word stay met (`전역` is not `전`, `notes`
        // is not `not`, `안내` is not `안`).
        for yes in [
            "충족 — 판정 함수는 ColorFlag 를 받음. 전역",
            "met (notes in PR body)",
            "충족 — 안내 줄 한 번",
            "끝(대용, 리더 결정 6868714203)",
            "충족 — 미리보기 출력 그대로",
            "충족 — 안전하게 거부",
            "충족(범위 밖 종료)",
            "충족 — POSIX 만(pty)",
        ] {
            assert!(is_met(yes), "{yes}");
        }
    }

    #[test]
    fn fenced_code_is_not_a_table() {
        let fenced = format!("{TABLE}\n\n```\n| a | 상태 |\n| - | - |\n| x | 미충족 |\n```\n~~~md\n| b | 상태 |\n| - | - |\n| y | 일부 |\n~~~\n");
        assert_eq!(unmet_rows(&fenced), (2, vec![]));
        // An unclosed fence hides the rest; a table after a closed fence is read again.
        let after = "```\nx\n```\n| a | 상태 |\n| - | - |\n| x | 일부 |";
        assert_eq!(unmet_rows(after).1.len(), 1);
        // A ``` line does not close a ~~~ fence.
        let mixed = "~~~\n```\n| a | 상태 |\n| - | - |\n| x | 일부 |\n~~~";
        assert_eq!(unmet_rows(mixed), (0, vec![]));
    }

    #[test]
    fn header_and_separator_are_not_rows() {
        let (rows, unmet) = unmet_rows(TABLE);
        assert_eq!((rows, unmet.len()), (2, 0));
        let (rows, unmet) = unmet_rows(&TABLE.replace(
            "| AC-2 | none | **해당 없음** |",
            "| AC-2 | later | 후속 #12 |",
        ));
        assert_eq!(rows, 2);
        assert_eq!(unmet, vec!["| AC-2 | later | 후속 #12 |"]);
        // Two tables: each has its own header.
        let two = format!("{TABLE}\n\n| a | 상태 |\n|:-|-:|\n| x | 미충족 |\n");
        let (rows, unmet) = unmet_rows(&two);
        assert_eq!((rows, unmet.len()), (3, 1));
    }

    #[test]
    fn the_verdict_column_is_found_by_its_heading() {
        // Verdict in the middle; the evidence cell after it would not be met.
        let mid = "| AC | 상태 | 근거 |\n| --- | --- | --- |\n| AC-1 | 충족 | test x |\n| AC-2 | 일부 | later |";
        let (rows, unmet) = unmet_rows(mid);
        assert_eq!(
            (rows, unmet),
            (2, vec!["| AC-2 | 일부 | later |".to_string()])
        );
        // The first column is never the verdict column, even when its heading is one.
        let first = "| 상태 | 근거 | 결과 |\n| - | - | - |\n| 충족 | x | 일부 |";
        assert_eq!(unmet_rows(first).1.len(), 1);
        // A table without a verdict column lists leftovers, not criteria.
        let left = "| 남은 것 | 근거 | 몫 |\n| --- | --- | --- |\n| N1 | x | 리더 |";
        assert_eq!(unmet_rows(left), (0, vec![]));
        // English headings, case ignored; a row too short for the column is not met.
        let en = "| item | Verdict (tip) |\n| --- | --- |\n| a | PASS |\n| b |";
        assert_eq!(unmet_rows(en), (2, vec!["| b |".to_string()]));
        // An escaped pipe stays inside its cell (#2806's close table).
        let esc = "| AC | 상태 | 근거 |\n| - | - | - |\n| `quiet\\|plain` | 충족 | x |\n| a\\| | 일부 | y |";
        assert_eq!(unmet_rows(esc).1, vec!["| a\\| | 일부 | y |".to_string()]);
        assert_eq!(cells("| a\\|b | c |"), vec!["a|b", "c"]);
        assert_eq!(cells("a | b"), vec!["a", "b"]);
    }

    #[test]
    fn the_table_must_name_this_issue() {
        assert!(is_table_of("CLOSE-TABLE #7 — x", 7));
        assert!(is_table_of("CLOSE-TABLE #7", 7));
        assert!(!is_table_of("CLOSE-TABLE #70 — x", 7));
        assert!(!is_table_of("CLOSE-TABLE #8", 7));
        assert!(!is_table_of("> CLOSE-TABLE #7", 7));
        assert!(!is_table_of("note\nCLOSE-TABLE #7", 7));
    }

    #[test]
    fn claim_issue_is_read_only_at_a_line_start() {
        assert_eq!(claim_issue("x\nMoai-Claim: 7#123\n"), Some(7));
        assert_eq!(claim_issue("receipt: `Moai-Claim: 7#123` in commits"), None);
        // Mid-line, even when the rest would parse (#2898's body before its edit).
        assert_eq!(claim_issue("trailer is Moai-Claim: 7#123"), None);
        assert_eq!(claim_issue("  Moai-Claim: 7#123  "), Some(7));
        assert_eq!(claim_issue("Moai-Claim: 7#abc"), None);
    }

    struct Fixed(RefCell<Vec<Value>>, Value);
    impl Transport for Fixed {
        fn send(&self, req: &HttpRequest) -> Result<HttpResponse, String> {
            self.0
                .borrow_mut()
                .push(req.body.clone().unwrap_or(Value::Null));
            Ok(HttpResponse {
                status: 200,
                headers: Default::default(),
                body: Some(self.1.clone()),
            })
        }
    }

    #[test]
    fn one_request_reads_the_issue_and_its_pull_requests() {
        let answer = json!({"data": {"repository": {"issue": {
        "number": 7, "state": "OPEN",
        "comments": {"totalCount": 101, "nodes": [{"databaseId": 5, "body": TABLE, "createdAt": "t"}]},
        "timelineItems": {"nodes": [
            {"source": {"number": 9, "state": "MERGED", "body": "Moai-Claim: 7#1"}},
            {"source": {"number": 9, "state": "MERGED", "body": "Moai-Claim: 7#1"}},
            {"source": {"number": 10, "state": "OPEN", "body": "mentions #7"}},
            {}
        ]}}}}});
        let t = Fixed(RefCell::new(Vec::new()), answer);
        let snap = collect("o", "r", 7, &t).unwrap();
        assert_eq!(t.0.borrow().len(), 1);
        assert_eq!(
            t.0.borrow()[0]["variables"],
            json!({"owner": "o", "repo": "r", "n": 7,
                   "claims": "repo:o/r is:pr is:open \"Moai-Claim: 7#\""})
        );
        assert_eq!(snap.prs.len(), 2);
        assert_eq!(snap.prs[0].claim_issue, Some(7));
        assert_eq!(snap.prs[1].claim_issue, None);
        assert_eq!(snap.partial, vec!["comments: read the newest 1 of 101"]);
        // A table found among the comments read decides even when older ones are unread.
        assert_eq!(decide(&snap)["verdict"], json!("close"));
        assert_eq!(decide(&snap)["merged_prs"], json!([9]));
    }

    /// Audit 6874717869 N1: an open pull request whose body carries only the claim line (no
    /// `#<n>`, so no cross-reference) is found by the search and keeps the Issue open; a search
    /// hit that claims another Issue does not.
    #[test]
    fn a_claim_only_pull_request_is_found_by_the_search() {
        let answer = json!({"data": {
            "repository": {"issue": {
                "number": 7, "state": "OPEN",
                "comments": {"totalCount": 1, "nodes": [{"databaseId": 5, "body": TABLE}]},
                "timelineItems": {"nodes": [
                    {"source": {"number": 9, "state": "MERGED", "body": "Moai-Claim: 7#1"}}]}}},
            "claimers": {"nodes": [
                {"number": 9, "state": "MERGED", "body": "Moai-Claim: 7#1"},
                {"number": 11, "state": "OPEN", "body": "## x\nMoai-Claim: 7#1"},
                {"number": 12, "state": "OPEN", "body": "Moai-Claim: 77#1"}]}}});
        let t = Fixed(RefCell::new(Vec::new()), answer);
        let snap = collect("o", "r", 7, &t).unwrap();
        assert_eq!(t.0.borrow().len(), 1);
        let numbers: Vec<u64> = snap.prs.iter().map(|p| p.number).collect();
        assert_eq!(numbers, vec![9, 11, 12]);
        let r = decide(&snap);
        assert_eq!(r["verdict"], json!("keep_open"), "{r}");
        assert_eq!(r["open_prs"], json!([11]));
    }

    #[test]
    fn a_missing_issue_is_an_error() {
        let t = Fixed(
            RefCell::new(Vec::new()),
            json!({"data": {"repository": {"issue": null}}, "errors": [{"message": "Could not resolve"}]}),
        );
        let e = collect("o", "r", 7, &t).unwrap_err();
        assert!(e.contains("Could not resolve"), "{e}");
    }
}
