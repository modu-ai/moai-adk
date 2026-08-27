---
id: SPEC-BINARY-LAG-VISIBILITY-001
title: "배포 지연 가시성 — 구현 계획"
version: "0.4.0"
status: in-progress
created: 2026-08-27
updated: 2026-08-28
author: manager-spec
priority: High
phase: "v3.1.4 target"
module: internal/cli, internal/hook, build
lifecycle: spec-anchored
tags: "deployment-lag, doctor, session-start, version-stamp, observability, fail-open"
tier: M
---

# SPEC-BINARY-LAG-VISIBILITY-001 — 구현 계획

마일스톤은 **되돌리기 어려운 결정 순서**로 배열했다. 위쪽일수록 바뀔 가능성이 크고 사람의 검토가 필요한 결정이며, 아래쪽은 기계적 작업이다.

---

## §A 맥락

`spec.md` §1이 실측 근거의 정본이다. 핵심만 재진술한다: **비교 로직은 이미 존재하고 정확하다**(`checkBinaryFreshness`, `internal/cli/doctor.go:495`). 결손은 도달이다 — 그것을 발화하는 유일한 경로가 사람이 손으로 `moai doctor`를 치는 것이고, 자동 호출 지점은 전수 조사 결과 0건이다. 따라서 이 계획에 **비교 알고리즘을 새로 쓰는 마일스톤은 없다.**

---

## §B 알려진 이슈 / 전제

1. `checkBinaryFreshness`는 현재 `os.Getwd()` 기준으로 동작하며 `verbose` 외 주입 지점이 없다 — 테스트에서 상태를 흔들 수 없다. M2의 seam 추출이 이를 푼다.
2. `doctorExitStatus`는 Fail 1건으로 exit 1을 만든다(`internal/cli/doctor.go:140`). 적용 불가를 Fail로 만들면 배포판 전체가 깨진다 — `acceptance.md` AC-BLV-003의 대표 뮤턴트가 바로 이것이다.
3. `computeDeferredAdvisory`(`internal/hook/session_start.go:593`)는 errgroup 태스크가 **자기 로컬 맵에만** 쓰는 계약이다. 지연 스캔 seam을 재사용한다면 그 동시성 계약은 지켜야 한다 — 다만 **판정 결과를 그 권고 맵에 남겨서는 안 된다.** 맵은 `:266`·`:574` → `:277` marshal → `:301` `HookOutput{Data:}`로 흘러가고 `Data`는 `json:"-"`이므로 직렬화되지 않는다. 빌려오는 것은 일정뿐이고, 결과는 `AdditionalContext`로 나와야 한다(`spec.md` §4 행 3 구속 조항).
4. `Makefile:6`의 `VERSION ?=`는 사용자 지정 가능(`?=`)하다. 릴리스 절차가 `VERSION=`을 명시로 넘기는 경로는 **보존**되어야 한다.

---

## §C 사전 점검 (착수 전 실행)

```bash
git rev-parse --short HEAD              # 계획이 세워진 트리와 일치하는지
git describe --tags                      # v3.1.2-494-g343399d2f 형태 확인
moai doctor --check "Binary Freshness"   # 현재 ok 인지 (회귀 기준선)
go test ./internal/cli/... ./internal/hook/... 2>&1 | tail -5   # 착수 전 GREEN 기준선
```

기준선이 이미 RED이면 이 SPEC과 무관한 실패이므로, 수리에 착수하기 전에 리드에 보고한다.

---

## §D 제약

- **로컬 전체 스위트 금지.** 영향 패키지만 돌리고 전 패키지 판정은 CI에 맡긴다(CLAUDE.local.md §4).
- **t317 워크트리에 쓰기 금지.** 다른 레인의 살아 있는 트리이며, 읽기 전용 스냅샷 인용만 허용된다.
- **신규 doctor 검사 이름 금지** — REQ-BLV-009 / AC-BLV-009 (spec.md 제약 C-2).
- 템플릿 아래(`internal/template/templates/`)를 건드리는 변경이 생기면 Template-First 규율과 `make build`가 따라붙는다. 현재 계획상 해당 없음이나, M1의 결정에 따라 훅 래퍼가 필요해지면 `.sh`/`.sh.tmpl` 쌍을 함께 고쳐야 한다(CLAUDE.local.md §2.3).

