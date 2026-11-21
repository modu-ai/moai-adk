//! `AGENTS.local.md` moves into `CLAUDE.local.md`, the git-ignored personal file Claude Code
//! reads: preview, apply, revert. A missing `CLAUDE.local.md` is created and its revert removes
//! it. Temporary projects and a temporary HOME only; no host session. Credential-shaped text is
//! built at run time so no fixture here is a secret.

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

    /// A project whose `CLAUDE.local.md` already holds `target` (and `AGENTS.local.md` when
    /// `local` is given).
    fn with(target: &str, local: Option<&str>) -> Fx {
        let fx = Fx::new();
        fx.write("CLAUDE.local.md", target);
        if let Some(l) = local {
            fx.write("AGENTS.local.md", l);
        }
        fx
    }

    /// A project with no `CLAUDE.local.md` yet; `AGENTS.local.md` holds `local`.
    fn fresh(local: &str) -> Fx {
        let fx = Fx::new();
        fx.write("AGENTS.local.md", local);
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
        assert_eq!(fx.read("CLAUDE.local.md"), RULES);
    }
}

/// R10-3, R10-4, R10-8: the text lands once at the end of `CLAUDE.local.md`; the preview shows the
/// whole text, the position and the commit line; a second run is `unchanged`.
#[test]
fn personal_text_merges_once_and_a_rerun_changes_nothing() {
    let fx = Fx::with(RULES, Some(MINE));
    let p = fx.plan();
    assert_eq!(p["action"], json!("merge"), "{p}");
    assert_eq!(p["source"]["kind"], json!("in_place"));
    assert_eq!(p["body"], json!(MINE));
    assert_eq!(p["target"]["starts_at_line"], json!(5), "{p}");
    assert_eq!(p["target"]["path"], json!("CLAUDE.local.md"), "{p}");
    let notice = p["commit_notice"].as_str().unwrap();
    assert!(
        notice.contains("CLAUDE.local.md") && notice.contains("not committed"),
        "{p}"
    );
    assert_eq!(p["secret_suspects"], json!([]));

    let applied = fx.merge();
    assert_eq!(applied["state"], json!("applied"), "{applied}");
    assert_eq!(
        fx.read("CLAUDE.local.md"),
        format!("{RULES}\n{}", block(MINE))
    );
    // The personal file stays where it is; the receipt keeps a copy.
    assert_eq!(fx.read("AGENTS.local.md"), MINE);
    let id = applied["receipt"].as_str().unwrap();
    assert_eq!(
        fx.read(&format!(".moai-cli/receipts/{id}/source/AGENTS.local.md")),
        MINE
    );
    assert_eq!(
        fx.read(&format!(".moai-cli/receipts/{id}/before/CLAUDE.local.md")),
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

/// R10-9: a wrong digest, a missing permission reference and an `CLAUDE.local.md` that changed after
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

    fx.write(
        "CLAUDE.local.md",
        &format!("{RULES}\nAdded after the preview.\n"),
    );
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
    let merged = fx.read("CLAUDE.local.md");

    // The block text edited by hand.
    fx.write(
        "CLAUDE.local.md",
        &merged.replace("small commits", "big commits"),
    );
    let p = fx.plan();
    assert_eq!(p["action"], json!("conflict"), "{p}");
    let edited = tree_state(&fx.root);
    let v = fx.apply_with(p["plan_digest"].as_str().unwrap(), REF);
    assert_eq!(v["ok"], json!(false), "{v}");
    assert_eq!(tree_state(&fx.root), edited);

    // The personal file changed after the merge: the old block is kept.
    fx.write("CLAUDE.local.md", &merged);
    fx.write("AGENTS.local.md", &format!("{MINE}One more line.\n"));
    let p = fx.plan();
    assert_eq!(p["action"], json!("conflict"), "{p}");
    assert!(p["reason"].as_str().unwrap().contains("changed"), "{p}");
    assert_eq!(fx.read("CLAUDE.local.md"), merged);

    // A marker whose hash is not 64 hex digits is damaged, not trusted.
    fx.write("AGENTS.local.md", MINE);
    fx.write("CLAUDE.local.md", &merged.replace("sha256=", "sha256=zz"));
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
    assert_eq!(fx.read("CLAUDE.local.md"), RULES);
}

/// R10-11: the size limit, a link, a missing `CLAUDE.local.md` and a source that holds a MoAI marker.
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

    // A `CLAUDE.local.md` that is a link is never followed or written through.
    let fx = Fx::with(RULES, Some(MINE));
    fs::rename(fx.root.join("CLAUDE.local.md"), fx.root.join("real.md")).unwrap();
    symlink(fx.root.join("real.md"), fx.root.join("CLAUDE.local.md")).unwrap();
    let p = fx.plan();
    assert_eq!(p["action"], json!("blocked"), "{p}");

    // A source that holds a managed-section marker would be read as a second section.
    let marked = format!("{MINE}<!-- moai:managed:begin id=x -->\n");
    let fx = Fx::with(RULES, Some(&marked));
    assert_eq!(fx.plan()["action"], json!("blocked"));
}

