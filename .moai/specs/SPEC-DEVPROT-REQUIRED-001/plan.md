# plan.md — SPEC-DEVPROT-REQUIRED-001 (Tier M, design-only)

> Run-phase 범위: **워크플로 YAML 편집 + 창 절차 문서 갱신 + 운영자 런북 작성**. 보호 설정 적용(gh api PUT)은 범위 밖(운영자 게이트). Go 코드 변경 없음.

---

## §A Context

| 항목 | 값 |
|---|---|
| card | t324 (design-only, 리드 배차 승인) |
| worktree | `.claude/worktrees/t324` (branch `WT-devprot-required`, base `origin/develop` @ `fa8ff89ba`) |
| SPEC 산출물 | `.moai/specs/SPEC-DEVPROT-REQUIRED-001/{spec,plan,acceptance,progress}.md` + research.md |
| 현재 develop 보호 | 없음(404 "Branch not protected", 2026-09-02 live GET) |
| 핵심 산출물 경로 | `.moai/docs/develop-protection-runbook.md`(신규), `.github/workflows/ci.yml`, (조건부) `.github/workflows/codeql.yml`, `.claude/rules/local/gitflow-lane-protocol.md`, `.moai/docs/git-local-workflow-doctrine.md` |

**연구 정정(반드시 반영)**: research.md §2.2의 "codeql marker 갭으로 `Analyze (Go) (go)`가 docs-only 직접 push에서 미보고" 전제는 반증됨 — codeql.yml `analyze` 잡이 `github.event_name == 'push'` 분기로 push에서 무조건 실행하며, docs-only 머지 `fa8ff89ba`에서 `Analyze (Go) (go)`=success를 실측했다(spec.md §1.3). run-phase 산출물은 정정된 사실에 기반해 작성한다.

### §A.1 결정 기록 (DECIDED — 오케스트레이터 해소, 2026-09-02, plan-audit iteration 1)

종전 해소 대기 마커 3건의 해소 값(오케스트레이터 확정). **잔여 마커 0** — 이 표가 정준 결정 기록이다.

| # | 결정 | 값 | 근거 |
|---|---|---|---|
| D-1 | phase-1 필수 세트 | `Test (ubuntu-latest)`, `Lint`, `Build (linux/amd64)`, **`Analyze (Go) (go)`** (4컨텍스트) | main 필수 세트 패리티(live API: 5컨텍스트에 Analyze 포함, 구조적 부재하는 `Release PR Multi-OS Gate`만 제외); codeql analyze 잡이 모든 push 이벤트에서 실행(codeql.yml:59-63 조건) → develop push에서 무조건 보고(spec.md §1.3 실측) |
| D-2 | verify ref | **카드별 `verify/<card-id>`** (단일 공유 verify 브랜치 아님) | 통합 락 변위(`--force` 획득 경로 존재)에도 강인; Actions 이력에서 카드별 실행 귀속 명확; 비정상 인터리빙에서도 cancel-in-progress 간섭 없음. **청소**: 해당 카드의 develop push가 착지하면 verify ref를 삭제한다 |
| D-3 | B1 CI 대기 | **자동 `gh run watch --exit-status <run-id>`** — timeout 상한 부여, run id는 verify-ref push에서 해석 | 창은 에이전트 레인이 수행; `gh run watch`는 이 리포의 기존 CI-watch 관용구(scripts/ci-watch). 모호한 watch 실패 시 폴백: commit status API 폴링 |

## §B Known Issues

- **B-T1 (delivery.md 미러 함정)**: `WT-*` 통합 절차의 정본은 배포 스킬 `.claude/skills/moai/workflows/sync/delivery.md` Step 3.2다(template-managed). 본 SPEC의 창 절차 갱신은 **local-only** 파일(`.claude/rules/local/gitflow-lane-protocol.md` §4 확장 + 런북)로 한정하는 것이 기본값이다. delivery.md를 직접 편집하려면 Template-First 미러 + 중립성 검토가 별도로 필요하다 — 기본값 밖으로 벗어나지 않는다.
- **B-T2 (워크플로 편집 안전성)**: push 트리거 `branches` 배열에 verify 패턴을 **추가만** 한다(기존 `main`·`develop` 보존). 편집 후 `actionlint`로 회귀 확인(변경 전 baseline exit 0, 이번 실행 측정).
- **B-T3 (7일 사전조건)**: 후보 컨텍스트는 현재 develop에서 초록으로 돌고 있다(실측 `fa8ff89ba`). 그러나 보호 적용 시점과 run-phase 사이에 간격이 생기면 조건이 비어질 수 있다 — 런북에 적용 직전 확인 절차(읽기 전용 check-runs 조회)를 넣는다(REQ-DPR-012).
- **B-T4 (동일 SHA 2회 실행)**: ci.yml concurrency group은 ref-scoped(`${{ github.workflow }}-${{ github.ref }}`, 실측)라 verify 실행과 develop push 실행이 상호 취소되지 않는다. 비용 문서화로 충분하다 — 그룹 키 변경은 하지 않는다(공유 상태 변경, 범위 밖). 카드별 verify ref(D-2)는 순차 창 사이 cancel-in-progress 충돌도 제거하며, 착지 후 ref 삭제 청소를 절차에 포함한다.
- **B-T5 (스코프 규율)**: 런타임 상태 파일(`.moai/state/`, `.moai/harness/`) 미수정. 타 SPEC 디렉터리 미수정.
- **B-T6 (언어)**: 워크플로 YAML 주석은 영어(`code_comments: en`), 런북·절차 문서는 한국어(`documentation: ko`).
- **B-T7 (해소 대기 마커는 이 문서에만)** — spec.md·acceptance.md에는 마커를 두지 않는다. (2026-09-02 해소: 3건 전부 §A.1 DECIDED로 전환, 잔여 마커 0.)

