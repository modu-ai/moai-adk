---
id: SPEC-UPDATE-OWNERSHIP-REPAIR-001
title: "design — ownership verdict, provenance evidence, and manifest additions"
version: "0.3.0"
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
- 처분 `removal-eligible` 의 정의는 spec.md § B 한 곳이다 — 이 문서는 인용만 한다. 판정 입력에 제거 스위치, 사용자 대응물 확인 결과, 프로필 도달 유지, 루트 경계 (심링크 탈출) 가 모두 들어 있다: 하나라도 빠지면 `keep(<이유>)` (스위치가 꺼진 기본값은 `keep(removal-gate-off)`) 이고, 판정 함수 바깥에 남는 것은 실행 시점에만 알 수 있는 조건 — 백업 사본 쓰기 성공, OS 읽기·삭제 오류 — 뿐이다. 그래서 `--dry-run` 미리보기가 알리는 후보 집합 (REQ-OWR-011) 과 판정의 `removal-eligible` 집합이 같고, 실제로 제거된 집합은 그 부분집합이다. 판정 함수는 제거하지 않는다. (종전 문장은 "t1668 의 게이트 앞까지의 후보" 였고 스위치·대응물을 판정 밖에 두어, 스위치가 꺼진 구성에서 acceptance.md 와 후보 집합이 어긋났다 — 감사 D8.)

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
| 이전 매니페스트 기록 | 이 경로가 그때 `template_managed` 로 기록되었음 | 이전 세대 매니페스트 항목: `Provenance == manifest.TemplateManaged` **정확히 이 값**, `TemplateHash` 기록. `UserModified`, `UserCreated`, `Deprecated`, `GeneratedManaged`, 빈 값, 알 수 없는 값은 기록이 있어도 증거가 아니다 (이유 `evidence-incomplete:prior-manifest-record`). 배포기가 매니페스트에 없는 기존 파일을 `UserCreated` 로 기록하고 건너뛰므로 (`internal/template/deployer.go:265-267`) 바이트가 템플릿과 같은 사용자 파일이 이전 세대에 `UserCreated` 로 남을 수 있다 — 해시 일치는 이를 `managed` 로 올리지 못한다 (감사 N2) |

세 증거를 모두 읽은 뒤에도 하나가 더 있다: 디스크 바이트의 해시가 이전 기록의 `TemplateHash` (또는 `DeployedHash`) 와 같아야 한다. 이는 증거가 아니라 일치 확인이다 — 해시 일치는 필요조건이지 충분조건이 아니다 (REQ-OWR-002). 일치하지 않으면 이유는 `record-hash-mismatch`.

원천은 닫힌 두 가지이다.

- **S1 — 실행 시작 시점의 매니페스트.** `runTemplateSyncWithReporter` 가 Backup 단계 앞에서 적재한 매니페스트의 사본 (`update_template_sync.go:285-286`). 마이그레이션이 공유 매니저를 바꾸기 전에 깊은 복사를 떠 두어야 한다 (판정이 자기가 바꾼 상태를 읽지 않도록). 미등록 파일은 정의상 S1 에 없으므로 S1 은 "등록된 파일의 기록" 을 확인하고, 미등록 파일의 증거 원천으로는 S2 만 남는다.
- **S2 — 보존된 이전 세대 사본.** 실행 백업 디렉터리의 `manifest.prior.json` (D-3). 최근 것부터 최대 `priorManifestGenerations` 개 (제안 기본값 8) 를 본다. 비용을 말뭉치 크기와 무관하게 상수로 묶기 위한 상한이다 (coding-standards 의 Advisory-Check Discipline: 시작 경로의 조회는 비용이 코퍼스에 비례하면 안 된다).

새 원천은 SPEC 개정 없이 추가하지 않는다. 특히 두 가지를 증거로 인정하지 않는다.

- 내장 템플릿 트리와의 바이트 일치 — 이 바이너리의 내장 트리가 출시 트리와 같다는 보장이 없다 (dirty 빌드 `ee981a932`, research.md D(1)).
- 정규 스킬 경로와의 기계적 관계 (`.agents/skills/<n>` ↔ `.claude/skills/<n>`) — 관계는 누가 그 복사본을 만들었는지를 증명하지 않는다 (D-5).

