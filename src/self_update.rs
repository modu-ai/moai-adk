//! Replacing the running moai-cli binary (Issue #2195 E3, REQ-E3-01; decision
//! 16:1x(11-08): the self-replace pattern via the `self-replace` crate 1.5.1).
//!
//! On Windows a running executable cannot be overwritten in place; `self-replace`
//! renames the running file aside, puts the new one at its path and schedules the old
//! one for deletion. On other platforms it is an atomic rename. This module adds the
//! known-good contract around that call:
//!
//! 1. the new binary must be a regular file whose digest equals the expected digest,
//!    checked BEFORE anything is touched;
//! 2. the current binary's digest is recorded as the known-good;
//! 3. on a replacer error the current path is hashed again — equal to the known-good
//!    means it was kept (`state: kept`), anything else is a loud error;
//! 4. on success the current path must now hash to the expected digest.
//!
//! The production caller is the binary update path (04-B, not built yet), which passes
//! [`SelfReplaceCrate`] and `std::env::current_exe()`. No CLI operation is added here.
//! Tests drive [`replace_running`] with a mock [`Replacer`]; the real replacer is never
//! run in tests because it would replace the test binary itself.

use crate::install::sha256_file;
use serde_json::{json, Value};
use std::{fs, io, path::Path};

type Result<T> = std::result::Result<T, String>;

/// Puts `new_binary` at `target`. `target` is the running executable in production.
pub trait Replacer {
    fn replace(&self, target: &Path, new_binary: &Path) -> io::Result<()>;
}

/// The production replacer: `self_replace::self_replace`, which always acts on
/// `std::env::current_exe()`. It refuses any other `target` so a caller cannot believe
/// it replaced a path it did not.
pub struct SelfReplaceCrate;

impl Replacer for SelfReplaceCrate {
    fn replace(&self, target: &Path, new_binary: &Path) -> io::Result<()> {
        let current = std::env::current_exe()?;
        if fs::canonicalize(&current)? != fs::canonicalize(target)? {
            return Err(io::Error::new(
                io::ErrorKind::InvalidInput,
                "self-replace only replaces the running executable",
            ));
        }
        self_replace::self_replace(new_binary)
    }
}