/// R10-10: the revert puts the earlier `CLAUDE.local.md` back, only while it still holds what the apply
/// wrote; it never touches the personal file.
#[test]
fn revert_restores_agents_md_only_while_it_is_unchanged() {
    let fx = Fx::with(RULES, Some(MINE));
    let id = fx.merge()["receipt"].as_str().unwrap().to_string();
    let merged = fx.read("CLAUDE.local.md");

    let v = fx.revert_with(&id, "");
    assert_eq!(v["ok"], json!(false), "{v}");
    let v = fx.revert_with("agents-local-merge-../x", REF);
    assert_eq!(v["ok"], json!(false), "{v}");
    assert_eq!(fx.read("CLAUDE.local.md"), merged);

    // Edited after the merge: refused, nothing written.
    fx.write("CLAUDE.local.md", &format!("{merged}\nlater\n"));
    let v = fx.revert_with(&id, REF);
    assert_eq!(v["ok"], json!(false), "{v}");
    assert_eq!(fx.read("CLAUDE.local.md"), format!("{merged}\nlater\n"));

    fx.write("CLAUDE.local.md", &merged);
    let v = fx.revert_with(&id, REF);
    assert_eq!(v["result"]["state"], json!("reverted"), "{v}");
    assert_eq!(fx.read("CLAUDE.local.md"), RULES);
    assert_eq!(fx.read("AGENTS.local.md"), MINE);
    // A second revert is a no-op; the text can be merged again.
    let v = fx.revert_with(&id, REF);
    assert_eq!(v["result"]["state"], json!("unchanged"), "{v}");
    assert_eq!(fx.plan()["action"], json!("merge"));
}

/// R10-7: after `go-state-apply` moved the file into its dated backup, that copy is the source
/// (a new generated file in place is ignored). Not covered here: a backup whose apply was
/// reverted by `go-state-revert` (the code skips it; no test).
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

    // The install step wrote a new guidance file and a new generated personal file; there is
    // no `CLAUDE.local.md` yet, so the merge creates it.
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
    assert_eq!(fx.read("CLAUDE.local.md"), block(MINE));
    assert_eq!(
        fx.read("AGENTS.md"),
        RULES,
        "the guidance file is not touched"
    );
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
    let merged = fs::read_to_string(root.path().join("CLAUDE.local.md")).unwrap();
    assert_eq!(merged, block(MINE), "the missing personal file was created");

    // The guidance file that generation owns (`GUIDANCE`) is regenerated; the personal file
    // is not part of that and keeps its one merged block.
    const GUIDANCE: &str = "AGENTS.md";
    let guidance_before = fs::read_to_string(root.path().join(GUIDANCE)).unwrap();
    agents("ko");
    let guidance_after = fs::read_to_string(root.path().join(GUIDANCE)).unwrap();
    assert_ne!(
        guidance_after, guidance_before,
        "the managed section was regenerated"
    );
    let personal = fs::read_to_string(root.path().join("CLAUDE.local.md")).unwrap();
    assert_eq!(
        personal, merged,
        "regenerating does not touch the personal file"
    );
    assert_eq!(personal.matches("moai:merged-local sha256=").count(), 1);
    let again = merge(&["agents-merge-plan", p]);
    assert_eq!(again["result"]["action"], json!("unchanged"), "{again}");
}

