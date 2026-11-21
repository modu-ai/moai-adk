//! A lane's own wake line, and the resume note each card keeps on its Issue.
//!
//! - Claude Code: `/loop 5m /moai --mode-agent <duty>` — a fixed five-minute loop that fires
//!   only while the session is idle. Claude Code is the only host.
//!
//! The resume note is one keyed first line, `moai-resume: lane= issue= claim= lease_until=
//! branch= head= uncommitted=0`, then `next:` lines and an optional `wait: … clears=<who>`.
//! After a compact or a clear, a lane's first action is to read the note of its card. Reads
//! and writes nothing.

use serde_json::{json, Value};
use std::path::Path;

/// What a lane does after a compact or a clear, before anything else.
pub const RESUME_FIRST: &str = "compact·clear 뒤 첫 행동은 내 카드 Issue 의 moai-resume 메모 읽기";

/// The key of a resume note's first line.
pub const RESUME_KEY: &str = "moai-resume:";

/// The subcommand a lane's `Stop` hook runs.
pub const STOP_HOOK_SUBCOMMAND: &str = "lane-stop-hook";

/// Who clears a wait.
pub const CLEARS: &[&str] = &["leader", "user", "lane", "author"];

/// Seconds the host may run the lane Stop hook (its handler `timeout`).
pub const STOP_HOOK_TIMEOUT_SECS: u64 = 120;

#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Host {
    Claude,
}

impl Host {
    pub fn parse(s: &str) -> Result<Host, String> {
        match s {
            "claude" => Ok(Host::Claude),
            other => Err(format!("unknown host {other:?}: claude is the only host")),
        }
    }

    fn name(self) -> &'static str {
        match self {
            Host::Claude => "claude",
        }
    }
}

/// The most a hook file or a saved comment body is read (1 MiB).
pub const READ_LIMIT: u64 = 1 << 20;

/// A regular file's bytes, refused whole when it is over [`READ_LIMIT`] (never cut: a cut file
/// could parse as something the whole file is not, AUDIT #2917 I1). A FIFO, a folder or a
/// device is refused before it is opened.
pub fn read_bounded(path: &Path) -> Result<Vec<u8>, String> {
    use std::io::Read as _;
    match std::fs::metadata(path) {
        Ok(m) if m.is_file() => {}
        Ok(_) => return Err("not a regular file".into()),
        Err(e) => return Err(format!("unreadable: {e}")),
    }
    let file = std::fs::File::open(path).map_err(|e| format!("unreadable: {e}"))?;
    let mut bytes = Vec::new();
    // One byte past the limit shows a file that is longer, also one that grew after the check.
    file.take(READ_LIMIT + 1)
        .read_to_end(&mut bytes)
        .map_err(|e| format!("unreadable: {e}"))?;
    if bytes.len() as u64 > READ_LIMIT {
        return Err(format!("file is over {READ_LIMIT} bytes"));
    }
    Ok(bytes)
}

/// 1–`max` of `A-Z a-z 0-9 . _ -`, not starting with `-` (it would read as a flag in the Stop
/// hook command, AUDIT #2917 N1).
fn token_ok(s: &str, max: usize) -> bool {
    !s.is_empty()
        && !s.starts_with('-')
        && s.len() <= max
        && s.chars()
            .all(|c| c.is_ascii_alphanumeric() || matches!(c, '.' | '_' | '-'))
}

/// A lane name as claims carry it: 1–64 of `A-Z a-z 0-9 . _ -`.
pub fn lane_ok(lane: &str) -> bool {
    token_ok(lane, 64)
}

/// `OWNER/REPO`, each part 1–100 of `A-Z a-z 0-9 . _ -`.
pub fn repo_ok(repo: &str) -> bool {
    repo.split_once('/')
        .is_some_and(|(o, r)| token_ok(o, 100) && token_ok(r, 100))
}

/// The duty a Claude lane's loop carries: `레인 <lane> — <duty>`, by default the round's
/// restart point and the resume-note rule.
pub fn duty(lane: &str, duty: Option<&str>) -> String {
    match duty {
        Some(d) => format!("레인 {lane} — {d}"),
        None => format!("레인 {lane} — 순환 계속, self-pick (1)부터. {RESUME_FIRST}"),
    }
}

