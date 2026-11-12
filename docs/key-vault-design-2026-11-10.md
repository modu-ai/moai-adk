# API 키 보관함과 경로 통일 — 조사·설계 1단계 (2026-11-10)

Issue #2393 (rc.2 런처·보안, 추적 #2389 B7). 근거: Issue 본문의 사용자 지시(2026-11-10T13:17Z, 리더 세션 직접) — "`~/.moai/.env.glm` … 키보관함을 만들어서 별도로 보관하도록 재설계", "`~/.moai` → `~/.moai-cli`로 변경해서 통일", "스킬명만 `/moai`". 정본: 「MoAI 홈 설치와 프로젝트 생성 기능 계약」(Space `page_83f162ef14d08191917913738184f167`).

> **이후 결정으로 대체된 부분(구현 기준).** 이 문서는 조사·설계 1단계의 기록이고, 아래 OS 키 보관함(keyring)·`key_store` 설정·평문 파일 항목·"옛 파일 읽기 유지"는 다음 사용자 결정으로 대체되었다. 구현은 이 결정들을 따른다.
> - Issue #2393 댓글 6777274025: OS 키 보관함은 쓰지 않는다. 세 OS 모두 같은 파일 보관을 쓰고, 내용은 암호화한다.
> - 6778150969: 순서는 환경 변수 → vault. 옛 `~/.moai/.env.glm`은 읽지 않고 `legacy_key_only`로 이전만 안내한다.
> - 6778671862: 열쇠는 머신 열쇠 파일 `~/.moai-cli/vault/.master.key`(32바이트, 첫 `key-set`에서 생성)다. 폴더는 `<이름>-<clone_id 앞 8자>`와 공통 `_shared`이고, 알고리즘은 XChaCha20-Poly1305(`chacha20poly1305`)다. 형식 이전 코드는 두지 않고, 읽는 순서는 환경 변수 → 프로젝트 → 공통이다.
> - 이 방식은 열쇠 파일 없이 복사·동기화·붙여넣기된 항목만 막고, 같은 사용자 권한의 프로그램은 막지 못한다.
> 현재 동작의 정본은 `docs/rust-product-path.md`의 `key.*`·`glm` 행이다.

## 0. 범위와 선언

