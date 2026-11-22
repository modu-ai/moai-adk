//! SPEC-20261122-next-flow-skills-agents, text slice A (AC-1, AC-2, AC-5): the `/moai` skill
//! package has eight routes, `/moai work` takes a card through the whole flow, and the retired
//! SETUP, VERIFY and OPTIMIZE routes leave their rules in PROJECT, RUN and UPDATE. Text slice B
//! (AC-3): one flow reference and one audit contract, each within its byte budget. Text slice C
//! (AC-4): the shipped texts name no retired host, mode, gate or tool.
//!
//! The tests read the skill files of this checkout; nothing is run and no HOME is touched. The
//! planted-copy tests write only into a temporary folder.

use serde_json::Value;
use std::{
    fs,
    path::{Path, PathBuf},
};

fn package() -> PathBuf {
    Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai")
}

fn read(rel: &str) -> String {
    fs::read_to_string(package().join(rel)).unwrap_or_else(|e| panic!("{rel}: {e}"))
}

fn routes() -> Value {
    serde_json::from_str(&read("references/routes.json")).expect("routes.json is JSON")
}

fn route_keys() -> Vec<String> {
    let mut keys: Vec<String> = routes()["routes"]
        .as_object()
        .expect("routes.json has a `routes` object")
        .keys()
        .cloned()
        .collect();
    keys.sort();
    keys
}

/// REQ-1: the eight routes, sorted.
const ROUTES: [&str; 8] = [
    "AUDIT", "PLAN", "PROJECT", "READ", "RUN", "SYNC", "UPDATE", "WORK",
];

/// REQ-2, the files text slices A, B and C remove. Slice B adds `references/internal/review.md`
/// (absorbed into the audit contract), slice C `references/internal/build-folder.md` (its rule,
/// rewritten, is part of RUN's "Check-run rules"); only the rule history is still to move: Plan
/// step 6 adds `references/rule-history.md` when it moves to `docs/`.
const REMOVED: [&str; 8] = [
    "workflows/mode-agent.md",
    "workflows/mode-leader.md",
    "workflows/setup.md",
    "workflows/verify.md",
    "workflows/optimize.md",
    "references/adapters/herdr.md",
    "references/internal/review.md",
    "references/internal/build-folder.md",
];

/// D-14: the build rule RUN's "Check-run rules" holds now that the build-folder reference left
/// the package: worktrees under `.claude/worktrees`, each building into its own folder, fixed
/// slots warmed by a clone of the seed's build folder, and removal when the work lands.
const BUILD_RULE: [&str; 8] = [
    "`.claude/worktrees/<name>`",
    "the build tool's default folder inside that worktree",
    "for cargo, the worktree's own `target/`",
    "leave `CARGO_TARGET_DIR` unset or point it at that folder",
    "`cp -c -R -p seed/target <slot>/target`",
    "are removed when its branch lands or is deleted",
    "an audit worktree when its verdict is posted",
    "by relative path, never through an absolute or home path",
];

/// REQ-5 and REQ-6: the byte budget of the flow reference and of the audit contract.
const REF_LIMIT: usize = 6144;

/// REQ-5: the flow reference and the meaning it alone explains.
const FLOW_REF: &str = "references/flows/next-flow.md";
const MEANING: &str = "New · engineering · eXperience · Together";
const FLOW_LINE: &str = "Work in parallel. Land once.";

/// REQ-5: the four technical rules, each with the letter of the name it must not be spelled as.
const RULES: [(char, &str); 4] = [
    ('n', "no tip chase"),
    ('e', "evidence before landing"),
    ('x', "exactly-once landing"),
    ('t', "a guarded tip"),
];

/// REQ-5: the condition the push sentence names.
const PUSH_CONDITION: &str = "only when the user has turned the remote copy on";

/// REQ-6: the audit contract.
const AUDIT_REF: &str = "references/audit.md";

