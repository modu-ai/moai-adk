---
id: SPEC-PERF-FIXTURE-WRITE-001
title: "perf 하네스 테스트가 추적 파일을 무조건 덮어쓰는 결함 — 명시적 opt-in 게이트로 좁힌다"
version: "0.3.3"
status: completed
created: 2026-08-29
updated: 2026-08-30
author: manager-spec
priority: P2
phase: "v3.1.4 target"
module: "internal/hook/perf"
lifecycle: spec-anchored
tags: "perf, test-hygiene, tracked-fixture, opt-in-gate, mutation-testing, working-tree-cleanliness"
tier: S
era: V3R6
related_specs: [SPEC-HOOK-PRETOOL-PERF-001]
---

# SPEC: perf 리포트 무조건 쓰기 차단

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.3.3 | 2026-08-31 | manager-spec | **수락 기준은 하나도 바뀌지 않았고 요구 8 / 수락 7 불변**(범위 증식을 찾는 독자는 여기서 멈춰도 된다). **REQ-PFW-006 의 “on **every** exit path” 를 “on every **normal** exit path” 로 좁히고 잔여 2건을 괄호로 이름 붙였다** — `t.Cleanup` 은 kill(`SIGINT`/`SIGKILL`)에서도 `go test -timeout` 패닉에서도 돌지 않으므로, 종전 문면은 **문자 그대로 만족 불가능**했고 구현도 만족하지 않았다. F2 수리가 가드 doc comment 를 “every NORMAL exit path”로 좁히며 잔여 2건을 이미 명명했는데 **요구는 함께 좁혀지지 않았다**. **형제 1건도 함께 쓸었다**: `plan.md` §B.4 가 같은 주장을 한국어로 되풀이하며(“모든 종료 경로에서 복원”) **한 술 더 떠 “패닉 경로에서도 돈다”**고 적고 있었다 — `-timeout` 패닉이 정확히 돌지 않는 자리다. `acceptance.md` 는 깨끗했다(GWT 가 `t.Fatal`/실패 경로로 이미 한정). 어떤 AC 도 건드리지 않았다 — 감사 판정대로 기준은 멀쩡했고 과장한 것은 요구의 문면뿐이다. **이 행이 진짜 기록하는 것 — 한 카드에서 “계약이 코드를 과장한다”가 세 번째다.** ① `REQ-PFW-007(a)` 의 파일 경로가 grep 인자를 통해 요구 지위에 올랐고(아무도 합의한 적 없는 요구), ② `REQ-PFW-005` 가 열거된 변수 목록을 못 박아 이식성 결함에 요구의 옷을 입혔고, ③ 이번 `REQ-PFW-006` 이 만족 불가능한 전칭을 실었다. **셋 다 스윕이 아니라 감사가 찾았다.** 재사용할 수 있는 형태로 적는다: **요구의 문면을 고칠 때 그 정정의 폭발 반경은 같은 주장을 공유하는 모든 형제 요구이지, 발견이 지목한 파일 하나가 아니다.** 0.3.2 의 F4 스윕은 파일 셋을 훑고도 **바로 다음 요구 하나를 앞두고 멈췄다** — 파일 단위로 훑고 요구 단위로 훑지 않았기 때문이다. 다음 문면 정정은 지목된 파일이 아니라 **같은 주장을 하는 요구 집합**을 대상으로 시작한다 |
| 0.3.2 | 2026-08-30 | manager-spec | **REQ-PFW-005 의 문면을 구현 모양에서 성질로 바꿨다 — 수락 기준은 하나도 바뀌지 않았고 요구 8 / 수락 7 불변**(범위 증식을 찾는 독자는 여기서 멈춰도 된다). **무엇이 틀렸나 (sync 감사 F4).** 종전 문면 “carrying the sentinel and nothing else”(자식 환경에 센티널만 실어 보낸다)는 **쓰이던 순간부터 구현과 맞은 적이 없다** — 가드는 변수 17개를 열거해 넘기고 있었다. 뒤이은 편집이 어긋나게 만든 것이 아니라 처음부터 어긋나 있었다. **그 열거가 왜 차단 결함이었나.** 목록이 POSIX 전용(`PATH HOME TMPDIR USER LOGNAME GOROOT GOPATH GOCACHE …`)이라 Windows 자식에게 `LOCALAPPDATA`·`TMP`·`TEMP`·`USERPROFILE`·`SystemRoot` 를 넘기지 못한다. `GOCACHE` 가 비면 `os.UserCacheDir` 이 `LocalAppData` 를 읽으므로 툴체인이 `GOCACHE is not defined` 로 죽는다. **부채가 아니라 차단으로 만든 것은 착지 지점이다**: PR 필수 잡 `ci.yml:41` 은 ubuntu 전용이고 `release-pr-multi-os.yml:91` 이 `windows-latest` 를, `:189` 가 `go test -race ./...` 를 돈다 — **이 카드의 CI 는 초록으로 돌아오고 파손은 릴리스에서 드러난다.** **수리(`2b2b726ee`)는 목록을 뒤집었다**: 자식은 `os.Environ()` 을 물려받고 `MOAI_HOOK_PERF_UPDATE`·`MOAI_HOOK_PERF_SKIP`·`GOFLAGS` 셋만 제거한 뒤 센티널을 붙인다. **요구가 이제 무는 것은 성질이다**: **음성 자식의 환경은 쓰기 게이트를 담지 않는다.** 그것이 이 요구가 존재하는 이유이며, 환경을 열거로 짜든 뺄셈으로 짜든 그것은 구현 선택이다. 종전 문면은 **모양을 요구로 못 박아 이식성 결함에 요구의 옷을 입혔다.** **뺄셈 방향을 계약이 승인하고, 허용목록으로 다시 조이는 것은 계약 위반으로 보이게 썼다** — 코드 주석에만 있으면 다음 사람이 “정리”로 되돌린다. 근거: 허용목록은 **아무도 열거하지 않은 변수를 어떤 플랫폼·툴체인이 필요로 하는 순간 조용히 실패한다**(Windows 는 첫 사례일 뿐 마지막이 아니다). 거부목록은 실제 위험을 이름으로 지목한다. `GOFLAGS` 는 자기 이유로 남는다 — `-short` 나 `-count` 를 실어 단언 밑에서 자식 호출 모양을 바꿔치기할 수 있다. **형제 파일의 같은 표현도 함께 고쳤다**(같은 문구가 이미 한 번 세 파일에서 틀렸던 전례): `plan.md` §B.2 음성 방향 표 셀과 §F 2단계, `acceptance.md` §D.3(d) 설명 산문과 AC-PFW-006 시나리오. §D.3(d) 행의 **명령과 기대값은 그대로 유효**하다 — 거부목록 구현에서도 음성 자식은 게이트를 받지 않으므로 결과가 같다. 다만 산문이 “상속이 위험”으로 읽히던 것을 **“막아야 할 것은 상속이 아니라 뺄셈 없는 상속”**으로 바로잡았다. 0.3.0 행의 “자식 환경을 상속하지 않고 명시적으로 구성한다”는 **그 판의 요구를 정확히 기술한 기록이므로 고치지 않는다** — 이 행이 그것을 대체한다 |
| 0.3.1 | 2026-08-29 | manager-spec | **0.3.0 의 limb 삭제가 남긴 죽은 상호참조 3건 수리 — 계약 문면만 고쳤고 어떤 요구의 뜻도 바꾸지 않았다. 요구 8 / 수락 7 불변**(범위 증식을 찾는 독자는 이 문장에서 멈춰도 된다). **(1) 사라진 limb 를 가리키던 `REQ-PFW-007(d)` 를 `(c)` 로 재표기**했다 — `plan.md` §D 제약과 `spec.md` §C Out of Scope 불릿 두 곳. 뒤엣것은 이번 스윕이 찾은 **네 번째 지점**으로 보고서 목록에 없었다. 이제 세 파일의 limb 참조는 `(a)` 3회 · `(c)` 4회이며 전부 실재하는 항을 가리킨다. **(2) `acceptance.md` §D.3 의 행 표기를 `(a) (b) (c1) (c2) (d)` 로 바꾸고 개수를 5행으로 맞췄다** — 틀린 것은 라벨이 아니라 **개수**였다(`AC-PFW-006` 이 "(a)-(d) 4항"이라 적는 동안 표에는 다섯 행이 있었고, 구현자는 다섯 개를 다 돌렸다). 다시 합쳐지지 않도록 상시 메모를 함께 넣었다: **행은 다섯, 항은 셋.** `(c1)`/`(c2)` 는 REQ-PFW-007(c) 한 항이 두 탈출구(`MOAI_HOOK_PERF_SKIP` · `-short`)라 명령 두 개로 갈린 것이고, **`(d)` 는 REQ-PFW-007 의 항이 아니다** — REQ-PFW-005 의 자식 환경 격리를 보는 별개 검증이며, 같은 AC 가 두 요구를 덮기 때문에 한 표에 있을 뿐이다. 오독의 출처(0.3.0 limb 삭제 잔재)도 메모에 적었다. **(3) 가드 파일 경로를 `REQ-PFW-007(a)` 에 함수명과 나란히 못 박았다** — 종전에는 `§D.3(a)` 의 grep 인자 안에만 있었다. **grep 인자를 통해 요구 지위에 오른 값은 아무도 합의한 적 없는 요구다.** 구현자가 그대로 따랐다는 사실이 합의를 만들지는 않는다. 실측으로 확인: 구현된 파일은 `internal/hook/perf/report_write_guard_test.go` 이고 함수는 `TestPerfReportWriteGuard` — 못 박은 값과 일치하므로 새 모순은 생기지 않았다. **같은 판에 비용 정정이 함께 들어왔다.** §G.1 의 추정이 실측으로 교체됐다(재측정 명령 먼저, 값은 날짜 붙은 참조값으로 뒤에). 참조값 2026-08-29, HEAD `15453140a`: 구현자 측정 가드 제외 10.139s / 포함 30.494s(증가분 ≈20.4s), 조정자 독립 재측정 10.765s / 32.146s(≈21.4s). **두 값의 벌어짐은 기계 부하이며, 자릿수가 같다는 것이 판독의 전부다** — 그래서 문턱이 아니라 참조값이다. §B.3 표의 “~2분” 셀도 실측 증가분으로 교체했다. **폐기된 추정 2건은 방향이 아니라 종류로 기록한다**: “~2분”(인용된 “~30s” 서술에 자식 2회를 곱함 — 실측의 약 6배)과 “~13s”(배차 추정, 자식 1회 가정 — 약 절반)는 반대 방향으로 틀렸지만 같은 잘못이다 — **인용된 산문 수치에 곱셈을 해서 측정값 자리에 놓은 것.** 반복하지 않을 것은 그 종류다. 이 판은 `status:` 를 건드리지 않는다 — `spec.md` 에만 있고 `plan.md`·`acceptance.md` 에는 없는 현 상태가 정답이며(iter-1 D5 가 금지 필드로 삭제했다), 형제 파일에 되살리는 것은 그 수리를 되돌리는 일이다 |
| 0.3.0 | 2026-08-29 | manager-spec | plan-audit iter-2(PASS-WITH-DEBT 0.88) 차단 2건 수리 + 선택 4건 반영. **D11(major) — REQ-PFW-007 limb (c) 는 발화할 수 없는 조항이었다.** `-short` 는 플래그이지 환경변수가 아니라 가드가 스스로 짜는 자식 명령줄에 애초에 없고, `MOAI_HOOK_PERF_SKIP` 은 limb (d) 가 먼저 가드를 스킵시켜 자식 구성 시점에 도달하지 못한다. 검증도 `perf-guard: child env scrubbed` 문자열 grep 이라 **아무것도 스크럽하지 않고 그 줄만 찍는 구현이 통과**했다 — 관측이 아니라 주장이었다. **구제가 아니라 삭제로 고쳤다**: limb (c) 를 없애고 REQ-PFW-007 을 3항으로 줄였으며, `spec.md`·`plan.md` 의 거짓 근거 문장(“외부 스킵이 음성 방향을 공허하게 초록으로 만든다”, “(c)와 (d)는 둘 다 필요하다”)을 삭제하고 근거를 **(d)**(현재 (c))에 실었다 — 상속된 `MOAI_HOOK_PERF_SKIP` 은 가드를 **공허하게 만드는 것이 아니라 부재하게** 만들며, 비용을 내지 말라는 운영자 요청에 대해 그것이 정직한 결말이다. **도달 가능한 잔여 1건은 경로와 함께 보고하고 다른 요구로 옮겼다**: 부모가 `MOAI_HOOK_PERF_UPDATE=1` 을 세운 채 가드가 도는 경로는 (d) 가 막지 않으며, 이 패키지의 기존 관용구(`harness_test.go:243` `cmd.Env = append(os.Environ(), …)`)를 그대로 베끼면 게이트가 **음성** 자식으로 새어 가드가 회귀가 아닌 이유로 붉어진다 → REQ-PFW-005 에 “자식 환경을 상속하지 않고 명시적으로 구성한다”를 추가하고 §D.3(d) 로 행동 관측한다. **D12(minor)**: AC-PFW-005 가 **끝 상태만** 재어, 자식이 죽어 아무것도 쓰이지 않은 채 붉어진 경우에도 세 기대가 모두 충족됐다 → REQ-PFW-003 에 귀속 조항(자식 rc·결합 출력 표면화, 해시 불일치 전용 문구)을 넣고 AC-PFW-005 에 **붉어진 이유**를 못 박는 기대 2건을 추가했으며, AC-PFW-003 의 되돌림(초록) 다리에 그 문구의 **부재** 단언을 붙여 “항상 찍는” 구현도 막았다. **D13(minor) — `tier:` 는 바꾸지 않는다.** 3-산출물이 Tier M 모양이라는 지적은 맞으나, 레인이 자기 카드의 tier 를 조용히 바꿔 문턱을 맞추는 것은 값싸서는 안 되는 수다. `tier: S` 를 유지하고 **선언된 부채**로 `plan.md` 서두에 기록했다(요구 8/수락 7 은 상한 안, 산출물 3 은 밖, 이번 감사는 더 엄격한 0.80 으로 채점, 재분류는 운영자 결정). **D14(minor)**: 0.2.0 행의 과장 2건을 그 행에서 정정 — 매치 수 단언은 “모든 `-run` 행”이 아니라 실행 행은 `=== RUN`·스킵 행은 `SKIP:`, D6 는 재작성이 아니라 흡수. **D15(minor)**: D9 철회가 순차 실행을 전제하므로 “가드는 `t.Parallel()` 이 아니다”를 REQ-PFW-006 조항으로 못 박았다. **D16(기록만)**: `scripts/ci-mirror/lib/go.sh:25` 가 `-short` 를 붙여 로컬 CI 미러가 가드를 건너뛴다 — 미러 충실도 문제이며 다른 카드 소관, `plan.md` §G.1 에 기록만 |
| 0.2.0 | 2026-08-29 | manager-spec | plan-audit iter-1(PASS-WITH-DEBT 0.84) 차단 등급 6건 수리. **D1(major) — 내용 해시 근거가 거꾸로 적혀 있었다.** 종전 §A.6 은 "mtime 을 키로 삼은 가드는 mtime-only 뮤턴트를 **통과시키므로** 공허하다"고 적었으나 반대다: mtime 가드는 그 뮤턴트에서 **붉어진다**(잡는다). 놓치는 쪽은 내용 해시 가드다. 결정은 살아남고 **근거만 바뀐다** — **git 이 재는 것은 내용이지 mtime 이 아니다.** 바이트가 그대로인 파일은 더럽지 않고 `git status` 에 뜨지 않으며 커밋에 쓸려 들어갈 수 없다 → 이 SPEC 이 막으려는 해악이 아니다. 아울러 **t256 의 mtime-only 뮤턴트는 이 SPEC 의 어떤 기준으로도 탐지되지 않으며, 그것이 결함이 아니기 때문에 의도적으로 그렇게 두었다**는 사실을 §A.6 에 명시했다(종전에는 `plan.md` §G 에만 있었다). 세 파일의 서술을 하나로 맞췄다. **D10(major)**: 복원 단언이 통과 경로에서만 검증됐다 — 가드가 **자기 단언이 실패한 채 끝난 직후, 뮤턴트를 되돌리기 전에** 픽스처가 이미 복원돼 있음을 관측하는 기준을 신설(AC-PFW-005). **D3(major)**: `-short`·외부 `MOAI_HOOK_PERF_SKIP` 상속 구멍이 산문에만 있었다 → REQ 로 승격(자식 환경 스크럽 + 로그, 가드 자신의 자가 스킵)하고 검증 AC 를 붙였다. **D4(major)**: 가드의 테스트 함수명을 REQ 로 못 박고(`TestPerfReportWriteGuard`), `-run` 으로 좁힌 행에 매치 수 단언을 추가했다 — 실행을 세는 행은 `=== RUN` 수, 스킵을 단언하는 §D.3 행은 `SKIP:` 수로 (0.3.0 에서 표현 정정, D14)(되돌림 다리가 0-매치로 공허 통과하던 구멍). **D7(major)**: Tier S 예산 초과(요구 10/상한 8) → 운영자가 채택한 절단 3건 적용 — AC-PFW-003(불변 동작 단언)·REQ+AC-PFW-004(선언 형태를 규정하는 HOW 요구)·AC-PFW-008 의 `.gitignore` 다리(공허)를 삭제. 요구 **8**, 수락 **7** 로 상한 안. **D5(minor)**: `plan.md`·`acceptance.md` 프론트매터의 금지 필드 `status:` 삭제. 선택 3건도 반영 — **D6** `With` 로 열던 요구를 REQ-PFW-001 의 후반절로 흡수(재작성이 아니라 흡수 — 0.3.0 에서 표현 정정, D14), **D8** `UPDATE_GOLDEN` 실측 정정(6개 파일 / 4개 패키지), **D9** 가드 파일명 정렬 의존성을 `plan.md` §F 에 명시 |
| 0.1.0 | 2026-08-29 | manager-spec | 최초 작성. 결함·소비자 분석은 배차 시점에 이미 끝나 있었고, 이 판은 그 결과를 요구사항으로 옮긴다. 후보 처방 3안 중 (c) gitignore·(a) `t.TempDir()` 는 **기각**(둘 다 감사자가 읽는 추적 증거를 없앤다), (b) **명시적 opt-in 게이트 채택**. 조정자 실측 보충 2건 반영: `MOAI_HOOK_PERF_SKIP` 을 세우는 곳이 리포 전체에 없고 CI 가 `-short` 도 안 붙이므로 **CI 는 매 실행 두 테스트를 실제로 돌리고 자기 체크아웃의 추적 파일 2개를 다시 쓴다** — 이에 따라 `harness_test.go:24-25` 주석의 "CI 에서는 스킵된다" 서술은 거짓이며 정정이 이 SPEC 범위에 든다 |

