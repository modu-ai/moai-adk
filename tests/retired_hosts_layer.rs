//! SPEC-20261121-retired-hosts-read-safely (REQ-1, REQ-2, REQ-10): the retired-host rule, the
//! profile list filter and the manifest entry check, called with an active list that leaves the
//! retired hosts out (the place a removal change switches the layer on) and with today's list.
use moai_cli::retired_hosts::{owned_entry_retired, retire_hosts, retired, RETIRED_HOSTS};

const TODAY: [&str; 3] = ["claude", "codex", "omp"];
const AFTER: [&str; 1] = ["claude"];

fn v(items: &[&str]) -> Vec<String> {
    items.iter().map(|s| s.to_string()).collect()
}

#[test]
fn the_retired_list_is_codex_and_omp() {
    assert_eq!(RETIRED_HOSTS, ["codex", "omp"]);
}

#[test]
fn a_host_is_retired_only_when_the_place_no_longer_lists_it() {
    // Active today: never retired, whatever the list.
    for host in TODAY {
        assert!(!retired(host, &TODAY), "{host}");
    }
    // After the removal the place lists claude only.
    assert!(retired("codex", &AFTER));
    assert!(retired("omp", &AFTER));
    assert!(!retired("claude", &AFTER));
    // A name that is not in the fixed list is never retired, listed or not.
    assert!(!retired("gemini", &AFTER));
    assert!(!retired("nope", &[]));
    // A place that still lists one of them keeps that one.
    assert!(retired("codex", &["claude", "omp"]));
    assert!(!retired("omp", &["claude", "omp"]));
}

#[test]
fn retire_hosts_removes_retired_names_and_falls_back_to_claude() {
    for (input, expected) in [
        (&["claude", "codex"][..], &["claude"][..]),
        (&["codex", "claude", "omp"][..], &["claude"][..]),
        (&["codex", "omp"][..], &["claude"][..]),
        (&["omp"][..], &["claude"][..]),
        (&["codex"][..], &["claude"][..]),
        (&["claude"][..], &["claude"][..]),
    ] {
        let mut hosts = v(input);
        retire_hosts(&mut hosts, &AFTER);
        assert_eq!(hosts, v(expected), "{input:?}");
    }
}

#[test]
fn retire_hosts_changes_nothing_while_the_hosts_are_active() {
    for input in [
        &["claude", "codex", "omp"][..],
        &["codex"][..],
        &["omp", "claude"][..],
        &[][..],
    ] {
        let mut hosts = v(input);
        retire_hosts(&mut hosts, &TODAY);
        assert_eq!(hosts, v(input), "{input:?}");
    }
}

#[test]
fn retire_hosts_keeps_an_empty_list_empty_and_unknown_names_as_they_are() {
    let mut empty: Vec<String> = Vec::new();
    retire_hosts(&mut empty, &AFTER);
    assert!(empty.is_empty());
    let mut hosts = v(&["gemini", "codex"]);
    retire_hosts(&mut hosts, &AFTER);
    assert_eq!(hosts, v(&["gemini"]));
}

#[test]
fn manifest_entries_of_a_retired_host_are_named_by_their_kind() {
    for kind in ["project-skill", "native-role", "skill-projection"] {
        let codex = format!("generation:{kind}:codex:moai");
        let omp = format!("generation:{kind}:omp:moai");
        let claude = format!("generation:{kind}:claude:moai");
        assert!(owned_entry_retired(&codex, &AFTER), "{codex}");
        assert!(owned_entry_retired(&omp, &AFTER), "{omp}");
        assert!(!owned_entry_retired(&claude, &AFTER), "{claude}");
        assert!(!owned_entry_retired(&codex, &TODAY), "{codex} while active");
    }
}

#[test]
fn manifest_entries_of_other_shapes_are_never_retired() {
    for id in [
        "",
        "generation:other:codex:moai",
        "generation:project-skill:codex",
        "agents:codex:x",
        "codex",
        "generation:project-skill:xcodex:moai",
    ] {
        assert!(!owned_entry_retired(id, &AFTER), "{id}");
    }
}

// ---- model settings (REQ-5, REQ-6) -------------------------------------------------------

mod model_settings {
    use moai_cli::model_effort_settings::{
        resolve_in, unsupported_entries_in, with_choice_in, Source, TOOLS,
    };
    use serde_json::{json, Map, Value};

