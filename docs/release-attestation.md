# MoAI 릴리스 attestation과 macOS quarantine 안내 (E4 #2196)

이 문서는 `moai-cli` 릴리스 산출물의 출처를 증명하는 방법(GitHub artifact attestation)과, 사용자가 받은 파일을 확인하고 macOS에서 여는 방법을 정한다. 결정 근거는 decisions 15:5x(04-A "GitHub artifact attestation만")와 01:1x(notarization 없음)다.

**상태(2026-11-09):** 설계와 로컬 dry-run까지 했다. 실제 attestation, 릴리스, tag, 게시용 workflow는 하나도 만들지 않았다. 아래 "운영자 승인이 필요한 일"을 승인받아야 실행할 수 있다.

## 1. checksum은 서명을 대체하지 않는다 (REQ-E4-01)

- `sha256.sum`과 archive별 `.sha256`은 **무결성 재확인**일 뿐이다. 내려받는 동안 파일이 깨지지 않았는지만 알려 준다.
- checksum 파일은 산출물과 같은 곳에서 함께 받는다. 그래서 공격자가 산출물을 바꿀 수 있다면 checksum도 같이 바꿀 수 있다. checksum이 맞는다고 해서 "누가, 어떤 소스와 workflow로 빌드했는가"가 증명되지는 않는다.
- 빌드 출처(provenance)는 attestation만 증명한다. 이 저장소의 GitHub Actions workflow가 어떤 커밋에서 그 파일을 만들었는지를 Sigstore 인증서로 서명해 남기는 방식이다.
- 그래서 이 프로젝트의 문서는 어디에서도 "checksum으로 서명을 확인한다"고 쓰지 않는다. 확인 절차는 항상 `gh attestation verify`이고, checksum은 보조로만 쓴다.

## 2. 조사 결과와 채택·기각 (원문 확인 2026-11-09T12:25Z)

| 항목 | 원문에서 확인한 사실 | 판정 |
|---|---|---|
| `actions/attest` | 최신 릴리스 `v4.2.2`(2026-08-04, 커밋 `1e69f48acb82d1966a394da916b4c1698aa569d6`). 필요한 권한은 `id-token: write`, `attestations: write`. 공개 저장소는 Sigstore public-good 인스턴스로 서명한다. 입력은 `subject-path`·`subject-digest`·`subject-checksums`(shasum 형식) 중 하나이고, 한 번에 최대 1024개 subject를 받는다. GitHub Enterprise Server는 지원하지 않는다 | **채택**(cargo-dist가 생성하는 단계가 이 액션을 쓴다) |
| `actions/attest-build-provenance` | v4부터 `actions/attest`의 래퍼다. 새 구현은 `actions/attest`를 쓰라고 안내한다 | **기각**(래퍼를 직접 쓰지 않음) |
| `gh attestation verify` | 기본으로 SLSA provenance v1(`https://slsa.dev/provenance/v1`)을 요구한다. 기본 OIDC issuer는 `https://token.actions.githubusercontent.com`이다. `--repo`/`--owner`, `--signer-workflow`, `--source-ref`, `--deny-self-hosted-runners`, `--bundle`(오프라인)을 지원한다 | **채택**(사용자 확인 절차) |
| cargo-dist 0.34.0 `github-attestations`·`github-attestations-phase` | 공식 설정 문서는 두 옵션을 모두 **experimental**로 표시한다. `ci = ["github"]`일 때 생성 workflow에 `actions/attest@v4` 단계를 넣는다. 기본 위치는 `build-local-artifacts` 단계이고, `github-attestations-phase = "host"`이면 게시 단계로 옮긴다 | **채택 제안**(experimental 상태와 host 단계 설정을 함께 운영자 승인 대상으로 둠) |
| 직접 쓴 attestation workflow | 같은 일을 cargo-dist 생성 workflow 밖에서 따로 하는 방식 | **기각**(dist 설정과 workflow가 어긋날 위험. dist가 생성 workflow를 관리한다) |
| Apple notarization | macOS는 기본적으로 서명·notarization된 소프트웨어를 요구한다. 그렇지 않은 앱은 "개인정보 보호 및 보안 → 그래도 열기"로 열 수 있다 | **해당 없음**(decisions 01:1x: notarization 없음). 아래 4절 안내로 대신한다 |

