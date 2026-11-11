# Linux ripgrep installer (#2469)

**Status: option A approved by user decision6801393490 (2026-11-11T16:32Z).** The leader authorized a PR despite the five-open-PR limit, with independent review by pr-6 before integration. Actual Linux CI evidence is recorded against the exact PR head; the local preparation alone does not establish it. Other workflows, rulesets and secrets remain outside scope.

## Behavior

The existing workflow checks `command -v rg` first. A working preinstalled rg remains untouched. On Linux without rg, the proposed step executes `tools/ci-install-ripgrep.sh`. The helper also checks an existing rg and returns its actual `--version` status; it does not replace a broken preinstalled executable silently. macOS retains the original Homebrew branch and version check; the Windows build job is unchanged.

APT keeps the update/install path. The helper makes an initial attempt plus **three retries**, with 5/10/20-second backoff. Each apt command runs under GNU `timeout`, with a 45-second deadline and 5-second termination grace; `sudo -n` cannot prompt. `--error-on=any` makes a transient update error fail. `Acquire::Retries=0` prevents nested per-file retry multiplication: upstream documents that option as retries for failed files, not a bound or backoff for the whole update/install pair. Explicit outer retries are adopted; relying on `Acquire::Retries=3` alone to guarantee three pair-level backoffs is rejected. Command-line options do not write an apt configuration file. A timeout can interrupt package installation on the disposable runner; the helper does not repair package state or weaken security settings.

If apt never provides a working rg, select the fixed official ripgrep **14.1.1** release for `uname -m` x86_64 or aarch64. This is a reproducible pin, not a claim that it is the latest release. An unsupported architecture fails without downloading. Curl uses HTTPS-only transfer/redirect protocols, ordinary certificate verification, disabled curlrc, connect/max-time limits and at most two retries. It does not use GitHub REST, an API token, a mutable latest URL or a downloaded installation script.

The archive is written under a unique `RUNNER_TEMP/moai-ripgrep.XXXXXX` directory. **The embedded SHA256 must match before tar extraction, executable invocation, or GITHUB_PATH publication.** Only the exact `<asset-stem>/rg` member is extracted. It must be a regular non-symlink file. Copy it to that temporary directory's bin, verify that the first line is exactly `ripgrep 14.1.1` or starts with that literal token followed by an ASCII space, then append the bin path to GITHUB_PATH for subsequent steps. The Linux workflow uses exec so its step result is the helper's verified result; it does not attempt a redundant current-shell PATH lookup before GitHub applies GITHUB_PATH. Artifacts remain in runner temp until the disposable job is cleaned up.

## Fixed assets and trust boundary

Both archives and their official `.tar.gz.sha256` files were actually downloaded on 2026-11-11 UTC (2026-11-12 KST). Independent local SHA256 recomputation matched the published checksum:

| Target | Archive SHA256 |
| --- | --- |
| `ripgrep-14.1.1-x86_64-unknown-linux-musl.tar.gz` | `4cf9f2741e6c465ffdb7c26f38056a59e2a2544b51f7cc128ef28337eeae4d8e` |
| `ripgrep-14.1.1-aarch64-unknown-linux-gnu.tar.gz` | `c827481c4ff4ea10c9dc7a4022c8de5db34a5737cb74484d62eb94a95841ab2f` |

Base URL: `https://github.com/BurntSushi/ripgrep/releases/download/14.1.1/` (GitHub may redirect release downloads to its asset CDN). The expected hashes are reviewed repository literals, never supplied by an environment variable or downloaded alongside the runtime payload. Integrity against those literals is checked at runtime. The initial published checksum and archive share an upstream release trust domain; the checksum is not an independent publisher signature or a promise against a compromised upstream. This adds the official release distribution source only and does not change secret/permission policy.

ELF headers were inspected: class2 and e_machine62 for x86_64, class2 and e_machine183 for aarch64. The real Linux executables were **not run on the development Mac**. These local preparation measurements do not establish compatibility with actual Ubuntu24.04 x64/arm64 runners. Cite exact-head hosted CI separately; the first PR run is the live Linux validation.

## Official sources and adopt/reject record

Checked 2026-11-11 UTC / 2026-11-12 KST:

