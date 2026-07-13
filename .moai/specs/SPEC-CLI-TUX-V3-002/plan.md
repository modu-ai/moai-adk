# SPEC-CLI-TUX-V3-002 — Implementation Plan

## §A Context

CLI TUX 현대화 4-마일스톤 중 M2(`moai init` 재설계). Tier M — I-1~I-5 5개 개선 + huh v2 스파이크. 방법론: `quality.yaml development_mode: tdd`, 단 표면별 분화:

- **스파이크(M2a)**: 실험 — 코드 산출물 없이 verdict 문서만 남긴다(progress.md §E.2).
- **셀프업데이트 순서 교정 + 경고 수집기(M2b/M2d)**: TDD RED-GREEN-REFACTOR — 채널·순서 계약을 테스트가 먼저 정의.
- **위저드 개편(M2c)**: characterization-first — 기존 위저드 테스트 + `WizardResult` 스키마가 회귀 앵커.

산출물 SSOT: spec.md(18 REQ) / acceptance.md(18 AC). 원본 계획: `.moai/reports/moai-cli-tux-modernization-plan-20260710.html` §4 M2 + §5 I-1~I-5.

§F 마일스톤 순서는 결정-가역성(decision-reversibility) 기준이다: 변경 가능성이 높은 결정(스파이크 verdict → 위저드 UX 구조 → 셀프업데이트 UX)을 앞에, 기계적 이행(수집기 배선·ratchet 감소·회귀 매트릭스)을 뒤에 배치해 사람 리뷰가 고변동 결정에 집중되게 한다.

## §B Known Issues (착수 전 인지 사항)

| # | 앵커 (run-phase 재실측) | 이슈 | 대응 |
|---|---|---|---|
| 1 | go.mod: bubbletea v1.3.10 / huh v1.0.0 / bubbles v1.0.0(indirect) | **M1 잔여 부채** — v2 미도입. 라이브 스피너(bubbles v2)와 통합 폼(huh v2)의 전제 | M2a 스파이크에서 v2 도입 가능성 확정 → REQ-TUX2-012 |
| 2 | wizard.go:99 `wizardTotalSteps = 6` + wizard.go:76 `TotalVisibleQuestions` | 스텝퍼 분모 이중 경로(하드코딩 6 vs 동적) — 실제 표시 질문 7~9 | 동적 단일 소스로 통일(REQ-TUX2-008); 상수 제거 grep AC |
| 3 | init.go:232 `runInit` 선두 `runBinaryUpdateStep` | 위저드 전 블로킹 네트워크 | 위저드 후 지연 + 비차단 알림(REQ-TUX2-001/002); re-exec 경로는 위저드 응답 수집 후 금지 |
| 4 | update.go:453 `shouldSkipBinaryUpdate` (templates-only/env/dev-build) | 스킵 시맨틱 3종은 behavior-preserving 의무 | characterization 테스트 선행(REQ-TUX2-003) |
| 5 | printer.go:241 spinnerFrame 단일 프레임 stateless | M1 printer는 애니메이션 없음 — Spinner/Progress 핸들 구현 교체 필요 | bubbles v2 spinner 백엔드로 핸들 내부만 교체 — Printer 인터페이스 계약(REQ-CTX-011~014) 불변 |
| 6 | init.go:509-510 `result.Warnings` uikit 나열 | 경고 집계점이 위저드 result에만 존재 — run 전체 경고 수집기 부재 | Printer `Warn` 경유 경고를 수집기로 계측(REQ-TUX2-013) |
| 7 | huh v2의 bubbletea v2 강제 | huh v2 도입 시 bubbletea v1 잔존 소비자(internal/merge/confirm.go)와 메이저 혼재 | confirm.go는 M3 소관 — v1/v2 공존은 M1 REQ-CTX-002 전례대로 허용; go.mod tidy diff 커밋별 확인 |
| 8 | reexecNewBinary(update.go:536) | 셀프업데이트 성공 시 프로세스 교체 — 위저드 후로 미루면 응답 유실 위험 | 위저드 후에는 re-exec 금지, 알림+힌트만(REQ-TUX2-002) |
| 9 | fang root(M1) 이후 help/에러 표면 변경 가능성 | init --help 골든이 M1 이후 상태 기준 | pre-flight에서 현행 골든 green 확인 후 착수 |
| 10 | 계획 보고서의 init.go:220 앵커 | M1 printer 배선으로 라인 드리프트 발생(현재 232) | 라인 번호는 전부 재실측 — content-token 앵커 우선 |
| 11 | wizard_test.go:559-560 `TestCharacterize_WizardTotalSteps` — `wizardTotalSteps == 6` 단언 (2026-07-14 실측) | REQ-TUX2-008이 상수를 제거하면 본 characterization 테스트가 제거 대상 식별자를 직접 참조하므로 **컴파일 실패** | **명시 supersession 승인**: 해당 테스트를 동적 분모(`TotalVisibleQuestions` 기반) 단언 테스트로 교체(또는 삭제) 허용 — REQ-TUX2-009의 escape("골든 변경 시 사유 기재")를 본 characterization supersession까지 확장; 교체/삭제 사유를 progress.md §E.2에 기재 |

