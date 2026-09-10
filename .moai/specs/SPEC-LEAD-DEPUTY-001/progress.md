# PROGRESS: SPEC-LEAD-DEPUTY-001 (card t471)

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-09-03
tier: M (확정 — 분할 판정 후. plan-auditor PASS threshold 0.80)
split_decision: 단일 SPEC, 마일스톤 A→B→C 순 (spec.md §1.4 근거)
audit_iterations: 3
audit_verdict: PASS 0.938 (threshold 0.80, iter 3 — D1·D2 수리 확인 패스, skip-eligible) — 최종 판정: .moai/reports/t471/plan-audit-iter3.md (tree e77ac4bb6). 이력: iter-1 PASS-WITH-DEBT 0.925 (.moai/reports/t471/plan-audit.md, tree 3ec569871) → D1·D2 수리 b7370371b → iter-3 PASS. 신규 optional 2건(NEW-1 exit-code 요소, NEW-2 비교 기준) — 비차단
plan_evidence: .moai/reports/t471/plan-evidence.md

## §E.2 Run-phase Evidence

run_base_pin: 615d18c1f (origin/develop 흡수 tip — fast-forward, 충돌 0. 흡수 전 HEAD bdcec58f0은 이미 origin/develop 조상, `origin/develop..HEAD = 0`)
run_evidence: .moai/reports/t471/run-evidence.md (명령 + 원문 출력 + 트리 귀속, VCI §3 5-섹션)

### AC 매트릭스

| AC | 판정 | 근거 |
|---|---|---|
| AC-LDP-001 | PASS (protocol readiness only) | 레시피·baseline·target이 교리 텍스트에 존재. 실측 감소는 §D.2가 정한 2단 검증의 2단 — 첫 채택 배치 소관 |
| AC-LDP-002 | **PASS (RED→GREEN)** | 편집 전 `:0/:0 exit=1` → 편집 후 `:1/:1 exit=0` (같은 트리). 기존 세 갈래 조건절 문단에 상호참조 **부착** — 축약 아닌 인용 |
| AC-LDP-003 | PASS (문면) | `kanban-dispatch.md` § Deputy dispatch surface [HARD] 상주 spawn 의무 + `manager-lead.md` § Resident mode |
| AC-LDP-004 | PASS (문면) | [HARD] `RECOMMEND:` 요약 경로 + 경로 명명 의무. 실적 확인은 채택 후 |
| AC-LDP-005 | PASS | delivery-shape 문단 무변경 (`grep -c` → 1, diff에서 미접촉) |
| AC-LDP-006 | PASS (문면) | [HARD] 측정자 귀속 의무 + 무귀속 = 결함. 실적 확인은 채택 후 |
| AC-LDP-007 | PASS (문면) | [HARD] 회차별 파일 + 인덱스. 실물 `.moai/reports/lead/` 미접촉 (리드 D2 판정) |
| AC-LDP-008 | PASS | neutrality grep → 0 (편집 전 기준선도 0) |
| AC-LDP-009 | **PASS (문언 편차 1건 명시)** | depth-seal `ok` · `go build` rc=0 · windows cross-build rc=0 · agentemit `ok` · `.codex` toml 재생성. **`.go` 변경 0**이나 `internal/template/catalog.yaml` 1행이 `make build`의 기계적 해시 재계산으로 변경 — 문언(`git diff --stat internal/` 공백) 대비 편차를 run-evidence.md Residual-risk 4에 기록 |
| AC-LDP-010 | PASS | `DEPUTY-RETAINED-BY-LEAD` 6항목·위임5/보유6 표·판정 소재·운영자 게이트 절 무변경 |

### 예산 — 목표 미달을 그대로 적는다

리드 승인 목표는 「always-loaded 순증 ≤ 0」(승인문의 ≤ +900 B에서 좁힌 값 — 흡수 base에서 가드가 이미 FAIL임을 실측하고 상신·승인). **실제는 +141 토큰 / +564 B로 목표 미달이다.**

