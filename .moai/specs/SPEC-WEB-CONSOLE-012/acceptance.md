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
| AC-WC12-009 | 011 | `grep -c 'SectionResearch' internal/settings/schema.go; grep -rn 'SectionResearch' internal/settings/ internal/web/ --include='*.go' \| wc -l` (0.2.0 D4: `&&`는 첫 grep이 0매치 exit 1일 때 둘째 명령을 단락시켜 성공 케이스가 미검증 — `;` 분리) | 두 출력 모두 `0` |
| AC-WC12-010 | 012 | research seam 거부 회귀 테스트 (신설 또는 기존 확장; `WriteSectionViaSeam(root, "research", …)` → error + research.yaml 무변경 단언) — `go test ./internal/settings/` | PASS |
| AC-WC12-011 | 020 | (0.2.0 inverted — 잔류 단언) `grep -c 'git_convention.auto_detection.enabled' internal/settings/schema.go; grep -c 'git_convention.auto_detection.confidence_threshold' internal/settings/schema.go; grep -c 'git_convention.auto_detection.sample_size' internal/settings/schema.go; grep -c 'f\.git_convention\.auto_detection' internal/web/assets/i18n.js` | 4개 출력 전부 `>= 1` (FieldDef 3종 + i18n 키 잔류) |
| AC-WC12-012 | 021 | `grep -rc 'min_coverage_per_commit' internal/settings/*.go \| grep -v ':0'` 존재 AND `grep -rc 'validation.enforce_on_push' internal/settings/*.go \| grep -v ':0'` 존재 | USED 필드 FieldDef 잔류 |
| AC-WC12-013 | 006 | `grep -c 'AutoDetection AutoDetectionConfig' internal/config/types.go; grep -c 'Opus   string' internal/config/types.go; grep -c 'func resolveGLMModels' internal/cli/glm.go` | 각 `>= 1` (struct + fallback 체인 보존, REQ-WC12-006) |
| AC-WC12-014 | 023 | `grep -c '"auto_detection"' internal/settings/schema_sections.go` | `>= 1` (harness 필드 무접촉) |
| AC-WC12-015 | 030/031 | Where-gate 통과 시: `grep -rc 'errDictKey' internal/web/ \| grep -v ':0'` → 무출력. fallback 시: assets.go 주석에 발견된 guard의 file:test-name 인용 존재 + 완료 보고 기재 | 분기별 판정 |
| AC-WC12-016 | 032 | `grep -rn 'WorkflowAgentPurposes' --include='*.go' internal/ cmd/ pkg/` | 무매치 |
| AC-WC12-017 | 040 | `grep -c '10 user-facing' internal/web/server.go` → `0`; `grep -c '10개' internal/web/projectconfig.go` → `0` (0.2.0 D5 — 두 파일 모두 기계 검증); 보조 수동 확인 1회: server.go/projectconfig.go에서 research/db가 편집 가능 열거가 아닌 제외군 열거에만 등장 | 기계 2건 `0` + 수동 확인 기록 |
| AC-WC12-018 | 050 | `go test ./internal/settings/... ./internal/web/... ./internal/cli/...` + 명시 실행 `go test ./internal/cli/ -run TestI18nKeySetParity -v` (실존 함수명 바인딩) | 전부 PASS |
| AC-WC12-019 | 051 | `git diff --name-only <M1-base>..HEAD \| grep -c 'internal/statusline/\|internal/template/templates/'` | `0` |
| AC-WC12-020 | 전체 | `go build ./... && GOOS=windows GOARCH=amd64 go build ./... && go vet ./... && go test ./...` | 전부 exit 0 |

## §D.2 A5 Per-Field Evidence (plan-phase 2026-07-10 실측, 0.2.0 iter-2 정정 — run-phase 재실행 의무)

verification-claim-integrity.md §1.1 surface 3 준수 기록. 각 항목: 실행 명령 + verbatim 매치. run-phase는 M3에서 동일 명령을 재실행해 분류가 유지됨을 확인한다 (stale-evidence 방지).

**강화 프로토콜 (0.2.0 D8 — 의무).** 모든 dead 판정 grep은 **쌍**으로 수행한다: (1) field-dot 패턴 grep, (2) **bare-symbol grep** (타입/필드 심볼 단독, 예: `AutoDetection`). field-dot 패턴 단독은 whole-struct bind(`ad := cfg.GitConvention.AutoDetection` 후 `ad.Enabled` 로컬 접근)와 미러 struct 복사를 구조적으로 매치하지 못한다 — iter-1이 정확히 이 결함으로 live 3필드를 DEAD 오분류했다 (plan-audit iter-1 D1, CRITICAL). bare-symbol grep의 추가 매치는 **전건 설명**되어야 분류가 확정된다.

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

