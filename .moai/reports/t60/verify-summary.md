# t60 — 크로스세션 통신 설정 MoAI 설정 표면 + moai web 토글 — 검증 요약

작성: 2026-08-17 (인수 워커 2 — 원본 워커 429 사망, 런처층 인수 이어받음)
워크트리: `.claude/worktrees/agent-a88a53c67fbaa94fe` [WT-t60]

## 카드 범위 대비 완료 상태

- **(1) 기기간 전송 승인** — 원본 워커 완료 (인수 시 GREEN 상태):
  - 가드 테스트 2종: `TestTemplateNeverShipsIsolatePeerMachines` (배포 템플릿 전수
    스캔 — settings-bearing 파일에 `isolatePeerMachines` 부재 고정),
    `TestLauncherNeverInjectsIsolatePeerMachinesByDefault` (중립 설정에서 일반/
    칸반 어느 페이로드도 동 키를 싣지 않음 고정).
  - 문서: crosssession.yaml(로컬+템플릿 동일 바이트) 주석에 '기본은 승인 없음 +
    true는 어느 스코프에서든 적용되어 하위에서 되돌릴 수 없음 + 기기간 메시지는
    Anthropic 서버 경유' 명시. 카드 (1b) 요구 충족.
- **(2) moai web 토글** — 인수 워커 완료 (본 세션 작업):
  - `internal/config`: 공유 closed set 접근자 `ValidCrossSessionInboundValues` /
    `ValidCrossSessionDialogExpiryValues` (+ 멤버십 헬퍼). 런처 번역과 웹 select가
    같은 단일 선언에서 파생 (재선언 금지 doctrine).
  - `internal/cli/crosssession_settings.go`: 사설 맵을 공유 접근자로 교체.
  - `internal/settings`: `SectionCrossSession` + 3필드 (inbound select /
    isolate_machines bool / dialog_expiry select), `RouteSeam` 배선,
    `sectionRootKeys` 등록, `EmptySubmits` 필드 신설 — select의 "" 제출이 키를
    중립 ""로 되돌리는 "끄기" 경로 (없으면 토글의 off 방향이 no-op).
  - `internal/web`: 11번째 탭 "crosssession" + 제네릭 패널 + "다음 실행부터
    적용" 정직 노트, i18n 4로케일(en/ko/ja/zh) 키 21종, 탭 순서 가드 갱신.
  - `internal/settings/yamlpatch`: 부재 파일을 빈 문서로 시작 (첫 편집이 파일
    생성 — 로더의 greenfield 허용과 대칭).

## Claim (주장)

1. 런처층(config 로더 + cli 번역 + 칸반 병합) 테스트 통과 — 원본 워커 인수분.
2. settings 스키마/seam/web 콘솔 계층 테스트 통과 — 본 세션 구현분.
3. `EmptySubmits` 되돌림 라운드트립: inbound=hold 기록 → "" 제출 → 로더가
   `Inbound=""` 로드 (주석 보존).
4. 비브라우저 게시(키 부재)는 여전히 preserve — atomic Save 계약 불변.
5. Windows 크로스빌드 + vet 통과. `make build` (templ 재생성 포함) 성공.

## Evidence (증거)

- `green-settings.txt` — `go test -count=1 ./internal/settings/...` →
  `ok ... settings 2.020s / agentfm 1.363s / yamlpatch 1.921s` (RC=0)
- `green-web.txt` — `go test -count=1 ./internal/web/ -run "TestCrossSession|
  TestParseSchemaFormEmptySubmits|TestConsoleTabsOrder|TestEveryTabHasAPanel|
  TestAgentFM"` → `ok ... 0.649s` (RC=0)
- `web-preexisting-failures.txt` — web 전체 수트: 실패 2건은 선행 결함(아래 Gaps).
- cli — `go test ./internal/cli/` → `ok ... 298.285s` (본 세션 관측; 이후 cli
  코드 무변경)
- config — `go test ./internal/config/` → 유일 실패 `TestAlwaysLoadedTokenBudget`
  (선행 결함, 아래 Gaps). crosssession 로더/감사 레지스트리 테스트 전부 통과.
- Windows: `GOOS=windows go build ./internal/settings/ ./internal/web/
  ./internal/config/ ./internal/cli/ ./internal/settings/yamlpatch/` → OK.
- `go vet` 동일 5패키지 → OK.

## Baseline-attribution (baseline 귀속)

- 워크트리 WT-t60 (HEAD 5c3141372), 미커밋 작업 트리 상태에서 위 커맨드 실행.
- 원본 워커분 증거: `red-config.txt`/`green-config.txt` (06:24), `red-cli.txt`/
  `green-cli.txt` (06:26-28) — 본 세션에서 파일 확인 및 현재 트리에서 재검증.

## Gaps (미검증)

1. **[선행 결함 — 본 카드 비관여] web i18n 2건 실패**: `TestI18nKeySetParity` /
   `TestDataI18nKeysSubsetOfDictionary` — `f.mcp.tools.glm_{task,job_status,
   job_result,job_cancel}.enabled.{title,desc}` 8키가 4로케일 전부에 부재.
   `git merge-base --is-ancestor 9865e87ed HEAD` = yes — glm_task 위임 도구
   패밀리 커밋이 카탈로그에는 도구를 추가했으나 i18n 키를 동반하지 않음.
   t60 diff에 `internal/mcp` 미포함 — 본 카드 이전부터 실패. 리드 라우팅 필요.
2. **[선행 결함 — 본 카드 비관여] config 예산 가드 실패**: `TestAlwaysLoadedTokenBudget`
   — always-loaded 표면 76,586 > 예산 76,000 (586 초과). 본 카드 diff에 `.claude/`
   룰 파일 0건 (측정 대상 아님). 최근 릴리스 레인 병합(t96/t92·t110 계열) 여파로
   추정 — t110(kanban-dispatch 3원칙) 착지 시 함께 정리 권장.
3. 실제 브라우저에서의 콘솔 조작(수동 E2E)은 미실행 — 위젯 렌더는 HTML 프로브
   테스트로 고정.
4. 칸반 모드 실기기 병합(accept 강제 + 사용자 extras)의 런타임 동작은 단위
   테스트 수준 — 실제 칸반 세션 발사 미검증.

## Residual-risk (잔여 위험)

- `moai update`는 `.moai/config` 관리 뿌리를 통째 삭제 후 재배포(§2.3) —
  crosssession.yaml의 사용자 편집(웹 콘솔이 쓴 값 포함)은 update 시 유실 가능.
  이는 기존 모든 섹션 yaml과 동일한 구조적 위험이며 t111(선백업-후삭제) 계열의
  소관 — 본 카드에서 새로 만든 위험 아님.
- EmptySubmits는 crosssession 2 select에만 적용 — 향후 다른 select가 "끄기"
  를 원하면 필드별 옵트인 필요 (의도된 최소성).
- 브라우저 폼은 select를 항상 제출하므로, 값이 ""인 상태에서 다른 패널만 저장해도
  `inbound: ""` 재기록(무해 no-op)이 발생할 수 있음 — 최초 1회 파일 생성 경로와
  함께 정상 동작.
