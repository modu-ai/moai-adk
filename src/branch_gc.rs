//! `branch.gc.plan` and `branch.gc.apply` (SPEC-20261123-branch-gc): stub that the tests
//! fail against before the code exists.

use serde_json::Value;
use std::path::Path;

type Result<T> = std::result::Result<T, String>;

pub fn plan(_project: &Path, _integration_ref: &str, _home: Option<&Path>) -> Result<Value> {
    Err("not implemented".into())
}

pub fn apply(
    _project: &Path,
    _integration_ref: &str,
    _plan_digest: &str,
    _permission_reference: &str,
    _home: Option<&Path>,
    _before_delete: &dyn Fn(&str),
) -> Result<Value> {
    Err("not implemented".into())
}