- **문서 전용.** 이 문서는 인수 기준 1(조사 기록, 설계 1편, 사용자 질문 모음)만 다룬다. 코드·런처 연결·이전 명령(인수 기준 2~4)은 열린 PR 상한(#2388) 아래에서 별도 PR로 한다.
- **실제 키 0회.** 사용자 키 파일(`~/.moai/.env.glm`, `~/.moai-cli/.env.glm`)의 내용을 읽지 않았다. login keychain, keychain 검색 목록, Secret Service, Windows 자격 증명 관리자에는 읽기·쓰기 모두 하지 않았다. 보안 설정 변경도 없다.
- **빌드 0회.** 이 worktree에서 cargo를 실행하지 않았다(#1975 공지 6768775536 규칙 4, 디스크 여유). 코드 사실은 tip `c6a8a1b2b39184e6e3517b8a85afa859f6d50f53`의 소스를 읽어서 확인했다.
- **확인 방식.** 출처는 모두 2026-11-10 13:51Z–13:53Z(UTC)에 HTTP GET으로 받아 sha256을 기록했다(§8). 받은 원문에서 읽은 문장만 "확인"으로 쓰고, 원문이 말하지 않는 것은 "미확인", 실측하지 않은 것은 not_run으로 적는다.
- **결정 기록 공백.** `src/session.rs:67-69` 주석은 "`~/.moai/.env.glm`이 Go 도구와 공유하는 전역 저장소이며 첫 후보"라는 사용자 결정(2026-11-10)을 기록한다. 이 Issue의 13:17Z 지시가 그것을 대체한다는 행은 decisions log에서 찾지 못했다(`grep -n -E '2393|키 ?보관|env\.glm|keychain' docs/decisions-*.md` 0줄). 그래서 §7 Q0으로 올린다.

## 1. 현재 상태 (tip `c6a8a1b2b3` 실측)

### 1.1 GLM 키를 읽는 길

| 순서 | 출처 | 코드 |
|---|---|---|
| 1 | 환경 변수 `ANTHROPIC_AUTH_TOKEN`(값이 비었거나 공백·제어 문자만 있으면 없는 것으로 봄) | `src/session.rs` `glm_token`, `usable_token` |
| 2 | `~/.moai/.env.glm` (Go 경로, 첫 파일 후보) | `GLM_KEY_FILES[0]`, `src/session.rs:70` |
| 3 | `~/.moai-cli/.env.glm` | `GLM_KEY_FILES[1]` |

- 파일에서는 `ANTHROPIC_AUTH_TOKEN` 또는 `GLM_API_KEY` 이름을 받아들이고, `export ` 접두와 따옴표를 벗긴다. 일반 파일이면서 64 KiB 이하일 때만 연다(FIFO·장치 차단). 파일은 읽기만 하며 만들거나 고치지 않는다.
- 값은 host 프로세스 환경으로만 넘기고 출력·로그·저장하지 않는다(#2359). 키가 없으면 `blocked:` 문구로 멈춘다(`src/session.rs:463-470`).
- 파일 권한(mode)·소유자·symlink는 지금 검사하지 않는다.
- Go 상태줄 스크립트 `.moai/status_line.sh`도 `~/.moai/.env.glm`을 source한다(Issue 본문, 리더 확인). 따라서 Go 도구를 지우기 전에 원본 파일을 옮기거나 지우면 Go 상태줄이 깨진다(#2383과 순서 연결).

### 1.2 `~/.moai`·`.moai` 경로 문자열 (인수 기준 3의 기준선)

두 가지 방법으로 셌다.

- **방법 A**: `git grep -n -E '"[^"]*\.moai/' -- 'src/*.rs' | grep -v '\.moai-cli'` → 5줄. 같은 줄에 `.moai-cli`가 함께 있는 `src/session.rs:70`을 빠뜨린다(줄 단위 제외의 한계).
- **방법 B**: Rust 문자열 리터럴을 하나씩 떼어 `.moai`가 경로 마디(뒤가 `/` 또는 끝)로 나오는 것을 세는 스크립트(`target/research/count_moai_paths.py`, 저장소에 넣지 않음) → 12개(실제 코드 10, `#[cfg(test)]` 뒤 시험 2).

방법 B의 실제 코드 10개를 용도로 나누면 다음과 같다.

| 위치 | 용도 | 기본 경로 읽기/쓰기인가 | 처리 |
|---|---|---|---|
| `src/session.rs:70` `.moai/.env.glm` | GLM 키 파일 첫 후보 | **예(읽기)** | 기본 경로에서 빼고 이전 명령 전용으로 옮김(§4) |
| `src/apply.rs:369` `.moai` | receipt 복구 금지 경로(가드) | 아니오(거부) | 유지 |
| `src/baseline.rs:24` `.moai` | Go 상태 보존 기준선 측정 대상 | 아니오(측정) | 유지 |
| `src/install.rs:609-610` `.moai` ×2 | `go_generation` 보고의 폴더 존재 여부 | 아니오(존재 확인만) | 유지 |
| `src/migrate.rs:383`(×2), `423`, `424`, `435` | 프로젝트 `.moai/config`·`.moai/specs` 이전 원본 | 이전 전용 | 유지(예외로 명시) |

방법 B의 한계: 파일에서 처음 나온 `#[cfg(test)]` 뒤를 모두 시험으로 본다. 그래서 시험 모듈 **뒤에** 놓인 실제 코드 문자열은 시험으로 잘못 분류된다. 변이 실험에서 실제로 확인했다. 복사본 `session.rs` 맨 앞에 `.moai/` 문자열을 넣으면 prod 11(검출), 맨 끝(시험 모듈 뒤)에 넣으면 prod 10(놓침), `.moai/.env.glm`을 빼면 prod 9(검출), `.moai-cli/`만 넣으면 prod 10(오검출 없음)이었다. 구현 PR의 checker는 `#[cfg(test)]` 모듈의 범위를 괄호 짝으로 정확히 잘라야 한다(§6).

결론: 기본 경로로 `~/.moai/` 아래 **내용**을 읽는 곳은 `src/session.rs:70` 한 곳이다. 쓰는 곳은 0이다. 구현 PR은 이 수를 0으로 만들고, 같은 두 방법으로 다시 센 수치를 PR에 적는다.

## 2. 조사 — 후보 저장소 (채택/기각)

| ID | 후보 | 원문에서 확인한 사실 | 무인 레인(AC2) 관점 | 판정 |
|---|---|---|---|---|
| R1 | macOS Keychain Services | 비밀을 "encrypted database called a keychain"에 저장한다(A2). macOS의 iCloud 밖 항목은 ACL을 가진다. 호출 앱이 신뢰 목록에 없으면 "the system prompts the user for confirmation"이고, Always Allow를 고르면 그 앱이 신뢰 목록에 들어간다(A5). `kSecUseAuthenticationUI`의 기본값은 `kSecUseAuthenticationUIAllow`(A6)다. `kSecUseAuthenticationUIFail`을 주면 인증이 필요할 때 `errSecInteractionNotAllowed`를 돌려준다(A7, A8). 기본 접근성은 "When Unlocked"(A3) | 확인 창을 막는 질의 옵션이 있어서 "멈추지 않고 차단 응답"을 만들 수 있다. 다만 **다시 빌드한 서명 없는 binary를 같은 신뢰 앱으로 보는지**는 받은 원문에 없다(미확인). 레인은 binary를 자주 다시 빌드하므로, 매번 신뢰 목록에서 빠지면 키를 읽을 때마다 `errSecInteractionNotAllowed`가 날 수 있다 | **선택 backend로 채택**, 기본값은 보류(Q2) |
| R2 | Linux Secret Service (freedesktop) | 잠긴 항목의 비밀은 읽을 수 없다. `Unlock()`은 확인 창 없이 바로 열 수 있는 객체만 돌려주고, 나머지에는 prompt 객체를 돌려준다. prompt는 클라이언트가 실행해야만 진행된다. 서비스는 언제든 다시 잠글 수 있다(S2) | prompt 객체를 실행하지 않으면 확인 창이 뜨지 않는다. 그러면 "잠김 → 차단 응답"이 결정적이다. 헤드리스·SSH 세션에서 D-Bus session bus나 데몬이 없을 때 어떻게 실패하는지는 받은 원문에 없다(미확인) | **선택 backend로 채택**, 잠김이면 prompt 없이 차단 |
| R3 | Windows Credential Manager | `CredWrite`는 자격 증명을 현재 token의 logon session에 묶는다. 실패 코드 `ERROR_NO_SUCH_LOGON_SESSION`이 있고, "Network logon sessions do not have an associated credential set"(W1). `CRED_TYPE_GENERIC`, blob 상한 `CRED_MAX_CREDENTIAL_BLOB_SIZE`(5×512 = 2,560 bytes)(W2) | SSH 같은 network logon에서는 저장소 자체가 없다. 그래서 그 경우를 차단 응답으로 다뤄야 한다. 키 길이 상한 2,560 bytes는 GLM 키에 충분하지만 시험에 경계값을 넣는다 | **선택 backend로 채택** |
| R4 | `keyring` crate (Rust) | 최신 4.2.0(crates.io `max_stable_version`, 갱신 2026-08-29)(K2). 기본 `v1` feature는 macOS Keychain Services, Windows Credential Manager, *nix Secret Service를 자동으로 고른다(K4). 4.x는 `keyring-core`와 플랫폼별 store crate들로 나뉘었다(K1, K5) | 각 store가 R1의 `kSecUseAuthenticationUIFail`을 쓰는지, R2에서 prompt를 실행하는지는 원문에 없다(미확인). 구현 카드에서 store crate 소스를 읽어 확인해야 한다 | **구현 수단 후보**. 채택 전에 소스 확인과 의존성 검토(Q3) |
| R5 | Linux 커널 keyring(keyutils) | user·user-session keyring은 UID 기록이 커널에 있는 동안만 존재한다. persistent keyring에는 만료 시간이 있다(L1) | 재부팅이나 만료 뒤 사라지므로 영구 저장소가 아니다 | **기각**(영구 보관 불가). 세션 캐시 용도는 범위 밖 |
| R6 | 권한 제한 파일(0600) | Claude Code: macOS는 Keychain에 쓰고, "When the Keychain rejects the write, such as when it's locked in an SSH session"이면 `~/.claude/.credentials.json`(mode 0600)에 쓴다. Linux는 처음부터 0600 파일, Windows는 사용자 프로필 ACL을 상속한 파일(C2). Codex `cli_auth_credentials_store`는 `file`(`CODEX_HOME/auth.json`) · `keyring`(OS 저장소, 없으면 실패) · `auto`(OS 저장소, 없으면 파일) · `ephemeral`(메모리) 네 가지다(X1, X2). gh는 시스템 credential store에 저장하고, 없거나 문제가 있으면 평문 파일로 떨어진다. `--insecure-storage`로 처음부터 평문을 고를 수 있다(G1) | 확인 창이 없다. 같은 사용자의 다른 프로세스는 읽을 수 있다(평문 at rest) | **기본 backend로 채택안**(Q1·Q2). 세 host 모두 같은 형태의 대체 경로를 둔다 |
| R7 | host 전역 설정에 키를 둠 | z.ai 공식 안내는 `~/.claude/settings.json`의 `env`에 `ANTHROPIC_AUTH_TOKEN`을 쓰는 스크립트를 제공한다(P1). Claude Code의 `apiKeyHelper`는 셸 스크립트로 API key를 돌려주고 설정 변경을 실행 중에 다시 읽는다(C1, C2). 관리형 설정에서 `forceLoginOrgUUID` 같은 정책이 있으면 `ANTHROPIC_AUTH_TOKEN`·`apiKeyHelper` 세션을 시작 시 막는다(C2) | 사용자 전역 설정을 고쳐야 한다. Issue 범위 2항("전역 설정 불변")과 #2359 원칙(세션 환경에만 주입)에 어긋난다. GLM 세션이 아닌 Claude 세션에도 키가 퍼진다 | **기각** |
| R8 | 환경 변수만 | 현행 1순위 | 사용자 셸 설정에 평문으로 남기 쉽다. 이 Issue가 바로잡으려는 대상이다 | **유지**(우선순위 1, 명시 override) |

## 3. 설계안

### 3.1 저장소 모양

- **보관함 위치**: `~/.moai-cli/vault/`(디렉터리 0700). provider마다 파일 하나를 둔다. 예: `~/.moai-cli/vault/glm.key`(0600, 값 한 줄, 이름=값 형식 아님). 이름은 사용자 결정 전까지 가칭이다(Q4).
- **backend 선택**: `~/.moai-cli/config`의 `key_store = file | keyring | auto`. 이름과 의미는 Codex `cli_auth_credentials_store`(X1)를 따르되, `ephemeral`은 환경 변수(R8)가 이미 그 역할을 하므로 두지 않는다.
  - `file`(제안 기본값): 위 0600 파일.
  - `keyring`: OS 저장소(R1~R3). 없거나 잠기면 차단한다(파일로 떨어지지 않음).
  - `auto`: OS 저장소를 쓰고, 저장소가 **없을 때만** 파일로 떨어진다. 잠김은 떨어지지 않고 차단한다. 잠긴 저장소를 피해 평문으로 조용히 쓰는 일을 막기 위해서다. 이 점은 Claude Code의 SSH 대체 동작(C2)과 일부러 다르게 잡았다(Q2).
- **OS 저장소 항목 이름**: service `moai-cli`, account `<provider>`(예: `glm`). Windows target 이름은 회사 접두를 권장한다(W2). 그래서 `moai-cli:glm` 형식으로 둔다.

### 3.2 파일 backend의 안전 규칙 (확인 창이 없는 대신 강제할 것)

1. 파일을 읽기 전에 `symlink_metadata`로 확인한다. 일반 파일이 아니거나 symlink이면 거부한다.
2. 소유자가 현재 사용자가 아니거나, group·other 권한 비트가 하나라도 있으면 거부한다(Unix). 이때 문구에는 고치는 명령(`chmod 600 <path>`)만 담고 값은 담지 않는다. Windows는 프로필 디렉터리 ACL을 상속한다는 점(C2 동형)을 문서에 적고, 이 검사는 하지 않는다.
3. 크기 상한은 현행 `GLM_KEY_FILE_MAX`(64 KiB)를 유지하고, 시험은 N/N+1 경계를 둔다.
4. 쓸 때는 같은 디렉터리의 임시 파일을 `0600`으로 만든다(umask에 기대지 않음). 그다음 rename으로 원자 교체한다. 디렉터리가 없으면 `0700`으로 만든다.
5. 값 검사는 현행 `usable_token` 규칙을 그대로 쓴다(공백·빈 값·제어 문자 거부).

### 3.3 런처 연결 (인수 기준 2)

읽는 순서는 다음과 같다.

1. 환경 변수 `ANTHROPIC_AUTH_TOKEN`(현행 유지)
2. 선택된 backend의 `glm` 항목
3. 그 밖의 경로는 읽지 않는다. `~/.moai/.env.glm`과 `~/.moai-cli/.env.glm`도 마찬가지다.

- 값은 host 프로세스 환경으로만 넘긴다. 키 값은 argv, dry-run, receipt, 로그, 오류 문구, `notes`에 나오지 않는다. 현행 #2359 시험을 backend별로 확장한다.
- OS 저장소 호출은 확인 창이 없는 질의로만 한다. macOS는 `kSecUseAuthenticationUIFail`(A7), Secret Service는 prompt 객체를 실행하지 않음(S2), Windows는 확인 창 없는 `CredRead`를 쓴다. 그리고 런처 쪽 시간 상한(제안 3초)을 둔다. 시간 상한은 D-Bus가 응답하지 않을 때를 위한 것이다(R2 미확인 항목).
- 차단 응답은 기계가 읽을 수 있게 고정한다. 무인 세션은 이것을 받고 멈추지 않고 보고한다.

| code | 조건 | 다음 행동 안내(문구) |
|---|---|---|
| `key_absent` | 환경 변수 없음 + 선택 backend에 항목 없음 | `moai-cli key set glm` (값은 stdin) |
| `key_store_locked` | OS 저장소가 잠김 / `errSecInteractionNotAllowed` / Secret Service가 prompt를 요구 | 사용자가 잠금을 해제하거나 `key_store = file` |
| `key_store_unavailable` | 저장소 없음(network logon `ERROR_NO_SUCH_LOGON_SESSION`, D-Bus 없음, 시간 초과) | `auto`면 파일 backend로 넘어가고, `keyring`이면 차단 |
| `key_file_unsafe` | 파일 backend 안전 규칙(§3.2) 위반 | `chmod 600 …` 등 고칠 명령 |
| `legacy_key_only` | 보관함에는 없고 `~/.moai/.env.glm` 또는 `~/.moai-cli/.env.glm`이 **존재**함(내용은 읽지 않음, 존재 확인만) | `moai-cli key import glm --preview` |

### 3.4 명령 표면 (구현 카드용 초안)

- `moai-cli key status [<provider>]`: backend, 항목 유무, 파일 mode·소유자 검사 결과를 보여준다. 값·길이·지문은 보여주지 않는다.
- `moai-cli key set <provider>`: 값을 stdin에서만 받는다. argv에 값을 받는 형태는 두지 않는다(프로세스 목록 노출). tty면 반향 없이 읽는다.
- `moai-cli key remove <provider>`: 보관함 항목만 지운다. 실행 전 확인을 받고, 원본 Go 파일은 건드리지 않는다.
- `moai-cli key import <provider> --preview | --apply`: §4의 이전 절차.

## 4. 이전 절차 (인수 기준 4)

1. **preview**(쓰기 0): 원본 후보(`~/.moai/.env.glm`, `~/.moai-cli/.env.glm`)의 존재, 안전 규칙 검사 결과, 받아들이는 키 이름이 있는지(예/아니오), 목적지 backend와 경로를 보여준다. 값은 보여주지 않는다. preview digest를 출력한다.
2. **사용자 확인**: `--apply --preview-digest <d>`. digest가 다르면 거부한다. 그 사이 원본이 바뀐 경우를 막기 위해서다.
3. **apply**: 보관함에 쓰고 다시 읽은 다음 원본 값과 바이트 비교만 한다(출력 없음). 결과를 receipt(`~/.moai-cli/` 아래)에 남긴다. receipt에는 경로·시각·backend·일치 여부만 적는다.
4. **원본 처리**: 하지 않는다. 원본 삭제나 이름 변경은 별도 명령과 별도 사용자 확인이 필요하다. Go 상태줄이 같은 파일을 source하므로(§1.1), Go 도구 제거(#2383) 뒤로 미룬다.
5. **되돌리기**: `key remove glm`. 원본이 그대로이므로 이전 전 상태로 돌아간다. 다만 런처는 원본을 기본 경로로 읽지 않으므로 `legacy_key_only` 차단 응답이 나온다. 이것은 의도한 동작이다.

## 5. 경로·명칭 통일 (범위 3)

- Rust가 읽고 쓰는 곳은 `~/.moai-cli/`와 `<project>/.moai-cli/`뿐이다. §1.2 표의 `session.rs:70`만 고치고, 나머지(가드·측정·존재 확인·프로젝트 이전 원본)는 "예외: 읽기 금지 가드·측정·이전 전용"으로 문서와 시험에 이름을 붙여 둔다.
- 제품·폴더·실행 파일 명칭은 `moai-cli`, 사용자 skill 이름만 `moai`(Issue 범위 3). #2363(사용자 대면 표기 'MoAI CLI' 제거, 'MoAI'로 통일)과 표현이 겹친다. 그래서 어느 쪽이 화면 문구를 정하는지 Q5로 묻는다.

## 6. 구현 카드 시험 계획 (인수 기준 2~4의 증거 모양)

- 모든 시험은 임시 HOME과 합성 키(`moai-test-key-<random>`)로 한다. 실제 OS 저장소에는 쓰지 않는다. OS backend는 trait 뒤에 두고, 시험은 가짜 구현으로 `locked`·`unavailable`·`timeout`·`absent`를 만든다.
- 출력 전체(stdout·stderr·receipt·dry-run JSON)에서 합성 키 문자열이 0회인지 확인한다. 각 차단 code마다 1개 이상 시험한다.
- 파일 backend: mode 0600 통과, 0640·0604 거부(각 비트 하나씩), symlink 거부, FIFO 거부, 64 KiB N/N+1 경계. 상한은 상수를 쓰지 않고 리터럴 값으로 적는다. 경계 변이 시험을 양쪽으로 한다.
- 기본 경로 0건: §1.2의 두 방법을 시험 또는 checker로 고정한다. 이때 방법 B는 시험 모듈 범위를 괄호 짝으로 자르도록 고친다. 그리고 "시험 모듈 뒤의 실제 코드 문자열" 변이가 FAILED가 되는지 확인한다.
- 이전 명령: preview는 쓰기 0(디렉터리 digest 전후 비교), digest가 다르면 거부, apply 뒤 원본 바이트·mtime 불변, remove 뒤 원본 불변.
- 실제 OS 저장소를 쓰는 확인(아래 not_run)은 사용자 확인 뒤, 전용 임시 keychain 또는 별도 계정에서만 한다.

## 7. 사용자 질문 (리더 경유)

- **Q0 결정 기록**: `src/session.rs:67-69`가 기록한 "`~/.moai/.env.glm`을 첫 저장소로 공유" 결정을 이 Issue의 13:17Z 지시가 대체한다고 decisions log에 행을 추가해도 되는가? 구현 PR은 이 행을 근거로 삼는다.
- **Q1 보관 방식**: "키 보관함"을 (a) OS 저장소(Keychain·Secret Service·Credential Manager)로 볼지, (b) `~/.moai-cli/vault/`의 권한 제한 파일로 볼지, (c) 둘 다 두고 설정으로 고르게 할지(제안: c).
- **Q2 기본 backend**: 제안은 `file`이다. 이유는 레인이 binary를 자주 다시 빌드하는데, macOS에서 다시 빌드한 binary가 Keychain 신뢰 목록에 남는지 미확인이기 때문이다(R1). 그래서 무인 레인이 `key_store_locked`로 반복 차단될 위험이 있다. `auto`나 `keyring`을 기본값으로 올리려면 not_run 항목 N1을 먼저 실측해야 한다.
- **Q3 의존성**: `keyring` 4.x(`keyring-core` + 플랫폼 store crate) 추가를 허용하는가, 아니면 플랫폼 API를 직접 부를까(macOS `security-framework`, Windows `windows-sys`, Linux `zbus`)? 어느 쪽이든 공급망 검토가 필요하다.
- **Q4 이름**: 보관함 경로 `~/.moai-cli/vault/`, 명령 `moai-cli key …`, OS 항목 service 이름 `moai-cli` — 이대로 가는가?
- **Q5 명칭 충돌**: 범위 3("제품 명칭은 `moai-cli`")과 #2363("사용자 대면 표기 'MoAI'")이 화면 문구에서 겹친다. 경로·실행 파일·저장소 표기는 `moai-cli`, 화면 제목은 `MoAI`로 나누면 되는가?
- **Q6 forge 토큰**: Issue 범위 1은 forge 토큰도 언급한다. gh는 이미 시스템 credential store에 저장한다(G1). moai-cli가 GitHub 토큰을 따로 보관하지 않고 `gh auth token`에 맡기는 쪽을 제안한다. 중복 보관하면 회전·폐기 지점이 두 곳으로 늘어난다. 이대로 가는가?
- **Q7 Go 원본 처리 시점**: 원본 `~/.moai/.env.glm`의 이름 변경·삭제는 Go 상태줄 제거(#2383) 뒤 별도 확인으로 미루는가?

## 8. 출처 (2026-11-10 UTC 확인, raw 사본은 worktree `target/research/raw/`에만 있고 저장소에 넣지 않음)

| ID | 확인 시각 | HTTP | sha256(앞 12자) | URL |
|---|---|---|---|---|
| A2 | 13:53:11Z | 200 | `ab3f36fc07fd` | https://developer.apple.com/documentation/security/keychain-services.md |
| A3 | 13:53:12Z | 200 | `472dff31b04c` | https://developer.apple.com/documentation/security/restricting-keychain-item-accessibility.md |
| A5 | 13:53:27Z | 200 | `217df19c75fe` | https://developer.apple.com/documentation/security/access-control-lists.md |
| A6 | 13:53:27Z | 200 | `828d80d58141` | https://developer.apple.com/documentation/security/ksecuseauthenticationui.md |
| A7 | 13:53:28Z | 200 | `f1c20c1eeb43` | https://developer.apple.com/documentation/security/ksecuseauthenticationuifail.md |
| A8 | 13:53:29Z | 200 | `a67558827f1a` | https://developer.apple.com/documentation/security/errsecinteractionnotallowed.md |
| S2 | 13:51:50Z | 200 | `fbe4d9efdf45` | https://specifications.freedesktop.org/secret-service-spec/latest/unlocking.html |
| W1 | 13:51:52Z | 200 | `dbad31ad4593` | https://learn.microsoft.com/en-us/windows/win32/api/wincred/nf-wincred-credwritew |
| W2 | 13:51:53Z | 200 | `7f8149603a5f` | https://learn.microsoft.com/en-us/windows/win32/api/wincred/ns-wincred-credentialw |
| K1 | 13:51:46Z | 200 | `59f446fed5cb` | https://docs.rs/keyring/latest/keyring/ |
| K2 | 13:51:47Z | 200 | `41efe12e9b43` | https://crates.io/api/v1/crates/keyring |
| K4 | 13:53:14Z | 200 | `b8a09c46c093` | https://docs.rs/keyring/latest/keyring/v1/index.html |
| K5 | 13:53:14Z | 200 | `d8c952c91246` | https://docs.rs/crate/keyring/latest/source/Cargo.toml.orig |
| L1 | 13:53:15Z | 200 | `1e7643795399` | https://man7.org/linux/man-pages/man7/keyrings.7.html |
| G1 | 13:51:53Z | 200 | `180ede83cf37` | https://cli.github.com/manual/gh_auth_login |
| C1 | 13:51:54Z | 200 | `2b1e5db28236` | https://code.claude.com/docs/en/settings |
| C2 | 13:51:54Z | 200 | `005a926f0785` | https://code.claude.com/docs/en/iam |
| X1 | 13:51:55Z | 200 | `00ffac9169c7` | https://developers.openai.com/codex/config-reference |
| X2 | 13:51:55Z | 200 | `f9a97c492631` | https://developers.openai.com/codex/auth |
| P1 | 13:53:17Z | 200 | `f445886abc3b` | https://docs.z.ai/devpack/tool/claude |

받았지만 쓰지 않은 것도 기록한다.

- K3(keyring README)는 K1·K4와 같은 내용이다.
- S1(Secret Service 목차)은 목차뿐이다.
- G2(`gh auth token`)에서는 관련 문장을 찾지 못했다.
- A1(Keychain HTML)은 JavaScript가 필요해서, 같은 내용의 Markdown 판 A2로 바꿨다.
- A4(`kSecAttrAccessibleWhenUnlockedThisDeviceOnly`)는 iOS 데이터 보호 속성이라 macOS 파일 기반 keychain 판단에는 쓰지 않았다.

## 9. 미확인·not_run

- **N1** macOS: 다시 빌드한(서명 없는·ad-hoc 서명) `moai-cli`가 이전 빌드의 Keychain "Always Allow" 신뢰를 이어받는지 — not_run. 실측하려면 임시 keychain을 만들어야 한다. 그런데 `security create-keychain`이 검색 목록을 건드릴 수 있어서 사용자 확인 뒤 별도 계정에서 한다.
- **N2** Linux: D-Bus session bus나 Secret Service 데몬이 없는 SSH·CI 세션에서 store crate가 어떤 오류를 얼마나 빨리 돌려주는지 — 미확인(원문 없음), not_run(Linux VM 필요).
- **N3** `keyring` 4.x의 apple·secret-service store가 확인 창 없는 질의(`kSecUseAuthenticationUIFail`, prompt 미실행)를 쓰는지 — 미확인. 구현 카드에서 store crate 소스를 읽어 확인한다.
- **N4** Windows: SSH(network logon)에서 `CredRead`가 `ERROR_NO_SUCH_LOGON_SESSION`을 돌려주는지 실측 — not_run(W1은 `CredWrite` 문서다).
- **N5** Go 상태줄 `.moai/status_line.sh`가 `~/.moai/.env.glm`을 source한다는 사실은 Issue 본문(리더 확인)에서 가져왔다. 이 문서 작성 중에는 해당 파일을 다시 읽지 않았다.
