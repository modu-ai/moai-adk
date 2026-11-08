//! Experimental issue report formatting. This module does not send or redact data.

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum CheckStatus { Passed, Failed, NotRun, Blocked, Stale, Unknown }

impl CheckStatus {
    fn label(self) -> &'static str {
        match self {
            Self::Passed => "passed", Self::Failed => "failed", Self::NotRun => "not_run",
            Self::Blocked => "blocked", Self::Stale => "stale", Self::Unknown => "unknown",
        }
    }
}

#[derive(Debug, PartialEq, Eq)]
pub struct Report<'a> {
    pub work: &'a str,
    pub check_name: &'a str,
    pub check_status: CheckStatus,
    pub next: &'a str,
}

/// Callers must review and redact fields before formatting or publishing.
pub fn format_report(report: &Report<'_>) -> Result<String, &'static str> {
    for field in [report.work, report.check_name, report.next] {
        if field.trim().is_empty() || field.chars().count() > 300 || field.chars().any(char::is_control) {
            return Err("report fields must be nonempty, single-line, and at most 300 characters");
        }
    }
    Ok(format!("Work: {}\nCheck: {} — {}\nNext: {}", report.work, report.check_name, report.check_status.label(), report.next))
}

#[cfg(test)]
mod tests {
    use super::*;

    fn report() -> Report<'static> {
        Report { work: "Implemented parser", check_name: "Unit tests", check_status: CheckStatus::NotRun, next: "Run tests" }
    }

    #[test]
    fn formats_concise_report_with_observed_status() {
        let text = format_report(&report()).unwrap();
        assert_eq!(text, "Work: Implemented parser\nCheck: Unit tests — not_run\nNext: Run tests");
    }

    #[test]
    fn all_statuses_remain_distinct() {
        let statuses = [CheckStatus::Passed, CheckStatus::Failed, CheckStatus::NotRun, CheckStatus::Blocked, CheckStatus::Stale, CheckStatus::Unknown];
        for status in statuses {
            let mut item = report(); item.check_status = status;
            assert!(format_report(&item).unwrap().contains(status.label()));
        }
    }

    #[test]
    fn rejects_empty_multiline_and_oversized_fields() {
        let mut item = report(); item.work = " "; assert!(format_report(&item).is_err());
        item.work = "log\nsecret"; assert!(format_report(&item).is_err());
        let long = "x".repeat(301);
        item.work = &long; assert!(format_report(&item).is_err());
    }
}