impl Fx {
    /// Run the command under a given umask (`"000"`, `"022"`, `"077"`).
    fn run_umask(&self, umask: &str, args: &[&str]) -> Value {
        let out = Command::new("/bin/sh")
            .args(["-c", "umask $0; exec \"$@\"", umask, EXE])
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
        serde_json::from_slice(&out.stdout)
            .unwrap_or_else(|_| panic!("{args:?}: {}", String::from_utf8_lossy(&out.stderr)))
    }

    fn receipt_path(&self, id: &str) -> PathBuf {
        self.root.join(".moai-cli/receipts").join(id)
    }

    fn set_status(&self, id: &str, status: &str) {
        let path = self.receipt_path(id).join("receipt.json");
        let mut v: Value = serde_json::from_slice(&fs::read(&path).unwrap()).unwrap();
        v["status"] = json!(status);
        fs::write(path, serde_json::to_vec_pretty(&v).unwrap()).unwrap();
    }
}

fn mode_of(path: &std::path::Path) -> u32 {
    use std::os::unix::fs::PermissionsExt;
    fs::symlink_metadata(path).unwrap().permissions().mode() & 0o777
}

/// #2982 audit B2: the receipt holds a copy of a personal file that may hold a token, so every
/// folder it makes is 0700 and every file 0600, whatever the umask.
#[test]
fn receipt_folders_and_files_are_owner_only_under_any_umask() {
    for umask in ["000", "022", "077"] {
        let fx = Fx::with(RULES, Some(MINE));
        let plan = fx.plan();
        let digest = plan["plan_digest"].as_str().unwrap();
        let v = fx.run_umask(
            umask,
            &[
                "agents-merge-apply",
                fx.p(),
                "--plan-digest",
                digest,
                "--permission-reference",
                REF,
            ],
        );
        assert_eq!(v["ok"], json!(true), "umask {umask}: {v}");
        let id = v["result"]["receipt"].as_str().unwrap().to_string();
        let dir = fx.receipt_path(&id);
        for d in [
            dir.parent().unwrap().to_path_buf(),
            dir.clone(),
            dir.join("before"),
            dir.join("source"),
        ] {
            assert_eq!(mode_of(&d), 0o700, "umask {umask}: {}", d.display());
        }
        for f in [
            dir.join("receipt.json"),
            dir.join("before/CLAUDE.local.md"),
            dir.join("source/AGENTS.local.md"),
        ] {
            assert_eq!(mode_of(&f), 0o600, "umask {umask}: {}", f.display());
        }
        // The revert rewrites the receipt; it stays owner-only.
        let r = fx.run_umask(
            umask,
            &[
                "agents-merge-revert",
                fx.p(),
                "--receipt",
                &id,
                "--permission-reference",
                REF,
            ],
        );
        assert_eq!(
            r["result"]["state"],
            json!("reverted"),
            "umask {umask}: {r}"
        );
        assert_eq!(mode_of(&dir.join("receipt.json")), 0o600, "umask {umask}");
    }
}

/// #2982 audit B1: the merged `CLAUDE.local.md` may be 262144 bytes (the limit every later read
/// accepts) and not one byte more; at the limit the merge and its revert both work.
#[test]
fn the_merged_file_stops_at_the_receipt_limit() {
    const LIMIT: usize = 262_144;
    let grow = block(MINE).len() + 1; // the blank line the separator adds after "\n"
    let agents_of = |extra: usize| format!("{}\n", "a".repeat(LIMIT - grow - 1 + extra));

    let fx = Fx::with(&agents_of(0), Some(MINE));
    let applied = fx.merge();
    assert_eq!(fx.read("CLAUDE.local.md").len(), LIMIT);
    let id = applied["receipt"].as_str().unwrap().to_string();
    assert_eq!(fx.plan()["action"], json!("unchanged"));
    let r = fx.revert_with(&id, REF);
    assert_eq!(r["result"]["state"], json!("reverted"), "{r}");
    assert_eq!(fx.read("CLAUDE.local.md"), agents_of(0));

    let fx = Fx::with(&agents_of(1), Some(MINE));
    let p = fx.plan();
    assert_eq!(p["action"], json!("blocked"), "{p}");
    assert!(p["reason"].as_str().unwrap().contains("262145"), "{p}");
    let before = tree_state(&fx.root);
    let v = fx.apply_with(p["plan_digest"].as_str().unwrap(), REF);
    assert_eq!(v["ok"], json!(false), "{v}");
    assert_eq!(tree_state(&fx.root), before, "nothing is written");
}

