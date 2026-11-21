//! The text a person wrote in `AGENTS.md` moves into `CLAUDE.md` (SPEC-20261121-agents-md-migration):
//! preview, apply, revert, through the library API of the merge engine (no command exists yet).
//! Temporary projects only. Credential-shaped text is built at run time so no fixture here is a
//! secret.

#![cfg(unix)]

use moai_cli::agents_local_merge::{
    agents_md_state, apply_for, plan_for, revert_for, AgentsMdState, Kind,
};
use serde_json::{json, Value};
use sha2::{Digest, Sha256};
use std::{fs, os::unix::fs::symlink, path::PathBuf};
use tempfile::TempDir;

const REF: &str = "test-permission-r10";
const MOAI_BLOCK: &str = "<!-- moai:managed:begin id=moai-workflow -->\n## MoAI workflow (managed section)\n\nGenerated text.\n<!-- moai:managed:end id=moai-workflow -->\n";
const TEXT: &str = "# Team rules\n\nWrite tests first.\n\n## Mine\n\nPrefer small commits.\n";

struct Fx {
    _tmp: TempDir,
    root: PathBuf,
}

impl Fx {
    fn new() -> Fx {
        let tmp = TempDir::new().unwrap();
        let root = fs::canonicalize(tmp.path()).unwrap().join("project");
        fs::create_dir_all(&root).unwrap();
        Fx { _tmp: tmp, root }
    }

    /// A project whose `AGENTS.md` holds the person's text around MoAI's managed section.
    fn with_text() -> Fx {
        let fx = Fx::new();
        fx.write(
            "AGENTS.md",
            "# Team rules\n\nWrite tests first.\n\n<MOAI>\n## Mine\n\nPrefer small commits.\n"
                .replace("<MOAI>\n", &format!("{MOAI_BLOCK}\n"))
                .as_str(),
        );
        fx
    }

    fn write(&self, rel: &str, body: &str) {
        let p = self.root.join(rel);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        fs::write(p, body).unwrap();
    }

    fn read(&self, rel: &str) -> String {
        fs::read_to_string(self.root.join(rel)).unwrap()
    }

    fn plan(&self) -> Value {
        plan_for(Kind::Guidance, &self.root).unwrap()
    }

    /// Plan and apply; the apply must succeed. Returns the apply result.
    fn migrate(&self) -> Value {
        let p = self.plan();
        assert_eq!(p["action"], json!("merge"), "{p}");
        apply_for(
            Kind::Guidance,
            &self.root,
            p["plan_digest"].as_str().unwrap(),
            REF,
        )
        .unwrap()
    }

    fn revert(&self, id: &str) -> Result<Value, String> {
        revert_for(Kind::Guidance, &self.root, id, REF)
    }
}

fn block(text: &str) -> String {
    format!(
        "<!-- moai:merged-agents sha256={:x} -->\n{}\n<!-- /moai:merged-agents -->\n",
        Sha256::digest(text.as_bytes()),
        text.trim_end_matches('\n')
    )
}

/// SPEC REQ-1, REQ-3, AC-4: the text of the person, without MoAI's section, goes into a new
/// `CLAUDE.md` once; `AGENTS.md` keeps every byte; a second run is `unchanged`.
#[test]
fn the_persons_text_moves_into_a_new_claude_md_once() {
    let fx = Fx::with_text();
    let agents_before = fx.read("AGENTS.md");
    let p = fx.plan();
    assert_eq!(p["action"], json!("merge"), "{p}");
    assert_eq!(p["target"]["path"], json!("CLAUDE.md"), "{p}");
    assert_eq!(p["target"]["exists"], json!(false), "{p}");
    assert_eq!(p["target"]["starts_at_line"], json!(1), "{p}");
    assert_eq!(p["source"]["path"], json!("AGENTS.md"), "{p}");
    assert_eq!(p["body"], json!(TEXT), "{p}");
    assert!(
        p["commit_notice"].as_str().unwrap().contains("committed"),
        "{p}"
    );
    assert!(!p["body"].as_str().unwrap().contains("Generated text"));

    let applied = fx.migrate();
    assert_eq!(applied["state"], json!("applied"), "{applied}");
    assert_eq!(applied["writes"], json!(["CLAUDE.md"]), "{applied}");
    assert_eq!(fx.read("CLAUDE.md"), block(TEXT));
    assert_eq!(
        fx.read("AGENTS.md"),
        agents_before,
        "AGENTS.md is never written"
    );
    let id = applied["receipt"].as_str().unwrap();
    assert!(id.starts_with("agents-md-migrate-"), "{id}");
    assert_eq!(
        fx.read(&format!(".moai-cli/receipts/{id}/source/AGENTS.md")),
        agents_before,
        "the receipt keeps a copy of the source file"
    );

    let again = fx.plan();
    assert_eq!(again["action"], json!("unchanged"), "{again}");
    let digest = again["plan_digest"].as_str().unwrap();
    let v = apply_for(Kind::Guidance, &fx.root, digest, REF).unwrap();
    assert_eq!(v["state"], json!("unchanged"), "{v}");
    assert_eq!(fx.read("CLAUDE.md"), block(TEXT));
}

