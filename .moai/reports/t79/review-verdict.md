# t79 Review Verdict — FAIL (차단 1건, 수정 후 재리뷰)

- Card: t79 — glm_task MCP 도구군 신설 (commit `9865e87ed`, branch WT-t80, base 정합: ride-along = 구현+release 머지 2커밋만)
- Lens: `--deep --security` (sync-auditor 위임 + orchestrator 결정론 검증)
- Reviewer session: release-v311 (2acd4be4), 2026-08-17
- Full audit: `deep-audit-report.md` (동일 디렉터리, verbatim)

## Claim (주장)

구현이 계약(커밋 메시지·defaults.go 문서·codex 대칭)을 지키는가 — 특히
DefaultGLMTaskTimeout 600s가 sync·background 양쪽 glm_task 호출의 실질 상한인가.

## Evidence (증거)

### 오케스트레이터 결정론 검증 (직접 실행)

| Check | Command | Observed |
|---|---|---|
| 빌드 | `go -C <t80> build ./...` | BUILD OK |
| 테스트 | `go -C <t80> test ./internal/cli/ -run 'GLM\|TestMoaiMCPServer\|TestAC_C' -count=1` | ok 1.357s |
| 테스트 | `go -C <t80> test ./internal/mcp/ -count=1` | ok 0.250s |
| 동시성 | `go -C <t80> test ./internal/cli/ -run 'GLM' -race -count=1` | ok 3.491s |
| 룰 미러 | blob `cmp` | byte-identical |
| super-advisor 헌크 | added-lines diff | identical (전체 파일 드리프트는 기존 — 감사관 4-blob 대조로도 확인) |
| neutrality | template blob grep | 신규 위반 0 (`REQ-AA2-003`은 부모 4004a2a06에도 존재하는 기존 토큰) |
| 카탈로그 | catalog.go 열거 + 핀 테스트 diff | 21 엔트리·write-capable 6 — 코멘트·핀·실제 삼중 정합; glm_task=true, status/result=false, cancel=true |
| RED→GREEN | /tmp/t80-red.txt | 진짜 RED (심볼 미정의 빌드 실패) |

### F1 독립 재확인 (FAIL의 근거 — 감사관 보고를 믿지 않고 직접)

- `git show 9865e87ed:internal/cli/glm_task.go` → sync arm **:144** `callGLMTask(ctx, …)` **WithTimeout 없음**; background만 **:200** `WithTimeout(ctx, config.DefaultGLMTaskTimeout)`.
- `callGLMTask` 실행부 **:256** `glmHTTPClient.Do(httpReq)`; `mcp_glm.go` **:79** `glmAuditHTTPTimeout = 120 * time.Second`, **:90** `glmHTTPClient = &http.Client{Timeout: glmAuditHTTPTimeout}` — `http.Client.Timeout`는 본문 읽기 포함 전체 요청 상한 → **양 form 모두 실질 120s**.
- `git grep DefaultGLMTaskTimeout 9865e87ed` → **리더 1곳**(glm_task.go:200). sync에서는 죽은 상수, background에서는 120s에 그림자.
- `defaults.go:352` "bounds ONE glm_task call — **sync or background**" 문서와 모순.
- codex 대칭 주장 반박: `codex_task.go:88` `WithTimeout(ctx, config.DefaultCodexTaskTimeout)` — codex는 진짜로 적용.

## Baseline-attribution (baseline 귀속)

커밋 `9865e87ed` blob(부모 `4004a2a06`) 대비, 본 세션(t80 워크트리 `go -C`, 공유 ODB `git show`)에서 측정. -race 런은 리뷰 세션이 추가(디스패치에는 없었음).

## Gaps (미검증)

- 실제 z.ai 세대에서 120s 초과 빈도(체감 타격률) 미측정 — 계약 위반 자체로 차단 근거는 충분.
- mcp-go ctx 수명은 stdio 배포 기준으로만 정적 확인(현행 배포 형태와 일치).
- 전체 스위트는 CI 몫(레인 로컬 규율상 미실행).

## Residual-risk (잔여 위험)

- **수정 시 주의**: 클라이언트 Timeout만 올리면 sync arm은 상한을 아예 잃음 — WithTimeout 랩과 함께 납입 필수(감사관 지적 동의).
- F2(cancel-vs-complete RMW 창)·F4(orphan running 폴링)·F5(Output 비레택트)·F6(queued 고아)·F3(job_id 포맷 검증 부재)는 codex에서 상속된 대칭 한계 — 본 카드 비차단, **양 패밀리 공통 후속 카드** 권장.

## Verdict

**FAIL** — F1(차단): `DefaultGLMTaskTimeout`(600s)가 sync에서 미적용·background에서
120s audit 클라이언트에 shadowing. 커밋 메시지·defaults.go 문서·"codex 대칭" 주장이
구현과 모순. 수정 범위는 작음: ① sync arm WithTimeout 추가 ② 태스크 호출의 120s
지배 해소(태스크 전용 client seam 또는 Timeout 해제+ctx 바운드) ③ 관련 코멘트
정정 ④ 상수를 120s 미만으로 줄여 지배를 증명하는 테스트. 수정 커밋 후 delta 재리뷰.

나머지 11개 주장 그룹(자격증명 격리·원자적 쓰기·cancel 의미론·stale 거부·fail-open
완전성·고루틴 위생·카탈로그·미러·모델 SSOT 등)은 적대 검증 통과 — 구조·위생·테스트
품질 자체는 우수.
