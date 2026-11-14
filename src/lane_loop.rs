//! A lane's own wake line per host, and the resume note each card keeps on its Issue (#2904,
//! O5 of the #2899 orchestration redesign, section B; user decision D-O2: a Codex lane keeps
//! one session with `/goal` plus a trusted `Stop` hook).
//!
//! - Claude Code: `/loop 5m /moai --mode-agent <duty>` — a fixed five-minute loop that fires
//!   only while the session is idle.
//! - Codex CLI: `/goal <end condition>` (enabled with `codex features enable goals`; at most
//!   4,000 characters) and a `Stop` hook whose `decision: "block"` makes Codex continue with the
//!   reason as the next prompt (docs/design-review/host-session-events-2026-11-09.md, Stop
//!   row). The hook is read here, never installed: when it is missing the lane reports
//!   `LANE-BLOCKED <lane> wake …`.
//! - OMP: `--goal "<end condition>"`, not measured (`verified: not_run`).
//!
//! The resume note is one keyed first line, `moai-resume: lane= issue= claim= lease_until=
//! branch= head= uncommitted=0`, then `next:` lines and an optional `wait: … clears=<who>`.
//! After a compact or a clear, a lane's first action is to read the note of its card. Reads
//! and writes nothing.

use serde_json::{json, Value};
use std::path::Path;

/// The lane-wide end condition (#2899 D-O1 (a)). The same words as `goal_block`'s `loop_line`;
/// a test pins the two together. The shared judgment of O2 (#2901) replaces this text.
pub const END_CONDITION: &str =
    "저장소에 열린 PR이 0이고 시작할 수 있는 Issue가 0임(HOLD·needs:decision 제외)";

/// What a lane does after a compact or a clear, before anything else.
pub const RESUME_FIRST: &str = "compact·clear 뒤 첫 행동은 내 카드 Issue 의 moai-resume 메모 읽기";

/// The key of a resume note's first line.
pub const RESUME_KEY: &str = "moai-resume:";

/// The subcommand a lane's Codex `Stop` hook runs.
pub const STOP_HOOK_SUBCOMMAND: &str = "lane-stop-hook";

/// The Codex `/goal` limit (follow-goals, "4,000자 이하").
pub const CODEX_GOAL_LIMIT: usize = 4000;

/// Who clears a wait.
pub const CLEARS: &[&str] = &["leader", "user", "lane", "author"];

/// Seconds Codex may run the lane Stop hook (its handler `timeout`).
pub const STOP_HOOK_TIMEOUT_SECS: u64 = 120;

/// The longest GitHub read of one Stop answer, inside [`STOP_HOOK_TIMEOUT_SECS`]; past it the
/// reading is unknown.
pub const STOP_READ_SECS: u64 = 90;

/// Bytes of the Codex Stop input read (and dropped).
pub const STOP_STDIN_LIMIT: u64 = 64 * 1024;

/// Blocks in a row for one next task before the hook lets the turn end (#2904 Q1 (a): the lane
/// hook ignores `stop_hook_active` and keeps its own bound). A new next task starts at zero.
pub const STOP_BLOCK_CAP: u32 = 48;

#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Host {
    Claude,
    Codex,
    Omp,
}

impl Host {
    pub fn parse(s: &str) -> Result<Host, String> {
        match s {
            "claude" => Ok(Host::Claude),
            "codex" => Ok(Host::Codex),
            "omp" => Ok(Host::Omp),
            other => Err(format!("unknown host {other:?}: claude, codex or omp")),
        }
    }

