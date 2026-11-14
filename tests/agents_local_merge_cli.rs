//! #2936 R10 (user decision 6880117157): `AGENTS.local.md` merges into `AGENTS.md` outside the
//! managed section — preview, apply, revert. Temporary projects and a temporary HOME only; no
//! host session. Credential-shaped text is built at run time so no fixture here is a secret.

#![cfg(unix)]

mod common;
use common::cli::{cli, create, ok_result, profile, tree_state};
use common::CleanEnv;
use serde_json::{json, Value};
use std::{
    fs,
    os::unix::fs::symlink,
    path::PathBuf,
    process::{Command, Stdio},
};
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;
const REF: &str = "test-permission-r10";
const RULES: &str = "# Rules\n\nWrite tests first.\n";
const MINE: &str = "# My notes\n\nReply in Korean.\nPrefer small commits.\n";
/// The text `agents::render_local` writes for English and an ignored personal file.
const GENERATED: &str = "# Personal supplement\n\nResponse language: English\n\n<!-- Personal file MoAI reads explicitly when invoked. Keep secrets out; never share it in the repo or in Issues. -->\n";

struct Fx {
    _tmp: TempDir,
    home: PathBuf,
    root: PathBuf,
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
        }
    }

    /// A project with `AGENTS.md` (and `AGENTS.local.md` when `local` is given).
    fn with(agents: &str, local: Option<&str>) -> Fx {
        let fx = Fx::new();
        fx.write("AGENTS.md", agents);
        if let Some(l) = local {
            fx.write("AGENTS.local.md", l);
        }
        fx
    }

    fn p(&self) -> &str {
        self.root.to_str().unwrap()
    }

    fn write(&self, rel: &str, body: &str) {
        let p = self.root.join(rel);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        fs::write(p, body).unwrap();
    }

    fn read(&self, rel: &str) -> String {
        fs::read_to_string(self.root.join(rel)).unwrap()
    }

    fn run(&self, args: &[&str]) -> Value {
        let out = Command::new(EXE)
            .args(args)
            .clean_env()
            .env("HOME", &self.home)
            .env("PATH", "/usr/bin:/bin")
            .env("LC_ALL", "C")
            .stdin(Stdio::null())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .output()
            .unwrap();
        serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
            panic!(
                "{args:?}: {}{}",
                String::from_utf8_lossy(&out.stdout),
                String::from_utf8_lossy(&out.stderr)
            )
        })
    }

    /// The preview; it must succeed and write nothing.
    fn plan(&self) -> Value {
        let before = tree_state(&self.root);
        let v = self.run(&["agents-merge-plan", self.p()]);
        assert_eq!(v["ok"], json!(true), "{v}");
        assert_eq!(tree_state(&self.root), before, "a preview writes nothing");
        v["result"].clone()
    }

    fn apply_with(&self, digest: &str, permission: &str) -> Value {
        self.run(&[
            "agents-merge-apply",
            self.p(),
            "--plan-digest",
            digest,
            "--permission-reference",
            permission,
        ])
    }

    /// Plan and apply; the apply must succeed. Returns the apply result.
    fn merge(&self) -> Value {
        let p = self.plan();
        assert_eq!(p["action"], json!("merge"), "{p}");
        let v = self.apply_with(p["plan_digest"].as_str().unwrap(), REF);
        assert_eq!(v["ok"], json!(true), "{v}");
        v["result"].clone()
    }

    fn revert_with(&self, receipt: &str, permission: &str) -> Value {
        self.run(&[
            "agents-merge-revert",
            self.p(),
            "--receipt",
            receipt,
            "--permission-reference",
            permission,
        ])
    }
}

fn block(source: &str) -> String {
    use sha2::{Digest, Sha256};
    format!(
        "<!-- moai:merged-local sha256={:x} -->\n{}\n<!-- /moai:merged-local -->\n",
        Sha256::digest(source.as_bytes()),
        source.trim_end_matches('\n')
    )
}

/// R10-1: no `AGENTS.local.md`, nothing to do.
#[test]
fn no_personal_file_means_no_change() {
    let fx = Fx::with(RULES, None);
    let p = fx.plan();
    assert_eq!(p["action"], json!("none"), "{p}");
    assert_eq!(p["source"], Value::Null);
    let before = tree_state(&fx.root);
    let v = fx.apply_with(p["plan_digest"].as_str().unwrap(), REF);
    assert_eq!(v["result"]["state"], json!("unchanged"), "{v}");
    assert_eq!(tree_state(&fx.root), before);
}

