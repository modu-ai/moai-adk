//! `lane-env --lane <name>` (#2841, leader decisions D1–D2): the build environment of one lane
//! as `export` lines a shell evaluates in one line, `e=$(moai lane-env --lane agent-7) && eval
//! "$e"`, so lanes stop prefixing every cargo command by hand. (A bare `eval "$(…)"` would go on
//! after a refusal: POSIX `eval` of an empty string returns 0.)
//!
//! - `CARGO_TARGET_DIR` is `<base>/moai-rc-target-<lane>`: one build folder per lane, kept when
//!   the lane moves to another worktree, so dependencies are reused. cargo decides what to
//!   rebuild from file mtimes, so after switching worktree, and after `git restore`,
//!   `git checkout --` or `git archive`, `touch` `build.rs` and the sources before the next build:
//!   otherwise the other worktree's build, or its build-script output (the `src/ops` command list),
//!   can stay (seen in a #2846 audit, 2026-11-14).
//! - `CARGO_HOME` is the lane's own `<base>/moai-rc-cargo-<lane>` when it holds a prepared
//!   registry (`registry/src`), else the shared `<base>/moai-rc-product-cargo` when that does,
//!   else it is left unset (cargo's own default) and a comment line says why. Nothing is created
//!   or copied: the command only reads.
//! - `CARGO_INCREMENTAL=0` always (#2930): incremental compilation is off for lane builds.
//! - `<base>` is `/tmp` on Unix (the folder lanes already use) and the system temporary folder
//!   on Windows, or `--base <absolute folder>`.
//!
//! Refused, with nothing printed on stdout: a lane name that is not one safe path component, a
//! relative base, and a target or cargo-home path that exists as a symbolic link or as anything
//! other than a folder (cargo would write through it).

use std::path::{Path, PathBuf};

/// The longest lane name accepted.
pub const MAX_LANE: usize = 64;

/// The shared, prepared cargo home lanes fall back to.
pub const SHARED_CARGO_HOME: &str = "moai-rc-product-cargo";

/// The default base folder: `/tmp` on Unix, the system temporary folder on Windows.
pub fn default_base() -> PathBuf {
    if cfg!(windows) {
        std::env::temp_dir()
    } else {
        PathBuf::from("/tmp")
    }
}

/// A lane name is one path component of `[A-Za-z0-9._-]`, 1 to [`MAX_LANE`] characters, not
/// led by `-` or `.` (so never `.`, `..` or an option).
pub fn check_lane(lane: &str) -> Result<(), String> {
    let ok = !lane.is_empty()
        && lane.len() <= MAX_LANE
        && !lane.starts_with(['-', '.'])
        && lane
            .chars()
            .all(|c| c.is_ascii_alphanumeric() || "-_.".contains(c));
    if ok {
        Ok(())
    } else {
        Err(format!(
            "lane-env: invalid lane name {lane:?}: use 1 to {MAX_LANE} letters, digits, `.`, `_` \
             or `-`, not starting with `-` or `.`"
        ))
    }
}

/// The lane's build environment.
#[derive(Debug, PartialEq, Eq)]
pub struct LaneEnv {
    pub target_dir: PathBuf,
    pub cargo_home: Option<PathBuf>,
    /// Why `cargo_home` is what it is, for the comment line.
    pub cargo_home_why: String,
}

