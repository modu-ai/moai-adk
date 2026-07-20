# SPEC-CLI-TUX-V3-002 — Acceptance Criteria

## §A Scenarios (Given-When-Then)

1. **Given** 네트워크가 차단된 환경에서 사용자가 대화식 `moai init myproj`를 실행할 때, **When** 위저드 첫 폼이 렌더되면, **Then** 그 시점까지 어떤 네트워크 호출도 시도되지 않았고(계측 seam 단언) 위저드는 정상 진행된다 — 셀프업데이트는 위저드 완료 후 비차단으로 시도되어 실패해도 init 결과에 영향이 없다.
2. **Given** init 실행 중 경고가 3건 발생하는 상황에서, **When** init이 종료되면, **Then** stderr에 경고 3건을 담은 요약 패널이 정확히 1회 출력되고 stdout에는 "Warning" 문자열이 한 번도 나타나지 않는다.
3. **Given** 이미 `.moai/`가 초기화된 디렉터리에서, **When** 사용자가 `--force` 없이 `moai init`을 재실행하면, **Then** 에러 안내에 `moai update` 리다이렉트 힌트가 포함된다.
4. **Given** CI 환경(non-TTY, `NO_COLOR=1`)에서, **When** 템플릿 배포가 진행되면, **Then** 스피너 애니메이션 프레임과 ANSI 이스케이프가 전혀 출력되지 않고 plain 진행 라인만 나타난다.

## §D AC Matrix (machine-verifiable)

| AC | REQ | Verification command | Expected outcome |
|---|---|---|---|
| AC-TUX2-001 | REQ-TUX2-001 | `go test ./internal/cli/ -run 'InitNetworkOrder\|NoNetworkBeforeWizard' -count=1 -v` | PASS — 주입된 update-check seam이 위저드 완료 이벤트 이후에만 호출됨을 순서 단언 (신규 테스트 실존을 `grep -rn 'func TestInitNoNetworkBeforeWizard' internal/cli --include='*_test.go'` ≥ 1 match로 병행 증명) |
| AC-TUX2-002 | REQ-TUX2-002 | `go test ./internal/cli/ -run 'DeferredUpdateNotice' -count=1 -v` | PASS — 신버전 감지 시 stderr 캡처에 알림 + `moai update` 힌트 존재, stdout 캡처에 부재, re-exec 함수 미호출 단언 |
| AC-TUX2-003 | REQ-TUX2-003 | `go test ./internal/cli/ -run 'SkipBinaryUpdate' -count=1 -v` | PASS — templates-only 플래그 / env guard / dev-build 3종 스킵 조건에서 지연 체크도 실행되지 않음 (characterization) |
| AC-TUX2-004 | REQ-TUX2-004 | `go test ./internal/cli/ -run 'InitNonInteractive.*Update\|NonTTYUpdateCheck' -count=1 -v` | PASS — non-TTY/플래그 완결 경로에서 update check가 phase 실행을 블로킹하지 않음 단언 |
| AC-TUX2-005 | REQ-TUX2-005 | progress.md §E.2 검사: `grep -n 'huh v2 spike verdict' .moai/specs/SPEC-CLI-TUX-V3-002/progress.md` | ≥ 1 match — verdict(성공/plan B) + 재현 증거 기재; M2c 첫 커밋보다 앞선 기록(커밋 타임라인 대조) |
| AC-TUX2-006 | REQ-TUX2-006, REQ-TUX2-007 | verdict=성공: `go test ./internal/cli/wizard/ -run 'UnifiedForm\|MultiGroup' -count=1 -v`; verdict=plan B: `go test ./internal/cli/wizard/ -count=1` + progress.md §E.2 plan-B 결정 기록 | 선택된 경로의 테스트 PASS + 결정 근거 문서화 (양 경로 중 실제 취한 쪽만 평가 — 미선택 경로 AC는 N/A로 명기) |
| AC-TUX2-007 | REQ-TUX2-008 | `grep -rn 'wizardTotalSteps' internal/cli/wizard --include='*.go' \| grep -v '_test.go' \| wc -l` + `go test ./internal/cli/wizard/ -run 'StepperTotal\|VisibleQuestions' -count=1 -v` | grep 0 + PASS — 스텝퍼 분모가 표시 질문 수에 따라 7~9로 동적 변화함을 단언 |
| AC-TUX2-008 | REQ-TUX2-009 | `go test ./internal/cli/wizard/ -count=1` | PASS — locale/standard/advanced 기존 스위트 green (골든 변경 시 파일별 사유 progress.md §E.2 기재) |
| AC-TUX2-009 | REQ-TUX2-012 | `go list -m charm.land/bubbletea/v2 charm.land/bubbles/v2` + `grep -rn 'charm.land/bubbletea/v2\|charm.land/bubbles/v2' internal/cli --include='*.go' \| grep -v '_test.go'` (보조: `go mod why charm.land/bubbletea/v2`) | 2개 모듈 resolve + 비테스트 internal/cli 소스의 import·사용 ≥ 1 (M1 잔여 부채 해소 — go.mod direct 블록 존재가 아닌 import-사용 증거로 앵커: bubbles/v2만 import될 경우 `go mod tidy`가 bubbletea/v2를 indirect로 강등할 수 있어 블록 검사만으로는 취약) |
| AC-TUX2-010 | REQ-TUX2-010 | `go test ./internal/cli/printer/ -run 'SpinnerAnimated\|ProgressLive' -count=1 -v` + `grep -rn 'charm.land/bubbles/v2' internal/cli/printer --include='*.go' \| grep -v '_test.go'` | PASS — TTY 모드 핸들이 다중 프레임 방출 + 파일 카운트 갱신 단언; grep ≥ 1 (bubbles v2 백엔드 도달 가능성 — 토큰 존재가 아닌 import+사용) |
| AC-TUX2-011 | REQ-TUX2-011 | `go test ./internal/cli/printer/ -run 'ReducedMotion\|NoANSI\|PlainFallback' -count=1 -v` | PASS — non-TTY/NO_COLOR/MOAI_REDUCED_MOTION에서 `\x1b[` 부재 + 단일 프레임 정적 출력 단언 |
| AC-TUX2-012 | REQ-TUX2-013 | `go test ./internal/cli/ -run 'WarningCollector\|WarningSummary' -count=1 -v` | PASS — 경고 N건 주입 시 stderr 요약 패널 1회(N건 전부 포함) + 개별 경고 재출력이 요약과 중복되지 않음 단언 |
| AC-TUX2-013 | REQ-TUX2-014 | `go test ./internal/cli/ -run 'InitStdoutClean\|WarningChannel' -count=1 -v` | PASS — init 전 경로 stdout 캡처에 "Warning" 부재 (골든) |
| AC-TUX2-014 | REQ-TUX2-016 | `go test ./internal/cli/ -run 'CompletionCard\|NextActions' -count=1 -v` | PASS — 성공 완료 카드에 `moai cc` + `/moai plan` 다음 액션 문자열 존재 단언 |
| AC-TUX2-015 | REQ-TUX2-015 | `go test ./internal/cli/ -run 'ExistingProject.*Hint\|UpdateRedirect' -count=1 -v` | PASS — `.moai/` 기존재 + `--force` 부재 시 에러 출력에 `moai update` 문자열 존재 |
| AC-TUX2-016 | REQ-TUX2-017 | `go test ./... -count=1 && golangci-lint run --timeout=5m && GOOS=windows GOARCH=amd64 go build ./... && GOOS=linux GOARCH=amd64 go build ./...` | 전체 green + lint NEW 0 (pre-flight baseline 대조) + 3-타깃 빌드 exit 0 |
| AC-TUX2-017 | REQ-TUX2-018 | `grep -rn 'fmt\.Printf\|fmt\.Println\|fmt\.Print(' internal/cli --include='*.go' \| grep -v '_test.go' \| wc -l` | pre-flight baseline(재실측; 2026-07-13 기준 40) **미만** — 실측치 progress.md §E.2 기록 |
| AC-TUX2-018 | REQ-TUX2-001, REQ-TUX2-011 | `NO_COLOR=1 go run ./cmd/moai init --help 2>&1 \| grep -c $'\x1b'` | 0 (grep exit 1 = no matches) — 파이프 + NO_COLOR에서 ANSI 무출력 |

