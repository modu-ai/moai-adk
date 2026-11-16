//! Versioning schemes of a project (card C-69): SemVer, HeadVer and CalVer.
//!
//! Three parts. [`classify`] tells the shape of one version string, a pure function.
//! [`fact`] is what `project-detect` reports as `versioning`: the scheme the newest tags (or,
//! without tags, the root manifests) show, with the evidence it was read from; it writes
//! nothing and reaches no network. [`VersioningSettings`] is the `versioning` key of the
//! project profile, the answer of the `/moai project` interview, and [`render_line`] is the one
//! line the `AGENTS.md` managed section carries for it.
//!
//! The rules the skill follows are in `skills/moai/references/internal/versioning.md`.

use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use std::{
    fs,
    path::Path,
    process::{Command, Stdio},
};

type Result<T> = std::result::Result<T, String>;

/// The shape of one version string.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Shape {
    Semver,
    Calver,
    Headver,
    /// Fits two schemes; it is reported and never counted.
    Ambiguous,
    Unrecognised,
}

impl Shape {
    pub fn name(self) -> &'static str {
        match self {
            Shape::Semver => "semver",
            Shape::Calver => "calver",
            Shape::Headver => "headver",
            Shape::Ambiguous => "ambiguous",
            Shape::Unrecognised => "unrecognised",
        }
    }
}

/// Newest tags read as evidence.
pub const MAX_TAGS: usize = 5;
/// Longest tag name or manifest version kept as evidence.
const MAX_VALUE: usize = 64;
/// Longest `custom` text in the profile.
const MAX_FORMAT: usize = 128;
/// Largest manifest read for its version.
const MAX_FILE: u64 = 256 * 1024;
/// Most output taken from one git call.
const GIT_OUTPUT_LIMIT: usize = 64 * 1024;

const NOTE_SEMVER_OR_CALVER: &str =
    "valid SemVer and a possible CalVer (two-digit year, month); not counted";
const NOTE_HEADVER_OR_CALVER: &str =
    "a HeadVer year-week segment behind a four-digit year; not counted";

/// A two-digit year is taken as one from 10 (2010): below that a first segment is far more
/// often a plain number.
const FIRST_SHORT_YEAR: u32 = 10;
const FIRST_FULL_YEAR: u32 = 2000;

/// The shape of one version string and, for an ambiguous one, why. A leading `v` and a prefix
/// that ends in `/` or `@` (`release/v1.2.3`, `pkg@1.2.3`) are dropped, and what follows the
/// first `-` or `+` (pre-release, build metadata) is not read.
fn shape_of(text: &str) -> (Shape, Option<&'static str>) {
    let name = text.rsplit(['/', '@']).next().unwrap_or(text);
    let name = name.strip_prefix(['v', 'V']).unwrap_or(name);
    let core = name.split(['-', '+']).next().unwrap_or(name);
    let segments: Vec<&str> = core.split('.').collect();
    let numeric = |s: &&str| !s.is_empty() && s.len() <= 9 && s.bytes().all(|b| b.is_ascii_digit());
    if !(2..=4).contains(&segments.len()) || !segments.iter().all(numeric) {
        return (Shape::Unrecognised, None);
    }
    let number = |s: &str| s.parse::<u32>().unwrap_or(u32::MAX);
    // SemVer 2.0.0 rule 2: no leading zeros.
    let plain = |s: &&str| *s == "0" || !s.starts_with('0');
    let semver = segments.len() == 3 && segments.iter().all(plain);
    let full_year = segments[0].len() == 4 && number(segments[0]) >= FIRST_FULL_YEAR;
    let short_year = segments[0].len() == 2
        && number(segments[0]) >= FIRST_SHORT_YEAR
        && segments[1].len() <= 2
        && (1..=12).contains(&number(segments[1]));
    let headver = segments.len() == 3
        && plain(&segments[0])
        && plain(&segments[2])
        && segments[1].len() == 4
        && number(&segments[1][..2]) >= FIRST_SHORT_YEAR
        && (1..=53).contains(&number(&segments[1][2..]));
    match (headver, full_year, short_year, semver) {
        (true, true, _, _) => (Shape::Ambiguous, Some(NOTE_HEADVER_OR_CALVER)),
        (true, ..) => (Shape::Headver, None),
        (_, true, ..) => (Shape::Calver, None),
        (_, _, true, true) => (Shape::Ambiguous, Some(NOTE_SEMVER_OR_CALVER)),
        (_, _, true, false) => (Shape::Calver, None),
        (_, _, _, true) => (Shape::Semver, None),
        _ => (Shape::Unrecognised, None),
    }
}