    /// A secret-looking marker: it is put into every refused value, and no refusal may contain it.
    const SECRET: &str = "SECRET-saved-text";
    const AFTER: [&str; 2] = ["claude", "glm"];

    fn object(value: Value) -> Map<String, Value> {
        match value {
            Value::Object(map) => map,
            other => panic!("test fixture must be an object, got {other}"),
        }
    }

    /// A launcher object saved by an earlier version: a valid Claude entry, a valid old Codex
    /// entry, an OMP entry and a sibling key.
    fn old_launcher() -> Map<String, Value> {
        object(json!({
            "advisor": "keep-me",
            "models": {
                "claude": {"leader": {"model": "claude-sonnet-5-5", "effort": "low"}},
                "codex": {"leader": {"model": "gpt-6.2-sol", "effort": "high"}},
                "omp": {"agent": {"model": "zai/glm-5.3-flash", "effort": "max"}}
            }
        }))
    }

    #[test]
    fn retired_tool_keys_of_any_shape_are_skipped_on_read() {
        for shape in [
            json!({"leader": {"model": "gpt-6.2-sol", "effort": "high"}}),
            json!({"leader": {"model": SECRET, "effort": SECRET}}),
            json!({"no-such-role": 3}),
            json!(SECRET),
            json!([SECRET]),
            json!(null),
        ] {
            let launcher = object(json!({
                "models": {
                    "claude": {"leader": {"model": "claude-sonnet-5-5", "effort": "low"}},
                    "codex": shape.clone(),
                    "omp": shape.clone()
                }
            }));
            let listed = unsupported_entries_in(&AFTER, &launcher)
                .unwrap_or_else(|e| panic!("{shape}: refused: {e}"));
            assert!(listed.is_empty(), "{shape}: {listed:?}");
            let (choice, source) = resolve_in(&AFTER, "claude", "leader", Some(&launcher))
                .unwrap_or_else(|e| panic!("{shape}: refused: {e}"));
            assert_eq!(source, Source::Saved, "{shape}");
            assert_eq!(choice.model, "claude-sonnet-5-5", "{shape}");
        }
    }

    #[test]
    fn an_unknown_tool_name_is_still_refused_without_its_text() {
        let launcher = object(json!({
            "models": {
                "claude": {"leader": {"model": "claude-sonnet-5-5", "effort": "low"}},
                "codex": {},
                SECRET: {}
            }
        }));
        let err = resolve_in(&AFTER, "claude", "leader", Some(&launcher)).unwrap_err();
        assert!(!err.contains(SECRET), "{err}");
        let err = unsupported_entries_in(&AFTER, &launcher).unwrap_err();
        assert!(!err.contains(SECRET), "{err}");
    }

    #[test]
    fn while_the_tools_are_active_a_malformed_value_is_refused_as_before() {
        // `glm` is still one of the launcher tools, so a malformed value under it is refused
        // without its text; `codex` and `omp` left the tool list with the launcher changes and
        // are skipped.
        let launcher = object(json!({"models": {"glm": SECRET}}));
        let err = resolve_in(&TOOLS, "claude", "leader", Some(&launcher)).unwrap_err();
        assert!(!err.contains(SECRET), "{err}");
        let err = unsupported_entries_in(&TOOLS, &launcher).unwrap_err();
        assert!(!err.contains(SECRET), "{err}");
        for host in ["codex", "omp"] {
            let retired = object(json!({"models": {host: SECRET}}));
            assert!(
                resolve_in(&TOOLS, "claude", "leader", Some(&retired)).is_ok(),
                "{host}"
            );
            assert!(
                unsupported_entries_in(&TOOLS, &retired).unwrap().is_empty(),
                "{host}"
            );
        }
    }

    #[test]
    fn saving_a_choice_leaves_retired_tool_keys_out_and_keeps_the_rest() {
        let launcher = old_launcher();
        let saved = with_choice_in(
            &AFTER,
            &launcher,
            "claude",
            "agent",
            "claude-haiku-5-5",
            "high",
        )
        .unwrap();
        assert_eq!(saved["advisor"], json!("keep-me"));
        assert_eq!(
            saved["models"],
            json!({
                "claude": {
                    "leader": {"model": "claude-sonnet-5-5", "effort": "low"},
                    "agent": {"model": "claude-haiku-5-5", "effort": "high"}
                }
            })
        );
        // The input is not changed.
        assert_eq!(launcher, old_launcher());
    }

