---
id: DOC-TAG-003
version: 0.0.1
status: draft
created: 2025-10-29
updated: 2025-10-29
author: "@Goos"
priority: high
category: Integration / Workflow / Migration
labels: [documentation, tags, migration, batch-processing]
depends_on: [DOC-TAG-001, DOC-TAG-002]
scope: "Phase 3 of 4-phase @DOC TAG automatic generation system - Batch migration of 33 untagged files"
---

# @SPEC:DOC-TAG-003: @DOC 태그 자동 생성 - 배치 마이그레이션 (Phase 3)

## HISTORY

### v0.0.1 (2025-10-29)
- **INITIAL**: Phase 1/2 완료 후 33개 미태깅 파일 배치 마이그레이션
- **AUTHOR**: @Goos
- **SCOPE**: 7개 배치 전략 (Quick Wins → Skills → Architecture → Concepts → Workflows → Tutorials → Polish)
- **CONTEXT**: 42.3% 갭 해소, 78/78 파일 완전 태깅 달성
- **STATUS**: 계획 승인됨, 구현 대기 중

---

## Environment (환경)

**WHEN** MoAI-ADK Phase 3 배치 마이그레이션 환경에서:

### Phase 1/2 완료 상태

- **Phase 1 라이브러리**: 완전히 작동하는 @DOC TAG 생성 라이브러리 (90.5% 테스트 커버리지)
  - `DocumentParser`: 마크다운 파싱 및 TAG 추출
  - `DocTagGenerator`: 도메인 기반 TAG ID 생성
  - `SpecDocMapper`: SPEC-DOC 매핑 및 신뢰도 점수
  - `TagInserter`: 마크다운 헤더에 TAG 삽입
  - `TagRegistry`: TAG 인벤토리 관리

- **Phase 2 워크플로우**: 완전히 통합된 자동 TAG 생성 워크플로우
  - Phase 1.5: `/alfred:3-sync`의 TAG 할당 체크
  - Phase 2.5: `doc-syncer`의 자동 TAG 생성
  - 사용자 승인 모델 (AskUserQuestion)
  - 백업 관리 시스템

### 현재 상태 (2025-10-29 기준)

- **총 마크다운 파일**: 78개
- **태깅 완료**: 45개 (57.7%)
- **미태깅 파일**: 33개 (42.3%)
- **도메인 분포**:
  - 기존 도메인: `@DOC:AUTH-*`, `@DOC:INSTALLER-*`, `@DOC:PLAN-*`, `@DOC:SYNC-*`, `@DOC:TAG-*`, `@DOC:TDD-*`, `@DOC:WORKFLOW-*`, `@DOC:CMD-*`
  - 신규 도메인 (예상): `@DOC:GUIDE-*`, `@DOC:SKILL-*`, `@DOC:STATUS-*`

### 미태깅 파일 분류 (7개 배치)

**Batch 1: Quick Wins** (5개 파일, 6.5시간)
- CLAUDE-AGENTS-GUIDE.md
- CLAUDE-PRACTICES.md
- CLAUDE-RULES.md
- CHANGELO G.md
- README.md

**Batch 2: Skills System** (5개 파일, 5.5시간)
- .claude/skills/.../SKILL.md (5개)

**Batch 3: Architecture** (3개 파일, 10시간)
- .claude/skills/moai-foundation-structure/SKILL.md
- .claude/skills/moai-foundation-product/SKILL.md
- .claude/skills/moai-foundation-tech/SKILL.md

**Batch 4: Concepts** (5개 파일, 17.5시간)
- .claude/skills/.../SKILL.md (5개 개념 스킬)

**Batch 5: Workflows** (6개 파일, 19시간)
- .claude/skills/.../SKILL.md (6개 워크플로우 스킬)

**Batch 6: Tutorials** (7개 파일, 26시간)
- .claude/skills/.../SKILL.md (7개 튜토리얼 스킬)

**Batch 7: Polish** (2개 파일, 3시간)
- .moai/project/structure.md
- .moai/project/tech.md

---

## Assumptions (전제조건)

**ASSUME THAT**:

