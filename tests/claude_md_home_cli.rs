//! The generic moai text lives in the Claude config folder, not in a project
//! (SPEC-20261122-claude-md-home-routing, SPEC-20261122-claude-md-flow-rule): the routing file
//! `rules/moai.md` and the flow rule file `rules/moai/flow.md` are planned together for each
//! config folder, shown in the preview and written only with the apply of the previewed plan;
//! extra config folders and the removal of the earlier `moai-routing` block in the user's
//! `CLAUDE.md` are separate picks. Every command runs through the real binary with a clean
//! environment, a temporary HOME and `MOAI_ACCOUNT_HOME` naming it, so the real HOME and the
//! real config folder are never read or written.
#![cfg(unix)]

mod common;

use common::home::TempHome;
use serde_json::{json, Value};
use std::ffi::OsString;
use std::fs;
use std::os::unix::fs::symlink;
use std::path::{Path, PathBuf};
use std::process::Output;

const PERM: &str = "test-permission";
const BEGIN: &str = "<!-- moai:managed:begin id=moai-routing -->\n";
const END: &str = "<!-- moai:managed:end id=moai-routing -->\n";

struct World {
    home: TempHome,
    base: PathBuf,
    pkg: PathBuf,
    vars: Vec<(String, OsString)>,
}

impl World {
    fn new() -> Self {
        let home = TempHome::new();
        let base = fs::canonicalize(home.path()).unwrap();
        World {
            home,
            base,
            pkg: Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai"),
            vars: Vec::new(),
        }
    }

    fn var(mut self, k: &str, v: impl Into<OsString>) -> Self {
        self.vars.push((k.into(), v.into()));
        self
    }

    fn package(mut self, p: &Path) -> Self {
        self.pkg = p.to_path_buf();
        self
    }

    fn h(&self) -> String {
        self.base.to_str().unwrap().to_string()
    }

    fn run(&self, args: &[&str]) -> Output {
        let mut cmd = self.home.command(args);
        cmd.env("HOME", &self.base)
            .env("MOAI_ACCOUNT_HOME", &self.base);
        for (k, v) in &self.vars {
            cmd.env(k, v);
        }
        cmd.output().expect("run moai")
    }

    /// The preview: `home-routing-plan HOME PACKAGE <extra>`.
    fn plan(&self, extra: &[&str]) -> Value {
        let h = self.h();
        let pkg = self.pkg.to_str().unwrap().to_string();
        let mut args = vec!["home-routing-plan", h.as_str(), pkg.as_str()];
        args.extend_from_slice(extra);
        ok(&self.run(&args))
    }

    fn plan_refused(&self, extra: &[&str]) -> String {
        let h = self.h();
        let pkg = self.pkg.to_str().unwrap().to_string();
        let mut args = vec!["home-routing-plan", h.as_str(), pkg.as_str()];
        args.extend_from_slice(extra);
        refused(&self.run(&args))
    }

    fn apply_out(&self, digest: &str, extra: &[&str]) -> Output {
        let h = self.h();
        let pkg = self.pkg.to_str().unwrap().to_string();
        let mut args = vec![
            "home-routing-apply",
            h.as_str(),
            pkg.as_str(),
            "--plan-digest",
            digest,
            "--permission-reference",
            PERM,
        ];
        args.extend_from_slice(extra);
        self.run(&args)
    }

    /// Previews with `extra`, applies that preview with the same `extra`, returns the result.
    fn apply(&self, extra: &[&str]) -> Value {
        let plan = self.plan(extra);
        ok(&self.apply_out(plan["plan_digest"].as_str().unwrap(), extra))
    }

    fn revert_out(&self, id: &str) -> Output {
        let h = self.h();
        self.run(&[
            "home-routing-revert",
            h.as_str(),
            "--receipt",
            id,
            "--permission-reference",
            PERM,
        ])
    }

    fn revert(&self, id: &str) -> Value {
        ok(&self.revert_out(id))
    }

    fn default_dir(&self) -> PathBuf {
        self.base.join(".claude")
    }

    fn state_dir(&self) -> PathBuf {
        moai_cli::home_root::path(&self.base).join("home-routing")
    }

    /// The names directly inside `dir` (a folder of this temporary home), sorted, from the
    /// home snapshot; a missing folder has none.
    fn children(&self, dir: &Path) -> Vec<String> {
        let rel = dir.strip_prefix(&self.base).unwrap().to_str().unwrap();
        let prefix = format!("{rel}/");
        let mut names: Vec<String> = self
            .home
            .snapshot()
            .into_iter()
            .filter_map(|e| {
                let rest = e.path.strip_prefix(&prefix)?.to_string();
                (!rest.contains('/')).then_some(rest)
            })
            .collect();
        names.sort();
        names
    }

    fn rules_state_files(&self) -> Vec<String> {
        let mut names: Vec<String> = self
            .children(&self.state_dir())
            .into_iter()
            .filter(|n| n.starts_with("rules-"))
            .collect();
        names.sort();
        names
    }
}

fn ok(out: &Output) -> Value {
    let v: Value = serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
        panic!(
            "stdout is not JSON ({e}): {} {}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    });
    assert_eq!(out.status.code(), Some(0), "{v}");
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn refused(out: &Output) -> String {
    let v: Value = serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
        panic!(
            "stdout is not JSON ({e}): {} {}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    });
    assert_eq!(out.status.code(), Some(1), "{v}");
    v["error"]["message"].as_str().unwrap().to_string()
}

fn s(p: &Path) -> String {
    p.to_str().unwrap().to_string()
}

fn row<'a>(plan: &'a Value, dir: &Path) -> &'a Value {
    plan["hosts"]
        .as_array()
        .unwrap()
        .iter()
        .find(|h| h["config_dir"] == json!(s(dir)))
        .unwrap_or_else(|| panic!("no row for {}: {plan}", dir.display()))
}

fn routing_action(plan: &Value, dir: &Path) -> String {
    row(plan, dir)["section"]["action"]
        .as_str()
        .unwrap_or("none")
        .to_string()
}

fn flow_action(plan: &Value, dir: &Path) -> String {
    row(plan, dir)["flow"]["action"]
        .as_str()
        .unwrap_or("none")
        .to_string()
}

fn write(path: &Path, text: impl AsRef<[u8]>) {
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, text).unwrap();
}

fn shipped(rel: &str) -> String {
    fs::read_to_string(
        Path::new(env!("CARGO_MANIFEST_DIR"))
            .join("skills/moai")
            .join(rel),
    )
    .unwrap()
}

fn shipped_routing_text() -> String {
    shipped("references/home/routing.md").replacen("{invoke}", "/moai", 1)
}

/// A package holding only what the plan reads.
fn temp_pkg(routing: &str, flow: Option<&str>) -> tempfile::TempDir {
    let pkg = tempfile::tempdir().unwrap();
    write(&pkg.path().join("SKILL.md"), "package\n");
    write(&pkg.path().join("references/home/routing.md"), routing);
    if let Some(f) = flow {
        write(&pkg.path().join("references/home/flow.md"), f);
    }
    pkg
}

const ROUTING_V1: &str = "## MoAI\n\nUse the skill ({invoke}).\n";
const ROUTING_V2: &str = "## MoAI\n\nUse the moai skill ({invoke}) first.\n";

/// `lines` lines, the first holding `{invoke}` once.
fn routing_of(lines: usize) -> String {
    let mut t = String::from("use ({invoke})\n");
    for _ in 1..lines {
        t.push_str("x\n");
    }
    t
}

fn flow_of(lines: usize) -> String {
    "rule\n".repeat(lines)
}

/// The earlier version's block and its recorded baseline in `<state>/home-routing/claude.json`.
fn old_block() -> String {
    format!("{BEGIN}## MoAI\n\nold text ({})\n{END}", "/moai")
}