테스트 가능성: 증거 원천은 인터페이스 (`provenanceSource`) 로 두어 시험이 이전 세대를 직접 공급할 수 있게 한다. 이것이 AC-OWR-05 의 양성 픽스처와 변이 네 가지를 운영 경로와 독립으로 만든다. 인터페이스와 해석기 `provenanceEvidence` 는 영속 기록 생산 (D-3, M1b) 과 분리된 새 파일에 두고 M1a 에서 만든다 — Q2 가 (나) 로 풀려 M1b 가 접혀도 AC-OWR-05 가 쓰는 코드는 남는다 (감사 D6).

운영 경로가 원천을 얻는 자리는 `removalEnabled` 와 같은 모양의 패키지 수준 변수 하나이다 (시험이 바꾸고 `t.Cleanup` 이 복원하며, 운영 코드 어디도 그 변수를 시험 값으로 대체하지 않는다). 증거 해석기는 이전 기록의 출처값을 반드시 확인한다. 출처값을 무시하고 버전·경로·해시만 보는 구현은 `user_created`·`user_modified` 기록을 `managed` 로 올리므로 AC-OWR-05 의 변이 (e)-(g) 가 이를 붉게 잡는다. 운영 원천 경유 한 건 (AC-OWR-05 (h)) 은 실제 배포기·`manifest.Track`·운영 이전 매니페스트 쓰기와 읽기를 거쳐 만든 이전 세대로 판정이 나오게 하여, 주입한 원천에서만 맞고 운영 원천에서 틀리는 구현을 잡는다. 이 변수가 있어서 AC-OWR-02 의 대조군이 **잘못된 승격 구현을 주입**할 수 있다: 파일 자신의 바이트와 내장 템플릿 트리의 일치만으로 세 증거를 모두 "있다" 고 답하는 대역 원천을 꽂고, 보존 단언이 그 대역 앞에서 붉어지는지를 시험이 직접 관측한다 (단언은 문제 목록을 돌려주는 순수 함수이고 메타 시험이 대역 입력에 대해 목록이 비어 있지 않음을 단언한다). 입력 쪽 변형만으로는 이 변이를 구성할 수 없다 — 내장 트리에 같은 경로 파일이 있는 픽스처여야 변이가 발화한다 (AC-OWR-02 의 전제 단계가 이를 먼저 단언한다).

## D-3 — 매니페스트 추가분과 호환 (decision-index.md Q2 (가) — 리더 조건부 결정, 계획의 기록)

Q2 는 리더가 옵션 (가) 로 조건부 결정했다 (보드 d-20261018T040918Z-def2). 아래 1-8 이 그 조건 (1)-(5) 을 싣는다. 이 결정은 설계 결정일 뿐 사용자 보안 승인·실제 홈 접근 승인·Kickoff 사전 승인이 아니다.

