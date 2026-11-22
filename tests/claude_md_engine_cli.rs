//! The four open findings of the AGENTS.md merge engine (SPEC-20261122-claude-md-engine-findings):
//! the warning about files Claude Code stops reading, the notice about joined blank lines, the
//! secret scan that reads only the copied text, and the receipts line in `.gitignore`.
//!
//! Every case runs the built binary in a temporary git repository (or a plain folder) with an empty
//! temporary HOME. The test names start with `hidden_`, `collapse_`, `scan_` and `receipts_`: the
//! acceptance criteria of the SPEC filter on those prefixes. Credential-shaped text is built at run
//! time so no fixture here is a secret.

#![cfg(unix)]

mod common;
use common::git::{git_ok, git_raw};
use common::CleanEnv;
use moai_cli::agents_local_merge::{BEGIN as LOCAL_BEGIN, END as LOCAL_END};
use serde_json::{json, Value};
use std::{
    fs,
    io::Write,
    path::PathBuf,
    process::{Command, Stdio},
};
use tempfile::TempDir;

const REF: &str = "fixture-approved";
const PLACEHOLDER: &str = "<permission-reference>";
const BEGIN: &str = "<!-- moai:managed:begin id=moai-workflow -->";
const END: &str = "<!-- moai:managed:end id=moai-workflow -->";
const AGENTS_MD: &str = "AGENTS.md";
const HIDDEN_MD: &str = ".claude/AGENTS.md";
const TEAM: &str = "# Team rules\n\nWrite tests first.\n";
const MINE: &str = "My own notes\nPrefer tabs.\n";
/// The receipt folder of a project that has been through `project-apply` (state folder `.moai`).
const RECEIPTS: &str = ".moai/receipts";
const RECEIPTS_LINE: &str = ".moai/receipts/";

struct Fx {
    _tmp: TempDir,
    root: PathBuf,
    home: PathBuf,
}

impl Fx {
    /// A git repository with nothing in it.
    fn new() -> Fx {
        let fx = Fx::plain();
        git_ok(&fx.root, &["init", "-q", "-b", "main"]);
        fx
    }

    /// A folder that is not a git repository.
    fn plain() -> Fx {
        let tmp = TempDir::new().unwrap();
        let base = fs::canonicalize(tmp.path()).unwrap();
        let (root, home) = (base.join("project"), base.join("home"));
        fs::create_dir_all(&root).unwrap();
        fs::create_dir_all(&home).unwrap();
        Fx {
            _tmp: tmp,
            root,
            home,
        }
    }

    fn p(&self) -> &str {
        self.root.to_str().unwrap()
    }

    fn write(&self, rel: &str, body: &str) {
        let path = self.root.join(rel);
        fs::create_dir_all(path.parent().unwrap()).unwrap();
        fs::write(path, body).unwrap();
    }

    fn read(&self, rel: &str) -> String {
        fs::read_to_string(self.root.join(rel)).unwrap()
    }

    /// The receipt folder of a project with no MoAI state yet, as the project state resolver
    /// answers it (SPEC-20261121-path-root-moai-default: `.moai` for a tree with no earlier
    /// `.moai-cli`).
    fn fresh_receipts(&self) -> String {
        format!("{}/receipts", moai_cli::state_root::dir(&self.root))
    }

    fn fresh_receipts_line(&self) -> String {
        format!("{}/", self.fresh_receipts())
    }

    fn exists(&self, rel: &str) -> bool {
        fs::symlink_metadata(self.root.join(rel)).is_ok()
    }

    /// `true` when git says the path is ignored (the folder must exist for a folder rule).
    fn ignored(&self, rel: &str) -> bool {
        git_raw(&self.root, &["check-ignore", "-q", "--", rel]).success
    }

