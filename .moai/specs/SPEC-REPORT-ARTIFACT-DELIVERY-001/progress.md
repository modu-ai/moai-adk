# SPEC-REPORT-ARTIFACT-DELIVERY-001 — progress

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-10-02
tier: M
artifacts: spec.md, plan.md, acceptance.md, progress.md, decision-index.md
baseline_tree: 2e700a15e
notes: 리더 배차(카드 t1427, Class C). 카드 전제 4종 실측 검증 완료 — report.format 폐쇄 집합 위치(internal/settings/schema_sections.go:531), 미러 패리티(diff 0), Codex 도메인 스킬 배포 경로 부재(.agents 19종 커맨드 스킬만), 리더 결정 보고서 미독 갭(spec.md §A.3 명명). RED-now 근거는 acceptance.md §D 표.

## §E.2 Run-phase Evidence

### M1 — report.format 폐쇄 집합 + UI 전파 (2026-10-02, 본 워크트리)

RED 관측(구현 전, 커밋 3eb674ada 트리 — 테스트 선커밋 없이 워킹트리에서 선작성·관측):

```text
cmd:   unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/settings/ -run TestReportFormatClosedSetArtifact
stdout: --- FAIL: TestReportFormatClosedSetArtifact (0.00s)
         schema_sections_test.go:1031: reportFormatValues = [html+md md], want [html+md md artifact] (existing values preserved, artifact appended)
        FAIL    github.com/modu-ai/moai-adk/internal/settings   0.218s
exit:  1

cmd:   go test ./internal/cli/wizard/ -run TestReportFormatQuestion
stdout: questions_test.go:23: report_format should have 3 options, got 2
        FAIL    github.com/modu-ai/moai-adk/internal/cli/wizard 0.411s
exit:  1

cmd:   go test ./internal/web/ -run TestReportFormatRendersAsRadio
stdout: schemaform_test.go:109: report.format radio missing option "artifact"
        FAIL    github.com/modu-ai/moai-adk/internal/web        0.865s
exit:  1
```

GREEN 재측정(같은 명령, 구현 후 — AC-RAD-001/002/003의 GREEN 셀):

```text
cmd:   unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/settings/ ./internal/cli/wizard/
stdout: ok  github.com/modu-ai/moai-adk/internal/settings   1.072s
        ok  github.com/modu-ai/moai-adk/internal/cli/wizard 8.604s
exit:  0

cmd:   go test ./internal/web/   (i18n 거버넌스·옵션 desc 4로캘 가드 포함 전체 스위트)
stdout: ok  github.com/modu-ai/moai-adk/internal/web        58.788s
exit:  0
```

변경 파일: internal/settings/schema_sections.go (폐쇄 집합 3값 + artifact OptionDesc 분기),
internal/cli/wizard/questions.go (artifact 옵션), internal/web/assets/i18n.js (4 로캘 ×
label/desc 신규 2키 + sec.report.desc·fieldDesc 갱신),
internal/core/project/initializer.go (주석 열거 정합), 테스트 3종 갱신.
AC-RAD-013 보존 축: report.yaml 무변경(기본 html+md 불변), 기존 2값 순서·의미 보존.

### M2 — SKILL.md 전달 단계 3분기 + artifact-contract.md 신설 (2026-10-02, 커밋 7cf9130f1)

GREEN 재측정(AC-RAD-005/007/009 SKILL 축 — 커밋 7cf9130f1 트리):

```text
cmd:   grep -ci 'republish' .claude/skills/moai-domain-html-report/SKILL.md
stdout: 1        exit: 0   (EV-06 GREEN — ≥1)
cmd:   grep -c 'artifact link' .claude/skills/moai-domain-html-report/SKILL.md
stdout: 1        exit: 0   (EV-07 GREEN — ≥1)
cmd:   grep -c 'Artifact tool' .claude/skills/moai-domain-html-report/SKILL.md
stdout: 3        exit: 0   (EV-08 GREEN — ≥1)
cmd:   grep -c 'pre.mermaid' .claude/skills/moai-domain-html-report/SKILL.md
stdout: 2        exit: 0   (EV-13 GREEN — ≥1)
cmd:   wc -c .claude/skills/moai-domain-html-report/references/artifact-contract.md
stdout: 5809 .../artifact-contract.md        exit: 0   (EV-14 GREEN — 파일 존재,
         2-4 단어 타이틀 규칙·"이름: 설명" 금지 표·한 문장 description 마커 동반)
```

비고 — SKILL.md 증분이 초안에서 26,475B로 EV-15 상한(26,000B)을 넘어 D3에 따라 산문을
축약했다: 축약 후 재측정 `wc -c` → `25900` exit 0 (≤ 26,000 — AC-RAD-014 유지).
대형 독트린(계약 전문)은 references/artifact-contract.md로 이관(AC-RAD-007이 운반).

### M3 — 6개 mustache artifact 블록 + fonts.md (2026-10-02, 커밋 612ff3960)

GREEN 재측정(AC-RAD-006/008 템플릿 축 — 커밋 612ff3960 트리):

