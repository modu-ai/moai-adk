---
id: SPEC-UPDATE-OWNERSHIP-REPAIR-001
title: "design — ownership verdict, provenance evidence, and manifest additions"
version: "0.1.0"
created: 2026-10-18
---

# SPEC-UPDATE-OWNERSHIP-REPAIR-001 — 설계 (design.md)

이 문서는 spec.md 가 정한 동작을 받치는 비자명한 데이터 모델 결정만 적는다. 함수 이름과 필드 이름은 run 단계의 제안이며, 요구사항이 아니라 근거를 설명하는 예시이다 (spec.md 는 함수 이름을 고정하지 않는다). 좌표는 HEAD `cdd72a814` 의 것이다.

## D-1 — 두 단계 분류: 소유권 분류와 처분

소유권 분류와 처분을 한 값으로 합치지 않는다. VCI 의 "분류와 처치는 별개 단계" 와 같은 이유이다: 분류가 곧 처치라고 믿는 독자는 처치의 종류를 분류의 종류만큼만 갖게 되어, 맞아 보이는 첫 번째 처치를 고른다.

```text
소유권 분류 (닫힌 집합, 4개)      처분 (2개)
  managed                          keep(<이유>)
  user-modified                    removal-eligible
  user-created
  unregistered
```

- `managed`: 매니페스트 항목이 있고 `Provenance` 가 `template_managed` (또는 `deprecated` — 현재 처리를 바꾸지 않는다), `TemplateHash` 가 기록되어 있고, 현재 바이트의 해시가 `TemplateHash` 와 같다. 미등록 파일이 증명되어 `managed` 가 되는 경우 (D-2) 도 같은 분류이며, 증거 기록이 붙는다.
- `user-modified`: 항목이 있으나 바이트가 `TemplateHash` 와 다르거나 `Provenance` 가 `user_modified`.
- `user-created`: `Provenance` 가 `user_created`.
- `unregistered`: 매니페스트 항목이 없고 증명도 되지 않음.
- 처분 `removal-eligible` 은 판정 함수가 말하는 "t1668 의 게이트 앞까지의 후보" 이다. 실제 제거는 `removalEnabled`, 백업 선행, 사용자 대응물 확인, 루트 경계 검사를 모두 지난 뒤에만 일어난다. 판정 함수는 제거하지 않는다.

이유 코드의 닫힌 집합 (제안): `user-modified`, `user-created`, `no-provenance-record`, `evidence-incomplete:<항목>`, `record-hash-mismatch`, `profile-reach`, `counterpart-unconfirmed`, `non-catalog`, `removal-gate-off`, `case-variant`, `symlink-escape`, `read-failed`, `no-template-hash`, `empty-dir-no-provenance`, `backup-failed`. 새 이유는 SPEC 개정 없이 늘리지 않는다 (잔존 목록이 열린 어휘가 되면 이유별 집계가 의미를 잃는다).

판정 함수는 순수하게 읽기만 한다: 파일시스템을 읽고, 매니페스트를 읽고, 사용자 매니페스트를 읽는다. 쓰기와 제거는 호출자 (마이그레이션) 가 한다. 그래서 같은 함수를 `--dry-run` 미리보기가 그대로 부를 수 있다 (REQ-OWR-011).

### 네 곳의 판정과의 관계

| 기존 판정 | 이 SPEC 에서의 처리 |
|-----------|----------------------|
| 마이그레이션 `migrate_project_assets.go:198-302` 의 인라인 순서 | 판정 함수 호출로 바뀐다. 처분은 현재와 같다 (특성화 시험 11개가 증거) |
| 미리보기 `update_dryrun_preview.go:120` `migrationRemovalCandidate` | 판정 함수 호출로 바뀐다 (REQ-OWR-011) |
| reconcile `update/reconcile_classify.go:282` | 바뀌지 않는다. 다루는 루트와 어휘가 다르다. 일치 규칙만 둔다: 한쪽이 보존으로 두는 경로를 다른 쪽이 제거 처분으로 보내지 않는다 |
| 휴면 분류기 `update/migrate_classify.go:172` | 바뀌지 않는다. 비시험 호출자가 없다 (research.md C). 정리는 이 SPEC 의 범위가 아니다 |

