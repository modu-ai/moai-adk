//! `moai card` (#3070). `command` is what `card list | show | add` do with the ledger and
//! `dispatch` what `card ready | take | release | handoff` do; `work` is the rest of a card's
//! life (`card next | note | criterion | pr | review | done | renew | check`, card C-53). The
//! other modules need no ledger connection.
//!
//! `export_format` is the file format of the ledger export: plain values in, bytes out, and
//! the same bytes back in; `export_dir` puts those bytes in a folder and reads them out, under the
//! rules a user's own folder needs. `export_db` reads the ledger's five exported tables into
//! those rows, and `command::export` joins the three for `card export`; restoring from a folder
//! is not here yet. `input` checks what a user types for `card add` and reads a card's name
//! (`C-12`, `#3066`). `spec_link` is the card's link to the SPEC it works on (`card spec`,
//! SDD-4). `inbox` is the leader's read-only list of the lane notes that wait for an answer
//! (`card inbox`, card C-116). `precheck` is the leader's read-only look before readying or
//! dispatching a card (card C-96): already-done on the tip, or colliding with a recorded
//! decision.

pub mod command;
pub mod dispatch;
pub mod early_reclaim;
pub mod export_db;
pub mod export_dir;
pub mod export_format;
pub mod inbox;
pub mod input;
pub mod precheck;
pub mod spec_link;
pub mod timeline;
pub mod work;
