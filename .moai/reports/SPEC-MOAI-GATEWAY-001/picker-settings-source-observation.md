# Picker 설정 출처와 native OAuth 저장소 분리 관측

## Claim

설치 Claude Code 2.1.268에서 다음 두 가지를 구분하여 확인했다.

1. **표시 출처:** 사용자 설정이 `{}`이고 프로젝트 설정이 없는 조건에서 `--settings` overlay에만 modelPicker를 넣어도 실제 TUI에 여섯 custom 행이 나타났다. 이를 기존 성공 실행의 script·argv·snapshot·화면을 기계적으로 대조하여 확인했다.
2. **저장·인증 분리:** private `CLAUDE_CONFIG_DIR`과 기존 보안 저장소 namespace를 보존하는 `CLAUDE_SECURESTORAGE_CONFIG_DIR` 조합으로 Opus 5 native OAuth 요청이 upstream 200, Claude exit 0, 정확한 `OK` 응답을 냈다. 원래 설정 여섯 경로의 전후 존재 상태·hash·symlink 상태가 동일했다.

modelPicker 표시만 위해 별도 사용자 프로필로 복사할 필요는 없다. 그러나 native `/model`의 Enter 저장은 `--settings` overlay가 아닌 userSettings에 쓰므로, GP-005의 기본값 저장 격리를 위해서는 별도 저장 경계가 필요하다. 이번 관측으로 확인한 가장 작은 후보는 **세션 전용 설정 디렉터리 + 원래 secure-storage namespace 명시 보존**이다. 자격 복사·새 로그인은 사용하지 않았다.

## Evidence

### A. 기존 native TUI 설정 출처 재검산

```sh
python3 .moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/gateway-entry/picker_settings_source_probe.py
```

출력의 판정 및 실제 행:

```json
{
  "status": "EXISTING_NATIVE_EVIDENCE_VERIFIED",
  "client_started_this_run": false,
  "settings_source": "--settings project/picker-overlay.json",
  "user_settings_initial": "{}",
  "project_settings_present": false,
  "observed_rows": [
    "2. gpt-6-astra            Local mock probe",
    "3. gpt-5.6-sol            Local mock probe",
    "4. gpt-5.6-terra          Local mock probe",
    "5. gpt-5.6-luna           Local mock probe",
    "6. claude-opus-5          Local mock probe",
    "❯ 7. claude-sonnet-5 ✔      Local mock probe"
  ],
  "default_write_before": "{}",
  "default_write_after": "{}\n",
  "default_write_changes_original_user_settings_bytes": true,
  "real_oauth_namespace_verified": false,
  "managed_precedence_runtime_verified": false
}
```

이 JSON의 original user는 기존 **합성 HOME의 사용자 설정**을 뜻한다. 실제 사용자 HOME을 변경한 시험이 아니다. 검사기는 실행 당시 사본을 AST parse하고, `CLAUDE_CONFIG_DIR`이 동일한 합성 HOME/.claude인지, user/project snapshot에 picker가 없는지, argv가 overlay를 지정하는지, overlay의 정확한 ID와 custom 설명이 화면에 있는지, snapshot 내용 hash가 맞는지 assert했다. 별도 picker 전용 프로필로 전환한 실험이 아니다.

출력 위치:

`.moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/gateway-entry/picker-settings-source-20260911T103936Z-9339efa8/readback.json`

원래 runtime 증거는 `picker-followup-20260911T102039Z-26558abd`다. 동일한 live TUI를 불필요하게 재실행하지 않았다.

### B. 문서와 설치 코드 — 서로 다른 근거

