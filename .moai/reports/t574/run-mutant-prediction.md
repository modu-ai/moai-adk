# t574 — run-phase mutant M-574 prediction (pinned and committed before injection)

카드 t574 · 트리 `.claude/worktrees/t574` · 브랜치 `WT-temp-roots-ac` · 주입 전 HEAD `d937cd68d` · go1.26.8 darwin/arm64.

이 파일은 plan-audit D2(뮤턴트 대조에 명령·뮤턴트 소스·exit 코드가 필드로 남지 않음)를 닫기 위한 run-phase 재측정의 예측이다. 뮤턴트를 주입하기 **전에** 커밋되며, 그 순서는 커밋 그래프로 확인한다(예측 커밋이 증거 커밋의 조상).

## 뮤턴트

- 대상: `internal/kanban/temp_origin.go` 의 `defaultTempRoots()` 본문 한 줄
- 원본: `return []string{os.TempDir(), "/tmp", "/var/folders"}`
- 뮤턴트: `return []string{os.TempDir()}`
- `os` 는 계속 쓰이므로 패키지는 컴파일된다. 컴파일 실패는 유효한 RED가 아니다.

## 판정 명령 (acceptance.md § AC-THG-006 그대로)

```
go test ./internal/kanban/ -run '^(TestTempOrigin_ComponentBoundary|TestDefaultTempRoots_Membership)$' -count=1 -v
```

## 예측

| 테스트 | 기준 트리 | 뮤턴트 아래 |
|---|---|---|
| `TestTempOrigin_ComponentBoundary` | PASS | **PASS** (경계 절은 스텁 루트와 음의 방향 `/tmpfoo` 만 판정 — 원소를 빼도 성립) |
| `TestDefaultTempRoots_Membership` | PASS | **FAIL** — `fixed temp root "/tmp" missing` · `fixed temp root "/var/folders" missing` · `has 1 members, want 3` |

- 최상위 `=== RUN` 은 두 실행 모두 정확히 2건.
- 뮤턴트 실행의 exit 코드는 1, 기준 실행은 0.

## 반증 조건

- `TestTempOrigin_ComponentBoundary` 가 뮤턴트 아래 FAIL → 경계 절이 이미 소속을 묶고 있다는 뜻이며 개정 근거의 서술이 틀렸다.
- `TestDefaultTempRoots_Membership` 가 뮤턴트 아래 PASS → 소속 절의 판정 테스트가 공허하다.
- 최상위 RUN 이 2건이 아님 → 선택자가 다른 집합을 골랐고 그 결과로는 판정하지 않는다.
- 빌드 실패 → 뮤턴트 주입 실패이며 RED로 세지 않는다.
