# SPEC-HARNESS-EVO-PIPE-REPAIR-001 — Acceptance Criteria

status: draft · Tier M · 검증 원칙: 모든 AC는 관측 가능한 명령 출력/파일 상태로 판정 (verification-claim-integrity §3 — Claim에는 반드시 verbatim Evidence 첨부)

---

## §D AC 매트릭스

### 파이프 수리 (Go)

**AC-HEP-001a (어휘 정렬 — 채택)** [MUST]
- Given: tier-promotions.jsonl 실데이터 형태의 fixture — `{"pattern_key":"user_prompt::","from_tier":"","to_tier":"auto_update","observation_count":196,"confidence":1}` 및 `to_tier:"rule"` 변형
- When: `proposalgen.MapPromotions` 호출
- Then: candidates ≥ 1 (auto_update 및 rule 각각 채택); `go test -run TestMapPromotions ./internal/harness/proposalgen/` PASS 출력 인용

**AC-HEP-001b (어휘 정렬 — 배제)** [MUST]
- Given: `to_tier: "observation"` 및 `to_tier: "heuristic"` promotion fixture (confidence ≥ 0.70)
- When: `MapPromotions` 호출
- Then: candidates = 0 (pre-actionable tier 배제 유지)

**AC-HEP-002a (스키마 정렬 — 실데이터 통과)** [MUST]
- Given: 실측 pattern_key 3형 — `user_prompt::`(빈 subject+hash), `agent_invocation:Bash:`(빈 hash), `session_stop::`
- When: mapper 형식 검사
- Then: prefix/빈-segment 사유의 거부 0건 (tier/confidence 조건 충족 시 채택)

**AC-HEP-002b (스키마 SSOT 파생)** [MUST]
- Given: 수리 후 mapper 소스
- When: `grep -n "code_change\|error_pattern\|tool_failure\|repeated_edit" internal/harness/proposalgen/mapper.go`
- Then: 수기 병렬 prefix 목록 잔존 0 (EventType enum 파생 구현으로 대체됨을 코드 리뷰로 확인); `apply_outcome`은 파생 집합에서 제외됨

**AC-HEP-003a (Stop 경로 classify 자동 실행)** [MUST]
- Given: usage-log에 이벤트가 축적된 테스트 프로젝트(t.TempDir)
- When: Stop 경로 훅 핸들러 실행 (e2e 테스트)
- Then: tier-promotions 갱신 또는 classify 실행 흔적(로그 라인) 관측; 테스트 출력 인용

**AC-HEP-003b (fail-open)** [MUST]
- Given: classify가 오류를 반환하는 조건 (손상된 usage-log 등)
- When: Stop 경로 훅 실행
- Then: 훅 exit 0 (세션 종료 비차단); stderr에 오류 표기 허용

**AC-HEP-004 (PostToolUse observer 등록)** [MUST]
- Given: `make build` 후 렌더된 settings
- When: `grep -n "handle-harness-observe.sh" internal/template/templates/.claude/settings.json.tmpl .claude/settings.json`
- Then: 양쪽 모두 PostToolUse 항목 존재 (기존 stop/subagent-stop/user-prompt-submit 3종 등록과 공존)

### 스모크 게이트

**AC-HEP-005a (doctor 정상 하네스 PASS)** [MUST]
- Given: 유효한 v4 하네스 fixture (command + `<name>/manifest.json` + Runner + 실존 specialist)
- When: `moai harness doctor` (또는 확정 verb)
- Then: exit 0 + 4축 cross-ref 결과 보고

**AC-HEP-005b (0-harness graceful)** [MUST]
- Given: 하네스 0개 프로젝트 (t.TempDir)
- When: doctor 실행
- Then: exit 0 + "no harnesses" 상당 안내

**AC-HEP-006 (결함 클래스 탐지 — 게이트 RED 증명)** [MUST]
- Given: B5 결함 재현 fixture — Runner의 manifest 경로 상수가 비실존 경로, agent 참조가 비실존 이름
- When: doctor 실행
- Then: findings ≥ 2 (wrong-manifest-path + unresolved-agent 각 1건 이상), 비-0 exit 또는 결함 보고 출력

**AC-HEP-007a (exemplar 수리 전 실전 탐지)** [MUST]
- Given: 수리 전 실제 `.claude/workflows/harness-release-update-run.js`
- When: doctor 실행 (run-phase M4 기록)
- Then: `:31` MANIFEST_PATH 결함 + `:56` `harness-devkit-release-update-specialist` 결함 2건 보고 — verbatim 출력을 progress.md §E.2에 기록

**AC-HEP-007b (exemplar 수리 후 클린)** [MUST]
- Given: 수리 후 Runner (`.claude/commands/harness/release-update/manifest.json` + `harness-release-update-specialist` 참조)
- When: doctor 재실행
- Then: 해당 하네스 findings 0; `git diff` 범위가 로컬 파일 1개로 한정(템플릿 무변경: `git status internal/template/` 청정)

**AC-HEP-014 (Builder ACTIVATE 계약)** [MUST]
- Given: 수리 후 `harness-builder.md` ACTIVATE 절 (live + template mirror)
- When: `grep -n -i "doctor\|smoke" .claude/skills/moai/workflows/harness-builder.md`
- Then: ACTIVATE 완료 조건에 스모크 게이트 실행 + 결함 시 활성 선언 금지 계약 문구 존재

### 디스패처 / 문서 정합

**AC-HEP-008 (가드 완전화)** [MUST]
- When: `grep -n "list\|edit\|remove\|doctor" .claude/skills/moai/workflows/harness-build-entry.md` (Phase 0 절)
- Then: 예약 동사 집합에 `list/edit/remove/doctor` 포함 (기존 4-verb + 신규 — 총 8-verb); template mirror 동일

