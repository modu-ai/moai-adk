# SPEC-TS-MIGRATION-001 수락 기준

## 개요

이 문서는 MoAI-ADK TypeScript 마이그레이션 프로젝트의 수락 기준을 정의합니다.
모든 시나리오는 Given/When/Then 형식으로 작성되었습니다.

---

## AC-1: 모노레포 초기화

### 시나리오 1.1: Turborepo 프로젝트 생성

```gherkin
Given 빈 디렉토리가 존재할 때
When `bunx create-turbo@latest moai-adk-ts` 명령을 실행하면
Then 다음 구조가 생성되어야 한다:
  - turbo.json 파일
  - package.json (workspaces 설정 포함)
  - apps/ 디렉토리
  - packages/ 디렉토리
And `bun install` 명령이 성공해야 한다
And `bun run build` 명령이 성공해야 한다
```

### 시나리오 1.2: Biome 설정

```gherkin
Given 모노레포가 초기화되어 있을 때
When biome.json 파일을 생성하면
Then `bun run lint` 명령이 성공해야 한다
And `bun run format` 명령이 성공해야 한다
And 모든 TypeScript 파일이 Biome 규칙을 준수해야 한다
```

### 시나리오 1.3: TypeScript 설정

```gherkin
Given 모노레포가 초기화되어 있을 때
When tsconfig.base.json 파일을 생성하면
Then strict 모드가 활성화되어야 한다
And 모든 패키지가 공유 설정을 상속해야 한다
And `bun run typecheck` 명령이 성공해야 한다
```

---

## AC-2: CLI init 명령

### 시나리오 2.1: 새 프로젝트 초기화

```gherkin
Given moai-adk CLI가 설치되어 있을 때
When `moai-adk init my-project` 명령을 실행하면
Then my-project 디렉토리가 생성되어야 한다
And .moai/config/config.yaml 파일이 생성되어야 한다
And .claude/ 디렉토리 구조가 복사되어야 한다
And CLAUDE.md 파일이 생성되어야 한다
```

### 시나리오 2.2: 대화형 초기화

```gherkin
Given moai-adk CLI가 설치되어 있을 때
When `moai-adk init` 명령을 실행하면 (프로젝트명 없이)
Then 프로젝트 이름을 묻는 프롬프트가 표시되어야 한다
And 프로젝트 모드 선택 프롬프트가 표시되어야 한다
And 언어 선택 프롬프트가 표시되어야 한다
And 모든 입력 후 프로젝트가 생성되어야 한다
```

### 시나리오 2.3: 기존 디렉토리 처리

```gherkin
Given my-project 디렉토리가 이미 존재할 때
When `moai-adk init my-project` 명령을 실행하면
Then 경고 메시지가 표시되어야 한다
And 덮어쓰기 여부를 묻는 프롬프트가 표시되어야 한다
And 사용자가 거부하면 작업이 취소되어야 한다
```

---

## AC-3: Dashboard 로드 성능

### 시나리오 3.1: 초기 페이지 로드

```gherkin
Given Dashboard 앱이 실행 중일 때
When 사용자가 / 경로에 접속하면
Then 3초 이내에 페이지가 로드되어야 한다
And 프로젝트 개요가 표시되어야 한다
And HeroUI 컴포넌트가 정상적으로 렌더링되어야 한다
```

### 시나리오 3.2: SPEC 목록 페이지

```gherkin
Given Dashboard 앱이 실행 중일 때
And 10개의 SPEC 파일이 존재할 때
When 사용자가 /specs 경로에 접속하면
Then 2초 이내에 SPEC 목록이 표시되어야 한다
And 각 SPEC의 ID, 제목, 상태가 표시되어야 한다
And 페이지네이션이 동작해야 한다
```

### 시나리오 3.3: 실시간 업데이트

```gherkin
Given Dashboard의 /hooks 페이지가 열려 있을 때
When Claude Code 훅이 실행되면
Then 1초 이내에 훅 로그가 화면에 표시되어야 한다
And SSE 연결이 유지되어야 한다
And 연결 끊김 시 자동 재연결되어야 한다
```

---

## AC-4: Config 로딩

### 시나리오 4.1: YAML 설정 파싱

```gherkin
Given .moai/config/config.yaml 파일이 존재할 때
When ConfigService.load() 메서드를 호출하면
Then Zod 스키마 검증이 통과해야 한다
And 설정 객체가 반환되어야 한다
And 캐시에 저장되어야 한다
```