/// The shape of one version string: a tag name or a manifest version.
pub fn classify(text: &str) -> Shape {
    shape_of(text).0
}

/// The scheme a set of shapes shows: the one scheme, `mixed` for two or more, `None` when no
/// shape is one of the three.
fn scheme_of(shapes: impl Iterator<Item = Shape>) -> Option<&'static str> {
    let mut found: Vec<Shape> = Vec::new();
    for shape in shapes {
        let definite = matches!(shape, Shape::Semver | Shape::Calver | Shape::Headver);
        if definite && !found.contains(&shape) {
            found.push(shape);
        }
    }
    match found.as_slice() {
        [] => None,
        [one] => Some(one.name()),
        _ => Some("mixed"),
    }
}

/// The versioning fact from tag names (newest first) and manifest versions (`(file, version)`).
/// The tags decide when one of them has a shape; the manifests decide only without that.
/// `latest` is the newest tag that is a version, else the first manifest version that is one.
pub fn summarize(tags: &[String], manifests: &[(&str, String)]) -> Value {
    let row = |source: &str, value: &str| {
        let (shape, note) = shape_of(value);
        let mut row = json!({"source": source, "value": value, "shape": shape.name()});
        if let Some(note) = note {
            row["note"] = json!(note);
        }
        row
    };
    let evidence: Vec<Value> = tags
        .iter()
        .map(|t| row("tag", t))
        .chain(manifests.iter().map(|(file, v)| row(file, v)))
        .collect();
    let scheme = scheme_of(tags.iter().map(|t| classify(t)))
        .or_else(|| scheme_of(manifests.iter().map(|(_, v)| classify(v))))
        .unwrap_or("none");
    let latest = tags
        .iter()
        .chain(manifests.iter().map(|(_, v)| v))
        .find(|v| classify(v) != Shape::Unrecognised);
    json!({"scheme": scheme, "latest": latest, "evidence": evidence})
}

/// Text short and plain enough to be shown as evidence: a tag name or a manifest version.
fn showable(value: &str) -> bool {
    !value.is_empty()
        && value.len() <= MAX_VALUE
        && value.bytes().all(|b| {
            b.is_ascii_alphanumeric() || matches!(b, b'.' | b'-' | b'_' | b'+' | b'/' | b'@')
        })
}

/// The regular file `root/<name>` as text; `None` for a link, a folder, a large or missing file.
fn root_file(root: &Path, name: &str) -> Option<String> {
    let path = root.join(name);
    let meta = fs::symlink_metadata(&path).ok()?;
    if !meta.file_type().is_file() || meta.len() > MAX_FILE {
        return None;
    }
    fs::read_to_string(path).ok()
}

/// The version each root manifest states, in a fixed order: `Cargo.toml` (`[package]`, else
/// `[workspace.package]`), `package.json`, `pyproject.toml` (`[project]`, else
/// `[tool.poetry]`). A value that is not plain version text is left out.
fn manifest_versions(root: &Path) -> Vec<(&'static str, String)> {
    let toml =
        |name: &str| root_file(root, name).and_then(|t| t.parse::<toml_edit::DocumentMut>().ok());
    let text = |item: Option<&toml_edit::Item>| {
        item.and_then(|i| i.get("version"))
            .and_then(toml_edit::Item::as_str)
            .map(String::from)
    };
    let cargo = toml("Cargo.toml").and_then(|doc| {
        text(doc.get("package"))
            .or_else(|| text(doc.get("workspace").and_then(|w| w.get("package"))))
    });
    let package_json = root_file(root, "package.json")
        .and_then(|t| serde_json::from_str::<Value>(&t).ok())
        .and_then(|v| v["version"].as_str().map(String::from));
    let pyproject = toml("pyproject.toml").and_then(|doc| {
        text(doc.get("project")).or_else(|| text(doc.get("tool").and_then(|t| t.get("poetry"))))
    });
    [
        ("Cargo.toml", cargo),
        ("package.json", package_json),
        ("pyproject.toml", pyproject),
    ]
    .into_iter()
    .filter_map(|(file, version)| Some((file, version.filter(|v| showable(v))?)))
    .collect()
}