/// REQ-4: the mode spellings SKILL.md no longer names.
const MODE_SPELLINGS: [&str; 5] = [
    "--mode-agent",
    "--mode-leader",
    "--autonomy",
    "MODE-AGENT",
    "MODE-LEADER",
];

/// REQ-4: the one sentence of SKILL.md that makes a dash option plain request text. It names
/// no retired flag (REQ-8).
const DASH_SENTENCE: &str =
    "A word of the request that is not a route, a dash option included, is request text and picks no route.";

/// REQ-3: the sentence of SKILL.md that picks WORK by the leading word.
const WORK_SENTENCE: &str = "A request whose first word after the skill name is `work` picks WORK";

/// REQ-3: the WORK procedure's step keys, in order.
const WORK_STEPS: [&str; 7] = ["take", "plan", "run", "sync", "audit", "land", "repeat"];

/// The mode spellings `text` holds, in the order of `MODE_SPELLINGS`.
fn mode_spellings(text: &str) -> Vec<&'static str> {
    MODE_SPELLINGS
        .iter()
        .copied()
        .filter(|s| text.contains(s))
        .collect()
}

/// The body of the `## <name>` section of `text`, up to the next `## ` heading.
fn section<'a>(text: &'a str, name: &str) -> &'a str {
    let head = format!("\n## {name}\n");
    let at = text
        .find(&head)
        .unwrap_or_else(|| panic!("no `## {name}` section"));
    let body = &text[at + head.len()..];
    &body[..body.find("\n## ").unwrap_or(body.len())]
}

/// The numbered steps of a procedure written as `N. **key** — …`: (number, key) per step.
fn step_keys(procedure: &str) -> Vec<(usize, String)> {
    procedure
        .lines()
        .filter_map(|line| {
            let (number, rest) = line.split_once(". **")?;
            let number: usize = number.parse().ok()?;
            let (key, after) = rest.split_once("** — ")?;
            (!after.is_empty()).then(|| (number, key.to_string()))
        })
        .collect()
}

/// Every shipped text: the files under `skills/moai/` and `agents/`, as (repository path, text).
fn shipped_texts() -> Vec<(String, String)> {
    shipped_texts_under(Path::new(env!("CARGO_MANIFEST_DIR")))
}

/// The shipped texts of the tree at `root` (this checkout, or a copy of it in a test).
fn shipped_texts_under(root: &Path) -> Vec<(String, String)> {
    fn walk(dir: &Path, out: &mut Vec<PathBuf>) {
        for entry in fs::read_dir(dir).unwrap_or_else(|e| panic!("{}: {e}", dir.display())) {
            let path = entry.unwrap().path();
            if path.is_dir() {
                walk(&path, out);
            } else {
                out.push(path);
            }
        }
    }
    let mut files = Vec::new();
    for top in ["skills/moai", "agents"] {
        walk(&root.join(top), &mut files);
    }
    files.sort();
    files
        .into_iter()
        .map(|p| {
            let rel = p.strip_prefix(root).unwrap().to_string_lossy().into_owned();
            let text = fs::read_to_string(&p).unwrap_or_else(|e| panic!("{rel}: {e}"));
            (rel, text)
        })
        .collect()
}

/// The (letter, rule) pairs `text` spells as the letters of the name: a rule with its letter set
/// before it (`N = no tip chase`, `- **T**: a guarded tip`) or marked inside it
/// (`e**X**actly-once landing`).
fn letter_spellings(text: &str) -> Vec<(char, &'static str)> {
    let separators = |c: char| c.is_whitespace() || "=:—–-*`_()→,.".contains(c);
    let mut hits = Vec::new();
    for line in text.lines() {
        let lower = line.to_lowercase();
        let bare: String = lower
            .chars()
            .filter(|c| !matches!(c, '*' | '`' | '_'))
            .collect();
        for (letter, rule) in RULES {
            // The rule without its article, so `T = guarded tip` is found as well.
            let core = rule.strip_prefix("a ").unwrap_or(rule);
            let before = lower.match_indices(core).any(|(at, _)| {
                let mut head = lower[..at].trim_end_matches(separators);
                if let Some(rest) = head.strip_suffix(" a") {
                    head = rest.trim_end_matches(separators);
                }
                let mut tail = head.chars().rev();
                tail.next() == Some(letter) && !tail.next().is_some_and(char::is_alphanumeric)
            });
            let inside = bare.contains(core) && !lower.contains(core);
            if before || inside {
                hits.push((letter, rule));
            }
        }
    }
    hits
}

