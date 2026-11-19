//! Write-scope path property tests (C-100, SPEC-20261119-property-tests): for generated
//! relative paths with `.` steps, `..` climbs, repeated separators and mixed case — checked
//! against a symlinked temp tree — the lexical scope check never allows a path whose
//! resolution leaves the scope, and its answer does not depend on the filesystem.

#[path = "common/prop.rs"]
mod prop;

use moai_cli::ledger::reservation::Reservation;
use moai_cli::ledger::workflow::{covering_pattern, covers};
use prop::{run_property, Rng, MIN_CASES};
use std::fs;
use tempfile::TempDir;

/// Scope patterns with a known inside shape: the literal prefix, and whether the last segment
/// must end in `.rs` (the `*.rs` pattern matches one file name).
const PATTERNS: &[(&str, &[&str], bool)] = &[
    ("src/**", &["src"], false),
    ("a/b/**", &["a", "b"], false),
    ("docs", &["docs"], false),
    ("tests/*.rs", &["tests"], true),
];

fn reservation_with(pattern: &str) -> Reservation {
    Reservation {
        reservation_id: 1,
        card_id: 1,
        generation: 1,
        holder: "prop".into(),
        status: "held".into(),
        acquired_at: 0,
        lease_until: i64::MAX,
        updated_at: 0,
        worktree: None,
        branch: None,
        write_scope: vec![pattern.to_string()],
    }
}

/// A generated case: the pattern the path was built against (and its prefix), the path, and
/// the answer the check owes — `expected` comes from how the path was built, never from
/// calling `covers`.
struct Generated {
    pattern: &'static str,
    prefix: &'static [&'static str],
    path: String,
    expected: bool,
}

fn generate(rng: &mut Rng) -> Generated {
    let (pattern, prefix, rs) = *rng.pick(PATTERNS);
    let mut segments: Vec<String> = prefix.iter().map(|s| s.to_string()).collect();
    // Extra folder names stay under `**`; the `*.rs` pattern matches one file name directly
    // under its prefix, so it takes none.
    let extras = if rs { 0 } else { rng.below(3) };
    for _ in 0..extras {
        segments.push(rng.name(8));
    }
    if rs {
        segments.push(format!("{}.rs", rng.name(6)));
    }
    // In-scope noise: `.` steps and a `x/..` pair inside the prefix stay inside.
    if rng.bool() {
        let at = 1 + rng.below(segments.len() as u64) as usize;
        segments.insert(at.min(segments.len()), ".".into());
    }
    if rng.bool() {
        let at = 1 + rng.below(segments.len() as u64) as usize;
        let name = rng.name(4);
        segments.insert(at.min(segments.len()), name.clone());
        segments.insert(at.min(segments.len()) + 1, "..".into());
    }
    // Case noise: an upper-case letter of a middle segment (the check is case-insensitive).
    if rng.bool() && segments.len() > 1 {
        let at = 1 + rng.below(segments.len() as u64 - 1) as usize;
        segments[at] = segments[at].to_uppercase();
    }
    let expected = true;
    if rng.bool() {
        // An escape: climb out of everything (extras, the prefix, one step more).
        for _ in 0..extras + prefix.len() as u64 + 1 {
            segments.push("..".into());
        }
        segments.push(rng.name(6));
        return Generated {
            pattern,
            prefix,
            path: segments.join("/"),
            expected: false,
        };
    }
    if rng.bool() {
        // A doubled separator stays one separator — only where a separator exists to double.
        let joined = segments.join("/");
        if let Some(at) = joined.find('/') {
            return Generated {
                pattern,
                prefix,
                path: format!("{}//{}", &joined[..at], &joined[at + 1..]),
                expected,
            };
        }
    }
    if rng.bool() {
        // A path under no pattern's prefix at all.
        return Generated {
            pattern,
            prefix,
            path: format!("{}/{}", rng.name(5), rng.name(5)),
            expected: false,
        };
    }
    Generated {
        pattern,
        prefix,
        path: segments.join("/"),
        expected,
    }
}

/// REQ-9, REQ-10: the check answers by generation, never allows an escape, and its answer does
/// not change when the tree (with symlinks over the in-scope prefixes) exists.
#[test]
fn the_scope_check_never_allows_an_escape_and_ignores_the_tree() {
    run_property("scope paths", MIN_CASES, |rng| {
        let case = generate(rng);
        let (pattern, prefix) = (case.pattern, case.prefix);
        let reservation = reservation_with(pattern);
        let answer = |path: &str| covering_pattern(&reservation, path).is_some();

        // Before any tree exists: the promised answer, and noise-invariance of the lowercase.
        assert_eq!(
            answer(&case.path),
            case.expected,
            "path {} against {}",
            case.path,
            pattern
        );
        assert_eq!(answer(&case.path), answer(&case.path.to_lowercase()));

        // The tree: the prefix as a real folder, or as a symlink out of the tree.
        let tree = TempDir::new().unwrap();
        let outside = TempDir::new().unwrap();
        let root = tree.path();
        let first = root.join(prefix[0]);
        if rng.bool() {
            fs::create_dir_all(&first).unwrap();
        } else {
            #[cfg(unix)]
            std::os::unix::fs::symlink(outside.path(), &first).unwrap();
            #[cfg(not(unix))]
            fs::create_dir_all(&first).unwrap();
        }
        // `covers` reads no file: the same strings answer the same with the tree present.
        assert_eq!(
            answer(&case.path),
            case.expected,
            "with the tree present: {} against {}",
            case.path,
            pattern
        );
        // An escaping path is refused whether or not anything exists on disk; the marker
        // segment never names a prefix segment, so the climb truly leaves the pattern.
        let escape = format!("{}/../escape-marker-{}", prefix.join("/"), rng.name(4));
        assert!(!covers(pattern, &escape), "{escape} against {pattern}");
    });
}
