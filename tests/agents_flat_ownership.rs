//! SPEC-20261124-update-flat-agent-ownership (REQ-1 .. REQ-9, REQ-11): a flat base agent file
//! that holds the bytes of its role's render or of its package file is ours; one that matches
//! neither its record nor those is a conflict the person chooses for (keep mine, use MoAI's,
//! show the difference); the person's bytes stay in the receipt; a conflict left over makes the
//! agents result `partial`. Every case runs the library on a temporary HOME; the real HOME and
//! the user's agents folder are never read or written.

use moai_cli::agent_defs::{package_agents, render};
use moai_cli::agent_install::{self, FilePlan, Plan};
use serde_json::Value;
use std::{
    collections::BTreeMap,
    fs,
    io::{Cursor, Read},
    path::{Path, PathBuf},
};
use tempfile::TempDir;

const PERM: &str = "fixture-approved";
const BASE: [&str; 5] = ["audit", "docs", "impl", "plan", "research"];
const HOSTS: [&str; 1] = ["claude"];

fn repo() -> PathBuf {
    PathBuf::from(env!("CARGO_MANIFEST_DIR"))
}

fn package() -> PathBuf {
    repo().join("skills/moai")
}

fn agents() -> PathBuf {
    repo().join("agents")
}

/// The text the package renders for the role: what the nested copy holds after an apply.
fn render_of(role: &str) -> String {
    let defs = package_agents(&agents(), &package()).unwrap();
    let def = defs
        .iter()
        .find(|d| d.name == format!("moai-{role}"))
        .unwrap();
    render(def, "claude").unwrap().text
}

/// The role's file in the package being applied (`agents/<role>.md`).
fn source_of(role: &str) -> String {
    fs::read_to_string(agents().join(format!("{role}.md"))).unwrap()
}

struct Fx {
    _tmp: TempDir,
    home: PathBuf,
}

fn fx() -> Fx {
    let tmp = TempDir::new().unwrap();
    let home = fs::canonicalize(tmp.path()).unwrap().join("home");
    fs::create_dir_all(&home).unwrap();
    Fx { _tmp: tmp, home }
}

impl Fx {
    fn claude_agents(&self) -> PathBuf {
        self.home.join(".claude/agents")
    }
    fn flat(&self, role: &str) -> PathBuf {
        self.claude_agents().join(format!("moai-{role}.md"))
    }
    fn nested(&self, role: &str) -> PathBuf {
        self.claude_agents().join(format!("moai/moai-{role}.md"))
    }
    fn plan(&self) -> Plan {
        agent_install::plan(&self.home, &package(), &agents()).unwrap()
    }
    /// The apply of a fresh plan with the given `host:file` choices.
    fn try_apply(&self, plan: &Plan, choices: &[&str]) -> Result<Value, String> {
        let choices: Vec<String> = choices.iter().map(|c| c.to_string()).collect();
        agent_install::apply_with_choices(
            &self.home,
            &package(),
            &agents(),
            &HOSTS,
            &plan.plan_digest,
            PERM,
            &choices,
        )
    }
    fn apply(&self, plan: &Plan, choices: &[&str]) -> Value {
        self.try_apply(plan, choices).unwrap()
    }
    fn write_flat(&self, role: &str, text: &str) {
        fs::create_dir_all(self.claude_agents()).unwrap();
        fs::write(self.flat(role), text).unwrap();
    }
    /// The earlier layout: the same install, then every file moved up to `agents/` (the record
    /// still names the bytes this tool wrote, as an earlier apply left it).
    fn earlier_layout(&self) {
        let first = self.apply(&self.plan(), &[]);
        assert_eq!(first["state"], "applied", "{first}");
        for role in BASE {
            fs::rename(self.nested(role), self.flat(role)).unwrap();
        }
        fs::remove_dir(self.claude_agents().join("moai")).unwrap();
    }
    fn state_owned(&self) -> BTreeMap<String, String> {
        let path = moai_cli::home_root::path(&self.home).join("agents/claude.json");
        let json: Value = serde_json::from_slice(&fs::read(path).unwrap()).unwrap();
        serde_json::from_value(json["owned"].clone()).unwrap()
    }
}

fn line(plan: &Plan, role: &str) -> FilePlan {
    let name = format!("moai-{role}.md");
    plan.hosts
        .iter()
        .find(|h| h.host == "claude")
        .unwrap()
        .files
        .iter()
        .find(|f| f.file == name)
        .unwrap_or_else(|| panic!("no plan line for {name}"))
        .clone()
}