    #[test]
    fn saving_a_choice_while_the_tools_are_active_keeps_every_active_tool_key() {
        let launcher = old_launcher();
        let saved = with_choice_in(
            &TOOLS,
            &launcher,
            "claude",
            "agent",
            "claude-haiku-5-5",
            "high",
        )
        .unwrap();
        // The Codex and OMP rows left the model table with the launcher changes: their saved
        // keys are dropped on this save, the key of a tool that is still active (Claude) is kept.
        assert!(saved["models"].get("codex").is_none());
        assert!(saved["models"].get("omp").is_none());
        assert_eq!(
            saved["models"]["claude"]["leader"],
            launcher["models"]["claude"]["leader"]
        );
    }
}

// ---- receipts of a retired host (REQ-7) ----------------------------------------------------

mod receipts {
    use moai_cli::{agent_install, home_routing, home_style, update_hooks};
    use serde_json::{json, Value};
    use std::collections::BTreeSet;
    use std::fs;
    use std::path::{Path, PathBuf};

    const PERM: &str = "test-permission";
    const AFTER: [&str; 1] = ["claude"];

    fn package() -> PathBuf {
        Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai")
    }

    fn agents() -> PathBuf {
        Path::new(env!("CARGO_MANIFEST_DIR")).join("agents")
    }

    fn home() -> (tempfile::TempDir, PathBuf) {
        let dir = tempfile::tempdir().unwrap();
        let path = fs::canonicalize(dir.path()).unwrap();
        (dir, path)
    }

    fn tree(root: &Path) -> BTreeSet<(String, Vec<u8>)> {
        fn walk(dir: &Path, root: &Path, out: &mut BTreeSet<(String, Vec<u8>)>) {
            let Ok(entries) = fs::read_dir(dir) else {
                return;
            };
            for e in entries {
                let p = e.unwrap().path();
                if p.is_dir() {
                    walk(&p, root, out);
                } else {
                    let rel = p.strip_prefix(root).unwrap().to_string_lossy().to_string();
                    out.insert((rel, fs::read(&p).unwrap()));
                }
            }
        }
        let mut out = BTreeSet::new();
        walk(root, root, &mut out);
        out
    }

    /// What an earlier version left in a home: files of the retired hosts the receipts name.
    fn leave_retired_files(home: &Path) {
        for (rel, text) in [
            (".codex/AGENTS.md", "# kept\n"),
            (".codex/agents/mine.md", "kept\n"),
            (".omp/agent/agents/mine.md", "kept\n"),
        ] {
            let path = home.join(rel);
            fs::create_dir_all(path.parent().unwrap()).unwrap();
            fs::write(path, text).unwrap();
        }
    }

    /// Adds an entry for `host` to receipt `id` of the part whose state folder is `part`: a copy
    /// of its first entry with the host changed, the way an earlier version wrote it (the
    /// apply of this version writes Claude entries only).
    fn add_entry(home: &Path, part: &str, id: &str, host: &str, kind: Option<&str>) {
        let path = moai_cli::home_root::path(home)
            .join(part)
            .join("receipts")
            .join(id)
            .join("receipt.json");
        let mut receipt: Value = serde_json::from_slice(&fs::read(&path).unwrap()).unwrap();
        let mut entry = receipt["entries"][0].clone();
        entry["host"] = json!(host);
        if let Some(kind) = kind {
            entry["kind"] = json!(kind);
        }
        receipt["entries"].as_array_mut().unwrap().push(entry);
        fs::write(&path, receipt.to_string()).unwrap();
    }

    #[test]
    fn a_routing_receipt_skips_the_retired_host_and_reverts_the_other() {
        let (_t, home) = home();
        fs::create_dir_all(home.join(".claude")).unwrap();
        fs::write(home.join(".claude/CLAUDE.md"), "mine\n").unwrap();
        leave_retired_files(&home);
        let p = home_routing::plan(&home, &package()).unwrap();
        let r = home_routing::apply(&home, &package(), &p.plan_digest, PERM).unwrap();
        let id = r["receipt"].as_str().unwrap().to_string();
        add_entry(&home, "home-routing", &id, "codex", None);
        let codex_before = fs::read(home.join(".codex/AGENTS.md")).unwrap();
        let v = home_routing::revert_in(&AFTER, &home, &id, PERM).unwrap();
        assert_eq!(v["state"], "reverted", "{v}");
        assert_eq!(v["reverted"], json!(["claude"]), "{v}");
        assert_eq!(
            v["skipped"],
            json!([{"host": "codex", "reason": "host_retired"}]),
            "{v}"
        );
        assert_eq!(
            fs::read_to_string(home.join(".claude/CLAUDE.md")).unwrap(),
            "mine\n"
        );
        assert_eq!(
            fs::read(home.join(".codex/AGENTS.md")).unwrap(),
            codex_before,
            "the retired host's file is left alone"
        );
    }

