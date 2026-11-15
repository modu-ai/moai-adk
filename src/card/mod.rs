//! `moai card` pieces that need no ledger connection (#3070).
//!
//! `export_format` is the file format of the ledger export: plain values in, bytes out, and
//! the same bytes back in; `export_dir` puts those bytes in a folder and reads them out, under the
//! rules a user's own folder needs. The commands that read the ledger and write the files come with
//! the ledger itself. `input` checks what a user types for `card add` and reads a card's name
//! (`C-12`, `#3066`).

pub mod export_dir;
pub mod export_format;
pub mod input;
