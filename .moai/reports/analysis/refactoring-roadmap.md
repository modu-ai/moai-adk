# ANALYSIS:ROADMAP-001: MoAI-ADK 리팩토링 로드맵

> 생성일: 2025-10-01
> 프로젝트: MoAI-ADK v0.0.1 → v1.0.0
> 목표: TRUST 5원칙 완전 준수 + 테스트 커버리지 85% 달성

---

## 🎯 전략적 목표

### 3개월 로드맵 (2025-10 → 2026-01)

**Phase 1: 기반 강화** (1개월)
- 긴급 기술 부채 해결
- 테스트 커버리지 70% 달성
- 대형 클래스 분해

**Phase 2: 품질 향상** (1개월)
- 아키텍처 개선
- 테스트 커버리지 85% 달성
- 성능 최적화

**Phase 3: 혁신 준비** (1개월)
- TAG System 고도화 (✅ 완료)
- 고급 기능 추가
- v1.0.0 릴리스 준비

---

## 📋 Phase 1: 기반 강화 (Week 1-4)

### Week 1: 긴급 테스트 커버리지 확대

**목표**: CLI Core & Command Handlers 테스트 작성

#### Day 1-2: doctor.ts 테스트 (437 LOC)
```typescript
// @CODE:CLI-TD-002 해결
// 파일: tests/cli/commands/doctor.test.ts

describe('DoctorCommand', () => {
  describe('System Diagnostics', () => {
    test('should detect TypeScript project correctly')
    test('should detect Python project correctly')
    test('should validate environment requirements')
  })

  describe('Backup Management', () => {
    test('should list existing backups')
    test('should identify stale backups')
    test('should handle missing backup directory')
  })

  describe('Error Handling', () => {
    test('should handle permission errors gracefully')
    test('should report missing dependencies')
  })
})
```

**예상 시간**: 16시간
**우선순위**: 🔴 긴급
**의존성**: 없음

#### Day 3-4: init.ts & init-prompts.ts 테스트 (655 LOC)
```typescript
// @CODE:CLI-TD-002 해결
// 파일: tests/cli/commands/init.test.ts

describe('InitCommand', () => {
  describe('Interactive Mode', () => {
    test('should prompt for project name')
    test('should validate project name format')
    test('should detect existing projects')
    test('should handle cancellation')
  })

  describe('Wizard Flow', () => {
    test('should complete full initialization flow')
    test('should generate correct config for TypeScript')
    test('should generate correct config for Python')
  })

  describe('Validation', () => {
    test('should reject invalid project names')
    test('should prevent overwriting existing projects')
  })
})
```

**예상 시간**: 24시간
**우선순위**: 🔴 긴급

#### Day 5: config-builder.ts 테스트 (202 LOC)
```typescript
// 파일: tests/cli/config/config-builder.test.ts

describe('ConfigDataBuilder', () => {
  test('should build TypeScript config correctly')
  test('should build Python config correctly')
  test('should detect project language from files')
  test('should apply defaults for missing fields')
})
```

**예상 시간**: 8시간

**Week 1 산출물**:
- ✅ CLI Core 테스트 커버리지: 40% → 75%
- ✅ 48시간 작업 완료
- ✅ @CODE:CLI-TD-002 해결

---

### Week 2: 대형 클래스 분해

#### Day 1-3: GitManager 분해 (690 LOC → 3 x 230 LOC)
```typescript
// @CODE:CORE-001 해결

// Before (690 LOC - 단일 클래스)
class GitManager {
  // 브랜치 관리 (250 LOC)
  // 커밋 관리 (200 LOC)
  // 원격 관리 (150 LOC)
  // 설정 관리 (90 LOC)
}

// After (3개 클래스)
class GitBranchManager {
  async createBranch(name: string): Promise<void>
  async switchBranch(name: string): Promise<void>
  async deleteBranch(name: string): Promise<void>
}

class GitCommitManager {
  async commit(message: string): Promise<void>
  async ammend(): Promise<void>
  async revert(hash: string): Promise<void>
}

class GitRemoteManager {
  async push(branch: string): Promise<void>
  async pull(): Promise<void>
  async fetch(): Promise<void>
}

// Facade (통합 인터페이스 유지)
class GitManager {
  constructor(
    private branch: GitBranchManager,
    private commit: GitCommitManager,
    private remote: GitRemoteManager
  ) {}
}
```