---

## §A Context & Motivation

### A.1 결함 — 테스트가 추적 파일 2개를 무조건 덮어쓴다

`internal/hook/perf/harness_test.go` 는 실행할 때마다 **추적된** git 파일 2개를 조건 없이 덮어쓴다.

| 위치 | 대상 | 함수 |
|---|---|---|
| `harness_test.go:48-52` | `.moai/specs/SPEC-HOOK-PRETOOL-PERF-001/baseline.md` | `TestPreToolProfilingBaseline` |
| `harness_test.go:84-88` | `.moai/specs/SPEC-HOOK-PRETOOL-PERF-001/postchange.md` | `TestPreToolProfilingWarmCache` |

경로는 `projectRoot(t)`(`harness_test.go:378-386`)가 패키지 디렉터리 기준 `filepath.Abs("../../..")` 로 잡으므로, **테스트가 도는 그 트리**에 쓴다. 워크트리든 primary 든 가리지 않는다.

두 리포트는 각각 `report.markdown()`·`report.markdownPostChange()` 가 렌더하며 **둘 다 `**Timestamp**:` 행을 담는다**(`:147`, `:179`). 그러므로 측정 수치가 우연히 같아도 내용은 매 실행 달라진다 — 더러워지는 것은 조건부가 아니라 **무조건**이다.