## §C Pre-flight (착수 전 의무 검증)

```bash
# 1. baseline 확인
git branch --show-current && git rev-parse HEAD

# 2. cross-platform build baseline
go build ./... && GOOS=windows GOARCH=amd64 go build ./...

# 3. lint baseline (NEW vs pre-existing 구분)
golangci-lint run --timeout=5m 2>&1 | tail -5

# 4. init/wizard/printer characterization baseline — 전량 green 확인
go test ./internal/cli/... ./internal/cli/wizard/... ./internal/cli/printer/... -count=1

# 5. ratchet baseline 재실측 (2026-07-13 실측 40과 대조)
grep -rn 'fmt\.Printf\|fmt\.Println\|fmt\.Print(' internal/cli --include='*.go' | grep -v '_test.go' | wc -l

# 6. huh v2 / bubbletea v2 / bubbles v2 최신 버전 확인
go list -m -versions charm.land/bubbletea/v2 | tr ' ' '\n' | tail -1
go list -m -versions charm.land/bubbles/v2 | tr ' ' '\n' | tail -1
go list -m -versions github.com/charmbracelet/huh | tr ' ' '\n' | tail -3   # v2 계열 존재 여부 + import 경로(charm.land 이동 여부) 확정

# 7. 앵커 재실측 — runBinaryUpdateStep 호출 위치 / wizardTotalSteps / Warnings 집계점
grep -n 'runBinaryUpdateStep' internal/cli/init.go
grep -n 'wizardTotalSteps' internal/cli/wizard/*.go
grep -n 'result.Warnings' internal/cli/init.go
```

## §D Constraints (DO NOT VIOLATE)

- **PRESERVE (절대 무접촉)**: `internal/statusline/**`, `internal/merge/confirm.go`(M3 소관 — @MX:DEBT 태그 유지), `internal/cli/update.go`의 템플릿 sync 파이프라인(M3 소관; `shouldSkipBinaryUpdate`/`runBinaryUpdateStep`의 **호출 위치 이동**은 허용, 함수 시맨틱 변경 금지), `internal/cli/uikit/banner.go`(M4 소관), 무관 SPEC 디렉터리, `.moai/state/**`.
- **Printer 공개 계약 보존**: `internal/cli/printer` Printer 인터페이스 메서드 집합·채널 규율·3모드 시그니처(REQ-CTX-011~014) 변경 금지 — 핸들 내부 구현만 교체.
- **WizardResult 스키마 보존**: 위저드 개편은 표시 계층만 — 질문 집합/기본값/검증/결과 필드 무변경.
- **스파이크 게이트**: M2a verdict가 progress.md §E.2에 기록되기 전 M2c 커밋 금지.
- **Template-First 비적용**: 내부 Go 코드만 — `internal/template/templates/**` 무접촉.
- **금지 명령**: `--no-verify`, `--amend`, force-push. Conventional Commits (`feat(SPEC-CLI-TUX-V3-002): M2a ...`) + `🗿 MoAI` trailer 의무.

## §E Self-Verification (완료 보고 의무 항목)