    fn name(self) -> &'static str {
        match self {
            Host::Claude => "claude",
            Host::Codex => "codex",
            Host::Omp => "omp",
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

/// The command a lane's Codex `Stop` hook runs.
pub fn stop_hook_command(lane: &str, repo: &str) -> String {
    format!(
        "{} {STOP_HOOK_SUBCOMMAND} --lane {lane} --repo {repo}",
        crate::hooks::HOOK_BINARY
    )
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
        Host::Codex => {
            let repo = repo.ok_or("--repo is required for codex (its Stop hook names it)")?;
            if duty.is_some() {
                return Err("codex takes no duty: its /goal is the end condition".into());
            }
            let goal = format!(
                "/goal {END_CONDITION} — 레인 {lane}: 매 회차 $moai --mode-agent 로 self-pick (1)부터. {RESUME_FIRST}"
            );
            out["lines"] = json!([
                line("shell", "codex features enable goals".into()),
                line("prompt", goal),
            ]);
            out["stop_hook"] = json!({
                "event": "Stop",
                "file": ".codex/hooks.json",
                "command": stop_hook_command(lane, repo),
            });
            out["verified"] = json!("not_run");
            out["note"] =
                json!("a live Codex session with /goal and this Stop hook is not measured yet");
        }
        Host::Omp => {
            out["lines"] = json!([line("flag", format!("--goal \"{END_CONDITION}\""))]);
            out["verified"] = json!("not_run");
            out["note"] = json!("OMP goal mode is not measured yet");
        }
    }
    Ok(out)
}

/// The lane's Codex `Stop` hook in `<project>/.codex/hooks.json`: `installed`, `missing` (no
/// file), `no_lane_hook`, `other_lane` or `unreadable`. Trust is Codex's own record and is
/// never read here.
pub fn codex_hook_state(project: &Path, lane: &str, repo: &str) -> &'static str {
    let path = project.join(".codex/hooks.json");
    if matches!(std::fs::symlink_metadata(&path), Err(e) if e.kind() == std::io::ErrorKind::NotFound)
    {
        return "missing";
    }
    // Only a regular file at most READ_LIMIT long is read (a FIFO would block the open).
    let Ok(bytes) = read_bounded(&path) else {
        return "unreadable";
    };
    let Ok(v) = serde_json::from_slice::<Value>(&bytes) else {
        return "unreadable";
    };
    let mine = stop_hook_command(lane, repo);
    let prefix = format!("{} {STOP_HOOK_SUBCOMMAND} ", crate::hooks::HOOK_BINARY);
    let mut other = false;
    for group in v["hooks"]["Stop"].as_array().into_iter().flatten() {
        for h in group["hooks"].as_array().into_iter().flatten() {
            let cmd = h["command"].as_str().unwrap_or("").trim();
            if cmd == mine {
                return "installed";
            }
            other |= cmd.starts_with(&prefix);
        }
    }
    if other {
        "other_lane"
    } else {
        "no_lane_hook"
    }
}

/// The line a Codex lane reports when its `Stop` hook is not in place (`state` from
/// [`codex_hook_state`]); none when it is. Installing the hook is the person's step: the hook
/// file is reviewed and trusted in Codex's own `/hooks`.
pub fn codex_blocked_line(lane: &str, state: &str) -> Option<String> {
    (state != "installed").then(|| {
        format!(
            "LANE-BLOCKED {lane} wake codex Stop hook {state} · kind=outside owner=user \
             next=install the lane Stop hook in .codex/hooks.json and trust it in Codex /hooks"
        )
    })
}

/// What a lane's Stop hook read: a snapshot, or nothing it can trust (a failed or timed-out
/// GitHub read, an unreadable `--snapshot`).
pub enum Reading {
    Snapshot(crate::lane_next::Snapshot),
    Unknown,
}

/// The Stop hook's count between stops, per lane under the project state root.
#[derive(Clone, Debug, Default, PartialEq, Eq, serde::Serialize, serde::Deserialize)]
pub struct StopState {
    /// The next task the blocks are for (`lane_next::summary_line`, or `wait`).
    #[serde(default)]
    pub key: String,
    /// Blocks in a row for `key`.
    #[serde(default)]
    pub blocks: u32,
    /// The last stop blocked on an unknown reading; the next unknown one passes.
    #[serde(default)]
    pub unknown_blocked: bool,
}