reconcile 의 `ClassUserModified` 는 "템플릿 운반 + 기록 없음" 도 포함하므로 (`reconcile_classify.go` 의 템플릿 운반 분기), 판정 함수의 `unregistered` 와 이름이 다르지만 둘 다 보존 쪽이다. 어휘를 합치지 않는 이유는 두 분류기가 다른 루트에서 다른 처분 (병합/충돌 sidecar vs 제거 후보) 을 정하기 때문이다.

## D-2 — 증거 세 가지와 두 원천

`managed` 로 인정받으려는 **미등록** 파일은 세 증거를 모두 가져야 한다 (REQ-OWR-002). 증거는 파일 자신의 바이트에서 나온 것이면 안 된다.

| 증거 | 뜻 | 확인 방법 |
|------|-----|-----------|
| 배포 버전 | 어느 moai 바이너리가 이 경로를 배포했는가 | 이전 세대 매니페스트 항목의 `deployed_version` (D-3) |
| 원래 배포 경로 | 배포 때 이 파일이 놓인 경로 | 이전 세대 매니페스트의 키가 현재 경로와 같음 (같은 철자, D-4 의 별칭 포함). 같은 해시의 기록이 다른 키에만 있으면 `evidence-incomplete:original-deployed-path` |
| 이전 매니페스트 기록 | 이 경로가 그때 `template_managed` 로 기록되었음 | 이전 세대 매니페스트 항목: `Provenance == template_managed`, `TemplateHash` 기록 |

세 증거를 모두 읽은 뒤에도 하나가 더 있다: 디스크 바이트의 해시가 이전 기록의 `TemplateHash` (또는 `DeployedHash`) 와 같아야 한다. 이는 증거가 아니라 일치 확인이다 — 해시 일치는 필요조건이지 충분조건이 아니다 (REQ-OWR-002). 일치하지 않으면 이유는 `record-hash-mismatch`.

원천은 닫힌 두 가지이다.

- **S1 — 실행 시작 시점의 매니페스트.** `runTemplateSyncWithReporter` 가 Backup 단계 앞에서 적재한 매니페스트의 사본 (`update_template_sync.go:285-286`). 마이그레이션이 공유 매니저를 바꾸기 전에 깊은 복사를 떠 두어야 한다 (판정이 자기가 바꾼 상태를 읽지 않도록). 미등록 파일은 정의상 S1 에 없으므로 S1 은 "등록된 파일의 기록" 을 확인하고, 미등록 파일의 증거 원천으로는 S2 만 남는다.
- **S2 — 보존된 이전 세대 사본.** 실행 백업 디렉터리의 `manifest.prior.json` (D-3). 최근 것부터 최대 `priorManifestGenerations` 개 (제안 기본값 8) 를 본다. 비용을 말뭉치 크기와 무관하게 상수로 묶기 위한 상한이다 (coding-standards 의 Advisory-Check Discipline: 시작 경로의 조회는 비용이 코퍼스에 비례하면 안 된다).

새 원천은 SPEC 개정 없이 추가하지 않는다. 특히 두 가지를 증거로 인정하지 않는다.

- 내장 템플릿 트리와의 바이트 일치 — 이 바이너리의 내장 트리가 출시 트리와 같다는 보장이 없다 (dirty 빌드 `ee981a932`, research.md D(1)).
- 정규 스킬 경로와의 기계적 관계 (`.agents/skills/<n>` ↔ `.claude/skills/<n>`) — 관계는 누가 그 복사본을 만들었는지를 증명하지 않는다 (D-5).

테스트 가능성: 증거 원천은 인터페이스 (`provenanceSource`) 로 두어 시험이 이전 세대를 직접 공급할 수 있게 한다. 이것이 AC-OWR-05 의 양성 픽스처와 변이 네 가지를 운영 경로와 독립으로 만든다.

## D-3 — 매니페스트 추가분과 호환 (decision-index.md Q2 (가) 에서만)

