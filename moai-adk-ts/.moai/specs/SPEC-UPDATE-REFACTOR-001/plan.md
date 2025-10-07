# @SPEC:UPDATE-REFACTOR-001 구현 계획서

## 개요

본 문서는 `/alfred:9-update` 명령어의 SPEC 문서화 작업에 대한 구현 계획을 제공합니다. 기존 구현(9-update.md v0.2.6, restore.ts)을 기반으로 하며, 추가 개선 사항 및 확장 가능성을 검토합니다.

## 현재 상태 분석 (As-Is)

### 기존 산출물 검토

**9-update.md (v0.2.6)**:
- ✅ Phase 1-5 실행 절차 완전히 문서화됨
- ✅ Alfred 직접 실행 방식 명시
- ✅ 템플릿 복사/병합 전략 상세화 (Step 1-11)
- ✅ 지능형 병합 시스템 (CLAUDE.md, config.json)
- ✅ 오류 복구 시나리오 4가지 정의
- ✅ trust-checker 선택적 연동 (--check-quality)

**restore.ts 구현**:
- ✅ BackupValidationResult 인터페이스 정의
- ✅ RestoreOptions 인터페이스 정의 (dryRun, force)
- ✅ RestoreResult 인터페이스 정의
- ✅ validateBackupPath() 메서드 구현
- ✅ performRestore() 메서드 구현
- ✅ run() 메서드 구현 (사용자 인터페이스)

### 작동 상태

| 컴포넌트 | 상태 | 검증 방법 |
|----------|------|-----------|
| `/alfred:9-update` 명령어 | ✅ 작동 중 | 문서 기반 실행 |
| Phase 1-5 실행 흐름 | ✅ 완전 | 9-update.md 참조 |
| 백업 생성 | ✅ 구현 | .moai-backup/ 생성 확인 |
| 복원 기능 | ✅ 구현 | `moai restore` 테스트 |
| 템플릿 병합 | ✅ 구현 | Alfred 직접 실행 |
| 품질 검증 | ✅ 선택적 | `--check-quality` 옵션 |

### 현재 제약사항

**기술적 제약**:
- Alfred가 Claude Code 도구만으로 모든 작업을 수행 (TypeScript 코드 최소화)
- JSON 파싱/병합이 복잡한 경우 오류 가능성 존재
- Windows 환경에서 `chmod` 실패 (경고만 출력)

**문서화 제약**:
- 9-update.md가 매우 상세하여 초보자에게는 복잡할 수 있음
- 오류 복구 시나리오가 문서로만 존재 (자동화 부족)

## 구현 계획 (To-Be)

### 마일스톤 1: SPEC 문서 작성 완료 (현재 단계)

**목표**: UPDATE-REFACTOR-001 SPEC 문서 완성

**산출물**:
- ✅ `spec.md` - EARS 명세, Phase 1-5 상세 정의
- ⏳ `plan.md` - 구현 계획서 (본 문서)
- ⏳ `acceptance.md` - 수락 기준, 테스트 시나리오

**주요 작업**:
1. 기존 9-update.md 내용을 EARS 구조로 변환
2. restore.ts 구현 내용을 Specifications에 반영
3. 템플릿 병합 전략 상세화 (Step 1-11)
4. 오류 복구 시나리오 정의 (4가지)
5. @TAG 체인 연결 (9-update.md, restore.ts)

**완료 기준**:
- [ ] YAML Front Matter 완전성 (7개 필수 필드)
- [ ] HISTORY 섹션 작성 (v0.1.0 INITIAL)
- [ ] EARS 구조 완전성 (4개 섹션)
- [ ] Phase 1-5 상세 명세
- [ ] @TAG 체인 무결성

### 마일스톤 2: 기존 구현 검증

**목표**: 9-update.md와 restore.ts의 정합성 확인

**검증 항목**:
1. **Phase 1-5 실행 흐름**:
   - [ ] Phase 1 (버전 확인) 실제 작동 확인
   - [ ] Phase 2 (백업 생성) 백업 디렉토리 생성 확인
   - [ ] Phase 3 (npm 업데이트) 패키지 설치 확인
   - [ ] Phase 4 (템플릿 복사/병합) 파일 복사 결과 확인
   - [ ] Phase 5 (검증) 검증 로직 작동 확인

2. **복원 기능**:
   - [ ] `moai restore <path>` 백업 경로 지정 작동 확인
   - [ ] `--dry-run` 옵션 미리보기 작동 확인
   - [ ] `--force` 옵션 강제 복원 작동 확인
   - [ ] 백업 검증 로직 (missing items) 작동 확인