1. `FileEntry.DeployedVersion string` — JSON `deployed_version,omitempty`. `Track` 이 기록할 때 배포 바이너리의 버전을 찍는다 (`pkg/version.GetVersion()`). 파일이 실제로 쓰일 때만 `Track` 이 불리므로 (`deployer.go:294`) 값은 "그 바이트를 마지막에 쓴 버전" 이다. **이번 실행이 배포한 항목에만 찍는다. 이번 실행이 배포하지 않은 항목에는 쓰지도, 채우지도, 추정하지도 않는다** — 다른 경로의 `Track` 호출(사용자 파일 기록, 재추적, 섹션 해시 재기록)이 값을 새로 찍거나 지우면 안 된다: `Track` 이 항목을 통째로 새 값으로 바꾸므로 (`internal/manifest/manifest.go` `Track`) 구현은 이미 있는 `deployed_version` 을 보존하는 쪽으로 쓰고, 버전을 찍는 호출은 배포 경로 하나로 좁힌다 (AC-OWR-06 (a) 의 "재추적 보존" 단언). 혼합 버전 상태가 실재한다: 사용자 수정으로 갱신이 건너뛰어진 파일은 매니페스트 수준 버전보다 오래된 버전의 바이트를 가진다.
2. 매니페스트 수준 `Version` 은 갱신의 배포 단계 저장에서 배포 버전으로 갱신한다 (`recordTemplateTree` 와 같은 자리, `update_template_sync.go:672-673`). `DeployedAt` 은 `Version` 이 바뀔 때만 바꾼다 — 같은 버전의 반복 갱신이 매니페스트 바이트를 바꾸면 REQ-OWR-012 (멱등) 가 깨진다. 매니페스트 저장은 기존 `Save` 의 임시 파일 + 이름 바꾸기 (`internal/manifest/manifest.go` `Save` → `atomicWriteFile`) 를 그대로 쓴다.
3. 실행 백업 디렉터리에 `manifest.prior.json` 을 쓴다: Backup 단계가 시작 시점 매니페스트의 **원본 바이트**를 그대로 쓴다 (재직렬화하지 않는다 — 바이트 동일을 위해). **원자적으로 만든다**: 같은 디렉터리에 임시 이름으로 쓰고 fsync 뒤 이름 바꾸기 — 중간에 멈춘 쓰기는 `manifest.prior.json` 이라는 이름을 갖지 못하고, 읽는 쪽은 정확한 이름만 읽는다. 임시 파일 이름은 읽는 쪽이 무시한다. 유효한 세대의 정의: 정확한 이름의 정규 파일이고, 읽을 수 있고, 매니페스트 JSON 으로 해석되고, `Files` 가 있다. 그 밖 (없음, 읽기 불가, 잘림, JSON 아님, 디렉터리) 은 증거 없음이며 오류로 실행을 죽이지 않고 보존 쪽으로 간다 (REQ-OWR-017).
4. **쓰기·복사 실패는 이 실행의 제거를 막는다 (fail-closed, REQ-OWR-016) — 파일 바이트 복사 실패도 같다.** 이 사본들은 제거된 파일의 복원이 소유권 정보를 함께 되돌리는 유일한 수단이다 (REQ-OWR-010). 구현 모양은 마이그레이션의 제거 암을 두 단계로 나누는 것이다: (i) 걷기가 `removal-eligible` 후보를 모으고 바이트를 읽는다, (ii) 모든 후보의 바이트 사본을 실행 백업에 쓰고 `manifest.prior.json` 이 유효한지 확인한다, (iii) (ii) 가 전부 성공했을 때만 제거와 빈 디렉터리 정리를 한다. (ii) 의 어느 하나가 실패하면 제거는 0건이고 모든 후보가 `keep(backup-failed)` 로 잔존 행과 경고 한 줄에 남는다. 종전 t1668 의 파일별 처리 (복사가 실패한 그 파일만 보존, 나머지는 계속 제거; `migrate_project_assets.go` 의 `backupRemovedAsset` 실패 분기) 보다 엄격하다 — 리더 조건 (2) "백업이 실패하면 그 실행은 파일을 0개 제거한다". 이 엄격화는 t1668 코드의 제거 암 구조를 바꾸므로 병합 충돌 노출이 커진다 (plan.md § D5, 수용 부채). 기존 시험 중 파일별 계속 제거를 단언하는 것은 없다 (`backupRemovedAsset` 실패를 다루는 시험 grep 0건, acceptance.md § L-14). 갱신의 나머지 (배포, 비제거 마이그레이션) 는 계속된다 — 막히는 것은 제거뿐이다. 증거 생산의 다른 쪽 (`deployed_version`, 매니페스트 수준 `Version`) 은 제거와 무관하므로 fail-open 이다.
5. **Backup 단계 → 마이그레이션 인계의 검증 장치.** Backup 단계는 사본 쓰기를 패키지 수준 변수 `priorManifestWriter` (기본은 운영 쓰기 함수) 로 부르고, 결과를 실행 문맥 (`migrationRunContext`) 의 한 필드에 넘긴다. 시험이 이 변수를 바꾸고 `t.Cleanup` 이 복원한다 (`removalEnabled` 와 `provenanceSource` 와 같은 모양, 운영 코드 어디도 시험 값으로 대체하지 않는다). 시험의 대역은 **운영 쓰기 함수를 감싼다**: 목적지 `manifest.prior.json` 자리에 내용이 있는 디렉터리를 먼저 만든 뒤 운영 함수를 부른다 — 디렉터리 경로에 파일을 이름 바꾸기로 놓는 일은 uid 와 무관하게 실패하므로 root 로 도는 실행에서도 성립한다. 실행 백업 디렉터리 이름이 시각에서 나와 (`update_disk_backup.go:73`) 시험이 이름을 미리 알 수 없다는 점은 이 대역이 `runDir` 인자를 받아 그 안에서 만들기 때문에 걸림돌이 아니다. 그리고 검증은 전체 갱신 흐름 (`runTemplateSyncWithProgress`, Backup 단계 → 마이그레이션 호출, `update_template_sync.go:815-817`) 을 `--force` 로 돌려서 한다 (AC-OWR-20). 인계 필드를 성공으로 두는 구현 (사본 실패를 삼키는 변이) 은 이 검증이 잡는다.
6. 호환: `deployed_version` 이 없는 기존 JSON 은 그대로 적재되고 저장 시 필드를 지어내지 않는다 (왕복에서 바이트가 같다 — AC-OWR-21 (c)). 증거 부족은 미등록 파일의 `managed` 승격만 막는다 — 등록된 항목의 기존 판정 (`TemplateHash` 와 바이트 비교) 은 바뀌지 않는다. 같은 버전의 반복 실행은 매니페스트 바이트를 바꾸지 않는다.
7. **다운그레이드.** 이전 바이너리가 새 매니페스트를 읽으면 알 수 없는 필드를 무시하고, 다시 저장하면 `deployed_version` 이 사라진다. 그러면 그 항목은 증거가 없는 것으로 읽혀 보존 쪽으로 간다 — 증거는 사라진 만큼 약해질 뿐 제거 방향의 오류는 만들지 않는다 (AC-OWR-21 (d): `deployed_version` 을 알지 못하는 구조체로 적재·저장한 JSON 을 새 해석기에 먹이면 `evidence-incomplete:deployed-version`).
8. **보존 기간과의 관계.** `priorManifestGenerations = 8` 은 **조회 한도**이다: 증거 원천이 최근 실행 백업에서 최대 8개의 `manifest.prior.json` 만 읽는다. 이 SPEC 은 삭제·정리·보존 기간 단축을 더하지 않는다. 기존 백업 보존 정책은 그대로이다 — `backup.PruneBackups(projectRoot, 5, pastKeep)` (`update_template_sync.go:944`) 이 최신 5개 실행을 통째로 두고, 더 오래된 실행은 선택 규칙이 요구하는 것만 남긴다 (`internal/cli/update/backup/retention_prune.go:82`). 따라서 실제로 읽을 수 있는 세대는 정책이 남긴 실행 수 이하이며 8 은 그 위의 상한이다 (5 < 8 이라 한도는 대부분 조회를 줄이지 않는다). `manifest.prior.json` 은 실행 백업 루트의 일반 파일이라 보존 정책의 센서스 (`retention_prune.go`, 장부 파일 판정 `isBookkeepingFile` 은 `RunMarkerName`·`.complete`·`backup_metadata.json` 만 장부로 본다) 에서 사용자 내용 파일처럼 해시되어 계산된다 — 정책 자체는 바뀌지 않는다는 점이 이 SPEC 의 약속이고, `internal/cli/update/backup/` 아래 파일은 이 카드의 커밋이 건드리지 않는다 (AC-OWR-21 (e) 의 파일 불변 증거, 양성 대조군 포함). 이 센서스가 `manifest.prior.json` 의 서로 다른 버전을 이유로 오래된 실행을 정책 예산 (`pastKeep`) 안에서 더 남길 수 있는지는 코드 읽기로 가능성만 확인했고 측정하지 않았다 (acceptance.md 갭 G-K) — 남긴다면 정책의 기존 한도 안이다. 기존 소비자 `internal/core/project/validator.go` 는 `backups/*/manifest.json` 을 읽는다 — 이름이 달라 충돌하지 않는다.