- E1: acceptance.md 18 AC 전 행 PASS/FAIL 매트릭스 + verbatim 출력 (vci 5-section 형식).
- E2: `go build ./... && GOOS=windows GOARCH=amd64 go build ./...` verbatim.
- E3: `go test ./internal/cli/... -count=1` green + init/wizard 신규 테스트 커버 경로 명시.
- E4: subagent boundary grep — `grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/wizard/ | grep -v _test` 0건.
- E5: `golangci-lint run --timeout=5m` NEW 항목 0 (baseline 대조).
- E6: ratchet 재실측 — pre-flight baseline 미만 + 실제 감소치 보고.
- E7: 스파이크 verdict + (해당 시) plan-B 결정 기록 위치 인용(progress.md §E.2).

## §F Milestones (priority order — 결정-가역성 우선)

### M2a — huh v2 호환성 스파이크 (Priority: High, 결정 게이트)

1. Pre-flight #6로 huh v2 계열 존재·import 경로 확정 (charm.land 이동 여부 포함).
2. `/tmp` 격리 스파이크 프로젝트에서 multi-group 폼 + 7~9 필드 렌더 → YOffset 스크롤 결함 재현 시도.
3. **Verdict 기록**: progress.md §E.2에 성공(→ M2c 통합 폼) / 실패(→ plan B: 현행 구조 + v2 스타일만) + 재현 증거. (REQ-TUX2-005)
4. bubbletea/bubbles v2 도입 커밋 (`go get charm.land/{bubbletea,bubbles}/v2@<pinned>`) — verdict와 무관하게 I-3 전제 (REQ-TUX2-012).
5. **게이트**: verdict 문서화 + `go build ./...` green.

### M2b — 셀프업데이트 순서 교정 (Priority: High, I-1 / UX 결정)

1. RED: "위저드 첫 인터랙션 전 네트워크 0회" 계약 테스트 — 네트워크 호출을 계측 가능한 seam(주입 가능한 update-check 함수)으로 추출 후 순서 단언.
2. GREEN: `runBinaryUpdateStep` 호출을 위저드 완료 후로 이동; 신버전 감지 시 stderr 비차단 알림 + `moai update` 힌트 (re-exec 금지). (REQ-TUX2-001/002)
3. characterization: `shouldSkipBinaryUpdate` 3종 스킵 시맨틱 보존 테스트. (REQ-TUX2-003)
4. non-interactive/non-TTY 경로의 비차단 속성 테스트. (REQ-TUX2-004)
5. **게이트**: AC-TUX2-001~004 통과.

### M2c — 위저드 개편 (Priority: High, I-2 / UX 구조 결정 — M2a verdict 종속)

1. characterization: 기존 위저드 테스트 green 앵커 + `WizardResult` 스키마 스냅숏.
2. verdict=성공: 단일 멀티그룹 폼(그룹 2-3개 + 스텝퍼) 구현; verdict=실패: 현행 구조 유지 + v2 스타일 적용만. (REQ-TUX2-006/007)
3. 스텝퍼 분모 동적화 — `wizardTotalSteps` 상수 제거, `TotalVisibleQuestions` 단일 소스. (REQ-TUX2-008)
4. locale/standard/advanced 모드 시맨틱 보존 검증. (REQ-TUX2-009)
5. **게이트**: AC-TUX2-005~009 통과 + 위저드 스위트 green.

### M2d — 라이브 진행 + 경고 수집기 + 완료 카드 (Priority: Medium, I-3/I-4 + 카드)

1. RED: Spinner/Progress 핸들의 TTY 애니메이션 + non-TTY/NO_COLOR/MOAI_REDUCED_MOTION 폴백 계약 테스트.
2. GREEN: printer 핸들 내부를 bubbles v2 spinner 백엔드로 교체 — 인터페이스 불변. PhaseExecutor 진행을 파일 카운트 연동. (REQ-TUX2-010/011)
3. RED→GREEN: 경고 수집기 — run 전체 `Warn` 계측 + 종료 시 stderr 요약 패널 1회. (REQ-TUX2-013/014)
4. 완료 카드에 다음 액션 시퀀스 + 경고 요약 포인터. (REQ-TUX2-016)
5. **게이트**: AC-TUX2-010~014, 016 통과.

