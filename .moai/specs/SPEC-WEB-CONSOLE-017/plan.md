# SPEC-WEB-CONSOLE-017 — Plan

카드 t1051 (issue #1709) — web save-failure observability. plan-phase 산출물; `status: draft`.

---

## §A. Context

- **작업 위치(워크트리)**: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1051`, 브랜치 `WT-save-observability` @ `616f7451d` (2026-09-22 SPEC 작성 시점; 커밋 직전 재판정 의무는 §C).
- **산출물 경로**: `.moai/specs/SPEC-WEB-CONSOLE-017/{spec.md, plan.md, acceptance.md, progress.md}` (Tier M 세트 + progress.md).
- **근거 계측**: 이 트리에서 본 레인이 재측정한 앵커 — `handlers.go:497,518,526,535,545,555,565,578,591`(9 seam), `:644`(renderErrorPage 정의), `shell.templ:274`, `settings_shell.go:40-42,63-72`, `root.templ:29-34,48-56`, `fieldsets.templ:607`, `server.go:252/:290`, `console.css:155`. 세부는 spec.md §1.1/§5.
- **기존 인프라**: PRESERVE 대상 `recordingSeams`(`internal/web/partial_apply_repro_test.go` — 주입 하니스, 형제 SPEC-WEB-CONSOLE-016 REQ-WC16-010 이 본 카드의 예속 경로로 명명). EXTEND 대상: 동일 하니스에 stderr 관측 축을 얹는다.
- **plan-auditor 판정**: 미실시(plan-phase 신규). 다음 관문에서 Tier M 임계 0.80.

## §B. Tier 결정 + 정당화

**결정: Tier M.** (frontmatter `tier: M`)

- LOC/파일 수 가이던스만 보면 Tier S 영역이다(예상 < 300 LOC, 5파일 미만 — `handlers.go` 로깅 배선 + 수송 응답 변경 + 테스트). 그러나 Tier 규정은 LOC 를 guidance-not-enforcement로 명시하고, 아래 세 이유가 acceptance.md 를 독립 파일(Tier M 세트)로 요구한다:
  1. **두 개의 독립 요구 표면**(REQ-A 사용자 / REQ-B 유지보수자)을 리드 지시로 절대 병합 금지 — 각 표면의 인수가 별도 추적을 갖는다.
  2. **공허-참 함정을 지닌 회귀 가드** — CSS 클래스 단정은 `fieldsets.templ:607` 이 사전에 같은 클래스를 렌더하므로 제출 전부터 참이다. 본문-문구 판별식(HARD-4)과 그 근거의 추적성이 acceptance.md 의 §D.2 반증력 규칙에 기록돼야 한다.
  3. **자격증명 비-유출 경계**(HARD-3, REQ-WC-017-005) — sentinel 주입 검증이 별도 AC 로 열거돼야 한다.
- 예산 대비: REQ 6개 / AC 5개 — Tier M 천장(각 16) 이내.
- 대가: plan-auditor 임계가 0.75(S) → 0.80(M). 이는 산출물 세트 정합성의 비용으로 수용한다.

## §C. Pre-flight (run-phase 진입 전 재측정 의무)

```bash
git branch --show-current && git rev-parse --short HEAD   # 배차문 값 재판정
grep -n 'a.renderErrorPage(' internal/web/handlers.go      # 9 call site 좌표 재확인
grep -rn 'os.Stderr' internal/web --include='*.go' | grep -v _test   # stderr 관용구 현황
grep -rn 'log/slog' internal/web --include='*.go' | wc -l  # 0 유지 확인 (§F: 로거 전환 금지)
ls internal/web/partial_apply_repro_test.go                # PRESERVE 하니스 존재
go test ./internal/web/...                                  # 변경 대상 패키지 기준선
```

전체 스위트(`go test ./...`)를 로컬에서 돌리지 않는다 — 레인 규율(CLAUDE.local.md §4). 기준선과 새 결함은 §E로 구분 보고한다.

## §D. Constraints (HARD)

- **PRESERVE**: `recordingSeams` seam 목록·시그니처(HARD-2) / 9 seam 의 순서·개수·존재(HARD-1) / `server.go:252` 의 기존 문구(HARD-5) / `root.templ:26-28` 단일 강조 설결 결정 주석(HARD-6 — 그 결정 자체를 건드리지 않는다) / 실패 문구의 Go 리터럴 관례(HARD-7).
- **금지**: stderr 라인에 `err.Error()` 포함(REQ-WC-017-005) / 어떤 새 표면에도 자격증명 값 포함(HARD-3) / CSS 클래스 기반 가드 단정(HARD-4) / `log/slog` 도입(§F) / REQ-A·REQ-B 를 하나의 REQ 로 합치기 / 9 seam 외 임의 경로에 로깅 확장 / `--no-verify` · 강제 push.
- **요구**: Conventional Commits + `🗿 MoAI` 트레일러 / 가드는 본문 문구 판별식(HARD-4) / M1에서 수송 기구 결정의 근거를 plan.md 회신 커밋 본문에 남긴다.

## §E. Self-Verification (각 마일스톤 GREEN 기준)

각 항목은 VCI 5-섹션 형식(Claim / Evidence / Baseline-attribution / Gaps / Residual-risk)으로 보고한다. 명령·관측 출력·HEAD SHA 를 매 항목에 귀속한다.

- **E1** AC PASS/FAIL 매트릭스 — `acceptance.md` §D.1의 5개 AC 전부, 검증 명령과 관측 출력 동반.
- **E2** 크로스 플랫폼 빌드 — `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` (exit 0).
- **E3** 커버리지 — `go test -cover ./internal/web/...` (패키지 목표 85%).
- **E4** 경계 grep — 새 stderr 라인에 `err.Error()` 0행, 자격증명 토큰 0행; `log/slog` 임포트 0행 유지.
- **E5** lint — `golangci-lint run` (기준선 대비 NEW 0).
- **E6** 커밋 HEAD + push 상태 — 본 저장소 레인 규율상 push는 리드 일괄(2026-09-02); 로컬 병합 SHA 보고까지가 레인 소관.
- **E7** blocker 보고(해당 시) — §F.1 운영자 대기 질문의 답이 M1 착수 전에 필요해지면 중단하지 말고 기구를 결정 없는 쪽으로 고정 가능한지 먼저 판정한다(HARD-6 — 가능. 따라서 이 질문은 blocker 가 아니다).

## §F. Milestones (되돌리기 비싼 결정 우선; 시간 추정 없음)

- **M1 — REQ-A 사용자 표면(수송 기구 결정)**: 실패 사유가 인라인 슬롯에 도달하게 한다. 이 마일스톤이 **가장 변경 가능성 높은 결정**(비-2xx 수송을 어떻게 바꾸는가 — 2xx 재응답+스왑 대상 지정 / 스왑 헤더 / `htmx:responseError` 핸들러)을 품는다. 구현 후 되돌리기가 가장 비싼 UX 흐름이므로 선두. AC-WC17-001, AC-WC17-002(inline 축), AC-WC17-005. `spec.md` §5.3의 기존 데이터 경로를 재사용하고 새 데이터를 만들지 않는다.
- **M2 — REQ-B 유지보수자 표면(stderr 층 로그)**: 단일 헬퍼(접두어 `moai web: ` 고정, `err.Error()` 미포함)를 만들고 9 seam call site에 배선한다. 결정이 이미 spec.md §5.2로 굳어 있어 M1보다 되돌리기 싸다. AC-WC17-002(stderr 축), AC-WC17-003, AC-WC17-004.
- **M3 — 가드·하니스 확장(기계적)**: `recordingSeams` 확장(HARD-2), 본문-문구 가드 — **양방향**: 제출 전 DOM에 어떤 seam 실패 문구도 없음 ∧ 강제 실패 후 해당 문구 존재(한 방향만 돌리면 공허 참/눈먼 패턴을 못 가른다). AC-WC17-001의 사전-부재 절, AC-WC17-005 비-회귀 재확인.

## §F.1 Operator-pending (non-blocking)

- **질문**: `banner--error` 를 신설해 오류/경고 배너 변형을 분리할 것인가? — `root.templ:26-28`의 단일 강조 설계 결정을 뒤집는지 여부. 리드가 운영자에게 상신했다.
- **본 SPEC 의 의존성: 없음(HARD-6).** 가드는 본문 문구 판별식이라 어느 답에서도 성립한다. 답이 「신설」이면 표현 계층의 후속 작업이 생길 뿐이다. `[NEEDS CLARIFICATION]` 마커를 붙이지 않는다 — 이 질문은 Kickoff 게이트를 막지 않는 것이 의도다.

## §G. Anti-Patterns (회피 목록)

- **클래스 기반 가드** — `banner--warn`/`banner--error` 단정은 공허하거나(사전 존재) 존재하지 않는다(HARD-4). 판별식은 문구 본문.
- **err.Error() 로깅** — 값이 자격증명 파편을 품을 수 있다(REQ-WC-017-005). 층 식별 + 안정 문구만.
- **REQ-A·REQ-B 병합** — 두 독자, 두 표면. 하나의 REQ 로 합치면 유지보수자 축이 사용자 축의 UX 논쟁에 인질로 잡힌다.
- **9 seam 외 확장** — 검증/거부(`:456` atomic reject)·프로필 rename/delete 등 다른 실패 경로로 로깅을 넓히는 것은 범위 침범이다.
- **접두어 혼용** — 새 줄에서 `web: ` 와 `moai web: ` 을 섞는다(REQ-WC-017-004).
- **하나의 방향만 돌리는 가드** — 사전-부재 또는 사후-존재 한쪽만 단정하면 공허 참과 눈먼 패턴을 구별 못 한다(§F M3).
- **전체 스위트 로컬 실행** — 레인 규율 위반; 판정은 패키지 테스트 + CI.

## §H. Cross-References

- `spec.md` §5.1 (9 seam 표) / §5.2 (접두어 결정) / §5.3 (기존 데이터 경로) — M1·M2 의 구현 입력.
- `acceptance.md` §D.1-§D.4 — AC 전문, 반증력 규칙, REQ↔AC 추적성, DoD.
- SPEC-WEB-CONSOLE-016 — 형제. `recordingSeams` 예속(REQ-WC16-010), 비밀 경계(REQ-WC16-005), 배너 문구 정밀도(REQ-WC16-006/007 — 미착수, §6-3 연동).
- SPEC-WEB-CONSOLE-011 — 콘솔 스코프 계약(10섹션 쓰기 경계; 본 SPEC 은 그 경계를 변경하지 않는다).
- SPEC-GLM-KEY-INPUT-001 / SPEC-JEV-OPTIN-MEASURE-001 — 자격증명 저장 seam(REQ-GKI-004-003: 오류 메시지 무-키-물질)의 원 REQ.
- `.claude/rules/moai/development/verification-completeness.md` — AC 채용 2-셀 규율(RED-now + green path)과 regression-guard 분류(§2.1 undecidable disposition)의 근거.
