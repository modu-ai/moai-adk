//! The step summary (the Go build's `gate_summary.go`): one row per configured step.

use super::fmtutil::{go_duration, local_stamp, round_ms};
use std::time::{Duration, SystemTime};

#[derive(Debug, Clone, PartialEq, Eq)]
pub enum Outcome {
    NotReached,
    Executed,
    Skipped,
    Disabled,
    /// This build does not run the step yet (never silent: it is named in the summary).
    NotPorted,
}

#[derive(Debug, Clone)]
pub struct Record {
    pub label: String,
    pub outcome: Outcome,
    pub command: String,
    pub duration: Duration,
    pub started_at: Option<SystemTime>,
    pub reason: String,
}

#[derive(Debug, Default)]
pub struct Summary {
    records: Vec<Record>,
    /// Axes that are not steps of any language (graph freshness, ast-grep) and that this build
    /// does not run yet; printed after the rows.
    axes_not_ported: Vec<String>,
}

pub const REASON_DISABLED: &str = "turned off by the gate.disabled_steps setting";
pub const REASON_SKIP_TESTS: &str = "turned off by the gate.skip_tests setting";

impl Summary {
    pub fn new() -> Self {
        Self::default()
    }

    /// Adds a `not reached` row unless the label is already there.
    pub fn seed(&mut self, label: &str) {
        if label.is_empty() || self.find(label).is_some() {
            return;
        }
        self.records.push(Record {
            label: label.into(),
            outcome: Outcome::NotReached,
            command: String::new(),
            duration: Duration::ZERO,
            started_at: None,
            reason: String::new(),
        });
    }

    fn find(&self, label: &str) -> Option<usize> {
        self.records.iter().position(|r| r.label == label)
    }

    pub fn relabel(&mut self, from: &str, to: &str) {
        if from == to {
            return;
        }
        if let Some(i) = self.find(from) {
            self.records[i].label = to.into();
        }
    }

    fn mark(&mut self, label: &str, outcome: Outcome, reason: &str, command: &str) -> &mut Record {
        self.seed(label);
        let i = self.find(label).expect("seeded");
        let r = &mut self.records[i];
        r.outcome = outcome;
        r.reason = reason.into();
        r.command = command.into();
        r.started_at = None;
        r.duration = Duration::ZERO;
        r
    }

    pub fn skipped(&mut self, label: &str, reason: &str) {
        self.mark(label, Outcome::Skipped, reason, "");
    }

    pub fn disabled(&mut self, label: &str, reason: &str) {
        self.mark(label, Outcome::Disabled, reason, "");
    }

    pub fn not_ported(&mut self, label: &str) {
        self.mark(label, Outcome::NotPorted, "", "");
    }

    pub fn executed(&mut self, label: &str, command: &str, started_at: SystemTime, d: Duration) {
        let r = self.mark(label, Outcome::Executed, "", command);
        r.started_at = Some(started_at);
        r.duration = d;
    }

    pub fn axis_not_ported(&mut self, axis: &str) {
        if !self.axes_not_ported.iter().any(|a| a == axis) {
            self.axes_not_ported.push(axis.into());
        }
    }

    pub fn outcome_of(&self, label: &str) -> Option<&Outcome> {
        self.find(label).map(|i| &self.records[i].outcome)
    }

    /// The summary text, or `""` when there is nothing to say.
    pub fn render(&self) -> String {
        if self.records.is_empty() && self.axes_not_ported.is_empty() {
            return String::new();
        }
        let mut out = format!("quality gate steps ({} configured):", self.records.len());
        for r in &self.records {
            out.push_str("\n  - ");
            out.push_str(&r.label);
            out.push_str(": ");
            match r.outcome {
                Outcome::Executed => {
                    let at = r.started_at.map(local_stamp).unwrap_or_default();
                    out.push_str(&format!(
                        "executed in {} at {at}",
                        go_duration(round_ms(r.duration))
                    ));
                    if !r.command.is_empty() {
                        out.push_str(" — ");
                        out.push_str(&r.command);
                    }
                }
                Outcome::Skipped | Outcome::Disabled => {
                    out.push_str(if r.outcome == Outcome::Skipped {
                        "skipped"
                    } else {
                        "disabled"
                    });
                    if !r.reason.is_empty() {
                        out.push_str(" — ");
                        out.push_str(&r.reason);
                    }
                }
                Outcome::NotReached => out.push_str("not reached"),
                Outcome::NotPorted => out.push_str("not run — not ported yet"),
            }
        }
        for axis in &self.axes_not_ported {
            out.push_str(&format!("\n  - {axis}: not run — not ported yet"));
        }
        out
    }
}

/// `binary arg arg`, as the summary shows a command.
pub fn command_line(binary: &str, args: &[String]) -> String {
    if args.is_empty() {
        binary.to_string()
    } else {
        format!("{binary} {}", args.join(" "))
    }
}

/// `"typecheck: skipped (why)"` -> `why` (the Go build's `condenseSkipReason`).
pub fn condense_skip_reason(label: &str, reason: &str) -> String {
    let r = reason.trim();
    let r = r
        .strip_prefix(&format!("{label}: skipped"))
        .unwrap_or(r)
        .trim();
    let inner = r.strip_prefix('(').unwrap_or(r);
    let inner = inner.strip_suffix(')').unwrap_or(inner).trim();
    if inner.is_empty() {
        reason.trim().to_string()
    } else {
        inner.to_string()
    }
}

/// Two text blocks joined by one blank line, either may be empty.
pub fn join_blocks(first: &str, second: &str) -> String {
    match (first.trim(), second.trim()) {
        ("", b) => b.to_string(),
        (a, "") => a.to_string(),
        (a, b) => format!("{a}\n\n{b}"),
    }
}
