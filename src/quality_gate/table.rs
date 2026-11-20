//! The 16-language toolchain table of the Go build (`toolchains` in `quality/gate.go`), the
//! per-language variants (Flutter, Go build tags, Python runner) and the step type.

use std::path::Path;

#[derive(Debug, Clone, Default, PartialEq)]
pub struct Step {
    pub name: String,
    pub binary: String,
    pub args: Vec<String>,
    /// Skip when the binary is not on PATH.
    pub optional: bool,
    /// Skip when none of these files exists in the module root.
    pub config_files: Vec<String>,
    /// Skip when git says files are staged and none carries one of these extensions.
    pub changed_exts: Vec<String>,
    /// Skip when the module holds no file with one of these extensions.
    pub source_exts: Vec<String>,
}

#[derive(Debug, Clone, Default, PartialEq)]
pub struct Toolchain {
    pub markers: Vec<String>,
    pub vet: Vec<Step>,
    pub lint: Vec<Step>,
    pub test: Option<Step>,
    /// Node projects: which script and package manager run is chosen from `package.json`, a
    /// decision this build does not make yet (its own sub-slice), so their steps are named
    /// in the summary and not run.
    pub node: bool,
}

pub const PYTEST: &str = "pytest";

fn s(name: &str, binary: &str, args: &[&str]) -> Step {
    Step {
        name: name.into(),
        binary: binary.into(),
        args: args.iter().map(|a| a.to_string()).collect(),
        ..Step::default()
    }
}

impl Step {
    fn opt(mut self) -> Self {
        self.optional = true;
        self
    }
    fn cfg(mut self, files: &[&str]) -> Self {
        self.config_files = files.iter().map(|f| f.to_string()).collect();
        self
    }
    fn changed(mut self, exts: &[&str]) -> Self {
        self.changed_exts = exts.iter().map(|f| f.to_string()).collect();
        self
    }
    fn src(mut self, exts: &[&str]) -> Self {
        self.source_exts = exts.iter().map(|f| f.to_string()).collect();
        self
    }
}

fn tc(markers: &[&str], vet: Vec<Step>, lint: Vec<Step>, test: Option<Step>) -> Toolchain {
    Toolchain {
        markers: markers.iter().map(|m| m.to_string()).collect(),
        vet,
        lint,
        test,
        node: false,
    }
}

/// The table, in the Go build's order: the first marker match at a directory wins.
pub fn toolchains() -> Vec<Toolchain> {
    let mut node = tc(
        &["package.json"],
        vec![],
        vec![
            s("eslint", "npx", &["eslint", "."]).opt().cfg(&[
                "eslint.config.js",
                "eslint.config.mjs",
                "eslint.config.cjs",
                "eslint.config.ts",
                "eslint.config.mts",
                "eslint.config.cts",
                ".eslintrc.js",
                ".eslintrc.cjs",
                ".eslintrc.yaml",
                ".eslintrc.yml",
                ".eslintrc.json",
                ".eslintrc",
            ]),
            s("biome", "npx", &["biome", "check", "."])
                .opt()
                .cfg(&["biome.json", "biome.jsonc"]),
            s("oxlint", "npx", &["oxlint"]).opt().cfg(&[
                ".oxlintrc.json",
                ".oxlintrc.jsonc",
                "oxlint.config.ts",
                "oxlint.config.mts",
            ]),
        ],
        Some(s("npm test", "npm", &["test", "--", "--passWithNoTests"])),
    );
    node.node = true;
    vec![
        tc(
            &["go.mod"],
            vec![s("go vet", "go", &["vet", "./..."]).src(&[".go"])],
            vec![s("golangci-lint", "golangci-lint", &["run"])
                .opt()
                .cfg(&[
                    ".golangci.yml",
                    ".golangci.yaml",
                    ".golangci.toml",
                    ".golangci.json",
                ])
                .src(&[".go"])],
            Some(s("go test", "go", &["test", "./..."]).src(&[".go"])),
        ),
        node,
        tc(
            &["pyproject.toml", "setup.py", "requirements.txt"],
            vec![],
            vec![
                s("ruff", "ruff", &["check", "."])
                    .opt()
                    .cfg(&["ruff.toml", ".ruff.toml", "pyproject.toml"])
                    .changed(&[".py"]),
                s("mypy", "mypy", &["."])
                    .opt()
                    .cfg(&["mypy.ini", ".mypy.ini", "pyproject.toml", "setup.cfg"])
                    .changed(&[".py"])
                    .src(&[".py", ".pyi"]),
            ],
            Some(s(PYTEST, "pytest", &[]).opt()),
        ),
        tc(
            &["Cargo.toml"],
            vec![],
            vec![s("cargo clippy", "cargo", &["clippy", "--", "-D", "warnings"]).opt()],
            Some(s("cargo test", "cargo", &["test"])),
        ),
        tc(
            &["pom.xml", "build.gradle", "build.gradle.kts"],
            vec![],
            vec![s(
                "checkstyle",
                "checkstyle",
                &["-c", "/google_checks.xml", "src/"],
            )
            .opt()
            .cfg(&["checkstyle.xml", "google_checks.xml", ".checkstyle"])],
            Some(s("mvn test", "mvn", &["test"]).opt()),
        ),
        tc(
            &["build.gradle.kts"],
            vec![],
            vec![s("ktlint", "ktlint", &[])
                .opt()
                .cfg(&[".editorconfig", ".ktlint"])],
            Some(s("gradle test", "gradle", &["test"]).opt()),
        ),
        tc(
            &["*.csproj", "*.sln"],
            vec![],
            vec![s(
                "dotnet format",
                "dotnet",
                &["format", "--verify-no-changes"],
            )
            .opt()
            .changed(&[".cs"])],
            Some(s("dotnet test", "dotnet", &["test"]).opt()),
        ),
        tc(
            &["Gemfile"],
            vec![],
            vec![s("rubocop", "rubocop", &[])
                .opt()
                .cfg(&[".rubocop.yml", ".rubocop.yaml"])],
            Some(s("rspec", "rspec", &[]).opt()),
        ),
        tc(
            &["composer.json"],
            vec![],
            vec![s("phpstan", "phpstan", &["analyse"]).opt().cfg(&[
                "phpstan.neon",
                "phpstan.neon.dist",
                "phpstan.dist.neon",
            ])],
            Some(s("phpunit", "phpunit", &[]).opt()),
        ),
        tc(
            &["Package.swift"],
            vec![],
            vec![s("swiftlint", "swiftlint", &[])
                .opt()
                .cfg(&[".swiftlint.yml", ".swiftlint.yaml"])],
            Some(s("swift test", "swift", &["test"])),
        ),
        tc(
            &["pubspec.yaml"],
            vec![s("dart analyze", "dart", &["analyze"])],
            vec![],
            Some(s("dart test", "dart", &["test"]).opt()),
        ),
        tc(
            &["mix.exs"],
            vec![],
            vec![s("credo", "mix", &["credo"]).opt().cfg(&[".credo.exs"])],
            Some(s("mix test", "mix", &["test"])),
        ),
        tc(
            &["build.sbt"],
            vec![],
            vec![s("scalafix", "scalafix", &[])
                .opt()
                .cfg(&[".scalafix.conf"])],
            Some(s("sbt test", "sbt", &["test"]).opt()),
        ),
        tc(
            &["*.cabal", "stack.yaml"],
            vec![],
            vec![],
            Some(s("cabal test", "cabal", &["test"]).opt()),
        ),
        tc(
            &["build.zig"],
            vec![],
            vec![],
            Some(s("zig test", "zig", &["test"]).opt()),
        ),
    ]
}