- [GitHub x64 Ubuntu24.04 software list](https://raw.githubusercontent.com/actions/runner-images/main/images/ubuntu/Ubuntu2404-Readme.md), [Arm image explanation](https://raw.githubusercontent.com/actions/partner-runner-images/main/images/arm-ubuntu-24-image.md), and [generated Arm list](https://raw.githubusercontent.com/actions/partner-runner-images/main/images/Ubuntu2404-Readme.md): ripgrep is not listed. These live documents do not prove absence on every actual job image. Adopt existence/version probing; reject assuming rg is guaranteed preinstalled and deleting installation. Document image versions observed were 20261101.360.1 and 0.0.1, not an observed live job inventory.
- Ubuntu Noble [apt.conf](https://manpages.ubuntu.com/manpages/noble/man5/apt.conf.5.html) and [apt-get](https://manpages.ubuntu.com/manpages/noble/man8/apt-get.8.html), served version2.8.4ubuntu0.1, were read with the web tool. [Debian apt.conf](https://manpages.debian.org/bookworm/apt/apt.conf.5.en.html) and [apt-get](https://manpages.debian.org/bookworm/apt/apt-get.8.en.html) were also downloaded/read. Adopt explicit bounded outer backoff and transient update failure detection; reject treating per-file retries as a total wall-clock/pair-level guarantee. Local Ubuntu HTML download failed DNS; no original HTML hash is claimed for those pages.
- [Official ripgrep14.1.1 release](https://github.com/BurntSushi/ripgrep/releases/tag/14.1.1) and [asset listing](https://github.com/BurntSushi/ripgrep/releases/expanded_assets/14.1.1): both target archives have separate sha256 files. Adopt fixed assets and embedded hashes after actual independent recomputation; reject unpinned latest downloads, unchecked extraction/execution, runtime checksum overrides and third-party installer actions.

Successfully downloaded source snapshots (hashes are of the actual bytes, not a claimed read of a link):

| Source | Retrieval UTC | SHA256 |
| --- | --- | --- |
| [ubuntu-x64](https://raw.githubusercontent.com/actions/runner-images/main/images/ubuntu/Ubuntu2404-Readme.md) | 2026-11-11T15:25:23.294760+00:00 | `5993072a61c810a3bf5057bba6322e576a5d83307b87afc1c8600e05b811b124` |
| [ubuntu-arm](https://raw.githubusercontent.com/actions/partner-runner-images/main/images/Ubuntu2404-Readme.md) | 2026-11-11T15:25:23.505816+00:00 | `99a4184ca62d3d40db372bcd6a9be6b4bbcf3d8e14b29c54d445edbd6517996c` |
| [apt-conf](https://manpages.debian.org/bookworm/apt/apt.conf.5.en.html) | 2026-11-11T15:25:24.774073+00:00 | `eadaf736ffb202ccdd00334b263c644ef02869197e4d16022a5dcb05d77460b9` |
| [apt-get](https://manpages.debian.org/bookworm/apt/apt-get.8.en.html) | 2026-11-11T15:25:26.222190+00:00 | `ca82c18bc89cda82a42b18724e353bdd4a51cb9b0c7f3ab8f7ce7f4c40ed16fd` |
| [rg-assets](https://github.com/BurntSushi/ripgrep/releases/expanded_assets/14.1.1) | 2026-11-11T15:25:26.639529+00:00 | `f126583b317352fda00e88256d86238f0086873fc7366a1404e28738672d4fab` |

Full archive/checksum byte receipts and the failed Ubuntu local-download record are retained in the author's `target/agent-20/2469/`; that evidence directory is not committed. The Issue design comment6799971011 records the preparation source review; decision6801393490 subsequently authorized option A and PR activation.

## Pin update procedure (after approval)

1. Read the selected official release and asset listing; confirm the x86_64 Linux musl and aarch64 Linux GNU asset names and supported runtime requirements.
2. Download each archive and its published checksum over HTTPS into a temporary directory. Independently recompute both archive SHA256 values and compare the complete digests; record URL, retrieval UTC and byte hashes. Do not copy a checksum from an unaudited proposal.
3. Update the helper's version, both targets and both digest literals together with this table. There is no runtime URL/hash override or automatic latest resolution.
4. Run the offline behavioral suite and checksum-guard mutation. On disposable Linux x64/arm64 runners, inject an apt acquisition failure and verify the real pinned fallback executable and version. Check checksum mismatch refusal before extraction/execution/export. Preserve macOS/Windows scope checks, then complete the required independent review/CI process. This procedure does not authorize changes outside the approved Issue scope.

## Verification and limits

Preparation commands:

```sh
bash -n tools/ci-install-ripgrep.sh
PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s tests -p 'ci_install_ripgrep_test.py'
PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s tests -p '*_test.py'
git diff --check
git diff --cached --check
```

The fixture launches the actual shell helper through isolated sudo/apt/curl/sleep/timeout commands and temporary runner paths. Python hashlib and tarfile process real synthetic archive bytes in deterministic command fixtures. Only bash is native; mktemp/tar/mkdir/install are fixture commands with checked arguments and confined paths. No host tar, gzip, or file utility is resolved from the fixture PATH. This tests helper ordering and publication, not the host tar implementation. **Only the fixture's private copy** substitutes a digest matching its synthetic archive; production has no such environment/argument bypass. A deliberate mismatched checksum must leave extraction, execution and GITHUB_PATH untouched. Mutation removes that barrier in a temporary source copy and must make the regression fail at runtime. Existing/apt-success paths must avoid curl, retries must be finite, download failure and unsupported architecture must refuse, and the original macOS workflow branch must still execute against a fake Homebrew.

AC-1 preparation covers injected command failure/timeout and a synthetic rg --version, plus separately observed checksums of real release assets. It is not the full live Linux AC: an actual bad mirror/failing apt acquisition with the official Linux ELF execution is not_run. Local preparation did not run hosted CI or change installed apt/user configuration; hosted results must be read separately from the current PR. AC-2 additionally compares the prepared workflow against its base: all bytes except the two Linux apt lines replaced by the helper exec must match, including Windows and macOS jobs/permissions. AC-3 is the source/pin/update record above. No Rust source changes are included; The preparation revision8fa3325 did not run the Rust full suite/clippy; cite the current PR Self-Audit for the final pre-PR check set. ShellCheck is not installed locally; its result is not_run, not PASS.

## CI evidence logs

The helper emits `ripgrep apt attempt N/4` before each bounded attempt, `ripgrep fallback download: <fixed URL>` before curl, and `ripgrep SHA256 verified: <digest> <asset>` only after the actual checksum command succeeds. The executable first line is logged before version validation, including on failure. Exact `ripgrep 14.1.1` and an ASCII-space-delimited revision suffix are accepted; `14.1.10`, `14x1x1`, `14.1.1-dev`, other prefixes and a matching later line are rejected. A nonzero version command exit remains a failure. The full successful version output follows extraction and execution. A checksum failure must not emit the success line. Use these lines plus the Linux job/step result to distinguish apt recovery from fallback; an install-step success alone does not prove the fallback ran. The PR ci line records the run/head, target and observed path, with job log receipts.

## First live Linux correction

The first PR run43061759020 reached the pinned fallback on both Linux targets and verified each archive SHA, then failed an overly exact version comparison. The synthetic fixture had omitted the official revision suffix. The corrected literal first-line boundary accepts official `ripgrep 14.1.1 (rev ...)` metadata without accepting `14.1.10`; the next exact-head CI log records the actual official executable output. Local boundary fixtures reach the version gate after a valid checksum, and command-failure/first-line-only cases keep GITHUB_PATH unchanged. This correction does not change URL/hash/architecture/TLS/extraction policy.

## Linux fixture portability correction

Run43064319196 exposed a harness dependency: GNU tar invokes external gzip, absent from the isolated PATH, while macOS tar did not. Its Check set log reported `tar (child): gzip: Cannot exec: No such file or directory` and exit2 after checksum verification. The fixture now implements file commands with Python standard-library operations and records their exact argv. A permanent regression exercises fallback with no gzip or host file utilities present and checks the installed executable mode. Production helper behavior remains unchanged. The old fixture fails and the corrected fixture passes on the lane-owned Colima Linux VM; hosted CI is recorded separately in the PR.
