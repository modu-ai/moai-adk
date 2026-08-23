# Plan — SPEC-CODEX-WIRING-001

> Tier M 구현 계획. 마일스톤은 **되돌릴 수 없는 정도 내림차순**으로 정렬했다 — 데이터 계약(산출물
> 형태·병합 모델)이 최상단, 기계적·부가적 단계가 최하단. 검토는 위에서부터.

## §A. Context

- 카드 t88 (Class C) / 에픽 Codex dual-harness M4. 목표: `moai init --agent claude|codex|both`로
  `.codex/hooks.json` + `.codex/config.toml`(mcp_servers.moai, writes 승인) 배선, codexadapter의
  첫 프로덕션 호출자 확보, 신뢰 안내 + doctor 점검.
- **베이스 제약 (§A.1 of spec.md)**: 작성 시점 origin/main(76b2c4ece)에는 듀얼 하네스 코어가
  전무하다 — `internal/codexadapter`·`.codex` 템플릿 전부가 미병합 PR #1602(release/v3.1.3)에
  있다. run-phase는 #1602 병합 후 베이스에서 시작해야 한다(§C Pre-flight #1).
- 구 블로커(프로젝트 hooks.json 미발화)는 운영자 실측으로 폐기(spec §A.3). 남은 차단은 배선 부재뿐.
- t187 정합성 결론(spec §A.4): `writes`는 capability 기반 → config에 도구명 열거 없음,
  병합 순서 독립, annotation 가드 테스트가 불변식.

## §B. Known issues carried into this plan

1. **codexadapter 호출자 0** — 본 SPEC이 첫 호출자. EventTable 소비 방식(리플렉션 아님, 패키지
   import)을 M1에서 확정한다.
2. **go.mod에 TOML 파서 없음** — agentemit이 손레nder링하는 것과 동일하게 최소 문법만
   렌더/탐지한다. 신규 의존성 추가 금지(사다리 규율).
3. **`.codex/`는 ManagedCleanTargets에 없음**(감사 §2.3) — update가 배선 파일을 지우지 않는다.
   존재 기반 opt-in(REQ-CW-009)은 이 부재와 정합.
4. **`features.hooks` 기본값 미확정**(spec §H) — run-phase 검증 항목 #1로 승계.

## §C. Pre-flight (Run-Phase Entry Checks)

1. `git ls-tree <run-base> --name-only internal/ | grep -q codexadapter` → 참. (PR #1602 병합 확인.
   거짓이면 run 중단 후 리드 보고 — 이 SPEC은 어댑터 없이 구현 불가.)
2. `go build ./...` green on run-base.
3. `go test ./internal/codexadapter/...` 전량 통과(M3 인수 상태 확인).
4. `.moai/reports/t83/precondition-measurement-round3.md` §4(화이트리스트 근거)의 행 번호가
   run-base와 일치하는지 육안 확인 — 인용이 아니라 원칙(최상위 {description, hooks})만 가져간다.
5. 검증 환경: 격리 `CODEX_HOME` 확보(t83 위생). 실제 `~/.codex/` mtime 무변경 전후 확인.

## §D. Milestones

### 선행 결정 블록 (마일스톤보다 먼저 확정 — 되돌림 비용 최대)

- **D1 hooks.json 데이터 계약**: 이벤트 키 PascalCase 6종(EventTable 유도), handler
  `{type:"command", command:"moai hook <arg> --harness codex", timeout:N}` (SessionEnd ≤3,
  타 이벤트 상수 — 제안 10, t83 관측 동작 사례와 동일), MoAI 관리 식별자 = command 접두
  `moai hook `. 최초 생성 시에만 MoAI description 부여(사용자 description 있으면 보존).
- **D2 병합 모델**: hooks.json = read → 낡은 MoAI handler 제거 → 현재 표 append → 원자적 쓰기.
  config.toml = `[mcp_servers.moai]` 테이블 create-if-absent만, 존재 시 미수정(doctor drift 보고).
- **D3 `--agent` 의미론**: claude(기본)=오늘과 동일(+.mcp.json 기본 provisioning),
  codex=.mcp.json provisioning 스킵 + Codex 배선, both=양쪽. 플래그가 wizard 답변에 우선.
  이 결정은 플래그 파싱 한 곳에 고립시켜 되돌림 비용을 1줄로 유지한다.

### M1 — 생성기 코어 (internal/codexwiring, 데이터 계약 구현)

- 패키지 신설: hooks.json 렌더러(EventTable 유도·결정론적 키 순서), 병합 엔진(D2),
  ValidateConfig 선검증 게이트(위반 시 미기록·중단), config.toml 테이블 렌더러(손레nder링,
  create-if-absent), 신뢰 사이드카(`.moai/state/codex-wiring.json`, sha256 기록).
- 단위 테스트: 화이트리스트 음성 샘플(version 키 등), 병합 보존, 멱등성(2회 렌더 바이트 동일),
  SessionEnd 타임아웃 상한, 결정론(키 순서·무타임스탬프).

### M2 — `--agent` 플래그 + init/update 배선

- `internal/cli/init.go`: 플래그 등록(`init_workflow_flags.go`의 opt-in tracker 패턴과
  `autonomy-tier`의 폐쇄집합 fail-loud 패턴 준용), `validateInitFlags` 검증, runInit 꼬리의
  `provisionMCPEntryUnlessDeclined` 인접 호출점에 배선 호출 배치.
