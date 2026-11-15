//! #2936 S1c-1: what the backup preview says about the Go generation's `.codex/` files (PLAN
//! 6878897673). Classify and preview only: every `.codex` item stays `leave`, the apply moves
//! nothing under `.codex`, and a preview writes nothing. Acceptance criteria, numbered as the
//! tests are named:
//!
//! - AC-1 `.codex/agents/moai/*.toml` that match the Go deploy are tool assets
//! - AC-2 an edited, `user_modified` or `user_created` agent file keeps its own class
//! - AC-3 a `generated_managed` bare-hex hash matches; only that provenance is normalized
//! - AC-4 `hooks.json` needs content proof beside the hash; only key names are printed
//! - AC-5 `config.toml` needs content proof; the exact Go table value; only key names printed
//! - AC-6 a hash that no longer matches the record is not proof
//! - AC-7 a file in the Rust generation manifest is never classified as Go
//! - AC-8 a `hooks.json` the Rust receipt owns is never classified as Go
//! - AC-9 a `config.toml` with keys the Rust receipt owns is never classified as Go
//! - AC-10 interrupted Go wiring: everything under `.codex` stays, with a warning
//! - AC-11 a symlink on the path keeps the file in place, unread
//! - AC-12 secret names under `.codex` are counted and never opened
//! - AC-13 `.codex` is counted among the unrecorded files, at its edge
//! - AC-14 the 262144-byte read limit of `hooks.json` and `config.toml`, at its edge
//! - AC-15 the apply moves nothing under `.codex`; `.codex/hooks/moai/**` stays (Q4)
//!
//! (AC-20, the backup step's own text and class counts, is in `project_update_plan_cli.rs`.)
//! Every case builds a temporary project; the real checkout and HOME are never read.

#![cfg(unix)]

mod common;
use common::CleanEnv;
use serde_json::{json, Value};
use sha2::{Digest, Sha256};
use std::{
    ffi::CString,
    fs,
    os::unix::fs::symlink,
    path::PathBuf,
    process::{Command, Stdio},
    time::{Duration, Instant},
};
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;
const REF: &str = "test-permission-ref";
/// `hosts::codex_config::CONFIG_LIMIT`.
const CONFIG_LIMIT: usize = 262_144;
/// A command that opened a FIFO would wait for a writer forever; this ends the test instead.
const DEADLINE: Duration = Duration::from_secs(120);
const LATER: &str = "classified; moves in S1c-2";

fn sha(bytes: &[u8]) -> String {
    format!("sha256:{:x}", Sha256::digest(bytes))
}

fn hex(bytes: &[u8]) -> String {
    format!("{:x}", Sha256::digest(bytes))
}

/// What the Go wiring writes into `.codex/hooks.json` (`internal/codexwiring/hooks.go`).
const GO_HOOKS: &str = r#"{
  "description": "MoAI-managed hook layer.",
  "hooks": {
    "PreToolUse": [
      {"matcher": "Bash", "hooks": [{"type": "command", "command": "moai hook pre-tool --harness codex", "timeout": 10}]}
    ],
    "SessionStart": [
      {"hooks": [{"type": "command", "command": "moai hook session-start --harness codex", "timeout": 10},
                 {"type": "command", "command": "sh /work/.codex/hooks/moai/start.sh", "timeout": 10}]}
    ]
  }
}
"#;

/// What the Go wiring writes into `.codex/config.toml` (`internal/codexwiring/configtoml.go`).
const GO_CONFIG: &str = r#"[mcp_servers.moai]
command = "moai"
args = ["mcp-server"]
env_vars = ["MOAI_HOME", "MOAI_KANBAN_ID", "MOAI_SESSION_PID", "MOAI_KANBAN_BACKEND", "MOAI_FACTORY_ROLE", "MOAI_FACTORY_WORKER", "MOAI_FACTORY_WORKERS", "CLAUDE_PROJECT_DIR", "CLAUDE_CODE_SESSION_ID"]
default_tools_approval_mode = "writes"

[tui]
status_line = ["model-with-reasoning", "codex-version", "context-remaining"]
"#;

/// One tree entry, never followed and never opened unless it is a regular file.
#[derive(Debug, PartialEq, Eq, PartialOrd, Ord, Clone)]
enum Node {
    File(Vec<u8>),
    Link(PathBuf),
    Fifo,
}

struct Fx {
    _tmp: TempDir,
    home: PathBuf,
    root: PathBuf,
    files: serde_json::Map<String, Value>,
}

impl Fx {
    fn new() -> Fx {
        let tmp = TempDir::new().unwrap();
        let base = fs::canonicalize(tmp.path()).unwrap();
        let (home, root) = (base.join("home"), base.join("project"));
        fs::create_dir_all(&home).unwrap();
        fs::create_dir_all(&root).unwrap();
        Fx {
            _tmp: tmp,
            home,
            root,
            files: serde_json::Map::new(),
        }
    }

    fn p(&self) -> &str {
        self.root.to_str().unwrap()
    }

    /// A folder next to the project, for a link target outside it.
    fn outside(&self, rel: &str, body: &str) -> PathBuf {
        let p = self.root.parent().unwrap().join("outside").join(rel);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        fs::write(&p, body).unwrap();
        p
    }

    fn write(&self, rel: &str, body: &str) {
        let p = self.root.join(rel);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        fs::write(p, body).unwrap();
    }

    fn link(&self, rel: &str, body: &str) {
        let p = self.root.join(rel);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        symlink(body, p).unwrap();
    }

    fn fifo(&self, rel: &str) {
        let p = self.root.join(rel);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        let c = CString::new(p.to_str().unwrap()).unwrap();
        // SAFETY: mkfifo with a valid NUL-terminated path.
        assert_eq!(unsafe { libc::mkfifo(c.as_ptr(), 0o600) }, 0, "{rel}");
    }

    /// A file the Go deployer wrote and recorded (`template_managed`, `sha256:` hash).
    fn deployed(&mut self, rel: &str, body: &str) {
        self.write(rel, body);
        self.record(rel, "template_managed", &sha(body.as_bytes()));
    }

    fn record(&mut self, rel: &str, provenance: &str, deployed: &str) {
        self.files.insert(
            rel.into(),
            json!({"provenance": provenance, "template_hash": deployed,
                   "deployed_hash": deployed, "current_hash": deployed}),
        );
    }

    /// A file the Go wiring wrote: `generated_managed`, a bare-hex hash, and its `parts`.
    fn wired(&mut self, rel: &str, body: &str) {
        self.write(rel, body);
        self.record_wired(rel, &hex(body.as_bytes()));
    }

    fn record_wired(&mut self, rel: &str, hash: &str) {
        self.files.insert(
            rel.into(),
            json!({"provenance": "generated_managed", "template_hash": "",
                   "deployed_hash": hash, "current_hash": hash,
                   "parts": [{"kind": "whole-file", "origin": "created"},
                             {"kind": "json-key", "key": "description", "origin": "created",
                              "hash": hash}]}),
        );
    }

    fn manifest(&self) {
        self.write(
            ".moai/manifest.json",
            &json!({"version": "v3.2.1-rc.6", "deployed_at": "2026-11-01T00:00:00Z",
                    "files": self.files})
            .to_string(),
        );
    }