fn write_old_state(w: &World, block: &str) {
    write(
        &w.state_dir().join("claude.json"),
        json!({"schema_version": 1, "block_sha256": moai_cli::digest(block.as_bytes())})
            .to_string(),
    );
}

fn claude_md_with_block(w: &World) -> (PathBuf, String) {
    let block = old_block();
    let text = format!("# Mine\n\nKeep this.\n\n{block}\nAfter.\n");
    let path = w.default_dir().join("CLAUDE.md");
    write(&path, &text);
    write_old_state(w, &block);
    (path, text)
}

// ---- flow_plan_ ----------------------------------------------------------------------------

#[test]
fn flow_plan_default_folder_gets_both_files_and_a_second_run_is_unchanged() {
    let w = World::new();
    let dir = w.default_dir();
    let plan = w.plan(&[]);
    assert_eq!(plan["hosts"].as_array().unwrap().len(), 1, "{plan}");
    assert_eq!(routing_action(&plan, &dir), "create", "{plan}");
    assert_eq!(flow_action(&plan, &dir), "create", "{plan}");
    assert_eq!(
        row(&plan, &dir)["file"],
        json!(s(&dir.join("rules/moai.md")))
    );
    assert_eq!(
        row(&plan, &dir)["flow"]["path"],
        json!(s(&dir.join("rules/moai/flow.md")))
    );
    // The preview writes nothing.
    assert!(!dir.exists());
    let applied = ok(&w.apply_out(plan["plan_digest"].as_str().unwrap(), &[]));
    assert_eq!(applied["state"], json!("applied"), "{applied}");
    assert_eq!(
        fs::read_to_string(dir.join("rules/moai.md")).unwrap(),
        shipped_routing_text()
    );
    let flow = fs::read_to_string(dir.join("rules/moai/flow.md")).unwrap();
    assert_eq!(flow, shipped("references/home/flow.md"));
    assert!(!dir.join("CLAUDE.md").exists());
    let again = w.plan(&[]);
    assert_eq!(routing_action(&again, &dir), "unchanged", "{again}");
    assert_eq!(flow_action(&again, &dir), "unchanged", "{again}");
    let r = ok(&w.apply_out(again["plan_digest"].as_str().unwrap(), &[]));
    assert_eq!(r["state"], json!("unchanged"), "{r}");
}

#[test]
fn flow_plan_text_has_no_paths_key_and_stays_within_its_limits() {
    let w = World::new();
    w.apply(&[]);
    let routing = fs::read_to_string(w.default_dir().join("rules/moai.md")).unwrap();
    let flow = fs::read(w.default_dir().join("rules/moai/flow.md")).unwrap();
    let flow = String::from_utf8(flow).unwrap();
    assert!(routing.lines().count() <= 12, "{routing}");
    assert!(flow.lines().count() <= 80);
    assert!(flow.len() <= 5 * 1024);
    for text in [&routing, &flow] {
        assert!(!text.starts_with("---"), "no frontmatter");
        assert!(!text.lines().any(|l| l.starts_with("paths:")));
    }
}

#[test]
fn flow_plan_only_the_default_folder_gets_the_flow_file_and_the_state() {
    // The default folder is the one copy; each picked folder is a row that writes nothing.
    let w = World::new();
    let a = w.base.join("extra/a");
    let b = w.base.join("extra/b");
    let extra = [
        "--config-dir",
        a.to_str().unwrap(),
        "--config-dir",
        b.to_str().unwrap(),
    ];
    let plan = w.plan(&extra);
    assert_eq!(plan["hosts"].as_array().unwrap().len(), 3, "{plan}");
    assert_eq!(flow_action(&plan, &w.default_dir()), "create", "{plan}");
    for d in [&a, &b] {
        assert_eq!(flow_action(&plan, d), "not_linked", "{plan}");
    }
    let r = ok(&w.apply_out(plan["plan_digest"].as_str().unwrap(), &extra));
    assert_eq!(r["state"], json!("applied"), "{r}");
    assert!(w.default_dir().join("rules/moai/flow.md").is_file());
    assert!(w.default_dir().join("rules/moai.md").is_file());
    for d in [&a, &b] {
        assert!(!d.exists(), "nothing is written in {}", d.display());
    }
    assert_eq!(
        w.rules_state_files().len(),
        1,
        "{:?}",
        w.rules_state_files()
    );
}

#[test]
fn flow_plan_applies_only_the_previewed_plan() {
    let w = World::new();
    let plan = w.plan(&[]);
    let digest = plan["plan_digest"].as_str().unwrap().to_string();
    // A digest of another plan, and a plan that went stale, are refused; nothing is written.
    let m = refused(&w.apply_out("sha256:0000", &[]));
    assert!(m.contains("stale"), "{m}");
    write(&w.default_dir().join("rules/moai.md"), "someone else's\n");
    let m = refused(&w.apply_out(&digest, &[]));
    assert!(m.contains("stale"), "{m}");
    assert!(!w.default_dir().join("rules/moai/flow.md").exists());
    // The picks are part of the preview: a plan made without a folder is not that of a run with it.
    let a = w.base.join("extra/a");
    let m = refused(&w.apply_out(&digest, &["--config-dir", a.to_str().unwrap()]));
    assert!(m.contains("stale"), "{m}");
    assert!(!a.exists());
}

#[test]
fn flow_plan_a_source_over_a_limit_is_refused_naming_the_limit() {
    // Each limit at N (passes) and N+1 (refused), and the empty file.
    let case = |routing: &str, flow: Option<&str>| {
        let pkg = temp_pkg(routing, flow);
        let w2 = World::new().package(pkg.path());
        let h = w2.h();
        let p = s(pkg.path());
        let out = w2.run(&["home-routing-plan", h.as_str(), p.as_str()]);
        let v: Value = serde_json::from_slice(&out.stdout).unwrap();
        if out.status.code() == Some(0) {
            Ok(v["result"].clone())
        } else {
            Err(v["error"]["message"].as_str().unwrap().to_string())
        }
    };
    assert!(case(&routing_of(12), None).is_ok());
    let e = case(&routing_of(13), None).unwrap_err();
    assert!(e.contains("12 lines"), "{e}");
    assert!(case(ROUTING_V1, Some(&flow_of(80))).is_ok());
    let e = case(ROUTING_V1, Some(&flow_of(81))).unwrap_err();
    assert!(e.contains("80 lines"), "{e}");
    let at_limit = format!("{}\n", "a".repeat(5 * 1024 - 1));
    assert_eq!(at_limit.len(), 5 * 1024);
    assert!(case(ROUTING_V1, Some(&at_limit)).is_ok());
    let over = format!("{}\n", "a".repeat(5 * 1024));
    let e = case(ROUTING_V1, Some(&over)).unwrap_err();
    assert!(e.contains("5 KiB"), "{e}");
    let e = case(ROUTING_V1, Some("")).unwrap_err();
    assert!(e.contains("newline"), "{e}");
    // A paths key (frontmatter) is refused in either source.
    let e = case(ROUTING_V1, Some("---\npaths: [\"*.rs\"]\n---\nrule\n")).unwrap_err();
    assert!(e.contains("paths"), "{e}");
    let e = case("---\npaths: x\n---\nuse ({invoke})\n", None).unwrap_err();
    assert!(e.contains("paths"), "{e}");
}

#[test]
fn flow_plan_a_package_without_a_flow_file_plans_none() {
    let pkg = temp_pkg(ROUTING_V1, None);
    let w = World::new().package(pkg.path());
    let plan = w.plan(&[]);
    assert_eq!(routing_action(&plan, &w.default_dir()), "create", "{plan}");
    assert_eq!(flow_action(&plan, &w.default_dir()), "none", "{plan}");
    w.apply(&[]);
    assert!(w.default_dir().join("rules/moai.md").is_file());
    assert!(!w.default_dir().join("rules/moai").exists());
}