1. `FileEntry.DeployedVersion string` — JSON `deployed_version,omitempty`. `Track` 이 기록할 때 배포 바이너리의 버전을 찍는다 (`pkg/version.GetVersion()`). 파일이 실제로 쓰일 때만 `Track` 이 불리므로 (`deployer.go:294`) 값은 "그 바이트를 마지막에 쓴 버전" 이다. 이것이 필요한 이유: 혼합 버전 상태가 실재한다 (사용자 수정으로 갱신이 건너뛰어진 파일은 매니페스트 수준 버전보다 오래된 버전의 바이트를 가진다).
2. 매니페스트 수준 `Version` 은 갱신의 배포 단계 저장에서 배포 버전으로 갱신한다 (`recordTemplateTree` 와 같은 자리, `update_template_sync.go:672-673`). `DeployedAt` 은 `Version` 이 바뀔 때만 바꾼다 — 같은 버전의 반복 갱신이 매니페스트 바이트를 바꾸면 REQ-OWR-012 (멱등) 가 깨진다.
3. 실행 백업 디렉터리에 `manifest.prior.json` 을 쓴다: Backup 단계가 시작 시점 매니페스트의 원본 바이트를 그대로 복사한다 (재직렬화하지 않는다 — 바이트 동일을 위해). 쓰기 실패는 경고이며 실행을 막지 않는다 (증거가 줄 뿐 데이터를 잃지 않는다). 단 이 파일이 있어야 제거된 항목의 복원이 소유권 정보를 함께 되돌린다 (REQ-OWR-010).
4. 호환: `deployed_version` 이 없는 기존 JSON 은 그대로 적재되고 저장 시 필드를 지어내지 않는다. 증거 부족은 미등록 파일의 `managed` 승격만 막는다 — 등록된 항목의 기존 판정 (`TemplateHash` 와 바이트 비교) 은 바뀌지 않는다. 이전 바이너리가 새 매니페스트를 읽으면 알 수 없는 필드를 무시하고, 다시 저장하면 필드가 사라진다. 그 경우 이후 증거는 사라진 만큼 약해질 뿐 보존 방향의 오류는 만들지 않는다 (증거 없음 = 보존).

Q2 가 (나) 로 풀리면 1-3 은 없고 4 만 남는다.

## D-4 — 대소문자 변형: 모드 프로브 없이 별칭 확정

걷기가 돌려준 경로 `rel` 이 매니페스트 키와 정확히 같으면 그대로 쓴다. 정확한 키가 없을 때만 별칭을 찾는다.

1. 같은 디렉터리의 매니페스트 키 중 `strings.EqualFold(key, rel)` 인 후보를 모은다.
2. 후보마다 `os.Stat(root/rel)` 과 `os.Stat(root/key)` 를 하고 `os.SameFile` 이 참이면 별칭이다. 후보가 정확히 하나일 때만 확정한다 (둘 이상이면 모호하므로 별칭 없음).
3. 대소문자 구분 파일시스템에서는 `root/key` 가 없거나 다른 파일이라 `SameFile` 이 거짓이다. 그러므로 파일시스템 모드를 따로 판정할 필요가 없다 — 판정 근거가 관측된 동일성 자체이다.

`probeCaseSensitiveFS` (`update_cleanup.go:387`) 는 프로젝트 루트에 프로브 파일을 쓴다. dry-run 은 아무것도 쓰면 안 되므로 (REQ-OWR-011) 판정 경로에서 쓰지 않는다.

시험 가능성: 별칭 확정 함수가 `stat` 과 `sameFile` 을 인자로 받게 하여 대소문자 구분/무구분 두 모드를 한 시험에서 주입으로 단언한다 (AC-OWR-11). 실제 호스트 시험은 호스트의 관측에 맞는 기대를 단언한다. 시험을 건너뛰지 않는다.

별칭은 이름을 바꾸지 않는다. 철자 정규화는 `skillresidue` 의 보류된 이름 바꾸기가 소유한다 (decision-index.md Q4 (가)). 마이그레이션 걷기 시점에는 그 이름 바꾸기가 이미 끝났을 수도 아닐 수도 있으므로 (`update_template_sync.go:369` 가 Backup 단계 앞), 판정은 두 경우를 모두 같은 파일로 읽는다.

## D-5 — 거울 복사본은 왜 증명할 수 없는가, 그리고 미뤄 둔 선택지

