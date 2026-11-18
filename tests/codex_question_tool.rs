//! SPEC-20261118-codex-question-tool: the research note behind `internal/questions.md` names
//! each host's question tool from its official source, with the URL and the UTC time it was
//! read, agrees with the table row, and states the Codex run that was not made as `not_run`
//! with its reason.

use std::fs;

const RESEARCH: &str = concat!(
    env!("CARGO_MANIFEST_DIR"),
    "/.moai/specs/SPEC-20261118-codex-question-tool/research.md"
);
const QUESTIONS: &str = concat!(
    env!("CARGO_MANIFEST_DIR"),
    "/skills/moai/references/internal/questions.md"
);

fn research() -> String {
    fs::read_to_string(RESEARCH).unwrap_or_else(|e| panic!("research.md is missing: {e}"))
}

/// The text under `## <heading>` up to the next `## ` heading.
fn section(text: &str, heading: &str) -> String {
    let start = format!("## {heading}");
    let mut out = String::new();
    let mut inside = false;
    for line in text.lines() {
        if line.starts_with("## ") {
            inside = line.trim_end() == start;
            continue;
        }
        if inside {
            out.push_str(line);
            out.push('\n');
        }
    }
    assert!(!out.is_empty(), "research.md has no section {start:?}");
    out
}

/// True for `YYYY-MM-DDTHH:MM:SSZ`.
fn is_utc_time(s: &str) -> bool {
    let b = s.as_bytes();
    b.len() == 20
        && b[19] == b'Z'
        && b[4] == b'-'
        && b[7] == b'-'
        && b[10] == b'T'
        && b[13] == b':'
        && b[16] == b':'
        && b.iter()
            .enumerate()
            .all(|(i, c)| matches!(i, 4 | 7 | 10 | 13 | 16 | 19) || c.is_ascii_digit())
}

/// Every `read: <UTC>` line of a section.
fn read_times(sec: &str) -> Vec<String> {
    sec.lines()
        .filter_map(|l| l.trim().strip_prefix("- read: "))
        .map(|t| t.trim().to_string())
        .collect()
}

fn urls(sec: &str) -> Vec<String> {
    sec.lines()
        .filter_map(|l| l.trim().strip_prefix("- url: "))
        .map(|u| u.trim().to_string())
        .collect()
}

/// AC-1 (card criterion 1): each host's facts carry a source URL and the UTC read time.
#[test]
fn research_note_quotes_each_host_from_its_source_with_url_and_read_time() {
    let text = research();
    for (heading, names) in [
        (
            "Codex CLI",
            &[
                "`request_user_input`",
                "`features.default_mode_request_user_input`",
            ][..],
        ),
        ("OMP", &["`ask`", "`ask.timeout`"][..]),
    ] {
        let sec = section(&text, heading);
        let found = urls(&sec);
        assert!(!found.is_empty(), "{heading}: no `- url:` line");
        for u in &found {
            assert!(
                u.starts_with("https://") || u.starts_with('/'),
                "{heading}: {u:?} is neither an https URL nor an installed file path"
            );
        }
        let times = read_times(&sec);
        assert!(!times.is_empty(), "{heading}: no `- read:` line");
        for t in &times {
            assert!(is_utc_time(t), "{heading}: {t:?} is not a UTC time");
        }
        assert_eq!(
            found.len(),
            times.len(),
            "{heading}: every source needs its own read time"
        );
        for n in names {
            assert!(sec.contains(n), "{heading} lacks {n}");
        }
        assert!(sec.contains("- quote: "), "{heading}: no quoted text");
    }
}

/// AC-2: the research and the table row of `questions.md` say the same about the tool names
/// and the flag.
#[test]
fn research_note_agrees_with_the_hosts_rows_of_questions_md() {
    let q = fs::read_to_string(QUESTIONS).unwrap();
    let row = |host: &str| {
        q.lines()
            .find(|l| l.starts_with(&format!("| {host} |")))
            .unwrap_or_else(|| panic!("no row for {host}"))
            .to_string()
    };
    let text = research();
    let codex = section(&text, "Codex CLI");
    let omp = section(&text, "OMP");
    let codex_row = row("Codex CLI");
    for n in [
        "`request_user_input`",
        "`features.default_mode_request_user_input`",
    ] {
        assert!(codex_row.contains(n) && codex.contains(n), "Codex: {n}");
    }
    assert!(
        codex_row.contains("off by default") && codex.contains("default: false"),
        "the flag default must be off in the row and in the research"
    );
    let omp_row = row("OMP");
    for n in ["`ask`", "`ask.timeout`"] {
        assert!(omp_row.contains(n) && omp.contains(n), "OMP: {n}");
    }
}

/// AC-3 (card criterion 4): the Codex run of `$moai project` that was not made is `not_run`
/// with its reason, and the read-only measurement that was made is recorded with its output.
#[test]
fn codex_run_gap_is_stated_as_not_run_with_its_reason() {
    let text = research();
    let sec = section(&text, "Codex run");
    let status = sec
        .lines()
        .find_map(|l| l.trim().strip_prefix("- status: "))
        .expect("`- status:` line");
    assert!(status.starts_with("not_run"), "status line: {status:?}");
    assert!(
        sec.lines()
            .any(|l| l.trim().starts_with("- reason: ") && l.trim().len() > 30),
        "the reason is missing or empty"
    );
    for measured in [
        "codex debug prompt-input",
        "codex features list",
        "codex-cli 0.170.0",
        "This tool is only available in Plan mode.",
        "This tool is only available in Default or Plan mode.",
    ] {
        assert!(sec.contains(measured), "measurement lacks {measured:?}");
    }
    let times = read_times(&sec);
    assert!(
        !times.is_empty() && times.iter().all(|t| is_utc_time(t)),
        "the measurement needs its UTC time"
    );
}
