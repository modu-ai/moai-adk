//! SPEC-20261126-project-harness-namespace: `generation-plan` and `generation-apply` write the
//! agents of a project in `.claude/agents/moai/` and its skills (project skills and skill
//! projections) in `.claude/skills/moai-<id>/`, with the front matter `name` equal to the folder
//! name. A skill id that would make the folder name longer than 64 characters is refused. A skill
//! file an earlier run wrote at `.claude/skills/<id>/SKILL.md` moves into the new folder through
//! the plan when the generation record owns it and its bytes are the recorded ones; an edited one
//! stays and is a conflict; a skill folder the record does not own is never touched.
//!
//! The agent path `.claude/agents/moai/<role>.md` was already in place before this SPEC; the
//! `fresh_` test pins it so that the three paths are fixed in one place.

mod common;

use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::Stdio;
use tempfile::TempDir;

const AGENT: &str = ".claude/agents/moai/prj-reviewer.md";
const SKILL: &str = ".claude/skills/moai-prj-checks/SKILL.md";
const SKILL_OLD: &str = ".claude/skills/prj-checks/SKILL.md";
const PROJECTION: &str = ".claude/skills/moai-ref-b/SKILL.md";
const PROJECTION_OLD: &str = ".claude/skills/ref-b/SKILL.md";

/// The longest folder name the platform documents for a skill name.
const FOLDER_NAME_LIMIT: usize = 64;
const FOLDER_PREFIX: &str = "moai-";

/// The project's state folder, named by the resolver the product uses.
fn state(root: &Path) -> PathBuf {
    root.join(moai_cli::state_root::dir(root))
}

/// Runs `moai` with a clean environment and a HOME of its own, never the real one.
fn cli(args: &[&str], input: &Value) -> std::process::Output {
    let home = common::home::TempHome::new();
    let mut child = home
        .command(args)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    common::feed_stdin(&mut child, input.to_string().as_bytes());
    child.wait_with_output().unwrap()
}

fn said(out: &std::process::Output) -> String {
    format!(
        "{}{}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    )
}

fn result(out: &std::process::Output) -> Value {
    let v: Value = serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|_| panic!("{}", String::from_utf8_lossy(&out.stderr)));
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

/// A project with a profile and a license-reviewed source for the projection source
/// `fixtures/skill-b`.
fn project() -> TempDir {
    let root = TempDir::new().unwrap();
    let r = root.path();
    fs::create_dir_all(r.join("skills/moai/workflows")).unwrap();
    fs::create_dir_all(state(r)).unwrap();
    let profile = json!({"schema_version": 2, "project_id": "demo", "purpose": "Build a small editor",
    "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude"],
    "interview_revision": "r1", "decisions": [
        {"id": "test_command", "topic": "test_command", "value": "cargo test --locked",
         "source": "user", "scope": "project"},
    ]});
    let mut bytes = serde_json::to_vec_pretty(&profile).unwrap();
    bytes.push(b'\n');
    fs::write(state(r).join("project.yaml"), bytes).unwrap();
    let sources = json!({"schema_version": 1, "sources": [
        {"source_url": "fixtures/skill-b", "license": {"status": "reviewed", "spdx": "MIT"},
         "resolved_revision": null}]});
    let at = moai_cli::state_root::path(r, "manifests/sources.json");
    fs::create_dir_all(at.parent().unwrap()).unwrap();
    fs::write(at, serde_json::to_vec_pretty(&sources).unwrap()).unwrap();
    root
}

fn request(roles: Value, skills: Value, projections: Value) -> Value {
    json!({"protocol_version": 1, "language": "en", "workflow_reference": "skills/moai/workflows",
        "local_file": "none", "workflow_id": "review-flow", "native_roles": roles,
        "skill_projections": projections, "project_skills": skills})
}

fn role_list() -> Value {
    json!([{"role_id": "prj-reviewer", "host": "claude", "purpose": "Review a change before it lands"}])
}

