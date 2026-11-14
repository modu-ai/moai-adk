//! #2936 R9 S3: the project `.omp/agents` files — OMP reads project agents from there and does
//! not read `.claude/agents`. Preview, apply, the user's own files kept, links refused.
//! Temporary folders only; the real HOME is never read or written.

use moai_cli::{agent_defs, omp_agents, state_root};
use serde_json::Value;
use std::{collections::BTreeMap, fs, path::Path, path::PathBuf};
use tempfile::TempDir;

const PERMISSION: &str = "fixture-approved";
const STATE: &str = "manifests/omp-agents.json";

fn copy_tree(from: &Path, to: &Path) {
    fs::create_dir_all(to).unwrap();
    for e in fs::read_dir(from).unwrap() {
        let e = e.unwrap();
        let target = to.join(e.file_name());
        if e.file_type().unwrap().is_dir() {
            copy_tree(&e.path(), &target);
        } else {
            fs::copy(e.path(), target).unwrap();
        }
    }
}

struct Fx {
    _tmp: TempDir,
    project: PathBuf,
    package: PathBuf,
    agents: PathBuf,
}

fn fx() -> Fx {
    let tmp = TempDir::new().unwrap();
    let base = fs::canonicalize(tmp.path()).unwrap();
    let repo = Path::new(env!("CARGO_MANIFEST_DIR"));
    let project = base.join("project");
    fs::create_dir_all(&project).unwrap();
    let package = base.join("package");
    copy_tree(&repo.join("skills/moai"), &package);
    let agents = base.join("agents");
    copy_tree(&repo.join("agents"), &agents);
    Fx {
        _tmp: tmp,
        project,
        package,
        agents,
    }
}

impl Fx {
    fn plan(&self) -> Result<Value, String> {
        omp_agents::plan(&self.project, &self.package, &self.agents)
    }
    fn apply(&self, digest: &str) -> Result<Value, String> {
        omp_agents::apply(
            &self.project,
            &self.package,
            &self.agents,
            digest,
            PERMISSION,
        )
    }
    /// The bytes OMP is meant to read: what the package renders for OMP, file name → text.
    fn rendered(&self) -> BTreeMap<String, String> {
        agent_defs::package_agents(&self.agents, &self.package)
            .unwrap()
            .iter()
            .map(|d| {
                let r = agent_defs::render(d, "omp").unwrap();
                (r.path.strip_prefix("agents/").unwrap().to_string(), r.text)
            })
            .collect()
    }
    fn dir(&self) -> PathBuf {
        self.project.join(".omp/agents")
    }
    fn state(&self) -> PathBuf {
        state_root::path(&self.project, STATE)
    }
    fn actions(plan: &Value) -> BTreeMap<String, String> {
        plan["files"]
            .as_array()
            .unwrap()
            .iter()
            .map(|f| {
                (
                    f["file"].as_str().unwrap().into(),
                    f["action"].as_str().unwrap().into(),
                )
            })
            .collect()
    }
    /// An agent source edit, so the next rendering of `moai-impl` differs.
    fn change_impl_source(&self) {
        let path = self.agents.join("impl.md");
        let mut text = fs::read_to_string(&path).unwrap();
        text.push_str("\nOne more line for the next rendering.\n");
        fs::write(path, text).unwrap();
    }
}

/// Every file under `root` with its bytes, links not followed; the way to say "nothing written".
fn snapshot(root: &Path) -> BTreeMap<String, Vec<u8>> {
    fn walk(root: &Path, dir: &Path, out: &mut BTreeMap<String, Vec<u8>>) {
        let Ok(rd) = fs::read_dir(dir) else { return };
        for e in rd {
            let e = e.unwrap();
            let rel = e
                .path()
                .strip_prefix(root)
                .unwrap()
                .to_string_lossy()
                .into_owned();
            let ft = e.file_type().unwrap();
            if ft.is_symlink() {
                out.insert(format!("{rel} -> link"), Vec::new());
            } else if ft.is_dir() {
                out.insert(format!("{rel}/"), Vec::new());
                walk(root, &e.path(), out);
            } else {
                out.insert(rel, fs::read(e.path()).unwrap());
            }
        }
    }
    let mut out = BTreeMap::new();
    walk(root, root, &mut out);
    out
}

fn digest(plan: &Value) -> String {
    plan["plan_digest"].as_str().unwrap().to_string()
}

#[test]
fn the_preview_lists_every_base_agent_as_create_and_writes_nothing() {
    let f = fx();
    let before = snapshot(&f.project);
    let plan = f.plan().unwrap();
    assert_eq!(plan["preflight"], "ok", "{plan}");
    assert_eq!(plan["dir"], ".omp/agents", "{plan}");
    assert_eq!(plan["writes"], false, "{plan}");
    let actions = Fx::actions(&plan);
    assert_eq!(
        actions.keys().cloned().collect::<Vec<_>>(),
        f.rendered().keys().cloned().collect::<Vec<_>>()
    );
    assert!(actions.values().all(|a| a == "create"), "{actions:?}");
    assert!(!actions.is_empty());
    // The user is told that the project folder shadows the user's own agents of the same name.
    let note = plan["shadows_user_agents"].as_str().unwrap_or_default();
    assert!(
        note.contains("shadow") && note.contains("~/.omp/agent/agents"),
        "{plan}"
    );
    assert_eq!(snapshot(&f.project), before, "a preview writes nothing");
}