### 시나리오 4.2: 잘못된 설정 처리

```gherkin
Given config.yaml에 잘못된 값이 있을 때
When ConfigService.load() 메서드를 호출하면
Then Zod 검증 오류가 발생해야 한다
And 오류 메시지에 잘못된 필드가 명시되어야 한다
And 기본값으로 폴백되지 않아야 한다
```

### 시나리오 4.3: 설정 저장

```gherkin
Given 유효한 설정 객체가 있을 때
When ConfigService.save(config) 메서드를 호출하면
Then config.yaml 파일이 업데이트되어야 한다
And 원자적 쓰기가 수행되어야 한다 (임시 파일 → 이름 변경)
And 캐시가 갱신되어야 한다
```

### 시나리오 4.4: 환경 변수 오버라이드

```gherkin
Given MOAI_CONVERSATION_LANG=ko 환경 변수가 설정되어 있을 때
When ConfigService.load() 메서드를 호출하면
Then conversation_language가 'ko'로 설정되어야 한다
And 환경 변수가 파일 설정보다 우선해야 한다
```

---

## AC-5: Hooks 실행

### 시나리오 5.1: SessionStart 훅

```gherkin
Given Claude Code 세션이 시작될 때
When SessionStart 훅이 트리거되면
Then hooks/session-start.ts가 실행되어야 한다
And 프로젝트 정보가 출력되어야 한다
And 15초 이내에 완료되어야 한다
```

### 시나리오 5.2: PreToolUse 훅

```gherkin
Given Claude Code가 도구를 사용하려고 할 때
When PreToolUse 훅이 트리거되면
Then 도구 이름과 파라미터가 로깅되어야 한다
And 훅이 도구 실행을 차단하지 않아야 한다
And 실행 시간이 5초를 초과하지 않아야 한다
```

### 시나리오 5.3: 훅 오류 처리

```gherkin
Given 훅 스크립트에 오류가 있을 때
When 훅이 실행되면
Then 오류가 로그에 기록되어야 한다
And Claude Code 세션이 중단되지 않아야 한다
And 다음 훅이 정상적으로 실행되어야 한다
```

---

## AC-6: SPEC 관리

### 시나리오 6.1: SPEC 생성

```gherkin
Given Dashboard의 /specs 페이지가 열려 있을 때
When 사용자가 "새 SPEC 생성" 버튼을 클릭하면
Then SPEC 생성 폼이 표시되어야 한다
And ID, 제목, 우선순위 필드가 있어야 한다
And 제출 시 .moai/specs/SPEC-{ID}/ 디렉토리가 생성되어야 한다
```

### 시나리오 6.2: SPEC 편집

```gherkin
Given SPEC-001 문서가 존재할 때
When 사용자가 /specs/SPEC-001/edit 페이지에서 내용을 수정하면
Then 변경 사항이 저장되어야 한다
And spec.md의 updated 필드가 갱신되어야 한다
And HISTORY 섹션에 변경 기록이 추가되어야 한다
```

### 시나리오 6.3: SPEC 상태 변경

```gherkin
Given SPEC-001의 상태가 "draft"일 때
When 사용자가 상태를 "in_progress"로 변경하면
Then spec.md의 status 필드가 업데이트되어야 한다
And Git 브랜치가 자동 생성되어야 한다 (설정에 따라)
And Dashboard에 상태 변경이 반영되어야 한다
```

---

## AC-7: Git 통합

### 시나리오 7.1: Git 상태 조회

```gherkin
Given Git 저장소가 초기화되어 있을 때
When GitService.getStatus() 메서드를 호출하면
Then 현재 브랜치 이름이 반환되어야 한다
And 변경된 파일 목록이 반환되어야 한다
And 커밋되지 않은 변경 여부가 반환되어야 한다
```

### 시나리오 7.2: 브랜치 생성

```gherkin
Given main 브랜치가 체크아웃되어 있을 때
When GitService.createBranch("feature/SPEC-001") 메서드를 호출하면
Then feature/SPEC-001 브랜치가 생성되어야 한다
And 해당 브랜치로 자동 전환되어야 한다
And 기존 브랜치가 존재하면 오류가 발생해야 한다
```

