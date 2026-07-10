# SPEC-WEB-CONSOLE-012 — Acceptance Criteria

> AC SSOT. 모든 AC는 기계 검증 명령 + 기대 출력을 명시한다. `-run` 무매치 exit-0 함정 회피를 위해 full-package `go test`를 1차 판정으로 삼는다.

## §D.1 AC Matrix

| AC | REQ | 검증 명령 | 기대 |
|----|-----|-----------|------|
| AC-WC12-001 | 001/002 | `sed -n '/^func llmFields/,/^}/p' internal/settings/schema_sections.go \| grep -c 'opus\|sonnet\|haiku'` | `0` |
| AC-WC12-002 | 001 | `sed -n '/^func llmFields/,/^}/p' internal/settings/schema_sections.go \| grep -c 'fable'` | `>= 1` |
| AC-WC12-003 | 003 | `sed -n '/^func applyLLMKey/,/^}/p' internal/settings/sectionapply.go \| grep -c 'glm.models.fable'` | `1` |
| AC-WC12-004 | 003 | `sed -n '/^func applyLLMKey/,/^}/p' internal/settings/sectionapply.go \| grep -c 'glm.models.opus\|glm.models.sonnet\|glm.models.haiku'` | `0` |
| AC-WC12-005 | 004 | `test $(grep -c 'f\.llm\.glm\.models\.fable\.title' internal/web/assets/i18n.js) -eq $(grep -c 'f\.llm\.glm\.models\.high\.title' internal/web/assets/i18n.js) && echo PARITY` | `PARITY` (sibling 파생 — locale 수 hard-pin 금지) |
| AC-WC12-006 | 004 | `grep -c 'f\.llm\.glm\.models\.opus\|f\.llm\.glm\.models\.sonnet\|f\.llm\.glm\.models\.haiku' internal/web/assets/i18n.js` | `0` |
| AC-WC12-007 | 005 | `grep -c 'opus/sonnet/haiku' internal/web/schemaform.go` | `0` |
| AC-WC12-008 | 010 | `grep -c '"research"' internal/settings/sectionroute.go internal/settings/sectionwrite.go` | 각 `0` |
| AC-WC12-009 | 011 | `grep -c 'SectionResearch' internal/settings/schema.go && grep -rc 'SectionResearch' internal/settings/ internal/web/ --include='*.go'` | 전부 `0` |
| AC-WC12-010 | 012 | research seam 거부 회귀 테스트 (신설 또는 기존 확장; `WriteSectionViaSeam(root, "research", …)` → error + research.yaml 무변경 단언) — `go test ./internal/settings/` | PASS |
| AC-WC12-011 | 020 | `grep -c 'git_convention.auto_detection' internal/settings/schema.go internal/web/assets/i18n.js` | 각 `0` |
| AC-WC12-012 | 021 | `grep -rc 'min_coverage_per_commit' internal/settings/*.go \| grep -v ':0'` 존재 AND `grep -rc 'validation.enforce_on_push' internal/settings/*.go \| grep -v ':0'` 존재 | USED 필드 FieldDef 잔류 |
| AC-WC12-013 | 022 | `grep -c 'AutoDetection AutoDetectionConfig' internal/config/types.go; grep -c 'Opus   string' internal/config/types.go` | 각 `>= 1` (struct 보존) |
| AC-WC12-014 | 023 | `grep -c '"auto_detection"' internal/settings/schema_sections.go` | `>= 1` (harness 필드 무접촉) |
| AC-WC12-015 | 030/031 | Where-gate 통과 시: `grep -rc 'errDictKey' internal/web/ \| grep -v ':0'` → 무출력. fallback 시: assets.go 주석에 발견된 guard의 file:test-name 인용 존재 + 완료 보고 기재 | 분기별 판정 |
| AC-WC12-016 | 032 | `grep -rn 'WorkflowAgentPurposes' --include='*.go' internal/ cmd/ pkg/` | 무매치 |
| AC-WC12-017 | 040 | `grep -c 'research' internal/web/server.go` 결과가 "제외군 열거 문맥"만 매치 (편집 가능 목록에서 research/db 부재 — 수동 확인 1회 + `grep -c '10 user-facing' internal/web/server.go` → `0`) | 충족 |
| AC-WC12-018 | 050 | `go test ./internal/settings/... ./internal/web/... ./internal/cli/...` + 명시 실행 `go test ./internal/cli/ -run TestI18nKeySetParity -v` (실존 함수명 바인딩) | 전부 PASS |
| AC-WC12-019 | 051 | `git diff --name-only <M1-base>..HEAD \| grep -c 'internal/statusline/\|internal/template/templates/'` | `0` |
| AC-WC12-020 | 전체 | `go build ./... && GOOS=windows GOARCH=amd64 go build ./... && go vet ./... && go test ./...` | 전부 exit 0 |

