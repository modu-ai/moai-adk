//! #3004 P4-3: the generated Claude `statusLine` command is `moai statusline`; the one the
//! releases before it wrote is `moai-cli statusline`. These tests start from the captured pre-P4
//! files (`tests/fixtures/hooks-pre-p4/statusline.*`) copied into a temporary project and check,
//! through the binary, that
//!
//! - a project with no status line gets `moai statusline`, and a pristine old install is an
//!   `update` previewed as a rewrite, written only when the first `moai` on PATH is the Rust
//!   build and answers `moai statusline --help` as the Rust build, and kept byte for byte so it
//!   can be put back;
//! - an edited, foreign or receipt-less status line is left as it is and its command text is
//!   never printed;
//! - the gate holds the rewrite, with its reason, under a Go `moai`, no `moai`, a `moai` inside
//!   the project, a routing script that sends `statusline` somewhere else and a `moai` that
//!   does not answer the probe; the hold is at plan time and again at apply time (the plan
//!   digest binds the files, not the PATH) and never touches the hooks.
//!
//! The fakes are shell scripts in a temporary folder; the project and HOME are temporary and the
//! real PATH, HOME and installed `moai` are never read.
#![cfg(unix)]

mod common;

use common::CleanEnv as _;
use moai_cli::status_line;
use serde_json::{json, Value};
use std::{
    fs,
    io::Write as _,
    path::{Path, PathBuf},
    process::{Command, Stdio},
};
use tempfile::TempDir;

const FIXTURES: &str = concat!(env!("CARGO_MANIFEST_DIR"), "/tests/fixtures/hooks-pre-p4");
const RUST: &str = "moai 1.2.3 (rust)";
const GO: &str = "moai-adk v3.2.1-rc.6";
/// What the Go build prints for `statusline --help` (its first stdout line is blank).
const GO_HELP: &str =
    "\n  Generate a compact statusline string for display in the Claude Code status bar.\n";
const HELP: &str = "Render the Claude Code status line";
const NEW: &str = "{\"type\":\"command\",\"command\":\"moai statusline\"}";
const SETTINGS: &str = ".claude/settings.json";

/// What a fake `moai` answers: `--version`, and `statusline --help`.
#[derive(Clone, Copy)]
struct Fake {
    version: &'static str,
    /// Printed for `statusline ...`; `None` prints nothing and exits 0.
    statusline: Option<&'static str>,
    /// Sleeps for 30 s on `statusline ...` instead of answering.
    hangs: bool,
}

const RUST_FAKE: Fake = Fake {
    version: RUST,
    statusline: Some(HELP),
    hangs: false,
};
const GO_FAKE: Fake = Fake {
    version: GO,
    statusline: Some(GO_HELP),
    hangs: false,
};
/// The leader's shim: `--version` goes to Rust, `statusline` to the Go build.
const ROUTED_FAKE: Fake = Fake {
    version: RUST,
    statusline: Some(GO_HELP),
    hangs: false,
};
const MUTE_FAKE: Fake = Fake {
    version: RUST,
    statusline: None,
    hangs: false,
};
const HANGING_FAKE: Fake = Fake {
    version: RUST,
    statusline: None,
    hangs: true,
};

fn install_fake(dir: &Path, f: Fake) {
    fs::create_dir_all(dir).unwrap();
    let status = if f.hangs {
        "sleep 30".to_string()
    } else {
        f.statusline
            .map(|s| format!("printf '%s\\n' '{s}'"))
            .unwrap_or_else(|| "exit 0".into())
    };
    common::write_executable(
        &dir.join("moai"),
        &format!(
            "#!/bin/sh\ncase \"$1\" in\n  statusline) {status} ;;\n  *) printf '%s\\n' '{}' ;;\nesac\n",
            f.version
        ),
    );
}

/// One temporary world: `root/project`, `root/home` and the PATH folders.
struct World {
    _dir: TempDir,
    root: PathBuf,
    project: PathBuf,
    home: PathBuf,
}

/// A project with the earlier releases' state folder `.moai-cli` (where the captured receipts
/// live; the update preview would move it aside first, so only the settings tests use it).
fn world() -> World {
    let w = modern();
    fs::create_dir_all(w.project.join(".moai-cli")).unwrap();
    w
}

