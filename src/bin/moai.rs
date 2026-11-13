//! `moai`: the public CLI name (user RC2 instruction, Issue #1975 comment 6859326988; #2790
//! R1). It runs the same `main` as `moai-cli`, loaded from `src/main.rs` as a module, so the
//! two binaries cannot drift: same commands, same output, same `--version` line
//! `moai-cli <ver> (rust)`.

#[path = "../main.rs"]
mod cli;

fn main() -> std::process::ExitCode {
    cli::main()
}
