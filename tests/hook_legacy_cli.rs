//! #3004 P4-2: the generated hook commands call `moai`; the hook sets of the releases before it
//! call `moai-cli`. These tests start from the captured pre-P4 files (`tests/fixtures/hooks-pre-p4/`)
//! copied into a temporary project and check, through the binary, that
//!
//! - a pristine old install is previewed as a rewrite (`binary_migration`), written only when the
//!   first `moai` on PATH is the Rust build, and reverted byte for byte;
//! - an edited, mixed or receipt-less install is left as it is, and the preview never repeats the
//!   edited command text;
//! - the gate refuses a rewrite under a Go `moai`, no `moai` or a `moai` inside the project, at
//!   plan time and again at apply time (the plan digest binds the files, not the PATH).
//!
//! The fakes are shell scripts in a temporary folder; the project and HOME are temporary and the
//! real PATH, HOME and installed `moai` are never read.
#![cfg(unix)]

mod common;

use common::CleanEnv as _;
use moai_cli::hooks;
use serde_json::{json, Value};
use std::{
    fs,
    path::{Path, PathBuf},
    process::{Command, Stdio},
};
use tempfile::TempDir;

const FIXTURES: &str = concat!(env!("CARGO_MANIFEST_DIR"), "/tests/fixtures/hooks-pre-p4");
const RUST: &str = "moai 1.2.3 (rust)";
const GO: &str = "moai-adk v3.2.1-rc.6";
const LANE_ARGS: [&str; 4] = ["--lane-loop", "agent-1", "--repo", "o/r"];

/// One host's files: where they go in the project and which fixtures hold the old bytes.
#[derive(Clone, Copy)]
struct Host {
    name: &'static str,
    file: &'static str,
    receipt: &'static str,
    fixture: &'static str,
    fixture_receipt: &'static str,
    /// `hooks-apply` flags of a fresh install of the same set.
    fresh_args: &'static [&'static str],
    /// The entries a rewrite replaces.
    entries: usize,
}

const CLAUDE: Host = Host {
    name: "claude",
    file: ".claude/settings.json",
    receipt: ".moai-cli/manifests/claude-settings.json",
    fixture: "claude.settings.json",
    fixture_receipt: "claude.receipt.json",
    fresh_args: &[],
    entries: 4,
};
const CODEX: Host = Host {
    name: "codex",
    file: ".codex/hooks.json",
    receipt: ".moai-cli/manifests/codex-hooks.json",
    fixture: "codex.hooks.json",
    fixture_receipt: "codex.receipt.json",
    fresh_args: &[],
    entries: 3,
};
const CODEX_LANE: Host = Host {
    name: "codex",
    file: ".codex/hooks.json",
    receipt: ".moai-cli/manifests/codex-hooks.json",
    fixture: "codex-lane.hooks.json",
    fixture_receipt: "codex-lane.receipt.json",
    fresh_args: &LANE_ARGS,
    entries: 4,
};
const OMP: Host = Host {
    name: "omp",
    file: ".omp/extensions/moai-hooks.ts",
    receipt: ".moai-cli/manifests/omp-hooks.json",
    fixture: "omp.moai-hooks.ts",
    fixture_receipt: "omp.receipt.json",
    fresh_args: &[],
    entries: 1,
};
const ALL: [Host; 4] = [CLAUDE, CODEX, CODEX_LANE, OMP];

/// One temporary world: `root/project`, `root/home` and the PATH folders `rust`, `go`.
struct World {
    _dir: TempDir,
    root: PathBuf,
    project: PathBuf,
    home: PathBuf,
}

fn script(dir: &Path, body: &str) {
    fs::create_dir_all(dir).unwrap();
    common::write_executable(
        &dir.join("moai"),
        &format!("#!/bin/sh\nprintf '%s\\n' '{body}'\n"),
    );
}

fn world() -> World {
    let dir = TempDir::new().unwrap();
    let root = fs::canonicalize(dir.path()).unwrap();
    let w = World {
        project: root.join("project"),
        home: root.join("home"),
        root,
        _dir: dir,
    };
    fs::create_dir_all(w.project.join(".moai-cli")).unwrap();
    fs::create_dir_all(&w.home).unwrap();
    script(&w.root.join("rust"), RUST);
    script(&w.root.join("go"), GO);
    w
}

