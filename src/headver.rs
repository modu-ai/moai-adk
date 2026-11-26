//! HeadVer build versions (card C-67 slice 1, SPEC-20261117-headver-version):
//! `{head}.{yearweek}.{build}` (github.com/line/headver; rules in
//! `skills/moai/references/internal/versioning.md` and `AGENTS.local.md`).
//!
//! - `head`: the one manual number, read from the tracked one-line file [`HEAD_FILE`].
//! - `yearweek`: the two-digit ISO-8601 week-year and the two-digit ISO week of the UTC build
//!   date, always four digits (2026-11-17 is `2647`, 2027-01-01 is `2653`, 2100-01-04 is
//!   `0001`: the two-digit year wraps in 2100, so `head` must rise before then to keep order).
//! - `build`: the build number; `MOAI_LOCAL_BUILD` for a local build, 0 for an unnumbered build.
//!
//! A build that is not a release ends in `-dev` (`4.2648.26-dev`, [`dev_version`]). A release
//! build (`MOAI_RELEASE_VERSION`, [`release_build`]) is `{head}.{release week}.{serial}` with the
//! serial counting from 0 inside its week and no suffix (`4.2649.0`), made on the branch
//! `release/v{head}.{yearweek}.{n}` (SPEC-20261125-headver-weekly-release).
//!
//! `build.rs` includes this file with `#[path]` to compute the build's version, so it uses
//! the standard library only. The library uses [`compare`] to order versions: every HeadVer
//! version is newer than every version of the earlier `YYYY.M.P[-rc.N]` scheme.

use std::cmp::Ordering;

/// The tracked file that holds `head`, at the repository root.
pub const HEAD_FILE: &str = "HEADVER";
/// The largest head (the profile's `versioning.head` limit).
pub const MAX_HEAD: u32 = 9999;

/// `head` from the text of [`HEAD_FILE`]: one decimal number from 0 to [`MAX_HEAD`], no sign,
/// no leading zero, optionally followed by one line ending.
pub fn parse_head(text: &str) -> Result<u32, String> {
    let line = text
        .strip_suffix("\r\n")
        .or_else(|| text.strip_suffix('\n'))
        .unwrap_or(text);
    let digits = !line.is_empty()
        && line.bytes().all(|b| b.is_ascii_digit())
        && (line == "0" || !line.starts_with('0'));
    match line.parse::<u32>() {
        Ok(head) if digits && head <= MAX_HEAD => Ok(head),
        _ => Err(format!(
            "{HEAD_FILE} refused: it must hold one line with the HeadVer head, a decimal number \
             from 0 to {MAX_HEAD} without a sign or leading zero"
        )),
    }
}

/// Days since 1970-01-01 of a proleptic Gregorian date (H. Hinnant's `days_from_civil`).
fn days_from_civil(y: i64, m: u32, d: u32) -> i64 {
    let y = if m <= 2 { y - 1 } else { y };
    let era = y.div_euclid(400);
    let yoe = y - era * 400;
    let m = i64::from(m);
    let doy = (153 * (if m > 2 { m - 3 } else { m + 9 }) + 2) / 5 + i64::from(d) - 1;
    let doe = yoe * 365 + yoe / 4 - yoe / 100 + doy;
    era * 146_097 + doe - 719_468
}

/// The proleptic Gregorian date of a day count since 1970-01-01 (`civil_from_days`).
pub fn civil_from_days(days: i64) -> (i64, u32, u32) {
    let z = days + 719_468;
    let era = z.div_euclid(146_097);
    let doe = z - era * 146_097;
    let yoe = (doe - doe / 1460 + doe / 36_524 - doe / 146_096) / 365;
    let doy = doe - (365 * yoe + yoe / 4 - yoe / 100);
    let mp = (5 * doy + 2) / 153;
    let d = (doy - (153 * mp + 2) / 5 + 1) as u32;
    let m = if mp < 10 { mp + 3 } else { mp - 9 } as u32;
    let y = yoe + era * 400 + i64::from(m <= 2);
    (y, m, d)
}

/// The ISO-8601 week-year and week of a date: the week runs Monday to Sunday and belongs to
/// the year that holds its Thursday.
pub fn iso_week(y: i64, m: u32, d: u32) -> (i64, u32) {
    let days = days_from_civil(y, m, d);
    // 1970-01-01 was a Thursday: Monday is 1, Sunday is 7.
    let weekday = (days + 3).rem_euclid(7) + 1;
    let thursday = days + 4 - weekday;
    let (year, _, _) = civil_from_days(thursday);
    let week = (thursday - days_from_civil(year, 1, 1)) / 7 + 1;
    (year, week as u32)
}