/// Replaces `current` with `new_binary` (see the module docs for the four steps).
/// Returns `{state: "replaced"|"kept", ...}`; a known-good that cannot be confirmed is
/// an `Err`, never a quiet success.
pub fn replace_running(
    replacer: &dyn Replacer,
    current: &Path,
    new_binary: &Path,
    expected_sha256: &str,
) -> Result<Value> {
    let link = fs::symlink_metadata(new_binary).map_err(|_| "new binary unavailable")?;
    if link.file_type().is_symlink() {
        return Err("new binary is a symlink; refused".into());
    }
    let new_digest = sha256_file(new_binary)?;
    if new_digest != expected_sha256 {
        return Err(format!(
            "new binary digest mismatch: expected {expected_sha256}, found {new_digest}; nothing replaced"
        ));
    }
    let known_good = sha256_file(current)?;
    match replacer.replace(current, new_binary) {
        Err(e) => {
            let now = sha256_file(current)
                .map_err(|_| format!("replace failed ({e}) and the known-good is unreadable"))?;
            if now != known_good {
                return Err(format!(
                    "replace failed ({e}) and the binary changed: known-good {known_good}, now {now}"
                ));
            }
            Ok(json!({
                "state": "kept",
                "reason": e.to_string(),
                "known_good_sha256": known_good,
            }))
        }
        Ok(()) => {
            let now = sha256_file(current)?;
            if now != expected_sha256 {
                return Err(format!(
                    "replace reported success but the binary hashes to {now}, not {expected_sha256}"
                ));
            }
            Ok(json!({
                "state": "replaced",
                "previous_sha256": known_good,
                "sha256": now,
            }))
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::cell::Cell;
    use tempfile::TempDir;

    /// Mock replacer: counts calls and does what `mode` says to the target.
    struct Mock {
        mode: &'static str,
        calls: Cell<u32>,
    }

    impl Mock {
        fn new(mode: &'static str) -> Self {
            Mock {
                mode,
                calls: Cell::new(0),
            }
        }
    }

    impl Replacer for Mock {
        fn replace(&self, target: &Path, new_binary: &Path) -> io::Result<()> {
            self.calls.set(self.calls.get() + 1);
            match self.mode {
                "ok" => fs::copy(new_binary, target).map(|_| ()),
                "fail" => Err(io::Error::new(io::ErrorKind::PermissionDenied, "locked")),
                "corrupt-then-fail" => {
                    fs::write(target, b"half written")?;
                    Err(io::Error::other("interrupted"))
                }
                "ok-but-wrong" => fs::write(target, b"something else"),
                _ => unreachable!(),
            }
        }
    }

    fn setup() -> (TempDir, std::path::PathBuf, std::path::PathBuf, String) {
        let d = TempDir::new().unwrap();
        let current = d.path().join("moai-cli.exe");
        let new = d.path().join("moai-cli-new.exe");
        fs::write(&current, b"known good build").unwrap();
        fs::write(&new, b"new build").unwrap();
        let expected = crate::digest(b"new build");
        (d, current, new, expected)
    }

    #[test]
    fn a_verified_new_binary_replaces_the_current_one() {
        let (_d, current, new, expected) = setup();
        let mock = Mock::new("ok");
        let v = replace_running(&mock, &current, &new, &expected).unwrap();
        assert_eq!(v["state"], "replaced");
        assert_eq!(v["previous_sha256"], crate::digest(b"known good build"));
        assert_eq!(fs::read(&current).unwrap(), b"new build");
        assert_eq!(mock.calls.get(), 1);
    }

    #[test]
    fn a_failed_replace_keeps_the_known_good() {
        let (_d, current, new, expected) = setup();
        let mock = Mock::new("fail");
        let v = replace_running(&mock, &current, &new, &expected).unwrap();
        assert_eq!(v["state"], "kept");
        assert_eq!(v["known_good_sha256"], crate::digest(b"known good build"));
        assert!(v["reason"].as_str().unwrap().contains("locked"));
        assert_eq!(fs::read(&current).unwrap(), b"known good build");
    }

    #[test]
    fn a_digest_mismatch_is_refused_before_the_replacer_runs() {
        let (_d, current, new, _) = setup();
        let mock = Mock::new("ok");
        let wrong = crate::digest(b"some other build");
        let err = replace_running(&mock, &current, &new, &wrong).unwrap_err();
        assert!(err.contains("digest mismatch"), "{err}");
        assert_eq!(mock.calls.get(), 0, "the replacer must not run");
        assert_eq!(fs::read(&current).unwrap(), b"known good build");
    }

    #[test]
    fn a_failed_replace_that_changed_the_binary_is_a_loud_error() {
        let (_d, current, new, expected) = setup();
        let mock = Mock::new("corrupt-then-fail");
        let err = replace_running(&mock, &current, &new, &expected).unwrap_err();
        assert!(err.contains("binary changed"), "{err}");
    }

    #[test]
    fn a_success_that_left_the_wrong_bytes_is_an_error() {
        let (_d, current, new, expected) = setup();
        let mock = Mock::new("ok-but-wrong");
        let err = replace_running(&mock, &current, &new, &expected).unwrap_err();
        assert!(err.contains("reported success"), "{err}");
    }

    #[cfg(unix)]
    #[test]
    fn a_symlinked_new_binary_is_refused() {
        let (d, current, new, expected) = setup();
        let link = d.path().join("link.exe");
        std::os::unix::fs::symlink(&new, &link).unwrap();
        let mock = Mock::new("ok");
        let err = replace_running(&mock, &current, &link, &expected).unwrap_err();
        assert!(err.contains("symlink"), "{err}");
        assert_eq!(mock.calls.get(), 0);
    }

    #[test]
    fn the_real_replacer_refuses_a_target_that_is_not_the_running_executable() {
        // Safe to run: the target check happens before self_replace is called, so the
        // test binary is never touched.
        let (_d, current, new, _) = setup();
        let err = SelfReplaceCrate.replace(&current, &new).unwrap_err();
        assert_eq!(err.kind(), io::ErrorKind::InvalidInput);
    }
}
