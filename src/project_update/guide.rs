//! What the project update plan says to do next with each step
//! (SPEC-20261125-update-next-say-ask-chain): the one word `next` of a step, the plain sentence
//! `say`, and the one question `ask` with the order `chain` and the end `finish` that the plan
//! carries on top. Every function here is pure: it reads the facts the plan hands over and no
//! file, so the plan stays a read-only preview.

use serde_json::{json, Value};

/// The steps whose entries read `guarded`: they change the project's settings, the Go tool state
/// or the user's reports, so the question names them.
pub const GUARDED: [&str; 6] = [
    "go-settings",
    "go-state",
    "reports",
    "claude-settings",
    "hooks",
    "claude-settings-merge",
];

/// Whether the step `id` is one the question names as guarded.
pub fn guarded(id: &str) -> bool {
    GUARDED.contains(&id)
}

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
    match state {
        "unchanged" | "skip" | "optional" => "none",
        "blocked" if waits_behind => "wait",
        "ready" | "needs_decision" | "conflict"
            if has_choices || (has_command && needs_confirmation) =>
        {
            "ask"
        }
        _ if has_command && !needs_confirmation && !has_choices => "apply",
        _ => "report",
    }
}

/// One line of words per step and situation. `topic` starts the sentences about waiting and
/// declining; `change` says what running the step does; `choose` says what the user picks when the
/// step asks; `current` says why nothing is left to do; `held` says why the step cannot run here.
struct Row {
    id: &'static str,
    topic: &'static str,
    change: &'static str,
    choose: Option<&'static str>,
    current: &'static str,
    held: &'static str,
}

/// The sentence table. A step without a row reads the generic sentences of [`say`].
const ROWS: [Row; 12] = [
    Row {
        id: "go-settings",
        topic: "Cleaning the old Go tool values out of your Claude Code settings",
        change: "Remove the old Go tool values from your Claude Code settings and keep a backup of the file.",
        choose: None,
        current: "Your Claude Code settings hold no old Go tool values.",
        held: "The old Go tool values in your Claude Code settings cannot be removed here and need a look first.",
    },
    Row {
        id: "go-state",
        topic: "Moving the old Go tool files into a backup",
        change: "Move the old Go tool files and state into a backup folder so the new layout can be installed.",
        choose: None,
        current: "There are no old Go tool files or state left to move.",
        held: "The old Go tool files cannot be moved here and need a look first.",
    },
    Row {
        id: "reports",
        topic: "Moving your earlier report files",
        change: "Move your earlier report files out of the project into your home folder.",
        choose: None,
        current: "No earlier report files need to move.",
        held: "Your earlier report files cannot be moved here and need a look first.",
    },
    Row {
        id: "profile",
        topic: "Writing the project description file",
        change: "Write the project description file that records what this project uses.",
        choose: None,
        current: "The project description file is already up to date.",
        held: "The project description file cannot be written here and needs a look first.",
    },
    Row {
        id: "generation",
        topic: "Writing the MoAI instruction files",
        change: "Write or refresh the MoAI instruction and workflow files in this project.",
        choose: Some("Some MoAI instruction files were edited by hand, so you choose how to keep them."),
        current: "The MoAI instruction and workflow files are already up to date.",
        held: "The MoAI instruction and workflow files cannot be written here and need a look first.",
    },
    Row {
        id: "claude-settings",
        topic: "Updating the Claude Code settings file",
        change: "Update the Claude Code settings file with the MoAI permissions and options.",
        choose: None,
        current: "The Claude Code settings file is already up to date.",
        held: "The Claude Code settings file cannot be updated here and needs a look first.",
    },
    Row {
        id: "hooks",
        topic: "Installing the Claude Code automatic checks",
        change: "Install the Claude Code automatic checks that MoAI relies on.",
        choose: None,
        current: "The Claude Code automatic checks are already installed.",
        held: "The Claude Code automatic checks cannot be installed here and need a look first.",
    },
    Row {
        id: "github",
        topic: "Adding the issue and pull request templates",
        change: "Add or refresh the issue and pull request templates in this project.",
        choose: None,
        current: "The issue and pull request templates are already up to date.",
        held: "The issue and pull request templates cannot be written here and need a look first.",
    },
    Row {
        id: "agents-picks",
        topic: "Moving your AGENTS.md text into the CLAUDE files",
        change: "Move the text of your AGENTS.md and AGENTS.local.md files into the CLAUDE files.",
        choose: Some("You choose whether the text of your AGENTS.md files moves into the CLAUDE files or stays where it is."),
        current: "Your AGENTS.md text needs no further move.",
        held: "Your AGENTS.md files cannot be moved here and need a look first.",
    },
    Row {
        id: "agents-local-merge",
        topic: "Copying your AGENTS.local.md text into CLAUDE.local.md",
        change: "Copy the text of your AGENTS.local.md file into CLAUDE.local.md.",
        choose: None,
        current: "Your AGENTS.local.md text needs no copying.",
        held: "Your AGENTS.local.md text cannot be copied here and needs a look first.",
    },
    Row {
        id: "claude-settings-merge",
        topic: "Merging your earlier settings back in",
        change: "Merge the settings you had before back into the new Claude Code settings file.",
        choose: Some("Some earlier settings differ from the new file, so you choose which ones to keep."),
        current: "No earlier settings are left to merge back.",
        held: "Your earlier settings cannot be merged back here and need a look first.",
    },
    Row {
        id: "retired-hosts",
        topic: "Checking the files of retired hosts",
        change: "MoAI no longer manages the files of retired hosts, so remove them yourself if you do not use them.",
        choose: None,
        current: "MoAI no longer manages the files of retired hosts, so remove them yourself if you do not use them.",
        held: "MoAI no longer manages the files of retired hosts, so remove them yourself if you do not use them.",
    },
];

