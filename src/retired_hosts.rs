//! SPEC-20261121-retired-hosts-read-safely (REQ-1, REQ-2, REQ-10): the hosts whose support is
//! withdrawn and the one rule that decides when a place treats a host as retired.
//!
//! A host is retired only when it is in [`RETIRED_HOSTS`] and not in the active list of the place
//! that asks. While `codex` and `omp` stay in the active lists the answer is always `false` and
//! nothing changes; a removal change switches the layer on by deleting the host from the active
//! list of each place.

/// The hosts whose support is withdrawn (the decision of 2026-11-21).
pub const RETIRED_HOSTS: [&str; 2] = ["codex", "omp"];

/// Whether `host` is retired for a place whose active hosts are `active`.
pub fn retired(host: &str, active: &[&str]) -> bool {
    let _ = (host, active);
    false
}

/// Removes the retired hosts from a profile's `hosts`. A list that held retired hosts and is
/// empty afterwards reads as `["claude"]`; a list that was empty stays empty (the profile check
/// still refuses it).
pub fn retire_hosts(hosts: &mut Vec<String>, active: &[&str]) {
    let _ = (hosts, active);
}

/// Whether a manifest entry (`owned_files[].section_id`) records a generated file of a retired
/// host: the three kinds `generation:project-skill:<host>:<id>`,
/// `generation:native-role:<host>:<role>` and `generation:skill-projection:<host>:<id>`.
pub fn owned_entry_retired(section_id: &str, active: &[&str]) -> bool {
    let _ = (section_id, active);
    false
}

#[cfg(test)]
mod tests {
    use super::*;

    const TODAY: [&str; 3] = ["claude", "codex", "omp"];
    const AFTER: [&str; 1] = ["claude"];

    fn v(items: &[&str]) -> Vec<String> {
        items.iter().map(|s| s.to_string()).collect()
    }

    #[test]
    fn the_retired_list_is_codex_and_omp() {
        assert_eq!(RETIRED_HOSTS, ["codex", "omp"]);
    }

    #[test]
    fn a_host_is_retired_only_when_the_place_no_longer_lists_it() {
        // Active today: never retired, whatever the list.
        for host in TODAY {
            assert!(!retired(host, &TODAY), "{host}");
        }
        // After the removal the place lists claude only.
        assert!(retired("codex", &AFTER));
        assert!(retired("omp", &AFTER));
        assert!(!retired("claude", &AFTER));
        // A name that is not in the fixed list is never retired, listed or not.
        assert!(!retired("gemini", &AFTER));
        assert!(!retired("nope", &[]));
        // A place that still lists one of them keeps that one.
        assert!(retired("codex", &["claude", "omp"]));
        assert!(!retired("omp", &["claude", "omp"]));
    }

    #[test]
    fn retire_hosts_removes_retired_names_and_falls_back_to_claude() {
        for (input, expected) in [
            (&["claude", "codex"][..], &["claude"][..]),
            (&["codex", "claude", "omp"][..], &["claude"][..]),
            (&["codex", "omp"][..], &["claude"][..]),
            (&["omp"][..], &["claude"][..]),
            (&["codex"][..], &["claude"][..]),
            (&["claude"][..], &["claude"][..]),
        ] {
            let mut hosts = v(input);
            retire_hosts(&mut hosts, &AFTER);
            assert_eq!(hosts, v(expected), "{input:?}");
        }
    }

    #[test]
    fn retire_hosts_changes_nothing_while_the_hosts_are_active() {
        for input in [
            &["claude", "codex", "omp"][..],
            &["codex"][..],
            &["omp", "claude"][..],
            &[][..],
        ] {
            let mut hosts = v(input);
            retire_hosts(&mut hosts, &TODAY);
            assert_eq!(hosts, v(input), "{input:?}");
        }
    }

    #[test]
    fn retire_hosts_keeps_an_empty_list_empty_and_unknown_names_as_they_are() {
        let mut empty: Vec<String> = Vec::new();
        retire_hosts(&mut empty, &AFTER);
        assert!(empty.is_empty());
        let mut hosts = v(&["gemini", "codex"]);
        retire_hosts(&mut hosts, &AFTER);
        assert_eq!(hosts, v(&["gemini"]));
    }

    #[test]
    fn manifest_entries_of_a_retired_host_are_named_by_their_kind() {
        for kind in ["project-skill", "native-role", "skill-projection"] {
            let codex = format!("generation:{kind}:codex:moai");
            let omp = format!("generation:{kind}:omp:moai");
            let claude = format!("generation:{kind}:claude:moai");
            assert!(owned_entry_retired(&codex, &AFTER), "{codex}");
            assert!(owned_entry_retired(&omp, &AFTER), "{omp}");
            assert!(!owned_entry_retired(&claude, &AFTER), "{claude}");
            assert!(!owned_entry_retired(&codex, &TODAY), "{codex} while active");
        }
    }

    #[test]
    fn manifest_entries_of_other_shapes_are_never_retired() {
        for id in [
            "",
            "generation:other:codex:moai",
            "generation:project-skill:codex",
            "agents:codex:x",
            "codex",
            "generation:project-skill:xcodex:moai",
        ] {
            assert!(!owned_entry_retired(id, &AFTER), "{id}");
        }
    }
}
