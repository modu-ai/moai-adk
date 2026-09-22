# t240 판정서 — §H 오버레이 문서 정정 (AC-AMP-006 반전 반영)

- 판정: **FIXED (카드 채택·구현 완료)** — 전제는 반증되지 않았다. 조사 초반 t175(구 방향)와의 충돌로 DROPPED를 검토했으나, `SPEC-V3R6-AUDIT-MODEL-PIN-001/spec.md §H:215-221` + `acceptance.md:70-106`의 AC-AMP-006 개정(2026-08-24, 리드 승인)이 **신 방향**(top-level `reasoning_effort` 실효 / thinking-budget 객체 무시, null 1.02)이고 §H가 "템플릿 llm.yaml 오버레이 문서 정정은 별도 후속 카드"로 명시한 것이 바로 본 카드다.
- 지배 기록 우선순위: AC-AMP-006 개정(4런 null-제어 differential, 경계 1.25/널 1.1, 비 1.34/1.85/1.48 — 신·리드승인) > t175 세션-경로 프로브(구 방향, superseded).

## 변경 (이름 나열)

- `internal/template/templates/.moai/config/sections/llm.yaml` — 상단 llm: 오버레이 노트(1행 "thinking cannot be disabled"→"reasoning", "(thinking enabled)" 제거) + 하단 glm.effort 와이어 스테이트 measured 절(:231-236)을 신 방향으로 교체, delivery-field 사실 추가. 템플릿 중립성 유지(SPEC-ID·카드id·날짜 0).
- `internal/template/glm_effort_overlay.go` — `SessionGLMReasoningState` 상단 주석: 구 measured 절(t175 인용) → 신 방향 + t175 superseded 명기.
- `internal/cli/glm.go` — `glmReasoningEnvVars()`(:391)·`glmReasoningEnvVarsForModel()`(:425) 두 "Delivery status MEASURED" 주석 블록 동일 정정. 주석 전용 변경, 동작 무변경.
- `.moai/specs/SPEC-MODEL-TIER-PLANTYPE-001/progress.md` — AC-MTP-032b 행에 잔여 종결 포인터 추가 + §E.4 KNOWN RESIDUAL 블록 뒤 ADDENDUM(카드 t240, 2026-09-02) 추가. 원문 보존·부가형.
- 증거: `.moai/reports/t240/verdict.md` (본 파일)

## 5-섹션 증거

**Claim**: 문서 3종+SPEC 기록 1종이 AC-AMP-006 개정 방향과 일치하고, 구 방향 주장이 배포 템플릿·코드 주석에 잔존하지 않는다.

**Evidence**:
- `make build` → `go build -ldflags ... -o bin/moai ./cmd/moai` (재임베드 성공, 2회)
- `go build ./internal/template/... ./internal/cli/...` → `BUILD_OK`
- `go vet ./internal/template/...` → (출력 없음) `VET_TEMPLATE_OK`
- `go test ./internal/template/` → `ok github.com/modu-ai/moai-adk/internal/template 25.300s`
- 구 방향 잔존 스윕: `grep -rn "silently IGNORES\|honors the Anthropic" internal/` → 0건. `grep -n "thinking" llm.yaml` → 신 방향 서술의 일부("a thinking-budget request object is ignored")만 잔존.
- 중립성: `grep -nE "t22[0-9]|t240|SPEC-|20[0-9]{2}-[0-9]{2}" llm.yaml` → 0건.
- `git diff --stat`: 4 files changed, 43 insertions(+), 29 deletions(-)

**Baseline-attribution**: 본 워크트리(`WT-overlay-effort-docs`), develop `2660bcd09` 흡수 기준(배차 문서의 b7462203a 이후 병렬 레인 t364 통합으로 develop 팁이 전진한 것을 흡수 — 병합 무충돌), 2026-09-02 본 런.

**Gaps**:
- `internal/cli` 전체 스위트 미실행(600s 하한·주석 전용 변경). `go build`+`go vet` 컴파일 판정으로 대체 — `git diff`가 glm.go를 주석 전용으로 증명.
- z.ai 엔드포인트의 물리적 재측정 미수행(카드 범위 밖 — 기록된 개정을 반영하는 문서 작업).

**Residual-risk**:
- `SPEC-GLM-EFFORT-MAX-001/spec.md:45`와 `CHANGELOG.md:256`은 여전히 구 방향("thinking honored / reasoning_effort ignored")을 MEASURED로 기록 — **본 카드 미수정**(전자는 종결 SPEC 본문이라 소관상 lane이 고칠 수 없고, 후자는 역사 기록). 후속 카드 권고: SPEC-GLM-EFFORT-MAX-001 본문 정정(manager-spec 소관)으로 리드 결정 바람. 이 방치 시 리포에 상반되는 "measured" 기록 2벌이 공존.
- AC-AMP-006은 audit 경로(`glm_audit`) 측정이고 t175는 세션 경로 프로브 — 두 경로의 물리적 동일성은 어느 기록도 증명하지 않았다. 본 판정은 기록된 지배 개정을 따른 것.
- t225 비차단 R1(insertAuditLeaf subtree-end 스캔이 주석 건너뜀)·R2(childless-ancestor 기본 step=2)는 **정리하지 않음** — 코드 경로 변경이라 본 문서 카드 범위 밖, 별도 카드 필요.
- 로컬 `.moai/config/sections/llm.yaml`(미추적, 사용자 값)은 미접촉 — 다음 `moai update`가 템플릿에서 재배포.