**예상 시간**: 24시간
**우선순위**: 🔴 긴급
**영향 범위**: git-manager 사용하는 모든 에이전트

#### Day 4-5: TemplateManager 리팩토링 (610 LOC → Strategy 패턴)
```typescript
// @CODE:CORE-002 해결

// Strategy 인터페이스
interface TemplateResolutionStrategy {
  resolve(templateName: string): string | null
}

class PackageRelativeStrategy implements TemplateResolutionStrategy {
  resolve(name: string): string | null {
    // package.json 기준 상대 경로
  }
}

class DevelopmentStrategy implements TemplateResolutionStrategy {
  resolve(name: string): string | null {
    // 개발 모드 로컬 경로
  }
}

class NodeModulesStrategy implements TemplateResolutionStrategy {
  resolve(name: string): string | null {
    // node_modules 내부 경로
  }
}

class GlobalStrategy implements TemplateResolutionStrategy {
  resolve(name: string): string | null {
    // 글로벌 설치 경로
  }
}

// Context
class TemplateManager {
  private strategies: TemplateResolutionStrategy[] = [
    new PackageRelativeStrategy(),
    new DevelopmentStrategy(),
    new NodeModulesStrategy(),
    new GlobalStrategy()
  ]

  resolveTemplate(name: string): string {
    for (const strategy of this.strategies) {
      const path = strategy.resolve(name)
      if (path && fs.existsSync(path)) return path
    }
    throw new Error(`Template not found: ${name}`)
  }
}
```

**예상 시간**: 16시간
**우선순위**: 🔴 긴급

**Week 2 산출물**:
- ✅ GitManager 복잡도: 690 LOC → 230 LOC (평균)
- ✅ TemplateManager 명시적 패턴 적용
- ✅ 40시간 작업 완료
- ✅ @CODE:CORE-001, @CODE:CORE-002 해결

---

### Week 3: Sync System 개선

#### Day 1-3: TransactionManager 구현
```typescript
// @CODE:SYNC-001 해결

class TransactionManager {
  private operations: Operation[] = []
  private backupPath: string

  async begin(): Promise<void> {
    // 트랜잭션 시작, 백업 생성
    this.backupPath = await this.createBackup()
  }

  async addOperation(op: Operation): Promise<void> {
    this.operations.push(op)
  }

  async commit(): Promise<void> {
    try {
      for (const op of this.operations) {
        await op.execute()
      }
      await this.removeBackup(this.backupPath)
    } catch (error) {
      await this.rollback()
      throw error
    }
  }

  async rollback(): Promise<void> {
    console.log('트랜잭션 실패, 롤백 중...')
    await this.restoreBackup(this.backupPath)
  }
}

// UpdateOrchestrator에 적용
class UpdateOrchestrator {
  async update(): Promise<void> {
    const tx = new TransactionManager()
    await tx.begin()

    try {
      await tx.addOperation(new BackupOperation())
      await tx.addOperation(new NpmUpdateOperation())
      await tx.addOperation(new TemplateSyncOperation())
      await tx.commit()
    } catch (error) {
      // 자동 롤백
      throw error
    }
  }
}
```

**예상 시간**: 24시간
**우선순위**: 🔴 긴급

