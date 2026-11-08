//! Shared secret-shape detection and message redaction (Issue #1996, 03-B owns the
//! module; scope decision 2026-11-08).
//!
//! The shape list itself stays `crate::SECRET_SHAPES` in lib.rs — forge::promote and
//! the profile filter already consume it, and moving the constant would touch files
//! outside this Issue's scope. This module is the shared read side: one predicate and
//! one redacting formatter so callers refuse or mask the same shapes instead of
//! restating the list. It is a conservative shape filter, not a secret detector; the
//! common redaction policy (what may be logged at all) belongs to the caller.

/// True when `s` carries a credential-shaped substring (lower-cased match).
pub fn contains_secret_shape(s: &str) -> bool {
    let lower = s.to_ascii_lowercase();
    crate::SECRET_SHAPES
        .iter()
        .any(|shape| lower.contains(shape))
}

/// Replaces every credential-shaped key and the value that follows it (up to the
/// next whitespace or the end of the string) with `[redacted]`, so diagnostics can
/// quote the surrounding context without carrying the secret value.
pub fn redact_secret_shapes(s: &str) -> String {
    let mut out = s.to_string();
    for shape in crate::SECRET_SHAPES {
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
}
