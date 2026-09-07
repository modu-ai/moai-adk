# SPEC Review Report: SPEC-FMT-GATE-001 (card t465)

- **Iteration**: 1/2 (Tier M ceiling)
- **Verdict**: **PASS** (skip-eligible)
- **Overall Score**: **0.923** (harmonic mean; Tier M PASS threshold 0.80)
- **Auditor**: plan-auditor (독립 감사, M1 Context Isolation 적용 — 저자 추론 맥락 없이 산출물만 판독)
- **Audit tree**: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t465` @ `WT-format-gate-zero` HEAD `9e1b6a379` (= base `d592b0551` + 문서 커밋만; `origin/develop` 실측 `d592b0551`)

---

## Claim

SPEC-FMT-GATE-001의 plan-phase 산출물 3종(spec.md / plan.md / acceptance.md)은Tier M
기준 0.80 이상으로 PASS 가능하며, 카드 t465의 제약 1–4가 각각 기계 판정 가능한 REQ/AC 쌍으로
인코딩되어 있고, 활성 전용(activation-only) 범위 규율이 지켜졌다. 기저 사실 주장(게이트 0개,
표면 분리, 기준선 154)은 전부 이 트리에서 재실측해 뒷받침된다. 차감 요인은 MINOR 3건뿐이다.

## Evidence — 카드 제약 → REQ/AC 쌍 커버리지 (task 절차 3)

| 카드 제약 | REQ/AC 쌍 | 기계 판정 형태 | 판정 |
|---|---|---|---|
| 1. t457 선행 착지 | REQ-FG-003 / AC-FG-003 (+ plan §C 전제 게이트·사후 감사 형태) | `git merge-base --is-ancestor e1fdf00d1 <act>` → exit 0. **고정 SHA 사용** — 이동 ref 의존 없음(R1 핀; `e1fdf00d1` 실측: `refs/heads/WT-gofmt-drift` tip `e1fdf00d16b348ba7df090318e23edba6693cf13`과 일치) | 커버됨 |
| 2. 활성 시점 녹색 | REQ-FG-004 / AC-FG-004 (+ REQ-FG-001/002 이진성, plan §E 뮤턴트 프로브 "1파일 변형 → exit 1") | `gofmt -l . | wc -l` → `0` @ activation tree | 커버됨 |
| 3. 표면 구분 (배포 vs 로컬) | §B 3-표면 표 + REQ-FG-001/002/004(루트 CI) · REQ-FG-006(로컬 Makefile) · REQ-FG-005(배포 금지); 기각 사유 문서화: pre-commit(§B 사유2 + Out of Scope), `moai gate` format 스텝(§B 사유3 + Out of Scope "배포 표면") | AC-FG-001..004(명령 이진), AC-FG-006(make fmt-check 이진), AC-FG-005(diff 0행) | 커버됨 |
| 4. 템플릿 중립성 | REQ-FG-005 / AC-FG-005 | `git diff --name-only d592b0551..HEAD -- internal/template/templates/ | wc -l` → `0` (base 고정 SHA 핀) | 커버됨 |

## Evidence — 기저 사실 주장 스팟 검증 (task 절차 5; 전부 이번 런·이 트리 실측)

| # | 주장 (spec.md 위치) | 내가 실행한 명령 | 결과 | 판정 |
|---|---|---|---|---|
| a | 배포 템플릿 `.github/workflows/`는 `label-sync.yml` 단 1개 (spec.md:L46, plan.md:L60-61) | `ls internal/template/templates/.github/workflows/` | `label-sync.yml` 1파일만 (루트는 19파일, ci.yml 포함) | **VERIFIED** |
| b | `core.hooksPath` = repo-local `/dev/null` (spec.md:L28, L47) | `git config --show-origin core.hooksPath` | `file:/Users/goos/MoAI/moai-adk-go/.git/config` → `/dev/null` | **VERIFIED** |
| c | 루트 CI Lint 잡에 포맷 스텝 부재·무조건 실행 (spec.md:L26, L51) | `grep -rn 'gofmt' .github/workflows/` → 0히트(exit 1); ci.yml L420-458 판독 | lint 잡(L422)에 `if:` 없음(무조건 required check), 스텝 = checkout→setup-go→golangci 설치→templ drift guard→`golangci-lint run`→build→agent lint — gofmt 스텝 없음 | **VERIFIED** |
| d | `.golangci.yml` enable-set `{errcheck, govet, ineffassign, staticcheck, unused}` — 포맷터 부재 (spec.md:L26) | `.golangci.yml` 판독 (L23-30) | enable = errcheck/govet/ineffassign/staticcheck/unused 정확 일치; L12 "widening the set is a deliberate future decision" 주석도 spec.md:L119-120 인용과 일치 | **VERIFIED** |
| e | 기준선 `gofmt -l .` @ `d592b0551` → 154 (spec.md:L31-32) | `wc -l .moai/reports/t465/gofmt-l.txt` → `154`; **신규 재실측** `gofmt -l . \| wc -l` → `154`, tracked 변형 `git ls-files -z '*.go' \| xargs -0 gofmt -l \| wc -l` → `154` (HEAD 9e1b6a379 = base+문서커밋만, .go 불변) | 154/154/154 — 카드의 낡은 3bdd5a803 수치가 아니라 **fresh baseline 인용** 요건 충족 | **VERIFIED** |
| f | 154 목록에 `*_templ.go` 생성물 0건, `_templ` 그렙 1히트는 `runner_template_test.go` (plan.md:L21-23) | `grep -c '_templ' .moai/reports/t465/gofmt-l.txt` → `1`; `grep '_templ' …` → `internal/harness/v4manifest/runner_template_test.go` | 주장과 정확 일치 | **VERIFIED** |
| g | testdata fixture `.go` 2개 (plan.md:L24-25, spec.md:L94-95) | `grep 'navigator/astx/testdata' .moai/reports/t465/gofmt-l.txt \| wc -l` | `2` | **VERIFIED** |

## Baseline-attribution

- 위 표의 모든 측정은 **이번 런, 이 트리**(worktree `t465` @ `9e1b6a379`)에서 직접 실행한
  명령과 그 출력이다. gofmt 재실측 154는 base `d592b0551`과 .go 내용이 동일한 트리(브랜치
  diff = SPEC 문서 4종 + 증거 1종, `git diff --name-only d592b0551..HEAD` 실측)에서의 재현이며,
  이는 곧 브랜치가 아직 `.go` 파일·템플릿을 0건 수정했음(AC-FG-005 사전 상태 양호)도 보여준다.
- `e1fdf00d1` 핀 검증: `git show-ref --verify refs/heads/WT-gofmt-drift` →
  `e1fdf00d16b348ba7df090318e23edba6693cf13` (plan.md:L4 "tip e1fdf00d1" 주장과 일치).
- SPEC 커밋: `git show -s 9e1b6a379` → `feat(SPEC-FMT-GATE-001): plan-phase artifacts (Tier M, 3 artifacts) (card t465)`.

## MUST-PASS criterion table

| MP | 항목 | 판정 | 근거 (spec.md 줄번호) |
|---|---|---|---|
| MP-1 | REQ 번호 정합성 | **PASS** | REQ-FG-001(L57)…REQ-FG-006(L84) — 6연속, 공백·중복 없음 |
| MP-2 | EARS/GEARS 형식 (판정 계층: **요구 계층 spec.md §C만** — AC는 검증 계층으로 Given-When-Then가 올바른 형식이며 본 항목에서 감점하지 않음) | **PASS** | When(이벤트) ×3: L59/L64/L75, Where ×2: L69/L86, shall not(unwanted) ×1: L80-82 — 6/6 GEARS 패턴. 부기: REQ-FG-006의 "Where 개발자가 … 실행하면"은 문맥상 When-동치로 읽힘(패턴 위반 아님, MINOR 문구 차원) |
| MP-3 | YAML frontmatter 12 필드 | **PASS** | L2-14: id/title/version("0.1.0" 인용)/status(draft∈enum)/created/updated(ISO)/author/priority(P1)/phase("v3.1.4 target" — 릴리스 타깃 라벨, 금지값 plan/run/sync/mx 아님)/module/lifecycle(spec-anchored)/tags(CSV) + tier: M(선택필드). snake_case 별칭 0건 |
| MP-4 | 언어 중립성(§22) | **N/A (auto-pass)** | 단일 언어(Go) 프로젝트 SPEC. 템플릿 중립성 제약은 REQ-FG-005/AC-FG-005로 별도 구속 (커버 표 참조) |
| MP-5 | D7 cross-SPEC 정합 | **PASS** | `grep -rn 'SPEC-…' .moai/specs/SPEC-FMT-GATE-001/ \| grep -v 자신` → 0매치(exit 1). 외부 SPEC 참조 없음 → BLOCKING 없음 |
| MP-6 | D8 cross-platform (syscall) | **PASS** | `grep -rn 'syscall' <SPEC dir>` → 0매치(exit 1) — 자동 통과 |
| MP-7 | Clarification gate | **PASS** | `grep -rn '\[NEEDS CLARIFICATION' plan.md research.md` → plan.md 0매치; research.md 부재(Tier M 정상 — 산출물 3종 + progress.md, `ls` 실측) |

## Category Scores (rubric-anchored)

| Dimension | Score | Band | 근거 |
|---|---|---|---|
| Clarity | 0.75 | "1-2개 요구에 경미한 모호성, 합리적 엔지니어가 일관되게 해석 가능" | 전 기계 판정 명시로 대체로 모호함 없음. 유일 예외 D1: REQ-FG-006의 문자 그대로 판정식(`gofmt -l .`, L86-87)이 spec §D(L97-98)·AC §D.3(L82-83)이 표준으로 지정하는 tracked-files 변형과 untracked `.go` 코너에서 충돌 — 같은 문서 내에서 해석은 수렴하므로 0.75 밴드 |
| Completeness | 1.0 | 전 섹션·전 필드 | HISTORY(L122-125)·WHY(§A)·범위(§B/§E)·REQUIREMENTS(§C 6건)·AC(acceptance.md 6건)·Out of Scope H3 3개 각 특정 불릿(L104/110/115) — 모두 존재. plan.md Tier M 전체 구조(Context/Known Issues/Pre-flight/Constraints/Self-Verification/Milestones/Anti-Patterns/Cross-References) |
| Testability | 1.0 | 전 AC 이진 판정 | AC-FG-001..006 전부 명령+기대 출력 이진 판정(Given-When-Then). 위즐 워드 0건. 제약1 인코딩은 고정 SHA(이동 ref 의존 없음), `<act>` 플레이스홀더는 §D.5 기록 의무로 폐쇄. 기준선 154는 내 재실측으로 재현됨 |
| Traceability | 1.0 | 전 쌍 커버 | REQ↔AC 1:1 (§D.2 매트릭스 L71-78), 마일스톤 매핑, 고아 AC/미커버 REQ 0건 |

**Harmonic mean** = 4 / (1/0.75 + 1 + 1 + 1) = **0.9231** ≥ 0.80 (Tier M) → **PASS, skip-eligible**

## 범위 규율 (task 절차 4) — activation-only, 정리 흡수 없음: 판정 **양호**

- 인도물 = ci.yml format-gate 스텝(M1) + Makefile `fmt-check` 타깃(M2) 단 2파일. 154개 정리는
  §E "Out of Scope — 154개 파일 포맷 정리"(L115-118) + plan §G "drive-by `gofmt -w` 금지" +
  §D.5 종결 게이트 "본 카드 커밋의 `gofmt -w` .go 수정 0건"(acceptance.md:L103) 삼중 차단.
- gofumpt 상향·`.golangci.yml` 확장·pre-commit·`moai gate` format 스텝 전부 Out of Scope로
  명시적 박제 — 배포 제품 변경 흡수 없음. 브랜치 실측 diff(base→HEAD)도 문서 5종뿐.

## Defects Found (D1..D3 — BLOCKING 0건, SHOULD-FIX 0건, MINOR 3건)

**D1** — `spec.md:L86-87` (REQ-FG-006) — REQ 문자 그대로의 판정식 "`gofmt -l .` 출력이 1행
이상일 때에만 non-zero"는 전수형이지만, 같은 SPEC의 spec §D(L97-98)와 acceptance §D.3
untracked 엣지 케이스(L82-83: "untracked `.go`가 판정을 바꾸지 않음")는 tracked-files 변형을
표준으로 지정한다. 워크트리에 정리 안 된 untracked `.go`가 있으면 두 규정이 충돌한다 — Severity:
**MINOR** — Class: optional — 구체적 실패: REQ 문면만 따라 전수형으로 구현한 엔지니어의
`make fmt-check`는 untracked 스크래치 파일에 붉어지고 AC §D.3 엣지 케이스가 실패한다. —
Required fix: REQ-FG-006의 판정식을 "`git ls-files -z '*.go' | xargs -0 gofmt -l` 출력이
1행 이상일 때"로 문면부터 tracked 변형으로 교체(§D·AC §D.3과 정렬).

**D2** — `acceptance.md:L44-46` (AC-FG-005) — 판정 범위의 끝점이 `d592b0551..HEAD`(branch
tip)인데 §D.5 종결 게이트(L100)가 progress.md에 기록하도록 요구하는 SHA는 activation SHA뿐이다.
M2의 Makefile 커밋은 activation **이후**에 착지하므로, 이후 세션이 AC-FG-005를 사후 재판정하려면
최종 tip SHA가 없어 `HEAD`를 써야 한다(그 시점 HEAD는 병합 후 다른 작업을 포함할 수 있음) —
Severity: **MINOR** — Class: optional — 구체적 실패: 사후 감사 세션이 AC-FG-005를 기록된
SHA만으로 재실행할 수 없다. — Required fix: §D.5에 "최종 branch tip SHA"를 기록 의무에 추가.

**D3** — `spec.md:L69-71` (REQ-FG-003) — 고정 핀 `e1fdf00d1`은 현재 `WT-gofmt-drift`
tip(실측 일치)이지만 t457이 아직 병합 창 대기 중이다. t457 브랜치가 병합 전 재작성(rebase/amend)
되거나 squash 병합되면 (a) 핀 SHA가 고아가 되거나 (b) 작업이 착지했어도 조상 관계가 성립하지
않아, 어느 쪽이든 전제 게이트가 **영구 통과 불능**이 되는데 SPEC이 그 사태의 감지·재고정
절차를 적지 않는다 — Severity: **MINOR** — Class: optional(공정 외부 가설; 이 리포의
--no-ff 레인 병합 규율로 실현 가능성 낮고, 카드가 이 인코딩을 명시적으로 처방함) —
Required fix(경화): "t457 tip 변경 시 SPEC 수정(MINOR amendment)으로 핀 재고정"이라는
구제 경로 한 줄을 plan §C 실패 분기에 추가.

## Gaps (명시적으로 관측하지 않은 것)

- `make fmt-check` 타깃의 실제 거동 — M2 산출물이 아니라 아직 존재하지 않아 판정 대상 없음
  (RED 상태 자체는 plan §E 표에 명시된 기대와 일치).
- activation 커밋 생성 이후의 CI Lint 잡 실측 판정 — run-phase 소관(본 감사는 plan-phase).
- `internal/cli/gate.go` 스텝 구성 주장(spec.md:L27)은 gate.yaml 선두부(vet/typecheck/lint/
  test 예산 + ast_grep_gate 언급)로만 확인 — format 스텝 부재는 확인됐으나 전체 스텝 열거의
  전수 대조는 하지 않았다.

## Residual-risk

- D3의 시나리오(t457 재작성/squash)가 현실화하면 활성이 교착된다 — 카드·리드 운영으로
  흡수 가능하나 문서상 구제 경로는 없다.
- CI Go 버전 상승이 gofmt 규칙을 바꾸면 활성 후 게이트가 붉어질 수 있다(acceptance
  §D.3 "CI Go 버전 드리프트" 항목이 이미 명시하고 있는 알려진 위험).
- 본 감사 중 브랜치·작업트리는 무변경(보고서 파일 1건 추가 커밋 외). 감사 산출물이므로
  별도 카네이션 없이 명시적 pathspec으로만 스테이징한다.

## Recommendation

**PASS (0.923, skip-eligible).** 4개 카드 제약 모두 기계 판정 가능한 REQ/AC 쌍으로 인코딩됐고,
기저 사실 주장 7건 전부 이 트리에서 재실측해 확인됐으며, 정리(t457) 흡수·템플릿 변경 모두
없다. MINOR 3건(D1-D3)은 run-phase 착수를 막지 않는다 — D1/D2는 run-phase 첫 커밋 전
문서 1-2줄 수정으로 반영 권장(선택), D3은 리드의 t457 병합 창 운영 시 유의사항으로 인계.
Implementation Kickoff Approval(인간 게이트)은 본 PASS와 무관하게 그대로 필수다.
