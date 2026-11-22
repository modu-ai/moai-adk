//! The picks `move`, `import` and `leave` that `moai update` offers a project that holds an
//! `AGENTS.md` or an `AGENTS.local.md` (SPEC-20261122-claude-md-project-picks), and how the pick is
//! written, refused and undone (SPEC-20261122-claude-md-project-layer, first slice).
//!
//! Every case runs the built binary in a temporary git repository with an empty temporary HOME.
//! The test names start with `picks_`, `shadow_`, `settled_`, `receipt_`, `refuse_` and
//! `baseline_`: the acceptance criteria of the two SPECs filter on those prefixes.

#![cfg(unix)]

mod common;
use common::git::git_ok;
use common::CleanEnv;
use moai_cli::agents_local_merge::{
    BEGIN as LOCAL_BEGIN, GUIDANCE_BEGIN, GUIDANCE_END, IMPORT_BEGIN, IMPORT_END, IMPORT_LINE,
};
use serde_json::{json, Value};
use std::{
    collections::BTreeMap,
    fs,
    io::Write,
    path::{Path, PathBuf},
    process::{Command, Stdio},
};
use tempfile::TempDir;

const REF: &str = "fixture-approved";
const PLACEHOLDER: &str = "<permission-reference>";
const BEGIN: &str = "<!-- moai:managed:begin id=moai-workflow -->";
const END: &str = "<!-- moai:managed:end id=moai-workflow -->";
/// What a person wrote in `AGENTS.md` and in `AGENTS.local.md`.
const AGENTS: &str = "# Team rules\n\nWrite tests first.\n";
const LOCAL: &str = "My own notes\nPrefer tabs.\n";
/// The text an earlier MoAI wrote into a new `AGENTS.local.md`: nothing the user wrote.
const GENERATED_LOCAL: &str = "# Personal supplement\n\nResponse language: English\n\n<!-- Personal file MoAI reads explicitly when invoked. Keep secrets out; never share it in the repo or in Issues. -->\n";
const PICKS: [&str; 3] = ["move", "import", "leave"];

struct Fx {
    _tmp: TempDir,
    root: PathBuf,
    home: PathBuf,
}

impl Fx {
    /// A git repository with nothing in it; write the files first, then call [`Fx::ready`].
    fn new() -> Fx {
        let tmp = TempDir::new().unwrap();
        let base = fs::canonicalize(tmp.path()).unwrap();
        let (root, home) = (base.join("project"), base.join("home"));
        fs::create_dir_all(&root).unwrap();
        fs::create_dir_all(&home).unwrap();
        git_ok(&root, &["init", "-q", "-b", "main"]);
        Fx {
            _tmp: tmp,
            root,
            home,
        }
    }