#[test]
fn flow_plan_a_flow_only_change_still_counts_as_a_write_on_the_routing_row() {
    // `update` asks the routing row alone whether the layer writes anything.
    let v1 = temp_pkg(ROUTING_V1, Some("one\n"));
    let w = World::new().package(v1.path());
    w.apply(&[]);
    let v2 = temp_pkg(ROUTING_V1, Some("two\n"));
    let w = w.package(v2.path());
    let plan = w.plan(&[]);
    let dir = w.default_dir();
    assert_eq!(flow_action(&plan, &dir), "merge", "{plan}");
    assert_eq!(routing_action(&plan, &dir), "merge", "{plan}");
}

// ---- flow_conflict_ ------------------------------------------------------------------------

#[test]
fn flow_conflict_a_hand_edited_file_keeps_its_bytes() {
    let w = World::new();
    w.apply(&[]);
    let flow = w.default_dir().join("rules/moai/flow.md");
    write(&flow, "my own rule\n");
    let plan = w.plan(&[]);
    assert_eq!(flow_action(&plan, &w.default_dir()), "conflict", "{plan}");
    let r = ok(&w.apply_out(plan["plan_digest"].as_str().unwrap(), &[]));
    assert_eq!(r["state"], json!("unchanged"), "{r}");
    assert_eq!(fs::read_to_string(&flow).unwrap(), "my own rule\n");
    // A file someone else wrote before moai ever did is a conflict too.
    let w = World::new();
    write(&w.default_dir().join("rules/moai/flow.md"), "theirs\n");
    let plan = w.plan(&[]);
    assert_eq!(flow_action(&plan, &w.default_dir()), "conflict", "{plan}");
    assert_eq!(routing_action(&plan, &w.default_dir()), "create", "{plan}");
}

#[test]
fn flow_conflict_a_linked_file_or_folder_is_kept_and_nothing_is_written_through_it() {
    // The file, `rules` and `rules/moai` are links; the config folder is a link.
    for what in ["file", "rules", "moai", "config"] {
        let w = World::new();
        let away = w.base.join("away");
        fs::create_dir_all(&away).unwrap();
        let dir = w.default_dir();
        match what {
            "file" => {
                fs::create_dir_all(dir.join("rules/moai")).unwrap();
                write(&away.join("flow.md"), "theirs\n");
                symlink(away.join("flow.md"), dir.join("rules/moai/flow.md")).unwrap();
            }
            "rules" => {
                fs::create_dir_all(&dir).unwrap();
                symlink(&away, dir.join("rules")).unwrap();
            }
            "moai" => {
                fs::create_dir_all(dir.join("rules")).unwrap();
                symlink(&away, dir.join("rules/moai")).unwrap();
            }
            _ => symlink(&away, &dir).unwrap(),
        }
        let before = w.home.snapshot();
        let plan = w.plan(&[]);
        let flow = &row(&plan, &dir)["flow"];
        assert_eq!(row(&plan, &dir)["preflight"], json!("ok"), "{what}: {plan}");
        assert_eq!(flow["action"], json!("conflict"), "{what}: {plan}");
        assert!(
            flow["detail"].as_str().unwrap().contains("link"),
            "{what}: {plan}"
        );
        let r = ok(&w.apply_out(plan["plan_digest"].as_str().unwrap(), &[]));
        // The routing file may still be written where no link is on its way.
        let _ = r;
        let after = w.home.snapshot();
        let flow_written = after
            .iter()
            .any(|e| e.path.ends_with("flow.md") && !before.iter().any(|b| b.path == e.path));
        assert!(!flow_written, "{what}: the flow file is not written");
        assert_eq!(
            w.children(&away).len(),
            usize::from(what == "file"),
            "{what}: nothing is written through the link"
        );
    }
}

#[test]
fn flow_conflict_a_file_that_is_not_utf8_is_kept() {
    let w = World::new();
    let flow = w.default_dir().join("rules/moai/flow.md");
    write(&flow, [0xff, 0xfe, 0x00, b'x']);
    let plan = w.plan(&[]);
    let f = &row(&plan, &w.default_dir())["flow"];
    assert_eq!(f["action"], json!("conflict"), "{plan}");
    assert!(f["detail"].as_str().unwrap().contains("UTF-8"), "{plan}");
    w.apply_out(plan["plan_digest"].as_str().unwrap(), &[]);
    assert_eq!(fs::read(&flow).unwrap(), [0xff, 0xfe, 0x00, b'x']);
}

#[test]
fn flow_conflict_a_new_source_merges_while_the_file_is_what_moai_wrote() {
    let v1 = temp_pkg(ROUTING_V1, Some("first\n"));
    let w = World::new().package(v1.path());
    w.apply(&[]);
    let v2 = temp_pkg(ROUTING_V1, Some("second\n"));
    let w = w.package(v2.path());
    let plan = w.plan(&[]);
    assert_eq!(flow_action(&plan, &w.default_dir()), "merge", "{plan}");
    w.apply(&[]);
    assert_eq!(
        fs::read_to_string(w.default_dir().join("rules/moai/flow.md")).unwrap(),
        "second\n"
    );
    // An edit after that turns the next new source into a conflict.
    write(&w.default_dir().join("rules/moai/flow.md"), "edited\n");
    let v3 = temp_pkg(ROUTING_V1, Some("third\n"));
    let w = w.package(v3.path());
    assert_eq!(flow_action(&w.plan(&[]), &w.default_dir()), "conflict");
}

// ---- flow_text_ ----------------------------------------------------------------------------

/// The one sentence both shipped rules files open with: they bind only a project configured for
/// MoAI or a session where the user invokes MoAI, so a project that has nothing to do with MoAI
/// keeps its own workflow although the user-level file is loaded in every session.
const CONDITION: &str = "only in a project configured for MoAI";

#[test]
fn flow_text_holds_the_evidence_and_handoff_rule() {
    let text = shipped("references/home/flow.md");
    for want in [
        "Missing evidence is never a PASS",
        "record the tested revision, the exact checks, and any check that was skipped or failed",
        "Keep implementation and independent review distinct",
        "a reviewer checks the candidate revision, not a stale one",
        "When blocked, record the scope, the revision, the remaining checks",
        "never bypass ownership, permissions or a release gate",
        "Keep secrets and private data out of shared artifacts",
        "a local merge does not authorize a push or a release",
        "never remove an absolute or home path",
    ] {
        assert!(text.contains(want), "the rule lacks {want:?}");
    }
    // The project's own CLAUDE.md wins: said once.
    let wins: Vec<&str> = text
        .lines()
        .filter(|l| l.contains("CLAUDE.md wins"))
        .collect();
    assert_eq!(wins.len(), 1, "{wins:?}");
    // Exactly one pointer to the full flow, which lives in the skill's flow reference.
    let pointer: Vec<&str> = text
        .lines()
        .filter(|l| l.contains("references/flows/next-flow.md"))
        .collect();
    assert_eq!(pointer.len(), 1, "{pointer:?}");
}