Q2 의 옵션 (나) 는 기록된 대안이다. (나) 였다면 1-8 이 없고 6 의 호환만 남는다 (plan.md § G.1).

## D-4 — 대소문자 변형: 모드 프로브 없이 별칭 확정

걷기가 돌려준 경로 `rel` 이 매니페스트 키와 정확히 같으면 그대로 쓴다. 정확한 키가 없을 때만 별칭을 찾는다.

1. 같은 디렉터리의 매니페스트 키 중 `strings.EqualFold(key, rel)` 인 후보를 모은다.
2. 후보마다 `os.Stat(root/rel)` 과 `os.Stat(root/key)` 를 하고 `os.SameFile` 이 참이면 별칭이다. 후보가 정확히 하나일 때만 확정한다 (둘 이상이면 모호하므로 별칭 없음).
3. 대소문자 구분 파일시스템에서는 `root/key` 가 없거나 다른 파일이라 `SameFile` 이 거짓이다. 그러므로 파일시스템 모드를 따로 판정할 필요가 없다 — 판정 근거가 관측된 동일성 자체이다.

`probeCaseSensitiveFS` (`update_cleanup.go:387`) 는 프로젝트 루트에 프로브 파일을 쓴다. dry-run 은 아무것도 쓰면 안 되므로 (REQ-OWR-011) 판정 경로에서 쓰지 않는다. 이 함수에는 비시험 호출자가 없다 (acceptance.md § L-9). plan.md M4 가 이 함수를 재사용한다고 적었던 종전 문장은 이 절과 모순이어서 고쳤다.

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