    fn exec(&self, args: &[&str], stdin: &[u8]) -> Value {
        let mut child = Command::new(common::MOAI_BIN)
            .args(args)
            .clean_env()
            .env("HOME", &self.home)
            .env("PATH", "/usr/bin:/bin")
            .env("LC_ALL", "C")
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .spawn()
            .unwrap();
        let mut pipe = child.stdin.take().unwrap();
        match pipe.write_all(stdin) {
            Ok(()) => {}
            Err(e) if e.kind() == std::io::ErrorKind::BrokenPipe => {}
            Err(e) => panic!("stdin write failed: {e}"),
        }
        drop(pipe);
        let out = child.wait_with_output().unwrap();
        serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
            panic!(
                "{args:?}: {}{}",
                String::from_utf8_lossy(&out.stdout),
                String::from_utf8_lossy(&out.stderr)
            )
        })
    }

    fn call(&self, args: &[&str], stdin: &[u8]) -> Value {
        let v = self.exec(args, stdin);
        assert_eq!(v["ok"], json!(true), "{args:?}: {v}");
        v["result"].clone()
    }

    /// Writes the project profile only.
    fn profile(&self) {
        let detect = self.call(&["project-detect", self.p()], b"");
        let profile = json!({
            "schema_version": 2, "project_id": "demo", "purpose": "Build a small editor",
            "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude"],
            "detected": detect["detected"],
        });
        let plan = self.call(
            &["project-plan", self.p()],
            &serde_json::to_vec(&profile).unwrap(),
        );
        let digest = plan["plan_digest"].as_str().unwrap().to_string();
        self.call(
            &[
                "project-apply",
                self.p(),
                "--plan-digest",
                &digest,
                "--permission-reference",
                REF,
            ],
            &serde_json::to_vec(&plan).unwrap(),
        );
    }

    fn update_plan(&self) -> Value {
        self.call(&["project-update-plan", self.p()], b"")
    }

    /// The profile applied and the generation step run, so a later preview offers the picks.
    fn ready(&self) {
        self.profile();
        let update = self.update_plan();
        let generation = step(&update, "generation");
        if generation["state"] == "ready" {
            let out = self.run_printed(generation);
            assert_eq!(out["ok"], json!(true), "{out}");
        }
        let after = step(&self.update_plan(), "generation").clone();
        assert_eq!(after["state"], "unchanged", "{after}");
    }

    /// Runs the command a step prints, the way a session does.
    fn run_printed(&self, with_apply: &Value) -> Value {
        let argv: Vec<String> = with_apply["apply"]["argv"]
            .as_array()
            .unwrap_or_else(|| panic!("no apply command: {with_apply}"))
            .iter()
            .map(|w| w.as_str().unwrap().to_string())
            .collect();
        let args: Vec<&str> = argv[1..]
            .iter()
            .map(|w| if w == PLACEHOLDER { REF } else { w.as_str() })
            .collect();
        self.exec(
            &args,
            &serde_json::to_vec(&with_apply["apply"]["stdin"]).unwrap(),
        )
    }

    /// `agents-merge-plan --pick <id>`: the result.
    fn pick_plan(&self, id: &str) -> Value {
        self.call(&["agents-merge-plan", self.p(), "--pick", id], b"")
    }

    /// Plans and applies a pick; the apply must succeed. Returns the apply result.
    fn pick(&self, id: &str) -> Value {
        let plan = self.pick_plan(id);
        assert_eq!(plan["state"], json!("ready"), "{id}: {plan}");
        let digest = plan["plan_digest"].as_str().unwrap();
        self.call(
            &[
                "agents-merge-apply",
                self.p(),
                "--plan-digest",
                digest,
                "--permission-reference",
                REF,
                "--pick",
                id,
            ],
            b"",
        )
    }

    /// The standalone merge of `AGENTS.local.md` into `CLAUDE.local.md`: the plan.
    fn merge_plan(&self) -> Value {
        self.call(&["agents-merge-plan", self.p()], b"")
    }

    /// Plans and applies the standalone merge; the apply must succeed. Returns the apply result.
    fn merge(&self) -> Value {
        let plan = self.merge_plan();
        assert_eq!(plan["action"], json!("merge"), "{plan}");
        self.call(
            &[
                "agents-merge-apply",
                self.p(),
                "--plan-digest",
                plan["plan_digest"].as_str().unwrap(),
                "--permission-reference",
                REF,
            ],
            b"",
        )
    }
}

