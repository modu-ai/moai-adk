# Phase 3 완료 보고서: AskUserQuestion 통합

**작성일**: 2025-10-19
**작성자**: @Alfred
**SPEC**: SPEC-UPDATE-004

---

## ✅ 완료 요약

Phase 3의 모든 작업이 성공적으로 완료되었습니다:
- **Commands**: 4개 전체 업데이트 완료
- **Sub-agents**: 7개 전체 업데이트 완료
- **총 파일 수정**: 11개

---

## 📋 Commands 업데이트 (4개)

모든 커맨드 파일에 **2단계 워크플로우** (Phase 1 계획 → Phase 2 실행) 패턴과 **AskUserQuestion 도구 호출**이 추가되었습니다.

### 1. /alfred:0-project

**파일**: `.claude/commands/alfred/0-project.md`

**추가된 섹션**:
- **Section 1.4**: AskUserQuestion으로 Phase 2 승인 요청
- **Section 2.2**: project-manager의 Nested AskUserQuestion (프로젝트 유형, 팀 모드 등)
- **Section 2.3**: 순차 실행 의존성 (문서 분석 → 인터뷰 → 문서 생성)

**주요 AskUserQuestion 시나리오**:
- Phase 2 승인 (진행/수정/중단)
- 프로젝트 유형 판단 (신규/레거시/하이브리드)
- 팀 모드 설정 (Personal/Team)

### 2. /alfred:1-plan

**파일**: `.claude/commands/alfred/1-plan.md`

**추가된 섹션**:
- **Section 1.4**: AskUserQuestion으로 Phase 2 승인 요청
- **Section 2.1**: Alfred Skills 자동 활성화 (moai-alfred-spec-metadata)
- **Section 2.2**: spec-builder의 Nested AskUserQuestion (SPEC 후보, 파일 충돌 등)
- **Section 2.3**: 순차 실행 의존성 (문서 분석 → SPEC 작성 → Git 작업)

**주요 AskUserQuestion 시나리오**:
- Phase 2 승인 (진행/수정/중단)
- 다중 SPEC 후보 선택
- 기존 SPEC 파일 충돌 처리

### 3. /alfred:2-run

**파일**: `.claude/commands/alfred/2-run.md`

**추가된 섹션**:
- **사용자 승인 단계**: AskUserQuestion으로 Phase 2 승인 요청
- **Section 2.1**: Alfred Skills 자동 활성화 (moai-alfred-trust-validation Level 2)
- **Section 2.2**: tdd-implementer의 Nested AskUserQuestion (테스트 실패 처리 등)
- **Section 2.3**: 순차 실행 의존성 (RED → GREEN → REFACTOR)

**주요 AskUserQuestion 시나리오**:
- Phase 2 승인 (진행/수정/중단)
- 테스트 반복 실패 처리 (5회 실패 시)
- 라이브러리 버전 충돌 해결

### 4. /alfred:3-sync

**파일**: `.claude/commands/alfred/3-sync.md`

**추가된 섹션**:
- **사용자 승인 단계**: AskUserQuestion으로 Phase 2 승인 요청
- **Section 2.1**: Alfred Skills 자동 활성화 (TAG 스캔 + TRUST 검증)
- **Section 2.2**: doc-syncer의 Nested AskUserQuestion (고아 TAG 처리 등)
- **Section 2.3**: 순차 실행 의존성 (TAG 스캔 → 동기화 → Git 작업)

**주요 AskUserQuestion 시나리오**:
- Phase 2 승인 (진행/수정/중단)
- 고아 TAG 처리 (SPEC 생성/TAG 제거/수동 처리)
- 문서-코드 불일치 해결

---

## 🤖 Sub-agents 업데이트 (7개)

모든 에이전트 파일에 **"🤝 사용자 상호작용"** 섹션이 추가되었습니다.

### 1. spec-builder

**파일**: `.claude/agents/alfred/spec-builder.md`

**추가 위치**: `🔗 SPEC 검증 기능` 섹션 이전 (line 51 이전)

