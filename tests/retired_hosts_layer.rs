//! SPEC-20261121-retired-hosts-read-safely (REQ-1, REQ-2, REQ-10): the retired-host rule, the
//! profile list filter and the manifest entry check, called with an active list that leaves the
//! retired hosts out (the place a removal change switches the layer on) and with today's list.
use moai_cli::retired_hosts::{owned_entry_retired, retire_hosts, retired, RETIRED_HOSTS};

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
