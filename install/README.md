# MoAI-ADK installers (Rust edition)

`install.sh` (macOS, Linux) and `install.ps1` (Windows; also runs under `pwsh` on macOS and Linux) install the Rust edition from the same addresses the Go edition used:

```sh
curl -fsSL https://adk.mo.ai.kr/install.sh | sh
```

```powershell
irm https://adk.mo.ai.kr/install.ps1 | iex
```

The addresses do not change; at release time only the files behind them are replaced (a release step that a person confirms). Until then the site serves the Go edition's scripts.

## What the installers do

1. Detect the target. Only the three supported platforms are accepted (`SPEC-20261117-supported-platforms`): `aarch64-apple-darwin` (an x86_64 shell under Rosetta 2 on Apple silicon counts as arm64), `x86_64-unknown-linux-gnu` and `x86_64-pc-windows-msvc`. Any other platform, Intel macOS and Linux arm64 included, stops here with `Unsupported platform: <os> <arch>` and the list of supported platforms, exit 1, before anything is downloaded or written.
2. Find the version: `--version` (a leading `v` or `moai-adk-v` is accepted), or the newest `moai-adk-v<version>` release. The scripts read GitHub's REST list first: the 100 newest releases, with drafts and prereleases flagged. When that cannot be read (no network to the API, or the unauthenticated limit of 60 requests an hour is spent), they read the releases feed (`releases.atom`). The feed is not rate limited but holds only the 10 newest entries. Release candidates (flagged prerelease, or a `-` in the version) are taken only with `--channel rc`. `/releases/latest` is not used, because the Rust edition's final releases are published with `--latest=false` so the Go edition's updater keeps its latest (`docs/release-checklist.md`).
3. Download `moai-cli-<target>.tar.xz` (`.zip` on Windows) and its `.sha256`, and verify the digest. A missing, malformed or different checksum stops the install before anything outside the temporary folder is written. The Go scripts continued without a checksum. Every release source must be `https://` (`file://` only for the tests). The `.sha256` comes from the same release as the archive, so it catches a damaged download, not a forged release. Checking the release's signature and attestation is part of the release step that replaces the published files (`docs/release-attestation.md`).
4. Ask the downloaded binary where its state lives: `moai-cli install-inspect` reports `home_state_root`. The rule for that folder (`~/.moai` or the earlier `~/.moai-cli`) is in the binary, not in the scripts.
5. Place it at `<state home>/versions/<version>/bin/moai-cli`, with `bin/moai` as a hard link to it (the layout of `src/bin_names.rs`). A new version folder is staged beside its final name and renamed into place. If the version folder already exists, it is kept when it holds the same bytes and refused otherwise.
6. Link `moai-cli` and `moai` in the link folder (`--install-dir`, default `<state home>/path`). Only a link into a version folder the installer placed is replaced, and the replacement is made under a temporary name and renamed over the old link, so the command never goes missing. A file that is not a link (for example someone's shim), or a link to anything else (the Go edition's `moai`, say), is left as it is. On Windows the commands are copies, because a symbolic link needs extra rights there. A copy that matches an installed version is replaced on an upgrade.
7. Never edit PATH or shell profiles. When the link folder is not on PATH, print the line to add instead. Putting the Rust `moai` first on PATH, with a receipt and a way back, is the `install-path-*` step.
8. When the `moai` first on PATH is the Go edition, change nothing of it and print `moai-cli update` (a preview that writes nothing) as the next step.

## After the install

Both scripts end by printing the next steps. Make a project folder and enter it (`mkdir my-project && cd my-project`; in PowerShell `New-Item -ItemType Directory my-project; Set-Location my-project`). Start a session there with `moai cc` or `moai glm`; add `--name <name>` to name it. A session starts with no role; type `/moai agent` (a lane) or `/moai leader` (the coordination window) inside it to take one, which setting up a project does not need (`moai cc -a` and `moai cc -l` start a named lane or coordination window directly). In the session, run `/moai project` to set the project up.

## Conversation language

