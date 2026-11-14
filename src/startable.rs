//! The one startable judgment (#2901 O2): whether an open Issue is work a lane may start now.
//! `lane-next`, `goal-block` and `queue-free` all call [`judge`], so the three commands cannot
//! disagree about a card, and [`end_condition`] is the lane-wide end of the loop (no open pull
//! request and no startable Issue, #2515).
//!
//! An open Issue is not startable, checked in this order, when it is
//!
//! 1. `held` — a label named `needs:decision`, `hold` or `<axis>:hold` (whole names only);
//! 2. `keep_set` — the active goal leaves it out (`goal-plan` exclusions, `goal-block`
//!    `keep_set`);
//! 3. `epic` — its title carries `][EP-`, or GitHub reports open sub-issues;
//! 4. `blocked_by` — its body's `Blocked-By:` line names an Issue that is still open, or GitHub
//!    reports an open Issue blocking it.
//!
//! A claim is not part of the judgment: a claimed card is still work left in the repository.
//! Whether another lane's claim keeps a card from this lane is the caller's step.

use crate::goal::Exclusion;
use serde_json::{json, Value};
use std::collections::BTreeSet;

/// Whether a label marks a card that waits for a decision (`needs:decision`, `HOLD`, …).
pub fn holds(label: &str) -> bool {
    // Whole names only (#2675 audit N4): a label that merely contains `hold` (`threshold`)
    // must not hide a card. `needs:decision` is the label scheme's gate marker
    // (`.github/moai/labels.yml`); `hold` and a `<axis>:hold` value are the HOLD spellings.
    let l = label.trim().to_ascii_lowercase();
    l == "needs:decision" || l == "hold" || l.ends_with(":hold")
}

/// What the judgment reads of one open Issue.
#[derive(Clone, Copy, Debug)]
pub struct Card<'a> {
    pub number: u64,
    pub title: &'a str,
    pub labels: &'a [String],
    pub body: &'a str,
    /// Sub-issues GitHub reports still open.
    pub open_children: u64,
    /// Open Issues GitHub reports as blocking this one.
    pub blocked_by: &'a [u64],
}

/// Why an open Issue is not startable.
#[derive(Clone, Debug, PartialEq, Eq)]
pub enum Skip {
    /// The hold labels it carries.
    Held(Vec<String>),
    /// The goal's reason for leaving it out.
    KeepSet(String),
    Epic {
        open_children: u64,
    },
    /// The open Issues it waits for.
    BlockedBy(Vec<u64>),
}

impl Skip {
    pub fn reason(&self) -> &'static str {
        match self {
            Skip::Held(_) => "held",
            Skip::KeepSet(_) => "keep_set",
            Skip::Epic { .. } => "epic",
            Skip::BlockedBy(_) => "blocked_by",
        }
    }

    /// The `skipped` / `held` entry: `{issue, reason}` and the detail of the reason.
    pub fn entry(&self, issue: u64) -> Value {
        let mut v = json!({"issue": issue, "reason": self.reason()});
        match self {
            Skip::Held(labels) => v["labels"] = json!(labels),
            Skip::KeepSet(why) => v["keep_reason"] = json!(why),
            Skip::Epic { open_children } => v["open_children"] = json!(open_children),
            Skip::BlockedBy(by) => v["blocked_by"] = json!(by),
        }
        v
    }
}

/// `None` when `card` is startable; otherwise the first reason it is not. `open` holds the
/// open Issue numbers (a `Blocked-By:` reference outside it is closed), `keep` the goal's
/// exclusions.
pub fn judge(card: &Card<'_>, open: &BTreeSet<u64>, keep: &[Exclusion]) -> Option<Skip> {
    let held: Vec<String> = card.labels.iter().filter(|l| holds(l)).cloned().collect();
    if !held.is_empty() {
        return Some(Skip::Held(held));
    }
    if let Some(k) = keep.iter().find(|k| k.number == card.number) {
        return Some(Skip::KeepSet(k.reason.clone()));
    }
    if card.title.contains("][EP-") || card.open_children > 0 {
        return Some(Skip::Epic {
            open_children: card.open_children,
        });
    }
    let mut blockers: Vec<u64> = crate::queue::refs_on(card.body, "Blocked-By:")
        .into_iter()
        .filter_map(|n| u64::try_from(n).ok())
        .filter(|n| open.contains(n))
        .chain(card.blocked_by.iter().copied())
        .filter(|n| *n != card.number)
        .collect();
    blockers.sort_unstable();
    blockers.dedup();
    (!blockers.is_empty()).then_some(Skip::BlockedBy(blockers))
}