1. **Phase 1/2 완료**: Phase 1 라이브러리와 Phase 2 워크플로우가 완전히 작동함
2. **사용자 승인 모델**: 각 배치 실행 시 사용자가 TAG 제안을 승인/거부함
3. **백업 시스템 작동**: 파일 수정 전 자동 백업이 생성되고 실패 시 롤백됨
4. **Phase 1.5/2.5 워크플로우**: `/alfred:3-sync` 및 `doc-syncer` 워크플로우가 정상 작동함
5. **배치 단위 실행**: 7개 배치를 순차적으로 실행하며, 각 배치는 독립적으로 승인/거부 가능
6. **도메인 규칙 준수**: 신규 도메인 생성 시 기존 TAG 규칙과 일관성 유지
7. **품질 우선**: 시간 예측은 참고용이며, 품질 검증 통과가 최우선
8. **점진적 검증**: 각 배치 완료 후 TAG 인벤토리 및 체인 무결성 검증

---

## Requirements (요구사항)

### Ubiquitous Requirements (보편 요구사항)

**THE SYSTEM SHALL**:

1. **모든 파일 태깅**: 33개 미태깅 파일에 대해 @DOC 태그를 자동으로 생성해야 함
2. **사용자 승인 필수**: 모든 TAG 삽입은 사용자 승인 후에만 실행되어야 함
3. **백업 보장**: 파일 수정 전 백업을 생성하고, 실패 시 자동 롤백해야 함
4. **도메인 일관성**: 신규 도메인 생성 시 기존 TAG 규칙과 일관성을 유지해야 함
5. **TAG ID 고유성**: 중복 TAG ID가 생성되지 않도록 TAG 인벤토리를 검증해야 함
6. **체인 무결성**: SPEC-DOC 매핑 시 Chain 참조를 포함하고 추적 가능해야 함
7. **품질 검증**: 각 배치 완료 후 TRUST 5 원칙에 따라 품질 게이트를 통과해야 함

### Event-Driven Requirements (이벤트 기반 요구사항)

**WHEN** 배치 X 실행이 시작되면 **THE SYSTEM SHALL**:

1. **파일 스캔**: 배치에 포함된 파일 목록을 스캔해야 함
2. **TAG 제안 생성**: 각 파일에 대해 @DOC TAG를 신뢰도 점수와 함께 제안해야 함
3. **사용자 승인 요청**: `AskUserQuestion`을 통해 배치 단위 승인/거부 선택을 제공해야 함

**WHEN** 사용자가 배치 X를 승인하면 **THE SYSTEM SHALL**:

4. **백업 생성**: 배치의 모든 파일에 대해 백업을 생성해야 함
5. **TAG 삽입**: 승인된 TAG를 마크다운 헤더에 삽입해야 함
6. **인벤토리 업데이트**: TAG 인벤토리를 업데이트하고 중복 검증해야 함
7. **진행 상태 보고**: 배치 진행 상태를 사용자에게 실시간으로 보고해야 함

**WHEN** 배치 X 실행이 완료되면 **THE SYSTEM SHALL**:

8. **품질 검증**: TAG 형식, 체인 무결성, 중복 검사를 수행해야 함
9. **완료 리포트 생성**: 배치 결과 리포트 (성공/실패/건너뜀)를 생성해야 함
10. **다음 배치 안내**: 다음 배치로 진행 또는 중단 옵션을 제공해야 함

**WHEN** 사용자가 배치 X를 거부하면 **THE SYSTEM SHALL**:

11. **파일 무수정**: 해당 배치의 파일을 수정하지 않아야 함
12. **다음 배치 제안**: 다음 배치로 건너뛰거나 중단 옵션을 제공해야 함

### State-Driven Requirements (상태 기반 요구사항)

**WHILE** 배치 처리가 진행 중인 동안 **THE SYSTEM SHALL**:

1. **백업 유지**: 모든 백업 파일을 `.moai/backups/` 디렉토리에 유지해야 함
2. **진행 상태 추적**: 현재 배치, 처리된 파일 수, 남은 파일 수를 표시해야 함
3. **인벤토리 동기화**: TAG 인벤토리를 실시간으로 업데이트해야 함

