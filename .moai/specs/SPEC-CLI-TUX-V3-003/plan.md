# SPEC-CLI-TUX-V3-003 — Implementation Plan

## §A Context

CLI TUX 현대화 4-마일스톤 중 M3(`moai update` 재설계). Tier L — 122KB/3,276 LOC 단일 파일의 behavior-preserving 분해 + Bubble Tea v2 프리뷰 TUI + confirm.go v2 승격이 걸린 대규모 작업이며, 최악 리스크는 **namespace 보호 회귀 = 사용자 자산 손실**(계획 §6)이다. 방법론 분화:

- **분해(M3d)**: DDD ANALYZE-PRESERVE-IMPROVE — characterization 테스트(M3b)가 선행 안전망.
- **프리뷰 TUI + outcome 카드(M3c/M3e)**: TDD RED-GREEN-REFACTOR — 신규 표시 계층은 계약 테스트가 먼저.
- **confirm.go 승격(M3e)**: characterization 보존 + v2 재구현.

산출물 SSOT: spec.md(20 REQ) / acceptance.md(20 AC). 원본 계획: `.moai/reports/moai-cli-tux-modernization-plan-20260710.html` §4 M3 + §5 U-1~U-4.

> **Tier L 산출물 노트**: Tier L 표준 5-artifact(design.md/research.md 포함) 대비 본 plan-phase는 오케스트레이터 지시에 따라 3-artifact(+progress 스켈레톤)로 저작되었다. design/research 수준의 실측·경계 결정은 본 plan.md §B(Known Issues)와 spec.md §A에 압축 수록 — plan-audit에서 보강 필요 판정 시 orchestrator가 재위임한다.

§F 마일스톤 순서는 결정-가역성(decision-reversibility) 기준: 변경 가능성이 높은 결정(분류 데이터 모델·패키지 경계·프리뷰 UX)을 앞에, 기계적 이행(파일 이동·ratchet·회귀 매트릭스)을 뒤에 배치한다. 단 실행 시에는 M3b(characterization 안전망)가 M3d(분해)보다 반드시 선행한다는 의존만 지키면 된다.

## §B Known Issues (착수 전 인지 사항)