출처(받은 파일의 sha256):

| 출처 | sha256 |
|---|---|
| https://raw.githubusercontent.com/actions/attest/main/README.md | `126485fd4507e05686dd46b61fdb0f76c8ec69e90f6e52289e2054ea36ee1354` |
| https://raw.githubusercontent.com/actions/attest-build-provenance/main/README.md | `d806124075bfbbccb47d447770efa204425fcfbcbe836e35089cb5683afa2b98` |
| https://cli.github.com/manual/gh_attestation_verify | `c43bdc359391e634ac1158f6079c3996a84b7a94ee493f55766b71f11971d714` |
| https://raw.githubusercontent.com/axodotdev/cargo-dist/main/book/src/reference/config.md | `42a68473c1eab68dfff24b02290977bb6907ffba8f869d38404da8bc009394a8` |
| cargo-dist v0.34.0 `cargo-dist/templates/ci/github/release.yml.j2`(tag 커밋 `294c1b25d372d86e74cafead8469b9f36d02da29`) | `1783c4e62b109ce40b17ed0b801f34c97646603e377b10a0fcb0ff7a75794d55` |
| https://support.apple.com/en-us/102445 | `78ee1279889dfac1eaa7dc3271ac7ba9095ba2ece2faf26d93b4e43a2345b38f` |
| https://support.apple.com/guide/security/gatekeeper-and-runtime-protection-sec5599b66df/web | `95587f80b642293775837dc4180da12bfe3d0d2b38bd3cefae7397e26deea6d3` |

F1·F2·N1 수리에서 원문을 다시 받았다(2026-11-09T12:58Z, HTTP 200). cargo-dist 소스는 v0.34.0의 커밋 `294c1b25d372d86e74cafead8469b9f36d02da29`로 고정했다. 설정 문서와 workflow template의 hash는 위 표와 동일했고, 추가 근거는 다음과 같다. 이 대조는 실제 릴리스·attestation 실행을 포함하지 않는다.

