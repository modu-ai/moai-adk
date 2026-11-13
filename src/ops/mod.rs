//! Internal operations that live one per file (#2617, #2388 AC-2).
//!
//! `build.rs` collects every `src/ops/<word>.rs` into the `Ops` subcommand enum below;
//! `src/main.rs` flattens it into `Command` once. To add an operation, add one file here
//! (and its document fragment `docs/design-review/appendix/operations/<word>.md`): the
//! subcommand is `<word>` with `_` as `-`, and the file declares
//!
//! - `pub struct Args` — the clap arguments; its `///` lines become the subcommand help,
//! - `pub const OPERATION: &str` — the envelope operation name,
//! - `pub fn run(args: Args) -> Result<serde_json::Value, String>`.
//!
//! Nothing else is shared, so two branches that each add an operation do not touch the same
//! lines. Operations that need the typed `OpError` path stay in `src/main.rs` for now.

include!(concat!(env!("OUT_DIR"), "/ops_registry.rs"));