    #[test]
    fn a_style_receipt_skips_the_retired_host_and_reverts_the_other() {
        let (_t, home) = home();
        leave_retired_files(&home);
        let p = home_style::plan(&home, &package()).unwrap();
        let r = home_style::apply(&home, &package(), &p.plan_digest, PERM).unwrap();
        let id = r["receipt"].as_str().unwrap().to_string();
        add_entry(&home, "home-style", &id, "codex", Some("block"));
        let codex_before = fs::read(home.join(".codex/AGENTS.md")).unwrap();
        let v = home_style::revert_in(&AFTER, &home, &id, PERM).unwrap();
        assert_eq!(v["state"], "reverted", "{v}");
        assert_eq!(
            v["skipped"],
            json!([{"target": "codex:block", "reason": "host_retired"}]),
            "{v}"
        );
        let reverted: Vec<String> = serde_json::from_value(v["reverted"].clone()).unwrap();
        assert!(!reverted.is_empty(), "{v}");
        assert!(reverted.iter().all(|t| t.starts_with("claude:")), "{v}");
        assert_eq!(
            fs::read(home.join(".codex/AGENTS.md")).unwrap(),
            codex_before
        );
    }

    #[test]
    fn an_agent_receipt_skips_the_retired_hosts_and_reverts_claude() {
        let (_t, home) = home();
        leave_retired_files(&home);
        let p = agent_install::plan(&home, &package(), &agents()).unwrap();
        let r = agent_install::apply(&home, &package(), &agents(), &p.plan_digest, PERM).unwrap();
        let id = r["receipt"].as_str().unwrap().to_string();
        add_entry(&home, "agents", &id, "codex", None);
        add_entry(&home, "agents", &id, "omp", None);
        let codex_before = tree(&home.join(".codex"));
        let omp_before = tree(&home.join(".omp"));
        assert!(!codex_before.is_empty() && !omp_before.is_empty());
        let v = agent_install::revert_in(&AFTER, &home, &id, PERM).unwrap();
        assert_eq!(v["state"], "reverted", "{v}");
        let reverted: Vec<String> = serde_json::from_value(v["reverted"].clone()).unwrap();
        assert!(!reverted.is_empty(), "{v}");
        assert!(reverted.iter().all(|t| t.starts_with("claude:")), "{v}");
        let skipped = v["skipped"].as_array().unwrap();
        assert!(!skipped.is_empty(), "{v}");
        assert!(skipped.iter().all(|s| s["reason"] == "host_retired"), "{v}");
        let hosts: BTreeSet<&str> = skipped
            .iter()
            .map(|s| s["host"].as_str().unwrap())
            .collect();
        assert_eq!(hosts, BTreeSet::from(["codex", "omp"]), "{v}");
        assert_eq!(tree(&home.join(".codex")), codex_before);
        assert_eq!(tree(&home.join(".omp")), omp_before);
    }

    /// A hand-written hook receipt: absent before and after for every file of the hosts, so a
    /// host the revert handles has nothing to restore.
    fn hook_receipt(project: &Path, hosts: &[&str]) -> String {
        let id = "hooks-retired-1";
        let entries: Vec<Value> = hosts
            .iter()
            .map(|h| {
                let (files, _) = moai_cli::hooks::written_files(project, h).unwrap();
                let files: Vec<Value> = files
                    .iter()
                    .map(|f| json!({"path": f, "before_sha256": null, "after_sha256": null}))
                    .collect();
                json!({"host": h, "files": files})
            })
            .collect();
        let dir = moai_cli::state_root::path(project, "receipts").join(id);
        fs::create_dir_all(&dir).unwrap();
        let receipt = json!({
            "schema_version": 1, "id": id, "plan_digest": "x",
            "permission_reference": "x", "status": "committed", "hosts": entries
        });
        fs::write(dir.join("receipt.json"), receipt.to_string()).unwrap();
        id.to_string()
    }