/// AC-4: the shipped texts the scrub scans leave out, each with its reason.
const SCRUB_EXCEPTIONS: [&str; 3] = [
    // The rule history records where each rule came from; it moves to `docs/` in Plan step 6.
    "skills/moai/references/rule-history.md",
    // Another card holds the interview question bank until it lands; its lines stay as they are.
    "skills/moai/references/internal/interview.md",
    // The manifest's origin reasons record where each file came from: history, not instructions.
    "skills/moai/references/source-manifest.json",
];

/// REQ-8: words no shipped text names, matched without regard to case: the retired hosts
/// (Codex in any spelling, OMP's skill command), the resident leader and its mode flags, the
/// pull-request audit gate, the merge train and the repository's own harness agents.
const REQ8_WORDS: [&str; 15] = [
    "codex",
    "$moai",
    "/skill:moai",
    "resident leader",
    "leader mode",
    "mode-agent",
    "mode-leader",
    "--autonomy",
    "/moai -a ",
    "/moai -l ",
    "merge train",
    "merge-train",
    "audit gate",
    "prj-reviewer",
    "prj-local-check",
];

/// REQ-8: OMP as a word of its own, without regard to case.
const REQ8_WHOLE_WORDS: [&str; 1] = ["omp"];

/// REQ-8: the older instruction files. A line may name them only as the older files moai moves
/// away from, which it says with the word in `OLDER`; CLAUDE.md is the instruction file.
const AGENTS_FILES: [&str; 2] = ["AGENTS.md", "AGENTS.local.md"];
const OLDER: &str = "older";

/// REQ-9: the tools and options Claude Code no longer has, matched as written.
const REQ9_WORDS: [&str; 7] = [
    "TodoWrite",
    "TaskOutput",
    "BashOutput",
    "agents wizard",
    "`/agents`",
    "`resume`",
    "`mode`",
];

/// REQ-9: the Grep and Glob tools as words of their own; a shipped text never assumes them.
const REQ9_WHOLE_WORDS: [&str; 2] = ["Grep", "Glob"];

/// D-10: TaskCreate is named only as a tool used when the host offers it.
const TASK_CREATE: &str = "TaskCreate";
const OFFERED: &str = "offer";

/// Whether `word` stands in `line` with no letter, digit or `_` on either side.
fn has_whole_word(line: &str, word: &str) -> bool {
    let edge = |c: Option<char>| !c.is_some_and(|c| c.is_alphanumeric() || c == '_');
    line.match_indices(word).any(|(at, w)| {
        edge(line[..at].chars().next_back()) && edge(line[at + w.len()..].chars().next())
    })
}

/// The REQ-8 hits of one text: (line number, word).
fn req8_hits(text: &str) -> Vec<(usize, &'static str)> {
    let mut hits = Vec::new();
    for (number, line) in text.lines().enumerate() {
        let lower = line.to_lowercase();
        for word in REQ8_WORDS {
            if lower.contains(word) {
                hits.push((number + 1, word));
            }
        }
        for word in REQ8_WHOLE_WORDS {
            if has_whole_word(&lower, word) {
                hits.push((number + 1, word));
            }
        }
        for file in AGENTS_FILES {
            if has_whole_word(line, file) && !has_whole_word(&lower, OLDER) {
                hits.push((number + 1, file));
            }
        }
    }
    hits
}