`git ls-files` 실측: 두 파일 모두 추적 중이며, `.gitignore` 에 이들을 덮는 규칙은 없다.

### A.2 관측된 피해 — 레인 4개가 각각 손으로 되돌렸다

`go test ./internal/hook/perf/...` 를 돈 레인은 영구적으로 더러운 작업 트리를 얻고, `git add -A` 는 **남의 SPEC 픽스처**를 그 카드의 커밋에 쓸어 담는다. 독립 관측 4회(2026-08-24 lane-7·lane-8, 2026-08-27 t306 핸드오프·lane-15), 매번 손으로 되돌림.

### A.3 CI 도 매 실행 이 파일들을 다시 쓴다 (조정자 실측, 이 트리 `a6bbbf82b`)

주석 `harness_test.go:24-25` 는 두 테스트가 "normal CI 에서는 `MOAI_HOOK_PERF_SKIP=1` 로 스킵된다"고 적지만, **그 전제는 거짓이다**:

- `MOAI_HOOK_PERF_SKIP` 을 세우는 곳이 `.github/`·`Makefile`·`scripts/` 어디에도 없다.
- `.github/workflows/ci.yml:183` `go test -coverprofile=… ./...`, `:238` `go test -race -count=1 ./...` — 둘 다 `-short` 를 붙이지 않는다.