impl World {
    fn rust(&self) -> String {
        format!("{}:/usr/bin:/bin", self.root.join("rust").display())
    }
    fn go(&self) -> String {
        format!("{}:/usr/bin:/bin", self.root.join("go").display())
    }
    /// A PATH with no `moai` on it.
    fn none(&self) -> String {
        "/usr/bin:/bin".to_string()
    }
    /// A PATH whose `moai` is inside the project (never run).
    fn inside(&self) -> String {
        let dir = self.project.join("bin");
        script(&dir, RUST);
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
    /// Puts the captured pre-P4 install of `h` into the project.
    fn install_legacy(&self, h: Host) {
        self.write(h.file, &fixture(h.fixture));
        self.write(h.receipt, &fixture(h.fixture_receipt));
    }
    /// The bytes of the host file and of its ownership receipt.
    fn files(&self, h: Host) -> (Vec<u8>, Vec<u8>) {
        (self.read(h.file), self.read(h.receipt))
    }
}

fn fixture(name: &str) -> Vec<u8> {
    fs::read(Path::new(FIXTURES).join(name)).unwrap()
}

/// Runs `moai ARGS` in a cleared environment; returns the whole envelope.
fn envelope(w: &World, args: &[&str], path_env: &str) -> Value {
    let out = Command::new(common::MOAI_BIN)
        .clean_env()
        .args(args)
        .current_dir(&w.root)
        .env("HOME", &w.home)
        .env("PATH", path_env)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .stdin(Stdio::null())
        .output()
        .unwrap();
    serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|_| panic!("not JSON: {}", String::from_utf8_lossy(&out.stderr)))
}

fn result(w: &World, args: &[&str], path_env: &str) -> Value {
    let v = envelope(w, args, path_env);
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn plan_args<'a>(w: &'a World, h: Host, verb: &'a str, extra: &[&'a str]) -> Vec<&'a str> {
    let mut a = vec![verb, w.project.to_str().unwrap(), "--host", h.name];
    a.extend_from_slice(extra);
    a
}

fn plan(w: &World, h: Host, path_env: &str) -> Value {
    result(w, &plan_args(w, h, "hooks-plan", &[]), path_env)
}

fn plan_envelope(w: &World, h: Host, path_env: &str) -> Value {
    envelope(w, &plan_args(w, h, "hooks-plan", &[]), path_env)
}

/// `hooks-apply` of `digest`; the whole envelope, since a refusal is part of the test.
fn apply(w: &World, h: Host, digest: &str, path_env: &str, extra: &[&str]) -> Value {
    let mut a = plan_args(w, h, "hooks-apply", extra);
    a.extend([
        "--plan-digest",
        digest,
        "--permission-reference",
        "test-3004",
    ]);
    envelope(w, &a, path_env)
}

fn digest_of(plan: &Value) -> String {
    plan["plan_digest"].as_str().unwrap().to_string()
}

/// A fresh install of `h` under the Rust PATH: what the generator writes now.
fn fresh(h: Host) -> (World, Value) {
    let w = world();
    let p = result(&w, &plan_args(&w, h, "hooks-plan", h.fresh_args), &w.rust());
    let v = apply(&w, h, &digest_of(&p), &w.rust(), h.fresh_args);
    assert_eq!(v["ok"], json!(true), "{v}");
    let next = plan(&w, h, &w.rust());
    (w, next)
}

fn commands(file: &[u8]) -> Vec<String> {
    let doc: Value = serde_json::from_slice(file).unwrap();
    let mut out = Vec::new();
    for groups in doc["hooks"].as_object().unwrap().values() {
        for g in groups.as_array().unwrap() {
            for h in g["hooks"].as_array().unwrap() {
                out.push(h["command"].as_str().unwrap().to_string());
            }
        }
    }
    out
}

/// AC-1: every command a fresh install writes starts with `moai ` and `moai-cli` is in no file
/// the install wrote, the receipts and the OMP module and the lane Stop group included.
#[test]
fn a_fresh_install_calls_moai_everywhere() {
    for h in ALL {
        let (w, next) = fresh(h);
        let (file, receipt) = w.files(h);
        for (name, bytes) in [("file", &file), ("receipt", &receipt)] {
            let text = String::from_utf8_lossy(bytes);
            assert!(!text.contains("moai-cli"), "{} {name}: {text}", h.fixture);
        }
        if h.name == "omp" {
            assert_eq!(file, hooks::omp_module("moai").as_bytes());
        } else {
            let cmds = commands(&file);
            assert_eq!(cmds.len(), h.entries, "{cmds:?}");
            assert!(cmds.iter().all(|c| c.starts_with("moai ")), "{cmds:?}");
        }
        assert_eq!(next["action"], json!("unchanged"), "{next}");
        assert_eq!(next["binary_migration"]["state"], json!("none"), "{next}");
    }
}

