---
title: 마이그레이션 가이드
description: 기존 .agency/ 프로젝트를 새 디자인 시스템으로 전환
weight: 60
draft: false
---

# 마이그레이션 가이드

SPEC-AGENCY-ABSORB-001 에 따라 `/agency` 명령어가 `/moai design` 으로 통합되었습니다. 기존 `.agency/` 디렉터리를 보유한 프로젝트는 **마이그레이션**을 통해 새 시스템으로 전환할 수 있습니다.

## 언제 마이그레이션하나?

**다음 중 하나에 해당하면 마이그레이션 필요:**

- `.agency/` 디렉터리가 존재함
- 기존 agency learnings/observations 사용 중
- 구 에이전트 (agency-copywriter, agency-designer 등) 활용 중

**마이그레이션 이후:**

- `.agency/` → `.agency.archived/` (백업)
- `.moai/project/brand/` (새로 생성)
- `.moai/config/sections/design.yaml` (새로 생성)
- 이전 learnings 병합 가능

## 마이그레이션 실행

### 1단계: 사전 확인

```bash
# .agency/ 존재 확인
ls -la .agency/
# brand-voice.md
# visual-identity.md
# learnings/
# observations/
```

### 2단계: 드라이런 (선택 사항)

마이그레이션 결과를 미리 확인:

```bash
moai migrate agency --dry-run
```

**출력 예시:**
```
[미리보기] .agency/ 에서 .moai/project/brand/ 로 마이그레이션
파일 3개 이동:
  ✓ brand-voice.md
  ✓ visual-identity.md
  ✓ target-audience.md
설정 병합:
  ✓ learnings 5개 수집
  ✓ observations 12개 수집
백업:
  ✓ .agency.archived/ 에 저장될 예정
```

### 3단계: 실제 마이그레이션 실행

```bash
moai migrate agency
```

**실행 순서 (6단계):**

1. **검증** — `.agency/` 존재 확인, 디스크 공간 확인
2. **Staging** — 임시 디렉터리에 복사
3. **컨텍스트 이전** — brand 파일을 `.moai/project/brand/` 복사
4. **설정 병합** — learnings/observations를 `.moai/config/` 병합
5. **Learning 이전** — 기존 heuristics를 새 구조로 변환
6. **Atomic Swap** — `.agency.archived/` 로 원본 백업, 완료

**완료 시 출력:**
```
마이그레이션 완료 [TX-abc123def456]

이동된 파일: 47개
  ✓ .moai/project/brand/ 3개
  ✓ .moai/config/sections/design.yaml 생성
  ✓ .moai/research/ 설정 병합

백업 위치: .agency.archived/

다음 단계:
  /moai design
```

## 마이그레이션 옵션

### --force 옵션

기존 대상 디렉터리 덮어쓰기:

```bash
moai migrate agency --force
```

**주의:** `.moai/project/brand/` 가 이미 존재하면 덮어씀. 사전 백업 권장.

### --resume 옵션

중단된 마이그레이션 재개:

```bash
# 이전에 SIGINT로 중단된 경우
moai migrate agency --resume TX-abc123def456
```

checkpoint 파일: `~/.moai/.migrate-tx-<txID>.json`

## 마이그레이션 에러 코드

마이그레이션 실패 시:

| 에러 코드 | 원인 | 해결 방법 |
|---|---|---|
| `MIGRATE_NO_SOURCE` | `.agency/` 없음 | 기존 agency 디렉터리 확인 |
| `MIGRATE_TARGET_EXISTS` | `.moai/project/brand/` 이미 존재 | `--force` 옵션 사용 |
| `MIGRATE_ARCHIVE_EXISTS` | `.agency.archived/` 이미 존재 | 기존 백업 파일 삭제 또는 이동 |
| `MIGRATE_DISK_FULL` | 디스크 공간 부족 | 여유 공간 확보 (최소 100MB) |
| `MIGRATE_MERGE_CONFLICT` | tech-preferences.md 충돌 | 기존 `.moai/project/tech.md` 백업, 재시도 |
| `MIGRATE_INTERRUPT` | SIGINT/SIGTERM 수신 | `--resume` 옵션으로 재개 |
| `MIGRATE_CHECKPOINT_CORRUPT` | 체크포인트 파일 손상 | `~/.moai/.migrate-tx-*.json` 삭제 후 재시도 |

