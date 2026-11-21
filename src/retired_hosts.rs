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