### 시나리오 7.3: Worktree 생성

```gherkin
Given main 브랜치가 체크아웃되어 있을 때
When WorktreeService.create("SPEC-001") 메서드를 호출하면
Then ~/worktrees/MoAI-ADK/SPEC-001/ 디렉토리가 생성되어야 한다
And feature/SPEC-001 브랜치가 생성되어야 한다
And worktree 레지스트리에 등록되어야 한다
```

---

## AC-8: 빌드 및 배포

### 시나리오 8.1: 전체 빌드

```gherkin
Given 모든 패키지 코드가 작성되어 있을 때
When `bun run build` 명령을 실행하면
Then 30초 이내에 빌드가 완료되어야 한다
And 모든 패키지의 dist/ 디렉토리가 생성되어야 한다
And TypeScript 오류가 없어야 한다
```

### 시나리오 8.2: CLI 패키지 발행

```gherkin
Given CLI 패키지가 빌드되어 있을 때
When `npm publish` 명령을 실행하면 (dry-run)
Then 패키지가 유효해야 한다
And bin 엔트리가 정확해야 한다
And 의존성이 올바르게 명시되어야 한다
```

### 시나리오 8.3: Dashboard 프로덕션 빌드

```gherkin
Given Dashboard 앱 코드가 작성되어 있을 때
When `bun run build` 명령을 apps/dashboard에서 실행하면
Then .next/ 디렉토리가 생성되어야 한다
And 빌드 크기가 5MB를 초과하지 않아야 한다
And 정적 최적화가 적용되어야 한다
```

---

## AC-9: 테스트 커버리지

### 시나리오 9.1: 단위 테스트 커버리지

```gherkin
Given 모든 패키지에 테스트가 작성되어 있을 때
When `bun run test:coverage` 명령을 실행하면
Then 전체 커버리지가 80% 이상이어야 한다
And packages/core 커버리지가 85% 이상이어야 한다
And packages/config 커버리지가 90% 이상이어야 한다
```

### 시나리오 9.2: E2E 테스트

```gherkin
Given Dashboard 앱이 실행 중일 때
When Playwright E2E 테스트를 실행하면
Then 모든 주요 사용자 플로우가 통과해야 한다
And 스크린샷 스냅샷이 일치해야 한다
And 테스트 실행 시간이 5분을 초과하지 않아야 한다
```

---

## 품질 게이트 체크리스트

### Phase 1 완료 기준
- [ ] AC-1.1: Turborepo 프로젝트 생성 통과
- [ ] AC-1.2: Biome 설정 통과
- [ ] AC-1.3: TypeScript 설정 통과

### Phase 2 완료 기준
- [ ] AC-4.1: YAML 설정 파싱 통과
- [ ] AC-4.2: 잘못된 설정 처리 통과
- [ ] AC-4.3: 설정 저장 통과
- [ ] AC-4.4: 환경 변수 오버라이드 통과

### Phase 3 완료 기준
- [ ] AC-7.1: Git 상태 조회 통과
- [ ] AC-7.2: 브랜치 생성 통과
- [ ] AC-7.3: Worktree 생성 통과

### Phase 4 완료 기준
- [ ] AC-2.1: 새 프로젝트 초기화 통과
- [ ] AC-2.2: 대화형 초기화 통과
- [ ] AC-2.3: 기존 디렉토리 처리 통과

### Phase 5 완료 기준
- [ ] AC-5.1: SessionStart 훅 통과
- [ ] AC-5.2: PreToolUse 훅 통과
- [ ] AC-5.3: 훅 오류 처리 통과

### Phase 6 완료 기준
- [ ] AC-3.1: 초기 페이지 로드 통과
- [ ] AC-3.2: SPEC 목록 페이지 통과
- [ ] AC-3.3: 실시간 업데이트 통과
- [ ] AC-6.1: SPEC 생성 통과
- [ ] AC-6.2: SPEC 편집 통과
- [ ] AC-6.3: SPEC 상태 변경 통과

### Phase 7 완료 기준
- [ ] AC-8.1: 전체 빌드 통과
- [ ] AC-8.2: CLI 패키지 발행 통과
- [ ] AC-8.3: Dashboard 프로덕션 빌드 통과
- [ ] AC-9.1: 단위 테스트 커버리지 80% 이상
- [ ] AC-9.2: E2E 테스트 통과
