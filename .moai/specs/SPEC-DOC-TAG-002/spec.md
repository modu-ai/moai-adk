---
id: DOC-TAG-002
version: 0.0.1
status: implementation-complete
created: 2025-10-29
updated: 2025-10-29
author: "@Goos"
priority: high
category: Integration / Workflow
labels: [documentation, tags, agents, workflow]
depends_on: [DOC-TAG-001]
scope: "Phase 2 of 4-phase @DOC TAG automatic generation system"
---

# @SPEC:DOC-TAG-002: @DOC 태그 자동 생성 - 에이전트 통합 (Phase 2)

## HISTORY

### v0.0.1 (2025-10-29)
- **INITIAL**: Phase 1 라이브러리를 MoAI-ADK 워크플로우에 통합하는 두 번째 단계
- **AUTHOR**: @Goos
- **SCOPE**: doc-syncer 에이전트와 /alfred:3-sync 명령 통합, Skill 업데이트
- **CONTEXT**: Phase 1 완료 후 자동화 워크플로우 구축, 사용자 상호작용 포함

---

## Environment (환경)

**WHEN** MoAI-ADK Phase 2 통합 환경에서:

- **Phase 1 라이브러리**: 완전히 작동하는 @DOC TAG 생성 라이브러리 (90.5% 테스트 커버리지)
  - `DocumentParser`: 마크다운 파싱 및 TAG 추출
  - `DocTagGenerator`: 도메인 기반 TAG ID 생성
  - `SpecDocMapper`: SPEC-DOC 매핑 및 신뢰도 점수
  - `TagInserter`: 마크다운 헤더에 TAG 삽입
  - `TagRegistry`: TAG 인벤토리 관리

- **MoAI-ADK 에이전트 시스템**:
  - `doc-syncer`: 문서 동기화 전문 에이전트
  - `tag-agent`: TAG 검증 및 체인 추적 에이전트
  - `alfred`: 워크플로우 오케스트레이션 SuperAgent

- **명령 시스템**:
  - `/alfred:3-sync`: 문서 동기화 명령 (Phase 0-3 워크플로우)

- **Skill 시스템**:
  - `moai-foundation-tags`: TAG 체계 및 Best Practice 문서

- **사용자 상호작용 도구**:
  - `AskUserQuestion`: TUI 기반 사용자 승인/거부 도구

---

## Assumptions (전제조건)

**ASSUME THAT**:

1. **Phase 1 완료**: Phase 1 라이브러리가 완전히 구현되고 테스트되었음
2. **사용자 승인 모델**: 사용자는 TAG 생성 전 제안을 승인/거부할 수 있음
3. **워크플로우 호환성**: 기존 doc-syncer 및 /alfred:3-sync 워크플로우는 변경 없이 유지됨
4. **파일 안전성**: TAG 삽입 전 파일 백업으로 안전성 보장
5. **신뢰도 기반 제안**: SPEC-DOC 매핑 신뢰도 점수에 따라 사용자 개입 결정
6. **점진적 통합**: Phase 1.5와 Phase 2.5로 기존 워크플로우에 자연스럽게 통합

---

## Requirements (요구사항)

### Ubiquitous Requirements (보편 요구사항)

**THE SYSTEM SHALL**:

1. **자동 TAG 제안**: 태그 없는 문서 파일에 대해 @DOC 태그를 자동으로 제안해야 함
2. **사용자 승인 우선**: 사용자 승인 이전에 TAG 생성을 적용하지 않아야 함
3. **기존 워크플로우 호환**: 기존 doc-syncer 워크플로우와 호환성을 유지해야 함
4. **수동 TAG 존중**: 수동으로 생성된 @DOC 태그와 충돌하지 않아야 함
5. **파일 백업**: TAG 삽입 전 파일 백업을 생성해야 함

### Event-Driven Requirements (이벤트 기반 요구사항)

**WHEN** `/alfred:3-sync` 명령이 실행되면 **THE SYSTEM SHALL**:

1. **Phase 1.5 TAG 할당 체크**: 태그 없는 문서 파일을 확인해야 함
2. **TAG 제안 표시**: 태그 없는 파일 발견 시, @DOC TAG를 신뢰도 점수와 함께 제안해야 함
3. **사용자 상호작용**: `AskUserQuestion`을 통해 승인/거부 선택을 제공해야 함

**WHEN** 사용자가 TAG 제안을 승인하면 **THE SYSTEM SHALL**:

