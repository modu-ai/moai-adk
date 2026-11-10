//! rc.2 E6 (#2198) RELEASE test suite: Go/RC feed isolation and interrupted-update
//! recovery. Design §17 RELEASE ("Go/RC paths and feed isolation, attestation and interrupted-update
//! recovery"); builds on
//! the 04-A characterization in `tests/install_cli.rs` without editing it.
//!
//! Every case is a fixture: fake release feeds, a model of documented GitHub semantics, and
//! a temporary HOME for the recovery sweep. No network call, no real release or tag, no
//! real HOME.
//!
//! Sources, read 2026-11-09 (sha256 of the fetched bodies):
//! - Go updater at `origin/main` 4fa2aa9db2e0add0f1299e35407141b2b41c5ec0:
//!   `internal/update/checker.go` (last changed d6b80a01c) and `internal/cli/deps.go`
//!   (last changed 0bd9d7fa7). Line numbers below are at that revision.
//! - GitHub REST "Releases" (`/en/rest/releases/releases`, 731dc3d1ac31…): "The latest
//!   release is the most recent non-prerelease, non-draft release"; `make_latest`
//!   "Defaults to true for newly published releases".
//! - GitHub workflow syntax "Filter pattern cheat sheet"
//!   (`/en/actions/reference/workflows-and-actions/workflow-syntax`, 3b8841b5372b…).
//!
//! The Go functions below are a test-side port (a model), duplicated from the 04-A port in
//! `tests/install_cli.rs` and extended with the stable `/releases/latest` path and the
//! `checksums.txt` gate. Expected cells are written as literals, never computed by the port.

#![cfg(unix)]

use moai_cli::tree_digest;
use serde_json::{json, Value};
use std::{
    fs,
    io::Write as _,
    path::{Path, PathBuf},
    process::{Command, Stdio},
};
use tempfile::TempDir;

// --- Go 3.x updater model --------------------------------------------------------------

#[derive(Clone, Debug)]
struct Release {
    tag: String,
    draft: bool,
    prerelease: bool,
    assets: Vec<String>,
}

/// A repository's releases, newest first (GitHub list order), and the release GitHub
/// reports at `/releases/latest`.
#[derive(Clone, Debug)]
struct Repo {
    list: Vec<Release>,
    latest: Option<String>,
}

impl Repo {
    /// The recorded feed (`tests/fixtures/go-releases-sample.json`, see `install_cli.rs`).
    /// Its first non-draft, non-prerelease entry is the latest release.
    fn recorded() -> Repo {
        let sample: Vec<Value> =
            serde_json::from_str(include_str!("fixtures/go-releases-sample.json")).unwrap();
        let list: Vec<Release> = sample
            .iter()
            .map(|r| Release {
                tag: r["tag_name"].as_str().unwrap().to_string(),
                draft: r["draft"] == json!(true),
                prerelease: r["prerelease"] == json!(true),
                assets: r["assets"]
                    .as_array()
                    .unwrap()
                    .iter()
                    .map(|a| a["name"].as_str().unwrap().to_string())
                    .collect(),
            })
            .collect();
        let latest = list
            .iter()
            .find(|r| !r.draft && !r.prerelease)
            .map(|r| r.tag.clone());
        Repo { list, latest }
    }

    /// Publishes a release on top of the list. GitHub REST: drafts and prereleases cannot be
    /// latest; `make_latest` defaults to true for a newly published release.
    fn publish(&mut self, r: Release, make_latest: bool) {
        if !r.draft && !r.prerelease && make_latest {
            self.latest = Some(r.tag.clone());
        }
        self.list.insert(0, r);
    }
}

/// Assets a cargo-dist 0.34.0 release carries (docs/release-attestation.md, dry run): an
/// archive, its `.sha256`, and `sha256.sum`. No `checksums.txt`.
fn dist_assets() -> Vec<String> {
    [
        "moai-cli-aarch64-apple-darwin.tar.xz",
        "moai-cli-aarch64-apple-darwin.tar.xz.sha256",
        "sha256.sum",
    ]
    .iter()
    .map(|s| s.to_string())
    .collect()
}

/// Assets a GoReleaser release carries (the recorded feed): `checksums.txt` plus archives.
fn goreleaser_assets(version: &str) -> Vec<String> {
    vec![
        "checksums.txt".to_string(),
        format!("moai-adk_{version}_darwin_arm64.tar.gz"),
    ]
}