## 마이그레이션 결과

### 생성 파일 구조

```
.moai/
├── project/
│   └── brand/
│       ├── brand-voice.md        (from .agency/)
│       ├── visual-identity.md    (from .agency/)
│       └── target-audience.md    (from .agency/)
├── config/
│   └── sections/
│       └── design.yaml           (새로 생성)
└── research/
    ├── learnings/                (병합됨)
    └── observations/             (병합됨)

.agency.archived/                  (원본 백업)
├── brand-voice.md
├── visual-identity.md
├── learnings/
└── observations/
```

### Learning 병합

기존 `.agency/learnings/` 의 항목들은:
- 새 구조로 변환
- `.moai/research/learnings/` 에 병합
- MIGRATED 태그 추가

**변환 예시:**

```yaml
# 기존 format
id: LEARN-20260401-001
category: copy
observation: "Hero 단락은 15단어 이하로 제한"
confidence: 0.85

# 마이그레이션 후
id: LEARN-20260420-001-MIGRATED-FROM-20260401-001
status: graduated
category: copy
observation: "Hero 단락은 15단어 이하로 제한"
confidence: 0.85
migrated_from_agency: true
```

## 롤백

마이그레이션 후 이전 상태로 되돌려야 할 경우:

### 옵션 1: 백업 파일에서 복원

```bash
# 백업 위치에서 복원
mv .agency.archived .agency

# 새로 생성된 파일 제거
rm -rf .moai/project/brand
rm .moai/config/sections/design.yaml
```

### 옵션 2: Git을 통한 복원

마이그레이션이 git commit을 생성한 경우:

```bash
git log --oneline | grep migrate
# abc1234 chore: migrate agency to moai design system

git revert abc1234
```

## 마이그레이션 후 다음 단계

마이그레이션이 완료되면:

1. **브랜드 컨텍스트 확인**
   ```bash
   cat .moai/project/brand/brand-voice.md
   cat .moai/project/brand/visual-identity.md
   ```

2. **새 디자인 워크플로우 시작**
   ```
   /moai design
   ```

3. **기존 learnings 검토**
   ```bash
   ls .moai/research/learnings/
   ```

4. **선택적: 기존 .agency/ 삭제**
   ```bash
   rm -rf .agency.archived
   ```

## 마이그레이션 상태 확인

마이그레이션 결과 요약 조회:

```bash
# 마이그레이션 로그 조회
cat ~/.moai/.migrate-tx-abc123def456.json

# 또는 CLI로 확인
moai status design
# Design System Status: MIGRATED (2026-04-20)
# Brand Files: 3/3 ✓
# Design Config: ✓
```

## SIGINT/SIGTERM 처리

마이그레이션 중 중단된 경우:

**Ctrl+C 입력 시:**
```
마이그레이션 중단됨 [TX-abc123def456]
완료된 단계: validation, staging, context-transfer
미완료 단계: config-merge, learning-transfer, atomic-swap

재개하려면:
  moai migrate agency --resume TX-abc123def456
```

**체크포인트 파일 위치:** `~/.moai/.migrate-tx-abc123def456.json`

미완료 단계부터 자동으로 재개됩니다.

## FAQ

### Q: 마이그레이션 후 기존 /agency 명령어 사용 가능?

**A:** 아니요. `/agency` 는 더 이상 지원되지 않습니다. 대신 `/moai design` 사용.

### Q: 마이그레이션을 여러 번 실행할 수 있나?

**A:** 첫 번째 완료 후 `.agency/` 가 `.agency.archived/` 로 변경되므로, 두 번째 실행 시 source not found 에러. `--force` 옵션으로 덮어쓰기 가능.

### Q: Learnings가 손실되지 않나?

**A:** 아니요. 모든 learnings/observations는 `.moai/research/` 로 병합되고, 백업도 `.agency.archived/` 에 보존됨.

### Q: 마이그레이션 중 네트워크 끊김?

**A:** 완료된 단계는 저장되므로, 네트워크 복구 후 `--resume` 옵션으로 재개 가능.