#[test]
fn apply_writes_exactly_the_omp_rendering_and_records_ownership() {
    let f = fx();
    let plan = f.plan().unwrap();
    let r = f.apply(&digest(&plan)).unwrap();
    assert_eq!(r["state"], "applied", "{r}");
    let rendered = f.rendered();
    for (file, text) in &rendered {
        assert_eq!(
            fs::read(f.dir().join(file)).unwrap(),
            text.as_bytes(),
            "{file}"
        );
    }
    let written: Vec<String> = fs::read_dir(f.dir())
        .unwrap()
        .map(|e| e.unwrap().file_name().to_string_lossy().into_owned())
        .collect();
    assert_eq!(written.len(), rendered.len(), "{written:?}");
    let state: Value = serde_json::from_slice(&fs::read(f.state()).unwrap()).unwrap();
    for (file, text) in &rendered {
        assert_eq!(
            state["owned"][file],
            moai_cli::digest(text.as_bytes()),
            "{file}"
        );
    }
    // Nothing was written outside .omp/agents and the state folder.
    for key in snapshot(&f.project).keys() {
        let state_dir = state_root::rel(&f.project, "");
        assert!(
            key.starts_with(".omp") || key.starts_with(&state_dir),
            "{key}"
        );
    }
}

#[test]
fn a_second_run_is_unchanged_and_writes_nothing() {
    let f = fx();
    f.apply(&digest(&f.plan().unwrap())).unwrap();
    let before = snapshot(&f.project);
    let plan = f.plan().unwrap();
    assert!(
        Fx::actions(&plan).values().all(|a| a == "unchanged"),
        "{plan}"
    );
    let r = f.apply(&digest(&plan)).unwrap();
    assert_eq!(r["state"], "unchanged", "{r}");
    assert_eq!(snapshot(&f.project), before);
}

#[test]
fn a_changed_source_updates_the_files_moai_wrote_and_only_those() {
    let f = fx();
    f.apply(&digest(&f.plan().unwrap())).unwrap();
    f.change_impl_source();
    let plan = f.plan().unwrap();
    let actions = Fx::actions(&plan);
    assert_eq!(actions["moai-impl.md"], "update", "{actions:?}");
    assert_eq!(
        actions.values().filter(|a| *a == "update").count(),
        1,
        "{actions:?}"
    );
    f.apply(&digest(&plan)).unwrap();
    assert_eq!(
        fs::read(f.dir().join("moai-impl.md")).unwrap(),
        f.rendered()["moai-impl.md"].as_bytes()
    );
}

#[test]
fn an_agent_file_the_user_edited_is_kept_as_preserved_then_as_conflict() {
    let f = fx();
    f.apply(&digest(&f.plan().unwrap())).unwrap();
    let mine = b"my own moai-audit, edited by hand\n";
    fs::write(f.dir().join("moai-audit.md"), mine).unwrap();
    // The rendering did not change since our last apply: the edit stays, nothing to do.
    let plan = f.plan().unwrap();
    assert_eq!(Fx::actions(&plan)["moai-audit.md"], "preserved", "{plan}");
    f.apply(&digest(&plan)).unwrap();
    assert_eq!(fs::read(f.dir().join("moai-audit.md")).unwrap(), mine);
    // Both the user and the package changed it: a conflict, still never overwritten.
    let audit_src = f.agents.join("audit.md");
    let mut text = fs::read_to_string(&audit_src).unwrap();
    text.push_str("\nA new package line.\n");
    fs::write(&audit_src, text).unwrap();
    let plan = f.plan().unwrap();
    assert_eq!(Fx::actions(&plan)["moai-audit.md"], "conflict", "{plan}");
    f.apply(&digest(&plan)).unwrap();
    assert_eq!(fs::read(f.dir().join("moai-audit.md")).unwrap(), mine);
}

#[test]
fn a_same_name_file_moai_never_wrote_is_never_overwritten() {
    let f = fx();
    fs::create_dir_all(f.dir()).unwrap();
    let theirs = b"---\nname: moai-impl\n---\nthe team's own agent\n";
    fs::write(f.dir().join("moai-impl.md"), theirs).unwrap();
    let plan = f.plan().unwrap();
    assert_eq!(Fx::actions(&plan)["moai-impl.md"], "conflict", "{plan}");
    f.apply(&digest(&plan)).unwrap();
    assert_eq!(fs::read(f.dir().join("moai-impl.md")).unwrap(), theirs);
    // The other agents were still written.
    assert!(f.dir().join("moai-audit.md").is_file());
}

