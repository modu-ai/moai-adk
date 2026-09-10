# plan — SPEC-CODEMAPS-ACCURACY-001

## §A Context

- 카드 t304 (Class B — 결함, 원인은 레인 조사로 확립됨). worktree `.claude/worktrees/t304`, branch `WT-codemaps-accuracy`, base origin/develop `65196a5a7`.
- 문제·분류·근거는 spec.md §1 (§1.1 분류표 8행 완결, §1.2 ListActive 3개 지점, §1.3 생성기 입력 재발 경로, §1.4 F1).
- 핵심 원칙: 신선도 게이트 초록 ≠ 정확성 증명 (t432 REQ-CMR-008 계승, REQ-CMA-010).
- 병합 의존: t432 (WT-codemaps-refresh @ `7e1c4d94f`)가 같은 6개 codemaps 파일을 재생성(신선도 60→0, 유령 본문 편집 0 — known-6 승계). 본 카드 편집은 t432 병합 후 착지.

## §B Known Issues

| 이슈 | 내용 | 대응 |
|------|------|------|
| 병합 순서 | t432 미병합 상태에서 codemaps 편집 시 충돌/덮어쓰기 | M0 게이트: 흡수 + 재측정 후 편집 (REQ-CMA-008) |
| 좌표 드리프트 | spec.md §1의 행번호(develop 기준)는 t432 병합 후 이동 | M0에서 §1.1·§1.2 좌표 전부 재측정해 progress.md에 기록 |
| bodp 이중 표현 | develop modules.md 양성 절 vs t432 dependencies.md 부정 각주 | 병합 후 실제 트리 상태 기준으로 D3 판정 적용 (REQ-CMA-004) |
| D2 위음성 | blockquote 내 양성 주장은 검사 우회 | 스킬 규약으로 blockquote를 부정 인용 전용 표기로 고정 (REQ-CMA-007); AC의 뮤턴트로 면제 정상작동만 검증 |
| 검사 대상 파일 수 | codemaps 6개 파일 중 data-flow.md에 mermaid 블록·코드펜스 내 경로 포함 | 추출기는 코드펜스·mermaid 블록도 주석이 아닌 이상 동일 규칙으로 스캔(인용은 블록 안에도 존재 — `ListActive` mermaid 노드가 실례), blockquote 행만 면제 |

## §C Pre-flight (run-phase 착수 전 검증)

1. `git -C .claude/worktrees/t304 fetch origin && git rev-parse origin/develop` — t432 병합 커밋 포함 확인 (WT-codemaps-refresh `7e1c4d94f`가 origin/develop의 조상인지 `git merge-base --is-ancestor 7e1c4d94f origin/develop`). **부성 분기**: 조상이 아니면(병합 전 후퇴·롤백) 레인은 blocker를 리드에게 보고하고 리드가 지정한 병합 순서를 기다린다 — M2 문서 편집으로 진행하지 않는다 (M1 Go 작업은 §D 편집 범위 밖이므로 계속 가능).
2. 흡수: 카드 워크트리에서 `git merge origin/develop`.
3. 재측정 (REQ-CMA-008): §1.1 부재 8개 재판정 + §1.2 ListActive 3개 지점 재좌표 — 명령과 출력을 progress.md §E.2에 기록. develop 기준 좌표가 유효하지 않으면 spec.md §1 좌표를 정정 커밋으로 갱신.
4. LSP baseline 캡처 (spec-workflow plan 단계 계약).

## §D Constraints