## D-8 — 프로필 루트 무접근을 어떻게 검증하는가 (감사 D2, D7, N1, N6, 리더 판정 OQ-5)

**왜 종전 방식이 부족했는가.** (1) 0o000 디렉터리 읽기 오류에 기대는 방식은 읽기 오류를 버리는 구현을 통과시키고, root 로 도는 실행에서는 권한 오류 자체가 나지 않는다. (2) `claude-skills@` 문자열 스캔은 공허하다: 이 트리의 Go 소스에는 그 문자열이 없다 (acceptance.md § L-7). (3) iteration 1 의 개정판은 `internal/cli`, `internal/userassets`, `internal/manifest` 세 패키지의 아홉 토큰만 봤다. 갱신 경로의 호출 닫힘에는 `internal/profile` 이 있고 (마이그레이션이 `profile.ResolveLaunchProfileForProject` 를 부른다), `profile.GetProfileDir` 같은 내보낸 도우미는 아홉 토큰에 없었다. 읽기를 하고 결과를 버리는 변이는 스캔·카나리·카드 이력 검사 어느 것도 통과했다 (감사 iteration 2 N1; codex 가 같은 종류의 overlay 변이로 `reads=3 ... all 3 observed outputs identical` 을 관측).

**선택한 방식 — 네 층, 서로 다른 구멍을 막는다.**

