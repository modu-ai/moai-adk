# t53 — install.bat rate-limit 버그 수정 (세 번째 미수정 표면)

- 카드: `moai todo` t53
- 브랜치: `WT-t53` (base: `release/v3.1.1` @ `36a12cf82`)
- 날짜: 2026-08-17

## 배경

install.sh(#1559) / install.ps1(#1567)이 `/releases/latest` 리다이렉트 계약으로
이주한 뒤 남아 있던 세 번째 표면. `install.bat:58`은 비인증
`api.github.com/repos/modu-ai/moai-adk/releases` **리스트** 엔드포인트를
PowerShell one-liner로 호출했다.

- 시간당 60회(IP당) 초과 시 API가 `tag_name` 없는 JSON 에러로 응답 → 빈 결과 →
  "No releases found"로 사망 (실제로는 릴리즈가 존재함)
- 리스트의 첫 항목은 가장 최근 *생성* 릴리즈이므로 RC를 최신 stable로 잘못
  잡는 결함도 공존 (`/releases/latest`는 draft/prerelease를 제외한 진짜 최신)

## 변경

### install.bat

버전 해석 one-liner를 리다이렉트 계약으로 교체했다
(install.sh `get_latest_version` / install.ps1 `Get-LatestVersion`과 동일 계약):

1. `HEAD https://github.com/modu-ai/moai-adk/releases/latest` (10초 상한,
   리다이렉트 최대 10회)
2. 최종 URI는 에디션별로 다른 속성에 노출 — PS6+는
   `BaseResponse.RequestMessage.RequestUri`, Windows PowerShell 5.1은
   `BaseResponse.ResponseUri` — 둘 다 탐침 (ps1과 동일)
3. `/tag/` 뒤 세그먼트 추출 → `go-` 접두사 제거 → `v` 접두사 제거
4. 실패 시 sh/ps1과 동일한 안내: "Could not determine the latest version from
   GitHub / GitHub may be unreachable from this network."

설계 결정:

- **native cmd curl 대신 PowerShell one-liner 유지** — 파일의 기존 관례(모든
  네트워크 동작이 PowerShell 위임), 헤더의 "Requires Windows 7 or later"
  바닥(curl.exe는 Win10 1803+ 번들), 그리고 cmd 이스케이프 노출면 최소화
- **페이로드에 파이프·리다이렉션 0개** — 인용부 안에서 캐럿 이스케이프가
  필요한 문자를 원천 배제. 남은 이스케이프는 `2^>nul` 하나(기존 라인과 동일
  위치). 괄호/캐럿의 인용부 내 사용은 기존 라인이 이미 파싱 가능함을 보이는
  형태만 사용
- `-NoProfile` 추가 — 사용자 PowerShell 프로필이 stdout에 배너를 출력하면
  `for /f` 캡처를 오염시켜 VERSION에 쓰레기가 들어가는 사고 예방 (데이터를
  캡처하는 호출에만 적용)
- TLS1.2 사전 고정 — 파일 내 다운로드 라인들의 기존 관례와 동일한 표현으로,
  Win7/PS5.1 조합에서의 협상 실패 예방

### .github/workflows/test-install.yml (test-bat 잡)

1. **정적 가드 (신규 단계)**: install.bat이 `/releases/latest` 리다이렉트 URL을
   포함하는지 확인 — REST API로의 회귀 방지
2. **엔드투엔드 실측 (신규 단계, 판정 경로)**:
   `call install.bat --install-dir "%RUNNER_TEMP%\moai-install-test"` — 실제
   네트워크 버전 해석 → 다운로드 → 체크섬 검증 → 압축 해제 → 설치 →
   `moai version` 실행까지 전부 실제 호출. cmd.exe 인용/for-f 이스케이프는
   macOS에서 검증 불가하므로 이 단계가 배치 표면의 판정 수단이다. 사용하는
   엔드포인트 전부 비인증·비레이트리밋(리다이렉트 + 릴리즈 브라우저 경로)이라
   토큰도, 공유 API 예산 소진도 없음

최종 판정은 릴리즈 PR 시점의 전량 매트릭스가 확정한다 (카드 지시와 동일).

## 검증 (로컬, macOS)

- **Claim**: 커밋된 페이로드가 실제 GitHub에서 최신 버전을 해석한다
- **Evidence**: 파일에서 직접 추출한 페이로드(`sed`로 `powershell
  -NoProfile -Command "…"` 인용부 발췌 — 재타이핑 사본 아님)를 pwsh 7.5.4로
  실행 → 출력 `3.1.0`, exit 0. PS6+ `RequestMessage` 분기 실측
- **YAML**: `python3 -c "import yaml; yaml.safe_load(open(...))"` → OK
- **잔존 스캔**: install.bat의 `api.github.com`은 REM 주석(설명문) 1곳만,
  `for /f`는 파일에 정확히 1개
- **Gaps**: cmd.exe 이스케이프와 PS 5.1 `ResponseUri` 분기는 로컬 미검증 →
  CI test-bat 엔드투엔드가 판정 (windows runner의 `powershell` = 5.1이므로
  정확히 그 분기를 실측)
- **Residual-risk**: 리다이렉트 엔드포인트가 API rate-limit 대상이 아니라는
  전제는 install.sh:74-83의 근거와 동일한 것에 기반. GH runner 공유 IP에서
  실제 동작은 dispatch 실행으로 확인 예정

## 후속 발견: LF 줄바꿈 결함 (첫 CI run 31960480175 실패 → 진단 → 수정)

t53 통합 직후 `test-install.yml`을 release ref에 dispatch했다 (run
31960480175). 결과: test-bat만 실패, 나머지 8개 잡 전부 통과. 로그 패턴:

- `'--install-dir' is not recognized` — :parse_args 괄호 블록 미파싱
- 배너 박스 문자 토큰들이 명령로 실행 시도 — 줄 경계 desync
- `[INFO] Fetching...` 누락 + `!VERSION!` 리터럴 — 버전 fetch 블록 통째로 스킵
- 최종 `Download failed` — `!VERSION!` 리터럴이 URL에 박힘

원인: `.gitattributes`의 전역 `* text=auto eol=lf`(11행)가 `*.bat` 예외 없이
적용되어 install.bat가 **LF로 저장·배포**되고 있었다. cmd.exe는 다중 줄
괄호 블록 파싱에 CRLF를 요구한다. 이전 CI는 정적 검사만 수행해 실제
실행 경로를 한 번도 검증한 적이 없어 미적발이었다 — 즉 카드 변경 이전부터
존재하던 결함이며, 엔드투엔드 단계가 설계대로 이를 적발했다.

수정:

- `.gitattributes`에 `*.bat -text` 예외 추가 (정규화 전면 비활성화)
- install.bat를 **CRLF 바이트로 커밋** — 클론 체크아웃, actions/checkout,
  raw.githubusercontent.com 다운로드까지 모든 배포 경로에서 CRLF가 전달된다
  (`text eol=crlf`는 인덱스가 LF로 남아 raw URL 경로가 여전히 깨진다)
- CRLF 변환 후 버전 해석 페이로드는 바이트 동일함을 확인
  (`cmp` — 589바이트 일치)

## 최종 판정 (2차 run 31960881976 — 전량 통과)

CRLF 수정 통합(`bcde9e570`) 후 재dispatch. **10개 잡 전부 success.**
엔드투엔드 단계 로그 (windows-latest, cmd.exe + Windows PowerShell 5.1):

```
[INFO] Detected platform: windows_amd64
[SUCCESS] Latest Go edition version: 3.1.0
[INFO] Downloading from: https://github.com/modu-ai/moai-adk/releases/download/v3.1.0/moai-adk_3.1.0_windows_amd64.zip
[SUCCESS] Checksum verified
[SUCCESS] Installed to: D:\a\_temp\moai-install-test\moai.exe
PASS: install.bat resolved, downloaded, and installed end-to-end
```

`3.1.0` 해석은 로컬 pwsh 7(`RequestMessage` 분기)과 CI PS 5.1
(`ResponseUri` 분기) 양쪽에서 각각 실측됐다. 1차 실패(31960480175)는 위에
기록한 LF 결함 — 엔드투엔드 단계가 노린 "검증 불가능한 표면의 기계 판정"이
설계대로 작동한 사례다. 정식 판정은 릴리즈 PR 시점의 전량 매트릭스.

## 프로세스 편차 (기록)

디스패치는 `EnterWorktree(t53)` 후 `git merge release/v3.1.1`을 지시했으나,
이 세션은 release-v311에 앵커된 상태라 `EnterWorktree(name)`의 신규 생성이
거부됐다. 대신 `git worktree add -b WT-t53 <경로> release/v3.1.1`로 생성 후
`EnterWorktree(<path>)`로 진입했다 — 프로토콜 표의 "Re-enter one from the
current session" 허용 형식이며, path 진입은 세션 종료 시 keep/remove 추적을
복원해 고아 트리 우려를 해소한다. 베이스를 release/v3.1.1로 직접 잡았으므로
지시된 병합 단계는 불필요해졌다 (결과 트리 내용 동일).