fn skill_list() -> Value {
    json!([{"id": "prj-checks", "purpose": "How this project is checked"}])
}

fn projection(id: &str) -> Value {
    json!({"id": id, "source_url": "fixtures/skill-b", "host_targets": ["claude"]})
}

/// One agent, one project skill and one skill projection.
fn full_request() -> Value {
    request(role_list(), skill_list(), json!([projection("ref-b")]))
}

fn plan(root: &Path, req: &Value) -> std::process::Output {
    cli(&["generation-plan", root.to_str().unwrap()], req)
}

fn apply(root: &Path, p: &Value) -> std::process::Output {
    cli(
        &[
            "generation-apply",
            root.to_str().unwrap(),
            "--plan-digest",
            p["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "fixture-approved",
        ],
        p,
    )
}

fn entry<'a>(p: &'a Value, path: &str) -> Option<&'a Value> {
    p["files"]
        .as_array()
        .unwrap()
        .iter()
        .find(|f| f["path"] == json!(path))
}

fn action(p: &Value, path: &str) -> Option<String> {
    entry(p, path).map(|f| f["action"].as_str().unwrap().to_string())
}

/// The paths of a plan that belong to the generation of agents and skills (not the workflow
/// files and not the files of the agents engine).
fn generated(p: &Value) -> Vec<String> {
    let mut paths: Vec<String> = p["files"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|f| {
            f["kind"].as_str().is_some_and(|k| {
                k.starts_with("generation:") && !k.starts_with("generation:workflow")
            })
        })
        .map(|f| f["path"].as_str().unwrap().to_string())
        .collect();
    paths.sort();
    paths
}

fn folder_name_of(id: &str) -> String {
    format!("{FOLDER_PREFIX}{id}")
}

fn read(root: &Path, rel: &str) -> String {
    fs::read_to_string(root.join(rel)).unwrap()
}

/// An id `prefix` plus letters, as long as `len` says.
fn id_of_length(prefix: &str, len: usize) -> String {
    format!("{prefix}{}", "a".repeat(len - prefix.len()))
}

/// Applies the full request, then leaves the project as an earlier version left it: each skill
/// file of ours is at `.claude/skills/<id>/SKILL.md` with the front matter `name` equal to the id,
/// and its record entry names that path with the digest of those bytes. This works on whatever
/// layout the current build writes, so the same seed serves before and after the change.
fn an_earlier_run() -> TempDir {
    let root = project();
    let r = root.path();
    result(&apply(r, &result(&plan(r, &full_request()))));
    let manifest_path = state(r).join("manifests/generation.json");
    let mut manifest: Value = serde_json::from_slice(&fs::read(&manifest_path).unwrap()).unwrap();
    let mut seeded = 0;
    for owned in manifest["owned_files"].as_array_mut().unwrap() {
        let kind = owned["section_id"].as_str().unwrap_or("").to_string();
        let Some(id) = [
            "generation:project-skill:claude:",
            "generation:skill-projection:claude:",
        ]
        .iter()
        .find_map(|p| kind.strip_prefix(p)) else {
            continue;
        };
        let now = owned["path"].as_str().unwrap().to_string();
        let former = format!(".claude/skills/{id}/SKILL.md");
        let text = read(r, &now);
        let earlier = text.replacen(
            &format!("name: {}\n", folder_name_of(id)),
            &format!("name: {id}\n"),
            1,
        );
        if now != former {
            fs::remove_file(r.join(&now)).unwrap();
            fs::remove_dir(r.join(&now).parent().unwrap()).unwrap();
        }
        fs::create_dir_all(r.join(&former).parent().unwrap()).unwrap();
        fs::write(r.join(&former), &earlier).unwrap();
        owned["path"] = json!(former);
        owned["baseline_sha256"] = json!(moai_cli::digest(earlier.as_bytes()));
        seeded += 1;
    }
    assert_eq!(seeded, 2, "one project skill and one projection are seeded");
    let mut bytes = serde_json::to_vec_pretty(&manifest).unwrap();
    bytes.push(b'\n');
    fs::write(manifest_path, bytes).unwrap();
    for former in [SKILL_OLD, PROJECTION_OLD] {
        assert!(
            read(r, former).starts_with(&format!(
                "---\nname: {}\n",
                Path::new(former)
                    .parent()
                    .unwrap()
                    .file_name()
                    .unwrap()
                    .to_str()
                    .unwrap()
            )),
            "{former} is as an earlier version wrote it"
        );
    }
    root
}