    /// Every entry under the project (`.git` included) and HOME; links read as links, FIFOs
    /// named, never opened.
    fn tree(&self) -> Vec<(PathBuf, Node)> {
        let mut out = Vec::new();
        let mut stack = vec![self.root.clone(), self.home.clone()];
        while let Some(d) = stack.pop() {
            for e in fs::read_dir(&d).unwrap() {
                let p = e.unwrap().path();
                let m = fs::symlink_metadata(&p).unwrap();
                if m.file_type().is_symlink() {
                    out.push((p.clone(), Node::Link(fs::read_link(&p).unwrap())));
                } else if m.is_dir() {
                    stack.push(p);
                } else if m.is_file() {
                    out.push((p.clone(), Node::File(fs::read(&p).unwrap())));
                } else {
                    out.push((p.clone(), Node::Fifo));
                }
            }
        }
        out.sort();
        out
    }

    fn run_env(&self, args: &[&str], env: &[(&str, &str)]) -> Value {
        let mut cmd = Command::new(EXE);
        cmd.args(args)
            .clean_env()
            .env("HOME", &self.home)
            .env("PATH", "/usr/bin:/bin")
            .env("LC_ALL", "C")
            .stdin(Stdio::null())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped());
        for (k, v) in env {
            cmd.env(k, v);
        }
        let mut child = cmd.spawn().unwrap();
        let start = Instant::now();
        while child.try_wait().unwrap().is_none() {
            if start.elapsed() > DEADLINE {
                let _ = child.kill();
                let _ = child.wait();
                panic!("{args:?} did not finish: a FIFO was opened, or the walk did not stop");
            }
            std::thread::sleep(Duration::from_millis(20));
        }
        let out = child.wait_with_output().unwrap();
        serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
            panic!(
                "{args:?}: {}{}",
                String::from_utf8_lossy(&out.stdout),
                String::from_utf8_lossy(&out.stderr)
            )
        })
    }

    /// `go-state-plan`, with the whole tree compared before and after.
    fn plan_env(&self, env: &[(&str, &str)]) -> Value {
        let before = self.tree();
        let v = self.run_env(&["go-state-plan", self.p()], env);
        assert_eq!(v["ok"], json!(true), "{v}");
        assert_eq!(self.tree(), before, "a preview writes nothing");
        v["result"].clone()
    }

    fn plan(&self) -> Value {
        self.plan_env(&[])
    }

    fn apply(&self, digest: &str) -> Value {
        self.run_env(
            &[
                "go-state-apply",
                self.p(),
                "--plan-digest",
                digest,
                "--permission-reference",
                REF,
            ],
            &[],
        )
    }
}

fn item<'a>(plan: &'a Value, rel: &str) -> &'a Value {
    plan["items"]
        .as_array()
        .unwrap()
        .iter()
        .find(|i| i["path"] == rel)
        .unwrap_or_else(|| panic!("{rel} not in plan: {plan}"))
}

fn has_item(plan: &Value, rel: &str) -> bool {
    plan["items"]
        .as_array()
        .unwrap()
        .iter()
        .any(|i| i["path"] == rel)
}

fn class<'a>(plan: &'a Value, rel: &str) -> &'a str {
    item(plan, rel)["class"].as_str().unwrap()
}

fn proof<'a>(plan: &'a Value, rel: &str) -> &'a str {
    item(plan, rel)["proof"].as_str().unwrap()
}

/// The class and the action together; no `.codex` item ever moves in this slice.
fn stays(plan: &Value, rel: &str, want: &str) {
    let i = item(plan, rel);
    assert_eq!(i["class"], json!(want), "{rel}: {plan}");
    assert_eq!(i["action"], json!("leave"), "{rel}: {plan}");
}

/// A Go project whose `.moai/` moves, so the whole preview runs.
fn go_fx() -> Fx {
    let mut fx = Fx::new();
    fx.deployed(".moai/config/sections/a.yaml", "a: 1\n");
    fx.deployed(".claude/skills/moai-a/SKILL.md", "# a\n");
    fx
}

/// Rust state files sit under the earlier state folder while the Go `.moai/` is in place.
fn rust_file(fx: &Fx, rel: &str, body: &str) {
    fx.write(&format!(".moai-cli/{rel}"), body);
}

