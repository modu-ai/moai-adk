---
id: INIT-002
version: 0.2.0
status: completed
created: 2025-10-06
updated: 2025-10-06
completed: 2025-10-06
author: @Goos
priority: medium
---

# @SPEC:INIT-002: Session Notice 초기화 체크 로직 Alfred 브랜딩 정렬

## HISTORY

### v0.2.0 (2025-10-06)
- **COMPLETED**: Session Notice Alfred 브랜딩 경로 정렬 완료
- **AUTHOR**: @goos, @alfred
- **IMPLEMENTATION**: `.claude/commands/moai` → `.claude/commands/alfred` 경로 변경 완료
  - session-notice/utils.ts 업데이트
  - Alfred 커맨드 경로 체크 로직 구현
  - 크로스 플랫폼 테스트 추가
- **EVIDENCE**:
  - `2510118 Merge feature/INIT-002: Alfred branding path detection`
  - `bc37263 feat(init): Add SPEC-INIT-002 documentation`
  - `e8cee54 fix(hooks): Update session-notice to check Alfred commands path`
  - @CODE:INIT-002 TAG 발견

### v0.1.0 (2025-10-06)
- **INITIAL**: Session Notice 초기화 체크 로직 Alfred 브랜딩 정렬 명세 작성
- **CONTEXT**: `.claude/commands/moai` → `.claude/commands/alfred` 경로 변경 필요
- **REASON**: Alfred 브랜딩 통일 및 프로젝트 인식 정확도 향상
- **AUTHOR**: @Goos
- **REVIEW**: @AI-Alfred (spec-builder)

---

## Environment (환경 및 전제조건)

### 시스템 환경
- **Node.js**: ≥18.0.0
- **TypeScript**: ≥5.0.0
- **빌드 도구**: tsup
- **배포 대상**: `.claude/hooks/alfred/session-notice.cjs`

### 프로젝트 구조
```
moai-adk-ts/
├── src/claude/hooks/session-notice/
│   ├── index.ts                    # 메인 진입점
│   ├── utils.ts                    # 유틸리티 함수 (수정 대상)
│   └── types.ts                    # 타입 정의
├── tsup.hooks.config.ts            # 빌드 설정
└── templates/.claude/hooks/alfred/ # 빌드 출력
```

### 관련 파일
- **수정 대상**: `moai-adk-ts/src/claude/hooks/session-notice/utils.ts:21-28`
- **빌드 출력**: `templates/.claude/hooks/alfred/session-notice.cjs`
- **최종 배포**: `.claude/hooks/alfred/session-notice.cjs`

---

## Assumptions (가정사항)

### 기술적 가정
1. MoAI-ADK 프로젝트는 `.moai` 디렉토리를 필수로 가진다
2. Alfred 명령어는 `.claude/commands/alfred` 디렉토리에 위치한다
3. 기존 `moai` 디렉토리는 레거시이며, `alfred`로 전환 중이다

### 비즈니스 가정
1. 모든 사용자는 `/alfred:8-project` 실행으로 초기화를 완료한다
2. Session Notice Hook은 매 세션 시작 시 자동 실행된다
3. 프로젝트 인식 실패 시 초기화 안내 메시지를 표시한다

### 호환성 가정
1. 기존 `moai` 경로 체크는 레거시 지원을 위해 유지할 수 있다 (선택적)
2. Alfred 브랜딩 이후 새 프로젝트는 모두 `alfred` 경로를 사용한다

---

## EARS 요구사항

### Ubiquitous Requirements (기본 기능)
- 시스템은 MoAI 프로젝트 여부를 `.moai` + `.claude/commands/alfred` 존재로 판단해야 한다
- 시스템은 Session Notice Hook에서 `isMoAIProject()` 함수를 사용해야 한다
- 시스템은 TypeScript 원본 수정 후 반드시 빌드를 실행해야 한다

### Event-driven Requirements (이벤트 기반)
- WHEN 세션이 시작되면, 시스템은 `isMoAIProject(projectRoot)`를 호출해야 한다
- WHEN 두 경로가 모두 존재하면, 시스템은 `true`를 반환해야 한다
- WHEN 하나라도 없으면, 시스템은 `false`를 반환하고 초기화 안내를 표시해야 한다

### State-driven Requirements (상태 기반)
- WHILE 프로젝트가 MoAI로 인식되면, 시스템은 현재 SPEC 상태를 표시해야 한다
- WHILE 프로젝트가 MoAI로 인식되지 않으면, 시스템은 `/alfred:8-project` 실행을 권장해야 한다