| 층 | 무엇을 | 잡는 것 | 못 잡는 것 (한계) |
|----|--------|---------|-------------------|
| L1 심볼 단위 소스 스캔 (호출 지점 동결) | 시험이 `go list -deps -f '{{.ImportPath}}' ./internal/cli` 로 **호출 닫힘을 도출한다** (손으로 쓴 패키지 목록 없음; 이 트리에서 모듈 안 패키지 142개, 여섯 핵심 패키지 포함 — § L-11). 닫힘 안 모든 비시험 `.go` 파일을 `go/parser` 로 해석해, 아래 심볼 집합의 **호출 지점**을 (파일, 심볼) 별 개수로 센다: ① `internal/profile` 이 **내보낸 모든 식별자**(시험이 그 패키지의 비시험 파일에서 기계적으로 열거 — 새 export 는 자동으로 집합에 들어간다; 이 트리의 목록 § L-12 의 `GetBaseDir`, `GetCurrentName`, `GetCurrentNameForProject`, `List`, `Delete`, `GetProfileDir`, `EnsureDir`, `ResolveLaunchProfile`, `ResolveLaunchProfileForProject`, `HasClaudeConfig`, `GetPreferencesPath`, `IsSetup`, `ReadPreferences`, `WritePreferences`, `RecordLastUsedProfile`, `RecordLastUsedProfileForProject`, `PruneStaleProjectEntries`, `SyncToProjectConfig` 등) 의 `profile.<식별자>` 선택자 사용, ② 문자열 토큰 `CLAUDE_CONFIG_DIR`, `claude-profiles`, `claude-skills@`, `claude-agents@`, `GetBaseDir`, `BaseDirOverride`, `EnvClaudeConfigDir`, `profilesDir`, `GetCurrentName` — **상수 접기 뒤에** 비교한다 (`"CLAUDE_"+"CONFIG_DIR"` 는 접혀 토큰이 된다), ③ 환경 변수 이름이 컴파일 시점 상수가 아닌 `os.Getenv`/`os.LookupEnv`/`os.Environ` 호출의 개수. 그리고 `internal/profile` 안에서는 **함수 단위로** 파일시스템 동사 (`os.Stat`, `os.Lstat`, `os.ReadDir`, `os.Open`, `os.OpenFile`, `os.ReadFile`, `os.WriteFile`, `os.MkdirAll`, `os.Remove*`, `os.Rename`, `filepath.Walk*`, `fs.*`) 호출의 개수를 센다. 기준선은 M0 에서 손대지 않은 트리로 측정해 얼린다 (`testdata/profile_callsite_baseline.txt`). **새 (파일, 심볼) 행이나 늘어난 개수는 붉다.** 이름 붙은 예외는 기준선의 **한 행**이다: `internal/profile` · 함수 `launchCandidateIsUsable` · `os.Stat` · 개수 정확히 1 (`profile.go:439` 는 참고 좌표; 비교는 함수 이름과 개수로 한다) — 같은 함수 안의 둘째 `os.Stat` 은 개수 2 라서, 다른 함수의 `os.Stat` 은 새 행이라서 붉다. 파일 전체나 패키지 전체를 예외로 두지 않는다. | 갱신 호출 닫힘 어디든 새로 생긴 프로필 심볼 호출 — 내보낸 도우미 (`GetProfileDir` 포함), 토큰 접합, 상수 이름이 아닌 환경 변수 읽기, `internal/profile` 안의 새 파일시스템 호출. **읽기 결과를 버리는 변이도 호출 지점이 새로 생기므로 잡힌다.** | 어떤 나열된 심볼·토큰도 쓰지 않고 조각난 문자열로 경로를 짓는 구현 (예: `filepath.Join(home, ".moai", "claude-"+x, ...)` 에서 `x` 가 실행 시점 값), 기준선에 이미 있는 호출 지점의 인자만 바꾸는 구현 (개수가 같다). 접근 기록 시임이 없으므로 이 구멍은 남는다 — decision-index.md 수용 부채에 이름으로 올렸다 |
| L2 스캔의 변이 대조군 | L1 의 스캔 함수는 파일 → 소스 맵을 받는 순수 함수이다. 메타 시험이 실제 소스에서 만든 맵에 변이 소스를 끼워 비어 있지 않은 문제 목록을 단언한다: M-a 마이그레이션 파일에 `_, _ = os.ReadDir(filepath.Join(profile.GetProfileDir("work"), "skills"))` 추가, M-b `os.Getenv("CLAUDE_"+"CONFIG_DIR")` 추가, M-c `internal/profile` 의 `launchCandidateIsUsable` 에 둘째 `os.Stat` 추가, M-d `internal/profile` 의 다른 함수에 `os.ReadDir` 추가, M-e `profile.HasClaudeConfig("work")` 호출 추가, M-f 상수가 아닌 이름의 `os.Getenv(name)` 추가. 그리고 음성 대조: 변이 없는 실제 맵은 빈 목록이며 예외 행의 개수가 정확히 1 로 일치한다. 공허 방지: 스캔한 패키지 집합이 `go list -deps` 결과와 같고 `internal/profile` 을 포함하며 닫힘의 패키지 수가 0 이 아님을 단언한다 | 스캔이 해당 변이에 대해 실제로 붉어지는지 — "시험이 붉음을 직접 관측" | 위 한계와 같다 |
| L3 차분 카나리 (정규화 관측) | 프로필 설정 디렉터리 (`profile.BaseDirOverride` 아래 `work/`) 의 `skills/`·`agents/` 에 대응 파일을 세 상태 — 없음, 템플릿과 바이트 같음, 해시 다름 — 로 놓은 같은 프로젝트 세 벌에 갱신 흐름을 돌려 **정규화한 관측 기록**이 세 번 모두 같음을 단언한다. 정규화(N6): 임시 루트 경로를 `<ROOT>` 로, 실행 백업 디렉터리 이름 (`defs.BackupTimestampFormat` 형태 `YYYYMMDD_HHMMSS`) 을 `<RUN>` 으로, RFC3339 타임스탬프와 매니페스트의 `deployed_at` 값을 `<TS>` 로 바꾼 뒤 비교하고, 백업 디렉터리 목록은 `<RUN>` 아래 상대 배치로 비교한다. 정규화 규칙 자체를 시험이 먼저 단언한다 (같은 입력을 두 번 정규화한 결과가 같고, 정규화로 지워진 부분이 위 세 종류뿐임). 지워진 영역은 비교 대상이 아니라는 점을 한계에 적는다 | 프로필 쪽 읽기 결과가 판정·출력·부수 효과에 영향을 주는 구현 | 읽기 결과를 버리는 접근 — L1 의 몫. 둘 다 못 잡는 경우(조각난 경로 + 버림)가 남는다 |
| L4 파일 불변 | `git log --oneline cdd72a814..HEAD --grep='(card t1865)' -- internal/config/protected_zone.go internal/config/userroots.go` 가 빈 출력·종료 0, 같은 형식의 양성 대조군(`-- .moai/specs`)이 한 줄 이상. 같은 형식으로 `internal/profile/` 아래 파일을 이 카드가 건드리지 않았음도 단언한다 (양성 대조군 포함) | 이 카드의 커밋이 보호 파일 둘이나 `internal/profile` 을 건드린 것. main 병합이 두 파일을 바꿔도 거짓 적색이 되지 않는다 | 카드 id 를 빠뜨린 커밋 (추적 운반자 규칙 AGENTS.md § 3 이 지킨다) |

