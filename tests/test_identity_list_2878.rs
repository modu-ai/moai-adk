//! #2878 audit N3 (6873276636): every identity email the repository's code and tests set —
//! `user.email`, `GIT_AUTHOR_EMAIL`, `GIT_COMMITTER_EMAIL` in `src/`, `tests/` and `tools/` (Rust
//! and Python) — is on the commit-identity guard's
//! test-address list, so a test identity that leaks into a real checkout is refused. The guard's
//! own tests use a few real-looking addresses on purpose; they are named here with the reason.
use moai_cli::hook_guards::TEST_IDENTITY_EMAILS;
use std::{fs, path::Path};

/// Addresses the commit-identity tests use as the *allowed* identity or as a `deny_emails`
/// entry; they must not be on the built-in list.
const GUARD_TEST_IDENTITIES: &[&str] = &["me@moai.dev", "someone@moai.dev", "ci@corp.example"];

const KEYS: &[&str] = &["user.email", "GIT_AUTHOR_EMAIL", "GIT_COMMITTER_EMAIL"];

/// The emails written right after one of [`KEYS`] (`"user.email", "x@y"`, `user.email=x@y`,
/// `("GIT_AUTHOR_EMAIL", "x@y")`), lowercased.
fn emails_in(text: &str) -> Vec<String> {
    let mut out = Vec::new();
    for key in KEYS {
        for (at, _) in text.match_indices(key) {
            let rest = &text[at + key.len()..];
            let rest = rest.trim_start_matches(|c: char| "\"', =:)(\\".contains(c));
            let word: String = rest
                .chars()
                .take_while(|c| c.is_ascii_alphanumeric() || "._%+-@".contains(*c))
                .collect();
            if word.contains('@') {
                out.push(word.to_ascii_lowercase());
            }
        }
    }
    out
}

fn source_files(dir: &Path, out: &mut Vec<std::path::PathBuf>) {
    for entry in fs::read_dir(dir).unwrap() {
        let path = entry.unwrap().path();
        if path.is_dir() {
            source_files(&path, out);
        } else if path.extension().is_some_and(|e| e == "rs" || e == "py") {
            out.push(path);
        }
    }
}

#[test]
fn every_identity_the_tests_set_is_a_listed_test_address() {
    let root = Path::new(env!("CARGO_MANIFEST_DIR"));
    let mut files = Vec::new();
    source_files(&root.join("src"), &mut files);
    source_files(&root.join("tests"), &mut files);
    source_files(&root.join("tools"), &mut files);
    let mut seen = 0;
    let mut unlisted = Vec::new();
    for file in files {
        // This file's own reader examples are not identities.
        if file
            .file_name()
            .is_some_and(|n| n == "test_identity_list_2878.rs")
        {
            continue;
        }
        let text = fs::read_to_string(&file).unwrap();
        for email in emails_in(&text) {
            seen += 1;
            if !TEST_IDENTITY_EMAILS.contains(&email.as_str())
                && !GUARD_TEST_IDENTITIES.contains(&email.as_str())
            {
                unlisted.push(format!("{}: {email}", file.display()));
            }
        }
    }
    assert!(seen > 20, "the scan found only {seen} identities");
    assert!(
        unlisted.is_empty(),
        "add these test identities to TEST_IDENTITY_EMAILS (src/hook_guards.rs):\n{}",
        unlisted.join("\n")
    );
}

/// The real-looking guard-test identities are never on the refused list.
#[test]
fn guard_test_identities_stay_off_the_list() {
    for e in GUARD_TEST_IDENTITIES {
        assert!(!TEST_IDENTITY_EMAILS.contains(e), "{e}");
    }
}

#[test]
fn the_reader_takes_the_value_after_each_key() {
    assert_eq!(
        emails_in(r#"git(&["config", "user.email", "A@B.invalid"]); "GIT_AUTHOR_EMAIL=c@d""#),
        ["a@b.invalid", "c@d"]
    );
    assert!(emails_in("user.email is unset").is_empty());
}
