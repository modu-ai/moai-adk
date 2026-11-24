//! The picks `move`, `leave` and `remove-originals` that `moai update` offers a project that holds
//! an `AGENTS.md` or an `AGENTS.local.md` (SPEC-20261122-claude-md-project-picks,
//! SPEC-20261124-claude-md-product-next), and how the pick is written, refused and undone
//! (SPEC-20261122-claude-md-project-layer, first slice).
//!
//! Every case runs the built binary in a temporary git repository with an empty temporary HOME.
//! The test names start with `picks_`, `shadow_`, `settled_`, `remove_`, `receipt_`, `refuse_` and
//! `baseline_`: the acceptance criteria of the SPECs filter on those prefixes.

#![cfg(unix)]

mod common;
use common::git::git_ok;
use common::CleanEnv;
use moai_cli::agents_local_merge::{
    BEGIN as LOCAL_BEGIN, END as LOCAL_END, GUIDANCE_BEGIN, GUIDANCE_END,
};
use serde_json::{json, Value};
use std::{
    collections::BTreeMap,
    fs,
    io::Write,
    os::unix::fs::PermissionsExt,
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
const PICKS: [&str; 2] = ["move", "leave"];
const REMOVE: &str = "remove-originals";
/// The line of the retired pick `import`: a project written by an earlier build may hold it.
const IMPORT_LINE: &str = "@AGENTS.md";
const IMPORT_BEGIN: &str = "<!-- moai:merged-import sha256=";
const IMPORT_END: &str = "<!-- /moai:merged-import -->";

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

    /// The unix permission bits of a file of the project.
    fn mode(&self, rel: &str) -> u32 {
        fs::metadata(self.root.join(rel))
            .unwrap()
            .permissions()
            .mode()
            & 0o7777
    }

    fn chmod(&self, rel: &str, mode: u32) {
        fs::set_permissions(self.root.join(rel), fs::Permissions::from_mode(mode)).unwrap();
    }

    /// The state folder of the project, as the resolver answers it.
    fn state_dir(&self) -> PathBuf {
        self.root.join(moai_cli::state_root::dir(&self.root))
    }

    /// A project whose originals were moved: `AGENTS.md` and `AGENTS.local.md` hold text, the
    /// pick `move` ran, and the originals carry the two unusual modes of the test.
    fn moved() -> Fx {
        let fx = Fx::new();
        fx.write("AGENTS.md", AGENTS);
        fx.write("AGENTS.local.md", LOCAL);
        fx.chmod("AGENTS.md", 0o755);
        fx.chmod("AGENTS.local.md", 0o640);
        fx.ready();
        fx.pick("move");
        fx
    }

    /// The printed command of `remove-originals` in the preview, or a panic when it is not offered.
    fn remove_choice(&self) -> Value {
        choice(&self.picks_step(), REMOVE).clone()
    }

    /// The step state and the choice ids of a fresh preview.
    fn picks_state(&self) -> (String, Vec<String>) {
        let step = self.picks_step();
        (
            step["state"].as_str().unwrap().to_string(),
            choice_ids(&step),
        )
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

/// A writer call that carries `move` without the picks step would write a `CLAUDE.md` that holds
/// only the managed block, and Claude Code would stop reading `AGENTS.md`. It writes nothing and
/// says the picks step carries the pick out.
#[test]
fn picks_a_direct_writer_call_with_move_writes_no_claude_md() {
    let pick = "move";
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

/// The retired pick `import` is an unknown pick everywhere a pick is given: the project picks, a
/// generation request, the agents request. Each refusal writes nothing.
#[test]
fn picks_import_is_refused_as_an_unknown_pick_and_writes_nothing() {
    let fx = Fx::both();
    let before = fx.snapshot(true);
    let digest = choice(&fx.picks_step(), "move")["plan_digest"]
        .as_str()
        .unwrap()
        .to_string();
    assert_eq!(fx.merge_plan("import")["ok"], json!(false));
    assert_eq!(fx.merge_apply("import", &digest)["ok"], json!(false));
    assert!(fx
        .merge_plan("import")
        .to_string()
        .contains("remove-originals"));

    // A generation request and an agents request that carry the pick: a project whose generation
    // is still to run prints the request.
    let other = Fx::new();
    other.write("AGENTS.md", AGENTS);
    other.profile();
    let update = other.update_plan();
    let mut request = step(&update, "generation")["apply"]["stdin"]["request"].clone();
    assert!(request.is_object(), "no request in {update}");
    request["picks"] = json!(["import"]);
    let kept = other.snapshot(true);
    let out = other.exec(
        &["generation-plan", other.p()],
        &serde_json::to_vec(&request).unwrap(),
    );
    assert_eq!(out["ok"], json!(false), "{out}");
    let agents_request = json!({
        "protocol_version": 1, "language": "en", "local_file": "none",
        "workflow_reference": request["workflow_reference"], "picks": ["import"],
    });
    let out = other.exec(
        &["agents-plan", other.p()],
        &serde_json::to_vec(&agents_request).unwrap(),
    );
    assert_eq!(out["ok"], json!(false), "{out}");
    assert_eq!(
        other.snapshot(true),
        kept,
        "a refused request wrote something"
    );
    assert_eq!(fx.snapshot(true), before, "a refused pick wrote something");
    assert!(!fx.exists("CLAUDE.md"));
    assert_originals(&fx, Some(AGENTS), Some(LOCAL));
}

// --- the preview ---------------------------------------------------------------------------------

#[test]
fn picks_preview_lists_move_leave_in_order_with_one_command_each_writes_nothing() {
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
    assert_eq!(digests.len(), 2, "one digest per pick");
    // The first pick says the originals stay unless the separate pick `remove-originals` is chosen.
    let move_effect = choice(picks, "move")["effect"].as_str().unwrap();
    for word in [
        "AGENTS.md",
        "AGENTS.local.md",
        "unchanged",
        "deleted",
        "remove-originals",
    ] {
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
    for bad in [
        "sideways",
        "Move",
        "",
        "move,leave",
        "personal-file",
        "import",
    ] {
        let out = fx.merge_apply(bad, &digest);
        assert_eq!(out["ok"], json!(false), "{bad:?}: {out}");
    }
    // A digest that belongs to another pick is stale for this one.
    let out = fx.merge_apply("leave", &digest);
    assert_eq!(out["ok"], json!(false), "{out}");
    assert_eq!(fx.snapshot(true), before, "a refused pick wrote something");
}

#[test]
fn picks_existing_claude_md_keeps_every_byte_and_adds_on_a_new_line() {
    for earlier in ["# Mine\nwith a final newline\n", "# Mine\nno final newline"] {
        let pick = "move";
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
}

#[test]
fn refuse_only_agents_local_md_with_move_blocked_is_a_conflict_not_a_decision() {
    // Only AGENTS.local.md exists, so the step lists `move` and `leave`. When CLAUDE.md
    // is not UTF-8 text, `move` is refused and nothing that writes is left: the state is a
    // conflict.
    let fx = Fx::new();
    fx.write("AGENTS.local.md", LOCAL);
    fx.ready();
    fs::write(fx.root.join("CLAUDE.md"), [0xff, 0xfe, b'\n']).unwrap();
    let picks = fx.picks_step();
    assert_eq!(choice_ids(&picks), ["leave"], "{picks}");
    assert_eq!(picks["state"], "conflict", "{picks}");
    assert!(picks.to_string().contains("CLAUDE.md"), "{picks}");
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

/// Once the text is copied, the step offers the one pick `remove-originals` and nothing else; the
/// step reads `optional`, so the preview still reads as up to date for a person who keeps the
/// originals.
#[test]
fn settled_project_with_the_copied_text_is_offered_only_remove_originals() {
    // The merged block a `move` wrote.
    let fx = Fx::both();
    fx.pick("move");
    let picks = fx.picks_step();
    assert_eq!(picks["state"], "optional", "{picks}");
    assert_eq!(choice_ids(&picks), [REMOVE], "{picks}");
    let remove = choice(&picks, REMOVE);
    assert_eq!(remove["recommended"], json!(false), "{remove}");
    assert_eq!(remove["deletes"], json!(["AGENTS.md", "AGENTS.local.md"]));
    let argv: Vec<&str> = remove["apply"]["argv"]
        .as_array()
        .unwrap()
        .iter()
        .map(|w| w.as_str().unwrap())
        .collect();
    assert_eq!(argv[..2], ["moai", "agents-merge-apply"], "{argv:?}");
    let at = argv.iter().position(|w| *w == "--pick").unwrap();
    assert_eq!(argv[at + 1], REMOVE, "{argv:?}");
    assert!(argv.contains(&remove["plan_digest"].as_str().unwrap()));
    let effect = remove["effect"].as_str().unwrap();
    for word in ["delete", "AGENTS.md", "AGENTS.local.md", "receipt", "bytes"] {
        assert!(effect.contains(word), "{word:?} not in {effect}");
    }
    // Every original carries one verdict, and the preview wrote nothing.
    let originals = remove["parts"][0]["originals"].as_array().unwrap();
    assert_eq!(originals.len(), 2, "{remove}");
    assert_eq!(originals[0]["path"], "AGENTS.md", "{remove}");
    assert_eq!(originals[0]["found_in"], "CLAUDE.md", "{remove}");
    assert_eq!(originals[1]["found_in"], "CLAUDE.local.md", "{remove}");
    assert!(originals
        .iter()
        .all(|o| o["verdict"] == "found" && o["sha256"].as_str().unwrap().starts_with("sha256:")));
    // The update reads as up to date once its other steps ran: the pick is optional.
    let settings = step(&fx.update_plan(), "claude-settings").clone();
    if settings["state"] == "ready" {
        let out = fx.run_printed(&settings);
        assert_eq!(out["ok"], json!(true), "{out}");
    }
    let plan = fx.update_plan();
    assert_eq!(plan["all_unchanged"], json!(true), "{plan}");
    assert_eq!(step(&plan, "agents-picks")["state"], "optional", "{plan}");

    // Text pasted by hand, without markers, among other text.
    let fx = Fx::new();
    fx.write("AGENTS.md", AGENTS);
    fx.write(
        "CLAUDE.md",
        "Intro\n\n# Team rules\n\nWrite tests first.\n\nOutro\n",
    );
    fx.ready();
    let picks = fx.picks_step();
    assert_eq!(picks["state"], "optional", "{picks}");
    assert_eq!(choice_ids(&picks), [REMOVE], "{picks}");
    assert_eq!(
        choice(&picks, REMOVE)["deletes"],
        json!(["AGENTS.md"]),
        "{picks}"
    );
    // Only a part of the text is not the text: nothing is offered to delete.
    let fx = Fx::new();
    fx.write("AGENTS.md", AGENTS);
    fx.write("CLAUDE.md", "Intro\n\n# Team rules\n");
    fx.ready();
    assert_eq!(choice_ids(&fx.picks_step()), PICKS);
}

/// The block an earlier generation merged into `AGENTS.md`: the personal text between its markers.
fn merged_local_block(text: &str) -> String {
    let sha = moai_cli::digest(text.as_bytes());
    let hex = sha.trim_start_matches("sha256:");
    format!("{LOCAL_BEGIN}{hex} -->\n{text}{LOCAL_END}\n")
}

/// A text that is not whole in its CLAUDE file keeps `remove-originals` from being offered, and
/// nothing can delete it, also when the plan of the text move reads `unchanged`.
#[test]
fn settled_text_not_whole_in_its_claude_file_is_never_offered_for_deletion() {
    // (1) The personal text sits only in the `moai:merged-local` block of AGENTS.md, where an
    // earlier generation put it: no CLAUDE file holds it, but the plan of the move reads unchanged.
    let fx = Fx::new();
    fx.write(
        "AGENTS.md",
        &format!(
            "{AGENTS}\n{}",
            merged_local_block("My own notes\nPrefer tabs.\n")
        ),
    );
    fx.ready();
    fx.pick("move");
    assert!(!fx.exists("CLAUDE.local.md"));
    let picks = fx.picks_step();
    assert_eq!(picks["state"], "unchanged", "{picks}");
    assert!(choice_ids(&picks).is_empty(), "{picks}");
    let before = fx.snapshot(true);
    let plan = fx.merge_plan(REMOVE);
    assert_ne!(plan["result"]["state"], "ready", "{plan}");
    let digest = plan["result"]["plan_digest"].as_str().unwrap().to_string();
    assert_eq!(fx.merge_apply(REMOVE, &digest)["ok"], json!(false));
    assert_eq!(fx.snapshot(true), before, "a refused pick wrote something");
    assert!(fx.read("AGENTS.md").contains("Prefer tabs."));
    // The same project is offered the pick once CLAUDE.local.md holds that text whole.
    fx.write("CLAUDE.local.md", "My own notes\nPrefer tabs.\n");
    let picks = fx.picks_step();
    assert_eq!(choice_ids(&picks), [REMOVE], "{picks}");

    // (2) AGENTS.local.md holds text that is only in the merged block of AGENTS.md.
    let fx = Fx::new();
    fx.write(
        "AGENTS.md",
        &format!("{AGENTS}\n{}", merged_local_block(LOCAL)),
    );
    fx.write("AGENTS.local.md", LOCAL);
    fx.ready();
    fx.pick("move");
    assert!(!fx.exists("CLAUDE.local.md"));
    let picks = fx.picks_step();
    assert_eq!(picks["state"], "unchanged", "{picks}");
    assert!(choice_ids(&picks).is_empty(), "{picks}");
    assert!(fx.exists("AGENTS.md") && fx.exists("AGENTS.local.md"));

    // (3) CLAUDE.local.md holds only part of the personal text, or none of it.
    let fx = Fx::moved();
    fx.write("CLAUDE.local.md", "My own notes\n");
    assert!(!fx.picks_state().1.contains(&REMOVE.to_string()));
    fx.write("CLAUDE.local.md", "unrelated\n");
    assert!(!fx.picks_state().1.contains(&REMOVE.to_string()));
}

/// A CLAUDE file that imports `@AGENTS.md` or links to it holds no copy of the text.
#[test]
fn settled_a_claude_file_that_imports_or_links_agents_md_gets_no_remove_originals() {
    let fx = Fx::moved();
    assert_eq!(fx.picks_state().1, [REMOVE]);
    fx.write(".claude/CLAUDE.md", "Notes\n@AGENTS.md\n");
    let (state, ids) = fx.picks_state();
    assert_eq!(state, "unchanged");
    assert!(ids.is_empty(), "{ids:?}");

    let fx = Fx::moved();
    fs::create_dir_all(fx.root.join(".claude")).unwrap();
    std::os::unix::fs::symlink("../AGENTS.md", fx.root.join(".claude/CLAUDE.md")).unwrap();
    let (state, ids) = fx.picks_state();
    assert_eq!(state, "unchanged");
    assert!(ids.is_empty(), "{ids:?}");
    let plan = fx.merge_plan(REMOVE);
    assert_ne!(plan["result"]["state"], "ready", "{plan}");
    assert!(fx.exists("AGENTS.md") && fx.exists("AGENTS.local.md"));
}

/// A blank or moai-generated `AGENTS.local.md` holds no text to find, and an `AGENTS.md` that holds
/// only the old managed block needs the managed block in `CLAUDE.md`.
#[test]
fn settled_blank_or_generated_originals_do_not_hold_the_pick_back() {
    for local in [GENERATED_LOCAL, "\n  \n"] {
        let fx = Fx::new();
        fx.write("AGENTS.md", AGENTS);
        fx.write("AGENTS.local.md", local);
        fx.ready();
        fx.pick("move");
        assert!(!fx.exists("CLAUDE.local.md"), "nothing is copied");
        let picks = fx.picks_step();
        assert_eq!(choice_ids(&picks), [REMOVE], "{picks}");
        let remove = choice(&picks, REMOVE);
        assert_eq!(remove["deletes"], json!(["AGENTS.md", "AGENTS.local.md"]));
        let local_row = &remove["parts"][0]["originals"][1];
        assert_eq!(local_row["verdict"], "no_text", "{remove}");
        assert!(local_row["found_in"].is_null(), "{remove}");
    }
    // An AGENTS.md of only a title and the old managed block.
    let fx = Fx::new();
    fx.write(
        "AGENTS.md",
        &format!("# demo\n\n{BEGIN}\n## MoAI\nOld.\n{END}\n"),
    );
    fx.ready();
    assert_eq!(choice_ids(&fx.picks_step()), PICKS);
    fx.pick("move");
    let picks = fx.picks_step();
    assert_eq!(choice_ids(&picks), [REMOVE], "{picks}");
}

/// The preview names the old managed section of `AGENTS.md` (its marker id and its line count) as
/// text that no CLAUDE file receives.
#[test]
fn settled_the_preview_names_the_old_managed_section_and_its_lines() {
    let fx = Fx::new();
    fx.write(
        "AGENTS.md",
        &format!("# Team rules\n\n{BEGIN}\n## MoAI\nGenerated.\n{END}\n\nWrite tests first.\n"),
    );
    fx.ready();
    fx.pick("move");
    let picks = fx.picks_step();
    assert_eq!(choice_ids(&picks), [REMOVE], "{picks}");
    let said = warnings(&picks);
    for word in ["moai-workflow", "4 lines", "no CLAUDE file", "AGENTS.md"] {
        assert!(said.contains(word), "{word:?} not in {said}");
    }
    let remove = choice(&picks, REMOVE);
    let listed = remove["parts"][0]["not_copied"].as_array().unwrap();
    assert_eq!(listed.len(), 1, "{remove}");
    assert_eq!(listed[0]["path"], "AGENTS.md", "{remove}");
    assert_eq!(listed[0]["section"], "moai-workflow", "{remove}");
    assert_eq!(listed[0]["lines"], json!(4), "{remove}");
    assert!(
        remove["warnings"].to_string().contains("moai-workflow"),
        "{remove}"
    );
    // A project whose AGENTS.md holds no such section names none.
    let fx = Fx::both();
    fx.pick("move");
    let picks = fx.picks_step();
    assert!(!warnings(&picks).contains("moai-workflow"), "{picks}");
}

// --- remove-originals ----------------------------------------------------------------------------

/// Before the text is copied nothing may be deleted: the pick is not offered, and a direct call is
/// refused with every file as it was.
#[test]
fn remove_is_not_offered_before_the_text_is_copied_and_a_direct_call_is_refused() {
    let fx = Fx::both();
    assert_eq!(choice_ids(&fx.picks_step()), PICKS);
    let before = fx.snapshot(true);
    let plan = fx.merge_plan(REMOVE);
    assert_ne!(plan["result"]["state"], "ready", "{plan}");
    let digest = plan["result"]["plan_digest"].as_str().unwrap().to_string();
    let out = fx.merge_apply(REMOVE, &digest);
    assert_eq!(out["ok"], json!(false), "{out}");
    assert_eq!(fx.snapshot(true), before);
    assert_originals(&fx, Some(AGENTS), Some(LOCAL));
}

#[test]
fn remove_deletes_both_originals_keeps_bytes_and_modes_in_an_owner_only_folder() {
    let fx = Fx::moved();
    let before = fx.snapshot(false);
    let out = fx.pick(REMOVE);
    assert_eq!(out["state"], "applied", "{out}");
    assert_eq!(
        out["deletes"],
        json!(["AGENTS.md", "AGENTS.local.md"]),
        "{out}"
    );
    assert_eq!(out["unchanged"], json!([]), "{out}");
    assert!(!fx.exists("AGENTS.md") && !fx.exists("AGENTS.local.md"));
    // Nothing else changed: the managed block, the copied text and `.gitignore` stand as they were.
    let mut expected = before.clone();
    expected.remove("AGENTS.md");
    expected.remove("AGENTS.local.md");
    assert_eq!(fx.snapshot(false), expected);

    // The receipt keeps the bytes, the modes and the digests in an owner-only folder.
    let id = out["receipt"].as_str().unwrap();
    let dir = fx.state_dir().join("receipts").join(id);
    let bits = |path: &Path| fs::metadata(path).unwrap().permissions().mode() & 0o777;
    assert_eq!(bits(&dir.join("originals")), 0o700);
    for (name, text) in [("AGENTS.md", AGENTS), ("AGENTS.local.md", LOCAL)] {
        let copy = dir.join("originals").join(name);
        assert_eq!(fs::read_to_string(&copy).unwrap(), text);
        assert_eq!(bits(&copy), 0o600, "{name}");
    }
    assert_eq!(bits(&dir.join("receipt.json")), 0o600);
    let receipt: Value =
        serde_json::from_slice(&fs::read(dir.join("receipt.json")).unwrap()).unwrap();
    assert_eq!(receipt["status"], "applied", "{receipt}");
    assert_eq!(receipt["pick"], REMOVE, "{receipt}");
    let part = receipt["parts"]
        .as_array()
        .unwrap()
        .iter()
        .find(|p| p["part"] == "originals")
        .unwrap_or_else(|| panic!("no originals part: {receipt}"));
    let files = part["files"].as_array().unwrap();
    assert_eq!(files.len(), 2, "{part}");
    for (row, (path, text, mode)) in files.iter().zip([
        ("AGENTS.md", AGENTS, 0o755u64),
        ("AGENTS.local.md", LOCAL, 0o640),
    ]) {
        assert_eq!(row["path"], path, "{row}");
        assert_eq!(row["sha256"], moai_cli::digest(text.as_bytes()), "{row}");
        assert_eq!(row["bytes"], json!(text.len()), "{row}");
        assert_eq!(row["mode"], json!(mode), "{row}");
    }
    // Asked again, the update has nothing left to offer.
    let again = fx.picks_step();
    assert_eq!(again["state"], "skip", "{again}");
}

/// The receipts folder is ignored by the same apply, through the receipt of the pick.
#[test]
fn remove_adds_the_receipts_ignore_rule_in_the_same_apply_and_the_revert_takes_it_back() {
    let fx = Fx::moved();
    // The earlier `move` already ignores the receipts folder: no second line is added.
    let step = fx.picks_step();
    assert!(
        !choice(&step, REMOVE)["writes"]
            .to_string()
            .contains(".gitignore"),
        "{step}"
    );
    // A project without that rule gets exactly one line.
    fx.write(".gitignore", "target/\n");
    let step = fx.picks_step();
    assert_eq!(
        choice(&step, REMOVE)["writes"],
        json!([".gitignore"]),
        "{step}"
    );
    let out = fx.pick(REMOVE);
    assert_eq!(out["writes"], json!([".gitignore"]), "{out}");
    let line = format!("{}/receipts/", moai_cli::state_root::dir(&fx.root));
    assert_eq!(fx.read(".gitignore"), format!("target/\n{line}\n"));
    let receipt = out["receipt"].as_str().unwrap().to_string();
    let reverted = fx.revert(&receipt);
    assert_eq!(reverted["result"]["state"], "reverted", "{reverted}");
    assert_eq!(fx.read(".gitignore"), "target/\n");
}

/// A refusal deletes no original and writes nothing.
#[test]
fn remove_refuses_a_changed_a_link_an_odd_or_a_large_original_and_deletes_nothing() {
    // An original changed after the preview: the digest of the preview is stale.
    let fx = Fx::moved();
    let digest = fx.remove_choice()["plan_digest"]
        .as_str()
        .unwrap()
        .to_string();
    fx.write("AGENTS.local.md", &format!("{LOCAL}one line more\n"));
    let before = fx.snapshot(true);
    let out = fx.merge_apply(REMOVE, &digest);
    assert_eq!(out["ok"], json!(false), "{out}");
    assert_eq!(fx.snapshot(true), before, "a refused pick wrote something");
    assert_eq!(fx.read("AGENTS.md"), AGENTS);

    // A link original, a file that is not UTF-8 and a file above 128 KiB: the pick is not offered,
    // a direct call is refused, and the other original stays too.
    type Break = fn(&Fx);
    let breaks: [(&str, Break); 3] = [
        ("a link", |fx| {
            fs::rename(fx.root.join("AGENTS.md"), fx.root.join("elsewhere.md")).unwrap();
            std::os::unix::fs::symlink("elsewhere.md", fx.root.join("AGENTS.md")).unwrap();
        }),
        ("not UTF-8", |fx| {
            fs::write(fx.root.join("AGENTS.local.md"), [0xff, 0xfe, b'\n']).unwrap();
        }),
        ("above 128 KiB", |fx| {
            fx.write("AGENTS.local.md", &"x".repeat(129 * 1024));
        }),
    ];
    for (what, make) in breaks {
        let fx = Fx::moved();
        let digest = fx.remove_choice()["plan_digest"]
            .as_str()
            .unwrap()
            .to_string();
        make(&fx);
        assert!(
            !fx.picks_state().1.contains(&REMOVE.to_string()),
            "{what}: still offered"
        );
        let before = fx.snapshot(true);
        let plan = fx.merge_plan(REMOVE);
        assert_ne!(plan["result"]["state"], "ready", "{what}: {plan}");
        for d in [
            digest.as_str(),
            plan["result"]["plan_digest"].as_str().unwrap(),
        ] {
            let out = fx.merge_apply(REMOVE, d);
            assert_eq!(out["ok"], json!(false), "{what}: {out}");
        }
        assert_eq!(
            fx.snapshot(true),
            before,
            "{what}: a refused pick wrote something"
        );
        assert!(
            fx.exists("AGENTS.md") && fx.exists("AGENTS.local.md"),
            "{what}"
        );
    }

    // The receipts folder is a link: the receipt cannot be made, so nothing is deleted.
    let fx = Fx::moved();
    let digest = fx.remove_choice()["plan_digest"]
        .as_str()
        .unwrap()
        .to_string();
    let receipts = fx.state_dir().join("receipts");
    let moved_away = fx.state_dir().join("receipts-away");
    fs::rename(&receipts, &moved_away).unwrap();
    std::os::unix::fs::symlink(&moved_away, &receipts).unwrap();
    let out = fx.merge_apply(REMOVE, &digest);
    assert_eq!(out["ok"], json!(false), "{out}");
    assert_originals(&fx, Some(AGENTS), Some(LOCAL));
}

// --- the receipt ---------------------------------------------------------------------------------

#[test]
fn receipt_revert_restores_first_bytes_after_move_and_removes_created_files() {
    let pick = "move";
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

/// The revert writes each deleted original back with its bytes and its unix mode, and a second
/// revert changes nothing.
#[test]
fn receipt_revert_writes_both_originals_back_with_their_bytes_and_modes() {
    let fx = Fx::moved();
    let first = fx.snapshot(false);
    let (agents_mode, local_mode) = (fx.mode("AGENTS.md"), fx.mode("AGENTS.local.md"));
    assert_eq!((agents_mode, local_mode), (0o755, 0o640));
    let out = fx.pick(REMOVE);
    let receipt = out["receipt"].as_str().unwrap().to_string();
    assert!(!fx.exists("AGENTS.md") && !fx.exists("AGENTS.local.md"));
    let reverted = fx.revert(&receipt);
    assert_eq!(reverted["ok"], json!(true), "{reverted}");
    assert_eq!(reverted["result"]["state"], "reverted", "{reverted}");
    assert_eq!(
        reverted["result"]["writes"],
        json!(["AGENTS.local.md", "AGENTS.md"]),
        "{reverted}"
    );
    assert_eq!(fx.snapshot(false), first, "every byte is back");
    assert_eq!(fx.mode("AGENTS.md"), agents_mode);
    assert_eq!(fx.mode("AGENTS.local.md"), local_mode);
    assert_originals(&fx, Some(AGENTS), Some(LOCAL));
    // The update offers the removal again.
    assert_eq!(choice_ids(&fx.picks_step()), [REMOVE]);
    // A second revert changes nothing.
    let again = fx.revert(&receipt);
    assert_eq!(again["ok"], json!(true), "{again}");
    assert_eq!(fx.snapshot(false), first);
    assert_eq!(fx.mode("AGENTS.md"), agents_mode);
}

/// A file that exists again at the path of a deleted original stays as it is and is named; the
/// other original still comes back.
#[test]
fn receipt_revert_leaves_a_recreated_original_untouched_and_names_it() {
    let fx = Fx::moved();
    let receipt = fx.pick(REMOVE)["receipt"].as_str().unwrap().to_string();
    fx.write("AGENTS.md", "A new file written since.\n");
    let out = fx.revert(&receipt);
    assert_eq!(out["ok"], json!(true), "{out}");
    assert_eq!(out["result"]["state"], "partial", "{out}");
    let kept = out["result"]["kept"].as_array().unwrap();
    assert_eq!(kept.len(), 1, "{out}");
    assert_eq!(kept[0]["path"], "AGENTS.md", "{out}");
    assert_eq!(fx.read("AGENTS.md"), "A new file written since.\n");
    assert_eq!(fx.read("AGENTS.local.md"), LOCAL);
    assert_eq!(fx.mode("AGENTS.local.md"), 0o640);
    // The copy of the deleted original stays in the receipt for the person to take.
    let copy = fx
        .state_dir()
        .join("receipts")
        .join(&receipt)
        .join("originals/AGENTS.md");
    assert_eq!(fs::read_to_string(copy).unwrap(), AGENTS);
}

/// A receipt written by the earlier build for the retired pick `import` still reverts: the revert
/// side of its `import-line` part stays, and a file that changed since is named and kept.
#[test]
fn receipt_revert_restores_an_import_line_part_built_by_the_earlier_version() {
    for edited in [false, true] {
        let fx = Fx::new();
        fx.write("AGENTS.md", AGENTS);
        fx.ready();
        let before = "# Mine\n";
        let body = format!("{IMPORT_LINE}\n");
        let sha = moai_cli::digest(body.as_bytes());
        let hex = sha.trim_start_matches("sha256:");
        let written = format!("{before}\n{IMPORT_BEGIN}{hex} -->\n{body}{IMPORT_END}\n");
        let current = if edited {
            format!("{written}and one line more\n")
        } else {
            written.clone()
        };
        fx.write("CLAUDE.md", &current);

        let store = fx.state_dir().join("receipts");
        let (engine, composite) = ("agents-import-1-0", "agents-pick-1-0");
        fs::create_dir_all(store.join(engine).join("before")).unwrap();
        fs::write(store.join(engine).join("before/CLAUDE.md"), before).unwrap();
        let engine_receipt = json!({
            "schema_version": 1, "id": engine, "at": "2026-11-22T00:00:00Z",
            "plan_digest": "sha256:old", "permission_reference": REF, "status": "applied",
            "agents_before": moai_cli::digest(before.as_bytes()),
            "agents_after": moai_cli::digest(written.as_bytes()),
            "block_sha256": hex, "source_kind": "in_place", "source_path": "AGENTS.md",
            "source_sha256": moai_cli::digest(AGENTS.as_bytes()),
            "target_path": "CLAUDE.md", "target_existed": true,
        });
        fs::write(
            store.join(engine).join("receipt.json"),
            serde_json::to_vec_pretty(&engine_receipt).unwrap(),
        )
        .unwrap();
        fs::create_dir_all(store.join(composite)).unwrap();
        let pick_receipt = json!({
            "schema_version": 1, "id": composite, "at": "2026-11-22T00:00:00Z",
            "pick": "import", "plan_digest": "sha256:old", "permission_reference": REF,
            "status": "applied",
            "parts": [{"part": "import-line", "receipts": [engine], "reverted": false}],
            "kept": [],
        });
        fs::write(
            store.join(composite).join("receipt.json"),
            serde_json::to_vec_pretty(&pick_receipt).unwrap(),
        )
        .unwrap();

        let out = fx.revert(composite);
        assert_eq!(out["ok"], json!(true), "edited={edited}: {out}");
        if edited {
            assert_eq!(out["result"]["state"], "partial", "{out}");
            assert!(
                out["result"]["kept"].to_string().contains("CLAUDE.md"),
                "{out}"
            );
            assert_eq!(fx.read("CLAUDE.md"), current, "the edited file was touched");
        } else {
            assert_eq!(out["result"]["state"], "reverted", "{out}");
            assert_eq!(fx.read("CLAUDE.md"), before);
        }
        assert_originals(&fx, Some(AGENTS), None);
    }
}

// --- what is refused -----------------------------------------------------------------------------

#[test]
fn refuse_claude_md_link_to_agents_md_writes_nothing_and_agents_md_stays() {
    // CLAUDE.md is a link to AGENTS.md: nothing may be written through it.
    {
        let pick = "move";
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
        // Nothing that writes is left: only `leave` stays, and the step reads a conflict.
        let picks = fx.picks_step();
        assert_eq!(choice_ids(&picks), ["leave"], "link={local_is_link}");
        assert_eq!(picks["state"], "conflict", "link={local_is_link}: {picks}");
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
