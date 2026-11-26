//! The one place a test starts cargo on a scratch crate (SPEC-20261126-scratch-cargo-wrapper).
//!
//! A scratch crate must not build through the compiler wrapper of the outer run. A test that does
//! inherits the state of a cache server it does not look at: a stopped server that cannot start
//! again under the test's long temporary folder fails the test for that reason alone.
//!
//! `tests/test_scratch_cargo.rs` refuses a test file that starts cargo without this module.
#![allow(dead_code)] // each test crate that includes this module uses its own part of it

use std::ffi::{OsStr, OsString};
use std::process::Command;

/// The variables that make cargo run a compiler wrapper.
pub const WRAPPER_VARS: [&str; 4] = [
    "RUSTC_WRAPPER",
    "CARGO_BUILD_RUSTC_WRAPPER",
    "RUSTC_WORKSPACE_WRAPPER",
    "CARGO_BUILD_RUSTC_WORKSPACE_WRAPPER",
];

/// `command` without the wrapper variables, for a program that starts cargo itself.
pub fn without_wrapper(command: &mut Command) -> &mut Command {
    for name in WRAPPER_VARS {
        command.env_remove(name);
    }
    command
}

/// A command that starts the cargo at `program`, without the wrapper variables.
pub fn cargo_at(program: impl AsRef<OsStr>) -> Command {
    let mut command = Command::new(program);
    without_wrapper(&mut command);
    command
}

/// The cargo of this test run, as Cargo or nextest name it, else `cargo` from `PATH`.
pub fn cargo_path() -> OsString {
    std::env::var_os("CARGO").unwrap_or_else(|| "cargo".into())
}