/// SPEC REQ-5: no MoAI section means the whole file is the person's text; a file that holds only
/// MoAI's section (with the heading MoAI wrote) has nothing to move.
#[test]
fn a_file_without_a_managed_section_is_all_the_persons_and_one_with_only_it_is_nothing() {
    let fx = Fx::new();
    fx.write("AGENTS.md", TEXT);
    let p = fx.plan();
    assert_eq!(p["action"], json!("merge"), "{p}");
    assert_eq!(p["body"], json!(TEXT), "{p}");

    let fx = Fx::new();
    fx.write("AGENTS.md", &format!("# my-project\n\n{MOAI_BLOCK}"));
    let p = fx.plan();
    assert_eq!(p["action"], json!("none"), "{p}");
    assert!(p["reason"].as_str().unwrap().contains("no text"), "{p}");

    let fx = Fx::new();
    let p = fx.plan();
    assert_eq!(p["action"], json!("none"), "no AGENTS.md at all: {p}");
}

/// SPEC REQ-2, REQ-4: an existing `CLAUDE.md` gets one block at its end and every earlier byte
/// stays; the revert puts the earlier bytes back, but not over a later edit.
#[test]
fn an_existing_claude_md_gets_a_block_and_the_revert_restores_it() {
    let fx = Fx::with_text();
    let mine = "# Team\n\nA rule the team already has here.\n";
    fx.write("CLAUDE.md", mine);
    let p = fx.plan();
    assert_eq!(p["target"]["exists"], json!(true), "{p}");
    assert_eq!(p["target"]["starts_at_line"], json!(5), "{p}");
    let applied = fx.migrate();
    let merged = fx.read("CLAUDE.md");
    assert_eq!(merged, format!("{mine}\n{}", block(TEXT)));
    let id = applied["receipt"].as_str().unwrap().to_string();

    fx.write("CLAUDE.md", &format!("{merged}\nlater\n"));
    assert!(fx.revert(&id).is_err(), "a later edit blocks the revert");
    assert_eq!(fx.read("CLAUDE.md"), format!("{merged}\nlater\n"));

    fx.write("CLAUDE.md", &merged);
    let v = fx.revert(&id).unwrap();
    assert_eq!(v["state"], json!("reverted"), "{v}");
    assert_eq!(fx.read("CLAUDE.md"), mine);
    assert_eq!(fx.revert(&id).unwrap()["state"], json!("unchanged"));
}

/// SPEC REQ-4: the revert removes a `CLAUDE.md` the apply created, and `AGENTS.md` is untouched.
#[test]
fn the_revert_removes_a_created_claude_md() {
    let fx = Fx::with_text();
    let agents = fx.read("AGENTS.md");
    let id = fx.migrate()["receipt"].as_str().unwrap().to_string();
    let v = fx.revert(&id).unwrap();
    assert_eq!(v["state"], json!("reverted"), "{v}");
    assert!(!fx.root.join("CLAUDE.md").exists());
    assert_eq!(fx.read("AGENTS.md"), agents);
    assert_eq!(
        fx.plan()["action"],
        json!("merge"),
        "it can be migrated again"
    );
}

/// SPEC REQ-6: a `CLAUDE.md` that imports `AGENTS.md` already carries the text.
#[test]
fn a_claude_md_that_imports_agents_md_is_left_alone() {
    let fx = Fx::with_text();
    let claude = "# Team\n\n@AGENTS.md\n";
    fx.write("CLAUDE.md", claude);
    let p = fx.plan();
    assert_eq!(p["action"], json!("unchanged"), "{p}");
    assert!(p["reason"].as_str().unwrap().contains("import"), "{p}");
    assert_eq!(fx.read("CLAUDE.md"), claude);
}