---

## §E 자기 검증

각 마일스톤 종료 시: 해당 AC를 GREEN으로 만들고, **대표 뮤턴트를 실제로 심어 RED를 관측한 뒤 원복**하며, 커맨드와 관측 출력을 `progress.md` §E.2에 기록한다. 뮤턴트를 심지 않은 GREEN은 근거로 채택하지 않는다 — 물려받은 테스트가 공허할 수 있기 때문이다.

---

## §F 마일스톤

### M1 — 권고 문면 확정 [되돌리기 가장 어려움 · 사용자 표면]

**표면은 결정됐다 (리드 결정 a): `additionalContext` 단독.** `systemMessage`도, 둘 다도 아니다. 근거와 선례 정정은 `spec.md` §1.7.1 — 요약하면 `HookOutput.Data`는 `json:"-"`라 도달하지 않고, `AdditionalContext`는 이 세션의 시작 컨텍스트에 실제로 도착한 것이 관측됐다.

[HARD] **선례는 「일정」에만 채택하고 「발화」에는 채택하지 않는다.** `computeDeferredAdvisory`는 지연 실행·최선 노력·경계된 조인·멱등 재유도의 선례이지, 결과를 어디에 쓰는지의 선례가 아니다. 이 구분을 놓치면 결손이 그대로 재생산된다(AC-BLV-008의 대표 뮤턴트).

이 마일스톤에 남은 결정은 **권고 문자열의 실제 문면**이다. 문면이 되돌리기 어려운 이유: 모든 세션 시작에 뜨는 문자열이므로, 소음이면 즉시 무시당하고 그 순간 이 SPEC 전체가 무의미해진다. 문면은 (a) 뒤처짐이 사실일 때만, (b) 두 SHA를 담아, (c) 수리 명령(`make build && make install`)을 지시해야 한다.

- 산출: 권고 문자열 확정 + 발화 조건 표
- AC: 없음(결정 마일스톤). M2가 이를 구현한다.

### M2 — 비교 seam 추출 + 두 표면 배선 [구조 결정]

`checkBinaryFreshness`의 판정 본체를 주입 가능한 단일 함수로 뽑고, doctor 항목과 세션 시작 권고 **양쪽**이 그것을 호출하게 한다. 적용가능성 술어(§5)는 이 함수 안에 산다.

- 기존 관용 다섯 갈래(커밋 메타데이터 없음 / cwd 불가 / git 아님 / HEAD 일치 / 조상 아님)를 **보존**한다. 축소는 배포판 회귀다.
- 기준점 해석은 t317 D9 선례를 따라 `.moai/` marker 상향 탐색을 채택한다 — doctor 배선이 `os.Getwd()` 원값을 넘기므로, 하위 디렉터리 실행에서 적용 가능한 트리가 적용 불가로 뒤집히지 않게 한다.
- 권고는 `AdditionalContext`에 **append-if-non-empty** 패턴으로 붙인다(`session_start.go:343-346`·`:369`의 확립된 형태 — 비어 있으면 대입, 아니면 `\n\n` 덧붙임). 기존 귀속 문자열·GLM 리마인더·팩토리 공지를 덮어쓰지 않는다.
- [HARD] doctor 항목은 **재배선만** 한다 — **세 레지스트리 어디에도** 새 이름을 등록하지 않는다(REQ-BLV-009): `systemChecks`(`internal/cli/doctor.go:187`), `moaiChecks`(`:195` 선언 ~ `:212`), `workspaceChecks`(`:214`). 셋은 `:245-249` → `:93-95` `allChecks`로 합류해 같은 exit 판정에 들어간다. t317도 `moaiChecks`에 등록할 예정이므로, 판정은 절대 개수가 아니라 **이 변경의 전후 이름 집합 합집합** 대조이며, 추출 단위는 따옴표 문자열이 아니라 **항목의 이름 표현식**이다(상수 식별자 회피 차단 — AC-BLV-009).
- AC: AC-BLV-001, AC-BLV-002, AC-BLV-003, AC-BLV-005, AC-BLV-008, AC-BLV-009