### Constraints (제약사항)
- IF TypeScript 원본을 수정하면, 시스템은 반드시 `npm run build:hooks`를 실행해야 한다
- IF 빌드를 생략하면, 시스템은 `.claude/hooks`에 변경사항이 반영되지 않는다
- IF 경로 체크 실패 시, 시스템은 절대 잘못된 프로젝트를 MoAI로 인식해서는 안 된다

---

## Specifications (상세 명세)

### 기능 명세

#### 1. `isMoAIProject()` 함수 수정

**위치**: `moai-adk-ts/src/claude/hooks/session-notice/utils.ts:21-28`

**현재 코드** (변경 전):
```typescript
export function isMoAIProject(projectRoot: string): boolean {
  const requiredPaths = [
    path.join(projectRoot, '.moai'),
    path.join(projectRoot, '.claude', 'commands', 'moai'),  // ❌ 레거시 경로
  ];

  return requiredPaths.every(p => fs.existsSync(p));
}
```

**변경 후**:
```typescript
export function isMoAIProject(projectRoot: string): boolean {
  const moaiDir = path.join(projectRoot, '.moai');
  const alfredCommands = path.join(projectRoot, '.claude', 'commands', 'alfred');  // ✅ Alfred 경로

  return fs.existsSync(moaiDir) && fs.existsSync(alfredCommands);
}
```

**변경 사유**:
1. **브랜딩 정렬**: Alfred SuperAgent 브랜딩과 일치
2. **명확성**: 변수명으로 의도 명확화
3. **단순성**: 배열 대신 직접 체크로 가독성 향상

#### 2. 빌드 프로세스

**명령어**:
```bash
npm run build:hooks
```

**빌드 설정**: `tsup.hooks.config.ts`
- **입력**: `src/claude/hooks/session-notice/index.ts`
- **출력**: `templates/.claude/hooks/alfred/session-notice.cjs`
- **포맷**: CommonJS (Node.js 호환)

**최종 배포**:
- 빌드된 `.cjs` 파일을 `.claude/hooks/alfred/session-notice.cjs`로 복사
- Claude Code가 세션 시작 시 자동 실행

### 비기능 명세

#### 성능
- 파일 존재 체크: O(1) 시간복잡도 유지
- 세션 시작 지연: <100ms

#### 보안
- 경로 순회(Path Traversal) 공격 방지: `path.join()` 사용
- 절대 경로 검증: `projectRoot` 파라미터 신뢰성 확인

#### 유지보수성
- 함수 복잡도: ≤3 (매우 단순)
- 의존성: Node.js 표준 라이브러리만 사용 (`fs`, `path`)

---

## Traceability (추적성)

### TAG 체인
- **@SPEC:INIT-002**: 본 문서
- **@CODE:INIT-002**: `moai-adk-ts/src/claude/hooks/session-notice/utils.ts:21-28`
- **@TEST:INIT-002**: 수동 검증 (세션 시작 시 프로젝트 인식 확인)
- **@DOC:INIT-002**: 본 SPEC 문서 (Living Document 자동 생성)

### 관련 SPEC
- **SPEC-SESSION-NOTICE-001**: Session Notice Hook 전체 시스템
- **SPEC-INIT-001**: 비대화형 모드 지원 (프로젝트 초기화)

### 의존성
- **선행 조건**: `/alfred:8-project` 실행 완료 (`.moai`, `.claude/commands/alfred` 생성)
- **후속 작업**: 없음 (독립 실행 가능)

---

## Definition of Done (완료 조건)

### 필수 체크리스트
- [ ] `utils.ts:21-28` 코드 수정 완료
- [ ] `@CODE:INIT-002` TAG 주석 추가
- [ ] `npm run build:hooks` 빌드 성공
- [ ] `.claude/hooks/alfred/session-notice.cjs` 배포 확인
- [ ] 새 세션 시작 시 프로젝트 인식 정상 동작 확인

### 검증 시나리오
1. **정상 인식**: `.moai` + `.claude/commands/alfred` 존재 → `true` 반환
2. **초기화 필요**: 둘 중 하나라도 없음 → `false` 반환 및 안내 메시지
3. **빌드 파일 검증**: `session-notice.cjs`에 `alfred` 포함, `moai` 미포함

---

_SPEC-INIT-002 작성 완료 | Alfred SuperAgent 브랜딩 정렬 명세_
