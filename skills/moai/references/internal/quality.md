# Internal: quality

When to read: at VERIFY, and at RUN before the first check run, in a project MoAI described with `/moai project`; also before you add a dependency. It says which checks to run for the project's own tools, how a test is tied to the acceptance criterion it proves, and how to show that a new dependency exists.

## The checks come from the project, never from a guess

1. Read the project's facts first: `moai project-detect <project>` (it writes nothing) or the `project.yaml` it describes. `detected.languages` and `detected.manifests` name the languages and manifest files; the profile's `tools` holds what the interview confirmed.
2. For each language, take each check from the first source that has it, and record that source as `file:line`:
   - the project's own declaration: a CI workflow step, a manifest script (`package.json` `scripts`, `composer.json` `scripts`), a `Makefile` or `justfile` target, a task runner file, or the project's `AGENTS.md`;
   - else the default of the language guide below, and only when the evidence that guide names is present (a config file, a lockfile, a dependency entry);
   - else nothing: record the check as `not_run: no <kind> configured in this project`.
3. Never install a tool, never change a tool's configuration to make a check pass, and never put a tool on PATH for the run. A configured tool that is missing on the machine is `blocked: <tool> not installed`, not a pass.
4. The kinds, in this order: format, lint, type check, test, build. Record each run as kind, command, source, exit code and the counts it prints. A test run that selected no test (`0 passed`, `no tests ran`, `[no test files]`) is `not_run`.

| Detected language | Guide |
| --- | --- |
| rust | languages/rust.md |
| python | languages/python.md |
| javascript | languages/javascript-typescript.md |
| typescript | languages/javascript-typescript.md |
| go | languages/go.md |
| java | languages/java-kotlin.md |
| kotlin | languages/java-kotlin.md |
| ruby | languages/ruby.md |
| php | languages/php.md |
| c | languages/c-cpp.md |
| cpp | languages/c-cpp.md |
| csharp | languages/csharp.md |
| swift | languages/swift.md |

The guides live under `references/languages/` of this skill. A language the detector does not know gets no guide: use the project's own declarations only and say that no guide exists.

## Tracing: requirement, criterion, check, test

- Each acceptance criterion (`AC-<n>`) of the SPEC names the requirements it covers (`(REQ-…)`) and one `verify:` command, or a `manual:` or `unknown:` reason. `moai spec check` refuses an AC without one.
- A test written for a criterion names it where `grep` finds it: in the test's name, or in a comment on the line above the test, as `<SPEC-ID> AC-<n>`. The ACC `@cc:proves` marker replaces that comment once its own SPEC lands.
- At VERIFY, run every `verify:` command of the SPEC and record its exit code next to the AC. A criterion whose command did not run is `not_run`, never `met`.
- A changed test that drops an assertion, gains a skip (`#[ignore]`, `@pytest.mark.skip`, `t.Skip`, `.skip(`), or is deleted is a finding: name it and the AC it proved.

## Before adding a dependency

Package names suggested by a model can be invented: in one study 5.2% of the packages that commercial models suggested, and 21.7% of those of open models, did not exist (Spracklen et al., USENIX Security 2025). An invented name can later be registered by someone else.

1. Check that the exact name exists in the registry the project already uses, with the command of the language guide, and keep its output line as evidence. Without network access, the check is `not_run` and the dependency is not added.
2. Use the name as the package's own documentation or repository spells it. A name one letter away from a popular package, or a package whose repository link does not match, is a stop and a question.
3. Add it with the package manager, so the lockfile records it; never edit a lockfile by hand. Show the lockfile's new entry.
4. A new dependency is a decision: it is in the SPEC's scope or asked for, and it is named in the change's report.
