//! The words a person reads in a refusal, a warning or a note name the product `moai`, never
//! the earlier name `moai-cli` (slice 2 of the name change, `SPEC-20261121-moai-name`).
//!
//! Each row is one fixed sentence of the program. The message is built deep inside an apply
//! or a launch that needs a real project or a key vault to reach, so the row pins the wording
//! where it is written: the sentence with the new name must be there and the sentence with the
//! old name must not. Sentences that name a file or a state folder on disk (`bin/moai-cli`,
//! the earlier home folder) are other slices and are not rows here.

use std::fs;
use std::path::Path;

/// (source file, sentence of the earlier name, sentence of the new name)
const SENTENCES: [(&str, &str, &str); 8] = [
    (
        "src/go_state.rs",
        "Stop moai-cli commands and agent sessions in this project before the apply: the",
        "Stop moai commands and agent sessions in this project before the apply: the",
    ),
    (
        "src/go_state.rs",
        "Stop moai-cli commands and agent sessions in this project before the apply: a",
        "Stop moai commands and agent sessions in this project before the apply: a",
    ),
    (
        "src/session.rs",
        "comes from the moai-cli key vault",
        "comes from the moai key vault",
    ),
    (
        "src/launcher_terminal.rs",
        "comes from the moai-cli key vault",
        "comes from the moai key vault",
    ),
    (
        "src/state_root.rs",
        "is from a newer moai-cli; this build leaves it alone",
        "is from a newer moai; this build leaves it alone",
    ),
    (
        "src/launcher_init.rs",
        "Created by moai-cli on first launch",
        "Created by moai on first launch",
    ),
    (
        "src/claude.rs",
        "# moai-cli settings backups",
        "# moai settings backups",
    ),
    (
        "src/update/switch.rs",
        "moai-cli update --revert-switch",
        "moai update --revert-switch",
    ),
];

/// The help goldens are the text a person sees. `moai-cli` may stay in them only as the name of
/// the earlier state folder (a dot comes right before it) until the folder moves (slice 7), and in the one
/// `bin-names-plan` sentence that names the file on disk until the file is renamed (slice 5,
/// decision D-6); at 60 columns that sentence wraps and the file name stands alone on a line.
#[test]
fn the_help_goldens_keep_the_old_name_only_for_folders_and_the_alias_file() {
    const NAME: &str = "moai-cli";
    const ALIAS_LEAD: &str = "Preview the moai link beside";
    // The eight goldens `tests/help_layout.rs` compares, named one by one;
    // a golden added there is added here.
    const GOLDENS: [&str; 8] = [
        "all-60.txt",
        "all-80.txt",
        "all-dumb.txt",
        "all.txt",
        "root-60.txt",
        "root-80.txt",
        "root-dumb.txt",
        "root.txt",
    ];
    let dir = Path::new(env!("CARGO_MANIFEST_DIR")).join("tests/fixtures/help");
    let mut wrong = Vec::new();
    for name in GOLDENS {
        let path = dir.join(name);
        let text = fs::read_to_string(&path).unwrap_or_else(|e| panic!("{name}: {e}"));
        // The last few lines, trimmed: the sentence may wrap before or inside its lead at another
        // column width, so the file name is the alias file when the words of the one sentence
        // `<lead> <name>` end on this line, however the wrap cuts them.
        let mut recent: Vec<&str> = Vec::new();
        for (i, line) in text.lines().enumerate() {
            recent.push(line.trim());
            if recent.len() > 4 {
                recent.remove(0);
            }
            let alias_file = line.trim_end().ends_with(NAME)
                && recent.join(" ").ends_with(&format!("{ALIAS_LEAD} {NAME}"));
            if !alias_file {
                for (at, _) in line.match_indices(NAME) {
                    if !line[..at].ends_with('.') {
                        wrong.push(format!("{name}:{}: {}", i + 1, line.trim()));
                    }
                }
            }
        }
    }
    assert!(wrong.is_empty(), "{}", wrong.join("\n"));
}

#[test]
fn every_refusal_and_note_sentence_names_moai() {
    let root = Path::new(env!("CARGO_MANIFEST_DIR"));
    let mut wrong = Vec::new();
    for (file, old, new) in SENTENCES {
        let text = fs::read_to_string(root.join(file)).unwrap_or_else(|e| panic!("{file}: {e}"));
        if text.contains(old) {
            wrong.push(format!("{file}: still says `{old}`"));
        }
        if !text.contains(new) {
            wrong.push(format!("{file}: does not say `{new}`"));
        }
    }
    assert!(wrong.is_empty(), "{}", wrong.join("\n"));
}