**WHILE** TAG 인벤토리가 업데이트되는 동안 **THE SYSTEM SHALL**:

4. **중복 검증**: 새로운 TAG ID가 기존 ID와 중복되지 않는지 확인해야 함
5. **도메인 일관성 검증**: 새로운 도메인이 명명 규칙을 준수하는지 확인해야 함

**WHILE** 전체 마이그레이션이 진행 중인 동안 **THE SYSTEM SHALL**:

6. **진행률 표시**: 전체 진행률 (예: "Batch 3/7, 18/33 files tagged")을 표시해야 함
7. **롤백 준비**: 언제든지 이전 배치로 롤백 가능한 상태를 유지해야 함

### Optional Requirements (선택적 요구사항)

**THE SYSTEM MAY**:

1. **배치 병합**: 사용자 요청 시 여러 배치를 하나로 병합하여 실행할 수 있음
2. **우선순위 변경**: 사용자 요청 시 배치 실행 순서를 변경할 수 있음
3. **상세 리포트**: 각 파일의 TAG 생성 이유와 신뢰도 점수를 상세 리포트로 제공할 수 있음
4. **자동 커밋**: 각 배치 완료 후 자동으로 Git 커밋을 생성할 수 있음 (Phase 4 후보)
5. **CLI 유틸리티**: 수동 배치 실행 및 검증을 위한 CLI 명령을 제공할 수 있음

---

## Unwanted Behaviors (제약조건 및 원하지 않는 동작)

**IF** 백업 생성이 실패하면 **THE SYSTEM SHALL**:

1. **TAG 삽입 중단**: 해당 배치의 TAG 삽입을 즉시 중단해야 함
2. **에러 리포트**: 실패 원인을 사용자에게 명확히 보고해야 함
3. **다음 배치 차단**: 다음 배치로 진행하지 않아야 함

**IF** TAG ID 중복이 감지되면 **THE SYSTEM SHALL**:

4. **자동 증분**: TAG ID를 자동으로 증분 (예: `GUIDE-001` → `GUIDE-002`)해야 함
5. **사용자 알림**: 중복 감지 및 자동 증분 사실을 사용자에게 알려야 함

**IF** 배치 실행 중 에러가 발생하면 **THE SYSTEM SHALL**:

6. **롤백 실행**: 해당 배치의 모든 변경 사항을 자동 롤백해야 함
7. **에러 로그**: 에러 로그를 `.moai/logs/`에 저장해야 함
8. **안전 모드**: 다음 배치 실행 전 사용자 확인을 요청해야 함

**IF** SPEC 매핑 신뢰도가 0.5 이하이면 **THE SYSTEM SHALL**:

9. **수동 검토 요청**: 사용자에게 수동 검토를 요청해야 함
10. **Chain 생략**: Chain 참조를 생략하고 단순 TAG만 삽입해야 함

**IF** 사용자가 전체 마이그레이션을 중단하면 **THE SYSTEM SHALL**:

11. **현재 배치 완료**: 진행 중인 배치를 완료하거나 롤백해야 함
12. **상태 저장**: 중단 시점의 상태를 `.moai/memory/migration-state.json`에 저장해야 함
13. **재개 가능**: 나중에 동일한 지점에서 재개 가능해야 함

---

## Specifications (세부 사양)

### 배치 실행 전략

#### Batch 1: Quick Wins (5개 파일, 6.5시간)

**목표**: 프로젝트 최상위 문서 태깅 (가시성 높음, 난이도 낮음)

**파일 목록**:
1. `CLAUDE-AGENTS-GUIDE.md` → `@DOC:GUIDE-AGENT-001`
2. `CLAUDE-PRACTICES.md` → `@DOC:GUIDE-PRACTICE-001`
3. `CLAUDE-RULES.md` → `@DOC:GUIDE-RULES-001`
4. `CHANGELOG.md` → `@DOC:STATUS-CHANGELOG-001`
5. `README.md` → `@DOC:STATUS-README-001`

**신규 도메인**:
- `@DOC:GUIDE-*`: 사용자 가이드 문서
- `@DOC:STATUS-*`: 프로젝트 상태 문서 (README, CHANGELOG)