/// Every file below `dir` with its bytes, to prove that nothing was written.
fn tree(dir: &Path) -> BTreeMap<PathBuf, Vec<u8>> {
    fn walk(dir: &Path, out: &mut BTreeMap<PathBuf, Vec<u8>>) {
        for e in fs::read_dir(dir).unwrap() {
            let e = e.unwrap();
            let p = e.path();
            if e.file_type().unwrap().is_dir() {
                walk(&p, out);
            } else {
                out.insert(p.clone(), fs::read(&p).unwrap());
            }
        }
    }
    let mut out = BTreeMap::new();
    walk(dir, &mut out);
    out
}

fn choose(plan: &Plan, answers: &str, korean: bool) -> (Vec<String>, String, String) {
    let mut input = Cursor::new(answers.as_bytes().to_vec());
    let mut out = Vec::new();
    let chosen = agent_install::choose_flat_conflicts(
        plan,
        &package(),
        &agents(),
        &mut input,
        &mut out,
        korean,
    )
    .unwrap();
    let mut rest = String::new();
    input.read_to_string(&mut rest).unwrap();
    (chosen, String::from_utf8(out).unwrap(), rest)
}

/// REQ-1: a flat file equal to the render, or equal to the package file of its role, is ours
/// whatever the record says (no record, or a record that names other bytes).
#[test]
fn a_flat_file_equal_to_the_render_or_the_package_file_is_ours() {
    // The render and the package file differ for every role, so the two cases are two cases.
    for role in BASE {
        assert_ne!(render_of(role), source_of(role), "{role}");
    }
    // No record at all: audit holds the render, research holds the package file.
    let none = fx();
    none.write_flat("audit", &render_of("audit"));
    none.write_flat("research", &source_of("research"));
    // A record that names other bytes: the earlier layout recorded the render of every role,
    // then impl and docs were replaced by the package file.
    let other = fx();
    other.earlier_layout();
    other.write_flat("impl", &source_of("impl"));
    other.write_flat("docs", &source_of("docs"));
    for (fx, roles) in [(&none, ["audit", "research"]), (&other, ["impl", "docs"])] {
        let plan = fx.plan();
        for role in roles {
            let l = line(&plan, role);
            assert_eq!(l.legacy.as_deref(), Some("remove"), "{role}: {l:?}");
            assert_eq!(l.action, "create", "{role}: {l:?}");
        }
        let v = fx.apply(&plan, &[]);
        assert_eq!(v["state"], "applied", "{v}");
        for role in roles {
            assert_eq!(
                fs::read_to_string(fx.nested(role)).unwrap(),
                render_of(role),
                "{role}: nested copy"
            );
            assert!(!fx.flat(role).exists(), "{role}: flat file");
        }
        let again = fx.apply(&fx.plan(), &[]);
        assert_eq!(again["state"], "unchanged", "{again}");
    }
}

/// REQ-2, REQ-7: a flat file that matches neither its record nor the render nor the package
/// file stays a conflict, and an apply with no choice leaves it and writes no nested copy.
#[test]
fn a_flat_file_that_matches_nothing_stays_a_conflict() {
    // Edited: the record names the render, the file says something else.
    let edited = fx();
    edited.earlier_layout();
    edited.write_flat("impl", "I edited this one\n");
    // Foreign: nothing was ever installed, the file is somebody else's.
    let foreign = fx();
    foreign.write_flat("plan", "someone else's plan agent\n");
    for (fx, role, legacy, bytes) in [
        (&edited, "impl", "kept_edited", "I edited this one\n"),
        (
            &foreign,
            "plan",
            "kept_foreign",
            "someone else's plan agent\n",
        ),
    ] {
        let plan = fx.plan();
        let l = line(&plan, role);
        assert_eq!(l.action, "conflict", "{l:?}");
        assert_eq!(l.legacy.as_deref(), Some(legacy), "{l:?}");
        fx.apply(&plan, &[]);
        assert_eq!(fs::read_to_string(fx.flat(role)).unwrap(), bytes);
        assert!(!fx.nested(role).exists(), "nested copy of {role}");
    }
}