    #[test]
    fn a_hook_receipt_skips_the_retired_hosts_and_reverts_the_other() {
        let tmp = tempfile::tempdir().unwrap();
        let project = fs::canonicalize(tmp.path()).unwrap();
        let id = hook_receipt(&project, &["claude", "codex", "omp"]);
        let v = update_hooks::revert_in(&AFTER, &project, &id, PERM).unwrap();
        assert_eq!(v["state"], "reverted", "{v}");
        assert_eq!(v["reverted"], json!(["claude"]), "{v}");
        assert_eq!(
            v["skipped"],
            json!([
                {"host": "codex", "reason": "host_retired"},
                {"host": "omp", "reason": "host_retired"}
            ]),
            "{v}"
        );
    }

    #[test]
    fn a_hook_receipt_of_retired_hosts_only_is_reverted_not_left_unchanged() {
        let tmp = tempfile::tempdir().unwrap();
        let project = fs::canonicalize(tmp.path()).unwrap();
        let id = hook_receipt(&project, &["codex"]);
        let v = update_hooks::revert_in(&AFTER, &project, &id, PERM).unwrap();
        assert_eq!(v["state"], "reverted", "{v}");
        assert_eq!(v["reverted"], json!([]), "{v}");
        assert_eq!(
            v["skipped"],
            json!([{"host": "codex", "reason": "host_retired"}]),
            "{v}"
        );
        // A second revert says so, like any receipt.
        assert!(update_hooks::revert_in(&AFTER, &project, &id, PERM)
            .unwrap_err()
            .contains("already"));
    }

    #[test]
    fn a_hook_receipt_naming_an_unknown_host_is_still_refused() {
        let tmp = tempfile::tempdir().unwrap();
        let project = fs::canonicalize(tmp.path()).unwrap();
        let id = hook_receipt(&project, &["claude"]);
        let path = moai_cli::state_root::path(&project, "receipts")
            .join(&id)
            .join("receipt.json");
        let text = fs::read_to_string(&path).unwrap();
        fs::write(
            &path,
            text.replace("\"host\":\"claude\"", "\"host\":\"gemini\""),
        )
        .unwrap();
        let err = update_hooks::revert_in(&AFTER, &project, &id, PERM).unwrap_err();
        assert!(err.contains("unknown host"), "{err}");
    }
}

// ---- generation manifest entries of a retired host (REQ-8) -----------------------------------

mod generation_manifest {
    use moai_cli::generation::{
        generation_apply, generation_apply_in, generation_plan, generation_plan_in, GenerationPlan,
        GenerationRequest,
    };
    use serde_json::{json, Value};
    use std::fs;
    use std::path::Path;

    const TODAY: [&str; 2] = ["claude", "codex"];
    const AFTER: [&str; 1] = ["claude"];
    const PERM: &str = "fixture-approved";
    const CLAUDE_CHECKS: &str = ".claude/skills/prj-checks/SKILL.md";
    const CODEX_CHECKS: &str = ".agents/skills/prj-checks/SKILL.md";
    const CLAUDE_EXTRA: &str = ".claude/skills/prj-extra/SKILL.md";
    const CODEX_EXTRA: &str = ".agents/skills/prj-extra/SKILL.md";
    const PROJECTION: &str = ".agents/skills/prj-proj/SKILL.md";
    const MANIFEST: &str = ".moai-cli/manifests/generation.json";

    fn project() -> tempfile::TempDir {
        let root = tempfile::tempdir().unwrap();
        fs::create_dir_all(root.path().join("skills/moai/workflows")).unwrap();
        fs::create_dir_all(root.path().join(".moai-cli")).unwrap();
        let decision = |id: &str, value: &str, source: &str| json!({"id": id, "topic": id, "value": value, "source": source, "scope": "project"});
        let profile = json!({"schema_version": 2, "project_id": "demo",
        "purpose": "Build a small editor", "tools": ["Rust"],
        "development_methods": ["TDD"], "hosts": ["claude", "codex"],
        "interview_revision": "r1", "decisions": [
            decision("test_command", "cargo test --locked", "user"),
            decision("review_blocks", "a change without a test", "user"),
            decision("protected", "the billing module", "user"),
            decision("build_command", "HOLD", "default"),
        ]});
        let mut bytes = serde_json::to_vec_pretty(&profile).unwrap();
        bytes.push(b'\n');
        fs::write(root.path().join(".moai-cli/project.yaml"), bytes).unwrap();
        root
    }

