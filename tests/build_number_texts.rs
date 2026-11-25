//! SPEC-20261125-install-build-number, AC-5: the shipped versioning text says that an install
//! build takes its number from `moai build-number next`, that a build which only checks a change
//! takes none, and that nobody picks a number by hand. The test reads the skill file of this
//! checkout; nothing is run and no HOME is touched.

use std::{fs, path::Path};

fn versioning() -> String {
    let path =
        Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai/references/internal/versioning.md");
    fs::read_to_string(&path).unwrap_or_else(|e| panic!("{}: {e}", path.display()))
}

/// The one bullet that describes a HeadVer development build.
fn dev_build_bullet(text: &str) -> String {
    text.lines()
        .find(|l| l.contains("HeadVer, a development build"))
        .expect("the versioning text has a HeadVer development build bullet")
        .to_string()
}

#[test]
fn an_install_build_takes_its_number_from_the_command() {
    let bullet = dev_build_bullet(&versioning());
    assert!(
        bullet.contains("`moai build-number next`"),
        "the development build bullet names no command that hands out the number: {bullet}"
    );
    assert!(
        bullet.contains("a build you install"),
        "the bullet does not say which build takes the number: {bullet}"
    );
}

#[test]
fn a_build_that_only_checks_takes_no_number() {
    let bullet = dev_build_bullet(&versioning());
    assert!(
        bullet.contains("a build that only checks a change takes no number"),
        "{bullet}"
    );
    assert!(bullet.contains("`unnumbered build`"), "{bullet}");
}

#[test]
fn no_text_tells_a_person_to_pick_a_number() {
    let text = versioning();
    for pick in ["pick a number", "choose a number", "pick the next number"] {
        assert!(!text.contains(pick), "the text says `{pick}`");
    }
}