fn release(tag: &str, prerelease: bool, assets: Vec<String>) -> Release {
    Release {
        tag: tag.to_string(),
        draft: false,
        prerelease,
        assets,
    }
}

/// deps.go L433-447: the current version's shape picks the endpoint. `rc`, `alpha`,
/// `beta`, `dev` or a `go-v` prefix select the `/releases` list; anything else
/// `/releases/latest`.
fn go_uses_list(current: &str) -> bool {
    current == "dev"
        || current.contains("rc")
        || current.contains("alpha")
        || current.contains("beta")
        || current.starts_with("go-v")
}

/// checker.go L343-382: strip `go-v` then `v`; at most 3 dot segments; each segment cut at
/// the first `-`/`+`; a non-numeric segment counts as 0.
fn go_semver(v: &str) -> [i64; 3] {
    let v = v.strip_prefix("go-v").unwrap_or(v);
    let v = v.strip_prefix('v').unwrap_or(v);
    let mut parts = [0i64; 3];
    for (i, seg) in v.splitn(3, '.').enumerate() {
        let seg = match seg.find(['-', '+']) {
            Some(idx) => &seg[..idx],
            None => seg,
        };
        if let Ok(n) = seg.parse::<i64>() {
            parts[i] = n;
        }
    }
    parts
}

