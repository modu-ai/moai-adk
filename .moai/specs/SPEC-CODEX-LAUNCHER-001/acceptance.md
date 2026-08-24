# SPEC-CODEX-LAUNCHER-001 — 인수 기준

모든 기준은 기계 판정이다. codex 실 바이너리에 의존하는 항목은 없다 — 프로브 seam (`codexLookPath`, `codexCommandRunner`) 을 스텁해 판정한다.

## AC-CL-001 — 커맨드 등록 (REQ-CL-001)

- **Given** 빌드된 moai 바이너리
- **When** `moai --help` 를 실행하면
- **Then** LAUNCH COMMANDS 그룹에 `codex` 행이 나타나고, `moai codex --help` 가 rc 0 으로 도움말을 출력한다.

## AC-CL-002 — 동사 라우팅 (REQ-CL-002)

- **Given** 등록된 `codex` 커맨드
- **When** `status` / `app` / 맨몸 각각을 스텁 exec seam 으로 호출하면
- **Then** `status` 는 exec 을 0회 수행하고 리드아웃만 내며, `app` 은 `codex app` 인자로 정확히 1회 exec 하고, 맨몸은 §B 확정 결과에 따른 동작을 한다.
- **평가 조건**: 맨몸 동작의 기대값은 plan §B 결정 (a)/(b) 확정 후 확정된다. 확정 전에는 이 절의 세 번째 절만 미정이며 앞의 둘은 확정 기준이다.

## AC-CL-003 — `--spawn` 패리티 (REQ-CL-003)

- **Given** tmux 부재를 흉내 낸 환경
- **When** `moai codex --spawn` 을 실행하면
- **Then** `moai cc --spawn` 이 같은 조건에서 내는 것과 동일한 계열의 진단으로 실패하고, codex 바이너리는 exec 되지 않는다.

## AC-CL-004 — 리드아웃 행 집합 (REQ-CL-004)

- **Given** 스텁 프로브가 설치됨/버전/auth 를 공급하는 상태
- **When** `moai codex status` 를 실행하면
- **Then** 출력에 바이너리 경로 · 버전 · CODEX_HOME · auth · 배선 상태 다섯 항목이 각각 한 행 이상으로 존재한다 (행 라벨 grep 5건 전부 히트).

## AC-CL-005 — CODEX_HOME 해석과 출처 표시 (REQ-CL-005)

- **Given** `CODEX_HOME=/tmp/xyz` 가 설정된 프로세스
- **When** 리드아웃을 조립하면
- **Then** 값은 `/tmp/xyz`, 출처는 `env` 로 보고된다.
- **And Given** `CODEX_HOME` 미설정
- **Then** 값은 `<home>/.codex`, 출처는 `default` 로 보고된다.

## AC-CL-006 — 배선 없는 프로젝트는 정보성 (REQ-CL-006)

- **Given** `.codex/` 가 없는 임시 프로젝트 루트
- **When** `moai codex status` 를 실행하면
- **Then** rc 는 0 이고, 배선 행은 미배선 상태로 적히며, 출력에 `moai init --agent codex` 문자열이 포함된다.

## AC-CL-007 — 분류 구현 단일성 (REQ-CL-007, REQ-CL-010)

- **When** `grep -rn "login status" internal/ --include="*.go" | grep -v _test` 를 실행하면
- **Then** 히트는 `classifyCodexAuth` 한 곳뿐이다 (신규 런처 코드에 두 번째 분류 경로가 없다).
- **And** `internal/web` 에는 auth 분류 로직이 여전히 0건이다.

## AC-CL-008 — stderr 로 나오는 상태 문구를 분류한다 (REQ-CL-008) [핵심]

- **Given** stdout 은 비우고 stderr 에만 `Logged in using ChatGPT` 를 쓰는 스텁 러너
- **When** `ProbeCodexSetup` 을 호출하면
- **Then** `AuthProvider == "chatgpt"` 이다.
- **기준선 근거**: 이 시험은 수정 전 트리에서 반드시 실패해야 한다 (현행은 `unknown`). 실패를 먼저 관측한 뒤 수정한다.

## AC-CL-009 — 비영 rc 라도 출력이 있으면 분류를 시도한다 (REQ-CL-008)

- **Given** rc 1 과 함께 `Logged in using ChatGPT` 를 stderr 로 내는 스텁
- **When** 분류하면
- **Then** `chatgpt` 로 분류된다 (rc 만으로 조기 `unknown` 하강하지 않는다).

## AC-CL-010 — 판정 불가는 조치와 함께 보고 (REQ-CL-009)

- **Given** 두 스트림 모두 비어 있는 스텁
- **When** `moai codex status` 를 실행하면
- **Then** auth 행은 `unknown` 이고, 출력에 `codex login status` 문자열이 포함되며, 로그아웃 단정 문구는 없다.

## AC-CL-011 — `--version` 경로 무회귀 (REQ-CL-008)

- **When** 기존 codex 관련 시험을 실행하면 (`go test ./internal/cli/... -run Codex`)
- **Then** 전부 통과한다 — 러너 인터페이스에 메서드를 추가했을 뿐 `run` 의 계약은 바뀌지 않았다.

## AC-CL-012 — 데스크톱 앱 위임 (REQ-CL-011)

- **Given** exec seam 스텁
- **When** `moai codex app` 을 실행하면
- **Then** 정확히 `codex app` 이 호출되고, 앱 경로 탐색이나 설치 시도 코드는 실행되지 않는다 (`grep` 으로 `/Applications` 류 하드코딩 0건).

## AC-CL-013 — codex 부재 시 exec 없음 (REQ-CL-012)

- **Given** `codexLookPath` 가 실패하도록 스텁된 상태
- **When** 세 동사 각각을 실행하면
- **Then** 모두 비영 rc 로 종료하고, exec 호출 횟수는 0 이며, 진단은 설치 조치를 명명한다.

## AC-CL-014 — 쓰기 없음 (REQ-CL-013)

- **Given** 임시 프로젝트 루트 + 임시 CODEX_HOME 의 파일 목록·mtime 스냅샷
- **When** 세 동사를 (exec 스텁 상태로) 각각 실행한 뒤 다시 스냅샷하면
- **Then** 두 스냅샷이 동일하다 (`.claude/settings.local.json` 포함 무변경, CODEX_HOME 하위 신규 파일 0).

## AC-CL-015 — 중립성 (REQ-CL-014)

- **When** 템플릿 중립성 가드를 실행하면 (`MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/...`)
- **Then** 통과한다. 도움말 문안에 SPEC ID · 카드 id · 내부 날짜 · 커밋 SHA 가 없다.

## AC-CL-016 — 게이트 (전 REQ)

- **When** `go build ./...` · `go vet ./...` · `GOOS=windows go vet ./...` · `golangci-lint run` 을 실행하면
- **Then** 전부 rc 0 이다.

---

## 판정 제외 (근거 명시)

- **실제 Codex 앱 기동**: CI 러너에 데스크톱 환경이 없다. 운영자 수동 확인 항목으로 `progress.md` 에 남긴다 — 확인 방법은 배선된 프로젝트에서 `moai codex app` 실행 후 앱 전면 등장 여부.
- **실 바이너리 auth 왕복**: 로그인 상태는 머신 상태에 의존한다. 기계 판정은 스텁으로 하고, 실 바이너리 확인은 `moai codex status` 출력 1회를 `progress.md` 에 붙이는 것으로 갈음한다 (§A.2 의 기준선 측정과 대칭).