#[test]
fn flow_text_binds_only_a_project_configured_for_moai() {
    for rel in ["references/home/flow.md", "references/home/routing.md"] {
        let text = shipped(rel);
        let first_rule = text
            .lines()
            .find(|l| l.contains(CONDITION))
            .unwrap_or_else(|| panic!("{rel} lacks the condition {CONDITION:?}"));
        // The condition is the first sentence of prose, ahead of every instruction.
        let before: Vec<&str> = text
            .lines()
            .take_while(|l| !l.contains(CONDITION))
            .filter(|l| !l.trim().is_empty() && !l.starts_with('#'))
            .collect();
        assert!(
            before.is_empty(),
            "{rel}: rules ahead of the condition: {before:?}"
        );
        assert!(!first_rule.trim().is_empty());
        // REQ-6 names both ways in: a configured project, or the user invoking MoAI.
        assert!(
            first_rule.contains("or when the user invokes MoAI"),
            "{rel}: the condition is narrower than REQ-6: {first_rule:?}"
        );
    }
    assert!(
        shipped("references/home/routing.md").contains("never changes another project's workflow"),
        "the routing text says what loading it does not do"
    );
}

#[test]
fn flow_text_carries_no_role_or_command_detail() {
    // Lane roles, card commands and the landing command are read from the skill when a card or
    // a landing is involved; hooks enforce what they can (force push, reset, clean). Neither
    // shipped rules file repeats any of it, because the user-level file loads in every project.
    for rel in ["references/home/flow.md", "references/home/routing.md"] {
        let text = shipped(rel);
        for banned in [
            "moai card",
            "LANE-",
            "--by leader",
            "lander",
            ".claude/worktrees",
            "integration branch",
            "Compact Instructions",
            "TaskCreate",
            "reset --hard",
            "--no-verify",
            "add -A",
        ] {
            assert!(!text.contains(banned), "{rel} carries {banned:?}");
        }
    }
}

#[test]
fn flow_text_names_neither_the_retired_name_nor_the_four_letters() {
    let text = shipped("references/home/flow.md");
    let lower = text.to_lowercase();
    assert!(!lower.contains("m-flow"), "the retired name");
    for letters in [
        "no tip chase",
        "evidence before landing",
        "exactly-once",
        "tip guarded",
        "four letters",
    ] {
        assert!(!lower.contains(letters), "{letters}");
    }
    // No card number in a shipped text.
    for t in [&text, &shipped("references/home/routing.md")] {
        let bytes = t.as_bytes();
        let card = bytes
            .windows(3)
            .any(|w| w[0] == b'C' && w[1] == b'-' && w[2].is_ascii_digit());
        assert!(!card, "a card number in a shipped text");
    }
    assert!(text.lines().count() <= 80 && text.len() <= 5 * 1024);
    assert!(!text.starts_with("---"));
}

// ---- target_ -------------------------------------------------------------------------------

#[test]
fn target_the_routing_file_lands_in_the_default_folder_and_claude_md_stays() {
    let w = World::new();
    write(&w.default_dir().join("CLAUDE.md"), "mine\n");
    let plan = w.plan(&[]);
    let h = row(&plan, &w.default_dir());
    assert_eq!(h["file"], json!(s(&w.default_dir().join("rules/moai.md"))));
    w.apply(&[]);
    assert_eq!(
        fs::read_to_string(w.default_dir().join("rules/moai.md")).unwrap(),
        shipped_routing_text()
    );
    assert_eq!(
        fs::read_to_string(w.default_dir().join("CLAUDE.md")).unwrap(),
        "mine\n"
    );
}

#[test]
fn target_claude_config_dir_never_moves_the_one_copy_out_of_the_default_folder() {
    // (b) A session whose CLAUDE_CONFIG_DIR names a profile: the default folder gets both files
    // and the profile gets nothing (the user's decision: `~/.claude` is the only source).
    let w0 = World::new();
    let conf = w0.base.join("elsewhere/conf");
    write(&conf.join("CLAUDE.md"), "theirs\n");
    let w = w0.var("CLAUDE_CONFIG_DIR", conf.as_os_str());
    let plan = w.plan(&[]);
    assert_eq!(plan["hosts"].as_array().unwrap().len(), 2, "{plan}");
    assert_eq!(
        plan["hosts"][0]["config_dir"],
        json!(s(&w.default_dir())),
        "the default folder is always the first row: {plan}"
    );
    assert_eq!(
        row(&plan, &w.default_dir())["file"],
        json!(s(&w.default_dir().join("rules/moai.md")))
    );
    assert_eq!(routing_action(&plan, &w.default_dir()), "create", "{plan}");
    assert_eq!(routing_action(&plan, &conf), "not_linked", "{plan}");
    assert_eq!(flow_action(&plan, &conf), "not_linked", "{plan}");
    // The row tells the user how the profile reads the one copy; moai makes no link.
    let detail = row(&plan, &conf)["section"]["detail"].as_str().unwrap();
    assert!(
        detail.contains(&s(&conf.join("rules")))
            && detail.contains(&s(&w.default_dir().join("rules"))),
        "{detail}"
    );
    let before = w.home.snapshot();
    w.apply(&[]);
    assert!(w.default_dir().join("rules/moai.md").is_file());
    assert!(w.default_dir().join("rules/moai/flow.md").is_file());
    assert!(
        !conf.join("rules").exists(),
        "no copy and no link in the profile"
    );
    assert_eq!(
        fs::read_to_string(conf.join("CLAUDE.md")).unwrap(),
        "theirs\n"
    );
    assert_ne!(w.home.snapshot(), before);
    // The state belongs to the default folder, so a plain session later sees nothing to do.
    assert_eq!(
        w.rules_state_files().len(),
        1,
        "{:?}",
        w.rules_state_files()
    );
    let plain = World {
        vars: Vec::new(),
        home: w.home,
        base: w.base.clone(),
        pkg: w.pkg.clone(),
    };
    let again = plain.plan(&[]);
    assert_eq!(
        routing_action(&again, &plain.default_dir()),
        "unchanged",
        "{again}"
    );
    assert_eq!(
        flow_action(&again, &plain.default_dir()),
        "unchanged",
        "{again}"
    );
}

#[test]
fn target_a_refused_variable_blocks_the_layer_and_names_it_without_its_value() {
    for (value, says) in [("", "set but empty"), ("secretword/conf", "absolute")] {
        let w = World::new().var("CLAUDE_CONFIG_DIR", value);
        let plan = w.plan(&[]);
        let hosts = plan["hosts"].as_array().unwrap();
        assert_eq!(hosts.len(), 1, "{plan}");
        assert_eq!(hosts[0]["preflight"], json!("blocked"), "{plan}");
        let why = hosts[0]["blocked_reason"].as_str().unwrap();
        assert!(
            why.contains("CLAUDE_CONFIG_DIR") && why.contains(says),
            "{why}"
        );
        assert!(!why.contains("secretword"), "{why}");
        assert!(hosts[0]["section"].is_null() && hosts[0]["flow"].is_null());
        let r = ok(&w.apply_out(plan["plan_digest"].as_str().unwrap(), &[]));
        assert_eq!(r["state"], json!("unchanged"), "{r}");
        assert!(!w.default_dir().exists());
    }
}

#[test]
fn target_an_edited_routing_file_is_a_conflict() {
    let w = World::new();
    w.apply(&[]);
    let file = w.default_dir().join("rules/moai.md");
    write(&file, "my routing\n");
    let plan = w.plan(&[]);
    assert_eq!(
        routing_action(&plan, &w.default_dir()),
        "conflict",
        "{plan}"
    );
    w.apply_out(plan["plan_digest"].as_str().unwrap(), &[]);
    assert_eq!(fs::read_to_string(&file).unwrap(), "my routing\n");
}

#[test]
fn target_a_new_source_merges_while_the_file_is_what_moai_wrote() {
    let v1 = temp_pkg(ROUTING_V1, None);
    let w = World::new().package(v1.path());
    w.apply(&[]);
    let v2 = temp_pkg(ROUTING_V2, None);
    let w = w.package(v2.path());
    assert_eq!(routing_action(&w.plan(&[]), &w.default_dir()), "merge");
    w.apply(&[]);
    assert!(fs::read_to_string(w.default_dir().join("rules/moai.md"))
        .unwrap()
        .contains("first."));
}

