# t149 — SPEC-HARNESS-GATE-TEST-001 유령 참조 판정

브랜치: `WT-t149` (base `release/v3.1.1` @ `b317f47c4`)

## 1. 판정

**(b) 존재한 적 없는 허수 참조.** 코드 주석 ID를 제거했다.

### 근거

| 확인 | 명령 | 관측 |
|---|---|---|
| SPEC 디렉터리가 전 ref 이력에 존재한 적 있는가 | `git log --all --oneline -- ".moai/specs/*HARNESS-GATE-TEST*"` | 출력 0건 |
| 현재 트리(아카이브 포함)에 있는가 | `find .moai/specs -maxdepth 1 -type d` + `.moai/specs/_archive` 대조 | 없음 |
| 코드 유입 커밋 | `git log --all -S "SPEC-HARNESS-GATE-TEST-001" -- internal/hook/quality/` | `f1ebd634a` 단독 (후속 `692d44586`) |
| 흡수처 SPEC이 있는가 | `grep -rln "resolveNodeTestStep\|npm run test:run\|watch-prone" .moai/specs/` | 0건 |

부수 관측: 로컬 브랜치 `spec-harness-gate-test-001`이 존재한다. 브랜치는 만들어졌으나
SPEC 문서는 작성되지 않았고, 두 커밋 모두 `release/v3.1.1`에 이미 들어가 있다.
즉 커밋 시점에 ID만 지어낸 상태다. 분기 (a)(작성 후 유실)와 (c)(다른 SPEC 흡수)는
위 4개 관측으로 각각 배제된다.

## 2. 교정 내역

주석 산문이 ID 없이도 자립하므로, SPEC을 소급 작성하지 않고 ID만 제거했다.
변경의 출처는 커밋 `f1ebd634a` / `692d44586` 메시지가 그대로 보존한다.

| 파일 | 위치 | 조치 |
|---|---|---|
| `internal/hook/quality/gate.go` | 308 (Run Step 3 주석) | ID 접두 제거 + 재랩 |
| `internal/hook/quality/gate.go` | 573 (`resolveNodeTestStep` godoc) | 괄호 ID 제거 + 재랩 |
| `internal/hook/quality/gate_node_resolution_test.go` | 3-4 (파일 헤더) | 괄호 ID 제거, 1행으로 통합 |
| `.moai/reports/t146/release-notes-draft.md` | `### Fixed` 항목 | 항목 제목의 ID 접두 제거 (배포되는 표면) |
| `.moai/reports/t146/release-notes-draft.md` | 미해결 항목 | 이력을 고쳐 쓰지 않고 t149 해소 결과를 하위 항목으로 추가 |

## 3. 검증

| 항목 | 명령 | 결과 |
|---|---|---|
| 잔여 참조 | `grep -rn "HARNESS-GATE-TEST" .` (`.git`, `reports/` 제외) | 0건 |
| 포맷 | `gofmt -l internal/hook/quality/` | 출력 없음 |
| 정적 분석 | `go vet ./internal/hook/quality/...` | exit 0 |
| 테스트 | `go test ./internal/hook/quality/... -count=1` | `ok ... 17.549s` |

전체 스위트는 카드 지시대로 돌리지 않았다 — 전 패키지 판정은 CI 몫이다.

## 4. 부수 점검 — 같은 형태 전수 조사

프로덕션 Go 파일(`*_test.go` 제외)이 참조하는 SPEC ID 270개를
`.moai/specs/` + `.moai/specs/_archive` 디렉터리 목록과 대조했다.
디렉터리가 없는 ID 18개가 나왔고, 성격별로 셋으로 갈린다.

### (가) 예시용 placeholder — 조치 불필요 (11건)

`e.g.` / 사용법 예시 문자열이라 실재하는 SPEC을 가리킬 의도가 없다.
`SPEC-AUTH-001`, `SPEC-FOO-001`, `SPEC-X-001`, `SPEC-ISSUE-123`,
`SPEC-NAV-001`, `SPEC-PROJ-001`, `SPEC-PROJ-INIT-001`, `SPEC-SPC-001`
(`mx_query.go`, `cc.go`, `graph.go`, `github.go`, `kanban.go`,
`worktree/shared.go`, `harness/install.go` 등).

### (나) 계획된 후속 작업에 대한 전방 참조 — 판단 보류 (5건)

"아직 안 만든 SPEC"을 의도적으로 가리키는 TODO/유보 표기다. 허위 출처 주장은
아니지만, 그 SPEC이 끝내 작성되지 않으면 (다)와 같아진다.

| ID | 위치 |
|---|---|
| `SPEC-V3R4-CATALOG-005` | `internal/template/slim_guard.go`, `internal/cli/init.go` ("forward reference to auto-bootstrap" 명시) |
| `SPEC-V3R5-PROJECT-MEGA-001` | `internal/cli/harness_mute.go`, `internal/harness/seeds/loader.go` ("W4 scope") |
| `SPEC-V3R6-GEARS-SWEEP-001` | `internal/spec/lint.go` ×2 |
| `SPEC-V3R6-V3-CUTOVER-001` | `internal/spec/lint.go` ×2 |
| `SPEC-V3R2-SCH-001` | `internal/session/checkpoint.go` ×2 (`TODO: implement once …`) |

### (다) t149와 같은 형태 — 출처 주장인데 대상이 없음 (4건)

과거형 출처 표기라 링크를 걸면 깨진다. 이 카드 범위 밖이므로 손대지 않았고,
후속 카드 후보로 넘긴다.

| ID | 위치 |
|---|---|
| `SPEC-CLI-WORKTREE-ADVISORY-001` | `internal/core/project/initializer.go:56`, `internal/cli/wizard/types.go:53` |
| `SPEC-MONITOR-001` | `internal/hook/post_tool.go:166` |
| `SPEC-V3R6-HOOK-RECOVERY-SIGNAL-001` | `internal/hook/user_decision_capture.go:73` |
| `SPEC-V3R6-LINT-CLASSIFYPRTITLE-001` | `internal/spec/transitions.go:113` (`@MX:REASON`) |

`_test.go`까지 포함하면 미존재 ID가 수백 건이지만 대부분 테스트 픽스처
(`SPEC-TEST-*`, `SPEC-BROKEN-*`, `SPEC-NONEXISTENT-*`)라 의도된 형태다.

## 5. 미검증 / 잔여 위험

- 전체 테스트 스위트를 돌리지 않았다. 이번 변경은 주석 전용이라 동작에 영향이 없지만,
  전 패키지 판정 근거는 CI다.
- 로컬 브랜치 `spec-harness-gate-test-001`은 남아 있다. 두 커밋이 이미 `release/v3.1.1`에
  들어갔으므로 정리 대상이지만, 브랜치 삭제는 이 카드가 하지 않는다 — 리드 판단.
- (나) 5건은 "언젠가 만들 SPEC"이라는 전제가 아직 유효한지 확인하지 않았다.
  전제가 죽었다면 (다)와 같은 교정 대상이 된다.