/// `yearweek`: the ISO week-year's last two digits, then the two-digit ISO week.
pub fn yearweek(y: i64, m: u32, d: u32) -> u32 {
    let (year, week) = iso_week(y, m, d);
    year.rem_euclid(100) as u32 * 100 + week
}

/// The version text: `{head}.{yearweek:04}.{build}`.
pub fn version(head: u32, yearweek: u32, build: u32) -> String {
    format!("{head}.{yearweek:04}.{build}")
}

/// The suffix every build that is not a release carries (SPEC-20261125-headver-weekly-release):
/// `4.2648.26-dev`. A release has none.
pub const DEV_SUFFIX: &str = "-dev";

/// The development build version: [`version`] with [`DEV_SUFFIX`].
pub fn dev_version(head: u32, yearweek: u32, build: u32) -> String {
    format!("{}{DEV_SUFFIX}", version(head, yearweek, build))
}

/// What one build is: its version and the second line `--version` prints.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Build {
    /// `4.2647.5`: what `--version`, the status line and update records show and compare, and
    /// the name of the build's `versions/<version>/` folder.
    pub version: String,
    /// `local build 5 (MOAI_LOCAL_BUILD), built 2026-11-17 UTC (ISO week 2026-W47), package
    /// 2026.11.0-rc.1`, or the same with `unnumbered build …`.
    pub note: String,
}

/// This build from the text of [`HEAD_FILE`] (`None` when it is missing), the build time in
/// seconds since 1970-01-01 UTC, the local build number and the package version.
pub fn build(
    head_text: Option<&str>,
    epoch_secs: i64,
    local: Option<u32>,
    package: &str,
) -> Result<Build, String> {
    build_numbered(head_text, epoch_secs, local, None, package)
}

/// [`build`] with a CI build number too (`MOAI_BUILD_NUMBER`, SPEC-20261117-headver-ci-number):
/// the note says which source the number came from, so a CI build is told apart from a local
/// build of the same number; a build given both numbers is refused, never guessed.
pub fn build_numbered(
    head_text: Option<&str>,
    epoch_secs: i64,
    local: Option<u32>,
    ci: Option<u32>,
    package: &str,
) -> Result<Build, String> {
    let head_text = head_text.ok_or_else(|| {
        format!("{HEAD_FILE} missing: the repository root must hold the HeadVer head (one line)")
    })?;
    let head = parse_head(head_text)?;
    let (y, m, d) = civil_from_days(epoch_secs.div_euclid(86_400));
    let (week_year, week) = iso_week(y, m, d);
    let (number, kind) = match (local, ci) {
        (Some(_), Some(_)) => {
            return Err(
                "MOAI_LOCAL_BUILD and MOAI_BUILD_NUMBER are both set: a build has one number \
                 source; unset one"
                    .into(),
            )
        }
        (Some(n), None) => (n, format!("local build {n} (MOAI_LOCAL_BUILD)")),
        (None, Some(n)) => (n, format!("CI build {n} (MOAI_BUILD_NUMBER)")),
        (None, None) => (
            0,
            "unnumbered build (no MOAI_LOCAL_BUILD or MOAI_BUILD_NUMBER)".to_string(),
        ),
    };
    Ok(Build {
        version: dev_version(head, yearweek(y, m, d), number),
        note: format!(
            "{kind}, built {y:04}-{m:02}-{d:02} UTC (ISO week {week_year}-W{week:02}), \
             package {package}"
        ),
    })
}