/// The install steps that wait for the project description before they can say anything else.
const NEEDS_SETUP: [&str; 5] = [
    "profile",
    "generation",
    "claude-settings",
    "hooks",
    "github",
];

/// The `say` of a step: one plain line of words, chosen by what the step prints. It never echoes
/// the `reason` of a step, which can carry an error text.
///
/// - `step`: the step id; `state` and `next` as the plan prints them.
/// - `declined`: the user declined the step and it stays as it is.
pub fn say(step: &str, state: &str, next: &str, declined: bool) -> String {
    let row = ROWS.iter().find(|r| r.id == step);
    let topic = row.map_or("This step", |r| r.topic);
    if declined {
        return format!(
            "{topic} was declined earlier and stays that way until its source changes."
        );
    }
    match next {
        "wait" => format!("{topic} comes after earlier steps and is checked again once they are done."),
        "apply" => row.map_or("Run this step as printed.", |r| r.change).to_string(),
        "ask" => row
            .map(|r| r.choose.unwrap_or(r.change))
            .unwrap_or("This step needs your choice before it runs.")
            .to_string(),
        "none" if state == "optional" && step == "agents-picks" => {
            "The text is already in the CLAUDE files, and you may delete the original AGENTS files if you wish.".into()
        }
        "none" => row
            .map_or("This step needs nothing.", |r| r.current)
            .to_string(),
        _ if state == "needs_decision" && NEEDS_SETUP.contains(&step) => {
            format!("{topic} needs the project setup to be run first.")
        }
        _ => row
            .map_or("This step needs a look first.", |r| r.held)
            .to_string(),
    }
}

/// What the overview needs to know about one step, in the order the plan prints the steps.
pub struct Fact {
    pub step: &'static str,
    /// The `next` of the step (see [`next_of`]).
    pub next: &'static str,
    /// The `say` of the step.
    pub say: String,
    /// For a step that asks: the id of its recommended choice when that choice needs no
    /// confirmation word.
    pub pick: Option<String>,
    /// The steps this step waits behind, runnable or not; only the runnable ones are named.
    pub behind: &'static [&'static str],
}

impl Fact {
    /// Whether the one question can run the step: it applies by itself, or it asks and has a
    /// recommended choice that needs no confirmation word.
    fn runnable(&self) -> bool {
        self.next == "apply" || (self.next == "ask" && self.pick.is_some())
    }
}

fn question(has_steps: bool) -> &'static str {
    if has_steps {
        "Tell me whether to apply the changes listed here to your project now."
    } else {
        "Tell me what to do with each step that needs your choice."
    }
}

fn options(has_steps: bool) -> Value {
    if has_steps {
        json!([
            {"id": "apply", "recommended": true,
             "say": "Apply every change listed, then check again for the steps that come after."},
            {"id": "choose", "recommended": false, "say": "Go through the steps one by one."},
            {"id": "leave", "recommended": false, "say": "Leave the project as it is."},
        ])
    } else {
        json!([
            {"id": "choose", "recommended": true,
             "say": "Go through the steps that need your choice one by one."},
            {"id": "leave", "recommended": false, "say": "Leave the project as it is."},
        ])
    }
}

