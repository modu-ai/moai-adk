# SPEC-SEC-HARDEN-003 — 구현 계획 (plan.md)

> Tier: S (minimal) · cycle_type: tdd (reproduction-first) · 선행: SEC-HARDEN-001/002 (`completed`)

## §A. Context

SEC-HARDEN §F.3 fast-follow. 2건의 MEDIUM 경로-봉쇄 결함을 외과적 가드로 해소한다. AC SSOT는 `acceptance.md`다.

- C-F1 — `internal/hook/file_changed.go` `runMXScan` (~L129-173): 비격리 MX 사이드카 스캔·쓰기.
- C-F2 — `internal/cli/update.go` `restoreMoaiConfigLegacy` (~L2040-2090) + in-scope sibling `restoreMoaiConfig` (~L1964-2035): 심볼릭 링크 추종 복원.

## §B. Known Issues / Ground-Truth (2026-06-14 verified)

| 결함 | 파일:함수 | 정확 라인 | 재사용 seam |
|------|-----------|-----------|-------------|
| C-F1 read | `internal/hook/file_changed.go:runMXScan` | L143 `scanner.ScanFile(input.FilePath)` | `resolveProjectRootFromInputOrEnv` (path_resolve.go) |
| C-F1 write | 동상 | L154→L156→L157-158 `mx.NewManager(stateDir).UpdateFile` | 동상 |
| C-F2 read | `internal/cli/update.go:restoreMoaiConfigLegacy` | L2062 `os.ReadFile(backupPath)` | SEC-HARDEN-002 specid idiom 모델 |
| C-F2 write | 동상 | L2060 `filepath.Join(configDir, relPath)` + L2072/L2082/L2085 `os.WriteFile` | 동상 |
| C-F2 sibling | `internal/cli/update.go:restoreMoaiConfig` | L1982 join + L1985 read + L1998/L2019/L2031/L2034 write | 동상 |

## §C. Pre-flight (구현 착수 전 확인)

- [ ] `resolveProjectRootFromInputOrEnv` 시그니처 재확인(`*HookInput`, caller string) → root string.
- [ ] `runMXScan`의 `input *HookInput` 가용성 확인(이미 인자로 받음).
- [ ] `configDir` 가용성 확인 — `restoreMoaiConfigLegacy`는 함수 인자, 모던 `restoreMoaiConfig`는 L1948 지역변수로 익명 walk 콜백 스코프 내(인자 아님).
- [ ] 새 root-relative 봉쇄 헬퍼가 specid 패키지와 책임이 다른지 확인(문자열 sanitizer vs 경로 봉쇄).

## §D. Constraints (HARD)

- 비동기 goroutine 구조·5s `asyncDeadline`·빈 payload 응답 계약 무변경 (NFR-SEC3-001).
- `os.Getenv` 인라인 금지 — `resolveProjectRootFromInputOrEnv`만 사용 (NFR-SEC3-002).
- 새 패키지·새 공개 플래그·새 추상화 금지 (REQ-SEC3-009).
- 모든 거부는 fail-closed + 구조화 로그, panic 금지 (NFR-SEC3-004).
- 크로스 플랫폼 (`filepath.Rel`/`os.Lstat`) (NFR-SEC3-005).

## §E. Self-Verification (구현 후 read-only 배치)

```bash
# 1. 변경 패키지 전체 테스트
go test ./internal/hook/... ./internal/cli/...
# 2. 커버리지 회귀 확인
go test -cover ./internal/hook/... ./internal/cli/...
# 3. os.Getenv 인라인 미사용 (C-F1 NFR-SEC3-002)
grep -n "os.Getenv" internal/hook/file_changed.go   # expect: 0 matches
# 4. 비동기 구조 무변경 확인 (asyncDeadline 토큰 보존)
grep -rn "asyncDeadline" internal/hook/file_changed.go
# 5. lint
golangci-lint run --timeout=2m ./internal/hook/... ./internal/cli/...
# 6. race (hook async goroutine)
go test -race ./internal/hook/...
# 7. 크로스 컴파일
GOOS=windows GOARCH=amd64 go build ./...
```

