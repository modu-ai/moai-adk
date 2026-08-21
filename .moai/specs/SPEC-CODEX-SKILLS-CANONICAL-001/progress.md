# progress — SPEC-CODEX-SKILLS-CANONICAL-001

## §E.1 Plan-phase Audit-Ready Signal

- Tier: **M** (spec.md + plan.md + acceptance.md). 근거: 예상 변경 파일 8-12개(`internal/template/` 신규 1 + `deployer.go` + `internal/cli/update/deploy/deploy.go` + `internal/defs/dirs.go` + 테스트 4-6 + 템플릿 미러), 예상 LOC 400-700 → Tier M 대역(300-1000 LOC / 5-15 files).
- 요구사항 12개 / 판정 기준 15개 — Tier M 상한 16/16 이내.
- SPEC ID 정규식 검사: `SPEC-CODEX-SKILLS-CANONICAL-001` → `PASS` (Bash `[[ =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]]`).
- 중복 검사: `.moai/specs/` 에 동일 ID 없음 (`SPEC-CODEX-PHASE2-001` 만 존재).
- 착수 시점 실측(작성 시점): 템플릿 스킬 34개 / core 21 / non-core 13 / 템플릿 트리 심볼릭 링크 0개.
- iter-5 (v0.5.0): iter-4 가 들여온 결함 3건(D2/D3/D3-b) + 인용 오류(D1/D7) + optional 4건 대응. 설계 무변경, 요구사항 계층 무변경. **REQ 16 / AC 16 불변 — 은퇴 없음.** AC-CSC-008 fixture 4형태 → **5형태**(`moai-linkprobe` 링크 추종 탐침 추가), 단언 4 → 6, 판정 syscall 을 `os.Lstat` 으로 [HARD] 고정. AC-CSC-012 2번 팔 측정 폭을 서브트리 전체 → **이름 단위**로 축소. §A.9b 신설(출력 표면 부재 실측 승격) + 출력 seam 인용 4곳 `§B.D6` → `§B.D3` + `§A.9b`. plan M3 본문·닫힘 조건 정정.
- iter-5 실측 기록: (a) dangling 링크 메커니즘 독립 재현 → `Stat err: … no such file or directory  IsNotExist: true` / `Lstat err: <nil>` / `glob: [a/.agents/skills/moai-x]`. (b) clean→deploy 순서 — `update_template_sync.go:297`(clean) < `:323`(deploy). (c) `ManagedCleanTargets` 는 목록 전체를 순회하며 4번째 항목이 `.claude/skills/moai*` → fixture 의 `moai-live` 정본은 링크와 무관하게 제거됨(D2 의 근거).
- iter-5 감사 쟁점: iter-3 감사의 지적 중 반박한 것 없음 — D1~D7 전건이 재현·확인됐고 전부 반영했다. 리드가 제시한 D2 대안 (b) 만 근거를 들어 채택하지 않았다(배포가 손상을 치유해 관측 불가; 사유는 AC-CSC-008 본문).
- iter-4 (v0.4.0): 독립 감사 2건(0.78 / 0.7625) 대응, 리드 상한 예외 승인 하에 수행. **REQ 16 / AC 16 불변 — 은퇴시킨 항목 없음**(모든 수정이 기존 번호 안의 절 추가·수정). 핵심은 §A.10 dangling 결함: REQ-CSC-008 에 `os.Lstat` 판정 + dangling 제거(본체) + 슬라이스 순서(이중 방어), AC-CSC-008 fixture 를 미러 4형태로 확장. 그 밖에 REQ-CSC-001 예외 절, REQ-CSC-010 백업 금지 한정(§A.11), REQ-CSC-005 반환 결과 seam, REQ-CSC-016 범위 축소, §A.9 접두 철자 정정(`moai-` → `moai`), REQ-CSC-007 문구 축소. optional 전건 반영. AC-006·015 를 SHOULD → MUST 로 승격(MUST 13 → 15).
- iter-4 실측 기록: (a) dangling 링크 — 직접 실행 → `Stat err: … no such file or directory  IsNotExist: true` / `Lstat err: <nil>` / `glob: [a/.agents/skills/moai-x]`. (b) 실행 순서 — `update_template_sync.go:297`(clean) < `:323`(deploy). (c) 출력 seam 부재 — `internal/template` 비-테스트 파일에 `io.Writer` 매치 0. (d) 접두 — `grep -cv '^moai-'` → **1**(`moai`), `grep -cv '^moai'` → 0.
- iter-4 감사 쟁점 확인: audit#1 의 N10(M3 모순)·N11(M5 미반영)·N12(백업 소유 마일스톤 부재)는 **디스크 실물 확인 결과 iter-3 에서 이미 닫혀 있었다** — `plan.md` 의 M3 제목은 "미러를 기록·백업 대상에서 제외 (Priority High)", M5 는 "`.gitignore` 에 `.agents/` 등록 (Priority **High**)", AC-CSC-012 는 M3 닫힘 조건에 등록돼 있다. audit#2 가 같은 결론을 냈다. audit#1 이 개정 전 스냅샷을 읽은 것으로 보인다.
- iter-3 (v0.3.0): plan-audit iteration 1 (FAIL 0.775) blocking 8건 대응. 요구사항 12 → **16**, 판정 15 → **16**(둘 다 Tier M 예산 상한 도달). §A 에 실측 4건 추가(§A.6 `manifest.Track` EISDIR, §A.7 pre-clean 백업 비대칭, §A.8 `.gitignore` 부재, §A.9 접두 우연 일치). REQ-CSC-010 방향 반전(기록한다 → 기록·백업하지 않는다). AC-CSC-001/010/013 재작성, AC-CSC-007 경로 구분자 중립화, AC-CSC-011 3-상태 확장, AC-CSC-016 신설. plan M3 방향 반전 · M5 Priority Low → High · M6 신설 · AP 5건 추가. D12 기각(사유 spec §G). 실측 재현 명령과 출력은 아래 iter-3 실측 기록 참조.
- iter-3 실측 기록: (a) `manifest.HashFile` 디렉터리 링크 — 별도 Go 프로그램 실행 → `open err: <nil>` / `copy err: read lnk: is a directory`. (b) 템플릿 `.gitignore` `.agents/` 항목 부재 — `grep` 결과 `.claude/` 계열만. (c) `optional-pack:*` tier — `grep -c 'tier: optional-pack'` → 13, `harness_generated.skills` → `[]`. (d) 비-`moai` 접두 템플릿 스킬 — `find … -exec basename {} \; | grep -cv '^moai'` → 0.
- iter-2 (v0.2.0): 리드 추가 제약 반영 — 청소 글롭 접두 `moai*` 한정을 spec §B.D5 [HARD] 로 고정, `ManagedCleanTargets` 확장이 `moai update` 동작 변경임을 §A.5 에 근거(두 차례 실측된 같은 실패 형태)와 함께 기록, AC-CSC-008 을 제거+생존 **양팔 단일 테스트**로 재작성. 요구사항 12 / 판정 15 불변, `moai spec lint` 무결점.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