4. **Phase 2.5 자동 생성**: doc-syncer Phase 2.5에서 @DOC TAG를 마크다운 파일에 삽입해야 함
5. **백업 관리**: 파일 수정 전 백업을 생성하고, 성공 시 원본 삭제 및 백업 보관해야 함
6. **인벤토리 업데이트**: TAG 인벤토리를 업데이트해야 함

**WHEN** doc-syncer가 실행되면 **THE SYSTEM SHALL**:

7. **자동 TAG 생성**: 누락된 @DOC TAG를 자동으로 생성해야 함 (Phase 2.5)
8. **체인 참조 추가**: SPEC-DOC 매핑 시 Chain 참조를 포함해야 함 (예: `@SPEC:AUTH-001 -> @DOC:AUTH-001`)

### State-Driven Requirements (상태 기반 요구사항)

**WHILE** 문서 파일이 @DOC 태그가 없는 동안 **THE SYSTEM SHALL**:

1. **제안 목록 표시**: 제안 목록에 파일을 표시해야 함

**WHILE** TAG 생성이 진행 중인 동안 **THE SYSTEM SHALL**:

2. **백업 생성**: 파일 수정 전 백업을 생성해야 함
3. **진행 상태 표시**: 생성 진행 상태를 사용자에게 표시해야 함

**WHILE** 동기화가 실행 중인 동안 **THE SYSTEM SHALL**:

4. **TAG 인벤토리 업데이트**: TAG 인벤토리를 실시간으로 업데이트해야 함

### Optional Requirements (선택적 요구사항)

**THE SYSTEM MAY**:

1. **CLI 명령 제공**: 수동 TAG 생성 및 검증을 위한 CLI 명령을 제공할 수 있음
   - `moai-adk tag-generate docs/`: 누락된 TAG 스캔 및 생성
   - `moai-adk tag-validate docs/`: 기존 TAG 검증
   - `moai-adk tag-map SPEC-ID docs/`: SPEC-DOC 매핑 표시

2. **상세 리포트 생성**: TAG 생성 결정에 대한 상세 리포트를 생성할 수 있음

3. **Pre-commit Hook 통합**: pre-commit hook과 통합하여 커밋 전 TAG 자동 생성할 수 있음

---

## Unwanted Behaviors (제약조건 및 원하지 않는 동작)

**IF** 사용자가 TAG 제안을 거부하면 **THE SYSTEM SHALL**:

1. **파일 무수정**: 어떤 파일도 수정하지 않아야 함
2. **워크플로우 계속**: 기존 워크플로우를 계속 진행해야 함

**IF** 파일 백업이 실패하면 **THE SYSTEM SHALL**:

3. **TAG 삽입 중단**: TAG 삽입을 중단해야 함
4. **변경 없이 반환**: 어떤 변경 사항도 없이 반환해야 함

**IF** SPEC 매핑 신뢰도가 임계값(0.5) 이하이면 **THE SYSTEM SHALL**:

5. **수동 개입 요청**: 사용자에게 수동 개입을 요청해야 함
6. **자동 매핑 생략**: 자동 매핑을 생략해야 함

**IF** 100개 이상의 파일에서 성능 저하가 감지되면 **THE SYSTEM SHALL**:

7. **증분 스캔 사용**: 전체 스캔 대신 증분 스캔을 사용해야 함
8. **배치 처리**: 파일을 배치로 나누어 처리해야 함

---

## Specifications (세부 사양)

### Phase 1.5: TAG 할당 체크 (in `/alfred:3-sync`)

**위치**: `.claude/commands/alfred/3-sync.md`

**워크플로우**:
```
Phase 0: 프로젝트 분석 (기존)
    ↓
Phase 1.5: TAG 할당 체크 (NEW)
    ├─ docs/ 디렉토리 스캔
    ├─ suggest_tag_for_file() 호출
    ├─ 신뢰도별 분류 (HIGH/MEDIUM/LOW)
    ├─ 제안 목록 표시
    └─ AskUserQuestion: "N개 파일에 TAG 생성?"
    ↓
Phase 1: doc-syncer 호출 (이제 Phase 2.5 포함)
    ↓
Phase 2: TAG 검증 (기존)
```

**입력**: `docs/` 디렉토리의 모든 마크다운 파일
**출력**: TAG 제안 목록 (신뢰도 점수 포함)
**사용자 상호작용**: `AskUserQuestion` - "10개 파일에 @DOC TAG를 생성하시겠습니까? [Y/n]"