## §C Pre-flight (run-phase 착수 전 읽기 전용 확인)

```bash
# 1. 현재 보호 상태 (적용 전 404가 정상)
gh api repos/modu-ai/moai-adk/branches/develop/protection
# 2. 트리거 baseline
yq -o=json '.on.push.branches' .github/workflows/ci.yml
yq -o=json '.on.push.branches' .github/workflows/codeql.yml
# 3. actionlint baseline (편집 전 exit 0이어야 함 — RED-now: 이번 실행에서 exit 0 측정됨)
actionlint .github/workflows/ci.yml .github/workflows/codeql.yml
# 4. develop HEAD 컨텍스트 보고 확인 (7일 사전조건의 근거)
gh api "repos/modu-ai/moai-adk/commits/$(gh api repos/modu-ai/moai-adk/commits/develop --jq .sha)/check-runs" --jq '.check_runs[].name'
```

## §D Constraints (DO NOT VIOLATE)

- **어떤 `gh api -X PUT/PATCH/DELETE`도 실행 금지** — 읽기(GET)만 허용.
- `.github/workflows/*.yml` 편집은 트리거 배열 추가로 한정; 잡 정의·skip-marker 구조를 건드리지 않는다.
- 보호 적용 절차는 런북 문서로만 존재한다(REQ-DPR-010).
- `gitflow-lane-protocol.md`·런북 외의 배포 스킬/템플릿 파일 미수정(B-T1).
- Conventional Commits + 카드 id(t324)를 커밋 메시지에 포함.

## §E Self-Verification (run-phase 완료 시)

acceptance.md의 AC 매트릭스(14개)를 명령과 함께 실행하고 PASS/FAIL를 원본 출력으로 보고한다. 쓰기 명령(gh api PUT)은 실행하지 않는다 — 해당 검증 차원은 "문서화된 명령의 형식·인자 검증 + GET dry-run"으로 대체된다.

## §F Milestones (결정 가역성 내림차순)

- **M1 — 운영자 런북 작성** (`.moai/docs/develop-protection-runbook.md`) — 결정 내용의 최종 운반체. §A.1의 DECIDED 값을 그대로 기록한다: 필수 세트 4컨텍스트(D-1), verify ref = 카드별 `verify/<card-id>` + 착지 후 ref 삭제 청소(D-2), B1 대기 = `gh run watch --exit-status` 자동(timeout 상한, 모호 시 commit status API 폴백)(D-3), B2→B1 단계화와 enforce_admins 트레이드오프, apply/rollback 명령, GH006 예상 거부·회복, 7일 사전조건 확인 절차, phase-2 후보 분류(Race Test — 동반 변경 사유 포함), 롤아웃 순서(위반=창 고착). — 아래 모든 마일스톤의 판단 기준이 여기에 실린다.
- **M2 — ci.yml + codeql.yml 트리거 확장** (양쪽 모두 — D-1로 Analyze가 필수 세트에 포함되어 codeql도 무조건 대상) — push `branches`에 `verify/*` 항목 추가(D-2 정준형). `actionlint` 회귀 + `yq` 단언.
- **M3 — 창 절차 갱신** — `gitflow-lane-protocol.md` §4에 B1 사전검증 절차(카드별 `verify/<card-id>` push → `gh run watch --exit-status` 자동 대기·timeout 상한·모호한 실패 시 commit status API 폴백(D-3) → 동일 SHA develop push → **착지 후 verify ref 삭제 청소**(D-2))와 락 홀드 연장(REQ-DPR-008) 추가.
- **M4 — doctrine 갱신** — `git-local-workflow-doctrine.md` §23.2 낡은 스냅샷 고지 + develop 보호 설계 기록(REQ-DPR-013).
- **M5 — AC 스윕** — acceptance.md 매트릭스 전항 재실행(two-cell flip), verify 패턴 단일 토큰 일관성 확인.

## §G Anti-Patterns

- **AP-1**: run-phase에서 보호를 적용하는 것(gh api PUT) — 운영자 게이트 침범.
- **AP-2**: ③(보호 적용)을 ①②보다 먼저 안내하거나, B1 부재 시 `enforce_admins: true`를 권장하는 런북 — 창 고착(REQ-DPR-011).
- **AP-3**: B2 단계에서 "필수 검사가 창을 게이트한다"고 기술 — 정직한 프레이밍 위반(REQ-DPR-006).
- **AP-4**: research.md §2.2의 반증된 codeql 전제를 런북·문서에 그대로 옮기는 것 — §1.3 정정본을 인용한다.
- **AP-5**: `Release PR Multi-OS Gate`·test-install 컨텍스트를 필수 목록에 넣는 것(REQ-DPR-002).
- **AP-6**: verify 패턴을 파일마다 다르게 쓰는 것 — 두 정준형(워크플로 `verify/*`, 문서 `verify/<card-id>`) 준수(AC-DPR-009).

## §H Cross-References

- research.md — 측정 원본 (§2.2 전제는 반증됨: spec.md §1.3)
- `.claude/rules/local/gitflow-lane-protocol.md` — 창 절차의 local 정본(M3 대상)
- `.moai/docs/git-workflow-doctrine.md` §18.7 — main 보호 선례 / `.moai/docs/git-local-workflow-doctrine.md` §23.2 — 낡은 스냅샷(M4 대상)
- GitHub docs(보호된 브랜치·필수 검사 트러블슈팅) — research.md §3의 검증된 명제와 URL
