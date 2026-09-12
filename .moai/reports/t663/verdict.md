# t663 — C++ 게이트가 `.cxx` / `.hpp` / `.hxx` 를 보지 않는 비대칭

카드 t663 (Tier S, Class B). 브랜치 `WT-cpp-ext-scan`, 기반 `2448be066` (로컬 develop, t664 병합 포함).

**상태: 구현·검증 완료.** 카드가 첫 판정으로 지정한 헤더 취급 문제를 실측해 **범위를 `.cxx` 로 좁혔고**, 그 결정을 고정하는 가드를 함께 넣었다.

---

## Claim

게이트의 `code_delta_pattern`(`:181`)은 `cpp|cc|cxx|h|hpp|hxx` 여섯 확장자를 C++ 로 인정하지만, 실제 검사(`:620` 부근)의 `find` 는 `*.cpp` 와 `*.cc` 둘만 찾는다. 즉 변경을 C++ 로 **분류해 놓고 검사는 건너뛴다**. 수정은 `.cxx` 를 스캔에 넣고, 헤더 3종은 **의도적으로 넣지 않는** 것이다.

## Evidence — 도달성 대조 (카드 [HARD] #2, 먼저 수행)

게이트 판정을 해석하기 전에 각 확장자 픽스처가 컴파일러에 직접 거부되는지부터 확인했다. 깨진 내용(`int broken( {`)을 여섯 확장자로 복제해 `g++ -fsyntax-only -std=c++17` 에 단독으로 넘겼다.

| 확장자 | exit |
|---|---|
| `.cpp` / `.cc` / `.cxx` | 1 / 1 / 1 |
| `.h` / `.hpp` / `.hxx` | 1 / 1 / 1 |

여섯 전부 거부된다 — 픽스처가 실제 탐지기이며, 스캔에 들어가면 오류를 잡는다는 뜻이다. 이 대조가 없으면 이후의 "게이트가 잡았다/못 잡았다"를 해석할 근거가 없다.

**툴체인 주의**: 이 머신의 `g++` 은 GNU gcc 가 아니라 **Apple clang 21.0.0 (clang-2100.1.1.101)** 이다. 게이트는 `g++` 라는 이름을 호출하므로 리눅스 CI 의 GNU gcc 와 반응이 다를 수 있다. 아래 판정은 이 툴체인에서의 관측이다.

## 첫 판정 (카드 [HARD] #1) — 헤더는 소스와 같이 취급하지 않는다

카드가 요구한 대로 헤더 단독 `-fsyntax-only` 반응을 먼저 쟀다.

### 자족적 헤더는 통과한다

`inline int a() { return 0; }` 를 여섯 확장자로 복제:

| 확장자 | exit | 출력 |
|---|---|---|
| `.cpp` / `.cc` / `.cxx` | 0 | 없음 |
| `.hpp` / `.hxx` | 0 | 없음 |
| `.h` | 0 | `warning: treating 'c-header' input as 'c++-header' when in C++ mode, this behavior is deprecated [-Wdeprecated]` |

부산물(`.gch` 등)은 생성되지 않았다. 즉 "헤더는 단독 검사가 불가능하다"는 것은 **사실이 아니다**.

### 그러나 문맥 의존 헤더는 실패한다 — 이것이 결정 근거다

앞서 include 된 헤더가 제공하는 타입을 쓰는 헤더는 정상적인 C++ 관용구이고, 단독 검사에서 실패한다:

```
$ g++ -fsyntax-only -std=c++17 dep.hpp     # 내용: void use(MyType* p);
dep.hpp:1:10: error: unknown type name 'MyType'
exit=1
```

대조로, 불완전 템플릿 선언(`template <class T> struct S; S<int> make();`)은 exit 0 이다 — 즉 모든 헤더가 실패하는 것이 아니라, **정상적인 프로젝트의 정상적인 헤더가 실패한다**. 게이트가 헤더를 스캔하면 그런 프로젝트는 게이트 실패로 바뀐다.

**판정**: 헤더(`.h` / `.hpp` / `.hxx`)를 소스와 같이 취급하는 것은 옳지 않다. 카드가 제시한 대비책대로 범위를 **`.cxx` 로 좁혔다**. `.h` 의 deprecation 경고는 부차적 근거이며, 결정을 만든 것은 위의 오탐 실측이다.

## 구현

| 파일 | 변경 |
|---|---|
| `.claude/hooks/moai/sync-phase-quality-gate.sh` | `find` 그룹에 `-o -name "*.cxx"` 추가. 0-타깃 메시지와 `log_gate_event` 문구를 `*.cpp/*.cc/*.cxx` 로 갱신. 왜 `.cxx` 는 넣고 헤더는 넣지 않는지를 기존 "load-bearing" 주석 블록에 같은 형식으로 추가 |
| `internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh` | 위와 **동일 변경** (§2 Template-First). 두 사본은 변경 전에도 바이트 동일이었고 변경 후에도 동일하다 |
| `internal/template/hook_cpp_gate_behavior_test.go` | 테스트 2본 추가 (아래) |

`code_delta_pattern` 은 **고치지 않았다.** 헤더를 패턴에서 빼면 헤더만 바뀐 커밋이 C++ 게이트를 통째로 건너뛰게 되어 더 나쁘다. 분류는 넓게, 검사는 오탐 없는 범위로 — 남는 비대칭은 아래 Residual 에 기록한다.

## 추가한 테스트와 RED 관측

기존 게이트 테스트 5본은 실행형(게이트를 실제로 돌린다)이라 그 형식을 그대로 따랐다.

