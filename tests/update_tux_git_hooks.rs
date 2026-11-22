//! The human `moai update` screens name the git-hook part: the preview row and the command that
//! confirms it, the retire folder with its revert line once the step ran, and a step that did not
//! finish as not finished. The results are the saved `update.run` fixtures with a `git_hooks`
//! part added, drawn in both languages.

use moai_cli::ui::Painter;
use moai_cli::ui_text::Lang;
use moai_cli::update_tux::{question, screen, Asked, View};
use serde_json::{json, Value};
use std::{fs, path::PathBuf};

fn fixture(name: &str) -> Value {
    let p = PathBuf::from(env!("CARGO_MANIFEST_DIR"))
        .join(format!("tests/fixtures/update_tux/{name}.json"));
    serde_json::from_str(&fs::read_to_string(&p).unwrap()).unwrap()
}

fn view(lang: Lang) -> View {
    View {
        painter: Painter {
            color: false,
            unicode: true,
            lang,
        },
        width: 100,
        home: "/home/u".into(),
        project: None,
    }
}

const FLAG: &str = "--confirm-git-hooks git-hooks:0123456789abcdef";
const REVERT: &str = "moai git-hooks-revert 20261122T010000Z";

fn part() -> Value {
    json!({
        "go_hooks": [{"name": "pre-commit", "evidence": "header"}],
        "install": ["pre-commit", "pre-push"],
        "kept": [],
        "confirm_token": "git-hooks:0123456789abcdef",
        "apply_flag": FLAG,
    })
}

fn preview() -> Value {
    let mut v = fixture("preview-changes");
    v["git_hooks"] = part();
    v
}

fn applied_step() -> Value {
    json!({
        "state": "applied",
        "stamp": "20261122T010000Z",
        "folder": "/proj/.git/hooks/moai-retired-20261122T010000Z",
    })
}

#[test]
fn the_question_names_the_git_hook_part_in_both_languages() {
    // The row names the part and counts what it would retire, install and keep.
    for (lang, word, retire) in [
        (Lang::En, "git hooks", "1 to retire"),
        (Lang::Ko, "git 훅", "1"),
    ] {
        let out = question(&preview(), &view(lang));
        let row = out
            .lines()
            .find(|l| l.contains(word))
            .unwrap_or_else(|| panic!("{lang:?}: no git-hook row: {out}"));
        assert!(row.contains(retire), "{lang:?}: {row}");
    }
}

#[test]
fn the_preview_screen_shows_the_confirm_command_whole() {
    for lang in [Lang::En, Lang::Ko] {
        let out = screen(&preview(), Asked::NotAsked, &view(lang));
        assert!(
            out.contains(&format!("moai update {FLAG}")),
            "{lang:?}: the command is shown whole: {out}"
        );
    }
}

#[test]
fn a_declined_question_still_offers_the_git_hook_command() {
    let out = screen(&preview(), Asked::No, &view(Lang::En));
    assert!(out.contains(&format!("moai update {FLAG}")), "{out}");
}

#[test]
fn a_part_that_ran_shows_the_retire_folder_and_the_revert_line() {
    let mut v = preview();
    v["git_hooks_result"] = applied_step();
    v["revert"] = json!({"git_hooks": REVERT});
    for lang in [Lang::En, Lang::Ko] {
        let out = screen(&v, Asked::NotAsked, &view(lang));
        assert!(
            out.contains("moai-retired-20261122T010000Z"),
            "{lang:?}: the retire folder: {out}"
        );
        assert!(out.contains(REVERT), "{lang:?}: the revert line: {out}");
        assert!(
            !out.contains(&format!("moai update {FLAG}")),
            "{lang:?}: the command is not offered again: {out}"
        );
    }
}

#[test]
fn the_applied_screen_shows_the_git_hook_step_and_its_revert_line() {
    let mut v = fixture("applied");
    v["git_hooks_result"] = applied_step();
    v["revert"]["git_hooks"] = json!(REVERT);
    for lang in [Lang::En, Lang::Ko] {
        let out = screen(&v, Asked::Yes, &view(lang));
        assert!(
            out.contains("moai-retired-20261122T010000Z"),
            "{lang:?}: {out}"
        );
        assert!(out.contains(REVERT), "{lang:?}: {out}");
    }
}

#[test]
fn a_step_that_did_not_finish_is_named_as_not_finished() {
    let mut v = fixture("applied");
    v["git_hooks_result"] =
        json!({"state": "failed", "message": "pre-commit appeared after the preview"});
    let out = screen(&v, Asked::Yes, &view(Lang::En));
    assert!(out.contains("did not finish"), "{out}");
    assert!(out.contains("appeared after the preview"), "{out}");
    assert!(
        !out.contains("Git hooks applied"),
        "never shown as applied: {out}"
    );
    let ok = screen(&fixture("applied"), Asked::Yes, &view(Lang::En));
    assert_ne!(
        out.lines().find(|l| l.contains('—')),
        ok.lines().find(|l| l.contains('—')),
        "a failed hook step makes the run partial, not done"
    );
}