3. **템플릿 병합**:
   - [ ] CLAUDE.md 지능형 병합 작동 확인
   - [ ] config.json 딥 병합 작동 확인
   - [ ] 프로젝트 문서 보존 정책 작동 확인

### 마일스톤 3: 개선 가능 영역 식별

**목표**: 사용자 경험 및 오류 처리 개선 방안 도출

**개선 후보**:

#### 3.1 사용자 경험 개선

**진행률 표시**:
```text
📄 Phase 4: 템플릿 복사/병합 중... [1/11]
   ✅ .claude/commands/alfred/ (10개 파일) [2/11]
   ✅ .claude/agents/alfred/ (9개 파일) [3/11]
   ...
   ✅ 템플릿 파일 처리 완료 [11/11]
```

**예상 소요 시간 안내** (금지):
- ❌ 시간 예측 표현 사용 금지
- ✅ 대신 Phase별 순서 표시 (1/5, 2/5 등)

**상호작용 개선**:
```text
⚠️ 검증 실패: 2개 파일 누락
다음 조치를 선택하세요:
  1. Phase 4 재실행
  2. 백업 복원
  3. 무시하고 진행 (위험)
선택 (1-3): _
```

#### 3.2 오류 처리 자동화

**자동 재시도 로직**:
- 파일 복사 실패 시 자동 재시도 (최대 2회)
- 디렉토리 없음 시 `mkdir -p` 자동 실행
- 권한 오류 시 `chmod` 자동 재실행

**오류 로그 수집**:
```text
.moai-backup/2025-10-06-15-30-00/
├── .claude/
├── .moai/
├── CLAUDE.md
└── update-error.log  # ← 오류 로그 추가
```

**오류 리포트 생성**:
```markdown
# 업데이트 오류 리포트

## 실패 파일 목록
- .claude/commands/alfred/2-build.md (Write 실패: 디스크 공간 부족)

## 권장 조치
1. 디스크 공간 확보 후 Phase 4 재실행
2. 또는 백업 복원: moai restore .moai-backup/2025-10-06-15-30-00
```

#### 3.3 백업 관리 개선

**자동 정리 정책**:
```bash
# 30일 이상 된 백업 자동 정리
find .moai-backup/ -type d -mtime +30 -exec rm -rf {} \;
```

**백업 목록 조회**:
```bash
/alfred:9-update --list-backups

# 출력 예시:
📦 백업 목록:
1. 2025-10-06-15-30-00 (2시간 전)
2. 2025-10-05-10-20-00 (1일 전)
3. 2025-10-01-09-00-00 (5일 전)

복원: moai restore .moai-backup/2025-10-06-15-30-00
```

#### 3.4 버전별 마이그레이션 가이드

**버전 업그레이드 감지**:
```text
🔍 버전 확인 중...
📦 현재 버전: v0.0.1
⚡ 최신 버전: v0.2.0 (메이저 업데이트)

⚠️  Breaking Changes 감지:
- config.json 구조 변경
- git_strategy 필드 추가

📖 마이그레이션 가이드: https://github.com/modu-ai/moai-adk/releases/v0.2.0
```

**자동 마이그레이션**:
```text
🔄 자동 마이그레이션 실행 중...
   ✅ config.json: git_strategy 필드 추가
   ✅ CLAUDE.md: 에이전트 목록 업데이트
   ✅ 마이그레이션 완료
```

### 마일스톤 4: 확장 기능 검토

**목표**: 장기적 확장 가능성 평가

**확장 후보**:

#### 4.1 업데이트 이력 추적

**changelog 자동 생성**:
```markdown
# 업데이트 이력

## v0.0.2 (2025-10-06)
- **업데이트 시각**: 2025-10-06 15:30:00
- **이전 버전**: v0.0.1
- **새 버전**: v0.0.2
- **변경 파일**: 23개
- **백업 경로**: .moai-backup/2025-10-06-15-30-00

### 변경 내용
- 명령어 파일 10개 업데이트
- 에이전트 파일 9개 업데이트
- CLAUDE.md 병합 (프로젝트 정보 유지)
- config.json 병합 (사용자 설정 유지)
```

#### 4.2 원격 템플릿 직접 조회

**GitHub API 활용**:
```bash
# 최신 릴리스 확인
curl -s https://api.github.com/repos/modu-ai/moai-adk/releases/latest
```