| # | 앵커 (run-phase 재실측) | 이슈 | 대응 |
|---|---|---|---|
| 1 | update.go 3,276 LOC / 122,035 bytes (2026-07-13 실측) | 분해 대상 규모 — 함수 인벤토리: runUpdate:136, runTemplateSync:569, runTemplateSyncWithReporter:579, analyzeFiles:1260, isUserOwnedNamespace:1387, runTemplateSyncWithProgress:1005 | M3a에서 함수→서브패키지 매핑표 확정 후 이동 |
| 2 | namespace 가드 테스트 5계열: internal/template/split_namespace_test.go + internal/cli/update_namespace_{hns,harness_v2,harness_v4}_test.go + update_security_m2_test.go | **무수정 green이 HARD 제약** — assertion 변경 금지 | import 경로 조정만 허용 + 파일별 사유 기록 (REQ-TUX3-005) |
| 3 | go.mod: bubbletea v1.3.10 direct, bubbles v1.0.0 indirect. **bubbletea v1 first-party importer 실측(2026-07-14)**: `internal/merge/confirm.go` + `confirm_test.go` + `confirm_coverage_test.go` 3파일 — 전부 M3e 승격 범위(`go mod why` 체인 = internal/merge 단독). internal/cli의 4개 테스트 파일(coverage_improvement/target_coverage/update_skip_sync/wizard coverage_boost)은 주석·Windows skip 문자열만 포함(grep 노이즈 — import 아님). huh v1(§B #5, 범위 외 잔존)이 bubbletea v1을 transitive로 유지하므로 **go.mod에서 v1 완전 소거는 불가** — direct(비-indirect) 항목 소거가 목표 | **M1 잔여 부채** — 프리뷰 TUI 전제 미충족. -002가 선행 랜딩했을 수 있음 | pre-flight #6에서 현재 상태 재실측; 미랜딩 시 본 SPEC이 도입; M3e에서 internal/merge 3파일의 v1 import 제거(또는 잔존 사유 progress.md §E.2 문서화) (REQ-TUX3-017) |
| 4 | internal/merge/confirm.go:871 @MX:DEBT + :874 tea.NewProgram (879 LOC) | M1 이연 결정의 승격 트리거가 본 SPEC — checkbox UI를 bubbles v2 list로 | 승격 후 @MX 3태그 제거 + confirm 스위트 green (REQ-TUX3-011) |
| 5 | update.go:165 huh.NewConfirm (하네스 setup 컨펌) | huh v1 사용처 — 본 SPEC 범위 외(§C Out of Scope) | 무접촉; 분해 시 파일 이동만 |
| 6 | update.go:596 부근 version-match 스킵 최적화 (70-80% faster) | 분해 중 성능 최적화 경로 유실 위험 | characterization에 Already up-to-date 고속 경로 포함 (REQ-TUX3-003) |
| 7 | `--yes` / non-TTY 폴백 이중 경로 (update.go:668-682, 1026-1040) | ConfirmMerge 호출이 2곳 — 프리뷰 전환 시 한쪽 누락 위험 | 두 호출부 모두 프리뷰(또는 폴백) 단일 진입점으로 수렴 |
| 8 | bubbletea v2 API 미실측 | v2 table/viewport/list 시그니처는 본 세션 미검증 | run-phase 착수 시 Context7/godoc으로 API 확정 후 착수 |
| 9 | reexecNewBinary/셀프업데이트 경로 (update.go:453-566) | M2(-002)가 호출 위치를 옮길 수 있음 — 병렬 진행 시 충돌 표면 | pre-spawn sync check + 분해 시 해당 함수군은 시맨틱 무변경 이동만 |
| 10 | exit code 계약 (cmd/moai ExitCoder 체인) | 분해로 에러 래핑 경로가 바뀌면 exit code 회귀 | characterization에 exit code 매트릭스 포함 (REQ-TUX3-007) |

## §C Pre-flight (착수 전 의무 검증)

```bash
# 1. baseline 확인
git branch --show-current && git rev-parse HEAD

# 2. cross-platform build baseline
go build ./... && GOOS=windows GOARCH=amd64 go build ./...

# 3. lint baseline
golangci-lint run --timeout=5m 2>&1 | tail -5

# 4. 가드 테스트 + update/merge 스위트 green 앵커
go test ./internal/template/ -run 'TestSplitHarnessNamespaceNoLeak' -count=1 -v
go test ./internal/cli/ -run 'Namespace' -count=1
go test ./internal/merge/... -count=1

# 5. update.go 규모 + 함수 인벤토리 재실측
wc -l -c internal/cli/update.go
grep -n '^func ' internal/cli/update.go | wc -l

# 6. M1 부채 현황 재실측 (-002 선행 랜딩 여부)
go list -m charm.land/bubbletea/v2 charm.land/bubbles/v2 2>&1
grep -n 'charmbracelet/bubbletea\|charmbracelet/bubbles' go.mod

# 7. confirm.go @MX:DEBT 태그 실존 확인 (승격 대상 앵커)
grep -n '@MX:DEBT\|@MX:CEILING\|@MX:UPGRADE' internal/merge/confirm.go

# 8. ratchet baseline 재실측
grep -rn 'fmt\.Printf\|fmt\.Println\|fmt\.Print(' internal/cli --include='*.go' | grep -v '_test.go' | wc -l

# 9. coverage baseline (비회귀 기준점)
go test -cover ./internal/cli/ ./internal/template/ ./internal/hook/... 2>&1 | grep coverage
```

## §D Constraints (DO NOT VIOLATE)

- **PRESERVE (절대 무접촉)**: namespace 보호 **정책**(user-owned 경로 집합), `internal/cli/init.go`·wizard(M2 소관), `internal/cli/uikit/banner.go`(M4 소관), `internal/statusline/**`, 무관 SPEC 디렉터리, `.moai/state/**`.
- **가드 테스트 assertion 무수정**: TestSplitHarnessNamespaceNoLeak 계열 + update_namespace_* + update_security_m2 — import 경로 조정 외 일절 금지. 조정 시 파일별 사유 progress.md §E.2 기재.
- **behavior-preserving 이동**: 분해 커밋에는 로직 변경 혼입 금지 — "이동 커밋"과 "표시 계층 신규 커밋"을 분리.
- **셀프업데이트 시맨틱 무변경**: `shouldSkipBinaryUpdate`/`runBinaryUpdateStep`은 이동만 허용(호출 순서 변경은 M2/-002 소관 — 병렬 진행 시 충돌 주의).
- **Template-First 비적용**: 내부 Go 코드만 — `internal/template/templates/**` 무접촉.
- **금지 명령**: `--no-verify`, `--amend`, force-push. Conventional Commits (`feat(SPEC-CLI-TUX-V3-003): M3a ...`) + `🗿 MoAI` trailer 의무.

## §E Self-Verification (완료 보고 의무 항목)

- E1: acceptance.md 20 AC 전 행 PASS/FAIL 매트릭스 + verbatim 출력 (vci 5-section 형식).
- E2: `go build ./... && GOOS=windows GOARCH=amd64 go build ./...` verbatim.
- E3: `go test -cover ./internal/cli/update/...` 서브패키지별 ≥ 85% + cli/template/hook 비회귀 (pre-flight #9 대조).
- E4: namespace 가드 테스트 5계열 전량 green + assertion 무수정 증명 (`git diff <preflight-HEAD>..HEAD -- <가드 테스트 경로>` 인용).
- E5: `golangci-lint run --timeout=5m` NEW 항목 0.
- E6: ratchet 재실측 — pre-flight baseline 미만 + 실제 감소치.
- E7: 신규 commit SHA 목록 + push 상태 + (해당 시) bubbletea v1 제거 여부/잔존 사유.

## §F Milestones (priority order — 결정-가역성 우선)

### M3a — 분류 데이터 모델 + 패키지 경계 확정 (Priority: High, 설계 결정)

1. 변경 분류 enum(add/update/preserve-user-owned/conflict) 타입 + 소비 계약 정의 — 프리뷰/폴백/보호 판정의 단일 원천. (REQ-TUX3-001/002)
2. update.go 함수 인벤토리(pre-flight #5) → `{plan,backup,deploy,merge,report}` 매핑표 작성, progress.md §E.2 기록.
3. bubbletea/bubbles v2 도입(또는 -002 선행 랜딩 확인). (REQ-TUX3-017)
4. **게이트**: 매핑표 + 분류 타입 계약 테스트 green.

### M3b — Characterization 안전망 (Priority: High, PRESERVE)

1. 플래그 매트릭스(`--yes`/templates-only/force/dry-run) × outcome 3종 × exit code characterization 테스트 작성 — 현행 green 확인. (REQ-TUX3-003)
2. Already up-to-date 고속 경로(version-match 스킵) 포함.
3. namespace 보호 시나리오(user-owned 파일 보존) end-to-end characterization.
4. **게이트**: characterization 스위트 green (분해 전 앵커 확정).

### M3c — 변경 프리뷰 TUI + 폴백 (Priority: High, UX 플로우)

1. RED: 프리뷰 계약 테스트 — 분류 테이블 렌더(카운트/행), diff viewport 도달, `--yes`/non-TTY 폴백, NO_COLOR ANSI 부재.
2. GREEN: Bubble Tea v2 테이블 + viewport 구현; 텍스트 폴백은 동일 분류 모델 소비. (REQ-TUX3-008/009/010)
3. `preserved (user-owned)` 라벨 가시화 — 보호 predicate 직접 소비. (REQ-TUX3-014)
4. ConfirmMerge 2개 호출부(Known Issue #7)를 단일 프리뷰 진입점으로 수렴.
5. **게이트**: AC-TUX3-008~010, 014, 015 통과.

### M3d — update.go 분해 (Priority: Medium, 기계적 — M3b 안전망 필수 선행)

1. M3a 매핑표대로 서브패키지 이동 — 모듈별 커밋 분리(plan→backup→deploy→merge→report), 각 커밋 후 characterization + 가드 테스트 green.
2. namespace 보호 로직 무손실 이동(assertion 무수정). (REQ-TUX3-004/005)
3. 3-way merge 오케스트레이션 이동 — internal/merge 스위트 green. (REQ-TUX3-006)
4. root update.go를 cobra 배선 + 오케스트레이션 글루로 축소.
5. **게이트**: AC-TUX3-003~007 + characterization 전량 green.

### M3e — confirm.go v2 승격 + outcome 카드 (Priority: Medium, UX + 부채 해소)

1. confirm.go characterization(선택 토글/전체 승인/취소/Windows non-TTY 폴백) 확보.
2. bubbles v2 list 재구현 + update 충돌 해소 재사용 배선; @MX:DEBT/@MX:CEILING/@MX:UPGRADE 제거. (REQ-TUX3-011)
3. 통일 outcome 카드(3 outcome 단일 렌더러) + 백업 경로·복구 커맨드 상시 표기. (REQ-TUX3-012/013)
4. bubbletea v1 direct 의존 제거 시도 — 잔존 소비자 있으면 사유 문서화. (REQ-TUX3-017 후반)
5. **게이트**: AC-TUX3-011~013 + confirm/merge 스위트 green.

### M3f — 정합성 통합 테스트 + 회귀 매트릭스 (Priority: Low, 게이트)

1. 프리뷰-집행 정합 통합 테스트: `preserve (user-owned)` 분류 파일이 배포 후 byte-identical. (REQ-TUX3-016)
2. 커버리지 측정 — 서브패키지 ≥ 85% + cli/template/hook 비회귀. (REQ-TUX3-018)
3. ratchet 재실측 + 회귀 매트릭스(non-TTY × NO_COLOR × GOOS=windows). (REQ-TUX3-019/020)
4. **게이트**: 전체 스위트 green + §E 자가 검증 전 항목.

### 검증 명령 배치 (M3f 종료 시, 단일 턴 병렬 실행)

```bash
go test ./... -count=1 > /tmp/moai-verify/1-full.log 2>&1; echo "exit=$?"; tail -20 /tmp/moai-verify/1-full.log
go test -cover ./internal/cli/update/... > /tmp/moai-verify/2-cover.log 2>&1; echo "exit=$?"; cat /tmp/moai-verify/2-cover.log
go test ./internal/template/ -run 'TestSplitHarnessNamespaceNoLeak' -count=1 -v > /tmp/moai-verify/3-guard.log 2>&1; echo "exit=$?"; tail -5 /tmp/moai-verify/3-guard.log
GOOS=windows GOARCH=amd64 go build ./... > /tmp/moai-verify/4-win.log 2>&1; echo "exit=$?"
golangci-lint run --timeout=5m > /tmp/moai-verify/5-lint.log 2>&1; echo "exit=$?"; tail -10 /tmp/moai-verify/5-lint.log
grep -rn 'fmt\.Printf\|fmt\.Println\|fmt\.Print(' internal/cli --include='*.go' | grep -v '_test.go' | wc -l
grep -n '@MX:DEBT' internal/merge/confirm.go | wc -l; echo "expect 0 (승격 후)"
```

## §G Anti-Patterns and Risks

- **Anti-pattern: 분해와 로직 변경 혼합 커밋** — behavior-preserving 검증 불가. 이동 커밋과 신규 표시 계층 커밋 엄격 분리.
- **Anti-pattern: 가드 테스트를 신규 구조에 맞춰 "정리"** — assertion 수정은 보호 회귀를 은폐한다. import 경로 외 무수정 (§D).
- **Anti-pattern: 프리뷰가 독자 분류 휴리스틱 구현** — 보호 predicate와 이중 원천이 되면 분류≠집행 괴리(신뢰 붕괴). REQ-TUX3-002/016이 방어.
- **Anti-pattern: characterization 없이 분해 착수** — M3b는 M3d의 HARD 선행 조건.
- **Risk: 122KB 이동의 리뷰 불가능성** — 모듈별 커밋 분리(M3d-1) + `git log --follow` 추적 가능성 유지.
- **Risk: bubbletea v2 Windows 콘솔 회귀** — confirm.go의 기존 Windows non-TTY 분기(confirm.go:858 주석) characterization 선행.
- **Risk: -002와의 병렬 진행 충돌 (update.go 공유)** — pre-spawn sync check + 셀프업데이트 함수군 시맨틱 무변경 원칙(§D).
- **Risk: 커버리지 목표 미달 (레거시 추출 코드)** — 서브패키지 ≥ 85%는 신규 테스트 작성 비용 포함 — Tier L 산정 근거.

## §H Cross-References

- 원본 계획: `.moai/reports/moai-cli-tux-modernization-plan-20260710.html` §4 M3 + §5 U-1~U-4 + §6 리스크(사용자 자산 손실) + §7 성공 지표(프리뷰 분류 100% 표기 / namespace 가드 전량 green).
- 선행: SPEC-CLI-TUX-V3-001 (completed — confirm.go @MX:UPGRADE가 본 SPEC 지목). 형제: -002 (bubbletea/bubbles v2 도입 조율). 후속: -004.
- 계약 승계: REQ-CTX-017 (ratchet), REQ-CTX-011~014 (Printer 계약), AC-CLI-TUI-013 (hex 금지).
- 규약: `internal/cli/CLAUDE.md` (output streams / exit codes), CLAUDE.local.md §2 (Template-First — 본 SPEC 비적용 확인) · §6 (테스트 격리) · §21/§24 (namespace 정책 배경).
- 방법론: moai-workflow-ddd (M3b/M3d characterization) + moai-workflow-tdd (M3c/M3e).
