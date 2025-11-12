
## AC-001: 파일/디렉토리 추적

**GIVEN** InstallationTransaction이 활성화되어 있을 때
**WHEN** 파일 또는 디렉토리가 생성되면
**THEN**
- transaction.track()이 호출되어야 한다
- 생성된 경로가 createdPaths에 추가되어야 한다
- 동일 경로 중복 추가 시 한 번만 기록되어야 한다

## AC-002: 트랜잭션 커밋

**GIVEN** 모든 설치 Phase가 성공적으로 완료되었을 때
**WHEN** transaction.commit()이 호출되면
**THEN**
- isCommitted 플래그가 true로 설정되어야 한다
- createdPaths가 비워져야 한다
- 이후 track() 호출이 무시되어야 한다

## AC-003: 자동 롤백 트리거

**GIVEN** 설치 중 에러가 발생했을 때
**WHEN** InstallerCore의 catch 블록이 실행되면
**THEN**
- transaction.rollback()이 자동으로 호출되어야 한다
- 롤백 결과가 reportRollback()에 전달되어야 한다
- 원본 에러가 다시 throw되어야 한다

## AC-004: 역순 삭제

**GIVEN** 파일 A, B, C가 순서대로 생성되었을 때
**WHEN** 롤백이 실행되면
**THEN**
- 삭제 순서는 C → B → A여야 한다
- 각 파일이 실제로 삭제되어야 한다
- 디렉토리는 재귀적으로 삭제되어야 한다

## AC-005: 부분 롤백 성공

**GIVEN** 3개 파일 중 1개가 삭제 권한이 없을 때
**WHEN** 롤백이 실행되면
**THEN**
- 권한 있는 2개 파일은 삭제되어야 한다
- 권한 없는 1개 파일은 failed 목록에 포함되어야 한다
- 롤백 프로세스는 중단되지 않고 계속되어야 한다

## AC-006: 롤백 리포트 생성

**GIVEN** 롤백이 완료되었을 때
**WHEN** reportRollback()이 호출되면
**THEN**
- 삭제 성공 개수가 출력되어야 한다
- 삭제 실패 개수가 출력되어야 한다
- 실패한 각 파일의 경로와 에러 메시지가 출력되어야 한다

## AC-007: TrackedFileSystem 통합

**GIVEN** TrackedFileSystem을 사용하여 파일을 생성할 때
**WHEN** ensureDir(), writeFile(), copy() 등을 호출하면
**THEN**
- 파일 시스템 작업이 정상적으로 수행되어야 한다
- 각 작업 후 transaction.track()이 자동으로 호출되어야 한다
- 사용자가 명시적으로 track()을 호출할 필요가 없어야 한다

## AC-008: Phase Executor 통합

**GIVEN** 각 Phase가 실행될 때
**WHEN** 파일/디렉토리 작업이 수행되면
**THEN**
- TrackedFileSystem을 사용하여 작업해야 한다
- 모든 생성된 리소스가 자동으로 추적되어야 한다
- Phase 간 transaction 객체가 공유되어야 한다

## AC-009: 빈 트랜잭션 롤백

**GIVEN** 아무 파일도 생성되지 않은 상태에서 에러가 발생했을 때
**WHEN** 롤백이 실행되면
**THEN**
- 에러 없이 정상 완료되어야 한다
- "Deleted: 0 files/directories" 메시지가 출력되어야 한다
- 원본 에러만 전파되어야 한다

## AC-010: --keep-partial 플래그

**GIVEN** 사용자가 --keep-partial 플래그를 사용했을 때
**WHEN** 설치가 실패하면
**THEN**
- 롤백이 실행되지 않아야 한다
- 생성된 파일이 그대로 유지되어야 한다
- 부분 설치 유지 경고가 출력되어야 한다

## AC-011: 중첩 디렉토리 롤백

**GIVEN** 중첩 디렉토리 구조가 생성되었을 때
```
/project
  /src
    /components
      Component.ts
  /tests
    test.ts
```
**WHEN** 롤백이 실행되면
**THEN**
- 파일이 먼저 삭제되어야 한다 (Component.ts, test.ts)
- 빈 디렉토리가 삭제되어야 한다 (/components, /src, /tests)
- 모든 리소스가 정리되어야 한다

## AC-012: 롤백 리포트 파일 저장

**GIVEN** --rollback-report 플래그가 지정되었을 때
**WHEN** 롤백이 완료되면
**THEN**
- 지정된 경로에 rollback-report.txt 파일이 생성되어야 한다
- 리포트에 타임스탬프가 포함되어야 한다
- 리포트에 모든 삭제/실패 항목이 기록되어야 한다

## AC-013: 재설치 충돌 방지

**GIVEN** 이전 설치가 실패하고 롤백이 완료되었을 때
**WHEN** 동일 경로에 재설치를 시도하면
**THEN**
- "File already exists" 에러가 발생하지 않아야 한다
- 새로운 설치가 정상적으로 진행되어야 한다
- 이전 설치 흔적이 남아있지 않아야 한다

## AC-014: 심볼릭 링크 처리

**GIVEN** 심볼릭 링크가 생성되었을 때
**WHEN** 롤백이 실행되면
**THEN**
- 심볼릭 링크만 삭제되어야 한다 (원본 파일 보존)
- 깨진 심볼릭 링크도 삭제되어야 한다
- 에러 없이 정상 완료되어야 한다

## AC-015: Windows 경로 지원

**GIVEN** Windows 환경에서 설치가 실패했을 때
**WHEN** 롤백이 실행되면
**THEN**
- Windows 경로(C:\Users\...)가 올바르게 처리되어야 한다
- 경로 구분자(\)가 정상 작동해야 한다
- 긴 경로 이름(Long Path)이 지원되어야 한다

## AC-016: 동시성 안전성

**GIVEN** 여러 Phase가 병렬로 실행될 때
**WHEN** 각 Phase에서 파일을 생성하면
**THEN**
- 모든 생성 경로가 정확히 추적되어야 한다
- 경합 조건(race condition)이 발생하지 않아야 한다
- 롤백 시 모든 파일이 삭제되어야 한다