/// SPEC REQ-10, REQ-11: a credential-shaped line and a MoAI marker in the text block the plan; the
/// value is never printed.
#[test]
fn credentials_and_markers_block_the_plan() {
    let secret = format!("{}{}", "sk-", "q".repeat(32));
    let fx = Fx::new();
    fx.write("AGENTS.md", &format!("{TEXT}api key: {secret}\n"));
    let p = fx.plan();
    assert_eq!(p["action"], json!("blocked"), "{p}");
    assert!(!p.to_string().contains(&secret), "the value is never shown");
    let digest = p["plan_digest"].as_str().unwrap();
    assert!(apply_for(Kind::Guidance, &fx.root, digest, REF).is_err());
    assert!(!fx.root.join("CLAUDE.md").exists());

    // An end marker with no start is damaged markers, not text to copy.
    let fx = Fx::new();
    fx.write(
        "AGENTS.md",
        &format!("{TEXT}<!-- moai:managed:end id=moai-workflow -->\n"),
    );
    assert_eq!(fx.plan()["action"], json!("blocked"));
}

/// The block an earlier personal-file merge wrote into `AGENTS.md` is not the team's text.
#[test]
fn an_earlier_personal_block_is_left_out() {
    let fx = Fx::new();
    let personal = "<!-- moai:merged-local sha256=".to_string()
        + &"a".repeat(64)
        + " -->\nMy own note.\n<!-- /moai:merged-local -->\n";
    fx.write("AGENTS.md", &format!("{TEXT}\n{personal}"));
    let p = fx.plan();
    assert_eq!(p["action"], json!("merge"), "{p}");
    assert!(!p["body"].as_str().unwrap().contains("My own note"), "{p}");
    assert_eq!(p["body"], json!(TEXT), "{p}");
}

/// SPEC REQ-7, REQ-8: a link as the target blocks; a changed project makes the old digest stale.
#[test]
fn a_link_blocks_and_a_stale_digest_writes_nothing() {
    let fx = Fx::with_text();
    fx.write("real.md", "# Elsewhere\n");
    symlink(fx.root.join("real.md"), fx.root.join("CLAUDE.md")).unwrap();
    let p = fx.plan();
    assert_eq!(p["action"], json!("blocked"), "{p}");
    fs::remove_file(fx.root.join("CLAUDE.md")).unwrap();

    let digest = fx.plan()["plan_digest"].as_str().unwrap().to_string();
    fx.write("AGENTS.md", &format!("{TEXT}One more rule.\n"));
    let err = apply_for(Kind::Guidance, &fx.root, &digest, REF).unwrap_err();
    assert!(err.contains("stale"), "{err}");
    assert!(!fx.root.join("CLAUDE.md").exists());
    assert!(apply_for(Kind::Guidance, &fx.root, &digest, "").is_err());
}

/// The guard that keeps a first `CLAUDE.md` from hiding `AGENTS.md` asks what the file holds: no
/// file, only MoAI's own section, text of its own, or something that blocks (SPEC
/// SPEC-20261121-claude-md-generation, plan step 1).
#[test]
fn the_agents_md_state_tells_the_guard_what_the_file_holds() {
    let state = |fx: &Fx| agents_md_state(&fx.root);

    let fx = Fx::new();
    assert!(matches!(state(&fx), AgentsMdState::Absent));

    // MoAI's section alone, and with the one title line MoAI writes above it, is not the
    // person's text.
    fx.write("AGENTS.md", MOAI_BLOCK);
    assert!(matches!(state(&fx), AgentsMdState::OnlyMoai));
    fx.write("AGENTS.md", &format!("# my-project\n\n{MOAI_BLOCK}"));
    assert!(matches!(state(&fx), AgentsMdState::OnlyMoai));

    // Text of the person's own, with or without MoAI's section.
    fx.write("AGENTS.md", TEXT);
    assert!(matches!(state(&fx), AgentsMdState::HasText));
    fx.write("AGENTS.md", &format!("{TEXT}\n{MOAI_BLOCK}"));
    assert!(matches!(state(&fx), AgentsMdState::HasText));

    // A credential-shaped line, a stray marker and a file that is not text each block, and the
    // reason never carries the value.
    let secret = format!("{}{}", "sk-", "q".repeat(32));
    fx.write("AGENTS.md", &format!("{TEXT}token: {secret}\n"));
    match state(&fx) {
        AgentsMdState::Blocked(why) => assert!(!why.contains(&secret), "{why}"),
        other => panic!("expected Blocked, got {other:?}"),
    }
    fx.write(
        "AGENTS.md",
        &format!("{TEXT}<!-- moai:managed:end id=moai-workflow -->\n"),
    );
    assert!(matches!(state(&fx), AgentsMdState::Blocked(_)));
    fs::write(fx.root.join("AGENTS.md"), [0xff, 0xfe, 0x00, 0x41]).unwrap();
    assert!(matches!(state(&fx), AgentsMdState::Blocked(_)));
}

