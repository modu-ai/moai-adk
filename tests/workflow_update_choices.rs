//! SPEC-20261120-update-workflow-choices (card C-57, criterion 5): the UPDATE workflow tells the
//! session to ask the `preview.choices` of a conflict step as one selection, `keep` first and
//! recommended, and says where the copy of a replacement is. Pinned here: what the sentence of
//! step 2 and the note of step 6 say, that they sit in their steps, and the byte budget.
//!
//! The tests read the workflow file of this checkout; nothing is run and no HOME is touched.

use std::{fs, path::Path};

const FILE: &str = "update.md";
/// Where the step 2 sentence starts and the words it ends with.
const CHOICES_START: &str = "Conflict choices (card C-57): ";
const CHOICES_END: &str = "stays as it is.";
/// SPEC REQ-4: the sentence is at most this many bytes.
const CHOICES_BUDGET: usize = 1200;
/// Where the step 6 note starts and the words it ends with.
const COPY_START: &str = "A replacement made from a conflict choice (card C-57) ";
const COPY_END: &str = "asks for exactly that.";
/// The bound `profile::read` puts on every package file.
const PACKAGE_FILE_BOUND: usize = 65536;

fn workflow() -> String {
    fs::read_to_string(
        Path::new(env!("CARGO_MANIFEST_DIR"))
            .join("skills/moai/workflows")
            .join(FILE),
    )
    .unwrap()
}

/// The text between two markers, the markers included.
fn cut<'a>(text: &'a str, start: &str, end: &str) -> &'a str {
    let from = text
        .find(start)
        .unwrap_or_else(|| panic!("{FILE} has no `{start}…` text"));
    let rest = &text[from..];
    let to = rest
        .find(end)
        .unwrap_or_else(|| panic!("`{start}…` does not end with `{end}`"));
    &rest[..to + end.len()]
}

/// One numbered step of the procedure, from its heading to the next one.
fn step<'a>(text: &'a str, from: &str, to: &str) -> &'a str {
    let start = text
        .find(from)
        .unwrap_or_else(|| panic!("{FILE} has no step `{from}`"));
    let rest = &text[start..];
    let end = rest
        .find(to)
        .unwrap_or_else(|| panic!("step `{from}` is not followed by `{to}`"));
    &rest[..end]
}

fn step_two(text: &str) -> &str {
    step(text, "2. **Ask.**", "3. **Apply in order.**")
}

fn step_six(text: &str) -> &str {
    step(
        text,
        "6. **Taking it back.**",
        "## Exceptions and Permissions",
    )
}

/// AC-1 (REQ-1, REQ-2): the sentence names where the choices are, asks them as one selection in
/// the printed order with `keep` recommended, and says which choice writes and how it runs.
#[test]
fn step_two_names_the_choices_and_how_each_one_runs() {
    let text = workflow();
    let sentence = cut(&text, CHOICES_START, CHOICES_END);
    for needle in [
        "`codex-config`",
        "`codex-hooks`",
        "`omp-hooks`",
        "`preview.choices`",
        "one selection",
        "printed order",
        "recommended",
        // `keep` and `diff` write nothing; `diff` is read-only and its answer shows both sides.
        "writes nothing",
        "read-only",
        "`preview.diff`",
        // `replace` writes, after the pick, as step 3 says.
        "only after the user picked it",
        "as step 3 says",
        "permission reference",
        "`command.stdin`",
        "keeps a copy of the file first",
        "`backup`",
    ] {
        assert!(sentence.contains(needle), "the sentence lacks `{needle}`");
    }
}

/// AC-1 (REQ-1): the three ids come in the printed order, `keep` first.
#[test]
fn step_two_lists_keep_diff_replace_in_that_order() {
    let text = workflow();
    let sentence = cut(&text, CHOICES_START, CHOICES_END);
    let at = |id: &str| {
        sentence
            .find(id)
            .unwrap_or_else(|| panic!("the sentence has no {id}"))
    };
    assert!(at("`keep`") < at("`diff`") && at("`diff`") < at("`replace`"));
    assert!(
        sentence[at("`keep`")..at("`diff`")].contains("recommended"),
        "`keep` is the recommended one"
    );
}

/// AC-1 (REQ-1): the sentence is part of step 2, after the sentence about a conflict, and a
/// conflict step without choices stays as it was.
#[test]
fn step_two_holds_the_sentence_after_the_conflict_rule() {
    let text = workflow();
    let two = step_two(&text);
    let sentence = cut(two, CHOICES_START, CHOICES_END);
    let conflict = two
        .find("never overwrite it by hand")
        .expect("step 2 still keeps a conflicting file and never overwrites it by hand");
    assert!(
        conflict < two.find(sentence).unwrap(),
        "the sentence comes after the conflict rule"
    );
    assert!(sentence.contains("without `preview.choices`"), "{sentence}");
}

/// AC-2 (REQ-3): step 6 says a replacement has no revert command, names the copy and says it is
/// copied back only on the user's ask; the old line about steps without a revert is still there.
#[test]
fn step_six_names_the_copy_of_a_replacement() {
    let text = workflow();
    let six = step_six(&text);
    let note = cut(six, COPY_START, COPY_END);
    for needle in [
        "no revert command",
        "the copy named by the answer's `backup`",
        "only when the user",
    ] {
        assert!(note.contains(needle), "the note lacks `{needle}`");
    }
    assert!(
        six.contains("(`codex-hooks`, `codex-config`, `omp-hooks`, or hooks written by `hooks-apply --host claude`) has no revert"),
        "the rule for steps that print no command is still there"
    );
}

/// AC-3 (REQ-4): the sentence stays inside its byte budget and the file inside the package bound.
#[test]
fn budget_holds_for_the_sentence_and_the_file() {
    let text = workflow();
    let sentence = cut(&text, CHOICES_START, CHOICES_END);
    assert!(
        sentence.len() <= CHOICES_BUDGET,
        "the sentence is {} bytes, the budget {CHOICES_BUDGET}",
        sentence.len()
    );
    assert!(
        text.len() <= PACKAGE_FILE_BOUND,
        "{FILE} is {} bytes, the package bound {PACKAGE_FILE_BOUND}",
        text.len()
    );
}