/// A project with no state folder yet: `project-apply` makes the current `.moai`.
fn modern() -> World {
    let dir = TempDir::new().unwrap();
    let root = fs::canonicalize(dir.path()).unwrap();
    let w = World {
        project: root.join("project"),
        home: root.join("home"),
        root,
        _dir: dir,
    };
    fs::create_dir_all(&w.project).unwrap();
    fs::create_dir_all(&w.home).unwrap();
    install_fake(&w.root.join("rust"), RUST_FAKE);
    install_fake(&w.root.join("go"), GO_FAKE);
    install_fake(&w.root.join("routed"), ROUTED_FAKE);
    install_fake(&w.root.join("mute"), MUTE_FAKE);
    install_fake(&w.root.join("hangs"), HANGING_FAKE);
    w
}

impl World {
    fn dir(&self, name: &str) -> String {
        format!("{}:/usr/bin:/bin", self.root.join(name).display())
    }
    fn rust(&self) -> String {
        self.dir("rust")
    }
    /// A PATH with no `moai` on it.
    fn none(&self) -> String {
        "/usr/bin:/bin".to_string()
    }
    /// A PATH whose `moai` is inside the project (never run).
    fn inside(&self) -> String {
        let dir = self.project.join("bin");
        install_fake(&dir, RUST_FAKE);
        format!("{}:/usr/bin:/bin", dir.display())
    }
    /// A PATH whose `moai` is a copy of the binary under test.
    fn real(&self) -> String {
        let dir = self.root.join("real");
        fs::create_dir_all(&dir).unwrap();
        common::copy_executable(Path::new(common::MOAI_BIN), &dir.join("moai"));
        format!("{}:/usr/bin:/bin", dir.display())
    }
    fn path(&self, rel: &str) -> PathBuf {
        self.project.join(rel)
    }
    fn read(&self, rel: &str) -> Vec<u8> {
        fs::read(self.path(rel)).unwrap()
    }
    fn write(&self, rel: &str, bytes: &[u8]) {
        let p = self.path(rel);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        fs::write(p, bytes).unwrap();
    }
    /// Puts the captured pre-P4 status line install (file and ownership receipt) into the project.
    fn install_legacy(&self) {
        self.write(SETTINGS, &fixture("statusline.settings.json"));
        self.write(&self.receipt(), &fixture("statusline.receipt.json"));
    }
    /// The ownership receipt of the settings file, in the state folder the project has.
    fn receipt(&self) -> String {
        let root = if self.path(".moai").is_dir() {
            ".moai"
        } else {
            ".moai-cli"
        };
        format!("{root}/manifests/claude-settings.json")
    }
    /// The settings file and its ownership receipt.
    fn files(&self) -> (Vec<u8>, Vec<u8>) {
        (self.read(SETTINGS), self.read(&self.receipt()))
    }
    fn settings_json(&self) -> Value {
        serde_json::from_slice(&self.read(SETTINGS)).unwrap()
    }
    /// A profile, so `project-update-plan` has an install phase (the `project_update_plan_cli`
    /// way: detect, `project-plan`, `project-apply`).
    fn with_profile(&self) {
        let p = self.project.to_str().unwrap();
        let detect = result(self, &["project-detect", p], b"", &self.none());
        let profile = json!({
            "schema_version": 2, "project_id": "demo", "purpose": "Build a small editor",
            "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude", "codex"],
            "detected": detect["detected"],
        });
        let plan = result(
            self,
            &["project-plan", p],
            &serde_json::to_vec(&profile).unwrap(),
            &self.none(),
        );
        let digest = plan["plan_digest"].as_str().unwrap().to_string();
        result(
            self,
            &[
                "project-apply",
                p,
                "--plan-digest",
                &digest,
                "--permission-reference",
                "test-3004",
            ],
            &serde_json::to_vec(&plan).unwrap(),
            &self.none(),
        );
    }
}

fn fixture(name: &str) -> Vec<u8> {
    fs::read(Path::new(FIXTURES).join(name)).unwrap()
}

/// Runs `moai ARGS` with `stdin` in a cleared environment; returns the whole envelope.
fn envelope_with(
    w: &World,
    args: &[&str],
    stdin: &[u8],
    path_env: &str,
    extra: &[(&str, &str)],
) -> Value {
    let mut child = Command::new(common::MOAI_BIN)
        .clean_env()
        .args(args)
        .current_dir(&w.root)
        .env("HOME", &w.home)
        .env("PATH", path_env)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .envs(extra.iter().copied())
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    let mut pipe = child.stdin.take().unwrap();
    let _ = pipe.write_all(stdin);
    drop(pipe);
    let out = child.wait_with_output().unwrap();
    serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|_| panic!("not JSON: {}", String::from_utf8_lossy(&out.stderr)))
}