fn step<'a>(plan: &'a Value, name: &str) -> &'a Value {
    plan["phases"]
        .as_array()
        .unwrap()
        .iter()
        .flat_map(|p| p["steps"].as_array().into_iter().flatten())
        .find(|s| s["step"] == name)
        .unwrap_or_else(|| panic!("no step {name}: {plan}"))
}

/// The strings of a JSON array of warnings.
fn texts(warnings: &Value) -> Vec<String> {
    warnings
        .as_array()
        .unwrap_or_else(|| panic!("no warnings array: {warnings}"))
        .iter()
        .map(|w| w.as_str().unwrap().to_string())
        .collect()
}

/// The warning that Claude Code stops reading exactly `files`, or `None`.
fn hidden_warning(warnings: &Value, files: &str) -> Option<String> {
    let reads = format!("Claude Code reads {files} only while no CLAUDE.md, .claude/CLAUDE.md or CLAUDE.local.md exists");
    let stops = format!("makes it stop reading {files} in this project");
    texts(warnings)
        .into_iter()
        .find(|w| w.contains(&reads) && w.contains(&stops))
}

/// Any warning that says Claude Code stops reading some file.
fn any_hidden(warnings: &Value) -> bool {
    texts(warnings)
        .iter()
        .any(|w| w.contains("only while no CLAUDE.md"))
}

