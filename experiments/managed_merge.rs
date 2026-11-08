//! Experimental three-way decision for a single managed setting.
//! The caller owns parsing, snapshot validation, persistence, and recovery.

#[derive(Debug, PartialEq, Eq)]
pub enum Decision<'a> {
    KeepCurrent,
    Apply(Option<&'a str>),
    Conflict,
}

/// Distinguishes missing values from present values and never overwrites drift.
pub fn decide<'a>(baseline: Option<&str>, current: Option<&str>, desired: Option<&'a str>) -> Decision<'a> {
    if current == desired || desired == baseline {
        Decision::KeepCurrent
    } else if current == baseline {
        Decision::Apply(desired)
    } else {
        Decision::Conflict
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn creates_absent_managed_value() {
        assert_eq!(decide(None, None, Some("new")), Decision::Apply(Some("new")));
    }

    #[test]
    fn updates_and_deletes_only_unchanged_baseline() {
        assert_eq!(decide(Some("old"), Some("old"), Some("new")), Decision::Apply(Some("new")));
        assert_eq!(decide(Some("old"), Some("old"), None), Decision::Apply(None));
    }

    #[test]
    fn preserves_user_edit_when_template_has_not_changed() {
        assert_eq!(decide(Some("old"), Some("user"), Some("old")), Decision::KeepCurrent);
        assert_eq!(decide(Some("old"), None, Some("old")), Decision::KeepCurrent);
    }

    #[test]
    fn refuses_unowned_or_concurrently_modified_values() {
        assert_eq!(decide(None, Some("user"), Some("new")), Decision::Conflict);
        assert_eq!(decide(Some("old"), Some("user"), Some("new")), Decision::Conflict);
        assert_eq!(decide(Some("old"), None, Some("new")), Decision::Conflict);
    }

    #[test]
    fn already_applied_is_idempotent() {
        assert_eq!(decide(Some("old"), Some("new"), Some("new")), Decision::KeepCurrent);
        assert_eq!(decide(None, None, None), Decision::KeepCurrent);
    }
}
