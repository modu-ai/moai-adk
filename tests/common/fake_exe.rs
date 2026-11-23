//! A fake executable that leaves a marker (SPEC-20261123-ready-mutant-line).
//!
//! A test that proves "this program was never run" starts a fake and checks that its marker is
//! absent. A fake that writes its marker with `touch` proves nothing when the test gives the
//! child a PATH without `touch`: the fake runs, cannot find `touch`, leaves no marker, and the
//! absence passes whatever the code under test did. The helper writes its marker with the
//! shell's own redirection, so it needs no external command and no PATH.
use std::path::Path;

/// `text` inside single quotes, for the shell.
fn quoted(text: &str) -> String {
    format!("'{}'", text.replace('\'', "'\\''"))
}

/// A `/bin/sh` script that creates `marker` first, with no command from PATH, then runs `rest`
/// (script lines, or an empty string).
#[allow(dead_code)] // not every test crate starts a fake executable
pub fn marker_script(marker: &Path, rest: &str) -> String {
    format!(
        "#!/bin/sh\n: > {}\n{rest}",
        quoted(&marker.display().to_string())
    )
}

/// Writes the script of [`marker_script`] as an executable at `path` (mode 0755, through
/// `common::write_executable`).
#[allow(dead_code)] // not every test crate starts a fake executable
pub fn write_marking_executable(path: &Path, marker: &Path, rest: &str) {
    super::write_executable(path, &marker_script(marker, rest));
}