/// REQ-3: the prompt asks once per conflicting file, offers the three answers, and never asks
/// about a file that is no conflict.
#[test]
fn the_choice_asks_once_per_conflicting_file() {
    let fx = fx();
    fx.earlier_layout();
    fx.write_flat("impl", "mine impl\n");
    fx.write_flat("plan", "mine plan\n");
    let plan = fx.plan();

    let (chosen, out, rest) = choose(&plan, "k\nm\nEXTRA\n", false);
    assert_eq!(chosen, ["claude:moai-plan.md"], "{out}");
    assert_eq!(rest, "EXTRA\n", "two answers were read: {out}");
    assert!(
        out.contains("moai-impl.md") && out.contains("moai-plan.md"),
        "{out}"
    );
    assert!(
        !out.contains("moai-audit.md"),
        "a file that moves is no question: {out}"
    );
    for word in ["keep mine", "use MoAI's", "show the difference"] {
        assert!(out.contains(word), "{word}: {out}");
    }

    let (_, ko, _) = choose(&plan, "k\nk\n", true);
    for word in ["내 것 유지", "MoAI 것 쓰기", "차이 보기"] {
        assert!(ko.contains(word), "{word}: {ko}");
    }
}

/// REQ-4: the difference lists only the lines that differ (`-` mine, `+` MoAI's) and the same
/// file is asked again; the answer after it decides.
#[test]
fn the_difference_is_shown_and_the_question_comes_again() {
    let fx = fx();
    fx.earlier_layout();
    let original = render_of("impl");
    let mine: String = original
        .lines()
        .map(|l| {
            if l.starts_with("description:") {
                "description: mine".to_string()
            } else {
                l.to_string()
            }
        })
        .collect::<Vec<_>>()
        .join("\n")
        + "\n";
    assert_ne!(mine, original);
    fx.write_flat("impl", &mine);
    let plan = fx.plan();

    let (chosen, out, _) = choose(&plan, "d\nm\n", false);
    assert_eq!(chosen, ["claude:moai-impl.md"], "{out}");
    assert_eq!(out.matches("keep mine").count(), 2, "asked again: {out}");
    let minus: Vec<&str> = out.lines().filter(|l| l.starts_with("- ")).collect();
    let plus: Vec<&str> = out.lines().filter(|l| l.starts_with("+ ")).collect();
    assert_eq!(minus, ["- description: mine"], "{out}");
    assert_eq!(plus.len(), 1, "{out}");
    assert!(plus[0].starts_with("+ description:"), "{out}");

    // A second look and then keep: the answer after the difference decides.
    let (chosen, _, _) = choose(&plan, "d\nd\nk\n", false);
    assert!(chosen.is_empty());
}

/// Edited flat impl, everything else as the earlier layout left it; returns the fixture, the
/// plan and the person's bytes.
fn edited_impl() -> (Fx, Plan, &'static str) {
    const MINE: &str = "I edited this one\n";
    let fx = fx();
    fx.earlier_layout();
    fx.write_flat("impl", MINE);
    let plan = fx.plan();
    (fx, plan, MINE)
}

/// REQ-5: use MoAI's writes the nested copy, removes the flat file and keeps the person's bytes
/// in the receipt; the record names the nested bytes.
#[test]
fn use_moais_keeps_the_bytes_and_moves_the_file() {
    let (fx, plan, mine) = edited_impl();
    let v = fx.apply(&plan, &["claude:moai-impl.md"]);
    assert_eq!(v["state"], "applied", "{v}");
    assert_eq!(
        fs::read_to_string(fx.nested("impl")).unwrap(),
        render_of("impl")
    );
    assert!(!fx.flat("impl").exists());
    let id = v["receipt"].as_str().unwrap();
    let kept = moai_cli::home_root::path(&fx.home)
        .join("agents/receipts")
        .join(id)
        .join("before/claude/legacy/moai-impl.md");
    assert_eq!(fs::read_to_string(kept).unwrap(), mine);
    assert_eq!(
        fx.state_owned()["moai-impl.md"],
        moai_cli::digest(render_of("impl").as_bytes())
    );
}

/// REQ-6: the revert of that receipt puts the person's flat file back and removes the nested copy.
#[test]
fn the_revert_restores_the_persons_file() {
    let (fx, plan, mine) = edited_impl();
    let v = fx.apply(&plan, &["claude:moai-impl.md"]);
    // The file moved first, so the revert has something to put back.
    assert_eq!(
        fs::read_to_string(fx.nested("impl")).unwrap(),
        render_of("impl"),
        "{v}"
    );
    assert!(!fx.flat("impl").exists(), "{v}");
    let id = v["receipt"].as_str().unwrap().to_string();
    let r = agent_install::revert(&fx.home, &id, PERM).unwrap();
    assert_eq!(r["state"], "reverted", "{r}");
    assert_eq!(fs::read_to_string(fx.flat("impl")).unwrap(), mine);
    assert!(!fx.nested("impl").exists());
    // The restored file is still the person's: the record does not name it, so the next update
    // asks again instead of moving it unasked.
    let l = line(&fx.plan(), "impl");
    assert_eq!(l.action, "conflict", "{l:?}");
    assert!(
        matches!(l.legacy.as_deref(), Some("kept_edited" | "kept_foreign")),
        "{l:?}"
    );
}