## §D.2 A5 Per-Field Evidence (plan-phase 2026-07-10 실측 — run-phase 재실행 의무)

verification-claim-integrity.md §1.1 surface 3 준수 기록. 각 행: 실행 명령 + verbatim 매치. run-phase는 제거 직전 동일 명령을 재실행해 분류가 유지됨을 확인한다 (stale-evidence 방지).

### quality.tdd_settings.min_coverage_per_commit — USED (잔류)

```
$ grep -rn 'MinCoveragePerCommit' --include='*.go' internal/ cmd/ pkg/ | grep -v '_test.go' | grep -v 'internal/web/' | grep -v 'internal/settings/'
internal/core/quality/trust.go:788:	minCov := g.config.TDDSettings.MinCoveragePerCommit
```
(그 외 매치는 defaults/validation/모델 선언 — 행동 reader는 trust.go:788.)

### git_convention.validation.enforce_on_push — USED (잔류)

```
$ grep -rn '\.EnforceOnPush\|EnvEnforceOnPush' --include='*.go' internal/ | grep -v '_test.go' | grep -v web | grep -v settings | grep -v profile_setup | grep -v envkeys
internal/cli/hook_pre_push.go:251:	if envVal := os.Getenv(config.EnvEnforceOnPush); envVal != "" {
internal/cli/hook_pre_push.go:258:			return cfg.GitConvention.Validation.EnforceOnPush
```
(선행 감사의 "dead 반증" 재확인 — pre-push hook 경로가 소비.)

### git_convention.auto_detection.{enabled, confidence_threshold, sample_size} — DEAD (제거)

```
$ grep -rn '\.AutoDetection\.' --include='*.go' internal/ cmd/ pkg/ | grep -v '_test.go' | grep -v 'internal/web/' | grep -v 'internal/settings/' | grep -v 'internal/config/'
(no output)
```
(존재하는 매치는 전부 config 인프라: validation.go 범위 검사, defaults.go 기본값, loader.go:332 키 인식 — 행동 reader 0. struct/yaml/validation은 REQ-WC12-022로 보존.)

## §D.3 Given-When-Then Scenarios

**S1 — fable 편집 왕복.** Given `moai web` 콘솔 LLM 섹션, When 사용자가 `glm.models.fable` 값을 수정 후 저장, Then `.moai/config/sections/llm.yaml`의 `glm.models.fable`이 갱신되고 기존 주석/미편집 키가 보존되며, 다음 `moai glm` 기동 시 `ANTHROPIC_DEFAULT_FABLE_MODEL`이 그 값으로 설정된다 (glm.go setGLMEnv 경로).

**S2 — research seam 거부.** Given 폐선 완료 상태, When `WriteSectionViaSeam(root, "research", edits)` 호출(또는 조작된 POST), Then not-seam-writable 오류 반환 + `research.yaml` 바이트 무변경.

**S3 — 양표면 파리티.** Given 스키마 필드셋 변경 완료, When `go test ./internal/web/... ./internal/cli/...` 실행, Then web-side i18n 파리티와 TUI-side bridge 파리티가 모두 PASS (011 M2b 회귀 부재).

## §D.4 Edge Cases

- **EC-1**: 저장 폼에서 fable 미제출/빈 값 → 기존 영속값 보존 (empty=preserve 계약 유지).
- **EC-2**: 라이브 llm.yaml에 legacy `opus/sonnet/haiku` 키가 값과 함께 잔존하는 상태에서 콘솔 저장 → typed re-marshal이 struct legacy 필드로 키를 보존, 데이터 파괴 없음 (roundtrip 테스트로 단언).
- **EC-3**: `harness.auto_detection.enabled`는 제거 후에도 스키마에 잔존·편집 가능 (AC-WC12-014).

## §D.5 Definition of Done

- [ ] AC-WC12-001..020 전부 PASS (E1 매트릭스에 verbatim 출력 인용)
- [ ] §D.2 3개 분류 명령 run-phase 재실행 결과가 plan-phase 분류와 일치
- [ ] `go test ./...` 전체 green + cross-platform build exit 0
- [ ] lint NEW issue 0 (baseline 별도 표기)
- [ ] 커밋 전부 specific-path add, `refactor(SPEC-WEB-CONSOLE-012): M{N} …` 형식, main push 완료
