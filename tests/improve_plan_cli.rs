//! `moai improve-plan` (SPEC-20261121-self-improvement-loop, slice S2): a read-only proposal for
//! the project's own skills and agents, built by deterministic code and held by static gates.
//!
//! Every command runs against a temporary HOME and a project inside it, so one snapshot of the
//! HOME shows that nothing was written. The generation manifest is written with the product's
//! own path function (the state folder is never spelled here). Acceptance AC-10 to AC-14 and
//! AC-17 of the SPEC are these tests (the pure budget function has its own unit tests in
//! `src/ops/improve_plan.rs`).
#![cfg(unix)]

mod common;

use common::home::TempHome;
use serde_json::{json, Value};
use sha2::{Digest, Sha256};
use std::os::unix::fs::symlink;
use std::path::{Path, PathBuf};
use std::process::Output;

const SKILL: &str = ".claude/skills/moai-prj-demo/SKILL.md";
const AGENT: &str = ".claude/agents/moai/prj-demo.md";
const BEGIN: &str = "<!-- moai:managed:begin id=moai-lessons -->";
const END: &str = "<!-- moai:managed:end id=moai-lessons -->";

struct World {
    home: TempHome,
    project: PathBuf,
}

fn world() -> World {
    let home = TempHome::new();
    let project = home.project("demo");
    World { home, project }
}

fn sha(bytes: &[u8]) -> String {
    format!("{:x}", Sha256::digest(bytes))
}

/// A front matter block with a name and a description of `description` characters.
fn front(description: usize) -> String {
    format!(
        "---\nname: demo\ndescription: {}\n---\n",
        "d".repeat(description)
    )
}

/// One item line in the form the operations document fixes.
fn line(id: &str, rule: &str, why: &str, source: &str) -> String {
    format!("- [{id}] {rule} | why: {why} | source: {source} | status: trial | helped: 0 | hurt: 0")
}

fn item(op: &str, id: &str, rule: &str, why: &str, source: &[&str]) -> String {
    json!({"op": op, "id": id, "rule": rule, "reason": why, "source": source}).to_string()
}

fn add(id: &str) -> String {
    item(
        "add",
        id,
        "Run the check first.",
        "A skipped check was claimed.",
        &["failure:cargo"],
    )
}

fn retire(id: &str) -> String {
    json!({"op": "retire", "id": id}).to_string()
}

impl World {
    fn path(&self, rel: &str) -> PathBuf {
        self.project.join(rel)
    }

    fn put(&self, rel: &str, bytes: &str) {
        let path = self.path(rel);
        std::fs::create_dir_all(path.parent().unwrap()).unwrap();
        std::fs::write(path, bytes).unwrap();
    }

    fn manifest_path(&self) -> PathBuf {
        moai_cli::state_root::path(&self.project, "manifests/generation.json")
    }

    /// The manifest the product writes: one `owned_files` entry per `(path, section_id,
    /// baseline)`.
    fn manifest(&self, entries: &[(&str, &str, &str)]) {
        let files: Vec<Value> = entries
            .iter()
            .map(|(path, section, baseline)| {
                json!({"path": path, "kind": "x", "section_id": section, "baseline_sha256": baseline})
            })
            .collect();
        let path = self.manifest_path();
        std::fs::create_dir_all(path.parent().unwrap()).unwrap();
        std::fs::write(path, json!({"owned_files": files}).to_string()).unwrap();
    }

    /// A project skill and a project agent the manifest owns, each with its baseline equal to
    /// its bytes.
    fn owned(&self, skill: &str, agent: &str) {
        self.put(SKILL, skill);
        self.put(AGENT, agent);
        self.manifest(&[
            (
                SKILL,
                "generation:project-skill:claude:demo",
                &sha(skill.as_bytes()),
            ),
            (
                AGENT,
                "generation:native-role:claude:prj-demo",
                &sha(agent.as_bytes()),
            ),
        ]);
    }