- [HARD] 편집 순서: M0 흡수·재측정 없이는 `.moai/project/codemaps/**`만 편집 금지 대상이다 — t432 병합은 문서 전용 병합이므로 `internal/`(M1 Go 축)은 흡수 대기에 멈추지 않는다.
- [HARD] Template-First: codemaps.md 스킬 편집은 `internal/template/templates/.../codemaps.md` 먼저 → 로컬 미러 동일 변경 → `make build`.
- [HARD] exit-code 계약 불변: `moai graph check` 0/1/2. citations 행은 `Failed()`에 참여하되 기존 3레이어 행의 의미·출력 형식은 유지.
- 검증은 카드 범위로 한정: `go test ./internal/graph/... ./internal/cli/...` (전체 스위트는 CI — CLAUDE.local.md §4/§6).
- 커밋은 레인 세션이 수행 (본 plan-phase는 커밋하지 않는다).

## §E Self-Verification (이 문서 작성 시점에 실행한 것)

- SPEC-ID 정규식 사전검사: `SPEC-CODEMAPS-ACCURACY-001` → `PASS` (Bash 실행 출력 인용).
- ID 충돌: `moai spec` 카탈로그(spec_progress, project_root=본 워크트리)에서 `SPEC-CODEMAPS-ACCURACY-*` 0건 — 인접 ID `SPEC-DWF-CODEMAPS-PILOT-001`·`SPEC-V3R6-DOCS-CODEMAPS-V3-001`는 다른 ID.
- R1: 스킬 양본(로컬 + 템플릿 정본) Phase 2/3/4 직독 — 재발 경로 특정 (spec.md §1.3).
- R2: `internal/cli/graph_check.go` 전문 + `internal/graph/check.go` 28–167행 직독 — LayerReport·Thresholds·CheckFreshness·metric 토큰 구조 확인 (spec.md §3 D1의 부착점 근거).
- R3: data-flow.md 185–224·345–363행 + `internal/session/registry.go` 150–270행 직독 — 실제 API 서명 확정 (spec.md §1.2 코드블록).
- modules.md 경고 노트 5개·factory 절·bodp 절 sed 직독으로 형식(blockquote) 확인 (D2·D3 근거).
- t432 보고서에서 정리 규칙(19행)·ListActive 판정(§3.1 #21)·F1(194행 "26항목" vs 251행 "27항목") 직독.

## §F Milestones

> 의사결정 가역성 내림차순 — 인터페이스를 정의하는 Go 축(M1)을 먼저 인간 검토한다. 우선순위 라벨은 High/Medium/Low (시간 추정 금지).

- **M0 (High) — 흡수 게이트 + 유령 재인벤토리** [REQ-CMA-008]
  - `origin/develop` 흡수(t432 병합 반영) → §1.1 8행·§1.2 3지점 재측정 → progress.md §E.2에 명령+출력 기록 → (필요시) spec.md §1 좌표 정정.
  - 완료 판정: 재측정 표가 progress.md에 존재하고 분류표가 post-absorb 트리에서 유효.
- **M1 (High) — cited-path-existence 검사 축 (Go)** [REQ-CMA-002, D1/D2]
  - `internal/graph`: citations 검사 신규 파일(예: `check_citations.go`) — `.moai/project/codemaps/*.md`에서 경로 토큰 추출, t432 정리 규칙(후행 슬래시 제거·`.go` 복원·`cmd/moai/main`→`cmd/moai/main.go`) 적용, blockquote 행 면제, 양성 부재 목록 산출. `CheckFreshness`의 Layers에 행 추가(`layer=citations`, `metric=positive-cited-path-absence`, threshold 0). metric 토큰은 `check.go` 상수 블록에 추가.
  - `internal/cli/graph_check.go`: 신규 행이 기존 렌더러·`OffendingLayers()`·`Failed()`를 경유해 exit 1로 이어짐을 확인 (부가 와이어링 최소화 — LayerReport 소비 경로 재사용).
  - 테스트: (a) 고정 픽스처에서 양성 유령 1개 → red; (b) blockquote 부정 인용만 있는 문서 → green; (c) 정리 규칙 3종 각각 단위 테스트; (d) 실측: 수정 전 develop 트리에서 red → M2 완료 후 green (뮤턴트 폐쇄: 수정된 문서에 가짜 양성 유령 1개 주입 시 재차 red).
  - 완료 판정: `go test ./internal/graph/... ./internal/cli/...` 초록 + 위 4종 테스트가 존재하고 통과.
- **M2 (High) — codemaps 사실 수정** [REQ-CMA-001/003/004/005/006, D3/D4]
  - modules.md: `### internal/factory` → `### internal/kanban` 절 재작성(Kanban 모드 서술, 진입점 `internal/cli/factory.go`·`internal/cli/launcher_blockcap_infinite.go`는 실존하므로 유지); `### internal/bodp` 양성 절 → blockquote 부정 인용 노트로 전환(t432 dependencies.md 각주와 형식 정렬); P1–P5 노트 불변.
  - data-flow.md: ListActive 3개 지점(mermaid 노드·흐름 단계·인터페이스 블록)을 실제 API로 수정 — 인터페이스 블록은 리시버 메서드 4종 + `Entry` 반환, 패키지 함수 `QueryActiveWork` 서술.
  - 완료 판정: `moai graph check` citations 행 green; §1.1 재판정 표의 조치 열 전부 "완료".
- **M3 (Medium) — 스킬 재발 방지 (Template-First)** [REQ-CMA-007, D2 완화]
  - 템플릿 정본 `internal/template/templates/.claude/skills/moai/workflows/codemaps.md` 편집: Phase 2 — 기존 codemaps 콘텐츠를 패키지 존재의 권위로 사용 금지(병합은 서술 품질용으로 한정); Phase 4 — 실행 가능한 검증 명령(`moai graph check` citations 행)과 부정 인용 blockquote 규약 명시. 같은 변경을 로컬 미러에 적용. `make build`.
  - 완료 판정: 양본 동일 변경 + `make build` 성공 + mirror-parity/중립성 기존 가드 통과.
- **M4 (Low) — F1 정정 기록 + 증거 반출** [REQ-CMA-009]
  - `.moai/reports/t304/`에 F1 정정(t432 보고서 §3.1 "26항목"→실제 27행 표) 기록, 리드 보고에 포함. t432 트리 무결성 확인(`git -C .claude/worktrees/t432 status --short`에서 본 카드 기인 변경 0).
  - 완료 판정: 증거 파일 존재 + t432 트리에 본 카드 쓰기 0 관측.

## §G Anti-Patterns

- **재스탬프로 완료 선언**: 신선도 게이트 green을 정확성 증명으로 제시 (REQ-CMA-010 위반).
- **흡수 전 편집**: t432 병합 전 codemaps 파일 수정 — 병합 시 소실.
- **로컬 미러 먼저**: Template-First 위반 — `moai update` 시 유실 (CLAUDE.local.md §2.3).
- **알려진 부정 인용 "정리"**: P1–P5 노트를 삭제하거나 양성 문구로 되돌리는 구동적 정리 (REQ-CMA-006 위반).
- **blockquote 우회**: 검사를 피하려 양성 서술을 blockquote로 옮기기 — D2 위음성을 악용하는 형태.
- **검사의 휴리스틱 과잉**: 존재 검사에 문맥 이해·의도 판별을 넣는 것 — blockquote 면제 + 경로 존재만. 나머지는 LLM(스킬) 소관.

## §H Cross-References

- t432 증거(읽기 전용): `.claude/worktrees/t432/.moai/reports/t432/codemaps-accuracy-verification.md` — 정리 규칙·전수검사 방법·F1.
- t432 SPEC: REQ-CMR-004(ListActive 기록만)/REQ-CMR-008(신선도≠정확성) — 본 SPEC REQ-CMA-005/010이 승계.
- `internal/graph/check.go` — LayerReport·Thresholds·metric 토큰(D1 부착점).
- `internal/session/registry.go:169-266` — 실제 Registry API(REQ-CMA-005의 정본).
- CLAUDE.local.md §2(Template-First)·§4/§6(검증 범위)·spec-workflow.md(LSP 게이트).