#[test]
fn fresh_plan_names_the_namespaced_agent_and_skill_paths() {
    let root = project();
    let r = root.path();
    let p = result(&plan(r, &full_request()));
    // The agent path was already namespaced before this SPEC; it is pinned here.
    assert_eq!(action(&p, AGENT).as_deref(), Some("create"), "{p}");
    assert_eq!(
        generated(&p),
        vec![AGENT, PROJECTION, SKILL],
        "a plan names the agent, the skill and the projection at their namespaced paths only"
    );
    for path in [SKILL, PROJECTION] {
        assert_eq!(action(&p, path).as_deref(), Some("create"), "{path}");
    }
    for former in [SKILL_OLD, PROJECTION_OLD] {
        assert!(entry(&p, former).is_none(), "nothing planned at {former}");
    }
    result(&apply(r, &p));
    for former in [SKILL_OLD, PROJECTION_OLD] {
        assert!(!r.join(former).exists(), "nothing written at {former}");
    }
    assert!(
        read(r, AGENT).starts_with("---\nname: prj-reviewer\n"),
        "the agent keeps its role id as its name"
    );
    assert!(
        read(r, SKILL).starts_with("---\nname: moai-prj-checks\n"),
        "the project skill name is its folder name"
    );
    assert!(
        read(r, PROJECTION).starts_with("---\nname: moai-ref-b\n"),
        "the projection name is its folder name"
    );
    let again = result(&plan(r, &full_request()));
    for path in [AGENT, SKILL, PROJECTION] {
        assert_eq!(action(&again, path).as_deref(), Some("unchanged"), "{path}");
    }
}

/// The id that makes `moai-<id>` exactly `folder_len` characters.
fn id_for_folder(prefix: &str, folder_len: usize) -> String {
    id_of_length(prefix, folder_len - FOLDER_PREFIX.len())
}

fn assert_refused_for(req: &Value, root: &Path, field: &str, id: &str) {
    let out = plan(root, req);
    assert!(!out.status.success(), "{}", said(&out));
    let text = said(&out);
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], json!(false), "{text}");
    assert!(text.contains(field), "names the field {field}: {text}");
    assert!(
        text.contains(&FOLDER_NAME_LIMIT.to_string()),
        "names the limit: {text}"
    );
    assert!(!text.contains(id), "never shows the id: {text}");
}

#[test]
fn folder_name_of_sixty_four_characters_is_planned_and_sixty_five_is_refused() {
    let root = project();
    let r = root.path();
    let longest = id_for_folder("prj-", FOLDER_NAME_LIMIT);
    let too_long = id_for_folder("prj-", FOLDER_NAME_LIMIT + 1);
    assert_eq!(folder_name_of(&longest).len(), FOLDER_NAME_LIMIT);
    assert_eq!(folder_name_of(&too_long).len(), FOLDER_NAME_LIMIT + 1);
    let skills = |id: &str| json!([{"id": id, "purpose": "How this project is checked"}]);
    // The first value over the limit is refused with the field and the limit, never the value.
    assert_refused_for(
        &request(json!([]), skills(&too_long), json!([])),
        r,
        "project_skills",
        &too_long,
    );
    // The last value under it is planned at the namespaced path.
    let p = result(&plan(r, &request(json!([]), skills(&longest), json!([]))));
    let path = format!(".claude/skills/{}/SKILL.md", folder_name_of(&longest));
    assert_eq!(action(&p, &path).as_deref(), Some("create"), "{p}");
    assert!(
        entry(&p, &format!(".claude/skills/{longest}/SKILL.md")).is_none(),
        "nothing planned at the former path"
    );
}

