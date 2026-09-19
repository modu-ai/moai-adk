---
id: SPEC-DOCTOR-PLUGIN-DIGEST-PREFILTER-001
title: "doctor Home Disk Usage — plugins 트리 sha256 앞에 (Size, Files) 사전필터를 둔다"
version: "0.1.0"
status: in-progress
created: 2026-09-19
updated: 2026-09-19
author: manager-spec
priority: P2
phase: "v3.1.5 target"
module: "internal/cli"
lifecycle: spec-first
tags: "doctor, home-disk, sha256, prefilter, test-budget"
tier: S
era: V3R6
related_specs: [SPEC-CI-DOCTOR-BIN-001]
---

# SPEC: plugins 다이제스트 사전필터

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-09-19 | manager-spec | 최초 작성 — 진단 카드 t962 의 verdict(`.moai/reports/t962/verdict.md`) 계열 A 실측 위에 구성. 처방은 배차 시점에 이미 확정돼 있었고 본 SPEC 은 그것을 부호화한다 |

---

## 1. 배경 — 무엇이 상한 없이 비싼가

`moai doctor` 는 매 실행마다 `checkHomeDisk`(`internal/cli/doctor_disk.go:71`)를 돈다. 그 안의
`findPluginHashClusters`(`:298`)는 `plugins` 카테고리가 비어 있지 않은 **모든** 프로파일에 대해
`treeDigest`(`:326`)를 부르고, `treeDigest` 는 정규 파일을 전부 열어 내용을 sha256 한다
(`io.Copy(h, file)`, `:356`). 상한이 없다 — 비용이 개발자 홈 트리의 크기에 그대로 비례한다.

t962 가 이 기계에서 실측한 값(그 카드가, 그 실행에서, 그 트리에서 잰 것):

| 대상 | 실측 |
|---|---|
| `~/.moai` | 14G · 148,478 파일 |
| `~/.claude` | 894M |
| `~/.moai` stat-only 전량 1회 | **1.37s** (더운 캐시) |
| `claude-profiles/*/plugins` | 1.7G · 43,929 파일 |
| 위 plugins 전량 **내용 읽기** | **5.73s** |
| 프로파일별 plugins 크기 | 6.3M · 6.4M · 6.5M · 782M · 970M — **전부 서로 다르다** |
| HOME 을 중화하지 않고 전체 체크 세트를 도는 `internal/cli` 테스트 | 27건, 각 **약 15초** |

마지막 두 줄이 이 SPEC 의 핵심이다. 이 기계의 다섯 프로파일은 `(Size, Files)` 가 전부 다르므로
**바이트 동일 클러스터가 존재할 수 없는데도** 43,929 파일을 전부 열어 해시한다. 즉 지금의 5.73초는
결과가 이미 결정돼 있는 계산에 쓰인다.

CI 는 이 비용을 원리상 보지 못한다 — CI 의 홈은 비어 있다. 그래서 이 결함은 개발자 기계에서만
나타나고, 나타나는 곳에서는 매 `doctor` 실행마다 나타난다.

## 2. 처방 — 순수 사전필터 (선택은 이미 끝났다)

바이트 동일한 두 트리는 **필연적으로** 같은 `(Size, Files)` 를 갖는다. 역은 성립하지 않는다
(같은 크기·개수인데 내용이 다를 수 있다 — 기존 회귀 가드가 바로 그 경우를 잠근다).

따라서 해싱 앞에 다음을 둔다: 프로파일을 `plugins` 의 `profileCategoryStat{Size, Files}` 로 묶고,
**2개 이상이 모인 그룹에 속한 프로파일만** 해시한다. 싱글턴은 해시하지 않는다.

이것이 출력을 바꿀 수 없는 이유가 두 방향으로 닫힌다:

- **놓침 없음** — 참인 클러스터의 구성원은 같은 `(Size, Files)` 를 공유하므로 반드시 필터를 통과한다.
- **날조 없음** — 통과한 후보는 여전히 해시돼 확인되므로, 필터가 클러스터를 만들어내지 않는다.

## 3. 요구사항 (GEARS)

**REQ-DPP-001** — Where a profile's `plugins` category stat is not shared by at least one other
profile, the doctor home-disk check shall not compute that profile's tree digest.

**REQ-DPP-002** — The cluster output of `findPluginHashClusters` shall be identical to its
pre-change output on every input: no cluster that the pre-change code reports shall be absent, and
no cluster that the pre-change code does not report shall appear.

**REQ-DPP-003** — The candidate selection shall be a separate pure function over the per-profile
category stats, returning the profile names that still require hashing in a deterministic sorted
order, so it is verifiable without instrumenting the hashing path.

**REQ-DPP-004** — The existing regression guard
`TestPluginHashDoesNotEquateSameSizeDifferentContent` shall keep passing unmodified: the two
profiles it declares both carry `Size: 4, Files: 1`, so both survive the prefilter and are still
hashed, and both of its directions (same-size-different-content must not cluster; byte-identical
must cluster) shall still hold.

**REQ-DPP-005** — The run-phase report shall state that the bytes-proportional term is removed and
that the file-count-proportional walk term remains, and shall not claim the check became
constant-time; no acceptance criterion shall assert wall-clock duration.

