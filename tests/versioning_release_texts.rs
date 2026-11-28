//! SPEC-20261125-skill-versioning-release-text, AC-1 to AC-3: the shipped versioning text states
//! the release rule once, in short sentences, with a week command that runs on macOS and Linux,
//! and the interview text only points to it. The tests read the skill files of this checkout and
//! start one `date` process; no HOME is read or written.

use std::{fs, path::Path, process::Command};

fn skill(rel: &str) -> String {
    let path = Path::new(env!("CARGO_MANIFEST_DIR"))
        .join("skills/moai/references/internal")
        .join(rel);
    fs::read_to_string(&path).unwrap_or_else(|e| panic!("{}: {e}", path.display()))
}

/// The opening passage of the versioning text: from its first paragraph to the first section.
fn profile_key_passage(text: &str) -> String {
    let start = text
        .find("MoAI offers three schemes")
        .expect("the versioning text opens with the three schemes");
    let end = text
        .find("\n## Keep the existing scheme")
        .expect("the versioning text has a `Keep the existing scheme` section");
    text[start..end].to_string()
}

/// A section body, from its heading line to the next `## ` heading.
fn section(text: &str, heading: &str) -> String {
    text.split(&format!("\n{heading}\n"))
        .nth(1)
        .unwrap_or_else(|| panic!("no {heading} section"))
        .split("\n## ")
        .next()
        .unwrap()
        .to_string()
}

#[test]
fn the_release_forms_and_both_branch_names_are_stated() {
    let text = skill("versioning.md");
    for form in [
        "{head}.{yearweek}.{build}-dev",
        "{head}.{release week}.{serial}",
        "`release/v{head}.{yearweek}.{n}`",
        "`release/rc/v{head}.{yearweek}.{n}`",
    ] {
        assert!(text.contains(form), "versioning.md lacks {form:?}");
    }
    for gone in ["release/<40hex>", "release/rc/<40hex>"] {
        assert!(!text.contains(gone), "versioning.md still names {gone:?}");
    }
}

#[test]
fn the_release_week_command_runs_on_macos_and_linux_and_prints_four_digits() {
    let text = skill("versioning.md");
    let bsd = "date -u -j -f %Y-%m-%d <that Monday> +%g%V";
    let gnu = "date -u -d <that Monday> +%g%V";
    for form in [bsd, gnu] {
        assert!(text.contains(form), "versioning.md lacks the form {form:?}");
    }
    assert!(
        !text.contains("+%G%V"),
        "the six-digit week form `%G%V` is still in the text"
    );
    // The Mondays: a week inside the year, a year start and a year with a 53rd week.
    let cases = [
        ("2026-11-30", "2649"),
        ("2027-01-04", "2701"),
        ("2026-12-28", "2653"),
    ];
    let host = if cfg!(target_os = "macos") { bsd } else { gnu };
    for (monday, want) in cases {
        let words: Vec<String> = host
            .replace("<that Monday>", monday)
            .split(' ')
            .map(str::to_string)
            .collect();
        let out = Command::new(&words[0])
            .args(&words[1..])
            .output()
            .unwrap_or_else(|e| panic!("{host}: {e}"));
        assert!(out.status.success(), "{host} for {monday} failed");
        let got = String::from_utf8_lossy(&out.stdout).trim().to_string();
        assert_eq!(got, want, "the week of Monday {monday}");
    }
}

#[test]
fn the_unicode_refusal_is_one_line() {
    let text = skill("versioning.md");
    // The two separators stay named: they are the only `U+` names the text keeps.
    let stripped = text.replace("U+2028", "").replace("U+2029", "");
    assert!(
        !stripped.contains("U+"),
        "versioning.md still lists a code point range"
    );
    let lines: Vec<&str> = text
        .lines()
        .filter(|l| l.contains("Default_Ignorable_Code_Point"))
        .collect();
    assert_eq!(lines.len(), 1, "the refusal sits in one line: {lines:?}");
    for word in ["Unicode 18.0.0", "General_Category Cf"] {
        assert!(lines[0].contains(word), "the refusal line lacks {word:?}");
    }
}

