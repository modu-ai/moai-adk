---
id: INSTALLER-ROLLBACK-001
version: 0.2.0
status: completed
created: 2025-10-06
updated: 2025-10-18
priority: high
---


## HISTORY

### v0.2.0 (2025-10-18)
- **CHANGED**: deprecated → completed (TypeScript 프로젝트 아카이브)
- **REASON**: TypeScript 프로젝트에서 구현 완료된 기능, Python 전환으로 deprecated 처리했으나 실제로는 완료된 것으로 간주

### v0.1.0 (2025-10-16)
- **DEPRECATED**: TypeScript 프로젝트용 SPEC, Python 프로젝트에는 적용 불가
- **REASON**: MoAI-ADK가 Python 프로젝트로 전환됨에 따라 TypeScript installer 관련 SPEC 불필요
- **ALTERNATIVE**: Python 프로젝트는 다른 설치 메커니즘 사용

### v0.0.1 (2025-10-06)
- **INITIAL**: 설치 실패 시 자동 롤백 메커니즘 명세 작성 (TypeScript용)
- **SCOPE**: 트랜잭션형 설치 프로세스 구현, 실패 시 자동 정리

## 1. 개요

### 1.1 목적
MoAI-ADK 설치 과정에서 실패가 발생했을 때, 생성된 파일과 디렉토리를 자동으로 정리하고 시스템을 이전 상태로 복원하는 롤백 메커니즘을 구현한다.

### 1.2 범위
- **포함 사항**:
  - 설치 과정 중 생성된 파일/디렉토리 추적
  - 실패 시 자동 롤백 실행
  - 상세한 실패 리포트 생성
  - Best-effort 정리 (권한 문제 등 경고 처리)

- **제외 사항**:
  - 외부 패키지 매니저 작업 롤백 (npm/pnpm은 자체 롤백 제공)
  - Git 저장소 초기화 롤백 (사용자가 명시적으로 요청한 경우 제외)

### 1.3 문제 정의
현재 설치 실패 시 부분적으로 생성된 파일들이 남아있어:
- 재설치 시 충돌 발생
- 디스크 공간 낭비
- 사용자 혼란 초래
- 수동 정리 필요

## 2. EARS 요구사항

### 2.1 Ubiquitous Requirements

**REQ-ROLLBACK-001**: 시스템은 설치 과정에서 생성되는 모든 파일과 디렉토리를 추적해야 한다.

**REQ-ROLLBACK-002**: 시스템은 설치 실패 시 추적된 모든 파일과 디렉토리를 자동으로 삭제해야 한다.

**REQ-ROLLBACK-003**: 시스템은 롤백 실패 시 상세한 에러 리포트를 생성해야 한다.

**REQ-ROLLBACK-004**: 시스템은 `InstallationTransaction` 클래스를 통해 트랜잭션형 설치를 제공해야 한다.

### 2.2 Event-driven Requirements

**REQ-ROLLBACK-010**: WHEN 파일 또는 디렉토리가 생성되면, 시스템은 해당 경로를 트랜잭션 컨텍스트에 기록해야 한다.

**REQ-ROLLBACK-011**: WHEN 설치 Phase에서 에러가 발생하면, 시스템은 즉시 롤백 프로세스를 시작해야 한다.

**REQ-ROLLBACK-012**: WHEN 롤백 중 파일 삭제 실패가 발생하면, 시스템은 경고를 출력하고 계속 진행해야 한다.

**REQ-ROLLBACK-013**: WHEN 롤백이 완료되면, 시스템은 삭제 성공/실패 통계를 출력해야 한다.

### 2.3 State-driven Requirements

**REQ-ROLLBACK-020**: WHILE 트랜잭션이 활성화된 상태일 때, 시스템은 모든 파일 시스템 작업을 추적해야 한다.

**REQ-ROLLBACK-021**: WHILE 롤백이 진행 중일 때, 시스템은 역순(생성 반대 순서)으로 파일을 삭제해야 한다.

### 2.4 Optional Requirements

**REQ-ROLLBACK-030**: WHERE 사용자가 `--keep-partial` 플래그를 사용하면, 시스템은 롤백을 건너뛸 수 있다.

**REQ-ROLLBACK-031**: WHERE 롤백 실패 파일이 있으면, 시스템은 수동 정리 스크립트를 생성할 수 있다.

### 2.5 Constraints