**원격 템플릿 다운로드**:
```bash
# npm 없이도 업데이트 가능
wget https://github.com/modu-ai/moai-adk/archive/refs/tags/v0.0.2.tar.gz
tar -xzf v0.0.2.tar.gz
```

#### 4.3 롤백 체인

**다단계 롤백**:
```bash
/alfred:9-update --rollback-to v0.0.1

# 여러 버전 건너뛰고 롤백
v0.0.3 → v0.0.2 → v0.0.1
```

**롤백 검증**:
```text
🔄 롤백 시뮬레이션 중...
   현재: v0.0.3
   목표: v0.0.1
   건너뛰는 버전: v0.0.2

⚠️  롤백 시 변경사항:
- config.json: git_strategy 필드 제거
- CLAUDE.md: 에이전트 목록 이전 버전 복원
- 명령어 파일 3개 다운그레이드

계속하시겠습니까? (y/n): _
```

## 기술적 접근 방법

### Alfred 직접 실행 방식

**핵심 원칙**:
- TypeScript 코드 최소화
- Claude Code 도구(Read, Write, Bash, Grep, Glob)만 사용
- 문서 기반 지침으로 모든 로직 표현

**장점**:
- ✅ 코드 복잡도 감소
- ✅ 유지보수 용이 (문서 수정만으로 동작 변경 가능)
- ✅ 에러 발생 시 debug-helper 자동 호출 가능

**단점**:
- ⚠️ 복잡한 JSON 파싱/병합 시 오류 가능성
- ⚠️ 플랫폼별 차이 (Windows vs Unix)

**최적화 전략**:
```text
Phase 4 실행 시:
1. Glob으로 파일 목록 일괄 조회
2. Read/Write를 순차 실행 (병렬 불가능)
3. 오류 발생 시 재시도 로직 내장
4. 실패 파일 목록 수집 후 마지막에 보고
```

### 템플릿 병합 전략

**3단계 보호 정책**:

| 정책 | 대상 | 판단 기준 | 처리 방식 |
|------|------|-----------|-----------|
| 완전 보존 🔒 | .moai/specs/, .moai/reports/ | 항상 | 절대 건드리지 않음 |
| 사용자 수정 보존 🔒 | product/structure/tech.md | `{{PROJECT_NAME}}` 없음 | 보존, 템플릿 참조 안내 |
| 전체 교체 ✅ | 시스템 파일 | 항상 | 최신 템플릿으로 덮어쓰기 |
| 지능형 병합 🔄 | CLAUDE.md | `{{PROJECT_NAME}}` 없음 | 프로젝트 정보 추출 + 템플릿 주입 |
| 딥 병합 🔄 | config.json | `{{PROJECT_NAME}}` 없음 | 필드별 병합 정책 적용 |

**병합 알고리즘** (config.json):
```javascript
// 의사 코드
function deepMerge(template, user) {
  return {
    project: user.project,  // 100% 보존
    constitution: { ...template.constitution, ...user.constitution },  // 덮어쓰기
    git_strategy: user.git_strategy,  // 100% 보존
    tags: {
      ...template.tags,
      categories: [...new Set([...template.tags.categories, ...user.tags.categories])]  // 병합
    },
    pipeline: { ...template.pipeline, current_stage: user.pipeline.current_stage },  // 부분 보존
    _meta: template._meta  // 템플릿 사용
  };
}
```

### 오류 복구 아키텍처

**3단계 복구 전략**:

1. **자동 복구** (Alfred가 자동 시도):
   - 디렉토리 없음 → `mkdir -p`
   - 파일 복사 실패 → 2회 재시도
   - 권한 오류 → `chmod +x`

2. **반자동 복구** (사용자 선택):
   - 파일 누락 → "Phase 4 재실행" 제안
   - 버전 불일치 → "Phase 3 재실행" 제안
   - 검증 실패 → "재시도 / 롤백 / 무시" 선택

3. **수동 복구** (사용자 직접):
   - 치명적 오류 → 백업 복원 명령어 안내
   - JSON 파싱 실패 → 백업 파일 참조 안내

## 리스크 및 대응 방안

### 1. JSON 파싱 실패 리스크

**위험도**: Medium

**시나리오**:
- config.json이 손상되어 파싱 불가능
- 사용자가 수동으로 JSON을 잘못 수정

**대응 방안**:
```text
IF JSON 파싱 실패:
  1. [Bash] cp .moai/config.json .moai/config.json.backup
  2. [Read] "{TEMPLATE_ROOT}/.moai/config.json"
  3. [Write] ".moai/config.json" (템플릿으로 교체)
  4. 사용자에게 안내:
     "⚠️ config.json 파싱 실패, 백업 생성: .moai/config.json.backup"
     "📝 수동 복구 필요: .moai/config.json.backup 참조"
```