/// The three project shapes of the first requirement: what the project holds and which files the
/// writer, the picks and the preview warn would stop being read. The root `AGENTS.md` is carried
/// into the new `CLAUDE.md` (copied text or an import), so only `.claude/AGENTS.md` is named.
#[allow(clippy::type_complexity)]
fn shapes() -> [(&'static str, Vec<&'static str>, Option<&'static str>); 3] {
    [
        (HIDDEN_MD, vec![HIDDEN_MD], Some(HIDDEN_MD)),
        (AGENTS_MD, vec![AGENTS_MD], None),
        (
            "AGENTS.md and .claude/AGENTS.md",
            vec![AGENTS_MD, HIDDEN_MD],
            Some(HIDDEN_MD),
        ),
    ]
}

/// `warned` is the file the warning names, or `None` when no warning about hidden files belongs.
fn assert_hidden(warnings: &Value, warned: Option<&str>, context: &str) {
    match warned {
        Some(files) => assert!(
            hidden_warning(warnings, files).is_some(),
            "{context}: {warnings}"
        ),
        None => assert!(!any_hidden(warnings), "{context}: {warnings}"),
    }
}

fn project_with(files: &[&str]) -> Fx {
    let fx = Fx::new();
    for f in files {
        fx.write(f, TEAM);
    }
    fx.profile();
    fx
}

/// The request of the generation step of a fresh preview.
fn generation_request(fx: &Fx) -> Value {
    let update = fx.update_plan();
    let request = step(&update, "generation")["apply"]["stdin"]["request"].clone();
    assert!(request.is_object(), "no request in {update}");
    request
}

// --- REQ-1: the files Claude Code stops reading ---------------------------------------------------

/// The generation writer names every file that would stop being read and that its block does not
/// carry: in the plan of the whole generation, and in the plan of the agents writer alone.
#[test]
fn hidden_writer_names_every_file_it_would_stop_claude_code_reading() {
    for (named, files, warned) in shapes() {
        let fx = project_with(&files);
        let request = generation_request(&fx);
        let plan = fx.call(
            &["generation-plan", fx.p()],
            &serde_json::to_vec(&request).unwrap(),
        );
        assert_hidden(&plan["warnings"], warned, named);

        // The agents writer alone needs the workflow reference folder to exist.
        let reference = request["workflow_reference"].as_str().unwrap();
        fx.write(
            &format!("{}/keep.md", reference.trim_end_matches('/')),
            "x\n",
        );
        let agents_request = json!({
            "protocol_version": 1, "language": "en", "local_file": "none",
            "workflow_reference": request["workflow_reference"], "picks": [],
        });
        let agents = fx.call(
            &["agents-plan", fx.p()],
            &serde_json::to_vec(&agents_request).unwrap(),
        );
        assert_hidden(&agents["warnings"], warned, named);
    }
}

/// A project that holds only `.claude/AGENTS.md` gets a CLAUDE.md from the writer, and the
/// warning says that file stops being read; the file itself stays as it is.
#[test]
fn hidden_claude_dir_agents_md_alone_is_warned_before_the_writer_creates_claude_md() {
    let fx = project_with(&[HIDDEN_MD]);
    let request = generation_request(&fx);
    let plan = fx.call(
        &["generation-plan", fx.p()],
        &serde_json::to_vec(&request).unwrap(),
    );
    let claude = plan["files"]
        .as_array()
        .unwrap()
        .iter()
        .find(|f| f["path"] == "CLAUDE.md")
        .unwrap_or_else(|| panic!("the writer plans no CLAUDE.md: {plan}"));
    assert_eq!(claude["action"], json!("create"), "{plan}");
    let warning = hidden_warning(&plan["warnings"], HIDDEN_MD).expect("warning");
    assert!(
        warning.contains("creating CLAUDE.md makes it stop reading"),
        "{warning}"
    );
    assert!(
        warning.contains("claude-md-and-agents-md"),
        "the warning names the official setting: {warning}"
    );
    assert_eq!(fx.read(HIDDEN_MD), TEAM);
}

/// The pick plan names `.claude/AGENTS.md` (no pick carries it) and not the root `AGENTS.md` (a
/// pick copies or imports it); a project that holds only `.claude/AGENTS.md` is warned but offered
/// no pick.
#[test]
fn hidden_pick_plan_names_every_file_and_offers_no_pick_for_the_claude_dir_file() {
    for (named, files, warned) in shapes() {
        let fx = project_with(&files);
        let plan = fx.pick_plan("move");
        assert_hidden(&plan["shadow_warnings"], warned, named);
        if files == [HIDDEN_MD] {
            assert_eq!(plan["state"], json!("none"), "{plan}");
            assert!(
                plan["blocked"].as_array().unwrap().is_empty(),
                "no pick is offered, so none is refused: {plan}"
            );
        } else {
            assert_eq!(plan["state"], json!("ready"), "{named}: {plan}");
        }
    }
}

/// The update preview warns in every shape: the generation step before it writes, and the picks
/// step once the project is ready.
#[test]
fn hidden_update_preview_warns_in_the_generation_and_the_picks_steps() {
    for (named, files, warned) in shapes() {
        let fx = project_with(&files);
        let update = fx.update_plan();
        let generation = step(&update, "generation");
        assert_hidden(&generation["preview"]["warnings"], warned, named);
        let picks = step(&update, "agents-picks");
        if files == [HIDDEN_MD] {
            // No pick is offered for this file (a separate feature), but the step carries the warning.
            assert_eq!(picks["state"], json!("skip"), "{picks}");
            assert_hidden(&picks["preview"]["warnings"], warned, named);
        }
    }
    for (named, files, warned) in shapes()
        .into_iter()
        .filter(|(_, f, _)| f.contains(&AGENTS_MD))
    {
        let fx = Fx::new();
        for f in &files {
            fx.write(f, TEAM);
        }
        fx.ready();
        let picks = step(&fx.update_plan(), "agents-picks").clone();
        assert_eq!(picks["state"], json!("needs_decision"), "{picks}");
        // A missing `warnings` key on the preview means no warning.
        let warnings = &picks["preview"]["warnings"];
        if warnings.is_null() {
            assert!(warned.is_none(), "{named}: {picks}");
        } else {
            assert_hidden(warnings, warned, named);
        }
    }
}

/// Once a CLAUDE file exists Claude Code has stopped reading the files already, so nothing is
/// warned about: not by the writer, the pick plan or the preview.
#[test]
fn hidden_nothing_is_warned_while_a_claude_file_exists() {
    for claude in ["CLAUDE.md", ".claude/CLAUDE.md", "CLAUDE.local.md"] {
        for (named, files, _) in shapes() {
            let fx = Fx::new();
            for f in &files {
                fx.write(f, TEAM);
            }
            fx.write(claude, "# My own rules\n");
            fx.profile();
            let request = generation_request(&fx);
            let plan = fx.call(
                &["generation-plan", fx.p()],
                &serde_json::to_vec(&request).unwrap(),
            );
            assert!(
                !any_hidden(&plan["warnings"]),
                "{claude} + {named}: {}",
                plan["warnings"]
            );
            let picks = fx.pick_plan("move");
            assert!(
                !any_hidden(&picks["shadow_warnings"]),
                "{claude} + {named}: {picks}"
            );
            let update = fx.update_plan();
            for name in ["generation", "agents-picks"] {
                let s = step(&update, name);
                let warnings = &s["preview"]["warnings"];
                assert!(
                    warnings.is_null() || !any_hidden(warnings),
                    "{claude} + {named} + {name}: {s}"
                );
            }
        }
    }
}

/// The writer that creates `CLAUDE.local.md` first (the personal file is wanted, no pick carries
/// the root `AGENTS.md` into a CLAUDE.md) names the root `AGENTS.md`: nothing carries it over, so
/// Claude Code would stop reading it without a word. Only a created CLAUDE.md carries it.
#[test]
fn hidden_writer_names_root_agents_md_when_only_the_personal_file_is_created() {
    let fx = project_with(&[AGENTS_MD]);
    let request = generation_request(&fx);
    let reference = request["workflow_reference"].as_str().unwrap();
    fx.write(
        &format!("{}/keep.md", reference.trim_end_matches('/')),
        "x\n",
    );
    let agents_request = json!({
        "protocol_version": 1, "language": "en", "local_file": "ignored",
        "workflow_reference": request["workflow_reference"], "picks": ["personal-file"],
    });
    let plan = fx.call(
        &["agents-plan", fx.p()],
        &serde_json::to_vec(&agents_request).unwrap(),
    );
    let action = |path: &str| {
        plan["files"]
            .as_array()
            .unwrap()
            .iter()
            .find(|f| f["path"] == path)
            .map(|f| f["action"].clone())
    };
    assert_eq!(action("CLAUDE.local.md"), Some(json!("create")), "{plan}");
    assert_ne!(action("CLAUDE.md"), Some(json!("create")), "{plan}");
    let warning = hidden_warning(&plan["warnings"], AGENTS_MD)
        .unwrap_or_else(|| panic!("no warning names AGENTS.md: {}", plan["warnings"]));
    assert!(
        warning.contains("creating CLAUDE.local.md makes it stop reading"),
        "{warning}"
    );
    assert_eq!(fx.read(AGENTS_MD), TEAM);
}

/// The personal-file merge keeps its own text and its confirmation token.
#[test]
fn hidden_personal_file_merge_keeps_its_warning_and_token() {
    let fx = Fx::new();
    fx.write(AGENTS_MD, TEAM);
    fx.write(HIDDEN_MD, TEAM);
    fx.write("AGENTS.local.md", MINE);
    let plan = fx.merge_plan();
    assert_eq!(plan["action"], json!("merge"), "{plan}");
    let warning = hidden_warning(&plan["warnings"], "AGENTS.md and .claude/AGENTS.md")
        .unwrap_or_else(|| panic!("{plan}"));
    assert!(
        warning.contains("creating CLAUDE.local.md makes it stop reading"),
        "{warning}"
    );
    assert_eq!(plan["needs_confirmation"], json!(true), "{plan}");
    assert!(plan["confirm_token"]
        .as_str()
        .is_some_and(|t| t.len() == 16));
}

// --- REQ-2: the joined blank lines ----------------------------------------------------------------

fn around_region() -> String {
    format!("# Team rules\n\nWrite tests first.\n\n\n{BEGIN}\n## MoAI workflow\n\nGenerated text.\n{END}\n\n\n\n## Mine\n\nPrefer small commits.\n")
}

/// `true` when some warning says the blank lines around the removed region are joined.
fn says_joined(warnings: &Value) -> bool {
    texts(warnings)
        .iter()
        .any(|w| w.contains("blank lines around") && w.contains("joined into one blank line"))
}

#[test]
fn collapse_a_managed_region_in_the_middle_says_the_blank_lines_are_joined() {
    let fx = Fx::new();
    fx.write(AGENTS_MD, &around_region());
    fx.ready();
    let plan = fx.pick_plan("move");
    assert_eq!(plan["state"], json!("ready"), "{plan}");
    assert!(says_joined(&plan["warnings"]), "{}", plan["warnings"]);
    let text = texts(&plan["warnings"])
        .into_iter()
        .find(|w| w.contains("joined into one blank line"))
        .unwrap();
    assert!(text.contains("AGENTS.md"), "{text}");

    // The copy is what the notice says: one blank line, and the original file is untouched.
    fx.pick("move");
    let claude = fx.read("CLAUDE.md");
    assert!(
        claude.contains("Write tests first.\n\n## Mine\n\nPrefer small commits.\n"),
        "{claude}"
    );
    assert!(!claude.contains("\n\n\n"), "{claude}");
    assert_eq!(fx.read(AGENTS_MD), around_region());
}

#[test]
fn collapse_no_region_or_text_on_one_side_only_says_nothing() {
    // No managed region at all.
    let fx = Fx::new();
    fx.write(
        AGENTS_MD,
        "# Team rules\n\nWrite tests first.\n\n\n## Mine\n\nPrefer tabs.\n",
    );
    fx.ready();
    let plan = fx.pick_plan("move");
    assert_eq!(plan["state"], json!("ready"), "{plan}");
    assert!(!says_joined(&plan["warnings"]), "{}", plan["warnings"]);

    // A region at the end: nothing is joined, the text stays on one side.
    let fx = Fx::new();
    fx.write(
        AGENTS_MD,
        &format!("# Team rules\n\nWrite tests first.\n\n{BEGIN}\n## MoAI\n{END}\n"),
    );
    fx.ready();
    let plan = fx.pick_plan("move");
    assert_eq!(plan["state"], json!("ready"), "{plan}");
    assert!(!says_joined(&plan["warnings"]), "{}", plan["warnings"]);

    // A region at the start.
    let fx = Fx::new();
    fx.write(
        AGENTS_MD,
        &format!("{BEGIN}\n## MoAI\n{END}\n\n\n# Team rules\n\nWrite tests first.\n"),
    );
    fx.ready();
    let plan = fx.pick_plan("move");
    assert_eq!(plan["state"], json!("ready"), "{plan}");
    assert!(!says_joined(&plan["warnings"]), "{}", plan["warnings"]);
}

// --- REQ-3: the secret scan reads only the copied text --------------------------------------------

/// Credential-shaped text, built here so no fixture in the source is itself a secret.
fn secret() -> String {
    format!("key: sk-{}", "a".repeat(24))
}

fn merged_local_block(body: &str) -> String {
    format!("{LOCAL_BEGIN}{} -->\n{body}{LOCAL_END}\n", "0".repeat(64))
}

fn reason_of(plan: &Value) -> String {
    plan["blocked"].to_string()
}

#[test]
fn scan_a_credential_inside_the_managed_region_does_not_block_the_plan() {
    let fx = Fx::new();
    fx.write(
        AGENTS_MD,
        &format!(
            "# Team rules\n\n{BEGIN}\n## MoAI workflow\n{}\n{END}\n\nWrite tests first.\n",
            secret()
        ),
    );
    fx.ready();
    let plan = fx.pick_plan("move");
    assert_eq!(plan["state"], json!("ready"), "{plan}");
    assert!(plan["blocked"].as_array().unwrap().is_empty(), "{plan}");
    fx.pick("move");
    let claude = fx.read("CLAUDE.md");
    assert!(claude.contains("Write tests first."), "{claude}");
    assert!(
        !claude.contains(&secret()),
        "the managed region is not copied"
    );
}

#[test]
fn scan_a_credential_inside_an_earlier_merged_personal_block_does_not_block_the_plan() {
    let fx = Fx::new();
    fx.write(
        AGENTS_MD,
        &format!(
            "# Team rules\n\n{}\nWrite tests first.\n",
            merged_local_block(&format!("{}\n", secret()))
        ),
    );
    fx.ready();
    let plan = fx.pick_plan("move");
    assert_eq!(plan["state"], json!("ready"), "{plan}");
    assert!(plan["blocked"].as_array().unwrap().is_empty(), "{plan}");
    fx.pick("move");
    let claude = fx.read("CLAUDE.md");
    assert!(claude.contains("Write tests first."), "{claude}");
    assert!(
        !claude.contains(&secret()),
        "the personal block is not copied"
    );
}

#[test]
fn scan_a_credential_in_the_copied_text_blocks_the_plan_with_the_line_of_the_whole_file() {
    // Line 9 of the file; the managed region above it is not part of the copy, so the copied text
    // alone would call it line 4.
    let fx = Fx::new();
    let file = format!(
        "# Team rules\n\n{BEGIN}\n## MoAI workflow\nGenerated text.\n{END}\n\nWrite tests first.\n{}\n",
        secret()
    );
    fx.write(AGENTS_MD, &file);
    fx.ready();
    let plan = fx.pick_plan("move");
    assert_eq!(plan["state"], json!("blocked"), "{plan}");
    let said = reason_of(&plan);
    assert!(said.contains("(lines 9)"), "{said}");
    assert!(!said.contains("sk-"), "the value is never printed: {said}");
    assert_eq!(fx.read(AGENTS_MD), file);
    assert!(!fx.exists("CLAUDE.md"));

    // The import pick copies none of the file, so it is not blocked by it.
    let import = fx.pick_plan("import");
    assert_eq!(import["state"], json!("ready"), "{import}");
}

#[test]
fn scan_line_numbers_count_through_both_removed_blocks() {
    // The managed region (3 lines) and an earlier personal block (3 lines) both come off; the
    // credential is on line 12 of the file.
    let fx = Fx::new();
    let file = format!(
        "# Team rules\n\n{BEGIN}\n## MoAI\n{END}\n\n{}\n\nWrite tests first.\n{}\n",
        merged_local_block("note\n").trim_end_matches('\n'),
        secret()
    );
    let at = file.lines().position(|l| l.contains("sk-")).unwrap() + 1;
    assert_eq!(at, 12, "{file}");
    fx.write(AGENTS_MD, &file);
    fx.ready();
    let plan = fx.pick_plan("move");
    assert_eq!(plan["state"], json!("blocked"), "{plan}");
    assert!(reason_of(&plan).contains("(lines 12)"), "{plan}");

    // No removed region: the number is the line itself.
    let fx = Fx::new();
    fx.write(AGENTS_MD, &format!("{}\nWrite tests first.\n", secret()));
    fx.ready();
    let plan = fx.pick_plan("move");
    assert_eq!(plan["state"], json!("blocked"), "{plan}");
    assert!(reason_of(&plan).contains("(lines 1)"), "{plan}");
}

// --- REQ-4: the receipts line in .gitignore -------------------------------------------------------

/// The exact `.gitignore` after the apply when the project holds none before.
fn ignore_lines(fx: &Fx) -> Vec<String> {
    fx.read(".gitignore").lines().map(String::from).collect()
}

fn count_of(lines: &[String], line: &str) -> usize {
    lines.iter().filter(|l| *l == line).count()
}

#[test]
fn receipts_a_standalone_merge_ignores_the_receipt_folder() {
    let fx = Fx::new();
    fx.write("AGENTS.local.md", MINE);
    let plan = fx.merge_plan();
    assert_eq!(plan["gitignore"]["action"], json!("add"), "{plan}");
    let applied = fx.merge();
    assert_eq!(
        applied["writes"],
        json!([".gitignore", "CLAUDE.local.md"]),
        "{applied}"
    );
    assert_eq!(
        fx.read(".gitignore"),
        format!("CLAUDE.local.md\n{}\n", fx.fresh_receipts_line())
    );
    assert!(
        fx.exists(&fx.fresh_receipts()),
        "the apply wrote its receipt folder"
    );
    assert!(
        fx.ignored(&fx.fresh_receipts()),
        "git sees the receipt folder"
    );
    assert!(fx.ignored("CLAUDE.local.md"));
}

#[test]
fn receipts_a_pick_in_a_project_without_a_managed_block_ignores_the_receipt_folder() {
    for id in ["move", "import"] {
        let fx = Fx::new();
        fx.write(AGENTS_MD, TEAM);
        fx.write("AGENTS.local.md", MINE);
        fx.profile();
        assert!(!fx.exists(".gitignore"), "no managed block yet");
        fx.pick(id);
        assert!(
            fx.exists(RECEIPTS),
            "{id}: the pick wrote its receipt folder"
        );
        assert!(fx.ignored(RECEIPTS), "{id}: git sees the receipt folder");
        let lines = ignore_lines(&fx);
        assert!(
            lines
                .iter()
                .any(|l| l.starts_with(RECEIPTS) || l == ".moai/*"),
            "{id}: {lines:?}"
        );
        assert!(
            count_of(&lines, RECEIPTS_LINE) <= 1,
            "{id}: one receipts line at most: {lines:?}"
        );
    }
}

#[test]
fn receipts_a_folder_that_is_already_ignored_gets_no_second_line() {
    // A line of its own, and a wider rule.
    for wide in [false, true] {
        let fx = Fx::new();
        let earlier = if wide {
            format!("{}/\n", moai_cli::state_root::dir(&fx.root))
        } else {
            format!("target/\n{}\n", fx.fresh_receipts_line())
        };
        let expect = format!("{earlier}CLAUDE.local.md\n");
        fx.write("AGENTS.local.md", MINE);
        fx.write(".gitignore", &earlier);
        let plan = fx.merge_plan();
        assert_eq!(
            plan["gitignore"]["receipts"]["action"],
            json!("present"),
            "{plan}"
        );
        fx.merge();
        assert_eq!(fx.read(".gitignore"), expect);
        assert!(fx.ignored(&fx.fresh_receipts()));
    }
}

#[test]
fn receipts_a_project_that_is_not_a_git_repository_gets_no_gitignore() {
    let fx = Fx::plain();
    fx.write("AGENTS.local.md", MINE);
    let plan = fx.merge_plan();
    assert_eq!(plan["gitignore"]["action"], json!("skip"), "{plan}");
    let applied = fx.merge();
    assert_eq!(applied["writes"], json!(["CLAUDE.local.md"]), "{applied}");
    assert!(!fx.exists(".gitignore"));
    assert!(fx.exists(&fx.fresh_receipts()));
}

#[test]
fn receipts_the_revert_puts_gitignore_back_as_it_was() {
    let fx = Fx::new();
    fx.write("AGENTS.local.md", MINE);
    let earlier = "target/\nnode_modules";
    fx.write(".gitignore", earlier);
    let applied = fx.merge();
    assert_eq!(
        fx.read(".gitignore"),
        format!("{earlier}\nCLAUDE.local.md\n{}\n", fx.fresh_receipts_line())
    );
    let id = applied["receipt"].as_str().unwrap();
    let out = fx.call(
        &[
            "agents-merge-revert",
            fx.p(),
            "--receipt",
            id,
            "--permission-reference",
            REF,
        ],
        b"",
    );
    assert_eq!(out["gitignore"], json!("reverted"), "{out}");
    assert_eq!(fx.read(".gitignore"), earlier);
}

// --- REQ-5: the generated block never restates the check-run rules --------------------------------

/// The managed block the writer puts into a project holds facts only. The check-run rules live
/// once in the VERIFY workflow, so the block carries neither the citation line format nor a line
/// of the role table.
#[test]
fn check_run_the_generated_block_never_restates_the_rules() {
    let fx = Fx::new();
    fx.ready();
    let text = fx.read("CLAUDE.md");
    let begin = text.find(BEGIN).expect("begin marker");
    let end = text.find(END).expect("end marker");
    let block = &text[begin..end];
    assert!(block.lines().count() > 2, "the block is empty: {text}");
    for restated in [
        "ci: ci run=<run id> head=<40-hex head SHA>",
        "ci run=",
        "Delta auditor: merge-only delta",
        "Author: new or changed code",
        "Auditor: the audited head changes code",
        "Repair lane: the repair changes code",
    ] {
        assert!(
            !block.contains(restated),
            "the block restates {restated:?}: {block}"
        );
    }
}