/// #2982 audit N2: an apply killed before it finished leaves a `pending` receipt; the revert
/// undoes it when `CLAUDE.local.md` was written, calls it a no-op when it was not, and refuses when
/// the file is neither.
#[test]
fn a_pending_receipt_is_reverted_by_what_agents_md_holds() {
    // Written: CLAUDE.local.md holds the merged bytes.
    let fx = Fx::with(RULES, Some(MINE));
    let id = fx.merge()["receipt"].as_str().unwrap().to_string();
    fx.set_status(&id, "pending");
    let v = fx.revert_with(&id, REF);
    assert_eq!(v["result"]["state"], json!("reverted"), "{v}");
    assert_eq!(fx.read("CLAUDE.local.md"), RULES);

    // Not written: CLAUDE.local.md holds its earlier bytes; the receipt becomes `failed`.
    let fx = Fx::with(RULES, Some(MINE));
    let id = fx.merge()["receipt"].as_str().unwrap().to_string();
    fx.set_status(&id, "pending");
    fx.write("CLAUDE.local.md", RULES);
    let v = fx.revert_with(&id, REF);
    assert_eq!(v["result"]["state"], json!("unchanged"), "{v}");
    assert_eq!(fx.read("CLAUDE.local.md"), RULES);
    let again = fx.revert_with(&id, REF);
    assert_eq!(
        again["ok"],
        json!(false),
        "a failed apply wrote nothing: {again}"
    );

    // Neither: edited by hand since; nothing is written.
    let fx = Fx::with(RULES, Some(MINE));
    let id = fx.merge()["receipt"].as_str().unwrap().to_string();
    fx.set_status(&id, "pending");
    fx.write("CLAUDE.local.md", "# something else\n");
    let v = fx.revert_with(&id, REF);
    assert_eq!(v["ok"], json!(false), "{v}");
    assert_eq!(fx.read("CLAUDE.local.md"), "# something else\n");

    // A `failed` receipt is never undone, even when CLAUDE.local.md holds the merged bytes.
    let fx = Fx::with(RULES, Some(MINE));
    let id = fx.merge()["receipt"].as_str().unwrap().to_string();
    let merged = fx.read("CLAUDE.local.md");
    fx.set_status(&id, "failed");
    let v = fx.revert_with(&id, REF);
    assert_eq!(v["ok"], json!(false), "{v}");
    assert_eq!(fx.read("CLAUDE.local.md"), merged);
}

/// #2982 audit M16–M18: only a well-formed receipt id is looked up, and a link anywhere on the
/// way to the receipt is refused.
#[test]
fn revert_refuses_odd_ids_and_links() {
    let fx = Fx::with(RULES, Some(MINE));
    let id = fx.merge()["receipt"].as_str().unwrap().to_string();
    let merged = fx.read("CLAUDE.local.md");
    let long = format!("agents-local-merge-{}", "1".repeat(64));
    for bad in [
        "agents-local-merge-a/b",
        "agents-local-merge-a b",
        "agents-local-merge-a.b",
        "other-123",
        "",
        long.as_str(),
    ] {
        let v = fx.revert_with(bad, REF);
        assert_eq!(v["ok"], json!(false), "{bad:?}: {v}");
    }
    assert_eq!(fx.read("CLAUDE.local.md"), merged);

    // `before/` as a link to a real copy.
    let dir = fx.receipt_path(&id);
    fs::rename(dir.join("before"), fx.root.join("elsewhere")).unwrap();
    symlink(fx.root.join("elsewhere"), dir.join("before")).unwrap();
    let v = fx.revert_with(&id, REF);
    assert_eq!(v["ok"], json!(false), "{v}");
    assert_eq!(fx.read("CLAUDE.local.md"), merged);
    fs::remove_file(dir.join("before")).unwrap();
    fs::rename(fx.root.join("elsewhere"), dir.join("before")).unwrap();

    // The receipt folder itself as a link.
    let real = fx.root.join("moved-receipt");
    fs::rename(&dir, &real).unwrap();
    symlink(&real, &dir).unwrap();
    let v = fx.revert_with(&id, REF);
    assert_eq!(v["ok"], json!(false), "{v}");
    assert_eq!(fx.read("CLAUDE.local.md"), merged);
    fs::remove_file(&dir).unwrap();
    fs::rename(&real, &dir).unwrap();

    // The receipt's own copy of the earlier file was altered.
    fs::write(dir.join("before/CLAUDE.local.md"), "# altered\n").unwrap();
    let v = fx.revert_with(&id, REF);
    assert_eq!(v["ok"], json!(false), "{v}");
    assert_eq!(fx.read("CLAUDE.local.md"), merged);
}

