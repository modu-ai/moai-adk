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

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