```
편집 전: always-loaded surface = 77723 tokens (budget 77600, headroom -123, 17 entries)  FAIL
편집 후: always-loaded surface = 77864 tokens (budget 77600, headroom -264, 17 entries)  FAIL
```

미달 사유: 잔여 상쇄 여지가 PRESERVE 대상뿐이었다. 신설 [HARD] 3건은 -k/-f 리드 **세션**을 구속하므로 always-loaded 표면에 있어야 하고, 기존 [HARD] 2건·보유 6항목 열거는 spec.md §5 PRESERVE다. 실질을 덜어 숫자를 맞추는 형태(「가드는 항목 수를 세지 내용을 세지 않는다」)는 취하지 않았다. 적자 상환은 이 카드 밖이다 — t473(미푸시, +1,044 보유)과 t492 소관.

### 수용된 optional 2건 — 미수리 사유

- **NEW-1** (AC-LDP-002 RED-now 셀의 종료코드 토큰) · **NEW-2** (AC-LDP-007 비교 기준 1절) — 둘 다 iter-3 optional. **수용하되 `acceptance.md`를 고치지 않았다.**
- 사유 ①: 지금 고치면 plan-audit **skip-eligible의 artifact-hash가 깨져** run-gate가 재감사를 요구한다.
- 사유 ②: 같은 배치에서 **Tier 상한 예외를 연달아 쓰지 않는다**(리드 판정 — 직전 카드가 이미 1회 예외를 소비).
- 보완: NEW-1이 요구하는 종료코드는 run-evidence.md가 실측으로 운반한다 — 편집 전 `exit=1`, 편집 후 `exit=0`.

## §E.3 Run-phase Audit-Ready Signal

run_status: audit-ready
run_complete_at: 2026-09-06
milestones: M1(상주 deputy) · M2(보고 위임 교리) · M3(idle 통지) · M4(기계 검증) 전부 착지
surfaces_edited: kanban-dispatch.md (always-loaded, +564 B) · kanban-dispatch-detail.md (lazy) · manager-lead.md (에이전트 정의) + 템플릿 미러 3본 + `.codex/agents/moai/manager-lead.toml` (C2→C3 방출) + `internal/template/catalog.yaml` (해시 1행)
known_deviation: always-loaded 순증 +141 토큰 — 목표(≤ 0) 미달, 사유 명시 (§E.2)

## §E.4 Sync-phase Audit-Ready Signal

sync_status: complete
sync_complete_at: 2026-09-06
sync_commit_sha: 94a940d03
three_phase_close: "`in-progress → implemented → completed` 를 이 sync 커밋에 병합 — 별도 Mx chore 커밋 없음. spec.md frontmatter는 `status: completed` + `updated: 2026-09-06` 만 변경, 본문 무편집. plan.md/acceptance.md는 `status:` 필드 자체가 없어 `updated:` 만 갱신 (SPEC-BINLAG-KEYGUARD-001 CHANGELOG 항목 선례와 동일 서술)."
changelog_entry_position: "CHANGELOG.md [Unreleased] > ### Added — 최상단 신규 항목 (편집 전 `grep -c 'SPEC-LEAD-DEPUTY-001' CHANGELOG.md` → 0, 중복 없음 확인 후 삽입)"

### B12 self-test (3건, 커밋 전 실행)

- pre_emission_grep: `grep -c 'SPEC-LEAD-DEPUTY-001' CHANGELOG.md` → `0` (삽입 전)
- ac_count_match: `grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' .moai/specs/SPEC-LEAD-DEPUTY-001/acceptance.md | sort -u | wc -l` → `10` (AC-LDP-001~010); CHANGELOG 항목이 "10건 AC-LDP-001..010" 을 명기 — 일치. **0은 아니었다** — 실측치이며 공허 비교가 아니다.
- file_path_verification: `ls`로 확인 — `.claude/rules/moai/workflow/kanban-dispatch.md` · `kanban-dispatch-detail.md` · `.claude/agents/moai/manager-lead.md` (+ 템플릿 미러 3본) · `internal/template/templates/.codex/agents/moai/manager-lead.toml` · `internal/template/catalog.yaml` 전부 존재 확인