/// #2982 audit M30: the plan digest binds the source bytes: a personal file changed after the
/// preview (same length, same name) makes the old digest stale.
#[test]
fn the_digest_binds_the_source_bytes() {
    let fx = Fx::with(RULES, Some(MINE));
    let p = fx.plan();
    let other = MINE.replace("Korean", "Korea!");
    assert_eq!(other.len(), MINE.len());
    fx.write("AGENTS.local.md", &other);
    let v = fx.apply_with(p["plan_digest"].as_str().unwrap(), REF);
    assert_eq!(v["ok"], json!(false), "{v}");
    assert!(v.to_string().contains("stale"), "{v}");
    assert_eq!(fx.read("CLAUDE.local.md"), RULES);
    assert_ne!(fx.plan()["plan_digest"], p["plan_digest"]);
}

/// #2982 audit N5: deleting only the start marker leaves the text in the file; a new merge
/// would double it, so the plan is a conflict.
#[test]
fn a_deleted_start_marker_is_a_conflict() {
    let fx = Fx::with(RULES, Some(MINE));
    fx.merge();
    let merged = fx.read("CLAUDE.local.md");
    let without_start: String = merged
        .lines()
        .filter(|l| !l.starts_with("<!-- moai:merged-local sha256="))
        .map(|l| format!("{l}\n"))
        .collect();
    fx.write("CLAUDE.local.md", &without_start);
    let p = fx.plan();
    assert_eq!(p["action"], json!("conflict"), "{p}");
}

/// #2982 audit O27/O09: when `CLAUDE.local.md` cannot be written, the apply fails, the receipt says
/// `failed` (it was written first as `pending`, with the earlier copy beside it), and
/// `CLAUDE.local.md` is untouched; a revert of that receipt writes nothing.
#[test]
fn a_failed_write_leaves_a_failed_receipt_and_changes_nothing() {
    use std::os::unix::fs::PermissionsExt;
    let fx = Fx::with(RULES, Some(MINE));
    fs::create_dir_all(fx.root.join(".moai-cli")).unwrap();
    let plan = fx.plan();
    let digest = plan["plan_digest"].as_str().unwrap().to_string();
    // A project folder nobody can add files to: CLAUDE.local.md cannot be replaced there.
    fs::set_permissions(&fx.root, fs::Permissions::from_mode(0o555)).unwrap();
    let probe = fs::write(fx.root.join("probe"), "x");
    let blocked = probe.is_err();
    if !blocked {
        // Running as a user who ignores modes (root): the case cannot be built here.
        fs::remove_file(fx.root.join("probe")).unwrap();
    }
    let v = fx.apply_with(&digest, REF);
    fs::set_permissions(&fx.root, fs::Permissions::from_mode(0o755)).unwrap();
    if !blocked {
        return;
    }
    assert_eq!(v["ok"], json!(false), "{v}");
    assert_eq!(fx.read("CLAUDE.local.md"), RULES);
    let receipts: Vec<_> = fs::read_dir(fx.root.join(".moai-cli/receipts"))
        .unwrap()
        .map(|e| e.unwrap().file_name().into_string().unwrap())
        .collect();
    assert_eq!(receipts.len(), 1, "{receipts:?}");
    let dir = fx.receipt_path(&receipts[0]);
    let r: Value = serde_json::from_slice(&fs::read(dir.join("receipt.json")).unwrap()).unwrap();
    assert_eq!(r["status"], json!("failed"), "{r}");
    assert_eq!(r["agents_after"], Value::Null, "{r}");
    assert_eq!(
        fs::read_to_string(dir.join("before/CLAUDE.local.md")).unwrap(),
        RULES
    );
    let v = fx.revert_with(&receipts[0], REF);
    assert_eq!(v["ok"], json!(false), "{v}");
    assert_eq!(fx.read("CLAUDE.local.md"), RULES);
    // The same text still merges once the folder is writable again.
    assert_eq!(fx.plan()["action"], json!("merge"));
}