#### Day 4-5: 3-Way Merge 구현
```typescript
// 사용자 수정 보존 메커니즘

interface MergeResult {
  merged: string
  conflicts: Conflict[]
}

class ThreeWayMerge {
  merge(
    base: string,      // 원본 템플릿
    theirs: string,    // 새 템플릿
    ours: string       // 사용자 수정본
  ): MergeResult {
    // 1. Diff 계산
    const baseToTheirs = diff(base, theirs)
    const baseToOurs = diff(base, ours)

    // 2. 자동 병합 가능 판단
    if (noConflict(baseToTheirs, baseToOurs)) {
      return { merged: applyBothChanges(), conflicts: [] }
    }

    // 3. 충돌 표시
    return {
      merged: createConflictMarkers(),
      conflicts: identifyConflicts()
    }
  }
}
```

**예상 시간**: 16시간
**우선순위**: 🔴 긴급

**Week 3 산출물**:
- ✅ Sync System 안정성: C+ → B+
- ✅ 트랜잭션 지원 추가
- ✅ 사용자 수정 보존
- ✅ 40시간 작업 완료
- ✅ @CODE:SYNC-001 해결

---

### Week 4: Core Managers 테스트 작성

#### Day 1-2: GitManager 테스트 (분해 후)
```typescript
// tests/core/git/git-branch-manager.test.ts
// tests/core/git/git-commit-manager.test.ts
// tests/core/git/git-remote-manager.test.ts

describe('GitBranchManager', () => {
  test('should create feature branch with correct naming')
  test('should prevent direct push to main')
  test('should handle merge conflicts')
})
```

**예상 시간**: 16시간

#### Day 3-4: TemplateManager 테스트
```typescript
// tests/core/template/template-manager.test.ts

describe('TemplateManager', () => {
  describe('Strategy Pattern', () => {
    test('should try PackageRelative first')
    test('should fallback to Development')
    test('should fallback to NodeModules')
    test('should fallback to Global')
    test('should throw if not found')
  })
})
```

**예상 시간**: 16시간

#### Day 5: UpdateOrchestrator 테스트
```typescript
// tests/core/update/update-orchestrator.test.ts

describe('UpdateOrchestrator', () => {
  test('should update npm package successfully')
  test('should sync templates after update')
  test('should rollback on failure')
})
```

**예상 시간**: 8시간

**Week 4 산출물**:
- ✅ Core Managers 테스트 커버리지: 40% → 80%
- ✅ 전체 테스트 커버리지: 58% → 70%
- ✅ 40시간 작업 완료

**Phase 1 총 작업량**: 160시간 (4주)
**Phase 1 산출물**:
- ✅ 테스트 커버리지: 58% → 70%
- ✅ 긴급 기술 부채 해결: 5개
- ✅ 대형 클래스 분해 완료
- ✅ Sync 안정성 대폭 향상

---

## 📋 Phase 2: 품질 향상 (Week 5-8)

### Week 5: 로깅 일관성 & 작은 개선

#### @CODE:CMD-001 - 로깅 통합 (1일)
```typescript
// Before: console.log 혼용 (72 logger vs 53 console.log)
console.log('✅ 설치 완료')

// After: logger 통일
this.logger.info('Installation completed successfully', {
  timestamp: Date.now(),
  projectPath: path
})
```

#### @CODE:QUALITY-001 - 커버리지 임계값 정렬 (1시간)
```typescript
// vitest.config.ts
coverage: {
  thresholds: {
    branches: 85,    // 80 → 85
    functions: 85,   // 80 → 85
    lines: 85,       // 80 → 85
    statements: 85   // 80 → 85
  }
}
```

#### @CODE:CLI-TD-003 - 메서드 복잡도 감소 (2일)
```typescript
// Before: InitCommand.runInteractive() - 186 LOC
async runInteractive(options): Promise<InitResult> {
  // 186 lines of mixed concerns
}

// After: 단계별 분리
async runInteractive(options): Promise<InitResult> {
  this.displayBanner()
  await this.verifySystem()
  const config = await this.collectConfiguration(options)
  const projectPath = this.determinePath(config, options)
  return await this.executeInstallation(projectPath, config)
}
```