/// The lane's Stop decision: the reason to keep working, or `None` to let the turn end.
/// - the lane has a next task → block (fixed words and `lane_next::summary_line` only, no text
///   read from GitHub);
/// - nothing of its own is startable but pull requests are open → block, telling it to wait
///   with `lane-next --wait 1800` (#2904 Q2 (a));
/// - no open pull request and no startable Issue (the end condition, as `lane_next` reads it)
///   → pass;
/// - unknown (no reading, or a partial reading with no next task) → block once, then pass;
/// - [`STOP_BLOCK_CAP`] blocks in a row for the same next task → pass.
pub fn stop_decision(
    lane: &str,
    repo: &str,
    reading: &Reading,
    state: &mut StopState,
) -> Option<String> {
    let known = match reading {
        Reading::Snapshot(s) => {
            let v = crate::lane_next::decide(lane, s);
            let open_prs = s.prs.len();
            (!v["next"].is_null() || s.partial.is_empty()).then_some((v, open_prs))
        }
        Reading::Unknown => None,
    };
    let Some((v, open_prs)) = known else {
        if state.unknown_blocked {
            state.unknown_blocked = false;
            return None;
        }
        state.unknown_blocked = true;
        return Some(format!(
            "레인 {lane}: 다음 일을 읽지 못했어요(GitHub 읽기 실패·시간 초과). `moai lane-next --lane {lane} --repo {repo}` 로 한 번 더 확인하고 이어 가세요. 다음 멈춤은 그대로 통과합니다."
        ));
    };
    state.unknown_blocked = false;
    let (key, reason) = if !v["next"].is_null() {
        let summary = crate::lane_next::summary_line(&v);
        let reason = format!(
            "레인 {lane}: 할 일이 남았어요({summary}). `moai lane-next --lane {lane} --repo {repo}` 로 그 일을 읽고 이어서 하세요. 끝 조건: {END_CONDITION}"
        );
        (summary, reason)
    } else if open_prs > 0 {
        let reason = format!(
            "레인 {lane}: 지금 시작할 내 일은 없지만 열린 PR {open_prs}개가 남았어요. `moai lane-next --lane {lane} --repo {repo} --wait 1800` 으로 다음 일을 기다리세요. 끝 조건: {END_CONDITION}"
        );
        ("wait".to_string(), reason)
    } else {
        *state = StopState::default();
        return None;
    };
    if key != state.key {
        state.key = key;
        state.blocks = 0;
    }
    if state.blocks >= STOP_BLOCK_CAP {
        return None;
    }
    state.blocks += 1;
    Some(reason)
}

/// The state file of `lane` in the project `root`.
pub fn stop_state_path(root: &Path, lane: &str) -> std::path::PathBuf {
    crate::state_root::path(root, &format!("lane-stop-hook/{lane}.json"))
}

fn load_state(path: &Path) -> StopState {
    match std::fs::symlink_metadata(path) {
        Ok(m) if m.is_file() && m.len() <= 4096 => std::fs::read(path)
            .ok()
            .and_then(|b| serde_json::from_slice(&b).ok())
            .unwrap_or_default(),
        _ => StopState::default(),
    }
}

/// Writes the state only under an existing state root, never through a link.
fn save_state(root: &Path, path: &Path, state: &StopState) -> Result<(), ()> {
    let real_dir = |p: &Path| std::fs::symlink_metadata(p).is_ok_and(|m| m.is_dir());
    if !real_dir(&crate::state_root::path(root, "")) {
        return Err(());
    }
    let dir = path.parent().ok_or(())?;
    if std::fs::symlink_metadata(dir).is_err() {
        std::fs::create_dir(dir).map_err(|_| ())?;
    }
    if !real_dir(dir) {
        return Err(());
    }
    if std::fs::symlink_metadata(path).is_ok_and(|m| !m.is_file()) {
        return Err(());
    }
    let bytes = serde_json::to_vec(state).map_err(|_| ())?;
    std::fs::write(path, bytes).map_err(|_| ())
}