    fn request(skills: Value, roles: Value) -> GenerationRequest {
        serde_json::from_value(json!({
            "protocol_version": 1, "language": "en",
            "workflow_reference": "skills/moai/workflows", "local_file": "none",
            "workflow_id": "review-flow", "native_roles": roles, "skill_projections": [],
            "project_skills": skills
        }))
        .unwrap()
    }

    fn both_skills() -> Value {
        json!([
            {"id": "prj-checks", "purpose": "How this project is checked"},
            {"id": "prj-extra", "purpose": "Something extra"}
        ])
    }

    fn codex_role() -> Value {
        json!([{"role_id": "prj-reviewer", "host": "codex", "purpose": "Reviews changes"}])
    }

    fn manifest(root: &Path) -> Value {
        serde_json::from_slice(&fs::read(root.join(MANIFEST)).unwrap()).unwrap()
    }

    fn section_ids(root: &Path) -> Vec<String> {
        manifest(root)["owned_files"]
            .as_array()
            .unwrap()
            .iter()
            .filter_map(|f| f["section_id"].as_str().map(str::to_string))
            .collect()
    }

    /// A project written by an earlier version: two project skills for both hosts, a Codex
    /// native role, and a skill projection entry recorded for Codex.
    fn applied() -> tempfile::TempDir {
        let root = project();
        let r = root.path();
        let plan = generation_plan(r, request(both_skills(), codex_role())).unwrap();
        generation_apply(r, plan.clone(), &plan.plan_digest, PERM).unwrap();
        let mut value = manifest(r);
        let baseline = moai_cli::digest(b"projection");
        fs::create_dir_all(r.join(".agents/skills/prj-proj")).unwrap();
        fs::write(r.join(PROJECTION), b"projection").unwrap();
        value["owned_files"].as_array_mut().unwrap().push(json!({
            "path": PROJECTION, "kind": "skill-projection",
            "section_id": "generation:skill-projection:codex:prj-proj",
            "baseline_sha256": baseline
        }));
        let mut bytes = serde_json::to_vec_pretty(&value).unwrap();
        bytes.push(b'\n');
        fs::write(r.join(MANIFEST), bytes).unwrap();
        root
    }

    fn checks_for_claude_only() -> GenerationRequest {
        request(
            json!([{"id": "prj-checks", "purpose": "How this project is checked",
                "hosts": ["claude"]}]),
            json!([]),
        )
    }

    fn action(plan: &GenerationPlan, path: &str) -> Option<String> {
        plan.files
            .iter()
            .find(|f| f.path == path)
            .map(|f| f.action.clone())
    }

    #[test]
    fn while_codex_is_active_its_recorded_skills_are_planned_for_removal_as_before() {
        let root = applied();
        let plan = generation_plan_in(&TODAY, root.path(), checks_for_claude_only()).unwrap();
        assert_eq!(action(&plan, CODEX_CHECKS).as_deref(), Some("remove"));
        assert_eq!(action(&plan, CODEX_EXTRA).as_deref(), Some("remove"));
        assert_eq!(action(&plan, CLAUDE_EXTRA).as_deref(), Some("remove"));
    }

    #[test]
    fn a_retired_hosts_recorded_skills_plan_no_conflict_and_no_removal() {
        let root = applied();
        let plan = generation_plan_in(&AFTER, root.path(), checks_for_claude_only()).unwrap();
        assert_eq!(action(&plan, CODEX_CHECKS), None, "{:?}", plan.files);
        assert_eq!(action(&plan, CODEX_EXTRA), None, "{:?}", plan.files);
        assert_eq!(action(&plan, CLAUDE_EXTRA).as_deref(), Some("remove"));
        assert!(
            plan.files.iter().all(|f| f.action != "conflict"),
            "{:?}",
            plan.files
        );
    }

    #[test]
    fn an_apply_drops_the_retired_entries_from_the_manifest_and_leaves_their_files() {
        let root = applied();
        let r = root.path();
        let ids_before = section_ids(r);
        for kind in ["project-skill", "native-role", "skill-projection"] {
            assert!(
                ids_before
                    .iter()
                    .any(|id| id.starts_with(&format!("generation:{kind}:codex:"))),
                "{kind}: {ids_before:?}"
            );
        }
        let plan = generation_plan_in(&AFTER, r, checks_for_claude_only()).unwrap();
        generation_apply_in(&AFTER, r, plan.clone(), &plan.plan_digest, PERM).unwrap();
        let ids = section_ids(r);
        assert!(
            ids.iter().all(|id| !id.contains(":codex:")),
            "no entry of the retired host stays: {ids:?}"
        );
        assert!(
            ids.iter()
                .any(|id| id == "generation:project-skill:claude:prj-checks"),
            "{ids:?}"
        );
        for kept in [CODEX_CHECKS, CODEX_EXTRA, PROJECTION] {
            assert!(r.join(kept).exists(), "{kept} is left alone");
        }
        assert!(r.join(CLAUDE_CHECKS).exists());
        assert!(
            !r.join(CLAUDE_EXTRA).exists(),
            "an active host still removes"
        );
    }

