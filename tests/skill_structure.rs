//! SPEC-20261123-skill-structure: the four roles of the `moai` skill hold. The entry routes, the
//! stage workflows say how to resume, one reference says when a timer may be armed, and the
//! stage order of the SDD reference ends in the landing and the next card.
//!
//! The tests read the named files of this checkout by path; they walk no folder, run nothing
//! and touch no HOME.

use std::{fs, path::Path};

fn read(rel: &str) -> String {
    let path = Path::new(env!("CARGO_MANIFEST_DIR")).join(rel);
    fs::read_to_string(&path).unwrap_or_else(|e| panic!("{}: {e}", path.display()))
}

const LOOP_REF: &str = "skills/moai/references/internal/loop.md";
const SKILL_ENTRY: &str = "skills/moai/SKILL.md";
const SDD_REF: &str = "skills/moai/references/internal/sdd.md";

/// The stage order written out in full, owned by the SDD reference.
const ORDER: &str = "PLAN → RUN → VERIFY → SYNC → AUDIT → merge";

/// REQ-9: the stage workflows and the words that show each one's `## Resume` is about its own
/// work. WORK and UPDATE are left out on purpose (SPEC decision D-4).
const RESUME_WORKFLOWS: [(&str, &str); 6] = [
    ("read.md", "source"),
    ("plan.md", "`moai spec check`"),
    ("run.md", "`moai card show"),
    ("sync.md", "`## Result`"),
    ("audit.md", "head"),
    ("project.md", "`moai project-update-plan"),
];

/// A section's text: from its heading line to the next `## ` heading or the end of the file.
fn section<'a>(text: &'a str, heading: &str) -> &'a str {
    let marker = format!("\n{heading}\n");
    let start = text
        .find(&marker)
        .unwrap_or_else(|| panic!("no `{heading}` section"))
        + 1;
    let rest = &text[start..];
    match rest[heading.len()..].find("\n## ") {
        Some(end) => &rest[..heading.len() + end + 1],
        None => rest,
    }
}

/// REQ-1, REQ-2, REQ-3: the loop reference exists, opens with its "When to read" line, refuses a
/// session that is not the confirmed leader and defines the confirmed leader with the two
/// commands that read the session id and the registered leader entry.
#[test]
fn loop_reference_refuses_an_unconfirmed_session_and_defines_the_leader() {
    let text = read(LOOP_REF);
    let head: Vec<&str> = text.lines().take(3).collect();
    assert!(
        head.len() == 3
            && head[0].starts_with("# Internal: loop")
            && head[1].is_empty()
            && head[2].starts_with("When to read: before "),
        "the loop reference opens with its title and the line that says when to read it: {head:?}"
    );
    for needle in [
        // REQ-2: the refusal, for the three ways to repeat a round.
        "A session that is not the confirmed leader arms no `/loop`, no cron task and no wakeup for the leader round.",
        // REQ-3: the definition and its two commands.
        "its own id, printed by `moai session-current`, equals the `session` of the live leader entry that `moai session-register --show --name leader` prints",
    ] {
        assert!(text.contains(needle), "the loop reference lacks: {needle}");
    }
}

/// REQ-4, REQ-5: a confirmed leader lists its timers, removes the earlier one for the same round
/// and arms exactly one; when the list cannot be read it arms nothing and says so.
#[test]
fn loop_reference_arms_one_timer() {
    let text = read(LOOP_REF);
    for needle in [
        "list its timers with `CronList`",
        "remove an earlier one for the same round with `CronDelete`",
        "arm exactly one",
        "If the timers cannot be listed, arm none and say so in one line.",
    ] {
        assert!(text.contains(needle), "the loop reference lacks: {needle}");
    }
    let list = text.find("list its timers").unwrap();
    let remove = text.find("remove an earlier one").unwrap();
    let arm = text.find("arm exactly one").unwrap();
    assert!(
        list < remove && remove < arm,
        "the order is: list, remove, arm"
    );
}

/// REQ-6, REQ-7, REQ-8: the entry points to the loop reference in one line, holds no sentence of
/// the rule and does not list the steps of the WORK route.
#[test]
fn skill_entry_points_to_the_loop_reference_and_does_not_repeat_it() {
    let entry = read(SKILL_ENTRY);
    assert_eq!(
        entry.matches("`internal/loop`").count(),
        1,
        "the entry names the loop reference once"
    );
    let pointer = entry
        .lines()
        .find(|l| l.contains("`internal/loop`"))
        .unwrap();
    assert!(
        pointer.starts_with(
            "Before a session arms `/loop`, a cron task or a wakeup for the leader round, read "
        ),
        "the pointer line says when to read it: {pointer}"
    );
    for word in [
        "session-register",
        "session-current",
        "CronList",
        "CronDelete",
        "confirmed leader",
    ] {
        assert!(
            !entry.contains(word),
            "the entry holds the loop rule's {word}"
        );
    }
}

#[test]
fn skill_entry_names_the_work_route_without_its_steps() {
    let entry = read(SKILL_ENTRY);
    assert!(
        entry.contains("`/moai work` picks WORK"),
        "the entry still names the WORK route"
    );
    for steps in [
        "take one card, plan it, run it, sync it, audit it",
        "land it with the landing command, and repeat until no card is left",
    ] {
        assert!(
            !entry.contains(steps),
            "the entry lists the WORK steps, which `workflows/work.md` owns: {steps}"
        );
    }
}

/// REQ-9: each stage workflow has a `## Resume` section directly before
/// `## Exceptions and Permissions`; it names what the session reads afresh, is about the
/// workflow's own work and stays short.
#[test]
fn resume_sits_before_the_exceptions_in_every_stage_workflow() {
    for (file, token) in RESUME_WORKFLOWS {
        let text = read(&format!("skills/moai/workflows/{file}"));
        let resume = section(&text, "## Resume");
        let after = &text[text.find(resume).unwrap() + resume.len()..];
        assert!(
            after.starts_with("## Exceptions and Permissions\n"),
            "{file}: `## Resume` is not directly before `## Exceptions and Permissions`"
        );
        assert!(
            resume.contains("afresh"),
            "{file}: the Resume section does not say what is read afresh"
        );
        assert!(
            resume.contains(token),
            "{file}: the Resume section does not name {token}"
        );
        assert!(
            resume.len() <= 700,
            "{file}: the Resume section is {} bytes, over 700",
            resume.len()
        );
    }
}

/// REQ-9, D-4: the two workflows another change holds keep their own text for now, so this
/// change does not add a Resume section to them.
#[test]
fn resume_is_not_added_to_the_workflows_another_change_holds() {
    for file in ["work.md", "update.md"] {
        let text = read(&format!("skills/moai/workflows/{file}"));
        assert!(
            !text.contains("\n## Resume\n"),
            "{file} gets its Resume section in the change that holds the file"
        );
    }
}

/// REQ-10: after the stage order the SDD reference names the landing, its closing of the card in
/// the same call and the next-card command, in that order; the full order stays in this file.
#[test]
fn sdd_order_ends_in_the_landing_and_the_next_card() {
    let sdd = read(SDD_REF);
    let order = sdd.find(ORDER).expect("the SDD reference writes the order");
    let rest = &sdd[order + ORDER.len()..];
    let land = rest.find("`moai card land`").expect("the landing command");
    let closes = rest
        .find("closes the card in the same call")
        .expect("the landing closes the card");
    let next = rest
        .find("`moai card next`")
        .expect("the next-card command");
    assert!(
        land < closes && closes < next,
        "the order is: land, close in the same call, next card"
    );
    assert_eq!(sdd.matches(ORDER).count(), 1, "the order is written once");
}