/// R10-2: an empty file, a blank one and the text MoAI generated hold nothing of the user's.
#[test]
fn empty_blank_and_generated_files_merge_nothing() {
    for (body, word) in [
        ("", "empty"),
        ("  \n\n\t\n", "empty"),
        (GENERATED, "generated"),
    ] {
        let fx = Fx::with(RULES, Some(body));
        let p = fx.plan();
        assert_eq!(p["action"], json!("none"), "{body:?}: {p}");
        assert!(
            p["reason"].as_str().unwrap().contains(word),
            "{body:?}: {p}"
        );
        assert_eq!(fx.read("AGENTS.md"), RULES);
    }
}

/// R10-3, R10-4, R10-8: the text lands once at the end of `AGENTS.md`; the preview shows the
/// whole text, the position and the commit line; a second run is `unchanged`.
#[test]
fn personal_text_merges_once_and_a_rerun_changes_nothing() {
    let fx = Fx::with(RULES, Some(MINE));
    let p = fx.plan();
    assert_eq!(p["action"], json!("merge"), "{p}");
    assert_eq!(p["source"]["kind"], json!("in_place"));
    assert_eq!(p["body"], json!(MINE));
    assert_eq!(p["target"]["starts_at_line"], json!(5), "{p}");
    assert!(
        p["commit_notice"].as_str().unwrap().contains("committed"),
        "{p}"
    );
    assert_eq!(p["secret_suspects"], json!([]));

    let applied = fx.merge();
    assert_eq!(applied["state"], json!("applied"), "{applied}");
    assert_eq!(fx.read("AGENTS.md"), format!("{RULES}\n{}", block(MINE)));
    // The personal file stays where it is; the receipt keeps a copy.
    assert_eq!(fx.read("AGENTS.local.md"), MINE);
    let id = applied["receipt"].as_str().unwrap();
    assert_eq!(
        fx.read(&format!(".moai-cli/receipts/{id}/source/AGENTS.local.md")),
        MINE
    );
    assert_eq!(
        fx.read(&format!(".moai-cli/receipts/{id}/before/AGENTS.md")),
        RULES
    );

    // R10-4: a second preview and apply change no byte.
    let again = fx.plan();
    assert_eq!(again["action"], json!("unchanged"), "{again}");
    let before = tree_state(&fx.root);
    let v = fx.apply_with(again["plan_digest"].as_str().unwrap(), REF);
    assert_eq!(v["result"]["state"], json!("unchanged"), "{v}");
    assert_eq!(tree_state(&fx.root), before);
}

/// R10-9: a wrong digest, a missing permission reference and an `AGENTS.md` that changed after
/// the preview write nothing.
#[test]
fn a_stale_digest_or_no_permission_writes_nothing() {
    let fx = Fx::with(RULES, Some(MINE));
    let digest = fx.plan()["plan_digest"].as_str().unwrap().to_string();
    let before = tree_state(&fx.root);

    let v = fx.apply_with("sha256:wrong", REF);
    assert_eq!(v["ok"], json!(false), "{v}");
    assert!(v.to_string().contains("stale"), "{v}");
    for permission in ["", "null", "placeholder"] {
        let v = fx.apply_with(&digest, permission);
        assert_eq!(v["ok"], json!(false), "{permission:?}: {v}");
    }
    assert_eq!(tree_state(&fx.root), before);

    fx.write("AGENTS.md", &format!("{RULES}\nAdded after the preview.\n"));
    let moved = tree_state(&fx.root);
    let v = fx.apply_with(&digest, REF);
    assert_eq!(v["ok"], json!(false), "{v}");
    assert_eq!(tree_state(&fx.root), moved);
}

/// R10-5: a block edited by hand, a source changed since the merge and a forged marker are
/// conflicts: nothing is overwritten and the apply refuses.
#[test]
fn edits_and_forged_markers_are_conflicts() {
    let fx = Fx::with(RULES, Some(MINE));
    fx.merge();
    let merged = fx.read("AGENTS.md");

    // The block text edited by hand.
    fx.write("AGENTS.md", &merged.replace("small commits", "big commits"));
    let p = fx.plan();
    assert_eq!(p["action"], json!("conflict"), "{p}");
    let edited = tree_state(&fx.root);
    let v = fx.apply_with(p["plan_digest"].as_str().unwrap(), REF);
    assert_eq!(v["ok"], json!(false), "{v}");
    assert_eq!(tree_state(&fx.root), edited);

    // The personal file changed after the merge: the old block is kept.
    fx.write("AGENTS.md", &merged);
    fx.write("AGENTS.local.md", &format!("{MINE}One more line.\n"));
    let p = fx.plan();
    assert_eq!(p["action"], json!("conflict"), "{p}");
    assert!(p["reason"].as_str().unwrap().contains("changed"), "{p}");
    assert_eq!(fx.read("AGENTS.md"), merged);

    // A marker whose hash is not 64 hex digits is damaged, not trusted.
    fx.write("AGENTS.local.md", MINE);
    fx.write("AGENTS.md", &merged.replace("sha256=", "sha256=zz"));
    assert_eq!(fx.plan()["action"], json!("conflict"));
}