/// The lane-wide end condition: `met` is true when no pull request is open and no Issue is
/// startable, false as soon as either count is known to be above zero, and null while a count
/// is unknown (a list not read, or read only in part).
pub fn end_condition(open_prs: Option<u64>, startable: Option<u64>) -> Value {
    let met = match (open_prs, startable) {
        (Some(p), _) if p > 0 => Some(false),
        (_, Some(s)) if s > 0 => Some(false),
        (Some(0), Some(0)) => Some(true),
        _ => None,
    };
    json!({"open_prs": open_prs, "startable": startable, "met": met})
}

#[cfg(test)]
mod tests {
    use super::*;

    fn card<'a>(number: u64, title: &'a str, labels: &'a [String], body: &'a str) -> Card<'a> {
        Card {
            number,
            title,
            labels,
            body,
            open_children: 0,
            blocked_by: &[],
        }
    }

    /// Whole label names only (#2675 audit N4): `threshold` holds nothing.
    #[test]
    fn hold_labels_are_whole_names() {
        for l in [
            "needs:decision",
            "HOLD",
            " hold ",
            "status:hold",
            "Needs:Decision",
        ] {
            assert!(holds(l), "{l}");
        }
        for l in [
            "threshold",
            "holder",
            "needs:decision-later",
            "on-hold-ish",
            "",
        ] {
            assert!(!holds(l), "{l}");
        }
    }

    /// Each reason, in the judgment's order, and a card that is startable.
    #[test]
    fn each_reason_in_order() {
        let open: BTreeSet<u64> = [1, 2, 3, 4, 5].into();
        let keep = vec![Exclusion {
            number: 2,
            reason: "운영자 보유".into(),
        }];
        let none: Vec<String> = vec![];
        let held = vec!["P2".to_string(), "needs:decision".to_string()];
        // Held wins over every later reason.
        let c = Card {
            open_children: 3,
            ..card(2, "[x][EP-1] e", &held, "Blocked-By: #1")
        };
        assert_eq!(
            judge(&c, &open, &keep),
            Some(Skip::Held(vec!["needs:decision".into()]))
        );
        assert_eq!(
            judge(&card(2, "[x][EP-1] e", &none, ""), &open, &keep),
            Some(Skip::KeepSet("운영자 보유".into()))
        );
        assert_eq!(
            judge(
                &card(3, "[x][EP-1] e", &none, "Blocked-By: #1"),
                &open,
                &keep
            ),
            Some(Skip::Epic { open_children: 0 })
        );
        let parent = Card {
            open_children: 1,
            ..card(3, "plain", &none, "")
        };
        assert_eq!(
            judge(&parent, &open, &keep),
            Some(Skip::Epic { open_children: 1 })
        );
        assert_eq!(
            judge(
                &card(4, "t", &none, "Blocked-By: #5, #1, #999, #4"),
                &open,
                &keep
            ),
            Some(Skip::BlockedBy(vec![1, 5]))
        );
        let gh = [5u64, 1];
        let by_github = Card {
            blocked_by: &gh,
            ..card(4, "t", &none, "Blocked-By: #1")
        };
        assert_eq!(
            judge(&by_github, &open, &keep),
            Some(Skip::BlockedBy(vec![1, 5]))
        );
        // A closed blocker, a self reference and no line at all: startable.
        assert_eq!(
            judge(&card(4, "t", &none, "Blocked-By: #999 #4"), &open, &keep),
            None
        );
        assert_eq!(judge(&card(5, "t", &none, ""), &open, &[]), None);
    }

    #[test]
    fn entries_name_the_reason_and_its_detail() {
        assert_eq!(
            Skip::Held(vec!["HOLD".into()]).entry(7),
            json!({"issue": 7, "reason": "held", "labels": ["HOLD"]})
        );
        assert_eq!(
            Skip::KeepSet("x".into()).entry(7),
            json!({"issue": 7, "reason": "keep_set", "keep_reason": "x"})
        );
        assert_eq!(
            Skip::Epic { open_children: 2 }.entry(7),
            json!({"issue": 7, "reason": "epic", "open_children": 2})
        );
        assert_eq!(
            Skip::BlockedBy(vec![3]).entry(7),
            json!({"issue": 7, "reason": "blocked_by", "blocked_by": [3]})
        );
    }

    /// #2901 AC5: the end condition is met only with both counts known and zero; any count
    /// above zero is not met; an unknown count otherwise leaves it open (null).
    #[test]
    fn end_condition_goldens() {
        let cases = [
            (Some(0), Some(0), json!(true)),
            (Some(1), Some(0), json!(false)),
            (Some(0), Some(1), json!(false)),
            (Some(2), None, json!(false)),
            (None, Some(3), json!(false)),
            (None, Some(0), Value::Null),
            (Some(0), None, Value::Null),
            (None, None, Value::Null),
        ];
        for (prs, startable, met) in cases {
            assert_eq!(
                end_condition(prs, startable),
                json!({"open_prs": prs, "startable": startable, "met": met}),
                "{prs:?} {startable:?}"
            );
        }
    }
}