#[derive(Debug, PartialEq)]
enum Go {
    /// `IsUpdateAvailable` returns true for this tag.
    Update(String),
    /// Latest is the same or older (checker.go L331-335).
    NoUpdate,
    /// The check returns an error; the client cannot see any update.
    Error(&'static str),
}

/// What a Go 3.x client at `current` reports. `authenticated` mirrors a token in GH_TOKEN /
/// GITHUB_TOKEN (checker.go L72-74): GitHub lists drafts only to such callers.
fn go_check(repo: &Repo, current: &str, authenticated: bool) -> Go {
    let selected = if go_uses_list(current) {
        // checker.go L26-30: the response struct has no draft/prerelease fields.
        // L99-115: keep `go-v`/`v` prefixed tags, take the first.
        repo.list
            .iter()
            .filter(|r| authenticated || !r.draft)
            .find(|r| r.tag.starts_with("go-v") || r.tag.starts_with('v'))
            .cloned()
            .ok_or("no_v_prefixed_release")
    } else {
        // checker.go L118-124: the single latest object, no prefix filter. A repository
        // without a latest release answers 404 (L82-84).
        repo.latest
            .as_ref()
            .and_then(|t| repo.list.iter().find(|r| &r.tag == t))
            .cloned()
            .ok_or("not_found_404")
    };
    let rel = match selected {
        Ok(r) => r,
        Err(e) => return Go::Error(e),
    };
    // checker.go L203-216: a release with no `checksums.txt` asset is refused with
    // ErrChecksumUnavailable before any version comparison.
    if !rel.assets.iter().any(|a| a == "checksums.txt") {
        return Go::Error("checksum_unavailable");
    }
    if go_semver(&rel.tag) > go_semver(current) {
        Go::Update(rel.tag)
    } else {
        Go::NoUpdate
    }
}

fn up(tag: &str) -> Go {
    Go::Update(tag.to_string())
}

/// Go clients used by the table: a stable 3.x install at the recorded latest, and an RC
/// install on the same line (its version contains `rc`, so it reads the list).
const STABLE: &str = "v3.1.2";
const RC: &str = "v3.1.2-rc.1";

/// REQ-E6-01 / AC-1: the feed contamination table, extended. Each row publishes one
/// candidate on top of the recorded feed and pins what a Go stable client and an anonymous
/// Go RC client report. The plain `v*` rows reproduce the RC-channel filter risk.
#[test]
fn go_feed_contamination_table_reproduces_plain_v_risk() {
    let base = Repo::recorded();
    assert_eq!(base.latest.as_deref(), Some("v3.1.2"));
    assert_eq!(go_check(&base, STABLE, false), Go::NoUpdate);
    assert_eq!(go_check(&base, RC, false), Go::NoUpdate);

    // (tag, prerelease, assets, make_latest, expected stable, expected rc)
    type Row = (&'static str, bool, Vec<String>, bool, Go, Go);
    let rows: Vec<Row> = vec![
        // Decided namespace, prerelease: invisible to both channels.
        (
            "moai-cli-v2026.11.0-rc.1",
            true,
            dist_assets(),
            true,
            Go::NoUpdate,
            Go::NoUpdate,
        ),
        // Plain `v` prerelease: the RC list selects it and the check errors (no checksums.txt).
        (
            "v2026.11.0-rc.1",
            true,
            dist_assets(),
            true,
            Go::NoUpdate,
            Go::Error("checksum_unavailable"),
        ),
        // Same tag carrying a `checksums.txt`: the RC client is offered a "2026.11.0" update.
        (
            "v2026.11.0-rc.1",
            true,
            goreleaser_assets("2026.11.0-rc.1"),
            true,
            Go::NoUpdate,
            up("v2026.11.0-rc.1"),
        ),
        // Plain `v` final: becomes latest and is first in the list: both channels break.
        (
            "v2026.11.0",
            false,
            dist_assets(),
            true,
            Go::Error("checksum_unavailable"),
            Go::Error("checksum_unavailable"),
        ),
        // Namespaced final with the dist default (make_latest true): stable breaks.
        (
            "moai-cli-v2026.11.0",
            false,
            dist_assets(),
            true,
            Go::Error("checksum_unavailable"),
            Go::NoUpdate,
        ),
        // Namespaced final published with make_latest=false: both channels stay on v3.1.2.
        (
            "moai-cli-v2026.11.0",
            false,
            dist_assets(),
            false,
            Go::NoUpdate,
            Go::NoUpdate,
        ),
        // Control: a historical `go-v` Go RC with checksums reaches the RC channel only.
        (
            "go-v3.1.3-rc.1",
            true,
            goreleaser_assets("3.1.3-rc.1"),
            true,
            Go::NoUpdate,
            up("go-v3.1.3-rc.1"),
        ),
    ];

    println!("E6 Go 3.x feed contamination table (recorded feed + one candidate)");
    println!("stable client {STABLE} (/releases/latest), rc client {RC} (/releases, anonymous)");
    for (tag, pre, assets, make_latest, want_stable, want_rc) in rows {
        let mut repo = base.clone();
        let checksums = assets.iter().any(|a| a == "checksums.txt");
        repo.publish(release(tag, pre, assets), make_latest);
        let stable = go_check(&repo, STABLE, false);
        let rc = go_check(&repo, RC, false);
        println!(
            "| {tag} | prerelease={pre} | checksums.txt={checksums} | make_latest={make_latest} \
             | stable={stable:?} | rc={rc:?} |"
        );
        assert_eq!(stable, want_stable, "stable client, candidate {tag}");
        assert_eq!(rc, want_rc, "rc client, candidate {tag}");
    }

    // The recorded feed's first entry is a draft (v3.1.3). The response struct has no
    // draft field, so an authenticated RC client is offered the draft; anonymous is not.
    assert!(base.list[0].draft);
    assert_eq!(go_check(&base, RC, true), up("v3.1.3"));
    assert_eq!(go_check(&base, RC, false), Go::NoUpdate);

    // The prerelease suffix does not protect a `v` tag in the comparison (L372-375).
    assert_eq!(go_semver("v2026.11.0-rc.1"), [2026, 11, 0]);
    assert_eq!(go_semver("moai-cli-v2026.11.0"), [0, 11, 0]);
}

/// docs/release-attestation.md approval item 6 (#2196 F1): a namespaced Rust final release
/// published with the dist default becomes `/releases/latest` and hides a newer Go release
/// from Go stable clients. With `make_latest=false` the Go release stays visible.
#[test]
fn rust_final_release_hides_newer_go_release_from_stable_clients() {
    let doc = include_str!("../docs/release-attestation.md");
    for needle in ["sha256.sum", "--latest=false", "/releases/latest"] {
        assert!(
            doc.contains(needle),
            "release-attestation.md lost {needle:?}"
        );
    }

    let mut repo = Repo::recorded();
    repo.publish(release("v3.2.0", false, goreleaser_assets("3.2.0")), true);
    assert_eq!(go_check(&repo, STABLE, false), up("v3.2.0"));

    // Dist default: the Rust final takes latest; the stable check now errors.
    let mut dist_default = repo.clone();
    dist_default.publish(release("moai-cli-v2026.11.0", false, dist_assets()), true);
    assert_eq!(dist_default.latest.as_deref(), Some("moai-cli-v2026.11.0"));
    assert_eq!(
        go_check(&dist_default, STABLE, false),
        Go::Error("checksum_unavailable")
    );
    // Even with a `checksums.txt`, the tag parses as 0.11.0 and the newer v3.2.0 is hidden.
    let mut with_checksums = repo.clone();
    with_checksums.publish(
        release("moai-cli-v2026.11.0", false, goreleaser_assets("2026.11.0")),
        true,
    );
    assert_eq!(go_check(&with_checksums, STABLE, false), Go::NoUpdate);

    // make_latest=false: latest stays v3.2.0 and stable clients still see it.
    let mut isolated = repo.clone();
    isolated.publish(release("moai-cli-v2026.11.0", false, dist_assets()), false);
    assert_eq!(isolated.latest.as_deref(), Some("v3.2.0"));
    assert_eq!(go_check(&isolated, STABLE, false), up("v3.2.0"));
    // The RC list never selected the namespaced tag in any of the three.
    for r in [&dist_default, &with_checksums, &isolated] {
        assert_eq!(go_check(r, "go-v3.2.0-rc.1", false), Go::NoUpdate);
    }
}

// --- release workflow trigger --------------------------------------------------------

/// A GitHub Actions branch/tag filter pattern as an anchored regex, per the cheat sheet:
/// `**` any characters, `*` any characters except `/`, `?` zero or one of the preceding
/// character, `+` one or more of the preceding character, `[...]` one character from the
/// listed ranges. Other characters, `.` included, are literal; a pattern names the whole
/// ref ("Matches the exact name", `v2*` "start with v2").
fn filter_regex(pattern: &str) -> String {
    let chars: Vec<char> = pattern.chars().collect();
    let mut out = String::from("^");
    let mut i = 0;
    while i < chars.len() {
        match chars[i] {
            '*' if chars.get(i + 1) == Some(&'*') => {
                out.push_str(".*");
                i += 1;
            }
            '*' => out.push_str("[^/]*"),
            '?' | '+' => out.push(chars[i]),
            '[' => {
                let end = i + chars[i..].iter().position(|c| *c == ']').unwrap();
                out.extend(&chars[i..=end]);
                i = end;
            }
            c if c.is_ascii_alphanumeric() || c == '-' || c == '_' || c == '/' => out.push(c),
            c => {
                out.push('\\');
                out.push(c);
            }
        }
        i += 1;
    }
    out.push('$');
    out
}

/// Whole-string match of the regex subset `filter_regex` emits: `^`/`$` anchors,
/// `.*`, `[^/]*`, `[...]` classes, `\x` escapes, literals, each optionally followed by
/// `+` or `?`. Backtracking, small inputs only.
fn regex_match(re: &str, s: &str) -> bool {
    #[derive(Clone)]
    enum Atom {
        Any,
        NotSlash,
        Class(Vec<(char, char)>),
        Lit(char),
    }
    #[derive(Clone, Copy, PartialEq)]
    enum Rep {
        One,
        Star,
        Plus,
        Opt,
    }
    let re: Vec<char> = re.chars().collect();
    assert!(re.first() == Some(&'^') && re.last() == Some(&'$'));
    let body = &re[1..re.len() - 1];
    let mut atoms: Vec<(Atom, Rep)> = Vec::new();
    let mut i = 0;
    while i < body.len() {
        let (atom, mut rep, next) = if body[i..].starts_with(&['.', '*']) {
            (Atom::Any, Rep::Star, i + 2)
        } else if body[i..].starts_with(&['[', '^', '/', ']', '*']) {
            (Atom::NotSlash, Rep::Star, i + 5)
        } else if body[i] == '[' {
            let end = i + body[i..].iter().position(|c| *c == ']').unwrap();
            let inner = &body[i + 1..end];
            let mut ranges = Vec::new();
            let mut j = 0;
            while j < inner.len() {
                if j + 2 < inner.len() && inner[j + 1] == '-' {
                    ranges.push((inner[j], inner[j + 2]));
                    j += 3;
                } else {
                    ranges.push((inner[j], inner[j]));
                    j += 1;
                }
            }
            (Atom::Class(ranges), Rep::One, end + 1)
        } else if body[i] == '\\' {
            (Atom::Lit(body[i + 1]), Rep::One, i + 2)
        } else {
            (Atom::Lit(body[i]), Rep::One, i + 1)
        };
        i = next;
        if rep == Rep::One && i < body.len() && (body[i] == '+' || body[i] == '?') {
            rep = if body[i] == '+' { Rep::Plus } else { Rep::Opt };
            i += 1;
        }
        atoms.push((atom, rep));
    }
    fn hit(a: &Atom, c: char) -> bool {
        match a {
            Atom::Any => true,
            Atom::NotSlash => c != '/',
            Atom::Class(r) => r.iter().any(|(lo, hi)| *lo <= c && c <= *hi),
            Atom::Lit(l) => *l == c,
        }
    }
    fn go(atoms: &[(Atom, Rep)], s: &[char]) -> bool {
        let Some(((atom, rep), rest)) = atoms.split_first() else {
            return s.is_empty();
        };
        let (min, max) = match rep {
            Rep::One => (1, 1),
            Rep::Opt => (0, 1),
            Rep::Plus => (1, s.len()),
            Rep::Star => (0, s.len()),
        };
        let mut n = 0;
        while n < max && n < s.len() && hit(atom, s[n]) {
            n += 1;
        }
        (min..=n).rev().any(|k| go(rest, &s[k..]))
    }
    let s: Vec<char> = s.chars().collect();
    go(&atoms, &s)
}

fn tag_filter_matches(pattern: &str, tag: &str) -> bool {
    regex_match(&filter_regex(pattern), tag)
}

/// docs/release-attestation.md: the cargo-dist 0.34.0 default release trigger also fires on
/// Go release tags; `tag-namespace = "moai-cli"` limits it to `moai-cli*` tags. Evaluated
/// with a model of the documented filter semantics, not a workflow run.
#[test]
fn dist_default_tag_trigger_fires_on_go_tags_and_namespace_isolates() {
    const DEFAULT: &str = "**[0-9]+.[0-9]+.[0-9]+*";
    const NAMESPACED: &str = "moai-cli**[0-9]+.[0-9]+.[0-9]+*";
    let doc = include_str!("../docs/release-attestation.md");
    assert!(
        doc.contains(&format!("'{DEFAULT}'")),
        "default trigger literal"
    );
    assert!(
        doc.contains(&format!("'{NAMESPACED}'")),
        "namespaced trigger literal"
    );

    // The cheat sheet's own examples, so the model is held to the documented rows.
    assert!(tag_filter_matches("v[12].[0-9]+.[0-9]+", "v1.10.1"));
    assert!(tag_filter_matches("v[12].[0-9]+.[0-9]+", "v2.0.0"));
    assert!(!tag_filter_matches("v[12].[0-9]+.[0-9]+", "v3.0.0"));
    assert!(tag_filter_matches("v2*", "v2.9"));
    assert!(!tag_filter_matches("v2*", "xv2"));
    assert!(tag_filter_matches("feature/**", "feature/mona/the/octocat"));
    assert!(!tag_filter_matches("feature/*", "feature/beta-a/my-branch"));
    assert!(tag_filter_matches("*feature", "mona-feature"));
    assert!(!tag_filter_matches("*feature", "feature-x"));

    // (tag, default fires, namespaced fires)
    let rows = [
        ("v3.1.3", true, false),
        ("go-v3.2.1-rc.2", true, false),
        ("v2026.11.0-rc.1", true, false),
        ("moai-cli-v2026.11.0-rc.1", true, true),
        ("moai-cli-v2026.11.0", true, true),
        ("x-moai-cli-v2026.11.0", true, false),
        ("moai-cli-nightly", false, false),
    ];
    println!("E6 release workflow trigger (model of the documented filter semantics)");
    for (tag, want_default, want_ns) in rows {
        let d = tag_filter_matches(DEFAULT, tag);
        let n = tag_filter_matches(NAMESPACED, tag);
        println!("| {tag} | default={d} | namespaced={n} |");
        assert_eq!(d, want_default, "default pattern on {tag}");
        assert_eq!(n, want_ns, "namespaced pattern on {tag}");
    }
}

// --- interrupted update recovery -------------------------------------------------------

const EXE: &str = env!("CARGO_BIN_EXE_moai-cli");
const SKILL_V1: &str = "---\nname: moai\ndescription: fixture package\n---\n# MoAI\nbody v1\n";
const SKILL_V2: &str = "---\nname: moai\ndescription: fixture package\n---\n# MoAI\nbody v2\n";
const UPDATE_V1: &str = "# UPDATE\nv1\n";
const UPDATE_V2: &str = "# UPDATE\nv2\n";
const NEW_V2: &str = "# NEW\nadded in v2\n";

fn write_pkg(dir: &Path, files: &[(&str, &str)]) {
    for (rel, body) in files {
        let p = dir.join(rel);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        fs::write(p, body).unwrap();
    }
}

fn run(args: &[&str], stdin: Option<&Value>) -> (bool, Value) {
    let mut child = Command::new(EXE)
        .args(args)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    let mut pipe = child.stdin.take().unwrap();
    if let Some(v) = stdin {
        match pipe.write_all(v.to_string().as_bytes()) {
            Ok(()) => {}
            Err(e) if e.kind() == std::io::ErrorKind::BrokenPipe => {}
            Err(e) => panic!("stdin write failed: {e}"),
        }
    }
    drop(pipe);
    let out = child.wait_with_output().unwrap();
    let v: Value = serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|_| panic!("no envelope: {}", String::from_utf8_lossy(&out.stderr)));
    (out.status.success(), v)
}

fn plan_and_apply(home: &Path, stage: &Path) {
    let (h, s) = (home.to_str().unwrap(), stage.to_str().unwrap());
    let (ok, v) = run(&["update-plan", h, s], None);
    assert!(ok, "{v}");
    let plan = v["result"].clone();
    let digest = plan["plan_digest"].as_str().unwrap().to_string();
    let args = [
        "update-apply",
        h,
        s,
        "--plan-digest",
        &digest,
        "--permission-reference",
        "fixture-approved",
    ];
    let (ok, v) = run(&args, Some(&plan));
    assert!(ok, "{v}");
    assert_eq!(v["result"]["status"], json!("applied"), "{v}");
}

fn digest(dir: &Path) -> String {
    tree_digest(dir, &[]).unwrap()
}

fn hosts_dir(home: &Path) -> PathBuf {
    let inst = home.join(".moai-cli/installations");
    let dirs: Vec<PathBuf> = fs::read_dir(&inst)
        .unwrap()
        .map(|e| e.unwrap().path())
        .filter(|p| p.is_dir())
        .collect();
    assert_eq!(dirs.len(), 1);
    dirs[0].join("hosts/moai")
}

fn sha_hex(bytes: &[u8]) -> String {
    let s = moai_cli::digest(bytes);
    s.strip_prefix("sha256:").unwrap().to_string()
}

fn host<'a>(v: &'a Value, name: &str) -> &'a Value {
    v["hosts"]
        .as_array()
        .unwrap()
        .iter()
        .find(|h| h["host"] == json!(name))
        .unwrap()
}

