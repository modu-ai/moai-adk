//! SPEC-20261121-nextest-local-checks: the project's nextest profile groups the binaries that
//! start the product binary, ends slow tests with their process group and retries twice, and
//! keeps the earlier rule for `ac_command_lint`. The file is read, never run: no nextest here.

use toml_edit::{DocumentMut, Item};

fn config() -> DocumentMut {
    include_str!("../.config/nextest.toml")
        .parse::<DocumentMut>()
        .expect(".config/nextest.toml is valid TOML")
}

/// The `[[profile.default.overrides]]` tables, in file order.
fn overrides(doc: &DocumentMut) -> Vec<&toml_edit::Table> {
    doc["profile"]["default"]["overrides"]
        .as_array_of_tables()
        .expect("profile.default.overrides is an array of tables")
        .iter()
        .collect()
}

fn text<'a>(t: &'a toml_edit::Table, key: &str) -> &'a str {
    t.get(key).and_then(Item::as_str).unwrap_or_default()
}

#[test]
fn the_profile_groups_the_spawning_binaries_ends_slow_tests_and_retries_twice() {
    let doc = config();

    // One test group of at most two threads, and the overrides that put the spawning binaries
    // into it (glob matchers: the default matcher of `binary()`).
    let groups = doc["test-groups"].as_table().expect("[test-groups]");
    assert_eq!(groups.len(), 1, "one group: {groups}");
    let (name, group) = groups.iter().next().unwrap();
    let max = group["max-threads"].as_integer().expect("max-threads");
    assert_eq!(max, 2, "{name} caps at two threads");
    let grouped: Vec<&str> = overrides(&doc)
        .into_iter()
        .filter(|o| text(o, "test-group") == name)
        .map(|o| text(o, "filter"))
        .collect();
    assert_eq!(grouped.len(), 1, "one override uses the group: {grouped:?}");
    for binary in [
        "binary(home_*)",
        "binary(update_*_cli)",
        "binary(statusline_golden)",
    ] {
        assert!(grouped[0].contains(binary), "{binary} in {}", grouped[0]);
    }

    // A test that runs past five periods of 60 seconds ends with its process group; the
    // default profile holds the setting (nextest signals the group, then waits its grace period).
    let slow = &doc["profile"]["default"]["slow-timeout"];
    assert_eq!(slow["period"].as_str(), Some("60s"), "{slow}");
    assert_eq!(slow["terminate-after"].as_integer(), Some(5), "{slow}");

    // A failed test is retried twice, so that nextest can label a pass on a retry FLAKY.
    assert_eq!(
        doc["profile"]["default"]["retries"].as_integer(),
        Some(2),
        "retries = 2"
    );
    // A flaky pass stays a pass here: the runner counts it in the summary instead (D-3).
    assert!(
        doc["profile"]["default"].get("flaky-result").is_none(),
        "flaky-result is not set"
    );

    // The earlier rule (every test thread to `ac_command_lint`) is unchanged.
    let lint = overrides(&doc)
        .into_iter()
        .find(|o| text(o, "filter") == "binary(ac_command_lint)")
        .expect("the ac_command_lint override stays");
    assert_eq!(text(lint, "threads-required"), "num-test-threads");

    // A longer limit for the one binary whose single test takes minutes under load.
    let long = overrides(&doc)
        .into_iter()
        .find(|o| {
            text(o, "filter") == "binary(statusline_golden)" && o.get("slow-timeout").is_some()
        })
        .expect("statusline_golden has its own, longer slow-timeout");
    let terminate = long["slow-timeout"]["terminate-after"]
        .as_integer()
        .expect("terminate-after");
    assert!(
        terminate > 5,
        "longer than the default five periods: {terminate}"
    );
}
