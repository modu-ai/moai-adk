//! `lane-env --lane <name>` (#2841, leader decisions D1–D2): the build environment of one lane
//! as `export` lines a shell evaluates in one line, `e=$(moai lane-env --lane agent-7) && eval
//! "$e"`, so lanes stop prefixing every cargo command by hand. (A bare `eval "$(…)"` would go on
//! after a refusal: POSIX `eval` of an empty string returns 0.)
//!
//! - `CARGO_TARGET_DIR` is the build folder of the worktree the command runs in (card C-82,
//!   `skills/moai/references/internal/build-folder.md`; follow-up asked by the user 2026-11-17
//!   08:22Z): `<HOME>/.moai/build/<project>-<worktree name>`, one per worktree, never shared,
//!   named as `tools/build_folder.py env` names it ([`folder_name`]; a test runs both). Like that
//!   tool it makes `~/.moai`, `~/.moai/build` and the folder (mode 0700) when missing and writes
//!   the owner record `.moai-build.json`, so `moai optimize-plan` can offer the folder for deletion
//!   once the worktree is gone. It replaces the per-lane `<base>/moai-rc-target-<lane>` (#2841).
//!   cargo decides what to rebuild from file mtimes, so after `git restore`, `git checkout --` or
//!   `git archive`, `touch` `build.rs` and the sources before the next build (#2846 audit).
//! - `CARGO_HOME` is the lane's own `<base>/moai-rc-cargo-<lane>` when it holds a prepared
//!   registry (`registry/src`), else the shared `<base>/moai-rc-product-cargo` when that does,
//!   else it is left unset (cargo's own default) and a comment line says why. Nothing is created
//!   or copied there. "Prepared" means `registry/src` exists, not that the
//!   registry holds every crate `Cargo.lock` names: on 2026-11-15 the shared home was prepared
//!   and still failed offline with `no matching package named cap-std` (#3054; `AGENTS.local.md`
//!   says how to fill a lane's own home without the network).
//! - `CARGO_INCREMENTAL=0` always (#2930): incremental compilation is off for lane builds.
//! - `<base>` is `/tmp` on Unix (the folder lanes already use) and the system temporary folder
//!   on Windows, or `--base <absolute folder>`.
//!
//! Refused, with nothing printed on stdout: a lane name that is not one safe path component, a
//! relative base, a cargo-home path that exists as a symbolic link or as anything other than a
//! folder (cargo would write through it), a HOME that is not absolute, a folder outside a git
//! worktree, a link or a non-folder on the build folder's path, and a build folder whose owner
//! record names another worktree. The cargo homes are checked before anything is written.

use serde_json::Value;
use std::ffi::OsStr;
use std::path::{Path, PathBuf};
use std::process::Command;

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

/// The environment of `lane` for the worktree that holds `cwd`: the cargo home under `base`
/// (read only), then the worktree's build folder under `home` (made and recorded).
pub fn resolve(
    lane: &str,
    base: &Path,
    home: Option<&OsStr>,
    cwd: &Path,
) -> Result<LaneEnv, String> {
    let (cargo_home, cargo_home_why) = cargo_home(lane, base)?;
    let home = home_dir(home)?;
    let target_dir = prepare(&home, &owner(cwd)?)?;
    Ok(LaneEnv {
        target_dir,
        cargo_home,
        cargo_home_why,
    })
}

/// The cargo home of `lane` under `base`, and why. Reads only.
pub fn cargo_home(lane: &str, base: &Path) -> Result<(Option<PathBuf>, String), String> {
    check_lane(lane)?;
    if !base.is_absolute() {
        return Err(format!(
            "lane-env: --base {} is not an absolute path",
            base.display()
        ));
    }
    let own = base.join(format!("moai-rc-cargo-{lane}"));
    let shared = base.join(SHARED_CARGO_HOME);
    Ok(if prepared(&own)? {
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
    })
}

/// The owner record in a build folder (C-82; `tools/build_folder.py` `RECORD`).
pub const RECORD: &str = crate::optimize::AGENT_BUILD_RECORD;

/// The largest owner record read (`tools/build_folder.py` `RECORD_MAX`).
const RECORD_MAX: u64 = 4096;

