//! What the project update plan says to do next with each step: the one word `next` of a step
//! (SPEC-20261125-update-next-say-ask-chain). A pure function of what the step prints; it reads
//! no file. RED skeleton: the real rules come with the next step of the SPEC.

/// The `next` of a step: one of `apply`, `ask`, `wait`, `report` and `none`.
///
/// - `state`: the state word of the step (`unchanged`, `skip`, `optional`, `ready`,
///   `needs_decision`, `conflict`, `blocked`).
/// - `has_command`: the step prints a command.
/// - `needs_confirmation`: that command carries `<warning-confirmation>` or `--confirm-replace`.
/// - `has_choices`: the step prints choices.
/// - `waits_behind`: the step waits behind other steps.
pub fn next_of(
    state: &str,
    has_command: bool,
    needs_confirmation: bool,
    has_choices: bool,
    waits_behind: bool,
) -> &'static str {
    let _ = (
        state,
        has_command,
        needs_confirmation,
        has_choices,
        waits_behind,
    );
    "report"
}
