//! Experimental host update decisions; performs no filesystem operations.

#[derive(Debug, PartialEq, Eq)]
pub enum RestoreDecision<'a> {
    Restore(Option<&'a str>),
    PreserveUserChange,
    InvalidReceipt,
}

/// A failed host may restore only a file still matching this run's after digest.
pub fn restore_decision<'a>(before: Option<&'a str>, after: &str, current: Option<&str>) -> RestoreDecision<'a> {
    if after.trim().is_empty() || before.is_some_and(|value| value.trim().is_empty()) {
        return RestoreDecision::InvalidReceipt;
    }
    if current == Some(after) {
        RestoreDecision::Restore(before)
    } else {
        RestoreDecision::PreserveUserChange
    }
}

/// The apply phase starts only after every targeted host passed preflight.
pub fn may_begin_apply(preflight: &[bool]) -> bool {
    !preflight.is_empty() && preflight.iter().all(|result| *result)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn every_target_needs_preflight() {
        assert!(!may_begin_apply(&[]));
        assert!(!may_begin_apply(&[true, false]));
        assert!(may_begin_apply(&[true, true]));
    }

    #[test]
    fn restores_only_matching_applied_content() {
        assert_eq!(restore_decision(Some("before"), "after", Some("after")), RestoreDecision::Restore(Some("before")));
        assert_eq!(restore_decision(None, "after", Some("after")), RestoreDecision::Restore(None));
    }

    #[test]
    fn preserves_subsequent_user_edit_or_deletion() {
        assert_eq!(restore_decision(Some("before"), "after", Some("user")), RestoreDecision::PreserveUserChange);
        assert_eq!(restore_decision(Some("before"), "after", None), RestoreDecision::PreserveUserChange);
    }

    #[test]
    fn incomplete_receipt_is_not_a_restore_authorization() {
        assert_eq!(restore_decision(Some("before"), "", Some("")), RestoreDecision::InvalidReceipt);
        assert_eq!(restore_decision(Some(" "), "after", Some("after")), RestoreDecision::InvalidReceipt);
    }
}
