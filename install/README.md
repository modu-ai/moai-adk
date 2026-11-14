# MoAI-ADK installers (Rust edition)

`install.sh` (macOS, Linux) and `install.ps1` (Windows; also runs under `pwsh` on macOS and Linux) install the Rust edition from the same addresses the Go edition used (#2935, #2871 I1):

```sh
curl -fsSL https://adk.mo.ai.kr/install.sh | sh
```

```powershell
irm https://adk.mo.ai.kr/install.ps1 | iex
```

The addresses do not change; at release time only the files behind them are replaced (step I1-e of #2871, a user-confirmed step). Until then the site serves the Go edition's scripts.

## What the installers do

1. Detect the target (`aarch64`/`x86_64` × `apple-darwin`, `unknown-linux-gnu`, `pc-windows-msvc`).
2. Find the version: `--version` (a leading `v` or `moai-cli-v` is accepted), or the newest `moai-cli-v<version>` release. The scripts read GitHub's REST list first: the 100 newest releases, with drafts and prereleases flagged. When that cannot be read (no network to the API, or the unauthenticated limit of 60 requests an hour is spent), they read the releases feed (`releases.atom`). The feed is not rate limited but holds only the 10 newest entries (audit #2943 B1). Release candidates (flagged prerelease, or a `-` in the version) are taken only with `--channel rc`. `/releases/latest` is not used, because the Rust edition's final releases are published with `--latest=false` so the Go edition's updater keeps its latest (`docs/release-checklist.md`).
3. Download `moai-cli-<target>.tar.xz` (`.zip` on Windows) and its `.sha256`, and verify the digest. A missing, malformed or different checksum stops the install before anything outside the temporary folder is written. The Go scripts continued without a checksum. Every release source must be `https://` (`file://` only for the tests). The `.sha256` comes from the same release as the archive, so it catches a damaged download, not a forged release. Checking the release's signature and attestation is part of the release step I1-e (`docs/release-attestation.md`).
4. Ask the downloaded binary where its state lives: `moai-cli install-inspect` reports `home_state_root`. The rule for that folder (`~/.moai` or the earlier `~/.moai-cli`) is in the binary (#2871 I2), not in the scripts.
5. Place it at `<state home>/versions/<version>/bin/moai-cli`, with `bin/moai` as a hard link to it (the layout of `src/bin_names.rs`). A new version folder is staged beside its final name and renamed into place. If the version folder already exists, it is kept when it holds the same bytes and refused otherwise.
6. Link `moai-cli` and `moai` in the link folder (`--install-dir`, default `<state home>/path`). Only a link into a version folder the installer placed is replaced, and the replacement is made under a temporary name and renamed over the old link, so the command never goes missing. A file that is not a link (for example someone's shim), or a link to anything else (the Go edition's `moai`, say), is left as it is. On Windows the commands are copies, because a symbolic link needs extra rights there. A copy that matches an installed version is replaced on an upgrade.
7. Never edit PATH or shell profiles. When the link folder is not on PATH, print the line to add instead. Putting the Rust `moai` first on PATH, with a receipt and a way back, is the `install-path-*` step (#2871 I1-b).
8. When the `moai` first on PATH is the Go edition, change nothing of it and print `moai-cli update` (a preview that writes nothing) as the next step.

## Options, compared with the Go scripts

| Option or variable | Go `install.sh` / `install.ps1` | Rust installers |
| --- | --- | --- |
| `--version V` (`-version` in PowerShell) | fixed version | fixed version (tag `moai-cli-vV`). Text other than a version is refused |
| `--install-dir D` (`-install-dir`) | puts the `moai` binary in D | puts the `moai` and `moai-cli` commands in D. The version folder stays the original |
| `MOAI_INSTALL_DIR` | PowerShell only | both scripts, same meaning as `--install-dir` |
| `--channel stable\|rc` | — | new: `rc` also takes release candidates |
| `-h`, `--help` (`-help`) | help | help |
| an unknown option | exit 1 | exit 1 |
| default location | `GOBIN`, `GOPATH/bin` or `~/.local/bin` (`%LOCALAPPDATA%\Programs\moai`) | `<state home>/versions/<v>/bin`, commands in `<state home>/path` |
| PATH | sh: a hint. ps1: writes the User PATH (appended at the end) | neither script edits PATH; both print a hint |
| checksum | skipped when `checksums.txt` is missing; ps1 continues on any failure | required; any failure stops the install |
| `MOAI_RELEASE_BASE`, `MOAI_RELEASE_FEED`, `MOAI_RELEASE_API` | — | where releases, the feed and the REST list are read from: `https://` only, or `file://` (the tests use a `file://` release) |

## Keeping the Go installers

When the site switches (I1-e), the Go repository's `docs-site/static/install.sh` and `install.ps1` are renamed to `install-go.sh` and `install-go.ps1` beside the new files. The Go edition then stays installable from `https://adk.mo.ai.kr/install-go.sh` and `install-go.ps1`, and the published addresses install the Rust edition. This is a site change and waits for the user's confirmation, like the release itself.

## Tests

`tests/install_script_test.py` runs both scripts against a temporary HOME and a fake release served from `file://`. The fake contains a feed, the archives in the cargo-dist shape and a fake `moai-cli` that answers `install-inspect` and `--version`. It covers the latest stable and `rc` versions, a pinned version, refused checksums (wrong, missing, malformed) with no files written, versions that are not versions, a Go `moai` left byte for byte, a second run that writes nothing, a version folder with other bytes refused, the state folder taken from the binary, `--install-dir`, and a non-link command file left alone. `shellcheck` runs when it is installed. The PowerShell tests run when `pwsh` is installed. Windows itself (zip, `.exe`, copies) is `not_run` there.