#[test]
fn target_a_link_or_a_file_that_is_not_utf8_is_a_conflict_not_a_block() {
    let w = World::new();
    let away = w.base.join("away/moai.md");
    write(&away, "theirs\n");
    fs::create_dir_all(w.default_dir().join("rules")).unwrap();
    symlink(&away, w.default_dir().join("rules/moai.md")).unwrap();
    let plan = w.plan(&[]);
    assert_eq!(
        row(&plan, &w.default_dir())["preflight"],
        json!("ok"),
        "{plan}"
    );
    assert_eq!(
        routing_action(&plan, &w.default_dir()),
        "conflict",
        "{plan}"
    );
    w.apply_out(plan["plan_digest"].as_str().unwrap(), &[]);
    assert_eq!(fs::read_to_string(&away).unwrap(), "theirs\n");

    let w = World::new();
    write(&w.default_dir().join("rules/moai.md"), [0xc3, 0x28]);
    let plan = w.plan(&[]);
    assert_eq!(
        routing_action(&plan, &w.default_dir()),
        "conflict",
        "{plan}"
    );
    assert_eq!(
        fs::read(w.default_dir().join("rules/moai.md")).unwrap(),
        [0xc3, 0x28]
    );
}

#[test]
fn target_a_profile_rules_link_to_the_default_folder_is_linked_and_any_other_is_not() {
    let w0 = World::new();
    let conf = w0.base.join("elsewhere/conf");
    let real = w0.default_dir().join("rules");
    fs::create_dir_all(&real).unwrap();
    fs::create_dir_all(&conf).unwrap();
    symlink(&real, conf.join("rules")).unwrap();
    let w = w0.var("CLAUDE_CONFIG_DIR", conf.as_os_str());
    let plan = w.plan(&[]);
    assert_eq!(routing_action(&plan, &w.default_dir()), "create", "{plan}");
    assert_eq!(routing_action(&plan, &conf), "linked", "{plan}");
    assert_eq!(flow_action(&plan, &conf), "linked", "{plan}");
    w.apply(&[]);
    assert!(real.join("moai.md").is_file());
    assert!(real.join("moai/flow.md").is_file());
    assert!(fs::symlink_metadata(conf.join("rules"))
        .unwrap()
        .is_symlink());

    // A link to any other folder is not a reference: the row names the link and writes nothing.
    let w0 = World::new();
    let conf = w0.base.join("elsewhere/conf2");
    let other = w0.base.join("elsewhere/other");
    fs::create_dir_all(&conf).unwrap();
    fs::create_dir_all(&other).unwrap();
    symlink(&other, conf.join("rules")).unwrap();
    let w = w0.var("CLAUDE_CONFIG_DIR", conf.as_os_str());
    let plan = w.plan(&[]);
    assert_eq!(routing_action(&plan, &conf), "not_linked", "{plan}");
    assert!(
        row(&plan, &conf)["section"]["detail"]
            .as_str()
            .unwrap()
            .contains(&s(&conf.join("rules"))),
        "{plan}"
    );
    w.apply(&[]);
    assert!(w.children(&other).is_empty());
    assert!(w.default_dir().join("rules/moai.md").is_file());
}

#[test]
fn target_a_profile_link_whose_target_is_not_there_yet_is_still_linked() {
    // (c2) The profile's `rules` points at `~/.claude/rules`, which no run has created yet: the
    // link is the right one, so the row says `linked`, never "a link to another folder".
    let w0 = World::new();
    let conf = w0.base.join("elsewhere/conf");
    fs::create_dir_all(&conf).unwrap();
    symlink(w0.default_dir().join("rules"), conf.join("rules")).unwrap();
    assert!(!w0.default_dir().exists());
    let w = w0.var("CLAUDE_CONFIG_DIR", conf.as_os_str());
    let plan = w.plan(&[]);
    assert_eq!(routing_action(&plan, &conf), "linked", "{plan}");
    assert_eq!(flow_action(&plan, &conf), "linked", "{plan}");
    let detail = row(&plan, &conf)["section"]["detail"].as_str().unwrap();
    assert!(!detail.contains("other than"), "{detail}");
    w.apply(&[]);
    assert!(w.default_dir().join("rules/moai.md").is_file());
    // The link now reaches the file, and a second preview has nothing left to write.
    assert!(conf.join("rules/moai.md").is_file());
    let again = w.plan(&[]);
    assert_eq!(
        routing_action(&again, &w.default_dir()),
        "unchanged",
        "{again}"
    );
    assert_eq!(routing_action(&again, &conf), "linked", "{again}");
}

#[test]
fn target_a_relative_profile_link_is_read_from_the_profile_folder() {
    // The profile's `rules` is the relative link `../.claude/rules`: it is read from the folder
    // the link sits in, so it reaches the default folder's `rules` and the rows say `linked`,
    // whatever folder the command runs in.
    let w0 = World::new();
    let conf = w0.base.join("conf");
    fs::create_dir_all(&conf).unwrap();
    symlink("../.claude/rules", conf.join("rules")).unwrap();
    let w = w0.var("CLAUDE_CONFIG_DIR", conf.as_os_str());
    let plan = w.plan(&[]);
    assert_eq!(routing_action(&plan, &w.default_dir()), "create", "{plan}");
    assert_eq!(routing_action(&plan, &conf), "linked", "{plan}");
    assert_eq!(flow_action(&plan, &conf), "linked", "{plan}");
    w.apply(&[]);
    assert!(w.default_dir().join("rules/moai.md").is_file());
    // Once the target exists the link still reaches it, and nothing is written into the profile.
    assert!(conf.join("rules/moai.md").is_file());
    let again = w.plan(&[]);
    assert_eq!(routing_action(&again, &conf), "linked", "{again}");

    // A relative link that leaves for another folder is still not a reference.
    let w0 = World::new();
    let conf = w0.base.join("conf");
    fs::create_dir_all(&conf).unwrap();
    fs::create_dir_all(w0.base.join("other")).unwrap();
    symlink("../other", conf.join("rules")).unwrap();
    let w = w0.var("CLAUDE_CONFIG_DIR", conf.as_os_str());
    let plan = w.plan(&[]);
    assert_eq!(routing_action(&plan, &conf), "not_linked", "{plan}");
}

#[test]
fn target_a_profile_with_its_own_rules_folder_is_not_linked_and_keeps_every_byte() {
    let w0 = World::new();
    let conf = w0.base.join("elsewhere/conf");
    write(&conf.join("rules/mine.md"), "mine\n");
    let w = w0.var("CLAUDE_CONFIG_DIR", conf.as_os_str());
    let plan = w.plan(&[]);
    assert_eq!(routing_action(&plan, &conf), "not_linked", "{plan}");
    w.apply(&[]);
    assert_eq!(w.children(&conf.join("rules")), ["mine.md"]);
    assert_eq!(
        fs::read_to_string(conf.join("rules/mine.md")).unwrap(),
        "mine\n"
    );
}

#[test]
fn target_a_refused_variable_still_leaves_a_blocked_default_row_and_no_consumer() {
    let w = World::new().var("CLAUDE_CONFIG_DIR", "relative/dir");
    let extra = w.base.join("extra/a");
    let plan = w.plan(&["--config-dir", extra.to_str().unwrap()]);
    for h in plan["hosts"].as_array().unwrap() {
        assert!(h["section"].is_null(), "{plan}");
        assert!(h["flow"].is_null(), "{plan}");
    }
    assert_eq!(
        plan["hosts"][0]["config_dir"],
        json!(s(&w.default_dir())),
        "{plan}"
    );
    assert_eq!(plan["hosts"][0]["preflight"], json!("blocked"), "{plan}");
    let r = ok(&w.apply_out(
        plan["plan_digest"].as_str().unwrap(),
        &["--config-dir", extra.to_str().unwrap()],
    ));
    assert_eq!(r["state"], json!("unchanged"), "{r}");
    assert!(!w.default_dir().exists() && !extra.exists());
}