/// A link as `AGENTS.md` blocks, even a dangling one.
#[test]
fn a_linked_agents_md_blocks_the_state() {
    let fx = Fx::new();
    fx.write("real.md", TEXT);
    symlink(fx.root.join("real.md"), fx.root.join("AGENTS.md")).unwrap();
    assert!(matches!(
        agents_md_state(&fx.root),
        AgentsMdState::Blocked(_)
    ));
    fs::remove_file(fx.root.join("real.md")).unwrap();
    assert!(matches!(
        agents_md_state(&fx.root),
        AgentsMdState::Blocked(_)
    ));
}

/// The two kinds keep their receipts apart: a receipt id of one kind is unknown to the other.
#[test]
fn a_receipt_id_belongs_to_its_kind() {
    let fx = Fx::with_text();
    let id = fx.migrate()["receipt"].as_str().unwrap().to_string();
    let err = revert_for(Kind::Local, &fx.root, &id, REF).unwrap_err();
    assert!(err.contains("unknown receipt id"), "{err}");
    assert!(fx.root.join("CLAUDE.md").exists(), "nothing was reverted");
}

/// Audit of the slice, C-M2: the revert of this kind takes only the receipt that names
/// `CLAUDE.md`. A receipt that names another file, even the personal file the other kind writes,
/// is refused and the file keeps every byte, although it holds exactly what the apply wrote.
#[test]
fn a_guidance_receipt_that_names_another_file_is_refused() {
    for other in [
        "README.md",
        "CLAUDE.local.md",
        "AGENTS.local.md",
        "AGENTS.md",
    ] {
        let fx = Fx::with_text();
        let id = fx.migrate()["receipt"].as_str().unwrap().to_string();
        let written = fx.read("CLAUDE.md");
        let agents_before = fx.read("AGENTS.md");
        if other != "AGENTS.md" {
            fx.write(other, &written);
        }
        let path = fx
            .root
            .join(format!(".moai-cli/receipts/{id}/receipt.json"));
        let mut r: Value = serde_json::from_slice(&fs::read(&path).unwrap()).unwrap();
        r["target_path"] = json!(other);
        fs::write(&path, serde_json::to_vec_pretty(&r).unwrap()).unwrap();

        let err = fx.revert(&id).unwrap_err();
        assert!(err.contains("never writes"), "{other}: {err}");
        assert_eq!(fx.read("CLAUDE.md"), written, "{other}: CLAUDE.md changed");
        if other == "AGENTS.md" {
            assert_eq!(fx.read("AGENTS.md"), agents_before);
        } else {
            assert_eq!(fx.read(other), written, "{other} was changed");
        }
    }
}

/// Audit of the slice, C-M3: a marker line of the personal-file merge that stayed in the text
/// (one that lost its other half, so it was not taken out as an earlier block) blocks the plan:
/// copied into `CLAUDE.md`, it would be read by the personal-file merge as its own block.
#[test]
fn a_damaged_personal_marker_in_the_text_blocks_the_plan() {
    let begin = format!("<!-- moai:merged-local sha256={} -->\n", "b".repeat(64));
    let end = "<!-- /moai:merged-local -->\n";
    for (what, text) in [
        ("a start marker alone", format!("{TEXT}{begin}My note.\n")),
        ("an end marker alone", format!("{TEXT}My note.\n{end}")),
    ] {
        let fx = Fx::new();
        fx.write("AGENTS.md", &text);
        let p = fx.plan();
        assert_eq!(p["action"], json!("blocked"), "{what}: {p}");
        assert!(
            p["reason"].as_str().unwrap().contains("marker"),
            "{what}: {p}"
        );
        let digest = p["plan_digest"].as_str().unwrap();
        assert!(apply_for(Kind::Guidance, &fx.root, digest, REF).is_err());
        assert!(!fx.root.join("CLAUDE.md").exists(), "{what}");
    }
}