/// One crash point of a real v1→v2 update: the codex apply stopped after the first `k`
/// receipt entries were published, before the receipt commit. Recovery must return the
/// whole codex projection to its v1 bytes from the known-good copies the real v2 apply
/// stored, leave the committed claude host and the fake Go state untouched, and end
/// `recovered`.
struct Sweep {
    k: usize,
    paths: Vec<String>,
    classes: Vec<String>,
    recovered: Vec<String>,
    v1_restored: bool,
}

fn crash_and_recover(k: usize) -> Sweep {
    let root = TempDir::new().unwrap();
    let home = root.path().join("home");
    fs::create_dir_all(home.join(".moai/cache")).unwrap();
    fs::write(home.join(".moai/config.yaml"), "generation: go\n").unwrap();
    let stage1 = root.path().join("stage-v1");
    let stage2 = root.path().join("stage-v2");
    write_pkg(
        &stage1,
        &[("SKILL.md", SKILL_V1), ("workflows/update.md", UPDATE_V1)],
    );
    write_pkg(
        &stage2,
        &[
            ("SKILL.md", SKILL_V2),
            ("workflows/update.md", UPDATE_V2),
            ("workflows/new.md", NEW_V2),
        ],
    );
    let codex = home.join(".agents/skills/moai");
    let claude = home.join(".claude/skills/moai");
    let go_before = digest(&home.join(".moai"));

    plan_and_apply(&home, &stage1);
    let codex_v1 = digest(&codex);
    plan_and_apply(&home, &stage2);
    let claude_v2 = digest(&claude);
    let hosts = hosts_dir(&home);

    // The real v2 apply stored the overwritten v1 bytes as known-good copies.
    for body in [SKILL_V1, UPDATE_V1] {
        let copy = hosts
            .join("codex/known-good")
            .join(sha_hex(body.as_bytes()));
        assert_eq!(fs::read_to_string(&copy).unwrap(), body, "{copy:?}");
    }

    // Rewind the committed v2 codex state to the crash point: entries at index >= k never
    // published (bytes back to v1, a created file absent); the receipt is pending. The
    // entry at k-1 keeps `published: false`, as a crash between its rename and the
    // receipt save leaves it.
    let receipt_path = hosts.join("codex.json");
    let mut receipt: Value = serde_json::from_slice(&fs::read(&receipt_path).unwrap()).unwrap();
    assert_eq!(receipt["status"], json!("committed"));
    let v1_bytes = |path: &str| match path {
        "SKILL.md" => Some(SKILL_V1),
        "workflows/update.md" => Some(UPDATE_V1),
        _ => None,
    };
    let mut paths = Vec::new();
    for (i, e) in receipt["entries"]
        .as_array_mut()
        .unwrap()
        .iter_mut()
        .enumerate()
    {
        let path = e["path"].as_str().unwrap().to_string();
        if i >= k {
            match v1_bytes(&path) {
                Some(body) => fs::write(codex.join(&path), body).unwrap(),
                None => fs::remove_file(codex.join(&path)).unwrap(),
            }
        }
        e["published"] = json!(i + 1 < k);
        paths.push(path);
    }
    receipt["status"] = json!("pending");
    receipt["finished_at"] = Value::Null;
    fs::write(&receipt_path, serde_json::to_vec_pretty(&receipt).unwrap()).unwrap();

    let h = home.to_str().unwrap();
    let (ok, v) = run(&["recovery-plan", h], None);
    assert!(ok, "{v}");
    let plan = v["result"].clone();
    let classes: Vec<String> = paths
        .iter()
        .map(|p| {
            host(&plan, "codex")["files"]
                .as_array()
                .unwrap()
                .iter()
                .find(|f| f["path"] == json!(p))
                .unwrap()["class"]
                .as_str()
                .unwrap()
                .to_string()
        })
        .collect();
    let digest_arg = plan["plan_digest"].as_str().unwrap().to_string();
    let (ok, v) = run(
        &[
            "recovery-apply",
            h,
            "--plan-digest",
            &digest_arg,
            "--permission-reference",
            "fixture-approved",
        ],
        Some(&plan),
    );
    assert!(ok, "k={k}: {v}");
    let result = host(&v["result"], "codex");
    assert_eq!(result["receipt_status"], json!("recovered"), "k={k}: {v}");
    assert_eq!(result["conflicts"], json!([]), "k={k}: {v}");
    let recovered: Vec<String> = result["recovered"]
        .as_array()
        .unwrap()
        .iter()
        .map(|p| p.as_str().unwrap().to_string())
        .collect();
    assert_eq!(digest(&claude), claude_v2, "k={k}: committed host moved");
    assert_eq!(host(&v["result"], "claude")["written"], json!(false));
    assert_eq!(
        digest(&home.join(".moai")),
        go_before,
        "k={k}: Go state moved"
    );
    Sweep {
        k,
        paths,
        classes,
        recovered,
        v1_restored: digest(&codex) == codex_v1,
    }
}