fn envelope(w: &World, args: &[&str], stdin: &[u8], path_env: &str) -> Value {
    envelope_with(w, args, stdin, path_env, &[])
}

fn result(w: &World, args: &[&str], stdin: &[u8], path_env: &str) -> Value {
    let v = envelope(w, args, stdin, path_env);
    assert_eq!(v["ok"], json!(true), "{args:?}: {v}");
    v["result"].clone()
}

fn digest_of(plan: &Value) -> String {
    plan["plan_digest"].as_str().unwrap().to_string()
}

/// The request that asks for MoAI's own status line.
fn want() -> Vec<u8> {
    format!("{{\"statusLine\":{NEW}}}").into_bytes()
}

fn settings_plan(w: &World, stdin: &[u8], path_env: &str) -> Value {
    result(
        w,
        &["claude-settings-plan", w.project.to_str().unwrap()],
        stdin,
        path_env,
    )
}

fn settings_apply(w: &World, stdin: &[u8], digest: &str, path_env: &str) -> Value {
    envelope(
        w,
        &[
            "claude-settings-apply",
            w.project.to_str().unwrap(),
            "--plan-digest",
            digest,
            "--permission-reference",
            "test-3004",
        ],
        stdin,
        path_env,
    )
}

fn update_plan(w: &World, path_env: &str) -> Value {
    result(
        w,
        &["project-update-plan", w.project.to_str().unwrap()],
        b"",
        path_env,
    )
}

fn step(plan: &Value, name: &str) -> Value {
    plan["phases"]
        .as_array()
        .unwrap()
        .iter()
        .flat_map(|p| p["steps"].as_array().unwrap().iter())
        .find(|s| s["step"] == name)
        .unwrap_or_else(|| panic!("no step {name}: {plan}"))
        .clone()
}

/// AC-3: the generator's two spellings are told apart, the old one is the captured bytes, and
/// `moai-cli` is in nothing the current one writes.
#[test]
fn the_generators_new_default_calls_moai_and_the_old_one_is_the_fixture() {
    assert_eq!(
        status_line::current(),
        json!({"type": "command", "command": "moai statusline"})
    );
    assert_ne!(status_line::current(), status_line::legacy());
    let fixed: Value = serde_json::from_slice(&fixture("statusline.settings.json")).unwrap();
    assert_eq!(fixed["statusLine"], status_line::legacy());
    let receipt: Value = serde_json::from_slice(&fixture("statusline.receipt.json")).unwrap();
    assert_eq!(receipt["owned"]["statusLine"], status_line::legacy());
}

/// The probe reads the first line of `statusline --help`; this is the line the binary under
/// test prints, so a changed help text fails here and not on a user's machine.
#[test]
fn the_probe_line_is_what_the_rust_build_prints() {
    let out = Command::new(common::MOAI_BIN)
        .clean_env()
        .args(["statusline", "--help"])
        .env("HOME", "/nonexistent")
        .stdin(Stdio::null())
        .output()
        .unwrap();
    assert!(out.status.success());
    let text = String::from_utf8(out.stdout).unwrap();
    assert_eq!(text.lines().next(), Some(status_line::RUST_HELP_LINE));
    assert_eq!(status_line::RUST_HELP_LINE, HELP);
}

