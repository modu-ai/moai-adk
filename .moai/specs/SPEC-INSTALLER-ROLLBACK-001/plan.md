
## 1. Phase 1: InstallationTransaction 클래스 구현

### 1.1 파일 생성
- **파일 경로**: `moai-adk-ts/src/core/installer/installation-transaction.ts`
- **테스트 파일**: `moai-adk-ts/tests/core/installer/installation-transaction.test.ts`

### 1.2 구현 내용
```typescript

export class InstallationTransaction {
  private createdPaths: Set<string>;
  private isCommitted: boolean;

  track(path: string): void;
  commit(): void;
  rollback(): Promise<RollbackResult>;
}
```

### 1.3 TDD 테스트 케이스
- [ ] track() 호출 시 경로 추가 검증
- [ ] commit() 호출 후 track() 무시 검증
- [ ] rollback() 역순 삭제 검증
- [ ] rollback() 실패 시 경고 출력 검증
- [ ] 빈 트랜잭션 롤백 검증

## 2. Phase 2: TrackedFileSystem 래퍼 구현

### 2.1 파일 생성
- **파일 경로**: `moai-adk-ts/src/core/installer/tracked-filesystem.ts`
- **테스트 파일**: `moai-adk-ts/tests/core/installer/tracked-filesystem.test.ts`

### 2.2 구현 내용
```typescript
export class TrackedFileSystem {
  constructor(private transaction: InstallationTransaction);

  ensureDir(path: string): Promise<void>;
  writeFile(path: string, content: string): Promise<void>;
  copy(src: string, dest: string): Promise<void>;
  copyDir(src: string, dest: string): Promise<void>;
}
```

### 2.3 TDD 테스트 케이스
- [ ] ensureDir() 디렉토리 생성 및 추적
- [ ] writeFile() 파일 생성 및 추적
- [ ] copy() 파일 복사 및 추적
- [ ] copyDir() 디렉토리 재귀 복사 및 추적
- [ ] 모든 작업 후 transaction.track() 호출 검증

## 3. Phase 3: InstallerCore 통합

### 3.1 수정 파일
- **파일 경로**: `moai-adk-ts/src/core/installer/installer-core.ts`
- **테스트 파일**: `moai-adk-ts/tests/core/installer/installer-core.test.ts`

### 3.2 수정 내용
```typescript
export class InstallerCore {
  async install(options: InstallOptions): Promise<void> {
    const transaction = new InstallationTransaction();

    try {
      // 모든 Phase에 transaction 전달
      await this.executeAllPhases(options, transaction);
      transaction.commit();
    } catch (error) {
      const result = await transaction.rollback();
      this.reportRollback(result);
      throw error;
    }
  }

  private reportRollback(result: RollbackResult): void {
    // 롤백 리포트 출력
  }
}
```

### 3.3 TDD 테스트 케이스
- [ ] 정상 설치 시 commit() 호출 검증
- [ ] 설치 실패 시 rollback() 호출 검증
- [ ] rollback() 결과 리포트 출력 검증
- [ ] 원본 에러 전파 검증

## 4. Phase 4: Phase Executor 수정

### 4.1 수정 파일
- **파일 경로**: `moai-adk-ts/src/core/installer/phase-executor.ts`

### 4.2 수정 내용
- 모든 Phase 메서드에 transaction 파라미터 추가
- 파일 시스템 작업 시 TrackedFileSystem 사용
- 각 Phase에서 생성되는 파일/디렉토리 추적

### 4.3 영향받는 Phase
- PreInstall
- DependencyInstall
- TemplateInstall
- TypeScriptSetup
- PostInstall

## 5. Phase 5: 롤백 리포트 개선

### 5.1 RollbackReport 클래스
```typescript
export class RollbackReport {
  generate(result: RollbackResult): string;
  saveToFile(path: string, result: RollbackResult): Promise<void>;
}
```

### 5.2 리포트 포맷
```
🔄 Installation Rollback Report
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📊 Summary:
  ✅ Successfully deleted: 23 items
  ⚠️  Failed to delete: 2 items

✅ Deleted Files/Directories:
  - /path/to/file1.ts
  - /path/to/dir1/
  ...

⚠️  Failed Deletions:
  - /path/to/protected-file.ts
    Error: EACCES: permission denied
  - /path/to/locked-dir/
    Error: Directory not empty

💡 Manual Cleanup Required:
  Run: rm -rf /path/to/failed-items
```

## 6. Phase 6: 에러 처리 개선

### 6.1 InstallationError 확장
```typescript
export class RollbackError extends InstallationError {
  constructor(
    message: string,
    public readonly rollbackResult: RollbackResult
  ) {
    super(message);
    this.name = 'RollbackError';
  }
}
```

### 6.2 에러 체인
```
Original Error (설치 실패)
  ↓
Rollback Attempted
  ↓
Rollback Success → Original Error 전파
Rollback Partial → RollbackError with context
```

## 7. Phase 7: CLI 플래그 추가

### 7.1 새로운 옵션
```typescript
export interface InstallOptions {
  // 기존 옵션...

  /**
   * 설치 실패 시 부분 설치 유지 (롤백 안 함)
   */
  keepPartial?: boolean;

  /**
   * 롤백 리포트 파일 저장 경로
   */
  rollbackReportPath?: string;
}
```

### 7.2 CLI 인터페이스
```bash
moai-adk install --keep-partial        # 롤백 비활성화
moai-adk install --rollback-report=./  # 리포트 저장
```

## 8. 테스트 전략

### 8.1 단위 테스트
- InstallationTransaction 클래스
- TrackedFileSystem 클래스
- RollbackReport 클래스

### 8.2 통합 테스트
- InstallerCore의 전체 설치 → 실패 → 롤백 플로우
- 각 Phase 실패 시나리오별 롤백 검증

### 8.3 Edge Case 테스트
- 권한 없는 파일 삭제 시도
- 이미 삭제된 파일 롤백 시도
- 중첩 디렉토리 롤백
- 심볼릭 링크 처리

## 9. 일정

- **Phase 1**: 0.5일 (InstallationTransaction)
- **Phase 2**: 0.5일 (TrackedFileSystem)
- **Phase 3**: 1일 (InstallerCore 통합)
- **Phase 4**: 1일 (Phase Executor 수정)
- **Phase 5**: 0.5일 (롤백 리포트)
- **Phase 6**: 0.5일 (에러 처리)
- **Phase 7**: 0.5일 (CLI 플래그)
- **Total**: 4.5일

## 10. 체크리스트

- [ ] InstallationTransaction 클래스 구현 및 테스트
- [ ] TrackedFileSystem 래퍼 구현 및 테스트
- [ ] InstallerCore 통합 및 테스트
- [ ] 모든 Phase에 transaction 전달
- [ ] 롤백 리포트 생성 기능
- [ ] CLI 플래그 추가
- [ ] 문서 업데이트 (README.md)