**AskUserQuestion 사용 시점** (5가지):
1. 다중 SPEC 후보 발견 시
2. 기존 SPEC 파일 충돌 시
3. EARS 검증 실패 시
4. 프로젝트 문서 누락 시
5. 모호한 요구사항 명확화 시

### 2. tdd-implementer (code-builder)

**파일**: `.claude/agents/alfred/tdd-implementer.md`

**추가 위치**: `🚫 제약사항` 섹션 이전 (line 172 이전)

**AskUserQuestion 사용 시점** (6가지):
1. 테스트 반복 실패 시 (5회 이상)
2. 라이브러리 버전 충돌 시
3. 기존 테스트 파괴 시 (Breaking Change)
4. 커버리지 부족 시 (목표 85% 미달)
5. 복잡도 초과 시 (제한 10 초과)
6. 환경 준비 실패 시

### 3. doc-syncer

**파일**: `.claude/agents/alfred/doc-syncer.md`

**추가 위치**: `@TAG 시스템 동기화` 섹션 이전 (line 117 이전)

**AskUserQuestion 사용 시점** (6가지):
1. 고아 TAG 발견 시
2. 문서-코드 불일치 시
3. 대규모 문서 업데이트 시 (50개 이상 파일)
4. TAG 체인 단절 시
5. 프로젝트 타입 변경 감지 시
6. CHANGELOG 자동 생성 시

### 4. git-manager

**파일**: `.claude/agents/alfred/git-manager.md`

**추가 위치**: 문서 끝 (line 328 이후)

**AskUserQuestion 사용 시점** (7가지):
1. 위험한 Git 작업 시 (force push 등)
2. 머지 충돌 발생 시
3. CI/CD 실패 시
4. 미커밋 변경사항 존재 시
5. GitFlow 규칙 위반 시
6. Auto-merge vs Manual merge 선택 시
7. 오래된 브랜치 정리 시

### 5. debug-helper

**파일**: `.claude/agents/alfred/debug-helper.md`

**추가 위치**: `⚠️ 제약사항` 섹션 이전 (line 117 이전)

**AskUserQuestion 사용 시점** (7가지):
1. 다중 근본 원인 가능성 시
2. 파괴적 수정 제안 시
3. 충돌하는 오류 신호 시
4. 미지의 오류 패턴 시
5. 다중 해결 경로 시
6. 긴급도 평가 시
7. 데이터 손실 위험 시

### 6. cc-manager

**파일**: `.claude/agents/alfred/cc-manager.md`

**추가 위치**: `💡 사용 가이드` 섹션 이전 (line 942 이전)

**AskUserQuestion 사용 시점** (7가지):
1. 파일 생성 시 구현 방식 선택 (Skill/Agent/Command)
2. 표준 위반 수정 방법 선택
3. Plugin 설치 보안 확인
4. settings.json 백업 확인
5. 템플릿 선택
6. 대규모 변경 확인 (10개 이상 파일)
7. Filesystem MCP 경로 확인

### 7. project-manager

**파일**: `.claude/agents/alfred/project-manager.md`

**추가 위치**: `📋 프로젝트 문서 구조 가이드` 섹션 이전 (line 67 이전)

**AskUserQuestion 사용 시점** (7가지):
1. 프로젝트 유형 판단 시 (신규/레거시/하이브리드)
2. 팀 모드 설정 시 (Personal/Team)
3. 누락 문서 생성 시
4. 레거시 분석 깊이 선택 시
5. 기술 스택 확정 시
6. Personal/Team 모드 의심 요소 발견 시
7. 문서 템플릿 선택 시

---

## 📊 통계

### 추가된 AskUserQuestion 예시 총 개수

