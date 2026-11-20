//! The mode-flag contract (SPEC-20261120-mode-flags): `/moai --leader` and `/moai --agent`
//! (short `-l` / `-a` right after the skill name) are the mode spellings, the old
//! `--mode-leader` / `--mode-agent` stay accepted for one release with a notice, and the
//! decisions log names the decision this replaces. The tests read the shipped skill text and
//! the decisions log; no command runs here.

use std::fs;
use std::path::Path;
use std::path::PathBuf;

fn root() -> PathBuf {
    PathBuf::from(env!("CARGO_MANIFEST_DIR"))
}

fn read(rel: &str) -> String {
    fs::read_to_string(root().join(rel)).unwrap_or_else(|e| panic!("{rel}: {e}"))
}

/// The mode-selection sentence of the shipped skill names the new spellings, the short
/// aliases right after the skill name, and the one-release acceptance of the old ones.
#[test]
fn the_skill_text_names_the_new_spellings_and_the_transition() {
    let skill = read("skills/moai/SKILL.md");
    assert!(
        skill.contains("`/moai --leader`") && skill.contains("`/moai --agent`"),
        "the skill text names --leader and --agent as the mode spellings"
    );
    assert!(
        skill.contains("-l") && skill.contains("-a"),
        "the skill text names the short aliases"
    );
    assert!(
        skill.contains("바로 뒤") || skill.contains("right after"),
        "the short alias counts only right after the skill name"
    );
    assert!(
        skill.contains("한 릴리스") || skill.contains("one release"),
        "the old spellings stay accepted for one release with a notice"
    );
}

/// No shipped text instructs the old spellings as the way to invoke a mode: the skills and
/// the shipped guides spell the new form. The decisions log and the dated design-review
/// records name the old spellings legitimately — they record what was decided, they do not
/// instruct.
#[test]
fn no_shipped_text_instructs_the_old_spellings() {
    let skip_prefixes = ["docs/decisions-2026-11-08.md", "docs/design-review/"];
    let mut offenders: Vec<String> = Vec::new();
    for dir in ["skills", "docs"] {
        let mut stack = vec![root().join(dir)];
        while let Some(d) = stack.pop() {
            for e in fs::read_dir(&d).unwrap().flatten() {
                let p = e.path();
                if p.is_dir() {
                    stack.push(p);
                } else if p.extension().is_some_and(|x| x == "md" || x == "json") {
                    let shown: &Path = p.strip_prefix(root()).unwrap_or(&p);
                    let rel = shown.display().to_string();
                    if skip_prefixes
                        .iter()
                        .any(|s| rel == *s || rel.starts_with(s))
                    {
                        continue;
                    }
                    let text = fs::read_to_string(&p).unwrap_or_default();
                    if text.contains("--mode-agent") || text.contains("--mode-leader") {
                        offenders.push(rel);
                    }
                }
            }
        }
    }
    assert!(
        offenders.is_empty(),
        "shipped text still instructs the old spellings: {offenders:?}"
    );
}

/// The decisions log carries the 2026-11-20 row that names the replaced decision.
#[test]
fn the_decisions_log_names_the_replaced_decision() {
    let log = read("docs/decisions-2026-11-08.md");
    assert!(
        log.contains("11:45Z(11-20)") && log.contains("--leader"),
        "the decisions log carries the 2026-11-20 mode-flag row"
    );
    assert!(
        log.contains("#2286"),
        "the row names the replaced 2026-11-10 decision"
    );
}