### M3 — 빌드 신원 단조성 수리 [사용자 표면 · `VERSION` 불변]

**결정됨 (운영자 결정 b): 별도 `BUILD_ID`를 도입하고 `VERSION`은 건드리지 않는다.**

근거 — 이 카드의 목적은 신호 도달이지 릴리스 넘버링 변경이 아니며, `VERSION`은 바깥으로 뻗는다: `Makefile:14`가 `RELEASE_BINARY`에 보간하고, `Makefile:35`가 `version.json`에 쓰며, `internal/update/local.go:65`가 그 파일을 읽는다. `VERSION`을 고정하면 릴리스 산출물 이름·업데이트 경로·GoReleaser가 이 카드의 폭발 반경 밖에 남고, 변경이 되돌리기 쉬운 상태로 유지된다.

- `BUILD_ID`는 `git describe --tags`(접미사 유지) 기반으로 파생하고, ldflags로 별도 심볼에 주입한다.
- [HARD] `Makefile:6`의 `VERSION ?=` 행은 **바이트 동일하게 남긴다** — AC-BLV-004 (c)가 이를 판정하며, `--abbrev=0`을 제거하는 구현은 AC-BLV-004의 대표 뮤턴트 1이다.
- 감수하는 비용: `moai version` 제목 줄은 여전히 `v3.1.2`로 읽힌다(`spec.md` §7.5 잔여 위험). 제목 줄 축은 **별도 카드**이며, 이 마일스톤에서 조용히 넓히지 않는다.
- AC: AC-BLV-004, AC-BLV-007

### M4 — 뮤턴트 대조 + 회귀 확인 [기계적]

`acceptance.md`의 대표 뮤턴트 전부(AC별 1~2종)를 순차로 심고 RED를 관측한 뒤 원복한다. AC-BLV-009는 처음부터 GREEN인 회귀 가드이므로 **뮤턴트 대조가 유일한 유효성 근거**다 — 반드시 심어 볼 것. 이어서:

```bash
go test ./internal/cli/... ./internal/hook/...
go vet ./internal/cli/... ./internal/hook/...
moai doctor            # 이 저장소에서 exit 0 유지
```

비-git 트리 확인은 **이 SPEC의 검사로 범위를 좁혀** 실행한다(감사 D8 — 전체 실행은 통제 밖 검사에 의존):

```bash
moai doctor --check "Binary Freshness"   # 비-git 디렉터리에서, exit 0 유지
```

> 감사 Gaps 기록: 감사관은 워크트리 격리 가드 때문에 트리 밖 디렉터리에서 실물 `moai doctor`를 돌리지 못했다. 위 좁힌 형태의 실측이 그 미관측 항목을 닫는다.

- AC: AC-BLV-006 + 완료 정의 전 항목

---

## §G 안티패턴

