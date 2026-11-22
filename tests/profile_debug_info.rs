// The dev and test profiles pin line-tables-only debug info for this crate and none for its
// dependencies (SPEC-20261121-debug-line-tables-only): file and line numbers stay in this crate's
// backtrace frames while variable and type info is dropped, which keeps the build folder and its
// file-event load small. The release-side profiles stay untouched.

use std::fs;

fn manifest() -> String {
    let manifest_dir = env!("CARGO_MANIFEST_DIR");
    fs::read_to_string(format!("{manifest_dir}/Cargo.toml")).expect("Cargo.toml reads")
}

/// The body of one `[header]` table, comparing header lines exactly so that
/// `[profile.dev]` never matches the `[profile.dev.package.sha2]` line.
fn table(toml: &str, header: &str) -> Option<String> {
    let lines: Vec<&str> = toml.lines().collect();
    let start = lines.iter().position(|line| line.trim() == header)?;
    let end = lines[start + 1..]
        .iter()
        .position(|line| line.trim().starts_with('['))
        .map(|i| i + start + 1)
        .unwrap_or(lines.len());
    Some(lines[start..end].join("\n"))
}

#[test]
fn dev_and_test_profiles_pin_line_tables_only() {
    let toml = manifest();
    for header in ["[profile.dev]", "[profile.test]"] {
        let body =
            table(&toml, header).unwrap_or_else(|| panic!("Cargo.toml lacks the {header} table"));
        assert!(
            body.contains("debug = \"line-tables-only\""),
            "{header} must pin debug = \"line-tables-only\" so backtraces keep file and line numbers"
        );
    }
    for header in [
        "[profile.dev.package.\"*\"]",
        "[profile.test.package.\"*\"]",
    ] {
        let body =
            table(&toml, header).unwrap_or_else(|| panic!("Cargo.toml lacks the {header} table"));
        assert!(
            body.contains("debug = false"),
            "{header} must build dependencies without debug info"
        );
    }
}

#[test]
fn dist_profile_and_sha2_override_stay_untouched() {
    let toml = manifest();
    let dist = table(&toml, "[profile.dist]").expect("Cargo.toml keeps [profile.dist]");
    assert!(
        dist.contains("inherits = \"release\""),
        "dist keeps its release inheritance"
    );
    assert!(dist.contains("lto = \"thin\""), "dist keeps thin lto");
    let sha2 = table(&toml, "[profile.dev.package.sha2]")
        .expect("Cargo.toml keeps the sha2 dev-package override");
    assert!(
        sha2.contains("opt-level = 3"),
        "sha2 stays optimized in dev builds (issue #2507)"
    );
    assert!(
        !toml.contains("\n[profile.release]"),
        "an explicit [profile.release] table must not appear"
    );
}