| 추가 출처 | sha256 |
|---|---|
| [cargo-dist 0.34.0 릴리스 명령 생성부](https://github.com/axodotdev/cargo-dist/blob/294c1b25d372d86e74cafead8469b9f36d02da29/cargo-dist/src/backend/ci/github.rs#L502-L528) | `80254ca74cf5844aeb847cecaa567363e37ddfb76549f547244303bef83a7414` |
| [GitHub REST Release 생성·수정의 `make_latest`](https://docs.github.com/en/rest/releases/releases#create-a-release) | `a9e5c962fff1357844e7fa5df39f3ac5f26fa4a0706b82eed730bcdd4761fb86` |
| [GitHub CLI `gh release edit`](https://cli.github.com/manual/gh_release_edit) | `a185a01c5ff0bf392ec0a4afafd7fc8d3bdfebeb198a100bcff6fd7a8af81c71` |

로컬 도구(5절 기존 측정): `gh version 2.92.0`, `cargo-dist 0.34.0`, macOS 27.0.1.

## 3. CI 단계 설계 (승인 전, 적용하지 않음)

### 3.1 바꿀 설정(제안)

`dist-workspace.toml`:

```toml
ci = ["github"]
github-attestations = true
github-attestations-phase = "host"
github-release = "host"
tag-namespace = "moai-cli"
pr-run-mode = "skip"
```

`Cargo.toml` `[package]`:

```toml
repository = "https://github.com/modu-ai/moai-adk"
```

그다음 `dist generate --mode ci`로 `.github/workflows/moai-cli-release.yml`을 생성한다.

### 3.2 설정마다 이유 (기존 생성 결과와 고정 소스 대조)

기존 측정은 `pr-run-mode`를 지정하지 않은 임시 사본에서 생성한 workflow다(`target/e4/release.variant2.yml`, 커밋하지 않음). 이번 수리의 `pr-run-mode = "skip"` 제안은 고정한 공식 설정 문서와 template의 조건문으로 대조했다. 이 추가 설정을 적용한 workflow 생성·실행은 not_run이다.

- **`tag-namespace = "moai-cli"`**
  - 기본 트리거는 `'**[0-9]+.[0-9]+.[0-9]+*'`이다. 이 패턴은 Go 릴리스 tag(`v3.1.3`, `go-v…`)에도 걸린다. 기본값 그대로면 Go 릴리스 tag를 push할 때 Rust 릴리스 workflow도 함께 돈다.
  - 이 설정을 넣으면 트리거가 `'moai-cli**[0-9]+.[0-9]+.[0-9]+*'`로 바뀌고, 파일 이름도 `moai-cli-release.yml`이 된다(실측).
  - 04-A 결정의 tag 접두사 `moai-cli-v`와 맞는다.
- **`github-attestations-phase = "host"`, `github-release = "host"`**
  - 기본(`build-local-artifacts`)이면 attest 단계가 빌드 job 안에 있다. 그래서 수동 dry-run이나 빌드만 하는 실행에서도 실제 attestation이 저장소에 기록될 수 있다.
  - host로 옮기면 attest 단계가 `host` job 안의 `Create GitHub Release` 바로 앞에 놓인다. 이 job은 `needs.plan.outputs.publishing == 'true'`일 때만 돈다(실측).
  - 그래서 PR에서 도는 `plan`은 attestation을 만들지 않는다.
  - 문서는 두 값을 같이 설정하라고 한다("Make sure `github-release = "host"` is also set").
- **PR 실행과 권한:** 기본 `pr-run-mode = "plan"`인 생성 workflow는 브랜치·경로 필터 없는 `pull_request` 이벤트에서도 `plan` job으로 `dist plan`을 실행하도록 되어 있다. 최상위 `permissions`는 `contents: write`이고, 별도 권한 선언이 없는 `plan` job도 이 요청 권한을 상속한다. 실제 토큰 권한은 fork·저장소 정책에 따라 낮아질 수 있지만, 생성 YAML의 쓰기 권한 요청 자체는 남는다. `host` job에는 `attestations: write`·`id-token: write`·`contents: write`가 명시된다.
  - 제안: `pr-run-mode = "skip"`. 공식 template은 이 값일 때 `pull_request` 트리거를 생성하지 않는다. PR에서 attestation이 생성되지 않는다는 사실만으로 PR workflow 실행과 쓰기 권한 요청까지 없다고 판단하지 않는다. 적용 여부와 권한은 운영자 승인 항목 1이다.
- **subject:** host 단계는 `subject-path: artifacts/*`로 내려받은 산출물 전부를 attest한다. 로컬 `dist build --artifacts=all`이 만든 산출물(아래 5절)이 그 대상이다.
- **prerelease와 정식판의 latest:** dist는 버전에 접미사(예: `-rc.1`)가 있으면 GitHub Release를 prerelease로 만든다. GitHub API는 draft·prerelease의 latest 지정을 허용하지 않는다. RC의 격리는 이 prerelease 속성에 근거한다.
  - cargo-dist 0.34.0의 릴리스 명령 생성부에는 `--latest` 인자가 없고, `$PRERELEASE_FLAG`만 전달된다. GitHub REST의 새 게시 릴리스는 `make_latest` 기본값이 `true`다. 따라서 접미사가 없는 첫 정식 `moai-cli-v2026.11.0`을 기본 경로로 게시하면 `/releases/latest`가 되어 Go 3.x 안정 업데이터에 노출될 수 있다. tag namespace는 이 latest 선택을 막지 않는다.
  - 이 기본 경로는 release-checklist (g)와 3단계의 `--latest=false` 요구를 자동으로 보장하지 않는다. 정식판 게시 경로를 운영자 승인 항목 6으로 결정하기 전까지 정식판 게시는 HOLD다.

### 3.3 함께 결정할 것

- **액션 고정:** 생성 workflow는 `actions/checkout@v6`, `actions/upload-artifact@v7`, `actions/download-artifact@v8`, `actions/attest@v4`처럼 tag로 액션을 부른다. cargo-dist 0.29+의 `[dist.github-action-commits]`로 커밋 SHA에 고정할 수 있다(예: `"actions/attest" = "1e69f48acb82d1966a394da916b4c1698aa569d6"`, v4.2.2). 고정 여부는 운영자 결정이다.
- **tag 불변성:** `moai-cli-v*`를 덮는 tag ruleset은 아직 없다(release-checklist (a) HOLD). attestation은 산출물 digest에 묶이므로 tag를 옮겨도 서명이 다른 파일로 옮겨 가지는 않는다. 다만 같은 버전 재배포 금지는 ruleset이 있어야 강제된다.

## 4. 사용자 안내 — 받은 파일 확인과 macOS에서 열기

### 4.1 받은 파일 확인 (권장 순서)

```sh
# 1) 출처 확인: 이 저장소의 workflow가 만든 파일인지
gh attestation verify moai-cli-aarch64-apple-darwin.tar.xz -R modu-ai/moai-adk

# 더 좁게: 릴리스 workflow와 tag까지 확인
gh attestation verify moai-cli-aarch64-apple-darwin.tar.xz -R modu-ai/moai-adk \
  --signer-workflow modu-ai/moai-adk/.github/workflows/moai-cli-release.yml \
  --source-ref refs/tags/moai-cli-v<버전> --deny-self-hosted-runners

# 2) 보조: 내려받는 중 손상 확인(서명을 대신하지 않음)
shasum -a 256 -c sha256.sum
```

- 1)이 실패하면 그 파일을 쓰지 않는다. checksum만 맞는 경우도 마찬가지다.
- 네트워크 없이 확인하려면 먼저 `gh attestation download <파일> -R modu-ai/moai-adk`로 bundle을 받는다. 그다음 `gh attestation verify <파일> -R modu-ai/moai-adk --bundle <bundle.jsonl>`로 확인한다. `gh attestation download`는 gh 2.92.0 도움말에 "public preview"로 적혀 있다.

### 4.2 macOS quarantine

- 브라우저 등 quarantine을 기록하는 앱으로 받은 파일에는 `com.apple.quarantine` 확장 속성이 붙는다. notarization 없는 실행 파일은 이 속성 때문에 처음 실행할 때 Gatekeeper가 막는다.
- **터미널 `curl`로 받은 파일에는 이 속성이 붙지 않는다(실측, 아래 5절).** 그래서 설치 스크립트나 `curl` 다운로드로 설치하면 이 단계가 필요 없다. 그래도 4.1의 확인은 필요하다.
- 속성이 붙은 경우의 방법은 두 가지다.
  1. **Apple 권장 경로:** 한 번 실행해 막히면 "시스템 설정 → 개인정보 보호 및 보안"에서 아래로 내려 "그래도 열기(Open Anyway)"를 누른다(Apple 지원 문서 102445).
  2. **터미널:** **4.1의 `gh attestation verify`가 통과한 뒤에만** 그 파일 하나에서 속성을 지운다.

     ```sh
     xattr -d com.apple.quarantine ~/.moai-cli/bin/moai-cli
     ```

- 하지 말 것:
  - `xattr -cr`로 폴더 전체의 속성을 지우기
  - `sudo spctl --master-disable`로 Gatekeeper 전체를 끄기
  - 확인하지 않은 파일의 속성을 지우기
- 이 안내는 notarization을 대신하지 않는다. notarization을 하지 않는 것은 결정 01:1x이고, 그래서 위 수동 단계가 남는다.

## 5. 로컬 dry-run 실측 (AC-1)

모두 이 PR의 worktree에서 실행했고, 게시·쓰기는 하지 않았다.

**1) attestation subject 만들기 — `dist build --artifacts=all`**
- 조건: `cargo-dist 0.34.0`, 2026-11-09T12:28Z, base `3a88d04ac`
- 결과: exit 0. 산출물은 `moai-cli-aarch64-apple-darwin.tar.xz`(+`.sha256`), `source.tar.gz`(+`.sha256`), `sha256.sum`이다.
- `shasum -a 256 -c sha256.sum`은 두 항목 모두 OK였다. 파일 끝 빈 줄 때문에 나오는 기존 경고 1줄은 그대로다.
- 이 `sha256.sum`은 `actions/attest`의 `subject-checksums` 형식(shasum)과 같다.

**2) 생성 workflow 확인 — `dist generate --mode ci`(임시 사본)**
- 처음에는 exit 255로 실패했다: "Github CI support requires you to specify the URL of your repository".
- `Cargo.toml`에 `repository`를 넣으니 exit 0이었다.
- 기본 설정과 3.1 설정을 각각 생성해 3.2를 확인했다.