기존 내려받은 `/tmp/gateway-claude-settings-reference-20260911.md:1064-1110`를 읽었다. [공식 settings-reference의 modelPicker 절](https://code.claude.com/docs/en/settings-reference#modelpicker)은 managed, `--settings`, user에서 picker를 읽고 project/local은 무시하며, 가장 높은 출처의 전체 lineup을 선택한다고 설명한다. 이 절은 표시 설정의 근거다. 이번 web open은 문서 크기 제한 오류가 났으므로 현재 원격 문서를 다시 내려받아 검증했다고 주장하지 않는다.

다음은 설치 실행 파일 자체에서 mmap으로 읽은 JS 코드의 byte offset이다. 실제 credential/keychain 값이나 credential 파일은 읽지 않았다.

| byte offset | 함수/연결 | 확인한 동작 |
|---:|---|---|
| 159952087 주변 | Se | CLAUDE_CONFIG_DIR 또는 homedir/.claude를 설정 root로 계산 |
| 160559138 | ibe | userSettings는 사용자 설정 root의 settings.json, flagSettings는 flagPath |
| 160872780 | tn / Gi | userSettings 갱신은 uo로 대상 경로를 구해 write 순회, flagSettings는 쓰기 대상 아님 |
| 176265612 주변 | mMe | 기본 모델 저장은 tn("userSettings", {model: ...})를 호출 |
| 190847456 | picker 호출부 | 내부 skipSettingsWrite=true여도 onSetDefault 콜백이 별도로 존재 |
| 190845000 주변 | picker 저장 완료부 | saveAsDefault가 참일 때 mMe 호출. 따라서 내부 UI prop 이름만으로 쓰기 차단을 주장할 수 없음 |
| **161483312** | **s_ / D0** | **CLAUDE_SECURESTORAGE_CONFIG_DIR가 정의되면 설정 root와 별도로 보안 저장소 root·service suffix 계산** |
| 161489207 | D0(s7) 호출 | native keychain read의 service 이름을 생성한 뒤 native security 호출에 전달 |
| 161491792 | file fallback u | 보안 저장소 root s_ 아래 .credentials.json 경로 계산 |
| 172342796 주변 | env 전달 | secure-storage override만은 빈 문자열도 제외하지 않고 전달 |

s_ / D0의 정확한 404바이트 JS를 추출하여 node에서 synthetic environment·homedir·OAuth suffix를 주고 함수 자체를 실행했다. credential I/O는 없었다. 추출 코드 SHA-256:

```text
36973f7ef846ad598d2d62d77a71a6dd4e0f33cce025e9dd9a4ccd16439d19ef
```

네 계산 모두 변경 전후 root와 service가 같았다.

| 원래 환경 | private 설정으로 바꿀 때 보존한 secure override | 결과 |
|---|---|---|
| 두 override 없음 | 명시 빈 문자열 | homedir/.claude 및 suffix 없는 service 유지 |
| config=/synthetic/profile | /synthetic/profile | 해당 root·hash suffix 유지 |
| secure=/synthetic/secure | /synthetic/secure | 해당 root·hash suffix 유지 |
| config=/synthetic/profile, secure="" | 명시 빈 문자열 | homedir/.claude 및 suffix 없는 service 유지 |

결과는 `picker-securestorage-pure-readback.json`, 추출 함수는 `picker-securestorage-functions.js`에 있다. **CLAUDE_CONFIG_DIR가 명시 빈 문자열인 별도 경우는 검증하지 않았으며 live 관측기는 그 입력을 거절한다.**

이 변수의 공개 문서 지원 상태는 확인하지 못했다. 설치 버전에 실제 존재하고 실행으로 관측한 내부 경로로 분류한다. 전역 모델 저장을 끄는 공개 CLI 플래그가 있다는 주장도 하지 않는다.

### C. 기존 native OAuth와 private 설정을 함께 사용한 단일 live 요청

live 실행 전에 관측한 환경은 HOME=/Users/goos이며 CLAUDE_CONFIG_DIR와 CLAUDE_SECURESTORAGE_CONFIG_DIR가 모두 미정의였다. 따라서 child에만 다음 조합을 주었다.

```text
CLAUDE_CONFIG_DIR=<새 private 임시 설정 디렉터리>
CLAUDE_SECURESTORAGE_CONFIG_DIR=""
```

기존 M0의 별도 X-MoAI-Session-Token loopback forwarder를 그대로 재사용했다. M0의 provider·token override 제거 규칙을 유지했고, refresh challenge는 사용하지 않았다. native Claude가 자기 보안 저장소를 읽으며 관측기는 그 값을 저장소에서 읽거나 복사하지 않는다. private 디렉터리에는 빈 settings, onboarding 표시, modelPicker/fallback overlay만 만들었다. 실제 계정 metadata도 복사하지 않았다.

```sh
/opt/homebrew/bin/timeout 160 python3 .moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/gateway-entry/picker_settings_oauth_probe.py
```

단일 handle 38796 최종 출력:

```json
{
  "status": "OBSERVED",
  "native_oauth_response_observed": true,
  "original_settings_unchanged": true,
  "client_count": 1,
  "evidence": "/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified/.moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/gateway-entry/picker-settings-oauth-20260911T104425Z-8bc33428"
}
```

result.json과 native-m0-evidence.json을 다시 읽었다. 실제 결과:

```text
model: claude-opus-5
native returncode: 0
stdout_exact_OK: true
authorization_kind: bearer
api_key_present: false
oauth_beta_present: true
custom_token_matches: true
upstream_status: 200
stop_reason: null
client_count: 1
```

인증 값은 출력·파일에 기록하지 않았다. M0는 수신 Authorization의 SHA-256만 기록한다. 실제 요청은 stream=true, max_tokens=64000, adaptive thinking, high effort, context management 있음, tools 0이었다.

다음 여섯 경로는 전후 existence·hash·symlink 상태가 일치했다. 내용은 복사하지 않고 hash만 계산했다.

- /Users/goos/.claude/.claude.json — 전후 없음
- /Users/goos/.claude/settings.json
- /Users/goos/.claude/settings.local.json
- /Users/goos/.claude.json
- 대상 worktree/.claude/settings.json
- 대상 worktree/.claude/settings.local.json

각 SHA는 result.json에 있다. 정리 결과:

```text
private_config_removed: true
client_pid_exists: false
listener_connect_ex: 61
```

## Baseline-attribution

이번 source/readback 작업에서 git HEAD를 다시 읽어 `81c1d58f9`를 확인했다. 설치 binary 경로는 `/Users/goos/.local/share/claude/versions/2.1.268`, 길이는 202081536바이트다.

```text
binary SHA-256: 06a96d5423f83770f120859f1c58e60d7252cc4c122aa13043b7e7cd716bc76a
live wrapper SHA-256: 96b97bcf61da59c1a091668653e6d26aaf907c1fedc4c18d57f55d36e63ed553
```

source/probe scripts, 산출물, 이 보고서만 작성했다. 제품 코드·원본 probe·SPEC body는 변경하지 않았다. live 실행은 2026-09-11 10:44:25 UTC에 시작했으며 gate는 10:00 UTC였다. 새 로그인·credential 복사·반복 auth retry·refresh 강제 발생은 없었다.

## Gaps

native TUI picker와 live OAuth 동작을 같은 한 프로세스에서 함께 시험한 것은 아니다. 앞의 합성 TUI는 표시·저장 동작을, 뒤의 live print는 private config와 기존 OAuth의 공존 및 원본 설정 불변을 각각 확인했다. GP-005 전체의 동시 세션·취소·재시작·다른 플랫폼·모든 설정 출처·managed 정책·named profile·사용자 수정 충돌을 입증하지 않는다.

원본 여섯 설정 경로 밖 모든 파일이 불변이라고 주장하지 않는다. native credential refresh가 저장소를 갱신하지 않았다는 뜻도 아니며, credential 저장소 내용 전후 비교는 수행하지 않았다. 공개 지원 약속은 확인하지 못했다. 전체 CI 및 제품 수용 판정은 PENDING이다.

## Residual-risk / 구현에 전달할 경계

설정 root와 secure-storage root를 분리하는 이 내부 변수는 대상 버전에서 직접 증명되었으나 다른 버전에서 달라질 수 있다. 구현 시 원래 secure override가 존재하면 빈 문자열까지 그대로 보존하고, 없으면 원래 config override 또는 기본 namespace를 선택해야 한다. private config로 바꾼 뒤 그 값을 보고 원래 namespace를 추측해서는 안 된다.

원본 settings와 계정 정보를 복사하지 않아도 이번 작은 native OAuth 요청은 성립했다. 필요한 비밀이 아닌 설정을 어떻게 읽고 세션 overlay에 적용할지, 사설 설정 안에서 model 저장이 어떤 범위로 남을지, 원래 프로젝트 경로의 세션과 resume를 어떻게 이어갈지는 후속 제품 계약이다. 이 보고서만으로 전체 launcher를 활성화하지 않는다.
