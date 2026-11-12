//! Shared secret-shape detection and message redaction (Issue #1996, 03-B owns the
//! module; scope decision 2026-11-08).
//!
//! One shared list, one predicate and one redacting formatter so callers refuse or
//! mask the same shapes instead of restating the list. The list is the union of the
//! conservative profile shapes (`crate::SECRET_SHAPES`, lib.rs) and the merged gh
//! token prefixes (`crate::forge::transport::TOKEN_PREFIXES`) — the prefixes stay
//! spelled only where they were already merged (#2351). forge::report consumes
//! this module's predicate instead of carrying its own copy. It is a conservative shape filter, not a secret detector; the common
//! redaction policy (what may be logged at all) belongs to the caller.

use crate::forge::transport::TOKEN_PREFIXES;

/// The shared shape list as one sequence: the conservative profile shapes plus the
/// merged gh token prefixes (#2351). Tests build their synthetic tokens from this
/// at runtime — no caller restates the prefixes.
pub fn secret_shapes() -> Vec<&'static str> {
    crate::SECRET_SHAPES
        .iter()
        .chain(TOKEN_PREFIXES.iter())
        .copied()
        .collect()
}

/// True when `s` carries a credential-shaped substring (lower-cased match).
pub fn contains_secret_shape(s: &str) -> bool {
    let lower = s.to_ascii_lowercase();
    crate::SECRET_SHAPES
        .iter()
        .chain(TOKEN_PREFIXES.iter())
        .any(|shape| lower.contains(shape))
}

/// Replaces every credential-shaped key and the value that follows it (up to the
/// next whitespace or the end of the string) with `[redacted]`, so diagnostics can
/// quote the surrounding context without carrying the secret value.
pub fn redact_secret_shapes(s: &str) -> String {
    let mut out = s.to_string();
    for shape in crate::SECRET_SHAPES.iter().chain(TOKEN_PREFIXES.iter()) {
        loop {
            let lower = out.to_ascii_lowercase();
            let Some(start) = lower.find(shape) else {
                break;
            };
            let value_end = out[start..]
                .find(char::is_whitespace)
                .map(|i| start + i)
                .unwrap_or(out.len());
            out.replace_range(start..value_end, "[redacted]");
        }
    }
    out
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn detects_the_shared_shapes_without_echoing() {
        assert!(contains_secret_shape("token=abc123"));
        assert!(contains_secret_shape("see PASSWORD=topsecret"));
        assert!(!contains_secret_shape("plain value"));
        let line = "failed at token=abc123 near the end";
        let redacted = redact_secret_shapes(line);
        assert!(!redacted.contains("abc123"), "{redacted}");
        assert!(redacted.contains("[redacted]"));
        assert!(redacted.contains("near the end"));
    }

    #[test]
    fn the_merged_gh_prefixes_are_part_of_the_shared_list() {
        // The four gh-shaped prefixes (start "gh", end "_", four chars) from the
        // merged transport list must be known to the shared predicate — spelled
        // here only through the list itself, never as literals.
        let gh: Vec<&str> = secret_shapes()
            .iter()
            .filter(|s| s.starts_with("gh") && s.ends_with('_') && s.len() == 4)
            .copied()
            .collect();
        assert!(
            gh.len() >= 4,
            "the merged gh token prefixes must be part of the shared shapes: {gh:?}"
        );
        for prefix in gh {
            assert!(contains_secret_shape(&format!("{prefix}synthetic-value")));
        }
    }
}