/// The release version `release` (`MOAI_RELEASE_VERSION`, SPEC-20261125-headver-weekly-release)
/// split into its numbers: `{head}.{yearweek}.{serial}` with the head of [`HEAD_FILE`], a
/// four-digit yearweek whose week is 01 to 53, and a serial from 0 in decimal digits without a
/// leading zero that fits `u32`. No prefix, no suffix, no metadata and no blank.
fn parse_release(release: &str, head: u32) -> Result<u32, String> {
    let refuse = |why: &str| {
        Err(format!(
            "MOAI_RELEASE_VERSION={release:?} refused: {why}; a release version is \
             {{head}}.{{yearweek}}.{{serial}} with no prefix or suffix"
        ))
    };
    let parts: Vec<&str> = release.split('.').collect();
    let digits = |p: &str| !p.is_empty() && p.bytes().all(|b| b.is_ascii_digit());
    if parts.len() != 3 || !parts.iter().all(|p| digits(p)) {
        return refuse("it is not three numbers joined by dots");
    }
    if parts[0] != head.to_string() {
        return refuse(&format!("its head is not the {HEAD_FILE} head {head}"));
    }
    let week = parts[1].parse::<u32>().unwrap_or(0) % 100;
    if parts[1].len() != 4 || !(1..=53).contains(&week) {
        return refuse("the yearweek must be four digits with an ISO week from 01 to 53");
    }
    if parts[2] != "0" && parts[2].starts_with('0') {
        return refuse("the serial has a leading zero");
    }
    match parts[2].parse::<u32>() {
        Ok(serial) => Ok(serial),
        Err(_) => refuse("the serial is over 4294967295"),
    }
}

/// A release build (SPEC-20261125-headver-weekly-release): its version is `release`, taken from
/// `MOAI_RELEASE_VERSION` without a suffix, and its note says
/// `release build <serial> (MOAI_RELEASE_VERSION), built <date> UTC (ISO week …), package …`.
/// The date is the build's own, the week of the version is the release week. A release build is
/// refused when `release` is not shaped as [`parse_release`] says or does not carry the
/// `HEADVER` head, and when `MOAI_LOCAL_BUILD` or `MOAI_BUILD_NUMBER` is set too.
pub fn release_build(
    head_text: Option<&str>,
    epoch_secs: i64,
    release: &str,
    local: Option<u32>,
    ci: Option<u32>,
    package: &str,
) -> Result<Build, String> {
    let head_text = head_text.ok_or_else(|| {
        format!("{HEAD_FILE} missing: the repository root must hold the HeadVer head (one line)")
    })?;
    let head = parse_head(head_text)?;
    for (set, name) in [
        (local.is_some(), "MOAI_LOCAL_BUILD"),
        (ci.is_some(), "MOAI_BUILD_NUMBER"),
    ] {
        if set {
            return Err(format!(
                "MOAI_RELEASE_VERSION and {name} are both set: a release build has no other \
                 number source; unset {name}"
            ));
        }
    }
    let serial = parse_release(release, head)?;
    let (y, m, d) = civil_from_days(epoch_secs.div_euclid(86_400));
    let (week_year, week) = iso_week(y, m, d);
    Ok(Build {
        version: release.to_string(),
        note: format!(
            "release build {serial} (MOAI_RELEASE_VERSION), built {y:04}-{m:02}-{d:02} UTC \
             (ISO week {week_year}-W{week:02}), package {package}"
        ),
    })
}

/// How [`compare`] reads one version.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
enum Shape {
    /// `head.yearweek.build`; the flag is true for a `-dev` suffix, which ranks below the
    /// release of the same numbers. Any other `-suffix` names the same build and is ignored.
    HeadVer(u64, u64, u64, bool),
    /// The earlier `YYYY.M.P` scheme with an optional `-rc.N`.
    Earlier(u64, u64, u64, Option<u64>),
}

fn shape(v: &str) -> Option<Shape> {
    let v = v.strip_prefix('v').unwrap_or(v);
    // Build metadata (`+local.4`, `+a`) is not part of the version.
    let core = v.split('+').next().unwrap_or(v);
    let (numbers, pre) = match core.split_once('-') {
        Some((n, p)) => (n, Some(p)),
        None => (core, None),
    };
    let parts: Vec<&str> = numbers.split('.').collect();
    if parts.len() != 3
        || parts
            .iter()
            .any(|p| p.is_empty() || !p.bytes().all(|b| b.is_ascii_digit()))
    {
        return None;
    }
    let n = |i: usize| parts[i].parse::<u64>().ok();
    let (a, b, c) = (n(0)?, n(1)?, n(2)?);
    // The earlier scheme is a four-digit year from 2000 and a month 1–12 written in one or two
    // digits (audit B1): a HeadVer middle is always the four-digit yearweek, so a head of 2000
    // or more (`2000.2647.5`) stays HeadVer. The third part is a patch number of any size.
    if parts[0].len() == 4 && a >= 2000 && parts[1].len() <= 2 && (1..=12).contains(&b) {
        let rc = match pre {
            None => None,
            Some(p) => Some(p.strip_prefix("rc.")?.parse::<u64>().ok()?),
        };
        return Some(Shape::Earlier(a, b, c, rc));
    }
    let week = b % 100;
    let dev = pre == Some("dev");
    (parts[1].len() == 4 && (1..=53).contains(&week)).then_some(Shape::HeadVer(a, b, c, dev))
}