#[test]
fn an_agent_under_another_name_is_neither_listed_nor_touched() {
    let f = fx();
    fs::create_dir_all(f.dir()).unwrap();
    let mine = b"---\nname: reviewer\n---\nmine\n";
    fs::write(f.dir().join("reviewer.md"), mine).unwrap();
    fs::write(f.dir().join("notes.txt"), b"x").unwrap();
    let plan = f.plan().unwrap();
    assert!(!Fx::actions(&plan).contains_key("reviewer.md"), "{plan}");
    assert!(!Fx::actions(&plan).contains_key("notes.txt"), "{plan}");
    f.apply(&digest(&plan)).unwrap();
    assert_eq!(fs::read(f.dir().join("reviewer.md")).unwrap(), mine);
    assert_eq!(fs::read(f.dir().join("notes.txt")).unwrap(), b"x");
}

#[test]
fn a_file_moai_wrote_that_the_package_no_longer_renders_is_reported_never_deleted() {
    let f = fx();
    f.apply(&digest(&f.plan().unwrap())).unwrap();
    fs::remove_file(f.agents.join("research.md")).unwrap();
    let plan = f.plan().unwrap();
    assert_eq!(
        Fx::actions(&plan)["moai-research.md"],
        "remove_blocked",
        "{plan}"
    );
    f.apply(&digest(&plan)).unwrap();
    assert!(f.dir().join("moai-research.md").is_file());
}

#[test]
fn a_stale_digest_and_a_missing_permission_are_refused_before_any_write() {
    let f = fx();
    let plan = f.plan().unwrap();
    let before = snapshot(&f.project);
    assert!(f.apply("sha256:0").is_err());
    assert!(omp_agents::apply(&f.project, &f.package, &f.agents, &digest(&plan), "").is_err());
    assert_eq!(snapshot(&f.project), before);
    // The folder changed after the preview: the digest no longer matches.
    fs::create_dir_all(f.dir()).unwrap();
    fs::write(
        f.dir().join("moai-impl.md"),
        b"appeared after the preview\n",
    )
    .unwrap();
    let after = snapshot(&f.project);
    let e = f.apply(&digest(&plan)).unwrap_err();
    assert!(e.contains("stale"), "{e}");
    assert_eq!(snapshot(&f.project), after);
}

#[cfg(unix)]
mod links {
    use super::*;
    use std::os::unix::fs::symlink;

    fn elsewhere(f: &Fx) -> PathBuf {
        let out = f.project.parent().unwrap().join("elsewhere");
        fs::create_dir_all(&out).unwrap();
        out
    }

    fn refused(f: &Fx, out: &Path) {
        let plan = f.plan().unwrap();
        assert_eq!(plan["preflight"], "blocked", "{plan}");
        assert!(
            plan["blocked_reason"].as_str().unwrap().contains("symlink"),
            "{plan}"
        );
        assert!(plan["files"].as_array().unwrap().is_empty(), "{plan}");
        let outside = snapshot(out);
        assert!(f.apply(&digest(&plan)).is_err());
        assert_eq!(
            snapshot(out),
            outside,
            "nothing is written through the link"
        );
    }

    #[test]
    fn a_linked_omp_folder_blocks_the_project() {
        let f = fx();
        let out = elsewhere(&f);
        symlink(&out, f.project.join(".omp")).unwrap();
        refused(&f, &out);
    }

    #[test]
    fn a_linked_agents_folder_blocks_the_project() {
        let f = fx();
        let out = elsewhere(&f);
        fs::create_dir_all(f.project.join(".omp")).unwrap();
        symlink(&out, f.dir()).unwrap();
        refused(&f, &out);
    }

    #[test]
    fn a_linked_state_folder_blocks_the_project() {
        let f = fx();
        let out = elsewhere(&f);
        let state = state_root::path(&f.project, "");
        symlink(&out, state).unwrap();
        refused(&f, &out);
    }

    #[test]
    fn a_linked_agent_file_blocks_the_project_and_is_not_followed() {
        let f = fx();
        let out = elsewhere(&f);
        let target = out.join("target.md");
        fs::write(&target, b"somebody else's file\n").unwrap();
        fs::create_dir_all(f.dir()).unwrap();
        symlink(&target, f.dir().join("moai-impl.md")).unwrap();
        refused(&f, &out);
        assert_eq!(fs::read(&target).unwrap(), b"somebody else's file\n");
    }

    #[test]
    fn a_folder_that_becomes_a_link_between_preview_and_apply_is_not_written_through() {
        let f = fx();
        let plan = f.plan().unwrap();
        let out = elsewhere(&f);
        symlink(&out, f.project.join(".omp")).unwrap();
        let outside = snapshot(&out);
        assert!(f.apply(&digest(&plan)).is_err());
        assert_eq!(snapshot(&out), outside);
    }
}