즉 CI 는 매 실행 두 프로파일링 테스트를 **실제로 돌리고** 자기 체크아웃의 추적 파일 2개를 다시 쓴다. 이 주석은 같은 doc comment 안에 있어 run-phase 가 어차피 손대는 자리이므로, **거짓 서술 1줄의 정정은 이 SPEC 범위에 든다**(드라이브-바이 편집이 아니다 — `plan.md` §D.1 에 명시).

### A.4 소비자 분석 — 처방이 여기서 결정된다

조사 범위를 먼저 못 박는다(부재 주장의 경계): `internal/`·`pkg/`·`cmd/` 아래 Go 소스, 리포 전역 파일명 참조, `.github/`·`scripts/`·`.claude/`·`Makefile`, `.gitignore`.

- **기계 판독자: 0건.** 두 파일을 파싱하는 코드는 없다.
- **사람 판독자: 3줄**, 전부 `SPEC-HOOK-PRETOOL-PERF-001/progress.md` — `:26`(M0 GATE 근거로 `baseline.md` 인용), `:29`(M3 검증 근거로 `postchange.md`), `:40`(AC-PERF-006 PASS 근거로 `postchange.md`). 더해 `plan.md:79`·`:114` 가 둘을 run-phase 증거 문서로 규정한다.