> REQ-DPP-005 의 근거는 두 가지 실측이다. (가) 이 저장소의 픽스처 헬퍼
> `writeHomeFixtureFile`(`internal/cli/doctor_disk_test.go:37-40`)은 의도적으로 **희소(sparse)**
> 파일을 만든다 — 주석이 그렇게 선언한다("truncated — sparse where the filesystem supports it").
> 큰 바이트 픽스처는 읽기 비용을 만들지 않으므로 크기-대-시간 비교는 공허해진다. (나) 벽시계 단언은
> 부하 의존이라 흔들린다. 따라서 수락은 **결정적인 후보 집합**으로만 진술한다.

## 4. 수락 기준 (Tier S — 인라인)

| REQ | 피복 AC |
|---|---|
| REQ-DPP-001 | AC-DPP-001, AC-DPP-003 |
| REQ-DPP-002 | AC-DPP-002, AC-DPP-004 |
| REQ-DPP-003 | AC-DPP-003 |
| REQ-DPP-004 | AC-DPP-004 |
| REQ-DPP-005 | AC-DPP-005 |

미피복 REQ 0건, 고아 AC 0건.

**AC-DPP-001 (음성 대조 — 해싱 0회)** — Given a per-profile stat map in which every profile's
`plugins` `(Size, Files)` pair is distinct from every other profile's, When the candidate function
is called, Then it returns an empty slice — no profile is a hashing candidate.

**AC-DPP-002 (양성 대조 — 해싱 유지)** — Given two profiles sharing one `(Size, Files)` pair,
When the candidate function is called, Then both profile names appear in the returned slice, so the
hashing and cluster-confirmation path still runs for them.

**AC-DPP-003 (혼합 + 결정성)** — Given three or more profiles of which exactly two share a
`(Size, Files)` pair and the rest are singletons, When the candidate function is called, Then the
returned slice contains exactly the two sharing profiles, in sorted order, and repeated calls on the
same input return the identical slice.

**AC-DPP-004 (출력 불변 — 비회귀)** — Given the fixture of
`TestPluginHashDoesNotEquateSameSizeDifferentContent` (two profiles, both `Size: 4, Files: 1`, first
with differing content then with identical content), When `findPluginHashClusters` runs on each
state, Then the same-size-different-content state yields zero clusters and the byte-identical state
yields exactly one `plugins sha256=` cluster with both profiles — 그 테스트 파일은 수정하지 않고
그대로 통과한다.

**AC-DPP-005 (보고 — 단언 아님)** — Given the run-phase completion report, When it describes the
effect, Then it states that the bytes-proportional term is removed while the file-count-proportional
walk term remains (이 기계에서 전량 1회 1.37초), does not claim constant time, and the test diff
contains no wall-clock duration assertion (검증: 추가된 테스트에서 `time.Since` / `time.Now()` 기반
경과시간 비교가 0건임을 grep 으로 확인).

## 5. 범위 밖 (exclusions)

### Out of Scope — 중복 전량 훑기의 통합
- `gatherHomeDiskReport` 가 홈 뿌리를 여러 번 훑는 구간(131 / 147 / 168 / 185 / 283 줄 계열)은
  이 카드가 건드리지 않는다. 파일 개수 비례 항은 그대로 남는다. **후속 카드 후보.**

### Out of Scope — `--check` 제외 경로
- 테스트가 `Home Disk Usage` 를 `--check` 로 빼는 경로(t962 처방 후보 2)는 이 카드의 소관이 아니다.
  **후속 카드 후보.**

### Out of Scope — go/ast 소스 스캔 가드
- 「필터 없이 전체 체크 세트를 도는가 AND HOME 을 중화하는가」 두 인자 판별식을 코드에 거는 가드
  (t962 갈래 (b) 의 플래그 값 추적)는 이 카드가 만들지 않는다. **후속 카드 후보.**

### Out of Scope — init 계열 테스트 실패
- t962 §E5 의 init 계열 26건 FAIL(툴체인 다운로드 · `t.TempDir()` 정리 `permission denied`)은
  원인이 다르고 이 카드가 고치지 않는다. **후속 카드 후보(t964 계열).**

### Out of Scope — 벽시계 성능 단언
- 성능 개선을 시간으로 단언하는 테스트는 만들지 않는다(REQ-DPP-005 근거).

## 6. 제약

- **영향 파일 2개**: `internal/cli/doctor_disk.go`(사전필터 함수 신설 + `findPluginHashClusters`
  내부 1개소 배선)와 `internal/cli/doctor_disk_test.go`(신규 테스트 추가만 — 기존 테스트 수정 금지).
  프로덕션 변경은 약 15줄 규모다. 이 범위를 넘으면 설계를 다시 본다.
- **`treeDigest` 는 손대지 않는다.** 해시 자체가 바뀌면 출력 불변 논증이 깨진다.
- **기존 클러스터 출력 계약 유지**: `Category` 는 `"plugins sha256=" + digest[:12]`, 프로파일 목록은
  정렬, 결과는 `Category` 로 정렬 — 전부 그대로다.
