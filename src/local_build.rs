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
    Ok(match number(raw)? {
        Some(n) => format!("{package}+local.{n}"),
        None => package.to_string(),
    })
}

/// The variable a CI build sets to its build number (SPEC-20261117-headver-ci-number): the
/// workflow's run number on the run's first attempt, nothing on a re-run.
pub const CI_ENV: &str = "MOAI_BUILD_NUMBER";

/// The local build number in the raw `MOAI_LOCAL_BUILD` value: `None` when unset or empty,
/// otherwise 1 to [`MAX`] in plain decimal digits (card C-67 S1: the HeadVer `{build}`).
pub fn number(raw: Option<&str>) -> Result<Option<u32>, String> {
    number_of(ENV, raw)
}

/// [`number`] for the variable `env` ([`ENV`] or [`CI_ENV`]): the same values, and a refusal
/// that names `env`.
pub fn number_of(env: &str, raw: Option<&str>) -> Result<Option<u32>, String> {
    let Some(raw) = raw.filter(|r| !r.is_empty()) else {
        return Ok(None);
    };
    let digits = raw.bytes().all(|b| b.is_ascii_digit());
    let number = if digits && !raw.starts_with('0') {
        raw.parse::<u32>().ok()
    } else {
        None
    };
    number.map(Some).ok_or_else(|| {
        format!(
            "{env}={raw:?} refused: use a build number from 1 to {MAX} in decimal digits, \
             without a sign or leading zero"
        )
    })
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

    /// SPEC-20261117-headver-ci-number: the CI number takes the same values and its refusal
    /// names its own variable.
    #[test]
    fn the_ci_number_has_the_same_bounds_and_names_its_variable() {
        use super::{number_of, CI_ENV, MAX};
        assert_eq!(CI_ENV, "MOAI_BUILD_NUMBER");
        assert_eq!(number_of(CI_ENV, None).unwrap(), None);
        assert_eq!(number_of(CI_ENV, Some("")).unwrap(), None);
        assert_eq!(number_of(CI_ENV, Some("1")).unwrap(), Some(1));
        assert_eq!(number_of(CI_ENV, Some("4294967295")).unwrap(), Some(MAX));
        for raw in ["0", "4294967296", "01", "x"] {
            let err = number_of(CI_ENV, Some(raw)).unwrap_err();
            assert!(err.starts_with("MOAI_BUILD_NUMBER="), "{raw:?}: {err}");
        }
    }

    #[test]
    fn anything_but_plain_decimal_digits_is_refused() {
        for raw in ["01", "+1", "-1", " 1", "1 ", "1.0", "1e3", "x", "rc.2", "١"] {
            let err = version(PKG, Some(raw)).unwrap_err();
            assert!(err.contains("MOAI_LOCAL_BUILD"), "{raw:?}: {err}");
        }
    }
}