/// AC-2: a pristine old install is an `update` previewed as a rewrite; the preview writes
/// nothing and the plan digest does not depend on the PATH.
#[test]
fn a_legacy_install_is_previewed_as_a_rewrite_and_nothing_is_written() {
    for h in ALL {
        let w = world();
        w.install_legacy(h);
        let before = w.files(h);
        let p = plan(&w, h, &w.rust());
        assert_eq!(p["action"], json!("update"), "{p}");
        let m = &p["binary_migration"];
        assert_eq!(m["state"], json!("rewrite"), "{p}");
        assert_eq!(m["from"], json!("moai-cli"));
        assert_eq!(m["to"], json!("moai"));
        assert_eq!(m["blocked_by"], Value::Null, "{p}");
        let entries = m["entries"].as_array().unwrap();
        assert_eq!(entries.len(), h.entries, "{p}");
        if h.name == "omp" {
            let old = hooks::omp_module(hooks::LEGACY_BINARY);
            let new = hooks::omp_module(hooks::HOOK_BINARY);
            assert_eq!(entries[0]["kind"], json!("module"));
            assert_eq!(
                entries[0]["before_sha256"],
                json!(moai_cli::digest(old.as_bytes()))
            );
            assert_eq!(
                entries[0]["after_sha256"],
                json!(moai_cli::digest(new.as_bytes()))
            );
        } else {
            for e in entries {
                let (b, a) = (e["before"].as_str().unwrap(), e["after"].as_str().unwrap());
                assert!(b.starts_with("moai-cli "), "{e}");
                assert_eq!(a, b.replacen("moai-cli", "moai", 1), "{e}");
                assert_eq!(e["decision"], json!("rewrite"), "{e}");
            }
        }
        assert_eq!(p["gate"]["allowed"], json!(true), "{p}");
        assert_eq!(w.files(h), before, "a plan writes nothing");
        // The PATH changes the verdict, not the digest.
        let under_go = plan(&w, h, &w.go());
        assert_eq!(digest_of(&under_go), digest_of(&p), "{}", h.fixture);
        assert_eq!(under_go["gate"]["allowed"], json!(false), "{under_go}");
    }
}

/// AC-3 / AC-4: the rewrite leaves no `moai-cli`, ends with the bytes (and the plan digest) of a
/// fresh install, and `hooks-revert` of the printed receipt restores both files byte for byte.
#[test]
fn the_rewrite_equals_a_fresh_install_and_reverts_byte_for_byte() {
    for h in ALL {
        let w = world();
        w.install_legacy(h);
        let before = w.files(h);
        let p = plan(&w, h, &w.rust());
        let v = apply(&w, h, &digest_of(&p), &w.rust(), &[]);
        assert_eq!(v["ok"], json!(true), "{v}");
        let r = &v["result"];
        let id = r["receipt"].as_str().unwrap_or_else(|| panic!("{v}"));
        assert!(r["revert"].as_str().unwrap().contains(id), "{v}");
        let (file, receipt) = w.files(h);
        assert!(!String::from_utf8_lossy(&file).contains("moai-cli"));
        assert!(!String::from_utf8_lossy(&receipt).contains("moai-cli"));

        let (fresh_world, fresh_next) = fresh(h);
        assert_eq!((file, receipt), fresh_world.files(h), "{}", h.fixture);
        let next = plan(&w, h, &w.rust());
        assert_eq!(next["action"], json!("unchanged"), "{next}");
        assert_eq!(next["binary_migration"]["state"], json!("none"), "{next}");
        assert_eq!(digest_of(&next), digest_of(&fresh_next), "{}", h.fixture);

        let reverted = result(
            &w,
            &[
                "hooks-revert",
                w.project.to_str().unwrap(),
                "--receipt",
                id,
                "--permission-reference",
                "test-3004",
            ],
            &w.rust(),
        );
        assert!(reverted.is_object(), "{reverted}");
        assert_eq!(w.files(h), before, "{}", h.fixture);
        assert_eq!(w.files(h).0, fixture(h.fixture));
        assert_eq!(w.files(h).1, fixture(h.fixture_receipt));
    }
}