/// The REQ-9 hits of one text at repository path `path`: (line number, what was found).
fn req9_hits(path: &str, text: &str) -> Vec<(usize, &'static str)> {
    let mut hits = Vec::new();
    let mut in_frontmatter = path.starts_with("skills/") && text.starts_with("---\n");
    for (number, line) in text.lines().enumerate() {
        if number > 0 && line == "---" {
            in_frontmatter = false;
        }
        for word in REQ9_WORDS {
            if line.contains(word) {
                hits.push((number + 1, word));
            }
        }
        for word in REQ9_WHOLE_WORDS {
            if has_whole_word(line, word) {
                hits.push((number + 1, word));
            }
        }
        if line.trim_start().starts_with("!`") {
            hits.push((number + 1, "shell-injection block"));
        }
        if in_frontmatter && number > 0 && line.starts_with("model:") {
            hits.push((number + 1, "model key"));
        }
        if line.contains(TASK_CREATE) && !line.contains(OFFERED) {
            hits.push((number + 1, TASK_CREATE));
        }
    }
    hits
}

/// `path:line: word` for each hit of `scan` in the shipped texts under `root`, the exceptions
/// left out.
fn scrub_scan(root: &Path, scan: fn(&str, &str) -> Vec<(usize, &'static str)>) -> Vec<String> {
    shipped_texts_under(root)
        .into_iter()
        .filter(|(path, _)| !SCRUB_EXCEPTIONS.contains(&path.as_str()))
        .flat_map(|(path, text)| {
            scan(&path, &text)
                .into_iter()
                .map(move |(line, word)| format!("{path}:{line}: {word}"))
        })
        .collect()
}

/// A copy of this checkout's shipped texts in a fresh temporary folder.
fn shipped_copy() -> tempfile::TempDir {
    let copy = tempfile::tempdir().unwrap();
    for (path, text) in shipped_texts() {
        let to = copy.path().join(&path);
        fs::create_dir_all(to.parent().unwrap()).unwrap();
        fs::write(&to, text).unwrap();
    }
    copy
}

/// The route cells of the chain table, one list per row.
fn chain_rows(text: &str) -> Vec<Vec<String>> {
    text.lines()
        .filter(|l| l.starts_with("| ") && !l.starts_with("| Request") && !l.starts_with("| ---"))
        .map(|l| {
            let cells: Vec<&str> = l.trim_matches('|').split('|').map(str::trim).collect();
            assert_eq!(cells.len(), 4, "a chain row has four cells: {l}");
            cells[1].split(", ").map(str::to_string).collect()
        })
        .collect()
}

#[test]
fn routes_table_names_exactly_the_eight_routes() {
    assert_eq!(route_keys(), ROUTES, "the route table");
}

#[test]
fn routes_every_reference_is_a_shipped_file() {
    let routes = routes();
    for key in route_keys() {
        let want = format!("../workflows/{}.md", key.to_lowercase());
        assert_eq!(
            routes["routes"][&key]["reference"],
            Value::String(want.clone()),
            "{key}"
        );
        let rel = want.strip_prefix("../").unwrap();
        assert!(package().join(rel).is_file(), "{key}: {rel} is not shipped");
    }
}

#[test]
fn routes_removed_files_are_gone() {
    let left: Vec<&str> = REMOVED
        .iter()
        .copied()
        .filter(|rel| package().join(rel).exists())
        .collect();
    assert!(left.is_empty(), "still in the package: {left:?}");
}

#[test]
fn routes_skill_entry_names_no_mode_flag() {
    // The checker finds a planted spelling, so an empty result below means none is there.
    assert_eq!(
        mode_spellings("pick `/moai --mode-agent` or MODE-LEADER"),
        ["--mode-agent", "MODE-LEADER"]
    );
    assert_eq!(
        mode_spellings(&read("SKILL.md")),
        Vec::<&str>::new(),
        "SKILL.md"
    );
    let modes: Vec<String> = route_keys()
        .into_iter()
        .filter(|k| k.starts_with("MODE-"))
        .collect();
    assert!(modes.is_empty(), "mode routes: {modes:?}");
}

#[test]
fn routes_dash_options_are_plain_text() {
    assert_eq!(
        read("SKILL.md").matches(DASH_SENTENCE).count(),
        1,
        "SKILL.md holds the dash-option sentence once"
    );
}

#[test]
fn routes_chain_table_uses_only_table_routes() {
    let rows = chain_rows(&read("references/chains/index.md"));
    assert!(!rows.is_empty(), "the chain table has rows");
    for row in &rows {
        for route in row {
            assert!(
                ROUTES.contains(&route.as_str()),
                "the chain table names {route}"
            );
        }
    }
    assert!(
        rows.iter().any(|r| r.iter().any(|route| route == "WORK")),
        "the chain table has a WORK row"
    );
}

#[test]
fn work_steps_run_in_order() {
    let work = read("workflows/work.md");
    let steps = step_keys(section(&work, "Procedure"));
    let numbers: Vec<usize> = steps.iter().map(|(n, _)| *n).collect();
    let keys: Vec<&str> = steps.iter().map(|(_, k)| k.as_str()).collect();
    assert_eq!(keys, WORK_STEPS, "the WORK steps");
    assert_eq!(numbers, (1..=WORK_STEPS.len()).collect::<Vec<_>>());
}

#[test]
fn work_takes_with_card_next_and_lands_with_card_land() {
    let work = read("workflows/work.md");
    for command in [
        "`moai card next`",
        "`moai card take",
        "`moai card check <card> --generation <g> --path <p>`",
        "`moai card land <card> --head <40 hex>`",
    ] {
        assert!(work.contains(command), "work.md lacks {command}");
    }
}

#[test]
fn work_repeats_until_no_card_is_left() {
    let work = read("workflows/work.md");
    let repeat = section(&work, "Procedure")
        .lines()
        .find(|l| l.contains(". **repeat** — "))
        .expect("work.md has a repeat step");
    assert!(
        repeat.contains("until no card is left"),
        "the repeat step: {repeat}"
    );
}

#[test]
fn work_decision_becomes_a_blocked_note_on_the_card() {
    let work = read("workflows/work.md");
    for phrase in [
        "`LANE-BLOCKED kind=decision",
        "as a note on the card",
        "never a question to the user",
    ] {
        assert!(work.contains(phrase), "work.md lacks {phrase:?}");
    }
    for tool in ["AskUserQuestion", "request_user_input", "question tool"] {
        assert!(!work.contains(tool), "work.md names {tool}");
    }
}

#[test]
fn work_is_picked_by_the_leading_word() {
    let skill = read("SKILL.md");
    assert!(
        skill.contains("`/moai work`"),
        "SKILL.md names `/moai work`"
    );
    assert!(
        skill.contains(WORK_SENTENCE),
        "SKILL.md lacks {WORK_SENTENCE:?}"
    );
}

#[test]
fn absorb_project_carries_the_setup_steps() {
    let project = read("workflows/project.md");
    for phrase in [
        "record the first work item with `moai card add`",
        "creating it is a write under the real HOME",
        "run `moai ledger-register-apply` only after its own pick in a selection",
        "Allowed initial local commit",
        "it never implies push or Issue publication",
        "existing commit is never amended",
    ] {
        assert!(project.contains(phrase), "project.md lacks {phrase:?}");
    }
}

#[test]
fn absorb_project_asks_no_language_the_user_scope_holds() {
    let project = read("workflows/project.md");
    for phrase in [
        "the local tool language question is the instruction-file language",
        "asks no conversation, document or commit language that the user scope already holds",
    ] {
        assert!(project.contains(phrase), "project.md lacks {phrase:?}");
    }
}

#[test]
fn absorb_update_carries_a_leftover_preview_that_deletes_nothing() {
    let update = read("workflows/update.md");
    for phrase in [
        "`moai optimize-plan",
        "deletes, moves and writes nothing",
        "Deleting is a `guarded` choice",
    ] {
        assert!(update.contains(phrase), "update.md lacks {phrase:?}");
    }
}

#[test]
fn absorb_run_carries_the_verify_sections() {
    let run = read("workflows/run.md");
    for name in [
        "Check-run rules",
        "Lane worktrees and build cache",
        "Change-specific rules",
        "Shell and helper rules",
    ] {
        assert_eq!(
            run.matches(&format!("\n## {name}\n")).count(),
            1,
            "run.md holds `## {name}` once"
        );
    }
}

#[test]
fn absorb_entry_names_no_retired_route() {
    for rel in ["SKILL.md", "references/chains/index.md"] {
        let text = read(rel);
        for word in ["SETUP", "VERIFY", "OPTIMIZE", "/moai optimize"] {
            assert!(!text.contains(word), "{rel} names {word}");
        }
    }
}

#[test]
fn refs_flow_reference_holds_each_stated_item() {
    let flow = read(FLOW_REF);
    assert!(
        flow.len() <= REF_LIMIT,
        "{FLOW_REF} is {} bytes, over {REF_LIMIT}",
        flow.len()
    );
    for phrase in [
        MEANING,
        FLOW_LINE,
        "`.claude/worktrees`",
        "submits two things and nothing else: its head commit and the base its checks ran on",
        "`moai card land <card> --head <40 hex> --by leader`",
        "protection hook",
        "Remote checks never block a landing",
    ] {
        assert!(flow.contains(phrase), "{FLOW_REF} lacks {phrase:?}");
    }
    let lower = flow.to_lowercase();
    for (_, rule) in RULES {
        assert!(lower.contains(rule), "{FLOW_REF} lacks the rule {rule:?}");
    }
}

#[test]
fn refs_meaning_is_in_exactly_one_shipped_text() {
    let holders: Vec<String> = shipped_texts()
        .into_iter()
        .filter(|(_, text)| text.contains(MEANING))
        .map(|(path, _)| path)
        .collect();
    assert_eq!(holders, [format!("skills/moai/{FLOW_REF}")]);
}

#[test]
fn refs_no_shipped_text_spells_the_rules_as_letters() {
    const NONE: [(char, &str); 0] = [];
    // The checker finds a planted spelling, so an empty result below means none is there.
    assert_eq!(
        letter_spellings("N = no tip chase"),
        [('n', "no tip chase")]
    );
    assert_eq!(
        letter_spellings("- **X** — exactly-once landing"),
        [('x', "exactly-once landing")]
    );
    assert_eq!(
        letter_spellings("e**X**actly-once landing"),
        [('x', "exactly-once landing")]
    );
    assert_eq!(
        letter_spellings("- **T**: a guarded tip"),
        [('t', "a guarded tip")]
    );
    assert_eq!(
        letter_spellings("T = guarded tip"),
        [('t', "a guarded tip")]
    );
    assert_eq!(
        letter_spellings(
            "Four rules: no tip chase, evidence before landing, exactly-once landing and a guarded tip."
        ),
        NONE
    );
    for (path, text) in shipped_texts() {
        assert_eq!(letter_spellings(&text), NONE, "{path}");
    }
}

#[test]
fn refs_push_sentence_names_the_remote_copy() {
    let flow = read(FLOW_REF);
    let pushes: Vec<&str> = flow
        .split(". ")
        .flat_map(|s| s.split('\n'))
        .filter(|s| s.contains("push") && s.contains("fast-forward"))
        .collect();
    assert!(!pushes.is_empty(), "{FLOW_REF} has no push sentence");
    for sentence in pushes {
        assert!(
            sentence.contains(PUSH_CONDITION),
            "the push sentence lacks its condition: {sentence}"
        );
    }
}

#[test]
fn refs_audit_contract_holds_each_stated_item() {
    let audit = read(AUDIT_REF);
    assert!(
        audit.len() <= REF_LIMIT,
        "{AUDIT_REF} is {} bytes, over {REF_LIMIT}",
        audit.len()
    );
    for name in [
        "What the auditor receives",
        "What the auditor returns",
        "Who records the verdict",
    ] {
        assert!(
            !section(&audit, name).trim().is_empty(),
            "{AUDIT_REF}: `## {name}` is empty"
        );
    }
    assert!(
        audit.contains("The author of a change never audits it"),
        "{AUDIT_REF} lacks the author rule"
    );
}

#[test]
fn refs_audit_contract_states_the_five_side_rules() {
    let audit = read(AUDIT_REF);
    for phrase in [
        "A local code review of the `<base>..<head>` range is allowed evidence",
        "never with its `--fix` or `--comment` option",
        "A re-audit goes to the first auditor",
        "A lane's tree-bound check receipt may be cited",
        "No forked session and no advisor session audits",
        "A finding that names no file and no line counts as no finding",
    ] {
        assert!(audit.contains(phrase), "{AUDIT_REF} lacks {phrase:?}");
    }
}

#[test]
fn refs_audit_contract_points_to_the_check_run_rules() {
    let audit = read(AUDIT_REF);
    assert!(
        audit.contains("`workflows/run.md`") && audit.contains("\"Check-run rules\""),
        "{AUDIT_REF} does not point to RUN's check-run rules"
    );
    // RUN's own words stay in RUN.
    for restated in ["Citation conditions", "| Role and change |", "head_sha"] {
        assert!(
            !audit.contains(restated),
            "{AUDIT_REF} restates {restated:?}"
        );
    }
}

#[test]
fn refs_run_check_run_rules_hold_the_build_rule() {
    let run = read("workflows/run.md");
    let rules = section(&run, "Check-run rules");
    for phrase in BUILD_RULE {
        assert!(
            rules.contains(phrase),
            "run.md \"Check-run rules\" lacks {phrase:?}"
        );
    }
    // The retired rule's home build folder is not the rule any more.
    let home_build = format!("~/{}/build", moai_cli::state_root::MOAI_DIR);
    assert!(
        !rules.contains(&home_build),
        "run.md \"Check-run rules\" names {home_build}"
    );
}

#[test]
fn refs_no_shipped_text_names_the_removed_build_folder_reference() {
    let naming: Vec<String> = shipped_texts()
        .into_iter()
        .filter(|(_, text)| text.contains("internal/build-folder"))
        .map(|(path, _)| path)
        .collect();
    assert!(
        naming.is_empty(),
        "still name internal/build-folder: {naming:?}"
    );
}

#[test]
fn scrub_shipped_texts_name_no_retired_host_mode_or_gate() {
    let root = Path::new(env!("CARGO_MANIFEST_DIR"));
    let hits = scrub_scan(root, |_, text| req8_hits(text));
    assert!(hits.is_empty(), "REQ-8 words:\n{}", hits.join("\n"));
}

#[test]
fn scrub_shipped_texts_name_no_retired_tool() {
    let root = Path::new(env!("CARGO_MANIFEST_DIR"));
    let hits = scrub_scan(root, req9_hits);
    assert!(hits.is_empty(), "REQ-9 tools:\n{}", hits.join("\n"));
}

#[test]
fn scrub_finds_a_word_planted_in_a_copy() {
    let planted = "skills/moai/workflows/work.md";
    for (word, scan) in [
        (
            "Run it as `$moai work` in Codex CLI.",
            (|_: &str, text: &str| req8_hits(text)) as fn(&str, &str) -> Vec<(usize, &'static str)>,
        ),
        ("Keep the list with TodoWrite.", req9_hits),
    ] {
        let copy = shipped_copy();
        let before = scrub_scan(copy.path(), scan);
        let file = copy.path().join(planted);
        let mut text = fs::read_to_string(&file).unwrap();
        let line = text.lines().count() + 1;
        text.push_str(word);
        text.push('\n');
        fs::write(&file, text).unwrap();
        let after = scrub_scan(copy.path(), scan);
        let new: Vec<&String> = after.iter().filter(|h| !before.contains(h)).collect();
        assert!(!new.is_empty(), "the scan missed {word:?}");
        for hit in new {
            assert!(
                hit.starts_with(&format!("{planted}:{line}: ")),
                "{word:?}: {hit}"
            );
        }
        // The same word in an excepted file is left out.
        let excepted = copy.path().join(SCRUB_EXCEPTIONS[0]);
        let mut text = fs::read_to_string(&excepted).unwrap();
        text.push_str(word);
        fs::write(&excepted, text).unwrap();
        assert_eq!(scrub_scan(copy.path(), scan).len(), after.len(), "{word:?}");
    }
}

#[test]
fn scrub_line_checks_read_what_they_say() {
    const NONE: [(usize, &str); 0] = [];
    // REQ-8: each spelling is found; a word inside another word is not OMP.
    assert_eq!(req8_hits("ask the OMP host"), [(1, "omp")]);
    assert_eq!(req8_hits("a compact table"), NONE);
    assert_eq!(req8_hits("x\nCODEX_HOME"), [(2, "codex")]);
    assert_eq!(req8_hits("the MODE-AGENT workflow"), [(1, "mode-agent")]);
    // REQ-8: AGENTS.md is named only as the older file.
    assert_eq!(
        req8_hits("read the project's AGENTS.md"),
        [(1, "AGENTS.md")]
    );
    // `older` is a word of its own: the `older` inside `folder` is no exception.
    assert_eq!(
        req8_hits("the state folder and the AGENTS.md section"),
        [(1, "AGENTS.md")]
    );
    assert_eq!(
        req8_hits("moves the older `AGENTS.md` and `AGENTS.local.md` into `CLAUDE.md`"),
        NONE
    );
    // REQ-9: tools as written, Grep and Glob as words, the shell-injection line, a skill's model
    // key, and TaskCreate only where the host offers it.
    assert_eq!(req9_hits("agents/x.md", "use the Grep tool"), [(1, "Grep")]);
    assert_eq!(req9_hits("agents/x.md", "a Globber, `grep -rn`"), NONE);
    assert_eq!(
        req9_hits("skills/moai/SKILL.md", "  !`git status`"),
        [(1, "shell-injection block")]
    );
    assert_eq!(
        req9_hits(
            "skills/moai/SKILL.md",
            "---\nname: moai\nmodel: opus\n---\nmodel: text\n"
        ),
        [(3, "model key")]
    );
    assert_eq!(req9_hits("agents/x.md", "---\nmodel: opus\n---\n"), NONE);
    assert_eq!(
        req9_hits("agents/x.md", "keep a TaskCreate list"),
        [(1, "TaskCreate")]
    );
    assert_eq!(
        req9_hits(
            "agents/x.md",
            "keep the task list current if the TaskCreate tool is offered"
        ),
        NONE
    );
}

#[test]
fn step_keys_and_sections_read_what_they_say() {
    let procedure =
        "1. **take** — a\n   - 2. **no** — nested\n2. **plan** — b\n3. **x**: no dash\n";
    assert_eq!(
        step_keys(procedure),
        [(1, "take".to_string()), (2, "plan".to_string())]
    );
    let text = "# T\n\n## Procedure\n\n1. **take** — a\n\n## Next\n\nz\n";
    assert_eq!(section(text, "Procedure"), "\n1. **take** — a\n");
}