**Week 5 작업량**: 24시간

---

### Week 6: Git Strategies 개선

#### @CODE:GIT-001 - 명시적 Strategy 패턴 (3일)
```typescript
// Strategy 인터페이스
interface GitWorkflowStrategy {
  createBranch(spec: SpecMetadata): Promise<string>
  createPR(spec: SpecMetadata): Promise<string>
  finalizeMerge(spec: SpecMetadata): Promise<void>
}

class PersonalWorkflowStrategy implements GitWorkflowStrategy {
  async createBranch(spec): Promise<string> {
    // 로컬 브랜치 생성
    return this.git.branch.create(`feature/${spec.id}`)
  }

  async createPR(spec): Promise<string> {
    // Personal 모드는 PR 생성 안 함
    return ''
  }
}

class TeamWorkflowStrategy implements GitWorkflowStrategy {
  async createBranch(spec): Promise<string> {
    const branch = await this.git.branch.create(`feature/${spec.id}`)
    await this.git.push(branch)
    return branch
  }

  async createPR(spec): Promise<string> {
    // GitHub CLI로 PR 생성
    return this.github.createPR(spec)
  }
}

// Context
class WorkflowAutomation {
  constructor(
    private strategy: GitWorkflowStrategy
  ) {}

  async executeSpecWorkflow(spec: SpecMetadata): Promise<void> {
    const branch = await this.strategy.createBranch(spec)
    // SPEC 파일 생성
    const prUrl = await this.strategy.createPR(spec)
    // ...
  }
}
```

**Week 6 작업량**: 24시간

---

### Week 7: Installation 개선

#### @CODE:INSTALL-001 - 자동 롤백 메커니즘 (2일)
```typescript
class RollbackManager {
  async createRollbackPoint(): Promise<RollbackPoint> {
    return {
      id: uuid(),
      timestamp: Date.now(),
      snapshot: await this.captureState()
    }
  }

  async rollback(point: RollbackPoint): Promise<void> {
    console.log('Rolling back to:', point.timestamp)
    await this.restoreState(point.snapshot)
  }
}

class InstallationOrchestrator {
  async install(): Promise<void> {
    const rollback = new RollbackManager()
    const point = await rollback.createRollbackPoint()

    try {
      await this.executePhases()
    } catch (error) {
      await rollback.rollback(point)
      throw error
    }
  }
}
```

**Week 7 작업량**: 16시간

---

### Week 8: 테스트 커버리지 최종 확대

#### 목표: 70% → 85%

**누락된 영역**:
- Installation system (Phase-driven 테스트)
- Sync system (Transaction 테스트)
- Documentation generator (템플릿 처리 테스트)

**Week 8 작업량**: 32시간

**Phase 2 총 작업량**: 120시간 (4주)
**Phase 2 산출물**:
- ✅ 테스트 커버리지: 70% → 85%
- ✅ 모든 중요 기술 부채 해결
- ✅ 아키텍처 개선 완료
- ✅ TRUST 5원칙 완전 준수

---

## 📋 Phase 3: 혁신 준비 (Week 9-12)

### Week 9-10: TAG System 단순화 (✅ 완료)

#### 8개 TAG → 4개 TAG 마이그레이션
```typescript
// Before (이전 버전 - 8개 TAG)
@REQ:AUTH-001
@DESIGN:AUTH-001
@TASK:AUTH-001
@TEST:AUTH-001
@FEATURE:AUTH-001
@API:AUTH-001
@UI:AUTH-001
@DATA:AUTH-001

// After (현재 버전 - 4개 TAG)
@SPEC:AUTH-001   (.moai/specs/)
@TEST:AUTH-001   (tests/)
@CODE:AUTH-001   (src/)
  @CODE:AUTH-001:API
  @CODE:AUTH-001:UI
  @CODE:AUTH-001:DATA
@DOC:AUTH-001    (docs/)
```