/// AC-3: a project with no status line is offered `moai statusline`, the printed command
/// writes exactly that, and neither the file nor the receipt names `moai-cli`.
#[test]
fn a_project_without_a_status_line_gets_moai_statusline() {
    let w = modern();
    w.with_profile();
    let plan = update_plan(&w, &w.rust());
    let s = step(&plan, "claude-settings");
    assert_eq!(s["state"], json!("ready"), "{s}");
    assert_eq!(
        s["apply"]["stdin"],
        json!({"statusLine": status_line::current()})
    );
    assert_eq!(
        s["preview"]["status_line_migration"]["state"],
        json!("create"),
        "{s}"
    );
    let stdin = serde_json::to_vec(&s["apply"]["stdin"]).unwrap();
    let p = settings_plan(&w, &stdin, &w.rust());
    let v = settings_apply(&w, &stdin, &digest_of(&p), &w.rust());
    assert_eq!(v["ok"], json!(true), "{v}");
    assert_eq!(
        w.settings_json(),
        json!({"statusLine": status_line::current()})
    );
    let (file, receipt) = w.files();
    for bytes in [&file, &receipt] {
        assert!(!String::from_utf8_lossy(bytes).contains("moai-cli"));
    }
    let next = update_plan(&w, &w.rust());
    let s = step(&next, "claude-settings");
    assert_eq!(s["state"], json!("unchanged"), "{s}");
    assert_eq!(
        s["preview"]["status_line_migration"]["state"],
        json!("none"),
        "{s}"
    );
}

/// A new default is not added where the `moai` first on PATH is a Go build or routes
/// `statusline` away; with no `moai` it is added (nothing would run), as a hook install is.
#[test]
fn the_new_default_waits_for_a_rust_statusline_when_a_moai_is_there() {
    for (name, held) in [
        ("go", "moai_wrong_generation"),
        ("routed", "statusline_routed_elsewhere"),
        ("mute", "statusline_unprobed"),
    ] {
        let w = modern();
        w.with_profile();
        let plan = update_plan(&w, &w.dir(name));
        let s = step(&plan, "claude-settings");
        assert_eq!(s["apply"], Value::Null, "{name}: {s}");
        assert_eq!(s["state"], json!("needs_decision"), "{name}: {s}");
        let m = &s["preview"]["status_line_migration"];
        assert_eq!(m["state"], json!("held"), "{name}: {s}");
        assert_eq!(m["held_action"], json!("create"), "{name}: {s}");
        assert_eq!(m["blocked_by"], json!(held), "{name}: {s}");
        // The direct writer refuses the same request, files untouched.
        let p = settings_plan(&w, &want(), &w.dir(name));
        let v = settings_apply(&w, &want(), &digest_of(&p), &w.dir(name));
        assert_eq!(v["ok"], json!(false), "{name}: {v}");
        assert!(!w.path(SETTINGS).exists(), "{name}");
    }
    let w = modern();
    w.with_profile();
    for path_env in [w.none(), w.inside()] {
        let s = step(&update_plan(&w, &path_env), "claude-settings");
        assert_eq!(s["state"], json!("ready"), "{s}");
        assert_eq!(
            s["preview"]["status_line_migration"]["state"],
            json!("create"),
            "{s}"
        );
    }
}

/// AC-2: a pristine old install is an `update` previewed as a rewrite (nothing is written and
/// the digest does not follow the PATH); the apply leaves the bytes and the receipt of a fresh
/// install, keeps the old file, and putting that copy back restores the install byte for byte.
#[test]
fn a_pristine_old_install_is_rewritten_behind_a_rust_moai_and_reverts() {
    let w = world();
    w.install_legacy();
    let before = w.files();
    let p = settings_plan(&w, &want(), &w.rust());
    assert_eq!(p["action"], json!("update"), "{p}");
    let m = &p["status_line_migration"];
    assert_eq!(m["state"], json!("rewrite"), "{p}");
    assert_eq!(m["from"], json!("moai-cli"));
    assert_eq!(m["to"], json!("moai"));
    assert_eq!(m["blocked_by"], Value::Null, "{p}");
    assert_eq!(w.files(), before, "a plan writes nothing");
    let under_go = settings_plan(&w, &want(), &w.dir("go"));
    assert_eq!(
        digest_of(&under_go),
        digest_of(&p),
        "the digest does not follow the PATH"
    );
    assert_eq!(
        under_go["status_line_migration"]["state"],
        json!("held"),
        "{under_go}"
    );

    let v = settings_apply(&w, &want(), &digest_of(&p), &w.rust());
    assert_eq!(v["ok"], json!(true), "{v}");
    let r = &v["result"];
    assert_eq!(
        r["status_line_migration"]["state"],
        json!("rewritten"),
        "{r}"
    );
    let backup = r["status_line_migration"]["backup_dir"]
        .as_str()
        .unwrap_or_else(|| panic!("{r}"));
    assert!(
        r["status_line_migration"]["revert"]
            .as_str()
            .unwrap()
            .contains(backup),
        "{r}"
    );
    assert_eq!(
        w.settings_json(),
        json!({"statusLine": status_line::current()})
    );
    let receipt: Value = serde_json::from_slice(&w.read(&w.receipt())).unwrap();
    assert_eq!(
        receipt["owned"]["statusLine"],
        status_line::current(),
        "ownership follows the rewrite"
    );
    assert!(!String::from_utf8_lossy(&w.read(&w.receipt())).contains("moai-cli"));
    assert_eq!(
        w.read(&format!("{backup}/settings.json")),
        fixture("statusline.settings.json"),
        "the old file is kept byte for byte"
    );

    // A fresh install of the same request ends in the same bytes.
    let f = world();
    let fp = settings_plan(&f, &want(), &f.rust());
    assert_eq!(fp["action"], json!("create"), "{fp}");
    let fv = settings_apply(&f, &want(), &digest_of(&fp), &f.rust());
    assert_eq!(fv["ok"], json!(true), "{fv}");
    assert_eq!(w.files(), f.files());
    let next = settings_plan(&w, &want(), &w.rust());
    assert_eq!(next["action"], json!("unchanged"), "{next}");

    // Put the kept file back: the file is the old one again and the next plan leaves it alone.
    w.write(SETTINGS, &w.read(&format!("{backup}/settings.json")));
    assert_eq!(w.read(SETTINGS), fixture("statusline.settings.json"));
    let after = settings_plan(&w, &want(), &w.rust());
    assert_eq!(after["action"], json!("unchanged"), "{after}");
}