**실행 조건**:
- Phase 2 워크플로우 정상 작동 확인
- 백업 시스템 테스트 완료

**검증 기준**:
- 5개 파일 모두 TAG 삽입 성공
- TAG ID 중복 없음
- TAG 인벤토리 업데이트 확인

---

#### Batch 2: Skills System (5개 파일, 5.5시간)

**목표**: Foundation Tier Skill 문서 태깅

**파일 목록**:
1. `moai-foundation-ears/SKILL.md` → `@DOC:SKILL-EARS-001`
2. `moai-foundation-specs/SKILL.md` → `@DOC:SKILL-SPECS-001`
3. `moai-foundation-tags/SKILL.md` → `@DOC:SKILL-TAGS-001`
4. `moai-foundation-trust/SKILL.md` → `@DOC:SKILL-TRUST-001`
5. `moai-foundation-hooks/SKILL.md` → `@DOC:SKILL-HOOKS-001`

**신규 도메인**:
- `@DOC:SKILL-*`: Skill 시스템 문서

**실행 조건**:
- Batch 1 완료
- Skill 도메인 명명 규칙 확정

**검증 기준**:
- 5개 Skill 문서 모두 TAG 삽입 성공
- `moai-foundation-tags` Skill 업데이트 (자기 참조 TAG 포함)

---

#### Batch 3: Architecture (3개 파일, 10시간)

**목표**: 프로젝트 아키텍처 핵심 Skill 태깅

**파일 목록**:
1. `moai-foundation-structure/SKILL.md` → `@DOC:SKILL-STRUCTURE-001`
2. `moai-foundation-product/SKILL.md` → `@DOC:SKILL-PRODUCT-001`
3. `moai-foundation-tech/SKILL.md` → `@DOC:SKILL-TECH-001`

**실행 조건**:
- Batch 2 완료
- 아키텍처 도메인 SPEC 매핑 확인

**검증 기준**:
- 3개 아키텍처 Skill 모두 TAG 삽입 성공
- Chain 참조 포함 (예: `@SPEC:PROJECT-001 -> @DOC:SKILL-STRUCTURE-001`)

---

#### Batch 4: Concepts (5개 파일, 17.5시간)

**목표**: 개념 설명 Skill 태깅

**파일 목록**:
1. `moai-essentials-context/SKILL.md` → `@DOC:SKILL-CONTEXT-001`
2. `moai-essentials-workflow/SKILL.md` → `@DOC:SKILL-WORKFLOW-001`
3. `moai-alfred-ears-authoring/SKILL.md` → `@DOC:SKILL-EARS-AUTHOR-001`
4. `moai-alfred-spec-metadata-validation/SKILL.md` → `@DOC:SKILL-SPEC-META-001`
5. `moai-alfred-tag-scanning/SKILL.md` → `@DOC:SKILL-TAG-SCAN-001`

**실행 조건**:
- Batch 3 완료
- 개념 도메인 분류 완료

**검증 기준**:
- 5개 개념 Skill 모두 TAG 삽입 성공
- Skill Tier 간 일관성 검증

---

#### Batch 5: Workflows (6개 파일, 19시간)

**목표**: 워크플로우 관련 Skill 태깅

**파일 목록**:
1. `moai-alfred-plan-workflow/SKILL.md` → `@DOC:SKILL-PLAN-WF-001`
2. `moai-alfred-run-workflow/SKILL.md` → `@DOC:SKILL-RUN-WF-001`
3. `moai-alfred-sync-workflow/SKILL.md` → `@DOC:SKILL-SYNC-WF-001`
4. `moai-alfred-project-workflow/SKILL.md` → `@DOC:SKILL-PROJECT-WF-001`
5. `moai-alfred-trust-validation/SKILL.md` → `@DOC:SKILL-TRUST-VAL-001`
6. `moai-alfred-interactive-questions/SKILL.md` → `@DOC:SKILL-INTERACTIVE-001`

**실행 조건**:
- Batch 4 완료
- 워크플로우 도메인 SPEC 매핑 확인