## §F. Milestones (Tier S — 2 milestone, reproduction-first)

### M1 — C-F1 MX 사이드카 봉쇄 (`internal/hook/file_changed.go`)

reproduction-first (Safe Dev Rule 4):

1. **RED**: `TestRunMXScan_RejectsUncontainedFilePath` + `TestRunMXScan_RejectsUncontainedSidecarCWD` 작성. 악의적 `input.FilePath`(루트 탈출) / 악의적 `input.CWD`(루트 밖 사이드카 쓰기)로 봉쇄 위반을 재현 → 현재 코드에서 실패(스캔/쓰기 발생) 확인.
2. **GREEN**: `runMXScan`에 봉쇄 가드 추가 — `resolveProjectRootFromInputOrEnv(input, "runMXScan")`로 루트 해소 후, `input.FilePath`와 `stateDir`(또는 쓰기 대상) root-relative 봉쇄 검증. 위반 시 `slog.Warn` + early return (REQ-SEC3-001/002/003). 빈 payload 계약 보존 (REQ-SEC3-004).
3. **REFACTOR**: root-relative 봉쇄 검사를 작은 사적 헬퍼로 정리(필요 시). 비동기 구조 무변경 검증.
4. 회귀 AC: `TestRunMXScan_AllowsInProjectPath` — 정상 in-project 경로는 기존대로 스캔·쓰기.

대상: REQ-SEC3-001, 002, 003, 004 / AC cluster C-F1.

### M2 — C-F2 백업 복원 봉쇄 (`internal/cli/update.go`, 레거시 + 모던 sibling)

reproduction-first:

1. **RED**: `TestRestoreMoaiConfigLegacy_SkipsSymlinkEntry` + `TestRestoreMoaiConfigLegacy_RejectsTraversalTarget` + `TestRestoreMoaiConfig_SkipsSymlinkEntry` 작성. 심볼릭 링크 백업 엔트리 / configDir 탈출 `relPath`로 취약점 재현 → 현재 코드에서 실패(링크 추종 읽기·쓰기 발생) 확인.
2. **GREEN**: 레거시·모던 두 `filepath.Walk` 콜백에 `os.Lstat` 심볼릭 링크 스킵(REQ-SEC3-005/006) + `os.WriteFile` 전 `filepath.Rel(configDir, targetPath)` 탈출 거부(REQ-SEC3-007) 추가. 거부는 fail-closed.
3. **REFACTOR**: 두 경로가 공유할 작은 사적 봉쇄 헬퍼로 중복 제거(specid import 없음).
4. 회귀 AC: `TestRestoreMoaiConfigLegacy_AllowsRegularInConfigFile` — 정규 파일 + configDir 내부 target은 기존 머지·복원 동작 보존 (REQ-SEC3-008).

대상: REQ-SEC3-005, 006, 007, 008, 009 / AC cluster C-F2.

> M1 → M2 순서. 단일 결합 milestone도 가능하나, 두 결함이 서로 다른 패키지(`internal/hook` vs `internal/cli`)이고 테스트 격리가 명확하므로 2-milestone으로 분리한다.

## §G. Anti-Patterns (회피)

- 새 봉쇄 패키지 생성(REQ-SEC3-009 위반) — 사적 헬퍼로 충분.
- `os.Getenv` 인라인 C-F1 루트 해소(NFR-SEC3-002 위반).
- 모던 경로 sibling 누락(레거시만 봉쇄 시 주 경로 취약 유지 — security theater).
- 비동기 side-effect 실패를 hook 응답에 전파(REQ-HAE-005 위반).
- 봉쇄 위반 시 panic(NFR-SEC3-004 위반) — fail-closed return만.

## §H. Cross-References

- `internal/hook/CLAUDE.md` — $CLAUDE_PROJECT_DIR resolution priority (B7).
- `internal/cli/CLAUDE.md` — absolute path / filepath.Rel discipline.
- `internal/cli/specid/specid.go` — SEC-HARDEN-002 봉쇄 idiom 모델.
- `.claude/rules/moai/languages/go.md` — 커버리지 ≥85%, race, 크로스 플랫폼.