    #[test]
    fn while_codex_is_active_an_apply_keeps_the_native_role_and_projection_entries() {
        let root = applied();
        let r = root.path();
        let plan = generation_plan_in(&TODAY, r, checks_for_claude_only()).unwrap();
        generation_apply_in(&TODAY, r, plan.clone(), &plan.plan_digest, PERM).unwrap();
        let ids = section_ids(r);
        assert!(
            ids.iter()
                .any(|id| id == "generation:native-role:codex:prj-reviewer"),
            "{ids:?}"
        );
        assert!(
            ids.iter()
                .any(|id| id == "generation:skill-projection:codex:prj-proj"),
            "{ids:?}"
        );
    }
}

// ---- profile hosts (REQ-2, REQ-3, REQ-4) ---------------------------------------------------

mod profile_hosts {
    use moai_cli::profile::{apply, apply_in, plan, plan_in, read_profile_in, Profile};
    use serde_json::{json, Value};
    use std::fs;
    use std::path::Path;

    const TODAY: [&str; 3] = ["claude", "codex", "omp"];
    const AFTER: [&str; 1] = ["claude"];
    const PERM: &str = "fixture-approved";

    fn profile(hosts: Value) -> Profile {
        serde_json::from_value(json!({
            "schema_version": 2, "project_id": "demo", "purpose": "Build a small editor",
            "tools": ["Rust"], "development_methods": ["TDD"], "hosts": hosts,
            "interview_revision": "r1"
        }))
        .unwrap()
    }

    fn file(root: &Path) -> std::path::PathBuf {
        moai_cli::state_root::path(root, "project.yaml")
    }

    fn baseline(root: &Path) -> std::path::PathBuf {
        moai_cli::state_root::path(root, "manifests/profile-baseline.json")
    }

    fn read_json(path: &Path) -> Value {
        serde_json::from_slice(&fs::read(path).unwrap()).unwrap()
    }

    /// A project whose profile an earlier version wrote with `claude` and `codex`, applied
    /// through the real engine so a baseline equal to the file exists.
    fn applied() -> tempfile::TempDir {
        let tmp = tempfile::tempdir().unwrap();
        let root = fs::canonicalize(tmp.path()).unwrap();
        let p = plan(&root, profile(json!(["claude", "codex"]))).unwrap();
        apply(&root, p.clone(), &p.plan_digest, PERM).unwrap();
        tmp
    }

    #[test]
    fn a_read_removes_retired_hosts_only_when_the_place_no_longer_lists_them() {
        let tmp = applied();
        let root = fs::canonicalize(tmp.path()).unwrap();
        let (kept, _) = read_profile_in(&TODAY, &root).unwrap();
        assert_eq!(kept.hosts, ["claude", "codex"]);
        let (read, _) = read_profile_in(&AFTER, &root).unwrap();
        assert_eq!(read.hosts, ["claude"]);
        // A file that only names retired hosts reads as claude.
        let mut value = read_json(&file(&root));
        value["hosts"] = json!(["omp", "codex"]);
        fs::write(file(&root), value.to_string()).unwrap();
        let (read, _) = read_profile_in(&AFTER, &root).unwrap();
        assert_eq!(read.hosts, ["claude"]);
    }

    #[test]
    fn a_plan_input_naming_a_retired_host_is_filtered_not_refused() {
        let tmp = tempfile::tempdir().unwrap();
        let root = fs::canonicalize(tmp.path()).unwrap();
        let planned = plan_in(&AFTER, &root, profile(json!(["codex", "omp"]))).unwrap();
        assert_eq!(planned.profile.hosts, ["claude"]);
        let kept = plan_in(&TODAY, &root, profile(json!(["codex"]))).unwrap();
        assert_eq!(kept.profile.hosts, ["codex"]);
        assert!(plan_in(&AFTER, &root, profile(json!(["gemini"]))).is_err());
    }

