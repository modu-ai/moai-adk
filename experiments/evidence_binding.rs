//! Experimental binding check; this is not full receipt admission.

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Subject<'a> {
    pub task_id: &'a str,
    pub attempt: u64,
    pub tree: &'a str,
    pub spec_digest: &'a str,
    pub policy_digest: &'a str,
    pub tool_version: &'a str,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Observation {
    Passed,
    Failed,
    NotRun,
    Blocked,
    Unknown,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum BindingResult {
    CandidatePass,
    Failed,
    NotRun,
    Blocked,
    Unknown,
    Stale,
    Invalid,
}

/// Checks identity and observed state only. A candidate pass is not admission.
pub fn check_binding(expected: &Subject<'_>, observed: &Subject<'_>, state: Observation) -> BindingResult {
    let valid = |s: &Subject<'_>| {
        s.attempt > 0 && [s.task_id, s.tree, s.spec_digest, s.policy_digest, s.tool_version]
            .iter()
            .all(|value| !value.trim().is_empty())
    };
    if !valid(expected) || !valid(observed) {
        return BindingResult::Invalid;
    }
    if expected != observed {
        return BindingResult::Stale;
    }
    match state {
        Observation::Passed => BindingResult::CandidatePass,
        Observation::Failed => BindingResult::Failed,
        Observation::NotRun => BindingResult::NotRun,
        Observation::Blocked => BindingResult::Blocked,
        Observation::Unknown => BindingResult::Unknown,
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn subject() -> Subject<'static> {
        Subject { task_id: "task-1", attempt: 1, tree: "tree-a", spec_digest: "spec-a", policy_digest: "policy-a", tool_version: "v1" }
    }

    #[test]
    fn current_pass_is_only_a_candidate() {
        assert_eq!(check_binding(&subject(), &subject(), Observation::Passed), BindingResult::CandidatePass);
    }

    #[test]
    fn changed_identity_is_stale() {
        let expected = subject();
        let variants = [
            Subject { task_id: "task-2", ..subject() },
            Subject { attempt: 2, ..subject() },
            Subject { tree: "tree-b", ..subject() },
            Subject { spec_digest: "spec-b", ..subject() },
            Subject { policy_digest: "policy-b", ..subject() },
            Subject { tool_version: "v2", ..subject() },
        ];
        for variant in variants {
            assert_eq!(check_binding(&expected, &variant, Observation::Passed), BindingResult::Stale);
        }
    }

    #[test]
    fn incomplete_identity_is_invalid() {
        assert_eq!(check_binding(&subject(), &Subject { tree: " ", ..subject() }, Observation::Passed), BindingResult::Invalid);
        assert_eq!(check_binding(&subject(), &Subject { attempt: 0, ..subject() }, Observation::Passed), BindingResult::Invalid);
    }

    #[test]
    fn nonpassing_observations_never_become_passes() {
        for (state, result) in [
            (Observation::Failed, BindingResult::Failed),
            (Observation::NotRun, BindingResult::NotRun),
            (Observation::Blocked, BindingResult::Blocked),
            (Observation::Unknown, BindingResult::Unknown),
        ] {
            assert_eq!(check_binding(&subject(), &subject(), state), result);
        }
    }
}