| 파일 | 카테고리 | 예시 개수 | 주요 시나리오 |
|------|---------|----------|-------------|
| /alfred:0-project | Command | 3 | Phase 승인, 프로젝트 유형, 팀 모드 |
| /alfred:1-plan | Command | 3 | Phase 승인, SPEC 후보, 파일 충돌 |
| /alfred:2-run | Command | 3 | Phase 승인, 테스트 실패, 버전 충돌 |
| /alfred:3-sync | Command | 3 | Phase 승인, 고아 TAG, 문서 불일치 |
| spec-builder | Agent | 5 | SPEC 후보, 충돌, 검증, 누락, 명확화 |
| tdd-implementer | Agent | 6 | 테스트 실패, 충돌, 파괴, 커버리지, 복잡도, 환경 |
| doc-syncer | Agent | 6 | 고아 TAG, 불일치, 대규모, 체인, 타입, CHANGELOG |
| git-manager | Agent | 7 | 위험, 충돌, CI/CD, 변경사항, GitFlow, 머지, 정리 |
| debug-helper | Agent | 7 | 원인, 수정, 신호, 패턴, 경로, 긴급도, 손실 |
| cc-manager | Agent | 7 | 방식, 위반, Plugin, 백업, 템플릿, 대규모, 경로 |
| project-manager | Agent | 7 | 유형, 모드, 문서, 분석, 스택, 의심, 템플릿 |
| **총계** | - | **57** | - |

### 코드 라인 추가 통계

| 파일 카테고리 | 파일 수 | 평균 추가 라인 수 | 총 추가 라인 수 |
|-------------|---------|----------------|---------------|
| Commands | 4 | ~100 라인 | ~400 라인 |
| Agents | 7 | ~150 라인 | ~1050 라인 |
| **총계** | **11** | - | **~1450 라인** |

---

## 🎯 패턴 일관성

모든 파일에서 다음 패턴이 일관되게 적용되었습니다:

### 1. 섹션 제목
```markdown
## 🤝 사용자 상호작용

### AskUserQuestion 사용 시점
```

### 2. 시나리오 구조
```markdown
#### N. [시나리오명]

**상황**: [구체적인 상황 설명]

```typescript
AskUserQuestion({
  questions: [{
    question: "[질문 내용]",
    header: "[헤더]",
    options: [
      { label: "[옵션1]", description: "[설명1]" },
      { label: "[옵션2]", description: "[설명2]" },
      { label: "[옵션3]", description: "[설명3]" }
    ],
    multiSelect: false
  }]
})
```
```

### 3. 사용 원칙
모든 파일에 "### 사용 원칙" 섹션이 포함되어 각 에이전트의 AskUserQuestion 사용 지침을 명시

---

## ✅ 검증 체크리스트

- [x] **Commands**: 4개 파일 모두 업데이트 완료
- [x] **Agents**: 7개 파일 모두 업데이트 완료
- [x] **패턴 일관성**: 모든 파일이 동일한 섹션 구조 사용
- [x] **TypeScript 예시**: 모든 AskUserQuestion이 정확한 TypeScript 구문 사용
- [x] **실무 시나리오**: 각 에이전트의 실제 사용 시나리오 반영
- [x] **multiSelect 활용**: 필요한 경우 multiSelect: true 사용 (예: 추가 정보 수집)
- [x] **명확한 옵션**: 각 옵션에 결과와 영향을 명시
- [x] **사용 원칙**: 모든 파일에 AskUserQuestion 사용 지침 포함

---

## 🔗 Next Steps (Phase 4)

Phase 3 완료 후 다음 단계:

1. **Phase 4: 전체 워크플로우 테스트**
   - Commands 실행 테스트 (모든 AskUserQuestion 동작 확인)
   - Agents 호출 테스트 (Nested AskUserQuestion 동작 확인)
   - Skills 자동 활성화 테스트

2. **문서 검증**
   - CLAUDE.md 내용과 일치성 확인
   - development-guide.md 참조 정확성 확인

3. **최종 PR 준비**
   - feature/SPEC-UPDATE-004 브랜치에서 develop으로 PR 생성
   - PR Ready 전환 및 리뷰 요청

---

**Phase 3 완료 시각**: 2025-10-19
**소요 시간**: 약 2시간
**다음 단계**: Phase 4 테스트 및 검증