**마이그레이션 전략**:
1. 코드베이스 전체 스캔
2. 자동 변환 스크립트 실행
3. 검증 및 수동 수정
4. 문서 업데이트

**작업량**: 40시간

---

### Week 11: 고급 기능 추가

#### 1. 증분 스캔 (대규모 프로젝트)
```typescript
class IncrementalScanner {
  async scan(since: Date): Promise<TagScanResult> {
    // Git diff로 변경된 파일만 스캔
    const changedFiles = await this.git.diff(since)
    return this.scanFiles(changedFiles)
  }
}
```

#### 2. 병렬 스캔 (성능 최적화)
```typescript
class ParallelScanner {
  async scan(files: string[]): Promise<TagScanResult[]> {
    const chunks = chunkArray(files, 10)
    return Promise.all(
      chunks.map(chunk => this.scanChunk(chunk))
    )
  }
}
```

**작업량**: 24시간

---

### Week 12: v1.0.0 릴리스 준비

#### 릴리스 체크리스트
- [ ] 모든 테스트 통과 (커버리지 ≥85%)
- [ ] 모든 긴급/중요 기술 부채 해결
- [ ] 문서 업데이트 (CHANGELOG, README, API docs)
- [ ] 성능 벤치마크 실행
- [ ] 보안 감사 (npm audit, dependency check)
- [ ] 크로스 플랫폼 테스트 (Windows, macOS, Linux)
- [ ] 릴리스 노트 작성
- [ ] GitHub Release 태그 생성

**작업량**: 16시간

**Phase 3 총 작업량**: 80시간 (4주)

---

## 📊 전체 로드맵 요약

| Phase | 기간 | 작업량 | 주요 목표 |
|-------|------|--------|----------|
| **Phase 1** | Week 1-4 | 160h | 기반 강화, 테스트 70% |
| **Phase 2** | Week 5-8 | 120h | 품질 향상, 테스트 85% |
| **Phase 3** | Week 9-12 | 80h | 혁신 준비, v1.0.0 |
| **총계** | 12주 | **360h** | TRUST 완전 준수 |

**1인 개발자 기준**: 주 30시간 × 12주 = 360시간
**팀 개발 (2인)**: 6주 완료 가능

---

## 🎯 성공 지표

### Phase 1 완료 기준
- ✅ 테스트 커버리지 ≥ 70%
- ✅ 긴급 기술 부채 0개
- ✅ 파일 평균 크기 ≤ 300 LOC

### Phase 2 완료 기준
- ✅ 테스트 커버리지 ≥ 85%
- ✅ 중요 기술 부채 ≤ 5개
- ✅ TRUST 5원칙 완전 준수

### Phase 3 완료 기준
- ✅ TAG System 단순화 적용
- ✅ v1.0.0 릴리스 완료
- ✅ 크로스 플랫폼 검증 완료

---

## 📅 마일스톤

- **2025-11-01**: Phase 1 완료 (v0.1.0 목표)
- **2025-12-01**: Phase 2 완료 (v0.5.0 목표)
- **2026-01-01**: Phase 3 완료 (v1.0.0 릴리스)

---

## 🚀 실행 방법

### 1. SPEC 생성
```bash
# 각 기술 부채에 대해 SPEC 작성
/alfred:1-spec "@CODE:CLI-TD-002 해결 - CLI Core 테스트 작성"
```

### 2. TDD 구현
```bash
# Red-Green-Refactor 사이클
/alfred:2-build "SPEC-CLI-TD-002"
```

### 3. 문서 동기화
```bash
# 코드와 문서 동기화, 보고서 업데이트
/alfred:3-sync
/alfred:8-project analyze --update
```

---

_이 로드맵은 10개 영역 분석 결과를 기반으로 자동 생성되었습니다._
_업데이트: 각 Phase 완료 시점_
