//! Maintainer tools of this repository, run as `cargo xtask <tool>` (the alias in
//! `.cargo/config.toml`). The crate has no dependencies, so it compiles in seconds, and it is not
//! part of the shipped binary, the skill package or a release (SPEC-20261121-xtask-local-check).

mod local_check;

/// The tools, each with the one line `cargo xtask` prints for it.
const TOOLS: [(&str, &str); 1] = [(
    "local-check",
    "the local check set: the CI check steps, in a candidate merge or in place",
)];

fn usage() -> String {
    let mut text = String::from("usage: cargo xtask <tool> [arguments]\n\ntools:\n");
    for (name, about) in TOOLS {
        text.push_str(&format!("  {name:<12} {about}\n"));
    }
    text
}

fn main() {
    let mut args = std::env::args().skip(1);
    let code = match args.next().as_deref() {
        Some("local-check") => local_check::run(),
        Some(other) => {
            eprintln!("xtask: unknown tool `{other}`\n\n{}", usage());
            2
        }
        None => {
            eprint!("{}", usage());
            2
        }
    };
    std::process::exit(code);
}