/// AC-2 through the update preview: the old install is a ready settings step that carries the
/// new status line and names the rewrite; applying the printed command ends `unchanged`.
#[test]
fn the_update_preview_prints_the_rewrite_behind_a_rust_moai() {
    let w = modern();
    w.with_profile();
    w.install_legacy();
    let s = step(&update_plan(&w, &w.rust()), "claude-settings");
    assert_eq!(s["state"], json!("ready"), "{s}");
    assert_eq!(
        s["apply"]["stdin"],
        json!({"statusLine": status_line::current()}),
        "{s}"
    );
    assert_eq!(
        s["preview"]["status_line_migration"]["state"],
        json!("rewrite"),
        "{s}"
    );
    assert_eq!(
        s["preview"]["status_line_migration"]["blocked_by"],
        Value::Null,
        "{s}"
    );
    let stdin = serde_json::to_vec(&s["apply"]["stdin"]).unwrap();
    let digest = s["plan_digest"].as_str().unwrap();
    let v = settings_apply(&w, &stdin, digest, &w.rust());
    assert_eq!(v["ok"], json!(true), "{v}");
    let next = step(&update_plan(&w, &w.rust()), "claude-settings");
    assert_eq!(next["state"], json!("unchanged"), "{next}");
}

/// AC-10: the rewrite needs the Rust build first on PATH that answers `statusline` as the Rust
/// build. Each other PATH holds it with its reason; the writer refuses at apply time with the
/// files untouched; the update preview keeps the old status line and says why.
#[test]
fn the_gate_holds_the_rewrite_unless_the_rust_statusline_is_first() {
    let w = modern();
    w.with_profile();
    w.install_legacy();
    let before = w.files();
    let digest = digest_of(&settings_plan(&w, &want(), &w.rust()));
    let cases = [
        (w.dir("go"), "moai_wrong_generation"),
        (w.none(), "moai_binary_unresolved"),
        (w.inside(), "moai_not_probed"),
        (w.dir("routed"), "statusline_routed_elsewhere"),
        (w.dir("mute"), "statusline_unprobed"),
    ];
    for (path_env, blocked) in cases {
        let p = settings_plan(&w, &want(), &path_env);
        let m = &p["status_line_migration"];
        assert_eq!(m["state"], json!("held"), "{blocked}: {p}");
        assert_eq!(m["held_action"], json!("rewrite"), "{blocked}: {p}");
        assert_eq!(m["blocked_by"], json!(blocked), "{blocked}: {p}");
        assert_eq!(digest_of(&p), digest, "{blocked}");
        let v = settings_apply(&w, &want(), &digest, &path_env);
        assert_eq!(v["ok"], json!(false), "{blocked}: {v}");
        assert!(v.to_string().contains("not rewritten"), "{blocked}: {v}");
        assert_eq!(w.files(), before, "{blocked}");

        let s = step(&update_plan(&w, &path_env), "claude-settings");
        assert_eq!(s["apply"], Value::Null, "{blocked}: {s}");
        assert_eq!(s["state"], json!("needs_decision"), "{blocked}: {s}");
        assert_eq!(
            s["preview"]["status_line_migration"]["blocked_by"],
            json!(blocked),
            "{s}"
        );
    }
    assert_eq!(w.files(), before);
    // The Rust PATH lets the same plan through.
    let v = settings_apply(&w, &want(), &digest, &w.rust());
    assert_eq!(v["ok"], json!(true), "{v}");
}