- **세 레지스트리 중 어디든 새 검사 이름 추가.** REQ-BLV-009 위반이며, t317과 같은 슬라이스를 두 레인이 동시에 건드리는 상황에서 경계를 무너뜨린다. AC-BLV-009의 대표 뮤턴트 1이다.
- **`workspaceChecks`로 우회 등록.** `moaiChecks`만 보는 판정을 통과하지만 요구는 그대로 위반한다 — 셋 다 같은 `allChecks`로 합류하기 때문이다. AC-BLV-009 뮤턴트 2.
- **상수 식별자 이름으로 우회 등록.** 따옴표 문자열만 추출하는 판정을 통과한다. AC-BLV-009 뮤턴트 3.
- **권고를 `systemMessage`에 쓰기.** 직렬화되므로 문서 전체 검색은 통과하지만 운영자 결정 (a)를 위반한다. AC-BLV-008 뮤턴트 2.
- **권고를 `HookOutput.Data`에 넣기.** `json:"-"`라 직렬화되지 않는다 — 올바른 판정을 아무도 렌더하지 않는 필드에 넣는 것이며, 이 SPEC이 닫으려는 결손의 재생산이다. `computeDeferredAdvisory`의 기존 권고 키들이 정확히 이 상태다(`spec.md` §1.7.1).
- **구조체 필드를 직접 읽어 「도달」을 주장하기.** `json:"-"` 필드도 통과시키므로 AC-BLV-008의 요점을 놓친다. 판정은 직렬화된 JSON 바이트에서 한다.
- **`VERSION` 파생을 고쳐 단조성을 얻기.** 운영자 결정 b 위반이며 `RELEASE_BINARY`·`version.json`·`internal/update/local.go`로 파급된다. 단조 신원은 `BUILD_ID`에 들어간다.
- **비교 로직 재작성.** `checkBinaryFreshness`는 정확하다. 새로 쓰면 이미 통과 중인 다섯 갈래 관용을 잃는다.
- **적용 불가를 Fail로.** 배포판 전체의 `moai doctor`가 exit 1이 된다.
- **`strings | grep`로 기능 존재 추론.** 이 세션에서 제안됐다가 기각됐다(spec.md §1.6). 새 바이너리에서도 0을 반환하는 공허한 판정이다.
- **semver 비교로 지연 판정.** 실측된 실제 역전에서 정확히 반대 결론을 낸다(AC-BLV-007).
- **grep으로 구현 개수를 세어 「단일 구현」 주장.** 문자열 개수는 대체 가능성을 증명하지 못한다(AC-BLV-005).
- **권고를 무조건 방출.** AC-BLV-002 대조군이 이를 잡는다. 소음이 되는 순간 이 SPEC은 무의미해진다.

---

## §H 상호 참조

- `spec.md` §6 — 세 고리 사슬 경계와 제약 C-1·C-2·C-3
- `acceptance.md` — AC 9건과 대표 뮤턴트
- `internal/cli/doctor.go:495` `checkBinaryFreshness` — 재사용 대상 (등록: `:199`)
- `internal/cli/doctor.go:140` `doctorExitStatus` — 배포판 회귀 위험의 근원 (세 레지스트리가 `:245-249` → `:93-95` `allChecks`로 합류해 여기로 들어온다)
- `internal/cli/doctor_mcp_version.go` — Link 3 선례(구조 + 「양성 증거에만 WARN」 규율)
- `internal/hook/session_start.go:243-257` — 호출자측 **기한 조인** 선례 (timer+select; AC-BLV-006)
- `internal/hook/session_start.go:593` `computeDeferredAdvisory` — **일정** 선례 (발화 선례 아님; 그 `:622-624` ctx-wrap은 기한 조인이 **아니다**)
- `internal/hook/types.go:394` `HookOutput.Data` `json:"-"` — 도달 불가의 구조적 근원
- `internal/hook/session_start.go:305` `AdditionalContext` — 도달이 실측된 표면 (append 패턴: `:343-346`, `:369`)
- `Makefile:6` `VERSION ?=` — 단조성 결손 지점 (**불변 유지**; 신원은 `BUILD_ID`)
- `Makefile:14` `RELEASE_BINARY` / `Makefile:35` `version.json` / `internal/update/local.go:65` — `VERSION` 소비 경로 (무변경 유지 근거)
- 카드 t317 / `SPEC-AGENT-EMIT-LINEAGE-001` REQ-AEL-004 — Link 2, 적용가능성 술어 추론의 출처(읽기 전용 스냅샷, 2026-08-27)
- 카드 t298 — 통합 락 수리(범위 밖, 반증 근거)