/// The order of two versions, or `None` when either is a shape this build does not read
/// (it is then never called newer or older). Every HeadVer version (head 0–9999) is newer than
/// every `YYYY.M.P[-rc.N]` version (`YYYY` ≥ 2000, `M` 1–12); HeadVer versions compare head, then yearweek, then build; the
/// earlier scheme compares its numbers, then a release above its `rc.N`, then `N`.
pub fn compare(a: &str, b: &str) -> Option<Ordering> {
    use Shape::{Earlier, HeadVer};
    Some(match (shape(a)?, shape(b)?) {
        (HeadVer(..), Earlier(..)) => Ordering::Greater,
        (Earlier(..), HeadVer(..)) => Ordering::Less,
        // Numbers decide; at the same numbers the release (no `-dev`) is above its `-dev`.
        (HeadVer(h1, y1, b1, d1), HeadVer(h2, y2, b2, d2)) => {
            (h1, y1, b1, !d1).cmp(&(h2, y2, b2, !d2))
        }
        (Earlier(y1, m1, p1, r1), Earlier(y2, m2, p2, r2)) => {
            (y1, m1, p1).cmp(&(y2, m2, p2)).then(match (r1, r2) {
                (None, None) => Ordering::Equal,
                (None, Some(_)) => Ordering::Greater,
                (Some(_), None) => Ordering::Less,
                (Some(x), Some(y)) => x.cmp(&y),
            })
        }
    })
}

/// `a` is newer than `b` (both read, and `a` orders above).
pub fn is_newer(a: &str, b: &str) -> bool {
    compare(a, b) == Some(Ordering::Greater)
}

#[cfg(test)]
mod tests {
    use super::*;

    /// Expected weeks from Python's `datetime.date.isocalendar()` (2026-11-17), an independent
    /// reading of ISO 8601.
    #[test]
    fn yearweek_follows_the_iso_week_year_at_every_boundary() {
        for (date, want) in [
            ((2026, 11, 17), 2647),
            ((2026, 12, 28), 2653),
            ((2026, 12, 31), 2653),
            ((2027, 1, 1), 2653),
            ((2027, 1, 3), 2653),
            ((2027, 1, 4), 2701),
            ((2025, 12, 29), 2601),
            ((2020, 12, 31), 2053),
            ((2021, 1, 3), 2053),
            ((2021, 1, 4), 2101),
            ((2099, 12, 31), 9953),
            ((2100, 1, 1), 9953),
            ((2100, 1, 4), 1),
            ((1970, 1, 1), 7001),
            ((2000, 2, 29), 9),
        ] {
            let (y, m, d) = date;
            assert_eq!(yearweek(y, m, d), want, "{date:?}");
        }
        assert_eq!(iso_week(2027, 1, 1), (2026, 53));
        assert_eq!(iso_week(2025, 12, 29), (2026, 1));
    }

    #[test]
    fn days_and_dates_round_trip() {
        for (days, date) in [
            (0, (1970, 1, 1)),
            (20_774, (2026, 11, 17)),
            (20_819, (2027, 1, 1)),
            (47_485, (2100, 1, 4)),
            (11_016, (2000, 2, 29)),
            (-1, (1969, 12, 31)),
        ] {
            assert_eq!(civil_from_days(days), date, "{days}");
            assert_eq!(days_from_civil(date.0, date.1, date.2), days, "{date:?}");
        }
    }

    #[test]
    fn the_version_has_four_yearweek_digits() {
        assert_eq!(version(4, 2647, 5), "4.2647.5");
        assert_eq!(version(4, 1, 0), "4.0001.0");
        assert_eq!(version(0, 9953, 4_294_967_295), "0.9953.4294967295");
    }

    /// The head file's bounds as literals (N accepted, N+1 refused) and its shapes.
    #[test]
    fn the_head_file_holds_one_number_from_0_to_9999() {
        assert_eq!(parse_head("4\n"), Ok(4));
        assert_eq!(parse_head("4"), Ok(4));
        assert_eq!(parse_head("4\r\n"), Ok(4));
        assert_eq!(parse_head("0\n"), Ok(0));
        assert_eq!(parse_head("9999\n"), Ok(9999));
        for bad in [
            "10000", "", "\n", "04", "+4", "-1", " 4", "4 ", "4\n\n", "4\n5\n", "four",
        ] {
            let err = parse_head(bad).unwrap_err();
            assert!(err.starts_with("HEADVER refused"), "{bad:?}: {err}");
        }
    }