따라서 후보 처방 3안의 판정:

| 안 | 판정 | 근거 |
|---|---|---|
| (a) `t.TempDir()` 로 리다이렉트 | **기각** | 증거가 테스트와 함께 사라진다 — 감사자가 M0/M3 판정 근거를 잃는다 |
| (b) 명시적 opt-in 갱신 게이트 | **채택** | 커밋된 증거는 남고, 재생성은 의도적으로 요청할 때만 일어난다 |
| (c) `.gitignore` 추가 | **기각** | (a) 와 같은 이유 — 추적 감사 증거를 버린다 |

### A.5 집의 관용구를 재사용한다 (단순성 사다리 2단)

리포에 이미 있는 패턴을 쓴다. 정본 형태는 `internal/cli/doctor_golden_test.go:14-15`:

```go
// updateDoctorGolden controls golden snapshot regeneration. Set via UPDATE_GOLDEN=1.
var updateDoctorGolden = os.Getenv("UPDATE_GOLDEN") == "1"
```

실측: 같은 형태가 **6개 파일 / 4개 패키지**에 이미 있다(`internal/tui` 1, `internal/tui/golden` 1, `internal/cli` 3, `internal/cli/uikit` 1). 재유도 명령은 `grep -rn 'os.Getenv("UPDATE_GOLDEN")' internal` 이며 수는 얼리지 않는다.