**검증 기준**:
- 6개 워크플로우 Skill 모두 TAG 삽입 성공
- Chain 참조 포함 (예: `@SPEC:WORKFLOW-001 -> @DOC:SKILL-PLAN-WF-001`)

---

#### Batch 6: Tutorials (7개 파일, 26시간)

**목표**: 튜토리얼 및 고급 Skill 태깅

**파일 목록**:
1. `moai-domain-python/SKILL.md` → `@DOC:SKILL-PYTHON-001`
2. `moai-domain-typescript/SKILL.md` → `@DOC:SKILL-TYPESCRIPT-001`
3. `moai-ops-git/SKILL.md` → `@DOC:SKILL-GIT-001`
4. `moai-ops-ci-cd/SKILL.md` → `@DOC:SKILL-CICD-001`
5. `moai-language-korean/SKILL.md` → `@DOC:SKILL-KOREAN-001`
6. `moai-language-japanese/SKILL.md` → `@DOC:SKILL-JAPANESE-001`
7. `moai-language-spanish/SKILL.md` → `@DOC:SKILL-SPANISH-001`

**실행 조건**:
- Batch 5 완료
- 튜토리얼 도메인 분류 완료

**검증 기준**:
- 7개 튜토리얼 Skill 모두 TAG 삽입 성공
- Language Tier 일관성 검증

---

#### Batch 7: Polish (2개 파일, 3시간)

**목표**: 프로젝트 메타 문서 태깅 (마이그레이션 완료)

**파일 목록**:
1. `.moai/project/structure.md` → `@DOC:PROJECT-STRUCTURE-001`
2. `.moai/project/tech.md` → `@DOC:PROJECT-TECH-001`

**실행 조건**:
- Batch 6 완료
- 모든 배치 품질 검증 통과

**검증 기준**:
- 2개 프로젝트 문서 모두 TAG 삽입 성공
- **최종 검증**: 78/78 파일 모두 태깅 완료 (100%)

---

### TAG 생성 규칙

#### 신규 도메인 명명 규칙

| 도메인 | 형식 | 설명 | 예시 |
|--------|------|------|------|
| `@DOC:GUIDE-*` | `@DOC:GUIDE-{TOPIC}-NNN` | 사용자 가이드 문서 | `@DOC:GUIDE-AGENT-001` |
| `@DOC:SKILL-*` | `@DOC:SKILL-{SKILL_NAME}-NNN` | Skill 시스템 문서 | `@DOC:SKILL-EARS-001` |
| `@DOC:STATUS-*` | `@DOC:STATUS-{TYPE}-NNN` | 프로젝트 상태 문서 | `@DOC:STATUS-README-001` |
| `@DOC:PROJECT-*` | `@DOC:PROJECT-{ASPECT}-NNN` | 프로젝트 메타 문서 | `@DOC:PROJECT-STRUCTURE-001` |

#### TAG 포맷 표준

**SPEC 매핑 있음**:
```markdown
# @DOC:GUIDE-AGENT-001 | Chain: @SPEC:DOC-TAG-003 -> @DOC:GUIDE-AGENT-001

# MoAI-ADK Agent System Guide
```

**SPEC 매핑 없음** (신뢰도 < 0.5):
```markdown
# @DOC:STATUS-README-001

# MoAI-ADK - MoAI-Agentic Development Kit
```

---

### 배치 실행 워크플로우

```
사용자: /alfred:3-sync
    ↓
Phase 1.5: TAG 할당 체크
    ├─ 33개 미태깅 파일 스캔
    ├─ Batch 1 제안 표시 (5개 파일)
    └─ AskUserQuestion: "Batch 1 (Quick Wins) 실행? [Y/n]"
    ↓
사용자 승인
    ↓
Phase 2.5: doc-syncer 자동 생성
    ├─ Batch 1 파일 백업 생성
    ├─ @DOC TAG 삽입 (신뢰도 기반)
    ├─ TAG 인벤토리 업데이트
    └─ 품질 검증 (TRUST 5)
    ↓
Batch 1 완료 리포트
    ├─ 성공: 5개 파일
    ├─ 실패: 0개 파일
    └─ AskUserQuestion: "다음 배치 (Batch 2) 진행? [Y/n]"
    ↓
반복 (Batch 2 → Batch 7)
    ↓
최종 검증
    ├─ 78/78 파일 태깅 확인
    ├─ TAG 체인 무결성 검증
    └─ 도메인 일관성 검증
    ↓
Phase 3 완료 (100%)
```