    /// 2026-11-17 is day 20 774 (1_794_873_600 s); 2027-01-04 is day 20 822 (1_799_020_800 s).
    #[test]
    fn a_build_names_its_kind_date_and_week() {
        let t = 1_794_873_600 + 8 * 3600;
        let local = build(Some("4\n"), t, Some(5), "2026.11.0-rc.1").unwrap();
        assert_eq!(local.version, "4.2647.5-dev");
        assert_eq!(
            local.note,
            "local build 5 (MOAI_LOCAL_BUILD), built 2026-11-17 UTC (ISO week 2026-W47), \
             package 2026.11.0-rc.1"
        );
        let unnumbered = build(Some("4\n"), t, None, "2026.11.0-rc.1").unwrap();
        assert_eq!(unnumbered.version, "4.2647.0-dev");
        assert!(unnumbered
            .note
            .starts_with("unnumbered build (no MOAI_LOCAL_BUILD or MOAI_BUILD_NUMBER), built"));
        // One second before midnight UTC on 2027-01-03 (Sunday) is still week 53 of 2026.
        let sunday = 1_799_020_800 - 1;
        assert_eq!(
            build(Some("4"), sunday, Some(1), "p").unwrap().version,
            "4.2653.1-dev"
        );
        assert_eq!(
            build(Some("4"), sunday + 1, Some(1), "p").unwrap().version,
            "4.2701.1-dev"
        );
        let missing = build(None, t, Some(5), "p").unwrap_err();
        assert!(missing.starts_with("HEADVER missing"), "{missing}");
        assert!(build(Some("x"), t, Some(5), "p").is_err());
    }

    /// SPEC-20261117-headver-ci-number: a CI build takes its number from `MOAI_BUILD_NUMBER`,
    /// says so on the `--version` note (so it is told apart from a local build of the same
    /// number), and a build given both numbers is refused rather than guessed.
    #[test]
    fn a_ci_build_names_its_number_and_two_sources_are_refused() {
        let t = 1_794_873_600 + 8 * 3600;
        let ci = build_numbered(Some("4\n"), t, None, Some(812), "2026.11.0-rc.1").unwrap();
        assert_eq!(ci.version, "4.2647.812-dev");
        assert_eq!(
            ci.note,
            "CI build 812 (MOAI_BUILD_NUMBER), built 2026-11-17 UTC (ISO week 2026-W47), \
             package 2026.11.0-rc.1"
        );
        let local = build_numbered(Some("4"), t, Some(5), None, "p").unwrap();
        assert_eq!(local, build(Some("4"), t, Some(5), "p").unwrap());
        let both = build_numbered(Some("4"), t, Some(5), Some(812), "p").unwrap_err();
        assert!(
            both.contains("MOAI_LOCAL_BUILD") && both.contains("MOAI_BUILD_NUMBER"),
            "{both}"
        );
    }

    /// SPEC-20261125-headver-weekly-release REQ-1: every build that is not a release ends in
    /// `-dev`, whichever source numbers it.
    #[test]
    fn a_development_build_ends_in_dev() {
        let t = 1_794_873_600 + 8 * 3600;
        for (local, ci, want) in [
            (Some(26), None, "4.2647.26-dev"),
            (None, Some(7), "4.2647.7-dev"),
            (None, None, "4.2647.0-dev"),
        ] {
            let b = build_numbered(Some("4"), t, local, ci, "p").unwrap();
            assert_eq!(b.version, want);
            assert!(b.version.ends_with("-dev"), "{}", b.version);
            assert!(!b.note.starts_with("release build"), "{}", b.note);
        }
        assert_eq!(dev_version(4, 2648, 26), "4.2648.26-dev");
    }

