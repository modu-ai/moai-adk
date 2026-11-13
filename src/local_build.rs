//! The version a local test build reports (#2722 AC-5, leader decision 6855056793): the
//! package version, plus build metadata `+local.<n>` when the build sets `MOAI_LOCAL_BUILD=<n>`.
//!
//! Build metadata does not change semver precedence and never names a release tag, so every
//! local build gets its own visible number without colliding with `rc.<n>` tags. `build.rs`
//! includes this file and sets `MOAI_BUILD_VERSION`; the crate reads it as
//! `crate::BUILD_VERSION`, which `--version` and the status line show. Install and update
//! records keep the plain package version, because they compare it with the skill package's.
//! A value that is not a positive decimal integer stops the build instead of being dropped.

/// The environment variable a local test build sets to its build number.
pub const ENV: &str = "MOAI_LOCAL_BUILD";

/// The largest accepted build number (`u32::MAX`).
pub const MAX: u32 = u32::MAX;

/// The version string for `package` and the raw `MOAI_LOCAL_BUILD` value, or why it is refused.
///
/// Unset or empty means no local number: the package version alone. Otherwise the value must be
/// 1 to [`MAX`] in plain decimal digits with no sign, no leading zero and no spaces.
pub fn version(package: &str, raw: Option<&str>) -> Result<String, String> {
    let Some(raw) = raw.filter(|r| !r.is_empty()) else {
        return Ok(package.to_string());
    };
    let digits = raw.bytes().all(|b| b.is_ascii_digit());
    let number = if digits && !raw.starts_with('0') {
        raw.parse::<u32>().ok()
    } else {
        None
    };
    match number {
        Some(n) => Ok(format!("{package}+local.{n}")),
        None => Err(format!(
            "{ENV}={raw:?} refused: use a build number from 1 to {MAX} in decimal digits, \
             without a sign or leading zero"
        )),
    }
}

#[cfg(test)]
mod tests {
    use super::version;

    const PKG: &str = "2026.11.0-rc.1";

    #[test]
    fn unset_or_empty_keeps_the_package_version() {
        assert_eq!(version(PKG, None).unwrap(), "2026.11.0-rc.1");
        assert_eq!(version(PKG, Some("")).unwrap(), "2026.11.0-rc.1");
    }

    #[test]
    fn a_build_number_becomes_local_build_metadata() {
        assert_eq!(version(PKG, Some("1")).unwrap(), "2026.11.0-rc.1+local.1");
        assert_eq!(version(PKG, Some("27")).unwrap(), "2026.11.0-rc.1+local.27");
    }

    /// Boundaries as literals, so a changed bound fails here (N accepted, N+1 refused).
    #[test]
    fn the_bounds_are_one_and_u32_max() {
        assert!(version(PKG, Some("0")).is_err());
        assert_eq!(version(PKG, Some("1")).unwrap(), "2026.11.0-rc.1+local.1");
        assert_eq!(
            version(PKG, Some("4294967295")).unwrap(),
            "2026.11.0-rc.1+local.4294967295"
        );
        assert!(version(PKG, Some("4294967296")).is_err());
    }

    #[test]
    fn anything_but_plain_decimal_digits_is_refused() {
        for raw in ["01", "+1", "-1", " 1", "1 ", "1.0", "1e3", "x", "rc.2", "١"] {
            let err = version(PKG, Some(raw)).unwrap_err();
            assert!(err.contains("MOAI_LOCAL_BUILD"), "{raw:?}: {err}");
        }
    }
}