**3) 확인 절차 — `gh attestation verify`(공개 저장소의 실제 attestation)**
- 대상: `cli/cli` v2.92.0의 `gh_2.92.0_macOS_arm64.zip`(sha256 `b11c54f6bd7d15ed6590475079e5b2fcf36f45d3991a80041b29c9d0cc1f1d07`). 이 저장소에는 아직 attestation이 없어서(API 404) 이것으로 절차를 실측했다.

| 경우 | 결과 |
|---|---|
| `-R cli/cli` | exit 0 |
| `--signer-workflow cli/cli/.github/workflows/deployment.yml` 추가 | exit 0 |
| 다른 저장소(`-R modu-ai/moai-adk`) | exit 1(HTTP 404) |
| 1바이트 덧붙인 사본 | exit 1("no attestations found") |

- JSON 출력에서 이 파일 digest와 일치하는 subject 1개, predicate `https://slsa.dev/provenance/v1`, issuer `https://token.actions.githubusercontent.com`, runner `github-hosted`를 확인했다.

**4) quarantine 속성 — 로컬 파일**

| 동작 | 결과 |
|---|---|
| `xattr -w com.apple.quarantine …`로 붙인 뒤 `-p`로 읽기 | 읽힘, exit 0 |
| `-d`로 지우기 | exit 0 |
| 지운 뒤 다시 읽기 | "No such xattr", exit 1 |
| `curl -sL`로 받은 파일 | quarantine 속성 0개 |

