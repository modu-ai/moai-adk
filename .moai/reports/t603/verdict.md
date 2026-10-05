# t603 — sync-phase 품질 게이트의 C++ 검사 복구

카드: t603 (hooks 감사 2026-09-11 · H08 · P2)
기준 커밋: `eabce7444` (origin/develop, t601 병합 직후)
워크트리: `.claude/worktrees/t603` · 브랜치 `WT-gate-cpp-checker`

## Claim

1. `.cpp` 파일이 C++ 문법 검사기에 전달되지 않던 결함(H08)을 수리했다.
2. 같은 줄에 있던 두 번째 결함 — 검사기가 **실패해도** 게이트가 통과시키던 문제 — 를 함께 수리했다.
3. 검사 대상이 0건인 경우를 통과한 검사와 구분해 감사 로그에 남긴다.
4. 프로젝트 사본과 템플릿 사본이 바이트 동일하다.

## Evidence

### 재현 (수리 전, 기준 `eabce7444`)

대상 줄은 감사 기준의 `:267`이 아니라 **`:574`**였다. t601이 같은 파일을 크게 고쳐 파일이 696행으로 늘었다. 명령은 그 줄에서 그대로 추출했다.

```
find . -name "*.cpp" -o -name "*.cc" -exec g++ -fsyntax-only -std=c++17 {} \; 2>&1
```

도달성 대조 — 두 fixture 모두 컴파일러가 직접 거부한다. 이 대조가 없으면 뒤따르는 실패를 해석할 수 없다.

```
g++ -fsyntax-only -std=c++17 bad.cpp  → rc=1
g++ -fsyntax-only -std=c++17 bad.cc   → rc=1
```

| 시험군 | c1 종료 코드 | 컴파일러 출력 | 게이트 판정 |
|---|---|---|---|
| `.cpp` 단독 | 0 | 없음 (0바이트) | allow |
| `.cc` 단독 | 0 | 있음 — `./bad.cc:1:14: error: ... 1 error generated.` | allow |

대조군 출력 원문: `repro-cc-output.txt`

### 원인

두 결함이 한 줄에 겹쳐 있었다.

**(A) `find` 연산자 우선순위.** `-name "*.cpp" -o -name "*.cc" -exec …`는 `-name "*.cpp" -o ( -name "*.cc" -a -exec … )`로 파싱된다. `.cpp` 파일은 첫 항에서 단락 평가되어 `-exec`에 도달하지 않는다. 표현식이 액션을 이미 가지고 있어 기본 `-print`도 붙지 않으므로 출력조차 남지 않는다 — 감사 기록의 `compiler_invoked=false`와 일치한다.

**(B) 종료 상태 미전파.** `find`는 `-exec … \;` 형식에서 실행한 유틸리티의 종료 상태를 자신의 종료 상태로 전파하지 않는다. 따라서 컴파일러가 실제로 실패해도 `c1=0`이 기록되고, 게이트의 `C1_EXIT -ne 0` 분기에 걸리지 않는다. 두 결함의 합은 **C++ 게이트가 어느 확장자로도 차단하지 못하는 상태**였다.

### 수리

```
out=$(find . \( -name "*.cpp" -o -name "*.cc" \) -print -exec g++ -fsyntax-only -std=c++17 {} + 2>&1); rc=$?; ...
```

- `\( … \)` — 확장자 조건을 묶어 공통 `-exec`가 모든 일치 파일에 적용된다. (A) 수리.
- `{} +` — POSIX는 `+` 형식에서만 유틸리티의 비0 종료를 `find`의 종료 상태로 전파한다. (B) 수리.
- `-print` — 검사기에 넘긴 파일 목록을 기록한다. 출력이 비면 대상이 0건이라는 뜻이고, 그 경우만 별도 문구를 남긴다.
- 0건 문구는 `GATE_TMPDIR`의 스크래치 로그에만 남으면 종료 시 삭제되어 아무도 읽지 못하므로, 기존 `log_gate_event` 헬퍼로 감사 로그에 승격한다.

### 수리 후 대조 (5개 시험군, 기준 `eabce7444` 워크트리)

| 시험군 | 기대 | 측정된 종료 코드 |
|---|---|---|
| 깨진 `.cpp` 단독 | 차단 | 1 |
| 깨진 `.cc` 단독 | 차단 | 1 |
| 정상 `.cpp` 단독 | 통과 | 0 (`./ok.cpp` 목록 출력) |
| 소스 0건 (헤더만) | 통과 + 0건 기록 | 0 (출력 0바이트) |
| 정상 `.cpp` + 깨진 `.cc` | 차단 | 1 (두 파일 모두 목록에 출력) |

### 회귀 테스트