| 테스트 | 무엇을 고정하는가 | RED 관측 방법 | 결과 |
|---|---|---|---|
| `TestSyncGateCpp_BrokenCxxBlocks` | 깨진 `.cxx` 가 게이트를 막는다 | 두 사본에서 `.cxx` 추가를 되돌림 | exit 1, `--- FAIL` (35.98s) |
| `TestSyncGateCpp_ContextDependentHeaderDoesNotBlock` | 문맥 의존 헤더가 게이트를 **막지 않는다** (헤더 제외 결정의 오탐 가드) | 두 사본의 스캔에 `-o -name "*.hpp"` 를 추가 | exit 1, `--- FAIL` (30.42s) |

두 mutant 모두 **두 사본에 동시에** 주입했다 — 한쪽만 바꾸면 기존 `TestSyncGateCpp_LocalAndTemplateCopiesIdentical` 이 먼저 터져서 어느 축을 재고 있는지 흐려진다. 주입 후 원본을 복원하고 md5 3개(로컬·백업·템플릿) 동일을 확인했다: `893586cd5820ef299d2a54d7d3fc0341`.

## 검증 실행 결과

| 검증 | exit | 관측 |
|---|---|---|
| `sh -n` 로컬 사본 | 0 | — |
| `sh -n` 템플릿 사본 | 0 | — |
| `diff -q` 두 사본 | 0 | 동일 |
| `gofmt -l` 테스트 파일 | — | 출력 없음 |
| `go vet ./internal/template/` | 0 | — |
| `make build` (템플릿 재임베드) | 0 | `catalog.yaml updated successfully (12899 bytes)`. 직후 변경 파일은 의도한 2본뿐 |
| `go test ./internal/template/ -run TestSyncGateCpp` | 0 | `=== RUN` 7행, `--- PASS` 7, `--- FAIL` 0 — 이름 7개 전부 실행 확인 |
| `go test ./internal/template/... -count=1` | 0 | `template 278.427s` / `agentemit 0.434s` / `commandemit 0.957s` 전부 `ok` (load 18.78 시점) |

### 수정 동작 확인 — 가드 우회 없이 두 조각으로

게이트의 실제 명령은 `find … -exec g++ …` 형태인데, 이 세션의 PreToolUse 가드가 그 형태를 거부한다. 스크립트 파일로 감싸면 가드가 내용을 읽지 못해 검사 전체가 우회되므로, 우회하지 않고 **검증을 두 조각으로 나눴다**:

- **선택**: 대조군 포함. 수정 전 표현식은 `good.cc`·`good.cpp` 2건만 고르고 `sub/bad.cxx` 를 **놓친다**. 수정 후 표현식은 3건을 고른다. 두 표현식 모두 `ctx.hpp` 를 고르지 않아 헤더 제외도 확인된다.
- **전파**: `{} +` 의 다중 인자 호출을 그대로 재현. 정상 2건만 넘기면 exit 0, 여기에 깨진 `.cxx` 를 더하면 exit 1.

두 조각이 붙은 **합성 자체**는 이 세션에서 직접 재지 못했다 — 그 몫은 위의 실행형 테스트 2본(`go test` 가 게이트를 진짜로 돌린다)이 담당하며, 그쪽에서 RED→GREEN 을 관측했다.

## Gaps — 아직 관측하지 않은 것

| 항목 | 사유 |
|---|---|
| GNU gcc 툴체인에서의 헤더 반응 | 이 머신 `g++` 은 Apple clang 이다. 헤더 오탐 결론은 clang 관측이며, 리눅스 CI 의 GNU gcc 에서 같은지는 미측정 |
| `.hxx` / `.h` 확장자의 게이트 경유 동작 | 스캔에 넣지 않기로 했으므로 게이트를 통과하는 경로 자체가 없다. 도달성 대조(단독 g++)만 측정 |
| `internal/cli` 전량 | 이 카드는 셸 훅과 템플릿 미러만 건드린다. 영향 패키지는 `internal/template` |
| `GOOS=windows` 빌드 | 미실행. 게이트는 POSIX 셸 스크립트이고 테스트도 windows 를 skip 한다. 판정은 CI 몫 |

## Residual-risk

- **비대칭이 완전히 사라진 것은 아니다.** `code_delta_pattern` 은 여전히 `.h` / `.hpp` / `.hxx` 를 C++ 로 인정하고, 검사는 그 3종을 건너뛴다. 이는 오탐(정상 프로젝트를 실패로 바꿈)과 미탐(헤더 문법 오류를 놓침) 사이의 **의도된 선택**이며, 위 실측이 그 근거다. 헤더를 검사하려면 프로젝트 문맥(include 순서·전제 헤더)을 알아야 하므로 게이트 구조가 달라져야 한다 — Tier S 범위 밖이다.
- 헤더만 바뀐 커밋은 `cpp_targets=0` 으로 감사 로그에 남는다(기존 `TestSyncGateCpp_ZeroTargetsReportedDistinctly` 가 지키는 동작). 즉 "검사 안 함"이 "통과"로 위장되지는 않는다 — 다만 그 로그를 사람이 읽어야 드러난다.
- `.h` 를 C++ 로 검사할 때 나오는 deprecation 경고는 exit 0 이라 판정에 영향이 없지만, 향후 툴체인이 이를 오류로 승격하면 `.h` 를 스캔에 넣는 선택지가 더 좁아진다.

## 적용 규칙

- `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 — 미실행 검증을 Claim 이 아니라 Gaps 로 기재. 툴체인 차이를 근거 한계로 명시
- `.claude/rules/moai/development/verification-completeness.md` §1.1 — 신규 테스트 2본 모두 RED 관측 후 GREEN
- `CLAUDE.local.md` §2 Template-First — 로컬·템플릿 두 사본 동시 수정 + `make build`