## 6. 운영자 승인이 필요한 일 (이 PR은 하지 않음)

1. 3.1 설정을 적용하고, 생성된 `.github/workflows/moai-cli-release.yml`을 커밋할지 결정한다. 기본 생성물은 모든 `pull_request`를 대상으로 `plan` job을 실행하며, 루트 `permissions: contents: write`를 그 job이 상속한다. `attestations: write`·`id-token: write`·`contents: write`를 명시하는 host job도 추가된다. PR 실행을 제거하는 `pr-run-mode = "skip"` 제안과 권한 범위를 함께 승인해야 한다. `github-attestations`·`github-attestations-phase`가 공식 문서상 experimental이라는 점도 이 승인에 포함한다.
2. 액션을 커밋 SHA로 고정할지 정한다(`[dist.github-action-commits]`).
3. `moai-cli-v*` tag ruleset을 추가한다(release-checklist (a)).
4. 첫 실제 attestation을 만든다. 첫 `moai-cli-v…` tag를 push하면 GitHub Release와 attestation이 공개 Sigstore 투명성 로그에 남는다. 되돌릴 수 없는 외부 기록이다.
5. 위가 끝난 뒤 4.1 절차를 이 저장소의 실제 산출물로 다시 실측하고, release-checklist (e)의 not_run 항목을 갱신한다.

6. **첫 정식판의 latest 격리 경로를 결정한다.** cargo-dist 0.34.0 기본 릴리스 명령에는 `--latest=false`가 없어 첫 `moai-cli-v2026.11.0`이 Go 안정 업데이터의 `/releases/latest`에 노출될 수 있다. 후보는 릴리스 뒤 `gh release edit <정식 tag> -R modu-ai/moai-adk --latest=false`를 수행하는 별도 job, 또는 정식판을 dist 게시 workflow 밖에서 release-checklist 3단계의 명시적 `--latest=false` 경로로 게시하는 방식이다. **사후 edit 방식에는 생성부터 edit까지 latest 노출 구간이 남는다.** Go updater 무노출을 보장해야 한다면 게시 시점부터 latest가 되지 않는 경로를 선택·검증해야 한다. 경로 결정과 검증 전까지 정식판 게시는 HOLD이고, 게시 뒤 `/releases/latest`가 Go stable을 가리키는지 다시 읽어야 한다. 이 항목은 해결안의 승인 요청이며 job 추가·설정 적용·릴리스 실행은 하지 않았다.
