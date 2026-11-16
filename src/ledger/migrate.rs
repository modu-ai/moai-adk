//! Schema version rules of the ledger (Issue #3068 P1). Version 1 is the first and only version,
//! so there is nothing to upgrade yet: this module decides whether a file's `user_version` may
//! be opened. When version 2 exists, the upgrade (with the one `ledger.db.bak` copy through the
//! SQLite backup method) lands here.

use super::{LedgerError, SCHEMA_VERSION};

/// Whether a file at `found` may be opened by this build: `found` equals [`SCHEMA_VERSION`].
/// A newer file is `schema_newer` and is left alone; a file below 1 has no schema (an empty or
/// foreign database) and is `ledger_unavailable`.
pub fn check_version(found: i64) -> Result<(), LedgerError> {
    match found {
        v if v == SCHEMA_VERSION => Ok(()),
        v if v > SCHEMA_VERSION => Err(LedgerError::SchemaNewer {
            found: v,
            supported: SCHEMA_VERSION,
        }),
        v => Err(LedgerError::Unavailable(format!(
            "the file has schema version {v}, not a ledger"
        ))),
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn only_the_current_version_opens_at_its_edges() {
        assert!(matches!(check_version(0), Err(LedgerError::Unavailable(_))));
        assert!(matches!(
            check_version(-1),
            Err(LedgerError::Unavailable(_))
        ));
        assert_eq!(check_version(1), Ok(()));
        assert_eq!(
            check_version(2),
            Err(LedgerError::SchemaNewer {
                found: 2,
                supported: 1
            })
        );
    }
}
