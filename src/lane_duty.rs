//! The lane duty line (#2902 O3a, plan 6874299470): the one line a leader types into a lane's
//! terminal to start or restart its non-stop loop. The line is typed into another terminal, so
//! a `--first` task that could move the cursor, end the line or reorder the text there (screen
//! injection) is refused, as is one over 300 characters.

/// The fixed duty wording. Its canonical source is the O1 #2900 document (leader decision
/// 6874308433): this text must follow that wording, and the golden in
/// `tests/lane_duty_line.rs` is updated when it changes.
pub const BODY: &str = "Issue/PR 을 배차받거나 self-pick 해서 열린 Issue·PR 이 모두 0 이 될 때까지 무중단 운영. PR 을 연 뒤 감사·CI 를 기다리지 말고 바로 다음 카드(Issue 또는 감사자 없는 PR)를 맡는다. 열린 PR 상한 8. 매 카드 끝에 LANE-REPORT 를 남기고 곧바로 다음 self-pick.";

/// Longest accepted `--first` task, in characters.
pub const MAX_FIRST_CHARS: usize = 300;

/// The hosts a lane runs on, with the prefix that starts the `moai` skill there.
pub const HOSTS: &[(&str, &str)] = &[("claude", "/moai"), ("codex", "$moai"), ("omp", "/moai")];

/// A character that must not reach another terminal: C0 and C1 controls (tab, line feed,
/// carriage return, ESC, DEL …), the Unicode line and paragraph separators, every format
/// character (general category Cf, Unicode 15.1: bidirectional and invisible controls, the
/// soft hyphen, tag characters …) and the combining grapheme joiner U+034F, which reorder
/// or hide text (#2912 audit I3).
fn unsafe_char(c: char) -> bool {
    c.is_control() || c == '\u{2028}' || c == '\u{2029}' || c == '\u{034f}' || format_char(c)
}

/// General category Cf (Unicode 15.1).
fn format_char(c: char) -> bool {
    matches!(c,
        '\u{00ad}'
        | '\u{0600}'..='\u{0605}'
        | '\u{061c}'
        | '\u{06dd}'
        | '\u{070f}'
        | '\u{0890}'..='\u{0891}'
        | '\u{08e2}'
        | '\u{180e}'
        | '\u{200b}'..='\u{200f}'
        | '\u{202a}'..='\u{202e}'
        | '\u{2060}'..='\u{2064}'
        | '\u{2066}'..='\u{206f}'
        | '\u{feff}'
        | '\u{fff9}'..='\u{fffb}'
        | '\u{110bd}'
        | '\u{110cd}'
        | '\u{13430}'..='\u{1343f}'
        | '\u{1bca0}'..='\u{1bca3}'
        | '\u{1d173}'..='\u{1d17a}'
        | '\u{e0001}'
        | '\u{e0020}'..='\u{e007f}')
}

/// The task text is not echoed in a refusal: it is what was refused.
fn check_first(first: &str) -> Result<(), String> {
    if first.is_empty() {
        return Err("--first is invalid: the task is empty".into());
    }
    if first.chars().count() > MAX_FIRST_CHARS {
        return Err(format!(
            "--first is invalid: longer than {MAX_FIRST_CHARS} characters"
        ));
    }
    if first.chars().any(unsafe_char) {
        return Err(
            "--first is invalid: control, line-separator and bidirectional characters are refused (the line is typed into another terminal)"
                .into(),
        );
    }
    Ok(())
}

/// `<prefix> --mode-agent 레인 <lane> — <BODY>[ 먼저: <first>]` for `host`.
pub fn line(lane: &str, first: Option<&str>, host: &str) -> Result<String, String> {
    if !crate::lane_registry::valid_name(lane) {
        return Err(format!(
            "lane name is invalid: use 1 to {} of A-Z a-z 0-9 . _ -, starting with a letter or digit",
            crate::lane_registry::MAX_NAME_LEN
        ));
    }
    let prefix = HOSTS
        .iter()
        .find(|(h, _)| *h == host)
        .map(|(_, p)| *p)
        .ok_or("host is invalid: use claude, codex or omp")?;
    let mut out = format!("{prefix} --mode-agent 레인 {lane} — {BODY}");
    if let Some(task) = first {
        check_first(task)?;
        out.push_str(" 먼저: ");
        out.push_str(task);
    }
    Ok(out)
}