/// The newest tag names of the repository whose `.git` is in `root`, newest first, at most
/// [`MAX_TAGS`], and how the read went: `read`, `no_repository` (no `.git` entry of its own;
/// a parent folder's repository is not the project's) or `unavailable` (git missing or
/// failing). One fixed `git for-each-ref` call on `root/.git` itself, so git discovers
/// nothing, with the shared git environment scrub; it takes no lock and writes nothing.
fn newest_tags(root: &Path) -> (Vec<String>, &'static str) {
    let git_dir = root.join(".git");
    match fs::symlink_metadata(&git_dir) {
        Ok(meta) if !meta.file_type().is_symlink() => {}
        _ => return (Vec::new(), "no_repository"),
    }
    let mut cmd = Command::new("git");
    cmd.arg("--no-optional-locks")
        .arg("--git-dir")
        .arg(&git_dir)
        .args([
            "for-each-ref",
            "--sort=-version:refname",
            "--sort=-creatordate",
        ])
        .arg(format!("--count={MAX_TAGS}"))
        .args(["--format=%(refname:short)", "refs/tags"]);
    crate::identity::scrub_git_env(&mut cmd);
    let out = cmd
        .env("GIT_OPTIONAL_LOCKS", "0")
        .env("GIT_TERMINAL_PROMPT", "0")
        .stdin(Stdio::null())
        .stderr(Stdio::null())
        .output();
    let Ok(out) = out else {
        return (Vec::new(), "unavailable");
    };
    if !out.status.success() || out.stdout.len() > GIT_OUTPUT_LIMIT {
        return (Vec::new(), "unavailable");
    }
    let tags = String::from_utf8_lossy(&out.stdout)
        .lines()
        .filter(|t| showable(t))
        .take(MAX_TAGS)
        .map(String::from)
        .collect();
    (tags, "read")
}

/// The `versioning` fact of `project-detect`: [`summarize`] over the project's newest tags and
/// root manifest versions, plus `tags`, how the tag read went. Reads only.
pub fn fact(root: &Path) -> Value {
    let (tags, status) = newest_tags(root);
    let mut fact = summarize(&tags, &manifest_versions(root));
    fact["tags"] = json!(status);
    fact
}

/// The schemes the profile accepts. `custom` is the user's own words, kept in `format`.
pub const SCHEMES: [&str; 4] = ["semver", "headver", "calver", "custom"];
/// The tokens of a CalVer format (calver.org, spec 26.0); `MM`, `WW` and `DD` are the
/// spelling the page used before that spec.
const CALVER_TOKENS: [&str; 16] = [
    "YYYY", "YY", "0Y", "M", "0M", "W", "0W", "D", "0D", "MM", "WW", "DD", "MAJOR", "MINOR",
    "MICRO", "MODIFIER",
];

/// The `versioning` key of the project profile: the scheme the project's releases follow, as
/// picked in the `/moai project` interview.
#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Eq)]
#[serde(deny_unknown_fields)]
pub struct VersioningSettings {
    /// One of [`SCHEMES`].
    pub scheme: String,
    /// The CalVer format (`YYYY.0M.MICRO`), or for `custom` the user's words in one line.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub format: Option<String>,
    /// The HeadVer head number.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub head: Option<u64>,
}

/// Two to four known tokens joined by dots, a year among them.
fn calver_format(format: &str) -> bool {
    let tokens: Vec<&str> = format.split('.').collect();
    (2..=4).contains(&tokens.len())
        && tokens.iter().all(|t| CALVER_TOKENS.contains(t))
        && tokens.iter().any(|t| ["YYYY", "YY", "0Y"].contains(t))
}

