//! SPEC-20261121-model-step-reachable: the model step of `moai update` opens on a person's
//! terminal also when nothing is saved, so a person whose earlier run applied everything can
//! still reach the questions. The decision is a pure function and the saved-or-not read is a
//! read of one file; both are tested here without a terminal. The terminal run itself is read by
//! `tests/update_model_step_reachable_tty_test.py`.

use moai_cli::update::{model_step_opens, models_saved};
use moai_cli::update_models::settings_path;
use std::fs;
use std::path::Path;
use tempfile::TempDir;

/// AC-1: all sixteen combinations. The step opens exactly for a run that is not scripted, has a
/// person's terminal, and either wrote files or has no saved model setting.
#[test]
fn the_model_step_opens_for_exactly_the_runs_the_spec_names() {
    for scripted in [false, true] {
        for terminal in [false, true] {
            for wrote in [false, true] {
                for saved in [false, true] {
                    let want = !scripted && terminal && (wrote || !saved);
                    assert_eq!(
                        model_step_opens(scripted, terminal, wrote, saved),
                        want,
                        "scripted={scripted} terminal={terminal} wrote={wrote} saved={saved}"
                    );
                }
            }
        }
    }
    // The three cases that matter most, written out.
    assert!(
        model_step_opens(false, true, false, false),
        "nothing applied, nothing saved: the step opens"
    );
    assert!(
        !model_step_opens(false, true, false, true),
        "nothing applied, something saved: it stays closed"
    );
    assert!(
        !model_step_opens(true, true, true, false),
        "a scripted run never asks"
    );
}

fn home_with(settings: Option<&str>) -> TempDir {
    let home = TempDir::new().unwrap();
    if let Some(text) = settings {
        let path = settings_path(home.path());
        fs::create_dir_all(path.parent().unwrap()).unwrap();
        fs::write(path, text).unwrap();
    }
    home
}

fn bytes(home: &Path) -> Option<Vec<u8>> {
    fs::read(settings_path(home)).ok()
}

/// AC-2: no saved model setting is `false`, whatever else the file holds.
#[test]
fn no_saved_model_setting_reads_as_not_saved() {
    for (what, settings) in [
        ("no file", None),
        ("an empty object", Some("{}")),
        (
            "a launcher without models",
            Some(r#"{"launcher":{"advisor":"fable"}}"#),
        ),
        (
            "an empty models object",
            Some(r#"{"launcher":{"models":{}}}"#),
        ),
    ] {
        let home = home_with(settings);
        let before = bytes(home.path());
        assert!(!models_saved(home.path()), "{what}");
        assert_eq!(bytes(home.path()), before, "{what}: the read never writes");
    }
}

/// AC-2: a saved entry is `true`, and so is a file the step cannot trust (it reports that
/// damage and is not forced open over it).
#[test]
fn a_saved_entry_and_a_damaged_file_read_as_saved() {
    let saved = r#"{"launcher":{"models":{"claude":{"leader":{"model":"claude-opus-5-5","effort":"high"}}}}}"#;
    for (what, settings) in [
        ("a saved entry", saved),
        (
            "models that is not an object",
            r#"{"launcher":{"models":"x"}}"#,
        ),
        ("a launcher that is not an object", r#"{"launcher":3}"#),
        ("text that is not JSON", "not json"),
    ] {
        let home = home_with(Some(settings));
        let before = bytes(home.path());
        assert!(models_saved(home.path()), "{what}");
        assert_eq!(bytes(home.path()), before, "{what}: the read never writes");
    }
    // A folder where the file should be is unreadable settings, so it counts as saved too.
    let home = TempDir::new().unwrap();
    fs::create_dir_all(settings_path(home.path())).unwrap();
    assert!(models_saved(home.path()), "a folder in place of the file");
}
