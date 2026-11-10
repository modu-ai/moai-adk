//! #2108: the SYNC workflow names the PR operation marker in the 06-C form, with the
//! repository id (stable across renames) in the project slot, never the repository name.
use moai_cli::forge::outbox::parse_op_markers;
use std::fs;

const TEMPLATE_06C: &str = "<!-- moai-op pr_create:<repository_id>:<task_key>:<epoch> -->";

fn read(path: &str) -> String {
    fs::read_to_string(path).unwrap_or_else(|e| panic!("read {path}: {e}"))
}

#[test]
fn sync_workflow_marker_template_matches_06c() {
    let contract = read("docs/pr-contract.md");
    assert!(
        contract.contains(TEMPLATE_06C),
        "docs/pr-contract.md no longer carries the 06-C template"
    );
    assert!(
        read("src/pr_link.rs").contains(TEMPLATE_06C),
        "src/pr_link.rs no longer documents the 06-C template"
    );
    let sync = read("skills/moai/workflows/sync.md");
    assert!(
        sync.contains(TEMPLATE_06C),
        "sync.md must show the 06-C PR marker template"
    );
    assert!(
        sync.contains("--jq .id"),
        "sync.md must say how the repository id is read"
    );
}

#[test]
fn sync_workflow_marker_examples_use_a_numeric_repository_id() {
    let sync = read("skills/moai/workflows/sync.md");
    let examples = parse_op_markers(&sync);
    assert!(
        !examples.is_empty(),
        "sync.md must carry a concrete marker example"
    );
    for m in examples {
        assert_eq!(m.operation, "pr_create");
        assert!(
            m.project_id.bytes().all(|b| b.is_ascii_digit()),
            "project slot must be the numeric repository id, got {:?}",
            m.project_id
        );
    }
}
