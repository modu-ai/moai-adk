# t1988 — GH #1926·#1924 재현·분류·수리 (Class B, run 기록)

## 리뷰 신선도 수리 기록 (2026-10-28, 리더 지시 2건)

**① 감사 SHA↔head 델타 실측** (`git diff --stat`, HEAD `1ab5ac7532`):
- `6e3cfbc4f7..HEAD`(verdict.md audited — merge-specless 판독기 바인딩): **보고서 2파일 전용** (progress.md +9, verdict.md ±4) — reports-only 예외 유지, 삼중 조건 만족.
- `7a263a0580..HEAD`(sync-audit-verdict.md audited — T11 기록 바인딩): **문서 전용 아님** — 108파일/+11,949. 원인: 그 범위 안에 origin/main 흡수 머지 `6e3cfbc4f7`(t1973 리바인드 정리·t1762 jevledger·t1977 docs-site — **이미 main에 착지·감사된 콘텐츠**)가 포함. 본 카드 고유 제품 델타는 e3b7e71450+8efe31c080의 4파일로 불변. T11/T13 게이트는 흡수 이전 정확 일치로 통과했고(소급 유효), 병합 시점 재검사는 verdict.md 바인딩(①)만 읽으므로 영향 없음.

**② codex review 재요청 1회**: `gh pr comment 1939 --body "@codex review"` → 요청 댓글 https://github.com/modu-ai/moai-adk/pull/1939#issuecomment-6452187560 (id 6452187560). 재요청 반복·수기 marker 없음.
- **수령 관측 완료**: 봇 `chatgpt-codex-connector` 응답 — **"Codex Review: Didn't find any major issues"**(지적 0건), 수령 마커 `codex-review-receipt:v1`, **바인딩 head `1ab5ac7532`**(요청 시점 head — 그 뒤 본 기록 커밋 6e39924b60는 보고서 1파일 델타), **수령 댓글 id 6452210595** (https://github.com/modu-ai/moai-adk/pull/1939#issuecomment-6452210595, 2026-10-27T20:00:08Z). 리뷰 head와 현 head 사이 델타는 보고서 전용.

## 체크포인트 (2026-10-27 11:58 KST 부근, 리더 지시 — 새 작업 금지·대기)

- 상태: PR #1939 OPEN @ HEAD `21703fd867` (pr-open 기록 완료, 리더 병합 창 대기)
- CI 봉쇄: main의 graph-freshness 적색(codemaps 53파일 stale, 임계 40)이 t1991 착지 전까지 모든 PR 검사(머지 ref 기준)를 봉쇄 — 본 카드 몫의 codemaps 수리는 t1991 소속으로 취소(착수했던 리프 작업은 산출 전 정지, 트리 무편집 확인: `git status --porcelain` 공)
- 병합: 리더 게이트(최신 head 전체 검사 + 리뷰 + 독립 PASS 무부채 + 양측 인계) — merged-pr 재실행도 리더가 다음 창에 수행
- t1984: 동결 유지 — §15.124 인계는 다음 조정자로 이관, 근거 기록 `.moai/reports/t1984/dispatch-record.md` 존속
- 잔여 후보: F2(템플릿 범위 diff 시 lane-local 필수 세트에 internal/template 해시 테스트 추가)·F3(kickoff 거부문의 런타임 해시 정의 포인터)
- run 전환: tnbn7e owner dead → tnjorw(pid 2314) live — 카드 레코드는 tnbn7e에 존속

Tree: `.moai/worktrees/t1988` branch `WT-triage-github-issues`, base main `cb30e05b3e` (확인: `git rev-parse --show-toplevel` 일치).
빌드: `make build` exit 0 — 프로브는 이 트리 빌드 `./bin/moai`로 수행(§2.2 tool-provenance: 설치 빌드 rc.38 @b1806498fe는 트리 HEAD의 조상 아님을 커밋 그래프로 확인).

## Issue #1926 (--spec 없는 Class C 카드가 plan-audit→kickoff 통과 불가)

| 주장 | 분류 | 근거 (명령 + 관측) |
|---|---|---|
| (a) 설치 빌드에 `factory set-spec` 부재 | **already-fixed** (구형 설치 빌드) | `./bin/moai factory set-spec --help` → 동사 헬프 출력("Set a picked card's SPEC identifier after pick (audit-trailed; gated by workflow.card_metadata_verbs.enabled)"); `git log -1 -- internal/cli/factory_card_metadata.go` → `bf0a35f10e [t1840] … (#1868)` 2026-10-21 21:38:02 +0000; `git merge-base --is-ancestor b1806498fe bf0a35f10e` → exit 1(설치 빌드 커밋이 착지 커밋의 조상 아님/부재). 등록: `factory_handoff_recover.go:94` |
| (b) `workflow.card_metadata_verbs.enabled` 기본 false 게이트 | **설계상 동작** (변경 금지 — 배차 제약) | `factory_card_metadata.go:152-154` 거부문이 키를 안내; `internal/config/types.go:675·1046` "Default false" 주석(REQ-CMV-001) |
| (c) queue state picked 한정 | **설계상 동작** | `factory_card_metadata.go:167` `requireQueuePicked`; 빌더 주석 :115-117 "the post-pick spec-id setter the unpick/re-pick round trip used to be the only route to" — unpick→set-spec→재임대가 정당 경로 |
| plan_artifact_hash 불일치 가설 | **여전히 결함(문서 드리프트) → 본 카드에서 수리** | 심사기는 verdict 줄과 런타임 해시를 직접 비교: `internal/homestate/card_evidence_readers.go:188-190` `runtime.NewInMemoryCache().ComputeHash(specDir)` 대조. 런타임은 byte-exact(`internal/runtime/audit_cache.go:127-159`, 경로+길이 구분자 프레이밍, t609 `1f9ce84d53` "preserve executable whitespace") 반면 안내 문서 `.claude/skills/moai/workflows/run/phase-execution.md:29`(와 템플릿 쌍둥이)는 "whitespace-normalized UTF-8"로 기술 — 감사자가 이 문구를 따라 계산하면 해시가 영구 불일치 → plan→kickoff 쐐기. SSOT(`spec-workflow-detail.md:101`)는 "as exact bytes"로 정확. **수리: 두 문서 사본을 byte-exact로 정정** (커밋 동반). 보강 writer `FileAuditCache.Store`는 비테스트 호출부 0건 — 카드 경로 해시 줄의 유일 생산자는 감사자 에이전트이므로 문서 정의가 곧 계약 |

## Issue #1924 (MOAI_FACTORY_ROLE=operator 미stamp · operator-action 카드 plan 탈출 불가)

| 주장 | 분류 | 근거 (명령 + 관측) |
|---|---|---|
| operator 마크를 찍는 코드 경로 없음 | **여전히 사실 — 단 게이트 자체는 의도적** | writer 전수: `os.Setenv(config.EnvFactoryRole` → `factory.go:858`(lane만); reader `mcp_refusal.go:149`; 상수 `envkeys.go:420` |
| 거부문 "the launcher or the operator environment stamps it" | **실결함(텍스트가 현실과 불일치) → 수리** | RED: `go test ./internal/cli/ -run TestFactoryUnknownAttributionErrorStatesManualStamp -count=1` → FAIL 2건("still claims the launcher stamps…", "must state stamping is manual") 원문 출력은 커밋 전 기록 본문 참조. 수리: `factoryUnknownAttributionError`가 수동 접두 형태(`MOAI_FACTORY_ROLE=operator`)를 명시하고 launcher는 lane만 stamp함을 밝히도록 정정. GREEN: 동일 명령 → `ok … internal/cli 1.276s`. gofmt -l → 0건 |
| (요청 2) 정당한 operator stamp 동사 신설 | **설계상 공백 — 리더/운영자 결정 대상** | 게이트 목적(무표시 세션 무권한)은 유지 배차 제약상 신기능 미구현. 이슈 자신도 "gate looks intentional and correct" 명시 |
| (요청 3) operator-action 카드의 plan 탈출 경로 | **설계상 공백 — 리더/운영자 결정 대상** | #1926 (b)(c)와 동일 가족: 현 정당 경로는 수동 `MOAI_FACTORY_ROLE=operator` 접두(거부문이 이제 정확히 안내) |

## 검증 배치 (이 트리, 이번 실행)

- RED: `go test ./internal/cli/ -run TestFactoryUnknownAttributionErrorStatesManualStamp -count=1` → FAIL (위 두 단언)
- GREEN: 동일 → `ok github.com/modu-ai/moai-adk/internal/cli 1.276s`
- 회귀: `go test ./internal/cli/ -run 'Attribution|FactoryDecide|OperatorAuthority|UnknownAttribution|MCPRefusal' -count=1` → `ok … 20.551s`
- 서식: `gofmt -l internal/cli/factory_card.go internal/cli/factory_unknown_attribution_test.go` → 출력 없음

## Gaps (명시적 미관측)

- plan-audit→kickoff end-to-end 재현 미실행 — 라이브 factory 레코드 변형이 필요해 배차 금지(원시 DB·운영자 역할 위조 금지)에 걸림; 메커니즘은 소스+심사기 레벨로 검증.
- internal/cli 전체 스위트 미실행(레인 로컬 범위 한정 — CI가 전체 실행). 플랫폼 매트릭스 미실행.
- 감사자가 문서가 아닌 임의 파이프라인으로 해시를 계산하는 경우의 런타임 방어 없음(심사기 거부문이 런타임 정의를 가리키지 않음) — 후속 후보, 본 카드 범위 밖.

## 잔여 위험

- 해시 알고리즘을 기술한 다른 문서가 남아 있을 가능성 — 검색 범위(.claude·templates·.moai/policies)에서는 두 사본뿐이었음.
- 템플릿↔로컬 문서 패리티는 본 수정으로 양쪽 동일 편집으로 유지.