/// AC-1: interrupted-update recovery at every crash point of a real three-file update,
/// k = 0 (nothing published) through 3 (all published, receipt never committed).
#[test]
fn interrupted_update_recovers_known_good_at_every_crash_point() {
    println!("E6 interrupted update recovery sweep (codex host, real known-good copies)");
    for k in 0..=3 {
        let s = crash_and_recover(k);
        // The receipt's entry order is the crash order; pinned so the sweep cannot follow
        // a reordering silently.
        assert_eq!(
            s.paths,
            ["SKILL.md", "workflows/new.md", "workflows/update.md"],
            "receipt entry order"
        );
        println!(
            "| k={} | entries={:?} | classes={:?} | recovered={:?} | codex==v1: {} |",
            s.k, s.paths, s.classes, s.recovered, s.v1_restored
        );
        let want_classes: Vec<&str> = (0..3)
            .map(|i| {
                if i < k {
                    "matches_after"
                } else {
                    "matches_before"
                }
            })
            .collect();
        assert_eq!(s.classes, want_classes, "k={k}");
        let all = ["SKILL.md", "workflows/new.md", "workflows/update.md"];
        assert_eq!(s.recovered, all[..k].to_vec(), "k={k}");
        assert!(s.v1_restored, "k={k}: codex projection not back at v1");
    }
}