/// #2982 audit O16: a receipt id is `agents-local-merge-` plus letters, digits and dashes, at
/// most 64 bytes in all; the edge of the length and a character just outside the set.
#[test]
fn a_receipt_id_is_checked_by_prefix_characters_and_length() {
    let fx = Fx::with(RULES, Some(MINE));
    let id = fx.merge()["receipt"].as_str().unwrap().to_string();
    let merged = fx.read("CLAUDE.local.md");
    let prefix = "agents-local-merge-";
    let at = |n: usize| format!("{prefix}{}", "1".repeat(n - prefix.len()));
    // 64 bytes passes the id check (and then finds no receipt); 65 is refused as an id.
    let v = fx.revert_with(&at(64), REF);
    assert!(v.to_string().contains("receipt not found"), "{v}");
    let v = fx.revert_with(&at(65), REF);
    assert!(v.to_string().contains("unknown receipt id"), "{v}");
    for bad in [
        format!("{id}_"),
        format!("{id}+"),
        format!("{id}é"),
        format!("{id}\n"),
    ] {
        let v = fx.revert_with(&bad, REF);
        assert!(v.to_string().contains("unknown receipt id"), "{bad:?}: {v}");
    }
    assert_eq!(fx.read("CLAUDE.local.md"), merged);
    let v = fx.revert_with(&id, REF);
    assert_eq!(v["result"]["state"], json!("reverted"), "{v}");
}

fn digest_of(text: &str) -> String {
    use sha2::{Digest, Sha256};
    format!("sha256:{:x}", Sha256::digest(text.as_bytes()))
}

/// SPEC REQ-2, REQ-4: a project without `CLAUDE.local.md` gets one that holds exactly the block;
/// its receipt keeps no copy of a file that did not exist, and the revert removes the file, but
/// only while it still holds the bytes the apply wrote.
#[test]
fn a_missing_personal_file_is_created_and_the_revert_removes_it() {
    let fx = Fx::fresh(MINE);
    let p = fx.plan();
    assert_eq!(p["action"], json!("merge"), "{p}");
    assert_eq!(p["target"]["path"], json!("CLAUDE.local.md"), "{p}");
    assert_eq!(p["target"]["exists"], json!(false), "{p}");
    assert_eq!(p["target"]["starts_at_line"], json!(1), "{p}");

    let applied = fx.merge();
    assert_eq!(applied["writes"], json!(["CLAUDE.local.md"]), "{applied}");
    assert_eq!(fx.read("CLAUDE.local.md"), block(MINE));
    assert_eq!(fx.read("AGENTS.local.md"), MINE, "the source stays");
    assert!(
        !fx.root.join("AGENTS.md").exists(),
        "no AGENTS.md is created"
    );
    let id = applied["receipt"].as_str().unwrap().to_string();
    assert!(
        !fx.receipt_path(&id).join("before/CLAUDE.local.md").exists(),
        "there is no earlier file to copy"
    );

    // Edited after the merge: the revert refuses and the file stays.
    let merged = fx.read("CLAUDE.local.md");
    fx.write("CLAUDE.local.md", &format!("{merged}\nmine too\n"));
    let v = fx.revert_with(&id, REF);
    assert_eq!(v["ok"], json!(false), "{v}");
    assert_eq!(fx.read("CLAUDE.local.md"), format!("{merged}\nmine too\n"));

    // Untouched: the revert removes the file the apply created.
    fx.write("CLAUDE.local.md", &merged);
    let v = fx.revert_with(&id, REF);
    assert_eq!(v["result"]["state"], json!("reverted"), "{v}");
    assert!(!fx.root.join("CLAUDE.local.md").exists(), "{v}");
    assert_eq!(fx.read("AGENTS.local.md"), MINE);
    // A second revert is a no-op and the text can be merged again.
    let v = fx.revert_with(&id, REF);
    assert_eq!(v["result"]["state"], json!("unchanged"), "{v}");
    assert_eq!(fx.plan()["action"], json!("merge"));
}