    /// The saved file keeps `codex`; the filtered desired profile is an update, never a conflict,
    /// whatever the baseline holds, and applying it writes `hosts` and nothing else.
    fn assert_update_to_claude(root: &Path, case: &str) {
        let before = read_json(&file(root));
        assert_eq!(before["hosts"], json!(["claude", "codex"]), "{case}");
        let (desired, _) = read_profile_in(&AFTER, root).unwrap();
        let planned = plan_in(&AFTER, root, desired).unwrap();
        assert_eq!(planned.action, "update", "{case}: {:?}", planned.delta);
        assert!(planned.delta.conflict.is_empty(), "{case}");
        assert_eq!(planned.delta.changed, ["hosts"], "{case}");
        apply_in(&AFTER, root, planned.clone(), &planned.plan_digest, PERM)
            .unwrap_or_else(|e| panic!("{case}: {e}"));
        let after = read_json(&file(root));
        let mut expected = before.clone();
        expected["hosts"] = json!(["claude"]);
        assert_eq!(after, expected, "{case}: only hosts changes");
    }

    #[test]
    fn a_saved_retired_host_is_updated_when_the_baseline_equals_the_file() {
        let tmp = applied();
        assert_update_to_claude(&fs::canonicalize(tmp.path()).unwrap(), "baseline == file");
    }

    #[test]
    fn a_saved_retired_host_is_updated_when_there_is_no_baseline() {
        let tmp = applied();
        let root = fs::canonicalize(tmp.path()).unwrap();
        fs::remove_file(baseline(&root)).unwrap();
        assert_update_to_claude(&root, "no baseline");
    }

    #[test]
    fn a_saved_retired_host_is_updated_when_the_baseline_was_edited_by_hand() {
        let tmp = applied();
        let root = fs::canonicalize(tmp.path()).unwrap();
        let mut record = read_json(&baseline(&root));
        record["profile"]["hosts"] = json!(["claude"]);
        fs::write(baseline(&root), record.to_string()).unwrap();
        assert_update_to_claude(&root, "baseline hand-edited");
    }

    #[test]
    fn while_the_hosts_are_active_the_saved_profile_plans_unchanged() {
        let tmp = applied();
        let root = fs::canonicalize(tmp.path()).unwrap();
        let (desired, _) = read_profile_in(&TODAY, &root).unwrap();
        let planned = plan_in(&TODAY, &root, desired).unwrap();
        assert_eq!(planned.action, "unchanged", "{:?}", planned.delta);
    }
}

// ---- the shared backup code leaves the Codex module (REQ-9) --------------------------------

mod backup_move {
    use std::fs;
    use std::path::Path;

    fn source(rel: &str) -> String {
        fs::read_to_string(Path::new(env!("CARGO_MANIFEST_DIR")).join(rel)).unwrap_or_default()
    }

    #[test]
    fn the_shared_backup_code_is_defined_in_config_backup() {
        let new = source("src/hosts/config_backup.rs");
        for def in [
            "pub(crate) struct Backup {",
            "pub(crate) const BACKUP_DIR: &str = \"backups\";",
            "pub(crate) fn write_backup(",
            "pub const CONFIG_LIMIT: u64 = 262_144;",
        ] {
            assert!(
                new.contains(def),
                "src/hosts/config_backup.rs lacks `{def}`"
            );
        }
    }

    #[test]
    fn the_codex_config_module_defines_none_of_them() {
        let old = source("src/hosts/codex_config.rs");
        for def in [
            "struct Backup",
            "const BACKUP_DIR",
            "fn write_backup(",
            "const CONFIG_LIMIT",
        ] {
            assert!(
                !old.contains(def),
                "src/hosts/codex_config.rs still defines `{def}`"
            );
        }
    }

    #[test]
    fn the_hooks_and_the_state_preview_use_the_new_module() {
        for rel in ["src/hooks.rs", "src/go_state.rs"] {
            let text = source(rel);
            for old in [
                "codex_config::Backup",
                "codex_config::BACKUP_DIR",
                "codex_config::write_backup",
                "codex_config::CONFIG_LIMIT",
            ] {
                assert!(!text.contains(old), "{rel} still names `{old}`");
            }
        }
        assert!(source("src/hooks.rs").contains("config_backup"));
        assert!(source("src/go_state.rs").contains("config_backup::CONFIG_LIMIT"));
    }
}