/// Accepts a `versioning` value: a known scheme, a `format` only on CalVer (a token format) or
/// `custom` (one bounded line), a `head` only on HeadVer. No value is echoed.
pub fn validate(v: &VersioningSettings) -> Result<()> {
    let scheme = v.scheme.as_str();
    let format_ok = match (scheme, v.format.as_deref()) {
        (_, None) => true,
        ("calver", Some(format)) => calver_format(format),
        ("custom", Some(text)) => {
            !text.trim().is_empty()
                && text.len() <= MAX_FORMAT
                && !text.chars().any(char::is_control)
        }
        _ => false,
    };
    if !SCHEMES.contains(&scheme) || !format_ok || (v.head.is_some() && scheme != "headver") {
        return Err("invalid versioning setting; no values were logged".into());
    }
    Ok(())
}

/// The line the `AGENTS.md` managed section carries for the choice, in the section's
/// language (`ko`, else English), newline included.
pub fn render_line(v: &VersioningSettings, lang: &str) -> String {
    let label = if lang == "ko" {
        "버전 규칙"
    } else {
        "Versioning"
    };
    let scheme = match (v.scheme.as_str(), v.format.as_deref(), v.head) {
        ("semver", ..) => "SemVer (MAJOR.MINOR.PATCH)".to_string(),
        ("headver", _, Some(head)) => {
            format!("HeadVer ({{head}}.{{yearweek}}.{{build}}), head {head}")
        }
        ("headver", ..) => "HeadVer ({head}.{yearweek}.{build})".to_string(),
        ("calver", Some(format), _) => format!("CalVer ({format})"),
        ("calver", ..) => "CalVer".to_string(),
        (_, Some(text), _) => format!("custom ({text})"),
        _ => "custom".to_string(),
    };
    format!("{label}: {scheme}.\n")
}

#[cfg(test)]
mod tests {
    use super::*;

