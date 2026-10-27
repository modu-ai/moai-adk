# t1988 verdict — GH #1926·#1924 트리아지·수리 (Class B)

verdict: PASS
audited_sha: 8efe31c080f1d49b86898620ee83cef99fe52bbe
machine-line-note: audited head 8efe31c080에서 레인 검증 배치와 codex_review가 관측됐고, 그로부터 현 HEAD까지의 델타는 .moai/reports/t1988/ 아래 증거 문서뿐임(diff --stat 관측) — merge_specless 판독기의 보고서 전용 델타 예외에 부합.

## Claim
1. #1924 거부문 결함("launcher가 operator 마크를 stamp한다")은 실재하며 본 카드에서 수리됐다 — 거부문이 수동 접두 형태를 정확히 안내한다.
2. #1926의 set-spec 부재 주장은 구형 설치 빌드에 한정된 관측으로, 현재 main 소스에는 존재한다(t1840/bf0a35f10e). 게이트 기본 false·picked 한정은 설계상 동작이다.
3. #1926의 plan_artifact_hash 불일치 가설은 문서 표면에서 실재 확인됐고 수리됐다 — 안내 문서가 런타임과 다른 알고리즘을 기술했던 드리프트.

## Evidence (명령 + 이번 실행 원본 출력)
- RED: `go test ./internal/cli/ -run TestFactoryUnknownAttributionErrorStatesManualStamp -count=1` → `--- FAIL: … refusal still claims the launcher stamps the operator mark … / refusal must state stamping is manual …` (2단언 실패)
- GREEN: 동일 명령 → `ok github.com/modu-ai/moai-adk/internal/cli 1.276s`
- 회귀: `go test ./internal/cli/ -run 'Attribution|FactoryDecide|OperatorAuthority|UnknownAttribution|MCPRefusal' -count=1` → `ok … 20.551s`
- 서식: `gofmt -l internal/cli/factory_card.go internal/cli/factory_unknown_attribution_test.go` → 출력 0건
- 바이너리 프로브(본 트리 빌드, `make build` exit 0): `./bin/moai factory set-spec --help` → 동사 헬프 출력
- 소스 귀속: writer `factory.go:858`(lane만)·reader `mcp_refusal.go:149`·상수 `envkeys.go:420`; 해시 런타임 `audit_cache.go:127-159`(byte-exact, t609 `1f9ce84d53`); 심사기 `card_evidence_readers.go:188-190`; SSOT `spec-workflow-detail.md:101`

## Baseline-attribution
- Tree: `.moai/worktrees/t1988`, base `cb30e05b3e`, head `e3b7e71450`(4파일 +36/−5, card id 커밋 메시지 포함)
- 도구 귀속: 트리 빌드 `./bin/moai`(§2.2) — 설치 빌드 rc.38 @b1806498fe는 `git merge-base --is-ancestor b1806498fe bf0a35f10e` exit 1로 착지 커밋의 조상 아님을 관측

## Gaps
- plan-audit→kickoff end-to-end 재현 미실행(라이브 factory 레코드 변형 수반 — 배차 금지 항목). 메커니즘은 소스·심사기 레벨 검증.
- internal/cli 전체 스위트·플랫폼 매트릭스 미실행(레인 로컬 한정, CI가 전체 담당).
- codex_review(card)는 정적 리뷰로 테스트·빌드 미실행을 명시 — pass, findings 0건.

## Residual-risk
- 감사자가 문서가 아닌 임의 파이프라인으로 해시 계산 시 심사기 거부문이 런타임 정의를 가리키지 않아 재발 여지(후속 후보).
- 기타 문서의 동일 패러프레이즈 잔존 가능성(검색 범위 내 0건).

판정: **PASS** — 수신 영수증 codex_review(card, base cb30e05b3e, pass) + 상기 배치. 최종 판정은 리더의 증거 읽기.