/// REQ-7: no choice, keep mine, an empty answer, end of input and a word that is no answer each
/// leave both paths of the file byte for byte as they were.
#[test]
fn keep_and_no_answer_leave_both_paths_untouched() {
    let (fx, plan, mine) = edited_impl();
    for answers in ["k\n", "\n", "", "x\n", "K\n"] {
        let (chosen, out, _) = choose(&plan, answers, false);
        assert!(chosen.is_empty(), "{answers:?}: {chosen:?} {out}");
    }
    fx.apply(&plan, &[]);
    assert_eq!(fs::read_to_string(fx.flat("impl")).unwrap(), mine);
    assert!(!fx.nested("impl").exists());
}

/// REQ-8: a choice for a file that is no flat-file conflict is refused as a whole, with nothing
/// written (no receipt, no moved file).
#[test]
fn a_choice_for_a_file_that_is_no_conflict_is_refused() {
    let (fx, plan, _) = edited_impl();
    let before = tree(&fx.home);
    for bad in [
        &["claude:moai-audit.md"][..],
        &["claude:nope.md"],
        &["codex:moai-impl.md"],
        &["moai-impl.md"],
        &["claude:../moai-impl.md"],
        &["claude:moai-impl.md", "claude:moai-audit.md"],
    ] {
        let e = fx.try_apply(&plan, bad);
        assert!(e.is_err(), "{bad:?} was accepted: {e:?}");
        assert_eq!(tree(&fx.home), before, "{bad:?} wrote something");
    }
}

/// REQ-9: a conflict left after the apply makes the result `partial` with the kept names, with
/// other writes and with none; a clean install keeps `applied` and `unchanged` and no `kept`.
#[test]
fn a_left_conflict_makes_the_agents_result_partial() {
    // With other writes: audit and the rest move, impl is kept.
    let (fx, plan, _) = edited_impl();
    let v = fx.apply(&plan, &[]);
    assert_eq!(v["state"], "partial", "{v}");
    assert_eq!(v["kept"], serde_json::json!(["claude:moai-impl.md"]), "{v}");
    assert!(!v["written"].as_array().unwrap().is_empty(), "{v}");

    // With no other write: everything is installed nested, and a flat copy sits beside it.
    let alone = self::fx();
    let first = alone.apply(&alone.plan(), &[]);
    assert_eq!(first["state"], "applied", "{first}");
    assert!(first.get("kept").is_none(), "{first}");
    alone.write_flat("impl", "my own old copy\n");
    let v = alone.apply(&alone.plan(), &[]);
    assert_eq!(v["state"], "partial", "{v}");
    assert_eq!(v["kept"], serde_json::json!(["claude:moai-impl.md"]), "{v}");
    assert_eq!(v["written"], serde_json::json!([]), "{v}");

    // Choosing MoAI's clears it: the next apply is a clean one.
    let v = alone.apply(&alone.plan(), &["claude:moai-impl.md"]);
    assert_eq!(v["state"], "applied", "{v}");
    assert!(v.get("kept").is_none(), "{v}");
    let again = alone.apply(&alone.plan(), &[]);
    assert_eq!(again["state"], "unchanged", "{again}");
    assert!(again.get("kept").is_none(), "{again}");
}

/// REQ-11: a plan whose only item is a flat-file conflict is worth a question, a plan with
/// nothing to write or choose is not, and the write predicate itself stays as it was.
#[test]
fn a_conflict_alone_is_worth_a_question() {
    let fx = fx();
    fx.apply(&fx.plan(), &[]);
    let clean = fx.plan();
    assert!(!agent_install::asks_anything(&clean));

    fx.write_flat("impl", "my own old copy\n");
    let conflict = fx.plan();
    let files = &conflict.hosts[0].files;
    assert!(
        files.iter().any(|f| f.action == "conflict"),
        "the setup must hold a conflict"
    );
    assert!(
        !files.iter().any(agent_install::writes_anything),
        "a conflict alone writes nothing"
    );
    assert!(agent_install::asks_anything(&conflict));
}

