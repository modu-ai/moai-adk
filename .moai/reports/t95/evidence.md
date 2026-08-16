# t95 증거 — 룰 문서 근거 교체 3건 (행동 변경 없음)

브랜치 `WT-t95`, base `origin/release/v3.1.1` = `5c31413728a252dfb8cc349529db408f6add670b` (카드 지시와 일치 확인).

## Claim (주장)

1. **항목 1 (t41 워크트리 락)**: `worktree-integration.md`의 락 서술 2곳(L1 용어표 Lifetime 칸, "`moai worktree` verbs are L2-only" [HARD] 블록)이 "설계된 동작" 프레임으로 교체됐다 — 세션이 실행되는 동안 유지·종료 시 해제, 죽은 세션의 락은 CC 2.1.210+에서 자동 해제, 폐기 시점의 잠긴 트리는 결함이 아니라 살아 있는 세션의 소유 표시이며 완화(unlock 안내)만 남는다.
2. **항목 2 (env -u 거부 설명)**: `kanban-dispatch.md`의 "argument-boundary misparse / argv[0]" 설명이 실제 의미론으로 교체됐다 — "정적으로 추적할 수 없는 셸 구조를 거부하며 끌 수 없다". 처방형(`unset … && <command>`)은 변경 없음. 같은 절의 서브셸 불렛도 통합된 의미론에 맞춰 정렬("다른 거부" → 같은 클래스의 거부).
3. **항목 3 (네이티브 /goal 금지 근거)**: `goal-directive.md` 스텁 1곳 + `goal-directive-detail.md` 3곳(77/79/91번 줄)의 "HUMAN-ONLY" 근거가 "평가자가 도구 호출을 못 해 기계 검증 조건 판정이 불가능"으로 교체됐다. `claude -p "/goal <condition>"` 비대화형 기동이 작동한다는 사실(같은 파일 81번 줄에 이미 문서화됨)을 근거로 명시.
4. 템플릿 미러 4종이 바이트 동일하게 갱신됐다(교체 전 로컬 파일과 바이트 동일임을 먼저 확인 — 중립화 델타 없음).

## Evidence (증거)