/// A `moai` that never answers `statusline --help` is `statusline_unprobed` after the probe
/// limit, and a stray child it left running does not keep the call waiting.
#[test]
fn a_moai_that_hangs_on_the_probe_is_unprobed() {
    let w = world();
    w.install_legacy();
    let before = w.files();
    let started = std::time::Instant::now();
    let v = envelope_with(
        &w,
        &["claude-settings-plan", w.project.to_str().unwrap()],
        &want(),
        &w.dir("hangs"),
        &[("MOAI_CLI_TEST_PROBE_TIMEOUT_MS", "300")],
    );
    assert!(
        started.elapsed().as_secs() < 20,
        "the probe waited for the hung child"
    );
    assert_eq!(v["ok"], json!(true), "{v}");
    let m = &v["result"]["status_line_migration"];
    assert_eq!(m["blocked_by"], json!("statusline_unprobed"), "{v}");
    assert_eq!(w.files(), before);
}

/// The real binary is a Rust `moai` that answers the probe: a copy of it named `moai` lets the
/// rewrite through, so the probe line and the build cannot drift apart unseen.
#[test]
fn the_built_binary_passes_its_own_gate() {
    let w = world();
    w.install_legacy();
    let path_env = w.real();
    let p = settings_plan(&w, &want(), &path_env);
    assert_eq!(p["status_line_migration"]["state"], json!("rewrite"), "{p}");
    assert_eq!(p["status_line_migration"]["blocked_by"], Value::Null, "{p}");
    let v = settings_apply(&w, &want(), &digest_of(&p), &path_env);
    assert_eq!(v["ok"], json!(true), "{v}");
}