/// The table entry `index` with the per-language variants applied for the module at `dir`:
/// Flutter instead of Dart, Go build tags, the project's own pytest runner.
pub fn resolve(index: usize, dir: &Path, go_build_tags: &str) -> Toolchain {
    let base = toolchains().swap_remove(index);
    let t = resolve_dart_flutter(base, dir);
    let t = resolve_go_build_tags(t, go_build_tags);
    resolve_python_runner(t, dir)
}

fn file_exists(path: &Path) -> bool {
    path.metadata().map(|m| !m.is_dir()).unwrap_or(false)
}

fn resolve_dart_flutter(mut t: Toolchain, dir: &Path) -> Toolchain {
    if t.markers.first().map(String::as_str) != Some("pubspec.yaml") {
        return t;
    }
    let Ok(content) = std::fs::read_to_string(dir.join("pubspec.yaml")) else {
        return t;
    };
    let flutter = content.contains("sdk: flutter")
        || content.contains("sdk:flutter")
        || content
            .lines()
            .any(|l| l.trim_end_matches([' ', '\t']) == "flutter:");
    if flutter {
        t.vet = vec![s("flutter analyze", "flutter", &["analyze"])];
        t.test = Some(s("flutter test", "flutter", &["test"]).opt());
    }
    t
}

fn apply_go_build_tags(mut step: Step, tags: &str) -> Step {
    match step.binary.as_str() {
        "go" => {
            // `-tags=<tags>` right after the subcommand.
            let mut args = Vec::with_capacity(step.args.len() + 1);
            if let Some(first) = step.args.first() {
                args.push(first.clone());
            }
            args.push(format!("-tags={tags}"));
            args.extend(step.args.iter().skip(1).cloned());
            step.args = args;
        }
        "golangci-lint" => step.args.push(format!("--build-tags={tags}")),
        _ => {}
    }
    step
}

fn resolve_go_build_tags(mut t: Toolchain, tags: &str) -> Toolchain {
    if tags.is_empty() || t.markers.first().map(String::as_str) != Some("go.mod") {
        return t;
    }
    t.vet = t
        .vet
        .into_iter()
        .map(|x| apply_go_build_tags(x, tags))
        .collect();
    t.lint = t
        .lint
        .into_iter()
        .map(|x| apply_go_build_tags(x, tags))
        .collect();
    t.test = t.test.map(|x| apply_go_build_tags(x, tags));
    t
}

fn has_uv_project(dir: &Path) -> bool {
    if file_exists(&dir.join("uv.lock")) {
        return true;
    }
    std::fs::read_to_string(dir.join("pyproject.toml"))
        .map(|text| text.lines().any(|l| l.trim().starts_with("[tool.uv")))
        .unwrap_or(false)
}

fn resolve_python_runner(mut t: Toolchain, dir: &Path) -> Toolchain {
    let Some(step) = t.test.as_mut() else {
        return t;
    };
    if step.name != PYTEST {
        return t;
    }
    let venv = [
        dir.join(".venv/bin/pytest"),
        dir.join(".venv/Scripts/pytest.exe"),
    ]
    .into_iter()
    .find(|p| file_exists(p));
    if let Some(path) = venv {
        // An absolute path already proved to exist: no longer optional.
        step.binary = path.to_string_lossy().into_owned();
        step.args.clear();
        step.optional = false;
    } else if has_uv_project(dir) {
        step.binary = "uv".into();
        step.args = vec!["run".into(), "pytest".into()];
    } else if file_exists(&dir.join("poetry.lock")) {
        step.binary = "poetry".into();
        step.args = vec!["run".into(), "pytest".into()];
    }
    t
}