### git_convention.auto_detection.{enabled, confidence_threshold, sample_size} — USED (잔류; 0.2.0 재분류)

iter-1 오분류 기록 (교훈 보존 — 이 grep은 불충분했다):

```
$ grep -rn '\.AutoDetection\.' --include='*.go' internal/ cmd/ pkg/ | grep -v '_test.go' | grep -v 'internal/web/' | grep -v 'internal/settings/' | grep -v 'internal/config/'
(no output)   ← trailing-dot 패턴은 whole-struct bind를 매치 불가 (false negative)
```

iter-2 bare-symbol 재실측 (분류 확정 근거):

```
$ grep -rn 'AutoDetection' --include='*.go' internal/cli/hook_pre_push.go internal/git/convention/
internal/cli/hook_pre_push.go:223: func resolveAutoDetectOptions() (convention.AutoDetectOptions, int) {
internal/cli/hook_pre_push.go:237:   ad := cfg.GitConvention.AutoDetection
internal/git/convention/manager.go:46: func (m *Manager) LoadConvention(name string, opts AutoDetectOptions) error {
```

live 소비 체인: `hook_pre_push.go:146` `opts, maxLength := resolveAutoDetectOptions()` → `:237` whole-struct bind → `ad.Enabled`/`ad.SampleSize`/`ad.ConfidenceThreshold`/`ad.Fallback` 4필드 읽기 → `mgr.LoadConvention(convName, opts)` — detection gate/표본 수/수용 임계값으로 동작 결정. 배선 출처: SPEC-WEB-CONSOLE-009 (completed). → 3필드 전부 **USED**, 노출 잔류 (REQ-WC12-020).

## §D.3 Given-When-Then Scenarios

**S1 — fable 편집 왕복.** Given `moai web` 콘솔 LLM 섹션, When 사용자가 `glm.models.fable` 값을 수정 후 저장, Then `.moai/config/sections/llm.yaml`의 `glm.models.fable`이 갱신되고 기존 주석/미편집 키가 보존되며, 다음 `moai glm` 기동 시 `ANTHROPIC_DEFAULT_FABLE_MODEL`이 그 값으로 설정된다 (glm.go setGLMEnv 경로).

**S2 — research seam 거부.** Given 폐선 완료 상태, When `WriteSectionViaSeam(root, "research", edits)` 호출(또는 조작된 POST), Then not-seam-writable 오류 반환 + `research.yaml` 바이트 무변경.

**S3 — 양표면 파리티.** Given 스키마 필드셋 변경 완료, When `go test ./internal/web/... ./internal/cli/...` 실행, Then web-side i18n 파리티와 TUI-side bridge 파리티가 모두 PASS (011 M2b 회귀 부재).

## §D.4 Edge Cases

- **EC-1**: 저장 폼에서 fable 미제출/빈 값 → 기존 영속값 보존 (empty=preserve 계약 유지).
- **EC-2**: 라이브 llm.yaml에 legacy `opus/sonnet/haiku` 키가 값과 함께 잔존하는 상태에서 콘솔 저장 → typed re-marshal이 struct legacy 필드로 키를 보존, 데이터 파괴 없음 (roundtrip 테스트로 단언).
- **EC-3**: `harness.auto_detection.enabled`와 `git_convention.auto_detection.*` 전부 본 SPEC 이후에도 스키마에 잔존·편집 가능 — auto_detection 계열은 전면 무접촉 (AC-WC12-011/014).

## §D.5 Definition of Done

- [ ] AC-WC12-001..020 전부 PASS (E1 매트릭스에 verbatim 출력 인용)
- [ ] §D.2 분류 명령(강화 프로토콜: field-dot + bare-symbol 쌍) run-phase 재실행 결과가 0.2.0 분류(5필드 전원 USED)와 일치
- [ ] `go test ./...` 전체 green + cross-platform build exit 0
- [ ] lint NEW issue 0 (baseline 별도 표기)
- [ ] 커밋 전부 specific-path add, `refactor(SPEC-WEB-CONSOLE-012): M{N} …` 형식, main push 완료
