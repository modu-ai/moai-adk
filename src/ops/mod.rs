//! Internal operations that live one per file (#2617, #2388 AC-2).
//!
//! `build.rs` lists every `src/ops/<word>.rs` by name in a generated `with_ops!` list, and
//! `declare_ops!` below turns it into the modules and the `Ops` subcommand enum;
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

/// Declares the listed operations next to this file: one `pub mod <word>;` each, the `Ops`
/// subcommand enum (each variant keeps the help lines of its `Args`), and its dispatch.
macro_rules! declare_ops {
    ($($word:ident $variant:ident [$($doc:literal),*])*) => {
        $(pub mod $word;)*

        /// The operations collected from the operation files, in file-name order.
        #[derive(clap::Subcommand)]
        pub enum Ops {
            $($(#[doc = $doc])* $variant($word::Args),)*
        }

        impl Ops {
            /// The envelope operation name.
            pub fn operation(&self) -> &'static str {
                match self {
                    $(Ops::$variant(_) => $word::OPERATION,)*
                }
            }

            /// Runs the operation; `main` wraps the result in the envelope.
            pub fn run(self) -> Result<serde_json::Value, String> {
                match self {
                    $(Ops::$variant(args) => $word::run(args),)*
                }
            }
        }
    };
}

// The generated `with_ops!` list: names and help lines only (SPEC-20261122-ops-registry-path-free).
include!(concat!(env!("OUT_DIR"), "/ops_registry.rs"));
with_ops!(declare_ops);