The installers ask which language the model should use when it talks with you. The answer is only the language of the conversation: everything the installer and `moai-cli` print stays English. The choice is, in order: `--language CODE` (`-language` in PowerShell), then the `MOAI_LANGUAGE` variable, then a numbered menu (`install.ps1` asks it only when both stdin and stdout are a terminal; `install.sh` follows the rule below) (1 English, 2 한국어, 3 日本語, 4 中文, 5 other: type a code; a wrong answer is asked again, three answers at most, then the locale's language is used), then the locale (`LC_ALL`, `LC_MESSAGES`, `LANG`; PowerShell also `Get-Culture`), then `en`. `install.sh` asks the menu only when stdout is a terminal, and takes the answer from standard input when that is a terminal, otherwise from the session's terminal (`/dev/tty`) when it can be opened, and otherwise nobody is asked, so `curl … | sh` at a terminal still shows the menu. At the menu, press Enter to take the default; Ctrl-D ends the input in `install.sh`, but PowerShell does not treat it as the end of input on every host. A code is two or three letters and optional `-` groups of two to eight letters or digits (`ko`, `pt-BR`); anything else is refused before anything is downloaded or written. A code is at most 35 characters and is passed on as typed.

When the commands are linked, the installer runs `moai-cli language-set --home <code>` once. It writes `config/language.yaml` in the state folder only when that file is not there yet (a second install never changes an earlier choice), and writes nothing when the HOME has no MoAI state folder. An older pinned version that has no `language-set` only costs one warning line; the install still succeeds. `moai-cli language-set` accepts any code of that shape and writes it as given; `moai-cli language-setup` offers the stored value as the conversation default when it is `en`, `ko`, `ja` or `zh` and otherwise falls through to the next source, so a code such as `pt-BR` is kept for a later reader without changing what `language-setup` does.

## The MoAI skill and agents

After the commands are linked and the language is recorded, both installers offer the MoAI skill and agents to the Claude Code they find. When they find none, they say so, ask nothing and name `moai update` for later. Otherwise they run `moai-cli update --json` from the version folder, which writes nothing. When everything is already installed and up to date they say that. When there is something to install, they print the skill and agents folders, say that MoAI's sections in the host's settings files are written too, and ask:

```text
Install them now? [y/N]
```

Only a yes installs, by confirming that preview's plan with `moai-cli update --confirm`: `install.sh` takes `y`, `Y`, `yes`, `Yes` or `YES`, and `install.ps1` takes `y` or `yes` in any case. With `--yes` (`-yes` in PowerShell) the preview is still shown and the plan is confirmed without a question. Without it the question is asked only where a person can answer: `install.sh` asks when stdout is a terminal and takes the answer from standard input when that is a terminal, otherwise from `/dev/tty` when it can be opened; `install.ps1` asks only when neither standard input nor standard output is redirected. In every other case nothing of the skill or the agents is written, and the message names `moai update` (or `--yes`). A failed preview or install prints a warning and leaves the installed command in place.

## Options, compared with the Go scripts

| Option or variable | Go `install.sh` / `install.ps1` | Rust installers |
| --- | --- | --- |
| `--version V` (`-version` in PowerShell) | fixed version | fixed version (tag `moai-adk-vV`). Text other than a version is refused |
| `--install-dir D` (`-install-dir`) | puts the `moai` binary in D | puts the `moai` and `moai-cli` commands in D. The version folder stays the original |
| `MOAI_INSTALL_DIR` | PowerShell only | both scripts, same meaning as `--install-dir` |
| `--channel stable\|rc` | — | new: `rc` also takes release candidates |
| `--language CODE` (`-language`), `MOAI_LANGUAGE` | — | new: the conversation language, see above |
| `--yes` (`-yes`) | — | new: installs the MoAI skill and agents without asking; without it and without anyone to ask, they are not installed |
| `-h`, `--help` (`-help`) | help | help |
| an unknown option | exit 1 | exit 1 |
| default location | `GOBIN`, `GOPATH/bin` or `~/.local/bin` (`%LOCALAPPDATA%\Programs\moai`) | `<state home>/versions/<v>/bin`, commands in `<state home>/path` |
| PATH | sh: a hint. ps1: writes the User PATH (appended at the end) | neither script edits PATH; both print a hint |
| checksum | skipped when `checksums.txt` is missing; ps1 continues on any failure | required; any failure stops the install |
| `MOAI_RELEASE_BASE`, `MOAI_RELEASE_FEED`, `MOAI_RELEASE_API` | — | where releases, the feed and the REST list are read from: `https://` only, or `file://` (the tests use a `file://` release) |

## Keeping the Go installers

When the site switches, the Go repository's `docs-site/static/install.sh` and `install.ps1` are renamed to `install-go.sh` and `install-go.ps1` beside the new files. The Go edition then stays installable from `https://adk.mo.ai.kr/install-go.sh` and `install-go.ps1`, and the published addresses install the Rust edition. This is a site change and waits for the user's confirmation, like the release itself.

## Tests

`tests/install_script_test.py` runs both scripts against a temporary HOME and a fake release served from `file://`. The fake contains a feed, the archives in the cargo-dist shape and a fake `moai-cli` that answers `install-inspect` and `--version`. It covers the latest stable and `rc` versions, a pinned version, refused checksums (wrong, missing, malformed) with no files written, versions that are not versions, a Go `moai` left byte for byte, a second run that writes nothing, a version folder with other bytes refused, the state folder taken from the binary, `--install-dir`, and a non-link command file left alone. For the conversation language it covers the flag, the variable and the locale order, the edges of the code pattern, the menu through a pseudo-terminal (every number, a typed code, the three-answer limit, an empty answer, the end of input), a fake `moai-cli` that records the exact `language-set --home <code>` call (`pt-BR` by flag, variable and menu choice 5, also against the real `moai-cli` when one is built), one without `language-set`, and a second run that writes nothing. `shellcheck` runs when it is installed. The PowerShell tests run when `pwsh` is installed. Windows itself (zip, `.exe`, copies) is `not_run` there.