/// Replaces `from` by `to` in the old Claude settings, leaving the ownership receipt as it was.
fn edited_claude(from: &str, to: &str) -> World {
    let w = world();
    w.install_legacy(CLAUDE);
    let text = String::from_utf8(fixture(CLAUDE.fixture)).unwrap();
    assert_eq!(text.matches(from).count(), 1, "{from}");
    w.write(CLAUDE.file, text.replace(from, to).as_bytes());
    w
}

/// AC-5: an entry the user changed is kept as it is: the plan is a conflict, nothing is written,
/// and the changed text is never repeated.
#[test]
fn an_edited_entry_keeps_the_whole_set_and_is_never_printed() {
    let edits = [
        ("\"timeout\": 10\n          }\n        ]\n      }\n    ],\n    \"PreToolUse\"", "\"timeout\": 31337\n          }\n        ]\n      }\n    ],\n    \"PreToolUse\""),
        ("moai-cli claim-verify . --hook", "moai-cli claim-verify . --hook --edited-arg"),
        ("moai-cli claim-verify . --hook", "/opt/edited/moai-cli claim-verify . --hook"),
        ("\"matcher\": \"Bash\",\n        \"hooks\": [\n          {\n            \"type\": \"command\",\n            \"command\": \"moai-cli claim-verify", "\"matcher\": \"Edited\",\n        \"hooks\": [\n          {\n            \"type\": \"command\",\n            \"command\": \"moai-cli claim-verify"),
        ("moai-cli claim-verify . --hook", "moai-cli  claim-verify . --hook"),
    ];
    for (from, to) in edits {
        let w = edited_claude(from, to);
        let before = w.files(CLAUDE);
        let p = plan(&w, CLAUDE, &w.rust());
        assert_eq!(p["action"], json!("conflict"), "{to}: {p}");
        let m = &p["binary_migration"];
        assert_eq!(m["state"], json!("held"), "{to}: {p}");
        assert_eq!(m["blocked_by"], json!("mixed_entries"), "{to}: {p}");
        assert!(
            m["kept"]
                .as_array()
                .unwrap()
                .iter()
                .any(|k| k["reason"] == json!("edited")),
            "{to}: {p}"
        );
        let shown = p.to_string();
        for text in [
            "--edited-arg",
            "/opt/edited",
            "31337",
            "moai-cli  claim-verify",
        ] {
            assert!(!shown.contains(text), "{text} was printed: {shown}");
        }
        let v = apply(
            &w,
            CLAUDE,
            p["plan_digest"].as_str().unwrap_or("sha256:0"),
            &w.rust(),
            &[],
        );
        assert_eq!(v["ok"], json!(false), "{to}: {v}");
        assert_eq!(w.files(CLAUDE), before, "{to}");
    }
}

/// AC-6: a Claude set with another hook beside MoAI's old entries is `mixed_entries`, an old set
/// with no ownership receipt is `not_owned`; both are held whole. Codex and OMP files that are
/// edited or have no receipt are left for review.
#[test]
fn mixed_and_receipt_less_installs_are_held() {
    // Mixed: the user added a group of their own.
    let w = world();
    w.install_legacy(CLAUDE);
    let mut doc: Value = serde_json::from_slice(&fixture(CLAUDE.fixture)).unwrap();
    doc["hooks"]["PreToolUse"].as_array_mut().unwrap().push(
        json!({"matcher": "Bash", "hooks": [{"type": "command", "command": "/usr/local/bin/mine"}]}),
    );
    w.write(CLAUDE.file, &serde_json::to_vec_pretty(&doc).unwrap());
    let before = w.files(CLAUDE);
    let p = plan(&w, CLAUDE, &w.rust());
    assert_eq!(p["action"], json!("conflict"), "{p}");
    assert_eq!(p["binary_migration"]["state"], json!("held"), "{p}");
    assert_eq!(p["binary_migration"]["blocked_by"], json!("mixed_entries"));
    assert!(!p.to_string().contains("/usr/local/bin/mine"), "{p}");
    let v = apply(
        &w,
        CLAUDE,
        p["plan_digest"].as_str().unwrap_or("sha256:0"),
        &w.rust(),
        &[],
    );
    assert_eq!(v["ok"], json!(false), "{v}");
    assert_eq!(w.files(CLAUDE), before);

    // Q8: the old set without its ownership receipt is not MoAI's to rewrite.
    let w = world();
    w.write(CLAUDE.file, &fixture(CLAUDE.fixture));
    let before = w.read(CLAUDE.file);
    let p = plan(&w, CLAUDE, &w.rust());
    assert_eq!(p["action"], json!("conflict"), "{p}");
    assert_eq!(p["binary_migration"]["state"], json!("held"), "{p}");
    assert_eq!(
        p["binary_migration"]["blocked_by"],
        json!("not_owned"),
        "{p}"
    );
    assert_eq!(w.read(CLAUDE.file), before);

    // Codex and OMP: a whole-file receipt; an edited file or no receipt is a review.
    for h in [CODEX, CODEX_LANE, OMP] {
        let w = world();
        w.write(h.file, &fixture(h.fixture));
        let before = w.read(h.file);
        let v = plan_envelope(&w, h, &w.rust());
        assert_eq!(v["ok"], json!(false), "{}: {v}", h.fixture);
        assert!(
            v.to_string().contains("requires preservation review"),
            "{v}"
        );
        assert_eq!(w.read(h.file), before);
        w.install_legacy(h);
        let mut edited = fixture(h.fixture);
        edited.extend_from_slice(b"\n// edited by the user\n");
        w.write(h.file, &edited);
        let v = plan_envelope(&w, h, &w.rust());
        assert_eq!(v["ok"], json!(false), "{}: {v}", h.fixture);
        assert!(
            v.to_string().contains("requires preservation review"),
            "{v}"
        );
        assert_eq!(w.read(h.file), edited);
    }
}