**REQ-ROLLBACK-040**: IF 파일 삭제 권한이 없으면, 시스템은 경고를 출력하고 해당 파일을 건너뛰어야 한다.

**REQ-ROLLBACK-041**: IF 디렉토리가 비어있지 않으면, 시스템은 재귀적으로 삭제해야 한다.

**REQ-ROLLBACK-042**: IF 트랜잭션이 커밋되면, 시스템은 추적 데이터를 정리하고 롤백을 비활성화해야 한다.

## 3. 기술 상세

### 3.1 InstallationTransaction 클래스

```typescript

export class InstallationTransaction {
  private createdPaths: Set<string> = new Set();
  private isCommitted: boolean = false;

  /**
   * 파일/디렉토리 생성 추적
   */
  track(path: string): void {
    if (!this.isCommitted) {
      this.createdPaths.add(path);
    }
  }

  /**
   * 트랜잭션 커밋 (성공 시 롤백 비활성화)
   */
  commit(): void {
    this.isCommitted = true;
    this.createdPaths.clear();
  }

  /**
   * 롤백 실행 (역순 삭제)
   */
  async rollback(): Promise<RollbackResult> {
    const deleted: string[] = [];
    const failed: Array<{ path: string; error: string }> = [];

    // 역순으로 삭제 (나중에 생성된 것부터)
    const paths = Array.from(this.createdPaths).reverse();

    for (const path of paths) {
      try {
        await fs.remove(path); // fs-extra의 recursive remove
        deleted.push(path);
      } catch (error) {
        failed.push({
          path,
          error: error instanceof Error ? error.message : String(error)
        });
      }
    }

    return { deleted, failed };
  }
}

export interface RollbackResult {
  deleted: string[];
  failed: Array<{ path: string; error: string }>;
}
```

### 3.2 InstallerCore 통합

```typescript
export class InstallerCore {
  async install(options: InstallOptions): Promise<void> {
    const transaction = new InstallationTransaction();

    try {
      // Phase 실행 시 transaction 전달
      await this.phaseExecutor.executePhase(Phase.PreInstall, {
        ...options,
        transaction
      });

      // ... 모든 Phase 실행

      // 성공 시 커밋
      transaction.commit();
    } catch (error) {
      // 실패 시 롤백
      const result = await transaction.rollback();

      // 롤백 결과 리포트
      this.reportRollback(result);

      throw error; // 원본 에러 전파
    }
  }

  private reportRollback(result: RollbackResult): void {
    console.error(`\n🔄 Rollback completed:`);
    console.error(`  ✅ Deleted: ${result.deleted.length} files/directories`);

    if (result.failed.length > 0) {
      console.error(`  ⚠️  Failed to delete: ${result.failed.length} items`);
      result.failed.forEach(({ path, error }) => {
        console.error(`     - ${path}: ${error}`);
      });
    }
  }
}
```

### 3.3 파일 시스템 작업 래퍼

```typescript
export class TrackedFileSystem {
  constructor(private transaction: InstallationTransaction) {}

  async ensureDir(path: string): Promise<void> {
    await fs.ensureDir(path);
    this.transaction.track(path);
  }

  async writeFile(path: string, content: string): Promise<void> {
    await fs.writeFile(path, content);
    this.transaction.track(path);
  }

  async copy(src: string, dest: string): Promise<void> {
    await fs.copy(src, dest);
    this.transaction.track(dest);
  }
}
```

## 4. 성공 기준

### 4.1 정량적 지표
- [ ] 설치 실패 시 100% 자동 롤백 시도
- [ ] 권한 문제 제외 95% 이상 정리 성공
- [ ] 롤백 시간 < 5초 (평균)

### 4.2 정성적 지표
- [ ] 사용자가 수동 정리 불필요
- [ ] 상세한 롤백 리포트 제공
- [ ] 재설치 시 충돌 없음

## 5. 참조

### 5.1 관련 SPEC
- `SPEC-REFACTOR-001`: Installer 패키지 리팩토링
- `SPEC-INSTALLER-TEST-001`: 테스트 커버리지 (병행 작업)
- `SPEC-INSTALLER-QUALITY-001`: 코드 품질 개선 (병행 작업)

### 5.2 관련 문서
- `.moai/memory/development-guide.md`: 에러 처리 가이드
- `moai-adk-ts/src/core/installer/installer-core.ts`: 핵심 설치 로직

### 5.3 관련 TAG
