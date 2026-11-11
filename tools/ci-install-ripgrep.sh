#!/usr/bin/env bash
# Issue #2469 option A, authorized by user decision 6801393490.
# Source/pin update procedure: docs/ci-ripgrep-install.md.
set -euo pipefail

if [[ "${RUNNER_OS:-}" != Linux ]]; then
  echo "This installer is only for GitHub-hosted Linux runners" >&2
  exit 1
fi
if command -v rg >/dev/null 2>&1; then
  exec rg --version
fi

# Initial attempt plus three explicit retries. Disable nested per-file retries;
# each apt command has a deadline, including a grace period for termination.
attempt=0
for delay in 0 5 10 20; do
  attempt=$((attempt + 1))
  printf 'ripgrep apt attempt %s/4 (backoff %ss)\n' "$attempt" "$delay"
  if [[ "$delay" -gt 0 ]]; then sleep "$delay"; fi
  if sudo -n timeout --kill-after=5s 45s apt-get -o Acquire::Retries=0 --error-on=any update &&
     sudo -n timeout --kill-after=5s 45s apt-get -o Acquire::Retries=0 install --yes ripgrep &&
     command -v rg >/dev/null 2>&1 && rg --version; then
    exit 0
  fi
  echo "apt did not provide a working rg; bounded attempt finished" >&2
done

version=14.1.1
case "$(uname -m)" in
  x86_64)
    target=x86_64-unknown-linux-musl
    sha=4cf9f2741e6c465ffdb7c26f38056a59e2a2544b51f7cc128ef28337eeae4d8e
    ;;
  aarch64)
    target=aarch64-unknown-linux-gnu
    sha=c827481c4ff4ea10c9dc7a4022c8de5db34a5737cb74484d62eb94a95841ab2f
    ;;
  *) echo "No pinned ripgrep asset for this Linux architecture" >&2; exit 1 ;;
esac
: "${RUNNER_TEMP:?RUNNER_TEMP is required}" "${GITHUB_PATH:?GITHUB_PATH is required}"
name="ripgrep-$version-$target"
work="$(mktemp -d "$RUNNER_TEMP/moai-ripgrep.XXXXXX")"
url="https://github.com/BurntSushi/ripgrep/releases/download/$version/$name.tar.gz"
# No API, credentials, mutable latest URL, runtime pin override, or downloaded script.
printf 'ripgrep fallback download: %s\n' "$url"
curl --disable --fail --location --silent --show-error \
  --proto '=https' --proto-redir '=https' --tlsv1.2 \
  --connect-timeout 15 --max-time 60 --retry 2 --retry-delay 2 \
  --output "$work/archive.tar.gz" "$url"
# This barrier must precede archive extraction, execution and PATH publication.
(cd "$work" && printf '%s  archive.tar.gz\n' "$sha" | sha256sum --check --status)
printf 'ripgrep SHA256 verified: %s  %s.tar.gz\n' "$sha" "$name"
tar -xzf "$work/archive.tar.gz" -C "$work" "$name/rg"
if [[ ! -f "$work/$name/rg" || -L "$work/$name/rg" ]]; then
  echo "Pinned ripgrep archive did not contain a regular executable" >&2
  exit 1
fi
mkdir "$work/bin"
install -m 0755 "$work/$name/rg" "$work/bin/rg"
version_status=0
observed="$("$work/bin/rg" --version)" || version_status=$?
first_line="${observed%%$'\n'*}"
printf 'ripgrep executable first line: %s\n' "$first_line"
if [[ "$version_status" -ne 0 ]]; then
  echo "Pinned ripgrep version command failed" >&2
  exit "$version_status"
fi
if [[ "$first_line" != "ripgrep $version" && "$first_line" != "ripgrep $version "* ]]; then
  echo "Pinned ripgrep executable reported an unexpected version" >&2
  exit 1
fi
printf '%s\n' "$observed"
printf '%s\n' "$work/bin" >> "$GITHUB_PATH"
