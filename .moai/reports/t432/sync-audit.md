# sync-audit — SPEC-CODEMAPS-REFRESH-001 (card t432)

**Auditor**: sync-auditor (독립 fresh-context 감사)
**일시**: 2026-09-02 / **측정 트리**: `.claude/worktrees/t432` @ `3b669d11f` (WT-codemaps-refresh, base `ad272be20`, 4커밋)
**Overall Verdict**: **PASS** (must-pass firewall: Functionality + Security 독립 통과)

---

## Evaluation Report

SPEC: SPEC-CODEMAPS-REFRESH-001 (docs-only, Tier M, evaluator profile: default — flat mode)

### Dimension Scores

| Dimension | Score | Verdict | Evidence |
|-----------|-------|---------|----------|
| Functionality (40%) | 100/100 | PASS | AC-CMR-001~008 전수 독립 재검증 PASS (아래 AC 표) — 8/8, 미이행 AC 없음 |
| Security (25%) | 100/100 | PASS | 변경 집합 14파일 전부 `.moai/` 문서 — Go 코드·시크릿·입력경로 변경 0 (`git diff --name-only ad272be20` 실측, 전부 허용 경로) |
| Craft (20%) | 90/100 | PASS | 문서 수치 클레임 12건 표본 재측정 전부 실측 일치 — 단, 증거 파일 §3.1 표제 "(26항목)" vs 본문 27행 불일치 1건 (F1, MINOR) |
| Consistency (15%) | 95/100 | PASS | 한국어 문서 컨벤션 유지, known-6 이월 규약 준수, frontmatter 전이 소유 규약 준수, close 주제 전체 SPEC-ID — §3.1 표제 오탈자 1건만 감점 |

- **Harmonic mean**: 96.1 · **Weighted**: 97.3
- **Must-pass firewall**: Functionality PASS · Security PASS → 강제 FAIL 조건 없음

### AC Must-Pass Matrix (감사자 독립 재측정 — 런 자가보고 재인용 아님)

| AC | 런 판정 | 감사 독립 재검증 | 감사 판정 |
|----|--------|----------------|----------|
| AC-CMR-001 재생성 완전성 | PASS | `git show 4548b947b --name-only` → 6문서 + provenance.json 전부 변경 실측; `ls .moai/project/codemaps/` → 7항목 | PASS |
| AC-CMR-002 경로 실존 표 | PASS | census 산수 자체 재계산 (85 디렉터리+15 파일=100; EXISTS 93=78+15; ABSENT 7=7 디렉터리) 정합; 표본 9경로(`internal/cli/factory.go` 등) 전부 실존 실측; absent 7건 전부 분류 기록 확인 | PASS |
| AC-CMR-003 패키지 대조 | PASS | `go list ./... \| wc -l` 재실행 → **137** (클레임과 일치); 유령=known-6+부정각주(bodp)만, `ls`로 7개 전부 트리 부재 실측; modules.md 패키지 수(64/131/137) 재측정 일치 | PASS |
| AC-CMR-004 식별자 hit/miss | PASS | 표본 5건 verbatim 일치 (`InitDependencies` deps.go:101, `RalphEngine.Decide` engine.go:34, `Registry.Register/Heartbeat/Deregister` registry.go:169/215/241); `ListActive` grep 0히트로 진짜 MISS 재확인; 카탈로그 11파일 `ls` 실측 = 12행 서술 정합 | PASS |
| AC-CMR-005 스탬프 도달성 | PASS | `git merge-base HEAD origin/develop` → `ad272be20abff…` = provenance `commit_sha` 일치; `--is-ancestor … origin/develop` → rc 0; worktree HEAD `3b669d11f` 아님 (merge-surviving) | PASS |
| AC-CMR-006 게이트 종결 | PASS | 감사자 직접 `moai graph check` 재실행 → `codemaps value=0 threshold=40 verdict=fresh`, mx-index/edges `verdict=absent` — progress §E.2 verbatim과 바이트 동일 | PASS |
| AC-CMR-007 범위 위생 | PASS | `git diff --name-only ad272be20` → 14파일 전부 허용 경로; `git status --porcelain` → 빈 집합 (untracked 0); gate.yaml `codemaps_changed_files: 40` 미변경 실측 | PASS |
| AC-CMR-008 증거 독립성 | PASS | 증거 파일 §1-§5 존재 + 감사자 자체 게이트 재판독 — 정확성 증거와 게이트 판정이 별개 근거로 병존 확인 | PASS |

### Accuracy substance (REQ-CMR-008 — "녹색 게이트 + 틀린 내용" 방어) 재측정

재생성 문서의 수치 클레임 12건을 표본 재측정 — **전부 실측 일치**:

| 클레임 (modules.md/overview.md/docs-truth.md) | 재측정 명령 | 결과 |
|---|---|---|
| cli 264 non-test / 최상위 202 | `find … ! -name '*_test.go' \| wc -l` | 264 / 202 일치 |
| hook 128 non-test | 동일 | 128 일치 |
| internal 디렉터리 64 / go-list internal 131 / 총 137 | `ls -d internal/*/` + `go list … grep -c` | 64 / 131 / 137 일치 |
| root 등록 60건, 33개 파일 | `grep -rh 'rootCmd.AddCommand' … \| wc -l` + `-l` | 60 / 33 일치 |
| cli 내부 패키지 임포트 93 | `go list -f '{{range .Imports}}' \| grep -c modu-ai` | 93 일치 |
| non-test 1074 / test 1714 | `find internal cmd pkg …` | 1074 / 1714 일치 |
| EventType 30 상수 | `grep -c 'EventType = '` | 30 일치 |
| go 1.26.4 | `grep '^go ' go.mod` | 1.26.4 일치 |
| 카탈로그 12 retained (11 파일) | `ls .claude/agents/moai/` | 11 파일 일치 |

**exit 계약 정정의 독립 확인**: data-flow.md §7의 정정 서술("0=전 계층 fresh / 1=임의 계층 stale **또는 absent** / 2=시스템 오류")를 감사자의 `moai graph check` 재실행이 직접 확인 — codemaps fresh + absent 2계층에서 **rc=1 실측** (정정된 계약과 행동 일치).

### Known-boundary integrity (t304 소관 보호)

- known-6 전부 원문 이월 확인 (modules.md): design(103행 경고 노트) · migrate(163) · state(248) · research(272) · evaluator(319) · **factory(169-173 서술 섹션 원문 유지)** — 삭제·무단 수리 0건. `internal/factory` 인용 파일 2건(`internal/cli/factory.go`, `launcher_blockcap_infinite.go`)은 트리 실존.
- `ListActive` 미적중 인용 본문 무수정 확인: data-flow.md 197/214/357행 원문 보존 + 증거 파일 §5.1 기록 — REQ-CMR-004("미적중은 기록만") 준수.
- bodp: "제거됐다"는 부정 인용 각주만, 패키지 서술 없음 — §1.2 분류와 일치.

### 진행 흔적 정합성

- 커밋 4건(a87e8ec2c→3b669d11f)의 파일 분포가 §E 서술과 1:1. run_commit_sha `4548b947b` 일치. sync_commit_sha `pending-backfill-t432` 플레이스홀더 + D3 근거 주석 존재 (커밋 자기참조 불가 — 규약 준수).
- 소유권 침범 없음: run 커밋의 spec.md diff = `status: draft→in-progress` 1행만, sync 커밋 = `in-progress→completed` 1행만 (`git show` 실측) — 본문 무수정, canonical owner 전이.
- close 주제 `docs(SPEC-CODEMAPS-REFRESH-001): sync-phase — 3-phase close (t432)` — 전체 SPEC-ID 규약 준수.

---

## Findings

- **F1** [MINOR] [optional] `.moai/reports/t432/codemaps-accuracy-verification.md` §3.1 표제 — "(26항목)"으로 표기했으나 표는 27행(HIT 26 + MISS 1)이고 판정 요약·progress.md는 27로 정확. 표제 단독 오탈자. - Required fix: 표제를 "27항목"으로 수정 (t304 후속 카드나 다음 문서 접촉 시 접촉 수리로 충분 — 단독 재커밋 불필요)

이 외 blocking·should-fix 결함 없음. 전 결함이 optional이므로 PASS→FAIL 전환 요인 없음 (finding-consumption discipline).

## Recommendations

- F1 표제 수정은 다음 접촉 수리로 흡수 권장 (단독 커밋으로 브랜치 흔들지 않음).
- provenance `generated_at`(14:31:10Z)이 data-flow.md 최종 mtime(14:32)보다 1분 이른 것은 게이트 중립(codemaps 문서는 described sources가 아님) — 기록만 남김.

---

## 5-Section Evidence-Bearing Verdict

- **Claim**: SPEC-CODEMAPS-REFRESH-001 sync-phase 종결 — 8/8 AC 충족, 재생성 문서가 현재 트리를 정확히 기술하며 t304 소관이 훼손되지 않았음.
- **Evidence**: 본 문서 전체 — 감사자 직접 실행 명령 18건의 관측 결과 (AC 표 + 수치 재측정 표 + `moai graph check` verbatim: `codemaps value=0 threshold=40 verdict=fresh` rc=1, absent 2계층).
- **Baseline-attribution**: 측정 트리 = `.claude/worktrees/t432` @ `3b669d11f` (HEAD), base `ad272be20`, 2026-09-02 본 감사 런 중 실행. divergence `22 4` (origin/develop 기준 — 리드 병합 창 소관).
- **Gaps**: (1) cross-model fan-out(`audit_multi`) 미실행 — 기계적 문서 전용 Tier M 판단으로 생략, claude 단일 감사. (2) 식별자 HIT 판정의 서명 수준 정합(인자·반환)은 census 자체가 범위 밖으로 선언 — 감사도 동일 범위. (3) census 전수 100행 중 9행 표본 인출(나머지는 산수 정합성으로 커버).
- **Residual-risk**: 스탬프 이후 described-source 변경이 다시 쌓이면 재적색 (임계값 40 — 재보정은 SPEC 명시 제외, new-findings #9 보고됨). 하위-목록 "언급" 토큰 매칭의 오판 여지는 census 자체가 공개한 대로 존재하나 NOTMENTIONED 0건 판정 방향에는 영향 없음.