/// The Codex Stop answer of `lane` in the project `root`: `{"decision":"block","reason":…}`,
/// or empty to let the turn end. A block whose count cannot be saved (no state root, a link)
/// is not sent: an uncounted block could not be bounded.
pub fn stop_hook(root: &Path, lane: &str, repo: &str, reading: &Reading) -> String {
    let path = stop_state_path(root, lane);
    let mut state = load_state(&path);
    let reason = stop_decision(lane, repo, reading, &mut state);
    match (reason, save_state(root, &path, &state)) {
        (Some(reason), Ok(())) => json!({"decision": "block", "reason": reason}).to_string(),
        _ => String::new(),
    }
}

/// A `--snapshot` file ([`read_bounded`]: a regular file of at most [`READ_LIMIT`]) holding a
/// `lane_next::Snapshot`; anything else is unknown.
pub fn read_snapshot(file: &Path) -> Reading {
    read_bounded(file)
        .ok()
        .and_then(|b| serde_json::from_slice(&b).ok())
        .map_or(Reading::Unknown, Reading::Snapshot)
}

/// One `lane_next::collect` read of GitHub (no wait), in the working directory's clone; a
/// transport that cannot start or a read past [`STOP_READ_SECS`] is unknown.
pub fn read_github(lane: &str, repo: &str) -> Reading {
    let (tx, rx) = std::sync::mpsc::channel();
    let (lane, repo) = (lane.to_string(), repo.to_string());
    std::thread::spawn(move || {
        let Ok(t) = crate::forge::transport::GhTransport::from_env() else {
            let _ = tx.send(None);
            return;
        };
        let Some((owner, name)) = repo.split_once('/') else {
            let _ = tx.send(None);
            return;
        };
        let dir = std::path::PathBuf::from(".");
        let target = crate::lane_next::Target {
            owner,
            repo: name,
            lane: &lane,
            tip_branch: None,
            dir: &dir,
        };
        let _ = tx.send(Some(crate::lane_next::collect(&target, &t)));
    });
    match rx.recv_timeout(std::time::Duration::from_secs(STOP_READ_SECS)) {
        Ok(Some(s)) => Reading::Snapshot(s),
        _ => Reading::Unknown,
    }
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
        let x = lines(Host::Codex, "agent-10", Some("modu-ai/moai-adk"), None).unwrap();
        let texts: Vec<&str> = x["lines"]
            .as_array()
            .unwrap()
            .iter()
            .map(|l| l["text"].as_str().unwrap())
            .collect();
        assert_eq!(texts[0], "codex features enable goals");
        assert!(texts[1].starts_with("/goal ") && texts[1].contains(END_CONDITION));
        assert!(texts[1].contains("$moai --mode-agent") && texts[1].contains(RESUME_FIRST));
        assert!(texts[1].chars().count() <= CODEX_GOAL_LIMIT);
        assert_eq!(
            x["stop_hook"]["command"],
            json!("moai-cli lane-stop-hook --lane agent-10 --repo modu-ai/moai-adk")
        );
        let o = lines(Host::Omp, "agent-10", None, None).unwrap();
        assert_eq!(
            o["lines"][0]["text"],
            json!(format!("--goal \"{END_CONDITION}\""))
        );
        assert_eq!(o["verified"], json!("not_run"));
        // Codex needs the repository for its hook; a bad lane or repository is refused.
        assert!(lines(Host::Codex, "agent-10", None, None).is_err());
        assert!(lines(Host::Codex, "agent-10", Some("no-slash"), None).is_err());
        assert!(lines(Host::Claude, "agent 10", None, None).is_err());
        assert!(lines(Host::Claude, "", None, None).is_err());
        assert!(lines(Host::Claude, &"a".repeat(65), None, None).is_err());
        assert!(lines(Host::Claude, &"a".repeat(64), None, None).is_ok());
        assert!(Host::parse("cursor").is_err());
    }

    /// The end condition is goal_block's lane-wide condition, so the two never drift apart.
    #[test]
    fn the_end_condition_is_goal_blocks_loop_condition() {
        let block = crate::goal_block::build(&json!({"milestone": "m",
            "issues": [{"number": 1, "state": "open"}]}))
        .unwrap();
        let line = block["agent"]["loop_line"].as_str().unwrap();
        let (head, tail) = END_CONDITION.split_once('(').unwrap();
        assert!(line.contains(head), "{line}");
        assert!(line.contains(tail.trim_end_matches(')')), "{line}");
    }

    #[test]
    fn a_long_duty_cannot_push_the_codex_goal_over_its_limit() {
        // Codex has no duty; the goal is fixed text, under the limit for the longest lane name.
        let x = lines(Host::Codex, &"a".repeat(64), Some("o/r"), None).unwrap();
        assert!(x["lines"][1]["text"].as_str().unwrap().chars().count() <= CODEX_GOAL_LIMIT);
        // A duty is one line.
        assert!(lines(Host::Claude, "a", None, Some("x\ny")).is_err());
    }

    #[test]
    fn the_codex_hook_state_reads_the_project_file() {
        let dir = tempfile::tempdir().unwrap();
        let p = dir.path();
        assert_eq!(codex_hook_state(p, "agent-10", "o/r"), "missing");
        std::fs::create_dir(p.join(".codex")).unwrap();
        let file = p.join(".codex/hooks.json");
        std::fs::write(&file, "{").unwrap();
        assert_eq!(codex_hook_state(p, "agent-10", "o/r"), "unreadable");
        let with = |cmd: &str| {
            json!({"hooks": {"Stop": [{"hooks": [
                {"type": "command", "command": "moai-cli hook-receipt --host codex", "timeout": 10},
            ]}, {"hooks": [{"type": "command", "command": cmd, "timeout": 120}]}]}})
            .to_string()
        };
        std::fs::write(&file, with("echo hi")).unwrap();
        assert_eq!(codex_hook_state(p, "agent-10", "o/r"), "no_lane_hook");
        std::fs::write(&file, with(&stop_hook_command("agent-9", "o/r"))).unwrap();
        assert_eq!(codex_hook_state(p, "agent-10", "o/r"), "other_lane");
        std::fs::write(&file, with(&stop_hook_command("agent-10", "o/x"))).unwrap();
        assert_eq!(codex_hook_state(p, "agent-10", "o/r"), "other_lane");
        std::fs::write(&file, with(&stop_hook_command("agent-10", "o/r"))).unwrap();
        assert_eq!(codex_hook_state(p, "agent-10", "o/r"), "installed");
        // The hook under another event is not a Stop hook.
        let pre = json!({"hooks": {"PreToolUse": [{"hooks": [{"type": "command",
            "command": stop_hook_command("agent-10", "o/r")}]}]}});
        std::fs::write(&file, pre.to_string()).unwrap();
        assert_eq!(codex_hook_state(p, "agent-10", "o/r"), "no_lane_hook");
    }

    /// A FIFO or a folder at the hook file's path is not opened (a FIFO would block the read).
    #[cfg(unix)]
    #[test]
    fn a_fifo_or_a_folder_at_the_hook_path_is_unreadable_without_blocking() {
        let dir = tempfile::tempdir().unwrap();
        let p = dir.path();
        std::fs::create_dir_all(p.join(".codex/hooks.json")).unwrap();
        assert_eq!(codex_hook_state(p, "a", "o/r"), "unreadable");
        let q = tempfile::tempdir().unwrap();
        std::fs::create_dir(q.path().join(".codex")).unwrap();
        let fifo = q.path().join(".codex/hooks.json");
        let made = std::process::Command::new("mkfifo")
            .arg(&fifo)
            .status()
            .unwrap();
        assert!(made.success());
        assert_eq!(codex_hook_state(q.path(), "a", "o/r"), "unreadable");
    }

    /// AUDIT #2917 I1: a file over the cap is refused whole, never judged on its first MiB.
    #[test]
    fn a_hook_file_over_the_cap_is_unreadable_and_at_the_cap_is_read() {
        let dir = tempfile::tempdir().unwrap();
        let p = dir.path();
        std::fs::create_dir(p.join(".codex")).unwrap();
        let hook = json!({"hooks": {"Stop": [{"hooks": [{"type": "command",
            "command": stop_hook_command("a", "o/r")}]}]}})
        .to_string();
        let pad = |total: usize| {
            let mut s = hook.clone();
            s.push_str(&" ".repeat(total - hook.len()));
            s
        };
        let file = p.join(".codex/hooks.json");
        std::fs::write(&file, pad(READ_LIMIT as usize)).unwrap();
        assert_eq!(codex_hook_state(p, "a", "o/r"), "installed");
        std::fs::write(&file, pad(READ_LIMIT as usize + 1)).unwrap();
        assert_eq!(codex_hook_state(p, "a", "o/r"), "unreadable");
        // The case the audit probed: valid JSON first, garbage after the cap.
        let mut over = pad(READ_LIMIT as usize + 10);
        over.push_str("GARBAGE");
        std::fs::write(&file, over).unwrap();
        assert_eq!(codex_hook_state(p, "a", "o/r"), "unreadable");
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
        assert!(lines(Host::Codex, "a", Some("-o/r"), None).is_err());
        assert!(lines(Host::Codex, "a", Some("o/-r"), None).is_err());
    }

    #[test]
    fn only_a_missing_hook_gives_the_blocked_line() {
        assert_eq!(codex_blocked_line("agent-10", "installed"), None);
        for state in ["missing", "no_lane_hook", "other_lane", "unreadable"] {
            let l = codex_blocked_line("agent-10", state).unwrap();
            assert!(l.starts_with(&format!(
                "LANE-BLOCKED agent-10 wake codex Stop hook {state} · "
            )));
            assert!(l.contains("kind=outside owner=user next="), "{l}");
        }
    }

    fn snap(v: Value) -> Reading {
        Reading::Snapshot(serde_json::from_value(v).unwrap())
    }

    fn issue(n: u64) -> Reading {
        snap(json!({"open_issues": [{"number": n, "title": "t"}]}))
    }

    #[test]
    fn the_stop_decision_blocks_until_the_end_condition() {
        let mut st = StopState::default();
        let r = stop_decision("a", "o/r", &issue(7), &mut st).unwrap();
        assert!(r.contains("lane-next: P2 pick_issue #7"), "{r}");
        assert!(r.contains(END_CONDITION), "{r}");
        assert_eq!(st.blocks, 1);
        // Another lane's open draft: nothing of its own, but a pull request is open.
        let draft = snap(json!({"prs": [{"number": 9, "head": "x", "draft": true,
            "author_lane": "b"}]}));
        let r = stop_decision("a", "o/r", &draft, &mut st).unwrap();
        assert!(
            r.contains("--wait 1800") && r.contains("열린 PR 1개"),
            "{r}"
        );
        assert_eq!((st.key.as_str(), st.blocks), ("wait", 1));
        // The end condition passes and clears the count.
        assert_eq!(stop_decision("a", "o/r", &snap(json!({})), &mut st), None);
        assert_eq!(st, StopState::default());
    }

    #[test]
    fn the_stop_cap_passes_at_48_blocks_in_a_row_until_the_next_task_changes() {
        let mut st = StopState::default();
        stop_decision("a", "o/r", &issue(7), &mut st).unwrap();
        st.blocks = 46;
        assert!(
            stop_decision("a", "o/r", &issue(7), &mut st).is_some(),
            "47th"
        );
        assert!(
            stop_decision("a", "o/r", &issue(7), &mut st).is_some(),
            "48th"
        );
        assert_eq!(st.blocks, STOP_BLOCK_CAP);
        assert_eq!(stop_decision("a", "o/r", &issue(7), &mut st), None, "49th");
        assert_eq!(stop_decision("a", "o/r", &issue(7), &mut st), None);
        // A new next task starts again.
        assert!(stop_decision("a", "o/r", &issue(8), &mut st).is_some());
        assert_eq!(st.blocks, 1);
    }

    #[test]
    fn an_unknown_reading_blocks_once_then_passes() {
        let mut st = StopState::default();
        for unknown in [Reading::Unknown, snap(json!({"partial": ["lists: down"]}))] {
            let r = stop_decision("a", "o/r", &unknown, &mut st).unwrap();
            assert!(r.contains("읽지 못했어요"), "{r}");
            assert_eq!(stop_decision("a", "o/r", &unknown, &mut st), None);
        }
        // A snapshot over the read cap is unknown, never judged on its first MiB.
        let dir = tempfile::tempdir().unwrap();
        let file = dir.path().join("s.json");
        let body = json!({"open_issues": [{"number": 7, "title": "t"}]}).to_string();
        std::fs::write(&file, &body).unwrap();
        assert!(matches!(read_snapshot(&file), Reading::Snapshot(_)));
        std::fs::write(&file, format!("{body}{}", " ".repeat(READ_LIMIT as usize))).unwrap();
        assert!(matches!(read_snapshot(&file), Reading::Unknown));
        // A partial reading that still names a next task is known.
        let partial_work = snap(json!({"partial": ["x"],
            "open_issues": [{"number": 7, "title": "t"}]}));
        assert!(stop_decision("a", "o/r", &partial_work, &mut st)
            .unwrap()
            .contains("#7"));
    }

    #[test]
    fn the_stop_hook_counts_only_under_a_state_root() {
        let dir = tempfile::tempdir().unwrap();
        let root = dir.path();
        // No state root: nothing can be counted, so nothing blocks.
        assert_eq!(stop_hook(root, "a", "o/r", &issue(7)), "");
        std::fs::create_dir_all(crate::state_root::path(root, "")).unwrap();
        let out: Value = serde_json::from_str(&stop_hook(root, "a", "o/r", &issue(7))).unwrap();
        assert_eq!(out["decision"], json!("block"));
        let saved: StopState =
            serde_json::from_slice(&std::fs::read(stop_state_path(root, "a")).unwrap()).unwrap();
        assert_eq!(saved.blocks, 1);
    }

    #[cfg(unix)]
    #[test]
    fn the_stop_hook_never_counts_through_a_link() {
        use std::os::unix::fs::symlink;
        let dir = tempfile::tempdir().unwrap();
        let outside = tempfile::tempdir().unwrap();
        // The state file is a link to a file elsewhere: no write through it, so no block.
        let root = dir.path().join("p1");
        std::fs::create_dir_all(crate::state_root::path(&root, "lane-stop-hook")).unwrap();
        let target = outside.path().join("victim.json");
        std::fs::write(&target, b"keep").unwrap();
        symlink(&target, stop_state_path(&root, "a")).unwrap();
        assert_eq!(stop_hook(&root, "a", "o/r", &issue(7)), "");
        assert_eq!(std::fs::read(&target).unwrap(), b"keep");
        // The `lane-stop-hook` folder is a link to a folder elsewhere: nothing is written there.
        let root = dir.path().join("p2");
        std::fs::create_dir_all(crate::state_root::path(&root, "")).unwrap();
        symlink(
            outside.path(),
            crate::state_root::path(&root, "lane-stop-hook"),
        )
        .unwrap();
        assert_eq!(stop_hook(&root, "a", "o/r", &issue(7)), "");
        assert!(!outside.path().join("a.json").exists());
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
