//! #2350 audit F2: the Go parity contract of `moai-cli statusline` runs inside `cargo test`,
//! so the check set and rc-ci fail when the built binary stops matching the Go goldens.
//!
//! The comparison itself is `tools/statusline_golden.py --check` (fresh HOME and project per
//! case, the binary copied under another name, on macOS the network/HOME/exec sandbox), run
//! here verbatim against the binary this build produced. Unix only: the harness builds its
//! fixtures with POSIX paths and `/usr/bin` git; Windows CI compiles this file without
//! running it.

#[cfg(unix)]
#[test]
fn statusline_matches_every_go_golden() {
    let root = env!("CARGO_MANIFEST_DIR");
    let out = std::process::Command::new("python3")
        .args(["-I", "tools/statusline_golden.py", "--bin"])
        .arg(env!("CARGO_BIN_EXE_moai-cli"))
        .args(["--argv", "statusline", "--check"])
        .current_dir(root)
        .env("PYTHONDONTWRITEBYTECODE", "1")
        .output()
        .expect("python3 runs the golden harness");
    let stdout = String::from_utf8_lossy(&out.stdout);
    assert!(
        out.status.success(),
        "golden --check exit {:?}\n{stdout}{}",
        out.status.code(),
        String::from_utf8_lossy(&out.stderr)
    );
    // A run that compared nothing must not pass: every case and variant reports `ok`.
    let ok = stdout.lines().filter(|l| l.starts_with("ok ")).count();
    assert_eq!(ok, 57, "{stdout}");
}