/// The command a lane's `Stop` hook runs, for any binary name (the legacy spelling is
/// `LEGACY_BINARY`).
pub fn stop_hook_command_with(binary: &str, lane: &str, repo: &str) -> String {
    format!("{binary} {STOP_HOOK_SUBCOMMAND} --lane {lane} --repo {repo}")
}

fn line(kind: &str, text: String) -> Value {
    json!({"kind": kind, "text": text})
}

/// The lines a lane on `host` pastes, with what is not measured.
pub fn lines(
    host: Host,
    lane: &str,
    repo: Option<&str>,
    duty: Option<&str>,
) -> Result<Value, String> {
    if !lane_ok(lane) {
        return Err(format!("lane {lane:?} must be 1-64 of A-Z a-z 0-9 . _ -"));
    }
    if let Some(r) = repo.filter(|r| !repo_ok(r)) {
        return Err(format!("repo {r:?} must be OWNER/REPO"));
    }
    if duty.is_some_and(|d| !one_line(d)) {
        return Err("duty must be one non-empty line".into());
    }
    let mut out = json!({"host": host.name(), "lane": lane, "writes": false});
    match host {
        Host::Claude => {
            out["lines"] = json!([line(
                "prompt",
                format!("/loop 5m /moai --mode-agent {}", self::duty(lane, duty))
            )]);
            out["verified"] = json!("documented");
            out["note"] =
                json!("/loop 5m fires only while the session is idle and expires after seven days");
        }
    }
    Ok(out)
}

#[derive(Clone, Debug, PartialEq, Eq)]
pub struct ResumeNote {
    pub lane: String,
    pub issue: u64,
    pub claim: u64,
    pub lease_until: String,
    pub branch: String,
    pub head: String,
    pub uncommitted: u64,
    pub next: Vec<String>,
    /// The wait condition and who clears it.
    pub wait: Option<(String, String)>,
}

/// `YYYY-MM-DDTHH:MM:SSZ`.
fn utc_ok(s: &str) -> bool {
    let b = s.as_bytes();
    b.len() == 20
        && b.iter().enumerate().all(|(i, c)| match i {
            4 | 7 => *c == b'-',
            10 => *c == b'T',
            13 | 16 => *c == b':',
            19 => *c == b'Z',
            _ => c.is_ascii_digit(),
        })
}

fn one_line(s: &str) -> bool {
    !s.trim().is_empty() && !s.chars().any(char::is_control)
}

/// What is wrong with a note; empty when it can be written.
pub fn problems(note: &ResumeNote) -> Vec<String> {
    let mut p = Vec::new();
    if note.uncommitted != 0 {
        p.push(format!(
            "uncommitted={}: commit or set the work aside before writing the note",
            note.uncommitted
        ));
    }
    let hex = |c: u8| c.is_ascii_digit() || (b'a'..=b'f').contains(&c);
    if !(note.head.len() == 40 && note.head.bytes().all(hex)) {
        p.push("head must be the full 40-hex commit id in lowercase".into());
    }
    // A comment id has at least seven digits (the rule of the Self-Audit `plan:` line).
    if note.claim < 1_000_000 {
        p.push("claim must be the claim comment id (at least 7 digits)".into());
    }
    if note.issue == 0 {
        p.push("issue must be a positive number".into());
    }
    if !utc_ok(&note.lease_until) {
        p.push("lease_until must be YYYY-MM-DDTHH:MM:SSZ".into());
    }
    if note.branch.is_empty()
        || note
            .branch
            .chars()
            .any(|c| c.is_whitespace() || c.is_control())
    {
        p.push("branch must be one word".into());
    }
    if !lane_ok(&note.lane) {
        p.push("lane must be 1-64 of A-Z a-z 0-9 . _ -".into());
    }
    if note.next.is_empty() || !note.next.iter().all(|n| one_line(n)) {
        p.push("next needs at least one step, each one line".into());
    }
    if let Some((cond, who)) = &note.wait {
        if !one_line(cond) || cond.contains(" clears=") {
            p.push("wait must be one line".into());
        }
        if !CLEARS.contains(&who.as_str()) {
            p.push(format!("clears must be one of {}", CLEARS.join(", ")));
        }
    }
    p
}