/// What a path is, as far as this command cares.
enum Kind {
    Missing,
    Dir,
    Refused(&'static str),
}

fn kind(path: &Path) -> Kind {
    match std::fs::symlink_metadata(path) {
        Err(_) => Kind::Missing,
        Ok(m) if m.file_type().is_symlink() => Kind::Refused("a symbolic link"),
        Ok(m) if m.is_dir() => Kind::Dir,
        Ok(_) => Kind::Refused("not a folder"),
    }
}

fn refuse(path: &Path, what: &str) -> String {
    format!(
        "lane-env: {} is {what}; cargo would write through it. Remove or rename it yourself, \
         or pass another --base",
        path.display()
    )
}

/// A cargo home counts as prepared when it is a real folder holding `registry/src`.
fn prepared(home: &Path) -> Result<bool, String> {
    match kind(home) {
        Kind::Missing => Ok(false),
        Kind::Refused(what) => Err(refuse(home, what)),
        Kind::Dir => Ok(matches!(
            kind(&home.join("registry").join("src")),
            Kind::Dir
        )),
    }
}

/// The environment of `lane` under `base`. Reads only.
pub fn resolve(lane: &str, base: &Path) -> Result<LaneEnv, String> {
    check_lane(lane)?;
    if !base.is_absolute() {
        return Err(format!(
            "lane-env: --base {} is not an absolute path",
            base.display()
        ));
    }
    let target_dir = base.join(format!("moai-rc-target-{lane}"));
    if let Kind::Refused(what) = kind(&target_dir) {
        return Err(refuse(&target_dir, what));
    }
    let own = base.join(format!("moai-rc-cargo-{lane}"));
    let shared = base.join(SHARED_CARGO_HOME);
    let (cargo_home, cargo_home_why) = if prepared(&own)? {
        (Some(own), "the lane's own prepared cargo home".to_string())
    } else if prepared(&shared)? {
        (
            Some(shared),
            format!(
                "the shared prepared cargo home ({} has no registry/src)",
                own.display()
            ),
        )
    } else {
        (
            None,
            format!(
                "CARGO_HOME left unset: neither {} nor {} holds a prepared registry/src",
                own.display(),
                shared.display()
            ),
        )
    };
    Ok(LaneEnv {
        target_dir,
        cargo_home,
        cargo_home_why,
    })
}

/// A POSIX shell single-quoted word.
fn quote(path: &Path) -> String {
    format!("'{}'", path.to_string_lossy().replace('\'', "'\\''"))
}

/// The lines to evaluate: comments start with `#`, every other line is one `export`.
pub fn render(lane: &str, env: &LaneEnv) -> String {
    let mut out = format!("# moai lane-env --lane {lane} (#2841)\n");
    out.push_str(&format!(
        "export CARGO_TARGET_DIR={}\n",
        quote(&env.target_dir)
    ));
    // #2930: incremental output only grows the shared build folder and cannot be cached
    // (Cargo: `CARGO_INCREMENTAL` "when set to 0 it will force disabling it",
    // doc.rust-lang.org/cargo/reference/environment-variables.html; sccache README:
    // "Incrementally compiled crates cannot be cached").
    out.push_str("export CARGO_INCREMENTAL=0\n");
    match &env.cargo_home {
        Some(home) => {
            out.push_str(&format!("# CARGO_HOME: {}\n", env.cargo_home_why));
            out.push_str(&format!("export CARGO_HOME={}\n", quote(home)));
        }
        None => out.push_str(&format!("# {}\n", env.cargo_home_why)),
    }
    out.push_str(
        "# One build folder for every worktree of this lane: after switching worktree or after \
         git restore/checkout/archive, touch build.rs and the sources before building.\n",
    );
    out
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::fs;

    fn base() -> (tempfile::TempDir, PathBuf) {
        let d = tempfile::TempDir::new().unwrap();
        let p = fs::canonicalize(d.path()).unwrap();
        (d, p)
    }

    #[test]
    fn lane_names() {
        for good in ["agent-7", "a", "agent_10.b", &"x".repeat(MAX_LANE)] {
            assert!(check_lane(good).is_ok(), "{good}");
        }
        let long = "x".repeat(MAX_LANE + 1);
        for bad in [
            "", ".", "..", "-x", ".x", "a/b", "a\\b", "a b", "../x", "한글", "a\n", &long,
        ] {
            assert!(check_lane(bad).is_err(), "{bad:?}");
        }
    }

    #[test]
    fn target_and_cargo_home_paths() {
        let (_d, b) = base();
        let e = resolve("agent-7", &b).unwrap();
        assert_eq!(e.target_dir, b.join("moai-rc-target-agent-7"));
        assert_eq!(e.cargo_home, None);
        assert!(e.cargo_home_why.contains("left unset"));

        // A shared prepared home is used when the lane's own is missing or unprepared.
        fs::create_dir_all(b.join("moai-rc-product-cargo/registry/src")).unwrap();
        fs::create_dir_all(b.join("moai-rc-cargo-agent-7/registry")).unwrap();
        let e = resolve("agent-7", &b).unwrap();
        assert_eq!(e.cargo_home, Some(b.join("moai-rc-product-cargo")));

        // The lane's own prepared home wins.
        fs::create_dir_all(b.join("moai-rc-cargo-agent-7/registry/src")).unwrap();
        let e = resolve("agent-7", &b).unwrap();
        assert_eq!(e.cargo_home, Some(b.join("moai-rc-cargo-agent-7")));
        // Another lane is not given this lane's home.
        let other = resolve("agent-8", &b).unwrap();
        assert_eq!(other.cargo_home, Some(b.join("moai-rc-product-cargo")));
    }

    #[test]
    fn relative_base_is_refused() {
        assert!(resolve("agent-7", Path::new("tmp")).is_err());
    }

    #[cfg(unix)]
    #[test]
    fn symbolic_links_and_files_are_refused() {
        let (_d, b) = base();
        let elsewhere = b.join("elsewhere");
        fs::create_dir(&elsewhere).unwrap();
        std::os::unix::fs::symlink(&elsewhere, b.join("moai-rc-target-agent-7")).unwrap();
        let err = resolve("agent-7", &b).unwrap_err();
        assert!(err.contains("is a symbolic link"), "{err}");

        fs::write(b.join("moai-rc-target-agent-8"), "x").unwrap();
        let err = resolve("agent-8", &b).unwrap_err();
        assert!(err.contains("is not a folder"), "{err}");

        // A symbolic-link cargo home is refused too, own or shared, prepared or not.
        fs::create_dir_all(elsewhere.join("registry/src")).unwrap();
        std::os::unix::fs::symlink(&elsewhere, b.join("moai-rc-cargo-agent-9")).unwrap();
        assert!(resolve("agent-9", &b)
            .unwrap_err()
            .contains("is a symbolic link"));
        std::os::unix::fs::symlink(&elsewhere, b.join("moai-rc-product-cargo")).unwrap();
        assert!(resolve("agent-10", &b)
            .unwrap_err()
            .contains("is a symbolic link"));
        // Nothing was created by any of these calls.
        let mut names: Vec<_> = fs::read_dir(&b)
            .unwrap()
            .map(|e| e.unwrap().file_name().into_string().unwrap())
            .collect();
        names.sort();
        assert_eq!(
            names,
            [
                "elsewhere",
                "moai-rc-cargo-agent-9",
                "moai-rc-product-cargo",
                "moai-rc-target-agent-7",
                "moai-rc-target-agent-8"
            ]
        );
    }

    #[test]
    fn render_quotes_and_comments() {
        let e = LaneEnv {
            target_dir: PathBuf::from("/tmp/it's here/moai-rc-target-a"),
            cargo_home: None,
            cargo_home_why: "CARGO_HOME left unset: x".into(),
        };
        let text = render("a", &e);
        assert!(text.contains("export CARGO_TARGET_DIR='/tmp/it'\\''s here/moai-rc-target-a'\n"));
        assert!(!text.contains("export CARGO_HOME"));
        assert_eq!(text.matches("export CARGO_INCREMENTAL=0\n").count(), 1);
        for line in text.lines() {
            assert!(
                line.starts_with('#') || line.starts_with("export "),
                "{line}"
            );
        }
    }
}