### 2. 템플릿 구조 변경 리스크

**위험도**: Medium

**시나리오**:
- 새 버전에서 파일 구조가 크게 변경됨
- 예: `.claude/commands/` → `.claude/commands/v2/`

**대응 방안**:
```text
1. 버전별 마이그레이션 가이드 제공
2. 메이저 버전 업그레이드 시 경고 표시
3. 백업 보존으로 안전 장치 확보
```

### 3. 크로스 플랫폼 호환성 리스크

**위험도**: Low

**시나리오**:
- Windows에서 `chmod` 실패
- 경로 구분자 차이 (`/` vs `\`)

**대응 방안**:
```text
1. Windows에서 chmod 실패는 경고로 처리 (계속 진행)
2. Node.js path 모듈 사용 (path.join)
3. 플랫폼별 테스트 수행
```

### 4. npm 캐시 손상 리스크

**위험도**: Low

**시나리오**:
- npm 캐시가 손상되어 템플릿 파일이 올바르지 않음

**대응 방안**:
```bash
# Phase 3 재실행 시 캐시 정리
npm cache clean --force
npm install moai-adk@latest
```

## 성능 고려사항

### 파일 복사 최적화

**현재 방식** (순차 처리):
```text
FOR EACH file:
  Read → Write
```

**최적화 불가능 이유**:
- Claude Code 도구는 순차 실행만 지원
- 병렬 처리는 Alfred 오케스트레이션 수준에서만 가능

**예상 실행 시간**:
- Phase 1: 2-3초 (npm 명령어 2회)
- Phase 2: 1-2초 (백업 생성)
- Phase 3: 10-30초 (npm 설치)
- Phase 4: 5-10초 (템플릿 복사, 약 30개 파일)
- Phase 5: 2-3초 (검증)
- **총 예상 시간**: 20-48초

**주의**: 시간 예측 표현은 문서에서 사용 금지, 내부 계획 참고용만

### 백업 용량 최적화

**현재 백업 크기**:
```bash
.moai-backup/2025-10-06-15-30-00/
├── .claude/ (~500KB)
├── .moai/ (~1MB, SPEC 파일 포함)
└── CLAUDE.md (~50KB)
# 총: ~1.5MB
```

**최적화 방안**:
- SPEC 파일 제외 가능 (항상 보존되므로)
- 압축 백업 (tar.gz) 고려
- 백업 개수 제한 (최근 10개만 보존)

## 의존성 관리

### 외부 의존성

| 의존성 | 버전 | 용도 | 필수 여부 |
|--------|------|------|-----------|
| npm | ≥6.0 | 패키지 관리 | ✅ 필수 |
| Node.js | ≥14.0 | 런타임 | ✅ 필수 |
| Git | ≥2.0 | 버전 관리 | ⚠️ 권장 |
| fs-extra | latest | 파일 시스템 | ✅ 필수 (restore.ts) |
| chalk | latest | 컬러 출력 | ⚠️ 선택 |

### 내부 의존성

| 컴포넌트 | 의존 대상 | 타입 | 비고 |
|----------|-----------|------|------|
| 9-update.md | Alfred | 직접 실행 | 중앙 오케스트레이터 |
| restore.ts | fs-extra | npm 패키지 | 파일 복사/검증 |
| Phase 4 | Glob, Read, Write | Claude Code 도구 | 템플릿 복사 |
| Phase 5.5 | trust-checker | 에이전트 | 선택적 호출 |

## 테스트 전략

### 단위 테스트

**restore.ts 테스트**:
```typescript
// @TEST:CLI-005
describe('RestoreCommand', () => {
  it('should validate backup path', async () => {
    const cmd = new RestoreCommand();
    const result = await cmd.validateBackupPath('.moai-backup/test');
    expect(result.isValid).toBe(true);
  });

  it('should detect missing items', async () => {
    const cmd = new RestoreCommand();
    const result = await cmd.validateBackupPath('.moai-backup/incomplete');
    expect(result.missingItems).toContain('.moai');
  });

  it('should perform dry-run restore', async () => {
    const cmd = new RestoreCommand();
    const result = await cmd.performRestore('.moai-backup/test', { dryRun: true });
    expect(result.isDryRun).toBe(true);
    expect(result.success).toBe(true);
  });
});
```

### 통합 테스트

**전체 업데이트 시나리오**:
```bash
# Given: 프로젝트 초기 상태 (v0.0.1)
moai init test-project

# When: 업데이트 실행
/alfred:9-update

# Then: 검증
- 백업 생성 확인: .moai-backup/ 디렉토리 존재
- 패키지 버전 확인: npm list moai-adk → v0.0.2
- 템플릿 파일 확인: .claude/commands/alfred/ 파일 개수
- 사용자 설정 보존: config.json project 필드 유지
```

**롤백 시나리오**:
```bash
# Given: 업데이트 완료 후 문제 발생
/alfred:9-update
# 오류 발생

# When: 롤백 실행
moai restore .moai-backup/2025-10-06-15-30-00

# Then: 검증
- 파일 복원 확인: .claude/, .moai/, CLAUDE.md
- 패키지 버전 확인: npm list moai-adk → v0.0.1
- 프로젝트 설정 확인: config.json 이전 상태 복원
```

### E2E 테스트

**크로스 플랫폼 테스트**:
```yaml
# .github/workflows/test.yml
jobs:
  test:
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
        node: [14, 16, 18, 20]
    runs-on: ${{ matrix.os }}
    steps:
      - run: npm test
      - run: npm run test:update
```

## 문서화 전략

### 사용자 문서

**9-update.md**:
- ✅ 이미 v0.2.6로 완전히 문서화됨
- ✅ Phase별 상세 절차 제공
- ✅ 오류 복구 시나리오 포함

**개선 방향**:
- 초보자를 위한 요약 가이드 추가
- 자주 묻는 질문 (FAQ) 섹션 추가
- 스크린샷 또는 애니메이션 GIF 추가 (선택)

### 개발자 문서

**SPEC-UPDATE-REFACTOR-001**:
- ✅ 본 문서 (plan.md) 작성 중
- ⏳ acceptance.md 작성 예정
- ✅ spec.md 완료

**코드 주석**:
```typescript
// @CODE:CLI-005 | SPEC: SPEC-UPDATE-REFACTOR-001.md

/**
 * Restore command for backup restoration
 * @tags @CODE:CLI-RESTORE-001
 * @see .moai/specs/SPEC-UPDATE-REFACTOR-001/spec.md
 */
export class RestoreCommand {
  // ...
}
```

## 다음 단계

### 즉시 수행 가능

1. **acceptance.md 작성**:
   - Given-When-Then 시나리오 작성
   - Definition of Done 정의
   - 품질 게이트 기준 설정

2. **기존 구현 검증**:
   - `/alfred:9-update` 실행 테스트
   - `moai restore` 실행 테스트
   - 템플릿 병합 결과 확인

3. **TAG 체인 검증**:
   ```bash
   rg '@SPEC:UPDATE-REFACTOR-001' -n
   rg '@CODE:CLI-005' -n
   rg '@DOC:UPDATE-001' -n
   ```

### 점진적 개선

1. **사용자 경험 개선** (마일스톤 3):
   - 진행률 표시 추가
   - 상호작용 개선 (사용자 선택)
   - 오류 메시지 개선

2. **오류 처리 자동화** (마일스톤 3):
   - 자동 재시도 로직 강화
   - 오류 로그 수집 및 리포트 생성
   - 백업 자동 정리 정책

3. **확장 기능 추가** (마일스톤 4):
   - 업데이트 이력 추적
   - 버전별 마이그레이션 가이드
   - 원격 템플릿 직접 조회

### 장기 로드맵

1. **Phase 4 최적화**:
   - 템플릿 복사 성능 개선
   - 병합 알고리즘 최적화

2. **Phase 5.5 확장**:
   - trust-checker Level 2/3 연동
   - 보안 취약점 자동 패치

3. **플랫폼 확장**:
   - Docker 환경 지원
   - CI/CD 파이프라인 통합

## 참고 자료

- **기존 문서**:
  - templates/.claude/commands/alfred/9-update.md (v0.2.6)
  - src/cli/commands/restore.ts
  - .moai/memory/development-guide.md

- **관련 SPEC**:
  - @SPEC:BACKUP-VALIDATION-001
  - @SPEC:RESTORE-OPTIONS-001
  - @SPEC:RESTORE-RESULT-001

- **외부 문서**:
  - npm CLI 문서: https://docs.npmjs.com/cli/
  - fs-extra 문서: https://github.com/jprichardson/node-fs-extra
  - Claude Code 도구: https://docs.anthropic.com/claude/docs/claude-code

---

**작성자**: @Goos
**작성일**: 2025-10-06
**버전**: 0.1.0
**상태**: Draft