/// Who a build folder belongs to: the repository's git common dir, the worktree root (both real
/// paths) and the folder's name under `~/.moai/build/`.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Owner {
    pub common: PathBuf,
    pub worktree: PathBuf,
    pub name: String,
}

/// `tools/build_folder.py` `safe`: each character outside `[A-Za-z0-9._-]` becomes `_`; an
/// empty result, `.` and `..` refuse.
fn safe(name: &str) -> Result<String, String> {
    let cleaned: String = name
        .chars()
        .map(|c| {
            if c.is_ascii_alphanumeric() || "._-".contains(c) {
                c
            } else {
                '_'
            }
        })
        .collect();
    if matches!(cleaned.as_str(), "" | "." | "..") {
        return Err(format!("lane-env: no usable folder name from {name:?}"));
    }
    Ok(cleaned)
}

fn base_name(path: &Path) -> String {
    path.file_name()
        .map(|n| n.to_string_lossy().into_owned())
        .unwrap_or_default()
}

/// The build folder name of `worktree` in the repository whose git common dir is `common` (both
/// real paths), as `tools/build_folder.py` `owner` names it: `<project>` is the primary
/// checkout's folder name (a trailing `.git` dropped); the primary checkout, and a worktree whose
/// name is `<project>` or starts with `<project>-`, use the name alone, any other worktree
/// `<project>-<name>`.
pub fn folder_name(common: &Path, worktree: &Path) -> Result<String, String> {
    let primary = if common.file_name() == Some(OsStr::new(".git")) {
        common.parent().unwrap_or(common)
    } else {
        common
    };
    let project = base_name(primary);
    let project = safe(project.strip_suffix(".git").unwrap_or(&project))?;
    let name = safe(&base_name(worktree))?;
    Ok(if worktree == primary {
        project
    } else if name != project && !name.starts_with(&format!("{project}-")) {
        format!("{project}-{name}")
    } else {
        name
    })
}

/// One git call in `cwd` (environment scrubbed, #2043); its trimmed stdout.
fn git(cwd: &Path, args: &[&str]) -> Result<String, String> {
    let mut cmd = Command::new("git");
    cmd.args(args).current_dir(cwd);
    crate::identity::scrub_git_env(&mut cmd);
    let out = cmd
        .output()
        .map_err(|e| format!("lane-env: git could not start: {e}"))?;
    if !out.status.success() {
        return Err(format!(
            "lane-env: git {} failed: {}",
            args.join(" "),
            String::from_utf8_lossy(&out.stderr).trim()
        ));
    }
    Ok(String::from_utf8_lossy(&out.stdout).trim().to_string())
}

fn real(path: &Path) -> Result<PathBuf, String> {
    std::fs::canonicalize(path).map_err(|e| format!("lane-env: {}: {e}", path.display()))
}

/// The owner of the worktree that holds `cwd`.
pub fn owner(cwd: &Path) -> Result<Owner, String> {
    let worktree = real(Path::new(&git(cwd, &["rev-parse", "--show-toplevel"])?))?;
    let common = PathBuf::from(git(cwd, &["rev-parse", "--git-common-dir"])?);
    let common = real(&cwd.join(common))?;
    let name = folder_name(&common, &worktree)?;
    Ok(Owner {
        common,
        worktree,
        name,
    })
}

/// The real paths of every worktree `git worktree list` names from `cwd` (ones that are gone
/// are left out).
pub fn worktrees(cwd: &Path) -> Result<Vec<PathBuf>, String> {
    Ok(git(cwd, &["worktree", "list", "--porcelain"])?
        .lines()
        .filter_map(|l| l.strip_prefix("worktree "))
        .filter_map(|p| std::fs::canonicalize(p).ok())
        .collect())
}

/// HOME as given; it must be absolute.
pub fn home_dir(home: Option<&OsStr>) -> Result<PathBuf, String> {
    match home.map(PathBuf::from) {
        Some(h) if h.is_absolute() => Ok(h),
        h => Err(format!(
            "lane-env: HOME must be an absolute folder, got {:?}",
            h.unwrap_or_default()
        )),
    }
}