/// The note's text, or its problems.
pub fn render(note: &ResumeNote) -> Result<String, Vec<String>> {
    let p = problems(note);
    if !p.is_empty() {
        return Err(p);
    }
    let mut out = format!(
        "{RESUME_KEY} lane={} issue={} claim={} lease_until={} branch={} head={} uncommitted={}\n",
        note.lane,
        note.issue,
        note.claim,
        note.lease_until,
        note.branch,
        note.head,
        note.uncommitted
    );
    for n in &note.next {
        out.push_str(&format!("next: {}\n", n.trim()));
    }
    if let Some((cond, who)) = &note.wait {
        out.push_str(&format!("wait: {} clears={who}\n", cond.trim()));
    }
    Ok(out)
}

const FIELDS: &[&str] = &[
    "lane",
    "issue",
    "claim",
    "lease_until",
    "branch",
    "head",
    "uncommitted",
];

/// The note in a comment body (its one `moai-resume:` line at a line start), or the problems.
pub fn parse(body: &str) -> Result<ResumeNote, Vec<String>> {
    let lines: Vec<&str> = body.lines().collect();
    let at: Vec<usize> = (0..lines.len())
        .filter(|&i| lines[i].starts_with(RESUME_KEY))
        .collect();
    let i = match at.as_slice() {
        [] => return Err(vec![format!("no {RESUME_KEY} line at a line start")]),
        [i] => *i,
        _ => return Err(vec![format!("more than one {RESUME_KEY} line")]),
    };
    let mut p = Vec::new();
    let mut f = std::collections::BTreeMap::new();
    for pair in lines[i][RESUME_KEY.len()..].split_whitespace() {
        match pair.split_once('=') {
            Some((k, v)) if FIELDS.contains(&k) => {
                if f.insert(k, v).is_some() {
                    p.push(format!("{k} given twice"));
                }
            }
            _ => p.push(format!("unknown field {pair:?}")),
        }
    }
    for k in FIELDS {
        if !f.contains_key(k) {
            p.push(format!("{k} missing"));
        }
    }
    let mut num = |k: &str| -> u64 {
        f.get(k).map_or(0, |v| {
            v.parse().unwrap_or_else(|_| {
                p.push(format!("{k} must be a number"));
                0
            })
        })
    };
    let (issue, claim, uncommitted) = (num("issue"), num("claim"), num("uncommitted"));
    let text = |k: &str| f.get(k).map_or(String::new(), |v| v.to_string());
    let mut note = ResumeNote {
        lane: text("lane"),
        issue,
        claim,
        lease_until: text("lease_until"),
        branch: text("branch"),
        head: text("head"),
        uncommitted,
        next: Vec::new(),
        wait: None,
    };
    // The note's own lines follow its first line, up to the first other line.
    for l in &lines[i + 1..] {
        if let Some(n) = l.strip_prefix("next: ") {
            note.next.push(n.trim().to_string());
        } else if let Some(w) = l.strip_prefix("wait: ") {
            match w.rsplit_once(" clears=") {
                Some((cond, who)) => note.wait = Some((cond.trim().into(), who.trim().into())),
                None => p.push("wait needs clears=<who>".into()),
            }
        } else {
            break;
        }
    }
    if p.is_empty() {
        p = problems(&note);
    }
    if p.is_empty() {
        Ok(note)
    } else {
        Err(p)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    const HEAD: &str = "0123456789abcdef0123456789abcdef01234567";

    fn note() -> ResumeNote {
        ResumeNote {
            lane: "agent-10".into(),
            issue: 2904,
            claim: 6874514212,
            lease_until: "2026-11-14T14:00:00Z".into(),
            branch: "issue/2904-lane-loop-line".into(),
            head: HEAD.into(),
            uncommitted: 0,
            next: vec!["RED tests".into(), "push and open the PR".into()],
            wait: Some(("PR #2910 merge".into(), "leader".into())),
        }
    }

    #[test]
    fn each_host_gets_its_line() {
        let c = lines(Host::Claude, "agent-10", None, None).unwrap();
        let text = c["lines"][0]["text"].as_str().unwrap();
        assert!(
            text.starts_with("/loop 5m /moai --mode-agent 레인 agent-10 "),
            "{text}"
        );
        assert!(text.contains(RESUME_FIRST), "{text}");
        let custom = lines(Host::Claude, "agent-10", None, Some("#2904 먼저")).unwrap();
        assert_eq!(
            custom["lines"][0]["text"],
            json!("/loop 5m /moai --mode-agent 레인 agent-10 — #2904 먼저")
        );
        // A bad lane or repository is refused.
        assert!(lines(Host::Claude, "agent-10", Some("no-slash"), None).is_err());
        assert!(lines(Host::Claude, "agent 10", None, None).is_err());
        assert!(lines(Host::Claude, "", None, None).is_err());
        assert!(lines(Host::Claude, &"a".repeat(65), None, None).is_err());
        assert!(lines(Host::Claude, &"a".repeat(64), None, None).is_ok());
        assert!(Host::parse("cursor").is_err());
        // Claude Code is the only host: a removed host word is refused and the line names claude.
        assert_eq!(Host::parse("claude"), Ok(Host::Claude));
        for word in ["codex", "omp"] {
            let e = Host::parse(word).unwrap_err();
            assert!(e.contains("claude is the only host"), "{e}");
        }
    }

    #[test]
    fn a_duty_is_one_line() {
        assert!(lines(Host::Claude, "a", None, Some("x\ny")).is_err());
        assert!(lines(Host::Claude, "a", None, Some("  ")).is_err());
        assert!(lines(Host::Claude, "a", None, Some("x")).is_ok());
    }

    /// A FIFO or a folder at a path is not opened (a FIFO would block the read).
    #[cfg(unix)]
    #[test]
    fn a_fifo_or_a_folder_is_refused_by_the_bounded_read_without_blocking() {
        let dir = tempfile::tempdir().unwrap();
        let folder = dir.path().join("folder");
        std::fs::create_dir(&folder).unwrap();
        assert_eq!(read_bounded(&folder).unwrap_err(), "not a regular file");
        let fifo = dir.path().join("fifo");
        let made = std::process::Command::new("mkfifo")
            .arg(&fifo)
            .status()
            .unwrap();
        assert!(made.success());
        assert_eq!(read_bounded(&fifo).unwrap_err(), "not a regular file");
    }

    /// A file over the cap is refused whole, never judged on its first MiB.
    #[test]
    fn a_file_over_the_cap_is_refused_and_at_the_cap_is_read() {
        let dir = tempfile::tempdir().unwrap();
        let file = dir.path().join("body.json");
        let pad = |total: usize| {
            let mut s = String::from("{}");
            s.push_str(&" ".repeat(total - s.len()));
            s
        };
        std::fs::write(&file, pad(READ_LIMIT as usize)).unwrap();
        assert_eq!(read_bounded(&file).unwrap().len(), READ_LIMIT as usize);
        std::fs::write(&file, pad(READ_LIMIT as usize + 1)).unwrap();
        assert_eq!(
            read_bounded(&file).unwrap_err(),
            "file is over 1048576 bytes"
        );
        // The case the audit probed: valid JSON first, garbage after the cap.
        let mut over = pad(READ_LIMIT as usize + 10);
        over.push_str("GARBAGE");
        std::fs::write(&file, over).unwrap();
        assert_eq!(
            read_bounded(&file).unwrap_err(),
            "file is over 1048576 bytes"
        );
        std::fs::write(&file, pad(READ_LIMIT as usize)).unwrap();
        assert_eq!(read_bounded(&file).unwrap().len(), READ_LIMIT as usize);
    }

    /// AUDIT #2917 N1: a lane or repository part starting with `-` would read as a flag in the
    /// Stop hook command.
    #[test]
    fn a_leading_dash_is_not_a_lane_or_repo() {
        assert!(!lane_ok("--help"));
        assert!(!lane_ok("-a"));
        assert!(lane_ok("a-"));
        assert!(lines(Host::Claude, "a", Some("-o/r"), None).is_err());
        assert!(lines(Host::Claude, "a", Some("o/-r"), None).is_err());
    }

    #[test]
    fn a_resume_note_round_trips() {
        let n = note();
        let text = render(&n).unwrap();
        assert!(text.starts_with(
            "moai-resume: lane=agent-10 issue=2904 claim=6874514212 lease_until=2026-11-14T14:00:00Z branch=issue/2904-lane-loop-line head=0123456789abcdef0123456789abcdef01234567 uncommitted=0\n"
        ), "{text}");
        assert!(
            text.contains("\nnext: RED tests\nnext: push and open the PR\n"),
            "{text}"
        );
        assert!(
            text.contains("\nwait: PR #2910 merge clears=leader"),
            "{text}"
        );
        assert_eq!(parse(&text).unwrap(), n);
        // Inside a longer comment, after other text.
        let body = format!("PLAN 메모\n\n{text}\n🗿 MoAI\n");
        assert_eq!(parse(&body).unwrap(), n);
        let mut no_wait = n.clone();
        no_wait.wait = None;
        assert_eq!(parse(&render(&no_wait).unwrap()).unwrap(), no_wait);
    }

    #[test]
    fn a_resume_note_with_a_gap_is_refused() {
        let bad = |f: &dyn Fn(&mut ResumeNote)| {
            let mut n = note();
            f(&mut n);
            render(&n).unwrap_err()
        };
        assert!(bad(&|n| n.uncommitted = 1)[0].contains("uncommitted"));
        assert!(bad(&|n| n.head = HEAD[..39].into())[0].contains("head"));
        assert!(bad(&|n| n.head = HEAD.to_uppercase())[0].contains("head"));
        assert!(bad(&|n| n.claim = 999_999)[0].contains("claim"));
        assert!(bad(&|n| n.issue = 0)[0].contains("issue"));
        assert!(bad(&|n| n.lease_until = "2026-11-14 14:00".into())[0].contains("lease_until"));
        assert!(bad(&|n| n.branch = "a b".into())[0].contains("branch"));
        assert!(bad(&|n| n.next.clear())[0].contains("next"));
        assert!(bad(&|n| n.next = vec!["a\nb".into()])[0].contains("next"));
        assert!(bad(&|n| n.wait = Some(("x".into(), "bot".into())))[0].contains("clears"));
        assert!(bad(&|n| n.lane = "a b".into())[0].contains("lane"));
        // The boundary: a seven-digit comment id is a claim id.
        let mut n = note();
        n.claim = 1_000_000;
        assert!(render(&n).is_ok());
    }

    #[test]
    fn parse_reads_only_one_keyed_first_line() {
        let text = render(&note()).unwrap();
        assert!(parse("no note here").unwrap_err()[0].contains("no moai-resume"));
        let twice = format!("{text}\n{text}");
        assert!(parse(&twice).unwrap_err()[0].contains("more than one"));
        // A quoted note is not this comment's note.
        let quoted = text
            .lines()
            .map(|l| format!("> {l}"))
            .collect::<Vec<_>>()
            .join("\n");
        assert!(parse(&quoted).unwrap_err()[0].contains("no moai-resume"));
        // A comment that quotes another note above its own reads its own, not "more than one".
        let mut mine = note();
        mine.next = vec!["mine".into()];
        let both = format!("{quoted}\n\n{}", render(&mine).unwrap());
        assert_eq!(parse(&both).unwrap(), mine);
        let missing = text.replacen(" branch=issue/2904-lane-loop-line", "", 1);
        assert!(parse(&missing)
            .unwrap_err()
            .iter()
            .any(|p| p.contains("branch")));
        let extra = text.replacen(" uncommitted=0", " uncommitted=0 colour=red", 1);
        assert!(parse(&extra)
            .unwrap_err()
            .iter()
            .any(|p| p.contains("colour")));
        let dirty = text.replacen("uncommitted=0", "uncommitted=3", 1);
        assert!(parse(&dirty)
            .unwrap_err()
            .iter()
            .any(|p| p.contains("uncommitted")));
    }
}