**차분 카나리가 도우미를 통한 읽기를 관측한다는 약속은 하지 않는다.** 이 트리에서 파일 접근 기록 없이 런타임에 "읽혔다" 를 관측할 방법은 없다 (atime 은 파일시스템 설정에 따라 갱신되지 않아 근거로 쓸 수 없다). 도우미를 통한 읽기를 잡는 것은 L1 의 심볼 단위 호출 지점 동결이고, L3 는 읽기 결과가 관측 가능한 곳에 새어 나오는 구현을 잡는다. 이 분담을 acceptance.md AC-OWR-08 한계 항에 그대로 적었다.

**접근 기록 시임은 만들지 않는다.** 마이그레이션은 `os.ReadFile`/`os.Stat` 를 직접 쓴다. 파일 접근을 모두 주입 가능한 파일시스템으로 감싸는 것은 판정 코드의 모양을 바꾸는 큰 변경이며 t1668 코드와 병합 충돌 면적을 넓힌다 (감사 D15). 감사가 제안한 접근 기록 시임은 이 SPEC 의 범위에서 채택하지 않았고, 대신 위 네 층의 한계를 있는 그대로 적고 남는 구멍 (조각난 경로 + 버림) 을 decision-index.md 수용 부채에 리더가 이름으로 수락 또는 거부할 항목으로 올렸다.

**0o000 모드는 판정 근거로 쓰지 않는다.** root 로 도는 실행에서 무력하고, 오류 전파를 전제하기 때문이다.

## D-9 — dry-run 무변경을 어떻게 검증하는가 (감사 D7)

REQ-OWR-011 은 명령 (`moai update --dry-run`) 에 걸린다. 검증은 그 명령이 디스패치하는 분기 전체 (`runUpdate`, `update.go:361`) 를 기존 시험 헬퍼 `runDryRunUpdateInProject` 로 구동한다. 트리 해시 비교만으로는 "만들었다가 지우는" 임시 쓰기 (예: `probeCaseSensitiveFS` 가 하는 일) 를 못 본다. 그래서 실행 전에 모든 디렉터리의 mtime 을 고정된 과거 시각으로 맞추고 (`os.Chtimes`), 트리 해시에 디렉터리 mtime 을 넣는다: 디렉터리 안에서 항목이 생기거나 사라지면 mtime 이 지금으로 바뀐다. 이 방식은 mtime 해상도와 무관하고 root 에서도 성립한다. 한계: 파일 내용을 바꾸지 않고 모드만 바꿨다 되돌리는 쓰기, 프로젝트 밖(HOME 등) 으로의 쓰기는 이 방식이 보지 못한다 — 기존 `TestUpdateDryRun_ZeroMutation` 이 기존 분기의 같은 성질을 지키고, 이 SPEC 이 더하는 코드 (판정·미리보기 새 파일) 는 쓰기 가능 호출이 없음을 소스 스캔으로 따로 단언한다 (AC-OWR-16).

