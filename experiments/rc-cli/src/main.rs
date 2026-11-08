use std::env;
use std::process::ExitCode;

fn main() -> ExitCode {
    let args: Vec<_> = env::args_os().skip(1).collect();
    match args.as_slice() {
        [arg] if arg == "--version" => {
            println!("moai-cli 0.0.0 (development experiment)");
            ExitCode::SUCCESS
        }
        [arg] if arg == "status" => {
            println!(r#"{{"implementation":"experiment","claude":"not_run","codex":"not_run","setup":"not_implemented","update":"not_implemented"}}"#);
            ExitCode::SUCCESS
        }
        _ => {
            eprintln!("usage: moai-cli --version | status");
            ExitCode::from(2)
        }
    }
}