    fn shapes(list: &[&str]) -> Vec<&'static str> {
        list.iter().map(|t| classify(t).name()).collect()
    }

    #[test]
    fn the_three_schemes_are_told_apart() {
        for semver in [
            "v3.2.1",
            "3.2.1",
            "0.1.0",
            "1.0.0-rc.1",
            "1.0.0+build.7",
            "pkg@1.4.0",
            "release/v2.0.1",
            "100.4.1",
        ] {
            assert_eq!(classify(semver), Shape::Semver, "{semver}");
        }
        for calver in [
            "2026.11.0",
            "v2026.11.0-rc.1",
            "2026.03.17",
            "2026.3",
            "26.04",
            "26.4",
            "26.04.1",
            "2026.11.17.1",
        ] {
            assert_eq!(classify(calver), Shape::Calver, "{calver}");
        }
        for headver in ["3.2647.59", "v0.2134.1", "12.2601.4+meta"] {
            assert_eq!(classify(headver), Shape::Headver, "{headver}");
        }
        for other in [
            "nightly",
            "1.2",
            "1.2.3.4",
            "01.2.3",
            "1..2",
            "",
            "v",
            "1.2.x",
            "1.2.3.4.5",
            "2026",
        ] {
            assert_eq!(classify(other), Shape::Unrecognised, "{other:?}");
        }
    }

    /// A shape that fits two schemes is named ambiguous, never guessed.
    #[test]
    fn a_shape_that_fits_two_schemes_is_ambiguous() {
        // Valid SemVer and a possible CalVer `YY.M.MICRO`.
        assert_eq!(shapes(&["26.4.1", "22.12.0", "10.1.0"]), ["ambiguous"; 3]);
        // A HeadVer middle segment behind a four-digit year.
        assert_eq!(classify("2026.2647.59"), Shape::Ambiguous);
    }

    /// Each limit at its edge: the last value inside and the first outside.
    #[test]
    fn the_limits_hold_at_their_edges() {
        // A four-digit year starts at 2000.
        assert_eq!(shapes(&["2000.1", "1999.1"]), ["calver", "unrecognised"]);
        assert_eq!(shapes(&["2000.1.0", "1999.1.0"]), ["calver", "semver"]);
        // A two-digit year needs a month from 1 to 12 behind it.
        assert_eq!(
            shapes(&["26.12", "26.13", "26.0", "26.00"]),
            ["calver", "unrecognised", "unrecognised", "unrecognised"]
        );
        assert_eq!(shapes(&["26.12.1", "26.13.1"]), ["ambiguous", "semver"]);
        // A two-digit year starts at 10.
        assert_eq!(shapes(&["10.04", "09.04"]), ["calver", "unrecognised"]);
        // The HeadVer week runs from 01 to 53, behind a two-digit year from 10.
        assert_eq!(
            shapes(&["3.2601.5", "3.2653.5", "3.2654.5", "3.2600.5"]),
            ["headver", "headver", "semver", "semver"]
        );
        assert_eq!(
            shapes(&["3.1001.5", "3.0901.5"]),
            ["headver", "unrecognised"]
        );
        // The middle segment has exactly four digits.
        assert_eq!(shapes(&["3.264.5", "3.26471.5"]), ["semver", "semver"]);
    }

    fn tags(list: &[&str]) -> Vec<String> {
        list.iter().map(|t| t.to_string()).collect()
    }

    #[test]
    fn tags_of_one_scheme_name_it_and_the_newest_is_latest() {
        let v = summarize(&tags(&["v3.2.1", "v3.2.0", "nightly"]), &[]);
        assert_eq!(v["scheme"], "semver");
        assert_eq!(v["latest"], "v3.2.1");
        assert_eq!(
            v["evidence"],
            json!([
                {"source": "tag", "value": "v3.2.1", "shape": "semver"},
                {"source": "tag", "value": "v3.2.0", "shape": "semver"},
                {"source": "tag", "value": "nightly", "shape": "unrecognised"},
            ])
        );
        assert_eq!(summarize(&tags(&["2026.11.0"]), &[])["scheme"], "calver");
        assert_eq!(summarize(&tags(&["3.2647.59"]), &[])["scheme"], "headver");
    }

    #[test]
    fn tags_of_two_schemes_are_mixed_and_nothing_is_none() {
        let v = summarize(&tags(&["2026.11.0", "v3.2.1"]), &[]);
        assert_eq!(v["scheme"], "mixed");
        assert_eq!(v["latest"], "2026.11.0");
        let v = summarize(&[], &[]);
        assert_eq!(v, json!({"scheme": "none", "latest": null, "evidence": []}));
        // Tags that are no versions are evidence and decide nothing.
        let v = summarize(&tags(&["nightly", "latest"]), &[]);
        assert_eq!(v["scheme"], "none");
        assert_eq!(v["latest"], Value::Null);
        assert_eq!(v["evidence"].as_array().unwrap().len(), 2);
    }

    /// An ambiguous shape says so in its evidence row and is not counted.
    #[test]
    fn an_ambiguous_version_is_reported_and_not_counted() {
        let v = summarize(&tags(&["26.4.1"]), &[]);
        assert_eq!(v["scheme"], "none");
        assert_eq!(v["latest"], "26.4.1");
        let row = &v["evidence"][0];
        assert_eq!(row["shape"], "ambiguous");
        assert!(
            row["note"].as_str().unwrap().contains("SemVer")
                && row["note"].as_str().unwrap().contains("CalVer"),
            "{row}"
        );
        // Beside a definite shape the definite one decides.
        let v = summarize(&tags(&["26.4.1", "26.04.0"]), &[]);
        assert_eq!(v["scheme"], "calver");
        // Rows of a definite shape carry no note.
        assert!(v["evidence"][1].get("note").is_none());
    }

    /// Tags decide when one of them has a shape; the manifests decide only without that.
    #[test]
    fn manifests_decide_only_without_a_recognised_tag() {
        let manifests = [("Cargo.toml", "0.3.1".to_string())];
        let v = summarize(&[], &manifests);
        assert_eq!(v["scheme"], "semver");
        assert_eq!(v["latest"], "0.3.1");
        assert_eq!(
            v["evidence"],
            json!([{"source": "Cargo.toml", "value": "0.3.1", "shape": "semver"}])
        );
        let v = summarize(&tags(&["2026.11.0"]), &manifests);
        assert_eq!(v["scheme"], "calver");
        assert_eq!(v["latest"], "2026.11.0");
        assert_eq!(v["evidence"].as_array().unwrap().len(), 2);
        let two = [
            ("Cargo.toml", "0.3.1".to_string()),
            ("package.json", "2026.11.0".to_string()),
        ];
        assert_eq!(summarize(&tags(&["nightly"]), &two)["scheme"], "mixed");
    }

    fn settings(scheme: &str, format: Option<&str>, head: Option<u64>) -> VersioningSettings {
        VersioningSettings {
            scheme: scheme.into(),
            format: format.map(String::from),
            head,
        }
    }

    #[test]
    fn settings_accept_the_four_schemes_and_refuse_the_rest() {
        for ok in [
            settings("semver", None, None),
            settings("headver", None, None),
            settings("headver", None, Some(0)),
            settings("headver", None, Some(3)),
            settings("calver", None, None),
            settings("calver", Some("YYYY.0M.MICRO"), None),
            settings("calver", Some("YY.0M.MICRO"), None),
            settings("calver", Some("YYYY.0M.0D"), None),
            settings("calver", Some("YY.MINOR.MICRO"), None),
            settings("calver", Some("YYYY.0M.0D.MICRO"), None),
            // The spelling calver.org used before spec 26.0 is still read.
            settings("calver", Some("YYYY.MM.DD"), None),
            settings("custom", None, None),
            settings(
                "custom",
                Some("release train letter plus build number"),
                None,
            ),
        ] {
            assert!(validate(&ok).is_ok(), "{ok:?}");
        }
        for bad in [
            settings("", None, None),
            settings("SemVer", None, None),
            settings("romantic", None, None),
            // A format belongs to CalVer or custom, a head to HeadVer.
            settings("semver", Some("MAJOR.MINOR.PATCH"), None),
            settings("semver", None, Some(1)),
            settings("calver", None, Some(1)),
            settings("headver", Some("YYYY.0M.MICRO"), None),
            // A CalVer format is two to four known tokens with a year among them.
            settings("calver", Some("YYYY"), None),
            settings("calver", Some("MAJOR.MINOR.MICRO"), None),
            settings("calver", Some("YYYY.0M.0D.MICRO.MODIFIER"), None),
            settings("calver", Some("YYYY.Q.MICRO"), None),
            settings("calver", Some("yyyy.0m"), None),
            settings("calver", Some(""), None),
            // Custom text is one bounded line.
            settings("custom", Some(""), None),
            settings("custom", Some("a\nb"), None),
            settings("custom", Some(&"x".repeat(129)), None),
        ] {
            assert!(validate(&bad).is_err(), "{bad:?}");
        }
        assert!(validate(&settings("custom", Some(&"x".repeat(128)), None)).is_ok());
    }

    #[test]
    fn the_managed_line_names_the_scheme() {
        assert_eq!(
            render_line(&settings("semver", None, None), "en"),
            "Versioning: SemVer (MAJOR.MINOR.PATCH).\n"
        );
        assert_eq!(
            render_line(&settings("headver", None, Some(3)), "en"),
            "Versioning: HeadVer ({head}.{yearweek}.{build}), head 3.\n"
        );
        assert_eq!(
            render_line(&settings("headver", None, None), "en"),
            "Versioning: HeadVer ({head}.{yearweek}.{build}).\n"
        );
        assert_eq!(
            render_line(&settings("calver", Some("YYYY.0M.MICRO"), None), "en"),
            "Versioning: CalVer (YYYY.0M.MICRO).\n"
        );
        assert_eq!(
            render_line(&settings("calver", None, None), "en"),
            "Versioning: CalVer.\n"
        );
        assert_eq!(
            render_line(&settings("custom", Some("train plus build"), None), "en"),
            "Versioning: custom (train plus build).\n"
        );
        assert_eq!(
            render_line(&settings("semver", None, None), "ko"),
            "버전 규칙: SemVer (MAJOR.MINOR.PATCH).\n"
        );
    }
}