#[test]
fn target_a_conflict_in_the_routing_file_holds_the_flow_file_and_the_preview_says_so() {
    // The `update` preview reads the routing row only; a flow write behind a routing conflict
    // would make "nothing to apply" untrue, so the flow file waits until the conflict is gone.
    let w = World::new();
    write(&w.default_dir().join("rules/moai.md"), "my routing\n");
    let plan = w.plan(&[]);
    assert_eq!(
        routing_action(&plan, &w.default_dir()),
        "conflict",
        "{plan}"
    );
    assert_eq!(flow_action(&plan, &w.default_dir()), "held", "{plan}");
    let r = ok(&w.apply_out(plan["plan_digest"].as_str().unwrap(), &[]));
    assert_eq!(r["state"], json!("unchanged"), "{r}");
    assert!(!w.default_dir().join("rules/moai").exists());
    assert_eq!(
        fs::read_to_string(w.default_dir().join("rules/moai.md")).unwrap(),
        "my routing\n"
    );
}

// ---- removal_ ------------------------------------------------------------------------------

fn old_block_of(plan: &Value, dir: &Path) -> Value {
    row(plan, dir)["old_block"].clone()
}

#[test]
fn removal_is_offered_with_the_warning_and_never_applied_by_default() {
    let w = World::new();
    let (claude_md, text) = claude_md_with_block(&w);
    // Before the routing file is in place the removal waits.
    let plan = w.plan(&[]);
    assert_eq!(
        old_block_of(&plan, &w.default_dir())["action"],
        json!("waiting"),
        "{plan}"
    );
    w.apply(&[]);
    let plan = w.plan(&[]);
    let old = old_block_of(&plan, &w.default_dir());
    assert_eq!(old["action"], json!("offered"), "{plan}");
    assert!(
        old["detail"].as_str().unwrap().contains("profile"),
        "{plan}"
    );
    // An apply without the pick leaves the user's file as it is.
    let r = ok(&w.apply_out(plan["plan_digest"].as_str().unwrap(), &[]));
    assert_eq!(r["state"], json!("unchanged"), "{r}");
    assert_eq!(fs::read_to_string(&claude_md).unwrap(), text);
}

#[test]
fn removal_is_not_offered_without_a_block() {
    let w = World::new();
    write(&w.default_dir().join("CLAUDE.md"), "# Mine\n");
    write_old_state(&w, &old_block());
    let plan = w.plan(&["--remove-old-block"]);
    assert!(
        row(&plan, &w.default_dir())["old_block"].is_null(),
        "{plan}"
    );
    // No CLAUDE.md at all: nothing to offer either.
    let w = World::new();
    let plan = w.plan(&["--remove-old-block"]);
    assert!(
        row(&plan, &w.default_dir())["old_block"].is_null(),
        "{plan}"
    );
}

#[test]
fn removal_with_the_pick_removes_exactly_the_block() {
    let w = World::new();
    let (claude_md, text) = claude_md_with_block(&w);
    w.apply(&[]);
    let extra = ["--remove-old-block"];
    let plan = w.plan(&extra);
    assert_eq!(
        old_block_of(&plan, &w.default_dir())["action"],
        json!("remove"),
        "{plan}"
    );
    let r = ok(&w.apply_out(plan["plan_digest"].as_str().unwrap(), &extra));
    assert_eq!(r["state"], json!("applied"), "{r}");
    assert_eq!(
        fs::read_to_string(&claude_md).unwrap(),
        text.replacen(&old_block(), "", 1)
    );
    // The text beside the block is every byte it was.
    let now = fs::read_to_string(&claude_md).unwrap();
    assert!(now.starts_with("# Mine\n\nKeep this.\n\n") && now.ends_with("\nAfter.\n"));
    assert!(!now.contains("moai:managed"));
}

#[test]
fn removal_waits_until_the_routing_file_is_in_place_and_moais_own() {
    let w = World::new();
    let (claude_md, text) = claude_md_with_block(&w);
    let extra = ["--remove-old-block"];
    // One run with the pick: the routing file is written, the block stays.
    let plan = w.plan(&extra);
    assert_eq!(
        old_block_of(&plan, &w.default_dir())["action"],
        json!("waiting"),
        "{plan}"
    );
    ok(&w.apply_out(plan["plan_digest"].as_str().unwrap(), &extra));
    assert!(w.default_dir().join("rules/moai.md").is_file());
    assert_eq!(fs::read_to_string(&claude_md).unwrap(), text);
    // A routing file that is not moai's own keeps it waiting.
    let w = World::new();
    claude_md_with_block(&w);
    write(&w.default_dir().join("rules/moai.md"), "mine\n");
    let plan = w.plan(&extra);
    assert_eq!(
        old_block_of(&plan, &w.default_dir())["action"],
        json!("waiting"),
        "{plan}"
    );
}

#[test]
fn removal_an_edited_block_stays_and_the_reason_is_named() {
    let w = World::new();
    let (claude_md, text) = claude_md_with_block(&w);
    w.apply(&[]);
    let edited = text.replace("old text", "my own text");
    write(&claude_md, &edited);
    let extra = ["--remove-old-block"];
    let plan = w.plan(&extra);
    let old = old_block_of(&plan, &w.default_dir());
    assert_eq!(old["action"], json!("kept"), "{plan}");
    assert!(old["detail"].as_str().unwrap().contains("edited"), "{plan}");
    ok(&w.apply_out(plan["plan_digest"].as_str().unwrap(), &extra));
    assert_eq!(fs::read_to_string(&claude_md).unwrap(), edited);
}

#[test]
fn removal_without_a_baseline_stays() {
    let w = World::new();
    let block = old_block();
    let claude_md = w.default_dir().join("CLAUDE.md");
    write(&claude_md, &block);
    w.apply(&[]);
    let extra = ["--remove-old-block"];
    let plan = w.plan(&extra);
    let old = old_block_of(&plan, &w.default_dir());
    assert_eq!(old["action"], json!("kept"), "{plan}");
    assert!(
        old["detail"].as_str().unwrap().contains("no record"),
        "{plan}"
    );
    ok(&w.apply_out(plan["plan_digest"].as_str().unwrap(), &extra));
    assert_eq!(fs::read_to_string(&claude_md).unwrap(), block);
}

#[test]
fn removal_writes_no_other_text_into_the_users_claude_md() {
    // No block and no pick: the file is not even created or touched by an apply.
    let w = World::new();
    w.apply(&[]);
    assert!(!w.default_dir().join("CLAUDE.md").exists());
    let w = World::new();
    write(&w.default_dir().join("CLAUDE.md"), "# Mine\n");
    w.apply(&["--remove-old-block"]);
    assert_eq!(
        fs::read_to_string(w.default_dir().join("CLAUDE.md")).unwrap(),
        "# Mine\n"
    );
}