/// `<home>/.moai/build/<name>`, without looking at the disk.
pub fn build_folder(home: &Path, name: &str) -> PathBuf {
    home.join(".moai").join("build").join(name)
}

/// `path` is a real folder, made with mode 0700 when missing; a link or anything else refuses
/// (C-82 audit B1: a link is never followed).
fn real_dir(path: &Path) -> Result<(), String> {
    if std::fs::symlink_metadata(path).is_err() {
        let mut b = std::fs::DirBuilder::new();
        #[cfg(unix)]
        std::os::unix::fs::DirBuilderExt::mode(&mut b, 0o700);
        match b.create(path) {
            Ok(()) => {}
            Err(e) if e.kind() == std::io::ErrorKind::AlreadyExists => {}
            Err(e) => {
                return Err(format!(
                    "lane-env: {} is not writable: {e}",
                    path.parent().unwrap_or(path).display()
                ))
            }
        }
    }
    match std::fs::symlink_metadata(path) {
        Ok(m) if m.file_type().is_symlink() => Err(format!(
            "lane-env: {} is a link; it is not followed",
            path.display()
        )),
        Ok(m) if m.is_dir() => Ok(()),
        _ => Err(format!("lane-env: {} is not a folder", path.display())),
    }
}

/// The owner record of `folder` as JSON, `None` when there is none or it is not an object; a
/// link, a non-file, an oversized or unreadable record refuses.
pub fn read_record(folder: &Path) -> Result<Option<Value>, String> {
    let rec = folder.join(RECORD);
    let Ok(meta) = std::fs::symlink_metadata(&rec) else {
        return Ok(None);
    };
    if !meta.is_file() || meta.len() > RECORD_MAX {
        return Err(format!(
            "lane-env: {} is not a regular file of at most {RECORD_MAX} bytes",
            rec.display()
        ));
    }
    let bytes = std::fs::read(&rec).map_err(|e| format!("lane-env: {}: {e}", rec.display()))?;
    let v: Value = serde_json::from_slice(&bytes)
        .map_err(|e| format!("lane-env: {} is not a record: {e}", rec.display()))?;
    Ok(v.is_object().then_some(v))
}

/// `tools/build_folder.py env`: makes the build folder of `owner` under `home` and writes its
/// owner record unless an equal one is there; refuses a record of another worktree.
pub fn prepare(home: &Path, owner: &Owner) -> Result<PathBuf, String> {
    let mut path = home.to_path_buf();
    for part in [".moai", "build", owner.name.as_str()] {
        path.push(part);
        real_dir(&path)?;
    }
    let record = serde_json::json!({
        "schema_version": 1,
        "git_common_dir": owner.common.to_string_lossy(),
        "worktree": owner.worktree.to_string_lossy(),
    });
    let old = read_record(&path)?;
    if let Some(w) = old.as_ref().map(|o| &o["worktree"]) {
        if !w.is_null() && *w != record["worktree"] {
            return Err(format!(
                "lane-env: {} belongs to the worktree {}",
                path.display(),
                w.as_str().map_or_else(|| w.to_string(), str::to_string)
            ));
        }
    }
    if old.as_ref() != Some(&record) {
        let tmp = path.join(format!("{RECORD}.tmp"));
        let mut o = std::fs::OpenOptions::new();
        o.write(true).create_new(true);
        #[cfg(unix)]
        std::os::unix::fs::OpenOptionsExt::mode(&mut o, 0o600);
        let f = o
            .open(&tmp)
            .map_err(|e| format!("lane-env: {} is in the way: {e}", tmp.display()))?;
        serde_json::to_writer(f, &record)
            .map_err(|e| format!("lane-env: {}: {e}", tmp.display()))?;
        std::fs::rename(&tmp, path.join(RECORD))
            .map_err(|e| format!("lane-env: {}: {e}", tmp.display()))?;
    }
    Ok(path)
}