/// R10-8: a credential-shaped line blocks the plan; its place is named, its value is not.
#[test]
fn a_credential_shaped_line_blocks_without_printing_it() {
    let secret = format!("{}{}", "sk-", "q".repeat(32));
    let fx = Fx::with(RULES, Some(&format!("{MINE}api key: {secret}\n")));
    let p = fx.plan();
    assert_eq!(p["action"], json!("blocked"), "{p}");
    assert_eq!(
        p["secret_suspects"],
        json!([{"line": 5, "kind": "api_key"}])
    );
    assert!(!p.to_string().contains(&secret), "the value is never shown");
    let v = fx.apply_with(p["plan_digest"].as_str().unwrap(), REF);
    assert_eq!(v["ok"], json!(false), "{v}");
    assert!(!v.to_string().contains(&secret));
    assert_eq!(fx.read("AGENTS.md"), RULES);
}

/// R10-11: the size limit, a link, a missing `AGENTS.md` and a source that holds a MoAI marker.
#[test]
fn limits_links_and_markers_block_the_plan() {
    // 128 KiB exactly merges; one byte more is blocked.
    let line = "word ".repeat(15) + "w\n"; // 77 bytes of text per line
    let big = |n: usize| {
        let mut s = line.repeat(n / line.len() + 1);
        s.truncate(n);
        s
    };
    let fx = Fx::with(RULES, Some(&big(128 * 1024)));
    assert_eq!(fx.plan()["action"], json!("merge"));
    let fx = Fx::with(RULES, Some(&big(128 * 1024 + 1)));
    let p = fx.plan();
    assert_eq!(p["action"], json!("blocked"), "{}", p["reason"]);
    assert!(p["reason"].as_str().unwrap().contains("larger"), "{p}");

    // A link is never followed.
    let fx = Fx::with(RULES, None);
    fx.write("elsewhere.md", MINE);
    symlink(
        fx.root.join("elsewhere.md"),
        fx.root.join("AGENTS.local.md"),
    )
    .unwrap();
    let p = fx.plan();
    assert_eq!(p["action"], json!("blocked"), "{p}");

    // No AGENTS.md yet: wait for the install phase.
    let fx = Fx::with(RULES, Some(MINE));
    fs::remove_file(fx.root.join("AGENTS.md")).unwrap();
    let p = fx.plan();
    assert_eq!(p["action"], json!("blocked"), "{p}");
    assert!(p["reason"].as_str().unwrap().contains("install"), "{p}");

    // A source that holds a managed-section marker would be read as a second section.
    let marked = format!("{MINE}<!-- moai:managed:begin id=x -->\n");
    let fx = Fx::with(RULES, Some(&marked));
    assert_eq!(fx.plan()["action"], json!("blocked"));
}

/// R10-10: the revert puts the earlier `AGENTS.md` back, only while it still holds what the apply
/// wrote; it never touches the personal file.
#[test]
fn revert_restores_agents_md_only_while_it_is_unchanged() {
    let fx = Fx::with(RULES, Some(MINE));
    let id = fx.merge()["receipt"].as_str().unwrap().to_string();
    let merged = fx.read("AGENTS.md");

    let v = fx.revert_with(&id, "");
    assert_eq!(v["ok"], json!(false), "{v}");
    let v = fx.revert_with("agents-local-merge-../x", REF);
    assert_eq!(v["ok"], json!(false), "{v}");
    assert_eq!(fx.read("AGENTS.md"), merged);

    // Edited after the merge: refused, nothing written.
    fx.write("AGENTS.md", &format!("{merged}\nlater\n"));
    let v = fx.revert_with(&id, REF);
    assert_eq!(v["ok"], json!(false), "{v}");
    assert_eq!(fx.read("AGENTS.md"), format!("{merged}\nlater\n"));

    fx.write("AGENTS.md", &merged);
    let v = fx.revert_with(&id, REF);
    assert_eq!(v["result"]["state"], json!("reverted"), "{v}");
    assert_eq!(fx.read("AGENTS.md"), RULES);
    assert_eq!(fx.read("AGENTS.local.md"), MINE);
    // A second revert is a no-op; the text can be merged again.
    let v = fx.revert_with(&id, REF);
    assert_eq!(v["result"]["state"], json!("unchanged"), "{v}");
    assert_eq!(fx.plan()["action"], json!("merge"));
}