---

### 백업 및 롤백 전략

#### 백업 구조

```
.moai/backups/
├── batch-1/
│   ├── CLAUDE-AGENTS-GUIDE.md.backup
│   ├── CLAUDE-PRACTICES.md.backup
│   ├── CLAUDE-RULES.md.backup
│   ├── CHANGELOG.md.backup
│   └── README.md.backup
├── batch-2/
│   └── [Skill 파일 백업]
└── migration-state.json  # 진행 상태 저장
```

#### 롤백 조건

1. **자동 롤백**:
   - 백업 생성 실패
   - TAG 삽입 중 에러 발생
   - TAG ID 중복 감지 (자동 증분 실패 시)

2. **수동 롤백**:
   - 사용자 요청 (`/alfred:3-sync --rollback batch-3`)
   - 품질 검증 실패

#### 롤백 프로세스

```bash
# Batch 3 롤백 예시
1. .moai/backups/batch-3/ 디렉토리 스캔
2. 각 백업 파일을 원본 위치로 복원
3. TAG 인벤토리에서 Batch 3 TAG 제거
4. migration-state.json 업데이트 (Batch 3 상태: "rolled_back")
```

---

### 품질 검증 체크리스트

#### 각 배치 완료 후

- ✅ 모든 파일에 TAG 삽입 성공
- ✅ TAG ID 중복 없음
- ✅ TAG 포맷 표준 준수
- ✅ Chain 참조 무결성 (SPEC 매핑 있는 경우)
- ✅ TAG 인벤토리 업데이트 확인
- ✅ 백업 파일 존재 확인

#### Phase 3 전체 완료 후

- ✅ **78/78 파일 모두 태깅 완료** (100%)
- ✅ TAG ID 전역 고유성 검증
- ✅ 도메인 명명 규칙 일관성 검증
- ✅ Chain 참조 전체 추적 가능
- ✅ TAG 인벤토리 최종 검증
- ✅ `.moai/memory/tag-registry.json` 업데이트

---

## Traceability (@TAG)

### SPEC 태그
- **@SPEC:DOC-TAG-003**: Phase 3 - Batch Migration

### 관련 SPEC
- **@SPEC:DOC-TAG-001**: Phase 1 - Library Infrastructure (의존성)
- **@SPEC:DOC-TAG-002**: Phase 2 - Agent Integration (의존성)
- **@SPEC:DOC-TAG-004**: Phase 4 - CLI & Automation (후속)

### 수정할 파일 (없음)
- Phase 3는 기존 파일 수정 없이 33개 파일에 TAG만 삽입

### TAG 대상 파일 (33개)

**Batch 1** (5개):
- `CLAUDE-AGENTS-GUIDE.md`
- `CLAUDE-PRACTICES.md`
- `CLAUDE-RULES.md`
- `CHANGELOG.md`
- `README.md`

**Batch 2-7** (28개):
- `.claude/skills/.../SKILL.md` (26개)
- `.moai/project/structure.md`
- `.moai/project/tech.md`

### TAG 체인

```
@SPEC:DOC-TAG-001 (Phase 1 - Library)
    ↓
@SPEC:DOC-TAG-002 (Phase 2 - Workflow)
    ↓
@SPEC:DOC-TAG-003 (Phase 3 - Migration) ← 현재 SPEC
    ↓
@SPEC:DOC-TAG-004 (Phase 4 - CLI) ← 계획됨
```

### 생성될 TAG 도메인

**신규 도메인** (3개):
- `@DOC:GUIDE-*`: 사용자 가이드 (5개 파일)
- `@DOC:SKILL-*`: Skill 시스템 (26개 파일)
- `@DOC:STATUS-*`: 프로젝트 상태 (2개 파일)

**확장 도메인**:
- `@DOC:PROJECT-*`: 프로젝트 메타 (2개 파일)

---

**END OF SPEC**