/// A status line hold never reaches the hooks: under a routing `moai` the old status line is
/// held and `hooks-apply` still installs the Claude hook set, leaving the status line bytes.
#[test]
fn a_status_line_hold_does_not_block_the_hooks() {
    let w = world();
    w.install_legacy();
    let routed = w.dir("routed");
    let held = settings_plan(&w, &want(), &routed);
    assert_eq!(
        held["status_line_migration"]["state"],
        json!("held"),
        "{held}"
    );
    let p = result(
        &w,
        &[
            "hooks-plan",
            w.project.to_str().unwrap(),
            "--host",
            "claude",
        ],
        b"",
        &routed,
    );
    assert_eq!(p["gate"]["allowed"], json!(true), "{p}");
    let v = envelope(
        &w,
        &[
            "hooks-apply",
            w.project.to_str().unwrap(),
            "--host",
            "claude",
            "--plan-digest",
            p["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "test-3004",
        ],
        b"",
        &routed,
    );
    assert_eq!(v["ok"], json!(true), "{v}");
    let after = w.settings_json();
    assert_eq!(after["statusLine"], status_line::legacy());
    assert!(after["hooks"].is_object(), "{after}");
    assert!(
        after["hooks"].to_string().contains("moai claim-verify"),
        "{after}"
    );
}

/// The old status line is recognised by equality with the old generator and nothing else: any
/// other text is the user's, stays as it is, and is never printed. The same holds for a status
/// line MoAI never owned.
#[test]
fn an_edited_or_foreign_status_line_is_kept_and_never_printed() {
    let edits = [
        "moai-cli statusline --edited-arg",
        "moai-cli  statusline",
        "/opt/edited/moai-cli statusline",
        "moai-cli statusline ",
        "moai-clistatusline",
        "my-status-line-31337",
    ];
    for command in edits {
        // Edited after MoAI wrote it: the receipt still holds the old value.
        let w = modern();
        w.with_profile();
        w.install_legacy();
        let edited = serde_json::to_vec_pretty(
            &json!({"statusLine": {"type": "command", "command": command}}),
        )
        .unwrap();
        w.write(SETTINGS, &edited);
        let before = w.files();
        for path_env in [w.rust(), w.dir("go")] {
            // The update preview and the writer's `status_line_migration` carry no command
            // text. (A conflicting request also gets the #2383 `confirm_replace`, which shows
            // the value it would replace by design; that is not checked here.)
            let plan = update_plan(&w, &path_env);
            let migration =
                settings_plan(&w, &want(), &path_env)["status_line_migration"].to_string();
            for shown in [plan.to_string(), migration] {
                for text in [
                    "--edited-arg",
                    "/opt/edited",
                    "31337",
                    "moai-cli  statusline",
                ] {
                    assert!(
                        !shown.contains(text),
                        "{command}: {text} was printed: {shown}"
                    );
                }
            }
            let s = step(&plan, "claude-settings");
            assert_eq!(s["state"], json!("unchanged"), "{command}: {s}");
            assert_eq!(s["apply"], Value::Null, "{command}: {s}");
            assert_eq!(
                s["preview"]["status_line_migration"]["state"],
                json!("none"),
                "{s}"
            );
            assert_eq!(
                s["preview"]["status_line_migration"]["reason"],
                json!("edited"),
                "{s}"
            );
        }
        // The writer: a request for MoAI's own line over an edited one is a conflict.
        let p = settings_plan(&w, &want(), &w.rust());
        assert_eq!(p["action"], json!("conflict"), "{command}: {p}");
        let v = settings_apply(
            &w,
            &want(),
            p["plan_digest"].as_str().unwrap_or("sha256:0"),
            &w.rust(),
        );
        assert_eq!(v["ok"], json!(false), "{command}: {v}");
        assert_eq!(w.files(), before, "{command}");

        // Never owned: the same text with no receipt.
        let w = modern();
        w.with_profile();
        w.write(SETTINGS, &edited);
        let before = w.read(SETTINGS);
        let s = step(&update_plan(&w, &w.rust()), "claude-settings");
        assert_eq!(s["apply"], Value::Null, "{command}: {s}");
        assert_eq!(
            s["preview"]["status_line_migration"]["state"],
            json!("none"),
            "{s}"
        );
        assert!(!s.to_string().contains(command.trim()), "{command}: {s}");
        let p = settings_plan(&w, &want(), &w.rust());
        assert_eq!(p["action"], json!("conflict"), "{command}: {p}");
        assert_eq!(w.read(SETTINGS), before, "{command}");
    }
}

/// An old status line the user removed stays removed, and the old value as the file's only
/// content without a receipt is not MoAI's to rewrite.
#[test]
fn a_removed_or_receipt_less_old_status_line_is_not_rewritten() {
    let w = modern();
    w.with_profile();
    w.install_legacy();
    w.write(SETTINGS, b"{}\n");
    let s = step(&update_plan(&w, &w.rust()), "claude-settings");
    assert_eq!(s["apply"], Value::Null, "{s}");
    assert_eq!(
        s["preview"]["status_line_migration"]["reason"],
        json!("removed"),
        "{s}"
    );
    assert_eq!(w.read(SETTINGS), b"{}\n");

    let w = world();
    w.write(SETTINGS, &fixture("statusline.settings.json"));
    let before = w.read(SETTINGS);
    let p = settings_plan(&w, &want(), &w.rust());
    assert_eq!(p["action"], json!("conflict"), "{p}");
    let v = settings_apply(
        &w,
        &want(),
        p["plan_digest"].as_str().unwrap_or("sha256:0"),
        &w.rust(),
    );
    assert_eq!(v["ok"], json!(false), "{v}");
    assert_eq!(w.read(SETTINGS), before);
}

/// The gate is for MoAI's own two spellings only: a status line request that is neither is the
/// caller's (`claude-settings-plan` has always carried it) and is not gated by the PATH.
#[test]
fn another_status_line_request_is_not_gated() {
    let w = world();
    let other = br#"{"statusLine":{"type":"command","command":"my-status"}}"#;
    let p = settings_plan(&w, other, &w.dir("go"));
    assert_eq!(p["action"], json!("create"), "{p}");
    assert!(p.get("status_line_migration").is_none(), "{p}");
    let v = settings_apply(&w, other, &digest_of(&p), &w.dir("go"));
    assert_eq!(v["ok"], json!(true), "{v}");
}