#[test]
fn the_profile_key_passage_has_short_sentences() {
    let passage = profile_key_passage(&skill("versioning.md"));
    for line in passage.lines() {
        for sentence in line.split(". ") {
            let words = sentence.split_whitespace().count();
            assert!(
                words <= 60,
                "a sentence of the profile key passage has {words} words: {sentence}"
            );
        }
    }
}

#[test]
fn the_versioning_text_holds_the_other_answers() {
    let text = skill("versioning.md");
    for needle in [
        "An answer through the host's \"Other\" to the `versioning` row is recorded as `{\"scheme\": \"custom\"",
        "An \"Other\" answer to `versioning_format` stays CalVer when it is a format the profile takes",
        "so the user can correct a wrong reading",
    ] {
        assert!(text.contains(needle), "versioning.md lacks {needle:?}");
    }
    // The refusal rules of custom text are one list, each rule on its own line.
    let passage = profile_key_passage(&text);
    for rule in [
        "at most 128 characters",
        "no control character",
        "no `#` at the start",
        "no credential shape",
    ] {
        let lines: Vec<&str> = passage.lines().filter(|l| l.contains(rule)).collect();
        assert_eq!(
            lines.len(),
            1,
            "the rule {rule:?} sits in one line: {lines:?}"
        );
        assert!(
            lines[0].trim_start().starts_with("- "),
            "{rule:?} is not a list item"
        );
    }
    assert!(passage.contains("at most 9999"), "the head limit is gone");
}

#[test]
fn the_interview_only_points_to_the_versioning_text() {
    let interview = skill("interview.md");
    let found: Vec<&str> = interview
        .lines()
        .filter(|l| l.starts_with("The `versioning` rows"))
        .collect();
    assert_eq!(
        found.len(),
        1,
        "one paragraph for the versioning rows: {found:?}"
    );
    let line = found[0];
    assert!(line.contains("`internal/versioning`"), "{line}");
    assert!(line.split_whitespace().count() <= 60, "{line}");
    // No card, SPEC or internal number, and no rule of its own.
    assert!(!line.contains("card "), "{line}");
    assert!(!line.contains("SPEC-"), "{line}");
    assert!(!line.contains("C-"), "{line}");
    for rule in ["custom", "128", "\"Other\"", "stays CalVer", "scheme\": "] {
        assert!(
            !line.contains(rule),
            "the interview line holds the rule {rule:?}: {line}"
        );
    }
    // The pointed-to section exists.
    assert!(!section(&skill("versioning.md"), "## Keep the existing scheme").is_empty());
}

/// The `versioning` question row tells the model to show a release tool's conflict beside the
/// question when detection reports one, in plain words and without a card, SPEC or internal
/// number, and to say which of the tool's files change if the person picks another scheme. That
/// `project-detect` prints the fact it names is pinned by `project_detect_release_tooling_cli`.
#[test]
fn the_versioning_question_shows_a_release_tool_conflict() {
    let interview = skill("interview.md");
    let rows: Vec<&str> = interview
        .lines()
        .filter(|l| l.starts_with("| `versioning` |"))
        .collect();
    assert_eq!(rows.len(), 1, "one `versioning` row: {rows:?}");
    let row = rows[0];
    let start = row
        .find("When detection reports `release_tooling.conflict`")
        .unwrap_or_else(|| panic!("the row names no release tool conflict: {row}"));
    let sentence = row[start..].split(". ").next().unwrap();
    for needle in [
        "beside the question",
        "in plain words",
        "which of the tool's files",
    ] {
        assert!(
            sentence.contains(needle),
            "{needle:?} is missing: {sentence}"
        );
    }
    for internal in ["card ", "SPEC-", "C-"] {
        assert!(!sentence.contains(internal), "{internal:?} in: {sentence}");
    }
}