**이름은 `UPDATE_GOLDEN` 을 재사용하지 않고 `MOAI_HOOK_PERF_UPDATE` 를 쓴다.** 근거 한 문장: `UPDATE_GOLDEN` 은 `Makefile:136-137` `tui-snapshot` 타깃이 이미 쓰는 **살아 있는 이름**이라, 재사용하면 결정적 골든 갱신 한 번이 ~30초짜리 벤치마크까지 함께 돌려 기계 의존적인 타이밍 수치를 추적 증거에 써 넣는다. 새 이름은 같은 패키지의 기존 접두 가족(`MOAI_HOOK_PERF_SKIP`, `MOAI_HOOK_PERF_TIMING`)과도 맞는다.

선언 형태 자체(패키지 수준 `var` + `os.Getenv(…) == "1"`)는 **요구가 아니라 권고**다 — 구현 형태를 규정하는 것은 WHAT 이 아니라 HOW 이므로 `plan.md` §A.2 가 가이던스로 담는다.

### A.6 회귀 가드는 왜 내용 해시를 키로 삼는가 — 그리고 무엇을 일부러 못 잡는가

**git 이 재는 것은 내용이지 수정 시각이 아니다.** 바이트가 그대로인 파일은 더럽지 않다 — `git status` 에 뜨지 않고, `git add -A` 가 집어 가지 않으며, 남의 카드 커밋에 쓸려 들어갈 수 없다. 이 SPEC 이 없애려는 해악은 정확히 그 **쓸려 들어감**이므로, 가드의 단언은 해악과 같은 축, 곧 **내용**에 걸어야 한다. 수정 시각을 키로 삼으면 축이 어긋나 **아무 해도 끼치지 않는 no-op 재작성까지 붉게** 만든다.