/// A choice about the flat file never overwrites a nested copy the person edited too: the apply
/// is refused before any write and both files stay as they were.
#[test]
fn a_choice_never_overwrites_an_edited_nested_copy() {
    let (fx, _, mine) = edited_impl();
    fs::create_dir_all(fx.claude_agents().join("moai")).unwrap();
    fs::write(fx.nested("impl"), "edited nested copy\n").unwrap();
    let plan = fx.plan();
    assert_eq!(line(&plan, "impl").action, "conflict");
    let before = tree(&fx.home);
    let e = fx.try_apply(&plan, &["claude:moai-impl.md"]);
    assert!(e.is_err(), "the choice overwrote the nested copy: {e:?}");
    assert_eq!(tree(&fx.home), before, "something was written");
    assert_eq!(fs::read_to_string(fx.flat("impl")).unwrap(), mine);
    assert_eq!(
        fs::read_to_string(fx.nested("impl")).unwrap(),
        "edited nested copy\n"
    );
}

/// REQ-6: when the nested copy was already there and the choice wrote none, the revert still
/// puts the person's file back and leaves it the person's (the record does not name it).
#[test]
fn the_revert_of_a_choice_that_wrote_no_nested_copy_keeps_the_file_the_persons() {
    let fx = fx();
    fx.apply(&fx.plan(), &[]);
    fx.write_flat("impl", "my own old copy\n");
    let v = fx.apply(&fx.plan(), &["claude:moai-impl.md"]);
    assert_eq!(v["state"], "applied", "{v}");
    assert!(!fx.flat("impl").exists(), "{v}");
    let id = v["receipt"].as_str().unwrap().to_string();
    let r = agent_install::revert(&fx.home, &id, PERM).unwrap();
    assert_eq!(r["state"], "reverted", "{r}");
    assert_eq!(
        fs::read_to_string(fx.flat("impl")).unwrap(),
        "my own old copy\n"
    );
    let l = line(&fx.plan(), "impl");
    assert_eq!(l.action, "conflict", "{l:?}");
    assert!(
        matches!(l.legacy.as_deref(), Some("kept_edited" | "kept_foreign")),
        "{l:?}"
    );
}

/// REQ-4: the difference shows a control character of the person's file as text, so the file
/// cannot move the cursor or retitle the window.
#[test]
fn the_difference_shows_control_characters_as_text() {
    let fx = fx();
    fx.earlier_layout();
    fx.write_flat("impl", "description: a\u{1b}[2Jb\n");
    let plan = fx.plan();
    let (_, out, _) = choose(&plan, "d\nk\n", false);
    assert!(!out.contains('\u{1b}'), "an escape was sent: {out:?}");
    assert!(
        out.lines().any(|l| l.starts_with("- description: a")),
        "{out:?}"
    );
}

/// The dangerous path of the choice: a link at the flat path, before the plan and between the
/// plan and the apply. The link is never followed, never removed, its target keeps its bytes,
/// and a choice for it writes nothing.
#[cfg(unix)]
#[test]
fn a_link_at_the_flat_path_is_never_followed_or_removed() {
    use std::os::unix::fs::symlink;
    // Before the plan: the host is blocked and no conflict is offered.
    let fx = fx();
    fx.earlier_layout();
    let outside = fx.home.join("outside.md");
    fs::write(&outside, "not an agent\n").unwrap();
    fs::remove_file(fx.flat("impl")).unwrap();
    symlink(&outside, fx.flat("impl")).unwrap();
    let plan = fx.plan();
    assert_eq!(plan.hosts[0].preflight, "blocked", "{:?}", plan.hosts[0]);
    assert!(agent_install::flat_conflicts(&plan).is_empty());
    assert!(fx.try_apply(&plan, &["claude:moai-impl.md"]).is_err());
    assert!(fs::symlink_metadata(fx.flat("impl"))
        .unwrap()
        .file_type()
        .is_symlink());
    assert_eq!(fs::read_to_string(&outside).unwrap(), "not an agent\n");
    assert!(!fx.nested("impl").exists());

    // Between the plan and the apply: a plain conflict file is swapped for a link.
    let (fx, plan, _) = edited_impl();
    let outside = fx.home.join("outside.md");
    fs::write(&outside, "not an agent\n").unwrap();
    fs::remove_file(fx.flat("impl")).unwrap();
    symlink(&outside, fx.flat("impl")).unwrap();
    let e = fx.try_apply(&plan, &["claude:moai-impl.md"]);
    assert!(e.is_err(), "the apply followed a link: {e:?}");
    assert!(fs::symlink_metadata(fx.flat("impl"))
        .unwrap()
        .file_type()
        .is_symlink());
    assert_eq!(fs::read_to_string(&outside).unwrap(), "not an agent\n");
    assert!(!fx.nested("impl").exists());
}
