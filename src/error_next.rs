//! The `next:` line of the stderr error frame (card #2804, design #2799 D2 §4).
//!
//! Only the kinds whose next step is the same for every message get a line here — the groups
//! the D1 addendum (Issue #2799 comment 6865269016) counted: `permission_reference_missing` and
//! `unsupported_protocol_version` (22 messages each, none naming a next step), `stale_plan`,
//! `input_too_large` and command-line `usage` errors. A message that already names its next
//! step keeps it and gets no second one. The envelope never carries this line: it is human
//! output on a terminal stderr only.

use crate::contract::Kind;

/// Phrases that mark a message as already naming its next step (the D1 count's rule).
const NEXT_PHRASES: &[&str] = &[
    "run `",
    "use `",
    "pass `",
    "see `",
    "see --help",
    "preview again",
    "re-preview",
    "repreview",
    "re-run",
    "rerun",
    "try ",
    "first",
    "; ask",
    "then ",
    "query again",
];

/// Whether `message` already tells the reader what to do next.
pub fn names_next_step(message: &str) -> bool {
    NEXT_PHRASES.iter().any(|p| message.contains(p))
}

/// The next step for a failure of `kind`, or `None` when the kind has no single next step or
/// the message already names one. `program` and `word` name the command as the user ran it
/// (`moai`, `project-plan`), for a usage error's help pointer.
pub fn next_action(kind: Kind, program: &str, word: Option<&str>, message: &str) -> Option<String> {
    if let Kind::Usage = kind {
        // A usage message ends in "see --help"; point at the command's own help instead.
        return Some(match word {
            Some(w) => format!("run `{program} {w} --help` to see its arguments"),
            None => format!("run `{program} --help` to see the commands"),
        });
    }
    if names_next_step(message) {
        return None;
    }
    let line = match kind {
        Kind::PermissionReferenceMissing => {
            "pass `--permission-reference <ref>`, the approval reference the host gives for this plan"
        }
        Kind::UnsupportedProtocolVersion => {
            "send version 1 (`protocol_version: 1`, or `schema_version: 1` in a request file)"
        }
        Kind::StalePlan => "make the plan again and apply it with the new digest",
        Kind::InputTooLarge => "send a smaller input",
        _ => return None,
    };
    Some(line.to_string())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn each_listed_kind_gets_its_line_and_other_kinds_none() {
        let m = "something failed";
        assert!(
            next_action(Kind::PermissionReferenceMissing, "moai", None, m)
                .unwrap()
                .contains("--permission-reference")
        );
        assert!(
            next_action(Kind::UnsupportedProtocolVersion, "moai", None, m)
                .unwrap()
                .contains("protocol_version: 1")
        );
        assert!(next_action(Kind::StalePlan, "moai", None, m)
            .unwrap()
            .contains("plan again"));
        assert_eq!(
            next_action(Kind::InputTooLarge, "moai", None, m).as_deref(),
            Some("send a smaller input")
        );
        for k in [
            Kind::Io,
            Kind::Internal,
            Kind::Conflict,
            Kind::InvalidInput,
            Kind::NotFound,
        ] {
            assert_eq!(next_action(k, "moai", None, m), None, "{k:?}");
        }
    }

    #[test]
    fn a_message_that_names_its_next_step_keeps_it() {
        let m = "the package changed while it was read; nothing applied, preview again";
        assert_eq!(next_action(Kind::StalePlan, "moai", None, m), None);
        assert_eq!(
            next_action(
                Kind::PermissionReferenceMissing,
                "moai",
                None,
                "x; run `moai y`"
            ),
            None
        );
    }

    #[test]
    fn a_usage_error_points_at_the_commands_own_help() {
        let m = "required arguments were not provided; see --help";
        assert_eq!(
            next_action(Kind::Usage, "moai", Some("project-plan"), m).as_deref(),
            Some("run `moai project-plan --help` to see its arguments")
        );
        assert_eq!(
            next_action(Kind::Usage, "moai-cli", None, m).as_deref(),
            Some("run `moai-cli --help` to see the commands")
        );
    }
}
