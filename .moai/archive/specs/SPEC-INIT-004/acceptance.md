# SPEC-INIT-004 Acceptance Criteria

## @ACCEPTANCE:INIT-004 수락 기준

---

## 1. 개요

**SPEC ID**: INIT-004
**제목**: moai-adk init 명령어 CLAUDE.md 의존성 제거 및 커맨드 생성 기능 수정
**우선순위**: High
**카테고리**: bugfix

---

## 2. 수락 기준 (Acceptance Criteria)

### AC1: CLAUDE.md 없이 신규 설치 성공

#### Given-When-Then

**Given**: 사용자가 신규 프로젝트에서 `moai-adk init` 실행
- 프로젝트 디렉토리가 비어있음
- `~/.claude/CLAUDE.md` 파일이 존재하지 않음
- `.moai/` 디렉토리가 없음

**When**: 사용자가 터미널에서 `moai-adk init` 실행

**Then**:
1. **CLAUDE.md 자동 생성**:
   - `~/.claude/CLAUDE.md` 파일이 템플릿에서 생성됨
   - 콘솔에 `[SUCCESS] CLAUDE.md created from template` 메시지 표시

2. **Alfred 커맨드 파일 생성**:
   - `.claude/commands/alfred/0-project.md` 생성
   - `.claude/commands/alfred/1-spec.md` 생성
   - `.claude/commands/alfred/2-build.md` 생성
   - `.claude/commands/alfred/3-sync.md` 생성
   - 콘솔에 각 파일별 `[SUCCESS] {파일명} copied` 메시지 표시

3. **검증 통과**:
   - 콘솔에 `[SUCCESS] All required files verified` 메시지 표시
   - 초기화 완료 메시지 표시

#### 검증 방법

```bash
# 1. 초기 상태 확인
ls ~/.claude/CLAUDE.md  # 존재하지 않음
ls .moai/               # 디렉토리 없음

# 2. 초기화 실행
moai-adk init

# 3. 결과 확인
ls ~/.claude/CLAUDE.md  # 파일 존재 확인
ls .claude/commands/alfred/
# 출력:
# 0-project.md
# 1-spec.md
# 2-build.md
# 3-sync.md

# 4. 파일 내용 검증
cat .claude/commands/alfred/1-spec.md  # SPEC 작성 커맨드 내용 확인
```

---

### AC2: 필수 커맨드 파일 생성 검증

#### Given-When-Then

**Given**: `moai-adk init` 실행이 완료됨
- 초기화 프로세스가 모든 단계 완료
- 검증 로직이 실행됨

**When**: 검증 로직이 필수 파일 존재 여부를 확인

**Then**:
1. **필수 파일 모두 존재**:
   - `.claude/commands/alfred/0-project.md` ✅
   - `.claude/commands/alfred/1-spec.md` ✅
   - `.claude/commands/alfred/2-build.md` ✅
   - `.claude/commands/alfred/3-sync.md` ✅
   - `.moai/config.json` ✅
   - `.moai/memory/development-guide.md` ✅

2. **파일 내용 검증**:
   - 각 커맨드 파일의 크기가 0보다 큼
   - 템플릿 구조와 일치
   - 필수 섹션 포함 확인

3. **검증 성공 메시지**:
   - 콘솔에 각 파일별 `✓` 표시
   - 최종 `[SUCCESS] All required files verified` 메시지

#### 검증 방법

```typescript
// 단위 테스트 (tests/unit/installer.test.ts)
describe('verifyInstallation', () => {
  it('모든 필수 파일 존재 시 성공', async () => {
    // Given
    const requiredFiles = [
      '.claude/commands/alfred/0-project.md',
      '.claude/commands/alfred/1-spec.md',
      '.claude/commands/alfred/2-build.md',
      '.claude/commands/alfred/3-sync.md',
      '.moai/config.json',
      '.moai/memory/development-guide.md'
    ];

    // 모든 파일 생성 (테스트 fixture)
    for (const file of requiredFiles) {
      await fs.writeFile(path.join(testDir, file), 'content');
    }

    // When
    await verifyInstallation();

    // Then: 에러 없이 완료
  });

  it('필수 파일 누락 시 에러 발생', async () => {
    // Given: 1-spec.md 누락

    // When/Then
    await expect(verifyInstallation()).rejects.toThrow(
      'Installation incomplete. Missing files:\n.claude/commands/alfred/1-spec.md'
    );
  });
});
```

---

### AC3: 템플릿 복사 실패 시 명확한 에러 처리

#### Given-When-Then