#[test]
fn folder_name_limit_holds_for_a_skill_projection_id() {
    let root = project();
    let r = root.path();
    let longest = id_for_folder("ref-", FOLDER_NAME_LIMIT);
    let too_long = id_for_folder("ref-", FOLDER_NAME_LIMIT + 1);
    assert_refused_for(
        &request(json!([]), json!([]), json!([projection(&too_long)])),
        r,
        "skill_projections",
        &too_long,
    );
    let p = result(&plan(
        r,
        &request(json!([]), json!([]), json!([projection(&longest)])),
    ));
    let path = format!(".claude/skills/{}/SKILL.md", folder_name_of(&longest));
    assert_eq!(action(&p, &path).as_deref(), Some("create"), "{p}");
    assert!(
        entry(&p, &format!(".claude/skills/{longest}/SKILL.md")).is_none(),
        "nothing planned at the former path"
    );
}

#[test]
fn moved_owned_unedited_skill_moves_with_its_body_kept() {
    let root = an_earlier_run();
    let r = root.path();
    let before: Vec<(&str, &str, String)> = [
        (SKILL_OLD, SKILL, read(r, SKILL_OLD)),
        (PROJECTION_OLD, PROJECTION, read(r, PROJECTION_OLD)),
    ]
    .into();
    let p = result(&plan(r, &full_request()));
    for (old, new, _) in &before {
        assert_eq!(action(&p, old).as_deref(), Some("remove"), "{old} leaves");
        assert_eq!(action(&p, new).as_deref(), Some("create"), "{new} arrives");
        assert!(r.join(old).exists(), "a plan writes nothing: {old}");
    }
    result(&apply(r, &p));
    for (old, new, was) in &before {
        assert!(!r.join(old).exists(), "{old} is gone");
        let folder = Path::new(old).parent().unwrap();
        assert!(!r.join(folder).exists(), "its emptied folder is gone too");
        // The name follows the folder; the rest of the file is what it was.
        let id = folder.file_name().unwrap().to_str().unwrap();
        let expected = was.replacen(
            &format!("name: {id}\n"),
            &format!("name: {}\n", folder_name_of(id)),
            1,
        );
        assert_eq!(read(r, new), expected, "{new} keeps the body of {old}");
    }
    let again = result(&plan(r, &full_request()));
    for path in [AGENT, SKILL, PROJECTION] {
        assert_eq!(action(&again, path).as_deref(), Some("unchanged"), "{path}");
    }
    for old in [SKILL_OLD, PROJECTION_OLD] {
        assert!(
            entry(&again, old).is_none(),
            "the record no longer names {old}"
        );
    }
}

#[test]
fn moved_edited_skill_stays_and_both_paths_are_conflict() {
    let root = an_earlier_run();
    let r = root.path();
    let edited = "---\nname: prj-checks\ndescription: mine\n---\n\nmine\n";
    fs::write(r.join(SKILL_OLD), edited).unwrap();
    let p = result(&plan(r, &full_request()));
    assert_eq!(action(&p, SKILL_OLD).as_deref(), Some("conflict"), "{p}");
    assert_eq!(action(&p, SKILL).as_deref(), Some("conflict"), "{p}");
    let why = entry(&p, SKILL_OLD).unwrap()["detail"]
        .as_str()
        .unwrap_or("");
    assert!(!why.is_empty(), "the conflict says why: {p}");
    assert!(!apply(r, &p).status.success(), "apply refuses a conflict");
    assert_eq!(read(r, SKILL_OLD), edited, "the edited file stays as it is");
    assert!(!r.join(SKILL).exists(), "no copy is written beside it");
}

