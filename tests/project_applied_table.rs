//! SPEC-20261119-language-reach-applied-table (card C-87, criterion 4): the PROJECT route ends with
//! a short table of which setting was written where, so the user can see that nothing was silently
//! dropped. Pinned here: the table paragraph sits after the closing-commit step and before the
//! permissions section, names its columns and the four states, forbids leaving a setting out, and
//! the Completion Check asks for it.
//!
//! The tests read one workflow file, compiled in with `include_str!`; nothing is run and no HOME
//! is touched.

const PROJECT: &str = include_str!("../skills/moai/workflows/project.md");

/// Where the paragraph starts.
const MARK: &str = "Final report (card C-87): ";

fn paragraph() -> &'static str {
    let start = PROJECT
        .find(MARK)
        .unwrap_or_else(|| panic!("project.md has no paragraph starting {MARK:?}"));
    assert_eq!(
        PROJECT.matches(MARK).count(),
        1,
        "the paragraph is written once"
    );
    let rest = &PROJECT[start..];
    let end = rest.find('\n').unwrap_or(rest.len());
    &rest[..end]
}

#[test]
fn the_final_report_has_an_applied_settings_table() {
    let p = paragraph();
    // It comes after the closing-commit step and before the permissions section.
    let closing = PROJECT
        .find("8. **Offer the closing commit")
        .expect("step 8 exists");
    let permissions = PROJECT
        .find("## Exceptions and Permissions")
        .expect("the permissions section exists");
    let at = PROJECT.find(MARK).unwrap();
    assert!(closing < at && at < permissions, "wrong place");
    // The columns and the four states.
    for word in [
        "table",
        "setting",
        "value",
        "written to",
        "`applied`",
        "`default`",
        "`HOLD`",
        "`not applied: <reason>`",
    ] {
        assert!(p.contains(word), "the paragraph must say {word:?}: {p}");
    }
    // A setting the user chose is never left out.
    assert!(
        p.contains("never left out") || p.contains("never leave"),
        "the paragraph must forbid leaving a setting out: {p}"
    );
}

#[test]
fn the_completion_check_names_the_table() {
    let at = PROJECT
        .find("## Completion Check")
        .expect("the Completion Check exists");
    let check = &PROJECT[at..];
    assert!(
        check.contains("applied-settings table"),
        "the Completion Check must name the applied-settings table"
    );
}