**Given**: 템플릿 파일이 패키지에 누락됨 (빌드 오류 시뮬레이션)
- `templates/.claude/commands/alfred/1-spec.md` 파일 부재
- 패키지 무결성 손상 상태

**When**: 사용자가 `moai-adk init` 실행

**Then**:
1. **즉시 중단**:
   - 초기화 프로세스 즉시 중단
   - 부분 설치된 파일 남김 (롤백 없음)

2. **명확한 에러 메시지**:
   ```
   [ERROR] Template command 1-spec.md not found

   This indicates a corrupted package installation.
   Please try:
   1. npm uninstall -g moai-adk
   2. npm install -g moai-adk

   If the issue persists, please report at:
   https://github.com/modu-ai/moai-adk/issues
   ```

3. **에러 타입 분류**:
   - `InstallerError` 타입
   - `type: 'TEMPLATE_MISSING'`
   - 스택 트레이스 포함

#### 검증 방법

```typescript
// 단위 테스트 (tests/unit/installer.test.ts)
describe('copyAlfredCommands - 에러 처리', () => {
  it('템플릿 파일 누락 시 InstallerError 발생', async () => {
    // Given: 1-spec.md 템플릿 누락 시뮬레이션
    const templateDir = path.join(__dirname, '..', 'templates');
    await fs.remove(path.join(templateDir, '.claude/commands/alfred/1-spec.md'));

    // When/Then
    await expect(copyAlfredCommands()).rejects.toThrow(InstallerError);
    await expect(copyAlfredCommands()).rejects.toMatchObject({
      type: 'TEMPLATE_MISSING',
      message: expect.stringContaining('1-spec.md not found')
    });
  });

  it('파일 시스템 권한 오류 시 적절한 에러', async () => {
    // Given: 쓰기 권한 없는 디렉토리
    const readonlyDir = path.join(testDir, 'readonly');
    await fs.mkdir(readonlyDir);
    await fs.chmod(readonlyDir, 0o444);

    // When/Then
    await expect(copyAlfredCommands()).rejects.toThrow(InstallerError);
    await expect(copyAlfredCommands()).rejects.toMatchObject({
      type: 'PERMISSION_DENIED'
    });
  });
});
```

---

## 3. 품질 게이트 (Quality Gates)

### 3.1 테스트 커버리지

**목표**: 85% 이상

**측정 방법**:
```bash
npm run test:coverage
```

**통과 기준**:
- 라인 커버리지: ≥ 85%
- 브랜치 커버리지: ≥ 80%
- 함수 커버리지: ≥ 90%
- 명령문 커버리지: ≥ 85%

### 3.2 TRUST 5원칙

#### T - Test First
- ✅ 모든 함수에 단위 테스트 존재
- ✅ 통합 테스트로 E2E 검증
- ✅ 에지 케이스 커버

#### R - Readable
- ✅ ESLint 규칙 준수
- ✅ Prettier 포맷 적용
- ✅ JSDoc 주석 추가
- ✅ 함수명 명확성 (동사+명사)

#### U - Unified
- ✅ 각 함수 ≤ 50 LOC
- ✅ 파일 ≤ 300 LOC
- ✅ 중첩 깊이 ≤ 3
- ✅ 단일 책임 원칙 준수

#### S - Secured
- ✅ 경로 순회 공격 방지
- ✅ 사용자 입력 검증
- ✅ 파일 시스템 권한 확인
- ✅ 보안 도구 스캔 통과

#### T - Trackable
- ✅ @TAG 체인 완전성
- ✅ Git 커밋 메시지 표준 준수
- ✅ 변경 이력 추적 가능

### 3.3 린트 및 타입 체크

**명령어**:
```bash
npm run lint       # ESLint 검증
npm run type-check # TypeScript 타입 체크
```

**통과 기준**:
- ESLint 에러 0개
- TypeScript 타입 에러 0개
- 경고는 허용 (점진적 개선)

---

## 4. 테스트 시나리오 (Test Scenarios)

### 시나리오 1: 완전 신규 설치

```bash
# 준비
rm -rf .moai .claude ~/.claude/CLAUDE.md

# 실행
moai-adk init

# 검증
ls -R .moai .claude ~/.claude
# 예상 출력:
# ~/.claude/CLAUDE.md
# .moai/config.json
# .moai/memory/development-guide.md
# .claude/commands/alfred/0-project.md
# .claude/commands/alfred/1-spec.md
# .claude/commands/alfred/2-build.md
# .claude/commands/alfred/3-sync.md
```

### 시나리오 2: 기존 CLAUDE.md 보존