/// R10-7: after `go-state-apply` moved the file into its dated backup, that copy is the source
/// (a new generated file in place is ignored); a reverted apply is not a source.
#[test]
fn a_go_state_backup_copy_is_the_source() {
    let fx = Fx::new();
    fx.write(
        ".moai/manifest.json",
        &json!({"version": "v3.2.1-rc.6", "deployed_at": "2026-11-01T00:00:00Z", "files": {}})
            .to_string(),
    );
    fx.write(".moai/config/sections/a.yaml", "a: 1\n");
    fx.write("AGENTS.local.md", MINE);
    let plan = fx.run(&["go-state-plan", fx.p()]);
    assert_eq!(plan["ok"], json!(true), "{plan}");
    let local = plan["result"]["root_files"]
        .as_array()
        .unwrap()
        .iter()
        .find(|f| f["path"] == "AGENTS.local.md")
        .cloned()
        .unwrap_or_else(|| panic!("{plan}"));
    assert_eq!(local["action"], json!("move"), "{plan}");
    let applied = fx.run(&[
        "go-state-apply",
        fx.p(),
        "--plan-digest",
        plan["result"]["plan_digest"].as_str().unwrap(),
        "--permission-reference",
        REF,
    ]);
    assert_eq!(applied["ok"], json!(true), "{applied}");
    let backup = applied["result"]["backup"].as_str().unwrap().to_string();
    assert!(!fx.root.join("AGENTS.local.md").exists());

    // The install step wrote a new AGENTS.md and a new generated personal file.
    fx.write("AGENTS.md", RULES);
    fx.write("AGENTS.local.md", GENERATED);
    let p = fx.plan();
    assert_eq!(p["action"], json!("merge"), "{p}");
    assert_eq!(p["source"]["kind"], json!("backup"), "{p}");
    assert_eq!(
        p["source"]["path"],
        json!(format!(".moai/backups/{backup}/root-files/AGENTS.local.md"))
    );
    let applied = fx.merge();
    assert_eq!(applied["source"]["kind"], json!("backup"));
    assert_eq!(fx.read("AGENTS.md"), format!("{RULES}\n{}", block(MINE)));
    // The generated file in place and the backup copy are not touched.
    assert_eq!(fx.read("AGENTS.local.md"), GENERATED);
    assert_eq!(
        fx.read(&format!(
            ".moai/backups/{backup}/root-files/AGENTS.local.md"
        )),
        MINE
    );
    assert_eq!(fx.plan()["action"], json!("unchanged"));

    // A backup copy that was edited after the move is not trusted.
    fx.write(
        &format!(".moai/backups/{backup}/root-files/AGENTS.local.md"),
        "# tampered\n",
    );
    assert_eq!(fx.plan()["action"], json!("blocked"));
}

/// R10-6: regenerating the managed section does not remove the merged block.
#[test]
fn regenerating_the_managed_section_keeps_the_merged_block() {
    let root = TempDir::new().unwrap();
    fs::create_dir_all(root.path().join("skills/moai/workflows")).unwrap();
    create(&root, &profile(None));
    let p = root.path().to_str().unwrap();
    let request = |language: &str| {
        json!({"protocol_version": 1, "language": language,
               "workflow_reference": "skills/moai/workflows/", "local_file": "ignored"})
    };
    let agents = |language: &str| {
        let plan = ok_result(&cli(&["agents-plan", p], &request(language)));
        let out = cli(
            &[
                "agents-apply",
                p,
                "--plan-digest",
                plan["plan_digest"].as_str().unwrap(),
                "--permission-reference",
                REF,
            ],
            &plan,
        );
        assert!(out.status.success(), "{:?}", out.stderr);
    };
    agents("en");
    // The generated personal file is replaced by the user's own text.
    fs::write(root.path().join("AGENTS.local.md"), MINE).unwrap();

    let merge = |args: &[&str]| -> Value {
        let out = Command::new(EXE)
            .args(args)
            .clean_env()
            .env("HOME", root.path())
            .stdin(Stdio::null())
            .output()
            .unwrap();
        serde_json::from_slice(&out.stdout).unwrap()
    };
    let plan = merge(&["agents-merge-plan", p]);
    assert_eq!(plan["result"]["action"], json!("merge"), "{plan}");
    let applied = merge(&[
        "agents-merge-apply",
        p,
        "--plan-digest",
        plan["result"]["plan_digest"].as_str().unwrap(),
        "--permission-reference",
        REF,
    ]);
    assert_eq!(applied["ok"], json!(true), "{applied}");
    let merged = fs::read_to_string(root.path().join("AGENTS.md")).unwrap();
    assert!(merged.ends_with(&block(MINE)), "{merged}");

    agents("ko");
    let regenerated = fs::read_to_string(root.path().join("AGENTS.md")).unwrap();
    assert_ne!(regenerated, merged, "the managed section was regenerated");
    assert!(regenerated.ends_with(&block(MINE)), "{regenerated}");
    assert_eq!(regenerated.matches("moai:merged-local sha256=").count(), 1);
    let again = merge(&["agents-merge-plan", p]);
    assert_eq!(again["result"]["action"], json!("unchanged"), "{again}");
}