/// `hooks.json` text of exactly `total` bytes: only the `description` grows.
fn hooks_of(total: usize) -> String {
    let at = |pad: usize| format!(r#"{{"description":"{}","hooks":{{}}}}"#, "x".repeat(pad));
    let text = at(total - at(0).len());
    assert_eq!(text.len(), total);
    text
}

/// JSON text of a Rust record of exactly `total` bytes: only the `pad` value grows.
fn record_of(total: usize, base: Value) -> String {
    let at = |pad: usize| {
        let mut v = base.clone();
        v["pad"] = json!("x".repeat(pad));
        v.to_string()
    };
    let text = at(total - at(0).len());
    assert_eq!(text.len(), total);
    text
}

/// `config.toml` text of exactly `total` bytes: only a trailing comment grows.
fn config_of(total: usize) -> String {
    let at = |pad: usize| format!("{GO_CONFIG}# {}\n", "x".repeat(pad));
    let text = at(total - at(0).len());
    assert_eq!(text.len(), total);
    text
}

// ---- AC-1, AC-2: the agent files ----

#[test]
fn ac1_agent_files_matching_the_go_deploy_are_tool_assets_that_stay_for_now() {
    let mut fx = go_fx();
    fx.deployed(".codex/agents/moai/expert.toml", "name = \"expert\"\n");
    fx.manifest();
    let p = fx.plan();
    stays(&p, ".codex/agents/moai/expert.toml", "tool_asset");
    assert_eq!(item(&p, ".codex/agents/moai/expert.toml")["reason"], LATER);
    assert!(
        proof(&p, ".codex/agents/moai/expert.toml").contains("matches"),
        "{p}"
    );
    // Not in the git column: nothing moves in this slice.
    assert!(item(&p, ".codex/agents/moai/expert.toml")["git"].is_null());
}

#[test]
fn ac2_an_edited_modified_or_created_agent_file_keeps_its_own_class() {
    let mut fx = go_fx();
    fx.deployed(".codex/agents/moai/edited.toml", "name = \"a\"\n");
    fx.write(".codex/agents/moai/edited.toml", "name = \"a\" # mine\n");
    fx.write(".codex/agents/moai/recorded.toml", "name = \"b\"\n");
    fx.record(
        ".codex/agents/moai/recorded.toml",
        "user_modified",
        &sha(b"name = \"old\"\n"),
    );
    fx.write(".codex/agents/moai/mine.toml", "name = \"c\"\n");
    fx.record(
        ".codex/agents/moai/mine.toml",
        "user_created",
        &sha(b"name = \"c\"\n"),
    );
    fx.manifest();
    let p = fx.plan();
    stays(&p, ".codex/agents/moai/edited.toml", "user_modified");
    assert_eq!(item(&p, ".codex/agents/moai/edited.toml")["reason"], LATER);
    assert!(proof(&p, ".codex/agents/moai/edited.toml").contains("differs"));
    stays(&p, ".codex/agents/moai/recorded.toml", "user_modified");
    assert!(proof(&p, ".codex/agents/moai/recorded.toml").contains("user_modified"));
    // The user's own file is never a candidate.
    stays(&p, ".codex/agents/moai/mine.toml", "user_data");
    assert_ne!(item(&p, ".codex/agents/moai/mine.toml")["reason"], LATER);
}

// ---- AC-3: the hash forms ----

#[test]
fn ac3_a_wiring_hash_is_bare_hex_and_only_that_provenance_is_normalized() {
    let mut fx = go_fx();
    fx.wired(".codex/hooks.json", GO_HOOKS);
    fx.wired(".codex/config.toml", GO_CONFIG);
    // The same hash written with the prefix also names the same bytes.
    fx.write(".codex/agents/moai/p.toml", "p\n");
    fx.record(
        ".codex/agents/moai/p.toml",
        "generated_managed",
        &sha(b"p\n"),
    );
    // A template-managed record keeps the prefixed form: bare hex is not its hash.
    fx.write(".codex/agents/moai/q.toml", "q\n");
    fx.record(
        ".codex/agents/moai/q.toml",
        "template_managed",
        &hex(b"q\n"),
    );
    fx.manifest();
    let p = fx.plan();
    stays(&p, ".codex/hooks.json", "go_hooks");
    stays(&p, ".codex/config.toml", "go_config");
    assert_eq!(item(&p, ".codex/hooks.json")["reason"], LATER);
    assert_eq!(item(&p, ".codex/config.toml")["reason"], LATER);
    assert!(proof(&p, ".codex/hooks.json").contains("generated_managed"));
    // A generated_managed agent file is not a template asset, whatever its hash form: only a
    // template_managed or user_modified record is judged by `agent_item` at all.
    stays(&p, ".codex/agents/moai/p.toml", "unknown");
    assert_eq!(
        item(&p, ".codex/agents/moai/p.toml")["reason"],
        json!("not template_managed"),
        "{p}"
    );
    // The miss direction: normalization is for generated_managed only.
    stays(&p, ".codex/agents/moai/q.toml", "user_modified");
    // One warning line: a running Codex session reads these two files.
    assert!(
        p["warnings"].to_string().contains("running Codex session"),
        "{p}"
    );
}

#[test]
fn ac3_a_prefixed_wiring_hash_matches_too() {
    let mut fx = go_fx();
    fx.write(".codex/hooks.json", GO_HOOKS);
    fx.record_wired(".codex/hooks.json", &sha(GO_HOOKS.as_bytes()));
    fx.manifest();
    let p = fx.plan();
    stays(&p, ".codex/hooks.json", "go_hooks");
}

// ---- AC-4, AC-5: content proof ----

#[test]
fn ac4_hooks_json_needs_content_proof_and_prints_key_names_only() {
    let secret_cmd = "curl https://example.invalid/token-9f3a";
    // Each case: the file text and a string that must name why it stays.
    let cases: [(String, &str); 5] = [
        // a top-level key the Go wiring does not write
        (
            r#"{"description":"d","hooks":{},"private_notes":"value-5c1e"}"#.to_string(),
            "private_notes",
        ),
        // a handler that is not MoAI's
        (
            format!(
                r#"{{"hooks":{{"Stop":[{{"hooks":[{{"type":"command","command":"{secret_cmd}"}}]}}]}}}}"#
            ),
            "handler",
        ),
        // a command that only contains the prefix is not MoAI's
        (
            r#"{"hooks":{"Stop":[{"hooks":[{"type":"command","command":"x moai hook y"}]}]}}"#
                .to_string(),
            "handler",
        ),
        // not JSON at all
        ("moai hook ?\n".to_string(), "parse"),
        // the empty file
        (String::new(), "parse"),
    ];
    for (n, (text, why)) in cases.iter().enumerate() {
        let mut fx = go_fx();
        fx.wired(".codex/hooks.json", text);
        fx.manifest();
        let p = fx.plan();
        stays(&p, ".codex/hooks.json", "user_modified");
        assert_ne!(item(&p, ".codex/hooks.json")["reason"], LATER, "case {n}");
        assert!(
            proof(&p, ".codex/hooks.json").contains(why),
            "case {n}: {p}"
        );
        let all = p.to_string();
        for leaked in ["value-5c1e", "token-9f3a", "example.invalid"] {
            assert!(
                !all.contains(leaked),
                "case {n}: a value reached the output"
            );
        }
    }
}

#[test]
fn ac4_hooks_json_with_only_moai_handlers_is_go_wiring_and_empty_hooks_are_too() {
    let mut fx = go_fx();
    fx.wired(".codex/hooks.json", r#"{"hooks":{}}"#);
    fx.manifest();
    let p = fx.plan();
    stays(&p, ".codex/hooks.json", "go_hooks");
    // The namespace form: a command under `/.codex/hooks/moai/` is MoAI's (GO_HOOKS holds one).
    let mut fx = go_fx();
    fx.wired(".codex/hooks.json", GO_HOOKS);
    fx.manifest();
    stays(&fx.plan(), ".codex/hooks.json", "go_hooks");
}

#[test]
fn ac4_hooks_json_fails_closed_on_a_shape_or_a_handler_the_go_wiring_does_not_write() {
    // Each case: the file text and a string that must name why it stays.
    let cases: [(&str, &str, &str); 6] = [
        // a handler without a command
        (
            "no command",
            r#"{"hooks":{"Stop":[{"hooks":[{"type":"prompt"}]}]}}"#,
            "handler",
        ),
        // a command that is not text
        (
            "command not text",
            r#"{"hooks":{"Stop":[{"hooks":[{"type":"command","command":["moai hook x"]}]}]}}"#,
            "handler",
        ),
        // the hooks table is not a table
        (
            "hooks a string",
            r#"{"hooks":"moai hook x"}"#,
            "hooks table",
        ),
        // an event holds no list of entries
        (
            "event not a list",
            r#"{"hooks":{"Stop":"moai hook x"}}"#,
            "hooks table",
        ),
        // an entry holds no list of handlers
        (
            "entry hooks not a list",
            r#"{"hooks":{"Stop":[{"hooks":"moai hook x"}]}}"#,
            "hooks table",
        ),
        // a script under `.codex/` that is not under the Go namespace `.codex/hooks/moai/`
        (
            "script outside the namespace",
            r#"{"hooks":{"Stop":[{"hooks":[{"type":"command","command":"sh /p/.codex/evil.sh"}]}]}}"#,
            "handler",
        ),
    ];
    for (n, (label, text, why)) in cases.iter().enumerate() {
        let mut fx = go_fx();
        fx.wired(".codex/hooks.json", text);
        fx.manifest();
        let p = fx.plan();
        stays(&p, ".codex/hooks.json", "user_modified");
        assert_ne!(item(&p, ".codex/hooks.json")["reason"], LATER, "{label}");
        assert!(
            proof(&p, ".codex/hooks.json").contains(why),
            "case {n} {label}: {p}"
        );
        assert!(
            !p.to_string().contains("evil.sh"),
            "case {n}: a command reached the output"
        );
    }
    // The pass direction beside the namespace case: the Go namespace itself is MoAI's.
    let mut fx = go_fx();
    fx.wired(
        ".codex/hooks.json",
        r#"{"hooks":{"Stop":[{"hooks":[{"type":"command","command":"sh /p/.codex/hooks/moai/stop.sh"}]}]}}"#,
    );
    fx.manifest();
    stays(&fx.plan(), ".codex/hooks.json", "go_hooks");
}

#[test]
fn ac5_config_toml_needs_content_proof_and_prints_key_names_only() {
    let cases: [(String, &str); 7] = [
        // a key the Go wiring does not write
        (format!("model = \"value-77aa\"\n{GO_CONFIG}"), "model"),
        // another key inside `tui`
        (format!("{GO_CONFIG}theme = \"value-77ab\"\n"), "tui.theme"),
        // another MCP server
        (
            format!("{GO_CONFIG}\n[mcp_servers.other]\ncommand = \"value-77ac\"\n"),
            "mcp_servers.other",
        ),
        // the table value is not the Go value
        (
            GO_CONFIG.replace("command = \"moai\"", "command = \"value-77ad\""),
            "mcp_servers.moai",
        ),
        // a key the Go table does not hold
        (
            GO_CONFIG.replace(
                "default_tools_approval_mode = \"writes\"",
                "default_tools_approval_mode = \"writes\"\nextra = \"value-77ae\"",
            ),
            "mcp_servers.moai",
        ),
        // not TOML
        ("[mcp_servers.moai\n".to_string(), "parse"),
        // the Go table is missing: only the status line is left, plus a foreign key
        (
            "stray_key = \"value-77af\"\n[tui]\nstatus_line = [\"model\"]\n".to_string(),
            "stray_key",
        ),
    ];
    for (n, (text, why)) in cases.iter().enumerate() {
        let mut fx = go_fx();
        fx.wired(".codex/config.toml", text);
        fx.manifest();
        let p = fx.plan();
        stays(&p, ".codex/config.toml", "user_modified");
        assert_ne!(item(&p, ".codex/config.toml")["reason"], LATER, "case {n}");
        assert!(
            proof(&p, ".codex/config.toml").contains(why),
            "case {n}: {p}"
        );
        let all = p.to_string();
        assert!(
            !all.contains("value-77a"),
            "case {n}: a value reached the output"
        );
    }
}

#[test]
fn ac5_config_toml_with_the_go_table_and_a_status_line_alone_is_go_wiring() {
    for text in [
        GO_CONFIG.to_string(),
        // either surface alone, a comment, and any status line of strings
        format!("# mine\n{}", GO_CONFIG.split("[tui]").next().unwrap()),
        "[tui]\nstatus_line = [\"anything\", \"at-all\"]\n".to_string(),
    ] {
        let mut fx = go_fx();
        fx.wired(".codex/config.toml", &text);
        fx.manifest();
        let p = fx.plan();
        stays(&p, ".codex/config.toml", "go_config");
    }
}

#[test]
fn ac5_each_value_of_the_go_table_and_the_status_line_type_is_compared() {
    // Each case: the text and the key name the proof must show. A canary value must not.
    let cases: [(&str, String, &str); 8] = [
        (
            "args",
            GO_CONFIG.replace("args = [\"mcp-server\"]", "args = [\"value-77b1\"]"),
            "mcp_servers.moai",
        ),
        (
            "args extra",
            GO_CONFIG.replace("args = [\"mcp-server\"]", "args = [\"mcp-server\", \"value-77b2\"]"),
            "mcp_servers.moai",
        ),
        (
            "env_vars changed",
            GO_CONFIG.replace("\"MOAI_HOME\"", "\"VALUE-77B3\""),
            "mcp_servers.moai",
        ),
        (
            "env_vars shorter",
            GO_CONFIG.replace(", \"CLAUDE_CODE_SESSION_ID\"", ""),
            "mcp_servers.moai",
        ),
        (
            "approval mode",
            GO_CONFIG.replace(
                "default_tools_approval_mode = \"writes\"",
                "default_tools_approval_mode = \"value-77b4\"",
            ),
            "mcp_servers.moai",
        ),
        (
            "approval mode not text",
            GO_CONFIG.replace(
                "default_tools_approval_mode = \"writes\"",
                "default_tools_approval_mode = 7",
            ),
            "mcp_servers.moai",
        ),
        (
            "status_line a string",
            GO_CONFIG.replace(
                "status_line = [\"model-with-reasoning\", \"codex-version\", \"context-remaining\"]",
                "status_line = \"value-77b5\"",
            ),
            "tui.status_line",
        ),
        (
            "status_line holds a number",
            GO_CONFIG.replace("\"codex-version\"", "7"),
            "tui.status_line",
        ),
    ];
    for (n, (label, text, why)) in cases.iter().enumerate() {
        assert_ne!(text, GO_CONFIG, "case {label} changed nothing");
        let mut fx = go_fx();
        fx.wired(".codex/config.toml", text);
        fx.manifest();
        let p = fx.plan();
        stays(&p, ".codex/config.toml", "user_modified");
        assert_ne!(item(&p, ".codex/config.toml")["reason"], LATER, "{label}");
        assert!(
            proof(&p, ".codex/config.toml").contains(why),
            "case {n} {label}: {p}"
        );
        let all = p.to_string().to_ascii_lowercase();
        assert!(
            !all.contains("value-77b"),
            "case {n} {label}: a value reached the output"
        );
    }
}

// ---- AC-6: a stale hash ----

#[test]
fn ac6_a_hash_that_no_longer_matches_is_not_proof_even_for_go_looking_content() {
    let mut fx = go_fx();
    fx.write(".codex/hooks.json", GO_HOOKS);
    fx.record_wired(".codex/hooks.json", &hex(b"another version\n"));
    fx.write(".codex/config.toml", GO_CONFIG);
    fx.record_wired(".codex/config.toml", "");
    fx.manifest();
    let p = fx.plan();
    stays(&p, ".codex/hooks.json", "user_modified");
    assert!(proof(&p, ".codex/hooks.json").contains("differs"), "{p}");
    // No recorded hash at all is no proof either.
    stays(&p, ".codex/config.toml", "user_modified");
}

// ---- AC-7, AC-8, AC-9: what this generation made is never Go's ----

#[test]
fn ac7_a_file_in_the_rust_generation_manifest_is_never_classified_as_go() {
    let mut fx = go_fx();
    // Recorded by the Go manifest with a matching hash, and also listed by the Rust manifest.
    fx.deployed(".codex/agents/prj-rev.toml", "name = \"rev\"\n");
    // Listed by the Rust manifest only.
    fx.write(".codex/agents/prj-new.toml", "name = \"new\"\n");
    // Recorded by Go, not listed by Rust: not Rust's.
    fx.deployed(".codex/agents/prj-other.toml", "name = \"other\"\n");
    // Under the Go namespace but listed by Rust: still Rust's.
    fx.deployed(".codex/agents/moai/both.toml", "name = \"both\"\n");
    fx.deployed(".codex/agents/moai/go.toml", "name = \"go\"\n");
    fx.manifest();
    rust_file(
        &fx,
        "manifests/generation.json",
        &json!({"owned_files": [
            {"path": ".codex/agents/prj-rev.toml"},
            {"path": ".codex/agents/prj-new.toml"},
            {"path": ".codex/agents/moai/both.toml"},
            {"path": "AGENTS.md"}
        ]})
        .to_string(),
    );
    let p = fx.plan();
    for rel in [
        ".codex/agents/prj-rev.toml",
        ".codex/agents/prj-new.toml",
        ".codex/agents/moai/both.toml",
    ] {
        stays(&p, rel, "rust");
        assert_ne!(item(&p, rel)["reason"], LATER, "{rel}");
    }
    assert_ne!(class(&p, ".codex/agents/prj-other.toml"), "rust", "{p}");
    stays(&p, ".codex/agents/moai/go.toml", "tool_asset");
}

#[test]
fn ac7_a_rust_record_that_cannot_be_read_keeps_every_codex_file_unclassified() {
    let mut fx = go_fx();
    fx.deployed(".codex/agents/moai/go.toml", "name = \"go\"\n");
    fx.manifest();
    rust_file(&fx, "manifests/generation.json", "{ not json");
    let p = fx.plan();
    stays(&p, ".codex/agents/moai/go.toml", "unknown");
    assert!(
        item(&p, ".codex/agents/moai/go.toml")["reason"]
            .as_str()
            .unwrap()
            .contains("Rust"),
        "{p}"
    );
}

#[test]
fn ac8_a_hooks_json_the_rust_receipt_owns_is_never_classified_as_go() {
    for (key, other) in [("sha256", "pending_sha256"), ("pending_sha256", "sha256")] {
        let mut fx = go_fx();
        fx.wired(".codex/hooks.json", GO_HOOKS);
        fx.manifest();
        rust_file(
            &fx,
            "manifests/codex-hooks.json",
            &json!({"schema_version": 1, key: sha(GO_HOOKS.as_bytes()),
                    other: sha(b"another")})
            .to_string(),
        );
        let p = fx.plan();
        stays(&p, ".codex/hooks.json", "rust");
    }
    // The miss direction: a receipt for other bytes owns nothing here.
    let mut fx = go_fx();
    fx.wired(".codex/hooks.json", GO_HOOKS);
    fx.manifest();
    rust_file(
        &fx,
        "manifests/codex-hooks.json",
        &json!({"schema_version": 1, "sha256": sha(b"another")}).to_string(),
    );
    stays(&fx.plan(), ".codex/hooks.json", "go_hooks");
    // A hooks.json only the receipt knows (the Go manifest does not record it) is listed too.
    let fx = go_fx();
    fx.write(".codex/hooks.json", "{\"hooks\":{}}\n");
    fx.manifest();
    rust_file(
        &fx,
        "manifests/codex-hooks.json",
        &json!({"schema_version": 1, "sha256": sha(b"{\"hooks\":{}}\n")}).to_string(),
    );
    stays(&fx.plan(), ".codex/hooks.json", "rust");
}

#[test]
fn ac9_a_config_toml_with_rust_owned_keys_is_never_classified_as_go() {
    let receipt = |owned: Value| {
        json!({"schema_version": 1, "config_path": ".codex/config.toml",
               "owned": owned, "config_sha256": null})
        .to_string()
    };
    let mut fx = go_fx();
    fx.wired(".codex/config.toml", GO_CONFIG);
    fx.manifest();
    rust_file(
        &fx,
        "manifests/codex-config.json",
        &receipt(json!({"tui.status_line": ["model"]})),
    );
    stays(&fx.plan(), ".codex/config.toml", "rust");

    // The miss direction: a receipt that owns no key owns nothing here.
    let mut fx = go_fx();
    fx.wired(".codex/config.toml", GO_CONFIG);
    fx.manifest();
    rust_file(&fx, "manifests/codex-config.json", &receipt(json!({})));
    stays(&fx.plan(), ".codex/config.toml", "go_config");

    // A receipt that cannot be read is not "owns nothing".
    let mut fx = go_fx();
    fx.wired(".codex/config.toml", GO_CONFIG);
    fx.manifest();
    rust_file(&fx, "manifests/codex-config.json", "[");
    stays(&fx.plan(), ".codex/config.toml", "unknown");
}

/// The Rust config receipt, the way `codex_config` writes it (`pending` only mid-write).
fn config_receipt(owned: Value, pending_owned: Option<Value>) -> String {
    let mut r = json!({"schema_version": 1, "config_path": ".codex/config.toml",
                       "owned": owned, "config_sha256": null});
    if let Some(o) = pending_owned {
        r["pending"] = json!({"owned": o, "config_sha256": sha(b"next")});
    }
    r.to_string()
}

#[test]
fn ac9_a_config_toml_only_the_pending_part_of_the_receipt_owns_is_never_go() {
    // An interrupted Rust write: `owned` is still empty, the keys are only in `pending`.
    let mut fx = go_fx();
    fx.wired(".codex/config.toml", GO_CONFIG);
    fx.manifest();
    rust_file(
        &fx,
        "manifests/codex-config.json",
        &config_receipt(json!({}), Some(json!({"tui.status_line": ["model"]}))),
    );
    stays(&fx.plan(), ".codex/config.toml", "rust");
    // The miss direction: a pending part that owns no key owns nothing here.
    let mut fx = go_fx();
    fx.wired(".codex/config.toml", GO_CONFIG);
    fx.manifest();
    rust_file(
        &fx,
        "manifests/codex-config.json",
        &config_receipt(json!({}), Some(json!({}))),
    );
    stays(&fx.plan(), ".codex/config.toml", "go_config");
}

#[test]
fn ac9_a_config_toml_only_the_receipt_knows_is_listed_as_rust() {
    // The config twin of the last AC-8 case: the Go manifest does not record the file.
    let fx = go_fx();
    fx.write(".codex/config.toml", GO_CONFIG);
    fx.manifest();
    rust_file(
        &fx,
        "manifests/codex-config.json",
        &config_receipt(json!({"tui.status_line": ["model"]}), None),
    );
    let p = fx.plan();
    stays(&p, ".codex/config.toml", "rust");
    assert_ne!(item(&p, ".codex/config.toml")["reason"], LATER, "{p}");
    // A receipt that owns no key lists nothing: the unrecorded file has no item.
    let fx = go_fx();
    fx.write(".codex/config.toml", GO_CONFIG);
    fx.manifest();
    rust_file(
        &fx,
        "manifests/codex-config.json",
        &config_receipt(json!({}), None),
    );
    let p = fx.plan();
    assert!(!has_item(&p, ".codex/config.toml"), "{p}");
}

#[test]
fn ac9_a_hooks_receipt_of_another_schema_is_a_broken_record_not_an_owner() {
    let hooks_sha = sha(GO_HOOKS.as_bytes());
    for (n, receipt) in [
        json!({"schema_version": 2, "sha256": hooks_sha}),
        json!({"schema_version": "1", "sha256": hooks_sha}),
        json!({"sha256": hooks_sha}),
    ]
    .iter()
    .enumerate()
    {
        let fx = wired_fx();
        fx.manifest();
        rust_file(&fx, "manifests/codex-hooks.json", &receipt.to_string());
        let p = fx.plan();
        // Not "owned by Rust" and not "Go's": nothing under `.codex` is proven.
        for rel in [".codex/hooks.json", ".codex/agents/moai/expert.toml"] {
            stays(&p, rel, "unknown");
            assert!(
                item(&p, rel)["reason"].as_str().unwrap().contains("Rust"),
                "case {n} {rel}: {p}"
            );
        }
    }
}

#[test]
fn ac9_a_config_receipt_of_another_schema_or_without_an_owned_table_is_broken() {
    let keys = json!({"tui.status_line": ["model"]});
    for (n, receipt) in [
        json!({"schema_version": 2, "owned": keys}),
        json!({"schema_version": "1", "owned": keys}),
        json!({"owned": keys}),
        json!({"schema_version": 1, "owned": ["tui.status_line"]}),
        json!({"schema_version": 1, "owned": "tui.status_line"}),
        json!({"schema_version": 1, "owned": null}),
        json!({"schema_version": 1}),
    ]
    .iter()
    .enumerate()
    {
        let fx = wired_fx();
        fx.manifest();
        rust_file(&fx, "manifests/codex-config.json", &receipt.to_string());
        let p = fx.plan();
        for rel in [".codex/config.toml", ".codex/agents/moai/expert.toml"] {
            stays(&p, rel, "unknown");
            assert!(
                item(&p, rel)["reason"].as_str().unwrap().contains("Rust"),
                "case {n} {rel}: {p}"
            );
        }
    }
}

#[test]
fn ac7_a_generation_manifest_whose_owned_files_is_not_a_list_is_a_broken_record() {
    for (n, owned_files) in [json!({}), json!("AGENTS.md"), json!(null), json!(7)]
        .iter()
        .enumerate()
    {
        let mut fx = go_fx();
        fx.deployed(".codex/agents/moai/go.toml", "name = \"go\"\n");
        fx.manifest();
        rust_file(
            &fx,
            "manifests/generation.json",
            &json!({ "owned_files": owned_files }).to_string(),
        );
        let p = fx.plan();
        stays(&p, ".codex/agents/moai/go.toml", "unknown");
        assert!(
            item(&p, ".codex/agents/moai/go.toml")["reason"]
                .as_str()
                .unwrap()
                .contains("Rust"),
            "case {n}: {p}"
        );
    }
    // The miss direction: no `owned_files` key at all owns nothing, and is not broken.
    let mut fx = go_fx();
    fx.deployed(".codex/agents/moai/go.toml", "name = \"go\"\n");
    fx.manifest();
    rust_file(&fx, "manifests/generation.json", "{}");
    stays(&fx.plan(), ".codex/agents/moai/go.toml", "tool_asset");
}

#[test]
fn ac9_a_hooks_receipt_that_cannot_be_read_keeps_every_codex_file_unclassified() {
    let receipt = json!({"schema_version": 1, "sha256": sha(GO_HOOKS.as_bytes())}).to_string();
    for kind in ["fifo", "link", "torn", "empty"] {
        let fx = wired_fx();
        fx.manifest();
        let at = ".moai-cli/manifests/codex-hooks.json";
        match kind {
            // A FIFO would hang a reader (`run_env` ends the test instead).
            "fifo" => fx.fifo(at),
            "link" => {
                let target = fx.outside("receipt.json", &receipt);
                fx.link(at, target.to_str().unwrap());
            }
            "torn" => fx.write(at, "{ torn"),
            _ => fx.write(at, ""),
        }
        let p = fx.plan();
        for rel in [".codex/hooks.json", ".codex/agents/moai/expert.toml"] {
            stays(&p, rel, "unknown");
            assert!(
                item(&p, rel)["reason"].as_str().unwrap().contains("Rust"),
                "{kind} {rel}: {p}"
            );
        }
    }
}

// ---- AC-10: interrupted wiring ----

fn wired_fx() -> Fx {
    let mut fx = go_fx();
    fx.wired(".codex/hooks.json", GO_HOOKS);
    fx.wired(".codex/config.toml", GO_CONFIG);
    fx.deployed(".codex/agents/moai/expert.toml", "name = \"expert\"\n");
    fx
}

fn journal(states: &[&str]) -> String {
    let entries: Vec<Value> = states
        .iter()
        .enumerate()
        .map(|(n, s)| {
            json!({"id": format!("e{n}"), "op": "write", "path": ".codex/hooks.json",
                   "pre_hash": "", "post_hash": "", "state": s})
        })
        .collect();
    json!({"entries": entries}).to_string()
}

fn assert_everything_stays(p: &Value) {
    for rel in [
        ".codex/hooks.json",
        ".codex/config.toml",
        ".codex/agents/moai/expert.toml",
    ] {
        stays(p, rel, "interrupted");
        assert_ne!(item(p, rel)["reason"], LATER, "{rel}");
    }
    assert!(
        p["warnings"]
            .to_string()
            .contains("interrupted Codex wiring"),
        "{p}"
    );
}

#[test]
fn ac10_a_wiring_temp_file_keeps_everything_under_codex_in_place() {
    let fx = wired_fx();
    fx.write(".codex/.codexwiring-9f2c", "half written");
    fx.manifest();
    assert_everything_stays(&fx.plan());
}

#[test]
fn ac10_a_go_journal_with_an_unfinished_entry_keeps_everything_in_place() {
    for states in [
        vec!["complete", "staged"],
        vec!["conflict"],
        vec!["diverged"],
    ] {
        let fx = wired_fx();
        fx.write(".moai/state/codex-wiring-journal.json", &journal(&states));
        fx.manifest();
        assert_everything_stays(&fx.plan());
    }
    // A journal that cannot be read says nothing either way: it stays.
    let fx = wired_fx();
    fx.write(".moai/state/codex-wiring-journal.json", "{ torn");
    fx.manifest();
    assert_everything_stays(&fx.plan());
}

#[test]
fn ac10_a_finished_journal_and_no_temp_file_are_not_interrupted() {
    let fx = wired_fx();
    fx.write(
        ".moai/state/codex-wiring-journal.json",
        &journal(&["complete", "discarded", "complete"]),
    );
    // A temp-looking name that is not the wiring prefix.
    fx.write(".codex/codexwiring-x", "not a wiring temp");
    fx.manifest();
    let p = fx.plan();
    stays(&p, ".codex/hooks.json", "go_hooks");
    stays(&p, ".codex/agents/moai/expert.toml", "tool_asset");
    assert!(
        !p["warnings"]
            .to_string()
            .contains("interrupted Codex wiring"),
        "{p}"
    );
}

#[test]
fn ac10_a_journal_that_cannot_be_read_as_a_file_keeps_everything_in_place() {
    for kind in ["link", "fifo", "directory"] {
        let fx = wired_fx();
        let at = ".moai/state/codex-wiring-journal.json";
        match kind {
            // A finished journal behind a link: the link alone makes it unreadable.
            "link" => {
                let target = fx.outside("journal.json", &journal(&["complete"]));
                fx.link(at, target.to_str().unwrap());
            }
            // A FIFO would hang a reader (`run_env` ends the test instead).
            "fifo" => fx.fifo(at),
            _ => fs::create_dir_all(fx.root.join(at)).unwrap(),
        }
        fx.manifest();
        let p = fx.plan();
        assert_everything_stays(&p);
        assert!(
            p["warnings"].to_string().contains("journal cannot be read"),
            "{kind}: {p}"
        );
    }
}

#[test]
fn ac10_a_journal_whose_entries_are_not_a_list_keeps_everything_in_place() {
    for (n, text) in [
        r#"{"entries": {"e0": {"state": "complete"}}}"#,
        r#"{"entries": "complete"}"#,
        r#"{"entries": null}"#,
    ]
    .iter()
    .enumerate()
    {
        let fx = wired_fx();
        fx.write(".moai/state/codex-wiring-journal.json", text);
        fx.manifest();
        let p = fx.plan();
        assert_everything_stays(&p);
        assert!(
            p["warnings"].to_string().contains("journal cannot be read"),
            "case {n}: {p}"
        );
    }
    // The miss direction: no `entries` key at all says nothing, and nothing is interrupted.
    let fx = wired_fx();
    fx.write(".moai/state/codex-wiring-journal.json", "{}");
    fx.manifest();
    stays(&fx.plan(), ".codex/hooks.json", "go_hooks");
}

#[test]
fn ac10_a_journal_entry_without_a_state_is_not_a_finished_one() {
    for (n, entries) in [
        json!([{"id": "e0", "op": "write", "path": ".codex/hooks.json"}]),
        json!([{"id": "e0", "state": "complete"}, {"id": "e1", "op": "write"}]),
        json!([{"id": "e0", "state": 7}]),
        json!([null]),
    ]
    .iter()
    .enumerate()
    {
        let fx = wired_fx();
        fx.write(
            ".moai/state/codex-wiring-journal.json",
            &json!({ "entries": entries }).to_string(),
        );
        fx.manifest();
        let p = fx.plan();
        assert_everything_stays(&p);
        assert!(
            p["warnings"].to_string().contains("unfinished entry"),
            "case {n}: {p}"
        );
    }
    // The miss direction: an empty list has no unfinished entry.
    let fx = wired_fx();
    fx.write(".moai/state/codex-wiring-journal.json", r#"{"entries":[]}"#);
    fx.manifest();
    stays(&fx.plan(), ".codex/hooks.json", "go_hooks");
}

// ---- AC-11: symlinks ----

#[test]
fn ac11_a_symlink_on_the_path_keeps_the_file_in_place_unread() {
    let mut fx = go_fx();
    // `.codex/agents` is a link to a folder outside the project.
    let body = "name = \"expert\"\n";
    let target = fx.outside("agents/moai/expert.toml", body);
    fx.link(
        ".codex/agents",
        target.parent().unwrap().parent().unwrap().to_str().unwrap(),
    );
    fx.record(
        ".codex/agents/moai/expert.toml",
        "template_managed",
        &sha(body.as_bytes()),
    );
    // `hooks.json` itself is a link to a file outside the project with Go content.
    let hooks = fx.outside("hooks.json", GO_HOOKS);
    fx.link(".codex/hooks.json", hooks.to_str().unwrap());
    fx.record_wired(".codex/hooks.json", &hex(GO_HOOKS.as_bytes()));
    // A link to a FIFO would hang a reader.
    fx.fifo("elsewhere.fifo");
    fx.link(".codex/config.toml", "../elsewhere.fifo");
    fx.record_wired(".codex/config.toml", &hex(GO_CONFIG.as_bytes()));
    fx.manifest();
    let p = fx.plan();
    for rel in [
        ".codex/agents/moai/expert.toml",
        ".codex/hooks.json",
        ".codex/config.toml",
    ] {
        stays(&p, rel, "unknown");
        assert!(
            item(&p, rel)["reason"]
                .as_str()
                .unwrap()
                .contains("symlink"),
            "{rel}: {p}"
        );
    }
}

#[test]
fn ac11_a_linked_codex_folder_keeps_everything_and_is_not_counted() {
    let mut fx = go_fx();
    let body = "name = \"expert\"\n";
    let target = fx.outside("codex/agents/moai/expert.toml", body);
    fx.link(
        ".codex",
        target
            .parent()
            .unwrap()
            .parent()
            .unwrap()
            .parent()
            .unwrap()
            .to_str()
            .unwrap(),
    );
    fx.record(
        ".codex/agents/moai/expert.toml",
        "template_managed",
        &sha(body.as_bytes()),
    );
    fx.manifest();
    let p = fx.plan();
    stays(&p, ".codex/agents/moai/expert.toml", "unknown");
    assert!(
        p["unrecorded"]
            .as_object()
            .unwrap()
            .keys()
            .all(|k| !k.starts_with(".codex")),
        "{p}"
    );
}

#[test]
fn ac11_a_fifo_named_hooks_json_is_not_a_regular_file_and_is_never_opened() {
    let mut fx = go_fx();
    // The Go manifest records it with the hash of the Go wiring; only a reader that opened the
    // FIFO would wait for a writer (`run_env` ends the test after `DEADLINE` instead).
    fx.fifo(".codex/hooks.json");
    fx.record_wired(".codex/hooks.json", &hex(GO_HOOKS.as_bytes()));
    fx.manifest();
    let p = fx.plan();
    stays(&p, ".codex/hooks.json", "unknown");
    assert_eq!(
        item(&p, ".codex/hooks.json")["reason"],
        json!("not a regular file"),
        "{p}"
    );
    // With a Rust hooks receipt beside it, the receipt's sha cannot be matched by opening it.
    let mut fx = go_fx();
    fx.fifo(".codex/hooks.json");
    fx.record_wired(".codex/hooks.json", &hex(GO_HOOKS.as_bytes()));
    fx.manifest();
    rust_file(
        &fx,
        "manifests/codex-hooks.json",
        &json!({"schema_version": 1, "sha256": sha(GO_HOOKS.as_bytes())}).to_string(),
    );
    let p = fx.plan();
    stays(&p, ".codex/hooks.json", "unknown");
    assert_eq!(
        item(&p, ".codex/hooks.json")["reason"],
        json!("not a regular file"),
        "{p}"
    );
}

#[test]
fn ac11_a_fifo_in_place_of_any_other_codex_file_is_not_a_regular_file_either() {
    let mut fx = go_fx();
    fx.fifo(".codex/config.toml");
    fx.record_wired(".codex/config.toml", &hex(GO_CONFIG.as_bytes()));
    // An agent file the Go manifest records, whose own check would say "not a bounded ..."
    fx.fifo(".codex/agents/moai/expert.toml");
    fx.record(
        ".codex/agents/moai/expert.toml",
        "template_managed",
        &sha(b"name = \"expert\"\n"),
    );
    // A FIFO the Rust generation manifest lists: it is not "Rust's file" either.
    fx.fifo(".codex/agents/prj-fifo.toml");
    fx.manifest();
    rust_file(
        &fx,
        "manifests/generation.json",
        &json!({"owned_files": [{"path": ".codex/agents/prj-fifo.toml"}]}).to_string(),
    );
    let p = fx.plan();
    for rel in [
        ".codex/config.toml",
        ".codex/agents/moai/expert.toml",
        ".codex/agents/prj-fifo.toml",
    ] {
        stays(&p, rel, "unknown");
        assert_eq!(item(&p, rel)["reason"], json!("not a regular file"), "{p}");
    }
}

// ---- AC-12: secret names ----

#[test]
fn ac12_secret_names_under_codex_are_counted_and_never_opened() {
    let mut fx = go_fx();
    // A FIFO under a secret name would hang a reader; a recorded one must still not be listed.
    fx.fifo(".codex/.env");
    fx.record(".codex/.env", "template_managed", &sha(b"x"));
    fx.fifo(".codex/agents/credentials.json");
    fx.write(".codex/agents/moai/server.pem", "not read\n");
    fx.record(
        ".codex/agents/moai/server.pem",
        "template_managed",
        &sha(b"not read\n"),
    );
    fx.deployed(".codex/agents/moai/ok.toml", "name = \"ok\"\n");
    fx.manifest();
    let p = fx.plan();
    for rel in [
        ".codex/.env",
        ".codex/agents/credentials.json",
        ".codex/agents/moai/server.pem",
    ] {
        assert!(!has_item(&p, rel), "{rel}: {p}");
    }
    assert_eq!(p["secrets"], json!(3), "{p}");
    stays(&p, ".codex/agents/moai/ok.toml", "tool_asset");
}

#[test]
fn ac12_a_secret_name_the_rust_generation_owns_has_no_item_and_is_counted() {
    // `credentials*` (home_state) and `*.key` (worktree_include) are secret names; the Rust
    // manifest lists both, and also a plain name as the control.
    let mut fx = go_fx();
    fx.deployed(".codex/agents/moai/ok.toml", "name = \"ok\"\n");
    fx.write(".codex/agents/credentials.toml", "token = \"x\"\n");
    fx.write(".codex/agents/server.key", "not read\n");
    fx.write(".codex/agents/prj-ok.toml", "name = \"prj\"\n");
    fx.manifest();
    rust_file(
        &fx,
        "manifests/generation.json",
        &json!({"owned_files": [
            {"path": ".codex/agents/credentials.toml"},
            {"path": ".codex/agents/server.key"},
            {"path": ".codex/agents/prj-ok.toml"}
        ]})
        .to_string(),
    );
    let p = fx.plan();
    for rel in [".codex/agents/credentials.toml", ".codex/agents/server.key"] {
        assert!(!has_item(&p, rel), "{rel}: {p}");
    }
    assert_eq!(p["secrets"], json!(2), "{p}");
    stays(&p, ".codex/agents/prj-ok.toml", "rust");
    stays(&p, ".codex/agents/moai/ok.toml", "tool_asset");
    // The name is not echoed anywhere else in the preview either.
    assert!(!p.to_string().contains("credentials.toml"), "{p}");
    assert!(!p.to_string().contains("server.key"), "{p}");
}

// ---- AC-13: counted among the unrecorded files ----

#[test]
fn ac13_unrecorded_codex_files_are_counted_and_the_limit_has_an_edge() {
    let mut fx = Fx::new();
    fx.deployed(".moai/config/a.yaml", "a\n");
    fx.write(".codex/agents/note.md", "n\n");
    fx.manifest();
    // .moai/manifest.json, .moai/config, .moai/config/a.yaml, .codex/agents,
    // .codex/agents/note.md: five entries.
    let stopped = |n: &str| {
        let p = fx.plan_env(&[("MOAI_CLI_TEST_UNRECORDED_LIMIT", n)]);
        p["warnings"].to_string().contains("stopped after")
    };
    assert!(!stopped("5"), "five entries fit a limit of five");
    assert!(stopped("4"), "the fifth entry passes a limit of four");
    let p = fx.plan();
    assert_eq!(p["unrecorded"][".codex/agents"], json!(1), "{p}");
    // A recorded file is not counted as unrecorded.
    let mut fx = go_fx();
    fx.deployed(".codex/agents/moai/expert.toml", "name = \"x\"\n");
    fx.manifest();
    let p = fx.plan();
    assert!(p["unrecorded"].get(".codex/agents/moai").is_none(), "{p}");
}

// ---- AC-14: the read limit ----

#[test]
fn ac14_hooks_json_and_config_toml_are_read_up_to_exactly_the_limit() {
    let mut fx = go_fx();
    fx.wired(".codex/hooks.json", &hooks_of(CONFIG_LIMIT));
    fx.wired(".codex/config.toml", &config_of(CONFIG_LIMIT));
    fx.manifest();
    let p = fx.plan();
    stays(&p, ".codex/hooks.json", "go_hooks");
    stays(&p, ".codex/config.toml", "go_config");

    let mut fx = go_fx();
    fx.wired(".codex/hooks.json", &hooks_of(CONFIG_LIMIT + 1));
    fx.wired(".codex/config.toml", &config_of(CONFIG_LIMIT + 1));
    fx.manifest();
    let p = fx.plan();
    for rel in [".codex/hooks.json", ".codex/config.toml"] {
        stays(&p, rel, "unknown");
        assert!(
            item(&p, rel)["reason"].as_str().unwrap().contains("262144"),
            "{rel}: {p}"
        );
    }
}

#[test]
fn ac14_the_rust_records_are_read_up_to_exactly_the_limit() {
    let reason = |p: &Value, rel: &str| item(p, rel)["reason"].as_str().unwrap().to_string();
    // The generation manifest: read at N (it owns nothing here), a broken record at N+1.
    for (len, want) in [(CONFIG_LIMIT, "tool_asset"), (CONFIG_LIMIT + 1, "unknown")] {
        let mut fx = go_fx();
        fx.deployed(".codex/agents/moai/go.toml", "name = \"go\"\n");
        fx.manifest();
        rust_file(
            &fx,
            "manifests/generation.json",
            &record_of(len, json!({"owned_files": [{"path": "AGENTS.md"}]})),
        );
        let p = fx.plan();
        stays(&p, ".codex/agents/moai/go.toml", want);
        assert_eq!(
            want == "unknown",
            reason(&p, ".codex/agents/moai/go.toml").contains("Rust")
        );
    }
    // The hooks receipt: read at N (its sha owns the file), a broken record at N+1.
    for (len, want) in [(CONFIG_LIMIT, "rust"), (CONFIG_LIMIT + 1, "unknown")] {
        let fx = wired_fx();
        fx.manifest();
        rust_file(
            &fx,
            "manifests/codex-hooks.json",
            &record_of(
                len,
                json!({"schema_version": 1, "sha256": sha(GO_HOOKS.as_bytes())}),
            ),
        );
        let p = fx.plan();
        stays(&p, ".codex/hooks.json", want);
    }
    // The config receipt: read at N (its keys own the file), a broken record at N+1.
    for (len, want) in [(CONFIG_LIMIT, "rust"), (CONFIG_LIMIT + 1, "unknown")] {
        let fx = wired_fx();
        fx.manifest();
        rust_file(
            &fx,
            "manifests/codex-config.json",
            &record_of(
                len,
                json!({"schema_version": 1, "owned": {"tui.status_line": ["model"]}}),
            ),
        );
        let p = fx.plan();
        stays(&p, ".codex/config.toml", want);
    }
}

// ---- AC-15: nothing moves ----

#[test]
fn ac15_the_apply_moves_nothing_under_codex_and_hook_scripts_stay_in_place() {
    let mut fx = wired_fx();
    fx.deployed(".codex/hooks/moai/start.sh", "#!/bin/sh\n");
    fx.manifest();
    let p = fx.plan();
    stays(&p, ".codex/hooks/moai/start.sh", "go_hook_script");
    assert!(
        item(&p, ".codex/hooks/moai/start.sh")["reason"]
            .as_str()
            .unwrap()
            .contains("Q4"),
        "{p}"
    );
    for i in p["items"].as_array().unwrap() {
        if i["path"].as_str().unwrap().starts_with(".codex") {
            assert_eq!(i["action"], json!("leave"), "{i}");
        }
    }
    let codex_before: Vec<_> = fx
        .tree()
        .into_iter()
        .filter(|(path, _)| path.starts_with(fx.root.join(".codex")))
        .collect();
    let v = fx.apply(p["plan_digest"].as_str().unwrap());
    assert_eq!(v["ok"], json!(true), "{v}");
    assert_eq!(v["result"]["state"], json!("committed"), "{v}");
    let codex_after: Vec<_> = fx
        .tree()
        .into_iter()
        .filter(|(path, _)| path.starts_with(fx.root.join(".codex")))
        .collect();
    assert_eq!(codex_before, codex_after, "the apply touched .codex");
    // The receipt names no `.codex` path.
    let backup = v["result"]["backup"].as_str().unwrap();
    let receipt = fs::read_to_string(
        fx.root
            .join(".moai/backups")
            .join(backup)
            .join("go-state-receipt.json"),
    )
    .unwrap();
    assert!(!receipt.contains(".codex"), "{receipt}");
}