- `.mcp.json` 게이팅(D3): `--agent codex`에서 declined 취급.
- `moai update`: 배선 파일 존재 시에만 갱신(REQ-CW-009), 내용 변경 시 재신뢰 안내.
- 안내 문구 상수화: 최초 신뢰 안내 + `/hooks to re-trust`(AC-CW-008 토큰).

### M3 — 런타임 `--harness codex` 모드

- `internal/cli/hook.go`: dispatcher 부속명령에 `--harness` 플래그. 모드 on 시
  MapOutput 래핑 → RecordDiscards 기록 → payload `hook_event_name` Resolve 일관성 검증.
  exit code·stderr 패스스루. `internal/hook/` 무변경(seam은 cli 계층).
- 테스트: canned payload(t83 골든 형식)로 `continue:false`→`decision:block` 치환,
  빈 reason 기본문구, systemMessage 폐기 기록(.moai/logs/codex-adapter.jsonl),
  이벤트 불일치 거부, exit code 통과.

### M4 — doctor "Codex Wiring" 진단

- `internal/cli/doctor_codex.go` 신설 + `runGroupedChecksObserved` workspaceChecks에 등록
  (`checkBinaryFreshness` t184 선례 형태 — advisory·fail-open).
- 검증: 파일 존재·ValidateConfig·사이드카 해시 divergence(`/hooks to re-trust` 안내)·
  `moai` PATH 해석·config 테이블 존재/일치. 비활성 프로젝트: 정보성 스킵.

### M5 — 검증/e2e + 인수 정리

- acceptance.md AC 전량 실행(스크래치 init 시나리오 포함), 커버리지 측정
  (`go test ./internal/codexwiring/... -cover` ≥85%), `GOOS=windows GOARCH=amd64 go vet ./internal/...`
  (크로스컴파일 게이트 — 로컬 전체 스위트 금지 규율 준수, affected-package만).
- annotation 가드 테스트(REQ-CW-011)를 M5가 아니라 **M2와 함께** 추가할 수도 있다 — catalog와
  mcp_server.go가 같은 패키지라 즉시 가능. 구현자 재량(AC-CW-012는 마일스톤 소속을 묻지 않음).

### PRESERVE 목록 (변경 금지 — 감사·카드 지정)

- `internal/cli/mcp_server.go`의 codex_task/codex_job_* · glm_* 핸들러·등록 전체 — 무변경
  (annotation 가드 테스트 추가만 허용, 선언 자체는 무수정).
- `internal/template/templates/.codex/agents/**` (M5 agentemit 산출물) — 재생성·수정 금지.
- `internal/codexadapter` 공개면 — 소비만, 수정 금지(M3 인수물).
- 기존 init 플래그 전량·플래그 부재 시 동작 — 바이트 수준 동일성(AC-CW-004).
- `.claude/` 템플릿·hook 래퍼 — 무변경. template 추가 발생 시에만 Template-First + `make build`.

## §E. Self-verification (run-phase가 §E.2/§E.3에 기록할 목록)

- AC 12/12 MUST 결과 + 명령 원문.
- `go test ./internal/codexwiring/ ./internal/cli/ -cover` 수치.
- `golangci-lint run` + `GOOS=windows GOARCH=amd64 go vet ./internal/...` 결과.
- 스크래치 e2e: init(--agent codex) → 재실행 no-diff → 사용자 엔트리 보존 → doctor 표시.
- template 무변경 확인: `git diff --stat internal/template/templates/` (공백 예상).

## §F. Anti-patterns

- **측정 안 된 config 키 주입** — `features.hooks`·`version`·handler 부가키(statusMessage 등)
 emit 금지. Codex의 무음 사망 계열을 우리 손으로 재현하지 않는다.
- **도구명 열거** — config.toml에 enabled_tools/default_tools_approval_mode 이외의 도구 목록
  남기는 것(§A.4 결론 위반). 카드의 6도구 나열을 문서가 아니라 config로 옮기는 실수.
- **사용자 hooks.json 재작성** — merge가 아니라 overwrite하는 어떤 경로. 레이어는 병합 실행된다.
- **init 실패 전파** — 배선 오류로 init 전체 실패. 경고-후-계행(단, 검증 실패 파일은 미기록).
- **`internal/hook/` 침입** — dispatcher 로직 안으로 어댑터를 넣는 것(M3 REQ-7 위반).
- **로컬 전체 스위트** — affected-package만; 전수 판정은 CI(2026-08-15 부하 사고).

## §G. Cross-references

- spec.md §A(측정 전제)·§D(REQ-CW-001..012) — 본 계획의 요구 원천.
- acceptance.md — AC 매트릭스·엣지·DoD.
- hns-moaiadk-patterns (Template-First·add-a-hook·CLI 아키텍처).
- 3-phase close 계획: **단일 sync 커밋**이 `implemented → completed` 전이를 carrying
  (Close-subject full-ID: `chore(SPEC-CODEX-WIRING-001): … 3-phase close`).
  progress.md §E.4의 `sync_commit_sha`는 `pending-backfill-` placeholder → 후속 커밋 backfill
  (D3 면제). **§E.5는 생성하지 않는다**(레거시 Mx-phase 마커 — era 오분류 유발).