/// AC-10: a rewrite needs the Rust build first on PATH. Under a Go `moai`, no `moai` or a `moai`
/// inside the project the plan says why and `hooks-apply` refuses, files untouched; a plan made
/// under the Rust PATH is refused again at apply time under another.
#[test]
fn the_gate_refuses_a_rewrite_unless_the_rust_moai_is_first() {
    for h in ALL {
        let w = world();
        w.install_legacy(h);
        let before = w.files(h);
        let rust_plan = plan(&w, h, &w.rust());
        let digest = digest_of(&rust_plan);
        let cases = [
            (w.go(), "moai_wrong_generation"),
            (w.none(), "moai_binary_unresolved"),
            (w.inside(), "moai_not_probed"),
        ];
        for (path_env, blocked) in cases {
            let p = plan(&w, h, &path_env);
            assert_eq!(p["gate"]["allowed"], json!(false), "{blocked}: {p}");
            assert_eq!(p["binary_migration"]["state"], json!("rewrite"), "{p}");
            assert_eq!(p["binary_migration"]["blocked_by"], json!(blocked), "{p}");
            assert_eq!(digest_of(&p), digest, "the digest does not follow the PATH");
            let v = apply(&w, h, &digest, &path_env, &[]);
            assert_eq!(v["ok"], json!(false), "{blocked}: {v}");
            assert!(v.to_string().contains("not rewritten"), "{blocked}: {v}");
            assert_eq!(w.files(h), before, "{blocked} {}", h.fixture);
        }
        let v = apply(&w, h, &digest, &w.rust(), &[]);
        assert_eq!(v["ok"], json!(true), "{v}");
    }
}

/// A plan that writes nothing is never gated: an install already on `moai` is `unchanged`
/// under any PATH.
#[test]
fn an_unchanged_install_is_not_gated() {
    let (w, _) = fresh(CLAUDE);
    for path_env in [w.go(), w.none()] {
        let p = plan(&w, CLAUDE, &path_env);
        assert_eq!(p["action"], json!("unchanged"), "{p}");
        assert_eq!(p["gate"]["allowed"], json!(true), "{p}");
    }
}

fn update_world(h: Host) -> World {
    let w = world();
    common::git::git_ok(&w.project, &["init", "-q"]);
    w.install_legacy(h);
    w
}