#[test]
fn moved_user_skill_folder_beside_ours_is_left_alone() {
    let root = project();
    let r = root.path();
    let user_files: [(&str, &str); 4] = [
        (
            ".claude/skills/my-own/SKILL.md",
            "---\nname: my-own\n---\nmine\n",
        ),
        (".claude/skills/my-own/extra.md", "extra\n"),
        // A skill of the user's at the former path of ours, with no record entry; its `name`
        // differs from the `name` of the new file, so the new path is not a conflict.
        (
            SKILL_OLD,
            "---\nname: prj-checks\ndescription: my own\n---\n\nmy own\n",
        ),
        (".claude/skills/prj-checks/notes.md", "notes\n"),
    ];
    for (rel, body) in user_files {
        fs::create_dir_all(r.join(rel).parent().unwrap()).unwrap();
        fs::write(r.join(rel), body).unwrap();
    }
    let p = result(&plan(r, &request(json!([]), skill_list(), json!([]))));
    for (rel, _) in user_files {
        assert!(entry(&p, rel).is_none(), "{rel} is not in the plan");
    }
    assert_eq!(
        action(&p, SKILL).as_deref(),
        Some("create"),
        "a user skill at the former path is no conflict for the new path: {p}"
    );
    result(&apply(r, &p));
    for (rel, body) in user_files {
        assert_eq!(read(r, rel), body, "{rel} is unchanged");
    }
    assert!(read(r, SKILL).starts_with("---\nname: moai-prj-checks\n"));
}

#[test]
fn moved_former_path_edited_after_the_plan_is_not_removed() {
    let root = an_earlier_run();
    let r = root.path();
    let p = result(&plan(r, &full_request()));
    assert_eq!(action(&p, SKILL_OLD).as_deref(), Some("remove"), "{p}");
    assert_eq!(action(&p, SKILL).as_deref(), Some("create"), "{p}");
    let edited = "---\nname: prj-checks\ndescription: edited after the plan\n---\n\nmine\n";
    fs::write(r.join(SKILL_OLD), edited).unwrap();
    assert!(
        !apply(r, &p).status.success(),
        "a plan over changed bytes is refused"
    );
    assert_eq!(read(r, SKILL_OLD), edited, "the file is not removed");
}

/// How many words of `text` are a clock time like `21:55:57Z`.
fn read_times(text: &str) -> usize {
    text.split(|c: char| !(c.is_ascii_alphanumeric() || c == ':'))
        .filter(|t| {
            let b = t.as_bytes();
            b.len() == 9
                && b[8] == b'Z'
                && b[2] == b':'
                && b[5] == b':'
                && [0, 1, 3, 4, 6, 7].iter().all(|&i| b[i].is_ascii_digit())
        })
        .count()
}

#[test]
fn research_read_time_counter_counts_only_clock_times() {
    assert_eq!(read_times("S, 21:55:57Z; K, 21:55:50Z."), 2);
    assert_eq!(read_times("21:55:57 1:55:57Z 21-55-57Z 21:55:57Zulu"), 0);
    assert_eq!(read_times(""), 0);
}

#[test]
fn research_holds_official_quotes_and_read_times() {
    let path = Path::new(env!("CARGO_MANIFEST_DIR"))
        .join(".moai/specs/SPEC-20261126-project-harness-namespace/research.md");
    let text = fs::read_to_string(&path).unwrap();
    for quote in [
        "identity comes only from the `name` frontmatter field",
        "`.claude/skills/<skill-name>/SKILL.md`",
        "Maximum 64 characters",
    ] {
        assert!(text.contains(quote), "research.md quotes: {quote}");
    }
    let times = read_times(&text);
    assert!(times >= 4, "research.md has {times} read times, wants 4");
}

#[test]
fn text_step_six_names_the_namespace_folders() {
    let path = Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai/workflows/project.md");
    let text = fs::read_to_string(&path).unwrap();
    let step = text
        .lines()
        .find(|l| l.starts_with("6. **Propose a project harness.**"))
        .expect("step 6 of the project route");
    for named in [".claude/agents/moai/", ".claude/skills/moai-"] {
        assert!(step.contains(named), "step 6 names {named}");
    }
}