    /// A project that holds `AGENTS.md` and `AGENTS.local.md`, with its generation settled.
    fn both() -> Fx {
        let fx = Fx::new();
        fx.write("AGENTS.md", AGENTS);
        fx.write("AGENTS.local.md", LOCAL);
        fx.ready();
        fx
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

    fn exists(&self, rel: &str) -> bool {
        fs::symlink_metadata(self.root.join(rel)).is_ok()
    }

    /// The envelope of one run; the child's HOME is this fixture's.
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

    /// Applies a project profile and runs the generation step, so a later preview offers the picks
    /// instead of waiting for the install phase.
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

    /// The `agents-picks` step of a fresh preview.
    fn picks_step(&self) -> Value {
        step(&self.update_plan(), "agents-picks").clone()
    }

    /// Runs the command a step or one of its choices prints, the way a session does.
    fn run_printed(&self, with_apply: &Value) -> Value {
        let argv: Vec<String> = with_apply["apply"]["argv"]
            .as_array()
            .unwrap_or_else(|| panic!("no apply command: {with_apply}"))
            .iter()
            .map(|w| w.as_str().unwrap().to_string())
            .collect();
        assert_eq!(argv[0], "moai", "{argv:?}");
        assert_eq!(
            argv.iter().filter(|w| w.contains('<')).collect::<Vec<_>>(),
            vec![PLACEHOLDER],
            "the only placeholder is the permission reference: {argv:?}"
        );
        let args: Vec<&str> = argv[1..]
            .iter()
            .map(|w| if w == PLACEHOLDER { REF } else { w.as_str() })
            .collect();
        self.exec(
            &args,
            &serde_json::to_vec(&with_apply["apply"]["stdin"]).unwrap(),
        )
    }

    /// The printed command of one pick, run; the apply must succeed. Returns its result.
    fn pick(&self, id: &str) -> Value {
        let step = self.picks_step();
        let out = self.run_printed(choice(&step, id));
        assert_eq!(out["ok"], json!(true), "{id}: {out}");
        out["result"].clone()
    }

    /// `agents-merge-plan --pick <id>`: the whole envelope.
    fn merge_plan(&self, id: &str) -> Value {
        self.exec(&["agents-merge-plan", self.p(), "--pick", id], b"")
    }

    fn merge_apply(&self, id: &str, digest: &str) -> Value {
        self.exec(
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

    fn revert(&self, receipt: &str) -> Value {
        self.exec(
            &[
                "agents-merge-revert",
                self.p(),
                "--receipt",
                receipt,
                "--permission-reference",
                REF,
            ],
            b"",
        )
    }

    /// Every file and link of the project except `.git`, by path. The receipt store is left out
    /// unless `receipts`: a revert keeps its receipts, so they are not part of "the first bytes".
    fn snapshot(&self, receipts: bool) -> BTreeMap<String, Vec<u8>> {
        let mut out = BTreeMap::new();
        walk(&self.root, &self.root, receipts, &mut out);
        out
    }
}

fn walk(root: &Path, dir: &Path, receipts: bool, out: &mut BTreeMap<String, Vec<u8>>) {
    for entry in fs::read_dir(dir).unwrap() {
        let path = entry.unwrap().path();
        let rel = path
            .strip_prefix(root)
            .unwrap()
            .to_str()
            .unwrap()
            .to_string();
        if rel == ".git" || (!receipts && rel.ends_with("/receipts")) {
            continue;
        }
        let meta = fs::symlink_metadata(&path).unwrap();
        if meta.file_type().is_symlink() {
            let target = fs::read_link(&path).unwrap();
            out.insert(rel, format!("link:{}", target.display()).into_bytes());
        } else if meta.is_dir() {
            walk(root, &path, receipts, out);
        } else {
            out.insert(rel, fs::read(&path).unwrap());
        }
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

fn choice_ids(step: &Value) -> Vec<String> {
    step["preview"]["choices"]
        .as_array()
        .map(|all| {
            all.iter()
                .map(|c| c["id"].as_str().unwrap().to_string())
                .collect()
        })
        .unwrap_or_default()
}

fn choice<'a>(step: &'a Value, id: &str) -> &'a Value {
    step["preview"]["choices"]
        .as_array()
        .unwrap_or_else(|| panic!("no choices: {step}"))
        .iter()
        .find(|c| c["id"] == id)
        .unwrap_or_else(|| panic!("no choice {id}: {step}"))
}

fn warnings(step: &Value) -> String {
    step["preview"]["warnings"].to_string()
}

/// The managed block of `text` and everything after it.
fn after_block(text: &str) -> &str {
    let end = text.find(END).expect("managed block end") + END.len();
    &text[end..]
}

/// The two originals are never written.
fn assert_originals(fx: &Fx, agents: Option<&str>, local: Option<&str>) {
    assert_eq!(fx.exists("AGENTS.md"), agents.is_some());
    if let Some(bytes) = agents {
        assert_eq!(fx.read("AGENTS.md"), bytes, "AGENTS.md changed");
    }
    assert_eq!(fx.exists("AGENTS.local.md"), local.is_some());
    if let Some(bytes) = local {
        assert_eq!(fx.read("AGENTS.local.md"), bytes, "AGENTS.local.md changed");
    }
}

// --- the writer alone ----------------------------------------------------------------------------

/// A writer call that carries `import` or `move` without the picks step would write a `CLAUDE.md`
/// that holds only the managed block, and Claude Code would stop reading `AGENTS.md`. It writes
/// nothing and says the picks step carries the pick out.
#[test]
fn picks_a_direct_writer_call_with_import_or_move_writes_no_claude_md() {
    for pick in ["import", "move"] {
        let fx = Fx::new();
        fx.write("AGENTS.md", AGENTS);
        fx.profile();
        let update = fx.update_plan();
        let request = step(&update, "generation")["apply"]["stdin"]["request"].clone();
        assert!(request.is_object(), "no request in {update}");
        let mut request = request;
        request["picks"] = json!([pick]);
        let before = fx.snapshot(false);

        let plan = fx.call(
            &["generation-plan", fx.p()],
            &serde_json::to_vec(&request).unwrap(),
        );
        let rows = plan["files"].as_array().unwrap();
        assert!(
            rows.iter().all(|f| f["path"] != "CLAUDE.md"),
            "{pick}: the plan names CLAUDE.md: {plan}"
        );
        let said = plan["warnings"].to_string();
        for word in ["agents-merge", "--pick", pick] {
            assert!(said.contains(word), "{pick}: {word:?} not in {said}");
        }
        assert_eq!(before, fx.snapshot(false), "{pick}: the plan wrote");

        let digest = plan["plan_digest"].as_str().unwrap().to_string();
        fx.call(
            &[
                "generation-apply",
                fx.p(),
                "--plan-digest",
                &digest,
                "--permission-reference",
                REF,
            ],
            &serde_json::to_vec(&plan).unwrap(),
        );
        assert!(!fx.exists("CLAUDE.md"), "{pick}: a CLAUDE.md was written");
        assert_originals(&fx, Some(AGENTS), None);

        // The agents writer alone says the same and writes the same nothing.
        let agents_request = json!({
            "protocol_version": 1, "language": "en", "local_file": "none",
            "workflow_reference": request["workflow_reference"], "picks": [pick],
        });
        let agents = fx.call(
            &["agents-plan", fx.p()],
            &serde_json::to_vec(&agents_request).unwrap(),
        );
        assert!(
            agents["files"]
                .as_array()
                .unwrap()
                .iter()
                .all(|f| f["path"] != "CLAUDE.md"),
            "{pick}: {agents}"
        );
        assert!(
            agents["warnings"].to_string().contains("--pick"),
            "{agents}"
        );
        assert!(!fx.exists("CLAUDE.md"));
    }
}

// --- the preview ---------------------------------------------------------------------------------

#[test]
fn picks_preview_lists_move_import_leave_in_order_with_one_command_each_writes_nothing() {
    let fx = Fx::both();
    let before = fx.snapshot(true);
    let plan = fx.update_plan();
    let picks = step(&plan, "agents-picks");
    assert_eq!(picks["state"], "needs_decision", "{picks}");
    assert_eq!(choice_ids(picks), PICKS);
    let all = picks["preview"]["choices"].as_array().unwrap();
    let recommended: Vec<&str> = all
        .iter()
        .filter(|c| c["recommended"] == json!(true))
        .map(|c| c["id"].as_str().unwrap())
        .collect();
    assert_eq!(recommended, ["move"]);
    let mut digests = Vec::new();
    for c in all {
        let id = c["id"].as_str().unwrap();
        let argv: Vec<&str> = c["apply"]["argv"]
            .as_array()
            .unwrap()
            .iter()
            .map(|w| w.as_str().unwrap())
            .collect();
        assert_eq!(argv[..2], ["moai", "agents-merge-apply"], "{id}: {argv:?}");
        let at = argv.iter().position(|w| *w == "--pick").unwrap();
        assert_eq!(argv[at + 1], id, "{argv:?}");
        assert_eq!(argv.iter().filter(|w| w.contains('<')).count(), 1);
        let digest = c["plan_digest"].as_str().unwrap();
        assert!(digest.starts_with("sha256:"), "{c}");
        assert!(argv.contains(&digest), "{argv:?}");
        assert!(!c["effect"].as_str().unwrap().is_empty());
        digests.push(digest.to_string());
    }
    digests.sort();
    digests.dedup();
    assert_eq!(digests.len(), 3, "one digest per pick");
    // The first pick says the originals stay and may be deleted later.
    let move_effect = choice(picks, "move")["effect"].as_str().unwrap();
    for word in ["AGENTS.md", "AGENTS.local.md", "unchanged", "deleted"] {
        assert!(move_effect.contains(word), "{word:?} not in {move_effect}");
    }
    // While a pick is pending the standalone personal-file merge has nothing to offer.
    let local = step(&plan, "agents-local-merge");
    assert_eq!(local["state"], "skip", "{local}");
    assert!(local["apply"].is_null());
    assert_eq!(fx.snapshot(true), before, "the preview wrote something");
}

// --- the three picks -----------------------------------------------------------------------------

#[test]
fn picks_move_copies_text_below_the_block_and_local_text_into_claude_local_md() {
    let fx = Fx::both();
    let out = fx.pick("move");
    assert!(
        out["receipt"].as_str().unwrap().starts_with("agents-pick-"),
        "{out}"
    );
    let claude = fx.read("CLAUDE.md");
    assert!(claude.contains(BEGIN) && claude.contains(END), "{claude}");
    let below = after_block(&claude);
    assert!(below.contains(GUIDANCE_BEGIN), "{claude}");
    assert!(below.contains(GUIDANCE_END));
    assert!(below.contains("Write tests first."), "{claude}");
    assert!(!claude.contains("Prefer tabs."), "{claude}");
    assert!(!claude.contains(IMPORT_LINE), "{claude}");
    let local = fx.read("CLAUDE.local.md");
    assert!(local.contains(LOCAL_BEGIN), "{local}");
    assert!(local.contains("Prefer tabs."), "{local}");
    assert!(!local.contains("Write tests first."), "{local}");
    assert_originals(&fx, Some(AGENTS), Some(LOCAL));

    // A personal file MoAI itself generated is not copied.
    assert!(moai_cli::agents::is_generated_local(
        GENERATED_LOCAL.as_bytes()
    ));
    let fx = Fx::new();
    fx.write("AGENTS.md", AGENTS);
    fx.write("AGENTS.local.md", GENERATED_LOCAL);
    fx.ready();
    fx.pick("move");
    assert!(fx.read("CLAUDE.md").contains("Write tests first."));
    assert!(!fx.exists("CLAUDE.local.md"), "a generated file is copied");
    assert_originals(&fx, Some(AGENTS), Some(GENERATED_LOCAL));
}

#[test]
fn picks_import_writes_block_and_one_import_line_and_leaves_local_alone() {
    let fx = Fx::both();
    let out = fx.pick("import");
    assert!(
        out["receipt"].as_str().unwrap().starts_with("agents-pick-"),
        "{out}"
    );
    let claude = fx.read("CLAUDE.md");
    assert!(claude.contains(BEGIN) && claude.contains(END), "{claude}");
    let below = after_block(&claude);
    let begin = below.find(IMPORT_BEGIN).expect("import start marker");
    let end = below.find(IMPORT_END).expect("import end marker");
    assert!(begin < end);
    let marked: Vec<&str> = below[begin..end].lines().collect();
    assert_eq!(
        marked.iter().filter(|l| **l == IMPORT_LINE).count(),
        1,
        "{marked:?}"
    );
    assert_eq!(claude.matches(IMPORT_LINE).count(), 1, "{claude}");
    assert!(
        !claude.contains(GUIDANCE_BEGIN),
        "the text is imported, not copied"
    );
    assert!(!claude.contains("Write tests first."), "{claude}");
    assert!(
        !fx.exists("CLAUDE.local.md"),
        "import never writes the personal file"
    );
    assert_originals(&fx, Some(AGENTS), Some(LOCAL));
    // The line is not written twice.
    let again = fx.picks_step();
    assert_eq!(again["state"], "unchanged", "{again}");
    assert!(choice_ids(&again).is_empty());
}

#[test]
fn picks_leave_writes_nothing_and_says_agents_md_stays_read() {
    let fx = Fx::both();
    let before = fx.snapshot(true);
    let out = fx.pick("leave");
    assert_eq!(fx.snapshot(true), before, "leave wrote something");
    assert_eq!(out["writes"], json!([]), "{out}");
    let said = out.to_string();
    assert!(
        said.contains("AGENTS.md") && said.contains("keeps reading"),
        "{said}"
    );
    assert!(!fx.exists("CLAUDE.md") && !fx.exists("CLAUDE.local.md"));
    // Leaving is not remembered: the next update offers the picks again.
    let again = fx.picks_step();
    assert_eq!(choice_ids(&again), PICKS, "{again}");
}

#[test]
fn picks_unknown_pick_is_refused_and_writes_nothing() {
    let fx = Fx::both();
    let before = fx.snapshot(true);
    let plan = fx.merge_plan("sideways");
    assert_eq!(plan["ok"], json!(false), "{plan}");
    let digest = choice(&fx.picks_step(), "move")["plan_digest"]
        .as_str()
        .unwrap()
        .to_string();
    for bad in ["sideways", "Move", "", "move,import", "personal-file"] {
        let out = fx.merge_apply(bad, &digest);
        assert_eq!(out["ok"], json!(false), "{bad:?}: {out}");
    }
    // A digest that belongs to another pick is stale for this one.
    let out = fx.merge_apply("import", &digest);
    assert_eq!(out["ok"], json!(false), "{out}");
    assert_eq!(fx.snapshot(true), before, "a refused pick wrote something");
}

#[test]
fn picks_existing_claude_md_keeps_every_byte_and_adds_on_a_new_line() {
    for earlier in ["# Mine\nwith a final newline\n", "# Mine\nno final newline"] {
        for pick in ["move", "import"] {
            let fx = Fx::new();
            fx.write("AGENTS.md", AGENTS);
            fx.write("AGENTS.local.md", LOCAL);
            fx.write("CLAUDE.md", earlier);
            fx.ready();
            fx.pick(pick);
            let claude = fx.read("CLAUDE.md");
            assert!(claude.starts_with(earlier), "{pick}: {claude:?}");
            let added = &claude[earlier.len()..];
            assert!(!added.is_empty());
            // The first added byte starts a new line, also when the file had no final newline.
            let first_added_line = if earlier.ends_with('\n') {
                added
            } else {
                assert!(added.starts_with('\n'), "{pick}: {claude:?}");
                &added[1..]
            };
            assert!(
                first_added_line.starts_with("<!-- moai:") || first_added_line.starts_with('\n'),
                "{pick}: {claude:?}"
            );
            assert_eq!(claude.matches(BEGIN).count(), 1, "{claude}");
            assert_originals(&fx, Some(AGENTS), Some(LOCAL));
        }
    }
    // The personal file too.
    for earlier in ["mine\n", "mine without a final newline"] {
        let fx = Fx::new();
        fx.write("AGENTS.md", AGENTS);
        fx.write("AGENTS.local.md", LOCAL);
        fx.write("CLAUDE.local.md", earlier);
        fx.ready();
        fx.pick("move");
        let local = fx.read("CLAUDE.local.md");
        assert!(local.starts_with(earlier), "{local:?}");
        let begin = local.find(LOCAL_BEGIN).unwrap();
        assert!(local[..begin].ends_with('\n'), "{local:?}");
        assert!(local.contains("Prefer tabs."));
    }
}

#[test]
fn picks_only_agents_local_md_offers_move_and_leave() {
    let fx = Fx::new();
    fx.write("AGENTS.local.md", LOCAL);
    fx.ready();
    let picks = fx.picks_step();
    assert_eq!(picks["state"], "needs_decision", "{picks}");
    assert_eq!(choice_ids(&picks), ["move", "leave"], "{picks}");
    fx.pick("move");
    assert!(fx.read("CLAUDE.local.md").contains("Prefer tabs."));
    assert!(!fx.exists("AGENTS.md"));
    assert_originals(&fx, None, Some(LOCAL));
    // `import` would name a file that is not there.
    let fx = Fx::new();
    fx.write("AGENTS.local.md", LOCAL);
    fx.ready();
    let plan = fx.merge_plan("import");
    let digest = plan["result"]["plan_digest"].as_str().unwrap().to_string();
    let before = fx.snapshot(true);
    assert_eq!(fx.merge_apply("import", &digest)["ok"], json!(false));
    assert_eq!(fx.snapshot(true), before);
}

// --- the shadow warning --------------------------------------------------------------------------

#[test]
fn shadow_warns_for_a_regular_claude_file_without_the_import() {
    for file in ["CLAUDE.md", ".claude/CLAUDE.md", "CLAUDE.local.md"] {
        let fx = Fx::new();
        fx.write("AGENTS.md", AGENTS);
        fx.write(file, "Our own notes\n");
        fx.ready();
        let picks = fx.picks_step();
        let said = warnings(&picks);
        assert!(said.contains(file), "{file}: {said}");
        assert!(said.contains("Claude Code reads only"), "{file}: {said}");
    }
}

#[test]
fn shadow_is_silent_for_an_importing_file_a_link_and_a_settled_project() {
    // A file that imports AGENTS.md says nothing, while a regular one beside it still does.
    let fx = Fx::new();
    fx.write("AGENTS.md", AGENTS);
    fx.write(".claude/CLAUDE.md", "Notes\n@AGENTS.md\n");
    fx.write("CLAUDE.local.md", "Mine\n");
    fx.ready();
    let said = warnings(&fx.picks_step());
    assert!(!said.contains(".claude/CLAUDE.md"), "{said}");
    assert!(said.contains("CLAUDE.local.md"), "{said}");

    // A link is neither read nor written.
    let fx = Fx::new();
    fx.write("AGENTS.md", AGENTS);
    fx.write("elsewhere.md", "Elsewhere\n");
    std::os::unix::fs::symlink("elsewhere.md", fx.root.join("CLAUDE.local.md")).unwrap();
    fx.ready();
    let said = warnings(&fx.picks_step());
    assert!(!said.contains("CLAUDE.local.md"), "{said}");

    // A settled project: the file imports, or already holds the text.
    for claude in [
        "Notes\n@AGENTS.md\n",
        "Intro\n\n# Team rules\n\nWrite tests first.\n",
    ] {
        let fx = Fx::new();
        fx.write("AGENTS.md", AGENTS);
        fx.write("CLAUDE.md", claude);
        fx.ready();
        let picks = fx.picks_step();
        assert!(
            !warnings(&picks).contains("Claude Code reads only"),
            "{picks}"
        );
    }
}

// --- a settled project ---------------------------------------------------------------------------

#[test]
fn settled_move_adds_the_ignore_rule_in_the_same_apply() {
    let fx = Fx::both();
    let out = fx.pick("move");
    let ignore = fx.read(".gitignore");
    assert_eq!(
        ignore.lines().filter(|l| *l == "CLAUDE.local.md").count(),
        1,
        "{ignore}"
    );
    assert!(out["writes"].to_string().contains(".gitignore"), "{out}");
    // A rule the project already has is not added a second time.
    let fx = Fx::new();
    fx.write("AGENTS.md", AGENTS);
    fx.write("AGENTS.local.md", LOCAL);
    fx.write(".gitignore", "target/\nCLAUDE.local.md\n");
    fx.ready();
    fx.pick("move");
    let ignore = fx.read(".gitignore");
    assert_eq!(
        ignore.lines().filter(|l| *l == "CLAUDE.local.md").count(),
        1,
        "{ignore}"
    );
}

#[test]
fn settled_project_with_the_import_line_is_offered_no_pick() {
    let fx = Fx::new();
    fx.write("AGENTS.md", AGENTS);
    fx.write("AGENTS.local.md", LOCAL);
    fx.write("CLAUDE.md", "# Mine\n\n@AGENTS.md\n");
    fx.ready();
    let picks = fx.picks_step();
    assert_eq!(picks["state"], "unchanged", "{picks}");
    assert!(choice_ids(&picks).is_empty(), "{picks}");
    assert!(picks["apply"].is_null());
}

#[test]
fn settled_project_with_the_copied_text_is_offered_no_pick() {
    // The merged block a `move` wrote.
    let fx = Fx::both();
    fx.pick("move");
    let picks = fx.picks_step();
    assert_eq!(picks["state"], "unchanged", "{picks}");
    assert!(choice_ids(&picks).is_empty());
    // Text pasted by hand, without markers, among other text.
    let fx = Fx::new();
    fx.write("AGENTS.md", AGENTS);
    fx.write(
        "CLAUDE.md",
        "Intro\n\n# Team rules\n\nWrite tests first.\n\nOutro\n",
    );
    fx.ready();
    let picks = fx.picks_step();
    assert_eq!(picks["state"], "unchanged", "{picks}");
    assert!(choice_ids(&picks).is_empty(), "{picks}");
    // Only a part of the text is not the text.
    let fx = Fx::new();
    fx.write("AGENTS.md", AGENTS);
    fx.write("CLAUDE.md", "Intro\n\n# Team rules\n");
    fx.ready();
    assert_eq!(choice_ids(&fx.picks_step()), PICKS);
}

// --- the receipt ---------------------------------------------------------------------------------

#[test]
fn receipt_revert_restores_first_bytes_after_move_and_import_and_removes_created_files() {
    for pick in ["move", "import"] {
        for existing in [false, true] {
            let fx = Fx::new();
            fx.write("AGENTS.md", AGENTS);
            fx.write("AGENTS.local.md", LOCAL);
            fx.write(".gitignore", "target/");
            if existing {
                fx.write("CLAUDE.md", "# Mine\nno final newline");
                fx.write("CLAUDE.local.md", "mine too");
            }
            fx.ready();
            let first = fx.snapshot(false);
            let out = fx.pick(pick);
            assert_ne!(fx.snapshot(false), first, "{pick}: nothing was written");
            let receipt = out["receipt"].as_str().unwrap().to_string();
            let reverted = fx.revert(&receipt);
            assert_eq!(reverted["ok"], json!(true), "{pick}: {reverted}");
            assert_eq!(
                reverted["result"]["state"], "reverted",
                "{pick}: {reverted}"
            );
            assert_eq!(fx.snapshot(false), first, "{pick} existing={existing}");
            if !existing {
                assert!(!fx.exists("CLAUDE.md") && !fx.exists("CLAUDE.local.md"));
            }
            // A second revert changes nothing.
            let again = fx.revert(&receipt);
            assert_eq!(again["ok"], json!(true), "{again}");
            assert_eq!(fx.snapshot(false), first);
        }
    }
}

#[test]
fn receipt_revert_skips_a_file_changed_after_the_apply_and_names_it() {
    // The personal file changed: it stays, everything else goes back.
    let fx = Fx::both();
    let first = fx.snapshot(false);
    let receipt = fx.pick("move")["receipt"].as_str().unwrap().to_string();
    let edited = format!("{}and one line more\n", fx.read("CLAUDE.local.md"));
    fx.write("CLAUDE.local.md", &edited);
    let out = fx.revert(&receipt);
    assert_eq!(out["ok"], json!(true), "{out}");
    assert!(
        out["result"]["kept"]
            .to_string()
            .contains("CLAUDE.local.md"),
        "{out}"
    );
    assert_eq!(
        fx.read("CLAUDE.local.md"),
        edited,
        "the edited file was touched"
    );
    assert!(!fx.exists("CLAUDE.md"), "the other files go back");
    assert_originals(&fx, Some(AGENTS), Some(LOCAL));
    let now = fx.snapshot(false);
    for (path, bytes) in &first {
        if path != ".gitignore" {
            assert_eq!(now.get(path), Some(bytes), "{path}");
        }
    }

    // The team file changed: it stays, the personal file goes.
    let fx = Fx::both();
    let receipt = fx.pick("move")["receipt"].as_str().unwrap().to_string();
    let edited = format!("{}and one line more\n", fx.read("CLAUDE.md"));
    fx.write("CLAUDE.md", &edited);
    let out = fx.revert(&receipt);
    assert_eq!(out["ok"], json!(true), "{out}");
    assert!(
        out["result"]["kept"].to_string().contains("CLAUDE.md"),
        "{out}"
    );
    assert_eq!(fx.read("CLAUDE.md"), edited);
    assert!(!fx.exists("CLAUDE.local.md"));
    assert_originals(&fx, Some(AGENTS), Some(LOCAL));
}

// --- what is refused -----------------------------------------------------------------------------

#[test]
fn refuse_claude_md_link_to_agents_md_writes_nothing_and_agents_md_stays() {
    // CLAUDE.md is a link to AGENTS.md: nothing may be written through it.
    for pick in ["move", "import"] {
        let fx = Fx::new();
        fx.write("AGENTS.md", AGENTS);
        fx.write("AGENTS.local.md", LOCAL);
        std::os::unix::fs::symlink("AGENTS.md", fx.root.join("CLAUDE.md")).unwrap();
        fx.ready();
        let before = fx.snapshot(true);
        let plan = fx.merge_plan(pick);
        assert_eq!(plan["ok"], json!(true), "{plan}");
        assert_eq!(plan["result"]["state"], "blocked", "{plan}");
        assert!(plan["result"].to_string().contains("CLAUDE.md"), "{plan}");
        let digest = plan["result"]["plan_digest"].as_str().unwrap().to_string();
        let out = fx.merge_apply(pick, &digest);
        assert_eq!(out["ok"], json!(false), "{pick}: {out}");
        assert!(out.to_string().contains("CLAUDE.md"), "{out}");
        assert_eq!(
            fx.snapshot(true),
            before,
            "{pick}: a refused pick wrote something"
        );
        assert_originals(&fx, Some(AGENTS), Some(LOCAL));
        // The step does not offer what it would refuse; leaving stays possible.
        let picks = fx.picks_step();
        assert_eq!(choice_ids(&picks), ["leave"], "{picks}");
        assert_eq!(picks["state"], "conflict", "{picks}");
        assert!(picks.to_string().contains("CLAUDE.md"), "{picks}");
    }

    // CLAUDE.md is not UTF-8 text.
    let fx = Fx::new();
    fx.write("AGENTS.md", AGENTS);
    fs::write(fx.root.join("CLAUDE.md"), [0xff, 0xfe, b'\n']).unwrap();
    fx.ready();
    let before = fx.snapshot(true);
    let plan = fx.merge_plan("move");
    assert_eq!(plan["result"]["state"], "blocked", "{plan}");
    let digest = plan["result"]["plan_digest"].as_str().unwrap().to_string();
    let out = fx.merge_apply("move", &digest);
    assert_eq!(out["ok"], json!(false), "{out}");
    assert!(out.to_string().contains("CLAUDE.md"), "{out}");
    assert_eq!(fx.snapshot(true), before);

    // CLAUDE.local.md is a link, or not UTF-8: the move that would write it is refused.
    for local_is_link in [true, false] {
        let fx = Fx::new();
        fx.write("AGENTS.md", AGENTS);
        fx.write("AGENTS.local.md", LOCAL);
        fx.write("elsewhere.md", "kept\n");
        if local_is_link {
            std::os::unix::fs::symlink("elsewhere.md", fx.root.join("CLAUDE.local.md")).unwrap();
        } else {
            fs::write(fx.root.join("CLAUDE.local.md"), [0xff, 0xfe, b'\n']).unwrap();
        }
        fx.ready();
        let before = fx.snapshot(true);
        let plan = fx.merge_plan("move");
        assert_eq!(plan["result"]["state"], "blocked", "{plan}");
        let digest = plan["result"]["plan_digest"].as_str().unwrap().to_string();
        let out = fx.merge_apply("move", &digest);
        assert_eq!(out["ok"], json!(false), "{out}");
        assert!(out.to_string().contains("CLAUDE.local.md"), "{out}");
        assert_eq!(fx.snapshot(true), before, "link={local_is_link}");
        assert_eq!(fx.read("elsewhere.md"), "kept\n");
        // `import` writes only CLAUDE.md, so it stays possible.
        assert_eq!(
            choice_ids(&fx.picks_step()),
            ["import", "leave"],
            "link={local_is_link}"
        );
    }
}

// --- the baseline --------------------------------------------------------------------------------

#[test]
fn baseline_next_generation_plan_reads_the_block_unchanged_after_each_pick() {
    for pick in PICKS {
        let fx = Fx::both();
        fx.pick(pick);
        let plan = fx.update_plan();
        let generation = step(&plan, "generation");
        assert_eq!(generation["state"], "unchanged", "{pick}: {generation}");
        let rows = generation["preview"]["files"].as_array().unwrap();
        assert!(
            rows.iter().all(|f| f["action"] == "unchanged"),
            "{pick}: {rows:?}"
        );
        if pick != "leave" {
            let claude = rows.iter().find(|f| f["path"] == "CLAUDE.md").unwrap();
            assert_eq!(claude["action"], "unchanged", "{pick}: {claude}");
        }
    }
}