/// The `moai update` hook layer: the preview counts the entries it would rewrite, a Go `moai`
/// turns the host into a conflict with the reason, and the PATH is read again at apply time.
#[test]
fn the_update_layer_counts_the_rewrite_and_gates_it() {
    use moai_cli::update_hooks;
    let w = update_world(CLAUDE);
    let before = w.files(CLAUDE);
    let layer = update_hooks::plan_in(&w.home, &w.project, &w.rust()).unwrap();
    let host = &layer.to_json()["hosts"][0];
    assert_eq!(host["action"], json!("update"), "{host}");
    assert_eq!(host["migration"], json!(CLAUDE.entries), "{host}");
    assert!(layer.writes());
    assert!(layer
        .summary()
        .contains("moai-cli hook entries become moai"));

    let go = update_hooks::plan_in(&w.home, &w.project, &w.go()).unwrap();
    let host = &go.to_json()["hosts"][0];
    assert_eq!(host["action"], json!("conflict"), "{host}");
    assert_eq!(host["plan_digest"], Value::Null, "{host}");
    assert!(
        host["reason"]
            .as_str()
            .unwrap()
            .contains("wrong_generation"),
        "{host}"
    );
    assert!(!go.writes());

    // Planned under the Rust PATH, applied under the Go one: refused, nothing written.
    let out =
        update_hooks::apply_in(&w.home, &layer, &layer.plan_digest, "test-3004", &w.go()).unwrap();
    assert_eq!(out["state"], json!("error"), "{out}");
    assert!(out.to_string().contains("not rewritten"), "{out}");
    assert_eq!(w.files(CLAUDE), before);

    let out = update_hooks::apply_in(&w.home, &layer, &layer.plan_digest, "test-3004", &w.rust())
        .unwrap();
    assert_eq!(out["state"], json!("applied"), "{out}");
    let (file, _) = w.files(CLAUDE);
    assert!(!String::from_utf8_lossy(&file).contains("moai-cli"));
}

/// `hook_status.configured` tells an old `moai-cli` install (it still runs) from a current one
/// and from a drifted file, for every host.
#[test]
fn the_status_names_an_old_install_legacy_binary() {
    for h in ALL {
        let w = world();
        w.install_legacy(h);
        let p = plan(&w, h, &w.rust());
        assert_eq!(
            p["hook_status"]["configured"],
            json!("legacy_binary"),
            "{}: {p}",
            h.fixture
        );
        let v = apply(&w, h, &digest_of(&p), &w.rust(), &[]);
        assert_eq!(v["ok"], json!(true), "{v}");
        let p = plan(&w, h, &w.rust());
        assert_eq!(p["hook_status"]["configured"], json!("yes"), "{p}");
    }
    let w = world();
    w.install_legacy(OMP);
    w.write(OMP.file, b"// something else\n");
    let p = plan_envelope(&w, OMP, &w.rust());
    assert!(p.to_string().contains("preservation review"), "{p}");
}

/// A file the receipt owns whole but that is neither the old nor the current set (an earlier
/// MoAI set) is replaced as a whole and named as such, never shown command by command.
#[test]
fn an_earlier_owned_set_is_replaced_whole_and_named() {
    // Claude: the `hooks` key equals the receipt's `owned.hooks`.
    let w = world();
    let earlier = json!({"SessionStart": [{"hooks": [{"type": "command",
        "command": "moai-cli hook-receipt --host claude --earlier", "timeout": 10}]}]});
    let mut settings = serde_json::to_vec_pretty(&json!({ "hooks": earlier })).unwrap();
    settings.push(b'\n');
    w.write(CLAUDE.file, &settings);
    let mut receipt: Value = serde_json::from_slice(&fixture(CLAUDE.fixture_receipt)).unwrap();
    receipt["owned"] = json!({ "hooks": earlier });
    receipt["settings_sha256"] = json!(moai_cli::digest(&settings));
    w.write(
        CLAUDE.receipt,
        &serde_json::to_vec_pretty(&receipt).unwrap(),
    );
    let p = plan(&w, CLAUDE, &w.rust());
    assert_eq!(p["action"], json!("update"), "{p}");
    let m = &p["binary_migration"];
    assert_eq!(m["state"], json!("rewrite"), "{p}");
    assert_eq!(m["entries"], json!([]), "{p}");
    assert_eq!(m["replaced"][0]["reason"], json!("earlier_moai_set"), "{p}");
    assert!(!p.to_string().contains("--earlier"), "{p}");

    // OMP: the receipt's sha256 matches a module that is not the old one.
    let w = world();
    let module = b"// an earlier moai module\n";
    w.write(OMP.file, module);
    let mut receipt: Value = serde_json::from_slice(&fixture(OMP.fixture_receipt)).unwrap();
    receipt["sha256"] = json!(moai_cli::digest(module));
    w.write(OMP.receipt, &serde_json::to_vec_pretty(&receipt).unwrap());
    let p = plan(&w, OMP, &w.rust());
    assert_eq!(p["action"], json!("update"), "{p}");
    let m = &p["binary_migration"];
    assert_eq!(m["state"], json!("rewrite"), "{p}");
    assert_eq!(m["entries"], json!([]), "{p}");
    assert_eq!(
        m["replaced"][0]["reason"],
        json!("earlier_moai_module"),
        "{p}"
    );
}