### 문서 표면 점검 — 무엇을 고쳤고 무엇을 고치지 않았나

- **CHANGELOG.md**: 갱신 — `[Unreleased] > ### Added` 최상단에 본 SPEC 신규 1건 삽입 (본 progress.md와 같은 커밋)
- **README.md (4로케일: ko/en/ja/zh)**: 무편집 — `grep -ni 'manager-lead\|kanban-dispatch\|deputy' README*.md` 재측정 결과 6줄 히트(실측, 최초 "0 히트"로 적었던 것은 오기 — grep 명령을 `-l`(파일명만)로 잘못 돌려 `-i` 없이 재확인하지 않은 실수였고, 이 문장을 쓰기 전에 바로잡았다). 히트는 전부 `manager-lead` 링크·요약 표 행 1개(각 로케일당 2줄)이며, 이 SPEC이 추가한 상주 deputy·`RECOMMEND:` 경로·idle 통지 어느 것도 언급하지 않는다 — 기존 문장이 이 SPEC으로 거짓이 되지 않았으므로 편집 대상이 아니다.
- **docs-site (adk.mo.ai.kr)**: 무편집 — `docs-site/content/{ko,en,ja,zh}/advanced/manager-lead.md` 4파일이 존재하고(실측, grep 전 `ls`로 확인 없이 "0 히트"라 적은 것도 위와 같은 오기), en판은 Role B(디스패치 사이클) 절에서 "parallel work ... is pushed out as background Agent() spawns"(41행)를 서술한다 — 상주 deputy 1개를 특정하지 않는 일반 서술이라 이 SPEC 이후에도 참이다. `grep -ni 'deputy\|resident\|idle'` 를 이 파일에 돌리면 0(en/ko 확인) — 새 메커니즘이 아직 문서화되지 않았다는 뜻이지, 지금 있는 문장이 거짓이 됐다는 뜻이 아니다. 과제 지침의 "거짓으로 만드는 구체적 문장" 기준을 충족하는 문장을 찾지 못해 무편집으로 남긴다 — 신설 상주 deputy 절 자체를 이 문서에 추가할지는 이 카드의 범위 밖(docs-site 보강은 별도 카드 후보).
- **`.moai/docs/*.md`**: 무편집 — 이 SPEC이 직접 편집한 대상(`kanban-dispatch.md`/`manager-lead.md` 자체가 `.claude/rules/moai/workflow/`·`.claude/agents/moai/` 아래에 있고, `.moai/docs/`는 별도 트리)이 편집 원본이므로 외부 참조 문서 갱신이 불요하다.

sync_scope: "markdown-only — `kanban-dispatch.md`(+detail) · `manager-lead.md` · 템플릿 미러 3본 · `.codex` toml 방출(C2→C3) · `catalog.yaml` 해시 1행 · SPEC 4파일 frontmatter(`status:`/`updated:`) · `CHANGELOG.md`. `internal/`·`pkg/`·`cmd/` 비접촉 — run-evidence.md AC-LDP-009 실측(`git diff --name-only internal/ pkg/ cmd/ | grep '\.go$' | wc -l` → 0)."
known_deviation_carried_forward: "always-loaded 표면 순증 +141 토큰/+564 B(§E.2 실측) — sync-phase에서 재론·상쇄 시도를 하지 않았다. 가드는 편집 전(-123)에도 편집 후(-264)에도 FAIL이며, 이 카드는 적자를 깊게 했을 뿐 만들지 않았다. 상환은 이 카드 밖(t473 미푸시 브랜치, t492) 소관 — §E.2 원문 그대로 보존, 문구를 부드럽게 하지 않았다."

