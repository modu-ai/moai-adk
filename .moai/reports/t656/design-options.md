# t656 설계 선택지 — settings.json 템플릿 값 변경 전달

- 카드: t656 (Class C) · 브랜치 `WT-update-value-merge` · 기준 커밋 `81c1d58f9`
- 작성: 2026-09-11 · 근거는 전부 코드 판독이다(실행한 명령·테스트 없음)

## 1. 현황 (판독 근거)

| 사실 | 위치 |
|---|---|
| 병합 base 는 배포된 템플릿을 사용자 파일이 가진 키로 좁혀 만든다. 공유 leaf 값은 base 와 updated 가 같아서 템플릿의 값 변경이 보이지 않는다 | `internal/cli/update/merge/base.go:29-34, 112-131` |
| JSON 3-way 규칙: 사용자만 바꿈 → 사용자, 템플릿만 바꿈 → 템플릿, 둘 다 바꿈 → 충돌 기록 후 사용자 값 | `internal/merge/strategies.go:418-461` |
| 배열은 leaf 로 통째 비교한다(원소 단위 병합 없음) | `internal/merge/strategies.go:437-459` |
| `MergeUserFiles` 는 일반 update 와 clean-reinstall 두 경로가 함께 쓴다 | `update_template_sync.go:544`, `update_clean_install.go:507` |
| sections 스냅숏은 **복원(병합) 뒤** 디스크 상태를 복사한다 | `update_template_sync.go:505-531`, `backup/snapshot.go:51` |
| sections 병합 규칙: old == base → 새 값 채택 | `backup/merge.go:74-76` |
| settings.json.tmpl 은 기계별 값(SmartPATH·Platform·GitMode·HookOptIn)으로 렌더된다 | `templates/.claude/settings.json.tmpl` (자리표시자 18개) |

## 2. 결정할 축 세 가지

### 축 A — 진짜 base 를 어디서 얻는가

| 안 | 내용 | 장점 | 단점 |
|---|---|---|---|
| **A1 (권장)** | 배포가 settings.json 을 렌더해 쓴 **직후, 병합 전** 바이트를 스냅숏으로 저장해 다음 update 의 base 로 쓴다. 스냅숏이 없으면 현행 `pruneToShared` 로 폴백 | 사용자가 건드리지 않은 키의 값 변경이 전달된다. 첫 사이클은 현행과 같아 퇴행이 없다 | 첫 update 한 번은 여전히 전달되지 않는다. 저장 지점을 init·일반 update·clean-reinstall 세 곳에 배선해야 한다 |
| A2 | sections 규칙을 그대로 복제 — **병합 뒤** 디스크 상태를 스냅숏으로 저장 | 기존 함수 재사용 | 아래 가설 참고. 사용자가 편집한 값이 base 에 들어가 다음 update 에서 "사용자가 안 바꿈"으로 읽혀 템플릿 값으로 덮일 수 있다 |
| A3 | 스냅숏 없이 매니페스트 `DeployedHash` 와 현재 파일 해시만 비교해, 파일 전체가 미편집이면 템플릿 전체를 채택 | 새 저장물이 없다 | 한 글자라도 편집한 사용자에게는 아무 값도 전달되지 않는다. `DeployedHash` 가 병합 전 값인지 미확인 |

### 축 B — 사용자와 템플릿이 같은 값을 둘 다 바꿨을 때

| 안 | 내용 | 영향 |
|---|---|---|
| **B1 (권장)** | 현행 엔진 규칙 유지 — 사용자 값 유지 + 충돌 보고 | 엔진 변경 없음. 단 `permissions.allow` 에 사용자가 한 줄만 더해도 그 배열의 템플릿 변경은 통째로 막힌다 |
| B2 | 배열을 원소 단위로 3-way 병합(추가·삭제만 반영) | 권한·훅 목록 전달력이 커진다. 순서가 의미 있는 배열(hooks)은 위험하고 공용 엔진(`internal/merge`)을 건드려 다른 파일 병합에도 번진다 |
| B3 | 지정한 키만 템플릿 값 강제(예: moai 훅 래퍼 경로, statusLine) | 전달이 확실하다. 관리 키 목록을 별도로 유지해야 하고 사용자 편집을 조용히 덮는다 |

### 축 C — sections `SnapshotSubdir` 규칙과의 관계

| 안 | 내용 |
|---|---|
| **C1 (권장)** | 캐시 루트(`.moai/cache/template-snapshot/`)만 공유하고 하위 경로를 나눈다(예: `claude/settings.json`). 쓰는 시점은 다르다(배포 직후 vs 복원 뒤). sections 규칙은 이번 SPEC 에서 건드리지 않는다 |
| C2 | 두 규칙을 하나의 "배포 직후 렌더 스냅숏"으로 통합 — sections 쪽도 교체한다. SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT-001 개정이 되어 범위가 커진다 |

## 3. 가설 — sections 스냅숏의 사용자 편집 유실 (미검증)

코드 판독만으로 세운 가설이며 테스트로 확인하지 않았다.

1. 사용자가 `quality.yaml` 의 `test_coverage_target` 을 80 → 90 으로 바꾼다.
2. update 1회차: 복원이 90 을 지키고, 복원 뒤 디스크(90)가 스냅숏이 된다.
3. update 2회차: base = 90, old = 90, new = 85. `old == base` → 새 값 85 채택 → 사용자 편집 유실.

기존 테스트(`backup/snapshot_provenance_test.go` AC-TBS-011/012)는 스냅숏을 손으로 심어 한 사이클만 보므로 이 두 사이클 경로를 다루지 않는다. 사실이면 A2 는 settings.json 에 같은 결함을 복제한다. 확인하려면 `internal/cli/update/backup` 패키지 테스트 하나(두 사이클 재현)가 필요하다 — 별도 카드로 분리할지는 리드 판단.

## 4. A1 을 택할 때 SPEC 에 담을 주의점

- 스냅숏은 병합 **전** 렌더 바이트여야 한다. 병합 뒤 단계(`stripRetiredV2DenyEntries` 등, `update.go:362-371`)의 수정분이 base 에 섞이면 안 된다.
- 기계가 바뀌어 렌더 값(PATH 등)이 달라지면 "템플릿 변경"으로 읽혀 새 값이 전달된다 — 의도한 동작. 사용자가 `env.PATH` 를 직접 고쳤다면 충돌 → 사용자 값 유지.
- 스냅숏이 깨졌거나 JSON 으로 읽히지 않으면 현행 경로로 폴백(비차단).
- 예상 변경 파일: `update/merge/base.go`, `update/merge/merge.go`, `update/backup/snapshot.go`(또는 새 파일), `update_template_sync.go`, `update_clean_install.go`, init 저장 지점 — 테스트 제외 6개 안팎.

## 5. 리드에게 요청

1. 축 A·B·C 결정(운영자 결정 필요).
2. §3 가설을 별도 카드로 뺄지 판단.
3. 결정 뒤 manager-spec 에 SPEC 작성을 위임하고, plan-auditor 판정을 `.moai/reports/t656/` 에 남긴다. 결정 전에는 SPEC 을 쓰지 않는다(재작성 방지).