    fn plan(&self, target: &str, items: &[String]) -> Output {
        let project = self.project.to_str().unwrap();
        let mut args = vec!["improve-plan", project, "--target", target];
        for i in items {
            args.push("--item");
            args.push(i);
        }
        self.home.command(&args).output().unwrap()
    }
}

fn text(out: &Output) -> String {
    format!(
        "{} {}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    )
}

fn result(out: &Output) -> Value {
    assert_eq!(out.status.code(), Some(0), "{}", text(out));
    let v: Value =
        serde_json::from_slice(&out.stdout).unwrap_or_else(|e| panic!("{e}: {}", text(out)));
    assert_eq!(v["ok"], json!(true), "{v}");
    assert_eq!(v["operation"], json!("improve.plan"), "{v}");
    v["result"].clone()
}

/// The refusal: the error kind, the message and the sorted reason words after `reasons=`.
fn refusal(out: &Output) -> (String, String, Vec<String>) {
    assert_eq!(out.status.code(), Some(1), "{}", text(out));
    let v: Value =
        serde_json::from_slice(&out.stdout).unwrap_or_else(|e| panic!("{e}: {}", text(out)));
    assert_eq!(v["ok"], json!(false), "{v}");
    let message = v["error"]["message"].as_str().unwrap().to_string();
    assert!(
        message.starts_with("improve-plan refused: reasons="),
        "{message}"
    );
    let reasons = message["improve-plan refused: reasons=".len()..]
        .split(|c: char| c == ' ' || c == '·')
        .next()
        .unwrap()
        .split(',')
        .map(String::from)
        .collect();
    (
        v["error"]["kind"].as_str().unwrap().to_string(),
        message,
        reasons,
    )
}

fn reasons_of(w: &World, target: &str, items: &[String]) -> Vec<String> {
    let before = w.home.snapshot();
    let out = w.plan(target, items);
    let (kind, message, reasons) = refusal(&out);
    assert_eq!(kind, "invalid_input", "{message}");
    assert_eq!(
        w.home.snapshot(),
        before,
        "a refusal writes nothing: {message}"
    );
    reasons
}

#[test]
fn a_listed_project_skill_and_agent_are_planned_and_nothing_is_written() {
    let w = world();
    let skill = format!("{}\n# demo\n\nBody.\n", front(20));
    let agent = format!("{}\nAgent body.\n", front(20));
    w.owned(&skill, &agent);
    let before = w.home.snapshot();

    for (rel, old, kind) in [
        (SKILL, &skill, "project-skill"),
        (AGENT, &agent, "native-role"),
    ] {
        let r = result(&w.plan(rel, &[add("run-check")]));
        assert_eq!(r["action"], json!("plan"), "{r}");
        assert_eq!(r["writes"], json!(false), "{r}");
        assert_eq!(r["grade"], json!("A"), "{r}");
        assert_eq!(r["target"], json!(rel), "{r}");
        assert_eq!(r["kind"], json!(kind), "{r}");
        assert_eq!(r["generation"], json!("unchanged"), "{r}");
        assert_eq!(r["old_sha256"], json!(sha(old.as_bytes())), "{r}");
        let item_line = line(
            "run-check",
            "Run the check first.",
            "A skipped check was claimed.",
            "failure:cargo",
        );
        let new = format!("{old}\n{BEGIN}\n{item_line}\n{END}\n");
        assert_eq!(r["new_sha256"], json!(sha(new.as_bytes())), "{r}");
        assert_eq!(
            r["items"],
            json!([{"op": "add", "id": "run-check", "line": item_line}]),
            "{r}"
        );
        assert!(
            r["plan_digest"].as_str().unwrap().starts_with("sha256:"),
            "{r}"
        );
        assert_eq!(
            w.home.snapshot(),
            before,
            "a plan writes no file, folder or time"
        );
    }
}

#[test]
fn a_target_the_manifest_does_not_own_as_a_project_file_is_refused_not_allowed() {
    let w = world();
    let skill = format!("{}\nBody.\n", front(5));
    w.owned(&skill, &skill);
    w.put(".claude/skills/other/SKILL.md", &skill);
    w.put(".claude/skills/moai/SKILL.md", &skill);
    w.put(".claude/settings.json", "{}\n");
    w.put(".claude/hooks/guard.sh", "#!/bin/sh\n");
    w.put(".mcp.json", "{}\n");
    w.put("CLAUDE.md", &skill);
    w.put("CLAUDE.local.md", &skill);
    w.put("docs/note.md", &skill);
    // Owned by the manifest but not a project skill or agent: the path shape and the section
    // kind each refuse on their own.
    w.put(".claude/skills/moai-prj-proj/SKILL.md", &skill);
    let h = sha(skill.as_bytes());
    w.manifest(&[
        (SKILL, "generation:project-skill:claude:demo", &h),
        (
            ".claude/settings.json",
            "generation:project-skill:claude:demo",
            &h,
        ),
        ("CLAUDE.md", "generation:native-role:claude:prj-demo", &h),
        (
            ".claude/skills/moai/SKILL.md",
            "generation:project-skill:claude:demo",
            &h,
        ),
        (
            ".claude/skills/moai-prj-proj/SKILL.md",
            "generation:skill-projection:claude:proj",
            &h,
        ),
        (
            ".claude/skills/moai-prj-codex/SKILL.md",
            "generation:project-skill:codex:codex",
            &h,
        ),
    ]);
    let outside = w.home.path().join("outside.md");
    std::fs::write(&outside, &skill).unwrap();
    symlink(&outside, w.path(".claude/skills/moai-prj-link.md")).unwrap();
    symlink(
        w.path(".claude/skills/moai-prj-demo"),
        w.path(".claude/skills/moai-prj-dirlink"),
    )
    .unwrap();
    std::fs::create_dir_all(w.path(".claude/skills/moai-prj-dir/SKILL.md")).unwrap();

    let absolute = w.path(SKILL).to_str().unwrap().to_string();
    for target in [
        ".claude/skills/other/SKILL.md",
        ".claude/skills/moai/SKILL.md",
        ".claude/settings.json",
        ".claude/hooks/guard.sh",
        ".mcp.json",
        "CLAUDE.md",
        "CLAUDE.local.md",
        "docs/note.md",
        ".claude/skills/moai-prj-proj/SKILL.md",
        ".claude/skills/moai-prj-codex/SKILL.md",
        ".claude/skills/moai-prj-demo/../moai-prj-demo/SKILL.md",
        "../outside.md",
        ".claude/skills/moai-prj-link.md",
        ".claude/skills/moai-prj-dirlink/SKILL.md",
        ".claude/skills/moai-prj-dir/SKILL.md",
        ".claude/skills/moai-prj-demo",
        "",
        absolute.as_str(),
    ] {
        let reasons = reasons_of(&w, target, &[add("x")]);
        assert!(
            reasons.contains(&"not_allowed".to_string()),
            "{target}: {reasons:?}"
        );
    }
    // A manifest that is absent owns nothing.
    std::fs::remove_file(w.manifest_path()).unwrap();
    let reasons = reasons_of(&w, SKILL, &[add("x")]);
    assert!(reasons.contains(&"not_allowed".to_string()), "{reasons:?}");
}

#[test]
fn a_link_in_the_manifest_or_in_the_path_is_never_followed() {
    let w = world();
    let skill = format!("{}\nBody.\n", front(5));
    w.owned(&skill, &skill);
    let real = w.path("real-skills");
    std::fs::create_dir_all(real.join("moai-prj-via")).unwrap();
    std::fs::write(real.join("moai-prj-via/SKILL.md"), &skill).unwrap();
    std::fs::remove_dir_all(w.path(".claude/skills")).unwrap();
    std::fs::create_dir_all(w.path(".claude")).unwrap();
    symlink(&real, w.path(".claude/skills")).unwrap();
    let h = sha(skill.as_bytes());
    w.manifest(&[(
        ".claude/skills/moai-prj-via/SKILL.md",
        "generation:project-skill:claude:via",
        &h,
    )]);
    let reasons = reasons_of(&w, ".claude/skills/moai-prj-via/SKILL.md", &[add("x")]);
    assert!(reasons.contains(&"not_allowed".to_string()), "{reasons:?}");
}

#[test]
fn a_target_in_a_test_or_the_evaluation_set_lists_both_reasons() {
    let w = world();
    let skill = format!("{}\nBody.\n", front(5));
    w.owned(&skill, &skill);
    let evals = moai_cli::state_root::rel(&w.project, "improve/evals/set.md");
    let evals_top = moai_cli::state_root::rel(&w.project, "evals/set.md");
    let h = sha(skill.as_bytes());
    for target in [
        "tests/check.rs",
        "pkg/tests/x.md",
        evals.as_str(),
        evals_top.as_str(),
        "src/foo_test.rs",
        "src/test_foo.py",
        "web/a.spec.ts",
    ] {
        w.put(target, &skill);
        w.manifest(&[
            (SKILL, "generation:project-skill:claude:demo", &h),
            (target, "generation:native-role:claude:prj-demo", &h),
        ]);
        let reasons = reasons_of(&w, target, &[add("x")]);
        assert_eq!(reasons, vec!["not_allowed", "test_weakening"], "{target}");
    }
    // An ordinary file names neither.
    let ok = result(&w.plan(SKILL, &[add("x")]));
    assert_eq!(ok["writes"], json!(false));
}

#[test]
fn one_two_or_three_items_pass_and_zero_or_four_are_refused() {
    let w = world();
    let skill = format!("{}\nBody.\n", front(5));
    w.owned(&skill, &skill);
    for n in 1..=3 {
        let items: Vec<String> = (0..n).map(|i| add(&format!("item-{i}"))).collect();
        let r = result(&w.plan(SKILL, &items));
        assert_eq!(r["items"].as_array().unwrap().len(), n, "{r}");
    }
    for n in [0usize, 4] {
        let items: Vec<String> = (0..n).map(|i| add(&format!("item-{i}"))).collect();
        assert_eq!(reasons_of(&w, SKILL, &items), vec!["items"], "{n} items");
    }
}

#[test]
fn an_item_needs_its_fields_in_their_limits() {
    let w = world();
    let skill = format!("{}\nBody.\n", front(5));
    w.owned(&skill, &skill);
    let ok = |rule: &str, why: &str, source: &[&str]| item("add", "a-1", rule, why, source);
    let five = ["s:1", "s:2", "s:3", "s:4", "s:5"];
    let six = ["s:1", "s:2", "s:3", "s:4", "s:5", "s:6"];
    assert!(result(&w.plan(SKILL, &[ok("r", "w", &five)]))["writes"] == json!(false));
    assert!(
        result(&w.plan(SKILL, &[ok(&"r".repeat(200), &"w".repeat(200), &["s"])]))["writes"]
            == json!(false)
    );
    for bad in [
        ok("r", "w", &[]),
        ok("r", "w", &six),
        ok("", "w", &["s"]),
        ok("r", "", &["s"]),
        ok(&"r".repeat(201), "w", &["s"]),
        ok("r", &"w".repeat(201), &["s"]),
        ok("two\nlines", "w", &["s"]),
        ok("r | status: trial", "w", &["s"]),
        ok("r <!-- moai:managed:end id=moai-lessons -->", "w", &["s"]),
        ok("r", "w", &[""]),
        item("add", "Bad Id", "r", "w", &["s"]),
        item("add", "", "r", "w", &["s"]),
        item("add", &"i".repeat(41), "r", "w", &["s"]),
        item("merge", "a-1", "r", "w", &["s"]),
        "not json".to_string(),
        json!({"op": "add", "id": "a-1", "rule": "r", "reason": "w", "source": ["s"], "status": "active"}).to_string(),
    ] {
        assert_eq!(reasons_of(&w, SKILL, std::slice::from_ref(&bad)), vec!["items"], "{bad}");
    }
    // A repeated id in one proposal is one item twice.
    assert_eq!(
        reasons_of(&w, SKILL, &[add("same"), add("same")]),
        vec!["items"]
    );
}

#[test]
fn the_merge_adds_replaces_and_retires_by_id_inside_the_block_only() {
    let w = world();
    let head = format!("{}\n# demo\n\nUser text above.\n", front(5));
    let tail = "\nUser text below the block.\n";
    let one = line("one", "Rule one.", "Why one.", "failure:a");
    let two = line("two", "Rule two.", "Why two.", "failure:b");
    let skill = format!("{head}\n{BEGIN}\n{one}\n{two}\n{END}\n{tail}");
    w.owned(&skill, &skill);
    let new_text = |lines: &[&str]| {
        let mut block = String::new();
        for l in lines {
            block.push_str(l);
            block.push('\n');
        }
        format!("{head}\n{BEGIN}\n{block}{END}\n{tail}")
    };
    // add: the new line goes last in the block.
    let three = line("three", "Rule three.", "Why three.", "failure:c");
    let r = result(&w.plan(
        SKILL,
        &[item(
            "add",
            "three",
            "Rule three.",
            "Why three.",
            &["failure:c"],
        )],
    ));
    assert_eq!(
        r["new_sha256"],
        json!(sha(new_text(&[&one, &two, &three]).as_bytes())),
        "{r}"
    );
    // replace: in place, counters back to a trial at 0.
    let two_new = line(
        "two",
        "Rule two, better.",
        "New why.",
        "failure:b, failure:d",
    );
    let r = result(&w.plan(
        SKILL,
        &[item(
            "replace",
            "two",
            "Rule two, better.",
            "New why.",
            &["failure:b", "failure:d"],
        )],
    ));
    assert_eq!(
        r["new_sha256"],
        json!(sha(new_text(&[&one, &two_new]).as_bytes())),
        "{r}"
    );
    // retire: the line is gone.
    let r = result(&w.plan(SKILL, &[retire("one")]));
    assert_eq!(
        r["new_sha256"],
        json!(sha(new_text(&[&two]).as_bytes())),
        "{r}"
    );
    assert_eq!(r["items"], json!([{"op": "retire", "id": "one"}]), "{r}");
    // one proposal may mix all three.
    let r = result(&w.plan(
        SKILL,
        &[
            retire("one"),
            item(
                "replace",
                "two",
                "Rule two, better.",
                "New why.",
                &["failure:b", "failure:d"],
            ),
            add("three"),
        ],
    ));
    assert_eq!(
        r["new_sha256"],
        json!(sha(new_text(&[
            &two_new,
            &line(
                "three",
                "Run the check first.",
                "A skipped check was claimed.",
                "failure:cargo"
            )
        ])
        .as_bytes())),
        "{r}"
    );
    // add of an id that exists, replace and retire of one that does not, and a retire of an
    // item that is no longer a trial are refused.
    assert_eq!(reasons_of(&w, SKILL, &[add("one")]), vec!["items"]);
    assert_eq!(reasons_of(&w, SKILL, &[retire("nope")]), vec!["items"]);
    assert_eq!(
        reasons_of(&w, SKILL, &[item("replace", "nope", "r", "w", &["s"])]),
        vec!["items"]
    );
    let kept = format!(
        "{head}\n{BEGIN}\n{}\n{END}\n",
        one.replace("status: trial", "status: kept")
    );
    w.put(SKILL, &kept);
    w.manifest(&[(
        SKILL,
        "generation:project-skill:claude:demo",
        &sha(kept.as_bytes()),
    )]);
    assert_eq!(reasons_of(&w, SKILL, &[retire("one")]), vec!["items"]);
    // a block that cannot be found exactly once is not edited.
    for broken in [
        format!("{head}\n{BEGIN}\n{one}\n"),
        format!("{head}\n{END}\n{one}\n{BEGIN}\n"),
        format!("{head}\n{BEGIN}\n{END}\n{BEGIN}\n{END}\n"),
    ] {
        w.put(SKILL, &broken);
        w.manifest(&[(
            SKILL,
            "generation:project-skill:claude:demo",
            &sha(broken.as_bytes()),
        )]);
        assert_eq!(
            reasons_of(&w, SKILL, &[add("x")]),
            vec!["not_allowed"],
            "{broken}"
        );
    }
}

/// A skill of `lines` lines in all that holds a block with one trial item `old` (so that one
/// more item is one more line), and its manifest.
fn filled(w: &World, rel: &str, section: &str, lines: usize, description: usize) {
    let mut body = front(description);
    body.push_str(BEGIN);
    body.push('\n');
    body.push_str(&line("old", "Old rule.", "Old why.", "failure:old"));
    body.push('\n');
    body.push_str(END);
    body.push('\n');
    let have = body.lines().count();
    assert!(have <= lines, "the fixture is shorter than {lines} lines");
    for _ in have..lines {
        body.push_str("filler\n");
    }
    assert_eq!(body.lines().count(), lines);
    w.put(rel, &body);
    w.manifest(&[(rel, section, &sha(body.as_bytes()))]);
}

#[test]
fn the_line_budget_holds_at_n_and_refuses_at_n_plus_one_unless_one_trial_item_is_retired() {
    for (rel, section, budget) in [
        (SKILL, "generation:project-skill:claude:demo", 500usize),
        (AGENT, "generation:native-role:claude:prj-demo", 200usize),
    ] {
        let w = world();
        // N - 1 lines plus one item is N lines.
        filled(&w, rel, section, budget - 1, 5);
        let r = result(&w.plan(rel, &[add("new")]));
        assert_eq!(r["new_lines"], json!(budget), "{r}");
        assert_eq!(r["budget_lines"], json!(budget), "{r}");
        // N lines plus one item is N + 1 lines.
        filled(&w, rel, section, budget, 5);
        assert_eq!(reasons_of(&w, rel, &[add("new")]), vec!["budget"], "{rel}");
        // The same proposal that also retires one trial item is back at N.
        let r = result(&w.plan(rel, &[add("new"), retire("old")]));
        assert_eq!(r["new_lines"], json!(budget), "{r}");
        // A file already over the budget stays refused when the proposal does not bring it back.
        filled(&w, rel, section, budget + 1, 5);
        assert_eq!(reasons_of(&w, rel, &[add("new")]), vec!["budget"], "{rel}");
        assert_eq!(
            reasons_of(&w, rel, &[add("new"), retire("old")]),
            vec!["budget"],
            "{rel}"
        );
        let r = result(&w.plan(rel, &[retire("old")]));
        assert_eq!(r["new_lines"], json!(budget), "{r}");
    }
}

#[test]
fn the_description_budget_is_1024_characters() {
    let w = world();
    filled(&w, SKILL, "generation:project-skill:claude:demo", 20, 1024);
    assert_eq!(
        result(&w.plan(SKILL, &[add("new")]))["writes"],
        json!(false)
    );
    filled(&w, SKILL, "generation:project-skill:claude:demo", 20, 1025);
    assert_eq!(reasons_of(&w, SKILL, &[add("new")]), vec!["front_matter"]);
}

#[test]
fn a_target_without_front_matter_with_a_name_and_a_description_is_refused() {
    let w = world();
    for bad in [
        "No front matter.\n",
        "---\nname: demo\n---\nBody.\n",
        "---\ndescription: only\n---\nBody.\n",
        "---\nname: demo\ndescription: unterminated\nBody.\n",
        "\n---\nname: demo\ndescription: not first\n---\n",
        "---\nname:\ndescription: empty name\n---\n",
        "---\nname: demo\ndescription:\n---\n",
    ] {
        w.put(SKILL, bad);
        w.manifest(&[(
            SKILL,
            "generation:project-skill:claude:demo",
            &sha(bad.as_bytes()),
        )]);
        assert_eq!(
            reasons_of(&w, SKILL, &[add("x")]),
            vec!["front_matter"],
            "{bad:?}"
        );
    }
    let not_text: &[u8] = &[0xff, 0xfe, 0x00];
    std::fs::write(w.path(SKILL), not_text).unwrap();
    assert_eq!(reasons_of(&w, SKILL, &[add("x")]), vec!["front_matter"]);
}

#[test]
fn a_rule_that_names_a_missing_path_or_a_missing_command_is_drift() {
    let w = world();
    let skill = format!("{}\nBody.\n", front(5));
    w.owned(&skill, &skill);
    w.put("docs/real.md", "x\n");
    let with = |rule: &str, why: &str| item("add", "d-1", rule, why, &["s"]);
    // An existing path and a command the binary has are not drift.
    assert_eq!(
        result(&w.plan(
            SKILL,
            &[with(
                "Read docs/real.md first.",
                "Run `moai improve-signals` then."
            )]
        ))["writes"],
        json!(false)
    );
    // A word that is only a command-looking word outside backticks is prose.
    assert_eq!(
        result(&w.plan(
            SKILL,
            &[with("Use the moai skill, not moai nonesuch.", "w")]
        ))["writes"],
        json!(false)
    );
    for bad in [
        with("Read docs/missing.md first.", "w"),
        with("r", "See src/gone/file.rs."),
        with("Run `moai nonesuch` first.", "w"),
        with(
            "r",
            "Run `moai card nonesuch-ever` then `moai notacommand`.",
        ),
    ] {
        assert_eq!(
            reasons_of(&w, SKILL, std::slice::from_ref(&bad)),
            vec!["drift"],
            "{bad}"
        );
    }
}

#[test]
fn a_secret_shaped_word_in_an_item_is_refused_and_never_printed() {
    let w = world();
    let skill = format!("{}\nBody.\n", front(5));
    w.owned(&skill, &skill);
    for bad in [
        item("add", "s-1", "Use password=hunter2 here.", "w", &["s"]),
        item("add", "s-1", "r", "token=abc123def456", &["s"]),
        item("add", "s-1", "r", "w", &["password=hunter2"]),
    ] {
        let out = w.plan(SKILL, std::slice::from_ref(&bad));
        let all = text(&out);
        assert!(
            !all.contains("hunter2") && !all.contains("abc123def456"),
            "{all}"
        );
        let (_, _, reasons) = refusal(&out);
        assert_eq!(reasons, vec!["secret"], "{bad}");
    }
}

#[test]
fn every_failing_gate_is_listed_and_the_target_is_never_written() {
    let w = world();
    let skill = format!("{}\nBody.\n", front(5));
    w.owned(&skill, &skill);
    w.put("tests/t.md", "No front matter.\n");
    let h = sha(b"No front matter.\n");
    w.manifest(&[("tests/t.md", "generation:project-skill:claude:demo", &h)]);
    let bad = item(
        "add",
        "m-1",
        "Read docs/none.md with token=abc123def456.",
        "Run `moai nonesuch`.",
        &["s"],
    );
    let reasons = reasons_of(&w, "tests/t.md", &[bad, add("a"), add("b"), add("c")]);
    assert_eq!(
        reasons,
        vec![
            "drift",
            "front_matter",
            "items",
            "not_allowed",
            "secret",
            "test_weakening"
        ]
    );
    assert_eq!(
        std::fs::read(w.path("tests/t.md")).unwrap(),
        b"No front matter.\n"
    );
}

#[test]
fn a_hand_edited_file_shows_conflict_and_a_generated_one_unchanged() {
    let w = world();
    let skill = format!("{}\nBody.\n", front(5));
    w.owned(&skill, &skill);
    assert_eq!(
        result(&w.plan(SKILL, &[add("x")]))["generation"],
        json!("unchanged")
    );
    let edited = format!("{skill}A line a person added.\n");
    w.put(SKILL, &edited);
    let r = result(&w.plan(SKILL, &[add("x")]));
    assert_eq!(r["generation"], json!("conflict"), "{r}");
    assert_eq!(r["old_sha256"], json!(sha(edited.as_bytes())), "{r}");
    // A manifest entry without a baseline reads as a conflict too: nothing is known to match.
    w.manifest(&[(SKILL, "generation:project-skill:claude:demo", "")]);
    assert_eq!(
        result(&w.plan(SKILL, &[add("x")]))["generation"],
        json!("conflict")
    );
}

#[test]
fn the_same_input_gives_the_same_digest_in_any_folder_and_any_change_gives_another() {
    let skill = format!("{}\nBody.\n", front(5));
    let digest = |w: &World, items: &[String]| {
        result(&w.plan(SKILL, items))["plan_digest"]
            .as_str()
            .unwrap()
            .to_string()
    };
    let a = world();
    a.owned(&skill, &skill);
    let b = {
        let home = TempHome::new();
        let project = home.project("another-demo-with-a-longer-name");
        World { home, project }
    };
    b.owned(&skill, &skill);
    let base = digest(&a, &[add("x")]);
    assert_eq!(
        base,
        digest(&b, &[add("x")]),
        "the folder and the clock are not input"
    );
    assert_eq!(
        base,
        digest(&a, &[add("x")]),
        "a second run gives the same digest"
    );
    // A changed item, a changed target byte and a changed baseline each give another digest.
    assert_ne!(
        base,
        digest(
            &a,
            &[item(
                "add",
                "x",
                "Another rule.",
                "A skipped check was claimed.",
                &["failure:cargo"]
            )]
        )
    );
    assert_ne!(base, digest(&a, &[add("y")]));
    let h = sha(skill.as_bytes());
    a.manifest(&[(
        SKILL,
        "generation:project-skill:claude:demo",
        &format!("{h}00"),
    )]);
    assert_ne!(base, digest(&a, &[add("x")]), "the baseline is input");
    let changed = format!("{skill}One more line.\n");
    b.put(SKILL, &changed);
    b.manifest(&[(
        SKILL,
        "generation:project-skill:claude:demo",
        &sha(skill.as_bytes()),
    )]);
    assert_ne!(base, digest(&b, &[add("x")]), "a target byte is input");
}

#[test]
fn a_target_over_the_read_limit_is_refused_by_the_size_budget() {
    let w = world();
    let mut big = front(5);
    big.push_str(&"x".repeat(1024 * 1024));
    w.put(SKILL, &big);
    w.manifest(&[(
        SKILL,
        "generation:project-skill:claude:demo",
        &sha(big.as_bytes()),
    )]);
    assert_eq!(reasons_of(&w, SKILL, &[add("x")]), vec!["budget"]);
}

#[test]
fn the_command_is_listed_in_the_help_and_has_an_example() {
    let w = world();
    let all = w.home.command(&["help", "--all"]).output().unwrap();
    let all = String::from_utf8_lossy(&all.stdout).to_string();
    assert!(all.contains("improve-plan"), "{all}");
    assert!(all.contains("improve-note"), "{all}");
    let help = w
        .home
        .command(&["improve-plan", "--help"])
        .output()
        .unwrap();
    let help = String::from_utf8_lossy(&help.stdout).to_string();
    assert!(
        help.contains("--target") && help.contains("--item"),
        "{help}"
    );
    assert!(Path::new(env!("CARGO_MANIFEST_DIR"))
        .join("docs/design-review/appendix/operations/improve_plan.md")
        .is_file());
}