### M2e — 리다이렉트 힌트 + 회귀 매트릭스 (Priority: Low, I-5 + 게이트, 기계적)

1. 기존 프로젝트 감지 시 `moai update` 리다이렉트 힌트. (REQ-TUX2-015)
2. ratchet 재실측 — init/wizard 잔여 fmt.Print* 흡수분 반영. (REQ-TUX2-018)
3. 회귀 매트릭스: non-TTY(`| cat`) × `NO_COLOR=1` × `GOOS=windows` 빌드 — AC-TUX2-015, 017~018.
4. **게이트**: 전체 스위트 green + lint NEW 0 + §E 자가 검증 전 항목.

### 검증 명령 배치 (M2e 종료 시, 단일 턴 병렬 실행)

```bash
go test ./... -count=1 > /tmp/moai-verify/1-full.log 2>&1; echo "exit=$?"; tail -20 /tmp/moai-verify/1-full.log
GOOS=windows GOARCH=amd64 go build ./... > /tmp/moai-verify/2-win.log 2>&1; echo "exit=$?"
golangci-lint run --timeout=5m > /tmp/moai-verify/3-lint.log 2>&1; echo "exit=$?"; tail -10 /tmp/moai-verify/3-lint.log
grep -rn 'fmt\.Printf\|fmt\.Println\|fmt\.Print(' internal/cli --include='*.go' | grep -v '_test.go' | wc -l
grep -rn 'wizardTotalSteps' internal/cli/wizard --include='*.go' | grep -v '_test.go' | wc -l; echo "expect 0"
go list -m charm.land/bubbletea/v2 charm.land/bubbles/v2; echo "expect resolve"
NO_COLOR=1 go run ./cmd/moai init --help 2>&1 | grep -c $'\x1b'; echo "expect 0"
```

## §G Anti-Patterns and Risks

- **Anti-pattern: 스파이크 없이 통합 폼 직행** — 스크롤 버그 잔존 시 되돌림 비용이 M2c 전체. verdict 게이트 의무(§D).
- **Anti-pattern: Printer 인터페이스에 bubbles v2 타입 노출** — 핸들 내부만 교체. v2 타입이 공개 시그니처로 새면 M1의 문자열-토큰 경계 설계가 무너진다.
- **Anti-pattern: 위저드 개편과 셀프업데이트 이동을 한 커밋에 혼합** — 회귀 원인 분리 불가. M2a~M2e 커밋 분리.
- **Anti-pattern: re-exec를 위저드 후로 단순 이동** — 위저드 응답이 프로세스 교체로 유실. 위저드 후에는 알림만(REQ-TUX2-002).
- **Risk: huh v2 import 경로 미확정** — charm.land 이동 여부를 pre-flight #6에서 확정; 스파이크 전 추측 import 금지.
- **Risk: bubbles v2 spinner의 고루틴 수명** — 테스트에서 goroutine leak 검출(`-race`) 필수; 핸들 Complete/Fail 후 프레임 방출 금지.
- **Risk: 셀프업데이트 지연으로 구버전 템플릿 배포** — 알림에 "신버전 존재, `moai update`로 템플릿 갱신" 문구를 포함해 사용자 인지 보장; 자동 재실행은 하지 않는다(명시적 트레이드오프).

## §H Cross-References

- 원본 계획: `.moai/reports/moai-cli-tux-modernization-plan-20260710.html` §4 M2 + §5 I-1~I-5 + §6 리스크 + §7 성공 지표.
- 선행: SPEC-CLI-TUX-V3-001 (completed — printer/fang/tui v2). 후속: -003 (M3 update), -004 (M4 폴리시).
- 계약 승계: REQ-CTX-011~014 (Printer 계약), REQ-CTX-017 (ratchet), AC-CLI-TUI-013 (hex 금지).
- 규약: `internal/cli/CLAUDE.md` (output streams), CLAUDE.local.md §6 (테스트 격리 — 스파이크는 /tmp) · §14 (하드코딩 금지).
- 방법론: moai-workflow-tdd (M2b/M2d) + characterization (M2c).