## §C Edge Cases

- **위저드 중단(Ctrl+C)**: 위저드 취소 시 지연 셀프업데이트 체크도 실행하지 않는다 — 취소 경로에 네트워크 부작용 금지.
- **경고 0건**: 수집기가 빈 요약 패널을 출력하지 않는다 — 경고 없으면 패널 자체 생략.
- **NO_COLOR 값 시맨틱**: 비어있지 않은 임의 문자열 = set (`NO_COLOR=0` 포함) — 기존 detect.go isEnvSet 시맨틱 승계.
- **스피너 고루틴 수명**: Complete/Fail 후 프레임 방출 금지; `go test -race` 필수 통과.
- **혼합 TTY(stdout 파이프 + stderr TTY)**: Printer 모드는 구성 시점 1회 고정(M1 결정 승계) — 두 채널 모두 동일 모드.
- **dev 빌드에서의 지연 체크**: `shouldSkipBinaryUpdate`의 dev-build 감지로 지연 체크도 스킵 — 로컬 개발 루프에 네트워크 소음 금지.
- **plan-B 경로의 스텝퍼**: 통합 폼 실패(plan B) 시에도 스텝퍼 동적 분모(REQ-TUX2-008)는 독립적으로 적용 — 스파이크 verdict에 종속되지 않는다.

## §D.5 Quality Gate / Definition of Done

- 18 AC 전 행 PASS(또는 미선택 스파이크 경로 N/A 명기) — verbatim 명령 출력을 progress.md §E.2에 인용 (vci §3 5-section 형식).
- M2a~M2e 마일스톤별 커밋 분리 + 스파이크 verdict가 M2c 커밋보다 선행함을 커밋 타임라인으로 증명.
- ratchet 실측치 baseline 미만 + `wizardTotalSteps` 상수 제거 grep 0.
- 채널 규율 골든(stdout 무경고) + non-TTY/NO_COLOR/REDUCED_MOTION 폴백 전부 green.
- frontmatter `status` 전이는 소유 매트릭스 준수 (draft→in-progress는 manager-develop).