`internal/template/hook_cpp_gate_behavior_test.go` — 게이트 스크립트를 임시 git 저장소에 대고 **실제로 실행**한다. 원 결함이 소스 검사로는 보이지 않았기 때문이다(명령이 올바르게 생겼고 조용히 아무것도 검사하지 않았다).

```
--- PASS: TestSyncGateCpp_BrokenCppBlocks (1.82s)
--- PASS: TestSyncGateCpp_BrokenCcBlocks (2.42s)
--- PASS: TestSyncGateCpp_ValidSourcePasses (3.55s)
--- PASS: TestSyncGateCpp_ZeroTargetsReportedDistinctly (2.50s)
--- PASS: TestSyncGateCpp_LocalAndTemplateCopiesIdentical (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/template	10.893s
```

### 변이 검사 (테스트가 공허하지 않음을 확인)

수리의 두 절반을 각각 되돌려, 각 절반이 실제로 작동하는지 측정했다.

| 변이 | 기대 | 측정 |
|---|---|---|
| 괄호 제거 (`+`는 유지) | `.cpp`만 빨강 | `BrokenCppBlocks` FAIL, `BrokenCcBlocks` PASS |
| `+` → `\;` (괄호는 유지) | 둘 다 빨강 | `BrokenCppBlocks` FAIL, `BrokenCcBlocks` FAIL |
| 0건 승격 블록 제거 | 0건 시험만 빨강 | `ZeroTargetsReportedDistinctly` FAIL |

세 변이 모두 되돌린 뒤 해시가 수리본과 일치함을 확인했다.

### 패키지 검증

```
go test ./internal/template/... -count=1
ok  	github.com/modu-ai/moai-adk/internal/template	55.180s
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	1.594s
ok  	github.com/modu-ai/moai-adk/internal/template/commandemit	1.847s

go vet ./internal/template/...   → 종료 코드 0
bash -n (두 사본)                 → 종료 코드 0
make build                        → 종료 코드 0 (catalog.yaml 드리프트 없음)
```

## Baseline-attribution

위 모든 수치는 워크트리 `.claude/worktrees/t603`(기준 `eabce7444`)에서 이번에 실행해 관측한 값이다. 감사 시점(2026-09-11)의 수치를 재사용한 항목은 없다. 감사 기준이 지목한 줄 번호(`:267`)는 이번 트리에서 `:574`로 재측정해 정정했다.

## Gaps

- **두 사본 동일성은 확인했으나, 배포된 바이너리로는 검증하지 않았다.** `make build`는 성공했고 `catalog.yaml` 드리프트도 없으나, 임베드된 사본이 실제로 이 내용인지를 별도로 대조하지는 않았다.
- **macOS(BSD find)에서만 측정했다.** GNU find은 미측정이다. `-exec … +`의 종료 상태 전파는 POSIX 규정이므로 동일할 것으로 보지만, 확정된 관측은 아니다.
- **게이트 스크립트의 다른 언어 분기는 건드리지 않았고 측정하지도 않았다.** `:582` R 분기는 `-o`를 쓰지만 `-exec`가 없어 기본 `-print`가 전체 표현식에 붙으므로 같은 결함이 **없다**고 판독했다 — 이것은 코드 판독이며 실행 측정이 아니다.
- **CI 전체 스위트는 미실행.** 범위는 `internal/template`으로 한정했다. 전 패키지 판정은 CI 몫이다.

## Residual-risk

- **이 수리는 게이트의 동작을 바꾼다.** 지금까지 C++ 프로젝트에서 sync-phase 게이트는 어떤 컴파일 오류로도 차단하지 못했다. 수리 후에는 차단한다. 조용히 통과하던 저장소가 갑자기 막히는 것으로 보일 수 있으며, 이는 결함이 아니라 게이트가 처음으로 제 일을 하는 것이다. CHANGELOG에 한 줄 남겼다.
- **`.cxx` / `.hpp` / `.hxx`는 여전히 검사되지 않는다.** 게이트의 `code_delta_pattern`(`:149`)은 이 확장자들을 C++로 인정하지만 `:574`의 `find`는 `*.cpp`와 `*.cc`만 찾는다. 이번 카드의 범위 밖이며 리드가 별도 발행 목록에 올렸다.
- **`{} +`는 파일을 한 번에 넘긴다.** 대상 파일이 매우 많은 저장소에서는 명령줄 길이 한계로 `find`가 호출을 여러 번으로 나눈다. 동작은 올바르지만 출력 순서가 파일 수에 따라 달라질 수 있다.
- **`t604`(H09)가 같은 파일의 `:62`를 동시에 수정한다.** 병합 창에서 순서대로 흡수해야 하며, 나중에 병합하는 쪽이 충돌을 해소한다.