**흡수 카드 t256 이 남긴 대표 뮤턴트 — 내용은 그대로 두고 mtime 만 갱신하는 구현 — 은 이 SPEC 의 어떤 기준으로도 탐지되지 않는다. 빠뜨린 것이 아니라 그렇게 두기로 한 것이다.** 그 뮤턴트가 만드는 상태는 **결함이 아니기** 때문이다(바이트가 같으니 트리는 깨끗하고, 쓸려 들어갈 것이 없다). 방향을 분명히 적어 둔다 — 그 뮤턴트에서 **붉어지는** 쪽은 mtime 가드이고, **통과시키는** 쪽은 내용 해시 가드다. 여기서 "mtime 가드가 더 약하다"는 결론이 나오지 않는다: 잡히는 것이 해악이 아니므로 그 RED 는 오탐이다.

나중에 이 가드를 mtime 기준으로 "강화"하려는 사람은 위 두 문단을 근거로 **하지 않는다**. 강화가 아니라 오탐원의 도입이다.

---

## §B Requirements (GEARS)

**REQ-PFW-001 (Ubiquitous — shall not):** The perf harness tests shall not write `baseline.md` or `postchange.md` unless the opt-in update gate `MOAI_HOOK_PERF_UPDATE=1` is set; the default execution path shall leave both tracked files byte-identical while still executing the profiling run and still emitting the rendered report through `t.Log`.

**REQ-PFW-002 (Where):** Where `MOAI_HOOK_PERF_UPDATE` is set to `1`, `TestPreToolProfilingBaseline` and `TestPreToolProfilingWarmCache` shall write their reports to the same two paths and with the same rendered content the current implementation produces.

**REQ-PFW-003 (Ubiquitous):** The regression guard shall decide the negative direction by comparing a **SHA-256 content hash** of each fixture taken before and after a gate-off child invocation, and shall not key that decision on modification time, file size, or existence alone; and it shall make that decision **attributable** — the negative child's exit code and combined output shall be surfaced in the guard's own report, a hash mismatch shall be announced by a distinct `perf-guard: fixture content changed` line, and a non-zero child exit code shall be reported as a distinct failure from a hash mismatch, so a RED guard is never ambiguous between "the write happened" and "the child never ran".

**REQ-PFW-004 (When):** When the guard performs a second child invocation with `MOAI_HOOK_PERF_UPDATE=1`, it shall assert that both fixtures' content hashes **differ** from the captured originals, so a "never writes at all" regression cannot pass on the negative direction alone.

**REQ-PFW-005 (Ubiquitous):** The gate-off child invocation shall use the CI invocation shape — the package glob `./internal/hook/perf/...` without `-short` — rather than a hand-narrowed `-run` selector, because that is the shape under which CI rewrites the fixtures today (§A.3); and **the negative child's environment shall not carry the write gate `MOAI_HOOK_PERF_UPDATE`**, so a value set in the parent cannot reach the negative leg and turn the guard red for a reason that is not a regression. That property shall be satisfied by **subtraction — inheriting the parent environment and removing the named variables — and the guard shall not enumerate an allowlist of variables to pass**: an allowlist fails silently whenever a platform or toolchain needs a variable nobody enumerated (Windows `LOCALAPPDATA` / `TMP` / `TEMP` / `USERPROFILE` / `SystemRoot` was the first instance, not the only possible one), whereas a denylist names the actual hazard. The removed set shall also carry `GOFLAGS`, for a reason of its own: it can carry `-short` or `-count` and would reshape the child invocation out from under the assertions.