    /// REQ-2: a release build reports the given version with no suffix, and its note says
    /// `release build <serial> (MOAI_RELEASE_VERSION)`; the week is the release week, not the
    /// build date's week.
    #[test]
    fn a_release_build_reports_its_version_without_a_suffix() {
        let t = 1_794_873_600 + 8 * 3600; // 2026-11-17, ISO week 2026-W47
        let r = release_build(Some("4\n"), t, "4.2649.0", None, None, "2026.11.0-rc.1").unwrap();
        assert_eq!(r.version, "4.2649.0");
        assert_eq!(
            r.note,
            "release build 0 (MOAI_RELEASE_VERSION), built 2026-11-17 UTC (ISO week 2026-W47), \
             package 2026.11.0-rc.1"
        );
        let second = release_build(Some("4"), t, "4.2653.12", None, None, "p").unwrap();
        assert_eq!(second.version, "4.2653.12");
        assert!(second
            .note
            .starts_with("release build 12 (MOAI_RELEASE_VERSION), built "));
        let last = release_build(Some("4"), t, "4.2649.4294967295", None, None, "p").unwrap();
        assert_eq!(last.version, "4.2649.4294967295");
    }

    /// REQ-3: each way the release version can be wrong stops the build, and the reason names
    /// the variable. Boundaries are literals: N accepted above, N+1 refused here.
    #[test]
    fn a_wrong_release_version_stops_the_build() {
        let t = 1_794_873_600;
        for bad in [
            "5.2649.0",          // head differs from HEADVER (4)
            "04.2649.0",         // head with a leading zero
            "4.649.0",           // yearweek not four digits
            "4.26490.0",         // yearweek over four digits
            "4.2600.0",          // week 00
            "4.2654.0",          // week 54
            "4.2649.00",         // serial with a leading zero
            "4.2649.01",         // serial with a leading zero
            "4.2649.4294967296", // serial over u32
            "4.2649.-1",         // sign
            "4.2649.x",          // not a number
            "4.2649",            // two parts
            "4.2649.0.1",        // four parts
            "4.2649.0-dev",      // a release has no suffix
            "4.2649.0+meta",     // nor metadata
            "v4.2649.0",         // nor a prefix
            " 4.2649.0",         // nor a blank
            "",
        ] {
            let err = release_build(Some("4"), t, bad, None, None, "p").unwrap_err();
            assert!(err.starts_with("MOAI_RELEASE_VERSION="), "{bad:?}: {err}");
        }
        let missing = release_build(None, t, "4.2649.0", None, None, "p").unwrap_err();
        assert!(missing.starts_with("HEADVER missing"), "{missing}");
        let local = release_build(Some("4"), t, "4.2649.0", Some(5), None, "p").unwrap_err();
        assert!(
            local.contains("MOAI_RELEASE_VERSION") && local.contains("MOAI_LOCAL_BUILD"),
            "{local}"
        );
        let ci = release_build(Some("4"), t, "4.2649.0", None, Some(5), "p").unwrap_err();
        assert!(
            ci.contains("MOAI_RELEASE_VERSION") && ci.contains("MOAI_BUILD_NUMBER"),
            "{ci}"
        );
    }

    /// REQ-4: at the same three numbers the release is above its `-dev`; every other pair keeps
    /// the numeric order, so a development build of a later number or week is newer, and an
    /// installed folder named without `-dev` reads as a release.
    #[test]
    fn a_release_ranks_above_the_dev_build_of_the_same_numbers() {
        assert_eq!(compare("4.2649.0", "4.2649.0-dev"), Some(Ordering::Greater));
        assert_eq!(compare("4.2649.0-dev", "4.2649.0"), Some(Ordering::Less));
        assert_eq!(
            compare("4.2649.0-dev", "4.2649.0-dev"),
            Some(Ordering::Equal)
        );
        assert_eq!(compare("4.2649.0", "4.2649.0"), Some(Ordering::Equal));
        // Numbers decide first.
        assert_eq!(compare("4.2649.5-dev", "4.2649.0"), Some(Ordering::Greater));
        assert_eq!(
            compare("4.2649.0", "4.2648.99-dev"),
            Some(Ordering::Greater)
        );
        assert_eq!(
            compare("4.2648.27-dev", "4.2648.26"),
            Some(Ordering::Greater)
        );
        assert_eq!(compare("5.0001.0-dev", "4.2649.9"), Some(Ordering::Greater));
        // A folder written before the suffix existed has no `-dev`: it is a release shape.
        assert!(is_newer("4.2648.26", "4.2648.26-dev"));
        assert!(!is_newer("4.2648.26-dev", "4.2648.26"));
        // Only `-dev` is read; another suffix keeps naming the same build.
        assert_eq!(compare("4.2647.5-x", "4.2647.5"), Some(Ordering::Equal));
    }