#[test]
fn removal_a_linked_or_non_utf8_claude_md_is_left() {
    let w = World::new();
    let away = w.base.join("away/CLAUDE.md");
    write(&away, old_block());
    fs::create_dir_all(w.default_dir()).unwrap();
    symlink(&away, w.default_dir().join("CLAUDE.md")).unwrap();
    write_old_state(&w, &old_block());
    w.apply(&[]);
    let extra = ["--remove-old-block"];
    let plan = w.plan(&extra);
    assert_eq!(
        old_block_of(&plan, &w.default_dir())["action"],
        json!("kept"),
        "{plan}"
    );
    ok(&w.apply_out(plan["plan_digest"].as_str().unwrap(), &extra));
    assert_eq!(fs::read_to_string(&away).unwrap(), old_block());

    let w = World::new();
    write(&w.default_dir().join("CLAUDE.md"), [0xff, 0xfe]);
    w.apply(&[]);
    let plan = w.plan(&extra);
    assert_eq!(
        old_block_of(&plan, &w.default_dir())["action"],
        json!("kept"),
        "{plan}"
    );
    ok(&w.apply_out(plan["plan_digest"].as_str().unwrap(), &extra));
    assert_eq!(
        fs::read(w.default_dir().join("CLAUDE.md")).unwrap(),
        [0xff, 0xfe]
    );
}

// ---- revert_ -------------------------------------------------------------------------------

/// Applies the first package, then the second with the removal pick; the two receipt ids.
fn two_applies(w: &mut World, first: &Path, second: &Path) -> (String, String) {
    claude_md_with_block(w);
    w.pkg = first.to_path_buf();
    let r1 = w.apply(&[]);
    w.pkg = second.to_path_buf();
    let r2 = w.apply(&["--remove-old-block"]);
    (
        r1["receipt"].as_str().unwrap().to_string(),
        r2["receipt"].as_str().unwrap().to_string(),
    )
}

#[test]
fn revert_restores_the_files_and_the_removed_block_and_offers_the_removal_again() {
    let v1 = temp_pkg(ROUTING_V1, Some("first\n"));
    let v2 = temp_pkg(ROUTING_V2, Some("second\n"));
    let mut w = World::new();
    let (_first, second) = two_applies(&mut w, v1.path(), v2.path());
    let dir = w.default_dir();
    assert_eq!(
        fs::read_to_string(dir.join("rules/moai/flow.md")).unwrap(),
        "second\n"
    );
    let claude_md = dir.join("CLAUDE.md");
    assert!(!fs::read_to_string(&claude_md)
        .unwrap()
        .contains("moai:managed"));
    let v = w.revert(&second);
    assert_eq!(v["state"], json!("reverted"), "{v}");
    assert_eq!(v["reverted"], json!(["claude"]), "{v}");
    assert_eq!(v["skipped"], json!([]), "{v}");
    assert_eq!(
        fs::read_to_string(dir.join("rules/moai.md")).unwrap(),
        ROUTING_V1.replacen("{invoke}", "/moai", 1)
    );
    assert_eq!(
        fs::read_to_string(dir.join("rules/moai/flow.md")).unwrap(),
        "first\n"
    );
    assert!(fs::read_to_string(&claude_md)
        .unwrap()
        .contains(&old_block()));
    // The baseline is back: with the first package the removal is offered again.
    let w = w.package(v1.path());
    let plan = w.plan(&["--remove-old-block"]);
    assert_eq!(
        old_block_of(&plan, &dir)["action"],
        json!("remove"),
        "{plan}"
    );
    assert_eq!(routing_action(&plan, &dir), "unchanged", "{plan}");
}

#[test]
fn revert_of_a_fresh_apply_removes_created_folders_only_when_empty() {
    let w = World::new();
    let id = w.apply(&[])["receipt"].as_str().unwrap().to_string();
    let dir = w.default_dir();
    write(&dir.join("rules/other.md"), "keep\n");
    let v = w.revert(&id);
    assert_eq!(v["reverted"], json!(["claude"]), "{v}");
    assert!(!dir.join("rules/moai.md").exists());
    assert!(
        !dir.join("rules/moai").exists(),
        "the empty folder is removed"
    );
    assert!(
        dir.join("rules/other.md").is_file(),
        "the folder with a file stays"
    );
    // With nothing else in it, the folders the apply made go and the config folder stays or goes
    // with the files only when empty.
    let w = World::new();
    let id = w.apply(&[])["receipt"].as_str().unwrap().to_string();
    w.revert(&id);
    assert!(!w.default_dir().join("rules").exists());
    assert_eq!(
        w.rules_state_files().len(),
        1,
        "the state stays, holding no digest"
    );
}

#[test]
fn revert_skips_each_changed_file_by_name_and_reverts_the_others() {
    for changed in ["rules/moai.md", "rules/moai/flow.md", "CLAUDE.md"] {
        let v1 = temp_pkg(ROUTING_V1, Some("first\n"));
        let v2 = temp_pkg(ROUTING_V2, Some("second\n"));
        let mut w = World::new();
        let (_, second) = two_applies(&mut w, v1.path(), v2.path());
        let dir = w.default_dir();
        write(&dir.join(changed), "changed after the apply\n");
        let v = w.revert(&second);
        let skipped = v["skipped"].as_array().unwrap();
        assert_eq!(skipped.len(), 1, "{changed}: {v}");
        assert_eq!(
            skipped[0]["reason"],
            json!("user_changed_since_apply"),
            "{v}"
        );
        assert_eq!(skipped[0]["path"], json!(s(&dir.join(changed))), "{v}");
        assert_eq!(
            fs::read_to_string(dir.join(changed)).unwrap(),
            "changed after the apply\n"
        );
        // The other two went back.
        if changed != "rules/moai.md" {
            assert_eq!(
                fs::read_to_string(dir.join("rules/moai.md")).unwrap(),
                ROUTING_V1.replacen("{invoke}", "/moai", 1)
            );
        }
        if changed != "rules/moai/flow.md" {
            assert_eq!(
                fs::read_to_string(dir.join("rules/moai/flow.md")).unwrap(),
                "first\n"
            );
        }
        if changed != "CLAUDE.md" {
            assert!(fs::read_to_string(dir.join("CLAUDE.md"))
                .unwrap()
                .contains("moai:managed"));
        }
    }
}

#[test]
fn revert_never_writes_through_a_link_made_after_the_apply() {
    let w = World::new();
    let id = w.apply(&[])["receipt"].as_str().unwrap().to_string();
    let dir = w.default_dir();
    let away = w.base.join("away");
    fs::create_dir_all(&away).unwrap();
    fs::rename(dir.join("rules"), away.join("rules")).unwrap();
    symlink(away.join("rules"), dir.join("rules")).unwrap();
    let v = w.revert(&id);
    assert_eq!(v["reverted"], json!([]), "{v}");
    let skipped = v["skipped"].as_array().unwrap();
    assert_eq!(skipped.len(), 2, "{v}");
    assert!(skipped.iter().all(|e| e["reason"]
        .as_str()
        .unwrap()
        .starts_with("host_folder_not_plain")));
    assert!(away.join("rules/moai.md").is_file());
    assert!(away.join("rules/moai/flow.md").is_file());
}

#[test]
fn revert_twice_is_refused() {
    let w = World::new();
    let id = w.apply(&[])["receipt"].as_str().unwrap().to_string();
    w.revert(&id);
    let m = refused(&w.revert_out(&id));
    assert!(m.contains("already reverted"), "{m}");
}