**AC-HEP-009 (help 현행화)** [MUST]
- When: `.claude/commands/moai/harness.md` argument-hint + `.claude/skills/moai/SKILL.md` 라우터 행(:70 상당) + §harness 디스패처 절 grep
- Then: `list/edit/remove/doctor` 열거; template mirror 동일

**AC-HEP-010 (manifest 단일 정본)** [MUST]
- When: `grep -n "\.claude/harness/" .claude/skills/moai/workflows/harness-builder.md`
- Then: 0 matches (OR-분기 제거); 정본 `.claude/commands/harness/<name>/manifest.json` 단일 서술; template mirror 동일

**AC-HEP-011 (FROZEN 정렬)** [MUST]
- When: harness.md Layer 1 목록 + learner SKILL.md L1 목록 grep
- Then: `.claude/agents/harness/`가 FROZEN 목록에서 제거되고 허용 쓰기 대상으로 서술; `.claude/agents/moai/`는 FROZEN 유지 — `frozen_guard.go` `allowedPrefixes`/`frozenPrefixes`와 문서가 일치; template mirror 동일

**AC-HEP-012 (rate-limit SSOT)** [MUST]
- When: `grep -n "7-day\|7일\|per week\|max_per_week" .claude/skills/moai/workflows/harness.md`
- Then: "1 per 7-day invariant" 서술 잔존 0; `harness.yaml` `rate_limit`(3/week + 24h cooldown) 기준 서술 + 구 REQ-HRN-FND-018 floor 주장의 supersede provenance note 존재; template mirror는 generic prose (SPEC ID 없음)

### 횡단

**AC-HEP-013a (Template-First parity)** [MUST]
- When: 수정된 8개 문서/설정 표면 각각 `diff <live> <template mirror>`
- Then: 의도된 차이(§25 중립성 목적의 SPEC-토큰 제거) 외 불일치 0; `make build` 성공 출력 인용

**AC-HEP-013b (중립성 + 네임스페이스 누출 차단)** [MUST]
- When: `grep -rn "HARNESS-EVO-PIPE-REPAIR\|REQ-HEP" internal/template/templates/` + `go test -run TestSplitHarnessNamespaceNoLeak ./internal/template/`
- Then: grep 0 matches + 테스트 PASS

---

## REQ ↔ AC 추적표

| REQ | AC |
|-----|-----|
| REQ-HEP-001 | AC-HEP-001a, 001b |
| REQ-HEP-002 | AC-HEP-002a, 002b |
| REQ-HEP-003 | AC-HEP-003a, 003b |
| REQ-HEP-004 | AC-HEP-004 |
| REQ-HEP-005 | AC-HEP-005a, 005b |
| REQ-HEP-006 | AC-HEP-014 |
| REQ-HEP-007 | AC-HEP-006, 007a, 007b |
| REQ-HEP-008 | AC-HEP-008 |
| REQ-HEP-009 | AC-HEP-009 |
| REQ-HEP-010 | AC-HEP-010 |
| REQ-HEP-011 | AC-HEP-011 |
| REQ-HEP-012 | AC-HEP-012 |
| REQ-HEP-013 | AC-HEP-013a, 013b |

---

## Edge Cases

- **EC-1**: promotions 파일 부재/빈 파일 → `MapPromotions` 빈(non-nil) slice, 무오류 (기존 tolerance 보존)
- **EC-2**: malformed JSONL 라인 혼재 → reader가 해당 라인만 skip (기존 계약 보존 — 회귀 테스트)
- **EC-3**: confidence < 0.70 promotion → tier 유효해도 배제 (ConfidenceThreshold 계약 유지)
- **EC-4**: doctor 실행 대상 하네스의 manifest는 있으나 Runner 부재 (thin command → specialist 직결형, 예: github/release) → Runner 축 검사는 "Runner 선언 시에만" 적용 — false-positive 금지
- **EC-5**: classify가 훅 시간 예산(5s)에 근접 → 시간 초과 시에도 훅 exit 0 (fail-open); 예산 실측치를 progress.md에 기록
- **EC-6**: Windows 렌더 경로 — settings.json.tmpl observer 등록이 기존 크로스 플랫폼 패턴(bash prefix 이중 항목)을 따름

---

## Quality Gates

- `go test ./...` full PASS + `go test -race` touched pkg PASS
- Coverage: `internal/harness/proposalgen` ≥ 85%, `internal/cli`(hook/harness 경로) 90% 목표 (critical pkg 기준)
- `golangci-lint run` clean
- `make build` 성공 (템플릿 임베드 재컴파일)
- `TestSplitHarnessNamespaceNoLeak` PASS + 중립성 grep 0 matches
- 커밋: Conventional Commits(한국어), pathspec 제한, `--no-verify` 금지

---

## Definition of Done

1. AC-HEP-001a ~ 013b 전건 PASS — 각 AC에 verbatim 명령 출력 증거 (progress.md §E.2)
2. 실데이터 스모크: 실제 `.moai/harness/learning-history/tier-promotions.jsonl` 대상 proposal 생성 경로 실행 → 후보 ≥ 1 관측 (구조적 0의 해소 실증)
3. exemplar Runner: 수리 전 2 findings → 수리 후 0 findings 시퀀스가 progress.md에 기록
4. 8개 문서/설정 표면 live↔template parity + 중립성 확인
5. spec.md frontmatter `status` 전이는 소유 매트릭스 준수 (draft→in-progress: manager-develop, →implemented→completed: manager-docs)