/// SPEC REQ-5: Claude Code reads `AGENTS.md` only while no `CLAUDE.md`, `.claude/CLAUDE.md` or
/// `CLAUDE.local.md` exists, so creating `CLAUDE.local.md` over an `AGENTS.md` hides that file.
/// The preview says so; it stays silent when nothing changes for the reader.
#[test]
fn the_plan_warns_when_agents_md_would_stop_being_read() {
    let warns = |p: &Value| {
        p["warnings"]
            .as_array()
            .map(|w| w.iter().any(|x| x.as_str().unwrap().contains("AGENTS.md")))
            .unwrap_or(false)
    };

    let fx = Fx::fresh(MINE);
    fx.write("AGENTS.md", RULES);
    let p = fx.plan();
    assert!(warns(&p), "{p}");
    assert_eq!(p["action"], json!("merge"));

    // A CLAUDE.md already wins over AGENTS.md: nothing is hidden by the new file.
    fx.write("CLAUDE.md", "# Team\n");
    assert!(!warns(&fx.plan()));
    fs::remove_file(fx.root.join("CLAUDE.md")).unwrap();
    fx.write(".claude/CLAUDE.md", "# Team\n");
    assert!(!warns(&fx.plan()));
    fs::remove_file(fx.root.join(".claude/CLAUDE.md")).unwrap();
    assert!(warns(&fx.plan()));

    // No AGENTS.md: nothing to hide.
    fs::remove_file(fx.root.join("AGENTS.md")).unwrap();
    assert!(!warns(&fx.plan()));

    // CLAUDE.local.md already exists: AGENTS.md is already not read, the merge changes nothing
    // for the reader.
    let fx = Fx::with(RULES, Some(MINE));
    fx.write("AGENTS.md", RULES);
    assert!(!warns(&fx.plan()));
}

/// SPEC REQ-4 for an older receipt: a merge applied before the personal file moved wrote into
/// `AGENTS.md`; its receipt has no `target_path`, and the revert still restores `AGENTS.md`.
#[test]
fn a_receipt_of_the_earlier_generation_still_reverts_agents_md() {
    let fx = Fx::fresh(MINE);
    let merged = format!("{RULES}\n{}", block(MINE));
    fx.write("AGENTS.md", &merged);
    let id = "agents-local-merge-1700000000000000000-0";
    let dir = fx.receipt_path(id);
    fs::create_dir_all(dir.join("before")).unwrap();
    fs::write(dir.join("before/AGENTS.md"), RULES).unwrap();
    let receipt = json!({
        "schema_version": 1, "id": id, "at": "2026-11-01T00:00:00Z",
        "plan_digest": "sha256:old", "permission_reference": REF, "status": "applied",
        "agents_before": digest_of(RULES), "agents_after": digest_of(&merged),
        "block_sha256": "x", "source_kind": "in_place",
        "source_path": "AGENTS.local.md", "source_sha256": digest_of(MINE),
    });
    fs::write(
        dir.join("receipt.json"),
        serde_json::to_vec_pretty(&receipt).unwrap(),
    )
    .unwrap();

    let v = fx.revert_with(id, REF);
    assert_eq!(v["result"]["state"], json!("reverted"), "{v}");
    assert_eq!(v["result"]["writes"], json!(["AGENTS.md"]), "{v}");
    assert_eq!(fx.read("AGENTS.md"), RULES);
    assert!(!fx.root.join("CLAUDE.local.md").exists());
}