/// A leftover of an earlier generation that waits for the user's answer
/// (SPEC-20261127-update-leftovers-action): it is not a step and never makes the plan ask, but
/// the end of the update lists it.
pub struct OpenLeftover<'a> {
    /// The path as the plan prints it.
    pub path: &'a str,
    /// Whether the plan offers to move it into a backup; otherwise it can only be kept.
    pub archive: bool,
}

/// The sentence for one open leftover.
fn leftover_say(l: &OpenLeftover) -> String {
    if l.archive {
        format!(
            "{} was left by an earlier version and is still open, so choose to move it into a backup or to keep it where it is.",
            l.path
        )
    } else {
        format!(
            "{} was left by an earlier version and cannot be moved here, so choose to keep it where it is.",
            l.path
        )
    }
}

/// The three keys the plan adds on top: `ask`, `chain` and `finish`.
///
/// - `ask` is one object while a step has `next` `apply` or `ask`; otherwise null.
/// - `chain` holds the `apply_order` (the ids of the steps of `ask`) and the places to preview
///   again; with no `ask` both lists are empty.
/// - `finish` is an object while `ask` is null, and null otherwise. Its `open` lists the steps
///   that still need attention and then the open `leftovers` (`{"leftover": path, "say": …}`);
///   the finish never says that nothing is left while one is listed.
pub fn overview(
    facts: &[Fact],
    all_unchanged: bool,
    leftovers: &[OpenLeftover],
) -> (Value, Value, Value) {
    let asks = facts.iter().any(|f| matches!(f.next, "apply" | "ask"));
    if !asks {
        let mut open: Vec<Value> = facts
            .iter()
            .filter(|f| matches!(f.next, "report" | "wait"))
            .map(|f| json!({"step": f.step, "next": f.next, "say": f.say}))
            .collect();
        let steps_open = !open.is_empty();
        open.extend(
            leftovers
                .iter()
                .map(|l| json!({"leftover": l.path, "say": leftover_say(l)})),
        );
        let say = match (leftovers.is_empty(), steps_open) {
            (true, _) if all_unchanged => {
                "Nothing is left to apply, so the update ends without a question."
            }
            (true, _) => {
                "No step can run on its own, so the update ends here without a question; the steps listed still need attention."
            }
            (false, true) => {
                "No step can run on its own, so the update ends here without a question; the steps and leftovers listed still need attention."
            }
            (false, false) => {
                "No step can run on its own, so the update ends here without a question; the leftovers listed are still open and need your answer."
            }
        };
        let chain = json!({"apply_order": [], "preview_again": []});
        return (Value::Null, chain, json!({"say": say, "open": open}));
    }
    let runnable: Vec<&Fact> = facts.iter().filter(|f| f.runnable()).collect();
    let order: Vec<&str> = runnable.iter().map(|f| f.step).collect();
    let steps: Vec<Value> = runnable
        .iter()
        .map(|f| json!({"step": f.step, "say": f.say, "guarded": guarded(f.step), "pick": f.pick}))
        .collect();
    // The runnable steps a waiting step is behind, in the order they run.
    let after = |f: &Fact| -> Vec<&str> {
        order
            .iter()
            .copied()
            .filter(|id| f.behind.contains(id))
            .collect()
    };
    let follows: Vec<(&Fact, Vec<&str>)> = facts
        .iter()
        .filter(|f| f.next == "wait")
        .map(|f| (f, after(f)))
        .collect();
    let follows_json: Vec<Value> = follows
        .iter()
        .map(|(f, after)| {
            json!({"step": f.step, "say": f.say, "guarded": guarded(f.step), "after": after})
        })
        .collect();
    let picks: Vec<Value> = facts
        .iter()
        .filter(|f| f.next == "ask" && !f.runnable())
        .map(|f| json!({"step": f.step, "say": f.say, "guarded": guarded(f.step)}))
        .collect();
    let preview_again: Vec<Value> = order
        .iter()
        .filter_map(|id| {
            let then: Vec<&str> = follows
                .iter()
                .filter(|(_, after)| after.contains(id))
                .map(|(f, _)| f.step)
                .collect();
            (!then.is_empty()).then(|| json!({"after": id, "then": then}))
        })
        .collect();
    let ask = json!({
        "question": question(!order.is_empty()),
        "options": options(!order.is_empty()),
        "steps": steps,
        "follows": follows_json,
        "picks": picks,
    });
    let chain = json!({"apply_order": order, "preview_again": preview_again});
    (ask, chain, Value::Null)
}