복사 폴백 거울은 의도적으로 매니페스트에 등록되지 않는다 (`internal/template/skill_mirror.go:1-21` 제약 2, `deployer.go:310-318`). 따라서 `.agents/skills/<n>/...` 복사본에는 "배포 때 이 경로에 기록되었음" 이라는 이전 기록이 구조적으로 없다. 정규 스킬 경로 (`.claude/skills/<n>/...`) 의 기록과 바이트 일치는 있어도, 그것은 복사본을 누가 만들었는지가 아니라 내용이 같다는 것만 말한다. 사용자가 폴더를 복사해 두었어도 같은 모양이 된다. 그래서 313개는 증거 없이 `unregistered` 이다 (REQ-OWR-003).

후속 후보 (이 SPEC 이 하지 않는다): 배포기가 복사 폴백 거울을 만들 때 경로, 원천 경로, 해시, 배포 버전을 별도 원장에 기록하면 그 파일들이 증거를 가진다. 이는 배포기 변경이며 매니페스트 등록 제약 (EISDIR 해싱 문제) 과 별개로 설계해야 한다.

## D-6 — 빈 디렉터리의 출처

정리 후보: 마이그레이션 루트 아래의 비어 있는 디렉터리 (`os.ReadDir` 결과가 0 개). 루트 자체와 에이전트 디렉터리 (`.claude/agents/moai`, `.codex/agents/moai`) 는 후보가 아니다 (REQ-OWR-008 은 스킬 디렉터리(`.claude/skills/` 또는 `.agents/skills/` 아래)만; `pruneEmptiedSkillDir` 의 경계 규칙과 같다).

"관리 파일만 있던" 의 판정: S1 과 S2 의 기록 중 그 디렉터리 아래에 기록된 모든 항목을 모은다. 그 집합이 비어 있지 않고, 모든 항목이 `template_managed` 이고, 모든 항목의 파일이 디스크에 없으면 출처가 있다. 하나라도 `user_created`/`user_modified` 이면, 또는 집합이 비어 있으면 (기록 없음) 출처가 없다 — 유지하고 `empty-dir-no-provenance`.

중첩: 아래쪽부터 위로 한 번씩 평가한다. 하위 폴더가 정리되어 부모가 비게 되면 부모도 같은 규칙으로 평가한다 (부모의 기록도 있어야 한다). `os.Remove` 는 빈 디렉터리만 지우므로 사용자 파일이 있는 디렉터리는 검사를 우회해도 남는다.

t1668 의 `pruneEmptiedSkillDir` (제거 직후 정확히 한 단계) 는 그대로 쓴다. 이 SPEC 의 정리 단계는 걷기 뒤에 도는 별도 단계이며, 제거 스위치와 무관하다 — 파일을 지우지 않고 빈 디렉터리만 지우기 때문이다.

## D-7 — 잔존 행과 프로필 대응물 상태

잔존 행: `{rel, ownership, reason, missing []string, counterpart, profileSide}`. `missing` 은 `deployed-version`, `original-deployed-path`, `prior-manifest-record` 중 빠진 것 (미등록 파일만). `counterpart` 는 `confirmed`/`absent`/`unconfirmed`/`n/a`, `profileSide` 는 프로필 지도 프로젝트에서 항상 `not-checked` 이고 그 외에는 `n/a`.

정렬: `rel` 사전순 (결정적). 렌더: 이유별 개수 한 줄이 기본이고 `--verbose` 에서 행마다 한 줄 (운영자 ruling d-20261012T221654Z-9ced 와 같은 형태). 마이그레이션이 행을 모아 돌려주고 (`migrationRunContext` 에 필드를 더하는 방식이 기존 `removed`/`keptProfileReach` 와 같은 결), 갱신 흐름이 터미널 참조 행으로 낸다 (decision-index.md Q3 (가)).

프로필 대응물: 사용자 루트의 대응물만 `userCounterpartConfirmed` 로 확인한다 (`migrate_project_assets.go:377-452`). 프로필 쪽은 읽지 않으므로 `not-checked` 로 기록한다 (REQ-OWR-005, decision-index.md Q5). `KeptBy` 표지는 Q1 의 DECIDED 대로 그대로 쓴다 — 표지의 뜻을 바꾸지 않는다. 표지를 읽어 사본을 되찾는 소비자 (후속 카드 t1642) 는 되찾기 시점에 대응물을 다시 확인해야 한다. 이 SPEC 은 소비자를 만들지 않는다.
