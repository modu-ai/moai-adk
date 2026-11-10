//! C10b executes all 42 scored tasks through three cumulative paths.
use serde_json::Value;
use std::{fs, path::Path, process::Command};

#[test]
fn frozen_corpus_three_arm_results_are_scored_and_each_stage_runs() {
    let out = tempfile::TempDir::new().unwrap();
    let run = Command::new("python3")
        .arg("-I")
        // -I excludes the script directory; runpy adds only our checked-in tools.
        .arg("-c")
        .arg("import runpy,sys; sys.path.insert(0,sys.argv.pop(1)); runpy.run_module('atlas_three_arm',run_name='__main__')")
        .arg(Path::new(env!("CARGO_MANIFEST_DIR")).join("tools"))
        .arg("--binary")
        .arg(env!("CARGO_BIN_EXE_moai-cli"))
        .arg("--output-dir")
        .arg(out.path())
        .output()
        .unwrap();
    assert!(
        run.status.success(),
        "{}{}",
        String::from_utf8_lossy(&run.stdout),
        String::from_utf8_lossy(&run.stderr)
    );
    let summary: Value =
        serde_json::from_slice(&fs::read(out.path().join("summary.json")).unwrap()).unwrap();
    let arms = summary["arms"].as_array().unwrap();
    assert_eq!(arms.len(), 3);
    for (i, name) in ["rg", "+Map", "+Graph"].iter().enumerate() {
        let arm = &arms[i];
        assert_eq!(arm["arm"], *name);
        assert_eq!(arm["tasks"], 42);
        assert_eq!(arm["passed"], true);
        assert_eq!(arm["operation_calls"]["search"], 42);
        if i >= 1 {
            assert_eq!(arm["operation_calls"]["atlas-map"], 41);
        }
        if i == 2 {
            assert_eq!(arm["operation_calls"]["atlas-graph"], 41);
        }
        assert_eq!(arm["tool_calls"], 42 + 41 * i);
        assert!(arm["response_bytes"].as_u64().unwrap() > 0);
        assert_eq!(arm["snapshot_stale_ratio"], 0.0);
    }
}
