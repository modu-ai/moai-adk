//! SPEC-20261122-bin-test-false: `src/bin/moai.rs` loads `src/main.rs` as a module, so a test
//! harness for the `moai` binary would build and run the same `main.rs` tests a second time.
//! The `moai` binary target is therefore not a test target; `moai-cli` stays one and runs those
//! tests once. A test written in `src/bin/moai.rs` itself would not run: keep that file a shell.

use std::fs;
use toml_edit::DocumentMut;

fn bin_tables() -> Vec<(String, Option<bool>)> {
    let text = fs::read_to_string(concat!(env!("CARGO_MANIFEST_DIR"), "/Cargo.toml"))
        .expect("Cargo.toml reads");
    let doc: DocumentMut = text.parse().expect("Cargo.toml parses");
    doc["bin"]
        .as_array_of_tables()
        .expect("Cargo.toml has [[bin]] tables")
        .iter()
        .map(|t| {
            let name = t["name"].as_str().expect("a [[bin]] name").to_string();
            let test = t
                .get("test")
                .map(|v| v.as_bool().expect("test is a boolean"));
            (name, test)
        })
        .collect()
}

#[test]
fn the_moai_binary_is_not_a_test_target_and_moai_cli_is() {
    let bins = bin_tables();
    let test_of = |name: &str| {
        bins.iter()
            .find(|(n, _)| n == name)
            .unwrap_or_else(|| panic!("no [[bin]] named {name}: {bins:?}"))
            .1
    };
    assert_eq!(
        test_of("moai"),
        Some(false),
        "[[bin]] moai must set test = false so the main.rs tests are not built twice"
    );
    assert_ne!(
        test_of("moai-cli"),
        Some(false),
        "[[bin]] moai-cli must stay a test target: it runs the main.rs tests"
    );
}