**공허 방지 (감사 N5 의 같은 부류, 이번 실행 코드 읽기).** `previewReconciliation` 은 `plan.Count() == 0 && len(plan.Symlinks) == 0` 이면 아무것도 출력하지 않고 돌아온다 (`update_dryrun_preview.go`), 후보 알림은 `removalEnabled && isCommonAssetCleanTarget(...) && migrationRemovalCandidate(...)` 가 모두 참일 때만이고, 분류 대상은 reconcile 이 `UserOwned` 로 분류한 파일에 한정된다 (공통 자산 루트는 `PreserveOnly` 라 `UserOwned` — `reconcile_classify.go:classifyManagedFile` 2번). 그래서 AC-OWR-16 은 비교 전에 세 가지를 단언한다: 출력에 "Dry-run: reconciliation preview" 머리가 있다, 픽스처의 `removal-eligible` 집합이 비어 있지 않다, 그 집합의 모든 경로가 reconcile 계획의 `UserOwned` 에 있다 (그렇지 않은 경로가 있으면 그 사실이 곧 측정 결과이며 비교 범위를 줄이는 근거로 RED 셀에 적는다 — 빈 집합끼리의 같음이 통과가 되지 않도록).

## D-10 — 여러 번 실행하는 기준과 버전 일치 건너뛰기 (감사 N5, 이번 실행 측정)

`runTemplateSyncWithProgress` 는 `packageVersion == projectVersion && !forceUpdate && versionMatchSkipsSync(projectRoot)` 이면 조정 단계 앞에서 "이미 최신" 으로 끝난다 (`update_template_sync.go:1239-1248`; `versionMatchSkipsSync` 는 매니페스트 `TemplateTreeHash` 가 내장 템플릿 트리 해시와 같으면 참 — `update_template_sync_guard.go:55-68`). 기존 헬퍼 `runUpdateFlowWithProgress` 는 `force` 플래그를 등록하고 값을 바꾸지 않는다 (`migrate_removal_backup_test.go`). 이번 실행 측정 (acceptance.md § L-13, overlay 탐침 — 트리 파일은 건드리지 않음): 같은 픽스처에서 일반 실행 1 은 건너뛰지 않고 (`skipped=false`), 일반 실행 2 는 건너뛰고 (`skipped=true`), `--force` 실행은 건너뛰지 않으며, 그 뒤 일반 실행은 다시 건너뛴다.

선택한 메커니즘은 **둘 다 시험한다**: (a) 여러 번 실행 기준 (AC-OWR-14 둘째 실행, AC-OWR-17, AC-OWR-18, AC-OWR-10) 은 조정에 닿게 하려고 `--force` 를 쓰는 새 헬퍼 `runUpdateFlowForced` (M0, `runTemplateSyncWithProgress` 의 `skipped` 반환값을 돌려준다) 로 구동한다. (b) 같은 기준에 **건너뛰는 실행 사례**를 따로 둔다: 일반 같은 버전 재실행이 `skipped == true` 이고 트리 해시·매니페스트 바이트·출력이 바뀌지 않는다 (운영 사실을 있는 그대로 단언). 어떤 실행도 읽혔다고 말하기 전에 공허 방지 단언을 건다: 모든 `--force` 실행이 `skipped == false` 를 돌려주고, 실행 백업 디렉터리에 그 실행의 `manifest.prior.json` 이 새로 생겼으며 (Backup 단계 도달의 증거; 바이트가 실행 전 매니페스트와 같음) 잔존 행 개수가 픽스처가 기대하는 양수이다 (마이그레이션 도달의 증거). 쓸어낸 실행 수를 단언한다 (3회 실행이면 `skipped == false` 인 실행이 정확히 N 개). 이 SPEC 은 건너뛰기 관문을 바꾸지 않는다. 일반 재실행이 건너뛰는 탓에 복구가 닿지 않는 경우는 plan.md OQ-6 으로 리더에게 올렸다.