#[cfg(test)]
mod tests {
    use super::*;

    /// Whole words a sentence for a person never holds: step ids and state words.
    const FORBIDDEN_WORDS: [&str; 19] = [
        "go-settings",
        "go-state",
        "reports",
        "profile",
        "generation",
        "claude-settings",
        "hooks",
        "github",
        "agents-picks",
        "agents-local-merge",
        "claude-settings-merge",
        "retired-hosts",
        "unchanged",
        "skip",
        "optional",
        "ready",
        "needs_decision",
        "conflict",
        "blocked",
    ];
    /// Text a sentence never holds anywhere.
    const FORBIDDEN_TEXT: [&str; 5] = ["digest", "argv", "spec-", "<", "`"];

    fn plain(text: &str) {
        assert!(!text.is_empty() && !text.contains('\n'), "{text:?}");
        assert!(text.chars().count() <= 200, "{text:?}");
        assert!(text.ends_with('.'), "{text:?}");
        let lower = text.to_lowercase();
        let words: Vec<&str> = lower
            .split(|c: char| !(c.is_ascii_alphanumeric() || c == '_' || c == '-'))
            .collect();
        for f in FORBIDDEN_WORDS {
            assert!(!words.contains(&f), "{text:?} holds {f}");
        }
        for f in FORBIDDEN_TEXT {
            assert!(!lower.contains(f), "{text:?} holds {f}");
        }
    }

    /// Every sentence of the table, in every situation, is one plain line.
    #[test]
    fn every_sentence_is_one_plain_line() {
        let ids = ROWS.iter().map(|r| r.id).chain(["unknown-step"]);
        for id in ids {
            for (state, next) in [
                ("ready", "apply"),
                ("conflict", "ask"),
                ("needs_decision", "ask"),
                ("blocked", "wait"),
                ("blocked", "report"),
                ("needs_decision", "report"),
                ("unchanged", "none"),
                ("optional", "none"),
            ] {
                plain(&say(id, state, next, false));
            }
            plain(&say(id, "skip", "none", true));
        }
        plain(question(true));
        plain(question(false));
        for has_steps in [true, false] {
            for o in options(has_steps).as_array().unwrap() {
                plain(o["say"].as_str().unwrap());
            }
        }
    }

    /// An open leftover is listed in `finish.open` with a plain sentence and never makes the plan
    /// ask; the finish does not say that nothing is left while one is open, and keeps saying it
    /// when none is.
    #[test]
    fn an_open_leftover_is_listed_and_is_not_settled() {
        let open = [
            OpenLeftover {
                path: crate::fixed_state!(""),
                archive: true,
            },
            OpenLeftover {
                path: ".moai/.update.lock",
                archive: false,
            },
        ];
        for l in &open {
            plain(&leftover_say(l));
        }
        let step_open = [Fact {
            step: "go-state",
            next: "report",
            say: "The old Go tool files cannot be moved here and need a look first.".into(),
            pick: None,
            behind: &[],
        }];
        for facts in [&[][..], &step_open[..]] {
            let (ask, _, finish) = overview(facts, false, &open);
            assert!(ask.is_null());
            let listed = finish["open"].as_array().unwrap();
            let leftovers: Vec<&Value> = listed
                .iter()
                .filter(|e| e.get("leftover").is_some())
                .collect();
            assert_eq!(leftovers.len(), 2, "{finish}");
            assert_eq!(leftovers[0]["leftover"], crate::fixed_state!(""));
            assert_ne!(leftovers[0]["say"], leftovers[1]["say"]);
            let say = finish["say"].as_str().unwrap();
            plain(say);
            assert!(!say.to_lowercase().contains("nothing is left"), "{say}");
            assert!(say.contains("leftovers"), "{say}");
        }
        let (_, _, done) = overview(&[], true, &[]);
        assert!(done["say"].as_str().unwrap().contains("Nothing is left"));
        assert_eq!(done["open"], json!([]));
        // Without a leftover the sentence of a step that needs attention is the one it always was.
        let (_, _, held) = overview(&step_open, false, &[]);
        assert!(held["say"]
            .as_str()
            .unwrap()
            .ends_with("the steps listed still need attention."));
    }
}