    #[test]
    fn every_headver_version_is_newer_than_the_earlier_scheme() {
        for old in [
            "2026.11.0-rc.1",
            "2026.11.0-rc.2+local.4",
            "v2026.11.0",
            "2099.12.9",
        ] {
            for new in ["4.2647.5", "0.0001.0", "4.2647.0", "v4.2647.5+meta"] {
                assert_eq!(compare(new, old), Some(Ordering::Greater), "{new} > {old}");
                assert_eq!(compare(old, new), Some(Ordering::Less), "{old} < {new}");
                assert!(is_newer(new, old) && !is_newer(old, new));
            }
        }
    }

    #[test]
    fn headver_orders_head_then_yearweek_then_build() {
        let ordered = [
            "3.9953.99",
            "4.2101.1",
            "4.2647.0",
            "4.2647.5",
            "4.2647.6",
            "4.2653.1",
            "4.2701.0",
            "5.0001.0",
        ];
        for pair in ordered.windows(2) {
            assert_eq!(compare(pair[0], pair[1]), Some(Ordering::Less), "{pair:?}");
            assert_eq!(
                compare(pair[1], pair[0]),
                Some(Ordering::Greater),
                "{pair:?}"
            );
        }
        // Numbers, not text: build 10 is after build 9.
        assert_eq!(compare("4.2647.10", "4.2647.9"), Some(Ordering::Greater));
        // Metadata names the same build; a `-dev` suffix ranks below the release of the same numbers.
        assert_eq!(compare("4.2647.5+local", "4.2647.5"), Some(Ordering::Equal));
        assert_eq!(compare("4.2647.5-dev", "4.2647.5"), Some(Ordering::Less));
    }

    #[test]
    fn the_earlier_scheme_orders_releases_above_their_candidates() {
        let ordered = [
            "2026.10.0",
            "2026.11.0-rc.1",
            "2026.11.0-rc.2",
            "2026.11.0-rc.10",
            "2026.11.0",
            "2026.11.1-rc.1",
        ];
        for pair in ordered.windows(2) {
            assert_eq!(compare(pair[0], pair[1]), Some(Ordering::Less), "{pair:?}");
        }
        assert_eq!(
            compare("2026.11.0-rc.2+local.4", "2026.11.0-rc.2+local.3"),
            Some(Ordering::Equal)
        );
    }

    /// Audit B1: a head of 2000 or more is still HeadVer (the middle part is a four-digit
    /// yearweek, never a month), so `parse_head`'s whole range 0–9999 orders above every
    /// version of the earlier scheme.
    #[test]
    fn a_head_of_2000_or_more_is_still_headver() {
        for new in ["1999.2647.5", "2000.2647.5", "9999.2647.5", "2026.0011.0"] {
            for old in [
                "4.2647.5",
                "2026.11.0-rc.2+local.4",
                "2026.12.0",
                "2026.11.0",
            ] {
                assert_eq!(compare(new, old), Some(Ordering::Greater), "{new} > {old}");
                assert_eq!(compare(old, new), Some(Ordering::Less), "{old} < {new}");
            }
        }
        assert_eq!(
            compare("2000.2647.5", "1999.2647.5"),
            Some(Ordering::Greater)
        );
        // The month bound: 12 is the earlier scheme, 13 is no shape this build reads.
        assert_eq!(compare("2026.12.0", "2026.11.0"), Some(Ordering::Greater));
        assert_eq!(compare("2026.12.0", "4.2647.5"), Some(Ordering::Less));
        assert_eq!(compare("2026.13.0", "4.2647.5"), None);
        assert_eq!(compare("2026.1.0", "2026.0.0"), None);
    }

    #[test]
    fn a_shape_this_build_does_not_read_is_never_newer_or_older() {
        for odd in [
            "1999.11.0",
            "2026.13.0",
            "2026.0.1",
            "3.2.1",
            "4.26.5",
            "4.2654.1",
            "4.2600.1",
            "2026.11",
            "2026.11.0-beta.1",
            "x.y.z",
            "",
            "4.2647.-1",
        ] {
            assert_eq!(compare(odd, "4.2647.5"), None, "{odd}");
            assert_eq!(compare("4.2647.5", odd), None, "{odd}");
            assert!(!is_newer(odd, "4.2647.5") && !is_newer("4.2647.5", odd));
        }
    }
}
