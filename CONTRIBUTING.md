# Contributing to MoAI ADK

Thank you for your interest in MoAI ADK. Anyone can contribute: bug reports, questions, documentation fixes and code are all welcome.

This guide is for the Rust release candidate on the `v2026.11.0` branch. The Go 3.x generation lives on `main` and has its own guide there; please base work for the release candidate on `v2026.11.0`.

By participating you agree to follow the [Code of Conduct](CODE_OF_CONDUCT.md).

## Before you start

1. **Search first.** Look through [open Issues](https://github.com/modu-ai/moai-adk/issues) and [Discussions](https://github.com/modu-ai/moai-adk/discussions) for the same topic.
2. **Open an Issue for anything non-trivial.** Use one of the Issue forms (bug fix, feature, investigation). For a larger change, agree on the scope in the Issue before writing code; a pull request that changes a lot without a prior discussion is hard to review.
3. **Questions go to Discussions**, not Issues. See [SUPPORT.md](SUPPORT.md).
4. **Security problems go to the private channel** in [SECURITY.md](SECURITY.md), never to a public Issue.

### Good first issues

Issues labelled [`good first issue`](https://github.com/modu-ai/moai-adk/labels/good%20first%20issue) are chosen to need little context about the internals. [`help wanted`](https://github.com/modu-ai/moai-adk/labels/help%20wanted) marks work where outside help is especially welcome. Leave a comment on the Issue before you start so two people do not do the same work.

## Development setup

You need Git, [rustup](https://rustup.rs/) and Python 3. The Rust toolchain is pinned in [`rust-toolchain.toml`](rust-toolchain.toml), so rustup selects the right version for you.

```sh
git clone https://github.com/<your-account>/moai-adk.git
cd moai-adk
git remote add upstream https://github.com/modu-ai/moai-adk.git
git fetch upstream
git checkout -b <your-branch> upstream/v2026.11.0
cargo fetch --locked
```

`cargo fetch --locked` downloads the pinned dependencies once; the checks below then run offline.

## The check set

Run all five before you push. CI (`rc-ci`) runs the same set.

```sh
cargo test --offline --locked
cargo clippy --offline --locked --all-targets -- -D warnings
cargo fmt --check
PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s tests -p '*_test.py'
git diff --check
```

Some Python tests import PyYAML; CI installs `PyYAML==6.0.2` (`python3 -m pip install "PyYAML==6.0.2"`).

This exact sequence (`cargo fetch --locked`, then the five commands) was run on 2026-11-11 in a fresh clone of `v2026.11.0` at commit `1aad1ce01a`, on macOS arm64 with an empty Cargo home and PyYAML 6.0.3 already installed. All six commands exited 0: 974 Rust tests passed and 208 Python tests passed. On other platforms, rely on CI or report what you see.

Rules that apply to every change:

- **Do not report a check you did not run as passed.** If you could not run something, say so in the pull request and give the reason.
- **Add or update tests** with code changes. When you add or change a limit, test both sides of the edge.
- **Keep secrets out** of code, Issues, pull requests and logs: write the variable name, never its value.
- **Comments and instruction files are in English.** User-facing text can follow the language of the document it belongs to.

## Commits and pull requests

- **Target branch:** `v2026.11.0`.
- **Link the Issue with `Refs #<number>`** in the pull request description. Please do not use closing keywords such as `Closes` or `Fixes`; maintainers close Issues after integration.
- **Commit messages:** a short imperative summary line, Conventional Commits style is welcome (`fix: …`, `docs: …`, `feat(scope): …`), followed by a body that explains why.
- **Keep the pull request focused** on one Issue. Unrelated cleanups belong in their own pull request.
- **Describe what you verified:** the commands you ran and their results.
- **Allow edits from maintainers** on your pull request. The integration branch must be up to date before a merge, and this lets a maintainer update your branch for you.

You do **not** need to use the internal pull request template, claim comments, commit trailers or the `## Self-Audit` section. Those belong to the maintainers' agent workflow, and maintainers fill them in for external contributions (see below).

## How an external pull request is reviewed and merged

MoAI ADK is developed with coding-agent sessions that follow the same workflow the kit provides. Your pull request joins that flow like this:

1. **Claim on your behalf.** A maintainer records a claim for your work on the linked Issue, so that no agent session starts the same work.
2. **CI approval.** For a first-time contributor, GitHub holds the workflow run of a pull request from a fork until a maintainer approves it. After approval, `rc-ci` runs the check set. It runs with read-only permissions and no secrets.
3. **Independent audit.** A reviewer session that did not write the change audits the exact head commit against the Issue, the diff and the evidence. It posts a verdict on the pull request, either PASS or FAIL. A FAIL lists what to fix.
4. **Merge gate.** Only a maintainer posts the `moai/audit` commit status on the audited head and merges. The branch ruleset requires that status, and the branch must be up to date.
5. **New pushes need a new audit.** Any new commit after the verdict invalidates it, and the audit runs again on the new head.

Expect review comments in English or Korean. A response-time target for first replies has not been set yet (**pending maintainer decision**).

## License of contributions

MoAI ADK is licensed under the [Apache License 2.0](LICENSE). There is no CLA and no DCO sign-off requirement. Under section 5 of the license, any contribution you intentionally submit for inclusion is provided under the same Apache-2.0 terms, unless you explicitly state otherwise.

## Contributing with a coding agent

You are welcome to use Claude Code, Codex or another agent. Please make sure that:

- the agent reads the repository's instruction files and the Issue before it changes anything;
- every check reported in the pull request was actually run, with its real output;
- you have read and understood the change you submit; you are responsible for it.

## Where to talk

- Questions and ideas: [GitHub Discussions](https://github.com/modu-ai/moai-adk/discussions)
- Bugs and tasks: [GitHub Issues](https://github.com/modu-ai/moai-adk/issues)
- Live chat: [Discord](https://discord.gg/Z7E7Mdc5aN). Decisions and anything that must be findable later go to an Issue or a Discussion.
