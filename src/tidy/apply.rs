//! `tidy-apply` (SPEC-20261126-home-tidy-apply): act on the picked items of a freshly
//! recomputed plan.

use super::Options;
use serde_json::Value;
use std::path::Path;

/// What the caller picked and approved.
pub struct Request<'a> {
    /// Plan item paths, relative to the moai folder.
    pub picks: &'a [String],
    /// The `digest` of the plan the picks were made from.
    pub confirm: &'a str,
    /// The caller's reference for the user's approval.
    pub permission: &'a str,
}

/// A point of a run a test may act at (SPEC D-4); the operation does nothing there.
pub enum Step<'a> {
    /// The archive and the list of `item` are written in `dir`; nothing is read back yet.
    Written { item: &'a str, dir: &'a Path },
    /// The checks of `item` passed and its first removal comes next; `run` holds the receipt.
    Removing { item: &'a str, run: &'a Path },
}

/// Applies the picked items of the plan of `<home>/.moai`.
pub fn apply(home: &Path, opts: &Options, request: &Request) -> Result<Value, String> {
    apply_with(home, opts, request, &mut |_| {})
}

/// [`apply`] with a hook at each [`Step`].
#[doc(hidden)]
pub fn apply_with(
    _home: &Path,
    _opts: &Options,
    _request: &Request,
    _at: &mut dyn FnMut(Step<'_>),
) -> Result<Value, String> {
    Err("not implemented".into())
}