**REQ-PFW-006 (Ubiquitous):** The guard shall restore the captured original bytes of both fixtures on every **normal** exit path — including the path on which its own assertion fails, but not on a kill (`SIGINT`/`SIGKILL`) or a `go test -timeout` panic, where `t.Cleanup` does not run and the fixtures stay modified — so that a RED guard leaves the two tracked files byte-identical and never becomes a fresh instance of the defect it exists to detect; and the guard shall not be `t.Parallel()`, because that restore contract assumes its cleanup fires before any later test in the package begins.

**REQ-PFW-007 (Ubiquitous):** The guard's invocation hygiene shall hold on all three counts: (a) it shall live at exactly `internal/hook/perf/report_write_guard_test.go` and its test function shall be named exactly `TestPerfReportWriteGuard`, so neither the acceptance selectors nor the file-scoped acceptance command can silently match nothing; (b) it shall pass a sentinel `MOAI_HOOK_PERF_GUARD_CHILD=1` to each child and shall skip itself when that sentinel is present, so the package-glob child cannot recurse; (c) it shall skip itself when `testing.Short()` holds or when `MOAI_HOOK_PERF_SKIP` is set in its own environment, so the package's two existing escape hatches keep working and an operator who asked not to pay the cost does not pay more of it.

**REQ-PFW-008 (Ubiquitous):** The false claim in the `TestPreToolProfilingBaseline` doc comment (`harness_test.go:24-25`) that the tests are skipped during normal CI via `MOAI_HOOK_PERF_SKIP=1` shall be corrected to state what is measured: nothing in the repository sets that variable, CI passes neither it nor `-short`, and the new gate suppresses the write rather than the run.

---

## §C Scope Exclusions

명시적으로 이 SPEC 밖이다.

### Out of Scope — mtime-only 재작성의 탐지

- 내용을 바꾸지 않고 수정 시각만 갱신하는 구현은 **어떤 기준도 잡지 않는다**. §A.6 의 결정이며, 근거는 그 상태가 결함이 아니라는 것이다 — git 은 내용을 재므로 트리가 더러워지지 않고, 커밋에 쓸려 들어갈 것이 없다. 빠뜨린 것이 아니라 선택한 것이다.

### Out of Scope — 리포트 내용·형식의 변경

- `report.markdown()` / `report.markdownPostChange()` 의 렌더 결과, `**Timestamp**:` 행의 존재 여부, 표 구성 — 일절 손대지 않는다. 게이트는 **쓰기 여부**만 바꾼다.
- 커밋되어 있는 `baseline.md` / `postchange.md` 의 현재 본문 재생성 — 하지 않는다. 이들은 M0/M3 판정 시점의 증거다.

### Out of Scope — 프로파일링 자체와 기존 스킵 분기의 동작

- 두 프로파일링 테스트의 `-short` 스킵과 `MOAI_HOOK_PERF_SKIP` 스킵 분기는 **현행 유지**한다. REQ-PFW-007(c)는 **새 가드**가 같은 두 탈출구를 존중하도록 요구할 뿐, 기존 분기를 바꾸지 않는다.
- CI 를 스킵시킬지 여부. §A.3 은 사실 기록이지 CI 설정을 바꾸라는 요구가 아니다.
- parallelism·batches 값, 측정 방법론, `timing.go` 의 계측 경로.

### Out of Scope — 같은 형태의 다른 테스트 전수 조사

- 리포 전역에서 "테스트가 추적 파일을 쓰는" 다른 사례를 찾아 함께 고치는 일. 이 카드는 관측된 1건만 닫는다.

### Out of Scope — 골든 게이트 통합

- 기존 4개 패키지의 `UPDATE_GOLDEN` 게이트를 새 이름으로 통합하거나, 반대로 이 게이트를 `UPDATE_GOLDEN` 아래로 접는 일 — §A.5 가 기각한 방향이다.

---

## §D Acceptance Criteria

정식 AC 매트릭스는 `acceptance.md` 가 SSOT 다. 요약: **AC-PFW-001 … AC-PFW-007**(7건), Given-When-Then 3건, edge case 3건. 요구 8 / 수락 7 — Tier S 상한(각 8) 안이다.