/// A POSIX shell single-quoted word (also the `claim-link` next-command paths, #3030).
pub(crate) fn quote(path: &Path) -> String {
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
        "# This worktree's own build folder (internal/build-folder, card C-82): after \
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
    fn cargo_home_paths() {
        let (_d, b) = base();
        let (home, why) = cargo_home("agent-7", &b).unwrap();
        assert_eq!(home, None);
        assert!(why.contains("left unset"));

        // A shared prepared home is used when the lane's own is missing or unprepared.
        fs::create_dir_all(b.join("moai-rc-product-cargo/registry/src")).unwrap();
        fs::create_dir_all(b.join("moai-rc-cargo-agent-7/registry")).unwrap();
        let (home, _) = cargo_home("agent-7", &b).unwrap();
        assert_eq!(home, Some(b.join("moai-rc-product-cargo")));

        // The lane's own prepared home wins.
        fs::create_dir_all(b.join("moai-rc-cargo-agent-7/registry/src")).unwrap();
        let (home, _) = cargo_home("agent-7", &b).unwrap();
        assert_eq!(home, Some(b.join("moai-rc-cargo-agent-7")));
        // Another lane is not given this lane's home.
        let (other, _) = cargo_home("agent-8", &b).unwrap();
        assert_eq!(other, Some(b.join("moai-rc-product-cargo")));
    }

    /// A relative base or HOME, or a folder outside git, refuses before anything is made.
    #[test]
    fn refusals_come_before_any_write() {
        let (_d, b) = base();
        let home = b.join("home");
        fs::create_dir(&home).unwrap();
        let h = Some(home.as_os_str());
        assert!(resolve("agent-7", Path::new("tmp"), h, &b).is_err());
        assert!(resolve("a b", &b, h, &b).is_err());
        let err = resolve("agent-7", &b, h, &b).unwrap_err();
        assert!(err.contains("git rev-parse"), "{err}");
        for bad in [None, Some(OsStr::new("")), Some(OsStr::new("rel/home"))] {
            assert!(home_dir(bad).is_err(), "{bad:?}");
        }
        assert_eq!(home_dir(h).unwrap(), home);
        assert_eq!(fs::read_dir(&home).unwrap().count(), 0);
    }

    #[cfg(unix)]
    #[test]
    fn symbolic_link_cargo_homes_are_refused() {
        let (_d, b) = base();
        let elsewhere = b.join("elsewhere");
        fs::create_dir_all(elsewhere.join("registry/src")).unwrap();
        // A symbolic-link cargo home is refused, own or shared, prepared or not.
        std::os::unix::fs::symlink(&elsewhere, b.join("moai-rc-cargo-agent-9")).unwrap();
        assert!(cargo_home("agent-9", &b)
            .unwrap_err()
            .contains("is a symbolic link"));
        fs::write(b.join("moai-rc-cargo-agent-8"), "x").unwrap();
        assert!(cargo_home("agent-8", &b)
            .unwrap_err()
            .contains("is not a folder"));
        std::os::unix::fs::symlink(&elsewhere, b.join("moai-rc-product-cargo")).unwrap();
        assert!(cargo_home("agent-10", &b)
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
                "moai-rc-cargo-agent-8",
                "moai-rc-cargo-agent-9",
                "moai-rc-product-cargo"
            ]
        );
    }

    /// The naming of `tools/build_folder.py` `owner` (the CLI test runs both on one repository).
    #[test]
    fn folder_names() {
        let common = Path::new("/r/repo/.git");
        for (worktree, want) in [
            ("/r/repo", "repo"),
            ("/r/wt-one", "repo-wt-one"),
            ("/r/repo-x", "repo-x"),
            ("/x/repo", "repo"),
            ("/r/repox", "repo-repox"),
            ("/r/a b+c", "repo-a_b_c"),
            ("/r/한글", "repo-__"),
        ] {
            assert_eq!(
                folder_name(common, Path::new(worktree)).unwrap(),
                want,
                "{worktree}"
            );
        }
        // A bare repository is its own primary; a trailing `.git` is dropped from the project.
        assert_eq!(
            folder_name(Path::new("/r/proj.git"), Path::new("/r/w")).unwrap(),
            "proj-w"
        );
        assert_eq!(
            folder_name(Path::new("/r/my.git/.git"), Path::new("/r/my.git")).unwrap(),
            "my"
        );
        for bad in ["", ".", ".."] {
            assert!(safe(bad).is_err(), "{bad:?}");
        }
        assert_eq!(safe("..x").unwrap(), "..x");
    }

    fn owner_of(root: &Path, worktree: &str) -> Owner {
        Owner {
            common: root.join("repo/.git"),
            worktree: root.join(worktree),
            name: format!("repo-{worktree}"),
        }
    }

    #[cfg(unix)]
    #[test]
    fn prepare_makes_the_folder_and_record_once() {
        use std::os::unix::fs::PermissionsExt;
        let (_d, b) = base();
        let home = b.join("home");
        fs::create_dir(&home).unwrap();
        let one = owner_of(&b, "one");
        let dir = prepare(&home, &one).unwrap();
        assert_eq!(dir, build_folder(&home, "repo-one"));
        for p in [home.join(".moai"), home.join(".moai/build"), dir.clone()] {
            let mode = fs::metadata(&p).unwrap().permissions().mode() & 0o777;
            assert_eq!(mode, 0o700, "{}", p.display());
        }
        let rec = read_record(&dir).unwrap().unwrap();
        assert_eq!(rec["schema_version"], 1);
        assert_eq!(rec["worktree"], b.join("one").to_str().unwrap());
        assert_eq!(rec["git_common_dir"], b.join("repo/.git").to_str().unwrap());
        // Again: the same record, no temporary file left.
        assert_eq!(prepare(&home, &one).unwrap(), dir);
        let names: Vec<_> = fs::read_dir(&dir)
            .unwrap()
            .map(|e| e.unwrap().file_name())
            .collect();
        assert_eq!(names, [RECORD]);
        // The same folder name from another worktree: refused, the record kept.
        let two = Owner {
            name: "repo-one".into(),
            ..owner_of(&b, "two")
        };
        let err = prepare(&home, &two).unwrap_err();
        assert!(err.contains("belongs to the worktree"), "{err}");
        assert_eq!(read_record(&dir).unwrap().unwrap(), rec);
        // A record without a worktree is replaced.
        fs::write(dir.join(RECORD), r#"{"schema_version": 1}"#).unwrap();
        prepare(&home, &two).unwrap();
        assert_eq!(
            read_record(&dir).unwrap().unwrap()["worktree"],
            b.join("two").to_str().unwrap()
        );
        // An oversized record refuses.
        fs::write(dir.join(RECORD), " ".repeat(RECORD_MAX as usize + 1)).unwrap();
        assert!(prepare(&home, &two)
            .unwrap_err()
            .contains("at most 4096 bytes"));
    }

    #[cfg(unix)]
    #[test]
    fn prepare_follows_no_link() {
        let (_d, b) = base();
        let elsewhere = b.join("elsewhere");
        fs::create_dir(&elsewhere).unwrap();
        for (i, at) in [".moai", ".moai/build", ".moai/build/repo-one"]
            .iter()
            .enumerate()
        {
            let home = b.join(format!("home{i}"));
            let link = home.join(at);
            fs::create_dir_all(link.parent().unwrap()).unwrap();
            std::os::unix::fs::symlink(&elsewhere, &link).unwrap();
            let err = prepare(&home, &owner_of(&b, "one")).unwrap_err();
            assert!(err.contains("is a link; it is not followed"), "{at}: {err}");
            fs::remove_file(&link).unwrap();
            fs::write(&link, "x").unwrap();
            let err = prepare(&home, &owner_of(&b, "one")).unwrap_err();
            assert!(err.contains("is not a folder"), "{at}: {err}");
        }
        assert_eq!(fs::read_dir(&elsewhere).unwrap().count(), 0);
    }

    #[test]
    fn render_quotes_and_comments() {
        let e = LaneEnv {
            target_dir: PathBuf::from("/h/it's here/.moai/build/repo-a"),
            cargo_home: None,
            cargo_home_why: "CARGO_HOME left unset: x".into(),
        };
        let text = render("a", &e);
        assert!(text.contains("export CARGO_TARGET_DIR='/h/it'\\''s here/.moai/build/repo-a'\n"));
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