- before/after 발췌: `before-item1-worktree-integration.txt` / `after-item1-worktree-integration.txt`, `before-item2-kanban-dispatch.txt` / `after-item2-kanban-dispatch.txt`, `before-item3-goal-directive{,-detail}.txt` / `after-item3-goal-directive{,-detail}.txt` (모두 이 디렉터리).
- 예산 테스트(패스 1, 편집 직후): `budget-test-pass1.txt` — 76,559토큰.
- 예산 테스트(패스 2, 서브셸 불렛 정리 후 = 최종): `budget-test-pass2.txt` — `always-loaded surface = 76554 tokens, exceeds budget 76000 (overflow 554); surface has 17 entries`.
- 템플릿 테스트: `template-tests.txt` — `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/ -count=1` → `ok github.com/modu-ai/moai-adk/internal/template 34.027s` (rc=0).
- 미러 패리티: `mirror-parity.txt` — 4쌍 `diff` 무출력 + `PARITY-OK-4-FILES`. (교체 전 성립 확인: `git show HEAD:<path>` 결과물 `preedit-*.md` 4종과 템플릿 파일 `diff` → 전부 IDENTICAL.)
- vet: `vet.txt` — `go vet ./internal/config/... ./internal/template/...` rc=0, 출력 없음.
- lint: `lint.txt` — `golangci-lint run ./internal/config/... ./internal/template/...` → `0 issues.` rc=0.
- 항목 1 전제의 외부 검증: 공식 changelog(https://code.claude.com/docs/en/changelog) 2.1.210 항목 — "Fixed killed background sessions leaving a permanent `git worktree lock` behind; the periodic sweep now releases locks whose owning process is gone".
- 파일별 기여 분해(진단): `budget_breakdown.go` + 실행 결과 — always-loaded 17개 표면 열거, 총 76,586토큰(베이스라인).

## Baseline-attribution (baseline 귀속)

- 예산 측정은 전부 이 워크트리(브랜치 `WT-t95`)에서 이번 실행으로 관측: 베이스라인(편집 전, HEAD = 5c3141372) 76,586토큰 → 최종 76,554토큰. 이 카드 diff의 순영향 = **−32 토큰**(칸반디스패치 −217바이트, goal-directive +96바이트, 서브셸 불렛 −13바이트 등).
- `worktree-integration.md`·`worktree-state-guard.md`·`goal-directive-detail.md`은 `paths:` 프론트매터로 범위 한정 → 예산 계상 제외(분해 출력에서 확인). 예산에 계상되는 편집은 항목 2(칸반디스패치)와 항목 3 스텁(goal-directive)뿐.

## 결정 사항

- **예산 게이트 = 구조적 블로커로 리드 상보(상세는 최종 보고)**: 카드 전제("여유 ≤3 토큰")는 실측과 다르다 — 릴리즈 팁이 이미 **586토큰 초과**(76,586/76,000). 76,000 상향 커밋 이후 릴리즈에 머지된 선행 레인들의 룰 추가분이 원인이며, 릴리즈 브랜치 push는 CI 트리거(main 전용) 밖이라 미검출됐다(상수 주석이 서술한 것과 동일한 경로). 이 카드의 지정 구간 안에서는 도저히 green 불가 — 잔여 554토큰(≈2.2KB)을 본 카드 범위 밖 always-loaded 룰에서 자르는 것은 별도 다이어트 카드의 영역이고, 예산 재상향은 Go 소스 변경이라 카드 지시("no Go source changes expected")와 충돌한다. 따라서 3개 항목은 순영향 음수로 완료하고 게이트는 블로커로 보고한다.
- **t41 시대 문구의 소재**: 카드가 지목한 "원인 조사" 프레임은 두 파일에 문자 그대로는 없다. 실측 결과, 해당 문구 군은 t72(커밋 1344d6ce9, 2026-08-17 00:41)가 남긴 락/폐기 서술이고 t41(901e5244f, 같은 날 02:53)의 `worktree done` 수정과 같은 라운드에서 만들어졌다. "설계된 동작이 아니라 조사 대상"으로 읽히는 잔여 프레임("after the session releases its lock" 등)을 카드 지시대로 교체했다. `worktree-state-guard.md`에는 락 관련 문구가 없음을 grep으로 확인(음성 결과, 미변경).
- **native-invocation-model.md는 미변경(범위 밖)**: 이 파일의 분류 매트릭스(44번 줄)와 Axis B(66/77번 줄)는 여전히 `/goal`을 HUMAN-ONLY로 분류한다. 카드는 goal-directive 쌍만 지정했으므로 손대지 않았고, 교체된 근거와 이 파일 사이의 긴장은 잔여 위험으로 기록한다.

## Gaps (미검증)

- `TestAlwaysLoadedTokenBudget` **최종 red**(554 초과) — 위 결정 사항 참조. green 불가 입증은 분해 결과(17 표면, 파일별 기여)로 대체.
- 네이티브 /goal 평가자의 "도구 호출 불가" 속성은 카드 전제 + 이 저장소 자체 문서(goal-directive-detail.md 81번 줄: 턴마다 소형 모델이 "대화에서 Claude가 표면화한 것"을 검사)로 뒷받침했다. CC 공식 문서에서 평가자 메커니즘 페이지를 직접 대조하지는 않았다.
- `claude -p "/goal"` 기동 실측은 하지 않았다(저장소 문서 81번 줄의 기술 명세를 인용).
- 전체 스위트 미실행(레인 로컬 규율) — 전 판정은 CI.

## Residual-risk (잔여 위험)

- **배치 PR이 main에서 CI를 돌리면 `TestAlwaysLoadedTokenBudget`는 여전히 실패한다** — 본 카드와 무관한 선결 결함이므로, 배치 PR 전에 (a) 별도 다이어트 카드로 554+ 토큰 절감, 또는 (b) 근거를 붙인 예산 재상향 중 하나가 선행돼야 한다.
- `native-invocation-model.md`(44/49/66/77번 줄)의 HUMAN-ONLY 분류와 교체된 근거 사이에 표면 긴장이 남는다 — 후속 카드로 해당 매트릭스의 `/goal` 행을 정정할 것을 권고.
- 항목 1의 "자동 해제는 2.1.210+" 서술은 분산 템플릿에도 반영됐다(버전 숫자는 중립 클래스 — 템플릿 내 기존 "v2.1.139" 사례와 동일 취급, strict 리크 테스트 통과로 기계 확인).