#[test]
fn revert_a_legacy_receipt_entry_without_a_kind_still_reverts() {
    // What the earlier version left: a block appended to the config folder's CLAUDE.md, its
    // receipt entry holding no `kind` and no `config_dir`, the before copy named after the host.
    let w = World::new();
    let before = "# Mine\n";
    let block = old_block();
    let after = format!("{before}\n{block}");
    let claude_md = w.default_dir().join("CLAUDE.md");
    write(&claude_md, &after);
    write_old_state(&w, &block);
    let id = "routing-1-0";
    let rdir = w.state_dir().join("receipts").join(id);
    write(&rdir.join("claude.before"), before);
    write(
        &rdir.join("receipt.json"),
        json!({
            "schema_version": 1,
            "id": id,
            "plan_digest": "sha256:0",
            "permission_reference": PERM,
            "status": "committed",
            "entries": [{
                "host": "claude",
                "action": "append",
                "before_sha256": moai_cli::digest(before.as_bytes()),
                "after_sha256": moai_cli::digest(after.as_bytes()),
                "baseline_before": null
            }]
        })
        .to_string(),
    );
    let v = w.revert(id);
    assert_eq!(v["reverted"], json!(["claude"]), "{v}");
    assert_eq!(fs::read_to_string(&claude_md).unwrap(), before);
    let state: Value =
        serde_json::from_slice(&fs::read(w.state_dir().join("claude.json")).unwrap()).unwrap();
    assert_eq!(state["block_sha256"], Value::Null, "{state}");
}

// ---- folders_ ------------------------------------------------------------------------------

#[test]
fn folders_a_picked_folder_gets_no_copy_and_the_default_folder_keeps_the_only_state() {
    let w = World::new();
    let a = w.base.join("extra/a");
    let b = w.base.join("extra/b");
    let extra = [
        "--config-dir",
        a.to_str().unwrap(),
        "--config-dir",
        b.to_str().unwrap(),
    ];
    // A folder given twice, or the default one given again, is planned once.
    let default_dir = w.default_dir();
    let dup = [
        "--config-dir",
        a.to_str().unwrap(),
        "--config-dir",
        a.to_str().unwrap(),
        "--config-dir",
        b.to_str().unwrap(),
        "--config-dir",
        default_dir.to_str().unwrap(),
    ];
    assert_eq!(w.plan(&dup)["hosts"].as_array().unwrap().len(), 3);
    let r = w.apply(&extra);
    assert_eq!(r["state"], json!("applied"), "{r}");
    assert_eq!(
        fs::read_to_string(default_dir.join("rules/moai.md")).unwrap(),
        shipped_routing_text()
    );
    for d in [&a, &b] {
        assert!(!d.exists(), "no copy in {}", d.display());
    }
    let names = w.rules_state_files();
    assert_eq!(names.len(), 1, "{names:?}");
    let hex = names[0]
        .strip_prefix("rules-")
        .and_then(|x| x.strip_suffix(".json"))
        .unwrap_or("");
    assert!(
        hex.len() == 12 && hex.bytes().all(|c| c.is_ascii_hexdigit()),
        "{names:?}"
    );
    // Each picked folder is told what to do, and the default folder is judged on its own.
    write(&a.join("rules/moai.md"), "mine\n");
    let plan = w.plan(&extra);
    assert_eq!(routing_action(&plan, &a), "not_linked", "{plan}");
    assert_eq!(routing_action(&plan, &b), "not_linked", "{plan}");
    assert_eq!(routing_action(&plan, &default_dir), "unchanged", "{plan}");
}

#[test]
fn folders_a_picked_folders_own_rules_never_stop_the_default_copy() {
    let w = World::new();
    let a = w.base.join("extra/a");
    let b = w.base.join("extra/b");
    write(&a.join("rules/moai.md"), "theirs\n");
    let extra = [
        "--config-dir",
        a.to_str().unwrap(),
        "--config-dir",
        b.to_str().unwrap(),
    ];
    w.apply(&extra);
    assert_eq!(
        fs::read_to_string(a.join("rules/moai.md")).unwrap(),
        "theirs\n"
    );
    assert!(!b.exists());
    assert!(w.default_dir().join("rules/moai.md").is_file());
}

#[test]
fn folders_a_profile_whose_rules_folder_links_to_the_default_one_is_a_reference_not_a_copy() {
    // An Orca profile shares the default folder's `rules` through a link, the way it shares
    // `agents` and `output-styles`; the profile must not get a second owner of the same file.
    let w = World::new();
    w.apply(&[]);
    let profile = w.base.join("profile");
    fs::create_dir_all(&profile).unwrap();
    symlink(w.default_dir().join("rules"), profile.join("rules")).unwrap();
    let extra = ["--config-dir", profile.to_str().unwrap()];
    let plan = w.plan(&extra);
    assert_eq!(routing_action(&plan, &profile), "linked", "{plan}");
    assert_eq!(flow_action(&plan, &profile), "linked", "{plan}");
    assert!(
        row(&plan, &profile)["section"]["detail"]
            .as_str()
            .unwrap()
            .contains("no second copy"),
        "{plan}"
    );
    // The default folder's own rows stay as they were.
    assert_eq!(routing_action(&plan, &w.default_dir()), "unchanged");
    // Applying the plan writes nothing and records no second state file.
    let before = w.home.snapshot();
    let r = ok(&w.apply_out(plan["plan_digest"].as_str().unwrap(), &extra));
    assert_eq!(r["state"], json!("unchanged"), "{r}");
    assert_eq!(w.home.snapshot(), before, "a reference writes nothing");
    assert_eq!(
        w.rules_state_files().len(),
        1,
        "{:?}",
        w.rules_state_files()
    );
}

#[test]
fn folders_a_reference_to_an_edited_file_stays_a_conflict() {
    // The linked profile reaches the user's own edit of the default file: it is reported as the
    // conflict it is, never relabelled as a harmless reference.
    let w = World::new();
    w.apply(&[]);
    let edited = w.default_dir().join("rules/moai.md");
    write(&edited, "my own routing notes\n");
    let profile = w.base.join("profile");
    fs::create_dir_all(&profile).unwrap();
    symlink(w.default_dir().join("rules"), profile.join("rules")).unwrap();
    let plan = w.plan(&["--config-dir", profile.to_str().unwrap()]);
    assert_eq!(
        routing_action(&plan, &w.default_dir()),
        "conflict",
        "{plan}"
    );
    assert_eq!(routing_action(&plan, &profile), "conflict", "{plan}");
    assert_eq!(
        fs::read_to_string(&edited).unwrap(),
        "my own routing notes\n"
    );
}

#[test]
fn folders_a_link_that_leaves_the_default_folder_is_not_a_reference() {
    // A `rules` link to a folder outside the default config folder is no reference: the row says
    // `not_linked` and nothing is written through it (SPEC decision D-2).
    let w = World::new();
    w.apply(&[]);
    let outside = w.base.join("elsewhere/rules");
    fs::create_dir_all(&outside).unwrap();
    let profile = w.base.join("profile");
    fs::create_dir_all(&profile).unwrap();
    symlink(&outside, profile.join("rules")).unwrap();
    let plan = w.plan(&["--config-dir", profile.to_str().unwrap()]);
    assert_eq!(routing_action(&plan, &profile), "not_linked", "{plan}");
    w.apply(&["--config-dir", profile.to_str().unwrap()]);
    assert!(
        !outside.join("moai.md").exists() && !outside.join("moai").exists(),
        "nothing was written through the link"
    );
}

#[test]
fn folders_a_relative_config_dir_is_refused() {
    let w = World::new();
    let m = w.plan_refused(&["--config-dir", "relative/dir"]);
    assert!(m.contains("--config-dir") && m.contains("absolute"), "{m}");
    assert!(!w.default_dir().exists());
}

#[test]
fn folders_revert_takes_the_files_from_the_receipt() {
    let w = World::new();
    let a = w.base.join("extra/a");
    let extra = ["--config-dir", a.to_str().unwrap()];
    let id = w.apply(&extra)["receipt"].as_str().unwrap().to_string();
    let v = w.revert(&id);
    assert_eq!(v["reverted"], json!(["claude"]), "{v}");
    assert_eq!(v["files"].as_array().unwrap().len(), 2, "{v}");
    assert!(!a.exists());
    assert!(!w.default_dir().join("rules/moai.md").exists());
}
