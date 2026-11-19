//! The AI decision model (Jev/OpenAI) was removed in full: no code, test or SPEC refers to
//! it any more. A profile written while it existed may still carry a `decisions` entry with
//! the id `decision_model` (values `off`, `typesafe`, `openai`), and reading the profile
//! must keep working: the entry is carried as the generic decision record it now is, and
//! nothing consumes that id. This pins the no-user-breaks rule for the retired id.

use std::path::Path;

fn write_profile(root: &Path, value: &str, source: &str) {
    let dir = root.join(".moai-cli");
    std::fs::create_dir_all(&dir).unwrap();
    std::fs::write(
        dir.join("project.yaml"),
        format!(
            "{{\"schema_version\":2,\"project_id\":\"demo\",\"purpose\":\"Build a small editor\",\
             \"tools\":[\"Rust\"],\"development_methods\":[\"TDD\"],\"hosts\":[\"claude\"],\
             \"interview_revision\":\"r1\",\"decisions\":[{{\"id\":\"decision_model\",\
             \"topic\":\"Decision model\",\"value\":\"{value}\",\"source\":\"{source}\",\
             \"scope\":\"project\"}}]}}"
        ),
    )
    .unwrap();
}

#[test]
fn a_profile_still_carrying_a_decision_model_entry_reads_unchanged() {
    for (value, source) in [("off", "default"), ("typesafe", "user"), ("openai", "user")] {
        let dir = tempfile::tempdir().unwrap();
        write_profile(dir.path(), value, source);
        let (profile, _bytes) =
            moai_cli::profile::read_profile(dir.path()).expect("the retired id still reads");
        assert_eq!(
            profile.decisions.len(),
            1,
            "the record is carried, not dropped"
        );
        assert_eq!(profile.decisions[0].id, "decision_model");
        assert_eq!(profile.decisions[0].value, value);
        assert_eq!(profile.decisions[0].source, source);
    }
}