```bash
# 준비
echo "# My Custom Config" > ~/.claude/CLAUDE.md
cat ~/.claude/CLAUDE.md  # 사용자 설정 확인

# 실행
moai-adk init

# 검증
cat ~/.claude/CLAUDE.md | grep "My Custom Config"
# 예상: 사용자 설정 보존됨
```

### 시나리오 3: 부분 설치 상태에서 재시도

```bash
# 준비: 부분 설치 시뮬레이션
mkdir -p .moai
echo '{}' > .moai/config.json
# .claude 디렉토리는 없음

# 실행
moai-adk init

# 검증
ls .claude/commands/alfred/
# 예상: 모든 커맨드 파일 생성됨
```

### 시나리오 4: 템플릿 손상 시뮬레이션

```bash
# 준비: 패키지 내 템플릿 파일 삭제 (개발 환경)
rm -rf node_modules/moai-adk/templates/.claude

# 실행
moai-adk init

# 검증
# 예상: [ERROR] Template command ... not found
# 예상: 패키지 재설치 권장 메시지
```

---

## 5. 회귀 테스트 (Regression Tests)

### 기존 기능 보존 확인

- ✅ `.moai/config.json` 생성 정상 작동
- ✅ `.moai/memory/development-guide.md` 복사 정상 작동
- ✅ Personal/Team 모드 설정 정상 작동
- ✅ 템플릿 병합 로직 정상 작동 (SPEC-TEMPLATE-001)
- ✅ 보안 검증 로직 정상 작동 (SPEC-INSTALLER-SEC-001)

### 테스트 방법

```bash
# 기존 테스트 스위트 실행
npm test

# 특정 기능 회귀 테스트
npm test -- --grep "template merge"
npm test -- --grep "security validation"
```

---

## 6. 성능 기준 (Performance Criteria)

### 응답 시간

- **목표**: `moai-adk init` 전체 실행 시간 ≤ 3초
- **측정**:
  ```bash
  time moai-adk init
  ```
- **최적화 포인트**:
  - 파일 I/O 병렬화 (Promise.all 활용)
  - 불필요한 동기 작업 제거

### 리소스 사용

- **메모리**: 최대 50MB
- **디스크 I/O**: 최소화 (한 번만 읽기/쓰기)

---

## 7. 사용자 경험 (User Experience)

### 콘솔 출력 예시

```
$ moai-adk init

[INFO] CLAUDE.md not found. Creating from template...
[SUCCESS] CLAUDE.md created from template

[INFO] Copying Alfred commands...
[SUCCESS] 0-project.md copied
[SUCCESS] 1-spec.md copied
[SUCCESS] 2-build.md copied
[SUCCESS] 3-sync.md copied

[INFO] Verifying installation...
✓ .claude/commands/alfred/0-project.md
✓ .claude/commands/alfred/1-spec.md
✓ .claude/commands/alfred/2-build.md
✓ .claude/commands/alfred/3-sync.md
✓ .moai/config.json
✓ .moai/memory/development-guide.md
[SUCCESS] All required files verified

✨ MoAI-ADK initialization complete!

Next steps:
1. Review .moai/config.json for project settings
2. Start with: /alfred:0-project (initialize project metadata)
3. Create your first SPEC: /alfred:1-plan
```

---

## 8. 완료 조건 체크리스트

### 기능 완료
- [ ] AC1: CLAUDE.md 없이 신규 설치 성공
- [ ] AC2: 필수 커맨드 파일 생성 검증
- [ ] AC3: 템플릿 복사 실패 시 명확한 에러 처리

### 품질 게이트
- [ ] 테스트 커버리지 ≥ 85%
- [ ] 모든 단위 테스트 통과
- [ ] 통합 테스트 통과
- [ ] TRUST 5원칙 준수
- [ ] ESLint/Prettier 통과
- [ ] TypeScript 타입 체크 통과

### 문서화
- [ ] CHANGELOG.md 업데이트
- [ ] README.md 수정 (설치 가이드)
- [ ] API 문서 업데이트 (필요 시)

### Git 워크플로우
- [ ] `feature/SPEC-INIT-004` 브랜치 생성
- [ ] 3회 커밋 (RED/GREEN/REFACTOR)
- [ ] Draft PR 생성
- [ ] CI/CD 통과
- [ ] PR 머지 준비

---

## 9. 관련 이슈

### GitHub Issue
- https://github.com/modu-ai/moai-adk/issues/26

### 관련 SPEC
- SPEC-INSTALLER-SEC-001 (템플릿 보안 검증)
- SPEC-TEMPLATE-001 (Template Processor)

---

_이 수락 기준은 `/alfred:2-run SPEC-INIT-004` 실행 시 검증됩니다._