### Phase 2.5: @DOC TAG 자동 생성 (in `doc-syncer`)

**위치**: `.claude/agents/alfred/doc-syncer.md`

**워크플로우**:
```
Phase 1: 상태 분석 (기존)
    ↓
Phase 2: 문서 동기화 (기존)
    ↓
Phase 2.5: @DOC TAG 자동 생성 (NEW)
    ├─ docs/ 스캔 → 태그 없는 파일
    ├─ generate_doc_tag() → TAG ID 생성
    ├─ find_related_spec() → SPEC-DOC 매핑
    ├─ AskUserQuestion → 승인/거부
    ├─ insert_tag_to_markdown() → TAG 삽입
    ├─ 백업 관리 (생성/삭제/보관)
    └─ TAG 인벤토리 업데이트
    ↓
Phase 3: 품질 검증 (기존)
```

**TAG 포맷**:
```markdown
# @DOC:DOMAIN-NNN | Chain: @SPEC:SOURCE-ID -> @DOC:DOMAIN-NNN

또는 (SPEC 매핑 없는 경우)

# @DOC:DOMAIN-NNN
```

**예시**:
```markdown
# @DOC:AUTH-001 | Chain: @SPEC:AUTH-001 -> @DOC:AUTH-001

# 사용자 인증 가이드

...
```

### 신뢰도 점수 기준

| 점수 범위 | 레벨 | 의미 | 처리 방식 |
|-----------|------|------|-----------|
| 0.8 - 1.0 | HIGH | 도메인 완전 일치, 파일명 일치 | 자동 제안 (사용자 승인 필요) |
| 0.5 - 0.8 | MEDIUM | 도메인 부분 일치, 키워드 일치 | 제안 + 확인 요청 |
| < 0.5 | LOW | 매핑 불확실 | 수동 개입 요청 |

### moai-foundation-tags Skill 업데이트

**위치**: `.claude/skills/moai-foundation-tags/SKILL.md`

**추가 섹션**:
1. **@DOC TAG 포맷 및 체인 참조**
2. **SPEC-DOC 매핑 예제**
3. **신뢰도 점수 설명**
4. **Phase 1.5/2.5 통합 설명**

**위치**: `.claude/skills/moai-foundation-tags/examples.md`

**추가 예제**:
1. **@DOC TAG 생성 예제** (도메인 기반 매핑)
2. **/alfred:3-sync TAG 제안 플로우**
3. **일반적인 문제 해결** (고아 DOC, 중복 ID)

### CLI 유틸리티 (선택사항)

**위치**: `src/moai_adk/cli/tag_commands.py` (~200 LOC)

**명령어**:
```bash
# 누락된 TAG 스캔 및 생성
moai-adk tag-generate docs/

# 기존 TAG 검증
moai-adk tag-validate docs/

# SPEC-DOC 매핑 표시
moai-adk tag-map AUTH-001 docs/
```

**테스트**: `tests/cli/test_tag_commands.py` (~150 LOC)

---

## Traceability (@TAG)

### SPEC 태그
- **@SPEC:DOC-TAG-002**: Phase 2 - Agent Integration

### 관련 SPEC
- **@SPEC:DOC-TAG-001**: Phase 1 - Library Infrastructure (의존성)

### 수정할 파일
- **`.claude/agents/alfred/doc-syncer.md`**: Phase 2.5 섹션 추가
- **`.claude/commands/alfred/3-sync.md`**: Phase 1.5 섹션 추가
- **`.claude/skills/moai-foundation-tags/SKILL.md`**: @DOC 섹션 추가
- **`.claude/skills/moai-foundation-tags/examples.md`**: @DOC 예제 추가

### 생성할 파일 (선택사항)
- **`src/moai_adk/cli/tag_commands.py`**: CLI 유틸리티 (~200 LOC)
- **`tests/cli/test_tag_commands.py`**: CLI 테스트 (~150 LOC)

### TAG 체인
```
@SPEC:DOC-TAG-001 (Phase 1)
    ↓
@SPEC:DOC-TAG-002 (Phase 2) ← 현재 SPEC
    ↓
@SPEC:DOC-TAG-003 (Phase 3) ← 계획됨
    ↓
@SPEC:DOC-TAG-004 (Phase 4) ← 계획됨
```

---

**END OF SPEC**