```text
cmd:   grep -c 'artifact:begin' .claude/skills/moai-domain-html-report/references/templates/*.mustache
stdout: financial:1 / explainer:1 / plan:1 / incident:1 / pr:1 / status:1    exit: 0
        (EV-11 GREEN — 파일당 1. artifact:end도 동일 관측, 파일당 1)

cmd:   grep -c 'prefers-color-scheme: dark' .../templates/*.mustache
stdout: 6파일 전부 :1    exit: 0   (EV-10 GREEN)
cmd:   grep -c 'data-theme="dark"' .../templates/*.mustache
stdout: 6파일 전부 :2    exit: 0
cmd:   grep -cF ':root:not([data-theme="light"])' .../templates/*.mustache
stdout: 6파일 전부 :2    exit: 0
        비고 — 이 마커는 grep 기본(BRE)에서 [..]를 문자 클래스로 읽어 0이 나온다
        (기계적 유의점: 고정 문자열 검색 -F 필요). -F로 관측 시 ≥1.

cmd:   sed -n '/artifact:begin/,/artifact:end/p' <파일> | grep -c 'fonts.googleapis.com'
stdout: 6파일 전부 ≥2 (financial 실측 2)    exit: 0   (AC-008 블록 내 ≥1)
cmd:   sed -n '/artifact:begin/,/artifact:end/p' <파일> | grep -c 'cdn.jsdelivr.net'
stdout: 6파일 전부 :0 (financial 실측 0)    exit: 1   (블록 내 0 — 요구 충족)
cmd:   head -1 <각 mustache>                 stdout: 6파일 전부 `<!doctype html>`
        (REQ-005 — Jev Q2 확정 해석: 자체 골격 유지, 래퍼 상속 제거)

cmd:   grep -ci 'pretendard' .../templates/*.mustache   (EV-12 보존 핀치 재측정)
stdout: financial:4 / explainer:0 / incident:4 / plan:3 / pr:4 / status:4    exit: 0
        (블록 삽입 후에도 전 파일 핀치와 동일 — 블록 내 pretendard 문자 0)
```

폰트 결정(계획 §F M3 ③ 이행): artifact 링크 세트는 세리프 모드(plan·explainer)에
Noto Serif KR을 포함하고 나머지 4모드는 Noto Sans KR + JetBrains Mono.
`references/fonts.md`에 artifact 매핑 절 + 근거 추가.

### M4 — skill-routing.md 소유 분리 문구 (2026-10-02, 커밋 c65172d82)

GREEN 재측정(AC-RAD-004 — 커밋 c65172d82 트리):

```text
cmd:   grep -c 'publication contract' .claude/rules/moai/workflow/skill-routing.md
stdout: 2        exit: 0   (EV-04 GREEN — ≥1)
cmd:   grep -cE '`html\+md`.{0,4}`md`.{0,4}`artifact`' .claude/rules/moai/workflow/skill-routing.md
stdout: 1        exit: 0   (EV-05 GREEN — 3값 결합 라인)
```

문구: 안티패턴 단락이 소유 분리 서술로 대체 — 콘텐츠·렌더링 = moai-domain-html-report,
아티팩트 게시 계약(publication contract) = artifact-design. format=artifact 전달 단계의
artifact-design 참조는 정상 라우팅으로 명시.

### M5 — 미러 동기화 + 중립성 + 카탈로그 해시 + 빌드·테스트 (2026-10-02)

미러 동기(REQ-012 — 라이브 10파일 cp 후 재생성 입증):

```text
cmd:   diff -rq .claude/skills/moai-domain-html-report internal/template/templates/.claude/skills/moai-domain-html-report
stdout: (없음)   exit: 0   (AC-RAD-011 GREEN — SKILL-MIRROR-PARITY-OK)
cmd:   diff -q .claude/rules/moai/workflow/skill-routing.md internal/template/templates/.claude/rules/moai/workflow/skill-routing.md
stdout: (없음)   exit: 0   (ROUTING-MIRROR-PARITY-OK)
```

중립성(AC-RAD-012 유지):

```text
cmd:   grep -rn 't1427\|SPEC-REPORT-ARTIFACT' internal/template/templates/
stdout: (없음)   exit: 1   (0히트 — 카드 내력 미유입)
```

출력 관례(AC-RAD-010 재검증):

```text
cmd:   go test ./internal/template/ -run '^TestHtmlReportOutputPathParity$' -v
stdout: === RUN   TestHtmlReportOutputPathParity
        --- PASS: TestHtmlReportOutputPathParity (0.00s)
        PASS
        ok      github.com/modu-ai/moai-adk/internal/template   0.294s
exit:  0
```

카탈로그 해시 갱신 — 신규 파일(artifact-contract.md) 포함 전체 트리 해시 무효화로
`TestCatalogHashCoversSkillSubfiles`·`TestManifestHashFormat` 2건 RED를 관측했고
생성기 단일 엔트리 갱신(`go run internal/template/scripts/gen-catalog-hashes.go
--entry moai-domain-html-report` → a7f209ba…)으로 수리했다. 수리 후:

```text
cmd:   go test ./internal/template/
stdout: ok      github.com/modu-ai/moai-adk/internal/template   162.070s
exit:  0
cmd:   make build
stdout: go build -ldflags ... -o bin/moai ./cmd/moai
exit:  0   (템플릿 재임베드 — Template-First 빌드 사이클)
```

최종 범위 스위트(-count=1 강제 재실행):

```text
cmd:   go test ./internal/web/ ./internal/settings/ ./internal/cli/wizard/ -count=1
stdout: ok  github.com/modu-ai/moai-adk/internal/web        112.967s
        ok  github.com/modu-ai/moai-adk/internal/settings   2.539s
        ok  github.com/modu-ai/moai-adk/internal/cli/wizard 12.552s
exit:  0
cmd:   go build ./...
exit:  0
```

**사전 존재 적색 2건 보고(본 카드 범위 밖 — 수리하지 않고 보고만)** — plan M5의
`go test ./internal/cli/...` 전체 실행에서 `internal/cli` 루트 패키지가 2개 테스트로
적색:

1. `TestSyncGateLanguageDetectionMatchesScript/kotlin_source` — 결정론적 적색(격리
   재실행에서도 동일): `app/src/Main.kt` + `build.gradle.kts` 트리에서 스크립트
   detect_languages가 `[kotlin java]`, Go detectSyncGateLanguages가 `[kotlin]` —
   Go 포트와 참조 스크립트의 감지 규칙 불일치. 관련 코드 최종 변경은 t1099(b82562b92
   이후)로 본 카드 base(2e700a15e) 이전 — 본 카드 변경 집합 19파일 중
   언어감지 코드 0(merge-base 범위 실측). 언어감지 정렬 수리는 별도 카드 대상.
2. `TestStopChainMemberCostWithinBudget` — 부하 민감 재현: 두 번의 격리 재실행에서
   초과 멤버 예산 대비 값이 매번 다르게 요동(member 2: 1.073s→1.558s, member 6:
   1.303s→1.065s, 예산 1s). 체인 데드라인(7.2s)은 두 번 모두 통과(최대 2.7s).
   코드 회귀라면 특정 멤버가 결정론적으로 적색이어야 하므로 기계 부하 요동으로 판정.
   측정 대상 훅 체인(sync-phase quality gate·codex-review-gate)은 본 카드 변경과
   무관. t1099가 캘리브레이션한 예산의 기계 상태 의존 — 리더 판정 대상.

gofmt: 본 SPEC 변경 파일 전부 clean. `gofmt -l`이 가리킨 2파일
(internal/cli/mcp_claude.go, internal/web/codex_panel_test.go)은 본 SPEC 변경 집합 밖의
기존 베이스라인 편차 — 미접촉(범위 규율).

## §E.3 Run-phase Audit-Ready Signal

run_status: audit-ready
run_complete_at: 2026-10-02
ac_matrix: AC-RAD-001~009 release-blocking 전부 GREEN 전환 완료(§E.2의 동일 명령
GREEN 재실행 기록) + AC-RAD-010~014 regression-guard 전부 유지 재입증.
commits: 6d8968e5a (M1) · 7cf9130f1 (M2) · 612ff3960 (M3) · c65172d82 (M4) ·
7bb5f7d3b (M5 미러+카탈로그) · 본 절 갱신 커밋(직후 착지)
final_tree_verification: 7bb5f7d3b에서 AC-001(2)/002(8)/003(1)/005(1·1·3)/007(5809B)/
009(2)/011(S:0·R:0)/012(0행 exit 1)/013(기본 html+md·미러 diff 0)/014(25900B) 전수
재실행 확인 — AC-004/006/008/010은 §E.2 M4/M3/M5 실측 기록 참조.
lint: golangci-lint v2.1.6(CI 판) — ./internal/settings/... ./internal/cli/wizard/...
./internal/core/... → 0 issues (exit 0).
evidence_ledger: acceptance.md §E(RED-now, 트리 5457f5832) ↔ 본 §E.2(GREEN 재실행,
커밋별 트리) — 전환 귀속은 acceptance.md §D 표 기준.
deviations: M2 SKILL.md 초산 증분이 예산 초과 → 산문 축약으로 수리(§E.2 M2 비고).
M5 카탈로그 해시 갱신이 계획 문언에 없던 필수 수리로 추가(전체 트리 해시 구조의
필연 결과 — 신규 파일 추가 시마다 갱신 대상).
open_for_sync: CHANGELOG 항목(report.format=artifact + 스킬 전달 개편),
미러 패리티 재입증(diff 0), 중립성 재입증 — §D.4 종결 게이트 기준.
pre_existing_reds: internal/cli 2건(§E.2 M5 말미 — kotlin 감지 규칙 불일치 결정론적
적색·t1099 기원 / stop-timing 예산 요동 부하 민감) — 본 카드 범위 밖, 리더 판정 대상.

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
